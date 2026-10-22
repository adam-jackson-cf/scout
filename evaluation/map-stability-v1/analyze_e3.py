#!/usr/bin/env python3
"""Compute E3 metrics from raw/<repo>/*.json; prints markdown tables and writes metrics.json."""
import json, os, itertools
from collections import Counter
D = os.path.dirname(os.path.abspath(__file__)) + "/raw"
out = {}


def L(repo, n):
    p = f"{D}/{repo}/{n}.json"
    return json.load(open(p)) if os.path.exists(p) else None


def ids(r, src=None):
    return {f["featureId"] for f in r["snapshot"] if src is None or (f["source"] == "agent-mapper") == (src == "agent")}


def owner(r):
    m = {}
    for f in r["snapshot"]:
        for o in f["ownedFiles"]:
            m.setdefault(o, set()).add(f["featureId"])
    return m


def jac(a, b):
    return len(a & b) / len(a | b) if a | b else 1.0


def cov(r, tracked):
    own = set(owner(r))
    return len(own & set(tracked)), len(tracked)


def active(r):  # features with a non-stale status
    return [f for f in r["snapshot"] if f["status"] != "stale"]


for repo, ext in (("fastapi", ".py"), ("axum", ".rs")):
    R = {n: L(repo, n) for n in ("H1", "H2", "H3", "A1", "A2", "A3", "A1_rerun1", "A1_rerun2", "M1_content", "M2_struct")}
    files = json.load(open(f"{D}/{repo}/files.json"))
    tracked = files["tracked_source"]
    print(f"\n## {repo}\n")
    print("| run | rc | s | total | heuristic-src | agent-mapper | stale-status | result(new/changed/stale) | inventory |")
    print("|---|---|---|---|---|---|---|---|---|")
    for n, r in R.items():
        if not r:
            continue
        s = r["snapshot"] or []
        c = Counter("agent" if f["source"] == "agent-mapper" else "heur" for f in s)
        st = sum(1 for f in s if f["status"] == "stale")
        res = r.get("result", {})
        print(f"| {n} | {r['rc']} | {r['seconds']} | {len(s)} | {c['heur']} | {c['agent']} | {st} | {res.get('new')}/{res.get('changed')}/{res.get('stale')} | {r['inventory']} |")
    m = {}
    # heuristic stability
    m["heur_jaccard"] = [jac(ids(R['H1']), ids(R[b])) for b in ("H2", "H3")]
    # fresh-state agent
    pairs = [("A1", "A2"), ("A1", "A3"), ("A2", "A3")]
    m["fresh_all_jaccard"] = {f"{a}-{b}": round(jac(ids(R[a]), ids(R[b])), 4) for a, b in pairs}
    m["fresh_agent_only_jaccard"] = {f"{a}-{b}": round(jac(ids(R[a], 'agent'), ids(R[b], 'agent')), 4) for a, b in pairs}
    m["fresh_agent_n"] = {n: len(ids(R[n], 'agent')) for n in ("A1", "A2", "A3")}
    m["fresh_agent_survival_A1_in"] = {b: round(len(ids(R['A1'], 'agent') & ids(R[b], 'agent')) / max(1, len(ids(R['A1'], 'agent'))), 4) for b in ("A2", "A3")}
    # same state
    chain = ["A1", "A1_rerun1", "A1_rerun2"]
    m["same_all_jaccard"] = {f"{a}-{b}": round(jac(ids(R[a]), ids(R[b])), 4) for a, b in zip(chain, chain[1:])}
    m["same_agent_only_jaccard"] = {f"{a}-{b}": round(jac(ids(R[a], 'agent'), ids(R[b], 'agent')), 4) for a, b in zip(chain, chain[1:])}
    m["same_agent_survival"] = {f"{a}->{b}": round(len(ids(R[a], 'agent') & ids(R[b], 'agent')) / max(1, len(ids(R[a], 'agent'))), 4) for a, b in zip(chain, chain[1:])}
    # coverage
    m["coverage"] = {n: cov(R[n], tracked) for n in ("H1", "A1", "A2", "A3", "A1_rerun2")}
    # file ownership change
    def own_sets(r):
        return {f: frozenset(v) for f, v in owner(r).items()}
    def moved(a, b):
        oa, ob = own_sets(R[a]), own_sets(R[b])
        common = set(oa) & set(ob)
        return {"files_owned_both": len(common), "owner_set_changed": sum(1 for f in common if oa[f] != ob[f]),
                "only_in_first": len(set(oa) - set(ob)), "only_in_second": len(set(ob) - set(oa))}
    m["ownership_change_fresh"] = {f"{a}-{b}": moved(a, b) for a, b in pairs}
    m["ownership_change_same"] = {f"{a}->{b}": moved(a, b) for a, b in zip(chain, chain[1:])}
    # which agent ids change in same-state: classify
    # mutations
    if R["M1_content"]:
        cc = json.load(open(f"{D}/{repo}/content_change.json"))
        base = R["A1_rerun2"]
        for tag, n in (("content", "M1_content"), ("struct", "M2_struct")):
            if not R[n] or not R[n]["snapshot"]:
                continue
            prev = R["A1_rerun2"] if tag == "content" else R["M1_content"]
            P = {f["featureId"]: f for f in prev["snapshot"]}
            N = {f["featureId"]: f for f in R[n]["snapshot"]}
            gone = [P[i] for i in P if i not in N]
            new = [N[i] for i in N if i not in P]
            stalef = [N[i] for i in N if N[i]["status"] == "stale" and P[i]["status"] != "stale"] if True else []
            m[f"mut_{tag}"] = {"prev": len(P), "now": len(N), "removed_from_state": [(f['featureId'], f['source'], f['title']) for f in gone][:40],
                               "new": [(f['featureId'], f['source'], f['title'], f['ownedFiles'][:3]) for f in new][:40],
                               "newly_stale_status": [(f['featureId'], f['source']) for f in stalef][:40],
                               "survival_all": round(len(set(P) & set(N)) / len(P), 4),
                               "agent_survival": round(len({i for i in P if P[i]['source'] == 'agent-mapper'} & set(N)) / max(1, sum(1 for i in P if P[i]['source'] == 'agent-mapper')), 4),
                               "heur_survival": round(len({i for i in P if P[i]['source'] != 'agent-mapper'} & set(N)) / max(1, sum(1 for i in P if P[i]['source'] != 'agent-mapper')), 4),
                               "result": R[n]["result"]}
        m["content_change"] = cc
    out[repo] = m
    print()
    print(json.dumps(m, indent=1, default=list)[:6000])
json.dump(out, open(os.path.dirname(D) + "/metrics.json", "w"), indent=1, default=list)
