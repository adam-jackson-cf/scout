# Jev design and implementation review against the `jev-actions` skill

Reviewer: `JevReview` (read-only; this is the only file written). Date 2026-10-22.
Status: **input to the pending re-scope round** (`PROPOSAL.md`). It is not a council sign-off and changes nothing by itself. Design changes in §A still need the innovation council (`AGENTS.md` design rules 1-2).

## 0. Method, sources and limits

Skill read in full: `~/.omp/agent/skills/jev-actions/SKILL.md` and all eight files under `references/`. Abbreviations used below: `DP` design-patterns, `QD` question-design, `CI` compose-and-integrate, `BED` building-eval-data, `LC` labeling-cases, `TT` threshold-tuning, `CC` calibration-check, `EA` error-analysis, `DOCS` docs-and-programming-model. Line numbers refer to those files.

"Read the live docs" (`DOCS:11-17`) was run on 2026-10-22 in two passes.

**Docs read** (all https://docs.typesafe.ai/…): `llms.txt` (index; no Jev 1.14 jaggedness page, no OpenRouter/decisions page, no versioning-policy page); `concepts/how-to-build-with-system-one.md` (lines 1-300 of 431); `concepts/state.md`; `primitives.md`; `primitives/choice.md`; `primitives/score.md` (lines 1-300 of 439); `primitives/noul.md`; `confidence.md`; `models.md`; `api/system-one.md` (OpenAPI, lines 1-300 of 905); `api/errors.md`; `model-jaggedness/jev-1.13.md`; `model-jaggedness/jev-1.14.md` (404); `patterns/fan-out.md`; `patterns/confidence-routing.md`; `patterns/composite-scoring.md`; `cookbooks/parallel_questions.md` (lines 1-300 of 352); `cookbooks/consistency_noul_cookbook.md` (lines 1-300 of 702).

**Not read:** the SDK pages (Python/JS), `agent-skill.md`, the remaining cookbooks, `hierarchical_classification` (cited from its index description only), the rest of the truncated pages above. Anything resting on those, or on facts the docs do not state, stays `[INFERENCE]`. `models.md` lists `jev-1.14.0` (released 2026-10-14) and says `jev-latest` points to `jev-1.13.0`. Doc-driven changes made in the second pass are in J1-4, J1-5, J1-7, J1-10, J1-14, J1-18, E2-6 and question 12 of §D.

One measured number below is mine, not the studies': I recomputed G5 recall/precision as a function of the labeller threshold from the **saved** scores and annotations in `evaluation/change-cost-v2/` (read-only, no model calls, no outcome-predictor association; it concerns a labeller's measurement quality, which `AGENTS.md` rule 3 permits). It is informational only (§C).

Severity: **blocker** = the design or rule cannot do what it claims; **major** = likely wrong result or wasted experiment unless changed; **minor** = tighten.

---

## 1. Product design (`docs/scout-design.md`, REQ-28..30)

Overall fit: the architecture follows the skill's core stance, with code owning control flow, eligibility, gate and identity and Jev limited to semantic judgments (`DOCS:13`, `DP:12`; design §6 table, lines 137-142). Annotation-only output, `issues` not importing `decisions`, cost cap, disclosed deterministic fallback, stored baseline and the audit sample are all sound. Findings follow.

### 1.1 §6, triage ranking (lines 140, 144, 146, 148)

**J1-1 (major). `Noul primary` is rejected on evidence that does not test the skill's pattern.**
Evidence: line 148 ("Rejected: … Noul primary") cites E2. E2's `v2_noul` (`evaluation/jev-triage-v1/questions.py:9-13`) is one Noul that bundles five conditions: shipped, hand-written, non-test, unguarded, not generated. Its 0.3/0.7 band was then mapped onto a three-way `cannot_tell` label (`RESULTS.md:12`). E2's own ranking metric does not show Noul worse: AUROC 0.932-1.000 against Choice 0.940-0.991 (`RESULTS.md:27-38`); the difference is only in 3-way routing. Skill: `QD:33-37` ("one Noul per reason or mode … combine 'any' in code", not one broad question) and `DP:52-59`.
Change: re-word line 148 to "single-Noul bundle rejected for 3-way routing (E2, n = 22)". Do not state that Noul was shown worse for ranking. See J1-2.

**J1-2 (major). The triage question bundles several judgments; part of it is decidable by code.**
Evidence: `questions.py:5-8` (v1) and `:16-19` (v3) put "shipped, hand-written, non-test code", "no visible guard", "concrete weakness" and "worth a human's attention" into one `real_defect` option. E2 itself notes that most FP discrimination is file-type visible from `path` (`RESULTS.md:71`, caveat 3). Design §6 line 146 already has `path role` in the deterministic baseline. Skill: `QD:33-39`, `EA:16-25` ("Rule a parser could decide: move it into code").
Change: (a) path/test/fixture/generated role is computed in code and supplied as a fact in state (`CI`/`EA`: "add … a code-computed fact"); (b) Jev gets narrow questions per family fanned out in one request (e.g. `mechanism_present_in_excerpt`, `guard_visible_in_excerpt`, `inputs_attacker_controlled`) with explicit Noul criteria; (c) the ranking score is composed in code from named per-question outputs (`CI:26-30`, `DOCS`/composite-scoring), so a weight change never needs a re-run (`CI:29-30`). If the council keeps one Choice, record why. This also makes X1's "incremental gain over baseline" meaningful: otherwise Jev's gain can be the path role the baseline already holds [INFERENCE].

**J1-3 (major). Question identity and version key.**
Evidence: line 144 says `question_id@version`, with no definition of `version`. Skill: `CI:35-38` (version key = hash of instructions, criteria, state shape and model; a stored threshold/mapping is stale when it changes); `CI:29` (question IDs are not sent to the model). E2 persisted variant names only (`run_jev.py:13-17` stores `variant`, no hash), so the three variants cannot be tied to their exact text from the raw files; E1b's `qhash` (`jev_type.py:64-65`) hashes the question JSON only, not state shape or model, and none of the 70,542 saved E1b records carries it (`prompts.json:5`).
Change: define `version = sha256(canonical JSON of {instructions, criteria, state_shape, resolved_model_id})`; store the full question as sent beside every annotation; treat any stored threshold, band or audit statistic keyed to a different version as stale (see §B).

