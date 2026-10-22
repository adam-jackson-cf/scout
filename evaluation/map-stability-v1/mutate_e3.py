#!/usr/bin/env python3
"""E3 phase 2: content-only and structural change, remap --source agent against A1 state.
Usage: mutate_e3.py <repo> <ext>   (<ext> is .py or .rs)"""
import json, os, subprocess, sys, time, glob, re, py_compile, shutil
repo, ext = sys.argv[1], sys.argv[2]
# reuse helpers (cp/snap/run) from run_e3.py: everything before the `runs = []` driver section
_src = open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "run_e3.py")).read().split("runs = []")[0]
exec(_src)
A1 = f"{E3}/state_{repo}_A1"
base = json.load(open(f"{OUT}/A1_rerun2.json"))["snapshot"]


def pick():
    for f in base:
        if f["source"] != "agent-mapper":
            continue
        for o in f["ownedFiles"]:
            if not o.endswith(ext) or "test" in o or "docs_src" in o:
                continue
            lines = open(f"{ROOT}/{o}").read().split("\n")
            for i, l in enumerate(lines[:-1]):
                if ext == ".py" and re.match(r"\s*def \w+\(.*\)( -> .*)?:\s*$", l) and not lines[i + 1].strip().startswith(('"""', "'''")):
                    return o, i, l
                if ext == ".rs" and re.match(r"\s*(pub(\(crate\))? )?(async )?fn \w+.*\{\s*$", l) and "where" not in l:
                    return o, i, l
    raise SystemExit("no candidate")


o, i, l = pick()
lines = open(f"{ROOT}/{o}").read().split("\n")
indent = re.match(r"\s*", l).group(0) + "    "
lines.insert(i + 1, indent + ("_e3_marker = 0  # e3 content-only change" if ext == ".py" else "let _e3_marker = 0; // e3 content-only change"))
open(f"{ROOT}/{o}", "w").write("\n".join(lines))
if ext == ".py":
    py_compile.compile(f"{ROOT}/{o}", doraise=True)
owner = [f["featureId"] for f in base if o in f["ownedFiles"]]
json.dump({"file": o, "line": i + 1, "sig": l.strip(), "owners_before": owner}, open(f"{OUT}/content_change.json", "w"), indent=1)
run("M1_content", "agent", A1, False)

newf = {".py": "fastapi/e3_new_module.py", ".rs": "axum/src/e3_new_module.rs"}[ext]
os.makedirs(os.path.dirname(f"{ROOT}/{newf}"), exist_ok=True)
body = {".py": "def e3_added(x: int) -> int:\n    return x * 2\n", ".rs": "pub fn e3_added(x: u32) -> u32 {\n    x * 2\n}\n"}[ext]
open(f"{ROOT}/{newf}", "w").write(body)
json.dump({"newfile": newf}, open(f"{OUT}/struct_change.json", "w"))
run("M2_struct", "agent", A1, False)
