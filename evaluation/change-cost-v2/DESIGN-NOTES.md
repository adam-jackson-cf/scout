# E1b-v2 design notes (for innovation-council review; NOT RUN)

Scope: how `PREREGISTRATION.md` and the scripts answer `change-cost-v1/METHOD-CRITIQUE.md` (M), `METHOD-CRITIQUE-challenger.md` (C) and the verified facts.
No predictor–outcome association and no model on real data was computed; real data was used only for collection and marginal distributions (§3) and
for outcome *construction* (Jev labels, no predictor shown). Council round files (`reviews/round-N.md`) are the council's to create.

## 1. Design in one paragraph
Unit = merged pull request (16 repositories, 24-month windows, 7,717 panel PRs in 12 repositories after gates, 480 area clusters). Predictors at the
snapshot `S` the author started from: complexity-burden density Q, degradation D on a fixed stable file set, hidden-coupling exposure. Outcomes: git-hours
active time (primary effort measure), Jev-judged corrective follow-up within 14 days. Task type (Jev, validated) and size, author experience, quarter, repository
(or area) fixed effects as controls. Five tests (T1–T5), one Holm layer, delete-one-cluster jackknife with full refit, SESOI + equivalence logic, MDE-based adequacy,
validation gates, synthetic harness, 67 software tests.

## 2. Critique → fix map
| # | Issue (source) | Disposition | Where |
|---|---|---|---|
| M1/C1 | Outcomes are not effort (O1 median 2 lines in fastapi root; O2 floor: 25/60 areas at floor [memo said 31]; O3 same-author/continued work; O4 hours = backlog) | Fixed: PR-level, Y_active (author-side session time, waiting excluded); follow-up requires creation after parent merge + semantic Jev label + validation; review latency demoted to S1 "coordination proxy"; task-type control; floor measured/gated (G3, per-repo >80 % rule, sensitivity a). Residual: single-commit PRs unobserved (42 %) — stated limitation | Prereg §3.3, §5 |
| M2 | Predictors have no variance (nest 23/26 erosion = 1.0; z ≈ −5 outlier) | Fixed: repo gate on Q>0 / D≠0 shares; 9 repositories added; PR-level touched-file measures. nest (Q>0 11 %), axum 20 %, tokio 21 % still pass at the 10 % gate but carry little leverage — jackknife handles leverage | Prereg §2; §3 table |
| M3/C3 | Baseline worse than grand mean (MSE 0.583 > Var 0.470); y uncentred vs z-scored X | Fixed: repo fixed effects (y effectively centred), within-repo predictors; no cross-repo MSE target; P1 reports `1−MSE(base)/MSE(const)` which must be >0 (harness: 0.09–0.22) | Prereg §4 |
| M4 | Ridge alpha at grid ceiling → intercept-only | Fixed: no ridge/tuning on effects (OLS; logit ridge 1e-2 only against separation) | analysis.py |
| M4 | Infeasible repo-count rule (3 big repos, P=0.5) | Fixed: rule removed; per-repo estimates and leave-one-repo-out are descriptive sensitivities | Prereg §8 |
| M5/C5 | No MDE; O1 CI excluded 5 % (tight null) | Fixed: power.py MDE80 vs SESOI, adequacy gate, equivalence verdict so a tight null is reportable only if adequate | §7 here |
| M6 | Bonferroni stacked on Holm | Fixed: one Holm layer over T1–T5; others exploratory | Prereg §4 |
| M7/C4 | Selection on ≥5 outcome-window changes / ≥5 linked PRs | Fixed: every merged PR; no activity count rule; remaining selection (a change was made) stated as threat | Prereg §6 |
| M8/C4 | Lagged outcome control | Not used; area FE (T2/T4) absorb time-invariant area effects; DAG written; T1/T3 labelled associational | Prereg §6 |
| M9a/C2 | Deletions counted as 0 churn | Fixed: GitHub additions+deletions; DELETED files enter F_S; unit test | collect.py, tests.py |
| M9b/C2 | Negative time deltas admitted as follow-ups | Fixed: `is_follow` (created ≥ parent merge, merged ≤ +14 d); 5 tests; neg_duration share = 0 | follow.py |
| M9c/C2 | Right-censoring, no complete-observation rule | Fixed: parents merged <14 d before T_end have no outcome (2.1 %) | follow.py |
| M9d | clone physical lines vs NLOC | Justified out of scope: clone predictor dropped (jscpd per base commit infeasible); if reintroduced it must use one line basis | Prereg §1 |
| M9e/C4 | Degradation compared different file populations | Fixed: D on stable set, same files and same paths at both ends | collect.py, tests.py |
| M10/C3 | Bootstrap without refit; cochanged areas treated independent | Fixed: area-cluster jackknife refitting everything; author-cluster sensitivity; harness shows coverage 0.93–0.98 under nulls (pairs bootstrap was ~20 % too narrow and was rejected) | analysis.py |
| M11/C5 | Regex fix classifier precision 0.52, single labeller, differential noise | Regex not used; Jev labels + two blinded independent annotators + third; gates G4/G5 with thresholds; differential error listed as threat | validate.py, Prereg §5 |
| Minor | hotspot collinear; `erosion` naming; bot regex matches "bot"; PR linkage misses rebase merges | hotspot dropped; Q named burden density (higher = worse); exact Bot type/login list (test); PR's own files/commits from the API, no linkage | |
| C2 | PR-window leakage / prior-window contamination | No windows-by-T0 design; membership by merge time; predictors from S and history ≤ S only | Prereg §2 |
| C1 | Hidden coupling includes import-explained coupling | Textual-reference proxy for static dependency; T5 holds total absent partners fixed. Proxy validated by gate G7 before T5 is decided (G7 thresholds accepted in council round 2, open question 3 closed) | collect.py, validate.py |
| C3 | Singleton junit5 repo | junit5, fastapi, eShopOnWeb, fastify fail the gate | repo-gate.json |
| §5 designs | A (PR panel) adopted; B (SlopCodeBench) out of scope: separate causal study; C (re-analysis) rejected, inherits M1/M2 | | |
| §6 / C Scout §8 | `docs/scout-design.md` §8 text | Out of scope; design rules require their own council review | |

