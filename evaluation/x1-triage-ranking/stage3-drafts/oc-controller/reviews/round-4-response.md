# OC controller — round-4 remediation response

**Draft; awaiting scoped zero-issue review.** Only the two prescribed fixes are implemented. Policy, ordering, seed bytes, disk reservation and accepted sibling contracts are unchanged. No preparation action, authentic OC or real repository data was used.

## Issue → targeted fix → catching test

| Issue | Implementation | Catching test and mutant |
|---|---|---|
| R4-01 (R3-01 residual) | `controller.py:292–320` parses a newline-less final JSON record. If its sequence and predecessor chain match, it applies that authoritative record, hashes `raw + b"\n"`, and retains the delimiter-restoration flag. The existing constructor restores the newline before a process-loss pause whose tail SHA is null. A non-parsing or non-chaining newline-less tail remains torn. No new recovery format or sidecar is introduced. | `test_controller.py:790` fault test adds `raw[:-1]` stage, keeps the concurrent-lane poison assertions, independently folds durable absolute deltas, and closes/reopens **twice**. Each restart must have no chain error and exact durable population state including that record. The newline stage requires null tail SHA. `mutate_round2.py:28` reverts newline-less parsing; it is assertion-killed. |
| R4-02 (R3-06 residual) | Production runtime guard is unchanged. `test_controller.py:768–788` runs optimization/version/implementation subprocess cases with mocked ample free space and requires the exact `requires CPython 3.11.11 with python_optimize=0` diagnostic and absence of freeze. Non-CPython is explicitly injected by patching `platform.python_implementation`. | `test_disk_refusal_and_interpreter_provenance`; `mutate_round2.py:29–31` separately deletes optimization, version and implementation clauses. Each now reaches a successful freeze under the ample-space fixture and is caught by an assertion, independent of host disk space. |

README documents stable newline-less replay and the isolated guard evidence. No nonblocking suggestion or unrelated hardening was added.

## Commands and complete outputs

From worktree root after `source .tools/activate`, using pinned Python 3.11.11 and `PYTHONDONTWRITEBYTECODE=1`:

```sh
D=evaluation/x1-triage-ranking/stage3-drafts/oc-controller
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/test_controller.py" -v
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/mutate_round2.py"
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/derive_known_answers.py"
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py
```

Full files, not excerpts: `round-4-test-output.txt` (35 tests), `round-4-mutant-output.txt` (all 20 substitutions/commands/assertion tracebacks), `round-4-derivation-output.json` (OpenSSL derivation), `round-4-frozen-output.json` (unchanged checker stdout), `round-4-evaluator-preservation.json` (all accepted sibling digests). **20/20 guard mutants were assertion-killed.** All four new reversions are independently caught; no check failure occurred during this round.

### Full suite output

