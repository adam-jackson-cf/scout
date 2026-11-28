# Binding v2 round 2 — response

Date: 2026-11-09. **All three required fixes supplied for council re-review; no sign-off or adoption claimed.** Changes are confined to binding source/tests and the amendment/request/response evidence. No Core change was needed. In particular `preparation.py` remains `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427`, and `verify_inputs.py` remains `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`; closure round 4's pinned bytes are unchanged. `test_core.py`, config, proposal, signed v1 policy/recipe, product code and the real 45-day clock were not edited. Files remain uncommitted.

No real preparation adoption, rehearsal, canonical acceptance, seal, fixture scanner, candidate scan, model CLI, Infisical operation or network fetch was run. The required real-code lifecycle tests operate only in temporary synthetic trees with scanner execution and explicitly permitted non-binding prerequisites substituted.

## X1B2-R2-01 — Major — exact protocol authority, repairable implementation

### Fix

`execution/binding/acceptance.py::reviewed_inputs` requires the authority receipt's reviewed set to be **exactly** `{proposal, recipe, amendment}` and its row count to be exactly three (duplicates/extra rows also refuse). Binding sources and the future source manifest no longer appear in adoption authority. Existing adopted-copy SHA and exact `adoption_review` pin checks remain.

Sources remain bound by the post-adoption implementation-reviewed manifest (`source_pins ⊆ manifest`) and by the rehearsal review. The council record still names every reviewed source/test digest; it is not necessary or permissible to place those source files in the immutable adoption receipt.

The amendment and current request state the freeze and repair procedure explicitly:

1. From adoption through seal, keep the proposal, recipe, amendment, three-row receipt and its council record unchanged (the existing Core writer/config/clock pins also remain unchanged).
2. If a rehearsal exposes a source defect, obtain a new zero-issue council round for the fix; regenerate the source manifest and implementation/runtime review; then perform a **fresh** rehearsal and obtain its review.
3. All repair, re-review and fresh rehearsal work must finish **before any canonical reservation**. An old report/source pin cannot authorize the changed implementation.

The existing exact canonical `preparation.py adopt` command is retained in the amendment/request; adoption is still Main-only after zero-issue review, never manual copying/timestamping.

### Catching tests

`tests/test_binding_adoption.py::BindingAdoptionIntegrationTests` now uses an exact three-row real-file receipt and retains real adopt/load/authority/pin validation. It replaces the obsolete “source absent from adoption receipt” test with:

- `test_one_source_pin_missing_from_implementation_manifest_refuses_binding`
- `test_extra_row_in_amendment_review_refuses_binding`
- `test_recipe_missing_from_amendment_review_refuses_binding`
- `test_amendment_missing_from_amendment_review_refuses_binding`
- `test_source_fix_after_real_adoption_accepts_fresh_implementation_review`

The last test changes a temporary binding source after a real synthetic adoption, regenerates its manifest/review, and proves real `load_config` and `reviewed_inputs` accept the new pins without changing adoption authority. The full lifecycle test below additionally performs this change **after an initial rehearsal**, refuses stale source/report authority, then reaches a real synthetic seal after fresh reviews/rehearsal.

## X1B2-R2-02 — Minor — seal-time rehearsal authority and accounting

### Fix

`verify_rehearsal` now returns all three absolute pins:

```text
{report, review, record: pin(review['record']['path'])}
```

Canonical synthetic evidence stores the result as `reviewed_rehearsal`. Existing Core `verify_tree_pins` therefore rechecks the actual council record at seal, and the normal source/output reference accounting includes it even if it is beneath OUT. No new Core reader, receipt type or required-review entry was introduced. The existing small rehearsal fixture now supplies its record pin too.

### Catching test

`tests/test_binding_seal.py::BindingSealIntegrationTests.test_rehearsal_record_mutation_invalidates_canonical_tree_and_seal` uses a real rehearsal receipt and council file, real canonical synthetic/Django evidence and real `collect_seal`. It first requires a successful seal with the rehearsal council record deliberately **under OUT** and asserts that record appears in the sealed reference set. It then both **changes** and **deletes** the council file, leaving the review JSON unchanged, and requires:

- `verify_tree_pins` on the already-created canonical evidence to refuse;
- real `collect_seal` to refuse.

These checks are not merely re-entry into `verify_rehearsal`, and they do not mock `review_receipt` or the recursive pin checker.

## X1B2-R2-03 — Minor — real positive v2 lifecycle through seal

### Fix and tests

New `tests/test_binding_seal.py` composes only reusable fixture operations from the adoption tests (not inherited test methods). It exercises:

**real adopt → real rehearse → real verify_rehearsal → canonical synthetic main → Django main → real collect_seal**.

Real code is retained for adoption authority/config loading, review validation, pins, clock checks, source-manifest closure, materialization verification, effective-record byte comparison, policy/proof predicates, parity, action journals, v2 binding checks, required-reviewed roles, and final output accounting. The synthetic proposal really adopts all four changes, including v2 acceptance keys; the base config is not pre-populated with v2 keys.

Only the explicitly permitted seams are substituted:
- scanner execution/materialization inputs: `run_revision`, `materialize`, `generate`, `git`;
- non-binding seal prerequisites: `documentary_gate`, installed-closure reinspection and object-store/Git inspection.

All receipts, council records, manifests, canonical outputs, clock and journal files are real temporary files. Required a/b/c/d journal pairs and the complete temporary EXEC manifest are processed by unchanged Core logic. The unit clock is fixed once before adoption and remains byte-identical through seal.

Passing tests:

1. `test_real_adopt_rehearse_canonical_django_and_v2_seal` — accepts both `scout.x1.binding-acceptance.v2` records; checks exact six keys, runtime closure, Django revisions/synthetic dependency and under-OUT rehearsal-record accounting. Removing **either** canonical binding digest from the real final review then makes `collect_seal` refuse with `scoped reviews do not bind all required evidence`.
2. `test_post_rehearsal_source_fix_requires_fresh_review_before_reservation` — runs an initial rehearsal, changes a temporary binding source, confirms load still succeeds but stale implementation pins refuse, creates a new real synthetic council record and fresh manifest/runtime/implementation authority, refuses the old rehearsal, confirms no canonical synthetic path exists yet, then performs a new rehearsal/review and reaches canonical synthetic, Django and seal. Original adoption receipt/record bytes stay unchanged.
3. `test_rehearsal_record_mutation_invalidates_canonical_tree_and_seal` — positive seal and mutation/deletion refusals described above.

This closes the missing end-to-end **synthetic authority** proof. Actual scanner trigger counts/fixture findings remain unmeasured; the test does not purport to be Main's real rehearsal or acceptance.

## Verification and authority status

Fresh full Python suite: **205 tests passed in 11.842s**, exit0, including the concurrent closure-round-3 tests and every new binding lifecycle case. Fresh Go vet: exit0, no output. Fresh Go test: exit0; the cached package result is recorded as such. No failed check was rerun merely to confirm it. Full outputs follow and are duplicated in [binding-v2-round-2-checks.json](binding-v2-round-2-checks.json), SHA-256 `07dc03e1c55c4a93b246d6ef60f7c5d3879b1b7321437f0df61a38d446d92c55`.

The current request replaces its stale 182-test summary with this fresh result. Prior review/check records are retained unchanged as historical evidence. No new timeout/sandbox policy or Core change was added. The amendment remains proposed and unadopted pending council round 3 with zero issues. Main's real action sequence and the existing R-064/R-087 provenance/45-day clock are unchanged.

## Commands and complete captured output

