# X1 preparation model-free binding v2 — amendment and implementation review — round 4

**Scope and method.** This round checks the fix for round-3 finding X1B2-R3-01 and the files that fix changed. It also checks what the Core change to `execution/preparation.py` (now `b1759ec7…`) does to the binding track. The closure prerequisite P2 (`closure-impl-round-4.md`, the "P2. The frontend is final" section) asks this review to sign `execution/model_free/main.go` by its digest. The review was read-only. It ran no command, test, hash, adoption, attestation, rehearsal, acceptance, seal or model call. Three analyses fed it:

- **My own trace.**
  - The repair text: amendment `:62-78`; request `:1-12` and `:60-80`.
  - In `main.go`: `resolve`, `boundInputs`, `configProof`, `runtimeProof`, `effectiveRecord`, `reference` and `run` (`:212-222`, `:271-450`).
  - `main_test.go:245-296` and `:405-457`.
  - The golden file: the bound effective record and the bound receipt.
  - `test_binding_contract.py` in full.
  - `acceptance.py`: `run_revision`, `verify_bound_policy`, `verify_model_proof`, `execute_fixture`, `acceptance_flags` and `complete_operation` (`:77-296`).
  - `interlock.compare` (`:66-87`).
  - In `preparation.py`: `verify_output_accounting` and `collect_seal` (`:444-572`).
  - In `attest_closure.py`: `SOURCE_KEYS` (`:34-35`), `go_metadata` (`:410-431`) and the Go-review row check (`:561-562`).
- **`innovation-challenger` (R4Challenger):** whether the golden test is genuine, which Python predicates are not run on the golden, a sweep of the repair text, and a trace of Core output accounting.
- **`experiment-peer` (R4Peer):** a guard-deletion table, a key-by-key check of the golden against `acceptance.py`, confounds, consistency of the text, and Core blast radius.

**Verdict: zero issues.**

## Files and SHA-256

All digests are Main's (`shasum -a 256`, `local://x1-packet-digests.md`). The council did not recompute them. Paths are relative to `evaluation/x1-triage-ranking/`.

**Under review (signed below)**

