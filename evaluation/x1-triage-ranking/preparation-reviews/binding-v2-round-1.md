# X1 preparation model-free binding v2 — amendment and implementation review — round 1

**Scope and method.** This round reviews the experiment-rule-5 amendment that replaces v1's process-observation requirement with the four-layer deterministic proof, and the v2 implementation. The review was read-only. It made no model calls beyond the two council peers and used no real inputs. Nothing was executed: no scanner, rehearsal, acceptance, adopt, seal or test. I did not recompute digests, because no shell was available to the council.

There were three independent analyses: `innovation-challenger` (seams and Q1), `experiment-peer` (requirement retention and mechanics, Q2 and Q4) and `experiment-peer` (implementation and prior-finding dispositions, Q3). I also ran my own source trace.

**Verdict: 5 issues open (1 Major, 4 Minor).** No sign-off. Amendment v2 is not adopted. Neither the scratch rehearsal nor the canonical binding may be run.

## Reviewed files and digests

Main supplied these digests; the council did not recompute them. Paths are relative to `evaluation/x1-triage-ranking/` unless shown otherwise.

**Signed basis.** Checked as still referenced; v1 bytes are not edited.

| File | SHA-256 |
|---|---|
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |
| PREREGISTRATION | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| rev9-digests | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| pins.json | `6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc` |
| `preparation/model-free-policy.json` | `87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f` |
| `preparation/model-free-binding-check.json` (v1) | `5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c` |
| runtime-attestation-rev6 | `19463957becfc23a391eac03582b32eaa0ff392069bccd9aed9000f390ad2d62` |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |

**Under review.**

| File | SHA-256 |
|---|---|
| `preparation/model-free-binding-check-v2.json` | `4788e60e43f988b7dee2c916ac675a027892dc5a23d3288ba46dd9bc584f180c` |
| `preparation/model-free-binding-amendment.md` | `225a4304d8e42f584a39675ecea1aa9aba5686db9eaaefae8c896fe726784324` |
| `execution/run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` |
| `execution/binding/acceptance.py` | `916d3d732f2efbdfa98e1119d9ab5e49401cac7c8e17edc91a404bb48a4a21c1` |
| `execution/binding/interlock.py` | `19464e8940d0458be3606ed0438ae778fd01b48b73f8c2fee801856b80b2e6a1` |
| `execution/binding/fixtures.py` | `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b` |
| `execution/binding/fixtures/**` | unchanged from binding round 2; per-file digests as listed in `binding-v2-REVIEW_REQUEST.md` |
| `execution/model_free/main.go` | `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` |
| `execution/model_free/main_test.go` | `16e66424f7bfef1a0faec66de9793c19e59018ffe8f78191af75524f9c271f70` |
| `execution/tests/test_binding.py` | `af82dc5165ec144754f5811a493ac67f454eecde252075038a4ca225fb51330a` |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `preparation-reviews/binding-v2-REVIEW_REQUEST.md` | `b888d562b5f0b73219bcbd9490d7fea721069ed3b41d91108c9b89a08af9ae76` |
| `preparation-reviews/binding-v2-checks.json` | `22a5d5bff095260f39a710fdd2f5b8161293f301a688cac0cf505377e0817ea6` |
| `preparation-reviews/binding-v1-unprivileged-probe.json` | `10a0c930a3b3d006e7a9c368ea7a93f709ebfda13a149b7521dc1c5721ef4881` |
| `execution/preparation.py` | supplied as `e883abc47ea9ea6a693a78ffe46e5793bb1a6b1ccdc61cb17bf2c815ab99db6e`; see X1B2-R1-05 |
| `execution/tests/test_core.py` | supplied as `2fa57759d64e2ed19a533358b79dc011375fdef82f7908a6728daad968314e41`; see X1B2-R1-05 |

**Audited product seams.** The 22 files listed in the request (`internal/engine/assess.go` `dbb20a1e…` through `internal/engine/lanes.go` `7bcddebb…`) were read at the cited lines.

## Basis

