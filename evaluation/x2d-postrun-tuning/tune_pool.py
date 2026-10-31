"""Throwaway: pool every gold-labelled V-sem case with saved scores and tune merge.covered / merge.which_area
with the pilot's pinned tune_threshold.py, the pilot's minimums and the pilot's row rules (tune.round_rows logic).
No model calls; reads the run worktree's data only; writes CSVs and reports to /tmp/x2d-tune/."""
import json, os, sys, subprocess, collections
S = sys.argv[1]  # study dir in the run worktree
sys.path.insert(0, S)
os.environ.setdefault("X2D_SKIP_SIGNOFF", "1")
import x2d, compose, tune

D = os.path.join(S, "data", "rounds", "0")
RG = os.path.join(D, "regate", "rounds", "0")
OUT = "/tmp/x2d-tune"
os.makedirs(OUT, exist_ok=True)


def rows_for(units, gold, scores, source):
    srows = x2d.read_jsonl(scores)
    flip_all = compose.header_flip_all(srows)
    ans, _ = compose._index(srows)
    out = {"merge.covered": [], "merge.which_area": []}
    for u in units:
        cid = u["case_id"]
        gr = gold.get(cid)
        if gr is None or not isinstance(gr.get("covered"), bool):
            continue
        acc = set(gr.get("acceptable") or []) if gr["covered"] else set()
        cov = ans.get((cid, "merge.covered", "V-sem", 1, None))
        if cov is None:
            continue
        meta = {"model": cov.get("model", ""), "question_version": cov.get("question_version", ""), "split": "dev",
                "stratum": source, "fallback": 0}
        out["merge.covered"].append({"case_id": cid, "label": int(not gr["covered"]), "score": 1.0 - cov["p_true"], **meta})
        need = compose.required_perms(u, flip_all)
        asked = [ans[k] for k in ((cid, "merge.which_area", "V-sem", p, None) for p in need) if k in ans]
        if len(asked) == len(need):
            same = len({r["choice"] for r in asked}) == 1
            out["merge.which_area"].append({"case_id": cid, "label": int(same and asked[0]["choice"] in acc),
                                            "score": sum(r["confidence"] for r in asked) / len(asked),
                                            "cov_score": 1.0 - cov["p_true"], **meta})
    return out


def gold_of(path):
    return {g["case_id"]: g for g in x2d.read_jsonl(path)}


def units_of(path, keep):
    return [u for u in x2d.read_jsonl(path) if keep(u)]


g0, gr = gold_of(os.path.join(D, "gold.jsonl")), gold_of(os.path.join(RG, "gold.jsonl"))
sources = {
    # round-0 main run: D/P units plus the dup/trap/prose/planted controls it answered (gold from labels)
    "r0-main": rows_for(units_of(os.path.join(D, "units.jsonl"), lambda u: u.get("control_kind") not in ("p1", "p2")),
                        g0, os.path.join(D, "jev", "scores.jsonl"), "r0-main"),
    # round-0 P1: gold = labellers' agreed acceptable set (gold.jsonl); round-0 P2 excluded (derived gold, mis-built)
    "r0-p1": rows_for(units_of(os.path.join(D, "units.jsonl"), lambda u: u.get("control_kind") == "p1"),
                      g0, os.path.join(D, "jev-p", "scores.p1.jsonl"), "r0-p1"),
    "regate-p1": rows_for(units_of(os.path.join(RG, "units.jsonl"), lambda u: u.get("control_kind") == "p1"),
                          gr, os.path.join(RG, "jev-p", "scores.p1.jsonl"), "regate-p1"),
    "regate-p2": rows_for(units_of(os.path.join(RG, "units.jsonl"), lambda u: u.get("control_kind") == "p2"),
                          gr, os.path.join(RG, "jev-p", "scores.p2.jsonl"), "regate-p2"),
}
pooled = {q: [r for s in sources.values() for r in s[q]] for q in ("merge.covered", "merge.which_area")}
counts = {s: {q: (sum(r["label"] for r in v[q]), len(v[q]) - sum(r["label"] for r in v[q])) for q in v} for s, v in sources.items()}
print("counts (pos, neg):", json.dumps(counts))
for q, rows in pooled.items():
    print(q, "pooled pos/neg:", sum(r["label"] for r in rows), len(rows) - sum(r["label"] for r in rows))

script = x2d.TUNE_SCRIPT
res = {}
for q in ("merge.covered", "merge.which_area"):
    e = tune.decision_entry(q)
    rows = pooled[q]
    p = os.path.join(OUT, "scores.%s.csv" % q)
    import csv
    with open(p, "w", newline="") as f:
        w = csv.DictWriter(f, tune.COLUMNS)
        w.writeheader()
        for r in rows:
            w.writerow({k: r.get(k, "") for k in tune.COLUMNS})
    pos = sum(r["label"] for r in rows); neg = len(rows) - pos
    cmd = ["python3", "-B", script, p, "--min-precision", str(e["min_precision"]), "--min-recall", str(e["min_recall"]),
           "--beta", str(e["beta"]), "--folds", "5", "--seed", str(x2d.SEED), "--repeats", "5", "--conservative", "--errors"]
    r = subprocess.run(cmd, capture_output=True, text=True)
    open(os.path.join(OUT, "report.%s.txt" % q), "w").write(r.stdout + r.stderr)
    rj = subprocess.run(cmd[:-1] + ["--json"], capture_output=True, text=True)
    open(os.path.join(OUT, "report.%s.json" % q), "w").write(rj.stdout)
    # point-estimate run too (no --conservative), for the best-guess value
    rp = subprocess.run([c for c in cmd if c not in ("--conservative", "--errors")] + ["--json"], capture_output=True, text=True)
    open(os.path.join(OUT, "report.%s.point.json" % q), "w").write(rp.stdout)
    print("==", q, "pos", pos, "neg", neg, "rc", r.returncode, rj.returncode, rp.returncode)
