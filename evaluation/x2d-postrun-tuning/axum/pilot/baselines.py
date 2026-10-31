"""X2d directory baselines B0/B1/B2 and contrast-B2 (PREREGISTRATION 6).

CLI (run AFTER gold exists for rounds >= 1: `run.sh round` order is units -> label -> labels-check -> gold -> baselines -> jev -> tune -> analysis):
    baselines.py --round R --repo CLONE [--data D]            writes D/rounds/R/arms/{B0,B1,B2,b2_round_start}.jsonl for every row of units.jsonl.
                                                              Rounds >= 1: B0/B1/B2 are computed against the membership at each unit's PARENT COMMIT = the round-start membership
                                                              + gold of window units of earlier commits of the round + inherit ops of earlier commits (`gold.replay_memberships`,
                                                              R1-08); `b2_round_start` = B2 against the round-start membership (reported sensitivity; arm name "b2_round_start").
                                                              Round 0: the D/P units and the controls, each against its own catalogue from control-key.jsonl (b2_round_start = B2 on
                                                              the same membership), plus D/rounds/0/arms-n1/B2.jsonl for the D units against catalogue-n1.json.
    baselines.py --round R --tag contrast --repo CLONE        writes arms/contrast-B2.jsonl against contrast-catalogue.json membership (static, no teacher forcing) from
                                                              units.contrast.jsonl: EVERY round (round 0 = the 20 D units; rounds 1..10 = the round's window units, R1-13; rounds 1 and 10 also the 20 D units, R2-06).
Importable: b0, b1, b2, arm_row, run_round(data, r, repo) -> {arm: rows} (no file written).
Trees are listed once per unit `parent` commit (`git ls-tree`, cached). Append-only (refuses when an output exists). X2D_SKIP_SIGNOFF=1 skips the sign-off check.

B0 = majority owner of the OWNED siblings in the same normalised directory (unique maximum, owned files counted once per owning area); a tie -> unresolved
(reason "tie"); no owned sibling -> none. B1 = B0, then (only when B0 found no owned sibling) walk <= 2 parent directory levels to the first directory with owned
files directly in it (root excluded); tie -> unresolved; else none. B2 = B0; if B0 does not assign (none or tie), the mirror rule (b2.py); else B1 (R1-29).
"""
import argparse, collections, json, os, sys

import x2d
import frontend as F
import gold as G
from b2 import b2_rule


def _decide(counter):
    if not counter:
        return ("none", None, None)
    top = counter.most_common()
    if len(top) > 1 and top[0][1] == top[1][1]:
        return ("unresolved", None, "tie")
    return ("assigned", top[0][0], None)


def b0(unit_members, tree_paths, ownership):
    """-> (output, area_id|None, unresolved_reason|None)."""
    mem = set(unit_members)
    tp = tree_paths if isinstance(tree_paths, (set, frozenset, dict)) else set(tree_paths)
    votes, seen = collections.Counter(), set()
    for m in sorted(mem):
        for s in ownership.sibling_owners(m, tp):
            k = (s["path"], s["area_id"])
            if s["path"] in mem or k in seen:
                continue
            seen.add(k)
            votes[s["area_id"]] += 1
    return _decide(votes)


def b1(unit_members, tree_paths, ownership):
    r0 = b0(unit_members, tree_paths, ownership)
    if r0[0] != "none":
        return r0
    tp = tree_paths if isinstance(tree_paths, (set, frozenset, dict)) else set(tree_paths)
    mem = set(unit_members)
    d = F.norm_dir(F.split_name(sorted(mem)[0])[0])
    for _ in range(2):
        if "/" not in d:
            break
        d = d.rsplit("/", 1)[0]
        votes = collections.Counter()
        for f in ownership.dir_files(d):
            if f in tp and f not in mem:
                for a in ownership.owners(f):
                    votes[a] += 1
        if votes:
            return _decide(votes)
    return ("none", None, None)


def b2(unit_members, tree_paths, ownership):
    r0 = b0(unit_members, tree_paths, ownership)
    if r0[0] == "assigned":
        return r0
    m = b2_rule(unit_members, tree_paths, ownership)
    if m:
        return ("assigned", m, None)
    return b1(unit_members, tree_paths, ownership)


def arm_row(case_id, arm, res):
    return {"case_id": case_id, "arm": arm, "output": res[0], "area_id": res[1], "confidence": None, "unresolved_reason": res[2]}


class _Trees:
    def __init__(self, repo):
        self.repo, self.c = repo, {}

    def get(self, rev):
        if rev not in self.c:
            self.c[rev] = set(F.tree_paths(self.repo, rev))
        return self.c[rev]


