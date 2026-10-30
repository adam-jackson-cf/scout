"""Rev-13 (R-053) fixtures: the one permitted re-gate of P1 and P2 on fresh items (PREREGISTRATION 5, rev 13; CLAUDE.md experiment rule 4: planted effect and planted null for each rebuilt gate).

Every function takes `V`, the validate module (its `check`, `run`, `clean_env`, `jl`, `wjl`, `make_inputs`, `x2d`, `A` = analysis_x2d, `G` = gold, `PY`, `S`, `ROOT`, `SRC`), so this file imports nothing from
validate.py. No model call, no secret. Check names start with `G13`.
  unit_checks(V, tmp)         pre-pipeline: the rebuilt gates on synthetic re-gate roots (planted effect / planted null, exclusions, boundaries, UNMEASURED, twins, retained clones), direct-label gold versus derived gold,
                              the leak fix, `x2d.gate_problems`, round 1 refusing without a re-gate PASS (run.sh), the frozen Jev side, the budget charge, the one-shot refusals of the stages.
  e2e_regate(V, tmp, e)       on copies of the finished round-0 fixture run (after `gates-analyse`): the fresh pool and its exclusions on the real clone, the build, the refusal on a short supply, the labelled and Jev-run
                              re-gate on stubs (planted effect PASS, path-following null FAIL), the unchanged round-0 bytes, the one-shot refusals.
"""
import collections
import copy
import hashlib
import json
import os
import shutil

PASS, FAIL, UNM = "PASS", "FAIL", "UNMEASURED"


# ---------------------------------------------------------------- the synthetic re-gate world (analysis-level fixtures)
def p1_item(stratum=1, out="assigned", area_out="A", acc=("A",), gold="ok", sib=("B",), raw=None, b1="B", b2="B", conf=0.9, reason=None):
    """One P1 item spec. `area_out` "A" = the item's own area, "B" = its sibling area, None = no area. `gold`: ok (agreed set `acc`), none (no agreed gold), notcov (agreed, not covered)."""
    return dict(stratum=stratum, out=out, area_out=area_out, acc=list(acc), gold=gold, sib=list(sib), raw=raw, b1=b1, b2=b2, conf=conf, reason=reason)


def p2_item(direct="none", kind="gap", derived="none", out="none", area=None, twin_of=None, origin="in_no_unit", acc=("R",)):
    """One P2 clone spec: its own label (`direct` none / covered / None = no agreed gold), the derived gold it is reported beside (`derived` none / covered / None), V-sem's output."""
    return dict(direct=direct, kind=kind, derived=derived, out=out, area=area, twin_of=twin_of, origin=origin, acc=list(acc))


def world(V, tmp, name, p1=(), p2=(), main_gates=("FAIL", "FAIL"), marker=True):
    """A study data directory D (rev-12 gates.json with the given P1/P2 statuses, the frozen Jev side) holding D/rounds/0/regate with the files `analysis_x2d.regate_gates` reads. -> (D, regate root)."""
    x2d, jl, wjl = V.x2d, V.jl, V.wjl
    main = os.path.join(tmp, "g13-" + name)
    shutil.rmtree(main, ignore_errors=True)
    rd0 = os.path.join(main, "rounds", "0")
    root = x2d.regate_root(main)
    rr = os.path.join(root, "rounds", "0")
    os.makedirs(os.path.join(main, "catalogue"))
    json.dump({"schema": "x2d-catalogue-v1", "areas": [{"area_id": "R", "title": "r", "summary": "r", "files": ["r.py"]}], "heuristic_only": []}, open(os.path.join(main, "catalogue", "catalogue.json"), "w"))
    V.make_inputs(main, 0)
    json.dump(dict(x2d.DEFAULT_SETTINGS), open(os.path.join(rd0, "settings.json"), "w"))
    th = {"decisions": {q: {"source": "provisional-0.7", "status": "provisional", "threshold": 0.7} for q in ("merge.area_covers", "merge.covered", "merge.which_area")}, "round": 0}
    json.dump(th, open(os.path.join(rd0, "thresholds.json"), "w"))
    json.dump({"gates": {"K2": {"status": "PASS"}, "P1": {"status": main_gates[0]}, "P2": {"status": main_gates[1]}}}, open(os.path.join(rd0, "gates.json"), "w"))
    os.makedirs(os.path.join(root, "catalogue"))
    shutil.copyfile(os.path.join(main, "catalogue", "catalogue.json"), os.path.join(root, "catalogue", "catalogue.json"))
    shutil.copytree(os.path.join(rd0, "inputs"), os.path.join(rr, "inputs"))
    for fn in ("settings.json", "thresholds.json"):
        shutil.copyfile(os.path.join(rd0, fn), os.path.join(rr, fn))
    if marker:
        json.dump({"gate_set": x2d.REGATE_GATE_SET}, open(os.path.join(root, "regate.json"), "w"))
    units, key, gold, vs1, b1, b2, vs2 = [], [], [], [], [], [], []
    for i, s in enumerate(p1):
        cid, A, B = "0:ctl-%d" % (i + 1), "A%02d" % i, "B%02d" % i
        sib = [B if x == "B" else x for x in s["sib"]]
        area = {"A": A, "B": B, None: None}.get(s["area_out"], s["area_out"])
        units.append({"case_id": cid, "control": True, "set": "ctl", "control_kind": "p1", "stratum": s["stratum"], "members": ["new/d%d/f%d.py" % (i, i)], "pr_key": "p%d" % i})
        key.append({"case_id": cid, "kind": "p1", "area_id": A, "sibling_area_ids": sib, "target_dir": "new/d%d" % i, "origin": "in_no_unit", "gold_rule": "labelled_acceptable_set_must_contain_A"})
        acc = [A if x == "A" else B if x == "B" else x for x in s["acc"]]
        if s["gold"] == "ok":
            gold.append({"case_id": cid, "covered": True, "acceptable": acc, "primary": acc[0], "none_kind": None, "agreement": "agreed"})
        elif s["gold"] == "notcov":
            gold.append({"case_id": cid, "covered": False, "acceptable": [], "primary": None, "none_kind": "gap", "agreement": "agreed"})
        raw = {"A": A, "B": B}.get(s["raw"], s["raw"]) if s["raw"] else area
        vs1.append({"case_id": cid, "arm": "V-sem", "output": s["out"], "area_id": area if s["out"] == "assigned" else None, "confidence": s["conf"], "unresolved_reason": s["reason"], "raw_choice": raw,
                    "raw_covered": True})
        for rows, nm, v in ((b1, "B1", s["b1"]), (b2, "B2", s["b2"])):
            v = {"A": A, "B": B}.get(v, v)
            rows.append({"case_id": cid, "arm": nm, "output": "assigned" if v else "none", "area_id": v})
    for i, s in enumerate(p2):
        cid = "0:ctl-%d" % (len(p1) + i + 1)
        mem = ["c0/f%d.py" % (s["twin_of"] if s["twin_of"] is not None else i)]
        units.append({"case_id": cid, "control": True, "set": "ctl", "control_kind": "p2", "stratum": 3, "members": mem, "pr_key": "q%d" % i})
        key.append({"case_id": cid, "kind": "p2", "gold_rule": "direct_label", "source_case_id": "0:src-%d" % i, "ablated_area_id": "X%d" % i, "ablated_area_ids": ["X%d" % i], "origin": s["origin"]})
        if s["direct"] is not None:
            g = {"case_id": cid, "covered": s["direct"] == "covered", "acceptable": list(s["acc"]) if s["direct"] == "covered" else [], "primary": s["acc"][0] if s["direct"] == "covered" else None,
                 "none_kind": None if s["direct"] == "covered" else s["kind"], "agreement": "agreed", "gold_rule": "direct_label"}
            if s["derived"] is not None:
                g["derived_gold"] = {"covered": s["derived"] == "covered", "acceptable": ["R"] if s["derived"] == "covered" else [], "primary": None, "none_kind": None if s["derived"] == "covered" else "gap",
                                     "source_case_id": "0:src-%d" % i, "source_agreement": "agreed"}
            gold.append(g)
        vs2.append({"case_id": cid, "arm": "V-sem", "output": s["out"], "area_id": s["area"] if s["out"] == "assigned" else None, "confidence": 0.9, "unresolved_reason": "below_t_cov" if s["out"] == "unresolved" else None,
                    "raw_choice": s["area"], "raw_covered": s["out"] != "none"})
    wjl(os.path.join(rr, "units.jsonl"), units)
    wjl(os.path.join(rr, "control-key.jsonl"), key)
    wjl(os.path.join(rr, "gold.jsonl"), gold or [{"case_id": "0:none", "covered": True, "acceptable": ["R"], "primary": "R", "none_kind": None, "agreement": "agreed"}])
    wjl(os.path.join(rr, "labels.jsonl"), [{"case_id": "0:none", "kind": "wide"}])
    wjl(os.path.join(rr, "arms-p1", "V-sem.jsonl"), vs1 or [{"case_id": "0:none", "arm": "V-sem", "output": "none"}])
    wjl(os.path.join(rr, "arms", "B1.jsonl"), b1 or [{"case_id": "0:none", "arm": "B1", "output": "none"}])
    wjl(os.path.join(rr, "arms", "B2.jsonl"), b2 or [{"case_id": "0:none", "arm": "B2", "output": "none"}])
    wjl(os.path.join(rr, "arms-p2", "V-sem.jsonl"), vs2 or [{"case_id": "0:none", "arm": "V-sem", "output": "none"}])
    for p in ("jev-p/scores.p1.jsonl", "jev-p/scores.p2.jsonl"):
        wjl(os.path.join(rr, p), [{"type": "header"}])
    return main, root


def gate(V, tmp, name, p1=(), p2=()):
    main, root = world(V, tmp, name, p1, p2)
    return V.A.regate_gates(root)


def _p1_good(n=24, **kw):
    return [p1_item(stratum=1, **kw) for _ in range(n)]


def _tail_p1(n2=6, n3=6):
    return [p1_item(stratum=2) for _ in range(n2)] + [p1_item(stratum=3, sib=()) for _ in range(n3)]


def _p2_pos(n, kinds=("gap", "new"), **kw):
    return [p2_item(direct="none", kind=kinds[i % len(kinds)], **kw) for i in range(n)]


