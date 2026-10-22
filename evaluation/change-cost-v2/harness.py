#!/usr/bin/env python3
"""E1b-v2 synthetic validation harness: planted-effect recovery and planted-null checks for the FULL pipeline
(analysis.analyse + analysis.decide). Synthetic data only; mimics the panel's structure (repo mean shifts, area
clusters with random effects, zero-inflated predictors, covariate confounding, floor/heavy-tailed outcome noise).

Usage: uv run --python 3.13 --with-requirements requirements.txt python harness.py [sims] [out.json]
Scenarios: recover | null | null_confounded | null_area_confounded | permutation | (reference big-sample truth)
"""
import json
import math
import multiprocessing as mp
import sys
import time

import analysis as an  # first: pins BLAS threads to 1 before numpy loads
import numpy as np
import pandas as pd

TRUTH_KEYS = ["T1", "T2", "T3", "T4", "T5"]


def synth(rng, beta1=0.0, beta2=0.0, theta1=0.0, theta2=0.0, beta5=0.0, cov_conf=0.0, area_conf=0.0, n_repos=8,
          size=200, sigma=1.0, floor_q=None):
    rows = []
    for r in range(n_repos):
        n = int(max(150, rng.lognormal(math.log(size), 0.6)))
        g = max(8, int(n / 25))
        w = rng.lognormal(0, 1.0, g)
        area = rng.choice(g, size=n, p=w / w.sum())
        za = rng.normal(size=g)               # area latent (Q level)
        ua = rng.normal(0, 0.5, g) + area_conf * za   # area difficulty (unmeasured), correlated with Q level if area_conf
        da = rng.normal(0, 0.3, g)             # area degradation trend
        ha = rng.normal(size=g)               # area hidden-coupling propensity
        task = rng.choice(4, size=n, p=[0.3, 0.3, 0.15, 0.25])
        era = rng.integers(0, 12)
        q = rng.integers(0, 8, size=n) + era
        log_churn = rng.normal(4 + 0.4 * task, 1.2, n)
        log_files = np.log1p(rng.poisson(1.5 + 0.5 * (task == 1), n))
        frac_new = rng.beta(1, 4, n)
        log_auth = np.clip(rng.normal(3, 1.5, n), 0, None)
        maint = rng.binomial(1, 0.5, n)
        log_nloc = rng.normal(6, 1.0, n)
        scale = rng.lognormal(0, 0.5)
        shift = rng.normal(0, 0.4)
        Qraw = np.maximum(0, 0.7 * za[area] + 0.7 * rng.normal(size=n) - 0.3 + shift + cov_conf * 0.3 * (log_churn - 4)) * scale
        Qs = Qraw + 0.05 * rng.normal(size=n) * (Qraw > 0)
        Draw = np.where(rng.random(n) < 0.4, 0.0, 0.5 * da[area] + 0.5 * rng.normal(size=n)) * 0.02
        lam = np.exp(0.3 + 0.5 * ha[area] + 0.3 * rng.normal(size=n) + cov_conf * 0.1 * (log_files - 0.9))
        hidden = rng.poisson(lam)
        explained = rng.poisson(0.5 + 0.3 * np.abs(ha[area]))
        rows.append(dict(
            repo=np.full(n, f"r{r}"), area=np.array([f"r{r}:a{a}" for a in area]), quarter=q.astype(str), task=task.astype(str),
            log_churn=log_churn, log_elig_files=log_files, frac_new=frac_new, log_author_prior=log_auth, maintainer=maint,
            log_nloc=log_nloc, Q=Qraw, Q_stable=Qs, D=Draw, log_hidden=np.log1p(hidden), log_partner_abs=np.log1p(hidden + explained),
            merged_ts=q * 90 * 86400.0 + rng.random(n) * 8e6,
            _ua=ua[area], _repo_int=rng.normal(0, 1.0), _n=n))
    df = pd.concat([pd.DataFrame({k: v for k, v in r.items() if not k.startswith("_") or k in ("_ua",)}).assign(_ri=r["_repo_int"])
                    for r in rows], ignore_index=True)
    xq, xd, xh = (an.within_repo_std(df[c].to_numpy(float), pd.factorize(df.repo)[0]) for c in ("Q", "D", "log_hidden"))
    tt = df.task.astype(int).to_numpy()
    eps = rng.normal(size=len(df))
    mean = (df._ri + 0.05 * (df.quarter.astype(int) % 5) + 0.2 * tt + 0.35 * (df.log_churn - 4) + 0.2 * df.log_elig_files
            + 0.1 * df.log_nloc + df._ua + beta1 * xq + beta2 * xd)
    sd = sigma * np.exp(theta1 * xq + theta2 * xd)
    y = mean + sd * eps
    df["y_active"] = np.maximum(y, np.quantile(y, floor_q)) if floor_q else y   # mass point like single-commit PRs
    cl = rng.normal(0, 0.3, df.area.nunique())
    cmap = dict(zip(df.area.unique(), cl))
    xp = an.within_repo_std(df.log_partner_abs.to_numpy(float), pd.factorize(df.repo)[0])
    eta = (-2.3 + 0.1 * (pd.factorize(df.repo)[0] % 3) + beta5 * xh + 0.25 * xq + 0.15 * (df.log_churn - 4) + df.area.map(cmap)
           + cov_conf * 0.4 * xp)   # follow-up opportunity grows with the number of absent partners (the T5 control)
    df["y_incomplete"] = rng.binomial(1, 1 / (1 + np.exp(-eta))).astype(float)
    df.loc[rng.random(len(df)) < 0.03, "y_incomplete"] = np.nan
    rev = rng.random(len(df)) < 0.6
    df["y_review"] = np.where(rev, rng.normal(2, 1.5, len(df)), np.nan)
    df["y_rework"] = np.where(rev, np.log1p(rng.poisson(1.0, len(df))), np.nan)
    return df.drop(columns=["_ua", "_ri"])


