# E1b-v2 — Innovation council review, round 2

Scope: `evaluation/change-cost-v2/` after the round-1 fixes (PREREGISTRATION.md, DESIGN-NOTES.md, 17 scripts, result JSONs, `reviews/round-1*.md`). Read-only; `.codegraph` untouched; no secrets printed; no predictor–outcome association computed (only marginals in `constructs.json`/`gates.json`/`power-results.json` read). Main agent ran `tests.py` (42 checks pass, 2026-10-20). An independent challenger (different model family) reviewed in parallel; synthesis at the end.

## Verdict: ISSUES FOUND — do not run.

All four round-1 blockers are substantively fixed; what remains is three must-fix items (one gate not mechanical, one validation sample misaligned with the proxy it validates, one power claim that the artefact contradicts) and documentary/edge-case minors. All are small and localised.

## Digests
All 18 digests in `/tmp/e1bv2-digests-round2.txt` (PREREGISTRATION.md + 17 scripts) match `DESIGN-NOTES.md` §10 line by line. Not verifiable here: `requirements.txt` and the result/data digests listed in §10 (absent from the main agent's file). No digests are certified in this round because fixes are required.

## Round-1 resolution check
| # | Status | Evidence |
|---|---|---|
| 1 | **Partially** | Any-via rule and comment/generic-stem filter implemented (`collect.py:119-135, 228-233`, tested `tests.py:104-113`); G7 exists (`validate.py:200-215`, prereg §5). But G7 shows the *best-confidence* via, not the file that produced `explained` → new issue 2. |
| 2 | Resolved (accepted limitation) | `harness.py:144-146,162-164` report/gate bias vs planted; `harness-results.json` recover: T1 −0.004, T2 +0.005, T5 +0.025, all checks true. Prereg §3.3 states non-separability; sensitivity (a) labelled post-treatment (`sensitivity.py:33`). Decision (b) below. |
| 3 | Resolved | IDs after shuffle (`validate.py:98-104,114-118,142-146`); blinded files drop `pipeline/kind/repo/stratum/lang/proxy` (`:147-154`); `annotate.py` pins models, temp 0, provider check, prompt SHA; `analysis.py:296-299` requires `--validation`, G4 fail exits, G5/G7 fail ⇒ `t5_decided=False` (`:269-285`). `cannot_tell` policy in prereg §5. Edge case → issue 4. |
| 4 | Resolved | `gates.py:26-27` reach = ok+no_snapshot+no_preexisting+no_measurable; `gates.json` axum 581/30 (= 540+5+6+30 in `funnel.json`); G3 n=7,030, 38.15 %. Panel regenerated (timestamps). One stale number → issue 9. |
| 5 | Resolved | (a) §3.2 "realised properties"; (b) §4 T5 row; (c) §3.3 "no model call", `finalize_panel.py:9-10`; (d) `decide()` drops T5 from Holm (`analysis.py:273`); (e) docstring (`analysis.py:231-238`). Residual wording → issue 6. |
| 6 | **Partially** | Own-sample standardisation (`power.py:41-50`), k-calibration and MC bounds (`:98-114`), tie rule coded (`adequate=mde_ub<=SESOI`, prereg §7). But T4 verification contradicts the MDE80 label → issue 3. |
| 7 | Resolved | Per-test `present`/G_t, t(G_t−1), hard error on non-finite decision replicate (`analysis.py:250-262`). |
| 8 | Resolved | `--power`/`--validation` required; `adequate.get(t, False)` (`analysis.py:280,296-297`); tested (`tests.py:177-179`). |
| 9 | Resolved | Prereg §4 SESOI paragraph, per scale, marked [INFERENCE]. Decision (c). |
| 10 | Resolved | `common.is_bot` used by `collect.py` and `follow.py:48`; `is_follow` ≥ boundary (`follow.py:30-33`, prereg §3.3); tests. |
| 11 | Resolved | Prereg cites H1/H2/H4 and R-IDs (§1); `reviews/round-1.md` exists. |
| 12 | Resolved | `fetch_prs.py:115-117` consumes pins; `requirements.txt` pinned; `qhash` in `jev_label.py:49,67`, `jev_type.py:46,64`; `prompts.json`; `OPENROUTER_API_KEY` only. |
| 13 | Resolved | All requested checks present (`tests.py:102-184`). No tests for the new G7/G5 scoring → issue 8. |
| 14–17 | Resolved | `validate.py:11`; `sensitivity.py:40-42`; prereg §6 last sentence; key presence-checked only. |

## New issues
**1. Must-fix — G3b/G3 consequences are not mechanical.** Prereg §5 says G3b failure ⇒ H1/H2 "outcome construct invalid" and G3 ⇒ H1/H2 not tested, but `analysis.py main()` (`:294-305`) reads only `--power` and `--validation`; `gates.py --g3b` only prints JSON (`:33-37`). This is the same human-set-flag pattern round-1 issue 3 removed. *Fix:* `--gates gates.json` required; if `G3.pass_` or `G3b.pass_` is false, T1–T4 get verdict `outcome_construct_invalid` and are removed from the Holm family (state this in §4/§5); add a `decide` test.

**2. Must-fix — G7 validates a different pair than the proxy decided.** `explained` is true if *any* via references g or vice-versa (`collect.py:229-230`), but the G7 item shows `pt["via"]` = highest-confidence via (`collect.py:232-233`, `validate.py:141-142`). An explained partner whose reference came from another touched file is shown as a non-referencing pair → counted against "explained precision"; a hidden partner with several vias is checked on one pair only → hidden precision overstated. *Fix:* `collect.py` records `via_ref` (the file and direction that triggered the match); G7 shows (`via_ref`, g) for explained items and, for hidden items, every via (≤3, pick by `rng`) with the item judged "no" only if all pairs are "no"; prereg §5 text accordingly. Also state that `jvm_dotnet` has no gate-passing repository (junit5/eShopOnWeb fail), so G7 has ≤ 96 items.

**3. Must-fix — T4 "MDE80" is not an 80 %-power effect.** `power-results.json`: T4 `power_at_mde_p01` 0.60, half-width 0.124 → upper 0.72 < 0.80 (T3 0.68+0.12 = 0.80 borderline). DESIGN-NOTES §7 says "within MC error" — false for T4. Adequacy itself is robust (0.044 ≪ 0.10) but the reported MDE is mislabelled. *Fix:* compute an empirical MDE80 (plant analytic MDE × {1, 1.25, 1.5}, ≥ 200 sims, smallest with power ≥ 0.80) or label the column "analytic MDE80" with verification power beside it; correct DESIGN-NOTES §7.

**4. Minor — identical `unparsable` answers become a consensus.** `validate.consensus()` (`:171-172`) returns `A[i]` when `A[i]==B[i]`, including `"unparsable"`; in G5/G7 such items are neither in `dec` nor counted in `ct` (`:196-198,210-214`) — silently dropped, contradicting prereg §5 ("an unparsable answer is a disagreement") and `annotate.py` docstring. *Fix:* treat `unparsable` as `None` in `consensus`.

**5. Minor — annotator identity check is provider-level only.** `annotate.py:46` compares the provider prefix, so any Anthropic/OpenAI/Google model would pass. *Fix:* require `resolved == MODELS[letter]` (or `startswith(MODELS[letter])` to allow dated suffixes) and record it.

**6. Minor — preregistration wording still contradicts §5/code.** §4 "Holm over the five tests T1–T5" and §5 header "failure ⇒ study stops" vs G5/G7 ⇒ T5 `not_decided` and the study continues; §4 "no shrinkage … of any effect" vs ridge 1e-2 on the T5 coefficient (`analysis.py:99-108`) — say "fixed ridge 1e-2, never tuned" and either substantiate "< 1 %" by a test (λ=0 vs 1e-2 on synthetic n=4,000) or mark [INFERENCE]; §3.3 "mean of 3 repetitions" vs `finalize_panel.py:41-55` mean of successful reps (all 60,342 ok, so state the rule); §6 "median area size ≈ 6" vs 5 (`constructs.json overall.ALL.median_area_size`).

**7. Minor — per-repository sensitivity loses django's T5.** `estimate()` on a repo with no `y_active` runs T1 on an empty sample → `dummies()` `codes.max()` raises → `safe()` records an error for the whole repo (`sensitivity.py:23-27,39`), although T5 is measurable. *Fix:* `estimate` skips a test whose sample is empty (NaN), exploratory only.

**8. Minor — untested new code.** No test exercises `validate.gates()` (kappa, IPW recall, G7 by-language, cannot_tell share), `annotate.parse`, or `analysis.main` with a validation/power file (DESIGN-NOTES §8 admits the CLI was last plumbing-tested before the round-1 change). *Fix:* fixture-based tests with synthetic sample/annotation files and a synthetic panel end-to-end CLI run.

**9. Minor — stale DESIGN-NOTES numbers.** §1 "31 software tests" (§8: 42); §3 "Hidden coupling: 23 % of PRs > 0" vs 24.8 % (`constructs.json hidden_zero_share` 0.752); §7 T4 sentence (issue 3).

Checked and found correct: all harness checks true in `harness-results.json` (null FWER 4.0–7.5 %, coverage ≥ 0.93 on nulls, recover coverage 0.90–0.96, detect T2/T4 0.91/1.00); permutation derangement; power plants on each test's own sample; jackknife per-test df; Holm; `is_follow` boundaries; `outcome_rule` NaN rules; pins as input; no prompt contains predictor or outcome values; nothing fits real predictor–outcome data (`gates.json G3b` null). Tolerances in `harness.checks` are fixed constants recorded in prereg §5; the harness seeds are fixed, so rule 4 is met once the above are fixed.

## Decisions on the open questions
**(a) LLM annotators — accepted** for G4/G5/G7: blinded, pinned, different family from Jev, temperature 0, resolved model recorded (tighten per issue 5). G7 thresholds (hidden ≥ 0.80, explained ≥ 0.70, per-language ≥ 0.60, κ ≥ 0.6, cannot_tell ≤ 15 %) are sound as preregistered point-estimate screens; with 12 items per cell the per-language rule has little power and must be described as a sanity check, and `validation-gates.json` should report cell sizes. Fix issue 2 first, else G7 measures the wrong pairs.

**(b) No two-part/Tobit primary required.** The single-commit floor is not censoring at a known threshold (the duration is unobserved, not bounded at 0.5 h), so Tobit is mis-specified; a two-part model's second stage conditions on commit count, a post-treatment variable — the same defect as sensitivity (a). The floored measure with the explicit non-separability statement (§3.3), the ungated `recover_floor` report and the T2/T4 pairing is the honest option. Condition: RESULTS must present T2 and T4 together and never state H1 support without the T4 estimate alongside (already implied by §3.3).

**(c) SESOI values accepted** (T1/T2 0.10 log-h, T3/T4 0.10 relative dispersion, T5 0.20 log-odds). Do not raise T5 to meet the MDE: that would be tuning the threshold to the design. T5 therefore stays inadequate (upper MDE80 0.215 > 0.20) and H4 can only be `supported` or `inconclusive`; this is conservative (the 90 % CI half-width ≈ 0.10 would often allow equivalence) and acceptable as frozen.

## Challenger synthesis
Adopted: G3b not consumed (their 19, also found independently), G7 best-via mismatch (1), `unparsable` consensus (3), exact model identity (18), django per-repo loss (21), T4 verification power (6), wording contradictions (5), stale 23 % (4). Rejected: coverage must target planted values (a CI should cover the estimand the estimator defines; the floored projection is reported), language-aware comment stripping (G7 governs the proxy), "sample quotas unenforced" (`annotate.py` writes every item; failures are `unparsable`), tolerances circular (constants are preregistered and seeds fixed; recorded here), `$U` notation. The challenger rated issues 1–3 blocking; I rate them must-fix because each is a bounded code/text change with no effect on the frozen data or models.

## Sign-off
Not granted. Re-review after fixes with regenerated `power-results.json`, `tests.py` (new checks) and recomputed digests.