# ---------------------------------------------------------------- P1
def p1_checks(V, tmp):
    check = V.check
    x2d = V.x2d
    # planted effect: Jev answers the item's own area A (acceptable), baselines answer the sibling area B (not acceptable) -> PASS
    g = gate(V, tmp, "p1-effect", p1=_p1_good() + _tail_p1())
    p1 = g["gates"]["P1"]
    v = p1["value"]
    check("G13 P1 planted effect: V-sem recovers A on every stratum-1 item, B1 / B2 answer the sibling area -> PASS; accuracy on the answered items 1.0, abstention 0, baselines 0, path-following 0, gate set tagged %s" % x2d.REGATE_GATE_SET,
          p1["status"] == PASS and p1["n"] == 24 and v["stratum1"]["accuracy_answered"]["value"] == 1.0 and v["stratum1"]["abstention_rate"]["value"] == 0.0 and v["baselines"]["B1"]["all_eligible"]["value"] == 0.0
          and v["baselines"]["B2"]["all_eligible"]["value"] == 0.0 and v["stratum1"]["path_following_answered"]["value"] == 0.0 and g["gate_set"] == x2d.REGATE_GATE_SET and v["items_built"] == 36
          and v["items_eligible"] == 36 and sorted(v["strata"]) == ["1", "2", "3"] and v["strata"]["2"]["n"] == 6 and v["strata"]["3"]["n"] == 6 and v["excluded"] == {"no_agreed_gold": [], "gold_not_covered": [], "gold_excludes_A": []}, p1)
    # planted null: a path-following Jev names the sibling area B (not acceptable) -> FAIL, the path-following rate is reported (answered 1.0, raw choice 1.0)
    g = gate(V, tmp, "p1-null", p1=_p1_good(area_out="B") + _tail_p1())
    p1 = g["gates"]["P1"]
    v = p1["value"]["stratum1"]
    check("G13 P1 planted null: a path-following V-sem (names sibling area B, never acceptable) FAILs; accuracy on the answered items 0.0, path-following rate 1.0 composed and 1.0 raw; strata 2-3 are reported, not decisive",
          p1["status"] == FAIL and v["accuracy_answered"]["value"] == 0.0 and v["path_following_answered"]["value"] == 1.0 and v["path_following_raw_choice"]["value"] == 1.0, p1)
    # B acceptable: an answer naming B is not path following when the labellers accept B
    g = gate(V, tmp, "p1-bacc", p1=_p1_good(area_out="B", acc=("A", "B"), b1=None, b2=None) + _tail_p1())
    v = g["gates"]["P1"]["value"]["stratum1"]
    check("G13 P1 an answer naming sibling area B is NOT path following when B is in the labellers' acceptable set (gold = the agreed set): PASS, path-following 0.0",
          g["gates"]["P1"]["status"] == PASS and v["path_following_answered"]["value"] == 0.0 and v["accuracy_answered"]["value"] == 1.0 and v["recovery_all_items_A_only"]["value"] == 0.0 and v["sibling_in_acceptable"] == 24,
          g["gates"]["P1"])
    # the accuracy bound: 16 of 20 answered stratum-1 items = 0.80 passes, 15 of 20 fails (a `none` is a resolved wrong answer)
    s16 = [p1_item(out="assigned", area_out="A") for _ in range(16)] + [p1_item(out="none", area_out=None) for _ in range(4)]
    s15 = [p1_item(out="assigned", area_out="A") for _ in range(15)] + [p1_item(out="none", area_out=None) for _ in range(5)]
    g16, g15 = gate(V, tmp, "p1-acc16", p1=s16)["gates"]["P1"], gate(V, tmp, "p1-acc15", p1=s15)["gates"]["P1"]
    check("G13 P1 accuracy bound at the R_assign minimum %.2f: 16 of 20 answered (4 false `none` = resolved wrong answers) = 0.80 -> PASS, 15 of 20 = 0.75 -> FAIL (none answers counted)" % x2d.REGATE_P1_ACC_MIN,
          g16["status"] == PASS and abs(g16["value"]["stratum1"]["accuracy_answered"]["value"] - 0.8) < 1e-12 and g15["status"] == FAIL and g15["value"]["stratum1"]["none_answers"] == 5
          and abs(g15["value"]["stratum1"]["accuracy_answered"]["value"] - 0.75) < 1e-12, (g16, g15))
    # the abstention bound: accuracy is judged on the answered items only, abstention has its own cap g = 0.15
    a3 = [p1_item(out="assigned", area_out="A") for _ in range(17)] + [p1_item(out="unresolved", area_out=None, reason="below_t_choice") for _ in range(3)]
    a4 = [p1_item(out="assigned", area_out="A") for _ in range(16)] + [p1_item(out="unresolved", area_out=None, reason="below_t_choice") for _ in range(4)]
    g3, g4 = gate(V, tmp, "p1-abst3", p1=a3)["gates"]["P1"], gate(V, tmp, "p1-abst4", p1=a4)["gates"]["P1"]
    check("G13 P1 abstention bound g = %.2f beside the accuracy on answered items: 3 of 20 abstained (0.15) with every answer right -> PASS (all-item recovery 0.85 reported); 4 of 20 abstained (0.20) with every answer right -> FAIL" % x2d.REGATE_P1_ABSTAIN_MAX,
          g3["status"] == PASS and g3["value"]["stratum1"]["accuracy_answered"]["value"] == 1.0 and abs(g3["value"]["stratum1"]["recovery_all_items"]["value"] - 0.85) < 1e-12
          and g4["status"] == FAIL and g4["value"]["stratum1"]["accuracy_answered"]["value"] == 1.0 and abs(g4["value"]["stratum1"]["abstention_rate"]["value"] - 0.2) < 1e-12, (g3, g4))
    # the directory baselines: B1 and B2 recover an acceptable area on <= 0.20 of ALL eligible items
    ok_items = lambda nb, name: [p1_item(**{name: "A"}) if i < nb else p1_item() for i in range(24)] + _tail_p1(8, 8)   # 40 eligible items
    gb8, gb9, gb2 = gate(V, tmp, "p1-b8", p1=ok_items(8, "b1")), gate(V, tmp, "p1-b9", p1=ok_items(9, "b1")), gate(V, tmp, "p1-b2-9", p1=ok_items(9, "b2"))
    check("G13 P1 baseline bar 0.20 over all eligible items (40): B1 recovering A on 8 of 40 = 0.20 -> PASS, on 9 of 40 -> FAIL; the same for B2",
          gb8["gates"]["P1"]["status"] == PASS and gb9["gates"]["P1"]["status"] == FAIL and gb2["gates"]["P1"]["status"] == FAIL and abs(gb9["gates"]["P1"]["value"]["baselines"]["B1"]["all_eligible"]["value"] - 9 / 40) < 1e-12
          and gb9["gates"]["P1"]["value"]["baselines"]["B2"]["all_eligible"]["value"] == 0.0, (gb8["gates"]["P1"]["value"]["baselines"], gb9["gates"]["P1"]["value"]["baselines"]))
    # exclusions: gold absent / not covered / not containing A are reported and excluded (not counted right or wrong); eligible stratum-1 items below 20 = UNMEASURED
    ex = [p1_item(gold="none"), p1_item(gold="none"), p1_item(gold="notcov"), p1_item(acc=("Z",)), p1_item(acc=("Z",), area_out="B")]   # 5 excluded; Jev right on 2 of them, wrong on 3
    g_un = gate(V, tmp, "p1-unm", p1=_p1_good(19) + ex + _tail_p1(2, 2))["gates"]["P1"]
    g_ok = gate(V, tmp, "p1-ex-ok", p1=_p1_good(20) + ex + _tail_p1(2, 2))["gates"]["P1"]
    check("G13 P1 exclusions: an item whose agreed gold is absent (2), not covered (1) or does not contain A (2) is reported by id and excluded (it counts as neither right nor wrong); below %d eligible stratum-1 items (19) the gate is UNMEASURED (not PASS), at 20 it decides (PASS)" % x2d.REGATE_P1_STRATUM1_MIN,
          g_un["status"] == UNM and g_un["n"] == 19 and g_ok["status"] == PASS and g_ok["n"] == 20 and len(g_ok["value"]["excluded"]["no_agreed_gold"]) == 2 and len(g_ok["value"]["excluded"]["gold_not_covered"]) == 1
          and len(g_ok["value"]["excluded"]["gold_excludes_A"]) == 2 and g_ok["value"]["items_built"] == 20 + 5 + 4 and g_ok["value"]["items_eligible"] == 24 and g_ok["value"]["stratum1_eligible"] == 20
          and g_ok["value"]["stratum1"]["accuracy_answered"]["value"] == 1.0, (g_un["status"], g_ok["value"]["excluded"]))
    # a missing arm row is a refusal, not a silent pass
    main, root = world(V, tmp, "p1-miss", p1=_p1_good(2))
    vp = os.path.join(root, "rounds", "0", "arms-p1", "V-sem.jsonl")
    V.wjl(vp, V.jl(vp)[:1])
    try:
        V.A.regate_gates(root)
        refused = False
    except SystemExit as ex_:
        refused = "lack a V-sem" in str(ex_.code)
    check("G13 P1 a control without a V-sem arm row refuses the analysis (no silent pass on a subset)", refused)


