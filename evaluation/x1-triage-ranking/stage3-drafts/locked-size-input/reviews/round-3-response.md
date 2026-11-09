# Locked-size input — round-3 remediation response

Date: 2026-11-09. **Draft only; submitted for scoped re-review, not self-signed off.** Implements the five findings in `round-3.md` and the three R2-02 residual paths. No preparation action, authentic inventory, binding remeasurement, OC, scan, model call or real repository data was used. The signed-off final evaluator is unchanged, including its `INPUTS` surface.

## Issue → fix → catching test

Paths below are relative to this artefact.

| Issue | Implementation | Catching test (`test_size_input.py`) |
|---|---|---|
| R3-01, R2-02 residual | `locked_commands.py:225–285` compares every component start/end closure with the fit closure and preserves all ten extracted values, exact `closure mismatch` and evidence digest; `:437–466` publishes completed-invalid create-once before nonzero exit. Input, host and interpreter mismatches still refuse. | `:426` `test_cross_component_closures_publish_then_nonzero`: A gate, flagged A→B benchmark, B calibration/fit, plus changes between commands; output exists, all ten values/classification/digest retained, nonzero. `:584` keeps mismatch refusals. |
| R3-02, R2-02 residual | `locked_commands.py:388` repeatable `--prior`; `:311` reads/parses/hashes each supplied file once; `:320–370` requires cumulative order, same input/inventory/host/runtime, exact qualifying cause/completion/evidence, and preserves prior causes; `:424–432` binds prior digests and appends only the new run. README items 7–8 require every prior fit. | `:453` `test_prior_cli_preserves_maxima_and_reads_once` invokes fit with a 9,000-byte failed prior and 3,000-byte rerun; the actual accepted evaluator sees both and returns maximum 9,000. `:484` refuses unflagged/different inventory/host and malformed prior sequences; `:562` preserves old causes when fit itself fails at its end. |
| R3-03, R2-02 residual | `locked_commands.py:116` tolerant end-source snapshot and `:437–466` input-end read retain null digests and exact OSError details, then publish completed-invalid and fail. No invented digest. | `:526` `test_missing_or_unreadable_end_snapshots_publish`: disposable source and input are each actually deleted; separate cases inject `PermissionError`. All publish/classify/nonzero and retain create-once evidence. |
| R3-04 | `test_size_input.py:613–648` heterogeneous fabricated fixture; `locked_commands.py:255–275` operation maxima. Every list maximum lies outside its first row, complete/full512 CPU maxima occur in different worker executions, calibration-frame CPU exceeds fixed-key CPU, gate CPU is nontrivial. | `:307` `test_extraction_uses_locked_maxima_not_slopes` asserts all ten hand answers. `round3_evidence.py:15–29` independent constants; `:41–72` assertion-kills minimum, first-row, wrong full512 selector, missing gate divisor and missing calibration-frame mutants. |
| R3-05 | `locked_commands.py:243–245` rejects component optimization other than integer 0 and version other than 3.11.11; `:398–399` refuses optimized invocation before input measurement. | `:584` `test_component_input_host_interpreter_and_flags_refused` includes optimize 1/version drift; `:600` `test_optimized_gate_subprocess_refused_before_measurement` actually invokes `-O`, checks nonzero and no output. |

## Commands and complete evidence

From the worktree root, `source .tools/activate`; every invocation sets `PYTHONDONTWRITEBYTECODE=1`. Numerical interpreter: `.tools/venv/bin/python` (3.11.11).

```sh
D=evaluation/x1-triage-ranking/stage3-drafts/locked-size-input
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/test_size_input.py" -v
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/round3_evidence.py" --output "$D/reviews/round-3-check-output.json"
PYTHONDONTWRITEBYTECODE=1 X1G_CLI_OUTPUT="$D/reviews/round-3-cli-evidence" .tools/venv/bin/python "$D/test_cli.py" -v
```

Output destinations are create-once; reproduction needs fresh paths. Full unit stdout/stderr is `round-3-test-output.txt`; full focused-check stdout and JSON are `round-3-check-output.txt`/`.json` (not an excerpt). The focused check reports **5/5 assertion-killed mutants**. Its five extraction wall times measure fabricated-record extraction only, not scientific CPU fit.

The actual CLI run retained every complete JSON output in `round-3-cli-evidence/`, including all benchmark trials and `cli-roundtrip.json` (exact argv, exit codes, stdout/stderr and evaluator-only resource result). It runs 100,000 gate redraws, eight cells × 2,000 trials × workers 1/2 = **32,000 actual helper trials**, B=1,999, actual locked-vector construction/calibration/enumeration and RSS. The unit suite separately performs the same full-count helper exercise. No reduced scientific mode or mocks were used for these integrations. Final receipts bind controller source `9fe453651816394a88272a98208f7e484c8bd73e895535638c68c273fc56478c`.

