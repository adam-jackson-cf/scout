#!/usr/bin/env python3
"""X2c analysis: X2's frozen `analysis.py` BY IMPORT on every (mapper arm x matching method) cell, plus the X2c blocks (PREREGISTRATION sections 6-8):
  - the cells: for each mapper arm the Jaccard method (X2's replay rows, unchanged) and the Jev method (`jev_match.jev_rows`, replay-schema rows on the same candidate
    pairs, arm name "<arm>+jev"); X2 decides each cell with its own rule (C1-C3, class, label and control gates, t* = max pass);
  - served check per arm (probes, flags, Clawpatch agent-start), Jev check (pin, question version, failed requests, billed cost, latency);
  - cross-repository confirmation of t* (tuned on one repository, scored on the other, both directions), the mapper coverage guard, per-arm map wall time, per-refresh Jev
    cost and time;
  - the selection rule (section 7): eligible cells, fastest refresh time, tie band, method preference, outcome adopt / no-adoption / inconclusive.
Usage: analysis_x2c.py --data D --replay R --pairs P --jev SCORES.jsonl --out results.json [--text analysis_output.txt]
"""
import argparse
import json
import math
import os
import shutil
import sys
import tempfile
from fractions import Fraction

import x2c  # noqa: E402 (puts the frozen X2 directory on sys.path)
import analysis as A  # noqa: E402 (X2's frozen analysis.py)
import common as C  # noqa: E402 (X2's frozen common.py)
import jev_match as J  # noqa: E402

FAMILY_RATES = (0.08, 0.10)  # true spurious rates for the per-cell false-pass probability (PREREGISTRATION section 8)
X2B_LUNA_JACCARD_VERDICT = "fallback"  # X2b RESULTS (../x2b-area-stability-luna): luna-xhigh, Jaccard method, the configuration X2c's luna-xhigh Jaccard cell repeats


def cells_of(arms):
    return [(a, m, a if m == "jaccard" else a + x2c.JEV_SUFFIX) for a in arms for m in x2c.METHODS]


def served_check(data, arm, repos):
    """-> {"ok", "reasons"}: every probe (before/after, each repository) names the arm's model and effort, every versions file and map records the arm's flags, and every
    successful map's Clawpatch agent-start reported the arm's provider and model option."""
    want, reasons = x2c.ARMS[arm], []
    for repo in sorted(repos):
        tvp = os.path.join(data, "tool-versions-%s-%s.json" % (arm, repo))
        if not os.path.exists(tvp):
            reasons.append("%s: no tool-versions file" % repo)
            continue
        tv = json.load(open(tvp))
        for which in ("before", "after"):
            p = tv.get("model_probe_" + which)
            if not isinstance(p, dict):
                reasons.append("%s: %s probe not taken (%s)" % (repo, which, p))
            elif p.get("error"):
                reasons.append("%s: %s probe failed (%s)" % (repo, which, p["error"]))
            else:
                if p.get("model") != want["model"]:
                    reasons.append("%s: %s probe served model %r, not %s" % (repo, which, p.get("model"), want["model"]))
                if p.get("reasoning_effort") != want["effort"]:
                    reasons.append("%s: %s probe served reasoning effort %r, not %s" % (repo, which, p.get("reasoning_effort"), want["effort"]))
        xb = tv.get("x2c") or {}
        if xb.get("map_flags") != x2c.arm_flags(arm):
            reasons.append("%s: versions file records map flags %r, not %r" % (repo, xb.get("map_flags"), x2c.arm_flags(arm)))
        for mid in C.MAP_IDS:
            m = C.read_meta(data, arm, repo, mid)
            if m is None or m.get("skipped"):
                continue
            if m.get("map_flags") != x2c.arm_flags(arm) or m.get("map_model") != want["model"] or m.get("map_reasoning_effort") != want["effort"]:
                reasons.append("%s/%s: meta.json records flags %r, not %r" % (repo, mid, m.get("map_flags"), x2c.arm_flags(arm)))
            if m.get("rc") == 0 and (m.get("provider_reported") != want["provider"] or m.get("model_option_reported") != want["model"]):
                reasons.append("%s/%s: Clawpatch agent-start reported provider=%r model=%r, not %s/%s" % (repo, mid, m.get("provider_reported"), m.get("model_option_reported"), want["provider"], want["model"]))
    return {"ok": not reasons, "reasons": reasons}


