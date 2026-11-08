# locked-size-input — round-1 response

**Status: fixes submitted for scoped re-review, not self-closure or sign-off.** This response addresses every issue in `round-1.md`; the review record is preserved byte-for-byte. All work and validation are synthetic-only. No preparation action (a)–(e), candidate fetch, scanner/model call, real-data inventory, attestation or binding OC was performed. No signed revision-9 byte was changed.

## Required fixes and catching regressions

| Issue | Fix | Implementation | Catching regression |
|---|---|---|---|
| R1-01 | Runnable locked benchmark/calibration/gate/fit modes; only guarded size-site substitutions; complete L_r in construction, calibration and both trial fixtures; unchanged enumeration, replicates and scientific equality. | `locked_commands.py:82`, `locked_commands.py:109`, `locked_commands.py:121`, `locked_commands.py:153`, `locked_commands.py:179`, `locked_commands.py:186` | `test_size_input.py:211` `test_construction_calibration_consume_every_locked_size`, `test_size_input.py:232` `test_trial_population_lengths_and_selected_ids`, `test_size_input.py:252` `test_actual_helper_one_two_worker_science`, `test_size_input.py:271` `test_size_site_drift_refused`, `test_size_input.py:276` `test_fit_uses_locked_maxima_not_slopes` |
| R1-02 | Fixed alias/repository/tip mapping and oldest-first SHA/position/size records, inventory seal and schema are all digest-bound; invalid permutations refused; independent record digest pinned. | `size_input.py:54` | `test_size_input.py:97` `test_identity_digest_and_invalid_order` |
| R1-03 | Exactly 100000 manifest-HMAC keyed SRS20 draws using controller.seed; fixed literal cell, exact integer cap, Wilson reporting, zero-success refusal; obsolete draw API removed. | `size_input.py:26`, `size_input.py:111`, `size_input.py:117`, `size_input.py:130` | `test_size_input.py:149` `test_seed_and_heterogeneous_selection_known_answers`, `test_size_input.py:178` `test_full_100000_key_estimates`, `test_size_input.py:170` `test_gate_boundaries` |
| R1-04 | Signed calibrator on unequal-length heterogeneous vectors, middle-repository global extrema, both size factors and effect/null; independent offsets and non-strippable truth mismatch refusal. | `population.py:15`, `size_input.py:98` | `test_size_input.py:40` `test_asymmetric_offsets_and_constant_zero`, `test_size_input.py:53` `test_planted_effect_null_and_truth_identity`, `test_size_input.py:64` `test_mismatched_truth_mutant_raises` |
| R1-05 | Inclusive run-count and run-size bounds, exact types, complete mapping and global extrema tested. | `size_input.py:54` | `test_size_input.py:71` `test_constructor_run_count_and_size_boundaries`, `test_size_input.py:97` `test_identity_digest_and_invalid_order`, `test_size_input.py:40` `test_asymmetric_offsets_and_constant_zero` |
| R1-06 | Frozen input object, tuple vectors, mapping proxy and independent canonical-byte snapshot refuse mutation. | `size_input.py:42`, `size_input.py:54` | `test_size_input.py:135` `test_immutable_snapshot` |
| R1-07 | README now describes implemented size portions precisely, the fixed mapping, executed validation and unexecuted full timing/RSS commands. | `README.md:1` | `test_size_input.py:252` `test_actual_helper_one_two_worker_science` |

## Conservative choices and independent evidence

The fixed pre-inventory mapping is the council's recommended mapping: slots 40/400 retain seed/identity labels, but each builds all complete vectors; trial eligible counts are each locked L_r; selected-frame content/8 cells/2000 trials/B=1999/one-two equality/worst-case enumeration are unchanged. Integer gate-cap arithmetic and immutable ordered identity records are frozen now, not chosen after inventory. Full CLI timing/RSS commands were not run and no authentic fit is claimed. The synthetic suite exercised 32000 helper trials. The first one-trial regression fixture was rejected by the signed helper and corrected to its full minimum; no signed helper or scientific algorithm changed.

Independent OpenSSL output is in `round-1-known-answer-output.txt`. It reproduces attempts 0 and 99999 using the exact literal gate cell and an independently constructed complete input-record digest (no SizeInput import). README lines 61–70 state the stdin/record recipe and expected values.

