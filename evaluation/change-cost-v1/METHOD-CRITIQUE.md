# E1/E1b method critique — council memo (read-only)

Inputs: `evaluation/change-cost-v1/*`, `evaluation/inertia-v1/*`, `docs/inertia-research.md`, `docs/scout-design.md §8`. Independent challenger memo: `local://e1b-challenger.md`. Hand-computed numbers are marked; nothing was re-run.

## 1. Recommendation

Treat E1b as a **precisely estimated null for its own proxy outcomes under an attenuating design**, not as evidence about H1/H2/H4. Do not pool adopter data with E1b's definitions (§8.3 "definitions follow E1b") until the measurement defects below are fixed. Replace the area×1-year, LORO-ridge design with a change-level within-repository panel (design A) and, for causal claims, the SlopCodeBench controlled continuation (design B).

## 2. Ranked issues

**Invalidates (for H1/H2/H4 inference)**

1. **Outcome constructs do not measure effort.** O1 = median lines changed per commit in the area. fastapi root area: 76 changes, `out_o1`=1.0986 → median **2 lines** (`outcomes-area.csv:2`) — release version bumps, not effort. O2 is floor-bound: 31/60 areas have `out_o2`=0.693 (median 1 file). O3 counts *any* later commit touching the same file within 7 days, including the same author's next commit; junit5 (median **7 commits/PR**, `out_o4a`=2.08 → e^2.08−1) has O3 up to 0.85 (`outcomes-area.csv` reporting area) — squash-vs-merge workflow, not incompleteness. O4c PR hours: fastapi repo mean 8.12 (≈140 days) vs caddy 4.03 (`model-results.json .area.o4c.y_mean_by_repo`) — backlog/review latency. Code Red measured per-issue cycle time; none of these are comparable. Effect: a null on these outcomes says nothing about effort/predictability.

2. **Principal predictors have (almost) no variance in the largest stratum.** nest = 26/60 units; 22 have `erosion`=1.0, `log_n_high_cc`=0, 23 have `d_erosion`=`d_log_n_high_cc`=0 (`features-area.csv:14-39`). junit5 5/8 at the floor. The "degradation" step therefore tests three nest outliers (platform-express `d_erosion`=−0.92, a 4-file area) after within-repo z-scoring, giving one unit a z of ≈−5 **[hand-computed]**. Clone deltas: `d_clone_pct` nonzero in ~half the sample but mostly |Δ|<0.02. Effect: the degradation hypothesis was never exercised; H1-degradation is untested, not unsupported.

3. **The baseline model is worse than the grand mean.** Var(out_o1) ≈ 0.470 **[hand-computed from `outcomes-area.csv`, n=60]**; LORO baseline MSE = 0.5829 (`model-results.json .area.o1.mse_base`). Out-of-sample R² ≈ −0.24. Cause: predictors are z-scored within repo (`model.py:46-52`) but `y` is not centred, so held-out repo mean shifts (O1 repo means 1.67–2.85) sit in the MSE denominator and no predictor can touch them; for O4c ≈38 % of MSE is between-repo **[hand-computed]**. "Relative MSE reduction" compares two models both inferior to a constant; 5 % of that denominator is a different quantity from 5 % of explainable variance. Also noted independently by the challenger (§3).

**Weakens**

4. **Ridge at/near grid ceiling → intercept-only comparisons.** Selected alpha = 1e5 for O2, O4a, O5b level+degradation, O4b/O4c coupling (`model-results.json alpha_median`); with n_train≈50 z-scored units, alpha=100 (O1) already shrinks coefficients to ≈1/3 **[algebra: n/(n+α)]**. The per-repo "improved" counts for nest/caddy are +0.00 %/+0.01 % (`tables.md:187`), so "3/3 repos improved" is sign noise at the fourth decimal; with 3 big repos the ≥all-but-one rule is P=0.5 under the null.

5. **Power/MDE.** O1 level CI half-width ≈2.1 points (−2.72…+1.42). To satisfy "≥5 % and lower bound >0" the point estimate needed ≈+7 %, i.e. a true gain of ≈20 % of explainable within-repo variance after shrinkage **[INFERENCE]**. O3 degradation CI spans 37 points. Challenger's correction, adopted: O1's CI *excludes* +5 % — it is not "underpowered", it is a tight null for this proxy/design; O4c (+9.05 % [−2.89, +18.04]) is the genuinely underpowered row.