def jev_check(header, answers, failures, unscored, arms_repos):
    """-> {"ok", "reasons", "arm_reasons", "per_arm", "failures", ...}. `reasons` (pin, question version, replies from another model, billed cost known) fail every Jev cell;
    `arm_reasons` (unscored share per (arm, repo) <= 1 %) fail the Jev cell of that arm only. `ok` is true iff neither lists anything. Failure rows are reported by kind and the
    number of rows that name a model other than the pin (a billed foreign-model reply later retried successfully is visible here)."""
    reasons, arm_reasons = [], {}
    kinds = {}
    for r in failures:
        kinds[r["kind"]] = kinds.get(r["kind"], 0) + 1
    foreign_failures = sum(1 for r in failures if r.get("model") and r["model"] != x2c.JEV_MODEL)
    if header is None:
        return {"ok": False, "reasons": ["no Jev scores file or no header"], "arm_reasons": {}, "per_arm": {}, "failures": len(failures), "failures_by_kind": kinds, "foreign_model_failure_rows": foreign_failures}
    if header.get("model") != x2c.JEV_MODEL:
        reasons.append("scorer ran model %r, not the pin %s" % (header.get("model"), x2c.JEV_MODEL))
    if header.get("question_version") != x2c.JEV_QUESTION_VERSION:
        reasons.append("question version %r is not the preregistered %s" % (header.get("question_version"), x2c.JEV_QUESTION_VERSION))
    bad_model = sorted({r["model"] for r in answers if r["model"] != x2c.JEV_MODEL})
    if bad_model:
        reasons.append("replies from other models: %s" % bad_model)
    if any(r["question_version"] != x2c.JEV_QUESTION_VERSION for r in answers):
        reasons.append("an answer row carries another question version")
    if any(not r["cost_known"] for r in answers):
        reasons.append("%d replies report no billed cost" % sum(1 for r in answers if not r["cost_known"]))
    per_arm = {}
    for arm, repo in sorted(arms_repos):
        n_un, n_all = unscored.get((arm, repo), [0, 0])
        share = None if n_all == 0 else n_un / n_all
        per_arm.setdefault(arm, {})[repo] = {"unscored": n_un, "candidates": n_all, "unscored_share": share}
        if share is not None and share > x2c.JEV_UNSCORED_MAX:
            arm_reasons.setdefault(arm, []).append("%s/%s: %d of %d candidate occurrences unscored (limit %.2f)" % (arm, repo, n_un, n_all, x2c.JEV_UNSCORED_MAX))
    return {"ok": not reasons and not arm_reasons, "reasons": reasons, "arm_reasons": arm_reasons, "per_arm": per_arm, "failures": len(failures), "failures_by_kind": kinds,
            "foreign_model_failure_rows": foreign_failures}


def jev_refresh_cost(key, answers, arms_repos, rows):
    """Per (arm, repo): mean over the MEASURED transitions (replay rows with status ok at the first grid threshold; a transition without candidates counts as 0) of the billed cost
    of the answer rows, of sum(latency)/JEV_WORKERS (the ASSUMED concurrent runner) and of sum(latency) (`seconds_sequential`: the product runner, decisions.Run, is sequential) of
    scoring that transition's candidates (both orders). The per-refresh figures are answers-only; failure rows enter only `billed_cost_usd` of `analyze`."""
    cost, lat = {}, {}
    for r in answers:
        cost[r["pair_id"]] = cost.get(r["pair_id"], 0.0) + r["cost_usd"]
        lat[r["pair_id"]] = lat.get(r["pair_id"], 0.0) + r["latency_ms"] / 1000.0
    per_tr = {(r["arm"], r["repo"], r["transition"]): [0.0, 0.0, 0, 0.0] for r in rows if r["status"] == "ok" and r["thresholdPct"] == C.GRID_PCT[0]}
    for k in key:
        t = per_tr.setdefault((k["arm"], k["repo"], k["transition"]), [0.0, 0.0, 0, 0.0])
        t[0] += cost.get(k["pair_id"], 0.0)
        t[1] += lat.get(k["pair_id"], 0.0) / x2c.JEV_WORKERS
        t[2] += 1
        t[3] += lat.get(k["pair_id"], 0.0)
    out = {}
    for arm, repo in arms_repos:
        trs = [v for (a, r, _t), v in per_tr.items() if a == arm and r == repo]
        out[(arm, repo)] = None if not trs else {"cost_usd": sum(v[0] for v in trs) / len(trs), "seconds": sum(v[1] for v in trs) / len(trs), "seconds_sequential": sum(v[3] for v in trs) / len(trs),
                                                  "requests_per_refresh": 2 * sum(v[2] for v in trs) / len(trs), "transitions": len(trs)}
    return out


