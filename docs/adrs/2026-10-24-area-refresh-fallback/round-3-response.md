# Round 3 response — revision 2.6 (agent map at onboarding and manual `scout map` only)

All 4 nits of `round-3.md` are accepted and fixed; the council's replacement text was used as given. `scout-design.proposed.md` stays at 477 lines (three in-line edits). Not signed off: round 4 re-reviews.

| # | Sev | Fix | Where |
|---|---|---|---|
| 1 | nit | R15 cite replaced by `Refresher.refresh` symbol cite and the `incomplete`-returns-before-reconciliation gloss; Phase 4 cite replaced by `reconcileAreas` / `next, _ := areas.Reconcile(…)`; DRAFT-NOTES item 9 replaced whole. `client.go:203-213` and `scout-nightly.yml:52` left as they are. | design 378, 408; DRAFT-NOTES |
| 2 | nit | Line 13 closing sentence now enumerates "the closing note on the Revision 2.5 paragraph that X2b has since run". | design 13 |
| 3 | nit | Status line, Fact 4, the abbreviations sentence and the (b) heading replaced with the council's text (pre-change tree / commit `f7b480e`; companion change in progress, commit after adoption). | PROPOSAL |
| 4 | nit | Companion (a) STATUS list gains lines 116 and 118 and the §5 "Remaining work" row in place of the "next-steps entry". | PROPOSAL |

Also updated: PROPOSAL status line (round-4 draft); `requirements.proposed.diff` line 3 recomputed (design digest, `round-4.md`; REQ-13 hunk unchanged; hunk sizes unchanged); DRAFT-NOTES digests, line 13/378/408 bullets and the requirements row.

## Digests (recomputed, `shasum -a 256`)

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` (= `docs/scout-design.md`, rev 2.5) | 473 | `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9` |
| `scout-design.proposed.md` | 477 | `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f` |
| `proposed.diff` (27 insertions, 23 deletions, 14 hunks; hunk headers unchanged) | 159 | `fc69e6bf8e49f90123bd7203f07fa9bfa714f6e61895c7168a9e2ebd278c2c8d` |
| `SCOUT_REBUILD_REQUIREMENTS.md` (base, unchanged) | | `9efcea1b0248d43d0ab831f0924f055f16bb39a3e008365d1e64c7d68a6ece2d` |
| `requirements.proposed.diff` (patched result `31d8c87dbcaa162928fd15210dd3030e41902ea81c2c59f82c9bdd6390875ac8`) | 20 | `9c39e6440a639db6ce295c10b6d63eff99616dd68033d23171f49caa7058c62e` |
| `PROPOSAL.md` | 179 | `97636b4cdd3788314eb1917c60f731af518bd72c9b6cd77d91e034ed0b4c6f67` |
| `DRAFT-NOTES.md` | 63 | `26c97515ae1556531c55cd0b240ecf31413fe42940be0a5b07386329ee48445d` |

Verification by command: `/usr/bin/diff -u --label … scout-design.base.md scout-design.proposed.md | cmp - proposed.diff` matched; `/usr/bin/patch -p1 --dry-run -i requirements.proposed.diff` from the repository root applied to `SCOUT_REBUILD_REQUIREMENTS.md` (digest `9efcea1b…`, unchanged); the patched result digest was computed with `patch -o` to a temp file (deleted).

Not changed: `scout-design.base.md`, `round-*.md` council files, `docs/scout-design.md`, `SCOUT_REBUILD_REQUIREMENTS.md`, the ledgers, `docs/STATUS.md`, code, `.codegraph`.
