#!/usr/bin/env python3
"""X2 implementation validation (AGENTS.md experiment rule 4). Synthetic recorded maps only: no model call, no real data.

Checks (all must pass; exit 1 otherwise):
  U  unit: Wilson intervals; decision logic (exact C1 on heterogeneous denominators, undefined spurious rate, non-contiguous pass set, precision
     guard, class gates, control gate (usable denominator, miss definition, minimum counts), replication rule); operating characteristics at n = 150/216/270.
  S1 planted effect (volatile concerns at Jaccard 0.33-0.45 share generic files): exact planted match and spurious counts, adopt t* = 0.75,
     chained survival and coverage diagnostics, provenance.   S1perm decision-level permutation null (S1 labels permuted -> fallback).
  S2 planted null (clean re-map).  S3 planted null (label noise independent of a Jaccard spread over 0.38-1.0): flat rate within binomial bounds, no
     gradient, planned verdict.  S4 survival boundary.  S5 unrelated maps (fallback, coverage ceiling).  S6 too few matches (precision guard).
  S7 unlabelled matches (label gate).  S8 failed maps (unmeasured transitions; remap and mutation class gates).  S9 invalid labels (control gate).
  S10 no prior areas, unparsable pairs, missing-field and reserved-key session answers, `--new-round` filtering; interrupted-run and concurrency guards.
  S11 skipped M2 mutation (no map, unmeasured T6).  S12 calibrated decision-level null/effect (Bernoulli labels at 3/5/7 %).
  R  replay cross-check: Go areas.Reconcile replay equals an independent Python reference (Go tie order: Jaccard, minted area ID, feature ID) on
     S1, S4, split-feature and cross-area-tie scenarios.  B blinding, side order against the key, control pairs.
  E  environment: collector environment equals Scout's (credential scrub checked against the Go function name for name; host auth context; BOTH provider
     timeouts on every call; the default equals client.go's); label sessions run through the same scrub.
  C  collector: stub Clawpatch, provenance and model probes, version pin, CLI-on-PATH preflight, append-only bookkeeping, retry from snapshot, timeout
     (SIGTERM then SIGKILL, detached child gone), abort record, stop rules, mutations; end to end pipeline.
Usage (repo root, toolchain active): .tools/venv/bin/python evaluation/x2-area-stability/validation/validate.py
"""
import glob
import hashlib
import json
import math
import os
import random
import re
import shutil
import signal
import subprocess
import sys
import time
import tempfile
from fractions import Fraction

HERE = os.path.dirname(os.path.abspath(__file__))
X2 = os.path.dirname(HERE)
ROOT = os.path.abspath(os.path.join(X2, "..", ".."))
sys.path.insert(0, X2)
import analysis as A  # noqa: E402
import common as C  # noqa: E402

FAILS = []
PY = sys.executable


def check(name, cond, detail=""):
    print(("PASS " if cond else "FAIL ") + name + ((" :: " + str(detail)) if detail != "" and not cond else ""))
    if not cond:
        FAILS.append(name)


def close(a, b, tol=1e-9):
    return a is not None and abs(a - b) <= tol


# ------------------------------------------------------------------ synthetic maps
def feat(fid, title, files, source="agent-mapper"):
    eps = [{"path": files[0], "symbol": "run_" + hashlib.sha1(title.encode()).hexdigest()[:6]}] if files else []
    return {"schemaVersion": 1, "featureId": fid, "title": title, "summary": "Handles " + title, "kind": "module", "source": source,
            "confidence": "medium", "entrypoints": eps, "contextFiles": [], "tests": [], "status": "mapped",
            "ownedFiles": [{"path": p, "reason": "synthetic"} for p in files]}


def write_map(root, arm, repo, mid, feats, rc=0, extra=None):
    d = C.map_dir(root, arm, repo, mid)
    os.makedirs(os.path.join(d, "features"))
    for f in feats:
        with open(os.path.join(d, "features", f["featureId"] + ".json"), "w") as fh:
            json.dump(f, fh)
    with open(os.path.join(d, "meta.json"), "w") as fh:
        json.dump({"schema": "x2-map-v2", "rc": rc, **(extra or {})}, fh)


EXTRA = {"F1": 0, "F2": 3, "F3": 3, "R1": 3, "R2": 0, "M1": 3, "M2": 0}  # extra files per map: regular J = 10/(10+e_prev+e_next) = 10/13 in every transition
VOL = [(3, 3), (5, 3), (3, 2), (4, 4), (5, 4), (7, 5), (6, 5), (5, 3)]  # (shared, unique): J = s/(s+2u), none on a 0.05 grid point
VOL_J = [s / (s + 2 * u) for s, u in VOL]  # .3333 .4545 .4286 .3333 .3846 .4118 .375 .4545
SPREAD_U = [0, 1, 2, 3, 4, 5, 6, 8]  # spread scenario: J = 10/(10+2u) = 1.0 .833 .714 .625 .556 .5 .455 .385 (none on a grid point except 1.0/.5)
SPREAD_J = [10 / (10 + 2 * u) for u in SPREAD_U]
TIE_STYLE = {"F1": "P", "F2": "N", "F3": "N", "R1": "N", "R2": "P", "M1": "N", "M2": "P"}  # tie scenario: every transition is P->N or N->P


def map_features(scn, arm, repo, mid):
    tag = "%s-%s-%s" % (arm, repo, mid)
    tok = hashlib.sha1(tag.encode()).hexdigest()[:8]  # opaque: makes pair signatures unique per (arm, repo, map) without naming them
    if scn["kind"] == "unrelated":
        return [feat("reg_%s_%d" % (tag, i), "Concern %d %s" % (i, tok), ["u/%s/%d/f%d" % (tag, i, k) for k in range(10)]) for i in range(scn["n_reg"])]
    out = []
    if scn["kind"] == "tie":
        for g in range(scn["n_reg"]):
            a = ["g%d/a%d" % (g, k) for k in range(4)]
            b = ["g%d/b%d" % (g, k) for k in range(4)]
            if TIE_STYLE[mid] == "P":  # two areas A, B (cross-area tie against both next features X and Y: J = 2/6 each)
                out.append(feat("tie_%s_%d_A" % (tag, g), "Tie %d A %s" % (g, tok), a))
                out.append(feat("tie_%s_%d_B" % (tag, g), "Tie %d B %s" % (g, tok), b))
            else:
                out.append(feat("tie_%s_%d_X" % (tag, g), "Tie %d X %s" % (g, tok), a[:2] + b[:2]))
                out.append(feat("tie_%s_%d_Y" % (tag, g), "Tie %d Y %s" % (g, tok), a[2:] + b[2:]))
        out.append(feat("heur_%s" % tag, "Heuristic group", ["heur/%s.py" % tag], source="python-source-group"))
        return out
    for i in range(scn["n_reg"]):
        e = 0 if scn["kind"] in ("clean", "spread") else EXTRA[mid]
        size = 6 if (scn["kind"] == "edge" and i >= scn["n_reg"] - scn["n_edge"]) else 10  # edge: 6-file areas have J = 6/9 = 0.667 successors
        base = ["base/%d/f%d" % (i, k) for k in range(size)]
        if scn["kind"] == "split" and mid != "F1":
            half = [base[:5], base[5:]]
            for h in (0, 1):
                out.append(feat("reg_%s_%d_%d" % (tag, i, h), "Concern %d%s %s" % (i, "ab"[h], tok), half[h]))
            continue
        files = base + ["ex/%s/%d/x%d" % (mid, i, k) for k in range(e)]
        if scn["kind"] == "spread":
            files = base + ["uq/%s/%d/u%d" % (mid, i, k) for k in range(SPREAD_U[i % len(SPREAD_U)])]
        out.append(feat("reg_%s_%d" % (tag, i), "Concern %d %s" % (i, tok), files))
    for j in range(scn.get("n_vol", 0)):
        s, u = VOL[j % len(VOL)]
        files = ["gen/%d/g%d" % (j, k) for k in range(s)] + ["vol/%s/%d/u%d" % (mid, j, k) for k in range(u)]
        out.append(feat("vol_%s_%d" % (tag, j), "Volatile concern %d %s" % (j, tok), files))
    out.append(feat("heur_%s" % tag, "Heuristic group", ["heur/%s.py" % tag], source="python-source-group"))
    return out


def gen_maps(root, scn, arms=("codex", "claude"), repos=("fastapi", "axum")):
    fails = scn.get("fail", [])
    for arm in arms:
        for repo in repos:
            for mid in C.MAP_IDS:
                write_map(root, arm, repo, mid, map_features(scn, arm, repo, mid), 1 if (arm, repo, mid) in fails else 0)


# ------------------------------------------------------------------ pipeline runners
_BIN = {}


def go_bin(tmp, name, pkg):
    if name not in _BIN:
        p = os.path.join(tmp, name)
        r = subprocess.run(["go", "build", "-o", p, pkg], cwd=ROOT, capture_output=True, text=True)
        if r.returncode != 0:
            print(r.stderr)
            sys.exit("go build failed")
        _BIN[name] = p
    return _BIN[name]


def run_replay(tmp, maps, trans, out):
    d = os.path.dirname(out)
    r = subprocess.run([go_bin(tmp, "replay-bin", "./evaluation/x2-area-stability/replay"), "-maps", maps, "-transitions", trans, "-grid",
                        ",".join(map(str, C.GRID_PCT)), "-out", out, "-chain-out", os.path.join(d, "chain.jsonl"),
                        "-meta-out", os.path.join(d, "replay-meta.json"), "-root", ROOT], capture_output=True, text=True)
    if r.returncode != 0:
        print(r.stderr)
        sys.exit("replay failed")


def truth_labels(keys, rule, ctl_keys=(), ctl_rule=None):
    """pair_id -> {A,B,C} stub session answers; every 10th pair has A wrong and C = truth (adjudication path). Control pairs get their expected
    label (or ctl_rule(control row))."""
    truth, by_pid = {}, {}
    for k in keys:
        by_pid.setdefault(k["pair_id"], []).append(k)
    labs = {pid: rule(rows[0]) for pid, rows in by_pid.items()}
    for c in ctl_keys:
        labs[c["pair_id"]] = (ctl_rule or (lambda c: c["expected"]))(c)
    for n, pid in enumerate(sorted(labs)):
        lab = labs[pid]
        if lab is None:
            truth[pid] = {"A": None, "B": None, "C": None}
        elif n % 10 == 0:
            truth[pid] = {"A": "cannot_tell" if lab != "cannot_tell" else "same_concern", "B": lab, "C": lab}
        else:
            truth[pid] = {"A": lab, "B": lab, "C": lab}
    return truth


def default_rule(k):
    return "different_concern" if k["next_feature_id"].startswith("vol_") else "same_concern"


def run_label(d, pairs, truth, extra=()):
    tf = os.path.join(d, "truth-%d.json" % random.randrange(10 ** 9))
    with open(tf, "w") as f:
        json.dump(truth, f)
    return subprocess.run([PY, os.path.join(X2, "label.py"), "--pairs", pairs, *extra], env=dict(os.environ, X2_LABEL_STUB=tf), capture_output=True, text=True)


def prepare(tmp, name, scn, arms=("codex", "claude"), meta_fix=None):
    d = os.path.join(tmp, name)
    maps, pairs = os.path.join(d, "data", "maps"), os.path.join(d, "pairs")
    os.makedirs(pairs)
    gen_maps(os.path.join(d, "data"), scn, arms)
    if meta_fix:
        meta_fix(os.path.join(d, "data"))
    trans = os.path.join(d, "transitions.json")
    C.write_transitions(trans, arms=list(arms))
    replay = os.path.join(d, "replay.jsonl")
    run_replay(tmp, maps, trans, replay)
    subprocess.run([PY, os.path.join(X2, "build_pairs.py"), "--data", os.path.join(d, "data"), "--replay", replay, "--out", pairs], check=True, capture_output=True)
    return d, replay, pairs


def run_scenario(tmp, name, scn, rule=default_rule, ctl_rule=None, meta_fix=None, arms=("codex", "claude")):
    d, replay, pairs = prepare(tmp, name, scn, arms, meta_fix)
    keys, ctl = C.read_jsonl(os.path.join(pairs, "key.jsonl")), C.read_jsonl(os.path.join(pairs, "control-key.jsonl"))
    r = run_label(d, pairs, truth_labels(keys, rule, ctl, ctl_rule))
    if r.returncode != 0:
        print(r.stdout, r.stderr)
        sys.exit("label.py failed in " + name)
    res = A.analyze(os.path.join(d, "data"), replay, pairs)
    return d, res, replay, pairs


def cell(res, arm, repo, stratum, pct):
    return res["arms"][arm]["tables"][repo][stratum][str(pct)]


# ------------------------------------------------------------------ python reference of areas.Reconcile (greedy best-first one-to-one)
def mint_ids(prev_agent):
    """internal/areas mintID for the prior areas Reconcile(nil, prev, ., Snapshot "prev") creates (features in feature-ID order)."""
    used, ids = set(), {}
    for fid, files in sorted(prev_agent):
        digest = hashlib.sha256(("prev" + "\0" + fid + "".join("\0" + p for p in files)).encode()).hexdigest()
        length = 16
        while True:
            length = min(length, len(digest))
            aid = "area:" + digest[:length]
            if aid not in used or length == len(digest):
                break
            length += 8
        used.add(aid)
        ids[fid] = aid
    return ids


def py_matches(prev, nxt, t, go_order=True):
    P = [(f["featureId"], C.owned_files(f)) for f in prev if f.get("source") == C.AGENT_SOURCE and C.owned_files(f)]
    N = [(f["featureId"], C.owned_files(f)) for f in nxt if f.get("source") == C.AGENT_SOURCE and C.owned_files(f)]
    aid = mint_ids(P)
    cands = []
    for pi, pf in P:
        for ni, nf in N:
            j = C.jaccard(pf, nf)
            if j >= t:
                cands.append((j, aid[pi] if go_order else pi, ni, pi))
    cands.sort(key=lambda x: (-x[0], x[1], x[2]))  # Go: Jaccard desc, area ID, feature ID
    ta, tf, out = set(), set(), set()
    for j, _a, ni, pi in cands:
        if pi in ta or ni in tf:
            continue
        ta.add(pi)
        tf.add(ni)
        out.add((pi, ni))
    return out


def crosscheck(d, go_order=True):
    rows = C.read_jsonl(os.path.join(d, "replay.jsonl"))
    data = os.path.join(d, "data")
    bad = 0
    maps_of = {t: (p, n) for t, _c, _s, p, n in C.TRANSITIONS}
    for r in rows:
        if r["status"] != "ok":
            continue
        pm, nm = maps_of[r["transition"]]
        want = py_matches(C.load_active_features(data, r["arm"], r["repo"], pm), C.load_active_features(data, r["arm"], r["repo"], nm), r["threshold"], go_order)
        got = {(m["prevFeatureId"], m["nextFeatureId"]) for m in r["matches"]}
        if want != got:
            bad += 1
    return bad, len(rows)


# ------------------------------------------------------------------ unit checks
def mk(surv_pairs=None, spur=0.0, n=160, surv=0.9, unlabelled=0):
    """Metrics of a decisive-stratum cell: six transitions of 20 prior areas (or the given (matched, prior) pairs), n matches of which
    round(spur*n) different and `unlabelled` without a label."""
    c = A.empty()
    pt = surv_pairs or [(round(surv * 20), 20)] * 6
    c["transitions"] = len(pt)
    c["per_transition"] = [{"transition": "T%d" % (i + 1), "matched": a, "prev_areas": b} for i, (a, b) in enumerate(pt)]
    c["prev_areas"], c["matched"] = max(sum(b for _, b in pt), n), n  # prev_areas only feeds the pooled (descriptive) interval
    lab = n - unlabelled
    c["different"] = round(spur * lab)
    c["same"] = lab - c["different"]
    c["never"] = unlabelled
    return A.metrics(c)


