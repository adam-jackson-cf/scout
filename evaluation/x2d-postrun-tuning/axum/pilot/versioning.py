#!/usr/bin/env python3
"""Per-round version freeze and refinement-envelope check (PREREGISTRATION section 10). No model call.

ensure_inputs(r) creates D/rounds/R/inputs/{GUIDE.md,questions/*.json} (R2-01): round 0 and 1 are copied from the study directory (the frozen v_0 base), round r >= 2 from
round r-1's inputs, unless the drafter's proposal already sits there (the council critique of round r-1 approves its digest, `next_inputs_sha256`). Everything below hashes the
round's inputs, never the study-directory files.

freeze_version(r) writes D/rounds/R/version.json (+ settings.json from x2d.round_settings when absent, validated against x2d.SETTING_RANGES): SHA-256 of the
question files (and of each file minus its wording keys = structure), the state builder (build_units.py, frontend.py, b2.py), the guide, the inputs digest, the frozen
thresholds.json, the catalogue, prepass/rounds.json; the settings; the model pin and PR cap; the question version keys (from the Jev header of an existing
jev/scores.jsonl, else null); the Go code-under-test digests. version_drift(r) recomputes it and lists every difference to version.json (R2-03: the settings, the inputs and the
code cannot change after the freeze), except `question_version_keys` and the guide fields of a recorded guide revision (rounds/R/guide-revisions.jsonl).

check_envelope(prev, new) -> [violations]. Permissible per-round changes only: question `instructions` / `criteria` (and the other wording keys: criteria
template, role, version_key, variant descriptions and instruction deltas), menu settings within ranges (sample files 3-8, excerpt 30-120 lines, v_sib / include_pr_title
bool (include_pr_title=False removes state.pr.title from V-sem/V-sib; the V-title arm is unaffected), field_order a permutation of the same four fields, order_rule in {permutation:unit_id, fixed}), the option-order rule of a question within the same set,
guide clarifications. Everything else (catalogue, state-builder code, frozen thresholds.json, question kind/state_shape/model/scoring, pin, PR cap, round
table, code under test) is a violation. From round 8 on ANY change relative to v_7 is a violation (tuned thresholds are not part of version.json).

Usage: versioning.py --round R              freeze (creates the inputs when absent; refuses when version.json exists)
       versioning.py --round R --inputs     only create the round's inputs (idempotent; used before round 0's quality labelling)
       versioning.py --round R --check      compare round R's version (existing file, else computed) with round R-1's (round 7's for R >= 8); non-zero on violations
       versioning.py --round R --drift      non-zero when the live inputs, settings, thresholds or code differ from the frozen version.json (run.sh require_version)
       versioning.py --round R --clarify "reason"   mid-round guide clarification (PREREGISTRATION 10 (v)), rounds 1-7, before jev/scores.jsonl exists: records the edited inputs/GUIDE.md
       X2D_SKIP_SIGNOFF=1  verify_pins(require_signoff=False) (validation only; refused outside a validation environment).
"""
import argparse
import glob
import json
import os
import shutil
import sys
import time

import x2d

WORDING_KEYS = ("instructions", "criteria", "criteria_template", "role", "version_key", "option_order_rule")
VARIANT_WORDING = ("description", "instructions_delta")
ORDER_RULES = ("permutation:unit_id", "fixed")
FIELDS = ["unit", "pr", "diff_summary", "catalogue"]
STATE_BUILDER = ("build_units.py", "frontend.py", "b2.py")
FROZEN_BASE_ROUND = 7
IGNORED_WHEN_FROZEN = ("round", "question_version_keys")


def _sha(path):
    return x2d.sha256_file(path) if os.path.exists(path) else None


