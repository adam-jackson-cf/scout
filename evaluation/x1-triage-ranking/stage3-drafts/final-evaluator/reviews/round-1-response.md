# final-evaluator — round-1 response

**Status: fixes submitted for scoped re-review, not self-closure or sign-off.** This response addresses every issue in `round-1.md`; the review record is preserved byte-for-byte. All work and validation are synthetic-only. No preparation action (a)–(e), candidate fetch, scanner/model call, real-data inventory, attestation or binding OC was performed. No signed revision-9 byte was changed.

## Required fixes and catching regressions

| Issue | Fix | Implementation | Catching regression |
|---|---|---|---|
| R1-01 | Evaluator consumes only digest-verified bytes: exact signed manifest, inventory sizes/p_gate, OC records seal/host, measurement runs, and post-dev cost evidence. Contradictory Mhat/c_U and vector bounds also refuse. | `evaluator.py:104`, `evaluator.py:195`, `provenance.py:41` | `test_evaluator.py:104` `test_byte_bundle_substitutions_and_no_parallel_inputs`, `test_evaluator.py:131` `test_seal_internal_references_and_inventory_types`, `test_evaluator.py:86` `test_full_grid_predev_postdev_and_selection_boundaries` |
| R1-02 | Per-operation maxima across every completed run feed the signed rev8 binding expression, locked units and p_gate; overhead once; wall=hours/192; exact CPU and RSS limits. | `evaluator.py:59`, `provenance.py:59` | `test_evaluator.py:211` `test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary`, `test_evaluator.py:238` `test_memory_boundary_and_bound_host` |
| R1-03 | Sealed OC host must match all remeasurements, for every required host field; mismatch refuses as a deviation pause. | `evaluator.py:104`, `provenance.py:52`, `provenance.py:59` | `test_evaluator.py:238` `test_memory_boundary_and_bound_host` |
| R1-04 | Exact bool flags; integrity and veto precede interval checks; unmeasured interval is inadequate; .10 upper boundary remains strict. | `evaluator.py:40` | `test_evaluator.py:295` `test_ordered_typed_verdict_and_interval_boundaries` |
| R1-05 | Strict predicates/completion types and finite non-bool gains, p_gate and measurements; malformed JSON/seals fail closed. | `provenance.py:37`, `provenance.py:59`, `evaluator.py:104` | `test_evaluator.py:131` `test_seal_internal_references_and_inventory_types`, `test_evaluator.py:199` `test_reachability_gain_and_missing_mask_types`, `test_evaluator.py:253` `test_rerun_qualifications_evidence_and_measurement_types` |
| R1-06 | Reconciled passing/first-failing execution and attempt counters isolate each limit, plus eligible=1999 with consistent inadequacy. | `evaluator.py:104` | `test_evaluator.py:163` `test_reconciled_counter_caps_and_eligible_boundary` |
| R1-07 | Both-tail last-pass/first-fail fixtures per cell kind, power/null/inadequacy boundaries, reachability and exact membership/type/manifest cases. | `evaluator.py:104`, `evaluator.py:30` | `test_evaluator.py:171` `test_each_tail_boundary_for_each_cell_kind`, `test_evaluator.py:182` `test_reference_power_null_and_inadequacy_boundaries`, `test_evaluator.py:199` `test_reachability_gain_and_missing_mask_types`, `test_evaluator.py:155` `test_exact_membership_B_and_integer_mutants` |
| R1-08 | Per-input maxima, CPU/p_gate/RSS/host, all rerun/evidence/refusal cases and exact Mhat selection boundaries have isolated fixtures. | `evaluator.py:59`, `provenance.py:59`, `evaluator.py:195` | `test_evaluator.py:211` `test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary`, `test_evaluator.py:238` `test_memory_boundary_and_bound_host`, `test_evaluator.py:253` `test_rerun_qualifications_evidence_and_measurement_types`, `test_evaluator.py:86` `test_full_grid_predev_postdev_and_selection_boundaries` |
| R1-09 | Separate pre-dev evaluation without Mhat returns sealed result and digest; post-dev selection uses that bound seal and cost evidence without recomputation or lower-row bypass. | `evaluator.py:104`, `evaluator.py:195` | `test_evaluator.py:272` `test_predev_none_no_bypass_and_postdev_tampering`, `test_evaluator.py:86` `test_full_grid_predev_postdev_and_selection_boundaries` |
| R1-10 | Every completed/interrupted run, original inputs and classification is retained in the result, including closure-invalid completed runs in maxima. | `provenance.py:59`, `evaluator.py:104` | `test_evaluator.py:86` `test_full_grid_predev_postdev_and_selection_boundaries`, `test_evaluator.py:211` `test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary` |
| R1-11 | One binding-cell enumerator/serialization schema, exact per-kind counts and pinned sorted identity-list SHA; README mandates generator imports. | `evaluator.py:9`, `evaluator.py:13`, `README.md:1` | `test_evaluator.py:78` `test_identity_counts_and_known_source_pins` |

## Conservative choices and independent evidence

The accepted conservative per-population power-inadequacy rule is retained. Resources now use the signed rev8 per-operation expression and per-input maxima, with wall=hours/192. No whole-run totals or caller-charged wall input remains. Pre-dev requires no invented Mhat; post-dev only consults sealed pre-dev rows and the bound, internally consistent Mhat/c_U evidence. The authoritative variant and authenticity of external trust anchors remain outside this synthetic draft and still need their own scoped zero-issue review.

