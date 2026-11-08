# X1 stage-3 draft review — locked-size input change (`stage3-drafts/locked-size-input/`) — round 2

Date: 2026-11-08. Reviewer: innovation council, scoped round 2. Two read-only peer audits ran independently, and neither saw the lead's leaning: `innovation-challenger` (X1GSizeReview2.LsiR2Challenger) and `experiment-peer` (X1GSizeReview2.LsiR2Peer). The lead wrote the synthesis.

The basis is X1 round 9, which recommends option G (`reviews/round-9.md` l.244–247). This review covers the locked-size input change only as a stage-3 artefact. It grants no authority for preparation, inventory, locked remeasurement, OC or any action on real data.

## Reviewed files

Worktree `/Users/adamjackson/Projects/work/scout-x1g`, PR #21, head `ed7f21ce2bc58f03c4dee6e0eeaa7092ab9dbc4f`. Paths are relative to `evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/`. Main recomputed these SHA-256 values with `shasum -a 256` at that head.

| File | SHA-256 |
|---|---|
| `README.md` | `9b0c44164f2efd1fa514d8b4bfaf6a1fd4939b000d076dbfd65d785b4bd64776` |
| `locked_commands.py` | `687937e1a4bb89cdd588b7bd25fa4a5f271f6c3a7d9ec9956672decb853a6f58` |
| `population.py` | `a32e33cd8061940067d94b56cead9a85608a3a07c79b8a9b89c248482586f0c6` (same as round 1) |
| `size_input.py` | `3b315fdf8c4b055fcb83ba3a5597e5c83f67605354d5966e89505ca228399d8f` |
| `test_size_input.py` | `9a33087e3562327c014b7ce3a77af5b050940b6b5a2ec20d462246af0ea1afc6` |
| `reviews/round-1.md` | `0af23777d7a4c8bc81e152c95581084ad3bcf194a759460d79d25a05898b9fc8` |
| `reviews/round-1-response.md` | `00cce4534953a9bbafd3b1f2d3bb83ff73cdf80bb08dd97cccc66fd1beaf3986` |
| `reviews/round-1-test-output.txt` | `bfccce29e41cee76006bfec9e86e025d04f8cdc5adf07fa502a13301f392b25f` |
| `reviews/round-1-known-answer-output.txt` | `ec746d03798e5c5d709b1fffa2808c97838d468eabc6169c4c8320285a252a95` |

## Signed basis (unchanged)

- `PREREGISTRATION.md` (P): `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`
- `README.md`: `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3`
- `oc-manifest.json` (M): `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d`
- `reviews/rev9-digests.json`: `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`
- `reviews/round-9.md`: `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`

Signed text relied on:
- **P:** l.37, l.161, l.163, l.169, l.177, l.179, l.181, l.185 and l.228.
- **M:** l.10 (`numerical_runtime`), l.9 (`primary_interval`) and l.41–63.
- **`round-9.md`:** l.326–329 (R9-reading-1).
- **Signed executables:**
  - `validation/compute_benchmark.py`: l.53–212 and l.219–283;
  - `validation/reproduce_reports.py`: l.76–123 and l.188–261.
- **Sibling drafts, read for contract only:**
  - `oc-controller/controller.py` l.52–59 and l.71–81;
  - `final-evaluator/provenance.py` l.7–9 and l.52–93;
  - `final-evaluator/evaluator.py` l.58–88;
  - `final-evaluator/README.md` l.12–13.

## Round-1 issue disposition