- **Authority.** Round 9 permits preparation actions (a)–(e).
- **Clock.** Started 2026-11-09T03:22:58.701909Z. This change does not touch it: `load_config` validates the clock's `execution_config` pin (`preparation.py:108-114`), and the rehearsal only reads the clock.
- **Owner challenge and delegation.** R-087 (ledger row pending) records the owner's challenge on the managed Mac; R-064 is the delegation.
- **Severity.** Calibrated as stated in the request.
- **Prior findings.** The prior binding findings are from `binding-impl-round-1.md` and `binding-impl-round-2.md`. Shared adoption code is concurrently under `identifiers-impl-round-7.md`.

## Answers to the four questions

**Q1: Does the four-layer proof answer the model-involvement question at least as well as v1? Yes, within the stated trust assumption, but the residual limitation is incomplete (X1B2-R1-02).**

The layers are real:

- **Layer 1.** Actual Config and Options are deep-equality-checked before the registry and scanners run (`main.go:291-304, 389-392`). `Assess` does not reload policy.
- **Layer 2.** The verbatim `result.Lanes` is checked for exactly-once cardinality and for skipped/disabled reasons that the product actually emits (`main.go:313-357`; `acceptance.py:132-171`; `internal/engine/lanes.go:150-225`; `internal/association/lane.go:33-55`).
- **Layer 3.** The private environment puts stand-ins first and keeps them at `CLAWPATCH_CLAUDE_BIN` (`acceptance.py:26-38`). `ScrubEnv` passes `CLAWPATCH_*` through (`internal/modelcli/modelcli.go:93-108`).
- **Layer 4.** The control runs on the contextual seam with production argv prefixes (`internal/contextual/cli.go:24-54`; `interlock.py:36-63`).

All 12 rows of the seam table were checked against source and are accurate. Two have imprecise pinpoints: trigger metadata (`plan.go:57` → `122-146`) and sandbox env (`run.go:35-39` → `35-47`).

**What v2 gives up relative to v1.** v1 would have caught an unexpected descendant or an absolute-path model launch. v2 detects only the audited seams. That loss is covered by the stated scanner-trust limitation for scanners, but not for the unsandboxed Node/Clawpatch tree.

**Q2: Are the v1 requirements retained? Yes in recipe text and in code.**

- Both fixtures, both revisions, parity/multiplicity against a scanner-only reference, the key-absence boolean, the positive control, sealed evidence and refusal semantics are all retained.
- Gaps:
  - The Django receipt does not pin the synthetic receipt it relies on (X1B2-R1-04).
  - Several retained or new boundaries are untested (X1B2-R1-03).

**Q3: Is the implementation correct, and are prior findings disposed of correctly?** By inspection the predicate, logs, stand-ins, Config/Options checks and lane checks are correct. The applicable findings are fixed and the observer-specific ones are correctly superseded (table below). Test gaps remain (X1B2-R1-03).

**Q4: Are the amendment mechanics sound? No.**

- **Sound:** v1 is unchanged, v2 is a new file, the proposal directory is never loaded, and the clock is untouched.
- **Not sound:** the amendment-review receipt that acceptance requires cannot be produced by any adopt sequence (X1B2-R1-01).
- **Stale digests:** the request names digests for `preparation.py` and `test_core.py` that the concurrent identifiers track has superseded (X1B2-R1-05).

## Disposition of prior findings

