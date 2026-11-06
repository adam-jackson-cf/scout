# Round 3 response — revision 2.5

All 12 issues fixed as specified, except issue 1 (fixed as directed by the main agent, below). Digests and the `cmp` check are in the final consistency pass.

| # | Fix |
|---|---|
| 1 | (a) `PROPOSAL.md` R-040/R-041 rows (and the HD5 History line in Companion (c)) are now the single canonical ledger text. R-040 Informed: "Design rev 2.5 header, §4 Refresh triggers, Areas and Model CLI, §6, §10.2–10.3, §12 R15, §13 X2 and X2b, §14 A1, Appendix 8 and 20; `SCOUT_REBUILD_REQUIREMENTS.md` REQ-31; X2b preregistration draft (`evaluation/x2b-area-stability-luna/`) and its hypothesis HD5b (X2b's ledger step)"; R-041 Informed and Caveat as the council wrote them. The ledger note (`PROPOSAL.md:76`) and the HD5b sentence replaced; Companion (a) and (b) as required. Departure from the council's wording, as the main agent directed: the ledger note states that the rows are appended in the rev 2.5 adoption change set, before X2b's final round, and that X2b's change set adds neither, instead of "whichever change set lands first". `evaluation/x2b-area-stability-luna/DRAFT-NOTES.md` section "Ledger text for the main agent to append before council round 2" now points to the PROPOSAL rows and keeps only HD5b's draft row (nothing else in that directory touched). `DRAFT-NOTES.md` judgement call 3 records the shared rows. |
| 2 | Design :416 "decides on the resulting proposal —"; :473 "decided by a later council design round."; P11 "decided by". |
| 3 | A1 :430 "fastapi (Python) and axum (Rust) only (no Go or TypeScript repository; no cross-repository claim; …" and "labels, unreviewed by a human, under"; P10 "fastapi (Python) and axum (Rust) only". |
| 4 | R15 :406 "never `refreshed` (`internal/engine/refresh.go:35-37,153-155`)". |
| 5 | R15 :406 "evidence only for whatever codex served as `gpt-6-luna` at `xhigh` during X2b's collection (X2b's probes echo the requested name and cannot see a same-name change)". |
| 6 | A1 "(`evaluation/x2-area-stability/RESULTS.md` §4.3, operating characteristics)"; R15 "`evaluation/x2-area-stability/RESULTS.md` §10.2". |
| 7 | :13 "signed off and run under Scout's experiment rules". |
| 8 | Companion (c) "adds at the top of `docs/hypothesis-ledger.md` History (newest first):". |
| 9 | HD5 History line replaced verbatim with the council's text (effort not recorded; X2b tests the pinned mapper under HD5b and does not move HD5's state). |
| 10 | Fact 6 replaced with the council's text (neither recorded a served model; neither is evidence about `gpt-6-luna` at `xhigh`). |
| 11 | `round-2-response.md:22` last clause replaced as specified. |
| 12 | `proposed.diff` regenerated with the documented `/usr/bin/diff -u --label …` command and confirmed with `cmp` (hunk list unchanged, 120 lines, 18 insertions, 13 deletions); every digest in `DRAFT-NOTES.md` recomputed with `shasum -a 256`; annotations for lines 13, 406, 416, 430, 473 and judgement call 3 updated; status lines changed to round-4 draft. `DRAFT-NOTES.md` cannot hold its own digest; it is reported by the main agent. |

`requirements.proposed.diff` unchanged (`4cfdd24f…`). No design, ledger or code file outside this directory and the one X2b DRAFT-NOTES section was edited.