Boundary regressions exercise each cell kind/tail: 313/314 pooled misses, 8079/8078 power supports, 457/458 null supports, and 84/2084 versus 85/2085 conditional inadequacy. Counter mutations use reconciled counts so cap deletion cannot hide behind reconciliation. The initial unittest fixture `run` shadowed TestCase.run; it was renamed `measurement_run` before the passing run.

All nine meaningful scratch-only mutation probes were killed by assertion failures in the named regression methods. Full substitutions and full failing-test traces are in `round-1-mutant-output.txt`; originals/signed files were never mutated. Each probe copied only evaluator/provenance/test source plus the exact signed manifest/reporter into a temporary matching directory layout, replaced exactly one source occurrence, and ran `.tools/venv/bin/python <scratch>/evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/test_evaluator.py -v EvaluatorTests.<method>` (method names and substitutions appear in that output). Literal CPU-only predicate removal is an equivalent mutant after the required wall derivation: wall<=15 independently enforces hours<=2880. The meaningful resource probe therefore removes the combined CPU/wall predicate; it was killed. No claim is made that the equivalent literal mutant can be killed.

## Exercised command and complete output

From `/Users/adamjackson/Projects/work/scout-x1g`:

```sh
source .tools/activate
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/test_evaluator.py -v
```

Observed: **14 tests passed** under pinned CPython 3.11.11. This is deterministic implementation validation, not evidence of an HD9 effect or authentic stage authority. Verbatim output (also retained in `round-1-test-output.txt`):

```text
test_byte_bundle_substitutions_and_no_parallel_inputs (__main__.EvaluatorTests.test_byte_bundle_substitutions_and_no_parallel_inputs) ... ok
test_each_tail_boundary_for_each_cell_kind (__main__.EvaluatorTests.test_each_tail_boundary_for_each_cell_kind) ... ok
test_exact_membership_B_and_integer_mutants (__main__.EvaluatorTests.test_exact_membership_B_and_integer_mutants) ... ok
test_full_grid_predev_postdev_and_selection_boundaries (__main__.EvaluatorTests.test_full_grid_predev_postdev_and_selection_boundaries) ... ok
test_identity_counts_and_known_source_pins (__main__.EvaluatorTests.test_identity_counts_and_known_source_pins) ... ok
test_memory_boundary_and_bound_host (__main__.EvaluatorTests.test_memory_boundary_and_bound_host) ... ok
test_ordered_typed_verdict_and_interval_boundaries (__main__.EvaluatorTests.test_ordered_typed_verdict_and_interval_boundaries) ... ok
test_predev_none_no_bypass_and_postdev_tampering (__main__.EvaluatorTests.test_predev_none_no_bypass_and_postdev_tampering) ... ok
test_reachability_gain_and_missing_mask_types (__main__.EvaluatorTests.test_reachability_gain_and_missing_mask_types) ... ok
test_reconciled_counter_caps_and_eligible_boundary (__main__.EvaluatorTests.test_reconciled_counter_caps_and_eligible_boundary) ... ok
test_reference_power_null_and_inadequacy_boundaries (__main__.EvaluatorTests.test_reference_power_null_and_inadequacy_boundaries) ... ok
test_rerun_qualifications_evidence_and_measurement_types (__main__.EvaluatorTests.test_rerun_qualifications_evidence_and_measurement_types) ... ok
test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary (__main__.EvaluatorTests.test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary) ... ok
test_seal_internal_references_and_inventory_types (__main__.EvaluatorTests.test_seal_internal_references_and_inventory_types) ... ok

----------------------------------------------------------------------
Ran 14 tests in 149.162s

OK
```

## Frozen revision-9 check

`PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py` passed: **126/131 literal file matches; 5/5 shared-file owned-extract groups valid; 131/131 files unchanged against baseline e2b94e3c**. Nineteen of 20 owned extracts match literally. HD9 differs only by the already-permitted round-9 condition-2 state transition; normalising that state alone reproduces its signed hash. The manifest remains `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`. Full machine-readable output is `evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/reviews/round-1-frozen-output.json`.

## Current file SHA-256s

Paths are relative to this artefact folder. The response excludes itself to avoid a circular digest; the delivery records its digest separately. The table includes the untouched incoming council record and all exercised evidence files.

| File | SHA-256 |
|---|---|
| `README.md` | `71c4f690e7840a8c07e3974525dc95cd8884be04847cf31a634d00dd89bd2979` |
| `evaluator.py` | `aa04fb87fbe6ba41b3790dee6b388160e1814daeb74e93ec3825339e3da7a2b6` |
| `provenance.py` | `a691afbbc22b83f4e5c6af0ca7b0f667e1be3a7540cca6c53936ffa0f9d285ab` |
| `test_evaluator.py` | `fe7caf1a76138510cbb79db068de15d6e2889ebc67d6c2943f9e946943774675` |
| `check_frozen.py` | `8130568a8e3e49fdf420c110bc9825ad20f483237f1f8745a2d96fa4009a56a1` |
| `reviews/round-1.md` | `d522104c10699ae05bf30514e1535feb556eddfc77802f3dd6d8ecfc48206388` |
| `reviews/round-1-test-output.txt` | `351158c1b87b040c9ad36615984ceb927190ebe5ebe5316bf9aedf131844fe13` |
| `reviews/round-1-mutant-output.txt` | `e37918a439014b1f6d8cecca2a329a82bdd5e15de593cb3dbf8f60673e7d961b` |
| `reviews/round-1-frozen-output.json` | `2c5d8a155e57cdd60a2b6ea295e9c0c1a2453acccf0ce662a60e711a85df50d0` |
