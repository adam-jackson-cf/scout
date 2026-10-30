"""X2d analysis (PREREGISTRATION sections 3, 4.3, 5, 8, 9, 10, 11; thresholds.json `study`). Standard library only.

Pure layer (importable, no I/O): wilson, wilson_bounds_one_sided, mcnemar_exact, pr_cluster_bootstrap, cohen_kappa, score_outcome, counts_for / rates,
summarise, go_verdict, informative, futility.

CLI (run with .tools/venv/bin/python; D = x2d.DATA, env X2D_DATA; every CLI calls x2d.verify_pins(); X2D_SKIP_SIGNOFF=1 makes it
call x2d.verify_pins(require_signoff=False), used by validation only):

    analysis_x2d.py --round R [--pack]   every section-9 figure of round R -> D/rounds/R/{results.json,report.md}  (append-only);
                                         --pack alone writes the council critique pack S/reviews/pilot-round-<R>-pack.md (facts only, no causes proposed)
    analysis_x2d.py --gates [--repo CLONE]  round-0 gates K2 K3 K4 K5 K6_N1 P1 P2 planted_new + core_gate (needs the clone: the C0 tree) -> D/rounds/0/gates.json; exit 1 if any gate FAILs
    analysis_x2d.py --labels-check R      before Jev scores round R: kappa on the joint answer over the GO population, adjudication rate and rule-citation concentration (report-only) of the round's wide labels (current guide version only);
                                         writes D/rounds/R/labels-check.json (ok, figures, labels_sha256, guide_version; run.sh binds gold/jev/tune/analysis to it, R2-03, R3-07); exit 1 ONLY for a refinement round (1-7) with kappa_answer < 0.6, no recorded guide
                                         revision and no rounds/R/clarify-refused.json (guide-clarification path, PREREGISTRATION 8/10 (v)); round 0, rounds 8-10, a clarified round and a round whose clarification was refused (R4-02) exit 0 (scored, non-informative)
    analysis_x2d.py --futility [--upto R]  futility rule over rounds 1..R (default: all rounds with results); exit 0 = not futile, 3 = fired
                                         (then D/verdict.json is written as NO-GO futility, append-only)
    analysis_x2d.py --verdict            ONCE on the pooled frozen rounds 8-10 -> D/verdict.json; refuses if it exists or a frozen round has no results.json

Assumptions about inputs the contract leaves open (every one is also reported in the outputs where it matters):
  * Scored units of round r >= 1 are units.jsonl rows with set == "window" and not control; set == "audit" rows are the front-end audit items; round 0 sets are D, P, ctl.
  * Gold comes from gold.jsonl (wide final rows). A unit without a gold row, or with covered not a bool, is `unmeasured` and in no metric. A missing arm row is `unmeasured` for that arm.
  * Round-0 layout (Main's addendum): jev/scores.jsonl (D+P and controls), jev-n1/scores.jsonl, jev-p/scores.p1.jsonl, scores.p2.jsonl; composed arm files
    arms/ arms-n1/ arms-p1/ arms-p2/; control gold comes from control-key.jsonl (`gold` with acceptable/area_id/primary, covered, none_kind); `k23.json`
    holds the K2/K3 shares as {"k2": x | {"share": x} | {"k": k, "n": n}, "k3": ...} (also accepted: "k2_share", "k3_share").
  * An integrity pause of round r is the file D/rounds/<r>/integrity_pause.json (any content; skipped by the futility rule).
  * Rule-resolved file counts for the end-to-end figure are the lookups.jsonl `counts` entries whose key matches modified/(a) (exact lookups, assumed correct)
    or renamed_inherit/variant_inherit/(c)/(d) (corrected by the audit's false-inherit estimate).
  * Wilson bounds named "90 %" in the GO rule are ONE-SIDED 90 % (z = inv_cdf(0.90)); bootstrap bounds: lower = quantile 0.10 (one-sided 90 %), futility upper = quantile 0.80 (one-sided 80 %).
  * Perception check (R1-01, R2-09, R2-17): the two jev-state questions (merge.covered, merge.which_area) are separate rows keyed (case_id, question). Per labeller and for the consensus the
    COMBINED answer is `none` if the covered question's final says not covered, else the which_area final's primary. The check measures whether the STATE suffices, not the threshold
    policy: it compares the V-sem arm's UNTHRESHOLDED combined answer (compose's `raw_covered` / `raw_choice`: none iff not raw_covered, else the permutation-1 top option) with the labellers'
    combined answer, whatever the primary arm is; the thresholded agreement, the unresolved rate and the primary arm's agreement are reported beside it. Labeller-labeller agreement = A's
    combined vs B's combined. Q1 and Q2 agreement are reported separately. Pass iff pooled units with both questions final >= 20 and jev_agree >= ll_agree - 0.10. Pooled over the frozen rounds 8-10 (v_7).
  * Primary arm (R1-03): x2d.primary_arm(D) (fan-out when rounds/0/gates.json fan_out_primary, else V-sem) drives pooling, perception, futility and the verdict; V-sem stays reported.
  * GO population (R1-14): units with stratum_reason == "core_gate" are excluded from the pooling, informativeness counts, futility and verdict; they stay in the stratum report.
  * Stratum 4 is assigned post hoc from gold none_kind == new (descriptive only).
"""
import argparse
import collections
import json
import math
import os
import random
import re
import sys
from statistics import NormalDist

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import x2d  # noqa: E402  (puts X2's dir and the study dir on sys.path)
import common as C  # noqa: E402

HERE = os.path.dirname(os.path.abspath(__file__))
ARMS = ["B0", "B1", "B2", "V-sem", "V-sib", "V-title", "fan-out", "reference-claude"]
CONTRAST_ARMS = ["contrast-B2", "contrast-V-sem"]
EPISODES = [14569, 14358, 14168, 16403, 14652]
CAUSES = ["label", "policy gap", "missing/clipped state", "wording", "parser-decidable rule", "model limit", "code failure"]
GO = {"p_assign_wilson_lb": 0.80, "level": 0.90, "guard_level": 0.95, "gain_min": 0.10, "spread_min": 3, "unresolved_rate_max": 0.15, "resolved_min": 50, "perception_margin": 0.10, "perception_min_n": 20}
RESOLVED = ("assigned", "none")  # a cluster is RESOLVED when the arm's composed output is one of these; `unresolved` (any reason) is an abstention (R4-13)
REASONS = ("below_t_cov", "below_t_choice", "permutation_disagreement", "failure")
FUTILITY = {"gain_upper_level": 0.80, "gain_upper_max": 0.02, "p_upper_level": 0.80, "run": 3}
N1_FLOOR = 0.20
RULE_CONC = 0.40


# ------------------------------------------------------------------ intervals
def wilson(k, n, z=C.Z95):
    """Wilson score interval (X2's frozen common.wilson); (nan, nan) when n == 0."""
    return C.wilson(k, n, z)


def z_one_sided(level):
    return NormalDist().inv_cdf(level)


def wilson_bounds_one_sided(k, n, level):
    """(lower, upper) one-sided Wilson bounds at `level` (e.g. 0.90): each is the matching end of the Wilson interval with z = inv_cdf(level)."""
    lo, hi = wilson(k, n, z_one_sided(level))
    return lo, hi


def mcnemar_exact(b, c):
    """Two-sided exact McNemar p on the discordant counts b, c: min(1, 2 * P(Binomial(b+c, 1/2) <= min(b, c)))."""
    n = b + c
    if n == 0:
        return 1.0
    m = min(b, c)
    tail = sum(math.comb(n, i) for i in range(m + 1)) / (2.0 ** n)
    return min(1.0, 2.0 * tail)


def _quantile(sorted_vals, q):
    n = len(sorted_vals)
    if n == 0:
        return float("nan")
    pos = max(0.0, min(1.0, q)) * (n - 1)
    lo = int(math.floor(pos))
    hi = min(n - 1, lo + 1)
    return sorted_vals[lo] + (sorted_vals[hi] - sorted_vals[lo]) * (pos - lo)


def pr_cluster_bootstrap(rows, stat_fn, B=10000, seed=x2d.SEED, key="pr_key", levels=(0.8, 0.9, 0.95)):
    """Resample PRs (the groups of `key`) with replacement, B times; stat_fn(list of rows) -> float | None (None replicates are dropped).
    -> {"point", "B", "n_valid", "two_sided": {lvl: [lo, hi]}, "lower": {lvl: one-sided lower bound}, "upper": {lvl: one-sided upper bound}}.
    Deterministic: PRs sorted by str(key), random.Random(seed), replicate b draws m indices with randrange(m)."""
    groups = collections.defaultdict(list)
    for r in rows:
        groups[str(r[key])].append(r)
    keys = sorted(groups)
    glist = [groups[k] for k in keys]
    m = len(glist)
    out = {"point": stat_fn(list(rows)) if rows else None, "B": B, "n_valid": 0, "n_prs": m, "two_sided": {}, "lower": {}, "upper": {}}
    if m == 0:
        return out
    rng = random.Random(seed)
    vals = []
    for _ in range(B):
        sample = []
        for _i in range(m):
            sample.extend(glist[rng.randrange(m)])
        v = stat_fn(sample)
        if v is not None and not (isinstance(v, float) and math.isnan(v)):
            vals.append(v)
    vals.sort()
    out["n_valid"] = len(vals)
    for lvl in levels:
        out["two_sided"][lvl] = [_quantile(vals, (1 - lvl) / 2), _quantile(vals, 1 - (1 - lvl) / 2)]
        out["lower"][lvl] = _quantile(vals, 1 - lvl)
        out["upper"][lvl] = _quantile(vals, lvl)
    return out


def cohen_kappa(a, b):
    """Cohen's kappa of two equal-length label lists. None when empty; degenerate chance agreement 1: 1.0 if all agree else 0.0."""
    n = len(a)
    if n == 0 or n != len(b):
        return None
    po = sum(1 for x, y in zip(a, b) if x == y) / n
    ca, cb = collections.Counter(a), collections.Counter(b)
    pe = sum(ca[k] * cb.get(k, 0) for k in ca) / (n * n)
    if pe >= 1.0:
        return 1.0 if po >= 1.0 else 0.0
    return (po - pe) / (1 - pe)


# ------------------------------------------------------------------ outcomes
def score_outcome(arm_row, gold_row):
    """One arm row against one gold row (section 9). -> {"output", "area_id", "kind", "correct": True|False|None, "gold_covered", "none_kind"}.
    kinds: assigned_correct, assigned_wrong, assigned_gold_none, none_correct_gap, none_correct_new, none_wrong, unresolved, missing (no arm row), ungolded (no usable gold).
    correct: assigned(a) iff a in acceptable (gold covered); none iff gold covered is False; unresolved is never correct; None when missing/ungolded."""
    if arm_row is None:
        return {"output": None, "area_id": None, "kind": "missing", "correct": None, "gold_covered": None, "none_kind": None, "reason": None,
                "raw_covered": None, "raw_choice": None, "raw_correct": None, "guess_correct": None}
    out = arm_row.get("output")
    base = {"output": out, "area_id": arm_row.get("area_id"), "raw_covered": arm_row.get("raw_covered"), "raw_choice": arm_row.get("raw_choice"), "gold_covered": None, "none_kind": None,
            "reason": arm_row.get("unresolved_reason"), "raw_correct": None, "guess_correct": None}
    if gold_row is None or not isinstance(gold_row.get("covered"), bool):
        base.update(kind="ungolded", correct=None)
        return base
    cov = gold_row["covered"]
    base["gold_covered"] = cov
    base["none_kind"] = gold_row.get("none_kind")
    acc = set(gold_row.get("acceptable") or [])
    # R5-01: raw_correct = correctness of the arm's UNTHRESHOLDED combined answer (`none` iff raw_covered is False, else raw_choice); a missing raw answer is wrong.
    rc = arm_row.get("raw_covered")
    if rc is False:
        base["raw_correct"] = not cov
    elif rc is True or rc is None:
        base["raw_correct"] = bool(rc is True and cov and arm_row.get("raw_choice") in acc and arm_row.get("raw_choice") is not None)
    if out == "assigned":
        if cov and arm_row.get("area_id") in acc:
            base.update(kind="assigned_correct", correct=True)
        else:
            base.update(kind="assigned_wrong" if cov else "assigned_gold_none", correct=False)
    elif out == "none":
        if not cov:
            base.update(kind="none_correct_" + ("gap" if gold_row.get("none_kind") == "gap" else "new"), correct=True)
        else:
            base.update(kind="none_wrong", correct=False)
    elif out == "unresolved":
        base.update(kind="unresolved", correct=False)
    else:
        base.update(kind="missing", correct=None)
    if base["correct"] is not None:
        base["guess_correct"] = bool(base["raw_correct"]) if out == "unresolved" else base["correct"]
    return base


def counts_for(outcomes):
    """Counts of section 9 from a list of score_outcome dicts (missing/ungolded are tallied, not counted)."""
    c = collections.Counter()
    for o in outcomes:
        k = o["kind"]
        if k == "missing":
            c["missing"] += 1
            continue
        if k == "ungolded":
            c["ungolded"] += 1
            continue
        c["n"] += 1
        c["correct"] += 1 if o["correct"] else 0
        cov = o["gold_covered"]
        c["gold_covered"] += 1 if cov else 0
        c["gold_none"] += 0 if cov else 1
        if not cov:
            c["gold_none_" + ("gap" if o["none_kind"] == "gap" else "new")] += 1
        if cov and o["output"] in RESOLVED:
            c["gold_covered_res"] += 1
        if o["output"] == "assigned":
            c["assigned"] += 1
            c["assigned_correct"] += 1 if o["correct"] else 0
        elif o["output"] == "none":
            c["none"] += 1
            c["none_correct"] += 1 if o["correct"] else 0
            if o["correct"]:
                c["none_correct_" + ("gap" if o["none_kind"] == "gap" else "new")] += 1
        elif o["output"] == "unresolved":
            c["unresolved"] += 1
            c["unresolved_" + (o.get("reason") if o.get("reason") in REASONS else "other")] += 1
    return dict(c)


def _rate(k, n):
    lo, hi = wilson(k, n) if n else (float("nan"), float("nan"))
    return {"k": k, "n": n, "value": (k / n) if n else None, "wilson95": [None if math.isnan(lo) else lo, None if math.isnan(hi) else hi]}


def rates(c):
    """section 9 figures with Wilson 95 % from counts_for output."""
    g = lambda k: c.get(k, 0)
    return {
        "n": g("n"), "missing": g("missing"), "ungolded": g("ungolded"),
        "P_assign": _rate(g("assigned_correct"), g("assigned")),
        "R_assign": _rate(g("assigned_correct"), g("gold_covered")),
        "R_assign_res": _rate(g("assigned_correct"), g("gold_covered_res")),
        "P_new": _rate(g("none_correct"), g("none")),
        "R_new": _rate(g("none_correct"), g("gold_none")),
        "R_new_gap": _rate(g("none_correct_gap"), g("gold_none_gap")),
        "R_new_new": _rate(g("none_correct_new"), g("gold_none_new")),
        "unresolved_rate": _rate(g("unresolved"), g("n")),
        "accuracy": _rate(g("correct"), g("n")),
        "assigned": g("assigned"), "none": g("none"), "unresolved": g("unresolved"),
        "gold_covered": g("gold_covered"), "gold_covered_res": g("gold_covered_res"), "gold_none": g("gold_none"), "gold_none_gap": g("gold_none_gap"), "gold_none_new": g("gold_none_new"),
    }


# ------------------------------------------------------------------ cases
def _pr_key(u):
    return str(u.get("pr_key") if u.get("pr_key") is not None else (u.get("pr") if u.get("pr") is not None else u.get("commit")))


def build_cases(units, gold_rows, arm_rows):
    """units: unit rows; gold_rows: {case_id: gold} or list; arm_rows: {arm: [rows] | {case_id: row}}.
    -> list of {"case_id","pr_key","role","stratum","stratum4","set","control","unit","gold","out": {arm: outcome}}."""
    if isinstance(gold_rows, list):
        gold_rows = {g["case_id"]: g for g in gold_rows}
    by_arm = {}
    for arm, rows in arm_rows.items():
        by_arm[arm] = rows if isinstance(rows, dict) else {r["case_id"]: r for r in rows}
    cases = []
    for u in units:
        g = gold_rows.get(u["case_id"])
        s = u.get("stratum")
        s4 = "4" if (s == 3 and g and g.get("covered") is False and g.get("none_kind") == "new") else (str(s) if s is not None else "?")
        cases.append({"case_id": u["case_id"], "pr_key": _pr_key(u), "role": u.get("role"), "stratum": s, "stratum4": s4, "set": u.get("set"),
                      "control": bool(u.get("control")), "unit": u, "gold": g,
                      "out": {arm: score_outcome(rows.get(u["case_id"]), g) for arm, rows in by_arm.items()}})
    return cases


def scored_units(units):
    return [u for u in units if u.get("set") == "window" and not u.get("control")]


def paired(cases, jev, base="B2", resolved_only=False):
    """Paired comparison of arm `jev` against `base` on clusters where both have a measured outcome. resolved_only (R4-13 (3)): only clusters `jev` RESOLVED (output assigned | none)
    enter the pair; a Jev-abstained cluster is excluded from the pair, not scored as wrong for Jev (B2 is scored on exactly the same clusters)."""
    b = c = both = neither = 0
    jev_only_prs = set()
    per_pr = collections.OrderedDict()
    for k in cases:
        oj, ob = k["out"].get(jev), k["out"].get(base)
        if not oj or not ob or oj["correct"] is None or ob["correct"] is None:
            continue
        if resolved_only and oj["output"] not in RESOLVED:
            continue
        d = int(oj["correct"]) - int(ob["correct"])
        row = per_pr.setdefault(k["pr_key"], {"pr_key": k["pr_key"], "n": 0, "d": 0})
        row["n"] += 1
        row["d"] += d
        if d == 1:
            b += 1
            jev_only_prs.add(k["pr_key"])
        elif d == -1:
            c += 1
        elif oj["correct"]:
            both += 1
        else:
            neither += 1
    n = b + c + both + neither
    return {"n": n, "jev_only": b, "base_only": c, "both": both, "neither": neither, "gain": ((b - c) / n) if n else None,
            "mcnemar_p": mcnemar_exact(b, c), "jev_only_prs": sorted(jev_only_prs), "n_jev_only_prs": len(jev_only_prs), "per_pr": list(per_pr.values())}


