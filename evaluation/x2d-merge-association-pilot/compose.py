#!/usr/bin/env python3
"""Compose arm outputs from raw Jev scores and the round's thresholds (PREREGISTRATION section 7). No model call.

V-sem / V-sib / V-title per cluster: merge.covered none iff 1 - p_true >= s (s = the tuner's score threshold, t_cov = 1 - s); else merge.which_area top confidence >= t_choice -> assigned(area), where a cluster
asked under a 2nd permutation needs both permutations to agree on the top option (else unresolved:permutation_disagreement) and its confidence is the mean of
the two; below t_choice -> unresolved:below_t_choice; any missing or failed request -> unresolved:failure (a cluster is never dropped).
fan-out (merge.area_covers per area): areas with p_true >= t_fan; non-empty -> assigned to the argmax area (+ acceptable_set), empty -> none; a failed or
missing area request -> unresolved:failure. Thresholds: thresholds["decisions"][q]["threshold"] (t_cov = merge.covered, t_choice = merge.which_area,
t_fan = merge.area_covers); merge.covered additionally stores `score_threshold` = s (absent on a provisional decision: s = 1 - threshold).

Usage: compose.py --round R [--tag contrast] [--scores PATH] [--units PATH] [--arms-dir NAME]
  defaults: D/rounds/R/jev/scores[.TAG].jsonl, units[.TAG].jsonl, arms/; reads D/rounds/R/thresholds.json (refuses when missing), writes
  arms-dir/{V-sem,V-sib,V-title,fan-out}.jsonl for the variants present in the scores file (tag contrast: contrast-V-sem.jsonl); refuses when an output exists.
  Relative --scores/--units paths are relative to the round directory.
  X2D_SKIP_SIGNOFF=1  verify_pins(require_signoff=False) (validation only).
"""
import argparse
import json
import os
import sys
from collections import defaultdict

import x2d

VARIANTS = ("V-sem", "V-sib", "V-title")


def _index(scores_rows):
    """-> (answers, failures): answers[(case_id, question, variant, perm, area_id)] = row (last wins); failures keyed the same (only when no answer)."""
    ans, fail = {}, {}
    for r in scores_rows:
        if r.get("type") not in ("answer", "failure"):
            continue
        k = (r.get("case_id"), r.get("question"), r.get("variant"), r.get("perm"), r.get("area_id"))
        (ans if r["type"] == "answer" else fail)[k] = r
    return ans, {k: v for k, v in fail.items() if k not in ans}


def _row(case_id, arm, output, area_id=None, conf=None, reason=None, **extra):
    return {"case_id": case_id, "arm": arm, "output": output, "area_id": area_id, "confidence": conf, "unresolved_reason": reason, **extra}


def _threshold(thresholds, q):
    return thresholds["decisions"][q]["threshold"]


def cov_score_threshold(thresholds):
    """s of merge.covered in score space (score = 1 - p_true; the tuner flags `none` when score >= s, R1-25)."""
    d = thresholds["decisions"]["merge.covered"]
    return d["score_threshold"] if d.get("score_threshold") is not None else 1.0 - d["threshold"]


def header_flip_all(scores_rows):
    """The `flip_all` flag of the scores header (the driver's -flip-all); False when the file has no header row or the header lacks it."""
    for r in scores_rows:
        if r.get("type") == "header":
            return bool(r.get("flip_all", False))
    return False


def required_perms(unit, flip_all=False):
    """Permutations of merge.which_area a unit must have answered: (1,) and 2 iff x2d.flip_selected(cluster_id) or the run's flip_all."""
    return (1, 2) if (flip_all or x2d.flip_selected(unit["cluster_id"])) else (1,)


def _compose_v(unit, arm, variant, ans, fail, s_cov, t_choice, flip_all=False):
    cid = unit["case_id"]
    cov = ans.get((cid, "merge.covered", variant, 1, None))
    if cov is None:
        cov = next((v for k, v in sorted(ans.items(), key=lambda kv: str(kv[0])) if k[:3] == (cid, "merge.covered", variant)), None)
    raw = {}
    if variant == "V-sem":  # unthresholded fields (R2-09): perception compares none iff not raw_covered else raw_choice
        c1 = ans.get((cid, "merge.which_area", variant, 1, None))
        raw = {"raw_covered": (cov["p_true"] >= 0.5) if cov is not None else None, "raw_choice": c1["choice"] if c1 is not None else None}
    if cov is None:
        return _row(cid, arm, "unresolved", reason="failure", **raw)
    if 1.0 - cov["p_true"] >= s_cov - 1e-12:  # none iff 1 - p_true >= s (the tuner's tie rule)
        return _row(cid, arm, "none", conf=1.0 - cov["p_true"], **raw)
    got = []
    for p in required_perms(unit, flip_all):  # a missing required permutation is a failure, never agreement
        r = ans.get((cid, "merge.which_area", variant, p, None))
        if r is None:
            return _row(cid, arm, "unresolved", reason="failure", **raw)
        got.append(r)
    if len({r["choice"] for r in got}) > 1:
        return _row(cid, arm, "unresolved", reason="permutation_disagreement", **raw)
    conf = sum(r["confidence"] for r in got) / len(got)
    if conf >= t_choice:
        return _row(cid, arm, "assigned", got[0]["choice"], conf, **raw)
    return _row(cid, arm, "unresolved", conf=conf, reason="below_t_choice", **raw)


