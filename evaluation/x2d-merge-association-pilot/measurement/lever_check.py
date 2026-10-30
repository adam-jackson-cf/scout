"""X2d measurement check (shape-decision section 0.a step 1; PREREGISTRATION section 12): is a per-state-dir `config.json` `include` scope honoured by the
Clawpatch AGENT-map inventory, and does `assertTrustedCodexConfig` accept it for the `state-dir` config source?

Exactly ONE real agent map (codex gpt-6-luna, xhigh) into a throwaway state directory under /tmp/x2d-lever/. Not part of the signed-off pipeline; the result is recorded in
DRAFT-NOTES.md. Invocation mirrors X2c's collect_x2c.py (`--root WS --state-dir STATE --json init`, then `map --source agent --provider codex --model M
--reasoning-effort E`, Scout's map environment from x2-area-stability/common.py). Between `init` and `map` it rewrites `<state>/config.json` `include` (what the
extension would do). Standard library only; run with .tools/venv/bin/python after `source .tools/activate`.

    .tools/venv/bin/python evaluation/x2d-merge-association-pilot/measurement/lever_check.py --source /tmp/scout-x2c-smoke/_src/fastapi --rev 1d50bad4555b951d5fa48eb8787fd5c39ed7effb
"""
import argparse, json, os, shutil, subprocess, sys, time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "..", "x2-area-stability"))
import common as C  # frozen X2 file; read-only import

ROOT = "/tmp/x2d-lever"


def run(cmd, env, cwd=None, timeout=1800):
    t0 = time.time()
    p = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True, text=True, timeout=timeout)
    return p.returncode, p.stdout, p.stderr, round(time.time() - t0, 1)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--source", required=True)
    ap.add_argument("--rev", required=True)
    ap.add_argument("--include", default="fastapi/**", help="comma-separated include globs written to the state-dir config.json")
    ap.add_argument("--model", default="gpt-6-luna")
    ap.add_argument("--effort", default="xhigh")
    ap.add_argument("--cli", default=os.environ.get("SCOUT_CLAWPATCH_BINARY"))
    a = ap.parse_args()
    if os.path.exists(ROOT):
        sys.exit("refusing: %s exists (the check is one map; remove it deliberately to repeat)" % ROOT)
    bad = [n for n in os.environ if n.startswith("CLAWPATCH_")]
    if bad:
        sys.exit("refusing: Clawpatch env names set: %s" % ", ".join(bad))
    os.makedirs(ROOT)
    ws, state = os.path.join(ROOT, "ws"), os.path.join(ROOT, "state")
    subprocess.run(["git", "clone", "-q", "--shared", "--no-checkout", a.source, ws], check=True)
    subprocess.run(["git", "-C", ws, "checkout", "-q", "--detach", a.rev], check=True)
    env = C.clawpatch_map_env(dict(os.environ))
    node = shutil.which("node")
    base = [node, a.cli, "--root", ws, "--state-dir", state, "--json"]
    rec = {"rev": a.rev, "include": a.include.split(","), "model": a.model, "effort": a.effort, "state_dir": state}
    rc, out, err, sec = run(base + ["init"], env)
    rec["init"] = {"rc": rc, "seconds": sec, "stderr_tail": err[-400:]}
    if rc != 0:
        print(json.dumps(rec, indent=1)); sys.exit(1)
    cfg_path = os.path.join(state, "config.json")
    cfg = json.load(open(cfg_path))
    rec["config_before"] = {"include": cfg["include"], "exclude": cfg["exclude"]}
    cfg["include"] = a.include.split(",")
    json.dump(cfg, open(cfg_path, "w"), indent=2)
    cmd = base + ["map", "--source", "agent", "--provider", "codex", "--model", a.model, "--reasoning-effort", a.effort]
    rec["map_argv_tail"] = cmd[len(base):]
    rc, out, err, sec = run(cmd, env)
    rec["map"] = {"rc": rc, "seconds": sec, "stderr_events": [l for l in err.splitlines() if l.startswith("clawpatch map")][:60], "stderr_tail": err[-600:]}
    try:
        rec["map"]["result_head"] = json.loads(out)
        rec["map"]["result_head"] = {k: v for k, v in rec["map"]["result_head"].items() if k != "features"} if isinstance(rec["map"]["result_head"], dict) else None
    except ValueError:
        rec["map"]["stdout_head"] = out[:600]
    feats = []
    fdir = os.path.join(state, "features")
    if os.path.isdir(fdir):
        for n in sorted(os.listdir(fdir)):
            f = json.load(open(os.path.join(fdir, n)))
            feats.append({"id": f.get("featureId"), "source": f.get("source"), "title": f.get("title"), "files": [r["path"] for r in f.get("ownedFiles", [])]})
    rec["features"] = feats
    agent = [f for f in feats if f["source"] == "agent-mapper"]  # internal/areas.AgentSource
    pref = tuple(i[:-3] + "/" for i in a.include.split(",") if i.endswith("/**"))
    rec["agent_feature_files_outside_include"] = sorted({p for f in agent for p in f["files"] if not p.startswith(pref)})
    json.dump(rec, open(os.path.join(ROOT, "lever-result.json"), "w"), indent=1)
    print(json.dumps({k: rec[k] for k in ("include", "init", "map") if k in rec}, indent=1)[:3000])
    print("agent features:", len(agent), "| files outside include:", len(rec["agent_feature_files_outside_include"]))


if __name__ == "__main__":
    main()
