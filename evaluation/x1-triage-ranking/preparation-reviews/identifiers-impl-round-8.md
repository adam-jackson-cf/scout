# X1 preparation (d) identifier projector — implementation review — round 8 (integration)

**Scope.** This round asks two questions:
- Is R7-01 fixed? That is, can a d journal written now, before adoption, survive a later `preparation.py adopt` of 001 → binding acceptance → seal, and can that sequence finish?
- Does the fix introduce any new concrete failure for d?

Out of scope: identifier completeness, the delta, and whether binding v2 is a good amendment. Binding v2 has its own track, and its round 1 left 5 issues open (`binding-v2-round-1.md:272`).

Method:
- I only read files. I did not run the projector, adopt, binding, seal or the tests, and I made no model calls beyond the two council peers.
- I used no real inputs.
- Main computed the digests below in `local://x1-packet-digests.md`. This council did not recompute them.

## Files and SHA-256

Paths are under `evaluation/x1-triage-ranking/`.

| File | SHA-256 | Since round 7 |
|---|---|---|
| `execution/preparation.py` | `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427` | unchanged |
| `execution/identifiers/projection.py` | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | unchanged |
| `execution/review_identifiers.py` | `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2` | unchanged |
| `execution/tests/test_identifiers.py` | `6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371` | unchanged |
| `execution/tests/test_core.py` | `94afb7638acbb0316cc019077a87d924fda0e665b752d3d57acdfadbd9657ace` | unchanged |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `execution/run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` | unchanged |
| `execution/binding/acceptance.py` | `61a95ede2c423c50b1ae0df130160608f72632484f41f7705f3e46f6cceb6b9b` | changed (R7-01 fix (a)) |
| `execution/tests/test_binding_adoption.py` | `7e15dab3cc8b21eeaf2651308e74c0e906b5e1039782c127f89c51410b00a9c5` | new |
| `preparation/model-free-binding-amendment.md` | `c8a8e418cdd99ca9c5fc691dd5d5c11766d94b4a888b956ce3fb0601ccf05362` | changed (procedure) |
| `preparation-reviews/identifiers-impl-round-7.md` | `6d279752335d4982754fd5639f729ecc008a23eacd66fa1207268caf1706b01f` | prior record |
| `preparation-reviews/binding-v2-round-1-response.md` | `e125096359b392d7e0d7ce481dfdf171c0b12afe2480d015faa849858c6e0c9a` | fix description |

Main's verification at these bytes: the Python suite ran 198 tests, OK; `go vet` was clean; `go test` was ok. I did not rerun them.

## Basis

The signed basis has not changed:
- PREREGISTRATION `abd206c3…`
- rev9-digests `2253911b…`
- round-9 authority `6f1d6ca2…`
- pins.json `6d556fb5…`
- model-free-policy `87de3949…`
- binding-check v1 `5056ae74…`
- runtime-attestation-rev6 `19463957…`
- exposure-recipe `c065347e…`, exposure-sources `67dc5a4b…`, frozen-exposure-inputs `ec92fba6…`

The 45-day clock is still the one started at `2026-11-09T03:22:58.701909Z`.

## Prior-round disposition

### R7-01 — resolved

**What was wrong.** `adopt` checked the authority receipt before it copied the amendment (`preparation.py:277–279`, then `:301–304`). Binding required that same immutable receipt to list the adopted copy, and indirectly the source manifest, neither of which could exist yet. So no adopt → binding → seal sequence could ever complete.

**What the fix does.** `binding/acceptance.py::reviewed_inputs` (`:312–357`) now splits the authority in two:

- **The pre-adoption receipt binds only files that already exist** (`:330–336`):
  - the proposal `run-config-amendments-proposed/001-binding-v2.json`;
  - the v2 recipe;
  - `model-free-binding-amendment.md`;
  - every binding source pin, which is every file under `binding/` plus `model_free/main.go` (`:422`).

  The adopted path and the source manifest are no longer required in that receipt.
- **Adoption provenance comes from the loaded config, not the receipt** (`:321–329`). Binding requires:
  - the adopted copy's SHA equals the proposal's SHA;
  - `config.amendments` has a row with exactly the adopted path and SHA;
  - that row's `adoption_review` is exactly `pin(OUT/reviews/binding-v2-amendment.json)`.

  Core writes that row from `pin(study/authority.path)` (`preparation.py:140–143`). In `run-config.json`, `OUT` is exactly `STUDY/preparation-run`, and `pin` uses `.absolute()` (`verify_inputs.py:25–28`), so the two path strings are identical.
- **The source manifest is bound after adoption, by `reviews/implementation.json`** (`:339–346`). Binding source pins must be a subset of that manifest.

**Tracing the sequence end to end:**

