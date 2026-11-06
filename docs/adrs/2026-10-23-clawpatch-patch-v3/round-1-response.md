# Round 1 response — revision 2.4 (clawpatch patch v3)

Choices: issues 4–5 folded into rev 2.4 (council-recommended option; one clause in the Revision 2.4 paragraph). Issue 3: the smoke-evidence clause is dropped from the design (kept in PROPOSAL fact 2, qualified as an observation).

| # | Fix | Where |
|---|---|---|
| 1 | Cite `internal/clawpatch/client.go:20-21` | proposed:106; PROPOSAL P3 "Now" |
| 2 | Added `toolchain/versions.env:10` to the pnpm parenthetical | proposed:106; PROPOSAL P3 |
| 3 | Deleted "without them Claude Code on macOS fails every call" from the design; PROPOSAL fact 2 now "Claude Code 2.1.309 on macOS failed every call (observed 2026-10-23; not in the design)" | proposed:9; PROPOSAL:8 |
| 4 | Status line: "release `v0.1.2` exists for CI to install and the dogfood caller pins `v0.1.4`, `docs/STATUS.md` §1"; Revision 2.4 paragraph ends "Also refreshes two code facts that drifted since revision 2.3 (status-line dogfood pin `v0.1.4`; §4 `defaultMapTimeoutMS`, `client.go:23`); no other statement changes."; PROPOSAL rows P5–P6, P2 reworded; DRAFT-NOTES updated | proposed:3, 9; PROPOSAL; DRAFT-NOTES |
| 5 | `defaultMapTimeoutMS = "900000"`, `internal/clawpatch/client.go:23` | proposed:99 |
| 6 | "revision 2.1, 2.2, 2.3 or 2.4" (PROPOSAL row P7) | proposed:11 |
| 7 | "proposed line 9" | PROPOSAL P2 |
| 8 | Shift sentence replaced with the council wording | DRAFT-NOTES |
| 9 | Stated `first 8…last 5`, rev 2.3 line used last 4 | PROPOSAL |
| 10 | Host-only block wording per council, incl. `CLAUDE_CODE_OAUTH_TOKEN` | PROPOSAL "Explicitly not changed" |

Note: dogfood pin described as on `main` (`git show HEAD:.github/workflows/scout-nightly.yml`), not the working-tree edits.

New digests (SHA-256) at round 1 (superseded by `round-2-response.md`): `scout-design.proposed.md` `49935688a9247a2a3f454919ccd8a3e17acedf941f3ca3873afea22a2281720c` (468 lines); `proposed.diff` `f5b84bec91b63711ccefb0b0c564a382889d8c8a9951e1d9781c074ee1319663` (45 lines); `PROPOSAL.md` `010658c8a42c22a50357faad464fcb81daec81a345c7e05242bc8a39913e62bf`; base unchanged `75ef1e66521b1f129e872f651bd0f8bb9989e38867fb5335ec5ad43158fb3061` (equals `docs/scout-design.md`).
