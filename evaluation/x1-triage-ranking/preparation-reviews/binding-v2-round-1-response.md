# Binding v2 round 1 — response

Date: 2026-11-09. **All five X1B2 source/document fixes and identifiers R7-01 preferred fix (a) are supplied for re-review. No sign-off or adoption is claimed.** No real preparation adoption, rehearsal, canonical acceptance, candidate scan, model CLI, Infisical command or network fetch was run. Tests use temporary synthetic records/repositories, including the real adoption implementation. Product source, `preparation.py`, `test_core.py`, the signed config, v1 recipe/policy and clock were not edited. Files remain uncommitted.

## Findings and catching evidence

### X1B2-R1-01 (Major) and identifiers R7-01 — fixed with preferred (a)

`execution/binding/acceptance.py::reviewed_inputs` now requires the immutable pre-adoption amendment review to bind **the proposal**, recipe, amendment and every binding source pin. It no longer requires the nonexistent adopted path or future source manifest in that receipt. Separately, it requires:

- the exact adopted path/digest in `config.amendments`;
- adopted SHA equal to the reviewed proposal SHA;
- that row's `adoption_review` equal to the exact pin of `OUT/reviews/binding-v2-amendment.json`.

The post-adoption `reviews/implementation.json` independently binds the source manifest and runtime; binding source pins must be a subset of that manifest. Neither a proposal-only config nor an alternate same-digest review path substitutes for adoption provenance. **No Core change is needed or made.**

Catching tests: new `tests/test_binding_adoption.py::BindingAdoptionIntegrationTests` runs real `preparation.adopt`, real `load_config`, real `review_receipt`/pins and the **entire** `reviewed_inputs` through runtime validation. The only environment redirection is the controller's `EXEC` constant to a private temporary tree and CLI-equivalent `sys.argv` for the direct adopt function call. No pin, authority validator, adopt, load or reviewed_inputs implementation is mocked. The pre-adoption receipt/record bytes remain identical throughout; adopted copy and source manifest are absent before adoption; the manifest is generated only afterward and contains the adopted file.

Tests passing in the full suite:
- `test_real_adopt_load_and_entire_binding_authority_chain`
- `test_proposal_digest_missing_refuses_real_adoption`
- `test_proposal_path_missing_refuses_binding_even_with_reviewed_digest`
- `test_one_source_pin_missing_from_amendment_review_refuses_binding`
- `test_adopted_sha_mismatch_refuses_load_and_binding`
- `test_loaded_configuration_requires_exact_adopted_row`
- `test_adoption_review_mismatched_pin_refuses_binding`

The old mocked `AdoptedAmendmentTests` that concealed the circularity was removed, replaced by these stronger real-file cases. This exercises adopt → binding-authority validation, **not a real seal**; no full real preparation run is claimed.

