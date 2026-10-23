# Draft notes — revision 2.2 proposal

Produced for the innovation-council review. Round 1 (25 issues) was resolved in `round-1-response.md`; the per-change line ranges below were written for round 0 and are still correct (line counts unchanged), except where the round-1 response gives newer wording (C1 Clawpatch provider binding at lines 101, 120, 150, 337, 346; Phase 3b and §8.7 principle 5 at 291, 365–366; X1/X3/X4/X5/X10 at 407–414; §6 audit, calibration and labels at 160–166). Not yet signed off.

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` (byte copy of `docs/scout-design.md`, rev 2.1) | 417 | `7ff21a7fba70a5e3af7a735b0aa40f1ad654d17714df8c3549da6ebcf89b89c5` |
| `scout-design.proposed.md` (rev 2.2, after round-5 response; C7 split into C7a/C7b) | 462 | `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4` |
| `proposed.diff` (`diff -u base proposed`) | — | 98 insertions, 53 deletions |

`docs/scout-design.md` is untouched. `docs/hypothesis-ledger.md` was edited directly (C2; listed below) and is part of the adopted change set. Line numbers below refer to `scout-design.proposed.md`.

## C1 — Model-CLI provider

Changed:
- Header summary item (1): line 5.
- §1 judgment-placement bullet: 19. Evidence-notation paragraph unchanged by C1.
- §3.1 component list: `contextual` via model-CLI adapter, line 61. §3.4: `requires` set `provider-egress` → `model-cli`, `skipped (no model CLI)` never `clean`, Jev lane health pointer, line 88.
- §4: new "Model CLI" bullet (agent refresh `skipped (no model CLI)`, heuristic map stays canonical), line 101.
- §5.1: lane 9 row, line 120; §5.3 matrix row, line 135.
- §6: owner-table row for reasoning LLMs, line 147; new "Model use (REQ-31; R-036)" paragraph, line 150; launch-default paragraph (key from Infisical / CI secret on trusted runs, model-CLI lanes capture their own consent), line 164.
- §7.1 nightly trigger wording, line 176. §7.4 trust boundary (no Anthropic/OpenAI keys; only secret is the Jev key, trusted runs only; model-CLI lanes self-hosted/local), line 195.
- §10.2: `internal/contextual/*` split out of the Keep row with provider rewrite, lines 329–330; `scout-nightly.yml` row drops `OPENAI_API_KEY`/`ANTHROPIC_API_KEY` (`scout-nightly.yml:172-173`), line 342. §10.3: model-CLI adapter and Jev provider in the adapter-contracts list, line 346.
- §12: new R13 (line 398). Appendix 8 (line 455).

Left out: no change to Phase 1–6 scenarios for C1 (out of scope per PROPOSAL); the Jev endpoint is not changed (kept on OpenRouter, REQ-31), only disclosed under C7/R14.

## C2 — Research duties move to inertia-bench

Changed:
- Header summary item (2): line 5. Evidence-notation paragraph: inertia-bench defined, evidence for design changes limited to Scout-experiment-rule or bench-rule studies plus grandfathered E1b-v2, line 7. SCB/FXL definitions kept: they are still cited for rubric patterns and scoring (§5.1, §9), not as X8 venues, which the base text never claimed.
- §1: inertia-monitoring bullet (pooled analysis run by the bench), line 18; new "Scope (revision 2.2)" paragraph, line 23.
- §2 "Not followed" paragraph: `§15 item 5` → `item 3`, line 43.
- §8.5: limit (5) no longer cites X8, line 267; `§15 item 5` → `item 3`, line 269; the "A further test, by strength (a)–(d)" paragraph replaced by "Further tests belong to inertia-bench", line 271; claim-upgrade rule kept, qualifying study may be signed off under Scout's or the bench's rules, causal claims need the same signing-off route, line 273.
- §8.6: `scout export --study` stays a product feature, Scout-owned consent (§15 item 1), pooled analyses run by the bench, line 279.
- §12 R4 (X8 independence removed), line 389.
- §13: X7 and X8 removed with a note that IDs are not reused, line 416. §14: A8 removed, with A10 (C4); gap note line 432. §15.2 removed (see C6).
- Appendix 11 (line 458).
- **`docs/hypothesis-ledger.md`** (edited in place, adopted with this change set): HA2 Tests `X8` → `inertia-bench` and date → 2026-10-22 (line 22); two history lines dated 2026-10-22 (lines 32–33): HA2 tests change, and the freeze of H\*/HA\*/HQ\* per `AGENTS.md` (HD\* and product hypotheses stay live).

