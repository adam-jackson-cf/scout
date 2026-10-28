# Cutover cleanup (revision 2.7) — response to round 6

`round-6.md` recorded 3 issues (0 blockers, 0 major, 1 minor, 2 nits). Each is fixed with the council's exact replacement text, in the council's order (design line 111, then `PROPOSAL.md` and `DRAFT-NOTES.md`, then regeneration of `proposed.diff` and `requirements.proposed.diff`, then digests). Only files in this directory were edited (not `scout-design.base.md`, any council `round-*.md`, code, ledgers, `docs/STATUS.md` or `docs/scout-design.md`). HEAD `5e1da48`.

| # | Fix | Where |
|---|---|---|
| 1 | Detection-order sentence: "selects the CLI when policy sets no provider (`SCOUT_MODEL_CLI` overrides it locally): the `contextual` prerequisite check resolves it, `scout readiness` reports whether it was found and never its name (the `model_cli` capability's ready state and reason, and the `contextual_provider` capability's, which runs that check), the capability matrix names it and the plan's model fingerprint records it." | `scout-design.proposed.md` line 111 (single line; hunk list and line count unchanged); `PROPOSAL.md` D4 cell (78); `DRAFT-NOTES.md` line 28 |
| 2 | P:166 fixture "… and no issues, `would_block` rows, signals, areas or `mapping-refresh.json`, then import …" | `PROPOSAL.md` line 166 (design unchanged by this issue) |
| 3 | Derived artifacts regenerated; status/bookkeeping moved to round 7 | `proposed.diff`, `requirements.proposed.diff` (header names `round-7.md` and the new design digest; REQ-31 unchanged); `DRAFT-NOTES.md` lines 3, 8–11, 17–20, 28, 40 (and 10 for the digests and `round-7.md`); `PROPOSAL.md` lines 3, 278 |

## Corrections to earlier records

- `round-5.md` issue 1's replacement text and its §3 "Only the reporting verb is wrong" overlooked the `contextual_provider` capability (`internal/engine/readiness.go:212`) (issue 1).
- `round-5.md` issue 1 cites `Capability` at `readiness.go:25-32`; it is `internal/engine/config.go:53-58`.
- Round-5 issue 2's P:166 fixture omitted `mapping-refresh.json` (issue 2).
- `round-5-response.md` sweep claim 3 should read: "the only `round-5.md` hits are DN:3 (the round-5 issue tally), DN:19's historical account of the round-4 version, and RD:7 (the `-` header line naming `area-refresh-fallback/round-5.md`)".

## Sweeps (enumerated from greps run on the final files)

1. Old qualifier: grep `reports it only` in D, P, DN, RD: no hit. New wording `reports whether it was found and never its name`: D:111, P:78 and `proposed.diff`:37 (a `+` line).
2. Statements that name a single readiness capability, grep `readiness report`/`readiness reports`/`readiness .model_cli. capability`: P:34 (call-site list: "readiness `model_cli` capability (`internal/engine/readiness.go:205`)" — a consumer cite, true); P:178 ("readiness reports whether a CLI is found", no capability named); P:182 ("sees readiness report `model_cli` ready" — a scenario outcome, true, no "only"); DN:10 and RD:37 (`requirements.proposed.diff`:37, a `+` line; "the CLI whose availability readiness reports (`model_cli`)" — no "only", true as the council ruled: REQ-31 stays unchanged); DN:28 (fixed, new round-6 account). No edit was made at P:34, P:178, P:182, DN:10 (REQ-31 quote) or RD:37, following the council's statement that they carry no "only" and stay true.
3. Stale round names: grep `round-[567]\.md` after the edits: DN:3 (`round-5.md`, `round-6.md` issue tally), DN:10 (`round-7.md`), DN:19 (`round-7.md` in the new round-6 account, `round-6.md` in the round-5 account, `round-5.md` in the round-4 account), P:278 (`round-7.md`), RD:7 (`-` line, `round-5.md` in the area-refresh-fallback path), RD:8 (`+` line, `round-7.md`); none in D. Old digests `bd447753`, `b2f52886`, `e17cfecb`, `7016ae5c`, `b5b4fcb8` no longer occur in DN or P.
4. Bookkeeping: DN:3, 17, 18, 19, 20, 40 read round 6; P:3 reads "round-6 fixes applied (`round-6-response.md`), awaiting innovation-council round 7"; P:278 "round 6 recorded 3".

## Regeneration

- `proposed.diff`: `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md > proposed.diff`; a fresh run piped into `cmp - proposed.diff` matches; same 12 hunks, 116 lines; only the line-111 `+` line changed.
- `requirements.proposed.diff`: regenerated via `/tmp/reqdiff27r6` (patch the repository file with the round-5 diff, edit the design digest and `round-6.md` → `round-7.md`, `/usr/bin/diff -u` with the two labels); 40 lines, 3 hunks, +6/−6; the only line differing from the round-5 diff is its `+` line 8 (header); `/usr/bin/patch -p1 --dry-run -i docs/design-reviews/2026-10-24-cutover-cleanup/requirements.proposed.diff` from the repository root applies cleanly; the repository file stays `bf6b8d29…`.

## Digests (SHA-256) and lines

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` | 477 | `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f` |
| `scout-design.proposed.md` | 486 | `00ffd3677acd7dbaec94d7c0a88be0b0d5a18841ceed5f7dbffa6df08d8bf005` |
| `proposed.diff` | 116 | `f3d055f04691375b714e04f26eb41bac081a6582e6259cf87b828b9fb7f72973` |
| `requirements.proposed.diff` | 40 | `f8d5cfc68c0d9602397923090ec619a6f43b6ad16749f105a28e5eed428d6ca8` (patched result `50b683f9c91fe6b3424187091306ea31754e4da675be8aa0b5b1034b754fcca9`; base `SCOUT_REBUILD_REQUIREMENTS.md` `bf6b8d2944abb757dea13c36ee85ebf4159ebf5d9535ce31827357362f8f649b`) |
| `PROPOSAL.md` | 282 | `a1fb5f7c19eb8ccd1815b513652f4dc411a10a1ca932a42e8a087a206cc25efa` |
| `DRAFT-NOTES.md` | 60 | `eacd00a46bfb3f3468f76bf8610770c159937a63172d28a604a121238bf69bac` |

`DRAFT-NOTES.md` records the first five digests and cannot contain its own; its digest is recorded here.

## Judgement calls

1. The edits were applied as exact-substring replacements (each asserted unique) in an eval cell, as in earlier rounds, because the target lines are over 1,800 characters.
2. DN:19 keeps the round-5 and round-4 accounts as history and the round-6 account is prepended in the same form, naming `round-7.md`.
3. Sweep 2: `requirements.proposed.diff`:37, DN:10, P:34 and P:182 name `model_cli` alone without "only"; left as the council ruled them true.