def unit_checks(tmp):
    lo, hi = C.wilson(0, 100)
    check("U wilson 0/100", close(lo, 0.0) and close(hi, 0.0369935, 1e-6), (lo, hi))
    lo, hi = C.wilson(1, 10)
    check("U wilson 1/10", close(lo, 0.0178754, 1e-6) and close(hi, 0.4041500, 1e-6), (lo, hi))
    lo, hi = C.wilson(50, 100)
    check("U wilson 50/100", close(lo, 0.4038315, 1e-6) and close(hi, 0.5961685, 1e-6), (lo, hi))
    check("U wilson n=0", C.wilson(0, 0)[0] != C.wilson(0, 0)[0])
    for k, n, want in ((7, 150, 0.093), (8, 150, 0.102), (0, 32, 0.107)):
        check("U wilson upper %d/%d ~ %.3f (prereg section 7)" % (k, n, want), close(C.wilson(k, n)[1], want, 6e-4), C.wilson(k, n))
    lo, hi = C.wilson(36, 45)
    check("U wilson 36/45 -> [0.66, 0.89]", close(lo, 0.66, 6e-3) and close(hi, 0.89, 6e-3), (lo, hi))

    CM = {"a": {"remap": 4, "mutation": 2}, "b": {"remap": 4, "mutation": 2}}

    def prod_for(rule):
        return {repo: {p: rule(p) for p in C.GRID_PCT} for repo in ("a", "b")}

    ok = A.decide(prod_for(lambda p: mk(surv=0.9)), CM)
    check("U decide all-pass -> adopt top of grid", ok["verdict"] == "adopt" and ok["threshold_pct"] == 90 and ok["pass_set_contiguous"], ok)
    nc = A.decide(prod_for(lambda p: mk(surv=0.9 if p in (30, 35, 45, 50) else 0.5)), CM)
    check("U decide non-contiguous pass set -> adopt max(pass), flagged non-contiguous", nc["verdict"] == "adopt" and nc["threshold_pct"] == 50 and not nc["pass_set_contiguous"], nc)
    fb = A.decide(prod_for(lambda p: mk(surv=0.9, spur=0.2)), CM)
    check("U decide spurious everywhere -> fallback", fb["verdict"] == "fallback", fb)
    one = {"a": {p: mk(surv=0.9) for p in C.GRID_PCT}, "b": {p: mk(surv=0.7) for p in C.GRID_PCT}}
    check("U decide one repo failing survival -> fallback", A.decide(one, CM)["verdict"] == "fallback")
    imp = A.decide(prod_for(lambda p: mk(surv=0.9, n=20)), CM)
    check("U decide precision guard -> inconclusive", imp["verdict"] == "inconclusive" and imp["imprecise_pct"] == C.GRID_PCT, imp)
    pt = A.decide(prod_for(lambda p: mk(surv=0.9, spur=0.06, n=1000)), CM)
    check("U decide point spurious 0.06 -> fallback", pt["verdict"] == "fallback", pt)
    edge = A.decide(prod_for(lambda p: mk(surv=0.8, spur=0.05, n=1000)), CM)
    check("U decide boundary survival 0.80 and spurious 0.05 pass", edge["verdict"] == "adopt", edge)
    # (c) exact rational C1 on heterogeneous denominators at the boundary
    hit = mk([(3, 4), (17, 20), (4, 5)] * 2, 0.0, 200)
    miss = mk([(3, 4), (16, 20), (4, 5)] * 2, 0.0, 200)
    check("U C1 exact: mean of 3/4, 17/20, 4/5 is exactly 4/5 -> pass; 3/4, 16/20, 4/5 -> fail",
          hit["survival_exact"] == "4/5" and A.classify_t({"a": hit}) == "pass" and A.classify_t({"a": miss}) == "fail", (hit["survival_exact"], miss["survival_exact"]))
    # (a) undefined spurious rate: evaluate every repo before returning
    undef = mk(surv=0.9, spur=0.0, n=10, unlabelled=10)
    bad_surv = mk(surv=0.5)
    check("U classify: undefined spurious (matches, none labelled) alone -> imprecise", undef["spurious"] is None and A.classify_t({"a": undef}) == "imprecise")
    check("U classify: undefined spurious in one repo and failing survival in the other -> fail (no early return)",
          A.classify_t({"a": undef, "b": bad_surv}) == "fail" and A.classify_t({"b": bad_surv, "a": undef}) == "fail")
    check("U classify: undefined spurious and a failing C2 elsewhere -> fail", A.classify_t({"a": undef, "b": mk(surv=0.9, spur=0.2)}) == "fail")
    # class gates and the gate named in the reason
    g = A.decide(prod_for(lambda p: mk(surv=0.9)), {"a": {"remap": 1, "mutation": 2}, "b": CM["b"]})
    check("U class gate: 1 measured remap -> inconclusive naming the class gate", g["verdict"] == "inconclusive" and g["tripped_gate"] == "class gate" and "remap" in g["reason"], g)
    g = A.decide(prod_for(lambda p: mk(surv=0.9)), {"a": CM["a"], "b": {"remap": 4, "mutation": 0}})
    check("U class gate: 0 measured mutation -> inconclusive", g["verdict"] == "inconclusive" and "mutation" in g["reason"], g)
    g = A.decide(prod_for(lambda p: mk(surv=0.9, unlabelled=20)), CM)
    check("U label gate: 12.5 % unlabelled at T_MIN -> inconclusive naming the label gate", g["verdict"] == "inconclusive" and g["tripped_gate"] == "label gate", g)
    g = A.decide(prod_for(lambda p: mk(surv=0.9, unlabelled=16)), CM)
    check("U label gate: exactly 10 % unlabelled passes", g["gates_ok"] and g["verdict"] == "adopt", g)
    g = A.decide(prod_for(lambda p: mk(surv=0.9)), CM, {"ok": False, "reason": "swapped controls: 3 of 10 labelled same_concern"})
    check("U control gate trips -> inconclusive naming the control gate", g["verdict"] == "inconclusive" and g["tripped_gate"] == "control gate", g)
    # replication
    dA = {"threshold_pct": 70}
    okp_, bad_b_, bad_a_ = ({"codex": {"consistent": a_}, "claude": {"consistent": b_}} for a_, b_ in ((True, True), (True, False), (False, True)))
    rep = lambda dB, pc=okp_: A.replication(dA, dB, pc)["status"]
    dB_ok = {"verdict": "adopt", "gates_ok": True, "per_threshold": {70: "pass"}}
    dB_no = {"verdict": "fallback", "gates_ok": True, "per_threshold": {70: "fail"}}
    check("U replication rule", rep(dB_ok) == "replicated"
          and rep({"verdict": "inconclusive", "gates_ok": False, "per_threshold": {70: "pass"}}) == "not replicated"
          and rep(dB_no) == "not replicated"
          and A.replication({"threshold_pct": None}, {"verdict": "fallback", "gates_ok": True, "per_threshold": {}}, okp_)["status"].startswith("not applicable"))
    check("U replication citability (section 4): a would-be-replicated arm B that is probe-inconsistent is `not citable`, a failing arm B stays `not replicated`, an inconsistent arm A carries no consequence (also `not applicable` when A did not adopt); missing probe data is inconsistent",
          rep(dB_ok, bad_b_) == "not citable (arm B probe-inconsistent)" and rep(dB_no, bad_b_) == "not replicated" and rep(dB_ok, bad_a_) == "not citable (arm A probe-inconsistent)"
          and A.replication({"threshold_pct": None}, dB_ok, bad_a_)["status"].startswith("not applicable") and rep(dB_ok, {}) == "not citable (arm A probe-inconsistent)"
          and rep(dB_no, bad_a_) == "not citable (arm A probe-inconsistent)"
          and A.replication(dA, dB_ok, bad_b_)["arm_B_citable"] is False and A.replication(dA, dB_ok, okp_)["arm_A_citable"] is True)
    # control gate (section 4 gate 3): USABLE3 denominator; a miss is any usable label other than the expected one; minimum usable counts; zero-kind failure
    D_, S_, O_, T_ = "different_concern", "same_concern", "overlapping_concern", "cannot_tell"

    def ctl_stats(name, sw, idl, n_real=250, kinds=("swapped", "identical")):
        d = os.path.join(tmp, "ctl-" + name)
        os.makedirs(d)
        with open(os.path.join(d, "key.jsonl"), "w") as f:
            for i in range(n_real):
                f.write(json.dumps({"pair_id": "r%d" % i, "jaccard": 0.3 + 0.7 * i / max(n_real, 1)}) + "\n")
        rows = [{"pair_id": "s%d" % i, "kind": "swapped", "expected": D_, "source_pair_id": "r%d" % i, "arm": "codex" if i % 2 else "claude", "repo": "x"} for i in range(len(sw))]
        rows += [{"pair_id": "i%d" % i, "kind": "identical", "expected": S_, "source_pair_id": "r%d" % i, "arm": "codex", "repo": "x"} for i in range(len(idl))]
        with open(os.path.join(d, "control-key.jsonl"), "w") as f:
            for c in rows:
                if c["kind"] in kinds:
                    f.write(json.dumps(c) + "\n")
        labs = {"s%d" % i: x for i, x in enumerate(sw)}
        labs.update({"i%d" % i: x for i, x in enumerate(idl)})
        return A.control_stats(d, labs)
    base_sw, base_id = [S_] * 4 + [D_] * 16, [S_] * 12
    cs = ctl_stats("base", base_sw, base_id)
    check("U controls: 4 of 20 swapped misses (20 %) and identical all right -> valid; wanted 20/10 for 250 real pairs",
          cs["ok"] and cs["kinds"]["swapped"]["wanted"] == 20 and cs["kinds"]["identical"]["wanted"] == 10 and cs["kinds"]["swapped"]["misses"] == 4, cs)
    cs = ctl_stats("over", [S_] * 5 + [D_] * 15, base_id)
    check("U controls: 5 of 20 swapped labelled same -> invalid, reason names swapped", not cs["ok"] and "swapped" in cs["reason"], cs["reason"])
    cs = ctl_stats("ovl", [S_] * 4 + [O_] + [D_] * 15, base_id)
    check("U controls: overlapping_concern on a swapped control is a miss (4 same + 1 overlapping of 20 -> invalid)", not cs["ok"] and cs["kinds"]["swapped"]["misses"] == 5, cs["reason"])
    cs = ctl_stats("ct", [T_] * 20, base_id)
    check("U controls: every swapped control cannot_tell -> invalid (no usable answer; the former gate passed here), cannot_tell reported separately",
          not cs["ok"] and cs["kinds"]["swapped"]["usable"] == 0 and cs["kinds"]["swapped"]["cannot_tell"] == 20 and "usable" in cs["reason"], cs["reason"])
    cs = ctl_stats("cov9", [D_] * 9 + [T_] * 11, base_id)
    check("U controls: 9 usable of 20 swapped (< minimum 10) -> invalid even with no miss", not cs["ok"] and cs["kinds"]["swapped"]["misses"] == 0, cs["reason"])
    cs = ctl_stats("cov10", [D_] * 10 + [T_] * 10, base_id)
    check("U controls: 10 usable of 20 swapped (minimum met) -> valid", cs["ok"], cs["reason"])
    cs = ctl_stats("idovl", base_sw, [S_] * 9 + [O_] * 3)
    check("U controls: identical controls labelled overlapping (3 of 12) -> invalid; different (3 of 12) likewise",
          not cs["ok"] and "identical" in cs["reason"] and not ctl_stats("iddiff", base_sw, [S_] * 9 + [D_] * 3)["ok"])
    cs = ctl_stats("nosw", [], base_id)
    check("U controls: zero swapped controls generated while 20 are wanted -> invalid (shortfall reported)", not cs["ok"] and cs["kinds"]["swapped"]["shortfall"] == 20 and "none generated" in cs["reason"], cs["reason"])
    cs = ctl_stats("none", [], [], n_real=0)
    check("U controls: no real pairs and no controls -> gate not applicable (valid)", cs["ok"] and not cs["kinds"]["swapped"]["applicable"], cs)
    cs = ctl_stats("band", base_sw, base_id)
    check("U controls: per-arm counts and the swapped miss rate by source-pair Jaccard band are reported",
          set(cs["per_arm"]) == {"codex", "claude"} and sum(b["n"] for b in cs["swapped_by_source_band_pct"].values()) == 20
          and sum(b["misses"] for b in cs["swapped_by_source_band_pct"].values()) == 4, (cs["per_arm"], cs["swapped_by_source_band_pct"]))
    # operating characteristics (prereg section 7): n = 150 (stress), 216 and 270 (expected size)
    oc = A.operating_characteristics()
    pb = {n: {k: v["p_both_repos"] for k, v in o["table"].items()} for n, o in oc["by_n"].items()}
    want = {"150": (7, (0.84, 0.56, 0.27, 0.03, 0.0002)), "216": (10, (0.878, 0.564, 0.232, 0.011, 0.0)), "270": (13, (0.931, 0.646, 0.267, 0.009, 0.0))}
    check("U operating characteristics at n = 150 / 216 / 270 (max spurious 7 / 10 / 13; P(adopt at one threshold, both repos) at 3, 4, 5, 7, 10 %)",
          all(oc["by_n"][n]["max_spurious_count"] == want[n][0] and all(close(pb[n][p], w, 0.005) for p, w in zip(("0.03", "0.04", "0.05", "0.07", "0.1"), want[n][1])) for n in want)
          and oc["c3_stops_binding_at_n"] == 140, (pb, oc["c3_stops_binding_at_n"]))
    print("     operating characteristics computed:", {n: {k: round(v, 4) for k, v in p.items()} for n, p in pb.items()}, "c3 binding n", oc["c3_stops_binding_at_n"])
    # build_state: canonical side order, per-occurrence left_is_a (round-2 N13a)
    import build_pairs as BP
    ra, rb = feat("a", "Alpha concern", ["x/1", "x/2", "x/3"]), feat("b", "Beta concern", ["x/2", "x/3", "x/4"])
    pid1, st1, la1 = BP.build_state(ra, rb)
    pid2, st2, la2 = BP.build_state(rb, ra)
    check("U build_state: the same pair in either orientation has one id and one state, and left_is_a is right per occurrence (and so flips)",
          pid1 == pid2 and st1 == st2 and (st1["left"]["title"] == "Alpha concern") == la1 and (st2["left"]["title"] == "Beta concern") == la2 and la1 != la2, (la1, la2))
    # joint leave-one-transition-out (round-3 issue 5): X passes at every threshold; Y's mean survival above 0.60 is 5/6 (28, 32, 40 of 40) and falls to 3/4 when T3 is held out
    lrows, lkey, llabels = [], [], {}
    for repo_, trs in {"X": {"T1": 40, "T2": 40, "T3": 40}, "Y": {"T1": 28, "T2": 32, "T3": 40}}.items():
        for tr_, full in trs.items():
            for i in range(40):
                pid_ = "%s-%s-%d" % (repo_, tr_, i)
                lkey.append({"pair_id": pid_, "arm": "codex", "repo": repo_, "transition": tr_, "prev_feature_id": "p%d" % i, "next_feature_id": "n%d" % i})
                llabels[pid_] = "same_concern"
            for pct_ in C.GRID_PCT:
                n_ = 40 if pct_ <= 60 else full
                lrows.append({"arm": "codex", "repo": repo_, "transition": tr_, "thresholdPct": pct_, "class": "remap", "sub": "fresh", "status": "ok", "prevAreas": 40,
                              "matched": n_, "created": 0, "retired": 0, "kept": 40 - n_,
                              "matches": [{"prevFeatureId": "p%d" % i, "nextFeatureId": "n%d" % i, "jaccard": 0.9} for i in range(n_)]})
    ltab = A.build_tables(lrows, lkey, llabels)["codex"]
    lprod = {repo_: {p: A.metrics(c) for p, c in ltab[repo_]["decisive"].items()} for repo_ in ("X", "Y")}
    lo = {x["held_out"]: x for x in A.loto(lrows, lkey, llabels, "codex", lprod)}
    check("U joint LOTO: the adopted t* over all transitions is 0.90 (both repositories pass everywhere)", all(A.classify_t({r_: lprod[r_][p] for r_ in lprod}) == "pass" for p in C.GRID_PCT))
    check("U joint LOTO: holding out T3 drops Y below 4/5 above 0.60, so the JOINT t* is 0.60 although X alone would keep 0.90 (the per-repository rule of round 2 would report 0.90); T1 and T2 keep 0.90",
          [lo[t]["t_star_pct"] for t in ("T1", "T2", "T3")] == [90, 90, 60], {t: lo[t]["t_star_pct"] for t in lo})
    check("U joint LOTO: held-out survival, matched, labelled and different counts per repository at the joint t* (Y T1 28/40 = 0.7 at 0.90; Y and X T3 40/40 at 0.60)",
          lo["T1"]["per_repo"]["Y"] == {"measured": True, "held_out_survival": 0.7, "matched": 28, "prev_areas": 40, "labelled": 28, "different": 0}
          and lo["T3"]["per_repo"]["Y"]["held_out_survival"] == 1.0 and lo["T3"]["per_repo"]["Y"]["labelled"] == 40 and lo["T3"]["per_repo"]["X"]["held_out_survival"] == 1.0, lo)
    # the model probe's claude environment is a port of claudeEnv(host) (round-3 issue 1)
    base_ = {"PATH": "/p", "Path": "/q", "HOME": "/h", "USERPROFILE": "/u", "CLAUDE_CONFIG_DIR": "/c", "ANTHROPIC_MODEL": "m", "X2_PLAIN": "1", "GOOGLE_CLOUD_PROJECT": "g",
             "AWS_CONFIG_FILE": "/a", "AWS_PROFILE": "pf", "CLAUDE_CODE_OAUTH_TOKEN": "t", "CLAUDE_CODE_USE_VERTEX": "", "CLAWPATCH_CLAUDE_BIN": "/b", "ANTHROPIC_API_KEY": "k",
             "USER": "op", "LOGNAME": "op2", "SHELL": "/bin/sh"}
    penv = C.claude_provider_env(base_, "/tmp/x")
    want_names = {"PATH", "HOME", "USERPROFILE", "CLAUDE_CONFIG_DIR", "XDG_CONFIG_HOME", "XDG_CACHE_HOME", "XDG_DATA_HOME", "TMPDIR", "TEMP", "TMP",
                  "CLAUDE_CODE_SUBPROCESS_ENV_SCRUB", "GOOGLE_CLOUD_PROJECT", "AWS_CONFIG_FILE", "AWS_PROFILE", "CLAUDE_CODE_OAUTH_TOKEN", "ANTHROPIC_API_KEY", "USER", "LOGNAME"}
    check("U claude provider environment port (as patched, scout-review-sanitize-v3): exactly PATH-like, HOME, USERPROFILE, CLAUDE_CONFIG_DIR, USER, LOGNAME, fresh XDG/TMP dirs, SUBPROCESS_ENV_SCRUB=1, the auth keys present and non-empty, AWS_PROFILE only with an AWS file, the OAuth token; nothing else (no ANTHROPIC_MODEL, no SHELL, no X2_*, no CLAWPATCH_CLAUDE_BIN)",
          set(penv) == want_names and penv["CLAUDE_CODE_SUBPROCESS_ENV_SCRUB"] == "1" and penv["TMP"] == penv["TMPDIR"] == penv["TEMP"] == "/tmp/x" and penv["XDG_CONFIG_HOME"] == "/tmp/x/xdg-config"
          and "AWS_PROFILE" not in C.claude_provider_env({"HOME": "/h", "AWS_PROFILE": "p"}, "/t") and C.claude_executable({"CLAWPATCH_CLAUDE_BIN": " /b "}) == "/b" and C.claude_executable({}) == "claude"
          and penv["USER"] == "op" and penv["LOGNAME"] == "op2" and "USER" not in C.claude_provider_env({"HOME": "/h", "USER": ""}, "/t"), sorted(set(penv) ^ want_names))
    pe = lambda d_: sorted(k for k in C.claude_provider_env(d_, "/t") if k.lower() == "path")
    check("U claude provider environment port, PATH: only the FIRST case-insensitive `path` key is considered (claude.ts copyPathEnv returns after it), in both insertion orders, and an empty first match copies nothing and stops the search",
          pe({"PATH": "/p", "Path": "/q"}) == ["PATH"] and pe({"Path": "/q", "PATH": "/p"}) == ["Path"] and pe({"PATH": "", "Path": "/q"}) == [] and pe({"Path": "", "PATH": "/p"}) == []
          and pe({"HOME": "/h", "path": "/z", "PATH": "/p"}) == ["path"] and pe({"HOME": "/h"}) == [] and C.claude_provider_env({"Path": "/q", "PATH": "/p"}, "/t")["Path"] == "/q",
          [pe({"PATH": "/p", "Path": "/q"}), pe({"Path": "/q", "PATH": "/p"}), pe({"PATH": "", "Path": "/q"})])
    # amendment 1: the drift guards read the source AS BUILT (pinned revision + scripts/clawpatch-safe.patch), not the unpatched submodule checkout
    pt = C.patched_sources(["src/providers/claude.ts", "src/providers/codex.ts", "src/config.ts", "src/prompt.ts"])
    claude_ts, codex_ts = pt["src/providers/claude.ts"], pt["src/providers/codex.ts"]
    mk_ts = re.search(r"CLAUDE_AUTH_ENV_KEYS = \[(.*?)\] as const", claude_ts, re.S)
    check("U drift guard (providers/claude.ts, codex.ts): the ported auth-key list equals CLAUDE_AUTH_ENV_KEYS; the flags and environment names the probes mirror are present; neither provider passes --setting-sources, --ignore-user-config or --ignore-rules",
          bool(mk_ts) and re.findall(r'"([A-Z0-9_]+)"', mk_ts.group(1)) == C.CLAUDE_AUTH_ENV_KEYS
          and all('"%s"' % f in claude_ts for f in ("--safe-mode", "--strict-mcp-config", "--disable-slash-commands", "--no-chrome", "--no-session-persistence", "--tools", "--output-format"))
          and all(n in claude_ts for n in ("CLAUDE_CODE_SUBPROCESS_ENV_SCRUB", "CLAWPATCH_CLAUDE_BIN", "CLAUDE_CODE_OAUTH_TOKEN", "CLAUDE_CONFIG_DIR"))
          and all('"%s"' % f in codex_ts for f in ("exec", "--cd", "--sandbox", "--skip-git-repo-check"))
          and not any(s in claude_ts + codex_ts for s in ("setting-sources", "ignore-user-config", "ignore-rules"))
          and "replaceEnv: true" in claude_ts and bool(re.search(r"function copyPathEnv\(.*?copyEnv\(target, key\);\s*return;", claude_ts, re.S)), bool(mk_ts))
    cfg_ts = pt["src/config.ts"]
    check("U drift guard (config.ts): Clawpatch reads CLAWPATCH_CONFIG, which the collector refuses (REFUSED_MODEL_ENV) because it can set model, reasoning effort and codexConfig for the maps only",
          "CLAWPATCH_CONFIG" in cfg_ts and "CLAWPATCH_CONFIG" in C.REFUSED_MODEL_ENV and set(C.REFUSED_MODEL_ENV) == {"CLAWPATCH_MODEL", "CLAWPATCH_REASONING_EFFORT", "CLAWPATCH_PROVIDER", "CLAWPATCH_CONFIG"})
    check("U drift guard (codex.ts): Clawpatch honours CLAWPATCH_CODEX_SANDBOX (bypass/none or --sandbox <value>), which the collector refuses (REFUSED_SANDBOX_ENV) so no map can write to the checkout",
          "CLAWPATCH_CODEX_SANDBOX" in codex_ts and "--dangerously-bypass-approvals-and-sandbox" in codex_ts and C.REFUSED_SANDBOX_ENV == ("CLAWPATCH_CODEX_SANDBOX",))
    sub_claude = open(os.path.join(ROOT, "third_party", "clawpatch", "src", "providers", "claude.ts")).read()
    check("U drift guard (patched claude.ts host branch): the patched source copies USER and LOGNAME right after CLAUDE_CONFIG_DIR (the port's additions), the source read is the patched one (differs from the unpatched submodule file), and the patched prompt.ts carries the sanitizer",
          bool(re.search(r'copyEnv\(env, "CLAUDE_CONFIG_DIR"\);\s*copyEnv\(env, "USER"\);\s*copyEnv\(env, "LOGNAME"\);\s*\} else \{', claude_ts))
          and claude_ts != sub_claude and 'copyEnv(env, "USER")' not in sub_claude and C.SANITIZER_TOKEN in pt["src/prompt.ts"] and C.SANITIZER_TOKEN not in open(os.path.join(ROOT, "third_party", "clawpatch", "src", "prompt.ts")).read())
    # amendment 1: the Clawpatch build must be the review-safety build (pins read from build-clawpatch.sh, cross-checked with review.go)
    pv, ps = C.review_safety_pins()
    check("U build-safety pins: read from scripts/build-clawpatch.sh, equal to internal/clawpatch/review.go and to the digest of scripts/clawpatch-safe.patch",
          re.fullmatch(r"scout-review-sanitize-v\d+", pv) and re.fullmatch(r"[0-9a-f]{64}", ps) and ps == C.sha256_file(os.path.join(ROOT, "scripts", "clawpatch-safe.patch")), (pv, ps))

    def fake_build(name, marker, dist_text, files=None):
        b = os.path.join(tmp, "build-" + name)
        os.makedirs(os.path.join(b, "dist"))
        if marker is not None:
            with open(os.path.join(b, C.BUILD_MARKER_FILE), "w") as f:
                f.write(marker)
        for fn, text in (files or {"prompt.js": dist_text}).items():
            with open(os.path.join(b, "dist", fn), "w") as f:
                f.write(text)
        return os.path.join(b, "dist", "cli.js")
    exp_marker = "%s %s\n" % (pv, ps)
    san = "const x = '[%s]';" % C.SANITIZER_TOKEN
    cases = {"patched": fake_build("ok", exp_marker, san), "patched, sanitizer in a second dist file": fake_build("ok2", exp_marker, "", {"cli.js": "//", "sub.js": san}),
             "patched, sanitizer in the entry file": fake_build("ok3", exp_marker, "", {"cli.js": san}),
             "no marker": fake_build("nomark", None, san), "old patch version": fake_build("oldv", "scout-review-sanitize-v2 f39d73da\n", san),
             "right version, wrong sha": fake_build("wrongsha", "%s %s\n" % (pv, "0" * 64), san), "marker but dist unpatched": fake_build("nosan", exp_marker, "const x = 1;")}
    got = {k: C.build_safety_check(p) for k, p in cases.items()}
    check("U build-safety check: passes a build whose marker equals the pins and whose dist/ holds the sanitizer (in the entry file cli.js, and in a non-entry file); refuses no marker, an old patch version, a wrong sha, and a marker over an unpatched dist/; the reason names the cause",
          got["patched"]["ok"] and got["patched, sanitizer in a second dist file"]["ok"] and got["patched, sanitizer in the entry file"]["ok"] and got["patched"]["marker"] == exp_marker.strip()
          and all(not got[k]["ok"] and got[k]["reason"] for k in ("no marker", "old patch version", "right version, wrong sha", "marker but dist unpatched"))
          and C.BUILD_MARKER_FILE in got["no marker"]["reason"] and got["old patch version"]["marker_ok"] is False and got["marker but dist unpatched"]["marker_ok"] is True
          and got["marker but dist unpatched"]["sanitizer_in_dist"] is False, got)
    ok_p = lambda m: {"model": m, "error": None}

    def tvf(before, after, v0=None, v1="same"):
        v0 = v0 or {"codex": "c1"}
        return {"model_probe_before": before, "model_probe_after": after, "cli_versions_after": v0 if v1 == "same" else v1, **v0}
    ar = [("codex", "fastapi"), ("codex", "axum"), ("claude", "fastapi")]
    tvs = {"tool-versions-codex-fastapi.json": tvf(ok_p("m1"), ok_p("m1")), "tool-versions-codex-axum.json": tvf(ok_p("m1"), ok_p("m1")),
           "tool-versions-claude-fastapi.json": tvf(ok_p("n1"), ok_p("n1"), {"claude": "v1"})}
    pc0 = A.probe_consistency(tvs, ar)
    check("U probe consistency: all probes taken without error naming one model (and unchanged CLI versions) is CONSISTENT per arm", pc0["codex"]["consistent"] and pc0["claude"]["consistent"]
          and pc0["codex"]["models"] == ["m1"] and pc0["codex"]["cli_versions_changed"] == {} and pc0["codex"]["reasons"] == [], pc0)
    cases = {"fastapi and axum name different models": ({"tool-versions-codex-axum.json": tvf(ok_p("m2"), ok_p("m2"))}, "probes name different models"),
             "after-probe names another model": ({"tool-versions-codex-axum.json": tvf(ok_p("m1"), ok_p("m2"))}, "probes name different models"),
             "after-probe failed": ({"tool-versions-codex-axum.json": tvf(ok_p("m1"), {"model": None, "error": "rc=1"})}, "axum: after probe failed (rc=1)"),
             "before-probe names no model": ({"tool-versions-codex-axum.json": tvf({"model": None, "error": None}, ok_p("m1"))}, "axum: before probe failed (no model)"),
             "after-probe not taken (abort)": ({"tool-versions-codex-axum.json": tvf(ok_p("m1"), "not taken: aborted")}, "axum: after probe not taken (not taken: aborted)"),
             "versions file missing": ({"tool-versions-codex-axum.json": None}, "axum: no tool-versions file")}
    bad = []
    for nm, (patch, reason) in cases.items():
        t_ = {k: v for k, v in {**tvs, **patch}.items() if v is not None}
        pc_ = A.probe_consistency(t_, ar)
        if pc_["codex"]["consistent"] or reason not in pc_["codex"]["reasons"] or not pc_["claude"]["consistent"]:
            bad.append((nm, pc_["codex"]["reasons"]))
    check("U probe consistency: each of the six disagreements (repository split, rolled after-probe, failed probe, no model, probe not taken, missing file) flags only the arm concerned with its reason", not bad, bad)
    pcv = A.probe_consistency({**tvs, "tool-versions-codex-axum.json": tvf(ok_p("m1"), ok_p("m1"), {"codex": "c1"}, {"codex": "c2"})}, ar)
    check("U probe consistency: a codex CLI-version change between the probes is listed with its distinct values but does not enter the flag", pcv["codex"]["consistent"] and pcv["codex"]["cli_versions_changed"] == {"codex": ["c1", "c2"]}, pcv["codex"])
    pcx = A.probe_consistency({**tvs, "tool-versions-codex-axum.json": tvf(ok_p("m1"), ok_p("m1"), {"codex": "c1", "claude": "v1"}, {"codex": "c1", "claude": "v2"}),
                               "tool-versions-claude-fastapi.json": tvf(ok_p("n1"), ok_p("n1"), {"claude": "v1", "codex": "c1"}, {"claude": "v1", "codex": "c9"})}, ar)
    check("U probe consistency: only the arm's own CLI is compared (a claude change during the codex arm and a codex change during the claude arm are not listed)",
          pcx["codex"]["consistent"] and pcx["codex"]["cli_versions_changed"] == {} and pcx["claude"]["cli_versions_changed"] == {}, (pcx["codex"]["cli_versions_changed"], pcx["claude"]["cli_versions_changed"]))
    pcg = A.probe_consistency({**tvs, "tool-versions-codex-fastapi.json": tvf(ok_p("m1"), ok_p("m1"), {"codex": "c1"}), "tool-versions-codex-axum.json": tvf(ok_p("m1"), ok_p("m1"), {"codex": "c2"})}, ar)
    pcb = A.probe_consistency({**tvs, "tool-versions-claude-fastapi.json": tvf(ok_p("n1"), ok_p("n1"), {"claude": "v1", "claude_executable_version": "e1"}, {"claude": "v1", "claude_executable_version": "e2"})}, ar)
    check("U probe consistency: a codex change between the repositories (after fastapi's probe, before axum's start) is listed although no single repository sees it; the claude arm compares `claude` and `claude_executable_version`",
          pcg["codex"]["consistent"] and pcg["codex"]["cli_versions_changed"] == {"codex": ["c1", "c2"]} and pcg["claude"]["cli_versions_changed"] == {}
          and pcb["claude"]["cli_versions_changed"] == {"claude_executable_version": ["e1", "e2"]}, (pcg["codex"]["cli_versions_changed"], pcb["claude"]["cli_versions_changed"]))
    pcw = A.probe_consistency({**tvs, "tool-versions-codex-axum.json": {**tvf(ok_p("m1"), ok_p("m1")), "workspace_config_presence": {".codex": True, ".claude/settings.json": False, "clawpatch.config.json": True}}}, ar)
    check("U probe consistency: a workspace-level configuration file present at the run makes only that arm INCONSISTENT, naming the files (presence entries that are false are not named)",
          not pcw["codex"]["consistent"] and "axum: workspace config present: .codex, clawpatch.config.json" in pcw["codex"]["reasons"] and pcw["claude"]["consistent"]
          and pcw["codex"]["per_repo"]["axum"]["workspace_config_presence"][".codex"] is True, pcw["codex"]["reasons"])
    # round-5 issue 2b: the exclusive open sets `_created` before the dump, so an interrupt during the dump leaves a file the abort path can rewrite
    import types
    from unittest import mock
    import collect as CO
    vdir = tempfile.mkdtemp(prefix="x2-versions-", dir=tmp)
    va, vv = types.SimpleNamespace(data=vdir, arm="codex", repo="tiny"), {"a": 1}
    with mock.patch.object(json, "dump", side_effect=KeyboardInterrupt("signal 15")):
        try:
            CO.write_versions(va, vv)
        except KeyboardInterrupt:
            pass
    vpath = os.path.join(vdir, "tool-versions-codex-tiny.json")
    partial_ok = os.path.exists(vpath) and vv.get("_created") is True
    vv["b"] = 2
    CO.write_versions(va, vv)  # what the abort path does: rewrite, no FileExistsError
    check("U write_versions: an interrupt during the first dump leaves `_created` set, and the rewrite from the abort path repairs the file instead of failing on the exclusive open",
          partial_ok and json.load(open(vpath)) == {"a": 1, "b": 2}, (partial_ok, open(vpath).read()))


