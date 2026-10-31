# Design review round 5 — DRAFT revision 2.8: Jev merge-time area association with a reasoning-model fallback

Date: 2026-10-31. Rules: CLAUDE.md design rules 1–3. Reviewers: innovation council (lead), with independent peers `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). Read-only review: no shell, no model calls, `.codegraph` untouched.

Reviewed files (SHA-256): `docs/scout-design.md` `58d8b53088bc814488254431a5536ec7724579af5ab96c04b5862139eb84d3ad`; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` `7370ce5e7a4c4d07cec0a47b3d503418161acc429c1bdc9ea9fb65e5072a77f8`

## Scope

This round checks the four sentences replaced for round 4, and anything they broke:
- `docs/scout-design.md` :114, :116 and :500;
- `proposal.md` :11, :14 and :56. Its lower-case start of the R4-02 sentence, which continues a sentence, is accepted.

Settled decisions are not reopened (R-059, R-061): a pending area at merge time; maintainer acceptance; provisional opt-in, off by default; the design rule 3 exception; the fallback `codex exec gpt-6.1-terra` at high effort only.

## Round-4 fixes: verification

| R4 | Applied | Check | Result |
|---|---|---|---|
| 01 inherited records | yes; design :114, proposal :11 | owners are checked first (`axum/pilot/frontend.py:355-360`); `is_pooled` = in the pool and not owned (`:283-285`); a pooled rename has owners `[]` (`:361-365`); `variant_owners` reads area owners only | plain cases correct; overlap and rule (f) cases wrong (R5-01) |
| 02 maintainer acceptance | yes; :114, proposal :14, :500(a) | on a match `Reconcile` rewrites file set, ids, title and origin (`internal/areas/areas.go:149-155`); an acceptance check in the retirement branch (`:161-168`) is enough | verified, no issue |
| 03 pending-new rate | yes; :116, proposal :56 | fastapi 20/44 × 14/20 = 14/44 = 0.32; axum 32/109 × 27/32 = 27/109 (`README.md` rounds 8–10; `axum/README.md`) | figures correct; no sample floor on the share (R5-03) |
| 04 union test | yes; :114, proposal :14 | wording now matches the pending-area test | stale paths block it (R5-02) |

## Issues

### R5-01 (medium): the pending-rename sentence overrides active ownership and rule (f)

**Where:** :114 and proposal :11.

**Text:** "A renamed file in a pending area gets no association record and is not an owner, as a pooled file owned nothing in the measured runs (`frontend.py` `owned_classify` `renamed_pooled`, `variant_owners`); it replaces its old path in that pending area's file set."

**Problem 1: a file can be pending and owned.** A pending area is superseded only when all its files are mapped. After a partial agent map, one of its files can therefore also belong to an active area. When that file is renamed, two rules apply: the inheritance rule ("inherits its old path's owner") and this sentence ("no association record"). The measured runs checked owners before the pool (`frontend.py:355-365`; `is_pooled` excludes owned files, `:283-285`). There, an owned rename moved to its owners and never stayed in the pool (`gold.py:174-182`).

**Problem 2: rule (f) would decide pooled renames.** Rule (f) says a heuristic-only file "is decided whether added, moved or modified". The measured pooled-rename branch returns before `add_residual` (`:361-365`), so a pooled rename was never decided. Deciding it can create a duplicate pending area.

**Fix.** In both places, replace the quoted sentence with:

> A renamed file whose old path is in a pending area and held by no owner gets no association record, is not an owner and is not decided, even when a heuristic feature owns it, as in the measured runs a pooled file owned nothing and a pooled rename was resolved before rule (f) (`frontend.py` `owned_classify` `renamed_pooled`, `is_pooled`, `variant_owners`); it replaces its old path in that pending area's file set. When an owner holds the old path, the renamed file inherits that owner as above and the old path leaves the pending area's file set (`gold.py`: an owned rename moves to its owners, never to the pool).