## 3. Construct checks (distributions only)
Funnel (`funnel.json`; `nondev` = Jev class; panel PRs before the repo gate) and gate (`repo-gate.json`):

| repo | T_end | fetched | bot | bulk/many | no eligible file | no snapshot | no pre-existing/measurable file | nondev | panel PRs | gate |
|---|---|---|---|---|---|---|---|---|---|---|
| fastapi | 2023-03-18 | 566 | 56 | 8 | 396 | 0 | 8 | 21 | 77 | FAIL (n<150, areas 7) |
| webpack | 2023-04-19 | 651 | 184 | 8 | 89 | 3 | 8 | 13 | 346 | pass |
| nest | 2023-04-05 | 2211 | 1834 | 8 | 91 | 0 | 11 | 15 | 252 | pass |
| caddy | 2023-02-14 | 418 | 4 | 1 | 52 | 3 | 4 | 30 | 324 | pass |
| axum | 2023-03-22 | 932 | 0 | 14 | 337 | 30 | 11 | 113 | 427 | pass |
| junit5 | 2023-01-10 | 146 | 9 | 2 | 58 | 0 | 1 | 10 | 66 | FAIL |
| eShopOnWeb | 2025-01-13 | 69 | 46 | 0 | 17 | 0 | 0 | 1 | 5 | FAIL |
| django | 2025-10-01 | 1538 | 0 | 14 | 785 | 4 | 1 | 47 | 687 | pass; Y_active dropped (87 % floor), stays in T5 |
| pandas | 2025-10-01 | 3589 | 54 | 21 | 1304 | 0 | 169 | 835 | 1206 | pass |
| scikit-learn | 2025-10-01 | 2452 | 12 | 46 | 1218 | 3 | 7 | 511 | 655 | pass |
| vite | 2025-10-01 | 2014 | 198 | 8 | 656 | 46 | 7 | 113 | 986 | pass |
| fastify | 2025-10-01 | 675 | 63 | 3 | 465 | 0 | 1 | 16 | 127 | FAIL (n, 3 areas) |
| prometheus | 2025-10-01 | 2229 | 581 | 31 | 490 | 0 | 6 | 161 | 960 | pass |
| terraform | 2025-10-01 | 1924 | 320 | 25 | 700 | 12 | 6 | 133 | 728 | pass |
| jellyfin | 2025-10-01 | 1404 | 378 | 16 | 149 | 0 | 1 | 16 | 844 | pass |
| tokio | 2025-10-01 | 659 | 0 | 8 | 171 | 0 | 5 | 173 | 302 | pass |