def settings_violations(s):
    v = []
    if not isinstance(s, dict) or set(s) != set(x2d.DEFAULT_SETTINGS):
        return ["settings keys %s != %s" % (sorted(s) if isinstance(s, dict) else s, sorted(x2d.DEFAULT_SETTINGS))]
    for k, (lo, hi) in x2d.SETTING_RANGES.items():
        if not isinstance(s[k], int) or isinstance(s[k], bool) or not lo <= s[k] <= hi:
            v.append("setting %s=%r outside %d-%d" % (k, s[k], lo, hi))
    for k in ("v_sib", "include_pr_title"):
        if not isinstance(s[k], bool):
            v.append("setting %s=%r is not a bool" % (k, s[k]))
    if not isinstance(s["field_order"], list) or sorted(s["field_order"]) != sorted(FIELDS):
        v.append("setting field_order=%r is not a permutation of %s" % (s["field_order"], FIELDS))
    if s["order_rule"] not in ORDER_RULES:
        v.append("setting order_rule=%r not in %s" % (s["order_rule"], list(ORDER_RULES)))
    return v


def _structure(q):
    q = {k: v for k, v in q.items() if k not in WORDING_KEYS}
    if isinstance(q.get("variants"), dict):
        q["variants"] = {n: {k: v for k, v in d.items() if k not in VARIANT_WORDING} for n, d in q["variants"].items()}
    return x2d.sha256_text(json.dumps(q, sort_keys=True))


def study_inputs():
    """-> (guide path, question directory) of the frozen v_0 base in the study directory."""
    return os.path.join(x2d.HERE, "GUIDE.md"), os.path.join(x2d.HERE, "questions")


def ensure_inputs(r, data=None):
    """Create rounds/<r>/inputs when absent (round 0/1: from the study directory, r >= 2: from round r-1's inputs). -> the inputs directory."""
    d = x2d.inputs_dir(r, data)
    if os.path.isdir(d):
        return d
    if r <= 1:
        g, qd = study_inputs()
    else:
        src = x2d.inputs_dir(r - 1, data)
        if not os.path.isdir(src):
            sys.exit("error: %s is missing (round %d's inputs are copied from round %d's)" % (src, r, r - 1))
        g, qd = os.path.join(src, "GUIDE.md"), os.path.join(src, "questions")
    os.makedirs(os.path.join(d, "questions"))
    shutil.copyfile(g, os.path.join(d, "GUIDE.md"))
    for p in sorted(glob.glob(os.path.join(qd, "*.json"))):
        shutil.copyfile(p, os.path.join(d, "questions", os.path.basename(p)))
    return d


def _inputs_of(r, data=None):
    """-> (guide path, question directory, inputs digest) the round uses: its inputs when they exist, else the ones freeze will copy (round 0: the study base)."""
    d = x2d.proposed_inputs_dir(r, data)
    if d is None or not os.path.isdir(d):
        g, qd = study_inputs()
        return g, qd, None
    return os.path.join(d, "GUIDE.md"), os.path.join(d, "questions"), x2d.inputs_digest(d)


def build_version(r, data=None, settings=None):
    settings = settings if settings is not None else x2d.round_settings(r, data)
    guide, qdir, inputs_sha = _inputs_of(r, data)
    qfiles, meta, structure = {}, {}, {}
    for p in sorted(glob.glob(os.path.join(qdir, "*.json"))):
        name = os.path.basename(p)
        q = json.load(open(p))
        qfiles[name] = x2d.sha256_file(p)
        structure[name] = _structure(q)
        meta[name] = {"kind": q.get("kind"), "model": q.get("model"), "state_shape": sorted(q.get("state_shape", [])), "option_order_rule": q.get("option_order_rule")}
    keys = None
    sp = os.path.join(x2d.round_dir(r, data), "jev", "scores.jsonl")
    rows = x2d.read_jsonl(sp)
    if rows and rows[0].get("type") == "header":
        keys = rows[0].get("versions")
    cat = x2d.catalogue_path(data)
    gsha = _sha(guide)
    return {"round": r, "question_files": qfiles, "question_structure": structure, "question_meta": meta,
            "state_builder": {n: _sha(os.path.join(x2d.HERE, n)) for n in STATE_BUILDER},
            "guide_sha256": gsha, "inputs_sha256": inputs_sha, "thresholds_sha256": _sha(os.path.join(x2d.HERE, "thresholds.json")),
            "catalogue_sha256": _sha(cat), "prepass_rounds_sha256": _sha(os.path.join(x2d.HERE, "prepass", "rounds.json")),
            "model": x2d.JEV_MODEL, "pr_cap": x2d.PR_CAP, "guide_version": x2d.GUIDE_VERSION_PREFIX + gsha[:12], "settings": settings,
            "question_version_keys": keys, "code_digests": x2d.go_code_digests()}