def coverage_check(rows, arms):
    """Mean over the measured transitions of (prevOwnedFiles + nextOwnedFiles) / 2 per (arm, repo): the mapper's file coverage; ok iff >= COVERAGE_FLOOR of the reference arm's
    in every repository. The reference arm unmeasured in a repository -> not evaluable (fail closed). Also reported, not gated: the AREA ratio, mean over the same rows of
    (prevAreas + nextAgentFeatures) / 2 relative to the reference's (a mapper owning as many files in fewer, larger areas passes the file guard)."""
    mean, areas = {}, {}
    for arm in arms:
        for repo in sorted({r["repo"] for r in rows}):
            sel = [r for r in rows if r["arm"] == arm and r["repo"] == repo and r["thresholdPct"] == C.GRID_PCT[0] and r["status"] == "ok"]
            v = [(r["prevOwnedFiles"] + r["nextOwnedFiles"]) / 2 for r in sel]
            mean[(arm, repo)] = None if not v else sum(v) / len(v)
            a = [(r["prevAreas"] + r["nextAgentFeatures"]) / 2 for r in sel]
            areas[(arm, repo)] = None if not a else sum(a) / len(a)
    out = {}
    for arm in arms:
        reasons, ratios, area_ratios = [], {}, {}
        for repo in sorted({r["repo"] for r in rows}):
            ref, got = mean.get((x2c.REFERENCE_ARM, repo)), mean.get((arm, repo))
            aref, agot = areas.get((x2c.REFERENCE_ARM, repo)), areas.get((arm, repo))
            if aref and agot is not None:
                area_ratios[repo] = agot / aref
            if ref is None or got is None:
                reasons.append("%s: coverage not evaluable (reference %s, arm %s)" % (repo, ref, got))
                continue
            ratios[repo] = got / ref
            if got < x2c.COVERAGE_FLOOR * ref:
                reasons.append("%s: mean owned files per map %.1f is %.2f of the reference's %.1f (floor %.2f)" % (repo, got, got / ref, ref, x2c.COVERAGE_FLOOR))
        out[arm] = {"ok": not reasons, "reasons": reasons, "ratio_to_reference": ratios, "area_ratio_to_reference": area_ratios,
                    "mean_owned_files": {r: mean.get((arm, r)) for r in sorted({r for a, r in mean})}, "mean_areas": {r: areas.get((arm, r)) for r in sorted({r for a, r in areas})}}
    return out


def candidate_ceiling(rows, arms):
    """Per arm and repository: mean over the measured transitions of 1 - zeroOverlapPriorAreas / prevAreas, the largest survival any matcher can reach on the candidate set
    (a prior area sharing no owned file with any next feature has no candidate)."""
    out = {}
    for arm in arms:
        for repo in sorted({r["repo"] for r in rows}):
            v = [1 - r["zeroOverlapPriorAreas"] / r["prevAreas"] for r in rows if r["arm"] == arm and r["repo"] == repo and r["thresholdPct"] == C.GRID_PCT[0] and r["status"] == "ok" and r["prevAreas"]]
            out.setdefault(arm, {})[repo] = None if not v else sum(v) / len(v)
    return out


def false_pass_probability(n, p):
    """Exact binomial probability that one repository with n labelled matches and a true spurious rate p meets C2 (point <= 1/20) and C3 (Wilson upper <= 0.10), as X2's
    `operating_characteristics` computes it (conditional on C1, independent matches: anticonservative). None when n is 0 or unknown."""
    if not n:
        return None
    kmax = max((k for k in range(n + 1) if Fraction(k, n) <= A.SPURIOUS_MAX_FRAC and C.wilson(k, n)[1] <= C.SPURIOUS_UPPER_GUARD), default=-1)  # -1: no count meets C2 and C3 at this n
    return sum(math.comb(n, k) * p ** k * (1 - p) ** (n - k) for k in range(kmax + 1))


