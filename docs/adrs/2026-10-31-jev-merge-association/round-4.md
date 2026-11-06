# Design review round 4: DRAFT revision 2.8, Jev merge-time area association with a reasoning-model fallback

Date: 2026-10-31. Rules: CLAUDE.md design rules 1–3. Reviewers: the innovation council (lead), with independent peers `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). Read-only review: no shell, no model calls, `.codegraph` untouched.

Reviewed files (SHA-256): `docs/scout-design.md` `acc075d56bdd1f8660ff60c22c6f0fb828d5edacfe07fd2840c72700dfbad151`; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` `877e1298102bc191139f19790c45026c96702fcea9a0aa965db744a80360820c`

## Scope

Lines changed by the six round-3 fixes:
- `docs/scout-design.md` :19, :114, :116, :361 and :500;
- `proposal.md` :11, :14, :16 and :56;
- anything those fixes broke.

Settled decisions are not reopened (R-059, R-061): pending area at merge time; maintainer acceptance; provisional opt-in, off by default; design rule 3 exception; fallback `codex exec gpt-6.1-terra` high only.

## Round-3 fixes: verification

| R3 | Status | Evidence |
|---|---|---|
| 01 inherited records | Applied; gap for renames inside a pending area (R4-01) | :114; proposal :11, :16. Matches measured rule effects for owned files (`gold.py:170-186` move/add ops), but not for pooled renames (`frontend.py:361-365`, `owners: []`, `pool: True`) |
| 02 supersede, no retire | Applied; protection lost after a match (R4-02); one-area vs union test (R4-04) | :114, :500(a); proposal :14; `areas.go:152` sets `Origin = OriginAgent` on match |
| 03 skipped fallback in denominator | Applied; dilution from Jev-assigned clusters remains (R4-03) | :116; proposal :56 |
| 04 translation rule | Applied and accurate | `_LANG_DIR` (`frontend.py:51-61`) matches `docs/api/` and `docs/src/` |
| 05 :19 Appendix amendments | Applied | consistent with :499, :500(e) |
| 06 :361 re-anchor target | Applied | consistent with :98, :378, :426, :462 |

Cited figures re-derived: rounds 8–9 give 20/44 escalated and 14/44 `new`; axum gives 27/109. No Terra call failed and no fallback was skipped in either run, so these figures hold under every denominator discussed here.

## Issues

### R4-01 (medium): a rename inside a pending area becomes an owner, which the measured runs never allowed

**Where:** :114 and proposal :11.

**Current text:** "A file resolved by the rename or variant-sibling rule gets an association record with outcome `inherited` (…the inherited area ids…). Its label comes from that record, and it counts as an owner afterwards."

**Why it is wrong:** The "or stays in its pending area" branch of the rename rule is also resolution by the rename rule, so that file gets an owning record. Two things then go wrong:
- When the file is modified later, it counts as "held by an owner" and is not decided. In the measured runs it was a coverage gap (`modified_unowned:pooled`, `frontend.py:337-342`).
- An added variant sibling would "inherit" an empty or pending owner, and would be left unowned without being reported. In the measured runs it was residual, because `variant_owners` reads area owners only (`frontend.py:263-268`).

Both contradict "`Area.Active()` is false for a pending area, so no deterministic lookup treats its files as owned".

**Fix.** At :114 and proposal :11, replace from "A file resolved by the rename or variant-sibling rule" to "…replaces its old path in that pending area's file set." with:

> A file that inherits an owner by the rename or variant-sibling rule gets an association record with outcome `inherited` (rule `rename` or `variant`, the old path or sibling, the inherited area ids, never a pending area; no model call), with the Jev and fallback fields empty. Its label comes from that record, and it counts as an owner afterwards. A renamed file in a pending area gets no association record and is not an owner, as a pooled file owned nothing in the measured runs (`frontend.py` `owned_classify` `renamed_pooled`, `variant_owners`); it replaces its old path in that pending area's file set.

At :114, replace "An active area or an association record counts as an owner" with "An active area, or an association record that assigns the file to an active area, counts as an owner".

### R4-02 (low): the retirement protection for accepted areas ends at their first match

**Where:** :114, :500(a) and proposal :14.

**Current text:** "`Reconcile` never retires an area of origin `association` (a maintainer accepted it)".

