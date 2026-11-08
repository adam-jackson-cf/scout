# oc-controller — round-1 response

**Status: fixes submitted for scoped re-review, not self-closure or sign-off.** This response addresses every issue in `round-1.md`; the review record is preserved byte-for-byte. All work and validation are synthetic-only. No preparation action (a)–(e), candidate fetch, scanner/model call, real-data inventory, attestation or binding OC was performed. No signed revision-9 byte was changed.

## Required fixes and catching regressions

| Issue | Fix | Implementation | Catching regression |
|---|---|---|---|
| R1-01 | Signed population_identity/criticality_key imported from digest-verified source; seven exact dimensions; UTF-8 identity bytes used for order/seeds; ordered JSON string-array SHA pinned independently with OpenSSL. | `controller.py:35`, `controller.py:52`, `controller.py:62`, `controller.py:352` | `test_controller.py:54` `test_signed_identity_exact_dimensions_and_independent_order_pin`, `test_controller.py:82` `test_seed_known_answers_openssl_and_input_refusals`, `test_controller.py:197` `test_criticality_exact_keys_order_and_auto_completion` |
| R1-02 | Verified signed manifest builds full 2016 identities internally, checks pinned order/policy, refuses mutation and writes create-once digest/source/timestamp freeze before dispatch. | `controller.py:176`, `controller.py:83`, `controller.py:223` | `test_controller.py:54` `test_signed_identity_exact_dimensions_and_independent_order_pin`, `test_controller.py:95` `test_manifest_substitution_and_create_once_freeze` |
| R1-03 | Every execution lease checks schedule state/meters. Gold/caps snapshot irreversibly; post-stop results are discarded; grid completion and integrity termination are logged. | `controller.py:120`, `controller.py:227`, `controller.py:408`, `controller.py:422`, `controller.py:430` | `test_controller.py:219` `test_gold_before_during_and_after_last_trial_is_irreversible`, `test_controller.py:239` `test_caps_at_execution_boundary_and_wall_clock_unpaused`, `test_controller.py:258` `test_cap_during_last_trial_and_criticality_freezes_partial_counts`, `test_controller.py:109` `test_serial_lease_refuses_reentry_and_gold_inside_trial_discards_result`, `test_controller.py:332` `test_unreachable_provenance_advances_in_order_and_grid_completion_logged` |
| R1-04 | Hashed operational policy fixes raw actual meters, unpaused first-full-OC wall clock, >= caps, evidence-only pauses/resume, serial lane/slot mapping and nonreentrant execution. | `controller.py:227`, `controller.py:286`, `controller.py:274`, `controller.py:311`, `controller.py:333` | `test_controller.py:179` `test_population_lane_mapping_and_finish_membership`, `test_controller.py:239` `test_caps_at_execution_boundary_and_wall_clock_unpaused`, `test_controller.py:284` `test_integrity_pause_resume_preserves_attempt_and_group`, `test_controller.py:353` `test_outcome_independent_dispatch_events_statuses_and_integrity_stop`, `test_controller.py:109` `test_serial_lease_refuses_reentry_and_gold_inside_trial_discards_result` |
| R1-05 | State-owned active/completed identity snapshots retain counters/Wilson/criticality/calibration; exhaustion only belongs to the active group; report requires termination and is immutable. | `controller.py:157`, `controller.py:396`, `controller.py:408`, `controller.py:430` | `test_controller.py:319` `test_exhaustion_is_bound_only_to_active_group`, `test_controller.py:219` `test_gold_before_during_and_after_last_trial_is_irreversible`, `test_controller.py:197` `test_criticality_exact_keys_order_and_auto_completion` |
| R1-06 | Explicit unreachable terminal state is bound to group, target, signed calibration source and expected record digest; zero counters retained; frozen order continues. | `controller.py:384`, `controller.py:402` | `test_controller.py:332` `test_unreachable_provenance_advances_in_order_and_grid_completion_logged` |
| R1-07 | Identity/index execution leases and exact same-group population membership; all four signed criticality keys are retained before automatic completion; duplicate/foreign/incomplete records refused. | `controller.py:92`, `controller.py:254`, `controller.py:370`, `controller.py:352` | `test_controller.py:179` `test_population_lane_mapping_and_finish_membership`, `test_controller.py:197` `test_criticality_exact_keys_order_and_auto_completion` |
| R1-08 | Counters are init=False; crashes/non-bool adequacy remain charged and classified, latch integrity pause and prevent completion/resume until recorded reviewed reconciliation. | `controller.py:92`, `controller.py:129`, `controller.py:319`, `controller.py:333` | `test_controller.py:302` `test_crash_and_nonbool_latch_integrity_with_reconciled_counts`, `test_controller.py:109` `test_serial_lease_refuses_reentry_and_gold_inside_trial_discards_result` |
| R1-09 | Exact manifest phases, strict nonnegative integer indices, population-phase attempt zero, NUL/UTF-8/first64-big-endian vectors independently checked by OpenSSL. | `controller.py:52` | `test_controller.py:82` `test_seed_known_answers_openssl_and_input_refusals` |
| R1-10 | All named counter/probability/meter/stop/mutation/Wilson/outcome-independence boundaries added; README citations and independent-evidence claims corrected. | `controller.py:113`, `controller.py:117`, `controller.py:71`, `README.md:1` | `test_controller.py:151` `test_completion_on_last_attempt_and_execution_wins`, `test_controller.py:364` `test_probability_meter_and_wilson_boundaries`, `test_controller.py:170` `test_safety_stops_at_executions_including_inadequacy`, `test_controller.py:95` `test_manifest_substitution_and_create_once_freeze`, `test_controller.py:353` `test_outcome_independent_dispatch_events_statuses_and_integrity_stop` |

