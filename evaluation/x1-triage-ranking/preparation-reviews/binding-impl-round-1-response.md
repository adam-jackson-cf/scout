# Binding implementation review round 1 — response

Status: all 13 requested source fixes applied with catching tests; submitted for round-2 review, **not signed off**. No acceptance, privileged observer proof, scanner, candidate scan, model CLI, bootstrap or network fetch was run. No product source, signed byte, preparation.py or run-config.json was edited. Files remain uncommitted.

Paths in the map are relative to `evaluation/x1-triage-ranking/execution/`. Updated review instructions moved from `execution/binding/REVIEW_REQUEST.md` to `preparation-reviews/binding-impl-REVIEW_REQUEST.md`; no review record remains in this component's execution folder.

## Issue-to-fix and catching-test map

| Issue | Applied source fix (file:line) | Catching test(s) |
|---|---|---|
| R1-01 | `binding/acceptance.py:26` places private stubs first on PATH and sets both configured Clawpatch provider paths, including CLAWPATCH_CLAUDE_BIN, for every route. `:89–100` launches the frontend directly through the observation gate, with no outer sandbox. `binding/interlock.py:35` rejects observed executables/scripts outside the pinned sets and unrestricted interpreter arguments. Obsolete outer exec-deny profile removed. `binding/descendant_fixture.py:14` has root/parent/leaf modes only; no denied mode. The exact nesting probe is recorded in `binding/unprivileged-probe.json:11`. | `test_binding.ProcessBoundaryTests.test_private_env_stubs_all_provider_paths_and_no_outer_sandbox`; `test_executable_interpreter_and_script_boundaries`; nested probe output below. |
| R1-02 | Unconditional attempt-coverage gate removed. `binding/observer_proof.py:23,84` generates a scratch-only, source/runtime/stub-pinned same-session proof and validates shebang interpreter/script/argv, Python orphan and pinned-Node spawn ancestry after parent exit, global sequence, exit 97 and zero live descendants. `binding/acceptance.py:83–86,277–279` verifies its absolute pin/session before revision launches and requires it before the fixture pipeline. | `test_binding_observer.ObserverProofTests.test_complete_proof_and_both_shebang_argv_shapes`, `test_proof_refuses_each_missing_boundary`, `test_runtime_pin_mismatch_refuses_before_launch`, `test_run_proof_writes_absolute_pin_without_executing_target`. These use synthetic event records/mocked target execution, not a real privileged proof. |
| R1-03 | `binding/acceptance.py:275–285` starts one ES session for the entire operation, proves it, passes that same instance to all resolution/control/bound/reference launches and checks final session failure. `binding/observer.py:134,179` uses the exact noninteractive sudoers-authorized command and resets only terminal trees, preserving global sequence across roots and idle periods. Revised review instructions specify owner-provided Full Disk Access and the exact NOPASSWD drop-in, not credential-cache refresh or log streaming. | `test_binding_observer.ObserverSessionTests.test_exact_sudoers_command_and_no_stderr_capture`, `test_sequential_trees_keep_global_sequence_between_trees`, `test_sequence_gap_between_trees_is_fatal`, `test_idle_stream_checks_sequence_without_retaining_payload`; `test_binding.CompletionTests.test_both_revisions_resolved_before_any_scanner` also asserts the same observer instance throughout. |
| R1-04 | `model_free/main.go:123` is the shared deterministic scanner-health predicate: all specs present, accounted file coverage/expected keys, no scanner/inventory/mapping/source-drift error, successful/reused heuristic map. Overall RunComplete is not required. `:147,314,338` separately requires every bound feature/review row to be skipped with ReasonNotOptedIn. Omissions/gaps are retained, normalized/sorted and compared by `binding/interlock.py:100` rather than required empty. | `TestSharedScannerHealthBoundaries`, `TestBindingRowsBoundaries`, `TestDuplicateBindingRowsRefuse`; Python `test_exact_sets_health_multiplicity_coverage_and_durable_link`. |
| R1-05 | `binding/fixtures/common/unchanged.py:5` uses the harmless, rule-matching `yaml.load(data, Loader=yaml.Loader)` construct unchanged across revisions. Parent `changed.py:4` has the PERF401 append loop; head `changed.py:5` moves it exactly one line and `:9` adds a head-only complexity-12 function. Go module/feature inputs are retained. `binding/acceptance.py:208` additionally refuses equal synthetic parent/head finding sets. | `test_binding.FixtureAndParityTests.test_fixture_trigger_constructs_and_revision_difference`; deterministic fixture generation/parity-shape tests. **Actual scanner trigger counts are unmeasured**: worker instructions prohibit scanner/acceptance execution. Main must measure the trigger set during the authorized fixture review/run after scoped source sign-off; no measured nonempty finding set is claimed here. |
| R1-06 | `binding/fixtures.py:66–70,112–114` stores repository metadata as repo_path/content_sha256, with literal revision and absolute root. Filesystem pin dictionaries remain reserved for actual absolute path/sha256/bytes records. `:118–128` verifies those metadata fields without turning them into filesystem pins. | `test_binding.FixtureAndParityTests.test_source_manifest_not_mistaken_for_filesystem_pin` creates only a synthetic Git fixture, materializes it, calls preparation.verify_tree_pins on the resulting receipt and checks deterministic commit identity. |
| R1-07 | `model_free/main.go:32,82` includes category, location StartLine/StartByte/path, family, full unchanged AnchorOf and rule version. `:96` requires actual persisted observation-to-issue provenance, a one-to-one fresh-parent mapping and exactly one durable ID for each head occurrence; `:306–327` reads the real store links and discloses carried IDs. `binding/interlock.py:116–123` checks complete key/link coverage before returning parity. | `TestOccurrenceKeyIndependentExpectedValues` fixes a literal full expected key and hash independently; `TestSecretOccurrenceKeyIndependentExpectedValue`; `TestDurableLinkBoundaries`; `TestOneOccurrenceCannotLinkToTwoIssues`; Python multiplicity/link comparisons. |
| R1-08 | `model_free/main.go:166,176,200` mirrors the planner's eligible/nonexcluded/readable file ordering and batch positions, records each spec's effective applicable/accounted paths and coverage keys, and drops off-file observations like splitAdapter. `binding/acceptance.py:23` uses equal-length bound/reference roots. `binding/interlock.py:100–104` requires exact coverage, effective scanner-input, omission and gap equality **before** comparing findings. | `TestPlannerInputsAndScannerManifest` covers eligibility/exclusion/content/readability/applicability; `test_equal_length_roots_and_missing_head_state_directory`; `test_exact_sets_health_multiplicity_coverage_and_durable_link`. |
| R1-09 | `model_free/main.go:273,286,295–297` adds --resolve-only and returns before registry/state/Assess. `binding/acceptance.py:180–190` resolves all parent/head bound/unbound policies (and reference records) before any scanner. `:72,120–121` validates each expected record's pin and requires later bytes to match exactly. | `TestResolveOnlyRequiredArguments`, `TestResolveOnlyWritesExclusiveMatchingRecordWithoutScanner`; `test_binding.CompletionTests.test_both_revisions_resolved_before_any_scanner`, `test_resolved_policy_must_match_later_bytes_and_pin`. |
| R1-10 | `binding/acceptance.py:128–141` creates head/.scout at 0700 if absent and copies parent state in sorted order while excluding policy.yaml and refusing conflicting/symlink state. | `test_binding.FixtureAndParityTests.test_equal_length_roots_and_missing_head_state_directory`. |
| R1-11 | `binding/observer.py:91–95` accepts script:null and validates/retains a path only for non-null script objects. Real-shaped synthetic fixture variants are used; actual host shape must pass the required proof before scanners. | `test_binding_observer.ObserverSessionTests.test_real_shape_null_and_script`; proof shape tests exercise interpreter-prefixed and direct script argv variants. |
| R1-12 | `binding/acceptance.py:253–254` exclusively reserves and fsyncs the canonical receipt before work. `:228–239` catches Exception, writes one terminal refused receipt and finishes the journal even for CalledProcessError/TimeoutExpired/TypeError-class ordinary failures. Crashes preserve the reservation rather than permit a silent rerun. | `test_binding.CompletionTests.test_calledprocesserror_always_terminates_refused_receipt`, injecting subprocess.CalledProcessError and asserting terminal journal/refusal plus create-once preservation. |
| R1-13 | Go table tests now use valid enabled-policy fixtures and independently fixed key/hash expectations. They cover digest drift, non-model-free policies, key presence, control authorization/enabled-lane preconditions, health and durable linkage. Python tests cover actual-executable/interpreter/Node/script refusals, multiplicity/input/health/link mismatch, unmatched positive-attempt PID/provider argv/exit and stub-source safety. | `main_test.go:31,48,60,72,92,110,137,151,180,192,225,231,236`; `tests/test_binding.py:59,65,97,125,139,245`; `tests/test_binding_observer.py:136`. |

