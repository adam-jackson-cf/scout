#!/usr/bin/env python3
"""X2c implementation validation (CLAUDE.md experiment rule 4). Synthetic data, a stub Clawpatch, fake CLIs and a stub Jev endpoint only: no model call, no real data.

X2c reuses X2's frozen analysis, labelling and replay by import, so validation has these parts:
  P  provenance and drift: the reused X2 files equal their pins; the copied collector equals X2's collector plus exactly the pinned diff; the Jev pin, the question version, the
     question text against GUIDE.md; Clawpatch-side flags (codex.ts, claude.ts, cli.ts).
  J  the Jev scorer (a thin driver over internal/decisions) against a stub endpoint: request shape, two orders, failures, resume, key hygiene, alias refusal.
  S  planted effect and planted nulls through the whole pipeline (maps -> replay -> candidates -> labels -> Jev scores -> analysis) for BOTH methods: exact counts at every checked
     threshold, the Jaccard method failing where the Jev method adopts, the two-order average removing an order artefact, uninformative / adverse scorers and permuted labels
     reporting no effect, the gates (failed requests, model mismatch, missing cost, cost cap, coverage, served check, unexpected arm), the four-arm selection rule.
  C  the multi-arm collector on a stub Clawpatch and fake codex/claude: exact argv per arm, probes, versions record, refusals, then run.sh end to end.
  X  X2's own validation, run unchanged on the reused code.
Usage (repo root, toolchain active): .tools/venv/bin/python evaluation/x2c-mapper-comparison/validation/validate.py [--skip-x2 (development only; exits 2)]
"""
import difflib
import hashlib
import importlib.util
import json
import math
import os
import random
import re
import shutil
import subprocess
import sys
import tempfile

HERE = os.path.dirname(os.path.abspath(__file__))
X2C = os.path.dirname(HERE)
sys.path.insert(0, X2C)
sys.path.insert(0, HERE)
import x2c  # noqa: E402 (puts the frozen X2 directory on sys.path)
import analysis as A  # noqa: E402
import common as C  # noqa: E402
import analysis_x2c as AX  # noqa: E402
import jev_match as J  # noqa: E402
import stub_jev  # noqa: E402

ROOT, X2, PY = x2c.ROOT, x2c.X2_DIR, sys.executable
spec_ = importlib.util.spec_from_file_location("x2_validate", os.path.join(X2, "validation", "validate.py"))
V = importlib.util.module_from_spec(spec_)
spec_.loader.exec_module(V)  # X2's helpers (synthetic maps, runners, labels); its main() is not run
COUNT = [0]
DIFF_FILE = os.path.join(HERE, "collect-x2c.diff")
KEY = "sk-or-v1-" + "validation-stub-key-0123456789"  # a fake key, only ever given to the scorer subprocess and the stub


def check(name, cond, detail=""):
    COUNT[0] += 1
    V.check(name, bool(cond), detail)


def collect_diff():
    a = open(os.path.join(X2, "collect.py")).read().splitlines(keepends=True)
    b = open(os.path.join(X2C, "collect_x2c.py")).read().splitlines(keepends=True)
    return "".join(difflib.unified_diff(a, b, "x2/collect.py", "x2c/collect_x2c.py", n=1))


def squash(s):
    return re.sub(r"\s+", " ", s).strip()


# ---------------------------------------------------------------- P provenance and drift guards
def provenance_checks(tmp):
    check("P every reused X2 file (analysis, common, build_pairs, label, replay, GUIDE, X2 PREREGISTRATION, validate, stub, credcheck) has the digest X2's round-10 sign-off names", not x2c.verify_pins(), x2c.verify_pins())
    check("P X2's collector, the base of collect_x2c.py, has the signed-off digest", C.sha256_file(os.path.join(X2, "collect.py")) == x2c.PINNED_COLLECT_BASE)
    check("P collect_x2c.py equals X2's collect.py plus exactly the committed diff (validation/collect-x2c.diff): every X2 safeguard outside the diff is the X2 code",
          os.path.exists(DIFF_FILE) and open(DIFF_FILE).read() == collect_diff(), "regenerate validation/collect-x2c.diff only with a new council round")
    check("P arms: exactly the four preregistered arms with their explicit flags; the reference is luna-xhigh; the collection order is serpentine and covers each (arm, repo) once",
          list(x2c.ARMS) == ["luna-xhigh", "terra-none", "terra-low", "haiku"] and x2c.REFERENCE_ARM == "luna-xhigh"
          and x2c.arm_flags("terra-none") == ["--model", "gpt-6.1-terra", "--reasoning-effort", "none"] and x2c.arm_flags("haiku") == ["--model", "claude-haiku-5-5", "--reasoning-effort", "none"]
          and sorted(x2c.COLLECTION_ORDER) == sorted((a, r) for a in x2c.ARMS for r in ("fastapi", "axum")) and x2c.COLLECTION_ORDER[0][0] == "luna-xhigh" and x2c.COLLECTION_ORDER[-1][0] == "luna-xhigh")
    check("P probe argv per arm mirrors the provider invocation (codex `--model M -c model_reasoning_effort=\"E\"`; claude `--model M` and no `--effort` for `none`, `--effort low` for low/minimal)",
          x2c.codex_probe_args("terra-low", "T") == ["exec", "--cd", "T", "--sandbox", "read-only", "--skip-git-repo-check", "--model", "gpt-6.1-terra", "-c", 'model_reasoning_effort="low"', "-"]
          and x2c.claude_probe_tail("haiku") == ["--model", "claude-haiku-5-5"] and x2c.claude_effort_flag("low") == ["--effort", "low"] and x2c.claude_effort_flag("minimal") == ["--effort", "low"]
          and x2c.claude_effort_flag("none") == [])
    pt = C.patched_sources(["src/providers/codex.ts", "src/providers/claude.ts", "src/cli.ts", "src/agent-mapper.ts"])
    cl, cx, cli, am = pt["src/providers/claude.ts"], pt["src/providers/codex.ts"], pt["src/cli.ts"], pt["src/agent-mapper.ts"]
    check("P drift guards (Clawpatch as built): codex passes `--model` and `-c model_reasoning_effort`; claude passes `--model`, passes `--effort` for any effort but `none` and maps `minimal` to `low`; `map` takes "
          "--model/--reasoning-effort (none allowed); the agent-start line reports the model option",
          'args.push("--model", options.model);' in cx and 'args.push("-c", `model_reasoning_effort="${options.reasoningEffort}"`);' in cx
          and 'args.push("--model", options.model);' in cl and 'options.reasoningEffort !== "none"' in cl and 'args.push("--effort", claudeEffort(options.reasoningEffort))' in cl and 'reasoningEffort === "minimal"' in cl
          and '"reasoning-effort"' in cli and "<none|minimal|low|medium|high|xhigh>" in cli and 'model: options.providerOptions.model ?? "default"' in am and 'options.onProgress?.("agent-start"' in am)
    check("P the Jev pin is Scout's (internal/policy DefaultModel equals the preregistered typesafe/jev-1.14-20261014); a changed pin is caught", x2c.check_jev_pin() == [] and x2c.JEV_MODEL == "typesafe/jev-1.14-20261014")
    bad = os.path.join(tmp, "consent.go")
    open(bad, "w").write(open(x2c.GO_POLICY).read().replace(x2c.JEV_MODEL, "typesafe/jev-1.15-20270101"))
    check("P Jev pin drift guard names the changed constant", len(x2c.check_jev_pin(bad)) == 1 and "DefaultModel" in x2c.check_jev_pin(bad)[0])
    scorer = V.go_bin(tmp, "jev-bin", x2c.JEV_SCORER_PKG)
    ver = subprocess.run([scorer, "-print-version", "-model", x2c.JEV_MODEL], capture_output=True, text=True).stdout.strip()
    check("P the question version key of the scorer equals the preregistered constant (any edit to the question, its state shape or the pin changes it)", ver == x2c.JEV_QUESTION_VERSION, (ver, x2c.JEV_QUESTION_VERSION))
    guide = squash(open(os.path.join(X2, "GUIDE.md")).read())
    src = squash(open(os.path.join(X2C, "jev", "main.go")).read())
    crit = re.findall(r"^- \*\*(\w+)\*\*: (.*?)(?=\n- \*\*|\n\n)", open(os.path.join(X2, "GUIDE.md")).read(), re.S | re.M)
    rules = re.findall(r"^- (R\d) (.*?)(?=\n- R|\n\n)", open(os.path.join(X2, "GUIDE.md")).read(), re.S | re.M)
    qtext = guide[guide.index("Two features were produced"): guide.index("## Evidence shown")]
    check("P the scorer's question is the guide's: the four criteria word for word (GUIDE.md label definitions) and the guide's question sentence; rules R1-R7 present in the instructions in the guide's words",
          len(crit) == 4 and all(squash(t).replace("\\", "") in src for _l, t in crit) and len(rules) == 7 and all(squash(t).replace("`", "")[:40] in src.replace('" + "', "").replace("`", "") for _r, t in rules)
          and "do they describe the same concern (the same area of responsibility a maintainer would point to for a question like" in qtext and "do they describe the same concern (the same area of responsibility a maintainer would point to for a" in src.replace('" + "', ""),
          (len(crit), len(rules)))


# ---------------------------------------------------------------- planted data
EXTRA, VOL, VOL_J = V.EXTRA, V.VOL, V.VOL_J