## Conservative choices and independent evidence

Conservative operational choices are explicit in the pinned policy: eight reserved slots, serial increasing-index population lanes, serial j phases, raw actual OC CPU without x2, and wall time from first full-OC start without pause/reset. Integrity pauses require recorded evidence; resume requires reviewed resolution. No scheduling API accepts safety outcomes. The opposite-outcome metamorphic regression produced identical dispatch, event and status sequences.

`known_answers.json` contains independent OpenSSL vectors. For the full ordered list, a separate construction enumerated the seven dimensions from the signed manifest, encoded `[kind, dimensions]` without importing controller code, used `openssl dgst -sha256 -hmac 169-x1-safety-order` for each identity, sorted digests/identity bytes, and passed the compact JSON array of strings to `openssl dgst -sha256`. Ordered-list SHA is `04aa2184f79cb3795f60305a6b8fa31687d5acd7069846e78b6acf1b1f4dd4ff`; policy SHA is `e2c8c9ec1da43b4eeface3cf3c52c711ecad8cef11be97c00defe51fc12353ca`. The tests independently invoke OpenSSL for the first order HMAC and all four seed vectors.

A pinned-Python exact-arithmetic enumeration checked all 100000 possible positive gate-success counts against `ceil(4000/(s/100000))`; it observed zero mismatches (`round-1-cap-output.txt`). Reproduction: `.tools/venv/bin/python -c 'import math; bad=[s for s in range(1,100001) if math.ceil(4000/(s/100000)) != (4000*100000+s-1)//s]; print(len(bad)); assert not bad'`. The initial hand-entered Wilson endpoint fixture was corrected using independent 60-digit Decimal evaluation of z²/(2000+z²), not by changing the Wilson implementation.

## Exercised command and complete output

From `/Users/adamjackson/Projects/work/scout-x1g`:

```sh
source .tools/activate
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/oc-controller/test_controller.py -v
```

Observed: **19 tests passed** under pinned CPython 3.11.11. This is deterministic implementation validation, not evidence of an HD9 effect or authentic stage authority. Verbatim output (also retained in `round-1-test-output.txt`):

