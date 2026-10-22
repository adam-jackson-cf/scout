# E1b-v2 — Innovation council review, round 14 (Amendment 1, revision 2)

Scope: `evaluation/change-cost-v2/AMENDMENT-1.md` (527bd820…), `annotate.py` (7b7bfe68…), `tests.py` (622127ee…), `validate.py` consensus/gates path (f8643b2d…, unchanged), `PREREGISTRATION.md` §5/§10 (7d87913c…, unchanged), `DESIGN-NOTES.md` §8/§10, `reviews/round-13.md`, `round-13-response.md`, `validation/` and `validation/run-1-failed/SHA256SUMS`, `/tmp/e1bv2-digests-round13.txt`, `/tmp/e1bv2-digests-round14.txt`, `/tmp/e1bv2-tests-r14.log`. Read-only; `.codegraph` untouched; no secrets printed; no API calls; no predictor–outcome association computed. Independent challenger (different model family) reviewed in parallel without being given a leaning; synthesis in §6. Digests were not recomputed (no shell); the main agent's list was checked entry by entry against §10 and the round-13 list.

## 1. Verdict: ZERO ISSUES — Amendment 1 revision 2 is signed off (§4).

## 2. Delta check since round 13

| Entry | Round 13 | Round 14 | Checked |
|---|---|---|---|
| `AMENDMENT-1.md` | 4b9a6c77… | 527bd820… | new §2a (abort record + bounded-reasoning diagnostic); §4 budgets 8192 / C reasoning 2048, manifest `reasoning` field; §6 budget asymmetry and the aborted attempt as Deviations. §1, §3, §5 unchanged in substance. |
| `annotate.py` | 9119dd18… | 7b7bfe68… | `:31` `MAX_TOKENS = 8192`; `:32` `REASONING = {"C": {"max_tokens": 2048}}`; `:35-39` `request_body()` adds `reasoning` only when `letter in REASONING`; `:115-116` manifest records `reasoning=REASONING`. `answer()` `:51-61`, `call()` `:64-83`, `main()` write order `:109-117` unchanged in logic from round 13. |
| `tests.py` | 5234ea8a… | 622127ee… | `:299-302` new check: C's `reasoning.max_tokens < max_tokens`; A and B carry no `reasoning`; temperature 0; pinned model id. |
| other 57 entries | identical | identical | including `PREREGISTRATION.md` 7d87913c…, `validate.py` f8643b2d…, `validation/annotations-A.jsonl` 20d46889…, `-B.jsonl` 10a4fd8d… (= `run-1-failed/SHA256SUMS:1-2`), `run-1-failed/SHA256SUMS` cfa85473…. |

`DESIGN-NOTES.md` §10 (recomputed for round 14) equals `/tmp/e1bv2-digests-round14.txt` entry by entry (60); §8 now names `annotate.request_body` and 67 checks. `/tmp/e1bv2-tests-r14.log`: 67 `ok` lines + 2 `jackknife 0 / 16` progress lines + `all tests passed`. `validation/` contains no `annotations-C.jsonl`; `annotator-prompts.json` (389 B) is the run-1 manifest (no `max_tokens`/`reasoning` field), consistent with AMENDMENT §4 and `SHA256SUMS:4` 383d3c42….

## 3. The four questions

**(a) Is bounding C's reasoning after one aborted attempt methodologically acceptable? Yes.**
- Prereg §5 (`PREREGISTRATION.md:146-150`) pins model ids, temperature 0, one call per item, rubric, pinned-provider check and the consensus rule. It pins neither `max_tokens` nor a reasoning budget; run 1's manifest had no such field. These are implementation parameters, frozen by rule 5 and correctly brought back to council.
- The change was decided on infrastructure evidence only: the abort message (`annotate.py:60`) carries the finish reason, not content; nothing was written (`:109-117` all writes follow the C `map`); `round-13-response.md` records no C label seen. Nothing label-dependent could have informed the choice of 2048/8192.
- The change applies uniformly to all 95 disagreement items (`:107-109`); the prompt is unchanged (`prompt()`, so `prompt_sha256` is unchanged), blinding untouched.
- Directionally, a thinking bound can only make C less accurate relative to an unbounded C; C is blinded to pipeline labels, so it cannot bias G4 accuracy or G5 precision towards passing. The study bears the risk, which is the conservative side. Recorded as a Deviation (§6).