Left out: HA1's Tests cell still reads "X8 tests HA2 only" (ledger line 21) because it is a frozen historical record and the 2026-10-21 history line explains it; changing it is outside the C2 scope (HA2 only). Council may decide otherwise.

## C3 — Subject-interface principles (new §8.7)

Changed: §3.1 `subject` component, line 66; new §8.7 with six principles (snapshot-only, offline, deterministic, never writes into the snapshot, explicit missingness with `unmeasured_reason` mapping and `scout.crap@1` → `not_applicable`, blind), lines 283–294; Phase 3b entry, scenario and acceptance, lines 364–366; §10.2 `subject` package and `cmd/scout` `subject capabilities|measure`, lines 338, 341; §10.3 pointer, line 346; Appendix 10 (line 457).

Left out (deliberately):
- The schema `ib-subject-*/1` and the exact `unmeasured` reason mapping: separate council round after inertia-bench's design signs off (bench review round 6: issues found). §8.7 lists the bench's closed reason set only "as it stands in the unsigned bench design".
- History/PR measures (§8.3): out of the first version.
- Phase 3b is not given a Phase number between 3 and 4 in other text (e.g. no change to the Phase 4 entry).

## C4 — §15.3 decided

Changed: §10.1 acceptance, `[ASSUMPTION target]` replaced by `[DECIDED 2026-10-22, Appendix 14]`, "development Mac", line 323; §14 A10 removed (gap note line 432); §15.3 removed; Appendix 14 (line 461); header summary item (4), line 5.

## C5 — E1b-v2 grandfathered

Changed: evidence-notation paragraph, line 7 (E1b item: OpenRouter annotators incl. Gemini before the model-use rule; stays citable; `evaluation/change-cost-v2/RESULTS.md:5`; `AGENTS.md` design rule 3); Appendix 9 (line 456). No other change to E1b citations.

## C6 — Renumber §15; Appendix

Changed: §15 now has items 1 (study-export consent, with Scout-owned consent and bench-run pooled analyses), 2 (was 4, PR metadata), 3 (was 5, inertia framing), lines 436–440; closing sentence explains items removed, line 442; every in-document reference to "§15 item 5" updated to item 3 (lines 43, 269) and the new reference to item 1 (line 279). Appendix heading extended and decisions 8–15 appended, lines 446, 455–462.

## C7 — Jev design sharpened (review §A, A1–A20)

| Review item | Where (proposed file) | Treatment |
|---|---|---|
| A1 Noul-primary wording | 168 | Applied (E2 AUROC parity cited, no "Noul worse" claim). |
| A2 code-computed facts, narrow questions | 146, 155 (State and questions bullet) | Applied. The example question names in the review (`mechanism_present_in_excerpt` …) are not written into the design; X1's preregistration fixes the questions. |
| A3 version key | 153 | Applied (sha256 over instructions, criteria, state shape, option order, resolved model; request body stored; stale on key change). |
| A4 pinned model, stale on bump | 154; 374 | Applied. **Which Jev version to pin at launch is not stated** (depends on review §D q12, unanswered). |
| A5 per-finding failure semantics → lane health | 158–159 | Applied: 429/529 backoff, 400-class non-retryable, cap-hit `partial (n/N)`, lane health `ran / skipped (reason) / failed`, optional-lane exit-code rule. |
| A6 data-not-instructions clause | 155 | Applied; sanitizer explicitly not an injection defence. |
| A7 escalation band unset | 147, 161 | Applied. Accuracy-loss value left to the preregistration (§D q7 unanswered). |
| A8 stratified audit, blind-label rule | 160, 166 | Applied (strata, inclusion probabilities, inverse-inclusion weights, annotation-visible flag, only blind or audit dispositions count). Rev 2.1's 5 % low-ranked sample retired; sample fractions left to X1 preregistration. |
| A9 calibration wording | 162, 168 | Applied ("calibration unchecked on Scout data; ranking only"; rank bucket not number; three number-reading uses named). The reliability thresholds (bin tolerance, cases per bin) are not stated. |
| A10 disclosure | 164 | Applied (`ranking: jev (unvalidated, X1 pending)`; Appendix 4 unchanged; review §D q13 treated as a disclosure only, no shadow mode). |
| A11 rubric primitive | 117 | Applied (one Noul per pattern, one request per function, per-pattern named threshold, `unmeasured` patterns out of lane). |
| A12 X4 | 410 | Applied (dev/test by repository, enriched positives, non-hit slice with IPW, recall minimum required). Numbers (positive count, recall minimum) left to preregistration. |
| A13 cost | 117, 157, 426 | Applied (billed per function; cap also in billed cost). $0.12 stays `[ASSUMPTION]`; no cost bound value invented. |
| A14 `fix_touches`, X5 | 243, 411 | Applied (Noul `is_fix`, state shape, sample sized by preregistration, G4 prior 0.80 [0.63, 0.91], error by path role and repo). X5's 0.85/0.8 and LB 0.75/0.7 retained as rev 2.1 values pending the user's minimums (§D q5/q6). |
| A15 X3 | 409 | Applied (retrieval in code or full list ≤ 255, companion Noul for `none`, options permuted, LB, dev/test, candidate coverage; "none-exclusive" defined). Recall minimum left to preregistration. |
| A16 X10 | 414 | Applied (Noul over pair, both orders, label guide, double labelling, Wilson LB 0.8, strata, no prevalence claim). |
| A17 X1 | 407 | Applied (split unit, `k`/budget unit defined, de-correlation defined, paired repository-clustered bootstrap, injected true defects, all critical below floor labelled, power-calculation sizing). `k` and sample sizes left to the preregistration. The ≥ 10-point gain and LB > 0 are rev 2.1's values, kept. The ≈ 0.08 SE statement is tagged `[INFERENCE]` as in the review. |
| A18 §9.7 | 308 | Applied (pooled; per-rule only above a minimum fixed by X1 preregistration, else `insufficient`). |
| A19 Phase 5 acceptance | 373–374 | Applied (stale on bump, failure/cap disclosure, order-flip rate). |
| A20 R10 | 395 | Applied. |
| Labelling (STATUS §6.8) | 166; 407, 409–411, 414; 373; Appendix 15 (462) | Protocol stated once in §6 "Labels" and referenced from X1/X3/X4/X5/X10 and Phase 5. X1's "blind labels" and §9.7's "blind post-launch labels" restated; maintainer dispositions kept as a separate `human` stream. Provenance `agent_reviewed` → `human`. |
| Jev endpoint risk (review §D q12) | 399 | R14 added; OpenRouter kept per REQ-31; whether to use the TypeSafe endpoint instead is declared a user decision not taken. |