def gain_res_stat(rows):
    """Paired gain on the RESOLVED clusters (condition d / e): sum d_res / sum n_res over per-PR pooling rows."""
    n = sum(r["n_res"] for r in rows)
    return (sum(r["d_res"] for r in rows) / n) if n else None


def gain_all_stat(rows):
    """All-cluster paired gain, every abstained cluster scored as Jev-wrong (reported beside the guard h, which is `gain_guess`): sum d_all / sum n_all."""
    n = sum(r["n_all"] for r in rows)
    return (sum(r["d_all"] for r in rows) / n) if n else None


def gain_guess_stat(rows):
    """R5-01 guard statistic: sum d_guess / sum n_all, d_guess = (guess_correct - B2 correct) per cluster; an abstained cluster is scored by Jev's own unthresholded guess."""
    n = sum(r["n_all"] for r in rows)
    return (sum(r["d_guess"] for r in rows) / n) if n else None


def gain_stat(rows):
    """Paired gain over `paired()` per-PR rows {pr_key, n, d} (descriptive tables, episodes; the GO rule's gain is gain_res_stat over pooling_rows)."""
    n = sum(r["n"] for r in rows)
    return (sum(r["d"] for r in rows) / n) if n else None


def p_assign_stat(rows):
    a = sum(r["a"] for r in rows)
    return (sum(r["ac"] for r in rows) / a) if a else None


def pooling_rows(cases, jev="V-sem", base="B2"):
    """Per-PR additive rows of Jev vs base over clusters with a measured Jev outcome: n_m measured, a assigned, ac assigned-correct, gc gold-covered, gc_res gold-covered and resolved, abst abstained
    (output unresolved, any reason); with a measured base outcome: n_all / d_all (every such cluster, d = jev correct - base correct; abstentions count as wrong), n_res / d_res (resolved clusters
    only) and b2_res (base-correct among them), abst_b2_n / abst_b2_correct (base outcomes on the Jev-abstained clusters)."""
    per = collections.OrderedDict()
    zero = ("n_m", "a", "ac", "gc", "gc_res", "abst", "n_all", "d_all", "d_guess", "n_res", "d_res", "b2_res", "abst_b2_n", "abst_b2_correct")
    for k in cases:
        oj, ob = k["out"].get(jev), k["out"].get(base)
        if not oj or oj["correct"] is None:
            continue
        row = per.setdefault(k["pr_key"], dict({"pr_key": k["pr_key"]}, **{f: 0 for f in zero}))
        res_ = oj["output"] in RESOLVED
        row["n_m"] += 1
        row["a"] += 1 if oj["output"] == "assigned" else 0
        row["ac"] += 1 if oj["kind"] == "assigned_correct" else 0
        row["gc"] += 1 if oj["gold_covered"] else 0
        row["gc_res"] += 1 if (oj["gold_covered"] and res_) else 0
        row["abst"] += 0 if res_ else 1
        if ob and ob["correct"] is not None:
            d = int(oj["correct"]) - int(ob["correct"])
            row["n_all"] += 1
            row["d_all"] += d
            row["d_guess"] += int(bool(oj.get("guess_correct"))) - int(ob["correct"])
            if res_:
                row["n_res"] += 1
                row["d_res"] += d
                row["b2_res"] += int(ob["correct"])
            else:
                row["abst_b2_n"] += 1
                row["abst_b2_correct"] += int(ob["correct"])
    return list(per.values())


def _lb(boot, level):
    """One-sided lower bound `level` of a pr_cluster_bootstrap result; None when there is no resamplable row (zero resolved clusters)."""
    return boot["lower"].get(level) if boot else None


def unresolved_block(outs):
    """Abstention figures over score_outcome dicts (R4-13 (2)): u = unresolved / measured clusters with its Wilson 95 % and the per-reason split; `missing` (no arm row) and `ungolded` are
    tallied, never counted in u (compose writes `unresolved: failure` for a missing request, so an absent arm row is an upstream defect that refuses earlier)."""
    meas = [o for o in outs if o["kind"] not in ("missing", "ungolded")]
    abst = [o for o in meas if o["output"] == "unresolved"]
    by = {r: sum(1 for o in abst if o.get("reason") == r) for r in REASONS}
    by["other"] = len(abst) - sum(by.values())
    r = _rate(len(abst), len(meas))
    return {"k": len(abst), "n": len(meas), "rate": r["value"], "wilson95": r["wilson95"], "by_reason": by,
            "missing": sum(1 for o in outs if o["kind"] == "missing"), "ungolded": sum(1 for o in outs if o["kind"] == "ungolded")}


def _gain_rows(cases, jev, base, B=10000, seed=x2d.SEED):
    """R5-08 report rows of one set of cases: `gain_res` (n_res, jev-only, B2-only, point, McNemar, 95 % CI, lb90, ub80, PRs with jev-only) and `gain_guess` (point, lb95, lb90, 95 % CI, counts)."""
    rows = [r for r in pooling_rows(cases, jev, base) if r["n_all"]]
    p_res = paired(cases, jev, base, resolved_only=True)
    br = pr_cluster_bootstrap([r for r in rows if r["n_res"]], gain_res_stat, B=B, seed=seed, levels=(0.8, 0.9, 0.95))
    bg = pr_cluster_bootstrap(rows, gain_guess_stat, B=B, seed=seed, levels=(0.8, 0.9, 0.95))
    gj = [k for k in cases if not k["control"] and k["out"][jev].get("guess_correct") is True and k["out"][base]["correct"] is False]
    gb = [k for k in cases if not k["control"] and k["out"][jev].get("guess_correct") is False and k["out"][base]["correct"] is True]
    return ({"n": p_res["n"], "jev_only": p_res["jev_only"], "base_only": p_res["base_only"], "point": p_res["gain"], "mcnemar_p": p_res["mcnemar_p"], "ci95": br["two_sided"].get(0.95),
             "lb90": _lb(br, 0.9), "ub80": br["upper"].get(0.8), "lb95": _lb(br, 0.95), "n_jev_only_prs": p_res["n_jev_only_prs"]},
            {"n": sum(r["n_all"] for r in rows), "jev_only": len(gj), "base_only": len(gb), "point": gain_guess_stat(rows) if rows else None, "mcnemar_p": mcnemar_exact(len(gj), len(gb)), "ci95": bg["two_sided"].get(0.95),
             "lb90": _lb(bg, 0.9), "ub80": bg["upper"].get(0.8), "lb95": _lb(bg, 0.95), "n_jev_only_prs": len({k["pr_key"] for k in gj})})


def abstention_table(rows):
    """The selection diagnostic (R4-13 (2), decides nothing): B2's accuracy on the Jev-abstained clusters, on the Jev-resolved clusters and on all, from pooling_rows."""
    ka, na = sum(r["abst_b2_correct"] for r in rows), sum(r["abst_b2_n"] for r in rows)
    kr, nr = sum(r["b2_res"] for r in rows), sum(r["n_res"] for r in rows)
    return {"n_abstained": sum(r["abst"] for r in rows), "b2_correct": ka, "b2_n_abstained": na, "b2_acc_abstained": _rate(ka, na), "b2_acc_resolved": _rate(kr, nr), "b2_acc_all": _rate(ka + kr, na + nr),
            "note": "descriptive selection diagnostic: B2's accuracy where Jev abstained vs where it answered; the guard h (gain_guess) decides, this does not"}


def summarise(arm_rows, gold_rows, units, jev="V-sem", base="B2", informative_all=None, perception=None, n1=None, B=10000, seed=x2d.SEED, cases=None):
    """Pooled summary consumed by go_verdict (R4-13). arm_rows {arm: rows}; units: the scored clusters (pooled over rounds); gold_rows {case_id: gold}.
    Keys: p_assign (assigned-correct / assigned, point and one-sided 90 % Wilson bound), r_assign_res (assigned-correct / gold-covered AND resolved), r_assign (all gold-covered, abstention = miss; descriptive),
    gain_res (paired gain on the RESOLVED clusters, the same clusters for both arms; point, 90 % PR-cluster-bootstrap lower bound, PR spread), gain_all (every abstention scored as a miss; reported), gain_guess (R5-01: an abstained cluster scored by Jev's own unthresholded guess; point, 95 % PR-cluster-bootstrap lower bound `lb95`: the guard h), n_res, unresolved
    {k, n, rate, wilson95, by_reason}, abstained (B2's accuracy where Jev abstained / answered), real_new / planted_new (descriptive, Wilson 95 %), perception, n1, informative_all.
    perception / n1: {"pass": bool|None,...} or None (unmeasured). Every numerator is read with .get (an arm or subset with no assignment still summarises, R4-10)."""
    cases = cases if cases is not None else build_cases(units, gold_rows, {jev: arm_rows[jev], base: arm_rows[base]})
    real = [k for k in cases if not k["control"]]
    cj = counts_for([k["out"][jev] for k in real])
    p_res, p_all = paired(real, jev, base, resolved_only=True), paired(real, jev, base)
    rows = pooling_rows(real, jev, base)
    boot_res = pr_cluster_bootstrap([r for r in rows if r["n_res"]], gain_res_stat, B=B, seed=seed, levels=(0.8, 0.9, 0.95)) if rows else None
    boot_all = pr_cluster_bootstrap([r for r in rows if r["n_all"]], gain_all_stat, B=B, seed=seed, levels=(0.8, 0.9, 0.95)) if rows else None
    boot_guess = pr_cluster_bootstrap([r for r in rows if r["n_all"]], gain_guess_stat, B=B, seed=seed, levels=(0.8, 0.9, 0.95)) if rows else None
    gg_rows = [r for r in rows if r["n_all"]]
    gg_point = gain_guess_stat(gg_rows) if gg_rows else None
    gg_jev = sum(1 for k in real if k["out"][jev].get("guess_correct") is True and k["out"][base]["correct"] is False)
    gg_base = sum(1 for k in real if k["out"][jev].get("guess_correct") is False and k["out"][base]["correct"] is True)
    gg_n = sum(r["n_all"] for r in gg_rows)
    ac, na = cj.get("assigned_correct", 0), cj.get("assigned", 0)
    pa_lo, _pa_hi = wilson_bounds_one_sided(ac, na, GO["level"]) if na else (None, None)
    gc, gcr = cj.get("gold_covered", 0), cj.get("gold_covered_res", 0)
    gn_none, nn = cj.get("none_correct", 0), cj.get("none", 0)
    ctl = [k for k in cases if k["control"] and jev in k["out"]]
    cp = counts_for([k["out"][jev] for k in ctl])
    pooled = {
        "jev_arm": jev, "base_arm": base, "n": cj.get("n", 0), "n_prs": len({k["pr_key"] for k in real}),
        "p_assign": {"k": ac, "n": na, "point": (ac / na) if na else None, "wilson_lb90": pa_lo},
        "r_assign_res": {"k": ac, "n": gcr, "point": (ac / gcr) if gcr else None},
        "r_assign": {"k": ac, "n": gc, "point": (ac / gc) if gc else None},
        "gain_res": {"point": p_res["gain"], "n": p_res["n"], "lb90": _lb(boot_res, 0.9), "ub80": boot_res["upper"].get(0.8) if boot_res else None,
                     "ci95": boot_res["two_sided"].get(0.95) if boot_res else None, "n_jev_only_prs": p_res["n_jev_only_prs"], "jev_only": p_res["jev_only"], "base_only": p_res["base_only"], "mcnemar_p": p_res["mcnemar_p"]},
        "gain_all": {"point": p_all["gain"], "n": p_all["n"], "lb90": _lb(boot_all, 0.9), "ci95": boot_all["two_sided"].get(0.95) if boot_all else None,
                     "jev_only": p_all["jev_only"], "base_only": p_all["base_only"], "mcnemar_p": p_all["mcnemar_p"]},
        "gain_guess": {"point": gg_point, "n": gg_n, "lb95": _lb(boot_guess, 0.95), "lb90": _lb(boot_guess, 0.9), "ci95": boot_guess["two_sided"].get(0.95) if boot_guess else None,
                       "jev_only": gg_jev, "base_only": gg_base, "mcnemar_p": mcnemar_exact(gg_jev, gg_base)},
        "n_res": p_res["n"],
        "unresolved": unresolved_block([k["out"][jev] for k in real]),
        "abstained": abstention_table(rows),
        "real_new": {"n": cj.get("gold_none", 0), "n_gap": cj.get("gold_none_gap", 0), "n_new": cj.get("gold_none_new", 0),
                     "P_new": (gn_none / nn) if nn else None, "P_new_k": gn_none, "P_new_n": nn, "P_new_wilson95": _rate(gn_none, nn)["wilson95"],
                     "R_new": (gn_none / cj["gold_none"]) if cj.get("gold_none") else None, "R_new_wilson95": _rate(gn_none, cj.get("gold_none", 0))["wilson95"],
                     "note": "n real positives; no veto on the verdict"},
        "planted_new": {"n": cp.get("gold_none", 0), "P_new": _rate(cp.get("none_correct", 0), cp.get("none", 0)), "R_new": _rate(cp.get("none_correct", 0), cp.get("gold_none", 0))},
        "perception": perception if perception is not None else {"pass": None, "note": "unmeasured"},
        "n1": n1 if n1 is not None else {"pass": None, "note": "unmeasured"},
        "informative_all": informative_all,
    }
    return pooled


def go_verdict(pooled, thresholds):
    """GO / INCONCLUSIVE / NO-GO (section 9, thresholds.json study.go; R4-13 resolved-cluster rule), pure over a `summarise` dict. Precedence (R1-19):
    1. a frozen round not informative (`informative_all` not True) -> INCONCLUSIVE, before any point test;
    2. a point condition (a, c, d, g) fails -> NO-GO; perception or N1 measured and failed -> NO-GO;
    3. everything passes -> GO; otherwise (a bound b / e, the PR spread f, the abstention guard h, resolved clusters i < 50, perception / N1 unmeasured) -> INCONCLUSIVE.
    a P_assign point >= 0.90 (assigned-correct / assigned); b its one-sided 90 % Wilson lower bound >= 0.80; c R_assign_res point >= 0.80 (assigned-correct / gold-covered AND resolved);
    d paired gain on the resolved clusters >= +0.10; e its one-sided 90 % PR-cluster-bootstrap lower bound > 0; f Jev-correct/B2-wrong resolved clusters from >= 3 PRs; g unresolved rate <= 0.15 (the cap);
    h guard: one-sided 95 % PR-cluster-bootstrap lower bound of gain_guess > 0 (abstained clusters answered by Jev's own unthresholded guess, R5-01); i resolved paired clusters >= 50; j perception passes; k N1 at chance; l every frozen round informative.
    P_new / R_new are reported (summarise real_new), never a condition."""
    mins = thresholds["minimums"]
    f = lambda x: x is not None and not (isinstance(x, float) and math.isnan(x))
    pa, rr, gr, ga, gg, un = pooled["p_assign"], pooled["r_assign_res"], pooled["gain_res"], pooled["gain_all"], pooled["gain_guess"], pooled["unresolved"]
    cond = collections.OrderedDict()

    def add(name, value, ok, kind):
        cond[name] = {"value": value, "pass": bool(ok), "kind": kind}

    add("P_assign point >= %.2f" % mins["P_assign"], pa["point"], f(pa["point"]) and pa["point"] >= mins["P_assign"], "point")
    add("R_assign_res point >= %.2f (resolved covered clusters)" % mins["R_assign"], rr["point"], f(rr["point"]) and rr["point"] >= mins["R_assign"], "point")
    add("gain_res point >= +%.2f (resolved clusters)" % GO["gain_min"], gr["point"], f(gr["point"]) and gr["point"] >= GO["gain_min"] - 1e-12, "point")
    add("unresolved rate <= %.2f (cap)" % GO["unresolved_rate_max"], un["rate"], f(un["rate"]) and un["rate"] <= GO["unresolved_rate_max"] + 1e-12, "point")
    add("P_assign one-sided 90%% Wilson lower bound >= %.2f" % GO["p_assign_wilson_lb"], pa["wilson_lb90"], f(pa["wilson_lb90"]) and pa["wilson_lb90"] >= GO["p_assign_wilson_lb"], "bound")
    add("gain_res one-sided 90% bootstrap lower bound > 0", gr["lb90"], f(gr["lb90"]) and gr["lb90"] > 0, "bound")
    add("Jev-correct/B2-wrong resolved clusters from >= %d distinct PRs" % GO["spread_min"], gr["n_jev_only_prs"], gr["n_jev_only_prs"] >= GO["spread_min"], "spread")
    add("gain_guess one-sided %d%% bootstrap lower bound > 0 (abstentions answered by Jev's own unthresholded guess; guard)" % round(GO["guard_level"] * 100), gg["lb95"], f(gg["lb95"]) and gg["lb95"] > 0, "guard")
    add("resolved paired clusters >= %d" % GO["resolved_min"], pooled["n_res"], pooled["n_res"] >= GO["resolved_min"], "data")
    percep, n1 = pooled["perception"].get("pass"), pooled["n1"].get("pass")
    add("perception check passes", percep, percep is True, "other")
    add("N1 at chance", n1, n1 is True, "other")
    inf_ok = pooled.get("informative_all") is True
    add("every frozen round informative", pooled.get("informative_all"), inf_ok, "other")
    points_ok = all(v["pass"] for v in cond.values() if v["kind"] == "point")
    all_ok = all(v["pass"] for v in cond.values())
    reasons = []
    if not inf_ok:
        verdict = "INCONCLUSIVE"
        reasons.append("a frozen round is not informative (decided before any point test)")
    elif not points_ok:
        verdict = "NO-GO"
        reasons.append("a point condition failed: " + "; ".join(k for k, v in cond.items() if v["kind"] == "point" and not v["pass"]))
    elif percep is False or n1 is False:
        verdict = "NO-GO"
        reasons.append("perception check or N1 measured and failed")
    elif all_ok:
        verdict = "GO"
    else:
        verdict = "INCONCLUSIVE"
    if verdict != "NO-GO":
        reasons += [(k + ": Jev shows no accuracy edge over B2 once its abstentions are answered by its own unthresholded guess (the resolved-cluster gain may be selection)") if v["kind"] == "guard" else k for k, v in cond.items() if not v["pass"]]
    return {"verdict": verdict, "conditions": cond, "reasons": reasons}


