#!/usr/bin/env python3
"""X2b analysis: X2's frozen `analysis.py` BY IMPORT (decision rule, tables, gates, controls, agreement, rendering: unchanged, pinned by x2b.verify_pins) plus
four X2b-specific blocks that read the collector's records:
  1. `served_check`: every probe (before/after, each repository) names model gpt-6-luna and reasoning effort xhigh, every versions file records the
     preregistered flags and Go configuration, every successful map recorded the preregistered flags and Clawpatch's `agent-start` reported `model=gpt-6-luna`
     from provider codex. Failing any of it makes the verdict NOT citable (X2 section 2's probe-consistency, extended to the pinned configuration).
  2. `combine`: the preregistered decision combination of X2b with X2's frozen codex result (X2b PREREGISTRATION section 4a): reported, never changes the verdict.
  3. no replication block (one arm).
  4. X2b's expected labelled n per repository and operating characteristics (sizes below X2's band), and a rendered citability line that names why a decision is not citable.
Usage: analysis_x2b.py --data D --replay R --pairs P --out results.json [--text analysis_output.txt]
"""
import argparse
import json
import os
import sys

import x2b  # noqa: E402 (puts the frozen X2 directory on sys.path)
import analysis as A  # noqa: E402 (X2's frozen analysis.py)
import common as C  # noqa: E402 (X2's frozen common.py)

X2_T_STAR_PCT = 30  # X2's frozen codex-arm t* (x2b.X2_RESULTS_JSON); used for the reverse cross-check only when X2's file is unavailable
OC_SIZES = (100, 125, 150, 216, 270)  # X2's three sizes plus the two below X2's band that the smoke feature counts make plausible (issue 9)
SMOKE_AGENT_FEATURES = {"fastapi": 26, "axum": 76}  # agent features of the two single smoke maps (PREREGISTRATION section 2, "Pre-run measurement checks"); a construct check on a predictor distribution
EXPECTED_MATCH_SHARE = 0.8  # C1 needs >= 0.8 of prior areas matched at a passing threshold
ARM = "codex"
FLAGS = ["--model", x2b.MAP_MODEL, "--reasoning-effort", x2b.MAP_REASONING_EFFORT]


def served_check(data, repos):
    """-> {"ok", "reasons", "per_repo"}; see the module docstring (1)."""
    reasons, per_repo = [], {}
    go_want = {"model": x2b.MAP_MODEL, "reasoning_effort": x2b.MAP_REASONING_EFFORT, "argv_tail_present": True, "provider_resolver_present": True}
    for repo in sorted(repos):
        tvp = os.path.join(data, "tool-versions-%s-%s.json" % (ARM, repo))
        if not os.path.exists(tvp):
            reasons.append("%s: no tool-versions file" % repo)
            per_repo[repo] = None
            continue
        tv = json.load(open(tvp))
        info = {"probes": {}, "maps": {}}
        for which in ("before", "after"):
            p = tv.get("model_probe_" + which)
            info["probes"][which] = p if isinstance(p, dict) else None
            if not isinstance(p, dict):
                reasons.append("%s: %s probe not taken (%s)" % (repo, which, p))
            elif p.get("error"):
                reasons.append("%s: %s probe failed (%s)" % (repo, which, p["error"]))
            else:
                if p.get("model") != x2b.MAP_MODEL:
                    reasons.append("%s: %s probe served model %r, not %s" % (repo, which, p.get("model"), x2b.MAP_MODEL))
                if p.get("reasoning_effort") != x2b.MAP_REASONING_EFFORT:
                    reasons.append("%s: %s probe served reasoning effort %r, not %s" % (repo, which, p.get("reasoning_effort"), x2b.MAP_REASONING_EFFORT))
        xb = tv.get("x2b") or {}
        if xb.get("map_flags") != FLAGS:
            reasons.append("%s: versions file records map flags %r, not %r" % (repo, xb.get("map_flags"), FLAGS))
        if xb.get("go_map_config") != go_want:
            reasons.append("%s: versions file records Go mapping configuration %r, not %r" % (repo, xb.get("go_map_config"), go_want))
        for mid in C.MAP_IDS:
            m = C.read_meta(data, ARM, repo, mid)
            if m is None or m.get("skipped"):
                continue  # not recorded / skipped: unmeasured by the replay, nothing was served
            row = {"rc": m.get("rc"), "map_flags": m.get("map_flags"), "provider_reported": m.get("provider_reported"), "model_option_reported": m.get("model_option_reported")}
            info["maps"][mid] = row
            if m.get("map_flags") != FLAGS or m.get("map_model") != x2b.MAP_MODEL or m.get("map_reasoning_effort") != x2b.MAP_REASONING_EFFORT:
                reasons.append("%s/%s: meta.json records flags %r, not the preregistered %r" % (repo, mid, m.get("map_flags"), FLAGS))
            if m.get("rc") == 0 and (m.get("provider_reported") != x2b.MAP_PROVIDER or m.get("model_option_reported") != x2b.MAP_MODEL):
                reasons.append("%s/%s: Clawpatch agent-start reported provider=%r model=%r, not %s/%s" % (
                    repo, mid, m.get("provider_reported"), m.get("model_option_reported"), x2b.MAP_PROVIDER, x2b.MAP_MODEL))
        per_repo[repo] = info
    return {"ok": not reasons, "reasons": reasons, "per_repo": per_repo}


