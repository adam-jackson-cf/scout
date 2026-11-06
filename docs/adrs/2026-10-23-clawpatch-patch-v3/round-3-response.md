# Response to round 3 — revision 2.4 (round-4 draft)

All three nits fixed with the council's exact wording.

| # | Fix | Location |
|---|---|---|
| 1 | Third sentence of the Revision 2.4 paragraph replaced by "Two code facts stale in revision 2.3 are also refreshed (status-line dogfood pin `v0.1.4`; §4 `defaultMapTimeoutMS`, `internal/clawpatch/client.go:23`); apart from revision bookkeeping (status line, evidence notation), no other statement changes." | `scout-design.proposed.md:9` |
| 2 | P5 "Where" = "Status line 3 (stale in revision 2.3, not from `06c9154`)" | `PROPOSAL.md:20` |
| 3 | "All nine nits fixed; eight as the council's required-fix column prescribes, issue 6 with the label advanced to round 3 (row 6)." | `round-2-response.md:3` |

Bookkeeping: `proposed.diff` regenerated with `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md` (still 45 lines, same hunks; only proposed line 9 changed); PROPOSAL and DRAFT-NOTES status advanced to the round-4 draft; DRAFT-NOTES digests updated. Line-number condition from round 3: `round-1-response.md` gained no line after round 2 (saved state compared: the round-1 fixer's write and the round-2 fixer's pre-edit read in their session transcripts both have the blank at `:17` and the `Note:` paragraph at `:18`); `round-2.md`'s `:17`/`:19` cites are off by one and read `:18`/`:20`.

| File | SHA-256 |
|---|---|
| `scout-design.base.md` | `75ef1e66521b1f129e872f651bd0f8bb9989e38867fb5335ec5ad43158fb3061` |
| `scout-design.proposed.md` | `5c16a807513d54667390278755b8daa504e8db5ae0733236f87938c51a9a90ae` |
| `proposed.diff` | `101f0af4fd1d48cdb70c2f1139071768bd77fa12be283c09f52648e1b00f8759` |
| `PROPOSAL.md` | `ad985bc71b315bf21de19316fae89bd3b793008fc6617228216331985f271b12` |
| `DRAFT-NOTES.md` | `d8ab52786bf40f3056a2b9a3b927be1909e9eb234759255c7ef03116ab47dd9a` |
