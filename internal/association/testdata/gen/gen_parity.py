#!/usr/bin/env python3
"""Regenerates internal/association/testdata/parity-<repo>.json.gz for the association parity test (about 2 minutes).

Read-only over the recorded X2d pilot inputs; writes only the two output files. Standard library only.

    python3 internal/association/testdata/gen/gen_parity.py \
        [--x2d     /Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot] \
        [--scout   .]     # this repository: axum pilot front end under evaluation/x2d-postrun-tuning/axum/pilot
        [--fastapi /tmp/scout-x2c-smoke/_src/fastapi] [--axum /tmp/scout-x2c-smoke/_src/axum]

Sources: fastapi window rounds 1-10 (<x2d>/data/postrun, <x2d>/prepass/rounds.json, <x2d>/frontend.py) and axum rounds 1-5
(<x2d>/data/postrun-axum, the axum pilot under --scout). The axum clone is the one in evaluation/x2d-postrun-tuning/axum/FROZEN-AXUM.md.

Per round: the round-start membership restricted to the paths that matter (touched paths plus owned variant siblings of added
paths); per first-parent commit that exercises a rule: the raw changes, the parent-tree siblings of added paths, and
  pilot     the recorded pilot front end (owned_classify + clusters + cap), asserted equal to the recorded units.jsonl ids;
  expected  the same front end with the two rules Scout does not port neutralised (cross-scope rename: scope() constant;
            translation: _LANG_DIR never matches) = the design's Scout result;
  affected  the rules for which pilot and expected differ ([] = parity case).
Commits that only modify owned files, delete files or touch nothing are not stored (counted in commits_total).
"""
import argparse, collections, gzip, hashlib, importlib.util, json, os, re

PR_CAP = 10
NEVER = re.compile(r"(?!)")


def load(name, path):
    spec = importlib.util.spec_from_file_location(name, path)
    m = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(m)
    return m


def cap_key(cid):
    return hashlib.sha256(("x2d-cap\0" + cid).encode()).hexdigest()


def run_front(F, entries, state, own, blobs, commit, neutralise=()):
    """Pilot front end for one commit with the named rules neutralised ("scope", "lang") -> comparable dict."""
    saved = (F.scope, F._LANG_DIR)
    if "scope" in neutralise:
        F.scope = lambda p: "x"
    if "lang" in neutralise:
        F._LANG_DIR = NEVER
    try:
        resolved = []
        counts, residual = F.owned_classify(entries, state, own, resolved)
        live, inert = F.clusters(residual, blobs)
    finally:
        F.scope, F._LANG_DIR = saved
    info = {p: d for p, _k, d in residual}
    for c in live:
        c["id"] = F.cluster_id(commit, c)
    order = sorted(live, key=lambda c: cap_key(c["id"]))
    counts = dict(counts, inert_clusters=len(inert), residual_clusters=len(live))
    return {
        "counts": {k: v for k, v in sorted(counts.items()) if v},
        "clusters": [{"id": c["id"], "dir": c["dir"], "stem": c["stem"], "files": c["files"], "kinds": c["kinds"],
                      "heuristic": [f for f in c["files"] if "heuristic_only" in info[f]]} for c in live],
        "inert": [{"dir": c["dir"], "stem": c["stem"], "files": c["files"]} for c in inert],
        "decided": [c["id"] for c in order[:PR_CAP]],
        "capped": [c["id"] for c in order[PR_CAP:]],
        "resolved": sorted(({"path": x["path"], "rule": x["rule"], "owners": x["owners"], "old": x["old"], "pool": x["pool"]} for x in resolved), key=lambda x: x["path"]),
    }