# ------------------------------------------------------------------ round informativeness and futility
def informative(round_info):
    """round_info: {"n_clusters", "n_prs", "kappa_answer", + reported: "kappa_covered", "raw_agreement_covered", "prevalence_covered", "labels_check_ok"}; informative iff >= 20 clusters, >= 6 PRs and
    kappa >= 0.6 on the JOINT ANSWER (none / area, acceptable-overlap mapped to one label) only (R1-06) over the GO population (R2-14), and the round's labels-check has not failed (rounds 8-10 cannot
    clarify the guide: a failed labels-check marks the round non-informative, R2-03): kappa on `covered` alone is degenerate at high prevalence, so it is reported with raw agreement and
    prevalence and decides nothing."""
    th = {"min_clusters": 20, "min_prs": 6, "min_kappa": x2d.MIN_KAPPA, "kappa_basis": "joint answer"}
    k = round_info.get("kappa_answer")
    ok_k = k is not None and k >= th["min_kappa"]
    lc_ok = round_info.get("labels_check_ok") is not False
    ok = round_info.get("n_clusters", 0) >= th["min_clusters"] and round_info.get("n_prs", 0) >= th["min_prs"] and ok_k and lc_ok
    return {"informative": bool(ok), "n_clusters_ok": round_info.get("n_clusters", 0) >= th["min_clusters"], "n_prs_ok": round_info.get("n_prs", 0) >= th["min_prs"],
            "kappa_ok": ok_k, "labels_check_ok": lc_ok, "kappa_answer": k, "kappa_covered": round_info.get("kappa_covered"), "raw_agreement_covered": round_info.get("raw_agreement_covered"),
            "prevalence_covered": round_info.get("prevalence_covered"), "thresholds": th}


def futility(round_summaries, minimum=None, B=10000, seed=x2d.SEED):
    """round_summaries: chronological [{"round", "informative": bool, "integrity_pause": bool, "pooling": [per-PR rows of pooling_rows]}].
    Integrity-pause rounds are skipped (not counted, do not break a run); a non-informative round breaks a run. Fires when some run of three consecutive informative
    rounds has pooled gain_res (resolved clusters, R4-13 (5)) 80 % upper bound <= +0.02 AND pooled P_assign 80 % upper bound < the P_assign minimum. No disjunct on the unresolved rate: u and
    the all-cluster gain are REPORTED per window (the per-round critique sees a drift towards the cap) and never fire the rule.
    -> {"fired": bool, "windows": [...], "counted_rounds": [...]}."""
    minimum = x2d.MINIMUMS["P_assign"] if minimum is None else minimum
    counted = [s for s in round_summaries if not s.get("integrity_pause")]
    windows = []
    for i in range(len(counted) - FUTILITY["run"] + 1):
        w = counted[i:i + FUTILITY["run"]]
        if not all(s.get("informative") for s in w):
            continue
        rows = []
        for s in w:
            for r in s["pooling"]:
                rows.append(dict(r, pr_key="%s:%s" % (s["round"], r["pr_key"])))
        gb = pr_cluster_bootstrap([r for r in rows if r["n_res"]], gain_res_stat, B=B, seed=seed, levels=(0.8,))
        pb = pr_cluster_bootstrap([r for r in rows if r["a"]], p_assign_stat, B=B, seed=seed, levels=(0.8,))
        gub, pub = gb["upper"].get(0.8), pb["upper"].get(0.8)
        hit = gub is not None and pub is not None and not math.isnan(gub) and not math.isnan(pub) and gub <= FUTILITY["gain_upper_max"] and pub < minimum
        n_m, n_all = sum(r["n_m"] for r in rows), sum(r["n_all"] for r in rows)
        windows.append({"rounds": [s["round"] for s in w], "gain_res_point": gb["point"], "gain_res_ub80": gub, "p_assign_point": pb["point"], "p_assign_ub80": pub,
                        "unresolved_rate": (sum(r["abst"] for r in rows) / n_m) if n_m else None, "gain_all_point": gain_all_stat(rows) if n_all else None, "fires": bool(hit)})
    return {"fired": any(w["fires"] for w in windows), "windows": windows, "counted_rounds": [s["round"] for s in counted], "minimum": minimum}


# ------------------------------------------------------------------ labels
def _sessions(labels, kind="wide", label_set="main", question=None):
    """{key: {"A": row, "B": row, "C": row, "final": row}} -- last row wins per (key, labeller) (labels are append-only; a relabel appends).
    key = case_id for wide / intruder rows; for kind "jev-state" (two rows per unit and labeller, one per question) key = (case_id, question), so the two questions never collide
    (R1-01). `question` restricts jev-state rows to one question."""
    out = {}
    for r in labels:
        if r.get("kind") != kind or (r.get("label_set") or "main") != label_set:
            continue
        if question is not None and r.get("question") != question:
            continue
        by = r.get("by") or ""
        slot = "final" if by in ("A+B agree", "C adjudicated") else ("A" if by.startswith("labeller_A") else "B" if by.startswith("labeller_B") else "C" if by.startswith("adjudicator_C") else None)
        if slot:
            key = (r["case_id"], r.get("question")) if kind == "jev-state" else r["case_id"]
            out.setdefault(key, {})[slot] = r
    return out


def _acc(row):
    return set(row.get("acceptable") or [])


def _pair_agree(a, b):
    """The contract's agreement rule (label_x2d.agree): covered equal; covered -> acceptable sets share >= 1 id; not covered -> none_kind equal."""
    if a.get("covered") is None or b.get("covered") is None or a["covered"] != b["covered"]:
        return False
    if a["covered"]:
        return bool(_acc(a) & _acc(b))
    return (a.get("none_kind") or None) == (b.get("none_kind") or None)


def _answer_labels(a, b):
    if not a["covered"] or not b["covered"]:
        return ("none" if not a["covered"] else (a.get("primary") or min(_acc(a), default="?")), "none" if not b["covered"] else (b.get("primary") or min(_acc(b), default="?")))
    ov = _acc(a) & _acc(b)
    if ov:
        m = min(ov)
        return m, m
    return a.get("primary") or min(_acc(a), default="?"), b.get("primary") or min(_acc(b), default="?")


def label_agreement(labels, case_ids=None, kind="wide", label_set="main"):
    """A vs B before adjudication on `case_ids` (default all): kappa on covered, kappa on the joint answer (none / area, acceptable-overlap mapped to a common label),
    acceptable-set agreement (the agree rule) and mean Jaccard among both-covered units, adjudication rate over final rows, rule-citation concentration."""
    ss = _sessions(labels, kind, label_set)
    ids = [c for c in ss if (case_ids is None or c in case_ids)]
    both = [c for c in ids if "A" in ss[c] and "B" in ss[c] and ss[c]["A"].get("covered") is not None and ss[c]["B"].get("covered") is not None]
    unparsable = [c for c in ids if c not in both]
    cov_a = [ss[c]["A"]["covered"] for c in both]
    cov_b = [ss[c]["B"]["covered"] for c in both]
    ans = [_answer_labels(ss[c]["A"], ss[c]["B"]) for c in both]
    agree_rule = [c for c in both if _pair_agree(ss[c]["A"], ss[c]["B"])]
    cc = [c for c in both if ss[c]["A"]["covered"] and ss[c]["B"]["covered"]]
    jac = [len(_acc(ss[c]["A"]) & _acc(ss[c]["B"])) / max(1, len(_acc(ss[c]["A"]) | _acc(ss[c]["B"]))) for c in cc]
    finals = [ss[c]["final"] for c in ids if "final" in ss[c]]
    adj = sum(1 for r in finals if r.get("agreement") == "adjudicated" or r.get("by") == "C adjudicated")
    raw_rules = collections.Counter()  # the raw rule strings, kept for the record
    cited = collections.Counter()  # R2-08: each rule id R<n> counted at most once per A/B session
    parsed = 0
    for c in ids:
        for s in ("A", "B"):
            row = ss[c].get(s)
            if row and row.get("rule"):
                raw_rules[str(row["rule"])] += 1
                found = {m.upper() for m in re.findall(r"R\d+", str(row["rule"]), re.I)}
                if found:
                    parsed += 1
                    cited.update(found)
    top = sorted(cited.items(), key=lambda kv: (-kv[1], kv[0]))[:1]
    top_share = (top[0][1] / parsed) if top else None
    return {"n_units": len(ids), "n_both_parsed": len(both), "n_unparsable_or_missing": len(unparsable),
            "kappa_covered": cohen_kappa(cov_a, cov_b), "kappa_answer": cohen_kappa([x for x, _ in ans], [y for _, y in ans]),
            "raw_agreement_covered": (sum(1 for x, y in zip(cov_a, cov_b) if x == y) / len(both)) if both else None,
            "prevalence_covered": ((sum(cov_a) + sum(cov_b)) / (2 * len(both))) if both else None,
            "raw_agreement_answer": (sum(1 for x, y in ans if x == y) / len(both)) if both else None,
            "agree_rule_rate": (len(agree_rule) / len(both)) if both else None,
            "acceptable_set_overlap_rate": (sum(1 for c in cc if _acc(ss[c]["A"]) & _acc(ss[c]["B"])) / len(cc)) if cc else None,
            "mean_jaccard": (sum(jac) / len(jac)) if jac else None,
            "adjudication_rate": (adj / len(finals)) if finals else None, "n_final": len(finals),
            "rule_counts": dict(raw_rules), "rule_session_counts": dict(cited), "rule_sessions_parsed": parsed, "rule_top": top[0][0] if top else None, "rule_top_share": top_share,
            "rule_concentration_trigger": bool(top and top_share > RULE_CONC)}


def _combined(q1, q2):
    """One labeller's (or the consensus') COMBINED answer from its covered-question row and its which_area row: "none" if not covered, else the which_area primary; None when unusable."""
    if not q1 or q1.get("covered") is None:
        return None
    if q1["covered"] is False:
        return "none"
    return (q2 or {}).get("primary") or None


def _jev_combined(o):
    """An arm's THRESHOLDED output as a combined answer: area id | "none" | "unresolved" (never equal to a labeller's answer)."""
    out = (o or {}).get("output")
    return o.get("area_id") if out == "assigned" else (out if out in ("none", "unresolved") else None)


def _jev_raw_combined(o):
    """The V-sem arm's UNTHRESHOLDED combined answer (R2-09): "none" iff not raw_covered, else raw_choice; None when the arm row has no raw answer (missing, or a request failed)."""
    if not o or o.get("raw_covered") is None:
        return None
    return "none" if o["raw_covered"] is False else (o.get("raw_choice") or None)


def perception_counts(labels, cases, jev="V-sem", primary=None):
    """Perception check inputs for one round: jev-state labels (two rows per unit and labeller, keyed (case_id, question)) vs the wide gold and the V-sem arm `jev` (always V-sem: the
    labellers saw the V-sem state, R2-17). The headline agreement compares the UNTHRESHOLDED combined answer (R2-09); reported beside it: the thresholded agreement, the unresolved count and the
    `primary` arm's thresholded agreement. Counts only; the pass decision is pooled (perception_pass). A unit counts only when BOTH questions have a final row and the combined consensus answer
    is defined (R1-01). -> {"n", "jev_agree_k", "jev_thresholded_k", "unresolved_k", "primary_k", "ll_n", "ll_agree_k", "state_sufficiency_k", "q1": {"n","jev_k","ll_n","ll_k"}, "q2": {...}};
    q1 = covered question (raw_covered equals the consensus), q2 = which_area primary (units whose consensus is covered; raw_choice equals it)."""
    s1, s2 = _sessions(labels, "jev-state", "main", "merge.covered"), _sessions(labels, "jev-state", "main", "merge.which_area")
    n = suff = jev_k = thr_k = unr_k = prim_k = ll_n = ll_k = 0
    q1 = {"n": 0, "jev_k": 0, "ll_n": 0, "ll_k": 0}
    q2 = {"n": 0, "jev_k": 0, "ll_n": 0, "ll_k": 0}
    for k in cases:
        if k["control"]:
            continue
        a1, a2 = s1.get((k["case_id"], "merge.covered")), s2.get((k["case_id"], "merge.which_area"))
        if not a1 or not a2 or "final" not in a1 or "final" not in a2:
            continue
        cons = _combined(a1["final"], a2["final"])
        if cons is None:
            continue
        n += 1
        g = k["gold"]
        if g and isinstance(g.get("covered"), bool):
            suff += 1 if ((cons == "none") if not g["covered"] else (cons in _acc(g))) else 0
        o = k["out"].get(jev)
        ja = _jev_raw_combined(o)
        jev_k += 1 if (ja is not None and ja == cons) else 0
        tj = _jev_combined(o)
        thr_k += 1 if (tj is not None and tj == cons) else 0
        unr_k += 1 if (o and o.get("output") == "unresolved") else 0
        if primary is not None:
            pj = _jev_combined(k["out"].get(primary))
            prim_k += 1 if (pj is not None and pj == cons) else 0
        q1["n"] += 1
        q1["jev_k"] += 1 if (o and o.get("raw_covered") is not None and o["raw_covered"] == a1["final"]["covered"]) else 0
        if cons != "none":
            q2["n"] += 1
            q2["jev_k"] += 1 if (o and o.get("raw_choice") == cons) else 0
        ca, cb = _combined(a1.get("A"), a2.get("A")), _combined(a1.get("B"), a2.get("B"))
        if ca is not None and cb is not None:
            ll_n += 1
            ll_k += 1 if ca == cb else 0
        if a1.get("A") and a1.get("B") and a1["A"].get("covered") is not None and a1["B"].get("covered") is not None:
            q1["ll_n"] += 1
            q1["ll_k"] += 1 if a1["A"]["covered"] == a1["B"]["covered"] else 0
        if a2.get("A") and a2.get("B") and a2["A"].get("primary") and a2["B"].get("primary"):
            q2["ll_n"] += 1
            q2["ll_k"] += 1 if a2["A"]["primary"] == a2["B"]["primary"] else 0
    return {"n": n, "state_sufficiency_k": suff, "jev_agree_k": jev_k, "jev_thresholded_k": thr_k, "unresolved_k": unr_k, "primary_k": prim_k, "primary_arm": primary,
            "ll_n": ll_n, "ll_agree_k": ll_k, "q1": q1, "q2": q2}


def perception_pass(parts):
    """Pooled perception verdict from per-round perception_counts: pass iff n (units with both questions final) >= 20 and the V-sem UNTHRESHOLDED jev_agree >= ll_agree - 0.10; None
    (unmeasured) when n < 20 or no labeller pairs. The thresholded agreement, the unresolved rate and the primary arm's agreement are reported, never decide. Q1 and Q2 agreement are
    reported separately (no separate pass rule)."""
    n = sum(p["n"] for p in parts)
    jk = sum(p["jev_agree_k"] for p in parts)
    lln = sum(p["ll_n"] for p in parts)
    llk = sum(p["ll_agree_k"] for p in parts)
    sk = sum(p["state_sufficiency_k"] for p in parts)
    tk, uk, pk = (sum(p.get(f, 0) for p in parts) for f in ("jev_thresholded_k", "unresolved_k", "primary_k"))
    res = {"n": n, "jev_agree": (jk / n) if n else None, "ll_n": lln, "ll_agree": (llk / lln) if lln else None, "state_sufficiency": (sk / n) if n else None,
           "jev_agree_thresholded": (tk / n) if n else None, "unresolved_rate": (uk / n) if n else None, "primary_agree": (pk / n) if n else None,
           "primary_arm": next((p.get("primary_arm") for p in parts if p.get("primary_arm")), None), "margin": GO["perception_margin"], "pass": None}
    for q in ("q1", "q2"):
        tot = {f: sum(p.get(q, {}).get(f, 0) for p in parts) for f in ("n", "jev_k", "ll_n", "ll_k")}
        res[q] = {"n": tot["n"], "jev_agree": (tot["jev_k"] / tot["n"]) if tot["n"] else None, "ll_n": tot["ll_n"], "ll_agree": (tot["ll_k"] / tot["ll_n"]) if tot["ll_n"] else None}
    if n >= GO["perception_min_n"] and lln:
        res["pass"] = bool(res["jev_agree"] >= res["ll_agree"] - GO["perception_margin"] - 1e-12)
    else:
        res["note"] = "unmeasured: needs >= %d jev-state units with both questions final and labeller pairs" % GO["perception_min_n"]
    return res


# ------------------------------------------------------------------ loading
def _rd(r, data=None):
    return x2d.round_dir(r, data)


def _need(path, what):
    if not os.path.exists(path) or (os.path.isfile(path) and os.path.getsize(path) == 0):
        sys.exit("missing input (%s): %s" % (what, path))
    return path


def _load_json(path):
    with open(path) as f:
        return json.load(f)


