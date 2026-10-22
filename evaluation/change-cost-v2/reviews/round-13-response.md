# E1b-v2 — Post-round-13 run record and Amendment 1 revision 2

Round 13 signed off Amendment 1 with zero issues. Its run step 1 was attempted and aborted as designed. That revision fixed `MAX_TOKENS` and forbade changing it without a new round, so this revision comes back to council (round 14).

## What happened

At 10:05, `shasum -a 256 -c` confirmed all 60 round-13 digests. `annotate.py --adjudicate` then ran (key injected by `infisical run`) and exited with:

`annotator C: incomplete or empty completion (finish_reason 'length'); return to council`

Nothing was written. `annotations-A/B.jsonl` and `annotator-prompts.json` are unchanged (20d46889…, 10a4fd8d…, 383d3c42…), and no `annotations-C.jsonl` exists. Under §5 the abort did not consume the attempt. No C label was seen, and no label from the aborted run was persisted or printed.

## Change (revision 2)

| Item | Change (file:line) |
|---|---|
| Unbounded Gemini thinking exhausted 4,096 tokens | `annotate.py:31-39`: `MAX_TOKENS` is now 8192, and `REASONING = {"C": {"max_tokens": 2048}}` is sent only for C via `request_body()`. A and B get no reasoning parameter. The manifest records `reasoning` (`:115-116`). Evidence that the bound is honoured comes from a dummy, non-study prompt (`AMENDMENT-1.md` §2a): 902 reasoning tokens used against a 1,024 budget and 1,887 against 2,048, both ending `stop`. |
| Test | `tests.py:299-302`: `annotate.request_body` gives C a reasoning budget below `max_tokens`; A and B carry no `reasoning`. |
| Amendment text | New §2a records the aborted attempt and the diagnostic. §4 and §6 now give the new budgets and the A/B-vs-C asymmetry. |

## Verification (main agent, 2026-10-21)

- `tests.py`: **67 checks pass** (log in `/tmp/e1bv2-tests-r14.log`).
- `PREREGISTRATION.md` is unchanged (`7d87913c…`).
- Digests: 60 entries in `/tmp/e1bv2-digests-round14.txt`, equal to DESIGN-NOTES §10. Changed since round 13: `AMENDMENT-1.md`, `annotate.py` and `tests.py`.