| File | SHA-256 | Since round 3 |
|---|---|---|
| `execution/model_free/main.go` | `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` | unchanged |
| `execution/model_free/main_test.go` | `54e7bc66d4df750796ae4201b1dbeba172373a34dc46c3d4cce629d17e54d617` | changed (was `16e66424…`) |
| `execution/tests/model_free_contract.json` | `bc88c34ae1f8edc800e2c6a6d7eb86af4fcda205899deaba085630edab0a334d` | new |
| `execution/tests/test_binding_contract.py` | `dc817736d39551e01c62acb2e6b3dffeb861de825b4fb56c9b5c3198dc0a9efa` | new |
| `execution/binding/acceptance.py` | `2144999a8aece522e72a86a429a1628a2833ada1490fa6e0fe7600e152373be1` | unchanged |
| `execution/binding/interlock.py` | `19464e8940d0458be3606ed0438ae778fd01b48b73f8c2fee801856b80b2e6a1` | unchanged |
| `execution/binding/fixtures.py` | `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b` | unchanged |
| `execution/binding/fixtures/common/go.mod` | `ec75b6e87719a439d7ec4959302e25fd588319c7f7cfd1e6e20b39a324aff0b0` | unchanged |
| `execution/binding/fixtures/common/index.js` | `2fa9c2c5a47269dd22d1da68d24423b1a2d3a2da635c4c60798c59e17ade5d23` | unchanged |
| `execution/binding/fixtures/common/main.go` | `d37528f6027aa09adca50fa54e40ffc247c5597ffb7d4effa633ccba6134536c` | unchanged |
| `execution/binding/fixtures/common/package.json` | `514ef8299051029ada3268f1fb2eba8b698548d59bb6d438d71c42c03a627f9e` | unchanged |
| `execution/binding/fixtures/common/unchanged.py` | `534f740e38f9a8c5e7226a676a427bc5b44ec64ab605c33c595a68aedf5a97c4` | unchanged |
| `execution/binding/fixtures/head/.scout/policy.yaml` | `6189fea6ad039f0eb0206a7986aa94a44e3bdf46669184298b8a558d06fa7255` | unchanged |
| `execution/binding/fixtures/head/changed.py` | `0950bd4989dfc76b1c8ed8641c91b4994cefee6bc5bb6d5105510f99f25ec109` | unchanged |
| `execution/binding/fixtures/parent/.scout/policy.yaml` | `3d3dae7d063c4fe94127de4e67ae9315e1495e4fa510ea587ae4c5a058604ec3` | unchanged |
| `execution/binding/fixtures/parent/changed.py` | `0ecf4e2fd0b76f8305f0729807a67bc7d2eeabe27ed637a98bda8b9d87738e3b` | unchanged |
| `execution/tests/test_binding.py` | `2458e4289d67a1a061fcf43a2de6e42f1d586daee6cb4452f17f857ee1d36ab5` | unchanged |
| `execution/tests/test_binding_adoption.py` | `4b186aa1c672bd7353d014e056ae881cd0e5a05613130ff44a4791450ce0065d` | changed (S1) |
| `execution/tests/test_binding_boundaries.py` | `c5c1d042ee440e39eab66bb3bdb0d0d13b4756d4fe0c1018a16d7e63b7c2b177` | unchanged |
| `execution/tests/test_binding_seal.py` | `b21763f3a94b229722381a36ed0b2a2a6054bdd6c476e326445bdf76aad38450` | changed (S2, S5, S6) |
| `preparation/model-free-binding-amendment.md` | `27f1100122f9867974049a88dfae14587aa6a9db7d96ae8065faca45fca284f7` | changed (`:74`) |
| `preparation/model-free-binding-check-v2.json` | `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc` | unchanged |
| `execution/run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` | unchanged |
| `preparation-reviews/binding-v2-REVIEW_REQUEST.md` | `40780de7872a475f75ae594639da481722e4df32b78cfb3ae98773fb412eb4fc` | changed (`:12`, `:65`, `:69`, `:75`) |
| `preparation-reviews/binding-v2-round-3-response.md` | `44a391c3b8d7a1aff2c0743fca2af2e642356e675a7bb9e15caa3b41d2cfea0a` | new |
| `preparation-reviews/binding-v2-round-3-checks.json` | `7272f8d2c7e0256bc86e44211bf89b9aebc3578f7f4627b03ae94bbe7e379e1e` | new |

**Shared contract, read for blast radius**

| File | SHA-256 |
|---|---|
| `execution/preparation.py` | `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916` (changed since round 3, which read `599a128d…`) |
| `execution/tests/test_core.py` | `4f4bfdaf50fc0d56ff775f8c86bb6bc579c932acc1a76764a55ba4a6f36da07e` |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `preparation-reviews/binding-v2-round-3.md` | `d3fcde7cfec9a9f3e9126b3284903d0be6796f0f414e947fcf71d27f1354985b` (digest as cited in the round-3 response) |
| `preparation-reviews/closure-impl-round-4.md` | `5e9fcd1bcd1a28ae3487505df11d1bb95c378d6c2d1a33a059b85a9ffd349627` (digest as cited in round 3) |

## Basis (unchanged)

- Authority: X1 round 9, `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`, preparation actions (a)–(e) only.
- Clock started 2026-11-09T03:22:58.701909Z.
- PREREGISTRATION `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`.
- rev9-digests `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`.
- pins.json `6d556fb5…`.
- model-free-policy `87de3949…` (`87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f`).
- binding-check v1 `5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c`.
- runtime-attestation-rev6 `19463957…`.
- exposure-recipe `c065347e…`.
- exposure-sources `67dc5a4b…`.
- frozen-exposure-inputs `ec92fba6…`.
- `run-config.json` `c817f678…`.