# ---------------------------------------------------------------- P2
def p2_checks(V, tmp):
    check, x2d = V.check, V.x2d
    # planted effect: 14 direct-label positives (8 gap, 6 new), V-sem answers `none` on all; 26 clones keep an area by their own label (reported): 20 assigned right, 3 none, 3 abstained
    ret = [p2_item(direct="covered", derived="none", out="assigned", area="R") for _ in range(20)] + [p2_item(direct="covered", derived="none", out="none") for _ in range(3)] + [
        p2_item(direct="covered", derived="none", out="unresolved") for _ in range(3)]
    pos = [p2_item(direct="none", kind="gap", derived="none") for _ in range(8)] + [p2_item(direct="none", kind="new", derived="covered") for _ in range(6)]
    g = gate(V, tmp, "p2-effect", p2=pos + ret)["gates"]["P2"]
    v = g["value"]
    check("G13 P2 planted effect: V-sem answers `none` on all 14 direct-label positives (gap 8 and new 6 count alike) -> PASS; none rate on the answered positives 1.0, abstention 0.0; the 26 clones whose own label keeps an area are reported "
          "(retained_covered: 20 assigned right, 3 none, 3 abstained), in no denominator; the derived gold is reported beside the direct label (8 none|none, 6 none|covered, 26 covered|none), never scored",
          g["status"] == PASS and g["n"] == 14 and v["none_rate_answered"]["value"] == 1.0 and v["abstention_rate"]["value"] == 0.0 and v["positives_by_none_kind"] == {"gap": 8, "new": 6}
          and v["retained_covered"] == {"n": 26, "answered": 23, "assigned_correct": 20, "none": 3, "assigned_wrong": 0, "abstained": 3} and v["derived_vs_direct"] == {"covered|none": 26, "none|covered": 6, "none|none": 8}
          and v["clones"] == 40 and v["twins_excluded"] == [] and v["no_agreed_gold"] == 0 and sum(v["positives_by_ablated_area"].values()) == 14 and len(v["positives_by_ablated_area"]) == 14, g)
    # planted null: a V-sem that assigns an area on the positives instead of `none`
    half = [p2_item(direct="none", out="none") for _ in range(7)] + [p2_item(direct="none", out="assigned", area="R") for _ in range(7)]
    allw = [p2_item(direct="none", out="assigned", area="R") for _ in range(14)]
    g_half, g_all = gate(V, tmp, "p2-half", p2=half)["gates"]["P2"], gate(V, tmp, "p2-all", p2=allw)["gates"]["P2"]
    check("G13 P2 planted null: a V-sem that assigns an area on the direct-label positives FAILs (7 of 14 answered `none` = 0.50; 0 of 14 = 0.0): the none rate on the answered positives is below %.2f" % x2d.REGATE_P2_NONE_MIN,
          g_half["status"] == FAIL and abs(g_half["value"]["none_rate_answered"]["value"] - 0.5) < 1e-12 and g_all["status"] == FAIL and g_all["value"]["none_rate_answered"]["value"] == 0.0
          and g_all["value"]["assigned_on_positives"] == 14, (g_half, g_all))
    # the none-rate bound 0.80 on the answered positives and the abstention bound 0.15
    r16 = [p2_item(out="none") for _ in range(16)] + [p2_item(out="assigned", area="R") for _ in range(4)]
    r15 = [p2_item(out="none") for _ in range(15)] + [p2_item(out="assigned", area="R") for _ in range(5)]
    a3 = [p2_item(out="none") for _ in range(17)] + [p2_item(out="unresolved") for _ in range(3)]
    a4 = [p2_item(out="none") for _ in range(16)] + [p2_item(out="unresolved") for _ in range(4)]
    g16, g15, ga3, ga4 = (gate(V, tmp, n, p2=s)["gates"]["P2"] for n, s in (("p2-r16", r16), ("p2-r15", r15), ("p2-a3", a3), ("p2-a4", a4)))
    check("G13 P2 bounds: none rate on the answered positives 16 of 20 = 0.80 -> PASS, 15 of 20 -> FAIL; abstention beside it: 3 of 20 abstained (0.15, every answer `none`) -> PASS, 4 of 20 (0.20) -> FAIL",
          g16["status"] == PASS and g15["status"] == FAIL and ga3["status"] == PASS and abs(ga3["value"]["abstention_rate"]["value"] - 0.15) < 1e-12 and ga3["value"]["none_rate_answered"]["value"] == 1.0
          and abs(ga3["value"]["none_rate_all_positives"]["value"] - 0.85) < 1e-12 and abs(g16["value"]["none_rate_all_positives"]["value"] - 0.8) < 1e-12
          and ga4["status"] == FAIL and ga4["value"]["none_rate_answered"]["value"] == 1.0 and abs(ga4["value"]["abstention_rate"]["value"] - 0.2) < 1e-12, (g16["status"], g15["status"], ga3["status"], ga4["status"]))
    # fewer than 8 direct-label positives: UNMEASURED (does not pass), 8: decided
    g7, g8 = gate(V, tmp, "p2-7", p2=_p2_pos(7) + [p2_item(direct="covered", out="assigned", area="R") for _ in range(30)])["gates"]["P2"], gate(V, tmp, "p2-8", p2=_p2_pos(8))["gates"]["P2"]
    check("G13 P2 below %d direct-label positives (7 of 37 clones) the gate is UNMEASURED, which does not pass; at 8 it decides (PASS)" % x2d.REGATE_P2_POSITIVES_MIN,
          g7["status"] == UNM and g7["n"] == 7 and g8["status"] == PASS and g8["n"] == 8, (g7["status"], g8["status"]))
    # a path twin of an earlier clone is one item counted twice: excluded; a clone without an agreed label is reported (no_agreed_gold) and in no denominator
    tw = _p2_pos(8) + [p2_item(direct="none", out="assigned", area="R", twin_of=0), p2_item(direct=None, out="assigned", area="R")]
    gt = gate(V, tmp, "p2-twin", p2=tw)["gates"]["P2"]
    check("G13 P2 a path twin (the same member file as an earlier clone) is excluded and listed, so a twin's wrong answer is not counted (PASS on 8 positives); a clone without an agreed label is counted under no_agreed_gold",
          gt["status"] == PASS and gt["n"] == 8 and len(gt["value"]["twins_excluded"]) == 1 and gt["value"]["no_agreed_gold"] == 1 and gt["value"]["clones"] == 10, gt)
    # a confident wrong assignment on a retained clone is reported, not decisive
    gw = gate(V, tmp, "p2-ret-wrong", p2=_p2_pos(10) + [p2_item(direct="covered", out="assigned", area="WRONG") for _ in range(3)])["gates"]["P2"]
    check("G13 P2 retained-covered clones are reported, not decisive: 3 confident wrong assignments on them leave the gate PASS and show as assigned_wrong 3",
          gw["status"] == PASS and gw["value"]["retained_covered"]["assigned_wrong"] == 3 and gw["value"]["retained_covered"]["n"] == 3, gw["value"]["retained_covered"])
    # the rev-12 FAIL is reported as it is, never re-scored: the re-gate analysis does not touch rounds/0/gates.json
    main, root = world(V, tmp, "p2-rev12", p2=_p2_pos(8))
    before = V.x2d.sha256_file(os.path.join(main, "rounds", "0", "gates.json"))
    V.A.regate_gates(root)
    check("G13 the re-gate analysis never reads or rewrites the rev-12 rounds/0/gates.json (its SHA-256 is unchanged)", V.x2d.sha256_file(os.path.join(main, "rounds", "0", "gates.json")) == before)


# ---------------------------------------------------------------- gold: direct label versus derived
def gold_checks(V, tmp):
    check, x2d, wjl = V.check, V.x2d, V.wjl
    data = os.path.join(tmp, "g13-gold")
    shutil.rmtree(data, ignore_errors=True)
    rd = os.path.join(data, "rounds", "0")
    V.make_inputs(data, 0)
    gv = x2d.guide_version(0, data)
    cat = {"areas": [{"area_id": a} for a in ("A", "B", "C")]}
    json.dump(cat, open(os.path.join(rd, "catalogue-p2.json"), "w"))
    units = [{"case_id": "0:s1", "set": "F", "control": False}, {"case_id": "0:s2", "set": "F", "control": False}, {"case_id": "0:s3", "set": "F", "control": False},
             {"case_id": "0:c1", "set": "ctl", "control": True}, {"case_id": "0:c2", "set": "ctl", "control": True}, {"case_id": "0:c3", "set": "ctl", "control": True}, {"case_id": "0:l1", "set": "ctl", "control": True}]
    wjl(os.path.join(rd, "units.jsonl"), units)
    key = [{"case_id": "0:c1", "kind": "p2", "gold_rule": "direct_label", "source_case_id": "0:s1", "ablated_area_ids": ["A"], "ablated_area_id": "A"},
           {"case_id": "0:c2", "kind": "p2", "gold_rule": "direct_label", "source_case_id": "0:s2", "ablated_area_ids": ["A"], "ablated_area_id": "A"},
           {"case_id": "0:c3", "kind": "p2", "gold_rule": "direct_label", "source_case_id": "0:s3", "ablated_area_ids": ["A"], "ablated_area_id": "A"},
           {"case_id": "0:l1", "kind": "p2", "gold_rule": "derive_ablation", "source_case_id": "0:s1", "ablated_area_ids": ["A"], "ablated_area_id": "A"}]
    wjl(os.path.join(rd, "control-key.jsonl"), key)

    def lab(cid, cov, acc, nk=None):
        return {"case_id": cid, "kind": "wide", "label_set": "main", "by": "A+B agree", "provenance": "agent_reviewed", "agreement": "agreed", "guide_version": gv, "covered": cov, "acceptable": acc,
                "primary": acc[0] if acc else None, "none_kind": nk}
    # s1: the source keeps B beside A (the rev-12 derivation would keep a covered gold on B); the clone's OWN label says none (gap): the direct label wins, the derived gold is reported beside it
    # s2: the source has only A (derived gap); the clone's own label is covered by C: direct covered C, derived gap
    # s3: the source's only area is A and the clone is `none` / new (the rev-12 derivation: gap)
    wjl(os.path.join(rd, "labels.jsonl"), [lab("0:s1", True, ["A", "B"]), lab("0:s2", True, ["A"]), lab("0:s3", True, ["A"]), lab("0:c1", False, [], "gap"), lab("0:c2", True, ["C"]), lab("0:c3", False, [], "new"),
                                           lab("0:l1", False, [], "gap")])
    rows = {g["case_id"]: g for g in V.G.build_gold(0, data, write=False)}
    d1, d2, d3, dl = rows["0:c1"], rows["0:c2"], rows["0:c3"], rows["0:l1"]
    check("G13 gold: a P2 clone whose key says `direct_label` takes its OWN agreed label as gold (gap and new alike: c1 none/gap, c2 covered by C, c3 none/new) and carries the rev-12 derived gold beside it "
          "(c1 derived covered by B, c2 derived gap, c3 derived gap), never replacing it",
          d1["covered"] is False and d1["none_kind"] == "gap" and d1["derived_gold"]["covered"] is True and d1["derived_gold"]["acceptable"] == ["B"] and d2["covered"] is True and d2["acceptable"] == ["C"]
          and d2["derived_gold"]["covered"] is False and d2["derived_gold"]["none_kind"] == "gap" and d3["covered"] is False and d3["none_kind"] == "new" and d3["derived_gold"]["none_kind"] == "gap"
          and all(r["gold_rule"] == "direct_label" and r["derived_gold"]["source_case_id"] in ("0:s1", "0:s2", "0:s3") for r in (d1, d2, d3)), (d1, d2, d3))
    check("G13 gold: a P2 clone WITHOUT `direct_label` (the rev-12 key) still gets the derived gold of its source and its own label is dropped (the rev-12 behaviour is unchanged): l1 derived covered by B from s1",
          dl["covered"] is True and dl["acceptable"] == ["B"] and dl.get("derived_from") == "0:s1" and "derived_gold" not in dl, dl)
    need = V.G.required_gold_ids(0, data=data)
    check("G13 gold: required_gold_ids covers the fresh source units (set F) and the direct-label clones (own gold), not a rev-12 derive-rule clone: {s1, s2, s3, c1, c2, c3}", need == {"0:s1", "0:s2", "0:s3", "0:c1", "0:c2", "0:c3"}, sorted(need))
    # a direct-label clone without its own label yields no row (unmeasured), never the derived gold
    wjl(os.path.join(rd, "labels.jsonl"), [lab("0:s1", True, ["A", "B"]), lab("0:c2", True, ["C"])])
    rows2 = {g["case_id"]: g for g in V.G.build_gold(0, data, write=False)}
    check("G13 gold: a direct-label clone without an agreed label of its own gets NO gold row (the derived gold never stands in for it)", "0:c1" not in rows2 and "0:c3" not in rows2 and "0:c2" in rows2, sorted(rows2))


