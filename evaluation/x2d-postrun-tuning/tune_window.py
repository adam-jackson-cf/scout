"""Development: tune merge.covered / merge.which_area on the V-sem scores of the window units of refinement rounds 1-7 (set == "window", the
pilot's tune.round_rows rules), alone (a) and together with the 164 earlier cases (b). Rounds 1-5 are split `dev`, rounds 6-7 `test` (held-out
confirmation at the shipped threshold). Pinned tune_threshold.py, the pilot's minimums and betas, --folds 5 --repeats 5 --seed 20261029, point-estimate
and --conservative runs. No model calls. Rounds 8-10 are never read.

  X2D_DATA=<worktree>/evaluation/x2d-merge-association-pilot/data/postrun \
  <worktree>/.tools/venv/bin/python tune_window.py <worktree>/evaluation/x2d-merge-association-pilot [OUT_DIR]

Writes OUT_DIR (default ./window): scores.<variant>.<question>.csv, report.<variant>.<question>.{point,conservative}.{json,txt}, summary.json.
"""
import json, os, sys

HERE = os.path.dirname(os.path.abspath(__file__))
S = os.path.abspath(sys.argv[1])
OUT = os.path.abspath(sys.argv[2]) if len(sys.argv) > 2 else os.path.join(HERE, "window")
os.environ.setdefault("X2D_DATA", os.path.join(S, "data", "postrun"))
sys.path.insert(0, HERE)
import tune_pool
x2d, compose, tune = tune_pool._setup(S)

DATA = os.environ["X2D_DATA"]
assert os.path.realpath(DATA) != os.path.realpath(os.path.join(S, "data")), "X2D_DATA must be the postrun root"
ROUNDS = range(1, 8)  # refinement rounds 1-7 only
TEST_ROUNDS = (6, 7)
QS = tune_pool.QUESTIONS


def window_rows():
    out = {q: [] for q in QS}
    for rr in ROUNDS:
        rows = tune.round_rows(rr, DATA, "test" if rr in TEST_ROUNDS else "dev")
        for q in QS:
            for r in rows[q]:
                out[q].append(dict(r, stratum="r%d" % rr))
    return out


def counts(rows):
    d = [r for r in rows if r["split"] == "dev"]
    t = [r for r in rows if r["split"] == "test"]
    f = lambda xs: [sum(r["label"] for r in xs), len(xs) - sum(r["label"] for r in xs)]
    return {"dev_pos_neg": f(d), "test_pos_neg": f(t)}


def main():
    win = window_rows()
    earlier, _ = tune_pool.earlier_rows(S)
    variants = {"window": win, "window+earlier": {q: win[q] + earlier[q] for q in QS}}
    summary = {"rounds": list(ROUNDS), "test_rounds": list(TEST_ROUNDS), "variants": {}}
    for v, rows in variants.items():
        tag = v.replace("+", "_")
        summary["variants"][v] = {}
        for q in QS:
            res = tune_pool.write_and_tune(x2d, tune, q, rows[q], OUT, tag)
            e = tune.decision_entry(q)
            summary["variants"][v][q] = {"counts": counts(rows[q]), "min_precision": e["min_precision"], "min_recall": e["min_recall"], "beta": e["beta"], **res}
            print(v, q, counts(rows[q]), {m: x.get("shipped_threshold", x.get("error")) for m, x in res.items()})
    json.dump(summary, open(os.path.join(OUT, "summary.json"), "w"), indent=1, sort_keys=True)


if __name__ == "__main__":
    main()
