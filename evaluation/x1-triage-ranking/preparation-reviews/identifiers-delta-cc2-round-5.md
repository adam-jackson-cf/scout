# X1 preparation (d) change-cost-v2 identifier delta — confirming implementation review — round 5

Paths are relative to `evaluation/x1-triage-ranking/` unless absolute. This round confirms the fix for CC2-R4-01 and only that issue. It also checks that all 16 bound pins are current.

The review was static and read-only. No CC2 data payload was opened and nothing was executed. The only model calls were two read-only peer analyses. The council had no shell, so every digest below is Main's `shasum -a 256` value. The council did not recompute any of them.

## Reviewed digests

| File | Bytes | SHA-256 | Status since round 4 |
|---|---:|---|---|
| `preparation-reviews/identifiers-delta-cc2-REQUEST.md` | 20861 | `ae6ba9800696fccd149dca18544c8eeb4f8f212481c4c9a76532c549002d0330` | changed (was `ba3e5816…`, 20873); CC2-R4-01 fix |
| `execution/preparation.py` | 50257 | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` | unchanged |
| `execution/verify_inputs.py` | 11211 | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | unchanged; already executed |
| `execution/identifiers/cc2_delta.py` | 14393 | `657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6` | unchanged |
| `execution/identifiers/cc2_readers.py` | 8012 | `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b` | unchanged |
| `execution/identifiers/cc2_schemas.py` | 4464 | `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f` | unchanged |
| `execution/tests/test_cc2_delta.py` | 22402 | `4844221de3ecb6c52b81a770ad7dadbcc42d8e0122636a29de2cdb1fc0e9d8d7` | unchanged |
| `execution/identifiers/projection.py` | 7637 | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | unchanged |
| `execution/run-config.json` | 5948 | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `preparation/exposure-sources.json` | 7792170 | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` | signed source manifest |
| `preparation-run/journal/906389d7212a401b8c7f781d37960d84.json` | 10301 | `6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c` | baseline action-d journal |
| `preparation-run/identifiers/manifest.json` | 27924749 | `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d` | baseline output |
| `preparation-run/identifiers/source-coverage.json` | 30151508 | `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724` | baseline output |
| `preparation-run/identifiers/delta-review.json` | 4635 | `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04` | baseline output |
| `preparation-run/identifiers/unresolved.json` | 91278 | `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9` | baseline output |
| `preparation-run/identifiers/familiarity.jsonl` | 20588100 | `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30` | baseline output |

The test files `test_cc2_seal.py` (`91e33dca…`), `test_binding_seal.py` (`7c05eaf2…`) and `test_journal_publication.py` (`7ea4fa92…`) are unchanged since round 4. They are not CC2 controls.

**Context records (not bound):**
- `preparation-reviews/core-impl-round-4-response.md` `c5d49e24db6e7818f9d862c4df644e38d87ee8a66bcd95279d4a4dcc388cbc2b`: the current external record of the REQUEST's own pin.
- `preparation-reviews/identifiers-delta-cc2-round-4.md` `fbf2b30f…`: a prefix supplied by Main.
- `preparation-reviews/core-impl-round-4.md`: read for its verdict.

## Prior-issue resolution

| ID | Status | Evidence |
|---|---|---|
| CC2-R4-01 (Minor): stale self-pin pointer at REQUEST `:99`, contradicting `:107` | **Closed. Fixed as prescribed.** | `REQUEST.md:99` now says the request's "final SHA-256 and byte count are recorded in `core-impl-round-4-response.md` after this file is written". Both cells of `:107` read "Final request pin in core-impl-round-4-response.md". So the two pointers name one record.<br>That record gives the REQUEST as 20861 / `ae6ba9800696fccd149dca18544c8eeb4f8f212481c4c9a76532c549002d0330` (`core-impl-round-4-response.md:20`), equal to Main's digest. It states that `preparation.py` was not edited and that its SHA-256 remains `9572aed3…`.<br>A grep of the REQUEST for `c32bd400`, `ba3e5816`, `7f8d390f`, `288073ce`, `858031`, `402f0a09`, `round-2-response` and `round-3-response` finds nothing. No competing self-pin source remains.<br>The historical records were not edited, as round 4 required: `identifiers-delta-cc2-round-2-response.md:63` still holds `c32bd400…` and `core-impl-round-3-response.md:90` still holds `ba3e5816…`. |
| R1-01–R1-03, R2-01, CB-1 / ID-01 allowlist | Closed or accepted earlier; not reopened | The producer files are byte-identical to rounds 3 and 4. The REQUEST sections that carry these controls are unchanged (see below). |
| Round 4's trace of producer and seal integration at `preparation.py` `9572aed3…` | Still holds | `preparation.py` is unchanged, so the round-4 lock-order and contract tables apply as recorded. |

