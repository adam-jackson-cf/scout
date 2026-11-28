# X1 preparation — binding v2 and runtime closure, confirming round at the final Core — round 2

**Scope.** This round closes X1BC-R1-01 from `binding-closure-confirm-round-1.md` (`c3eeb652…`, prefix as supplied by Main). That finding was that the closure-signed tests `test_closure_review.py` and `test_closure_round2.py` had changed after closure round 5 with no zero-issue round naming their current bytes, and that the current digests of the other closure-signed files had not been supplied. This round also:
- reviews the Core-owned parser test `test_preparation_evidence.py`, which no zero-issue record named;
- restates runbook rule RB-1;
- provides the receipt authority that `binding-v2-round-4.md` cannot provide.

Round 1's answers to Q1–Q4 are not reopened. Main reports that no file has been edited since round 1: the 21 digests in round 1's table are unchanged, and `preparation.py` is still `9572aed3…`.

**Method.** Read-only. The council ran no command, test, hash, Git command, adoption, attestation, rehearsal, acceptance, seal or model call, other than the two peer analyses below. All digests are Main's (`shasum -a 256`). Paths are relative to `evaluation/x1-triage-ranking/`. Three analyses fed this record:
- **My own reading:**
  - the full test lists of both closure files, compared with the recorded closure suite output in `closure-impl-round-3-response.md:521–589`;
  - the authority fixtures and every `patch.object` in both files;
  - `test_closure_round2.py:309–385`;
  - all of `test_preparation_evidence.py`.
- **`innovation-challenger` (R2Challenger):**
  - a per-file comparison of the test-name sets with `binding-v2-round-3-checks.json`;
  - checks that the authority tests reach production code and are not bypassed;
  - an audit of `test_preparation_evidence.py`, test by test.
- **`experiment-peer` (R2Peer):**
  - falsification checks O1–O5 against the hypothesis that the changes touch only the fixtures;
  - a guard-deletion table for `test_preparation_evidence.py`.

**Outcome: zero issues.** X1BC-R1-01 is resolved, and the binding v2 amendment and the closure frontend are signed off at the digests below.

## Reviewed digests (Main's)

**Supplied for X1BC-R1-01:**

| File | SHA-256 | Against the last closure sign-off (`closure-impl-round-5.md:197–202`) |
|---|---|---|
| `execution/closure/native.py` | `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb` | equal |
| `execution/tests/test_closure.py` | `3a1eae6082d0a9f42074e7ea5db8a2d74720808a91c2e2d5770bf44a82e599dd` | equal |
| `execution/tests/test_closure_review.py` | `534e4754f6100a69e60c7e12b6e7740a39fc84c8f0d81b094ab4b19f26cb2906` | changed (was `8762b45d…`); equals `core-impl-round-1-response.md:100` |
| `execution/tests/test_closure_round2.py` | `70efac053f0b92b5686455481bab1cd074874d96019091bb493f76cff465f274` | changed (was `7b22a59c…`); equals `core-impl-round-1-response.md:101` |
| `execution/tests/test_closure_round3.py` | `f6eec06e95a7227a02ffcd7ded9598957cffe2f840fe7976bde7e9e1e9e599e8` | equal |
| `execution/tests/test_preparation_evidence.py` | `634f5242d5d4190b5373bc741398a834cd3ee5296ec86a59b4748ce74b080aeb` | Core-owned; previously named only in `core-impl-round-2.md:23`, a round that closed with an issue open |

**Unchanged since round 1 (Main):**

