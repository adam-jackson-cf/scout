#!/usr/bin/env python3
"""X2c measurement check (NOT a study run): one agent map per (arm, repo) on a disposable clone. Records rc, seconds, feature/owned-file counts, probe model.
No area matching, no labels, no Jev. Reuses frozen X2 common.py read-only (env scrubbing, host auth, timeouts).
usage: smoke_run.py <arm-id> <repo>   arms: codex-terra-none, codex-terra-low, claude-haiku-default, claude-haiku-low"""
import glob, json, os, re, shutil, subprocess, sys, tempfile, time
HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", ".."))
sys.path.insert(0, os.path.join(ROOT, "evaluation", "x2-area-stability"))
import common as C

ARMS = {"codex-terra-none": ("codex", "gpt-6.1-terra", "none"), "codex-terra-low": ("codex", "gpt-6.1-terra", "low"),
        "claude-haiku-default": ("claude", "claude-haiku-5-5", None), "claude-haiku-low": ("claude", "claude-haiku-5-5", "low")}
SRC = "/tmp/scout-x2c-smoke/_src"
TIMEOUT = 3300
arm, repo = sys.argv[1], sys.argv[2]
prov, model, eff = ARMS[arm]
cli = os.environ["SCOUT_CLAWPATCH_BINARY"]
env = C.clawpatch_map_env(os.environ)
base = "/tmp/scout-x2c-smoke/%s/%s" % (arm, repo)
ws, state = base + "/ws", base + "/state"
os.makedirs(base)
subprocess.run(["git", "clone", "-q", "--no-hardlinks", SRC + "/" + repo, ws], check=True)


def run(cmd, env, inp=None, cwd=None, timeout=TIMEOUT):
    t0 = time.time()
    try:
        r = subprocess.run(cmd, env=env, input=inp, capture_output=True, text=True, timeout=timeout, cwd=cwd)
        return r.returncode, r.stdout, r.stderr, round(time.time() - t0, 1)
    except subprocess.TimeoutExpired as e:
        return 124, "", "timeout", round(time.time() - t0, 1)


def probe():
    with tempfile.TemporaryDirectory() as d:
        if prov == "claude":
            cmd = [C.claude_executable(env), "-p", "--output-format", "json", "--model", model, "--tools", "", "--no-session-persistence", "--safe-mode", "--strict-mcp-config",
                   "--mcp-config", json.dumps({"mcpServers": {}}), "--disable-slash-commands", "--no-chrome"] + (["--effort", eff] if eff else [])
            penv = C.claude_provider_env(env, d)
        else:
            cmd = ["codex", "exec", "--cd", d, "--sandbox", "read-only", "--skip-git-repo-check", "--model", model, "-c", 'model_reasoning_effort="%s"' % eff, "-"]
            penv = env
        rc, out, err, s = run(cmd, penv, inp="Reply with the single word: ok", cwd=d, timeout=300)
    res = {"rc": rc, "seconds": s}
    if prov == "claude":
        try:
            j = json.loads(out)
            res["modelUsage"] = j.get("modelUsage"); res["total_cost_usd"] = j.get("total_cost_usd"); res["usage"] = j.get("usage")
        except ValueError:
            res["raw"] = out[-500:]
    else:
        txt = out + err
        for k in ("model", "reasoning effort", "provider"):
            m = re.search(r"^%s:\s*(.+)$" % k, txt, re.M)
            res[k] = m.group(1).strip() if m else None
        m = re.search(r"tokens used\s*\n?\s*([\d,]+)", txt)
        res["tokens_used"] = m.group(1) if m else None
    return res


pre = probe()
flags = ["--model", model] + (["--reasoning-effort", eff] if eff else [])
g = ["node", cli, "--root", ws, "--state-dir", state, "--json"]
rc0, o0, e0, s0 = run(g + ["init"], env)
t0 = time.time()
rc, out, err, s = run(g + ["map", "--source", "agent", "--provider", prov, *flags], env) if rc0 == 0 else (rc0, o0, e0, 0)
feats = [json.load(open(p)) for p in sorted(glob.glob(state + "/features/*.json"))]
act = [f for f in feats if f.get("status") != "stale"]
by_src = {}
for f in act:
    by_src[f.get("source")] = by_src.get(f.get("source"), 0) + 1
owned = lambda fs: {o["path"] if isinstance(o, dict) else o for f in fs for o in f.get("ownedFiles", [])}
agent = [f for f in act if f.get("source") == "agent"]
dirty = bool(subprocess.run(["git", "status", "--porcelain"], cwd=ws, capture_output=True, text=True).stdout.strip())
post = probe()
rec = {"arm": arm, "repo": repo, "provider": prov, "model_flag": model, "effort_flag": eff, "init_rc": rc0, "rc": rc, "seconds": s,
       "features_total": len(feats), "features_active": len(act), "features_by_source": by_src, "features_agent": len(agent),
       "owned_files_agent": len(owned(agent)), "owned_files_all_active": len(owned(act)), "workspace_dirty": dirty,
       "probe_before": pre, "probe_after": post, "stdout_tail": out[-1500:], "stderr_tail": err[-2500:]}
os.makedirs(HERE + "/results", exist_ok=True)
json.dump(rec, open("%s/results/%s.%s.json" % (HERE, arm, repo), "w"), indent=1)
print(arm, repo, "rc", rc, s, "s agent", len(agent), "owned", len(owned(agent)), by_src, flush=True)