Cwd for all commands: `/Users/adamjackson/Projects/work/scout-x1prep`. Go commands activate the pinned toolchain and disable network/module mutation. Empty streams are explicitly marked; nonempty streams below are verbatim.

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
test_amendment_missing_from_amendment_review_refuses_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_amendment_missing_from_amendment_review_refuses_binding) ... ok
test_extra_row_in_amendment_review_refuses_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_extra_row_in_amendment_review_refuses_binding) ... ok
test_loaded_configuration_requires_exact_adopted_row (test_binding_adoption.BindingAdoptionIntegrationTests.test_loaded_configuration_requires_exact_adopted_row) ... ok
test_one_source_pin_missing_from_implementation_manifest_refuses_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_one_source_pin_missing_from_implementation_manifest_refuses_binding) ... ok
test_proposal_digest_missing_refuses_real_adoption (test_binding_adoption.BindingAdoptionIntegrationTests.test_proposal_digest_missing_refuses_real_adoption) ... ok
test_proposal_path_missing_refuses_binding_even_with_reviewed_digest (test_binding_adoption.BindingAdoptionIntegrationTests.test_proposal_path_missing_refuses_binding_even_with_reviewed_digest) ... ok
test_real_adopt_load_and_entire_binding_authority_chain (test_binding_adoption.BindingAdoptionIntegrationTests.test_real_adopt_load_and_entire_binding_authority_chain) ... ok
test_recipe_missing_from_amendment_review_refuses_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_recipe_missing_from_amendment_review_refuses_binding) ... ok
test_source_fix_after_real_adoption_accepts_fresh_implementation_review (test_binding_adoption.BindingAdoptionIntegrationTests.test_source_fix_after_real_adoption_accepts_fresh_implementation_review) ... ok
test_every_launch_checks_key_presence_before_direct_execution (test_binding_boundaries.ControllerKeyBoundaryTests.test_every_launch_checks_key_presence_before_direct_execution) ... ok
test_rehearsal_entry_checks_presence_before_output_reservation (test_binding_boundaries.ControllerKeyBoundaryTests.test_rehearsal_entry_checks_presence_before_output_reservation) ... ok
test_reviewed_inputs_checks_presence_not_truthiness (test_binding_boundaries.ControllerKeyBoundaryTests.test_reviewed_inputs_checks_presence_not_truthiness) ... ok
test_django_evidence_pins_exact_synthetic_receipt_and_detects_change (test_binding_boundaries.DjangoFixtureBoundaryTests.test_django_evidence_pins_exact_synthetic_receipt_and_detects_change) ... ok
test_django_manifest_refusal_table (test_binding_boundaries.DjangoFixtureBoundaryTests.test_django_manifest_refusal_table) ... ok
test_equal_synthetic_parent_head_issue_sets_refuse (test_binding_boundaries.DjangoFixtureBoundaryTests.test_equal_synthetic_parent_head_issue_sets_refuse) ... ok
test_each_bound_reference_attempt_count_must_be_zero (test_binding_boundaries.PredicateBoundaryTests.test_each_bound_reference_attempt_count_must_be_zero) ... ok
test_positive_control_false_boundaries (test_binding_boundaries.PredicateBoundaryTests.test_positive_control_false_boundaries) ... ok
test_resolve_attempt_and_exit_status_boundaries (test_binding_boundaries.PredicateBoundaryTests.test_resolve_attempt_and_exit_status_boundaries) ... ok
test_post_rehearsal_source_fix_requires_fresh_review_before_reservation (test_binding_seal.BindingSealIntegrationTests.test_post_rehearsal_source_fix_requires_fresh_review_before_reservation) ... ok
test_real_adopt_rehearse_canonical_django_and_v2_seal (test_binding_seal.BindingSealIntegrationTests.test_real_adopt_rehearse_canonical_django_and_v2_seal) ... ok
test_rehearsal_record_mutation_invalidates_canonical_tree_and_seal (test_binding_seal.BindingSealIntegrationTests.test_rehearsal_record_mutation_invalidates_canonical_tree_and_seal) ... ok
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
test_r301_45_slice_native_collision_never_skipped (test_closure_round3.RoundThreeTests.test_r301_45_slice_native_collision_never_skipped) ... ok
test_r301_bounded_fat_refuses_even_if_stricter_checks_would_fail (test_closure_round3.RoundThreeTests.test_r301_bounded_fat_refuses_even_if_stricter_checks_would_fail) ... ok
test_r301_java_52_0_and_45_3_remain_non_native (test_closure_round3.RoundThreeTests.test_r301_java_52_0_and_45_3_remain_non_native) ... ok
test_r301_java_only_skipped_when_a_fat_slice_escapes_file (test_closure_round3.RoundThreeTests.test_r301_java_only_skipped_when_a_fat_slice_escapes_file) ... ok
test_s4_alignment_exponent_and_table_overlap (test_closure_round3.RoundThreeTests.test_s4_alignment_exponent_and_table_overlap) ... ok
test_s4_cpu_disagreement (test_closure_round3.RoundThreeTests.test_s4_cpu_disagreement) ... ok
test_s4_empty_slice (test_closure_round3.RoundThreeTests.test_s4_empty_slice) ... ok
test_s4_fat64_reserved_field (test_closure_round3.RoundThreeTests.test_s4_fat64_reserved_field) ... ok
test_s4_invalid_architecture_counts (test_closure_round3.RoundThreeTests.test_s4_invalid_architecture_counts) ... ok
test_s4_invalid_slice_magic (test_closure_round3.RoundThreeTests.test_s4_invalid_slice_magic) ... ok
test_s4_invalid_thin_filetype_and_command_bounds (test_closure_round3.RoundThreeTests.test_s4_invalid_thin_filetype_and_command_bounds) ... ok
test_s4_missing_or_duplicate_arm64 (test_closure_round3.RoundThreeTests.test_s4_missing_or_duplicate_arm64) ... ok
test_s4_non_native_null (test_closure_round3.RoundThreeTests.test_s4_non_native_null) ... ok
test_s4_overlapping_slices (test_closure_round3.RoundThreeTests.test_s4_overlapping_slices) ... ok
test_s4_truncated_fat_version (test_closure_round3.RoundThreeTests.test_s4_truncated_fat_version) ... ok
test_s4_x86_and_arm64e_null_and_swapped_fat (test_closure_round3.RoundThreeTests.test_s4_x86_and_arm64e_null_and_swapped_fat) ... ok
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
Ran 205 tests in 11.842s

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

