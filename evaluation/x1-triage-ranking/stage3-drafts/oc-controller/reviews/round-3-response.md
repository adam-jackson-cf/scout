# OC controller — round-3 remediation response

Date: 2026-11-09. **Non-authoritative draft; proposed fixes await scoped zero-issue review.** Covers all six round-3 findings, including the R2-07 and R2-11 residuals. No preparation action, authentic inventory, binding implementation, real repository data, scan or model call occurred. The accepted final evaluator, its tests/evidence/README and `provenance.INPUTS` are unchanged; its new `round-3.md` is committed without alteration.

## Issue → implementation → catching test

C = `controller.py`; T = `test_controller.py`. Paths/lines are current artefact-relative anchors.

| Issue | Fix | Catching test |
|---|---|---|
| R3-01, R2-07 residual | C:320–373 writer poison is set under the shared lock before append; partial/short write, flush, fsync and reducer failures leave it set and notify waiters. Every later commit/boundary/work/pause/resume refuses, without trying to journal an uncertain pause. Unbuffered append avoids a later buffered flush. C:292 replay treats a complete record as authoritative; torn tail retains reviewed process-loss recovery. | T:785 `test_journal_failure_poison_blocks_other_lane_and_replays`: five failure sites (including short write), a second lane mid-attempt, no subsequent bytes, blocked operations, restart into process-loss pause, counters equal durable records. Two new poison reversion mutants are assertion-killed. |
| R3-02, R2-07 residual | C:326–378 compact deltas use a short population index and coalesce gate changes until a durable linearisation point; no standalone meter records. C:485 / :500 trial start/result write two records; rejected attempts write none. C:119 storage projection, :217–243 freeze/free-space refusal. `benchmark_journal.py` measures deterministic temporary-file I/O, including separate final gate flush and full-population projections at frozen synthetic p_gate. | T:720 `test_compact_record_counts_and_lost_gate_replay` pins passing=2 fsyncs, rejected=0, deferred flush/replay; T:588 `test_process_loss_between_passing_gate_and_trial_start` injects a loss before durable trial start, closes/reopens and checks same-index replay. The retained mid-safety test separately uses actual subprocess loss. T:749 `test_disk_refusal_and_interpreter_provenance` checks recorded refusal without freeze. Full benchmark output: `round-3-journal-output.json`. |
| R3-03, R2-11 residual | C:462–465 refuses binding leases while run owns its lock, including paused admission; C:699–701 rechecks all leased binding completion atomically with first dispatch. | T:672 `test_paused_run_atomically_closes_binding_admission`: deterministic waiting run, resume and attempted lease; exactly safety admission succeeds and the lease refuses. Removing the run-lock guard is assertion-killed. |
| R3-04 | `policy.json:6`, README construction disclosure, C:609–614 retain the council-approved κ/j-free construction-base bytes using signed `safety-profile` kind, not the outcome-cell identity. Policy checksum re-pinned at C:26. Independent `derive_known_answers.py` adds the exact population-phase construction vector. | T:701 `test_construction_base_vector_and_336_by_six` checks the actual seed supplied to preparation, 336 distinct bases and six κ groups each; independent five-vector fixture also exercised by `test_independent_known_answers_and_canonical_seed_refusals`. |
| R3-05 | Isolating behavioral tests target gate classification, gate/calibration `BaseException`, and forbidden external safety stepping; all four requested reversions are added to `mutate_round2.py`. | T:614 `test_gate_window_classifications_are_isolated`: gold/cap/pause/pause+resume, zero classified passes, pending gate, no late population record and same-index gate retry after resume. T:646/:649 gate/calibration `KeyboardInterrupt` pause with incident. T:660 `test_safety_handle_guard`. Each of four mutants fails an assertion, not an uncaught exception. |
| R3-06 | C:160 rejects non-CPython/non-3.11.11/optimized execution before creating freeze. C:211–225 binds executable real path, SHA-256, version, implementation and optimize flag in expected freeze; replay refuses changed interpreter provenance. | T:749 `test_disk_refusal_and_interpreter_provenance`: actual `-O` subprocess creates no freeze; version/implementation refusals and changed replay interpreter provenance are checked. |

