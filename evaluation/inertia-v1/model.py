#!/usr/bin/env python3
"""E1 inertia-v1 models per PREREGISTRATION.md. Reads features.csv, writes model-results.json.

uv run --with numpy,pandas,scikit-learn python model.py
"""
import json
import warnings
from pathlib import Path

import numpy as np
import pandas as pd
from sklearn.linear_model import LogisticRegression, Ridge
from sklearn.metrics import roc_auc_score

warnings.filterwarnings("ignore")
HERE = Path(__file__).resolve().parent
SEED = 20261016
B = 2000
ALPHAS = np.logspace(-2, 5, 15)  # grid not fixed by prereg; extended after first run hit the 1e3 ceiling
REPO_ORDER = ["fastapi", "webpack", "nest", "caddy", "axum", "junit5", "eShopOnWeb"]

df = pd.read_csv(HERE / "features.csv")
df["Y"] = np.log1p(df.fix_commits)
df["Y_any"] = (df.fix_commits > 0).astype(int)
X = pd.DataFrame({
    "log_sloc": np.log1p(df.sloc), "log_prior_commits": np.log1p(df.prior_commits),
    "log_prior_fix": np.log1p(df.prior_fix), "log_prior_lines": np.log1p(df.prior_lines),
    "log_age_days": np.log1p(df.age_days), "log_prior_authors": np.log1p(df.prior_authors),
    "erosion": df.erosion, "cc_max": df.cc_max, "log_n_high_cc": np.log1p(df.n_high_cc),
    "clone_pct": df.clone_pct, "rework": df.rework,
}).astype(float)
BASE = list(X.columns[:6])
QUAL = list(X.columns[6:])
missing_repos = [r for r in REPO_ORDER if r in set(df.repo) and X.loc[df.repo == r, "clone_pct"].isna().all()]
# z-score within repo; constant or missing -> 0 (mean imputation after scaling)
Z = X.copy()
for r in df.repo.unique():
    m = (df.repo == r).values
    sub = X.loc[m]
    mu, sd = sub.mean(), sub.std(ddof=0)
    Z.loc[m] = ((sub - mu) / sd.replace(0, np.nan)).fillna(0.0).values
repos = df.repo.values
y = df.Y.values
ya = df.Y_any.values
present = [r for r in REPO_ORDER if r in set(repos)]


def best_alpha(cols, tr_idx, kind):
    """inner leave-one-repository-out CV over training repositories."""
    trr = sorted(set(repos[tr_idx]))
    scores = []
    for a in ALPHAS:
        tot, n = 0.0, 0
        for g in trr:
            va = tr_idx[repos[tr_idx] == g]
            fi = tr_idx[repos[tr_idx] != g]
            if kind == "ridge":
                m = Ridge(alpha=a).fit(Z.loc[fi, cols], y[fi])
                tot += ((m.predict(Z.loc[va, cols]) - y[va]) ** 2).sum()
            else:
                m = LogisticRegression(C=1 / a, max_iter=500).fit(Z.loc[fi, cols], ya[fi])
                p = np.clip(m.predict_proba(Z.loc[va, cols])[:, 1], 1e-6, 1 - 1e-6)
                tot += -(ya[va] * np.log(p) + (1 - ya[va]) * np.log(1 - p)).sum()
            n += len(va)
        scores.append(tot / n)
    return ALPHAS[int(np.argmin(scores))]


def loro(cols, kind):
    pred = np.zeros(len(df))
    alphas = {}
    for g in present:
        te = np.where(repos == g)[0]
        tr = np.where(repos != g)[0]
        a = best_alpha(cols, tr, kind)
        alphas[g] = float(a)
        if kind == "ridge":
            m = Ridge(alpha=a).fit(Z.loc[tr, cols], y[tr])
            pred[te] = m.predict(Z.loc[te, cols])
        else:
            m = LogisticRegression(C=1 / a, max_iter=500).fit(Z.loc[tr, cols], ya[tr])
            pred[te] = m.predict_proba(Z.loc[te, cols])[:, 1]
    return pred, alphas


rng = np.random.default_rng(SEED)
idx_by_repo = {g: np.where(repos == g)[0] for g in present}
boot_idx = []
for _ in range(B):
    boot_idx.append(np.concatenate([rng.choice(idx_by_repo[g], len(idx_by_repo[g]), replace=True) for g in present]))


def rel_red(pb, pq, idx):
    mb = ((pb[idx] - y[idx]) ** 2).mean()
    mq = ((pq[idx] - y[idx]) ** 2).mean()
    return (mb - mq) / mb