### Full final unit-suite output

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
test_component_input_host_interpreter_and_flags_refused (__main__.RoundThreeTests.test_component_input_host_interpreter_and_flags_refused) ... ok
test_cross_component_closures_publish_then_nonzero (__main__.RoundThreeTests.test_cross_component_closures_publish_then_nonzero) ... ok
test_fit_end_failure_only_marks_new_run (__main__.RoundThreeTests.test_fit_end_failure_only_marks_new_run) ... ok
test_missing_or_unreadable_end_snapshots_publish (__main__.RoundThreeTests.test_missing_or_unreadable_end_snapshots_publish) ... ok
test_optimized_gate_subprocess_refused_before_measurement (__main__.RoundThreeTests.test_optimized_gate_subprocess_refused_before_measurement) ... ok
test_prior_cli_preserves_maxima_and_reads_once (__main__.RoundThreeTests.test_prior_cli_preserves_maxima_and_reads_once) ... ok
test_prior_validation_and_cumulative_order (__main__.RoundThreeTests.test_prior_validation_and_cumulative_order) ... ok
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
Ran 30 tests in 596.312s

OK
```

### Full actual CLI test output

```text
test_full_cli_sequence_and_provenance_roundtrip (__main__.SyntheticCLITests.test_full_cli_sequence_and_provenance_roundtrip) ... ok

----------------------------------------------------------------------
Ran 1 test in 593.939s