```text
test_actual_process_loss_mid_safety_replays_same_attempt_and_seeds (__main__.ControllerTests.test_actual_process_loss_mid_safety_replays_same_attempt_and_seeds) ... ok
test_baseexception_is_latched (__main__.ControllerTests.test_baseexception_is_latched) ... ok
test_binding_completion_caps_and_gate_failures (__main__.ControllerTests.test_binding_completion_caps_and_gate_failures) ... ok
test_calibration_baseexception_is_latched (__main__.ControllerTests.test_calibration_baseexception_is_latched) ... ok
test_compact_record_counts_and_lost_gate_replay (__main__.ControllerTests.test_compact_record_counts_and_lost_gate_replay) ... ok
test_construction_base_vector_and_336_by_six (__main__.ControllerTests.test_construction_base_vector_and_336_by_six) ... ok
test_criticality_baseexception_is_latched (__main__.ControllerTests.test_criticality_baseexception_is_latched) ... ok
test_criticality_exception_is_latched (__main__.ControllerTests.test_criticality_exception_is_latched) ... ok
test_criticality_malformed_is_latched (__main__.ControllerTests.test_criticality_malformed_is_latched) ... ok
test_disk_refusal_and_interpreter_provenance (__main__.ControllerTests.test_disk_refusal_and_interpreter_provenance) ... ok
test_eight_concurrent_binding_slots_equal_serial_and_refuse_reentry (__main__.ControllerTests.test_eight_concurrent_binding_slots_equal_serial_and_refuse_reentry) ... ok
test_fault_arriving_after_reviewed_resume_latches_new_incident (__main__.ControllerTests.test_fault_arriving_after_reviewed_resume_latches_new_incident) ... ok
test_gate_baseexception_is_latched (__main__.ControllerTests.test_gate_baseexception_is_latched) ... ok
test_gate_exception_is_latched (__main__.ControllerTests.test_gate_exception_is_latched) ... ok
test_gate_window_classifications_are_isolated (__main__.ControllerTests.test_gate_window_classifications_are_isolated) ... ok
test_gate_window_gold_pause_and_cap_never_start_trial (__main__.ControllerTests.test_gate_window_gold_pause_and_cap_never_start_trial) ... ok
test_gold_cap_and_completion_races_are_single_atomic_termination (__main__.ControllerTests.test_gold_cap_and_completion_races_are_single_atomic_termination) ... ok
test_independent_known_answers_and_canonical_seed_refusals (__main__.ControllerTests.test_independent_known_answers_and_canonical_seed_refusals) ... ok
test_journal_failure_poison_blocks_other_lane_and_replays (__main__.ControllerTests.test_journal_failure_poison_blocks_other_lane_and_replays) ... ok
test_meter_origin_decrease_and_fault_latch_without_meter_read (__main__.ControllerTests.test_meter_origin_decrease_and_fault_latch_without_meter_read) ... ok
test_nonbool_gate_is_latched_not_gate_failure (__main__.ControllerTests.test_nonbool_gate_is_latched_not_gate_failure) ... ok
test_nonbool_trial_is_latched (__main__.ControllerTests.test_nonbool_trial_is_latched) ... ok
test_owned_loop_calibration_order_disclosure_and_outcome_independence (__main__.ControllerTests.test_owned_loop_calibration_order_disclosure_and_outcome_independence) ... ok
test_paused_run_atomically_closes_binding_admission (__main__.ControllerTests.test_paused_run_atomically_closes_binding_admission) ... ok
test_private_readonly_state_and_completion_guard (__main__.ControllerTests.test_private_readonly_state_and_completion_guard) ... ok
test_process_lock_and_torn_tail_preserves_evidence (__main__.ControllerTests.test_process_lock_and_torn_tail_preserves_evidence) ... ok
test_process_loss_between_passing_gate_and_trial_start (__main__.ControllerTests.test_process_loss_between_passing_gate_and_trial_start) ... ok
test_replay_mid_attempt_and_after_gold_and_fixed_freeze (__main__.ControllerTests.test_replay_mid_attempt_and_after_gold_and_fixed_freeze) ... ok
test_safety_handle_guard (__main__.ControllerTests.test_safety_handle_guard) ... ok
test_source_order_policy_refusals_before_freeze (__main__.ControllerTests.test_source_order_policy_refusals_before_freeze) ... ok
test_trial_exception_is_latched (__main__.ControllerTests.test_trial_exception_is_latched) ... ok
test_trial_window_gold_pause_and_cap_classification (__main__.ControllerTests.test_trial_window_gold_pause_and_cap_classification) ... ok
test_unfinished_binding_refuses_safety_and_grid_complete (__main__.ControllerTests.test_unfinished_binding_refuses_safety_and_grid_complete) ... ok
test_unreachable_owned_loop_advances_in_frozen_order (__main__.ControllerTests.test_unreachable_owned_loop_advances_in_frozen_order) ... ok
test_unreachable_population_three_before_attempt_and_profile_reuse (__main__.ControllerTests.test_unreachable_population_three_before_attempt_and_profile_reuse) ... ok

----------------------------------------------------------------------
Ran 35 tests in 1.270s

OK
```

The fresh derivation stdout is byte-identical to unchanged `known_answers.json` (SHA-256 `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5`). It retains policy checksum `9b6de56992b00129edcb09d8c0b1e95e7f3bfb538dc3fea02c0529e94886bfd5`, ordered-list SHA `04aa2184f79cb3795f60305a6b8fa31687d5acd7069846e78b6acf1b1f4dd4ff`, and construction-base seed64 `9337516894895418925`. Policy hashing remains a checksum, not independent design evidence. Full derivation bytes are retained, not reconstructed in this response.

The size unit suite passed 32 tests in 586.982 seconds, all six focused extraction mutants were killed, and the actual full-count final-controller CLI passed in 602.705 seconds. Its complete receipts are under `../locked-size-input/reviews/round-4-cli-evidence/`. Scientific sizes/counts were not reduced. No journal-cost remeasurement was needed: the only production change is replay, not the measured append path; historical cost evidence remains labelled with its original source digest.

## Preserved signed boundaries

Frozen checker: 126/131 literal files; 5/5 owned-extract groups valid; 19/20 literal extracts with permitted HD9 normalization; **131/131 unchanged against baseline `e2b94e3c`**. Rev9 manifest SHA remains `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`. All 17 accepted-evaluator files remain unchanged. No output was written under that signed-off draft, and its `INPUTS` surface was not changed.

No new author/owner choice is introduced. The round-4 council accepted the existing conservative disk reservation, which is unchanged. Size/controller still require scoped zero-issue re-review; this writer requests none. Validation grants no preparation or run authority.

## Recomputed per-file SHA-256

All artefact files except this response itself; its self-digest accompanies the commit head. Historical files are included unchanged.

