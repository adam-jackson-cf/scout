#!/usr/bin/env python3
"""Per-round threshold tuning (PREREGISTRATION section 10 step 5; thresholds.json `tuning`). No model call.

Dev rows = gold-labelled V-sem scores of rounds < r that have gold (never the frozen rounds 8-10) plus the round-0 D and P units. Round r's own rows are written
to the CSVs as split `test` and are never given to the tuning script (it is run on a dev-only CSV; the dev set is built from rounds < r only, asserted).
Decisions (score, label):
  merge.covered     score = 1 - p_true, label 1 = gold none; the tuned score threshold s maps to t_cov = 1 - s (compose: none iff 1 - p_true >= s; s is stored as
                    `score_threshold`)
  merge.which_area  Choice rows for EVERY gold-labelled cluster whose Noul passed t_cov (1 - p_true < s of the covered decision; gold-none clusters included, so a false
                    assignment lowers precision); score = mean top confidence over the asked permutations; label 1 = all asked permutations agree on one option AND that
                    option is acceptable (always 0 for a gold-none cluster)
  merge.area_covers per (cluster, area) pair from the fan-out requests; label 1 = area in acceptable (case_id `<case_id>@<area_id>`)
Rules (thresholds.json tuning): min(n_pos, n_neg) < 30 -> provisional 0.7, no tuning; 30..50 -> tune with --repeats 5 --conservative, status tuned-provisional;
> 50 -> tune, status `ready` only if the script found a qualifying threshold in every fold, the fold thresholds agree within 0.05, no 0.5 replacement was made and the
composed dev minimums are met, else tuned-provisional. Always `--seed 20261029 --folds 5`; minimums and beta from the decision entry of thresholds.json. A tuned value of
exactly 0.5 is replaced by the next grid value in score space BELOW it for merge.covered (toward `none`, the cheaper error) and ABOVE it for the other decisions
(recorded); no threshold meeting the minimums keeps 0.7 and records `no_threshold_meets_minimums`.
Composed-rule check (R1-10, R4-13 (4)): after tuning, the PRIMARY arm (x2d.primary_arm: V-sem composes merge.covered + merge.which_area, fan-out composes merge.area_covers) is
composed at the tuned thresholds over the dev rounds' units and scored against gold with the verdict's definitions (`composed_dev`): P_assign (assigned-correct / assigned) and R_assign
(= R_assign_res: assigned-correct / (gold-covered AND resolved)) against x2d.MINIMUMS, AND the dev unresolved rate (abstained / dev clusters) <= x2d.UNRESOLVED_RATE_MAX. A value with a
denominator below its bound (or no assignment at all, or an unresolved rate over the cap) makes the tuned decisions of that arm revert to provisional 0.7 with
`no_threshold_meets_minimums` (a minimum with no denominator, e.g. no dev unit, is recorded as unevaluated and does not fail). P_new / R_new (and R_assign over all covered clusters,
abstention = miss) are recorded in `composed_dev`, never enforced. Refinement and frozen rounds (r >= 1) also refuse before writing anything when the round's own jev/scores.jsonl is incomplete
over its window units (compose.missing_required with the catalogue's area ids; R4-11: re-run jev R). The v_7 re-score's header `input_sha256` must equal the digest of its units.jsonl (R4-06).
Frozen rounds (R1-10): before round 8 the dev rounds 0-7 are re-scored under v_7 (dev only) into D/rounds/<rr>/jev-v7/scores.jsonl (x2d.V7_RESCORE_SUB); round 8
tunes once on those rows (all one question version); rounds 9 and 10 reuse round 8's decisions unchanged (`reused_from`: 8). Refinement rounds 1-7 pool dev rows over the
versions run so far (the version column of the tuning CSV is blanked for the script when versions differ; scores.csv keeps `question_version`).
Writes D/rounds/R/{scores.<question>.csv, scores.csv, thresholds.json}; refuses when D/rounds/R/thresholds.json exists. `--round 0` writes the
all-provisional file (round 0 tunes nothing). The tuning script (x2d.TUNE_SCRIPT) is run with `python3 -B` and its SHA-256 recorded (and pinned in signoff.json).

Usage: tune.py --round R      (X2D_SKIP_SIGNOFF=1: verify_pins(require_signoff=False), validation only)
"""
import argparse
import csv
import json
import os
import subprocess
import sys
import tempfile

import x2d
import compose