def x2c_features(spec, tag, mid):
    tok = hashlib.sha1(tag.encode()).hexdigest()[:8]
    out = []
    if spec["kind"] == "unrelated":
        return [V.feat("reg_%s_%d" % (tag, i), "Concern %d %s" % (i, tok), ["u/%s/%d/f%d" % (tag, i, k) for k in range(10)]) for i in range(spec["n_reg"])]
    e = EXTRA[mid] if spec["kind"] == "planted" else 0
    for i in range(spec["n_reg"]):
        out.append(V.feat("reg_%s_%d" % (tag, i), "Concern %d %s" % (i, tok), ["base/%d/f%d" % (i, k) for k in range(10)] + ["ex/%s/%d/x%d" % (mid, i, k) for k in range(e)]))
    if spec["kind"] == "planted":
        for k in range(spec["n_red"]):  # redrawn: the same concern keeps two of ten files (Jaccard 2/18, below T_MIN)
            out.append(V.feat("red_%s_%d" % (tag, k), "Redrawn %d %s" % (k, tok), ["red/%d/f0" % k, "red/%d/f1" % k] + ["red/%s/%d/u%d" % (mid, k, u) for u in range(8)]))
        for j in range(spec["n_vol"]):  # volatile: generic files shared at Jaccard 0.33-0.45, a different concern
            s, u = VOL[j % len(VOL)]
            out.append(V.feat("vol_%s_%d" % (tag, j), "Volatile concern %d %s" % (j, tok), ["gen/%d/g%d" % (j, k) for k in range(s)] + ["vol/%s/%d/u%d" % (mid, j, k) for k in range(u)]))
        for k in range(spec["n_trap"]):  # trap: one generic file in common, different concerns (Jaccard 1/9)
            out.append(V.feat("trap_%s_%d" % (tag, k), "Trap %d %s" % (k, tok), ["trap/shared.py"] + ["trap/%s/%d/u%d" % (mid, k, u) for u in range(4)]))
    out.append(V.feat("heur_%s" % tag, "Heuristic group", ["heur/%s.py" % tag], source="python-source-group"))
    return out


def plant_data(data, arms, repos=("fastapi", "axum"), probe_effort=None, probe_model=None, flags_override=None, option_override=None):
    """arms: {arm: spec}; writes maps + versions files that look like collector output (probe-consistent, preregistered flags, spec["seconds"] per map)."""
    for arm, spec in arms.items():
        a = x2c.ARMS[arm]
        for repo in repos:
            for mid in C.MAP_IDS:
                tag = "%s-%s-%s" % (arm, repo, mid)
                V.write_map(data, arm, repo, mid, x2c_features(spec, tag, mid), 0, {
                    "schema": "x2c-map-v1", "arm": arm, "repo": repo, "map_id": mid, "provider": a["provider"], "map_model": a["model"], "map_reasoning_effort": a["effort"],
                    "map_flags": (flags_override or {}).get(arm, x2c.arm_flags(arm)), "provider_reported": a["provider"], "model_option_reported": (option_override or {}).get(arm, a["model"]),
                    "attempt_count": 1, "seconds_total": spec.get("seconds", 60.0), "seconds_last": spec.get("seconds", 60.0), "attempts": [{"started_utc": "2026-10-28T00:00:00Z"}]})
            pm, pe = (probe_model or {}).get(arm, a["model"]), (probe_effort or {}).get(arm, a["effort"])
            probe = {"model": pm, "reasoning_effort": pe, "error": None}
            with open(os.path.join(data, "tool-versions-%s-%s.json" % (arm, repo)), "w") as f:
                json.dump({"model_probe_before": probe, "model_probe_after": probe, "cli_versions_after": {a["provider"]: "c1"}, a["provider"]: "c1",
                           "workspace_config_presence": {".codex": False}, "x2c": {"map_flags": x2c.arm_flags(arm)}}, f)


def rule_pair(k):
    """The planted concern truth: regular and redrawn successors of the same index are the same concern; everything else (volatile, trap) is a different concern."""
    pf, nf = k["prev_feature_id"], k["next_feature_id"]
    fam = nf.split("_")[0]
    return "same_concern" if fam in ("reg", "red") and pf.split("_")[0] == fam and pf.rsplit("_", 1)[1] == nf.rsplit("_", 1)[1] else "different_concern"


def jev_env():
    return dict(os.environ, OPENROUTER_API_KEY=KEY)


def run_scorer(tmp, pairs, out, stub, extra=()):
    exe = V.go_bin(tmp, "jev-bin", x2c.JEV_SCORER_PKG)
    return subprocess.run([exe, "-in", os.path.join(pairs, "jev-input.jsonl"), "-out", out, "-endpoint", stub.url, "-workers", "8", *extra], env=jev_env(), capture_output=True, text=True)


def pipeline(tmp, name, arms, rule=rule_pair, mode="planted", stub_kw=None, plant_kw=None, perm_seed=None, jev=True):
    d = os.path.join(tmp, name)
    data, pairs = os.path.join(d, "data"), os.path.join(d, "pairs")
    os.makedirs(pairs)
    plant_data(data, arms, **(plant_kw or {}))
    trans = os.path.join(d, "transitions.json")
    C.write_transitions(trans, arms=list(arms))
    replay = os.path.join(d, "replay.jsonl")
    V.run_replay(tmp, os.path.join(data, "maps"), trans, replay)
    r = subprocess.run([PY, os.path.join(X2C, "build_candidates.py"), "--data", data, "--replay", replay, "--out", pairs], capture_output=True, text=True)
    assert r.returncode == 0, r.stderr + r.stdout
    keys, ctl = C.read_jsonl(os.path.join(pairs, "key.jsonl")), C.read_jsonl(os.path.join(pairs, "control-key.jsonl"))
    rl = rule
    if perm_seed is not None:
        pids = sorted({k["pair_id"] for k in keys})
        vals = [rule(next(k for k in keys if k["pair_id"] == p)) for p in pids]
        random.Random(perm_seed).shuffle(vals)
        perm = dict(zip(pids, vals))
        rl = lambda k: perm[k["pair_id"]]
    lr = V.run_label(d, pairs, V.truth_labels(keys, rl, ctl))
    assert lr.returncode == 0, lr.stderr
    out = {"d": d, "data": data, "pairs": pairs, "replay": replay, "scores": os.path.join(d, "scores.jsonl"), "keys": keys, "ctl": ctl}
    if jev:
        stub = stub_jev.Stub(x2c.JEV_MODEL, mode, log_path=os.path.join(d, "stub.log"), **(stub_kw or {})).start()
        sr = run_scorer(tmp, pairs, out["scores"], stub)
        stub.stop()
        out.update({"scorer": sr, "stub": stub})
    out["res"] = AX.analyze(data, replay, pairs, out["scores"]) if jev else None
    return out


def cell(res, arm, repo, pct):
    return res["arms"][arm]["tables"][repo]["decisive"][str(pct)]


def volc(n, t):
    return sum(1 for j in range(n) if VOL_J[j % len(VOL_J)] >= t)


PLANTED = {"kind": "planted", "n_reg": 40, "n_red": 8, "n_vol": 6, "n_trap": 4, "seconds": 600.0}


