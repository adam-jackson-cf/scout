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
5. Both measurement outputs contain every L_r, complete input record/digest, exact size-site substitutions and source closure, including `size_input.py`, `population.py`, command driver, parent controller, signed numerical files and manifest. Sources and input are rechecked before create-once publication. Host records include hw.model, hw.memsize, logical CPU count and OS build. Closure is compared by canonical bytes after JSON round trips.
6. `fit` uses **full locked cost maxima**, separately measured enumeration, and max-phase lifetime-high-water RSS, not the planning L40/L400 slope or intercept. Binding and total-planning costs are separate; safety cannot cause the binding stop. Binding counts follow signed reporter arithmetic: 90,860 constructions, 80,640 requested evaluations, 10,220 calibrations/enumerations, and 4,000 expensive executions per binding population, separately charging the keyed integer attempt cap for cheap gate checks. All costs are multiplied by two. The fixed-eight-worker checks are ≤2,880 core-hours, ≤15 wall days and `8*max(phase RSS) <= hw.memsize - 8 GiB`. Zero gate successes or failed/unmeasured fit means not testable. This synthetic result never establishes an OC binding gate.

### Keyed gate estimator (R1-03)

`SizeInput.estimate_p_gate()` executes exactly **100,000** calls to the parent controller's `seed('{"kind":"sampling-gate-estimate"}', 0, i, 'sampling-gate-estimate')`, i=0…99,999, imported explicitly by file path. No second HMAC convention is implemented here. One `random.Random(seed)` samples `range(L_r)` without replacement, 20 per repository in the fixed mapping order. The estimator reports successes, p_gate, Wilson MC95, `ceil(4000*100000/successes)` via integer arithmetic, the independent 4,000 execution cap, and `not testable` for zero successes. The old caller-supplied-seed `draw` and unused `measurement_sizes` APIs are removed.

## Exact commands

From the repository root, after activating the already-pinned toolchain. Outputs must not exist. These are **synthetic-only reproduction commands**, not permission to perform real locked remeasurement. The full benchmark is intentionally not shortened.

```sh
D=evaluation/x1-triage-ranking/stage3-drafts/locked-size-input
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/test_size_input.py"

.tools/venv/bin/python "$D/locked_commands.py" synthetic-input --output /tmp/x1g-synthetic-sizes.json
.tools/venv/bin/python "$D/locked_commands.py" gate --input /tmp/x1g-synthetic-sizes.json --output /tmp/x1g-synthetic-gate.json
.tools/venv/bin/python "$D/locked_commands.py" benchmark --input /tmp/x1g-synthetic-sizes.json --method studentized --trials 2000 --output /tmp/x1g-synthetic-benchmark.json
.tools/venv/bin/python "$D/locked_commands.py" calibration --input /tmp/x1g-synthetic-sizes.json --output /tmp/x1g-synthetic-calibration.json
.tools/venv/bin/python "$D/locked_commands.py" fit --input /tmp/x1g-synthetic-sizes.json --benchmark /tmp/x1g-synthetic-benchmark.json --calibration /tmp/x1g-synthetic-calibration.json --gate /tmp/x1g-synthetic-gate.json --output /tmp/x1g-synthetic-fit.json
```

**Execution status:** the pinned command above (with `-v`) passed **16 tests in 586.708 seconds**. Full output is in `reviews/round-1-test-output.txt`; the issue mapping and file digests are in `reviews/round-1-response.md`. The actual-helper regression built the unchanged Go helper and compared all eight cells with one/two workers on small synthetic L_r, retaining **2,000 trials/cell and B=1,999** (32,000 helper trials overall). The initial one-trial fixture was refused by the signed helper; the fixture was corrected to the full required count, not the helper or algorithm. The separate full CLI timing/RSS commands above were not run, and no authentic locked fit is claimed. This is synthetic regression evidence only.

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
| R1-01 | `locked_commands.py`: `SIZE_SITES`, `locked_functions`, `population_task`, `trial_task`, measurement modes, `locked_fit`; removed unused size exposure | `test_construction_calibration_consume_every_locked_size`, `test_trial_population_lengths_and_selected_ids`, `test_actual_helper_one_two_worker_science`, `test_size_site_drift_refused`, `test_fit_uses_locked_maxima_not_slopes`, kernel AST parity |
| R1-02 | `SizeInput.__init__`, fixed schema/mapping/tips, canonical complete record | `test_identity_digest_and_invalid_order` (size/SHA/seal sensitivity, swapped run/repository hash change and refusal, pinned known answer) |
| R1-03 | `controller_seed`, `gate_selection`, `estimate_p_gate`, `gate_passes` | `test_seed_and_heterogeneous_selection_known_answers`, `test_full_100000_key_estimates`, `test_gate_boundaries` |
| R1-04 | complete locked constructor plus unchanged signed calibrator | `test_planted_effect_null_and_truth_identity` (unequal L_r, middle-repository 20/2000 extrema, factors .05/.20, targets .20/0), `test_mismatched_truth_mutant_raises`, `test_asymmetric_offsets_and_constant_zero` |
| R1-05 | constructor bounds/type/key/order checks | `test_constructor_run_count_and_size_boundaries`, `test_identity_digest_and_invalid_order`, middle-extrema offsets |
| R1-06 | frozen input, tuples/mapping proxy, immutable canonical bytes | `test_immutable_snapshot` |
| R1-07 | this README replaces unsupported scope/result claims with exact implementation and commands | parent records actual execution evidence; no worker test-result claim |
