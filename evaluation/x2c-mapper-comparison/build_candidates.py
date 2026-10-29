#!/usr/bin/env python3
"""Build the X2c candidate pairs: the units both matching methods are judged on (X2c PREREGISTRATION section 5).

X2's `build_pairs.py` labels only the matches the Jaccard rule makes at T_MIN. The Jev method can match pairs the Jaccard rule never makes (a redrawn area that keeps
two of its ten files), so every method is judged on the same, larger candidate set: for every measured transition of every arm, **every pair (prior area, next agent
feature) that shares at least one owned file**. Prior areas are the prev map's agent features with >= 1 owned file and next features the next map's (the replay's own
definition, PREREGISTRATION section 3.1 of X2). A Jaccard match at any threshold >= T_MIN shares files, so the candidate set contains every Jaccard match; this script
verifies that on the recorded replay (every T_MIN match is a candidate, with the replay's Jaccard) and aborts otherwise.

It writes the same three files `build_pairs.py` writes, so X2's frozen `label.py` and `analysis.py` run on them unchanged, plus the Jev input:
  <out>/blinded.jsonl, key.jsonl (one row per candidate occurrence; `jaccard` is the file-set Jaccard, 0 < J <= 1), control-key.jsonl: as X2 (state, side order, pair_id, controls
  from the same hash ranking: ceil(8 %) content-swapped, ceil(4 %) identical-record, of the UNIQUE real candidate pairs);
  <out>/jev-input.jsonl: {pair_id, state} of the real candidate pairs only (controls are never sent to Jev).
Usage: build_candidates.py --data DIR --replay replay.jsonl --out DIR
"""
import argparse
import json
import math
import os
import sys

import x2c  # noqa: E402 (puts the frozen X2 directory on sys.path)
import common as C  # noqa: E402
import build_pairs as B  # noqa: E402 (X2's frozen build_pairs: states, ids, controls)


def agent_areas(feats):
    return {f["featureId"]: f for f in feats if f.get("source") == C.AGENT_SOURCE and C.owned_files(f)}


def candidates(prev, nxt):
    """[(prev id, next id, jaccard, shared count)] of every pair sharing >= 1 owned file, sorted by (prev id, next id)."""
    out = []
    pf = {k: set(C.owned_files(f)) for k, f in prev.items()}
    nf = {k: set(C.owned_files(f)) for k, f in nxt.items()}
    for p in sorted(pf):
        for n in sorted(nf):
            s = len(pf[p] & nf[n])
            if s:
                out.append((p, n, s / len(pf[p] | nf[n]), s))
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default=x2c.DATA)
    ap.add_argument("--replay", required=True)
    ap.add_argument("--out", required=True)
    a = ap.parse_args()
    bad = x2c.verify_pins()
    if bad:
        sys.exit("refusing: frozen X2 files differ from their pinned digests: %s" % "; ".join(bad))
    for old in C.read_jsonl(os.path.join(a.out, "labels.jsonl")):
        if old.get("guide_version") == C.GUIDE_VERSION:
            sys.exit("refusing to rebuild pairs: %s already has labels for %s (labels are bound to the pair ids)" % (a.out, C.GUIDE_VERSION))
    rows = C.read_jsonl(a.replay)
    base_rows = {}
    for r in rows:
        if r["thresholdPct"] == C.GRID_PCT[0]:
            base_rows[(r["arm"], r["repo"], r["transition"])] = r
    maps_of = {t: (p, n) for t, _c, _s, p, n in C.TRANSITIONS}
    real, key, sources, feats_cache = {}, [], {}, {}
    for (arm, repo, tr), base in sorted(base_rows.items()):
        if base["status"] != "ok":
            continue
        pm, nm = maps_of[tr]
        prev_all, next_all = C.load_active_features(a.data, arm, repo, pm), C.load_active_features(a.data, arm, repo, nm)
        prev, nxt = agent_areas(prev_all), agent_areas(next_all)
        feats_cache[(arm, repo, pm)], feats_cache[(arm, repo, nm)] = list(prev_all), list(next_all)
        cand = candidates(prev, nxt)
        cand_j = {(p, n): j for p, n, j, _s in cand}
        for m in base["matches"]:  # every Jaccard match must be a candidate, with the replay's Jaccard
            j = cand_j.get((m["prevFeatureId"], m["nextFeatureId"]))
            if j is None or abs(j - m["jaccard"]) > 1e-9:
                sys.exit("replay match %s/%s/%s %s->%s is not a candidate or its Jaccard differs (%r vs %r): the candidate definition and the replay disagree" %
                         (arm, repo, tr, m["prevFeatureId"], m["nextFeatureId"], j, m["jaccard"]))
        for p, n, j, s in cand:
            pid, state, left_is_prev = B.build_state(prev[p], nxt[n])
            if pid not in real:
                real[pid] = state
                sources[pid] = {"pid": pid, "prev": prev[p], "next": nxt[n], "prev_map": (arm, repo, pm), "next_map": (arm, repo, nm), "arm": arm, "repo": repo}
            key.append({"pair_id": pid, "arm": arm, "repo": repo, "transition": tr, "class": base["class"], "sub": base.get("sub", ""),
                        "prev_feature_id": p, "next_feature_id": n, "jaccard": j, "shared": s, "left_is_prev": left_is_prev})
    blinded, ctl = dict(real), []
    n = len(real)
    n_swap, n_ident = math.ceil(C.CONTROL_SWAP_FRACTION * n), math.ceil(C.CONTROL_IDENT_FRACTION * n)
    swapped_from = set()
    for pid in B.rank(list(real), "ctl-swap:"):
        if len(swapped_from) >= n_swap:
            break
        s = sources[pid]
        pair = B.swapped_control(s, feats_cache)
        if pair is None:
            continue
        cid, state, _ = B.build_state(*pair)
        if cid in blinded:
            continue
        blinded[cid] = state
        swapped_from.add(pid)
        ctl.append({"pair_id": cid, "kind": "swapped", "expected": "different_concern", "source_pair_id": pid, "arm": s["arm"], "repo": s["repo"]})
    idents = 0
    for pid in B.rank([p for p in real if p not in swapped_from], "ctl-ident:"):
        if idents >= n_ident:
            break
        s = sources[pid]
        cid, state, _ = B.build_state(s["prev"], s["prev"])
        if cid in blinded:
            continue
        blinded[cid] = state
        idents += 1
        ctl.append({"pair_id": cid, "kind": "identical", "expected": "same_concern", "source_pair_id": pid, "arm": s["arm"], "repo": s["repo"]})
    os.makedirs(a.out, exist_ok=True)
    for name, items in (("blinded.jsonl", [{"pair_id": p, "state": blinded[p]} for p in sorted(blinded)]), ("key.jsonl", key), ("control-key.jsonl", ctl),
                        ("jev-input.jsonl", [{"pair_id": p, "state": real[p]} for p in sorted(real)])):
        with open(os.path.join(a.out, name), "w") as f:
            for it in items:
                f.write(json.dumps(it) + "\n")
    n_sw = len(ctl) - idents
    print("candidates: %d unique real pairs, %d occurrences over %d measured transitions, %d control pairs (%d swapped, %d identical)" % (n, len(key), len({(k['arm'], k['repo'], k['transition']) for k in key}), len(ctl), n_sw, idents))
    for kind, want, got in (("swapped", n_swap, n_sw), ("identical", n_ident, idents)):
        if got < want:
            print("SHORTFALL: %d %s control pair(s) generated of %d wanted (no donor or collision); the control gate reports this" % (got, kind, want))


if __name__ == "__main__":
    main()