```text
test_cap_during_last_trial_and_criticality_freezes_partial_counts (__main__.ControllerTests.test_cap_during_last_trial_and_criticality_freezes_partial_counts) ... ok
test_caps_at_execution_boundary_and_wall_clock_unpaused (__main__.ControllerTests.test_caps_at_execution_boundary_and_wall_clock_unpaused) ... ok
test_completion_on_last_attempt_and_execution_wins (__main__.ControllerTests.test_completion_on_last_attempt_and_execution_wins) ... ok
test_crash_and_nonbool_latch_integrity_with_reconciled_counts (__main__.ControllerTests.test_crash_and_nonbool_latch_integrity_with_reconciled_counts) ... ok
test_criticality_exact_keys_order_and_auto_completion (__main__.ControllerTests.test_criticality_exact_keys_order_and_auto_completion) ... ok
test_execution_cap_is_independent_inadequacy_is_charged (__main__.ControllerTests.test_execution_cap_is_independent_inadequacy_is_charged) ... ok
test_exhaustion_is_bound_only_to_active_group (__main__.ControllerTests.test_exhaustion_is_bound_only_to_active_group) ... ok
test_gate_rejections_never_execute_and_attempt_cap (__main__.ControllerTests.test_gate_rejections_never_execute_and_attempt_cap) ... ok
test_gold_before_during_and_after_last_trial_is_irreversible (__main__.ControllerTests.test_gold_before_during_and_after_last_trial_is_irreversible) ... ok
test_integrity_pause_resume_preserves_attempt_and_group (__main__.ControllerTests.test_integrity_pause_resume_preserves_attempt_and_group) ... ok
test_manifest_substitution_and_create_once_freeze (__main__.ControllerTests.test_manifest_substitution_and_create_once_freeze) ... ok
test_outcome_independent_dispatch_events_statuses_and_integrity_stop (__main__.ControllerTests.test_outcome_independent_dispatch_events_statuses_and_integrity_stop) ... ok
test_population_lane_mapping_and_finish_membership (__main__.ControllerTests.test_population_lane_mapping_and_finish_membership) ... ok
test_probability_meter_and_wilson_boundaries (__main__.ControllerTests.test_probability_meter_and_wilson_boundaries) ... ok
test_safety_stops_at_executions_including_inadequacy (__main__.ControllerTests.test_safety_stops_at_executions_including_inadequacy) ... ok
test_seed_known_answers_openssl_and_input_refusals (__main__.ControllerTests.test_seed_known_answers_openssl_and_input_refusals) ... ok
test_serial_lease_refuses_reentry_and_gold_inside_trial_discards_result (__main__.ControllerTests.test_serial_lease_refuses_reentry_and_gold_inside_trial_discards_result) ... ok
test_signed_identity_exact_dimensions_and_independent_order_pin (__main__.ControllerTests.test_signed_identity_exact_dimensions_and_independent_order_pin) ... ok
test_unreachable_provenance_advances_in_order_and_grid_completion_logged (__main__.ControllerTests.test_unreachable_provenance_advances_in_order_and_grid_completion_logged) ... ok

----------------------------------------------------------------------
Ran 19 tests in 0.795s

OK
```

## Frozen revision-9 check

`PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py` passed: **126/131 literal file matches; 5/5 shared-file owned-extract groups valid; 131/131 files unchanged against baseline e2b94e3c**. Nineteen of 20 owned extracts match literally. HD9 differs only by the already-permitted round-9 condition-2 state transition; normalising that state alone reproduces its signed hash. The manifest remains `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`. Full machine-readable output is `evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/reviews/round-1-frozen-output.json`.

## Current file SHA-256s

Paths are relative to this artefact folder. The response excludes itself to avoid a circular digest; the delivery records its digest separately. The table includes the untouched incoming council record and all exercised evidence files.

| File | SHA-256 |
|---|---|
| `README.md` | `583967f896c4ee0b602769ff9986173e125003b53966605f5180afa9bd1897b4` |
| `controller.py` | `e2740ddc4e7d7bb5ffe5a826c292a1b87dd9ad74c97463aaf13b1f81491c94bb` |
| `test_controller.py` | `41419d66799b525b82c0f857fd0dede4f813a21351cb63744ea4e16b3a405f5e` |
| `known_answers.json` | `24b0b48497cac0955737bef05661b109462223304eac068898151a13cd15a694` |
| `reviews/round-1.md` | `f5a1c845f81d3d22e9d357c03e7c233c382317d8894e67551a10a3558a90539b` |
| `reviews/round-1-test-output.txt` | `4d970a1f1c77544377c8dff41a23a6a3efae447b0ec28b8e324d467eb534f91a` |
| `reviews/round-1-cap-output.txt` | `1d442d16d51c04f4120587d56c4638fa9035e88280376da4fccd70cb690cfeb4` |
