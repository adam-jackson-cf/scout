#!/usr/bin/env python3
"""E1 inertia-v1: build per-file features and outcomes per PREREGISTRATION.md.

Usage: uv run --with lizard,numpy,pandas python collect.py [repo ...]
Stage order is enforced: exclusions.json must exist (written before outcomes).
Outputs: features.csv, commits.jsonl.gz, meta.json (in this directory).
"""
import concurrent.futures as cf
import gzip
import json
import os
import re
import shutil
import subprocess
import sys
import time
from collections import defaultdict
from math import sqrt
from pathlib import Path

import lizard

HERE = Path(__file__).resolve().parent
REPOS = Path("/tmp/scout-evidence/repos")
SNAP = Path("/tmp/scout-evidence/inertia-snap")
CACHE = Path("/tmp/scout-evidence/inertia-cache")
NAMES = ["fastapi", "webpack", "nest", "caddy", "axum", "junit5", "eShopOnWeb"]
EXTS = {".py", ".js", ".mjs", ".cjs", ".ts", ".tsx", ".go", ".rs", ".java", ".cs"}
JSCPD_FORMATS = "python,javascript,typescript,tsx,go,rust,java,csharp"
DAY = 86400
FIX_RE = re.compile(r"(?i)\b(fix(e[sd]|ing)?|bug|regression|revert(ed|s)?|hotfix|crash(es|ed)?|broken|incorrect|wrong)\b")
BULK = 50


def git(repo, *args, check=True):
    r = subprocess.run(["git", "-c", "core.quotepath=off", "-C", str(REPOS / repo), *args],
                       capture_output=True, text=True, errors="replace", check=False)
    if check and r.returncode:
        raise RuntimeError(f"git {args}: {r.stderr[:300]}")
    return r.stdout


def glob_re(g):
    out, i = "", 0
    while i < len(g):
        if g.startswith("**/", i):
            out += "(?:.*/)?"; i += 3
        elif g.startswith("**", i):
            out += ".*"; i += 2
        elif g[i] == "*":
            out += "[^/]*"; i += 1
        elif g[i] == "?":
            out += "[^/]"; i += 1
        else:
            out += re.escape(g[i]); i += 1
    return re.compile("^" + out + "$")


def is_fix(text):
    return bool(FIX_RE.search(text)) or text.lstrip().lower().startswith("revert")


def is_fix_msg(subject, body):
    return bool(FIX_RE.search(subject + "\n" + body)) or subject.startswith("Revert")


def parse_log_status(out):
    """-> list of (hash, ct, email, [(status, old, new)])"""
    commits = []
    cur = None
    for line in out.splitlines():
        if line.startswith("@@"):
            h, ct, em = line[2:].split("\t")
            cur = (h, int(ct), em, [])
            commits.append(cur)
        elif line and cur is not None:
            p = line.split("\t")
            st = p[0]
            if st[0] in "RC":
                cur[3].append((st, p[1], p[2]))
            else:
                cur[3].append((st[0], p[1], p[1]))
    return commits


def parse_numstat(out):
    """-> {hash: {newpath: (added, removed)}}"""
    res, cur = {}, None
    for line in out.splitlines():
        if line.startswith("@@"):
            cur = res.setdefault(line[2:], {})
        elif line and cur is not None:
            a, d, p = line.split("\t", 2)
            if "{" in p:
                p = re.sub(r"\{(.*?) => (.*?)\}", lambda m: m.group(2), p).replace("//", "/")
            elif " => " in p:
                p = p.split(" => ")[1]
            cur[p] = (0 if a == "-" else int(a), 0 if d == "-" else int(d))
    return res


# --------------------------------------------------------------------------- lizard
def _liz(paths):
    out = []
    for fi in lizard.analyze_files(paths, threads=1):
        funcs = [(f.cyclomatic_complexity, f.nloc) for f in fi.function_list]
        out.append((fi.filename, fi.nloc, funcs))
    return out


def run_lizard(root, rels):
    paths = [str(root / r) for r in rels]
    chunks = [paths[i:i + 40] for i in range(0, len(paths), 40)]
    res = {}
    with cf.ProcessPoolExecutor(max_workers=8) as ex:
        for part in ex.map(_liz, chunks):
            for fn, nloc, funcs in part:
                res[os.path.relpath(fn, root)] = (nloc, funcs)
    return res


# --------------------------------------------------------------------------- rework
_WIN = set()