## Prior disposition

| Finding | Council |
|---|---|
| **X1B2-R3-01 (Major)** | **Fixed as required.** |

**1. Text.** Amendment `:74` now states each required point:
- (b) must come before rehearsal.
- `main.go` and every file `closure.json` pins are frozen from the first `attest_closure.py` journal, preflight included.
- Repair covers only `execution/binding/**`, the source manifest, `binding/runtime.json` and the implementation review.
- A frontend defect found at rehearsal has no in-code recovery; it needs a reviewed recovery of (b), outside the amendment.
- Runtime-only frontend defects remain a residual risk.

The request says the same at `:12`, `:69` and `:75`. Request `:65` corrects the scope of round 2's lifecycle evidence to `binding/**` only. A sweep by the challenger and the peer found no remaining text that promises repair of `main.go` or of closure-pinned files. The phrase "only … frozen by that authority" at amendment `:64` refers to the adoption authority alone, so it does not conflict with the closure freeze.

**2. Pre-freeze contract test.**
- **Go side.** `main_test.go:416-444` builds all three effective records through the real `resolve`, `boundInputs` and `effectiveRecord`. Both receipts use the production `receipt` type, its JSON tags and the real `runtimeProof`.
  - The bound route reads the real signed `preparation/model-free-policy.json` through `frozenPolicy`.
  - `:448-456` round-trips through JSON and compares the whole structure with `reflect.DeepEqual`. Nothing is projected, and the golden is never regenerated. A missing golden file fails the test.
  - `main_test.go` is not in the closure's Go set. `go_metadata` selects files through `SOURCE_KEYS`, which has `GoFiles` but no `TestGoFiles` (`attest_closure.py:34-35`, `:425-431`), so the closure is unaffected.
  - Only `main.go` and `main_test.go` are left in `model_free/`; the generator was removed.
- **Python side.** `test_binding_contract.py` runs the production `verify_bound_policy`, `verify_model_proof(policy)` and `verify_model_proof(result, route=…)` on the golden, and they pass (`:30-35`).
  - Mutations refuse with the intended message: each of the six `model_proof` keys removed, on 2 routes × 2 collections (`:37-49`); `association_nil` flipped (`:51-57`); each lane reason changed (`:59-76`); and the reference `lane_report` set to `[]` (`:81-84`).
  - In the peer's deletion table, deleting each targeted guard turns its test red. Deleting `set(proof)!=fields` leaves a `KeyError`, which does not match the expected `InterlockFailure` message.
- **Key-by-key check.** Every field that `acceptance.py` and `interlock.compare` read from real effective records and receipts is present in the golden with the right name, type and nesting. This includes `original_policy.model_cli.provider` (read at `acceptance.py:95`), `mapping.Source` with a capital S, the lane reason literals, and every key `compare` reads.

**3. Order.** Amendment `:74` and request `:69` require both tests to run on `edb37ec2…` before the first `attest_closure.py` journal. The round-3 checks record shows them passing at these bytes.

**Round-3 suggestions.**
- **S1–S6: done.** S1 is covered by `test_duplicate_fourth_amendment_row_refuses_binding`. S2 is covered by `test_rehearsal_review_omitting_current_runtime_or_source_refuses`; both its runtime and source subtests go red if the `:364-366` guard is deleted or narrowed. S3, S4 and S6 are the text and comment changes. S5 uses the production scope `binding-acceptance`.
- **S7–S8: carried** as suggestions.

## Core `b1759ec7` blast radius on the binding track

No regression.