Deliberately left out of rev 2.2 because they depend on unanswered §D questions: minimum precision/recall per decision, conservative-bound policy for X3/X10 beyond the review's Wilson-LB wording (applied only where A15/A16 state it), costs of false actions, written policies per question, escalation accuracy loss, Jev model pin, X1/X3/X4/X5/X10 sample sizes and `k`, thresholds' `provisional` handling beyond "stale → disclosed fallback". No numeric minimum was invented; where numbers appear they are rev 2.1's existing values or the review's tagged inferences.

Not applied from the review: its sections B (implementation requirements; implementation work, not design text beyond what §6 now states), C (lessons for frozen studies) and the E2/E1b findings (no edit to those studies).

## Whole-document consistency pass

Searched the proposed file for X7, X8, A8, A10, §15, `provider-egress`, "provider/decision secrets", "Noul primary", "not calibrated", "blind post-launch": remaining hits are the new removal notes (lines 416, 432, 442), the header/Appendix summaries and the updated `§15 item` references. Mentions of "§15.1/§15.2/§15.3" in the base were limited to the sections now rewritten. A6 (decision budget 2,000/run) is unchanged. Appendix 1–7 are byte-identical to the base (diff confirms).

## Checks for the main agent

`shasum -a 256 docs/scout-design.md docs/design-reviews/2026-10-22-rescope/scout-design.base.md` must match `7ff21a7f…`; `diff -u` base proposed must reproduce `proposed.diff`; the council's final round must name the proposed file's SHA-256 (see the table at the top of this file; any later edit changes it). Round 2 (9 issues) is resolved in `round-2-response.md`; its edits are at lines 3, 155, 157, 160, 166, 300, 308, 373–374, 407, 410, 414 plus six rows of `SCOUT_REBUILD_REQUIREMENTS.md`.

Round 3 split C7: C7a (adapter, versions, failure handling, primitives, calibration wording, §6.8 protocol, R14, four audit/preregistration principles) stays in 2.2; C7b (X1/X3/X4/X5/X10 estimator, split, interval, population mechanics; interim sampler mechanics; §9.7 inputs; Phase 5 replay) is deferred to the preregistrations. §13 rows X1/X3/X4/X5/X10 are at lines 407–414 in short form; §6 audit/labels at 155–166. The carry-forward list is in `round-3-response.md`. Round 4 (7 issues) is resolved in `round-4-response.md`; its edits touch lines 157, 160, 162, 166, 374, 409. The per-row "Applied" marks for A8/A12/A13/A16/A17/A18 above are superseded by this split.
