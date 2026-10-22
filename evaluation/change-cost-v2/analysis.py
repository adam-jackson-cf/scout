#!/usr/bin/env python3
"""E1b-v2 analysis pipeline (the ONLY code that touches predictor-outcome relationships).

Imported by harness.py (synthetic validation), power.py (design power) and run when the study is executed:
    uv run --python 3.13 --with-requirements requirements.txt python analysis.py panel-final.csv.gz results.json --power power-results.json --validation validation-gates.json --gates gates.json

Estimands (all coefficients are per +1 within-repository SD of the predictor):
  T1 H1-level          y_active ~ Q_std      + controls + FE(repo, quarter, task)            OLS
  T2 H1-degradation    y_active ~ D_std      + controls + Q_stable_std + FE(area, quarter, task)   OLS (area absorbed)
  T3 H2-level          u ~ Q_std   ; u = |e| / repo-mean|e|, e = residual of the T1 model       OLS, same FE as T1
  T4 H2-degradation    u ~ D_std   ; e from the T2 model                                       OLS, same FE as T2
  T5 H4-hidden-coupling logit(y_incomplete) ~ logH_std + log1p(absent partners)_std + Q_std + controls + FE(repo, quarter, task)
Secondary (not in the decision family): S1 y_review (log review latency, reviewed PRs), S2 rework commits,
P1 temporal-holdout incremental predictive gain.
Uncertainty: delete-one-area-cluster jackknife that REFITS the whole pipeline (standardisation, stage-1/stage-2, logit) on every
leave-one-cluster-out sample. SE = jackknife SE with per-test cluster count G_t; p and CIs from t with G_t-1 df.
Decisions need power-results.json (adequacy) and validation-gates.json (G4/G5/G7) -- no human-set flags.
The logit ridge (1e-2) is a numerical stabiliser against separation, not shrinkage of the effect.
"""
import json
import os
import sys

for _v in ("OMP_NUM_THREADS", "OPENBLAS_NUM_THREADS", "VECLIB_MAXIMUM_THREADS", "MKL_NUM_THREADS"):
    os.environ.setdefault(_v, "1")   # jackknife replicates are small; BLAS threads only oversubscribe worker pools

import numpy as np
import pandas as pd

CONTROLS = ["log_churn", "log_elig_files", "frac_new", "log_author_prior", "maintainer", "log_nloc"]
SESOI = dict(T1=0.10, T2=0.10, T3=0.10, T4=0.10, T5=0.20)   # per within-repo SD: log hours, relative |e|, log-odds
FAMILY = ["T1", "T2", "T3", "T4", "T5"]
ALPHA = 0.05
MIN_TEST_N = 1   # a test whose complete-case sample is empty is skipped (exploratory per-repository runs only; the decision run requires all tests, see main)


# ------------------------------------------------------------------ data preparation
def encode(df):
    """-> dict of numpy arrays (numeric columns + integer codes for factors)."""
    A = {c: df[c].to_numpy(float) for c in CONTROLS + ["Q", "D", "Q_stable", "log_hidden", "y_active", "y_incomplete",
                                                          "y_review", "y_rework", "log_partner_abs"] if c in df}
    for c in ["repo", "quarter", "task", "area"] + (["author"] if "author" in df else []):
        A[c] = pd.factorize(df[c].fillna("NA"))[0]
    A["merged_ts"] = df["merged_ts"].to_numpy(float)
    return A


def dummies(codes):
    k = codes.max() + 1
    if k <= 1:
        return np.zeros((len(codes), 0))
    M = np.zeros((len(codes), k - 1))
    m = codes > 0
    M[np.nonzero(m)[0], codes[m] - 1] = 1.0
    return M


def relabel(codes):
    return pd.factorize(codes)[0]


def within_repo_std(x, repo):
    """(x - repo mean) / pooled within-repo SD."""
    r = relabel(repo)
    cnt = np.bincount(r)
    mean = np.bincount(r, weights=x) / cnt
    xc = x - mean[r]
    sd = np.sqrt((xc ** 2).sum() / max(len(x) - len(cnt), 1))
    return xc / sd if sd > 0 else xc * 0.0


