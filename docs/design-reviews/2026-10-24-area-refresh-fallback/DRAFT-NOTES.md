# Draft notes — revision 2.6 proposal

Round-5 draft (round-4 nits fixed, `round-4-response.md`); not yet re-reviewed by the council, not signed off. Line numbers refer to `scout-design.proposed.md` unless stated.

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` (byte copy of `docs/scout-design.md`, rev 2.5) | 473 | `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9` |
| `scout-design.proposed.md` (rev 2.6) | 477 | `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f` |
| `proposed.diff` | 159 | 27 insertions (26 text, 1 blank), 23 deletions (473 + 27 − 23 = 477); 14 hunks (`-1,6 +1,6`, `-8,10 +8,12`, `-20,7 +22,7`, `-95,16 +97,16`, `-179,7 +181,7`, `-198,7 +200,7`, `-340,7 +342,7`, `-349,7 +351,7`, `-371,9 +373,9`, `-403,7 +405,8`, `-412,8 +415,8`, `-427,12 +430,12`, `-449,9 +452,9`, `-470,4 +473,5`); `fc69e6bf8e49f90123bd7203f07fa9bfa714f6e61895c7168a9e2ebd278c2c8d` |
| `requirements.proposed.diff` (base `SCOUT_REBUILD_REQUIREMENTS.md`, SHA-256 `9efcea1b0248d43d0ab831f0924f055f16bb39a3e008365d1e64c7d68a6ece2d`; patched result `bf6b8d2944abb757dea13c36ee85ebf4159ebf5d9535ce31827357362f8f649b`) | 20 | two hunks (`-1,6 +1,6`, `-24,7 +24,7`), 3 insertions, 3 deletions (header line 1, header line 3 with two substitutions, REQ-13 line 27); line 3 names revision 2.6, the design digest `be49efd5…1a6f` and `round-5.md` (the sign-off round; a later round updates both); `35df910ef6d512a6a96a8372394f0fa00ebfd25e7fea7e6023de467b7ccfc6a4` |
| `PROPOSAL.md` | 179 | see the digest line below (recomputed after the last edit) |

PROPOSAL.md digest: `45ff11221040f65bdb9b7cd33b2d77f26f9725ad469028dd0659235df8c520d6`. This file (`DRAFT-NOTES.md`) cannot contain its own digest; it is recorded in `round-4-response.md`.

Digests via `shasum -a 256`; line counts via `/usr/bin/grep -c ''`; insertion and deletion counts count `+`/`-` diff lines excluding the two header lines. Generation:

- `proposed.diff`: `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md > proposed.diff` (BSD diff, no timestamps; never the shell builtin). Re-running it piped into `cmp - proposed.diff` matched. `/usr/bin/patch --dry-run -i proposed.diff -o /dev/null scout-design.base.md` applied cleanly.
- `requirements.proposed.diff`: `/usr/bin/diff -u --label SCOUT_REBUILD_REQUIREMENTS.md --label SCOUT_REBUILD_REQUIREMENTS.proposed.md SCOUT_REBUILD_REQUIREMENTS.md SCOUT_REBUILD_REQUIREMENTS.proposed.md`, run in `/tmp/reqdiff4` (round 3 submission; the round-4 resubmission edits line 3 in place: `round-4.md` → `round-5.md`, regenerated with the same command) on a byte copy of the repository file and a copy with the REQ-13 line and header lines 1 and 3 replaced (not kept). `/usr/bin/patch -p1 --dry-run -i requirements.proposed.diff` from the repository root applied cleanly to `SCOUT_REBUILD_REQUIREMENTS.md` (digest above).
- `scout-design.proposed.md` was produced from the base by exact-substring replacements (each asserted unique) in an eval cell; the replaced strings are the "Was" columns of `PROPOSAL.md`.
- The base digest equals `docs/scout-design.md` at the time of this draft (`shasum -a 256 docs/scout-design.md` → `032285b3…`).

## Every changed line (proposed file)

- Line 3: status line, **revision 2.6 (2026-10-24)** (rev 2.5 no longer bold) and two Phase 4 qualifiers: "Phases 1–4 are implemented (`docs/STATUS.md` §1; Phase 4 to the pre-2.6 acceptance: the revision 2.6 agent-map schedule awaits the companion code change, §4, §11)" and "the Phase 4 scenario ran (under the pre-2.6 acceptance; revision 2.6 amends it, §11)" (D1; round-1 issue 5).
- Line 11: one sentence appended to the Revision 2.5 paragraph, "(X2b has since run; the statements above that X2b is planned are historical, and revision 2.6 applies its result.)" (D2). Line 12 is an unchanged blank.
- Line 13 (+ blank line 14): new **Revision 2.6** paragraph (D2); round 2: section list "§§1, 4, …" and R15 "records the onboarding re-attempts (§4)"; round 3: closing sentence enumerates "the closing note on the Revision 2.5 paragraph that X2b has since run".
- Line 15: evidence notation, "2.1, 2.2, 2.3, 2.4, 2.5 or 2.6" and the sentence "Revision 2.6's agent-map schedule rests on X2b (R-042) … and on the user's direction (R-043)" (D3).
- Line 25: §1 Mapping bullet (D4).
- Line 100: §4 heading "(DECIDED 1, 7 and 21 applied)" (round-1 issue 6c).
- Line 103: §4 **Agent-map schedule** (was "Refresh triggers (DECIDED)") (D5); timing parenthetical's closing clause replaced by X2b's recorded wall times (round-1 issue 4), the rest kept verbatim after "Timing per agent map ("; clauses (a) and (b) restated in round 2 (`round-2.md` issue 2: `skipped` decided before the egress check, `--allow-egress` as the only authorization, the documented caller authorizing only on `workflow_dispatch`).
- Line 104: §4 Areas tag (D6).
- Line 109: §4 Model CLI, two wording changes (D7).
- Line 184: §7.1 `push` clause (D8).
- Line 203: §7.4 trusted-job clause (D9).
- Line 345: §10.2 clawpatch row (D10).
- Line 354: §10.3 (D11; REQ-13 added to the synchronisation list, round-1 issue 10).
- Lines 376–378: §11 Phase 4 title, scenario, acceptance (three edits in the acceptance sentence) (D12); round 2: the scenario states every step runs with `codex` on `PATH` and egress authorized.
- Line 408: §12 R15 wording (D13; round 2: the retry ends only when the pin changes; round 3: `Refresher.refresh` symbol cite replaces refresh.go line cites); line 378: `reconcileAreas` symbol cite replaces `map.go:102` (round 3); line 409: **new R16** row (D13).
- Line 418: §13 X2 decision cell tail (D14); line 419: §13 X2b row replaced (D14).
- Line 433: §14 A1 title and last sentence; line 438: A6 (D15).
- Line 455: Appendix header; line 457: item 1 marked superseded; line 476: item 20 restored verbatim from rev 2.5 plus the marker "Pending clause closed 2026-10-24 by item 21" (round-1 issue 3); line 477: **new item 21** (D16).

Offsets: +2 (Revision 2.6 paragraph and its blank), +1 (R16), +1 (item 21); total +4 (477 = 473 + 4).

Unchanged on purpose: Revision 2.2, 2.3, 2.4 paragraphs; §3.4; §5.1; §6; §9 line 4 ("agent-refresh value"); §10.2 `cmd/scout/main.go` and workflow rows; R13, R14; Appendix 8, 15; every `refresh-map` / `scout map --refresh` mention (the manual path).

## Judgement calls for the council

The nine calls are in `PROPOSAL.md` ("Judgement calls for the council"); the ones most likely to change the text:

1. **Onboarding semantics** (retry on every default-branch push/nightly run until one agent map succeeds, as `Refresher.Run` (`internal/engine/refresh.go`) does, at commit `f7b480e` (`:126`) and in the companion change). Written into §4, §7.1, Appendix 21, REQ-13, R15 and Phase 4; the council confirmed it in round 1 (`round-1.md` issue 1: onboarding, not an automatic remap). Alternatives (rejected): attempt once; onboard only from `scout init`/manual.
2. **R-042 (X2b) and R-043 (user direction) as two ledger rows**, following R-040/R-041.
3. **0.5 kept** as an explicit **[ASSUMPTION]** (no threshold evidence-backed); nothing proposed.
4. **`mapping.refresh.debounce` becomes a hard policy parse error** (strict `KnownFields`, `internal/policy/policy.go:204,232`), per the clean-cutover rule; no shim.
5. **`scout map` flags**: keep `--event`/`--default-branch`/`--refresh`, drop `--base`/`--head`.
6. **Appendix 1 marked superseded, not rewritten; Revision 2.5 paragraph gets one appended sentence.**
7. **Phase 4 scenario must be re-run** after the code change; the design says the recorded run was under the old acceptance.
8. **R16** (areas stale between agent maps) is added content the user did not ask for; it records the consequence and adds no mechanism.
9. **Code cites.** `internal/clawpatch/client.go`, `provider.go`, `internal/modelcli/modelcli.go`, `internal/engine/assess.go` and the workflow cites are to the working tree (files the companion change does not rewrite); R15 and Phase 4 cite `internal/engine/refresh.go` and `cmd/scout/map.go` by symbol (`Refresher.refresh`, `reconcileAreas`), not by line, because the companion change rewrites both files (in progress in the working tree at round 3).

## Checks for the main agent

`shasum -a 256 docs/scout-design.md docs/design-reviews/2026-10-24-area-refresh-fallback/scout-design.base.md` must both equal `032285b3…`; `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md` (run in this directory) must reproduce `proposed.diff` byte for byte; `shasum -a 256 SCOUT_REBUILD_REQUIREMENTS.md` must equal `9efcea1b…` before `requirements.proposed.diff` applies (`/usr/bin/patch -p1 --dry-run -i … ` from the repository root); the council's final round must name the proposed file's and the requirements diff's SHA-256 (any later edit changes them). After editing `PROPOSAL.md`, recompute its digest here. The sign-off round also names the `PROPOSAL.md` and `DRAFT-NOTES.md` digests (`round-2.md`, sign-off conditions); `DRAFT-NOTES.md`'s is in the latest `round-<N>-response.md`.