# ---------------------------------------------------------------- the leak fix
def leak_checks(V, tmp):
    check = V.check
    import baselines as BL
    import build_controls as BC
    import frontend as F
    import gold as G
    mem = {"round": 0, "areas": {"A": ["tests/a.py", "tests/src/x_old.py"], "B": ["tests/bdir/b.py"]}, "pool": [], "heuristic_only": [], "pool_areas": {}, "unabsorbed": []}
    own = F.Ownership(mem)
    tree = {"tests/a.py", "tests/src/x_old.py", "tests/bdir/b.py", "tests/zzz/u.py"}
    state = F.TreeState(tree)
    src = {"role": "test", "scope": "examples-tests", "dir": "tests/src", "cluster_id": "cl1", "members": ["tests/src/x_new.py"]}
    legacy = BC._p1_target(src, "A", own, tree, state, 3)
    leak = BC._p1_target(src, "A", own, tree, state, 3, leak_check=True)
    b1_legacy, b2_legacy = BL.b1(legacy[0], tree, own), BL.b2(legacy[0], tree, own)
    check("G13 leak fix (ctl-23): without the check a stratum-3 target (tests/zzz, no owned file) is chosen although B1 and B2 fall back to the parent directory `tests`, which A owns, and answer A (the rev-12 leak)",
          legacy is not None and legacy[1] == "tests/zzz" and b1_legacy[1] == "A" and b2_legacy[1] == "A", (legacy, b1_legacy, b2_legacy))
    check("G13 leak fix: with leak_check the target is a directory where neither B1 nor B2 answers A (here the stratum-1 directory of area B: both answer B)",
          leak is not None and leak[1] == "tests/bdir" and BL.b1(leak[0], tree, own)[1] == "B" and BL.b2(leak[0], tree, own)[1] == "B", leak)
    only = {"round": 0, "areas": {"A": ["tests/a.py", "tests/src/x_old.py"]}, "pool": [], "heuristic_only": [], "pool_areas": {}, "unabsorbed": []}
    own2, tree2 = F.Ownership(only), {"tests/a.py", "tests/src/x_old.py", "tests/zzz/u.py"}
    check("G13 leak fix: when the only free target leaks (a stratum-3 directory under A's own parent), the item is unplaceable under leak_check (None) rather than built with a leaking baseline",
          BC._p1_target(src, "A", own2, tree2, F.TreeState(tree2), 3) is not None and BC._p1_target(src, "A", own2, tree2, F.TreeState(tree2), 3, leak_check=True) is None)
    two = {"round": 0, "areas": {"A": ["tests/a.py"], "B": ["tests/bdir/b.py"]}, "pool": [], "heuristic_only": [], "pool_areas": {}, "unabsorbed": []}
    own3, tree3 = F.Ownership(two), {"tests/a.py", "tests/bdir/b.py", "tests/bdir2/c.py"}
    s2 = dict(src, members=["tests/src/y_new.py"], cluster_id="cl2")
    t1 = BC._p1_target(s2, "A", own3, tree3, F.TreeState(tree3), 1, leak_check=True, taken=())
    t2 = BC._p1_target(s2, "A", own3, tree3, F.TreeState(tree3), 1, leak_check=True, taken={t1[1]})
    check("G13 leak fix: a directory in `taken` (hosting an earlier stratum-1 item) is never reused: the second item moves to another directory or falls back",
          t1 is not None and t1[1] == "tests/bdir" and (t2 is None or t2[1] != "tests/bdir"), (t1, t2))
    # R17-01: a rev-12 target directory (forbid_dirs) and a presented path (forbid_paths) are never used, for any stratum, and are counted
    rej_d, rej_p = collections.Counter(), collections.Counter()
    td = BC._p1_target(s2, "A", own3, tree3, F.TreeState(tree3), 1, leak_check=True, forbid_dirs={"tests/bdir"}, rej=rej_d)
    tp = BC._p1_target(s2, "A", own3, tree3, F.TreeState(tree3), 1, leak_check=True, forbid_paths={"tests/bdir/y_new.py"}, rej=rej_p)
    check("G13 R17-01 _p1_target: the only stratum-1 host tests/bdir is chosen without a forbidden set (%s) but never when it is a rev-12 target directory (forbid_dirs: %s, counted %d) or when the new path equals a presented path "
          "of a round-0 control (forbid_paths: %s, counted %d)" % (t1 and t1[1], td and td[1], rej_d["dir"], tp and tp[1], rej_p["path"]),
          t1 is not None and t1[1] == "tests/bdir" and (td is None or td[1] != "tests/bdir") and rej_d["dir"] == 1 and (tp is None or tp[1] != "tests/bdir") and rej_p["path"] == 1, (td, tp))


# ---------------------------------------------------------------- gate_problems, run.sh round-1 refusal, frozen side, budget, one shot
def _gates_json(main, **status):
    G_ = {k: {"status": v} for k, v in status.items()}
    os.makedirs(os.path.join(main, "rounds", "0"), exist_ok=True)
    json.dump({"gates": G_}, open(os.path.join(main, "rounds", "0", "gates.json"), "w"))


def _regate_json(main, gate_set=None, **status):
    root = os.path.join(main, "rounds", "0", "regate")
    os.makedirs(root, exist_ok=True)
    json.dump({"gate_set": gate_set or "x2d-regate-v1", "gates": {k: {"status": v} for k, v in status.items()}}, open(os.path.join(root, "gates.json"), "w"))


def gate_problem_checks(V, tmp):
    check, x2d = V.check, V.x2d
    d = os.path.join(tmp, "g13-gp")
    shutil.rmtree(d, ignore_errors=True)
    os.makedirs(d)
    miss = x2d.gate_problems(d)
    _gates_json(d, K2=PASS, P1=PASS, P2=PASS)
    ok = x2d.gate_problems(d)
    _gates_json(d, K2=PASS, P1=FAIL, P2=FAIL)
    none_yet = x2d.gate_problems(d)
    _regate_json(d, P1=PASS, P2=PASS)
    both = x2d.gate_problems(d)
    _regate_json(d, P1=PASS, P2=PASS, gate_set="x2d-regate-v0")
    tag = x2d.gate_problems(d)
    os.remove(os.path.join(d, "rounds", "0", "regate", "gates.json"))
    cases = {}
    for k, (s1, s2) in {"p1fail": (FAIL, PASS), "p2fail": (PASS, FAIL), "p2unm": (PASS, UNM), "p1unm": (UNM, PASS)}.items():
        _regate_json(d, P1=s1, P2=s2)
        cases[k] = x2d.gate_problems(d)
        os.remove(os.path.join(d, "rounds", "0", "regate", "gates.json"))
    _gates_json(d, K2=PASS, P1=FAIL, P2=PASS)  # only P1 failed in round 0: the re-gate must still PASS BOTH
    _regate_json(d, P1=PASS, P2=FAIL)
    only_p1 = x2d.gate_problems(d)
    _gates_json(d, K2=FAIL, P1=FAIL, P2=FAIL)
    _regate_json(d, P1=PASS, P2=PASS)
    other = x2d.gate_problems(d)
    check("G13 round 1 and the re-gate: gate_problems is empty when every round-0 gate passed, names a missing gates.json, and for a round-0 P1/P2 FAIL demands regate/gates.json (tag x2d-regate-v1) with PASS for BOTH P1 and P2",
          "gates.json missing" in miss[0] and ok == [] and len(none_yet) == 1 and "re-gate" in none_yet[0] and "does not exist" in none_yet[0] and both == [] and len(tag) == 1 and "x2d-regate-v0" in tag[0],
          dict(miss=miss, ok=ok, none_yet=none_yet, both=both, tag=tag))
    check("G13 a re-gate that is not PASS stops the pilot (one shot): P1 FAIL, P2 FAIL, P2 UNMEASURED, P1 UNMEASURED and a P2 FAIL when only P1 failed in round 0 are all refused, each naming the gate and `stops the pilot`",
          all(len(v) == 1 and "stops the pilot" in v[0] for v in list(cases.values()) + [only_p1]) and "re-gate P1 is FAIL" in cases["p1fail"][0] and "re-gate P2 is UNMEASURED" in cases["p2unm"][0] and "re-gate P2 is FAIL" in only_p1[0],
          dict(cases=cases, only_p1=only_p1))
    check("G13 a round-0 FAIL of a gate other than P1 / P2 stops the pilot even with a re-gate PASS (it is not re-gated)", other == ["round-0 gate(s) failed: K2, P1, P2"], other)


