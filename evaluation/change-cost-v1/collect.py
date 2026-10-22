#!/usr/bin/env python3
"""E1b change-cost-v1 stage 1: git/lizard/jscpd collection per repository (adapted from ../inertia-v1/collect.py).

Usage: uv run --with lizard,numpy,pandas python collect.py [repo ...]
Requires areas.json (written before outcomes). Writes /tmp/scout-evidence/e1b/cache/<repo>.json.gz and
the per-repo metadata copy meta.json here. features.py consumes the cache plus raw/prs-<repo>.json.
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
WORK = Path("/tmp/scout-evidence/e1b")
SNAP = WORK / "snap"
CACHE = WORK / "cache"
NAMES = ["fastapi", "webpack", "nest", "caddy", "axum", "junit5", "eShopOnWeb"]
EXTS = {".py", ".js", ".mjs", ".cjs", ".ts", ".tsx", ".go", ".rs", ".java", ".cs"}
JSCPD_FORMATS = "python,javascript,typescript,tsx,go,rust,java,csharp"
DAY = 86400
BULK = 50


def git(repo, *args):
    r = subprocess.run(["git", "-c", "core.quotepath=off", "-C", str(REPOS / repo), *args],
                       capture_output=True, text=True, errors="replace")
    if r.returncode:
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


def parse_log_status(out):
    """-> list of (hash, ct, author_name, [(status, old, new)]) in log order"""
    commits, cur = [], None
    for line in out.splitlines():
        if line.startswith("@@"):
            h, ct, an = line[2:].split("\t", 2)
            cur = (h, int(ct), an, [])
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


def _liz(paths):
    out = []
    for fi in lizard.analyze_files(paths, threads=1):
        out.append((fi.filename, fi.nloc, [(f.cyclomatic_complexity, f.nloc) for f in fi.function_list]))
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


def run_jscpd(root, outdir, timeout=3000):
    if outdir.exists():
        shutil.rmtree(outdir)
    outdir.mkdir(parents=True)
    cmd = ["jscpd", str(root), "--format", JSCPD_FORMATS, "--reporters", "json", "--output", str(outdir),
           "--silent", "--max-lines", "10000", "--max-size", "500kb"]
    r = subprocess.run(cmd, capture_output=True, text=True, timeout=timeout)
    rep = outdir / "jscpd-report.json"
    if not rep.exists():
        raise RuntimeError(f"jscpd failed rc={r.returncode}: {r.stderr[-300:]}")
    data = json.loads(rep.read_text())
    lines = defaultdict(set)
    for d in data.get("duplicates", []):
        for k in ("firstFile", "secondFile"):
            f = d[k]
            name = os.path.relpath(os.path.abspath(f["name"]), root)
            s = f.get("start", f.get("startLoc", {}).get("line"))
            e = f.get("end", f.get("endLoc", {}).get("line"))
            lines[name].update(range(s, e + 1))
    return {k: len(v) for k, v in lines.items()}


def snapshot_metrics(repo, rev, tag, ok, meta):
    """lizard + jscpd on eligible files (>=10 SLOC) of `rev`. -> {path: dict(sloc, num, hi, nhigh, dup)}"""
    files = [p for p in git(repo, "ls-tree", "-r", "--name-only", rev).splitlines() if ok(p)]
    snap = SNAP / f"{repo}-{tag}"
    if snap.exists():
        shutil.rmtree(snap)
    snap.mkdir(parents=True)
    ar = subprocess.Popen(["git", "-C", str(REPOS / repo), "archive", rev], stdout=subprocess.PIPE)
    subprocess.run(["tar", "-x", "-C", str(snap)], stdin=ar.stdout, check=True)
    ar.wait()
    liz = run_lizard(snap, files)
    rows = {}
    for p in files:
        if p not in liz or liz[p][0] < 10:
            continue
        nloc, funcs = liz[p]
        rows[p] = dict(sloc=nloc, num=sum(cc * sqrt(n) for cc, n in funcs),
                       hi=sum(cc * sqrt(n) for cc, n in funcs if cc > 10),
                       nhigh=sum(1 for cc, _ in funcs if cc > 10))
    meta[f"n_candidate_{tag}"] = len(files)
    meta[f"n_sloc10_{tag}"] = len(rows)
    pruned = SNAP / f"{repo}-{tag}-pruned"
    if pruned.exists():
        shutil.rmtree(pruned)
    for p in rows:
        dst = pruned / p
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(snap / p, dst)
    t = time.time()
    try:
        dup = run_jscpd(pruned, WORK / "jscpd" / f"{repo}-{tag}")
        for p, r in rows.items():
            r["dup"] = min(r["sloc"], dup.get(p, 0))
    except Exception as e:  # documented as deviation if it happens
        meta[f"jscpd_error_{tag}"] = str(e)[:300]
        for r in rows.values():
            r["dup"] = None
    meta[f"jscpd_seconds_{tag}"] = round(time.time() - t, 1)
    shutil.rmtree(pruned, ignore_errors=True)
    shutil.rmtree(snap, ignore_errors=True)
    return rows


def process(repo, excl, cfg):
    t_start = time.time()
    meta = {"repo": repo}
    bot_re = re.compile(cfg["bot_author_regex"])
    tend = git(repo, "rev-parse", "HEAD").strip()
    tend_ts = int(git(repo, "log", "-1", "--format=%ct", tend))
    t0 = git(repo, "rev-list", "--first-parent", "-1", f"--before={tend_ts - 365 * DAY}", tend).strip()
    t0_ts = int(git(repo, "log", "-1", "--format=%ct", t0))
    t1 = git(repo, "rev-list", "--first-parent", "-1", f"--before={t0_ts - 365 * DAY}", t0).strip()
    if not t1:
        meta.update(t_end=tend, t0=t0, t0_ts=t0_ts,
                    skipped="no first-parent commit at or before T0-365d: history starts after T-1")
        return dict(meta=meta, skipped=True)
    t1_ts = int(git(repo, "log", "-1", "--format=%ct", t1))
    meta.update(t_end=tend, t_end_ts=tend_ts, t0=t0, t0_ts=t0_ts, t1=t1, t1_ts=t1_ts)
    ex_re = [glob_re(g) for g in excl["repos"][repo]]

    def ok(p):
        return Path(p).suffix in EXTS and not any(r.match(p) for r in ex_re)

    m0 = snapshot_metrics(repo, t0, "t0", ok, meta)
    m1 = snapshot_metrics(repo, t1, "t1", ok, meta)

    FMT = "--format=@@%H%x09%ct%x09%an"
    # ---- backward walk, prior window (t1..t0): T0 file ids, path-at-T-1, per-file prior touches
    prior_log = parse_log_status(git(repo, "log", "--no-merges", "-M", "--topo-order", "--name-status", FMT, f"{t1}..{t0}"))
    prior_ns = parse_numstat(git(repo, "log", "--no-merges", "-M", "--topo-order", "--numstat", "--format=@@%H", f"{t1}..{t0}"))
    pmap = {p: p for p in m0}
    first_ts = {}
    fprior = defaultdict(list)  # fid -> [(hash, path_at_commit, add, del)]

    def back_step(h, ct, ents, record):
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
            first_ts[fid] = ct if fid not in first_ts else min(first_ts[fid], ct)
            if record and stt != "R100":
                a, d = prior_ns.get(h, {}).get(new, (0, 0))
                fprior[fid].append((h, new, a, d))

    for h, ct, an, ents in prior_log:
        back_step(h, ct, ents, True)
    path_t1 = {fid: p for p, fid in pmap.items()}  # fid -> path at T-1 (absent => created in prior window)
    for h, ct, an, ents in parse_log_status(git(repo, "log", "--no-merges", "-M", "--topo-order", "--name-status", FMT, t1)):
        back_step(h, ct, ents, False)

    # ---- outcome window forward walk (t0..tend)
    out_log = parse_log_status(git(repo, "log", "--no-merges", "-M", "--topo-order", "--reverse", "--name-status", FMT, f"{t0}..{tend}"))
    out_ns = parse_numstat(git(repo, "log", "--no-merges", "-M", "--topo-order", "--numstat", "--format=@@%H", f"{t0}..{tend}"))
    fmap = {p: p for p in m0}
    fout = defaultdict(list)  # fid -> [(hash, path_at_commit, add, del)]
    for h, ct, an, ents in out_log:
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
            if fid is None or stt == "R100":
                continue
            a, d = out_ns.get(h, {}).get(new, (0, 0))
            fout[fid].append((h, new, a, d))

    # ---- commit table (oldest first within each window)
    def table(log, ns, win):
        rows = []
        for seq, (h, ct, an, ents) in enumerate(log):
            el = []
            for stt, old, new in ents:
                if stt[0] == "D":
                    paths = [(old, 0, 0)]
                else:
                    a, d = ns.get(h, {}).get(new, (0, 0))
                    paths = [(new, a, d)]
                for p, a, d in paths:
                    if ok(p):
                        el.append([stt, old, p, a, d])
                if stt[0] == "R" and ok(old) and old != new:
                    el.append(["R-old", old, old, 0, 0])
            rows.append(dict(h=h, ts=ct, an=an, bot=bool(bot_re.search(an)), win=win, seq=seq,
                             nfiles=len(ents), ents=el))
        return rows

    commits = table(list(reversed(prior_log)), prior_ns, "prior") + table(out_log, out_ns, "outcome")
    meta["n_prior_commits"] = len(prior_log)
    meta["n_outcome_commits"] = len(out_log)
    meta["n_bot_commits"] = sum(c["bot"] for c in commits)
    meta["seconds"] = round(time.time() - t_start, 1)
    data = dict(meta=meta, m0=m0, m1=m1, first_ts=first_ts, path_t1=path_t1,
                fprior=fprior, fout=fout, commits=commits)
    return data


def main():
    excl = json.loads((HERE / "exclusions.json").read_text())
    cfg = json.loads((HERE / "areas.json").read_text())
    CACHE.mkdir(parents=True, exist_ok=True)
    metas = []
    for repo in sys.argv[1:] or NAMES:
        f = CACHE / f"{repo}.json.gz"
        if not f.exists():
            print("processing", repo, flush=True)
            data = process(repo, excl, cfg)
            with gzip.open(f, "wt") as fh:
                json.dump(data, fh)
            print(data["meta"], flush=True)
        else:
            print("cached", repo)
    for repo in NAMES:
        f = CACHE / f"{repo}.json.gz"
        if f.exists():
            with gzip.open(f, "rt") as fh:
                metas.append(json.load(fh)["meta"])
    (HERE / "meta.json").write_text(json.dumps(metas, indent=1))


if __name__ == "__main__":
    main()
