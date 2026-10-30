"""Council round-4 fixtures (infra side): R4-01, R4-02, R4-06, R4-07, R4-08, R4-11 and the infra parts of R4-13 (round-0 critique lines, composed_dev cap).

Every function takes `V`, the validate module (its `check`, `run`, `clean_env`, `jl`, `wjl`, `x2d`, `TU`, `G`, `PY`, `S`, `ROOT`), so this file imports nothing from validate.py. No model call, no secret.
Unit checks: `r4_infra_unit_checks(V, tmp, repo, res)` (pre-pipeline). End-to-end checks run on copies of the finished fixture run: `e2e_r4_round0(V, tmp, e, lp)` (after `gates-label`) and
`e2e_r4_rounds(V, tmp, e)` (after round 10); `round0_lines(V, data)` is the complete round-0 critique body that `validate.write_critique` writes.
"""
import json
import os
import shutil

POWER_CHECK_KEYS = ("P3_u0", "P3_u05", "P3_u10", "N3", "N3_PR", "N4a", "N4b", "N4b_cap", "N4b_rho0")


def _out(r):
    return r.stdout + r.stderr


def _pops(k23):
    """{round: (GO-population clusters (core-gate excluded), PRs)} of the frozen rounds from a k23.json document."""
    return {r: (k23["per_round"][str(r)]["n"] - (k23["per_round"][str(r)].get("core_gate") or 0), len(k23["frozen_pr_counts"][str(r)])) for r in (8, 9, 10)}


def round0_lines(V, data):
    """The recorded-value lines of a complete round-0 critique (R4-13 (7), R4-08), from data/rounds/0/{k23,gates}.json and data/power.json: real_none_proxy (INFO), p3_real_structure, p3_abstention,
    n4_max, frozen_go_population, `user_decision: proceed` when a power check is false or a frozen round is thin, `p2_unmeasured: accepted` when P2 is UNMEASURED."""
    x2d = V.x2d
    k23 = json.load(open(os.path.join(data, "rounds", "0", "k23.json")))
    pw = json.load(open(os.path.join(data, "power.json")))
    gj = json.load(open(os.path.join(data, "rounds", "0", "gates.json")))
    pops = _pops(k23)
    out = [x2d.CRITIQUE_NONE_PROXY_LINE + str(k23["real_none_proxy"]), x2d.CRITIQUE_P3_LINE + "%.4f" % pw["p3_real_structure"]["share"],
           x2d.CRITIQUE_ABSTAIN_LINE + "0.05=%.4f 0.10=%.4f" % (pw["p3_abstention"]["0.05"], pw["p3_abstention"]["0.10"]), x2d.CRITIQUE_N4_LINE + "%.4f" % pw["n4_max"],
           x2d.CRITIQUE_POP_LINE + x2d.line_pop_text(pops)]
    if any(not v for k, v in pw["checks"].items() if k not in x2d.POWER_NULL_KEYS) or any(n < x2d.FROZEN_MIN_CLUSTERS or p < x2d.FROZEN_MIN_PRS for n, p in pops.values()):
        out.append(x2d.CRITIQUE_DECISION_LINE + "proceed")
    if (gj["gates"].get("P2") or {}).get("status") == "UNMEASURED":
        out.append(x2d.CRITIQUE_P2_LINE + "accepted")
    return out


