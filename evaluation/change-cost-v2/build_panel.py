#!/usr/bin/env python3
"""E1b-v2 stage 3: assemble the PR-level panel (outcomes, predictors, covariates) and the exclusion funnel.

Usage: uv run --with pandas,numpy python build_panel.py
Reads data/pr-<repo>.jsonl.gz and raw/prs-<repo>.json.gz; writes panel.csv.gz (eligible rows) and funnel.json.
Everything here is a deterministic definition from the preregistration; no modelling.
"""
import gzip
import json
import math
from collections import Counter, defaultdict

import numpy as np
import pandas as pd

from common import HERE, NAMES, load_pins

SESSION_GAP_H = 4.0     # git-hours session gap
FIRST_COMMIT_H = 0.5    # credit for the first commit of each session
AREA_DEPTH0, AREA_MAX_SHARE, AREA_DEPTH_MAX = 2, 0.15, 6
GATE_MIN_N, GATE_MIN_SHARE, GATE_MIN_AREAS, GATE_MAX_AREA_SHARE, GATE_MAX_FLOOR = 150, 0.10, 10, 0.50, 0.80
GATE_MIN_COVERAGE = 0.85

CLASSES = ["fix", "feature", "refactor", "other_dev", "nondev"]


def snapshot_coverage(f):
    """G1: share of PRs that needed a snapshot (ok + no_snapshot + no_preexisting_file + no_measurable_file; `nondev` is a subset of ok)
    for which one was found. `f` is a funnel entry. 0.0 when no PR reached the snapshot stage."""
    reach = sum(f.get(k, 0) for k in ("ok", "no_snapshot", "no_preexisting_file", "no_measurable_file"))
    return (1 - f.get("no_snapshot", 0) / reach) if reach else 0.0


def load_types():
    """(repo, number) -> task class: argmax of the rep-mean Jev class probabilities (ties -> other_dev); from types.jsonl."""
    acc = defaultdict(list)
    f = HERE / "types.jsonl"
    if f.exists():
        for l in f.read_text().splitlines():
            r = json.loads(l)
            if r.get("ok"):
                acc[(r["repo"], r["number"])].append(r["probs"])
    out = {}
    for k, ps in acc.items():
        m = {c: sum(p[c] for p in ps) / len(ps) for c in CLASSES}
        best = max(m.values())
        winners = [c for c in CLASSES if m[c] == best]
        out[k] = winners[0] if len(winners) == 1 else "other_dev"
    return out


def ts(s):
    return pd.Timestamp(s).timestamp()


def active_hours(authored, gap_h=SESSION_GAP_H):
    t = sorted(authored)
    h = FIRST_COMMIT_H
    for a, b in zip(t, t[1:]):
        gap = (b - a) / 3600.0
        h += gap if gap <= gap_h else FIRST_COMMIT_H
    return h


def area_of(path, d):
    parts = path.split("/")[:-1]
    return "/".join(parts[:d]) if parts else "."


def assign_areas(df):
    """df has repo, number, main_path. Adaptive depth: areas holding > 15% of a repo's PRs are deepened (<= depth 6)."""
    out = {}
    for repo, g in df.groupby("repo"):
        depth = {k: AREA_DEPTH0 for k in g.index}
        for _ in range(AREA_DEPTH_MAX - AREA_DEPTH0):
            keys = {k: area_of(g.at[k, "main_path"], depth[k]) for k in g.index}
            cnt = Counter(keys.values())
            big = {a for a, c in cnt.items() if c / len(g) > AREA_MAX_SHARE}
            if not big:
                break
            for k in g.index:
                if keys[k] in big:
                    depth[k] += 1
        for k in g.index:
            out[k] = repo + ":" + area_of(g.at[k, "main_path"], depth[k])
    return pd.Series(out)


