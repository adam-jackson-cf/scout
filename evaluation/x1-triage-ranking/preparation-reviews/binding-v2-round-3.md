# X1 preparation model-free binding v2 — amendment and implementation review — round 3

**Scope and method.** This round checks the fixes for round-2 findings X1B2-R2-01 to -03 and the files those fixes changed. It also answers Main's question: does an order exist from council sign-off through `preparation.py adopt`, rehearsal, canonical synthetic and Django to seal, and is that order tested? The closure prerequisite (`closure-impl-round-4.md:163`) asks this review to sign `model_free/main.go`, so its blast radius was read too.

The review was read-only. It ran no command, test, hash, adopt, rehearsal, acceptance, seal or model call. Three analyses fed it:
- **Own trace:** `acceptance.py` (lines 300–445), `preparation.py` (`verify_adoption`, `load_config`, `adopt`, `Journal`, `review_receipt`, `collect_seal`), `test_binding_seal.py` and `test_binding_adoption.py` in full, the `test_binding.py` rehearsal fixtures, the amendment's adoption/repair/staged lines (`:64`, `:74`, `:76`), and the request.
- **`innovation-challenger` (OrderAndSealChallenger):** order and seal trace, plus every patch in the lifecycle test.
- **`experiment-peer` (FixDeletionPeer):** for each guard, would deleting it turn a test red; plus the Go↔Python emission contract.

**Verdict: 1 issue open (X1B2-R3-01, Major).** All three round-2 findings are fixed as required, and each has a catching test. However, the repair procedure that round 2 itself prescribed, and that the amendment now states, does not work for `model_free/main.go`: the installed closure freezes that file before any rehearsal can run.

## Files and SHA-256

All digests are Main's (`shasum -a 256`); the council did not recompute them. Paths are relative to `evaluation/x1-triage-ranking/`.

**Under review**

| File | SHA-256 | Since round 2 |
|---|---|---|
| `execution/binding/acceptance.py` | `2144999a8aece522e72a86a429a1628a2833ada1490fa6e0fe7600e152373be1` | changed |
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
| `execution/model_free/main.go` | `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` | unchanged |
| `execution/model_free/main_test.go` | `16e66424f7bfef1a0faec66de9793c19e59018ffe8f78191af75524f9c271f70` | unchanged |
| `execution/tests/test_binding.py` | `2458e4289d67a1a061fcf43a2de6e42f1d586daee6cb4452f17f857ee1d36ab5` | changed |
| `execution/tests/test_binding_adoption.py` | `c08775b684218d660fbf761f5c364331bce2a120a8eb10d164de5a794dd98003` | changed |
| `execution/tests/test_binding_boundaries.py` | `c5c1d042ee440e39eab66bb3bdb0d0d13b4756d4fe0c1018a16d7e63b7c2b177` | unchanged |
| `execution/tests/test_binding_seal.py` | `c867fa0e60bd2c2491e95ed90822d0746801747d6a4480a95a21bd7cba3bfb7b` | new |
| `preparation/model-free-binding-check-v2.json` | `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc` | unchanged |
| `preparation/model-free-binding-amendment.md` | `634b352cbf893042a1e4b525c59492d5ae00f3bc4c8ce34e999065efb7054edf` | changed |
| `execution/run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` | unchanged |
| `preparation-reviews/binding-v2-REVIEW_REQUEST.md` | `f633f53f26e59e7c7bdb3e020f50d2d9a1384536cacfe8dcc48a317a6ac67575` | changed (round 2 reviewed `f221a010…`) |
| `preparation-reviews/binding-v2-round-2-response.md` | `8eb9ecf3c3a06177418ed82ca81b331d8d068fc61c9398dd5b4298a335f707e0` | new |
| `preparation-reviews/binding-v2-round-2-checks.json` | `07dc03e1c55c4a93b246d6ef60f7c5d3879b1b7321437f0df61a38d446d92c55` | new (digest from the request) |

