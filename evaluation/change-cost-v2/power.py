#!/usr/bin/env python3
"""E1b-v2 power / minimum-detectable-effect analysis by cluster-aware simulation of the FULL pipeline.

Usage: uv run --with numpy,pandas,scipy python power.py [panel-final.csv.gz] [sims] [out.json]
Design = the REAL panel's predictors, covariates, FE groups and area clusters (real X). Outcomes are SIMULATED from planted
effects with nuisance parameters taken from OUTCOME marginals only (within-repo SD and area ICC of the outcome; repo base
rates and area dispersion of the binary outcome): no predictor-outcome association of the real data is used or computed.
Covariates get zero outcome effect, so residual variance is overstated -> MDEs are conservative.
Step 1: sims with every planted effect = 0 give the pipeline's jackknife SE for each test. ANALYTIC MDE80 = (t_{1-a/2,G-1} + z_0.8) * SE
        with a = 0.01 (= 0.05/5, the strictest Holm threshold; the largest effect Holm can require). It is a normal approximation,
        not reported as an 80 %-power effect.
Step 2: sims with the test's effect planted at the analytic MDE80 verify empirical power (p < 0.01) and the null rejection rate.
Step 3: EMPIRICAL MDE80 = analytic MDE80 x the smallest multiplier in {1, 1.25, 1.5} whose simulated power (200 sims) is >= 0.80.
Units of the planted effect: T1, T2 log hours per within-repo SD; T3, T4 log-SD multiplier per within-repo SD (relative
|residual| slope is about equal); T5 log-odds per within-repo SD.
Design adequacy (used by analysis.decide): upper Monte Carlo bound of the EMPIRICAL MDE80 <= SESOI (prereg §7).
"""
import json
import multiprocessing as mp
import sys
from pathlib import Path

import analysis as an  # first: pins BLAS threads to 1 before numpy loads
import numpy as np
import pandas as pd
from scipy import stats as st

from constructs import icc, within_sd

TESTS = ["T1", "T2", "T3", "T4", "T5"]
ALPHA_SINGLE = 0.01
HERE = Path(__file__).resolve().parent


def simulate(df0, rng, betas, nuis):
    """betas: dict test -> planted effect (missing = 0)."""
    df = df0.copy()
    repo = pd.factorize(df.repo)[0]
    area = pd.factorize(df.area)[0]
    # plant on each test's OWN complete-case sample and scale (analysis.estimate standardises per test sample)
    def std_on(col, cols):
        ok = np.ones(len(df), bool)
        for c in cols:
            ok &= np.isfinite(df[c].to_numpy(float))
        x = np.zeros(len(df))
        x[ok] = an.within_repo_std(df[col].to_numpy(float)[ok], repo[ok])
        return x
    xq = std_on("Q", an.TEST_COLS["T1"][:-1] + ["y_active"])
    xd = std_on("D", an.TEST_COLS["T2"])
    xh = std_on("log_hidden", an.TEST_COLS["T5"])
    sd_tot, icc_a = nuis["sd"], nuis["icc"]
    su, se = sd_tot * np.sqrt(icc_a), sd_tot * np.sqrt(1 - icc_a)
    u = rng.normal(0, su, area.max() + 1)[area]
    eps = rng.normal(size=len(df))
    mean = u + betas.get("T1", 0.0) * xq + betas.get("T2", 0.0) * xd
    scale = se * np.exp(betas.get("T3", 0.0) * xq + betas.get("T4", 0.0) * xd)
    df["y_active"] = mean + scale * eps
    df.loc[df0.y_active.isna(), "y_active"] = np.nan   # same repositories/rows as the real analysis
    p0 = df.repo.map(nuis["rate_by_repo"]).fillna(nuis["rate"]).to_numpy(float)
    cl = rng.normal(0, nuis["cluster_logit_sd"], area.max() + 1)[area]
    eta = np.log(p0 / (1 - p0)) + cl + betas.get("T5", 0.0) * xh
    df["y_incomplete"] = rng.binomial(1, 1 / (1 + np.exp(-eta))).astype(float)
    if nuis["keep_missing"]:
        df.loc[df0.y_incomplete.isna(), "y_incomplete"] = np.nan
    df["y_review"] = np.nan
    df["y_rework"] = np.nan
    return df


def one(args):
    df0, betas, seed, nuis = args
    res = an.analyse(simulate(df0, np.random.default_rng(seed), betas, nuis))
    return {t: (res[t]["se"], res[t]["p"], res[t]["est"]) for t in TESTS}


MULTS = (1.0, 1.25, 1.5)   # analytic-MDE multipliers tried for the empirical MDE80 (preregistered grid)
EMP_SIMS = 200