### Planner/filter clarification for R1-08

The required effective input filtering is present. Product `inventory/plan.go:104–118` plans eligible, nonexcluded, content-digested files; job order is fixed at `:447–467`. `Assess` preserves those files in `Spec.Batches`; `Runner.selectInputs` applies `Spec.Applies` inside each batch (`internal/analyzers/runner.go:343–369`). Prefiltering applicability **before** Batches would change the very argv-budget boundaries under review. The reference therefore preserves planner batch positions but records/compares only effective applicable/readable/accounted paths within them. Tests cover nonapplicable/excluded/missing files, and bound/reference roots have equal byte lengths. The finding comparison still retains exact full sets and multiplicity, never just intersections/counts.

Additional nonblocking review suggestions adopted: positive controls must match the production provider argv and stub exit 97, and the bound predicate must reject the same positive-control evidence; Python -c and unpinned scripts are refused; six flags are derived from their respective evidence rather than unconditionally assigned true.

## Exercised verification and limits

No acceptance command or real privileged observer proof was run. Tests launch only synthetic Git/resolve-only subprocesses or mocked/synthetic observation records, never scanners/models. The owner still must supply responsible-launcher Full Disk Access and the limited sudoers drop-in; updated instructions are in `binding-impl-REVIEW_REQUEST.md`. Re-review must reach zero issues before the observer proof or fixture scanners.