def _init_win(w):
    _WIN.update(w)


def _rework_job(args):
    repo, h, items = args
    win_keys = _WIN  # items: [(id, oldpath, newpath)]
    parents = git(repo, "rev-list", "--parents", "-n", "1", h).split()
    if len(parents) < 2:
        return []
    parent = parents[1]
    diff = git(repo, "diff", "-U0", "-M", "--no-color", parent, h)
    byfile = defaultdict(list)  # newpath -> (oldpath, [(s,c)])
    oldp = newp = None
    for line in diff.splitlines():
        if line.startswith("diff --git "):
            oldp = newp = None
        elif line.startswith("--- "):
            oldp = line[4:].split("\t")[0]
            oldp = oldp[2:] if oldp.startswith("a/") else None
        elif line.startswith("+++ "):
            n = line[4:].split("\t")[0]
            newp = n[2:] if n.startswith("b/") else None
        elif line.startswith("@@") and oldp:
            m = re.match(r"@@ -(\d+)(?:,(\d+))? ", line)
            s, c = int(m.group(1)), int(m.group(2)) if m.group(2) is not None else 1
            if c > 0:
                byfile[newp if newp else oldp].append((oldp, s, c))
    res = []
    for fid, op, np_ in items:
        hunks = byfile.get(np_) or byfile.get(op)
        if not hunks:
            continue
        old_path = hunks[0][0]
        total = sum(c for _, _, c in hunks)
        inwin = 0
        ranges = [f"{s},{s + c - 1}" for _, s, c in hunks]
        for i in range(0, len(ranges), 100):
            args = ["blame", "-l", "-s"]
            for r in ranges[i:i + 100]:
                args += ["-L", r]
            args += [parent, "--", old_path]
            out = git(repo, *args, check=False)
            for line in out.splitlines():
                tok = line.split(" ", 1)[0].lstrip("^")
                if tok[:39] in win_keys:
                    inwin += 1
        res.append((fid, total, inwin))
    return res


# --------------------------------------------------------------------------- jscpd
def run_jscpd(root, outdir, timeout=3000):
    outdir.mkdir(parents=True, exist_ok=True)
    cmd = ["jscpd", str(root), "--format", JSCPD_FORMATS, "--reporters", "json", "--output", str(outdir),
           "--silent", "--max-lines", "10000", "--max-size", "500kb"]
    # default min-tokens (50) / min-lines (5) kept
    t = time.time()
    r = subprocess.run(cmd, capture_output=True, text=True, timeout=timeout)
    rep = outdir / "jscpd-report.json"
    if not rep.exists():
        raise RuntimeError(f"jscpd failed rc={r.returncode}: {r.stderr[-300:]}")
    data = json.loads(rep.read_text())
    lines = defaultdict(set)
    for d in data.get("duplicates", []):
        for k in ("firstFile", "secondFile"):
            f = d[k]
            name = os.path.relpath(os.path.abspath(f["name"]), root)  # jscpd names are relative to cwd
            s = f.get("start", f.get("startLoc", {}).get("line"))
            e = f.get("end", f.get("endLoc", {}).get("line"))
            lines[name].update(range(s, e + 1))
    return {k: len(v) for k, v in lines.items()}, time.time() - t


