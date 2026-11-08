# oc-controller: round-2 remediation response

**Status: implemented and exercised; submitted for scoped re-review, not council acceptance.** Responds to every R2 required fix and each named R1 residual in the byte-preserved `round-2.md`. All work is synthetic-only and pre-clock. No preparation action (a)–(e), real repository data, candidate fetch, bootstrap/attestation, binding implementation, identifier review, authentic sealing/inventory, scans or model calls occurred.

## Issue → fix → catching regression

| Issue | Fix | Implementation | Catching test |
|---|---|---|---|
| R2-01 / R1-04 residual | Per-population nonreentrant locks and eight fixed binding slots share a short state lock and raw meter; no one-worker global execution lock. | `controller.py:385` (`binding_population`), `controller.py:422` (`_attempt`) | `test_controller.py:132` (`test_eight_concurrent_binding_slots_equal_serial_and_refuse_reentry`) |
| R2-02 / R1-03 residual | One state lock serializes commits/snapshots/stops; gate-to-trial reauthorization; hashed explicit in-flight rule retains completed cap result but leaves gold/pause/integrity work charged/unclassified. | `controller.py:422` (`_attempt`), `controller.py:279` (`_emit`), `controller.py:683` (`_terminate`), `controller.py:671` (`_build_report`) | `test_controller.py:184` (`test_gate_window_gold_pause_and_cap_never_start_trial`), `test_controller.py:211` (`test_trial_window_gold_pause_and_cap_classification`), `test_controller.py:518` (`test_gold_cap_and_completion_races_are_single_atomic_termination`) |
| R2-03 / R1-03 residual | run owns safety pacing; public per-step safety/finish/next-group APIs removed; durable dispatch, lane, calibration/criticality and completion timestamps. | `controller.py:596` (`run`), `controller.py:530` (`_calibrate`), `controller.py:570` (`_run_criticality`), `controller.py:279` (`_emit`) | `test_controller.py:274` (`test_owned_loop_calibration_order_disclosure_and_outcome_independence`) |
| R2-04 / R1-06 residual | Controller calls signed kernel on prepared data for all five populations before attempts, retains complete disclosures and own digests, reuses first-profile results across kappa, refuses contradictions; no asserted unreachable API. | `controller.py:530` (`_calibrate`), `controller.py:525` (`_profile`), `controller.py:596` (`run`), `controller.py:659` (`_snapshot`) | `test_controller.py:303` (`test_unreachable_population_three_before_attempt_and_profile_reuse`), `test_controller.py:343` (`test_unreachable_owned_loop_advances_in_frozen_order`) |
| R2-05 | No asserted outcome resolution; reviewed resume automatically reexecutes the pending phase/attempt with original seeds under the lease. | `controller.py:510` (`resume`), `controller.py:422` (`_attempt`), `controller.py:596` (`run`) | `test_controller.py:243` (`fault_case`), `test_controller.py:463` (`test_actual_process_loss_mid_safety_replays_same_attempt_and_seeds`), `test_controller.py:211` (`test_trial_window_gold_pause_and_cap_classification`) |
| R2-06 | Unified gate/trial/calibration/criticality BaseException/malformed-result latch; separate gate-fault counter, no fabricated gate rejection; reviewed same-phase recovery. | `controller.py:411` (`_fault`), `controller.py:422` (`_attempt`), `controller.py:570` (`_run_criticality`), `controller.py:646` (`_population_report`) | `test_controller.py:268` (`test_gate_exception_is_latched`), `test_controller.py:269` (`test_nonbool_gate_is_latched_not_gate_failure`), `test_controller.py:270` (`test_trial_exception_is_latched`), `test_controller.py:271` (`test_nonbool_trial_is_latched`), `test_controller.py:272` (`test_baseexception_is_latched`), `test_controller.py:381` (`test_criticality_exception_is_latched`), `test_controller.py:382` (`test_criticality_malformed_is_latched`), `test_controller.py:383` (`test_criticality_baseexception_is_latched`), `test_controller.py:546` (`test_fault_arriving_after_reviewed_resume_latches_new_incident`) |
| R2-07 | Fixed run directory/exclusive process owner; create-once input/origin/source freeze; hash-chained fsync append-before-apply journal/replay; immutable terminated replay and recorded process-loss/torn-tail/meter integrity handling. | `controller.py:139` (`__init__`), `controller.py:279` (`_emit`), `controller.py:306` (`_apply`), `controller.py:251` (`_replay`), `controller.py:357` (`_boundary`) | `test_controller.py:385` (`test_replay_mid_attempt_and_after_gold_and_fixed_freeze`), `test_controller.py:463` (`test_actual_process_loss_mid_safety_replays_same_attempt_and_seeds`), `test_controller.py:410` (`test_process_lock_and_torn_tail_preserves_evidence`), `test_controller.py:428` (`test_meter_origin_decrease_and_fault_latch_without_meter_read`) |
| R2-08 | Private decision state/read-only properties and immutable nested manifest; completion is private and loop-owned. | `controller.py:222` (`manifest`), `controller.py:236` (`terminated`), `controller.py:234` (`paused`), `controller.py:230` (`oc_started_at`), `controller.py:665` (`_complete`) | `test_controller.py:449` (`test_private_readonly_state_and_completion_guard`) |
| R2-09 / R1-10 residual | Isolating negative guards for kernel/order/policy before freeze, completed-population criticality refusal, nonbool gate, earlier OC origin, decreasing meters and independent interior Wilson value. | `controller.py:37` (`_kernel`), `controller.py:139` (`__init__`), `controller.py:665` (`_complete`), `controller.py:357` (`_boundary`), `controller.py:85` (`wilson`) | `test_controller.py:119` (`test_source_order_policy_refusals_before_freeze`), `test_controller.py:449` (`test_private_readonly_state_and_completion_guard`), `test_controller.py:269` (`test_nonbool_gate_is_latched_not_gate_failure`), `test_controller.py:428` (`test_meter_origin_decrease_and_fault_latch_without_meter_read`), `test_controller.py:100` (`test_independent_known_answers_and_canonical_seed_refusals`) |
| R2-10 | Committed standalone no-controller/no-kernel derivation plus full OpenSSL output; published operational policy is correctly described as checksummed, not independently designed. | `derive_known_answers.py:22` (`derive`), `derive_known_answers.py:15` (`openssl`) | `test_controller.py:100` (`test_independent_known_answers_and_canonical_seed_refusals`) |
| R2-11 | Safety cannot start with any unfinished/pending binding lease; string seed cells must equal their compact sorted-key JSON bytes. | `controller.py:596` (`run`), `controller.py:62` (`seed`) | `test_controller.py:501` (`test_unfinished_binding_refuses_safety_and_grid_complete`), `test_controller.py:100` (`test_independent_known_answers_and_canonical_seed_refusals`) |