def _own(membership_path_or_obj):
    return F.Ownership(membership_path_or_obj)


def _cat_membership(path):
    return G.catalogue_membership(json.load(open(path)), 0)


def _run(units, owner_for, trees, arms):
    """owner_for(unit, arm) -> Ownership."""
    out = {a: [] for a in arms}
    fn = {"B0": b0, "B1": b1, "B2": b2, "b2_round_start": b2, "contrast-B2": b2}
    for u in units:
        t = trees.get(u["parent"])
        for a in arms:
            out[a].append(arm_row(u["case_id"], a, fn[a](u["members"], t, owner_for(u, a))))
    return out


def run_round(data, r, repo, trees=None):
    """-> {"B0","B1","B2","b2_round_start": rows} (+ {"n1-B2": rows} in round 0) for round r's units.jsonl; no file is written. Rounds >= 1 need gold.jsonl (R1-08)."""
    trees = trees or _Trees(repo)
    rd = x2d.round_dir(r, data)
    units = x2d.read_jsonl(os.path.join(rd, "units.jsonl"))
    arms = ["B0", "B1", "B2", "b2_round_start"]
    if r == 0:
        base = os.path.join(data or x2d.DATA, "catalogue", "catalogue.json")
        paths = {"catalogue": base}
        for n in ("catalogue-p1", "catalogue-p2", "catalogue-n1"):
            paths[n] = os.path.join(rd, n + ".json")
        cache = {}

        def get(name):
            if name not in cache:
                cache[name] = _own(_cat_membership(paths[name]))
            return cache[name]

        key = {k["case_id"]: k.get("catalogue", "catalogue") for k in x2d.read_jsonl(os.path.join(rd, "control-key.jsonl"))}
        res = _run(units, lambda u, a: get(key.get(u["case_id"], "catalogue")), trees, arms)
        dp_units = [u for u in units if u["set"] in ("D", "P") and not u.get("control")]
        res["n1-B2"] = _run(dp_units, lambda u, a: get("catalogue-n1"), trees, ["B2"])["B2"]
        return res
    if not os.path.exists(os.path.join(rd, "gold.jsonl")):
        sys.exit("refusing: round %d gold.jsonl missing; baselines are replayed per parent commit after gold (R1-08)" % r)
    try:
        mems = G.replay_memberships(r, data)
    except ValueError as e:  # a window unit without usable gold (R2-04): the stage refuses with the cases named
        sys.exit("refusing: round %d baselines need gold for every window unit: %s" % (r, e))
    start = _own(G.read_membership(r, data))
    cache = {}

    def owner_for(u, arm):
        if arm == "b2_round_start":
            return start
        m = mems[u["case_id"]]
        if id(m) not in cache:
            cache[id(m)] = _own(m)
        return cache[id(m)]

    return _run(units, owner_for, trees, arms)


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--round", type=int, required=True)
    ap.add_argument("--tag", default=None, choices=[None, "contrast"])
    ap.add_argument("--repo", required=True)
    ap.add_argument("--data", default=None)
    a = ap.parse_args(argv)
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pins: " + "; ".join(bad))
    r, data = a.round, a.data
    rd = x2d.round_dir(r, data)
    trees = _Trees(a.repo)
    if a.tag == "contrast":
        p = os.path.join(rd, "units.contrast.jsonl")
        units = x2d.read_jsonl(p)
        if not units:
            sys.exit("refusing: %s missing or empty (run build_units --tag contrast first; round 0: build_controls)" % p)
        own = _own(_cat_membership(os.path.join(data or x2d.DATA, "catalogue", "contrast-catalogue.json")))
        dest = os.path.join(rd, "arms", "contrast-B2.jsonl")
        x2d.refuse_if_exists(dest)
        res = _run(units, lambda u, a_: own, trees, ["contrast-B2"])
        x2d.write_jsonl_new(dest, res["contrast-B2"])
        print("wrote", dest, len(units))
        return
    arms = ["B0", "B1", "B2", "b2_round_start"]
    dests = {k: os.path.join(rd, "arms", k + ".jsonl") for k in arms}
    n1_dest = os.path.join(rd, "arms-n1", "B2.jsonl")
    x2d.refuse_if_exists(*dests.values(), *([n1_dest] if r == 0 else []))
    res = run_round(data, r, a.repo, trees)
    if r == 0:
        x2d.write_jsonl_new(n1_dest, res["n1-B2"])
    for k in arms:
        x2d.write_jsonl_new(dests[k], res[k])
    print("wrote arms for %d units" % len(res["B0"]))


if __name__ == "__main__":
    main()
