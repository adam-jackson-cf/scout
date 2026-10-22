#!/usr/bin/env python3
"""E1b stage 3: nested ridge models, leave-one-repository-out, bootstrap, Holm, decisions (PREREGISTRATION.md).

Usage: uv run --with numpy,pandas,scikit-learn python model.py
Reads features-{area,file}.csv + outcomes-{area,file}.csv; writes model-results.json.
Adapted from ../inertia-v1/model.py (same seed scheme, alpha grid, bootstrap).
"""
import json
import warnings
from pathlib import Path

import numpy as np
import pandas as pd
from sklearn.linear_model import Ridge

warnings.filterwarnings("ignore")
HERE = Path(__file__).resolve().parent
SEED = 20261019
B = 2000
ALPHAS = np.logspace(-2, 5, 15)
REPO_ORDER = ["fastapi", "webpack", "nest", "caddy", "axum", "junit5", "eShopOnWeb"]

BASE0 = {"area": ["log_sloc", "log_prior_changes", "log_prior_authors", "log_nfiles", "median_age"],
         "file": ["log_sloc", "log_prior_changes", "log_prior_authors", "median_age"]}
LEVEL = ["erosion", "log_n_high_cc", "clone_pct", "hotspot"]
DEGR = ["d_erosion", "d_log_n_high_cc", "d_clone_pct"]
COUP, OWN = ["coupling"], ["ownership"]
# favourable direction => expected coefficient sign if the predictor behaves as intended (higher = costlier)
EXPECT = {"erosion": -1, "log_n_high_cc": 1, "clone_pct": 1, "hotspot": 1,
          "d_erosion": -1, "d_log_n_high_cc": 1, "d_clone_pct": 1, "coupling": 1}
OUTCOMES = {"area": ["o1", "o2", "o3", "o4a", "o4b", "o4c", "o5a", "o5b"], "file": ["o1", "o2", "o3", "o5a"]}
FAMILY = {"o1": "O1", "o2": "O2", "o3": "O3", "o4a": "O4", "o4b": "O4", "o4c": "O4", "o5a": "O5", "o5b": "O5"}


def run(level, oc):
    feat = pd.read_csv(HERE / f"features-{level}.csv")
    outs = pd.read_csv(HERE / f"outcomes-{level}.csv")
    df = feat.merge(outs[["repo", "unit", "out_" + oc]], on=["repo", "unit"])
    y_col, lag_col = "out_" + oc, "lag_" + oc
    df = df[df[y_col].notna() & df[lag_col].notna()].reset_index(drop=True)
    res = dict(outcome=oc, n=int(len(df)), counts=df.groupby("repo").size().to_dict())
    if len(df) < 30 or df.repo.nunique() < 3:
        res["skipped"] = "too few units"
        return res
    base = BASE0[level] + [lag_col]
    cols_all = base + LEVEL + DEGR + COUP + OWN
    X = df[cols_all].astype(float)
    Z = X.copy()
    for r in df.repo.unique():
        m = (df.repo == r).values
        sub = X.loc[m]
        mu, sd = sub.mean(), sub.std(ddof=0)
        Z.loc[m] = ((sub - mu) / sd.replace(0, np.nan)).fillna(0.0).values
    Z = Z.fillna(0.0)  # missing clone metrics (jscpd failure) -> repo mean
    repos = df.repo.values
    y = df[y_col].values
    present = [r for r in REPO_ORDER if r in set(repos)]
    big = [r for r in present if (repos == r).sum() >= 10]
    idx_by = {g: np.where(repos == g)[0] for g in present}
    res.update(repos_ge10=big, y_mean_by_repo=df.groupby("repo")[y_col].mean().to_dict())

    def best_alpha(cols, tr):
        trr = sorted(set(repos[tr]))
        if len(trr) < 2:
            return 1.0
        scores = []
        for a in ALPHAS:
            tot, n = 0.0, 0
            for g in trr:
                va, fi = tr[repos[tr] == g], tr[repos[tr] != g]
                m = Ridge(alpha=a).fit(Z.loc[fi, cols], y[fi])
                tot += ((m.predict(Z.loc[va, cols]) - y[va]) ** 2).sum()
                n += len(va)
            scores.append(tot / n)
        return ALPHAS[int(np.argmin(scores))]

    def loro(cols):
        pred, alphas = np.zeros(len(df)), {}
        for g in present:
            te, tr = idx_by[g], np.where(repos != g)[0]
            a = best_alpha(cols, tr)
            alphas[g] = float(a)
            pred[te] = Ridge(alpha=a).fit(Z.loc[tr, cols], y[tr]).predict(Z.loc[te, cols])
        return pred, alphas

    rng = np.random.default_rng(SEED)
    boot = np.stack([np.concatenate([rng.choice(idx_by[g], len(idx_by[g]), replace=True) for g in present])
                     for _ in range(B)])
    cache = {}

    def pred(cols):
        k = tuple(cols)
        if k not in cache:
            cache[k] = loro(cols)
        return cache[k]

    def compare(name, cb, cq):
        pb, _ = pred(cb)
        pq, aq = pred(cq)
        seb, seq = (pb - y) ** 2, (pq - y) ** 2
        pooled = 1 - seq.mean() / seb.mean()
        bs = 1 - seq[boot].mean(1) / seb[boot].mean(1)
        per = {g: float(1 - seq[idx_by[g]].mean() / seb[idx_by[g]].mean()) for g in present}
        need = len(big) if len(big) < 6 else 5
        if len(big) < 6:
            need = max(len(big) - 1, 1)
        improved = int(sum(per[g] > 0 for g in big))
        lo = float(np.percentile(bs, 2.5))
        p = float((1 + (bs <= 0).sum()) / (B + 1))
        return dict(name=name, added=[c for c in cq if c not in cb], mse_base=float(seb.mean()), mse_model=float(seq.mean()),
                    rel_reduction=float(pooled), ci95=[lo, float(np.percentile(bs, 97.5))], p_one_sided=p,
                    per_repo=per, repos_improved_big=improved, repos_needed=need, n_big=len(big),
                    rule_met=bool(pooled >= 0.05 and lo > 0 and improved >= need), alpha_median=float(np.median(list(aq.values()))))

    comps = {
        "level": compare("base vs base+level", base, base + LEVEL),
        "degradation": compare("base+level vs base+level+degradation", base + LEVEL, base + LEVEL + DEGR),
        "coupling": compare("base vs base+coupling", base, base + COUP),
        "ownership": compare("base vs base+ownership", base, base + OWN),
    }
    res["comparisons"] = comps
    res["mse_base"] = comps["level"]["mse_base"]
    res["ablations"] = {q: compare(f"base vs base+{q}", base, base + [q]) for q in LEVEL + DEGR + COUP + OWN}

    def boot_coefs(cols, alpha, targets):
        cs = np.zeros((B, len(targets)))
        for b in range(B):
            i = boot[b]
            m = Ridge(alpha=alpha).fit(Z.iloc[i][cols], y[i])
            cs[b] = [m.coef_[cols.index(t)] for t in targets]
        m = Ridge(alpha=alpha).fit(Z[cols], y)
        out = {}
        for j, t in enumerate(targets):
            ci = [float(np.percentile(cs[:, j], 2.5)), float(np.percentile(cs[:, j], 97.5))]
            sign = EXPECT.get(t)
            invalid = bool(sign is not None and ((sign > 0 and ci[1] < 0) or (sign < 0 and ci[0] > 0)))
            out[t] = dict(coef=float(m.coef_[cols.index(t)]), ci95=ci, construct_invalid=invalid)
        return out
    a_deg = comps["degradation"]["alpha_median"]
    res["coefs_level_model"] = boot_coefs(base + LEVEL, comps["level"]["alpha_median"], LEVEL)
    res["coefs_full_model"] = boot_coefs(base + LEVEL + DEGR, a_deg, LEVEL + DEGR)
    res["coefs_coupling_model"] = boot_coefs(base + COUP, comps["coupling"]["alpha_median"], COUP)
    return res


