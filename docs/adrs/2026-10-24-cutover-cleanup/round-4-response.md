# Cutover cleanup (revision 2.7) — response to round 4

`round-4.md` recorded 6 issues (0 blockers, 0 major, 3 minor, 3 nits). Each is fixed with the council's exact replacement text, in the council's order (design line 352, then `PROPOSAL.md` and `DRAFT-NOTES.md`, then regeneration of `proposed.diff` and `requirements.proposed.diff`, then digests). Only files in this directory were edited (not `scout-design.base.md`, any council `round-*.md`, code, ledgers, `docs/STATUS.md` or `docs/scout-design.md`). HEAD `5e1da48`.

| # | Fix | Where |
|---|---|---|
| 1 | `Store.Import` "replaces `export_base` with the manifest's metadata when the manifest carries `uninspected` and otherwise deletes any recorded `export_base`, in the import transaction" | `scout-design.proposed.md` line 352 (single line; hunk list and line count unchanged); `PROPOSAL.md` D12 cell (line 87), Decision 2 recommendation (146), code-impact row (164), new-tests row (166, the second-import scenario appended), "Explicitly not changed" (230); `DRAFT-NOTES.md` line 32 |
| 2 | `install-release.sh`: define `die` (`::error::`, exit 3) before sourcing `scripts/sha256.sh`, citing `sha256.sh:3,9`; test row says "defining `die`"; `scripts/` row adds the `sha256.sh:1` header-comment edit | `PROPOSAL.md` lines 219, 221, 222 |
| 3 | "no `model_cli.provider`, no local `SCOUT_MODEL_CLI` override and `CLAWPATCH_PROVIDER=codex` set" | `PROPOSAL.md` line 182 |
| 4 | "the only manifests that need it come from v0.1.0/v0.1.1 (published 2026-10-23) or from an earlier build's publish with no finalized run and no imported lane health, and none was observed (17 of 17 …)" | `PROPOSAL.md` line 142 |
| 5 | "The round-1 to round-4 fixes were applied the same way … (`round-1-response.md` to `round-4-response.md`)"; "The latest round response (`round-<N>-response.md`) lists the current digests and the sweep." | `DRAFT-NOTES.md` lines 20, 60 |
| 6 | Derived artifacts regenerated; status/bookkeeping lines updated to round 5; `DRAFT-NOTES.md` digest row added below | `proposed.diff`, `requirements.proposed.diff`, `DRAFT-NOTES.md` lines 3, 10; `PROPOSAL.md` lines 3, 278 |

## Corrections to earlier records

- `round-3-response.md` sweep claim 2 (no other statement omits `SCOUT_MODEL_CLI`) was incomplete: P:182 omitted it (issue 3).
- `round-3-response.md` sweep claim 4 (all other statements use the canonical lane-health wording) was incomplete: P:142 named only the v0.1.0/v0.1.1 publishes by age (issue 4).
- `round-3.md` §4's "not adopted: define `die`" rested on not reading `scripts/sha256.sh:3` (issue 2).

## Sweep of the same classes (verified by reading the grep hits, not asserted)

1. Import rule: every statement of what `Store.Import` does with `export_base` in the design and `PROPOSAL.md` now says replace/delete or "records or deletes": design 352 (grep `export_base` in the design hits only 352); P:87, 146, 164, 166, 230. P:28 (Fact 5) describes HEAD's `Import` (upsert whenever revision or digests are set) and is left; P:153/154 describe observable results (imports records, no comparison identifiers; recorded `null` counts as none) and agree. No "records `export_base` only"/"writes `export_base` only"/"only when the manifest carries" text remains in D, P, DN or RD.
2. `sha256.sh` sourcing: P:208 (option 4E summary, "sourcing `scripts/sha256.sh` rather than a second copy") makes no claim about `die` and is left; P:219, 221, 222 carry the fix. No other "source/sourcing `scripts/sha256`" text remains.
3. `SCOUT_MODEL_CLI` override: D:111, REQ-31 (RD:8's hunk `-42,16`), P:78 and P:182 carry it; P:36 states it for reporting; P:38 and P:192 do not describe the default CLI. No statement of "readiness/matrix report `claude`" omits it.
4. Lane-health origin wording: D:15, 190, 416; P:26, 84, 142, 143, 153; DN judgement call 3 name both origins (v0.1.0/v0.1.1 and an earlier build's publish with no finalized run and no imported lane health). "2 days old" no longer occurs.
5. Stale round names/digests: grep for `5d58b242`, `9a3bcfc0`, `ae5a1448`, `b529fa47`, `e8b63d81`, `313622e1` in D, P, DN, RD returns no hit after the edits; the only occurrences of `round-4.md` as a prospective sign-off round (DN:10, P:278, RD:8) now read `round-5.md`.

## Regeneration

- `proposed.diff`: `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md > proposed.diff`; a fresh run piped into `cmp - proposed.diff` matches; same 12 hunk headers, 19 insertions (18 text, 1 blank) and 10 deletions, 116 lines; only the line-352 `+` line changed.
- `requirements.proposed.diff`: regenerated via `/tmp/reqdiff27r4` (patch the repository file with the round-3 diff, edit the design digest and `round-4.md` → `round-5.md`, `/usr/bin/diff -u` with the two labels); 40 lines, 3 hunks (`-1,6`, `-18,7`, `-42,16`), +6/−6; the only difference from the round-3 diff is header line 3 (digest and round name); `/usr/bin/patch -p1 --dry-run -i docs/design-reviews/2026-10-24-cutover-cleanup/requirements.proposed.diff` from the repository root applies cleanly; the repository file stays `bf6b8d29…`.

## Digests (SHA-256) and lines

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` | 477 | `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f` |
| `scout-design.proposed.md` | 486 | `5774189c9fd644d36959477ca7bee1a2371f65a01ba2334a4807b91bafd13d5f` |
| `proposed.diff` | 116 | `7b8b12bc55022b65f70be9b0292066b37be30c7ba1c8926c7617bb821427402c` |
| `requirements.proposed.diff` | 40 | `d89cb9c52ef67406267fe2e81c4e223f367d6c63c440db5b5c38f5a5ebb90ef9` (patched result `1dcae6aa8972b85c7f53d60b757440d5f634f621e383ecd2bb98c05e65cdffc5`; base `SCOUT_REBUILD_REQUIREMENTS.md` `bf6b8d2944abb757dea13c36ee85ebf4159ebf5d9535ce31827357362f8f649b`) |
| `PROPOSAL.md` | 282 | `061010387fcd647c69ef6af20092796086a6c33b12ea902ce4a747a0b69924f1` |
| `DRAFT-NOTES.md` | 60 | `07d0f3cd9ec8ea8ac10dc5ba95f683b1bd3e1f2b5f348517beeda358b91c3d9d` |

`DRAFT-NOTES.md` records the first five digests and cannot contain its own; its digest is recorded here.

## Judgement calls

1. DN:32: the council said "append"; the line ends "(D7).", so the round-4 clause is inserted inside the row-351–354 parenthetical ("…issues 2, 3; round 4: row 352 deletes a recorded `export_base` on an import without the field, issue 1); line 356: …"), same words, so the line does not end in "(D7); round 4: …".
2. DN prose that listed re-run rounds ("re-checked after the round-1, round-2 and round-3 fixes", "Re-run after the round-2 and round-3 fixes", "unchanged by rounds 2 and 3") was updated to include round 4 (line 352 reworded only); not in the council's list, done for consistency.
3. P:208 (option 4E summary) was not changed to mention `die`: it describes the sourcing, not the contract.