def main():
    f = sys.argv[1] if len(sys.argv) > 1 else "panel-final.csv.gz"
    sims = int(sys.argv[2]) if len(sys.argv) > 2 else 40
    outp = sys.argv[3] if len(sys.argv) > 3 else "power-results.json"
    df0 = pd.read_csv(HERE / f)
    yi = df0.y_incomplete.dropna() if "y_incomplete" in df0 else pd.Series(dtype=float)
    rate = float(yi.mean()) if len(yi) else 0.05
    cl_sd = 0.3
    if len(yi) > 100:   # logit-scale cluster SD by method of moments on the OUTCOME alone (excess area variance), floor 0.15
        g = df0.dropna(subset=["y_incomplete"]).groupby("area").y_incomplete.agg(["mean", "count"])
        excess = np.average((g["mean"] - rate) ** 2 - rate * (1 - rate) / g["count"], weights=g["count"])
        cl_sd = float(max(0.15, np.sqrt(max(excess, 0)) / (rate * (1 - rate))))
    nuis = dict(sd=within_sd(df0, "y_active"), icc=icc(df0, "y_active", "area"), rate=rate,
                rate_by_repo=(df0.dropna(subset=["y_incomplete"]).groupby("repo").y_incomplete.mean().to_dict() if len(yi) else {}),
                cluster_logit_sd=cl_sd, keep_missing=bool(len(yi)))
    G = df0.area.nunique()
    t_crit = st.t.ppf(1 - ALPHA_SINGLE / 2, G - 1)
    ctx = mp.get_context("fork")
    out = dict(nuisance={k: v for k, v in nuis.items() if k != "rate_by_repo"}, G=G, n=len(df0), sims=sims, tests={})
    with ctx.Pool(10) as pool:
        null = pool.map(one, [(df0, {}, 5000 + i, nuis) for i in range(sims)])
    for t in TESTS:
        ses = np.array([r[t][0] for r in null])
        se = float(ses.mean())
        se_ub = float(se + 1.96 * ses.std(ddof=1) / np.sqrt(len(ses)))   # upper MC bound of the null SE
        mde = float((t_crit + st.norm.ppf(0.8)) * se)                    # on the ESTIMATE scale
        mde_ub = float((t_crit + st.norm.ppf(0.8)) * se_ub)
        # planted-parameter scale differs from the estimate scale (T3/T4: log-SD multiplier vs relative |e| slope): calibrate k = est/planted
        with ctx.Pool(10) as pool:
            cal = pool.map(one, [(df0, {t: mde}, 7000 + i, nuis) for i in range(sims)])
        k = float(np.mean([r[t][2] for r in cal]) / mde)
        with ctx.Pool(10) as pool:
            alt = pool.map(one, [(df0, {t: mde / k}, 9000 + i, nuis) for i in range(sims)])
        pw = float(np.mean([r[t][1] < ALPHA_SINGLE for r in alt]))
        half = 1.96 * float(np.sqrt(pw * (1 - pw) / len(alt)))
        # EMPIRICAL MDE80: plant analytic MDE x m for m in MULTS (EMP_SIMS sims each), smallest m with simulated power >= 0.80
        emp = {}
        mult = None
        for m in MULTS:
            with ctx.Pool(10) as pool:
                rr = pool.map(one, [(df0, {t: m * mde / k}, 11000 + i + int(1000 * m), nuis) for i in range(EMP_SIMS)])
            p_m = float(np.mean([r[t][1] < ALPHA_SINGLE for r in rr]))
            emp[str(m)] = dict(power=p_m, halfwidth95=1.96 * float(np.sqrt(p_m * (1 - p_m) / EMP_SIMS)), sims=EMP_SIMS)
            if p_m >= 0.80:
                mult = m
                break
        mde_emp = float(mult * mde) if mult else None
        mde_emp_ub = float(mult * mde_ub) if mult else None
        out["tests"][t] = dict(mean_se_null=se, se_null_mc_upper=se_ub, analytic_mde80=mde, analytic_mde80_mc_upper=mde_ub,
                               empirical_mde80_multiplier=mult, empirical_mde80=mde_emp, empirical_mde80_mc_upper=mde_emp_ub,
                               empirical_grid=emp, sesoi=an.SESOI[t],
                               adequate=bool(mde_emp_ub is not None and mde_emp_ub <= an.SESOI[t]),   # preregistered: empirical MDE80 (upper MC) <= SESOI
                               est_per_planted_k=k, planted_for_verification=float(mde / k),
                               null_reject_p01=float(np.mean([r[t][1] < ALPHA_SINGLE for r in null])),
                               power_at_analytic_mde_p01=pw, power_mc_halfwidth95=half,
                               mean_est_at_verification=float(np.mean([r[t][2] for r in alt])))
        print(t, out["tests"][t], flush=True)
    json.dump(out, open(HERE / outp, "w"), indent=1)


if __name__ == "__main__":
    main()