OK
```

The heterogeneous hand answers are: construction 6/3,000 = .002; requested base 9/3,000 = .003; calibration 12/3,000 = .004; enumeration 4; gate 25,000/100,000 = .25; complete .03; full512 .07; construction/calibration/trial RSS 3,000/4,000/5,000. The focused receipt records all ten independently and all five mutant hashes/results.

## Preservation and review choices

The unchanged evaluator's frozen checker passed: 126/131 literal file matches, 5/5 owned-extract groups valid, 19/20 extracts literal with only allowed HD9 normalization, 131/131 unchanged against baseline `e2b94e3c`; manifest SHA `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`. Its full output is in the controller's `reviews/round-3-frozen-output.json`, outside the accepted evaluator. `../oc-controller/reviews/round-3-evaluator-preservation.json` confirms all 17 accepted-evaluator files unchanged, including the new sign-off receipt.

No owner choices are deferred. Every cumulative prior fit is mandatory rather than selecting only a recent payload; invalid prior bindings fail closed. Historical synthetic receipts remain separate non-selectable fixtures. No evaluator contract or signed byte was changed. Acceptance and any preparation/run authority still require the prescribed separate reviews.

## Recomputed per-file SHA-256

All files of this artefact, including retained historical evidence, are listed below except this response itself (its self-digest is delivered with the commit head).

| File | SHA-256 |
|---|---|
| `README.md` | `92666d5a955a9906b8cbdf21be0b659b7a7feb3b2c3cbe33391b8395e8d9fed5` |
| `locked_commands.py` | `d086af364594f779bd3546f9fa018dda66c412cdf729a86a04f25f62c681542c` |
| `population.py` | `a32e33cd8061940067d94b56cead9a85608a3a07c79b8a9b89c248482586f0c6` |
| `reviews/round-1-known-answer-output.txt` | `ec746d03798e5c5d709b1fffa2808c97838d468eabc6169c4c8320285a252a95` |
| `reviews/round-1-response.md` | `00cce4534953a9bbafd3b1f2d3bb83ff73cdf80bb08dd97cccc66fd1beaf3986` |
| `reviews/round-1-test-output.txt` | `bfccce29e41cee76006bfec9e86e025d04f8cdc5adf07fa502a13301f392b25f` |
| `reviews/round-1.md` | `0af23777d7a4c8bc81e152c95581084ad3bcf194a759460d79d25a05898b9fc8` |
| `reviews/round-2-cli-evidence/benchmark.json` | `236fb26e1740e9abbce26b20da7560a0f63a4d2c70fcb18e70088b90dfdd4ae5` |
| `reviews/round-2-cli-evidence/calibration.json` | `a2d519e9165dbc339ebe9ac824213de75efe5305d938a524e0d41f272eb7bb94` |
| `reviews/round-2-cli-evidence/cli-roundtrip.json` | `ced9da9ec9ada650f4dde19d8d02547527078fcd0d0cf258779d4079d56ec1c1` |
| `reviews/round-2-cli-evidence/fit.json` | `8473d78f1af5a8d1394d8e157a706bb7a00a4692aad4d6121b0fa93dadd63e36` |
| `reviews/round-2-cli-evidence/gate.json` | `91523d6f344caab731dd2af7d529c5fd919e42518641290680d511eb4388046b` |
| `reviews/round-2-cli-evidence/synthetic-input.json` | `6f8880093a21621ba32fdfc2df2441fa4c8c4e18a84082c1501a194ac23e33c6` |
| `reviews/round-2-cli-final-evidence/benchmark.json` | `d62f4c27ff416cb55186252ddc0dc86e45adff13ec5949dabf20d106bde7cfa8` |
| `reviews/round-2-cli-final-evidence/calibration.json` | `55c9542f3c4136193d9d45f168a06c510ba76c691e35d8b80d612ea6998e679e` |
| `reviews/round-2-cli-final-evidence/cli-roundtrip.json` | `74aaf2e421f35f263cf9955904a334e5b744d6f84cb979edfd67c5733766cf7c` |
| `reviews/round-2-cli-final-evidence/fit.json` | `e76828a035cc402d8c4c7790a5c14cbfb8d456b4d19cce095401411c9231c31c` |
| `reviews/round-2-cli-final-evidence/gate.json` | `0ab5cd09d4ebf4e2cc80a1edf51ed42577712920a669f888c4ccf7cdf3b1c469` |
| `reviews/round-2-cli-final-evidence/synthetic-input.json` | `6f8880093a21621ba32fdfc2df2441fa4c8c4e18a84082c1501a194ac23e33c6` |
| `reviews/round-2-cli-final-test-output.txt` | `fb7b075e7a0abc8def566e6ca5449c23fc6359c7b374a08da389670bd9ce46f3` |
| `reviews/round-2-cli-test-output.txt` | `01d7bb31333d2f2eab01a38b7263e652da642ecd0cac5016534801eb94a227ed` |
| `reviews/round-2-response.md` | `bbcba191886038cd4d586d18d13d72ebbada4a6cfe3d6965d39ea6b8c002fc9b` |
| `reviews/round-2-test-output.txt` | `1d89aa852e39f51d90fcdc3f85b78fd305f62955d9a5dc9acc778e1fd51f519d` |
| `reviews/round-2.md` | `ef1f7b6e46423ef978feedbb765ab1df3d29562eb5b55c69769a0df9ec3d69ff` |
| `reviews/round-3-check-output.json` | `3adcc0e4e1b70a32f70f5ed81af08cd20785099ec36379c86d01ae0cdccefa1d` |
| `reviews/round-3-check-output.txt` | `c7143a7940795358a6fd6548f3260fc0832c6dd58268f94f736e0bf2038e489f` |
| `reviews/round-3-cli-evidence/benchmark.json` | `00ad9c731be589c5c0da73abb461221e75fcefb9cea7ee176f9ce7f7929b6103` |
| `reviews/round-3-cli-evidence/calibration.json` | `60ebee3592e37c0f5a7a7e070e09d7c03ef80f9b460cf522bd9e235d7988ba08` |
| `reviews/round-3-cli-evidence/cli-roundtrip.json` | `ee0a505980474285e2251cc7b2f23a8ce8a6b4239d69678449441091ec7e0bbc` |
| `reviews/round-3-cli-evidence/fit.json` | `9b8571e4e4331cad00aa8e5bf5a8b64cde97366944193dce8222ecfadbf590e0` |
| `reviews/round-3-cli-evidence/gate.json` | `15a0f87ae2eda92fb544e8a60566c8a360a716cf17cdaee2ae3748dd7b1a9a3b` |
| `reviews/round-3-cli-evidence/synthetic-input.json` | `6f8880093a21621ba32fdfc2df2441fa4c8c4e18a84082c1501a194ac23e33c6` |
| `reviews/round-3-cli-test-output.txt` | `0172772554128669568689e0c1011ed08c528846bb37408e2524b3eb3b1a9a82` |
| `reviews/round-3-test-output.txt` | `09a48a081908429ad544d9ee0f91cd3db538d9b8413aa22a32a6ef09e3649634` |
| `reviews/round-3.md` | `e7a78488ad6e415b5dd151e0c37ad8558121114cd4153e70bb3651d8bd0bada1` |
| `round3_evidence.py` | `8ac46bba81686316bda40be38555bce8803cc79d935ba18d5c9c04997da2a300` |
| `size_input.py` | `579c5119a812f64ffaa574f4ec9d7f77d12b89f47d407511d6379ee7ce647cdc` |
| `test_cli.py` | `d6e22f609f5caa529d5ec2380edf7209aff4d10f323aeae968b161854c1e979d` |
| `test_size_input.py` | `c7da0a766bad346165af21c1e0cca1c62f0fefa816c59543774f23835f3ad154` |
