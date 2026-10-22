# E1b-v2 — Innovation council review, round 13 (Amendment 1)

Scope: `evaluation/change-cost-v2/AMENDMENT-1.md` (4b9a6c77…), `annotate.py` (9119dd18…), `tests.py` (5234ea8a…), `validate.py` consensus/gates path (f8643b2d…, unchanged), `DESIGN-NOTES.md` §8/§10, `reviews/round-11.md`, `round-11-response.md`, `round-12.md`, `round-12-response.md`, `validation/` and `validation/run-1-failed/`, `/tmp/e1bv2-digests-round12.txt`, `/tmp/e1bv2-digests-round13.txt`, `/tmp/e1bv2-tests-r13.log`. Read-only; `.codegraph` untouched; no secrets printed; no API calls; no predictor–outcome association computed. Independent challenger (different model family) reviewed in parallel; synthesis in §5. Digests were not recomputed by the council (no shell); the main agent's list was checked entry by entry against §10, the round-12 list and `run-1-failed/SHA256SUMS`.

## 1. Verdict: ZERO ISSUES — Amendment 1 is signed off (§4).

## 2. Round-12 resolution check — resolved

| # | Round-12 issue | Resolved? | Evidence |
|---|---|---|---|
| 1 Minor | `answer()` parsed non-`stop` completions (`content_filter`, `error`, `None`) into labels | **Yes** | `annotate.py:51-52`: `if c.get("finish_reason") != "stop" or not (c["message"].get("content") or "").strip(): sys.exit(...)`; message names the annotator and finish reason and says "return to council". `parse()` (`:53`) is reached only after that guard, so AMENDMENT-1 §4 bullet 3 ("the only way to get `unparsable` is now a complete answer outside the label list") is now true of the code. `tests.py:277` imports the real module; `:292-298` call `annotate.answer` with `length`, `content_filter`, `error`, `None`, empty content (`resp(None)`), wrong model (each `SystemExit` via `_raises`, `:266-270`), plus the normal path `("google/gemini-2.5-pro","yes")` and a complete off-list answer → `unparsable`. `AMENDMENT-1.md` §4 bullet 2 reads "aborts … unless … `finish_reason == "stop"` … `length`, `content_filter`, `error` and a missing finish reason all abort". `DESIGN-NOTES.md:106` 66 checks (count unchanged because the existing check was extended; `:109` names the `answer`/`call` checks). |
| (interpretation) | manifest-write failure after C is written | **Recorded** | `AMENDMENT-1.md` §5 bullet 3: once `annotations-C.jsonl` is written the attempt has happened; a later manifest-write failure is a Deviation, not a rerun (council round 12). |

## 3. Issues

None.

## 4. Sign-off

The council signs off Amendment 1 to E1b-v2 with zero issues. The reviewed set (all 60 entries of `/tmp/e1bv2-digests-round13.txt` equal `DESIGN-NOTES.md:121-180` entry by entry, same order):