def _arm_rows(d, arm, sub="arms", tag=None):
    for name in ([arm] + (["reference"] if arm == "reference-claude" else [])):
        p = os.path.join(d, sub, name + ".jsonl")
        if os.path.exists(p):
            return x2d.read_jsonl(p)
    return None


def current_labels(rows, guide_version):
    """R2-01 (5): wide label rows of the round's CURRENT guide version only (a clarification relabels under a new version; earlier rows stay on file). jev-state and intruder rows carry no guide."""
    return [l for l in rows if l.get("kind") != "wide" or l.get("guide_version") == guide_version]


def load_round(r, data=None, tag=None):
    """-> dict(units, gold, labels, arms {arm: rows}, lookups, dir). Missing arm files are absent from `arms` (unmeasured). `labels` = the current guide version's rows (current_labels)."""
    d = _rd(r, data)
    sfx = "." + tag if tag else ""
    units = x2d.read_jsonl(_need(os.path.join(d, "units%s.jsonl" % sfx), "units"))
    gold = {g["case_id"]: g for g in x2d.read_jsonl(os.path.join(d, "gold%s.jsonl" % sfx))}
    labels = current_labels(x2d.read_jsonl(os.path.join(d, "labels%s.jsonl" % sfx)), x2d.guide_version(r, data))
    arms = {}
    for arm in (CONTRAST_ARMS if tag else ARMS):
        rows = _arm_rows(d, arm)
        if rows is not None:
            arms[arm] = rows
    return {"units": units, "gold": gold, "labels": labels, "arms": arms, "lookups": x2d.read_jsonl(os.path.join(d, "lookups.jsonl")), "dir": d}


def _arm_summary(cases, arm):
    real = [k for k in cases if not k["control"] and arm in k["out"]]
    planted = [k for k in cases if k["control"] and arm in k["out"]]
    return {"real": rates(counts_for([k["out"][arm] for k in real])), "planted": rates(counts_for([k["out"][arm] for k in planted])) if planted else None}


def _breakdown(cases, arm, keyfn):
    groups = collections.defaultdict(list)
    for k in cases:
        if not k["control"] and arm in k["out"]:
            groups[str(keyfn(k))].append(k["out"][arm])
    res = {}
    for g, outs in sorted(groups.items()):
        r = rates(counts_for(outs))
        res[g] = {"n": r["n"], "accuracy": r["accuracy"], "P_assign": r["P_assign"], "R_assign": r["R_assign"], "unresolved": r["unresolved"]}
    return res


def _confusion(cases, arm):
    c = collections.Counter()
    for k in cases:
        o = k["out"].get(arm)
        if k["control"] or not o or o["correct"] is not False or o["kind"] == "unresolved":
            continue
        g = k["gold"]
        gl = "none" if not g["covered"] else (g.get("primary") or min(g.get("acceptable") or ["?"]))
        c[(gl, o["area_id"] if o["output"] == "assigned" else "none")] += 1
    pairs = [{"gold": a, "answered": b, "n": n} for (a, b), n in c.most_common()]
    return {"errors": sum(c.values()), "pairs": pairs, "pairs_repeated": sum(1 for p in pairs if p["n"] >= 2), "errors_in_repeated_pairs": sum(p["n"] for p in pairs if p["n"] >= 2)}


def _end_to_end(cases, lookups, audit_units, gold, arms=("V-sem", "B2"), total=None):
    """Files resolved by rule (a) `modified_owned` count as exact lookups; `modified_unowned` (pooled / never_owned) files are reported, in no accuracy (R1-12);
    renamed_inherit / variant_inherit / (c) / (d) files are inherited (corrected by the audit's false-inherit estimate). `total` = rounds/<r>/counts.json (build_units writes the round's
    rule counts there): the lookups rows must sum to it, key by key, or a commit's counts were dropped (R2-07; ValueError)."""
    if total is not None:
        have = collections.Counter()
        for row in lookups:
            have.update({str(k): int(v) for k, v in (row.get("counts") or {}).items()})
        have = {k: v for k, v in have.items() if v}
        want = {str(k): int(v) for k, v in total.items() if int(v)}
        if have != want:
            raise ValueError("lookups.jsonl rows sum to %s but counts.json says %s: a commit's rule counts were dropped" % (json.dumps(have, sort_keys=True), json.dumps(want, sort_keys=True)))
    exact = inherit = 0
    unowned = collections.Counter()
    for row in lookups:
        for key, v in (row.get("counts") or {}).items():
            k = str(key).lower()
            if k == "modified_owned" or k.startswith("(a)") or k == "a":
                exact += int(v)
            elif k.startswith("modified_unowned"):
                unowned[k] += int(v)
            elif k.startswith("renamed_inherit") or k.startswith("variant_inherit") or k.startswith("(c)") or k.startswith("(d)") or k in ("c", "d"):
                inherit += int(v)
    items = {a["case_id"]: a for row in lookups for a in (row.get("audit") or [])}
    wrong = measured = 0
    for cid, a in items.items():
        g = gold.get(cid)
        if g is None or not isinstance(g.get("covered"), bool):
            continue
        measured += 1
        if not (g["covered"] and a.get("inherited_area_id") in set(g.get("acceptable") or [])):
            wrong += 1
    fi = _rate(wrong, measured)
    f = fi["value"] or 0.0
    res = {"rule_exact_files": exact, "rule_inherit_files": inherit, "modified_unowned": dict(unowned), "modified_unowned_total": int(unowned.get("modified_unowned", 0)),
           "audit_items": len(items), "audit_measured": measured, "audit_unmeasured": len(items) - measured, "false_inherit": fi, "arms": {}}
    for arm in arms:
        w = ok = unm = 0
        for k in cases:
            o = k["out"].get(arm)
            if k["control"] or not o or o["correct"] is None:
                unm += 1
                continue
            m = len(k["unit"].get("members") or [1])
            w += m
            ok += m * int(o["correct"])
        tot = exact + inherit + w
        res["arms"][arm] = {"files_total": tot, "files_correct_est": exact + inherit * (1 - f) + ok, "end_to_end_accuracy": ((exact + inherit * (1 - f) + ok) / tot) if tot else None,
                            "residual_files": w, "residual_accuracy_file_weighted": (ok / w) if w else None, "unmeasured_clusters": unm}
    return res


def _divergence(r, data, arms):
    """Replayed free-running divergence (PREREGISTRATION 10): for each arm, gold.divergence(gold.next_membership(r), gold.free_membership(r, arm)). Per-arm status
    measured | unmeasured (with the reason); never silently dropped."""
    res = {"name": "replayed free-running divergence", "arms": {}}
    try:
        import gold as G
    except Exception as e:  # gold.py is another module's; absent -> unmeasured
        res["status"] = "unmeasured"
        res["reason"] = "gold module not importable: %s" % e
        return res
    for arm in arms:
        try:
            res["arms"][arm] = {"status": "measured", "value": G.divergence(G.next_membership(r, data), G.free_membership(r, arm, data))}
        except Exception as e:
            res["arms"][arm] = {"status": "unmeasured", "reason": "%s: %s" % (type(e).__name__, e)}
    res["status"] = "measured" if res["arms"] and all(v["status"] == "measured" for v in res["arms"].values()) else "partly unmeasured"
    return res


def _episodes(cases, jev="V-sem", base="B2"):
    res = {}
    for pr in EPISODES:
        mine = [k for k in cases if k["pr_key"] == str(pr) and not k["control"]]
        rest = [k for k in cases if k["pr_key"] != str(pr) and not k["control"]]
        if not mine:
            continue
        pj, pr_rest = paired(mine, jev, base), paired(rest, jev, base)
        res[str(pr)] = {"n_clusters": len(mine), "with": {"P_assign": rates(counts_for([k["out"][jev] for k in cases if not k["control"]]))["P_assign"]["value"], "gain": paired([k for k in cases if not k["control"]], jev, base)["gain"]},
                        "without": {"P_assign": rates(counts_for([k["out"][jev] for k in rest]))["P_assign"]["value"], "gain": pr_rest["gain"], "n": pr_rest["n"]},
                        "episode_only": {"P_assign": rates(counts_for([k["out"][jev] for k in mine]))["P_assign"]["value"], "gain": pj["gain"], "n": pj["n"]}}
    return res


def _jev_scores(d, sub="jev", name="scores.jsonl"):
    return x2d.read_jsonl(os.path.join(d, sub, name))


def _go_population(cases):
    """The GO population (R1-14): scored, non-control cases without stratum_reason == "core_gate" (those stay in the stratum report)."""
    return [k for k in cases if not k["control"] and (k["unit"].get("stratum_reason") != "core_gate")]


def _output_rates(rows):
    c = collections.Counter((r.get("output") or "missing") for r in rows)
    n = len(rows)
    return {"n": n, **{k: {"k": c.get(k, 0), "rate": (c.get(k, 0) / n) if n else None} for k in ("assigned", "none", "unresolved")}}


def _same_output(a, b):
    """Two arm rows give the same answer: same output and, for an assignment, the same area."""
    if not a or not b or a.get("output") is None or b.get("output") is None:
        return None
    return a["output"] == b["output"] and (a.get("area_id") if a["output"] == "assigned" else None) == (b.get("area_id") if b["output"] == "assigned" else None)


def controls_section(units, gold, arm_rows, primary):
    """Per-round controls (R1-07, R2-10), report-only: units with control true and control_kind dup | trap | prose.
    dup (test-retest of a window unit): gold agreement with the source (covered equal; acceptable overlap; primary equal among both-covered) and primary-arm output equality with the source;
    trap (empty-file rename): gold covered = false rate and the arm's none rate; prose (misleading PR prose, same unit as its source): primary-arm output equals the source's.
    Every rate is {k, n, value, wilson95}; units lacking gold or an arm row are counted in `unmeasured`, never in a rate."""
    by = {u["case_id"]: u for u in units}
    ctl = collections.defaultdict(list)
    for u in units:
        if u.get("control") and u.get("control_kind") in ("dup", "trap", "prose"):
            ctl[u["control_kind"]].append(u)
    res = {"primary_arm": primary, "unmeasured": collections.Counter()}
    gcov = lambda g: g is not None and isinstance(g.get("covered"), bool)
    dup = ctl.get("dup", [])
    cov, ovl, prim, arm = [0, 0], [0, 0], [0, 0], [0, 0]
    for u in dup:
        g, gs = gold.get(u["case_id"]), gold.get(u.get("source_case_id"))
        if gcov(g) and gcov(gs):
            cov[1] += 1
            cov[0] += int(g["covered"] == gs["covered"])
            ovl[1] += 1
            ovl[0] += int((not g["covered"] and not gs["covered"]) or (g["covered"] and gs["covered"] and bool(_acc(g) & _acc(gs))))
            if g["covered"] and gs["covered"]:
                prim[1] += 1
                prim[0] += int(g.get("primary") == gs.get("primary"))
        else:
            res["unmeasured"]["dup_gold"] += 1
        eq = _same_output(arm_rows.get(u["case_id"]), arm_rows.get(u.get("source_case_id")))
        if eq is None:
            res["unmeasured"]["dup_arm"] += 1
        else:
            arm[1] += 1
            arm[0] += int(eq)
    res["dup"] = {"n": len(dup), "gold_covered_agrees": _rate(*cov), "gold_acceptable_overlap": _rate(*ovl), "gold_primary_agrees": _rate(*prim), "arm_output_equals_source": _rate(*arm)}
    trap = ctl.get("trap", [])
    gn, an = [0, 0], [0, 0]
    nc = [0, 0, 0]  # R5-07 / R6-05 joint counts, only over traps with gold and an arm row: arm none AND gold not covered; gold not covered; arm none
    for u in trap:
        g, o = gold.get(u["case_id"]), arm_rows.get(u["case_id"])
        if gcov(g):
            gn[1] += 1
            gn[0] += int(g["covered"] is False)
        else:
            res["unmeasured"]["trap_gold"] += 1
        armed = bool(o and o.get("output") is not None)
        if armed:
            an[1] += 1
            an[0] += int(o["output"] == "none")
            nc[0] += int(o["output"] == "none" and gcov(g) and g["covered"] is False)
            if gcov(g):
                nc[1] += int(g["covered"] is False)
                nc[2] += int(o["output"] == "none")
        else:
            res["unmeasured"]["trap_arm"] += 1
    res["trap"] = {"n": len(trap), "gold_not_covered": _rate(*gn), "arm_none": _rate(*an), "none_correct": {"k": nc[0], "n": nc[1]}, "none_n": nc[2]}
    prose = ctl.get("prose", [])
    pe = [0, 0]
    for u in prose:
        eq = _same_output(arm_rows.get(u["case_id"]), arm_rows.get(u.get("source_case_id")))
        if eq is None or u.get("source_case_id") not in by:
            res["unmeasured"]["prose_arm"] += 1
        else:
            pe[1] += 1
            pe[0] += int(eq)
    res["prose"] = {"n": len(prose), "arm_output_equals_source": _rate(*pe)}
    res["unmeasured"] = dict(res["unmeasured"])
    return res


def contrast_section(r, data=None):
    """Contrast catalogue (R1-13, R2-06), reported every round, never part of the GO rule: units.contrast.jsonl (the round's window units restratified; in rounds 1 and 10 also the 20 D
    units, set "D"); arms contrast-B2 (every round) and contrast-V-sem (when run). A unit is SCORED only where contrast gold exists (D: round-0 gold.contrast.jsonl; window: the <= 30
    hash-sampled units labelled in rounds 1 and 10, the round's own gold.contrast.jsonl); every unit contributes to the assigned / none / unresolved rates. The window part and D are
    reported separately, never pooled (labels: `labels` = the window sample's, `labels_D` = round 0's)."""
    d = _rd(r, data)
    cu = os.path.join(d, "units.contrast.jsonl")
    if not os.path.exists(cu):
        return None
    units = [u for u in x2d.read_jsonl(cu) if not u.get("control")]
    gold, labels = {}, []
    for rr in sorted({0, r}):
        for name in ("gold.contrast.jsonl", "gold.contrast_window.jsonl"):
            gold.update({g["case_id"]: g for g in x2d.read_jsonl(os.path.join(_rd(rr, data), name))})
    gv, gv0 = x2d.guide_version(r, data), x2d.guide_version(0, data)
    for name in ("labels.contrast.jsonl", "labels.contrast_window.jsonl"):
        labels += x2d.read_jsonl(os.path.join(d, name))
    labels = current_labels(labels, gv)
    labels_d = current_labels(x2d.read_jsonl(os.path.join(_rd(0, data), "labels.contrast.jsonl")), gv0)
    arms = {a: rows for a in CONTRAST_ARMS for rows in [_arm_rows(d, a)] if rows is not None}
    cases = build_cases(units, gold, arms)
    sec = {"n_units": len(units), "n_scored": sum(1 for k in cases if k["gold"] and isinstance(k["gold"].get("covered"), bool)),
           "labels": label_agreement(labels, None, "wide", "contrast"), "labels_D": label_agreement(labels_d, {u["case_id"] for u in units if u.get("set") == "D"}, "wide", "contrast"), "parts": {}}
    for part, pred in (("window", lambda k: k["set"] != "D"), ("D", lambda k: k["set"] == "D")):
        sub = [k for k in cases if pred(k)]
        ids = {k["case_id"] for k in sub}
        sec["parts"][part] = {
            "n_units": len(sub), "n_scored": sum(1 for k in sub if k["gold"] and isinstance(k["gold"].get("covered"), bool)),
            "arms": {a: {"scored": rates(counts_for([k["out"][a] for k in sub if a in k["out"]])), "outputs": _output_rates([x for x in arms[a] if x["case_id"] in ids])} for a in arms},
            "paired": ({k: v for k, v in paired(sub, "contrast-V-sem", "contrast-B2").items() if k != "per_pr"} if all(a in arms for a in CONTRAST_ARMS) else None)}
    return sec


