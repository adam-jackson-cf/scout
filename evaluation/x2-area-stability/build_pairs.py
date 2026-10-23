#!/usr/bin/env python3
"""Build the labelling units from the replay output (PREREGISTRATION section 5).

Greedy best-first one-to-one matching means the matches at threshold t are exactly the matches at T_MIN whose Jaccard >= t, so the
pairs to label are the matches at T_MIN. This script verifies that property on the recorded data and writes:
  <out>/blinded.jsonl      {pair_id, state}   evidence only (what labellers see): no IDs, scores, arm, transition, class or repository name.
                           Real pairs and known-answer control pairs are mixed and carry no marker. The file paths in the evidence can
                           identify the repository and the mutation markers (`x2_newpkg/`, `*_x2moved/`), and file overlap is inferable from
                           the evidence; both are by design (rules R2/R3 need the files).
  <out>/key.jsonl          {pair_id, arm, repo, transition, class, sub, prev_feature_id, next_feature_id, jaccard, left_is_prev}  unblinded join
  <out>/control-key.jsonl  {pair_id, kind, expected, source_pair_id, arm, repo}   sealed key of the control pairs (section 5)
Controls (deterministic, hash-ranked): ceil(8 %) of the unique real pairs are content-swapped (one side's title, summary and entrypoints are
replaced by those of an unrelated feature of the same map; same files; expected different_concern by rule R3) and ceil(4 %) are
identical-record pairs (expected same_concern).
pair_id = sha256 of the whole blinded evidence in a canonical (side-order independent) form, so equal evidence is labelled once and
different evidence (including entrypoints) never shares a label.
Usage: build_pairs.py --data DIR --replay replay.jsonl --out DIR
"""
import argparse
import hashlib
import json
import math
import os
import sys

import common as C

CAP_FILES, CAP_SHARED, CAP_ENTRY = 40, 20, 8  # files and shared files are truncated alphabetically (section 5)


def side(rec):
    files = C.owned_files(rec)
    entry = []
    for e in rec.get("entrypoints") or []:
        entry.append(e.get("path", "") + (":" + e["symbol"] if e.get("symbol") else ""))
    return {"title": rec.get("title", ""), "summary": rec.get("summary", ""), "kind": rec.get("kind", ""),
            "entrypoints": entry[:CAP_ENTRY], "owned_files": files[:CAP_FILES], "owned_files_total": len(files)}, set(files)


def canonical(a, b, shared, n_shared):
    sides = sorted(json.dumps(x, sort_keys=True) for x in (a, b))
    return json.dumps({"sides": sides, "shared_files_total": n_shared, "shared_files": shared}, sort_keys=True)


def build_state(rec_a, rec_b):
    """-> (pair_id, blinded state, left_is_a). The two sides are first put in canonical order (by their serialised evidence), then the pair id's hash
    decides which is shown on the left, so the position carries no information and a byte-identical pair seen in either orientation gets the same
    state; `left_is_a` is computed per call from that canonical order. The Jaccard is reconstructible from the totals the state keeps
    (owned_files_total per side, shared_files_total; kept so truncation is visible), by design (section 5)."""
    a, fa = side(rec_a)
    b, fb = side(rec_b)
    shared = sorted(fa & fb)
    pid = hashlib.sha256(canonical(a, b, shared[:CAP_SHARED], len(shared)).encode()).hexdigest()[:20]
    a_first = json.dumps(a, sort_keys=True) <= json.dumps(b, sort_keys=True)
    first, second = (a, b) if a_first else (b, a)
    swap = int(hashlib.sha256(("order:" + pid).encode()).hexdigest(), 16) % 2 == 1
    left, right = (second, first) if swap else (first, second)
    return pid, {"left": left, "right": right, "shared_files_total": len(shared), "shared_files": shared[:CAP_SHARED]}, a_first != swap


def rank(pids, tag):
    return sorted(pids, key=lambda p: hashlib.sha256((tag + p).encode()).hexdigest())


