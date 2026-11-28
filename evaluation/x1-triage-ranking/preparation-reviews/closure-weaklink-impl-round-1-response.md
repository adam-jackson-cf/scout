# Closure weak-link implementation round 1 response

Status: implementation/evidence submitted for review; **not signed off**. Steps 0–2 only. No GR, new preflight, canonical closure, real journal, bootstrap, network fetch, model CLI, secret use, or privilege escalation was performed.

## Authority and immutable inputs

Authority: `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-preflight-remedy-round-2.md`, SHA-256 `6ebcc683fc32f960d275c4786b40a5e8bd50354f9b11075e25f2261dc268fa01`.

The five signed dependencies remain byte-identical to the authority's digests. No consumer, signed test, accepted record, journal, or existing receipt was edited. All new files remain uncommitted.

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/attest_closure.py` | 46996 | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native.py` | 10692 | `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/preparation.py` | 50257 | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/verify_inputs.py` | 11211 | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/run-config.json` | 5948 | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |

## Step 0: recovery

Created and successfully validated the recovery receipt for refused preflight `082476b0335b4deebe7390ca6b426c0a`. No new operation was opened.

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/reviews/operation-recovery-082476b0335b4deebe7390ca6b426c0a.json` | 1875 | `d27f95c8d45492028ecadc0390896903491991d16d428e060a736912a17c85a7` |

## Step 1: M1–M7

The original stopped prescan is retained unchanged. Main expressly authorised treating rejected `-help` as informational and using the no-argument usage probe. That probe exited 0 and listed `-all_dyld_cache`; the original `-help` exit 1 and diagnostics remain recorded.

- M1: RemoteProcessingBlock is not an active-cache member; CoreAudio, AudioDSPGraph, and TCC are members.
- M2: the deliberately absent image calibration exits 0, with empty stdout and missing-image stderr. Absence is never inferred from exit 0 alone.
- M3: no RemoteProcessingBlock entry exists at `/` or the enumerated immediate child overlays of `/System/Cryptexes` and `/System/Volumes/Preboot/Cryptexes`. Positive-control observations are retained.
- M4: 3656 header occurrences, 3649 unique names, 76990 edges, zero unknown grammar findings. Alias-complete graph: 639 reachable cached names, zero required-absent edges, six reachable effectively weak absent edges. Two aliases individually retain their requested identity spelling. The first incomplete alias-expansion analysis is retained and explicitly superseded.
- M5: unchanged signed `attest()` with a non-journalling sentinel captured 612 Mach-O images, including 14 fat images and 32 cached seeds. Every captured Mach-O received real `otool -L` and `otool -l` reads. Zero unresolvable dependencies and zero absolute missing on-disk dependencies. The exact harness and captured list are pinned by the prescan. There are zero direct M5 edges to the three present system alternates discussed under I7.
- M6: available bytes 1599120711680; cache-family plus unique-file bytes 7584293085. Available space exceeds 2.2 times the measured total. UUID-matched cache selection and file sizes are retained.
- M7: `AutomaticDownload = 1`, `SplatEnabled = 1`. No settings were changed. Main reports an MDM-managed Mac without sudo/root; no escalation was attempted. **The required update/security-response freeze before Step 5 remains unresolved.** This is not claimed as a completed mitigation.

The final additive prescan revision pins the earlier revisions, every measurement command/status/output digest/timestamp, retained full-output artifacts, recovery receipt, M5 harness, corrected graph, I7 evidence, and final suite result. Revision 2's earlier no-I7-amendment conclusion is superseded, not overwritten.

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-weaklink-prescan-2026-11-10.json` | 14877 | `2475337905376d64bba2aec29204a225edffc89abc416a4ed43943c5228aa19a` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-weaklink-prescan-2026-11-10-rev2.json` | 34105 | `79dabedcbafc06b037487b7614811253633461db5c8f0f49dc0c67fe2a089446` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-weaklink-prescan-2026-11-10-rev3.json` | 36087 | `4d982c94f3752c84463fd492c7de8f8d13d783042a3379cf6e5998482a69e1e0` |

## Step 2: I1–I5 implementation