- Each binding run's `Journal('c')` pins only `pin(canonical)` (`acceptance.py:293`). That path is already in `outputs` with role `scout.x1.binding-acceptance.v2` before journal accounting runs.
- `verify_output_accounting` checks every journal output pin, but adds only paths not already referenced, with role `journal-output` (`preparation.py:444-452`). The canonical binding role is therefore kept, and no output is counted twice.
- The change adds no new refusal: journal `output_pins` were already checked by `verify_tree_pins(value)` (`:563`).
- The rehearsal creates no journal. Its report, review receipt and council record are still accounted for through canonical evidence.
- Tests:
  - `test_binding_seal.py` keeps its assertions on both binding roles.
  - Core adds `test_canonical_output_role_preserved_when_journal_also_pins_it` and `test_duplicate_journal_outputs_are_deduplicated_but_all_verified`.

## New issues

None.

## Suggestions (not blocking)

- **S1. Closure P1 before (b).**
  - `preparation.py` is now `b1759ec7…`, not the `599a128d…` that `closure-impl-round-4.md` P1 names. P1 therefore requires a short closure round confirming that dependency check before any `attest_closure.py` journal.
  - The response mentions only "any separate Core review". Name this closure round explicitly in Main's order.
  - If Core changes again before (b), rerun the binding lifecycle and contract tests, and hold a confirming binding round. `acceptance.py` imports Core.
- **S2. Make the pre-freeze run auditable.** Just before the preflight journal, write a checks artifact containing:
  - the argv, cwd, exit code and a timestamp;
  - the digests of `main.go`, `main_test.go`, the golden, `test_binding_contract.py` and `acceptance.py`;
  - the worktree commit.
- **S3.** Rerun `test_binding_contract.py` after any `binding/**` repair made after (b), before the fresh rehearsal.
- **S4. State the contract's limits in the request.**
  - The fixture copies the receipt assembly and empty-value normalisation from `run()` (`main.go:437`, `:443-448`) instead of calling it. These lines match at `edb37ec2…`.
  - The lane rows are hand-built: four rows with empty `mode`/`requires`. The real `LaneHealthOf` emits more rows with these fields set.
  - So `acceptance.py`'s check for unexpected model-capable lanes, `compare`, `acceptance_flags` and the positive-control evaluation are never run on Go-serialised values. The test is a serialization and proof-predicate contract, not end-to-end coverage of how receipts are produced.
  - **Optional alternative:** add synthetic findings and coverage to the golden, and run `compare` on them with refusals.
- **S5. Link the binary to its source.** Record in `binding/runtime.json`, or in the implementation review, the go build argv, toolchain and `main.go` digest used to build the frontend binary that rehearsal runs. This ties the executed binary to the source that was contract-tested.
- **S6 (carried, round 3 S7).** Keep the records of ordinary required reviews outside `OUT`.
- **S7 (carried, round 3 S8).** Django is never rehearsed, and stale inputs are found only after reservation. Both remain disclosed risks of the staged design.

## Verification

- **Main reports, at these bytes:** full `unittest discover` over `execution/tests` ran 237 tests, OK; `go vet ./evaluation/x1-triage-ranking/execution/...` clean; `go test -count=1 ./evaluation/x1-triage-ranking/execution/...` ok.
- **Round-3 checks record** (`7272f8d2…`): keeps the exact argv, exit codes and output, including the first failed run and its fixture-only correction.
- **Not exercised:** I ran nothing. Real frontend, engine and scanner integration, the trigger set and the rehearsal remain **unmeasured** until Main runs the authorized scratch rehearsal.

## Risks

- Every finding comes from reading code statically [INFERENCE]; nothing was reproduced.
- **Frontend defects that show only at runtime stay unrecoverable once (b) starts.** Examples: real lane rows or Clawpatch `mapping.Source`. The amendment discloses this risk; the contract test reduces it but does not remove it.
- **Adoption cannot be undone.** Compare the three receipt rows with `pin()` output first. Use the final writer at its canonical absolute path.
- The Clawpatch/Node tree is still trusted as a disclosed assumption.

## Execution outline (Main)

