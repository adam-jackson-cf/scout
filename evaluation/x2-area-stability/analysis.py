#!/usr/bin/env python3
"""X2 analysis: join replayed matches with agent labels, compute survival and spurious-merge rates per threshold, apply the
preregistered decision rule (PREREGISTRATION sections 3-7). Standard library only.

Usage: analysis.py --data DIR --replay replay.jsonl --pairs DIR --out results.json [--text analysis_output.txt]
                   [--chain chain.jsonl] [--replay-meta replay-meta.json]   (the last two default to files next to --replay when present)
"""
import argparse
import json
import math
import os
import sys
from fractions import Fraction

import common as C

SURVIVAL_MIN_FRAC = Fraction(4, 5)  # C1 is decided on exact rationals (section 3.2)
SPURIOUS_MAX_FRAC = Fraction(1, 20)  # C2
CONTROL_FAIL_MAX_FRAC = Fraction(1, 5)
UNMEASURED_MAX_FRAC = Fraction(1, 10)
USABLE3 = ("same_concern", "overlapping_concern", "different_concern")


# ---------------------------------------------------------------- label agreement and controls
def agreement(pairs_dir, include=None, exclude=None):
    """A/B raw agreement, Cohen's kappa over pairs where both sessions parsed, adjudication and unparsable counts. Round-1 calls only (calls
    of a later `--new-round` are rows with a `round` key and are excluded), optionally restricted to / excluding a set of pair ids. Session accounting:
    `codex_sessions` counts codex sessions run for these pairs (sum of B's attempts, one retry allowed); `pairs_with_exec_event` counts pairs in which any
    B attempt printed an `exec` event (a per-attempt split is not recorded: `exec_events` is summed over the attempts of a call). The share of adjudicated pairs
    whose C model equals A's is over the adjudicated pairs C answered validly."""
    calls = [c for c in C.read_jsonl(os.path.join(pairs_dir, "calls.jsonl")) if "round" not in c]
    if include is not None:
        calls = [c for c in calls if c["pair_id"] in include]
    if exclude is not None:
        calls = [c for c in calls if c["pair_id"] not in exclude]
    both = [(c["A"]["label"], c["B"]["label"]) for c in calls if c["A"]["ok"] and c["B"]["ok"]]
    n = len(both)
    adj = [c for c in calls if c["C"] is not None]
    c_ok = sum(1 for c in adj if c["C"]["ok"])
    c_eq_a = sum(1 for c in adj if c["C"]["ok"] and c["C"]["model"] == c["A"]["model"])  # C is `--model sonnet`, A the default model
    res = {"pairs": len(calls), "both_parsed": n, "adjudicated": len(adj), "adjudication_rate": None if not calls else len(adj) / len(calls),
           "adjudicated_c_ok": c_ok, "adjudicated_c_model_equals_a": c_eq_a, "adjudicated_c_model_equals_a_of": c_ok,
           "adjudicated_c_model_equals_a_share": None if c_ok == 0 else c_eq_a / c_ok,
           "codex_sessions": sum(c["B"].get("attempts") or 1 for c in calls), "pairs_with_exec_event": sum(1 for c in calls if (c["B"].get("exec_events") or 0) > 0),
           "unparsable_final": sum(1 for c in calls if c["C"] is not None and not c["C"]["ok"] and not (c["A"]["ok"] and c["B"]["ok"] and c["A"]["label"] == c["B"]["label"]))}
    if n:
        po = sum(1 for a, b in both if a == b) / n
        pe = sum((sum(1 for a, _ in both if a == k) / n) * (sum(1 for _, b in both if b == k) / n) for k in C.LABELS)
        res.update({"raw_agreement": po, "cohen_kappa": None if pe == 1 else (po - pe) / (1 - pe)})
    return res


def label_models(pairs_dir):
    """Distinct served models per session role (A claude, B codex, C adjudicator) with the number of call rows, over EVERY call row (real and control pairs,
    all rounds), so a model roll during the labelling run is visible (section 6). A call row keeps only its FINAL attempt's model (label.py overwrites it on
    a retry): a retried call is one row, and a model seen only on a failed first attempt is not recorded."""
    out = {"A": {}, "B": {}, "C": {}}
    for c in C.read_jsonl(os.path.join(pairs_dir, "calls.jsonl")):
        for role in out:
            s = c.get(role)
            if s is not None:
                out[role][s["model"]] = out[role].get(s["model"], 0) + 1
    return out


