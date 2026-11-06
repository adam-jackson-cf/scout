# Cutover cleanup (revision 2.7) — response to round 3

`round-3.md` recorded 8 issues (0 blockers, 0 major, 4 minor, 4 nits). Each is fixed with the council's exact replacement text, in the council's order (design line 15, the requirements diff, `PROPOSAL.md` and `DRAFT-NOTES.md`, then regeneration). Only files in this directory were edited (not `scout-design.base.md`, any council `round-*.md`, code, ledgers, `docs/STATUS.md` or `docs/scout-design.md`). HEAD `5e1da48`.

| # | Fix | Where |
|---|---|---|
| 1 | Revision 2.7 paragraph now reads "Appendix 22 records the decision; §10.2 gains one row per decision and §12 adds R17, the refusal cost of (1) and (2). The shipped code and workflows" | `scout-design.proposed.md` line 15 (single line; hunk list and line count unchanged) |
| 2 | REQ-31: "with none set and no local `SCOUT_MODEL_CLI` override, readiness and the capability matrix report the first of `claude`, `codex` found" | `requirements.proposed.diff`, hunk `-42,16` |
| 3 | `install-release.sh` / `ci-toolchain.sh` cell rewritten (input checks stay in the workflow; installs into `$RUNNER_TEMP/scout-release`; `SCOUT_BIN`/`CLAWPATCH_BIN` exported by `install-release.sh`; bubblewrap, toolchain directory, `SCOUT_TOOLS_DIR` and `bin` on `$GITHUB_PATH` by `ci-toolchain.sh`); workflow row says the step keeps `TOOLCHAIN_DIR` and `RELEASE_REPO` in its `env` | `PROPOSAL.md` Decision 4 inventory rows |
| 4 | Canonical "branch tip was published by v0.1.0/v0.1.1, or by an earlier build from a store with no finalized run and no imported lane health" wording | `PROPOSAL.md` option 2B "Against" and consequence row; `DRAFT-NOTES.md` judgement call 3 |
| 5 | The "never beside identifiers…" clause replaced with the council's sentence (empty identifiers only without a finalized run or a recorded non-null `export_base`) | `PROPOSAL.md` Decision 2 recommendation |
| 6 | Full schema-equality definition in judgement call 11 and in the D8 "Now" cell | `PROPOSAL.md` |
| 7 | `immutable=1` rationale replaced (asserts the file cannot change, skips locking and change detection); R-047 row title, source, "Used in" and key-point cells extended | `PROPOSAL.md` judgement call 12 and R-047 row |
| 8 | Derived artifacts regenerated and bookkeeping updated (below) | `proposed.diff`, `requirements.proposed.diff`, `DRAFT-NOTES.md`, `PROPOSAL.md` |

## Sweep of the same classes (each fixed consistently)

1. Change record naming every touched section: `PROPOSAL.md` D2 row now says the paragraph names "the §10.2 rows and §12 R17 they add"; `DRAFT-NOTES.md` line-15 entry appends "; round 3: §10.2 rows and R17 named in the paragraph (issue 1)". The Revision 2.2–2.6 paragraphs were not changed (council verified D:13).
2. REQ-31 override omission: no other statement of the first-of-`claude`/`codex` rule omits `SCOUT_MODEL_CLI` (design line 111 and `PROPOSAL.md` Decision 3 already carry it); `DRAFT-NOTES.md` requirements row and `PROPOSAL.md` Companion (c) now name REQ-31's override clause.
3. Script inputs/outputs: `PROPOSAL.md` Decision 4 summary rows (P:81, P:211) and design line 205 attribute exports to "the scripts" jointly, which is correct; no other `DEST` mention exists.
4. Lane-health wording: all other statements (design lines 15, 190, 416; P:26, P:84, P:263; DN:29) already use the canonical wording. P:28's "in any release" describes pre-cutover behaviour and was left.
5–7. No further abbreviated schema definitions, "never beside" clauses or `immutable` statements exist.

## Regeneration

- `proposed.diff`: `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md …` regenerated; `cmp` against a fresh run matches; 12 hunks, +19/−10, 116 lines; `patch -s -o` on the base is `cmp`-equal to the proposed design (486 lines).
- `requirements.proposed.diff`: regenerated from the repository file via `/tmp/reqdiff27r3` (patch with the round-2 diff, edit digest, `round-4.md` and REQ-31, diff); 40 lines, 3 hunks, +6/−6; `/usr/bin/patch -p1 --dry-run -i …` from the repository root applies cleanly; `patch -o` is `cmp`-equal to the edited copy.

## Digests (SHA-256) and lines

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` | 477 | `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f` |
| `scout-design.proposed.md` | 486 | `5d58b242df7486bebbe6805fc31383d9768383510389649391819a3a9431e87e` |
| `proposed.diff` | 116 | `ae5a1448bd0e0d3c8b10bf093085eb6a7fe4e3b7d27cbcda3c0cff5fabb41f46` |
| `requirements.proposed.diff` | 40 | `9a3bcfc0fa4a8e593ea93a3c1b7a8eb31e2399315a830c3d471560bfad19d916` (patched result `b529fa47d6ec917b06c4e28e041ea92a9770507811b09f4e9d5e24b607c0175a`) |
| `PROPOSAL.md` | 282 | `e8b63d811cc558aec36911a269b089de27791b173cdd5dcd5edf0934c296fcb4` |

`DRAFT-NOTES.md` records these digests and cannot contain its own.

## Judgement calls

1. `requirements.proposed.diff` was regenerated rather than text-patched, as instructed; the result differs from the round-3 diff only in header line 3 (digest, `round-4.md`) and the REQ-31 `+` line.
2. Issue 8 asks for P's Companion (c) to "name REQ-31's override clause": added as "REQ-10, 25, 27 and 31, whose model-CLI clause keeps the local `SCOUT_MODEL_CLI` override".
3. Companion (c) also records "round 3 recorded 8" in the issue tally alongside the `round-4.md` rename.
