# Round 2 response — revision 2.2

Proposed file after this response: 462 lines (numbering unchanged), SHA-256 `fe5005e7eef356a5696eb577c4002c30242f946aef2d055bbf9d68b41bf93ceb`. Base unchanged (`7ff21a7f…`); `docs/scout-design.md` untouched; Appendix 1–7 untouched. Line numbers are in `scout-design.proposed.md`. No new number is introduced: the interim sampler values are rev 2.1's 5 % total, split equally, all tagged `[ASSUMPTION]` and replaced by X1's preregistration.

| # | Resolution | Lines |
|---|---|---|
| 1 | "Audit label" defined once in §6: a label on an audit-sampled unit with a stored inclusion probability, whichever labeller. §9.7 uses the maintainer's audit dispositions (`human` stream); X1 uses only §6.8-protocol labels on audit-sampled units drawn under the frozen question and formula versions (earlier dispositions excluded or re-labelled). Audit items are presented outside the ranked queue with annotation, stratum and rank hidden. | 160, 166, 407 |
| 2 | Executable interim sampler, each part `[ASSUMPTION]`: strata = top and bottom halves of annotated units by rank position within the run; the 5 % total split equally; sampling unit `(issue id, rank-formula version)` at its first eligible nightly run; unannotated units form their own stratum at the same fraction (never excluded); certainty stratum (p = 1, stored) for critical findings below the audit floor once fixed. Phase 5's injection fixture lands in the certainty stratum. | 160, 374, 407 |
| 3 | X1's preregistration names the §9.7 line and the Phase 5 replay as covered analyses with the repositories and period they may read; §9.7 withheld for X1's test repositories until X1 is recorded for that formula version. Phase 5 now records the X1 result obtained under the signed-off preregistration instead of re-evaluating the rule. | 166, 308, 373–374 |
| 4 | Rank-formula version (formula id, weights, constituent question versions) stored with every `ranking` disclosure; "shipped version" in X1 = formula version; X1 verdict and §9.7 keyed to it and stale on change; dev tunes weights only, questions frozen before labels. | 155, 308, 407 |
| 5 | X4: enrichment serves the positive count and non-hit slice, not the hit sample; precision is on an equal-probability sample of the hit population, or inverse-inclusion weighted with the §9 design-based interval if unequal. §9 example reworded to "equal-probability sample of the hit population, as E1b's G5". X10: precision population-weighted to the candidate-pair population with the §9 interval; per-stratum precision reported alongside. | 300, 410, 414 |
| 6 | Council allowed either route; I synchronised the rows (the cleaner route) and kept the existing §10.3 sentence. `SCOUT_REBUILD_REQUIREMENTS.md`: REQ-31 (rev 2.2, Clawpatch provider, exit-code rule), REQ-35 (§15.3), REQ-36 (§15.1 contents open, Appendix 11), REQ-38 (Appendix 14), REQ-41 (Wilson or design-based), §10 (items 1–3 renumbered). Phase 5 "prints Wilson intervals" → "the §9 intervals". | 346, 374; requirements file lines 54, 61–62, 70, 73, 77 |
| 7 | X1 reports Jev-minus-baseline recall@k on the injected set, descriptive; an injected defect below the audit floor counts as a critical true issue below it when critical. | 407 |
| 8 | Only the call cap applies until the Phase 5 measurement sets a cost bound, which covers every question (triage and rubric). X1 runs are uncapped or exclude `partial` runs, as fixed in its preregistration. | 157, 407 |
| 9 | Status line now: "Phase 1 is implemented in the working tree (`docs/STATUS.md` §1); every later acceptance scenario is future work." | 3 |

## Files changed outside the proposed file
`proposed.diff` (98 insertions, 53 deletions), `DRAFT-NOTES.md`, `round-2-response.md`, `SCOUT_REBUILD_REQUIREMENTS.md`.

Not changed (council pointers for adoption): `docs/STATUS.md` handoff summaries are to be refreshed at adoption, outside this round.
