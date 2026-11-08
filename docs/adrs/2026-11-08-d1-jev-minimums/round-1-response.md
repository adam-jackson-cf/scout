# ADR-010: response to council round 1

Date: 2026-11-08. This is the response for the next council review, not a sign-off or adoption record. Round 1 found 13 issues and remains preserved unchanged. **ADR-010 has not been adopted:** signed-set bytes below are prepared in adopted form for route (a), following ADR-009 round-13 step 5. Actual adoption must be recorded outside that signed set after an explicit zero-issue council sign-off; no later proposed→accepted edits to signed bytes are needed.

## Owner answers

R-063 now quotes the owner's scope clarification verbatim: “(a) Triage only” — “P 0.80 / R 0.90 applies to triage.real_defect only. X3, X5 and X10 keep REQ-98's per-decision defaults until you set them or a preregistration replaces them.” It also records “Confirm: misses cost more” separately from the earlier “use the data not costs” direction. The owner set the pair from data, not explicit costs; separate confirmation establishes the costlier side, not a retrospective cost-based derivation.

## Per-issue fixes

| Council issue | Fix |
| --- | --- |
| ADR010-R1-01 | Triage-only scope is consistent in REQ-98, ADR-010, its card, STATUS, R-063 and the plan. D1-T is done; non-triage D1 remains open with D1→Track C. X3/X5/X10 retain their own defaults, and REQ-169–REQ-175 are unchanged. R-063 quotes the clarification and corrects the earlier K3 misdescription. |
| ADR010-R1-02 | Restored REQ-98's general `provisional`, pending costs/minimums and carried rev-2.1 wording. Its added clause is triage-only and states data, not explicit costs, with the owner's separate miss-cost confirmation. ADR Changes enumerates these edits and quotes the historical “until the user sets costs and minimums” correctly. **Frozen-proof exception:** `thresholds_test.go` is unchanged, including its comment. ADR-010 explicitly corrects that comment's rationale: data set the pair; the costlier side was separately confirmed. No assertion was changed or weakened. |
| ADR010-R1-03 | ADR Informed by/Conditions, card, R-063 and STATUS explicitly name REQ-102 and CLAUDE.md General “Jev labelling”, describe owner-directed analysis and acceptance after unpreregistered/not-citable disclosure, and limit this deviation to R-063: not precedent and no design-rule-3 exception. The exploratory README and plan carry the same limitation. |
| ADR010-R1-04 | ADR Conditions, card, D1-R, R-063, STATUS, defaults assumptions and exploratory README quarantine all 52 label cases plus E2 v2_noul score files. X1/D1-R must disclose them, use them only for development and draw confirmatory tests from unexamined cases. Legacy v1_choice-guide labels do not count toward the current-question 30-per-label floor. |
| ADR010-R1-05 | D1-R is part of X1 or its own council-signed preregistration everywhere; its threshold does not change X1's ordinal decision rule. |
| ADR010-R1-06 | Committed exact `scores.csv`, regenerated `--json` output and reproduction README under `evaluation/d1-triage-minimums-exploratory/`. R-063 distinguishes pooled out-of-fold metrics from the in-sample 0.385 optimum, gives exact counts, descriptive Wilson calculation and effective denominators, and records repeated-CV dependence. Added production-question/state and F1-versus-F2 mismatch to ADR, R-063 and defaults. README gives script SHA-256, exact command and byte-matching CSV construction recipe; no new model calls. Owner statements are quoted. |
| ADR010-R1-07 | Restored ADR-003 card body byte-for-byte against base; only its relation span now names ADR-010. ADR-003 `ADR.md` changes only its status line; README names the amendment. An explicit comparison confirmed the card equality apart from `rel` and ADR body equality apart from status. |
| ADR010-R1-08 | Chose route (a). ADR-010, HTML requirement/card/footer, index and STATUS are in adopted form. The prior ADR-003 status/rel/index amendment is likewise included in the review bytes. This response, and the later council/adoption records, are outside the signed set and establish that accepted-form text is not yet adoption. Signed-set files will not be rewritten to mark adoption. |
| ADR010-R1-09 | Updated governed audit Fixed row 22 to owner-agreed triage minimums, separately confirmed miss cost, and explicit missing current-question labels / preregistered tuning conditions instead of a nonexistent visible TODO. Audit is listed in ADR Changes and the signed set. |
| ADR010-R1-10 | Defaults assumptions now state E2 human labels exist but no label set exists for this question version; minimums are owner-set, threshold untuned, status provisional pending preregistered D1-R readiness. Added non-citable exploratory status and F1/question mismatch. `label_set.summary` remains `none`. |
| ADR010-R1-11 | Card's Before adoption section contains only council sign-off. D1-R's floor, preregistration, quarantine and later work are under a separately labelled D1-R details section; ADR Conditions makes the distinction explicit. |
| ADR010-R1-12 | STATUS's work and latest-ledger pointers now reflect triage complete, non-triage D1 open, D1-R outstanding and R-063. Plan indexes ADR-000…ADR-010; D1 has its blocking edge and D1-R has an explicit X1/dedicated-preregistration relation. |
| ADR010-R1-13 | ADR Changes lists every affected implementation/document/artifact path, the proof assertions, and test-map change. Added Records in this folder. Eligible D1-R/default labels are human or agent_reviewed under the current-question protocol; removed `outcome` from defaults. ADR explains audit B-2's narrower agent_reviewed collection request as a subset of eligible sources, not a ban on human labels. |

## Reproduction and invariant checks exercised

- Regenerated tuning JSON with the requested jev-actions script and parameters plus `--json`, from saved CSV only; no model calls.
- Reconstructed the CSV from the named labels, score files and repo metadata: byte-for-byte match to the retained CSV; three score repetitions per case.
- Compared ADR-003 card against base `04d5e632`: byte-identical apart from the relation span; compared ADR-003 summary: byte-identical apart from the status line.
- Compared REQ-169–REQ-175 against base: each byte-identical.
- Descriptive Wilson bounds were recomputed from per-repeat averaged counts, not independent repeated observations; README records exact bounds and supersedes the earlier rounded recall bounds.

## Signed set for next council submission

The main agent must recompute SHA-256 on these bytes and report `git diff --stat 04d5e632..HEAD`. Review records, this response and the later adoption record are excluded, because they record the sign-off/adoption rather than alter the signed content.

1. `docs/scout-design.html`
2. `docs/adrs/2026-11-08-d1-jev-minimums/ADR.md`
3. `docs/adrs/README.md`
4. `docs/STATUS.md`
5. `docs/jev-adapter-audit.md`
6. `docs/adrs/2026-10-23-decisions/ADR.md` (status line only changed)
7. `docs/research-ledger.md`
8. `docs/plans/2026-11-06-remaining-work.md`
9. `internal/decisions/thresholds/defaults.json`
10. `internal/decisions/thresholds/thresholds_test.go` (unchanged since the supplied proof)
11. `.omp/test-map.json`
12. `evaluation/d1-triage-minimums-exploratory/README.md`
13. `evaluation/d1-triage-minimums-exploratory/scores.csv`
14. `evaluation/d1-triage-minimums-exploratory/tuning.json`