def run_refusal_checks(V, tmp):
    """run.sh: round 1 refuses without a re-gate PASS; the one-shot refusals of the stages; the frozen Jev side."""
    check, x2d = V.check, V.x2d
    rv = os.path.join(tmp, "g13-rv")
    os.makedirs(rv, exist_ok=True)

    def sh(main, *args, **kw):
        return V.run([os.path.join(V.S, "run.sh"), *args], cwd=V.ROOT, env=V.clean_env(X2D_DATA=main, X2D_SRC=V.SRC, X2D_REVIEWS_DIR=rv, **kw))
    main, root = world(V, tmp, "rsh", p1=_p1_good(2), p2=_p2_pos(2))
    r_no = sh(main, "units", "1")
    out_no = r_no.stdout + r_no.stderr
    _regate_json(main, P1=FAIL, P2=PASS)
    r_f = sh(main, "units", "1")
    out_f = r_f.stdout + r_f.stderr
    os.remove(os.path.join(root, "gates.json"))
    _regate_json(main, P1=PASS, P2=UNM)
    r_u = sh(main, "units", "1")
    out_u = r_u.stdout + r_u.stderr
    os.remove(os.path.join(root, "gates.json"))
    _regate_json(main, P1=PASS, P2=PASS)
    r_p = sh(main, "units", "1")
    out_p = r_p.stdout + r_p.stderr
    check("G13 run.sh units 1 (round 1) refuses without a re-gate PASS: no regate/gates.json, a P1 FAIL and a P2 UNMEASURED are refused naming the re-gate; with PASS for both it passes the gate check and stops at the NEXT prerequisite (the round-0 critique)",
          r_no.returncode != 0 and "re-gate" in out_no and "does not exist" in out_no and r_f.returncode != 0 and "re-gate P1 is FAIL" in out_f and r_u.returncode != 0 and "re-gate P2 is UNMEASURED" in out_u
          and r_p.returncode != 0 and "re-gate" not in out_p and "pilot-round-0.md" in out_p, (out_no[-250:], out_f[-250:], out_u[-250:], out_p[-250:]))
    # one shot: build refuses when the directory exists; the stages refuse once gates.json exists; every stage refuses a tampered frozen copy
    r_b = sh(main, "regate-build")
    out_b = r_b.stdout + r_b.stderr
    check("G13 run.sh regate-build refuses when D/rounds/0/regate exists (one shot), writing nothing", r_b.returncode != 0 and "one shot" in out_b, out_b[-300:])
    outs = {s: sh(main, s) for s in ("regate-label", "regate-jev", "regate-analyse")}
    check("G13 once regate/gates.json exists, regate-label, regate-jev and regate-analyse refuse (one shot: the re-gate was analysed)",
          all(r.returncode != 0 and "one shot" in (r.stdout + r.stderr) and "gates.json exists" in (r.stdout + r.stderr) for r in outs.values()), {s: (r.stdout + r.stderr)[-200:] for s, r in outs.items()})
    os.remove(os.path.join(root, "gates.json"))
    th = os.path.join(root, "rounds", "0", "thresholds.json")
    keep = open(th).read()
    json.dump({"decisions": {"merge.covered": {"source": "tuned", "threshold": 0.6}}}, open(th, "w"))
    probs_t = x2d.regate_frozen_problems(main)
    r_t = sh(main, "regate-analyse")
    open(th, "w").write(keep)
    gp = os.path.join(root, "rounds", "0", "inputs", "GUIDE.md")
    keep_g = open(gp).read()
    open(gp, "w").write(keep_g + "\nan edit after the freeze\n")
    probs_g = x2d.regate_frozen_problems(main)
    r_g = sh(main, "regate-jev")
    open(gp, "w").write(keep_g)
    cat = os.path.join(root, "catalogue", "catalogue.json")
    keep_c = open(cat).read()
    open(cat, "w").write(keep_c.replace('"title": "r"', '"title": "changed"'))
    probs_c = x2d.regate_frozen_problems(main)
    open(cat, "w").write(keep_c)
    check("G13 the frozen Jev side: a thresholds.json that is not the round-0 provisional 0.70, an edited guide copy and an edited catalogue copy are each reported by regate_frozen_problems and refused by the stages; the untouched copy has no problem",
          any("thresholds.json" in p for p in probs_t) and any("not the round-0 provisional 0.70" in p for p in probs_t) and r_t.returncode != 0 and "refusing" in (r_t.stdout + r_t.stderr)
          and any("inputs digest" in p for p in probs_g) and r_g.returncode != 0 and any("catalogue.json" in p for p in probs_c) and x2d.regate_frozen_problems(main) == [], (probs_t, probs_g, probs_c))
    os.remove(os.path.join(root, "regate.json"))
    r_m = sh(main, "regate-label")
    check("G13 every re-gate stage after the build refuses without regate/regate.json (run.sh regate-build first)", r_m.returncode != 0 and "regate-build first" in (r_m.stdout + r_m.stderr), (r_m.stdout + r_m.stderr)[-200:])
    # the build precondition: nothing to re-gate when no round-0 gate FAILed, and a failed K-gate is not re-gated
    for nm, gs, needle in (("pre-pass", (PASS, PASS), "nothing to re-gate"), ("pre-other", (FAIL, FAIL), "other than P1 / P2")):
        m2, _r2 = world(V, tmp, nm, main_gates=gs)
        shutil.rmtree(os.path.join(m2, "rounds", "0", "regate"))
        if nm == "pre-other":
            _gates_json(m2, K2=FAIL, P1=FAIL, P2=FAIL)
        r_x = sh(m2, "regate-build")
        check("G13 run.sh regate-build refuses when %s (the re-gate follows a P1 / P2 stop only), writing no regate directory" % ("no round-0 gate FAILed" if nm == "pre-pass" else "a gate other than P1 / P2 FAILed"),
              r_x.returncode != 0 and needle in (r_x.stdout + r_x.stderr) and not os.path.exists(os.path.join(m2, "rounds", "0", "regate")), (r_x.stdout + r_x.stderr)[-300:])


def budget_checks(V, tmp):
    check, x2d = V.check, V.x2d
    main, root = world(V, tmp, "budget")
    row = lambda c: {"type": "answer", "cost_known": True, "cost_usd": c}
    V.wjl(os.path.join(main, "rounds", "0", "jev", "scores.jsonl"), [row(0.60), row(0.50)])   # round 0's own spend: $1.10 of the $2 round cap
    out1 = os.path.join(root, "rounds", "0", "jev-p", "scores.p1.jsonl")
    out2 = os.path.join(root, "rounds", "0", "jev-p", "scores.p2.jsonl")
    b1 = x2d.jev_budget(0, out1, data=root)
    V.wjl(out1, [row(0.30)])
    b2 = x2d.jev_budget(0, out2, data=root)
    b2_main = x2d.jev_budget(0, out2, data=main)
    b_r1 = x2d.jev_budget(1, os.path.join(main, "rounds", "1", "jev", "scores.jsonl"), data=main)
    check("G13 budget: the re-gate's Jev spend is charged to the round-0 cap ($2, $1.10 spent): the first re-gate run may spend $0.90, after $0.30 the next $0.60 (the same figure from the study data directory or from the re-gate root); "
          "round 1's own cap is untouched ($2 round cap, $20 total less $1.40)", abs(b1 - 0.90) < 1e-9 and abs(b2 - 0.60) < 1e-9 and abs(b2_main - 0.60) < 1e-9 and abs(b_r1 - 2.0) < 1e-9, (b1, b2, b2_main, b_r1))
    V.wjl(os.path.join(root, "rounds", "0", "jev-p", "scores.p3.jsonl"), [row(0.70)])
    b_cap = x2d.jev_budget(0, os.path.join(root, "rounds", "0", "jev-p", "scores.p4.jsonl"), data=root)
    check("G13 budget: the re-gate cannot exceed the round-0 cap: with $1.10 + $0.30 + $0.70 spent the next run's allowance is 0", b_cap == 0.0, b_cap)
    check("G13 is_regate_root / main_data: the re-gate root maps to the study data directory; a data directory is its own main_data", x2d.is_regate_root(root) and not x2d.is_regate_root(main) and x2d.main_data(root) == os.path.abspath(main)
          and x2d.main_data(main) == os.path.abspath(main) and x2d.regate_root(main) == root)


def precheck_checks(V, tmp):
    """R17-02: the label-only and construction conjuncts are checked before any Jev request, from files that need no Jev answer."""
    check, x2d = V.check, V.x2d
    tail = _tail_p1(8, 8)

    def pre(name, p1, p2, mutate=None):
        main, root = world(V, tmp, name, p1, p2)
        rr = os.path.join(root, "rounds", "0")
        for sub in ("arms-p1", "arms-p2", "jev-p"):
            shutil.rmtree(os.path.join(rr, sub), ignore_errors=True)
        if mutate:
            mutate(rr)
        problems, counts = V.A.regate_label_problems(root)
        return problems, counts, not any(os.path.exists(os.path.join(rr, s)) for s in ("arms-p1", "arms-p2", "jev-p"))

    s1 = lambda n, nb=0: [p1_item(**({"b1": "A"} if i < nb else {})) for i in range(n)] + tail
    ok, c_ok, absent = pre("pre-ok", s1(20), _p2_pos(8))
    check("G13 R17-02 precheck: with exactly 20 eligible stratum-1 items, 8 direct-label positives, no rules-resolved unit and baselines 0 there is no problem; it ran with no arms-p1/, arms-p2/ or jev-p/ present (no Jev file is read)",
          ok == [] and absent and c_ok["stratum1_eligible"] == 20 and c_ok["direct_label_positives"] == 8 and c_ok["rules_resolved_units"] == 0, (ok, c_ok))
    b8, _c8, _a = pre("pre-b8", s1(24, 8), _p2_pos(8))
    p19, c19, _a = pre("pre-s19", s1(19), _p2_pos(8))
    p7, c7, _a = pre("pre-p7", s1(20), _p2_pos(7) + [p2_item(direct="covered", out="assigned", area="R") for _ in range(20)])
    b9, _c9, _a = pre("pre-b9", s1(24, 9), _p2_pos(8))
    rr_, _cr, _a = pre("pre-rules", s1(20), _p2_pos(8), mutate=lambda rr: V.wjl(os.path.join(rr, "units.jsonl"), [u for u in V.jl(os.path.join(rr, "units.jsonl")) if u["case_id"] != "0:ctl-30"]))
    check("G13 R17-02 precheck: 19 eligible stratum-1 items, 7 direct-label positives, B1 recovering A on 9 of 40 (0.225; 8 of 40 passes) and one rules-resolved unit each give their own problem",
          b8 == [] and len(p19) == 1 and "stratum-1" in p19[0] and c19["stratum1_eligible"] == 19 and len(p7) == 1 and "positives 7 < 8" in p7[0] and len(b9) == 1 and b9[0].startswith("B1 recovery")
          and len(rr_) == 1 and "rules_resolved_units 1" in rr_[0], (b8, p19, p7, b9, rr_))
    # the CLI flag runs only inside a re-gate root
    try:
        V.A.regate_label_problems(os.path.join(tmp, "g13-pre-ok"))
        refused = False
    except SystemExit as ex_:
        refused = "inside D/rounds/0/regate" in str(ex_.code)
    check("G13 R17-02 --regate-precheck (regate_label_problems) refuses outside the re-gate root", refused)
    gg = gate(V, tmp, "p1-fields", p1=_p1_good(20), p2=_p2_pos(8))
    check("G13 R17-03 p1_items and p2_items of gates.json carry `date` and `origin` (RESULTS tabulates the figures by year and origin)", {"date", "origin"} <= set(gg["p1_items"][0]) and {"date", "origin"} <= set(gg["p2_items"][0]))


def unit_checks(V, tmp):
    precheck_checks(V, tmp)
    p1_checks(V, tmp)
    p2_checks(V, tmp)
    gold_checks(V, tmp)
    leak_checks(V, tmp)
    gate_problem_checks(V, tmp)
    budget_checks(V, tmp)
    run_refusal_checks(V, tmp)


# ---------------------------------------------------------------- end to end on the finished round-0 fixture run (stub Clawpatch catalogue, real clone, stub labellers, stub Jev)
def _manifest(root):
    out = {}
    for dp, _dn, fns in os.walk(root):
        for fn in fns:
            p = os.path.join(dp, fn)
            out[os.path.relpath(p, root)] = hashlib.sha256(open(p, "rb").read()).hexdigest()
    return out


def _copy_main(V, e, tmp, name):
    """A copy of the finished fixture run's catalogue and round 0 (the rev-12 gates.json carries a PLANTED P1 / P2 FAIL: validation only, so that a re-gate may be built)."""
    dst = os.path.join(tmp, "g13-e2e-" + name)
    shutil.rmtree(dst, ignore_errors=True)
    os.makedirs(dst)
    shutil.copytree(os.path.join(e.data, "catalogue"), os.path.join(dst, "catalogue"))
    shutil.copytree(os.path.join(e.data, "rounds", "0"), os.path.join(dst, "rounds", "0"))
    gp = os.path.join(dst, "rounds", "0", "gates.json")
    g = json.load(open(gp))
    g["gates"]["P1"]["status"] = g["gates"]["P2"]["status"] = FAIL
    json.dump(g, open(gp, "w"), indent=1, sort_keys=True)
    return dst


