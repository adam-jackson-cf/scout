# locked-size-input: round-2 remediation response

**Status: implemented and exercised; submitted for scoped re-review, not council acceptance.** Responds to every R2 required fix and each named R1 residual in the byte-preserved `round-2.md`. All work is synthetic-only and pre-clock. No preparation action (a)–(e), real repository data, candidate fetch, bootstrap/attestation, binding implementation, identifier review, authentic sealing/inventory, scans or model calls occurred.

## Issue → fix → catching regression

| Issue | Fix | Implementation | Catching test |
|---|---|---|---|
| R2-01 | Stable four-field host; start/end load/process snapshots and executable/digest/optimize provenance; evaluator-consumable inventory/run records; read/hash/parse the same evidence bytes. | `locked_commands.py:161` (`host`), `locked_commands.py:169` (`host_load`), `locked_commands.py:175` (`interpreter`), `locked_commands.py:215` (`measurement_run`), `locked_commands.py:286` (`locked_inventory`), `locked_commands.py:293` (`read_evidence`) | `test_size_input.py:385` (`test_evaluator_provenance_roundtrip`), `test_size_input.py:347` (`test_evidence_single_read_hashes_parsed_bytes`), `test_cli.py:19` (`test_full_cli_sequence_and_provenance_roundtrip`) |
| R2-02 | Completed closure-invalid evidence is published create-once with both closures and full values, then raises nonzero; its run classification remains included in per-input maxima. | `locked_commands.py:354` (`finish`), `locked_commands.py:215` (`measurement_run`) | `test_size_input.py:354` (`test_closure_failure_published_create_once_then_nonzero`), `test_size_input.py:385` (`test_evaluator_provenance_roundtrip`) |
| R2-03 | Removed benchmark method option; fixed studentized execution and exact frozen cells/workers/full trial-count validation before run extraction. | `locked_commands.py:181` (`benchmark_mode`), `locked_commands.py:270` (`validate_benchmark`), `locked_commands.py:308` (`main`) | `test_size_input.py:323` (`test_frozen_method_and_cell_set_refused`), `test_cli.py:19` (`test_full_cli_sequence_and_provenance_roundtrip`) |
| R2-04 | Deleted duplicate locked_fit and all stop decisions/binding hours; fit only extracts the evaluator input/run convention. | `locked_commands.py:215` (`measurement_run`), `locked_commands.py:308` (`main`) | `test_size_input.py:307` (`test_extraction_uses_locked_maxima_not_slopes`), `test_size_input.py:385` (`test_evaluator_provenance_roundtrip`) |
| R2-05 | Estimator calls controller.wilson; interior 3/100000 fixture checks independent bounds and ceiling 133333334. | `size_input.py:120` (`estimate_p_gate`), `size_input.py:25` (`controller`) | `test_size_input.py:217` (`test_interior_gate_wilson_and_ceiling`) |
| R2-06 / R1-05 residual | Direct record mutations reach constructor guards for repository count/field sets and positions 1999/2000, rather than failing in the fixture builder. | `size_input.py:57` (`__init__`) | `test_size_input.py:91` (`test_direct_malformed_records_and_position_boundary`) |
| R2-07 | All six signed rev9 source hashes are pinned before import or measurement; each substituted expected digest is independently refused. | `locked_commands.py:39` (`pin_sources`), `locked_commands.py:99` (`locked_functions`) | `test_size_input.py:341` (`test_patched_signed_digest_refused_before_measurement`) |

## Conservative implementation choices and independent evidence

The evaluator is the sole resource decision convention; the size artefact extracts records only. The two retained full CLI executions are independent **synthetic validation fixtures**, not two real binding runs or a choice of measurement minima. Both receipts are preserved. The earlier controller closure is under `round-2-cli-evidence/` (605.024s); after the controller's late-arriving-fault latch was strengthened, a fresh full-count sequence under `round-2-cli-final-evidence/` passed in 606.215s and binds the final controller source. Neither output authorizes locked remeasurement. The unit helper run and each CLI integration executed 32000 actual helper trials (eight cells, 2000 trials, B=1999, workers one and two). The initial unit timing/RSS fixtures remain explicitly fabricated; only the retained CLI outputs are actual synthetic measurements.

The closure-failure regression modifies an actual disposable source after synthetic measurement values are supplied to the publisher. It verifies completed-invalid evidence persists, both closures are retained and duplicate publication fails. A separate evaluator roundtrip proves completed-invalid inputs enter the maximum. No signed source is modified. The single-read regression uses a read-once-changing byte source and asserts the parsed value and digest describe that one read.