## Conservative implementation choices and independent evidence

The council's recommended design replaces the caller-paced library: fixed directory and ownership lock, append-before-apply fsync journal, shared state lock, per-lane binding locks, owned serial safety loop, signed-kernel calibration before any attempt, immutable runtime state, and reviewed same-seed recovery. No public per-step safety, finish, next_group, mark_unreachable or asserted-adequacy crash-resolution API remains.

The policy explicitly fixes the in-flight linearization/commit rule and eight binding slots. Its canonical checksum is now `71c848cdf686b5424629098f4a8bd30ea8b8ceae6ea050652e2a7f1d230dc33b`. The independently constructed order remains `04aa2184f79cb3795f60305a6b8fa31687d5acd7069846e78b6acf1b1f4dd4ff`. `derive_known_answers.py` imports neither controller nor kernel; it runs OpenSSL for all 2016 order HMACs and final digest and includes the first/last identities, four seed vectors and an independent Decimal-60 Wilson(50,2000). Its full output is `round-2-derivation-output.json` and the current `known_answers.json`. Policy hashing is described correctly as a checksum of published `policy.json`, not independent policy-design evidence.

All 9 committed scratch-only guard probes were killed by assertion failures, not import errors (`round-2-mutant-output.txt`). Actual subprocess loss during the first safety trial is exercised with `os._exit(17)`; replay preserves the first index and all seed values, requires reviewed resume, and a cap after the recovered result retains that result. Tests seed explicitly labelled feasible prior counter states for some endpoint transitions; they do not claim a full OC was executed.

