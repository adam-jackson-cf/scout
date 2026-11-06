# ADR-009: One design document with requirements and ADR appendices

- **Status:** accepted
- **Design revision:** `docs/scout-design.html` (adopted 2026-11-06; replaces revision 2.8 of `docs/scout-design.md`)
- **Decided:** 2026-11-06; **sources:** R-062, user direction

## Decision

- `docs/scout-design.html` is the governed design. It has a visual Overview, Appendix A (requirements `REQ-1`…`REQ-183`, normative) and Appendix B (ADR cards). Where they differ, Appendix A prevails.
- It replaces `docs/scout-design.md`, `SCOUT_REBUILD_REQUIREMENTS.md` and `docs/scout-allowlist.md`. `REQ-1`…`REQ-41` keep their IDs.
- It ends the planned Phase 6 spec cutover. There is no `docs/scout-product-contract.md` or `docs/scout-adapter-contracts.md`. The requirements file is retired now, not at Phase 6.
- Council records move from `docs/design-reviews/` to `docs/adrs/<YYYY-MM-DD>-<topic>/`, each with an `ADR.md` summary. IDs: ADR-000 (v1 user decisions, no council round) to ADR-009.
- Legacy documents are deleted: `docs/archive/`, `docs/scout-design-brief.md`, `docs/inertia-research.md` and `docs/enaible-findings.md`. The last one is not moved to `docs/evidence/`. Git keeps them all at `9a3d19f5`.
- REQ-37 follows the design's principles-only subject interface (design §8.7 at `9a3d19f5`). The old requirement's exact `ib-subject-*/1` schema, wall-clock cap and sort order are not adopted.
- REQ-42 adopts the experiment list X1–X6, X9–X11 (design §13 at `9a3d19f5`).
- An accepted ADR is never edited, except its status line. The commit that adopts a later ADR updates that line.

## Informed by

- User direction, 2026-11-06 (R-062): one simple visual design document; requirements and ADRs as appendices that grow over time; no legacy or already-implemented documents.
- The 163 KB revision-2.8 design had grown by revision paragraphs, and readers could not find what is decided or required.

## Changes

- Supersedes part of ADR-000: item 6, the spec cutover to separate contract documents.
- Supersedes part of ADR-002: Appendix 12, which kept the requirements file live until Phase 6, and its experiment list. Base design §1 and Appendix 11 said "X1–X6, X9, X10"; REQ-42 adopts X1–X6, X9–X11, which includes ADR-008's X11.
- `requirements-map.md` maps every design heading and subsection, Appendix item, allow-list section and old REQ to a REQ, an ADR or a listed drop.
- `design-section-map.json` maps old section citations (`design §n`, `Appendix N`, `AL §n`, `req §n`) to REQ and ADR IDs. Code comments and docs cite those IDs now.
- `AGENTS.md` design rules 1 and 2 name `docs/scout-design.html`, the documents it governs and the ADR folders. The experiment list there is X1–X6, X9–X11.
- Research ledger R-062 and a hypothesis-ledger History line record the moved paths. Earlier rows are unchanged.

## Conditions

- Frozen evaluation records, signed review rounds and earlier ledger rows keep their old paths. `docs/adrs/README.md` maps them.

## Review

- This summary is among the signed files, so it does not hold its own outcome or digests (`AGENTS.md` design rule 2).
- Review outcome and signed digests: the zero-issue final round in this folder (`round-<N>.md` with the highest N). Earlier rounds record the issues found and fixed.

## Records in this folder

- `ADR.md`: this summary.
- `proposal.md`: the change set and the questions put to the council.
- `requirements-map.md`: disposition of every source section and requirement.
- `design-section-map.json`: old section citation → REQ/ADR IDs.
- `completeness-audit.md`: phrase-level check of Appendix A against the base documents, with a decision per candidate.
- `round-<N>.md`: council review rounds.