Exactly three new production/test files were added:

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native_v2.py` | 12932 | `a22a2c7f594672eae845a7c56412ab0c846030c08ee40a0919e7a945e975a739` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/attest_closure_v2.py` | 1168 | `c28a9853f826d46e05ddebbd207ac85cc112aec90179a84b7a0800aeecfd6d84` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests/test_closure_weak_links.py` | 23252 | `e19b0d25d3017923a2f7fba591db3865db48f35ea323526cf827f72b4b2d0ed7` |

- I1: the wrapper imports the signed base and collector at module top; both signed-base hashes are checked before exactly the two permitted rebinding assignments. Original source ordering is preserved, followed by wrapper and collector pins. Import itself does not mutate the base. W1–W5 pass, including the real inventory route, journal source pins/argv, and the original own-source assertion.
- I2: membership/filesystem/overlay absence proofs, strict UUID/header/section/attribute parsing, effective weakness, per-edge requiredness, absent-probe memoisation, and explicit-stack cached traversal. A 5,000-image synthetic chain passes without recursion. Required-after-weak, reverse order, and duplicate-target strength regressions pass.
- I3: on-disk executable context, rpaths, IDs, candidate resolution, sorting, and pending traversal remain the signed behaviour; load-command sets determine effective weakness. Required, mixed, unresolved-rpath, and non-system absence cases still refuse. One evidence-backed I7 cached-source amendment is documented below.
- I4: v2 output has top-level cache pins, cached link attributes/requirements, and separately evidenced absent weak links; no nonexistent file pin is emitted. Both unchanged signed consumers accept the v2 fixture. Synthetic executable-context and cached null comparisons match v1 after only the specified added-field/per-image-cache duplication normalisation.
- I5: all synthetic contracts and wrapper tests pass. All three frozen-host real-byte tests execute, not skip: R1 passes, R2 passes, R3 reports the expressly anticipated signed-v1 refusal below. Real passing-null parity is not claimed.

## I7: minimal evidence-backed active-cache system-alternate amendment

The first R1 run reached `/usr/lib/system/libsystem_kernel.dylib`, `/usr/lib/system/libsystem_platform.dylib`, and `/usr/lib/system/libsystem_pthread.dylib`. Each is both a present file and an active-cache member; signed `image_filetype()` refuses each with `Mach-O has no unique arm64 slice`.

Read-only M4 extensions establish that all three are actually loaded from the active cache address range. Both individual and `-all_dyld_cache` CLI output nevertheless emit x86_64, arm64, and arm64e headers for their on-disk alternates. An intermediate census-based extraction therefore also correctly refused ambiguous identity; that failed suite is retained, and the intermediate extraction code was removed.

The final amendment is deliberately limited:

1. For a present `SYSTEM_PREFIXES` target proven to belong to the active cache, use its cached-image representation instead of queuing its on-disk alternate.
2. For that present cached target only, derive the architecture strictly from the already UUID-matched cache file's magic and pass `-arch <architecture>` to the individual `dyld_info` command. The measured magic is `dyld_v1  arm64e\0`. A calibrated `-arch arm64e` query produces one identity.
3. Keep the strict one-header/one-UUID parser, requiredness checks, unknown-grammar refusal, and every negative absence probe unchanged. An unidentifiable cache architecture refuses; no guessed architecture, swallowed classifier exception, hard-coded target exemption, or fallback exists.

Three catching regressions cover present-member source choice/architecture routing, strict identity after architecture selection, and invalid cache magic. Final R1 passes through the real reachable graph. This amendment remains subject to council implementation review; it is not adopted by this response.

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-m4/i7-active-architecture-analysis.json` | 5094 | `dd57dfd0bfffe1be5ae5a401f51fe1ac814e361405b2577701eb3ab90589e562` |
| `/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-m4/alias-disk-classification.json` | 1711 | `2ed439b5628fef790704c1e8543a1e141c60f0cfb0897b0ce49b9e13e0204dd3` |
| `/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-m4/reachable-disk-classification.json` | 50542 | `b0dd1ebbfb3c19d6fab310becb2545fb353c5be988baa3abc017c39f68f992fb` |
| `/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-m4/actual-loaded-system-image-source.json` | 2027 | `e0c4b783992eaa316aea5438bff3bb75b8629c7be5f6959ce028d7c6f68d7e4d` |
| `/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-m4/present-kernel-dyld-info.json` | 1051 | `d801e8ce086ac02313c85761be6f12af1ba952ec3187c5d3b182850467e4592d` |
| `/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-m4/present-platform-arm64e-dyld-info.json` | 829 | `cf5853d4e355777428fe5c7c1e51ccb3a1e70ac38500b53bcf51ce03862ed852` |

