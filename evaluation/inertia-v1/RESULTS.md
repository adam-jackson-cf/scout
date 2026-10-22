# E1 — Inertia signal predictive validity (inertia-v1): results

Design: `PREREGISTRATION.md` (unchanged). Every number below is computed by the scripts in this directory
from the pinned repositories; raw outputs are `features.csv` (1,732 files), `model-results.json`,
`classifier-metrics.json`, `classifier-labels.jsonl`, `meta.json`, `exclusions.json`.

## Decision (preregistered rules)

**No predictive support for the quality set, and for none of the five quality predictors individually.**
Under the preregistration the quality signals stay **descriptive** in Scout: they may be reported and trended,
but not used as CI gates or severity inputs. No predictor is construct-invalid.

| Quality set | pooled relative MSE reduction | 95% CI (2,000 bootstrap, stratified by repo) | repos improved | rule (≥5%, lower>0, ≥5/7) |
|---|---|---|---|---|
| baseline + {erosion, cc_max, log n_high_cc, clone_pct, rework} | **+1.18%** (MSE 0.18025 → 0.17813) | **[−0.62%, +3.24%]** | 6 / 7 | **fails** (size and CI lower bound) |

Secondary (Y_any, logistic ridge, LORO): pooled AUROC baseline 0.642 → baseline+quality 0.623,
Δ = −0.019, 95% CI [−0.027, −0.012]. Quality signals do **not** improve, and pooled even slightly reduce, discrimination of
"touched by any fix commit" (mean of per-repo AUROC: 0.783 → 0.777).

## Dataset

Targets/revisions from `evaluation/reference-v1.json`. `T_end` = pinned revision; `T0` = last first-parent commit with
committer time ≤ `T_end − 365 d`. Unit = production source file (preregistered extensions) at T0 with lizard NLOC ≥ 10 that
survives to `T_end` (rename-followed). Exclusion globs per repo: `exclusions.json` (written before any outcome was computed).

| repo | T0 | T_end | candidate files at T0 | ≥10 SLOC | analysed (survive) | eligible commits, outcome window | regex-fix commits, outcome window | eligible commits, prior window | regex-fix, prior window | Σ fix touches on analysed files | files with ≥1 fix | median SLOC |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| fastapi | 2022-03-18 | 2023-03-18 | 43 | 24 | 23 | 748 | 65 | 486 | 29 | 18 | 9 | 116 |
| webpack | 2022-04-19 | 2023-04-19 | 549 | 529 | 529 | 380 | 151 | 1012 | 225 | 107 | 51 | 63 |
| nest | 2022-04-05 | 2023-04-05 | 504 | 308 | 304 | 1590 | 219 | 1312 | 379 | 200 | 87 | 35 |
| caddy | 2022-02-07 | 2023-02-14 | 139 | 136 | 135 | 364 | 128 | 246 | 93 | 177 | 72 | 135 |
| axum | 2022-03-21 | 2023-03-22 | 66 | 61 | 58 | 457 | 103 | 620 | 183 | 107 | 36 | 128 |
| junit5 | 2022-01-06 | 2023-01-10 | 651 | 536 | 536 | 343 | 48 | 547 | 73 | 32 | 27 | 35 |
| eShopOnWeb | 2023-12-28 | 2025-01-13 | 201 | 147 | 147 | 17 | 0 | 97 | 11 | 0 | 0 | 18 |
| **total** | | | 2,153 | 1,741 | **1,732** | 3,899 | 714 | 4,320 | 993 | 641 | | |

"Eligible" = non-merge, ≤50 files. Prior-window counts here are by committer time in `(T0−365 d, T0]`.
Outcome rate (Y_any share): fastapi 0.39, webpack 0.10, nest 0.29, caddy 0.53, axum 0.62, junit5 0.05, **eShopOnWeb 0.00**
(only 17 eligible commits in its outcome window, none fix-matching, so Y is constant 0 there; it contributes MSE, but its AUROC is undefined).
Mean Y by repo: 0.39, 0.095, 0.30, 0.59, 0.77, 0.038, 0.0 (order as in table).

## Commit classifier validation (regex frozen per prereg)

Seeded (20261016) stratified sample of 100 eligible commits (14–15 per repo; `classifier-sample.jsonl`). I labelled each from
message + diff stat in `classifier-sample-view.txt`, which does not display the regex verdict (single labeller; the labeller
also wrote the pipeline, so independence is procedural, not statistical). Rubric (`classifier_eval.py` header): a fix repairs a
defect (wrong/broken behaviour, crash, leak, deadlock, regression, or a revert because something broke) in production or test code;
dependency bumps, CI/config, docs/typo edits, releases, features, refactors, new tests are **not** fixes. One-line rationale per
commit: `classifier-labels.jsonl`.