| File | SHA-256 |
|---|---|
| `execution/preparation.py` | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` |
| `execution/model_free/main.go` | `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` |
| `execution/model_free/main_test.go` | `54e7bc66d4df750796ae4201b1dbeba172373a34dc46c3d4cce629d17e54d617` |
| `execution/tests/model_free_contract.json` | `bc88c34ae1f8edc800e2c6a6d7eb86af4fcda205899deaba085630edab0a334d` |
| `execution/tests/test_binding_contract.py` | `dc817736d39551e01c62acb2e6b3dffeb861de825b4fb56c9b5c3198dc0a9efa` |
| `execution/binding/acceptance.py` | `2144999a8aece522e72a86a429a1628a2833ada1490fa6e0fe7600e152373be1` |
| `execution/binding/interlock.py` | `19464e8940d0458be3606ed0438ae778fd01b48b73f8c2fee801856b80b2e6a1` |
| `execution/binding/fixtures.py` | `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b` |
| `execution/binding/fixtures/common/go.mod` | `ec75b6e87719a439d7ec4959302e25fd588319c7f7cfd1e6e20b39a324aff0b0` |
| `execution/binding/fixtures/common/index.js` | `2fa9c2c5a47269dd22d1da68d24423b1a2d3a2da635c4c60798c59e17ade5d23` |
| `execution/binding/fixtures/common/main.go` | `d37528f6027aa09adca50fa54e40ffc247c5597ffb7d4effa633ccba6134536c` |
| `execution/binding/fixtures/common/package.json` | `514ef8299051029ada3268f1fb2eba8b698548d59bb6d438d71c42c03a627f9e` |
| `execution/binding/fixtures/common/unchanged.py` | `534f740e38f9a8c5e7226a676a427bc5b44ec64ab605c33c595a68aedf5a97c4` |
| `execution/binding/fixtures/head/.scout/policy.yaml` | `6189fea6ad039f0eb0206a7986aa94a44e3bdf46669184298b8a558d06fa7255` |
| `execution/binding/fixtures/head/changed.py` | `0950bd4989dfc76b1c8ed8641c91b4994cefee6bc5bb6d5105510f99f25ec109` |
| `execution/binding/fixtures/parent/.scout/policy.yaml` | `3d3dae7d063c4fe94127de4e67ae9315e1495e4fa510ea587ae4c5a058604ec3` |
| `execution/binding/fixtures/parent/changed.py` | `0ecf4e2fd0b76f8305f0729807a67bc7d2eeabe27ed637a98bda8b9d87738e3b` |
| `execution/tests/test_binding.py` | `2458e4289d67a1a061fcf43a2de6e42f1d586daee6cb4452f17f857ee1d36ab5` |
| `execution/tests/test_binding_adoption.py` | `0396765fc3d51de609bf2d935d5b0b8979cd3b84ff65c5f12080e384912ce432` |
| `execution/tests/test_binding_boundaries.py` | `c5c1d042ee440e39eab66bb3bdb0d0d13b4756d4fe0c1018a16d7e63b7c2b177` |
| `execution/tests/test_binding_seal.py` | `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593` |
| `execution/tests/test_core.py` | `7bdc801955675437a095121b28cfbf2d85ca75c004abbe870f1e14ca5a1d2628` |
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` |
| `execution/fetch_objects.py` | `64bb3dd3b99d5227d0abd8e460f4dbabda18e15b37c16c71358241566e8c6a9f` |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `execution/run-config-amendments-proposed/001-binding-v2.json` (proposal) | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` |
| `preparation/model-free-binding-check-v2.json` (recipe) | `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc` |
| `preparation/model-free-binding-amendment.md` (amendment) | `27f1100122f9867974049a88dfae14587aa6a9db7d96ae8065faca45fca284f7` |

The nine `fixtures/**` digests are the ones signed in `binding-v2-round-4.md`, which Main reports unchanged.

**Main's reported checks at these bytes, not re-run here:**
- the full Python suite: 289 tests OK;
- `go vet` and `go test -count=1` on `execution/model_free`: pass.

## Disposition of X1BC-R1-01: resolved

### 1. The two changed closure tests

- **No test removed, renamed or added.**
  - `test_closure_review.py` defines the same 35 tests in `ClosureReviewTests` as the closure suite run recorded at `8762b45d…` (`closure-impl-round-3-response.md:521–555`; also `binding-v2-round-3-checks.json`):
    - r101 ×3, r102 ×2, r103, r104 ×3, r105, r106 ×2, r107 ×2, r108, r109 ×5, r110;
    - r111: recheck ×6, deferred ×5, Go ×3.
  - `test_closure_round2.py` defines the same 33 tests in `RoundTwoTests` as the run at `7b22a59c…` (`:556–589`): r201 ×8, r202 ×3, r203 ×8, r204 ×14.
- **The changed fixtures use the production receipt format and are not bypassed.**
  - `test_closure_review.py:178–182` and `test_closure_round2.py:109–118` write a council record under the study's `preparation-reviews/`. The record contains the exact verdict line, a scope marker for `closure-model-free-go` and the reviewed digest, under a receipt at `OUT/reviews/closure-model-free.json`.
  - Nothing in either file patches `review_receipt`, `reviewed_go_sources`, `validate_pin` or `preparation`. The only `patch.object` targets are platform and sys fields and the external-metadata seams that closure rounds 2–3 described: `main()`'s `load_config`, `read_clock`, `operation_lock`, `Journal` and `attest` (`round2:180`), and the producer seams at `round2:360–379`.
- **Each test still catches what the closure rounds said it catches:**
  - r107 frontend drift → `pin mismatch` (`test_closure_review.py:186–188`, from `validate_pin` on the reviewed row);
  - r203 wrong scope and extra key → `invalid scoped review` (`preparation.py:439–440`);
  - drifted record → `pin mismatch` (`:441`). If that guard were deleted, the text would fail the verdict check with a different message, so the test would still go red.
  - unrelated reviewed file → `only unique model_free` (`attest_closure.py:561–562`). The only needed adaptation is at `round2:154–156`: the fixture appends the unrelated digest to the record's block and re-pins the record. Without it, the new parser would refuse first, with "scope block does not name reviewed digest", and the test would no longer reach the `model_free` guard.
  - exact selected set → `exactly the selected`;
  - the journal binds both the receipt and the record (`round2:182–183`);
  - the producer's `inputs` carry both pins (`round2:383`).
- **The edits are where the fixtures are [INFERENCE].** In `test_closure_round2.py`, the r201 fixtures are at the lines closure round 3 cited (`:43–73`). The producer assertions moved by exactly two lines (from `:379–383` to `:381–385`), which matches the longer `authority()` fixture and the re-pin.
- **Limit:** no older copy of either file was found, and the council could not read Git objects. So "only the receipt fixtures changed" is supported by the evidence above and by the author's description (`core-impl-round-1-response.md:19`; "fixtures moved to the new receipt format", `core-impl-round-2.md:24`), but it is not a byte-level diff. Rule 5 is met regardless: this round reviewed the current bytes of both files in full against their claimed coverage and found no defect. S1 offers Main a decisive diff if wanted.

### 2. `test_preparation_evidence.py` (`634f5242…`)

- **Parser tests** (`:41–75`). Every one calls the real `review_receipt`:
  - exact approval;
  - real non-approval records, including `binding-v2-round-4.md`;
  - a negated verdict;
  - another scope;
  - a digest in another scope's block;
  - a record outside the review directories.
  
  The asserted messages map to separate branches (`preparation.py:445–460`). The peer's guard-deletion table shows each test turns red if its guard is deleted.
- **Scoping, recovery, provenance, journals and adoption** (`:77–276`). These cover:
  - cross-scope pooling (`verify_scoped_evidence`);
  - recovery, its coverage, and a deleted prior pair;
  - refusing a duplicate fetch before a journal starts;
  - the shared-file filter;
  - `seal --check` with no writes;
  - producer pins and the legacy deviation;
  - the r107 journal refusals ×8;
  - adoption refusals ×4;
  - preserving a first-seal refusal.
  
  Each asserts the specific refusal it claims. The patched seams are only `read_clock`, `documentary_gate`, `subprocess.check_output`, and `Journal`/`exclusive_json` used as call detectors. `collect_seal` is mocked only in the two tests about seal wrapper mechanics (`:179–187`, `:268–276`).
- **Integrated seal tests** (`:280–347`). They drive the binding seal fixture through the real `collect_seal`, checking predecessors, unaccounted output, the failed-producer check and clock refusals.
- **Dependencies.** The file borrows fixtures from `test_core.py` (`7bdc8019…`) and `test_binding_seal.py` / `test_binding_adoption.py`. All three are named in this round's sign-off.

### 3. The other closure-signed files

`closure/native.py`, `test_closure.py`, `test_closure_round3.py`, `attest_closure.py` and `run-config.json` equal closure round 5. Together with round 1's dependency analysis at `preparation.py` `9572aed3…` (round 1, Q1 and Q3), **closure prerequisite P1 is met.** This record also supplies P2, the closure scope block below for `main.go` `edb37ec2…`. P3 and P4 are still standing prerequisites for running the closure.

## Round 1, Q1–Q4: unchanged

No file changed, so round 1's findings hold as written:
- The Core changes keep the binding and closure contracts.
- The changed binding tests still catch what binding round 4 said they catch.
- `adopt`, closure preflight and canonical, the rehearsal and the seal have no code-level refusal beyond the authority-format requirement, which this record satisfies.
- The receipt rules in round 1, Q4 apply to receipts that cite this record.

## Runbook rule RB-1 (restated; in force with this sign-off)

RB-1 resolves Core round 4 S3, which Core round 5 required to be settled before canonical binding (c).
1. Run exactly one canonical `preparation.py adopt` of `execution/run-config-amendments-proposed/001-binding-v2.json`. Use the writer at its canonical absolute path, never under the `journal` wrapper.
2. Confirm that adoption succeeded before any `binding/acceptance.py` process starts, rehearsal included: `load_config` must show one adopted row whose `adoption_review` equals the pin of `OUT/reviews/binding-v2-amendment.json`.
3. Never run `adopt` while a binding CLI may be starting.
4. Make no further amendment once canonical (c) has begun.

Breaking RB-1 can leave an empty create-once `binding/*.json` behind a refused `Journal` (`acceptance.py:423`, `:436–439`; `preparation.py:359–360`), and the seal would then refuse.

## New issues

None.

## Suggestions (not blocking)

- **S1. Decisive diff (optional).** Find the committed blobs at `8762b45d…` and `7b22a59c…` with read-only Git, and run `git diff` against them. Expected hunks: only `test_closure_review.py` `:178–182` and `test_closure_round2.py` `:109–118`, `:150–156`. A hunk anywhere else would not invalidate this review of the current bytes, but it should be recorded.
- **S2. Parser hardening** (peer; already noted in `core-impl-round-2.md`). Add tests for:
  - a verdict line embedded in a longer line, because reverting `:449` to a substring check would leave every current test green;
  - a duplicated scope line;
  - an empty `reviewed` list;
  - a digest inside a longer hex run.
  
  All consumers fail closed, so this is hardening only, and adding the tests would change a bound test.
- **S3.** Rename `test_r101_real_nonapproval_records_refused`: `binding-v2-round-4.md` is an approval in a superseded format, not a non-approval.
- **S4 (carried from round 1).**
  - Pre-check both receipts with `review_receipt` and `reviewed_go_sources` before `adopt` and the preflight.
  - Never edit or delete this record, either receipt, the proposal or the adopted copy.
  - Add a closure-shaped journal-output fixture.
  - Use message regexes in the bare negatives of `test_binding_adoption.py`.

## Recommendation

Adopt this record as the authority for both scopes:
- **Binding amendment.** Write `OUT/reviews/binding-v2-amendment.json` as `{scope: "model-free-binding-v2-amendment", record: pin(this record), reviewed: [pin(proposal), pin(recipe), pin(amendment)]}`, with exactly those three rows at their canonical paths.
- **Closure frontend.** Write `OUT/reviews/closure-model-free.json` as `{scope: "closure-model-free-go", record: pin(this record), reviewed: [pin(<ROOT>/evaluation/x1-triage-ranking/execution/model_free/main.go)]}`, with no `main_test.go`.

Then follow round 1's execution outline, under RB-1.

## Alternatives

- **A. Sign off now (recommended).** The current bytes of every changed file were reviewed against their claimed coverage. Nothing else changed.
- **B. Wait for the byte-level diff (S1) first.** This adds a round, while rule 5 is already met by reviewing the current bytes.
- **C. Separate records per scope.** Equally valid for the parser, but it adds a record and a round for no gain.

## Evidence

- **Test files:**
  - `test_closure_review.py`: test definitions, `:174–188`;
  - `test_closure_round2.py`: `:109–193`, `:309–385`, all `patch.object` sites;
  - `test_preparation_evidence.py:20–348`.
- **Recorded suite output:** `closure-impl-round-3-response.md:521–589` and `:704–717` (the old digests); `binding-v2-round-3-checks.json`.
- **Records:**
  - `closure-impl-round-5.md:197–202`;
  - `core-impl-round-1-response.md:19`, `:91–101`;
  - `core-impl-round-2.md:19–24`;
  - `core-impl-round-4.md:129–143`;
  - `core-impl-round-5.md`;
  - `binding-closure-confirm-round-1.md`.
- **Code:** `preparation.py:437–461`, `:359–360`; `attest_closure.py:32–33`, `:554–564`; `acceptance.py:319–336`, `:423–439`.

## Risks

- **"Fixture-only" is [INFERENCE]** (see S1). That inference does not bear on this sign-off, which rests on the current bytes.
- **Irreversibility.** This record and both receipts are frozen from the first `attest_closure.py` journal or from adoption, whichever comes first. `main.go` is frozen from the first closure journal. `adopt`, every closure journal and every canonical binding reservation are create-once.
- **Residual risk:** frontend defects that show only at runtime (amendment `:74`).
- **Core cannot change:** CC2 journal `0378e156…` already pins it.
- **Clock:** 45-day deadline `2026-12-24T03:22:58.701909Z`.

## Execution outline

1. Main writes this record (LF line endings), computes its pin, and writes both receipts as stated under Recommendation. Pre-check them with `review_receipt` and `reviewed_go_sources`.
2. Run both contract tests on `main.go` `edb37ec2…` and record the run, before any `attest_closure.py` journal.
3. Run the closure all-roots dry check, one journalled preflight, then the single canonical `OUT/closure.json` (P3 and P4).
4. Under RB-1, run the canonical `adopt` and confirm it with `load_config` before any `acceptance.py` process.
5. Produce the manifest, `binding/runtime.json` and `reviews/implementation.json`, then the scratch rehearsal and its review.
6. Run canonical synthetic, then Django, then `reviews/binding.json`.
7. Finish the remaining reviews, including the CC2 output review. Run `seal --check`, then `seal`.

## How the peers shaped this record

- **experiment-peer:**
  - Ran O1–O5 against the fixture-only hypothesis. O1–O4 were refuted or not falsified; O5 was not falsified, with its limit stated.
  - Supplied the line-shift inference and the guard-deletion table for `test_preparation_evidence.py`.
  - Flagged that `test_core.py` `7bdc8019…` needs explicit naming; it is named below.
  - Supplied S1 and S2.
- **innovation-challenger:**
  - Compared the test-name sets independently and confirmed that nothing bypasses authority validation.
  - Confirmed that `round2:150–157` reaches the `model_free` guard.
  - Bounded the claim that the changes are fixture-only.
  - Supplied S3.
- **Mine:**
  - The comparison of test lists against the recorded run at the old digests.
  - The `patch.object` sweep.
  - The reading of the r204 orchestration test.
  - The restatement of RB-1 and the authority layout.
- **Neither peer found an issue, and none changed the verdict.**

## Sign-off

The innovation council reports zero issues in this confirming round and signs off, at these exact digests:

**Binding v2 amendment and implementation:**
- proposal `execution/run-config-amendments-proposed/001-binding-v2.json` `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121`
- recipe `preparation/model-free-binding-check-v2.json` `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc`
- amendment `preparation/model-free-binding-amendment.md` `27f1100122f9867974049a88dfae14587aa6a9db7d96ae8065faca45fca284f7`
- `execution/model_free/main.go` `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` (closure prerequisite P2)
- `execution/model_free/main_test.go` `54e7bc66d4df750796ae4201b1dbeba172373a34dc46c3d4cce629d17e54d617`
- `execution/tests/model_free_contract.json` `bc88c34ae1f8edc800e2c6a6d7eb86af4fcda205899deaba085630edab0a334d`
- `execution/tests/test_binding_contract.py` `dc817736d39551e01c62acb2e6b3dffeb861de825b4fb56c9b5c3198dc0a9efa`
- `execution/binding/acceptance.py` `2144999a8aece522e72a86a429a1628a2833ada1490fa6e0fe7600e152373be1`
- `execution/binding/interlock.py` `19464e8940d0458be3606ed0438ae778fd01b48b73f8c2fee801856b80b2e6a1`
- `execution/binding/fixtures.py` `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b`
- `execution/binding/fixtures/common/go.mod` `ec75b6e87719a439d7ec4959302e25fd588319c7f7cfd1e6e20b39a324aff0b0`
- `execution/binding/fixtures/common/index.js` `2fa9c2c5a47269dd22d1da68d24423b1a2d3a2da635c4c60798c59e17ade5d23`
- `execution/binding/fixtures/common/main.go` `d37528f6027aa09adca50fa54e40ffc247c5597ffb7d4effa633ccba6134536c`
- `execution/binding/fixtures/common/package.json` `514ef8299051029ada3268f1fb2eba8b698548d59bb6d438d71c42c03a627f9e`
- `execution/binding/fixtures/common/unchanged.py` `534f740e38f9a8c5e7226a676a427bc5b44ec64ab605c33c595a68aedf5a97c4`
- `execution/binding/fixtures/head/.scout/policy.yaml` `6189fea6ad039f0eb0206a7986aa94a44e3bdf46669184298b8a558d06fa7255`
- `execution/binding/fixtures/head/changed.py` `0950bd4989dfc76b1c8ed8641c91b4994cefee6bc5bb6d5105510f99f25ec109`
- `execution/binding/fixtures/parent/.scout/policy.yaml` `3d3dae7d063c4fe94127de4e67ae9315e1495e4fa510ea587ae4c5a058604ec3`
- `execution/binding/fixtures/parent/changed.py` `0ecf4e2fd0b76f8305f0729807a67bc7d2eeabe27ed637a98bda8b9d87738e3b`
- `execution/tests/test_binding.py` `2458e4289d67a1a061fcf43a2de6e42f1d586daee6cb4452f17f857ee1d36ab5`
- `execution/tests/test_binding_adoption.py` `0396765fc3d51de609bf2d935d5b0b8979cd3b84ff65c5f12080e384912ce432`
- `execution/tests/test_binding_boundaries.py` `c5c1d042ee440e39eab66bb3bdb0d0d13b4756d4fe0c1018a16d7e63b7c2b177`
- `execution/tests/test_binding_seal.py` `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593`

**Runtime closure (b), confirmed against the final Core (P1 met):**
- `execution/attest_closure.py` `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892`
- `execution/closure/native.py` `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb`
- `execution/tests/test_closure.py` `3a1eae6082d0a9f42074e7ea5db8a2d74720808a91c2e2d5770bf44a82e599dd`
- `execution/tests/test_closure_review.py` `534e4754f6100a69e60c7e12b6e7740a39fc84c8f0d81b094ab4b19f26cb2906`
- `execution/tests/test_closure_round2.py` `70efac053f0b92b5686455481bab1cd074874d96019091bb493f76cff465f274`
- `execution/tests/test_closure_round3.py` `f6eec06e95a7227a02ffcd7ded9598957cffe2f840fe7976bde7e9e1e9e599e8`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`

**Shared Core, reviewed and depended on.** Core's own authority remains `core-impl-round-5.md`.
- `execution/preparation.py` `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`
- `execution/verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`
- `execution/fetch_objects.py` `64bb3dd3b99d5227d0abd8e460f4dbabda18e15b37c16c71358241566e8c6a9f`
- `execution/tests/test_core.py` `7bdc801955675437a095121b28cfbf2d85ca75c004abbe870f1e14ca5a1d2628`
- `execution/tests/test_preparation_evidence.py` `634f5242d5d4190b5373bc741398a834cd3ee5296ec86a59b4748ce74b080aeb`

**Prior record:** `binding-closure-confirm-round-1.md` `c3eeb652…` (prefix as supplied by Main).

**Conditions:**
- The run is still gated by:
  - RB-1;
  - closure P3 and P4;
  - the amendment's staged order (the rehearsal and its review before canonical binding);
  - the remaining reviews before seal.
- Any change to a file named above needs a new zero-issue round before the run continues.
- This record, both receipts, the proposal and the adopted copy must not change from the first closure journal or adoption until seal.

**Zero issues. Signed off.**

Verdict: ZERO ISSUES — signed off

Signed-off scope: model-free-binding-v2-amendment

- proposal `execution/run-config-amendments-proposed/001-binding-v2.json` `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121`
- recipe `preparation/model-free-binding-check-v2.json` `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc`
- amendment `preparation/model-free-binding-amendment.md` `27f1100122f9867974049a88dfae14587aa6a9db7d96ae8065faca45fca284f7`

Signed-off scope: closure-model-free-go

- `execution/model_free/main.go` `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078`