QUESTIONS = ("merge.covered", "merge.which_area", "merge.area_covers")
COLUMNS = ["case_id", "label", "score", "model", "question_version", "split", "stratum", "fallback"]
PROVISIONAL = 0.7
VARIANT = "V-sem"
ARM_QUESTIONS = {"V-sem": ("merge.covered", "merge.which_area"), "fan-out": ("merge.area_covers",)}
REUSE_FROM = x2d.FROZEN_ROUNDS[0]  # rounds 9 and 10 reuse round 8's thresholds


def decision_entry(name):
    th = json.load(open(os.path.join(x2d.HERE, "thresholds.json")))
    return next(d for d in th["decisions"] if d["question_id"] == name)


def _gold(rr, data):
    if not os.path.exists(os.path.join(x2d.round_dir(rr, data), "gold.jsonl")):
        return None
    import gold as G
    return G.load_gold(rr, data)


def _scores_path(rr, data, sub):
    return os.path.join(x2d.round_dir(rr, data), sub, "scores.jsonl")


def _dev_units(rr, g, rd):
    """Units admitted to tuning / composed dev: gold-labelled, never control or audit; round 0: the D and P sets, every later round: `set == "window"` only (R2-13)."""
    return [u for u in x2d.read_jsonl(os.path.join(rd, "units.jsonl")) if not u.get("control") and u["case_id"] in g
            and (u.get("set") in ("D", "P") if rr == 0 else u.get("set") == "window")]


def round_rows(rr, data, split, sub="jev"):
    """-> {question: [csv row dict]} for round rr's gold-labelled V-sem units from D/rounds/rr/<sub>/scores.jsonl (round 0: the D and P units only; later rounds: set == "window"
    only; controls and audit units never); {} when the round has no gold. merge.which_area rows are produced for every gold-labelled cluster that has a merge.covered answer AND
    every required permutation answered (compose.required_perms: permutation 2 iff x2d.flip_selected(cluster_id) or the scores header's flip_all) and carry `cov_score`
    (= 1 - p_true) for `choice_rows` to filter on the covered threshold; they are not CSV columns."""
    g = _gold(rr, data)
    if g is None:
        return {}
    rd = x2d.round_dir(rr, data)
    units = _dev_units(rr, g, rd)
    srows = x2d.read_jsonl(_scores_path(rr, data, sub))
    flip_all = compose.header_flip_all(srows)
    ans, _ = compose._index(srows)
    out = {q: [] for q in QUESTIONS}
    for u in units:
        cid, gr = u["case_id"], g[u["case_id"]]
        if not isinstance(gr.get("covered"), bool):
            continue
        acc = set(gr.get("acceptable") or []) if gr["covered"] else set()
        meta = lambda r: {"model": r.get("model", ""), "question_version": r.get("question_version", ""), "split": split,
                          "stratum": u.get("stratum", ""), "fallback": 0}
        cov = ans.get((cid, "merge.covered", VARIANT, 1, None))
        if cov is not None:
            out["merge.covered"].append({"case_id": cid, "label": int(not gr["covered"]), "score": 1.0 - cov["p_true"], **meta(cov)})
            need = compose.required_perms(u, flip_all)
            asked = [ans[k] for k in ((cid, "merge.which_area", VARIANT, p, None) for p in need) if k in ans]
            if len(asked) == len(need):  # an incomplete permutation set is never agreement
                same = len({r["choice"] for r in asked}) == 1
                out["merge.which_area"].append({"case_id": cid, "label": int(same and asked[0]["choice"] in acc),
                                                "score": sum(r["confidence"] for r in asked) / len(asked), "cov_score": 1.0 - cov["p_true"], **meta(asked[0])})
        for k, r in sorted(ans.items(), key=lambda kv: str(kv[0])):
            if k[:3] == (cid, "merge.area_covers", VARIANT):
                out["merge.area_covers"].append({"case_id": "%s@%s" % (cid, k[4]), "label": int(k[4] in acc), "score": r["p_true"], **meta(r)})
    return out


def choice_rows(rows, s_cov):
    """merge.which_area rows whose Noul passed the covered threshold (compose: none iff 1 - p_true >= s; so passed iff cov_score < s)."""
    return [r for r in rows if r["cov_score"] < s_cov - 1e-12]


def _header(path):
    return next((r for r in x2d.read_jsonl(path) if r.get("type") == "header"), None)


V7_QUESTIONS = ("merge.covered", "merge.which_area", "merge.area_covers")


