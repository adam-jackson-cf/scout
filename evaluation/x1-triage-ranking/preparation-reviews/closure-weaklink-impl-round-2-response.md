# Closure weak-link implementation round 2 response

Status: all three requested fixes implemented and exercised; submitted for council review, not signed off. No GR, real journal, preflight, canonical closure, O0, P3-D observation, network fetch, model CLI, secret use, settings change, or privilege escalation was performed.

## Authority and reviewed files

Authority: `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-weaklink-impl-round-1.md`, SHA-256 `b50e016181651e8b98c33973c3a2869c563403025f91cd2785707c0940c7bc99`. This response follows the authority's response filename. All prior review records and measurement artifacts remain unchanged. Only the previously unsigned collector and its test file were edited; the wrapper remains unchanged. Nothing is committed.

### Three execution files

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native_v2.py` | 15849 | `5e320ed37a5b106e3a37984e6137e78f87422aa525792018032117d66f4dee09` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/attest_closure_v2.py` | 1168 | `c28a9853f826d46e05ddebbd207ac85cc112aec90179a84b7a0800aeecfd6d84` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests/test_closure_weak_links.py` | 32601 | `cf0385d711d97fae5dc7656c4368bf2badb66cb2a07b0bab44e7f27cdebde417` |

### Five unchanged signed dependencies

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/attest_closure.py` | 46996 | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native.py` | 10692 | `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/preparation.py` | 50257 | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/verify_inputs.py` | 11211 | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/run-config.json` | 5948 | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |

The full-suite record pins all eight files before and after execution. The before and after rows are identical. Both signed-base guards remain unchanged.

## X1CW-R1-01: loaded-image identity

Added patchable module-level `loaded_cache_image_uuid(name)`. It uses ctypes on `CDLL(None)` and requires exactly one matching loaded image name, a complete 64-bit Mach-O header within the active cache address range, `MH_MAGIC_64`, bounded `ncmds`/`sizeofcmds` traversal, valid command extents, and exactly one `LC_UUID`. Any unsuccessful outcome raises the required `cannot establish active-cache identity` ValueError.

Only the present-file/cache-member I7 branch invokes the helper after strict CLI parsing. It compares the loaded UUID against the disk-slice UUID case-insensitively, refuses a mismatch with the required message, and adds exactly the requested `identity_check` fields. Other cached records have no identity-check field and retain the unqualified individual query without `-arch`.

The identity is now described honestly: the CLI supplies the disk arm64e slice UUID/link list; the in-memory UUID comparison ties that identity to the image actually loaded inside the active cache. The earlier claim that `-arch` itself reads the active-cache representation is superseded.

Synthetic tests pass for matching identity/field emission, mismatched identity, image not loaded, duplicate loaded name, header outside cache, invalid magic, malformed/out-of-range load commands, missing/duplicate UUID, and the no-`-arch` absent-disk member branch. R1 now asserts the identity checks on all three real system alternates.

### Required rule-3 measurement

Fresh pinned Python process; closed environment; each metadata command and its complete outputs/timestamps retained. All three in-memory `LC_UUID` values equal their measured `dyld_info -arch arm64e` UUIDs:

| Image | Loaded LC_UUID | Disk arm64e UUID | Equal |
|---|---|---|---|
| `/usr/lib/system/libsystem_kernel.dylib` | `7CFB3A10-ADB2-32C4-99B5-922A111B3224` | `7CFB3A10-ADB2-32C4-99B5-922A111B3224` | True |
| `/usr/lib/system/libsystem_platform.dylib` | `4C6BD0FE-25F2-38D8-BE5C-24DAC1DEDCFB` | `4C6BD0FE-25F2-38D8-BE5C-24DAC1DEDCFB` | True |
| `/usr/lib/system/libsystem_pthread.dylib` | `12342372-0084-37A5-9EA6-2E4DC940C685` | `12342372-0084-37A5-9EA6-2E4DC940C685` | True |

Measurement record:

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-impl2/identity-measurement.json` | 4204 | `fc26f3de0ba5d5f5f138c161970c8e97df0e51eed18c635027c5bfd839cc7f2a` |

## X1CW-R1-02: R3 replaced by R3′a–c

