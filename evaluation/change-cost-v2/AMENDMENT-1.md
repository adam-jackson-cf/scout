# E1b-v2 — Amendment 1 (post-sign-off; annotator C truncation)

Status: proposed for council round 11. Under `AGENTS.md` experiment rules, the steps below run only after a zero-issue round names this file's digest. `PREREGISTRATION.md` is unchanged (byte-identical to the round-10 sign-off); this file amends its §10 execution order and nothing else.

## 1. What ran (2026-10-21)

Signed-off §10 steps 10–12 ran once:

- `validate.py sample`: 150 type, 200 pair and 128 dep items. Pair strata: jev_pos population 1,138 (100 sampled), jev_neg population 16,782 (100 sampled).
- `annotate.py`: A and B labelled all 478 items; C was called on the 95 A/B disagreements.
- `validate.py gates`.

No later step ran: no `constructs.py` re-run, no `gates.py --g3b`, no `analysis.py`, no `sensitivity.py`, and no predictor–outcome association was computed. The run-1 artefacts are preserved unchanged in `validation/run-1-failed/`, with digests in `validation/run-1-failed/SHA256SUMS`. The key was supplied by injecting `OPENROUTER_API_KEY` into the process environment with `infisical run` (user-designated project); it was never printed or stored.

## 2. Defect

All 95 C answers were `unparsable`; A and B had 0. The cause is `annotate.py`, which sent `max_tokens: 20`. `google/gemini-2.5-pro` is a reasoning model and spends its completion budget on thinking. A diagnostic call on a dummy prompt (not study data) returned:

- with `max_tokens` 20: `finish_reason=length`, content `'\n**Analyzing'`, 17 of 18 completion tokens spent on reasoning;
- with 2,000: `finish_reason=stop`, content `'yes'`.

The script turned these infrastructure truncations into `unparsable` labels, so C never adjudicated. The review rounds did not catch this, because network calls were untestable before sign-off.

### 2a. Aborted adjudication under the round-13 sign-off (revision 1 of this amendment)

The round-13 version (`MAX_TOKENS` 4096, no reasoning bound; `annotate.py` 9119dd18…) ran `annotate.py --adjudicate` once at 10:05 and aborted: `annotator C: incomplete or empty completion (finish_reason 'length'); return to council`. Unbounded Gemini thinking exhausted 4,096 tokens on at least one real item. As §5 designed, the abort wrote nothing: `annotations-A/B.jsonl` and `annotator-prompts.json` are byte-identical to run 1 (20d46889…, 10a4fd8d…, 383d3c42…), and no `annotations-C.jsonl` exists. Under §5 the abort did not consume the attempt, but changing `MAX_TOKENS` requires a new round, hence revision 2.

A second diagnostic, on a dummy reasoning-heavy prompt (not study data), showed that OpenRouter's `reasoning.max_tokens` bounds Gemini's thinking:

| Reasoning budget | `max_tokens` | `finish_reason` | Reasoning tokens used |
|---|---|---|---|
| none | 4,096 | stop | 1,796 |
| 1,024 | 8,192 | stop | 902 |
| 2,048 | 8,192 | stop | 1,887 |

## 3. What was seen (gate metrics only; run-1 `validation-gates.json`)

| Gate | Result | Failing criterion | Depends on C? |
|---|---|---|---|
| G4 | fail | unresolved 0.173 > 0.10 (κ 0.771 ≥ 0.5, accuracy 0.823 ≥ 0.70 pass) | yes: all 26 unresolved items are A/B disagreements left to C |
| G5 | fail | unresolved 0.165 > 0.15 (κ 0.715, precision 0.898, IPW recall 0.728, 81 positives pass) | yes |
| G7 | fail | κ(A,B) 0.514 < 0.6; hidden precision 0.645 < 0.80; unresolved 0.342 > 0.15; go 0.0 and jvm_dotnet 0.5 < 0.60 | κ does not depend on C, so G7 fails whatever C answers |

