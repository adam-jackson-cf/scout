# E1b-v2 — Response to council review round 11 (Amendment 1)

I fixed all four issues; there are no rebuttals. No API call was made since round 11, nothing after step 12 has run, and `PREREGISTRATION.md` is unchanged.

| # | Issue | Fix (file:line) |
|---|---|---|
| 1 (major) | C transport failures still became `unparsable` | `annotate.py:55-75`: `call()` now aborts (`sys.exit`) after three failed attempts and names the item and the last error class. It accepts an injectable `urlopen` for tests. The only remaining route to `unparsable` is a complete answer outside the label list (`answer()` → `parse`). Test `tests.py:299-302`: a synthetic `TimeoutError` outage aborts with `SystemExit`. |
| 2 (minor) | One-shot rule under-specified; truncation message said "raise MAX_TOKENS" | `AMENDMENT-1.md` §5 (`:56-61`) now defines the attempt: the first `--adjudicate` that completes and writes `annotations-C.jsonl` (95 rows, all `model_resolved` non-empty), followed by `validate.py gates`. An abort writes nothing and does not consume the attempt; the rerun uses byte-identical scripts and is reported. Any `MAX_TOKENS` or script change needs a new zero-issue round. The `annotate.py:51` message now says "return to council". |
| 3 (minor) | `--adjudicate` would falsely record `max_tokens` 4096 for A and B | `annotate.py:104-108`: `max_tokens` is recorded per annotator, `{"A": 20, "B": 20, "C": 4096}` in `--adjudicate` mode, plus `ab_reused_from_run_1`. `AMENDMENT-1.md` §4 (`:41`) and §6 (`:67-68`) state the asymmetry and the run-1 manifest digest `383d3c42…`. |
| 4 (nit) | `DESIGN-NOTES.md:106` said 64 checks | It now says 66 (`:12`, `:106`); `:109` lists the `answer`/`call` checks. |

## Verification (main agent, 2026-10-21)

- `tests.py`: **66 checks pass** (log in `/tmp/e1bv2-tests-r12.log`).
- `PREREGISTRATION.md` digest is unchanged (`7d87913c…`).
- Digests: 60 entries in `/tmp/e1bv2-digests-round12.txt`, equal to `DESIGN-NOTES.md` §10. Changed since round 11: `AMENDMENT-1.md`, `annotate.py`, `tests.py`.
