# Cutover cleanup (revision 2.7) — response to round 5

`round-5.md` recorded 3 issues (0 blockers, 0 major, 1 minor, 2 nits). Each is fixed with the council's exact replacement text, in the council's order (design line 111, then `PROPOSAL.md` and `DRAFT-NOTES.md`, then regeneration of `proposed.diff` and `requirements.proposed.diff`, then digests). Only files in this directory were edited (not `scout-design.base.md`, any council `round-*.md`, code, ledgers, `docs/STATUS.md` or `docs/scout-design.md`). HEAD `5e1da48`.

| # | Fix | Where |
|---|---|---|
| 1 | Detection-order sentence: "selects the CLI when policy sets no provider (`SCOUT_MODEL_CLI` overrides it locally): `scout readiness` reports it only as the `model_cli` capability's ready state and reason, the capability matrix names it, the `contextual` prerequisite check resolves it and the plan's model fingerprint records it." | `scout-design.proposed.md` line 111 (single line; hunk list and line count unchanged); `PROPOSAL.md` D4 cell (78), 3C (177), 3D (178), 182; REQ-31 in `requirements.proposed.diff` line 37 (`+` line); `DRAFT-NOTES.md` lines 10, 28 |
| 2 | P:146 "`importRows` refuses a store that holds issues, so a store that imported a zero-issue manifest can take a second import of the same repository, …"; P:166 fixture "restore of a manifest with identifiers, `uninspected: []` and no issues, `would_block` rows, signals or areas, then import of a no-`uninspected` manifest of the same repository into the same store" | `PROPOSAL.md` lines 146, 166 (design unchanged by this issue) |
| 3 | Derived artifacts regenerated; status/bookkeeping moved to round 6 | `proposed.diff`, `requirements.proposed.diff` (header names `round-6.md` and the new design digest; REQ-31 as in issue 1); `DRAFT-NOTES.md` lines 3, 8–11, 17–20, 28, 40; `PROPOSAL.md` lines 3, 278 |

## Corrections to earlier records

- `round-3.md` issue 2's REQ-31 text, and the `round-3.md`/`round-4.md` §3 verifications of D:111 and RD:37, overstated what readiness reports: plain `scout readiness` reports only the `model_cli` capability's ready state and reason (`internal/engine/readiness.go:205-206`); the capability matrix names the CLI (issue 1).
- `round-4.md` issue 1 cited P:84 for the D12 cell; it is P:87 (the response edited P:87, correctly).
- Round-4 issue 1's P:146 parenthetical and P:166 fixture were imprecise (the `ensureRepositoryIdentity` check; shared IDs/signals across two manifests) (issue 2).
- `round-4-response.md` sweep claim 3's locator should read "REQ-31 (RD:37, hunk `-42,16`)" (issue 3).

## Sweep of the same classes (verified by reading grep hits, not asserted)

1. "readiness reports/prints the model CLI" class: grep of D, P, DN, RD for `readiness`/`matrix` with report/name/print/found. After the edits the hits are: D:111 (new sentence); P:78 (same sentence); P:177, 178, 182 (fixed); RD:37 and DN:10 (fixed). Left unchanged because true: D:98, D:376, D:412 and D:138 (`readiness --matrix` prints/reports the matrix and the skip); D:15, D:353, D:486(c), P:87, P:180 ("detection and reporting": the matrix reports, readiness detects); P:34 (Fact 2, the readiness `model_cli` capability); P:36 ("what is reported or fingerprinted"); P:176 ("matrix/readiness text" is unchanged under 3B: `Detect` is untouched); P:38 and P:192/194/195 (docs references, "first CLI found on `PATH`"). No remaining "readiness … report `claude`/which CLI is found" text in D, P, DN or RD.
2. "importRows checks only issues" class: grep `importRows|zero-issue|second import` in D, P, DN: P:146 (fixed) and P:166 (fixed); D:352 states only the replace/delete rule and is unchanged; no other statement of the admission check.
3. Stale round names/digests: `5774189c`, `d89cb9c5`, `7b8b12bc`, `06101038`, `1dcae6aa` no longer occur in D, P, DN, RD (the only `round-5.md` hits are DN:19's historical account of the round-4 version and RD's context line for `area-refresh-fallback/round-5.md`); P:3, P:278, DN:3, DN:10 and RD:8 name round 6.
4. "rounds 2 to 4" / "round-1 to round-4" bookkeeping in DN (lines 3, 17, 18, 19, 20, 40) now reads round 5.

## Regeneration

- `proposed.diff`: `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md > proposed.diff`; a fresh run piped into `cmp - proposed.diff` matches; same 12 hunk headers, 116 lines; only the line-111 `+` line changed.
- `requirements.proposed.diff`: regenerated via `/tmp/reqdiff27r5` (patch the repository file with the round-4 diff, edit the design digest, `round-5.md` → `round-6.md` and REQ-31, `/usr/bin/diff -u` with the two labels); 40 lines, 3 hunks (`-1,6`, `-18,7`, `-42,16`), +6/−6; `/usr/bin/patch -p1 --dry-run -i docs/design-reviews/2026-10-24-cutover-cleanup/requirements.proposed.diff` from the repository root applies cleanly; the repository file stays `bf6b8d29…`. The patched result digest changed to `7016ae5c0a0b8ddf600d944cc36754958ca8eac155da1a1ab2bd626242f16337`.

## Digests (SHA-256) and lines

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` | 477 | `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f` |
| `scout-design.proposed.md` | 486 | `bd4477538b52c60fa7f29c10bcd533d4c790144c86af0432cacd7cab18d2b2d1` |
| `proposed.diff` | 116 | `b2f52886dc409c19de62b9af87fba7074ae2449a36493e809075f8c65736183c` |
| `requirements.proposed.diff` | 40 | `e17cfecb3812b01dcc1de8db08913ef61ab39679f916479ca877d072290a85bc` (patched result `7016ae5c0a0b8ddf600d944cc36754958ca8eac155da1a1ab2bd626242f16337`; base `SCOUT_REBUILD_REQUIREMENTS.md` `bf6b8d2944abb757dea13c36ee85ebf4159ebf5d9535ce31827357362f8f649b`) |
| `PROPOSAL.md` | 282 | `b5b4fcb85e043d5d4a69300105dbd0ae31b7199cbef7e5c719ca6c564066ca7f` |
| `DRAFT-NOTES.md` | 60 | `330b75dcbfacf41a4409495e7f5e17a17216754ba487245f9c8527622740f3be` |

`DRAFT-NOTES.md` records the first five digests and cannot contain its own; its digest is recorded here.

## Judgement calls

1. The edits were applied as exact-substring replacements (each asserted unique) in an eval cell, as in earlier rounds, because the target lines are over 1,800 characters.
2. DN:19 keeps the round-4 account's "(`round-5.md`)" as history and the round-5 account is prepended in the same form, naming `round-6.md`.
3. P:176 ("matrix/readiness text") and P:180 ("detection and reporting (readiness, the matrix, …)") were not in the council's list and were left: readiness text is unchanged under 3B, and readiness does detect.
