# Design review — revision 2.6 (agent map at onboarding and manual `scout map` only) — round 5

Reviewed: `PROPOSAL.md` (179 lines), `DRAFT-NOTES.md` (63 lines), `requirements.proposed.diff` (20 lines; base `SCOUT_REBUILD_REQUIREMENTS.md`), `scout-design.proposed.md` (477 lines), `proposed.diff` (159 lines), `round-4.md`, `round-4-response.md`; base `scout-design.base.md` = `docs/scout-design.md` rev 2.5. Digests below are those recorded in `round-4-response.md`; the council cannot compute SHA-256. Evidence read: the working-tree code named below (`internal/engine/refresh.go:68-120`, `cmd/scout/map.go:60-95`, `internal/clawpatch/client.go:200-238`, `.github/workflows/scout-nightly.yml:48-54`), `SCOUT_REBUILD_REQUIREMENTS.md:1-6,24-30,54`, `.git/HEAD` → `refs/heads/main` = `f7b480e89bc5d9ce13dfaea6f2da3ad1d929acbf`, `round-1-response.md:20`, `round-2-response.md:18` (requirements generation directories `/tmp/reqdiff3`, `/tmp/reqdiff4`).

Reviewed files (SHA-256): `scout-design.base.md` `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9`; `scout-design.proposed.md` `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f`; `proposed.diff` `fc69e6bf8e49f90123bd7203f07fa9bfa714f6e61895c7168a9e2ebd278c2c8d`; `requirements.proposed.diff` `35df910ef6d512a6a96a8372394f0fa00ebfd25e7fea7e6023de467b7ccfc6a4`; `PROPOSAL.md` `45ff11221040f65bdb9b7cd33b2d77f26f9725ad469028dd0659235df8c520d6`; `DRAFT-NOTES.md` `17930c799b5c25d9010a5093fecf99c7f578f658f8ad705dc88f487f8c21192b`

Digests as submitted (`round-4-response.md`): `scout-design.base.md` `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9` (473 lines); `scout-design.proposed.md` `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f` (477); `proposed.diff` `fc69e6bf8e49f90123bd7203f07fa9bfa714f6e61895c7168a9e2ebd278c2c8d` (159); `SCOUT_REBUILD_REQUIREMENTS.md` (base) `9efcea1b0248d43d0ab831f0924f055f16bb39a3e008365d1e64c7d68a6ece2d`; `requirements.proposed.diff` `35df910ef6d512a6a96a8372394f0fa00ebfd25e7fea7e6023de467b7ccfc6a4` (20; patched result `bf6b8d2944abb757dea13c36ee85ebf4159ebf5d9535ce31827357362f8f649b`); `PROPOSAL.md` `45ff11221040f65bdb9b7cd33b2d77f26f9725ad469028dd0659235df8c520d6` (179); `DRAFT-NOTES.md` `17930c799b5c25d9010a5093fecf99c7f578f658f8ad705dc88f487f8c21192b` (63).

## Verdict: **ZERO ISSUES — signed off** (0 major, 0 minor, 0 nits), conditional on the main-agent checks in "Sign-off" below.

All four round-4 nits landed with the council's replacement text verbatim; the two same-class sweep edits of `round-4-response.md` landed; no same-class residue ("today" / "working tree" / "now" attached to a code line cite that the companion change has moved) remains in `PROPOSAL.md` or `DRAFT-NOTES.md`. `scout-design.proposed.md` and `proposed.diff` are unchanged since round 4 by every observable the council has (line counts 477 / 159; lines 3, 13, 354, 376–378, 408–409, 455–477 read as round 4 verified them; `proposed.diff` lines 6–7, 87–88, 100–101, 109–111 carry the same text). `requirements.proposed.diff` differs from the round-4 file in the one token sign-off condition 1 allowed. Every factual claim introduced by the four fixes holds against the working tree and the proposed design.

## Round-4 fixes — verification