The required real seed was measured before implementation continued. Signed v1 completed successfully; no stop condition or seed substitution occurred.

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/.local/share/mise/installs/go/1.26.10/src/runtime/race/race_darwin_arm64.syso` | 502032 | `6ca6a32e8b650ac03b6a11b40c7a0cdd938d00046bd1c942f7c530e0fdaf23b2` |
| `/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-impl2/seed-v1-measurement.json` | 9079 | `1ac03a9761f91ab7bc1c60faf40570d8a13ec5eea299cef15664cded41f57ccb` |

The seed SHA-256 is pinned literally in the test. Each real test uses its own recording `Commands`, the same closed environment, and the required cache root. All real-byte command records are retained permanently under `/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-impl2`, in separate test/run/collector directories rather than cleanup-managed temporary directories.

- R3′a asserts the exact signed-v1 `/bin/sh` refusal for RemoteProcessingBlock.
- R3′b completes both real null walks on the pinned syso, compares normalised output, asserts exactly one image/no dependencies/no cached or absent images, equal cache pins/UUID, and identical ordered argv lists.
- R3′c captures the signed-v1 `cached` dictionary by read-only exception traceback introspection, completes v2 on `/bin/sh`, checks every prefix image UUID and dependency list (only the specified iOSSupport name allowance), checks multiset inclusion of `(argv, exit_code, stdout)`, checks the matching RemoteProcessingBlock absence probe and every absent-row proof, and asserts all three I7 identity checks.

No signed code was patched or pruned. The former red R3 test was replaced only under this explicit review authority.

## X1CW-R1-03: present non-member stays on disk

Added a cached-parent synthetic case with a present SYSTEM_PREFIXES target and membership false. It asserts that the target is queued and classified by `image_filetype`, its real fixture pin appears in `images`, it does not become a cached record, and no dyld_info call is made for that target. The test passes.

## Full suite

Exact command, cwd `/Users/adamjackson/Projects/work/scout-x1prep`:

```sh
/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python -I -B -m unittest discover -s /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests -p 'test_*.py' -v
```

**332 tests in 71.030s; OK; exit 0; zero errors, failures, or skips.** All five gated real-byte tests ran and passed:

```text
test_r1_coreaudio_weak_absence_and_present_weak_targets (test_closure_weak_links.RealByteTests.test_r1_coreaudio_weak_absence_and_present_weak_targets) ... ok
test_r2_required_remote_processing_block_refuses (test_closure_weak_links.RealByteTests.test_r2_required_remote_processing_block_refuses) ... ok
test_r3a_signed_v1_bin_sh_refusal_remains_visible (test_closure_weak_links.RealByteTests.test_r3a_signed_v1_bin_sh_refusal_remains_visible) ... ok
test_r3b_pinned_real_object_null_matches_v1 (test_closure_weak_links.RealByteTests.test_r3b_pinned_real_object_null_matches_v1) ... ok
test_r3c_real_cached_prefix_matches_signed_v1 (test_closure_weak_links.RealByteTests.test_r3c_real_cached_prefix_matches_signed_v1) ... ok
```

The suite started at `2026-11-10T17:16:37.026702+00:00` and ended at `2026-11-10T17:17:48.300932+00:00`. Its complete output and start/end pins are retained here:

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-impl2/full-suite-round2.json` | 50169 | `95bd2c0ac7562e41186bd54944032ce247915f9188a09bfa4debc0d11c0c659b` |

### Complete stdout

Empty. SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.

