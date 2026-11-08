# X1 stage-3 draft review — locked-size input change (`stage3-drafts/locked-size-input/`) — round 1

Date: 2026-11-08. Reviewer: innovation council, scoped round 1. Two independent read-only peer audits ran without seeing the lead's leaning: `innovation-challenger` (X1GSizeReview.LsiChallenger) and `experiment-peer` (X1GSizeReview.LsiPeer). The lead wrote the synthesis.

**Basis for this review:** X1 round 9 recommends option G (`reviews/round-9.md` l.244–247 and §11 step 5). This review covers the locked-size input change as a stage-3 artefact only. It does not authorise preparation, inventory, the locked remeasurement, OC or any real-data action.

## Reviewed files

Worktree `/Users/adamjackson/Projects/work/scout-x1g`, branch `agent/x1-option-g`, head `7168c0e1b2cbfaacbd3ab81da0910eeaa09b531e` (PR #21). Paths are relative to `evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/`.

| File | SHA-256 |
|---|---|
| `README.md` | `dff0e3b5f54d62b4427cf8d5fc3e930c7f50b6446d8dfe962f818b240d4f2543` |
| `size_input.py` | `8275658513a54f7da4b9927f6a49a9485aefe7af1b1fc787fbf727df5187aedf` |
| `population.py` | `a32e33cd8061940067d94b56cead9a85608a3a07c79b8a9b89c248482586f0c6` |
| `test_size_input.py` | `3a4034dfaf353c2eb568131e4faf0ee032d672044cb642e02e4eb1a9fb6d5368` |

These digests come from Main's report. This session had no shell, so the council did not recompute them.

## Signed basis

- `PREREGISTRATION.md` (P): `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`.
- `README.md`: `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3`.
- `reviews/rev9-digests.json`: `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`.
- `reviews/round-9.md`: `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`.

Text relied on:
- **P:** l.37, l.39, l.41 (frame order, keyed repository names, SRS20), l.153, l.161, l.163, l.169, l.177, l.179, l.181 and l.224.
- **`oc-manifest.json`:** l.41–49 and l.57–63 (sizes, seed key `169-x1-oc-v4`, phases, gates, sampling-gate estimate).
- **`reviews/round-8.md`:** l.252 (§2 E), l.254 (option G = "write the locked-size benchmark and calibration mode … against synthetic sizes"), l.305 (hard-coded size sites) and l.480 ("a reviewed locked-size command").
- **`reviews/round-9.md`:** §6 l.337–339, §7 l.358, §9 l.390 and §12 l.451.
- **Signed executables:** `validation/compute_benchmark.py`, `validation/reproduce_reports.py` and `validation/interval_diagnostic.go`.

## What is sound

- **Population construction.** `population.py` `prepare_population` is identical to the signed `fixed_population.prepare_population` (l.47–88) except for the four size-source substitutions:
  - the `size_input` parameter;
  - the run-loop bound;
  - `size_input.size`;
  - `size_input.offset`.
- **Offset formula.** The offset uses the same left-to-right float expression as the signed kernel. Its extremes are global across all three repositories, as P l.161 requires. Constant sizes give zero through `max(1, …)`.
- **Validation limits.** The run-count limit of 20–2,000 matches P l.41 (at least 20 eligible runs) and P l.37 (a 2,000-commit envelope). The per-run size limit of 20–2,000 matches P l.41. Booleans and non-integers are refused, which is fail-closed.
- **Gate predicate.** `min(totals) >= 1000 and sum(totals) <= 60000` matches P l.163, manifest l.57 and the signed benchmark gate (`compute_benchmark.py` gate block).

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R1-01 | **blocking** | `README.md:11`, `:16`; `size_input.py:43–45`. Signed size sites: `compute_benchmark.py:31–32`, `:163`, `:185–191`, `:268`; `reproduce_reports.py:76–82`, `:108`, `:151`; `interval_diagnostic.go` trial loop `chosen:=rng.Perm(len(pop))[:20]` | **The rule-5 size input for the locked remeasurement is not implemented.** P l.181 requires the reviewed benchmark and calibration commands "changing only the size source to the locked sizes; that size input is the only permitted code change and is reviewed under rule 5". P l.179 requires remeasuring "these costs, RSS and both hard counters on locked sizes through L = 2,000".<br>Round 8 defines option G as "the locked-size benchmark and calibration mode" (l.254) and requires "a reviewed locked-size command" (l.480).<br>The draft changes, wraps and specifies none of the signed size sites. Open choice 2 defers "the concrete benchmark size-source mapping" to a review held after inventory. That is the post-inventory freedom round 8 §2 E rejected: the binding trial CPU and RSS are maxima, so a structure chosen once the sizes are known moves the binding input.<br>The signed trial helper permutes every run of the population on each trial, so trial worker-plus-helper RSS and CPU depend on the eligible-run count L. P l.181's "locked-size peak RSS of … trial worker plus helper" therefore needs locked L in the trial cells.<br>`measurement_sizes()` has no caller. | Write and test the locked mode now, on synthetic vectors. The only code difference from the signed bytes may be the size source.<br>**Recommended mapping:**<br>1. **Population slots.** Keep both signed slots (40 and 400) as seed and identity labels only, so the existing `seed_for(...)` keys and the cell set stay unchanged. In each slot, build the complete locked population through `population.prepare_population` in a fresh process, for both `benchmark_population` and `calibration_measurement`, with its 22 fixed-key evaluations and its ≤21-truth/one-frame/one-enumeration check. Units are the sum over all three locked vectors.<br>2. **Gate-CPU block.** Draw a per-repository SRS20 over the locked vectors. Its p_gate output is labelled CPU-only and non-binding; the binding p̂_gate comes from R1-03.<br>3. **Trial cells.** Keep the eight cells, 2,000 trials, B = 1,999, the one/two-worker equality check and the worst-case 60,000-unit selected-frame content. In uniform-srs cells, the population run count becomes each repository's locked L_r. In mixed-explicit cells, draw selected IDs from `range(L_r)`.<br>4. **Enumeration.** Leave the worst-case enumeration unchanged (P l.169).<br>5. **Closure and provenance.** Add `size_input.py`, `population.py` and the size-input digest to both commands' source closure and outputs, and record every L_r.<br>6. **Fit.** The binding fit uses the locked maxima (P l.181), not the planning `projection` slope.<br>Test that construction, calibration and trial phases all consume the locked vectors, and that the scientific output is identical with one and two workers on a small synthetic vector. Connect `measurement_sizes()` or delete it. |
| R1-02 | major | `size_input.py:6`, `:16–25`; `README.md:15`; `test_size_input.py:34` | **The input record binds neither the identity nor the order of the locked inputs.**<br>The digest covers only `{synthetic alias: sizes}`. It includes no repository name, pinned tip, run commit SHA, inventory seal or schema tag. The alias-to-repository mapping and the order of runs inside each vector are never fixed. The kernel draws random numbers sequentially in alias order and then run-index order (`population.py:23–24`), so permuting aliases or runs yields different populations and gate draws under the same seeds. That is an unreviewed post-inventory choice, against P l.41 ("never re-seed/redraw").<br>P l.177 requires "authentic locked-input … provenance", P l.37 fixes the frame order, and manifest l.44 requires "complete eligible-run size vectors from the locked inventory".<br>The only digest test is insertion-order invariance, and any deterministic hash passes it. | **Fix the mapping:**<br>• `synthetic-python` → `django/django`<br>• `synthetic-typescript` → `microsoft/TypeScript`<br>• `synthetic-go` → `prometheus/prometheus`<br>This follows P l.33–35 order and matches each repository's language.<br>Order runs by first-parent topological position, oldest to newest (P l.37).<br>Hash one canonical record containing: a schema tag, the mapping, the pinned tips, the full 40-hex SHA and size of each run, and the inventory seal SHA-256.<br>Test that each of these changes the digest: one size, two swapped runs, two swapped repositories, a changed seal. Pin one known-answer digest. |
| R1-03 | major | `size_input.py:37–41`; `README.md:7`; `test_size_input.py:28–33` | **The "keyed deterministic SRS20" is neither keyed nor tested.**<br>`draw` passes any caller seed to `random.Random`. P l.163 and manifest l.46–49 and l.63 require manifest HMAC seeds (`169-x1-oc-v4`, phase `sampling-gate-estimate`) and 100,000 redraws reporting successes, p̂_gate and Wilson MC95. The attempt cap is ceil(4,000/p̂_gate), and zero means not testable. Nothing in the drafts computes p̂_gate, although it sets one of the "both hard counters" (P l.179). The controller and evaluator drafts only accept it as an input.<br>Every gate test vector has exactly 20 runs, so the SRS takes the whole frame and the seed is never exercised. The N_r ≥ 1,000 boundary is never reached (each repository totals 20,000), so deleting that condition survives the tests [INFERENCE: mutants not executed]. | Add `estimate_p_gate()`:<br>• seeds from the existing `oc-controller/controller.py` `seed(cell, 0, i, "sampling-gate-estimate")` for i = 0…99,999 (reuse it; do not add a second seed convention);<br>• a fixed literal cell identity, frozen now: `{"kind":"sampling-gate-estimate"}`;<br>• per-repository `sample(range(L_r), 20)` in the fixed mapping order;<br>• return successes, p̂, Wilson MC95, the integer cap `ceil(4000·100000/successes)`, and "not testable" when there are zero successes.<br>Either connect `draw` to it or delete `draw`.<br>Test:<br>• against independently computed HMAC digests;<br>• on heterogeneous vectors with L_r > 20, with independently expected selections;<br>• on planted p = 0 and p = 1 vectors;<br>• N_r 999 rejected and 1,000 accepted;<br>• ΣN 60,000 accepted and 60,001 rejected. |
| R1-04 | major | `test_size_input.py:11–26` | **No planted-effect or planted-null check on the new input shape** (rule 4).<br>Inherited parity covers only identical length-20 vectors. The asymmetric test checks only structure. Its offset assertion is circular: it compares `size_offset` with the same `inputs.offset`. Apart from l.25, it does not independently check the global offset.<br>Nothing calibrates a population with unequal L_r and heterogeneous sizes across repositories, so nothing shows a planted .20 is reachable, a null of ≤ 0 is reached, or the P l.169 identity `_population_gain(prepared, latent_gain) == realised_gain(frame)` holds. | On asymmetric vectors (unequal lengths, sizes spanning 20–2,000, global minimum and maximum placed inside the middle repository), run the signed `calibrate_population` on `prepare_population(size_input=…)` with targets .20 and 0 at size factors .05 and .20.<br>Assert reachability, the truth-equals-frame identity, and that a mismatched-gain mutant raises.<br>Add an independently computed expected offset per run, and an all-zero offset vector for constant sizes. |
| R1-05 | minor | `size_input.py:17–24`; `test_size_input.py:35–36` | **Constructor boundaries are untested.** Only the boolean case is checked. Mutants that survive by inspection [INFERENCE]:<br>• `len <= 2000` changed to `< 2000`;<br>• `20 <= n` changed to `20 < n`;<br>• `minimum`/`maximum` taken from the first or last repository (the test vectors place the global extremes there). | Test:<br>• 19, 20, 2,000 and 2,001 runs;<br>• sizes 19, 20, 2,000 and 2,001;<br>• floats;<br>• extra and missing repository keys;<br>• global extremes placed inside one repository. |
| R1-06 | minor | `size_input.py:19`, `:25–27` | **The validated vectors can be changed after construction.** `vectors` is a mutable public dict, but the digest, minimum and maximum are cached. Reassigning one entry leaves a stale digest and wrong global offsets. P l.228 says "Refuse substituted sources". | Store the vectors immutably (for example `MappingProxyType` over tuples, or a frozen dataclass). Add a mutation-refusal test. |
| R1-07 | minor | `README.md:5`, `:7`, `:27` | **The README overclaims.** "Implements … 163, 179–185" is untrue (R1-01, R1-03). "4 tests passed" is author-reported only. | Correct the line claims after the fixes. Report the reproduced test result with its exact command. |

## Open-choice verdicts

1. **Positional aliases for the three locked repositories: raise R1-02.** Aliasing in kernel traversal order is acceptable. The mapping must be explicit (django → python, TypeScript → typescript, prometheus → go, following P l.33–35 order) and bound into the digest together with run order and the inventory seal.
2. **Complete, ordered, unequal-length vectors with no prefixes, resampling or replication: accept** for population construction. This matches P l.153, P l.161 and manifest l.44. **The deferral in the same choice raises R1-01**: the concrete size-source mapping for the benchmark and calibration commands must be fixed in this artefact before the sizes are known.
3. **Keep the worst-case 60,000-unit trial fixtures and replicates: accept in part.**
   - **Accepted:** keeping the selected-frame content, 2,000 trials, B = 1,999, the eight cells, the one/two-worker equality check and the worst-case enumeration. This matches P l.179 and P l.181 and is fail-closed in units.
   - **Issue (in R1-01):** the eligible-run count of the trial cells is a size source. Under P l.181 it must become each repository's locked L_r, so that trial worker-plus-helper RSS is measured at locked size.
   - The challenger accepted this choice in full. The council follows the experiment peer, because the signed helper permutes every run of the population on each trial (see R1-01).

## Verification run

- **No tests were run.** This session, and both peer sessions, had only read, grep and glob tools, with no shell and no write access. The README command `PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/test_size_input.py` was not executed, and "4 tests passed" was not reproduced. No mutation probes were executed. Every mutant named above is marked [INFERENCE] from reading the code.
- **File digests were not recomputed** (see Reviewed files).
- **Static checks made:**
  - compared the draft `prepare_population` line by line with signed `fixed_population.py` l.47–88;
  - read the size sites and seed keys in `compute_benchmark.py` and `reproduce_reports.py`, and the trial permutation in `interval_diagnostic.go`;
  - confirmed `seed()` and the `p_gate` inputs in `stage3-drafts/oc-controller/controller.py` l.14–16 and l.40–50, and in `final-evaluator/evaluator.py` l.85 and l.103;
  - located the P, manifest and round-8/9 text cited above.
- **LSP:** the challenger's check returned "No language servers configured for this project."
- **Effect of the peer audits on the recommendation:**
  - The experiment peer added R1-03's estimator and seed contract, R1-04, and the trial-L evidence behind the partial verdict on choice 3.
  - The challenger added R1-06.
  - Both peers independently raised R1-01, R1-02 and the boundary gaps. Severities are the council's calibration: R1-01 is blocking because the artefact cannot serve as the P l.181 rule-5 input without it.
- **No real data, repositories, model calls or secrets were touched.** No files were changed.

## Verdict

**7 issues open** (R1-01 blocking; R1-02, R1-03 and R1-04 major; R1-05, R1-06 and R1-07 minor). The draft is not accepted for its stage-3 role. Once fixed, it needs another scoped review round with zero issues. Before that round, the tests must be run in a session that has a shell, and the four file digests recomputed. This review confers no authority to run the locked remeasurement, OC, inventory or any action on real data.