def build_repo(name, F, repo, data, rounds_info, rounds):
    out = {"repo": name, "rounds": []}
    for r in rounds:
        info = next(x for x in rounds_info["rounds"] if x["round"] == r)
        rd = os.path.join(data, "rounds", str(r))
        membership = json.load(open(os.path.join(rd, "membership.json")))
        own = F.Ownership(membership)
        recorded, capped_rec = collections.defaultdict(list), collections.defaultdict(list)
        for line in open(os.path.join(rd, "units.jsonl")):
            u = json.loads(line)
            if u.get("set") == "window" and not u.get("control"):
                recorded[u["commit"]].append(u["cluster_id"])
        for line in open(os.path.join(rd, "capped.jsonl")):
            if line.strip():
                row = json.loads(line)
                capped_rec[row["pr"]].append(row["cluster_id"])
        state = F.TreeState(F.tree_paths(repo, info["after_commit"]))
        commits, touched, added_paths, total = [], set(), [], 0
        for sha, parent, date, subject in F.first_parent_commits(repo, info["after_commit"], info["through_commit"]):
            total += 1
            entries = F.diff_entries(repo, parent, sha)
            blobs = {e[3]: e[4] for e in entries if e[0] in "ARC"}
            pilot = run_front(F, entries, state, own, blobs, sha)
            expected = run_front(F, entries, state, own, blobs, sha, ("scope", "lang"))
            assert sorted(recorded.get(sha, [])) == sorted(pilot["decided"]), (name, r, sha, "recorded units differ from the pilot run")
            pr = str(F.pr_number(subject) or sha[:8])
            assert sorted(capped_rec.get(pr, [])) == sorted(pilot["capped"]), (name, r, sha, "recorded capped differ from the pilot run")
            affected = [rule for rule, n in (("cross_scope_rename", ("scope",)), ("translation", ("lang",)))
                        if run_front(F, entries, state, own, blobs, sha, n) != pilot]
            if pilot != expected and not affected:
                affected = ["cross_scope_rename+translation"]
            interesting = affected or pilot != expected or expected["clusters"] or expected["inert"] or expected["resolved"] \
                or set(expected["counts"]) - {"modified_owned", "deleted"}
            if interesting:
                added = [e[3] for e in entries if e[0] in "AC"]
                commits.append({"sha": sha, "parent": parent,
                                "entries": [[e[0], e[2], e[3], 1 if e[4] == F.EMPTY_BLOB else 0] for e in entries],
                                "tree": sorted({f for p in added for f in own._by_sib.get(F.sibling_key(p), []) if f != p and f in state.paths}),
                                "pilot": pilot, "expected": expected, "affected": affected})
                for e in entries:
                    touched.update((e[2], e[3]))
                added_paths += added
            state.apply(entries)
        keep = set(touched)
        for p in added_paths:
            keep.update(own._by_sib.get(F.sibling_key(p), []))
        heur, pool = set(membership.get("heuristic_only", [])), set(membership.get("pool", []))
        out["rounds"].append({"round": r, "commits_total": total, "commits": commits,
                              "owners": {f: sorted(own._owners[f]) for f in sorted(keep) if f in own._owners},
                              "heuristic_only": sorted(touched & heur), "pool": sorted(touched & pool)})
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--x2d", default="/Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot")
    ap.add_argument("--scout", default=".")
    ap.add_argument("--fastapi", default="/tmp/scout-x2c-smoke/_src/fastapi")
    ap.add_argument("--axum", default="/tmp/scout-x2c-smoke/_src/axum")
    ap.add_argument("--out", default=os.path.join(os.path.dirname(os.path.abspath(__file__)), ".."))
    a = ap.parse_args()
    axum_pilot = os.path.join(a.scout, "evaluation/x2d-postrun-tuning/axum/pilot")
    jobs = [
        ("fastapi", load("frontend_fastapi", os.path.join(a.x2d, "frontend.py")), a.fastapi, os.path.join(a.x2d, "data/postrun"), os.path.join(a.x2d, "prepass/rounds.json"), range(1, 11)),
        ("axum", load("frontend_axum", os.path.join(axum_pilot, "frontend.py")), a.axum, os.path.join(a.x2d, "data/postrun-axum"), os.path.join(axum_pilot, "prepass/rounds.json"), range(1, 6)),
    ]
    for name, F, repo, data, rj, rounds in jobs:
        res = build_repo(name, F, repo, data, json.load(open(rj)), rounds)
        path = os.path.join(a.out, "parity-%s.json.gz" % name)
        with gzip.GzipFile(path, "wb", mtime=0) as f:
            f.write(json.dumps(res, sort_keys=True, separators=(",", ":")).encode())
        print(name, "commits stored", sum(len(r["commits"]) for r in res["rounds"]), "of", sum(r["commits_total"] for r in res["rounds"]), "bytes", os.path.getsize(path))


if __name__ == "__main__":
    main()
