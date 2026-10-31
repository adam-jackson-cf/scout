# Design review round 3: DRAFT revision 2.8, Jev merge-time area association with a reasoning-model fallback

Date: 2026-10-31. Rules: CLAUDE.md design rules 1–3. Reviewers: the innovation council (lead), plus independent peers `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). The review was read-only: no shell, no model calls, `.codegraph` untouched.

Reviewed files (SHA-256): `docs/scout-design.md` `7bf173c04c484147e810858ea4bd6c610530cbac727812724ca942bdf7182072`; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` `571ba17f6bd53c90fd13f39a867c1cbc8b16b5d44e716d6d2362482d9d003e92`; `docs/research-ledger.md` `5819f8b2f354b9d4c737737ba32b39086da1a77aa4193b915090ea98641135c2`

## Scope

This round checks:
- the changed lines of `docs/scout-design.md` (:19, :114, :115, :116, :173, :176, :378, :462, :499, :500);
- the changed lines of `proposal.md` (:3, :11, :16, :20, :34, :56, :64, :71);
- R-061 (`docs/research-ledger.md:69`);
- anything these fixes broke elsewhere.

Settled decisions are not reopened (R-059, R-061): pending area at merge time; maintainer acceptance; provisional opt-in, off by default; design rule 3 exception; fallback `codex exec gpt-6.1-terra` at high effort only.

## Round-2 fixes: verification

| R2 | Status | Evidence |
|---|---|---|
| 01 residual rules | applied; matches `owned_classify` (`axum/pilot/frontend.py:311-380`; `build_units.py:247`); two gaps (R3-01, R3-02) and one repository-specific rule (R3-04) | `:114`; proposal step 1 |
| 02 guard rates | applied; 14/44, 20/44, 16 % and axum 25 per 100 are consistent with the new denominator (neither run had `fallback_cap` or `terra_failure`); one gap (R3-03) | `:116`; proposal "Refresh guard" |
| 03 enable/consent | applied | `:114` opening; proposal :20, :24 |
| 04 §6 output | applied verbatim | `:176` |
| 05 threshold versions | applied | `:114`, `:173`; proposal :20 |
| 06 guard rebuild | applied | `:114` row fields, window and flag rule; `:378` acceptance; `normalizeRefresh` pins UTC, so time equality survives a round trip (`internal/state/mapping_refresh.go:83-89`) |
| 07 schema 13, old exports | applied; `:19` and `:361` not updated (R3-05, R3-06) | `:114`, `:462`, `:499`, `:500(e)`; proposal step 6; old manifests need no `absent` entry (`internal/state/export.go` manifest checks) |
| 08 R-061, `:19`, citations | applied | `research-ledger.md:69`; `:19`, `:114`, `:500(c)`; proposal :3, :64 |
| 09 clusters per PR | applied | `:114`; proposal :20, :71 |
| 10 Jev-only recall | applied | `:115` 0.217 (10/46) and 0.500 (23/46), `axum/README.md`; proposal :34 |

## Numbers re-derived

| Figure | Source | Result |
|---|---|---|
| rounds 8–9: 21 PRs, 44 clusters; fallback rate 20/44; pending-new 14/44 | `README.md` rounds 8–10 (units/PRs 21/10, 23/11; escalations 7 + 13; `none` outputs 3 + 11) | correct |
| 0.40 ≈ 1.5 × 27.2; 0.20 ≈ 1.25 × 16.3 % | 25/92; 15/92 | correct |
| axum 25 per 100 `none`; 29.4 vs 27.2 per 100 | 27/109; 32/109, 25/92 | correct |
| 2.2 clusters per PR (92 in 41); axum 120 / 69 / 905 | `README.md`; `FROZEN-AXUM.md` window paragraph | correct as worded |
| R-061: 0.622 [0.514, 0.719] overlaps B2 [0.396, 0.705] | `axum/README.md` | correct |