**Why it is wrong:** On every match `Reconcile` sets `area.Origin = OriginAgent` (`internal/areas/areas.go:152`; :110 "matched areas update file set and provenance"). After one matching map, an unmatched, fully heuristic-owned later map retires the area (`areas.go:161-168`). That undoes the maintainer acceptance the sentence cites as its reason.

**Fix.**
- At :114 and proposal :14, replace "An accepted area becomes active with origin `association`, and from then on `Reconcile` matches it like any area, except that `Reconcile` never retires an area of origin `association` (a maintainer accepted it); an unmatched one is kept." with:

  > An accepted area becomes active with origin `association` and records the maintainer's acceptance (`accepted_by`, `accepted_at`) in the `areas` payload. From then on `Reconcile` matches it like any area, updating file set, title, Clawpatch ids and origin. It never retires an area that carries a maintainer acceptance, before or after a match; an unmatched one is kept.

- At :500(a), replace "; `Reconcile` does not retire an accepted area of origin `association`" with "; `Reconcile` never retires an area carrying a maintainer acceptance, which survives matches".

### R4-03 (low): partial fallback availability still biases the pending-new rate low

**Where:** :116 and proposal :56.

**Current text:** "(valid `new` answers ÷ the residual clusters counted in the fallback rate, less those left `unresolved (fallback_cap)` or `unresolved (terra_failure)` and those not sent because `association-fallback` was skipped)".

**Why it is wrong:** The Jev-assigned clusters of runs where the fallback was skipped or capped stay in the denominator, and they can never be `new`. So the guard still under-fires in the window R3-03 targeted. Example: run A has 20 clusters, 6 escalated and 4 `new`; run B has 20 clusters and 6 escalated, with the fallback skipped. The current text gives 4/34 = 0.118, while run A alone shows 0.20.

**Fix.** Replace the parenthesis with:

> (the fallback rate × the share of valid `new` answers among the clusters the fallback answered; clusters left `unresolved (fallback_cap)` or `unresolved (terra_failure)`, or not sent because `association-fallback` was skipped, are not answered)

The cited figures are unchanged: fastapi 20/44 × 14/20 = 14/44 = 0.32; axum 32/109 × 27/32 = 27/109 ≈ 25 per 100. The example gives (12/40) × (4/6) = 0.20. "`unmeasured` when the fallback answered no cluster" still holds.

### R4-04 (low): association records are superseded by a one-area test, pending areas by a union test

**Where:** :114 and proposal :14.

**Current text:** "an association record whose files are all in an active area's `FileSet` is superseded by the map".

**Why it is wrong:** A cluster whose files a refreshed map spreads over two areas is never superseded. Its stale label and owner status then persist beside the map's labels, even after the manual refresh the guard recommends. The preceding pending-area rule uses "active areas' `FileSet`s", and no reason is given for the difference.

**Fix.** Replace the quoted text with "an association record whose files are all in active areas' `FileSet`s is superseded by the map".

## Considered and not raised

- **Accuracy of inherited labels.** Nothing measured it. The measured runs applied the rename and variant-sibling rules deterministically and reported no agreement figure. The design makes no accuracy claim for them, so this is noted for X11's preregistration only.
- **The translation regex also matches `docs/pt-BR/`.** Because the rule is not applied, the "2–3-letter" wording has no effect.
- **Inherited records carry the `stale` and `thresholds provisional` marks.** This is harmless.
- **Proposal :3 counts round 1 and round 2 issues but not round 3.** Editorial.

## Where the peers changed the lead's view

- **R4-01:** found by the lead and both peers; the replacement text is the experiment peer's.
- **R4-02:** the challenger found it. The lead and the experiment peer had first read the origin flip as consistent with "matches it like any area". The council kept it because the sentence's own reason, maintainer acceptance, does not survive the flip. It prescribes recording the acceptance rather than freezing the origin, because `Reconcile` rewrites origin on every match.
- **R4-03 and R4-04:** the experiment peer.

## Open questions that need the user

None.

## Verdict

**4 ISSUES** (medium: R4-01; low: R4-02, R4-03, R4-04). No sign-off in this round. Any later sign-off must name the SHA-256 digests of the revised `docs/scout-design.md` and `proposal.md`, and holds only while those digests are unchanged. Round 5 checks only the sentences these fixes change.