def main():
    pins = load_pins()
    types = load_types()
    rows, funnel = [], {}
    for repo in NAMES:
        f = HERE / "data" / f"pr-{repo}.jsonl.gz"
        if not f.exists():
            continue
        data = [json.loads(l) for l in gzip.open(f, "rt")]
        raw = {p["number"]: p for p in json.loads(gzip.open(HERE / "raw" / f"prs-{repo}.json.gz", "rt").read())}
        funnel[repo] = dict(fetched=len(data), **Counter(r.get("skip", "ok") for r in data))
        t_end = pins[repo]["t_end_ts"]
        for r in data:
            if "skip" in r:
                continue
            p = raw[r["number"]]
            ttype = types.get((repo, r["number"]))
            if ttype is None or ttype == "nondev":
                key = "type_unlabelled" if ttype is None else "nondev"
                funnel[repo][key] = funnel[repo].get(key, 0) + 1
                continue
            authored = [ts(c["authored"]) for c in p["commits"]]
            merged, created = ts(p["merged"]), ts(p["created"])
            hours = active_hours(authored)
            revs = sorted((ts(x["at"]), x["state"], x["by"]) for x in p["reviews"] if x["at"] and x["by"] != p["author"])
            first_rev = revs[0][0] if revs else None
            fs = r["files"]
            nloc = sum(x["nloc"] for x in fs)
            hi = sum(x["hi"] for x in fs)
            st = [x for x in fs if x["stable"]]
            nS, hS = sum(x["nloc"] for x in st), sum(x["hi"] for x in st)
            nT, hT = sum(x["nloc_t"] for x in st), sum(x["hi_t"] for x in st)
            measured = {y["p"] for y in fs}
            main_path = max(((x["add"] + x["dele"], x["path"]) for x in p["files"] if x["path"] in measured))[1]
            dt_ = pd.Timestamp(p["merged"])
            rows.append(dict(
                repo=repo, number=r["number"], merged=p["merged"], created=p["created"], merged_ts=merged,
                quarter=f"{dt_.year}Q{(dt_.month - 1) // 3 + 1}", author=r["author"], task=ttype, main_path=main_path,
                maintainer=int(r["assoc"] in ("OWNER", "MEMBER", "COLLABORATOR")),
                log_churn=math.log1p(r["churn"]), log_elig_files=math.log(r["n_elig"]),
                frac_new=r["n_new"] / r["n_elig"], log_nloc=math.log(nloc),
                log_author_prior=math.log1p(r["author_prior_commits"]) if r["author_prior_commits"] is not None else np.nan,
                # predictors
                Q=hi / nloc, n_measured=r["n_measured"],
                D=(hS / nS - hT / nT) if nS > 0 else np.nan, n_stable=r["n_stable"], Q_stable=(hS / nS) if nS > 0 else np.nan,
                log_hidden=math.log1p(r["n_hidden"]), n_hidden=r["n_hidden"], n_partner_abs=r["n_partner_abs"],
                log_partner_abs=math.log1p(r["n_partner_abs"]),
                # outcomes
                active_hours=hours, y_active=math.log(hours), n_commits=len(authored),
                y_active_g2=math.log(active_hours(authored, 2.0)), y_active_g8=math.log(active_hours(authored, 8.0)),
                has_review=int(first_rev is not None),
                review_latency_h=((merged - first_rev) / 3600.0) if first_rev is not None else np.nan,
                author_wait_h=((first_rev - created) / 3600.0) if first_rev is not None else np.nan,
                n_reviews=len(revs), n_changes_requested=sum(1 for x in revs if x[1] == "CHANGES_REQUESTED"),
                n_rework_commits=(sum(1 for a in authored if a > first_rev) if first_rev is not None else np.nan),
                y_review=math.log1p((merged - first_rev) / 3600.0) if first_rev is not None and merged >= first_rev else np.nan,
                y_rework=math.log1p(sum(1 for a in authored if a > first_rev)) if first_rev is not None else np.nan,
                censored_followup=int(merged > t_end - 14 * 86400),
                neg_duration=int(max(authored) > merged + 1), S_src=r["S_src"]))
    df = pd.DataFrame(rows)
    df["area"] = assign_areas(df)
    df = df.sort_values(["repo", "number"]).reset_index(drop=True)
    # ---- repository gates (predictor and outcome MARGINALS only; PREREGISTRATION §2, §5)
    gate, keep = {}, []
    for repo, g in df.groupby("repo"):
        floor = float((g.y_active <= math.log(FIRST_COMMIT_H) + 1e-9).mean())
        d = g.D.dropna()
        gate[repo] = dict(n=len(g), n_areas=int(g.area.nunique()), max_area_share=float(g.groupby("area").size().max() / len(g)),
                          q_pos_share=float((g.Q > 0).mean()), d_measured=int(len(d)),
                          d_nonzero_share=float((d != 0).mean()) if len(d) else 0.0, y_active_floor_share=floor)
        gate[repo]["snapshot_coverage"] = snapshot_coverage(funnel[repo])
        gate[repo]["pass"] = bool(len(g) >= GATE_MIN_N and gate[repo]["q_pos_share"] >= GATE_MIN_SHARE and
                                  gate[repo]["d_nonzero_share"] >= GATE_MIN_SHARE and gate[repo]["n_areas"] >= GATE_MIN_AREAS and
                                  gate[repo]["max_area_share"] <= GATE_MAX_AREA_SHARE and
                                  gate[repo]["snapshot_coverage"] >= GATE_MIN_COVERAGE)   # G1: else the repository is excluded
        gate[repo]["y_active_informative"] = bool(floor <= GATE_MAX_FLOOR)   # else the repo drops out of T1-T4 only
        if gate[repo]["pass"]:
            keep.append(repo)
    (HERE / "repo-gate.json").write_text(json.dumps(gate, indent=1))
    df = df[df.repo.isin(keep)].copy()
    for repo in keep:
        if not gate[repo]["y_active_informative"]:
            m = df.repo == repo
            for c in ("y_active", "y_active_g2", "y_active_g8"):
                df.loc[m, c] = np.nan
    df.to_csv(HERE / "panel.csv.gz", index=False)
    (HERE / "funnel.json").write_text(json.dumps(funnel, indent=1))
    print(len(df), "panel rows;", df.repo.nunique(), "repos;", df.area.nunique(), "areas; failed gate:",
          sorted(r for r, v in gate.items() if not v["pass"]))


if __name__ == "__main__":
    main()