## Issues

### R3-01 (medium): files resolved by rename or variant sibling have no label source

`:114` and proposal step 1 say: "a renamed file inherits its old path's owner, or stays in its pending area, unless it is empty; an added file with an owned variant sibling … inherits that owner". But owners exist only in an active area's `FileSet`, which is never written between maps (`:114`; `internal/areas/areas.go:150`), and in association records, which are written only for decided clusters. So an inherited file has no recorded owner. On its next modification it becomes a coverage gap, or under rule (f) it is decided again, which was never measured. The measured runs added such files to the membership (`axum/pilot/gold.py` rule effects "(c) renames move, (d) variant siblings join").

**Fix.** At `:114` and in proposal step 1, insert after "…and clusters made only of empty files are inert.":

> A file resolved by the rename or variant-sibling rule gets an association record with outcome `inherited` (rule `rename` or `variant`, the old path or sibling, the inherited area ids; no model call), with the Jev and fallback fields empty. Its label comes from that record, and it counts as an owner afterwards. A renamed file in a pending area replaces its old path in that pending area's file set. Inherited records are not decided clusters, enter neither guard rate, and are not counted by `max_per_run` or the per-merge cluster cap.

At `:114` (and proposal step 6), change "one per cluster per run" to "one per decided cluster per run, plus one per inherited file".

### R3-02 (medium): with rule (f), heuristic ownership still counts as cover at map time

Since R2-01, a file owned only by a heuristic feature is decided by the lane (rule (f)). Three texts still treat heuristic ownership as cover:
- `:114`: "a pending area whose files are all owned by an active area or a heuristic feature is marked `superseded`". Every pending area made of heuristic-only files is therefore superseded at the next agent map, although no area covers it.
- `:114`: "an association record whose files the map now owns is superseded by the map". This does not say whether heuristic ownership counts as the map owning the files.
- `:114`: "An accepted area … from then on `Reconcile` matches it like any area". `Reconcile` retires an unmatched area whose files are all heuristic-owned (`internal/areas/areas.go:161-171`; `:110` "fully heuristic-owned areas retire"). So an accepted area made of heuristic-only files is retired at the next map, undoing the maintainer's acceptance.

**Fix.** At `:114` (and proposal step 4):
- Replace "a pending area whose files are all owned by an active area or a heuristic feature is marked `superseded`" with "a pending area whose files are all in active areas' `FileSet`s is marked `superseded`".
- Replace "an association record whose files the map now owns is superseded by the map" with "an association record whose files are all in an active area's `FileSet` is superseded by the map; heuristic-feature ownership supersedes neither, because rule (f) decides such files".
- After "from then on `Reconcile` matches it like any area", append: ", except that `Reconcile` never retires an area of origin `association` (a maintainer accepted it); an unmatched one is kept".
- Append to Appendix 23(a): "; `Reconcile` does not retire an accepted area of origin `association`".

### R3-03 (low): clusters skipped by the fallback dilute the pending-new rate

`:116` (and proposal "Refresh guard") uses the denominator "the residual clusters counted in the fallback rate, less those left `unresolved (fallback_cap)` or `unresolved (terra_failure)`". Escalated clusters left unowned because `association-fallback` was skipped (`no model CLI`, `no consent`) have no such reason, so they stay in the denominator. In a window where `codex` or the consent was present for only some runs, the guard under-fires.

**Fix.** In both places:
- Extend the parenthesis to "…less those left `unresolved (fallback_cap)` or `unresolved (terra_failure)` and those not sent because `association-fallback` was skipped".
- Replace "Clusters `unresolved (fallback_cap)` or `unresolved (terra_failure)` count in the fallback rate only." with "Clusters `unresolved (fallback_cap)` or `unresolved (terra_failure)`, and clusters not sent because `association-fallback` was skipped (`no model CLI`, `no consent`), count in the fallback rate only."

The cited figures are unchanged, because neither run had a skip.

