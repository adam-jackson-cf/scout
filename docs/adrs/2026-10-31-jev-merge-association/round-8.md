# Design review round 8: DRAFT revision 2.8, Jev merge-time area association with a reasoning-model fallback

Date: 2026-10-31. Rules: CLAUDE.md design rules 1–3. Reviewers: the innovation council (lead), with independent peers `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). The review was read-only: no shell, no model calls, and `.codegraph` was not touched.

Reviewed files (SHA-256): `docs/scout-design.md` `7b102ebbabddebb9f49bdec5cd559bf6908ed32b1bd136f01ee1f5a4dc4b0551`; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` `36d082ea61b6024091a14dad5120e0ef5d97d486ec83a8523eecb6d6fce52671`

## Scope

This round checks the R7-01 replacement at `docs/scout-design.md:116` and `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md:56`, and anything it broke in the rest of :116 / :56 and :114.

Settled decisions are not reopened (R-059, R-061): a pending area is created at merge time; a maintainer accepts it; the lane is a provisional opt-in, off by default; the user's exception to design rule 3 stands; the fallback is `codex exec gpt-6.1-terra` at high effort only.

## Round-7 fix: verification

The replaced sentence, identical in both files:

> It is `unmeasured` while those merges hold fewer than `min_clusters` residual clusters (for example, on hosted runners without `codex`) or no cluster the fallback answered (for example, when every fallback call fails).

| Check | Where | Finding | Result |
|---|---|---|---|
| Text | :116, proposal :56 | identical to the round-7 replacement | applied |
| What "It" refers to | :116 | the pending-new rate only; the fallback rate stays measured "whether or not the fallback ran" | consistent |
| Every fallback call fails | :116 | the filtered merges then hold no answered cluster, so the rate reads `unmeasured`, not 0.00 | closes R7-01 |
| Hosted runner without `codex` | :114, :116 | `skipped (no model CLI)` leaves out every merge, so the rate is `unmeasured`; clusters not sent still count in the fallback rate | correct |
| Healthy window with no escalations | :116 | the rate is `unmeasured` and the fallback rate is 0, so the guard does not fire either way | correct |
| `min_clusters` floors | :116 | the pending-new floor (filtered merges) and the window floor (fallback rate) cover different sets and do not conflict | consistent |
| Count before the pending cap; `jev_failure`/`jev_cap`/`pr_cap` in neither rate; `fallback_cap`/`terra_failure`/skipped in the fallback rate only | :116 | the new clause adds no new place where a cluster is counted | consistent |
| Rebuilt only from exported rows | :114 | answered clusters are read from run id, escalation reason, raw answer and outcome, plus the run-level skip flag; no new field or stored state is needed; inherited rows stay out of both rates | implementable |
| fastapi rounds 8–9 | `evaluation/x2d-postrun-tuning/README.md` rounds 8–10 | 21 + 23 = 44 clusters; 7 + 13 = 20 escalated; 0 failed or unparsable calls, so 20 answered; 3 + 11 = 14 `new`. Fallback rate 20/44 = 0.45, pending-new rate 14/44 = 0.32, and 44 ≥ 20 | still measured |
| axum rounds 1–5 | `evaluation/x2d-postrun-tuning/axum/README.md` | 32 of 109 escalated; 0 failed calls, so 32 answered (the "5 Terra errors" are wrong answers, not failed calls); 27 `new`, so 27/109 ≈ 0.25 | still measured |
| Other places that state the guard | :19, :427, :444, :463 | none restates the filter; :463's defaults (20, 20, 0.40, 0.20) match :116 | no conflict |
| Settled decisions and the Appendix | :497–500 | manual refresh only; settled decisions untouched | no break |

## Issues

None.

## Considered and not raised

- **A window with only one answered cluster** is measured on a thin basis. The guard's recall and false-fire rate are already disclosed as unknown and are tested by X11 (:116, :444).
- **Bias when some calls fail or the cap is reached** was considered in round 7 and is disclosed. The rate still rests on real answers.
- **Implementation note (challenger).** Measurability can be implemented as two explicit counts over the filtered merges: residual clusters ≥ `min_clusters`, and valid fallback answers > 0. This matches the sentence and needs no lane-health status.

## Peers

Both peers and the lead independently report zero issues for this round.

## Review history

Rounds 1–8: 14 → 10 → 6 → 4 → 3 → 1 → 1 → 0 issues. Records are in `round-1.md` through `round-8.md`.

The ledger entries this design cites are R-055 to R-061 (R-060 corrects the reading of R-058; R-061 records the user's confirmation after seeing axum) and HD8, which is `proposed` with X11 as its test. User question U1 was answered (R-061). No question for the user is open.

## Sign-off

The innovation council signs off DRAFT revision 2.8, merge-time area association with a reasoning-model fallback, with zero issues, under CLAUDE.md design rules 1 and 2. The sign-off covers the two reviewed files:

- `docs/scout-design.md`: its revision 2.8 text, namely the status line, the revision paragraph at :19, §3.3, §4 (merge-time association, evidence and limits, refresh guard), the §6 rows and adapter guardrails, §7.2, the §11 Phase 1 cutover row and acceptance, R17, R18, X3, X11, A13, A14, and Appendix 21, 22(a) and 23;
- `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md`.

This sign-off holds only while the SHA-256 digests recorded in `Reviewed files (SHA-256): `docs/scout-design.md` `7b102ebbabddebb9f49bdec5cd559bf6908ed32b1bd136f01ee1f5a4dc4b0551`; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` `36d082ea61b6024091a14dad5120e0ef5d97d486ec83a8523eecb6d6fce52671`` match both files byte for byte, as reviewed in this round. The council had no shell and computed no digests; the main agent records them. Any later change to either file voids this sign-off and needs a new council round with zero issues.

The council did not review the adoption steps: moving the status line from revision 2.7 to 2.8, and updating `docs/STATUS.md` and the requirements file. As revision 2.8 itself states:
- X11 needs a council-signed preregistration before any run;
- the `orderflip` flip rate on `merge.which_area` must be recorded before a release lets the opt-in lane be enabled (Appendix 19, B-3).

## Verdict

**ZERO ISSUES — signed off**
