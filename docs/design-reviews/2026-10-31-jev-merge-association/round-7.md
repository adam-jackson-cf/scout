# Design review round 7 — DRAFT revision 2.8: Jev merge-time area association with a reasoning-model fallback

Date: 2026-10-31. Rules: CLAUDE.md design rules 1–3. Reviewers: the innovation council (lead), with independent peers `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). Read-only: no shell, no model calls, `.codegraph` untouched.

Reviewed files (SHA-256): `docs/scout-design.md` `2f082520f9f9a14cbbbc22d6d9a6d3510a696a2e0e8572098c8f1a49db4a0d6e`; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` `cbc8658ff114231ca2cd53e810396c03fe69a597e40bbe36a17c70113232de68`

## Scope

The two R6-01 replacements, and anything they broke:
- the pending-new-rate parenthesis: `docs/scout-design.md:116`, `proposal.md:56`;
- the `associations` row field list: design :114, proposal :16.

Settled decisions are not reopened (R-059, R-061): pending area at merge time; maintainer acceptance; provisional opt-in, off by default; design rule 3 exception; fallback `codex exec gpt-6.1-terra` at high effort only.

## Round-6 fix: verification

| Item | Where | Check | Result |
|---|---|---|---|
| R6-01 (1) parenthesis | :116, proposal :56 | identical to the round-6 text | applied |
| R6-01 (2) row field | :114, proposal :16 | "…latency, outcome, whether `association-fallback` was skipped in that run and why, and `stale`" | applied |
| Hosted runner without `codex` | :116 | The lane reports `skipped (no model CLI)` whether or not there is work, so every merge is left out and the rate is `unmeasured` | now correct |
| Rebuilt from exported rows only | :114 | Every row, Jev-assigned ones included, records the run-level skip | implementable |
| Adjacent sentences | :116 | `min_clusters`, counting before the pending cap, and "count in the fallback rate only" stay consistent | consistent |
| fastapi rounds 8–9 | `README.md:128-141` | 21 + 23 clusters, 7 + 13 escalated, 3 + 11 `new`, 0 failed calls, no skip, 20 < 40 so no `fallback_cap`: 20/44 = 0.45, 14/44 = 0.32 | holds |
| axum | `axum/README.md:12,20` | 32 escalated, 0 failed calls, 27 `new`: 27/109 ≈ 0.25 | holds |
| Other statements of the filter | :427, :444, :463 | none restate it | no conflict |

## Issue

### R7-01 (low): when every fallback call fails, the pending-new rate still reads a measured 0.00

**Where:** `docs/scout-design.md:116` and `proposal.md:56`.

**Text:**

> It is `unmeasured` while those merges hold fewer than `min_clusters` residual clusters (for example, on hosted runners without `codex`).

**Problem:**
- The fallback lane is skipped only when `codex` is not on `PATH` (:114). Detection is a `PATH` lookup, and "a missing login surfaces as a failed run" (`internal/modelcli/modelcli.go:45-58`).
- Some failures repeat on every call: an expired login, an account that cannot use `gpt-6.1-terra`, or a CLI that reports another model (which voids the answer, :114). Each one ends `unresolved (terra_failure)` and the lane reports `failed`, not `skipped`.
- The R6-01 filter then drops every merge that had an escalation. It keeps every merge without one, because "answered every cluster" is trivially true when there are none.
- Leaving out `failed` runs would not help. A run with no escalation reports `ran` with 0 calls (§6 Lane health, :179).
- So once `min_clusters` is reached the rate reads a measured 0.00, and the pending-new half of the guard cannot fire. This is the R6-01 defect, reached through `failed` instead of `skipped`.
- [INFERENCE] At about 2.2 clusters per merge and 27 of 100 escalated, roughly half of a 20-merge window's ~44 clusters fall in merges with no escalation, which is near the floor.

**Fix**, in both files, replace that sentence with:

> It is `unmeasured` while those merges hold fewer than `min_clusters` residual clusters (for example, on hosted runners without `codex`) or no cluster the fallback answered (for example, when every fallback call fails).

**Effect:**
- The answered clusters are readable from existing row fields (outcome, raw answer, run id), so no new field is needed.
- The cited figures are unchanged: fastapi has 20 answered clusters and axum 32, so they stay measured at 14/44 and 27/109.
- A window with no escalations and a working fallback now reads `unmeasured`, not 0.00. Its fallback rate is 0, so the guard does not fire either way.

## Considered and not raised

- **Calls that fail only sometimes.** These leave a downward bias, as `fallback_cap` does. The rate still rests on real answers, and the guard's recall is already disclosed as unknown (X11).
- **A merge decided in more than one run** (a workflow re-run). Taken together, the two filter clauses give a clear, conservative rule: such a merge is left out if any of its runs skipped the fallback.
- **A login check.** :170 describes a "logged in" check, but the code only checks `PATH`. This predates round 6. A real login check would still not prevent R7-01, because model availability and model mismatch also fail every call.

## Where the peers changed the lead's view

The lead and the challenger found R6-01 complete. The challenger reported zero issues. The experiment peer found R7-01, the failure twin of R6-01's skip case, and the council adopts its replacement text.

## Open questions that need the user

None.

## Verdict

**1 ISSUE** (low: R7-01). No sign-off in this round. A later sign-off must name the SHA-256 digests of the revised `docs/scout-design.md` and `proposal.md`, and holds only while those digests are unchanged. Round 8 checks only the sentence this fix changes.