def pool_checks(V, tmp, dst, info):
    """The fresh pool on the real clone: every candidate is fresh (membership only), and planted exposures disappear from a copy of the data."""
    check, x2d = V.check, V.x2d
    import build_controls as BC
    import frontend as F
    import gold as G
    ex = BC.exposed(dst)
    pool = info["pool"]
    cat = x2d.read_catalogue(dst)
    own = F.Ownership(G.catalogue_membership(cat, 0))
    c0 = set(F.tree_paths(V.SRC, x2d.C0))
    prs = [r["pr_key"] for r in pool]
    paths = [m for r in pool for m in r["members"]]
    check("G13 fresh pool (real clone, membership only): %d candidates; none is a source of the rev-12 P1 / P2 gates (%d), shares a file with one (%d paths, a path twin), belongs to the PR of one (%d PR episodes) or to a PR the guide names (%d)"
          % (len(pool), len(ex["sources"]), len(ex["paths"]), len(ex["prs"]), len(ex["guide_prs"])),
          len(pool) > 0 and not any(r["case_id"] in set(ex["sources"]) or set(r["members"]) & ex["paths"] or r["pr_key"] in ex["prs"] or r["pr_key"] in ex["guide_prs"] for r in pool), len(pool))
    check("G13 fresh pool: one cluster per PR and no file in two clusters; every file is in the C0 tree and owned by exactly one area, the same for the whole cluster; the hash order is sha256('x2d-regate-order' + cluster id); origins are D / P / G / in_no_unit",
          len(set(prs)) == len(prs) and len(set(paths)) == len(paths) and all(set(r["members"]) <= c0 and len({a for m in r["members"] for a in own.owners(m)}) == 1 and all(len(own.owners(m)) == 1 for m in r["members"])
                                                                              and next(iter(own.owners(r["members"][0]))) == r["area_id"] for r in pool)
          and [r["cluster_id"] for r in pool] == sorted((r["cluster_id"] for r in pool), key=lambda c: BC._h("x2d-regate-order", c)) and {r["origin"] for r in pool} <= {"D", "P", "G", "in_no_unit"}, len(pool))
    # planted exposures in a minimal copy: a path twin, a PR episode, a guide-example PR and a used source leave the pool
    d2 = os.path.join(tmp, "g13-pool-plant")
    shutil.rmtree(d2, ignore_errors=True)
    shutil.copytree(os.path.join(dst, "catalogue"), os.path.join(d2, "catalogue"))
    os.makedirs(os.path.join(d2, "rounds", "0"))
    shutil.copytree(os.path.join(dst, "rounds", "0", "inputs"), os.path.join(d2, "rounds", "0", "inputs"))
    for fn in ("units.jsonl", "control-key.jsonl"):
        shutil.copyfile(os.path.join(dst, "rounds", "0", fn), os.path.join(d2, "rounds", "0", fn))
    c_twin, c_ep, c_guide, c_src = pool[0], pool[1], pool[2], pool[3]
    units = V.jl(os.path.join(d2, "rounds", "0", "units.jsonl"))
    key = V.jl(os.path.join(d2, "rounds", "0", "control-key.jsonl"))
    units += [{"case_id": "0:fake-twin", "members": list(c_twin["members"]), "pr_key": "fake-pr-1"}, {"case_id": "0:fake-episode", "members": ["no/such/path.py"], "pr_key": c_ep["pr_key"]},
              {"case_id": c_src["case_id"], "members": list(c_src["members"]), "pr_key": c_src["pr_key"]}]
    key += [{"case_id": "0:ctl-901", "kind": "p1", "source_case_id": "0:fake-twin", "original_paths": []}, {"case_id": "0:ctl-902", "kind": "p2", "source_case_id": "0:fake-episode"},
            {"case_id": "0:ctl-903", "kind": "p1", "source_case_id": c_src["case_id"], "original_paths": []}]
    V.wjl(os.path.join(d2, "rounds", "0", "units.jsonl"), units)
    V.wjl(os.path.join(d2, "rounds", "0", "control-key.jsonl"), key)
    with open(os.path.join(d2, "rounds", "0", "inputs", "GUIDE.md"), "a") as f:
        f.write("\n10. A worked example of PR #%s.\n" % c_guide["pr"])
    info2 = BC.fresh_pool(V.SRC, d2)
    ids2 = {r["cluster_id"] for r in info2["pool"]}
    gone = {c["cluster_id"] for c in (c_twin, c_ep, c_guide, c_src)}
    bad_prs = {c_ep["pr_key"], c_src["pr_key"], str(c_guide["pr"])}
    keep = [r for r in pool if r["cluster_id"] not in gone and r["pr_key"] not in bad_prs and not set(r["members"]) & set(c_twin["members"])]
    tot = lambda i, k: sum(v.get(k, 0) for v in i["funnel"].values())
    check("G13 fresh pool planted exposures: a path twin of an exposed source, a cluster of an exposed PR (episode), a cluster of a guide-example PR and a used source all leave the pool (the others stay), "
          "and the funnel counts one more of each reason (excluded_path_twin, excluded_pr_episode, excluded_guide_example_pr, excluded_used_as_source)",
          not (gone & ids2) and all(r["cluster_id"] in ids2 for r in keep) and len(keep) > 0 and tot(info2, "excluded_path_twin") > tot(info, "excluded_path_twin") and tot(info2, "excluded_pr_episode") > tot(info, "excluded_pr_episode")
          and tot(info2, "excluded_guide_example_pr") > tot(info, "excluded_guide_example_pr") and tot(info2, "excluded_used_as_source") > tot(info, "excluded_used_as_source"),
          (sorted(gone & ids2), len(keep)))
    # the supply is never shrunk: a build whose minimums cannot be met refuses with nothing written
    plan = BC.regate_plan(V.SRC, dst, info=info)
    saved = (x2d.REGATE_P1_N, x2d.REGATE_P2_CLONES, x2d.REGATE_P1_STRATUM1_MIN)
    refusals = {}
    try:
        for nm, patch in (("p1", {"REGATE_P1_N": 10000, "REGATE_P1_STRATUM1_MIN": 5000}), ("p2", {"REGATE_P2_CLONES": 10000})):
            for k, v in patch.items():
                setattr(x2d, k, v)
            pl = BC.regate_plan(V.SRC, dst, info=info)
            try:
                BC.make_regate(V.SRC, dst, plan=pl)
                refusals[nm] = None
            except SystemExit as ex_:
                refusals[nm] = (pl["problems"], str(ex_.code))
            for k, v in zip(("REGATE_P1_N", "REGATE_P2_CLONES", "REGATE_P1_STRATUM1_MIN"), saved):
                setattr(x2d, k, v)
    finally:
        x2d.REGATE_P1_N, x2d.REGATE_P2_CLONES, x2d.REGATE_P1_STRATUM1_MIN = saved
    check("G13 build refuses, nothing written, when the fresh supply cannot meet the minimums (they are never shrunk): P1 items short -> 'only N fresh P1 items can be placed' and 'stratum 1'; P2 clones short -> 'only N fresh P2 clones'; "
          "the real plan has no problem (%d P1 items, %d P2 clones)" % (len(plan["p1"]), len(plan["p2"])),
          plan["problems"] == [] and len(plan["p1"]) == x2d.REGATE_P1_N and len(plan["p2"]) == x2d.REGATE_P2_CLONES and refusals["p1"] is not None and any("fresh P1 items" in p for p in refusals["p1"][0])
          and any("stratum 1" in p for p in refusals["p1"][0]) and "not shrunk" in refusals["p1"][1] and refusals["p2"] is not None and any("fresh P2 clones" in p for p in refusals["p2"][0]) and "not shrunk" in refusals["p2"][1],
          (plan["problems"], refusals))
    # R17-01: no planned P1 placement reuses the presented path of a round-0 control or the target directory of a rev-12 P1 item; a planted rev-12 control at a placement displaces it
    exr = BC.exposed(dst)
    clean = all(pl[1] not in exr["target_dirs"] and not (set(pl[0]) & exr["presented"]) for _r, _a, pl, _s in plan["p1"])
    d3 = os.path.join(tmp, "g13-plant-presented")
    shutil.rmtree(d3, ignore_errors=True)
    shutil.copytree(os.path.join(dst, "catalogue"), os.path.join(d3, "catalogue"))
    os.makedirs(os.path.join(d3, "rounds", "0"))
    shutil.copytree(os.path.join(dst, "rounds", "0", "inputs"), os.path.join(d3, "rounds", "0", "inputs"))
    pl0 = next(it for it in plan["p1"] if it[3] == 1)
    u3 = V.jl(os.path.join(dst, "rounds", "0", "units.jsonl")) + [{"case_id": "0:ctl-950", "control": True, "members": list(pl0[2][0]), "pr_key": "planted"}]
    k3 = V.jl(os.path.join(dst, "rounds", "0", "control-key.jsonl")) + [{"case_id": "0:ctl-950", "kind": "p1", "target_dir": pl0[2][1], "original_paths": []}]
    V.wjl(os.path.join(d3, "rounds", "0", "units.jsonl"), u3)
    V.wjl(os.path.join(d3, "rounds", "0", "control-key.jsonl"), k3)
    plan3 = BC.regate_plan(V.SRC, d3, info=info)
    mv = [it for it in plan3["p1"] if it[0]["cluster_id"] == pl0[0]["cluster_id"]]
    check("G13 R17-01 fresh P1 placements: none of the %d planned items is presented at the path of a round-0 control or in the target directory of a rev-12 P1 item (%d presented paths, %d target directories); with a rev-12 control planted at "
          "the stratum-1 placement of candidate %s (dir %s) the plan no longer places any item there (the candidate moves or drops), and the supply records the rejections"
          % (len(plan["p1"]), len(exr["presented"]), len(exr["target_dirs"]), pl0[0]["cluster_id"][:8], pl0[2][1]),
          clean and not any(it[2][1] == pl0[2][1] or set(it[2][0]) & set(pl0[2][0]) for it in plan3["p1"]) and (not mv or mv[0][2][1] != pl0[2][1])
          and plan3["supply"]["exposed"]["rev12_p1_target_dirs"] == len(exr["target_dirs"]) + 1 and plan3["supply"]["exposed"]["p1_directory_candidates_rejected_as_rev12_target"] > 0, (clean, pl0[2][1]))
    return plan