The amendment and current review request now give this exact, Main-only procedure, after a zero-issue council round and materialization of the immutable `{scope, record, reviewed}` authority receipt:

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/preparation.py adopt --config /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/run-config.json --amendment /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/run-config-amendments-proposed/001-binding-v2.json --review /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/reviews/binding-v2-amendment.json
```

Do not copy by hand, write adoption timestamps by hand or edit the authority receipt after adoption. The command writes the adopted copy and timestamp/operation/adoption receipt. Keep proposal and final signed-off writer bytes/paths unchanged. Only **after adoption** produce the source manifest and implementation/runtime reviews, then the reviewed scratch rehearsal; its own report review precedes canonical synthetic and Django acceptance. The command above is an instruction, not an exercised real action.

### X1B2-R1-02 (Minor) — corrected residual scope

Both `preparation/model-free-binding-check-v2.json::residual_limitation` and the amendment's “What the proof answers” section now explicitly distinguish:

1. Sandboxed scanners: hash-pinned binary trust plus fail-closed analyzer network sandbox. The sandbox rebuilds PATH; controller stand-ins do **not** reach scanner descendants.
2. **Unsandboxed frontend/Node/Clawpatch dependency tree and git/go-list children, with network available**: trust rests on the cited non-transitive audit, pinned closure and configured stand-ins, not on the analyzer sandbox or global observation. Unexpected descendants/absolute-path launches are no longer globally observed.
3. Positive-control liveness is proved **only for Scout's contextual seam**. Clawpatch provider routes are configured/statically checked but not exercised by the control. Their exclusion rests on empty opt-ins, false egress and heuristic mapping with a null provider; stand-ins are defense in depth.

The current review request repeats this scope, rather than retaining an overstated summary. This is the requested **text-only fix**; no runtime behavior or new sandbox was introduced, and no behavioral test of Node/scanner model absence is claimed. Catching evidence is the explicit revised recipe/amendment text and their newly recomputed digests below; existing real contextual-stub tests continue to prove only their stated seam.

### X1B2-R1-03 (Minor) — every enumerated boundary tested

New `tests/test_binding_boundaries.py` supplies table regressions:

| Boundary | Catching test |
|---|---|
| Controller key presence, including empty value, at reviewed_inputs; before rehearsal output reservation; before each bound/reference/control launch (execution never reached) | `ControllerKeyBoundaryTests.test_reviewed_inputs_checks_presence_not_truthiness`, `test_rehearsal_entry_checks_presence_before_output_reservation`, `test_every_launch_checks_key_presence_before_direct_execution` |
| Control count1, providers swapped/duplicated, attempts0 and wrong stub exit all make the positive flag false | `PredicateBoundaryTests.test_positive_control_false_boundaries` |
| A nonzero count on any bound/reference revision makes zero_model_attempts false | `PredicateBoundaryTests.test_each_bound_reference_attempt_count_must_be_zero` |
| A resolve-only unbound attempt refuses; nonzero bound exit refuses; genuine positive-control nonzero exit is tolerated | `PredicateBoundaryTests.test_resolve_attempt_and_exit_status_boundaries` |
| Django substituted tip, head^1 mismatch, noncanonical or failed objects, and non-passing synthetic receipt refuse | `DjangoFixtureBoundaryTests.test_django_manifest_refusal_table` |
| Equal synthetic parent/head finding sets refuse | `DjangoFixtureBoundaryTests.test_equal_synthetic_parent_head_issue_sets_refuse` |

Existing valid-zero/positive baselines remain; these are refusal/false-predicate tests, not weakened happy-path counts. External scanner routes are mocked, and no real scanner or model is invoked.

### X1B2-R1-04 (Minor) — synthetic dependency sealed in Django evidence

`fixture_manifest` pins the exact passing synthetic receipt it reads. The Django manifest and top-level `execute_fixture` evidence both carry **`synthetic_receipt: pin(OUT/binding/synthetic.json)`**. This is ordinary absolute path/hash evidence, so the existing recursive `verify_tree_pins` at seal rechecks it; no Core modification is necessary.

Catching test: `DjangoFixtureBoundaryTests.test_django_evidence_pins_exact_synthetic_receipt_and_detects_change` checks the precise manifest/top-level pin and then changes the temporary synthetic receipt, requiring `verify_tree_pins` to refuse. Thus a later reviewed recovery cannot silently substitute the dependency.

### X1B2-R1-05 (Minor) — current Core and all files re-pinned, checks fresh

The current unchanged Core bytes were hashed before the new suite and again for this manifest:

- `preparation.py`: `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427`
- `tests/test_core.py`: `94afb7638acbb0316cc019077a87d924fda0e665b752d3d57acdfadbd9657ace`

The request's stale Core/test rows and 164-test evidence summary have been replaced. Every binding file, **every fixture file**, new tests, shared authority/dependency and the 22 audited product seams were rehashed. The old check artifact is preserved as historical evidence, not reused as current proof. Fresh authorized results: **182 Python tests passed in 14.884s**, Go vet exit0 (no output), Go test exit0 (cached result honestly reported). The real adoption integration and current Core adoption/journal tests are in that fresh run, so the evidence cannot silently predate the adoption writer.

## Scope and remaining authority

The requested fixes do not change v1 signed bytes, product design, the 45-day clock, scientific outcomes/analysis or frame/dev authority. R-064/R-087 provenance remains as stated in the amendment, with Main owning the pending ledger row. Root/FDA is not a v2 prerequisite. Actual fixture trigger counts and real runtime lane outputs remain unmeasured until Main's reviewed scratch rehearsal. No real adoption, rehearsal or canonical acceptance is authorized by this response; council re-review must reach zero issues first.

The nonblocking suggestions beyond the requested adoption-procedure correction were not silently turned into new execution policies. In particular no new outer sandbox, timeout policy, extra receipt type or Core edit was added. The source-level heuristic/health and existing refusal semantics remain unchanged.

## Commands and complete captured output

All three commands ran with cwd `/Users/adamjackson/Projects/work/scout-x1prep`, after both independent synthetic-test slices landed. Go commands used the pinned activation with network/module mutation disabled. Capture contains exact stdout/stderr separately; empty streams are explicitly marked. Machine-readable duplicate: [binding-v2-round-1-checks.json](binding-v2-round-1-checks.json), SHA-256 `8d567ea225c388c19bfb6e82c7e93076f2f5c27e1f26572cf8b2ff3a24ce10e4`.

### Check 1

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py' -v
```

Exit code: **0**.

stdout (complete, empty):

stderr (complete):