SCEN = {
    "recover": dict(beta1=0.15, beta2=0.15, theta1=0.30, theta2=0.30, beta5=0.40),
    "recover_floor": dict(beta1=0.15, beta2=0.15, theta1=0.30, theta2=0.30, beta5=0.40, floor_q=0.45),
    "null": dict(),
    "null_floor": dict(floor_q=0.45),
    "null_confounded": dict(cov_conf=1.0),
    "null_area_confounded": dict(area_conf=0.6),
}


def one(args):
    scen, seed, perm = args
    rng = np.random.default_rng(seed)
    kw = SCEN["recover" if perm else scen]
    df = synth(rng, **kw)
    if perm:  # permute (Q, Q_stable, D, log_hidden) rows jointly within repo x task strata by a random cyclic shift of
        # 1..m-1 positions (a derangement: no row keeps its own predictor, so no planted signal survives via fixed points)
        key = (df.repo + df.task).to_numpy()
        for k in np.unique(key):
            idx = np.nonzero(key == k)[0]
            idx = idx[rng.permutation(len(idx))]
            shift = int(rng.integers(1, len(idx)))
            src = np.roll(idx, shift)
            for c in ("Q", "Q_stable", "D", "log_hidden"):
                df.loc[idx, c] = df.loc[src, c].to_numpy()
    res = an.analyse(df)
    ver, hyp = an.decide(res, adequate=dict.fromkeys(TRUTH_KEYS, True), t5_decided=True)
    return dict(est={t: res[t]["est"] for t in TRUTH_KEYS}, se={t: res[t]["se"] for t in TRUTH_KEYS},
                ci95={t: res[t]["ci95"] for t in TRUTH_KEYS}, p={t: res[t]["p"] for t in TRUTH_KEYS},
                adj={t: ver[t]["adj_p"] for t in TRUTH_KEYS}, verdict={t: ver[t]["verdict"] for t in TRUTH_KEYS},
                P1=dict(gain=res.get("P1_rel_gain", {}).get("est"), base_vs_const=res.get("P1_base_vs_const", {}).get("est")),
                n=res["n"])


def truth(scen, perm, seed=777, n_repos=8, size=2000, n_ref=8):
    """Reference estimand: mean of point estimates over `n_ref` large independent populations (~13k PRs each) from the same DGP
    (repository-level properties are redrawn per population, so a single reference draw is itself noisy). Zero for permutations."""
    if perm:
        return {t: 0.0 for t in TRUTH_KEYS}
    kw = SCEN[scen]
    vals = []
    for k in range(n_ref):
        df = synth(np.random.default_rng(seed + k), n_repos=n_repos, size=size, **kw)
        A = an.encode(df)
        pt = an.estimate(A, np.arange(len(df)), A["area"])
        vals.append([pt[t] for t in TRUTH_KEYS])
    return dict(zip(TRUTH_KEYS, np.mean(vals, axis=0).astype(float)))


