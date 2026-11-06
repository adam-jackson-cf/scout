# Round 2 response — Cutover cleanup (revision 2.7)

Round 2 (`round-2.md`): 13 issues (1 blocker, 2 major, 7 minor, 3 nits). Every issue is fixed with the council's replacement text. The files edited are in this directory only (`scout-design.proposed.md`, `PROPOSAL.md`, `DRAFT-NOTES.md`, `proposed.diff`, `requirements.proposed.diff`); `scout-design.base.md`, the council rounds, code, ledgers, `docs/STATUS.md` and `docs/scout-design.md` are untouched. Code claims are cited by symbol.

## Per-issue

| # | Sev. | Fix | Where |
|---|---|---|---|
| 1 | blocker | §7.2: lane health is stated as a property of a published manifest first; the comparability sentence now requires only matching digests and the merge-base revision | design line 190; PROPOSAL D5 row |
| 2 | major | §7.2 parenthetical and export sentence use the three-way rule (latest finalized run, else a recorded imported base with non-null `uninspected`, else empty identifiers with `uninspected: []`; absent or `null` counts as none); Revision 2.7 (2) parenthetical; §10.2 row 352. PROPOSAL: recommendation (`baseInfo` re-emits a recorded `export_base` only when non-null), `export.go` row, the two new tests | design lines 15, 190, 352; PROPOSAL Decision 2, D5, D12 |
| 3 | major | Phase 1 guarantee restated (file not written, no `writer.lock`/`artifacts/`/journal-mode change, nothing migrated; `-wal`/`-shm` may appear, R-047); §3.3 and Revision 2.7 (1) "read on a read-only connection before any lock, journal-mode change or migration"; §10.2 row 351 "when the database file exists" and "any write to the database file"; R-047 added to the Revision 2.7 opening and the evidence notation. PROPOSAL: test (b), 1B consequence row, D8 row, `Store.Open` row ordering wording, the R-047 ledger row (Companion (a): four rows), STATUS and Process notes (R-044–R-047) | design lines 15, 17, 94, 351, 368; PROPOSAL |
| 4 | minor | Schema comparison covers indexed columns with sort order and collation, `sqlite_schema` objects by type and name, `index_xinfo` | design line 368; PROPOSAL test (a), D8 row, judgement call 11 |
| 5 | minor | §7.4 release wording ("the release that adds them, and every later one"; callers move workflow and `scout-version` together, the dogfood caller by its own revision); REQ-25 reworded | design line 205; `requirements.proposed.diff`; PROPOSAL D6 row, adopter consequence |
| 6 | minor | §7.4 interface reads `SCOUT_VERSION`, `RELEASE_REPO` and the toolchain directory from the environment | design line 205; PROPOSAL D6 row, 4E recommendation |
| 7 | minor | R17 release statement, Fact 2.3, consequence rows (152, 154), R-046 key point narrowed | design line 416; PROPOSAL Fact 2.3, persisted-state table, R-046 row |
| 8 | minor | §3.3 and the PROPOSAL D3 row: reasons recorded for every ambiguous row, `ambiguous_with` only for indistinguishable occurrences | design line 94; PROPOSAL D3 row |
| 9 | minor | "Explicitly not changed" lists corrected (`Store.Import` changes only when it records `export_base`; §7.3 attribution except the legacy-recompute clause; REQ-27 for a manifest that carries the field) | PROPOSAL |
| 10 | minor | Re-pointed test: `install-release.sh` (sourcing `sha256.sh`, calling `sha256`), `ci-toolchain.sh` (`bootstrap-tools.sh` from `$(dirname "$0")/..`, no `git clone`); the workflows assert the clone, the existence check and the two invocations | PROPOSAL test row |
| 11 | nit | "keeps only" clause lists the check that the clone holds both scripts | design line 205; PROPOSAL D6 row |
| 12 | nit | `reanchorLegacyIssue` renamed `reanchorIssue` in the inventory (persisted names kept) | PROPOSAL `issue_migration.go` row |
| 13 | nit | `proposed.diff` and `requirements.proposed.diff` regenerated (header line 3 names `round-3.md` and the new design digest); DRAFT-NOTES digests, hunk list, changed-line list and judgement-call count updated; Companion (c) names `round-3.md` | all |

