# X1 preparation (d) change-cost-v2 identifier delta — confirming implementation review — round 4

Paths are relative to `evaluation/x1-triage-ranking/` unless absolute.

This round confirms the CC2 producer and its seal integration at the shared-helper bytes produced by the Core round-3 fix (X1C-R3-01). The round-3 sign-off (`identifiers-delta-cc2-round-3.md`) was given at `preparation.py` `7f8d390f…` and REQUEST `c32bd400…`. It does not carry over to the new bytes (REQUEST `:83`).

The review was static and read-only:
- No CC2 data payload was opened and nothing was executed.
- No model was called beyond two read-only peer analyses.
- The council had no shell, so every digest below is Main's `shasum -a 256` value. The council did not recompute any of them.

## Reviewed digests

| File | Bytes | SHA-256 | Status since round 3 |
|---|---:|---|---|
| `execution/preparation.py` | 50257 | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` | changed (was `7f8d390f…`, 50213); X1C-R3-01 lock-scope fix |
| `execution/tests/test_journal_publication.py` | 8900 | `7ea4fa923b963993067782f4aad5a13aaa9d2c0159c7dc49b5112cb5ea97ee0c` | changed (was `6d8a0f1a…`); Core test, not a CC2 control |
| `preparation-reviews/identifiers-delta-cc2-REQUEST.md` | 20873 | `ba3e5816eadbda4af7a541e66ea6e75fc0d525cd2c3e8a439608c45be4856119` | changed (was `c32bd400…`, 20847) |
| `execution/tests/test_cc2_seal.py` | — | `91e33dcac2d9ab6fcf0a9160cc01755b0fbb051f9c9f56e4e6504136f6190f82` | unchanged (equals `core-impl-round-3.md:27`) |
| `execution/tests/test_binding_seal.py` | — | `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593` | unchanged (equals `core-impl-round-3.md:28`) |
| `execution/tests/test_cc2_delta.py` | 22402 | `4844221de3ecb6c52b81a770ad7dadbcc42d8e0122636a29de2cdb1fc0e9d8d7` | unchanged |
| `execution/identifiers/cc2_delta.py` | 14393 | `657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6` | unchanged |
| `execution/identifiers/cc2_readers.py` | 8012 | `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b` | unchanged since round 1 |
| `execution/identifiers/cc2_schemas.py` | 4464 | `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f` | unchanged since round 1 |
| `execution/verify_inputs.py` | 11211 | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | unchanged; already executed |

Context records, not bound:
- `preparation-reviews/core-impl-round-3.md` `2e937a529e44bdf95500abe15ebe4f0e45dc38c754211076c4a3ce900fa3097b`
- `preparation-reviews/core-impl-round-3-response.md` `900fbcfa5c73401fee4be45ff03275d95a3277ecaf07b1bbe4694b819f04a711`
- `preparation-reviews/identifiers-delta-cc2-round-3.md` `be86e13ef89695314d6a70e51fa86abd46a6dc8fc2ffc0d4c63dd7f188655dc8`
- `preparation-reviews/identifiers-delta-cc2-round-2-response.md`: read for its `:63–66` packet rows.

The other bound controls are unchanged from the round-3 table and are re-listed in the REQUEST at `:108,110–118`:
- `projection.py` `2408ed45…`
- `run-config.json` `c817f678…`
- `exposure-sources.json` `67dc5a4b…`
- baseline journal `6aa8802a…`
- five baseline outputs: `23a911ae…`, `24f93c47…`, `240aa16d…`, `19d10781…`, `96751811…`

**Pin check.**
- **Helper pins.** `REQUEST.md:79` names `preparation.py` `9572aed3…` and `verify_inputs.py` `7d6f89ce…`. `REQUEST.md:109` gives `preparation.py` as 50257 / `9572aed3…`. Both equal Main's digests and `core-impl-round-3-response.md:88`.
- **The other 14 literal rows** (`:103–106,108,110–118`) equal the round-3 signed values in path, bytes and digest.
- **Self-pin row.** Row `:107` defers to `core-impl-round-3-response.md`. That file gives 20873 / `ba3e5816…` (`:90`), equal to Main's REQUEST digest.
- **Exception: `REQUEST.md:99`.** See CC2-R4-01.

## Prior-issue resolution

| ID | Status | Evidence |
|---|---|---|
| R1-01, R1-02, R1-03 | Closed (round 2); not reopened | `cc2_delta.py` is byte-identical (`657aeb47…`). |
| Round-1 CB-1 (`raw/q` transport-cache boundary), ID-01 allowlist | Accepted; carried forward | `cc2_readers.py` and `cc2_schemas.py` are unchanged. The REQUEST sections that define them are unchanged (see "REQUEST changes"). |
| R2-01 (shared code frozen by a passed CC2 journal while still changing) | Closed in round 3; the control still holds | `REQUEST.md:83` (freeze order) and `:85` (post-run freeze) are unchanged. This round exists because of that control: the helper changed, so round 3's authority lapsed as its sign-off conditions state. |
| Round-3 sign-off condition (Core round 3 zero issues at the same helper digests) | Superseded | Core round 3 did not pass. It raised X1C-R3-01 (`core-impl-round-3.md:3,53`). The helper changed to `9572aed3…`. No `core-impl-round-4.md` exists on disk. The condition now attaches to Core round 4 at the final helper bytes. |

## What changed in `preparation.py`, and the effect on CC2

**Byte accounting.** The writer describes one change: the publication lock is now taken before the `RunConfig` re-check (`core-impl-round-3-response.md`, "Required fix").
- Core round 3 placed the reload at `:357` and the publication lock at `:369` (`core-impl-round-3.md:45,53`).
- In the current file, `self.directory` is at `:357` and `with journal_publication(self.directory):` at `:358`.
- Eleven lines moved one level deeper, into the lock: `:359–369` (`if` / `raise` / `self.id` / `started_at` / the five-line `self.base = dict(…)` / `start_path` / `end_path`).
- 11 × 4 spaces = 44 bytes, which equals 50257 − 50213 exactly. The line count is unchanged.

Every seal-side line that round 3 cited is still at the same line number with the content round 3 described:
- `read_clock` `:205–208`
- `verify_output_accounting` `:499–514`
- `collect_cc2_delta` `:532–593`, including `:535–538`, `:539–553`, `:563`, `:567`, `:570`, `:576–583`, `:584–588` and `:590`
- `collect_journals` `:617–637`
- `collect_seal` → `collect_cc2_delta` `:732–735`

[INFERENCE] A byte count cannot rule out an edit that keeps the same length. The old bytes were not available to diff (Suggestion 2).

**The CC2 producer under the new lock order** (`cc2_delta.py:217–281`, `preparation.py:237–258,340–381`):
1. Unlocked `load_config` and canonical-path, interpreter and source checks (`:218–236`). This is Core round 3's accepted S1: a spurious refusal here reserves nothing.
2. `operation_lock(config)`: shared and non-blocking (`:237`).
3. `Journal(...)` (`:238`). Its `operation_lock(blocking=True)` re-enters the lock already held (`preparation.py:241–245`), with no second flock and no upgrade.
4. Inside the journal, the clock is read with `read_clock`. Then the directory flock is taken (`:358`), and under it run `load_config` (`:359`), the predecessor pins and the exclusive start write (`:370–371`).
   - The calls inside `load_config` — `review_receipt`, `verify_adoption`, `verify_journal_configuration` — take no lock.
   - The only lock call sites in `preparation.py` are `:264,355,374,404,753,829` (operation lock) and `:358,379` (publication lock).
   - So the publication lock never nests, and every acquirer takes the operation lock first, then the publication lock. No deadlock is possible.
5. The lock is released before `baseline`, `review_gate`, projection and `publish`. `finish` takes the publication lock again separately (`:379`).
6. A configuration mismatch or a parse failure still raises before the `uuid` and the start record, and before `root.mkdir()` (`cc2_delta.py:187`).

**Contracts between producer and seal** (re-traced at the current bytes; unchanged from round 3):

| Contract | Producer (`cc2_delta.py`) | Seal (`preparation.py`) | Result |
|---|---|---|---|
| Imports | `Journal`, `load_config`, `operation_lock`, `paths`, `review_receipt`, `pin` | All present, with the same signatures. `Journal(config, config_path, clock, action, source_pins=…)` is unchanged (`:353`). | Match |
| 16 controls | `:231–236` + source + config + baseline (`:243`) | `:539–553` | Same `(path, sha256)` set |
| Implementation receipt | Canonical path and scope `identifier-delta-cc2-implementation` (`:21,222–227`) | `:535–538`; `implementation_review == pin(review_path)` (`:567`) | Match |
| Outputs, kinds and order | `:185–214` | `:554–560` | Match |
| Common fields | `:271–274`; manifest `:188–191`; familiarity first line `:201–202` | `:563–575` | Match |
| Producer identification | Action `d`; `source_pins` = 8 implementation pins + source (`:238`); role-free output pins in creation order (`:214,276`) | `:576–583` | Match |
| Output review | — | `OUT/reviews/identifier-delta-cc2-output.json`, scope `identifier-delta-cc2-output` (`:584–588`) | Matches `REQUEST.md:87` |
| Output accounting | Three files under its own root | `:570,590`; `verify_output_accounting` `:499–514` | No unaccounted bytes |

`test_journal_publication.py` is not a CC2 control, and the producer does not import it. Its change does not affect CC2.

The peer's H4 analysis found that the new positive test and the counterfactual catch the old ordering. B fails with `JSONDecodeError`, which trips `test_journal_publication.py:124,140`, and the counterfactual asserts this deterministically at `:120–122,136–142`. The only exposure is the 0.5 s timing caveat already accepted as Core S8. The full adjudication of that test belongs to Core round 4.

## REQUEST changes

- **The +26 bytes come from `:107` alone.**
  - The digest swaps at `:79` and `:109` keep the same length: 64 hex characters, and 50213 and 50257 both have five digits.
  - The new `:107` cells read "Final request pin in core-impl-round-3-response.md", 50 characters each, twice.
  - +26 means the old cells were 37 characters each. "Final request pin in round-2 response" is exactly 37 characters and matches round 3's description ("defers to the round-2 response", `identifiers-delta-cc2-round-3.md:30`). [INFERENCE: the old bytes were not seen.]
- **Nothing else changed in substance.** The schema mapping, the source dispositions, the transport-cache boundary, the selective-reading contract, the execution interlock, the baseline pins, the merge contract, the freeze order (`:83`), the post-run freeze (`:85`), the output review (`:87`) and the command are as signed in round 3. `test_cc2_delta.py:277–290` still finds both helper digests and every freeze-order phrase (Main: 289 tests OK).
- **One sentence was not updated:** `:99`. See CC2-R4-01.

## New issues

| ID | Severity | File and line | Concrete failure | Required fix |
|---|---|---|---|---|
| CC2-R4-01 | Minor | `preparation-reviews/identifiers-delta-cc2-REQUEST.md:99` (contradicts `:107`) | See below | See below |

**CC2-R4-01 — the stale self-pin pointer at REQUEST `:99`.**

*Failure.*
- `:99` still says the request's "final SHA-256 and byte count are recorded in `identifiers-delta-cc2-round-2-response.md`". That file's packet, labelled as the final confirming packet, gives the request as 20847 / `c32bd400…` (`:63`) and `preparation.py` as `7f8d390f…` (`:65`). Both are now wrong.
- `:107` names `core-impl-round-3-response.md` instead. The response's claim that "the request's external self-pin pointer now names this response" (`core-impl-round-3-response.md:28`) is true of only one of the two pointers.
- If Main follows the request's own instruction at `:99` when building the receipt, `review_receipt` refuses the stale row: `validate_pin` and the scope-block search at `preparation.py:456–460`. That refusal happens inside `review_gate`, which runs only after the action-d start record already exists (`cc2_delta.py:238` precedes `:244`). The result is a permanent refused action-d journal. The seal then cannot pass without an `operation-recovery` review (`preparation.py:621–631`).
- The REQUEST is one of the 16 bound files and is frozen from the run through the seal (`:85`). The contradiction would therefore become permanent in sealed evidence and could not be corrected later.
- Nothing wrong is sealed and no check is bypassed. The defect is a signed control that, if obeyed, produces an irreversible refused operation.

*Required fix.*
1. Edit `REQUEST.md:99` so that it names the same external self-pin record as `:107`. Preferably both should name one new record that will not go stale: e.g. `identifiers-delta-cc2-round-4-response.md` giving the REQUEST's final bytes and SHA-256, or wording that defers to "the external record named in the self-pin table row".
2. Do not edit any historical record: not `core-impl-round-3-response.md` and not the round-2 response.
3. Change nothing else in the REQUEST, apart from the helper pins if Core round 4 changes `preparation.py`.
4. Recompute the REQUEST pin and bring the full 16-pin packet to CC2 confirming round 5.

## Suggestions (not blocking)

1. **Pointer-consistency assertion.** A test could assert that the REQUEST's self-pin pointers agree, or at least that the request no longer names `identifiers-delta-cc2-round-2-response.md` as its pin location. Helper-digest presence checks alone cannot catch CC2-R4-01. `test_cc2_delta.py` is a bound control, so adding this changes its pin. It is optional.
2. **Diff evidence.** Core round 4 should record Main's diff of `7f8d390f…` against `9572aed3…`. That would close the limit of the byte-count accounting above.
3. **Validate the receipt before reserving the journal.** Carried from round 3, suggestion 5. Running `review_gate` before `Journal(...)` would make any receipt or pointer mistake refuse without leaving a failed action-d journal.
4. **Round-3 suggestions 1–4 still apply:**
   - freeze the receipt and this record after the run;
   - check every literal packet row against `pin()`;
   - add an end-to-end `run` → `collect_cc2_delta` test;
   - bind the producer journal's `producer_pins` to the reviewed controls.

## Peers and where they changed the recommendation

- **innovation-challenger (Cc2R4Challenger):**
  - Independently confirmed the lock-order trace (no recursive flock and no reversed order on the producer path), the producer/seal contract table, and the match between the helper pins and the table.
  - Found CC2-R4-01 independently and rated it an issue. It traced the refusal to `review_gate` running after the start journal.
  - Proposed the wording that defers to the table row.
- **experiment-peer (Cc2R4Peer):**
  - Refuted H1 with the exact 11 × 4 = 44-byte accounting and confirmed every round-3 seal citation at its current line.
  - Refuted H3 (deadlock) and H4 (test misses the defect).
  - Explained the REQUEST's +26 bytes by `:107` alone.
  - Found CC2-R4-01 independently, rated it Minor, noted that the response's claim about the pointer is inaccurate, and contributed Suggestion 2.
- **Council:** found the `:99`/`:107` contradiction independently, together with the byte accounting and the lock call-site inventory. Its first leaning, under the severity calibration, was to make `:99` a suggestion: the code refuses mismatches and nothing wrong can be sealed.
  - Both peers independently rated it an issue, on two grounds: the producer reserves the action-d start journal before `review_gate`, so following the request's own text leaves an irreversible refused operation; and the REQUEST is frozen into the seal, so the contradiction cannot be fixed later.
  - Those two facts moved the council to **Minor issue**.
  - This differs from round 3's sequencing condition. That residual failure required Main to *break* the signed order; this one follows from *obeying* the signed text.

## Verification

- **Static reading:**
  - `preparation.py:100–262, 300–443, 433–663`, plus grep of every lock call site
  - `cc2_delta.py:160–293`
  - `test_cc2_delta.py:269–300`
  - `test_journal_publication.py` (structure)
  - REQUEST `:75–118`
  - `core-impl-round-3.md` (grep)
  - `core-impl-round-3-response.md`
  - `identifiers-delta-cc2-round-3.md`
  - `identifiers-delta-cc2-round-2-response.md:62–66`
- **Main's reported results, not re-run by the council:**
  - `test_journal_publication.py`: 5 OK.
  - Full Python suite: 289 OK.
  - `go vet` and `go test -count=1` on `execution/model_free`: OK.
- **Before round 5, Main should:**
  - apply the CC2-R4-01 fix;
  - hold Core round 4 at the final `preparation.py`, with zero issues;
  - recompute `shasum -a 256` and `wc -c` for all 16 bound paths;
  - bring the packet, including the new REQUEST pin, to CC2 confirming round 5.
- **The receipt** may be created only after round 5 signs off, with rows generated by `verify_inputs.pin()` on the exact absolute paths in the REQUEST table.

No receipt, journal or CC2 producer run may use these bytes. The CC2 sign-off given in round 3 has lapsed, and this round grants none.

Verdict: 1 ISSUES