def _check_round_scores(r, data):
    """R4-11 (r >= 1): round r's own jev/scores.jsonl must be complete over the round's window units (compose.missing_required with the round's area ids: merge.covered, every required
    merge.which_area permutation, every catalogue area of a fan-out): the driver writes its header before any request and exits 4 on a cost stop, and a partial file must not become the
    round's recorded test rows. Refuses (SystemExit) before anything is written."""
    rd = x2d.round_dir(r, data)
    p = _scores_path(r, data, "jev")
    rows = x2d.read_jsonl(p)
    units = [u for u in x2d.read_jsonl(os.path.join(rd, "units.jsonl")) if not u.get("control") and u.get("set") == "window"]
    fan = any(row.get("question") == "merge.area_covers" for row in rows)  # the catalogue is read only once the file has a fan-out row (compose.missing_required's own rule)
    miss = compose.missing_required(rows, units, area_ids=_round_area_ids(r, data) if fan else None)
    if miss:
        sys.exit("error: %s is incomplete for %d required request(s) of the window units (e.g. %s): re-run jev %d" % (p, len(miss), miss[:3], r))


def _round_area_ids(rr, data):
    """Area ids of the catalogue round rr's main Jev run used: round 0 the manifest's `main` catalogue, later rounds the study catalogue (R3-18)."""
    p = x2d.catalogue_path(data)
    if rr == 0:
        man = {e["name"]: e for e in json.load(open(os.path.join(x2d.round_dir(0, data), "jev-in", "manifest.json")))}["main"]
        p = man["catalogue"]
    return [a["area_id"] for a in json.load(open(p))["areas"]]


def _check_v7_scores(rounds, data, sub):
    """R2-02, R3-18: each dev round's <sub>/scores.jsonl must carry a header whose model and catalogue digest equal those of rounds/7/jev/scores.jsonl's header and whose `versions` are
    exactly the three question versions of the sem and fan arms (merge.covered, merge.which_area, merge.area_covers), each equal to round 7's, and must be complete:
    compose.missing_required (merge.covered, every required merge.which_area permutation, every area of the round's catalogue for the fan-out) is empty over the round's dev units;
    refuses (SystemExit) empty, headerless, mismatched or incomplete files."""
    ref_p = _scores_path(7, data, "jev")
    ref = _header(ref_p)
    if ref is None:
        sys.exit("error: %s has no header row (the v_7 header reference)" % ref_p)
    for rr in rounds:
        p = _scores_path(rr, data, sub)
        rows = x2d.read_jsonl(p)
        if not rows:
            sys.exit("error: %s is empty" % p)
        h = _header(p)
        if h is None:
            sys.exit("error: %s has no header row" % p)
        hv, rv = h.get("versions"), ref.get("versions") or {}
        bad = [k for k in ("model", "catalogue_sha256") if h.get(k) != ref.get(k)]
        want_keys = {q + "/V-sem" for q in V7_QUESTIONS}
        if not isinstance(hv, dict) or set(hv) != want_keys or any(rv.get(k) != hv.get(k) for k in want_keys):
            bad.insert(0, "versions")
        if bad:
            sys.exit("error: %s header differs from %s in %s (re-score round %d with the round-7 versions, model and catalogue)" % (p, ref_p, ", ".join(bad), rr))
        up = os.path.join(os.path.dirname(p), "units.jsonl")  # R3-05, R4-06: the re-score's units carry round 7's excerpt_lines (build_units.rerender_units); the header's input_sha256 (the SHA-256 of the -in file the driver read) must equal this file
        if not os.path.isfile(up) or h.get("input_sha256") != x2d.sha256_file(up):
            sys.exit("error: %s: the v_7 units file is not the one the re-score read (header input_sha256 %s, %s digests to %s)" % (
                p, h.get("input_sha256"), up, x2d.sha256_file(up) if os.path.isfile(up) else "missing"))
        want_lines = int(x2d.round_settings(7, data)["excerpt_lines"])
        got_lines = {u.get("excerpt_lines") for u in x2d.read_jsonl(up)}
        if got_lines != {want_lines}:
            sys.exit("error: %s rows carry excerpt_lines %s, round 7's setting is %d (the v_7 re-score must re-render the excerpt under v_7)" % (up, sorted(map(str, got_lines)), want_lines))
        if not any(r.get("question") == "merge.area_covers" for r in rows):
            sys.exit("error: %s has no merge.area_covers rows (the re-score runs the sem and fan arms): incomplete re-score" % p)
        g = _gold(rr, data)
        miss = compose.missing_required(rows, _dev_units(rr, g, x2d.round_dir(rr, data)), area_ids=_round_area_ids(rr, data))
        if miss:
            sys.exit("error: %s is incomplete for %d required request(s) of the dev units (e.g. %s): incomplete re-score" % (p, len(miss), miss[:3]))