### Complete stderr

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
test_duplicate_fourth_amendment_row_refuses_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_duplicate_fourth_amendment_row_refuses_binding) ... ok
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
test_actual_lane_reasons_and_structural_absence (test_binding_contract.ModelFreeContractTests.test_actual_lane_reasons_and_structural_absence) ... ok
test_control_cannot_claim_bound_model_proof (test_binding_contract.ModelFreeContractTests.test_control_cannot_claim_bound_model_proof) ... ok
test_exact_record_routes (test_binding_contract.ModelFreeContractTests.test_exact_record_routes) ... ok
test_flipped_association_nil_refuses_bound_policy (test_binding_contract.ModelFreeContractTests.test_flipped_association_nil_refuses_bound_policy) ... ok
test_missing_model_proof_key_refuses_policies_and_receipts (test_binding_contract.ModelFreeContractTests.test_missing_model_proof_key_refuses_policies_and_receipts) ... ok
test_real_effective_and_receipt_proofs_pass (test_binding_contract.ModelFreeContractTests.test_real_effective_and_receipt_proofs_pass) ... ok
test_post_rehearsal_source_fix_requires_fresh_review_before_reservation (test_binding_seal.BindingSealIntegrationTests.test_post_rehearsal_source_fix_requires_fresh_review_before_reservation) ... ok
test_real_adopt_rehearse_canonical_django_and_v2_seal (test_binding_seal.BindingSealIntegrationTests.test_real_adopt_rehearse_canonical_django_and_v2_seal) ... ok
test_rehearsal_record_mutation_invalidates_canonical_tree_and_seal (test_binding_seal.BindingSealIntegrationTests.test_rehearsal_record_mutation_invalidates_canonical_tree_and_seal) ... ok
test_rehearsal_review_omitting_current_runtime_or_source_refuses (test_binding_seal.BindingSealIntegrationTests.test_rehearsal_review_omitting_current_runtime_or_source_refuses) ... ok
test_changed_imported_dependency_refuses_review_gate (test_cc2_delta.DeltaPublicationTests.test_changed_imported_dependency_refuses_review_gate) ... ok
test_journal_pins_partial_familiarity_from_current_run (test_cc2_delta.DeltaPublicationTests.test_journal_pins_partial_familiarity_from_current_run) ... ok
test_partial_publication_retained (test_cc2_delta.DeltaPublicationTests.test_partial_publication_retained) ... ok
test_publish_is_create_once_union_and_preserves_original (test_cc2_delta.DeltaPublicationTests.test_publish_is_create_once_union_and_preserves_original) ... ok
test_request_names_exact_baseline_journal_pin (test_cc2_delta.DeltaPublicationTests.test_request_names_exact_baseline_journal_pin) ... ok
test_request_requires_final_shared_code_freeze_order (test_cc2_delta.DeltaPublicationTests.test_request_requires_final_shared_code_freeze_order) ... ok
test_review_gate_requires_each_hash (test_cc2_delta.DeltaPublicationTests.test_review_gate_requires_each_hash) ... ok
test_source_pin_mismatch_and_symlink_refuse (test_cc2_delta.DeltaPublicationTests.test_source_pin_mismatch_and_symlink_refuse) ... ok
test_synthetic_journalled_run_and_refusal_keep_baseline (test_cc2_delta.DeltaPublicationTests.test_synthetic_journalled_run_and_refusal_keep_baseline) ... ok
test_all_schema_families (test_cc2_delta.ReaderTests.test_all_schema_families) ... ok
test_csv_never_decodes_excluded_cells (test_cc2_delta.ReaderTests.test_csv_never_decodes_excluded_cells) ... ok
test_csv_quoted_newline_skipped_without_decode (test_cc2_delta.ReaderTests.test_csv_quoted_newline_skipped_without_decode) ... ok
test_csv_unknown_header_duplicate_and_column_counts_refuse (test_cc2_delta.ReaderTests.test_csv_unknown_header_duplicate_and_column_counts_refuse) ... ok
test_excluded_fields_never_decoded_or_emitted (test_cc2_delta.ReaderTests.test_excluded_fields_never_decoded_or_emitted) ... ok
test_json_bad_framing_refuses (test_cc2_delta.ReaderTests.test_json_bad_framing_refuses) ... ok
test_missing_scope_unknown_scope_bad_pr_sha_path_refuse (test_cc2_delta.ReaderTests.test_missing_scope_unknown_scope_bad_pr_sha_path_refuse) ... ok
test_null_parent_and_snapshot_do_not_invent_identifiers (test_cc2_delta.ReaderTests.test_null_parent_and_snapshot_do_not_invent_identifiers) ... ok
test_oid_and_parent_are_commits_never_blobs (test_cc2_delta.ReaderTests.test_oid_and_parent_are_commits_never_blobs) ... ok
test_other_repository_rows_not_used (test_cc2_delta.ReaderTests.test_other_repository_rows_not_used) ... ok
test_parent_follow_prs_not_commit_sha (test_cc2_delta.ReaderTests.test_parent_follow_prs_not_commit_sha) ... ok
test_snapshot_trail_measured_and_all_partner_paths (test_cc2_delta.ReaderTests.test_snapshot_trail_measured_and_all_partner_paths) ... ok
test_unknown_forbidden_key_refuses_before_its_value (test_cc2_delta.ReaderTests.test_unknown_forbidden_key_refuses_before_its_value) ... ok
test_unknown_nested_structures_and_duplicate_keys_refuse (test_cc2_delta.ReaderTests.test_unknown_nested_structures_and_duplicate_keys_refuse) ... ok
test_baseline_outputs_cannot_be_substituted (test_cc2_seal.DeltaSealTests.test_baseline_outputs_cannot_be_substituted) ... ok
test_both_review_receipts_are_required_and_scope_specific (test_cc2_seal.DeltaSealTests.test_both_review_receipts_are_required_and_scope_specific) ... ok
test_delta_schema_clock_chain_and_union_semantics_are_required (test_cc2_seal.DeltaSealTests.test_delta_schema_clock_chain_and_union_semantics_are_required) ... ok
test_familiarity_header_must_match_manifest_sources (test_cc2_seal.DeltaSealTests.test_familiarity_header_must_match_manifest_sources) ... ok
test_full_seal_refuses_changed_execution_source (test_cc2_seal.DeltaSealTests.test_full_seal_refuses_changed_execution_source) ... ok
test_full_seal_refuses_missing_delta_output (test_cc2_seal.DeltaSealTests.test_full_seal_refuses_missing_delta_output) ... ok
test_full_seal_retains_baseline_and_canonical_delta_roles_and_scopes (test_cc2_seal.DeltaSealTests.test_full_seal_retains_baseline_and_canonical_delta_roles_and_scopes) ... ok
test_implementation_review_must_bind_every_control (test_cc2_seal.DeltaSealTests.test_implementation_review_must_bind_every_control) ... ok
test_output_review_must_bind_all_outputs_and_the_passed_journal (test_cc2_seal.DeltaSealTests.test_output_review_must_bind_all_outputs_and_the_passed_journal) ... ok
test_unique_passed_action_d_and_producer_control_pins_are_required (test_cc2_seal.DeltaSealTests.test_unique_passed_action_d_and_producer_control_pins_are_required) ... ok
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
test_duplicate_loaded_name_refuses (test_closure_weak_links.LoadedCacheIdentityTests.test_duplicate_loaded_name_refuses) ... ok
test_header_outside_cache_refuses (test_closure_weak_links.LoadedCacheIdentityTests.test_header_outside_cache_refuses) ... ok
test_image_not_loaded_refuses (test_closure_weak_links.LoadedCacheIdentityTests.test_image_not_loaded_refuses) ... ok
test_invalid_header_and_command_bounds_refuse (test_closure_weak_links.LoadedCacheIdentityTests.test_invalid_header_and_command_bounds_refuse) ... ok
test_loaded_uuid_matches_memory (test_closure_weak_links.LoadedCacheIdentityTests.test_loaded_uuid_matches_memory) ... ok
test_r1_coreaudio_weak_absence_and_present_weak_targets (test_closure_weak_links.RealByteTests.test_r1_coreaudio_weak_absence_and_present_weak_targets) ... ok
test_r2_required_remote_processing_block_refuses (test_closure_weak_links.RealByteTests.test_r2_required_remote_processing_block_refuses) ... ok
test_r3a_signed_v1_bin_sh_refusal_remains_visible (test_closure_weak_links.RealByteTests.test_r3a_signed_v1_bin_sh_refusal_remains_visible) ... ok
test_r3b_pinned_real_object_null_matches_v1 (test_closure_weak_links.RealByteTests.test_r3b_pinned_real_object_null_matches_v1) ... ok
test_r3c_real_cached_prefix_matches_signed_v1 (test_closure_weak_links.RealByteTests.test_r3c_real_cached_prefix_matches_signed_v1) ... ok
test_filesystem_permission_failure_refuses (test_closure_weak_links.WeakLinkTests.test_filesystem_permission_failure_refuses) ... ok
test_i7_absent_disk_cache_member_is_queried_without_architecture (test_closure_weak_links.WeakLinkTests.test_i7_absent_disk_cache_member_is_queried_without_architecture) ... ok
test_i7_loaded_uuid_mismatch_refuses (test_closure_weak_links.WeakLinkTests.test_i7_loaded_uuid_mismatch_refuses) ... ok
test_i7_present_nonmember_stays_on_disk (test_closure_weak_links.WeakLinkTests.test_i7_present_nonmember_stays_on_disk) ... ok
test_i7_present_system_cache_member_uses_active_cache_not_disk_alternate (test_closure_weak_links.WeakLinkTests.test_i7_present_system_cache_member_uses_active_cache_not_disk_alternate) ... ok
test_i7_selected_architecture_identity_must_be_unique_and_present (test_closure_weak_links.WeakLinkTests.test_i7_selected_architecture_identity_must_be_unique_and_present) ... ok
test_i7_unidentifiable_cache_architecture_refuses_without_guessing (test_closure_weak_links.WeakLinkTests.test_i7_unidentifiable_cache_architecture_refuses_without_guessing) ... ok
test_membership_symbol_missing_refuses (test_closure_weak_links.WeakLinkTests.test_membership_symbol_missing_refuses) ... ok
test_overlay_permission_failure_refuses (test_closure_weak_links.WeakLinkTests.test_overlay_permission_failure_refuses) ... ok
test_s10_disk_required_and_mixed_commands_refuse (test_closure_weak_links.WeakLinkTests.test_s10_disk_required_and_mixed_commands_refuse) ... ok
test_s10_disk_weak_absence (test_closure_weak_links.WeakLinkTests.test_s10_disk_weak_absence) ... ok
test_s10_unresolved_rpath_or_non_system_weak_refuses (test_closure_weak_links.WeakLinkTests.test_s10_unresolved_rpath_or_non_system_weak_refuses) ... ok
test_s11_cached_chain_of_5000_is_iterative (test_closure_weak_links.WeakLinkTests.test_s11_cached_chain_of_5000_is_iterative) ... ok
test_s12_cached_null_matches_v1_and_visit_order (test_closure_weak_links.WeakLinkTests.test_s12_cached_null_matches_v1_and_visit_order) ... ok
test_s12_executable_context_null_matches_v1 (test_closure_weak_links.WeakLinkTests.test_s12_executable_context_null_matches_v1) ... ok
test_s13_signed_consumers_accept_absent_rows (test_closure_weak_links.WeakLinkTests.test_s13_signed_consumers_accept_absent_rows) ... ok
test_s14_absence_requires_a_cache_family (test_closure_weak_links.WeakLinkTests.test_s14_absence_requires_a_cache_family) ... ok
test_s1_absent_weak_evidence_and_probe_memoization (test_closure_weak_links.WeakLinkTests.test_s1_absent_weak_evidence_and_probe_memoization) ... ok
test_s2_required_absence_refuses (test_closure_weak_links.WeakLinkTests.test_s2_required_absence_refuses) ... ok
test_s3_delay_init_is_not_weak (test_closure_weak_links.WeakLinkTests.test_s3_delay_init_is_not_weak) ... ok
test_s4_reexport_weak_is_required (test_closure_weak_links.WeakLinkTests.test_s4_reexport_weak_is_required) ... ok
test_s5_present_cache_image_parse_errors_never_become_absent (test_closure_weak_links.WeakLinkTests.test_s5_present_cache_image_parse_errors_never_become_absent) ... ok
test_s6_disk_or_overlay_entry_prevents_absence (test_closure_weak_links.WeakLinkTests.test_s6_disk_or_overlay_entry_prevents_absence) ... ok
test_s7_nonempty_probe_stdout_prevents_absence (test_closure_weak_links.WeakLinkTests.test_s7_nonempty_probe_stdout_prevents_absence) ... ok
test_s8_duplicate_target_strength_is_required (test_closure_weak_links.WeakLinkTests.test_s8_duplicate_target_strength_is_required) ... ok
test_s8_required_edge_refuses_in_both_visit_orders (test_closure_weak_links.WeakLinkTests.test_s8_required_edge_refuses_in_both_visit_orders) ... ok
test_s9_parser_refuses_unknown_repeated_and_malformed_grammar (test_closure_weak_links.WeakLinkTests.test_s9_parser_refuses_unknown_repeated_and_malformed_grammar) ... ok
test_s9_rpaths_end_links_and_iossupport_path_stays_whole (test_closure_weak_links.WeakLinkTests.test_s9_rpaths_end_links_and_iossupport_path_stays_whole) ... ok
test_w1_drift_refuses_before_rebinding (test_closure_weak_links.WrapperTests.test_w1_drift_refuses_before_rebinding) ... ok
test_w2_exact_rebinding_and_real_inventory_route (test_closure_weak_links.WrapperTests.test_w2_exact_rebinding_and_real_inventory_route) ... ok
test_w3_exact_source_order (test_closure_weak_links.WrapperTests.test_w3_exact_source_order) ... ok
test_w4_main_binds_v2_journal_sources_and_argv (test_closure_weak_links.WrapperTests.test_w4_main_binds_v2_journal_sources_and_argv) ... ok
test_w5_import_has_no_mutation_and_signed_source_assertion_passes (test_closure_weak_links.WrapperTests.test_w5_import_has_no_mutation_and_signed_source_assertion_passes) ... ok
test_adopted_amendment_without_review_refused_at_load (test_core.CoreTests.test_adopted_amendment_without_review_refused_at_load) ... ok
test_adoption_racing_journal_is_serialized (test_core.CoreTests.test_adoption_racing_journal_is_serialized) ... ok
test_amendment_chain_applies_in_order_and_preserves_original (test_core.CoreTests.test_amendment_chain_applies_in_order_and_preserves_original) ... ok
test_amendment_review_scope_and_exact_digest_required (test_core.CoreTests.test_amendment_review_scope_and_exact_digest_required) ... ok
test_canonical_output_role_preserved_when_journal_also_pins_it (test_core.CoreTests.test_canonical_output_role_preserved_when_journal_also_pins_it) ... ok
test_clock_pinned_original_config_cannot_drift (test_core.CoreTests.test_clock_pinned_original_config_cannot_drift) ... ok
test_deadline_exact_seconds_and_leap_boundary (test_core.CoreTests.test_deadline_exact_seconds_and_leap_boundary) ... ok
test_duplicate_journal_outputs_are_deduplicated_but_all_verified (test_core.CoreTests.test_duplicate_journal_outputs_are_deduplicated_but_all_verified) ... ok
test_edit_delete_or_rename_adopted_amendment_refused (test_core.CoreTests.test_edit_delete_or_rename_adopted_amendment_refused) ... ok
test_exception_classification (test_core.CoreTests.test_exception_classification) ... ok
test_exclusive_creation_preserves_original (test_core.CoreTests.test_exclusive_creation_preserves_original) ... ok
test_hand_written_and_backdated_adoption_receipts_refused (test_core.CoreTests.test_hand_written_and_backdated_adoption_receipts_refused) ... ok
test_hd9_only_state_date_normalized (test_core.CoreTests.test_hd9_only_state_date_normalized) ... ok
test_initialized_submodule_checks_working_bytes (test_core.CoreTests.test_initialized_submodule_checks_working_bytes) ... ok
test_initialized_submodule_mismatch_refused (test_core.CoreTests.test_initialized_submodule_mismatch_refused) ... ok
test_journal_pinned_output_is_accounted_and_sealed (test_core.CoreTests.test_journal_pinned_output_is_accounted_and_sealed) ... ok
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
test_tampered_journal_pinned_output_is_refused (test_core.CoreTests.test_tampered_journal_pinned_output_is_refused) ... ok
test_uninitialized_submodule_deferred_not_passed (test_core.CoreTests.test_uninitialized_submodule_deferred_not_passed) ... ok
test_unpinned_extra_output_is_refused (test_core.CoreTests.test_unpinned_extra_output_is_refused) ... ok
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
test_run_config_reload_waits_for_start_publication (test_journal_publication.JournalPublicationTests.test_run_config_reload_waits_for_start_publication) ... ok
test_shared_operation_lock_alone_reproduces_both_publication_races (test_journal_publication.JournalPublicationTests.test_shared_operation_lock_alone_reproduces_both_publication_races) ... ok
test_start_publication_blocks_predecessor_snapshot (test_journal_publication.JournalPublicationTests.test_start_publication_blocks_predecessor_snapshot) ... ok
test_terminal_publication_blocks_predecessor_snapshot (test_journal_publication.JournalPublicationTests.test_terminal_publication_blocks_predecessor_snapshot) ... ok
test_unlocked_run_config_reload_reproduces_parse_failure_and_orphan (test_journal_publication.JournalPublicationTests.test_unlocked_run_config_reload_reproduces_parse_failure_and_orphan) ... ok
test_r101_authority_outside_study_review_dirs_refused (test_preparation_evidence.CoreReviewTests.test_r101_authority_outside_study_review_dirs_refused) ... ok
test_r101_cross_scope_pooling_refused (test_preparation_evidence.CoreReviewTests.test_r101_cross_scope_pooling_refused) ... ok
test_r101_digest_in_other_scope_refused (test_preparation_evidence.CoreReviewTests.test_r101_digest_in_other_scope_refused) ... ok
test_r101_exact_approval_scope_and_digest (test_preparation_evidence.CoreReviewTests.test_r101_exact_approval_scope_and_digest) ... ok
test_r101_negated_verdict_refused (test_preparation_evidence.CoreReviewTests.test_r101_negated_verdict_refused) ... ok
test_r101_other_scope_refused (test_preparation_evidence.CoreReviewTests.test_r101_other_scope_refused) ... ok
test_r101_real_nonapproval_records_refused (test_preparation_evidence.CoreReviewTests.test_r101_real_nonapproval_records_refused) ... ok
test_r103_deleted_prior_pair_detected (test_preparation_evidence.CoreReviewTests.test_r103_deleted_prior_pair_detected) ... ok
test_r103_duplicate_fetch_refuses_before_journal (test_preparation_evidence.CoreReviewTests.test_r103_duplicate_fetch_refuses_before_journal) ... ok
test_r103_failed_unreviewed_operation_refused (test_preparation_evidence.CoreReviewTests.test_r103_failed_unreviewed_operation_refused) ... ok
test_r103_recovery_must_cover_all_output_pins (test_preparation_evidence.CoreReviewTests.test_r103_recovery_must_cover_all_output_pins) ... ok
test_r103_reviewed_failure_retained_outside_operations (test_preparation_evidence.CoreReviewTests.test_r103_reviewed_failure_retained_outside_operations) ... ok
test_r104_shared_whole_file_observations_can_change (test_preparation_evidence.CoreReviewTests.test_r104_shared_whole_file_observations_can_change) ... ok
test_r105_seal_check_has_no_reservation_journal_or_writes (test_preparation_evidence.CoreReviewTests.test_r105_seal_check_has_no_reservation_journal_or_writes) ... ok
test_r106_legacy_producer_gap_disclosed (test_preparation_evidence.CoreReviewTests.test_r106_legacy_producer_gap_disclosed) ... ok
test_r106_new_journal_pins_entry_and_loaded_execution_modules (test_preparation_evidence.CoreReviewTests.test_r106_new_journal_pins_entry_and_loaded_execution_modules) ... ok
test_r107_adoption_existing_reservation (test_preparation_evidence.CoreReviewTests.test_r107_adoption_existing_reservation) ... ok
test_r107_adoption_filename_order (test_preparation_evidence.CoreReviewTests.test_r107_adoption_filename_order) ... ok
test_r107_adoption_twice (test_preparation_evidence.CoreReviewTests.test_r107_adoption_twice) ... ok
test_r107_adoption_unfinished_operation (test_preparation_evidence.CoreReviewTests.test_r107_adoption_unfinished_operation) ... ok
test_r107_id_stem_mismatch (test_preparation_evidence.CoreReviewTests.test_r107_id_stem_mismatch) ... ok
test_r107_initial_seal_refusal_preserves_reservation_and_failure (test_preparation_evidence.CoreReviewTests.test_r107_initial_seal_refusal_preserves_reservation_and_failure) ... ok
test_r107_journal_before_clock (test_preparation_evidence.CoreReviewTests.test_r107_journal_before_clock) ... ok
test_r107_missing_required_action (test_preparation_evidence.CoreReviewTests.test_r107_missing_required_action) ... ok
test_r107_noncanonical_start_record (test_preparation_evidence.CoreReviewTests.test_r107_noncanonical_start_record) ... ok
test_r107_stale_journal_clock (test_preparation_evidence.CoreReviewTests.test_r107_stale_journal_clock) ... ok
test_r107_start_terminal_mismatch (test_preparation_evidence.CoreReviewTests.test_r107_start_terminal_mismatch) ... ok
test_r107_terminal_before_start (test_preparation_evidence.CoreReviewTests.test_r107_terminal_before_start) ... ok
test_r107_unfinished_journal (test_preparation_evidence.CoreReviewTests.test_r107_unfinished_journal) ... ok
test_r102_predecessors_are_exactly_documentary_gate_pins (test_preparation_evidence.CoreSealReviewTests.test_r102_predecessors_are_exactly_documentary_gate_pins) ... ok
test_r103_required_record_from_failed_operation_refused (test_preparation_evidence.CoreSealReviewTests.test_r103_required_record_from_failed_operation_refused) ... ok
test_r107_clock_create_once_and_refusals (test_preparation_evidence.CoreSealReviewTests.test_r107_clock_create_once_and_refusals) ... ok
test_r107_integrated_unaccounted_output_refused (test_preparation_evidence.CoreSealReviewTests.test_r107_integrated_unaccounted_output_refused) ... ok