def scenario_checks(tmp):
    n_prior = 40 + 8 + 6 + 4
    p = pipeline(tmp, "S1", {"luna-xhigh": PLANTED})
    res = p["res"]
    check("S1 scorer exits 0 on the planted run; one candidate set serves both methods (every Jaccard match is a candidate: build_candidates verified it); the candidate set contains pairs the Jaccard rule never matches",
          p["scorer"].returncode == 0 and any(k["jaccard"] < C.T_MIN for k in p["keys"]) and any(k["jaccard"] >= C.T_MIN for k in p["keys"]), p["scorer"].stderr[-300:])
    ok = True
    for repo in ("fastapi", "axum"):
        for pct in (30, 45, 50, 75, 80, 85, 90):
            t = pct / 100
            cj, cv = cell(res, "luna-xhigh", repo, pct), cell(res, "luna-xhigh+jev", repo, pct)
            vol = volc(6, t)
            ok &= cj["matched"] == 6 * ((40 if pct <= 75 else 0) + vol) and cj["different"] == 6 * vol
            ok &= cv["matched"] == 6 * (48 if pct <= 85 else 40) and cv["different"] == 0 and cv["prev_areas"] == 6 * n_prior
    check("S1 planted effect, exact counts in both repositories at 0.30/0.45/0.50/0.75/0.80/0.85/0.90: Jaccard matches the 40 regular areas (and the volatile ones at <= 0.45, all spurious), the Jev method matches regular + redrawn "
          "(48 of 58) with no spurious merge up to 0.85 and only the regular ones at 0.90", ok)
    dj, dv = res["arms"]["luna-xhigh"]["decision"], res["arms"]["luna-xhigh+jev"]["decision"]
    check("S1 verdicts: the Jaccard cell is `fallback` (survival 40/58 below 4/5 where it is clean; spurious volatile merges where survival suffices) and the Jev cell is `adopt` t* = 0.85 (survival 48/58 >= 4/5, none spurious)",
          dj["verdict"] == "fallback" and dv["verdict"] == "adopt" and dv["threshold_pct"] == 85 and dv["pass_set_contiguous"], (dj["verdict"], dv))
    cj = res["x2c"]["cells"]
    sel = res["x2c"]["selection"]
    check("S1 cells and selection: Jaccard `not-adopted`; Jev `eligible` (confirmed across repositories, coverage ok, cost and time under the caps); outcome adopt luna-xhigh+jev, proposal names the arm, the method and tau 0.85",
          cj["luna-xhigh"]["state"] == "not-adopted" and cj["luna-xhigh+jev"]["state"] == "eligible" and cj["luna-xhigh+jev"]["confirmation"]["confirmed"] and sel["outcome"] == "adopt" and sel["cell"] == "luna-xhigh+jev"
          and "mapper luna-xhigh with the jev method" in sel["proposal"] and "tau = 0.85" in sel["proposal"], (cj["luna-xhigh+jev"], sel))
    jr = res["x2c"]["jev_requests"]
    check("S1 Jev accounting: requests = 2 x unique candidate pairs, billed cost and latency recorded, every reply from the pin, no failure; refresh seconds = map seconds (2 x 600) + Jev seconds",
          jr["answers"] == 2 * len({k["pair_id"] for k in p["keys"]}) and jr["failures"] == 0 and jr["billed_cost_usd"] > 0 and jr["mean_latency_s"] is not None and res["x2c"]["jev_check"]["ok"]
          and 1200 < cj["luna-xhigh+jev"]["refresh_seconds"] < 1210 and cj["luna-xhigh"]["refresh_seconds"] == 1200.0, (jr, cj["luna-xhigh+jev"]["refresh_seconds"]))
    text = AX.render(res)
    check("S1 rendered report carries the X2 tables, the cells and the SELECTION line", all(s in text for s in ("== X2c", "cell luna-xhigh+jev", "SELECTION: adopt", "not recorded: Clawpatch prints no tokens")), text[-500:])
    check("S1 the report states that `<arm>+jev` created/kept/retired and kept-inclusive survival are schema placeholders (unmeasured) and that the inventory's `not recorded` entries for `+jev` arms are the same placeholders",
          "NOTE `<arm>+jev` rows: created/kept/retired and kept-inclusive survival are schema placeholders" in text and "(Jev cells run no maps)" in text)
    sq = res["x2c"]["cells"]["luna-xhigh+jev"]["jev"]
    fm = res["x2c"]["family"]
    check("S1 sequential-runner and reporting items are computed and rendered: `seconds_sequential` (= the 8-worker figure x 8, the product runner being sequential), the share of answers with attempts > 1, failure rows by kind, the candidate ceiling, "
          "the area ratio, the family-wise sum of maxima over the eight cells, the per-(arm, repo) Jev refresh line and the X2b comparison",
          sq["seconds_sequential"] > sq["seconds"] > 0 and abs(sq["seconds_sequential"] - x2c.JEV_WORKERS * sq["seconds"]) < 1e-9 and res["x2c"]["jev_requests"]["attempts_gt1_share"] == 0 and res["x2c"]["jev_requests"]["failures_by_kind"] == {}
          and res["x2c"]["jev_requests"]["cost_unknown_rows"] == 0
          and 0 < res["x2c"]["candidate_ceiling"]["luna-xhigh"]["fastapi"] <= 1 and fm["cells_examined"] == 2 and fm["cells_eligible"] == 1 and 0 < fm["sum_of_maxima"]["0.1"] <= fm["sum_of_maxima"]["0.08"]
          and res["x2c"]["x2b_comparison"]["agree"] is True and all(s in text for s in ("seconds_sequential", "candidate ceiling", "area_ratio", "family-wise:", "X2b comparison:", "Jev code digests", "Jev per (arm, repo) refresh")), (sq, fm, res["x2c"]["x2b_comparison"]))
    fam_ok = True
    for cid_, v_ in fm["per_cell"].items():
        for p_ in AX.FAMILY_RATES:
            floor_ = [q_ for q_ in (AX.false_pass_probability(n_, p_) for n_ in v_["n_labelled"][str(C.GRID_PCT[0])].values()) if q_ is not None]
            fam_ok = fam_ok and v_["p_false_pass"][str(p_)] is not None and all(v_["p_false_pass"][str(p_)] >= q_ for q_ in floor_) and v_["at"][str(p_)]["pct"] in C.GRID_PCT and v_["at"][str(p_)]["repo"] in ("fastapi", "axum")
    check("S1 family-wise per cell: the figure is the maximum over thresholds and repositories, so it is at least the one-repository figure at the lowest grid threshold for each repository, and `at` names a grid threshold and a repository", fam_ok, fm["per_cell"])

    # ---- the order artefact: P(same) depends on which side is feature_a; the two-order average must remove it
    po = pipeline(tmp, "S2-order", {"luna-xhigh": PLANTED}, mode="order")
    _h, sc, answers, _f = J.read_scores(po["scores"])
    check("S2 order artefact: every pair's `ab` and `ba` answers differ (0.95 vs 0.05) and the averaged score is exactly 0.5 for every pair, so the artefact does not select pairs", answers and all(abs(J.pair_score(o)[0] - 0.5) < 1e-12 and abs(o["ab"][0] - o["ba"][0]) > 0.8 for o in sc.values()))

    # ---- planted nulls: an uninformative scorer, an adverse scorer, permuted labels, unrelated maps
    pc = pipeline(tmp, "S3-const", {"luna-xhigh": PLANTED}, mode="constant")
    pr = pipeline(tmp, "S3-rev", {"luna-xhigh": PLANTED}, mode="reverse")
    pp = pipeline(tmp, "S3-perm", {"luna-xhigh": PLANTED}, perm_seed=7)
    pu = pipeline(tmp, "S3-unrel", {"luna-xhigh": {"kind": "unrelated", "n_reg": 30, "seconds": 60.0}})
    vs = {n: q["res"]["arms"]["luna-xhigh+jev"]["decision"]["verdict"] for n, q in (("const", pc), ("rev", pr), ("perm", pp), ("unrel", pu))}
    vj = {n: q["res"]["arms"]["luna-xhigh"]["decision"]["verdict"] for n, q in (("perm", pp), ("unrel", pu))}
    check("S3 planted nulls report no effect: an uninformative Jev scorer (P = 0.5 for every pair), an adverse one (planted scores inverted) and permuted labels give `fallback` for the Jev method; permuted labels and "
          "unrelated maps give `fallback` for the Jaccard method; the selection is no-adoption in each case",
          set(vs.values()) == {"fallback"} and set(vj.values()) == {"fallback"} and all(q["res"]["x2c"]["selection"]["outcome"] == "no-adoption" for q in (pc, pr, pp, pu)), (vs, vj))

    # ---- gates: failed requests, model mismatch, missing cost, cost cap, served check, coverage, unexpected arm
    s1_titles = lambda rx: sorted({k["state"][s]["title"] for k in C.read_jsonl(os.path.join(tmp, "S1", "pairs", "jev-input.jsonl")) for s in ("left", "right") if re.match(rx, k["state"][s]["title"])})
    pf = pipeline(tmp, "S4-fail", {"luna-xhigh": PLANTED}, stub_kw={"fail_pairs": s1_titles(r"^Concern [3-7] ")})
    rf = pf["res"]["x2c"]
    _h, _s, _a, fails = J.read_scores(pf["scores"])
    check("S4 failed requests are never a pass: HTTP 500 for some features -> the scorer exits 3 with failure rows (kind http), those pairs stay unscored, the unscored share exceeds 1 %, the Jev check FAILS naming it (arm-only), "
          "the Jev cell is `inconclusive` and not citable, the Jaccard cell is unaffected, the outcome is inconclusive",
          pf["scorer"].returncode == 3 and fails and {f["kind"] for f in fails} == {"http"} and not rf["jev_check"]["ok"] and not rf["jev_check"]["reasons"]
          and any("unscored" in r for r in rf["jev_check"]["arm_reasons"].get("luna-xhigh", [])) and rf["jev_check"]["failures_by_kind"] == {"http": len(fails)}
          and rf["cells"]["luna-xhigh+jev"]["state"] == "inconclusive" and rf["cells"]["luna-xhigh"]["state"] == "not-adopted" and rf["selection"]["outcome"] == "inconclusive", (pf["scorer"].stderr[-200:], rf["jev_check"]))
    luna_titles = sorted({f["title"] for repo in ("fastapi", "axum") for mid in C.MAP_IDS for f in x2c_features(PLANTED, "luna-xhigh-%s-%s" % (repo, mid), mid) if re.match(r"^Concern [3-7] ", f["title"])})
    pa = pipeline(tmp, "S4-armonly", {"luna-xhigh": PLANTED, "terra-none": {"kind": "clean", "n_reg": 60, "seconds": 100.0}}, stub_kw={"fail_pairs": luna_titles})
    ra = pa["res"]["x2c"]
    check("S4 an unscored-share breach fails the Jev check of that arm only: failed requests on luna features leave `luna-xhigh+jev` inconclusive and `terra-none+jev` eligible; pin, version, model and cost failures stay global",
          pa["scorer"].returncode == 3 and not ra["jev_check"]["ok"] and not ra["jev_check"]["reasons"] and sorted(ra["jev_check"]["arm_reasons"]) == ["luna-xhigh"]
          and ra["cells"]["luna-xhigh+jev"]["state"] == "inconclusive" and ra["cells"]["terra-none+jev"]["state"] == "eligible" and ra["cells"]["terra-none"]["state"] == "eligible" and ra["selection"]["outcome"] == "adopt", (ra["jev_check"], {c: v["state"] for c, v in ra["cells"].items()}))
    pm = pipeline(tmp, "S4-model", {"luna-xhigh": PLANTED}, stub_kw={"model_for": {t: "typesafe/jev-9.9-rerouted" for t in s1_titles(r"^Concern 5 ")}})
    pm_h, _s, pm_a, pm_f = J.read_scores(pm["scores"])
    check("S4 a reply from another model than the pin is a `model-mismatch` failure (the product client's pin check), never an answer; the run reports it and the Jev check names the unscored pairs",
          pm["scorer"].returncode == 3 and {f["kind"] for f in pm_f} <= {"model-mismatch"} and all(a["model"] == x2c.JEV_MODEL for a in pm_a) and bool(pm_f))
    rm = pm["res"]["x2c"]["jev_check"]
    check("S4 a model-mismatch failure row is reported, not gated (PREREGISTRATION section 6): counted in foreign_model_failure_rows, no global Jev-check reason, its units unscored in that arm only",
          rm["foreign_model_failure_rows"] == len(pm_f) > 0 and not rm["reasons"] and sum(v["unscored"] for v in rm["per_arm"]["luna-xhigh"].values()) > 0, rm)
    pn = pipeline(tmp, "S4-nocost", {"luna-xhigh": PLANTED}, stub_kw={"no_cost": True})
    check("S4 a reply without billed cost makes the Jev check fail (cost is a gate input) and the Jev cell inconclusive", not pn["res"]["x2c"]["jev_check"]["ok"] and any("no billed cost" in r for r in pn["res"]["x2c"]["jev_check"]["reasons"])
          and pn["res"]["x2c"]["cells"]["luna-xhigh+jev"]["state"] == "inconclusive")
    exp = AX.jev_refresh_cost(p["keys"], J.read_scores(p["scores"])[2], [("luna-xhigh", "fastapi"), ("luna-xhigh", "axum")], C.read_jsonl(p["replay"]))
    check("S4 per-refresh Jev cost is computed (> 0) on the planted run, from the measured transitions; the caps are exercised by the next check", exp[("luna-xhigh", "fastapi")]["cost_usd"] > 0 and exp[("luna-xhigh", "fastapi")]["transitions"] == 6)
    old = (x2c.JEV_COST_PER_REFRESH_MAX_USD, x2c.JEV_SECONDS_PER_REFRESH_MAX)
    x2c.JEV_COST_PER_REFRESH_MAX_USD = 1e-9
    r_cost = AX.analyze(p["data"], p["replay"], p["pairs"], p["scores"])
    x2c.JEV_COST_PER_REFRESH_MAX_USD, x2c.JEV_SECONDS_PER_REFRESH_MAX = old[0], 1e-9
    r_time = AX.analyze(p["data"], p["replay"], p["pairs"], p["scores"])
    x2c.JEV_COST_PER_REFRESH_MAX_USD, x2c.JEV_SECONDS_PER_REFRESH_MAX = old
    check("S4 with the cost cap lowered the Jev cell is `ineligible` (reason names the cost) and with the time cap lowered it is `ineligible` (reason names the time); the outcome is no-adoption",
          r_cost["x2c"]["cells"]["luna-xhigh+jev"]["state"] == "ineligible" and any("cost" in r for r in r_cost["x2c"]["cells"]["luna-xhigh+jev"]["reasons"]) and r_time["x2c"]["cells"]["luna-xhigh+jev"]["state"] == "ineligible"
          and any("time" in r for r in r_time["x2c"]["cells"]["luna-xhigh+jev"]["reasons"]) and r_cost["x2c"]["selection"]["outcome"] == "no-adoption")

    # served check negatives on the S1 data (clone + re-plant)
    def served_variant(name, spec=None, **kw):
        dd = os.path.join(tmp, name)
        shutil.copytree(p["data"], dd)
        shutil.rmtree(os.path.join(dd, "maps"))
        plant_data(dd, {"luna-xhigh": spec or PLANTED}, **kw)
        return AX.analyze(dd, p["replay"], p["pairs"], p["scores"])
    v1 = served_variant("S4-eff", probe_effort={"luna-xhigh": "low"})
    v2 = served_variant("S4-mod", probe_model={"luna-xhigh": "gpt-6.1-sol"})
    v3 = served_variant("S4-flag", flags_override={"luna-xhigh": ["--model", "gpt-6.1-terra", "--reasoning-effort", "low"]})
    v4 = served_variant("S4-opt", option_override={"luna-xhigh": "default"})
    for nm, v, frag in (("a probe serving effort low", v1, "reasoning effort 'low'"), ("a probe serving model gpt-6.1-sol", v2, "served model 'gpt-6.1-sol'"), ("maps recorded with other flags", v3, "meta.json records flags"),
                        ("Clawpatch agent-start reporting model=default", v4, "agent-start reported")):
        sc = v["x2c"]["served"]["luna-xhigh"]
        check("S4 served check: %s -> FAILED naming it, both cells of the arm inconclusive (not citable), outcome inconclusive" % nm,
              not sc["ok"] and any(frag in r for r in sc["reasons"]) and v["x2c"]["cells"]["luna-xhigh"]["state"] == "inconclusive" and v["x2c"]["cells"]["luna-xhigh+jev"]["state"] == "inconclusive"
              and v["x2c"]["selection"]["outcome"] == "inconclusive", sc["reasons"][:2])
    dn = os.path.join(tmp, "S4-nov")
    shutil.copytree(p["data"], dn)
    os.remove(os.path.join(dn, "tool-versions-luna-xhigh-axum.json"))
    vn = AX.analyze(dn, p["replay"], p["pairs"], p["scores"])
    check("S4 served check: a missing versions file fails it", any("axum: no tool-versions file" in r for r in vn["x2c"]["served"]["luna-xhigh"]["reasons"]) and vn["x2c"]["selection"]["outcome"] == "inconclusive")
    v5 = served_variant("S4-noseconds", spec=dict(PLANTED, seconds=None))
    c5 = v5["x2c"]["cells"]["luna-xhigh+jev"]
    check("S5 a missing map wall time makes an adopt cell ineligible (`no map wall time recorded`) through analyze() and the outcome is no-adoption",
          c5["state"] == "ineligible" and any("no map wall time recorded" in x for x in c5["reasons"]) and v5["x2c"]["selection"]["outcome"] == "no-adoption", c5)
    dd = os.path.join(tmp, "S5-halfseconds")
    shutil.copytree(p["data"], dd)
    shutil.rmtree(os.path.join(dd, "maps"))
    plant_data(dd, {"luna-xhigh": PLANTED}, repos=("fastapi",))
    plant_data(dd, {"luna-xhigh": dict(PLANTED, seconds=None)}, repos=("axum",))
    v6 = AX.analyze(dd, p["replay"], p["pairs"], p["scores"])
    c6 = v6["x2c"]["cells"]["luna-xhigh+jev"]
    check("S5 a map wall time recorded in one repository only is no refresh time: the adopt cell is ineligible (`no map wall time recorded`) and the outcome is no-adoption",
          c6["refresh_seconds"] is None and c6["state"] == "ineligible" and any("no map wall time recorded" in x for x in c6["reasons"]) and v6["x2c"]["selection"]["outcome"] == "no-adoption", c6)

    # ---- four arms: selection by refresh time, coverage guard, method preference and the tie band
    arms4 = {"luna-xhigh": PLANTED, "terra-none": {"kind": "clean", "n_reg": 60, "seconds": 100.0}, "terra-low": {"kind": "clean", "n_reg": 60, "seconds": 300.0}, "haiku": {"kind": "clean", "n_reg": 20, "seconds": 50.0}}
    p4 = pipeline(tmp, "S5-four", arms4)
    r4 = p4["res"]["x2c"]
    st = {c: v["state"] for c, v in r4["cells"].items()}
    check("S5 four arms: luna Jaccard not-adopted / Jev eligible; terra-none and terra-low eligible with both methods (clean maps adopt at 0.90); haiku adopts but is ineligible by the coverage guard (its maps own about a third of the reference's files)",
          st == {"luna-xhigh": "not-adopted", "luna-xhigh+jev": "eligible", "terra-none": "eligible", "terra-none+jev": "eligible", "terra-low": "eligible", "terra-low+jev": "eligible", "haiku": "ineligible", "haiku+jev": "ineligible"}
          and not r4["coverage"]["haiku"]["ok"] and r4["coverage"]["terra-none"]["ok"] and any("coverage guard" in x for x in r4["cells"]["haiku"]["reasons"]), (st, r4["coverage"]["haiku"]["reasons"]))
    s4 = r4["selection"]
    check("S5 selection: the fastest eligible refresh time is terra-none (2 x 100 s); its Jaccard and Jev cells are tied within 10 %, the Jaccard method is preferred (no egress, no cost): adopt terra-none with the jaccard method at 0.90",
          s4["outcome"] == "adopt" and s4["cell"] == "terra-none" and sorted(s4["tied"]) == ["terra-none", "terra-none+jev"] and "match_jaccard = 0.90" in s4["proposal"], s4)
    check("S5 the area ratio is reported, not gated: haiku owns about a third of the reference's areas per map (20 against 58 per map: ratio between 0.30 and 0.40 in both repositories) and the adopt proposal for terra-none carries "
          "`areas per map` with its ratios",
          all(0.30 < r4["coverage"]["haiku"]["area_ratio_to_reference"][repo] < 0.40 for repo in ("fastapi", "axum")) and r4["coverage"]["luna-xhigh"]["area_ratio_to_reference"] == {"axum": 1.0, "fastapi": 1.0}
          and "areas per map: axum x" in s4["proposal"] and "fastapi x" in s4["proposal"], (r4["coverage"]["haiku"], s4["proposal"]))
    check("S5 the coverage floor is the preregistered 0.5 on owned files (option B: haiku's file ratio is below it, terra's above)", x2c.COVERAGE_FLOOR == 0.5 and all(r4["coverage"]["haiku"]["ratio_to_reference"][repo] < 0.5 for repo in ("fastapi", "axum"))
          and all(r4["coverage"]["terra-none"]["ratio_to_reference"][repo] >= 0.5 for repo in ("fastapi", "axum")), (r4["coverage"]["haiku"]["ratio_to_reference"], r4["coverage"]["terra-none"]["ratio_to_reference"]))
    rows4 = C.read_jsonl(p4["scores"])
    fa4 = next(i for i, r_ in enumerate(rows4) if r_["type"] == "answer")
    OTHER_MODEL, OTHER_VERSION = "typesafe/jev-1.13-20260901", "x2c.concern@0"
    for gi_, (nm_, patched_, frag_) in enumerate((
            ("the header names another model", [dict(r_, model=OTHER_MODEL) if r_["type"] == "header" else r_ for r_ in rows4], "not the pin"),
            ("the header names another question version", [dict(r_, question_version=OTHER_VERSION) if r_["type"] == "header" else r_ for r_ in rows4], "not the preregistered"),
            ("an answer row names another model", [dict(r_, model=OTHER_MODEL) if i == fa4 else r_ for i, r_ in enumerate(rows4)], "replies from other models"),
            ("an answer row carries another question version", [dict(r_, question_version=OTHER_VERSION) if i == fa4 else r_ for i, r_ in enumerate(rows4)], "another question version"))):
        sp_ = os.path.join(tmp, "S5-gate-%d.jsonl" % gi_)
        J.write_jsonl(sp_, patched_)
        rg_ = AX.analyze(p4["data"], p4["replay"], p4["pairs"], sp_)["x2c"]
        check("S5 global Jev gate: %s -> a global Jev-check reason, every +jev cell inconclusive, Jaccard cells unaffected (adopt terra-none, jaccard)" % nm_,
              any(frag_ in r for r in rg_["jev_check"]["reasons"]) and all(rg_["cells"][a + x2c.JEV_SUFFIX]["state"] == "inconclusive" for a in x2c.ARMS)
              and rg_["cells"]["terra-none"]["state"] == "eligible" and rg_["selection"]["cell"] == "terra-none", rg_["jev_check"]["reasons"])
    sn = AX.select({"a": {"arm": "terra-none", "method": "jev", "state": "eligible", "refresh_seconds": 200.0, "margin": 0.1, "t_star_pct": 70},
                    "b": {"arm": "terra-low", "method": "jaccard", "state": "eligible", "refresh_seconds": 205.0, "margin": 0.01, "t_star_pct": 60},
                    "c": {"arm": "luna-xhigh", "method": "jaccard", "state": "eligible", "refresh_seconds": 900.0, "margin": 0.5, "t_star_pct": 50}})
    s_m = AX.select({"a": {"arm": "terra-none", "method": "jaccard", "state": "eligible", "refresh_seconds": 200.0, "margin": 0.01, "t_star_pct": 70}, "b": {"arm": "terra-low", "method": "jaccard", "state": "eligible", "refresh_seconds": 210.0, "margin": 0.2, "t_star_pct": 60}})
    s_o = AX.select({"a": {"arm": "terra-none", "method": "jaccard", "state": "eligible", "refresh_seconds": 200.0, "margin": 0.2, "t_star_pct": 70}, "b": {"arm": "terra-low", "method": "jaccard", "state": "eligible", "refresh_seconds": 400.0, "margin": 0.9, "t_star_pct": 60}})
    s_e = AX.select({"a": {"arm": "haiku", "method": "jaccard", "state": "not-adopted", "refresh_seconds": None, "margin": None, "t_star_pct": None}, "b": {"arm": "haiku", "method": "jev", "state": "ineligible", "refresh_seconds": None, "margin": None, "t_star_pct": None}})
    s_i = AX.select({"a": {"arm": "haiku", "method": "jaccard", "state": "inconclusive", "refresh_seconds": None, "margin": None, "t_star_pct": None}})
    check("S5 select() rules: the Jaccard cell wins a tie against Jev even with a smaller margin (200 vs 205 s); among tied cells of one method the larger survival margin wins; outside the band the faster cell wins; no eligible cell -> "
          "no-adoption, or inconclusive when any cell is inconclusive",
          sn["cell"] == "b" and s_m["cell"] == "b" and s_o["cell"] == "a" and s_e["outcome"] == "no-adoption" and s_i["outcome"] == "inconclusive", (sn, s_m, s_o))
    cf = AX.confirm({"fastapi": {"decisive": {str(q): {"survival_exact": "9/10", "labelled": 100, "different": 0, "spurious_wilson": [0, 0.03], "matched": 90} for q in C.GRID_PCT}},
                     "axum": {"decisive": {str(q): {"survival_exact": "7/10" if q >= 50 else "9/10", "labelled": 100, "different": 0 if q < 50 else 0, "spurious_wilson": [0, 0.03], "matched": 90} for q in C.GRID_PCT}}}, ["fastapi", "axum"])
    check("S5 cross-repository confirmation is directional: tuned on fastapi (passes everywhere, so t = 0.90) axum's survival 7/10 fails C1; tuned on axum (passes only at <= 0.45) fastapi passes at 0.45; so the threshold is not confirmed",
          cf["per_repo"]["axum"]["t_pct"] == 90 and not cf["per_repo"]["axum"]["ok"] and cf["per_repo"]["fastapi"]["t_pct"] == 45 and cf["per_repo"]["fastapi"]["ok"] and not cf["confirmed"], cf)
    real_confirm = AX.confirm
    AX.confirm = lambda t, r: {"confirmed": False, "per_repo": {}}
    try:
        r_nc = AX.analyze(p["data"], p["replay"], p["pairs"], p["scores"])
    finally:
        AX.confirm = real_confirm
    cn = r_nc["x2c"]["cells"]["luna-xhigh+jev"]
    check("S5 an adopt cell whose cross-repository confirmation fails is ineligible through analyze() (reason names the confirmation) and the outcome is no-adoption",
          cn["state"] == "ineligible" and any("confirmation" in x for x in cn["reasons"]) and r_nc["x2c"]["selection"]["outcome"] == "no-adoption", cn)
    pe = pipeline(tmp, "S5-unexp", {"luna-xhigh": PLANTED, "terra-none": {"kind": "clean", "n_reg": 60, "seconds": 100.0}}, jev=False)
    rows_u, pairs_u = C.read_jsonl(pe["replay"]), os.path.join(pe["d"], "pairs-renamed")
    shutil.copytree(pe["pairs"], pairs_u)
    for r in rows_u:
        if r["arm"] == "terra-none":
            r["arm"] = "mystery"
    J.write_jsonl(pe["replay"], rows_u)
    J.write_jsonl(os.path.join(pairs_u, "key.jsonl"), [dict(k, arm="mystery") if k["arm"] == "terra-none" else k for k in C.read_jsonl(os.path.join(pe["pairs"], "key.jsonl"))])
    ru = AX.analyze(pe["data"], pe["replay"], pairs_u, os.path.join(tmp, "no-scores.jsonl"))
    check("S5 an arm outside the preregistered four is reported and makes every decision not citable; a missing Jev scores file makes the Jev cells inconclusive (no header)",
          ru["x2c"]["unexpected_arms"] == ["mystery"] and all(not c["citable"] for c in ru["x2c"]["cells"].values() if "citable" in c) and not ru["x2c"]["jev_check"]["ok"]
          and ru["x2c"]["cells"]["luna-xhigh+jev"]["state"] == "inconclusive")

    # ---- candidate construction
    keys_by = {}
    for k in p["keys"]:
        keys_by.setdefault((k["arm"], k["repo"], k["transition"]), []).append(k)
    t1 = keys_by[("luna-xhigh", "fastapi", "T1")]
    red = [k for k in t1 if k["prev_feature_id"].startswith("red_") and k["next_feature_id"].startswith("red_")]
    check("S6 candidates: a transition's candidates are exactly the pairs sharing a file: 40 regular + 8 redrawn (Jaccard 2/18, below T_MIN: the Jaccard rule never matches them) + 6 volatile + 16 trap = 70; every candidate shares >= 1 file; "
          "no candidate has Jaccard 0; controls are generated from the unique candidates (ceil 8 % / 4 %)",
          len(t1) == 70 and len(red) == 8 and all(abs(k["jaccard"] - 2 / 18) < 1e-9 for k in red) and all(k["shared"] >= 1 and k["jaccard"] > 0 for k in t1)
          and len([c for c in p["ctl"] if c["kind"] == "swapped"]) == math.ceil(0.08 * len({k["pair_id"] for k in p["keys"]})) and len([c for c in p["ctl"] if c["kind"] == "identical"]) == math.ceil(0.04 * len({k["pair_id"] for k in p["keys"]})))
    rows_m = C.read_jsonl(p["replay"])
    for r in rows_m:
        if r["thresholdPct"] == 30 and r["status"] == "ok" and r["matches"]:
            r["matches"][0]["jaccard"] = 0.9999
            break
    bad_replay = os.path.join(tmp, "S6-bad-replay.jsonl")
    J.write_jsonl(bad_replay, rows_m)
    rb = subprocess.run([PY, os.path.join(X2C, "build_candidates.py"), "--data", p["data"], "--replay", bad_replay, "--out", os.path.join(tmp, "S6-bad-out")], capture_output=True, text=True)
    check("S6 build_candidates aborts when a replay match is not a candidate or its Jaccard differs (the candidate definition and the replay must agree)", rb.returncode != 0 and "disagree" in rb.stderr + rb.stdout, rb.stderr[-300:])
    rb2 = subprocess.run([PY, os.path.join(X2C, "build_candidates.py"), "--data", p["data"], "--replay", p["replay"], "--out", p["pairs"]], capture_output=True, text=True)
    check("S6 build_candidates refuses to rebuild once labels exist", rb2.returncode != 0 and "refusing" in rb2.stderr + rb2.stdout)
    check("S6 the Jev matcher is greedy best-first one-to-one with ties broken by (-s, -P(same), prior id, next id) (unit): the best pair is taken first, a taken area or feature is skipped, scores below tau are not matched; "
          "among equal s the larger P(same_concern) wins, then the ids",
          [(m["prevFeatureId"], m["nextFeatureId"]) for m in J.greedy([(0.9, 0.9, "a", "x"), (0.95, 0.95, "b", "x"), (0.9, 0.9, "a", "y"), (0.4, 0.4, "c", "z")], 50)] == [("b", "x"), ("a", "y")]
          and J.greedy([(0.5, 0.5, "b", "y"), (0.5, 0.5, "a", "y")], 50)[0]["prevFeatureId"] == "a" and J.greedy([(1.0, 0.3, "a", "y"), (1.0, 0.9, "b", "y")], 50)[0]["prevFeatureId"] == "b"
          and J.pair_score({"ab": (0.2, 0.1)}) is None and all(abs(u - v) < 1e-12 for u, v in zip(J.pair_score({"ab": (0.2, 0.1), "ba": (0.4, 0.3)}), (0.3, 0.2))))
    ov = os.path.join(tmp, "S6-overlap-scores.jsonl")
    J.write_jsonl(ov, [{"type": "header"}] + [{"type": "answer", "pair_id": "q", "order": o, "probabilities": {"same_concern": 0.3, "overlapping_concern": 0.6, "different_concern": 0.1, "cannot_tell": 0.0}}
                                               for o in ("ab", "ba")])
    s_ov, ps_ov = J.pair_score(J.read_scores(ov)[1]["q"])
    check("S6 the Jev score counts overlapping_concern (the answer C2 does not count as spurious): P(same) 0.3 + P(overlapping) 0.6 gives s = 0.9 with P(same) 0.3 kept for the tie-break",
          abs(s_ov - 0.9) < 1e-12 and abs(ps_ov - 0.3) < 1e-12, (s_ov, ps_ov))


