#!/usr/bin/env python3
"""E1b-v2 stage 0/1: pin each repository and fetch merged PRs (GitHub GraphQL) over (T_end - 730 d, T_end].

Usage: python3 fetch_prs.py [repo ...]     (needs authenticated `gh`; responses cached under raw/q/)
Output: pins.json (written for all repos when no args), raw/prs-<repo>.json.gz = list of PR dicts.
Fields kept: number, title, body (<=2000 chars), labels, created, merged, author login/type/association,
commit list (oid, authoredDate, committedDate, author email, first parent), file list (path, additions,
deletions, changeType), reviews (state, submittedAt, author), merge commit, base/head oids.
"""
import datetime as dt
import gzip
import hashlib
import json
import subprocess
import sys
import time

from common import DAY, HERE, NAMES, REPO_CFG, WINDOW_DAYS, pin

RAW = HERE / "raw"
PAGE = 20
Q_COUNT = "query($q:String!){search(query:$q,type:ISSUE,first:1){issueCount}}"
Q_PAGE = """query($q:String!,$after:String){search(query:$q,type:ISSUE,first:%d,after:$after){
 issueCount pageInfo{hasNextPage endCursor}
 nodes{... on PullRequest{number title bodyText createdAt mergedAt authorAssociation
   additions deletions changedFiles baseRefOid headRefOid
   author{__typename login} mergeCommit{oid parents(first:3){nodes{oid}}}
   labels(first:20){nodes{name}}
   commits(first:100){totalCount nodes{commit{oid authoredDate committedDate author{email}
      parents(first:1){nodes{oid}}}}}
   files(first:100){totalCount nodes{path additions deletions changeType}}
   reviews(first:100){totalCount nodes{state submittedAt author{login}}}}}}
 rateLimit{remaining resetAt cost}}""" % PAGE


def gql(query, **vars):
    key = hashlib.sha1(json.dumps([query, vars], sort_keys=True).encode()).hexdigest()
    f = RAW / "q" / (key + ".json.gz")
    if f.exists():
        return json.loads(gzip.open(f, "rt").read())
    args = ["gh", "api", "graphql", "-f", "query=" + query]
    for k, v in vars.items():
        if v is not None:
            args += ["-f", f"{k}={v}"]
    err = ""
    for attempt in range(30):
        r = subprocess.run(args, capture_output=True, text=True)
        if r.returncode == 0:
            d = json.loads(r.stdout)
            if "errors" in d and not (d.get("data") or {}).get("search"):
                err = str(d["errors"])[:300]
            else:
                f.parent.mkdir(parents=True, exist_ok=True)
                f.write_bytes(gzip.compress(r.stdout.encode()))
                rl = (d.get("data") or {}).get("rateLimit")
                if rl and rl["remaining"] < 200:
                    time.sleep(90)
                return d
        else:
            err = r.stderr[:300]
        time.sleep(120 if "rate limit" in err.lower() else 5 * (attempt + 1))
    raise RuntimeError("gh failed: " + err)


def day(ts):
    return dt.datetime.fromtimestamp(ts, dt.timezone.utc).date()


def fetch_range(slug, a, b, out):
    q = f"repo:{slug} is:pr is:merged merged:{a}..{b}"
    n = gql(Q_COUNT, q=q)["data"]["search"]["issueCount"]
    if n > 1000 and a < b:
        mid = a + (b - a) // 2
        fetch_range(slug, a, mid, out)
        fetch_range(slug, mid + dt.timedelta(days=1), b, out)
        return
    after = None
    while True:
        d = gql(Q_PAGE, q=q, after=after)["data"]["search"]
        out.extend(x for x in d["nodes"] if x)
        if not d["pageInfo"]["hasNextPage"]:
            break
        after = d["pageInfo"]["endCursor"]


def slim(p):
    mc = p.get("mergeCommit") or {}
    return dict(
        number=p["number"], title=p["title"], body=(p.get("bodyText") or "")[:2000],
        labels=[n["name"] for n in p["labels"]["nodes"]],
        created=p["createdAt"], merged=p["mergedAt"], assoc=p["authorAssociation"],
        additions=p["additions"], deletions=p["deletions"], changed_files=p["changedFiles"],
        base_oid=p.get("baseRefOid"), head_oid=p.get("headRefOid"),
        author=(p.get("author") or {}).get("login"), author_type=(p.get("author") or {}).get("__typename"),
        merge_commit=mc.get("oid"), merge_parents=[n["oid"] for n in (mc.get("parents") or {}).get("nodes", [])],
        n_commits=p["commits"]["totalCount"],
        commits=[dict(oid=n["commit"]["oid"], authored=n["commit"]["authoredDate"],
                      committed=n["commit"]["committedDate"],
                      email=((n["commit"].get("author") or {}).get("email")),
                      parent=(((n["commit"].get("parents") or {}).get("nodes") or [{}])[0] or {}).get("oid"))
                 for n in p["commits"]["nodes"]],
        n_files=p["files"]["totalCount"],
        files=[dict(path=n["path"], add=n["additions"], dele=n["deletions"], ct=n["changeType"])
               for n in p["files"]["nodes"]],
        n_reviews=p["reviews"]["totalCount"],
        reviews=[dict(state=n["state"], at=n["submittedAt"], by=(n.get("author") or {}).get("login"))
                 for n in p["reviews"]["nodes"]])


def main():
    RAW.mkdir(exist_ok=True)
    names = sys.argv[1:] or NAMES
    pf = HERE / "pins.json"
    for repo in names:
        pins = json.loads(pf.read_text()) if pf.exists() else {}
        if repo in pins:   # pins.json is an INPUT once written (frozen, digest in DESIGN-NOTES); never recomputed
            ts = pins[repo]["t_end_ts"]
        else:
            sha, ts = pin(repo)
            pins[repo] = dict(slug=REPO_CFG[repo][0], sha=sha, t_end_ts=ts,
                              t_end=dt.datetime.fromtimestamp(ts, dt.timezone.utc).isoformat(),
                              window_start_ts=ts - WINDOW_DAYS * DAY)
            pf.write_text(json.dumps(pins, indent=1, sort_keys=True))
        out, cur = [], day(ts - WINDOW_DAYS * DAY)
        end = day(ts)
        while cur <= end:  # 90-day chunks, split further when > 1000 hits
            nxt = min(end, cur + dt.timedelta(days=90))
            fetch_range(REPO_CFG[repo][0], cur, nxt, out)
            cur = nxt + dt.timedelta(days=1)
        seen = {p["number"]: p for p in out}
        prs = [slim(p) for p in seen.values()]
        with gzip.open(RAW / f"prs-{repo}.json.gz", "wt") as fh:
            json.dump(prs, fh)
        print(repo, len(prs), flush=True)


if __name__ == "__main__":
    main()