def _compose_fan(cid, arm, ans, fail, t_fan, area_ids=None):
    keys = [k for k in list(ans) + list(fail) if k[:2] == (cid, "merge.area_covers")]
    variants = {k[2] for k in keys}
    variant = "V-sem" if "V-sem" in variants else (sorted(variants)[0] if variants else None)
    if variant is None or any(k in fail for k in keys if k[2] == variant):
        return _row(cid, arm, "unresolved", reason="failure", acceptable_set=[], raw_covered=None, raw_choice=None)
    if area_ids is not None and set(area_ids) - {k[4] for k in keys if k[2] == variant}:  # R3-06: a partial fan-out (a cost stop between a unit's area requests) is never composed
        return _row(cid, arm, "unresolved", reason="failure", acceptable_set=[], raw_covered=None, raw_choice=None)
    rows = {k[4]: v for k, v in ans.items() if k[:3] == (cid, "merge.area_covers", variant)}
    # R5-01: the fan-out's unthresholded guess (the guard h): covered iff the top area's p_true >= 0.5, the choice = argmax p_true (ties by area id)
    top = max(sorted(rows), key=lambda a: rows[a]["p_true"])
    raw = {"raw_covered": rows[top]["p_true"] >= 0.5, "raw_choice": top}
    hit = {a: r["p_true"] for a, r in rows.items() if r["p_true"] >= t_fan}
    if not hit:
        return _row(cid, arm, "none", acceptable_set=[], **raw)
    best = max(sorted(hit), key=lambda a: hit[a])
    return _row(cid, arm, "assigned", best, hit[best], acceptable_set=sorted(hit), **raw)


def compose(scores_rows, units, thresholds, variant, arm=None, flip_all=None, area_ids=None):
    """-> [arm rows], one per unit (units: unit rows with case_id and cluster_id), variant in V-sem|V-sib|V-title|fan-out. `arm` overrides the arm name. `flip_all` None -> the
    scores header's flag (False without a header). `area_ids`: the catalogue's area ids the fan-out must have answered per unit (None: not checked; a missing area row -> unresolved:failure, R3-06). V-sem rows carry `raw_covered` (p_true >= 0.5 on merge.covered, None when unanswered) and `raw_choice` (permutation-1 top option or None); fan-out rows carry `raw_covered` (max area p_true >= 0.5) and `raw_choice` (argmax p_true, ties by area id; None on a failure row) (R5-01)."""
    flip_all = header_flip_all(scores_rows) if flip_all is None else flip_all
    ans, fail = _index(scores_rows)
    arm = arm or variant
    out = []
    for u in units:
        cid = u["case_id"]
        if variant == "fan-out":
            out.append(_compose_fan(cid, arm, ans, fail, _threshold(thresholds, "merge.area_covers"), area_ids))
        elif variant in VARIANTS:
            out.append(_compose_v(u, arm, variant, ans, fail, cov_score_threshold(thresholds), _threshold(thresholds, "merge.which_area"), flip_all))
        else:
            raise ValueError("unknown arm variant %r" % variant)
    return out


def missing_required(scores, units, flip_all=None, area_ids=None):
    """-> sorted [(case_id, "<variant>/<question>" or "merge.area_covers", permutation or None)] of required requests that have neither an answer nor a failure row.
    scores: a scores.jsonl path or its rows. Required per unit: merge.covered perm 1 and merge.which_area perms required_perms() for every variant present in the file (V-sem always);
    fan-out (any merge.area_covers row in the file): an answer or failure row for EVERY area of the catalogue (`area_ids`; default: the study catalogue's ids, R3-06). `flip_all` None -> the header's flag."""
    rows = x2d.read_jsonl(scores) if isinstance(scores, str) else list(scores)
    flip_all = header_flip_all(rows) if flip_all is None else flip_all
    have = {(r.get("case_id"), r.get("question"), r.get("variant"), r.get("perm")) for r in rows if r.get("type") in ("answer", "failure")}
    variants = {r.get("variant") for r in rows if r.get("type") in ("answer", "failure") and r.get("question") in ("merge.covered", "merge.which_area")} | {"V-sem"}
    fan = any(r.get("question") == "merge.area_covers" for r in rows)
    fan_have = {(r.get("case_id"), r.get("area_id")) for r in rows if r.get("type") in ("answer", "failure") and r.get("question") == "merge.area_covers"}
    if fan and area_ids is None:
        area_ids = [a["area_id"] for a in x2d.read_catalogue()["areas"]]
    out = []
    for u in units:
        cid = u["case_id"]
        for v in sorted(variants):
            if (cid, "merge.covered", v, 1) not in have:
                out.append((cid, v + "/merge.covered", 1))
            for p in required_perms(u, flip_all):
                if (cid, "merge.which_area", v, p) not in have:
                    out.append((cid, v + "/merge.which_area", p))
        if fan:
            for aid in sorted(area_ids):
                if (cid, aid) not in fan_have:
                    out.append((cid, "fan-out/merge.area_covers", aid))
    return sorted(out, key=str)


