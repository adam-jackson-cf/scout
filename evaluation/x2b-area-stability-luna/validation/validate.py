#!/usr/bin/env python3
"""X2b implementation validation (AGENTS.md experiment rule 4). Synthetic data and fake CLIs only: no model call, no real data.

X2b reuses X2's frozen analysis, pairs, labelling and replay by import, so validation has four parts:
  P  provenance: the reused X2 files equal their pins; the copied collector equals X2's collector plus exactly the pinned diff (collect-x2b.diff).
  X  X2's own validation (evaluation/x2-area-stability/validation/validate.py) run unchanged: the reused code is still green (its check count is reported).
  S  the planted effect / planted null / fidelity checks of X2 re-run through X2b's ONE-ARM configuration (codex only) and X2b's analysis wrapper:
     S1x planted effect (exact counts, adopt 0.75), S2x clean null (adopt 0.90), S3x label-noise null (adopt 0.45), permutation null (fallback), Go-vs-Python replay
     fidelity (S1x and a tie case), one-arm control-gate sizing (180 pairs, and the 225/226 boundary), the decision combination (adopt / fallback / inconclusive / not citable /
     X2 cross-check unavailable), the served/flags check and its rendered citability line, X2b's operating-characteristics block.
  C  the X2b collector on a stub Clawpatch and fake codex: Scout's exact argv (`--model gpt-6-luna --reasoning-effort xhigh`), the probe mirror and the recorded
     served model/effort, refusals (the five Clawpatch env names, claude arm, drift of the Go constants, a changed reused X2 file, a probe without effort),
     environment (host auth context, both timeouts, credential scrub), provenance fields, retry, then replay -> pairs -> labels -> analysis end to end, and run.sh itself
     (replay, pairs, analysis under X2B_DATA, then the append-only refusals) compared with the Python path.
Usage (repo root, toolchain active): .tools/venv/bin/python evaluation/x2b-area-stability-luna/validation/validate.py [--skip-x2 (development only; exits 2)]
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
X2B = os.path.dirname(HERE)
sys.path.insert(0, X2B)
import x2b  # noqa: E402 (puts the frozen X2 directory on sys.path)
import analysis as A  # noqa: E402 (X2's frozen analysis.py)
import common as C  # noqa: E402 (X2's frozen common.py)
import analysis_x2b as AX  # noqa: E402

ROOT = x2b.ROOT
X2 = x2b.X2_DIR
PY = sys.executable
spec_ = importlib.util.spec_from_file_location("x2_validate", os.path.join(X2, "validation", "validate.py"))
V = importlib.util.module_from_spec(spec_)
spec_.loader.exec_module(V)  # X2's helpers (synthetic maps, runners, Python reference); its main() is not run

COUNT = [0]
FLAGS = ["--model", "gpt-6-luna", "--reasoning-effort", "xhigh"]
GO_WANT = {"model": "gpt-6-luna", "reasoning_effort": "xhigh", "argv_tail_present": True, "provider_resolver_present": True}
DIFF_FILE = os.path.join(HERE, "collect-x2b.diff")


def check(name, cond, detail=""):
    COUNT[0] += 1
    V.check(name, bool(cond), detail)


def collect_diff():
    a = open(os.path.join(X2, "collect.py")).read().splitlines(keepends=True)
    b = open(os.path.join(X2B, "collect_x2b.py")).read().splitlines(keepends=True)
    return "".join(difflib.unified_diff(a, b, "x2/collect.py", "x2b/collect_x2b.py", n=1))


# ---------------------------------------------------------------- P provenance and drift guards
def provenance_checks(tmp):
    check("P every reused X2 file (analysis, common, build_pairs, label, replay, GUIDE, X2 PREREGISTRATION, validate, stub, credcheck) has the digest X2's round-10 sign-off names", not x2b.verify_pins(), x2b.verify_pins())
    check("P X2's collector, the base of collect_x2b.py, has the signed-off digest", C.sha256_file(os.path.join(X2, "collect.py")) == x2b.PINNED_COLLECT_BASE)
    check("P collect_x2b.py equals X2's collect.py plus exactly the committed diff (validation/collect-x2b.diff): every X2 safeguard outside the diff is the X2 code",
          os.path.exists(DIFF_FILE) and open(DIFF_FILE).read() == collect_diff(), "regenerate validation/collect-x2b.diff only with a new council round")
    check("P arms: X2b runs one arm (codex); X2's ARMS still lists claude (frozen) but write_transitions with arms=x2b.ARMS yields codex rows only",
          x2b.ARMS == ("codex",) and {r["arm"] for r in C.write_transitions(os.path.join(tmp, "tr.json"), arms=list(x2b.ARMS))} == {"codex"}
          and len(json.load(open(os.path.join(tmp, "tr.json")))) == 12)
    # Go drift guard: the real source states the preregistered configuration; each mutation is caught with its own reason
    check("P drift guard: internal/clawpatch/client.go + provider.go state gpt-6-luna / xhigh, append them after `--source agent` and resolve the map provider to codex only",
          x2b.check_go_config() == [] and x2b.go_map_config() == GO_WANT and x2b.map_flags() == FLAGS, x2b.go_map_config())
    csrc, psrc = open(x2b.GO_CLIENT).read(), open(x2b.GO_PROVIDER).read()
    d = os.path.join(tmp, "gomut")
    os.makedirs(d)

    def mut(name, cs=None, ps=None):
        cp, pp = os.path.join(d, name + "-client.go"), os.path.join(d, name + "-provider.go")
        open(cp, "w").write(cs if cs is not None else csrc)
        open(pp, "w").write(ps if ps is not None else psrc)
        return x2b.check_go_config(cp, pp)
    r_model = mut("model", cs=re.sub(r'(mapModel\s*=\s*)"[^"]*"', r'\1"gpt-6-sol"', csrc, 1))
    r_eff = mut("eff", cs=re.sub(r'(mapReasoningEffort\s*=\s*)"[^"]*"', r'\1"low"', csrc, 1))
    r_argv = mut("argv", cs=csrc.replace('agentArgs = append(agentArgs, "--model", mapModel, "--reasoning-effort", mapReasoningEffort)', ""))
    r_prov = mut("prov", ps=psrc.replace("return ProviderCodex, nil", 'return "claude", nil'))
    r_resolver = mut("resolver", cs=csrc.replace("c.resolveMapProvider()", "c.resolveProvider()"))
    check("P drift guard catches each mutation with its own reason (model, effort, argv line, resolver returning claude, Client.Map calling resolveProvider)",
          len(r_model) == 1 and "mapModel" in r_model[0] and len(r_eff) == 1 and "mapReasoningEffort" in r_eff[0] and len(r_argv) == 1 and "no longer appends" in r_argv[0]
          and len(r_prov) == 1 and "resolveMapProvider" in r_prov[0] and len(r_resolver) == 1 and "no longer appends" in r_resolver[0], (r_model, r_eff, r_argv, r_prov, r_resolver))
    # drift guards on the Clawpatch side: the flags and the codex provider invocation the probe mirrors
    pt = C.patched_sources(["src/providers/codex.ts", "src/cli.ts", "src/agent-mapper.ts", "src/config.ts"])
    codex_ts, cli_ts, am_ts = pt["src/providers/codex.ts"], pt["src/cli.ts"], pt["src/agent-mapper.ts"]
    check("P drift guard (codex.ts): the codex provider invocation is `exec --cd <root> --output-schema .. --output-last-message .. --sandbox read-only [--skip-git-repo-check] [codexConfig -c] --model <m> -c model_reasoning_effort=\"<e>\" -`, which x2b.codex_probe_args mirrors (minus the schema/output files)",
          'addCodexSandboxArgs(args, sandbox);\n    addCodexModelArgs(args, options);\n    args.push("-");' in codex_ts and 'args.push("--model", options.model);' in codex_ts
          and 'args.push("-c", `model_reasoning_effort="${options.reasoningEffort}"`);' in codex_ts and 'args.push("--skip-git-repo-check")' in codex_ts
          and x2b.codex_probe_args("T") == ["exec", "--cd", "T", "--sandbox", "read-only", "--skip-git-repo-check", "--model", "gpt-6-luna", "-c", 'model_reasoning_effort="xhigh"', "-"])
    check("P drift guard (cli.ts, agent-mapper.ts): `map` accepts --model and --reasoning-effort (xhigh allowed), and the agent-start progress line reports the model option (`model=<option>`), the collector's model_option_reported",
          '"reasoning-effort"' in cli_ts and "xhigh" in cli_ts and 'model: options.providerOptions.model ?? "default"' in am_ts and 'options.onProgress?.("agent-start"' in am_ts)



# ---------------------------------------------------------------- planted data helpers (one arm)
def plant_x2b(data, repos=("fastapi", "axum"), arm="codex", after_effort="xhigh", option="gpt-6-luna", flags=None, drop_flags_in=None, probe_model="gpt-6-luna"):
    """Make synthetic maps look like collector output: meta.json fields and a per-repo versions file (probe-consistent, preregistered flags)."""
    for repo in repos:
        for mid in C.MAP_IDS:
            p = os.path.join(C.map_dir(data, arm, repo, mid), "meta.json")
            m = json.load(open(p))
            m.update({"map_flags": FLAGS if flags is None else flags, "map_model": "gpt-6-luna", "map_reasoning_effort": "xhigh", "provider_reported": "codex", "model_option_reported": option})
            if drop_flags_in == (repo, mid):
                m.pop("map_flags")
            json.dump(m, open(p, "w"))
        tv = {"model_probe_before": {"model": probe_model, "reasoning_effort": "xhigh", "error": None}, "model_probe_after": {"model": probe_model, "reasoning_effort": after_effort, "error": None},
              "cli_versions_after": {"codex": "c1"}, "codex": "c1", "workspace_config_presence": {".codex": False, "clawpatch.config.json": False},
              "x2b": {"map_flags": FLAGS, "go_map_config": GO_WANT}}
        with open(os.path.join(data, "tool-versions-%s-%s.json" % (arm, repo)), "w") as f:
            json.dump(tv, f)


def clone_data(src_data, name, tmp):
    dst = os.path.join(tmp, name)
    shutil.copytree(src_data, dst)
    return dst


def volc(n_vol, t):
    return sum(1 for j in range(n_vol) if V.VOL_J[j % len(V.VOL_J)] >= t)


def scenario_checks(tmp):
    n_reg, n_vol = 32, 8  # X2's own S1 size; one arm gives 480 unique real pairs, above the pooled control gate's identical-kind minimum (226)
    s1 = {"kind": "planted", "n_reg": n_reg, "n_vol": n_vol}
    d1, r1, rp1, p1 = V.run_scenario(tmp, "S1x", s1, arms=("codex",), meta_fix=plant_x2b)
    data1 = os.path.join(d1, "data")
    rx = AX.analyze(data1, rp1, p1)
    check("S1x one arm: the replay, the key and the results carry the codex arm only; no replication block; transitions file has 12 rows",
          sorted(rx["arms"]) == ["codex"] and "replication" not in rx and {r["arm"] for r in C.read_jsonl(rp1)} == {"codex"} and {k["arm"] for k in C.read_jsonl(os.path.join(p1, "key.jsonl"))} == {"codex"}
          and len(json.load(open(os.path.join(d1, "transitions.json")))) == 12 and set(rx["controls"]["per_arm"]) == {"codex"} and rx["x2b"]["arm_check"]["unexpected_arms"] == [], sorted(rx["arms"]))
    ok = True
    for repo in ("fastapi", "axum"):
        for pct in (30, 35, 40, 45, 50, 75, 80):
            t = pct / 100
            vol = volc(n_vol, t)
            reg = 6 * n_reg if pct <= 75 else 0
            c = V.cell(rx, "codex", repo, "decisive", pct)
            ok &= c["matched"] == reg + 6 * vol and c["different"] == 6 * vol
            if pct == 30:
                ok &= V.close(c["spurious"], 6 * vol / (reg + 6 * vol))
            if pct == 50:
                ok &= c["different"] == 0 and c["survival_exact"] == "4/5"
    check("S1x planted effect: exact matched and different counts at t = 0.30..0.50, 0.75, 0.80 in both repositories (spurious 48/240 at 0.30, none and survival 4/5 at 0.50)", ok)
    dec = rx["arms"]["codex"]["decision"]
    check("S1x verdict adopt t* = 0.75, feasible 0.50..0.75 contiguous, low thresholds fail on spurious merges, controls gate ok (one arm, %d unique real pairs)" % rx["controls"]["n_real_pairs"],
          dec["verdict"] == "adopt" and dec["threshold_pct"] == 75 and dec["feasible_pct"] == list(range(50, 76, 5)) and dec["pass_set_contiguous"] and all(dec["per_threshold"][p] == "fail" for p in (30, 35, 40, 45))
          and rx["controls"]["ok"], (dec, rx["controls"]["reason"]))
    g = rx["arms"]["codex"]["gradient_exploratory"]
    check("S1x Jaccard gradient strongly negative", g["r"] is not None and g["r"] < -0.5, g)
    bad, n = V.crosscheck(d1)
    check("R S1x replay (Go) equals the independent Python reference at every threshold (%d rows, one arm)" % n, bad == 0 and n == 12 * 13, (bad, n))
    sc = rx["x2b"]["served_check"]
    comb = rx["x2b"]["combination"]
    check("S1x X2b blocks: served/flags check ok on probe/flag-consistent planted records, decision citable, proposal match_jaccard = 0.75 stated specific to gpt-6-luna/xhigh and outside X2's pass set, X2's frozen codex result read (adopt 0.30, pass set [30]), "
          "X2's 0.30 reported as `fail` in X2b (the reverse cross-check), expected labelled n per repository 125 / 365 and the five-size operating-characteristics block present",
          sc["ok"] and dec["citable"] is True and comb["proposal"].startswith("match_jaccard = 0.75") and "specific to gpt-6-luna/xhigh" in comb["proposal"] and "in X2's pass set: no" in comb["proposal"]
          and "UNAVAILABLE" not in comb["proposal"] and comb["x2_codex"]["available"] and comb["x2_codex"]["verdict"] == "adopt"
          and comb["x2_codex"]["threshold_pct"] == 30 and comb["x2_codex"]["pass_pct"] == [30] and comb["x2b_t_star_in_x2_pass_set"] is False and comb["x2_t_star_status_in_x2b"] == "fail"
          and rx["x2b"]["expected_labelled_n_per_repo"] == {"fastapi": 125, "axum": 365} and sorted(rx["x2b"]["operating_characteristics"]["by_n"], key=int) == ["100", "125", "150", "216", "270"]
          and rx["x2b"]["operating_characteristics"]["by_n"]["125"]["n_per_repo"] == 125, (sc["reasons"], comb))
    text = AX.render(rx)
    check("S1x rendered report carries the X2 tables, the citability line (yes, probe-consistency ok, served/flags ok), the served/flags line, X2b's expected labelled n and operating characteristics (n=100 and 125), "
          "the combination line with X2's t* status in X2b, and the proposal; no `replication:` line",
          all(s in text for s in ("== X2b", "served/flags check: ok", "combination: X2b verdict adopt, t* 75", "PROPOSAL (for a design round, not adopted here): match_jaccard = 0.75", "AT t*=75", "controls: pooled across arms",
                                  "citable for a design change (X2b section 4): yes -- probe-consistency ok; served/flags ok; unexpected arms []", "X2b expected labelled n per repository ~ {\"axum\": 365, \"fastapi\": 125}",
                                  "the '216-270' band printed above is X2's, not X2b's", "X2b operating characteristics (n=100 per repo", "X2b operating characteristics (n=125 per repo",
                                  "X2's t* status in X2b (per-threshold at X2's t*): fail"))
          and text.index("citable for a design change (X2b section 4)") > text.index("== arm codex:") and "\nreplication:" not in text, text[-900:])

    # ---- served/flags check negatives, each on a copy of S1x's data
    def variant(name, **kw):
        dd = clone_data(data1, name, tmp)
        plant_x2b(dd, **kw)
        return AX.analyze(dd, rp1, p1)
    v_eff = variant("S1x-eff", after_effort="low")
    v_mod = variant("S1x-model", probe_model="gpt-6.1-sol")
    v_opt = variant("S1x-opt", option="default")
    v_flag = variant("S1x-flag", flags=["--model", "gpt-6.1-sol", "--reasoning-effort", "low"])
    v_drop = variant("S1x-drop", drop_flags_in=("axum", "M1"))
    for nm, v, frag in (("an after-probe serving effort low", v_eff, "reasoning effort 'low'"), ("probes serving model gpt-6.1-sol", v_mod, "served model 'gpt-6.1-sol'"),
                        ("Clawpatch agent-start reporting model=default", v_opt, "agent-start reported"), ("maps recorded with other flags", v_flag, "meta.json records flags"),
                        ("one map without flags", v_drop, "axum/M1")):
        d_ = v["arms"]["codex"]["decision"]
        vtext = AX.render(v)
        citable_line = next((l for l in vtext.split("\n") if "citable for a design change (X2b section 4)" in l), "")
        check("S1x served/flags check: %s -> check FAILED naming it, the verdict unchanged (adopt 75) but NOT citable, proposal `none: not citable`, and the rendered citability line says NO with `served/flags FAILED` naming the reason (not X2's `arm probe-inconsistent`)" % nm,
              not v["x2b"]["served_check"]["ok"] and any(frag in r for r in v["x2b"]["served_check"]["reasons"]) and d_["verdict"] == "adopt" and d_["threshold_pct"] == 75 and d_["citable"] is False
              and v["x2b"]["combination"]["proposal"].startswith("none: not citable")
              and "(X2b section 4): NO -- probe-consistency ok; served/flags FAILED" in citable_line and frag in citable_line and "served/flags check: FAILED" in vtext, (v["x2b"]["served_check"]["reasons"][:2], citable_line))
    # no versions file at all
    dn = clone_data(data1, "S1x-noversions", tmp)
    os.remove(os.path.join(dn, "tool-versions-codex-axum.json"))
    vn = AX.analyze(dn, rp1, p1)
    vn_line = next((l for l in AX.render(vn).split("\n") if "citable for a design change (X2b section 4)" in l), "")
    check("S1x served/flags check: a missing versions file -> FAILED, not citable, and the rendered citability line says NO and names the missing versions file", not vn["x2b"]["served_check"]["ok"] and any("axum: no tool-versions file" in r for r in vn["x2b"]["served_check"]["reasons"]) and vn["arms"]["codex"]["decision"]["citable"] is False
          and "(X2b section 4): NO" in vn_line and "no tool-versions file" in vn_line, vn_line)
    # a failed map is not asked to have reported the model option; a skipped one is ignored
    dfl = clone_data(data1, "S1x-failedmap", tmp)
    pm = os.path.join(C.map_dir(dfl, "codex", "axum", "F2"), "meta.json")
    mm = json.load(open(pm))
    mm.update({"rc": 1, "model_option_reported": None, "provider_reported": None})
    json.dump(mm, open(pm, "w"))
    pm2 = os.path.join(C.map_dir(dfl, "codex", "axum", "R2"), "meta.json")
    mm2 = json.load(open(pm2))
    mm2.update({"rc": None, "skipped": "test"})
    mm2.pop("map_flags")
    json.dump(mm2, open(pm2, "w"))
    check("S1x served/flags check: a failed map (rc != 0) carries no agent-start model and a skipped map no flags; neither makes the check fail", AX.served_check(dfl, {"fastapi", "axum"})["ok"])
    # unexpected arm in the replay
    d2c, _rc, rpc, pc = V.run_scenario(tmp, "S1x-two-arms", {"kind": "planted", "n_reg": 32, "n_vol": 8}, arms=("codex", "claude"), meta_fix=plant_x2b)
    r_two = AX.analyze(os.path.join(d2c, "data"), rpc, pc)
    two_line = next((l for l in AX.render(r_two).split("\n") if "citable for a design change (X2b section 4)" in l), "")
    check("S1x an unexpected second arm in the replay is reported and makes the codex decision not citable; the rendered citability line names the unexpected arm", r_two["x2b"]["arm_check"]["unexpected_arms"] == ["claude"]
          and r_two["arms"]["codex"]["decision"]["citable"] is False and "(X2b section 4): NO" in two_line and 'unexpected arms ["claude"]' in two_line, two_line)
    # X2 results file integrity
    x2d = os.path.join(tmp, "x2copy")
    os.makedirs(os.path.join(x2d, "data"))
    shutil.copy(os.path.join(X2, "data", "results.json"), os.path.join(x2d, "data", "results.json"))
    with open(os.path.join(x2d, "data", "results.json"), "a") as f:
        f.write(" ")
    xt = AX.x2_codex_result(x2d)
    check("S1x X2's frozen results file is read only when its digest equals the pin (a changed file -> unavailable, reason names the digests); absent -> unavailable",
          xt["available"] is False and "pinned" in xt["reason"] and AX.x2_codex_result(os.path.join(tmp, "nope"))["available"] is False and AX.x2_codex_result()["available"] is True)

    # ---- decision combination for every verdict (X2b PREREGISTRATION section 4a)
    d5, r5, rp5, p5 = V.run_scenario(tmp, "S5x", {"kind": "unrelated", "n_reg": 60}, arms=("codex",), meta_fix=plant_x2b)
    r5x = AX.analyze(os.path.join(d5, "data"), rp5, p5)
    d6, r6, rp6, p6 = V.run_scenario(tmp, "S6x", {"kind": "clean", "n_reg": 0}, arms=("codex",), meta_fix=plant_x2b)
    r6x = AX.analyze(os.path.join(d6, "data"), rp6, p6)
    c5, c6 = r5x["x2b"]["combination"], r6x["x2b"]["combination"]
    check("S5x unrelated maps: verdict fallback, proposal = the design fallback naming the 0.30-0.90 grid and that nothing below 0.30 was measured, X2's adopt not applied",
          r5x["arms"]["codex"]["decision"]["verdict"] == "fallback" and c5["proposal"].startswith("no threshold on the 0.30-0.90 grid meets C1-C3 for gpt-6-luna xhigh (nothing below 0.30 measured)")
          and "X2's codex adopt is not applied" in c5["proposal"] and c5["x2b_t_star_in_x2_pass_set"] is None, (r5x["arms"]["codex"]["decision"], c5))
    # combine() unit cases: `adopt` at the grid minimum, and an unavailable X2 cross-check (also through analyze + render with a missing X2 directory)
    x2_ok = AX.x2_codex_result()
    c_min = AX.combine({"verdict": "adopt", "threshold_pct": 30, "per_threshold": {30: "pass"}}, True, {"ok": True}, x2_ok)
    c_in = AX.combine({"verdict": "adopt", "threshold_pct": 30, "per_threshold": {30: "pass"}}, True, {"ok": True}, {"available": False, "reason": "file absent"})
    rn = AX.analyze(data1, rp1, p1, x2_dir=os.path.join(tmp, "no-x2-here"))
    tn = AX.render(rn)
    check("S1x combination strings: `adopt` at 0.30 says specific to gpt-6-luna/xhigh, in X2's pass set: yes and `grid minimum`; an unavailable X2 cross-check appends `X2 cross-check UNAVAILABLE (<reason>)` to the proposal (unit, and through analyze and render), keeps the verdict and still states `in X2's pass set: unknown`",
          c_min["proposal"].startswith("match_jaccard = 0.30 (X2b t*; specific to gpt-6-luna/xhigh; in X2's pass set: yes; grid minimum") and c_min["x2b_t_star_in_x2_pass_set"] is True
          and c_in["proposal"].endswith("X2 cross-check UNAVAILABLE (file absent)") and "in X2's pass set: unknown" in c_in["proposal"]
          and rn["x2b"]["combination"]["x2_codex"] == {"available": False, "reason": "file absent"} and rn["x2b"]["combination"]["proposal"].endswith("X2 cross-check UNAVAILABLE (file absent)")
          and rn["arms"]["codex"]["decision"]["verdict"] == "adopt" and "X2 cross-check UNAVAILABLE (file absent)" in tn
          and rn["x2b"]["combination"]["x2_t_star_status_in_x2b"] == "fail", (c_min["proposal"], c_in["proposal"], rn["x2b"]["combination"]["proposal"]))
    check("S6x no prior areas: verdict inconclusive, proposal keeps 0.5 and does not adopt X2's 0.30", r6x["arms"]["codex"]["decision"]["verdict"] == "inconclusive" and c6["proposal"].startswith("none: inconclusive")
          and "match_jaccard stays 0.5" in c6["proposal"] and "X2's 0.30 is not adopted" in c6["proposal"], c6)

    # ---- planted nulls (single arm)
    d2, r2, rp2, _p2 = V.run_scenario(tmp, "S2x", {"kind": "clean", "n_reg": 40}, arms=("codex",), meta_fix=plant_x2b)
    r2x = AX.analyze(os.path.join(d2, "data"), rp2, _p2)
    ok2 = all(V.cell(r2x, "codex", repo, "decisive", pct)["different"] == 0 and V.close(V.cell(r2x, "codex", repo, "decisive", pct)["survival"], 1.0) for repo in ("fastapi", "axum") for pct in C.GRID_PCT)
    dec2 = r2x["arms"]["codex"]["decision"]
    check("S2x planted null (identical re-map): survival 1 and no spurious merge at every threshold, adopt 0.90, proposal 0.90", ok2 and dec2["verdict"] == "adopt" and dec2["threshold_pct"] == 90
          and r2x["x2b"]["combination"]["proposal"].startswith("match_jaccard = 0.90"), dec2)

    def rule3(k):
        return "different_concern" if int(hashlib.sha256(("noise" + k["pair_id"]).encode()).hexdigest()[:8], 16) % 100 < 2 else "same_concern"
    d3, r3, rp3, p3 = V.run_scenario(tmp, "S3x", {"kind": "spread", "n_reg": 40}, rule=rule3, arms=("codex",), meta_fix=plant_x2b)
    r3x = AX.analyze(os.path.join(d3, "data"), rp3, p3)
    flat = counts_ok = True
    for repo in ("fastapi", "axum"):
        for p in C.GRID_PCT:
            c = V.cell(r3x, "codex", repo, "decisive", p)
            counts_ok &= c["matched"] == 6 * 5 * sum(1 for x in V.SPREAD_J if x >= p / 100.0)
            if c["labelled"] >= 60:
                flat &= abs(c["spurious"] - 0.02) <= 4 * math.sqrt(0.02 * 0.98 / c["labelled"])
    d3c = r3x["arms"]["codex"]["decision"]
    check("S3x planted null (2 % label noise independent of a Jaccard spread 0.38-1.0): matched counts equal the planted spread, rate within binomial bounds, gradient ~0, planned verdict adopt 0.45",
          counts_ok and flat and abs(r3x["arms"]["codex"]["gradient_exploratory"]["r"]) < 0.2 and d3c["verdict"] == "adopt" and d3c["threshold_pct"] == 45 and d3c["feasible_pct"] == list(range(30, 46, 5)), d3c)
    # permutation null at decision level
    keys, ctl = C.read_jsonl(os.path.join(p1, "key.jsonl")), C.read_jsonl(os.path.join(p1, "control-key.jsonl"))
    labels = C.final_labels(p1)
    pids = sorted({k["pair_id"] for k in keys})
    verdicts, spur = [], []
    for seed in range(6):
        rnd = random.Random(20261023 + seed)
        vals = [labels[p] for p in pids]
        rnd.shuffle(vals)
        perm = dict(zip(pids, vals))
        pd = os.path.join(tmp, "S1xperm%d" % seed)
        os.makedirs(pd)
        for fn in ("blinded.jsonl", "key.jsonl", "control-key.jsonl"):
            shutil.copy(os.path.join(p1, fn), pd)
        rr = V.run_label(tmp, pd, V.truth_labels(keys, lambda k, perm=perm: perm[k["pair_id"]], ctl))
        assert rr.returncode == 0, rr.stderr
        rres = AX.analyze(data1, rp1, pd)
        verdicts.append(rres["arms"]["codex"]["decision"]["verdict"])
        spur.append(V.cell(rres, "codex", "fastapi", "decisive", 50)["spurious"])
    check("P permutation null: S1x labels permuted over 6 seeds -> fallback every time, spurious ~0.2 at 0.50", set(verdicts) == {"fallback"} and 0.1 < sum(spur) / len(spur) < 0.3, (verdicts, spur))
    # replay fidelity on a cross-area tie
    dt, rt, rpt, pt_ = V.run_scenario(tmp, "Tx", {"kind": "tie", "n_reg": 20}, arms=("codex",))
    badt, nt = V.crosscheck(dt, go_order=True)
    check("R tie case (one arm): the Go replay equals the Python reference with Go's tie order at every threshold (%d rows)" % nt, badt == 0 and nt == 12 * 13, (badt, nt))
    # one-arm control-gate sizing: below ~226 unique real pairs the identical-control kind cannot reach its 10-usable minimum (X2 section 4 gate 3)
    ds, rs_, rps, ps_ = V.run_scenario(tmp, "S1x-small", {"kind": "planted", "n_reg": 12, "n_vol": 3}, arms=("codex",))
    cs = rs_["controls"]
    check("S1x one-arm sizing: with 180 unique real pairs (n_reg 12) the identical-kind minimum (226) cannot be met -> the control gate fails, its reason names the identical controls, and the verdict is inconclusive naming the gate; at X2's per-arm size (480 pairs, above) it passes",
          cs["n_real_pairs"] < 226 and not cs["ok"] and "identical controls" in cs["reason"] and rs_["arms"]["codex"]["decision"]["verdict"] == "inconclusive" and rs_["arms"]["codex"]["decision"]["tripped_gate"] == "control gate",
          (cs["n_real_pairs"], cs["reason"]))
    # the exact boundary on a synthetic pairs directory (A.control_stats is X2's frozen code): 225 unique real pairs generate 9 identical controls, below the 10-usable minimum, so the gate
    # fails with every control usable; 226 generate 10, and the gate passes when all 10 are usable but fails again when one of them is `cannot_tell`
    def boundary(n_real, bad_identical=0):
        pd_ = os.path.join(tmp, "boundary-%d-%d" % (n_real, bad_identical))
        os.makedirs(pd_)
        keys_ = [{"pair_id": "r%04d" % i, "jaccard": 0.5, "arm": "codex"} for i in range(n_real)]
        n_ident, n_swap = math.ceil(C.CONTROL_IDENT_FRACTION * n_real), math.ceil(C.CONTROL_SWAP_FRACTION * n_real)
        ctl_ = [{"pair_id": "ci%03d" % i, "kind": "identical", "expected": "same_concern", "arm": "codex", "source_pair_id": "r%04d" % i} for i in range(n_ident)]
        ctl_ += [{"pair_id": "cs%03d" % i, "kind": "swapped", "expected": "different_concern", "arm": "codex", "source_pair_id": "r%04d" % i} for i in range(n_swap)]
        for fn, rows in (("key.jsonl", keys_), ("control-key.jsonl", ctl_), ("calls.jsonl", [])):
            with open(os.path.join(pd_, fn), "w") as f:
                f.write("".join(json.dumps(r) + "\n" for r in rows))
        labs = {c["pair_id"]: c["expected"] for c in ctl_}
        for i in range(bad_identical):
            labs["ci%03d" % i] = "cannot_tell"
        return A.control_stats(pd_, labs), n_ident
    b225, n225 = boundary(225)
    b226, n226 = boundary(226)
    b226x, _ = boundary(226, bad_identical=1)
    check("S1x one-arm sizing boundary (X2's control_stats on a synthetic pairs directory): 225 unique real pairs -> %d identical controls, all usable, gate FAILS (minimum 10); 226 -> %d, all usable, gate passes; "
          "226 with one identical control `cannot_tell` fails again (226 is necessary, not sufficient)" % (n225, n226),
          n225 == 9 and not b225["ok"] and "identical controls" in b225["reason"] and n226 == 10 and b226["ok"] and not b226x["ok"] and "identical controls: 9 usable (minimum 10 of 10 generated)" in b226x["reason"],
          (b225["reason"], b226["reason"], b226x["reason"]))


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
    shutil.copy(os.path.join(HERE, "stub_clawpatch_x2b.js"), stub)
    open(os.path.join(sb, "dist", "prompt.js"), "w").write("// fixture: %s\n" % C.SANITIZER_TOKEN)
    open(os.path.join(sb, C.BUILD_MARKER_FILE), "w").write("%s %s\n" % (pv_, ps_))
    bins = os.path.join(tmp, "fakebin")
    os.makedirs(bins)
    # a fake codex that, like the real CLI, serves what its arguments and "user config" say: --model / -c model_reasoning_effort win; without them it serves the
    # config's model and effort (the operator's ~/.codex/config.toml: gpt-6.1-sol, low), and --ignore-user-config changes the model (so a probe that drops the flags would be detected)
    body = ('import json, os, sys\na = sys.argv[1:]\nif a[:1] == ["--version"]:\n    print("codex-cli 0.0.0-fake")\n    sys.exit(0)\nsys.stdin.read()\n'
            'open(os.path.abspath(__file__) + ".calls", "a").write(json.dumps({"argv": a, "env": sorted(os.environ)}) + "\\n")\n'
            'model = a[a.index("--model") + 1] if "--model" in a else "gpt-6.1-sol"\neff = "low"\n'
            'for i, x in enumerate(a):\n    if x == "-c" and a[i + 1].startswith("model_reasoning_effort="):\n        eff = a[i + 1].split("=", 1)[1].strip(chr(34))\n'
            'if "--ignore-user-config" in a:\n    model += "-isolated"\n'
            'print("model: " + model)\n'
            'if not os.environ.get("X2B_FAKE_NO_EFFORT"):\n    print("reasoning effort: " + (os.environ.get("X2B_FAKE_EFFORT") or eff))\n')
    cp = os.path.join(bins, "codex")
    open(cp, "w").write("#!%s\n%s" % (PY, body))
    os.chmod(cp, 0o755)
    poison = {"ANTHROPIC_API_KEY": "k", "OPENAI_API_KEY": "k", "MY_SECRET_THING": "k", "GITHUB_TOKEN": "k", "OPENROUTER_FOO": "k", "OPENAI_BASE_URL": "k", "X2_PLAIN": "keep",
              "CLAWPATCH_CLAUDE_AUTH_CONTEXT": "isolated"}
    calls = os.path.join(bins, "codex.calls")

    def ncalls():
        return len(open(calls).read().splitlines()) if os.path.exists(calls) else 0

    def collect(name, fail_calls="", env_extra=None, args=(), arm="codex", wrapper=None, repo="tiny", data=None, work=None, spec_=None):
        d = data or os.path.join(tmp, "col-" + name)
        env = dict(os.environ, X2_STUB_COUNTER=d + "-counter", X2_STUB_FAIL_CALLS=fail_calls, X2_STUB_ENVLOG=d + "-envlog.jsonl", X2B_STUB_ARGVLOG=d + "-argv.jsonl",
                   X2B_BASE_STUB=os.path.join(X2, "validation", "stub_clawpatch.js"), **poison)
        env["PATH"] = bins + os.pathsep + os.environ["PATH"]
        for n_ in (*C.TIMEOUT_ENV_NAMES, *C.REFUSED_MODEL_ENV, *C.REFUSED_SANDBOX_ENV, "CLAWPATCH_CLAUDE_BIN", "CODEX_HOME", "X2B_FAKE_NO_EFFORT", "X2B_FAKE_EFFORT"):
            env.pop(n_, None)
        env.update(env_extra or {})
        cmd = [PY, os.path.join(X2B, "collect_x2b.py"), "--arm", arm, "--repo", repo, "--source", src, "--work", work or d + "-work", "--data", d, "--node", node, "--cli", stub,
               "--spec", spec_ or spec, "--expect-clawpatch-version", "stub-0", "--term-grace", "2", *args]
        if wrapper:
            cmd = [PY, "-c", "import sys; sys.path.insert(0, %r)\n" % X2B + wrapper, *cmd[1:]]
        return d, subprocess.run(cmd, env=env, capture_output=True, text=True)

    # ---- normal run: one failing call (F2's first attempt) -> retry
    d, r = collect("ok", "2")
    check("C collector exits 0 (7 maps, F2 retried once)", r.returncode == 0, r.stderr[-800:])
    metas = {m: C.read_meta(d, "codex", "tiny", m) for m in C.MAP_IDS}
    check("C all 7 maps recorded with rc 0; F2 has 2 attempts and the second ran after the first failed", all(v is not None and v["rc"] == 0 for v in metas.values()) and metas["F2"]["attempt_count"] == 2
          and all(metas[m]["attempt_count"] == 1 for m in C.MAP_IDS if m != "F2"), {k: (v or {}).get("rc") for k, v in metas.items()})
    argv_log = [json.loads(line)["args"] for line in open(d + "-argv.jsonl")]
    maps_ = [a for a in argv_log if "map" in a]
    inits = [a for a in argv_log if "init" in a]
    want_tail = ["--json", "map", "--source", "agent", "--provider", "codex", "--model", "gpt-6-luna", "--reasoning-effort", "xhigh"]
    check("E every map invocation (8 = 7 maps + 1 retry) is exactly Scout's argv `--root R --state-dir S --json map --source agent --provider codex --model gpt-6-luna --reasoning-effort xhigh` (client.go Map + resolveMapProvider); init calls carry no model flags",
          len(maps_) == 8 and all(a[0] == "--root" and a[2] == "--state-dir" and a[4:] == want_tail for a in maps_) and len(inits) >= 4 and all("--model" not in a and "--reasoning-effort" not in a for a in inits), maps_[:1])
    check("E the flags in the argv equal the Go constants (derived at run time, not copied): x2b.map_flags() == the tail",
          [x for x in maps_[0][-4:]] == x2b.map_flags() == ["--model", x2b.go_map_config()["model"], "--reasoning-effort", x2b.go_map_config()["reasoning_effort"]])
    log = [json.loads(line) for line in open(d + "-envlog.jsonl")]
    names = sorted({n for e in log for n in e["names"]})
    cred = subprocess.run([V.go_bin(tmp, "credcheck", "./evaluation/x2-area-stability/validation/credcheck")], input="\n".join(names) + "\n", capture_output=True, text=True).stdout
    leaked = [line.split("\t")[0] for line in cred.splitlines() if line.endswith("\t1")]
    check("E every Clawpatch call (%d) runs with the host auth context, BOTH provider timeouts at 900000 ms and no credential name; non-credentials are kept" % len(log),
          len(log) >= 12 and not leaked and all(e["claudeAuth"] == "host" and e["codexTimeout"] == "900000" and e["claudeTimeout"] == "900000" for e in log) and "X2_PLAIN" in names
          and not any(n in names for n in ("ANTHROPIC_API_KEY", "OPENAI_API_KEY", "MY_SECRET_THING", "GITHUB_TOKEN", "OPENROUTER_FOO", "OPENAI_BASE_URL")), (leaked, names[:5]))
    check("C every map's meta.json records schema x2b-map-v1, the flags, model and effort, provider codex and Clawpatch's agent-start model option gpt-6-luna (the option, not the served model)",
          all(v["schema"] == "x2b-map-v1" and v["map_flags"] == FLAGS and v["map_model"] == "gpt-6-luna" and v["map_reasoning_effort"] == "xhigh" and v["provider_reported"] == "codex"
              and v["model_option_reported"] == "gpt-6-luna" and v["workspace_dirty"] is False and all(re.fullmatch(r"\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ", x["started_utc"]) for x in v["attempts"]) for v in metas.values()),
          {k: v.get("model_option_reported") for k, v in metas.items()})
    probe_calls = [json.loads(line) for line in open(calls)]
    want_probe = lambda dd: ["exec", "--cd", dd, "--sandbox", "read-only", "--skip-git-repo-check", "--model", "gpt-6-luna", "-c", 'model_reasoning_effort="xhigh"', "-"]
    tv = json.load(open(os.path.join(d, "tool-versions-codex-tiny.json")))
    check("C the before- and after-probe are the codex provider invocation for Scout's options (`exec --cd <tmp> --sandbox read-only --skip-git-repo-check --model gpt-6-luna -c model_reasoning_effort=\"xhigh\" -`), without --ignore-user-config/--ignore-rules, under Clawpatch's scrubbed environment",
          len(probe_calls) == 2 and all(p["argv"] == want_probe(p["argv"][2]) for p in probe_calls) and not any("OPENAI_API_KEY" in p["env"] or "ANTHROPIC_API_KEY" in p["env"] or "GITHUB_TOKEN" in p["env"] for p in probe_calls)
          and "X2_PLAIN" in probe_calls[0]["env"], probe_calls[:1])
    check("C versions file records the served model AND effort of both probes (gpt-6-luna, xhigh: the fake serves the flags, not its config's gpt-6.1-sol/low), the x2b block (flags, Go configuration, client.go/provider.go/collect_x2b.py digests, the X2 pins) and the standard X2 provenance",
          tv["model_probe_before"]["model"] == tv["model_probe_after"]["model"] == "gpt-6-luna" and tv["model_probe_before"]["reasoning_effort"] == tv["model_probe_after"]["reasoning_effort"] == "xhigh"
          and tv["x2b"]["map_flags"] == FLAGS and tv["x2b"]["go_map_config"] == GO_WANT and tv["x2b"]["probe_argv_tail"] == want_probe("<tmp>") and tv["x2b"]["client_go_sha256"] == C.sha256_file(x2b.GO_CLIENT)
          and tv["x2b"]["provider_go_sha256"] == C.sha256_file(x2b.GO_PROVIDER) and tv["x2b"]["collect_x2b_sha256"] == C.sha256_file(os.path.join(X2B, "collect_x2b.py")) and tv["x2b"]["x2_pins"] == x2b.X2_PINS
          and tv["clawpatch_build_safety"]["ok"] is True and tv["codex_timeout_ms"] == tv["claude_timeout_ms"] == "900000" and tv["call_timeout_seconds"] == 1800 and tv["wall_max_seconds"] == 8 * 3600
          and tv["areas_go_sha256"] == C.sha256_file(os.path.join(ROOT, "internal", "areas", "areas.go")) and tv["cli_versions_after"]["codex"] == tv["codex"], tv["model_probe_before"])
    # sanity: the fake discriminates (so the check above would fail for a probe that dropped the flags)
    plain = subprocess.run([cp, "exec", "-"], input="x", capture_output=True, text=True).stdout
    iso = subprocess.run([cp, "exec", "--ignore-user-config", "-"], input="x", capture_output=True, text=True).stdout
    check("C the fake codex is flag-sensitive (no flags: gpt-6.1-sol / low; --ignore-user-config: another model), so the probe checks discriminate", "model: gpt-6.1-sol\n" in plain and "reasoning effort: low" in plain and "gpt-6.1-sol-isolated" in iso)
    # append-only and attempts
    r_again = collect("ok", "", data=d)[1]
    check("C append-only: a second collection into the same data directory is refused before any model call and before cloning", r_again.returncode != 0 and "already recorded" in r_again.stderr + r_again.stdout)

    # ---- refusals: every refusal happens before the workspace, any data and any probe call
    refused_ok, detail = True, []
    for env_name in (*C.REFUSED_MODEL_ENV, *C.REFUSED_SANDBOX_ENV):
        for value in ("x", ""):
            before = ncalls()
            dd, rr = collect("ref-%s-%s" % (env_name, "v" if value else "empty"), env_extra={env_name: value})
            ok_ = rr.returncode != 0 and env_name in rr.stderr + rr.stdout and not os.path.exists(dd + "-work") and not os.path.exists(os.path.join(dd, "maps")) and ncalls() == before
            refused_ok &= ok_
            if not ok_:
                detail.append((env_name, value, rr.returncode, rr.stderr[-200:]))
    check("E collector refuses each of CLAWPATCH_MODEL, CLAWPATCH_REASONING_EFFORT, CLAWPATCH_PROVIDER, CLAWPATCH_CONFIG, CLAWPATCH_CODEX_SANDBOX (set to a value and set empty): non-zero, names the variable, no workspace, no data, no probe call",
          refused_ok and set(C.REFUSED_MODEL_ENV) == {"CLAWPATCH_MODEL", "CLAWPATCH_REASONING_EFFORT", "CLAWPATCH_PROVIDER", "CLAWPATCH_CONFIG"} and C.REFUSED_SANDBOX_ENV == ("CLAWPATCH_CODEX_SANDBOX",), detail)
    dc, rc_ = collect("claude", arm="claude")
    check("E the claude arm does not exist in X2b: --arm claude is rejected by argument parsing before anything runs", rc_.returncode == 2 and "invalid choice" in rc_.stderr and not os.path.exists(os.path.join(dc, "maps")))
    before = ncalls()
    dne, rne = collect("noeffort", env_extra={"X2B_FAKE_NO_EFFORT": "1"})
    check("C a before-probe that names no reasoning effort fails closed: refused with the reason, no workspace clone, no data", rne.returncode != 0 and "no reasoning effort named" in rne.stderr + rne.stdout
          and not os.path.exists(os.path.join(dne + "-work", "codex", "tiny", "ws")) and not os.path.exists(os.path.join(dne, "maps")) and ncalls() == before + 1, rne.stderr[-300:])
    dmf = os.path.join(tmp, "col-drift")
    wrapper_drift = ("import sys, runpy, x2b, tempfile, re\n"
                     "t = tempfile.mkdtemp(); p = t + '/client.go'\n"
                     "open(p, 'w').write(re.sub(r'(mapModel\\s*=\\s*)\"[^\"]*\"', r'\\1\"gpt-6-sol\"', open(x2b.GO_CLIENT).read(), 1))\n"
                     "x2b.GO_CLIENT = p\n"
                     "sys.argv = sys.argv[1:]\nrunpy.run_path(sys.argv[0], run_name='__main__')\n")
    d_, r_drift = collect("drift", wrapper=wrapper_drift, data=dmf)
    check("C the collector refuses when Scout's Go mapModel drifted from gpt-6-luna (derived constants, not copies): non-zero, names mapModel, no workspace, no data", r_drift.returncode != 0 and "drifted" in r_drift.stderr + r_drift.stdout
          and "mapModel" in r_drift.stderr + r_drift.stdout and not os.path.exists(dmf + "-work") and not os.path.exists(os.path.join(dmf, "maps")), r_drift.stderr[-300:])
    wrapper_pin = ("import sys, runpy, x2b\nx2b.X2_PINS['GUIDE.md'] = '0' * 64\nsys.argv = sys.argv[1:]\nrunpy.run_path(sys.argv[0], run_name='__main__')\n")
    dpin = os.path.join(tmp, "col-pin")
    d_, r_pin = collect("pin", wrapper=wrapper_pin, data=dpin)
    check("C the collector refuses when a reused frozen X2 file differs from its pin (GUIDE.md simulated): non-zero, names the file, no workspace, no data", r_pin.returncode != 0 and "GUIDE.md" in r_pin.stderr + r_pin.stdout
          and not os.path.exists(dpin + "-work") and not os.path.exists(os.path.join(dpin, "maps")), r_pin.stderr[-300:])

    # ---- the served/flags check on a real collector run whose probe serves another effort
    dlow, rlow = collect("lowprobe", env_extra={"X2B_FAKE_EFFORT": "low"})
    tvl = json.load(open(os.path.join(dlow, "tool-versions-codex-tiny.json"))) if os.path.exists(os.path.join(dlow, "tool-versions-codex-tiny.json")) else {}
    sl = AX.served_check(dlow, {"tiny"})
    check("C a collection whose probes serve effort low (the fake ignores the flag) completes and its served/flags check FAILS naming the effort", rlow.returncode == 0 and tvl["model_probe_before"]["reasoning_effort"] == "low"
          and not sl["ok"] and any("reasoning effort 'low'" in x for x in sl["reasons"]), sl["reasons"][:2])

    # ---- end to end on the real collector output: replay -> pairs -> labels -> X2b analysis
    sc_ = AX.served_check(d, {"tiny"})
    trans = os.path.join(tmp, "e2e-transitions.json")
    C.write_transitions(trans, arms=list(x2b.ARMS), repos=["tiny"])
    rp = os.path.join(tmp, "e2e-replay.jsonl")
    V.run_replay(tmp, os.path.join(d, "maps"), trans, rp)
    pdir = os.path.join(tmp, "e2e-pairs")
    os.makedirs(pdir)
    subprocess.run([PY, os.path.join(X2, "build_pairs.py"), "--data", d, "--replay", rp, "--out", pdir], check=True, capture_output=True)
    keys, ctl = C.read_jsonl(os.path.join(pdir, "key.jsonl")), C.read_jsonl(os.path.join(pdir, "control-key.jsonl"))
    lr = V.run_label(d, pdir, V.truth_labels(keys, lambda k: "same_concern", ctl))
    res = AX.analyze(d, rp, pdir)
    dec = res["arms"]["codex"]["decision"]
    pc = res["probe_consistency"]["codex"]
    check("C end to end on the collector's own output: replay -> pairs -> labels -> X2b analysis; served/flags check ok, the arm probe-consistent (model gpt-6-luna), no replication block, the combination proposal follows the verdict",
          sc_["ok"] and lr.returncode == 0 and sorted(res["arms"]) == ["codex"] and pc["consistent"] and pc["models"] == ["gpt-6-luna"] and "replication" not in res
          and res["x2b"]["combination"]["x2b_verdict"] == dec["verdict"] and dec["citable"] is True
          and ((dec["verdict"] == "adopt" and res["x2b"]["combination"]["proposal"].startswith("match_jaccard = ")) or dec["verdict"] != "adopt"), (sc_["reasons"], lr.stderr[-300:], dec))

    # ---- run.sh itself on the collector's own output (fastapi and axum under X2B_DATA): replay -> pairs -> (stub) labels -> analysis, then the append-only refusals
    spec_both = json.dumps({k: {"rev": rev, "url": src, "ext": ".py"} for k in ("fastapi", "axum")})
    dr = os.path.join(tmp, "col-runsh")
    rc_cols = [collect("runsh-" + k, repo=k, data=dr, spec_=spec_both)[1].returncode for k in ("fastapi", "axum")]
    runsh = os.path.join(X2B, "run.sh")
    renv = dict(os.environ, X2B_DATA=dr)

    def sh_(*a):
        return subprocess.run(["bash", runsh, *a], env=renv, capture_output=True, text=True)
    r_rep, r_prs = sh_("replay"), sh_("pairs")
    prs = os.path.join(dr, "pairs")
    keys_s = C.read_jsonl(os.path.join(prs, "key.jsonl")) if os.path.exists(os.path.join(prs, "key.jsonl")) else []
    ctl_s = C.read_jsonl(os.path.join(prs, "control-key.jsonl")) if os.path.exists(os.path.join(prs, "control-key.jsonl")) else []
    lr_s = V.run_label(dr, prs, V.truth_labels(keys_s, lambda k: "same_concern", ctl_s)) if keys_s else None  # run.sh label calls claude and codex: the same label.py, stubbed
    r_an = sh_("analysis")
    res_s = json.load(open(os.path.join(dr, "results.json"))) if os.path.exists(os.path.join(dr, "results.json")) else {}
    # independent Python path: the stage functions called directly on the same data, into a separate directory
    ind = os.path.join(tmp, "runsh-independent")
    os.makedirs(ind)
    trans_i = os.path.join(ind, "transitions.json")
    C.write_transitions(trans_i, arms=list(x2b.ARMS))
    rp_i = os.path.join(ind, "replay.jsonl")
    V.run_replay(tmp, os.path.join(dr, "maps"), trans_i, rp_i)
    pd_i = os.path.join(ind, "pairs")
    os.makedirs(pd_i)
    subprocess.run([PY, os.path.join(X2, "build_pairs.py"), "--data", dr, "--replay", rp_i, "--out", pd_i], check=True, capture_output=True)
    lr_i = V.run_label(ind, pd_i, V.truth_labels(C.read_jsonl(os.path.join(pd_i, "key.jsonl")), lambda k: "same_concern", C.read_jsonl(os.path.join(pd_i, "control-key.jsonl"))))
    res_i = AX.analyze(dr, rp_i, pd_i)
    r_rep2, r_prs2 = sh_("replay"), sh_("pairs")
    out_txt = os.path.join(dr, "analysis_output.txt")
    check("C run.sh end to end under X2B_DATA on two collector runs (fastapi, axum): `run.sh replay` writes the 12-row codex transitions and the replay, `run.sh pairs` builds the pairs, `run.sh analysis` writes results.json and analysis_output.txt "
          "whose decision, controls and served/flags check equal the independent Python path (same stage functions called directly into another directory); then `run.sh replay` and `run.sh pairs` exit 1 with `refusing` once labels exist",
          rc_cols == [0, 0] and r_rep.returncode == 0 and r_prs.returncode == 0 and lr_s is not None and lr_s.returncode == 0 and lr_i.returncode == 0 and r_an.returncode == 0
          and len(json.load(open(os.path.join(dr, "transitions.json")))) == 12 and {r["arm"] for r in C.read_jsonl(os.path.join(dr, "replay.jsonl"))} == {"codex"}
          and os.path.exists(out_txt) and "== X2b" in open(out_txt).read() and sorted(res_s.get("arms", {})) == ["codex"]
          and json.dumps(res_s["arms"]["codex"]["decision"], sort_keys=True) == json.dumps(res_i["arms"]["codex"]["decision"], sort_keys=True)
          and json.dumps(res_s["controls"], sort_keys=True) == json.dumps(res_i["controls"], sort_keys=True)
          and json.dumps(res_s["x2b"]["served_check"], sort_keys=True) == json.dumps(res_i["x2b"]["served_check"], sort_keys=True)
          and res_s["x2b"]["combination"]["proposal"] == res_i["x2b"]["combination"]["proposal"] and res_s["arms"]["codex"]["decision"]["citable"] is True
          and r_rep2.returncode == 1 and "refusing" in r_rep2.stderr and r_prs2.returncode == 1 and "refusing" in r_prs2.stderr,
          (rc_cols, r_rep.stderr[-300:], r_prs.stderr[-300:], r_an.stderr[-300:], r_rep2.returncode, r_prs2.returncode))


def run_x2_validation():
    r = subprocess.run([PY, os.path.join(X2, "validation", "validate.py")], capture_output=True, text=True)
    n_pass = sum(1 for l in r.stdout.splitlines() if l.startswith("PASS "))
    n_fail = sum(1 for l in r.stdout.splitlines() if l.startswith("FAIL "))
    check("X X2's own validation, run unchanged on the reused code, is green (%d PASS, %d FAIL, rc %s)" % (n_pass, n_fail, r.returncode),
          r.returncode == 0 and n_fail == 0 and n_pass >= 233 and "all validation checks passed" in r.stdout, r.stdout[-800:] + r.stderr[-400:])
    return n_pass


def main():
    skip_x2 = "--skip-x2" in sys.argv
    tmp = tempfile.mkdtemp(prefix="x2b-validate-")
    print("work dir", tmp)
    provenance_checks(tmp)
    scenario_checks(tmp)
    collector_checks(tmp)
    n_x2 = None if skip_x2 else run_x2_validation()
    print("X2b checks: %d (plus X2's %s)" % (COUNT[0], "SKIPPED" if skip_x2 else n_x2))
    if V.FAILS:
        print("FAILED:", *V.FAILS, sep="\n  ")
        sys.exit(1)
    if skip_x2:
        print("X2b-specific checks passed; X2 validation SKIPPED: not a complete validation")
        sys.exit(2)
    print("all validation checks passed")


if __name__ == "__main__":
    main()
