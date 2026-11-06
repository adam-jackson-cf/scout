# Round 4 response — revision 2.6 (agent map at onboarding and manual `scout map` only)

All 4 nits of `round-4.md` are accepted and fixed with the council's replacement text as given. `scout-design.proposed.md` and `proposed.diff` are not edited. Not signed off: round 5 re-reviews.

| # | Sev | Fix | Where |
|---|---|---|---|
| 1 | nit | `as `internal/engine/refresh.go:126` does today` replaced by `as `Refresher.Run` (`internal/engine/refresh.go`) does, at commit `f7b480e` (`:126`) and in the companion change`. | DRAFT-NOTES judgement call 1 |
| 2 | nit | Heading now `## Facts (read 2026-10-24; code cites are to the pre-change tree, see the note after the change set)`. | PROPOSAL Facts heading |
| 3 | nit | `Today (…)` replaced by `At commit `f7b480e` (…; the companion change's `Refresher.Run` keeps the same test)`. | PROPOSAL judgement call 1 |
| 4 | nit | D12 "Now" clause and D11 "Now" cell replaced as given (round-1 issue 12, round-3 issue 1, round-1 issue 10); DRAFT-NOTES "Line 354" bullet now records the REQ-13 addition. | PROPOSAL D11, D12; DRAFT-NOTES |

Sign-off condition 1: `requirements.proposed.diff` changes in one token (line 8, `round-4.md` → `round-5.md`); 20 lines, hunks `-1,6 +1,6` and `-24,7 +24,7` unchanged; regenerated with `/usr/bin/diff -u --label SCOUT_REBUILD_REQUIREMENTS.md --label SCOUT_REBUILD_REQUIREMENTS.proposed.md`. DRAFT-NOTES: header "Round-5 draft", requirements row (new digests, `round-5.md`), PROPOSAL digest line, "Line 354" bullet, item 1, generation note. PROPOSAL status line: round-5 draft.

Same-class sweep (rounds 3–4) of PROPOSAL and DRAFT-NOTES, beyond the four nits: (i) PROPOSAL (b) table column header `What it does today` → `What it does (pre-change tree)`; (ii) DRAFT-NOTES `equals `docs/scout-design.md` today` → `at the time of this draft`. Left as is, covered by the abbreviations note or true: Fact 5 and the (b) inventory line cites, D8/D12 evidence cells (`refresh.go:106-115,126,129`, `map.go:90-104`), JC1's `refresh.go:147-152`/`:129,163`, the "Explicitly not changed" cites; the status line's "being implemented in the working tree" and Fact 4's "the working tree holds the in-progress companion change" (both true as of this round); the heading "Companion changes (adoption change set; none made now)" (refers to the adoption change set (a)-(c), none of which is applied).

## Digests (`shasum -a 256`)

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` (= `docs/scout-design.md`, rev 2.5; unchanged) | 473 | `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9` |
| `scout-design.proposed.md` (unchanged) | 477 | `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f` |
| `proposed.diff` (unchanged) | 159 | `fc69e6bf8e49f90123bd7203f07fa9bfa714f6e61895c7168a9e2ebd278c2c8d` |
| `SCOUT_REBUILD_REQUIREMENTS.md` (base, unchanged) | | `9efcea1b0248d43d0ab831f0924f055f16bb39a3e008365d1e64c7d68a6ece2d` |
| `requirements.proposed.diff` (patched result `bf6b8d2944abb757dea13c36ee85ebf4159ebf5d9535ce31827357362f8f649b`) | 20 | `35df910ef6d512a6a96a8372394f0fa00ebfd25e7fea7e6023de467b7ccfc6a4` |
| `PROPOSAL.md` | 179 | `45ff11221040f65bdb9b7cd33b2d77f26f9725ad469028dd0659235df8c520d6` |
| `DRAFT-NOTES.md` | 63 | `17930c799b5c25d9010a5093fecf99c7f578f658f8ad705dc88f487f8c21192b` |

Verification by command: `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md | cmp - proposed.diff` matched (design and `proposed.diff` untouched); `/usr/bin/patch -p1 --dry-run -i requirements.proposed.diff` from the repository root applied cleanly to `SCOUT_REBUILD_REQUIREMENTS.md` (digest `9efcea1b…`, unchanged); the patched-result digest was computed with `patch -o` to a temp file, then `round-4.md` → `round-5.md` (deleted).

Not changed: `scout-design.proposed.md`, `proposed.diff`, `scout-design.base.md`, `round-*.md` council files, `docs/scout-design.md`, `SCOUT_REBUILD_REQUIREMENTS.md`, the ledgers, `docs/STATUS.md`, code, `.codegraph`.
