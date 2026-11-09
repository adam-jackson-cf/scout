# Locked-size input — Option G draft

**DRAFT, synthetic only; requires a new zero-issue scoped council review.** These commands confer no preparation, inventory, locked remeasurement, OC or model authority. They accept only the synthetic schema below. No signed revision-9 file is modified. No real repositories are accessed.

## Implemented boundary and modes

The implementation addresses the size-source portions of P153/161/163/169/179–185. It is not an implementation of every stage-3 requirement or an authentic inventory/provenance authority.

- `size_input.py`: immutable complete eligible-run vectors (20–2,000 runs/repository, integer sizes 20–2,000), global offsets, identity/order-bound canonical SHA-256, manifest-seeded sampling-gate estimate.
- `population.py`: signed population constructor with only its size parameter, loop bound, size lookup and global offset lookup changed. The original numerical dataclasses and helpers are reused.
- `locked_commands.py`: runnable `benchmark`, `calibration`, `gate` and `fit` modes; `synthetic-input` makes a synthetic fixture. The signed function bodies are compiled with the explicit exact-AST `SIZE_SITES` substitutions. Each site must match once or execution refuses. No fallback to planning sizes is possible. All non-size scientific instructions remain those of the signed functions.

### Fixed identity and ordering

`SizeInput(record)` accepts exactly `schema`, `synthetic_only`, `inventory_seal_sha256` and `repositories`. The schema is `x1-locked-size-input/synthetic-v1`, `synthetic_only` must be true, and the seal is 64 lowercase hexadecimal characters. `synthetic_record(vectors)` fabricates synthetic commit identities and an all-zero **non-authentic** seal for tests only.

Repository records appear in this fixed kernel traversal order:

| Alias | Repository | Pinned tip |
|---|---|---|
| synthetic-python | django/django | a05e16a7230fe33d3f43f4a3aaed103cebb42efc |
| synthetic-typescript | microsoft/TypeScript | 6161772c1370bdd4a82f808e6823bd755eb1870d |
| synthetic-go | prometheus/prometheus | ef78ff64267724c8ecd1a1283cad49b365a8b67f |

Each repository record has exactly `alias`, `repository`, `pinned_tip`, `runs`. Each run has exactly `commit_sha` (unique full 40-lowercase-hex SHA), `first_parent_position` (strictly increasing integer 0–1,999, oldest first) and `size`. Positions specify first-parent topological order, not dates or lexicographic SHA order. This boundary validates the supplied ordering, not repository ancestry or completeness. Those require separately authorised inventory provenance. Swapped run/repository records change the canonical hash and are rejected for violating fixed order, rather than silently normalised. Wrong tips/mappings are also rejected.

The digest is SHA-256 of the **whole record**, compact sorted-key JSON with list order preserved, UTF-8 encoded. It binds schema, aliases/mapping, pinned tips, ordered full SHAs/positions/sizes and inventory seal. Validated vectors are mapping-proxied tuples, metadata is frozen, and record bytes are snapshotted; caller mutation cannot invalidate offsets or digest.

### Frozen size-source mapping (R1-01)