def demean(M, groups):
    g = relabel(groups)
    cnt = np.bincount(g)
    if M.ndim == 1:
        return M - (np.bincount(g, weights=M) / cnt)[g]
    out = np.empty_like(M)
    for j in range(M.shape[1]):
        out[:, j] = M[:, j] - (np.bincount(g, weights=M[:, j]) / cnt)[g]
    return out


# ------------------------------------------------------------------ estimators
def ols(y, X, fe, absorb=None):
    """X columns: target first. fe: list of code arrays dummy-encoded. absorb: code array demeaned out (else intercept)."""
    Z = np.column_stack([X] + [dummies(relabel(f)) for f in fe])
    if absorb is not None:
        Z, y = demean(Z, absorb), demean(y, absorb)
    else:
        Z = np.column_stack([np.ones(len(y)), Z])
    b, *_ = np.linalg.lstsq(Z, y, rcond=None)
    off = 0 if absorb is not None else 1
    return b[off], y - Z @ b, Z, y


def logit(y, X, fe, lam=1e-2, iters=60):
    """Ridge-stabilised IRLS (lam on every coefficient except the intercept; negligible at n in the thousands, but prevents
    separation blow-ups in jackknife replicates). Steps are capped at 2 per iteration; non-convergence raises."""
    Z = np.column_stack([np.ones(len(y)), X] + [dummies(relabel(f)) for f in fe])
    pen = np.full(Z.shape[1], lam)
    pen[0] = 0.0
    b = np.zeros(Z.shape[1])
    b[0] = np.log(max(y.mean(), 1e-3) / max(1 - y.mean(), 1e-3))
    for _ in range(iters):
        eta = np.clip(Z @ b, -30, 30)
        p = 1 / (1 + np.exp(-eta))
        w = np.maximum(p * (1 - p), 1e-9)
        H = Z.T @ (Z * w[:, None]) + np.diag(pen)
        g = Z.T @ (y - p) - pen * b
        step = np.clip(np.linalg.solve(H, g), -2.0, 2.0)
        b += step
        if np.abs(step).max() < 1e-8:
            return b[1]
    raise RuntimeError(f"logit IRLS did not converge in {iters} iterations (max |step| {np.abs(step).max():.3g})")


def rel_abs_resid(e, repo):
    r = relabel(repo)
    a = np.abs(e)
    m = np.bincount(r, weights=a) / np.bincount(r)
    return a / np.maximum(m[r], 1e-12)


def _sel(A, ix, mask_cols):
    ok = np.ones(len(ix), bool)
    for c in mask_cols:
        ok &= np.isfinite(A[c][ix])
    return ix[ok]