def summarize(name, cols, pb, pq):
    allidx = np.arange(len(df))
    pooled = rel_red(pb, pq, allidx)
    bs = np.array([rel_red(pb, pq, i) for i in boot_idx])
    per = {g: float(rel_red(pb, pq, idx_by_repo[g])) for g in present}
    return dict(name=name, predictors=cols, pooled_mse_base=float(((pb - y) ** 2).mean()),
                pooled_mse_model=float(((pq - y) ** 2).mean()), rel_reduction=float(pooled),
                ci95=[float(np.percentile(bs, 2.5)), float(np.percentile(bs, 97.5))],
                per_repo=per, repos_improved=int(sum(v > 0 for v in per.values())),
                support=bool(pooled >= 0.05 and np.percentile(bs, 2.5) > 0 and sum(v > 0 for v in per.values()) >= 5))


def boot_coef(cols, target, alpha):
    """Bootstrap (files within repo) coefficient of `target` in baseline+cols ridge."""
    cs = []
    for i in boot_idx[:B]:
        m = Ridge(alpha=alpha).fit(Z.iloc[i][cols], y[i])
        cs.append(m.coef_[cols.index(target)])
    cs = np.array(cs)
    m = Ridge(alpha=alpha).fit(Z[cols], y)
    return float(m.coef_[cols.index(target)]), [float(np.percentile(cs, 2.5)), float(np.percentile(cs, 97.5))]


out = {"n_files": int(len(df)), "counts": df.groupby("repo").size().to_dict(),
       "y_mean_by_repo": df.groupby("repo").Y.mean().to_dict(),
       "frac_any_by_repo": df.groupby("repo").Y_any.mean().to_dict(),
       "clone_pct_missing_repos": missing_repos}
pb, ab = loro(BASE, "ridge")
out["alphas_base"] = ab
pq, aq = loro(BASE + QUAL, "ridge")
out["alphas_quality"] = aq
out["main"] = summarize("baseline+quality", QUAL, pb, pq)
# direction of each quality predictor: expected sign if the predictor behaves as intended (higher = riskier)
EXPECT = {"erosion": -1, "cc_max": 1, "log_n_high_cc": 1, "clone_pct": 1, "rework": 1}
out["ablations"] = []
for q in QUAL:
    p1, a1 = loro(BASE + [q], "ridge")
    s = summarize(f"baseline+{q}", [q], pb, p1)
    alpha_med = float(np.median(list(a1.values())))
    coef, ci = boot_coef(BASE + [q], q, alpha_med)
    s["coef"], s["coef_ci95"], s["coef_alpha"] = coef, ci, alpha_med
    s["expected_sign"] = EXPECT[q]
    s["construct_invalid"] = bool(coef * EXPECT[q] < 0 and (ci[0] > 0 or ci[1] < 0))
    out["ablations"].append(s)
alpha_m = float(np.median(list(aq.values())))
full_cols = BASE + QUAL
coefs = {}
for q in QUAL:
    c, ci = boot_coef(full_cols, q, alpha_m)
    coefs[q] = dict(coef=c, ci95=ci)
out["main"]["coefs_in_full_model"] = coefs
# baseline-only coefficients (sanity)
m = Ridge(alpha=float(np.median(list(ab.values())))).fit(Z[BASE], y)
out["baseline_coefs"] = dict(zip(BASE, map(float, m.coef_)))

# secondary: Y_any AUROC
lb, lab = loro(BASE, "logit")
lq, laq = loro(BASE + QUAL, "logit")


def auc(p, idx):
    return roc_auc_score(ya[idx], p[idx]) if len(set(ya[idx])) == 2 else float("nan")


allidx = np.arange(len(df))
dbs = []
for i in boot_idx:
    if len(set(ya[i])) == 2:
        dbs.append(roc_auc_score(ya[i], lq[i]) - roc_auc_score(ya[i], lb[i]))
out["auroc"] = dict(
    pooled_base=float(auc(lb, allidx)), pooled_quality=float(auc(lq, allidx)),
    delta=float(auc(lq, allidx) - auc(lb, allidx)),
    delta_ci95=[float(np.percentile(dbs, 2.5)), float(np.percentile(dbs, 97.5))],
    per_repo={g: dict(base=float(auc(lb, idx_by_repo[g])), quality=float(auc(lq, idx_by_repo[g]))) for g in present},
    mean_per_repo_base=float(np.nanmean([auc(lb, idx_by_repo[g]) for g in present])),
    mean_per_repo_quality=float(np.nanmean([auc(lq, idx_by_repo[g]) for g in present])))
# descriptive: Spearman of each quality predictor with Y within repo (pooled mean)
from scipy.stats import spearmanr  # noqa: E402
desc = {}
for q in QUAL + ["log_sloc", "log_prior_fix"]:
    rs = []
    for g in present:
        i = idx_by_repo[g]
        if Z.loc[i, q].std() > 0 and y[i].std() > 0:
            rs.append(spearmanr(Z.loc[i, q], y[i])[0])
    desc[q] = dict(mean_within_repo_spearman=float(np.mean(rs)), n_repos=len(rs))
out["descriptive_spearman"] = desc
(HERE / "model-results.json").write_text(json.dumps(out, indent=1))
print(json.dumps({k: out[k] for k in ("n_files", "counts", "main")}, indent=1)[:3000])