These are measurement-validity statistics on the annotation samples, not predictor–outcome associations.

## 4. Change

`annotate.py`:

- `MAX_TOKENS = 8192` for every annotator. C alone also gets `REASONING = {"C": {"max_tokens": 2048}}` (OpenRouter `reasoning` parameter), which bounds its thinking and leaves ≥ 6,144 tokens for the label. A and B get no reasoning parameter (`request_body()`).
- `answer()` aborts the run unless the completion finished normally (`finish_reason == "stop"`) with non-empty content: `length` (truncation), `content_filter`, `error` and a missing finish reason all abort instead of recording `unparsable`.
- `call()` aborts the run after three failed attempts (transport error, timeout, malformed response) instead of returning `unparsable`. The only way to get `unparsable` is now a complete answer outside the label list.
- `annotator-prompts.json` records `max_tokens` per annotator and the `reasoning` setting. In `--adjudicate` mode that is `max_tokens` `{"A": 20, "B": 20, "C": 8192}` and `reasoning` `{"C": {"max_tokens": 2048}}`, with `ab_reused_from_run_1: true`, because the reused A/B labels were produced at `max_tokens` 20 (run-1 manifest `383d3c42…`, which has no `max_tokens` field).
- A new `--adjudicate` mode keeps `validation/annotations-{A,B}.jsonl` and reruns only C on the A/B disagreements. It refuses to run unless A/B cover exactly the blinded items.

`tests.py` adds three checks. `annotate.answer`: `finish_reason` `length`, `content_filter`, `error` or `None`, empty content, and a wrong model each abort, and a normal answer parses. `annotate.request_body`: C's reasoning budget is below `max_tokens`, and A/B carry no reasoning parameter. `annotate.call`: a synthetic outage aborts after three attempts. 67 checks pass.

Unchanged: samples, blinding, rubric, pinned models, temperature, the A/B labels (each a complete in-list label, 0 unparsable), consensus rules, every gate threshold, and all analysis code.

## 5. Amended execution order (replaces §10 step 11 only)

```
$U annotate.py --adjudicate        # C only, on the same 95 A/B disagreements (needs OPENROUTER_API_KEY)
$U validate.py gates               # validation-gates.json
... §10 continues unchanged: constructs.py -> gates.py --g3b -> power.py -> analysis.py -> sensitivity.py -> RESULTS.md
```

**What counts as the attempt.** The attempt is the first `annotate.py --adjudicate` that completes and writes `annotations-C.jsonl` (95 rows, every `model_resolved` non-empty), followed by `validate.py gates`. Rules:

- An aborted `--adjudicate` (truncation, wrong model, outage, missing key) writes nothing, because every write happens after the C `map` finishes and `SystemExit` propagates out of it first. It does not consume the attempt; it is rerun with byte-identical scripts, and every abort and rerun is reported in `RESULTS.md` Deviations.
- Any change to `MAX_TOKENS` or any other script byte requires a new zero-issue council round first.
- Once `annotations-C.jsonl` is written, the attempt has happened. If a later write (the `annotator-prompts.json` manifest) fails, that is reported as a Deviation, not rerun (council round 12).

**One shot:** if G4 fails after the attempt, the study stops under prereg §5. There is no further annotation rerun, threshold change or re-sampling. If G5 or G7 fails, T5 is `not_decided` under prereg §5; given §3 above, G7 is expected to fail [INFERENCE: κ is fixed by the A/B labels], making H4 exploratory only.

## 6. Deviations to report in `RESULTS.md`

- This amendment: the run-1 failure, its cause, and the reuse of the A/B labels.
- Key supplied via `infisical run` environment injection.
- Annotator budgets differ: A and B ran at `max_tokens` 20 (run 1, reused); C runs at 8,192 with thinking bounded at 2,048. Both are recorded in `annotator-prompts.json`.
- The aborted round-13 `--adjudicate` attempt (§2a), and any further abort, with its abort message.