## Exact commands and complete evidence

From the worktree root after `source .tools/activate`; numerical interpreter is `.tools/venv/bin/python` 3.11.11. All calls set `PYTHONDONTWRITEBYTECODE=1`.

```sh
D=evaluation/x1-triage-ranking/stage3-drafts/oc-controller
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/derive_known_answers.py"
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/test_controller.py" -v
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/mutate_round2.py"
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/benchmark_journal.py" --attempts 1000 --p-gate .5
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py
```

The derivation stdout was saved verbatim as both `known_answers.json` and `reviews/round-3-derivation-output.json` **before** suite/mutant execution. No signed source was edited. Complete final outputs (not excerpts):

- `round-3-test-output.txt`: all **35 tests passed in 1.249 s**; reproduced in full below.
- `round-3-mutant-output.txt`: exact substitutions, commands and every assertion traceback; **16/16 assertion-killed** (nine retained probes plus seven requested R3 probes).
- `round-3-journal-output.json`: full measured sample/freezer/projection/runtime/source evidence.
- `round-3-frozen-output.json`: full unchanged-checker stdout, stored outside the accepted evaluator.
- `round-3-evaluator-preservation.json`: all 17 accepted-evaluator files retain their pre-remediation SHA-256s, including the incoming sign-off record.

### Validation-harness corrections, not hidden failures

Initial mutation execution stopped at the poison mutant: assertions correctly failed but stranded the peer thread, producing cleanup errors. `round-3-mutant-initial-output.txt` preserves that complete failed run. The test now always releases/joins its peer and closes the original unbuffered file in `finally`. The initial passing 35-test suite (1.171 s) is retained in `round-3-initial-test-output.txt`.

A second mutation execution reached the gate `BaseException` mutant but escaped as `KeyboardInterrupt`; full output is `round-3-mutant-second-output.txt`. The common fault helper now catches `BaseException` explicitly and asserts its type is `Stopped`. This makes a reversion fail an assertion, rather than abort unittest. The final suite and all 16 mutants passed after these test-only fixes. Controller source/policy did not change during validation. A passing intermediate suite also ran in 3.227 s before the second assertion correction; final evidence below is the current test source.

### Full final unit-suite output

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
Ran 35 tests in 1.249s

