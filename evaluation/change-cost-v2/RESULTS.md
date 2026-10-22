# E1b-v2 — Results

Run 2026-10-21. Preregistration `PREREGISTRATION.md` (7d87913c…, signed off in `reviews/round-10.md`); Amendment 1 revision 2 (`AMENDMENT-1.md` 527bd820…, signed off in `reviews/round-14.md`). Every number below comes from `results.json`, `validation-gates.json`, `gates.json`, `power-results.json` and `sensitivity.json` in this directory. Everything not decided by the preregistered rule is labelled exploratory.

## 1. Answer

On merged PRs in 12 open-source repositories (7,717 PRs, 480 directory areas), **changes to code whose complexity burden rose over the prior year did not take a meaningfully larger, or less predictable, amount of author active time than comparable changes to code whose burden did not rise.** Both decision tests return `equivalent`: the 90 % CI lies inside the preregistered smallest effect of interest (SESOI), ±0.10 log-hours per within-repo SD (≈ ±10.5 % active time), and the design was adequately powered for that (empirical MDE80 0.063 for T2 and 0.053 for T4).

**H4** (hidden coupling → corrective follow-up) **was not decided.** Two of the measurement-validity gates failed: the Jev follow-up labeller's recall (G5) and the hidden-coupling proxy (G7). H4 is therefore exploratory only.

| Hypothesis | Test | Estimate (per SD) | 95 % CI | 90 % CI | Holm p | Verdict |
|---|---|---|---|---|---|---|
| H1 degradation → effort | T2 | +0.023 log-h (≈ +2.3 %) | [+0.001, +0.045] | [+0.004, +0.041] | 0.165 | **equivalent** (no effect of interest) |
| H2 degradation → unpredictability | T4 | +0.006 | [−0.014, +0.026] | [−0.010, +0.023] | 0.599 | **equivalent** |
| H1-L burden level → effort (associational) | T1 | −0.011 | [−0.032, +0.010] | [−0.029, +0.007] | 0.599 | equivalent |
| H2-L burden level → unpredictability (associational) | T3 | +0.018 | [−0.007, +0.042] | [−0.003, +0.038] | 0.493 | equivalent |
| H4 hidden coupling → follow-up | T5 | −0.024 log-odds | [−0.151, +0.102] | [−0.130, +0.082] | — | **not_decided** (G5, G7 failed); exploratory |

Holm family = T1–T4 (four decided tests; T5 removed under prereg §5). Jackknife clusters: T1/T3 451, T2/T4 417, T5 480. Sample sizes: n_T1 = 7,028, n_T2 = 6,110, n_T5 = 7,550.

How to read T2 (the H1 test). Its point estimate is positive, and the unadjusted 95 % CI just excludes 0 (p = 0.041). It is not significant after Holm correction. Its 90 % CI upper bound is 0.041 (≈ +4.2 % active time per SD of degradation), which is under half the SESOI. A small positive association is compatible with the data; an effect of the size the study set out to detect (≥ 10 % per SD) is not. Per council decision b, T2 and T4 are read together: degradation shows no material effect on either the level or the spread of effort.

## 2. Gates

| Gate | Result | Values |
|---|---|---|
| G1 snapshot coverage ≥ 0.85 | pass, all 16 repositories | lowest axum 0.948 |
| G2 predictor variance (repository gate) | 12 of 16 repositories pass | junit5, fastapi, eShopOnWeb, fastify excluded (as frozen in `repo-gate.json`) |
| G3 outcome floor ≤ 60 % | pass | 38.2 % of 7,030 non-missing `Y_active` |
| G3b `Y_active` vs log churn (construct) | pass | Spearman 0.460 |
| G4 task type (Jev) | **pass** | κ(A,B) 0.771; Jev-vs-consensus accuracy 0.767; unresolved 0.0; n 150 |
| G5 follow-up labeller (Jev) | **fail** | κ 0.715; precision 0.878; **IPW recall 0.422 < 0.60**; 94 consensus positives; unresolved 0.015 |
| G7 hidden-coupling proxy | **fail** | **pair κ 0.514 < 0.6**; **hidden precision 0.630 < 0.80**; explained precision 0.932; per language: go 0.20, jvm_dotnet 0.583 (< 0.60), python 0.636, rust 0.818, js_ts 0.90; unresolved 0.058 |
| G6 software tests and harness | pass | `tests.py` 67 checks; `harness-results.json` (frozen) |

