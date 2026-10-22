#!/usr/bin/env python3
"""Hand labels for the 100-commit classifier sample (labelled from message + diff stat in
classifier-sample-view.txt, which does NOT show the regex verdict) and precision/recall of the
preregistered regex. Writes classifier-labels.jsonl and classifier-metrics.json.

Rubric: fix = the commit repairs a defect (wrong/broken behaviour, crash, leak, regression, deadlock) in
production or test code, or reverts a change because it broke something. Not fixes: dependency bumps,
CI/config tweaks, docs/typo edits, releases, features, refactors, new tests.
"""
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent
T = {  # index -> rationale for TRUE (defect fix)
    10: "fixes duplicated DbContext registration bug in test fixture",
    15: "fixes httploader resource leak (fixes #5158)",
    18: "fixes memory leaks in mini-css via dependency bump; intent is a defect fix",
    19: "fix: logic with escaped chars in CSS parser, code change plus test",
    31: "fixes wrong reporting of enclosed parameterized tests; one-line code fix plus tests",
    46: "corrects Origin header on API requests (community bug report)",
    51: "fixes wrong fallback for error handler chain (mistaken assumption), small code fix",
    59: "reorders init to avoid admin-API deadlock (bug report link)",
    73: "repairs failing test cases under node 17",
    89: "bug fix for issue #16045, one-line code change",
    95: "fix context: one-line code fix with snapshot update",
    98: "revert terser dependency change (reverts a breaking change)",
}
F_RAT = {
    "deps": "dependency/lockfile bump, not a defect fix",
    "docs": "documentation/readme/release-notes/typo edit, not a code defect",
    "ci": "CI/build/config change, not a code defect",
    "release": "release/version bump",
    "feature": "new feature or enhancement",
    "refactor": "refactor/tweak/cleanup with no stated defect",
    "test": "adds or adjusts tests only",
}
F = {  # index -> category
    **{i: "deps" for i in (1, 3, 12, 23, 27, 32, 40, 49, 52, 54, 55, 61, 63, 65, 67, 72, 79, 82, 85, 88, 91, 93, 96)},
    **{i: "docs" for i in (2, 5, 7, 11, 26, 30, 33, 34, 36, 37, 38, 41, 43, 44, 45, 47, 48, 50, 53, 60, 64, 76, 78, 80, 81, 84, 86, 90)},
    **{i: "ci" for i in (4, 14, 39, 42, 71, 77)},
    **{i: "release" for i in (17, 62, 87)},
    **{i: "feature" for i in (24, 25, 56, 57, 58, 66, 68, 69, 74, 75, 83, 92, 99)},
    **{i: "refactor" for i in (0, 6, 8, 9, 13, 20, 21, 22, 28, 29, 35, 70)},
    **{i: "test" for i in (16, 94, 97)},
}
assert not (set(T) & set(F)), set(T) & set(F)
missing = set(range(100)) - set(T) - set(F)
assert not missing, missing

sample = [json.loads(l) for l in open(HERE / "classifier-sample.jsonl")]
tp = fp = fn = tn = 0
rows = []
for i, c in enumerate(sample):
    truth = i in T
    pred = bool(c["regex_fix"])
    rat = T[i] if truth else F_RAT[F[i]]
    tp += truth and pred
    fp += (not truth) and pred
    fn += truth and not pred
    tn += (not truth) and not pred
    rows.append(dict(idx=i, repo=c["repo"], hash=c["hash"], subject=c["subject"], label_fix=truth,
                     regex_fix=pred, rationale=rat))
with open(HERE / "classifier-labels.jsonl", "w") as f:
    for r in rows:
        f.write(json.dumps(r, ensure_ascii=False) + "\n")
m = dict(n=len(rows), true_fix=tp + fn, tp=tp, fp=fp, fn=fn, tn=tn,
         precision=tp / (tp + fp) if tp + fp else None, recall=tp / (tp + fn) if tp + fn else None,
         false_positives=[(r["idx"], r["subject"]) for r in rows if r["regex_fix"] and not r["label_fix"]],
         false_negatives=[(r["idx"], r["subject"]) for r in rows if r["label_fix"] and not r["regex_fix"]])
(HERE / "classifier-metrics.json").write_text(json.dumps(m, indent=1, ensure_ascii=False))
print(json.dumps(m, indent=1, ensure_ascii=False))

# ---- secondary: restrict to commits touching >=1 in-scope production source file (exclusions.json)
import subprocess  # noqa: E402
from pathlib import Path as _P  # noqa: E402
from collect import glob_re, EXTS  # noqa: E402

excl = json.loads((HERE / "exclusions.json").read_text())
sub = dict(tp=0, fp=0, fn=0, tn=0)
for r, c in zip(rows, sample):
    files = subprocess.run(["git", "-C", f"/tmp/scout-evidence/repos/{c['repo']}", "show", "--name-only", "--format=", c["hash"]],
                           capture_output=True, text=True, errors="replace").stdout.split("\n")
    rx = [glob_re(g) for g in excl["repos"][c["repo"]]]
    insrc = any(f and _P(f).suffix in EXTS and not any(x.match(f) for x in rx) for f in files)
    r["touches_source"] = insrc
    if insrc:
        sub[("tp" if r["regex_fix"] else "fn") if r["label_fix"] else ("fp" if r["regex_fix"] else "tn")] += 1
m["source_touching_subset"] = dict(sub, n=sum(sub.values()),
                                   precision=sub["tp"] / (sub["tp"] + sub["fp"]) if sub["tp"] + sub["fp"] else None,
                                   recall=sub["tp"] / (sub["tp"] + sub["fn"]) if sub["tp"] + sub["fn"] else None)
(HERE / "classifier-metrics.json").write_text(json.dumps(m, indent=1, ensure_ascii=False))
with open(HERE / "classifier-labels.jsonl", "w") as f:
    for r in rows:
        f.write(json.dumps(r, ensure_ascii=False) + "\n")
print(m["source_touching_subset"])