# --------------------------------------------------------------------------- main per repo
def process(repo, excl):
    t_start = time.time()
    meta = {"repo": repo}
    tend = git(repo, "rev-parse", "HEAD").strip()
    tend_ts = int(git(repo, "log", "-1", "--format=%ct", tend))
    t0 = git(repo, "rev-list", "--first-parent", "-1", f"--before={tend_ts - 365 * DAY}", tend).strip()
    t0_ts = int(git(repo, "log", "-1", "--format=%ct", t0))
    prior_start = t0_ts - 365 * DAY
    meta.update(t_end=tend, t_end_ts=tend_ts, t0=t0, t0_ts=t0_ts)
    ex_re = [glob_re(g) for g in excl["repos"][repo]]

    def ok(p):
        return Path(p).suffix in EXTS and not any(r.match(p) for r in ex_re)

    files0 = [p for p in git(repo, "ls-tree", "-r", "--name-only", t0).splitlines() if ok(p)]
    files_end = set(git(repo, "ls-tree", "-r", "--name-only", tend).splitlines())
    meta["n_candidate_t0"] = len(files0)

    # snapshot
    snap = SNAP / repo
    if snap.exists():
        shutil.rmtree(snap)
    snap.mkdir(parents=True)
    ar = subprocess.Popen(["git", "-C", str(REPOS / repo), "archive", t0], stdout=subprocess.PIPE)
    subprocess.run(["tar", "-x", "-C", str(snap)], stdin=ar.stdout, check=True)
    ar.wait()

    # lizard
    liz = run_lizard(snap, files0)
    rows = {}
    for p in files0:
        if p not in liz:
            continue
        nloc, funcs = liz[p]
        if nloc < 10:
            continue
        num = sum(cc * sqrt(n) for cc, n in funcs)
        hi = sum(cc * sqrt(n) for cc, n in funcs if cc > 10)
        rows[p] = dict(sloc=nloc, n_funcs=len(funcs),
                       erosion=1.0 if num == 0 else 1 - hi / num,
                       cc_max=max((cc for cc, _ in funcs), default=0),
                       n_high_cc=sum(1 for cc, _ in funcs if cc > 10))
    meta["n_sloc_ge10"] = len(rows)

    # prune snapshot to analysed files and run jscpd
    keep = set(rows)
    pruned = SNAP / (repo + "-pruned")
    if pruned.exists():
        shutil.rmtree(pruned)
    for p in keep:
        dst = pruned / p
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(snap / p, dst)
    try:
        dup, secs = run_jscpd(pruned, CACHE / ("jscpd-" + repo))
        meta["jscpd_seconds"] = round(secs, 1)
        for p, r in rows.items():
            r["clone_pct"] = min(1.0, dup.get(p, 0) / r["sloc"])
    except Exception as e:  # recorded as documented deviation
        meta["jscpd_error"] = str(e)[:300]
        for r in rows.values():
            r["clone_pct"] = float("nan")
    shutil.rmtree(pruned, ignore_errors=True)
    shutil.rmtree(snap, ignore_errors=True)

    # ---- history walk <= T0 (newest first, backwards path tracking)
    st = git(repo, "log", "--no-merges", "-M", "--topo-order", "--name-status", "--format=@@%H%x09%ct%x09%ae", t0)
    commits = parse_log_status(st)
    ns = parse_numstat(git(repo, "log", "--no-merges", "-M", "--topo-order", "--numstat", "--format=@@%H", t0))
    msgs = {}
    for rec in git(repo, "log", "--no-merges", "--format=%H%x1f%s%x1f%b%x1e", t0, f"{tend}").split("\x1e"):
        rec = rec.strip("\n")
        if rec:
            h, s, b = (rec.split("\x1f") + ["", ""])[:3]
            msgs[h] = (s, b)
    pmap = {p: p for p in rows}
    stats = defaultdict(lambda: dict(prior_commits=0, prior_fix=0, prior_lines=0, first_ts=None))
    authors = defaultdict(set)
    win_keys = set()
    touches = defaultdict(list)  # hash -> [(id, old, new)]
    for h, ct, em, ents in commits:
        in_prior = ct > prior_start
        bulk = len(ents) > BULK
        if in_prior and not bulk:
            win_keys.add(h[:39])
        fix = False
        if in_prior and not bulk:
            s, b = msgs.get(h, ("", ""))
            fix = is_fix_msg(s, b)
        for stt, old, new in ents:
            fid = None
            if stt[0] == "R":
                if new in pmap:
                    fid = pmap.pop(new)
                    pmap[old] = fid
            elif stt[0] == "C":
                pmap.pop(new, None)
                continue
            elif stt == "A":
                fid = pmap.pop(new, None)
            elif stt == "D":
                continue
            else:
                fid = pmap.get(new)
            if fid is None:
                continue
            S = stats[fid]
            S["first_ts"] = ct if S["first_ts"] is None else min(S["first_ts"], ct)
            if in_prior and not bulk and stt != "R100":
                S["prior_commits"] += 1
                S["prior_fix"] += int(fix)
                a, d = ns.get(h, {}).get(new, (0, 0))
                S["prior_lines"] += a + d
                authors[fid].add(em)
                if stt[0] in "MR":
                    touches[h].append((fid, old, new))
    # commit sample material (prior + outcome windows)
    # ---- rework via blame
    jobs = [(repo, h, items) for h, items in touches.items()]
    rw = defaultdict(lambda: [0, 0])
    with cf.ProcessPoolExecutor(max_workers=8, initializer=_init_win, initargs=(win_keys,)) as ex:
        for part in ex.map(_rework_job, jobs, chunksize=4):
            for fid, tot, inw in part:
                rw[fid][0] += tot
                rw[fid][1] += inw
    meta["n_blame_jobs"] = len(jobs)

    for p, r in rows.items():
        S = stats.get(p)
        r["prior_commits"] = S["prior_commits"] if S else 0
        r["prior_fix"] = S["prior_fix"] if S else 0
        r["prior_lines"] = S["prior_lines"] if S else 0
        r["age_days"] = (t0_ts - S["first_ts"]) / DAY if S and S["first_ts"] else 0.0
        r["prior_authors"] = len(authors.get(p, ()))
        tot, inw = rw.get(p, (0, 0))
        r["rework_removed"] = tot
        r["rework"] = inw / tot if tot else 0.0

    # ---- outcome walk (T0, T_end] forward
    st2 = git(repo, "log", "--no-merges", "-M", "--topo-order", "--reverse", "--name-status",
              "--format=@@%H%x09%ct%x09%ae", f"{t0}..{tend}")
    oc = parse_log_status(st2)
    fmap = {p: p for p in rows}
    yfix, ytot = defaultdict(int), defaultdict(int)
    for h, ct, em, ents in oc:
        bulk = len(ents) > BULK
        s, b = msgs.get(h, ("", ""))
        fix = (not bulk) and is_fix_msg(s, b)
        for stt, old, new in ents:
            fid = None
            if stt[0] == "R":
                if old in fmap:
                    fid = fmap.pop(old)
                    fmap[new] = fid
            elif stt[0] == "D":
                fmap.pop(old, None)
                continue
            elif stt[0] in "AC":
                continue
            else:
                fid = fmap.get(old)
            if fid is None or bulk or stt == "R100":
                continue
            ytot[fid] += 1
            yfix[fid] += int(fix)
    surv = {fid: path for path, fid in fmap.items() if path in files_end}
    meta["n_outcome_commits"] = len(oc)

    out = []
    for p, r in rows.items():
        if p not in surv:
            continue
        r.update(repo=repo, path=p, path_end=surv[p], fix_commits=yfix.get(p, 0), commits=ytot.get(p, 0))
        out.append(r)
    meta["n_survivors"] = len(out)

    # commit material for classifier validation (non-merge, <=50 files, in prior or outcome window)
    cm = []
    nfiles = {h: len(e) for h, _, _, e in commits}
    nfiles.update({h: len(e) for h, _, _, e in oc})
    allc = {h: ct for h, ct, _, _ in commits}
    allc.update({h: ct for h, ct, _, _ in oc})
    prior_set = {c[0] for c in commits}
    for h, ct in allc.items():
        if ct > prior_start and nfiles[h] <= BULK and h in msgs:
            s, b = msgs[h]
            cm.append(dict(repo=repo, hash=h, ts=ct, nfiles=nfiles[h], subject=s, body=b[:1500],
                           window="prior" if ct <= t0_ts and h in prior_set else "outcome",
                           regex_fix=is_fix_msg(s, b)))
    meta["seconds"] = round(time.time() - t_start, 1)
    return out, cm, meta


def main():
    excl = json.loads((HERE / "exclusions.json").read_text())
    names = sys.argv[1:] or NAMES
    CACHE.mkdir(parents=True, exist_ok=True)
    for repo in names:
        f = CACHE / f"{repo}.json"
        if f.exists():
            print("cached", repo)
            continue
        print("processing", repo, flush=True)
        rows, cm, meta = process(repo, excl)
        f.write_text(json.dumps(dict(rows=rows, commits=cm, meta=meta), default=lambda x: None if x != x else x))
        print(meta, flush=True)
    import pandas as pd
    allrows, allc, metas = [], [], []
    for repo in NAMES:
        f = CACHE / f"{repo}.json"
        if f.exists():
            d = json.loads(f.read_text())
            allrows += d["rows"]; allc += d["commits"]; metas.append(d["meta"])
    pd.DataFrame(allrows).to_csv(HERE / "features.csv", index=False)
    with gzip.open(HERE / "commits.jsonl.gz", "wt") as fh:
        for c in allc:
            fh.write(json.dumps(c) + "\n")
    (HERE / "meta.json").write_text(json.dumps(metas, indent=1))


if __name__ == "__main__":
    main()
