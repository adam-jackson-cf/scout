import json, collections, sys
from analyze import pred_of
cases = {c["case_id"]: c for c in map(json.loads, open("unlabelled.jsonl"))}
out = {}
for f in sys.argv[1:]:
    recs = [json.loads(l) for l in open(f) if l.strip()]
    ok = [r for r in recs if r["ok"]]
    pool = collections.Counter(pred_of(r)[0] for r in ok)
    bysrc = collections.defaultdict(collections.Counter)
    for r in ok: bysrc[cases[r["case_id"]]["state"]["tool_signal"].split(" rule ")[0]][pred_of(r)[0]] += 1
    out[f] = {"calls": len(recs), "ok": len(ok), "pooled": dict(pool), "by_source": {k: dict(v) for k, v in bysrc.items()},
              "cost_per_1k_usd": round(1000 * sum(r["usage"]["cost"] for r in ok) / len(ok), 4)}
json.dump(out, open("unlabelled_metrics.json", "w"), indent=1)
print(json.dumps(out, indent=1))