def map_seconds(inventory, arms, repos):
    """Per arm: mean seconds_total (attempts included) of the successfully recorded maps per repository, and their sum over repositories (one fastapi + one axum map); also the
    dispersion per (arm, repo): n, mean, min, max and the seconds of each map id."""
    out = {}
    for arm in arms:
        per, by_map = {}, {}
        for m in inventory["maps"]:
            if m["arm"] == arm and m["status"] == "ok" and m.get("seconds_total") is not None:
                per.setdefault(m["repo"], []).append(m["seconds_total"])
                by_map.setdefault(m["repo"], {})[m["map"]] = m["seconds_total"]
        means = {r: sum(v) / len(v) for r, v in per.items()}
        out[arm] = {"per_repo_mean": means, "maps": {r: len(v) for r, v in per.items()}, "sum_over_repos": sum(means.values()) if set(means) == set(repos) else None,
                    "dispersion": {r: {"n": len(v), "mean": means[r], "min": min(v), "max": max(v), "by_map": by_map[r]} for r, v in per.items()}}
    return out


def confirm(tables, repos):
    """Two-fold cross-repository confirmation of the threshold (jev-actions: tune on one unit, score on another). For each repository R: t = max threshold at which the OTHER
    repository alone is `pass` (X2's classification); R is scored at t on C1 (mean survival >= 4/5) and C2 (spurious <= 1/20), exact. confirmed iff both repositories are."""
    out = {"per_repo": {}, "confirmed": False}
    if len(repos) != 2:
        out["reason"] = "needs exactly two repositories"
        return out
    ok = True
    for held in repos:
        other = [r for r in repos if r != held][0]
        passing = [p for p in C.GRID_PCT if A.classify_t({other: tables[other]["decisive"][str(p)]}) == "pass"] if other in tables and "decisive" in tables[other] else []
        if not passing:
            out["per_repo"][held] = {"tuned_on": other, "t_pct": None, "ok": False, "reason": "no threshold passes in %s alone" % other}
            ok = False
            continue
        t = max(passing)
        m = tables[held]["decisive"][str(t)]
        c1 = m["survival_exact"] is not None and Fraction(m["survival_exact"]) >= A.SURVIVAL_MIN_FRAC
        c2 = m["labelled"] > 0 and Fraction(m["different"], m["labelled"]) <= A.SPURIOUS_MAX_FRAC
        out["per_repo"][held] = {"tuned_on": other, "t_pct": t, "ok": bool(c1 and c2), "survival": m.get("survival"), "spurious": m.get("spurious"), "labelled": m["labelled"], "C1": c1, "C2": c2}
        ok &= bool(c1 and c2)
    out["confirmed"] = ok
    return out


def select(cells):
    """PREREGISTRATION section 7. cells: {cell id: {arm, method, state, refresh_seconds, margin, t_star_pct}} -> the outcome."""
    eligible = {c: v for c, v in cells.items() if v["state"] == "eligible"}
    if not eligible:
        if any(v["state"] == "inconclusive" for v in cells.values()):
            return {"outcome": "inconclusive", "cell": None, "proposal": "none: inconclusive (no eligible cell and at least one cell inconclusive or not citable); a new preregistration is needed"}
        return {"outcome": "no-adoption", "cell": None, "proposal": "none: no cell is eligible; the design's fallback stays (agent map at onboarding and manually; no refresh-on-merge)"}
    fastest = min(v["refresh_seconds"] for v in eligible.values())
    tied = {c: v for c, v in eligible.items() if v["refresh_seconds"] <= fastest * (1 + x2c.TIE_BAND)}
    order = list(x2c.ARMS)
    best = sorted(tied, key=lambda c: (tied[c]["method"] != "jaccard", -(tied[c]["margin"] if tied[c]["margin"] is not None else -1), order.index(tied[c]["arm"])))[0]
    v = tied[best]
    rule = "match_jaccard = %.2f" % (v["t_star_pct"] / 100) if v["method"] == "jaccard" else "Jev match threshold tau = %.2f on s = P(same_concern) + P(overlapping_concern)" % (v["t_star_pct"] / 100)
    proposal = "adopt mapper %s with the %s method: %s (for a design round, not adopted here; specific to this arm's served model and effort)" % (v["arm"], v["method"], rule)
    ratios = v.get("area_ratio")
    if ratios:
        proposal += "; areas per map: " + ", ".join("%s x%.2f" % (r, q) for r, q in sorted(ratios.items())) + " of the reference"
        if any(q < 0.8 for q in ratios.values()):
            proposal += " (a coarser area granularity: the design round decides whether that is acceptable before this cell is cited)"
    if v["arm"] == x2c.REFERENCE_ARM and v["method"] == "jaccard":
        proposal += "; X2b (same configuration): %s (one-draw verdicts for this configuration are not reproducible)" % X2B_LUNA_JACCARD_VERDICT
    return {"outcome": "adopt", "cell": best, "tied": sorted(tied), "fastest_refresh_seconds": fastest, "proposal": proposal}


