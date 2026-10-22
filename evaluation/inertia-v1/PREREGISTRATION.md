# Inertia signal predictive validity on real repositories — preregistration (2026-10-16)

Written before any outcome in this study was computed or inspected. It adapts
SlopCodeBench's `experiment_analysis/011026-predictive-validity-preregistration.md`
from benchmark checkpoints to Git history.

## Question

Do file-level code-quality signals measured at time T0 predict later
defect-fix activity on that file beyond size, history, and ownership controls
available at T0?

## Targets

The seven pinned targets in `evaluation/reference-v1.json` (FastAPI, webpack,
Nest, Caddy, Axum, JUnit 5, eShopOnWeb). `T_end` is each pinned revision.
`T0` is the last first-parent commit at or before `T_end − 365 days`.
Prior window: `(T0 − 365 days, T0]`. Outcome window: `(T0, T_end]`.

## Unit and inclusion

- Unit: one production source file (`.py`, `.js/.mjs/.cjs`, `.ts/.tsx`, `.go`,
  `.rs`, `.java`, `.cs`) present at both T0 and T_end (renames followed with
  Git rename detection).
- Exclude tests, examples, docs, benchmarks, fixtures, vendored, generated and
  type-declaration (`.d.ts`) files; exclusion globs are recorded per repository
  before outcomes are computed.
- Exclude files with fewer than 10 SLOC at T0.

## Commit classification

- Ignore merge commits and commits touching more than 50 files (bulk/format).
- A **fix commit** has a subject or body matching
  `(?i)\b(fix(e[sd]|ing)?|bug|regression|revert(ed|s)?|hotfix|crash(es|ed)?|broken|incorrect|wrong)\b`
  or a subject starting with `Revert`.
- The classifier is validated on a seeded random sample of 100 commits labelled
  independently from message and diff stat; precision and recall are reported.
  The classifier is not changed after validation within this study.

## Outcome (outcome window)

- Primary `Y = log1p(fix commits touching the file)`.
- Secondary `Y_any = 1 if any fix commit touched the file`.

## Predictors (T0 snapshot and prior window only)

- **Baseline:** `log1p(SLOC)`, `log1p(prior commits)`, `log1p(prior fix commits)`,
  `log1p(prior lines added+removed)`, `log1p(file age days)`, `log1p(prior distinct authors)`.
- **Quality** (added to baseline):
  - `erosion = 1 − Σ_{CC>10} CC·√SLOC / Σ CC·√SLOC` over functions (lizard; 1 when no functions);
  - `cc_max`, `log1p(count of functions with CC > 10)`;
  - `clone_pct`: duplicated lines / SLOC (jscpd, per repository, T0 snapshot);
  - `rework`: of lines removed from the file in prior-window commits, the share
    whose originating commit (blame at the parent) is itself inside the prior window.
- Import-cycle architecture mass is excluded: no uniform multi-language import
  graph is available in this study.

## Model and validation

- Predictors z-scored within repository.
- Ridge regression (alpha by inner grouped CV over repositories), fitted
  separately for baseline and baseline+quality; logistic ridge for `Y_any`.
- Outer validation: leave-one-repository-out.
- Metric: MSE on held-out files (primary); AUROC for `Y_any` (secondary).
- Effect: relative MSE reduction of baseline+quality over baseline.
- Uncertainty: 95% percentile interval from 2,000 bootstrap resamples of
  held-out files stratified by repository.
- Ablations: baseline + each quality predictor alone.

## Decision rules

- **Predictive support** for the quality set: pooled relative MSE reduction ≥ 5%,
  interval lower bound > 0, and reduction > 0 in at least 5 of 7 held-out repositories.
- A single predictor has support under the same rule.
- **No support:** the quality signals stay descriptive in Scout; they may be
  reported and trended but not used as CI gates or severity inputs.
- A predictor whose favourable direction predicts more fixes (opposite to its
  definition, interval excluding 0) is reported as construct-invalid for prediction.
- Known threats reported with the result: commit-message classifier error,
  fix-proneness confounded with usage/centrality, single time split, survivorship
  (files deleted before T_end are excluded).
