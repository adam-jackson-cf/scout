# E1b-v2 — Does degraded code make later changes costlier, less predictable and less complete? Preregistration

Status: sign-off state and the run record live in `reviews/` (council rounds; the final round names the signed-off digests) and `RESULTS.md` (run); this document does not change after sign-off. Replaces `evaluation/change-cost-v1/` (superseded; see its
`METHOD-CRITIQUE.md`, `METHOD-CRITIQUE-challenger.md`). Governed by `AGENTS.md` experiment rules: council sign-off (zero issues) on
this document **and** the scripts before any predictor–outcome relationship is computed on real data; frozen after sign-off.
Authored before any predictor–outcome relationship was computed on real data. Only marginal distributions of predictors and
outcomes were inspected (recorded in `DESIGN-NOTES.md` and `reviews/round-1-response.md`); no model was fitted to real data. Critique → fix mapping: `DESIGN-NOTES.md`.

## 1. Questions

The user's hypotheses **H1, H2, H4 of `docs/hypothesis-ledger.md`** (sources R-005, R-010, R-011, R-012, R-015, R-017, R-022, R-033, R-034; labelling with Jev: R-008, R-030;
narrative in `docs/inertia-research.md` §4), unchanged in intent, now measured at the level where effort
is observable (the individual change, a merged pull request), within repository, controlling for task type:

- **H1** — a change made to code that has *degraded* (its complexity burden rose over the preceding year) costs more effort than
  the same kind and size of change to code that has not.
- **H1-L (level association, labelled associational)** — changes to code with a higher complexity burden cost more effort,
  beyond controls. Not a causal claim: unmeasured area difficulty confounds it (§6, harness scenario `null_area_confounded`).
- **H2** — a change to degraded code has a less predictable effort (larger deviation from its own forecast). **H2-L** — same for burden level.
- **H4** — a change that leaves behind historically co-changing partner files that no static reference explains (*hidden* coupling) is more
  often followed by a correction/completion of that change than a change that leaves the same *number* of absent partners that a static
  reference does explain (T5 holds total absent partners fixed; see below).

Out of scope (justified in DESIGN-NOTES §5): causal identification (the controlled SlopCodeBench continuation, critique design B,
is a separate study), agent-magnification, changes to `docs/scout-design.md` §8 (needs its own design review), clone-duplication
and hotspot predictors.

## 2. Data

**Repositories** (listed here, fixed before outcomes): fastapi, webpack, nest, caddy, axum, junit5, eShopOnWeb (pinned revisions of
`evaluation/reference-v1.json`) and django, pandas, scikit-learn, vite, fastify, prometheus, terraform, jellyfin, tokio
(pinned at the last first-parent commit before 2025-10-02T00:00Z; SHA recorded in `pins.json`). Full-history clones in
`/tmp/scout-evidence/repos`. A repository enters the analysis iff the **repository gate** (`build_panel.py`, `repo-gate.json`) holds, computed on
predictor and cluster-structure marginals only (never associations): ≥ 150 panel PRs, ≥ 10 % of its PRs with Q > 0, ≥ 10 % of its D-measured PRs
with D ≠ 0, ≥ 10 areas, no area holding > 50 % of its PRs, and G1 snapshot coverage ≥ 0.85 (§5). Additionally a repository whose `Y_active` floor share exceeds 80 % (one commit per PR in
practice) drops out of T1–T4 only (its `Y_active` is set missing; it stays in T5). Excluded repositories are reported with the numbers.

**Window.** PRs *merged* in `(T_end − 730 d, T_end]` per repository (GitHub GraphQL, `fetch_prs.py`). PR membership is by merge
time only; no activity-count inclusion rule.

**Panel population** (funnel counted per repository and reported): merged PR; author not a bot (GraphQL `__typename == "Bot"` or
exact login in the list in `collect.py`; **no substring matching**); ≤ 50 files and ≤ 100 commits; ≥ 1 *eligible* file
(production source by one universal rule, `common.eligible`: ext ∈ {py, js/mjs/cjs, ts/tsx, go, rs, java, cs}, excluding test, doc, example,
vendor, generated, build, tool and `.d.ts` paths) that the PR modifies or deletes and that exists at the snapshot commit
`S` with ≥ 10 NLOC; snapshot `S` found; task type ≠ `nondev`; complete cases on the model's covariates.