```text
test_both_revisions_resolved_before_any_scanner (test_binding.CompletionTests.test_both_revisions_resolved_before_any_scanner) ... ok
test_bound_policy_refuses_each_input_boundary (test_binding.CompletionTests.test_bound_policy_refuses_each_input_boundary) ... ok
test_calledprocesserror_always_terminates_refused_receipt (test_binding.CompletionTests.test_calledprocesserror_always_terminates_refused_receipt) ... ok
test_invalid_reference_head_resolution_prevents_every_scanner (test_binding.CompletionTests.test_invalid_reference_head_resolution_prevents_every_scanner) ... ok
test_resolved_policy_must_match_later_bytes_and_pin (test_binding.CompletionTests.test_resolved_policy_must_match_later_bytes_and_pin) ... ok
test_all_six_flags_derived_from_complete_evidence (test_binding.DeterministicProofTests.test_all_six_flags_derived_from_complete_evidence) ... ok
test_config_proof_and_actual_lanes_each_required (test_binding.DeterministicProofTests.test_config_proof_and_actual_lanes_each_required) ... ok
test_layout_and_identity_cannot_pass_vacuously (test_binding.DeterministicProofTests.test_layout_and_identity_cannot_pass_vacuously) ... ok
test_actual_frontend_launch_is_direct_and_private (test_binding.EnvironmentAndSeamTests.test_actual_frontend_launch_is_direct_and_private) ... ok
test_missing_attempt_log_refuses_instead_of_passing_zero (test_binding.EnvironmentAndSeamTests.test_missing_attempt_log_refuses_instead_of_passing_zero) ... ok
test_private_env_stubs_all_provider_paths_and_no_outer_sandbox (test_binding.EnvironmentAndSeamTests.test_private_env_stubs_all_provider_paths_and_no_outer_sandbox) ... ok
test_static_provider_binary_seams (test_binding.EnvironmentAndSeamTests.test_static_provider_binary_seams) ... ok
test_stub_is_logging_only (test_binding.EnvironmentAndSeamTests.test_stub_is_logging_only) ... ok
test_equal_length_roots_and_missing_head_state_directory (test_binding.FixtureAndParityTests.test_equal_length_roots_and_missing_head_state_directory) ... ok
test_exact_sets_health_multiplicity_coverage_and_durable_link (test_binding.FixtureAndParityTests.test_exact_sets_health_multiplicity_coverage_and_durable_link) ... ok
test_fixture_trigger_constructs_and_revision_difference (test_binding.FixtureAndParityTests.test_fixture_trigger_constructs_and_revision_difference) ... ok
test_path_escape_and_codegraph_refused (test_binding.FixtureAndParityTests.test_path_escape_and_codegraph_refused) ... ok
test_source_manifest_not_mistaken_for_filesystem_pin (test_binding.FixtureAndParityTests.test_source_manifest_not_mistaken_for_filesystem_pin) ... ok
test_actual_failing_stub_evidence (test_binding.ModelAttemptTests.test_actual_failing_stub_evidence) ... ok
test_positive_boundaries (test_binding.ModelAttemptTests.test_positive_boundaries) ... ok
test_zero_and_positive_logs (test_binding.ModelAttemptTests.test_zero_and_positive_logs) ... ok
test_canonical_gate_happens_before_lock_reservation_and_journal (test_binding.RehearsalTests.test_canonical_gate_happens_before_lock_reservation_and_journal) ... ok
test_rehearsal_main_never_opens_out_or_journal (test_binding.RehearsalTests.test_rehearsal_main_never_opens_out_or_journal) ... ok
test_rehearsal_rejects_canonical_escape_and_symlinks (test_binding.RehearsalTests.test_rehearsal_rejects_canonical_escape_and_symlinks) ... ok
test_reviewed_rehearsal_requires_current_pins_and_every_predicate (test_binding.RehearsalTests.test_reviewed_rehearsal_requires_current_pins_and_every_predicate) ... ok
test_wrong_or_missing_rehearsal_clock_refuses_before_scratch_writes (test_binding.RehearsalTests.test_wrong_or_missing_rehearsal_clock_refuses_before_scratch_writes) ... ok
test_adopted_sha_mismatch_refuses_load_and_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_adopted_sha_mismatch_refuses_load_and_binding) ... ok
test_adoption_review_mismatched_pin_refuses_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_adoption_review_mismatched_pin_refuses_binding) ... ok
test_loaded_configuration_requires_exact_adopted_row (test_binding_adoption.BindingAdoptionIntegrationTests.test_loaded_configuration_requires_exact_adopted_row) ... ok
test_one_source_pin_missing_from_amendment_review_refuses_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_one_source_pin_missing_from_amendment_review_refuses_binding) ... ok
test_proposal_digest_missing_refuses_real_adoption (test_binding_adoption.BindingAdoptionIntegrationTests.test_proposal_digest_missing_refuses_real_adoption) ... ok
test_proposal_path_missing_refuses_binding_even_with_reviewed_digest (test_binding_adoption.BindingAdoptionIntegrationTests.test_proposal_path_missing_refuses_binding_even_with_reviewed_digest) ... ok
test_real_adopt_load_and_entire_binding_authority_chain (test_binding_adoption.BindingAdoptionIntegrationTests.test_real_adopt_load_and_entire_binding_authority_chain) ... ok
test_every_launch_checks_key_presence_before_direct_execution (test_binding_boundaries.ControllerKeyBoundaryTests.test_every_launch_checks_key_presence_before_direct_execution) ... ok
test_rehearsal_entry_checks_presence_before_output_reservation (test_binding_boundaries.ControllerKeyBoundaryTests.test_rehearsal_entry_checks_presence_before_output_reservation) ... ok
test_reviewed_inputs_checks_presence_not_truthiness (test_binding_boundaries.ControllerKeyBoundaryTests.test_reviewed_inputs_checks_presence_not_truthiness) ... ok
test_django_evidence_pins_exact_synthetic_receipt_and_detects_change (test_binding_boundaries.DjangoFixtureBoundaryTests.test_django_evidence_pins_exact_synthetic_receipt_and_detects_change) ... ok
test_django_manifest_refusal_table (test_binding_boundaries.DjangoFixtureBoundaryTests.test_django_manifest_refusal_table) ... ok
test_equal_synthetic_parent_head_issue_sets_refuse (test_binding_boundaries.DjangoFixtureBoundaryTests.test_equal_synthetic_parent_head_issue_sets_refuse) ... ok
test_each_bound_reference_attempt_count_must_be_zero (test_binding_boundaries.PredicateBoundaryTests.test_each_bound_reference_attempt_count_must_be_zero) ... ok
test_positive_control_false_boundaries (test_binding_boundaries.PredicateBoundaryTests.test_positive_control_false_boundaries) ... ok
test_resolve_attempt_and_exit_status_boundaries (test_binding_boundaries.PredicateBoundaryTests.test_resolve_attempt_and_exit_status_boundaries) ... ok
test_codegraph_refused_even_via_link (test_closure.ClosureTests.test_codegraph_refused_even_via_link) ... ok
test_cycle_refused (test_closure.ClosureTests.test_cycle_refused) ... ok
test_explicit_external_root_counted (test_closure.ClosureTests.test_explicit_external_root_counted) ... ok
test_incomplete_retained_object_refused (test_closure.ClosureTests.test_incomplete_retained_object_refused) ... ok
test_inventory_preserves_links_and_files (test_closure.ClosureTests.test_inventory_preserves_links_and_files) ... ok
test_json_stream_rejects_partial (test_closure.ClosureTests.test_json_stream_rejects_partial) ... ok
test_launcher_bytes_never_normalized (test_closure.ClosureTests.test_launcher_bytes_never_normalized) ... ok
test_link_chain (test_closure.ClosureTests.test_link_chain) ... ok
test_missing_node_dependency_refused (test_closure.ClosureTests.test_missing_node_dependency_refused) ... ok
test_model_cli_never_launched (test_closure.ClosureTests.test_model_cli_never_launched) ... ok
test_native_missing_dependency_refused (test_closure.ClosureTests.test_native_missing_dependency_refused) ... ok
test_native_self_identity_not_dependency (test_closure.ClosureTests.test_native_self_identity_not_dependency) ... ok
test_npm_identity_not_just_version (test_closure.ClosureTests.test_npm_identity_not_just_version) ... ok
test_python_record_verified (test_closure.ClosureTests.test_python_record_verified) ... ok
test_python_tamper_refused (test_closure.ClosureTests.test_python_tamper_refused) ... ok
test_requirements_frozen_hashes (test_closure.ClosureTests.test_requirements_frozen_hashes) ... ok
test_requirements_unhashed_refused (test_closure.ClosureTests.test_requirements_unhashed_refused) ... ok
test_retention_deduplicates_and_preserves_original_mode (test_closure.ClosureTests.test_retention_deduplicates_and_preserves_original_mode) ... ok
test_unknown_external_link_refused (test_closure.ClosureTests.test_unknown_external_link_refused) ... ok
test_r101_duplicate_applicable_row (test_closure_review.ClosureReviewTests.test_r101_duplicate_applicable_row) ... ok
test_r101_real_marker_lines (test_closure_review.ClosureReviewTests.test_r101_real_marker_lines) ... ok
test_r101_unknown_marker_variable_and_operator (test_closure_review.ClosureReviewTests.test_r101_unknown_marker_variable_and_operator) ... ok
test_r102_executable_context_carried_through_rpath (test_closure_review.ClosureReviewTests.test_r102_executable_context_carried_through_rpath) ... ok
test_r102_missing_executable_context_refuses (test_closure_review.ClosureReviewTests.test_r102_missing_executable_context_refuses) ... ok
test_r103_nested_peer_keys_are_not_packages (test_closure_review.ClosureReviewTests.test_r103_nested_peer_keys_are_not_packages) ... ok
test_r104_internal_tarball_node_modules_pass (test_closure_review.ClosureReviewTests.test_r104_internal_tarball_node_modules_pass) ... ok
test_r104_stray_nested_dependency_refuses (test_closure_review.ClosureReviewTests.test_r104_stray_nested_dependency_refuses) ... ok
test_r104_stray_top_level_package_refuses (test_closure_review.ClosureReviewTests.test_r104_stray_top_level_package_refuses) ... ok
test_r105_consistently_substituted_installed_locks_refuse (test_closure_review.ClosureReviewTests.test_r105_consistently_substituted_installed_locks_refuse) ... ok
test_r106_git_agreement_and_helper_root (test_closure_review.ClosureReviewTests.test_r106_git_agreement_and_helper_root) ... ok
test_r106_git_disagreement_refuses (test_closure_review.ClosureReviewTests.test_r106_git_disagreement_refuses) ... ok
test_r107_reviewed_frontend_bytes_must_stay_signed (test_closure_review.ClosureReviewTests.test_r107_reviewed_frontend_bytes_must_stay_signed) ... ok
test_r107_unpinned_selected_source_refuses (test_closure_review.ClosureReviewTests.test_r107_unpinned_selected_source_refuses) ... ok
test_r108_own_closure_is_not_whole_execution_manifest (test_closure_review.ClosureReviewTests.test_r108_own_closure_is_not_whole_execution_manifest) ... ok
test_r109_interpreter_outside_inventory_refuses (test_closure_review.ClosureReviewTests.test_r109_interpreter_outside_inventory_refuses) ... ok
test_r109_relocated_runtime_refuses (test_closure_review.ClosureReviewTests.test_r109_relocated_runtime_refuses) ... ok
test_r109_resolved_path_must_equal_signed_path (test_closure_review.ClosureReviewTests.test_r109_resolved_path_must_equal_signed_path) ... ok
test_r109_unisolated_interpreter_refuses (test_closure_review.ClosureReviewTests.test_r109_unisolated_interpreter_refuses) ... ok
test_r109_wrong_running_interpreter_refuses (test_closure_review.ClosureReviewTests.test_r109_wrong_running_interpreter_refuses) ... ok
test_r110_nonexecutable_node_addon_is_inspected (test_closure_review.ClosureReviewTests.test_r110_nonexecutable_node_addon_is_inspected) ... ok
test_r111_deferred_count_refuses (test_closure_review.ClosureReviewTests.test_r111_deferred_count_refuses) ... ok
test_r111_deferred_empty_still_verifies_working_files (test_closure_review.ClosureReviewTests.test_r111_deferred_empty_still_verifies_working_files) ... ok
test_r111_deferred_field_refuses (test_closure_review.ClosureReviewTests.test_r111_deferred_field_refuses) ... ok
test_r111_deferred_nonempty_null_pass (test_closure_review.ClosureReviewTests.test_r111_deferred_nonempty_null_pass) ... ok
test_r111_deferred_set_refuses (test_closure_review.ClosureReviewTests.test_r111_deferred_set_refuses) ... ok
test_r111_go_metadata_planted_null_pass (test_closure_review.ClosureReviewTests.test_r111_go_metadata_planted_null_pass) ... ok
test_r111_go_replace_refuses (test_closure_review.ClosureReviewTests.test_r111_go_replace_refuses) ... ok
test_r111_go_sum_mismatch_refuses (test_closure_review.ClosureReviewTests.test_r111_go_sum_mismatch_refuses) ... ok
test_r111_recheck_added_file_refuses (test_closure_review.ClosureReviewTests.test_r111_recheck_added_file_refuses) ... ok
test_r111_recheck_entrypoint_resolution_drift_refuses (test_closure_review.ClosureReviewTests.test_r111_recheck_entrypoint_resolution_drift_refuses) ... ok
test_r111_recheck_mode_change_refuses (test_closure_review.ClosureReviewTests.test_r111_recheck_mode_change_refuses) ... ok
test_r111_recheck_planted_null_pass (test_closure_review.ClosureReviewTests.test_r111_recheck_planted_null_pass) ... ok
test_r111_recheck_retargeted_link_refuses (test_closure_review.ClosureReviewTests.test_r111_recheck_retargeted_link_refuses) ... ok
test_r111_recheck_writable_retained_object_refuses (test_closure_review.ClosureReviewTests.test_r111_recheck_writable_retained_object_refuses) ... ok
test_r201_fat32_uses_arm64_slice (test_closure_round2.RoundTwoTests.test_r201_fat32_uses_arm64_slice) ... ok
test_r201_fat64_uses_arm64_slice (test_closure_round2.RoundTwoTests.test_r201_fat64_uses_arm64_slice) ... ok
test_r201_java_is_non_native_without_suffix_filter (test_closure_round2.RoundTwoTests.test_r201_java_is_non_native_without_suffix_filter) ... ok
test_r201_misaligned_fat_refuses (test_closure_round2.RoundTwoTests.test_r201_misaligned_fat_refuses) ... ok
test_r201_out_of_bounds_fat_refuses (test_closure_round2.RoundTwoTests.test_r201_out_of_bounds_fat_refuses) ... ok
test_r201_thin_arm64 (test_closure_round2.RoundTwoTests.test_r201_thin_arm64) ... ok
test_r201_truncated_fat_refuses (test_closure_round2.RoundTwoTests.test_r201_truncated_fat_refuses) ... ok
test_r201_truncated_thin_refuses (test_closure_round2.RoundTwoTests.test_r201_truncated_thin_refuses) ... ok
test_r202_go_metadata_escaping_file_symlink_refuses (test_closure_round2.RoundTwoTests.test_r202_go_metadata_escaping_file_symlink_refuses) ... ok
test_r202_go_metadata_logical_source_null (test_closure_round2.RoundTwoTests.test_r202_go_metadata_logical_source_null) ... ok
test_r202_go_metadata_unpinned_inside_root_refuses (test_closure_round2.RoundTwoTests.test_r202_go_metadata_unpinned_inside_root_refuses) ... ok
test_r203_drifted_council_record_refuses (test_closure_round2.RoundTwoTests.test_r203_drifted_council_record_refuses) ... ok
test_r203_drifted_receipt_refuses (test_closure_round2.RoundTwoTests.test_r203_drifted_receipt_refuses) ... ok
test_r203_exact_selected_frontend_review_set (test_closure_round2.RoundTwoTests.test_r203_exact_selected_frontend_review_set) ... ok
test_r203_main_journal_binds_dedicated_receipt_and_record (test_closure_round2.RoundTwoTests.test_r203_main_journal_binds_dedicated_receipt_and_record) ... ok
test_r203_missing_dedicated_receipt_refuses (test_closure_round2.RoundTwoTests.test_r203_missing_dedicated_receipt_refuses) ... ok
test_r203_unrelated_reviewed_byte_refuses (test_closure_round2.RoundTwoTests.test_r203_unrelated_reviewed_byte_refuses) ... ok
test_r203_valid_receipt_and_council_record_are_pinned (test_closure_round2.RoundTwoTests.test_r203_valid_receipt_and_council_record_are_pinned) ... ok
test_r203_wrong_scope_refuses (test_closure_round2.RoundTwoTests.test_r203_wrong_scope_refuses) ... ok
test_r204_attest_producer_wires_helpers_and_review_inputs (test_closure_round2.RoundTwoTests.test_r204_attest_producer_wires_helpers_and_review_inputs)
Fake only external metadata; execute the real attestation orchestration. ... ok
test_r204_claw_assets_precede_pnpm_null (test_closure_round2.RoundTwoTests.test_r204_claw_assets_precede_pnpm_null) ... ok
test_r204_git_head_mismatch_refuses (test_closure_round2.RoundTwoTests.test_r204_git_head_mismatch_refuses) ... ok
test_r204_git_non_bin_refuses (test_closure_round2.RoundTwoTests.test_r204_git_non_bin_refuses) ... ok
test_r204_git_shim_refuses (test_closure_round2.RoundTwoTests.test_r204_git_shim_refuses) ... ok
test_r204_git_version_drift_refuses (test_closure_round2.RoundTwoTests.test_r204_git_version_drift_refuses) ... ok
test_r204_keg_git_version_head_null_and_command_identity (test_closure_round2.RoundTwoTests.test_r204_keg_git_version_head_null_and_command_identity) ... ok
test_r204_missing_git_core_refuses (test_closure_round2.RoundTwoTests.test_r204_missing_git_core_refuses) ... ok
test_r204_native_inventory_selection_addon_failure_caught (test_closure_round2.RoundTwoTests.test_r204_native_inventory_selection_addon_failure_caught) ... ok
test_r204_native_inventory_selection_null (test_closure_round2.RoundTwoTests.test_r204_native_inventory_selection_null) ... ok
test_r204_nonpassing_object_receipt_refuses (test_closure_round2.RoundTwoTests.test_r204_nonpassing_object_receipt_refuses) ... ok
test_r204_runtime_inventory_includes_keg_null (test_closure_round2.RoundTwoTests.test_r204_runtime_inventory_includes_keg_null) ... ok
test_r204_runtime_inventory_keg_escape_refuses (test_closure_round2.RoundTwoTests.test_r204_runtime_inventory_keg_escape_refuses) ... ok
test_r204_substituted_claw_refuses_before_pnpm (test_closure_round2.RoundTwoTests.test_r204_substituted_claw_refuses_before_pnpm) ... ok
test_adopted_amendment_without_review_refused_at_load (test_core.CoreTests.test_adopted_amendment_without_review_refused_at_load) ... ok
test_adoption_racing_journal_is_serialized (test_core.CoreTests.test_adoption_racing_journal_is_serialized) ... ok
test_amendment_chain_applies_in_order_and_preserves_original (test_core.CoreTests.test_amendment_chain_applies_in_order_and_preserves_original) ... ok
test_amendment_review_scope_and_exact_digest_required (test_core.CoreTests.test_amendment_review_scope_and_exact_digest_required) ... ok
test_clock_pinned_original_config_cannot_drift (test_core.CoreTests.test_clock_pinned_original_config_cannot_drift) ... ok
test_deadline_exact_seconds_and_leap_boundary (test_core.CoreTests.test_deadline_exact_seconds_and_leap_boundary) ... ok
test_edit_delete_or_rename_adopted_amendment_refused (test_core.CoreTests.test_edit_delete_or_rename_adopted_amendment_refused) ... ok
test_exception_classification (test_core.CoreTests.test_exception_classification) ... ok
test_exclusive_creation_preserves_original (test_core.CoreTests.test_exclusive_creation_preserves_original) ... ok
test_hand_written_and_backdated_adoption_receipts_refused (test_core.CoreTests.test_hand_written_and_backdated_adoption_receipts_refused) ... ok
test_hd9_only_state_date_normalized (test_core.CoreTests.test_hd9_only_state_date_normalized) ... ok
test_initialized_submodule_checks_working_bytes (test_core.CoreTests.test_initialized_submodule_checks_working_bytes) ... ok
test_initialized_submodule_mismatch_refused (test_core.CoreTests.test_initialized_submodule_mismatch_refused) ... ok
test_load_and_seal_accept_legacy_ab_and_d_before_adoption (test_core.CoreTests.test_load_and_seal_accept_legacy_ab_and_d_before_adoption) ... ok
test_merge_first_parent_and_offpath_refusal (test_core.CoreTests.test_merge_first_parent_and_offpath_refusal) ... ok
test_original_config_is_unchanged (test_core.CoreTests.test_original_config_is_unchanged) ... ok
test_promisor_and_extra_ref_refused (test_core.CoreTests.test_promisor_and_extra_ref_refused) ... ok
test_proposed_amendment_is_ignored (test_core.CoreTests.test_proposed_amendment_is_ignored) ... ok
test_recursive_pin_mismatch_and_missing (test_core.CoreTests.test_recursive_pin_mismatch_and_missing) ... ok
test_seal_deferred_rows_must_all_be_verified (test_core.CoreTests.test_seal_deferred_rows_must_all_be_verified) ... ok
test_seal_journal_validator_accepts_pre_and_post_adoption (test_core.CoreTests.test_seal_journal_validator_accepts_pre_and_post_adoption) ... ok
test_seal_missing_and_mismatched_canonical_record (test_core.CoreTests.test_seal_missing_and_mismatched_canonical_record) ... ok
test_seal_missing_records_reservation_not_replaced (test_core.CoreTests.test_seal_missing_records_reservation_not_replaced) ... ok
test_start_binds_uninitialized_submodule_receipt_without_hashing_deferred_rows (test_core.CoreTests.test_start_binds_uninitialized_submodule_receipt_without_hashing_deferred_rows) ... ok
test_uninitialized_submodule_deferred_not_passed (test_core.CoreTests.test_uninitialized_submodule_deferred_not_passed) ... ok
test_unreviewed_amendment_refused_at_seal (test_core.CoreTests.test_unreviewed_amendment_refused_at_seal) ... ok
test_unsafe_existing_repository_refused (test_core.CoreTests.test_unsafe_existing_repository_refused) ... ok
test_wrong_prior_digest_refused (test_core.CoreTests.test_wrong_prior_digest_refused) ... ok
test_context_identity_refuses (test_identifiers.PipelineTests.test_context_identity_refuses) ... ok
test_coverage_and_enumeration_identity_refuse (test_identifiers.PipelineTests.test_coverage_and_enumeration_identity_refuse) ... ok
test_dropped_record_refuses (test_identifiers.PipelineTests.test_dropped_record_refuses) ... ok
test_existing_root_refuses_and_preserves_bytes (test_identifiers.PipelineTests.test_existing_root_refuses_and_preserves_bytes) ... ok
test_frozen_pin_duplicate_and_count_refusals (test_identifiers.PipelineTests.test_frozen_pin_duplicate_and_count_refusals) ... ok
test_frozen_valid_eight_file_fixture (test_identifiers.PipelineTests.test_frozen_valid_eight_file_fixture) ... ok
test_group_conservation_pin_once_and_source_table (test_identifiers.PipelineTests.test_group_conservation_pin_once_and_source_table) ... ok
test_independent_sources_digest_and_canonical_input_refusals (test_identifiers.PipelineTests.test_independent_sources_digest_and_canonical_input_refusals) ... ok
test_listing_mismatch_and_duplicates_refuse (test_identifiers.PipelineTests.test_listing_mismatch_and_duplicates_refuse) ... ok
test_metadata_delta_and_duplicate (test_identifiers.PipelineTests.test_metadata_delta_and_duplicate) ... ok
test_missing_and_ambiguous_source_lookup (test_identifiers.PipelineTests.test_missing_and_ambiguous_source_lookup) ... ok
test_nonparsed_statuses_all_unresolved (test_identifiers.PipelineTests.test_nonparsed_statuses_all_unresolved) ... ok
test_partial_publication_retained (test_identifiers.PipelineTests.test_partial_publication_retained) ... ok
test_planted_emitted_range_loss_refuses (test_identifiers.PipelineTests.test_planted_emitted_range_loss_refuses) ... ok
test_required_header_and_table_mutants_fail_oracle (test_identifiers.PipelineTests.test_required_header_and_table_mutants_fail_oracle) ... ok
test_runtime_noncanonical_signed_and_source_count_refuse (test_identifiers.PipelineTests.test_runtime_noncanonical_signed_and_source_count_refuse) ... ok
test_seal_identity_and_crosslinks_refuse (test_identifiers.PipelineTests.test_seal_identity_and_crosslinks_refuse) ... ok
test_successful_publication_groups_and_pin_once (test_identifiers.PipelineTests.test_successful_publication_groups_and_pin_once) ... ok
test_summary_count_refusals (test_identifiers.PipelineTests.test_summary_count_refusals) ... ok
test_aliases_and_case (test_identifiers.ProjectionTests.test_aliases_and_case) ... ok
test_any_scope_sha_and_repository_scoped_path_pr (test_identifiers.ProjectionTests.test_any_scope_sha_and_repository_scoped_path_pr) ... ok
test_rejection_boundaries (test_identifiers.ProjectionTests.test_rejection_boundaries) ... ok
test_scoped_paths_prs_and_familiarity (test_identifiers.ProjectionTests.test_scoped_paths_prs_and_familiarity) ... ok
test_unattributed_boundaries (test_identifiers.ProjectionTests.test_unattributed_boundaries) ... ok
test_unknown_top_level_and_duplicate_json (test_identifiers.ProjectionTests.test_unknown_top_level_and_duplicate_json) ... ok

----------------------------------------------------------------------
Ran 182 tests in 14.884s

OK
```