- `PREREGISTRATION.md` 7d87913c535b286fe7e2c54419297fdfc29eebfae4a64269e33990963cb49bc1 (byte-identical to the round-10 sign-off)
- `AMENDMENT-1.md` 4b9a6c7779998eacb391af4eb453da3922a625285e26090caca76b4ad9e5f04f
- `annotate.py` 9119dd185691e2230fd0faf495bb51454657c6a13cb40902db38e447211b0d3a
- `tests.py` 5234ea8a7f7d38b94aa2ac65dfe63637575f22268bd4d967adeddb3918373479
- `validate.py` f8643b2d10ac1487a84afad00fcaac7cdfeec03a6c9e7fa4193e7fc0084c526d
- `analysis.py` 651cc1ba3e95edff8f335ffcc0174b5e39d93e30d8ba71b6aa2e4659b0e428a9
- `build_panel.py` f2e41c96c6868c81a19163abb5d079d48e99f1129ccbd42863c1af9a72461e71
- `collect.py` 6ff799d402a261f295b810e3cea708d99e1fc4d00f6b0b36a3d4853ccd6eb087
- `common.py` 0ec6de5e41c69ab261c167fa172a10cd2f1f88794a2212afcd810fda34a85baa
- `constructs.py` 80482f25a00c6214219e6674b432e28f78e1227c786f7c49bbead691a1af90d0
- `fetch_prs.py` 8db9613116fd62e803883e736cd8dfe55898a2947a7e954e89b5654e6160416a
- `finalize_panel.py` b2187dcd2b570b52d8891c18fe55cc978d128833120784cda3cccfbde93193e8
- `follow.py` e258ad83b96c066f00b137aa80426d97ac401f82f05e657570ccd1dd9bd48dcb
- `gates.py` 800e84bb7c0f3eaf494f7e400aa4ac8b324ab240f2fd59107a759dd2b42716d5
- `harness.py` 233b91a184fa6a788de4a2e97216b465e4805cd5a91b1d9e39886b07ea609016
- `jev_label.py` d94dfa090669c898766970b967b689cca1faddf30ab5e48c99099468b660ab1c
- `jev_type.py` d9accb9d419c4417998274db99c70fbc327331795de3ee64bba2eb70322c4bcb
- `power.py` 6ac3d72083a6cd2642155048f999da503c571868823b360e20f5888c48788506
- `sensitivity.py` 687665d7946bce66a5b90d644f04f79ad5b34a64df895eba6993c174b49816c9
- `requirements.txt` d01ffb51…, `prompts.json` 0b54289a…, and the 39 data/result/validation artefacts exactly as listed in §10, including `validation/annotations-A.jsonl` 20d46889…, `validation/annotations-B.jsonl` 10a4fd8d… (= `run-1-failed/SHA256SUMS:1-2`) and `validation/run-1-failed/SHA256SUMS` cfa85473….

Delta since round 12 (`/tmp/e1bv2-digests-round12.txt` vs `round13.txt`): only `AMENDMENT-1.md` (1444a6b3… → 4b9a6c77…), `annotate.py` (026bb1e2… → 9119dd18…), `tests.py` (5151ac82… → 5234ea8a…); the other 57 entries are identical. Tests: `/tmp/e1bv2-tests-r13.log` = 66 `ok` lines + 2 `jackknife 0 / 16` progress lines + `all tests passed`.

**Ordered remaining run steps** (AMENDMENT-1 §5; §10 step numbering from `PREREGISTRATION.md`):

1. `$U annotate.py --adjudicate` with `OPENROUTER_API_KEY` injected via `infisical run` (never printed). Writes `validation/annotations-C.jsonl` (95 rows, every `model_resolved` non-empty) and `annotator-prompts.json` (`max_tokens` A 20 / B 20 / C 4096, `ab_reused_from_run_1: true`). Any abort (`SystemExit` before the writes) writes nothing, does not consume the attempt, is rerun with byte-identical scripts and recorded; any change to `MAX_TOKENS` or any script byte returns to council first.
2. `$U validate.py gates` → `validation-gates.json`. G4 fail ⇒ stop, return to council (prereg §5, one shot). G5/G7 fail ⇒ T5 `not_decided`; G7 expected to fail on κ(A,B) = 0.514 regardless of C.
3. `constructs.py` — must reproduce `constructs.json` 68325887….
4. `gates.py --g3b` — `gates.json` digest changes by design (`DESIGN-NOTES.md:182`).
5. `power.py` — must reproduce `power-results.json` 2c924dcb….
6. `analysis.py` (refuses a gates file without G3b).
7. `sensitivity.py`.
8. `RESULTS.md` with everything preregistered, uncertainty intervals, `unmeasured` counts, reproduce commands, and Deviations: run-1 failure and cause, A/B reuse at `max_tokens` 20 vs C 4096, key via `infisical run`, every aborted `--adjudicate` attempt, any manifest-write failure.
9. Hypothesis ledger update; file this report as `reviews/round-13.md`.