## Exact commands and complete suite output

Commands ran from `/Users/adamjackson/Projects/work/scout-x1g` after `source .tools/activate`, using pinned CPython 3.11.11. For CLI reproduction select a **new nonexistent destination**; the committed create-once receipts must not be overwritten.

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/test_size_input.py -v
```

Complete output, also `round-2-test-output.txt`:

```text
test_actual_helper_one_two_worker_science (__main__.LockedModeTests.test_actual_helper_one_two_worker_science) ... ok
test_closure_failure_published_create_once_then_nonzero (__main__.LockedModeTests.test_closure_failure_published_create_once_then_nonzero) ... ok
test_construction_calibration_consume_every_locked_size (__main__.LockedModeTests.test_construction_calibration_consume_every_locked_size) ... ok
test_evaluator_provenance_roundtrip (__main__.LockedModeTests.test_evaluator_provenance_roundtrip) ... ok
test_evidence_single_read_hashes_parsed_bytes (__main__.LockedModeTests.test_evidence_single_read_hashes_parsed_bytes) ... ok
test_extraction_uses_locked_maxima_not_slopes (__main__.LockedModeTests.test_extraction_uses_locked_maxima_not_slopes) ... ok
test_frozen_method_and_cell_set_refused (__main__.LockedModeTests.test_frozen_method_and_cell_set_refused) ... ok
test_patched_signed_digest_refused_before_measurement (__main__.LockedModeTests.test_patched_signed_digest_refused_before_measurement) ... ok
test_size_site_drift_refused (__main__.LockedModeTests.test_size_site_drift_refused) ... ok
test_trial_population_lengths_and_selected_ids (__main__.LockedModeTests.test_trial_population_lengths_and_selected_ids) ... ok
test_asymmetric_offsets_and_constant_zero (__main__.SizeTests.test_asymmetric_offsets_and_constant_zero) ... ok
test_constructor_run_count_and_size_boundaries (__main__.SizeTests.test_constructor_run_count_and_size_boundaries) ... ok
test_direct_malformed_records_and_position_boundary (__main__.SizeTests.test_direct_malformed_records_and_position_boundary) ... ok
test_equal_vectors_preserve_exact_population (__main__.SizeTests.test_equal_vectors_preserve_exact_population) ... ok
test_full_100000_key_estimates (__main__.SizeTests.test_full_100000_key_estimates) ... ok
test_gate_boundaries (__main__.SizeTests.test_gate_boundaries) ... ok
test_identity_digest_and_invalid_order (__main__.SizeTests.test_identity_digest_and_invalid_order) ... ok
test_immutable_snapshot (__main__.SizeTests.test_immutable_snapshot) ... ok
test_interior_gate_wilson_and_ceiling (__main__.SizeTests.test_interior_gate_wilson_and_ceiling) ... ok
test_mismatched_truth_mutant_raises (__main__.SizeTests.test_mismatched_truth_mutant_raises) ... ok
test_only_size_source_changed_in_kernel (__main__.SizeTests.test_only_size_source_changed_in_kernel) ... ok
test_planted_effect_null_and_truth_identity (__main__.SizeTests.test_planted_effect_null_and_truth_identity) ... ok
test_seed_and_heterogeneous_selection_known_answers (__main__.SizeTests.test_seed_and_heterogeneous_selection_known_answers) ... ok

----------------------------------------------------------------------
Ran 23 tests in 588.587s

OK
```

```sh
PYTHONDONTWRITEBYTECODE=1 X1G_CLI_OUTPUT=/Users/adamjackson/Projects/work/scout-x1g/evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/reviews/round-2-cli-evidence .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/test_cli.py -v
```

Complete output, also `round-2-cli-test-output.txt`:

```text
test_full_cli_sequence_and_provenance_roundtrip (__main__.SyntheticCLITests.test_full_cli_sequence_and_provenance_roundtrip) ... ok

----------------------------------------------------------------------
Ran 1 test in 605.024s

OK
```

```sh
PYTHONDONTWRITEBYTECODE=1 X1G_CLI_OUTPUT=/Users/adamjackson/Projects/work/scout-x1g/evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/reviews/round-2-cli-final-evidence .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/test_cli.py -v
```

Complete output, also `round-2-cli-final-test-output.txt`:

```text
test_full_cli_sequence_and_provenance_roundtrip (__main__.SyntheticCLITests.test_full_cli_sequence_and_provenance_roundtrip) ... ok

