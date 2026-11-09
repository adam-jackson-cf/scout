# X1 stage-3 draft review — locked-size input (`stage3-drafts/locked-size-input/`) — round 5

Date: 2026-11-09. Reviewer: innovation council, scoped round 5. Two peers audited the draft independently and read-only. Neither saw the lead's leaning:
- `innovation-challenger` (X1GSizeReview5.LsiR5Challenger)
- `experiment-peer` (X1GSizeReview5.LsiR5Peer)

The lead wrote the synthesis.

**Basis:** X1 round 9 recommends option G (`reviews/round-9.md` l.244–247). This review covers the locked-size input change only as a stage-3 artefact. It grants no authority for preparation, inventory, locked remeasurement, OC or any action on real data.

## Reviewed files

- **Worktree:** `/Users/adamjackson/Projects/work/scout-x1g`, PR #21, head `66bfb1408d24f3587ff4afc02bc6afbb1b54fba9`.
- **Paths:** relative to `evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/`.
- **Digests:** Main recomputed these with `shasum -a 256` at that head. Files marked "unchanged" are byte-identical to the digests reviewed in round 4.

| File | SHA-256 | Status |
|---|---|---|
| `README.md` | `7b9070c90a9e356dc3a23a777bbe5ca80096018d80b45637a7d62659f636981c` | changed |
| `round3_evidence.py` | `3d67fefb4748effb3b317efe8c796e39b8b7b0295a7cf0aa4183cdd1ae4e2ad8` | changed |
| `test_size_input.py` | `04ce64997ca13f955b1ccdbab49677288ef39e8bc61373a39d5b40659568bece` | changed |
| `locked_commands.py` | `d086af364594f779bd3546f9fa018dda66c412cdf729a86a04f25f62c681542c` | unchanged |
| `population.py` | `a32e33cd8061940067d94b56cead9a85608a3a07c79b8a9b89c248482586f0c6` | unchanged |
| `size_input.py` | `579c5119a812f64ffaa574f4ec9d7f77d12b89f47d407511d6379ee7ce647cdc` | unchanged |
| `test_cli.py` | `d6e22f609f5caa529d5ec2380edf7209aff4d10f323aeae968b161854c1e979d` | unchanged |
| `reviews/round-4.md` | `0f33e54581c30efde196199445fdaa7409645ad7c63798870eca64277ce50dd9` | record |
| `reviews/round-4-response.md` | `4e9b80ed5bafb82644bb2e13a195cd6d77838a08c0e13305c7e7950d42a4274d` | record |
| `reviews/round-4-test-output.txt` | `ef5a7da7fce45b847c0a682ba609fbd85ed569db5d3c1c56508dbae510b2b0d4` | evidence |
| `reviews/round-4-cli-test-output.txt` | `0c470baaa070ac6c05f531cfb4ede3aa0c624112747d68fadbe10478df5975b4` | evidence |
| `reviews/round-4-check-output.json` | `652f147fd755a0c70891ad709841dd1d63ae337afce904808e5d64302e40eef0` | evidence |
| `reviews/round-4-check-output.txt` | `5f18689a3d129772ee46a7efe08bb964bba14d54219df9b6617965978db40bb2` | evidence |
| `reviews/round-4-cli-evidence/benchmark.json` | `62309dc7b1dd7f2e4709e4643973f0ce8efc7c3256dd545a93f51d197255d43c` | evidence |
| `…/calibration.json` | `5ad5b3f5c3158b0af0a2e9eb205a8f1321c3c0b7410dc61d400cdf3ab3385dce` | evidence |
| `…/cli-roundtrip.json` | `edf6f8464203e2a3f94d5c6eba126eb43adec6f681daec8b9eeb2273136cbdd6` | evidence |
| `…/fit.json` | `f0423de61d004ade3b373840a22ca68940b3818aac0b9d6d08a24704490a05c1` | evidence |
| `…/gate.json` | `427f16da7600308c884a00c22db6918eadc3221a64255ec6e5152383e8b78036` | evidence |
| `…/synthetic-input.json` | `6f8880093a21621ba32fdfc2df2441fa4c8c4e18a84082c1501a194ac23e33c6` | evidence |

Main supplied 8-character prefixes for the CLI evidence. Each full value above is the one listed in `reviews/round-4-response.md`. The `source_sha256` block in `round-4-check-output.json` names the current `locked_commands.py` (`d086af36…`), `test_size_input.py` (`04ce6499…`) and `round3_evidence.py` (`3d67fefb…`).

## Signed basis (unchanged)

| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` (P) | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `oc-manifest.json` (M) | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |

**Signed text relied on:**
- **P l.181:** rerun only after a recorded integrity failure; record the reason and preserve the evidence; every binding input is the maximum over all completed runs.
- **M l.10:** optimize=0.
- **R9-reading-1** (`round-9.md` l.322–329): a run that reaches its end is completed even if its closure check fails, and its values enter the maximum.
- **`round-9.md` l.341–344:** recording crashed or interrupted runs is not counted.

The sibling evaluator contract is unchanged: `final-evaluator/provenance.py` `maximum_completed_inputs` and `evaluator.py` `resource_fit`.

## Round-4 disposition

| ID | Fixed as required? | Evidence and catching test |
|---|---|---|
| R4-01 (major) | **Yes** | **README** (finalise-before-rerun section, after "Exact commands"):<br>• After a component exits non-zero with published closure-mismatch evidence, keep its output, finish the remaining components, and run `fit` on the original set. That fit publishes completed-invalid and is the mandatory first `--prior`.<br>• "Never rerun or replace a component before this fit exists."<br>• Separate exact command blocks cover a non-zero benchmark (finish calibration, then `fit`) and a non-zero calibration (`fit` immediately). The commands must not be chained with `&&`.<br>• The rerun command uses fresh paths with `--prior`. A second rerun repeats `--prior` for both fits, in order.<br>• Item 8 still says "No command here authorises a rerun", so there is no conflict with P l.181.<br>**Works with the unchanged producer:** `measurement_run` accepts a completed component flagged `"closure mismatch"` and flags the run (`locked_commands.py:229–237`). `finish` publishes, then raises (`:450–466`). The published fit meets `prior_measurement_runs`, and `main` assembles `prior_runs + [record]`.<br>**Benchmark recovery is feasible:** `calibration_mode` reads no benchmark output (both peers checked this).<br>**Test:** `test_size_input.py:480–511` `test_calibration_failure_finalised_before_rerun`.<br>1. The real `finish` publishes calibration RSS 12,000 with closure A→B, then raises.<br>2. The real `main` fit on the original set publishes completed-invalid, then raises.<br>3. A `main --prior` rerun is built inside the B patch, so its components really record closure B. It has RSS 4,000.<br>4. Asserted: `runs[0] == original`; the evaluator maximum `calibration_rss` is **12,000**; the classifications are [invalid-included, valid-included].<br>**Regressions this test catches [INFERENCE]:**<br>• dropping priors in `main`;<br>• refusing flagged components (`ValueError`, not the expected `RuntimeError`);<br>• raising before publication;<br>• excluding completed-invalid values from the maximum.<br>**Confound (peer):** the original fit is also flagged by A≠B provenance, so this test alone would not catch a mutant that drops the component-flag term. `test_prior_cli_preserves_maxima_and_reads_once` already catches that mutant, because there the flag is the only cause. |
| R4-02 (minor) | **Yes** | **Test:** `test_size_input.py:325–327` sets `population_measurements[1].fixed_key_evaluations[1]=10`. That is outside the first row and above every frame maximum (2 and 9; fixture l.705). It asserts exactly `10/3000`. The frame-dominant ten-literal assertion is kept.<br>**Mutants killed by the new assertion [INFERENCE]:** dropping the fixed-key list, taking the first row only, taking population 0 only, and `min` all give 9/3,000 or 2/3,000.<br>**Evidence script:** `round3_evidence.py:53–55` adds `fixed-key-dropped`. The site must occur exactly once, and only an `AssertionError` counts as a kill. The receipt `round-4-check-output.json` shows **6/6 killed** and binds the current source digests.<br>**Real-path relevance (checked with jq on the round-4 CLI evidence):**<br>• fixed-key maxima are [0.182577, 0.177164];<br>• frame maxima are [0.175326, 0.173325];<br>• `fit.json` gives `requested_base` 2.778949771689498e-6 = 0.182577/65,700.<br>The branch now under test is the one that decides the real value. |
| R4-03 (minor) | **Yes** | **Test:** `test_size_input.py:456–478` `test_agreeing_components_differ_from_fit_closure`. It goes through `main` (`run_fit`) with three unflagged components that agree on closure A and have no end snapshot. `provenance` is patched to B, so `current`, the start closure in `main` and the end closure in `finish` are all B. `finish` therefore cannot flag the run, and the non-zero exit can only come from `measurement_run`'s comparison with `current` (`:235–237`).<br>**Asserted:**<br>• completed-invalid at the top level and in the run;<br>• `preserved_evidence_sha256` equals the SHA of canonical `evidence_sha256`;<br>• the exact ten-literal dict;<br>• the `RuntimeError` / non-zero path.<br>**Mutants killed [INFERENCE]:** replacing `current` with `data["provenance"]`, comparing components only with each other, and comparing only start or only end closures. Each yields a valid run and no raise. |

## New issues

**None.** The lead and both peers tested these hypotheses and found no defect against signed text that changes a numeric outcome:

1. **The finalising fit cannot be run.** This happens when a pinned signed source changed during a component (`pin_sources` refuses at the next import), or when the `--input` file was replaced (`measurement_run` raises "size input mismatch").
   - The failure is closed: nothing is published, and the README forbids a rerun until the fit exists.
   - No maximum is lost. Recovery means restoring the file.
   - This is the input-recovery suggestion accepted in round 4; it is not new evidence. → Suggestion.
2. **A rerun component itself fails, and the operator later passes only the latest fit.**
   - README item 8 requires every prior fit.
   - Round 4 accepted that an omitted prior cannot be detected in a file-based design. → Suggestion.
3. **A controller change alters the rerun gate's inventory.** The prior is refused on the inventory digest (`:331–334`). This fails closed.
4. **The R4-01 test passes for an unintended reason.** Refuted. The component-flag term is pinned by the existing prior-CLI test, and the rerun fixture is built under B, so it is a true B rerun.

## Non-blocking suggestions

None of these is counted.

- **Restore before finalising** (peer; carried from round 4). In the recovery section, add one sentence:
  - restore any pinned file that a refusal names;
  - rebuild `--input` from `provenance.size_input_record` and check its digest before running the remaining components.
- **Nested failure** (peer). State that finalising a failed *rerun* must pass every earlier fit as `--prior`, oldest first.
- **Fixture realism** (peer). In the R4-03 and R3-01 tests, B adds a closure key instead of changing an existing path's digest. A subset-comparison mutant would survive; it is contrived.
- **Branch placement** (peer). The fixed-key value 10 is in the last row of the last population, so a `[-1]` mutant would survive. That mutant is contrived. An interior variant is optional.
- **Stale paragraph.** The README's "Observed round-3 validation" paragraph (30 tests, 5 mutants) is accurate only for round 3. A round-4 line (32 tests, 6/6 mutants) would avoid confusion.
- **Workflow wrapper** (challenger). A wrapper that finalises the original set automatically after a published component failure would reduce dependence on operator sequencing.
- **Carried forward unchanged** from the round-4 suggestions list:
  - publication is not atomic;
  - several `validate_benchmark` branches are untested;
  - a race remains between `pin_sources` and use;
  - a `host_load()` failure inside `finish` happens before publication;
  - `synthetic-input` mode runs before the `-O` guard;
  - priors accept causes that no producer emits;
  - component reuse is not refused;
  - `--prior` outside `fit` is ignored;
  - interrupted runs have no recording path (accepted at `round-9.md` l.341–344).

## Verification

- **Main's runs** at head `66bfb140`, on a clean tree, with pinned `.tools/venv/bin/python` 3.11.11 and `PYTHONDONTWRITEBYTECODE=1`:

  | Suite | Tests | Time | Result |
  |---|---|---|---|
  | `locked-size-input/test_size_input.py` | 32 | 593.805 s | OK |
  | `locked-size-input/test_cli.py` | 1 | 602.449 s | OK |
  | `oc-controller/test_controller.py` | 35 | 1.366 s | OK |

  Main recomputed every digest above. The writer's retained logs agree: 32 tests in 586.982 s and 1 test in 602.705 s.
- **What this session did not do:** the council session and both peers ran no tests and no mutants and recomputed no digests. Every claim about mutant kills, regressions or failure paths is static [INFERENCE]. The exception is the 6/6 kill record in the retained `round-4-check-output.json`, which binds the current sources. `round3_evidence.py` was not rerun at this head.
- **Static and jq checks made:**
  - `locked_commands.py` `measurement_run`, `prior_measurement_runs`, `main` and `finish` against the README recovery procedure, P l.181 and R9-reading-1;
  - the three new or changed test bodies and `measurement_fixture`;
  - the `round3_evidence.py` mutant table and harness;
  - jq on the round-4 CLI evidence: fixed-key maxima [0.182577, 0.177164], frame maxima [0.175326, 0.173325], and a fit with `requested_base` 2.778949771689498e-6, `integrity_failure` null, optimize 0 and version 3.11.11.
- **How the peers changed the recommendation:**
  - **Lead:** found no issues independently. The challenger agreed on all three fixes.
  - **Experiment peer:** found the R4-01 confound and showed that an existing test covers it, which confirmed the R4-01 verdict.
  - **Experiment peer:** found the two finalisation-infeasible paths. The lead classified them as fail-closed and as the round-4 input suggestion carried forward, so they are suggestions, not issues.
  - **Experiment peer:** found the branch-placement and closure-realism residuals, both classed as contrived.
  - **Challenger:** contributed the wrapper alternative, recorded as a suggestion.
  - Neither peer changed the verdict.
- **Nothing else was touched:** no real data, repositories, model calls or secrets, and no files were changed.

## Verdict

**zero issues — draft accepted for its stage-3 role, conferring no run authority.**

**Sign-off:** the innovation council signs off the locked-size input draft at PR #21 head `66bfb1408d24f3587ff4afc02bc6afbb1b54fba9` with zero open issues. R4-01, R4-02 and R4-03 are fixed as round 4 required, each with a catching test.

**Reviewed sources and tests:**

| File | SHA-256 |
|---|---|
| `README.md` | `7b9070c90a9e356dc3a23a777bbe5ca80096018d80b45637a7d62659f636981c` |
| `locked_commands.py` | `d086af364594f779bd3546f9fa018dda66c412cdf729a86a04f25f62c681542c` |
| `population.py` | `a32e33cd8061940067d94b56cead9a85608a3a07c79b8a9b89c248482586f0c6` |
| `size_input.py` | `579c5119a812f64ffaa574f4ec9d7f77d12b89f47d407511d6379ee7ce647cdc` |
| `round3_evidence.py` | `3d67fefb4748effb3b317efe8c796e39b8b7b0295a7cf0aa4183cdd1ae4e2ad8` |
| `test_size_input.py` | `04ce64997ca13f955b1ccdbab49677288ef39e8bc61373a39d5b40659568bece` |
| `test_cli.py` | `d6e22f609f5caa529d5ec2380edf7209aff4d10f323aeae968b161854c1e979d` |

**Records and evidence:**

| File | SHA-256 |
|---|---|
| `reviews/round-4.md` | `0f33e54581c30efde196199445fdaa7409645ad7c63798870eca64277ce50dd9` |
| `reviews/round-4-response.md` | `4e9b80ed5bafb82644bb2e13a195cd6d77838a08c0e13305c7e7950d42a4274d` |
| `reviews/round-4-test-output.txt` | `ef5a7da7fce45b847c0a682ba609fbd85ed569db5d3c1c56508dbae510b2b0d4` |
| `reviews/round-4-cli-test-output.txt` | `0c470baaa070ac6c05f531cfb4ede3aa0c624112747d68fadbe10478df5975b4` |
| `reviews/round-4-check-output.json` | `652f147fd755a0c70891ad709841dd1d63ae337afce904808e5d64302e40eef0` |
| `reviews/round-4-check-output.txt` | `5f18689a3d129772ee46a7efe08bb964bba14d54219df9b6617965978db40bb2` |
| `reviews/round-4-cli-evidence/benchmark.json` | `62309dc7b1dd7f2e4709e4643973f0ce8efc7c3256dd545a93f51d197255d43c` |
| `…/calibration.json` | `5ad5b3f5c3158b0af0a2e9eb205a8f1321c3c0b7410dc61d400cdf3ab3385dce` |
| `…/cli-roundtrip.json` | `edf6f8464203e2a3f94d5c6eba126eb43adec6f681daec8b9eeb2273136cbdd6` |
| `…/fit.json` | `f0423de61d004ade3b373840a22ca68940b3818aac0b9d6d08a24704490a05c1` |
| `…/gate.json` | `427f16da7600308c884a00c22db6918eadc3221a64255ec6e5152383e8b78036` |
| `…/synthetic-input.json` | `6f8880093a21621ba32fdfc2df2441fa4c8c4e18a84082c1501a194ac23e33c6` |

**Signed basis:** `PREREGISTRATION.md` `abd206c3…6fdb`, `README.md` `22e2f368…c3202c3`, `oc-manifest.json` `1caea1a5…96761d`, `reviews/rev9-digests.json` `2253911b…9c2da` and `reviews/round-9.md` `6f1d6ca2…c6cc`, all unchanged.

Any change to a signed file above needs a new scoped council round with zero issues. This sign-off grants no authority for preparation, inventory, locked remeasurement, OC or any action on real data.