## Recomputed SHA-256

Study-relative paths. These hashes include every binding source, fixture, test and current evidence file plus the unchanged Core/closure sign-off inputs. The response's own hash is delivered separately to avoid self-reference.

### Binding source, fixtures, tests and evidence

| File | SHA-256 |
|---|---|
| `execution/binding/acceptance.py` | `2144999a8aece522e72a86a429a1628a2833ada1490fa6e0fe7600e152373be1` |
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
| `execution/tests/test_binding.py` | `2458e4289d67a1a061fcf43a2de6e42f1d586daee6cb4452f17f857ee1d36ab5` |
| `execution/tests/test_binding_adoption.py` | `c08775b684218d660fbf761f5c364331bce2a120a8eb10d164de5a794dd98003` |
| `execution/tests/test_binding_boundaries.py` | `c5c1d042ee440e39eab66bb3bdb0d0d13b4756d4fe0c1018a16d7e63b7c2b177` |
| `execution/tests/test_binding_seal.py` | `c867fa0e60bd2c2491e95ed90822d0746801747d6a4480a95a21bd7cba3bfb7b` |
| `preparation/model-free-binding-amendment.md` | `634b352cbf893042a1e4b525c59492d5ae00f3bc4c8ce34e999065efb7054edf` |
| `preparation/model-free-binding-check-v2.json` | `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc` |
| `preparation-reviews/binding-v2-round-2-checks.json` | `07dc03e1c55c4a93b246d6ef60f7c5d3879b1b7321437f0df61a38d446d92c55` |
| `preparation-reviews/binding-v2-REVIEW_REQUEST.md` | `f633f53f26e59e7c7bdb3e020f50d2d9a1384536cacfe8dcc48a317a6ac67575` |

### Shared contract, immutable basis and council inputs — unchanged by Binding

| File | SHA-256 |
|---|---|
| `execution/preparation.py` | `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427` |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `execution/tests/test_core.py` | `94afb7638acbb0316cc019077a87d924fda0e665b752d3d57acdfadbd9657ace` |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `execution/run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` |
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` |
| `preparation/model-free-binding-check.json` | `5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c` |
| `preparation/model-free-policy.json` | `87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |
| `preparation-reviews/binding-v2-round-2.md` | `29c1651f01bc5c87543f182de84c2008307e4e17dcb69ec9b3f60e4890327dc0` |
| `preparation-reviews/closure-impl-round-4.md` | `5e9fcd1bcd1a28ae3487505df11d1bb95c378d6c2d1a33a059b85a9ffd349627` |
