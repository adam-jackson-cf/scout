#!/usr/bin/env python3
"""E1b-v2 stage 5: Jev labelling of candidate follow-up pairs ("does the later change correct/complete/revert the earlier one?").

Usage: the run REQUIRES OPENROUTER_API_KEY in the environment (presence-checked, never printed or stored). Supply it by exporting it
in your shell; no secret store or project is assumed. Do not run before council sign-off (labels already collected in design review
are the frozen outcome construction; any re-run needs a new review round).
  uv run --with pandas python <abs>/jev_label.py <abs>/candidates.jsonl.gz <abs>/labels.jsonl [model] [reps]
  uv run python <abs>/jev_label.py --digest        # prints the sha256 of the question definition; no network, no key
Writes labels.jsonl: one record per call {repo,parent,follow,rep,ok,choice,p_corrective,confidence,model_resolved,qhash};
finalize_panel.py averages p_corrective over reps. Resumable (skips done calls). Records written before the qhash field existed
were produced with the identical question text (digest recorded in prompts.json).
The state shows titles, body excerpts and changed-file paths of BOTH PRs only: no predictor value (hidden-coupling, Q, D),
no repository-level statistics, no task type.
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
QUESTION = {"q": {"type": "choice",
    "instructions": ("Two pull requests to the same repository are shown: `earlier` was merged first, `later` was merged within "
                     "14 days after it. Decide from the titles, descriptions and changed files whether `later` exists because "
                     "`earlier` was wrong or incomplete."),
    "criteria": {
        "corrective": ("`later` fixes a defect or regression introduced by `earlier`, completes something `earlier` omitted "
                       "(a missed call site, config, migration, doc/code update the first change should have contained), or "
                       "reverts `earlier`."),
        "independent": ("`later` is ordinary continued development, an unrelated fix or feature, a refactor, a dependency or "
                        "release chore, or merely touches the same files for a different purpose."),
        "cannot_tell": "The descriptions do not give enough to decide."}}}


def state(parent, follow, cand, raw):
    def view(p):
        return dict(title=p["title"], description=p["body"][:700], files=[f["path"] for f in p["files"]][:15])
    return dict(earlier=view(parent), later=view(follow), hours_between_merges=round(cand["gap_h"], 1))


def call(job, key, model):
    body = json.dumps({"model": model, "state": job["state"], "questions": QUESTION}).encode()
    rec = dict(repo=job["repo"], parent=job["parent"], follow=job["follow"], rep=job["rep"], qhash=qhash())
    for attempt in range(3):
        try:
            r = urllib.request.Request(URL, body, {"Authorization": "Bearer " + key, "Content-Type": "application/json"})
            d = json.load(urllib.request.urlopen(r, timeout=60))
            a = next(iter(d["answers"].values()))
            rec.update(ok=True, model_resolved=d.get("model"), choice=a["choice"],
                       p_corrective=a["probabilities"].get("corrective", 0.0), confidence=a.get("confidence"))
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
    cand_f, out_f = sys.argv[1], Path(sys.argv[2])
    model = sys.argv[3] if len(sys.argv) > 3 else DEFAULT_MODEL
    reps = int(sys.argv[4]) if len(sys.argv) > 4 else 3
    key = os.environ.get("OPENROUTER_API_KEY")
    if not key:
        sys.exit("OPENROUTER_API_KEY missing: export it in the environment (required by the preregistration)")
    cands = [json.loads(l) for l in gzip.open(cand_f, "rt")]
    cands = [c for c in cands if not c["censored"]]
    raws = {}
    done = set()
    if out_f.exists():
        for l in out_f.read_text().splitlines():
            r = json.loads(l)
            if r.get("ok"):
                done.add((r["repo"], r["parent"], r["follow"], r["rep"]))
    jobs = []
    for c in cands:
        if c["repo"] not in raws:
            raws[c["repo"]] = {p["number"]: p for p in json.loads(gzip.open(HERE / "raw" / f"prs-{c['repo']}.json.gz", "rt").read())}
        st = state(raws[c["repo"]][c["parent"]], raws[c["repo"]][c["follow"]], c, raws)
        for rep in range(reps):
            if (c["repo"], c["parent"], c["follow"], rep) not in done:
                jobs.append(dict(repo=c["repo"], parent=c["parent"], follow=c["follow"], rep=rep, state=st))
    with ThreadPoolExecutor(8) as ex, out_f.open("a") as fh:
        for rec in ex.map(lambda j: call(j, key, model), jobs):
            fh.write(json.dumps(rec) + "\n")
            fh.flush()
    print(len(jobs), "calls")


if __name__ == "__main__":
    main()