### R3-04 (low): the measured translation exclusion is a fastapi path pattern

`:114` and proposal step 1 say: "and so are markdown, non-English translation and media files (`excluded_reason`)". The measured rule is `_LANG_DIR = ^docs/([a-z]{2,3}(?:-[A-Za-z]{2,4})?)/` with a language other than `en` (`axum/pilot/frontend.py`). In other repositories it would drop every file under `docs/api/`, `docs/src/` or `docs/css/`. On axum it was inert.

**Fix.**
- Replace "and so are markdown, non-English translation and media files (`excluded_reason`)" with "and so are markdown and media files (`excluded_reason`)".
- Replace "The measured cross-scope rename rule (`frontend.py` `scope`, repository-specific prefixes) has no Scout equivalent and is not applied; this is an unmeasured difference." with:

  > The measured cross-scope rename rule (`frontend.py` `scope`, repository-specific prefixes) and the measured translation rule (`excluded_reason`: any `docs/<2–3-letter directory>/` other than `docs/en/`, a fastapi layout that also matches `docs/api/` or `docs/src/`) have no Scout equivalent and are not applied; these are unmeasured differences.

### R3-05 (low): the revision paragraph omits the Appendix 22 amendment

`:19` says: "and the Appendix (new item 23; item 21 gets a pointer: Appendix 21 is reopened for the pending-area case only)". Since R2-07, item 22(a) also carries a DRAFT amendment (`:499`, `:500(e)`).

**Fix.** Replace that parenthesis with "(new item 23; items 21 and 22 get DRAFT amendments: Appendix 21 is reopened for the pending-area case only, and item 22(a)'s current schema becomes `user_version` 13)".

### R3-06 (low): the cutover row still targets schema 12 for the v1 re-anchor

`:361` says: "the v1 re-anchor path (`user_version` 3 → 12: the plan-jobs backfill, the issue re-anchor, the tables added since)". Under the amended Appendix 22(a) and §3.3 (`:98`, "brought to the current version (12; DRAFT 13)"), a store built to this row would be refused.

**Fix.** Replace that parenthetical with "(`user_version` 3 → current: 12 through revision 2.7, DRAFT revision 2.8 13 including `associations`; the plan-jobs backfill, the issue re-anchor, the tables added since)".

## Considered and not raised

- **Guard flag versus policy changes (challenger).** The flag is recomputed from the exported rows under the current policy. The result is deterministic and the same with or without a restore. A maintainer changing a guard limit re-evaluates an advisory flag; this is acceptable, so no stored flag is required.
- **A modified heuristic-only file in a pending area is decided again.** The measured run behaves the same way (`is_heuristic_only` ignores the pool). That one new concern can produce several pending areas is already disclosed.
- **Which copy of `frontend.py` is cited.** The fastapi run used `x2d-merge-association-pilot/frontend.py`. Its `owned_classify` and `excluded_reason` bodies match the cited axum copy; only `scope()` differs, and `scope()` is not applied.
- **Ownership fixed at round start in the measured runs.** This is covered by the teacher-forcing disclosure at `:115`.
- **R-061.** It matches the user's choice and is limited to the provisional opt-in.

## Where the peers changed the lead's view

- **Experiment peer:** R3-01, R3-03, R3-04 and R3-05.
- **Challenger:** R3-06. Its guard-flag concern was assessed and not raised (above).
- **Lead:** R3-02.
- The R2-01 text written by the council brought rule (f) in without updating the rules for superseding at map time and for retirement in `Reconcile`.

## Open questions that need the user

None.

## Verdict

**6 ISSUES** (medium: R3-01, R3-02; low: R3-03, R3-04, R3-05, R3-06). No sign-off in this round. A later sign-off must name the SHA-256 digests of the revised `docs/scout-design.md` and `proposal.md`, and holds only while those digests are unchanged. Round 4 checks only the lines these fixes change.