**Snapshot `S`** = parent of the PR's first commit if present in the clone, else `baseRefOid`; else the PR is `unmeasured`
(counted). `S` is the code the author started from, independent of the PR's later work. All predictors use `S` and git history
at or before `S`'s committer time. Outcomes use only the PR's own commits/reviews and PRs created after its merge. **No
information dated after `S` enters any predictor and no prior-window outcome is used as a control** (§6 DAG).

**Area (cluster).** Directory of the PR's largest-churn measured file truncated to 2 components; areas holding > 15 % of a
repository's panel PRs are deepened one level at a time (up to 6). Areas are clusters for inference and fixed effects in T2/T4; they are
directory areas, not Clawpatch features.

## 3. Measures

### 3.1 Predictors (per PR; lizard CC and NLOC on the blobs at `S`)
- **Q (complexity-burden density, higher = worse).** `Q = Σ_f hi_f / Σ_f nloc_f` over the measured files `F_S` (modified/deleted
  eligible files existing at `S`, ≥ 10 NLOC), `hi_f = Σ_{functions CC>10} CC·√nloc_fn`. (Renamed from v1 `erosion`, which was a health score.)
- **D (degradation, higher = worse).** On the **stable population** `F_stable ⊆ F_S` = files that also exist, same path, with ≥ 10 NLOC at
  `S_trail` (last first-parent commit at least 365 d before `S`'s time): `D = Σ hi/Σ nloc at S − Σ hi/Σ nloc at S_trail`, **the same
  file set at both ends**. PRs with no stable file have D missing. `Q_stable` = the first term, used as the level control in T2/T4.
- **Hidden-coupling exposure `log_hidden = log1p(n_hidden)`.** For each file in `F_S`, co-change history in `(S_time − 365 d, S_time]`
  over non-merge commits touching ≤ 50 files and eligible paths: partner `g` qualifies if `n(f,g) ≥ 3` and `n(f,g)/n(f) ≥ 0.3`. Among the
  qualifying partners **absent from the PR**, the top 10 per file by confidence `n(f,g)/n(f)` (then `n(f,g)`, then path) are kept; a kept partner existing at `S` counts as *expected-but-absent*; it is
  **hidden** unless `g`'s module stem occurs as a word in `f`'s source at `S` or `f`'s stem in `g`'s (textual-reference proxy for a
  static dependency; no uniform multi-language import graph exists). `n_hidden` = number of hidden absent partners.
- Predictors are standardised **within repository** (subtract repo mean, divide by the pooled within-repo SD) inside the pipeline so
  coefficients are "per +1 within-repo SD" and between-repo differences never enter a slope.

### 3.2 Controls (pre-change characteristics and realised properties of the change)
`log1p(PR additions+deletions)` (deletions counted), `log(eligible files touched)`, share of touched eligible files that are new, `log(Σ nloc of F_S)` and the
task type and calendar quarter of the merge are **realised properties of the change itself** (they are not measured at `S`, may be affected by the code state, and
can mediate part of an effect: H1/H2 are therefore effects at comparable size, task and time; §6). `log1p(prior commits by the first commit's author e-mail before the
PR's first commit)` and maintainer (`authorAssociation` ∈ OWNER/MEMBER/COLLABORATOR) are pre-change. Fixed effects: task type, quarter, and repository
(T1, T3, T5) or area (T2, T4).
**Task type** (`jev_type.py` → `types.jsonl`; Jev choice on title, labels, description and changed-file paths; 1 repetition; argmax of class probabilities; validated by G4):
fix / feature / refactor / other_dev. `nondev` PRs (dependency or version bump, release, documentation- or test-only, CI/build, backport/cherry-pick, generated)
are excluded; PRs whose type call failed are `type_unlabelled` (counted, excluded). A deterministic title/label regex was tried in design review
and abandoned (DESIGN-NOTES §4) because it left ~45 % of PRs as "other" and mislabelled backports and documentation PRs.

### 3.3 Outcomes
- **Y_active (H1, H2 — primary effort measure).** `log(active hours)` by the git-hours session method on the PR's commit **author**
  timestamps: sort, gaps ≤ 4 h add the gap, a longer gap starts a new session; each session gets 0.5 h credit. Review/queue waiting is
  excluded by construction (fixes v1's backlog-latency problem). A single-commit PR gets the floor `log(0.5)`: the work duration is
  unobserved for it (≈ 42 % of panel PRs are single-commit; see DESIGN-NOTES §3 for the floor share over non-missing rows).
  **Preregistered single-commit handling:** `Y_active` is used as a *floored* measure with no further modelling. T1–T4 therefore estimate slopes of a
  left-censored (floored) outcome: a predictor that changes dispersion also moves the floored mean, so **location (H1) and dispersion (H2) are not separable on this
  scale**, and T1/T2 are read as "effect on the floored active-time measure", T3/T4 as "deviation of the floored measure" (the harness `recover_floor` scenario reports
  this projection and is not gated). The floor is bounded by gate G3 and the per-repository rule (§2). Sensitivity (a) (≥ 2 commits) conditions on a post-treatment variable
  and is descriptive only.
- **Y_incomplete (H4).** 1 if any PR merged within 14 d after the parent's merge, **created at or after the parent's merge instant** (concurrent work
  never counts; `follow.is_follow`), non-bot (shared `common.is_bot`: GraphQL Bot type or exact login in the list; same predicate as the panel filter), ≤ 50 files, that shares an
  eligible file with the parent, or touches one of its co-change partners, or names it (`#N`), is judged *corrective* by Jev (§5, G5: P(corrective) ≥ 0.5, mean over the successful repetitions; all 60,342 calls succeeded, 3 repetitions each);
  0 otherwise. Parents merged < 14 d before `T_end` are right-censored and get no outcome (not zero). The labeller sees titles, descriptions and file paths of the two PRs
  only; never the parent's predictors. Question digests are recorded (`prompts.json`, `qhash`).
- **Secondary (no decision weight):** S1 `log1p(hours from first review to merge)` (reviewed PRs; coordination latency proxy, not effort);
  S2 `log1p(commits authored after the first review)` (rework); exploratory outcome `y_incomplete_explicit` (follow-up names the parent and
  its text matches fix/revert/regress/follow-up/missing; deterministic regex rule, no model call).
- **Dispersion (H2)** is defined from the primary outcome's residual, §4.

## 4. Models (the whole pipeline is `analysis.py`; nothing else touches predictor–outcome relationships)

All coefficients are per +1 within-repo SD of the predictor.

| Test | Hypothesis | Model | FE |
|---|---|---|---|
| T1 | H1-L | OLS `Y_active ~ Q + controls` | repo, quarter, task |
| T2 | **H1** | OLS `Y_active ~ D + Q_stable + controls` | **area (absorbed)**, quarter, task |
| T3 | H2-L | stage 2: OLS `u ~ Q + controls`, `u = |e| / repo-mean(|e|)`, `e` = residual of the T1 model | as T1 |
| T4 | **H2** | stage 2: same with `D`, `e` from the T2 model | as T2 |
| T5 | **H4** | logistic `Y_incomplete ~ log_hidden + log1p(absent partners) + Q + controls` (fixed ridge 1e-2 on all but the intercept, never tuned: a numerical stabiliser against separation, step cap 2; `tests.py` checks on a synthetic n = 4,000 panel that λ = 0 vs 1e-2 changes the coefficient by < 2 %) | repo, quarter, task |

T5 controls for `log1p(absent partners)` (hidden + statically explained) because the number of absent partners also sets how many candidate follow-up PRs
exist (opportunity for a labelled follow-up); the T5 coefficient is therefore the extra effect of a partner being *hidden* rather than explained.

- **No lagged-outcome control, no tuning of any effect (the only shrinkage is the fixed ridge above), no outcome-driven selection.** Area fixed effects
  (T2/T4) absorb all time-invariant area difficulty.
- **Uncertainty: delete-one-cluster jackknife (clusters = areas) with full refit** of every step (standardisation, stage 1 + 2, logit) on each
  leave-one-cluster-out sample; `se² = (G_t−1)/G_t Σ(θ₍₋g₎ − θ̄)²` with **G_t = the clusters present in test t's own complete-case sample**; p-values and CIs from t with G_t−1 df.
  A non-finite replicate for a decision test is a hard error (no silent dropping). (A pairs-cluster bootstrap was rejected in harness runs: SE ≈ 20 % too small
  with 100–300 uneven clusters.) Deterministic; no seed.
- **Multiplicity: one layer.** Holm over the *decided* tests among T1–T5 (T5 only if G5/G7 pass; T1–T4 only if G3/G3b pass), two-sided, family α = 0.05. No Bonferroni stacking, no uncorrected
  primary rule. S1, S2, sensitivity analyses and P1 are outside the family and labelled exploratory.
- **Smallest effect of interest (SESOI)**, each justified separately; all are decision-use thresholds, not measured truths, and are open to council challenge:
  T1, T2 `0.10` log-hours per SD (≈ +10.5 % active time per SD of burden): Code Red (R-010) reports +124 % resolution time between low and high code health, a span of roughly
  4 within-repo SDs [INFERENCE], i.e. ≈ 0.20 log/SD; SESOI is half of that. T3, T4 `0.10` relative dispersion per SD (typical absolute deviation 10 % larger per SD): chosen as the
  same relative size as the T1/T2 location threshold, so a flag would be raised for a change in spread of the same proportional magnitude as a change in mean; Code Red's ~9× maximum
  time (R-010) is a tail ratio, not a per-SD slope, and cannot be transferred. T5 `0.20` log-odds per SD (OR ≈ 1.22; ≈ +2.2 percentage points on the 11 % base rate): no published
  effect size for hidden coupling → incomplete change exists (R-015 reports defect associations, R-017 coordination/resolution-time effects, R-022 is conceptual; none reports an effect size for incomplete changes), so this is a decision-use threshold: Scout would act on a signal
  that moves the follow-up probability by about a fifth in relative terms [INFERENCE].
- **Verdict per test** (`analysis.decide`): `supported` — Holm-adjusted p < 0.05, estimate > 0, 95 % CI lower bound > 0 **and** estimate ≥ SESOI;
  `detected_below_SESOI` — significant, positive, estimate < SESOI; `construct_invalid` — significant with the opposite sign;
  `equivalent` — 90 % CI inside ±SESOI **and** the test is adequate per `power-results.json` (upper-MC-bound **empirical** MDE80 ≤ SESOI, §7); else `inconclusive`. A null is
  therefore reportable as "no effect of interest" only where the design could have detected it.
- **Hypothesis verdicts** are the verdict of one test each: **H1 = T2, H2 = T4, H4 = T5**; the level associations H1-L = T1 and H2-L = T3 are reported as
  associations only and are never described as H1/H2 support.
- **P1 (exploratory): temporal incremental predictive gain.** Per repo, first 60 % of merge times train / last 40 % test, predictors standardised
  with train-period statistics, relative MSE reduction of baseline + Q + D over baseline, plus the sanity statistic `1 − MSE(baseline)/MSE(train
  repo mean)` that **must be > 0** (v1's baseline was worse than a constant). Uncertainty by the same jackknife (refit-inclusive).

## 5. Measurement validation gates (all before any model is fitted; G4 failure ⇒ study stops; G5/G7 failure ⇒ T5 `not_decided`; G3/G3b failure ⇒ T1–T4 `outcome_construct_invalid`)

Gates G4, G5 and G7 are **fully scripted and mechanical**: `validate.py sample` draws blinded samples (item ids assigned after shuffling; blinded files carry no
pipeline label, Jev output, stratum, candidate kind or repository); `annotate.py` labels them with **pinned annotator models of a different family from Jev**
(A = `anthropic/claude-sonnet-4.5`, B = `openai/gpt-4.1`, C = `google/gemini-2.5-pro`, temperature 0, one call per item; C labels only items where A and B
disagree; rubric text in `validate.py RUBRIC`, digests in `validation/annotator-prompts.json`; the resolved model must come from the pinned provider or the script
aborts; an unparsable answer is a disagreement: C labels it, even when A and B are both `unparsable`, and κ never counts it as agreement nor in chance agreement); `validate.py gates` computes the metrics and pass/fail into `validation-gates.json`, which `analysis.py` reads —
no human-set flag exists. These annotation calls are a post-sign-off, pre-analysis step (§10). Consensus = A when A = B and parsable, else C; no C label or an unparsable C ⇒ unresolved.
- **G1 Snapshot coverage:** per repository ≥ 85 % of PRs reaching the snapshot step (`ok` + no_snapshot + no_preexisting_file + no_measurable_file) find `S`; else the repository is excluded. **Mechanical:** `build_panel.snapshot_coverage` computes it from the funnel and it is part of the repository gate (`repo-gate.json` `snapshot_coverage`; `pass` requires ≥ 0.85); `gates.py` reports the same numbers.
- **G2 Predictor variance:** the repository gate in §2; also reports overall share Q > 0 and D ≠ 0.
- **G3 Outcome floor:** share of non-missing `Y_active` rows at the floor ≤ 60 % overall; else `Y_active` is declared uninformative and H1/H2 are not tested
  (per-repository floor rule in §2).
  Run-time construct gate **G3b** (a covariate–outcome association, not a predictor of interest): Spearman(`Y_active`, `log1p(churn)`) ≥ 0.15
  after repo demeaning; else the outcome is not tracking work size. **Mechanical:** `analysis.py` requires `--gates gates.json` produced with `gates.py --g3b` (exits if G3b is absent); if `G3.pass_` or `G3b.pass_` is false,
  T1–T4 receive verdict `outcome_construct_invalid`, are removed from the Holm family and H1/H2 are not tested (`decide(..., h12_valid=False)`, tested).
- **G4 Task type (Jev):** 150 PRs, 30 per Jev class; annotator κ(A,B) ≥ 0.5, Jev-vs-consensus accuracy ≥ 0.70, unresolved ≤ 10 %; else the study stops and
  returns to council (`analysis.py` exits).
- **G5 Follow-up labeller (Jev):** 200 candidate pairs (100 Jev-positive, 100 Jev-negative); κ ≥ 0.6; Jev precision ≥ 0.70 and inverse-probability-weighted
  recall ≥ 0.60 against consensus; ≥ 20 consensus positives. **`cannot_tell` policy:** pairs with consensus `cannot_tell` or unresolved are excluded from precision
  and recall and must be ≤ 15 % of the sample, else G5 fails.
- **G7 Hidden-coupling proxy:** 12 partner pairs per cell of (language group: python, js_ts, go, rust, jvm_dotnet) × (proxy-explained, proxy-hidden), shown as the first 150 lines
  of the pair(s) at `S`; annotators judge static dependence (yes / no / cannot_tell; `yes` if ANY listed pair depends). κ ≥ 0.6; proxy-hidden items judged `no` ≥ 80 %; proxy-explained items judged `yes` ≥ 70 %;
  hidden precision ≥ 60 % in every language group (a sanity check only: 12 items per cell give it little power; cell sizes are reported in `validation-gates.json`); `cannot_tell`/unresolved ≤ 15 %. The proxy (`collect.referenced`): partner's stem occurs as a whole word on a non-comment
  line of a touched co-changing file, or the touched file's stem in the partner (generic stems use the directory name and never count); a partner is *explained* if ANY qualifying
  touched file does so (recorded as `via_ref`), else *hidden*. **G7 items are pair-level and blinded:** each annotated item is one (file, partner) pair with an opaque id assigned after a global shuffle, so item structure carries no proxy information; an explained partner contributes its (`via_ref`, partner) pair, a hidden partner each of ≤ 3 qualifying vias drawn by the seeded `rng`; the partner grouping exists only in the unblinded sample file.
  `validate.gates()` aggregates per partner: any `yes` pair ⇒ the partner is `yes` (it depends); otherwise any `cannot_tell`/unresolved pair ⇒ the partner is unresolved; otherwise `no`. κ is on pair-level labels; precision, cell sizes and the unresolved share are partner-level. Every sampled hidden language must have ≥ 1 decided partner and precision ≥ 60 % (a group with none fails the gate). An `unparsable` annotator answer is a disagreement: C adjudicates the pair, and if C gives no parsable label the pair is unresolved.
- **G5/G7 failure ⇒ T5 is not decided**: it is dropped from the Holm family and reported `not_decided`; H4 is then exploratory only (T5 on `y_incomplete_explicit`
  in `sensitivity.py`).
- **G6 Software tests and harness** (`tests.py`, `harness.py`; tolerances coded in `harness.checks`) pass: unbiased recovery of the planted T1/T2/T5 values
  (|bias| ≤ 0.05 / 0.05 / 0.10), CI coverage ≥ 0.88 (planted) and ≥ 0.90 (nulls), any-test Holm rejection ≤ 10 % on null and permuted data, T2/T4 detection ≥ 80 %,
  P1 baseline beats the constant, documented area-confounding behaviour. The floored scenario is reported, not gated (see §3.3).

## 6. Identification statement (temporal DAG)

`area difficulty (u) → {Q, D history, hidden coupling}` and `u → effort`; `Q(S)`, `D(S)`, `hidden(S)` precede the PR; `author experience, task type, size, quarter`
precede or define the work and are controlled; effort and correction occur after. T1/T3 compare PRs *across* areas within a repository
and are therefore confounded by unmeasured area difficulty `u` (demonstrated by harness scenario `null_area_confounded`: T1 rejects, T2 does not);
they are reported as associations ("incremental predictive value beyond controls"). T2/T4 use within-area variation (area FE), removing time-invariant `u`;
they still do not identify causation (time-varying confounding by e.g. a refactoring campaign is possible). Size and files are controlled as
confounders of task scale, which means H1 is the effect on effort *for a change of comparable size*; effects that act through larger changes (change
amplification) are not in T1/T2 and are not claimed. Selection: the unit is a change that was made; code avoided because it is costly is not in the panel
(reported as a threat; v1's additional ≥ 5-change activity selection is removed). H2 uses in-sample residuals; with median area size 5 (`constructs.json`) the area absorption in
stage 2 can shrink residuals (a leverage artefact that would bias T4 toward zero); this is a stated threat, and the harness null and planted scenarios exercise the same estimator.

## 7. Power and design adequacy

`power.py` simulates the full pipeline on the real panel's predictor/covariate/cluster structure with outcomes simulated from outcome marginals only; each effect is planted on
**that test's own complete-case sample and standardisation scale** (the planted-to-estimate ratio is calibrated per test, because T3/T4 plant a log-SD multiplier but estimate a
relative-|residual| slope). It reports each test's **analytic** MDE80 (estimate scale) at the strictest Holm threshold (α = 0.01) with Monte-Carlo bounds, the verification power at it, and an **empirical MDE80**: the analytic MDE × the smallest
multiplier in {1, 1.25, 1.5} whose simulated power (200 simulations) is ≥ 0.80 (none ⇒ inadequate). **Adequacy rule: a test is adequate iff the upper Monte-Carlo bound of its empirical MDE80 (multiplier × (mean null SE + 1.96·SD/√sims)·(t+z)) is ≤ its SESOI.** `analysis.py` requires
`power-results.json`; a test absent from it is inadequate and can never receive an `equivalent` verdict. **Reporting rule (council decision b):** RESULTS presents T2 and T4 together and never states H1 support without the T4 estimate alongside (location and dispersion are not separable on the floored scale, §3.3).

## 8. Preregistered sensitivity analyses (reported, not decision-bearing)
(a) T1–T5 restricted to PRs with ≥ 2 commits — **conditioning on a post-treatment variable**, descriptive only; (b) session gap 2 h and 8 h; (c) leave-one-repository-out;
(d) per-repository estimates and a cohort split (PRs merged before 2024 vs 2024 or later: mixed eras); (e) `y_incomplete_explicit` (deterministic regex rule, no model call) for T5;
(f) inference clustered by author instead of area (`sensitivity.py`). S1 and S2 (Q on review latency / rework) are reported by `analysis.py` itself.

## 9. Known threats (reported with results)
Active time from commit timestamps is a proxy (single-commit PRs unobserved; rebased commits keep author dates); squash vs merge workflows do not affect
author timestamps but commit habits differ by author; complexity density is a coarse quality measure and the only one confirmed per PR at `S`;
hidden-coupling textual-reference proxy misses real static dependencies and flags some real ones; the Jev task-type and follow-up labellers are
imperfect (validated; errors may be differential); directory areas are not features; merged-PR-only population; 16 repositories chosen for activity and
language coverage, not randomly; one pinned window per repository; LLM annotators are not human experts; association, not causation.

## 10. Execution order, artefacts, deviations

Environment: Python 3.13 via `uv`, dependencies pinned in `requirements.txt` (lizard 1.24.1, numpy 2.5.4, pandas 3.0.6, scipy 1.18.1; every command
below is `U = uv run --python 3.13 --with-requirements requirements.txt python`). **Secrets:** the Jev and annotator steps REQUIRE the environment variable
`OPENROUTER_API_KEY` (presence-checked, never printed or stored); no secret store or project is assumed. `pins.json` is an input, not recomputed
(`fetch_prs.py` only writes entries for repositories not yet pinned).

```
cd evaluation/change-cost-v2
$U tests.py                                            # G6 (software tests)
$U harness.py 200 harness-results.json                 # G6 (planted effects / nulls / permutations, tolerances in harness.checks)
$U fetch_prs.py                                        # pins.json (input), raw/prs-*.json.gz   (gh auth)
$U collect.py                                          # data/pr-*.jsonl.gz
$U jev_type.py types.jsonl                             # Jev task types            (needs OPENROUTER_API_KEY)
$U build_panel.py                                      # panel.csv.gz, funnel.json, repo-gate.json
$U follow.py                                           # candidates.jsonl.gz
$U jev_label.py candidates.jsonl.gz labels.jsonl       # Jev follow-up labels      (needs OPENROUTER_API_KEY)
$U finalize_panel.py                                   # panel-final.csv.gz
$U validate.py sample                                  # validation/*-sample.jsonl, *-blinded.jsonl
$U annotate.py                                         # pinned blinded annotators (needs OPENROUTER_API_KEY)
$U validate.py gates                                   # validation-gates.json: G4, G5, G7 (mechanical)
$U constructs.py panel-final.csv.gz > constructs.json  # distributions
$U gates.py --g3b > gates.json                         # G1-G3 (+G3b at run time only)
$U power.py panel-final.csv.gz 60 power-results.json   # MDE / adequacy
$U analysis.py panel-final.csv.gz results.json --power power-results.json --validation validation-gates.json --gates gates.json
$U sensitivity.py panel-final.csv.gz sensitivity.json
```

**Design-review status of the Jev steps:** the Jev task-type and follow-up labels (`types.jsonl`, `labels.jsonl`) were generated during design review because
the harness/power/construct checks need the outcome *construction* (no predictor was shown to Jev; no association computed). They are frozen inputs with
digests; question digests are in `prompts.json`. **No further real Jev or annotator calls are made until sign-off.** Annotation (G4/G5/G7) is a post-sign-off,
pre-analysis step: the scripts and pinned annotators are part of the reviewed implementation, and `analysis.py` refuses to run without `validation-gates.json`.

Frozen artefacts after sign-off: this file and all scripts listed with SHA-256 at the end of `DESIGN-NOTES.md`. Runtime deviations go under `Deviations` in
`RESULTS.md` (the v1 file is superseded and untouched). Everything preregistered is reported with estimates, SEs, 95 % CIs, `unmeasured` counts, gate results and commands.