def estimate(A, ix, cl):
    """Point estimates for every statistic on rows `ix` (indices into A) with cluster ids `cl` (aligned with A rows)."""
    out = {}
    ctrl = lambda ii: np.column_stack([A[c][ii] for c in CONTROLS])
    # --- T1 / T3
    ii = _sel(A, ix, CONTROLS + ["Q", "y_active"])
    out["n_T1"] = len(ii)
    if len(ii) >= MIN_TEST_N:   # a test whose own sample is (nearly) empty is skipped, e.g. django has no y_active
        x = within_repo_std(A["Q"][ii], A["repo"][ii])
        b, e, Z, _ = ols(A["y_active"][ii], np.column_stack([x, ctrl(ii)]), [A["repo"][ii], A["quarter"][ii], A["task"][ii]])
        out["T1"] = b
        u = rel_abs_resid(e, A["repo"][ii])
        out["T3"] = ols(u, np.column_stack([x, ctrl(ii)]), [A["repo"][ii], A["quarter"][ii], A["task"][ii]])[0]
    # --- T2 / T4 (area absorbed)
    ii = _sel(A, ix, CONTROLS + ["D", "Q_stable", "y_active"])
    out["n_T2"] = len(ii)
    if len(ii) >= MIN_TEST_N:
        d = within_repo_std(A["D"][ii], A["repo"][ii])
        q = within_repo_std(A["Q_stable"][ii], A["repo"][ii])
        X = np.column_stack([d, q, ctrl(ii)])
        b, e, _, _ = ols(A["y_active"][ii], X, [A["quarter"][ii], A["task"][ii]], absorb=cl[ii])
        out["T2"] = b
        u = rel_abs_resid(e, A["repo"][ii])
        out["T4"] = ols(u, X, [A["quarter"][ii], A["task"][ii]], absorb=cl[ii])[0]
    # --- T5
    ii = _sel(A, ix, CONTROLS + ["log_hidden", "log_partner_abs", "Q", "y_incomplete"])
    out["n_T5"] = len(ii)
    if len(ii) >= MIN_TEST_N:
        h = within_repo_std(A["log_hidden"][ii], A["repo"][ii])
        pa = within_repo_std(A["log_partner_abs"][ii], A["repo"][ii])   # total absent partners: removes follow-up OPPORTUNITY
        q = within_repo_std(A["Q"][ii], A["repo"][ii])
        out["T5"] = logit(A["y_incomplete"][ii], np.column_stack([h, pa, q, ctrl(ii)]),
                          [A["repo"][ii], A["quarter"][ii], A["task"][ii]])
    # --- secondary
    for name, ycol in (("S1", "y_review"), ("S2", "y_rework")):
        if ycol in A:
            ii = _sel(A, ix, EXPLORATORY_COLS[name])
            if len(ii) > 50:
                x = within_repo_std(A["Q"][ii], A["repo"][ii])
                out[name] = ols(A[ycol][ii], np.column_stack([x, ctrl(ii)]),
                                [A["repo"][ii], A["quarter"][ii], A["task"][ii]])[0]
    out.update(predictive_gain(A, ix))
    return out


def predictive_gain(A, ix):
    """P1: per repo, first 60% of merge times = train, last 40% = test. Models: constant(repo mean), baseline
    (controls + repo/task FE), +Q +D (needs D). Returns MSEs and relative reductions. Refit on every call."""
    ii = _sel(A, ix, EXPLORATORY_COLS["P1"])
    if len(ii) < 200:
        return {}
    ts = A["merged_ts"][ii]
    train = np.zeros(len(ii), bool)
    for r in np.unique(A["repo"][ii]):
        m = np.nonzero(A["repo"][ii] == r)[0]
        cut = np.quantile(ts[m], 0.6)
        train[m] = ts[m] <= cut
    te = ~train
    if train.sum() < 100 or te.sum() < 50:
        return {}
    y = A["y_active"][ii]
    C = np.column_stack([A[c][ii] for c in CONTROLS])
    repo, task = A["repo"][ii], A["task"][ii]
    Dm = dummies(relabel(repo)), dummies(relabel(task))
    base = np.column_stack([np.ones(len(ii)), C, *Dm])
    # standardise predictors with TRAIN-period within-repo statistics only
    qx = np.zeros(len(ii))
    dx = np.zeros(len(ii))
    for r in np.unique(repo):
        m = repo == r
        mt = m & train
        for src, dst in ((A["Q"][ii], qx), (A["D"][ii], dx)):
            dst[m] = src[m] - src[mt].mean()
    sq = np.sqrt((qx[train] ** 2).mean()) or 1.0
    sd = np.sqrt((dx[train] ** 2).mean()) or 1.0
    full = np.column_stack([base, qx / sq, dx / sd])
    const = np.zeros(len(ii))
    for r in np.unique(repo):
        m = repo == r
        const[m] = y[m & train].mean()

    def fit(Z):
        keep = np.abs(Z[train]).sum(0) > 0
        bb, *_ = np.linalg.lstsq(Z[train][:, keep], y[train], rcond=None)
        return Z[:, keep] @ bb

    mse = lambda p: float(np.mean((y[te] - p[te]) ** 2))
    m0, m1, mc = mse(fit(base)), mse(fit(full)), mse(const)
    return dict(P1_mse_const=mc, P1_mse_base=m0, P1_mse_full=m1, P1_rel_gain=1 - m1 / m0, P1_base_vs_const=1 - m0 / mc,
                P1_n_test=int(te.sum()))