def swapped_control(src, maps_feats):
    """Replace one side's title/summary/entrypoints by those of an unrelated feature of the same map. src: (prev_rec, next_rec, prev_map, next_map)."""
    pid = src["pid"]
    use_next = int(hashlib.sha256(("ctl-side:" + pid).encode()).hexdigest(), 16) % 2 == 1
    rec = src["next"] if use_next else src["prev"]
    pool = [f for f in maps_feats[src["next_map" if use_next else "prev_map"]]
            if f.get("source") == C.AGENT_SOURCE and f["featureId"] != rec["featureId"] and C.owned_files(f)
            and not (set(C.owned_files(f)) & set(C.owned_files(rec))) and f.get("title") != rec.get("title")]
    if not pool:
        return None
    donor = min(pool, key=lambda f: hashlib.sha256(("ctl-donor:" + pid + f["featureId"]).encode()).hexdigest())
    mod = dict(rec, title=donor.get("title", ""), summary=donor.get("summary", ""), entrypoints=donor.get("entrypoints") or [])
    return (src["prev"], mod) if use_next else (mod, src["next"])


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default=C.DATA)
    ap.add_argument("--replay", required=True)
    ap.add_argument("--out", required=True)
    a = ap.parse_args()
    for old in C.read_jsonl(os.path.join(a.out, "labels.jsonl")):
        if old.get("guide_version") == C.GUIDE_VERSION:
            sys.exit("refusing to rebuild pairs: %s already has labels for %s (labels are bound to the pair ids; PREREGISTRATION section 6)" % (a.out, C.GUIDE_VERSION))
    rows = C.read_jsonl(a.replay)
    by_tr = {}
    for r in rows:
        by_tr.setdefault((r["arm"], r["repo"], r["transition"]), {})[r["thresholdPct"]] = r
    maps_of = {t: (p, n) for t, _c, _s, p, n in C.TRANSITIONS}
    real, key, sources = {}, [], {}
    feats_cache = {}
    for (arm, repo, tr), by_t in sorted(by_tr.items()):
        base = by_t.get(C.GRID_PCT[0])
        if base is None or base["status"] != "ok":
            continue
        base_pairs = {(m["prevFeatureId"], m["nextFeatureId"]): m["jaccard"] for m in base["matches"]}
        for pct, r in by_t.items():
            if r["status"] != "ok":
                sys.exit("inconsistent status across thresholds for %s/%s/%s" % (arm, repo, tr))
            want = {k for k, j in base_pairs.items() if j >= pct / 100.0}
            got = {(m["prevFeatureId"], m["nextFeatureId"]) for m in r["matches"]}
            if want != got:
                sys.exit("prefix property violated for %s/%s/%s at %d%%: matches(t) != matches(T_MIN) with J>=t" % (arm, repo, tr, pct))
        pm, nm = maps_of[tr]
        prev = {r_["featureId"]: r_ for r_ in C.load_active_features(a.data, arm, repo, pm)}
        nxt = {r_["featureId"]: r_ for r_ in C.load_active_features(a.data, arm, repo, nm)}
        feats_cache[(arm, repo, pm)], feats_cache[(arm, repo, nm)] = list(prev.values()), list(nxt.values())
        for (pf, nf), j in sorted(base_pairs.items()):
            pid, state, left_is_prev = build_state(prev[pf], nxt[nf])
            if pid not in real:
                real[pid] = state
                sources[pid] = {"pid": pid, "prev": prev[pf], "next": nxt[nf], "prev_map": (arm, repo, pm), "next_map": (arm, repo, nm), "arm": arm, "repo": repo}
            key.append({"pair_id": pid, "arm": arm, "repo": repo, "transition": tr, "class": base["class"], "sub": base.get("sub", ""),
                        "prev_feature_id": pf, "next_feature_id": nf, "jaccard": j, "left_is_prev": left_is_prev})
    blinded, ctl = dict(real), []
    n = len(real)
    n_swap, n_ident = math.ceil(C.CONTROL_SWAP_FRACTION * n), math.ceil(C.CONTROL_IDENT_FRACTION * n)
    swapped_from = set()
    for pid in rank(list(real), "ctl-swap:"):
        if len(swapped_from) >= n_swap:
            break
        s = sources[pid]
        pair = swapped_control(s, feats_cache)
        if pair is None:
            continue
        cid, state, _ = build_state(*pair)
        if cid in blinded:
            continue
        blinded[cid] = state
        swapped_from.add(pid)
        ctl.append({"pair_id": cid, "kind": "swapped", "expected": "different_concern", "source_pair_id": pid, "arm": s["arm"], "repo": s["repo"]})
    idents = 0
    for pid in rank([p for p in real if p not in swapped_from], "ctl-ident:"):
        if idents >= n_ident:
            break
        s = sources[pid]
        cid, state, _ = build_state(s["prev"], s["prev"])
        if cid in blinded:
            continue
        blinded[cid] = state
        idents += 1
        ctl.append({"pair_id": cid, "kind": "identical", "expected": "same_concern", "source_pair_id": pid, "arm": s["arm"], "repo": s["repo"]})
    os.makedirs(a.out, exist_ok=True)
    with open(os.path.join(a.out, "blinded.jsonl"), "w") as f:
        for pid in sorted(blinded):
            f.write(json.dumps({"pair_id": pid, "state": blinded[pid]}) + "\n")
    with open(os.path.join(a.out, "key.jsonl"), "w") as f:
        for k in key:
            f.write(json.dumps(k) + "\n")
    with open(os.path.join(a.out, "control-key.jsonl"), "w") as f:
        for c in ctl:
            f.write(json.dumps(c) + "\n")
    n_sw = len(ctl) - idents
    print("pairs: %d unique real, %d transition occurrences, %d control pairs (%d swapped, %d identical)" % (n, len(key), len(ctl), n_sw, idents))
    for kind, want, got in (("swapped", n_swap, n_sw), ("identical", n_ident, idents)):
        if got < want:
            print("SHORTFALL: %d %s control pair(s) generated of %d wanted (no donor or collision); the control gate reports this" % (got, kind, want))


if __name__ == "__main__":
    main()