| Finding | Claimed | Council |
|---|---|---|
| R1-01 | Direct launch and private environment retained; ES allowlist part superseded | **Verified.** `acceptance.py:77-105`; `test_private_env_stubs_all_provider_paths_and_no_outer_sandbox`, `test_actual_frontend_launch_is_direct_and_private`, `test_static_provider_binary_seams`. No observer files remain. |
| R1-02, R1-03 | Superseded | **Verified.** Their only basis was v1's "capture all child-process starts", which v2 removes explicitly (`supersedes_on_adoption`). No privilege flag is bypassed. |
| R1-04 / R2-01 | Fixed | **Verified.** `main.go:139-143` (ineligible is healthy only when all rows are not_applicable, nonempty and failure-free); `TestRealIneligibleRunnerHealth`. |
| R1-05 residual | Source fixed; trigger set still unmeasured | **Verified as stated.** The rehearsal exists (`acceptance.py:374-405`) and is gated before lock (`:420-423`), but it is currently unreachable because of X1B2-R1-01. |
| R1-06, R1-07, R1-08, R1-10 | Retained | **Verified.** Tests: `test_source_manifest_not_mistaken_for_filesystem_pin`; `TestOccurrenceKeyIndependentExpectedValues` and `TestDurableLinkBoundaries`; `TestPlannerInputsAndScannerManifest` and `TestReferenceOffFileObservationFiltering`; `test_equal_length_roots_and_missing_head_state_directory`. |
| R1-09 | Fixed | **Verified.** Six resolve-only runs precede any scanner (`acceptance.py:202-211`); the later run must be byte-identical (`:70-75`). Tests: `test_both_revisions_resolved_before_any_scanner`, `test_invalid_reference_head_resolution_prevents_every_scanner`. |
| R1-11 | Superseded | **Verified.** No ES parser remains. |
| R1-12 | Retained | **Partly verified.** Refused receipts are create-once (`test_calledprocesserror_always_terminates_refused_receipt`). The rehearsal gate runs before reservation, but the adopted-pin and Django prerequisites run inside `work()` after reservation (`acceptance.py:428-439`). For synthetic this is gated in practice by the rehearsal; the Django case is suggestion S2. |
| R1-13 | Tests added | **Mostly verified.** Gaps are in X1B2-R1-03. |
| R2-02 | Superseded | **Verified.** No interpreter allowlist remains. A real shebang stub with `-p` is accepted by the positive check and rejected by the bound predicate (`test_actual_failing_stub_evidence`). |
| R2-03 | Superseded | **Verified.** No observer stop path remains. |
| S1–S9 (round 2) | As claimed in the request | **Accepted.** S8's map-source check is effectively constant because `Assess` hard-codes heuristic mapping (`assess.go:551,577`); see suggestion S1. |

## New issues

### X1B2-R1-01 — Major — `acceptance.py:316-327` against `preparation.py:277-279, 297-304, 394-405, 79-81, 514-516`

**Finding.** `reviewed_inputs` requires `OUT/reviews/binding-v2-amendment.json` to list two things that cannot be in it:

- **(a) The adopted path** `EXEC/run-config-amendments/001-binding-v2.json`. That same file is the overlay's adoption authority:
  - `adopt()` validates it with `review_receipt` (which re-pins every listed path, `:403`) before the adopted copy exists (`:297-304`).
  - After adoption the receipt's bytes are pinned by the adoption receipt (`:79-81`).
- **(b) `pin(OUT/execution-source-manifest.json)`.** This manifest does not exist yet, so this record cannot name its digest, and `review_receipt` requires the council record to contain every reviewed digest. Seal also requires the manifest to equal the post-adoption EXEC file set (`:514-516`), which includes the adopted copy.

**Failure.** Every sequence fails:

- If the receipt lists the adopted path, `adopt` raises `FileNotFoundError`.
- If it lists only the proposal, adoption succeeds but every rehearsal and canonical run refuses with "v2 amendment review does not bind…".
- If the receipt is edited after adoption, every `load_config` refuses.

So the binding, and therefore seal, can never pass.

**Why the tests miss it.** `AdoptedAmendmentTests` patches `pin` and `review_receipt` (`test_binding.py:380-407`). Core's fixture reviews only the proposal (`test_core.py:177-187`). This is the same defect as R7-01 in identifiers round 7.

**Related.** The procedure text says Main copies the file and writes an `adopted_at` receipt by hand. That contradicts `preparation.py adopt` and the review schema `{scope, record, reviewed}`, which has no `adopted_at` key (`model-free-binding-amendment.md:66`; `binding-v2-REVIEW_REQUEST.md:12`).

**Required fix.** This is R7-01's fix (a). It changes only binding-track files and leaves `preparation.py` alone.

1. In `reviewed_inputs`, require the amendment review to bind:
   - the **proposal** pin;
   - the recipe;
   - `model-free-binding-amendment.md`;
   - every `source_pins` entry.

   All of these exist before adoption.