**Shared contract, read for blast radius and unchanged**

| File | SHA-256 |
|---|---|
| `execution/preparation.py` | `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427` |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` (digest as cited in `closure-impl-round-4.md:22`; read only for X1B2-R3-01) |
| `preparation-reviews/binding-v2-round-2.md` | `29c1651f01bc5c87543f182de84c2008307e4e17dcb69ec9b3f60e4890327dc0` |
| `preparation-reviews/closure-impl-round-4.md` | `5e9fcd1bcd1a28ae3487505df11d1bb95c378d6c2d1a33a059b85a9ffd349627` |

**Basis unchanged:**
- Authority round 9 `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`.
- Clock started 2026-11-09T03:22:58.701909Z.
- Signed basis as listed in round 2.
- `OUT` has no closure, adoption or review files. Its journals are five (a) pairs, two (b) pairs and one (d) pair. Both (b) pairs are generic `preparation.py journal` runs with no `source_pins`, so no `attest_closure.py` journal exists yet.

## Round-2 disposition

| Finding | Council |
|---|---|
| **X1B2-R2-01 (Major)** | **Fixed as required for every binding source except `main.go`; the `main.go` remainder is X1B2-R3-01.** |
| **X1B2-R2-02 (Minor)** | **Fixed.** |
| **X1B2-R2-03 (Minor)** | **Fixed** up to `collect_seal`, the boundary round 2 asked for. |

**X1B2-R2-01 detail.**
- **Code.** `acceptance.py:334-336` requires the amendment review to bind exactly `{proposal, recipe, amendment}`, compared as (path, sha256), and exactly 3 rows.
- **Where sources are bound now:** only through the implementation-reviewed manifest (`:339-346`, `source_pins ⊆ manifest`) and the rehearsal review (`:364-366`).
- **Catching tests.** Each of these has no other refusal path that could make it pass:
  - extra row, missing recipe, missing amendment, and proposal-at-alternate-path (`test_binding_adoption.py`);
  - source missing from the manifest (matched by message).
  - Deleting the exact-set check turns all four `InterlockFailure` tests red. `InterlockFailure` is a `RuntimeError` subclass (`interlock.py`), so Core's `ValueError` refusals cannot satisfy those assertions.
- **Repair path.** `test_source_fix_after_real_adoption_accepts_fresh_implementation_review` and `test_post_rehearsal_source_fix_requires_fresh_review_before_reservation` show that, after a real adoption and a real rehearsal:
  - a changed binding source leaves `load_config` working;
  - the stale manifest refuses (`pin mismatch`);
  - the old rehearsal refuses;
  - fresh reviews reach canonical evidence and seal;
  - the adoption receipt and its record stay byte-identical.
- **Text.** Amendment `:64` and `:74` and request `:12` state the three-row receipt and the freeze.

**X1B2-R2-02 detail.** `verify_rehearsal` returns `record: pin(review['record']['path'])` (`acceptance.py:381`), and canonical synthetic stores it as `reviewed_rehearsal`. Seal then re-checks the record through `verify_tree_pins` of the required record and accounts for it through `sources` (`preparation.py:476`, `:550-558`). `test_rehearsal_record_mutation_invalidates_canonical_tree_and_seal` changes and deletes the real record after canonical acceptance, and requires both `verify_tree_pins(evidence)` and real `collect_seal` to refuse. If the record pin were removed, three things would turn red:
- the helper's key/equality asserts;
- the mutation checks;
- the positive seal, which would refuse the record placed under `OUT` as an unaccounted output.

**X1B2-R2-03 detail.** `test_binding_seal.py` runs, for real: `adopt` → `rehearse` → `verify_rehearsal` → `main` (synthetic) → `main` (Django) → `collect_seal`. The challenger listed every patch:
- **Permitted substitutions:** `generate`, `materialize`, `run_revision`, `git`, `documentary_gate`, `verify_installed_closure`, and `fetch_objects.Git` / `inspect_store` / `first_parent_boundary`.
- **Fixture relocation:** `controller.HERE` and `controller.EXEC`; the environment is cleared and `sys.argv` set for adopt.
- **Not mocked:** `adopt`, `load_config`, `review_receipt`, pins, clock, `Journal`, locks, `reviewed_inputs`, `verify_rehearsal` and `collect_seal`.
- **Assertions:** both v2 roles must be covered by the final review (removing either digest refuses with the expected message); Django revisions; the synthetic dependency; the clock bytes; and the authority receipt left unchanged.

**Prior suggestions.** Round 2 S1 (stale test count) is done. S2–S4 and round 1's S1–S7 remain suggestions.

## End-to-end order

**The order exists in code and is acyclic.** Its real-tree form is:
1. A council round with zero open issues that names `main.go`.
2. Main writes `OUT/reviews/closure-model-free.json` (closure-round-4 P2) and the three-row `OUT/reviews/binding-v2-amendment.json`.
3. Action (b): the `attest_closure.py` preflight and then the canonical `OUT/closure.json`. Rehearsal needs a passed, complete closure (`acceptance.py:347-350`) and `verify_installed_closure` on every launch (`:82`).
4. Canonical `preparation.py adopt`. It may run before or after (b), but not while any journal is non-terminal. The argv in amendment `:68` and the request match argparse.
5. Manifest (including the adopted copy), `binding/runtime.json` and `reviews/implementation.json`.
6. Scratch rehearsal, then `OUT/reviews/binding-v2-rehearsal.json`.
7. Canonical synthetic: `verify_rehearsal` runs before the lock.
8. Django: pins `synthetic.json`, checks the objects receipt and `^1`.
9. `reviews/binding.json`, then the remaining tracks, then seal.

The challenger confirmed:
- Relative `--config`/`--clock` paths resolve from the documented working directory.
- Rehearsal and canonical scratch paths do not collide.
- The rehearsal council record is accounted for whether it sits inside or outside `OUT`.

**Tested with real code: yes, through `collect_seal`.** Gaps:
- `seal()` itself (reservation, the (e) journal) is covered only by Core tests.
- The rehearsal is called through `rehearse()`, not `main --rehearsal`. Canonical `main` re-derives `source_pins`, so drift would still be caught.
- The closure re-check is patched, and the synthetic closure contains no Go files. This is why X1B2-R3-01 was not caught.
- In the test, journals a/b/d are written after adoption. Journals written before adoption are covered only by `test_core`.

## New issues

### X1B2-R3-01 — Major — amendment `:74`, request `:12`/`:75`; `acceptance.py:82, 347-350, 424`; `attest_closure.py:413-431, 566-579, 722-723, 733, 742, 830-831, 835-840`

**Finding.** `main.go` is one of the binding `source_pins` (`acceptance.py:424`). It is frozen before any rehearsal can run, through two paths:
- **Through `closure.json`.** `go_metadata` selects `model_free/*.go` through `selected_source`, which requires a reviewed pin and returns `pin(main.go)`. That pin enters `inventory.files` and then `original_runtime_files` (`:722-723`, `:742`). These pins are re-validated by `verify_installed_closure` before every `run_revision`, by `verify_tree_pins(closure)` in `reviewed_inputs`, and at seal. `closure.json` is create-once (`:830-831`).
- **Through the (b) journal.** Every `attest_closure.py` (b) journal, the preflight included (`:835-840`), pins `OUT/reviews/closure-model-free.json` and its record. That receipt binds `main.go`'s digest, and the closure re-validates it.

So `main.go` is frozen from the first `attest_closure` journal. Rehearsal cannot run before (b).

The amendment's new "Repair before canonical reservation" paragraph (`:74`) and request `:75` nevertheless state that any source defect found by the rehearsal can be repaired. That paragraph becomes frozen protocol text at adoption. Round 2's required fix prescribed this procedure and named "a `main.go` emission mismatch" as a case it should make repairable. That was the council's own error, not an implementation error.

**Failure.**
1. The first real rehearsal shows a frontend defect: a field missing from or misnamed in the effective record or receipt, or a mis-set `Config`/`Options`.
2. Main follows `:74`: a new council round, a fresh manifest and implementation review, and a fresh rehearsal.
3. The fresh rehearsal refuses at `verify_tree_pins(closure)` and `verify_installed_closure` with a pin mismatch.
4. Replacing `closure-model-free.json` breaks the (b) journal's pins at seal, and `closure.json` cannot be recreated.

The binding track, and with it the seal, is then lost on the 45-day clock, after a repair round the signed text promised would work. Nothing is accepted wrongly, because every gate fails closed. The only test of the Go-to-Python field contract today is the real rehearsal, and that runs after the freeze. Python tests use hand-built dictionaries; Go tests do not run the Python predicates.

**Required fix.** Change only binding-track files; `main.go` stays `edb37ec2…`, and Core and the closure are unchanged.
1. **Text: amendment `:74` and request `:12`/`:75`.**
   - Limit the repair procedure to `execution/binding/**`, the source manifest, `binding/runtime.json` and the implementation review.
   - State that `model_free/main.go`, like every file `closure.json` pins, is frozen from the first `attest_closure.py` journal, preflight included.
   - State that (b) must precede rehearsal.
   - State that a frontend defect found at rehearsal has no in-code recovery and needs a reviewed recovery of (b), outside this amendment.
   - Disclose this as a residual risk.
2. **Pre-freeze contract test.**
   - **Golden file.** Add a checked-in golden JSON under `execution/tests/` holding the bound, reference and synthetic-unbound effective records, and the bound and reference receipts.
   - **Go test.** A `main_test.go` test uses a fixed synthetic root and the existing fixtures (`boundFixture` / `disabledResult` or equivalent). It builds those records through the real `effectiveRecord` and `runtimeProof`/receipt path and requires equality with the golden file.
   - **Python test.** A test loads the same golden and requires `verify_bound_policy`, `verify_model_proof(policy)`, `verify_model_proof(result, route=…)` and the lane-reason checks to pass. A mutated copy (one `model_proof` key removed; `association_nil` flipped) must refuse.
   - **Closure impact.** `main_test.go` is outside the closure's reviewed Go set (`closure-impl-round-4.md:164`), so the closure is not affected.
3. **Main's order.** Main runs both tests on `main.go` `edb37ec2…` before the first `attest_closure.py` journal. The round-3 response states that the round-2 lifecycle test proves repair only for `binding/**`.

## Suggestions (not blocking)

- **S1.** Add a test with a duplicate fourth row in the amendment receipt, so the row-count guard (`acceptance.py:335`) is covered. Otherwise drop "duplicates refuse" from the response.
- **S2.** Isolate the rehearsal-review binding guard (`:364-366`): a review that omits the runtime or one source while the report's `sources` are current. Today the stale-rehearsal refusal in the lifecycle test comes from Core `review_receipt` (stale row), and its `assertRaises` matches no message.
- **S3. Repair mechanics.**
  - A fresh rehearsal needs a new `--output`, because `rehearse` refuses an existing report or `.work` directory, and the request hard-codes `report.json`.
  - `implementation.json`, the manifest and `binding-v2-rehearsal.json` are replaced at their fixed paths.
- **S4. Exact receipt rows.**
  - List the three receipt rows as the exact absolute strings that `pin()` produces.
  - `adopt` checks digests only (`preparation.py:278`), so a wrong path is adopted and then refused for good at `acceptance.py:335`.
  - Main should compare the receipt rows with `pin()` output before adopting.
- **S5.** Use the production scope `binding-acceptance` for the final binding review in `test_binding_seal.py`.
- **S6.** The test's `council/source-fix.json` is never read by code; the gate is the record of `implementation.json`. Say so in the test.
- **S7 (carried).** Keep the records of ordinary required reviews outside `OUT`. Core lists them only in `scoped_reviews`, so a record placed under `OUT` would be an unaccounted output.
- **S8 (carried; round-1 S2).** Django is never rehearsed. A Django-only defect found after canonical synthetic cannot be repaired, because `synthetic.json` pins the sources and Django re-checks them. Also, a stale manifest or closure is found only after reservation, inside `work()`. This is the cleared staged design; it is recorded as a risk, not re-raised.

## Recommendation

Do not adopt or start (b) yet. Apply X1B2-R3-01's text and contract-test fix; this changes `main_test.go`, a new test-data file, a Python test, the amendment and the request. Then hold round 4, which should be short. A clean round 4 lets Main write `closure-model-free.json` and proceed in the order above.

## Alternatives

- **A. Accept with text only.** Rejected: the field contract would still first meet a real run after the freeze.
- **B. Rehearse before (b) on an unattested runtime.** Rejected: it changes the signed staged-authority basis (the rehearsal is meant to run on the attested runtime) and adds a second way to validate the runtime.
- **C. Closure-round-3 S7: rehearse on a preflight closure.** Insufficient: the preflight journal also pins the Go review receipt (`attest_closure.py:835-840`).
- **D. A Core/closure change allowing the Go review to be re-issued.** Rejected: it reopens two signed tracks for a risk the contract test mostly removes.

## Risks

- All findings come from static reading [INFERENCE]; nothing was reproduced.
- Defects in the frontend that only show at runtime, with the real engine and scanners, remain unrecoverable after (b) even with the fix. The fix discloses this risk; it does not remove it.
- Adoption is irreversible. Run it only with the final signed writer at its canonical absolute path, and with a three-row receipt whose rows match `pin()` exactly (S4).
- Trust in the Clawpatch/Node tree remains a disclosed assumption.

## Execution outline

1. Edit amendment `:74` and request `:12`/`:75` as specified in X1B2-R3-01.
2. Add the golden file, the Go equality test and the Python predicate test, including the mutation refusals.
3. Optionally apply S1–S6.
4. Main runs the full unittest suite, `go vet` and `go test -count=1`, re-pins every file and submits round 4.
5. After a clean round:
   1. write `closure-model-free.json`;
   2. write the three-row receipt;
   3. run (b);
   4. run `adopt`;
   5. write the manifest, runtime and implementation review;
   6. run the rehearsal and write its review;
   7. run canonical synthetic, then Django;
   8. write `binding.json`;
   9. seal.

## Peer influence

- **Challenger.** Found no issues. It supplied the patch inventory and the order trace, and suggestions S1, S5 and S7. It also showed why the rehearsal record is safe under `OUT`.
- **Experiment-peer.** Found X1B2-R3-01, which changed my verdict: my trace had cleared R2-01 without linking the closure's Go pin to the new repair text. It also provided the guard-deletion table, the Go/Python key-by-key consistency check (no mismatch found) and S2–S4 and S6. I kept its Major grade, for consistency with R2-01: the consequence is the same unrecoverable loss of the preparation. I added that the preflight journal also freezes `main.go` (Alternative C). This narrows the fix to text plus a pre-freeze test.

## Verification

Main reports, at these bytes: 205 Python tests OK, `go vet` clean, `go test -count=1` ok. I ran nothing; what I read is listed under "Scope and method".

## Verdict

**1 issue open (X1B2-R3-01, Major).** No sign-off is given, and the clean-review wording is withheld; this record must not be used as a review record. This record also does not sign `model_free/main.go` (`edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078`), so closure-round-4 prerequisite P2 is still unmet. Amendment v2 stays proposed and unadopted. `preparation.py adopt`, (b), the rehearsal, canonical synthetic and Django acceptance, and seal must not run on the v2 path. The signed v1 recipe, `run-config.json` `c817f678…` and the clock remain in force.