6. **Multiplicity is double-corrected.** `decide()` (`model.py:154-160`) takes Bonferroni-within-family then Holm across 5 families, and requires rule AND adjusted p<0.05. With one-sided percentile-bootstrap p's at n=35–60 the effective threshold is ≈0.003–0.005; O2–O5/H4 support was unreachable by construction.

7. **Selection on activity.** Inclusion requires ≥5 outcome-window changes (`features.py:193`); `n_changes` ranges 5–92. Avoided/stable areas (a plausible inertia consequence) are excluded; conditioning on post-T0 activity selects on a descendant of exposure **[INFERENCE]**. PR outcomes further require ≥5 linked PRs in both windows (n=35; junit5 1 unit → all its z-scores 0, fastapi 2 units → ±1).

8. **Lagged-outcome control.** The lag is a prior-window value computed with no minimum count: 11/60 areas have <5 prior changes, 5 have ≤2 (`features-area.csv n_prior_changes`), so it is noisy and absorbs *less* than feared. Causally, prior-window churn precedes Q(T0) and can cause it (churn→degradation), so it is a confounder, not a mediator/collider — controlling for it is correct for a causal reading of Q(T0). What it does do: make the test one of *incremental* value over history, which is the right product question but a harder bar. Main agent's "absorbs quality's effect" is therefore overstated.

9. **Code defects (verified by reading).** (a) `collect.py` `table()`: deletions get churn 0 (`paths=[(old,0,0)]`) — removals are effort. (b) `features.py follow_flags`: `t - ts <= 7*DAY` admits negative differences; seq is topo-order, so parallel-branch commits with earlier timestamps count as follow-ups (inflates O3 in merge-commit repos). (c) Follow-up stream limited to the window → right-censoring at T_end, acknowledged. (d) jscpd lines are physical, SLOC is NLOC → `clone_pct` biased up (capped at 1). (e) Degradation deltas compare different file populations (`t1_files` keyed by the T0 map, `features.py:181-186`) — population change masquerades as degradation (§8.2 already fixes this for Scout).

10. **Bootstrap without refit** (`model.py:86-112`): captures evaluation-set noise only; training/tuning variance across 5 folds is omitted, so CIs are too narrow for the 1e5-alpha rows and the repo-count rule.

11. **E1-specific.** Regex-fix precision 0.52 (0.64 on source-touching commits; `classifier-metrics.json`); nest `fix(deps)` dominates. Non-differential label noise attenuates toward zero; differential noise (bot `fix(deps)` concentrated in some repos) can bias either way (challenger, adopted). Note E1's own ablation `log_n_high_cc` +1.91 % [+0.39, +3.68], 6/7 repos, and within-repo Spearman +0.31 (`RESULTS.md`): a real but small incremental signal beyond size/history — consistent with the defect-prediction literature where process metrics dominate.

**Minor**: `hotspot` is `Σ num·log1p(prior commits)/SLOC`, collinear with baseline `log_prior_changes` and `log_n_high_cc`; `erosion` is a health score (1 − burden), confusingly named (RESULTS flags it); bot regex `(?i)(bot|…)` matches names containing "bot"; PR linkage misses rebase merges (junit5 24/151 linked).

## 3. Verdict on the main agent's summary

Right: no support; sample too small; conclusions are "not shown" not "disproved"; descriptive-only consequence. Wrong/overstated: (i) "lagged outcome may absorb quality's effect" — minor; the lag is noisy and causally a confounder (issue 8); (ii) "too small" misses that O1's CI excludes 5 % — the primary result is a tight null for the proxy, and the reasons it is uninformative are construct validity (1), zero-variance predictors (2) and the MSE denominator (3), not n alone; (iii) it omitted that the degradation hypothesis was never exercised (2) and that secondary-outcome support was unreachable under the double correction (6).

## 4. What can and cannot be concluded