| ID | Fixed | Evidence (file:line + test) |
|---|---|---|
| R1-01 | **Yes** for the required size-source mapping. The new driver has defects of its own, recorded as R2-01 to R2-04 and R2-07. | **Substitutions:** `locked_commands.py:47–65` (`SIZE_SITES`) and `:69–95`. Each is an exact AST match and must occur exactly once, otherwise the driver raises "signed size site drift".<br>**Fresh process per slot:** `:136–140`. **Trial cells and worker equality:** `:153–183`. **Closure and every L_r:** `:98–106`. **Enumeration unchanged:** `:178–183`. **Maxima-based fit:** `:186–237`. **`measurement_sizes` removed.**<br>**Every size site is covered.** The lead checked this against the signed bodies, and both peers agree. The remaining uses of `L` and `frame_runs` are only seed or identity labels. `sizes` becomes the `SizeInput` object, so any site left unsubstituted raises a TypeError (fail-closed).<br>**Tests:** `test_construction_calibration_consume_every_locked_size` (:211), `test_trial_population_lengths_and_selected_ids` (:232), `test_actual_helper_one_two_worker_science` (:252; 8 cells × 2,000 trials, B = 1,999, one versus two workers), `test_size_site_drift_refused` (:271), `test_fit_uses_locked_maxima_not_slopes` (:276). |
| R1-02 | **Yes** | **Fixed mapping, tips, schema and seal:** `size_input.py:13–16` and `:54–93`. The tips match P l.33–35, and position order follows P l.37.<br>**Test:** `test_identity_digest_and_invalid_order` (:97). A change to size, SHA or seal changes the digest. Swapped runs or swapped repositories change the hash and are refused. Wrong tip, mapping, schema or `synthetic_only` values are refused. The pinned known answer `7f97e552…` matches `round-1-known-answer-output.txt` l.3. |
| R1-03 | **Yes** against its required test list. Interior coverage is still missing (R2-05). | **Controller seed loaded by file:** `size_input.py:26–34`. **Literal cell, 100,000 draws and integer cap:** `:111–127`.<br>**Tests:**<br>• `test_seed_and_heterogeneous_selection_known_answers` (:149): OpenSSL HMAC vectors for attempts 0 and 99,999, and L_r = 23/27/31 selections;<br>• `test_full_100000_key_estimates` (:178): attempts exactly 0…99,999, planted p = 1 and p = 0, "not testable";<br>• `test_gate_boundaries` (:170): 999/1,000 and 60,000/60,001, also exercised through `gate_selection`. |
| R1-04 | **Yes** | **Tests:**<br>• `test_planted_effect_null_and_truth_identity` (:53): unequal L_r of 22/23/24, both global extremes (20 and 2,000) in the middle repository, size factors .05/.20 × targets .20/0, reachability, the P l.169 identity;<br>• `test_mismatched_truth_mutant_raises` (:64);<br>• `test_asymmetric_offsets_and_constant_zero` (:40): offsets computed independently as `factor*(n-20)/1980`, and all zero for constant sizes. |
| R1-05 | **Partly.** The residual is R2-06. | **Covered:** `test_constructor_run_count_and_size_boundaries` (:71) tests 19/20/2,000/2,001 runs, sizes 19/20/2,000/2,001, `True` and `50.0`. Middle-repository extremes are covered at :40.<br>**Not covered:** the extra-key and missing-key cases (:88–95) are refused by the fixture builder (`size_input.py:136–137`) before `SizeInput` runs, so they do not test the constructor. |
| R1-06 | **Yes** | **Implementation:** `size_input.py:41–42` and `:88–93` (frozen dataclass, `MappingProxyType` over tuples, snapshot of the record bytes).<br>**Test:** `test_immutable_snapshot` (:135). |
| R1-07 | **Yes** | The README scope is qualified (`README.md:3–11`). The exact command and output are recorded (`round-1-response.md` l.28–62; `round-1-test-output.txt`). Full timing and RSS CLI runs are correctly disclaimed. Main reproduced 16/16. |