**(b) Is the dummy-prompt evidence adequate? Yes, for what it claims.** §2a demonstrates that OpenRouter's `reasoning.max_tokens` is honoured by Gemini (902 of 1,024; 1,887 of 2,048; both `stop`). It does not and cannot show that real items complete — that evidence could only come from calling C on study items, which would consume or peek at the attempt. The failure mode is handled: any `length`/empty/non-`stop` completion aborts before any write (`:59-60`, `:109-111`), and 6,144 tokens of headroom remain for the label.

**(c) Does the abort consume the attempt? No.** §5 defines the attempt as the first `--adjudicate` that completes and writes `annotations-C.jsonl`. Code: `out["C"] = list(ex.map(...))` `:109` must finish before the loop at `:110-111` writes; a worker `SystemExit` is re-raised in the main thread from the map iterator [INFERENCE from CPython executor semantics, as rounds 11–13] and nothing is written. A rerun with byte-identical scripts at temperature 0 selects on nothing (abort is content-independent).

**(d) Remaining infrastructure path? None found.** `answer()` rejects wrong model, non-`stop` and empty content; malformed bodies (`KeyError`/`AttributeError`) are caught at `:81` and abort after three attempts (`:83`); 401/402 abort at once (`:79`). Only a complete, non-empty, out-of-list answer becomes `unparsable` (`:46-47`), which `validate.consensus` (`:183-187`) maps to `None` → unresolved, never dropped (`gates()` `:194-218`). The key is read at `:90` and used only in the header (`:73`); no message, row, manifest or print contains it.

## 4. Sign-off

The council signs off Amendment 1 revision 2 to E1b-v2 with zero issues. The reviewed set (all 60 entries of `/tmp/e1bv2-digests-round14.txt` = `DESIGN-NOTES.md` §10):

- `PREREGISTRATION.md` 7d87913c535b286fe7e2c54419297fdfc29eebfae4a64269e33990963cb49bc1 (byte-identical to the round-10 sign-off)
- `AMENDMENT-1.md` 527bd820fca3c6b66a165b5bd785d306062bf73c7e4a280662ce2e4b1ff4b6c5
- `annotate.py` 7b7bfe6887eac6715fd637a0ada6bce02e80645ba45309762b7077471a3f5a8e
- `tests.py` 622127eec69c1a68015afdc27b646a10b0ee8d28715330089c4967904b7a2e08
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
- `requirements.txt` d01ffb51…, `prompts.json` 0b54289a…, and the 39 data/result/validation artefacts exactly as listed in §10, including `validation/annotations-A.jsonl` 20d46889…, `validation/annotations-B.jsonl` 10a4fd8d… and `validation/run-1-failed/SHA256SUMS` cfa85473….

**Ordered remaining run steps** (AMENDMENT-1 §5; §10 numbering from `PREREGISTRATION.md`):

1. `$U annotate.py --adjudicate` with `OPENROUTER_API_KEY` injected via `infisical run` (never printed). Expected: `validation/annotations-C.jsonl` (95 rows, every `model_resolved` non-empty) and `annotator-prompts.json` with `max_tokens` A 20 / B 20 / C 8192, `reasoning` `{"C": {"max_tokens": 2048}}`, `ab_reused_from_run_1: true`. Any abort writes nothing, does not consume the attempt, is rerun with byte-identical scripts and recorded; any change to `MAX_TOKENS`, `REASONING` or any script byte returns to council first.
2. `$U validate.py gates` → `validation-gates.json`. G4 fail ⇒ stop, return to council (prereg §5, one shot). G5/G7 fail ⇒ T5 `not_decided`; G7 expected to fail on κ(A,B) = 0.514 regardless of C.
3. `constructs.py` — must reproduce `constructs.json` 68325887….
4. `gates.py --g3b` — `gates.json` digest changes by design.
5. `power.py` — must reproduce `power-results.json` 2c924dcb….
6. `analysis.py` (refuses a gates file without G3b).
7. `sensitivity.py`.
8. `RESULTS.md` with everything preregistered, uncertainty intervals, `unmeasured` counts, reproduce commands, and Deviations: run-1 failure and cause; A/B reuse at `max_tokens` 20 vs C at 8192 with thinking bounded at 2048; key via `infisical run`; the aborted round-13 attempt (§2a) and every further abort with its message; any manifest-write failure.
9. Hypothesis ledger update; file this report as `reviews/round-14.md`.