Implementation-validation note: an initial controller suite timed out after 600 seconds. Its child journal recorded a calibration `ModuleNotFoundError`; the signed kernel lazily imports `ranking`. The controller now explicitly digest-pins and loads that dependency, and the corrected suite passed. A later regression also covers a fault arriving after an earlier pause was reviewed/resumed: it must latch a new incident, never reuse that earlier review. The final suite below has 25 tests. No numerical signed byte or scientific kernel changed.

## Exact commands and complete suite output

Commands ran from `/Users/adamjackson/Projects/work/scout-x1g` after `source .tools/activate`, using pinned CPython 3.11.11. For CLI reproduction select a **new nonexistent destination**; the committed create-once receipts must not be overwritten.

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/oc-controller/test_controller.py -v
```

Complete output, also `round-2-test-output.txt`:

```text
test_actual_process_loss_mid_safety_replays_same_attempt_and_seeds (__main__.ControllerTests.test_actual_process_loss_mid_safety_replays_same_attempt_and_seeds) ... ok
test_baseexception_is_latched (__main__.ControllerTests.test_baseexception_is_latched) ... ok
test_binding_completion_caps_and_gate_failures (__main__.ControllerTests.test_binding_completion_caps_and_gate_failures) ... ok
test_criticality_baseexception_is_latched (__main__.ControllerTests.test_criticality_baseexception_is_latched) ... ok
test_criticality_exception_is_latched (__main__.ControllerTests.test_criticality_exception_is_latched) ... ok
test_criticality_malformed_is_latched (__main__.ControllerTests.test_criticality_malformed_is_latched) ... ok
test_eight_concurrent_binding_slots_equal_serial_and_refuse_reentry (__main__.ControllerTests.test_eight_concurrent_binding_slots_equal_serial_and_refuse_reentry) ... ok
test_fault_arriving_after_reviewed_resume_latches_new_incident (__main__.ControllerTests.test_fault_arriving_after_reviewed_resume_latches_new_incident) ... ok
test_gate_exception_is_latched (__main__.ControllerTests.test_gate_exception_is_latched) ... ok
test_gate_window_gold_pause_and_cap_never_start_trial (__main__.ControllerTests.test_gate_window_gold_pause_and_cap_never_start_trial) ... ok
test_gold_cap_and_completion_races_are_single_atomic_termination (__main__.ControllerTests.test_gold_cap_and_completion_races_are_single_atomic_termination) ... ok
test_independent_known_answers_and_canonical_seed_refusals (__main__.ControllerTests.test_independent_known_answers_and_canonical_seed_refusals) ... ok
test_meter_origin_decrease_and_fault_latch_without_meter_read (__main__.ControllerTests.test_meter_origin_decrease_and_fault_latch_without_meter_read) ... ok
test_nonbool_gate_is_latched_not_gate_failure (__main__.ControllerTests.test_nonbool_gate_is_latched_not_gate_failure) ... ok
test_nonbool_trial_is_latched (__main__.ControllerTests.test_nonbool_trial_is_latched) ... ok
test_owned_loop_calibration_order_disclosure_and_outcome_independence (__main__.ControllerTests.test_owned_loop_calibration_order_disclosure_and_outcome_independence) ... ok
test_private_readonly_state_and_completion_guard (__main__.ControllerTests.test_private_readonly_state_and_completion_guard) ... ok
test_process_lock_and_torn_tail_preserves_evidence (__main__.ControllerTests.test_process_lock_and_torn_tail_preserves_evidence) ... ok
test_replay_mid_attempt_and_after_gold_and_fixed_freeze (__main__.ControllerTests.test_replay_mid_attempt_and_after_gold_and_fixed_freeze) ... ok
test_source_order_policy_refusals_before_freeze (__main__.ControllerTests.test_source_order_policy_refusals_before_freeze) ... ok
test_trial_exception_is_latched (__main__.ControllerTests.test_trial_exception_is_latched) ... ok
test_trial_window_gold_pause_and_cap_classification (__main__.ControllerTests.test_trial_window_gold_pause_and_cap_classification) ... ok
test_unfinished_binding_refuses_safety_and_grid_complete (__main__.ControllerTests.test_unfinished_binding_refuses_safety_and_grid_complete) ... ok
test_unreachable_owned_loop_advances_in_frozen_order (__main__.ControllerTests.test_unreachable_owned_loop_advances_in_frozen_order) ... ok
test_unreachable_population_three_before_attempt_and_profile_reuse (__main__.ControllerTests.test_unreachable_population_three_before_attempt_and_profile_reuse) ... ok