def x2_codex_result(x2_dir=None):
    """X2's frozen codex-arm result (verified by digest): {verdict, threshold_pct, pass_pct, citable, models}; None fields when the file is absent."""
    p = os.path.join(x2_dir or x2b.X2_DIR, x2b.X2_RESULTS_JSON)
    if not os.path.exists(p):
        return {"available": False, "reason": "file absent"}
    got = x2b.sha256_file(p)
    if got != x2b.X2_RESULTS_JSON_SHA256:
        return {"available": False, "reason": "sha256 %s != pinned %s" % (got, x2b.X2_RESULTS_JSON_SHA256)}
    r = json.load(open(p))
    d = r["arms"]["codex"]["decision"]
    pt = d.get("per_threshold") or {}
    return {"available": True, "verdict": d["verdict"], "threshold_pct": d.get("threshold_pct"), "pass_pct": sorted(int(k) for k, v in pt.items() if v == "pass"),
            "citable": d.get("citable"), "models": (r.get("probe_consistency", {}).get("codex") or {}).get("models")}


def combine(dec, citable, served, x2):
    """X2b PREREGISTRATION section 4a: what the X2b verdict proposes for `mapping.areas.match_jaccard`. X2b alone decides (it is Scout's production
    mapper); X2's codex result is a reported cross-check on a different model and never overrides, rescues or vetoes. A design round decides; this block
    states the proposal the preregistration fixed. `x2_t_star_status_in_x2b` is X2b's own status at X2's t* (30) so the report states the reverse of
    `x2b_t_star_in_x2_pass_set` too."""
    v, t = dec["verdict"], dec.get("threshold_pct")
    x2_t = x2.get("threshold_pct") if x2.get("available") else None
    out = {"x2b_verdict": v, "x2b_threshold_pct": t, "citable": citable, "served_check_ok": served["ok"], "x2_codex": x2, "x2b_t_star_in_x2_pass_set": None,
           "x2_t_star_status_in_x2b": (dec.get("per_threshold") or {}).get(x2_t if x2_t is not None else X2_T_STAR_PCT)}
    if t is not None and x2.get("available"):
        out["x2b_t_star_in_x2_pass_set"] = t in x2["pass_pct"]
    if not citable:
        out["proposal"] = "none: not citable (probe-inconsistent, served/flags check failed or unexpected arm); match_jaccard stays 0.5 [ASSUMPTION; X2 did not confirm 0.5, design A1]; X2's 0.30 is not adopted (different model); a new preregistration is needed"
    elif v == "adopt":
        in_x2 = out["x2b_t_star_in_x2_pass_set"]
        out["proposal"] = "match_jaccard = %.2f (X2b t*; specific to gpt-6-luna/xhigh; in X2's pass set: %s%s; state non-contiguity, margin and unlabelled share from this file in any design proposal)" % (
            t / 100.0, "unknown" if in_x2 is None else ("yes" if in_x2 else "no"), "; grid minimum: nothing below 0.30 was measured" if t == C.GRID_PCT[0] else "")
    elif v == "fallback":
        out["proposal"] = ("no threshold on the 0.30-0.90 grid meets C1-C3 for gpt-6-luna xhigh (nothing below 0.30 measured): the design's fallback applies (agent map at onboarding and manually); "
                           "X2's codex adopt is not applied to a model Scout no longer uses")
    else:
        out["proposal"] = "none: inconclusive (%s); match_jaccard stays 0.5 [ASSUMPTION; X2 did not confirm 0.5, design A1]; X2's 0.30 is not adopted; a new preregistration is needed" % dec.get("reason")
    if not x2.get("available"):
        out["proposal"] += "; X2 cross-check UNAVAILABLE (%s)" % x2.get("reason")
    return out