def control_stats(pairs_dir, labels):
    """Known-answer controls (section 5) and the label-validity gate (section 4, gate 3), POOLED across arms (a protocol-level property; per-arm
    counts are descriptive). Swapped controls expect different_concern, identical-record controls same_concern. Per kind, with USABLE3 as the
    denominator: a miss is any usable label other than the expected one (overlapping_concern included); cannot_tell, unparsable and never-labelled are
    reported separately. The gate fails when a kind with controls wanted (ceil(fraction * unique real pairs)) has none generated, when fewer than
    max(10, half of the generated controls of the kind) carry a usable label, or when misses exceed 20 % of the usable. Generation shortfalls are reported."""
    ctl = C.read_jsonl(os.path.join(pairs_dir, "control-key.jsonl"))
    keyrows = C.read_jsonl(os.path.join(pairs_dir, "key.jsonl"))
    n_real = len({k["pair_id"] for k in keyrows})
    first_j = {}
    for k in keyrows:
        first_j.setdefault(k["pair_id"], k["jaccard"])
    want = {"swapped": math.ceil(C.CONTROL_SWAP_FRACTION * n_real), "identical": math.ceil(C.CONTROL_IDENT_FRACTION * n_real)}
    default_expected = {"swapped": "different_concern", "identical": "same_concern"}
    out = {"n": len(ctl), "n_real_pairs": n_real, "kinds": {}, "per_arm": {}, "swapped_by_source_band_pct": {}, "ok": True, "reason": None}

    def tally(rows):
        usable = sum(1 for c in rows if labels.get(c["pair_id"]) in USABLE3)
        miss = sum(1 for c in rows if labels.get(c["pair_id"]) in USABLE3 and labels[c["pair_id"]] != c["expected"])
        return {"n": len(rows), "usable": usable, "misses": miss, "miss_rate": None if usable == 0 else miss / usable}
    fails = []
    for kind in ("swapped", "identical"):
        rows = [c for c in ctl if c["kind"] == kind]
        dist = {}
        for c in rows:
            lab = labels.get(c["pair_id"], "never_labelled")
            dist[lab] = dist.get(lab, 0) + 1
        t = tally(rows)
        min_usable = max(C.CONTROL_MIN_USABLE, math.ceil(len(rows) / 2))
        k = {**t, "expected": rows[0]["expected"] if rows else default_expected[kind], "wanted": want[kind], "shortfall": max(0, want[kind] - len(rows)),
             "labels": dist, "cannot_tell": dist.get("cannot_tell", 0), "unparsable": dist.get("unparsable", 0), "never_labelled": dist.get("never_labelled", 0),
             "min_usable": min_usable, "applicable": want[kind] > 0 or bool(rows)}
        if k["applicable"]:
            if not rows:
                fails.append("%s controls: none generated (%d wanted)" % (kind, want[kind]))
            elif t["usable"] < min_usable:
                fails.append("%s controls: %d usable (minimum %d of %d generated)" % (kind, t["usable"], min_usable, len(rows)))
            elif Fraction(t["misses"], t["usable"]) > CONTROL_FAIL_MAX_FRAC:
                fails.append("%s controls: %d of %d usable labels are not %s (limit %.2f)" % (kind, t["misses"], t["usable"], k["expected"], float(CONTROL_FAIL_MAX_FRAC)))
        out["kinds"][kind] = k
        for arm in sorted({c["arm"] for c in rows}):
            out["per_arm"].setdefault(arm, {})[kind] = tally([c for c in rows if c["arm"] == arm])
    for c in (c for c in ctl if c["kind"] == "swapped"):
        j = first_j.get(c["source_pair_id"])
        band = C.GRID_PCT[0] if j is None else max([p for p in C.GRID_PCT if j >= p / 100.0] or [C.GRID_PCT[0]])
        b = out["swapped_by_source_band_pct"].setdefault(str(band), {"n": 0, "usable": 0, "misses": 0})
        b["n"] += 1
        if labels.get(c["pair_id"]) in USABLE3:
            b["usable"] += 1
            b["misses"] += labels[c["pair_id"]] != c["expected"]
    if fails:
        out["ok"], out["reason"] = False, "; ".join(fails)
    out["agreement"] = agreement(pairs_dir, include={c["pair_id"] for c in ctl}) if ctl else None
    return out


# ---------------------------------------------------------------- tables
def empty():
    return {"transitions": 0, "unmeasured_transitions": [], "prev_areas": 0, "matched": 0, "same": 0, "overlap": 0, "different": 0,
            "cannot_tell": 0, "unparsable": 0, "never": 0, "created": 0, "retired": 0, "kept": 0, "jaccards": [], "per_transition": [],
            "pairs_labelled": set(), "pairs_all": set()}


def build_tables(rows, key, labels):
    """rows: replay rows. Returns {arm: {repo: {stratum: {pct: cell}}}}; strata: decisive (all six transitions), remap, mutation (classes),
    fresh, same_state (sub-strata of remap, descriptive). A skipped mutation runs no map, so the replay reports its transition unmeasured."""
    index = {(k["arm"], k["repo"], k["transition"], k["prev_feature_id"], k["next_feature_id"]): k["pair_id"] for k in key}
    tables = {}
    for r in rows:
        arm, repo, pct, cls, sub = r["arm"], r["repo"], r["thresholdPct"], r["class"], r.get("sub", "")
        strata = ["decisive", cls] + ([sub] if sub in ("fresh", "same_state") else [])
        for st in strata:
            cell = tables.setdefault(arm, {}).setdefault(repo, {}).setdefault(st, {}).setdefault(pct, empty())
            cell["transitions"] += 1
            reason = r.get("reason") if r["status"] != "ok" else None
            if reason is None and r["prevAreas"] == 0:
                reason = "no prior areas"
            if reason is not None:
                cell["unmeasured_transitions"].append({"transition": r["transition"], "reason": reason})
                continue
            cell["per_transition"].append({"transition": r["transition"], "matched": r["matched"], "prev_areas": r["prevAreas"], "kept": r["kept"]})
            for f in ("prev_areas", "matched", "created", "retired", "kept"):
                cell[f] += r[{"prev_areas": "prevAreas"}.get(f, f)]
            for m in r["matches"]:
                cell["jaccards"].append(m["jaccard"])
                pid = index.get((arm, repo, r["transition"], m["prevFeatureId"], m["nextFeatureId"]))
                if pid is None:
                    sys.exit("match without a labelling key (replay and pairs out of sync): %s/%s/%s %s->%s" %
                             (arm, repo, r["transition"], m["prevFeatureId"], m["nextFeatureId"]))
                lab = labels.get(pid)
                cell["pairs_all"].add(pid)
                if lab == "same_concern":
                    cell["same"] += 1
                elif lab == "overlapping_concern":
                    cell["overlap"] += 1
                elif lab == "different_concern":
                    cell["different"] += 1
                elif lab == "cannot_tell":
                    cell["cannot_tell"] += 1
                elif lab == "unparsable":
                    cell["unparsable"] += 1
                else:
                    cell["never"] += 1
                if lab in USABLE3:
                    cell["pairs_labelled"].add(pid)
    return tables