| R4 # | Sev | Council replacement text | Verified at | Result |
|---|---|---|---|---|
| 1 | nit | `as `Refresher.Run` (`internal/engine/refresh.go`) does, at commit `f7b480e` (`:126`) and in the companion change` | `DRAFT-NOTES.md:51` (judgement call 1) | Landed verbatim. Claim holds: working tree `Refresher.Run` (`refresh.go:71-89`) has `case record.LastAgentRefreshAt != nil:` → `RefreshNone`, "onboarded; run `scout map --refresh` to remap" (`:85-86`), else `r.refresh` (`:88`). |
| 2 | nit | `## Facts (read 2026-10-24; code cites are to the pre-change tree, see the note after the change set)` | `PROPOSAL.md:5` | Landed verbatim; the note it points to is the abbreviations paragraph at `PROPOSAL.md:71`, which follows the change-set table. |
| 3 | nit | `At commit `f7b480e` (`internal/engine/refresh.go:106-115,126,129`; the companion change's `Refresher.Run` keeps the same test) onboarding is` | `PROPOSAL.md:83` (judgement call 1, first sentence) | Landed verbatim. `f7b480e` is HEAD (`.git/refs/heads/main`); the "same test" claim holds (`refresh.go:85-86`, above). |
| 4 | nit | D12 "Now" clause: `…unchanged except that inertia series keys are "unchanged by the onboarding and manual agent maps" (round-1 issue 12) and the reconciliation-report cite `cmd/scout/map.go:102` becomes "(`reconcileAreas`, `cmd/scout/map.go`: `next, _ := areas.Reconcile(…)`)" (round-3 issue 1)`; D11 "Now": `…; the synchronisation list "its REQ-29, 31, …" becomes "its REQ-13, 29, 31, …" (round-1 issue 10)`; DRAFT-NOTES `- Line 354: §10.3 (D11; REQ-13 added to the synchronisation list, round-1 issue 10).` | `PROPOSAL.md:64` (D12), `:63` (D11); `DRAFT-NOTES.md:36` | All three landed verbatim (backticks, quotation marks, `…`, round references). Anchors hold: proposed:354 reads "its REQ-13, 29, 31, 35, 36, 38, 39, 41 and §10 are synchronised"; proposed:378 contains "(`reconcileAreas`, `cmd/scout/map.go`: `next, _ := areas.Reconcile(…)`)"; working tree `reconcileAreas` is `map.go:83-95` with `next, _ := areas.Reconcile(` at `:93`. |

Same-class sweep (`round-4-response.md` line 14): (i) `PROPOSAL.md:110` column header reads `What it does (pre-change tree)`; (ii) `DRAFT-NOTES.md:20` reads "The base digest equals `docs/scout-design.md` at the time of this draft". Both landed. Residue search of both files for `today|now|current|working tree|still`: `today` no longer occurs; each remaining hit is true as of this round (`PROPOSAL.md:3` "being implemented in the working tree", `:16` Fact 4 "the working tree holds the in-progress companion change"), not attached to a code line (`:33`, `:93` "none made now", `:101`, `:103`), or explicitly anchored to the pre-change tree (`:71`; `DRAFT-NOTES.md:59` "at round 3").

Sign-off condition 1 bookkeeping: `DRAFT-NOTES.md:3` "Round-5 draft (round-4 nits fixed, `round-4-response.md`)"; `:10` requirements row names `round-5.md`, digest `35df910e…c6a4`, patched result `bf6b8d29…649b`; `:13` PROPOSAL digest `45ff1122…20d6` and the statement that DRAFT-NOTES' own digest lives in `round-4-response.md` (it does: `17930c79…192b`); `:18` generation note names "the round-4 resubmission edits line 3 in place: `round-4.md` → `round-5.md`, regenerated with the same command" (the `/tmp/reqdiff4` directory is the round-3 submission's per `round-2-response.md:18`, consistent); `PROPOSAL.md:3` "round-5 draft (round-4 nits fixed per `round-4-response.md`)". All digests in `DRAFT-NOTES.md` match `round-4-response.md` character for character.