An initial integrated Go test exposed invalid test-policy fixtures (missing the policy parser's required enabled-decision model/acknowledgement and provider fields). Those literals were corrected; the final Go test passes. The complete failing and final outputs are preserved below, not omitted. Go vet passed. The shared Python suite initially found the newly introduced test's empty mocked runtime; it was corrected to include synthetic-git before the final shared run. The final complete Python suite was executed once by X1PrepClosure under Main's explicit verification request and reused here to avoid duplicate integrated runs. Its full output was read from `closure-impl-round-1-response.md:42–169`, not inferred; 121 tests passed.

All commands below used cwd `/Users/adamjackson/Projects/work/scout-x1prep`.

## Commands and complete outputs

### Required unprivileged nesting probe

```sh
/usr/bin/sandbox-exec -p '(version 1)(allow default)' /usr/bin/sandbox-exec -p '(version 1)(allow default)' /usr/bin/true
```

Exit code: **0**. stdout and stderr were both empty. The initial tool delivery backgrounded and omitted an exit code, so the same harmless probe was captured via subprocess.run to obtain the exact status. This host observation does not support claiming sandbox_apply failed. The outer sandbox was nevertheless removed exactly as requested by R1-01.

### Shared full execution suite

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py' -v
```

Complete captured combined stdout/stderr (source: sibling's persisted verification record):

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

EXIT_CODE: 0```

### Go check 1

```sh
source .tools/activate && GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 go vet ./evaluation/x1-triage-ranking/execution/...
```

Exit code: 0.

stdout (complete):
```text
```

stderr (complete):
```text
```

### Go check 2

```sh
source .tools/activate && GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 go test ./evaluation/x1-triage-ranking/execution/...
```

Exit code: 1.

stdout (complete):
```text
--- FAIL: TestFrozenDigestAndModelFreeBoundaries (0.00s)
    --- FAIL: TestFrozenDigestAndModelFreeBoundaries/decisions_enabled (0.00s)
        main_test.go:45: invalid policy: decisions.model is required when decisions.enabled is true
    --- FAIL: TestFrozenDigestAndModelFreeBoundaries/contextual_enabled (0.00s)
        main_test.go:45: invalid policy: model_cli.provider is required when model_cli.lanes is set
    --- FAIL: TestFrozenDigestAndModelFreeBoundaries/review_enabled (0.00s)
        main_test.go:45: invalid policy: model_cli.provider is required when model_cli.lanes is set
--- FAIL: TestKeyPresenceAndControlBoundaries (0.00s)
    --- FAIL: TestKeyPresenceAndControlBoundaries/authorized_enabled (0.00s)
        main_test.go:57: invalid policy: decisions.model is required when decisions.enabled is true
    --- FAIL: TestKeyPresenceAndControlBoundaries/absent_authorization (0.00s)
        main_test.go:57: invalid policy: decisions.model is required when decisions.enabled is true
    --- FAIL: TestKeyPresenceAndControlBoundaries/nonexact_authorization (0.00s)
        main_test.go:57: invalid policy: decisions.model is required when decisions.enabled is true
    --- FAIL: TestKeyPresenceAndControlBoundaries/decisions_disabled (0.00s)
        main_test.go:57: invalid policy: model_cli.provider is required when model_cli.lanes is set
    --- FAIL: TestKeyPresenceAndControlBoundaries/lane_absent (0.00s)
        main_test.go:57: invalid policy: decisions.model is required when decisions.enabled is true
    --- FAIL: TestKeyPresenceAndControlBoundaries/review_alone (0.00s)
        main_test.go:57: invalid policy: decisions.model is required when decisions.enabled is true
FAIL
FAIL	scout/evaluation/x1-triage-ranking/execution/model_free	0.685s
FAIL
```

stderr (complete):
```text
```

### Go check 3

```sh
source .tools/activate && GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 go test ./evaluation/x1-triage-ranking/execution/...
```

Exit code: 0.

stdout (complete):
```text
ok  	scout/evaluation/x1-triage-ranking/execution/model_free	0.660s
```

stderr (complete):
```text
```

## Recomputed owned-file SHA-256

Paths relative to `evaluation/x1-triage-ranking/`. This response's own SHA-256 is supplied separately in the final handoff, avoiding a self-hash cycle. The request was moved out of execution; no old-path compatibility shim remains.

```text
5e75e907392663394f01afa9d9d0f365568c3ae0f8bdc7cd0c222aa5ec59180c  execution/binding/acceptance.py
669c2a58dedc063e21a617479020b2eeacbaca684fd0123b871c89b68493420c  execution/binding/descendant_fixture.py
ec75b6e87719a439d7ec4959302e25fd588319c7f7cfd1e6e20b39a324aff0b0  execution/binding/fixtures/common/go.mod
2fa9c2c5a47269dd22d1da68d24423b1a2d3a2da635c4c60798c59e17ade5d23  execution/binding/fixtures/common/index.js
d37528f6027aa09adca50fa54e40ffc247c5597ffb7d4effa633ccba6134536c  execution/binding/fixtures/common/main.go
514ef8299051029ada3268f1fb2eba8b698548d59bb6d438d71c42c03a627f9e  execution/binding/fixtures/common/package.json
534f740e38f9a8c5e7226a676a427bc5b44ec64ab605c33c595a68aedf5a97c4  execution/binding/fixtures/common/unchanged.py
6189fea6ad039f0eb0206a7986aa94a44e3bdf46669184298b8a558d06fa7255  execution/binding/fixtures/head/.scout/policy.yaml
0950bd4989dfc76b1c8ed8641c91b4994cefee6bc5bb6d5105510f99f25ec109  execution/binding/fixtures/head/changed.py
3d3dae7d063c4fe94127de4e67ae9315e1495e4fa510ea587ae4c5a058604ec3  execution/binding/fixtures/parent/.scout/policy.yaml
0ecf4e2fd0b76f8305f0729807a67bc7d2eeabe27ed637a98bda8b9d87738e3b  execution/binding/fixtures/parent/changed.py
48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b  execution/binding/fixtures.py
72006f7874f481dcbbc7ee61c7472aca9b08c0bd7286052e39c1960794860ff0  execution/binding/gate.py
dc12b4c45be9e01f182c6253941ab51543787ba1b436a1ce037871cf9ecb4116  execution/binding/interlock.py
af674a9bbf22c94df02eac5970a6d1f414be99e8dc2b13ce983c588ffa905553  execution/binding/observer.py
91dad45fc0e8f6fe43315dcafc8bb5566589b2e392e2d11ddf6af0e62595a4c3  execution/binding/observer_proof.py
10a0c930a3b3d006e7a9c368ea7a93f709ebfda13a149b7521dc1c5721ef4881  execution/binding/unprivileged-probe.json
235ce0fc95e8bffbc33ac1929c22d440838bd9db9dfebddb664f716bb852222f  execution/model_free/main.go
c013b037fb26d61374a0edff84a316bda05bafa93a428946f8271783457ca205  execution/model_free/main_test.go
870e6eaf514f557703399ed8649e1a401856e58ff6f9a07f387bf43a46ac05d9  execution/tests/test_binding.py
645133a7a0231714b0a648c31b66c4a01432405d825f675aacb568d5afdc85db  execution/tests/test_binding_observer.py
b11f46ee492915771576f5bb0376535753c614f1db6c6551ecad3f7deb22eb64  preparation-reviews/binding-impl-REVIEW_REQUEST.md
```
