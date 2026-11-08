# X1 stage-3 draft review — draft final evaluator (`stage3-drafts/final-evaluator/`) — round 2

## Scope and authority

- This is a scoped innovation-council review, round 2, of one Option G stage-3 draft. It checks the fixes for the round-1 issues in `reviews/round-1.md`, as described in the author's `reviews/round-1-response.md`.
- Worktree: `/Users/adamjackson/Projects/work/scout-x1g`. PR #21 head: `ed7f21ce2bc58f03c4dee6e0eeaa7092ab9dbc4f`.
- **This review gives no authority to run anything.** That includes the OC, the inventory, model calls, scans and anything on real repositories.
- How the round was done:
  - The lead read the files and the signed text.
  - An `innovation-challenger` (adversarial pass) and an `experiment-peer` (competing readings) worked in parallel. Neither had a shell.
  - The lead checked every finding below against the files.

## Reviewed files (SHA-256 recomputed by Main with `shasum -a 256` at the head above)

| File | SHA-256 |
|---|---|
| `README.md` | `71c4f690e7840a8c07e3974525dc95cd8884be04847cf31a634d00dd89bd2979` |
| `check_frozen.py` | `8130568a8e3e49fdf420c110bc9825ad20f483237f1f8745a2d96fa4009a56a1` |
| `evaluator.py` | `aa04fb87fbe6ba41b3790dee6b388160e1814daeb74e93ec3825339e3da7a2b6` |
| `provenance.py` | `a691afbbc22b83f4e5c6af0ca7b0f667e1be3a7540cca6c53936ffa0f9d285ab` |
| `test_evaluator.py` | `fe7caf1a76138510cbb79db068de15d6e2889ebc67d6c2943f9e946943774675` |
| `reviews/round-1.md` | `d522104c10699ae05bf30514e1535feb556eddfc77802f3dd6d8ecfc48206388` |
| `reviews/round-1-response.md` | `9f708f5b07e052690c44372fc7e271d0da737b71698b2626ee88d83cfd3ceb62` |
| `reviews/round-1-test-output.txt` | `351158c1b87b040c9ad36615984ceb927190ebe5ebe5316bf9aedf131844fe13` |
| `reviews/round-1-mutant-output.txt` | `e37918a439014b1f6d8cecca2a329a82bdd5e15de593cb3dbf8f60673e7d961b` |
| `reviews/round-1-frozen-output.json` | `2c5d8a155e57cdd60a2b6ea295e9c0c1a2453acccf0ce662a60e711a85df50d0` |

## Signed basis (unchanged)

| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| X1 `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `oc-manifest.json` | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |

Signed text applied:
- P §7 l.127–135.
- P §8 l.163, l.169, l.175–177, l.181 and l.185.
- P §10 l.206.
- P §11 l.220–228.
- Manifest `final_evaluator_contract` (l.196–200), `row_timing` (l.150–153) and `populations` (l.37–55).
- The rev8 method in `validation/reproduce_reports.py:205-330`.

`check_frozen.py` is byte-identical to round 1.

## Round-1 issue disposition

Paths are relative to the artefact folder. Test names are methods of `EvaluatorTests` in `test_evaluator.py`.

| ID | Fixed | Evidence (code; catching test) |
|---|---|---|
| R1-01 | **Yes** | **Code:**<br>• `provenance.py:41-49` checks exact bundle membership and every member's digest, and pins the manifest bytes to `1caea1a5…`.<br>• `evaluator.py:108-109`: the manifest argument must equal the bundle member.<br>• `evaluator.py:113-123`: `p_gate` is `gate_successes/100000` with 100,000 redraws, and the sizes come from `locked_inventory`.<br>• `evaluator.py:126-134`: the OC output carries `records_sha256` and references the inventory and manifest digests; the measurement runs are bound to the inventory.<br>• `evaluator.py:204-210`: the M̂ evidence is bound to the pre-dev digest and the inventory.<br>• No parallel arguments exist.<br>**Tests:** `test_byte_bundle_substitutions_and_no_parallel_inputs` (:104); `test_seal_internal_references_and_inventory_types` (:131); the M̂ swap in `test_predev_none_no_bypass_and_postdev_tampering` (:288-293).<br>**Gap:** the signed-manifest pin itself has no isolating test (R2-02). |
| R1-02 | **Yes** | **Code:**<br>• `provenance.py:62-97` takes the maximum of each input over every completed run.<br>• `evaluator.py:59-88` matches term by term the binding expression in `reproduce_reports.py:229-237,244-247`: construction 18,172×5×U; requested-base 16,128×5×U; (U×calibration + enumeration)×11,900×2044/2380; 5×4,000×(17,920×complete + 252×full512); gate 18,172×5×`ceil(4000/p_gate)`; ×2 once; wall = hours/192.<br>**Test:** `test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary` (:211). Exactly 2,880 passes and ×(1+1e−10) fails; changing only `p_gate` (to .49999) flips the fit; maxima from two runs are combined per input; a hand-computed rev8 expression matches.<br>**Gap:** nothing tests that a failed fit makes the rows fail (R2-01). |
| R1-03 | **Yes** | **Code:** `provenance.py:70` compares every remeasurement host with the sealed OC host (`evaluator.py:135-136`).<br>**Test:** `test_memory_boundary_and_bound_host` (:244-246) changes each of the four OC host fields alone and expects a refusal.<br>**Gap:** a host that omits fields is never tested (R2-04). |
| R1-04 | **Yes** | **Code:** `evaluator.py:40-56`. Steps 1–2 come before the interval check; an unmeasured or reversed interval is inadequate; flags must be exact `bool`.<br>**Test:** `test_ordered_typed_verdict_and_interval_boundaries` (:295) covers NaN after an integrity failure, NaN after a violating case, NaN alone, upper = .10 and upper just below .10.<br>**Gap:** step 6 has a new untested boundary (R2-05). |
| R1-05 | **Partly** (code yes; one required test does not reach the code) | **Code:**<br>• `is True` predicates at `evaluator.py:167,174`.<br>• Booleans are excluded from gains at `evaluator.py:165,177`.<br>• `p_gate` and completion types are checked at `evaluator.py:115-117` and `provenance.py:85-91`.<br>**Tests:** `test_reachability_gain_and_missing_mask_types` (:199); `test_seal_internal_references_and_inventory_types` (:144); `test_rerun_qualifications_evidence_and_measurement_types` (:253-267).<br>**Gap:** the non-finite gain tests raise inside the fixture's own `encode` before `evaluate` runs. Carried as R2-03. |
| R1-06 | **Yes** | **Code:** `evaluator.py:155-158`.<br>**Test:** `test_reconciled_counter_caps_and_eligible_boundary` (:163-169) uses the exact reconciled fixtures R1-06 asked for (4000/4001, 8000/8001 and n = 1999). The mutant output shows the execution-cap and attempt-cap mutants killed. |
| R1-07 | **Yes** | **Tests:**<br>• `test_each_tail_boundary_for_each_cell_kind` (:171): 313/314 for each tail and each of the three cell kinds.<br>• `test_reference_power_null_and_inadequacy_boundaries` (:182): 8079/8078, 457/458, 84/2084 vs 85/2085, and realised gain 0/.001 and .2/.199.<br>• Missing-cell targets and the mask (:199-206).<br>• Membership, an extra identity key and population 5 (:155-158).<br>• Manifest argument swap (:111-113). |
| R1-08 | **Yes** | **Tests:**<br>• Memory exactly at hw.memsize − 8 GiB passes and +1 fails (:239-243).<br>• Supplied resource and memory overrides are impossible (:247-251).<br>• Rerun after a valid run, duplicate run ID, trailing unresolved failure, interrupted run with no cause, missing failure or OS evidence, and non-qualifying causes (:253-269).<br>• M̂ at 399/400 and 699/700 (:90-93).<br>Each is in its named test. |
| R1-09 | **Yes** | **Code:**<br>• `evaluate` takes no M̂ and returns `not testable before dev` when no row passes (`evaluator.py:187`).<br>• `select_postdev` (:195-218) picks the largest registered M ≤ M̂ and then reads that row's sealed status.<br>**Test:** `test_predev_none_no_bypass_and_postdev_tampering` (:272-293) covers no rows passing, a tampered sealed row, no bypass of the failed 700 row, and refusal of the original records. |
| R1-10 | **Yes** | **Code:** `provenance.py:93-94`; `evaluator.py:189`.<br>**Tests:** `test_full_grid_predev_postdev_and_selection_boundaries` (:94-95); `test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary` (:233-235). Every run is reported with its inputs, cause and classification. |
| R1-11 | **Yes** | **Code:** `evaluator.py:9-27`. README l.29 says the generator must import `expected_cells` and `identity`.<br>**Test:** `test_identity_counts_and_known_source_pins` (:78-84) pins 16,128 / 1,792 / 252 cells, 90,860 records and identity-list SHA-256 `3499e423…`. |

## New issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R2-01 | major | `evaluator.py:146`; `test_evaluator.py:211-243` (esp. :233-236) | **A failed resource fit is never shown to make the rows fail.**<br>• Every CPU and memory boundary test calls `resource_fit` directly.<br>• The only `evaluate` call with a failing fit is at :233. Its gate term alone is 2×18,172×5×8,000×.02/3600 ≈ 8,076 core-hours. That test asserts only the run list and the maxima.<br>• So a mutant that sets `passed = {m: True …}` at :146 would survive [INFERENCE: from reading; not executed].<br>• P l.181 says a failed or unmeasured CPU or memory fit means not testable before dev. P l.175 and `row_timing.before_dev` make compute limits a row gate. This is an untested decision boundary of the kind round 1 rated major. | • At :233, assert status `not testable before dev` and every row `False`.<br>• Add an `evaluate` fixture where only memory fails (8×RSS = budget + 8) and one where CPU is exactly 2,880 (must pass).<br>• Run the `{m: True}` mutant and record it as killed. |
| R2-02 | minor | `provenance.py:47-48`; `test_evaluator.py:104-113` | **The signed-manifest digest pin has no isolating test.**<br>• Every manifest substitution test either leaves a member's digest stale or passes argument bytes that differ from the bundle member. Both are refused by other checks.<br>• `test_evaluator.py:83` checks the fixture file, not the evaluator.<br>• So deleting :47-48 alone would survive [INFERENCE: from reading; not executed].<br>• R1-01's required fix and P l.177 / manifest l.198 ("signed manifest identity") need this check to be caught. | Change the manifest bytes (for example, reserialise them). Supply the changed bytes as both the argument and the bundle member, update the bundle digest and the OC output's `signed_manifest_sha256`, and assert a refusal. |
| R2-03 | minor | `test_evaluator.py:152-153,208-209`; `evaluator.py:165,177-178`; `provenance.py:32` | **The non-finite gain tests never reach the evaluator.**<br>• The fixture's `encode` uses `allow_nan=False`, so it raises before `evaluate` runs.<br>• `decode` rejects only the `NaN` and `Infinity` literals. JSON `1e400` / `-1e400` decodes to ±inf [INFERENCE: standard Python `json` behaviour; not executed].<br>• `math.isfinite` at :177-178 is then the only guard for missing cells. Without it, inf ≥ target passes, and so does −inf ≤ 0 at target 0.<br>• R1-05 asked for one test per field; the manifest l.199 requires unmeasured input to fail closed. | Build OC payload bytes by hand with `1e400` and `-1e400` as `realised_gain`, for a reference cell and a missing cell at targets 0 and .2. Assert each cell fails. Record the `isfinite`-deletion mutant as killed. |
| R2-04 | minor | `provenance.py:52-56,70-72` | **The host-identity and load-provenance presence checks have no tests.**<br>• Neither `validate_host`'s field-set and type check nor the three-load-averages and concurrent-process-count check is exercised. Hosts are only changed to other valid values (:244-246), and every run carries valid load fields.<br>• If both the OC host and the run host omit `os_build`, or a run omits `load_averages`, deleting these checks would survive [INFERENCE: from reading; not executed].<br>• P l.181 requires recording hw.model, hw.memsize, logical CPU count, OS build, load averages and concurrent-process count. R1-03 required the OC output to carry all four host fields. | Add refusal fixtures for each of the following, with the OC host and the run host kept equal so equality cannot catch it:<br>• each host field missing;<br>• each host field with the wrong type, including bool `logical_cpus`;<br>• `load_averages` with 2 entries, or containing a negative or non-number value;<br>• `concurrent_process_count` of 0, bool or missing. |
| R2-05 | minor | `evaluator.py:48-49,54`; `test_evaluator.py:298-311` | **Step 6's point cutoff is untested on its failing side, and so is the reversed interval.**<br>• No test has point < .10 with lower > 0 and upper ≥ .10. So deleting `point >= .10` would survive [INFERENCE: from reading; not executed].<br>• That change would turn the rule into "lower > 0", which P l.132 and l.135 ("the point cutoff is not LB ≥ 0.10") rule out.<br>• The `lower > upper` branch is also never exercised. | Add `point = nextafter(.10, 0), lower = .001, upper = .4` → `inconclusive`, and a reversed interval (`lower = .3, upper = .2`) → `inconclusive`. |
| R2-06 | minor | `evaluator.py:204-210` (esp. :209); README l.21 | **The cost evidence after dev is accepted on truthiness.**<br>• The only check on `bootstrap_evidence` is that it is non-empty. `True` or `"false"` passes, so a correctly digested payload that declares no bootstrap or receipt completeness can select a row [INFERENCE: from the code path; not executed].<br>• P l.206 fixes c_U as the one-sided 95% upper bound from 10,000 per-run-cluster bootstrap means, seed 169300, with complete receipts. It also says unknown or incomplete evidence blocks test. Manifest l.199 says unmeasured input fails closed.<br>• This was the author's choice, made while fixing R1-01(d). It is the same truthiness defect class as R1-05. | **Recommended:** replace the truthiness test with an exact typed schema: `bootstrap_means == 10000`, `bootstrap_seed == 169300`, `resampling_unit == "run-cluster"`, `one_sided_level == .95`, and `unknown_or_incomplete_receipts == 0`. Use exact `int`/`str`/`float` types with no bool, and refuse otherwise. Add one refusal test per field, plus `bootstrap_evidence=True`.<br>Recomputing c_U from receipts stays with the later dev-gate stage. Authenticity of the anchor stays under round-1 Open choice 1. |

### Considered and not counted

- **Checking all four flag types before step 1** (`evaluator.py:41-42`). The peer suggested this reopens R1-04. It does not:
  - a non-bool flag is malformed input, not the legitimate unmeasured state that R1-04 was about;
  - raising emits no `support`;
  - round 1 already accepted exceptions as fail-closed.
- **The OC output and measurements are not tied to `controller_freeze` or `source_runtime_closure`.**
  - The manifest's `required_checks` (l.198) do not name the controller or closure. Round 1 listed the R1-01 binding items specifically.
  - Proving which producer made the output is authenticity of the trust anchors, deferred under round-1 Open choice 1.
  - Both members are still sealed into the result's `provenance_sha256`. This is a prerequisite for the authoritative variant.
- **Forging a pre-dev result by minting a new external `predev_sha256` anchor:** deferred under Open choice 1.
- **The evaluator sums seconds where rev8 sums per-term hours.**
  - The two are algebraically equal.
  - The 2,880 boundary test asserts exact equality and passes.
- **`ceil(4000/p_gate)` is computed in floating point:** it is the same expression as the signed rev8 code (`reproduce_reports.py:229`).
- **New choices that are only stricter:** size vectors of 20–2,000 runs, and positive-only measured inputs. Both are fail-closed.
- **Owner-facing open choices:** README.md and round-1-response.md pose none (R-064 satisfied).

## Verification

- **The lead and both peers had no shell.** No tests, mutants or hashes were run in this round.
- Main's independent run at head `ed7f21ce…` (clean tree, pinned `.tools/venv/bin/python` 3.11.11, `PYTHONDONTWRITEBYTECODE=1`): `final-evaluator/test_evaluator.py` ran 14 tests in 143.781 s, OK.
- The author's committed outputs were read but not reproduced:
  - 14 tests OK in 149.162 s;
  - 9/9 mutants killed;
  - frozen check: 126/131 literal matches, 5/5 groups valid, 131/131 unchanged against `e2b94e3c`, manifest digest `2253911b…`.
- The mutant probes do not include the R2-01 to R2-05 mutants. Those survival claims are marked [INFERENCE].
- Reproduce after the fixes:

```sh
cd /Users/adamjackson/Projects/work/scout-x1g
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/test_evaluator.py -v
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py
```

- Mutants to record as killed, each in a scratch copy:
  - `evaluator.py:146` `resources["passed"]` → `True`;
  - delete `provenance.py:47-48`;
  - delete `math.isfinite(realised) and` at `evaluator.py:177`;
  - make `validate_host` a no-op;
  - delete the load-provenance clauses at `provenance.py:70-72`;
  - delete `point >= .10 and` at `evaluator.py:54`;
  - delete `or lower > upper` at `evaluator.py:48`;
  - change `not evidence["bootstrap_evidence"]` to the new schema check, and its inverse.

## Recommendation, alternatives, risks and fix order

- **Recommendation:** keep this evaluator. Fix R2-01 to R2-06 in the same draft. All are test additions, except R2-06, which is a small schema check. Then resubmit for round 3.
- **Alternative for R2-06:** take c_U from a separately reviewed cost-computation artefact's sealed output instead of a typed schema. Rejected for now: it adds a second artefact and a second review. The typed schema closes the defect, and recomputation stays with the dev-gate stage.
- **Risks:**
  - The R2-06 field names become part of the dev-gate contract, and can change later only through a new council round.
  - All the other fixes are reversible test additions.
  - The authoritative variant (real trust anchors, controller and closure binding, `binding_gate_established`) still needs its own scoped zero-issue review.
- **Fix order:**
  1. R2-01.
  2. R2-02 and R2-04.
  3. R2-03 and R2-05.
  4. R2-06.
  5. Rerun the suite and the mutants above, and record the SHA-256s.

## How the peers changed the outcome

- **Challenger:**
  - added R2-06; the lead downgraded it to minor because it is the same class as R1-05;
  - added R2-02;
  - independently confirmed the rev8 term-by-term equivalence.
- **Peer:**
  - added R2-01, R2-03 and R2-05;
  - proposed the flag-ordering and controller-binding candidates; the lead did not count either, with the reasons given above.
- **Lead:** added R2-04.
- Neither peer found a regression in any round-1 fix.

## Verdict

**6 issues open** (1 major, 5 minor). The draft is not yet accepted for its stage-3 role. This review gives no authority to run anything.