**J1-4 (major). Model pinning is specified but not its source, scope or upgrade path.**
Evidence: line 144 "resolved-model pinning (mismatch → discarded, unit `incomplete`)". Skill: `CI:32-34` pin a versioned ID, never `jev-latest`; record what the response reports. Docs `models.md`: `jev-latest` → `jev-1.13.0`, `jev-preview` → `jev-1.14.0`; aliases move without a code change. E2 and E1b both pinned a dated OpenRouter slug (`typesafe/jev-1.13-20260917`, `typesafe/jev-1.14-20261014`) and checked the resolved model equals the request only by eye (`RESULTS.md:9`); `jev_type.py:52` records `model_resolved` but never compares it.
Gaps: where the pinned model is set (`.scout/policy.yaml`?), what happens on a legitimate slug roll, and what a model bump does (E2 itself says "re-measure on any model bump", `RESULTS.md:78`). **Docs settle the principle:** the response `model` "may differ from the alias supplied" (https://docs.typesafe.ai/api/system-one.md, `SystemOneResponse.model`), aliases move without a code change, and versioned IDs such as `jev-1.14.0` are accepted whether or not listed (`models.md` "Aliases", "Listing models"); the docs' own cookbook keys its cache on a hash of state + questions + model and stores the returned model "because an alias can resolve to a different version later" (https://docs.typesafe.ai/cookbooks/consistency_noul_cookbook.md). This confirms the finding and the J1-3 version-key shape. **Not settled `[INFERENCE]`:** whether OpenRouter's dated slug `typesafe/jev-1.14-20261014` is retired or re-pointed, and the `https://openrouter.ai/api/alpha/decisions` contract, are in no TypeSafe doc (the documented endpoint is `https://api.typesafe.ai/v1/systemone`; `llms.txt` lists no OpenRouter page). A forced mismatch that makes every unit `incomplete` indefinitely needs a defined operator path.
Change: pinned model ID in policy, required; a bump = new version key = old thresholds/bands stale = annotations marked stale-version and ranking falls back disclosed until re-measured (or the maintainer accepts `provisional`). Phase 5 acceptance (line 336, "annotations carry resolved model and question version") should add "bump marks prior annotations stale".

**J1-5 (major). Failure semantics are partly specified.**
Evidence: line 144 covers model mismatch only. Skill: `CI:39-40` ("a failed or invalid Jev call is a failed evaluation or decision, never a pass or clean result"). E1b's outcome rule already applies this (`finalize_panel.py:25-30`: any pair lacking a successful label → outcome `None`, never 0).
Missing cases: HTTP errors, timeouts, schema-invalid answers, and units not reached because `max_per_run` was hit. The docs list the status codes: 400 (including `max_tokens_exceeded` when state + questions exceed the context), 401, 413, 415, 422, 429 and 529 Overloaded, with exponential backoff for 429/529 (https://docs.typesafe.ai/api/errors.md "Errors", "Handling rate limits"); the adapter must treat each as a failed unit, with `max_tokens_exceeded` and 422 as non-retryable. 402 appears only on E2/E1b's OpenRouter path (`run_jev.py:26`) and is not in the TypeSafe docs.
Change: add to §6: a unit with no valid annotation gets `ranking: deterministic (reason)` per finding (not run-level only), is never treated as low-ranked, and cap-hit selection is by deterministic baseline order and disclosed as `ranking: partial (n/N)`.

**J1-6 (minor). Untrusted-text instruction.**
Evidence: E2 and E1b questions never tell the model to treat state as data (`questions.py:4-19`, `jev_label.py:28-38`); the E2 state is raw source with comments. Skill: `QD:31-32`; jaggedness 1.13 §6 (state "can move the answer"). X1 already tests injection negatives (line 367).
Change: every Scout question carries a data-not-instructions clause and refers to state fields by backticked path (`QD:26-27`); redaction (the `contextual` sanitizer) is not an injection defence and the spec should not imply it is.

**J1-7 (minor). Choice option order.**
Evidence: jaggedness 1.13 §8 ("leans toward the option that comes first"); QD:49-51. Every E2/E1b Choice lists the positive class first (`questions.py:6,17`; `jev_label.py:33`; `jev_type.py:29`) and no order permutation was run. E2's unlabelled sample shows an over-call of `real_defect` (v3 63/90, `RESULTS.md:53`, `:80`), which is consistent with an order effect [INFERENCE; also consistent with wording]. **Docs limit:** the only jaggedness page is for 1.13 (https://docs.typesafe.ai/model-jaggedness/jev-1.13.md); `model-jaggedness/jev-1.14.md` returns 404 and is not in `llms.txt`, so whether option-order bias applies to the pinned `jev-1.14` (used by E2's second run and all of E1b-v2) is undocumented and stays `[INFERENCE]`; it must be measured, not assumed absent.
Change: Phase 2/5 harness runs each Choice in two orders on the X1 set and reports the flip rate; options order stored in the version key.

### 1.2 §6, uncertainty and escalation (line 141, 148; risk R10)

**J1-8 (major). "Second opinion on Jev-uncertain high-severity items" has no definition of uncertain, no measured band and no fallback evidence.**
Evidence: line 141. Skill `EA:34-48` (escalate only after measuring both stages on the same labelled cases, band by hand from saved results, per question, with an agreed accuracy loss; "if no band qualifies, set none"), `EA:51-52` (fewer than ~30 cases per label, or no agreed accuracy loss: set no band, record a TODO), `EA:55-57` (a fallback with the same blind spots cannot fix confident errors; prefer different error patterns or a person), `EA:58-61`.
Facts: the E2 reasoning-LLM baseline was Claude Opus on one rep and 22 cases and missed the same python-multipart case as Jev (`RESULTS.md:64`) — a shared blind spot, the very case where escalation was needed. No band exists. Under REQ-31 the second opinion runs through `claude -p`, so its cost and latency (≈ $55-65 / 1k and 8-9 s, E2) must be in the measurement.
Change: state in §6 that the band is **unset** at launch (no escalation on confidence), that "high-severity" is a deterministic rule attribute, and that setting a band requires a preregistered measurement under the experiment rules; until then the opt-in second opinion is requested by the maintainer, not routed by Jev confidence. Add to X1 or a new experiment: the cascade accuracy and escalation rate on dev and test.

**J1-9 (major). The audit sample is designed for one tail only and not wired to measurement.**
Evidence: line 144 "5 % random low-ranked audit sample nightly"; line 146 "Every disposition on an annotated issue becomes a blind label". Skill `EA:65-66` (label a random sample of **confident** decisions, because confident errors never reach the queue); `BED:56-59`; E1b's G5 shows the working method: stratify by Jev score, record stratum sizes and weight by inverse inclusion probability (`validate.py:210-221`).
Problems: (a) dispositions are made on what the maintainer sees first, so dispositions are not a random sample of findings; ranking changes which items get labelled and therefore the measured recall; (b) "blind" is unmet if the UI shows rank/annotation at disposition time (`LC:20`); (c) 5 % of a low-ranked tail cannot find rare critical items: with 100 audited and a 0.5 % critical-miss rate the expected hits are 0.5, so "no critical true issue below the audit floor" (X1) is untestable as stated [INFERENCE from simple expectation].
Change: (i) draw the audit sample from both the top-ranked and low-ranked strata (stratum recorded per finding; inclusion probability stored); (ii) the UI records whether the annotation was visible when the disposition was made, and X1 uses only dispositions from the audit sample or labelled blind; (iii) X1's critical-miss test labels **all** critical-severity findings below the floor (a deterministic set), not a 5 % sample; (iv) metrics use the inclusion weights.

**J1-10 (minor). `P(real_defect)` "not calibrated" should read "calibration not checked on Scout data; used for ranking only".**
Evidence: lines 140, 148; REQ-30 (`SCOUT_REBUILD_REQUIREMENTS.md:53`); E2 caveat 6 (`RESULTS.md:74`). Skill `CC:8-12,17-18,40-42`: a probability is read as confidence only after a user-data calibration check; ordering alone needs none. Docs `confidence.md` and `how-to-build…` call Jev's output calibrated by training, so "not calibrated" over-claims in the other direction; E2 meant "not measured".
Change: wording as above. The three places that do read the number must be named: any display of P as a percent or "likelihood" (needs every held-out reliability bin within 0.10, ~20 cases per bin, `CC:40-50`, so ≥ 100 labels over five bins, more than X1's minimum); any hybrid rank that weights P against `rule weight` (composite, so `CC:10-12`); any threshold that defines "Jev-uncertain" or the audit "floor" (threshold-tuning on a single cut, calibration for a two-cut band). Until measured, show rank bucket (high/normal/low), never the number. Docs add (https://docs.typesafe.ai/confidence.md, "Thresholds scale with risk": "start with conservative thresholds, test with your own data"; https://docs.typesafe.ai/primitives/noul.md "Reading a Noul": 0.7/0.3 is a starting point, "avoid fitting thresholds closely to a handful of test cases") that the docs' own gates are starting points, consistent with the skill. Run-to-run noise is documented as small but non-zero: some Noul answers vary by ~0.005-0.01 across repeats, and one `covered` answer spanned 0.43-0.53 over 15 repeats (https://docs.typesafe.ai/cookbooks/parallel_questions.md; https://docs.typesafe.ai/cookbooks/consistency_noul_cookbook.md), so a cut placed near a probability can flip on resampling; X1's audit "floor" and any "uncertain" band should be evaluated over the repeat distribution, not one call.

**J1-11 (minor). "Jev on from launch" precedes its validation; the default is the user's decision and is kept.**
Evidence: Appendix 4 (line 414), Phase 2 vs Phase 5 acceptance (lines 322, 336), X1 (line 367). Skill `BED:27-30` keeps automated actions in shadow mode until confirmed. The action here is reversible ranking with a stored baseline, so this review does **not** propose reopening it. Change: disclose in the product (`ranking: jev (unvalidated, X1 pending)`) until X1 is recorded, and keep the E2 set (n = 22) described as an engineering observation, not as evidence for the default.

### 1.3 §5.1 lane 6 `rubric` (line 111), X4 (line 370), A5 (line 386)

**J1-12 (major). Wrong primitive for the "presence" decision; per-pattern questions and thresholds are missing.**
Evidence: line 111 "Jev Choice where semantic". Twelve patterns can co-occur in one function; "does this excerpt show pattern P" is a yes/no. Skill `QD:14-22,34-37` (Noul per label when several may apply; one Noul per mode; "any" in code), `DP:52-59`, `CI:11-13` (fan out over the same state in one request). Also `CI:17-18`: gate on named per-question thresholds.
Change: 12 Nouls, one request per changed function, each with a `true/false` checklist and the Go idiom exclusions the 45-pattern rejection (line 111) already names; each pattern has its own named threshold from X4 labels (`TT:90-92`), never a literal. Pattern results stay report-only (already true).

**J1-13 (major). X4's labelling scheme cannot support tuning or a recall minimum.**
Evidence: line 370 "12 patterns × 300 changed … functions; double-labelled hits; keep iff Wilson LB ≥ 0.6". Skill `BED:57,60-61,92-98` (≥ ~30 per label before tuning; both labels per stratum, ideally ≥ 10; collect more rare positives; do not tune below 30), `TT:18-21` (precision and recall minimums).
Problems: labelling only **hits** gives precision but no recall; at a realistic rare-pattern base rate 300 functions give a handful of positives per pattern [INFERENCE: no base rate is cited anywhere], so the 12 patterns are individually below the 30-label floor; a threshold is needed before "hits" exist, and X4 does not say how it is chosen; there is no dev/test split.
Change: (a) sample strata by Jev score and a deterministic Semgrep pre-filter and label a random slice of non-hits (IPW, as G5 did); (b) enrich candidates per pattern to ≥ 30 positives; if a pattern cannot reach it, report `unmeasured` and keep it out of the lane (not "provisional with a number"); (c) dev/test by repository, thresholds tuned on dev, test scored once; (d) record a recall minimum even though the lane is non-blocking, because a lane that never flags is useless.

**J1-14 (minor). Cost assumption.**
Evidence: line 111 (≈ $0.12 per 200-function PR, `[ASSUMPTION]`), A5 (line 386). **Settled by docs:** questions in one request share one state, and "adding questions … costs only the tokens for the extra questions" (https://docs.typesafe.ai/primitives.md, "Ask multiple questions together"); the parallel-questions cookbook measured a 13-question batch as 12.2× cheaper than 13 single calls because the article is billed once (https://docs.typesafe.ai/cookbooks/parallel_questions.md); price is $0.042 per Mtok input, output free (`models.md`). The earlier worry that 12 questions may bill the state 12× is **withdrawn**: cost ≈ (state tokens once + 12 question/criteria texts) × $0.042/Mtok. The $0.12 figure is still unverified (no per-function token count exists) and stays an `[ASSUMPTION]`. Change: Phase 5 measures billed `usage` per function (E2 recorded it per call, `RESULTS.md:17`); the per-run cap (§6) should be expressed in tokens/cost, not calls, once rubric questions fan out; keep the state under the 32k per-question / 64k per-request default (long context is slower and less accurate, `models.md` "Long context").

**J1-15 (minor). Reasoning-about-code questions.**
Evidence: patterns such as `retry_loops_hiding_failures`, `validation_theater`, `ignoring_return_values` require control-flow reading, not recognition. Skill `QD:52-54` keeps "code correctness" out of Jev, and jaggedness §4 (indirection) lists multi-hop reasoning as a weak spot. Change: where a parser can decide (`ignoring_return_values`, `silent_truncation` forms, bare `except`) use Semgrep (already "where structural"); X4 reports per-pattern results by family so semantic families can be dropped without revisiting the rest.

### 1.4 §8.3 `fix_touches` and X5 (lines 223, 371)

**J1-16 (major). The X5 bar may already be contradicted by E1b's G4, and the sample is too small for the lower bounds.**
Evidence: `evaluation/change-cost-v2/validation-gates.json` G4 confusion (pipeline = Jev label → consensus): `fix→fix` 24, `fix→feature` 6; `other_dev→fix` 3 and `nondev→fix` 2 are other classes' errors. Jev's `fix` precision on 30 sampled = 24/30 = 0.80, Wilson 95 % [0.63, 0.91] (my arithmetic). X5 requires precision ≥ 0.85 with LB ≥ 0.75 and recall ≥ 0.8 with LB ≥ 0.70 on 200 commits. Wilson LB for an observed 0.80 recall is 0.70 only at ≥ 80 positives (0.65 at 40, 0.68 at 60); LB for 0.85 precision is 0.75 at ~80 flagged. At a typical `fix` share of 25-35 % [INFERENCE; no share is cited], 200 commits yield 50-70 positives, so the rule passes only if observed recall is ≈ 0.85+. Skill `BED:92-98`, `TT:18-21,64-65`.
Change: before X5 is preregistered, (i) the minimums are agreed (§D) and the sample is sized to put the LB on the required side (≥ 100 `fix` positives → ~300-400 commits at 25-35 % prevalence, or stratified over-sampling with IPW); (ii) `fix` vs not-`fix` is asked as its own Noul with a tuned threshold, because `fix_touches` consumes only a binary and the 5-class argmax made `fix` compete with `feature` (6/30); (iii) the state shape is stated: E1b's classifier saw PR title, labels, ≤ 700 chars body and ≤ 15 paths (`jev_type.py:40-41`), a commit has no labels, so G4 is not evidence for X5's input and X5 must measure it.

**J1-17 (minor). Aggregate bias.** `fix_touches` is a ratio over classified changes (line 223). Misclassification that varies with change type (docs-heavy vs code-heavy scopes) biases it between scopes, not just in level. X5 should also report classification error by path role and by repository (the 7 repos, line 371), not pooled only.

### 1.5 X3 (line 369), X10 (line 376), X1 (line 367)

**J1-18 (major). X3: primitive and decision rule are under-specified.**
Evidence: line 369 "Choice on 200 unowned files … precision ≥ 0.9 on `none`-exclusive decisions". Skill `QD:48` ("the model cannot choose an omitted value"), `QD:46-47` (no-match outcome), `DP:29-36` (retrieve candidates in code; rerank with per-candidate Nouls or hierarchical Choice), jaggedness 1.13 §5 and §8 (large state, option order). **Docs correction (confirmed in part, withdrawn in part):** a Choice accepts up to 255 options and the docs advise giving the full list rather than a shortlist (https://docs.typesafe.ai/primitives/choice.md, "Good practice: ask more than one question per call"), so "one Choice over all features is a shape to avoid" is withdrawn for ≤ 255 features; beyond that the docs chain Choices hierarchically (https://docs.typesafe.ai/cookbooks/hierarchical_classification.md). The same section warns that a `none of the above` option "may not work well in many use cases" and recommends a companion Noul ("does anything fit at all") read before the Choice, which supports a code-side or Noul-side `none` rather than a `none` option, and the first-option bias in jaggedness §8 still applies to a long option list.
Further issues: "`none`-exclusive" is not defined (decisions excluding `none`? excluding files whose truth is `none`?); no dev/test split; the rule is a point estimate while X4, X5, X9 use Wilson LBs; no recall, so a classifier that assigns three files looks perfect; no threshold or confidence gate named though precision 0.9 is only reachable with one (`TT:9-20`).
Change: code retrieves the top-K candidate features per file (path proximity, import edges) or supplies the full list when ≤ 255; `none` is decided by a companion Noul ("does any listed feature own this file?") per the docs tip, with the Choice read only on yes; option order permuted and stored in the version key; define the rule as "precision LB ≥ 0.9 at a threshold tuned on dev, scored once on test"; record candidate coverage (how often the true feature was in the candidate set) as its own number.

**J1-19 (major). X10: primitive mismatch, pairwise order effect and no label guide.**
Evidence: line 376 "Jev Choice 'same concern, different implementation?'" is a yes/no, i.e. a Noul (`QD:17`), over a **pair**. Skill `QD:57-58` (pairwise: judge both presentation orders and average), `LC:9-11` (a one-page label guide per question, from the question's own criteria), `LC:22-24` (two independent labellers; X4 requires double labelling, X10 does not). "Same concern, different implementation" is itself a multi-hop semantic judgment (jaggedness §4); the intended positives are subtle ("elaborate but wrong" cases `QD:57`). Rule: precision ≥ 0.8 with no interval and no recall.
Change: Noul, both orders averaged, per-pair state limited to the two signatures/bodies; label guide and double labelling like X4; sample strata (clone-group, signature-similar, random) with stratum sizes recorded; rule on a Wilson LB; an explicit statement that without recall the result can license "exploratory" only (already the outcome) but not any prevalence claim.

**J1-20 (blocker for the *rule as written*). X1 has no estimator or sizing for its own decision rule.**
Evidence: line 367. "recall@k at fixed budget" does not define `k` or the budget unit; "Wilson LB > 0" is used for a **difference** of two paired recalls on the same items, where Wilson (a single-proportion interval) is the wrong tool — a paired method (bootstrap over repositories, or McNemar-based) is required. Sizing: with ≥ 30 semantic positives and, say, 20 % discordant pairs the paired SE of the recall difference is ≈ 0.08 [INFERENCE; arithmetic: √(0.2/30)], so observing the stated 10-point gain gives an LB below 0; the rule passes only at ≳ 16 points or with far more positives. "De-correlated" is not defined (E2's own split put the near-duplicate pair `csharp-architecture-…` / `csharp-quality-…`, same file, on both sides, `cases.jsonl` split field), and no dev/test split is declared. `BED:62-69` requires split by an independent unit (repo), question wording frozen before test, test scored once per shipped version.
Change: rewrite as a preregistered design: unit = repository for the split; `k` and budget defined; paired interval on repo-clustered resampling; positives sized by a power calculation or the threshold restated; injection items defined (see next); question version frozen before the labels are shown to anyone. This rule should be fixed during preregistration, but the sizing problem exists in the design text now, so it is a proposed change in §A.

**J1-21 (minor). Injection set wording.** "≥ 20 injection/indirection negatives" (line 367) is ambiguous: the harmful case is a *real* finding whose excerpt contains text asserting it is a false positive (the attacker lowers a rank, line 144). Name them "injected true defects", and keep a separate set of indirection cases (QD / jaggedness §4).

**J1-22 (minor). §9 item 7 (line 275).** "per-rule AUROC … on blind post-launch labels" has too few labels per rule; report pooled AUROC with per-rule only above a minimum count and mark others `insufficient`.

---

## 2. E2 study (`evaluation/jev-triage-v1/`; pre-rules, informs the product question)

E2's own caveats (`RESULTS.md:68-74`) already admit n = 22, leakage, narrow label mix, single adjudicator and uncalibrated confidence. Items below add skill-based findings; none asks for an edit to E2 (it predates the rules). They matter because §6 cites E2 as evidence.

**E2-1 (major). Splits are by row, not by independent unit.** `build_cases.py:44-50` shuffles case IDs; the Rust and C# repos have cases in both splits, and `csharp-architecture-generated-migration-length-4da8212` (tuning) and `csharp-quality-generated-migration-length-4da8212` (held-out) are the same file and lines 1-45 (`build_cases.py:24-25`). Mutable-GitHub-Action cases (6 of 9 positives) cross splits. Skill `BED:62-64`. Effect: held-out accuracy is optimistic beyond what caveat 1 says.

**E2-2 (major). The held-out split was seen before v3 was written.** v1/v2 results on all 22 (including held-out) were reviewed before v3's boundary examples were chosen "from tuning errors only" (`RESULTS.md:13`). Skill `BED:65-69`: test seen ⇒ fresh test set. v3 held-out 11/12 is optimistic; v3 criteria also list the exact mutable-tag and loopback examples of the labelled cases (`questions.py:17-18`), i.e. rule knowledge written into the question.

**E2-3 (major). Labels fail the skill's labeling rules.** Single adjudicator who also read the mechanism text while building the states (`RESULTS.md:69,73`); no per-question label guide; no blind double labelling. Skill `LC:9-11,20-24`. Consequence: E2 measures agreement with one person's reading, not a ground truth; its 9/13 composition is not a stratified design (`BED:56-58`).

**E2-4 (minor). Quantities for design use.** Reusable as engineering observations: cost ≈ $0.049-0.054 / 1k decisions (billed), latency p50 ≈ 0.3 s, 0/396 failed calls, same price/behaviour for 1.13 and 1.14 at n = 22 (`RESULTS.md:17,45`). Not reusable as evidence for: the default-on decision, "Noul worse" (J1-1), AUROC .93-.99 as an expected production figure (the set is mostly path-visible, caveat 3).

**E2-5 (minor). Instability evidence worth keeping.** On the unlabelled sample, v1 and v3 differ sharply on the same inputs (mutable-action: v1 15/42 FP, v3 0/42; v3 never says `cannot_tell`; v2 34/42 `cannot_tell`, `RESULTS.md:52-55`). This shows question wording dominates outcome, which supports J1-3 (version key) and J1-7 (order/wording sensitivity), and means X1 must pin one question set per run.

**E2-6 (minor). Implementation.** `run_jev.py`: the three-retry loop uses a fixed 2 s sleep for every failure (`:24-29`). Docs confirm what the adapter should do on the TypeSafe API: exponential backoff for 429 and 529, no retry for 400/422, SDKs follow `retry-after` (https://docs.typesafe.ai/api/errors.md "Handling rate limits"; `models.md` "Rate limits"). Whether the OpenRouter route returns the same codes and headers is undocumented `[INFERENCE]`. Not an issue at 4 workers; the Phase 2 adapter should implement it. The script records `model_resolved` but does not compare it to the request (`:22`).

---

## 3. E1b-v2 frozen labellers (`evaluation/change-cost-v2/`)

Informational only (§C). Findings are lessons for future questions and for inertia-bench; **no edit to the study is proposed**.

**E1-1 (lesson, major). `TAU = 0.5` was never tuned and the G5 sample cannot pass at any threshold.**
`validate.py:31` and `finalize_panel.py:22` fix `TAU = 0.5` for both the G5 strata and the outcome rule; `prompts.json` records no tuning. Skill `TT:12-17` (never ship 0.5 untuned), `TT:21-25` (F-beta; a miss costing more → beta 2). Recomputed from saved scores (`labels.jsonl`, `validation/annotations-{A,B,C}.jsonl`, `pair-strata.json`; strata weights 16,782 and 1,138; consensus = A when A = B else C; decided items only) — **[measured, informational]**:

| threshold | precision (IPW) | recall (IPW) |
|---|---|---|
| 0.50 (as run) | 0.878 | 0.422 |
| 0.40 | 0.790 | 0.494 |
| 0.30 | 0.673 | 0.566 |
| 0.20 | 0.647 | 0.639 |
| 0.10 | 0.500 | 0.855 |

No threshold meets precision ≥ 0.70 and recall ≥ 0.60 together on these 197 decided pairs. The recall figure rests on 8 consensus-corrective pairs in the 99 decided `jev_neg` items (Wilson 95 % [0.04, 0.15] → IPW recall roughly 0.30-0.60, my arithmetic); the G5 gate used the point estimate. So G5's failure is mainly the labeller, not the cut-off, but the lesson stands: a threshold must be tuned from saved scores before a gate is judged, and the gate interval should be reported.

**E1-2 (lesson, major). Likely cause of the low recall is missing state, not wording.** `jev_label.py:41-44` shows the labeller sees titles, ≤ 700-char bodies, ≤ 15 paths and a number; "completes something `earlier` omitted (a missed call site …)" (`:33-35`) is only decidable from the diff or the omitted call site, which is not in state. Skill `QD:23-25` ("Jev cannot judge facts it has not seen"), `EA:19-20` ("Missing or clipped state: add the evidence"). Only 1.5 % of G5 items were consensus `cannot_tell` or unresolved (`validation-gates.json` G5), and Jev's own `cannot_tell` use was not measured here, so I cannot say the model abstained; the "any pair ≥ τ" rule over up to many candidates per parent also raises false positives with candidate count [INFERENCE].

**E1-3 (lesson, minor). G4 passed but `other_dev` is a catch-all.** κ 0.771, accuracy 0.767 (`validation-gates.json` G4); `other_dev→refactor` 11 and `→feature` 10 of 30, `other_dev` agreed 6/30. Skill `EA:21`: "Policy gap: agree the rule, update the guide, relabel". `jev_type.py:33-34` defines `other_dev` as "not clearly one of the above", which makes a residual class contested by design; type was a control (RESULTS.md:38), so the cost was noise only. Argmax of 1 repetition with ties → `other_dev` (`build_panel.py:34-48`) is sensitive to option order (J1-7) and was not checked.

**E1-4 (lesson, minor). What the study did well, to reuse.** Blinded opaque IDs and stratum-weighted IPW recall (`validate.py:9-12,208-225`); mechanical gates with no hand-set flag; a failed call yields no outcome rather than 0 (`finalize_panel.py:25-30`); resolved model recorded on every call; excerpts bounded; three annotator families with adjudication. Weaknesses against the skill: annotators are LLMs (`LC:15-19` weak evidence; already caveated, RESULTS.md:84), no model-resolved check, no state-untrusted clause, no question-version key over state shape/model, no order check.

---

## A. Proposed design changes for the re-scope round (`docs/scout-design.md`)

Each is an input to the council; none is applied. Items touching only wording are marked (W).

| # | Section / line | Proposed change | From |
|---|---|---|---|
| A1 | §6 line 148 | (W) Replace "Noul primary" with "a single bundled Noul for 3-way routing"; cite E2 `RESULTS.md:27-38` for ranking AUROC parity | J1-1 |
| A2 | §6 line 140 | Add: path role/test/generated/fixture are code-computed facts in state; Jev questions are narrow, fanned out in one request, composed in code | J1-2 |
| A3 | §6 line 144 | Define `question_id@version` as the SHA-256 version key over instructions, criteria, state shape, option order and resolved model; full question stored with each annotation; stale thresholds on key change | J1-3 |
| A4 | §6 line 144 | Pinned model ID is a required policy field; a bump marks prior annotations stale and returns the disclosed fallback until re-measured | J1-4 |
| A5 | §6 line 144 | Per-finding failure semantics: error/invalid/mismatch/cap ⇒ deterministic rank, `ranking: deterministic (reason)` or `partial`, never "low-ranked" | J1-5 |
| A6 | §6 line 144 | Every Jev question carries a data-not-instructions clause; state fields referenced by path | J1-6 |
| A7 | §6 line 141 | Escalation band unset at launch; "high-severity" is deterministic; second opinion on maintainer request only, until a measured cascade exists | J1-8 |
| A8 | §6 line 144, 146 | Audit sample stratified across ranks with stored inclusion probabilities; blind-label rule requires annotation hidden at disposition or the flag recorded; labels weighted by inclusion probability | J1-9 |
| A9 | §6 line 148; REQ-30 | (W) "calibration unchecked on Scout data; rank-only"; show buckets not numbers until the reliability criterion (CC:40-42) is met | J1-10 |
| A10 | §6 line 146 | (W) disclose `ranking: jev (unvalidated)` until X1 is recorded; default stays (Appendix 4) | J1-11 |
| A11 | §5.1 line 111 | `Jev Choice where semantic` → `one Noul per pattern, fanned out in one request, per-pattern named thresholds` | J1-12 |
| A12 | §13 X4 (line 370) | Add dev/test by repository, enriched positives (≥ 30 per pattern or `unmeasured`), a random non-hit slice for recall (IPW), recall minimum | J1-13 |
| A13 | §14 A5 (line 386), §5.1 line 111 | Cost to be measured as billed per function; per-run cap in tokens/cost, not calls | J1-14 |
| A14 | §8.3 line 223; §13 X5 (line 371) | Primitive: Noul `is_fix` with tuned threshold; state shape stated; sample sized so Wilson LBs are reachable (≥ 80-100 positives); report error by path role/repo; cite E1b G4 `fix` precision 0.80 [0.63, 0.91] as prior | J1-16, J1-17 |
| A15 | §13 X3 (line 369) | Candidate retrieval in code (or the full feature list when ≤ 255 options, per the Choice docs); `none` via a companion Noul read before the Choice, not a `none` option; options order permuted; define "`none`-exclusive"; Wilson LB; dev/test; candidate-coverage metric | J1-18 |
| A16 | §13 X10 (line 376) | Noul over a pair, both orders; label guide; double labelling; Wilson LB; strata recorded | J1-19 |
| A17 | §13 X1 (line 367) | Define `k`, budget, de-correlation, split unit, paired repo-clustered interval, sizing; "injected true defects"; label all critical-severity findings below the floor | J1-9, J1-20, J1-21 |
| A18 | §9 item 7 (line 275) | Per-rule AUROC only above a label-count minimum, else `insufficient` | J1-22 |
| A19 | Phase 5 acceptance (line 336) | Add: model bump marks annotations stale; order-flip rate reported; failure/cap cases ranked deterministic and disclosed | J1-4, J1-5, J1-7 |
| A20 | §12 R10 (line 357) | Add mitigation: version key, stale-on-bump, stratified audit | J1-3, J1-9 |

---

## B. Requirements for the Phase 2 `internal/decisions` implementation

No code exists yet. These are requirements for the adapter and the stores it feeds (skill `CI`, `TT`, `BED`, `EA`).

1. **Typed questions** in code as data: `{type: noul|choice|score, instructions, criteria, state_shape}`; named fields in state; backticked path references; untrusted-state clause; Choice option order recorded.
2. **Version key** = SHA-256 over canonical `{instructions, criteria, state_shape, option_order, resolved_model_id}`; `question_id` is for code only; the exact request body is stored (redacted) next to every annotation (`BED:75`, `CI:35`).
3. **Pinned model** from policy; reject aliases (`jev-latest`, `jev-preview`) at config load; record the response's reported model; compare it to the pinned ID on every call; mismatch ⇒ discard, unit `incomplete`, counter; a configured upgrade path creates a new version key.
4. **One request per unit** with all independent questions fanned out (`CI:11-13`); a second request only when an answer is needed to build the next state.
5. **Failure semantics** (`CI:39-40`): network error, HTTP 4xx/5xx, 429 (honour `retry-after`), timeout, schema-invalid answer, missing key, fork PR, egress off, cap reached ⇒ no annotation row, per-finding disclosed `ranking: deterministic (reason)`; never a zero or a low score.
6. **Cap** `policy.decisions.max_per_run` expressed in calls and billed cost, chosen in deterministic-baseline order; record billed `usage` per call.
7. **Annotation row** fields: finding id, question id, version key, resolved model, answer (probabilities, never only the label), run id, timestamps, state digest, `ranking` provenance, audit stratum and inclusion probability, flag whether it was shown at disposition time.
8. **Thresholds in config, never literals** (`TT:90-92`): `decisions/thresholds.json` (or the policy section) with per-question threshold, minimums, metric, label-set summary, version key, score space (`raw`), status `provisional|ready`, cost rule, assumptions. A key mismatch ⇒ stale ⇒ disclosed fallback/`provisional`.
9. **Composition in code**: the rank is a named formula over per-question outputs; changing weights re-ranks stored answers with no model call (`CI:29-30`); the deterministic baseline stored for every finding.
10. **Data feed for self-eval**: `scout selfeval triage` reads annotations + blind dispositions with inclusion weights; dev/test split by repository; test scored once per shipped version (`BED:62-69`); `scores.csv`-compatible export (`case_id, label, score, model, question_version, split, stratum`) so `scripts/tune_threshold.py` can run unchanged.
11. **Tests**: a forced model mismatch; an injection fixture (an excerpt asserting "false positive") changes rank only and is audited; 429/timeout never yields a low rank; a version-key change marks thresholds stale; order permutation harness reports a flip rate.
12. **Escalation**: no confidence-routed band at launch; if added later, per question, from saved scores, with the fallback measured on the same cases (`EA:36-52`).

---

## C. Lessons only (frozen studies; no edit proposed)

- **E1b-v2 G5/TAU:** a labeller threshold (0.5) was fixed a priori; saved scores show no threshold meets precision ≥ 0.70 and recall ≥ 0.60 together (table in E1-1). Future labellers: tune from saved scores with F-beta, keep dev/test, and report the interval on IPW recall; one repetition set per version key.
- **E1b-v2 state:** an outcome like "completes something the earlier change omitted" needs the diff or call-site evidence in state; titles/bodies/paths are insufficient (E1-2).
- **E1b-v2 G4:** `other_dev` as a residual class produced most disagreement; for a control variable the cost was noise, but a future classifier whose output is a predictor should avoid a residual class, or split it.
- **E1b-v2 G4 as prior for X5:** `fix` precision 0.80 [0.63, 0.91] at the PR level (n = 30) (A14).
- **E2:** split by repository/pattern, not row; keep a test set unseen by anyone writing question text; do not put rule-specific examples that mirror labelled cases into criteria; blind second labeller (E2-1..3). For inertia-bench: record the question version key, resolved model, `qhash`, and untrusted-state clause in every Jev run script.
- Neither study checked option order, calibration, or the cascade; none was designed for them (pre-rules or out of scope).

---

## D. Questions for the user (skill STOP gates)

The skill requires one message with the answers below before any case collection, threshold tuning, or Jev question wording for a gating decision (`BED:20-24`, `TT:33-46`; QD:10-13). I have assumed nothing. Per-decision draft questions are listed only here (`QD:12-13`).

**Decisions covered:** (1) triage ranking `P(real_defect)`; (2) rubric presence, 12 patterns (X4); (3) `fix` classifier (X5); (4) unowned-file → feature suggestion (X3); (5) competing-implementation pair (X10); (6) Jev-uncertain high-severity escalation.

For **each** decision:
1. **Action:** what happens on a positive, who bears it. (My reading: triage = reorder and disclose, no state change; rubric = report only; X5 = enables `fix_touches`; X3 = reconciliation aid; X10 = exploratory. Is that right?)
2. **Baseline:** deterministic rank (triage), Semgrep (rubric), regex (X5), path heuristics (X3), clone-group alone (X10) — confirm, or name another.
3. **Costs:** what a false action and a miss each cost (e.g. critical issue ranked low; maintainer time on a false rubric hit).
4. **Written policy:** the plain-words rule with exclusions and tie-breaks (e.g. is a mutable action tag a `real_defect` regardless of workflow privilege; is a test fixture always `false_positive`?).
5. **Minimum precision and recall** with reasons (without an answer the skill assumes 0.8/0.8, raised to 0.9 on the costlier side, status `provisional`, `TT:43-46`).
6. **Conservative bounds:** may the minimums be held at the 95 % lower bound (`--conservative`), which design X4/X5/X9 already do but X3/X10 do not?

Additional:
7. **Escalation:** acceptable accuracy loss for a second opinion (`EA:44-45`); without it no band is set.
8. **Who labels — answered 2026-10-22 (`docs/STATUS.md` §6.8):** independent agent sessions do a blind first pass: two with no shared context, one Anthropic via `claude -p` and one OpenAI via `codex exec`, following a frozen per-question guide, with a third adjudicating session where they disagree. The user then reviews every case and confirms or changes it. Labels are `agent_reviewed` before review and `human` after. First batch: the 52 E2 triage cases (`evaluation/jev-triage-labels-v1/`). Volumes still to plan: ~100 (X1), ~200 (X3), ~200 plus ≥ 100 `fix` positives (X5), 200 (X10), and 12 × ≥ 30 positives (X4).
9. **Logs and corpora:** access to PR/commit and findings history for replay (`BED:43-47`) beyond the 22-case E2 set and the 12 E1b repos.
10. **Rulings on ambiguous cases** you already hold (e.g. `cannot_tell` handling, whether tests-only changes count as `fix`).
11. **User-only report fields** for `thresholds.json`: label provenance, who reviewed, and a decision on whether `unknown` is acceptable for any left blank (`TT:33-37,46-47`).
12. **Model pin:** which Jev version Scout pins for launch (1.13 or 1.14; same price per `models.md`, E2 saw no difference at n = 22; no jaggedness page exists for 1.14, so its weak spots are undocumented) and whether the OpenRouter `alpha/decisions` route is acceptable for production use. The documented endpoint is `https://api.typesafe.ai/v1/systemone`; the OpenRouter route, its dated slugs and their retention are in no TypeSafe doc `[INFERENCE on stability]`. Should production use the TypeSafe endpoint directly, or does REQ-31 require OpenRouter for Jev?
13. **Shadow vs on:** you decided Jev is on from launch (Appendix 4). Confirm that the ranking is acceptable as `provisional` with disclosure until X1 (J1-11), or say if it should start in shadow.

---

## Summary of top findings

1. **X1's decision rule is not executable as written (J1-20):** no definition of `k`/budget/split unit, Wilson used on a paired difference, and ≥ 30 positives cannot give LB > 0 at a 10-point gain; the 5 % audit sample cannot test "no critical miss" (J1-9).
2. **Version key and model pinning are incomplete (J1-3, J1-4, J1-5):** define the key as a hash including state shape, option order and resolved model, mark thresholds stale on change, and specify per-finding failure semantics.
3. **Primitive choices:** rubric lane should be 12 Nouls not a Choice (J1-12, J1-13); X10 is a Noul over a pair with both orders (J1-19); X3 needs retrieved candidates, not one Choice over all features (J1-18); `Noul primary` was rejected on a bundled single Noul whose ranking AUROC matched Choice (J1-1).
4. **X5's 0.85 precision bar is at risk (J1-16):** E1b G4 `fix` precision is 0.80 [0.63, 0.91] and a 200-commit sample cannot reach the Wilson LBs unless observed recall is ≈ 0.85+.
5. **The triage question bundles judgments code can decide (J1-2)** and E2's n = 22 evidence has split, labelling and leakage problems (E2-1..3).
6. **E1b-v2 (lesson only):** TAU 0.5 untuned; saved-score recomputation shows no threshold satisfies the G5 minimums (E1-1); likely missing diff evidence in state (E1-2).
7. **Escalation and calibration wording (J1-8, J1-10):** no band exists, none should be set at launch; replace "not calibrated" with "unchecked, ranking only".