### Check 2

```sh
source .tools/activate && GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 go vet ./evaluation/x1-triage-ranking/execution/...
```

Exit code: **0**.

stdout (complete, empty):

stderr (complete, empty):

### Check 3

```sh
source .tools/activate && GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 go test ./evaluation/x1-triage-ranking/execution/...
```

Exit code: **0**.

stdout (complete):

```text
ok  	scout/evaluation/x1-triage-ranking/execution/model_free	(cached)
```

stderr (complete, empty):

## Recomputed SHA-256 manifest

Paths are study-relative unless in the audited-source table. The response's own hash is supplied separately to avoid a self-hash cycle. The updated request is included here after its metadata refresh.

### Binding source, tests, fixtures and current evidence

| File | SHA-256 |
|---|---|
| `execution/binding/acceptance.py` | `61a95ede2c423c50b1ae0df130160608f72632484f41f7705f3e46f6cceb6b9b` |
| `execution/binding/fixtures/common/go.mod` | `ec75b6e87719a439d7ec4959302e25fd588319c7f7cfd1e6e20b39a324aff0b0` |
| `execution/binding/fixtures/common/index.js` | `2fa9c2c5a47269dd22d1da68d24423b1a2d3a2da635c4c60798c59e17ade5d23` |
| `execution/binding/fixtures/common/main.go` | `d37528f6027aa09adca50fa54e40ffc247c5597ffb7d4effa633ccba6134536c` |
| `execution/binding/fixtures/common/package.json` | `514ef8299051029ada3268f1fb2eba8b698548d59bb6d438d71c42c03a627f9e` |
| `execution/binding/fixtures/common/unchanged.py` | `534f740e38f9a8c5e7226a676a427bc5b44ec64ab605c33c595a68aedf5a97c4` |
| `execution/binding/fixtures/head/.scout/policy.yaml` | `6189fea6ad039f0eb0206a7986aa94a44e3bdf46669184298b8a558d06fa7255` |
| `execution/binding/fixtures/head/changed.py` | `0950bd4989dfc76b1c8ed8641c91b4994cefee6bc5bb6d5105510f99f25ec109` |
| `execution/binding/fixtures/parent/.scout/policy.yaml` | `3d3dae7d063c4fe94127de4e67ae9315e1495e4fa510ea587ae4c5a058604ec3` |
| `execution/binding/fixtures/parent/changed.py` | `0ecf4e2fd0b76f8305f0729807a67bc7d2eeabe27ed637a98bda8b9d87738e3b` |
| `execution/binding/fixtures.py` | `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b` |
| `execution/binding/interlock.py` | `19464e8940d0458be3606ed0438ae778fd01b48b73f8c2fee801856b80b2e6a1` |
| `execution/model_free/main.go` | `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` |
| `execution/model_free/main_test.go` | `16e66424f7bfef1a0faec66de9793c19e59018ffe8f78191af75524f9c271f70` |
| `execution/tests/test_binding.py` | `e106e98770284c91532d99a8b07a5a61cb6c3fe4f0a3a4c2b3b2a257fcda73f2` |
| `execution/tests/test_binding_adoption.py` | `7e15dab3cc8b21eeaf2651308e74c0e906b5e1039782c127f89c51410b00a9c5` |
| `execution/tests/test_binding_boundaries.py` | `c5c1d042ee440e39eab66bb3bdb0d0d13b4756d4fe0c1018a16d7e63b7c2b177` |
| `preparation/model-free-binding-amendment.md` | `c8a8e418cdd99ca9c5fc691dd5d5c11766d94b4a888b956ce3fb0601ccf05362` |
| `preparation/model-free-binding-check-v2.json` | `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc` |
| `preparation-reviews/binding-v2-round-1-checks.json` | `8d567ea225c388c19bfb6e82c7e93076f2f5c27e1f26572cf8b2ff3a24ce10e4` |
| `preparation-reviews/binding-v2-REVIEW_REQUEST.md` | `f221a010118cd7b7c31f178396a45cafe4a6d8e6acfe390d041dc6591fed3c5d` |

