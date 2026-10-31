"""Throwaway: pool every gold-labelled V-sem case with saved scores and tune merge.covered / merge.which_area
with the pilot's pinned tune_threshold.py, the pilot's minimums and the pilot's row rules (tune.round_rows logic).
No model calls; reads the run worktree's data only; writes CSVs and reports to /tmp/x2d-tune/ (or argv[2]).

Importable: `earlier_rows(S)` -> {question: [row]} for the 164 earlier cases (round-0 main, round-0 P1, re-gate P1, re-gate P2);
`write_and_tune(rows, out, tag)` runs the pinned script on one question's rows (point and conservative)."""
import csv, json, os, sys, subprocess

QUESTIONS = ("merge.covered", "merge.which_area")


def _setup(S):
    sys.path.insert(0, S)
    os.environ.setdefault("X2D_SKIP_SIGNOFF", "1")
    import x2d, compose, tune
    return x2d, compose, tune


def rows_for(x2d, compose, units, gold, scores, source):
    srows = x2d.read_jsonl(scores)
    flip_all = compose.header_flip_all(srows)
    ans, _ = compose._index(srows)
    out = {q: [] for q in QUESTIONS}
    for u in units:
        cid = u["case_id"]
        gr = gold.get(cid)
        if gr is None or not isinstance(gr.get("covered"), bool):
            continue
        acc = set(gr.get("acceptable") or []) if gr["covered"] else set()
        cov = ans.get((cid, "merge.covered", "V-sem", 1, None))
        if cov is None:
            continue
        meta = lambda r: {"model": r.get("model", ""), "question_version": r.get("question_version", ""), "split": "dev",
                          "stratum": source, "fallback": 0}
        out["merge.covered"].append({"case_id": cid, "label": int(not gr["covered"]), "score": 1.0 - cov["p_true"], **meta(cov)})
        need = compose.required_perms(u, flip_all)
        asked = [ans[k] for k in ((cid, "merge.which_area", "V-sem", p, None) for p in need) if k in ans]
        if len(asked) == len(need):
            same = len({r["choice"] for r in asked}) == 1
            # the which_area row carries the which_area question version (asked[0]), not the covered row's
            out["merge.which_area"].append({"case_id": cid, "label": int(same and asked[0]["choice"] in acc),
                                            "score": sum(r["confidence"] for r in asked) / len(asked),
                                            "cov_score": 1.0 - cov["p_true"], **meta(asked[0])})
    return out


def earlier_rows(S):
    """The 164 earlier gold-labelled V-sem cases (round-0 main, round-0 P1, re-gate P1, re-gate P2) from the study data root."""
    x2d, compose, tune = _setup(S)
    D = os.path.join(S, "data", "rounds", "0")
    RG = os.path.join(D, "regate", "rounds", "0")
    gold_of = lambda path: {g["case_id"]: g for g in x2d.read_jsonl(path)}
    units_of = lambda path, keep: [u for u in x2d.read_jsonl(path) if keep(u)]
    g0, gr = gold_of(os.path.join(D, "gold.jsonl")), gold_of(os.path.join(RG, "gold.jsonl"))
    sources = {
        "r0-main": rows_for(x2d, compose, units_of(os.path.join(D, "units.jsonl"), lambda u: u.get("control_kind") not in ("p1", "p2")),
                            g0, os.path.join(D, "jev", "scores.jsonl"), "r0-main"),
        # round-0 P2 excluded: derived gold, shown mis-built
        "r0-p1": rows_for(x2d, compose, units_of(os.path.join(D, "units.jsonl"), lambda u: u.get("control_kind") == "p1"),
                          g0, os.path.join(D, "jev-p", "scores.p1.jsonl"), "r0-p1"),
        "regate-p1": rows_for(x2d, compose, units_of(os.path.join(RG, "units.jsonl"), lambda u: u.get("control_kind") == "p1"),
                              gr, os.path.join(RG, "jev-p", "scores.p1.jsonl"), "regate-p1"),
        "regate-p2": rows_for(x2d, compose, units_of(os.path.join(RG, "units.jsonl"), lambda u: u.get("control_kind") == "p2"),
                              gr, os.path.join(RG, "jev-p", "scores.p2.jsonl"), "regate-p2"),
    }
    return {q: [r for s in sources.values() for r in s[q]] for q in QUESTIONS}, sources


def write_and_tune(x2d, tune, q, rows, out, tag):
    """Write scores CSV and run the pinned script (point estimate, conservative, conservative + --errors). -> {"point": json, "conservative": json}."""
    os.makedirs(out, exist_ok=True)
    e = tune.decision_entry(q)
    versions = {r["question_version"] for r in rows}
    keep = next(iter(versions)) if len(versions) == 1 else ""  # one group per (model, version): mixed versions are pooled, as tune.run_script does
    p = os.path.join(out, "scores.%s.%s.csv" % (tag, q))
    with open(p, "w", newline="") as f:
        w = csv.DictWriter(f, tune.COLUMNS)
        w.writeheader()
        for r in rows:
            w.writerow({k: r.get(k, "") for k in tune.COLUMNS} | {"question_version": keep})
    base = ["python3", "-B", x2d.TUNE_SCRIPT, p, "--min-precision", str(e["min_precision"]), "--min-recall", str(e["min_recall"]),
            "--beta", str(e["beta"]), "--folds", "5", "--seed", str(x2d.SEED), "--repeats", "5"]
    res = {}
    for mode, extra in (("point", []), ("conservative", ["--conservative"])):
        rj = subprocess.run(base + extra + ["--json"], capture_output=True, text=True)
        open(os.path.join(out, "report.%s.%s.%s.json" % (tag, q, mode)), "w").write(rj.stdout)
        rt = subprocess.run(base + extra + ["--errors"], capture_output=True, text=True)
        open(os.path.join(out, "report.%s.%s.%s.txt" % (tag, q, mode)), "w").write(rt.stdout + rt.stderr)
        res[mode] = json.loads(rj.stdout)[0] if rj.returncode == 0 and rj.stdout.strip() else {"error": rj.stderr[-300:]}
    return res


if __name__ == "__main__":
    S = sys.argv[1]  # study dir in the run worktree
    OUT = sys.argv[2] if len(sys.argv) > 2 else "/tmp/x2d-tune"
    x2d, compose, tune = _setup(S)
    pooled, sources = earlier_rows(S)
    print("counts (pos, neg):", json.dumps({s: {q: (sum(r["label"] for r in v[q]), len(v[q]) - sum(r["label"] for r in v[q])) for q in v}
                                            for s, v in sources.items()}))
    for q in QUESTIONS:
        pos = sum(r["label"] for r in pooled[q])
        res = write_and_tune(x2d, tune, q, pooled[q], OUT, "earlier")
        print("==", q, "pos", pos, "neg", len(pooled[q]) - pos, {m: v.get("shipped_threshold", v.get("error")) for m, v in res.items()})