# ------------------------------------------------------------------ main
def main():
    tmp = tempfile.mkdtemp(prefix="x2-validate-")
    print("work dir", tmp)
    unit_checks(tmp)

    # ---- S1 planted effect (all six transitions decisive)
    s1 = {"kind": "planted", "n_reg": 32, "n_vol": 8}
    d1, r1, rp1, p1 = run_scenario(tmp, "S1", s1)
    j = sorted(VOL_J)
    for repo in ("fastapi", "axum"):
        for pct in (30, 35, 40, 45, 50, 75, 80):
            t = pct / 100
            vol = sum(1 for x in j if x >= t)
            reg = 6 * 32 if pct <= 75 else 0
            c = cell(r1, "codex", repo, "decisive", pct)
            exp_match, exp_diff = reg + 6 * vol, 6 * vol
            check("S1 %s t=%.2f matched/different" % (repo, t), c["matched"] == exp_match and c["different"] == exp_diff, (c["matched"], c["different"], exp_match, exp_diff))
            if pct == 30:
                check("S1 %s t=0.30 spurious = 48/240 planted" % repo, close(c["spurious"], 48 / 240), c["spurious"])
            if pct == 50:
                check("S1 %s t=0.50 no spurious and survival 32/40" % repo, c["different"] == 0 and c["survival_exact"] == "4/5", (c["different"], c["survival_exact"]))
    dec = r1["arms"]["codex"]["decision"]
    check("S1 verdict adopt threshold 75 (planted)", dec["verdict"] == "adopt" and dec["threshold_pct"] == 75, dec)
    check("S1 feasible set 50..75, contiguous", dec["feasible_pct"] == list(range(50, 76, 5)) and dec["pass_set_contiguous"], dec["feasible_pct"])
    check("S1 low thresholds fail on spurious merges", all(dec["per_threshold"][p] == "fail" for p in (30, 35, 40, 45)), dec["per_threshold"])
    g = r1["arms"]["codex"]["gradient_exploratory"]
    check("S1 Jaccard gradient strongly negative", g["r"] is not None and g["r"] < -0.5, g)
    rp = r1["replication"]
    okp1 = {"codex": {"consistent": True}, "claude": {"consistent": True}}
    dA1, dB1 = r1["arms"]["codex"]["decision"], r1["arms"]["claude"]["decision"]
    rp_ok = A.replication(dA1, dB1, okp1)
    check("S1 replication: arm B replicated at t* (both arms probe-consistent), non-gating", rp_ok["status"] == "replicated" and rp_ok["arm_B_status_at_threshold"] == "pass" and rp_ok["gating"] is False, rp_ok)
    rp_b = A.replication(dA1, dB1, {"codex": {"consistent": True}, "claude": {"consistent": False}})
    rp_a = A.replication(dA1, dB1, {"codex": {"consistent": False}, "claude": {"consistent": True}})
    text1 = A.render(dict(r1, replication=rp_b))
    check("S1 citability (section 4): the decisions are unchanged (adopt 75), but with the claude probes rolled the replication status is `not citable (arm B probe-inconsistent)`, not `replicated`, and the rendered replication line names the reason; with the codex probes inconsistent it is `not citable (arm A ...)`; the scenario's own analysis has no versions files, so both arms are not citable",
          dA1["verdict"] == "adopt" and dA1["threshold_pct"] == 75 and rp_b["status"] == "not citable (arm B probe-inconsistent)" and rp_b["arm_B_status_at_threshold"] == "pass"
          and rp_a["status"] == "not citable (arm A probe-inconsistent)" and "not citable (arm B probe-inconsistent)" in [l_ for l_ in text1.split("\n") if l_.startswith("replication:")][0]
          and dA1["citable"] is False and dB1["citable"] is False and r1["replication"]["status"] == "not citable (arm A probe-inconsistent)"
          and "[citable for a design change: NO (arm probe-inconsistent)]" in text1, (rp_b, r1["replication"]["status"]))

    # ---- round-7 issue 1: analyze -> replication wiring with real tool-versions files (both arms), not direct replication() calls
    def plant_versions(data, claude_axum_after="n1"):
        for arm_, repo_ in (("codex", "fastapi"), ("codex", "axum"), ("claude", "fastapi"), ("claude", "axum")):
            m_, cli_ = ("m1", {"codex": "c1"}) if arm_ == "codex" else ("n1", {"claude": "v1", "claude_executable_version": "e1"})
            after_m = claude_axum_after if (arm_, repo_) == ("claude", "axum") else m_
            with open(os.path.join(data, "tool-versions-%s-%s.json" % (arm_, repo_)), "w") as f_:
                json.dump({"model_probe_before": {"model": m_, "error": None}, "model_probe_after": {"model": after_m, "error": None}, "cli_versions_after": dict(cli_), **cli_,
                           "workspace_config_presence": {".claude/settings.json": False, ".codex": False, "clawpatch.config.json": False}}, f_)
    dv_ok = os.path.join(d1, "data-versions-ok")
    shutil.copytree(os.path.join(d1, "data"), dv_ok)
    plant_versions(dv_ok)
    rv = A.analyze(dv_ok, rp1, p1)
    check("S1 analyze -> replication with probe-consistent tool-versions files for both arms and repositories: both decisions are citable and the replication status is `replicated` with both arms citable",
          rv["probe_consistency"]["codex"]["consistent"] and rv["probe_consistency"]["claude"]["consistent"]
          and rv["arms"]["codex"]["decision"]["citable"] is True and rv["arms"]["claude"]["decision"]["citable"] is True
          and rv["replication"]["status"] == "replicated" and rv["replication"]["arm_A_citable"] is True and rv["replication"]["arm_B_citable"] is True
          and rv["replication"]["arm_B_verdict"] == rv["arms"]["claude"]["decision"]["verdict"] == "adopt" and rv["replication"]["arm_B_threshold_pct"] == 75,
          (rv["replication"], rv["probe_consistency"]))
    dv_roll = os.path.join(d1, "data-versions-roll")
    shutil.copytree(dv_ok, dv_roll)
    plant_versions(dv_roll, claude_axum_after="n2")
    rr_ = A.analyze(dv_roll, rp1, p1)
    check("S1 analyze -> replication with the claude axum after-probe rolled: the claude decision is not citable and the replication status is `not citable (arm B probe-inconsistent)`, the codex decision stays citable",
          rr_["arms"]["codex"]["decision"]["citable"] is True and rr_["arms"]["claude"]["decision"]["citable"] is False
          and rr_["replication"]["status"] == "not citable (arm B probe-inconsistent)" and rr_["replication"]["arm_A_citable"] is True and rr_["replication"]["arm_B_citable"] is False,
          (rr_["replication"], rr_["probe_consistency"]["claude"]["reasons"]))
    d1b, r1b_, rp1b, p1b = run_scenario(tmp, "S1b", s1, rule=lambda k: "different_concern" if k["arm"] == "claude" else default_rule(k))
    plant_versions(os.path.join(d1b, "data"))
    rb_ = A.analyze(os.path.join(d1b, "data"), rp1b, p1b)
    check("S1b analyze -> replication, both arms probe-consistent, only the claude arm's real pairs labelled different_concern: arm A still adopts 75, arm B's decision differs, and the status is `not replicated` with arm_B_verdict taken from the claude decision",
          rb_["arms"]["codex"]["decision"]["verdict"] == "adopt" and rb_["arms"]["codex"]["decision"]["threshold_pct"] == 75
          and rb_["arms"]["claude"]["decision"]["verdict"] != rb_["arms"]["codex"]["decision"]["verdict"]
          and rb_["replication"]["status"] == "not replicated" and rb_["replication"]["arm_B_verdict"] == rb_["arms"]["claude"]["decision"]["verdict"]
          and rb_["replication"]["arm_B_threshold_pct"] == rb_["arms"]["claude"]["decision"].get("threshold_pct") and rb_["replication"]["arm_B_citable"] is True,
          (rb_["replication"], rb_["arms"]["claude"]["decision"]["verdict"]))
    fr, ss, rm, mu = (cell(r1, "codex", "fastapi", s, 50) for s in ("fresh", "same_state", "remap", "mutation"))
    check("S1 sub-strata and classes reported (fresh 2, same_state 2, remap 4, mutation 2)",
          (fr["transitions"], ss["transitions"], rm["transitions"], mu["transitions"]) == (2, 2, 4, 2) and fr["prev_areas"] == 80, (fr["transitions"], ss["transitions"], rm["transitions"], mu["transitions"]))
    check("S1 per-transition survival keeps transition IDs", [x["transition"] for x in cell(r1, "codex", "fastapi", "decisive", 50)["survival_per_transition"]] == ["T%d" % i for i in range(1, 7)])
    la = r1["label_agreement"]
    check("S1 adjudication path exercised; agreement counts real pairs only (controls reported separately)", la["adjudicated"] > 0 and la["both_parsed"] == la["pairs"] and la["cohen_kappa"] is not None and la["pairs"] == len({k["pair_id"] for k in C.read_jsonl(os.path.join(p1, "key.jsonl"))}), la)
    bad, n = crosscheck(d1)
    check("R S1 replay equals Python reference at every threshold (%d rows)" % n, bad == 0, bad)
    # chained survival (label-free), coverage diagnostics, provenance
    ch = {s["step"]: s for s in r1["arms"]["codex"]["chain"]["fastapi"]["50"]}
    check("S1 chained fold at 0.50: T3 40 active prior, 32 matched, 8 created, 8 kept, 0 retired; 32 of 40 F1 areas matched at every step through T6",
          (ch["T3"]["activePrior"], ch["T3"]["matched"], ch["T3"]["created"], ch["T3"]["kept"], ch["T3"]["retired"]) == (40, 32, 8, 8, 0) and ch["T6"]["f1MatchedAllSoFar"] == 32 and ch["T6"]["f1Areas"] == 40
          and r1["arms"]["codex"]["chain"]["fastapi"]["80"][-1]["f1MatchedAllSoFar"] == 0, ch)
    cov = r1["arms"]["codex"]["coverage"]["fastapi"]["T1"]
    check("S1 coverage diagnostics present (prior areas, owned files, zero-overlap prior areas = 0)", cov["prevAreas"] == 40 and cov["zeroOverlapPriorAreas"] == 0 and cov["prevOwnedFiles"] > 0 and cov["sharedOwnedFiles"] > 0, cov)
    om = r1["replay_meta"] or {}
    head = subprocess.run(["git", "rev-parse", "HEAD"], cwd=ROOT, capture_output=True, text=True).stdout.strip()
    check("S1 replay-meta: Scout HEAD, sha256 of areas.go and records.go, grid, Go version copied into results.json",
          om.get("scoutHead") == head and om.get("areasGoSha256") == C.sha256_file(os.path.join(ROOT, "internal", "areas", "areas.go"))
          and om.get("recordsGoSha256") == C.sha256_file(os.path.join(ROOT, "internal", "clawpatch", "records.go")) and om.get("gridPct") == C.GRID_PCT and om.get("goVersion", "").startswith("go"), om)
    text = A.render(r1)
    check("S1 rendered report covers transitions, worst-case, controls, chained survival, label distribution, the survival margin, the joint LOTO lines, the t* prominence line, the over-call check and the label models",
          all(s in text for s in ("T1", "worst", "controls:", "chained survival", "label distribution", "per-transition survival", "replication:", "operating characteristics",
                                  "survival margin at t*", "leave-one-transition-out, joint t*", "hold out T3", "AT t*=75", "over-call check", "label call rows by role")))
    ld = r1["arms"]["codex"]["label_distribution"]["fastapi"]
    check("S1 label distribution by class and by band [t, t+0.05)", set(ld["by_class"]) == {"remap", "mutation"} and "30" in ld["by_band_pct"] and "75" in ld["by_band_pct"], ld)
    u = cell(r1, "codex", "fastapi", "decisive", 30)
    check("S1 unique labelled pairs reported beside occurrences", 0 < u["pairs_unique_labelled"] <= u["labelled"], (u["pairs_unique_labelled"], u["labelled"]))
    loto1 = r1["arms"]["codex"]["loto"]
    check("S1 joint LOTO: six held-out transitions, each re-selects t* = 0.75 jointly; held-out survival 32/40 with 32 labelled and 0 different in both repositories",
          [x["held_out"] for x in loto1] == ["T%d" % i for i in range(1, 7)]
          and all(x["t_star_pct"] == 75 and set(x["per_repo"]) == {"fastapi", "axum"} and all(v["held_out_survival"] == 0.8 and v["matched"] == 32 and v["prev_areas"] == 40
                                                                                         and v["labelled"] == 32 and v["different"] == 0 for v in x["per_repo"].values()) for x in loto1), loto1[:1])
    at1 = r1["arms"]["codex"]["at_t_star"]
    check("S1 unlabelled share and worst-case spurious rate reported at t* = 0.75 per repository (192 matches, 0 unlabelled, worst case 0)",
          set(at1) == {"fastapi", "axum"} and all(v["matched"] == 192 and v["unlabelled"] == 0 and v["unlabelled_share"] == 0.0 and v["spurious_worst_case"] == 0.0 for v in at1.values()), at1)
    oc1 = ld["over_call_check_band_90"]
    check("S1 over-call check is reported with its denominators; the S1 matches all lie below Jaccard 0.90, so the band is empty (0 of 0 labelled, share undefined)",
          oc1 == {"different": 0, "labelled": 0, "matches": 0, "share": None}, oc1)
    calls1 = C.read_jsonl(os.path.join(p1, "calls.jsonl"))
    lm1 = r1["label_models"]
    check("S1 labeller models per role with call-row counts over every call row (controls included): A, B on every call, C on the adjudicated ones",
          lm1["A"] == {"stub-A": len(calls1)} and lm1["B"] == {"stub-B": len(calls1)} and lm1["C"] == {"stub-C": sum(1 for c in calls1 if c["C"] is not None)} and lm1["C"]["stub-C"] > 0, lm1)
    check("S1 agreement: C-equals-A share reported with its denominator (C answered validly; stub models differ -> 0 of N), sessions counted (no retries -> one per pair), exec-event pairs 0, old field gone",
          la["adjudicated_c_model_equals_a_of"] == la["adjudicated_c_ok"] > 0 and la["adjudicated_c_model_equals_a_share"] == 0.0 and la["codex_sessions"] == la["pairs"]
          and la["pairs_with_exec_event"] == 0 and "codex_sessions_with_exec" not in la, la)

    # ---- B blinding, side order against the key, control pairs
    leak = []
    blinded = open(os.path.join(p1, "blinded.jsonl")).read()
    for needle in ("reg_", "vol_", "heur_", "codex", "claude", "fastapi", "axum", "jaccard", "transition", "featureId", "only_left", "only_right", "control", "class"):
        if needle in blinded:
            leak.append(needle)
    check("B blinded states carry no identifiers, scores, arm, repo, transition, set-difference counts or control markers", not leak, leak)
    states = [json.loads(line) for line in blinded.splitlines()]
    check("B blinded row keys", all(set(s) == {"pair_id", "state"} for s in states))
    check("B state keys (no only_left/only_right)", all(set(s["state"]) == {"left", "right", "shared_files_total", "shared_files"} for s in states))
    keys = C.read_jsonl(os.path.join(p1, "key.jsonl"))
    ctl = C.read_jsonl(os.path.join(p1, "control-key.jsonl"))
    first = {}
    for k in keys:
        first.setdefault(k["pair_id"], k["left_is_prev"])
    share = sum(1 for v in first.values() if v) / len(first)
    check("B side order randomised against the key (share of unique pairs with the earlier map on the left %.3f)" % share, 0.4 < share < 0.6 and len(set(first.values())) == 2, share)
    n_real = len(first)
    kinds = {k: [c for c in ctl if c["kind"] == k] for k in ("swapped", "identical")}
    check("B controls: ceil(8 %) swapped and ceil(4 %) identical, expected labels sealed in control-key",
          len(kinds["swapped"]) == math.ceil(0.08 * n_real) and len(kinds["identical"]) == math.ceil(0.04 * n_real)
          and all(c["expected"] == "different_concern" for c in kinds["swapped"]) and all(c["expected"] == "same_concern" for c in kinds["identical"]), (n_real, len(kinds["swapped"]), len(kinds["identical"])))
    bl = {s["pair_id"]: s["state"] for s in states}
    check("B control pairs are in blinded.jsonl, disjoint from real pairs, and indistinguishable by row shape",
          all(c["pair_id"] in bl and c["pair_id"] not in first for c in ctl) and len(bl) == n_real + len(ctl))
    sw = kinds["swapped"][0]
    src, cst = bl[sw["source_pair_id"]], bl[sw["pair_id"]]
    sides = lambda s: sorted(json.dumps({k: v for k, v in x.items() if k in ("title", "summary", "entrypoints")}, sort_keys=True) for x in (s["left"], s["right"]))
    files = lambda s: sorted(json.dumps(x["owned_files"]) for x in (s["left"], s["right"]))
    check("B swapped control: exactly one side's title/summary replaced, file lists unchanged",
          files(src) == files(cst) and len(set(sides(src)) & set(sides(cst))) == 1, (sides(src), sides(cst)))
    # round-3 issue 8a: the replaced side carries the DONOR's entrypoints (a feature of the same map with that title), not the source side's own
    ent = lambda f: [e.get("path", "") + (":" + e["symbol"] if e.get("symbol") else "") for e in f.get("entrypoints") or []]
    donors = {}
    for mid_ in C.MAP_IDS:
        for f_ in C.load_active_features(os.path.join(d1, "data"), sw["arm"], sw["repo"], mid_) or []:
            donors.setdefault(f_["title"], ent(f_))
    src_sides = [src["left"], src["right"]]
    new_sides = [s_ for s_ in (cst["left"], cst["right"]) if s_["title"] not in {x["title"] for x in src_sides}]
    orig_side = next(x for x in src_sides if x["owned_files"] == new_sides[0]["owned_files"]) if len(new_sides) == 1 else None
    check("B swapped control: the replaced side's entrypoints equal the donor feature's (found by title in the same arm and repository) and differ from the source side's own entrypoints; removing the replacement would fail this",
          len(new_sides) == 1 and new_sides[0]["title"] in donors and new_sides[0]["entrypoints"] == donors[new_sides[0]["title"]] and orig_side is not None
          and new_sides[0]["entrypoints"] != orig_side["entrypoints"] and len(new_sides[0]["entrypoints"]) > 0, (new_sides, orig_side))
    idc = bl[kinds["identical"][0]["pair_id"]]
    check("B identical control: both sides equal", idc["left"] == idc["right"])
    check("S1 controls excluded from every rate, validity gate ok, kappa reported", r1["controls"]["ok"] and r1["controls"]["agreement"] is not None
          and r1["controls"]["kinds"]["swapped"]["misses"] == 0, r1["controls"]["kinds"]["swapped"])

    # ---- P permutation nulls on S1: gradient (exploratory) and decision level
    rows = C.read_jsonl(rp1)
    labels = C.final_labels(p1)
    pids = sorted({k["pair_id"] for k in keys})
    rs = []
    for seed in range(20):
        rnd = random.Random(20261023 + seed)
        vals = [labels[p] for p in pids]
        rnd.shuffle(vals)
        rs.append(A.gradient(rows, keys, dict(zip(pids, vals)), "codex")["r"])
    rs = [x for x in rs if x is not None]
    check("P permuted labels: mean |r| < 0.15 and max |r| < 0.3 over 20 seeds", sum(abs(x) for x in rs) / len(rs) < 0.15 and max(abs(x) for x in rs) < 0.3,
          (sum(abs(x) for x in rs) / len(rs), max(abs(x) for x in rs)))
    verdicts, spur50 = [], []
    for seed in range(10):
        rnd = random.Random(7000 + seed)
        vals = [labels[p] for p in pids]
        rnd.shuffle(vals)
        perm = dict(zip(pids, vals))
        pd = os.path.join(tmp, "S1perm%d" % seed)
        os.makedirs(pd)
        for fn in ("blinded.jsonl", "key.jsonl", "control-key.jsonl"):
            shutil.copy(os.path.join(p1, fn), pd)
        rr = run_label(tmp, pd, truth_labels(keys, lambda k, perm=perm: perm[k["pair_id"]], ctl))
        assert rr.returncode == 0, rr.stderr
        rres = A.analyze(os.path.join(d1, "data"), rp1, pd)
        verdicts.append(rres["arms"]["codex"]["decision"]["verdict"])
        spur50.append(cell(rres, "codex", "fastapi", "decisive", 50)["spurious"])
    check("P decision-level permutation null: S1 labels permuted over 10 seeds -> fallback every time, spurious ~0.2 at t=0.50", set(verdicts) == {"fallback"} and 0.1 < sum(spur50) / len(spur50) < 0.3, (verdicts, spur50))

    # ---- S2 planted null: clean re-map
    d2, r2, rp2, _p = run_scenario(tmp, "S2", {"kind": "clean", "n_reg": 40})
    ok2 = True
    for arm in ("codex", "claude"):
        for repo in ("fastapi", "axum"):
            for pct in C.GRID_PCT:
                c = cell(r2, arm, repo, "decisive", pct)
                ok2 &= c["different"] == 0 and close(c["survival"], 1.0)
    check("S2 survival 1 and no spurious merge at every threshold", ok2)
    dec2 = r2["arms"]["codex"]["decision"]
    check("S2 verdict adopt threshold 90 (largest on grid)", dec2["verdict"] == "adopt" and dec2["threshold_pct"] == 90, dec2)
    check("S2 gradient undefined (no spurious label)", r2["arms"]["codex"]["gradient_exploratory"]["r"] is None)
    rows2 = [r for r in C.read_jsonl(rp2) if r["status"] == "ok" and r["thresholdPct"] == 30]
    check("S2 coverage diagnostics: no zero-overlap prior areas in a clean re-map", all(r["zeroOverlapPriorAreas"] == 0 for r in rows2))
    oc2 = r2["arms"]["codex"]["label_distribution"]["fastapi"]["over_call_check_band_90"]
    check("S2 over-call check: every match of the identical re-map lies in the [0.90, 1.0] band, all labelled, 0 `different` (share 0.0 of N labelled, N = band matches > 0)",
          oc2["different"] == 0 and oc2["labelled"] == oc2["matches"] > 0 and oc2["share"] == 0.0, oc2)

    # ---- S3 planted null: label noise independent of a Jaccard spread (2 % hash-selected `different`)
    def rule3(k):
        return "different_concern" if int(hashlib.sha256(("noise" + k["pair_id"]).encode()).hexdigest()[:8], 16) % 100 < 2 else "same_concern"
    d3, r3, _r3, _p3 = run_scenario(tmp, "S3", {"kind": "spread", "n_reg": 40}, rule=rule3)
    flat, counts_ok = True, True
    for repo in ("fastapi", "axum"):
        for p in C.GRID_PCT:
            c = cell(r3, "codex", repo, "decisive", p)
            exp = 6 * 5 * sum(1 for x in SPREAD_J if x >= p / 100.0)
            counts_ok &= c["matched"] == exp
            if c["labelled"] >= 60:
                flat &= abs(c["spurious"] - 0.02) <= 4 * math.sqrt(0.02 * 0.98 / c["labelled"])
    check("S3 matched counts equal the planted Jaccard spread at every threshold", counts_ok)
    check("S3 spurious rate within binomial bounds of the planted 2 % at every threshold with >= 60 labelled (J spread 0.38-1.0)", flat)
    gr3 = r3["arms"]["codex"]["gradient_exploratory"]["r"]
    check("S3 gradient defined (J varies) and ~0 under J-independent noise", gr3 is not None and abs(gr3) < 0.2, gr3)
    d3c = r3["arms"]["codex"]["decision"]
    check("S3 planned verdict: survival edge at 0.45 (35/40 areas have J >= 0.45; 30/40 at 0.50) -> adopt 45 despite 2 % label noise",
          d3c["verdict"] == "adopt" and d3c["threshold_pct"] == 45 and d3c["feasible_pct"] == list(range(30, 46, 5)), d3c)
    # S3b over-call check with a planted over-call: `different` exactly where the match Jaccard is >= 0.90
    _d3b, r3b, _x3, _y3 = run_scenario(tmp, "S3b", {"kind": "spread", "n_reg": 40}, rule=lambda k: "different_concern" if k["jaccard"] >= 0.9 else "same_concern")
    oc3 = r3b["arms"]["codex"]["label_distribution"]["fastapi"]["over_call_check_band_90"]
    t90 = cell(r3b, "codex", "fastapi", "decisive", 90)
    check("S3b over-call check: labels `different` exactly for Jaccard >= 0.90 -> the [0.90, 1.0] band share is 1.0 (= matches = the t=0.90 cell's matched and different), reported with denominators",
          oc3["labelled"] > 0 and oc3["share"] == 1.0 and oc3["different"] == oc3["labelled"] == oc3["matches"] == t90["matched"] == t90["different"], (oc3, t90["matched"], t90["different"]))

    # ---- S4 survival boundary
    d4, r4, _r4, _p4 = run_scenario(tmp, "S4", {"kind": "edge", "n_reg": 40, "n_edge": 12})
    dec4 = r4["arms"]["codex"]["decision"]
    c65, c70 = cell(r4, "codex", "fastapi", "decisive", 65), cell(r4, "codex", "fastapi", "decisive", 70)
    check("S4 survival 1.0 at 0.65 and 28/40 at 0.70", close(c65["survival"], 1.0) and c70["survival_exact"] == "7/10", (c65["survival"], c70["survival_exact"]))
    check("S4 verdict adopt threshold 65 (survival edge, no spurious merge)", dec4["verdict"] == "adopt" and dec4["threshold_pct"] == 65 and dec4["feasible_pct"] == list(range(30, 66, 5)), dec4)
    bad, n = crosscheck(d4)
    check("R S4 replay equals Python reference at every threshold (%d rows)" % n, bad == 0, bad)
    d4s, _r4s, _x, _y = run_scenario(tmp, "S4split", {"kind": "split", "n_reg": 40})
    bad, n = crosscheck(d4s)
    check("R split-feature maps: replay equals Python reference at every threshold (%d rows)" % n, bad == 0, bad)
    d4t, r4t, _x, _y = run_scenario(tmp, "S4tie", {"kind": "tie", "n_reg": 12}, rule=lambda k: "same_concern")
    bad, n = crosscheck(d4t)
    check("R cross-area ties: replay equals the Python reference with Go's tie order (Jaccard, minted area ID, feature ID) (%d rows)" % n, bad == 0, bad)
    bad_old, _ = crosscheck(d4t, go_order=False)
    check("R tie scenario discriminates: ordering by feature ID alone (the round-1 reference) disagrees with Go on %d rows" % bad_old, bad_old > 0)
    tr = [r for r in C.read_jsonl(os.path.join(d4t, "replay.jsonl")) if r["status"] == "ok" and r["thresholdPct"] == 30 and r["arm"] == "codex" and r["repo"] == "fastapi"]
    check("R tie scenario: every tied group matches both areas (12 groups x 2 per transition)", all(r["matched"] == 24 and r["prevAreas"] == 24 for r in tr), [(r["matched"], r["prevAreas"]) for r in tr])

    # ---- S5 unrelated
    d5, r5, rp5, _p = run_scenario(tmp, "S5", {"kind": "unrelated", "n_reg": 20})
    dec5 = r5["arms"]["codex"]["decision"]
    check("S5 unrelated maps -> fallback, nothing matched", dec5["verdict"] == "fallback" and cell(r5, "codex", "axum", "decisive", 30)["matched"] == 0, dec5)
    cov5 = r5["arms"]["codex"]["coverage"]["axum"]
    check("S5 coverage diagnostics: every prior area has zero overlap (survival ceiling 0)", all(v["zeroOverlapPriorAreas"] == v["prevAreas"] == 20 for v in cov5.values()), cov5)
    chn = r5["arms"]["codex"]["chain"]["axum"]["30"]
    check("S5 chained fold: nothing matched, new areas created each step, F1 share 0", all(s["matched"] == 0 for s in chn) and chn[0]["created"] == 20 and chn[-1]["f1MatchedAllSoFar"] == 0, chn)

    # ---- S6 small n: 5 areas x 6 transitions = 30 matches per repo (Wilson upper of 0/30 is 0.114)
    _d6, r6, _r, _p = run_scenario(tmp, "S6", {"kind": "planted", "n_reg": 5, "n_vol": 0})
    dec6 = r6["arms"]["codex"]["decision"]
    check("S6 30 matches/repo: Wilson upper > 0.10 -> every grid threshold imprecise (precision guard), verdict inconclusive (the small control set also fails its minimum counts)",
          dec6["verdict"] == "inconclusive" and dec6["per_threshold"][50] == "imprecise" and cell(r6, "codex", "axum", "decisive", 50)["spurious_wilson"][1] > 0.10, dec6)

    # ---- S7 unlabelled share: a stable 30% of pair ids get no valid session answer
    def unlabelled_30(k):
        return None if int(k["pair_id"][:4], 16) % 10 < 3 else "same_concern"
    _d7, r7, _r, _p = run_scenario(tmp, "S7", {"kind": "planted", "n_reg": 40, "n_vol": 0}, rule=unlabelled_30)
    dec7 = r7["arms"]["codex"]["decision"]
    c7 = cell(r7, "codex", "fastapi", "decisive", 30)
    check("S7 unlabelled matches counted as unmeasured (never labelled vs unparsable split) and the label gate trips",
          c7["unlabelled"] > 0 and c7["unlabelled_unparsable"] == c7["unlabelled"] and dec7["verdict"] == "inconclusive" and dec7["tripped_gate"] == "label gate" and "label gate" in dec7["reason"], (c7["unlabelled"], dec7))
    k7 = C.read_jsonl(os.path.join(_p, "key.jsonl"))
    real7 = {k["pair_id"] for k in k7}
    calls7 = [c for c in C.read_jsonl(os.path.join(_p, "calls.jsonl")) if c["pair_id"] in real7]
    bad_b = sum(1 for c in calls7 if not c["B"]["ok"])
    ag7 = r7["label_agreement"]
    check("S7 session accounting: a pair whose codex session never parses costs two sessions (one retry), so codex_sessions = pairs + such pairs; pairs_with_exec_event is a pair count (0 here)",
          bad_b > 0 and ag7["pairs"] == len(calls7) and ag7["codex_sessions"] == len(calls7) + bad_b and ag7["pairs_with_exec_event"] == 0, (ag7, bad_b))

    # ---- S8 failed maps: unmeasured transitions, class gates
    def fail_maps(*maps, arm="codex", repo="fastapi"):
        def fix(data):
            for mid in maps:
                p = os.path.join(C.map_dir(data, arm, repo, mid), "meta.json")
                m = json.load(open(p))
                m["rc"] = 1
                json.dump(m, open(p, "w"))
        return fix
    _d8, r8, _r, _p = run_scenario(tmp, "S8a", {"kind": "clean", "n_reg": 40}, meta_fix=fail_maps("R1"))
    dec8 = r8["arms"]["codex"]["decision"]
    c8 = cell(r8, "codex", "fastapi", "same_state", 50)
    check("S8a R1 failure: T3 and T4 unmeasured with reasons; 2 remap + 2 mutation still measured -> class gates pass, adopt",
          len(c8["unmeasured_transitions"]) == 2 and all("R1" in u["reason"] for u in c8["unmeasured_transitions"]) and dec8["verdict"] == "adopt" and dec8["class_gate"]["fastapi"]["remap"]["measured"] == 2, (c8["unmeasured_transitions"], dec8))
    check("S8a other arm unaffected", r8["arms"]["claude"]["decision"]["verdict"] == "adopt")
    _d8, r8b, _r, _p = run_scenario(tmp, "S8b", {"kind": "clean", "n_reg": 40}, meta_fix=fail_maps("F2", "R1"))
    dec8b = r8b["arms"]["codex"]["decision"]
    check("S8b F2 and R1 fail: 1 measured remap transition -> class gate, inconclusive, reason names repo, class and gate",
          dec8b["verdict"] == "inconclusive" and dec8b["tripped_gate"] == "class gate" and "fastapi" in dec8b["reason"] and "remap" in dec8b["reason"], dec8b)
    _d8, r8c, _r, _p = run_scenario(tmp, "S8c", {"kind": "clean", "n_reg": 40}, meta_fix=fail_maps("M1"))
    dec8c = r8c["arms"]["codex"]["decision"]
    check("S8c M1 fails: T5 and T6 unmeasured -> mutation class gate, inconclusive",
          dec8c["verdict"] == "inconclusive" and dec8c["tripped_gate"] == "class gate" and "mutation" in dec8c["reason"], dec8c)
    check("S8c replication arm B not applicable when A is inconclusive", r8c["replication"]["status"].startswith("not applicable"), r8c["replication"])

    # ---- S9 invalid labels: swapped controls labelled same_concern
    _d9, r9, rp9, p9 = run_scenario(tmp, "S9", {"kind": "clean", "n_reg": 40}, ctl_rule=lambda c: "same_concern")
    dec9 = r9["arms"]["codex"]["decision"]
    check("S9 label-validity gate: all swapped controls labelled same_concern -> inconclusive naming the control gate despite perfect real-pair rates",
          dec9["verdict"] == "inconclusive" and dec9["tripped_gate"] == "control gate" and not r9["controls"]["ok"] and dec9["per_threshold"][90] == "pass", dec9)
    k9, c9 = C.read_jsonl(os.path.join(p9, "key.jsonl")), C.read_jsonl(os.path.join(p9, "control-key.jsonl"))

    def relabel9(name, ctl_rule):
        pd = os.path.join(tmp, name)
        os.makedirs(pd)
        for fn in ("blinded.jsonl", "key.jsonl", "control-key.jsonl"):
            shutil.copy(os.path.join(p9, fn), pd)
        rr = run_label(tmp, pd, truth_labels(k9, default_rule, c9, ctl_rule))
        assert rr.returncode == 0, rr.stderr
        return A.analyze(os.path.join(_d9, "data"), rp9, pd)
    for name, rule, kind in (("S9ct", lambda c: "cannot_tell", "swapped"),
                             ("S9ov", lambda c: "overlapping_concern" if c["kind"] == "swapped" else c["expected"], "swapped"),
                             ("S9io", lambda c: "overlapping_concern" if c["kind"] == "identical" else c["expected"], "identical")):
        rv = relabel9(name, rule)
        dv = rv["arms"]["codex"]["decision"]
        check("S9 %s: controls answered %s -> control gate trips (inconclusive) naming %s although real pairs pass" % (
            name, "cannot_tell on every control" if name == "S9ct" else "overlapping_concern on every %s control" % kind, kind),
              dv["verdict"] == "inconclusive" and dv["tripped_gate"] == "control gate" and kind in rv["controls"]["reason"] and dv["per_threshold"][90] == "pass", (dv, rv["controls"]["reason"]))
    ds_, rps_, ps_ = prepare(tmp, "S9short", {"kind": "clean", "n_reg": 1}, arms=("codex",))
    bp = subprocess.run([PY, os.path.join(X2, "build_pairs.py"), "--data", os.path.join(ds_, "data"), "--replay", rps_, "--out", ps_], capture_output=True, text=True)
    ks_, cs2 = C.read_jsonl(os.path.join(ps_, "key.jsonl")), C.read_jsonl(os.path.join(ps_, "control-key.jsonl"))
    assert run_label(ds_, ps_, truth_labels(ks_, default_rule, cs2)).returncode == 0
    rsh = A.analyze(os.path.join(ds_, "data"), rps_, ps_)
    check("S9 generation shortfall (one agent feature per map: no donor for swapped controls): build_pairs prints SHORTFALL, the control gate fails with 'none generated'",
          "SHORTFALL" in bp.stdout and rsh["controls"]["kinds"]["swapped"]["shortfall"] >= 1 and not rsh["controls"]["ok"] and "none generated" in rsh["controls"]["reason"], (bp.stdout, rsh["controls"]["reason"]))
    r9b = relabel9("S9claude", lambda c: "same_concern" if c["kind"] == "swapped" and c["arm"] == "claude" else c["expected"])
    dca, dcb = r9b["arms"]["codex"]["decision"], r9b["arms"]["claude"]["decision"]
    pa9b = r9b["controls"]["per_arm"]
    check("S9 arm-B-only misses (swapped controls whose arm is claude answered same_concern, arm A's controls correct): the pooled gate trips and BOTH arms are inconclusive naming the control gate although both arms' real pairs pass; per_arm shows the asymmetry (codex 0, claude 1.0)",
          dca["verdict"] == "inconclusive" and dcb["verdict"] == "inconclusive" and dca["tripped_gate"] == dcb["tripped_gate"] == "control gate" and dca["per_threshold"][90] == "pass"
          and dcb["per_threshold"][90] == "pass" and pa9b["codex"]["swapped"]["miss_rate"] == 0.0 and pa9b["claude"]["swapped"]["miss_rate"] == 1.0
          and pa9b["claude"]["swapped"]["usable"] > 0 and r9b["controls"]["kinds"]["swapped"]["misses"] == pa9b["claude"]["swapped"]["misses"], (dca, dcb, pa9b))

    # ---- S10 no prior areas; unparsable pairs; missing fields; reserved keys; --new-round
    _d10, r10, _r, _p = run_scenario(tmp, "S10a", {"kind": "clean", "n_reg": 0})
    dec10 = r10["arms"]["codex"]["decision"]
    c10 = cell(r10, "codex", "fastapi", "decisive", 50)
    check("S10a no prior areas: every transition unmeasured ('no prior areas'), class gate, inconclusive",
          len(c10["unmeasured_transitions"]) == 6 and all(u["reason"] == "no prior areas" for u in c10["unmeasured_transitions"]) and dec10["verdict"] == "inconclusive"
          and dec10["tripped_gate"] == "class gate", (c10["unmeasured_transitions"], dec10))
    d, replay, pairs = prepare(tmp, "S10b", {"kind": "clean", "n_reg": 30}, arms=("codex",))
    keys, ctl = C.read_jsonl(os.path.join(pairs, "key.jsonl")), C.read_jsonl(os.path.join(pairs, "control-key.jsonl"))
    real = sorted({k["pair_id"] for k in keys})
    unp = {p for p in real if int(p[:2], 16) % 6 == 0}
    odd = {p for p in real if int(p[:2], 16) % 6 == 1}
    truth = truth_labels(keys, lambda k: "same_concern", ctl)
    for p in unp:
        truth[p] = {"A": None, "B": None, "C": None}
    evil = {"label": "same_concern", "omit": ["rule", "reason"], "extra": {"ok": False, "model": "evil", "attempts": 9, "timestamp": "x", "raw": "x"}}
    for p in odd:
        truth[p] = {"A": evil, "B": "same_concern", "C": "same_concern"}
    r = run_label(d, pairs, truth)
    check("S10b label.py survives answers with missing rule/reason and reserved keys, and unparsable pairs", r.returncode == 0, r.stderr[-500:])
    calls = {c["pair_id"]: c for c in C.read_jsonl(os.path.join(pairs, "calls.jsonl"))}
    check("S10b session bookkeeping is not overridable by the model answer (ok, model, attempts)", all(calls[p]["A"]["ok"] is True and calls[p]["A"]["model"] == "stub-A" and calls[p]["A"]["attempts"] == 1 for p in odd) and len(calls) == len(real) + len(ctl))
    fl = C.final_labels(pairs, with_unparsable=True)
    check("S10b final labels: unparsable pairs recorded as unparsable, missing-field pairs labelled", all(fl[p] == "unparsable" for p in unp) and all(fl[p] == "same_concern" for p in odd) and len(unp) > 0 and len(odd) > 0, (len(unp), len(odd)))
    r1b = A.analyze(os.path.join(d, "data"), replay, pairs)
    cu = cell(r1b, "codex", "fastapi", "decisive", 30)
    check("S10b unmeasured matches split (unparsable vs cannot_tell vs never labelled)", cu["unlabelled_unparsable"] > 0 and cu["unlabelled_cannot_tell"] == 0 and cu["unlabelled_never_labelled"] == 0, cu)
    refuse = run_label(d, pairs, truth)
    check("S10b second run without --new-round refused (append-only)", refuse.returncode != 0 and "append-only" in refuse.stderr + refuse.stdout)
    n_calls_before = len(C.read_jsonl(os.path.join(pairs, "calls.jsonl")))
    truth2 = {p: {"A": "same_concern", "B": "same_concern", "C": "same_concern"} for p in list(truth)}
    nr = run_label(d, pairs, truth2, ("--new-round", "round2"))
    calls2 = C.read_jsonl(os.path.join(pairs, "calls.jsonl"))
    new = [c for c in calls2 if c.get("round") == "round2"]
    check("S10b --new-round labels only the pairs without a usable label (exactly the unparsable ones)", nr.returncode == 0 and {c["pair_id"] for c in new} == unp and len(calls2) == n_calls_before + len(unp), (nr.stderr[-300:], len(new), len(unp)))
    r2b = A.analyze(os.path.join(d, "data"), replay, pairs)
    cu2 = cell(r2b, "codex", "fastapi", "decisive", 30)
    check("S10b after the new round: no unparsable matches, agreement still counts round-1 calls only, verdict adopt",
          cu2["unlabelled_unparsable"] == 0 and r2b["label_agreement"]["pairs"] == len(real) and r2b["arms"]["codex"]["decision"]["verdict"] == "adopt", (cu2, r2b["label_agreement"]))
    again = run_label(d, pairs, truth2, ("--new-round", "round3"))
    check("S10b a further new round finds nothing to label", again.returncode == 0 and len(C.read_jsonl(os.path.join(pairs, "calls.jsonl"))) == len(calls2))
    lock = os.path.join(pairs, "labels.lock")
    check("S10b labels.lock is removed after a run", not os.path.exists(lock))
    open(lock, "w").close()
    lk = run_label(d, pairs, truth2, ("--new-round", "round4"))
    check("S10b a leftover or concurrent labels.lock refuses a labelling run", lk.returncode != 0 and "labels.lock" in lk.stderr + lk.stdout and os.path.exists(lock), lk.stderr[-300:])
    os.unlink(lock)
    check("S10b an uninterrupted run leaves no labelled pair without a call row", r2b["label_integrity"]["labelled_pairs_without_call_row"] == 0, r2b["label_integrity"])
    cut = os.path.join(tmp, "S10b-cut")
    shutil.copytree(pairs, cut)
    cp_rows = open(os.path.join(cut, "calls.jsonl")).read().splitlines()
    with open(os.path.join(cut, "calls.jsonl"), "w") as f:
        f.write("\n".join(cp_rows[1:]) + "\n")
    rcut = A.analyze(os.path.join(d, "data"), replay, cut)
    check("S10b a kill between the label rows and the call row (call row missing) is reported, never an orphan call row", rcut["label_integrity"]["labelled_pairs_without_call_row"] == 1, rcut["label_integrity"])
    with open(os.path.join(cut, "calls.jsonl"), "a") as f:
        f.write('{"pair_id": "partial", "A": {"la')
    try:
        C.read_jsonl(os.path.join(cut, "calls.jsonl"))
        msg = None
    except SystemExit as e:
        msg = str(e)
    check("S10b a partial trailing line makes read_jsonl fail loudly naming file and line", msg is not None and "malformed" in msg and "calls.jsonl:%d" % len(cp_rows) in msg, msg)
    before = C.sha256_file(os.path.join(pairs, "blinded.jsonl"))
    bp2 = subprocess.run([PY, os.path.join(X2, "build_pairs.py"), "--data", os.path.join(d, "data"), "--replay", replay, "--out", pairs], capture_output=True, text=True)
    check("S10b build_pairs refuses once labels exist for the guide version and leaves blinded.jsonl untouched",
          bp2.returncode != 0 and "refusing" in bp2.stderr + bp2.stdout and C.sha256_file(os.path.join(pairs, "blinded.jsonl")) == before, bp2.stderr[-300:])
    xd = os.path.join(tmp, "rsdata")
    os.makedirs(os.path.join(xd, "pairs"))
    with open(os.path.join(xd, "pairs", "labels.jsonl"), "w") as f:
        f.write(json.dumps({"pair_id": "x", "label": "same_concern", "provenance": "agent_reviewed", "guide_version": C.GUIDE_VERSION}) + "\n")
    rs = [subprocess.run(["bash", os.path.join(X2, "run.sh"), st], env=dict(os.environ, X2_DATA=xd), capture_output=True, text=True) for st in ("replay", "pairs")]
    check("S10b run.sh replay and run.sh pairs refuse once labels.jsonl has rows", all(r_.returncode != 0 and "refusing" in r_.stderr for r_ in rs), [r_.stderr[-200:] for r_ in rs])

    # ---- S12 calibrated decision-level null and effect: Jaccard-independent Bernoulli `different` labels at 3, 5, 7 % through label.py (stub) -> analyze()
    d12, rp12, p12 = prepare(tmp, "S12", {"kind": "spread", "n_reg": 40}, arms=("codex",))
    k12, c12 = C.read_jsonl(os.path.join(p12, "key.jsonl")), C.read_jsonl(os.path.join(p12, "control-key.jsonl"))
    seeds = 30
    rates = {}
    for p in (0.03, 0.05, 0.07):
        adopt = at30 = 0
        for seed in range(seeds):
            pd = os.path.join(tmp, "S12-%s-%d" % (p, seed))
            os.makedirs(pd)
            for fn in ("blinded.jsonl", "key.jsonl", "control-key.jsonl"):
                shutil.copy(os.path.join(p12, fn), pd)

            def bern(k, p=p, seed=seed):
                u = int(hashlib.sha256(("cal%s-%d-%s" % (p, seed, k["pair_id"])).encode()).hexdigest()[:8], 16) / 2 ** 32
                return "different_concern" if u < p else "same_concern"
            rr = run_label(tmp, pd, truth_labels(k12, bern, c12))
            assert rr.returncode == 0, rr.stderr
            dv = A.analyze(os.path.join(d12, "data"), rp12, pd)["arms"]["codex"]["decision"]
            adopt += dv["verdict"] == "adopt"
            at30 += dv["per_threshold"][30] == "pass"
            shutil.rmtree(pd)
        rates[p] = (adopt / seeds, at30 / seeds)
    oc240 = A.operating_characteristics((240,))["by_n"]["240"]["table"]
    print("     S12 P(adopt any t) / P(pass at t=0.30) over %d seeds:" % seeds, {p: rates[p] for p in rates}, "exact table n=240:", {p: round(oc240[str(p)]["p_both_repos"], 4) for p in rates})
    tol = lambda q: max(4 * math.sqrt(q * (1 - q) / seeds), 0.05)
    check("S12 P(adopt at some threshold) is ordered 3 % > 5 % > 7 % and no less than P(pass at t=0.30)", rates[0.03][0] > rates[0.05][0] > rates[0.07][0] and all(a_ >= b_ for a_, b_ in rates.values()), rates)
    check("S12 P(pass at t=0.30), where all 240 matches per repo are labelled, is within Monte-Carlo tolerance (4 sigma, >= 0.05) of the exact table at n = 240",
          all(abs(rates[p][1] - oc240[str(p)]["p_both_repos"]) <= tol(oc240[str(p)]["p_both_repos"]) for p in rates), (rates, {p: oc240[str(p)]["p_both_repos"] for p in rates}))

    # ---- E environment, C collector end to end
    environment_and_collector_checks(tmp)

    shutil.rmtree(tmp, ignore_errors=True)
    print("\n%d check(s) failed" % len(FAILS) if FAILS else "\nall validation checks passed")
    sys.exit(1 if FAILS else 0)


