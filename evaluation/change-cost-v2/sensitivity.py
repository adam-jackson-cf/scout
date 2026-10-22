#!/usr/bin/env python3
"""E1b-v2 preregistered sensitivity analyses (PREREGISTRATION §8); reported, not decision-bearing.

Usage: uv run --with numpy,pandas,scipy python sensitivity.py panel-final.csv.gz sensitivity.json
(a) multi-commit PRs only; (b) session gap 2 h / 8 h; (c) leave-one-repository-out; (d) per-repository estimates;
(e) explicit-reference follow-up outcome for T5; (f) inference clustered by AUTHOR instead of area (same estimates).
Every variant refits the whole pipeline (jackknife SEs) via analysis.analyse.
"""
import json
import sys

import analysis as an  # first: pins BLAS threads
import pandas as pd

TESTS = ["T1", "T2", "T3", "T4", "T5"]


def brief(res):
    return {t: dict(est=res[t]["est"], se=res[t]["se"], ci95=res[t]["ci95"], p=res[t]["p"]) for t in TESTS if t in res} | {
        "n": res["n"], "G": res["G"]}


def safe(df, **kw):
    try:
        return brief(an.analyse(df, **kw))
    except Exception as e:   # tiny repositories can lack a test's data
        return dict(error=f"{type(e).__name__}: {e}"[:200])


def main():
    df = pd.read_csv(sys.argv[1])
    # (a) conditions on n_commits >= 2, a post-treatment variable: a descriptive robustness view, NOT a causal contrast
    out = dict(a_multicommit_post_treatment_conditioning=safe(df[df.n_commits >= 2]))
    for g in (2, 8):
        d = df.copy()
        d["y_active"] = d[f"y_active_g{g}"]
        out[f"b_session_gap_{g}h"] = safe(d)
    out["c_leave_one_repo_out"] = {r: safe(df[df.repo != r]) for r in sorted(df.repo.unique())}
    out["d_per_repo"] = {r: safe(df[df.repo == r]) for r in sorted(df.repo.unique())}
    old = pd.to_datetime(df.merged).dt.year < 2024   # mixed eras: windows ending 2023 vs 2025
    out["d_cohort_merged_before_2024"] = safe(df[old])
    out["d_cohort_merged_2024_or_later"] = safe(df[~old])
    d = df.copy()
    d["y_incomplete"] = d["y_incomplete_explicit"]
    out["e_explicit_followup"] = safe(d)
    out["f_author_clusters"] = safe(df, cluster="author")
    json.dump(out, open(sys.argv[2], "w"), indent=1)


if __name__ == "__main__":
    main()
