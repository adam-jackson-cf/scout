# Round 3 response — revision 2.2 (C7 split)

Proposed file: 462 lines, SHA-256 `41f5abf8bac3d0cfbaef278b57b1d20e15aae8db984639c120e8e3ca7799306d`. Base unchanged (`7ff21a7f…`); `docs/scout-design.md` and Appendix 1–7 untouched. No new numbers.

**Split.** C7a stays in 2.2: adapter version key; rank-formula and baseline versions stored with every ranking and disposition; failure handling and lane health; 2,000-call cap; primitive choices (rubric Nouls, X10 Noul in both orders, X3 Choice + companion Noul); calibration wording; §6.8 labelling protocol; R14; four principles in §6 (probability-sampled audits with stored inclusion probabilities, L160; blind audit presentation, L160; no Jev-versus-label comparison, threshold, band or default change outside a signed-off preregistration, L166; provisional `[ASSUMPTION]` audit fraction and `insufficient` Jev-quality lines until X1's preregistration is signed off, L160, L308). C7b moves to the preregistrations. §13 rows X1/X3/X4/X5/X10 (L407–414) now state question, primitive, label protocol, the rev 2.1 rule flagged pending the user's minimums, and that estimator, split, population, interval and sample size are fixed by the council-signed preregistration.

| # | Outcome |
|---|---|
| 1 | Resolved by C7b deferral: dispositions/audit-label definitions removed (old L160, L166); §9.7 now reads "audit-view blind dispositions only, else `insufficient`" (L308); finer rules → X1 prereg. |
| 2 | Resolved by C7b deferral: draw operation, strata overlap and `fork` text removed (L160); governed by X1 prereg. |
| 3 | Resolved by C7b deferral: sampling vs shipped formula version lifecycle removed (L155, L166, L407); X1 prereg. |
| 4 | Resolved by C7b deferral: test-set reuse across versions removed from L407; X1 prereg. |
| 5 | Resolved by C7b deferral: capped/failed-unit estimand removed (L157, L407); X1 prereg. Cap sentence kept without X1 claims. |
| 6 | Direct fix in kept text: baseline version (formula id, rule-weight table digest, path-role classifier version) stored with every `ranking` and disposition (L155); keying of statistics fixed by X1 prereg. |
| 7 | Direct fix: L166 now matches `AGENTS.md:67` without qualifier; §9.7 withheld until X1's preregistration (which defines its inputs, repositories and period) is signed off (L308). |
| 8 | Resolved by C7b deferral: candidate-pair population left to X10 prereg (L414). |
| 9 | Resolved by C7b deferral: X4 reads "precision lower bound ≥ 0.6" with interval type left to prereg (L410). |
| 10 | Direct fix: Phase 5 fixture is a critical-severity finding that changes at most its rank; Jev replay only as a preregistration-named analysis; certainty-stratum claim removed (L373–374). |
| 11 | Direct fix: "until X1's preregistration is signed off" (L160, L308, L374). |
| 12 | Direct fix: REQ-29 rewritten to the kept principles (`SCOUT_REBUILD_REQUIREMENTS.md:52`); R10 now "blind probability-sampled audit with stored inclusion probabilities" (L395); §10.3 lists REQ-29, 31, 35, 36, 38, 39, 41 and §10 (L346). |

## C7b carry-forward (seeds the X1 and other preregistrations)
- **X1:** split unit repository; dev tunes weights only, questions frozen before labels; sampling vs shipped formula version; each later formula version scored on a not-previously-scored test set; baseline-version keying; de-correlation (no file/clone group across split); injected true defects (counted separately, ≥ 20 floor, recall@k descriptive, count below floor as critical when critical); audit floor as a rank position; all critical findings below floor labelled (certainty stratum); `k` = review budget; 95 % repository-clustered paired bootstrap of recall@k difference; power calculation (≥ 100 labels, ≥ 30 semantic positives floors; paired SE ≈ 0.08 `[INFERENCE]`); treatment of capped, `partial` and failed units (ITT or exclusion with counts; no unit lost by first-eligible exclusion); labels = §6.8 on audit units under frozen versions; which repositories/period §9.7 and the Phase 5 replay may read; withholding for test repositories until recorded.
- **Interim sampler:** top/bottom halves by rank, equal split of the 5 %, independent per-unit Bernoulli at the stored probability, sampling unit `(issue id, formula version)` at first eligible run, unannotated stratum, certainty stratum removed from the halves, drop `fork` from nightly causes; rank-bucket cuts by rank position.
- **§9.7:** audit-view disposition flag; excluded and counted when annotation visible; sampled-but-undispositioned reported `unmeasured`; off-sample dispositions reported separately, unweighted; per-rule label minimum.
- **X3:** option-order rule seeded by file id; second diagnostic order; top-K/full list ≤ 255; candidate coverage; recall = owned files assigned the correct feature; dev-tuned threshold, test scored once.
- **X4:** enrichment serves positive count and non-hit slice, not hit sample; equal-probability hit sample (Wilson) or inverse-inclusion weights (design-based interval); Semgrep pre-filter strata; recall minimum; `unmeasured` patterns; per-family reporting.
- **X5:** commit state shape fields/bounds; split by repository; G4 24/30 prior with derived Wilson [0.63, 0.91]; sample sizing for LBs 0.75/0.70; stratified over-sampling; error by path role and repository.
- **X10:** candidate-pair population = pairs the shipped lane judges (clone-group ∪ signature-similar), role of the random stratum; population-weighted precision with design-based interval, per-stratum alongside; dev-tuned threshold; both orders averaged.
- **Phase 5:** replay of the labelled X1 set is X1's single test scoring or dev-only.

## Files changed outside the proposed file
`proposed.diff`, `DRAFT-NOTES.md`, `PROPOSAL.md` (C7 split note), `round-3-response.md`, `SCOUT_REBUILD_REQUIREMENTS.md` (REQ-29).
