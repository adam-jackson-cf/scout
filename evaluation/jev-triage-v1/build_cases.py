import json,subprocess,random
R=json.load(open('../reference-v1.json'))
repomap={'fastapi':'fastapi','webpack':'webpack','nest':'nest','caddy':'caddy','axum':'axum','junit-framework':'junit5','junit5':'junit5','eShopOnWeb':'eShopOnWeb'}
# id -> (tool/rule signal neutral, [(path,start,end)...], deleak note)
S={
'architecture-routing-utils-cycle-d666ccb6':("architecture analyzer rule ARCHITECTURE007: circular import between fastapi/routing.py and fastapi/utils.py",[('fastapi/routing.py',20,56),('fastapi/utils.py',1,32)],"Dropped 'not a runtime cycle' reasoning; signal restated as the rule's claim. Excerpts = cited lines +-15."),
'security-python-multipart-redos-test-extra-d666ccb6':("dependency/security analyzer: python-multipart version constraint in pyproject.toml reported as a security finding (CVE-affected range check)",[('pyproject.toml',45,78),('fastapi/routing.py',175,205)],"Removed GHSA id, patched version and fix commit; kept only that the constraint was reported by a dependency check."),
'security-unpinned-publish-actions-d666ccb6':("semgrep rules SECURITY022-SECURITY025: GitHub Action referenced by version tag instead of commit SHA",[('.github/workflows/publish.yml',1,45)],"Removed note about PyPI token exposure/verdict; excerpt is raw workflow file."),
'quality-long-openapi-test-fixture-d666ccb6':("lizard rule COMPLEXITY001: function length 194 lines, cyclomatic complexity 1",[('tests/test_generate_unique_id_function.py',20,75)],"Removed 'test fixture / expected OpenAPI document' description; excerpt shows file head only (function is 194 lines, excerpt truncated to cited lines +-25)."),
'security-auto-public-path-dom-clobbering':("security analyzer structural signal: lib/runtime/AutoPublicPathRuntimeModule.js reads document.currentScript.src and assigns the public path",[('lib/runtime/AutoPublicPathRuntimeModule.js',25,80)],"Removed GHSA id, affected versions and the upstream fix description; excerpt only the cited module lines."),
'typescript-architecture-loopback-url-676c98ca':("architecture scalability analyzer rule ARCHITECTURE001 at packages/core/nest-application.ts:361: hardcoded configuration value",[('packages/core/nest-application.ts',340,380)],"Removed 'only assigned within URL formatting' explanation."),
'typescript-security-mutable-github-actions-676c98ca':("semgrep rules SECURITY003-SECURITY006: GitHub Action referenced by version tag instead of commit SHA",[('.github/workflows/codeql-analysis.yml',1,70)],"Removed note about security-events permission as justification."),
'typescript-quality-route-params-complexity-676c98ca':("lizard: cyclomatic complexity 21 at line 9 of packages/core/router/route-params-factory.ts (high complexity)",[('packages/core/router/route-params-factory.ts',1,60)],"Removed 'exhaustive dispatch / bounded mapping' description."),
'typescript-performance-sponsor-images-676c98ca':("frontend performance analyzer rule PERFORMANCE012 at Readme.md:63: unoptimized image format",[('Readme.md',55,90)],"Removed 'sponsor table' framing; excerpt is raw file lines."),
'go-security-release-mutable-actions-0db29e2c':("semgrep rule SECURITY002 and related: GitHub Action referenced by version tag or branch instead of commit SHA",[('.github/workflows/release.yml',1,125)],"Removed 'privileged release job' reasoning and permission commentary."),
'rust-architecture-benchmark-blocking-false-positive-562fa206':("architecture scalability analyzer rule ARCHITECTURE001 at axum/benches/benches.rs lines 12-25 region: blocking call (child.wait / synchronous I/O)",[('axum/benches/benches.rs',175,225)],"Removed 'standalone benchmark runner' framing and verdict in id (id not shown to Jev)."),
'rust-security-mutable-action-references-562fa206':("semgrep rules SECURITY001-SECURITY049: GitHub Action referenced by branch/version tag instead of commit SHA",[('.github/workflows/CI.yml',1,80)],"Removed GitHub guidance URL and verdict text."),
'rust-security-local-websocket-test-false-positive-562fa206':("security analyzer rule SECURITY053: plain ws:// (unencrypted WebSocket) URL",[('examples/testing-websockets/src/main.rs',80,125)],"Removed 'loopback-only integration test' explanation."),
'rust-quality-documented-query-overlap-false-positive-562fa206':("jscpd rules QUALITY209-QUALITY210: duplicate code between axum/src/extract/query.rs and axum-extra/src/extract/query.rs",[('axum/src/extract/query.rs',1,60),('axum-extra/src/extract/query.rs',1,65)],"Removed note that the latter documents a deliberate different extractor."),
'rust-performance-websocket-demo-resource-false-positive-562fa206':("frontend performance analyzer rules PERFORMANCE001-PERFORMANCE004: timers / event listeners without cleanup (resource leak)",[('examples/websockets/assets/script.js',1,40)],"Removed lifetime/one-shot reasoning."),
'java-architecture-fixture-pom-parse-8ed3c66':("architecture dependency analyzer parsed platform-tests/src/test/resources/listeners/uidtracking/maven/pom.xml as a dependency manifest",[('platform-tests/src/test/resources/listeners/uidtracking/maven/pom.xml',1,40)],"Removed statement that line 1 says it's a test fixture as reasoning (raw file still shows it)."),
'java-security-unpinned-github-actions-8ed3c66':("semgrep rules SECURITY001-SECURITY030: GitHub Action referenced by version tag instead of commit SHA",[('.github/workflows/main.yml',1,60),('.github/actions/run-gradle/action.yml',1,40)],"Removed 'workflow credentials available' conclusion."),
'java-quality-snapshot-duplication-8ed3c66':("jscpd: duplicate code block between Skip-skipWithSingleLineReason-summary-unicode.out.txt and ...-summary-ascii.out.txt lines 2-14",[('platform-tests/src/test/resources/console/details/skip/Skip-skipWithSingleLineReason-summary-unicode.out.txt',1,20),('platform-tests/src/test/resources/console/details/skip/Skip-skipWithSingleLineReason-summary-ascii.out.txt',1,20)],"Removed 'intentionally similar fixtures' phrase."),
'csharp-architecture-generated-migration-length-4da8212':("architecture patterns analyzer rule ARCHITECTURE060: method too long (src/Infrastructure/Data/Migrations/20201202111507_InitialModel.cs)",[('src/Infrastructure/Data/Migrations/20201202111507_InitialModel.cs',1,45)],"Removed 'generated' / 'EF Core' wording from signal."),
'csharp-quality-generated-migration-length-4da8212':("lizard rule COMPLEXITY008: function length 152 lines, cyclomatic complexity 1 (high severity)",[('src/Infrastructure/Data/Migrations/20201202111507_InitialModel.cs',1,45)],"Removed 'generated migration' framing."),
'csharp-security-mutable-actions-4da8212':("semgrep rules SECURITY002, SECURITY003: GitHub Action referenced by version tag instead of commit SHA",[('.github/workflows/dotnetcore.yml',1,40)],"Removed 'privileged' wording."),
'csharp-performance-identity-height-transition-4da8212':("frontend performance analyzer rules PERFORMANCE013-PERFORMANCE016: CSS transition on layout-affecting property (height)",[('src/Web/wwwroot/css/shared/components/identity/identity.css',15,60),('src/Web/Views/Shared/_Layout.cshtml',10,35)],"Removed 'served in production bundle' conclusion; layout file is included as raw excerpt."),
}
rows=[]
for e in R['entries']:
  repo=e['target']['repository'].split('/')[-1].removesuffix('.git'); repo=repomap.get(repo,repo)
  rev=e['target']['revision']
  for i in e['issues']:
    if i['adjudication'] not in('validated_true','false_positive'):continue
    sig,rngs,note=S[i['id']]
    ex=[]
    for p,a,b in rngs:
      t=subprocess.run(['git','-C',f'/tmp/scout-evidence/repos/{repo}','show',f'{rev}:{p}'],capture_output=True,text=True)
      if t.returncode: raise SystemExit(p+t.stderr)
      L=t.stdout.split('\n'); a=max(1,a); b=min(len(L),b)
      ex.append({'file':p,'lines':f'{a}-{b}','text':'\n'.join(f'{n}: {L[n-1]}' for n in range(a,b+1))})
    rows.append({'case_id':i['id'],'label':'real_defect' if i['adjudication']=='validated_true' else 'false_positive','repo':repo,'rev':rev,
      'state':{'tool_signal':sig,'path':i['path'],'excerpts':ex},'deleak':note})
random.seed(7)
ids=[r['case_id'] for r in rows]; random.shuffle(ids)
# stratified split
pos=[r['case_id'] for r in rows if r['label']=='real_defect']; neg=[r['case_id'] for r in rows if r['label']!='real_defect']
random.shuffle(pos);random.shuffle(neg)
held=set(pos[:5]+neg[:7])
for r in rows: r['split']='heldout' if r['case_id'] in held else 'tuning'
open('cases.jsonl','w').write('\n'.join(json.dumps(r) for r in rows)+'\n')
print(len(rows),sum(r['split']=='heldout' for r in rows))
# unlabelled
Q=json.load(open('../scout-requests-v2.json'))
un=[]
for q in Q:
  for o in q['observations']:
    if o.get('adjudication') in('pending_review','incomplete'): un.append((q,o))
print(len(un))
import collections
print(collections.Counter((q['target']['repository'].split('/')[-1],o['source']) for q,o in un).most_common(40))