### Unchanged shared contract / authority and review inputs

| File | SHA-256 |
|---|---|
| `execution/preparation.py` | `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427` |
| `execution/tests/test_core.py` | `94afb7638acbb0316cc019077a87d924fda0e665b752d3d57acdfadbd9657ace` |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `execution/run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `preparation/model-free-binding-check.json` | `5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c` |
| `preparation/model-free-policy.json` | `87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |
| `preparation-reviews/binding-v2-round-1.md` | `91105b4598f4d971ee78933a6578b5560476b966b27260109adc15fd4b0eabcd` |
| `preparation-reviews/identifiers-impl-round-7.md` | `6d279752335d4982754fd5639f729ecc008a23eacd66fa1207268caf1706b01f` |

### Unchanged audited product sources (repository-relative)

| File | SHA-256 |
|---|---|
| `internal/engine/assess.go` | `dbb20a1eccea29f7f46b88337fbdea73a1ee2c4f54e64174f9d03916e59fd1d1` |
| `internal/engine/model_cli_policy.go` | `cb7694f7f44af75d9e71370431de568c1e460aef7a9d8f73efb43c9359ead734` |
| `internal/association/lane.go` | `8f54f8d361f65b74b061df79d96cd99e5cd728e7faf000a2eb68bd5d1946f334` |
| `internal/decisions/client.go` | `0f46173a3d787e1577d97699f64623ebddedb62c3bf31b9428ac5c23538c2fa0` |
| `internal/decisions/run.go` | `bc8d5d38c5281446d6e9ed54c151a6e4fcddb6c9fad3a5b30ac9cf5a113d849c` |
| `internal/engine/ranking.go` | `ca49c3c088df15e5c3e4d2227099e8c1030e7cc909514c7e1dfa196f7cd70380` |
| `internal/clawpatch/client.go` | `a53ca6056b7de521d1dc4c262e593aff00a744ef331440136179e7315b74f180` |
| `internal/clawpatch/review.go` | `1a0bf3825d0228c8372c81c3d79b65edd86b2633b83acf346f2ef558df1ac117` |
| `third_party/clawpatch/src/app.ts` | `c6336d07cf2c763cc677523cc26f98fed90fe3dc926d927aa95a174ad34ad575` |
| `third_party/clawpatch/src/detect.ts` | `9bcd0a495a4f794661c9aff3600d2825ae730d7a80cca24683430fd76e477e46` |
| `third_party/clawpatch/src/git.ts` | `dfc98e3764b1cafa1c7f5a46490009ffef10e0e6efb3ceadcd0431841108efcb` |
| `internal/modelcli/modelcli.go` | `b02b4692a8a6e44db48fc07c3cab4e09b843cb8fb6db8850d7d5838f9e7f9096` |
| `third_party/clawpatch/src/providers/claude.ts` | `fadf1e447d9b0b9a32ba7db6ca9101d29a84acce5c575f37423995456d03d059` |
| `third_party/clawpatch/src/providers/codex.ts` | `994e8066be12c3051c4cb63b32e2144bcf4bd495c5d17a8033f94814bd7a6e92` |
| `internal/sandbox/sandbox.go` | `18886a0b81e57c1e1eb992ae43844750a0538777b1f7e26324a3120a1cc25c73` |
| `internal/sandbox/run.go` | `d78faa34ec691361900d986ff4aeb4cba11f88e7815bbb8cc481393ecba82afd` |
| `internal/analyzers/runner.go` | `8ec374d7b1e67fa04999d5ce8934c5428285ab3b1eeaa4d109342990b6e979f3` |
| `internal/measures/tools.go` | `966e94868d45bae749b8fed74a237cf8a1348b21db6512ca19e74e786e2864f1` |
| `internal/prmeta/fetch.go` | `2cac2d7813b5a602180cc247096824a93527f042216e19de6a03a9a59e9d24cd` |
| `internal/baseline/baseline.go` | `bdc94bb7876d03203110bcff82b4780f204e53abd2a0750ccbdb2e9ffef9d1b7` |
| `internal/inventory/plan.go` | `becf04b94d3b175de6f62fb67fdb63dd3c09f6b526d616978cb1e481ce1a278a` |
| `internal/engine/lanes.go` | `7bcddebbfea020b8e67a1d1b17e2557545f38f03bd6af1619213657925c98ca9` |
