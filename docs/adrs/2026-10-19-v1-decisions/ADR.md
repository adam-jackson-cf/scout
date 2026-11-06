# ADR-000: V1 baseline decisions: map refresh, baseline branch, allow-list, Jev, language tiers, spec cutover, Clawpatch pin

- **Status:** partly superseded by ADR-004, ADR-005, ADR-006, ADR-007, ADR-008, ADR-009; amended by ADR-001, ADR-002, ADR-003
- **Design revision:** v1 baseline decisions, carried into v2 (adopted 2026-10-19)
- **Decided:** 2026-10-19; **sources:** user decisions recorded in `docs/archive/scout-design-v1.md`@9a3d19f5 §15 and `docs/scout-design.md`@9a3d19f5 Appendix items 1–7

## Decision

- Run both mappers at onboarding and, originally, on debounced default-branch merges that change monitored areas; ADR-006 supersedes this trigger clause.
- Keep the areas reconciliation layer and never run the agent map on PR runs.
- Store the baseline on the protected `scout/state` git branch.
- Adopt `docs/scout-allowlist.md`@9a3d19f5 as an eligible-not-active list with two-track promotion.
- Enable Jev at launch for routing and ranking only; never let it change issue state or the gate.
- Make Go, Python and TypeScript tier-1 languages.
- Plan to archive `SCOUT_REBUILD_REQUIREMENTS.md`@9a3d19f5 in favour of the governed specification documents listed in `docs/scout-design.md`@9a3d19f5 §10.3; ADR-009 replaces this cutover with `docs/scout-design.html` Appendix A, without separate product-contract or adapter-contract documents.
- Pin Clawpatch v0.7.3 at commit `734f81bc1514b52a0627edb0c24ebe1a18b36105` with the ported redaction patch; ADR-004 updates that patch to `scout-review-sanitize-v3`.

## Informed by

- `docs/archive/scout-design-v1.md`@9a3d19f5 §15 records the seven decisions as of 2026-10-19.
- `docs/scout-design.md`@9a3d19f5 Appendix items 1–7 provided the authoritative wording before ADR-009.
- `docs/archive/scout-design-v1.md`@9a3d19f5 §15 item 1 cites E3 and the Clawpatch 0.7.3 review for persistent agent-ID churn and retaining the areas layer.
- `docs/archive/scout-design-v1.md`@9a3d19f5 §15 item 7 cites `evaluation/clawpatch-0.7-compat/RESULTS.md` §7 for the Clawpatch upgrade and ported patch.
- `docs/archive/scout-design-v1.md`@9a3d19f5 §15 item 7 distinguishes the pinned commit from annotated tag object `c46c0c6`.

## Changes

- Item 1 replaced `docs/archive/scout-design-v1.md`@9a3d19f5 §4.1's nightly refresh default with monitored-area changes on default-branch merges.
- Item 3 replaced the memo's 20-PR shadow minimum with at least 5 comparable-base PR runs.
- Item 3 retained a minimum shadow duration of 14 days.
- Item 3 replaced `docs/archive/scout-design-v1.md`@9a3d19f5 §7.4's promotion procedure with the two-track procedure in `docs/scout-allowlist.md`@9a3d19f5.
- Item 6 selected `docs/archive/scout-design-v1.md`@9a3d19f5 §9.2's specification documents; `docs/scout-design.md`@9a3d19f5 §10.3 listed their successors before ADR-009 replaced that plan.
- ADR-002 confirms that the specification cutover happens at Phase 6 and keeps `SCOUT_REBUILD_REQUIREMENTS.md`@9a3d19f5 live until then (`docs/scout-design.md`@9a3d19f5 Appendix 12).
- ADR-003 adds Jev's model pin, provisional launch disclosure and tuning defaults without changing route/rank-only authority (`docs/scout-design.md`@9a3d19f5 Appendix 19).
- ADR-004 replaces the earlier Clawpatch review-safety patch with `scout-review-sanitize-v3` (revision 2.4; `docs/scout-design.md`@9a3d19f5 Appendix 7).
- ADR-005 narrows agent mapping to `codex`, `gpt-6-luna`, effort `xhigh`, with no provider fallback (`docs/scout-design.md`@9a3d19f5 Appendix 20).
- ADR-006 supersedes item 1's trigger clause with onboarding and manual `scout map --refresh` only (`docs/scout-design.md`@9a3d19f5 Appendix 21).
- ADR-008 reopens `docs/scout-design.md`@9a3d19f5 Appendix 21 only for pending areas from the opt-in merge-time association lane; the agent-map schedule stays unchanged (Appendix 23).
- ADR-001 amends the inertia framing without reopening items 1–7.
- ADR-007 amends the v1 migration commitment to retain only the version-3 re-anchor.
- ADR-009 supersedes item 6 and the Phase-6 specification cutover with requirements in `docs/scout-design.html` Appendix A, without separate product-contract or adapter-contract documents.

## Conditions

- Defer S3 until the branch backend is proven, then allow it as an optional adapter.
- Require both at least 14 days and at least 5 comparable-base PR runs for the shadow window; make both configurable.
- Require a maintainer for blocking overrides.
- Allow rare security checks to activate with `provisional: true` on track A alone as a conservative default to confirm through use.
- Treat Clawpatch 0.8.x as a later, separate review.

## Review

- `docs/archive/scout-design-v1.md`@9a3d19f5 §15 records the decisions, updated 2026-10-19; its original header dates the council memo to 2026-10-17.
- Final sign-off: the v1 council decision memo, `docs/archive/scout-design-v1.md`@9a3d19f5 §15; no round files exist for this record.
- Signed digests: none recorded for these seven decisions in the cited memo.
- Adoption record: `docs/scout-design.md`@9a3d19f5 Appendix items 1–7 carried the decisions forward as authoritative before ADR-009.
- The archived memo is preserved as `docs/archive/scout-design-v1.md`@9a3d19f5.

## Records in this folder

- `ADR.md`: summary of the seven v1 baseline decisions and their later amendments.