----------------------------------------------------------------------
Ran 332 tests in 71.030s

OK
```

## P3-D adopted for subsequent execution; observations deferred

P3-D replaces the prior requirement to disable updates before Step 5. No root or settings change is required. This response adopts the specified detection-and-stop protocol; it does not perform O0 or claim any observation baseline. It takes effect for execution when restated in the zero-issue implementation round.

Each future observation records argv, exit status, complete stdout, digests and timestamps:

1. `/usr/bin/sw_vers`: ProductVersion `26.6`, BuildVersion `25G72`, no ProductVersionExtra.
2. `/usr/sbin/sysctl -n kern.bootsessionuuid kern.boottime kern.osversion`.
3. Fresh pinned Python `-I -B` process queries `_dyld_get_shared_cache_uuid`; UUID must be `d64a142b-8533-3e09-968a-6ad3e176e3f6` and equal bytes 88–104 of `/System/Volumes/Preboot/Cryptexes/OS/System/Library/dyld/dyld_shared_cache_arm64e`.
4. SHA-256 of every active-cache-family file at O0, O2, O4 and before seal.
5. `/usr/bin/defaults read /Library/Preferences/com.apple.SoftwareUpdate`; existence and readable content of `/Library/Managed Preferences/com.apple.SoftwareUpdate.plist`; `/usr/sbin/softwareupdate --history`. Never `--list`.

Schedule: O0 before GR; O1 before Step 5; O2 after Step 5; O3 before Step 7; O4 after Step 7. Items 1–3 also run before each binding launch session and before seal.

Stop rules are unchanged from the council record: items 1, 3 and 4 equal O0 at every observation; boot-session UUID stays constant O1–O4; preflight/canonical closure active-cache UUID and cache-file pins equal O0 and their platform.version equals O0 sw_vers. Any difference or failed observation means stop before the next step, retain outputs, create nothing further, and open a council round. From O1 until seal the owner does not restart or accept update prompts.

Asking the MDM administrator to defer updates, including Background Security Improvements, until 2026-12-24 is non-blocking. The later RESULTS.md deviation must record that M7 prevention could not be achieved and P3-D replaced it. No existing RESULTS.md was edited in this code-only assignment.

## Hand-off

All three findings have catching tests and passing exercised results. Council round 2 must review these new digests and restate P3-D plus P3/P4 with the v2 entrypoint before subsequent execution. O0, GR and Steps 5–10 remain unrun and require zero-issue implementation sign-off. This response grants no execution authority itself.
