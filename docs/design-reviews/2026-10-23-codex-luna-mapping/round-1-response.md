# Round 1 response — revision 2.5 (codex `gpt-6-luna` `xhigh` mapping)

All 17 issues of `round-1.md` are accepted and fixed; the council's replacement text was used where given. Line numbers are in the revised `scout-design.proposed.md` (473 lines). Digests of the revised files are in `DRAFT-NOTES.md`; `proposed.diff` was regenerated with the documented command and verified with `cmp`. Not signed off: round 2 re-reviews.

| # | Sev | Fix | Where |
|---|---|---|---|
| 1 | major | Design-rule-3 statement replaced: "R-040, R-041, HD5; X2's signed-off result (R-041) is cited for the first time, only to state what it did and did not cover; no default or decision rule is changed on its evidence". PROPOSAL status line replaced with the council's sentence. | design 11; PROPOSAL 3 |
| 2 | major | A1 replaced with the council's text (HD5, R-041, model option `default` probe-served `gpt-6.1-sol`, pass set {0.30} contiguous, 0.30 grid minimum, margin 0.0035, 0.50 failed C1, arm B `fallback`/`not replicated`, scope limits). Line 11 gained "pass set {0.30}, contiguous;" and "`fallback`, `not replicated`; scope limits in A1". | design 430, 11 |
| 3 | major | X2 decision cell extended with the run result and its non-application; new §13 X2b row added (text as required). PROPOSAL "Explicitly not changed" and DRAFT-NOTES updated (X2 row no longer "unchanged"). | design 415–416; PROPOSAL 55; DRAFT-NOTES |
| 4 | minor | Appendix 20 last sentence and line 11 clause replaced with the council's wording (X2's rule including `else`; combination fixed in X2b's preregistration, adopted by a later council design round). PROPOSAL P11 updated. | design 473, 11 |
| 5 | minor | Companion change in PROPOSAL now lists the two ledger rows (R-040 as drafted, R-041 = X2 row added to the draft block) and the `docs/STATUS.md:162` edit; MOSS candidate "takes the next free ID at adoption"; judgement call 3 rewritten. The ledger and STATUS themselves are not edited (outside this directory). | PROPOSAL 64, 67–80 |
| 6 | minor | §4 sentence replaced: `skipped (no model CLI)` where `codex` is not on `PATH` (`modelcli.go:41-42`); installed-but-failing call keeps the heuristic map and reports exploration incomplete (`client.go:230-237`). PROPOSAL P4 mirrored. | design 107; PROPOSAL 41 |
| 7 | minor | Clause replaced with the council's text; mirrored in PROPOSAL P5 (DRAFT-NOTES line list rewritten). | design 101; PROPOSAL 42 |
| 8 | minor | Tag replaced with "**[ASSUMPTION; X2 did not confirm 0.5 for its model, A1; X2b pending]**". | design 102 |
| 9 | minor | R15 added after R14 in §12 (council's row text). | design 406 |
| 10 | nit | "model option `default`, probe-served `gpt-6.1-sol`" at lines 11, 107, 430 (and §13 X2 cell). | design 11, 107, 415, 430 |
| 11 | nit | Line 11 "so no provider value is refused for mapping"; Appendix 20 "never `claude` and with no provider fallback (the heuristic-map fallback on agent failure, §4, is unchanged)". | design 11, 473 |
| 12 | nit | argv `:219-220` at lines 11 and 107; PROPOSAL P2 also corrected from `:220`. | design 11, 107; PROPOSAL 39 |
| 13 | nit | CP07 churn figure scoped: "(codex, model option `default`; the churn figure is model-scoped, the need for reconciliation is not)". | design 102 |
| 14 | nit | PROPOSAL Fact 2 sentence replaced with the council's (`review.go:169`; `contextual/client.go:253`). | PROPOSAL 10 |
| 15 | nit | Fact 5 and ledger note rewritten against the X2b draft (DRAFT rev 1, not signed off, HD5b draft; enters the hypothesis ledger as `proposed` when the draft is reviewed). | PROPOSAL 14, 76 |
| 16 | nit | Table row split into the §3.1 tree row (65) and the §5.1 row 9 (124). | PROPOSAL 22–23 |
| 17 | nit | `proposed.diff` regenerated (120 lines, +18 / −13, `cmp` OK); all digests recomputed; changed-line, offset, hunk and unchanged-on-purpose lists rewritten; PROPOSAL P12 added for the new §4/§12/§13 changes; status labels set to round-2 draft. | DRAFT-NOTES; PROPOSAL 3 |

Verification by command after the fixes: `docs/scout-design.md` digest equals the base digest (`5c16a807…`); `requirements.proposed.diff` unchanged (`4cfdd24f…`); line counts base 468, proposed 473 (+5: paragraph and blank at 11–12, R15, X2b row, new Appendix 20).

Not changed (out of scope per the council's Observations): pre-existing "logged in" wording in §3.4, §6, R13, Appendix 8 and REQ-31; X2b draft defects (for X2b's own review).
