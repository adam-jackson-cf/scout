# Draft notes — revision 2.5 proposal

Round-6 draft, revised after council round 5 (`round-5.md`: 3 issues; fixes in `round-5-response.md`), round 4 (`round-4.md`: 4 issues; fixes in `round-4-response.md`), round 3 (`round-3.md`: 12 issues; fixes in `round-3-response.md`), round 2 (`round-2.md`: 14 issues; fixes in `round-2-response.md`) and round 1 (`round-1.md`: 17 issues; `round-1-response.md`). Not yet signed off. Line numbers refer to `scout-design.proposed.md` unless stated.

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` (byte copy of `docs/scout-design.md`, rev 2.4) | 468 | `5c16a807513d54667390278755b8daa504e8db5ae0733236f87938c51a9a90ae` |
| `scout-design.proposed.md` (rev 2.5) | 473 | `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9` |
| `proposed.diff` | 120 | 18 insertions (17 text, 1 blank), 13 deletions; hunks `-1,6 +1,6`, `-8,8 +8,10`, `-96,13 +98,13`, `-148,10 +150,10`, `-338,7 +340,7`, `-347,7 +349,7`, `-401,6 +403,7`, `-409,7 +412,8`, `-423,7 +427,7`, `-445,7 +449,7`, `-454,7 +458,7`, `-466,3 +470,4` (the hunk ranges are unchanged from rounds 2 and 3; the `+` text changed); `778842853667a5a882b261ca8875e4a47c7697e72cebcf02eaf2f9bc749b0fda` |
| `requirements.proposed.diff` (base `SCOUT_REBUILD_REQUIREMENTS.md`, SHA-256 `cd84592fc287fd5e4041c13fed827900c4143bdd583aee447ac59c6f65f71da2`) | 11 | 1 insertion, 1 deletion (REQ-31, line 54); hunk `-51,7 +51,7`; `4cfdd24fb96bec040bc4112b79b5b80e75c92bc953bef22876b60f742a2b5463` (unchanged since round 1) |
| `PROPOSAL.md` | 84 | `1a59951c3536e8ac82d377f48996f5b8ceb3eddba618889974267857314a6686` |

Digests via `shasum -a 256`; line counts via `wc -l`; insertion/deletion counts count `+`/`-` diff lines excluding the two header lines. `proposed.diff` generated with `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md > proposed.diff` (BSD diff, no timestamps); re-running it piped to `cmp - proposed.diff` matched (round 5 regeneration). `requirements.proposed.diff` generated with `/usr/bin/diff -u --label SCOUT_REBUILD_REQUIREMENTS.md --label SCOUT_REBUILD_REQUIREMENTS.proposed.md <base> <proposed>`, where `<base>` is a byte copy of `SCOUT_REBUILD_REQUIREMENTS.md` (digest above) and `<proposed>` applies the one REQ-31 substitution; produced in `/tmp/reqdiff`, not kept. `docs/scout-design.md` digest equals the base digest (checked after the round-4 fixes).

## Every changed line (proposed file)

- Line 3: status line, `**revision 2.5 (2026-10-23)**` (rev 2.4 no longer bold) (P1).
- Line 11: **Revision 2.5** paragraph (P2); round 1 issues 1, 2, 4, 10, 11, 12 (design-rule-3 statement now names R-041 and HD5; pass set, contiguity, `fallback`; user direction under X2b's rule; "model option `default`, probe-served"; "no provider value is refused"; argv `:219-220`); round 2 issue 1 (the "Also:" sentence listing §4's E3/CP07 scoping, R15, the §13 X2 record and X2b row before the closing bookkeeping sentence). Line 12 is an unchanged blank.
- Line 13: evidence notation, "2.1, 2.2, 2.3, 2.4 or 2.5" (P3) "product experiments X1–X6 including X2b, X9, X10" (round 2 issue 7) and "signed off and run under Scout's experiment rules" (round 3 issue 7); line 14 is the diff's added blank line.
- Line 101: §4 Refresh triggers (base 99), E3/CP07 model-option scoping (P5; round 1 issue 7; round 2 issue 2: "neither recorded a served model", "neither is evidence about `gpt-6-luna` at `xhigh`", smokes "not evidence"; no smoke timings).
- Line 102: §4 Areas (base 100), assumption tag and CP07 churn scoping (P12; issues 8, 13).
- Line 107: §4 Model CLI bullet (base 105) (P4); issues 6, 10, 12 (`PATH` vs failed call, argv `:219-220`, "model option `default`"); round 2 issue 6 (`internal/modelcli/modelcli.go:41-42`).
- Line 153: §6 table row (base 151) (P6).
- Line 156: §6 Model use paragraph (base 154) (P7).
- Line 343: §10.2 clawpatch row (base 341) (P8).
- Line 352: §10.3 (base 350) (P9).
- Line 406: §12 new R15, model pinned in code (P12; round 1 issue 9; round 2 issue 5: "no model or provider fallback", same-name change undetected, new council-signed study for a new model; round 3 issues 4, 5, 6: "never `refreshed` (`internal/engine/refresh.go:35-37,153-155`)", "whatever codex served as `gpt-6-luna` at `xhigh` during X2b's collection", full path `evaluation/x2-area-stability/RESULTS.md` §10.2).
- Line 415: §13 X2 decision cell, run result and non-application appended (P12; round 1 issue 3; round 2 issue 4: pass set, contiguity, grid minimum, scope limits in A1, the X2/X2b combination fixed in X2b's council-signed preregistration).
- Line 416: §13 new X2b row (P12; round 1 issue 3; round 2 issue 4: existence-of-threshold question, "map schedule", consequence options incl. the `else` branch and 0.5 kept; round 3 issue 2: "decides on the resulting proposal").
- Line 430: §14 A1 (base 426) (P10; round 1 issues 2, 10; round 2 issue 3: two clusters with no cluster-robust interval, the operating-characteristics scope sentence; round 3 issues 3, 6: "fastapi (Python) and axum (Rust) only (no Go or TypeScript repository; …)", "labels, unreviewed by a human", "`evaluation/x2-area-stability/RESULTS.md` §4.3, operating characteristics").
- Line 452: Appendix header, items 8–20, R-040 (base 448) (P11).
- Line 461: Appendix item 8 (base 457) (P11).
- Line 473: new Appendix item 20 (P11; issues 4, 11; round 3 issue 2 and round 4 issue 3: "a later council design round decides on the resulting proposal").

Offsets: +2 from proposed line 11 (paragraph and its blank), +1 after line 406 (R15), +1 after line 415 (X2b), +1 for the new Appendix item 20 at the end; total +5 (473 = 468 + 5).

Unchanged on purpose: §5.1 row 9, §3.1 tree, R13, R14, §4 Failure semantics bullet, Revision 2.2 paragraph, §10.2 workflow row, Appendix 15, and the pre-existing "logged in" wording in §3.4, §6, R13, Appendix 8 and REQ-31 (policy-level, signed off in rev 2.2–2.4; recorded by the council as out of scope).

## Judgement calls for the council

1. Smoke timings (92 s / 169 s) are not in the design text; they remain in `PROPOSAL.md` Fact 3 as operational observations; the unsigned X2b draft repeats them (`evaluation/x2b-area-stability-luna/PREREGISTRATION.md` §2, "Time") (round 2 issue 2).
2. Mapping ignores the provider setting and `CLAWPATCH_PROVIDER` (code-verified, `provider.go:57-66`); they govern only `clawpatch-review`. §10.2's workflow row still says "constrain `CLAWPATCH_PROVIDER` to `codex`/`claude`" and is left (accepted in round 1).
3. Ledger IDs (round 1 issue 5; round 2 issues 8, 12): R-040 = user direction, R-041 = X2; both rows, the `docs/STATUS.md:162` edit and a wording-only `docs/hypothesis-ledger.md` History line for HD5 (`PROPOSAL.md` Companion change (c)) belong to the adoption change set; HD5b is X2b's own ledger step; no ledger, `docs/STATUS.md` or `docs/scout-design.md` edit is made now; the one file edited outside this directory, by the round-3 and round-4 responses, is `evaluation/x2b-area-stability-luna/DRAFT-NOTES.md` (its ledger section in both rounds, its first "Not done" bullet in round 4). Round 3 issue 1 and round 4 issue 1: the rows are shared with the X2b draft and added once; the single canonical text is `PROPOSAL.md`'s R-040/R-041 rows and History line (`evaluation/x2b-area-stability-luna/DRAFT-NOTES.md` now points to them and keeps only HD5b's draft row and History line); the rows are appended, and the History line added at the top, in the rev 2.5 adoption change set, which precedes X2b's HD5b row (it cites both) and X2b's final round.
4. X2's codex arm effort is not recorded in its probes; "low" comes from `~/.codex/config.toml` and is marked [INFERENCE] in `PROPOSAL.md`. The design text says "effort not recorded".

## Checks for the main agent

`shasum -a 256 docs/scout-design.md docs/design-reviews/2026-10-23-codex-luna-mapping/scout-design.base.md` must both equal `5c16a807…`; `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md` (run in this directory) must reproduce `proposed.diff` byte for byte; `shasum -a 256 SCOUT_REBUILD_REQUIREMENTS.md` must equal `cd84592f…` before `requirements.proposed.diff` applies; the council's final round must name the proposed file's and the requirements diff's SHA-256 (any later edit changes them). After editing PROPOSAL.md or DRAFT-NOTES.md, recompute the PROPOSAL digest here.