def analyze(data, replay_path, pairs_dir, chain_path=None, meta_path=None, x2_dir=None):
    res = A.analyze(data, replay_path, pairs_dir, chain_path, meta_path)
    repos = {r for a, r in {(r["arm"], r["repo"]) for r in C.read_jsonl(replay_path)} if a == ARM}
    extra = sorted(a for a in res["arms"] if a != ARM)
    res["x2b"] = {"arm_check": {"unexpected_arms": extra}, "operating_characteristics": A.operating_characteristics(sizes=OC_SIZES),
                  "expected_labelled_n_per_repo": {r: round(EXPECTED_MATCH_SHARE * 6 * n) for r, n in SMOKE_AGENT_FEATURES.items()}}
    if ARM not in res["arms"]:
        res["x2b"]["served_check"] = {"ok": False, "reasons": ["no codex arm in the replay"], "per_repo": {}}
        res["x2b"]["combination"] = None
        return res
    served = served_check(data, repos)
    dec = res["arms"][ARM]["decision"]
    dec["citable"] = bool(dec.get("citable")) and served["ok"] and not extra
    res["x2b"]["served_check"] = served
    res["x2b"]["combination"] = combine(dec, dec["citable"], served, x2_codex_result(x2_dir))
    res.pop("replication", None)  # one arm: there is no arm B
    return res


def render(res):
    text = A.render(res)
    x = res["x2b"]
    sc = x["served_check"]
    arm = res["arms"].get(ARM)
    lines = text.split("\n")
    if arm is not None:
        pc = res["probe_consistency"].get(ARM) or {}
        line = "   citable for a design change (X2b section 4): %s -- probe-consistency %s; served/flags %s; unexpected arms %s  (X2's own bracket on the arm line above says 'arm probe-inconsistent' for any false value; this line gives the reason)" % (
            "yes" if arm["decision"].get("citable") else "NO", "ok" if pc.get("consistent") else "FAILED %s" % json.dumps(pc.get("reasons")),
            "ok" if sc["ok"] else "FAILED %s" % json.dumps(sc["reasons"]), json.dumps(x["arm_check"]["unexpected_arms"]))
        at = next((i for i, l in enumerate(lines) if l.startswith("== arm %s:" % ARM)), None)
        if at is not None:
            lines.insert(at + 1, line)
        else:
            lines.append(line)
    L = ["", "== X2b (X2b PREREGISTRATION sections 2, 4a)"]
    L.append("served/flags check: %s%s" % ("ok" if sc["ok"] else "FAILED", "" if sc["ok"] else " " + json.dumps(sc["reasons"])))
    if x["arm_check"]["unexpected_arms"]:
        L.append("UNEXPECTED ARMS in the replay (X2b has one arm): %s" % x["arm_check"]["unexpected_arms"])
    L.append("X2b expected labelled n per repository ~ %s [INFERENCE: %.1f x 6 transitions x the smoke-map agent-feature counts %s]; the '216-270' band printed above is X2's, not X2b's; "
             "X2's c3_stops_binding_at_n=%d is a multiple-of-20 search; at integer n C3 can bind only for n <= 126 (PREREGISTRATION section 4(iv))" % (
                 json.dumps(x["expected_labelled_n_per_repo"], sort_keys=True), EXPECTED_MATCH_SHARE, json.dumps(SMOKE_AGENT_FEATURES, sort_keys=True), res["operating_characteristics"]["c3_stops_binding_at_n"]))
    for n, o in x["operating_characteristics"]["by_n"].items():
        L.append("X2b operating characteristics (n=%s per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= %d spurious; both repos): %s" %
                 (n, o["max_spurious_count"], json.dumps({p: round(v["p_both_repos"], 4) for p, v in o["table"].items()})))
    c = x["combination"]
    if c:
        L.append("combination: X2b verdict %s, t* %s, citable %s; X2 codex (frozen) %s; X2b t* in X2 pass set: %s; X2's t* status in X2b (per-threshold at X2's t*): %s" % (
            c["x2b_verdict"], c["x2b_threshold_pct"], c["citable"], json.dumps(c["x2_codex"], sort_keys=True), c["x2b_t_star_in_x2_pass_set"], c["x2_t_star_status_in_x2b"]))
        L.append("PROPOSAL (for a design round, not adopted here): %s" % c["proposal"])
    return "\n".join(lines) + "\n".join(L) + "\n"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default=x2b.DATA)
    ap.add_argument("--replay", required=True)
    ap.add_argument("--pairs", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--text")
    ap.add_argument("--chain")
    ap.add_argument("--replay-meta")
    a = ap.parse_args()
    bad = x2b.verify_pins()
    if bad:
        sys.exit("refusing to analyse: frozen X2 files differ from their pinned digests: %s" % "; ".join(bad))
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
