"""Development: sweep the covered score threshold s (none iff 1 - p_true >= s, i.e. p_true <= 1 - s; "never none" = s 2.0) with which_area held at 0.535
and 0.6125, composed V-sem over the window units of rounds 1-7, saved scores only. Best pair chosen on rounds 1-5 (max R_assign_res s.t. P_assign >= 0.90 and
abstention <= 0.15), then reported on rounds 6-7.

  X2D_DATA=<worktree>/.../data/postrun <worktree>/.tools/venv/bin/python sweep_window.py <study dir> [OUT_DIR]
"""
import json, os, sys
import compose_window as cw
compose, A, x2d = cw.compose, cw.A, cw.x2d

S_LIST = [("0.30 (= old t_cov 0.70)", 0.3), ("0.40", 0.4), ("0.50", 0.5), ("0.60", 0.6), ("0.70", 0.7), ("0.80", 0.8), ("0.85", 0.85), ("0.90", 0.9), ("0.95", 0.95), ("0.99", 0.99), ("never none", 2.0)]
T_CHOICE = (0.535, 0.6125)
PARTS = {"all": [1, 2, 3, 4, 5, 6, 7], "dev": [1, 2, 3, 4, 5], "test": [6, 7]}


def run(data, s, tc, rs):
    th = {"decisions": {"merge.covered": {"threshold": None, "score_threshold": s}, "merge.which_area": {"threshold": tc}}}
    outs = []
    for r in rs:
        g, units, scores, _ = data[r]
        rows = {x["case_id"]: x for x in compose.compose(scores, units, th, "V-sem")}
        outs += [A.score_outcome(rows[u["case_id"]], g[u["case_id"]]) for u in units]
    f = cw.figures(A.counts_for(outs))
    f["gold_covered_answered_none"] = sum(1 for o in outs if o["kind"] == "none_wrong")
    return f


def main():
    data = {r: cw.load(r) for r in cw.ROUNDS}
    res = []
    for tc in T_CHOICE:
        for label, s in S_LIST:
            res.append({"t_choice": tc, "s_cov": label, "s": s, **{p: run(data, s, tc, rs) for p, rs in PARTS.items()}})
    ok = [x for x in res if x["dev"]["P_assign"] is not None and x["dev"]["P_assign"] >= 0.90 and x["dev"]["abstention"] <= 0.15]
    best = max(ok, key=lambda x: (x["dev"]["R_assign_res"], -x["dev"]["abstention"])) if ok else None
    json.dump({"rows": res, "best_on_dev": best}, open(os.path.join(cw.OUT, "sweep.json"), "w"), indent=1)
    for x in res:
        for p in ("all", "test"):
            f = x[p]
            print(x["t_choice"], x["s_cov"], p, "P %.3f R %.3f abst %.3f none %d noneP %s noneR %s covNone %d" % (
                f["P_assign"], f["R_assign_res"], f["abstention"], f["none"], f["P_none"] and round(f["P_none"], 3), round(f["R_none"], 3), f["gold_covered_answered_none"]))
    print("BEST", best and (best["t_choice"], best["s_cov"]), best and {p: best[p] for p in ("dev", "test")})


if __name__ == "__main__":
    main()
