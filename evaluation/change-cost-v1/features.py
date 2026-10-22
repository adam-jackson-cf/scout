#!/usr/bin/env python3
"""E1b stage 2: areas, predictors and outcomes per PREREGISTRATION.md.

Usage: uv run --with numpy,pandas python features.py
Reads /tmp/scout-evidence/e1b/cache/<repo>.json.gz (collect.py), raw/prs-<repo>.json (fetch_prs.py), areas.json.
Writes features-area.csv, outcomes-area.csv, features-file.csv, outcomes-file.csv, dataset-counts.json.
"""
import bisect
import gzip
import json
import re
from collections import Counter, defaultdict
from pathlib import Path

import numpy as np
import pandas as pd

HERE = Path(__file__).resolve().parent
CACHE = Path("/tmp/scout-evidence/e1b/cache")
NAMES = ["fastapi", "webpack", "nest", "caddy", "axum", "junit5", "eShopOnWeb"]
BULK = 50
DAY = 86400
cfg = json.loads((HERE / "areas.json").read_text())
MIN_FILES, MIN_CHANGES, MIN_PR = cfg["min_area_files"], cfg["min_area_changes"], 5


# ------------------------------------------------------------------ areas
def split_root(path, roots):
    """-> (root_id, comps_below_root). Longest segment-prefix match; '*' matches any one segment."""
    d = path.split("/")[:-1]
    best = None
    for root in roots:
        rs = [] if root == "." else root.split("/")
        if len(rs) <= len(d) and all(a == "*" or a == b for a, b in zip(rs, d)):
            if best is None or len(rs) > len(best[0]):
                best = (d[:len(rs)], d[len(rs):])
    root_id = "/".join(best[0]) if best[0] else "."
    return root_id, tuple(best[1][:2])


def area_id(root, comps):
    return "/".join(([] if root == "." else [root]) + list(comps)) or "."


def raw_area(path, roots):
    r, c = split_root(path, roots)
    return area_id(r, c)


def parent_area(path_sample, roots, raw):
    r, c = split_root(path_sample, roots)
    return None if not c else area_id(r, c[:-1])


def build_area_map(m0, roots):
    """raw area -> final key (or None). Merge areas with < MIN_FILES eligible T0 files into the parent once."""
    cnt, sample = Counter(), {}
    for p in m0:
        a = raw_area(p, roots)
        cnt[a] += 1
        sample.setdefault(a, p)
    target = {}
    for a, n in cnt.items():
        target[a] = a if n >= MIN_FILES else parent_area(sample[a], roots, a)
    final = Counter()
    for a, t in target.items():
        if t is not None:
            final[t] += cnt[a]
    return {a: (t if t is not None and final[t] >= MIN_FILES else None) for a, t in target.items()}, cnt


# ------------------------------------------------------------------ metrics helpers
def metrics(files):
    """files: list of dict(sloc,num,hi,nhigh,dup) -> aggregate"""
    sloc = sum(f["sloc"] for f in files)
    num = sum(f["num"] for f in files)
    hi = sum(f["hi"] for f in files)
    dups = [f["dup"] for f in files]
    return dict(sloc=sloc, nfiles=len(files), erosion=1.0 if num == 0 else 1 - hi / num,
                nhigh=sum(f["nhigh"] for f in files),
                clone=float("nan") if any(d is None for d in dups) or sloc == 0 else sum(dups) / sloc)


def p90_ratio(xs, floor=1.0):
    xs = np.asarray(xs, float)
    return float(np.log1p(np.percentile(xs, 90) / max(np.median(xs), floor)))


def outcome_block(chs, prs):
    """chs: list of change dicts (churn, nel, follow); prs: list of PR dicts. -> dict of outcomes (NaN if undefined)"""
    o = dict(o1=np.nan, o2=np.nan, o3=np.nan, o4a=np.nan, o4b=np.nan, o4c=np.nan, o5a=np.nan, o5b=np.nan,
             n_changes=len(chs), n_prs=len(prs))
    if chs:
        ch = [c["churn"] for c in chs]
        o["o1"] = float(np.log1p(np.median(ch)))
        o["o2"] = float(np.log1p(np.median([c["nel"] for c in chs])))
        o["o3"] = float(np.mean([c["follow"] for c in chs]))
        o["o5a"] = p90_ratio(ch)
    if len(prs) >= MIN_PR:
        o["o4a"] = float(np.log1p(np.median([p["n_commits"] for p in prs])))
        o["o4b"] = float(np.log1p(np.median([p["n_reviews"] + p["n_review_comments"] for p in prs])))
        hrs = [p["hours"] for p in prs]
        o["o4c"] = float(np.log1p(np.median(hrs)))
        o["o5b"] = p90_ratio(hrs)
    return o


def follow_flags(stream, seq, ts, paths):
    for p in paths:
        L = stream.get(p)
        if not L:
            continue
        i = bisect.bisect_right(L[0], seq)
        for t in L[1][i:]:
            if t - ts <= 7 * DAY:
                return True
    return False