1. Hold the closure P1 confirming round for Core `b1759ec7…` (S1).
2. Run the Python and Go contract tests on `main.go` `edb37ec2…` and record the run (S2), before any `attest_closure.py` journal.
3. Write `OUT/reviews/closure-model-free.json` with scope `closure-model-free-go`, `record` = a pin of this record, and `reviewed` = `[pin(main.go)]` only.
4. Write the three-row `OUT/reviews/binding-v2-amendment.json`, checking each row against `pin()`.
5. Run (b): preflight, then canonical.
6. Run canonical `preparation.py adopt`.
7. Write the manifest, `binding/runtime.json` and `reviews/implementation.json`.
8. Run the scratch rehearsal at its scratch output, then write `OUT/reviews/binding-v2-rehearsal.json`.
9. Run canonical synthetic, then Django.
10. Write `reviews/binding.json`, finish the remaining tracks, then seal.

## Peer influence

- **Challenger.** Found no issues. It confirmed that the golden runs the real proof functions and that Core keeps canonical roles. It listed the canonical checks not run on the golden, which became S4, and offered the `compare` extension as an alternative.
- **Experiment-peer.** Found no issues. It supplied:
  - the deletion table, which confirmed that every new guard has a test that catches it;
  - the key-by-key agreement between the golden and `acceptance.py`;
  - confounds C1 and C2 (into S4) and C3 (into S5);
  - the closure P1 consequence of the Core digest change (S1), which I had noted but rated lower.
- My verdict did not change; the peers added S1–S5.

## Verdict

**zero issues — amendment v2 adopted for action (c); the binding may be rehearsed and then run at its canonical paths**

Per the signed staged design, canonical runs still need the authentic `model-free-binding-v2-rehearsal` review of a successful scratch rehearsal. Action (b) still needs Main's closure P1 confirming round for Core `b1759ec7…` (S1).

**Sign-off.** The innovation council signs, with zero open issues, the binding v2 amendment and implementation at these digests:

- `execution/model_free/main.go` `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078`. This satisfies closure-impl-round-4 P2: the binding record that `closure-model-free.json` pins.
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
- `execution/tests/test_binding_adoption.py` `4b186aa1c672bd7353d014e056ae881cd0e5a05613130ff44a4791450ce0065d`
- `execution/tests/test_binding_boundaries.py` `c5c1d042ee440e39eab66bb3bdb0d0d13b4756d4fe0c1018a16d7e63b7c2b177`
- `execution/tests/test_binding_seal.py` `b21763f3a94b229722381a36ed0b2a2a6054bdd6c476e326445bdf76aad38450`
- `preparation/model-free-binding-amendment.md` `27f1100122f9867974049a88dfae14587aa6a9db7d96ae8065faca45fca284f7`
- `preparation/model-free-binding-check-v2.json` `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc`
- `execution/run-config-amendments-proposed/001-binding-v2.json` `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121`
- `preparation-reviews/binding-v2-REVIEW_REQUEST.md` `40780de7872a475f75ae594639da481722e4df32b78cfb3ae98773fb412eb4fc`
- `preparation-reviews/binding-v2-round-3-response.md` `44a391c3b8d7a1aff2c0743fca2af2e642356e675a7bb9e15caa3b41d2cfea0a`
- `preparation-reviews/binding-v2-round-3-checks.json` `7272f8d2c7e0256bc86e44211bf89b9aebc3578f7f4627b03ae94bbe7e379e1e`

**Shared contract read and depended on:**
- `execution/preparation.py` `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916`
- `execution/tests/test_core.py` `4f4bfdaf50fc0d56ff775f8c86bb6bc579c932acc1a76764a55ba4a6f36da07e`
- `execution/verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`
- `execution/attest_closure.py` `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`

This sign-off does not give Core or the closure their own authority. Any change to a signed file, or to the shared Core it depends on, needs a new zero-issue round before the binding run continues.