| scope | n | true fixes | TP | FP | FN | precision | recall |
|---|---|---|---|---|---|---|---|
| all 100 sampled commits | 100 | 12 | 11 | 10 | 1 | **0.524** | **0.917** |
| subset touching ≥1 in-scope source file | 41 | 8 | 7 | 4 | 1 | 0.636 | 0.875 |

False positives (10): `fix(deps): …` ×2 (Renovate dependency bumps), docs/typo "Fix …" ×3, "Remove outdated documentation",
release `axum-debug: Version 0.2.2` (changelog mentions regression), `fix node.js because of v8 bug` (CI pin), `Add Field::chunk`
(body: "fixes not being able to stream"), docs "Fix seed property name documentation". False negative (1): "core: Initialize
logging before admin" (deadlock fix with no keyword). Precision is poor mainly because of Renovate `fix(deps)` titles and
docs-typo commits; with a 12-positive sample the intervals are wide ([INFERENCE]: precision 95% CI roughly 0.32–0.72 by Wilson on 11/21).
The regex was not changed after validation, per prereg. Effect on this study: Y is noisy (label noise in the outcome biases
effect sizes toward zero), especially for nest whose history is dominated by `fix(deps)` bot commits ([INFERENCE] those commits mostly touch excluded `sample/` or lockfiles, not analysed source).

## Main model

Ridge (alpha from inner leave-one-repository-out CV over the 6 training repos; grid 10^-2…10^5, 15 points — grid not fixed by
prereg), predictors z-scored within repo, outer LORO. Chosen alpha (all outer folds) was 316–1000 for baseline and 1000 for
baseline+quality, i.e. strongly shrunk; baseline z-coefficients are all ≤ 0.07 (log SLOC 0.072, prior fix 0.033, prior commits 0.026).
Baseline pooled held-out MSE 0.18025; baseline+quality 0.17813.

Per-repo relative MSE reduction of baseline+quality over baseline:

| fastapi | webpack | nest | caddy | axum | junit5 | eShopOnWeb |
|---|---|---|---|---|---|---|
| +6.4% | +3.9% | +0.7% | +1.1% | +1.3% | **−3.0%** | +1.9% |

AUROC for Y_any by repo (baseline → +quality): fastapi 0.778 → 0.802; webpack 0.753 → 0.737; nest 0.799 → 0.784; caddy 0.796 → 0.794;
axum 0.745 → 0.739; junit5 0.826 → 0.809; eShopOnWeb undefined (no positives).

Coefficients in the full baseline+quality ridge (alpha 1000, z-scored; 95% bootstrap CI refitting on resampled files within repo):
erosion −0.0007 [−0.015, 0.012]; cc_max +0.022 [0.006, 0.039]; log n_high_cc +0.035 [0.018, 0.054]; clone_pct −0.004 [−0.013, 0.005];
rework −0.011 [−0.024, 0.003].

## Ablations (baseline + one quality predictor), LORO

| predictor | pooled rel. MSE reduction | 95% CI | repos improved (of 7) | support (≥5%, lower>0, ≥5/7) | coefficient [95% CI] (alpha 1000) | expected sign | construct-invalid |
|---|---|---|---|---|---|---|---|
| erosion (as defined: 1 − Σ_{CC>10}…/Σ…) | +0.98% | [0.07%, 1.97%] | 6 | no | −0.026 [−0.048, −0.008] | − | no |
| cc_max | +0.74% | [−0.32%, 1.85%] | 4 | no | +0.035 [0.016, 0.056] | + | no |
| log1p(n functions CC>10) | **+1.91%** | [0.39%, 3.68%] | 6 | **no** (<5%) | +0.043 [0.022, 0.066] | + | no |
| clone_pct | +0.49% | [−0.03%, 0.99%] | 4 | no | −0.004 [−0.013, 0.006] | + | no (CI includes 0) |
| rework | −0.10% | [−0.49%, 0.28%] | 3 | no | −0.012 [−0.025, 0.003] | + | no (CI includes 0) |

Per-repo ablation reductions are in `model-results.json` (`ablations[].per_repo`).

Direction convention [INFERENCE]: the preregistered erosion formula as written is high when little complexity mass sits in CC>10
functions, so "favourable" = high and the expected sign of its coefficient is negative; the observed negative coefficient therefore
matches the intended direction. Under the opposite reading (erosion = share *in* complex functions) it would be the same
predictor with flipped sign and the same MSE numbers, and still no support.

Descriptive within-repo Spearman with Y (mean over the 6 repos with variation in Y): log SLOC +0.44, cc_max +0.35,
n_high_cc +0.31, erosion −0.28, log prior fix +0.24, rework +0.12, clone_pct +0.09. The raw complexity correlations are real but
largely carried by file size/history already in the baseline, which is why the incremental MSE gain is ≈1–2%.