1. **d runs now.** `run-config-amendments/` is empty.
   - d's start record has `amendments: []`, `no-adopted-amendments`, and the base config digest (`preparation.py:332–339`, `:144–149`).
   - d's `source_pins` are only `review_identifiers.py` and `projection.py` (`review_identifiers.py:259–260`). It pins no binding file and no adoption-dependent file.
2. **adopt.**
   - The receipt re-pins only files that already exist (`:277–279`, `:394–406`).
   - adopt refuses unless every journal is terminal (`:284–287`) and strictly earlier than `adopted_at` (`:289–294`).
   - It then copies the amendment and writes the operation record and receipt (`:301–317`).
   - Its final `load_config` (`:318`) runs `verify_journal_configuration` on d. The expected amendment set is empty (`:411–412`) and matches d's record (`:418–426`).
3. **Post-adoption steps.**
   - The source manifest is produced next and includes the adopted copy. Then come `implementation.json`, the runtime review, the rehearsal, canonical synthetic and Django.
   - `reviewed_inputs` passes. The new real-adopt test shows the pre-adoption receipt bytes stay unchanged throughout (`test_binding_adoption.py:118–140`, `test_real_adopt_load_and_entire_binding_authority_chain`).
   - Binding outputs are written to SCRATCH or OUT, never to EXEC (peer H5: `fixtures.py` writes only to scratch). So the manifest stays exact.
4. **Seal.**
   - `verify_amendment_reviews` (`:160–174`) checks the amendment digest, which the proposal provides, and re-checks the adoption.
   - The manifest must equal the post-adoption EXEC file set (`:512–516`).
   - Each journal must satisfy `verify_journal_configuration` and `verify_tree_pins` (`:544–545`). d's configuration, clock, source and output pins do not change when adoption happens.
   - The adoption files and the authority receipt are counted as accounted-for outputs (`:452–457`).

**Result.** With these bytes and the documented procedure (`model-free-binding-amendment.md:64–72`), d run now, then adopt → binding → seal, can complete. The fix touches no Core file: `preparation.py` is still `599a128d…`, so `effective_configuration` and `verify_journal_configuration` are unchanged, which is what round 7 required. Both peers reached the same conclusion on their own.

**Round-7 suggestion 1 (stale procedure document)** is also done. `model-free-binding-amendment.md:66–72` now gives the exact `preparation.py adopt` command at canonical absolute paths and forbids hand-copying the file or hand-writing timestamps.

## New issues

None.

## Suggestions (not issues)

1. **Route to binding-v2 round 2: a post-adoption binding fix would lock every command** (peer S1).
   - Every `load_config` re-pins every path in the immutable 001 receipt (`preparation.py:132`, `:394–405`), and that now includes all of `binding/**` and its fixtures.
   - The rehearsal can only run after adoption, and fixture trigger counts are still unmeasured. If the rehearsal shows a fixture or controller defect, any edit makes every command refuse, including adopt of a 002, binding, seal and any d recovery. There is no way out inside the tool.
   - This already existed before the fix (round 7 already required the receipt to bind the binding sources), and it does not depend on d: d's journal pins no binding file, and running d does not narrow binding's choices.
   - Binding round 2 should either accept this explicitly or define a reviewed recovery route before anyone adopts.
2. **Keep council records outside `OUT`.**
   - Records named by the required reviews and the rehearsal review are collected only into `reviews` (`:525`). They are never added to the referenced set, so a record stored under `OUT` makes seal refuse with "unaccounted run output" (`:550–558`).
   - `test_binding_adoption.py:39–45` puts records in `OUT/reviews/`, but that test never seals.
   - The procedure should state that records live in `preparation-reviews/`.
3. **Never edit pinned binding documents after adoption.** The 001 receipt pins `model-free-binding-amendment.md`, including its "Proposed; unadopted" status line (`:3`). Marking it "adopted" afterwards would break every load. Record the adoption elsewhere, for example in the R-087 ledger row, which is not pinned.
4. **Use one canonical, symlink-free path for everything** (round-7 suggestion 2, still applies).
   - `acceptance.EXEC` is built from `resolve()`.
   - The adopted row and `verify_adoption`'s writer and config pins are built from `absolute()`.
   - Run d, adopt, binding and seal with the same canonical absolute paths used by the a/b argv. The new test patches `controller.EXEC` to the config tree, which would hide any mismatch.
5. **Test gaps** (carried forward):
   - No test seals with a d journal written before adoption. `test_load_and_seal_accept_legacy_ab_and_d_before_adoption` never calls `collect_seal`.
   - `test_binding_adoption.py` has no pre-adoption journals and never seals.
   - Add one synthetic test that runs the whole chain: pre-adoption journal → adopt → `reviewed_inputs` → `collect_seal`.