| File | SHA-256 |
|---|---|
| `README.md` | `e6727b1fa2787ad3fcad3dcd50e2b8c816f5527a39e3421f4df07ccf2e802093` |
| `benchmark_journal.py` | `36d692e228d5f3afbb8951bfdb12aacc6a07a257a48a8f4a2e2f234a269fdf55` |
| `controller.py` | `4627b3bb419dcf0fe5de230609e9e768570b8a616a485015b7c766c02c90f8a0` |
| `derive_known_answers.py` | `c789ea6b2b7d4dbef54ed70ae2ec96a02b89fa6822e3d77181fc443c280edc33` |
| `known_answers.json` | `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5` |
| `mutate_round2.py` | `f7df46e237ed69b5d6598a0828ea18ed770923560694060892957794b955cb10` |
| `policy.json` | `ceb2ca53ffee1699ffabe03d10bd3a067d58f2bd92bdc3e20369a4daec88ebba` |
| `reviews/round-1-cap-output.txt` | `1d442d16d51c04f4120587d56c4638fa9035e88280376da4fccd70cb690cfeb4` |
| `reviews/round-1-response.md` | `682af9a486173b56c3a302e93251134f2592a3a0a9100ac174c4d6c8296b8ce8` |
| `reviews/round-1-test-output.txt` | `4d970a1f1c77544377c8dff41a23a6a3efae447b0ec28b8e324d467eb534f91a` |
| `reviews/round-1.md` | `f5a1c845f81d3d22e9d357c03e7c233c382317d8894e67551a10a3558a90539b` |
| `reviews/round-2-derivation-output.json` | `012699629663f464af546f8e5048e730af711b49f9932e3e4b6db115a5a24449` |
| `reviews/round-2-mutant-output.txt` | `48d9c17a3e41adfd12186ecc2559d842ad54d4dae778383afe26d5884fe58467` |
| `reviews/round-2-response.md` | `02f0cac73fd13f66699c974e8a4dd7dab64cfa5281928b1c25acdd41c1084dbf` |
| `reviews/round-2-test-output.txt` | `5b25a2d02a93e746da046c9233da84c2df22295243279f8eb181151962db79dd` |
| `reviews/round-2.md` | `d3b770d948fa2b7162f57a281b27a9201c3283e3f22f94436c40a4499945e22e` |
| `reviews/round-3-derivation-output.json` | `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5` |
| `reviews/round-3-evaluator-preservation.json` | `c6266f87a13019f656ed26f5490a8c739fa0664a79f51588408a965ee9cfd48c` |
| `reviews/round-3-frozen-output.json` | `2c5d8a155e57cdd60a2b6ea295e9c0c1a2453acccf0ce662a60e711a85df50d0` |
| `reviews/round-3-initial-test-output.txt` | `e49ae6bc05b7624c77b59e39e71f460f1d1cb9560d53a412bf62a01a582b1553` |
| `reviews/round-3-journal-output.json` | `24a698b5cc9e6e0642318d19a4395f75868c2e18d7aa783756ae275d5ea223dc` |
| `reviews/round-3-mutant-initial-output.txt` | `92312b4191a7e02e2b385f3ba74305d113829cb1a8eac82f919b9a2ff246f015` |
| `reviews/round-3-mutant-output.txt` | `7232edb21609784fbe8137e8a124c8a1f9cc7740a35cd08c48075f67f199a5d1` |
| `reviews/round-3-mutant-second-output.txt` | `64258f557b451ca81413c6fc5e557efc4e1ffa1892a5829a8993ada1a52c7260` |
| `reviews/round-3-response.md` | `aded0fa8fe622492cba94bad14528c28b7b81dbc4158231733274a2adb393d52` |
| `reviews/round-3-test-output.txt` | `3fd9a9f8a63538f974564ccaddffa05906107fea860115bb77ade987e5732959` |
| `reviews/round-3.md` | `d8e708a3e133d7f73ae2584d80e5a874f51ffe6c28cfa26ce8b676e44010e4d6` |
| `reviews/round-4-derivation-output.json` | `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5` |
| `reviews/round-4-evaluator-preservation.json` | `c6266f87a13019f656ed26f5490a8c739fa0664a79f51588408a965ee9cfd48c` |
| `reviews/round-4-frozen-output.json` | `2c5d8a155e57cdd60a2b6ea295e9c0c1a2453acccf0ce662a60e711a85df50d0` |
| `reviews/round-4-mutant-output.txt` | `702ad39561cfc84eef0f2f8f9d150b9f59e9b0c270e82ed632d132e206846744` |
| `reviews/round-4-test-output.txt` | `a2070f3f5ae1e0e8ccc54c87501a2abdc667f6bd05a9a5fdbc184f95211dd0b2` |
| `reviews/round-4.md` | `9d9a0d52937a60f7fcdeb1945d5c242cca4abde0c80f2e72ab7e27bae71d466f` |
| `test_controller.py` | `16cc62ea3d8b497ee301e300a41103ecd197cf92129157abfad248e66ba62726` |