def freeze_version(r, data=None):
    rd = x2d.round_dir(r, data)
    vp, sp = os.path.join(rd, "version.json"), os.path.join(rd, "settings.json")
    x2d.refuse_if_exists(vp)
    ensure_inputs(r, data)
    settings = x2d.round_settings(r, data)
    bad = settings_violations(settings)
    if bad:
        sys.exit("settings invalid: " + "; ".join(bad))
    v = build_version(r, data, settings)
    if not os.path.exists(sp):
        x2d.write_json_new(sp, settings)
    x2d.write_json_new(vp, v)
    return v


def revisions_path(r, data=None):
    return os.path.join(x2d.round_dir(r, data), "guide-revisions.jsonl")


def effective_version(r, data=None):
    """version.json of round r with the guide fields of its last recorded guide revision (PREREGISTRATION 10 (v)) applied. Raises SystemExit when version.json is missing."""
    v = _load(os.path.join(x2d.round_dir(r, data), "version.json"))
    revs = x2d.read_jsonl(revisions_path(r, data))
    if revs:
        for k in ("guide_sha256", "guide_version", "inputs_sha256"):
            v[k] = revs[-1][k]
    return v


def version_drift(r, data=None):
    """-> [keys] in which the live inputs/settings/thresholds/code differ from the frozen version of round r (R2-03); question_version_keys are written by the Jev run and ignored."""
    want = effective_version(r, data)
    got = build_version(r, data)
    return [k for k in sorted(set(want) | set(got)) if k != "question_version_keys" and want.get(k) != got.get(k)]


def clarify(r, reason, data=None):
    """Record a guide clarification of round r (R2-01 (4)): rounds 1-7, before the round's Jev scores exist; the edited inputs/GUIDE.md must differ from the last known guide. -> the row."""
    if r not in x2d.REFINEMENT_ROUNDS:
        sys.exit("refusing: guide clarification is permitted in rounds 1-7 only")
    if os.path.exists(os.path.join(x2d.round_dir(r, data), "jev", "scores.jsonl")):
        sys.exit("refusing: round %d has Jev scores; a clarification must precede scoring" % r)
    if x2d.read_jsonl(revisions_path(r, data)):
        sys.exit("refusing: round %d already has its one recorded guide clarification (C <= 1, PREREGISTRATION 10 (v))" % r)
    if not reason.strip():
        sys.exit("refusing: a clarification needs a reason")
    known = effective_version(r, data)
    gsha = _sha(x2d.guide_path(r, data))
    if gsha is None or gsha == known["guide_sha256"]:
        sys.exit("refusing: rounds/%d/inputs/GUIDE.md is unchanged since the last recorded guide" % r)
    row = {"utc": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()), "guide_sha256": gsha, "guide_version": x2d.GUIDE_VERSION_PREFIX + gsha[:12],
           "inputs_sha256": x2d.inputs_digest(x2d.inputs_dir(r, data)), "reason": reason.strip()}
    x2d.append_jsonl(revisions_path(r, data), [row])
    return row


def _diff(prev, new, what):
    return ["%s changed: %s -> %s" % (what, json.dumps(prev, sort_keys=True)[:200], json.dumps(new, sort_keys=True)[:200])] if prev != new else []