Any change to the files above after this sign-off requires a new zero-issue round (AGENTS.md experiment rule 5).

## 5. Checked and found correct (council + challenger)

- Amendment ↔ code: §4 budgets = `:31-32`; manifest content = `:114-116`; "≥ 6,144 tokens for the label" = 8192 − 2048; §2a abort message = `:60` text; `round-13-response.md` line references (`:31-39`, `:115-116`, `tests.py:299-302`) are exact.
- Tests exercise the real module (`tests.py:277` import; `:290-307` `answer`, `request_body`, `call`).
- `DESIGN-NOTES.md:109` names `annotate.request_body (bounded reasoning for C only)`; §8 count 67 = log.
- Module docstring `annotate.py:12-14` ("max_tokens = MAX_TOKENS"; manifest "models, temperature, max_tokens, rubric digests") is true and non-exhaustive (omits `reasoning`, `ab_reused_from_run_1`); same class as the round-13 ruling on `:11-12` — not an issue. `REASONING` is documented inline at `:32` and in AMENDMENT §4.
- Prior decisions stand: C-only rerun with reused A/B (round 11); abort-not-attempt and manifest-failure-as-Deviation (rounds 12–13); "one call per item" = one recorded answer (round 13).

## 6. Challenger synthesis

Challenger verdict: ISSUES FOUND — one nit: `AMENDMENT-1.md:3` "Status: proposed for council round 11" while §2a describes revision 2 for round 14. **Rejected by the council — prior decision, no new evidence.** Round 12 (`round-12.md:32`) ruled the line "literally true (first proposed then) and the gating sentence is round-agnostic — not an issue"; round 13 accepted it again; revision 2 does not change the sentence's truth value or the gating clause ("runs only after a zero-issue round names this file's digest"), which remains the operative statement. Changing one header line would alter the digest and force a round 15 for no validity gain — the same trade-off round 13 made on its nit. Interpretation recorded: *the Status line records origin; the digest named in the latest zero-issue round governs.* The challenger agreed on all seven substantive questions (acceptability conditional on this round and transparent reporting; dummy evidence shows feasibility not assurance; abort writes nothing; no infrastructure-to-label path; provenance and secrets correct). It did not change the recommendation; it sharpened §3(a)–(b) wording (bounded thinking is not behaviour-neutral; the abort disclosed operational, not label, information) and the alternative below.

Alternative considered: raise `MAX_TOKENS` further without bounding reasoning. Rejected: unbounded thinking already exhausted 4,096 on a real item; no study-independent evidence bounds it, so the abort risk would persist; a bound plus headroom is the only configuration with a feasibility demonstration.

## 7. Risks (recorded, not issues)

- Gemini may overflow `reasoning.max_tokens` [INFERENCE: Google documents the budget as guidance]; 6,144 tokens of headroom make a `length` abort unlikely but not impossible; it aborts cleanly, and a different budget needs a round 15.
- C's labels under bounded thinking are not label-equivalent to an unbounded C; direction is conservative for the gates; reported as a Deviation.
- Rows do not record per-item reasoning usage; the recorded request configuration is the provenance. Executor `SystemExit` propagation and other workers' discarded completions remain reasoned, not tested (rounds 11–13).
- The one-shot rule makes the C outcome final; G4 fails if more than 15 of the 26 C-dependent type items stay unresolved.