G4 confusion (pipeline → consensus): most errors are `other_dev` PRs that the annotators call `feature` (10) or `refactor` (11), and `fix` called `feature` (6). Task type is a control, so these errors add noise to the controls only.

G5 means the Jev follow-up labeller finds fewer than half of the corrective follow-ups the annotators identify, so `y_incomplete` is undercounted. G7 means the textual-reference proxy often calls a partner "hidden" when the annotators judge that a static dependency exists, so `n_hidden` is contaminated. Each failure on its own invalidates T5 as a test of H4.

## 3. Design adequacy (`power-results.json`, reproduced byte-identically at run time, 2c924dcb…)

| Test | SESOI | Empirical MDE80 (upper MC bound) | Adequate |
|---|---|---|---|
| T1 | 0.10 | 0.045 (0.046) | yes |
| T2 | 0.10 | 0.063 (0.065) | yes |
| T3 | 0.10 | 0.042 (0.043) | yes |
| T4 | 0.10 | 0.053 (0.055) | yes |
| T5 | 0.20 | 0.263 (0.269) | no |

## 4. Exploratory (outside the decision family)

- **P1, temporal incremental prediction:** adding Q and D to the baseline model changes held-out MSE by −0.15 % [−0.69, +0.38] (relative gain). The baseline does beat the constant (sanity statistic 0.333 [0.295, 0.372] > 0), so v1's defect is absent. Complexity burden and its change add no predictive value for PR effort.
- **S1** (review latency ~ Q): −0.023 [−0.095, +0.050]. **S2** (rework commits ~ Q): +0.006 [−0.013, +0.024].
- **T5** (H4, exploratory because G5/G7 failed): −0.024 [−0.151, +0.102]. With the explicit-reference follow-up outcome (sensitivity e): −0.038 [−0.188, +0.111].

## 5. Sensitivity (`sensitivity.json`; preregistered §8, not decision-bearing)

| Variant | T2 (H1) | T4 (H2) | T5 |
|---|---|---|---|
| Primary | +0.023 [+0.001, +0.045] | +0.006 [−0.014, +0.026] | −0.024 [−0.151, +0.102] |
| (a) ≥ 2 commits (post-treatment conditioning, descriptive) | +0.021 [−0.002, +0.043] | +0.004 [−0.022, +0.030] | −0.006 [−0.154, +0.142] |
| (b) session gap 2 h | +0.014 [−0.006, +0.035] | −0.003 [−0.022, +0.015] | −0.024 |
| (b) session gap 8 h | +0.029 [+0.003, +0.054] | +0.009 [−0.017, +0.036] | −0.024 |
| (c) leave-one-repository-out, range of estimates | +0.006 (vite out) … +0.029 | +0.001 … +0.011 | −0.065 … +0.028 |
| (d) cohort merged < 2024 | +0.024 [−0.013, +0.060] | −0.003 [−0.031, +0.025] | −0.282 [−0.513, −0.051] |
| (d) cohort merged ≥ 2024 | +0.024 [−0.004, +0.051] | +0.010 [−0.018, +0.039] | +0.107 [−0.024, +0.237] |
| (e) explicit-reference follow-up | (= primary) | (= primary) | −0.038 [−0.188, +0.111] |
| (f) author-clustered inference | +0.023 [−0.008, +0.053] | +0.006 [−0.016, +0.028] | −0.024 [−0.169, +0.120] |

Every pooled variant keeps T2 and T4 well below the SESOI; every 95 % upper bound is ≤ 0.061 (largest: T2 in PRs merged before 2024, 0.0603).

**(d) per repository.** T2 is heterogeneous: pandas +0.096 [+0.048, +0.144], vite +0.061 [+0.009, +0.113], tokio −0.114 [−0.245, +0.016], jellyfin −0.048, scikit-learn −0.043; the other repositories are within ±0.03. For axum and caddy the per-repository refit stopped on the logit non-convergence guard (`RuntimeError`, T5 model on a small single-repository sample), so those two rows have no estimates. django contributes only T5 (its `Y_active` is masked by the floor rule). Per-repository estimates are single-repository and exploratory; with 12 repositories, two nominally positive and one nominally negative interval are within what chance produces [INFERENCE].

T1 in the pre-2024 cohort is −0.039 [−0.072, −0.005], negative for the level association. T5 changes sign between cohorts. Both are exploratory, unadjusted and not interpreted.

## 6. Known threats (prereg §9, as they apply to this result)

