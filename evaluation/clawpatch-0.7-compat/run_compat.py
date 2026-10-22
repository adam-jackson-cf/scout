#!/usr/bin/env python3
"""Reduced E3 on Clawpatch 0.7.3 (+ 0.4.0 heuristic baseline).
Usage: run_compat.py <repo> [--no-agent]
Fresh copies of /tmp/scout-evidence/repos/<repo> go to /tmp/scout-evidence/cp07-work/<repo> (originals untouched)."""
import json, os, subprocess, sys, time, shutil, re, glob

repo = sys.argv[1]
agent = "--no-agent" not in sys.argv
W = "/tmp/scout-evidence/cp07-work"
ROOT = f"{W}/{repo}"
CP04 = glob.glob("/tmp/scout-evidence/clawpatch-build/*/dist/cli.js")[0]
CP07 = "/tmp/scout-evidence/cp073-safe/dist/cli.js"
OUT = f"/Users/adamjackson/Projects/inflight/scout/evaluation/clawpatch-0.7-compat/raw/{repo}"
os.makedirs(OUT, exist_ok=True)
if not os.path.isdir(ROOT):
    os.makedirs(W, exist_ok=True)
    shutil.copytree(f"/tmp/scout-evidence/repos/{repo}", ROOT, symlinks=True)

def cp(cli, args, state):
    t = time.time()
    p = subprocess.run(["node", cli, *args, "--root", ROOT, "--state-dir", state, "--json"],
                       capture_output=True, text=True)
    return p, time.time() - t

def snap(state):
    feats = []
    for f in sorted(glob.glob(f"{state}/features/*.json")):
        d = json.load(open(f))
        feats.append({k: d.get(k) for k in ("featureId", "title", "source", "kind", "status", "confidence")} |
                     {"ownedFiles": [o["path"] for o in d.get("ownedFiles", [])]})
    return feats

def run(name, cli, source, state, fresh):
    if fresh:
        shutil.rmtree(state, ignore_errors=True)
        p, _ = cp(cli, ["init"], state)
        assert p.returncode == 0, p.stderr
    p, dt = cp(cli, ["map", "--source", source], state)
    rec = {"name": name, "cli": cli, "source": source, "rc": p.returncode, "seconds": round(dt, 1),
           "stderr_tail": p.stderr[-3000:]}
    try: rec["result"] = json.loads(p.stdout)
    except Exception: rec["stdout"] = p.stdout[-2000:]
    m = re.findall(r"inventory-done (.*)", p.stderr)
    rec["inventory"] = m[-1] if m else None
    rec["snapshot"] = snap(state) if p.returncode == 0 else None
    json.dump(rec, open(f"{OUT}/{name}.json", "w"), indent=1)
    print(name, rec["rc"], rec["seconds"], rec.get("result"), flush=True)

if "--agent-only" not in sys.argv: run("H04_1", CP04, "heuristic", f"{W}/s_{repo}_H04_1", True)
if "--agent-only" not in sys.argv: run("H07_1", CP07, "heuristic", f"{W}/s_{repo}_H07_1", True)
if "--agent-only" not in sys.argv: run("H07_2", CP07, "heuristic", f"{W}/s_{repo}_H07_2", True)
if agent:
    import concurrent.futures as cf
    with cf.ThreadPoolExecutor(2) as ex:
        list(ex.map(lambda n: run(f"A07_{n}", CP07, "agent", f"{W}/s_{repo}_A07_{n}", True), (1, 2)))
