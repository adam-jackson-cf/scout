# Round 4 response — revision 2.5

All 4 issues fixed as specified (issue 1 with the main agent's fixed sequence stated: rev 2.5 is adopted, appending R-040, R-041 and the HD5 History line, before X2b council round 2; HD5b is appended only after that). Digests and the `cmp` check are in the final consistency pass.

| # | Fix |
|---|---|
| 1 | `PROPOSAL.md:76`: "keeps only HD5b's draft row and History line, which belong to X2b"; "in the rev 2.5 adoption change set; X2b's change set adds neither row, and X2b's HD5b row, which cites both, is added only after they exist (`.agents/skills/hypothesis-ledger/SKILL.md`: every cited source must already exist in the research ledger), so this adoption precedes HD5b's row and X2b's final round." `:80` (a): "(once, before X2b's HD5b hypothesis-ledger row, which cites both, and before X2b's final round; Ledger ID note)". `DRAFT-NOTES.md:43`: the council's replacement (rows appended and History line added at the top in the rev 2.5 adoption, which precedes HD5b's row and X2b's final round). X2b `DRAFT-NOTES.md` ledger section: the council's HD5b sentence ("if X2b's round 2 comes before that adoption, the row waits for it"); the paragraph sentence now says the rows are appended "before the HD5b row below and before X2b's final round"; the stale "the ledger rows above appended" in "Not done" now reads "the R-040/R-041 rows appended by the rev 2.5 adoption, then the HD5b row above appended". |
| 2 | `DRAFT-NOTES.md:43`: "no ledger, `docs/STATUS.md` or `docs/scout-design.md` edit is made now; the one edit outside this directory, by the round-3 response, is the ledger section of `evaluation/x2b-area-stability-luna/DRAFT-NOTES.md`." |
| 3 | `scout-design.proposed.md:473`: "fixed in X2b's preregistration before X2b runs; a later council design round decides on the resulting proposal." P11 (`PROPOSAL.md:48`): "combination fixed in X2b's preregistration; a later council design round decides on the resulting proposal)". `DRAFT-NOTES.md:33`: "round 3 issue 2 and round 4 issue 3: "a later council design round decides on the resulting proposal"". |
| 4 | `proposed.diff` regenerated with `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md > proposed.diff` and confirmed with `cmp` (OK; 120 lines, hunks unchanged); digests recomputed with `shasum -a 256`; status lines set to round-5 draft. |

## Digests (SHA-256)

- `scout-design.base.md` and `docs/scout-design.md`: `5c16a807513d54667390278755b8daa504e8db5ae0733236f87938c51a9a90ae`
- `scout-design.proposed.md` (473 lines): `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9`
- `proposed.diff`: `778842853667a5a882b261ca8875e4a47c7697e72cebcf02eaf2f9bc749b0fda`
- `requirements.proposed.diff` (unchanged): `4cfdd24fb96bec040bc4112b79b5b80e75c92bc953bef22876b60f742a2b5463`
- `PROPOSAL.md`: `86d5685f4ac4b7356b22cf83d1d06d77fea5365f2f954df376847a8d3a76021f`
- `DRAFT-NOTES.md`: not self-hostable; reported by the main agent.

No `docs/scout-design.md`, ledger or code file was edited; outside this directory only the X2b `DRAFT-NOTES.md` ledger section and its first "Not done" bullet changed.
