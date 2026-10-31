"""Development: per-round counts and the composed V-sem check on the window units of rounds 1-7, at the pilot's 0.70/0.70, the tuned values and the
B2 baseline on the same units. Reuses compose.compose and analysis_x2d.score_outcome / counts_for / rates. No model calls.

  X2D_DATA=<worktree>/evaluation/x2d-merge-association-pilot/data/postrun <worktree>/.tools/venv/bin/python compose_window.py <study dir> [OUT_DIR]

Writes OUT_DIR/rounds.json and OUT_DIR/composed.json (the tuned values are read from OUT_DIR/summary.json written by tune_window.py).
"""
import json, os, sys

HERE = os.path.dirname(os.path.abspath(__file__))
S = os.path.abspath(sys.argv[1])
OUT = os.path.abspath(sys.argv[2]) if len(sys.argv) > 2 else os.path.join(HERE, "window")
os.environ.setdefault("X2D_DATA", os.path.join(S, "data", "postrun"))
sys.path.insert(0, S)
os.environ.setdefault("X2D_SKIP_SIGNOFF", "1")
import x2d, compose, gold as G, analysis_x2d as A

DATA = os.environ["X2D_DATA"]
assert os.path.realpath(DATA) != os.path.realpath(os.path.join(S, "data"))
ROUNDS = range(1, 8)
TEST = (6, 7)


def rd(r, *p):
    return os.path.join(x2d.round_dir(r, DATA), *p)


def load(r):
    g = G.load_gold(r, DATA)
    units = [u for u in x2d.read_jsonl(rd(r, "units.jsonl")) if u.get("set") == "window" and not u.get("control")]
    scores = x2d.read_jsonl(rd(r, "jev", "scores.jsonl"))
    b2 = {a["case_id"]: a for a in x2d.read_jsonl(rd(r, "arms", "B2.jsonl"))}
    assert not compose.missing_required(scores, units), "round %d jev incomplete" % r
    assert all(u["case_id"] in g and isinstance(g[u["case_id"]].get("covered"), bool) for u in units), "round %d: window unit without gold" % r
    return g, units, scores, b2


def figures(c):
    r = A.rates(c)
    f = lambda k: r[k]["value"]
    return {"n": r["n"], "assigned": r["assigned"], "none": r["none"], "unresolved": r["unresolved"],
            "gold_covered": r["gold_covered"], "gold_covered_res": r["gold_covered_res"], "gold_none": r["gold_none"],
            "P_assign": f("P_assign"), "R_assign_res": f("R_assign_res"), "R_assign": f("R_assign"), "abstention": f("unresolved_rate"),
            "P_none": f("P_new"), "R_none": f("R_new"),
            "assigned_correct": c.get("assigned_correct", 0), "none_correct": c.get("none_correct", 0)}


def th(t_cov, t_choice):
    return {"decisions": {"merge.covered": {"threshold": t_cov, "score_threshold": None}, "merge.which_area": {"threshold": t_choice}}}


def main():
    summ = json.load(open(os.path.join(OUT, "summary.json")))
    tuned = {v: summ["variants"][v]["merge.which_area"]["point"].get("shipped_threshold") for v in summ["variants"]}
    settings = {"0.70/0.70": th(0.7, 0.7)}
    for v, t in tuned.items():  # covered has no qualifying threshold in any run: it stays at the provisional 0.7
        if t is not None:
            settings["tuned %s (t_cov 0.70 / t_choice %.4f)" % (v, t)] = th(0.7, t)
    data = {r: load(r) for r in ROUNDS}
    parts = {"all (rounds 1-7)": list(ROUNDS), "dev (rounds 1-5)": [1, 2, 3, 4, 5], "test (rounds 6-7)": list(TEST)}
    res = {"tuned_thresholds_from": {"which_area_point_shipped": tuned, "covered": "no qualifying threshold (provisional 0.70)"}, "settings": {}}
    for name, t in list(settings.items()) + [("B2", None)]:
        res["settings"][name] = {}
        for pname, rs in parts.items():
            outs = []
            for r in rs:
                g, units, scores, b2 = data[r]
                rows = {x["case_id"]: x for x in (compose.compose(scores, units, t, "V-sem") if t else [b2[u["case_id"]] for u in units])}
                outs += [A.score_outcome(rows.get(u["case_id"]), g[u["case_id"]]) for u in units]
            res["settings"][name][pname] = figures(A.counts_for(outs))
    # per-round counts
    per = {}
    for r in ROUNDS:
        g, units, scores, b2 = data[r]
        allu = x2d.read_jsonl(rd(r, "units.jsonl"))
        sess = json.load(open(rd(r, "sessions.json")))
        gold_none = sum(1 for u in units if not g[u["case_id"]]["covered"])
        cost = sum(x.get("cost_usd") or 0 for x in scores if x.get("type") in ("answer", "failure") and x.get("cost_known"))
        rows = tune_rows(r)
        per[r] = {"units_total": len(allu), "window": len(units), "audit": sum(1 for u in allu if u.get("set") == "audit"),
                  "controls": sum(1 for u in allu if u.get("control")), "prs": len({u["pr_key"] for u in units}),
                  "labelled_cases": len(x2d.read_jsonl(rd(r, "gold.jsonl"))), "sessions_used": sess["used"], "session_cap": sess["cap"],
                  "covered_pos_none_neg_covered": [gold_none, len(units) - gold_none],
                  "which_area_pos_neg": [sum(x["label"] for x in rows), len(rows) - sum(x["label"] for x in rows)],
                  "jev_requests": sum(1 for x in scores if x.get("type") == "answer"), "jev_failures": sum(1 for x in scores if x.get("type") == "failure"),
                  "jev_cost_usd": round(cost, 6)}
    res_rounds = {"per_round": per, "jev_cost_usd_total": round(sum(p["jev_cost_usd"] for p in per.values()), 6),
                  "sessions_total": sum(p["sessions_used"] for p in per.values())}
    json.dump(res, open(os.path.join(OUT, "composed.json"), "w"), indent=1, sort_keys=True)
    json.dump(res_rounds, open(os.path.join(OUT, "rounds.json"), "w"), indent=1, sort_keys=True)
    for name, d in res["settings"].items():
        for pname, f in d.items():
            print(name, "|", pname, "|", {k: (round(v, 3) if isinstance(v, float) else v) for k, v in f.items()})
    print(json.dumps(res_rounds, sort_keys=True))


def tune_rows(r):
    import tune
    return tune.round_rows(r, DATA, "dev")["merge.which_area"]


if __name__ == "__main__":
    main()