def dev_rounds(r, data, sub="jev"):
    rounds = [rr for rr in range(0, r) if rr not in x2d.FROZEN_ROUNDS and _gold(rr, data) is not None]
    assert all(rr < r for rr in rounds), "dev set must be built from rounds < r only"
    assert not set(rounds) & set(x2d.FROZEN_ROUNDS)
    if sub != "jev":  # v_7 rows: every dev round must have been re-scored
        missing = [rr for rr in rounds if not os.path.exists(_scores_path(rr, data, sub))]
        if missing:
            sys.exit("error: dev rounds %s lack %s/scores.jsonl (re-score rounds 0-7 under v_7 before round 8, PREREGISTRATION 10)" % (missing, sub))
        _check_v7_scores(rounds, data, sub)
    return rounds


def run_script(rows, script, entry, repeats, conservative):
    versions = {r["question_version"] for r in rows}
    keep = next(iter(versions)) if len(versions) == 1 else ""  # the script reports one group per (model, version); mixed versions are pooled
    with tempfile.TemporaryDirectory() as d:
        p = os.path.join(d, "dev.csv")
        with open(p, "w", newline="") as f:
            w = csv.DictWriter(f, COLUMNS)
            w.writeheader()
            for r in rows:
                w.writerow({**{k: r[k] for k in COLUMNS}, "model": x2d.JEV_MODEL, "question_version": keep, "split": "dev"})
        cmd = ["python3", "-B", script, p, "--min-precision", str(entry["min_precision"]), "--min-recall", str(entry["min_recall"]),
               "--beta", str(entry["beta"]), "--folds", "5", "--seed", str(x2d.SEED), "--repeats", str(repeats), "--json"]
        cmd += ["--conservative"] if conservative else []
        res = subprocess.run(cmd, capture_output=True, text=True)
    if res.returncode != 0:
        raise RuntimeError("tune_threshold.py failed: " + (res.stderr or res.stdout)[-500:])
    return json.loads(res.stdout)[0]


def next_half(rows, below):
    """Next grid value (the script's candidate grid: the lowest score and midpoints of adjacent distinct scores) strictly on one side of 0.5 in SCORE space."""
    u = sorted({r["score"] for r in rows})
    cands = u[:1] + [(lo + hi) / 2 for lo, hi in zip(u, u[1:])]
    side = [c for c in cands if (c < 0.5 if below else c > 0.5)]
    if not side:
        return None
    return side[-1] if below else side[0]


def decide(question, rows, script):
    entry = decision_entry(question)
    pos = sum(r["label"] for r in rows)
    neg = len(rows) - pos
    d = {"threshold": PROVISIONAL, "status": "provisional", "n_pos": pos, "n_neg": neg, "source": "provisional-0.7", "fold_report": None}
    if question == "merge.covered":
        d["score_threshold"] = 1.0 - PROVISIONAL
    m = min(pos, neg)
    if m < 30:
        return d
    rep = run_script(rows, script, entry, 5 if m <= 50 else 1, m <= 50)
    d["fold_report"] = rep
    t = rep.get("shipped_threshold")
    if t is None:
        d["note"] = "no_threshold_meets_minimums"
        return d
    if t == 0.5:
        below = question == "merge.covered"  # covered: toward `none` (R1-25)
        nt = next_half(rows, below)
        if nt is None:
            d["note"] = "tuned_0.5_no_grid_value_%s" % ("below" if below else "above")
            return d
        d["note"] = "tuned_value_was_0.5_took_next_grid_value_%s" % ("below" if below else "above")
        t = nt
    d["source"] = "tuned"
    d["score_threshold"] = t
    d["threshold"] = (1.0 - t) if question == "merge.covered" else t
    ts = rep.get("fold_thresholds") or []
    clean = not rep.get("failed_folds") and rep.get("tuned_folds") == rep.get("total_folds") and (not ts or max(ts) - min(ts) <= 0.05)
    d["status"] = "ready" if (m > 50 and clean and "note" not in d) else "tuned-provisional"
    return d