G1 snapshot coverage ≥ 94.8 % in every repository (axum lowest: 581 reached, 30 without snapshot; `gates.json`, denominators corrected in round 1). Overall (panel 7,717; `constructs.json`):
- Task type (Jev): fix 3,958, feature 1,688, refactor 1,616, other_dev 455 (nondev 2,208 excluded). 10,200 type calls, 100 % ok.
- Q: 75 % > 0, within-repo SD 0.65 (per repo: 11 % nest … 91 % pandas). D measured 6,777, 78 % ≠ 0, SD 0.22. Q area ICC 0.41, D area ICC 0.20.
- Hidden coupling: 24.8 % of PRs > 0 (`hidden_zero_share` 0.752; p95 = 4); 44 % have ≥ 1 absent partner.
- Active hours (all 7,717 panel rows, raw hours): median 1.0 h, p75 2.3 h, p95 8.4 h; single-commit 42 %; negative durations 0. `Y_active` (log hours, n = 7,030 non-missing after django's mask): median 0 (= 1.0 h), p75 0.933 (≈ 2.5 h), p95 2.187 (≈ 8.9 h); floor share 38.2 % (G3 pass; django's 87 % floor makes it missing for T1–T4); within-repo SD 0.94 log; area ICC 0.006.
- Review latency (S1): median 22.5 h, p95 1,627 h (SD 2.3 log) — heavy tail, confirms it is a queueing proxy. Reviewed share 93 %. Rework commits: 0 for >50 %.
- Y_incomplete: 60,342 Jev pair calls (3 reps) all ok, resolved model `typesafe/jev-1.14-20261014`, rep-spread mean 0.011 (reps nearly redundant). 11.3 % of 7,552 non-censored parents positive (explicit-reference variant 6.9 %); by repo 5 %–20 %. 2.1 % censored.
Other: all snapshots resolved via parent of first commit; 2 PRs lack author history.

## 4. Decision log (what changed during design and why)
1. Regex task typing abandoned: left 45 % "other", mislabelled backports/docs (marginal title inspection) → Jev 5-class labeller + validation gate.
2. Primary outcome: cycle time rejected (median 24 h, 90th pct ≈ 53 days, backlog); git-hours chosen; 44 % floor found → gates, per-repo rule, sensitivity.
3. Pairs bootstrap → jackknife (harness coverage 0.85 on nulls vs 0.93–0.98).
4. Logit blow-ups in replicates → ridge 1e-2 + step cap. Permutation check first leaked through fixed points → cyclic derangement. Reference estimand averaged over 8 populations (one draw was noisy).
5. T5 opportunity confound (more absent partners → more candidate follow-ups) → control for log1p(absent partners); harness null scenario covers it.
6. T1/T3 level tests are exposed to area-level confounding of Q (harness `null_area_confounded`: T1 rejects 96 %; T3, whose residual-spread outcome the planted level shift does not reach, rejects 0.035) → H1/H2 defined on the within-area degradation tests T2/T4; T1/T3 associational.
7. Repo gates (n, variance, areas, floor) were set after seeing predictor/cluster/floor marginals, never associations.
8. Jev key: `~/.omp/jev-lab/fastino-bench` does not exist here. Round 1: the run now requires an explicit `OPENROUTER_API_KEY` environment variable (presence-checked, never printed); no secret store or project is referenced; no Jev/annotator call has been made since design review.
9. Jev labels (pairs, types) were generated before validation (they are outcome construction); gates must pass before any model.