OK
```

## Measured synthetic journal cost and storage admission

These are deterministic **instrumentation samples**, not scientific OC trials or new CPU/memory-fit inputs. Each mode performs 1,000 callback attempts with actual temporary-file writes/fsyncs; only synthetic disk admission is mocked. No controller-side randomness is used. The final deferred-gate flush is reported separately and excluded from the per-attempt figures.

| Sample | Executions / rejected gates | Durable records / fsyncs | Bytes per attempt | Microseconds per attempt |
|---|---|---|---|---|
| Passing | 1,000 / 0 | 2,000 / 2,000 | 544.245 | 114.903625 |
| Rejected | 0 / 1,000 | 0 / 0 | 0.0 | 16.196291 |
| Mixed at p_gate .5 | 500 / 500 | 1,000 / 1,000 | 281.276 | 69.76975 |

Mixed final flush: 1 record/fsync, 344 bytes, 58.583 μs. Passing final control flush: 275 bytes, 60.417 μs; rejected gate flush: 342 bytes, 98.834 μs. Full output also records initialization and largest observed record sizes.

At the synthetic frozen comparison p_gate .5, 90,860 binding + 10,080 safety populations imply a minimum 201,880,000 executions and expected 403,760,000 attempts. **Linear sample-scaled inference, not an upper fit bound:** 113,647,942,240 bytes and 28,170,234,260 serial μs. Binding inadequacy can require up to 4,000 executions/population; real control payloads/platform I/O differ. No authentic full-OC CPU/memory/wall or disk measurement is claimed.

The separate conservative freeze reservation is **809,135,040 records × 2,048 = 1,657,108,561,920 bytes**, using 4,000 executions and 16 control records per population. It records p_gate and its 8,000-attempt cap at .5; coalesced gate rejection does not add a record. A smaller volume gets a create-once `freeze-refused.json`, observed free-space evidence and no freeze. This intentionally may refuse a volume that holds the sample-scaled estimate. It is disclosed under README **Open choices for review**, not offered as operator discretion. Reproduce instrumentation at the actual frozen p_gate when that independently authorised input exists. No accepted evaluator `INPUTS` change was needed.

## Independent policy/order/construction seed derivation

Policy checksum: `9b6de56992b00129edcb09d8c0b1e95e7f3bfb538dc3fea02c0529e94886bfd5`.
Order SHA-256: `04aa2184f79cb3795f60305a6b8fa31687d5acd7069846e78b6acf1b1f4dd4ff`.
Construction-base population seed64: `9337516894895418925`; full HMAC: `81958624c4732e2da8d3dadd96e026a8e3756f03b933b53d05988ee674ef2644`.

The standalone script imports neither controller nor signed kernel; it independently enumerates all 2,016 group identities and uses OpenSSL for all order HMACs, array SHA and five seed vectors. Policy SHA is a checksum of the published policy, **not independent design evidence**. Full derivation stdout:

```json
{
  "construction_base": "[\"safety-profile\",{\"cap_M\":2000,\"prevalence\":0.1,\"rank_structure\":\"uniform\",\"realised_gain\":0.2,\"scanner_control_true_at_prevalence\":true,\"size_dependence\":0.05}]",
  "first_identity": "[\"safety-interval-group\",{\"cap_M\":2000,\"prevalence\":0.1,\"rank_structure\":\"uniform\",\"realised_gain\":0.2,\"scanner_control_true_at_prevalence\":true,\"size_dependence\":0.05,\"unresolved_kappa\":3.0}]",
  "group_count": 2016,
  "hmac_order_first": "002310bb001376178f0714c972be461256f769320d18e857e64a2d1d524b1d33",
  "last_identity": "[\"safety-interval-group\",{\"cap_M\":700,\"prevalence\":0.1,\"rank_structure\":\"uniform\",\"realised_gain\":0.2,\"scanner_control_true_at_prevalence\":false,\"size_dependence\":0.05,\"unresolved_kappa\":10.0}]",
  "method": "Independent recipe in derive_known_answers.py; every order HMAC and SHA uses openssl; policy is a checksum of published policy.json, not an independent design derivation",
  "ordered_groups_sha256": "04aa2184f79cb3795f60305a6b8fa31687d5acd7069846e78b6acf1b1f4dd4ff",
  "policy_sha256": "9b6de56992b00129edcb09d8c0b1e95e7f3bfb538dc3fea02c0529e94886bfd5",
  "seeds": [
    {
      "attempt": 0,
      "cell": "[\"safety-interval-group\",{\"cap_M\":2000,\"prevalence\":0.1,\"rank_structure\":\"uniform\",\"realised_gain\":0.2,\"scanner_control_true_at_prevalence\":true,\"size_dependence\":0.05,\"unresolved_kappa\":3.0}]",
      "hmac_sha256": "bb2bb6cbad068c03f323d2a9de146f0ac602a9ef5445732dc1fd23917a079fe6",
      "phase": "population",
      "population": 0,
      "seed64": 13487074494988585987
    },
    {
      "attempt": 0,
      "cell": "[\"safety-interval-group\",{\"cap_M\":2000,\"prevalence\":0.1,\"rank_structure\":\"uniform\",\"realised_gain\":0.2,\"scanner_control_true_at_prevalence\":true,\"size_dependence\":0.05,\"unresolved_kappa\":3.0}]",
      "hmac_sha256": "0f4b3a2681211f53079e8c2e20d5887919b556b563b66236b8502faa8ae7b361",
      "phase": "audit",
      "population": 1,
      "seed64": 1102038470871818067
    },
    {
      "attempt": 17,
      "cell": "[\"safety-interval-group\",{\"cap_M\":2000,\"prevalence\":0.1,\"rank_structure\":\"uniform\",\"realised_gain\":0.2,\"scanner_control_true_at_prevalence\":true,\"size_dependence\":0.05,\"unresolved_kappa\":3.0}]",
      "hmac_sha256": "1ef9732de3ec42b5899893b9cfa11239b5fabd66701a935e3e408571b309afdc",
      "phase": "bootstrap",
      "population": 0,
      "seed64": 2231941731273425589
    },
    {
      "attempt": 0,
      "cell": {
        "kind": "sampling-gate-estimate"
      },
      "hmac_sha256": "f98555ac5c97329b205df7cf34ef7c58400eafd0ee5b3013253ee37552e418a3",
      "phase": "sampling-gate-estimate",
      "population": 0,
      "seed64": 17979871286122721947
    },
    {
      "attempt": 0,
      "cell": "[\"safety-profile\",{\"cap_M\":2000,\"prevalence\":0.1,\"rank_structure\":\"uniform\",\"realised_gain\":0.2,\"scanner_control_true_at_prevalence\":true,\"size_dependence\":0.05}]",
      "hmac_sha256": "81958624c4732e2da8d3dadd96e026a8e3756f03b933b53d05988ee674ef2644",
      "phase": "population",
      "population": 0,
      "seed64": 9337516894895418925
    }
  ],
  "wilson_50_2000_decimal60": [
    "0.0190144235039924754912500557522667442502430374162904373713250",
    "0.0328067714131978119323191293954135255572157765564168470906098"
  ]
}
```

## Frozen basis and accepted sibling

The unchanged checker reports 126/131 literal files, 5/5 owned-extract groups valid, 19/20 literal extracts (only allowed HD9 normalization), and **131/131 unchanged against baseline `e2b94e3c`**. Rev9 manifest SHA remains `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`. All 17 accepted-evaluator files remain byte-identical; only its incoming round-3 receipt is newly committed. No output or response was written beneath that accepted artefact.

The size suite separately passed 30 tests and the actual full-count CLI integration passed with the final controller source, as recorded in its round-3 response. Synthetic validation confers neither study evidence nor any preparation/stage/run authority. Size/controller require another scoped council round; this writer requests none.

## Recomputed per-file SHA-256

All artefact files, including historical evidence, are listed except this response itself (its self-digest is delivered with the commit head).

| File | SHA-256 |
|---|---|
| `README.md` | `8bf16f39f36f4afb6dd0986b8dc7ea8f4eaabe4e2faa46b6a4e3dfc355cb2bd7` |
| `benchmark_journal.py` | `36d692e228d5f3afbb8951bfdb12aacc6a07a257a48a8f4a2e2f234a269fdf55` |
| `controller.py` | `9fe453651816394a88272a98208f7e484c8bd73e895535638c68c273fc56478c` |
| `derive_known_answers.py` | `c789ea6b2b7d4dbef54ed70ae2ec96a02b89fa6822e3d77181fc443c280edc33` |
| `known_answers.json` | `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5` |
| `mutate_round2.py` | `2b944f4583fefbf654a9853023a22a2f02487f8c2c94b9a833239f50d1772f4c` |
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
| `reviews/round-3-test-output.txt` | `3fd9a9f8a63538f974564ccaddffa05906107fea860115bb77ade987e5732959` |
| `reviews/round-3.md` | `d8e708a3e133d7f73ae2584d80e5a874f51ffe6c28cfa26ce8b676e44010e4d6` |
| `test_controller.py` | `30c16339e173f8f925c1b7219e6587a4679b30a0addabed270573a6a516ca542` |
