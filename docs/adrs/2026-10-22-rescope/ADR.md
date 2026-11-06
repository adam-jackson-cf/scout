# ADR-002: Re-scope Scout to a product and an inertia-bench subject

- **Status:** partly superseded by ADR-003, ADR-005, ADR-007, ADR-009; amended by ADR-008 (X11 added to the experiment list)
- **Design revision:** 2.2 (adopted 2026-10-23)
- **Decided:** 2026-10-22/23; **sources:** R-036, R-037, R-038; user direction in `docs/STATUS.md`@9a3d19f5 §6

## Decision

- Scout runs product experiments X1–X6, X9 and X10; inertia-bench owns codebase-inertia hypothesis studies.
- Reasoning-LLM lanes use subscription-authenticated `claude -p` or `codex exec` with Anthropic/OpenAI models; CI carries no Anthropic/OpenAI API keys.
- Jev stays on OpenRouter, with Infisical supplying the local key and a GitHub secret supplying it only on trusted CI runs.
- Scout owns an offline, deterministic, snapshot-only subject interface in its binary, implemented after Phase 3 measures.
- The Phase-1 bar is `scout assess` on fastapi `d666ccb`, heuristic map, egress off, in less than 10 minutes on the development Mac.
- Jev annotations carry question, model, rank-formula and baseline versions, disclose failures and capped coverage, and remain advisory.
- Jev uses probability-sampled blind audits with stored inclusion probabilities; Jev-quality lines remain `insufficient` until X1's preregistration is signed off.
- This revision required two blind independent agent labelling sessions, an adjudicator and per-case user review; ADR-003 later removes the user-review step.

## Informed by

- R-036 records the user's product-only scope and model-use direction of 2026-10-22.
- R-037 supplies inertia-bench subject-interface context; this revision adopts principles rather than its unsigned exact schema.
- R-038 and `jev-actions-review.md` inform question versioning, primitive choices, failure semantics, blind audits and calibration wording.
- The Jev review distinguishes E2's rejection of a single-Noul bundle for three-way routing from ranking quality; E2 did not show Noul worse for ranking.
- No new experiment supports this revision.
- E1b-v2 remains citable despite its earlier OpenRouter annotators, including Gemini; its frozen preregistration and scripts stay untouched.

## Changes

- Amends ADR-001's revision 2.1 across `scout-design.proposed.md` (signed revision 2.2) §§1, 3–8, 10–15 and Appendix decisions 8–15.
- Replaces the HTTP contextual provider with model-CLI integration and constrains Clawpatch to its `codex` or host-authenticated `claude` provider.
- Optional lanes without a logged-in model CLI report `skipped (no model CLI)` in health, summaries and readiness without changing the exit code.
- Removes `scout-design.base.md` (revision 2.1) §8.5's further-test list, X7, X8, §15.2 and A8; experiment IDs are not reused.
- Keeps the claim-upgrade rule and consented `scout export --study`; inertia-bench runs pooled analyses.
- Moves HA2's test home to inertia-bench and freezes H*/HA*/HQ* as historical records while product hypotheses stay live.
- Adds `scout-design.proposed.md` (signed revision 2.2) §8.7 and Phase 3b subject-interface principles; the first version has no coverage-artifact channel, so `scout.crap@1` is `not_applicable`.
- Replaces the 5% low-ranked-only audit with a probability-sampling principle and a provisional 5% total audit fraction.
- Synchronises `SCOUT_REBUILD_REQUIREMENTS.md` as adopted with revision 2.2 (recorded in `PROPOSAL.md`) REQ-29, REQ-31, REQ-35, REQ-36, REQ-38, REQ-39, REQ-41 and §10.
- Amend ADR-000's specification-cutover timing while preserving `scout-design.proposed.md` (signed revision 2.2) Appendix decisions 1–7: keep the requirements file live until Phase 6 (Appendix 12). ADR-009 ends that plan and places requirements in `docs/scout-design.html` Appendix A.
- ADR-003 later decides the remaining export, PR-metadata and framing questions and amends Jev launch policy and labelling (`docs/scout-design.md`@9a3d19f5 revision 2.3 summary and Appendix 15–19).
- ADR-005 later narrows agent mapping to `codex`, `gpt-6-luna`, effort `xhigh` (`docs/scout-design.md`@9a3d19f5 Appendix 8 and 20).
- ADR-007 later amends the model-CLI design by removing Scout’s `CLAWPATCH_PROVIDER` use.
- ADR-008 later amends the experiment list by adding X11 for merge-time association (HD8).
- ADR-009 supersedes the requirements-file lifetime in `docs/scout-design.md`@9a3d19f5 Appendix 12 and replaces the Phase-6 specification cutover with `docs/scout-design.html` Appendix A. It also replaces the experiment list "X1–X6, X9, X10" with REQ-42's X1–X6, X9–X11.

## Conditions

- The exact subject schema and missingness mapping require a separate council round after inertia-bench's design signs off.
- History and PR measures stay outside the first subject-interface version.
- X1/X3/X4/X5/X10 require council-signed preregistrations before Jev-versus-label comparisons, thresholds, bands or default changes.
- `round-3-response.md` lines 22–30 bind those preregistrations to the deferred estimator, population, split, interval, sampler and replay questions.
- Round 6 also carries forward what X1 must capture before interim audit draws count.
- The 2,000-call cap schedules triage first in deterministic-baseline order, then other question families in stable unit-key order, with unreached coverage disclosed as `n/N`.
- Only the call cap applies until Phase 5 measures billed usage and sets a cost bound.
- Adoption required the main agent to confirm the proposed and base digests and reproduce `proposed.diff`.

## Review

- Six rounds: 25 → 9 → 12 → 7 → 2 → 0 issues.
- Round 3 splits adapter and audit principles into C7a and defers experiment mechanics as C7b.
- Rounds 4–6 close write-time audit capture, cap scheduling and requirements-synchronisation gaps.
- Final sign-off: `round-6.md`; signed digests: `scout-design.proposed.md`: `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4`; `scout-design.base.md`: `7ff21a7fba70a5e3af7a735b0aa40f1ad654d17714df8c3549da6ebcf89b89c5`.
- Adoption record: `PROPOSAL.md` records signed-off adoption on 2026-10-23; no separate round-6 response appears in this folder.

## Records in this folder

- `ADR.md`: this decision summary.
- `PROPOSAL.md`: adopted scope, C1–C7 change set and council procedure.
- `DRAFT-NOTES.md`: change locations, digest table and implementation-review context.
- `jev-actions-review.md`: advisory Jev design and implementation audit that informed C7.
- `scout-design.base.md`: frozen revision 2.1 comparison baseline.
- `scout-design.proposed.md`: signed-off revision 2.2 design snapshot.
- `proposed.diff`: base-to-proposed design changes.
- `round-1.md`: first council review with 25 issues.
- `round-1-response.md`: fixes for the first review.
- `round-2.md`: second council review with nine issues.
- `round-2-response.md`: fixes for the second review.
- `round-3.md`: third council review with 12 issues.
- `round-3-response.md`: C7 split, fixes and binding preregistration carry-forward.
- `round-4.md`: fourth council review with seven issues.
- `round-4-response.md`: audit-capture, acceptance and cap-order fixes.
- `round-5.md`: fifth council review with two remaining nits.
- `round-5-response.md`: complete cap scheduling and interim-only independence fixes.
- `round-6.md`: zero-issue sign-off, adopted companion changes and digest-confirmation condition.
