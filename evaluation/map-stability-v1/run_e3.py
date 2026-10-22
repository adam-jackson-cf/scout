#!/usr/bin/env python3
"""E3: Clawpatch map-ID stability. Usage: run_e3.py <repo> <ext> ; repo copy at /tmp/scout-evidence/e3/<repo>."""
import json, os, subprocess, sys, time, glob, shutil, re
repo, ext = sys.argv[1], sys.argv[2]
E3 = "/tmp/scout-evidence/e3"
ROOT = f"{E3}/{repo}"
CP = glob.glob("/tmp/scout-evidence/clawpatch-build/*/dist/cli.js")[0]
OUT = f"/Users/adamjackson/Projects/inflight/scout/evaluation/map-stability-v1/raw/{repo}"
os.makedirs(OUT, exist_ok=True)

def cp(args, state):
    t = time.time()
    p = subprocess.run(["node", CP, *args, "--root", ROOT, "--state-dir", state, "--json"],
                       capture_output=True, text=True)
    return p, time.time() - t

def snap(state):
    feats = []
    for f in sorted(glob.glob(f"{state}/features/*.json")):
        d = json.load(open(f))
        feats.append({k: d.get(k) for k in ("featureId", "title", "source", "kind", "status", "confidence")} |
                     {"ownedFiles": [o["path"] for o in d.get("ownedFiles", [])],
                      "entry": [e.get("symbol") or e.get("path") for e in d.get("entrypoints", [])]})
    return feats

def run(name, source, state, fresh, extra=()):
    if fresh:
        shutil.rmtree(state, ignore_errors=True)
        p, _ = cp(["init"], state)
        assert p.returncode == 0, p.stderr
    p, dt = cp(["map", "--source", source, *extra], state)
    rec = {"name": name, "source": source, "state": state, "fresh": fresh, "rc": p.returncode,
           "seconds": round(dt, 1), "stderr_tail": p.stderr[-3000:]}
    try: rec["result"] = json.loads(p.stdout)
    except Exception: rec["stdout"] = p.stdout[-2000:]
    m = re.findall(r"inventory-done (.*)", p.stderr)
    rec["inventory"] = m[-1] if m else None
    rec["snapshot"] = snap(state) if p.returncode == 0 else None
    json.dump(rec, open(f"{OUT}/{name}.json", "w"), indent=1)
    print(name, rec["rc"], rec["seconds"], rec.get("result"), flush=True)
    return rec

runs = []
for i in (1, 2, 3):
    runs.append(run(f"H{i}", "heuristic", f"{E3}/state_{repo}_H{i}", True))
for i in (1, 2, 3):
    runs.append(run(f"A{i}", "agent", f"{E3}/state_{repo}_A{i}", True))
A1 = f"{E3}/state_{repo}_A1"
for i in (1, 2):
    runs.append(run(f"A1_rerun{i}", "agent", A1, False))

# mutations on repo copy, same state as A1
src = [f for f in runs[3]["snapshot"] if f["source"] == "agent-mapper" and any(o.endswith(ext) and "test" not in o for o in f["ownedFiles"])]
tracked = subprocess.run(["git", "-C", ROOT, "ls-files"], capture_output=True, text=True).stdout.split("\n")
owned_all = sorted({o for f in runs[3]["snapshot"] for o in f["ownedFiles"]})
json.dump({"owned_all": owned_all, "tracked_source": [t for t in tracked if t.endswith(ext)]}, open(f"{OUT}/files.json", "w"))
sys.stdout.flush()
