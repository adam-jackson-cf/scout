# Round 2 response — revision 2.4 (clawpatch patch v3)

All nine nits fixed; eight as the council's required-fix column prescribes, issue 6 with the label advanced to round 3 (row 6).

| # | Fix | Where |
|---|---|---|
| 1 | Deleted " only"; closing clause now "apart from revision bookkeeping (status line, evidence notation), no other statement changes."; mirrored in PROPOSAL P2 | proposed:9; PROPOSAL:17 |
| 2 | `client.go:23` → `internal/clawpatch/client.go:23` | proposed:9 |
| 3 | "base 97" / "Line 99 … (base 97)" / "proposed:99" | PROPOSAL:21; DRAFT-NOTES:23; round-1-response:11 |
| 4 | "7 insertions (6 text, 1 blank), 5 deletions"; "context shifts" parenthetical dropped | DRAFT-NOTES:9, 12 |
| 5 | Judgement call 2 now "…beyond the pin, the v3 change and the two drift refreshes (P5, P6)" | DRAFT-NOTES:32 |
| 6 | PROPOSAL status now "round-3 draft (round-2 issues fixed, `round-2-response.md`), awaiting innovation-council review" (updated from the prescribed "round-2 draft" because this is the round-3 submission); DRAFT-NOTES:3 uses the same label | PROPOSAL:3; DRAFT-NOTES:3 |
| 7 | Shift sentence: "Every base line from the evidence-notation line (base line 9) onward shifts by +2 (the new paragraph and its blank line at proposed 9–10)." | DRAFT-NOTES:27 |
| 8 | "…copies `CLAUDE_CODE_OAUTH_TOKEN` in host context when auth is included (`third_party/clawpatch/src/providers/claude.ts:258,268-270`)" | PROPOSAL:28 |
| 9 | P5 "Where" = "Status line 3 (drift since revision 2.3, not from `06c9154`)"; evidence adds `docs/STATUS.md:72,94` | PROPOSAL:20 |

## Consistency pass (each verified by command)

- `proposed.diff` regenerated with the `DRAFT-NOTES.md:14` command; `diff … | cmp - proposed.diff` identical. 45 lines; hunks `-1,12 +1,14`, `-94,14 +96,14`, `-451,7 +453,7`; 5 `-`, 7 `+` (6 text, 1 blank, at diff line 15).
- Line counts: base 466, proposed 468. Line map: §4 refresh base 97 → 99; pin base 104 → 106; Appendix 7 base 454 → 456; evidence notation base 9 → 11; Revision 2.4 paragraph at proposed 9, blank at 10.
- Correction found in the pass: `scout-nightly.yml` carries `scout-version: v0.1.4` at `:40,50` on `main` (`git show HEAD:`), `:40,49` in the working tree (uncommitted edits). PROPOSAL P5 evidence now states both. `git show HEAD:.github/workflows/release.yml | sed -n 89p` = `"safetyPatchSha256": "dc867503…e6ce3"`, so `:106`'s `release.yml:89` holds.
- Code cites re-read: `client.go:20-21` (`Revision`, `Version`), `:23` `defaultMapTimeoutMS = "900000"`; `review.go:20-21` (v3, `dc867503…`); `clawpatch-safe.patch:332-336`; `docs/STATUS.md:72,94`.
- `round-1-response.md:20` digests are marked as round-1 values superseded by this file; its line 10 is the round-1 record of the then-text and is left as written.
- `docs/scout-design.md` and `scout-design.base.md` both `75ef1e66…3061`.

## New digests (SHA-256, `shasum -a 256`)

- `scout-design.base.md` = `docs/scout-design.md`: `75ef1e66521b1f129e872f651bd0f8bb9989e38867fb5335ec5ad43158fb3061`
- `scout-design.proposed.md` (468 lines): `cdc97065079f034e6acd3d7df7d44f7b068408dbc7ba71f33048221fd18d6709`
- `proposed.diff` (45 lines): `0a9bd12aa4ab5094f4d166c4d978ac5c8860a0f4d4f590a04d81c36bdc8e4db0`
- `PROPOSAL.md`: `5831bf55385031a7dfdd348bf85c3e575ee7b9fa8fd9e3a99a86e740cc16f268`
- `DRAFT-NOTES.md`: `bb93273e29d940a4f3b689378abac8c40772695366fa6eb0463266394e6b9f41`

Requesting round 3.
