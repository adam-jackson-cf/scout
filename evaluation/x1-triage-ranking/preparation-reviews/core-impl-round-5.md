# X1 preparation core implementation review — round 5 (confirming)

**Verdict: zero issues. X1C-R4-01 is resolved, and Core is signed off at the digests below.**
- The REQUEST's two self-pin pointers (`:99` and both cells of `:107`) now name the same file, `core-impl-round-4-response.md`.
- That file records the REQUEST's current 20,861 bytes and SHA-256 `ae6ba9800696fccd149dca18544c8eeb4f8f212481c4c9a76532c549002d0330`.
- No other REQUEST text or bound file changed.

**Scope and method.** This review was read-only and limited to X1C-R4-01. Nothing was run: no code, tests, hashes, Git commands, adoption, preflight, producer or seal. No models were called beyond the two read-only peer analyses. Three analyses fed this record:
- **My own reading** of the REQUEST's integration paragraph (`:79`), its self-pin paragraph (`:99`) and all 16 table rows (`:101–118`), plus:
  - `core-impl-round-4-response.md`, in full;
  - the tests in `execution/tests/test_cc2_delta.py` that read the REQUEST;
  - the review-directory listing.
- **`innovation-challenger` (CoreR5Challenger):** a full read of the REQUEST and the response, looking for stale pointers or digests, byte arithmetic, and tests that read the REQUEST.
- **`experiment-peer` (CoreR5Peer):** the observations that would show X1C-R4-01 is not fixed, each checked against the files.

This record does not contain the exact round-9 verdict line or a whole-line scope marker, so `review_receipt` cannot accept it as authority. The CC2 confirming round and the implementation receipt still have to provide that authority.

## Reviewed digests

The digests are Main's (`shasum -a 256`). I did not recompute them. Paths are relative to `evaluation/x1-triage-ranking/`.