## New issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R2-01 | **major** | `locked_commands.py:143–150` (`host`), `:264–265` and `:285–287`; gate output `size_input.py:122–127`. Contract: `final-evaluator/provenance.py:52–56` and `:59–72`, `final-evaluator/README.md:12–13` | **The single locked run's outputs lack provenance the signed text requires, and the reviewed evaluator cannot consume them.**<br>• **Load provenance.** P l.181 says "Record load averages and concurrent-process count in provenance". No output records either.<br>• **Interpreter provenance.** M l.10 requires "executable path/digest and sys.flags.optimize=0 in every new output". The outputs omit `python_optimize` and the executable path. The signed benchmark records both (`compute_benchmark.py` l.258–262).<br>• **Host schema.** The host block emits `logical_cpu_count` plus extra keys. The evaluator's `validate_host` requires exactly `hw.model`, `hw.memsize`, `logical_cpus` and `os_build`.<br>• **No consumable records.** No `measurement_runs` record (with `binding_inputs`, `completed` and `integrity_failure`) is emitted. The gate output uses `draws`/`successes`, but the evaluator's `locked_inventory` uses `gate_redraws`/`gate_successes`.<br>• **Consequence.** P l.181 allows one create-once run, and load data cannot be reconstructed afterwards. A real run made with this command would therefore be refused by `maximum_completed_inputs` ("missing load provenance"), leaving the fit unmeasured, which means not testable.<br>• **Fit evidence hashes.** `fit` parses the evidence at :280 but hashes a second read at :281–282, so the recorded digest need not describe the bytes that were fitted (P l.228). | **Recommended:**<br>1. In every gate, benchmark and calibration output, record `os.getloadavg()` and a concurrent-process count (lines of `ps -A -o pid=`) at the start and at the end. Put them in a `host_load` block kept separate from the stable host identity, so host equality is not broken. Also record `python_executable` and `python_optimize`.<br>2. Emit `host` in the evaluator's four-key schema.<br>3. Emit the evaluator's `measurement_runs` run record, with the 10 `INPUTS` taken from the existing per-unit fields, and the gate's `locked_inventory` fields, both bound to `size_input_sha256`.<br>4. Read each evidence file once, then hash and parse those same bytes.<br>**Tests:**<br>• a round trip of synthetic outputs through `provenance.maximum_completed_inputs` and `evaluator.resource_fit`;<br>• one recorded synthetic end-to-end run of the README command sequence. |
| R2-02 | **major** | `locked_commands.py:283–284`; `:188–190` | **A completed run that then fails its closure check is discarded.**<br>• After every measurement has finished, a source or input mismatch raises "output not published", and nothing is written.<br>• R9-reading-1 (`round-9.md` l.326–329): "A run that reached its end is a completed run, even if its closure check then fails. Its values therefore enter the maximum, and the closure mismatch also permits one rerun."<br>• P l.181: "preserve that evidence … each binding input is the maximum over all completed runs".<br>• Discarding the run lets the rerun become the only evidence, which can lower the binding maxima. That weakens the safeguard round 9 recorded "so that no operator choice arises later". | On a mismatch after measurement:<br>1. Publish the full result create-once, with `completed: true`, `integrity_failure: "closure mismatch"`, both the start and end closures and the measurements.<br>2. Then exit non-zero.<br>The R2-01 run record must carry this classification. **Test:** change a closure source between measurement and publication, then assert that the file exists with that classification and its values. |
| R2-03 | minor | `locked_commands.py:251` and `:193–194` | **The locked benchmark accepts `--method percentile`, and nothing downstream refuses it.** P l.181 fixes the invocation as "`--method studentized --trials 2000 --cores 8` with its cell set". M l.9 makes studentized the primary interval. Percentile evidence would feed the fit unchecked. | 1. Remove the `--method` option: locked `benchmark` mode always runs studentized.<br>2. When building the R2-01 run record, refuse any evidence whose method is not `studentized`, or whose cell set is not the frozen one. The frozen set is slots {40, 400} × missing {false, true} × two fixtures, workers [1, 2] and `measured_trials == trials`.<br>**Test:** both refusals. |
| R2-04 | minor | `locked_commands.py:186–237`; `test_size_input.py:276–303` | **`locked_fit` is a second, untested implementation of the P l.181 binding stop rule.**<br>• P l.177 assigns the resource decision to the final evaluator, whose `resource_fit` is boundary-tested.<br>• `locked_fit`'s arithmetic matches signed `projection` term by term (lead check, confirmed by both peers).<br>• However, its test asserts only 3 of the 6 binding terms, at about 182 core-hours, far below the cap. These mutants survive **[INFERENCE: not executed]**:<br>  – the ×2 overhead dropped;<br>  – the 2,880 threshold or the 15-day check changed;<br>  – the 4,000 trial term or the requested/gate terms changed;<br>  – `min` used in place of `max`. | Delete `locked_fit`'s stop decision (`fits_caps`, `status`, binding hours). `fit` then only emits the R2-01 run record, so the evaluator is the single fit convention. Update README item 6 and remove the test's stop-rule assertions. |
| R2-05 | minor | `size_input.py:119–124`; `test_size_input.py:178–200` | **The estimator is checked only at p̂ ∈ {0, 1}.** At those points the Wilson term `p(1−p)` is zero, and the cap is an exact integer.<br>• A variance-term mutant and a floor-for-ceiling mutant would both survive **[INFERENCE]**.<br>• The Wilson code duplicates `controller.wilson` (`controller.py:71–81`), which is the same formula in a second copy. | 1. Call the controller's `wilson(successes, DRAWS)` through the existing file import.<br>2. Add an interior fixture, for example `gate_selection` patched to give 3 successes, which must yield cap 133,333,334. Check the Wilson bounds against an independently computed value. |
| R2-06 | minor | R1-05 residual. `test_size_input.py:88–95`; `size_input.py:55–56`, `:62–63`, `:66–67`, `:75–76`, `:78–79` | **Constructor guards added by the R1-02 fix are never reached by a test:**<br>• the repository count — if it is deleted, `zip` truncates a 4-record list and accepts it **[INFERENCE]**;<br>• the exact field sets for the record, each repository and each run;<br>• the `first_parent_position` < 2,000 bound (P l.37). | Mutate the record dicts directly: a fourth repository, a dropped repository, an extra key at each level, and positions 1,999 (accepted) and 2,000 (refused). Assert that `SizeInput` raises where it should. |
| R2-07 | minor | `locked_commands.py:41–42`, `:85–95` and `:98–106` | **Signed sources are not pinned before measurement.** `locked_functions` compiles whatever signed bytes are on disk. Hashes are compared only within the run, so an edited `compute_benchmark.py` whose size sites still match would be measured.<br>P l.228 says "Refuse substituted sources". The sibling controller already pins its sources (`controller.py:19–20`). | Pin the `rev9-digests.json` values before any measurement, and refuse to start on a mismatch:<br>• `compute_benchmark.py` `d39f40a2…`;<br>• `reproduce_reports.py` `023e2e5b…`;<br>• `fixed_population.py` `ed670a6b…`;<br>• `ranking.py` `8bc2e7cb…`;<br>• `interval_diagnostic.go` `437f9c25…`;<br>• M `1caea1a5…`.<br>**Test:** a patched digest is refused. |