----------------------------------------------------------------------
Ran 1 test in 606.215s

OK
```

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
| `README.md` | `d9ba449bee07dddab0d0bbfc9e39d305385516beb1c0fb0050c005f3c831e28b` |
| `size_input.py` | `579c5119a812f64ffaa574f4ec9d7f77d12b89f47d407511d6379ee7ce647cdc` |
| `population.py` | `a32e33cd8061940067d94b56cead9a85608a3a07c79b8a9b89c248482586f0c6` |
| `locked_commands.py` | `6c33a8d0aee617f19ce086a57a25a3a66a0bd9826c95a6d310fc15b7481bc28f` |
| `test_size_input.py` | `3551c0af440d10fd6c439453495f37bc5ad48380f5d705d0347f2a3426176cc9` |
| `reviews/round-1.md` | `0af23777d7a4c8bc81e152c95581084ad3bcf194a759460d79d25a05898b9fc8` |
| `reviews/round-1-response.md` | `00cce4534953a9bbafd3b1f2d3bb83ff73cdf80bb08dd97cccc66fd1beaf3986` |
| `reviews/round-1-test-output.txt` | `bfccce29e41cee76006bfec9e86e025d04f8cdc5adf07fa502a13301f392b25f` |
| `reviews/round-1-known-answer-output.txt` | `ec746d03798e5c5d709b1fffa2808c97838d468eabc6169c4c8320285a252a95` |
| `test_cli.py` | `d6e22f609f5caa529d5ec2380edf7209aff4d10f323aeae968b161854c1e979d` |
| `reviews/round-2.md` | `ef1f7b6e46423ef978feedbb765ab1df3d29562eb5b55c69769a0df9ec3d69ff` |
| `reviews/round-2-test-output.txt` | `1d89aa852e39f51d90fcdc3f85b78fd305f62955d9a5dc9acc778e1fd51f519d` |
| `reviews/round-2-cli-test-output.txt` | `01d7bb31333d2f2eab01a38b7263e652da642ecd0cac5016534801eb94a227ed` |
| `reviews/round-2-cli-final-test-output.txt` | `fb7b075e7a0abc8def566e6ca5449c23fc6359c7b374a08da389670bd9ce46f3` |
| `reviews/round-2-cli-evidence/synthetic-input.json` | `6f8880093a21621ba32fdfc2df2441fa4c8c4e18a84082c1501a194ac23e33c6` |
| `reviews/round-2-cli-evidence/gate.json` | `91523d6f344caab731dd2af7d529c5fd919e42518641290680d511eb4388046b` |
| `reviews/round-2-cli-evidence/benchmark.json` | `236fb26e1740e9abbce26b20da7560a0f63a4d2c70fcb18e70088b90dfdd4ae5` |
| `reviews/round-2-cli-evidence/calibration.json` | `a2d519e9165dbc339ebe9ac824213de75efe5305d938a524e0d41f272eb7bb94` |
| `reviews/round-2-cli-evidence/fit.json` | `8473d78f1af5a8d1394d8e157a706bb7a00a4692aad4d6121b0fa93dadd63e36` |
| `reviews/round-2-cli-evidence/cli-roundtrip.json` | `ced9da9ec9ada650f4dde19d8d02547527078fcd0d0cf258779d4079d56ec1c1` |
| `reviews/round-2-cli-final-evidence/synthetic-input.json` | `6f8880093a21621ba32fdfc2df2441fa4c8c4e18a84082c1501a194ac23e33c6` |
| `reviews/round-2-cli-final-evidence/gate.json` | `0ab5cd09d4ebf4e2cc80a1edf51ed42577712920a669f888c4ccf7cdf3b1c469` |
| `reviews/round-2-cli-final-evidence/benchmark.json` | `d62f4c27ff416cb55186252ddc0dc86e45adff13ec5949dabf20d106bde7cfa8` |
| `reviews/round-2-cli-final-evidence/calibration.json` | `55c9542f3c4136193d9d45f168a06c510ba76c691e35d8b80d612ea6998e679e` |
| `reviews/round-2-cli-final-evidence/fit.json` | `e76828a035cc402d8c4c7790a5c14cbfb8d456b4d19cce095401411c9231c31c` |
| `reviews/round-2-cli-final-evidence/cli-roundtrip.json` | `74aaf2e421f35f263cf9955904a334e5b744d6f84cb979edfd67c5733766cf7c` |
