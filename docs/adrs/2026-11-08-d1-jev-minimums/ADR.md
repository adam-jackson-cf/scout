# ADR-010: Record owner-agreed Jev triage minimums

- **Status:** proposed; council review pending, not adopted
- **Design revision:** proposed amendment to `docs/scout-design.html` REQ-98 and Appendix B
- **Decided:** owner direction 2026-11-08; **sources:** R-063

## Decision

- Record `triage.real_defect` minimum precision 0.80 and minimum recall 0.90 as owner-agreed, with misses the costlier side. The numerical values are unchanged; `minimums_provenance` changes from `assumed` to `agreed`.
- Keep threshold 0.3, status `provisional`, and `gating: false`. Agreement on minimums does not establish a tuned or ready threshold, validate Jev ranking, or change REQ-102.
- Use the agreed minimums as the starting point for X3/X5/X10 precision/recall point minimums. Each preregistration still confirms or replaces its minimums and names the costlier side (X10 recall n/a). These defaults do not replace X1 gain or X4/X5 lower bounds.
- Retain the owner-requested **D1-R: review and optimise the Jev triage minimums and threshold** in `docs/plans/2026-11-06-remaining-work.md`. Entry requires at least 30 human/`agent_reviewed` labels per label (jev-actions B-2), from the frozen-guide blind independent two-labeller protocol with adjudication, and a council-signed preregistration (part of X3/X10 or its own). Work is preregistered tuning on a repository-split dev/test set, checking 0.80/0.90 feasibility per stratum, setting or replacing the threshold, and reviewing the minimums with the owner.

## Informed by

- R-063 records owner direction: set minimums from data rather than costs, accept precision 0.80 / recall 0.90, and request later review and optimisation.
- The owner waived preregistration for the exploratory analysis described in R-063 only. It used 51 human-confirmed labels (28 real_defect / 23 false_positive, one cannot_tell excluded), saved Jev 1.14 scores averaged over three repetitions, and F1 tuning with 5-fold cross-validation × 5 repeats, seed 7. R-063 records the reported metrics and uncertainty intervals.
- The analysis had fewer than 30 cases per label, no test split, v1_choice label-guide versus v2_noul score mismatch, weak nest/webpack strata, and unstable fold thresholds 0.24–0.39 (spread >0.05). It is **not citable as experiment evidence for design changes** under design rule 3. This ADR records the owner's decision, not an empirical validation claim or an exception to that rule.

## Changes

- Proposes to amend ADR-003's conditions treating triage recall 0.9 as an assumption until the user sets minimums. The accepted ADR body remains a historical record; its status line is updated to name ADR-010 only upon adoption.
- Updates REQ-98 and the Appendix B open-conditions summary; adds the ADR-010 card and index entry. Other per-decision defaults and all experiment sign-off requirements remain unchanged.
- Aligns `docs/STATUS.md` with D1, R-063 and the already-recorded D1-R plan. The embedded-default provenance and frozen proof were committed in `b8d81025`.

## Conditions

- This design amendment is not adopted until an innovation-council review reports zero issues and signs the reviewed file digests.
- No new experiment, Jev comparison or threshold tuning is authorised by this ADR. Future runs follow the existing experiment rules and REQ-102.

## Review

- Pending; no council round or sign-off exists yet for this draft.
- This summary is intended to be among the signed files. The review outcome and SHA-256 digests will be recorded in this folder's explicit zero-issue final `round-<N>.md`, rather than embedded here.