**Choices the author left open:** none. The README and the response pose no choices to the owner. The author's own choices are either the council's round-1 recommendation (slots used only as labels, the fixed mapping, gate CPU labelled non-binding, the integer cap) or stricter than required (same host, a synthetic-only schema). The one consequential choice the author did not flag is the separate fit, and R2-04 resolves it with a single recommendation.

**Considered and not counted:**
- **Peer's float versus integer cap divergence.** Refuted by executed evidence: an exhaustive check over all 100,000 success counts found 0 mismatches (`oc-controller/reviews/round-1-cap-output.txt`).
- **Peer's test with L_r > 40 and planted-null equivalence.**
  - A revert to `frame_runs` is already caught, because the test L_r ≠ 40.
  - The `min()` mutant is contrived.
  - Kernel parity is covered by `test_only_size_source_changed_in_kernel`.
- **No CLI execution.** Folded into the tests required by R2-01 and R2-02.
- **Schema.** Real inputs need a later reviewed schema change. This is fail-closed and within the round-1 scope.
- **`--cores` and `--oc-days` dropped.** The signed code validates both but never uses them.
- **Gate-CPU proxy.** It uses the signed `seed_for` timing, which the size-only rule retains.
- **Conservative RSS.** The fresh-process baseline is higher than the signed one, and maxima are taken over both slots and all cells.
- **`calibration_measurement` runs in the benchmark namespace.** Its `peak_rss_bytes()` is identical to the reporter's RUSAGE_SELF version (`compute_benchmark.py:45–47`, `reproduce_reports.py:71–73`).

## Verification

- **Main ran the tests** at this head on a clean tree, with pinned `.tools/venv/bin/python` 3.11.11 and `PYTHONDONTWRITEBYTECODE=1`: `locked-size-input/test_size_input.py` ran 16 tests in 650.663 s, OK. Main also recomputed the file digests above.
- **The council session and both peers had no shell.** They ran no tests and executed no mutants, and they did not recompute any digest. Every mutant-survival claim above is marked [INFERENCE].
- **Static checks made:**
  - compared every `SIZE_SITES` substitution against the signed bodies of `benchmark_population`, `trial_frame`, `execute_cell` and `calibration_measurement`;
  - compared `locked_fit` term by term against `reproduce_reports.projection` (l.214–254) and P l.181/185: counts 90,860 / 80,640 / 10,220, 4,000 versus 2,000 executions, separate gate charging, ×2 overhead, 8 workers, `8·max RSS ≤ hw.memsize − 8 GiB`;
  - checked the gate seed against M l.46–50 and `controller.seed`;
  - checked the pinned tips against P l.33–35;
  - checked the run-record contract against the final-evaluator draft;
  - read R9-reading-1.
- **How the peers changed the recommendation:**
  - **Experiment peer:**
    - found the host-key and run-record mismatch with the evaluator, which raised R2-01 to major;
    - gave R2-04's single-convention fix, R2-07 and the Wilson-reuse part of R2-05;
    - independently confirmed R2-06, which the lead also found.
  - **Challenger:**
    - added R2-03 and the M l.10 interpreter fields and evidence-hash race, both folded into R2-01;
    - framed the boundary gaps behind R2-04 and R2-05.
  - **Both peers** independently raised R2-02. The lead first leaned minor and raised it to major, because round 9 recorded that reading specifically to remove operator choice.
  - **The lead** independently found the missing load provenance (R2-01) and R2-06.
- **Nothing else was touched:** no real data, repositories, model calls or secrets, and no files were changed.

## Verdict

**7 issues open** (R2-01 and R2-02 major; R2-03 to R2-07 minor, with R2-06 carrying the R1-05 residual). R1-01 to R1-04, R1-06 and R1-07 are fixed. The draft is not accepted for its stage-3 role. Once the fixes are in, it needs another scoped round with zero issues, with the tests rerun and the digests recomputed. This review confers no authority to run the locked remeasurement, OC, inventory or any action on real data.
