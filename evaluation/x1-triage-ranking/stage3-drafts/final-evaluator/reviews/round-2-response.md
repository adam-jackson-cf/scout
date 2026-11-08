# final-evaluator: round-2 remediation response

**Status: implemented and exercised; submitted for scoped re-review, not council acceptance.** Responds to every R2 required fix and each named R1 residual in the byte-preserved `round-2.md`. All work is synthetic-only and pre-clock. No preparation action (a)–(e), real repository data, candidate fetch, bootstrap/attestation, binding implementation, identifier review, authentic sealing/inventory, scans or model calls occurred.

## Issue → fix → catching regression

| Issue | Fix | Implementation | Catching test |
|---|---|---|---|
| R2-01 | Failing resource fits explicitly fail every row; independent memory-only failure and exact CPU=2880 passing integration fixture; row-gate mutant killed. | `evaluator.py:104` (`evaluate`), `evaluator.py:59` (`resource_fit`) | `test_evaluator.py:284` (`test_resource_fit_propagates_to_every_row`), `test_evaluator.py:255` (`test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary`) |
| R2-02 | Consistently resealed changed manifest argument/member/internal OC digest still fails the independent signed pin. | `provenance.py:72` (`verify_bundle`) | `test_evaluator.py:134` (`test_signed_manifest_pin_with_consistent_bundle_resealing`) |
| R2-03 / R1-05 residual | Hand-built exponent-overflow OC bytes reach finite-gain cell decisions. Exact records-member byte sealing avoids premature float reserialization; nonfinite gains fail each cell. | `provenance.py:36` (`member_bytes`), `evaluator.py:104` (`evaluate`) | `test_evaluator.py:226` (`test_nonfinite_exponent_gain_bytes_fail_cells`), `test_evaluator.py:246` (`test_exact_record_member_bytes`) |
| R2-04 | Isolating host field/type and load/process provenance refusals keep run/OC hosts equal; missing process count also fails. | `provenance.py:83` (`validate_host`), `provenance.py:90` (`maximum_completed_inputs`) | `test_evaluator.py:315` (`test_complete_host_and_load_provenance`) |
| R2-05 | Point immediately below .10 with positive lower bound remains inconclusive; reversed interval remains inconclusive. | `evaluator.py:40` (`verdict`) | `test_evaluator.py:396` (`test_ordered_typed_verdict_and_interval_boundaries`) |
| R2-06 | Exact complete bootstrap schema fixes values/types and rejects bool, missing/extra fields and truthy stand-ins; positive complete evidence and inverse mutation tested. | `evaluator.py:195` (`complete_bootstrap_evidence`), `evaluator.py:205` (`select_postdev`) | `test_evaluator.py:374` (`test_complete_typed_bootstrap_evidence`) |

## Conservative implementation choices and independent evidence

The typed bootstrap declaration is exact: int bootstrap_means=10000, int bootstrap_seed=169300, str resampling_unit="run-cluster", float one_sided_level=.95, int unknown_or_incomplete_receipts=0; bool, extra/missing fields and wrong values/types refuse. Computing c_U from receipts remains the later dev-gate stage, as the council recommended. Trust-anchor authenticity and authoritative use remain separately reviewed prerequisites.

For R2-03, the records seal now hashes the exact UTF-8 JSON value bytes of the top-level records member, not reserialized floating-point objects. Existing compact sorted-key producers keep the same digest. This lets correctly byte-sealed 1e400/-1e400 test inputs reach each finite-gain check without weakening the JSON literal/duplicate-key refusal or inventing sentinel gains. The parsed-member extractor has separate structural tests.

All 10 committed scratch-only mutants were caught by assertion failures. The host-guard removal additionally produces downstream missing-field/type errors for two malformed hosts; the many other host fixtures fail their explicit assertions, so this is not an import/runtime-error-only kill. Full outputs are preserved. No real-data inference or HD9 result is claimed.

## Exact commands and complete suite output