def metrics(cell):
    m = cell["matched"]
    la = cell["same"] + cell["overlap"] + cell["different"]
    un = cell["cannot_tell"] + cell["unparsable"] + cell["never"]
    pt = cell["per_transition"]
    out = {"transitions": cell["transitions"], "unmeasured_transitions": cell["unmeasured_transitions"], "prev_areas": cell["prev_areas"],
           "matched": m, "created": cell["created"], "retired": cell["retired"], "kept": cell["kept"], "labelled": la,
           "same": cell["same"], "overlapping": cell["overlap"], "different": cell["different"], "unlabelled": un,
           "unlabelled_cannot_tell": cell["cannot_tell"], "unlabelled_unparsable": cell["unparsable"], "unlabelled_never_labelled": cell["never"],
           "pairs_unique_matched": len(cell["pairs_all"]), "pairs_unique_labelled": len(cell["pairs_labelled"])}
    # primary survival: equal weight per refresh (mean of per-transition survival, exact rational), so a large map cannot dilute a failing one
    mean = None if not pt else sum(Fraction(x["matched"], x["prev_areas"]) for x in pt) / len(pt)
    out["survival_exact"] = None if mean is None else "%d/%d" % (mean.numerator, mean.denominator)
    out["survival"] = None if mean is None else float(mean)
    kinc = lambda x: (x["matched"] + x.get("kept", 0)) / x["prev_areas"]  # kept-inclusive survival (rev 2.3 Phase 4 construct: matched or kept), cross-walk only
    out["survival_per_transition"] = [{**x, "survival": x["matched"] / x["prev_areas"], "survival_kept_inclusive": kinc(x)} for x in pt]
    out["survival_kept_inclusive_mean"] = None if not pt else float(sum(Fraction(x["matched"] + x.get("kept", 0), x["prev_areas"]) for x in pt) / len(pt))
    out["survival_worst_transition"] = None if not pt else min(x["matched"] / x["prev_areas"] for x in pt)
    out["survival_pooled"] = None if cell["prev_areas"] == 0 else m / cell["prev_areas"]
    out["survival_pooled_wilson"] = list(C.wilson(m, cell["prev_areas"])) if cell["prev_areas"] else None
    out["spurious"] = None if la == 0 else cell["different"] / la
    out["spurious_wilson"] = list(C.wilson(cell["different"], la)) if la else None
    out["spurious_strict"] = None if la == 0 else (cell["different"] + cell["overlap"]) / la
    out["spurious_worst_case"] = None if m == 0 else (cell["different"] + un) / m
    out["unlabelled_share"] = None if m == 0 else un / m
    js = sorted(cell["jaccards"])
    out["match_jaccard_upper_median"] = None if not js else js[len(js) // 2]  # upper median for an even count
    out["share_matches_jaccard_below_0_8"] = None if not js else sum(1 for x in js if x < 0.8) / len(js)
    return out


# ---------------------------------------------------------------- decision
def classify_t(by_repo):
    """by_repo: {repo: metrics of the decisive stratum at one threshold}. -> pass | imprecise | fail. Every repository is evaluated before a
    status is returned: a failure in any repository is a `fail` even if another repository's spurious rate is undefined (section 4)."""
    fail, imprecise = False, False
    for m in by_repo.values():
        surv = None if m["survival_exact"] is None else Fraction(m["survival_exact"])
        if surv is None or surv < SURVIVAL_MIN_FRAC:  # C1, exact
            fail = True
        elif m["labelled"] == 0:
            # spurious rate undefined: with matches the data cannot show compliance (imprecise); without matches survival has already failed
            if m["matched"] == 0:
                fail = True
            else:
                imprecise = True
        elif Fraction(m["different"], m["labelled"]) > SPURIOUS_MAX_FRAC:  # C2, exact
            fail = True
        elif m["spurious_wilson"][1] > C.SPURIOUS_UPPER_GUARD:  # C3
            imprecise = True
    return "fail" if fail else ("imprecise" if imprecise else "pass")


def decide(prod, class_measured, controls=None):
    """prod: {repo: {pct: metrics}} (decisive stratum, one arm); class_measured: {repo: {class: measured transitions}}; controls: control_stats().
    Returns the preregistered verdict. Gates (section 4) are applied first and the tripped gate is named in `reason`."""
    repos = sorted(prod)
    base = {"per_threshold": {}, "class_gate": {}, "label_gate": {}, "control_gate": None, "tripped_gate": None}
    for pct in C.GRID_PCT:
        base["per_threshold"][pct] = classify_t({r: prod[r][pct] for r in repos})
    trips = []
    for repo in repos:
        need = C.CLASS_MIN_MEASURED
        got = class_measured.get(repo, {})
        base["class_gate"][repo] = {cls: {"measured": got.get(cls, 0), "required": need[cls]} for cls in need}
        for cls in need:
            if got.get(cls, 0) < need[cls]:
                trips.append(("class gate", "%s has %d measured %s transition(s), %d required" % (repo, got.get(cls, 0), cls, need[cls])))
    for repo in repos:
        share = prod[repo][C.GRID_PCT[0]]["unlabelled_share"]  # None = no match at T_MIN (nothing to label)
        mm = prod[repo][C.GRID_PCT[0]]
        ok = share is None or Fraction(mm["unlabelled"], mm["matched"]) <= UNMEASURED_MAX_FRAC
        base["label_gate"][repo] = {"unlabelled_share_at_tmin": share, "ok": ok}
        if not ok:
            trips.append(("label gate", "%s: unlabelled share of matches at T_MIN %.3f above %.2f" % (repo, share, float(UNMEASURED_MAX_FRAC))))
    if controls is not None:
        base["control_gate"] = {"ok": controls["ok"], "reason": controls["reason"]}
        if not controls["ok"]:
            trips.append(("control gate", "label validity: " + controls["reason"]))
    passing = [p for p in C.GRID_PCT if base["per_threshold"][p] == "pass"]
    imprecise = [p for p in C.GRID_PCT if base["per_threshold"][p] == "imprecise"]
    idx = [C.GRID_PCT.index(p) for p in passing]
    base["pass_set_contiguous"] = (not idx) or idx == list(range(idx[0], idx[-1] + 1))
    base["gates_ok"] = not trips
    if trips:
        verdict, t = "inconclusive", None
        base["tripped_gate"] = trips[0][0]
        base["reason"] = "%s: %s" % trips[0] + ("" if len(trips) == 1 else " (+%d more gate(s): %s)" % (len(trips) - 1, "; ".join("%s: %s" % x for x in trips[1:])))
    elif passing:
        verdict, t = "adopt", max(passing)
        base["reason"] = "largest passing threshold (survival, spurious-merge and precision criteria in both repositories)"
    elif imprecise:
        verdict, t = "inconclusive", None
        base["reason"] = "survival and point spurious rate pass (or the spurious rate is undefined) at some threshold but the data cannot show compliance (Wilson upper bound above %.2f)" % C.SPURIOUS_UPPER_GUARD
    else:
        verdict, t = "fallback", None
        base["reason"] = "no threshold meets the criteria: refresh only at onboarding and manually"
    base.update({"verdict": verdict, "threshold_pct": t, "feasible_pct": passing, "imprecise_pct": imprecise})
    margin = None if t is None else min(Fraction(prod[r][t]["survival_exact"]) - SURVIVAL_MIN_FRAC for r in repos)
    base["survival_margin_at_t_star_exact"] = None if margin is None else str(margin)  # min over repositories of (mean survival at t*) - 4/5
    base["survival_margin_at_t_star"] = None if margin is None else float(margin)
    return base


def replication(dec_a, dec_b, probe_consistency):
    """Arm B reported beside arm A (never changes A's verdict). replicated = B's gates pass, A's t* is `pass` in B and arm B is probe-consistent (section 4);
    an arm A that is probe-inconsistent carries no section 4 consequence, an arm B that is would-be-replicated but inconsistent is `not citable`."""
    t = dec_a.get("threshold_pct")
    cit = lambda arm: bool((probe_consistency.get(arm) or {}).get("consistent"))
    out = {"gating": False, "arm_A_threshold_pct": t, "arm_B_verdict": dec_b["verdict"], "arm_B_threshold_pct": dec_b.get("threshold_pct"),
           "arm_B_gates_ok": dec_b.get("gates_ok"), "arm_B_status_at_threshold": None, "arm_A_citable": cit("codex"), "arm_B_citable": cit("claude")}
    if t is None:
        out["status"] = "not applicable (arm A did not adopt)"
        return out
    status = (dec_b.get("per_threshold") or {}).get(t)
    out["arm_B_status_at_threshold"] = status
    replicated = bool(dec_b.get("gates_ok")) and status == "pass"
    if not out["arm_A_citable"]:
        out["status"] = "not citable (arm A probe-inconsistent)"
    elif replicated and not out["arm_B_citable"]:
        out["status"] = "not citable (arm B probe-inconsistent)"
    else:
        out["status"] = "replicated" if replicated else "not replicated"
    return out


def gradient(rows, key, labels, arm):
    """Point-biserial correlation between match Jaccard and label==different over T_MIN matches of the arm's decisive transitions.
    Exploratory diagnostic (a threshold-dependent spurious effect shows as a negative value)."""
    index = {(k["arm"], k["repo"], k["transition"], k["prev_feature_id"], k["next_feature_id"]): k["pair_id"] for k in key}
    xs, ys = [], []
    for r in rows:
        if r["arm"] != arm or r["thresholdPct"] != C.GRID_PCT[0] or r["status"] != "ok" or r["class"] not in C.DECISIVE_CLASSES:
            continue
        for m in r["matches"]:
            lab = labels.get(index[(arm, r["repo"], r["transition"], m["prevFeatureId"], m["nextFeatureId"])])
            if lab in USABLE3:
                xs.append(m["jaccard"])
                ys.append(1.0 if lab == "different_concern" else 0.0)
    n = len(xs)
    if n < 3:
        return {"n": n, "r": None}
    mx, my = sum(xs) / n, sum(ys) / n
    sx = math.sqrt(sum((x - mx) ** 2 for x in xs))
    sy = math.sqrt(sum((y - my) ** 2 for y in ys))
    return {"n": n, "r": None if sx == 0 or sy == 0 else sum((x - mx) * (y - my) for x, y in zip(xs, ys)) / (sx * sy)}


def label_distribution(rows, key, labels, arm, repo):
    """Label distribution of the T_MIN matches (decisive transitions) by class and by threshold band [t, t+0.05) (the last band [0.90, 1.0])."""
    index = {(k["arm"], k["repo"], k["transition"], k["prev_feature_id"], k["next_feature_id"]): k["pair_id"] for k in key}
    by_class, by_band = {}, {}
    for r in rows:
        if r["arm"] != arm or r["repo"] != repo or r["thresholdPct"] != C.GRID_PCT[0] or r["status"] != "ok":
            continue
        for m in r["matches"]:
            lab = labels.get(index[(arm, repo, r["transition"], m["prevFeatureId"], m["nextFeatureId"])], "never_labelled")
            band = max([p for p in C.GRID_PCT if m["jaccard"] >= p / 100.0] or [C.GRID_PCT[0]])  # same doubles as the replay's >= test
            for tbl, k in ((by_class, r["class"]), (by_band, band)):
                d = tbl.setdefault(str(k), {})
                d[lab] = d.get(lab, 0) + 1
    # descriptive over-call check (section 5): the share of `different_concern` among the labelled matches of the top band [0.90, 1.0]. Consistent with the
    # labeller over-calling `different` or with a genuine different concern at a near-identical file set (GUIDE R3); the two are not separable here.
    # Occurrences, with the labelled and total match counts as denominators
    top = by_band.get(str(C.GRID_PCT[-1]), {})
    labelled = sum(top.get(k, 0) for k in USABLE3)
    return {"by_class": by_class, "by_band_pct": by_band,
            "over_call_check_band_90": {"different": top.get("different_concern", 0), "labelled": labelled, "matches": sum(top.values()),
                                        "share": None if labelled == 0 else top.get("different_concern", 0) / labelled}}


def operating_characteristics(sizes=(150, 216, 270)):
    """Probability that one repository with n labelled matches and a true spurious rate p meets C2 and C3 (point <= 5 % and Wilson upper <= 0.10), for each
    n in `sizes`; both repositories independent. Exact binomial, ONE threshold, conditional on C1 passing there, matches treated as independent (anticonservative:
    no selection over the nested thresholds, no cluster effects). Expected n: about 216-270 labelled occurrences per repository at a passing threshold
    (C1 needs >= 0.8 of about 270 prior areas matched); 150 is a stress case (fewer matches)."""
    by_n = {}
    for n in sizes:
        kmax = max(k for k in range(n + 1) if Fraction(k, n) <= SPURIOUS_MAX_FRAC and C.wilson(k, n)[1] <= C.SPURIOUS_UPPER_GUARD)
        table = {}
        for p in (0.03, 0.04, 0.05, 0.07, 0.10):
            one = sum(math.comb(n, k) * p ** k * (1 - p) ** (n - k) for k in range(kmax + 1))
            table[str(p)] = {"p_one_repo": one, "p_both_repos": one * one}
        by_n[str(n)] = {"n_per_repo": n, "max_spurious_count": kmax, "table": table}
    c3_n = next(m for m in range(20, 2000, 20) if C.wilson(m // 20, m)[1] <= C.SPURIOUS_UPPER_GUARD)  # smallest n whose exact-5 % count passes C3
    return {"by_n": by_n, "c3_stops_binding_at_n": c3_n}


def loto(rows, key, labels, arm, prod):
    """Leave-one-transition-out (descriptive), by the ADOPTED rule: hold out transition h in every repository, classify each threshold JOINTLY over the
    repositories on the remaining transitions (C1-C3 only; the gates are not re-applied), `t* = max(pass)`, then report each repository's held-out
    survival, matched count, labelled count and `different` count of h at that `t*`. prod: {repo: {pct: metrics}} of the decisive stratum."""
    index = {(k["arm"], k["repo"], k["transition"], k["prev_feature_id"], k["next_feature_id"]): k["pair_id"] for k in key}
    held_rows = {(r["repo"], r["transition"], r["thresholdPct"]): r for r in rows if r["arm"] == arm and r["class"] in C.DECISIVE_CLASSES}
    measured = {repo: [x["transition"] for x in next(iter(pr.values()))["survival_per_transition"]] for repo, pr in prod.items()}
    out = []
    for h in sorted({t for ts in measured.values() for t in ts}):
        sub = [r for r in rows if r["arm"] == arm and r["transition"] != h]
        cells = build_tables(sub, key, labels).get(arm, {})
        by_pct = {p: {repo: metrics(c["decisive"][p]) for repo, c in cells.items() if "decisive" in c and p in c["decisive"]} for p in C.GRID_PCT}
        passing = [p for p in C.GRID_PCT if len(by_pct[p]) == len(prod) and classify_t(by_pct[p]) == "pass"]
        tstar = max(passing) if passing else None
        per_repo = {}
        for repo in sorted(prod):
            r = None if tstar is None or h not in measured[repo] else held_rows.get((repo, h, tstar))
            if r is None:
                per_repo[repo] = {"measured": h in measured[repo], "held_out_survival": None, "matched": None, "prev_areas": None, "labelled": None, "different": None}
                continue
            labs = [labels.get(index[(arm, repo, h, m["prevFeatureId"], m["nextFeatureId"])]) for m in r["matches"]]
            per_repo[repo] = {"measured": True, "held_out_survival": r["matched"] / r["prevAreas"], "matched": r["matched"], "prev_areas": r["prevAreas"],
                              "labelled": sum(1 for lab in labs if lab in USABLE3), "different": sum(1 for lab in labs if lab == "different_concern")}
        out.append({"held_out": h, "t_star_pct": tstar, "per_repo": per_repo})
    return out


def inventory(data, arms_repos):
    """Map inventory: status, attempts, wall seconds, mutation notes, per recorded map; stop records and tool versions found in data."""
    inv = {"maps": [], "stopped": [], "tool_versions": {}}
    for arm, repo in arms_repos:
        for mid in C.MAP_IDS:
            m = C.read_meta(data, arm, repo, mid)
            if m is None:
                inv["maps"].append({"arm": arm, "repo": repo, "map": mid, "status": "not recorded"})
            else:
                inv["maps"].append({"arm": arm, "repo": repo, "map": mid, "status": "skipped" if m.get("skipped") else ("ok" if m.get("rc") == 0 else "failed"), "rc": m.get("rc"),
                                    "workspace_dirty": m.get("workspace_dirty"),
                                    "attempts": m.get("attempt_count", m.get("attempts")), "seconds_total": m.get("seconds_total", m.get("seconds")),
                                    "first_attempt_started_utc": ((m.get("attempts") or [{}])[0] or {}).get("started_utc") if isinstance(m.get("attempts"), list) else None,
                                    "model_option": m.get("model_option_reported"), "skipped": m.get("skipped")})
    for name in sorted(os.listdir(data)) if os.path.isdir(data) else []:
        p = os.path.join(data, name)
        if name.startswith("stopped-") and name.endswith(".json"):
            inv["stopped"].append(json.load(open(p)))
        elif name.startswith("tool-versions-") and name.endswith(".json"):
            inv["tool_versions"][name] = json.load(open(p))
    return inv


ARM_CLI_KEYS = {"codex": ("codex",), "claude": ("claude", "claude_executable_version")}  # the CLIs an arm's maps and probes run (the other arm's CLI is not its construct)


def probe_consistency(tool_versions, arms_repos):
    """Section 2 / section 10: per arm, the model probes (before and after each repository's collection) are *consistent* iff every (arm, repo) of the
    analysis has a tool-versions file, all its probes were taken without error, they carry one identical value (codex: the `model:` line; claude: the sorted
    comma-joined `modelUsage` key set) and no workspace-level configuration file was present. Otherwise the arm's verdict is reported with the deviation and
    is not citable for a design change. The arm's own CLI `--version`s (ARM_CLI_KEYS) over all its records (both repositories, start and after-probe) are
    listed with their distinct values when they differ; that does not enter the flag."""
    out = {}
    for arm in sorted({a for a, _ in arms_repos}):
        reasons, models, per_repo, seen = [], set(), {}, {k: set() for k in ARM_CLI_KEYS.get(arm, ())}
        for repo in sorted(r for a, r in arms_repos if a == arm):
            tv = tool_versions.get("tool-versions-%s-%s.json" % (arm, repo))
            if tv is None:
                reasons.append("%s: no tool-versions file" % repo)
                per_repo[repo] = {"before": None, "after": None, "cli_versions_before": None, "cli_versions_after": None, "workspace_config_presence": None}
                continue
            before = {k: tv[k] for k in seen if k in tv}
            after = tv.get("cli_versions_after")
            after_arm = {k: after[k] for k in seen if k in after} if isinstance(after, dict) else after
            pres = tv.get("workspace_config_presence")
            per_repo[repo] = {"before": tv.get("model_probe_before"), "after": tv.get("model_probe_after"), "cli_versions_before": before, "cli_versions_after": after_arm,
                              "workspace_config_presence": pres}
            for which in ("before", "after"):
                p = per_repo[repo][which]
                if not isinstance(p, dict):
                    reasons.append("%s: %s probe not taken (%s)" % (repo, which, p))
                elif p.get("error") or not p.get("model"):
                    reasons.append("%s: %s probe failed (%s)" % (repo, which, p.get("error") or "no model"))
                else:
                    models.add(p["model"])
            for src in (before, after_arm):
                if isinstance(src, dict):
                    for k, x in src.items():
                        seen[k].add(x)
            present = sorted(k for k, x in (pres or {}).items() if x)
            if present:
                reasons.append("%s: workspace config present: %s" % (repo, ", ".join(present)))
        if len(models) > 1:
            reasons.append("probes name different models")
        out[arm] = {"consistent": not reasons, "reasons": reasons, "models": sorted(models), "cli_versions_changed": {k: sorted(x) for k, x in seen.items() if len(x) > 1},
                    "per_repo": per_repo}
    return out


def analyze(data, replay_path, pairs_dir, chain_path=None, meta_path=None):
    rows = C.read_jsonl(replay_path)
    key = C.read_jsonl(os.path.join(pairs_dir, "key.jsonl"))
    labels = C.final_labels(pairs_dir, with_unparsable=True)
    ctl_ids = {c["pair_id"] for c in C.read_jsonl(os.path.join(pairs_dir, "control-key.jsonl"))}
    arms_repos = sorted({(r["arm"], r["repo"]) for r in rows})
    tables = build_tables(rows, key, labels)
    inv = inventory(data, arms_repos)
    controls = control_stats(pairs_dir, labels)
    res = {"preregistered": {"grid_pct": C.GRID_PCT, "survival_min": "4/5", "spurious_max": "1/20", "spurious_upper_guard": C.SPURIOUS_UPPER_GUARD,
                             "unmeasured_max": "1/10", "class_min_measured": C.CLASS_MIN_MEASURED, "control_fail_max": "1/5"},
           "label_agreement": agreement(pairs_dir, exclude=ctl_ids), "label_models": label_models(pairs_dir), "controls": controls, "arms": {},
           "operating_characteristics": operating_characteristics(), "inventory": inv, "replay_meta": None,
           "label_integrity": {"labelled_pairs_without_call_row": len(set(C.final_labels(pairs_dir)) - {c["pair_id"] for c in C.read_jsonl(os.path.join(pairs_dir, "calls.jsonl"))})}}
    res["probe_consistency"] = probe_consistency(inv["tool_versions"], arms_repos)
    meta_path = meta_path or os.path.join(os.path.dirname(os.path.abspath(replay_path)), "replay-meta.json")
    if os.path.exists(meta_path):
        res["replay_meta"] = json.load(open(meta_path))
    chain_path = chain_path or os.path.join(os.path.dirname(os.path.abspath(replay_path)), "chain.jsonl")
    chain = C.read_jsonl(chain_path)
    for arm, by_repo in sorted(tables.items()):
        a = {"tables": {}, "gradient_exploratory": gradient(rows, key, labels, arm), "label_distribution": {}, "coverage": {}, "chain": {}, "loto": {}}
        for repo, strata in by_repo.items():
            a["tables"][repo] = {st: {str(p): metrics(c) for p, c in sorted(cells.items())} for st, cells in strata.items()}
            a["label_distribution"][repo] = label_distribution(rows, key, labels, arm, repo)
            a["coverage"][repo] = {r["transition"]: {k: r[k] for k in ("class", "sub", "prevAreas", "nextAgentFeatures", "prevOwnedFiles", "nextOwnedFiles",
                                                                       "sharedOwnedFiles", "zeroOverlapPriorAreas")} if r["status"] == "ok" else {"unmeasured": r.get("reason")}
                                   for r in rows if r["arm"] == arm and r["repo"] == repo and r["thresholdPct"] == C.GRID_PCT[0]}
            ch = {}
            for r in chain:
                if r["arm"] == arm and r["repo"] == repo:
                    ch.setdefault(str(r["thresholdPct"]), []).append(r)
            a["chain"][repo] = ch
        prod = {repo: {p: metrics(c) for p, c in strata["decisive"].items()} for repo, strata in by_repo.items() if "decisive" in strata}
        a["loto"] = loto(rows, key, labels, arm, prod) if prod else []
        class_measured = {repo: {cls: (lambda c: c["transitions"] - len(c["unmeasured_transitions"]))(strata[cls][C.GRID_PCT[0]]) for cls in C.DECISIVE_CLASSES
                                 if cls in strata} for repo, strata in by_repo.items()}
        a["decision"] = decide(prod, class_measured, controls) if prod else {"verdict": "inconclusive", "reason": "no decisive transitions", "threshold_pct": None,
                                                                              "gates_ok": False}
        t_star = a["decision"].get("threshold_pct")
        a["at_t_star"] = None if t_star is None else {repo: {k: a["tables"][repo]["decisive"][str(t_star)][k] for k in ("matched", "labelled", "unlabelled", "unlabelled_share", "spurious_worst_case")}
                                                      for repo in sorted(prod)}
        a["decision"]["citable"] = bool(res["probe_consistency"].get(arm, {}).get("consistent"))  # section 4: a probe-inconsistent arm's verdict is not citable
        res["arms"][arm] = a
    dec, rep = res["arms"].get("codex", {}).get("decision"), res["arms"].get("claude", {}).get("decision")
    if dec and rep:
        res["replication"] = replication(dec, rep, res["probe_consistency"])
    return res


# ---------------------------------------------------------------- rendering
def render(res):
    L = []
    f3 = lambda v: "  -  " if v is None else "%.3f" % v
    wl = lambda v: "[  -  ,  -  ]" if v is None else "[%.3f,%.3f]" % tuple(v)
    om = res.get("replay_meta")
    if om:
        L.append("replay: scout %s%s areas.go %s records.go %s %s grid %s" % (om["scoutHead"][:12], " (dirty)" if om.get("areasOrRecordsWorkingTreeDirty") else "",
                                                                          om["areasGoSha256"][:12], om["recordsGoSha256"][:12], om["goVersion"], om["gridPct"]))
    ag = res["label_agreement"]
    L.append("label agreement (real pairs, round 1): %s" % json.dumps(ag))
    L.append("label call rows by role, served model (final attempt's model per call row; every call row incl. controls and later rounds; a model seen only on a failed first attempt is not recorded): %s" % json.dumps(res["label_models"], sort_keys=True))
    c = res["controls"]
    L.append("controls: pooled across arms; %d unique real pairs; %d pairs, gate %s%s" % (c["n_real_pairs"], c["n"], "ok" if c["ok"] else "FAILED", "" if c["ok"] else " (%s)" % c["reason"]))
    for kind, k in c["kinds"].items():
        L.append("   %-9s n=%d (wanted %d, shortfall %d) usable=%d (min %d) misses(not %s)=%d rate=%s cannot_tell=%d unparsable=%d never=%d labels=%s" % (
            kind, k["n"], k["wanted"], k["shortfall"], k["usable"], k["min_usable"], k["expected"], k["misses"], f3(k["miss_rate"]), k["cannot_tell"], k["unparsable"],
            k["never_labelled"], json.dumps(k["labels"], sort_keys=True)))
    L.append("   per arm (descriptive): %s" % json.dumps(c["per_arm"], sort_keys=True))
    L.append("   swapped controls by source-pair Jaccard band: %s" % json.dumps(c["swapped_by_source_band_pct"], sort_keys=True))
    L.append("   label integrity: %s" % json.dumps(res["label_integrity"]))
    L.append("   control agreement: %s" % json.dumps(c["agreement"]))
    inv = res["inventory"]
    bad = [m for m in inv["maps"] if m["status"] != "ok" or (m.get("attempts") or 1) > 1 or m.get("skipped") or m.get("workspace_dirty")]
    L.append("map inventory: %d recorded ok of %d; retried/failed/skipped/unrecorded/workspace dirty after the map: %s" % (sum(1 for m in inv["maps"] if m["status"] == "ok"), len(inv["maps"]), json.dumps(bad)))
    if inv["stopped"]:
        L.append("STOPPED (section 9): %s" % json.dumps(inv["stopped"]))
    for arm, pc in sorted(res["probe_consistency"].items()):
        L.append("probe consistency (section 2, arm %s): %s%s; models %s; CLI versions changed between the probes: %s" % (
            arm, "CONSISTENT" if pc["consistent"] else "INCONSISTENT (the arm's verdict is reported with this deviation and is not citable for a design change)",
            "" if pc["consistent"] else " -- " + "; ".join(pc["reasons"]), json.dumps(pc["models"]), json.dumps(pc["cli_versions_changed"])))
        for repo, p in sorted(pc["per_repo"].items()):
            L.append("   %s probes: before %s, after %s; CLI versions before %s, after %s" % (
                repo, json.dumps(p["before"], sort_keys=True), json.dumps(p["after"], sort_keys=True), json.dumps(p["cli_versions_before"], sort_keys=True),
                json.dumps(p["cli_versions_after"], sort_keys=True)))
            L.append("   %s workspace config presence (names only): %s" % (repo, json.dumps(p["workspace_config_presence"], sort_keys=True)))
    oc = res["operating_characteristics"]
    for n, o in oc["by_n"].items():
        L.append("operating characteristics (n=%s per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= %d spurious; both repos): %s" %
                 (n, o["max_spurious_count"], json.dumps({p: round(v["p_both_repos"], 4) for p, v in o["table"].items()})))
    L.append("   expected n 216-270 labelled occurrences per repo; C3 stops binding at n=%d" % oc["c3_stops_binding_at_n"])
    if "replication" in res:
        L.append("replication: %s (arm A citable: %s, arm B citable: %s)" % (json.dumps(res["replication"]), res["replication"]["arm_A_citable"], res["replication"]["arm_B_citable"]))
    for arm, a in sorted(res["arms"].items()):
        d = a["decision"]
        L.append("== arm %s: verdict %s, threshold %s (%s)%s [citable for a design change: %s]" % (
            arm, d["verdict"], d.get("threshold_pct"), d.get("reason"), "" if d.get("pass_set_contiguous", True) else "  [pass set NOT contiguous: %s]" % d.get("feasible_pct"),
            "yes" if d.get("citable") else "NO (arm probe-inconsistent)"))
        L.append("   per-threshold status: %s" % json.dumps(d.get("per_threshold")))
        L.append("   gradient (Jaccard vs different, exploratory): %s" % a["gradient_exploratory"])
        L.append("   survival margin at t* (min over repos of mean survival - 4/5): %s (%s)" % (d.get("survival_margin_at_t_star"), d.get("survival_margin_at_t_star_exact")))
        if d.get("threshold_pct") is not None:
            L.append("   AT t*=%s (section 7, prominent): unlabelled share of matches and worst-case spurious rate ((different + unlabelled) / matches) per repo: %s" % (
                d["threshold_pct"], json.dumps(a["at_t_star"], sort_keys=True)))
        for lo in a["loto"]:
            L.append("   leave-one-transition-out, joint t* (descriptive) hold out %s: t*=%s per repo (held-out survival, matched, labelled, different): %s" % (
                lo["held_out"], lo["t_star_pct"], json.dumps(lo["per_repo"], sort_keys=True)))
        t = d.get("threshold_pct")
        for repo, cov in sorted(a["coverage"].items()):
            L.append("-- %s coverage (T_MIN rows): %s" % (repo, json.dumps(cov)))
            L.append("-- %s label distribution: %s" % (repo, json.dumps(a["label_distribution"][repo], sort_keys=True)))
            oc_ = a["label_distribution"][repo]["over_call_check_band_90"]
            L.append("-- %s over-call check (`different` share of labelled matches in the [0.90, 1.0] band): %d of %d labelled (%d matches): share %s" % (
                repo, oc_["different"], oc_["labelled"], oc_["matches"], f3(oc_["share"])))
            chain = a["chain"].get(repo, {})
            for ct in ([str(t)] if t else []) + (["50"] if not t else []):
                if ct in chain:
                    L.append("-- %s chained survival at %s%%%s: %s" % (repo, ct, "" if t else " (design default; no adopted threshold)",
                                                                    json.dumps([{k: s.get(k) for k in ("step", "status", "activePrior", "matched", "created", "kept", "retired", "f1MatchedAllSoFar", "f1Areas")} for s in chain[ct]])))
        for repo, strata in sorted(a["tables"].items()):
            for st in ("decisive", "remap", "mutation", "fresh", "same_state"):
                if st not in strata:
                    continue
                L.append("-- %s / %s" % (repo, st))
                L.append("   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst")
                for p, m in sorted(strata[st].items(), key=lambda kv: int(kv[0])):
                    L.append("   %.2f %5d %5d  %s  %s  %s %s %5d %5d %5d %s %s | %5d/%-5d %4d %5d %5d (%d/%d/%d)  %s %s %s %s" % (
                        int(p) / 100, m["prev_areas"], m["matched"], f3(m["survival"]), f3(m["survival_worst_transition"]), f3(m["survival_pooled"]),
                        wl(m["survival_pooled_wilson"]), m["created"], m["retired"], m["kept"], f3(m["match_jaccard_upper_median"]),
                        f3(m["share_matches_jaccard_below_0_8"]), m["labelled"], m["pairs_unique_labelled"], m["different"], m["overlapping"], m["unlabelled"],
                        m["unlabelled_cannot_tell"], m["unlabelled_unparsable"], m["unlabelled_never_labelled"], f3(m["spurious"]), wl(m["spurious_wilson"]),
                        f3(m["spurious_strict"]), f3(m["spurious_worst_case"])))
                pts = next(iter(strata[st].values()))
                L.append("   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: " + json.dumps(
                    {p: [(x["transition"], "%d/%d" % (x["matched"], x["prev_areas"]), "%.3f" % x["survival_kept_inclusive"]) for x in m["survival_per_transition"]]
                     for p, m in sorted(strata[st].items(), key=lambda kv: int(kv[0])) if int(p) in (30, 50, 70, 90) or (t is not None and int(p) == t)}))
                if pts["unmeasured_transitions"]:
                    L.append("   unmeasured transitions: %s" % json.dumps(pts["unmeasured_transitions"]))
    return "\n".join(L) + "\n"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default=C.DATA)
    ap.add_argument("--replay", required=True)
    ap.add_argument("--pairs", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--text")
    ap.add_argument("--chain")
    ap.add_argument("--replay-meta")
    a = ap.parse_args()
    res = analyze(a.data, a.replay, a.pairs, a.chain, a.replay_meta)
    with open(a.out, "w") as f:
        json.dump(res, f, indent=1, sort_keys=True)
    text = render(res)
    if a.text:
        with open(a.text, "w") as f:
            f.write(text)
    print(text)


if __name__ == "__main__":
    main()