def round_results(r, data=None, B=10000):
    """Every section-9 figure of round r (pure over files; no writing). -> results dict. The headline tables, pooling rows, perception and futility use the PRIMARY arm
    (x2d.primary_arm) and the GO population (no core_gate units); V-sem and every other arm stay reported."""
    if r == 0:
        sys.exit("round 0 has gates, not a results table: use --gates")
    L = load_round(r, data)
    d = L["dir"]
    primary = x2d.primary_arm(data)
    su = scored_units(L["units"])
    go_su = [u for u in su if u.get("stratum_reason") != "core_gate"]
    cases = build_cases(su, L["gold"], L["arms"])
    gcases = _go_population(cases)
    audit_units = [u for u in L["units"] if u.get("set") == "audit"]
    prs = {k["pr_key"] for k in gcases}
    gv = x2d.guide_version(r, data)
    lcp = os.path.join(d, "labels-check.json")
    lc = _load_json(lcp) if os.path.exists(lcp) else None
    lc_ok = (lc["ok"] if lc and lc.get("guide_version") == gv else None)
    la_all = label_agreement(L["labels"], {u["case_id"] for u in su})  # every scored unit (the report)
    la = label_agreement(L["labels"], {u["case_id"] for u in go_su})  # the GO population: informative and futility use this (R2-14)
    info = informative({"n_clusters": len(go_su), "n_prs": len(prs), "kappa_covered": la["kappa_covered"], "kappa_answer": la["kappa_answer"],
                        "raw_agreement_covered": la["raw_agreement_covered"], "prevalence_covered": la["prevalence_covered"], "labels_check_ok": lc_ok})
    paused = os.path.exists(os.path.join(d, "integrity_pause.json"))
    res = {"round": r, "primary_arm": primary, "n_scored_clusters": len(go_su), "n_core_gate_excluded": len(su) - len(go_su), "n_prs": len(prs),
           "capped": len(x2d.read_jsonl(os.path.join(d, "capped.jsonl"))),
           "integrity_pause": paused, "informative": info,
           "unmeasured": {"units_without_gold": sum(1 for k in cases if not k["gold"] or not isinstance(k["gold"].get("covered"), bool)),
                          "arms_missing_file": [a for a in ARMS if a not in L["arms"]],
                          "arm_rows_missing": {a: sum(1 for k in cases if k["out"].get(a, {}).get("kind") == "missing") for a in L["arms"]}},
           "labels": la_all, "labels_go_population": la, "labels_check": lc, "arms": {}, "paired_vs_B2": {}, "strata": {}, "roles": {}, "confusion": {}, "pooling": [], "core_gate_units": {}}
    for arm in ARMS:
        if arm not in L["arms"]:
            continue
        res["arms"][arm] = _arm_summary(gcases, arm)
        res["strata"][arm] = _breakdown(cases, arm, lambda k: k["stratum4"])  # all scored units, core-gate ones included (the core stratum report)
        res["roles"][arm] = _breakdown(cases, arm, lambda k: k["role"])
        cg = [k for k in cases if k["unit"].get("stratum_reason") == "core_gate"]
        if cg:
            res["core_gate_units"][arm] = _arm_summary(cg, arm)
        if arm in ("V-sem", "V-sib", "fan-out", "B2"):
            res["confusion"][arm] = _confusion(cases, arm)
        if arm == "reference-claude":
            sub = [k for k in gcases if k["out"]["reference-claude"]["kind"] != "missing"]
            res["arms"][arm]["subset_n"] = len(sub)
            res["arms"][arm]["same_subset"] = {a: _arm_summary(sub, a)["real"]["accuracy"] for a in ("V-sem", "B2") if a in L["arms"]}
        if arm != "B2" and "B2" in L["arms"]:
            p = paired(gcases, arm, "B2")
            boot = pr_cluster_bootstrap([x for x in p["per_pr"] if x["n"]], gain_stat, B=B, levels=(0.8, 0.9, 0.95))
            res["paired_vs_B2"][arm] = {k: v for k, v in p.items() if k != "per_pr"}
            res["paired_vs_B2"][arm]["bootstrap"] = {"B": B, "ci95": boot["two_sided"][0.95], "lb90_one_sided": boot["lower"][0.9], "ub80_one_sided": boot["upper"][0.8]}
    if primary in L["arms"] and "B2" in L["arms"]:
        _require_complete(gcases, primary, "B2", "round %d" % r)
        res["pooling"] = pooling_rows(gcases, primary, "B2")
    if primary in L["arms"] and "B2" in L["arms"]:
        res["abstention"] = {"primary_arm": primary, "unresolved": unresolved_block([k["out"][primary] for k in gcases]), "abstained": abstention_table(res["pooling"]),
                             "n_res": sum(x["n_res"] for x in res["pooling"]), "gain_res_point": gain_res_stat(res["pooling"]), "gain_all_point": gain_all_stat(res["pooling"]),
                             "gain_guess_point": gain_guess_stat(res["pooling"]),
                             "cap": GO["unresolved_rate_max"], "note": "u and the all-cluster gain are reported every round (a drift towards the cap is visible to the critique); the verdict is evaluated once on the pooled frozen rounds"}
        res["abstention"]["gain_res_row"], res["abstention"]["gain_guess_row"] = _gain_rows(gcases, primary, "B2", B)
    res["perception"] = perception_counts(L["labels"], cases, "V-sem", primary)
    res["end_to_end"] = _end_to_end(gcases, L["lookups"], audit_units, L["gold"], tuple(dict.fromkeys([primary, "V-sem", "B2"])), _load_json(_need(os.path.join(d, "counts.json"), "counts.json")))
    if primary in L["arms"]:
        res["controls"] = controls_section(L["units"], L["gold"], {x["case_id"]: x for x in L["arms"][primary]}, primary)
    res["divergence"] = _divergence(r, data, [a for a in ("V-sem", "V-sib", "V-title", "fan-out", "B2") if a in L["arms"]])
    res["episodes"] = _episodes(gcases, primary) if primary in L["arms"] and "B2" in L["arms"] else {}
    res["gold_none_split"] = {"gap": sum(1 for k in gcases if k["gold"] and k["gold"].get("covered") is False and k["gold"].get("none_kind") == "gap"),
                              "new": sum(1 for k in gcases if k["gold"] and k["gold"].get("covered") is False and k["gold"].get("none_kind") == "new")}
    con = contrast_section(r, data)
    if con is not None:
        res["contrast"] = con
    prev = _prior_summaries(r, data)
    cur = {"round": r, "informative": info["informative"], "integrity_pause": paused, "pooling": res["pooling"]}
    res["futility_state"] = futility(prev + [cur], B=B)
    desc = [s for s in prev + [cur] if s["informative"] and not s["integrity_pause"]][-4:]
    res["descriptive_rule"] = {"rounds": [s["round"] for s in desc], "note": "point minimums met in >= 3 of the last 4 informative rounds (decides nothing)",
                               "points_met": [_points_met(s["pooling"]) for s in desc]}
    return res


def _points_met(pooling):
    a = sum(x["a"] for x in pooling)
    ac = sum(x["ac"] for x in pooling)
    gc = sum(x["gc_res"] for x in pooling)
    return bool(a and gc and ac / a >= x2d.MINIMUMS["P_assign"] and ac / gc >= x2d.MINIMUMS["R_assign"])


def _prior_summaries(r, data=None):
    out = []
    for q in range(1, r):
        p = os.path.join(_rd(q, data), "results.json")
        if os.path.exists(p):
            j = _load_json(p)
            out.append({"round": q, "informative": j["informative"]["informative"], "integrity_pause": j.get("integrity_pause", False), "pooling": j.get("pooling", [])})
    return out


# ------------------------------------------------------------------ report
def _fmt(x, p=3):
    if x is None:
        return "n/a"
    if isinstance(x, float):
        return "nan" if math.isnan(x) else ("%.*f" % (p, x))
    return str(x)


def _rate_cell(r):
    if r["n"] == 0:
        return "n/a (0)"
    lo, hi = r["wilson95"]
    return "%s (%d/%d) [%s, %s]" % (_fmt(r["value"]), r["k"], r["n"], _fmt(lo), _fmt(hi))


def render_report(res):
    L = ["# X2d round %d report" % res["round"], "",
         "Primary arm: %s. Scored clusters %d (GO population; %d core-gate units excluded and reported in the stratum tables) from %d PRs (capped, reported not scored: %d). Informative: **%s** (clusters ok %s, PRs ok %s, "
         "kappa joint answer %s; reported only: kappa covered %s, raw agreement covered %s, prevalence covered %s). Integrity pause: %s." % (
             res["primary_arm"], res["n_scored_clusters"], res["n_core_gate_excluded"], res["n_prs"], res["capped"], res["informative"]["informative"], res["informative"]["n_clusters_ok"], res["informative"]["n_prs_ok"],
             _fmt(res["informative"]["kappa_answer"]), _fmt(res["informative"]["kappa_covered"]), _fmt(res["informative"]["raw_agreement_covered"]), _fmt(res["informative"]["prevalence_covered"]), res["integrity_pause"]), "",
         "Unmeasured: %s" % json.dumps(res["unmeasured"], sort_keys=True), "",
         "## Arms (real clusters; Wilson 95 %)", "", "| arm | n | P_assign | R_assign (all covered; abstention = miss) | R_assign_res (resolved) | P_new | R_new | unresolved | accuracy |", "|---|---|---|---|---|---|---|---|---|"]
    for arm, s in res["arms"].items():
        r = s["real"]
        L.append("| %s | %d | %s | %s | %s | %s | %s | %s | %s |" % (arm, r["n"], _rate_cell(r["P_assign"]), _rate_cell(r["R_assign"]), _rate_cell(r["R_assign_res"]), _rate_cell(r["P_new"]), _rate_cell(r["R_new"]), _rate_cell(r["unresolved_rate"]), _rate_cell(r["accuracy"])))
    L += ["", "## Abstention (primary arm; cap u <= %.2f; B2's accuracy on abstained vs resolved clusters is a descriptive selection diagnostic and decides nothing)" % GO["unresolved_rate_max"], "", "```", json.dumps(res.get("abstention"), indent=1, sort_keys=True), "```"]
    L += ["", "gold none split (gap / new): %s" % json.dumps(res["gold_none_split"]), "", "## Paired gain vs B2, all clusters (abstention = miss; `gain_all`) (Jev-arm correct - B2 correct; exact McNemar; PR-cluster bootstrap)", "",
          "| arm | n | jev-only | B2-only | gain | McNemar p | 95 % CI | lb90 | PRs with jev-only |", "|---|---|---|---|---|---|---|---|---|"]
    for arm, p in res["paired_vs_B2"].items():
        b = p["bootstrap"]
        L.append("| %s | %d | %d | %d | %s | %s | [%s, %s] | %s | %d |" % (arm, p["n"], p["jev_only"], p["base_only"], _fmt(p["gain"]), _fmt(p["mcnemar_p"]), _fmt(b["ci95"][0]), _fmt(b["ci95"][1]), _fmt(b["lb90_one_sided"]), p["n_jev_only_prs"]))
    ab = res.get("abstention") or {}
    if ab.get("gain_res_row"):
        g = ab["gain_res_row"]
        L += ["", "## Primary arm: resolved-cluster gain `gain_res` (the GO rule's d / e) and `gain_guess` (the guard h; abstentions answered by Jev's own unthresholded guess)", "",
              "| statistic | n | jev-only | B2-only | point | McNemar p | 95 % CI | lb90 | ub80 | lb95 | PRs with jev-only |", "|---|---|---|---|---|---|---|---|---|---|---|"]
        for nm, row in (("gain_res", g), ("gain_guess", ab["gain_guess_row"])):
            ci = row["ci95"] or [None, None]
            L.append("| %s | %s | %s | %s | %s | %s | [%s, %s] | %s | %s | %s | %s |" % (nm, row["n"], row["jev_only"], row["base_only"], _fmt(row["point"]), _fmt(row["mcnemar_p"]), _fmt(ci[0]), _fmt(ci[1]), _fmt(row["lb90"]), _fmt(row["ub80"]), _fmt(row["lb95"]), row["n_jev_only_prs"]))
    for title, key in (("Strata (3 = code stratum 3 not gold-new; 4 = stratum 3 with gold none_kind new)", "strata"), ("Roles", "roles")):
        L += ["", "## " + title, "", "| arm | group | n | accuracy | P_assign |", "|---|---|---|---|---|"]
        for arm, g in res[key].items():
            for name, v in g.items():
                L.append("| %s | %s | %d | %s | %s |" % (arm, name, v["n"], _rate_cell(v["accuracy"]), _rate_cell(v["P_assign"])))
    la = res["labels"]
    L += ["", "## Labels", "", "A vs B before adjudication: kappa answer %s, kappa covered %s (raw agreement %s, prevalence covered %s), agree-rule rate %s, acceptable-set overlap %s, mean Jaccard %s; adjudication rate %s (final rows %d); "
          "top rule %s share %s (trigger > 0.40: %s); unparsable/missing %d." % (_fmt(la["kappa_answer"]), _fmt(la["kappa_covered"]), _fmt(la["raw_agreement_covered"]), _fmt(la["prevalence_covered"]), _fmt(la["agree_rule_rate"]), _fmt(la["acceptable_set_overlap_rate"]),
                                                                             _fmt(la["mean_jaccard"]), _fmt(la["adjudication_rate"]), la["n_final"], la["rule_top"], _fmt(la["rule_top_share"]), la["rule_concentration_trigger"], la["n_unparsable_or_missing"]),
          "", "## Confusion (errors only; pairs repeated >= 2 are candidate near-duplicate areas)", ""]
    for arm, c in res["confusion"].items():
        L.append("- %s: %d errors, %d repeated pairs covering %d errors; top: %s" % (arm, c["errors"], c["pairs_repeated"], c["errors_in_repeated_pairs"], json.dumps(c["pairs"][:5])))
    L += ["", "## Perception (jev-state labels)", "", json.dumps(res["perception"]), "", "## End to end (rule-resolved files counted; audit false-inherit)", "", "```", json.dumps(res["end_to_end"], indent=1, sort_keys=True), "```",
          "", "## Free-running divergence", "", json.dumps(res["divergence"]), "", "## Episode sensitivity", "", "```", json.dumps(res["episodes"], indent=1, sort_keys=True), "```", "",
          "## Futility state", "", json.dumps(res["futility_state"]), "", "## Descriptive rule", "", json.dumps(res["descriptive_rule"])]
    L += ["", "## Controls (report-only, R2-10)", "", "```", json.dumps(res.get("controls"), indent=1, sort_keys=True), "```"]
    if "contrast" in res:
        L += ["", "## Contrast catalogue (never part of the GO rule; D and the window sample never pooled)", "", "```", json.dumps(res["contrast"], indent=1, sort_keys=True), "```"]
    return "\n".join(L) + "\n"


# ------------------------------------------------------------------ critique pack
def _fmt_err(k, arm, cat, thr, jev_rows):
    u, g, o = k["unit"], k["gold"], k["out"][arm]
    title = lambda a: ("%s (%s)" % (cat.get(a, {}).get("title", "?"), a)) if a else "none"
    gl = ("none (%s)" % g.get("none_kind")) if not g["covered"] else "; ".join(title(a) for a in g.get("acceptable") or [])
    ans = o["output"] if o["output"] != "assigned" else "assigned " + title(o["area_id"])
    lines = ["#### %s | %s | PR %s \"%s\"" % (k["case_id"], arm, u.get("pr"), (u.get("pr_title") or "").replace("\n", " ")),
             "- paths: %s" % ", ".join((u.get("members") or [])[:10]), "- change_kind: %s | role %s | stratum %s (code) / %s (with gold) | set %s" % (u.get("change_kind"), k["role"], k["stratum"], k["stratum4"], k["set"]),
             "- gold: %s (primary %s; label agreement %s; control_kind %s)" % (gl, g.get("primary"), g.get("agreement"), g.get("control_kind")),
             "- answer: %s | outcome kind %s | confidence %s | reason %s" % (ans, o["kind"], (k.get("conf") or {}).get(arm), (k.get("reasons") or {}).get(arm)),
             "- thresholds in force: %s" % json.dumps(thr),
             "- Jev scores (this case): %s" % (json.dumps(jev_rows.get(k["case_id"], {}), sort_keys=True) or "none"),
             "- sibling owners shown to V-sib: %s" % json.dumps(u.get("siblings") or []),
             "- cause: ____   (menu: %s)" % " | ".join(CAUSES), ""]
    return lines


def _jev_index(d):
    """{case_id: {variant/question/perm: summary, 'failures': [...]}} from jev/scores.jsonl."""
    idx = collections.defaultdict(dict)
    for row in x2d.read_jsonl(os.path.join(d, "jev", "scores.jsonl")):
        if row.get("type") == "answer":
            key = "%s/%s/p%s%s" % (row.get("variant"), row.get("question", "").replace("merge.", ""), row.get("perm"), ("/" + row["area_id"]) if row.get("area_id") else "")
            idx[row["case_id"]][key] = {k: row.get(k) for k in ("p_true", "choice", "confidence") if row.get(k) is not None}
        elif row.get("type") == "failure":
            idx[row["case_id"]].setdefault("failures", []).append({k: row.get(k) for k in ("question", "variant", "perm", "kind", "status", "reason")})
    return idx