1. Both signed population slots **40 and 400 remain seed/identity labels only**. Each slot constructs every run from all three complete vectors in a fresh spawned interpreter, for both benchmark and calibration. Units are the sum of every vector. No prefixes, padding or size replication are used for population construction. The 22 fixed-key evaluations and calibration's ≤21 truth calls/one final frame/one enumeration check are retained.
2. The signed benchmark gate-CPU block samples 20 indices independently from each full vector. Its p_gate is explicitly **CPU-only, non-binding**; the standalone keyed `gate` result supplies the binding attempt cap.
3. All eight trial cells, production minimum 2,000 trials, B=1,999, worst-case 60,000-unit selected-frame fixtures, complete/full512 algorithms and one/two-worker scientific equality are retained. Uniform-SRS helper population lengths are each repository's locked L_r (including unequal lengths); mixed-explicit selected IDs come from each `range(L_r)`. The signed selected-frame templates are intentionally unchanged and are not population-preparation evidence.
4. Worst-case every-size enumeration is the unmodified signed reporter function.
5. Every measurement output contains every L_r, complete input record/digest, exact size-site substitutions and start/end source closure. The six signed source/manifest digests from rev9 are pinned before signed imports and again before measurement. Stable `host` has exactly `hw.model`, `hw.memsize`, `logical_cpus`, `os_build`. Separate `host_load.start` and `.end` record three load averages and the concurrent-process count (`ps -A -o pid=`). Every gate, benchmark, calibration and fit output records interpreter path/digest/version and `python_optimize=0`.
6. `fit` only extracts the ten evaluator binding inputs: full locked operation CPU maxima divided by total locked units where appropriate, separate enumeration and cheap gate costs, complete/full512 trial maxima, and phase RSS maxima. It emits **no resource decision, hours, stop status or duplicate arithmetic**. The final evaluator's `maximum_completed_inputs` and `resource_fit` alone apply the binding resource convention. Evidence is read once, then parsed and hashed from the same bytes. Only studentized evidence with the frozen eight cells, workers [1,2], B=1,999 and every measured trial count equal to the requested count (at least 2,000) is accepted.
7. A completed measurement with a closure mismatch is still published create-once with all values, both closures, `completed: true`, `integrity_failure: "closure mismatch"`, then exits nonzero. Missing/unreadable end snapshots retain a `null` digest and the read error, never an invented hash. `fit` also publishes completed-invalid when component start/end closures differ from each other or the current closure, even if an earlier component was unflagged. Size-input, host and interpreter mismatches still refuse. Components must be completed, have an exact supported failure flag, and record unoptimized CPython 3.11.11.
8. A rerun `fit` **must include every prior published fit**, oldest first, using repeatable `--prior FIT.json`. Each prior is read once, parsed and hashed from those bytes, and its digest joins `evidence_sha256`. The first prior contains one run; each subsequent prior must preserve that exact prefix and append one run. Inventory/input/host/interpreter bindings must match. Every prior's last run must have a qualifying integrity failure with its exact cause and required preserved evidence; an unflagged prior is refused. The new payload contains all prior runs followed by the new run, so a 9,000-byte failed construction maximum cannot be replaced by a 3,000-byte rerun. A fit-time closure failure changes only the new run, not prior causes. No command here authorises a rerun.

### Evaluator payload contract

`gate.locked_inventory` and `fit.locked_inventory` contain `eligible_sizes`, `gate_redraws=100000`, `gate_successes`, `p_gate`, `L_r` and `size_input_sha256`. `fit.measurement_runs` contains `runs` and `locked_inventory_sha256`, the SHA-256 of **canonical compact sorted JSON bytes** of the inventory object (no newline). Each run has the ten `provenance.INPUTS`, normalized host, interpreter, load provenance, completion/integrity classification, and size-input binding. To build evaluator bundle members, encode these two nested objects with `size_input.canonical`; hashing a pretty-printed variant instead would correctly fail the inventory binding. This is a consumable synthetic resource payload, not an authentic full OC bundle or authority claim.

### Keyed gate estimator (R1-03)

`SizeInput.estimate_p_gate()` executes exactly **100,000** calls to the parent controller's `seed('{"kind":"sampling-gate-estimate"}', 0, i, 'sampling-gate-estimate')`, i=0…99,999, imported explicitly by file path. The literal cell is canonical compact sorted JSON. No second HMAC or Wilson convention is implemented here: the interval calls `controller.wilson(successes,100000)`. One `random.Random(seed)` samples `range(L_r)` without replacement, 20 per repository in fixed mapping order. The estimator reports successes, p_gate, Wilson MC95, integer `ceil(4000*100000/successes)`, the independent 4,000 execution cap, and `not testable` for zero successes.

## Exact commands

From the repository root, after activating the already-pinned toolchain. Outputs must not exist. These are **synthetic-only reproduction commands**, not permission to perform real locked remeasurement. The full benchmark is intentionally not shortened.