Commands ran from `/Users/adamjackson/Projects/work/scout-x1g` after `source .tools/activate`, using pinned CPython 3.11.11. For CLI reproduction select a **new nonexistent destination**; the committed create-once receipts must not be overwritten.

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/test_evaluator.py -v
```

Complete output, also `round-2-test-output.txt`:

```text
test_byte_bundle_substitutions_and_no_parallel_inputs (__main__.EvaluatorTests.test_byte_bundle_substitutions_and_no_parallel_inputs) ... ok
test_complete_host_and_load_provenance (__main__.EvaluatorTests.test_complete_host_and_load_provenance) ... ok
test_complete_typed_bootstrap_evidence (__main__.EvaluatorTests.test_complete_typed_bootstrap_evidence) ... ok
test_each_tail_boundary_for_each_cell_kind (__main__.EvaluatorTests.test_each_tail_boundary_for_each_cell_kind) ... ok
test_exact_membership_B_and_integer_mutants (__main__.EvaluatorTests.test_exact_membership_B_and_integer_mutants) ... ok
test_exact_record_member_bytes (__main__.EvaluatorTests.test_exact_record_member_bytes) ... ok
test_full_grid_predev_postdev_and_selection_boundaries (__main__.EvaluatorTests.test_full_grid_predev_postdev_and_selection_boundaries) ... ok
test_identity_counts_and_known_source_pins (__main__.EvaluatorTests.test_identity_counts_and_known_source_pins) ... ok
test_memory_boundary_and_bound_host (__main__.EvaluatorTests.test_memory_boundary_and_bound_host) ... ok
test_nonfinite_exponent_gain_bytes_fail_cells (__main__.EvaluatorTests.test_nonfinite_exponent_gain_bytes_fail_cells) ... ok
test_ordered_typed_verdict_and_interval_boundaries (__main__.EvaluatorTests.test_ordered_typed_verdict_and_interval_boundaries) ... ok
test_predev_none_no_bypass_and_postdev_tampering (__main__.EvaluatorTests.test_predev_none_no_bypass_and_postdev_tampering) ... ok
test_reachability_gain_and_missing_mask_types (__main__.EvaluatorTests.test_reachability_gain_and_missing_mask_types) ... ok
test_reconciled_counter_caps_and_eligible_boundary (__main__.EvaluatorTests.test_reconciled_counter_caps_and_eligible_boundary) ... ok
test_reference_power_null_and_inadequacy_boundaries (__main__.EvaluatorTests.test_reference_power_null_and_inadequacy_boundaries) ... ok
test_rerun_qualifications_evidence_and_measurement_types (__main__.EvaluatorTests.test_rerun_qualifications_evidence_and_measurement_types) ... ok
test_resource_fit_propagates_to_every_row (__main__.EvaluatorTests.test_resource_fit_propagates_to_every_row) ... ok
test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary (__main__.EvaluatorTests.test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary) ... ok
test_seal_internal_references_and_inventory_types (__main__.EvaluatorTests.test_seal_internal_references_and_inventory_types) ... ok
test_signed_manifest_pin_with_consistent_bundle_resealing (__main__.EvaluatorTests.test_signed_manifest_pin_with_consistent_bundle_resealing) ... ok

----------------------------------------------------------------------
Ran 20 tests in 256.070s

OK
```

### Reproducible scratch-only guard mutants

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/mutate_round2.py
```

The committed runner contains exact one-occurrence substitutions and isolated copied-fixture layouts. Full command-by-command output is `round-2-mutant-output.txt`; it is included in the digest table below.

## Frozen rev9 bytes

The following check passed with manifest SHA `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`:

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py
```

**126/131 literal file matches, 5/5 shared-file owned-extract groups valid, 131/131 unchanged against baseline e2b94e3c.** Nineteen of 20 owned extracts match literally; the only state-only normalization is the already-permitted HD9 transition under round-9 condition 2. Full JSON is `final-evaluator/reviews/round-2-frozen-output.json` under stage3-drafts. No new change to any signed file occurred.

## Current per-file SHA-256s

Paths below are relative to this artefact folder. Historical review/evidence records are included and untouched. This response excludes its own hash to avoid circularity; delivery reports that hash separately.

| File | SHA-256 |
|---|---|
| `README.md` | `1acc107a64e1306fbe346528ab8e2f46760cf37ff7655532a5af08ed2162c30f` |
| `evaluator.py` | `940480b9190971549acb2abd20cd9853c67a377a45cf3b1ab2ccd927d96f7eed` |
| `provenance.py` | `5742d827f9e080a9934b39f94670cdecaba8bf74c6b58bcfd65d3a230e8c3932` |
| `test_evaluator.py` | `9948edea3d09902c6609362fc13cbeb010813097526a1b1d023859369ea60703` |
| `check_frozen.py` | `8130568a8e3e49fdf420c110bc9825ad20f483237f1f8745a2d96fa4009a56a1` |
| `reviews/round-1.md` | `d522104c10699ae05bf30514e1535feb556eddfc77802f3dd6d8ecfc48206388` |
| `reviews/round-1-response.md` | `9f708f5b07e052690c44372fc7e271d0da737b71698b2626ee88d83cfd3ceb62` |
| `reviews/round-1-test-output.txt` | `351158c1b87b040c9ad36615984ceb927190ebe5ebe5316bf9aedf131844fe13` |
| `reviews/round-1-mutant-output.txt` | `e37918a439014b1f6d8cecca2a329a82bdd5e15de593cb3dbf8f60673e7d961b` |
| `reviews/round-1-frozen-output.json` | `2c5d8a155e57cdd60a2b6ea295e9c0c1a2453acccf0ce662a60e711a85df50d0` |
| `mutate_round2.py` | `e09b53c2a19ec2f09c207bfbec0c08fd633b932839f4ecff334d860c2a3191a9` |
| `reviews/round-2.md` | `301bb41f025f2f592b9907a87a725fb645550a8ef491abd5496bab2e2b0e00a4` |
| `reviews/round-2-test-output.txt` | `fe20fd766c02f16537145bce289c5a7a7447d1d7004b215931a15999e4277dab` |
| `reviews/round-2-mutant-output.txt` | `6a9747089ee3dd0736833941f3fcf1a36401843ee9bda32765ef2fdfa4478302` |
| `reviews/round-2-frozen-output.json` | `2c5d8a155e57cdd60a2b6ea295e9c0c1a2453acccf0ce662a60e711a85df50d0` |