def build_pack(r, data=None):
    p = os.path.join(_rd(r, data), "results.json")
    res = _load_json(_need(p, "results.json (run --round first)"))
    L_ = load_round(r, data)
    d = L_["dir"]
    cases = build_cases(scored_units(L_["units"]), L_["gold"], L_["arms"])
    ver = _load_json(os.path.join(d, "version.json")) if os.path.exists(os.path.join(d, "version.json")) else None
    thr = _load_json(os.path.join(d, "thresholds.json")) if os.path.exists(os.path.join(d, "thresholds.json")) else None
    tdec = {q: v.get("threshold") for q, v in (thr or {}).get("decisions", {}).items()}
    cat = {}
    if os.path.exists(x2d.catalogue_path(data)):
        cat = {a["area_id"]: a for a in x2d.read_catalogue(data)["areas"]}
    jev_rows = _jev_index(d)
    for k in cases:
        k["conf"] = {a: (L_["arms"][a] and {x["case_id"]: x for x in L_["arms"][a]}.get(k["case_id"], {}).get("confidence")) for a in L_["arms"]}
        k["reasons"] = {a: {x["case_id"]: x for x in L_["arms"][a]}.get(k["case_id"], {}).get("unresolved_reason") for a in L_["arms"]}
    out = ["# Pilot round %d critique pack (facts for the council; no cause is proposed here)" % r, "",
           "## Version and digests", "", "```", json.dumps(ver, indent=1, sort_keys=True) if ver else "version.json absent: unmeasured", "```",
           "thresholds: %s" % json.dumps(thr, sort_keys=True) if thr else "thresholds.json absent: unmeasured", "",
           "## Population and counts", "", "- scored clusters %d, PRs %d, capped (reported) %d, informative %s, integrity pause %s" % (res["n_scored_clusters"], res["n_prs"], res["capped"], res["informative"]["informative"], res["integrity_pause"]),
           "- strata (V-sem): %s" % json.dumps({g: v["n"] for g, v in res["strata"].get("V-sem", {}).items()}), "- roles (V-sem): %s" % json.dumps({g: v["n"] for g, v in res["roles"].get("V-sem", {}).items()}),
           "- labels: kappa answer %s, kappa covered %s, adjudication rate %s, top rule %s share %s" % (_fmt(res["labels"]["kappa_answer"]), _fmt(res["labels"]["kappa_covered"]), _fmt(res["labels"]["adjudication_rate"]), res["labels"]["rule_top"], _fmt(res["labels"]["rule_top_share"])),
           "- unmeasured: %s" % json.dumps(res["unmeasured"], sort_keys=True), "", "## Metrics with intervals (real clusters, Wilson 95 %)", "", "| arm | P_assign | R_assign (abstention = miss) | R_assign_res | P_new | R_new | unresolved | accuracy |", "|---|---|---|---|---|---|---|---|"]
    for arm, s in res["arms"].items():
        r_ = s["real"]
        out.append("| %s | %s | %s | %s | %s | %s | %s | %s |" % (arm, _rate_cell(r_["P_assign"]), _rate_cell(r_["R_assign"]), _rate_cell(r_["R_assign_res"]), _rate_cell(r_["P_new"]), _rate_cell(r_["R_new"]), _rate_cell(r_["unresolved_rate"]), _rate_cell(r_["accuracy"])))
    out += ["", "## Abstention (primary arm; cap u <= %.2f)" % GO["unresolved_rate_max"], "", json.dumps(res.get("abstention"), sort_keys=True)]
    out += ["", "## Paired delta vs B2, all clusters (abstention = miss; `gain_all`); the primary arm's `gain_res` and `gain_guess` rows are in the abstention JSON above", ""]
    for arm, p_ in res["paired_vs_B2"].items():
        out.append("- %s: n %d, jev-only %d, B2-only %d, gain %s, McNemar p %s, bootstrap 95 %% CI %s, one-sided 90 %% lower bound %s" % (arm, p_["n"], p_["jev_only"], p_["base_only"], _fmt(p_["gain"]), _fmt(p_["mcnemar_p"]), json.dumps(p_["bootstrap"]["ci95"]), _fmt(p_["bootstrap"]["lb90_one_sided"])))
    out += ["", "## Controls (test-retest of the labels and of the arm; report-only)", "", "```", json.dumps(res.get("controls"), indent=1, sort_keys=True), "```", "",
            "## Perception (V-sem unthresholded vs labellers; thresholded agreement, unresolved rate and primary arm beside it)", "", json.dumps(res.get("perception")), "",
            "## Errors (every error of V-sem, fan-out and, for comparison, B2; each listed once, exactly one cause per error)", ""]
    for arm in ("V-sem", "fan-out", "B2"):
        errs = [k for k in cases if arm in k["out"] and k["out"][arm]["correct"] is False and k["gold"]]
        out += ["### Arm %s: %d errors of %d measured" % (arm, len(errs), sum(1 for k in cases if arm in k["out"] and k["out"][arm]["correct"] is not None)), ""]
        for k in errs:
            out += _fmt_err(k, arm, cat, tdec, jev_rows)
    out += ["## Proposed changes (version keys and prior-data justification; filled in by the council, none proposed here)", "", "- ____", "",
            "## Attestation", "", "I attest that no data of round %d or later was used to prepare these proposals: ____ (name, date)" % (r + 1), "",
            "## Verdict (zero issues to proceed): ____", ""]
    return "\n".join(out)


# ------------------------------------------------------------------ gates (round 0)
def _key_gold(k):
    g = k.get("gold") or {}
    acc = g.get("acceptable")
    if acc is None:
        a = g.get("area_id") or g.get("primary")
        acc = [a] if a else []
    cov = g.get("covered")
    if cov is None:
        cov = bool(acc)
    return {"case_id": k["case_id"], "covered": bool(cov), "acceptable": list(acc), "primary": g.get("primary") or (acc[0] if acc else None), "none_kind": g.get("none_kind")}


def _share(v):
    if isinstance(v, dict):
        if "share" in v:
            return v["share"], v.get("n")
        if "k" in v and v.get("n"):
            return v["k"] / v["n"], v["n"]
        return None, None
    return v, None


def _gate(status, value, bound, n, **extra):
    return dict({"status": "PASS" if status else "FAIL", "value": value, "bound": bound, "n": n}, **extra)


def compute_core_gate(data=None, repo=None):
    """Core gate (PREREGISTRATION 4.3), decided ONCE here on the frozen round-0 catalogue membership and the C0 tree: core `fastapi/` source files < 50 % area-owned ->
    core units are stratum 3 with stratum_reason "core_gate" and are excluded from the GO population. -> {"enabled": bool, "core_share_owned": float|None}."""
    import frontend as F
    import gold as GM
    if not repo:
        sys.exit("error: the core gate needs the repository clone (--repo CLONE or X2D_REPO)")
    try:
        ccg = F.compute_core_gate
    except AttributeError:
        import build_units as BU
        ccg = BU.compute_core_gate
    own = F.Ownership(GM.catalogue_membership(x2d.read_catalogue(data), 0))
    tree = F.tree_paths(repo, x2d.C0)
    core = [p for p in tree if p.startswith("fastapi/") and p.endswith(".py")]
    share = (sum(1 for p in core if own.owners(p)) / len(core)) if core else None
    return {"enabled": bool(ccg(own, tree)), "core_share_owned": share}


def gates(data=None, repo=None):
    d = _rd(0, data)
    for p, w in (("units.jsonl", "round-0 units"), ("control-key.jsonl", "control key"), ("gold.jsonl", "round-0 gold"), ("labels.jsonl", "round-0 labels"), ("k23.json", "K2/K3 shares"),
                 ("jev/scores.jsonl", "main Jev run"), ("jev-n1/scores.jsonl", "N1 Jev run"), ("catalogue-n1.json", "N1 catalogue"),
                 ("arms/V-sem.jsonl", "main V-sem arm"), ("arms/B1.jsonl", "main B1 arm"), ("arms/B2.jsonl", "main B2 arm"), ("arms-n1/V-sem.jsonl", "N1 V-sem arm"), ("arms-n1/B2.jsonl", "N1 B2 arm"),
                 ("arms-p1/V-sem.jsonl", "P1 V-sem arm"), ("jev-p/scores.p1.jsonl", "P1 Jev run"), ("arms-p2/V-sem.jsonl", "P2 (planted-gap) V-sem arm"), ("jev-p/scores.p2.jsonl", "P2 Jev run")):
        _need(os.path.join(d, p), w)
    units = x2d.read_jsonl(os.path.join(d, "units.jsonl"))
    gold = {g["case_id"]: g for g in x2d.read_jsonl(os.path.join(d, "gold.jsonl"))}
    labels = current_labels(x2d.read_jsonl(os.path.join(d, "labels.jsonl")), x2d.guide_version(0, data))
    key = {k["case_id"]: k for k in x2d.read_jsonl(os.path.join(d, "control-key.jsonl"))}
    k23 = _load_json(os.path.join(d, "k23.json"))
    arm = lambda sub, name: {r["case_id"]: r for r in x2d.read_jsonl(os.path.join(d, sub, name + ".jsonl"))}
    G = collections.OrderedDict()
    real = [u for u in units if not u.get("control") and u.get("set") in ("D", "P")]
    D_ids = {u["case_id"] for u in real if u.get("set") == "D"}
    DP_ids = {u["case_id"] for u in real}
    # K2 / K3
    for name, lim, strict in (("K2", 0.90, False), ("K3", 0.40, True)):
        v, n = _share(k23.get(name.lower(), k23.get(name.lower() + "_share")))
        if v is None:
            sys.exit("k23.json has no usable %s" % name.lower())
        G[name] = _gate((v < lim) if not strict else (v <= lim), v, lim, n, note="stops when %s %s" % ("share >=" if not strict else "share >", lim))
    # K4
    la = label_agreement(labels, D_ids)
    k4ok = la["n_both_parsed"] >= 20 and la["kappa_answer"] is not None and la["kappa_answer"] >= x2d.MIN_KAPPA  # joint answer only (R1-06)
    G["K4"] = _gate(k4ok, {"kappa_answer": la["kappa_answer"], "kappa_covered": la["kappa_covered"], "raw_agreement_covered": la["raw_agreement_covered"],
                           "prevalence_covered": la["prevalence_covered"], "raw_agreement_answer": la["raw_agreement_answer"]}, x2d.MIN_KAPPA, la["n_both_parsed"],
                    note="needs >= 20 D units with both labeller sessions parsed; decides on kappa of the joint answer only (kappa_covered reported with raw agreement and prevalence)")
    # K5
    sc = collections.defaultdict(dict)
    for row in x2d.read_jsonl(os.path.join(d, "jev", "scores.jsonl")):
        if row.get("type") == "answer" and row.get("question") == "merge.which_area" and row.get("variant") == "V-sem" and row["case_id"] in DP_ids:
            sc[row["case_id"]][row.get("perm")] = row.get("choice")
    both = [c for c, v in sc.items() if 1 in v and 2 in v]
    flips = sum(1 for c in both if sc[c][1] != sc[c][2])
    rate = (flips / len(both)) if both else None
    if len(both) < 40:
        G["K5"] = _gate(False, rate, 0.15, len(both), note="needs >= 40 D+P units asked under two permutations")
        fan_primary = False
    else:
        fan_primary = rate > 0.15
        G["K5"] = _gate(True, rate, 0.15, len(both), note=("flip rate > 0.15: the Noul fan-out (merge.area_covers) becomes the primary arm" if fan_primary else "flip rate within 0.15"))
    # K6 / N1
    cat1 = _load_json(os.path.join(d, "catalogue-n1.json"))
    K = len(cat1["areas"] if isinstance(cat1, dict) else cat1)
    a_n1, b_n1, b_main = arm("arms-n1", "V-sem"), arm("arms-n1", "B2"), arm("arms", "B2")
    DPcases = [{"case_id": c, "gold": gold.get(c)} for c in sorted(DP_ids)]  # N1 runs on D u P (R1-18)
    jn = counts_for([score_outcome(a_n1.get(k["case_id"]), k["gold"]) for k in DPcases])
    corr = jn.get("assigned_correct", 0)
    nden = jn.get("gold_covered", 0)
    bound = max(2.0 / K, N1_FLOOR) if K else N1_FLOOR
    ub = wilson_bounds_one_sided(corr, nden, 0.90)[1] if nden else None
    n1_missing = sorted(c for c in DP_ids if c not in a_n1)
    b_same = all((b_n1.get(c) or {}).get("area_id") == (b_main.get(c) or {}).get("area_id") and (b_n1.get(c) or {}).get("output") == (b_main.get(c) or {}).get("output") for c in DP_ids)
    G["K6_N1"] = _gate(ub is not None and ub <= bound and b_same and not n1_missing, {"jev_accuracy": (corr / nden) if nden else None, "wilson_ub90": ub, "K": K, "B2_unaffected": b_same,
                                                                                      "units_without_arm_row": len(n1_missing)}, bound, nden,
                       note="N1: the Choice follows option text, not position or id (P1 is the path-cue test). Jev accuracy on the shuffled catalogue over gold-covered D+P units (manifest n1.units); one-sided 90 % Wilson upper bound <= max(2/K, 0.20)")
    # P1
    p1 = sorted(c for c, k in key.items() if k.get("kind") == "p1")
    if len(p1) < 20:
        G["P1"] = _gate(False, None, "Jev >= 0.60, B1 and B2 <= 0.20", len(p1), note="needs >= 20 P1 controls in control-key.jsonl")
    else:
        rec = {}
        for name in ("V-sem", "B1", "B2"):
            rows = arm("arms-p1" if name == "V-sem" else "arms", name)  # baselines are computed once per unit against its own catalogue (control-key.catalogue) into arms/
            rec[name] = sum(1 for c in p1 if (rows.get(c) or {}).get("output") == "assigned" and rows[c].get("area_id") in set(_key_gold(key[c])["acceptable"])) / len(p1)
        resolved_by_rules = [c for c in p1 if c not in {u["case_id"] for u in units}]
        G["P1"] = _gate(rec["V-sem"] >= 0.60 and rec["B1"] <= 0.20 and rec["B2"] <= 0.20 and not resolved_by_rules, rec, {"jev_min": 0.60, "baseline_max": 0.20}, len(p1),
                        rules_resolved_units=len(resolved_by_rules), note="a P1 control missing from units.jsonl (rules (a)-(f) resolved it) fails the gate")
    # planted_new: planted new-directory items -> none (gold none_kind new)
    vs_main = arm("arms", "V-sem")
    ids = sorted(c for c, k in key.items() if k.get("kind") == "planted_new")
    outs = [score_outcome(vs_main.get(c), _key_gold(key[c])) for c in ids]
    ok = sum(1 for o in outs if o["correct"])
    kind_ok = sum(1 for o in outs if o["kind"] == "none_correct_new")
    rate_new = (ok / len(ids)) if ids else None
    G["planted_new"] = _gate(bool(ids) and rate_new >= x2d.NEW_AREA_GATE_MIN and kind_ok == ok, {"none_rate": rate_new, "reported_as": "none_correct_new", "reported_as_n": kind_ok},
                             x2d.NEW_AREA_GATE_MIN, len(ids), note="planted item -> none (gold none_kind new from the control key); bound = the round-0 new-area gate minimum (x2d.NEW_AREA_GATE_MIN; a round-0 gate, not a verdict condition)")
    # P2 = the planted-gap gate (R1-07, R2-05): gold = gold.derive_ablation(source gold, the SAME ablated set the catalogue removes) -> none / gap for the `gap positives`; V-sem must answer none on
    # them. A clone whose derived gold stays covered (a retained acceptable area) is no gap positive: reported separately, its correct answer is that area. The gate needs >= P2_MIN_GAP_POSITIVES.
    import build_controls as BC
    p2 = sorted(c for c, k in key.items() if k.get("kind") == "p2")
    vs_p2 = arm("arms-p2", "V-sem")
    gap_pos = [c for c in p2 if (gold.get(c) or {}).get("covered") is False]
    retained = [c for c in p2 if isinstance((gold.get(c) or {}).get("covered"), bool) and gold[c]["covered"]]
    outs = [score_outcome(vs_p2.get(c), gold.get(c)) for c in gap_pos]  # the derived gold (gold.jsonl, derive_ablation) -- the control key carries none
    none_n = sum(1 for o in outs if o["output"] == "none")
    gap_ok = sum(1 for o in outs if o["kind"] == "none_correct_gap")
    rate_gap = (none_n / len(gap_pos)) if gap_pos else None
    ret_ok = sum(1 for o in (score_outcome(vs_p2.get(c), gold.get(c)) for c in retained) if o["correct"])
    p2_val = {"none_rate": rate_gap, "none_correct_gap": gap_ok, "none_correct_gap_rate": (gap_ok / len(gap_pos)) if gap_pos else None, "gap_positives": len(gap_pos),
              "retained_covered": {"n": len(retained), "assigned_correct": ret_ok}, "p2_clones": len(p2), "no_agreed_gold": sum(1 for c in p2 if not isinstance((gold.get(c) or {}).get("covered"), bool))}
    p2_note = ("planted-gap gate: V-sem none rate over the gap positives (P2 clones whose derived gold is none / gap; >= %d) >= the new-area gate minimum; every none must score as none_correct_gap; "
               "clones whose derived gold stays covered are reported in retained_covered (correct = the retained acceptable area), not in the denominator; below %d label-derived gap positives the gate is "
               "FAIL when at least 2 gap positives were not answered none (no denominator of %d could reach the minimum) or a none is not none_correct_gap, else UNMEASURED (counts reported either way; the round-0 "
               "critique rules on it, R3-04, R4-04); when UNMEASURED RESULTS states \"R_new on catalogue gaps has no planted test\"" % (BC.P2_MIN_GAP_POSITIVES, BC.P2_MIN_GAP_POSITIVES, BC.P2_MIN_GAP_POSITIVES))
    if len(gap_pos) < BC.P2_MIN_GAP_POSITIVES:
        if len(gap_pos) - none_n >= 2 or gap_ok != none_n:  # R4-04: a decisive failure is never hidden by a thin denominator
            G["P2"] = _gate(False, p2_val, x2d.NEW_AREA_GATE_MIN, len(gap_pos), note=p2_note)
        else:
            G["P2"] = {"status": "UNMEASURED", "value": p2_val, "bound": x2d.NEW_AREA_GATE_MIN, "n": len(gap_pos), "note": p2_note}
    else:
        G["P2"] = _gate(rate_gap >= x2d.NEW_AREA_GATE_MIN and gap_ok == none_n, p2_val, x2d.NEW_AREA_GATE_MIN, len(gap_pos), note=p2_note)
    # real-none proxy (R2-11 (1)): frozen-round stratum-3 (K3) + new_dir clusters, no core_gate: INFO only (R4-13: ~10 real positives cannot veto; it limits the descriptive new-area table), never a stop
    if "real_none_proxy" not in k23 or "frozen_pr_counts" not in k23:
        sys.exit("k23.json has no real_none_proxy / frozen_pr_counts: rebuild it with build_units.py --gates")
    G["real_none_proxy"] = {"status": "INFO", "value": k23["real_none_proxy"], "bound": None, "n": sum(sum(v.values()) for v in k23["frozen_pr_counts"].values()),
                            "note": "frozen-round stratum-3 clusters whose parent directory exists in the C0 tree (K3) plus new_dir clusters, excluding core_gate: an indicator of the real gold-none positives "
                                    "available in rounds 8-10 (stratum 3 != gold none); descriptive only: P_new / R_new are reported with intervals and never veto the verdict (R4-13)"}
    out = {"gates": G, "fan_out_primary": fan_primary, "core_gate": compute_core_gate(data, repo or os.environ.get("X2D_REPO")),
           "pr_note": "K5 FAIL is recorded as PASS with fan_out_primary (documented switch)"}
    return out