```sh
D=evaluation/x1-triage-ranking/stage3-drafts/locked-size-input
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/test_size_input.py"

.tools/venv/bin/python "$D/locked_commands.py" synthetic-input --output /tmp/x1g-synthetic-sizes.json
.tools/venv/bin/python "$D/locked_commands.py" gate --input /tmp/x1g-synthetic-sizes.json --output /tmp/x1g-synthetic-gate.json
.tools/venv/bin/python "$D/locked_commands.py" benchmark --input /tmp/x1g-synthetic-sizes.json --trials 2000 --output /tmp/x1g-synthetic-benchmark.json
.tools/venv/bin/python "$D/locked_commands.py" calibration --input /tmp/x1g-synthetic-sizes.json --output /tmp/x1g-synthetic-calibration.json
.tools/venv/bin/python "$D/locked_commands.py" fit --input /tmp/x1g-synthetic-sizes.json --benchmark /tmp/x1g-synthetic-benchmark.json --calibration /tmp/x1g-synthetic-calibration.json --gate /tmp/x1g-synthetic-gate.json --output /tmp/x1g-synthetic-fit.json
```

### Finalise the original run before any rerun

After **any component exits nonzero with published closure-mismatch evidence**, keep that output, finish the remaining components and run `fit` on the **original component set**. The fit publishes completed-invalid and exits nonzero; it is the mandatory first `--prior` of any rerun. **Never rerun or replace a component before this fit exists.** Run the following commands separately, not chained with `&&`: a completed-invalid command intentionally returns nonzero after publishing its evidence. A command that publishes no output is not this recovery path.

If `benchmark` published `/tmp/x1g-synthetic-benchmark.json` and exited nonzero, retain it and the original gate, then finish calibration and finalise:

```sh
.tools/venv/bin/python "$D/locked_commands.py" calibration --input /tmp/x1g-synthetic-sizes.json --output /tmp/x1g-synthetic-calibration.json
.tools/venv/bin/python "$D/locked_commands.py" fit --input /tmp/x1g-synthetic-sizes.json --benchmark /tmp/x1g-synthetic-benchmark.json --calibration /tmp/x1g-synthetic-calibration.json --gate /tmp/x1g-synthetic-gate.json --output /tmp/x1g-synthetic-fit.json
```

If `calibration` published `/tmp/x1g-synthetic-calibration.json` and exited nonzero, all components already exist; retain them and finalise immediately:

```sh
.tools/venv/bin/python "$D/locked_commands.py" fit --input /tmp/x1g-synthetic-sizes.json --benchmark /tmp/x1g-synthetic-benchmark.json --calibration /tmp/x1g-synthetic-calibration.json --gate /tmp/x1g-synthetic-gate.json --output /tmp/x1g-synthetic-fit.json
```

If a separately authorised rerun is required, retain the failed fit and include it explicitly (use fresh rerun component/output paths):

```sh
.tools/venv/bin/python "$D/locked_commands.py" fit --input /tmp/x1g-synthetic-sizes.json --benchmark /tmp/x1g-rerun-benchmark.json --calibration /tmp/x1g-rerun-calibration.json --gate /tmp/x1g-rerun-gate.json --prior /tmp/x1g-synthetic-fit.json --output /tmp/x1g-rerun-fit.json
```

For another authorised rerun, repeat `--prior` for **both** `/tmp/x1g-synthetic-fit.json` and `/tmp/x1g-rerun-fit.json` in that order. Never substitute only the most recent cumulative fit: the sequence check requires every published fit.

