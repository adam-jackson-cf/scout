# Locked-size input — round-4 remediation response

**Draft; awaiting scoped zero-issue review.** Only the three prescribed residuals are addressed. The measurement producer and evaluator contract are unchanged. Historical reviews/receipts are preserved; no preparation action or real-data work occurred.

## Issue → targeted fix → catching test

| Issue | Implementation/documentation | Catching test |
|---|---|---|
| R4-01 (R3-02/R2-02 residual) | `README.md:63–78` mandates finalise-before-rerun: retain a component's published closure-mismatch output, execute remaining components, run fit on the original set, retain its completed-invalid fit as mandatory first prior, and never replace/rerun any component before it exists. Separate exact recovery command blocks cover a nonzero benchmark and a nonzero calibration. Existing `main --prior` behavior is reused unchanged. | `test_size_input.py:480` `test_calibration_failure_finalised_before_rerun`: fabricated calibration RSS 12,000, actual `finish` publishes at A→B and fails; original-set `main fit` publishes completed-invalid and fails; rerun `main fit --prior` succeeds with RSS 4,000 while accepted evaluator maximum remains 12,000. Both run classifications and exact retained prior are checked. |
| R4-02 (R3-04 residual) | `test_size_input.py:325–327` adds fixed-key-dominant second-row CPU 10, exceeding every frame maximum; expected requested-base CPU is exactly 10/3,000. Existing frame-dominant case remains. `round3_evidence.py:53–55` adds only the matching `fixed-key-dropped` mutant. | `test_extraction_uses_locked_maxima_not_slopes` now exercises both selecting branches. Focused evidence reports **6/6 assertion-killed mutants**, including `fixed-key-dropped`. |
| R4-03 (R3-01 residual) | No producer change: add the missing fit-time closure boundary case through `main`. | `test_size_input.py:456` `test_agreeing_components_differ_from_fit_closure`: three unflagged agreeing A components, fit closure B; checks completed-invalid publication and nonzero, preserved-evidence SHA and all ten independent literal answers unchanged. |

## Commands and complete outputs

From the worktree root, `source .tools/activate`. Every command uses pinned Python 3.11.11 and disables bytecode writes.

```sh
D=evaluation/x1-triage-ranking/stage3-drafts/locked-size-input
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/test_size_input.py" -v
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/round3_evidence.py" --output "$D/reviews/round-4-check-output.json"
PYTHONDONTWRITEBYTECODE=1 X1G_CLI_OUTPUT="$D/reviews/round-4-cli-evidence" .tools/venv/bin/python "$D/test_cli.py" -v
```

Use fresh output paths for reproduction. Complete stdout/stderr is in `round-4-test-output.txt`, `round-4-check-output.txt` and `round-4-cli-test-output.txt`; focused JSON is `round-4-check-output.json`. All six complete CLI receipts (including the unabridged benchmark JSON, exact argv, return codes and evaluator roundtrip) are in `round-4-cli-evidence/`. They use 100,000 keyed draws, eight cells × 2,000 trials × one/two-worker execution = 32,000 actual helper trials, B=1,999. The unit suite separately exercises the same full-count helper science. These are synthetic validation fixtures, never selectable authentic runs.

### Full unit output

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
test_agreeing_components_differ_from_fit_closure (__main__.RoundThreeTests.test_agreeing_components_differ_from_fit_closure) ... ok
test_calibration_failure_finalised_before_rerun (__main__.RoundThreeTests.test_calibration_failure_finalised_before_rerun) ... ok
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
Ran 32 tests in 586.982s

OK
```

### Full CLI test output

```text
test_full_cli_sequence_and_provenance_roundtrip (__main__.SyntheticCLITests.test_full_cli_sequence_and_provenance_roundtrip) ... ok

----------------------------------------------------------------------
Ran 1 test in 602.705s