# ------------------------------------------------------------------ the re-gate (rev 13, R-053): P1 and P2 rebuilt on fresh items, gate set x2d.REGATE_GATE_SET
def _p1_stats(rows):
    """Descriptive and decisive figures of V-sem over eligible P1 items (gold = the labellers' agreed acceptable set, which contains A): accuracy on the ANSWERED items (resolved = composed output assigned or none;
    correct = assigned to an acceptable area, a `none` is a resolved wrong answer), the abstention rate over all the items, the all-item recovery (abstention = miss, the rev-12 figure), and the path-following
    rates: answers naming a sibling area of the target directory that is not acceptable (composed answers over the answered items; the unthresholded permutation-1 choice over all the items)."""
    n = len(rows)
    res = [r for r in rows if r["V-sem"]["output"] in RESOLVED]
    corr = [r for r in res if r["V-sem"]["output"] == "assigned" and r["V-sem"]["area_id"] in r["acceptable"]]
    pf = lambda r, a: a is not None and a in (set(r["sibling_area_ids"]) - set(r["acceptable"]))
    return {"n": n, "resolved": len(res), "abstained": n - len(res), "accuracy_answered": _rate(len(corr), len(res)), "abstention_rate": _rate(n - len(res), n), "recovery_all_items": _rate(len(corr), n),
            "recovery_all_items_A_only": _rate(sum(1 for r in rows if r["V-sem"]["output"] == "assigned" and r["V-sem"]["area_id"] == r["area_id"]), n),
            "sibling_in_acceptable": sum(1 for r in rows if set(r["sibling_area_ids"]) & set(r["acceptable"])),
            "none_answers": sum(1 for r in res if r["V-sem"]["output"] == "none"),
            "path_following_answered": _rate(sum(1 for r in res if r["V-sem"]["output"] == "assigned" and pf(r, r["V-sem"]["area_id"])), len(res)),
            "path_following_raw_choice": _rate(sum(1 for r in rows if pf(r, r["V-sem"].get("raw_choice"))), n)}


def _recovery(rows, name):
    return _rate(sum(1 for r in rows if r[name]["output"] == "assigned" and r[name]["area_id"] in r["acceptable"]), len(rows))


def _regate_denominators(d):
    """The label-only and construction part of the re-gate gates (R17-02): everything that needs no Jev answer. Reads ONLY units.jsonl, control-key.jsonl, gold.jsonl, arms/B1.jsonl and arms/B2.jsonl of the
    re-gate root `d` (never arms-p1/, arms-p2/ or jev-p/). -> dict(units, key, gold, b1, b2, p1, p2, items [P1 items without a V-sem field], excluded, base, resolved_by_rules, twins, kept, rows2 [P2 rows without
    a V-sem answer])."""
    for p, w in (("units.jsonl", "re-gate units"), ("control-key.jsonl", "re-gate control key"), ("gold.jsonl", "re-gate gold"), ("labels.jsonl", "re-gate labels"), ("arms/B1.jsonl", "B1 arm"), ("arms/B2.jsonl", "B2 arm")):
        _need(os.path.join(d, p), w)
    units = {u["case_id"]: u for u in x2d.read_jsonl(os.path.join(d, "units.jsonl"))}
    key = {k["case_id"]: k for k in x2d.read_jsonl(os.path.join(d, "control-key.jsonl"))}
    gold = {g["case_id"]: g for g in x2d.read_jsonl(os.path.join(d, "gold.jsonl"))}
    arm = lambda sub, name: {r["case_id"]: r for r in x2d.read_jsonl(os.path.join(d, sub, name + ".jsonl"))}
    b1, b2 = arm("arms", "B1"), arm("arms", "B2")
    p1 = sorted(c for c, k in key.items() if k.get("kind") == "p1")
    p2 = sorted(c for c, k in key.items() if k.get("kind") == "p2")
    miss = [c for c in p1 if c not in b1 or c not in b2]
    if miss:
        sys.exit("refusing: %d re-gate control(s) lack a B1 / B2 arm row: %s" % (len(miss), ", ".join(miss[:10])))
    items, excluded = [], {"no_agreed_gold": [], "gold_not_covered": [], "gold_excludes_A": []}
    for c in p1:
        k, g = key[c], gold.get(c)
        A = k["area_id"]
        acc = sorted((g or {}).get("acceptable") or [])
        why = None
        if g is None or not isinstance(g.get("covered"), bool):
            why = "no_agreed_gold"
        elif not g["covered"]:
            why = "gold_not_covered"
        elif A not in acc:
            why = "gold_excludes_A"
        if why:
            excluded[why].append(c)
        items.append({"case_id": c, "stratum": (units.get(c) or {}).get("stratum"), "area_id": A, "acceptable": acc, "gold_agreement": (g or {}).get("agreement"), "excluded": why,
                      "sibling_area_ids": sorted(k.get("sibling_area_ids") or []), "target_dir": k.get("target_dir"), "origin": k.get("origin"), "date": (units.get(c) or {}).get("date"),
                      "B1": {"output": b1[c].get("output"), "area_id": b1[c].get("area_id")}, "B2": {"output": b2[c].get("output"), "area_id": b2[c].get("area_id")}})
    elig = [r for r in items if not r["excluded"]]
    s1 = [r for r in elig if r["stratum"] == 1]
    base = {name: {"all_eligible": _recovery(elig, name), "stratum1": _recovery(s1, name)} for name in ("B1", "B2")}
    resolved_by_rules = [c for c in p1 if c not in units]
    kept, twins, seen = [], [], set()
    for c in p2:
        mem = set(units[c]["members"])
        if mem & seen:  # a path twin of an earlier clone is one item counted twice (R10-02): excluded
            twins.append(c)
            continue
        seen |= mem
        kept.append(c)
    rows2 = []
    for c in kept:
        g, k = gold.get(c), key[c]
        direct = None if g is None or not isinstance(g.get("covered"), bool) else ("covered" if g["covered"] else "none")
        dg = (g or {}).get("derived_gold")
        rows2.append({"case_id": c, "direct_label": direct, "none_kind": (g or {}).get("none_kind"), "direct_acceptable": sorted((g or {}).get("acceptable") or []), "gold_agreement": (g or {}).get("agreement"),
                      "derived_gold": None if not dg else ("covered" if dg.get("covered") else "none"), "derived_acceptable": sorted((dg or {}).get("acceptable") or []), "origin": k.get("origin"),
                      "date": (units.get(c) or {}).get("date"), "ablated_area_id": k.get("ablated_area_id")})
    return {"units": units, "key": key, "gold": gold, "b1": b1, "b2": b2, "p1": p1, "p2": p2, "items": items, "excluded": excluded, "base": base, "resolved_by_rules": resolved_by_rules,
            "twins": twins, "kept": kept, "rows2": rows2, "elig": elig, "s1": s1}


def regate_label_problems(data=None):
    """R17-02: the conjuncts of the re-gate gates that need no Jev answer, checked after gold and BEFORE any re-gate Jev request. -> (problems, counts). Problems: eligible stratum-1 P1 items below
    x2d.REGATE_P1_STRATUM1_MIN; direct-label P2 positives (after twin exclusion) below x2d.REGATE_P2_POSITIVES_MIN; rules-resolved P1 units; B1 or B2 over the eligible items above x2d.REGATE_BASELINE_MAX.
    Never opens arms-p1/, arms-p2/ or jev-p/."""
    data = data or x2d.DATA
    if not x2d.is_regate_root(data):
        sys.exit("refusing: the re-gate precheck runs inside D/rounds/0/regate (X2D_DATA)")
    den = _regate_denominators(_rd(0, data))
    pos = [r for r in den["rows2"] if r["direct_label"] == "none"]
    counts = {"items_built": len(den["p1"]), "items_eligible": len(den["elig"]), "stratum1_eligible": len(den["s1"]), "direct_label_positives": len(pos), "twins_excluded": len(den["twins"]),
              "rules_resolved_units": len(den["resolved_by_rules"]), "baseline_all_eligible": {n_: den["base"][n_]["all_eligible"] for n_ in ("B1", "B2")}}
    problems = []
    if len(den["s1"]) < x2d.REGATE_P1_STRATUM1_MIN:
        problems.append("eligible stratum-1 P1 items %d < %d" % (len(den["s1"]), x2d.REGATE_P1_STRATUM1_MIN))
    if len(pos) < x2d.REGATE_P2_POSITIVES_MIN:
        problems.append("direct-label P2 positives %d < %d" % (len(pos), x2d.REGATE_P2_POSITIVES_MIN))
    if den["resolved_by_rules"]:
        problems.append("rules_resolved_units %d > 0" % len(den["resolved_by_rules"]))
    for n_ in ("B1", "B2"):
        v = den["base"][n_]["all_eligible"]["value"]
        if v is None or v > x2d.REGATE_BASELINE_MAX:
            problems.append("%s recovery over the eligible items %s > %s" % (n_, v, x2d.REGATE_BASELINE_MAX))
    return problems, counts


def regate_gates(data=None):
    """The one permitted re-gate of P1 and P2 on fresh items (PREREGISTRATION 5, rev 13), computed in D/rounds/0/regate (X2D_DATA). Gold, bars and exclusions are fixed by the preregistration:
    P1 gold = the labellers' agreed acceptable set of the item on catalogue-p1, which must contain A (an item whose agreed gold is absent, not covered or does not contain A is reported and excluded);
    STRATUM 1 decides: PASS iff V-sem's accuracy on the answered stratum-1 items >= x2d.REGATE_P1_ACC_MIN, its abstention rate on them <= x2d.REGATE_P1_ABSTAIN_MAX and B1 and B2 each recover an acceptable area
    on <= x2d.REGATE_BASELINE_MAX of all eligible items; UNMEASURED below x2d.REGATE_P1_STRATUM1_MIN eligible stratum-1 items. Strata 2-3 and the path-following rates are reported.
    P2 gold = the clone's OWN agreed label on catalogue-p2 (gap and new alike; the gold the rev-12 derivation would have given is reported beside it, never scored; a path twin of an earlier clone is excluded):
    over the direct-label positives (own label not covered) PASS iff V-sem's `none` rate on the answered positives >= x2d.REGATE_P2_NONE_MIN and its abstention rate <= x2d.REGATE_P2_ABSTAIN_MAX;
    UNMEASURED below x2d.REGATE_P2_POSITIVES_MIN positives. Pure; the CLI writes D/rounds/0/regate/gates.json once."""
    data = data or x2d.DATA
    if not x2d.is_regate_root(data):
        sys.exit("refusing: the re-gate is analysed inside D/rounds/0/regate (X2D_DATA)")
    frozen_bad = x2d.regate_frozen_problems(data)
    if frozen_bad:
        sys.exit("refusing: the frozen Jev side of the re-gate differs from round 0: " + "; ".join(frozen_bad))
    d = _rd(0, data)
    den = _regate_denominators(d)
    for p, w in (("arms-p1/V-sem.jsonl", "P1 V-sem arm"), ("jev-p/scores.p1.jsonl", "P1 Jev run"), ("arms-p2/V-sem.jsonl", "P2 V-sem arm"), ("jev-p/scores.p2.jsonl", "P2 Jev run")):
        _need(os.path.join(d, p), w)
    key, gold, units = den["key"], den["gold"], den["units"]
    arm = lambda sub, name: {r["case_id"]: r for r in x2d.read_jsonl(os.path.join(d, sub, name + ".jsonl"))}
    vs1, vs2 = arm("arms-p1", "V-sem"), arm("arms-p2", "V-sem")
    p1, p2 = den["p1"], den["p2"]
    miss = [c for c in p1 if c not in vs1] + [c for c in p2 if c not in vs2]
    if miss:
        sys.exit("refusing: %d re-gate control(s) lack a V-sem arm row: %s" % (len(miss), ", ".join(miss[:10])))
    G_ = collections.OrderedDict()
    # ---- P1
    excluded = den["excluded"]
    items = []
    for it in den["items"]:
        v = vs1[it["case_id"]]
        items.append(dict(it, **{"V-sem": {"output": v.get("output"), "area_id": v.get("area_id"), "confidence": v.get("confidence"), "reason": v.get("unresolved_reason"), "raw_choice": v.get("raw_choice"),
                                          "raw_covered": v.get("raw_covered")}}))
    elig = [r for r in items if not r["excluded"]]
    s1 = [r for r in elig if r["stratum"] == 1]
    base = den["base"]
    st1 = _p1_stats(s1)
    by_stratum = {str(s): _p1_stats([r for r in elig if r["stratum"] == s]) for s in (1, 2, 3)}
    resolved_by_rules = den["resolved_by_rules"]
    acc_v, abst_v = st1["accuracy_answered"]["value"], st1["abstention_rate"]["value"]
    base_ok = all(base[n_]["all_eligible"]["value"] is not None and base[n_]["all_eligible"]["value"] <= x2d.REGATE_BASELINE_MAX for n_ in ("B1", "B2"))
    ok1 = acc_v is not None and acc_v >= x2d.REGATE_P1_ACC_MIN and abst_v is not None and abst_v <= x2d.REGATE_P1_ABSTAIN_MAX and base_ok and not resolved_by_rules
    v1 = {"stratum1": st1, "strata": by_stratum, "all_eligible": _p1_stats(elig), "baselines": base, "items_built": len(p1), "items_eligible": len(elig), "stratum1_eligible": len(s1),
          "excluded": excluded, "rules_resolved_units": len(resolved_by_rules)}
    bound1 = {"accuracy_answered_stratum1_min": x2d.REGATE_P1_ACC_MIN, "abstention_stratum1_max": x2d.REGATE_P1_ABSTAIN_MAX, "baseline_max": x2d.REGATE_BASELINE_MAX}
    note1 = ("P1 re-gate: gold = the labellers' agreed acceptable set (contains A); stratum 1 decides: V-sem accuracy on the answered items >= the R_assign minimum, abstention <= the GO cap g, B1 and B2 <= the rev-12 baseline bar "
             "over all eligible items; strata 2-3 and the path-following rates are reported; UNMEASURED below %d eligible stratum-1 items" % x2d.REGATE_P1_STRATUM1_MIN)
    if len(s1) < x2d.REGATE_P1_STRATUM1_MIN:
        G_["P1"] = {"status": "UNMEASURED", "value": v1, "bound": bound1, "n": len(s1), "note": note1}
    else:
        G_["P1"] = _gate(ok1, v1, bound1, len(s1), note=note1)
    # ---- P2
    twins = den["twins"]
    rows2 = []
    for r0 in den["rows2"]:
        v = vs2[r0["case_id"]]
        rows2.append(dict(r0, output=v.get("output"), area_id=v.get("area_id"), confidence=v.get("confidence"), reason=v.get("unresolved_reason"), raw_choice=v.get("raw_choice"), raw_covered=v.get("raw_covered")))
    pos = [r for r in rows2 if r["direct_label"] == "none"]
    ret = [r for r in rows2 if r["direct_label"] == "covered"]
    res_pos = [r for r in pos if r["output"] in RESOLVED]
    none_pos = [r for r in res_pos if r["output"] == "none"]
    rate_none = _rate(len(none_pos), len(res_pos))
    abst2 = _rate(len(pos) - len(res_pos), len(pos))
    res_ret = [r for r in ret if r["output"] in RESOLVED]
    dvd = collections.Counter("%s|%s" % (r["direct_label"], r["derived_gold"] or "no_source_label") for r in rows2)
    v2 = {"clones": len(p2), "twins_excluded": twins, "direct_label_positives": len(pos), "positives_by_none_kind": dict(collections.Counter(r["none_kind"] or "unspecified" for r in pos)),
          "positives_by_origin": dict(collections.Counter(r["origin"] for r in pos)), "positives_by_ablated_area": dict(sorted(collections.Counter(r["ablated_area_id"] for r in pos).items())),
          "none_rate_answered": rate_none, "none_rate_all_positives": _rate(len(none_pos), len(pos)), "abstention_rate": abst2,
          "none_answers": len(none_pos), "assigned_on_positives": len(res_pos) - len(none_pos),
          "retained_covered": {"n": len(ret), "answered": len(res_ret), "assigned_correct": sum(1 for r in res_ret if r["output"] == "assigned" and r["area_id"] in r["direct_acceptable"]),
                               "none": sum(1 for r in res_ret if r["output"] == "none"), "assigned_wrong": sum(1 for r in res_ret if r["output"] == "assigned" and r["area_id"] not in r["direct_acceptable"]),
                               "abstained": len(ret) - len(res_ret)},
          "no_agreed_gold": sum(1 for r in rows2 if r["direct_label"] is None), "derived_vs_direct": dict(sorted(dvd.items()))}
    bound2 = {"none_rate_answered_min": x2d.REGATE_P2_NONE_MIN, "abstention_max": x2d.REGATE_P2_ABSTAIN_MAX, "positives_min": x2d.REGATE_P2_POSITIVES_MIN}
    note2 = ("P2 re-gate: gold = the clone's own blind label on catalogue-p2 (gap or new; the rev-12 derived gold is reported in derived_vs_direct, never scored); over the direct-label positives V-sem's `none` rate on the "
             "answered positives >= the new-area gate bound and abstention <= the GO cap g; UNMEASURED below %d positives (does not pass: one shot)" % x2d.REGATE_P2_POSITIVES_MIN)
    if len(pos) < x2d.REGATE_P2_POSITIVES_MIN:
        G_["P2"] = {"status": "UNMEASURED", "value": v2, "bound": bound2, "n": len(pos), "note": note2}
    else:
        G_["P2"] = _gate(rate_none["value"] is not None and rate_none["value"] >= x2d.REGATE_P2_NONE_MIN and abst2["value"] <= x2d.REGATE_P2_ABSTAIN_MAX, v2, bound2, len(pos), note=note2)
    return {"gate_set": x2d.REGATE_GATE_SET, "gates": G_, "bars": {"P1": bound1, "P2": bound2}, "p1_items": items, "p2_items": rows2,
            "note": "one shot (PREREGISTRATION 5, rev 13): a P1 or P2 that is not PASS stops the pilot; the rev-12 round-0 gates.json is never rewritten"}