def build_checks(V, tmp, e, dst, before):
    check, x2d = V.check, V.x2d
    import build_controls as BC
    r = e.rs("regate-build", X2D_DATA=dst)
    root = x2d.regate_root(dst)
    rr = os.path.join(root, "rounds", "0")
    check("G13 run.sh regate-build builds D/rounds/0/regate (marker, supply table, frozen Jev side, units, control key, catalogue-p1 / -p2, Jev manifest) and the baselines B0 / B1 / B2 inside it",
          r.returncode == 0 and all(os.path.exists(os.path.join(root, p)) for p in ("regate.json", "supply.json", "catalogue/catalogue.json", "rounds/0/units.jsonl", "rounds/0/control-key.jsonl", "rounds/0/catalogue-p1.json",
                                                                                   "rounds/0/catalogue-p2.json", "rounds/0/jev-in/manifest.json", "rounds/0/inputs/GUIDE.md", "rounds/0/thresholds.json", "rounds/0/arms/B1.jsonl",
                                                                                   "rounds/0/arms/B2.jsonl")), (r.stdout + r.stderr)[-500:])
    if r.returncode != 0:
        return None
    units, key = V.jl(os.path.join(rr, "units.jsonl")), V.jl(os.path.join(rr, "control-key.jsonl"))
    kinds = collections.Counter(k["kind"] for k in key)
    src = [u for u in units if u.get("set") == "F"]
    ctl = {u["case_id"]: u for u in units if u.get("control")}
    p1 = [k for k in key if k["kind"] == "p1"]
    p2 = [k for k in key if k["kind"] == "p2"]
    s1 = sum(1 for k in p1 if ctl[k["case_id"]]["stratum"] == 1)
    check("G13 build: %d P1 items (>= %d, stratum 1: %d >= %d), %d P2 clones, %d fresh source units (set F) for the clones, no other control (no dup / trap / prose / planted_new: those gates stand from round 0); every P1 key is leak_checked "
          "and its achieved stratum equals the unit's" % (len(p1), x2d.REGATE_P1_N, s1, x2d.REGATE_P1_STRATUM1_MIN, len(p2), len(src)),
          kinds == collections.Counter({"p1": x2d.REGATE_P1_N, "p2": x2d.REGATE_P2_CLONES}) and s1 >= x2d.REGATE_P1_STRATUM1_MIN and len(src) == x2d.REGATE_P2_CLONES and all(k["leak_checked"] and k["achieved_stratum"] == ctl[k["case_id"]]["stratum"] for k in p1)
          and len(units) == len(src) + len(ctl) and {u["control_kind"] for u in ctl.values()} == {"p1", "p2"}, (kinds, s1, len(src)))
    arm = lambda n: {a["case_id"]: a for a in V.jl(os.path.join(rr, "arms", n + ".jsonl"))}
    b1, b2 = arm("B1"), arm("B2")
    leaks = [k["case_id"] for k in p1 if k["area_id"] in (b1[k["case_id"]]["area_id"], b2[k["case_id"]]["area_id"])]
    check("G13 leak fix at pipeline level: on every one of the %d built P1 items neither B1 nor B2 answers the item's own area A (rev 12's ctl-23 leak)" % len(p1), not leaks and len(b1) >= len(p1), leaks)
    cat = x2d.read_catalogue(dst)
    cp1 = json.load(open(os.path.join(rr, "catalogue-p1.json")))
    cp2 = json.load(open(os.path.join(rr, "catalogue-p2.json")))
    p1_members = {m for k in p1 for m in k["original_paths"]}
    ablated = set(p2[0]["ablated_area_ids"])
    check("G13 build: catalogue-p1 keeps every area of the catalogue and none of the P1 files; catalogue-p2 is the catalogue minus the %d ablated areas (all clones share that union); the P2 clones keep their C0 files, none of which is a P1 file; "
          "every clone has a distinct source unit and distinct PR" % len(ablated),
          {a["area_id"] for a in cp1["areas"]} == {a["area_id"] for a in cat["areas"]} and not (p1_members & {f for a in cp1["areas"] for f in a["files"]})
          and {a["area_id"] for a in cp2["areas"]} == {a["area_id"] for a in cat["areas"]} - ablated and all(set(k["ablated_area_ids"]) == ablated for k in p2)
          and not ({m for c in ctl.values() if c["control_kind"] == "p2" for m in c["members"]} & p1_members) and len({k["source_case_id"] for k in p2}) == len(p2) and len({c["pr_key"] for c in src}) == len(src)
          and all(k["ablated_area_id"] in ablated and k["gap_positive"] for k in p2), (len(ablated), len(cp2["areas"])))
    check("G13 build: the frozen Jev side (guide, v1 questions, settings, the provisional 0.70 thresholds, catalogue) in the re-gate root is the round-0 one, byte for byte", x2d.regate_frozen_problems(dst) == [], x2d.regate_frozen_problems(dst))
    sup = json.load(open(os.path.join(root, "supply.json")))
    check("G13 build: supply.json records the membership table before any label (funnel per origin, pool, P1 strata, P2 areas) and no problem; every P1 and P2 item carries its commit date (R17-03) and the exposure counts of the "
          "presented paths and rev-12 target directories (R17-01)", sup["problems"] == [] and sup["p1"]["n"] == x2d.REGATE_P1_N and sup["p2"]["clones"] == x2d.REGATE_P2_CLONES
          and sup["membership_only"] is True and sup["gate_set"] == x2d.REGATE_GATE_SET and set(sup["funnel"]) >= {"D", "P", "G", "in_no_unit"}
          and all(i.get("date") for i in sup["p1"]["items"] + sup["p2"]["items"]) and sup["exposed"]["presented_paths"] > 0 and "rev12_p1_target_dirs" in sup["exposed"], list(sup))
    after = _manifest(dst)
    new = sorted(set(after) - set(before))
    changed = sorted(k for k in before if after.get(k) != before[k])
    check("G13 the build writes only under rounds/0/regate/ (%d new files) and leaves every existing byte of the study data directory unchanged (SHA-256 manifest of %d files before / after)" % (len(new), len(before)),
          new and all(p.startswith("rounds/0/regate/") for p in new) and not changed, (changed[:3], [p for p in new if not p.startswith("rounds/0/regate/")][:3]))
    return units, key