def check_envelope(prev_version, new_version):
    p, n = prev_version, new_version
    if n.get("round", 0) >= 8:
        v = []
        for k in sorted((set(p) | set(n)) - set(IGNORED_WHEN_FROZEN)):
            if p.get(k) != n.get(k):
                v.append("frozen round %s: %s differs from v_%s (rounds 8-10 use v_7 unchanged)" % (n.get("round"), k, p.get("round")))
        return v
    v = []
    v += _diff(p.get("catalogue_sha256"), n.get("catalogue_sha256"), "catalogue digest")
    for k in sorted(set(p.get("state_builder", {})) | set(n.get("state_builder", {}))):
        v += _diff(p.get("state_builder", {}).get(k), n.get("state_builder", {}).get(k), "state builder %s sha256" % k)
    v += _diff(p.get("thresholds_sha256"), n.get("thresholds_sha256"), "frozen thresholds.json (minimums, GO rule)")
    v += _diff(p.get("prepass_rounds_sha256"), n.get("prepass_rounds_sha256"), "round table prepass/rounds.json")
    v += _diff(p.get("model"), n.get("model"), "model pin")
    v += _diff(p.get("pr_cap"), n.get("pr_cap"), "PR cap")
    v += _diff(p.get("code_digests"), n.get("code_digests"), "code under test")
    pm, nm = p.get("question_meta", {}), n.get("question_meta", {})
    v += _diff(sorted(pm), sorted(nm), "question file set")
    for name in sorted(set(pm) & set(nm)):
        v += _diff(pm[name].get("kind"), nm[name].get("kind"), "%s kind" % name)
        v += _diff(pm[name].get("model"), nm[name].get("model"), "%s model" % name)
        v += _diff(pm[name].get("state_shape"), nm[name].get("state_shape"), "%s state_shape (outside the menu)" % name)
        rule = nm[name].get("option_order_rule")
        if not isinstance(rule, dict) or rule.get("rule") not in ORDER_RULES:
            v.append("%s option_order_rule %r not within %s" % (name, rule, list(ORDER_RULES)))
        v += _diff(p.get("question_structure", {}).get(name), n.get("question_structure", {}).get(name), "%s structure (anything but wording/criteria text)" % name)
    v += ["new " + x for x in settings_violations(n.get("settings"))]
    return v


def _load(path):
    if not os.path.exists(path):
        sys.exit("error: %s missing" % path)
    return json.load(open(path))


def main(argv=None):
    ap = argparse.ArgumentParser()
    ap.add_argument("--round", type=int, required=True)
    g = ap.add_mutually_exclusive_group()
    g.add_argument("--check", action="store_true")
    g.add_argument("--inputs", action="store_true")
    g.add_argument("--drift", action="store_true")
    g.add_argument("--clarify", metavar="REASON")
    a = ap.parse_args(argv)
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pin check failed:\n" + "\n".join(bad))
    if a.inputs:
        print("inputs of round %d: %s" % (a.round, ensure_inputs(a.round)))
        return
    if a.drift:
        d = version_drift(a.round)
        if d:
            sys.exit("refusing: round %d differs from its frozen version.json in: %s" % (a.round, ", ".join(d)))
        print("round %d: no drift from version.json" % a.round)
        return
    if a.clarify is not None:
        row = clarify(a.round, a.clarify)
        print("recorded guide revision %s of round %d" % (row["guide_version"], a.round))
        return
    if not a.check:
        v = freeze_version(a.round)
        print("froze version of round %d: %s" % (a.round, os.path.join(x2d.round_dir(a.round), "version.json")))
        return v
    base = FROZEN_BASE_ROUND if a.round >= 8 else a.round - 1
    prev = effective_version(base)
    np_ = os.path.join(x2d.round_dir(a.round), "version.json")
    new = _load(np_) if os.path.exists(np_) else build_version(a.round)
    viol = check_envelope(prev, new)
    for x in viol:
        print("VIOLATION: " + x, file=sys.stderr)
    if viol:
        sys.exit(1)
    print("round %d within the refinement envelope of round %d" % (a.round, base))


if __name__ == "__main__":
    main()