# ---------------------------------------------------------------- J the Jev scorer
def scorer_checks(tmp):
    p = pipeline(tmp, "J0", {"luna-xhigh": {"kind": "clean", "n_reg": 6, "seconds": 5.0}}, jev=False)
    inp = os.path.join(p["pairs"], "jev-input.jsonl")
    rows = C.read_jsonl(inp)
    log = os.path.join(tmp, "J-stub.log")
    stub = stub_jev.Stub(x2c.JEV_MODEL, "planted", log_path=log).start()
    out = os.path.join(tmp, "J-scores.jsonl")
    r = run_scorer(tmp, p["pairs"], out, stub, ["-questions-dir", os.path.join(tmp, "J-questions")])
    stub.stop()
    reqs = [json.loads(line) for line in open(log)]
    hdr, sc, ans, fl = J.read_scores(out)
    first = reqs[0]
    state_keys = {"feature_a", "feature_b", "shared_files_total", "shared_files"}
    check("J scorer: exit 0, a header and two answer rows per pair (orders ab and ba), one request per unit with the single question key `q1` of type choice, four criteria keys, the pin as model, and the key sent as a Bearer header only",
          r.returncode == 0 and hdr and len(ans) == 2 * len(rows) and len(reqs) == 2 * len(rows) and all(q["question_keys"] == ["q1"] and q["type"] == "choice" and sorted(q["criteria_keys"]) == sorted(
              ["same_concern", "overlapping_concern", "different_concern", "cannot_tell"]) and q["model"] == x2c.JEV_MODEL and q["auth_present"] and not q["body_has_key"] and set(q["state"]) == state_keys for q in reqs), (r.stderr[-300:], len(reqs)))
    ab = [q for q in reqs if q["state"]["feature_a"]["title"] == rows[0]["state"]["left"]["title"] and q["state"]["feature_b"]["title"] == rows[0]["state"]["right"]["title"]]
    ba = [q for q in reqs if q["state"]["feature_a"]["title"] == rows[0]["state"]["right"]["title"] and q["state"]["feature_b"]["title"] == rows[0]["state"]["left"]["title"]]
    check("J scorer: the two orders of one pair swap feature_a and feature_b (each exactly once); the state carries titles, summaries, kinds, entrypoints, file lists and the shared files as the labellers see them",
          len(ab) == 1 and len(ba) == 1 and ab[0]["state"]["feature_a"]["owned_files"] == rows[0]["state"]["left"]["owned_files"] and ab[0]["state"]["shared_files"] == rows[0]["state"]["shared_files"]
          and ab[0]["state"]["feature_a"]["summary"] == rows[0]["state"]["left"]["summary"])
    blob = open(out).read() + r.stdout + r.stderr + open(log).read()
    check("J key hygiene: the Jev key appears in no output row, log line, stdout or stderr; the version key and model are recorded per answer; the question file was written under its version key",
          KEY not in blob and all(a["question_version"] == x2c.JEV_QUESTION_VERSION and a["model"] == x2c.JEV_MODEL for a in ans) and os.path.exists(os.path.join(tmp, "J-questions", x2c.JEV_QUESTION_VERSION + ".json")))
    check("J scorer: criteria sent are the guide's four label definitions word for word (the same strings the labellers were given)",
          all(squash(reqs[0]["criteria"][lab]) in squash(open(os.path.join(X2, "GUIDE.md")).read()) for lab in ("same_concern", "overlapping_concern", "different_concern", "cannot_tell")))
    # resume: nothing left to do -> no new request; failed units are retried
    stub2 = stub_jev.Stub(x2c.JEV_MODEL, "planted", log_path=log + "2", fail_pairs=[rows[0]["state"]["left"]["title"]]).start()
    out2 = os.path.join(tmp, "J-scores2.jsonl")
    r2 = run_scorer(tmp, p["pairs"], out2, stub2)
    stub2.stop()
    _h2, _s2, a2, f2 = J.read_scores(out2)
    stub3 = stub_jev.Stub(x2c.JEV_MODEL, "planted", log_path=log + "3").start()
    r3 = run_scorer(tmp, p["pairs"], out2, stub3)
    stub3.stop()
    n3 = len(open(log + "3").read().splitlines()) if os.path.exists(log + "3") else 0
    _h3, _s3, a3, f3 = J.read_scores(out2)
    check("J resume: a run with failures exits 3 and records them; the re-run retries exactly the failed units (%d requests), appends answers and exits 0; the failure rows stay (append-only)" % n3,
          r2.returncode == 3 and f2 and r3.returncode == 0 and n3 == len(f2) and len(a3) == 2 * len(rows) and len(f3) == len(f2), (r2.returncode, r3.returncode, n3, len(f2)))
    other = os.path.join(tmp, "J-other.jsonl")
    open(other, "w").write(open(inp).read().replace(rows[0]["pair_id"], "0" * 20, 1))
    stub4 = stub_jev.Stub(x2c.JEV_MODEL, "planted").start()
    exe = V.go_bin(tmp, "jev-bin", x2c.JEV_SCORER_PKG)
    r4 = subprocess.run([exe, "-in", other, "-out", out2, "-endpoint", stub4.url], env=jev_env(), capture_output=True, text=True)
    r5 = subprocess.run([exe, "-in", inp, "-out", os.path.join(tmp, "J-alias.jsonl"), "-endpoint", stub4.url, "-model", "jev-latest"], env=jev_env(), capture_output=True, text=True)
    r6 = subprocess.run([exe, "-in", inp, "-out", os.path.join(tmp, "J-nokey.jsonl"), "-endpoint", stub4.url], env={k: v for k, v in os.environ.items() if k != "OPENROUTER_API_KEY"}, capture_output=True, text=True)
    stub4.stop()
    check("J scorer refusals: resuming into an output whose header names another input is refused; an alias (`jev-latest`) is refused; a missing key is refused (naming only that it is unset); none writes an answer",
          r4.returncode == 2 and "refusing to resume" in r4.stderr and r5.returncode == 2 and "alias" in r5.stderr and r6.returncode == 2 and "OPENROUTER_API_KEY is not set" in r6.stderr
          and not os.path.exists(os.path.join(tmp, "J-alias.jsonl")) and not os.path.exists(os.path.join(tmp, "J-nokey.jsonl")))
    stub5 = stub_jev.Stub(x2c.JEV_MODEL, "planted", rate_limit_first=3).start()
    r7 = run_scorer(tmp, p["pairs"], os.path.join(tmp, "J-429.jsonl"), stub5)
    stub5.stop()
    check("J rate limits: HTTP 429 with Retry-After is retried by the product client and every unit still answers", r7.returncode == 0 and len(J.read_scores(os.path.join(tmp, "J-429.jsonl"))[2]) == 2 * len(rows), r7.stderr[-200:])
    stub6 = stub_jev.Stub(x2c.JEV_MODEL, "planted").start()
    r8 = run_scorer(tmp, p["pairs"], os.path.join(tmp, "J-cap.jsonl"), stub6, ["-max-cost-usd", "1e-12", "-workers", "1"])
    stub6.stop()
    check("J cost stop: the run stops once the billed cost exceeds -max-cost-usd (exit 4, fewer answers than units)", r8.returncode == 4 and len(J.read_scores(os.path.join(tmp, "J-cap.jsonl"))[2]) < 2 * len(rows), r8.stderr[-200:])
    stub7 = stub_jev.Stub(x2c.JEV_MODEL, "planted", drop_option="cannot_tell").start()
    out7 = os.path.join(tmp, "J-schema.jsonl")
    r9 = run_scorer(tmp, p["pairs"], out7, stub7)
    stub7.stop()
    _h7, _s7, a7, f7 = J.read_scores(out7)
    check("J schema: a reply that gives three of the four options is a failure row of kind `schema` (never a scored answer): exit 3, every unit unscored, the row keeps the billed cost and model",
          r9.returncode == 3 and not a7 and len(f7) == 2 * len(rows) and {f["kind"] for f in f7} == {"schema"} and all(f["cost_known"] and f["model"] == x2c.JEV_MODEL for f in f7), (r9.returncode, len(a7), len(f7)))
    stub9 = stub_jev.Stub(x2c.JEV_MODEL, "planted", scale=0.9).start()
    out9 = os.path.join(tmp, "J-unnorm.jsonl")
    r11 = run_scorer(tmp, p["pairs"], out9, stub9)
    stub9.stop()
    _h9, _s9, a9, f9 = J.read_scores(out9)
    check("J schema: four options summing to 0.9 -> failure rows of kind `schema`, exit 3, every unit unscored", r11.returncode == 3 and not a9 and len(f9) == 2 * len(rows) and {f["kind"] for f in f9} == {"schema"}, (r11.returncode, len(a9), len(f9)))
    stub8 = stub_jev.Stub(x2c.JEV_MODEL, "planted").start()
    r10 = run_scorer(tmp, p["pairs"], os.path.join(tmp, "J-cap.jsonl"), stub8, ["-max-cost-usd", "1e-12"])
    n_req8 = stub8.requests
    stub8.stop()
    check("J cost stop is cumulative over the output: re-running after the cost stop refuses to start (exit 2, `budget already exhausted`) and sends no request",
          r10.returncode == 2 and "budget already exhausted" in r10.stderr and n_req8 == 0, (r10.returncode, r10.stderr[-200:], n_req8))


