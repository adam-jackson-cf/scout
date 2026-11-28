# X1 preparation model-free binding v2 — amendment and implementation review — round 2

**Scope and method.** This round re-reviews the fixes for X1B2-R1-01 to -05 and for identifiers R7-01 (fix (a), in `binding/acceptance.py`), and the files those fixes touch. It also answers Main's question: does a complete order exist, and is it tested, from council sign-off through `preparation.py adopt`, rehearsal, canonical synthetic and Django to seal?

The review was read-only. It made no model calls beyond the two council peers, used no real inputs, and ran no adopt, rehearsal, acceptance, seal, test or hash. Three analyses fed it:
- my own trace of `acceptance.py`, `preparation.py` (adopt, load, review, journal, seal) and the three binding test files;
- `innovation-challenger` (EndToEndOrderChallenger), which traced the end-to-end order;
- `experiment-peer` (FixVerificationPeer), which checked each fix by asking whether deleting the guard would fail a test.

**Verdict: 3 issues open (1 Major, 2 Minor).** The circularity from round 1 is fixed. However, the way it was fixed means a defect found in the rehearsal could never be repaired, and the rehearsal authority is not re-checked at seal.

## Files and SHA-256

All digests are Main's (`local://x1-packet-digests.md`, `shasum -a 256`). The council did not recompute them. Paths are relative to `evaluation/x1-triage-ranking/`.

**Under review: amendment, binding implementation and tests**

| File | SHA-256 |
|---|---|
| `preparation/model-free-binding-check-v2.json` | `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc` |
| `preparation/model-free-binding-amendment.md` | `c8a8e418cdd99ca9c5fc691dd5d5c11766d94b4a888b956ce3fb0601ccf05362` |
| `execution/run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` |
| `execution/binding/acceptance.py` | `61a95ede2c423c50b1ae0df130160608f72632484f41f7705f3e46f6cceb6b9b` |
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
| `execution/model_free/main.go` | `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` |
| `execution/model_free/main_test.go` | `16e66424f7bfef1a0faec66de9793c19e59018ffe8f78191af75524f9c271f70` |
| `execution/tests/test_binding.py` | `e106e98770284c91532d99a8b07a5a61cb6c3fe4f0a3a4c2b3b2a257fcda73f2` |
| `execution/tests/test_binding_adoption.py` | `7e15dab3cc8b21eeaf2651308e74c0e906b5e1039782c127f89c51410b00a9c5` |
| `execution/tests/test_binding_boundaries.py` | `c5c1d042ee440e39eab66bb3bdb0d0d13b4756d4fe0c1018a16d7e63b7c2b177` |
| `preparation-reviews/binding-v2-REVIEW_REQUEST.md` | `f221a010118cd7b7c31f178396a45cafe4a6d8e6acfe390d041dc6591fed3c5d` |
| `preparation-reviews/binding-v2-round-1-response.md` | `e125096359b392d7e0d7ce481dfdf171c0b12afe2480d015faa849858c6e0c9a` |
| `preparation-reviews/binding-v2-round-1-checks.json` | `8d567ea225c388c19bfb6e82c7e93076f2f5c27e1f26572cf8b2ff3a24ce10e4` |

The globbed file set under `execution/binding/` is exactly the 12 files above. This is the `source_pins` set (`acceptance.py:424`), plus `model_free/main.go`.

**Shared contract, read for this review's blast radius, and unchanged since earlier rounds**

