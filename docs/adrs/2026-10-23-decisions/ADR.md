# ADR-003: Record user decisions and align the design with shipped behavior

- **Status:** partly superseded by ADR-007; amended by ADR-010 (triage minimums), ADR-012 (readiness observations/eligibility, maintainer triage history, trusted readiness publication)
- **Design revision:** 2.3 (adopted 2026-10-23)
- **Decided:** 2026-10-23; **sources:** R-039, user direction in `docs/STATUS.md`@9a3d19f5 §6.9

## Decision

- Confirm the separate opt-in study export of pseudonymous scope aggregates and a sanitised change table, with Scout's dogfood ledger as the first pooled dataset.
- Keep code quality as the primary family and frame principal series as descriptive “structure to watch”, not cost.
- Split reusable workflows into read-only, secret-free PR runs and trusted default-branch runs with state publication, Jev and `pull-requests: read` for PR metadata.
- Pin Jev to `typesafe/jev-1.14-20261014` through OpenRouter `alpha/decisions` and launch provisionally with per-finding disclosure until X1 reports.
- Generate Jev labels through blind independent `claude -p` and `codex exec` sessions plus adjudication, without per-case user review.
- Derive attribution from a paired merge-base recompute when the baseline is absent or incomparable, rather than treating every digest mismatch as `unknown`.
- Classify otherwise-`introduced`, non-`guardrails.*` findings as `unknown` when the base did not inspect their family, and exclude them from `would_block` and the gate.
- Replace Phase 4's Jaccard ≥ 0.8 acceptance bar with non-gating reports of area-ID retention and retained-area Jaccard under the existing ≥ 0.5 matching rule.

## Informed by

- R-039 records the user's decisions on export, PR metadata, framing, Jev pinning, provisional launch and delegated labelling.
- `PROPOSAL.md` D4 records coverage-aware attribution on `main` in `480949a` and pinned scanner provisioning in `404f69b`.
- `round-3-response.md` records both changes in release `v0.1.2` and successful scanner provisioning in run 39966674494.
- `round-1-response.md` records 7,134 unchanged axum series keys after an agent refresh and a content-only commit, with none added or removed.
- `PROPOSAL.md` D6 records area-ID retention of 0.95 on fastapi and 0.85 on axum as scenario observations, not X2 results.
- `PROPOSAL.md` D2 cites the Jev adapter audit and leaves Design-level rows 5, 35 and 38 open.
- This revision cites no new experiment.

## Changes

- Amend ADR-001's provisional framing and ADR-002's revision 2.2 by closing `scout-design.base.md` §15 items 1–3 and adding Appendix decisions 16–19 in `scout-design.proposed.md` (signed revision 2.3).
- Amend ADR-000's Jev launch policy with the model pin, provisional disclosure and tuning defaults, without changing route/rank-only authority.
- Replace `scout-design.base.md` (revision 2.2) Appendix 15's per-case human labelling review with agent-only adjudication.
- Preserve the 52 user-confirmed triage cases as `human`, with human labels taking precedence for the same unit and question.
- Amend `scout-design.proposed.md` (signed revision 2.3) §§7.2–7.4, Phase 2 acceptance and A2 to make recomputed bases comparable and count a PR run once toward the ≤ 10 % unknown-attribution limit.
- Amend `scout-design.proposed.md` (signed revision 2.3) §§6, 8.5–8.6, 10.2, 11 and 15, and the X1/X3/X4/X5/X10 label-protocol references.
- Change Phase 3's scenario from lowering `ruff` strictness to adding a `ruff` ignore without changing the allow-list.
- Update `SCOUT_REBUILD_REQUIREMENTS.md` REQ-25, REQ-27, REQ-29, REQ-35, REQ-36 and its open-decisions list through the signed `requirements.proposed.diff`.
- ADR-007 later removes this revision's legacy-baseline recompute when the manifest lacks `uninspected` (`docs/scout-design.md`@9a3d19f5 Appendix 22(b)).

## Conditions

- Jev tuning minimums remain provisional at 0.8/0.8, raised to 0.9 on the costlier side, until the user sets costs and minimums.
- The 0.9 triage-recall judgement remains an assumption and gates nothing while ranking is ordinal.
- Defaults substitute only for precision/recall point minimums, not X1's gain threshold or X4/X5 lower bounds.
- Jev adds question `kind` to the version key and a $0.50 assumed billed-cost cap, stopping when billed cost is unreported.
- Area-ID retention counts kept areas and does not establish X2's matched-area survival or change X2's decision rule.
- The final review requires matching recomputed digests and a reproducible design diff before adoption.
- The historical header leaves `pr_effort` in CI unobserved and the per-finding base-gap clause covered by unit tests, not yet the Phase 2 scenario.

## Review

- Seven rounds: 16 → 9 → 6 → 5 → 3 → 1 → 0 issues.
- Final sign-off: `round-7.md`; signed digests: `scout-design.proposed.md`: `75ef1e66521b1f129e872f651bd0f8bb9989e38867fb5335ec5ad43158fb3061`; `proposed.diff`: `73da0f865fb72158ee5ee3e367b6b5dfcb7f8e0945c2e10144f121145c1eb6c1`; `requirements.proposed.diff`: `4e7adaccaa3ffe0e627fdd06c068a8469afbefc5a6dd09d985ca73803ab927c3`; base: `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4`.
- Adoption record: `docs/STATUS.md`@9a3d19f5 §1, revision 2.3 history entry; `round-6-response.md` records the final fix, not adoption.

## Records in this folder

- `ADR.md`: decision summary and record index.
- `PROPOSAL.md`: D1–D8 changes, sources, exclusions and adoption procedure.
- `DRAFT-NOTES.md`: draft digests, line map and judgement calls.
- `scout-design.base.md`: revision 2.2 baseline snapshot.
- `scout-design.proposed.md`: signed revision 2.3 snapshot.
- `proposed.diff`: signed design amendment against the baseline.
- `requirements.proposed.diff`: signed companion requirements amendment.
- `round-1.md`: council review with 16 issues.
- `round-1-response.md`: first fixes and choices for the second draft.
- `round-2.md`: council review with nine issues.
- `round-2-response.md`: nine fixes and newly committed attribution and CI facts.
- `round-3.md`: council review with six issues.
- `round-3-response.md`: attribution qualifiers, pooled Jaccard reporting and release evidence.
- `round-4.md`: council review with five issues.
- `round-4-response.md`: area-ID retention terminology and four wording fixes.
- `round-5.md`: council review with three nits.
- `round-5-response.md`: artifact-output wording and review-label fixes.
- `round-6.md`: council review with one stale-label nit.
- `round-6-response.md`: final label fix and unchanged design-artifact digests.
- `round-7.md`: zero-issue sign-off and conditional adoption instructions.
