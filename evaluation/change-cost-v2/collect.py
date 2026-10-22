#!/usr/bin/env python3
"""E1b-v2 stage 2: per-PR structural measurement from git (lizard on the blobs the PR touched).

Usage: uv run --with lizard python collect.py [repo ...]
Reads pins.json + raw/prs-<repo>.json.gz + local clones; writes data/pr-<repo>.jsonl.gz (one row per fetched PR in
the window; `skip` names the exclusion reason, counted later). No outcome-predictor relationship is computed here.

Per PR:
  S  = snapshot commit = parent of the PR's first commit if present in the clone, else baseRefOid, else unmeasured.
  F_S = eligible files the PR modifies/deletes that exist at S with >= 10 NLOC (lizard).
  F_stable = F_S files that also exist at S_trail (last first-parent commit of S at least 365 d before S) with >= 10 NLOC
             under the SAME path. Degradation is computed on this fixed population only.
  Hidden-coupling exposure from co-change history in (S_ts-365 d, S_ts] (no later information).
"""
import bisect
import gzip
import json
import math
import multiprocessing as mp
import re
import subprocess
import sys
import time
from collections import Counter, defaultdict

from lizard import analyze_file

from common import (BULK_FILES, COCHANGE_DAYS, DAY, HERE, NAMES, REPO_CFG, REPOS, TRAIL_DAYS, eligible, git,
                    is_bot, load_pins)

OUT = HERE / "data"
MAXBLOB = 400_000
MIN_NLOC = 10
CONF_MIN, SUPPORT_MIN, TOP_PARTNERS = 0.3, 3, 10
GENERIC_STEMS = {"__init__", "index", "mod", "lib", "main", "utils", "util", "types", "type", "config", "common", "base",
                 "helpers", "helper", "constants", "const", "core", "models", "model", "api", "app", "server", "client"}

G = {}  # per-process globals: repo, hist arrays


def iso_ts(s):
    import datetime as dt
    return int(dt.datetime.fromisoformat(s.replace("Z", "+00:00")).timestamp())


def load_history(repo, sha):
    """Non-merge history up to sha (oldest first): ts, author email, eligible file tuple, total file count."""
    out = git(repo, "log", "--no-merges", "--name-only", "--format=@@%H\t%ct\t%ae", sha)
    commits, cur = [], None
    for line in out.splitlines():
        if line.startswith("@@"):
            h, ct, ae = line[2:].split("\t", 2)
            cur = [h, int(ct), ae, [], 0]
            commits.append(cur)
        elif line and cur is not None:
            cur[4] += 1
            if eligible(line):
                cur[3].append(line)
    commits.sort(key=lambda c: c[1])
    return commits


def build_index(commits):
    ts = [c[1] for c in commits]
    by_file = defaultdict(lambda: ([], []))  # path -> (commit idx asc, their ts asc); commits are sorted by ts
    authors = defaultdict(list)
    for i, c in enumerate(commits):
        authors[c[2]].append(c[1])
        if c[4] <= BULK_FILES:
            for p in c[3]:
                by_file[p][0].append(i)
                by_file[p][1].append(c[1])
    return ts, dict(by_file), authors


def catfile_batch(repo, specs):
    """specs: list of 'rev:path' or 'rev'. -> list of bytes|None (None = missing)."""
    if not specs:
        return []
    inp = ("\n".join(specs) + "\n").encode()
    r = subprocess.run(["git", "-C", str(REPOS / REPO_CFG[repo][1]), "cat-file", "--batch"],
                       input=inp, capture_output=True)
    buf, pos, res = r.stdout, 0, []
    for _ in specs:
        nl = buf.index(b"\n", pos)
        head = buf[pos:nl].split()
        pos = nl + 1
        if len(head) == 2 and head[1] == b"missing":
            res.append(None)
            continue
        size = int(head[2])
        res.append((head[1].decode(), buf[pos:pos + size]))
        pos += size + 1
    return res


def measure(path, data):
    """-> (nloc, hi_mass, n_high) or None when unreadable/too big."""
    if data is None or len(data) > MAXBLOB:
        return None
    try:
        fi = analyze_file.analyze_source_code(path, data.decode("utf-8", "replace"))
    except Exception:
        return None
    hi = sum(f.cyclomatic_complexity * math.sqrt(max(f.nloc, 1)) for f in fi.function_list if f.cyclomatic_complexity > 10)
    nh = sum(1 for f in fi.function_list if f.cyclomatic_complexity > 10)
    return fi.nloc, hi, nh