def holm(ps):
    order = sorted(ps, key=ps.get)
    adj, run_max = {}, 0.0
    for i, k in enumerate(order):
        run_max = max(run_max, min(1.0, (len(order) - i) * ps[k]))
        adj[k] = run_max
    return adj


def decide(area):
    """Preregistered decisions from area-level results."""
    R = {oc: area[oc] for oc in OUTCOMES["area"] if "comparisons" in area[oc]}
    tests = {}  # family -> list of p
    for oc, r in R.items():
        names = ["level", "degradation"] + (["coupling"] if oc == "o3" else [])
        for n in names:
            tests.setdefault(FAMILY[oc], []).append(r["comparisons"][n]["p_one_sided"])
    fam_p = {f: min(1.0, min(ps) * len(ps)) for f, ps in tests.items()}
    fam_adj = holm(fam_p)
    dec = dict(family_p_bonferroni_within=fam_p, family_p_holm=fam_adj)

    def sup(oc, n, corrected):
        c = R[oc]["comparisons"][n]
        ok = c["rule_met"] and (not corrected or fam_adj[FAMILY[oc]] < 0.05)
        return bool(ok)
    d = {}
    d["H1_primary_O1"] = {n: sup("o1", n, False) for n in ("level", "degradation")} if "o1" in R else None
    d["H1_supported"] = bool(d["H1_primary_O1"] and any(d["H1_primary_O1"].values()))
    d["H1_partial_secondary"] = {oc: {n: sup(oc, n, True) for n in ("level", "degradation")}
                                 for oc in ("o2", "o3", "o4a", "o4b", "o4c") if oc in R}
    d["H2_O5"] = {oc: {n: sup(oc, n, True) for n in ("level", "degradation")} for oc in ("o5a", "o5b") if oc in R}
    d["H2_supported"] = bool(any(v for oc in d["H2_O5"].values() for v in oc.values()))
    d["H4_coupling_O3"] = sup("o3", "coupling", True) if "o3" in R else None
    d["H4_supported"] = bool(d["H4_coupling_O3"])
    construct_invalid = []
    for oc, r in R.items():
        for key in ("coefs_level_model", "coefs_full_model", "coefs_coupling_model"):
            for t, v in r[key].items():
                if v["construct_invalid"]:
                    construct_invalid.append(f"{oc}:{t}({key})")
    d["construct_invalid"] = sorted(set(construct_invalid))
    dec["decisions"] = d
    return dec


def main():
    out = {}
    for level in ("area", "file"):
        out[level] = {}
        for oc in OUTCOMES[level]:
            print(level, oc, flush=True)
            out[level][oc] = run(level, oc)
    out["decision"] = decide(out["area"])
    # file-level secondary: support flags only (no Holm)
    out["file_secondary_rule_met"] = {
        oc: {n: out["file"][oc]["comparisons"][n]["rule_met"] for n in ("level", "degradation", "coupling", "ownership")}
        for oc in OUTCOMES["file"] if "comparisons" in out["file"][oc]}
    (HERE / "model-results.json").write_text(json.dumps(out, indent=1, default=lambda o: float(o) if hasattr(o, "item") else str(o)))
    print(json.dumps(out["decision"], indent=1))


if __name__ == "__main__":
    main()