Any change to the files above after this sign-off requires a new zero-issue round (AGENTS.md experiment rule 5).

## 5. Checked and found correct (council + challenger)

- **Abort semantics.** `answer()` (`:47-53`) aborts on wrong model, non-`stop` finish reason, empty content. `call()` (`:56-76`): `SystemExit` (BaseException) escapes `except urllib.error.HTTPError` and `except Exception`, so those paths do not retry; 401/402 abort immediately; `KeyError`/`AttributeError` from malformed bodies (missing `choices`/`message`, non-string content) are caught as a failed attempt and abort after three (`:76`). The only `return` (`:69`) follows a successful `answer()`. `list(ex.map(...))` (`:98-103`) completes before any write (`:104-109`); worker `SystemExit` is stored on the future and re-raised in the main thread [INFERENCE from CPython semantics, as in rounds 11–12].
- **`--adjudicate`.** `:91-96` reads A/B, refuses unless id sets equal the blinded items; `:97` recomputes `dis` with `validate.disagree` (`validate.py:163-165`); `:98-103` runs and writes only C; manifest provenance per annotator (`:106-109`). The guard is id-set coverage, not a digest check; the inputs are digest-frozen (A/B = `SHA256SUMS:1-2`) and prior rounds accepted this.
- **`validate.py gates`** (`:194-218`): missing/unparsable C ⇒ `consensus` None ⇒ unresolved; thresholds unchanged. No stale `annotations-C.jsonl` or `validation-gates.json` outside `run-1-failed/`.
- **Secrets.** Key read from env (`:84`), presence-checked, used only in the header (`:66`); no abort message, row, manifest or the final `print` (`:110`) contains it.
- **Amendment ↔ prereg.** §5 replaces step 11 only; §1/§3/§6 unchanged since round 12 except the §4 wording and the §5 manifest interpretation.
- **Documentation.** `DESIGN-NOTES.md:109` "truncated/empty/wrong-model completions abort" is an abbreviated, true description (round 12 required a wording change only if the count changed; it did not). The module docstring `annotate.py:11-12` "A truncated (`finish_reason` length) or empty completion also aborts" is true but non-exhaustive; the `answer()` docstring (`:44-46`) is complete.

## 6. Challenger synthesis

Challenger verdict: ISSUES FOUND — one nit: `annotate.py:7` "one call per item" vs the three-attempt retry loop (`:64-67`). **Rejected by the council, with a recorded interpretation and no file change:** the retry loop and that clause pre-date Amendment 1 (the loop was accepted in round 10 and cited in `round-11.md` issue 1; neither round-11 nor round-12 response touched the clause [INFERENCE from the responses' file:line lists]), so this is a prior decision with no new evidence; "one call per item" states the design property that matters for validity — one recorded answer per item, no multi-sampling or voting — and a transport retry yields no additional answer; the `call()` docstring (`:57-58`) documents the three attempts. Interpretation recorded: *"one call per item" means one recorded answer per item; up to three transport attempts may precede it.* Alternative considered: raise the nit, change the digest and hold a round 14 — rejected as a documentation-only change with no effect on behaviour or the record. The challenger independently confirmed the round-12 resolution, the 60-digest match, the three-file delta, the 66-check log, run-1 identity, and the abort/propagation analysis; it did not change the council's verdict.

## 7. Risks (recorded, not issues)

- `MAX_TOKENS` 4096 may still be exceeded by C's reasoning on a real item (diagnostic used a dummy prompt at 2,000): the run then aborts with "return to council", writes nothing, and a byte-identical rerun at temperature 0 will likely abort again; raising `MAX_TOKENS` needs a round 14. Pre-specified in §5; reversible.
- Executor propagation of `SystemExit` is reasoned, not tested; `tests.py:302-303` tests `call` directly. Accepted in rounds 11–12.
- The one-shot rule makes any irreversible C outcome final; G4's unresolved share fails above 15 of the 26 C-dependent type items.
