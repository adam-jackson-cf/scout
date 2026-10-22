# E1b — Does quality degradation predict the cost of later changes? Results (change-cost-v1)

> **Superseded (2026-10-20).** Invalidated by the method critique in `METHOD-CRITIQUE.md`: the baseline model was worse than a constant (held-out MSE 0.583 vs outcome variance 0.470), the outcomes did not measure effort, and the quality predictors barely varied in the sample. It was also run before the project's experiment rules (`AGENTS.md`) existed. Replacement: `evaluation/change-cost-v2/`.

Design: `PREREGISTRATION.md` (unchanged). All numbers are produced by the scripts here from the pinned repositories and GitHub
GraphQL responses cached in `raw/`. Full generated tables (every comparison, ablation, per-repo reduction, bootstrap
coefficient): `tables.md`; machine-readable: `model-results.json`, `features-*.csv`, `outcomes-*.csv`, `dataset-counts.json`.

## Decisions (preregistered rules)

| Hypothesis | Rule | Result |
|---|---|---|
| **H1** (degradation/level raises change effort), primary O1 churn per change | support if level or degradation meets ≥5%, CI lower > 0, ≥ all-but-one of the repositories with ≥10 units improved (uncorrected) | **Not supported.** Level: −0.44% [−2.72%, +1.42%]; degradation (on top of level): +0.49% [−2.89%, +4.33%] |
| H1 partial (O2, O3, O4a-c) | same rule + Holm | **No partial support on any secondary outcome.** Largest nominal: O4c (PR hours) level +9.05% [−2.89%, +18.04%], p=0.058; fails the CI rule and Holm (family adj. p = 1.0) |
| **H2** (predictability, O5) | level or degradation on O5, Holm-corrected | **Not supported.** O5a (churn dispersion) level +4.08% [−9.70%, +14.37%], degradation +2.13% [−6.55%, +14.41%]; O5b (PR-hours dispersion) level −5.43% [−12.97%, +0.77%], degradation +2.78% [−0.38%, +6.32%] (p=0.046 one-sided, but < 5% and CI includes 0; Holm-adjusted family p = 0.93) |
| **H4** (coupling → incomplete changes), O3 follow-up rate | coupling on O3, Holm-corrected | **Not supported.** +1.49% [−3.23%, +5.68%], 2/3 repos improved (need 2), below 5%; Holm-adjusted p = 1.0 |

**Consequence under the preregistration:** the quality signals (erosion, n high-CC functions, clone %, hotspot, and their T−1→T0
deltas) and the coupling and ownership signals stay **descriptive** in Scout; none is documented as predictive of later change cost. No CI gating.
The result is *absence of demonstrated predictive value in a small sample*, not demonstrated absence: see power below.

Construct-invalid rule: two area coefficients are flagged by the literal rule (O5b, `d_erosion` +1.1e-4 [5e-6, 2.2e-4] and
`d_log_n_high_cc` −1.3e-4 [−2.3e-4, −3e-5], favourable direction predicting worse dispersion). **These are numerical
artefacts**: the inner CV chose the grid ceiling alpha = 1e5, so all coefficients are ≈1e-4 and the bootstrap interval excludes 0 only
by epsilon. Reported as flagged, not interpreted. No other predictor flagged.
For O1 level, `erosion` has a coefficient of −0.045 [−0.098, −0.003] (higher erosion-health → lower churn: intended direction).

## Dataset

Windows: `T0` = last first-parent commit with committer time ≤ `T_end − 365 d`; `T−1` = same rule at `T0 − 365 d`; `T_end` = pinned
revision (all seven repository HEADs verified equal to `reference-v1.json` / E1 `meta.json` revisions). Bot commits excluded by author-name regex
`(?i)(bot|renovate|dependabot)` (`areas.json`). Source roots in `areas.json` (written 2026-10-19 16:09:46, sha256 `6764739f…e2f4`, before any
collection/outcome step; first PR response cached 16:11:03, first collection cache 16:11:49).