## Sweep (same classes, beyond the council's list)

- Comparability/refusal contradiction class: PROPOSAL D5 row now describes the split sentence and the three-way export rule (it still carried the round-1 export wording); PROPOSAL D12 row (import/export) adds the recorded-`export_base` rule.
- WAL/`-wal`/`-shm` over-claim class: PROPOSAL D8 row (a third copy of the Phase 1 claim) and the 1A consequence row (read-only refusal table) fixed in the same words; the design §10.2 row 351 and Revision 2.7 (1) as listed.
- `index_list` class: the PROPOSAL D8 row and judgement call 11 (neither in the council's list) use the extended definition.
- Over-broad v0.1.2 statements: PROPOSAL D9 row ("(v0.1.2+) manifests with `uninspected`") and the Decision 2 recommendation ("written by v0.1.2+") narrowed.
- Release-naming class: PROPOSAL D6 row, the 4E recommendation (inputs now named beside outputs) and the Decision 4 adopter consequence ("releases that share the scripts") aligned with §7.4; the D6 row also gets the keeps-only and environment-input text.
- Ambiguity class: PROPOSAL D3 row had two statements (abbreviated and tail); both fixed.
- `legacy` naming class: no other `legacy` behaviour name was found outside persisted names and the 1C option's list of symbols to delete.
- Judgement call 12 added (the `Store.Open` guarantee is stated as the council's text; `immutable=1` not adopted; driver not run, **[INFERENCE]**); DRAFT-NOTES says "twelve calls".
- Not changed: Appendix 22's evidence list (R-045, R-046 only; the council text gave R-047 only for the header and the evidence notation); the design text's A13; the historical Revision 2.5 sentence.

## Judgement calls

1. D5 (PROPOSAL change-set row) gets a sentence recording the issue-1 restructure so the "Now" cell matches the design.
2. The council's P:146 insertion was placed verbatim after "`[]UninspectedFamily{}`; ", so the original "; `uninspected: []` then sits beside empty identifiers…" clause now follows the new `null` sentence; left as the council wrote it.
3. The requirements diff was regenerated by patching the repository file with the round-1 diff, editing the header digest, round name and REQ-25, and diffing again with `/usr/bin/diff` (labels `SCOUT_REBUILD_REQUIREMENTS.md` / `SCOUT_REBUILD_REQUIREMENTS.proposed.md`); hunks `-1,6`, `-18,7`, `-42,16` are unchanged.

## Digests (SHA-256; `shasum -a 256`; lines via `/usr/bin/grep -c ''`)

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` (rev 2.6, unchanged) | 477 | `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f` |
| `scout-design.proposed.md` | 486 | `8a0d7f7ad1709aec79e88ccbd63952642accb17faaaec599b69df790bfe0d3cd` |
| `proposed.diff` | 116 | `25f0e5648ae67e33619cb95b40de47aaec96de2e721dfc1bba4113ef485d135c` |
| `requirements.proposed.diff` | 40 | `4e4bbc42baa61c8a976f4b2d66f8623088c8f7e08a799eb3b65058b565451e81` |
| `SCOUT_REBUILD_REQUIREMENTS.md` base (unchanged) | n/a | `bf6b8d2944abb757dea13c36ee85ebf4159ebf5d9535ce31827357362f8f649b` |
| requirements patched result | n/a | `eb80e188b848814e27a2f6c63bc99ad0ed573a058864b72515cfce85dee02ae4` |
| `PROPOSAL.md` | 282 | `3fa3b3d3bee56278dbd8478a0592798c2c88e0a88e32f080b662ffaf207ff8d9` |
| `DRAFT-NOTES.md` | 60 | `128f9d96865f70f9bf79573a6dcbc90f1793acaf89a9883906603ecea915280d` |

Checks run: `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md | cmp - proposed.diff` equal; `proposed.diff` has 12 hunks, 19 insertions and 10 deletions (477 + 19 − 10 = 486) and patches the base to a file `cmp`-equal to the proposed design; `/usr/bin/patch -p1 --dry-run -i docs/design-reviews/2026-10-24-cutover-cleanup/requirements.proposed.diff` from the repository root applies cleanly. `round-2-response.md` itself is not digested here.
