# Design review round 6 — DRAFT revision 2.8: Jev merge-time area association with a reasoning-model fallback

Date: 2026-10-31. Rules: CLAUDE.md design rules 1–3. Reviewers: innovation council (lead), with independent peers `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). Read-only review: no shell, no model calls, `.codegraph` untouched.

Reviewed files (SHA-256): `docs/scout-design.md` `cec57f8eb18516530a3ab835a8afa043ffad1425e2040ac872e6c1d841a232ad`; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` `e205bf13635c61718db53a49949d2faa2f37a5aef4cab17ac168389e4a05536b`

## Scope

This round checks the round-5 replacements and anything they broke:
- `docs/scout-design.md` :114 (R5-01, and both R5-02 clauses) and :116 (R5-03);
- `proposal.md` :11 (R5-01), :14 (R5-02) and :56 (R5-03).

Settled decisions are not reopened (R-059, R-061):
- a pending area is created at merge time;
- a maintainer must accept it;
- the lane is a provisional opt-in, off by default;
- the user accepts the development evidence as an exception to design rule 3;
- the fallback is `codex exec gpt-6.1-terra` at high effort only.

## Round-5 fixes: verification

| R5 | Applied | Check | Result |
|---|---|---|---|
| 01 pending rename | yes, text-identical (:114, proposal :11) | `axum/pilot/frontend.py:347-369`: empty-file trap, then `owners(old)`, then `is_pooled(old)` (pooled = in the pool and owned by no area, `:283-285`). The pooled-rename branch returns before `add_residual`, so a heuristic-only file is not decided. `gold.py:153-162` `move` takes the old path's owners and removes the old path from every area and from the pool. The new path goes to the pool only if it has no owner and the old path was pooled | correct |
| 02 supersede tests | yes, both clauses (:114, proposal :14) | Paths that were deleted or renamed away no longer block the test. A pending area or record with no surviving file is marked `superseded` and listed, never deleted or accepted (the adjacent maintainer-only rule holds). The measured runs dropped removed paths (`build_units.py:255-260`, `gold.py:150-151`). `Reconcile` replaces a matched area's `FileSet` (`internal/areas/areas.go:150`), so the test reads the new map | correct |
| 03 pending-new rate | yes (:116, proposal :56) | fastapi rounds 8–9: 7 + 13 = 20 escalated, 3 + 11 = 14 `new`, 0 failed calls (`README.md` rounds 8–10), giving 20/44 = 0.45 and 14/44 = 0.32. axum: 32 escalated, `terra_failures: 0`, 27 `new`, giving 27/109. Both exceed `min_clusters` 20 | figures correct; the merge filter is wrong (R6-01) |

## Issue

### R6-01 (low): the merge filter keeps merges where the fallback never ran, so the rate reads 0.00 instead of `unmeasured`

**Where:** `docs/scout-design.md:116` and `proposal.md:56`.

**Text:**

> both counted over the window's merges in which the fallback answered every cluster Jev left `unresolved` or `none`; a merge with a cluster left `unresolved (fallback_cap)` or `unresolved (terra_failure)`, or not sent because `association-fallback` was skipped, is left out

and

> It is `unmeasured` while those merges hold fewer than `min_clusters` residual clusters (for example, on hosted runners without `codex`).

**Problem:**
- A merge where Jev escalated nothing passes the filter even when `association-fallback` was skipped, because it has no unsent cluster that would exclude it.
- On runners without `codex`, only such merges are kept. They add residual clusters and can never add a `new` answer. Once they reach `min_clusters`, the rate reports a measured 0.00, which contradicts the hosted-runner example in the same sentence.
- The same filter biases the rate down whenever the fallback runs only some of the time: merges with many escalations are dropped, while merges where Jev assigned everything are kept.
- The guard is computed only from exported association rows ("no other guard state is stored"). A cluster Jev assigned has empty fallback fields, so no row shows whether the fallback was available in that run.
- [INFERENCE] With about 2.2 clusters per merge and about 27 of 100 clusters escalated, roughly half the clusters of a 20-merge window fall in merges with no escalation, which is around the 20-cluster floor.

**Fix**, in both files:

1. Replace the parenthesis with:

   > (valid `new` answers ÷ residual clusters, both counted over the window's merges decided in runs in which `association-fallback` was not skipped and in which the fallback answered every cluster Jev left `unresolved` or `none`; a merge decided in a run in which `association-fallback` was skipped, even one with no such cluster, or with a cluster left `unresolved (fallback_cap)` or `unresolved (terra_failure)`, is left out)

2. In the `associations` row field list (design :114 and proposal :16), change "outcome and `stale`" to:

   > outcome, whether `association-fallback` was skipped in that run and why, and `stale`

The cited figures do not change: in both runs the fallback ran and answered every escalation.

## Considered and not raised

- **Which clusters the floor counts.** "Fewer than `min_clusters` residual clusters" uses the rate's own denominator, which "enter neither rate" already narrows by removing `jev_failure`, `jev_cap` and `pr_cap` clusters. That is the consistent reading.
- **Pending areas with no surviving file.** They are superseded because the test is trivially true. This is a coverage marker, not a status change: acceptance and merge stay maintainer-only (Appendix 23(a–b)).
- **A file in more than one pending area.** Rule (f) can decide a heuristic-only pending file again. This follows from the earlier rule (f) decision, not from R5-01. "Replaces its old path" is implementable for every pending area that holds the path.
- **Stale paths after renames that do not inherit an owner.** The R5-02 filter on "files that exist at the mapped revision" makes them harmless for the supersede tests.
- **A live-file projection at map time** would be an equivalent implementation. The stated predicates keep historical records intact without another membership representation.

## Where the peers changed the lead's view

- The lead and the challenger both found the three round-5 fixes correct. The challenger reported zero issues.
- The experiment peer found R6-01, an empty-merge case in the R5-03 filter that the council's round-5 text did not cover. The council adopts the peer's replacement and its row field.

## Open questions that need the user

None.

## Verdict

**1 ISSUE** (low: R6-01). No sign-off in this round.

Any later sign-off must name the SHA-256 digests of the revised `docs/scout-design.md` and `proposal.md`, and holds only while those digests are unchanged. Round 7 checks only the two sentences this fix changes.