# ---------------------------------------------------------------- unit checks
def critique_checks(V, tmp):
    """R3-16 as extended by R4-08 and R4-13 (7): `critique_problems(0)` over a fixture data dir."""
    x2d, check = V.x2d, V.check
    cd, rv = os.path.join(tmp, "crit0-r4"), os.path.join(tmp, "crit0-r4-reviews")
    os.makedirs(os.path.join(cd, "rounds", "0", "inputs"), exist_ok=True)
    open(os.path.join(cd, "rounds", "0", "inputs", "GUIDE.md"), "w").write("guide\n")
    open(os.path.join(cd, "rounds", "0", "version.json"), "w").write("{}\n")
    os.makedirs(rv, exist_ok=True)
    base = dict(proxy=39, p3=0.84, ab=(0.81, 0.66), n4=0.03, checks={k: True for k in POWER_CHECK_KEYS},
                per_round={"8": {"n": 26, "core_gate": 1}, "9": {"n": 30, "core_gate": 0}, "10": {"n": 35, "core_gate": 2}}, prs={"8": 7, "9": 8, "10": 9})

    def crit(line=None, rec=None, extra=(), drop=(), power=True, p2="PASS"):
        r_ = dict(base, **(rec or {}))
        json.dump({"gates": {"real_none_proxy": {"status": "INFO", "value": r_["proxy"]}, "P2": {"status": p2}}}, open(os.path.join(cd, "rounds", "0", "gates.json"), "w"))
        k23 = {"real_none_proxy": r_["proxy"], "per_round": r_["per_round"], "frozen_pr_counts": {r: {"pr%d" % i: 3 for i in range(r_["prs"][r])} for r in ("8", "9", "10")}}
        json.dump(k23, open(os.path.join(cd, "rounds", "0", "k23.json"), "w"))
        pp = os.path.join(cd, "power.json")
        if power:
            json.dump({"schema": "x2d-power-v5", "checks": r_["checks"], "p3_real_structure": {"share": r_["p3"]}, "p3_abstention": {"0.05": r_["ab"][0], "0.10": r_["ab"][1]}, "n4_max": r_["n4"]}, open(pp, "w"))
        elif os.path.exists(pp):
            os.remove(pp)
        v = dict({"proxy": r_["proxy"], "p3": r_["p3"], "ab": r_["ab"], "n4": r_["n4"], "pop": x2d.line_pop_text(_pops(k23))}, **(line or {}))
        lines = ["Verdict: ZERO ISSUES", x2d.CRITIQUE_VERSION_LINE + x2d.sha256_file(os.path.join(cd, "rounds", "0", "version.json")),
                 x2d.CRITIQUE_INPUTS_LINE + x2d.inputs_digest(os.path.join(cd, "rounds", "0", "inputs")), x2d.CRITIQUE_NONE_PROXY_LINE + str(v["proxy"]), x2d.CRITIQUE_P3_LINE + "%.4f" % v["p3"],
                 x2d.CRITIQUE_ABSTAIN_LINE + "0.05=%.4f 0.10=%.4f" % v["ab"], x2d.CRITIQUE_N4_LINE + "%.4f" % v["n4"], x2d.CRITIQUE_POP_LINE + v["pop"]]
        lines = [l_ for l_ in lines if not any(l_.startswith(d_) for d_ in drop)] + list(extra)
        open(x2d.critique_path(0, rv), "w").write("\n".join(lines) + "\n")
        return x2d.critique_problems(0, data=cd, reviews_dir_=rv)

    has = lambda probs, s: any(s in p_ for p_ in probs)
    dec = x2d.CRITIQUE_DECISION_LINE + "proceed"
    p_ok = crit()
    p_nopower = crit(power=False)
    p_proxy, p_p3, p_ab, p_n4 = crit({"proxy": 38}), crit({"p3": 0.85}), crit({"ab": (0.82, 0.66)}), crit({"n4": 0.05})
    p_ab10 = crit({"ab": (0.81, 0.67)})
    p_pop = crit({"pop": "8=26/7 9=30/8 10=35/9"})
    p_no_ab, p_no_n4, p_no_pop = crit(drop=(x2d.CRITIQUE_ABSTAIN_LINE,)), crit(drop=(x2d.CRITIQUE_N4_LINE,)), crit(drop=(x2d.CRITIQUE_POP_LINE,))
    p_inside = crit(extra=("p3_inside_min: 0.1070",))
    check("R4-13 (7) / R4-08 critique_problems(0): a complete critique passes (a stale `p3_inside_min` line is ignored); power.json must exist; real_none_proxy / p3_real_structure / p3_abstention (both cells) / n4_max must equal gates.json / power.json "
          "(4 decimals) and frozen_go_population must equal rounds/0/k23.json (core-gate units excluded); each line is required",
          p_ok == [] and p_inside == [] and has(p_nopower, "power.json") and has(p_proxy, "real_none_proxy") and has(p_proxy, "does not equal") and has(p_p3, "p3_real_structure") and has(p_ab, "p3_abstention") and has(p_ab10, "p3_abstention")
          and has(p_n4, "n4_max") and has(p_pop, "frozen_go_population") and has(p_pop, "does not equal") and has(p_no_ab, "p3_abstention") and has(p_no_n4, "n4_max") and has(p_no_pop, "frozen_go_population"),
          dict(ok=p_ok, inside=p_inside, nopower=p_nopower, proxy=p_proxy, p3=p_p3, ab=p_ab, ab10=p_ab10, n4=p_n4, pop=p_pop, no_ab=p_no_ab, no_n4=p_no_n4, no_pop=p_no_pop))
    bad_checks = dict(base["checks"], P3_u10=False)
    p_low_proxy = crit({"proxy": 8}, {"proxy": 8})
    p_chk, p_chk_ok, p_chk_no = crit(rec={"checks": bad_checks}), crit(rec={"checks": bad_checks}, extra=(dec,)), crit(rec={"checks": bad_checks}, extra=(x2d.CRITIQUE_DECISION_LINE + "no",))
    p_chk_two = crit(rec={"checks": bad_checks}, extra=(dec, dec))
    thin_n = {"8": {"n": 21, "core_gate": 2}, "9": {"n": 30, "core_gate": 0}, "10": {"n": 35, "core_gate": 2}}  # 21 - 2 = 19 < 20 once core-gate units leave the GO population
    edge_n = {"8": {"n": 20, "core_gate": 0}, "9": {"n": 30, "core_gate": 0}, "10": {"n": 35, "core_gate": 2}}  # exactly 20 clusters, 6 PRs: not thin
    p_thin, p_thin_ok = crit(rec={"per_round": thin_n}), crit(rec={"per_round": thin_n}, extra=(dec,))
    p_edge = crit(rec={"per_round": edge_n, "prs": {"8": 6, "9": 8, "10": 9}})
    p_pr, p_pr_ok = crit(rec={"prs": {"8": 7, "9": 5, "10": 9}}), crit(rec={"prs": {"8": 7, "9": 5, "10": 9}}, extra=(dec,))
    p_nochecks = (lambda: (json.dump({"schema": "x2d-power-v5", "p3_real_structure": {"share": 0.84}, "p3_abstention": {"0.05": 0.81, "0.10": 0.66}, "n4_max": 0.03}, open(os.path.join(cd, "power.json"), "w")),
                           x2d.critique_problems(0, data=cd, reviews_dir_=rv))[1])()
    p_p2, p_p2_ok = crit(p2="UNMEASURED"), crit(p2="UNMEASURED", extra=(x2d.CRITIQUE_P2_LINE + "accepted",))
    check("R4-13 (7) / R4-08 round-0 `user_decision: proceed` is required exactly when a power.json `checks` value is false or a frozen round has < 20 GO-population clusters (core-gate excluded) or < 6 PRs, and not for a small real_none_proxy "
          "(INFO only); a missing `checks` block is refused; an UNMEASURED P2 gate needs `p2_unmeasured: accepted`",
          p_low_proxy == [] and has(p_chk, "user_decision") and has(p_chk, "P3_u10") and p_chk_ok == [] and has(p_chk_no, "user_decision") and has(p_chk_two, "user_decision") and has(p_thin, "user_decision") and has(p_thin, "frozen round")
          and p_thin_ok == [] and p_edge == [] and has(p_pr, "user_decision") and p_pr_ok == [] and has(p_nochecks, "checks") and has(p_p2, "p2_unmeasured") and p_p2_ok == [],
          dict(low_proxy=p_low_proxy, chk=p_chk, chk_ok=p_chk_ok, chk_no=p_chk_no, chk_two=p_chk_two, thin=p_thin, thin_ok=p_thin_ok, edge=p_edge, pr=p_pr, pr_ok=p_pr_ok, nochecks=p_nochecks, p2=p_p2, p2_ok=p_p2_ok))
    # R5-02 (3), (5): a false NULL check is a problem no line clears; a false P3 check needs the line; `checks` must hold exactly POWER_CHECK_KEYS
    null_bad = dict(base["checks"], N4b=False)
    p_null, p_null_dec = crit(rec={"checks": null_bad}), crit(rec={"checks": null_bad}, extra=(dec,))
    p_cap_null = crit(rec={"checks": dict(base["checks"], N4b_cap=False)}, extra=(dec,))
    p_p3_dec = crit(rec={"checks": dict(base["checks"], P3_u10=False)}, extra=(dec,))
    p_missing_key = crit(rec={"checks": {k: True for k in POWER_CHECK_KEYS if k != "N4b_rho0"}})
    p_extra_key = crit(rec={"checks": dict(base["checks"], X=True)})
    check("R5-02 round-0 critique: a false NULL check (N4b) is refused naming the cell whether or not `user_decision: proceed` is present (the GO rule is unsound there: new council round); N4b_cap likewise; a false P3 check (P3_u10) with the line is ok; "
          "`checks` must hold exactly the nine POWER_CHECK_KEYS (a missing or an extra key is refused); the key sets of x2d and the fixtures agree",
          has(p_null, "N4b") and has(p_null, "unsound") and len(p_null) == 1 and has(p_null_dec, "N4b") and has(p_null_dec, "unsound") and len(p_null_dec) == 1 and has(p_cap_null, "N4b_cap") and p_p3_dec == []
          and has(p_missing_key, "exactly") and has(p_extra_key, "exactly") and tuple(POWER_CHECK_KEYS) == tuple(x2d.POWER_CHECK_KEYS) and set(x2d.POWER_NULL_KEYS) <= set(POWER_CHECK_KEYS),
          dict(null=p_null, null_dec=p_null_dec, cap=p_cap_null, p3=p_p3_dec, missing=p_missing_key, extra=p_extra_key))


