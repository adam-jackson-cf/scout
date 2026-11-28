# X1 preparation (d) change-cost-v2 identifier delta — allowlist and implementation review — round 2

## Files reviewed (SHA-256)

The digests come from Main's packet (`local://x1-packet-digests.md`, made with `shasum -a 256`). I had no shell and did not recompute any of them. Paths are relative to `evaluation/x1-triage-ranking/`.

| File | SHA-256 | Role |
|---|---|---|
| `execution/identifiers/cc2_delta.py` | `657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6` | Under review (changed since round 1) |
| `execution/identifiers/cc2_readers.py` | `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b` | Under review (unchanged) |
| `execution/identifiers/cc2_schemas.py` | `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f` | Under review (unchanged) |
| `execution/tests/test_cc2_delta.py` | `ba96da085fe3cc1ad2d2433c199e0abedacebb6065f52ccea3d8be83441a20c2` | Under review (changed) |
| `preparation-reviews/identifiers-delta-cc2-REQUEST.md` | `288073ce3a47d952db958f33b8423bac8b4f4f0fce340ffd2f9a98ebb456ad0c` | Under review (changed) |
| `preparation-reviews/identifiers-delta-cc2-round-1-response.md` | `d11a9a4c11cc006cc6770e22d2112be8250db306fd1abeabc97fbb172bf5460d` | Response to round 1 |
| `execution/identifiers/projection.py` | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | Bound dependency (unchanged) |
| `execution/preparation.py` | `858031133e1d28ecf71cb9b4e79b64df9ad03f572c2ab4eca32d5fc48f019255` | Bound dependency. Changed by the concurrent Core round; checked only for its effect on the CC2 run |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | Bound dependency (unchanged) |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | Bound config (unchanged) |
| `preparation/exposure-sources.json` | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` | Signed source manifest |
| `preparation-run/journal/906389d7212a401b8c7f781d37960d84.json` | `6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c` (10,301 bytes) | Baseline journal. The digest is the implementer's `shasum` output in the response; it is not in Main's packet. The run checks it at runtime |
| Baseline start record `906389d7….start.json` | `c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738` (8,524 bytes) | Pinned inside the journal |
| Baseline outputs | `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d`, `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724`, `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04`, `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9`, `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30` | `manifest.json`, `source-coverage.json`, `delta-review.json`, `unresolved.json`, `familiarity.jsonl` |
| Context: `preparation-reviews/core-impl-round-1.md`, `core-impl-round-1-response.md` (`0d13e3356dcace31fa6d1c96db927d53347bcfbb7253b3a72c7a510d52f9c938`) | — | The Core change to `preparation.py` |

## Basis

- X1 round 9 `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`, preparation (a)–(e) only. The clock started 2026-11-09T03:22:58.701909Z.
- Signed inputs, unchanged:
  - PREREGISTRATION `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`
  - rev9-digests `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`
  - pins.json `6d556fb5…`
  - exposure-recipe `c065347e…`
  - exposure-sources `67dc5a4b…`
  - frozen-exposure-inputs `ec92fba6…`
- Round 1 accepted the ID-01 allowlist, the code-derived schema registry and the `raw/q` transport-cache boundary (CB-1). `cc2_readers.py` and `cc2_schemas.py` have not changed since, so those verdicts carry forward.

## Prior disposition

| ID | Status | Evidence |
|---|---|---|
| R1-01 (executed code not bound) | **Closed** | `cc2_delta.py:235–236` adds pins for `projection.py`, `preparation.py` and `verify_inputs.py`. These reach the journal's `source_pins` (`:238`) and the gate's required set (`:243–244`).<br>The `identifiers/` directory has no `__init__.py`, and the import closure is the 8 bound files (peer check). So everything `producer_pins()` records is also in the bound set.<br>The test that catches the old code is `test_cc2_delta.py:348–359`: it requires each helper in both the `source_pins` and every `review_gate` binding set. |
| R1-02 (refused run pins outputs it did not create) | **Closed** | `create_json` and the familiarity log record each file in `created` right after the exclusive open (`:176–182`, `:198–200`). The `except` block pins only that list (`:277–279`).<br>If `root.mkdir()` fails, `created` is empty, so the journal pins nothing. `:347` asserts that the refused rerun has `output_pins == []`; the old code gave 3.<br>Partial coverage output and a partial familiarity log are retained and pinned (`:233–246`, `:361–383`). |
| R1-03 (baseline journal digest missing) | **Closed** | `REQUEST.md:55–65` names the journal (10,301 bytes, `6aa8802a…`), its start record and the five outputs in order. The deferral sentence is gone (asserted at `test_cc2_delta.py:269–275`).<br>I read the journal metadata. It has action `d`, status `passed` and exit 0. Its `output_pins` (`journal:187–213`) match the request's table in bytes, digest and order, and its `start_record` (`:227–230`) is `c400e46e…`, 8,524 bytes.<br>I could not recompute the journal file's own digest. If it is wrong, `review_receipt`'s `validate_pin` and the gate's subset check refuse, so a wrong digest cannot pass. |

## Effect of `preparation.py` `858031…` on the CC2 run

The new bytes are compatible with how the producer calls them:
- **Imports and locks.** `Journal`, `load_config`, `operation_lock`, `paths` and `review_receipt` all still exist. `operation_lock` gained `read_only=False`, and its locks re-enter (`preparation.py:237–255`).
- **Journal construction.** `Journal.__init__`/`finish` signatures are unchanged. `producer_pins()` resolves the CC2 script entry. `prior_journals` pins the existing journals before the new start record is written (`:324–358`).
- **Receipt location.** The receipt sits under `preparation-reviews/`, not `preparation-run/`, so the study resolves to `EXEC.parent` and the record directory check passes (`:428–433`).
- **Refused runs.** Each refused CC2 journal needs an `operation-recovery` review that binds its start record, its terminal record and every output it pinned (`:543–553`). Before this change there was no recovery path at all.
- **Output accounting.** The three delta files are still accounted as `journal-output` (`:485–500`).

Two consequences are new:
1. **Receipt format.** `review_receipt` now needs, inside the council record:
   - the exact standalone line `**Zero issues. Signed off.**`;
   - exactly one line `Signed-off scope: identifier-delta-cc2-implementation`;
   - all 16 bound digests in full, in the block after that line (`:434–447`).
   
   This is met by writing the confirming record correctly, so it is a suggestion, not an issue (S2).
2. **Freezing order.** A CC2 journal written now would be the first journal to pin `preparation.py`. That makes R2-01 below.

## New issues

| ID | Severity | Finding | Concrete failure | Required fix |
|---|---|---|---|---|
| R2-01 | Moderate | **The authorisation has no ordering precondition, so the CC2 journal can freeze shared code that is still changing.**<br>The CC2 journal pins `preparation.py`, `verify_inputs.py`, `projection.py`, the test and the request in `source_pins` (`cc2_delta.py:235–238`), and every loaded EXEC module in `producer_pins` (`preparation.py:324–337,350`).<br>`collect_journals` runs `verify_tree_pins(value)` on every journal, passed or not (`:537`). `operation-recovery` applies only to journals that did not pass (`:546`), so a passed journal with stale pins has no recovery route.<br>No existing journal pins `preparation.py`: legacy journals have no `producer_pins`, and the baseline pins only `review_identifiers.py` and `projection.py` (`journal:215–226`).<br>Two pending changes will edit `preparation.py`. Core round 2 is still open. The delta seal integration the request requires (`REQUEST.md:79`) also needs it: `verify_scoped_evidence` reads fixed config keys (`:503–516`), and amendments may change only binding keys (`apply_binding_changes`, `:43–62`), so that integration has to change `preparation.py` [INFERENCE]. | Main runs CC2 at `858031…` once this round passes. Core round 2 or the delta seal integration then edits `preparation.py`.<br>From then on, `seal --check` and `seal` both raise `pin mismatch: …/preparation.py`. The create-once output root (`cc2_delta.py` `root.mkdir()`) cannot be re-run, and no review can waive the pins of a passed journal.<br>The preparation stays unsealable until another Core code change exists to allow it. | 1. Amend `REQUEST.md`'s "Main command" section: the producer runs only after the Core implementation review **and** the delta seal-integration review have both passed with zero issues, and their sign-offs name the same `preparation.py` and `verify_inputs.py` digests that the CC2 receipt binds. After the run, none of the bound files may change before the seal.<br>2. In the same edit, update `:79` to the final integrated `preparation.py` digest, and change `:53` to quote the exact verdict line and the scope-block rule (`preparation.py:434–447`).<br>3. Bring the final digests to a confirming CC2 round. That round must name all 16 bound digests in a `identifier-delta-cc2-implementation` scope block. CC2 code changes are needed only if the integration touches CC2 files. |

## Suggestions (not blocking)

1. **Correct the R1-01 catching-test claim.** `test_changed_imported_dependency_refuses_review_gate` (`test_cc2_delta.py:258–267`) builds its required list by hand, and `review_gate` did not change, so it passes on the pre-fix code too. The test that actually catches R1-01 is `:348–359`; the response should say so.
2. **Exercise the real receipt validator in a CC2 test.** `review_gate` is mocked (`:327`) and `review_receipt` is patched (`:249–256`). Add one synthetic end-to-end gate test that uses the real `review_receipt` with the new record format.
3. **Validate before reserving the journal** (round-1 suggestion 3, now with more evidence). Today, an existing output root is detected only after every table has been decoded, and a bad receipt only after the journal is reserved. Each such refusal costs an `operation-recovery` review. Run `baseline()` and `review_gate()`, and check that `root` does not exist, before constructing the `Journal`.
4. **Use `except BaseException`** at `cc2_delta.py:277`, as the seal and closure code do (`preparation.py:687`). An interrupt otherwise leaves a start record with no terminal record, which the seal refuses (`:524`), and no recovery path exists for that. The baseline producer has the same pattern, so this is hardening, not a regression.
5. **Carried forward from round 1:**
   - add a refusal test for a changed baseline output (`cc2_delta.py:163–164`);
   - put the remaining round-1 suggestions 4–8 into the output-review checklist.
6. **Build the receipt rows from `pin()`.** `review_gate` compares `(path, sha256)` pairs using the absolute path strings from `verify_inputs.pin` (`verify_inputs.py:25–28`). Main should generate each row by calling `pin()` on the exact paths the producer uses.

## Recommendation, alternatives, risks and execution outline

- **Recommendation.** Keep the allowlist and implementation as they are. Do not authorise the run yet. Close R2-01 by documenting the ordering, then hold a short confirming round at the final shared-code digests.
- **Alternatives considered:**
  - *Approve now and rely on Main's discipline.* Rejected: with the seal integration still pending, the failure becomes near-certain, and it cannot be undone.
  - *Approve now and refresh the receipt later.* Rejected: a new receipt cannot repair the pins of an immutable passed journal.
  - *Add a "superseded producer pins" recovery to the seal.* Rejected for now: it widens Core scope and would weaken a safeguard.
- **Risks:**
  - The added review sequencing costs time on the 45-day clock (deadline 2026-12-24T03:22:58.701909Z).
  - The baseline journal digest has not yet been confirmed by Main. It is checked at runtime, but the failure would come after the journal is reserved.
- **Execution outline:**
  1. Core round 2 reaches zero issues.
  2. Implement and review the delta seal integration, adding the CC2 kinds and output-review scope to the seal check.
  3. Main amends `REQUEST.md` as R2-01 requires and recomputes the 16 bound digests, the baseline journal included.
  4. Hold the confirming CC2 round.
  5. Main builds the receipt from `pin()` rows and makes the single canonical run.
  6. Hold the output review.

## Peers and where they changed the recommendation

- **experiment-peer (Cc2R2Peer):**
  - Its R2-02 matches my R2-01 and confirmed it independently, including the absence of a recovery route for passed journals and that the baseline journal does not pin `preparation.py`.
  - I folded its R2-01 (stale `:79` and `:53` wording) into the R2-01 fix and suggestion 2. On its own it is documentary, and a correctly formatted council record avoids it.
  - Its S1 became suggestion 1 and its S3 became suggestion 4.
- **innovation-challenger (Cc2R2Challenger):**
  - Agreed that R1-01 to R1-03 are closed, and raised the same ordering issue.
  - It added the deferred delta seal integration as a second certain editor of `preparation.py`, which raised R2-01 from "plausible" to "expected".
  - Its receipt-format point became suggestion 2.
- **Mine:** the compatibility trace of `858031…` against the producer, and the evidence from `apply_binding_changes` that the integration must change `preparation.py`.

## Verification

- This was a static, read-only review of the delta modules, the tests, the request, `preparation.py` `858031…` (`:1–760`), `verify_inputs.pin`, and the baseline journal metadata.
- No CC2 data payloads were opened. Nothing was executed, no digests were recomputed, and no model calls were made beyond the two read-only peer analyses.
- Main reports 273 Python tests OK and `go vet`/`go test` ok at these bytes; I did not re-run them.
- **Main should run after the fix:**
  - `shasum -a 256` and `wc -c` on the baseline journal, `preparation.py`, `verify_inputs.py` and the amended request;
  - the full Python suite under the pinned CPython 3.11.11.

## Verdict

**1 issue open** (R2-01, Moderate). R1-01, R1-02 and R1-03 are closed, and the ID-01 allowlist and its implementation stay correct. The delta producer must not run until a confirming round at the final `preparation.py` and `verify_inputs.py` digests reports no open issues and names all 16 bound digests in an `identifier-delta-cc2-implementation` scope block. No authorisation is given in this round.
