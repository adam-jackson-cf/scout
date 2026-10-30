"""X2d gold, membership layers (teacher forcing, PREREGISTRATION 10) and free-running membership.

CLI:
    gold.py --round R [--tag contrast] [--data D]
        writes D/rounds/R/gold[.<tag>].jsonl from the final `kind:"wide"` label rows (provenance agent_reviewed, agreement agreed|adjudicated, first usable final row per
        case; `control_kind` joined from D/rounds/0/control-key.jsonl (round 0) or the units' `control_kind` (rounds >= 1)) and, without --tag and for R < 10, the NEXT round's start membership
        D/rounds/<R+1>/membership.json = round R's start membership + gold primaries of round R's window units (gold `none` -> pool; nothing else is minted) + the
        deterministic rule effects of round R (`inherit.jsonl`: deletions leave, (c) renames move, (d) variant siblings join).
Append-only: refuses when an output exists. Gold and next membership are computed in memory first; a window unit without usable gold makes `main` exit writing NOTHING (R2-04), naming the cases and
`label_x2d --round R --new-round <name>`; `_gold_outputs`/`next_membership`/`replay_memberships` raise ValueError for it. X2D_SKIP_SIGNOFF=1 -> verify_pins(require_signoff=False) (validation).

Importable: load_gold, build_gold, catalogue_membership, round_membership, next_membership, replay_memberships, derive_ablation, free_membership, divergence, pool_areas.
Absorption is chronological (R1-11): events are grouped by `seq` (the commit's index in the round, written into units.jsonl and inherit.jsonl by build_units); a commit's
inherit ops are applied before the gold of its own window units. Controls and audit units are never absorbed. A pooled file renamed by rule (c) stays pooled (R1-04).
Replayed free-running divergence (R1-12): `divergence(next_membership(r), free_membership(r, arm))` (gold membership vs the arm's own, files-Jaccard per area).
`replay_memberships(r)` -> {case_id: membership at the unit's parent commit} for the baselines (R1-08).
Departures recorded here: a covered gold row with no primary but exactly one acceptable id is absorbed into that id; with several acceptable ids and no primary it is
not absorbed (listed under `unabsorbed` in membership.json; label_x2d.agree now always picks a primary from the common ids, R1-11). Pool areas (leaf directories with >= 3 pooled files) are reported only, never an option.
"""
import argparse, collections, json, os, sys

import x2d
import frontend as F


def _rd(r, data=None):
    return x2d.round_dir(r, data)


def _labels_path(r, data, tag):
    return os.path.join(_rd(r, data), "labels.%s.jsonl" % tag if tag else "labels.jsonl")


def _gold_path(r, data, tag):
    return os.path.join(_rd(r, data), "gold.%s.jsonl" % tag if tag else "gold.jsonl")


def control_kinds(data=None, r=0):
    """{case_id: kind}: round 0 from control-key.jsonl (labellers and Jev never see it); rounds >= 1 from the `control_kind` of the units.jsonl rows (per-round controls)."""
    if r == 0:
        return {k["case_id"]: k.get("kind") for k in x2d.read_jsonl(os.path.join(_rd(0, data), "control-key.jsonl"))}
    return {u["case_id"]: u.get("control_kind") for u in x2d.read_jsonl(os.path.join(_rd(r, data), "units.jsonl")) if u.get("control")}


