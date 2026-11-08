# ADR-010: response to council round 2

Date: 2026-11-08. This records fixes for the next review, not a sign-off or adoption. Round 2's six minor issues remain preserved in `round-2.md`. Route (a) continues: the signed set is in adopted form, while this response and the final review/adoption records stay outside it. ADR-010 has not yet been adopted; accepted-form review bytes do not establish sign-off.

| Council issue | Fix |
| --- | --- |
| ADR010-R2-01 | STATUS's title/current-state dates now say 2026-11-08. Its main design pointer names ADR-000…ADR-010, the ADR-010 review folder as latest amendment and ADR-009's review/digest as the restructure provenance. |
| ADR010-R2-02 | Removed the accepted card's Before adoption section. ADR Conditions no longer contains the pending-adoption sentence; Review uses the requested adopted-form wording pointing to the explicit zero-issue final round, with adoption recorded outside the signed set. This response records the actual pending-review state without requiring later changes to signed bytes. |
| ADR010-R2-03 | ADR Changes explicitly lists removal of the old triage recall `[ASSUMPTION]` parenthesis, replacing the inaccurate “inserts only” description. It also records the deletion of the independent-agent blindness sentence in the previous rewrite and its restoration: defaults again says independent sessions never see Jev's answer, with REQ-100 cited. |
| ADR010-R2-04 | README, ADR Informed by and R-063 now say the metrics pool per-fold thresholds: 0.385 in 13, 0.26 in 9 and 0.24333349999999998 in 3 of 25 fold tests, not one fixed threshold. All three add weak axum precision 0.5 (TP 1, FP 1, n=8, one positive), three-call means versus one production request (potential transfer failure), and out-of-fold precision 0.7950310559006211 below 0.80: joint attainability is not established and D1-R must re-check it. CSV and JSON are unchanged; no new analysis/model calls. |
| ADR010-R2-05 | The main agent supplied the exact Q1/Q2 ask record on 2026-11-08. R-063 now identifies that source and quotes both questions, every option label/description, and each selection verbatim. It explicitly distinguishes assistant-drafted option text from the owner's act of selecting it and records the initial “you should use jev-actions…” direction verbatim, including its original “datat” spelling. The supplied ask record has no additional free text. ADR Decision now says the owner selected the assistant-drafted miss-cost option for triage, rather than attributing its wording to the owner. The Q2 rationale is explicitly attributed to the assistant; minimums were data-set and the costlier side separately confirmed. |
| ADR010-R2-06 | Reversed the mermaid dependency: signed X1 or dedicated triage preregistration → D1-R. ADR Changes states the same prerequisite direction. D1-R does not block drafting the preregistration. |

## Submission invariants

- The frozen `internal/decisions/thresholds/thresholds_test.go` remains unchanged. Council round 2 accepts ADR text as the correction to its comment; no proof mutation is needed.
- Recompared the ADR-003 card with base `04d5e632`: byte-identical except the permitted `rel` span.
- Confirmed every changed file against base, excluding review/response records, is in the same fourteen-file signed set listed in `round-1-response.md`. No new signed-set path is required.
- Main must recompute all fourteen signed-set SHA-256 digests and attach `git diff --stat 04d5e632..HEAD` for the next council round. Review/adoption records remain outside the signed set.
