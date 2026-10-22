#!/usr/bin/env python3
"""Analyze raw/<repo>/*.json from run_compat.py (plus E3 0.4.0 numbers by reference)."""
import json, os, collections, sys
BASE = os.path.dirname(os.path.abspath(__file__)) + "/raw"

def load(repo, name):
    p = f"{BASE}/{repo}/{name}.json"
    return json.load(open(p)) if os.path.exists(p) else None

def ids(r, src=None):
    return {f["featureId"] for f in r["snapshot"] if src is None or (f["source"] == "agent-mapper") == (src == "agent")}

def jac(a, b):
    return len(a & b) / len(a | b) if a | b else 1.0

for repo in ("fastapi", "axum"):
    print(f"=== {repo}")
    for n in ("H04_1", "H07_1", "H07_2", "A07_1", "A07_2"):
        r = load(repo, n)
        if not r:
            print(n, "missing"); continue
        c = collections.Counter(f["source"] for f in r["snapshot"] or [])
        print(n, "rc", r["rc"], "s", r["seconds"], "features", len(r["snapshot"] or []), dict(c), "result", r.get("result"), "inv", r.get("inventory"))
    h4, h7a, h7b = (load(repo, n) for n in ("H04_1", "H07_1", "H07_2"))
    print("H07 vs H07 Jaccard", jac(ids(h7a), ids(h7b)))
    a, b = ids(h4), ids(h7a)
    print(f"0.4.0 heuristic {len(a)} vs 0.7.3 heuristic {len(b)}: common={len(a&b)} only04={len(a-b)} only07={len(b-a)} Jaccard={jac(a,b):.3f} survival_of_04_ids={len(a&b)/len(a):.3f}")
    by04 = {f["featureId"]: f for f in h4["snapshot"]}
    by07 = {f["featureId"]: f for f in h7a["snapshot"]}
    print(" removed (0.4.0-only) by kind:", dict(collections.Counter(by04[i]["kind"] for i in a - b)))
    print(" added (0.7.3-only) by kind:", dict(collections.Counter(by07[i]["kind"] for i in b - a)))
    # for removed IDs: does the same owned-file set exist under a new ID? (file-set migration)
    sets07 = {tuple(sorted(f["ownedFiles"])): f["featureId"] for f in h7a["snapshot"]}
    mig = sum(1 for i in a - b if tuple(sorted(by04[i]["ownedFiles"])) in sets07)
    print(f" removed 0.4.0 IDs whose exact owned-file set reappears under a new 0.7.3 ID: {mig}/{len(a-b)}")
    files04 = {o for f in h4["snapshot"] for o in f["ownedFiles"]}
    files07 = {o for f in h7a["snapshot"] for o in f["ownedFiles"]}
    print(f" owned-file coverage (distinct files): 0.4.0={len(files04)} 0.7.3={len(files07)}")
    a1, a2 = load(repo, "A07_1"), load(repo, "A07_2")
    if a1 and a2 and a1["snapshot"] and a2["snapshot"]:
        x, y = ids(a1, "agent"), ids(a2, "agent")
        print(f"agent-only A1={len(x)} A2={len(y)} common={len(x&y)} Jaccard={jac(x,y):.3f} share_of_A1_in_A2={len(x&y)/len(x) if x else 0:.3f}")
        print(f"all-feature Jaccard {jac(ids(a1), ids(a2)):.3f}")
        fa = {o for f in a1["snapshot"] if f["source"] == "agent-mapper" for o in f["ownedFiles"]}
        fb = {o for f in a2["snapshot"] if f["source"] == "agent-mapper" for o in f["ownedFiles"]}
        print(f"agent-owned files A1={len(fa)} A2={len(fb)} overlap={len(fa&fb)}")