| File | SHA-256 | Role |
|---|---|---|
| `preparation-reviews/identifiers-delta-cc2-REQUEST.md` | `ae6ba9800696fccd149dca18544c8eeb4f8f212481c4c9a76532c549002d0330` | Under review (changed; 20,861 bytes). Supersedes `ba3e5816…` |
| `preparation-reviews/core-impl-round-4-response.md` | `c5d49e24db6e7818f9d862c4df644e38d87ee8a66bcd95279d4a4dcc388cbc2b` | Under review (new; the REQUEST's self-pin home) |
| `preparation-reviews/core-impl-round-4.md` | `270d3e40…` (prefix as supplied by Main) | Prior round |
| `execution/preparation.py` | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` | Unchanged since round 4 |
| `execution/tests/test_journal_publication.py` | `7ea4fa923b963993067782f4aad5a13aaa9d2c0159c7dc49b5112cb5ea97ee0c` | Unchanged since round 4 |
| `execution/tests/test_cc2_seal.py` | `91e33dcac2d9ab6fcf0a9160cc01755b0fbb051f9c9f56e4e6504136f6190f82` | Unchanged |
| `execution/tests/test_binding_seal.py` | `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593` | Unchanged |
| `execution/tests/test_cc2_delta.py` | `4844221de3ecb6c52b81a770ad7dadbcc42d8e0122636a29de2cdb1fc0e9d8d7` | Unchanged |
| `execution/identifiers/cc2_delta.py` | `657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6` | Unchanged |
| `execution/identifiers/cc2_readers.py` | `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b` | Unchanged |
| `execution/identifiers/cc2_schemas.py` | `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f` | Unchanged |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | Unchanged; executed |
| `preparation-reviews/core-impl-round-3.md` | `2e937a529e44bdf95500abe15ebe4f0e45dc38c754211076c4a3ce900fa3097b` | Prior round |
| `preparation-reviews/core-impl-round-3-response.md` | `900fbcfa5c73401fee4be45ff03275d95a3277ecaf07b1bbe4694b819f04a711` | Prior response |
| `preparation-reviews/identifiers-delta-cc2-round-3.md` | `be86e13ef89695314d6a70e51fa86abd46a6dc8fc2ffc0d4c63dd7f188655dc8` | Context |

**Main's test results at these bytes** (recorded output: `core-impl-round-4-response.md`). I did not re-run any of them.
- `test_request_requires_final_shared_code_freeze_order`: 1 test OK.
- Full Python suite: 289 OK, the same count as round 4.
- No Go check was re-run. This round changed no code.

## Disposition of prior issues

| ID | Status | Evidence |
|---|---|---|
| X1C-R4-01 (Minor) | **Resolved as specified** | See the five checks below. |
| X1C-R3-01 and every earlier issue | Still resolved | No code or test file changed. `preparation.py` is still `9572aed3…`, as in REQUEST `:79` and `:109` and the response. |
| Round-4 S1–S5 | Still suggestions | Not reopened. S3 (adoption can orphan a binding output) is still a binding-lane or runbook matter, and must be resolved before canonical binding (c). |

**How X1C-R4-01 meets the round-4 fix:**

1. **One pointer target.** REQUEST `:99` says the request's "final SHA-256 and byte count are recorded in `core-impl-round-4-response.md`". Both cells of the request row at `:107` read "Final request pin in core-impl-round-4-response.md". Those are the REQUEST's only references to a response file. It contains none of `7f8d390f`, `c32bd400`, `ba3e5816`, `402f0a09`, `50213`, `20847`, `20873`, `round-2-response` or `round-3-response`.
2. **The file it names records the current pin.** The self-pin table in `core-impl-round-4-response.md` (`:18–20`) gives the REQUEST as `20861` bytes and `ae6ba9800696fccd149dca18544c8eeb4f8f212481c4c9a76532c549002d0330`, matching Main's brief. No `identifiers-delta-cc2-round-4-response.md` exists, so no other record competes. The round-2 response still carries 20,847 / `c32bd400…` at `:63`, but it is a historical record that nothing points to any more, so that is correct.
3. **No other text changed.** Only the two pointer lines differ from `ba3e5816…`, and the byte count accounts for every changed byte:
   - At `:99`, `identifiers-delta-cc2-round-2-response.md` (41 ASCII bytes) became `core-impl-round-4-response.md` (29 bytes), a change of −12 bytes.
   - At `:107`, both cells changed `round-3` to `round-4`, a change of 0 bytes.
   - 20,873 − 12 = 20,861, which matches the response.
   
   The other text is still at round-4 values:
   - `:79` still names `preparation.py` `9572aed3…` and `verify_inputs.py` `7d6f89ce…`.
   - `:109` still reads 50257 / `9572aed3…`.
   - The other 14 rows match round 4: cc2_delta 14393 / `657aeb47…`, cc2_readers 8012 / `1ad375a9…`, cc2_schemas 4464 / `db845451…`, test_cc2_delta 22402 / `4844221d…`, projection 7637 / `2408ed45…`, verify_inputs 11211 / `7d6f89ce…`, run-config 5948 / `c817f678…`, exposure-sources 7792170 / `67dc5a4b…`, journal `906389d7…` 10301 / `6aa8802a…`, manifest 27924749 / `23a911ae…`, source-coverage 30151508 / `24f93c47…`, delta-review 4635 / `240aa16d…`, unresolved 91278 / `19d10781…`, familiarity 20588100 / `96751811…`.
   
   **[INFERENCE]** Reading cannot rule out an edit elsewhere that adds and removes exactly the same number of bytes. Nothing read suggests one. That the digest is correct rests on Main's `shasum`.
4. **No other file was edited.** The response says only the REQUEST and the response changed, and that no code, test, receipt, journal, clock, amendment or historical record changed. This fits the unchanged `test_cc2_delta.py` pin and the same 289 test count. The response repeats only the REQUEST pin and states that the other 15 controls are unchanged. That is all round 4 required; it creates no conflicting copy of the table.
5. **The tests do not see this defect.** `test_cc2_delta.py:269–275` and `:277–289` read the REQUEST but never check which response file it names. So the passing tests show that the existing checks still hold; they are not evidence for this fix. The evidence is the reading above.

## New issues

None.

## Suggestions (not required for approval)

- **S1. Record the measured byte counts** (peer). In the CC2 confirming round, record `shasum -a 256` and `wc -c` output for the REQUEST, `core-impl-round-4-response.md` and `preparation.py`, so the pins don't rest only on response prose.
- **S2. Add a test that the two pointers agree** (peer and challenger; CC2 round-4 suggestion 1). An assertion that `:99` and `:107` name the same file would catch this class of defect. It is optional, and adding it would change the pin for `test_cc2_delta.py` and need another review.
- **S3. Check the receipt before the journal starts** (peer; CC2 round-4 suggestion 3, CC2 lane). Run `review_gate` before `Journal(...)` in `cc2_delta.py`, so a wrong receipt is refused without first publishing a start journal.
- **S4. Keep the current pointer wording** (challenger). Round 4's indirect wording ("the response named in the request row") would also have worked. Switching to it now would only invalidate the current REQUEST pin.

## Recommendation

Adopt these bytes as Core's signed-off state. Then proceed in the order fixed by the REQUEST and round 2:
1. the CC2 confirming round at all 16 actual digests, including the REQUEST at `ae6ba980…`;
2. the authentic implementation receipt;
3. the single producer run;
4. the output review;
5. the remaining actions and the seal.

Resolve round-4 S3 before canonical binding (c). Any further edit to a signed file needs a new Core round.

## Alternatives

- **A. Sign off now (recommended).** The fix is exactly what was required, and no issue remains.
- **B. Add the pointer-agreement test (S2) first.** It changes a bound test pin and costs another round for a defect that is already fixed by reading.
- **C. Switch `:99` to indirect wording.** It invalidates the REQUEST pin for no gain.

## Risks

- **Byte checks rest on Main.** The byte and digest checks depend on Main's `shasum`. The arithmetic and the reading are consistent with them.
- **This is not execution authority by itself.** The CC2 confirming round and the receipt are still required.
- **Freeze from the producer run.** From the producer run onwards, none of the 16 bound files may change before the seal, and the run is irreversible.
- **Binding is still exposed.** Canonical binding stays exposed to round-4 S3 until the binding lane or the runbook addresses it.
- **Clock time.** Every further round uses time on the 45-day clock (deadline `2026-12-24T03:22:58.701909Z`).

## Execution outline

1. Main writes this record and computes its digest.
2. Hold the CC2 confirming round. Its scope block names the 16 actual digests, including REQUEST `ae6ba9800696fccd149dca18544c8eeb4f8f212481c4c9a76532c549002d0330` and `preparation.py` `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`.
3. Create the authentic `identifiers-delta-cc2-implementation` receipt, and only after that the single canonical producer run. Then the output review, the remaining actions and the seal.
4. Before canonical binding (c), resolve round-4 S3.

## How the peers shaped this record

- **Peer:**
  - Tested the four observations that would show X1C-R4-01 is not fixed. All were refuted.
  - Checked all 14 other rows against round 4.
  - Supplied S1–S3.
- **Challenger:**
  - Confirmed there are no stale pointers or digest strings anywhere in the REQUEST.
  - Confirmed that the response's partial table is enough.
  - Confirmed that neither REQUEST-reading test checks the pointer.
  - Supplied S4.
- **Mine:**
  - The byte arithmetic.
  - The row-by-row comparison of the table with round 4.
  - The check that no competing response file exists.
  - The verdict and the sign-off scope.

## Sign-off

The innovation council reports zero issues in Core implementation review round 5 and signs off Core at these exact digests:
- **Under review:**
  - `identifiers-delta-cc2-REQUEST.md` `ae6ba9800696fccd149dca18544c8eeb4f8f212481c4c9a76532c549002d0330` (20,861 bytes)
  - `core-impl-round-4-response.md` `c5d49e24db6e7818f9d862c4df644e38d87ee8a66bcd95279d4a4dcc388cbc2b`
- **Prior record:** `core-impl-round-4.md` `270d3e40…` (prefix as supplied by Main).
- **Round-4 digests, unchanged:**
  - `preparation.py` `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`
  - `test_journal_publication.py` `7ea4fa923b963993067782f4aad5a13aaa9d2c0159c7dc49b5112cb5ea97ee0c`
  - `test_cc2_seal.py` `91e33dcac2d9ab6fcf0a9160cc01755b0fbb051f9c9f56e4e6504136f6190f82`
  - `test_binding_seal.py` `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593`
  - `test_cc2_delta.py` `4844221de3ecb6c52b81a770ad7dadbcc42d8e0122636a29de2cdb1fc0e9d8d7`
  - `cc2_delta.py` `657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6`
  - `cc2_readers.py` `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b`
  - `cc2_schemas.py` `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f`
  - `verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`
  - `core-impl-round-3.md` `2e937a529e44bdf95500abe15ebe4f0e45dc38c754211076c4a3ce900fa3097b`
  - `core-impl-round-3-response.md` `900fbcfa5c73401fee4be45ff03275d95a3277ecaf07b1bbe4694b819f04a711`
  - `identifiers-delta-cc2-round-3.md` `be86e13ef89695314d6a70e51fa86abd46a6dc8fc2ffc0d4c63dd7f188655dc8`

The REQUEST at `ba3e5816eadbda4af7a541e66ea6e75fc0d525cd2c3e8a439608c45be4856119` is superseded and is not signed off.

**What this sign-off does not authorize.** It covers Core only. It does not authorize the CC2 producer run, the receipt or the seal. Those still need the CC2 confirming round and the receipt chain in the REQUEST. Any change to a file named above needs a new Core round with zero issues.

Verdict: ZERO ISSUES — signed off