def _dev_fixture(V, d, n, n_low, n_false_none=0):
    """Round-0 dev data: n gold-covered D units (area A). The first `n_false_none` get a false `none` (merge.covered p_true 0.05), the next `n_low` an unresolved Choice (confidence 0.2 < t_choice 0.6), the rest an
    assigned-correct answer (A, 0.9)."""
    x2d = V.x2d
    rd = os.path.join(d, "rounds", "0")
    units, gold, rows = [], [], [{"type": "header", "versions": {}, "model": x2d.JEV_MODEL, "catalogue_sha256": "c" * 64}]
    base = {"type": "answer", "variant": "V-sem", "area_id": None, "cost_known": True, "cost_usd": 0, "model": x2d.JEV_MODEL, "question_version": "q"}
    for i in range(n):
        cid, cl = "0:u%03d" % i, "cl%d" % i
        units.append({"case_id": cid, "cluster_id": cl, "pr_key": str(i // 3), "set": "D", "control": False, "stratum": 1, "members": ["f%d.py" % i]})
        gold.append({"case_id": cid, "covered": True, "acceptable": ["A"], "primary": "A", "none_kind": None, "agreement": "agreed", "control_kind": None})
        false_none = i < n_false_none
        low = n_false_none <= i < n_false_none + n_low
        rows.append(dict(base, case_id=cid, question="merge.covered", perm=1, p_true=0.05 if false_none else 0.9))
        if not false_none:
            for perm in ((1, 2) if x2d.flip_selected(cl) else (1,)):
                rows.append(dict(base, case_id=cid, question="merge.which_area", perm=perm, choice="A", confidence=0.2 if low else 0.9))
    V.wjl(os.path.join(rd, "units.jsonl"), units)
    V.wjl(os.path.join(rd, "gold.jsonl"), gold)
    V.wjl(os.path.join(rd, "jev", "scores.jsonl"), rows)


def composed_dev_checks(V, tmp):
    """R4-13 (4): tune.composed_dev uses the verdict's definitions (R_assign over resolved covered clusters) and enforces the dev unresolved-rate cap; P_new / R_new are recorded, never enforced; a tuned decision
    that misses any of them reverts to the provisional 0.7 with `no_threshold_meets_minimums`."""
    TU, x2d, check = V.TU, V.x2d, V.check
    dec = {"merge.covered": {"threshold": 0.7, "score_threshold": 0.3}, "merge.which_area": {"threshold": 0.6}, "merge.area_covers": {"threshold": 0.5}}
    da, db, dc = (os.path.join(tmp, "r4-dev-" + t) for t in ("ok", "cap", "newonly"))
    _dev_fixture(V, da, 40, 4)
    _dev_fixture(V, db, 40, 8)
    _dev_fixture(V, dc, 60, 0, n_false_none=6)
    ca, cb, cc = (TU.composed_dev([0], d, dec, "V-sem") for d in (da, db, dc))
    check("R4-13 (4) composed_dev on 40 dev clusters, 4 abstained (u = 0.10 <= cap %.2f): P_assign 1.0, R_assign (resolved) 1.0, unresolved_rate recorded with its cap, minimums met" % x2d.UNRESOLVED_RATE_MAX,
          ca["failed"] == [] and ca["minimums_met"] is True and ca["P_assign"]["value"] == 1.0 and ca["R_assign"]["value"] == 1.0 and ca["R_assign"]["n"] == 36 and abs(ca["unresolved_rate"]["value"] - 0.10) < 1e-12
          and ca["unresolved_rate"]["max"] == x2d.UNRESOLVED_RATE_MAX, ca)
    check("R4-13 (4) composed_dev with 8 of 40 dev clusters abstained (u = 0.20 > cap) while P_assign and R_assign (resolved) are both 1.0: the unresolved rate alone fails the minimums (R_assign over all covered clusters, "
          "abstention = miss, is 0.80 and only recorded)",
          cb["failed"] == ["unresolved_rate"] and cb["minimums_met"] is False and cb["P_assign"]["value"] == 1.0 and cb["R_assign"]["value"] == 1.0 and abs(cb["unresolved_rate"]["value"] - 0.20) < 1e-12
          and abs(cb["R_assign_all"]["value"] - 0.80) < 1e-12, cb)
    check("R4-13 (4) composed_dev records P_new / R_new without enforcing them: 6 false `none` answers on 60 covered clusters give P_new 0 on 6 `none` outputs (recorded), R_assign (resolved) 54/60 = 0.9, and the minimums are met",
          cc["failed"] == [] and cc["minimums_met"] is True and cc["P_new"]["n"] == 6 and cc["P_new"]["value"] == 0.0 and "R_new" in cc and abs(cc["R_assign"]["value"] - 0.9) < 1e-12 and cc["R_assign"]["n"] == 60, cc)
    orig = TU.decide

    def fake_decide(question, rows, script):  # a tuned, ready decision for every question: the revert logic is what is under test
        d = {"threshold": 0.7 if question == "merge.covered" else 0.6, "status": "ready", "n_pos": 60, "n_neg": 60, "source": "tuned", "fold_report": {"stub": True}}
        if question == "merge.covered":
            d["score_threshold"] = 0.3
        return d
    TU.decide = fake_decide
    try:
        dec_ok, _, comp_ok = TU._tune_dev([0], da, "unused", "jev")
        dec_cap, _, comp_cap = TU._tune_dev([0], db, "unused", "jev")
    finally:
        TU.decide = orig
    prim = ("merge.covered", "merge.which_area")
    check("R4-13 (4) a tuned primary decision stays tuned when the composed dev figures meet P_assign, R_assign and the unresolved cap, and reverts to the provisional 0.7 with `no_threshold_meets_minimums` (naming unresolved_rate) when the cap is breached",
          all(dec_ok[q]["source"] == "tuned" for q in prim) and comp_ok["minimums_met"] is True
          and all(dec_cap[q]["source"] == "provisional-0.7" and dec_cap[q]["threshold"] == 0.7 and "no_threshold_meets_minimums" in dec_cap[q]["note"] and "unresolved_rate" in dec_cap[q]["note"] for q in prim) and comp_cap["minimums_met"] is False,
          ({q: dec_ok[q]["source"] for q in prim}, {q: dec_cap[q].get("note") for q in prim}))
    cp = comp_cap.get("composed_dev_provisional") or {}
    check("R5-11 when the tuned decisions revert to the provisional 0.7, `composed_dev_provisional` records the same composed figures at the provisional decisions (P_assign, R_assign, unresolved_rate, minimums_met) beside `composed_dev`; "
          "no reversion, no record; the decisions themselves are unchanged by the record",
          cp.get("arm") == comp_cap["arm"] and all(k in cp for k in ("P_assign", "R_assign", "unresolved_rate", "minimums_met")) and cp["unresolved_rate"]["n"] == comp_cap["unresolved_rate"]["n"] and "composed_dev_provisional" not in comp_ok
          and all(dec_cap[q]["threshold"] == 0.7 for q in prim), (cp, sorted(comp_ok)))


def required_gold_checks(V, tmp):
    """R4-07: required_gold_ids(0) = D u P u every non-P2 control of units.jsonl, whether it was labelled, attempted or never started; a P2 clone is required through its source."""
    d = os.path.join(tmp, "r4-req0")
    rd = os.path.join(d, "rounds", "0")
    units = [{"case_id": "0:d1", "set": "D", "control": False}, {"case_id": "0:p1", "set": "P", "control": False}, {"case_id": "0:ctl-1", "set": "ctl", "control": True},
             {"case_id": "0:ctl-2", "set": "ctl", "control": True}, {"case_id": "0:ctl-3", "set": "ctl", "control": True}, {"case_id": "0:x", "set": "reserve", "control": False}]
    V.wjl(os.path.join(rd, "units.jsonl"), units)
    V.wjl(os.path.join(rd, "control-key.jsonl"), [{"case_id": "0:ctl-1", "kind": "p1"}, {"case_id": "0:ctl-2", "kind": "p2", "source_case_id": "0:d1"}, {"case_id": "0:ctl-3", "kind": "trap"}])
    got = V.G.required_gold_ids(0, data=d)
    V.check("R4-07 required_gold_ids(0) is D u P u every non-P2 control of units.jsonl with no label row and no call row at all (a never-started capped control is required, the P2 clone is not: its gold derives from its source)",
            got == {"0:d1", "0:p1", "0:ctl-1", "0:ctl-3"}, sorted(got))


def r4_infra_unit_checks(V, tmp, repo, res):
    critique_checks(V, tmp)
    required_gold_checks(V, tmp)
    composed_dev_checks(V, tmp)


# ---------------------------------------------------------------- end-to-end checks (copies of the fixture run)
def e2e_r4_round0(V, tmp, e, lp):
    """After `gates-label` (round 0): R4-07 (a never-started non-P2 control) and R4-01 (the contrast gold missing while the main gold exists) on copies of the data."""
    x2d, jl, wjl, check = V.x2d, V.jl, V.wjl, V.check
    # R4-07
    d1 = os.path.join(tmp, "e2e-r4-ctl")
    shutil.copytree(e.data, d1)
    rd = os.path.join(d1, "rounds", "0")
    os.remove(os.path.join(rd, "gold.jsonl"))
    ctl = next(k["case_id"] for k in jl(os.path.join(rd, "control-key.jsonl")) if k["kind"] != "p2")
    for f in ("labels.jsonl", "calls.jsonl"):
        wjl(os.path.join(rd, f), [x_ for x_ in jl(os.path.join(rd, f)) if x_["case_id"] != ctl])
    r_g = e.py("gold.py", "--round", 0, X2D_DATA=d1)
    f1 = e.py("label_x2d.py", "--round", 0, "--new-round", "fill", X2D_DATA=d1, X2D_LABEL_STUB=lp)
    g1 = e.py("gold.py", "--round", 0, X2D_DATA=d1)
    gold_ids = {g_["case_id"] for g_ in jl(os.path.join(rd, "gold.jsonl"))} if os.path.exists(os.path.join(rd, "gold.jsonl")) else set()
    check("E R4-07 round 0: a non-P2 control that was never started (no label row, no call row) makes `gold.py --round 0` refuse, naming it, with nothing written; after `label --new-round fill` gold succeeds and covers it (%s)" % ctl,
          r_g.returncode != 0 and ctl in _out(r_g) and "nothing written" in _out(r_g) and f1.returncode == 0 and g1.returncode == 0 and ctl in gold_ids, ((_out(r_g))[-250:], (_out(f1))[-150:], (_out(g1))[-150:]))
    shutil.rmtree(d1)
    # R4-01 (round 0): the main gold exists, the contrast gold does not
    d2 = os.path.join(tmp, "e2e-r4-r0gold")
    shutil.copytree(e.data, d2)
    rd2 = os.path.join(d2, "rounds", "0")
    main_p, con_p = os.path.join(rd2, "gold.jsonl"), os.path.join(rd2, "gold.contrast.jsonl")
    c_unit = jl(os.path.join(rd2, "units.contrast.jsonl"))[0]["case_id"]
    os.remove(con_p)
    wjl(os.path.join(rd2, "labels.contrast.jsonl"), [x_ for x_ in jl(os.path.join(rd2, "labels.contrast.jsonl")) if x_["case_id"] != c_unit])
    main_before = x2d.sha256_file(main_p)
    r = e.rs("gates-label", "--new-round", "fill", X2D_DATA=d2, X2D_LABEL_STUB=lp)
    con_ids = {g_["case_id"] for g_ in jl(con_p)} if os.path.exists(con_p) else set()
    check("E R4-01 round 0: with the main gold present and the contrast gold absent (one contrast unit unlabelled), `gates-label --new-round fill` fills it and writes the contrast gold only: the main gold is skipped (`exists (append-only): skipped`) and byte-identical",
          r.returncode == 0 and "exists (append-only): skipped" in _out(r) and c_unit in con_ids and x2d.sha256_file(main_p) == main_before, (_out(r))[-400:])
    shutil.rmtree(d2)


def _flip_a(V, d, r):
    p = os.path.join(d, "rounds", str(r), "labels.jsonl")
    rows = V.jl(p)
    last = {}
    for i, x in enumerate(rows):
        if x.get("kind") == "wide" and str(x.get("by", "")).startswith("labeller_A"):
            last[x["case_id"]] = i
    for i in last.values():
        x = rows[i]
        cov = not x["covered"]
        rows[i] = dict(x, covered=cov, acceptable=["zz"] if cov else [], primary="zz" if cov else None, none_kind=None if cov else "gap")
    V.wjl(p, rows)


def e2e_r4_rounds(V, tmp, e):
    """After round 10 on copies of the finished run: R4-01 (round 1), R4-02, R4-06, R4-11."""
    x2d, jl, wjl, check = V.x2d, V.jl, V.wjl, V.check
    sha = x2d.sha256_file
    # ---- R4-01 (round 1: main gold + rounds/2/membership.json present, contrast-window gold absent and one contrast-window unit unlabelled)
    d = os.path.join(tmp, "e2e-r4-r1gold")
    shutil.copytree(e.data, d)
    rd = os.path.join(d, "rounds", "1")
    main_p, con_p, mem_p = os.path.join(rd, "gold.jsonl"), os.path.join(rd, "gold.contrast.jsonl"), os.path.join(d, "rounds", "2", "membership.json")
    c_unit = jl(os.path.join(rd, "units.contrast_window.jsonl"))[0]["case_id"]
    os.remove(con_p)
    wjl(os.path.join(rd, "labels.contrast.jsonl"), [x_ for x_ in jl(os.path.join(rd, "labels.contrast.jsonl")) if x_["case_id"] != c_unit])
    before = sha(main_p), sha(mem_p)
    r_ref = e.rs("gold", 1, X2D_DATA=d)
    check("E R4-01 round 1: with the main gold and rounds/2/membership.json present and one contrast-window unit unlabelled, `run.sh gold 1` skips the existing main gold and refuses on the contrast gold (names the unit, non-zero, nothing written)",
          r_ref.returncode != 0 and "exists (append-only): skipped" in _out(r_ref) and c_unit in _out(r_ref) and "nothing written" in _out(r_ref) and not os.path.exists(con_p) and (sha(main_p), sha(mem_p)) == before, (_out(r_ref))[-400:])
    r_fill = e.rs("label", 1, "--new-round", "fill", X2D_DATA=d, X2D_LABEL_STUB=os.path.join(tmp, "e2e-label-r1.json"))
    r_lc = e.rs("labels-check", 1, X2D_DATA=d)
    r_ok = e.rs("gold", 1, X2D_DATA=d)
    con_ids = {g_["case_id"] for g_ in jl(con_p)} if os.path.exists(con_p) else set()
    check("E R4-01 round 1: after `label --new-round fill` and `labels-check 1`, `run.sh gold 1` writes the contrast gold only (the unit has gold); the main gold and rounds/2/membership.json are byte-identical",
          r_fill.returncode == 0 and r_lc.returncode == 0 and r_ok.returncode == 0 and c_unit in con_ids and (sha(main_p), sha(mem_p)) == before, [(_out(x_))[-200:] for x_ in (r_fill, r_lc, r_ok)])
    shutil.rmtree(d)
    # ---- R4-02: a refused pre-check of a failing round
    d = os.path.join(tmp, "e2e-r4-refused")
    shutil.copytree(e.data, d)
    rd5 = os.path.join(d, "rounds", "5")
    scores, away = os.path.join(rd5, "jev", "scores.jsonl"), os.path.join(rd5, "jev", "scores.jsonl.away")
    os.replace(scores, away)  # `clarify` must precede scoring and gold: the copy is set back to that point
    for p_ in (os.path.join(rd5, "gold.jsonl"), os.path.join(d, "rounds", "6", "membership.json")):
        os.remove(p_)
    _flip_a(V, d, 5)  # kappa_answer < 0.6: the round must clarify
    lp5 = os.path.join(tmp, "e2e-label-r5.json")
    refused_p = os.path.join(rd5, "clarify-refused.json")
    r_c = e.rs("clarify", 5, "validation: cap breach", X2D_DATA=d, X2D_LABEL_STUB=lp5, X2D_STUB_SESSION_CAP="1")
    doc = json.load(open(refused_p)) if os.path.exists(refused_p) else {}
    b_before = open(refused_p, "rb").read() if os.path.exists(refused_p) else b""
    check("E R4-02 round 5: a `clarify` whose pre-check refuses (cap 1) exits non-zero, records NO revision, and writes create-once rounds/5/clarify-refused.json {utc, reason, cap, used, planned}",
          r_c.returncode != 0 and "cannot be relabelled within the session cap" in _out(r_c) and set(doc) == {"utc", "reason", "cap", "used", "planned"} and doc["cap"] == 1 and doc["planned"] > 0 and doc["used"] >= 0
          and not os.path.exists(os.path.join(rd5, "guide-revisions.jsonl")), (_out(r_c))[-300:])
    r_c2 = e.rs("clarify", 5, "validation: again", X2D_DATA=d, X2D_LABEL_STUB=lp5, X2D_STUB_SESSION_CAP="1")
    check("E R4-02 clarify-refused.json cannot be written twice: a second refused `clarify` exits non-zero and leaves the file byte-identical",
          r_c2.returncode != 0 and "already refused" in _out(r_c2) and open(refused_p, "rb").read() == b_before, (_out(r_c2))[-300:])
    r_lc = e.rs("labels-check", 5, X2D_DATA=d)
    lc = json.load(open(os.path.join(rd5, "labels-check.json")))
    r_g = e.rs("gold", 5, X2D_DATA=d)
    os.replace(away, scores)
    for fn in ("results.json", "report.md"):
        if os.path.exists(os.path.join(rd5, fn)):
            os.remove(os.path.join(rd5, fn))
    xa = V.run([V.PY, os.path.join(V.S, "analysis_x2d.py"), "--round", "5"], cwd=V.ROOT, env=V.clean_env(X2D_DATA=d, X2D_REVIEWS_DIR=e.reviews))
    res5 = json.load(open(os.path.join(rd5, "results.json"))) if xa.returncode == 0 and os.path.exists(os.path.join(rd5, "results.json")) else {}
    check("E R4-02 round 5 with clarify-refused.json (no guide revision): the failing labels-check (kappa_answer %s) exits 0 with ok false, `gold` proceeds (gold.jsonl and rounds/6/membership.json written), and the scored round is NON-INFORMATIVE with "
          "results.json informative.labels_check_ok false" % lc.get("kappa_answer"),
          r_lc.returncode == 0 and lc["ok"] is False and r_g.returncode == 0 and os.path.exists(os.path.join(rd5, "gold.jsonl")) and os.path.exists(os.path.join(d, "rounds", "6", "membership.json")) and xa.returncode == 0
          and (res5.get("informative") or {}).get("informative") is False and (res5.get("informative") or {}).get("labels_check_ok") is False, ((_out(r_lc))[-200:], (_out(r_g))[-200:], xa.stderr[-200:], res5.get("informative")))
    shutil.rmtree(d)
    # ---- R5-05: the guide was edited BEFORE a refused pre-check: the round is stuck on drift until the frozen guide bytes are restored; then it proceeds scored and non-informative
    d = os.path.join(tmp, "e2e-r5-edited-guide")
    shutil.copytree(e.data, d)
    rd5 = os.path.join(d, "rounds", "5")
    scores, away = os.path.join(rd5, "jev", "scores.jsonl"), os.path.join(rd5, "jev", "scores.jsonl.away")
    os.replace(scores, away)
    for p_ in (os.path.join(rd5, "gold.jsonl"), os.path.join(d, "rounds", "6", "membership.json")):
        os.remove(p_)
    _flip_a(V, d, 5)
    lp5b = os.path.join(tmp, "e2e-label-r5b.json")
    refused_p = os.path.join(rd5, "clarify-refused.json")
    gpath = os.path.join(rd5, "inputs", "GUIDE.md")
    frozen_guide = open(gpath, "rb").read()
    r_pc = e.rs("clarify-precheck", 5, X2D_DATA=d, X2D_LABEL_STUB=lp5b, X2D_STUB_SESSION_CAP="1")
    refused_ok = r_pc.returncode != 0 and os.path.isfile(refused_p) and not os.path.exists(os.path.join(rd5, "guide-revisions.jsonl"))
    open(gpath, "ab").write(b"\nproposed clarification (edited before the refusal was known)\n")
    r_gold_bad = e.rs("gold", 5, X2D_DATA=d)
    stuck = r_gold_bad.returncode != 0 and not os.path.exists(os.path.join(rd5, "gold.jsonl"))
    open(gpath, "wb").write(frozen_guide)  # recovery: restore the frozen bytes (the proposed text is kept under reviews/ by the operator)
    prev_guide = os.path.join(d, "rounds", "4", "inputs", "GUIDE.md")
    same_as_prev = (not os.path.exists(prev_guide)) or open(prev_guide, "rb").read() == frozen_guide
    r_lc = e.rs("labels-check", 5, X2D_DATA=d)
    r_g = e.rs("gold", 5, X2D_DATA=d)
    os.replace(away, scores)
    for fn in ("results.json", "report.md"):
        if os.path.exists(os.path.join(rd5, fn)):
            os.remove(os.path.join(rd5, fn))
    xa = V.run([V.PY, os.path.join(V.S, "analysis_x2d.py"), "--round", "5"], cwd=V.ROOT, env=V.clean_env(X2D_DATA=d, X2D_REVIEWS_DIR=e.reviews))
    res5 = json.load(open(os.path.join(rd5, "results.json"))) if xa.returncode == 0 and os.path.exists(os.path.join(rd5, "results.json")) else {}
    check("E R5-05 `clarify-precheck 5` before any edit refuses (cap 1: non-zero, clarify-refused.json written, no revision); with the guide edited first, `gold 5` is refused (drift, no gold written); after restoring the frozen guide bytes `labels-check 5` exits 0, "
          "`gold 5` proceeds and the scored round is non-informative (labels_check_ok false)",
          refused_ok and stuck and same_as_prev and r_lc.returncode == 0 and r_g.returncode == 0 and os.path.exists(os.path.join(rd5, "gold.jsonl")) and xa.returncode == 0 and (res5.get("informative") or {}).get("informative") is False,
          ((_out(r_pc))[-200:], (_out(r_gold_bad))[-200:], (_out(r_lc))[-150:], (_out(r_g))[-150:], xa.stderr[-150:], res5.get("informative")))
    shutil.rmtree(d)
    # ---- R4-06: one byte of jev-v7/units.jsonl changed after the re-score
    d = os.path.join(tmp, "e2e-r4-v7sha")
    shutil.copytree(e.data, d)
    up = os.path.join(d, "rounds", "2", x2d.V7_RESCORE_SUB, "units.jsonl")
    open(up, "ab").write(b"\n")  # one byte more: the units file is no longer the one the driver read (header input_sha256)
    r8 = os.path.join(d, "rounds", "8")
    keep = {f: sha(os.path.join(r8, f)) for f in ("version.json", "units.jsonl")}
    r_u = e.rs("units", 8, X2D_DATA=d)
    for fn in os.listdir(r8):
        if fn == "thresholds.json" or (fn.startswith("scores") and fn.endswith(".csv")):
            os.remove(os.path.join(r8, fn))
    r_t = e.rs("tune", 8, X2D_DATA=d)
    msg = "the v_7 units file is not the one the re-score read"
    check("E R4-06 one byte of rounds/2/jev-v7/units.jsonl changed after scoring: `units 8` and `tune 8` are refused (the header's input_sha256 is compared with the file), nothing is written (round 8's version.json / units.jsonl unchanged, no thresholds.json, no scores csv)",
          r_u.returncode != 0 and msg in _out(r_u) and r_t.returncode != 0 and msg in _out(r_t) and {f: sha(os.path.join(r8, f)) for f in keep} == keep
          and not any(fn == "thresholds.json" or (fn.startswith("scores") and fn.endswith(".csv")) for fn in os.listdir(r8)), ((_out(r_u))[-250:], (_out(r_t))[-250:]))
    shutil.rmtree(d)
    # ---- R4-11: a header-only (or partial) jev/scores.jsonl must not become the round's recorded test rows
    d = os.path.join(tmp, "e2e-r4-partial")
    shutil.copytree(e.data, d)
    r4d = os.path.join(d, "rounds", "4")
    sp = os.path.join(r4d, "jev", "scores.jsonl")
    full = jl(sp)

    def clear():
        for fn in os.listdir(r4d):
            if fn == "thresholds.json" or (fn.startswith("scores") and fn.endswith(".csv")):
                os.remove(os.path.join(r4d, fn))
    clear()
    wjl(sp, [x_ for x_ in full if x_.get("type") == "header"])
    r_h = e.rs("tune", 4, X2D_DATA=d)
    wrote_h = [fn for fn in os.listdir(r4d) if fn == "thresholds.json" or (fn.startswith("scores") and fn.endswith(".csv"))]
    win_ids = {u_["case_id"] for u_ in jl(os.path.join(r4d, "units.jsonl")) if u_.get("set") == "window" and not u_.get("control")}
    drop = next(x_ for x_ in full if x_.get("type") == "answer" and x_.get("question") == "merge.covered" and x_.get("variant") == "V-sem" and x_["case_id"] in win_ids)
    wjl(sp, [x_ for x_ in full if x_ is not drop])
    r_p = e.rs("tune", 4, X2D_DATA=d)
    wrote_p = [fn for fn in os.listdir(r4d) if fn == "thresholds.json" or (fn.startswith("scores") and fn.endswith(".csv"))]
    check("E R4-11 round 4: a header-only jev/scores.jsonl (the driver writes its header before any request) and a file missing one required merge.covered row are refused by `tune 4` (\"re-run jev 4\"); thresholds.json and scores csv files are not written",
          r_h.returncode != 0 and "re-run jev 4" in _out(r_h) and not wrote_h and r_p.returncode != 0 and "re-run jev 4" in _out(r_p) and not wrote_p, ((_out(r_h))[-250:], (_out(r_p))[-250:], wrote_h, wrote_p))
    shutil.rmtree(d)