def build_gold(r, data=None, tag=None, write=True):
    """Write gold(.<tag>).jsonl; returns the rows. Append-only. Round 0, no tag: the gold row of a P2 clone is `derive_ablation(source gold, ablated area)` (the planted-gap
    gate, R1-07); a P2 clone whose source has no agreed gold gets no row (unmeasured). Rev 13 (R-053, the re-gate root only): a P2 control-key row with `gold_rule: "direct_label"` keeps the clone's OWN
    agreed label as its gold (gap and new alike) and carries the derived gold beside it as `derived_gold`."""
    out = _gold_path(r, data, tag)
    x2d.refuse_if_exists(out)
    ck = control_kinds(data, r) if not tag else {}
    guide = x2d.guide_version(r, data)
    last = {}
    for row in x2d.read_jsonl(_labels_path(r, data, tag)):
        if row.get("kind") != "wide" or row.get("provenance") != "agent_reviewed" or row.get("agreement") not in ("agreed", "adjudicated"):
            continue
        if row.get("guide_version") != guide:  # only labels under this round's guide (R2-01); other rounds' gold is never rewritten
            continue
        ls = row.get("label_set")
        if tag == "contrast":
            if ls != "contrast":
                continue
        elif ls not in (None, "main"):
            continue
        last[row["case_id"]] = row  # the LAST usable final row per case wins
    rows = []
    for cid, row in last.items():
        rows.append({"case_id": cid, "covered": bool(row.get("covered")), "acceptable": list(row.get("acceptable") or []), "primary": row.get("primary"),
                     "none_kind": row.get("none_kind"), "agreement": row["agreement"], "control_kind": ck.get(cid)})
    if r == 0 and not tag:
        by = {g["case_id"]: g for g in rows}
        key = {k["case_id"]: k for k in x2d.read_jsonl(os.path.join(_rd(0, data), "control-key.jsonl")) if k.get("kind") == "p2"}
        direct = {cid for cid, k in key.items() if k.get("gold_rule") == "direct_label"}  # rev 13 (R-053): the clone's OWN agreed label on its designated catalogue is its gold
        rows = [g for g in rows if g["case_id"] not in key or g["case_id"] in direct]
        for cid, k in key.items():
            src = by.get(k.get("source_case_id"))
            if cid in direct:
                own = by.get(cid)
                if own is not None and src is not None:  # reported beside the direct gold, never scored: the gold the rev-12 derivation would have given (derive_ablation of the source's label on the full catalogue)
                    d = derive_ablation(src, k.get("ablated_area_ids") or [k["ablated_area_id"]])
                    own["gold_rule"] = "direct_label"
                    own["derived_gold"] = {"covered": d["covered"], "acceptable": d["acceptable"], "primary": d["primary"], "none_kind": d["none_kind"], "source_case_id": src["case_id"],
                                           "source_agreement": src["agreement"]}
                elif own is not None:
                    own["gold_rule"] = "direct_label"
                continue
            if src is not None:
                d = derive_ablation(src, k.get("ablated_area_ids") or [k["ablated_area_id"]])
                d.update(case_id=cid, control_kind="p2", agreement=src["agreement"], derived_from=src["case_id"])
                rows.append(d)
    rows.sort(key=lambda g: g["case_id"])
    if write:
        x2d.write_jsonl_new(out, rows)
    return rows


def load_gold(r, data=None, tag=None):
    p = _gold_path(r, data, tag)
    if not os.path.exists(p):
        raise FileNotFoundError("gold file missing: " + p)
    return {g["case_id"]: g for g in x2d.read_jsonl(p)}


def derive_ablation(gold_row, ablated_set):
    """P2: acceptable minus the ablated areas (the SAME set the designated catalogue removes: control-key `ablated_area_ids`, R2-05); if empty -> covered False,
    none_kind gap (primary cleared when it was ablated). `ablated_set`: iterable of area ids (a single id string is accepted)."""
    ab = {ablated_set} if isinstance(ablated_set, str) else set(ablated_set)
    g = dict(gold_row)
    acc = [a for a in (g.get("acceptable") or []) if a not in ab]
    g["acceptable"] = acc
    if g.get("primary") in ab:
        g["primary"] = None
    if not acc:
        g["covered"], g["primary"], g["none_kind"] = False, None, "gap"
    return g


# ---- membership layers
def catalogue_membership(cat, round_no=1):
    return {"round": round_no, "areas": {a["area_id"]: sorted(a["files"]) for a in cat["areas"]}, "pool": [],
            "heuristic_only": sorted(h["path"] if isinstance(h, dict) else h for h in cat.get("heuristic_only", [])), "pool_areas": {}, "unabsorbed": []}


