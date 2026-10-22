import json, random, subprocess, re, collections
Q = json.load(open('../scout-requests-v2.json'))
repomap = {'junit5': 'junit5'}
groups = collections.defaultdict(list)
for q in Q:
    repo = q['target']['repository'].split('/')[-1].removesuffix('.git')
    for o in q['observations']:
        if o.get('adjudication') in ('pending_review', 'incomplete'):
            m = re.search(r':(\d+);', o['evidence'])
            if m: groups[(repo, o['source'])].append((q, o, repo, int(m.group(1))))
random.seed(11)
for g in groups.values(): random.shuffle(g)
pick = []; keys = sorted(groups)
while len(pick) < 30 and any(groups.values()):
    for k in keys:
        if groups[k] and len(pick) < 30: pick.append(groups[k].pop())
rows = []
for q, o, repo, line in pick:
    rev = q['target']['revision']; p = o['path']
    t = subprocess.run(['git', '-C', f'/tmp/scout-evidence/repos/{repo}', 'show', f'{rev}:{p}'], capture_output=True, text=True)
    if t.returncode: continue
    L = t.stdout.split('\n'); a = max(1, line - 25); b = min(len(L), line + 25)
    rule = o['mechanism'].replace('unreviewed ', '')
    rows.append({'case_id': o['id'][:12], 'label': None, 'split': 'unlabelled', 'repo': repo, 'rev': rev,
        'state': {'tool_signal': f"{o['source']} rule {rule} at {p}:{line}", 'path': p,
                  'excerpts': [{'file': p, 'lines': f'{a}-{b}', 'text': '\n'.join(f'{n}: {L[n-1]}' for n in range(a, b + 1))}]},
        'deleak': 'No adjudication exists; state = observation source/rule/path:line (run ids stripped) + file lines line+-25.'})
open('unlabelled.jsonl', 'w').write('\n'.join(json.dumps(r) for r in rows) + '\n')
print(len(rows))