## Exercised command and complete output

From `/Users/adamjackson/Projects/work/scout-x1g`:

```sh
source .tools/activate
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/test_size_input.py -v
```

Observed: **16 tests passed** under pinned CPython 3.11.11. This is deterministic implementation validation, not evidence of an HD9 effect or authentic stage authority. Verbatim output (also retained in `round-1-test-output.txt`):

```text
test_actual_helper_one_two_worker_science (__main__.LockedModeTests.test_actual_helper_one_two_worker_science) ... ok
test_construction_calibration_consume_every_locked_size (__main__.LockedModeTests.test_construction_calibration_consume_every_locked_size) ... ok
test_fit_uses_locked_maxima_not_slopes (__main__.LockedModeTests.test_fit_uses_locked_maxima_not_slopes) ... ok
test_size_site_drift_refused (__main__.LockedModeTests.test_size_site_drift_refused) ... ok
test_trial_population_lengths_and_selected_ids (__main__.LockedModeTests.test_trial_population_lengths_and_selected_ids) ... ok
test_asymmetric_offsets_and_constant_zero (__main__.SizeTests.test_asymmetric_offsets_and_constant_zero) ... ok
test_constructor_run_count_and_size_boundaries (__main__.SizeTests.test_constructor_run_count_and_size_boundaries) ... ok
test_equal_vectors_preserve_exact_population (__main__.SizeTests.test_equal_vectors_preserve_exact_population) ... ok
test_full_100000_key_estimates (__main__.SizeTests.test_full_100000_key_estimates) ... ok
test_gate_boundaries (__main__.SizeTests.test_gate_boundaries) ... ok
test_identity_digest_and_invalid_order (__main__.SizeTests.test_identity_digest_and_invalid_order) ... ok
test_immutable_snapshot (__main__.SizeTests.test_immutable_snapshot) ... ok
test_mismatched_truth_mutant_raises (__main__.SizeTests.test_mismatched_truth_mutant_raises) ... ok
test_only_size_source_changed_in_kernel (__main__.SizeTests.test_only_size_source_changed_in_kernel) ... ok
test_planted_effect_null_and_truth_identity (__main__.SizeTests.test_planted_effect_null_and_truth_identity) ... ok
test_seed_and_heterogeneous_selection_known_answers (__main__.SizeTests.test_seed_and_heterogeneous_selection_known_answers) ... ok

----------------------------------------------------------------------
Ran 16 tests in 586.708s

OK
```

## Frozen revision-9 check

`PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py` passed: **126/131 literal file matches; 5/5 shared-file owned-extract groups valid; 131/131 files unchanged against baseline e2b94e3c**. Nineteen of 20 owned extracts match literally. HD9 differs only by the already-permitted round-9 condition-2 state transition; normalising that state alone reproduces its signed hash. The manifest remains `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`. Full machine-readable output is `evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/reviews/round-1-frozen-output.json`.

## Current file SHA-256s

Paths are relative to this artefact folder. The response excludes itself to avoid a circular digest; the delivery records its digest separately. The table includes the untouched incoming council record and all exercised evidence files.

| File | SHA-256 |
|---|---|
| `README.md` | `9b0c44164f2efd1fa514d8b4bfaf6a1fd4939b000d076dbfd65d785b4bd64776` |
| `size_input.py` | `3b315fdf8c4b055fcb83ba3a5597e5c83f67605354d5966e89505ca228399d8f` |
| `population.py` | `a32e33cd8061940067d94b56cead9a85608a3a07c79b8a9b89c248482586f0c6` |
| `locked_commands.py` | `687937e1a4bb89cdd588b7bd25fa4a5f271f6c3a7d9ec9956672decb853a6f58` |
| `test_size_input.py` | `9a33087e3562327c014b7ce3a77af5b050940b6b5a2ec20d462246af0ea1afc6` |
| `reviews/round-1.md` | `0af23777d7a4c8bc81e152c95581084ad3bcf194a759460d79d25a05898b9fc8` |
| `reviews/round-1-test-output.txt` | `bfccce29e41cee76006bfec9e86e025d04f8cdc5adf07fa502a13301f392b25f` |
| `reviews/round-1-known-answer-output.txt` | `ec746d03798e5c5d709b1fffa2808c97838d468eabc6169c4c8320285a252a95` |