def blob_metrics(repo, rev, paths):
    got = catfile_batch(repo, [f"{rev}:{p}" for p in paths])
    out = {}
    for p, g in zip(paths, got):
        out[p] = (None, None) if g is None or g[0] != "blob" else (measure(p, g[1]), g[1])
    return out


def stem(p):
    base = p.rsplit("/", 1)[-1].rsplit(".", 1)[0]
    if base in GENERIC_STEMS and "/" in p:
        base = p.rsplit("/", 2)[-2]   # generic file names are identified by their directory name
    return base


_COMMENT = re.compile(r"^\s*(#|//|\*|/\*)")


def referenced(text_f, path_g):
    """Proxy for a static reference: the stem of g occurs as a whole word on a non-comment line of f. Unvalidated by itself:
    gate G7 (annotators, by language) decides whether T5 may be decided."""
    s = stem(path_g)
    if len(s) < 3 or s in GENERIC_STEMS:
        return False
    body = "\n".join(l for l in text_f.splitlines() if not _COMMENT.match(l))
    return re.search(r"(?<![A-Za-z0-9_])" + re.escape(s) + r"(?![A-Za-z0-9_])", body) is not None


def process(pr):
    repo = G["repo"]
    base = dict(repo=repo, number=pr["number"], created=pr["created"], merged=pr["merged"], author=pr["author"],
                assoc=pr["assoc"], title=pr["title"], labels=pr["labels"], n_commits=pr["n_commits"],
                n_files=pr["n_files"], additions=pr["additions"], deletions=pr["deletions"],
                n_reviews=pr["n_reviews"])
    if is_bot(pr["author"], pr["author_type"]):
        return dict(base, skip="bot")
    if pr["n_files"] > BULK_FILES or len(pr["files"]) < pr["n_files"]:
        return dict(base, skip="bulk")
    if pr["n_commits"] > len(pr["commits"]) or not pr["commits"]:
        return dict(base, skip="many_commits")
    elig = [f for f in pr["files"] if eligible(f["path"])]
    if not elig:
        return dict(base, skip="no_eligible_file")
    # snapshot
    cands = [c for c in (pr["commits"][0].get("parent"), pr["base_oid"]) if c]
    got = catfile_batch(repo, cands)
    S, Ssrc = None, None
    for i, (c, g) in enumerate(zip(cands, got)):
        if g is not None and g[0] == "commit":
            S, Ssrc = c, "parent" if c == pr["commits"][0].get("parent") else "base"
            break
    if S is None:
        return dict(base, skip="no_snapshot")
    S_ts = int(git(repo, "show", "-s", "--format=%ct", S).strip())
    churn_all = pr["additions"] + pr["deletions"]
    ech = sum(f["add"] + f["dele"] for f in elig)
    n_new = sum(1 for f in elig if f["ct"] == "ADDED")
    n_ren = sum(1 for f in elig if f["ct"] in ("RENAMED", "COPIED"))
    old = [f for f in elig if f["ct"] in ("MODIFIED", "CHANGED", "DELETED")]
    base.update(S=S, S_src=Ssrc, S_ts=S_ts, churn=churn_all, elig_churn=ech, n_elig=len(elig), n_new=n_new,
                n_renamed=n_ren, n_old=len(old))
    if not old:
        return dict(base, skip="no_preexisting_file")
    paths = [f["path"] for f in old]
    mS = blob_metrics(repo, S, paths)
    FS = {p: m[0] for p, m in mS.items() if m[0] is not None and m[0][0] >= MIN_NLOC}
    if not FS:
        return dict(base, skip="no_measurable_file")
    # trailing snapshot
    trail = git(repo, "rev-list", "-1", "--first-parent", f"--before={S_ts - TRAIL_DAYS * DAY}", S, check=False).strip()
    files = []
    stable_paths = []
    mT = {}
    if trail:
        mT = blob_metrics(repo, trail, list(FS))
        stable_paths = [p for p in FS if mT[p][0] is not None and mT[p][0][0] >= MIN_NLOC]
    for p, m in FS.items():
        files.append(dict(p=p, nloc=m[0], hi=round(m[1], 3), nh=m[2], stable=p in stable_paths,
                          nloc_t=mT[p][0][0] if p in stable_paths else None,
                          hi_t=round(mT[p][0][1], 3) if p in stable_paths else None,
                          nh_t=mT[p][0][2] if p in stable_paths else None))
    row = dict(base, trail=trail or None, n_measured=len(FS), n_stable=len(stable_paths), files=files)
    # hidden-coupling exposure (history strictly before S_ts, 365 d)
    ts, by_file, authors = G["idx"]
    lo, hi_ts = S_ts - COCHANGE_DAYS * DAY, S_ts
    in_pr = {f["path"] for f in pr["files"]}
    commits = G["hist"]
    cand = {}  # partner path -> {best (file, n, conf), vias: {file: n}}
    for f in FS:
        idxs, tss = by_file.get(f, ([], []))
        a, b = bisect.bisect_right(tss, lo), bisect.bisect_right(tss, hi_ts)
        win = idxs[a:b]
        nf = len(win)
        if nf < SUPPORT_MIN:
            continue
        cnt = Counter()
        for i in win:
            for g in commits[i][3]:
                if g != f:
                    cnt[g] += 1
        top = sorted(((n, g) for g, n in cnt.items() if n >= SUPPORT_MIN and n / nf >= CONF_MIN and g not in in_pr),
                     key=lambda t: (-t[0] / nf, -t[0], t[1]))[:TOP_PARTNERS]
        for n, g in top:
            c = cand.setdefault(g, dict(best=(f, n, n / nf), vias={}))
            c["vias"][f] = n
            if n / nf > c["best"][2]:
                c["best"] = (f, n, n / nf)
    n_abs = n_hidden = 0
    partners = []
    if cand:
        gp = list(cand)
        gm = catfile_batch(repo, [f"{S}:{g}" for g in gp])
        fm = {f: d for f, (m, d) in mS.items() if f in FS}
        for g, gg in zip(gp, gm):
            if gg is None or gg[0] != "blob":
                continue  # partner did not exist at S
            n_abs += 1
            tg = gg[1].decode("utf-8", "replace") if len(gg[1]) <= MAXBLOB else ""
            # explained if ANY qualifying touched file f (co-changing with g) references g or is referenced by g; record the pair that triggered it
            via_ref = None
            for f in sorted(cand[g]["vias"]):
                if referenced(fm[f].decode("utf-8", "replace") if fm.get(f) else "", g) or referenced(tg, f):
                    via_ref = f
                    break
            explained = via_ref is not None
            n_hidden += 0 if explained else 1
            b = cand[g]["best"]
            partners.append(dict(g=g, via=b[0], vias=sorted(cand[g]["vias"]), via_ref=via_ref, n=b[1], conf=round(b[2], 3), explained=explained))
    row.update(n_partner_abs=n_abs, n_hidden=n_hidden, partners=partners)
    # author history (git author e-mail of the PR's first commit) before the PR's first commit
    em = pr["commits"][0].get("email")
    first_auth = min(iso_ts(c["authored"]) for c in pr["commits"])
    row["author_prior_commits"] = bisect.bisect_left(authors.get(em, []), first_auth) if em else None
    return row


def run_repo(repo, pins, procs):
    t0 = time.time()
    pin = pins[repo]
    prs = json.loads(gzip.open(HERE / "raw" / f"prs-{repo}.json.gz", "rt").read())
    prs = [p for p in prs if p["merged"] and pin["window_start_ts"] < iso_ts(p["merged"]) <= pin["t_end_ts"]]
    prs.sort(key=lambda p: p["number"])
    hist = load_history(repo, pin["sha"])
    idx = build_index(hist)
    G.update(repo=repo, hist=hist, idx=idx)
    ctx = mp.get_context("fork")
    with ctx.Pool(procs) as pool:
        rows = pool.map(process, prs, chunksize=8)
    OUT.mkdir(exist_ok=True)
    with gzip.open(OUT / f"pr-{repo}.jsonl.gz", "wt") as fh:
        for r in rows:
            fh.write(json.dumps(r) + "\n")
    c = Counter(r.get("skip", "ok") for r in rows)
    print(repo, len(prs), dict(c), f"{time.time() - t0:.0f}s", flush=True)


def main():
    pins = load_pins()
    for repo in sys.argv[1:] or NAMES:
        run_repo(repo, pins, 8)


if __name__ == "__main__":
    main()