2. Separately require all of the following:
   - `config.amendments` contains the adopted pin;
   - its SHA equals the reviewed proposal SHA;
   - its `adoption_review` equals `pin(OUT/reviews/binding-v2-amendment.json)`.
3. Remove `pin(source_path)` from the amendment-review requirement. The post-adoption `reviews/implementation.json` already binds the manifest, and `source_pins` is already checked to be a subset of the manifest (`acceptance.py:330-337`).
4. Add an unmocked test that runs the real `adopt()` with a real receipt, then `load_config`, then `reviewed_inputs`, and gets past the amendment check. Add negative cases: the proposal digest missing, one `source_pins` entry missing, and an adopted SHA that differs from the reviewed one.
5. Rewrite amendment line 66 and the request around `preparation.py adopt --config … --amendment <proposed> --review <receipt>`. State that the source manifest is produced **after** adoption.

### X1B2-R1-02 — Minor — `model-free-binding-check-v2.json` `residual_limitation`; amendment "What the proof answers" section

**Finding.** The residual limitation names one trust root: scanner binaries, trusted via the closure and the analyzer network sandbox. It omits two others:

- **The frontend, Node/Clawpatch and their children run outside any sandbox with network available.** Clawpatch init and heuristic map are launched directly (`internal/clawpatch/client.go:322-329`), and they spawn `git` and `go list` (`third_party/clawpatch/src/mappers/go.ts:110-121`). Their model-freeness rests on the cited, non-transitive source audit, the hash-pinned closure and the stand-ins, not on the sandbox.
- **The positive control proves the detector is live only for Scout's contextual seam.**
  - Clawpatch review is blocked before any spawn, because egress is unauthorized (`internal/clawpatch/review.go:175-177`).
  - So Clawpatch's own `CLAWPATCH_CLAUDE_BIN` and literal-`codex` routes to the stand-ins are configured and statically checked, but never shown to be live.
  - The amendment's sentence "the enabled control demonstrates the attempt detector is live" overstates this.

The stand-ins also never reach sandboxed scanner descendants, because the sandbox rebuilds PATH (`internal/sandbox/run.go:35-47`). For scanners, the sandbox is the guard.

**Failure.** Under rule 6 the sealed scope claim would understate the residual trust: a reader would not learn that a network-capable, partially audited Node dependency tree is trusted.

**Required fix.** Text only, in both the recipe and the amendment:

- name the unsandboxed frontend/Clawpatch/Node/git/go-list trust root and what it rests on;
- limit the liveness claim to the contextual seam, and state that exclusion of the Clawpatch provider seams rests on layers 1–2 (lanes not opted in, egress false, heuristic map with a null provider), with the stand-ins as defence in depth;
- note that stand-ins do not reach scanner descendants.

### X1B2-R1-03 — Minor — `acceptance.py:79-80, 94-99, 178-196, 233-234, 269-274, 310-311, 378-379`

**Finding.** These proof-layer decision boundaries have no catching test:

- **Controller key refusal** while `OPENROUTER_API_KEY` is present, including an empty value: at entry (`reviewed_inputs`), at rehearsal entry, and at each launch. Only Go and `private_env` are tested; the controller check is the only one that sees the caller's environment.
- **`acceptance_flags` returning false** for:
  - `positive_control_expected_failure`: one control only, providers swapped or duplicated, `model_attempts==0`, `stub_exit_code≠97`;
  - `zero_model_attempts` with a nonzero count.
- **A resolve-only synthetic-unbound run** with a logged attempt must refuse.
- **Exit-code handling:** a bound run's nonzero exit must refuse, while the positive control's nonzero exit is tolerated.
- **Django `fixture_manifest`:** substituted tip, a `head^1` mismatch, a noncanonical or failed objects receipt, and a non-passing synthetic receipt.
- **Equal synthetic parent/head sets** must refuse.

**Failure.** Regressions would keep the suite green. Examples: dropping the provider-order check, `model_attempts>=0`, or deleting the `head^1` comparison at `:193-194`. Each silently weakens layer 4 or the Django pair.