PLANTED = {"T1": "beta1", "T2": "beta2", "T5": "beta5"}
NULL_SCEN = ("null", "null_floor", "null_confounded", "permutation")


def summarise(runs, tr, name, planted_kw):
    out = {}
    for t in TRUTH_KEYS:
        est = np.array([r["est"][t] for r in runs])
        se = np.array([r["se"][t] for r in runs])
        cov = np.mean([r["ci95"][t][0] <= tr[t] <= r["ci95"][t][1] for r in runs])
        out[t] = dict(truth=tr[t], mean_est=float(est.mean()), bias=float(est.mean() - tr[t]), sd_est=float(est.std(ddof=1)),
                      mean_se=float(se.mean()), coverage95=float(cov),
                      reject_unadj=float(np.mean([r["p"][t] < 0.05 for r in runs])),
                      reject_holm=float(np.mean([r["adj"][t] < 0.05 for r in runs])),
                      supported=float(np.mean([r["verdict"][t] == "supported" for r in runs])))
        if t in PLANTED and name not in NULL_SCEN and name != "permutation":
            out[t]["planted"] = planted_kw.get(PLANTED[t], 0.0)
            out[t]["bias_vs_planted"] = float(est.mean() - out[t]["planted"])
    out["FWER_any_holm_reject"] = float(np.mean([any(r["adj"][t] < 0.05 for t in TRUTH_KEYS) for r in runs]))
    g = [r["P1"]["gain"] for r in runs if r["P1"]["gain"] is not None]
    bc = [r["P1"]["base_vs_const"] for r in runs if r["P1"]["base_vs_const"] is not None]
    out["P1"] = dict(mean_gain=float(np.mean(g)), mean_base_vs_const=float(np.mean(bc)))
    out["n_sims"] = len(runs)
    out["checks"] = checks(name, out)
    return out


def checks(name, o):
    """Preregistered G6 tolerances (PREREGISTRATION section 5). Returns {check: bool}; the floored scenario is reported only."""
    c = {}
    if name in NULL_SCEN:
        c["any_holm_reject<=0.10"] = o["FWER_any_holm_reject"] <= 0.10
        c["coverage95>=0.90_all_tests"] = all(o[t]["coverage95"] >= 0.90 for t in TRUTH_KEYS)
    elif name == "recover":
        c["|bias_vs_planted|<=0.05_T1_T2"] = all(abs(o[t]["bias_vs_planted"]) <= 0.05 for t in ("T1", "T2"))
        c["|bias_vs_planted|<=0.10_T5"] = abs(o["T5"]["bias_vs_planted"]) <= 0.10
        c["coverage95>=0.88_all_tests"] = all(o[t]["coverage95"] >= 0.88 for t in TRUTH_KEYS)
        c["detect_T2_T4>=0.80"] = o["T2"]["reject_holm"] >= 0.80 and o["T4"]["reject_holm"] >= 0.80
    elif name == "null_area_confounded":
        c["T2_T3_T4_T5_reject_unadj<=0.10"] = all(o[t]["reject_unadj"] <= 0.10 for t in ("T2", "T3", "T4", "T5"))
        c["T1_rejects>=0.50 (documented confounding)"] = o["T1"]["reject_unadj"] >= 0.50
    c["P1_base_beats_constant"] = o["P1"]["mean_base_vs_const"] > 0
    return c


def main():
    sims = int(sys.argv[1]) if len(sys.argv) > 1 else 60
    outp = sys.argv[2] if len(sys.argv) > 2 else "harness-results.json"
    results, t0 = {}, time.time()
    ctx = mp.get_context("fork")
    for name, scen, perm in (("recover", "recover", False), ("recover_floor", "recover_floor", False),
                             ("null", "null", False), ("null_floor", "null_floor", False),
                             ("null_confounded", "null_confounded", False),
                             ("null_area_confounded", "null_area_confounded", False), ("permutation", "recover", True)):
        tr = truth(scen, perm)
        with ctx.Pool(10) as pool:
            runs = pool.map(one, [(scen, 1000 + i + (500 if perm else 0), perm) for i in range(sims)])
        results[name] = summarise(runs, tr, name, SCEN["recover" if perm else scen])
        print(name, f"{time.time() - t0:.0f}s", flush=True)
    results["settings"] = dict(sims=sims, scenarios=SCEN)
    json.dump(results, open(outp, "w"), indent=1)


if __name__ == "__main__":
    main()