## 5. Out of scope, justified
Causality (separate controlled study); agent magnification; clone/hotspot predictors (M9d, collinearity); Scout §8 edits (design review rule).

## 6. Synthetic harness (`harness.py`, 200 simulations per scenario, full pipeline; `harness-results.json`)
Synthetic panels: 8 repos, uneven area clusters, repo mean shifts, zero-inflated predictors, area random effects, covariate confounding. Reference estimand = mean over 8 large populations.
- Planted-null scenarios (null, null_floor, null_confounded, permutation): unadjusted rejection 2–7 % per test, 95 % CI coverage 0.930–0.980, any-test Holm rejection 4.0 %, 5.5 %, 4.0 %, 7.5 % (MC SE ≈ 1.5–1.9 %; permutation slightly high but within ~1.4 SE).
- Planted effects (recover): estimates near the reference (bias ≤ 0.02 for T1–T5), coverage 0.90–0.96, detection at harness N≈2–3 k: T2 0.95, T3 0.97, T4 1.00, T5 0.82, T1 0.39 (low because synthetic area variance is far larger than real, ICC 0.006). Floored outcome: still recovered (coverage 0.82–0.96; T1 0.85, T3 0.82 mildly low).
- null_area_confounded: T1 rejects 96 % (expected: documented confounding), T2/T3/T4/T5 do not (0.035–0.065) — the reason H1/H2 are defined on T2/T4.
- P1 sanity: base vs constant 0.09–0.22 > 0 in every scenario; incremental gain ≈ 0 under null, 0.01–0.08 under effects.

## 7. Power / MDE (`power.py`, real design, simulated outcomes planted on each test's own sample/scale, 60 sims for SE/calibration/verification, 200 sims per empirical-grid point; `power-results.json`)
Nuisance from outcome marginals: SD 0.943, ICC 0.006, base rate 0.113, cluster logit SD 0.33; N = 7,717, G = 480. The ANALYTIC MDE80 = (t + z_.8)·mean null SE at α = 0.01 is a normal approximation and is NOT reported as an 80 %-power effect.
The EMPIRICAL MDE80 is the analytic MDE × the smallest multiplier in {1, 1.25, 1.5} whose simulated power (200 sims) is ≥ 0.80. Adequacy: upper MC bound of the empirical MDE80 ≤ SESOI.
| test | SESOI | analytic MDE80 | power at analytic (60 sims, ±95 % MC) | grid power (200 sims) | multiplier | empirical MDE80 | adequate |
|---|---|---|---|---|---|---|---|
| T1 | 0.10 | 0.045 | 0.90 ± 0.08 | 1.0: 0.90 | 1.0 | 0.045 | yes |
| T2 | 0.10 | 0.050 | 0.85 ± 0.09 | 1.0: 0.735; 1.25: 0.94 | 1.25 | 0.063 | yes |
| T3 | 0.10 | 0.034 | 0.68 ± 0.12 | 1.0: 0.74; 1.25: 0.915 | 1.25 | 0.042 | yes |
| T4 | 0.10 | 0.042 | 0.60 ± 0.12 | 1.0: 0.69; 1.25: 0.91 | 1.25 | 0.053 | yes |
| T5 | 0.20 | 0.211 | 0.83 ± 0.09 | 1.0: 0.79; 1.25: 0.96 | 1.25 | 0.263 | **no** |
The analytic MDE overstates power for T2–T5 (earlier note "within MC error" was wrong for T4: 0.60 ± 0.12 and 0.69 on 200 sims). T1–T4 stay adequate with wide margins (empirical MDE80 ≤ 0.063 ≪ 0.10); **T5 is inadequate: it can be `supported`, `detected_below_SESOI`, `construct_invalid`, `inconclusive` or `not_decided` (G5/G7 failure), but never `equivalent`.**
Null rejection at p<.01: 0 for T1–T4, 1.7 % for T5. MDEs are conservative for the continuous outcome (covariates given zero effect) but rely on the very low outcome ICC.