| File | SHA-256 |
|---|---|
| `execution/preparation.py` | `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427` |
| `execution/tests/test_core.py` | `94afb7638acbb0316cc019077a87d924fda0e665b752d3d57acdfadbd9657ace` |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` (closure track; reached only through `verify_installed_closure`) |
| `preparation-reviews/binding-v2-round-1.md` | `91105b4598f4d971ee78933a6578b5560476b966b27260109adc15fd4b0eabcd` |
| `preparation-reviews/identifiers-impl-round-7.md` | `6d279752335d4982754fd5639f729ecc008a23eacd66fa1207268caf1706b01f` |
| `preparation-reviews/closure-impl-round-3.md` / `-response.md` | `32b4b78e0f73887e83c0a1cf1e33c5efc7c5cbf3805d01b8b96f11637adc9ab3` / `f82335907b2efecfcb0241da0bd9006ffc4c17527cbfe067f31774798da9c982` |

The other packet files (`closure/native.py` `c264d61a…`, `fetch_objects.py` `737b5ec7…`, `identifiers/projection.py` `2408ed45…`, `review_identifiers.py` `adc42d38…`, closure and identifier tests) belong to other tracks. They were not reviewed here.

## Basis

- **Authority.** Round 9 (`6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`) authorises preparation (a)–(e).
- **Signed basis, unchanged:**
  - PREREGISTRATION `abd206c3…0fdb`
  - rev9-digests `2253911b…c2da`
  - pins.json `6d556fb5…00fc`
  - model-free-policy `87de3949…245f9f`
  - binding-check v1 `5056ae74…aedc59c`
  - runtime-attestation-rev6 `19463957…62d62`
  - exposure-recipe `c065347e…`, exposure-sources `67dc5a4b…`, frozen-exposure-inputs `ec92fba6…`
- **Clock.** Started 2026-11-09T03:22:58.701909Z. This change does not touch it.
- **Severity.** Calibrated as in the request. R-064: one recommendation per issue, no choices posed.
- **Current state of `OUT`.** It holds only `clock.json`, the input receipt, seven terminal journal pairs and the object receipts. There is no closure, runtime, review or adoption file yet, and `run-config-amendments/` is empty.

## Prior-round disposition

| Finding | Council |
|---|---|
| **X1B2-R1-01 (Major) / R7-01** | **Circularity fixed; the fix introduces X1B2-R2-01.** See detail below. |
| **X1B2-R1-02** | **Verified.** The recipe's `residual_limitation` and the amendment's "What the proof answers" section name the three required points (detail below). No "detector is live" overstatement remains. |
| **X1B2-R1-03** | **Verified, and the tests are not vacuous.** For each boundary, deleting or weakening the guard turns a test red (detail below). |
| **X1B2-R1-04** | **Verified.** Detail below. |
| **X1B2-R1-05** | **Verified on Main's current digests**, which match the response's re-pin for every binding, fixture, Core and test row. The request still cites 182 tests; Main's 198 includes the concurrent closure-round-3 tests (see S1). |

**R1-01 / R7-01 detail.** The fixed code:
- requires the pre-adoption review to bind exactly `[pin(recipe), pin(amendment), proposal, *source_pins]` (`acceptance.py:334`);
- requires the adopted SHA to equal the proposal SHA, and the exact adopted row with `adoption_review == pin(OUT/reviews/binding-v2-amendment.json)` (`:321-329`);
- binds the source manifest only through the post-adoption `implementation.json`, and checks `source_pins ⊆ manifest` (`:339-346`).

`test_binding_adoption.py` runs the real `adopt` → `load_config` → `reviewed_inputs` chain. It patches only `acceptance.EXEC`, the environment and `sys.argv`. `adopt` takes its paths from `config_path.parent` (`preparation.py:266,281`), so those patches do not reach adopt's paths, and the real writer pin is recorded and re-checked (`:94-95`, `:312`). On the real tree, paths agree: OUT equals `STUDY/preparation-run`, and there are no symlinks under `/Users/adamjackson/Projects/work` (the objects receipt's `database_realpath` and the clock's `cwd` are resolved paths that match the configured ones).

**R1-02 detail.** Both texts state that:
- the stand-ins do not reach scanner descendants;
- the unsandboxed frontend/Node/Clawpatch/git/go-list tree, which has network access, is trusted through the audit, the closure and the stand-ins;
- liveness is shown only at the contextual seam, and the Clawpatch providers are excluded by layers 1–2.

**R1-03 detail.** Each boundary and its catching test:
- **Key presence including `''`** at `reviewed_inputs`, at rehearsal entry before `open('x')`, and at every launch for 3 routes × resolve/full.
- **Provider order and duplication**, `model_attempts>0` (the case keeps `expected_interlock_failure=True`) and stub exit `97`.
- **`or resolve_only`** in the exit check, and **`and not resolve_only`** in the positive flag.
- **Django tip and `^1` check**: the valid case asserts the exact argv.
- **Re-derived synthetic flags** and **equal synthetic parent/head sets**.

**R1-04 detail.** `synthetic_receipt` is pinned in both the manifest and the top-level Django evidence (`:179,191,242-243`). The test then changes `synthetic.json` and expects `verify_tree_pins` to raise (`test_binding_boundaries.py:244-260`).

**Round-1 suggestions S1–S7.** Not implemented, as the response says. They remain suggestions, except where an issue below now covers them.

## Is there a complete end-to-end order, and is it tested?

**The order exists in code.** It is now acyclic:

| Step | Writer | Bound by | Satisfiable |
|---|---|---|---|
| Council round N with zero issues → `OUT/reviews/binding-v2-amendment.json` | Main | Its own `record`. `adopt` re-checks every reviewed pin. | Yes: proposal, recipe, amendment and sources all exist. |
| `preparation.py adopt` (canonical absolute path, `-I -B`) | `adopt` (`preparation.py:260-318`) | Adoption receipt and command record, writer/config/clock pins, prior journals | Yes. The adopt argv in amendment `:69` and request `:72` matches argparse (`:617-620`). |
| Source manifest, `binding/runtime.json`, `reviews/implementation.json` | Main, after adoption | `reviewed_inputs:339-355` | Yes. The manifest includes the adopted copy, as seal requires (`:512-517`). |
| Scratch rehearsal | `rehearse` (`:383-414`) | No Journal. Writes under SCRATCH only. | Yes. `SCRATCH/binding-v2-rehearsal/…` does not collide with `SCRATCH/synthetic`. |
| `reviews/binding-v2-rehearsal.json` → canonical synthetic | Main, then `main` (`:417-449`) | `verify_rehearsal` runs before the lock | Yes, but see X1B2-R2-02. |
| Django | `main` | Pins the synthetic receipt; checks objects and `^1` | Yes |
| `reviews/binding.json` → seal | Main, then `seal` | Record kind v2, v2 acceptance keys, `required_reviewed` role v2, revisions, `runtime_closure`, output accounting | Yes, by static trace. |

The relative `--config`/`--clock` paths in the request's commands resolve to the same absolute pins that seal compares (`pin` uses `.absolute()`; `read_clock` compares absolute paths). The R-087 ledger append does not trip the documentary gate, because shared ledgers are checked by owned extracts (`verify_inputs.py:17,63-90`).

**The order is not tested end to end.** Real code runs only from adopt to `reviewed_inputs`. After that:
- the rehearsal and rehearsal-review tests mock `load_config`, `review_receipt` and execution (`test_binding.py:291-295,349-352`);
- the Django tests mock `git`, `materialize` and `run_revision`;
- no test reaches `collect_seal` with v2 binding records after a real adoption (`test_core.py` reaches seal only on refusal paths).

This gap is why X1B2-R2-01 and X1B2-R2-02 were not caught (see X1B2-R2-03).

## New issues

### X1B2-R2-01 — Major — `acceptance.py:334` with `preparation.py:129-130, 394-405, 79-81`; amendment `:64-72`; request `:12, :75`

**Finding.** Fix (a) lists every `source_pins` entry (all 12 `binding/**` files plus `model_free/main.go`) in the **adoption authority** receipt. After adoption:
- every `load_config` calls `review_receipt` on that receipt, and `review_receipt` runs `validate_pin` on every reviewed row;
- the adoption receipt pins the review receipt's bytes (`verify_adoption`, `:79-81`).

So from adoption onwards, any change to a binding source makes `load_config` refuse. That blocks every preparation action: identifiers (d), closure (b), binding and seal. There is no way to re-adopt or to amend the receipt, and a second amendment does not help, because 001 is re-checked on every load. Rehearsal can only run after adoption (`:317` needs the adopted `binding_recipe`, and the amendment orders adopt → manifest → rehearsal).

**Failure.** The scratch rehearsal is the first real measurement of trigger sets, lane rows and scanner health; rounds 1 and 2 both record these as unmeasured. Suppose it reveals a defect that needs a source change, for example:
- an unexpected lane reason (`:156-166`);
- equal synthetic sets (`:233-234`);
- a `main.go` emission mismatch.

The fix then makes `load_config` raise `pin mismatch` for every action from then on. The preparation can then never seal while the 45-day clock runs. The staged rehearsal exists to protect irreversible state, but here the irreversible step (adoption) comes first and freezes the code the rehearsal tests. This is new evidence: round 1 required these pins without tracing how they interact with load-time re-checking and the post-adoption rehearsal.

**Required fix.** Change only binding-track files; Core is unchanged.

1. In `reviewed_inputs`, require the amendment receipt's `reviewed` set to **equal exactly** `{proposal, recipe, amendment}`. Exact equality stops the trap from being reintroduced by listing extra files.
2. Leave sources to the two post-adoption reviews that already bind them:
   - `implementation.json`'s manifest, with `source_pins ⊆ manifest` (`:345-346`);
   - the rehearsal review (`:364`).
3. **Amendment and request text.**
   - The pre-adoption receipt lists exactly the proposal, the recipe and the amendment. The council record still names every reviewed digest.
   - From adoption until seal, the proposal, recipe, amendment and that record are frozen.
   - A source fix after the rehearsal needs: a new council round → a regenerated manifest and `implementation.json`/runtime review → a fresh rehearsal and rehearsal review. All of this happens before any canonical reservation.
4. **Tests.**
   - Replace `test_one_source_pin_missing_from_amendment_review_refuses_binding` with three tests:
     - a source pin missing from the implementation-reviewed manifest refuses;
     - an amendment receipt with an extra row refuses;
     - the recipe, or the amendment, missing from the receipt refuses (today, deleting either from `:334` leaves the suite green).
   - Add a real-adoption regression: change a binding source after adoption, regenerate the manifest and implementation review, then check that `load_config` succeeds and `reviewed_inputs` passes with the new pins.

### X1B2-R2-02 — Minor — `acceptance.py:359-380, 443`; `preparation.py:380-391, 476, 521-526, 550-558`

**Finding.** `verify_rehearsal` validates the rehearsal review, including its council record (via `review_receipt`). It returns only `{report, review}`. Seal's `verify_tree_pins` checks those two pins but does not open the review JSON to follow its `record` pin. The rehearsal review is not in `required_reviews`.

By contrast, the amendment-review record is re-checked at seal (`authority_record`, `verify_amendment_reviews`). Closure round 2's R2-03 required the same treatment for the closure review: pin both the receipt and its record.

**Failure.**
1. After canonical synthetic acceptance, the rehearsal council record is edited or deleted, with the review JSON unchanged. Seal still passes, but the sealed evidence names a rehearsal authority that no longer verifies. That contradicts the recipe's sealed-evidence requirement: "passing scoped review records; no JSON approval flag substitutes" (`model-free-binding-check-v2.json` `required_sealed_evidence[2]`).
2. If that record is placed under `OUT/`, canonical acceptance accepts it, but seal later refuses it as an "unaccounted run output". Canonical receipts are create-once, so there is no recovery in code.

**Required fix.**
- Return `record: pin(review['record']['path'])` from `verify_rehearsal`, so it is stored in `reviewed_rehearsal`. Seal then re-checks it and accounts for it.
- Add a test with a real receipt: mutate the record, then require `verify_tree_pins` on the canonical evidence to refuse.

### X1B2-R2-03 — Minor — `tests/test_binding_adoption.py` (ends at `reviewed_inputs`); no positive v2 seal test

**Finding.** As shown above, real code is exercised only from adopt to `reviewed_inputs`. The later steps are tested with their authority checks mocked, and `collect_seal`'s handling of v2 binding records (kind, keys, `required_reviewed` role, revisions, output accounting) is never exercised positively after a real adoption.

**Failure.** Defects at these handoffs leave the suite green. X1B2-R2-02 is one; the R2-01 fix also changes which reviews bind sources. Such defects surface only at seal, after an irreversible adoption and create-once canonical receipts.

**Required fix.** Extend the integration test, keeping real authority, pin, journal and seal logic:
- real `adopt` → real `rehearse` → real `verify_rehearsal` → canonical synthetic `main` → Django `main` → `collect_seal`;
- substitute only scanner execution (`run_revision`/`materialize`/`generate`/`git`) and the non-binding seal prerequisites (documentary gate, object-store inspection, `verify_installed_closure`);
- assert that seal accepts both v2 records;
- include R2-01's post-rehearsal source-fix path and R2-02's record mutation as refusal cases.

## Suggestions (not blocking)

- **S1.** Update the request's evidence to Main's 198-test run at the current bytes. It still cites 182 tests and `round-1-checks.json`.
- **S2.** In `test_loaded_configuration_requires_exact_adopted_row`, two of the three cases are refused by Core before the binding gate. Assert the specific message so each case shows which gate refused.
- **S3.** `fixture_manifest` reads `synthetic.json` (`:178`) and pins it separately (`:179`). Pin from the bytes already read, to close a race that is negligible under the lock.
- **S4.** Keep every council record outside `OUT/`. Seal accounts for none of the records of required reviews (`preparation.py:521-526` puts them only in `scoped_reviews`).
- **Carried forward.** Round 1's S1–S7 still stand. Round 1's S2 (canonical prerequisites before reservation) remains relevant to Django.

## Recommendation

Do not adopt yet. Apply the three fixes in binding-track files only:
- `acceptance.py`;
- the binding tests;
- the amendment and request text.

None of these changes `preparation.py`, `test_core.py`, `run-config.json`, the proposal or the identifiers-round-7 digests. The proposal's bytes stay `88eb95a5…`. Then hold round 3. If it reports zero issues, adoption can proceed in the order traced above.

## Alternatives

- **Rehearse before adoption.** `rehearse` would apply the proposal's changes to a deep copy. Defects would surface before the irreversible step, but this means more new code, a pre-adoption manifest that seal would treat as stale, and a second configuration path. Rejected in favour of R2-01's fix.
- **Accept the freeze in text only.** State that a failed rehearsal ends in reviewed recovery that blocks every track. Rejected: it accepts a plausible, unrecoverable loss of the 45-day preparation for no gain in safety.
- **Add a post-adoption binding-authorisation receipt (round 1's alternative B).** This changes Core and reopens the identifiers track.

## Evidence

- Load-time freeze: `preparation.py:129-130` (`review_receipt` on the adopted authority), `:398-404` (`validate_pin` on every reviewed row), `:79-81` (adoption pins the review receipt), with `acceptance.py:317,334`.
- Rehearsal provenance: `acceptance.py:361-364, 380, 443`; seal `preparation.py:476, 521-526, 550-558`. Precedent: `closure-impl-round-2.md:60,115`.
- Order trace: `adopt` `:260-318`; `Journal` `:324-339`; `collect_seal` `:459-560`; `run-config.json` `required_records`/`required_reviews`; documented commands in request `:72, :80, :86-87`.
- Path identity: the objects receipt's `database_realpath` and the clock's `cwd` are resolved paths that equal the configured ones.

## Risks

- Every finding comes from static reading [INFERENCE]; nothing was reproduced. Line numbers agree across my reading and both peers.
- Adoption is still irreversible. Adopt only with the final signed-off `preparation.py`, invoked at its canonical absolute path (round 1 S7).
- Fixture triggers and lane rows remain unmeasured until the rehearsal runs. R2-01 makes acting on that measurement possible; it does not make the measurement itself any more likely to pass.
- Trust in the Clawpatch/Node tree remains a disclosed assumption.

## Execution outline

1. R2-01:
   - make the amendment-receipt check an exact set of `{proposal, recipe, amendment}`;
   - rewrite amendment `:64-72` and request `:12, :75`;
   - replace and add the tests listed under R2-01.
2. R2-02: add the `record` pin to `verify_rehearsal`'s return value, with a mutation test.
3. R2-03: extend the integration test from real adopt through `collect_seal`, as specified above.
4. Main runs the full unittest suite, `go vet` and `go test`, re-pins every file, attaches the output, and submits round 3.
5. After a zero-issue round, Main proceeds in order:
   1. materialise the `{scope, record, reviewed}` receipt with exactly the three rows;
   2. run `preparation.py adopt` (canonical command);
   3. generate the manifest, runtime and implementation review;
   4. run the rehearsal and its review;
   5. run canonical synthetic, then Django;
   6. produce `reviews/binding.json`, then seal.

## Peer influence

- **Challenger.** Found X1B2-R2-02 and supplied the order trace. Its seal-placement failure mode turned out to apply to every review record, so it became S4; I kept the seal-time re-check as the issue, on closure R2-03 precedent. Its observation that real adoption is never followed through to seal became R2-03.
- **Experiment-peer.** Found the load-time freeze (its N1). My own trace had verified the circularity fix but missed the freeze, so this changed my verdict. It offered three fixes; following R-064 I chose the exact-set binding-only change. Its deletion analysis confirmed R1-03. Its S2 (recipe/amendment negative cases) is folded into R2-01's tests, and its S1/S3 are kept as suggestions.

## Verification

**What I read:**
- `acceptance.py` in full;
- in `preparation.py`: lines 36–320 and 320–662 (adopt, `verify_adoption`, `load_config`, `review_receipt`, `Journal`, `collect_seal`, `seal`, `main`);
- `test_binding_adoption.py` and `test_binding_boundaries.py` in full;
- the recipe's `residual_limitation`, `scratch_rehearsal` and `required_sealed_evidence`;
- the amendment, the request's order and command lines, and the proposal;
- `run-config.json` paths, records and reviews;
- the shared-file handling in `verify_inputs.py`;
- the rev9 ledger rows;
- the current `OUT` listing, the clock and the Django objects fields.

I ran no command, test or hash. Main reports, at these bytes: 198 Python tests OK; `go vet` clean; `go test` ok.

## Verdict

**3 issues open (X1B2-R2-01 Major; X1B2-R2-02 and X1B2-R2-03 Minor).** No sign-off, and no zero-issue wording is recorded. Amendment v2 stays proposed and unadopted. `preparation.py adopt`, the scratch rehearsal, canonical synthetic and Django acceptance, and seal must not run on the binding v2 path. The signed v1 recipe, `run-config.json` `c817f678…` and the clock remain in force, unchanged.