| repo | raw areas T0 | eligible areas (≥3 files after merge) | exist at T−1 | ≥5 outcome changes | **included** | Σ changes in included areas | PRs fetched (T−1…T_end) | PRs linked to included-area changes | included areas with O4 unmeasured (<5 linked PRs) | files in file-level analysis |
|---|---|---|---|---|---|---|---|---|---|---|
| fastapi | 6 | 3 | 3 | 2 | 2 | 82 | 568 | 53 | 0 | 6 |
| webpack | 38 | 30 | 28 | 11 | 10 | 154 | 665 | 75 | 4 | 11 |
| nest | 62 | 38 | 37 | 26 | 26 | 526 | 2,244 | 158 | 3 | 43 |
| caddy | 28 | 17 | 16 | 14 | 14 | 408 | 419 | 207 | 2 | 39 |
| axum | **excluded** — repository history begins after T−1 (no first-parent commit ≤ T0−365 d) | | | | | | | | | |
| junit5 | 28 | 15 | 14 | 9 | 8 | 131 | 151 (`junit-team/junit-framework`) | 24 | 7 | 4 |
| eShopOnWeb | 38 | 25 | 25 | 0 | **0** | 0 | 74 | 0 | – | 0 |
| **total** | | | | | **60** | 1,301 | 4,121 | 517 | 16 | 103 |

* eShopOnWeb has only 17 commits in its outcome window (E1 saw the same), so no area reaches 5 changes. It contributes nothing.
* Eligible outcome-window changes (non-merge, ≤50 files, non-bot, touching ≥1 eligible-path file): fastapi 84, webpack 171, nest 450, caddy 302, junit5 122, eShopOnWeb 0.
  Bot commits removed (both windows): fastapi 56, webpack 192, nest 1,852, caddy 4, junit5 9, eShopOnWeb 49.
* Area-level model samples: O1/O2/O3/O5a n = 60 (fastapi 2, webpack 10, nest 26, caddy 14, junit5 8); PR outcomes need ≥5 linked PRs in **both** windows (lagged outcome) → n = 35 (fastapi 2, webpack 6, nest 15, caddy 11, junit5 1).
  Repositories with ≥10 units (used in the repository-count rule): webpack, nest, caddy for O1/O2/O3/O5a (need ≥2 of 3, "all but one"); nest, caddy for PR outcomes (need ≥1 of 2).
* File-level secondary: 103 files (fastapi 6, webpack 11, nest 43, caddy 39, junit5 4).

## Main area-level results (ridge, z-scored within repository, leave-one-repository-out, 2,000 stratified bootstrap resamples)

Pooled relative MSE reduction (positive = better than the comparison base), 95% bootstrap CI, one-sided bootstrap p (fraction of resamples ≤ 0).
`rule` = preregistered support rule (≥5%, CI lower > 0, repository-count condition). **No row meets it.**

| outcome | n | comparison | reduction | 95% CI | p | repos improved (≥10 units) | rule |
|---|---|---|---|---|---|---|---|
| O1 churn/change | 60 | base → +level | −0.44% | [−2.72%, +1.42%] | 0.679 | 3/3 | no |
| O1 | 60 | +level → +level+degradation | +0.49% | [−2.89%, +4.33%] | 0.432 | 2/3 | no |
| O2 amplification | 60 | base → +level | +3.88% | [−3.03%, +11.44%] | 0.137 | 2/3 | no |
| O2 | 60 | +level → +degradation | −0.52% | [−1.68%, +0.30%] | 0.869 | 1/3 | no |
| O3 follow-up rate | 60 | base → +level | −0.92% | [−6.33%, +3.54%] | 0.636 | 3/3 | no |
| O3 | 60 | +level → +degradation | −2.78% | [−23.61%, +13.55%] | 0.607 | 1/3 | no |
| **O3 (H4)** | 60 | base → +coupling | +1.49% | [−3.23%, +5.68%] | 0.256 | 2/3 | no |
| O3 | 60 | base → +ownership | −1.73% | [−6.35%, +2.36%] | 0.785 | 0/3 | no |
| O4a PR commits | 35 | base → +level | −0.27% | [−3.55%, +2.73%] | 0.534 | 0/2 | no |
| O4a | 35 | +level → +degradation | +3.46% | [−1.66%, +8.89%] | 0.093 | 0/2 | no |
| O4b PR reviews+comments | 35 | base → +level | +0.21% | [−8.82%, +7.60%] | 0.460 | 1/2 | no |
| O4b | 35 | +level → +degradation | −4.09% | [−11.24%, +2.28%] | 0.893 | 1/2 | no |
| O4c PR hours | 35 | base → +level | +9.05% | [−2.89%, +18.04%] | 0.058 | 1/2 | no |
| O4c | 35 | +level → +degradation | +1.08% | [−6.41%, +7.30%] | 0.372 | 1/2 | no |
| O5a churn dispersion (H2) | 60 | base → +level | +4.08% | [−9.70%, +14.37%] | 0.287 | 2/3 | no |
| O5a | 60 | +level → +degradation | +2.13% | [−6.55%, +14.41%] | 0.327 | 1/3 | no |
| O5b PR-hours dispersion (H2) | 35 | base → +level | −5.43% | [−12.97%, +0.77%] | 0.962 | 1/2 | no |
| O5b | 35 | +level → +degradation | +2.78% | [−0.38%, +6.32%] | 0.046 | 2/2 | no |

