#!/usr/bin/env python3
"""E1b: fetch merged PRs (GitHub GraphQL search) over [T-1, T_end] for each repo into raw/.

Usage: python3 fetch_prs.py [repo ...]     (needs authenticated `gh`; cached per query page under raw/)
Output: raw/prs-<repo>.json = list of PR dicts. Windows are day-granular merged: ranges, split until <= 1000 hits.
"""
import datetime as dt
import hashlib
import json
import subprocess
import sys
import time
from pathlib import Path

HERE = Path(__file__).resolve().parent
RAW = HERE / "raw"
REPOS = Path("/tmp/scout-evidence/repos")
NAMES = ["fastapi", "webpack", "nest", "caddy", "axum", "junit5", "eShopOnWeb"]
SLUG = {"fastapi": "fastapi/fastapi", "webpack": "webpack/webpack", "nest": "nestjs/nest", "caddy": "caddyserver/caddy",
        "axum": "tokio-rs/axum", "junit5": "junit-team/junit-framework", "eShopOnWeb": "dotnet-architecture/eShopOnWeb"}
DAY = 86400
Q_COUNT = "query($q:String!){search(query:$q,type:ISSUE,first:1){issueCount}}"
Q_PAGE = """query($q:String!,$after:String){search(query:$q,type:ISSUE,first:50,after:$after){
 issueCount pageInfo{hasNextPage endCursor}
 nodes{... on PullRequest{number createdAt mergedAt author{login} mergeCommit{oid}
   commits(first:100){totalCount nodes{commit{oid}}}
   reviews(first:100){totalCount nodes{comments{totalCount}}}}}}
 rateLimit{remaining resetAt cost}}"""


def gql(query, **vars):
    key = hashlib.sha1(json.dumps([query, vars], sort_keys=True).encode()).hexdigest()
    f = RAW / "q" / (key + ".json")
    if f.exists():
        return json.loads(f.read_text())
    args = ["gh", "api", "graphql", "-f", "query=" + query]
    for k, v in vars.items():
        if v is not None:
            args += ["-f", f"{k}={v}"]
    for attempt in range(6):
        r = subprocess.run(args, capture_output=True, text=True)
        if r.returncode == 0:
            d = json.loads(r.stdout)
            if "errors" in d and "data" not in d:
                raise RuntimeError(str(d["errors"])[:300])
            f.parent.mkdir(parents=True, exist_ok=True)
            f.write_text(json.dumps(d))
            rl = (d.get("data") or {}).get("rateLimit")
            if rl and rl["remaining"] < 100:
                time.sleep(60)
            return d
        time.sleep(5 * (attempt + 1))
    raise RuntimeError("gh failed: " + r.stderr[:300])


def window(repo):
    def g(*a):
        return subprocess.run(["git", "-C", str(REPOS / repo), *a], capture_output=True, text=True).stdout.strip()
    tend = g("rev-parse", "HEAD")
    tend_ts = int(g("log", "-1", "--format=%ct", tend))
    t0 = g("rev-list", "--first-parent", "-1", f"--before={tend_ts - 365 * DAY}", tend)
    t0_ts = int(g("log", "-1", "--format=%ct", t0))
    t1 = g("rev-list", "--first-parent", "-1", f"--before={t0_ts - 365 * DAY}", t0)
    if not t1:
        return None
    t1_ts = int(g("log", "-1", "--format=%ct", t1))
    return t1_ts, tend_ts


def day(ts):
    return dt.datetime.fromtimestamp(ts, dt.timezone.utc).date()


def fetch_range(repo, a, b, out):
    q = f"repo:{SLUG[repo]} is:pr is:merged merged:{a}..{b}"
    n = gql(Q_COUNT, q=q)["data"]["search"]["issueCount"]
    if n > 1000 and a < b:
        mid = a + (b - a) // 2
        fetch_range(repo, a, mid, out)
        fetch_range(repo, mid + dt.timedelta(days=1), b, out)
        return
    after = None
    while True:
        d = gql(Q_PAGE, q=q, after=after)["data"]["search"]
        out.extend(x for x in d["nodes"] if x)
        if not d["pageInfo"]["hasNextPage"]:
            break
        after = d["pageInfo"]["endCursor"]


def main():
    RAW.mkdir(exist_ok=True)
    for repo in sys.argv[1:] or NAMES:
        w = window(repo)
        if w is None:
            print(repo, "skipped: no T-1 snapshot", flush=True)
            continue
        t1_ts, tend_ts = w
        a, b = day(t1_ts), day(tend_ts)
        out = []
        cur = a
        while cur <= b:  # month-sized chunks, split further when > 1000 hits
            nxt = min(b, cur + dt.timedelta(days=90))
            fetch_range(repo, cur, nxt, out)
            cur = nxt + dt.timedelta(days=1)
        seen = {}
        for p in out:
            seen[p["number"]] = p
        prs = []
        for p in seen.values():
            prs.append(dict(number=p["number"], created=p["createdAt"], merged=p["mergedAt"],
                            author=(p.get("author") or {}).get("login"),
                            merge_commit=(p.get("mergeCommit") or {}).get("oid"),
                            n_commits=p["commits"]["totalCount"],
                            commit_oids=[n["commit"]["oid"] for n in p["commits"]["nodes"]],
                            n_reviews=p["reviews"]["totalCount"],
                            n_review_comments=sum(n["comments"]["totalCount"] for n in p["reviews"]["nodes"])))
        (RAW / f"prs-{repo}.json").write_text(json.dumps(prs))
        print(repo, a, b, len(prs), flush=True)


if __name__ == "__main__":
    main()