Focused round-3 reproduction (fabricated records and disposable files only; no measurement modes or real data):

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/test_size_input.py" RoundThreeTests LockedModeTests.test_extraction_uses_locked_maxima_not_slopes -v
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/round3_evidence.py" --output /tmp/x1g-r3-size-checks.json
```

`round3_evidence.py` publishes the independent ten-input derivation, six focused fault-mutant results (construction minimum/first row, full512 mapped to complete, dropped gate divisor, dropped calibration frame, dropped fixed-key evaluations), function/source digests and five fabricated-record extraction wall times. It exits nonzero for a surviving mutant. These timings are not binding benchmark evidence. The heterogeneous fixture's hand answers use 3,000 total units: construction 6/3,000=.002, requested base 9/3,000=.003, calibration 12/3,000=.004, enumeration 4, gate 25,000/100,000=.25, complete .03, full512 .07, and RSS 3,000/4,000/5,000. Each list maximum is outside its first row; trial CPU maxima occur in different worker executions. The frame-dominant fixture is complemented by a fixed-key-dominant variant with second-row CPU 10, whose requested-base answer is exactly 10/3,000.

**Observed round-3 validation:** parent verification passed 30 tests in 596.312 seconds, all five focused extraction mutants were assertion-killed, and the actual full-count CLI integration passed in 593.939 seconds against the final controller closure. `reviews/round-3-response.md` maps all five findings and R2-02 residuals to lines/tests, exact commands, full output and SHA-256s. Logs and all six CLI receipts are preserved under `reviews/round-3-*`; historical round-2 receipts below are unchanged. These are synthetic validation results, not study results or run authority.

For a **single reproducible actual CLI integration run**, instead of the five manual measurement commands, run the following with a fresh destination. It invokes that exact five-mode sequence with no mocks, 100,000 keyed redraws, all eight studentized cells × 2,000 trials × both worker counts (32,000 helper trials), B=1,999, actual construction/calibration/enumeration and RSS. It checks the provenance round trip through `maximum_completed_inputs` and `resource_fit`, then create-once refusal. It retains all JSON evidence plus `cli-roundtrip.json` with exact argv, exit codes and the evaluator-only resource result.

```sh
PYTHONDONTWRITEBYTECODE=1 X1G_CLI_OUTPUT=/tmp/x1g-r2-cli .tools/venv/bin/python "$D/test_cli.py" -v
```

**Observed validation:** the round-2 unit suite passed 23 tests in 588.587 seconds. Actual synthetic CLI integrations passed in 605.024 seconds and, after the controller's late-fault-latch fix, 606.215 seconds. Each run exercised 32,000 actual helper trials with frozen scientific counts. Full logs are `reviews/round-2-test-output.txt`, `reviews/round-2-cli-test-output.txt` and `reviews/round-2-cli-final-test-output.txt`. Both six-record CLI receipts are preserved: `reviews/round-2-cli-evidence/` is the earlier controller closure; `reviews/round-2-cli-final-evidence/` binds the final controller source. Each directory retains exact invocations and the evaluator roundtrip; these are separate synthetic validation fixtures, not selectable real binding runs. The response records mappings and digests. Unit timing fixtures are explicitly fabricated; retained CLI measurements are actual synthetic measurements. No authentic locked fit is claimed.

## Independent known answers

The tests pin OpenSSL-computed vectors; the following reconstructs the first HMAC without importing any production code (the public manifest seed key is not a secret):

```sh
printf '%s\000%s\000%s\000%s' '{"kind":"sampling-gate-estimate"}' 0 0 sampling-gate-estimate | openssl dgst -sha256 -hmac 169-x1-oc-v4
printf '%s\000%s\000%s\000%s' '{"kind":"sampling-gate-estimate"}' 0 99999 sampling-gate-estimate | openssl dgst -sha256 -hmac 169-x1-oc-v4
```

Expected full HMACs: `f98555ac5c97329b205df7cf34ef7c58400eafd0ee5b3013253ee37552e418a3` and `a66f0e557732e338843ccea658cde74eb941e7fc5f6d0e83d1a8879a8725d494`. The first seed is decimal `17979871286122721947`. Expected heterogeneous L_r=(23,27,31) selections are literal vectors in `test_seed_and_heterogeneous_selection_known_answers`, derived from that literal seed independently of the controller.

The independent canonical-record digest fixture uses 20 runs of size 50 in each repository, zero seal, positions 0…19 and commit SHAs `format((repository_index+1)*10000+position, '040x')`, with the mapping/tips above. Compact sorted-key JSON, no newline, passed to `openssl dgst -sha256` gives `7f97e552a3e6ec6b4b056c3905056ea58ce1c82251a9cd0cfba0cddc12aa496a`. This expected constant was computed independently, not by importing `SizeInput`.

## Required-fix / regression map

| Issue | Implementation | Regression methods in `test_size_input.py` |
|---|---|---|
| R1-01 | `SIZE_SITES`, `locked_functions`, task functions, measurement modes and `measurement_run` | construction/calibration full-vector tests, selected-ID tests, actual-helper worker equality, size-site drift, extraction maxima, kernel AST parity |
| R1-02 | `SizeInput.__init__`, fixed schema/mapping/tips, canonical complete record | `test_identity_digest_and_invalid_order` (size/SHA/seal sensitivity, swapped run/repository hash change and refusal, pinned known answer) |
| R1-03 | `controller_seed`, `gate_selection`, `estimate_p_gate`, `gate_passes` | `test_seed_and_heterogeneous_selection_known_answers`, `test_full_100000_key_estimates`, `test_gate_boundaries` |
| R1-04 | complete locked constructor plus unchanged signed calibrator | `test_planted_effect_null_and_truth_identity` (unequal L_r, middle-repository 20/2000 extrema, factors .05/.20, targets .20/0), `test_mismatched_truth_mutant_raises`, `test_asymmetric_offsets_and_constant_zero` |
| R1-05 residual / R2-06 | `SizeInput.__init__` unchanged guards | `test_direct_malformed_records_and_position_boundary`: direct fourth/dropped repository, extra/missing keys at all three levels, positions 1,999 accepted / 2,000 refused; existing count/size boundaries retained |
| R1-06 | frozen input, tuples/mapping proxy, immutable canonical bytes | `test_immutable_snapshot` |
| R1-07 | this README replaces unsupported scope/result claims with exact implementation and commands | parent records actual execution evidence; no worker test-result claim |
| R2-01 | `host`, `host_load`, `interpreter`, `finish`, `read_evidence`, `measurement_run`, `locked_inventory` | `test_evidence_single_read_hashes_parsed_bytes`, `test_evaluator_provenance_roundtrip`, full `test_cli.py` integration |
| R2-02 and residuals R3-01–03 | `measurement_run`, `prior_measurement_runs`, `main`, `provenance`, `finish` preserve completed-invalid values across component/end-closure failures and reruns | `RoundThreeTests` cross-component publication, prior CLI maxima/read-once, prior validation/order, absent/unreadable snapshots and fit-end propagation tests |
| R2-03 | no method CLI option; `benchmark_mode` fixes studentized; `validate_benchmark` freezes cells/workers/counts | `test_frozen_method_and_cell_set_refused`; full CLI retains scientific counts |
| R2-04 | removed `locked_fit`; `measurement_run` extracts inputs only | `test_extraction_uses_locked_maxima_not_slopes`; evaluator round trip exercises sole resource decision |
| R2-05 | shared controller module supplies seed and Wilson | `test_interior_gate_wilson_and_ceiling` (3/100,000; cap 133,333,334; independently calculated bounds) |
| R2-07 | `SIGNED`, `pin_sources` before imports and measurement | `test_patched_signed_digest_refused_before_measurement` for every pinned source |
| R3-01 | `measurement_run` classifies component closure drift; `finish` publishes and exits nonzero | `test_cross_component_closures_publish_then_nonzero`, `test_fit_end_failure_only_marks_new_run` |
| R3-02 | repeatable `main --prior`, `read_evidence`, `prior_measurement_runs` preserve ordered cumulative fits and bind byte digests | `test_prior_cli_preserves_maxima_and_reads_once`, `test_prior_validation_and_cumulative_order` |
| R3-03 | tolerant end-only `provenance` and `finish` input snapshot preserve null/error | `test_missing_or_unreadable_end_snapshots_publish` (missing and unreadable, source and input) |
| R3-04 | heterogeneous `measurement_fixture`, all-ten literal assertions; independent `round3_evidence.py` derivation/mutants | `test_extraction_uses_locked_maxima_not_slopes` |
| R3-05 | `measurement_run` pinned component guards, existing `main` premeasurement guard | `test_component_input_host_interpreter_and_flags_refused`, `test_optimized_gate_subprocess_refused_before_measurement` |
| R4-01 | finalise-before-rerun recovery commands above; existing fit/prior assembly | `test_calibration_failure_finalised_before_rerun` preserves original calibration RSS 12,000 over rerun 4,000 |
| R4-02 | fixed-key-dominant extraction variant and `fixed-key-dropped` mutant | `test_extraction_uses_locked_maxima_not_slopes` |
| R4-03 | A/A/A→B fit-time closure regression | `test_agreeing_components_differ_from_fit_closure` asserts all ten literal inputs, preserved digest and completed-invalid nonzero publication |