Coupling/ownership on the other outcomes and all ablations (baseline + one predictor, 9 predictors × 8 outcomes): `tables.md`.
No ablation meets the rule at area level. Highlights: O1 `d_erosion` +3.00% [−0.80%, +7.14%] (p 0.063, 3/3 repos); O4c `d_erosion`
+8.54% [−5.77%, +20.38%], `hotspot` +7.34%, `erosion` +6.55% (all CIs span 0); a few ablations are significantly *worse* than baseline
(O4a `hotspot` −0.21%, O5a `d_log_n_high_cc` −1.72% [−4.02%, −0.40%], O5b `hotspot` −3.51%).

Holm (five outcome families; family p = min over the family's tested comparisons × number of comparisons, then Holm):
adjusted p = O1 1.00, O2 1.00, O3 1.00, O4 1.00, O5 0.93; before Holm: O1 0.86, O2 0.27, O3 0.77, O4 0.35, O5 0.19 (all Bonferroni-within-family). Nothing survives.

**Per-repository reductions** (area level; "-" = no included units): in `tables.md` ("per-repo relative MSE reduction"). Reductions are near
0.00% for nest and caddy on most O2/O4/O5b comparisons because inner CV chose the ridge ceiling (alpha 1e2…1e5), making the model ≈ intercept
for both base and candidate. Examples (level / degradation / coupling reductions, O1): fastapi +4.84% / −7.04% / −9.57% (2 units), webpack +2.18% / +0.98% / +0.31%,
nest +0.00% / +0.00% / −3.43%, caddy +0.01% / −0.00% / +0.00%, junit5 −7.41% / +2.76% / +1.22%. For O3 coupling: fastapi +24.89% (2 units), webpack −0.79%, nest +0.00%, caddy +6.15%, junit5 +1.63%.
Per-repo reductions have no stable sign for any predictor set.

## Secondary: file level (unit = T0 file with ≥5 changes in the outcome window, n = 103; O1, O2, O3, O5a)

No preregistered file-level decision rule exists beyond the same support rule; none of the base → level / degradation / coupling / ownership
comparisons meets it (`model-results.json: file_secondary_rule_met`). Nominal observations (uncorrected, exploratory): O5a `clone_pct`
ablation +6.45% [+2.02%, +11.85%] meets the rule on its own (2/3 repos), and `log_n_high_cc` +3.38% [+1.12%, +6.26%], `erosion` +3.06% [+0.81%, +5.98%];
O2 `ownership` +3.64% [+0.77%, +6.31%] (< 5%). The file-level family has 36 ablation tests with no Holm correction, 5 repositories with data,
n = 103: treat as hypothesis-generating at most; they do not alter the area-level decisions. Degradation at file level never helps
(O1 `d_erosion` −2.71% [−4.56%, −1.29%], `d_n_high_cc` −1.62%, O5a degradation step −2.69% [−6.14%, −0.65%]).

## Threats / limits (read with the decisions)

* **Power.** 60 areas from 5 repositories, three with ≥10 units. The 95% intervals are wide (O1 level CI [−2.7%, +1.4%], O3 degradation [−23.6%, +13.6%]); an effect of 5% could not be reliably detected. Absence of support ≠ absence of effect.
* Ridge shrinkage: inner CV picked the grid ceiling (1e5) for several PR-outcome models (O2 level, O4a level, O4b/O4c coupling, O5b), driving predictions to the intercept; those comparisons are close to degenerate. The alpha grid is the preregistered 10⁻²…10⁵.
* PR outcomes: only 44 areas have ≥5 linked PRs in the outcome window and 35 in both windows; PRs link through the merge/squash commit and the first 100 PR commit oids only (rebase merges whose rewritten oids differ are unlinked; reviews/comments are counted from the first 100 reviews). junit5 merges few PRs (151 in two years).
* Churn and PR metrics are proxies for effort; squash vs merge workflows differ per repository; bot filter is name-based (the regex also matches any author name containing "bot").
* Directory areas are not Clawpatch features; a single time split; coupling is raw co-change with no import-graph filtering (preregistered scope).
* O3 follow-ups are observed within the same window only (a change in the last 7 days before T0/T_end can't be followed past the boundary): mild right-censoring.
* fastapi (2 areas) and junit5 (1 PR-measured area) are almost empty folds; they affect the repository-count rule but contribute little.

## Deviations from the preregistration / impossible steps

1. **axum excluded**: no first-parent commit exists at or before `T0 − 365 d` (axum history starts mid-2021; T0 = 2022-03-21), so there is no T−1 snapshot or prior window. Recorded in `meta.json`/`dataset-counts.json`; the study has six repositories, one (eShopOnWeb) contributing no units.
2. junit5 GitHub slug is now `junit-team/junit-framework` (`junit-team/junit5` search returns 0 PRs); fetched there.
3. **Area membership of a change is by path at commit time** (ext + exclusion-glob eligible, area rule + the T0 merge map), including files created after T0; the ≥10 SLOC filter applies to T0/T−1 snapshot files only (not checkable per commit). Renames count toward both old and new path areas. Paths in areas absent at T0 are not tracked.
4. "Area exists at T−1" = ≥1 eligible file (≥10 SLOC) under the same final area key at T−1. Merge rule: areas with <3 files at T0 merge once into the parent directory; the parent keeps ≥3 total files or is excluded.
5. Median guards: p90/median uses `max(median, 1)` (churn lines, and hours for PR dispersion); lagged outcome is 0 when the prior window has no changes (O1/O2/O3/O5a); PR-based lags are NaN → area excluded from that outcome's model (preregistered).
6. O5 is split into O5a (churn dispersion, all areas) and O5b (PR-hours dispersion, PR-measured areas) and both are tested; Holm family O5 combines them. O3 follow-up counts any non-merge non-bot commit (any size) touching the change's files in the area within 7 d, after it in topological/time order.
7. Holm: one-sided bootstrap p-values (B+1 smoothing); family p = Bonferroni-within-family min; Holm across the five families; support for O2–O5/H4 requires the rule *and* Holm-adjusted p < 0.05. O1 uses the uncorrected rule per the prereg.
8. Repository-count rule counts repositories with ≥10 *included* units for that outcome; 3 repos (O1/O2/O3/O5a) or 2 (PR outcomes): "all but one" (fewer than 6 big repos).
9. File age = raw days since the file's first commit (rename-followed); no log transform. Inner CV is leave-one-repository-out over training repositories (E1's implementation of "inner grouped CV").
10. Author-name matching only (`%an`); PR window is by merged date (UTC) over `[T−1, T_end]`.
11. File-level units follow renames (E1 logic) and require a path at T−1; area-level units do not rename-track files.

## Files

`areas.json` (written first), `exclusions.json` (copy of E1's), `fetch_prs.py`, `collect.py`, `features.py`, `model.py`, `report.py`;
`raw/` (cached GraphQL responses and `prs-<repo>.json`), `cache/` (per-repo collection cache: snapshots' lizard/jscpd metrics, commit tables),
`meta.json`, `dataset-counts.json`, `features-{area,file}.csv`, `outcomes-{area,file}.csv`, `model-results.json`, `tables.md`.

## Reproduce

```
cd evaluation/change-cost-v1
python3 fetch_prs.py                                          # needs `gh` auth; cached in raw/; ~5 min
uv run --with lizard,numpy,pandas python collect.py           # uses /tmp/scout-evidence/repos + jscpd on PATH; caches to /tmp/scout-evidence/e1b/cache (copy in cache/)
uv run --with numpy,pandas python features.py
uv run --with numpy,pandas,scikit-learn python model.py       # seed 20261019, B = 2000
uv run python report.py                                       # regenerates tables.md
```
