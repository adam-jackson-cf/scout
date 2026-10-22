#!/usr/bin/env python3
"""E1b-v2 stage 2b: Jev task-type labelling of every PR that survived the collect.py filters (the task-type control).

Usage: the run REQUIRES OPENROUTER_API_KEY in the environment (presence-checked, never printed or stored); export it in your shell, no
secret store or project is assumed. Do not run before council sign-off (types collected in design review are the frozen construction).
  uv run --with pandas python <abs>/jev_type.py <abs>/types.jsonl [model] [reps]
  uv run python <abs>/jev_type.py --digest        # sha256 of the question definition; no network, no key
Writes types.jsonl: one record per call {repo,number,rep,ok,choice,probs{class: p},model_resolved,qhash}; build_panel.py takes the argmax of the
rep-mean probabilities (ties -> other_dev). The state shows title, labels, description and changed-file paths ONLY (no predictor,
no outcome). Resumable. Validated by the annotator gate G4 (validate.py).
"""
import gzip
import hashlib
import json
import os
import sys
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

URL = "https://openrouter.ai/api/alpha/decisions"
DEFAULT_MODEL = "typesafe/jev-1.14-20261014"
HERE = Path(__file__).resolve().parent
CLASSES = ["fix", "feature", "refactor", "other_dev", "nondev"]
QUESTION = {"q": {"type": "choice",
    "instructions": "Classify this pull request by the kind of work it is, from its title, labels, description and changed files.",
    "criteria": {
        "fix": "Repairs incorrect behaviour: a bug, crash, panic, regression, security flaw, wrong result or missing error handling.",
        "feature": "Adds a capability, API, option, command or support that did not exist before.",
        "refactor": ("Restructures, renames, cleans up, simplifies, deprecates, removes dead code or tunes performance without "
                     "intending to change observable behaviour."),
        "other_dev": ("Other source-code development work that is not clearly one of the above: intentional behaviour or API changes, "
                      "compatibility work, configuration changes, mixed changes."),
        "nondev": ("Not development effort on the product code: dependency or version bump, release or changelog, documentation-only "
                   "or comment-only, tests-only, CI or build configuration, backport or cherry-pick of another pull request, "
                   "automated or generated change.")}}}


def state(p):
    return dict(title=p["title"], labels=p["labels"], description=p["body"][:700], files=[f["path"] for f in p["files"]][:15])


def call(job, key, model):
    body = json.dumps({"model": model, "state": job["state"], "questions": QUESTION}).encode()
    rec = dict(repo=job["repo"], number=job["number"], rep=job["rep"], qhash=qhash())
    for _ in range(3):
        try:
            r = urllib.request.Request(URL, body, {"Authorization": "Bearer " + key, "Content-Type": "application/json"})
            d = json.load(urllib.request.urlopen(r, timeout=60))
            a = next(iter(d["answers"].values()))
            rec.update(ok=True, model_resolved=d.get("model"), choice=a["choice"],
                       probs={c: a["probabilities"].get(c, 0.0) for c in CLASSES})
            return rec
        except urllib.error.HTTPError as e:
            rec.update(ok=False, http=e.code)
            if e.code in (401, 402):
                return rec
        except Exception as e:
            rec.update(ok=False, error=type(e).__name__)
    return rec


def qhash():
    return hashlib.sha256(json.dumps(QUESTION, sort_keys=True).encode()).hexdigest()


def main():
    if sys.argv[1:2] == ["--digest"]:
        print(qhash())
        return
    out_f = Path(sys.argv[1])
    model = sys.argv[2] if len(sys.argv) > 2 else DEFAULT_MODEL
    reps = int(sys.argv[3]) if len(sys.argv) > 3 else 1
    key = os.environ.get("OPENROUTER_API_KEY")
    if not key:
        sys.exit("OPENROUTER_API_KEY missing: export it in the environment (required by the preregistration)")
    done = set()
    if out_f.exists():
        for l in out_f.read_text().splitlines():
            r = json.loads(l)
            if r.get("ok"):
                done.add((r["repo"], r["number"], r["rep"]))
    jobs = []
    for f in sorted((HERE / "data").glob("pr-*.jsonl.gz")):
        repo = f.name[3:-9]
        rows = [json.loads(l) for l in gzip.open(f, "rt")]
        raw = {p["number"]: p for p in json.loads(gzip.open(HERE / "raw" / f"prs-{repo}.json.gz", "rt").read())}
        for r in rows:
            if "skip" in r:
                continue
            for rep in range(reps):
                if (repo, r["number"], rep) not in done:
                    jobs.append(dict(repo=repo, number=r["number"], rep=rep, state=state(raw[r["number"]])))
    with ThreadPoolExecutor(8) as ex, out_f.open("a") as fh:
        for rec in ex.map(lambda j: call(j, key, model), jobs):
            fh.write(json.dumps(rec) + "\n")
            fh.flush()
    print(len(jobs), "calls")


if __name__ == "__main__":
    main()
