# Round 4 response — revision 2.2

Proposed file: 462 lines, SHA-256 `223d89fdffc4cd11e82e91f4a1bae4bfc4bafadbbec351380d1796461b633852`. Base unchanged; `docs/scout-design.md` and Appendix 1–7 untouched. Only cited lines edited; no new numbers.

| # | Fix | Where |
|---|---|---|
| 1 | One sentence added: interim draw is an independent per-unit draw at the stored probability (stored value is the realised inclusion probability), under a stored audit-sampler version; unit, strata, fractions, floor are `[ASSUMPTION]` until X1's preregistration is signed off, which may discard interim draws. Mirrored in REQ-29. | L160; REQ-29 |
| 2 | L166: UI records, with every disposition, the view it was made in (audit view or not) and annotation visibility. L160: "audit labels" → "audit dispositions or labels on audited units". Mirrored in REQ-29. | L160, L166; REQ-29 |
| 3 | "Once X1's preregistration is signed off" (L162); REQ-29 "until X1's preregistration is signed off (sampling unit, strata, draws and floor fixed there)". | L162; REQ-29 |
| 4 | X3 labels read "unowned files (rev 2.1: 200)". | L409 |
| 5 | Phase 5 acceptance adds: audit sample drawn with a stored inclusion probability per unit, presented blind; on a synthetic frame selection frequency matches the stored probability. | L374 |
| 6 | Cap order: "ties broken by issue id; `changed-in-PR` = 1 outside PR runs; both part of the baseline formula id". | L157 |
| 7 | R-038 Informed → "stratified-audit principle; X1 estimator carried to X1's preregistration (C7b)"; STATUS row 8 cite → `AGENTS.md:46-49`. | `research-ledger.md:46`; `STATUS.md:90` |

Files changed outside the proposed file: `proposed.diff`, `DRAFT-NOTES.md`, `round-4-response.md`, `SCOUT_REBUILD_REQUIREMENTS.md` (REQ-29), `docs/research-ledger.md`, `docs/STATUS.md` (row 8).