## Tests and observed blocker

Exact command, run in `/Users/adamjackson/Projects/work/scout-x1prep`:

```sh
/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python -I -B -m unittest discover -s /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests -p 'test_*.py' -v
```

Complete command/timing/stdout/stderr records for all three runs are retained below. The artifact named `suite-final` is the intermediate census attempt, not the final architecture-selected result.

| File | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-weaklink-impl-round-1-suite-initial.json` | 47460 | `d48cedb51fa216f7dd02779fbe493adce1b1245f8799e245ee1b6f446af94086` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-weaklink-impl-round-1-suite-final.json` | 47922 | `b38ea422e86c7a298bbe8a2d4c2dab42acaf4a5aa2b8ee06ba7aa9fda4056c91` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-weaklink-impl-round-1-suite-architecture.json` | 46272 | `f81561aaeb3d9ad01e8a019ab30cebc27126f9336ea966bdc487277213e13570` |

1. Initial collector: 319 tests in 35.950s; exit 1; R1 multi-arm64-slice error and R3 signed-v1 refusal.
2. Intermediate census extraction: 321 tests in 40.088s; exit 1; R1 multi-header identity error and unchanged R3 refusal.
3. Final architecture-selected collector: **322 tests in 26.309s; 321 pass, one error, zero skips; exit 1**. R1 and R2 pass. All synthetic/wrapper/existing tests pass. Sole remaining error: R3.

R3 keeps the exact `/bin/sh` seed and equality assertions. Signed v1 refuses before the v2 call at absent RemoteProcessingBlock. This is precisely the finding the remedy says to report rather than replacing the seed. Consequently a passing real-null parity comparison is **unmeasured**, and this response does not assert zero issues or a passing suite. No signed byte, real seed, or assertion was changed to manufacture a pass.

### Final full stdout

Empty (SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`).

### Final full stderr

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
test_r1_coreaudio_weak_absence_and_present_weak_targets (test_closure_weak_links.RealByteTests.test_r1_coreaudio_weak_absence_and_present_weak_targets) ... ok
test_r2_required_remote_processing_block_refuses (test_closure_weak_links.RealByteTests.test_r2_required_remote_processing_block_refuses) ... ok
test_r3_bin_sh_null_is_identical_to_signed_v1 (test_closure_weak_links.RealByteTests.test_r3_bin_sh_null_is_identical_to_signed_v1) ... ERROR
test_filesystem_permission_failure_refuses (test_closure_weak_links.WeakLinkTests.test_filesystem_permission_failure_refuses) ... ok
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

======================================================================
ERROR: test_r3_bin_sh_null_is_identical_to_signed_v1 (test_closure_weak_links.RealByteTests.test_r3_bin_sh_null_is_identical_to_signed_v1)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests/test_closure_weak_links.py", line 396, in test_r3_bin_sh_null_is_identical_to_signed_v1
    one = original.collect(*args)
          ^^^^^^^^^^^^^^^^^^^^^^^
  File "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native.py", line 201, in collect
    cached_image(dependency)
  File "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native.py", line 167, in cached_image
    cached_image(dependency)
  File "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native.py", line 167, in cached_image
    cached_image(dependency)
  File "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native.py", line 167, in cached_image
    cached_image(dependency)
  [Previous line repeated 32 more times]
  File "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native.py", line 148, in cached_image
    raise ValueError('ambiguous cached native image identity: ' + name)
ValueError: ambiguous cached native image identity: /System/Library/PrivateFrameworks/RemoteProcessingBlock.framework/Versions/A/RemoteProcessingBlock

----------------------------------------------------------------------
Ran 322 tests in 26.309s

FAILED (errors=1)
```

## Main's remaining prerequisites and commands

- Council must disposition the retained R3 signed-v1 refusal and review the I7 amendment before any GR/new preflight/canonical closure. The specified passing-null real criterion cannot be met on this frozen host with the unchanged signed v1 and required `/bin/sh` seed; this implementation does not silently relax it.
- An authorised host/MDM administrator must provide the required update/security-response-freeze mitigation before Step 5; no sudo/root was available and no privileged command was attempted.
- The full suite command above is the exact reproduce command, including real-byte tests. No additional real action is authorised or recommended by this response. Do not run GR, another preflight, or canonical closure under this response alone.