- **Proxy outcome.** Active time comes from commit timestamps; single-commit PRs sit at the floor (38 % of rows). G3b supports the construct (ρ = 0.46 with churn), but effort spent outside commits (reading, review, design) is unobserved. An effect that acts mainly through comprehension time before the first commit would be missed [INFERENCE].
- **Coarse quality measure.** Complexity burden density (lizard CC > 10 mass) is the only quality measure confirmed per PR at the snapshot. Duplication, coupling, cohesion and architecture erosion were not tested here.
- **Population.** Merged PRs only (abandoned or reverted work is unobserved); 16 repositories chosen for activity and language coverage; one pinned window each; 12 survive the gate.
- **Association, not causation.** Area fixed effects absorb time-invariant area difficulty for T2/T4, but time-varying confounding remains possible.
- **Validation.** LLM annotators are not human experts. G4 passed with κ 0.77.

## 7. Deviations

1. **Amendment 1 (signed off in rounds 11–14).**
   - **Run 1.** All 95 adjudications by annotator C (gemini-2.5-pro) were `unparsable`: `max_tokens` 20 truncated the reasoning model. G4 failed on the unresolved share alone (0.173), and the study stopped as prereg §5 requires. Run-1 artefacts are preserved in `validation/run-1-failed/` (SHA256SUMS).
   - **Fix.** `annotate.py` now aborts on any non-`stop` or empty completion and after three transport failures, rather than recording `unparsable`. A new `--adjudicate` mode reruns C alone, reusing the run-1 A/B labels.
   - **One-shot rule.** A G4 failure after the rerun would have stopped the study.
2. **Aborted attempt under revision 1 (round 13).** At 10:05, `annotate.py --adjudicate` aborted with `annotator C: incomplete or empty completion (finish_reason 'length'); return to council`: unbounded Gemini thinking exhausted 4,096 tokens. Nothing was written and no C label was seen. Revision 2 bounds C's thinking (`reasoning.max_tokens` 2048, `max_tokens` 8192). The 10:15 attempt completed: 95 labels, 0 `unparsable`.
3. **Annotator budgets differ.** A and B ran at `max_tokens` 20 (run 1, reused); C ran at 8,192 with thinking bounded at 2,048. `validation/annotator-prompts.json` records `{"A": 20, "B": 20, "C": 8192}`, `reasoning {"C": {"max_tokens": 2048}}` and `ab_reused_from_run_1: true`. Bounded thinking may make C a weaker adjudicator than unbounded C would be [INFERENCE].
4. **Key source.** `OPENROUTER_API_KEY` was injected into the process environment with `infisical run` (a user-designated project, `dev` environment). It was never printed or stored by the study scripts.
5. **Power and constructs.** Both were re-run at steps 13 and 15 and reproduced their frozen files byte for byte (`2c924dcb…`, `68325887…`). `gates.json` gained G3b at step 14 by design.

## 8. Reproduce

```
U="uv run --python 3.13 --with-requirements requirements.txt python"
cd evaluation/change-cost-v2
$U validate.py sample
infisical run --projectId <dev> --env dev -- $U annotate.py              # run 1 (A, B; failed C, archived)
infisical run --projectId <dev> --env dev -- $U annotate.py --adjudicate # Amendment 1 rev 2
$U validate.py gates
$U constructs.py panel-final.csv.gz > constructs.json
$U gates.py --g3b > gates.json
$U power.py panel-final.csv.gz 60 power-results.json
$U analysis.py panel-final.csv.gz results.json --power power-results.json --validation validation-gates.json --gates gates.json
$U sensitivity.py panel-final.csv.gz sensitivity.json
```

Run artefacts (SHA-256):

```
1f4b801007299974bb15ce3dabeec3fa93dee6cac5db7c5911d1d6a27720f761  results.json
50360c3e5464e171995088764b3d601af61f494a92b59347dc44527e0fd540ee  sensitivity.json
4a9dc44fdd2bde67f26aff46e2f09eafe2ceb3e1e0c1279c172a612587fdadd3  validation-gates.json
788c93c0cc2ed0f4ed1242b63969c09176a1eb24dfbfc87c083fb9ed985eaa0b  gates.json
5ffc215c016055bf2eb057693cf9db930e51ae3fb4fa5ab60bbda80cbfd9b184  validation/annotations-C.jsonl
c7918ad0b7e2e7271b7c25e4bb341c52d0ebe9230e6c4ed40c50067b0c1f1ffd  validation/annotator-prompts.json
2c924dcb939c725923238629321f891d0b49d5665b0cfcc78d779b523cd35f2b  power-results.json
683258878ea9694c6449089b2a76b2f4933458cb4701f0d1e94203becb170438  constructs.json
```
