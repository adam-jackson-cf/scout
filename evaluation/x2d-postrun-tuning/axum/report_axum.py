"""Axum replication report (FROZEN-AXUM.md): three arms (Jev-only, Jev+Terra, B2) pooled over all rounds, per round and on the units merged on or after 2025-01-01 (the originally
requested window, descriptive), with Wilson 95 % intervals; per-round counts, Jev cost, sessions, escalation and Terra figures. Model-free; reads what escalate.py wrote.

  X2D_DATA=<axum data root> <venv python> report_axum.py <fork pilot dir> FIRST-LAST      ->  report.md and report.json next to this file
"""
import json, os, sys

HERE = os.path.dirname(os.path.abspath(__file__))
S, RNG = os.path.abspath(sys.argv[1]), sys.argv[2]
sys.argv = [sys.argv[0], S, "--prompt", os.path.join(HERE, "..", "terra_prompt_v1.txt"), "--tag", "axum-r" + RNG, "--rounds", RNG]
sys.path.insert(0, HERE)
import escalate as E  # noqa: E402  (parses the argv above; main() is not run)

cw, A, x2d = E.cw, E.A, E.x2d
ROUNDS = E.ROUNDS
data = {r: cw.load(r) for r in ROUNDS}
jev = E.jev_rows(data)
esc = [c for c, (r, u, row) in jev.items() if row["output"] in ("none", "unresolved")]
done = {}
for row in (x2d.read_jsonl(E.CALLS) if os.path.exists(E.CALLS) else []):
    if row["prompt_sha256"] == E.PSHA and (row["case_id"] not in done or row.get("parsed") is not None):
        done[row["case_id"]] = row
fin = {c: E.final_row(c, jev[c][2], done.get(c)) for c in jev}
ARMS = (("Jev-only", lambda r, u: jev[u["case_id"]][2]), ("Jev+Terra", lambda r, u: fin[u["case_id"]]), ("B2", lambda r, u: data[r][3][u["case_id"]]))
PARTS = [("all rounds %s" % RNG, lambda r, u: True)] + [("round %d" % k, (lambda kk: lambda r, u: r == kk)(k)) for k in ROUNDS] + \
        [("merged >= 2025-01-01", lambda r, u: u["date"][:10] >= "2025-01-01"), ("merged < 2025-01-01", lambda r, u: u["date"][:10] < "2025-01-01")]


def f(v):
    return "n/a" if v is None else "%.3f" % v


def cnt(r):
    if not r["n"]:
        return "n/a (0/0)"
    lo, hi = r["wilson95"]
    return "%s [%s, %s] (%d/%d)" % (f(r["value"]), f(lo), f(hi), r["k"], r["n"])


res = {}
for part, pred in PARTS:
    for arm, fn in ARMS:
        outs = []
        for r in ROUNDS:
            g, units, scores, b2 = data[r]
            outs += [A.score_outcome(fn(r, u), g[u["case_id"]]) for u in units if pred(r, u)]
        if not outs:
            continue
        c = A.counts_for(outs)
        rt = A.rates(c)
        res[(part, arm)] = {"n": rt["n"], "assigned": rt["assigned"], "none": rt["none"], "unresolved": rt["unresolved"], "gold_covered": rt["gold_covered"], "gold_none": rt["gold_none"],
                            "rates": {k: rt[k] for k in ("P_assign", "R_assign_res", "R_assign", "unresolved_rate", "P_new", "R_new")},
                            "gold_covered_answered_none": sum(1 for o in outs if o["kind"] == "none_wrong")}
json.dump({"%s | %s" % k: v for k, v in res.items()}, open(os.path.join(HERE, "report.json"), "w"), indent=1, default=str)

md = ["| part | arm | n | assigned | P_assign [Wilson 95 %] | R_assign_res | R_assign (abstention = miss) | unresolved | none P | none R | gold none | covered answered none |",
      "|---|---|---|---|---|---|---|---|---|---|---|---|"]
for (part, arm), v in res.items():
    r = v["rates"]
    md.append("| %s | %s | %d | %d | %s | %s | %s | %s | %s | %s | %d | %d |" % (part, arm, v["n"], v["assigned"], cnt(r["P_assign"]), cnt(r["R_assign_res"]), cnt(r["R_assign"]),
                                                                       cnt(r["unresolved_rate"]), cnt(r["P_new"]), cnt(r["R_new"]), v["gold_none"], v["gold_covered_answered_none"]))
open(os.path.join(HERE, "report.md"), "w").write("\n".join(md) + "\n")
print("\n".join(md))
print()
for r in ROUNDS:
    g, units, scores, b2 = data[r]
    sess = json.load(open(cw.rd(r, "sessions.json")))
    rows = x2d.read_jsonl(cw.rd(r, "jev", "scores.jsonl"))
    cost = sum(d.get("cost_usd") or 0 for d in rows if d.get("cost_known"))
    nreq = sum(1 for d in rows if "case_id" in d)
    ngc = sum(1 for u in units if g[u["case_id"]]["covered"])
    print("round %d units %d PRs %d gold none/covered %d/%d sessions %d/%d jev requests %d cost %.6f first %s last %s" % (
        r, len(units), len({u["pr_key"] for u in units}), len(units) - ngc, ngc, sess["used"], sess["cap"], nreq, cost, min(u["date"] for u in units)[:10], max(u["date"] for u in units)[:10]))
calls = [done[c] for c in esc if c in done]
lat = sorted(c["latency_s"] for c in calls if c.get("latency_s") is not None)
nunits = sum(len(data[r][1]) for r in ROUNDS)
print("escalations %d of %d (%.1f per 100); terra failures %d; models %s efforts %s; latency n=%d median %s p90 %s max %s total %s" % (
    len(esc), nunits, 100 * len(esc) / nunits, sum(1 for c in esc if (done.get(c) or {}).get("parsed") is None), sorted({c.get("model") for c in calls}),
    sorted({c.get("effort") for c in calls}), len(lat), lat[len(lat) // 2] if lat else None, lat[int(0.9 * (len(lat) - 1))] if lat else None, lat[-1] if lat else None, round(sum(lat), 1)))
rp = os.path.join(E.EDIR, E.A_.tag + ".report.json")
if os.path.exists(rp):
    print("terra correct/escalated", json.load(open(rp))["escalated_correct"])
