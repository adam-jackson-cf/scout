"""Check revision-9 frozen bytes, including its recorded adoption exceptions."""
import hashlib
import json
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parents[4]
BASELINE = "e2b94e3c"
MANIFEST = "evaluation/x1-triage-ranking/reviews/rev9-digests.json"
MANIFEST_SHA256 = "2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da"


def sha(data):
    return hashlib.sha256(data).hexdigest()


def check():
    raw = (ROOT/MANIFEST).read_bytes()
    if sha(raw) != MANIFEST_SHA256:
        raise RuntimeError("revision-9 manifest changed")
    manifest = json.loads(raw)
    literal = 0
    unchanged = 0
    shared = set()
    for path, expected in manifest["files_sha256"].items():
        data = (ROOT/path).read_bytes()
        literal += sha(data) == expected
        baseline = subprocess.check_output(["git", "show", BASELINE+":"+path], cwd=ROOT)
        if data != baseline:
            raise RuntimeError("frozen baseline byte changed: "+path)
        unchanged += 1
        if sha(data) != expected:
            shared.add(path)
    groups = set()
    exact_extracts = 0
    normalized_extracts = []
    for name, entry in manifest["owned_extracts_sha256"].items():
        text = (ROOT/entry["source"]).read_text()
        if name.startswith("R-") or name in ("HD9", "ADR014Index"):
            marker = "ADR-014" if name == "ADR014Index" else name
            matches = [line for line in text.splitlines(keepends=True) if line.startswith("| "+marker+" |")]
        elif name == "ADR014Card":
            matches = [part for part in re.findall(r"  <article[^>]*>[\s\S]*?  </article>\n", text) if ">ADR-014<" in part]
        else:
            prefix = entry["selector"].split("beginning ", 1)[1]
            matches = [line for line in text.splitlines(keepends=True) if line.startswith(prefix)]
        if len(matches) != 1:
            raise RuntimeError("ambiguous owned extract: "+name)
        value = matches[0]
        if sha(value.encode()) == entry["sha256"]:
            exact_extracts += 1
        elif name == "HD9" and sha(value.replace("| preregistered |", "| proposed |").encode()) == entry["sha256"]:
            normalized_extracts.append(name)
        else:
            raise RuntimeError("owned extract changed: "+name)
        groups.add(entry["source"])
    if shared != groups or len(groups) != 5 or literal != 126 or unchanged != 131:
        raise RuntimeError("unexpected baseline exception membership")
    return {"baseline": BASELINE, "manifest_sha256": sha(raw), "literal_file_matches": literal,
            "frozen_file_count": unchanged, "unchanged_against_baseline": unchanged,
            "owned_extract_file_groups": len(groups), "literal_owned_extract_matches": exact_extracts,
            "permitted_round9_condition2_state_normalization": normalized_extracts,
            "shared_files_not_modified": sorted(shared)}


if __name__ == "__main__":
    print(json.dumps(check(), indent=2, sort_keys=True))