Can: within these 5 repos, CC-burden/clone/hotspot levels at T0 add nothing to a shrunk ridge forecast of next-year median commit churn beyond size/history, cross-repo; E1: static levels add ≈1–2 % to file fix-touch forecasting beyond size/history — small, positive, below the bar. Cannot: anything about effort, predictability or hidden coupling (constructs not measured); anything about *degradation* (no variance); anything causal; anything about agent magnification. The file-level O5a `clone_pct` +6.45 % [+2.02, +11.85] (36 uncorrected tests) is hypothesis-generating only.

## 5. Better designs

**A. Change-level within-repository panel (observational, feasible now).** Unit = PR (or squash/merge-equivalent change), 7 repos × 2 years ≈ 4,100 merged PRs already cached (`RESULTS.md` dataset table), ≈1,300 included-area changes. Outcome: log PR cycle time split into *review latency* (first review → merge) vs *author time* (created → first review), plus review rounds and post-merge corrective PRs adjudicated by Jev (replace regex; precision measured). Predictors: principal of the *files/functions actually touched* at the PR base commit (CC burden, clone lines, cycle membership), measured on a stable population. Fixed effects: repo, author, quarter; task category covariate (Jev-labelled: feature/fix/docs/deps). Cluster SEs by area; validate on future quarters within repo and on held-out repos separately. H2 via conditional dispersion (residual variance by principal tercile). H4 via co-change pairs *without* an import edge (tier-1 languages only). Power **[INFERENCE, normal approximation]**: n≈3,000 PRs, ICC≈0.1 with ≈40 per area → design effect ≈5, n_eff≈600, MDE for a standardised slope at 80 %/α=0.05 ≈ 2.8/√600 ≈ 0.11 SD — far below Code Red's 78–124 % effects. Needs: lizard at each PR base (minutes per snapshot), Jev labels, the existing PR cache.

**B. Controlled continuation on SlopCodeBench (X8, causal).** Paired tasks on behaviour-equivalent healthy vs degraded snapshots, fresh sessions, blinded scoring; outcomes success, steps, cost, corrective edits, conditional variance. Challenger's estimate adopted: d=0.3 at 80 % needs ≈88 independent task pairs; seeds do not substitute for tasks. Only design that tests agent magnification (IR §5).

**C (cheap re-analysis of E1b data).** Quarterly area-windows (60 areas × 8 ≈ 480 area-windows) with area fixed effects, Δinterest on lagged Δprincipal, y centred within repo. Requires 8 lizard snapshots per repo; detects within-area change-on-change at MDE ≈0.2 SD **[INFERENCE]** but inherits issues 1 and 2.

## 6. Implications for `docs/scout-design.md §8`

- §8.3 "Definitions follow E1b's preregistration so adopter data pools with the study": pooling with defective definitions is not an asset. Fix before shipping: count deletion churn; require t>ts for follow-ups and exclude same-author/same-PR commits; normalise to PR-level changes (squash-equivalent) so workflows compare; split PR hours into author vs review latency; store the change table (already planned, §8.6) so outcomes can be redefined.
- §8.5 "a 5 % effect could not be reliably detected" and "≥300 scope-windows from ≥15 repositories" should be replaced: the gap is construct and design, not n; state design A as the upgrade path and derive n by cluster-aware simulation.
- §8.5 upgrade rule reuses E1b's rule (relative MSE, LORO, repo-count). Replace with: within-repo-centred outcomes, pre-specified MDE, refit-inclusive uncertainty, one correction layer.
- Keep: descriptive claim level, no composite, population digest for deltas (§8.2 already fixes issue 9e), `insufficient` below n=20.
- Do not describe `scout.churn_per_change` or `follow_up_rate` as effort/incompleteness measures in panels; label them activity proxies.

## 7. Where the challenger changed this memo

Adopted: O1 is a tight null, not underpowered (§2.5); deletion-churn and negative-time follow-up defects (§2.9a-b); differential label noise can bias either way (§2.11); do not make "more windows" the first remedy; the ≈88-pair figure for design B. Rejected: challenger's PR-window leakage claim — PRs link through commits that are themselves inside the window's ancestry (`features.py linked_prs` over `area_chg[win]`), so a PR cannot enter the prior-window lag from post-T0 commits; only its created-at can predate the window (measurement, not leakage).