def pool_areas(membership):
    """Leaf directories (normalised) holding >= 3 pooled files: reported, never a Choice option."""
    by = collections.defaultdict(list)
    for f in membership.get("pool", []):
        by[F.norm_dir(F.split_name(f)[0])].append(f)
    return {d: sorted(v) for d, v in sorted(by.items()) if len(v) >= 3}


def _finish(areas, pool, heur, round_no, unabsorbed):
    m = {"round": round_no, "areas": {a: sorted(v) for a, v in sorted(areas.items())}, "pool": sorted(pool), "heuristic_only": sorted(heur), "unabsorbed": sorted(unabsorbed)}
    m["pool_areas"] = pool_areas(m)
    return m


def read_membership(r, data=None):
    p = os.path.join(_rd(r, data), "membership.json")
    with open(p) as f:
        return json.load(f)


def round_membership(r, data=None):
    """Round r's start membership: round 1 = the catalogue; r > 1 = gold.next_membership(r-1) (an existing membership.json is returned as is by `read_membership`)."""
    if r == 1:
        return catalogue_membership(x2d.read_catalogue(data), 1)
    return next_membership(r - 1, data)


def _inherit_rows(r, data=None):
    return x2d.read_jsonl(os.path.join(_rd(r, data), "inherit.jsonl"))


class _Layer:
    """Mutable membership layer advanced event by event (R1-11): a commit's inherit ops, then the gold of that commit's window units."""

    def __init__(self, membership, recorded_owners):
        self.areas = {a: set(v) for a, v in membership["areas"].items()}
        self.pool = set(membership.get("pool", []))
        self.heur = membership.get("heuristic_only", [])
        self.recorded = recorded_owners
        self.unabs = []

    def _drop(self, f):
        for s in self.areas.values():
            s.discard(f)
        self.pool.discard(f)

    def op(self, op):
        if op["op"] == "remove":
            self._drop(op["path"])
        elif op["op"] == "move":
            own = [a for a, s in self.areas.items() if op["old"] in s]
            was_pooled = op["old"] in self.pool
            if not own and self.recorded and not was_pooled:
                own = [a for a in op.get("area_ids", []) if a in self.areas]
            self._drop(op["old"])
            for a in own:
                self.areas[a].add(op["path"])
            if not own and was_pooled:  # a pooled file stays pooled when renamed
                self.pool.add(op["path"])
        elif op["op"] == "add":
            for a in op.get("area_ids", []):
                if a in self.areas:
                    self.areas[a].add(op["path"])

    def unit(self, u, o):
        if o is None:
            self.unabs.extend(u["members"])
        elif o == "none":
            for f in u["members"]:
                self.pool.add(f)
        elif o in self.areas:
            for f in u["members"]:
                self.pool.discard(f)
                self.areas[o].add(f)
        else:
            self.unabs.extend(u["members"])

    def finish(self, round_no):
        return _finish(self.areas, self.pool, self.heur, round_no, self.unabs)


def _events(units, inherit):
    """-> {seq: ([ops], [window units])}; `seq` = index of the commit in the round's first-parent order (absent -> 0). Only window units are absorbed (controls, audit: never)."""
    ev = collections.defaultdict(lambda: ([], []))
    for op in inherit:
        ev[op.get("seq", 0)][0].append(op)
    for u in units:
        if u.get("set") == "window":
            ev[u.get("seq", 0)][1].append(u)
    return ev


def _absorb(membership, units, outputs, inherit, round_no, recorded_owners):
    """Chronological absorption: commits in round order; a commit's inherit ops are applied before the gold of its own units (R1-11)."""
    layer = _Layer(membership, recorded_owners)
    ev = _events(units, inherit)
    for s in sorted(ev):
        ops, us = ev[s]
        for op in ops:
            layer.op(op)
        for u in us:
            layer.unit(u, outputs.get(u["case_id"]))
    return layer.finish(round_no)