**Required fix.** Add table tests for each boundary.

### X1B2-R1-04 — Minor — `acceptance.py:178-182, 195`

**Finding.** The v2 recipe's check 4 says Django acceptance "must bind the already-passing sealed synthetic receipt". `fixture_manifest` reads `OUT/binding/synthetic.json` and copies its `positive_control`, but the Django receipt records no pin of that file.

**Failure.** After a reviewed recovery that recreates `synthetic.json`, nothing in the Django receipt or at seal shows which synthetic receipt it relied on.

**Required fix.** Record `synthetic_receipt: pin(OUT/binding/synthetic.json)` in the Django evidence; `verify_tree_pins` already re-checks it at seal. Add a test.

### X1B2-R1-05 — Minor — `binding-v2-REVIEW_REQUEST.md` shared-contract table; `binding-v2-checks.json`

**Finding.** The on-disk `preparation.py` contains `adopt()`, marked "R6-01 code-enforced choice", and `verify_adoption`. `identifiers-impl-round-6.md:45-47` says the version at `e883abc4…` had no such writer. `identifiers-impl-round-7.md:15-16` records the current files as:

- `preparation.py` `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427`
- `test_core.py` `94afb7638acbb0316cc019077a87d924fda0e665b752d3d57acdfadbd9657ace`

[INFERENCE: the council did not hash these files.] The experiment-peer also saw `preparation.py` change mid-review, from 528 to 662 lines.

**Failure.** A sign-off would name digests for files other than those that execute adoption. The 164-test evidence may also predate them.

**Required fix.** Re-pin every file, including the current `preparation.py` and `test_core.py`, plus the per-file `binding/fixtures/**` digests. Attach a fresh full suite, `go vet` and `go test` record to the next request.

## Suggestions (not blocking)

- **S1.** Make `heuristic_only` also require `!Mapping.ExplorationIncomplete`. Today `Source` just echoes the requested source, and the agent-fallback path returns `Source=heuristic` (`internal/clawpatch/client.go:231-263`). Scanner health is the only thing that catches it today.
- **S2.** Move these checks before `canonical.open('x')`: the key and clock checks, Django's prerequisites (passing synthetic receipt, objects receipt), and absence of `SCRATCH/{synthetic,django}`. Also refuse rehearsal outputs under `SCRATCH/{synthetic,django}`, where a later canonical `mkdir(exist_ok=False)` would burn a create-once receipt.
- **S3.** Pin `preparation/model-free-policy.json` explicitly in each receipt. Today its digest is evidenced only by `frozenDigest` (`main.go:27`).
- **S4.** Rehearsals have no journal, so require the rehearsal review to list every rehearsal report produced.
- **S5.** At each launch, re-check the frontend pin and compare stub bytes with `stub_source`.
- **S6.** Seam inventory additions:
  - the precise pinpoints noted under Q1;
  - other Clawpatch providers (acpx, cursor, grok, opencode, pi), which are unreachable on the heuristic path (`app.ts:80`);
  - the npm `registry-verifier` HTTP, which is reachable only from review.
- **S7.** Adoption pins the `preparation.py` writer (`:89-90`). Adopt only with the final signed-off `preparation.py`, invoked by its canonical absolute path (identifiers round 7, S2).

## Recommendation

Do not adopt v2 yet. Fix X1B2-R1-01 with R7-01 fix (a), confined to `acceptance.py`, `test_binding.py`, the amendment and the request. Then fix R1-02 to R1-05 and hold round 2. With those fixes the four-layer design is adequate for X1's question, and it is the right replacement for the root-dependent v1 observer.

## Alternatives

- **(A) Recommended: v2 with binding-track fix (a).** No digest in the identifiers track changes.
- **(B) Two separate authorities.** Keep the immutable pre-adoption adoption receipt, and add a post-adoption binding-authorization receipt. This changes `preparation.py` or adds a receipt type, which forces re-review in the identifiers track. More moving parts, no gain.
- **(C) Option A plus a network-deny sandbox around the frontend/Clawpatch subtree.** This removes the unsandboxed Node trust root, but needs a new reviewed sandbox profile and full-route compatibility evidence. The successful nesting probe is not route evidence. Optional hardening; R1-02 asks only that the trust be disclosed.
- **(D) Keep v1.** Not feasible: it needs root or an ES client, and the probe failed with `ES_NEW_CLIENT_RESULT_ERR_NOT_PRIVILEGED` (`binding-v1-unprivileged-probe.json`).

