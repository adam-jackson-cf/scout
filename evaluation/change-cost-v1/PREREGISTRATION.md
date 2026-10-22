# E1b — Does quality degradation predict the cost of later changes? Preregistration (2026-10-19)

Written before any outcome in this study was computed or inspected. Follows E1
(`evaluation/inertia-v1/`), which tested fix-proneness from quality *levels* at file level and
found no support. This study tests the user's hypotheses H1, H2 and H4 (`docs/inertia-research.md` §4):
degradation raises the effort and unpredictability of later changes in the same area, and hidden
coupling causes incomplete changes.

## Targets and windows

The seven pinned targets in `evaluation/reference-v1.json`, same `T_end` as E1.
`T0` = last first-parent commit at or before `T_end − 365 d`; `T−1` = same rule at `T0 − 365 d`.
Prior window `(T−1, T0]`; outcome window `(T0, T_end]`. Repository history clones at
`/tmp/scout-evidence/repos/` (re-clone at the pinned revisions if absent).

## Unit: area

- Eligible files: E1's production-source rules and exclusion globs (`evaluation/inertia-v1/exclusions.json`), with ≥ 10 SLOC.
- **Area** = the file's directory path truncated to at most 2 components below the repository's
  source root. Source roots per repository are recorded in `areas.json` **before** outcomes are computed.
  Areas with fewer than 3 eligible files at T0 are merged into their parent directory once; if still
  below 3, they are excluded.
- Include an area when it exists at T−1 and T0, and has ≥ 5 *changes* in the outcome window.
- A **change** is a non-merge commit touching ≥ 1 eligible file in the area and ≤ 50 files in total
  (E1's bulk rule). A change is counted once per area it touches.
- Secondary analysis: the same at file level (unit = file, ≥ 5 changes).

## Outcomes (outcome window, per area)

Primary (H1):
- **O1 churn per change** = `log1p(median(lines added + removed in the area's files per change))`.

Secondary (H1):
- **O2 change amplification** = `log1p(median(eligible files touched per change, whole repo))`.
- **O3 follow-up rate** = share of changes followed within 7 days by another non-merge commit touching
  at least one of the same files (an incomplete or corrected change).
- **O4 PR effort** (GitHub API, merged PRs whose merge or squash commit is a change in the area):
  `log1p(median commits per PR)`, `log1p(median review comments + review submissions per PR)`,
  `log1p(median hours from PR creation to merge)`. Areas with < 5 linked PRs are `unmeasured` for O4.

H2 (predictability):
- **O5 dispersion** = `log1p(p90 / median)` of per-change churn, and of PR hours-to-merge where O4 is measured.

## Predictors (T0 snapshot and prior window only)

- **Baseline:** `log1p(area SLOC at T0)`, `log1p(prior changes)`, `log1p(prior distinct authors)`,
  `log1p(file count)`, median file age, **and the prior-window value of the same outcome**
  (lagged outcome; for O4 computed on prior-window PRs, `unmeasured` ⇒ that outcome's model excludes the area).
- **Quality level** (added to baseline): area erosion
  `1 − Σ_{CC>10} CC·√SLOC / Σ CC·√SLOC`; `log1p(functions with CC > 10)`; `clone_pct` (jscpd, T0 snapshot);
  hotspot = `Σ_files CC mass × log1p(prior commits to file)`, area-normalised by SLOC.
- **Degradation** (added to baseline + level): Δerosion, Δ`log1p(functions CC>10)`, Δ`clone_pct`
  between T−1 and T0 snapshots.
- **Coupling (H4):** share of the area's prior-window changes that also touch an eligible file in another area.
  (Import-graph filtering of "explained" coupling is out of scope: no uniform 7-language import graph.)
- **Ownership:** top author's share of the area's prior-window changes.

## Models and validation

Identical to E1: predictors z-scored within repository; ridge regression with inner grouped CV for alpha
(grid 10^-2…10^5); outer leave-one-repository-out; held-out MSE; 2,000 bootstrap resamples of held-out
units stratified by repository. Fitted per outcome. Nested comparisons:
1. baseline vs baseline + level;
2. baseline + level vs baseline + level + degradation;
3. baseline vs baseline + coupling; baseline vs baseline + ownership.
Ablations: baseline + each quality predictor alone.

## Decision rules

- **Support** for a predictor set on an outcome: pooled relative MSE reduction ≥ 5%, 95% interval
  lower bound > 0, and reduction > 0 in at least 5 of the repositories with ≥ 10 included units
  (if fewer than 6 such repositories, at least all but one).
- **H1 supported** if level or degradation has support on O1. Support on O2–O4 only is reported as
  partial support for H1 on that outcome.
- **H2 supported** if level or degradation has support on O5.
- **H4 supported** if coupling has support on O3.
- No support ⇒ the corresponding signals stay descriptive in Scout. Support ⇒ documented as predictive
  for that outcome; still no CI gating without a separate decision.
- A predictor whose favourable direction predicts worse outcomes (interval excluding 0) is reported as
  construct-invalid for that outcome.
- Holm correction across the five outcome families (O1…O5) is reported alongside uncorrected results;
  the decision uses the uncorrected rule for O1 (primary) and Holm-corrected for the others.

## Known threats (reported with results)

Churn and PR metrics are proxies for effort, not hours; squash vs merge workflows differ by repository;
bot commits (Renovate, Dependabot) inflate change counts and are excluded by author name match
`(?i)(bot|renovate|dependabot)` recorded in `areas.json`; directory areas are not Clawpatch features;
single time split; seven repositories.