def analyze(data, replay_path, pairs_dir, jev_path, chain_path=None, meta_path=None):
    rows = C.read_jsonl(replay_path)
    key = C.read_jsonl(os.path.join(pairs_dir, "key.jsonl"))
    arms = [a for a in x2c.ARMS if any(r["arm"] == a for r in rows)]
    unexpected = sorted({r["arm"] for r in rows} - set(x2c.ARMS))
    header, scores, answers, failures = J.read_scores(jev_path) if os.path.exists(jev_path) else (None, {}, [], [])
    jrows, unscored = J.jev_rows(rows, key, scores, arms=set(arms))
    work = tempfile.mkdtemp(prefix="x2c-analysis-")
    try:
        pdir = os.path.join(work, "pairs")
        os.makedirs(pdir)
        for fn in ("blinded.jsonl", "control-key.jsonl", "labels.jsonl", "calls.jsonl"):
            if os.path.exists(os.path.join(pairs_dir, fn)):
                shutil.copy(os.path.join(pairs_dir, fn), pdir)
        with open(os.path.join(pdir, "key.jsonl"), "w") as f:
            for k in key:
                f.write(json.dumps(k) + "\n")
            for k in key:
                f.write(json.dumps(dict(k, arm=k["arm"] + x2c.JEV_SUFFIX)) + "\n")
        rp = os.path.join(work, "replay.jsonl")
        J.write_jsonl(rp, rows + jrows)
        chain = chain_path or os.path.join(os.path.dirname(os.path.abspath(replay_path)), "chain.jsonl")
        meta = meta_path or os.path.join(os.path.dirname(os.path.abspath(replay_path)), "replay-meta.json")
        res = A.analyze(data, rp, pdir, chain, meta)
    finally:
        shutil.rmtree(work, ignore_errors=True)
    repos = sorted({r["repo"] for r in rows})
    arms_repos = sorted({(r["arm"], r["repo"]) for r in rows})
    served = {a: served_check(data, a, repos) for a in arms}
    jc = jev_check(header, answers, failures, unscored, arms_repos)
    jcost = jev_refresh_cost(key, answers, arms_repos, rows)
    cov = coverage_check(rows, arms)
    ceiling = candidate_ceiling(rows, arms)
    fam = {}
    secs = map_seconds(res["inventory"], arms, repos)
    cells = {}
    for arm, method, cid in cells_of(arms):
        a = res["arms"].get(cid)
        if a is None:
            cells[cid] = {"arm": arm, "method": method, "state": "inconclusive", "reasons": ["no rows"], "refresh_seconds": None, "margin": None, "t_star_pct": None}
            continue
        d = a["decision"]
        reasons, cit = [], True
        if not served[arm]["ok"]:
            cit = False
            reasons += ["served check failed: " + served[arm]["reasons"][0]]
        pc = res["probe_consistency"].get(arm)
        if pc is not None and not pc["consistent"]:
            cit = False
            reasons.append("arm probes inconsistent")
        jr = jc["reasons"] + jc["arm_reasons"].get(arm, [])
        if method == "jev" and jr:
            cit = False
            reasons += ["Jev check failed: " + jr[0]]
        d["citable"] = cit and not unexpected
        conf = confirm(a["tables"], repos) if d["verdict"] == "adopt" else {"confirmed": None, "per_repo": {}}
        jev_caps = None
        refresh = secs[arm]["sum_over_repos"]
        if method == "jev":
            per = [jcost.get((arm, r)) for r in repos]
            if all(per):
                jev_caps = {"cost_usd": sum(p["cost_usd"] for p in per), "seconds": sum(p["seconds"] for p in per), "seconds_sequential": sum(p["seconds_sequential"] for p in per),
                            "requests_per_refresh": {r: jcost[(arm, r)]["requests_per_refresh"] for r in repos}}
                refresh = None if refresh is None else refresh + jev_caps["seconds"]
            else:
                jev_caps = {"cost_usd": None, "seconds": None, "seconds_sequential": None}
        state = None
        if d["verdict"] == "inconclusive" or not d["citable"]:
            state = "inconclusive"
            reasons.append(d.get("reason") or "not citable")
        elif d["verdict"] == "fallback":
            state = "not-adopted"
        else:
            no = []
            if not conf["confirmed"]:
                no.append("cross-repository confirmation failed")
            if not cov[arm]["ok"]:
                no.append("coverage guard: " + "; ".join(cov[arm]["reasons"]))
            if refresh is None:
                no.append("no map wall time recorded")
            if jev_caps is not None:
                if jev_caps["cost_usd"] is None or jev_caps["cost_usd"] > x2c.JEV_COST_PER_REFRESH_MAX_USD:
                    no.append("Jev cost per refresh %s above %.2f USD" % (jev_caps["cost_usd"], x2c.JEV_COST_PER_REFRESH_MAX_USD))
                if jev_caps["seconds"] is None or jev_caps["seconds"] > x2c.JEV_SECONDS_PER_REFRESH_MAX:
                    no.append("Jev time per refresh %s above %.0f s (sum of latencies / %d workers, ASSUMED concurrent runner)" % (jev_caps["seconds"], x2c.JEV_SECONDS_PER_REFRESH_MAX, x2c.JEV_WORKERS))
            state = "ineligible" if no else "eligible"
            reasons += no
        cells[cid] = {"arm": arm, "method": method, "state": state, "reasons": reasons, "verdict": d["verdict"], "t_star_pct": d.get("threshold_pct"), "margin": d.get("survival_margin_at_t_star"),
                      "confirmation": conf, "refresh_seconds": refresh, "jev": jev_caps, "citable": d["citable"], "area_ratio": cov[arm]["area_ratio_to_reference"]}
        ns = {str(t): {repo: ((a["tables"].get(repo) or {}).get("decisive") or {}).get(str(t), {}).get("labelled") for repo in repos} for t in C.GRID_PCT}
        fam[cid] = {"n_labelled": ns, "p_false_pass": {}, "at": {}}
        for p in FAMILY_RATES:
            cands = [(q, t, repo) for t, by in ns.items() for repo, n in by.items() for q in [false_pass_probability(n, p)] if q is not None]
            best = max(cands, default=None)
            fam[cid]["p_false_pass"][str(p)] = None if best is None else best[0]
            fam[cid]["at"][str(p)] = None if best is None else {"pct": int(best[1]), "repo": best[2]}
    sel = select(cells)
    jev_lines = C.read_jsonl(os.path.join(os.path.dirname(os.path.abspath(jev_path)), "code-digests.jsonl")) if os.path.exists(os.path.join(os.path.dirname(os.path.abspath(jev_path)), "code-digests.jsonl")) else []
    family = {"cells_examined": len(cells), "cells_eligible": sum(1 for c in cells.values() if c["state"] == "eligible"), "per_cell": fam, "sum_of_maxima": {
        str(p): (None if any(v["p_false_pass"][str(p)] is None for v in fam.values()) else sum(v["p_false_pass"][str(p)] for v in fam.values())) for p in FAMILY_RATES},
        "note": "per cell: the largest over the 13 grid thresholds and the two repositories of the probability that one repository with that threshold's realised labelled count meets C2/C3 at the true spurious rate "
                "(the one-bad-repository null, the same rule for every cell); sum over cells. Single-threshold, conditional on C1, independent matches: an approximation to the family-wise false-support probability, not an upper bound"}
    l_ver = cells.get(x2c.REFERENCE_ARM, {}).get("verdict")
    res["x2c"] = {"cells": cells, "selection": sel, "served": served, "jev_check": jc, "coverage": cov, "map_seconds": secs, "unexpected_arms": unexpected, "candidate_ceiling": ceiling, "family": family,
                  "jev_code_digests": jev_lines, "x2b_comparison": {"cell": x2c.REFERENCE_ARM, "x2c_verdict": l_ver, "x2b_verdict": X2B_LUNA_JACCARD_VERDICT, "agree": None if l_ver is None else l_ver == X2B_LUNA_JACCARD_VERDICT},
                  "jev_refresh": {"%s/%s" % k: v for k, v in jcost.items()},
                  "jev_requests": {"answers": len(answers), "failures": len(failures), "failures_by_kind": jc["failures_by_kind"], "cost_unknown_rows": sum(1 for r in answers + failures if not r.get("cost_known")),
                                   "billed_cost_usd": sum(r["cost_usd"] for r in answers + failures if r.get("cost_known")),
                                   "attempts_gt1_share": None if not answers else sum(1 for r in answers if r["attempts"] > 1) / len(answers),
                                   "mean_latency_s": None if not answers else sum(r["latency_ms"] for r in answers) / len(answers) / 1000.0},
                  "map_cost": "not recorded: Clawpatch prints no tokens or cost (X2 section 9); map cost is not a decision input"}
    return res


