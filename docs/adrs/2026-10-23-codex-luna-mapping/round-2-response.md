# Round 2 response — revision 2.5 (codex `gpt-6-luna` `xhigh` mapping)

All 14 issues of `round-2.md` are accepted and fixed; the council's replacement text was used where given. Line numbers are in the revised `scout-design.proposed.md` (473 lines) and `PROPOSAL.md` (84 lines). Digests are in `DRAFT-NOTES.md`; `proposed.diff` was regenerated with the documented command and verified with `cmp`. Not signed off: round 3 re-reviews. X2b draft citations use section names, not line numbers, because that draft is being revised in parallel.

| # | Sev | Fix | Where |
|---|---|---|---|
| 1 | minor | "Also: §4 scopes E3's and CP07's … X2b row lists the planned repeat." inserted before "Apart from revision bookkeeping"; P2 reads "also §4's E3/CP07 scoping, R15, the §13 X2 record and X2b row; nothing else beyond bookkeeping". | design 11; PROPOSAL 39 |
| 2 | minor | Line 101 clause replaced (neither recorded a served model; "neither is evidence about `gpt-6-luna` at `xhigh`"; smokes "not evidence"). Fact 3 names the unsigned X2b draft's repeat of the timings (cited as PREREGISTRATION §2 "Time"); Fact 5 sentence replaced; P5 mirrors line 101; R-040 caveat reads "one-run operational observations (repeated in the unsigned X2b draft), not evidence"; judgement call 1 reworded in PROPOSAL and DRAFT-NOTES. | design 101; PROPOSAL 12, 14, 42, 62, 72; DRAFT-NOTES 41 |
| 3 | minor | A1: "(no cross-repository claim; two clusters, no cluster-robust interval)" and the 5 % operating-characteristics sentence after the guide name. P10 mirrors. | design 430; PROPOSAL 47 |
| 4 | minor | X2 cell and X2b row replaced with the council's text (pass set, contiguity, grid minimum, scope limits; combination fixed in X2b's council-signed preregistration; existence-of-threshold question; map schedule; adoption of `match_jaccard` value, `else` branch or 0.5 kept). PROPOSAL "Explicitly not changed" updated. | design 415–416; PROPOSAL 55 |
| 5 | minor | R15 replaced with the council's row (no model or provider fallback; same-name change undetected; new council-signed study for a new model). P12 mirrors. | design 406; PROPOSAL 49 |
| 6 | nit | `internal/modelcli/modelcli.go:41-42`. | design 107 |
| 7 | nit | "X1–X6 including X2b, X9, X10". The §1 scope line (29) keeps its pre-existing "X1–X6, X9, X10" list (not a changed line). | design 13 |
| 8 | minor | R-040 and R-041 rows rewritten (Informed, Key point "no provider fallback", Caveat); ledger note appended with the competing X2b DRAFT-NOTES drafts (cited by file name, not line). | PROPOSAL 72–73, 76 |
| 9 | nit | Fact 5: pass set, contiguity, grid minimum, arm B `fallback`, `not replicated`, combined with the issue-2 sentence. | PROPOSAL 14 |
| 10 | nit | "No provider value is refused for mapping and no `claude` path exists." | PROPOSAL 10 |
| 11 | nit | Title "The 0.5 default value (A1, §4 Areas threshold)."; "signed off and run". | PROPOSAL 55 |
| 12 | nit | Ledger note: HD5b is X2b's own ledger step; Companion change (c) adds the wording-only HD5 History line; DRAFT-NOTES judgement call 3 lists it. | PROPOSAL 76, 80; DRAFT-NOTES 43 |
| 13 | nit | "PROPOSAL 14, 76". | `round-1-response.md:21` |
| 14 | nit | `proposed.diff` regenerated (120 lines, +18 / −13, hunk ranges unchanged, `cmp` OK); digests recomputed with `shasum -a 256`; changed-line annotations and judgement calls 1 and 3 rewritten; status lines set to round-3 draft. | DRAFT-NOTES; PROPOSAL 3 |

Consistency pass (by command): `docs/scout-design.md` digest equals the base digest (`5c16a807…`); `requirements.proposed.diff` unchanged (`4cfdd24f…`); line counts base 468, proposed 473 (changed proposed lines 3, 11, 13, 14, 101, 102, 107, 153, 156, 343, 352, 406, 415, 416, 430, 452, 461, 473); hunk list identical to round 2; PROPOSAL line references (P-rows, ledger note at 76) re-checked against the file. Also aligned without a council item: Fact 5 "DRAFT rev 1" → "DRAFT" (the X2b draft is at rev 2); PROPOSAL P10 lists the new A1 scope items; the ledger note (PROPOSAL 76) cites the X2b preregistration by section name (§12 "Sources and ledgers") instead of the old `:100` line.