OK
```

No failures occurred in this round's checks. Fixed-key extraction was tested without changing the signed numerical instructions, fit arithmetic, prior schema or any evaluator bytes.

## Immutable boundaries

Unchanged `../final-evaluator/check_frozen.py` passed, with full output stored outside that artefact as `../oc-controller/reviews/round-4-frozen-output.json`: 126/131 literal file matches, all five owned-extract groups valid, 19/20 literal extracts with only permitted HD9 normalization, 131/131 unchanged against baseline `e2b94e3c`. Manifest SHA remains `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`. All 17 accepted-evaluator files retain their round-3 preservation digests (`../oc-controller/reviews/round-4-evaluator-preservation.json`).

No new owner choice or API surface is introduced. A new scoped council round is required; none is requested by this writer. No preparation or run authority is granted.

## Recomputed per-file SHA-256

All artefact files except this response itself; its self-digest accompanies the commit head. Historical files are included unchanged.

| File | SHA-256 |
|---|---|
| `README.md` | `7b9070c90a9e356dc3a23a777bbe5ca80096018d80b45637a7d62659f636981c` |
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
| `reviews/round-3-response.md` | `812ced7d8cb7da492c49f1d8a13a0cb759047dfbffcf91d20a29cdedd23883ff` |
| `reviews/round-3-test-output.txt` | `09a48a081908429ad544d9ee0f91cd3db538d9b8413aa22a32a6ef09e3649634` |
| `reviews/round-3.md` | `e7a78488ad6e415b5dd151e0c37ad8558121114cd4153e70bb3651d8bd0bada1` |
| `reviews/round-4-check-output.json` | `652f147fd755a0c70891ad709841dd1d63ae337afce904808e5d64302e40eef0` |
| `reviews/round-4-check-output.txt` | `5f18689a3d129772ee46a7efe08bb964bba14d54219df9b6617965978db40bb2` |
| `reviews/round-4-cli-evidence/benchmark.json` | `62309dc7b1dd7f2e4709e4643973f0ce8efc7c3256dd545a93f51d197255d43c` |
| `reviews/round-4-cli-evidence/calibration.json` | `5ad5b3f5c3158b0af0a2e9eb205a8f1321c3c0b7410dc61d400cdf3ab3385dce` |
| `reviews/round-4-cli-evidence/cli-roundtrip.json` | `edf6f8464203e2a3f94d5c6eba126eb43adec6f681daec8b9eeb2273136cbdd6` |
| `reviews/round-4-cli-evidence/fit.json` | `f0423de61d004ade3b373840a22ca68940b3818aac0b9d6d08a24704490a05c1` |
| `reviews/round-4-cli-evidence/gate.json` | `427f16da7600308c884a00c22db6918eadc3221a64255ec6e5152383e8b78036` |
| `reviews/round-4-cli-evidence/synthetic-input.json` | `6f8880093a21621ba32fdfc2df2441fa4c8c4e18a84082c1501a194ac23e33c6` |
| `reviews/round-4-cli-test-output.txt` | `0c470baaa070ac6c05f531cfb4ede3aa0c624112747d68fadbe10478df5975b4` |
| `reviews/round-4-test-output.txt` | `ef5a7da7fce45b847c0a682ba609fbd85ed569db5d3c1c56508dbae510b2b0d4` |
| `reviews/round-4.md` | `0f33e54581c30efde196199445fdaa7409645ad7c63798870eca64277ce50dd9` |
| `round3_evidence.py` | `3d67fefb4748effb3b317efe8c796e39b8b7b0295a7cf0aa4183cdd1ae4e2ad8` |
| `size_input.py` | `579c5119a812f64ffaa574f4ec9d7f77d12b89f47d407511d6379ee7ce647cdc` |
| `test_cli.py` | `d6e22f609f5caa529d5ec2380edf7209aff4d10f323aeae968b161854c1e979d` |
| `test_size_input.py` | `04ce64997ca13f955b1ccdbab49677288ef39e8bc61373a39d5b40659568bece` |