def render(res):
    x = res["x2c"]
    note = ("NOTE `<arm>+jev` rows: created/kept/retired and kept-inclusive survival are schema placeholders (kept = prior - matched, retired = 0), unmeasured; the inventory's "
            "'not recorded' entries for `+jev` arms are the same placeholders (Jev cells run no maps).")
    L = [A.render(res), note, "", "== X2c (mapper x matching method)"]
    L.append("served checks: " + json.dumps({a: v["ok"] for a, v in x["served"].items()}) + (" unexpected arms %s" % x["unexpected_arms"] if x["unexpected_arms"] else ""))
    jc = x["jev_check"]
    L.append("Jev check: %s%s; arm-only failures %s; failure rows by kind %s, foreign-model failure rows %s; requests %s" % (
        "ok" if jc["ok"] else "FAILED", "" if not jc["reasons"] else " -- " + "; ".join(jc["reasons"]), json.dumps(jc["arm_reasons"], sort_keys=True), json.dumps(jc["failures_by_kind"], sort_keys=True),
        jc["foreign_model_failure_rows"], json.dumps(x["jev_requests"], sort_keys=True)))
    L.append("Jev unscored per (arm, repo): " + json.dumps(jc["per_arm"], sort_keys=True))
    L.append("Jev per (arm, repo) refresh (cost, seconds / %d workers, seconds_sequential, requests, transitions): " % x2c.JEV_WORKERS + json.dumps(x["jev_refresh"], sort_keys=True))
    L.append("Jev code digests (one line per `run.sh jev` invocation): " + json.dumps(x["jev_code_digests"], sort_keys=True))
    L.append("map wall seconds: " + json.dumps(x["map_seconds"], sort_keys=True))
    L.append("coverage: " + json.dumps({a: {"ok": v["ok"], "file_ratio": v["ratio_to_reference"], "area_ratio": v["area_ratio_to_reference"]} for a, v in x["coverage"].items()}, sort_keys=True))
    L.append("candidate ceiling (1 - zeroOverlapPriorAreas / prevAreas, mean over measured transitions; the most survival any matcher reaches on the candidate set): " + json.dumps(x["candidate_ceiling"], sort_keys=True))
    for cid, c in x["cells"].items():
        L.append("cell %-18s %-12s verdict %-12s t* %s margin %s refresh %ss confirmed %s jev %s %s" % (cid, c["state"], c.get("verdict"), c["t_star_pct"], c["margin"], c["refresh_seconds"],
                                                                                                (c.get("confirmation") or {}).get("confirmed"), json.dumps(c.get("jev"), sort_keys=True), "; ".join(c["reasons"])))
    L.append("family-wise: %s" % json.dumps(x["family"], sort_keys=True))
    L.append("X2b comparison: %s" % json.dumps(x["x2b_comparison"], sort_keys=True))
    L.append("SELECTION: %s: %s" % (x["selection"]["outcome"], x["selection"]["proposal"]))
    L.append(x["map_cost"])
    return "\n".join(L) + "\n"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default=x2c.DATA)
    ap.add_argument("--replay", required=True)
    ap.add_argument("--pairs", required=True)
    ap.add_argument("--jev", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--text")
    a = ap.parse_args()
    bad = x2c.verify_pins()
    if bad:
        sys.exit("refusing to analyse: frozen X2 files differ from their pinned digests: %s" % "; ".join(bad))
    res = analyze(a.data, a.replay, a.pairs, a.jev)
    with open(a.out, "w") as f:
        json.dump(res, f, indent=1, sort_keys=True, default=str)
    text = render(res)
    if a.text:
        with open(a.text, "w") as f:
            f.write(text)
    print(text)


if __name__ == "__main__":
    main()