# ------------------------------------------------------------------ labels check (before Jev scores a round)
def _check_units(r, data=None):
    """-> (ids of the checked units, ids of the GO population among them): round 0 the D and P units, later rounds the window units; controls never. The GO population excludes core_gate units (R3-10)."""
    d = _rd(r, data)
    units = x2d.read_jsonl(_need(os.path.join(d, "units.jsonl"), "units"))
    sets = ("D", "P") if r == 0 else ("window",)
    chosen = [u for u in units if not u.get("control") and u.get("set") in sets]
    return {u["case_id"] for u in chosen}, {u["case_id"] for u in chosen if u.get("stratum_reason") != "core_gate"}


def labels_digest(r, data=None):
    """R3-07: SHA-256 of the sorted JSON of the current-guide wide rows (main label set) of the units the labels-check judges; any later label row for them (`label --new-round` fill) changes it."""
    d = _rd(r, data)
    ids, _go = _check_units(r, data)
    rows = [l for l in current_labels(x2d.read_jsonl(os.path.join(d, "labels.jsonl")), x2d.guide_version(r, data))
            if l.get("kind") == "wide" and (l.get("label_set") or "main") == "main" and l.get("case_id") in ids]
    return x2d.sha256_text(json.dumps(sorted(rows, key=lambda l: json.dumps(l, sort_keys=True)), sort_keys=True))


def labels_check(data=None, r=1):
    """Pre-score label-quality gate of round r (PREREGISTRATION 8, 10 (v); R1-09) over the round's scored wide labels of the CURRENT guide version (round 0: D and P units; controls never): kappa on the
    JOINT answer (A vs B before adjudication) over the GO population (core_gate units excluded; the all-unit kappa is reported, R3-10), adjudication rate, rule-citation concentration (R2-08: the top
    rule's share of the A/B sessions that cite a parsable rule id R<n>, each id counted at most once per session) -- REPORT-ONLY, never a trigger (R3-02: the guide's general rules R1/R2 are cited by
    construction) -- plus kappa_covered, raw agreement and prevalence for the record. Fails (ok False) only when kappa_answer < 0.6 (undefined kappa fails). `labels_sha256` binds the check to the label
    rows it judged (R3-07). Pure; the CLI writes rounds/<r>/labels-check.json."""
    d = _rd(r, data)
    gv = x2d.guide_version(r, data)
    ids, go_ids = _check_units(r, data)
    labels = current_labels(x2d.read_jsonl(os.path.join(d, "labels.jsonl")), gv)
    la_all = label_agreement(labels, ids)
    la = label_agreement(labels, go_ids)
    failed = []
    if la["kappa_answer"] is None or la["kappa_answer"] < x2d.MIN_KAPPA:
        failed.append("kappa_answer %s < %.1f (GO population)" % (_fmt(la["kappa_answer"]), x2d.MIN_KAPPA))
    return {"round": r, "guide_version": gv, "ok": not failed, "failed": failed, "labels_sha256": labels_digest(r, data), "n_units": la["n_units"], "n_both_parsed": la["n_both_parsed"],
            "kappa_answer": la["kappa_answer"], "kappa_covered": la["kappa_covered"], "raw_agreement_covered": la["raw_agreement_covered"], "prevalence_covered": la["prevalence_covered"],
            "raw_agreement_answer": la["raw_agreement_answer"], "adjudication_rate": la["adjudication_rate"], "n_units_all": la_all["n_units"], "kappa_answer_all_units": la_all["kappa_answer"],
            "rule_top": la["rule_top"], "rule_top_share": la["rule_top_share"], "rule_top_share_over_0_40": bool(la["rule_top_share"] is not None and la["rule_top_share"] > RULE_CONC),
            "rule_counts": la["rule_counts"], "rule_session_counts": la["rule_session_counts"], "rule_sessions_parsed": la["rule_sessions_parsed"]}


def _planted_new_from_rounds(res_by):
    """Planted new-area positives of the frozen rounds (the empty-file rename traps), pooled from the rounds' `controls.trap` joint counts (R5-07, R6-05; only traps with gold and an arm row): R_new = sum k / sum n
    (arm `none` AND gold not covered / gold not covered), P_new = sum k / sum none_n (arm `none`), Wilson 95 % (descriptive, no veto); `unmeasured` counts traps with no gold or no arm row."""
    k = n = nn = 0
    unm = 0
    for r in res_by.values():
        t = (r.get("controls") or {}).get("trap") or {}
        nc = t.get("none_correct")
        if nc:
            k += nc["k"]
            n += nc["n"]
            nn += t.get("none_n", 0)
        um = (r.get("controls") or {}).get("unmeasured") or {}
        unm += (um.get("trap_gold", 0) or 0) + (um.get("trap_arm", 0) or 0)
    return {"source": "frozen-round empty-file rename traps (one per round); the round-0 planted new concern and the P2 gap gate are in rounds/0/gates.json", "n": n, "none_n": nn, "P_new": _rate(k, nn), "R_new": _rate(k, n), "unmeasured": unm,
            "note": "n planted positives (gold not covered); reported with Wilson 95 %, no veto on the verdict"}


def _require_complete(cases, primary, base, where):
    """R5-09 integrity assertion: every GO-population case has a primary-arm row and a B2 outcome (pooling_rows / paired silently drop a cluster whose outcome is missing)."""
    bad = [k["case_id"] for k in cases if k["out"].get(primary, {}).get("kind", "missing") == "missing" or k["out"].get(base, {}).get("kind", "missing") == "missing"]
    if bad:
        sys.exit("refusing: %s: %d case(s) lack a %s row or a %s outcome: %s" % (where, len(bad), primary, base, ", ".join(bad[:10])))


# ------------------------------------------------------------------ verdict
def verdict(data=None, B=10000):
    vp = os.path.join(data or x2d.DATA, "verdict.json")
    if os.path.exists(vp):
        sys.exit("refusing: %s exists (the verdict is evaluated ONCE)" % vp)
    res_by, parts = {}, []
    for r in x2d.FROZEN_ROUNDS:
        p = os.path.join(_rd(r, data), "results.json")
        if not os.path.exists(p):
            sys.exit("refusing: round %d has no results.json (run --round %d first)" % (r, r))
        res_by[r] = _load_json(p)
    primary = x2d.primary_arm(data)
    su_all, gold, armrows = [], {}, {a: [] for a in dict.fromkeys([primary, "V-sem", "B2"])}
    for r in x2d.FROZEN_ROUNDS:
        L = load_round(r, data)
        for a in armrows:
            if a not in L["arms"]:
                sys.exit("round %d lacks arm %s" % (r, a))
            armrows[a] += L["arms"][a]
        su_all += scored_units(L["units"])
        gold.update(L["gold"])
    units = [u for u in su_all if u.get("stratum_reason") != "core_gate"]  # the GO population (R1-14)
    cases = build_cases(units, gold, armrows)
    pcases = build_cases(su_all, gold, armrows)  # perception samples the scored window units, whatever their stratum
    for r in x2d.FROZEN_ROUNDS:
        cr = [k for k in pcases if k["case_id"].split(":")[0] == str(r)]
        parts.append(perception_counts(load_round(r, data)["labels"], cr, "V-sem", primary))
    percep = perception_pass(parts)
    gp = os.path.join(_rd(0, data), "gates.json")
    n1 = None
    if os.path.exists(gp):
        g = _load_json(gp)["gates"].get("K6_N1")
        if g:
            n1 = {"pass": g["status"] == "PASS", "value": g["value"], "bound": g["bound"]}
    all_inf = all(res_by[r]["informative"]["informative"] for r in x2d.FROZEN_ROUNDS)
    thr = _load_json(os.path.join(HERE, "thresholds.json"))
    _require_complete([k for k in cases if not k["control"]], primary, "B2", "frozen rounds")
    pooled = summarise(armrows, gold, units, jev=primary, informative_all=all_inf, perception=percep, n1=n1, B=B, cases=cases)
    ns = (pooled["gain_guess"]["n"], pooled["gain_all"]["n"], pooled["unresolved"]["n"], pooled["n"])
    if len(set(ns)) != 1:
        sys.exit("refusing: the GO population is not complete (gain_guess.n, gain_all.n, unresolved.n, n = %s; a case with a Jev outcome but no B2 outcome or the reverse; unreachable after _require_complete)" % (ns,))
    v = go_verdict(pooled, thr)
    pooled["planted_new"] = _planted_new_from_rounds(res_by)
    ubr = {str(r): unresolved_block([k["out"][primary] for k in cases if not k["control"] and k["case_id"].split(":")[0] == str(r)]) for r in x2d.FROZEN_ROUNDS}
    pooled_boot = pr_cluster_bootstrap(pooling_rows([k for k in cases if not k["control"]], primary), p_assign_stat, B=B, levels=(0.9, 0.95))
    loo = {}  # R3-15: leave-one-round-out figures of the pooled rounds, descriptive, no verdict (RESULTS reports them)
    for r in x2d.FROZEN_ROUNDS:
        keep = (lambda cid, r=r: str(cid).split(":")[0] != str(r))
        loo[str(r)] = {"left_out": r, "rounds": [q for q in x2d.FROZEN_ROUNDS if q != r],
                       "pooled": summarise({a: [x for x in rows if keep(x["case_id"])] for a, rows in armrows.items()}, gold, [u for u in units if keep(u["case_id"])], jev=primary, informative_all=all(res_by[q]["informative"]["informative"] for q in x2d.FROZEN_ROUNDS if q != r),
                                           perception=perception_pass([p for q, p in zip(x2d.FROZEN_ROUNDS, parts) if q != r]), n1=n1, B=B, cases=[k for k in cases if keep(k["case_id"])])}
    out = {"verdict": v["verdict"], "primary_arm": primary, "go_verdict": v, "pooled": pooled, "leave_one_round_out": loo, "rounds": x2d.FROZEN_ROUNDS, "informative": {str(r): res_by[r]["informative"] for r in res_by},
           "n_core_gate_excluded": len(su_all) - len(units), "unresolved_by_round": ubr,
           "p_assign_bootstrap": {"point": pooled_boot["point"], "ci95": pooled_boot["two_sided"][0.95], "lb90_one_sided": pooled_boot["lower"][0.9], "B": B, "seed": x2d.SEED},
           "per_round": {str(r): {primary: res_by[r]["arms"].get(primary), "V-sem": res_by[r]["arms"].get("V-sem"), "B2": res_by[r]["arms"].get("B2")} for r in res_by},
           "episodes_pooled": _episodes(cases, primary), "perception": percep, "n1": n1,
           "thresholds_sha256": x2d.sha256_file(os.path.join(HERE, "thresholds.json")),
           "reference_arm": {str(r): res_by[r]["arms"].get("reference-claude") for r in res_by}}
    x2d.write_json_new(vp, out)
    return out


def futility_check(upto=None, data=None, write=True):
    summ = []
    r = 1
    while (upto is None or r <= upto) and os.path.exists(os.path.join(_rd(r, data), "results.json")):
        j = _load_json(os.path.join(_rd(r, data), "results.json"))
        summ.append({"round": r, "informative": j["informative"]["informative"], "integrity_pause": j.get("integrity_pause", False), "pooling": j.get("pooling", [])})
        r += 1
    f = futility(summ)
    if f["fired"] and write:
        vp = os.path.join(data or x2d.DATA, "verdict.json")
        if not os.path.exists(vp):
            x2d.write_json_new(vp, {"verdict": "NO-GO", "reason": "futility", "futility": f, "note": "refinement stopped; the frozen rounds are not run (thresholds.json study.futility)"})
    return f


# ------------------------------------------------------------------ CLI
def _verify():
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pin verification failed:\n  " + "\n  ".join(bad))


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--round", type=int)
    ap.add_argument("--pack", action="store_true")
    ap.add_argument("--gates", action="store_true")
    ap.add_argument("--regate-gates", action="store_true", help="rev 13: the re-gate's P1 and P2 on fresh items (run inside D/rounds/0/regate: X2D_DATA); writes regate/gates.json once; exit 1 unless both PASS")
    ap.add_argument("--regate-precheck", action="store_true", help="rev 13 (R17-02): the label-only and construction conjuncts of the re-gate, before any re-gate Jev request (inside D/rounds/0/regate); prints the counts, exit 1 on any problem; reads no Jev file")
    ap.add_argument("--verdict", action="store_true")
    ap.add_argument("--futility", action="store_true")
    ap.add_argument("--upto", type=int)
    ap.add_argument("--repo", help="repository clone (for --gates: the C0 tree of the core gate); default env X2D_REPO")
    ap.add_argument("--labels-check", type=int, metavar="R", help="pre-score label gate of round R (writes rounds/R/labels-check.json; exit 1 only when kappa_answer < 0.6 (GO population) in rounds 1-7 with no guide revision recorded; round 0, rounds 8-10 and a clarified round exit 0 and are scored / non-informative)")
    ap.add_argument("--B", type=int, default=10000)
    a = ap.parse_args(argv)
    _verify()
    if a.labels_check is not None:
        res = labels_check(r=a.labels_check)
        d = _rd(a.labels_check)
        os.makedirs(d, exist_ok=True)
        tmp = os.path.join(d, "labels-check.json.tmp")
        with open(tmp, "w") as f:  # the latest check wins; every check also stays in labels-check.jsonl (a guide clarification re-runs it under a new guide version)
            json.dump(res, f, indent=1, sort_keys=True)
            f.write("\n")
        os.replace(tmp, os.path.join(d, "labels-check.json"))
        x2d.append_jsonl(os.path.join(d, "labels-check.jsonl"), [res])
        print(json.dumps(res, sort_keys=True))
        print("LABELS-CHECK %s" % ("PASS" if res["ok"] else "FAIL: " + "; ".join(res["failed"])))
        d_r = _rd(a.labels_check)
        if not res["ok"] and a.labels_check == 0:
            print("round 0: the labels-check is report-only; K4 on the D units is the round-0 kappa gate")
            return 0
        if not res["ok"] and a.labels_check in x2d.FROZEN_ROUNDS:
            print("round %d is frozen: the guide cannot be clarified; the round is scored and marked non-informative (INCONCLUSIVE)" % a.labels_check)
            return 0
        if not res["ok"] and (x2d.read_jsonl(os.path.join(d_r, "guide-revisions.jsonl")) or os.path.exists(os.path.join(d_r, "clarify-refused.json"))):
            print("round %d already has its one guide clarification (or its pre-check refused it, R4-02): the round proceeds scored and non-informative" % a.labels_check)
            return 0
        return 0 if res["ok"] else 1
    if a.regate_precheck:
        problems, counts = regate_label_problems()
        print("regate precheck: " + json.dumps(counts, sort_keys=True))
        for pr in problems:
            print("PROBLEM: " + pr)
        return 1 if problems else 0
    if a.regate_gates:
        p = os.path.join(x2d.DATA, "gates.json")
        if not x2d.is_regate_root():
            sys.exit("refusing: --regate-gates runs inside D/rounds/0/regate (X2D_DATA)")
        x2d.refuse_if_exists(p)
        res = regate_gates()
        x2d.write_json_new(p, res)
        for k, g in res["gates"].items():
            print("%s %s n=%s bound=%s" % (g["status"], k, g["n"], json.dumps(g["bound"])))
        bad = [k for k, g in res["gates"].items() if g["status"] != "PASS"]
        print("REGATE NOT PASSED (the pilot stops): " + ", ".join(bad) if bad else "REGATE PASS")
        return 1 if bad else 0
    if a.gates:
        if x2d.is_regate_root():
            sys.exit("refusing: the round-0 gates do not run inside D/rounds/0/regate")
        p = os.path.join(_rd(0), "gates.json")
        x2d.refuse_if_exists(p)
        res = gates(repo=a.repo)
        x2d.write_json_new(p, res)
        for k, g in res["gates"].items():
            print("%s %s value=%s bound=%s n=%s" % (g["status"], k, json.dumps(g["value"]), json.dumps(g["bound"]), g["n"]))
        print("fan_out_primary=%s" % res["fan_out_primary"])
        bad = [k for k, g in res["gates"].items() if g["status"] == "FAIL"]
        print("GATES FAIL: " + ", ".join(bad) if bad else "GATES PASS")
        return 1 if bad else 0
    if a.verdict:
        out = verdict(B=a.B)
        print("VERDICT %s" % out["verdict"])
        return 0
    if a.futility:
        f = futility_check(a.upto)
        print(json.dumps(f, sort_keys=True))
        if f["fired"]:
            print("FUTILITY: rule fired; D/verdict.json NO-GO (futility)")
            return 3
        return 0
    if a.round is not None:
        d = _rd(a.round)
        if a.pack and not os.path.exists(os.path.join(d, "results.json")):
            sys.exit("missing results.json: run --round %d first" % a.round)
        if a.pack:
            p = os.path.join(x2d.reviews_dir(), "pilot-round-%d-pack.md" % a.round)
            x2d.refuse_if_exists(p)
            os.makedirs(os.path.dirname(p), exist_ok=True)
            with open(p, "x") as f:
                f.write(build_pack(a.round))
            print("wrote " + p)
            return 0
        rp, mp = os.path.join(d, "results.json"), os.path.join(d, "report.md")
        x2d.refuse_if_exists(rp, mp)
        res = round_results(a.round, B=a.B)
        sp = os.path.join(d, "sessions.json")
        ses = _load_json(sp) if os.path.exists(sp) else {}
        res["session_cap"], res["sessions_used"], res["session_plan"] = ses.get("cap"), ses.get("used"), ses.get("plan")
        x2d.write_json_new(rp, res)
        with open(mp, "x") as f:
            f.write(render_report(res))
        print("round %d: informative=%s clusters=%d prs=%d" % (a.round, res["informative"]["informative"], res["n_scored_clusters"], res["n_prs"]))
        return 0
    ap.error("one of --round, --gates, --labels-check, --verdict, --futility")


if __name__ == "__main__":
    sys.exit(main())