## REQUEST changes and the pin check

- **Byte accounting.**
  - `:99`: `identifiers-delta-cc2-round-2-response.md` (41 characters) became `core-impl-round-4-response.md` (29 characters), a change of −12.
  - `:107`: `round-3` became `round-4` in two cells of equal length, a change of 0.
  - The net is −12, which equals 20873 → 20861 exactly.
  - [INFERENCE] A byte count cannot rule out a same-length edit elsewhere, and the old bytes were not available to diff. The council read the whole REQUEST. These sections read as signed in round 4: the interlock and receipt rules (`:57`), baseline pins (`:59–69`), merge contract, integration paragraph (`:79`), freeze order (`:83`), post-run freeze (`:85`), output review (`:87`) and command (`:91`). The writer states that nothing else changed (`core-impl-round-4-response.md:9–14`). Suggestion 1 would close this limit.
- **Helper pins.** `REQUEST.md:79` names `preparation.py` `9572aed3…` and `verify_inputs.py` `7d6f89ce…`. Rows `:109` (50257 / `9572aed3…`) and `:110` (11211 / `7d6f89ce…`) agree. Both equal Main's digests.
- **16-row table.** The 15 literal rows (`:103–106,108–118`) equal the reviewed digests above in absolute path, bytes and full SHA-256. The sixteenth row (`:107`) defers to `core-impl-round-4-response.md`, which gives 20861 / `ae6ba980…`. That equals Main's REQUEST digest. **All 16 pins are current.**
- **Test coverage.** `test_cc2_delta.py:278–290` (`test_request_requires_final_shared_code_freeze_order`) still asserts the freeze-order phrases and the live `pin()` digests of both helpers against the REQUEST. Main reports that targeted test as 1 OK and the full suite as 289 OK (`core-impl-round-4-response.md:39–58`). The council did not re-run them.

## New issues

None.

## Sign-off condition: Core approval is not yet on disk

`core-impl-round-4.md:3,205` records one open issue (X1C-R4-01, the same pointer defect), and no `core-impl-round-5.md` exists. `REQUEST.md:83` requires zero-issue Core and delta seal-integration reviews naming the same final `preparation.py` and `verify_inputs.py` digests **before the receipt and the producer run**. It does not require them before this confirming round.

The council records this as a sign-off condition, not an issue, for the reasons round 3 gave:
- The REQUEST already forbids running out of order.
- Any Core-driven change to a bound file changes a receipt row, so `review_gate` refuses.
- The residual failure requires Main to break the signed order.

Core round 4's outline places the CC2 confirming round after Core round 5. Holding it first departs from that recommendation but not from the signed REQUEST.

## Suggestions (not blocking)

1. **Record the diff.** Have the next Core round record Main's `diff` of REQUEST `ba3e5816…` against `ae6ba980…`; two changed lines are expected. That removes the same-length limit noted above.
2. **Keep the pointer target unchanged.** Do not edit `core-impl-round-4-response.md`; its digest is recorded above as context. The receipt and the scope block bind the literal digest, not the pointer (`REQUEST.md:99`: "a self-pin description cannot substitute"). So a later edit there cannot corrupt the gate, but it would confuse the record.
3. **Round-4 suggestions 1 and 3 and round-3 suggestions 1–4 remain open:**
   - a pointer-consistency test;
   - validate the receipt before reserving the journal;
   - freeze the receipt and this record after the run;
   - check every literal packet row against `pin()`;
   - an end-to-end producer-to-seal test;
   - bind `producer_pins` to the reviewed controls.