def replay_memberships(r, data=None):
    """{case_id: membership} for every row of round r's units.jsonl (window, audit, controls): the membership at the row's PARENT commit = the round-start
    membership + the gold of window units of earlier commits of the round + the inherit ops of earlier commits (`_absorb` applied chronologically, R1-08).
    Rows of one commit share one membership object. The baselines are computed against this (after gold exists); units and labels stay built against the
    round-start membership."""
    start = read_membership(r, data)
    units = x2d.read_jsonl(os.path.join(_rd(r, data), "units.jsonl"))
    outputs = _gold_outputs(units, load_gold(r, data), r)
    ev = _events(units, _inherit_rows(r, data))
    layer = _Layer(start, True)
    seqs = sorted({u.get("seq", 0) for u in units} | set(ev))
    snap, out = {}, {}
    for s in seqs:
        snap[s] = layer.finish(r)  # state before commit s: everything strictly earlier
        ops, us = ev.get(s, ([], []))
        for op in ops:
            layer.op(op)
        for u in us:
            layer.unit(u, outputs.get(u["case_id"]))
    for u in units:
        out[u["case_id"]] = snap[u.get("seq", 0)]
    return out


def _gold_outputs(units, gold, r=None):
    """{case_id: output} of the window units; raises ValueError naming the cases when a window unit has no usable gold (R2-04)."""
    out, missing = {}, []
    for u in units:
        if u.get("set") != "window":
            continue
        g = gold.get(u["case_id"])
        if g is None:
            missing.append(u["case_id"])
        elif not g["covered"]:
            out[u["case_id"]] = "none"
        elif g.get("primary"):
            out[u["case_id"]] = g["primary"]
        elif len(g.get("acceptable") or []) == 1:
            out[u["case_id"]] = g["acceptable"][0]
        else:
            out[u["case_id"]] = None
    if missing:
        raise ValueError("gold missing for %d window units (%s%s): label them with `label_x2d --round %s --new-round <name>` and rerun" % (
            len(missing), ", ".join(missing[:10]), ", ..." if len(missing) > 10 else "", "r" if r is None else r))
    return out


def next_membership(r, data=None, gold=None):
    """Membership at the END of round r under gold (= round r+1's start): round r's start membership + gold primaries + rule effects. `gold` (optional, {case_id: row}) defaults to the gold.jsonl file. Raises ValueError when a window unit has no gold."""
    start = read_membership(r, data)
    units = x2d.read_jsonl(os.path.join(_rd(r, data), "units.jsonl"))
    gold = load_gold(r, data) if gold is None else gold
    return _absorb(start, units, _gold_outputs(units, gold, r), _inherit_rows(r, data), r + 1, True)


def _arm_outputs(r, arm, data):
    p = os.path.join(_rd(r, data), "arms", arm + ".jsonl")
    if not os.path.exists(p):
        raise FileNotFoundError("arm file missing: " + p)
    out = {}
    for row in x2d.read_jsonl(p):
        out[row["case_id"]] = row["area_id"] if row["output"] == "assigned" else ("none" if row["output"] == "none" else None)
    return out


def free_membership(r, arm, data=None):
    """Free-running membership of `arm` at the end of round r (secondary analysis, PREREGISTRATION 10): starts from the catalogue and absorbs the arm's own
    assigned (-> area) / none (-> pool) outputs round by round; unresolved are not absorbed. Rule effects (inherit.jsonl) are replayed on the arm's own
    membership (a (c) move carries the arm's owners of the old path; (d) adds use the recorded owners)."""
    m = round_membership(1, data)
    for k in range(1, r + 1):
        units = x2d.read_jsonl(os.path.join(_rd(k, data), "units.jsonl"))
        m = _absorb(m, units, _arm_outputs(k, arm, data), _inherit_rows(k, data), k + 1, False)
    return m