# ------------------------------------------------------------------ uncertainty: delete-one-cluster jackknife with full refit
TEST_COLS = {"T1": CONTROLS + ["Q", "y_active"], "T3": CONTROLS + ["Q", "y_active"],
             "T2": CONTROLS + ["D", "Q_stable", "y_active"], "T4": CONTROLS + ["D", "Q_stable", "y_active"],
             "T5": CONTROLS + ["log_hidden", "log_partner_abs", "Q", "y_incomplete"]}
# exploratory statistics (not in the decision family): each statistic's OWN complete-case columns, so cluster presence,
# G and df count only clusters that contribute rows to that statistic (estimate() / predictive_gain() use the same sets)
EXPLORATORY_COLS = {"S1": CONTROLS + ["Q", "y_review"], "S2": CONTROLS + ["Q", "y_rework"],
                    "P1": CONTROLS + ["Q", "D", "y_active"]}


def stat_cols(name):
    if name in TEST_COLS:
        return TEST_COLS[name]
    key = "P1" if name.startswith("P1_") else name
    if key not in EXPLORATORY_COLS:
        raise KeyError(f"no column set for statistic {name}")
    return EXPLORATORY_COLS[key]


def analyse(df, progress=False, cluster="area"):
    """Point estimates on all rows; SE from the delete-one-cluster jackknife, REFITTING the whole pipeline (within-repo
    standardisation, stage-1 + stage-2 models, logit, P1 split) on each leave-one-cluster-out sample:
        se^2 = (G_t-1)/G_t * sum_g (theta_(-g) - mean theta_(-g))^2          (CR3-type; conservative with few/uneven clusters)
    G_t = number of clusters that contribute rows to test t's own complete-case sample (replicates that drop a cluster outside
    that sample are identical to the full fit and are excluded); p-values and CIs use t with G_t-1 df. A non-finite replicate
    for a test whose sample contains the dropped cluster is a hard error (no silent dropping). Chosen over a pairs-cluster
    bootstrap because harness runs showed the bootstrap SE was ~20% too small with ~100-300 uneven clusters."""
    from scipy import stats as st
    A = encode(df)
    n = len(df)
    absorb = A["area"]          # area FE in T2/T4 regardless of the inference cluster
    pt = estimate(A, np.arange(n), absorb)
    cl = A[cluster]
    order = np.argsort(cl, kind="stable")
    bounds = np.flatnonzero(np.diff(cl[order])) + 1
    groups = np.split(order, bounds)
    reps = []
    for k, g in enumerate(groups):
        keep = np.ones(n, bool)
        keep[g] = False
        reps.append(estimate(A, np.nonzero(keep)[0], absorb))
        if progress and k % 100 == 0:
            print("jackknife", k, "/", len(groups), flush=True)
    res = {}
    for nme in [k for k in pt if not k.startswith("n_")]:
        cols = stat_cols(nme)
        exploratory = nme not in TEST_COLS   # S1, S2, P1_*: may be undefined on a replicate; not in the decision family
        rows = _sel(A, np.arange(n), cols)
        present = np.array([len(np.intersect1d(g, rows)) > 0 for g in groups])
        G = int(present.sum())
        v = np.array([r.get(nme, np.nan) for r in reps])[present]
        if exploratory:
            v = v[np.isfinite(v)]
            G = max(len(v), 2)
        elif not np.isfinite(v).all():
            raise RuntimeError(f"non-finite jackknife replicate for {nme}: {int((~np.isfinite(v)).sum())} of {len(v)}")
        est = float(pt[nme])
        se = float(np.sqrt((G - 1) / G * ((v - v.mean()) ** 2).sum())) if G > 1 else float("nan")
        tq95, tq90 = st.t.ppf(0.975, G - 1), st.t.ppf(0.95, G - 1)
        res[nme] = dict(est=est, se=se, ci95=[est - tq95 * se, est + tq95 * se], ci90=[est - tq90 * se, est + tq90 * se],
                        p=float(2 * st.t.sf(abs(est / se), G - 1)) if se > 0 else float("nan"), n_jack=int(len(v)), df=G - 1)
    res["n"] = {k: v for k, v in pt.items() if k.startswith("n_")}
    res["G"] = len(groups)
    return res


