#!/usr/bin/env python3
"""E1b-v2 construct checks: DISTRIBUTIONS ONLY (no predictor-outcome association is computed anywhere here).

Usage: uv run --with pandas,numpy python constructs.py [panel.csv.gz|panel-final.csv.gz] > constructs.json
Reports, per repository and overall: funnel counts, outcome distributions (quantiles, floor mass, negative durations),
predictor distributions (zero share, within-repo SD, quantiles), cluster structure, outcome variance decomposition
by repo / by area (marginal ICC of the OUTCOME alone; used by power.py), task-type mix, snapshot provenance.
"""
import json
import sys

import numpy as np
import pandas as pd

from common import HERE

Q = [0.05, 0.25, 0.5, 0.75, 0.95]


def qs(x):
    x = pd.Series(x).dropna()
    return {str(q): round(float(x.quantile(q)), 4) for q in Q} if len(x) else {}


def icc(df, col, grp):
    """One-way ANOVA ICC of `col` by `grp` after removing repo means (outcome marginal only)."""
    g = df[[col, grp, "repo"]].dropna()
    y = g[col] - g.groupby("repo")[col].transform("mean")
    k = g[grp].nunique()
    n = len(g)
    ms_b = (g.assign(y=y).groupby(grp).y.agg(["sum", "count"]).eval("sum**2/count").sum()) / max(k - 1, 1)
    ms_w = ((y - y.groupby(g[grp]).transform("mean")) ** 2).sum() / max(n - k, 1)
    n0 = (n - (g.groupby(grp).size() ** 2).sum() / n) / max(k - 1, 1)
    return float(max(0.0, (ms_b - ms_w) / (ms_b + (n0 - 1) * ms_w))) if ms_b + (n0 - 1) * ms_w > 0 else 0.0


def within_sd(df, col):
    g = df[[col, "repo"]].dropna()
    r = g[col] - g.groupby("repo")[col].transform("mean")
    return float(np.sqrt((r ** 2).sum() / max(len(g) - g.repo.nunique(), 1)))


def main():
    f = sys.argv[1] if len(sys.argv) > 1 else "panel.csv.gz"
    df = pd.read_csv(HERE / f)
    funnel = json.loads((HERE / "funnel.json").read_text())
    out = dict(funnel=funnel, n_panel=len(df), repos={}, overall={})
    floor = np.log(0.5)
    for repo, g in list(df.groupby("repo")) + [("ALL", df)]:
        d = dict(n=len(g), n_areas=int(g.area.nunique()), median_area_size=float(g.groupby("area").size().median()),
                 max_area_share=float(g.groupby("area").size().max() / len(g)),
                 task=g.task.value_counts().to_dict(),
                 y_active_q=qs(g.y_active), y_active_n=int(g.y_active.notna().sum()),
                 y_active_floor_share=float((g.y_active.dropna() <= floor + 1e-9).mean()) if g.y_active.notna().any() else None,
                 active_hours_q=qs(g.active_hours), single_commit_share=float((g.n_commits == 1).mean()),
                 neg_duration_share=float(g.neg_duration.mean()), reviewed_share=float(g.has_review.mean()),
                 review_latency_h_q=qs(g.review_latency_h), n_rework_commits_q=qs(g.n_rework_commits),
                 Q_zero_share=float((g.Q == 0).mean()), Q_q=qs(g.Q), Q_within_sd=within_sd(g, "Q"),
                 D_measured=int(g.D.notna().sum()), D_zero_share=float((g.D.dropna() == 0).mean()) if g.D.notna().any() else None,
                 D_q=qs(g.D), D_within_sd=within_sd(g, "D") if g.D.notna().sum() > 2 else None,
                 hidden_zero_share=float((g.n_hidden == 0).mean()), n_hidden_q=qs(g.n_hidden),
                 partner_abs_nonzero_share=float((g.n_partner_abs > 0).mean()),
                 log_author_prior_missing=int(g.log_author_prior.isna().sum()),
                 S_src=g.S_src.value_counts().to_dict(), frac_new_q=qs(g.frac_new),
                 censored_followup_share=float(g.censored_followup.mean()))
        if "y_incomplete" in g:
            yi = g.y_incomplete.dropna()
            d["y_incomplete_labelled"] = int(len(yi))
            d["y_incomplete_rate"] = float(yi.mean()) if len(yi) else None
            if "y_incomplete_explicit" in g:
                d["y_incomplete_explicit_rate"] = float(g.y_incomplete_explicit.dropna().mean())
        (out["overall"] if repo == "ALL" else out["repos"])[repo] = d
    out["variance"] = dict(
        y_active_within_repo_sd=within_sd(df, "y_active"), y_active_area_icc=icc(df, "y_active", "area"),
        y_review_within_repo_sd=within_sd(df, "y_review") if "y_review" in df else None,
        Q_area_icc=icc(df, "Q", "area"), D_area_icc=icc(df.dropna(subset=["D"]), "D", "area"),
        log_hidden_area_icc=icc(df, "log_hidden", "area"))
    if "y_incomplete" in df:
        yi = df.dropna(subset=["y_incomplete"])
        out["variance"]["y_incomplete_area_icc"] = icc(yi, "y_incomplete", "area") if len(yi) > 100 else None
    print(json.dumps(out, indent=1, default=lambda o: None))


if __name__ == "__main__":
    main()