## Issues

None.

## Verified correct

- **`requirements.proposed.diff`:** 20 lines; hunks `-1,6 +1,6` and `-24,7 +24,7`; context lines match `SCOUT_REBUILD_REQUIREMENTS.md:1-6,24-30` (REQ-13 at `:27`); line 8 (patched line 3) embeds `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f` and `docs/design-reviews/2026-10-24-area-refresh-fallback/round-5.md`; the patched REQ-13 agrees with design §4 (a)/(b) (proposed:103) and Appendix 21 (proposed:477): onboarding attempted on default-branch push and nightly until one completes, rerun only on manual `scout map --refresh`, never on PR runs, series keyed to heuristic features and never re-keyed. No other difference from the round-4 description.
- **Change set vs `proposed.diff` (peer, exhaustive):** rows D1–D16 each have their "Now" text at the proposed line(s) DRAFT-NOTES lists (3; 11, 13; 15; 25; 103; 104; 109; 184; 203; 345; 354; 376–378; 408–409; 418–419; 433, 438; 455, 457, 476, 477) and their "Was" text in the base; every one of the 14 hunks' base→proposed line mapping re-derived from context (23→25, 98→100, 101/102/107→103/104/109, 182→184, 352→354, 406→408, 415/416→418/419, 473→476); 23 deletions, 27 insertions (one blank), 473 − 23 + 27 = 477; the "Every changed line" list accounts for all 27 insertions.
- **Design code cites vs the working tree (unchanged from round 4, re-sampled):** `client.go:203-213` skipped-before-egress ordering, `:219-220`, `:230-237`; `scout-nightly.yml:52` `inputs.authorize-egress || false`; pins still `v0.1.4` at `scout-nightly.yml:49`, so PROPOSAL (b) Cutover is undone and "in progress" / "being implemented" remain accurate.
- **Sequencing (round-4 condition 4):** HEAD is `f7b480e`; the companion change is uncommitted; design:3 "awaits the companion code change" and :13 "still implements the debounced trigger until the companion code change lands" are true at adoption.
- **Design-rule compliance:** rule 3 — the change cites X2b (preregistration rev 4 signed off, `reviews/round-4.md`; run commit `f2baaa8`) and the user direction (R-043); rounds 1–5 recorded in this directory with issues, severity, fixes and verdicts; this round carries the explicit sign-off statement naming the reviewed digests.

## Considered and not raised

- **`PROPOSAL.md:93` heading "Companion changes (adoption change set; none made now)"** while the (b) code edits sit in the working tree: the heading refers to the adoption change set (a)–(c), none of which is applied (ledgers, STATUS, design and requirements untouched; (b) is incomplete by its own Cutover row). Round 4 and the round-4 response treated it the same way.
- **Line 100 heading edit** ("DECIDED 1, 7 and 21 applied", round-1 issue 6c) has no change-set row; `DRAFT-NOTES.md:29` records it and no file claims the table is exhaustive — unlike round-4 issue 4, where neither file recorded the edit.
- **Requirements header "REQ-31 to revision 2.5 and this header and REQ-13 to revision 2.6 on 2026-10-24"**: REQ-31 (`:54`) itself says "user, 2026-10-23; design rev 2.5"; reviewed since round 2; nothing misstated.
- **`docs/references/policy-file.md:51`** ("in the working tree but not yet committed"): companion-change documentation, not design text; out of scope.
- **DRAFT-NOTES item 9's claim that the companion change does not rewrite `client.go`, `provider.go`, `modelcli.go`, `assess.go`:** not verifiable without git; a search of those files for refresh/debounce/onboarding wording found nothing related, and every design cite into them holds (round 4 list). No contrary evidence.
- **Round-4's own rationale (issue 2) says `Refresher.Run` is called at `map.go:74`; the working tree has it at `map.go:70`, `reconcileAreas` at `:83-95` (round 4 wrote `:83-94`).** Council-file slips, not submission text; the design cites both files by symbol, so nothing in the reviewed files depends on them. Recorded for the record.
- **Inherited rev 2.5 cites outside the change set** (design:96, :197, :327, §10.2 identity/CLI lines): recorded in round 4; out of this proposal's scope.