## Peers and where they changed the recommendation

- **innovation-challenger (Cc2R5Challenger):** independently confirmed:
  - the pointer agreement and the −12-byte accounting;
  - all 15 literal rows and `:79`;
  - the absence of stale digests;
  - that Core round 4 is not zero-issue approval.

  It advised against rewording `:99` again, since that would force another pin. It changed nothing in the verdict.
- **experiment-peer (Cc2R5Peer):**
  - Refuted H1 (other edits, with the same-length limit), H2 (stale pin) and H5 (helper test). It found no other mention of `ae6ba980` that could compete.
  - Classified H4 (Core sequencing) as a sign-off condition, consistent with round 3.
  - Contributed suggestions 1 and 2 and the wording of the condition below.
- **Council:** made the same findings by direct reading of `REQUEST.md:79–118`, a stale-reference grep and the Core round-4 verdict. No peer changed the recommendation.

## Verification

- **Static reading:**
  - `REQUEST.md:79–118`, plus a stale-reference grep;
  - `core-impl-round-4-response.md`;
  - `core-impl-round-4.md` (verdict and issue rows, by grep);
  - a directory listing of `preparation-reviews/`;
  - for unchanged files, the round-4 trace.
- **Main's reported results** (not re-run by the council): targeted helper-pin test 1 OK; full Python suite 289 OK. No Go check was re-run in this REQUEST-only revision.
- **Before building the receipt**, Main should:
  1. Re-run `shasum -a 256` and `wc -c` on all 16 bound paths and confirm they equal the scope block below.
  2. Generate the receipt rows with `verify_inputs.pin()` on those exact absolute paths.
  3. Check that the receipt names this record and passes `preparation.review_receipt` for scope `identifier-delta-cc2-implementation` before the single canonical run.

Verdict: ZERO ISSUES — signed off

**Zero issues. Signed off.**

## Sign-off

The innovation council signs off the CC2 identifier-delta implementation, its allowlist, its consistency with the seal integration and the confirming request packet, at the 16 digests named in the scope block below.

**Conditions** (from `REQUEST.md:83–87`):
- **Before the receipt and the run.** This record is execution authority only once a Core implementation review and the delta seal-integration review have both recorded zero issues naming `preparation.py` `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` and `verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`. Core round 4 does not meet this condition.
- **If any of the 16 bound files changes before the run,** including any change to the REQUEST that Core round 5 requires, this sign-off lapses and a new confirming round is required.
- **After the run,** none of the 16 files, the receipt or this record may change before the seal.

Signed-off scope: identifier-delta-cc2-implementation

- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/cc2_delta.py` — 14393 bytes — `657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/cc2_readers.py` — 8012 bytes — `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/cc2_schemas.py` — 4464 bytes — `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests/test_cc2_delta.py` — 22402 bytes — `4844221de3ecb6c52b81a770ad7dadbcc42d8e0122636a29de2cdb1fc0e9d8d7`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/identifiers-delta-cc2-REQUEST.md` — 20861 bytes — `ae6ba9800696fccd149dca18544c8eeb4f8f212481c4c9a76532c549002d0330`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/projection.py` — 7637 bytes — `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/preparation.py` — 50257 bytes — `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/verify_inputs.py` — 11211 bytes — `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/run-config.json` — 5948 bytes — `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation/exposure-sources.json` — 7792170 bytes — `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/journal/906389d7212a401b8c7f781d37960d84.json` — 10301 bytes — `6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/manifest.json` — 27924749 bytes — `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/source-coverage.json` — 30151508 bytes — `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/delta-review.json` — 4635 bytes — `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/unresolved.json` — 91278 bytes — `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/familiarity.jsonl` — 20588100 bytes — `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30`
