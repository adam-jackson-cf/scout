# Closure implementation review round 1 — response

Status: all 11 required fixes applied; submitted for round-2 review, not a council sign-off. Read-only reuse of the signed `scout-x1/.tools` installation is retained. No bootstrap, canonical closure/preflight, scanner acceptance, model CLI, network fetch or installation modification was performed.

Paths below are relative to `evaluation/x1-triage-ranking/execution/`. Catching tests are in `tests/test_closure_review.py`, unless stated otherwise.

## Issue-to-fix and catching-test map

| Issue | Applied fix (file:line) | Catching tests |
|---|---|---|
| R1-01 | `attest_closure.py:163` restricts marker variables/operators and compares against the running interpreter; `:183` retains applicable hash-pinned rows only and refuses duplicate applicable names; `:206` compares that set with installed distributions. | `test_r101_real_marker_lines`, `test_r101_unknown_marker_variable_and_operator`, `test_r101_duplicate_applicable_row`; real-byte dry check below finds 53 applicable/installed names, deprecated 1.3.1 and librt 0.16.0. |
| R1-02 | `closure/native.py:13` reads the Mach-O filetype (including arm64 slices in fat binaries); `:44` expands executable-relative names/rpaths; `:59` carries executable contexts through the native dependency queue. Executables seed contexts before initial standalone-library entries; contextless unresolved loads still refuse. | `test_r102_executable_context_carried_through_rpath`, `test_r102_missing_executable_context_refuses`; real-byte dry check resolves the signed interpreter's exact libpython load to the existing pinned installation. |
| R1-03 | `attest_closure.py:320` accepts only exactly-two-space pnpm package headers and strips one balanced quote pair. Nested peer keys cannot become packages. | `test_r103_nested_peer_keys_are_not_packages`, using the real @vitest/mocker peerDependencies/peerDependenciesMeta block; dry check finds 133 locked identities and 54 installed virtual-store packages. |
| R1-04 | `attest_closure.py:247` walks only root/lock-listed-package installation-layout node_modules, not arbitrary directories within package tarballs. Internal dist/node_modules bytes remain in the full Inventory. | `test_r104_internal_tarball_node_modules_pass`, `test_r104_stray_top_level_package_refuses`, `test_r104_stray_nested_dependency_refuses`; dry check validates 122 npm lock-listed entries. |
| R1-05 | `attest_closure.py:501`, called before pnpm interpretation, verifies installed package.json, pnpm-lock.yaml and pnpm-workspace.yaml against signed submodule working-file sha256/bytes. Those observations enter source_pins. | `test_r105_consistently_substituted_installed_locks_refuse`; real-byte dry check also verifies all three signed installed assets. |
| R1-06 | `attest_closure.py:507` requires five passing object receipts with identical Git identities, validates the resolved executable and its git-core keg; `:563` uses it for Git version/submodule HEAD, puts its bin first in PATH, enumerates the keg and includes every regular file in native selection. Output includes `git` and object-receipt pins. Extra dylib kegs are supplied through `--external-root` as specified by the review. | `test_r106_git_agreement_and_helper_root`, `test_r106_git_disagreement_refuses`. |
| R1-07 | `attest_closure.py:554`, invoked by `go_metadata` at `:408`, refuses selected ROOT files absent from source pins or the authenticated, unchanged reviewed model_free/*.go set from `:546`. The execution-implementation review is a run prerequisite; selected Go bytes remain retained/rechecked. | `test_r107_unpinned_selected_source_refuses`, `test_r107_reviewed_frontend_bytes_must_stay_signed`. |
| R1-08 | `attest_closure.py:461` defines exactly the five-file own import/config closure for journal source_pins and output. `:466` no longer requires execution_sources/source_manifest; the whole-manifest loader and both fields were removed. Go selected files remain separate. Core confirmed no run-config contract change is necessary; binding/seal retain whole-execution coverage. | `test_r108_own_closure_is_not_whole_execution_manifest`; `test_r111_recheck_planted_null_pass` succeeds without either removed field. |
| R1-09 | `attest_closure.py:530` requires the literal signed numerical command location and the pinned running interpreter/-I/-B; `:68` enforces each provided resolved_path; `:540` requires launcher interpreters to lie inside Inventory roots. Runtime provenance now records an identity mapping for historical read-only reuse. | `test_r109_relocated_runtime_refuses`, `test_r109_resolved_path_must_equal_signed_path`, `test_r109_interpreter_outside_inventory_refuses`, `test_r109_wrong_running_interpreter_refuses`, `test_r109_unisolated_interpreter_refuses`. |
| R1-10 | `attest_closure.py:563` passes every enumerated regular file to `closure/native.py:59`; `:13` selects Mach-O images by bytes, never executable mode or suffix. | `test_r110_nonexecutable_node_addon_is_inspected`. |
| R1-11 | `tests/test_closure_review.py:18` adds planted-null/drift boundary tests for the installed-closure recheck, deferred-row count/set/fields/empty and nonempty cases, Go go.sum/replace/null cases, and all above requirements. Existing fake Mach-O fixtures in `tests/test_closure.py` now have valid filetype headers. | All `test_r111_*` tests, plus R1-01–R1-10 tests above; full execution suite passes below. |

Also adopted review suggestion S1: `--preflight` derives its output exclusively under SCRATCH/closure-preflight/<operation-id>.json and journals action b; `--dyld-cache-root` is required. No preflight was executed. The review-only `--dry-check` is distinct: it runs only the R1-01–R1-04 real-byte checks, prints stdout, and creates no journal, output or cache. It is not a complete native/runtime attestation.

## Verification history and limits

An initial full-suite run reported 120 tests with two errors: the newly added retarget-link test caught only ValueError while a missing target correctly refused with FileNotFoundError, and the binding owner's policy-order test supplied an empty runtime rather than a synthetic git identity. The former now plants an existing-target retarget and catches enumeration drift; the binding owner fixed its owned fixture to supply synthetic-git. The subsequent complete run is reproduced in full below: 121 tests passed. No product code, preparation.py, run-config.json or signed byte was edited by this response.

The dry check is deliberately narrow: it verifies all Python RECORD rows, the applicable signed lock syntax, signed installed Clawpatch manifests, installed pnpm/npm graphs and the real Python executable-relative libpython resolution. It does not claim the remaining full native dependency/cache walk, Go metadata, immutable retention or canonical closure succeeded. Those remain Main's reviewed preflight/canonical actions after round-2 zero-issue sign-off and the unchanged model_free Go source review prerequisite. Historical installation and shared main .tools must not be bootstrapped or rewritten.

## Commands and complete final outputs

All commands ran with cwd `/Users/adamjackson/Projects/work/scout-x1prep`. Unit-test subprocesses were synthetic only. Captured output combines stdout/stderr without truncation.

### Full execution suite

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py' -v
```

```text
test_both_revisions_resolved_before_any_scanner (test_binding.CompletionTests.test_both_revisions_resolved_before_any_scanner) ... ok
test_bound_policy_refuses_each_input_boundary (test_binding.CompletionTests.test_bound_policy_refuses_each_input_boundary) ... ok
test_calledprocesserror_always_terminates_refused_receipt (test_binding.CompletionTests.test_calledprocesserror_always_terminates_refused_receipt) ... ok
test_resolved_policy_must_match_later_bytes_and_pin (test_binding.CompletionTests.test_resolved_policy_must_match_later_bytes_and_pin) ... ok
test_equal_length_roots_and_missing_head_state_directory (test_binding.FixtureAndParityTests.test_equal_length_roots_and_missing_head_state_directory) ... ok
test_exact_sets_health_multiplicity_coverage_and_durable_link (test_binding.FixtureAndParityTests.test_exact_sets_health_multiplicity_coverage_and_durable_link) ... ok
test_fixture_trigger_constructs_and_revision_difference (test_binding.FixtureAndParityTests.test_fixture_trigger_constructs_and_revision_difference) ... ok
test_path_escape_and_codegraph_refused (test_binding.FixtureAndParityTests.test_path_escape_and_codegraph_refused) ... ok
test_source_manifest_not_mistaken_for_filesystem_pin (test_binding.FixtureAndParityTests.test_source_manifest_not_mistaken_for_filesystem_pin) ... ok
test_bound_zero_start_synthetic (test_binding.ObserverTests.test_bound_zero_start_synthetic) ... ok
test_grandchild_absolute_path_detected (test_binding.ObserverTests.test_grandchild_absolute_path_detected) ... ok
test_live_grandchild_refuses (test_binding.ObserverTests.test_live_grandchild_refuses) ... ok
test_loss_and_incomplete_argv_refuse (test_binding.ObserverTests.test_loss_and_incomplete_argv_refuse) ... ok
test_positive_control_boundaries (test_binding.ObserverTests.test_positive_control_boundaries) ... ok
test_positive_control_requires_bound_failure_and_provider_argv (test_binding.ObserverTests.test_positive_control_requires_bound_failure_and_provider_argv) ... ok
test_executable_interpreter_and_script_boundaries (test_binding.ProcessBoundaryTests.test_executable_interpreter_and_script_boundaries) ... ok
test_private_env_stubs_all_provider_paths_and_no_outer_sandbox (test_binding.ProcessBoundaryTests.test_private_env_stubs_all_provider_paths_and_no_outer_sandbox) ... ok
test_stub_is_logging_only (test_binding.ProcessBoundaryTests.test_stub_is_logging_only) ... ok
test_complete_proof_and_both_shebang_argv_shapes (test_binding_observer.ObserverProofTests.test_complete_proof_and_both_shebang_argv_shapes) ... ok
test_proof_refuses_each_missing_boundary (test_binding_observer.ObserverProofTests.test_proof_refuses_each_missing_boundary) ... ok
test_run_proof_writes_absolute_pin_without_executing_target (test_binding_observer.ObserverProofTests.test_run_proof_writes_absolute_pin_without_executing_target) ... ok
test_runtime_pin_mismatch_refuses_before_launch (test_binding_observer.ObserverProofTests.test_runtime_pin_mismatch_refuses_before_launch) ... ok
test_exact_sudoers_command_and_no_stderr_capture (test_binding_observer.ObserverSessionTests.test_exact_sudoers_command_and_no_stderr_capture) ... ok
test_idle_stream_checks_sequence_without_retaining_payload (test_binding_observer.ObserverSessionTests.test_idle_stream_checks_sequence_without_retaining_payload) ... ok
test_live_orphan_prevents_reset (test_binding_observer.ObserverSessionTests.test_live_orphan_prevents_reset) ... ok
test_real_shape_null_and_script (test_binding_observer.ObserverSessionTests.test_real_shape_null_and_script) ... ok
test_sequence_gap_between_trees_is_fatal (test_binding_observer.ObserverSessionTests.test_sequence_gap_between_trees_is_fatal) ... ok
test_sequential_trees_keep_global_sequence_between_trees (test_binding_observer.ObserverSessionTests.test_sequential_trees_keep_global_sequence_between_trees) ... ok
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
test_deadline_exact_seconds_and_leap_boundary (test_core.CoreTests.test_deadline_exact_seconds_and_leap_boundary) ... ok
test_exception_classification (test_core.CoreTests.test_exception_classification) ... ok
test_exclusive_creation_preserves_original (test_core.CoreTests.test_exclusive_creation_preserves_original) ... ok
test_hd9_only_state_date_normalized (test_core.CoreTests.test_hd9_only_state_date_normalized) ... ok
test_initialized_submodule_checks_working_bytes (test_core.CoreTests.test_initialized_submodule_checks_working_bytes) ... ok
test_initialized_submodule_mismatch_refused (test_core.CoreTests.test_initialized_submodule_mismatch_refused) ... ok
test_merge_first_parent_and_offpath_refusal (test_core.CoreTests.test_merge_first_parent_and_offpath_refusal) ... ok
test_promisor_and_extra_ref_refused (test_core.CoreTests.test_promisor_and_extra_ref_refused) ... ok
test_recursive_pin_mismatch_and_missing (test_core.CoreTests.test_recursive_pin_mismatch_and_missing) ... ok
test_seal_deferred_rows_must_all_be_verified (test_core.CoreTests.test_seal_deferred_rows_must_all_be_verified) ... ok
test_seal_missing_and_mismatched_canonical_record (test_core.CoreTests.test_seal_missing_and_mismatched_canonical_record) ... ok
test_seal_missing_records_reservation_not_replaced (test_core.CoreTests.test_seal_missing_records_reservation_not_replaced) ... ok
test_start_binds_uninitialized_submodule_receipt_without_hashing_deferred_rows (test_core.CoreTests.test_start_binds_uninitialized_submodule_receipt_without_hashing_deferred_rows) ... ok
test_uninitialized_submodule_deferred_not_passed (test_core.CoreTests.test_uninitialized_submodule_deferred_not_passed) ... ok
test_unsafe_existing_repository_refused (test_core.CoreTests.test_unsafe_existing_repository_refused) ... ok
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
Ran 121 tests in 4.583s

OK

EXIT_CODE: 0
```

### Read-only historical dry check

```sh
/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python -I -B evaluation/x1-triage-ranking/execution/attest_closure.py --dry-check --runtime-root /Users/adamjackson/Projects/work/scout-x1
```

```text
{
  "R1-01": {
    "applicable_names": 53,
    "deprecated": "1.3.1",
    "installed_distributions": 53,
    "librt": "0.16.0"
  },
  "R1-02": {
    "load_command": "@executable_path/../lib/libpython3.11.dylib",
    "python": {
      "bytes": 49968,
      "expected_resolved_path": null,
      "mode": 493,
      "path": "/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11",
      "requested_path": "/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python",
      "sha256": "570b41051d135bcfc011b41272224f7cc66b0d0e7f735761dcb8611029f51d6d",
      "shebang": null,
      "symlink_chain": [
        {
          "link_path": "/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python",
          "link_target": "/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11"
        }
      ]
    },
    "resolved_libpython": {
      "bytes": 18495136,
      "path": "/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/lib/libpython3.11.dylib",
      "sha256": "9bb0abed613dca8b5a578d2ebf0465a59e3c25946dddade90c2dbd6a90853efb"
    }
  },
  "R1-03": {
    "installed_packages": 54,
    "locked_packages": 133,
    "signed_assets": [
      {
        "bytes": 1562,
        "expected_resolved_path": null,
        "mode": 420,
        "path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/package.json",
        "requested_path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/package.json",
        "sha256": "b878e7ecf4f988a11ad9de8ae4412a32abdeff89ebf6fcc89296aea39ff4e818",
        "shebang": null,
        "symlink_chain": []
      },
      {
        "bytes": 45009,
        "expected_resolved_path": null,
        "mode": 420,
        "path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/pnpm-lock.yaml",
        "requested_path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/pnpm-lock.yaml",
        "sha256": "f546ec43c1dced67f3b3015dd1ba387fa3a999c648a30eda80475dc546de2b44",
        "shebang": null,
        "symlink_chain": []
      },
      {
        "bytes": 69,
        "expected_resolved_path": null,
        "mode": 420,
        "path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/pnpm-workspace.yaml",
        "requested_path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/pnpm-workspace.yaml",
        "sha256": "beb856bbf5b1db5da0186b42aeb420f2a09526acdc76ce3db2843ec56844d45f",
        "shebang": null,
        "symlink_chain": []
      }
    ]
  },
  "R1-04": {
    "lock_listed_packages": 122
  },
  "canonical_writes": false,
  "commands": [
    {
      "argv": [
        "/usr/bin/otool",
        "-L",
        "/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11"
      ],
      "cwd": "/Users/adamjackson/Projects/work/scout-x1",
      "ended_at": "2026-11-09T04:05:01.637278Z",
      "exit_code": 0,
      "started_at": "2026-11-09T04:05:01.493893Z",
      "stderr": "",
      "stdout": "/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11:\n\t/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation (compatibility version 150.0.0, current version 2202.0.0)\n\t@executable_path/../lib/libpython3.11.dylib (compatibility version 3.11.0, current version 3.11.0)\n\t/usr/lib/libSystem.B.dylib (compatibility version 1.0.0, current version 1336.61.1)\n"
    }
  ],
  "runtime_root": "/Users/adamjackson/Projects/work/scout-x1",
  "schema": "scout.x1.closure-review-dry-check.v1",
  "status": "passed"
}

EXIT_CODE: 0
```

### Recomputed SHA-256s (including unchanged shared imports)

```sh
/usr/bin/shasum -a 256 evaluation/x1-triage-ranking/execution/attest_closure.py evaluation/x1-triage-ranking/execution/closure/native.py evaluation/x1-triage-ranking/execution/tests/test_closure.py evaluation/x1-triage-ranking/execution/tests/test_closure_review.py evaluation/x1-triage-ranking/execution/preparation.py evaluation/x1-triage-ranking/execution/verify_inputs.py evaluation/x1-triage-ranking/execution/run-config.json
```

```text
22f23e4c83b458035a081b18fc79fc5f4d8f82a450300a44333657c2aa7f18ff  evaluation/x1-triage-ranking/execution/attest_closure.py
5ae267b380c3a841d4058685593a5fab9ca4e351616207d6b0e5a15f34859d07  evaluation/x1-triage-ranking/execution/closure/native.py
e19d3193908101811dd022711489198fe0f68a9d77ae1f7585205e65ecfaa447  evaluation/x1-triage-ranking/execution/tests/test_closure.py
3fa002a1362e754d6528bc296d774f995bc1ac095bf1b28d446a44b53dbe6ea1  evaluation/x1-triage-ranking/execution/tests/test_closure_review.py
1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf  evaluation/x1-triage-ranking/execution/preparation.py
7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310  evaluation/x1-triage-ranking/execution/verify_inputs.py
c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54  evaluation/x1-triage-ranking/execution/run-config.json

EXIT_CODE: 0
```