# ---------------------------------------------------------------- C collector and E environment
def collector_checks(tmp):
    node = shutil.which("node")
    check("C node available (source .tools/activate)", bool(node))
    if not node:
        return
    src = os.path.join(tmp, "tiny-src")
    os.makedirs(src)
    git = lambda *a, cwd=src: subprocess.run(["git", "-c", "user.name=t", "-c", "user.email=t@e.invalid", *a], cwd=cwd, check=True, capture_output=True, text=True)
    git("init", "-q")
    for dname in ("alpha", "beta", "gamma"):
        os.makedirs(os.path.join(src, dname))
        for k in range(4):
            open(os.path.join(src, dname, "m%d.py" % k), "w").write("x = %d\n" % k)
    git("add", "-A")
    git("commit", "-q", "-m", "init")
    rev = git("rev-parse", "HEAD").stdout.strip()
    spec = json.dumps({"tiny": {"rev": rev, "url": src, "ext": ".py"}})
    pv_, ps_ = C.review_safety_pins()
    sb = os.path.join(tmp, "stubbuild")
    os.makedirs(os.path.join(sb, "dist"))
    stub = os.path.join(sb, "dist", "cli.js")
    shutil.copy(os.path.join(HERE, "stub_clawpatch_x2c.js"), stub)
    open(os.path.join(sb, "dist", "prompt.js"), "w").write("// fixture: %s\n" % C.SANITIZER_TOKEN)
    open(os.path.join(sb, C.BUILD_MARKER_FILE), "w").write("%s %s\n" % (pv_, ps_))
    bins = os.path.join(tmp, "fakebin")
    os.makedirs(bins)
    # fake codex: serves what its flags say (like the real CLI); without them it serves the "user config" (gpt-6.1-sol, low). Prints the banner lines the probe parses.
    cbody = ('import json, os, sys\na = sys.argv[1:]\nif a[:1] == ["--version"]:\n    print("codex-cli 0.0.0-fake")\n    sys.exit(0)\nsys.stdin.read()\n'
             'open(os.path.abspath(__file__) + ".calls", "a").write(json.dumps({"argv": a, "env": sorted(os.environ)}) + "\\n")\n'
             'model = a[a.index("--model") + 1] if "--model" in a else "gpt-6.1-sol"\neff = "low"\n'
             'for i, x in enumerate(a):\n    if x == "-c" and a[i + 1].startswith("model_reasoning_effort="):\n        eff = a[i + 1].split("=", 1)[1].strip(chr(34))\n'
             'print("model: " + model)\nif not os.environ.get("X2C_FAKE_NO_EFFORT"):\n    print("reasoning effort: " + (os.environ.get("X2C_FAKE_EFFORT") or eff))\n')
    # fake claude: --version for the host-auth check; -p prints the JSON the probe parses (modelUsage keys = the requested model) and logs argv
    kbody = ('import json, os, sys\na = sys.argv[1:]\nif a[:1] == ["--version"]:\n    print("2.1.312 (Claude Code)")\n    sys.exit(0)\nsys.stdin.read()\n'
             'open(os.path.abspath(__file__) + ".calls", "a").write(json.dumps({"argv": a, "env": sorted(os.environ)}) + "\\n")\n'
             'model = a[a.index("--model") + 1] if "--model" in a else "claude-default"\nprint(json.dumps({"result": "ok", "modelUsage": {model: {}}}))\n')
    for nm, body in (("codex", cbody), ("claude", kbody)):
        open(os.path.join(bins, nm), "w").write("#!%s\n%s" % (PY, body))
        os.chmod(os.path.join(bins, nm), 0o755)
    poison = {"ANTHROPIC_API_KEY": "k", "OPENAI_API_KEY": "k", "OPENROUTER_API_KEY": "k", "MY_SECRET_THING": "k", "GITHUB_TOKEN": "k", "X2_PLAIN": "keep", "CLAWPATCH_CLAUDE_AUTH_CONTEXT": "isolated"}

    def ncalls(tool):
        p_ = os.path.join(bins, tool + ".calls")
        return len(open(p_).read().splitlines()) if os.path.exists(p_) else 0

    def collect(name, arm, fail_calls="", env_extra=None, args=(), wrapper=None, repo="tiny", data=None, work=None, spec_=None):
        d = data or os.path.join(tmp, "col-" + name)
        env = dict(os.environ, X2_STUB_COUNTER=d + "-counter", X2_STUB_FAIL_CALLS=fail_calls, X2_STUB_ENVLOG=d + "-envlog.jsonl", X2C_STUB_ARGVLOG=d + "-argv.jsonl",
                   X2C_BASE_STUB=os.path.join(X2, "validation", "stub_clawpatch.js"), **poison)
        env["PATH"] = bins + os.pathsep + os.environ["PATH"]
        for n_ in (*C.TIMEOUT_ENV_NAMES, *C.REFUSED_MODEL_ENV, *C.REFUSED_SANDBOX_ENV, "CLAWPATCH_CLAUDE_BIN", "CODEX_HOME", "X2C_FAKE_NO_EFFORT", "X2C_FAKE_EFFORT"):
            env.pop(n_, None)
        env.update(env_extra or {})
        cmd = [PY, os.path.join(X2C, "collect_x2c.py"), "--arm", arm, "--repo", repo, "--source", src, "--work", work or d + "-work", "--data", d, "--node", node, "--cli", stub,
               "--spec", spec_ or spec, "--expect-clawpatch-version", "stub-0", "--term-grace", "2", *args]
        if wrapper:
            cmd = [PY, "-c", "import sys; sys.path.insert(0, %r)\n" % X2C + wrapper, *cmd[1:]]
        return d, subprocess.run(cmd, env=env, capture_output=True, text=True)

    runs = {}
    for arm in x2c.ARMS:
        runs[arm] = collect("ok-" + arm, arm, "2" if arm == "terra-none" else "")
    ok_all = True
    for arm, (d, r) in runs.items():
        a = x2c.ARMS[arm]
        metas = {m: C.read_meta(d, arm, "tiny", m) for m in C.MAP_IDS}
        argv_log = [json.loads(line)["args"] for line in open(d + "-argv.jsonl")]
        maps_, inits = [x for x in argv_log if "map" in x], [x for x in argv_log if "init" in x]
        want_tail = ["--json", "map", "--source", "agent", "--provider", a["provider"], "--model", a["model"], "--reasoning-effort", a["effort"]]
        tv = json.load(open(os.path.join(d, "tool-versions-%s-tiny.json" % arm)))
        ok_all &= (r.returncode == 0 and all(v is not None and v["rc"] == 0 for v in metas.values()) and len(maps_) == (8 if arm == "terra-none" else 7) and all(x[0] == "--root" and x[2] == "--state-dir" and x[4:] == want_tail for x in maps_)
                   and all("--model" not in x and "--reasoning-effort" not in x for x in inits) and all(v["schema"] == "x2c-map-v1" and v["map_flags"] == x2c.arm_flags(arm) and v["provider_reported"] == a["provider"]
                                                                                                         and v["model_option_reported"] == a["model"] and v["workspace_dirty"] is False for v in metas.values())
                   and tv["x2c"]["map_flags"] == x2c.arm_flags(arm) and tv["x2c"]["effort"] == a["effort"] and tv["x2c"]["x2_pins"] == x2c.X2_PINS and tv["x2c"]["collect_x2c_sha256"] == C.sha256_file(os.path.join(X2C, "collect_x2c.py"))
                   and tv["model_probe_before"]["model"] == tv["model_probe_after"]["model"] == a["model"] and tv["model_probe_before"]["reasoning_effort"] == tv["model_probe_after"]["reasoning_effort"] == a["effort"]
                   and tv["clawpatch_build_safety"]["ok"] is True)
        if not ok_all:
            print("collector detail", arm, r.returncode, r.stderr[-500:], [x for x in maps_[:1]], tv.get("model_probe_before"))
            break
    check("C every arm's collector run (stub Clawpatch): exit 0, 7 maps rc 0 (terra-none: one injected failure retried, 8 map invocations), every map argv is exactly `--root R --state-dir S --json map --source agent --provider P --model M "
          "--reasoning-effort E` for that arm (init carries no model flags), meta.json records schema x2c-map-v1, flags, provider and Clawpatch's agent-start option, the versions file records flags, effort, the X2 pins, the collector digest and "
          "the build-safety check, and both probes served the arm's model and effort", ok_all)
    codex_calls = [json.loads(line) for line in open(os.path.join(bins, "codex.calls"))]
    claude_calls = [json.loads(line) for line in open(os.path.join(bins, "claude.calls"))]
    want_probe = lambda arm, dd: ["exec", "--cd", dd, "--sandbox", "read-only", "--skip-git-repo-check", "--model", x2c.ARMS[arm]["model"], "-c", 'model_reasoning_effort="%s"' % x2c.ARMS[arm]["effort"], "-"]
    efforts = [c["argv"][c["argv"].index("-c") + 1] for c in codex_calls]
    check("C probes: 6 codex probes (3 codex arms x before/after) mirror the provider invocation with the arm's model and effort (no --ignore-user-config), 2 claude probes carry `--model claude-haiku-5-5` and NO --effort (effort none) "
          "with `--safe-mode`, all under the scrubbed environment (no credential, X2_PLAIN kept)",
          len(codex_calls) == 6 and sorted(efforts) == sorted(['model_reasoning_effort="xhigh"'] * 2 + ['model_reasoning_effort="none"'] * 2 + ['model_reasoning_effort="low"'] * 2) and all(c["argv"][:3] == ["exec", "--cd", c["argv"][2]] for c in codex_calls)
          and not any("--ignore-user-config" in c["argv"] for c in codex_calls) and len(claude_calls) == 2 and all(c["argv"][-2:] == ["--model", "claude-haiku-5-5"] and "--effort" not in c["argv"] and "--safe-mode" in c["argv"] for c in claude_calls)
          and not any(n in c["env"] for c in codex_calls + claude_calls for n in ("ANTHROPIC_API_KEY", "OPENAI_API_KEY", "OPENROUTER_API_KEY", "GITHUB_TOKEN", "MY_SECRET_THING")) and all("X2_PLAIN" in c["env"] for c in codex_calls), (efforts, claude_calls[:1]))
    log = [json.loads(line) for line in open(runs["haiku"][0] + "-envlog.jsonl")]
    check("C every Clawpatch call of the claude arm runs with the host auth context and both provider timeouts at 900000 ms", len(log) >= 8 and all(e["claudeAuth"] == "host" and e["codexTimeout"] == "900000" and e["claudeTimeout"] == "900000" for e in log))
    r_again = collect("ok-luna-xhigh", "luna-xhigh", data=runs["luna-xhigh"][0])[1]
    check("C append-only: a second collection into the same data directory is refused before any model call and before cloning", r_again.returncode != 0 and "already recorded" in r_again.stderr + r_again.stdout)
    refused_ok, detail = True, []
    for env_name in (*C.REFUSED_MODEL_ENV, *C.REFUSED_SANDBOX_ENV):
        for value in ("x", ""):
            before = (ncalls("codex"), ncalls("claude"))
            dd, rr = collect("ref-%s-%s" % (env_name, "v" if value else "empty"), "haiku", env_extra={env_name: value})
            ok_ = rr.returncode != 0 and env_name in rr.stderr + rr.stdout and not os.path.exists(dd + "-work") and not os.path.exists(os.path.join(dd, "maps")) and (ncalls("codex"), ncalls("claude")) == before
            refused_ok &= ok_
            if not ok_:
                detail.append((env_name, value, rr.returncode, rr.stderr[-200:]))
    check("E collector refuses each of CLAWPATCH_MODEL, CLAWPATCH_REASONING_EFFORT, CLAWPATCH_PROVIDER, CLAWPATCH_CONFIG, CLAWPATCH_CODEX_SANDBOX (set and empty): non-zero, names the variable, no workspace, no data, no probe", refused_ok, detail)
    dc, rc_ = collect("badarm", "sonnet")
    check("E an arm outside the four is rejected by argument parsing before anything runs", rc_.returncode == 2 and "invalid choice" in rc_.stderr and not os.path.exists(os.path.join(dc, "maps")))
    before = ncalls("codex")
    dne, rne = collect("noeffort", "terra-low", env_extra={"X2C_FAKE_NO_EFFORT": "1"})
    check("C a codex before-probe that names no reasoning effort fails closed: refused with the reason, no clone, no data", rne.returncode != 0 and "no reasoning effort named" in rne.stderr + rne.stdout and ncalls("codex") == before + 1
          and not os.path.exists(os.path.join(dne + "-work", "terra-low", "tiny", "ws")) and not os.path.exists(os.path.join(dne, "maps")), rne.stderr[-300:])
    wrapper_pin = "import sys, runpy, x2c\nx2c.X2_PINS['GUIDE.md'] = '0' * 64\nsys.argv = sys.argv[1:]\nrunpy.run_path(sys.argv[0], run_name='__main__')\n"
    dpin = os.path.join(tmp, "col-pin")
    _d, r_pin = collect("pin", "luna-xhigh", wrapper=wrapper_pin, data=dpin)
    check("C the collector refuses when a reused frozen X2 file differs from its pin (GUIDE.md simulated): non-zero, names the file, no workspace, no data", r_pin.returncode != 0 and "GUIDE.md" in r_pin.stderr + r_pin.stdout
          and not os.path.exists(dpin + "-work") and not os.path.exists(os.path.join(dpin, "maps")))
    dlow, rlow = collect("lowprobe", "luna-xhigh", env_extra={"X2C_FAKE_EFFORT": "low"})
    sl = AX.served_check(dlow, "luna-xhigh", {"tiny"})
    check("C a collection whose probes serve effort low (the fake ignores the flag) completes and its served check FAILS naming the effort", rlow.returncode == 0 and not sl["ok"] and any("reasoning effort 'low'" in x for x in sl["reasons"]), sl["reasons"][:2])
    check("C the served check passes on every arm's real collector output", all(AX.served_check(d, arm, {"tiny"})["ok"] for arm, (d, _r) in runs.items()))

    # ---- run.sh end to end on collector output of one arm (fastapi and axum under X2C_DATA): replay -> pairs -> stub labels -> jev (stub endpoint) -> analysis
    spec_both = json.dumps({k: {"rev": rev, "url": src, "ext": ".py"} for k in ("fastapi", "axum")})
    dr = os.path.join(tmp, "col-runsh")
    rcs = [collect("runsh-" + k, "luna-xhigh", repo=k, data=dr, spec_=spec_both)[1].returncode for k in ("fastapi", "axum")]
    renv = dict(os.environ, X2C_DATA=dr)

    def sh_(*a, env=None):
        return subprocess.run(["bash", os.path.join(X2C, "run.sh"), *a], env=env or renv, capture_output=True, text=True)
    r_rep, r_prs = sh_("replay"), sh_("pairs")
    keys_s = C.read_jsonl(os.path.join(dr, "pairs", "key.jsonl"))
    ctl_s = C.read_jsonl(os.path.join(dr, "pairs", "control-key.jsonl"))
    lr = V.run_label(dr, os.path.join(dr, "pairs"), V.truth_labels(keys_s, lambda k: "same_concern", ctl_s)) if keys_s else None
    stub = stub_jev.Stub(x2c.JEV_MODEL, "planted").start()
    r_jev = sh_("jev", env=dict(renv, OPENROUTER_API_KEY=KEY, X2C_JEV_ENDPOINT=stub.url))
    stub.stop()
    r_an = sh_("analysis")
    res_s = json.load(open(os.path.join(dr, "results.json"))) if os.path.exists(os.path.join(dr, "results.json")) else {}
    r_rep2, r_prs2 = sh_("replay"), sh_("pairs")
    check("C run.sh end to end on collector output (fastapi + axum, stub Clawpatch, stub labels, stub Jev): replay writes the 4-arm (48-row) transitions and the replay, pairs builds candidates, jev scores them, analysis writes results.json and "
          "analysis_output.txt with the cells; replay and pairs then refuse (labels exist); the Jev header records the preregistered question version and pin",
          rcs == [0, 0] and r_rep.returncode == 0 and r_prs.returncode == 0 and lr is not None and lr.returncode == 0 and r_jev.returncode == 0 and r_an.returncode == 0 and len(json.load(open(os.path.join(dr, "transitions.json")))) == 48
          and "cells" in res_s.get("x2c", {}) and "luna-xhigh+jev" in res_s["x2c"]["cells"] and "== X2c" in open(os.path.join(dr, "analysis_output.txt")).read()
          and J.read_scores(os.path.join(dr, "jev", "scores.jsonl"))[0]["question_version"] == x2c.JEV_QUESTION_VERSION and os.path.isdir(os.path.join(dr, "jev", "questions"))
          and r_rep2.returncode == 1 and "refusing" in r_rep2.stderr and r_prs2.returncode == 1 and "refusing" in r_prs2.stderr, (rcs, r_rep.stderr[-200:], r_prs.stderr[-200:], r_jev.stderr[-300:], r_an.stderr[-400:]))
    dg = res_s.get("x2c", {}).get("jev_code_digests", [])
    dg_files = sorted(dg[0]["digests"]) if dg else []
    check("C run.sh jev records the Jev code digests before the driver runs (<data>/jev/code-digests.jsonl, one line per invocation): every non-test file of internal/decisions plus consent.go, areas.go, records.go, go.mod and go.sum; the analysis stores and renders them",
          os.path.exists(os.path.join(dr, "jev", "code-digests.jsonl")) and len(dg) == 1 and "internal/decisions/client.go" in dg_files and "internal/decisions/triage.go" in dg_files and "go.mod" in dg_files and "go.sum" in dg_files
          and not any(f.endswith("_test.go") for f in dg_files) and dg[0]["digests"] == x2c.go_code_digests() and "Jev code digests" in open(os.path.join(dr, "analysis_output.txt")).read(), dg_files)
    dxc = os.path.join(tmp, "col-runsh-data")
    os.makedirs(dxc)
    marker = os.path.join(dxc, "tool-versions-luna-xhigh-fastapi.json")
    open(marker, "w").write("{}")
    data_before = os.path.exists(x2c.DATA)
    r_col = sh_("collect", "/nonexistent/fastapi", "/nonexistent/axum", env=dict(os.environ, X2C_DATA=dxc, PATH=bins + os.pathsep + os.environ["PATH"]))
    check("C run.sh collect passes X2C_DATA to the collector: with a record planted under $X2C_DATA the first collection is refused as already recorded (naming that file) before any clone or model call, and nothing is created under the default data directory",
          r_col.returncode != 0 and marker in r_col.stderr + r_col.stdout and "already recorded" in r_col.stderr + r_col.stdout and os.path.exists(x2c.DATA) == data_before, (r_col.returncode, r_col.stderr[-300:]))