def compose_k5(scores_rows):
    """Order-flip stats over units with both permutations of merge.which_area answered: {"n", "flips", "rate", "by_variant": {v: {...}}}."""
    ans, _ = _index(scores_rows)
    per = defaultdict(lambda: {"n": 0, "flips": 0})
    for k, r in ans.items():
        if k[1] == "merge.which_area" and k[3] == 1:
            r2 = ans.get((k[0], k[1], k[2], 2, k[4]))
            if r2 is not None:
                per[k[2]]["n"] += 1
                per[k[2]]["flips"] += int(r["choice"] != r2["choice"])
    n, f = sum(v["n"] for v in per.values()), sum(v["flips"] for v in per.values())
    return {"n": n, "flips": f, "rate": (f / n) if n else None,
            "by_variant": {v: {**s, "rate": (s["flips"] / s["n"]) if s["n"] else None} for v, s in sorted(per.items())}}


def main(argv=None):
    ap = argparse.ArgumentParser()
    ap.add_argument("--round", type=int, required=True)
    ap.add_argument("--tag")
    ap.add_argument("--scores")
    ap.add_argument("--units")
    ap.add_argument("--arms-dir", default="arms")
    ap.add_argument("--catalogue", help="catalogue whose area ids the fan-out must have answered per unit (default: the study catalogue; run.sh passes the manifest catalogue of a round-0 run)")
    a = ap.parse_args(argv)
    if a.tag not in (None, "contrast"):
        ap.error("--tag must be contrast")
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pin check failed:\n" + "\n".join(bad))
    if a.catalogue:
        area_ids = [x["area_id"] for x in json.load(open(a.catalogue))["areas"]]
    else:  # the study catalogue; absent only in a fixture without fan-out rows (missing_required then reads it when it is needed)
        area_ids = [x["area_id"] for x in x2d.read_catalogue()["areas"]] if os.path.isfile(x2d.catalogue_path()) else None
    rd = x2d.round_dir(a.round)
    sfx = "." + a.tag if a.tag else ""
    scores_p = os.path.join(rd, a.scores or "jev/scores%s.jsonl" % sfx)
    units_p = os.path.join(rd, a.units or "units%s.jsonl" % sfx)
    th_p = os.path.join(rd, "thresholds.json")
    if not os.path.exists(th_p):
        sys.exit("error: %s missing (run tune.py first)" % th_p)
    thresholds = json.load(open(th_p))
    scores, units = x2d.read_jsonl(scores_p), x2d.read_jsonl(units_p)
    if not scores or not units:
        sys.exit("error: empty or missing scores/units: %s, %s" % (scores_p, units_p))
    miss = missing_required(scores, units, area_ids=area_ids)
    if miss:  # R2-13: a cost stop or a killed driver leaves required requests unanswered; the driver is resumable, composing now would turn them into agreement
        sys.exit("refusing: %d required Jev request(s) of %s have no answer or failure row (re-run the jev stage; first: %s)" % (len(miss), scores_p, miss[:3]))
    present = {r.get("variant") for r in scores if r.get("type") in ("answer", "failure") and r.get("question") in ("merge.covered", "merge.which_area")}
    jobs = []
    if a.tag:
        if "V-sem" in present:
            jobs.append(("V-sem", "contrast-V-sem"))
    else:
        jobs += [(v, v) for v in VARIANTS if v in present]
        if any(r.get("question") == "merge.area_covers" for r in scores):
            jobs.append(("fan-out", "fan-out"))
    if not jobs:
        sys.exit("error: no composable variant in %s" % scores_p)
    outs = {arm: os.path.join(rd, a.arms_dir, arm + ".jsonl") for _, arm in jobs}
    x2d.refuse_if_exists(*outs.values())
    for variant, arm in jobs:
        x2d.write_jsonl_new(outs[arm], compose(scores, units, thresholds, variant, arm, area_ids=area_ids))
        print("composed %s: %d clusters -> %s" % (arm, len(units), outs[arm]))


if __name__ == "__main__":
    main()