----------------------------------------------------------------------
Ran 25 tests in 2.262s

OK
```

### Reproducible scratch-only guard mutants

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/oc-controller/mutate_round2.py
```

The committed runner contains exact one-occurrence substitutions and isolated copied-fixture layouts. Full command-by-command output is `round-2-mutant-output.txt`; it is included in the digest table below.

### Independent full-order derivation

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/oc-controller/derive_known_answers.py
```

Full output is `round-2-derivation-output.json`, byte-identical to the current known-answer fixture and included in the digest table.

## Frozen rev9 bytes

The following check passed with manifest SHA `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`:

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py
```

**126/131 literal file matches, 5/5 shared-file owned-extract groups valid, 131/131 unchanged against baseline e2b94e3c.** Nineteen of 20 owned extracts match literally; the only state-only normalization is the already-permitted HD9 transition under round-9 condition 2. Full JSON is `final-evaluator/reviews/round-2-frozen-output.json` under stage3-drafts. No new change to any signed file occurred.

## Current per-file SHA-256s

Paths below are relative to this artefact folder. Historical review/evidence records are included and untouched. This response excludes its own hash to avoid circularity; delivery reports that hash separately.

| File | SHA-256 |
|---|---|
| `README.md` | `db902fcc72ac24f41cdbfa40b17a8be57091b52285afdbd4ff9a8073f5299a8c` |
| `controller.py` | `331770134461efeb09828bbc6e906251b3e5082ab36b8758e915f11e6f0061a9` |
| `test_controller.py` | `e029d1ca0375b0388692b341f310a9c5dbb9cfaa9ecc0fff4b66473c125ba7cf` |
| `known_answers.json` | `012699629663f464af546f8e5048e730af711b49f9932e3e4b6db115a5a24449` |
| `reviews/round-1.md` | `f5a1c845f81d3d22e9d357c03e7c233c382317d8894e67551a10a3558a90539b` |
| `reviews/round-1-response.md` | `682af9a486173b56c3a302e93251134f2592a3a0a9100ac174c4d6c8296b8ce8` |
| `reviews/round-1-test-output.txt` | `4d970a1f1c77544377c8dff41a23a6a3efae447b0ec28b8e324d467eb534f91a` |
| `reviews/round-1-cap-output.txt` | `1d442d16d51c04f4120587d56c4638fa9035e88280376da4fccd70cb690cfeb4` |
| `policy.json` | `decd2f3aa55fa61e8a8e5cf2a4abc7de33403a480f959fd8fa832030cfd0f918` |
| `derive_known_answers.py` | `cf86e79f65565388ed029cc5184d71e7874750f1d33f30ac8635cafe8d2e1b8b` |
| `mutate_round2.py` | `7732b41ca846c0eeedde88b91e0714da1f7841963d6a343f205b80e3a3fe9c66` |
| `reviews/round-2.md` | `d3b770d948fa2b7162f57a281b27a9201c3283e3f22f94436c40a4499945e22e` |
| `reviews/round-2-derivation-output.json` | `012699629663f464af546f8e5048e730af711b49f9932e3e4b6db115a5a24449` |
| `reviews/round-2-test-output.txt` | `5b25a2d02a93e746da046c9233da84c2df22295243279f8eb181151962db79dd` |
| `reviews/round-2-mutant-output.txt` | `48d9c17a3e41adfd12186ecc2559d842ad54d4dae778383afe26d5884fe58467` |