def label_truth(V, tmp, dst, units, key):
    """Stub labeller truth for the re-gate units. P1: covered by its own area A. P2 clones: every third (case order) is a direct-label positive (gap and new alternately), the others are covered by an area of catalogue-p2;
    their F source is covered by the ablated area (derived gold: gap). -> (labeller stub path, planted P2 world {case_id: "none"|"covered"})."""
    x2d = V.x2d
    rr = os.path.join(x2d.regate_root(dst), "rounds", "0")
    cp2 = json.load(open(os.path.join(rr, "catalogue-p2.json")))
    r2 = sorted(a["area_id"] for a in cp2["areas"])[0]
    kmap = {k["case_id"]: k for k in key}
    ans = lambda cov, acc, nk=None: {"covered": cov, "acceptable": acc, "primary": acc[0] if acc else None, "none_kind": nk, "rule": "R1", "reason": "stub"}
    t, world = {}, {}
    clones = sorted(k["case_id"] for k in key if k["kind"] == "p2")
    src_of = {kmap[c]["source_case_id"]: c for c in clones}
    for k in key:
        if k["kind"] == "p1":
            a = ans(True, [k["area_id"]])
            t[k["case_id"]] = {"A": a, "B": dict(a), "C": dict(a)}
    for i, c in enumerate(clones):
        if i % 3 == 0:
            a = ans(False, [], "gap" if (i // 3) % 2 == 0 else "new")
            world[c] = "none"
        else:
            a = ans(True, [r2])
            world[c] = "covered"
        t[c] = {"A": a, "B": dict(a), "C": dict(a)}
        s = ans(True, [kmap[c]["ablated_area_id"]])
        t[kmap[c]["source_case_id"]] = {"A": s, "B": dict(s), "C": dict(s)}
    p = os.path.join(tmp, "g13-label-%s.json" % os.path.basename(dst))
    json.dump(t, open(p, "w"))
    return p, world, r2


def plant_jev(V, e, dst, key, units, world, r2, follow_path=False):
    """The stub Jev's planted truth by first member path. Effect: P1 -> its own area A, P2 positives -> none (not covered), retained clones -> r2. Null (`follow_path`): P1 -> the sibling area B, positives -> assigned (covered, r2)."""
    ctl = {u["case_id"]: u for u in units if u.get("control")}
    for k in key:
        path = ctl[k["case_id"]]["members"][0]
        if k["kind"] == "p1":
            area = (k["sibling_area_ids"] or [r2])[0] if follow_path else k["area_id"]
            e.stub.truth[path] = {"area": area, "covered": True}
        else:
            cov = follow_path or world[k["case_id"]] == "covered"
            e.stub.truth[path] = {"area": r2 if cov else None, "covered": bool(cov)}


def e2e_regate(V, tmp, e):
    """The re-gate end to end on copies of the finished round-0 fixture run (called after `gates-analyse`, before the rounds)."""
    check, x2d = V.check, V.x2d
    import build_controls as BC
    if not os.path.isdir(V.SRC):
        check("G13 the end-to-end re-gate needs the fastapi clone X2D_SRC", False, V.SRC)
        return
    dst = _copy_main(V, e, tmp, "a")
    before = _manifest(dst)
    info = BC.fresh_pool(V.SRC, dst)
    pool_checks(V, tmp, dst, info)
    built = build_checks(V, tmp, e, dst, before)
    if built is None:
        return
    units, key = built
    rr = os.path.join(x2d.regate_root(dst), "rounds", "0")
    lp, world_, r2 = label_truth(V, tmp, dst, units, key)
    rl = e.rs("regate-label", X2D_DATA=dst, X2D_LABEL_STUB=lp)
    gold = {g["case_id"]: g for g in V.jl(os.path.join(rr, "gold.jsonl"))}
    clones = sorted(k["case_id"] for k in key if k["kind"] == "p2")
    pos = [c for c in clones if world_[c] == "none"]
    ret = [c for c in clones if world_[c] == "covered"]
    sess = json.load(open(os.path.join(rr, "sessions.json"))) if os.path.exists(os.path.join(rr, "sessions.json")) else {}
    check("G13 run.sh regate-label: both labellers and gold for every re-gate unit (%d P1 items, %d P2 clones, %d sources) inside the re-gate root, under the round-0 guide version; the re-gate has its OWN session cap by the round-0 formula "
          "(W = %d units: cap %s, used %s)" % (sum(1 for k in key if k["kind"] == "p1"), len(clones), len(units) - len(key), len(units), sess.get("cap"), sess.get("used")),
          rl.returncode == 0 and set(gold) == {u["case_id"] for u in units} and sess.get("plan", {}).get("W") == len(units) and sess.get("cap") == -(-22 * len(units) // 8) and sess.get("used", 0) > 0
          and x2d.guide_version(0, x2d.regate_root(dst)) == x2d.guide_version(0, dst), ((rl.stdout + rl.stderr)[-400:], sorted(set(units_ids(units)) - set(gold))[:3]))
    if rl.returncode != 0:
        return
    kinds_pos = collections.Counter(gold[c]["none_kind"] for c in pos)
    check("G13 P2 gold at pipeline level: every clone's gold is its OWN agreed label (%d positives: gap %d and new %d alike; %d covered by %s), the derived gold of its source (rev 12 rule) sits beside it as derived_gold "
          "(gap for all: the source's only area is the ablated one) and is never the gold" % (len(pos), kinds_pos["gap"], kinds_pos["new"], len(ret), r2[:6]),
          all(gold[c]["covered"] is False and gold[c]["gold_rule"] == "direct_label" and gold[c]["derived_gold"]["covered"] is False and gold[c]["derived_gold"]["none_kind"] == "gap" for c in pos)
          and all(gold[c]["covered"] is True and gold[c]["acceptable"] == [r2] and gold[c]["derived_gold"]["covered"] is False for c in ret) and kinds_pos["gap"] > 0 and kinds_pos["new"] > 0,
          {c: (gold[c]["covered"], gold[c].get("derived_gold")) for c in clones[:2]})
    check("G13 P1 gold at pipeline level: the labellers' agreed acceptable set of every item (here exactly its own area A)", all(gold[k["case_id"]]["covered"] and gold[k["case_id"]]["acceptable"] == [k["area_id"]] for k in key if k["kind"] == "p1"))
    # the null variant is cut from the labelled state, before Jev
    dst_n = os.path.join(tmp, "g13-e2e-n")
    shutil.rmtree(dst_n, ignore_errors=True)
    shutil.copytree(dst, dst_n)
    reg_n = x2d.regate_root(dst_n)
    # the copies are byte copies with absolute manifest paths: the null copy's manifest must point at the null copy
    mf = os.path.join(reg_n, "rounds", "0", "jev-in", "manifest.json")
    man = json.load(open(mf))
    for ent in man:
        ent["units"], ent["catalogue"] = ent["units"].replace(dst, dst_n), ent["catalogue"].replace(dst, dst_n)
    json.dump(man, open(mf, "w"), indent=1, sort_keys=True)
    # R17-02 / R17-06: a copy cut from the labelled state with a planted LOW YIELD (the own labels of the P2 clones keep an area: no direct-label positive)
    dst_l = os.path.join(tmp, "g13-e2e-l")
    shutil.rmtree(dst_l, ignore_errors=True)
    shutil.copytree(dst, dst_l)
    rr_l = os.path.join(x2d.regate_root(dst_l), "rounds", "0")
    gl = V.jl(os.path.join(rr_l, "gold.jsonl"))
    for g_ in gl:
        if world_.get(g_["case_id"]) == "none":
            g_.update(covered=True, acceptable=[r2], primary=r2, none_kind=None)
    V.wjl(os.path.join(rr_l, "gold.jsonl"), gl)
    orig_truth = dict(e.stub.truth)
    try:
        m_l = _manifest(dst_l)
        rl_ = e.rs("regate-jev", X2D_DATA=dst_l)
        check("G13 R17-02 end to end: a planted low yield (no direct-label P2 positive in the labels) makes run.sh regate-jev REFUSE before any Jev request: non-zero, 'no Jev request was made', no score file under regate/rounds/0/jev-p/, "
              "no regate/gates.json, and the SHA-256 manifest of the whole data directory (the study's round-0 files included) unchanged",
              rl_.returncode != 0 and "no Jev request was made" in (rl_.stdout + rl_.stderr) and "direct-label P2 positives 0 < %d" % x2d.REGATE_P2_POSITIVES_MIN in (rl_.stdout + rl_.stderr)
              and not any(os.path.exists(os.path.join(rr_l, "jev-p", n)) for n in ("scores.p1.jsonl", "scores.p2.jsonl")) and not os.path.exists(os.path.join(x2d.regate_root(dst_l), "gates.json"))
              and _manifest(dst_l) == m_l, (rl_.stdout + rl_.stderr)[-400:])
        # R17-06: the legacy round-0 stages refuse once round 0's gates.json exists and write nothing under rounds/0 (regate/ excluded)
        r0_pre = {k: v for k, v in _manifest(dst_l).items() if not k.startswith("rounds/0/regate/")}
        lo = {s: e.rs(s, X2D_DATA=dst_l) for s in ("gates-build", "gates-label", "gates-jev")}
        r0_post = {k: v for k, v in _manifest(dst_l).items() if not k.startswith("rounds/0/regate/")}
        check("G13 R17-06 gates-build, gates-label and gates-jev each refuse on a finished round 0 (rounds/0/gates.json exists; 'never written again') and the SHA-256 manifest of rounds/0 (regate/ excluded) is unchanged",
              all(r_.returncode != 0 and "never written again" in (r_.stdout + r_.stderr) for r_ in lo.values()) and r0_pre == r0_post and len(r0_pre) > 0, {s: (r_.stdout + r_.stderr)[-150:] for s, r_ in lo.items()})
    finally:
        e.stub.truth.clear()
        e.stub.truth.update(orig_truth)
    orig_truth = dict(e.stub.truth)
    try:
        # ---- planted effect
        plant_jev(V, e, dst, key, units, world_, r2)
        rj = e.rs("regate-jev", X2D_DATA=dst)
        check("G13 run.sh regate-jev: the p1 and p2 Jev runs (v1 questions, the frozen settings, the round-0 provisional 0.70 thresholds) and compose inside the re-gate root; no n1 / main run, no fan-out",
              rj.returncode == 0 and all(os.path.exists(os.path.join(rr, p)) for p in ("jev-p/scores.p1.jsonl", "jev-p/scores.p2.jsonl", "arms-p1/V-sem.jsonl", "arms-p2/V-sem.jsonl")) and not os.path.exists(os.path.join(rr, "jev", "scores.jsonl"))
              and not os.path.exists(os.path.join(rr, "arms", "fan-out.jsonl")), (rj.stdout + rj.stderr)[-400:])
        ra = e.rs("regate-analyse", X2D_DATA=dst)
        gj = json.load(open(os.path.join(x2d.regate_root(dst), "gates.json"))) if os.path.exists(os.path.join(x2d.regate_root(dst), "gates.json")) else {}
        g = gj.get("gates", {})
        v1, v2 = (g.get("P1") or {}).get("value", {}), (g.get("P2") or {}).get("value", {})
        n_pos, n_ret = len(pos), len(ret)
        check("G13 re-gate on the planted effect (end to end, stubs): analyse exits 0, gate set %s, P1 PASS (stratum-1 accuracy on the answered items 1.0, abstention 0, B1 / B2 0, path-following 0) and P2 PASS (%d direct-label positives, "
              "none rate 1.0, derived vs direct: none|none %d, covered|none %d)" % (x2d.REGATE_GATE_SET, n_pos, n_pos, n_ret),
              ra.returncode == 0 and gj.get("gate_set") == x2d.REGATE_GATE_SET and (g.get("P1") or {}).get("status") == PASS and (g.get("P2") or {}).get("status") == PASS
              and v1["stratum1"]["accuracy_answered"]["value"] == 1.0 and v1["stratum1"]["abstention_rate"]["value"] == 0.0 and v1["baselines"]["B1"]["all_eligible"]["value"] == 0.0 and v1["baselines"]["B2"]["all_eligible"]["value"] == 0.0
              and v1["stratum1"]["path_following_answered"]["value"] == 0.0 and v2["direct_label_positives"] == n_pos and v2["none_rate_answered"]["value"] == 1.0 and v2["derived_vs_direct"] == {"none|none": n_pos, "covered|none": n_ret}
              and v2["twins_excluded"] == [] and v2["retained_covered"]["assigned_correct"] == n_ret and v1["excluded"] == {"no_agreed_gold": [], "gold_not_covered": [], "gold_excludes_A": []},
              ((ra.stdout + ra.stderr)[-500:], {k: (x or {}).get("status") for k, x in g.items()}))
        check("G13 round 1 after a re-gate PASS: gate_problems is empty (the rev-12 FAIL stands in rounds/0/gates.json and the re-gate PASS lifts it) and run.sh units 1 passes the gate check, stopping at the next prerequisite (the round-0 critique)",
              x2d.gate_problems(dst) == [] and json.load(open(os.path.join(dst, "rounds", "0", "gates.json")))["gates"]["P1"]["status"] == FAIL, x2d.gate_problems(dst))
        # ---- one shot
        m_pre = _manifest(dst)
        outs = {s: e.rs(s, X2D_DATA=dst) for s in ("regate-build", "regate-label", "regate-jev", "regate-analyse")}
        check("G13 one shot (end to end): after the analysis a second regate-build, regate-label, regate-jev and regate-analyse all refuse and write nothing (SHA-256 manifest of the data directory unchanged)",
              all(r_.returncode != 0 and "one shot" in (r_.stdout + r_.stderr) for r_ in outs.values()) and _manifest(dst) == m_pre, {s: (r_.stdout + r_.stderr)[-150:] for s, r_ in outs.items()})
        after = _manifest(dst)
        before_main = {k: v for k, v in before.items()}
        new = sorted(set(after) - set(before_main))
        changed = sorted(k for k in before_main if after.get(k) != before_main[k])
        check("G13 after build, label, Jev and analysis the study data directory differs from its starting state ONLY by new files under rounds/0/regate/ (%d new; code-digests, the Jev scores, labels, gold and the rev-12 gates.json of round 0 are unchanged: "
              "SHA-256 manifest of %d files)" % (len(new), len(before_main)), new and all(p.startswith("rounds/0/regate/") for p in new) and not changed, (changed[:3], [p for p in new if not p.startswith("rounds/0/regate/")][:3]))
        # ---- planted null: a path-following Jev, assigning on the positives
        plant_jev(V, e, dst_n, key, units, world_, r2, follow_path=True)
        rjn = e.rs("regate-jev", X2D_DATA=dst_n)
        ran = e.rs("regate-analyse", X2D_DATA=dst_n)
        gn = json.load(open(os.path.join(reg_n, "gates.json")))["gates"] if os.path.exists(os.path.join(reg_n, "gates.json")) else {}
        n1, n2 = (gn.get("P1") or {}).get("value", {}), (gn.get("P2") or {}).get("value", {})
        ru = e.rs("units", "1", X2D_DATA=dst_n)
        check("G13 re-gate on the planted null (end to end, stubs): a path-following Jev (names the sibling area on every P1 item, assigns an area on every P2 positive) FAILs both gates (analyse exits non-zero; P1 accuracy on the answered items 0.0, "
              "path-following 1.0; P2 none rate 0.0); the rev-12 FAIL is not lifted: gate_problems names both, and run.sh units 1 refuses (the pilot stops)",
              rjn.returncode == 0 and ran.returncode != 0 and (gn.get("P1") or {}).get("status") == FAIL and (gn.get("P2") or {}).get("status") == FAIL and n1["stratum1"]["accuracy_answered"]["value"] == 0.0
              and n1["stratum1"]["path_following_answered"]["value"] == 1.0 and n2["none_rate_answered"]["value"] == 0.0 and len(x2d.gate_problems(dst_n)) == 1 and "re-gate P1 is FAIL" in x2d.gate_problems(dst_n)[0]
              and "re-gate P2 is FAIL" in x2d.gate_problems(dst_n)[0] and ru.returncode != 0 and "stops the pilot" in (ru.stdout + ru.stderr), ((ran.stdout + ran.stderr)[-300:], (ru.stdout + ru.stderr)[-300:]))
    finally:
        e.stub.truth.clear()
        e.stub.truth.update(orig_truth)


def units_ids(units):
    return [u["case_id"] for u in units]