## Where the peers changed the outcome

- `experiment-peer`: did not change the verdict. Added the exhaustive D1–D16 row check and hunk line mapping, which the main reviewer had only sampled; found the `map.go:70` correction to round-4's rationale; tested and rejected three hypotheses (stale self/PROPOSAL digest in DRAFT-NOTES; an unrecorded change-set edit beyond round-4 issue 4; REQ-13 drift from §4/Appendix 21). The line-100 and REQ-31-date observations were considered and not raised, as above.
- `innovation-challenger`: did not change the verdict. Independently confirmed all four fixes and both sweep edits verbatim, the digest cross-consistency between `DRAFT-NOTES.md` and `round-4-response.md`, the design/diff spot checks (lines 13, 378, 408; 477/159 lines), and that no misleading temporal residue remains; it again rejected "complete, uncommitted" in favour of "in progress" on the `v0.1.4` pins.

## Not verified by the council (main agent)

SHA-256 digests; `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md | cmp - proposed.diff`; `/usr/bin/patch -p1 --dry-run -i requirements.proposed.diff` from the repository root; HEAD's content (the council read only the ref); byte identity of the round-4 and round-5 `requirements.proposed.diff` outside line 8 (textual comparison with the round-4 description only). No LSP server was used; symbol checks used source reads and text search.

## Sign-off

The innovation council signs off design revision 2.6 with zero issues. The reviewed files are:

- `scout-design.proposed.md` (rev 2.6, 477 lines) SHA-256 `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f`
- `proposed.diff` (159 lines) SHA-256 `fc69e6bf8e49f90123bd7203f07fa9bfa714f6e61895c7168a9e2ebd278c2c8d`
- `requirements.proposed.diff` (20 lines) SHA-256 `35df910ef6d512a6a96a8372394f0fa00ebfd25e7fea7e6023de467b7ccfc6a4`; patched result `bf6b8d2944abb757dea13c36ee85ebf4159ebf5d9535ce31827357362f8f649b`
- base `scout-design.base.md` (= `docs/scout-design.md` rev 2.5, 473 lines) SHA-256 `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9`
- requirements base `SCOUT_REBUILD_REQUIREMENTS.md` SHA-256 `9efcea1b0248d43d0ab831f0924f055f16bb39a3e008365d1e64c7d68a6ece2d`
- `PROPOSAL.md` (179 lines) SHA-256 `45ff11221040f65bdb9b7cd33b2d77f26f9725ad469028dd0659235df8c520d6`
- `DRAFT-NOTES.md` (63 lines) SHA-256 `17930c799b5c25d9010a5093fecf99c7f578f658f8ad705dc88f487f8c21192b`

The sign-off is conditional on the main agent, on the files exactly as reviewed and before any adoption step:

1. recomputing all seven digests with `shasum -a 256` and finding each equal to the value above (and `shasum -a 256 docs/scout-design.md` = `032285b3…acf9`, `shasum -a 256 SCOUT_REBUILD_REQUIREMENTS.md` = `9efcea1b…ce2d`);
2. `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md | cmp - proposed.diff` (run in this directory) exiting 0;
3. `/usr/bin/patch -p1 --dry-run -i docs/design-reviews/2026-10-24-area-refresh-fallback/requirements.proposed.diff` from the repository root applying cleanly.

Any mismatch voids this sign-off and requires a new round. On success, adopt in the order PROPOSAL (c) fixes: replace `docs/scout-design.md` with the signed-off proposed file; apply exactly `requirements.proposed.diff`; then the ledger rows R-042/R-043 and the HD5b History line (PROPOSAL (a)); **then** commit the companion code change (b) (committing it before adoption would falsify design:3 and :13 and need its own round); then `docs/STATUS.md` with the signed-off digest. Any later edit to the design or the requirements needs a new council round.
