"""Usage: python3 analyze.py raw/jev_x.jsonl [more...]  -> prints markdown tables + writes metrics.json"""
import sys, json, math, statistics as st
cases = {c["case_id"]: c for c in map(json.loads, open("cases.jsonl"))}
LO, HI = 0.3, 0.7  # noul routing band: p<=LO false_positive, p>=HI real_defect, else cannot_tell

def wilson(k, n, z=1.96):
    if n == 0: return (float("nan"),) * 2
    p = k / n; d = 1 + z * z / n
    c = (p + z * z / (2 * n)) / d; h = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / d
    return (round(c - h, 3), round(c + h, 3))

def auroc(pairs):  # (score, is_pos)
    pos = [s for s, y in pairs if y]; neg = [s for s, y in pairs if not y]
    if not pos or not neg: return float("nan")
    w = sum((p > n) + 0.5 * (p == n) for p in pos for n in neg)
    return round(w / (len(pos) * len(neg)), 3)

def pct(xs, q):
    xs = sorted(xs); return xs[min(len(xs) - 1, round(q * (len(xs) - 1)))]

def pred_of(rec):
    a = next(iter(rec["answers"].values()))
    if a["type"] == "choice": return a["choice"], a["probabilities"].get("real_defect", 0.0), a["confidence"]
    p = a["noul"]; return ("real_defect" if p >= HI else "false_positive" if p <= LO else "cannot_tell"), p, abs(p - .5) * 2

def report(path):
    recs = [json.loads(l) for l in open(path)]
    ok = [r for r in recs if r["ok"]]
    out = {"file": path, "variant": recs[0]["variant"], "requested": recs[0]["model_requested"],
           "resolved": sorted({r.get("model_resolved") for r in ok}), "calls": len(recs), "failed": len(recs) - len(ok)}
    out["cost_per_1k_usd"] = round(1000 * sum(r["usage"]["cost"] for r in ok) / len(ok), 4)
    lat = [r["latency_ms"] for r in ok]; out["latency_ms_p50"] = pct(lat, .5); out["latency_ms_p95"] = pct(lat, .95)
    out["splits"] = {}
    for split in ("all", "tuning", "heldout"):
        by = {}
        for r in ok:
            if split == "all" or cases[r["case_id"]]["split"] == split: by.setdefault(r["case_id"], []).append(pred_of(r))
        # case-level: mean P(real_defect); label = majority of rep labels (tie -> cannot_tell)
        rows = []; agree = 0
        for cid, ps in by.items():
            labs = [p[0] for p in ps]; maj = max(set(labs), key=labs.count)
            if labs.count(maj) <= len(labs) / 2: maj = "cannot_tell"
            agree += len(set(labs)) == 1
            rows.append((cid, cases[cid]["label"], maj, sum(p[1] for p in ps) / len(ps)))
        n = len(rows); dec = [r for r in rows if r[2] != "cannot_tell"]
        correct = sum(r[1] == r[2] for r in rows); dcorrect = sum(r[1] == r[2] for r in dec)
        conf = {}
        for _, y, p, _ in rows: conf[f"{y}->{p}"] = conf.get(f"{y}->{p}", 0) + 1
        rep_acc = [sum(cases[cid]["label"] == p[0] for p in ps) / len(ps) for cid, ps in by.items()]
        out["splits"][split] = {"n": n, "pos": sum(r[1] == "real_defect" for r in rows),
            "acc_all(cannot_tell=wrong)": f"{correct}/{n}", "wilson95": wilson(correct, n),
            "coverage(non-cannot_tell)": f"{len(dec)}/{n}", "acc_on_decided": f"{dcorrect}/{len(dec)}",
            "wilson95_decided": wilson(dcorrect, len(dec)), "auroc_meanP": auroc([(r[3], r[1] == "real_defect") for r in rows]),
            "rep_agreement": f"{agree}/{n}", "confusion": conf, "mean_rep_acc": round(sum(rep_acc) / n, 3),
            "wrong_cases": [r[0] for r in rows if r[1] != r[2]]}
    return out

if __name__ == "__main__":
    allm = [report(p) for p in sys.argv[1:]]
    json.dump(allm, open("metrics.json", "w"), indent=1)
    for m in allm: print(json.dumps(m, indent=1))