## 8. Software tests
`tests.py`: 67 checks pass (fixture git repo: deletion counted, snapshot choice, stable-population D, trailing window ≥ 365 d, any-via hidden/explained partners and `via_ref`, author history, login-list/type/substring bot rules, bulk;
follow-up timing; eligibility; active hours; Jev type argmax/ties; within-repo standardisation; Holm; area deepening (two fixtures); logit recovery and ridge λ=0 vs 1e-2; `rel_abs_resid`; every `decide` branch incl. G3/G3b
construct-invalid, no-adequacy and T5-undecided; `finalize_panel` NaN rules; `validate.consensus`/`disagree`/`kappa` treat `unparsable` as disagreement (C adjudicates; κ counts it neither as agreement nor in chance agreement); `validate.gates()` on synthetic annotations (unresolved share, IPW recall, G7 pair-to-partner aggregation, partner-level cell sizes, by-language incl. a sampled language with zero decided partners failing); G1 `snapshot_coverage` gate formula;
`annotate.parse`; `annotate.answer` (non-`stop`/empty/wrong-model completions abort), `annotate.request_body` (bounded reasoning for C only) and `annotate.call` (three failed attempts abort; Amendment 1); `estimate()` skip of an empty-sample test; jackknife cluster presence per statistic's own columns (P1 excludes a repository without `y_active`) and `stat_cols`; logit non-convergence raises; `analysis.main` end-to-end on a synthetic panel with synthetic power/validation/gates files (normal, G3b failure, missing G3b)). `annotate.py` network calls and the real G4/G5/G7 are not exercised by tests.

## 9. Open questions
Closed in council round 2: (1) LLM annotators accepted for G4/G5/G7 (decision a; G7 thresholds accepted as point-estimate screens, per-language rule a sanity check); (2) floored primary kept, no Tobit/two-part (decision b; RESULTS presents T2 and T4 together); (3) G7 thresholds accepted (decision a); (4) SESOIs accepted, T5 stays inadequate and never `equivalent` (decision c).
Still open (not decidable before the run):
5. T1/T3 stay in the Holm family although associational.
6. Mixed eras (2021–23 vs 2023–25 windows): quarter FE plus the cohort split in sensitivity (d); not resolved structurally.
7. nest/axum/tokio contribute little predictor variance; junit5/fastapi/eShopOnWeb/fastify lost to the gate.
8. `.codegraph` untouched; raw GraphQL cache (`raw/q`) kept for reproducibility.

## 10. SHA-256 (round-10 sign-off set plus Amendment 1 rev 2 and the run-1 validation inputs; recomputed for round 14, 2026-10-21; `shasum -a 256`)
```
7d87913c535b286fe7e2c54419297fdfc29eebfae4a64269e33990963cb49bc1  PREREGISTRATION.md
527bd820fca3c6b66a165b5bd785d306062bf73c7e4a280662ce2e4b1ff4b6c5  AMENDMENT-1.md
d01ffb5168fd92095fab0782287e708584f729d71c30baca1b6fcf65e3544681  requirements.txt
651cc1ba3e95edff8f335ffcc0174b5e39d93e30d8ba71b6aa2e4659b0e428a9  analysis.py
7b7bfe6887eac6715fd637a0ada6bce02e80645ba45309762b7077471a3f5a8e  annotate.py
f2e41c96c6868c81a19163abb5d079d48e99f1129ccbd42863c1af9a72461e71  build_panel.py
6ff799d402a261f295b810e3cea708d99e1fc4d00f6b0b36a3d4853ccd6eb087  collect.py
0ec6de5e41c69ab261c167fa172a10cd2f1f88794a2212afcd810fda34a85baa  common.py
80482f25a00c6214219e6674b432e28f78e1227c786f7c49bbead691a1af90d0  constructs.py
8db9613116fd62e803883e736cd8dfe55898a2947a7e954e89b5654e6160416a  fetch_prs.py
b2187dcd2b570b52d8891c18fe55cc978d128833120784cda3cccfbde93193e8  finalize_panel.py
e258ad83b96c066f00b137aa80426d97ac401f82f05e657570ccd1dd9bd48dcb  follow.py
800e84bb7c0f3eaf494f7e400aa4ac8b324ab240f2fd59107a759dd2b42716d5  gates.py
233b91a184fa6a788de4a2e97216b465e4805cd5a91b1d9e39886b07ea609016  harness.py
d94dfa090669c898766970b967b689cca1faddf30ab5e48c99099468b660ab1c  jev_label.py
d9accb9d419c4417998274db99c70fbc327331795de3ee64bba2eb70322c4bcb  jev_type.py
6ac3d72083a6cd2642155048f999da503c571868823b360e20f5888c48788506  power.py
687665d7946bce66a5b90d644f04f79ad5b34a64df895eba6993c174b49816c9  sensitivity.py
622127eec69c1a68015afdc27b646a10b0ee8d28715330089c4967904b7a2e08  tests.py
f8643b2d10ac1487a84afad00fcaac7cdfeec03a6c9e7fa4193e7fc0084c526d  validate.py
0b54289a40f64b8fa49183afbd5fb23651a5a4a7c7d1f490cf559fc368748198  prompts.json
0ce59a217d22b8460ed6f5d023a7dcd55aae5343621a58639e350f45a2fc318f  harness-results.json
2c924dcb939c725923238629321f891d0b49d5665b0cfcc78d779b523cd35f2b  power-results.json
2b1843962c97a12739d067a322e1904148bd6cc3eada5a3151ba4551c4010509  gates.json
683258878ea9694c6449089b2a76b2f4933458cb4701f0d1e94203becb170438  constructs.json
9042f9746a183f14220478572c8bf7a9dbb4665aee480315733af55ea1b584e2  funnel.json
9fe21ce7bbc67b681183a2ef4b2915490560c7fa18cb9643f6d1024abb5b029d  repo-gate.json
66ae08f7636c4681caf1dfa4308277d2f625e17b0da7da12dc1610bd3a8acc26  panel-final.csv.gz
48e0e7e30583d4f926eff153275607f698dadbcf1990092b679ae6034c2e5626  panel.csv.gz
4712eea7490aec3eae915371f97f5f0754bfc8de1c378e3db038b589493752b7  labels.jsonl
7b8eb164c5dc6851e571df78fc15b1a2cb99e732eda77947c94f00b2cede48de  types.jsonl
5005e10b940a361c1a4a93d1e7ec7a631a76a1df371c3dcb214adcaa65508517  pins.json
f9bff032dbd72d3d11fc2f66e4713246d3bcb5654f4d025a48f5debf5ac6dcc5  candidates.jsonl.gz
b98bfc9fd28c74faee03b47f95d9352bed1116a21c8c4da29af3c24d425e399b  data/pr-axum.jsonl.gz
8dfde92a7adb263fd93cc689ec0154e3ffc690bd03756caf8f090fea3990c30f  data/pr-caddy.jsonl.gz
fdcb80f9083e376da51321c67eb1bf129297668f6221061825bde3724e608b80  data/pr-django.jsonl.gz
9263c2bd7309a1b4f7d9d1ea01aa51e4fc51548341c159a73375b07d26f29f23  data/pr-eShopOnWeb.jsonl.gz
f09518be764269dce078ba1eecc42933a0e86b120c63cc7db177e48c99b4071f  data/pr-fastapi.jsonl.gz
818e63998b5d52105fad6c77a1523445bd7b6eda75c6e6bee9a16d0b4083b36f  data/pr-fastify.jsonl.gz
c85a72de276a0fd9b872ee03a8e8b4d21099e93affc7fa5e966e747cb6e07cb7  data/pr-jellyfin.jsonl.gz
4929d7ad85ebfa13367c82a15130822eee528c1d150b99b74650c3e6882231c4  data/pr-junit5.jsonl.gz
ce5b8f75e9d836a83227823427d41dcf4320169f521fef0868fee453a71ac314  data/pr-nest.jsonl.gz
7bf9398697cd0677873b9b7ed66a2279147250ac84fac3551401aa1566795e89  data/pr-pandas.jsonl.gz
86da8e5c155fb222804fc8a141042fcd32f341e91100536eda93093036496aba  data/pr-prometheus.jsonl.gz
7b031b915247df5c7a27d5a8fc706860d4d1ce394b9b9a75ab49b4cf78c3ec11  data/pr-scikit-learn.jsonl.gz
ef01df61becab28d870300e9ee76e5d83d996f3c48a6d9c673b36102bb3ea155  data/pr-terraform.jsonl.gz
1b81955a1d47af36b566ce1c9d8555c011ee6af2c68c38007dccab149b6f2773  data/pr-tokio.jsonl.gz
ce809283abad6fb8e1d9da414e5e20f3c4072098dbd987197a16832e90267ea3  data/pr-vite.jsonl.gz
230002dc4cc7ee2deb24a85a1a0ffb7325cde274de55ebe516d93c4f3173968c  data/pr-webpack.jsonl.gz
43673b35c4dc26b08e703fd83bf807825a0a52630b1d5f59da6b53b3da1a757f  validation/RUBRIC.json
8bb6626ac4f37c9ddf850a4e15844e050e106a5d128efe9d6f3b20babc7acda0  validation/pair-strata.json
5f8cbc844ef0d9ac779050d8bffcc148796b0c3fea88cd5834a199f0b006ad5f  validation/type-blinded.jsonl
3c7858fee012aed7a9cf0dfaf3dee02aad6090f1a65311256407d25b32849f74  validation/pair-blinded.jsonl
81a444c8445c5f1581628a164642ba96f7cc0e3071dc562dc7f21b3bd4aef63d  validation/dep-blinded.jsonl
e1df8d2d3b17d6b69f005ada26ed35ec2be3e85ae23a2f1eedb00d5a9ad5cdc2  validation/type-sample.jsonl
cfd8c5f08a64c64e73fc6412e345ff09315fe9266066079a05c0c6f44533a2ba  validation/pair-sample.jsonl
f8ca34878df6ccfd9e4ae321b65f0afd4f17a0f66e1a0c484c980ffb12e02c2a  validation/dep-sample.jsonl
20d468895e1e1b10ab46205f3f731fbb50d71f57d6e125749ea6d788e97db0a6  validation/annotations-A.jsonl
10a4fd8d7c503845f0b781edbeebe6473de5ad7a8438264d480d7a0deee11dc0  validation/annotations-B.jsonl
cfa854737f52febf60858393706519776c0e4d3fb1919b1fc49dc23625b54297  validation/run-1-failed/SHA256SUMS
```
`gates.json` is regenerated at run time with `gates.py --g3b` (adds G3b; `analysis.py` refuses a gates file without it), so its digest changes then by design.
