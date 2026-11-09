# X1 stage-3 draft review — locked-size input (`stage3-drafts/locked-size-input/`) — round 4

Date: 2026-11-09. Reviewer: innovation council, scoped round 4. Two peers audited independently, read-only, and neither saw the lead's leaning: `innovation-challenger` (X1GSizeReview4.LsiR4Challenger) and `experiment-peer` (X1GSizeReview4.LsiR4Peer). The lead wrote the synthesis.

**Basis:** X1 round 9 recommends option G (`reviews/round-9.md` l.244–247). This review covers the locked-size input change only as a stage-3 artefact. It grants no authority for preparation, inventory, locked remeasurement, OC or any action on real data.

## Reviewed files

- **Worktree:** `/Users/adamjackson/Projects/work/scout-x1g`, PR #21, head `fcdfda5b566d89fe193b88bedcfe36419c902236`.
- **Digests:** Main recomputed these with `shasum -a 256` at that head.
- **Paths:** relative to `evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/`.
- **CLI evidence:** Main supplied 8-character prefixes for these files. Each matches the full value in `reviews/round-3-response.md`.

| File | SHA-256 |
|---|---|
| `README.md` | `92666d5a955a9906b8cbdf21be0b659b7a7feb3b2c3cbe33391b8395e8d9fed5` |
| `locked_commands.py` | `d086af364594f779bd3546f9fa018dda66c412cdf729a86a04f25f62c681542c` |
| `population.py` | `a32e33cd8061940067d94b56cead9a85608a3a07c79b8a9b89c248482586f0c6` (unchanged since round 1) |
| `size_input.py` | `579c5119a812f64ffaa574f4ec9d7f77d12b89f47d407511d6379ee7ce647cdc` (unchanged since round 3) |
| `round3_evidence.py` | `8ac46bba81686316bda40be38555bce8803cc79d935ba18d5c9c04997da2a300` |
| `test_size_input.py` | `c7da0a766bad346165af21c1e0cca1c62f0fefa816c59543774f23835f3ad154` |
| `test_cli.py` | `d6e22f609f5caa529d5ec2380edf7209aff4d10f323aeae968b161854c1e979d` (unchanged since round 2) |
| `reviews/round-3.md` | `e7a78488ad6e415b5dd151e0c37ad8558121114cd4153e70bb3651d8bd0bada1` |
| `reviews/round-3-response.md` | `812ced7d8cb7da492c49f1d8a13a0cb759047dfbffcf91d20a29cdedd23883ff` |
| `reviews/round-3-test-output.txt` | `09a48a081908429ad544d9ee0f91cd3db538d9b8413aa22a32a6ef09e3649634` |
| `reviews/round-3-cli-test-output.txt` | `0172772554128669568689e0c1011ed08c528846bb37408e2524b3eb3b1a9a82` |
| `reviews/round-3-check-output.json` | `3adcc0e4e1b70a32f70f5ed81af08cd20785099ec36379c86d01ae0cdccefa1d` |
| `reviews/round-3-check-output.txt` | `c7143a7940795358a6fd6548f3260fc0832c6dd58268f94f736e0bf2038e489f` |
| `reviews/round-3-cli-evidence/benchmark.json` | `00ad9c731be589c5c0da73abb461221e75fcefb9cea7ee176f9ce7f7929b6103` |
| `…/calibration.json` | `60ebee3592e37c0f5a7a7e070e09d7c03ef80f9b460cf522bd9e235d7988ba08` |
| `…/cli-roundtrip.json` | `ee0a505980474285e2251cc7b2f23a8ce8a6b4239d69678449441091ec7e0bbc` |
| `…/fit.json` | `9b8571e4e4331cad00aa8e5bf5a8b64cde97366944193dce8222ecfadbf590e0` |
| `…/gate.json` | `15a0f87ae2eda92fb544e8a60566c8a360a716cf17cdaee2ae3748dd7b1a9a3b` |
| `…/synthetic-input.json` | `6f8880093a21621ba32fdfc2df2441fa4c8c4e18a84082c1501a194ac23e33c6` |

## Signed basis (unchanged)

| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` (P) | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `oc-manifest.json` (M) | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |

**Signed text relied on:**
- **P l.181:**
  - "Rerun only after a recorded integrity failure … a closure mismatch …"
  - "record the reason and preserve that evidence"
  - "Report every completed run; each binding input is the maximum over all completed runs"
- **M l.10:** "sys.flags.optimize=0 in every new output; non-zero optimization refused".
- **R9-reading-1** (`round-9.md` l.322, 326–329): a run that reached its end is a completed run even if its closure check fails, and its values enter the maximum. The reading was recorded "so that no operator choice arises later".
- **`round-9.md` l.341–344:** crash and interrupted-run recording is not counted.
- **Sibling contract:** `final-evaluator/provenance.py` `maximum_completed_inputs` (l.90–125) and `evaluator.py` `resource_fit` (l.65–90).

## Round-3 disposition

| ID | Fixed as required? | Evidence and catching test |
|---|---|---|
| R3-01 | **Yes**, with a coverage residual (R4-03) | **Code:** `locked_commands.py:227–237` flags a run when a component's own flag, start closure or end closure differs from the fit-time closure. `finish` `:437–466` publishes create-once, then raises. Size-input mismatch (`:230–232`), host mismatch (`:238`) and interpreter mismatch (`:240–245`) still refuse, as round 3 required.<br>**Test:** `test_cross_component_closures_publish_then_nonzero` (`test_size_input.py:426–451`) covers gate A, benchmark flagged A→B, calibration and fit B, and a between-commands variant. It asserts the published file, the classification, ten keys, the preserved digest and a non-zero exit. The old raise-before-publish behaviour fails this test.<br>**Real path:** the real CLI `fit.json` shows a valid run and start closure equal to end closure, so the comparisons hold on actual JSON round trips. |
| R3-02 | **Yes** for the interface round 3 specified. The workflow residual is R4-01. | **Code:** `--prior` is repeatable (`:388`). Each prior is read once and hashed (`read_evidence`). Prior digests join `evidence_sha256` before the run ID is computed. `prior_measurement_runs` (`:320–370`) checks:<br>• the same input, inventory, host and runtime;<br>• a cumulative prefix in order;<br>• the exact qualifying cause and preserved evidence.<br>The payload is the prior runs plus the new run (`:431`).<br>**Tests:**<br>• `:453–482` runs through `main`: one read, exact prior run retained, digest bound, evaluator maximum **9,000** (not 3,000), both classifications.<br>• `:484–524` refuses an unflagged prior, an inventory mismatch, a host mismatch, an input mismatch, an optimize mismatch, a bad completed type, a bad cause, missing evidence, and the orderings `[second]`, `[second, prior]` and `[prior, prior]`.<br>• `:562–582` shows a failure at fit end marks only the new run.<br>README item 8 makes every prior fit mandatory. |
| R3-03 | **Yes** | **Code:** `provenance(…, tolerate_missing=True)` (`:116–134`) records a `null` digest plus `source_errors`. `finish` records `input_end_sha256: null` plus `input_end_error` and classifies either as a closure mismatch. The start snapshot stays strict. No hash is invented.<br>**Test:** `:526–560` deletes a disposable closure file and, separately, the input, and also injects `PermissionError`. It asserts publication, the classification, the null digest and the error. A strict end read would fail it. Create-once is covered at `:361`. |
| R3-04 | **Yes**, as prescribed. A residual comes from the prescription itself (R4-02). | **Fixture** (`:613–648`, checked by hand):<br>• every list maximum lies outside its first row;<br>• the complete-trial maximum (.03) is in workers-1 cell 4 and the full512 maximum (.07) in workers-2 cell 7;<br>• calibration frame CPU (9) exceeds every fixed-key value (≤ 4);<br>• gate CPU is 25,000/100,000.<br>**Assertions:** `:313–323` assert all ten literals exactly (6/3,000, 9/3,000, 12/3,000 and 25,000/100,000 are each the double nearest the literal).<br>**Mutants:** `round3_evidence.py:41–72` uses five single-site mutants, and only an AssertionError counts as a kill. The receipt `round-3-check-output.json` shows 5/5 killed and binds the current `locked_commands.py`, `test_size_input.py` and `round3_evidence.py` digests. |
| R3-05 | **Yes** | **Code:** `:243–245` refuses `python_optimize` that is not the integer 0 (a bool is refused) and any version other than 3.11.11. Prior runs are pinned through runtime equality. `main` `:398–399` refuses before any measurement.<br>**Tests:**<br>• `:584–598` sets optimize 1 or version 3.11.10 *equally* on all three components, so the new guard alone must refuse.<br>• `:600–610` runs a real `-O … gate` subprocess and asserts a non-zero exit, the exact message and no output. Deleting `or sys.flags.optimize` would publish after 100,000 draws and fail the test. |

## New issues

| ID | Severity | Location | Concrete failure | Required fix |
|---|---|---|---|---|
| R4-01 | **major** (R2-02/R3-02 residual) | README items 7–8 (l.37–38) and the "Exact commands" rerun paragraph; `main` `:417–432` (priors must be *fits*) | **The gap:** `--prior` accepts only published fits. A component that exits non-zero after publishing `"closure mismatch"` evidence (R3-01 and R3-03 paths) has no fit yet.<br>**Input:**<br>1. Gate and benchmark complete under A; benchmark construction RSS is 9,000 B.<br>2. Calibration completes, a closure file changes A→B, and calibration publishes completed-invalid and exits non-zero. That evidence is itself a recorded integrity failure.<br>3. The operator reruns the components into fresh paths under B (RSS 3,000 B) and runs `fit` on them. No prior fit exists, so nothing can be supplied.<br>**Wrong outcome:** `runs = [rerun]`. `maximum_completed_inputs` accepts a lone valid first run, so the binding maximum is 3,000 B. The completed run's values are dropped, contrary to P l.181 and R9-reading-1. Up to 7 of the 10 inputs come from the benchmark alone.<br>**Why it is ordinary:** the README never says that the original components must be finalised as a fit before any rerun. This is the same kind of undocumented operator step as R3-01. | **README:** prescribe finalise-before-rerun.<br>• After any component exits with published closure-mismatch evidence, keep its output, run the remaining components, and run `fit` on the original component set. That fit publishes completed-invalid.<br>• That failed fit is the mandatory first `--prior` of any rerun.<br>• Never rerun or replace a component before that fit exists.<br>• Give exact recovery commands for a non-zero benchmark and a non-zero calibration.<br>**Test:** with fabricated records:<br>1. a calibration flagged at its end (calibration RSS larger than in the rerun);<br>2. continue to `fit`, which publishes completed-invalid and exits non-zero;<br>3. rerun `fit` with `--prior` through `main`;<br>4. assert that the evaluator maximum keeps the original calibration RSS. |
| R4-02 | minor (R3-04 residual) | `locked_commands.py:256–257` (`reevaluate`); fixture `test_size_input.py:613–648`; `round3_evidence.py` | **The gap:** round 3 prescribed "calibration frame CPU > fixed-key CPU". The fixed-key branch of `requested_base_cpu_per_unit` therefore never decides the maximum in any test.<br>**Mutants that survive [INFERENCE]:** dropping the fixed-key list, taking its first row, or using `min` all still give 9/3,000.<br>**In the real CLI evidence the binding value comes from exactly this branch:**<br>• fixed-key maxima are [0.190001, 0.197569] (`benchmark.json`);<br>• frame maxima are [0.182216, 0.172808] (`calibration.json`);<br>• `fit.json` gives 3.0071385e-6 = 0.197569/65,700.<br>**Wrong outcome:** a mutant that drops the fixed-key term gives 0.182216/65,700, which is 7.8 % low. `requested_base` enters `resource_fit` linearly as `requested·populations·units` (`evaluator.py:77–78`). The understatement can therefore flip a fit near the 2,880 cap, and no test fails [INFERENCE on the magnitude]. | **Test:** add a fixture variant or subtest in which a fixed-key evaluation outside the first row (for example 10) exceeds every frame maximum, and assert `requested_base_cpu_per_unit == 10/3,000`.<br>**Evidence script:** add a matching `fixed-key-dropped` mutant to `round3_evidence.py` and record the kill. |
| R4-03 | minor (R3-01 residual) | `locked_commands.py:235–237`; `test_size_input.py:426–451` | **The gap:** round 3's required behaviour includes components that differ "from the fit-time closure". Every test case has disagreement among the components themselves, or a flag on the benchmark. No test has three components that agree with each other while the fit's closure differs.<br>**Mutant that survives [INFERENCE]:** replace `current` with `data["provenance"]`, or compare the components only with each other.<br>**Input:** gate, benchmark and calibration all complete unflagged under A. `controller.py` (which this PR has already edited) changes to B, and then `fit` runs under B.<br>**Wrong outcome:** under that mutant the run is classified valid. The fit binds closure B to values measured under A, and the closure-mismatch rerun that P l.181 permits never arises. | **Test:** add an A/A/A→B case through `main`. Assert:<br>• completed-invalid publication;<br>• `preserved_evidence_sha256`;<br>• all ten binding inputs unchanged against the literal answers;<br>• a non-zero exit. |

**Choices posed to the owner:** none.
- The README and the response recommend concrete behaviour throughout. "No owner choices are deferred" is stated, so R-064 is met.
- Requiring every cumulative prior fit is fail-closed.
- Refusing a prior with no failure flag matches "a failed fit never permits a rerun".

## Non-blocking suggestions

None of these is counted. Each needs tampered evidence, an operator departing from the documented fresh-path workflow, a contrived mutant, or a risk already accepted.

- **Three-fit assembly through `main`** (challenger). Cumulative validation of two priors is tested at function level, and `main` concatenates `prior_runs + [record]`, which the one-prior test asserts. A full RSS sequence of 9,000 → 12,000 → 3,000 through `main` would also pin `main`'s concatenation for n ≥ 2.
- **Narrow accepted prior causes to what the producer emits** (peer; round-3 F2 carried forward). `prior_measurement_runs` accepts `"non-resource crash"`, `"independent OS malfunction"` and `completed: false`, which no producer emits. Restrict priors to `"closure mismatch"` with `completed is True`.
- **Refuse component reuse in a rerun** (peer). A rerun can point back to an old, unflagged component file. Refuse any component whose digest already appears in a prior's `evidence_sha256`. Reuse cannot lower a measured value.
- **Untested tamper-only prior branches:** run-level host or runtime mismatch while the top level matches, a duplicate run ID, missing load fields, edited `locked_inventory` content with the stated digest left in place, and the OS-evidence digest.
- **Leaving out the newest prior cannot be detected** in a file-based design. README item 8 forbids it. An append-only run journal would enforce it mechanically.
- **Document recovery of `--input`** after the R3-03 deletion path: rebuild it from `provenance.size_input_record` and check its digest. This fits naturally into the R4-01 README text.
- **The R3-01 test checks only one value.** It asserts ten keys and `construction_rss`; compare against the full literal answer dict instead (R4-03's test can do this).
- **Reject `--prior` outside `fit`.** Other modes silently ignore it.
- **Carried forward unchanged:**
  - Publication is not atomic.
  - The `trials < 2000`, `replicates` and byte-equality branches of `validate_benchmark` are untested.
  - The check/use race after `pin_sources` remains.
  - A `host_load()` failure inside `finish` happens before publication.
  - `synthetic-input` mode runs before the `-O` guard. It is fixture-only and predates this round.
  - Interrupted runs have no recording path; this was accepted at `round-9.md` l.341–344.

## Verification

- **Main's runs** at head `fcdfda5b`, on a clean tree, with pinned `.tools/venv/bin/python` 3.11.11 and `PYTHONDONTWRITEBYTECODE=1`:

  | Suite | Tests | Time | Result |
  |---|---|---|---|
  | `locked-size-input/test_size_input.py` | 30 | 592.176 s | OK |
  | `locked-size-input/test_cli.py` | 1 | 600.320 s | OK |
  | `oc-controller/test_controller.py` | 35 | 3.400 s | OK |

  Main recomputed every digest above.
- **What this session did not do:** the council session and both peers had no shell. They ran no tests and no mutants and recomputed no digests. Every mutant-survival and failure-path claim is static [INFERENCE], except the jq checks below. `round3_evidence.py` was not rerun at this head; its retained receipt binds the current source digests.
- **Static and jq checks made:**
  - all of `locked_commands.py` against the R3 required fixes, P l.181, M l.10, R9-reading-1 and `provenance.py` l.90–125;
  - the hand-checked fixture maxima and literal answers;
  - real CLI fixed-key maxima of [0.190001, 0.197569] against frame maxima of [0.182216, 0.172808], with units 65,700 and `requested_base` 3.0071385e-6 (R4-02);
  - the real CLI fit: run valid, optimize 0, version 3.11.11, start closure equal to end closure.
- **How the peers changed the recommendation:**
  - **Lead:** found R4-01 independently.
    - The challenger confirmed it with a concrete calibration-failure sequence.
    - The experiment peer did not raise it. Its suggestion on component reuse is a related but distinct path.
    - The lead rated it major, consistent with R3-01: completed values are lost through an undocumented operator step.
  - **Experiment peer:** found R4-02. The lead confirmed it on the real CLI evidence and counted it as an untested decision boundary that decides a real binding value. Round 3's own fixture prescription caused it.
  - **Challenger:** found R4-03. The lead counted it as an untested boundary of behaviour round 3 named as required, consistent with R3-05. The lead moved the challenger's three-fit assembly finding to suggestions, because the cumulative logic is tested at function level and the surviving mutant is a contrived truncation.
- **Nothing else was touched:** no real data, repositories, model calls or secrets, and no files were changed.

## Verdict

**3 issues open:**
- **Major:** R4-01.
- **Minor:** R4-02 and R4-03.

R3-01 to R3-05 are fixed as round 3 required, each with a catching test. The residuals of R3-01, R3-02 and R3-04 are counted as R4-03, R4-01 and R4-02. The draft is not accepted for its stage-3 role. After the fixes it needs another scoped round with zero issues, with tests rerun and digests recomputed. This review confers no authority to run the locked remeasurement, OC, inventory or any action on real data.