# ------------------------------------------------------------------ decisions
def holm(ps):
    order = sorted(ps, key=ps.get)
    out, run = {}, 0.0
    for i, k in enumerate(order):
        run = max(run, min(1.0, (len(ps) - i) * ps[k]))
        out[k] = run
    return out


def decide(res, adequate, t5_decided, h12_valid=True):
    """Single correction layer: Holm over the DECIDED tests among T1..T5 (two-sided). `adequate` (test -> bool, from power.py)
    is required: a test absent from it is inadequate and can never be `equivalent`. If `t5_decided` is False (gate G5 or G7
    failed) T5 is neither fitted into the family nor given a verdict. If `h12_valid` is False (gate G3 or G3b failed) T1-T4 get
    `outcome_construct_invalid` and are removed from the family."""
    fam = [t for t in FAMILY if t in res and (t != "T5" or t5_decided) and (h12_valid or t == "T5")]
    adj = holm({t: res[t]["p"] for t in fam})
    ver = {}
    for t in fam:
        r, s = res[t], SESOI[t]
        sig = adj[t] < ALPHA
        if sig and r["est"] > 0 and r["ci95"][0] > 0:
            v = "supported" if r["est"] >= s else "detected_below_SESOI"
        elif sig and r["est"] < 0 and r["ci95"][1] < 0:
            v = "construct_invalid (favourable direction predicts opposite)"
        elif r["ci90"][0] > -s and r["ci90"][1] < s and adequate.get(t, False):
            v = "equivalent (no effect of interest)"
        else:
            v = "inconclusive"
        ver[t] = dict(adj_p=adj[t], verdict=v)
    if not t5_decided:
        ver["T5"] = dict(adj_p=None, verdict="not_decided (G5/G7 failed; H4 exploratory only)")
    if not h12_valid:
        for t in ("T1", "T2", "T3", "T4"):
            ver[t] = dict(adj_p=None, verdict="outcome_construct_invalid (G3/G3b failed; H1/H2 not tested)")
    # H1/H2 are about DEGRADATION (within-area change) -> T2/T4; level tests are labelled associational (area-level confounding)
    hyp = {"H1": "T2", "H1_level_association": "T1", "H2": "T4", "H2_level_association": "T3", "H4": "T5"}
    out = {h: ver[t]["verdict"] for h, t in hyp.items() if t in ver}
    return ver, out


def main():
    import argparse
    ap = argparse.ArgumentParser()
    ap.add_argument("panel")
    ap.add_argument("out")
    # no tuning parameters and no randomness: the jackknife is deterministic
    ap.add_argument("--power", required=True, help="power-results.json (per-test `adequate`)")
    ap.add_argument("--validation", required=True, help="validation-gates.json written by `validate.py gates` (G4, G5, G7)")
    ap.add_argument("--gates", required=True, help="gates.json written by `gates.py --g3b` (G3, G3b); G3b MUST be present")
    a = ap.parse_args()
    val = json.load(open(a.validation))
    gt = json.load(open(a.gates))
    if gt.get("G3b") is None:
        raise SystemExit("gates.json has no G3b: run `gates.py --g3b` first (PREREGISTRATION §5)")
    h12 = bool(gt["G3"]["pass_"] and gt["G3b"]["pass_"])
    if not val["G4"]["pass"]:
        raise SystemExit("G4 (task-type validation) failed: the study stops and returns to council (PREREGISTRATION §5)")
    df = pd.read_csv(a.panel)
    res = analyse(df, progress=True)
    adequate = {t: v["adequate"] for t, v in json.load(open(a.power))["tests"].items()}
    t5 = bool(val["G5"]["pass"] and val["G7"]["pass"])
    missing = [t for t in FAMILY if t not in res and (t != "T5" or t5) and (h12 or t == "T5")]
    if missing:
        raise SystemExit(f"decision tests without a sample on the full panel: {missing}")
    ver, hyp = decide(res, adequate, t5, h12)
    json.dump(dict(estimates=res, verdicts=ver, hypotheses=hyp, adequate=adequate, validation=val, gates=gt), open(a.out, "w"), indent=1)


if __name__ == "__main__":
    main()