### R5-02 (low): stale paths keep both supersede tests from ever passing

**Where:** :114 and proposal :14.

**Text:** "a pending area whose files are all in active areas' `FileSet`s is marked `superseded` … an association record whose files are all in active areas' `FileSet`s is superseded by the map".

**Problem:**
- A record keeps the paths it was decided on. A later rename creates a new `inherited` record but leaves the old path in the old record.
- A deleted file is only dropped from classification. Nothing removes it from a record or from a pending area's file set.
- A path that no longer exists is in no map's `FileSet`, so the test never passes. The stale label and its owner status then persist beside the map, which is what R4-04 targeted.
- Manual-only maps let these stale paths accumulate. The measured membership step deleted removed files and moved renamed ones (`gold.py:8,166-182`).

**Fix.** In both clauses of both places, change "whose files are all in active areas' `FileSet`s" to "whose files that exist at the mapped revision are all in active areas' `FileSet`s".

### R5-03 (low): the pending-new rate can latch the stale flag on a single answer

**Where:** :116 and proposal :56.

**Text:** "(the fallback rate × the share of valid `new` answers among the clusters the fallback answered; …)" and "It is `unmeasured` when the fallback answered no cluster in the window".

**Problem:**
- `min_clusters` bounds residual clusters, not answered clusters, so the share can rest on a single answer.
- Example: 20 residual clusters, 6 of them escalated (fallback rate 0.30). The fallback is skipped on every merge except one, and that merge's single answered cluster is `new`. The rate is 0.30 × 1 = 0.30, over the 0.20 limit, and every association output is then flagged `catalogue-stale` until a manual map runs.
- The round-4 formula gave about 0.07 on the same window, so the new behaviour comes from the R4-03 fix. It conflicts with the stated aim that the rate be "neither one push's noise nor months old".

**Fix.** In both places:
- Replace the parenthesis with:

  > (valid `new` answers ÷ residual clusters, both counted over the window's merges in which the fallback answered every cluster Jev left `unresolved` or `none`; a merge with a cluster left `unresolved (fallback_cap)` or `unresolved (terra_failure)`, or not sent because `association-fallback` was skipped, is left out)

- Replace "It is `unmeasured` when the fallback answered no cluster in the window" with "It is `unmeasured` while those merges hold fewer than `min_clusters` residual clusters".

The cited figures are unchanged, since every escalation was answered in both runs: fastapi 14/44 = 0.32 and axum 27/109. The round-4 example becomes run A alone, 4/20 = 0.20.

## Considered and not raised

- **Variant siblings must exist in the parent tree** (`variant_owners` filters on `state.paths`). The design omits this condition, but it predates round 4 and does not change the measured outcome.
- **:110 "fully heuristic-owned areas retire"** omits the acceptance exception. :114 and Appendix 23(a) state that exception for this DRAFT.
- **Merging a pending area into an existing area** is not specified as recording an acceptance. No round-4 sentence depends on it.
- **Other R4-01 cases match the measured runs:** a modified pending file owned by a heuristic feature is decided (`frontend.py:340-341`); empty renames are trapped before the owner check (`:347-350`); an inherited record whose area retires stops counting as an owner.

## Where the peers changed the lead's view

- The lead found R4-01's pending-only case correct.
- The challenger found the overlap case in R5-01.
- The experiment peer found the rule (f) case in R5-01, plus R5-02 and R5-03. Its replacement text is adopted for all three.
- The council chose a merge-level filter for R5-03 over a new minimum on answered clusters, because the filter reuses `min_clusters`, adds no new assumption key and keeps every cited figure.

## Open questions that need the user

None.

## Verdict

**3 ISSUES** (medium R5-01; low R5-02, R5-03). No sign-off in this round. A later sign-off must name the SHA-256 digests of the revised `docs/scout-design.md` and `proposal.md`, and holds only while those digests are unchanged. Round 6 checks only the sentences these fixes change.
