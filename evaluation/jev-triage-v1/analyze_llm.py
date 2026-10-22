import json, sys
from analyze import wilson, pct
cases = {c["case_id"]: c for c in map(json.loads, open("cases.jsonl"))}
res = {}
for f in sys.argv[1:]:
    recs = [json.loads(l) for l in open(f)]
    o = {"calls": len(recs), "parsed": sum(1 for r in recs if r.get("label")),
         "cost_total_usd": round(sum(r.get("cost_usd") or 0 for r in recs), 4),
         "cost_per_1k_usd": round(1000 * sum(r.get("cost_usd") or 0 for r in recs) / len(recs), 2),
         "latency_ms_p50": pct([r["latency_ms"] for r in recs], .5), "latency_ms_p95": pct([r["latency_ms"] for r in recs], .95),
         "models": sorted({m for r in recs for m in r.get("model_usage", [])}), "splits": {}}
    for s in ("all", "tuning", "heldout"):
        rs = [r for r in recs if s == "all" or cases[r["case_id"]]["split"] == s]
        k = sum(r.get("label") == cases[r["case_id"]]["label"] for r in rs)
        conf = {}
        for r in rs:
            key = f'{cases[r["case_id"]]["label"]}->{r.get("label")}'; conf[key] = conf.get(key, 0) + 1
        o["splits"][s] = {"acc": f"{k}/{len(rs)}", "wilson95": wilson(k, len(rs)), "confusion": conf,
                          "wrong": [r["case_id"] for r in rs if r.get("label") != cases[r["case_id"]]["label"]]}
    res[f] = o
json.dump(res, open("llm_metrics.json", "w"), indent=1)
print(json.dumps(res, indent=1))