## Threats (preregistered list plus observed)

- Commit-message classifier error: precision 0.52 / recall 0.92 on 100 hand-labelled commits (above); Y includes ~half false
  fix labels in expectation overall (fewer on source-touching commits, 0.64).
- Fix-proneness confounded with usage/centrality; no centrality control was available.
- Single time split per repo, one-year windows; repos are large-to-small and Y is very sparse in junit5 (5% of files) and absent in eShopOnWeb.
- Survivorship: files deleted before T_end excluded (survival was 95–100% of ≥10-SLOC files, so small effect here).
- Only 7 repos, 4 of them with ≤ 150 analysed files at the ≥10 SLOC cut (fastapi has only 23 files); the "≥5/7 repos" criterion is noisy.
- Rust inline `#[cfg(test)]` modules (axum) inflate SLOC/CC of production files; Go/Java/C# generated code not covered by globs may remain.
- Ridge alpha is heavily shrunk (1000 on z-scored features, n≈1,500): partly a consequence of Y's heavy-tailed, low-signal structure;
  an unpenalised fit was not run (not in prereg).
- Single labeller for the classifier validation and the same agent built the pipeline.

## Deviations from the preregistration / environment

1. **Repos were not present at start.** `/tmp/scout-evidence/repos` and `targets.tsv` did not exist when I started. I cloned the seven
   repos myself (`git clone --no-checkout` from github.com, full history) and checked out the pinned revisions from
   `reference-v1.json` detached; full histories were verified (e.g. fastapi 2,748, webpack 15,615, nest 13,157 commits reachable).
2. `T0` computed with `git rev-list --first-parent --before=<T_end ts − 365 d>` (committer date).
3. Windows: prior window = commits reachable from T0 with committer time in `(T0−365 d, T0]`; outcome window = `T0..T_end`
   by ancestry (not by date). Walks use `--topo-order`; rename tracking uses `git log -M` name-status along that walk
   (backwards for history, forwards for outcomes). Copies are not followed; a pure rename (R100) is not counted as a touch.
4. Bulk (>50 files) and merge commits are excluded from **all** counts (prior commits, prior fixes, lines, authors, outcome) but
   still processed for rename continuity. File age uses the earliest commit touching the file (including bulk).
5. SLOC = lizard NLOC (comment/blank-free). Function CC/NLOC from lizard 1.21.x Python API (`uv run --with lizard`); `.mjs/.cjs/.tsx` are handled by lizard.
6. `clone_pct` = union of line spans of jscpd (4.0.8, defaults: min-tokens 50, min-lines 5; `--max-lines 10000 --max-size 500kb` raised
   so large files are not skipped) duplicates for the file ÷ lizard NLOC, capped at 1. jscpd ran only on a pruned T0 snapshot containing
   the analysed files (so clones are within the analysed set). jscpd succeeded on all 7 repos (≤12 s each); nothing is missing.
7. `rework`: blame (`git blame -l -s -L …` at the parent) of lines removed (`git diff -U0 -M`) in non-bulk prior-window commits;
   share whose originating commit is in the prior window (matched on the first 39 hex chars of the prior-window commit set).
   Files with no removed lines in the prior window get `rework = 0` (not missing).
8. Models: ridge with intercept (sklearn), logistic L2 with `C = 1/alpha` for Y_any, same alpha grid, inner selection by log-loss.
   Bootstrap = 2,000 resamples (seed 20261016) of held-out files within each repository applied to the fixed LORO predictions (no refit),
   percentile 2.5/97.5. Per-predictor coefficient CIs (used only for the construct-invalid test) refit ridge on resampled files with alpha
   fixed at the median selected alpha (1000), so they are shrunk and conservative.
9. The alpha grid was first 10^-2…10^3; I widened it to 10^-2…10^5 after the selection hit the upper edge; the selected alphas did not
   change (1000 is interior to the final grid). No other outcome-dependent choice was made.
10. Classifier sample: drawn from eligible commits in the union of both windows, stratified ~equal per repo (not proportional).
    Labelling rubric excludes docs/dependency commits as non-defect fixes; a reader using a looser rubric would get higher precision.
11. AUROC for eShopOnWeb is undefined (no positives); the pooled AUROC includes its negatives.

## Reproduce

```
cd evaluation/inertia-v1
uv run --with lizard,pandas python collect.py                       # features.csv, commits.jsonl.gz, meta.json (needs repos in /tmp/scout-evidence/repos, jscpd on PATH)
python classifier_sample.py                                          # classifier-sample.jsonl (+ view for labelling)
uv run --with lizard,pandas python classifier_eval.py                # labels + precision/recall
uv run --with numpy,pandas,scikit-learn,scipy python model.py        # model-results.json
```