def composed_dev(rounds, data, decisions, arm, sub="jev"):
    """The composed rule of `arm` at the tuned decisions over the dev rounds' units (round 0: D and P) against gold, with the verdict's definitions (R4-13 (4)): P_assign and
    R_assign (= rates()["R_assign_res"]: assigned-correct / (gold-covered AND resolved)) against x2d.MINIMUMS, AND the dev unresolved rate (abstained / all dev clusters) <= x2d.UNRESOLVED_RATE_MAX.
    P_new / R_new (and R_assign over all covered clusters) are recorded under `P_new`, `R_new`, `R_assign_all`, never enforced.
    -> {"arm", "n", "P_assign", "R_assign", "R_assign_all", "P_new", "R_new" (each {"k","n","value"}), "unresolved_rate" {"k","n","value","max"}, "minimums_met": bool|None,
    "failed": [...], "unevaluated": [...]}."""
    import analysis_x2d as A
    thresholds = {"decisions": decisions}
    outs = []
    for rr in rounds:
        g = _gold(rr, data)
        rd = x2d.round_dir(rr, data)
        units = _dev_units(rr, g, rd)
        srows = x2d.read_jsonl(_scores_path(rr, data, sub))
        rows = compose.compose(srows, units, thresholds, arm, flip_all=compose.header_flip_all(srows))
        outs += [A.score_outcome(row, g[row["case_id"]]) for row in rows]
    c = A.counts_for(outs)
    rt = A.rates(c)
    res = {"arm": arm, "n": rt["n"], "failed": [], "unevaluated": []}
    for name, key in (("P_assign", "P_assign"), ("R_assign", "R_assign_res"), ("R_assign_all", "R_assign"), ("P_new", "P_new"), ("R_new", "R_new")):
        r_ = rt[key]
        res[name] = {"k": r_["k"], "n": r_["n"], "value": r_["value"]}
        if name not in x2d.MINIMUMS:  # recorded, not enforced (R4-13: ~10 new-area cases cannot veto; all-covered recall counts abstention as a miss)
            continue
        if r_["n"] == 0:
            if rt["n"]:
                res["failed"].append(name + " (no assignment / no denominator)")
            else:
                res["unevaluated"].append(name)
        elif r_["value"] < x2d.MINIMUMS[name]:
            res["failed"].append(name)
    u = rt["unresolved_rate"]
    res["unresolved_rate"] = {"k": u["k"], "n": u["n"], "value": u["value"], "max": x2d.UNRESOLVED_RATE_MAX}
    if u["n"] and u["value"] > x2d.UNRESOLVED_RATE_MAX:
        res["failed"].append("unresolved_rate")
    res["minimums_met"] = (not res["failed"]) if rt["n"] else None
    return res


def write_csv(path, rows):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "x", newline="") as f:
        w = csv.DictWriter(f, COLUMNS)
        w.writeheader()
        for r in rows:
            w.writerow({k: r[k] for k in COLUMNS})


def _tune_dev(rounds, data, script, sub):
    """Tune the three questions on dev rows of `rounds` (scores under `sub`); -> (decisions, dev rows, composed_dev|None)."""
    dev = {q: [] for q in QUESTIONS}
    for rr in rounds:
        for q, rows in round_rows(rr, data, "dev", sub).items():
            dev[q] += rows
    covered = decide("merge.covered", dev["merge.covered"], script)
    s_cov = covered["score_threshold"]
    dev["merge.which_area"] = choice_rows(dev["merge.which_area"], s_cov)
    decisions = {"merge.covered": covered, "merge.which_area": decide("merge.which_area", dev["merge.which_area"], script),
                 "merge.area_covers": decide("merge.area_covers", dev["merge.area_covers"], script)}
    arm = x2d.primary_arm(data)
    comp = None
    if rounds:
        comp = composed_dev(rounds, data, {q: {"threshold": d["threshold"], **({"score_threshold": d["score_threshold"]} if "score_threshold" in d else {})} for q, d in decisions.items()}, arm, sub)
        tuned = [q for q in ARM_QUESTIONS[arm if arm in ARM_QUESTIONS else "V-sem"] if decisions[q]["source"] == "tuned"]
        if tuned and comp["minimums_met"] is False:
            for q in tuned:
                prov = {"threshold": PROVISIONAL, "status": "provisional", "n_pos": decisions[q]["n_pos"], "n_neg": decisions[q]["n_neg"], "source": "provisional-0.7",
                        "fold_report": decisions[q]["fold_report"], "note": "no_threshold_meets_minimums (composed dev rule: %s)" % ", ".join(comp["failed"]),
                        "tuned_but_rejected": {k: decisions[q][k] for k in ("threshold", "score_threshold", "status") if k in decisions[q]}}
                if q == "merge.covered":
                    prov["score_threshold"] = 1.0 - PROVISIONAL
                decisions[q] = prov
            # R5-11: the same figures at the provisional decisions (record only; tune_round writes it as `composed_dev_provisional` beside `composed_dev`)
            comp["composed_dev_provisional"] = composed_dev(rounds, data, {q: {"threshold": d["threshold"], **({"score_threshold": d["score_threshold"]} if "score_threshold" in d else {})} for q, d in decisions.items()}, arm, sub)
        elif tuned and comp["minimums_met"] is None:
            for q in tuned:
                decisions[q]["status"] = "tuned-provisional"
    return decisions, dev, comp