def environment_and_collector_checks(tmp):
    node = shutil.which("node")
    if not node:
        check("C node available (source .tools/activate)", False)
        return
    # ---- E: the credential scrub is Scout's, name for name
    names = ["PATH", "HOME", "ANTHROPIC_API_KEY", "OPENAI_API_KEY", "MY_SECRET_THING", "GITHUB_TOKEN", "GH_TOKEN", "INFISICAL_TOKEN", "OPENROUTER_API_KEY", "openrouter_x",
             "ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_BASE_URL", "OPENAI_BASE_URL", "OPENAI_ORG_ID", "CLAUDE_CODE_OAUTH_TOKEN", "ANTHROPIC_MODEL", "OPENAI_MODEL", "CODEX_HOME",
             "SOMETHING_API_KEY", "x_api_key", "SECRETARY", "APIKEY", "TOKEN", "CLAWPATCH_CLAUDE_AUTH_CONTEXT", "CLAWPATCH_CODEX_TIMEOUT_MS", "X2_PLAIN"]
    out = subprocess.run([go_bin(tmp, "credcheck", "./evaluation/x2-area-stability/validation/credcheck")], input="\n".join(names) + "\n", capture_output=True, text=True).stdout
    go = {line.split("\t")[0]: line.split("\t")[1] == "1" for line in out.splitlines()}
    mismatch = [n for n in names if go.get(n) != C.is_credential(n)]
    check("E common.is_credential equals modelcli.IsCredential on %d names" % len(names), not mismatch and go.get("CLAUDE_CODE_OAUTH_TOKEN") is False and go.get("SECRETARY") is True, mismatch)
    scrubbed = C.scrub_env({n: "v" for n in names}, ["CLAWPATCH_CLAUDE_AUTH_CONTEXT=host"])
    check("E scrub_env keeps non-credentials, drops credentials, extra wins", "CLAUDE_CODE_OAUTH_TOKEN" in scrubbed and "ANTHROPIC_MODEL" in scrubbed and "ANTHROPIC_API_KEY" not in scrubbed
          and scrubbed["CLAWPATCH_CLAUDE_AUTH_CONTEXT"] == "host" and "MY_SECRET_THING" not in scrubbed)

    src = os.path.join(tmp, "tiny-src")
    os.makedirs(src)
    git = lambda *a, cwd=src: subprocess.run(["git", "-c", "user.name=t", "-c", "user.email=t@e.invalid", *a], cwd=cwd, check=True, capture_output=True, text=True)
    git("init", "-q")
    for dname in ("alpha", "beta", "gamma"):
        os.makedirs(os.path.join(src, dname))
        for k in range(4):
            open(os.path.join(src, dname, "m%d.py" % k), "w").write("x = %d\n" % k)
    os.makedirs(os.path.join(src, "tests"))
    for k in range(3):
        open(os.path.join(src, "tests", "t%d.py" % k), "w").write("y = %d\n" % k)
    git("add", "-A")
    git("commit", "-q", "-m", "init")
    rev = git("rev-parse", "HEAD").stdout.strip()
    # a repository in which no leaf directory can qualify for the M2 rename (only excluded directories and root files)
    src3 = os.path.join(tmp, "tiny3-src")
    os.makedirs(src3)
    git3 = lambda *a: git(*a, cwd=src3)
    git3("init", "-q")
    for dname, stem in (("tests", "t"), ("docs", "d"), (".", "a")):
        os.makedirs(os.path.join(src3, dname), exist_ok=True)
        for k in range(3 if dname != "." else 4):
            open(os.path.join(src3, dname, "%s%d.py" % (stem, k)), "w").write("z = %d\n" % k)
    git3("add", "-A")
    git3("commit", "-q", "-m", "init")
    rev3 = git3("rev-parse", "HEAD").stdout.strip()
    spec = json.dumps({"tiny": {"rev": rev, "url": src, "ext": ".py"}, "tiny2": {"rev": rev, "url": src, "ext": ".py"}, "tiny3": {"rev": rev3, "url": src3, "ext": ".py"}})
    # amendment 1: the collector accepts only the review-safety build, so the stub lives in a fixture build directory (marker = the real pins, sanitizer token in dist/)
    pv_, ps_ = C.review_safety_pins()
    sb = os.path.join(tmp, "stubbuild")
    os.makedirs(os.path.join(sb, "dist"))
    stub = os.path.join(sb, "dist", "cli.js")
    shutil.copy(os.path.join(HERE, "stub_clawpatch.js"), stub)
    open(os.path.join(sb, "dist", "prompt.js"), "w").write("// fixture: %s\n" % C.SANITIZER_TOKEN)
    open(os.path.join(sb, C.BUILD_MARKER_FILE), "w").write("%s %s\n" % (pv_, ps_))
    bins = os.path.join(tmp, "fakebin")
    os.makedirs(bins)

    def fake_cli(name, body, where=bins):
        p = os.path.join(where, name)
        open(p, "w").write("#!%s\n%s" % (PY, body))
        os.chmod(p, 0o755)

    def fake_claude(version, where=bins, name="claude", tag=""):
        fake_cli(name, 'import json, os, sys\na = sys.argv[1:]\nif a[:1] == ["--version"]:\n    print("@VER@ (Claude Code)")\n    sys.exit(0)\nsys.stdin.read()\n'
                 'log = os.environ.get("X2_FAKE_ENVLOG")\nif log:\n    open(log, "a").write("\\n".join(sorted(os.environ)) + "\\n")\n'
                 'open(os.path.abspath(__file__) + ".calls", "a").write(json.dumps({"argv": a, "env": sorted(os.environ)}) + "\\n")\n'
                 'm = "fake-claude@TAG@-model" + ("-overridden" if "--setting-sources" in a or os.environ.get("ANTHROPIC_MODEL") else "")\n'
                 'print(json.dumps({"result": json.dumps({"label": "same_concern", "rule": "R1", "reason": "fake"}), "modelUsage": {m: {}}}))\n'.replace("@VER@", version).replace("@TAG@", tag), where)

    def fake_codex(where=bins):
        fake_cli("codex", 'import json, os, sys, time\na = sys.argv[1:]\nif a[:1] == ["--version"]:\n    vf, vs = os.environ.get("X2_FAKE_COUNTER"), os.environ.get("X2_FAKE_VERSION_SLEEP_CALL")\n    if vf and vs:\n        vn = (int(open(vf + ".ver").read()) if os.path.exists(vf + ".ver") else 0) + 1\n        open(vf + ".ver", "w").write(str(vn))\n        if str(vn) == vs:\n            time.sleep(120)\n    print("codex-cli 0.0.0-fake" + ("-rolled" if os.path.exists(os.environ.get("X2_FAKE_COUNTER", "-") + ".rolled") else ""))\n    sys.exit(0)\nsys.stdin.read()\n'
                 'log = os.environ.get("X2_FAKE_ENVLOG")\nif log:\n    open(log, "a").write("\\n".join(sorted(os.environ)) + "\\n")\n'
                 'open(os.path.abspath(__file__) + ".calls", "a").write(json.dumps({"argv": a, "env": sorted(os.environ)}) + "\\n")\n'
                 'rolled = False\ncf = os.environ.get("X2_FAKE_COUNTER")\nif cf:\n    n = (int(open(cf).read()) if os.path.exists(cf) else 0) + 1\n    open(cf, "w").write(str(n))\n'
                 '    if os.environ.get("X2_FAKE_SLEEP_CALL") == str(n):\n        time.sleep(120)\n'
                 'if "-o" in a:\n    open(a[a.index("-o") + 1], "w").write(json.dumps({"label": "same_concern", "rule": "R1", "reason": "fake"}))\n'
                 'if os.environ.get("X2_FAKE_ROLL_FROM") and cf and n >= int(os.environ["X2_FAKE_ROLL_FROM"]):\n    rolled = True\n    open(cf + ".rolled", "w").write("1")\n'
                 'print("model: fake-codex-model" + ("-rolled" if rolled else "") + ("-isolated" if "--ignore-user-config" in a else ""))\n'
                 'if os.environ.get("X2_FAKE_EXEC"):\n    print("exec")\n    print("ls")\n', where)
    fake_claude("2.1.169")
    fake_codex()

    poison = {"ANTHROPIC_API_KEY": "k", "OPENAI_API_KEY": "k", "MY_SECRET_THING": "k", "GITHUB_TOKEN": "k", "OPENROUTER_FOO": "k", "OPENAI_BASE_URL": "k",
              "CLAUDE_CODE_OAUTH_TOKEN": "keep", "ANTHROPIC_MODEL": "keep", "X2_PLAIN": "keep", "CLAWPATCH_CLAUDE_AUTH_CONTEXT": "isolated"}

    def collect(name, fail_calls="", arm="codex", env_extra=None, args=(), data=None, work=None, repo="tiny", source=None, wait=True, hup_ignored=False):
        d = data or os.path.join(tmp, "col-" + name)
        env = dict(os.environ, X2_STUB_COUNTER=d + "-counter", X2_STUB_FAIL_CALLS=fail_calls, X2_STUB_ENVLOG=d + "-envlog.jsonl", **poison)
        env["PATH"] = bins + os.pathsep + os.environ["PATH"]
        for n_ in C.TIMEOUT_ENV_NAMES:
            env.pop(n_, None)
        for n_ in (*C.REFUSED_MODEL_ENV, *C.REFUSED_SANDBOX_ENV, "CLAWPATCH_CLAUDE_BIN", "CODEX_HOME"):  # the operator's own setting must not reach the collector under test
            env.pop(n_, None)
        env.update(env_extra or {})
        cmd = [PY, os.path.join(X2, "collect.py"), "--arm", arm, "--repo", repo, "--source", source or src, "--work", work or d + "-work", "--data", d,
               "--node", node, "--cli", stub, "--spec", spec, "--expect-clawpatch-version", "stub-0", "--term-grace", "2", *args]
        ign = (lambda: signal.signal(signal.SIGHUP, signal.SIG_IGN)) if hup_ignored else None  # what `nohup` does before exec
        if not wait:
            return d, subprocess.Popen(cmd, env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, preexec_fn=ign)
        return d, subprocess.run(cmd, env=env, capture_output=True, text=True, preexec_fn=ign)

    # ---- normal run (codex arm): environment, provenance, retry, mutations
    d, r = collect("ok", "2")  # call 2 = F2's first attempt fails once, retry succeeds
    check("C collector exits 0", r.returncode == 0, r.stderr[-800:])
    metas = {m: C.read_meta(d, "codex", "tiny", m) for m in C.MAP_IDS}
    check("C all 7 maps recorded with rc 0", all(v is not None and v["rc"] == 0 for v in metas.values()), {k: (v or {}).get("rc") for k, v in metas.items()})
    check("C every map's meta.json records workspace_dirty (git status --porcelain after the map) and it is false on the normal run; the inventory carries it",
          all(v["workspace_dirty"] is False for v in metas.values()) and all(m_["workspace_dirty"] is False for m_ in A.inventory(d, [("codex", "tiny")])["maps"]), {k: v.get("workspace_dirty") for k, v in metas.items()})
    log = [json.loads(line) for line in open(d + "-envlog.jsonl")]
    all_names = sorted({n for e in log for n in e["names"]})
    cred = subprocess.run([go_bin(tmp, "credcheck", "x")], input="\n".join(all_names) + "\n", capture_output=True, text=True).stdout
    leaked = [line.split("\t")[0] for line in cred.splitlines() if line.endswith("\t1")]
    check("E every Clawpatch call (init and map, %d invocations) runs with CLAWPATCH_CLAUDE_AUTH_CONTEXT=host, BOTH provider timeouts at 900000 ms and no credential name" % len(log),
          len(log) >= 8 and not leaked and all(e["claudeAuth"] == "host" and e["codexTimeout"] == "900000" and e["claudeTimeout"] == "900000" for e in log),
          (leaked, [(e["claudeAuth"], e["codexTimeout"], e["claudeTimeout"]) for e in log][:3]))
    check("E non-credential names Scout keeps are kept (CLAUDE_CODE_OAUTH_TOKEN, ANTHROPIC_MODEL, X2_PLAIN); poisoned credentials are gone",
          all(n in all_names for n in ("CLAUDE_CODE_OAUTH_TOKEN", "ANTHROPIC_MODEL", "X2_PLAIN")) and not any(n in all_names for n in ("ANTHROPIC_API_KEY", "OPENAI_API_KEY", "MY_SECRET_THING", "GITHUB_TOKEN", "OPENROUTER_FOO", "OPENAI_BASE_URL")))
    maplog = [e for e in log if e["cmd"] == "map"]
    check("E map invocations use Scout's arguments (--source agent --provider codex --json)", all("--json" in e["args"] and e["args"][e["args"].index("--source") + 1] == "agent" and e["args"][e["args"].index("--provider") + 1] == "codex" for e in maplog))
    tv = json.load(open(os.path.join(d, "tool-versions-codex-tiny.json")))
    check("C provenance: per-(arm, repo) tool-versions with Scout HEAD, areas.go and records.go sha256, env names, timeouts",
          tv["scout_head"] == subprocess.run(["git", "rev-parse", "HEAD"], cwd=ROOT, capture_output=True, text=True).stdout.strip() and tv["areas_go_sha256"] == C.sha256_file(os.path.join(ROOT, "internal", "areas", "areas.go"))
          and tv["records_go_sha256"] == C.sha256_file(os.path.join(ROOT, "internal", "clawpatch", "records.go")) and tv["clawpatch"] == "stub-0" and tv["claude_auth_context"] == "host"
          and "ANTHROPIC_API_KEY" not in tv["env_names"] and tv["call_timeout_seconds"] == 1800 and tv["wall_max_seconds"] == 8 * 3600
          and tv["codex_timeout_ms"] == "900000" and tv["claude_timeout_ms"] == "900000" and tv["clawpatch_cli_sha256"] == C.sha256_file(stub)
          and tv["model_probe_before"]["model"] == "fake-codex-model" and tv["model_probe_after"]["model"] == "fake-codex-model" and tv["term_grace_seconds"] == 2, tv)
    check("C tool-versions records the PRESENCE (never the value) of CLAWPATCH_CLAUDE_BIN, ANTHROPIC_MODEL and CODEX_HOME",
          tv["env_presence"] == {"CLAWPATCH_CLAUDE_BIN": False, "ANTHROPIC_MODEL": True, "CODEX_HOME": False} and "keep" not in json.dumps(tv["env_presence"]), tv["env_presence"])
    cbs = tv.get("clawpatch_build_safety") or {}
    check("C tool-versions records the build-safety check: marker text equal to the expected pinned \"<version> <sha256>\", marker_ok, sanitizer_in_dist, ok",
          cbs.get("ok") is True and cbs.get("marker_ok") is True and cbs.get("sanitizer_in_dist") is True and cbs.get("marker") == cbs.get("expected") == "%s %s" % (pv_, ps_), cbs)
    check("C tool-versions records the CLI versions again after the after-probe (cli_versions_after equals the start versions when nothing changed) and, on the synthetic repo, no workspace-level configuration file",
          tv["cli_versions_after"] == {k_: tv[k_] for k_ in ("claude", "codex")} and set(tv["workspace_config_presence"]) == set(C.WORKSPACE_CONFIG_PATHS) and not any(tv["workspace_config_presence"].values()), tv["cli_versions_after"])
    check("C every attempt records a UTC start time (started_utc) in the attempt files and meta.json attempts",
          all(re.fullmatch(r"\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ", a_["started_utc"]) for m_ in metas.values() for a_ in m_["attempts"])
          and re.fullmatch(r"\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ", json.load(open(os.path.join(d, "attempts", "codex", "tiny", "F2.attempt1.json")))["started_utc"]), metas["F2"]["attempts"])
    check("C F2 retried once: attempts kept (attempt files), per-attempt provider/model/seconds in meta", metas["F2"]["attempt_count"] == 2 and os.path.exists(os.path.join(d, "attempts", "codex", "tiny", "F2.attempt1.json"))
          and metas["F2"]["attempts"][0]["rc"] == 3 and metas["F2"]["attempts"][1]["rc"] == 0 and metas["F2"]["attempts"][1]["provider"] == "codex" and metas["F2"]["attempts"][1]["model_option"] == "default"
          and metas["F2"]["model_option_reported"] == "default" and "model_reported" not in metas["F2"] and metas["F2"]["seconds_total"] >= metas["F2"]["seconds_last"], metas["F2"])
    check("C M1 added package; M2 renamed the deterministic target `alpha` (fewest files tie -> path ascending; `tests` and `x2_newpkg` excluded)",
          metas["M1"].get("mutation") == "M1" and metas["M2"].get("renamed") == "alpha", (metas["M1"].get("mutation"), metas["M2"]))
    all_r1 = [json.load(open(p)) for p in sorted(glob.glob(os.path.join(C.map_dir(d, "codex", "tiny", "R1"), "features", "*.json")))]
    active_r1 = C.load_active_features(d, "codex", "tiny", "R1")
    check("C same-state map keeps stale records but analysis loads only active ones", any(f["status"] == "skipped" for f in all_r1) and len(active_r1) == 5, (len(all_r1), len(active_r1)))
    w2 = os.path.join(tmp, "col-ok-work2")
    _d, r_again = collect("ok", data=d, work=w2)
    check("C all seven map directories are checked before cloning: rerun refused with no workspace created", r_again.returncode != 0 and "append-only" in r_again.stderr + r_again.stdout and not os.path.exists(w2), r_again.stderr[-300:])

    # ---- full pipeline on the collected maps
    trans = os.path.join(d, "t.json")
    C.write_transitions(trans, arms=["codex"], repos=["tiny"])
    rp = os.path.join(d, "replay.jsonl")
    run_replay(tmp, os.path.join(d, "maps"), trans, rp)
    os.makedirs(os.path.join(d, "pairs"))
    subprocess.run([PY, os.path.join(X2, "build_pairs.py"), "--data", d, "--replay", rp, "--out", os.path.join(d, "pairs")], check=True, capture_output=True)
    keys, ctl = C.read_jsonl(os.path.join(d, "pairs", "key.jsonl")), C.read_jsonl(os.path.join(d, "pairs", "control-key.jsonl"))
    rl = run_label(d, os.path.join(d, "pairs"), truth_labels(keys, lambda k: "same_concern", ctl))
    check("C labelling on collected maps", rl.returncode == 0, rl.stderr[-300:])
    res = A.analyze(d, rp, os.path.join(d, "pairs"))
    t6 = [x for x in C.read_jsonl(rp) if x["transition"] == "T6" and x["thresholdPct"] == 50][0]
    check("C replay on collected maps: churned IDs still match by file set; renamed dir does not (T6: 5 prior areas, 4 matched)", t6["prevAreas"] == 5 and t6["matched"] == 4, t6)
    pm = res["arms"]["codex"]["tables"]["tiny"]["decisive"]["50"]
    check("C analysis on collected maps: decisive pooled survival 24/25, per-refresh mean 29/30 over six transitions", close(pm["survival_pooled"], 24 / 25) and pm["survival_exact"] == "29/30", pm)
    check("C analysis on collected maps: the unrolled collection's arm is probe-CONSISTENT (before and after name one model, CLI versions unchanged) and therefore citable, and the render says so",
          res["probe_consistency"]["codex"]["consistent"] and res["probe_consistency"]["codex"]["models"] == ["fake-codex-model"] and res["probe_consistency"]["codex"]["cli_versions_changed"] == {}
          and res["arms"]["codex"]["decision"]["citable"] is True and "[citable for a design change: yes]" in A.render(res), res["probe_consistency"])
    inv_d = A.inventory(d, [("codex", "tiny")])
    inv_dirty = {**inv_d, "maps": [{**inv_d["maps"][0], "workspace_dirty": True}, *inv_d["maps"][1:]]}
    check("C a map that left the workspace dirty is listed in the rendered inventory line; a clean run lists none", '"workspace_dirty": true' in A.render(dict(res, inventory=inv_dirty)) and '"workspace_dirty": true' not in A.render(res),
          "workspace_dirty" in A.render(dict(res, inventory=inv_dirty)))

    # ---- persistent failure; stop rules; timeout; partial write; failed M1; claude arm
    d2, r2 = collect("fail", "2,3")  # F2 fails both attempts (map calls 2 and 3)
    m2 = C.read_meta(d2, "codex", "tiny", "F2")
    check("C persistent failure recorded as rc!=0, collection continues", r2.returncode == 0 and m2["rc"] != 0 and C.read_meta(d2, "codex", "tiny", "F3")["rc"] == 0, (r2.stderr[-300:], m2 and m2["rc"]))
    trans2 = os.path.join(d2, "t.json")
    C.write_transitions(trans2, arms=["codex"], repos=["tiny"])
    rp2 = os.path.join(d2, "replay.jsonl")
    run_replay(tmp, os.path.join(d2, "maps"), trans2, rp2)
    t1 = [x for x in C.read_jsonl(rp2) if x["transition"] == "T1" and x["thresholdPct"] == 50][0]
    check("C failed map makes its transition unmeasured", t1["status"] == "unmeasured" and "F2" in t1["reason"], t1)

    d3, r3 = collect("stop3", "1,2,3,4,5,6")  # F1, F2, F3 fail twice each
    st = json.load(open(os.path.join(d3, "stopped-codex-tiny.json"))) if os.path.exists(os.path.join(d3, "stopped-codex-tiny.json")) else {}
    check("C stop rule: three consecutive failed maps stop the (arm, repo) before R1 with stopped.json; later maps absent",
          r3.returncode == 0 and "consecutive" in st.get("reason", "") and st.get("stopped_before") == "R1" and st.get("not_attempted") == ["R1", "R2", "M1", "M2"]
          and all(C.read_meta(d3, "codex", "tiny", m)["rc"] != 0 for m in ("F1", "F2", "F3")) and C.read_meta(d3, "codex", "tiny", "R1") is None, (st, r3.stdout[-300:]))
    trans3 = os.path.join(d3, "t.json")
    C.write_transitions(trans3, arms=["codex"], repos=["tiny"])
    rp3 = os.path.join(d3, "replay.jsonl")
    run_replay(tmp, os.path.join(d3, "maps"), trans3, rp3)
    t3 = [x for x in C.read_jsonl(rp3) if x["transition"] == "T4" and x["thresholdPct"] == 50][0]
    check("C stopped maps are unmeasured (not recorded)", t3["status"] == "unmeasured" and "not recorded" in t3["reason"], t3)
    d4, r4 = collect("wall", args=("--wall-max", "1"), env_extra={"X2_STUB_SLEEP_ALL": "2"})
    st4 = json.load(open(os.path.join(d4, "stopped-codex-tiny.json"))) if os.path.exists(os.path.join(d4, "stopped-codex-tiny.json")) else {}
    check("C stop rule: wall-clock limit stops before the next map", r4.returncode == 0 and "wall-clock" in st4.get("reason", "") and C.read_meta(d4, "codex", "tiny", "F1")["rc"] == 0 and C.read_meta(d4, "codex", "tiny", "F2") is None, (st4, r4.stderr[-200:]))
    d5, r5 = collect("timeout", args=("--call-timeout", "3"), env_extra={"X2_STUB_SLEEP_CALLS": "1", "X2_STUB_SLEEP_SECONDS": "60"})
    m5 = C.read_meta(d5, "codex", "tiny", "F1") or {}
    a5 = m5.get("attempts", [{}, {}])
    check("C per-call timeout: hung attempt killed (rc 124, timed_out), retried, map recorded", r5.returncode == 0 and a5[0].get("timed_out") is True and a5[0].get("rc") == 124 and a5[0].get("seconds", 99) < 20 and a5[1].get("rc") == 0 and m5.get("rc") == 0, m5)
    d6, r6 = collect("partial", env_extra={"X2_STUB_PARTIAL_CALLS": "4"})  # call 4 = R1's first attempt: leaves a junk active record and exits 3
    m6 = C.read_meta(d6, "codex", "tiny", "R1") or {}
    a6 = m6.get("attempts", [{}, {}])
    check("C retry of a non-fresh map restores the pre-map chain state (snapshot): no junk from the partial attempt, restoration recorded",
          r6.returncode == 0 and m6.get("rc") == 0 and a6[0].get("rc") == 3 and a6[1].get("restored_snapshot") is True and a6[0].get("restored_snapshot") is False
          and not os.path.exists(os.path.join(C.map_dir(d6, "codex", "tiny", "R1"), "features", "feat_partial_junk.json")), (m6, r6.stderr[-200:]))
    d7, r7 = collect("m1fail", "6,7")  # M1 fails twice (calls 6 and 7)
    m7 = C.read_meta(d7, "codex", "tiny", "M2") or {}
    check("C failed M1: M2 target cannot be selected from M1 -> mutation skipped with that reason; M1 recorded as failed", r7.returncode == 0 and (C.read_meta(d7, "codex", "tiny", "M1") or {}).get("rc") != 0 and "M1" in str(m7.get("skipped")), (m7, r7.stderr[-200:]))
    # claude arm: version gate and Scout's environment with the claude provider
    cb = os.path.join(tmp, "claude-old")
    os.makedirs(cb)
    fake_claude("2.0.0", cb)
    w8 = os.path.join(tmp, "col-claudeold-work")
    d8, r8 = collect("claudeold", arm="claude", env_extra={"PATH": cb + os.pathsep + os.environ["PATH"]}, work=w8)
    check("C claude arm refuses Claude Code < 2.1.169 before cloning", r8.returncode != 0 and "2.1.169" in r8.stderr + r8.stdout and not os.path.exists(w8) and not os.path.exists(os.path.join(d8, "maps")), r8.stderr[-300:])
    fake_claude("2.1.169")
    d9, r9 = collect("claudeok", arm="claude", env_extra={"CLAWPATCH_CODEX_TIMEOUT_MS": "123456", "CLAWPATCH_CLAUDE_TIMEOUT_MS": "654321"})
    log9 = [json.loads(line) for line in open(d9 + "-envlog.jsonl")]
    tv9 = json.load(open(os.path.join(d9, "tool-versions-claude-tiny.json"))) if os.path.exists(os.path.join(d9, "tool-versions-claude-tiny.json")) else {}
    check("E claude arm: --provider claude, host auth context, Scout's own CLAWPATCH_CODEX_TIMEOUT_MS and CLAWPATCH_CLAUDE_TIMEOUT_MS win on every call, claude version and served model recorded",
          r9.returncode == 0 and all(e["claudeAuth"] == "host" and e["codexTimeout"] == "123456" and e["claudeTimeout"] == "654321" for e in log9)
          and all(e["args"][e["args"].index("--provider") + 1] == "claude" for e in log9 if e["cmd"] == "map") and "2.1.169" in tv9.get("claude", "")
          and tv9.get("claude_timeout_ms") == "654321" and tv9.get("codex_timeout_ms") == "123456"
          and tv9["model_probe_before"]["model"] == "fake-claude-model" and tv9["model_probe_after"]["model"] == "fake-claude-model", (r9.stderr[-300:], tv9))
    check("E claude arm default: with no Scout value both timeouts default to 900000 (the claude provider's own 180 s cap is raised, N1)",
          all(e["claudeTimeout"] == "900000" for e in [json.loads(line) for line in open(d + "-envlog.jsonl")]) and C.clawpatch_map_env({})["CLAWPATCH_CLAUDE_TIMEOUT_MS"] == "900000")
    d10, r10 = collect("claudeok", arm="claude", env_extra={"PATH": bins + os.pathsep + os.environ["PATH"]}, data=d9, repo="tiny2")
    # round-3 issue 1: the probes run in the model-resolution context of the map calls
    cl_calls = [json.loads(line) for line in open(os.path.join(bins, "claude.calls"))]
    allowed_names = set(C.CLAUDE_AUTH_ENV_KEYS) | {"PATH", "HOME", "USERPROFILE", "CLAUDE_CONFIG_DIR", "XDG_CONFIG_HOME", "XDG_CACHE_HOME", "XDG_DATA_HOME", "TMPDIR", "TEMP", "TMP",
                                                   "CLAUDE_CODE_SUBPROCESS_ENV_SCRUB", "CLAUDE_CODE_OAUTH_TOKEN", "AWS_PROFILE", "SystemRoot", "ComSpec", "PATHEXT", "USER", "LOGNAME"}
    interp_added = {"LC_CTYPE", "__CF_USER_TEXT_ENCODING", "__PYVENV_LAUNCHER__", "PYTHONCOERCECLOCALE"}  # names the fake's own Python interpreter / macOS add to an otherwise empty environment
    check("C claude model probes (%d calls): flags mirror providers/claude.ts (-p json, --tools \"\", --safe-mode, --strict-mcp-config + empty MCP config, --disable-slash-commands, --no-chrome, --no-session-persistence) and NO --setting-sources" % len(cl_calls),
          len(cl_calls) >= 4 and all(all(f in c["argv"] for f in ("-p", "--safe-mode", "--strict-mcp-config", "--disable-slash-commands", "--no-chrome", "--no-session-persistence"))
                                      and c["argv"][c["argv"].index("--tools") + 1] == "" and c["argv"][c["argv"].index("--mcp-config") + 1] == '{"mcpServers":{}}'
                                      and c["argv"][c["argv"].index("--output-format") + 1] == "json" and "--setting-sources" not in c["argv"] for c in cl_calls), cl_calls[:1])
    check("C claude model probes run in the ported claudeEnv(host) environment: names within the allow-list, required names present, and no ANTHROPIC_MODEL, X2_PLAIN, X2_STUB_*, credential or CLAWPATCH_* name from the operator environment",
          all(set(c["env"]) - interp_added <= allowed_names and {"PATH", "HOME", "CLAUDE_CODE_SUBPROCESS_ENV_SCRUB", "CLAUDE_CODE_OAUTH_TOKEN", "TMPDIR", "XDG_CONFIG_HOME"} <= set(c["env"])
              and not ({"ANTHROPIC_MODEL", "X2_PLAIN", "MY_SECRET_THING", "OPENAI_API_KEY", "X2_STUB_COUNTER"} & set(c["env"])) and not any(n.startswith("CLAWPATCH_") for n in c["env"]) for c in cl_calls),
          [sorted(set(c["env"]) - allowed_names) for c in cl_calls][:2])
    cx_calls = [json.loads(line) for line in open(os.path.join(bins, "codex.calls"))]
    check("C codex model probes (%d calls) mirror providers/codex.ts: exec --cd <tmp> --sandbox read-only --skip-git-repo-check -, NO --ignore-user-config/--ignore-rules/--ephemeral/-o, under Clawpatch's own environment (credentials scrubbed, timeouts and auth context present, ANTHROPIC_MODEL kept)" % len(cx_calls),
          len(cx_calls) >= 4 and all(c["argv"][0] == "exec" and c["argv"][c["argv"].index("--sandbox") + 1] == "read-only" and "--cd" in c["argv"] and "--skip-git-repo-check" in c["argv"] and c["argv"][-1] == "-"
                                      and not ({"--ignore-user-config", "--ignore-rules", "--ephemeral", "-o", "-C", "-s"} & set(c["argv"]))
                                      and {"CLAWPATCH_CLAUDE_AUTH_CONTEXT", "CLAWPATCH_CODEX_TIMEOUT_MS", "CLAWPATCH_CLAUDE_TIMEOUT_MS", "ANTHROPIC_MODEL"} <= set(c["env"])
                                      and not ({"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "MY_SECRET_THING"} & set(c["env"])) for c in cx_calls), cx_calls[:1])

    def run_fake(name, args, env):
        return subprocess.run([os.path.join(bins, name), *args], input="x", capture_output=True, text=True, env=env)
    cx_iso = run_fake("codex", ["exec", "--ignore-user-config", "-"], {}).stdout
    cx_plain = run_fake("codex", ["exec", "-"], {}).stdout
    cl_set = json.loads(run_fake("claude", ["-p", "--setting-sources", "", "--output-format", "json"], {}).stdout)["modelUsage"]
    cl_env = json.loads(run_fake("claude", ["-p", "--output-format", "json"], {"ANTHROPIC_MODEL": "x"}).stdout)["modelUsage"]
    cl_plain = json.loads(run_fake("claude", ["-p", "--output-format", "json"], {}).stdout)["modelUsage"]
    check("C the fake CLIs are flag- and environment-sensitive (so the probe checks discriminate): codex answers a different model with --ignore-user-config; claude answers a different model with --setting-sources or ANTHROPIC_MODEL",
          "model: fake-codex-model-isolated" in cx_iso and cx_plain.strip() == "model: fake-codex-model" and set(cl_set) == {"fake-claude-model-overridden"}
          and set(cl_env) == {"fake-claude-model-overridden"} and set(cl_plain) == {"fake-claude-model"}, (cx_iso, cx_plain, cl_set, cl_env, cl_plain))
    for nm_ in (*C.REFUSED_MODEL_ENV, *C.REFUSED_SANDBOX_ENV):
        wr = os.path.join(tmp, "col-refuse-%s-work" % nm_)
        dr, rr = collect("refuse-" + nm_, env_extra={nm_: "" if nm_ in ("CLAWPATCH_PROVIDER", "CLAWPATCH_CONFIG", "CLAWPATCH_CODEX_SANDBOX") else "x"}, work=wr)
        check("C preflight refuses before cloning when %s is set (even empty: Clawpatch's `??` lets it override), so no Clawpatch-level model, effort, provider, config or sandbox override reaches the maps" % nm_,
              rr.returncode != 0 and nm_ in rr.stderr + rr.stdout and not os.path.exists(wr) and not os.path.exists(os.path.join(dr, "maps")) and not os.path.exists(os.path.join(dr, "tool-versions-codex-tiny.json")), rr.stderr[-300:])
    alt = os.path.join(tmp, "altbin")
    os.makedirs(alt)
    fake_claude("2.1.169", alt, name="claude-alt", tag="-alt")
    dalt, ralt = collect("claudealt", arm="claude", env_extra={"CLAWPATCH_CLAUDE_BIN": os.path.join(alt, "claude-alt")})
    tvalt = json.load(open(os.path.join(dalt, "tool-versions-claude-tiny.json"))) if os.path.exists(os.path.join(dalt, "tool-versions-claude-tiny.json")) else {}
    check("C CLAWPATCH_CLAUDE_BIN is honoured by the claude probes (the executable the maps run) and recorded by presence and version",
          ralt.returncode == 0 and tvalt["model_probe_before"]["model"] == "fake-claude-alt-model" and tvalt["model_probe_after"]["model"] == "fake-claude-alt-model"
          and tvalt["env_presence"]["CLAWPATCH_CLAUDE_BIN"] is True and "2.1.169" in tvalt["claude_executable_version"], (ralt.stderr[-300:], tvalt.get("model_probe_before")))
    check("C a second repository's versions file does not overwrite the first (names per arm and repo)", r10.returncode == 0 and os.path.exists(os.path.join(d9, "tool-versions-claude-tiny.json"))
          and os.path.exists(os.path.join(d9, "tool-versions-claude-tiny2.json")), r10.stderr[-300:])

    # ---- round-2 guards: product drift, CLI preflight, version pin, process termination, abort record, skipped M2, label environment
    gosrc = open(os.path.join(ROOT, "internal", "clawpatch", "client.go")).read()
    mm = re.search(r'defaultMapTimeoutMS\s*=\s*"(\d+)"', gosrc)
    check("E common's default map timeout equals client.go's defaultMapTimeoutMS and both timeout variables are set by mapEnvironment",
          bool(mm) and mm.group(1) == C.DEFAULT_MAP_TIMEOUT_MS and all('"%s"' % n in gosrc for n in C.TIMEOUT_ENV_NAMES), mm and mm.group(1))
    eb = os.path.join(tmp, "emptybin")
    os.makedirs(eb)
    w11 = os.path.join(tmp, "col-nocli-work")
    dn, rn = collect("nocli", env_extra={"PATH": eb}, work=w11)
    check("C preflight refuses before cloning when the arm's CLI is not on PATH (no workspace, no versions or stop record)",
          rn.returncode != 0 and "not on PATH" in rn.stderr + rn.stdout and not os.path.exists(w11) and not os.path.exists(os.path.join(dn, "tool-versions-codex-tiny.json"))
          and not os.path.exists(os.path.join(dn, "stopped-codex-tiny.json")), rn.stderr[-300:])
    w12 = os.path.join(tmp, "col-wrongver-work")
    dw, rw = collect("wrongver", args=("--expect-clawpatch-version", "0.7.3"), work=w12)
    check("C a Clawpatch --version other than the pinned 0.7.3 is refused before cloning", rw.returncode != 0 and "0.7.3" in rw.stderr + rw.stdout and not os.path.exists(w12)
          and not os.path.exists(os.path.join(dw, "maps")), rw.stderr[-300:])
    ub = os.path.join(tmp, "stubbuild-unpatched")  # an unpatched local build: no marker file, no sanitizer
    os.makedirs(os.path.join(ub, "dist"))
    shutil.copy(os.path.join(HERE, "stub_clawpatch.js"), os.path.join(ub, "dist", "cli.js"))
    ob = os.path.join(tmp, "stubbuild-old")  # a marker from an older patch revision over a patched dist/
    os.makedirs(os.path.join(ob, "dist"))
    shutil.copy(os.path.join(HERE, "stub_clawpatch.js"), os.path.join(ob, "dist", "cli.js"))
    open(os.path.join(ob, "dist", "prompt.js"), "w").write(C.SANITIZER_TOKEN)
    open(os.path.join(ob, C.BUILD_MARKER_FILE), "w").write("scout-review-sanitize-v2 f39d73da\n")
    nb = os.path.join(tmp, "stubbuild-nosan")  # the pinned marker over a dist/ without the sanitizer
    os.makedirs(os.path.join(nb, "dist"))
    shutil.copy(os.path.join(HERE, "stub_clawpatch.js"), os.path.join(nb, "dist", "cli.js"))
    open(os.path.join(nb, C.BUILD_MARKER_FILE), "w").write("%s %s\n" % (pv_, ps_))
    wsb = os.path.join(tmp, "stubbuild-wrongsha")  # the pinned version with a wrong sha over a patched dist/
    os.makedirs(os.path.join(wsb, "dist"))
    shutil.copy(os.path.join(HERE, "stub_clawpatch.js"), os.path.join(wsb, "dist", "cli.js"))
    open(os.path.join(wsb, "dist", "prompt.js"), "w").write(C.SANITIZER_TOKEN)
    open(os.path.join(wsb, C.BUILD_MARKER_FILE), "w").write("%s %s\n" % (pv_, "0" * 64))

    def probe_calls():  # model calls the fake CLIs have logged so far (every model call, hence the probe, appends one line to <cli>.calls; --version does not)
        n_ = 0
        for cli_ in ("codex", "claude"):
            f_ = os.path.join(bins, cli_ + ".calls")
            n_ += len(open(f_).read().splitlines()) if os.path.exists(f_) else 0
        return n_
    for tag_, bdir_, why_ in (("nomarker", ub, "no .scout-review-safety"), ("oldmarker", ob, "not the pinned"), ("wrongsha", wsb, "not the pinned"), ("nosanitizer", nb, C.SANITIZER_TOKEN)):
        wb_ = os.path.join(tmp, "col-%s-work" % tag_)
        calls_before_ = probe_calls()
        db_, rb_ = collect(tag_, args=("--cli", os.path.join(bdir_, "dist", "cli.js")), work=wb_)
        check("C a Clawpatch build that is not the review-safety build (%s) is refused before any probe, clone or data file (no fake-CLI model call logged)" % tag_,
              rb_.returncode != 0 and "review-safety build" in rb_.stderr + rb_.stdout and why_ in rb_.stderr + rb_.stdout and not os.path.exists(wb_)
              and not os.path.exists(os.path.join(db_, "maps")) and not glob.glob(os.path.join(db_, "tool-versions-*")) and probe_calls() == calls_before_ and calls_before_ > 0,
              ((rb_.stderr + rb_.stdout)[-300:], calls_before_, probe_calls()))
    import collect as COL
    try:
        COL.sh(["sleep", "5"], timeout=1)
        msg = None
    except SystemExit as e:
        msg = str(e)
    check("C version and other bounded calls: sh() with a timeout exits naming the command", msg is not None and "timed out" in msg and "sleep" in msg, msg)

    def alive(pid):
        try:
            os.kill(pid, 0)
            return True
        except ProcessLookupError:
            return False

    def gone(pid, secs=10):
        for _ in range(int(secs * 10)):
            if not alive(pid):
                return True
            time.sleep(0.1)
        return False
    pid1 = os.path.join(tmp, "child1.pid")
    dc, rcc = collect("child", env_extra={"X2_STUB_CHILD_CALLS": "1", "X2_STUB_CHILD_PIDFILE": pid1}, args=("--call-timeout", "3"))
    mc = C.read_meta(dc, "codex", "tiny", "F1") or {}
    ac = mc.get("attempts", [{}, {}])
    child1 = int(open(pid1).read())
    check("C per-call timeout ends the Clawpatch process group with SIGTERM (recorded), the detached provider child is gone after run_limited returns, the retry succeeds",
          rcc.returncode == 0 and ac[0].get("timed_out") is True and ac[0].get("rc") == 124 and ac[0].get("ended_by") == "SIGTERM" and gone(child1) and ac[1].get("rc") == 0 and mc.get("rc") == 0,
          (mc, rcc.stderr[-300:]))
    pid2 = os.path.join(tmp, "child2.pid")
    dk, rkk = collect("childkill", env_extra={"X2_STUB_CHILD_CALLS": "1", "X2_STUB_CHILD_PIDFILE": pid2, "X2_STUB_IGNORE_TERM": "1"}, args=("--call-timeout", "3"))
    mk_ = C.read_meta(dk, "codex", "tiny", "F1") or {}
    ak = mk_.get("attempts", [{}, {}])
    child2 = int(open(pid2).read())
    survives = alive(child2)
    os.kill(child2, signal.SIGKILL)
    check("C a Clawpatch that ignores SIGTERM is SIGKILLed after the grace (ended_by SIGKILL); the check discriminates: the detached child then survives, so child cleanup rests on Clawpatch's SIGTERM handler",
          rkk.returncode == 0 and ak[0].get("ended_by") == "SIGKILL" and ak[0].get("timed_out") is True and survives and ak[1].get("rc") == 0, (mk_, survives))
    pid3 = os.path.join(tmp, "child3.pid")
    da, pa = collect("abort", env_extra={"X2_STUB_CHILD_CALLS": "1", "X2_STUB_CHILD_PIDFILE": pid3}, wait=False)
    for _ in range(600):
        if os.path.exists(pid3) and open(pid3).read().strip():
            break
        time.sleep(0.1)
    time.sleep(0.5)
    pa.send_signal(signal.SIGTERM)
    try:
        _o, _e = pa.communicate(timeout=60)
    except subprocess.TimeoutExpired:
        pa.kill()
        _o, _e = pa.communicate()
    sp = os.path.join(da, "stopped-codex-tiny.json")
    st9 = json.load(open(sp)) if os.path.exists(sp) else {}
    tva = json.load(open(os.path.join(da, "tool-versions-codex-tiny.json")))
    child3 = int(open(pid3).read())
    check("C abort (SIGTERM to the collector during a map): stop record 'aborted: ...' written, the Clawpatch process group ended and its detached child gone, no map recorded, versions file notes the missing probe",
          pa.returncode != 0 and st9.get("reason", "").startswith("aborted") and st9.get("stopped_before") == "F1" and st9.get("not_attempted", [])[:2] == ["F1", "F2"]
          and gone(child3) and C.read_meta(da, "codex", "tiny", "F1") is None and tva.get("model_probe_after") == "not taken: aborted", (st9, _e[-300:]))
    _d, rre = collect("abort", data=da, work=os.path.join(tmp, "col-abort-work2"))
    check("C a rerun after an abort is refused (stop record, versions file, attempts are append-only)", rre.returncode != 0 and "append-only" in rre.stderr + rre.stdout)
    # round-3 issue 4: an abort during the after-probe and a SIGHUP still write the stop record
    probe_counter = os.path.join(tmp, "probe-counter.txt")
    dap, pap = collect("abortprobe", env_extra={"X2_FAKE_COUNTER": probe_counter, "X2_FAKE_SLEEP_CALL": "2"}, wait=False)  # codex call 2 is the after-probe: it sleeps
    for _ in range(1200):
        if os.path.exists(probe_counter) and open(probe_counter).read().strip() == "2":
            break
        time.sleep(0.1)
    time.sleep(0.5)
    pap.send_signal(signal.SIGTERM)
    try:
        _o, _e = pap.communicate(timeout=60)
    except subprocess.TimeoutExpired:
        pap.kill()
        _o, _e = pap.communicate()
    sap = os.path.join(dap, "stopped-codex-tiny.json")
    stap = json.load(open(sap)) if os.path.exists(sap) else {}
    tvap = json.load(open(os.path.join(dap, "tool-versions-codex-tiny.json")))
    check("C abort during the after-probe: all seven maps recorded, stop record `aborted` with phase after-probe and nothing left to attempt, versions file keeps the before-probe and notes `not taken: aborted`, and a rerun is refused",
          pap.returncode != 0 and all((C.read_meta(dap, "codex", "tiny", m_) or {}).get("rc") == 0 for m_ in C.MAP_IDS) and stap.get("reason", "").startswith("aborted") and stap.get("phase") == "after-probe"
          and stap.get("stopped_before") == "after-probe" and stap.get("not_attempted") == [] and tvap.get("model_probe_after") == "not taken: aborted" and tvap["model_probe_before"]["model"] == "fake-codex-model"
          and collect("abortprobe", data=dap, work=os.path.join(tmp, "col-abortprobe-work2"))[1].returncode != 0, (stap, tvap.get("model_probe_after"), _e[-300:]))
    # round-5 issue 2a: an abort during the CLI-version collection that follows a SUCCESSFUL after-probe keeps the probe; only what was not taken is marked
    ver_ctr = os.path.join(tmp, "ver-counter.txt")
    dvr, pvr = collect("abortver", env_extra={"X2_FAKE_COUNTER": ver_ctr, "X2_FAKE_VERSION_SLEEP_CALL": "2"}, wait=False)  # codex --version call 2 is cli_versions() after the after-probe
    for _ in range(1200):
        if os.path.exists(ver_ctr + ".ver") and open(ver_ctr + ".ver").read().strip() == "2":
            break
        time.sleep(0.1)
    time.sleep(0.5)
    pvr.send_signal(signal.SIGTERM)
    try:
        _o, _e = pvr.communicate(timeout=60)
    except subprocess.TimeoutExpired:
        pvr.kill()
        _o, _e = pvr.communicate()
    svr = os.path.join(dvr, "stopped-codex-tiny.json")
    stvr = json.load(open(svr)) if os.path.exists(svr) else {}
    tvvr = json.load(open(os.path.join(dvr, "tool-versions-codex-tiny.json")))
    check("C abort during the CLI-version collection after a successful after-probe: the after-probe is KEPT (model named), `cli_versions_after` is `not taken: aborted`, stop record phase after-probe, all seven maps recorded",
          pvr.returncode != 0 and isinstance(tvvr.get("model_probe_after"), dict) and tvvr["model_probe_after"].get("model") == "fake-codex-model" and tvvr.get("cli_versions_after") == "not taken: aborted"
          and stvr.get("phase") == "after-probe" and stvr.get("reason", "").startswith("aborted") and all((C.read_meta(dvr, "codex", "tiny", m_) or {}).get("rc") == 0 for m_ in C.MAP_IDS), (stvr, tvvr.get("model_probe_after"), tvvr.get("cli_versions_after"), _e[-300:]))
    pid4 = os.path.join(tmp, "child4.pid")
    dh, ph = collect("hup", env_extra={"X2_STUB_CHILD_CALLS": "1", "X2_STUB_CHILD_PIDFILE": pid4}, wait=False)
    for _ in range(600):
        if os.path.exists(pid4) and open(pid4).read().strip():
            break
        time.sleep(0.1)
    time.sleep(0.5)
    ph.send_signal(signal.SIGHUP)
    try:
        _o, _e = ph.communicate(timeout=60)
    except subprocess.TimeoutExpired:
        ph.kill()
        _o, _e = ph.communicate()
    sh_ = os.path.join(dh, "stopped-codex-tiny.json")
    sthup = json.load(open(sh_)) if os.path.exists(sh_) else {}
    check("C SIGHUP to the collector (dropped terminal) is routed like SIGTERM: stop record `aborted ... signal 1`, Clawpatch process group ended and its detached child gone, no map recorded, probe note written",
          ph.returncode != 0 and "signal 1" in sthup.get("reason", "") and sthup.get("reason", "").startswith("aborted") and sthup.get("phase") == "maps" and sthup.get("stopped_before") == "F1"
          and gone(int(open(pid4).read())) and C.read_meta(dh, "codex", "tiny", "F1") is None
          and json.load(open(os.path.join(dh, "tool-versions-codex-tiny.json"))).get("model_probe_after") == "not taken: aborted", (sthup, _e[-300:]))

    # round-4 issue 3: a SIGHUP ignored at start (what `nohup` sets) stays ignored: the collector is not aborted by it
    dI, pI = collect("hupign", env_extra={"X2_STUB_SLEEP_ALL": "1.5"}, wait=False, hup_ignored=True)
    for _ in range(600):
        if os.path.exists(dI + "-counter"):
            break
        time.sleep(0.1)
    time.sleep(0.3)
    pI.send_signal(signal.SIGHUP)
    try:
        _o, _e = pI.communicate(timeout=180)
    except subprocess.TimeoutExpired:
        pI.kill()
        _o, _e = pI.communicate()
    tvI = json.load(open(os.path.join(dI, "tool-versions-codex-tiny.json"))) if os.path.exists(os.path.join(dI, "tool-versions-codex-tiny.json")) else {}
    check("C a SIGHUP ignored at start (`nohup`) is kept: a SIGHUP during the maps does not abort; all seven maps recorded, no stop record, after-probe taken (exit 0)",
          pI.returncode == 0 and not os.path.exists(os.path.join(dI, "stopped-codex-tiny.json")) and all((C.read_meta(dI, "codex", "tiny", m_) or {}).get("rc") == 0 for m_ in C.MAP_IDS)
          and isinstance(tvI.get("model_probe_after"), dict), (pI.returncode, _e[-300:]))
    # round-4 issue 4: the after-probe names another model and the CLI version changed between the probes: collector records both, analysis flags the arm
    roll_ctr = os.path.join(tmp, "roll-counter.txt")
    dR, rR = collect("roll", env_extra={"X2_FAKE_COUNTER": roll_ctr, "X2_FAKE_ROLL_FROM": "2"})
    tvR = json.load(open(os.path.join(dR, "tool-versions-codex-tiny.json")))
    check("C a roll between the probes is recorded: after-probe model differs from the before-probe, CLI versions are taken again (`cli_versions_after`) and differ for codex",
          rR.returncode == 0 and tvR["model_probe_before"]["model"] == "fake-codex-model" and tvR["model_probe_after"]["model"] == "fake-codex-model-rolled"
          and tvR["cli_versions_after"]["codex"] == tvR["codex"] + "-rolled" and tvR["cli_versions_after"]["claude"] == tvR["claude"], (tvR["model_probe_after"], tvR["codex"], tvR["cli_versions_after"]))
    pcR = A.probe_consistency(A.inventory(dR, [("codex", "tiny")])["tool_versions"], [("codex", "tiny")])["codex"]
    resR = dict(res, probe_consistency={"codex": pcR})
    textR = A.render(resR)
    line_tiny = "   tiny probes: before %s, after %s; CLI versions before %s, after %s" % (json.dumps(tvR["model_probe_before"], sort_keys=True), json.dumps(tvR["model_probe_after"], sort_keys=True),
                                                                                  json.dumps({"codex": tvR["codex"]}, sort_keys=True), json.dumps({"codex": tvR["cli_versions_after"]["codex"]}, sort_keys=True))
    check("C analysis flags the arm INCONSISTENT when the probes name different models (reason named, the codex CLI-version change listed with both values, not citable); the per-repository line carries both probes and both CLI versions exactly; the unrolled run is CONSISTENT",
          not pcR["consistent"] and "probes name different models" in pcR["reasons"] and pcR["cli_versions_changed"] == {"codex": sorted([tvR["codex"], tvR["codex"] + "-rolled"])} and pcR["models"] == ["fake-codex-model", "fake-codex-model-rolled"]
          and "probe consistency (section 2, arm codex): INCONSISTENT" in textR and "not citable for a design change" in textR and line_tiny in textR.split("\n")
          and "   tiny workspace config presence (names only): " in textR
          and res["probe_consistency"]["codex"]["consistent"] is True and res["probe_consistency"]["codex"]["cli_versions_changed"] == {}
          and "probe consistency (section 2, arm codex): CONSISTENT" in A.render(res), (pcR, line_tiny, textR[-600:]))
    # round-4 issue 1: workspace-level configuration is recorded by presence (names only) after the clone
    src4 = os.path.join(tmp, "tiny4-src")
    subprocess.run(["git", "clone", "-q", "--no-local", src, src4], check=True, capture_output=True)
    os.makedirs(os.path.join(src4, ".claude"))
    os.makedirs(os.path.join(src4, ".codex"))
    open(os.path.join(src4, ".claude", "settings.local.json"), "w").write('{"model": "SECRET-MODEL-NAME"}\n')
    open(os.path.join(src4, ".codex", "config.toml"), "w").write('model = "x"\n')
    open(os.path.join(src4, "clawpatch.config.json"), "w").write("{}\n")
    git("add", "-f", "-A", cwd=src4)  # -f: a global gitignore may hide .claude/settings.local.json
    git("commit", "-q", "-m", "workspace configuration", cwd=src4)
    rev4 = git("rev-parse", "HEAD", cwd=src4).stdout.strip()
    spec4 = json.dumps({"tiny": {"rev": rev4, "url": src4, "ext": ".py"}})
    d4c, r4c = collect("wscfg", source=src4, args=("--spec", spec4))
    tv4c = json.load(open(os.path.join(d4c, "tool-versions-codex-tiny.json")))
    check("C tool-versions records which workspace-level configuration files exist (names only; none on the synthetic repo, three on the planted one) and never their content",
          r4c.returncode == 0 and tv["workspace_config_presence"] == {p_: False for p_ in C.WORKSPACE_CONFIG_PATHS}
          and tv4c["workspace_config_presence"] == {".claude/settings.json": False, ".claude/settings.local.json": True, ".codex": True, "clawpatch.config.json": True, ".clawpatch/config.json": False}
          and "SECRET-MODEL-NAME" not in json.dumps(tv4c), (r4c.stderr[-300:], tv4c.get("workspace_config_presence")))
    # round-4 issue 5: a failure during the setup (here: the pinned revision is absent from the source) writes no stop record and no data file; the rerun needs only the workspace removed
    src5 = os.path.join(tmp, "tiny5-src")
    os.makedirs(src5)
    git("init", "-q", cwd=src5)
    open(os.path.join(src5, "x.py"), "w").write("x = 1\n")
    git("add", "-A", cwd=src5)
    git("commit", "-q", "-m", "other", cwd=src5)
    w5 = os.path.join(tmp, "col-setup-work")
    d5, r5 = collect("setup", source=src5, work=w5)
    check("C a failure during the setup (clone/checkout) exits non-zero and leaves no stop record, no versions file and no attempt or map file",
          r5.returncode != 0 and "command failed" in r5.stderr and (not os.path.exists(d5) or os.listdir(d5) == []) and os.path.isdir(os.path.join(w5, "codex", "tiny")), (r5.stderr[-300:], os.path.exists(d5) and os.listdir(d5)))
    _d, r5b = collect("setup", data=d5, source=src5, work=w5)
    shutil.rmtree(os.path.join(w5, "codex", "tiny"))
    _d, r5c = collect("setup", data=d5, work=w5)
    check("C after a setup failure the rerun is refused until the workspace is removed (message names it), then succeeds with nothing else cleaned",
          r5b.returncode != 0 and "workspace exists" in r5b.stderr and r5c.returncode == 0 and all((C.read_meta(d5, "codex", "tiny", m_) or {}).get("rc") == 0 for m_ in C.MAP_IDS), (r5b.stderr[-200:], r5c.stderr[-300:]))
    # round-5 issue 2c: a leftover `base` without `ws` (a failed clone, an interrupted setup) is refused in the preflight, before any model call
    wB = os.path.join(tmp, "col-baseonly-work")
    os.makedirs(os.path.join(wB, "codex", "tiny"))
    ctrB = os.path.join(tmp, "baseonly-counter.txt")
    dBo, rBo = collect("baseonly", work=wB, env_extra={"X2_FAKE_COUNTER": ctrB})
    check("C a leftover workspace directory without a clone is refused in the preflight before the before-probe (no fake model call), the message names it, and nothing is written to data",
          rBo.returncode != 0 and "workspace exists" in rBo.stderr and os.path.join(wB, "codex", "tiny") in rBo.stderr and not os.path.exists(ctrB)
          and not os.path.exists(os.path.join(dBo, "tool-versions-codex-tiny.json")) and not os.path.exists(os.path.join(dBo, "stopped-codex-tiny.json")), (rBo.stderr[-300:], os.path.exists(ctrB)))
    # N4: skipped M2 (no leaf directory can qualify): no map call, rc null, T6 unmeasured everywhere, class gate satisfied by T5
    d11, r11 = collect("skip", repo="tiny3", source=src3)
    m11 = C.read_meta(d11, "codex", "tiny3", "M2") or {}
    f11 = os.path.join(C.map_dir(d11, "codex", "tiny3", "M2"), "features")
    check("S11 skipped M2 runs no map: meta has `skipped` and rc null, empty features/, no attempt files, 6 map calls only, not a failure for the stop rule",
          r11.returncode == 0 and m11.get("rc") is None and "no leaf directory qualifies" in str(m11.get("skipped")) and int(open(d11 + "-counter").read()) == 6
          and os.listdir(f11) == [] and not glob.glob(os.path.join(d11, "attempts", "codex", "tiny3", "M2*")) and C.load_active_features(d11, "codex", "tiny3", "M2") is None
          and not os.path.exists(os.path.join(d11, "stopped-codex-tiny3.json")), (m11, r11.stderr[-300:]))
    t11 = os.path.join(d11, "t.json")
    C.write_transitions(t11, arms=["codex"], repos=["tiny3"])
    rp11 = os.path.join(d11, "replay.jsonl")
    run_replay(tmp, os.path.join(d11, "maps"), t11, rp11)
    os.makedirs(os.path.join(d11, "pairs"))
    subprocess.run([PY, os.path.join(X2, "build_pairs.py"), "--data", d11, "--replay", rp11, "--out", os.path.join(d11, "pairs")], check=True, capture_output=True)
    k11, c11 = C.read_jsonl(os.path.join(d11, "pairs", "key.jsonl")), C.read_jsonl(os.path.join(d11, "pairs", "control-key.jsonl"))
    assert run_label(d11, os.path.join(d11, "pairs"), truth_labels(k11, lambda k: "same_concern", c11)).returncode == 0
    res11 = A.analyze(d11, rp11, os.path.join(d11, "pairs"))
    t6rows = [x for x in C.read_jsonl(rp11) if x["transition"] == "T6"]
    chain11 = [x for x in C.read_jsonl(os.path.join(d11, "chain.jsonl")) if x["step"] == "T6"]
    a11 = res11["arms"]["codex"]
    check("S11 replay rows, chain step, coverage, pairs, label distribution and labelling all exclude the skipped T6; class gate still satisfied (4 remap, T5 mutation)",
          t6rows and all(x["status"] == "unmeasured" and "mutation skipped" in x["reason"] for x in t6rows)
          and chain11 and all(x["status"] == "unmeasured" and "mutation skipped" in x["reason"] for x in chain11)
          and "T6" not in {k["transition"] for k in k11} and "T5" in {k["transition"] for k in k11}
          and "unmeasured" in a11["coverage"]["tiny3"]["T6"] and "measured" not in a11["coverage"]["tiny3"]["T6"]
          and sum(a11["label_distribution"]["tiny3"]["by_class"]["mutation"].values()) == len([k for k in k11 if k["transition"] == "T5"])
          and a11["decision"]["class_gate"]["tiny3"]["mutation"]["measured"] == 1 and a11["decision"]["class_gate"]["tiny3"]["remap"]["measured"] == 4
          and a11["decision"]["tripped_gate"] != "class gate", (t6rows[:1], chain11[:1], a11["decision"]["class_gate"]))
    inv11 = {m_["map"]: m_ for m_ in res11["inventory"]["maps"]}
    check("S11 inventory: the skipped M2 is `skipped` (not `failed`), carries its reason, M1 is ok; the rendered inventory lists it among retried/failed/skipped",
          inv11["M2"]["status"] == "skipped" and "no leaf directory qualifies" in str(inv11["M2"]["skipped"]) and inv11["M1"]["status"] == "ok" and all(m_["status"] != "failed" for m_ in inv11.values())
          and '"status": "skipped"' in A.render(res11), inv11["M2"])

    # N5/N13d: label sessions run with the scrubbed environment; codex exec events are recorded and reported
    lpd = os.path.join(tmp, "labelenv")
    os.makedirs(lpd)
    with open(os.path.join(lpd, "blinded.jsonl"), "w") as f:
        for pid in ("p1", "p2"):
            side_ = {"title": "T " + pid, "summary": "S", "kind": "module", "entrypoints": [], "owned_files": ["a/b.py"], "owned_files_total": 1}
            f.write(json.dumps({"pair_id": pid, "state": {"left": side_, "right": side_, "shared_files_total": 1, "shared_files": ["a/b.py"]}}) + "\n")
    elog = os.path.join(tmp, "label-envlog.txt")
    lenv = dict(os.environ, X2_FAKE_ENVLOG=elog, X2_FAKE_EXEC="1", **poison)
    lenv["PATH"] = bins + os.pathsep + os.environ["PATH"]
    lenv.pop("X2_LABEL_STUB", None)
    lr = subprocess.run([PY, os.path.join(X2, "label.py"), "--pairs", lpd, "--workers", "1"], env=lenv, capture_output=True, text=True)
    seen = sorted({n for n in open(elog).read().split("\n") if n})
    cred = subprocess.run([go_bin(tmp, "credcheck", "x")], input="\n".join(seen) + "\n", capture_output=True, text=True).stdout
    leaked_l = [line.split("\t")[0] for line in cred.splitlines() if line.endswith("\t1")]
    calls_l = {c["pair_id"]: c for c in C.read_jsonl(os.path.join(lpd, "calls.jsonl"))}
    ag = A.agreement(lpd)
    check("E label sessions (claude and codex runners, real code path, fake CLIs on PATH) see no credential name from a poisoned parent environment and keep non-credentials",
          lr.returncode == 0 and len(seen) > 5 and not leaked_l and "X2_PLAIN" in seen and "CLAUDE_CODE_OAUTH_TOKEN" in seen and "ANTHROPIC_API_KEY" not in seen
          and "OPENAI_API_KEY" not in seen and "OPENAI_BASE_URL" not in seen, (lr.stderr[-300:], leaked_l, len(seen)))
    check("C label calls record the served model (the label sessions' own flags select the isolated/overridden fake models), codex exec events per call, session accounting (2 sessions, 2 pairs with an exec event) and an undefined C-equals-A share without adjudications",
          all(c["A"]["model"] == "fake-claude-model-overridden" and c["B"]["model"] == "fake-codex-model-isolated" and c["B"]["exec_events"] == 1 for c in calls_l.values()) and len(calls_l) == 2
          and ag["codex_sessions"] == 2 and ag["pairs_with_exec_event"] == 2 and ag["adjudication_rate"] == 0.0 and ag["adjudicated_c_model_equals_a_of"] == 0
          and ag["adjudicated_c_model_equals_a_share"] is None, (calls_l, ag))


if __name__ == "__main__":
    main()