## Evidence

Cited inline above. Key traces:

- Adoption deadlock: `acceptance.py:316-327` ↔ `preparation.py:277, 297-304, 403, 79-81`.
- `Assess` uses only heuristic map (`assess.go:551, 577`) and the nil-association `disabled` path (`association/lane.go:45-48`).
- `ApplyModelCLIPolicy` leaves Clawpatch egress unset (`model_cli_policy.go:17-26`).
- Sandbox PATH rebuild (`sandbox/run.go:35-47`).
- Clawpatch launched unsandboxed (`clawpatch/client.go:322-329`).

## Risks

- **Static reading only.** All findings come from reading code; nothing was reproduced [INFERENCE].
- **Unmeasured fixture behaviour.** Trigger counts and actual lane rows on the fixtures remain unmeasured until the rehearsal runs.
- **Adoption is irreversible.** There is no re-adopt path, so adopt only after round 2 signs and with final `preparation.py` bytes.
- **Trust in the Clawpatch/Node closure** remains a stated assumption, not an observation.

## Peer influence

- **Challenger.** Reported no issues and confirmed every citation. It supplied the sandbox PATH rebuild, the `go list` spawn and the extra-provider facts, which strengthened my R1-02. I raised R1-02 to an issue because Q1 asks explicitly whether the limitation is correctly stated.
- **Retention/mechanics peer.** Reported B-01, which matched my source-manifest circularity and merged into R1-01. B-04 became R1-04 and B-05 became R1-03. I downgraded B-06 to S2 because the documented path does not collide. Its digest-staleness note led me to identifiers round 7, which produced R1-05.
- **Implementation peer.** Raised N1 as Critical. I rate it Major: the canonical reservation is unreachable because `verify_rehearsal` runs before the lock, so nothing irreversible is written. N2 and N3 merged into R1-03 and R1-04. Its disposition table matched my own reading.

## Execution outline

1. Implement R1-01 fix (a) with the unmocked adopt → `load_config` → `reviewed_inputs` test, and rewrite the adoption procedure text.
2. Apply the text fix for R1-02 in both the recipe and the amendment.
3. Add the R1-03 table tests and the R1-04 `synthetic_receipt` pin with its test.
4. Main runs the full Python suite, `go vet` and `go test`, and re-pins every file (R1-05).
5. Hold council round 2; it must report zero issues before Main runs `preparation.py adopt`.
6. Only after adoption: generate the source manifest, the implementation review and the runtime review, then the scratch rehearsal, the rehearsal-report review, canonical synthetic, and Django.

## Verification

I read, at the cited lines:

- the v1 and v2 recipes, the amendment, the request and the proposal overlay;
- `acceptance.py`, `interlock.py`, `main.go` and `preparation.py` (adopt, review_receipt, load_config, seal);
- the test inventories;
- the product seams listed in the request, plus `contextual/cli.go`, `contextual/client.go`, `clawpatch/client.go`, `clawpatch/review.go`, `sandbox/run.go`, `domain` lane/status types, and Clawpatch `app.ts`, `agent-mapper.ts` and `mappers/go.ts`;
- `identifiers-impl-round-6.md` and `identifiers-impl-round-7.md`.

I ran no tests, hashes or commands. Main reports its own verification: suite OK, `go vet` clean, `go test` ok.

## Verdict

**5 issues open (X1B2-R1-01 Major; X1B2-R1-02, -03, -04, -05 Minor).** No sign-off, and no zero-issue wording is recorded. Amendment v2 stays proposed and unadopted. No rehearsal, canonical binding or `preparation.py adopt` may run. The signed v1 recipe, `run-config.json` `c817f678…` and the clock remain in force, unchanged.