def tune_round(r, data=None, script=None):
    script = script or x2d.TUNE_SCRIPT
    rd = x2d.round_dir(r, data)
    th_p = os.path.join(rd, "thresholds.json")
    if os.path.exists(th_p):
        sys.exit("refusing: %s exists (append-only)" % th_p)
    if r >= 1:
        _check_round_scores(r, data)  # R4-11: before anything is written
    comp, comp_prov, reused, sub = None, None, None, "jev"
    if r == 0:
        dev = {q: [] for q in QUESTIONS}
        rounds = []
        decisions = {q: decide(q, [], script) for q in QUESTIONS}
    elif r in x2d.FROZEN_ROUNDS and r != REUSE_FROM:
        src = os.path.join(x2d.round_dir(REUSE_FROM, data), "thresholds.json")
        if not os.path.exists(src):
            sys.exit("error: %s missing: round %d reuses round %d's thresholds (tune round %d first)" % (src, r, REUSE_FROM, REUSE_FROM))
        prior = json.load(open(src))
        decisions, rounds, reused = prior["decisions"], prior["dev_rounds"], REUSE_FROM
        dev = {q: [] for q in QUESTIONS}
        comp = prior.get("composed_dev")
        comp_prov = prior.get("composed_dev_provisional")
    else:
        if not os.path.exists(script):
            sys.exit("error: tuning script %s not found" % script)
        sub = x2d.V7_RESCORE_SUB if r == REUSE_FROM else "jev"
        rounds = dev_rounds(r, data, sub)
        decisions, dev, comp = _tune_dev(rounds, data, script, sub)
        comp_prov = comp.pop("composed_dev_provisional", None) if comp else None
    test = round_rows(r, data, "test")  # written for the analysis only; never given to the tuning script
    if test:
        s_cov = decisions["merge.covered"]["score_threshold"] if decisions["merge.covered"].get("score_threshold") is not None else 1.0 - decisions["merge.covered"]["threshold"]
        test["merge.which_area"] = choice_rows(test["merge.which_area"], s_cov)
    csvs = [os.path.join(rd, "scores.%s.csv" % q) for q in QUESTIONS] + [os.path.join(rd, "scores.csv")]
    x2d.refuse_if_exists(*csvs)
    allrows = []
    for q in QUESTIONS:
        rows = dev[q] + test.get(q, [])
        write_csv(os.path.join(rd, "scores.%s.csv" % q), rows)
        allrows += [{**row, "case_id": row["case_id"] + "|" + q} for row in rows]
    write_csv(os.path.join(rd, "scores.csv"), allrows)
    out = {"round": r, "decisions": decisions, "dev_rounds": rounds, "seed": x2d.SEED, "folds": 5, "skill_script": script,
           "skill_script_sha256": x2d.sha256_file(script) if os.path.exists(script) else None, "primary_arm": x2d.primary_arm(data),
           "dev_scores_sub": sub, "reused_from": reused, "composed_dev": comp, "composed_dev_provisional": comp_prov}
    x2d.write_json_new(th_p, out)
    return out


def main(argv=None):
    ap = argparse.ArgumentParser()
    ap.add_argument("--round", type=int, required=True)
    a = ap.parse_args(argv)
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pin check failed:\n" + "\n".join(bad))
    out = tune_round(a.round)
    for q, d in out["decisions"].items():
        print("%s: threshold %.4f status %s (%s) n_pos %d n_neg %d%s" % (q, d["threshold"], d["status"], d["source"], d["n_pos"], d["n_neg"],
                                                                  " [" + d["note"] + "]" if d.get("note") else ""))
    if out.get("composed_dev"):
        print("composed dev (%s): %s" % (out["composed_dev"]["arm"], json.dumps({k: out["composed_dev"][k] for k in ("P_assign", "R_assign", "unresolved_rate", "P_new", "R_new", "minimums_met")}, sort_keys=True)))


if __name__ == "__main__":
    main()