def run_x2_validation():
    r = subprocess.run([PY, os.path.join(X2, "validation", "validate.py")], capture_output=True, text=True)
    n_pass = sum(1 for line in r.stdout.splitlines() if line.startswith("PASS "))
    n_fail = sum(1 for line in r.stdout.splitlines() if line.startswith("FAIL "))
    check("X X2's own validation, run unchanged on the reused code, is green (%d PASS, %d FAIL, rc %s)" % (n_pass, n_fail, r.returncode),
          r.returncode == 0 and n_fail == 0 and n_pass >= 233 and "all validation checks passed" in r.stdout, r.stdout[-800:] + r.stderr[-400:])
    return n_pass


def main():
    skip_x2 = "--skip-x2" in sys.argv
    tmp = tempfile.mkdtemp(prefix="x2c-validate-")
    print("work dir", tmp)
    provenance_checks(tmp)
    scenario_checks(tmp)
    scorer_checks(tmp)
    collector_checks(tmp)
    n_x2 = None if skip_x2 else run_x2_validation()
    print("X2c checks: %d (plus X2's %s)" % (COUNT[0], "SKIPPED" if skip_x2 else n_x2))
    if V.FAILS:
        print("FAILED:", *V.FAILS, sep="\n  ")
        sys.exit(1)
    if skip_x2:
        print("X2c-specific checks passed; X2 validation SKIPPED: not a complete validation")
        sys.exit(2)
    print("all validation checks passed")


if __name__ == "__main__":
    main()
