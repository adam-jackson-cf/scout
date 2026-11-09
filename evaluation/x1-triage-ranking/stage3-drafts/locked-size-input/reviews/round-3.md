# X1 stage-3 draft review — locked-size input change (`stage3-drafts/locked-size-input/`) — round 3

Date: 2026-11-09. Reviewer: innovation council, scoped round 3. Two read-only peer audits ran independently, and neither saw the lead's leaning: `innovation-challenger` (X1GSizeReview3.LsiR3Challenger) and `experiment-peer` (X1GSizeReview3.LsiR3Peer). The lead wrote the synthesis.

**Basis:** X1 round 9 recommends option G (`reviews/round-9.md` l.244–247). This review covers the locked-size input change only as a stage-3 artefact. It grants no authority for preparation, inventory, locked remeasurement, OC or any action on real data.

## Reviewed files

- **Worktree:** `/Users/adamjackson/Projects/work/scout-x1g`, PR #21, head `38cead909de6de815ef8e9453630f404f6e2628f`.
- **Paths:** relative to `evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/`.
- **Digests:** Main recomputed these with `shasum -a 256` at that head. For the CLI evidence files, Main supplied 8-character prefixes; each matches the full value listed in `round-2-response.md`.

| File | SHA-256 |
|---|---|
| `README.md` | `d9ba449bee07dddab0d0bbfc9e39d305385516beb1c0fb0050c005f3c831e28b` |
| `locked_commands.py` | `6c33a8d0aee617f19ce086a57a25a3a66a0bd9826c95a6d310fc15b7481bc28f` |
| `population.py` | `a32e33cd8061940067d94b56cead9a85608a3a07c79b8a9b89c248482586f0c6` (unchanged since round 1) |
| `size_input.py` | `579c5119a812f64ffaa574f4ec9d7f77d12b89f47d407511d6379ee7ce647cdc` |
| `test_size_input.py` | `3551c0af440d10fd6c439453495f37bc5ad48380f5d705d0347f2a3426176cc9` |
| `test_cli.py` | `d6e22f609f5caa529d5ec2380edf7209aff4d10f323aeae968b161854c1e979d` |
| `reviews/round-2.md` | `ef1f7b6e46423ef978feedbb765ab1df3d29562eb5b55c69769a0df9ec3d69ff` |
| `reviews/round-2-response.md` | `bbcba191886038cd4d586d18d13d72ebbada4a6cfe3d6965d39ea6b8c002fc9b` |
| `reviews/round-2-test-output.txt` | `1d89aa852e39f51d90fcdc3f85b78fd305f62955d9a5dc9acc778e1fd51f519d` |
| `reviews/round-2-cli-final-test-output.txt` | `fb7b075e7a0abc8def566e6ca5449c23fc6359c7b374a08da389670bd9ce46f3` |
| `reviews/round-2-cli-final-evidence/benchmark.json` | `d62f4c27ff416cb55186252ddc0dc86e45adff13ec5949dabf20d106bde7cfa8` |
| `…/calibration.json` | `55c9542f3c4136193d9d45f168a06c510ba76c691e35d8b80d612ea6998e679e` |
| `…/cli-roundtrip.json` | `74aaf2e421f35f263cf9955904a334e5b744d6f84cb979edfd67c5733766cf7c` |
| `…/fit.json` | `e76828a035cc402d8c4c7790a5c14cbfb8d456b4d19cce095401411c9231c31c` |
| `…/gate.json` | `0ab5cd09d4ebf4e2cc80a1edf51ed42577712920a669f888c4ccf7cdf3b1c469` |
| `…/synthetic-input.json` | `6f8880093a21621ba32fdfc2df2441fa4c8c4e18a84082c1501a194ac23e33c6` |

## Signed basis (unchanged)

| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` (P) | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `oc-manifest.json` (M) | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |

Signed text relied on:
- **P l.181:**
  - "Record load averages and concurrent-process count in provenance"
  - "Rerun only after a recorded integrity failure … a closure mismatch …"
  - "Report every completed run; each binding input is the maximum over all completed runs"
- **M l.10:** "executable path/digest and sys.flags.optimize=0 in every new output; non-zero optimization refused".
- **R9-reading-1** (`round-9.md` l.326–329): "A run that reached its end is a completed run, even if its closure check then fails. Its values therefore enter the maximum". It was recorded "so that no operator choice arises later" (l.322).
- **Sibling contract:**
  - `final-evaluator/provenance.py` `INPUTS`, `validate_host`, `maximum_completed_inputs` (l.90–125; a single valid run is accepted at i = 0);
  - `final-evaluator/README.md` l.12–13.

## Round-2 disposition

| ID | Fixed | Evidence (file:line + catching test) |
|---|---|---|
| R2-01 | **Yes** | **Code:**<br>• `host` `locked_commands.py:161–166` emits the evaluator's exact four keys.<br>• `host_load` `:169–172` records start and end load averages and the `ps -A -o pid=` count.<br>• `interpreter` `:175–178` records path, version, `python_optimize` and the resolved-executable SHA-256. All of these are added to every gate, benchmark, calibration and fit output by `finish` `:366–371`.<br>• The run record `:258–266` carries the 10 `INPUTS`, the top-level load fields from the benchmark start, `completed` and `integrity_failure`.<br>• `locked_inventory` `:286–290` uses `gate_redraws`/`gate_successes`.<br>• `read_evidence` `:293–299` reads once, then hashes and parses those bytes.<br>**Tests:** `test_evidence_single_read_hashes_parsed_bytes` (:347), `test_evaluator_provenance_roundtrip` (:385), and `test_cli.py:19`, a full five-mode CLI run checked against `maximum_completed_inputs` and `resource_fit`.<br>**Final CLI evidence:** `fit.json` has a four-key host, three load averages, optimize 0 and executable digest `570b4105…`, which matches the round-9 toolchain pin. |
| R2-02 | **Partly** | **Fixed path:** `finish` `:354–375` publishes create-once when a closure source's *content* changes during one command, with `completed: true`, `integrity_failure: "closure mismatch"` and both closures, then raises.<br>**Test:** `test_closure_failure_published_create_once_then_nonzero` (:354) kills the "raise before publish" mutant.<br>**Not fixed:** the outcome R2-02 exists for — a completed run's values reaching the binding maximum alongside its rerun — fails on three ordinary paths. These are R3-01, R3-02 and R3-03. |
| R2-03 | **Yes** | **Code:** no `--method` option in the parser (`:310–316`). Studentized is fixed in `benchmark_mode` (`:182`). `validate_benchmark` (`:270–283`) freezes the eight cells, workers [1, 2], B = 1,999 and `measured_trials == trials`.<br>**Test:** `test_frozen_method_and_cell_set_refused` (:323) covers top-level method, per-cell method, a missing cell, a duplicate cell, wrong workers and a 1,999 trial count. |
| R2-04 | **Yes** | **Code:** `locked_fit`, every stop decision and the hour arithmetic are gone. `measurement_run` only extracts inputs. README item 6 and the issue map are updated.<br>**Test:** `test_extraction_uses_locked_maxima_not_slopes` (:307) asserts that `fits_caps` is absent.<br>The replacement extraction is under-tested; that is R3-04. |
| R2-05 | **Yes** | **Code:** `size_input.py:124` calls `controller().wilson(successes, DRAWS)`.<br>**Test:** `test_interior_gate_wilson_and_ceiling` (:217): 100,000 calls, cap 133,333,334, and bounds checked to 15 places with an independent formula.<br>**Lead and peer hand check:** interval ≈ [1.0203×10⁻⁵, 8.8208×10⁻⁵]; ⌈4×10⁸/3⌉ = 133,333,334. A floor mutant gives 133,333,333 and is killed. A mutant that drops the variance term shifts the radius by about 4×10⁻¹⁰ and is killed [INFERENCE]. |
| R2-06 (R1-05 residual) | **Yes** | **Test:** `test_direct_malformed_records_and_position_boundary` (:91) mutates the records directly:<br>• 2 or 4 repositories ("exactly three"), which kills the `zip`-truncation mutant;<br>• an extra or missing key at the record, repository and run levels;<br>• position 1,999 accepted and 2,000 refused.<br>R1-05 is now closed. |
| R2-07 | **Yes** | **Code:** `SIGNED` (`:29–36`) matches `rev9-digests.json` `files_sha256` for all six files (lead and peer checked with jq). `pin_sources` runs at import (`:44`), before the signed modules load, again in `locked_functions` (`:100`) in every worker, and in `main`.<br>**Test:** `test_patched_signed_digest_refused_before_measurement` (:341) patches each pin in turn.<br>The challenger's check/use race is listed under suggestions. |

## New issues

| ID | Severity | Location | Concrete failure | Required fix |
|---|---|---|---|---|
| R3-01 | **major** (R2-02 residual) | `locked_commands.py:220–221` | **Input:** the README order runs `gate` under closure A. `benchmark` starts under A, and a closure file changes to B during the run (for example `controller.py`, which this PR has already edited once). `benchmark.json` is correctly published as `"closure mismatch"`. `calibration` and `fit` then run under B.<br>**Wrong outcome:** at fit time the gate's provenance A differs from the current closure B, and the gate's `integrity_failure` is `None`. Fit therefore raises "source/input closure mismatch" and publishes nothing. No run record ever carries the completed run's values or its classification.<br>Getting a record needs an undocumented operator step: revert to A, or rerun a component, which drops completed values. A source change *between* commands fails the same way. That is the operator choice R9-reading-1 was recorded to remove.<br>The tests cannot reach this branch: :385 sets the failure flag but keeps every provenance equal to the current closure. | **Code:** when any component's start or end provenance differs from another component's, or from the fit-time closure:<br>1. publish the run record with `completed: true`, `integrity_failure: "closure mismatch"` and `preserved_evidence_sha256`;<br>2. then exit non-zero, instead of raising before publication.<br>Keep refusing a size-input, host or interpreter mismatch.<br>**Test:** gate under A, benchmark flagged A→B, calibration and fit under B. Assert that `fit.json` exists with the classification and all ten values, and that the exit is non-zero. |
| R3-02 | **major** (R2-02 residual) | `main` `:341–349` (always `runs: [one record]`); parser `:310–316` (no way to pass a prior run); README item 7 and "Evaluator payload contract" | **Input:** a completed run fails its closure check with construction RSS 9,000 B. A full rerun measures 3,000 B.<br>**Wrong outcome:** the rerun's `fit` emits `measurement_runs.runs = [rerun]`. The README tells the operator to encode `fit.measurement_runs` directly as the bundle member. `maximum_completed_inputs` accepts a lone valid first run (`provenance.py:98`), so the binding maximum becomes 3,000 B.<br>This breaks P l.181 ("maximum over all completed runs") and R9-reading-1. README item 7 promises the failed run is "included with the permitted rerun, never replaced", but no code does this. No sibling draft assembles payloads with several runs, and round 9 assigns this to no later stage.<br>The only test of this path (:385) hand-assembles `[invalid, record]`. | **Code:** `fit` accepts the prior published fit outputs in order, for example a repeatable `--prior FIT.json`. It must:<br>• read each prior once and hash it;<br>• require the same `size_input_sha256`, `locked_inventory_sha256` and host;<br>• require that the last prior run carries a qualifying integrity failure, keeping its exact cause;<br>• emit `runs = prior runs + new run`;<br>• bind the prior digests into `evidence_sha256`.<br>**README:** state that a rerun fit must include every prior fit.<br>**Tests:**<br>• a flagged prior plus a valid rerun gives one payload with both runs, and the evaluator maximum is 9,000;<br>• a prior with no failure is refused;<br>• a prior with a different inventory or host is refused. |
| R3-03 | minor (R2-02 residual) | `finish` `:355–356` | **Input:** after measurement, a closure file or the `--input` file disappears. Two ways this happens: `git checkout` in the PR worktree removes `stage3-drafts/*`, or the README's `/tmp` input is cleaned up.<br>**Wrong outcome:** `provenance()` or `input_path.read_bytes()` raises FileNotFoundError before `publish`. The completed benchmark or calibration result is lost. A *changed* file is preserved, but a *missing* one is not, although both are closure mismatches. | **Code:** while computing the end closure, record a missing or unreadable path or input as absent, with a `null` digest plus the error. Classify that as `"closure mismatch"`, publish create-once, then exit non-zero. Do not invent hashes.<br>**Tests:** after supplying completed measurements, delete a disposable closure file, and separately the input. Assert publication with the classification and a non-zero exit. |
| R3-04 | minor | `measurement_run` `:236–256`; `measurement_fixture` and `test_size_input.py:307` | **The extraction of binding inputs is not regression-tested** — the same defect class as R2-04, now in the replacement code.<br>The fixture is degenerate:<br>• one population row and one calibration row;<br>• complete and full512 trial CPU both 1e-6;<br>• fixed-key CPU equal to calibration frame CPU;<br>• equal RSS across cells.<br>`gate_cpu_per_check`, `complete_trial_cpu` and `full512_trial_cpu` are never asserted, and the CLI round trip compares the record only with itself.<br>**Mutants that survive [INFERENCE]:**<br>• `min` or `[0]` in place of `max`;<br>• `full512_trial_cpu: worst[False]`;<br>• the `/100000` gate divisor dropped;<br>• the calibration-frame term dropped from `reevaluate`.<br>**Impact:** with the final CLI values (0.0474335 versus 0.0010781 s), the full512 mutant understates binding CPU by about 252 × 5 × 4,000 × 2 × 0.0463554 / 3600 ≈ 129.8 core-hours, which is 4.5% of the 2,880 cap (peer arithmetic, lead-checked). That can flip a fit near the cap. | **Fixture:** use a heterogeneous fixture:<br>• at least two rows in each list, with each maximum outside the first row;<br>• different complete and full512 maxima, placed in different worker executions;<br>• calibration frame CPU greater than fixed-key CPU;<br>• a non-trivial gate CPU.<br>**Assertions:** assert all ten extracted values against hand-computed constants. |
| R3-05 | minor | `main` `:325`; `measurement_run` `:224–227` | **M l.10 requires "non-zero optimization refused".**<br>• The only guard is `:325`, and no test reaches it.<br>• `measurement_run` checks that the components share an interpreter, but never checks that `python_optimize == 0` or that the version is 3.11.11.<br>**Failure:** delete `or sys.flags.optimize` and `python -O locked_commands.py benchmark …` runs, recording `python_optimize: 1`. Fit and the evaluator then accept that output, and no test fails [INFERENCE]. This is an untested boundary of an explicit signed safeguard. | **Code:** `measurement_run` refuses any component with `python_optimize != 0` or `python_version != "3.11.11"`.<br>**Tests:**<br>• a fixture with `python_optimize: 1` is refused;<br>• a `-O` subprocess `gate` invocation exits non-zero before any measurement. This is cheap, because the refusal comes before the 100,000 draws. |

**Choices the author left open:** none.
- The README and response pose no choices to the owner (R-064 is met).
- Making the evaluator the only resource convention follows R2-04's own recommendation.
- Keeping two synthetic CLI receipts is disclosed, and they are labelled as not selectable binding runs.
- Accepting `--trials` > 2,000 is fail-closed: more trials can only raise a maximum.

## Non-blocking suggestions

These are not counted. Each needs hand-crafted or tampered evidence, an adversarial swap, or an unlikely trigger.

- **Laundered completion or failure flags** (challenger F2).
  - `measurement_run` never validates the incoming `completed` flag, and it maps any truthy failure to `"closure mismatch"`. The producers emit only `completed: true` with `None` or `"closure mismatch"`, so no current path triggers this.
  - If R3-02's prior-fit input is added, validate completion and the exact cause there, and refuse rather than reclassify.
- **Population and calibration slot membership** (challenger F3). There is no check that both identity slots 40 and 400 are present exactly once. Only edited evidence could drop a row, because `measurements` always maps both slots. A cheap membership check would mirror R2-03.
- **Check/use race** (challenger F4). After `pin_sources`, the bytes are re-read for import, `inspect.getsource` and `go build`. A swap-and-restore attack is adversarial. Executing from the verified snapshots would close the race.
- **Run-record propagation of the fit's own closure** (`:362–365`) is untested. R3-01's test may cover it.
- **A flagged gate feeds `locked_inventory`.** Refusing a flagged gate for the inventory would only be stricter. The gate is deterministic.
- **Publication is not atomic.** `json.dump(..., allow_nan=False)` streams into a file already opened with `"x"`, so a non-finite value leaves a partial file that blocks republication. Serialize first, then write once.
- **The evaluator's inventory validator** (`evaluator.py:114–124`) is not part of the round trip, so renaming the inventory keys would fail closed only at evaluation time.
- **Untested `validate_benchmark` branches:** `trials < 2000`, `replicates != 1999` and a false byte-equality flag.
- **For the final-evaluator review, not this artefact:** `maximum_completed_inputs` allows chained reruns. Whether that matches "permits one rerun" belongs to that review.

## Verification

- **Main's runs:** at head `38cead90`, on a clean tree, with pinned `.tools/venv/bin/python` 3.11.11 and `PYTHONDONTWRITEBYTECODE=1`:

  | Suite | Tests | Time | Result |
  |---|---|---|---|
  | `locked-size-input/test_cli.py` | 1 | 597.955 s | OK |
  | `locked-size-input/test_size_input.py` | 23 | 585.722 s | OK |
  | `oc-controller/test_controller.py` | 25 | 2.175 s | OK |
  | `final-evaluator/test_evaluator.py` | 20 | 258.191 s | OK |

  Main recomputed all the digests above.
- **What this session did not do:**
  - The council session and both peers had no shell. They ran no tests and no mutants, and recomputed no digests.
  - Every mutant-survival and failure-path claim above is static [INFERENCE], except the hand arithmetic, which is labelled.
- **Static checks made:**
  - `SIGNED` against `rev9-digests.json`;
  - every new code path against P l.181, M l.10, R9-reading-1 and `provenance.py` l.80–125;
  - the README order (gate → benchmark → calibration → fit) traced through `measurement_run` `:220–221` for R3-01;
  - the final CLI `fit.json`: four-key host, load fields, interpreter digest, L_r 25/30/35 and `p_gate` 1.0, against the evaluator contract;
  - the Wilson bounds and the cap for 3/100,000 by hand.
- **How the peers changed the recommendation:**
  - **Lead:** found R3-01 and R3-02 independently. Both peers confirmed them, and the experiment peer gave both their concrete RSS and closure scenarios.
  - **Both peers:** independently raised R3-03 and R3-04. The lead first had extraction coverage as a suggestion and counted it after the peer's 129.8 core-hour impact, for consistency with R2-04 and R2-05.
  - **Experiment peer:** raised R3-05. The lead counted it as an untested boundary of an explicit M l.10 safeguard, consistent with R2-06.
  - **Challenger's F2, F3 and F4:** the lead moved these to suggestions, because each needs tampered evidence or an adversarial swap that the producers cannot generate.
- **Nothing else was touched:** no real data, repositories, model calls or secrets, and no files were changed.

## Verdict

**5 issues open:**
- **Major:** R3-01 and R3-02.
- **Minor:** R3-03, R3-04 and R3-05.

R2-01, R2-03, R2-04, R2-05, R2-06 (closing R1-05) and R2-07 are fixed. R2-02 is fixed only for a content change within one command; its residuals are R3-01, R3-02 and R3-03.

The draft is not accepted for its stage-3 role. Once the fixes are in, it needs another scoped round with zero issues, with the tests rerun and the digests recomputed. This review confers no authority to run the locked remeasurement, OC, inventory or any action on real data.