6. **acceptance.py may still change in binding round 2.** If it does, binding round 2 must preserve the R7-01 property: the pre-adoption receipt names only files that exist before adoption. d does not depend on `acceptance.py`'s bytes, so this does not affect whether d may run.

## Recommendation, alternatives, risks, execution

**Recommendation.** Record zero issues for d and run the projector once at its canonical paths.

**Alternatives considered:**
- **Option (b): a separate post-adoption binding-authorization receipt.** It would also work, but it adds a new artifact and validation path. The challenger recommends against it.
- **Wait for binding round 2 before running d.** Rejected. d's journal is independent of binding's bytes, and waiting uses up study time on the 45-day clock.

**Risks:**
- Every conclusion comes from reading code [INFERENCE]; nothing was reproduced. The full chain has no end-to-end test (suggestion 5).
- Adoption cannot be undone (suggestion 1).
- Once d has run, any later change to `verify_journal_configuration` or the `Journal` record shape would invalidate d's journals. `preparation.py` must stay at `599a128d…`, which adoption's writer pin requires anyway.
- If d fails, seal is blocked (`:533–534`) until a reviewed recovery. This is d's own recovery path, not something introduced here.

**Execution outline:**
1. Immediately before running d, confirm that `run-config-amendments/` and `preparation-run/config-adoptions/` are absent or empty, and that all seven a/b journal pairs are terminal and passed.
2. Run `review_identifiers.py` once with the pinned CPython 3.11.11, using the canonical `--frozen`, `--sources`, `--output-root` and clock paths.
3. Hold the separate completeness/delta output review.
4. Binding round 2 resolves its open issues plus suggestions 1–3. Then follow `model-free-binding-amendment.md:64–72`: adopt, then the source manifest and implementation/runtime reviews, then rehearsal, synthetic, Django, and finally seal.

## Verification and peers

**What I read:**
- `preparation.py` lines 1–163, 160–443 and 444–662
- `acceptance.py` lines 1–21, 305–363 and 383–451
- `review_identifiers.py` lines 234–293
- `test_binding_adoption.py` lines 19–143
- the 001 proposal
- the `run-config.json` paths and `source_manifest`
- `model-free-binding-amendment.md` (grep of the adopt procedure)
- the round-7 record, the binding-v2 round-1 response, and the verdict of binding-v2 round 1

**innovation-challenger:** 0 issues. It found R7-01 resolved, checked seven attack cases (an alternate-path review, a missing row, a same-content receipt, a post-adoption receipt edit, substituted bytes, omitted sources, a manifest made before adoption) and found every one refused. It recommends fix (a) over (b).

**experiment-peer:** 0 issues. It ruled out seven hypotheses H1–H7, covering path strings, EXEC resolution, files created after adoption, seal re-checks, binding writing to EXEC, an existing `implementation.json`, and review cycles. Its points became suggestions here: S1 (post-adoption lock-in) became suggestion 1, S3 became 2, S2 became 3, and S4 together with its test confounds became 4 and 5.

**Where the peers changed my recommendation:** they did not change the verdict. The peer added suggestions 1–3. I checked suggestion 1 against `preparation.py:132` and `:394–405`, and suggestion 2 against `:525` and `:550–558`. I classed them as binding-track or procedure items rather than d issues, because d's journal pins neither binding files nor council-record locations.

## Verdict

**zero issues — projector may be run once at its canonical paths with preparation.py 599a128d…; completeness/delta sign-off remains a separate output review.**

Signed off at these reviewed digests:
- `execution/preparation.py` `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427`
- `execution/identifiers/projection.py` `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072`
- `execution/review_identifiers.py` `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2`
- `execution/tests/test_identifiers.py` `6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371`
- `execution/tests/test_core.py` `94afb7638acbb0316cc019077a87d924fda0e665b752d3d57acdfadbd9657ace`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`
- `execution/run-config-amendments-proposed/001-binding-v2.json` `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121`
- `execution/binding/acceptance.py` `61a95ede2c423c50b1ae0df130160608f72632484f41f7705f3e46f6cceb6b9b`. Signed off only for the R7-01 integration question; binding-v2's own sign-off is still pending.
- `execution/tests/test_binding_adoption.py` `7e15dab3cc8b21eeaf2651308e74c0e906b5e1039782c127f89c51410b00a9c5`
- `preparation/model-free-binding-amendment.md` `c8a8e418cdd99ca9c5fc691dd5d5c11766d94b4a888b956ce3fb0601ccf05362`
- `preparation-reviews/identifiers-impl-round-7.md` `6d279752335d4982754fd5639f729ecc008a23eacd66fa1207268caf1706b01f`
- `preparation-reviews/binding-v2-round-1-response.md` `e125096359b392d7e0d7ce481dfdf171c0b12afe2480d015faa849858c6e0c9a`