def run_repo(repo):
    with gzip.open(CACHE / f"{repo}.json.gz", "rt") as fh:
        D = json.load(fh)
    roots = cfg["roots"][repo]
    M0, M1 = D["m0"], D["m1"]
    amap, raw_cnt = build_area_map(M0, roots)

    def key_of(path):  # tracked final key or None
        return amap.get(raw_area(path, roots))

    def id_of(path):  # identifier used for coupling: final key if tracked, else raw area
        a = raw_area(path, roots)
        return amap.get(a) or "raw:" + a

    commits = {c["h"]: c for c in D["commits"]}
    valid = {h: c for h, c in commits.items() if not c["bot"] and c["nfiles"] <= BULK}
    # follow-up streams per window: path -> ([seq...], [ts...]) over all non-bot non-merge commits
    streams = {"prior": defaultdict(lambda: ([], [])), "outcome": defaultdict(lambda: ([], []))}
    for c in D["commits"]:
        if c["bot"]:
            continue
        for p in {e[2] for e in c["ents"]} | {e[1] for e in c["ents"] if e[0] == "R-old"}:
            s = streams[c["win"]][p]
            s[0].append(c["seq"]); s[1].append(c["ts"])
    # PRs
    prs = json.loads((HERE / "raw" / f"prs-{repo}.json").read_text())
    from datetime import datetime
    for p in prs:
        p["hours"] = max(0.0, (datetime.fromisoformat(p["merged"].replace("Z", "+00:00")) -
                               datetime.fromisoformat(p["created"].replace("Z", "+00:00"))).total_seconds() / 3600)
    by_oid = defaultdict(set)
    for i, p in enumerate(prs):
        for oid in ([p["merge_commit"]] if p["merge_commit"] else []) + p["commit_oids"]:
            by_oid[oid].add(i)

    # ---------------- area change table
    area_chg = {"prior": defaultdict(dict), "outcome": defaultdict(dict)}  # win -> key -> hash -> change dict
    for h, c in valid.items():
        paths_by_key = defaultdict(list)
        churn_by_key = defaultdict(int)
        for stt, old, p, a, d in c["ents"]:
            k = key_of(p)
            if k is None:
                continue
            paths_by_key[k].append(p)
            churn_by_key[k] += a + d
        if not paths_by_key:
            continue
        nel = len({e[2] for e in c["ents"] if e[0] != "R-old"})
        ids = {id_of(e[2]) for e in c["ents"]}
        for k, ps in paths_by_key.items():
            area_chg[c["win"]][k][h] = dict(
                churn=churn_by_key[k], nel=nel, an=c["an"], coupled=bool(ids - {k}),
                follow=int(follow_flags(streams[c["win"]], c["seq"], c["ts"], set(ps))))

    def linked_prs(win, k):
        s = set()
        for h in area_chg[win][k]:
            s |= by_oid.get(h, set())
        return [prs[i] for i in s]

    # ---------------- area-level snapshot aggregates
    t0_files, t1_files = defaultdict(list), defaultdict(list)
    for p, f in M0.items():
        k = key_of(p)
        if k:
            t0_files[k].append(p)
    for p, f in M1.items():
        k = key_of(p)
        if k:
            t1_files[k].append(p)
    t0 = D["meta"]["t0_ts"]
    prior_cnt_file = {fid: sum(1 for x in L if x[0] in valid) for fid, L in D["fprior"].items()}
    rows, counts = [], dict(repo=repo, raw_areas_t0=len(raw_cnt), areas_eligible=len(t0_files))
    counts["areas_exist_t1"] = sum(1 for k in t0_files if k in t1_files)
    counts["areas_ge5_changes"] = sum(1 for k in t0_files if len(area_chg["outcome"][k]) >= MIN_CHANGES)
    for k, fl in t0_files.items():
        if k not in t1_files or len(area_chg["outcome"][k]) < MIN_CHANGES:
            continue
        m0 = metrics([M0[p] for p in fl])
        m1 = metrics([M1[p] for p in t1_files[k]])
        ages = [(t0 - D["first_ts"][p]) / DAY if p in D["first_ts"] else 0.0 for p in fl]
        hot = sum(M0[p]["num"] * np.log1p(prior_cnt_file.get(p, 0)) for p in fl) / max(m0["sloc"], 1)
        pc = list(area_chg["prior"][k].values())
        oc = list(area_chg["outcome"][k].values())
        auth = Counter(c["an"] for c in pc)
        ob = outcome_block(oc, linked_prs("outcome", k))
        pb = outcome_block(pc, linked_prs("prior", k))
        row = dict(repo=repo, unit=k, level="area",
                   log_sloc=np.log1p(m0["sloc"]), log_prior_changes=np.log1p(len(pc)),
                   log_prior_authors=np.log1p(len(auth)), log_nfiles=np.log1p(m0["nfiles"]),
                   median_age=float(np.median(ages)),
                   erosion=m0["erosion"], log_n_high_cc=np.log1p(m0["nhigh"]), clone_pct=m0["clone"], hotspot=hot,
                   d_erosion=m0["erosion"] - m1["erosion"], d_log_n_high_cc=np.log1p(m0["nhigh"]) - np.log1p(m1["nhigh"]),
                   d_clone_pct=m0["clone"] - m1["clone"],
                   coupling=float(np.mean([c["coupled"] for c in pc])) if pc else 0.0,
                   ownership=max(auth.values()) / len(pc) if pc else 0.0)
        for o in ("o1", "o2", "o3", "o4a", "o4b", "o4c", "o5a", "o5b"):
            row["out_" + o] = ob[o]
            v = pb[o]
            # lagged outcome: no prior change -> 0 (documented); unmeasured PR outcomes stay NaN
            row["lag_" + o] = (0.0 if (pb["n_changes"] == 0 and o in ("o1", "o2", "o3", "o5a")) else v)
        row.update(n_changes=ob["n_changes"], n_prs=ob["n_prs"], n_prior_prs=pb["n_prs"], n_prior_changes=len(pc))
        rows.append(row)
    counts["areas_included"] = len(rows)
    counts["changes_in_included"] = int(sum(r["n_changes"] for r in rows))
    counts["prs_linked_outcome"] = len({i for k in t0_files if len(area_chg["outcome"][k]) >= MIN_CHANGES
                                        for h in area_chg["outcome"][k] for i in by_oid.get(h, set())})
    counts["prs_fetched"] = len(prs)
    counts["areas_o4_unmeasured"] = int(sum(1 for r in rows if np.isnan(r["out_o4a"])))
    counts["valid_changes_outcome"] = sum(1 for c in valid.values() if c["win"] == "outcome" and c["ents"])
    counts["bot_commits"] = D["meta"]["n_bot_commits"]

    # ---------------- file level (secondary)
    frows = []
    for p in M0:
        if p not in D["path_t1"]:
            continue
        p1 = D["path_t1"][p]
        if p1 not in M1:
            continue
        oc_raw = [x for x in D["fout"].get(p, []) if x[0] in valid]
        pc_raw = [x for x in D["fprior"].get(p, []) if x[0] in valid]
        if len({x[0] for x in oc_raw}) < MIN_CHANGES:
            continue
        own = id_of(p)

        def mk(lst, win):
            out = []
            for h, path, a, d in {x[0]: x for x in lst}.values():
                c = valid[h]
                ids = {id_of(e[2]) for e in c["ents"]}
                out.append(dict(churn=a + d, nel=len({e[2] for e in c["ents"] if e[0] != "R-old"}), an=c["an"],
                                coupled=bool(ids - {own}),
                                follow=int(follow_flags(streams[win], c["seq"], c["ts"], {path}))))
            return out
        oc, pc = mk(oc_raw, "outcome"), mk(pc_raw, "prior")
        ob, pb = outcome_block(oc, []), outcome_block(pc, [])
        m0, m1 = metrics([M0[p]]), metrics([M1[p1]])
        auth = Counter(c["an"] for c in pc)
        row = dict(repo=repo, unit=p, level="file",
                   log_sloc=np.log1p(m0["sloc"]), log_prior_changes=np.log1p(len(pc)),
                   log_prior_authors=np.log1p(len(auth)),
                   median_age=(t0 - D["first_ts"][p]) / DAY if p in D["first_ts"] else 0.0,
                   erosion=m0["erosion"], log_n_high_cc=np.log1p(m0["nhigh"]), clone_pct=m0["clone"],
                   hotspot=M0[p]["num"] * np.log1p(len(pc)) / max(m0["sloc"], 1),
                   d_erosion=m0["erosion"] - m1["erosion"], d_log_n_high_cc=np.log1p(m0["nhigh"]) - np.log1p(m1["nhigh"]),
                   d_clone_pct=m0["clone"] - m1["clone"],
                   coupling=float(np.mean([c["coupled"] for c in pc])) if pc else 0.0,
                   ownership=max(auth.values()) / len(pc) if pc else 0.0)
        for o in ("o1", "o2", "o3", "o5a"):
            row["out_" + o] = ob[o]
            row["lag_" + o] = 0.0 if not pc else pb[o]
        row.update(n_changes=ob["n_changes"], n_prior_changes=len(pc))
        frows.append(row)
    counts["files_included"] = len(frows)
    return rows, frows, counts


def main():
    arows, frows, counts = [], [], []
    for repo in NAMES:
        with gzip.open(CACHE / f"{repo}.json.gz", "rt") as fh:
            skipped = json.load(fh).get("skipped")
        if skipped:
            counts.append(dict(repo=repo, skipped="no T-1 snapshot (history starts after T0-365d)"))
            print(counts[-1], flush=True)
            continue
        a, f, c = run_repo(repo)
        arows += a; frows += f; counts.append(c)
        print(c, flush=True)
    for name, rows in (("area", arows), ("file", frows)):
        df = pd.DataFrame(rows)
        key = ["repo", "unit", "level", "n_changes"]
        outs = [c for c in df.columns if c.startswith("out_")]
        df[[c for c in df.columns if c not in outs]].to_csv(HERE / f"features-{name}.csv", index=False)
        df[key + outs].to_csv(HERE / f"outcomes-{name}.csv", index=False)
    (HERE / "dataset-counts.json").write_text(json.dumps(counts, indent=1))


if __name__ == "__main__":
    main()