def divergence(membership_gold, membership_arm):
    """File-set Jaccard per area (gold vs arm membership) and the pool; areas absent from both are skipped."""
    per = {}
    for a in sorted(set(membership_gold["areas"]) | set(membership_arm["areas"])):
        g, b = set(membership_gold["areas"].get(a, [])), set(membership_arm["areas"].get(a, []))
        if g or b:
            per[a] = len(g & b) / len(g | b)
    gp, bp = set(membership_gold.get("pool", [])), set(membership_arm.get("pool", []))
    pj = len(gp & bp) / len(gp | bp) if gp | bp else 1.0
    vals = list(per.values())
    return {"per_area": per, "mean": sum(vals) / len(vals) if vals else 1.0, "min": min(vals) if vals else 1.0, "pool": pj,
            "areas_diverged": sum(1 for v in vals if v < 1.0)}


def required_gold_ids(r, data=None, tag=None):
    """R3-03, R4-07: the case ids a gold file must cover. Main, r >= 1: the window units. Main, round 0: the D and P units and EVERY non-P2 control of units.jsonl, labelled, attempted or never
    started (P2 gold derives from its source, a D/P unit). A never-started capped control has neither a label row nor a call row, so it is required by membership of units.jsonl, never silently absent.
    Contrast tag: every unit of the labelled contrast set (round 0 units.contrast.jsonl, later rounds units.contrast_window.jsonl)."""
    rd = _rd(r, data)
    if tag == "contrast":
        return {u["case_id"] for u in x2d.read_jsonl(os.path.join(rd, "units.contrast.jsonl" if r == 0 else "units.contrast_window.jsonl"))}
    units = x2d.read_jsonl(os.path.join(rd, "units.jsonl"))
    if r >= 1:
        return {u["case_id"] for u in units if u.get("set") == "window"}
    need = {u["case_id"] for u in units if not u.get("control") and u.get("set") in ("D", "P", "F")}  # F: the fresh source units of the rev-13 re-gate root (labelled on the full catalogue)
    derived = {k["case_id"] for k in x2d.read_jsonl(os.path.join(rd, "control-key.jsonl")) if k.get("kind") == "p2" and k.get("gold_rule") != "direct_label"}  # a direct-label P2 clone needs its own gold
    return need | {u["case_id"] for u in units if u.get("control") and u["case_id"] not in derived}


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--round", type=int, required=True)
    ap.add_argument("--tag", default=None)
    ap.add_argument("--data", default=None)
    a = ap.parse_args(argv)
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pins: " + "; ".join(bad))
    r, data = a.round, a.data
    nxt = os.path.join(_rd(r + 1, data), "membership.json")
    with_next = not a.tag and 1 <= r < x2d.N_ROUNDS
    x2d.refuse_if_exists(_gold_path(r, data, a.tag), *([nxt] if with_next else []))
    rows = build_gold(r, data, a.tag, write=False)
    miss = sorted(required_gold_ids(r, data, a.tag) - {g["case_id"] for g in rows})  # R3-03: independent of with_next: every gold file is complete or not written
    if miss:
        sys.exit("refusing, nothing written: gold missing for %d unit(s) (%s%s): label them with `label_x2d --round %d%s --new-round <name>` and rerun" % (
            len(miss), ", ".join(miss[:10]), ", ..." if len(miss) > 10 else "", r, " --tag contrast" if a.tag else ""))
    m = None
    if with_next:
        try:  # everything is computed in memory first: on missing gold nothing is written (R2-04)
            m = next_membership(r, data, gold={g["case_id"]: g for g in rows})
        except ValueError as e:
            sys.exit("refusing, nothing written: %s" % e)
    x2d.write_jsonl_new(_gold_path(r, data, a.tag), rows)
    print("gold rows:", len(rows))
    if m is not None:
        x2d.write_json_new(nxt, m)
        print("wrote", nxt)


if __name__ == "__main__":
    main()
