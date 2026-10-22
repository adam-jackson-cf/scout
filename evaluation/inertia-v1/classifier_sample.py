#!/usr/bin/env python3
"""Draw the seeded 100-commit validation sample (stratified ~equal per repository) from commits.jsonl.gz
and print it for MANUAL labelling WITHOUT showing the regex verdict. Writes classifier-sample.jsonl.

python classifier_sample.py
"""
import gzip
import json
import random
import subprocess
from collections import defaultdict
from pathlib import Path

HERE = Path(__file__).resolve().parent
SEED = 20261016
commits = [json.loads(l) for l in gzip.open(HERE / "commits.jsonl.gz", "rt")]
by = defaultdict(list)
for c in commits:
    by[c["repo"]].append(c)
rng = random.Random(SEED)
repos = sorted(by)
quota = {r: 100 // len(repos) for r in repos}
for r in rng.sample(repos, 100 - sum(quota.values())):
    quota[r] += 1
sample = []
for r in repos:
    pool = sorted(by[r], key=lambda c: c["hash"])
    sample += rng.sample(pool, min(quota[r], len(pool)))
rng.shuffle(sample)
sample = sample[:100]
with open(HERE / "classifier-sample.jsonl", "w") as f:
    for c in sample:
        f.write(json.dumps(c) + "\n")
out = []
for i, c in enumerate(sample):
    stat = subprocess.run(["git", "-C", f"/tmp/scout-evidence/repos/{c['repo']}", "show", "--stat=100,60", "--format=", c["hash"]],
                          capture_output=True, text=True, errors="replace").stdout.strip().splitlines()
    out.append(f"#{i} {c['repo']} {c['hash'][:10]} files={c['nfiles']}\nSUBJ: {c['subject']}\nBODY: {c['body'][:400].strip()!r}\nSTAT:\n  " +
               "\n  ".join(stat[:6] + (stat[-1:] if len(stat) > 6 else [])))
(HERE / "classifier-sample-view.txt").write_text("\n\n".join(out))
print(len(sample), {r: sum(1 for c in sample if c["repo"] == r) for r in repos})
