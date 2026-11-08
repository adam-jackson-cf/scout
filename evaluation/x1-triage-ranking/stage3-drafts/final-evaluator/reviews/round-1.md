# X1 stage-3 draft review — draft final evaluator (`stage3-drafts/final-evaluator/`) — round 1

## Scope and authority

- This is a scoped innovation-council review, round 1, of one Option G stage-3 draft. Round 9 recommended it (`reviews/round-9.md` l.244–247 and §11 step 5, l.436–438). The draft is a prerequisite for "a reviewed final evaluator" in round 9 l.451.
- Worktree: `/Users/adamjackson/Projects/work/scout-x1g`, branch `agent/x1-option-g`, head `7168c0e1b2cbfaacbd3ab81da0910eeaa09b531e`, PR #21.
- **This review gives no authority** to run the OC, the inventory, model calls, scans or anything on real repositories. Even a future zero-issue outcome on this draft would not authorise any of those. Only preparation actions (a)–(e) are authorised, under round 9 §12.
- How the round was done:
  - The lead read the files, ran `grep` and did the arithmetic by hand.
  - An `innovation-challenger` (adversarial pass) and an `experiment-peer` (competing readings) ran in parallel. Neither had a shell either.
  - Each finding below was checked by the lead against the files.

## Reviewed files (SHA-256 as given by Main; not recomputed here, so they rest on Main's report)

| File | SHA-256 |
|---|---|
| `README.md` | `67bc5aeee9fb1034f041aaa179a38c087eddc44c971e9873656f9aa8ada5767a` |
| `evaluator.py` | `1efc8328665aa393145a0c49a681ded0ad6d8bc40845ec41fd232b36d395e3d6` |
| `provenance.py` | `e4ca8f4e472b27be7860b5af62d2ae1950fbc50b6113c387345232844d0c6d27` |
| `test_evaluator.py` | `82b264993946b4f7f5cbcae8b3b65ba24be79932f7c7d5c84aa0f536bde2045f` |
| `check_frozen.py` | `8130568a8e3e49fdf420c110bc9825ad20f483237f1f8745a2d96fa4009a56a1` |

## Signed basis

| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` (`rev9-digests.json` l.406) |
| X1 `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` (l.407) |
| `oc-manifest.json` | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` (l.409) |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |

Signed text applied:
- P §7 l.127–139.
- §8 l.153–185, especially l.163 (counters), l.169 (references), l.175–177 (binding gate and final evaluator), l.181 (binding resource stop rule) and l.185 (how compute is charged).
- §10 l.206 (row selection).
- §11 l.220–228.
- The manifest's `final_evaluator_contract` and `row_timing`.
- R9-reading-1 and R9-reading-2 (round 9 l.322–333).

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R1-01 | major | `provenance.py:13-18`; `evaluator.py:94-104`; `test_evaluator.py:24,29` | The provenance bundle is never tied to the inputs being evaluated. `verify_bundle` only checks that each payload matches its supplied digest. Nothing then checks the evaluated objects against those payloads. **(a)** The manifest check hashes a re-serialised parsed object (`bfc7bb2d…`). That value appears nowhere in the signed record. The `signed_manifest` payload is never compared with the manifest being evaluated, and the signed byte digest `1caea1a5…` is never used. **(b)** `p_gate` sets the attempt cap, but it is a bare argument rather than a value taken from the locked inventory's 100,000 redraws (P l.163). **(c)** No digest of the records or the OC output is bound. **(d)** The measurement runs and `mhat` (from the c_U bootstrap, l.206) are not bound. The test fixture passes with payloads like `b"synthetic:signed_manifest"`. This falls short of P l.177 ("authentic locked-input and signed-manifest provenance") and l.228 ("all source and output hashes"). Authenticity of the trust anchors themselves is deferred under Open choice 1 and is **not** counted here. The binding logic between bundle and inputs is in scope. | `evaluate` must take the manifest **bytes**, check their SHA-256 against the signed `oc-manifest.json` digest, and parse those same bytes. `p_gate` and the locked sizes must come from the verified `locked_inventory` payload. The measurement runs, the sealed OC output (records digest) and the M̂/c_U evidence must be bundle members, and the evaluator must use them rather than parallel arguments. Add synthetic tests that keep the bundle and digests identical while swapping, one at a time, the manifest, records, `p_gate`, measurement runs, resources and M̂. Each swap must be refused. |
| R1-02 | major | `evaluator.py:58-69,97-102`; README l.18 (Open choice 3) | The resource fit does not implement the charging rule in P l.185 or the signed rev8 method (`validation/reproduce_reports.py:214-249,322-324`), which derives binding core-hours from measured per-operation maxima. Those maxima are: construction CPU per unit, calibration excluding enumeration per unit, worst-case enumeration per calibrated population, requested-base re-evaluation per unit, gate CPU per check × `ceil(4000/p_gate)`, and 4,000 × maximum trial CPU (complete or full512) per binding population. All are ×2, with the binding fraction (1792+252)/2380. The draft instead takes six caller-supplied totals. **(i)** `p_gate` never enters the CPU fit. **(ii)** The draft requires the totals to equal the maxima over remeasurement runs. That takes the larger of whole-run totals rather than the maximum of each binding input (l.181), which can under-charge. A locked remeasurement also cannot produce full-grid totals. **(iii)** `wall_days` is a free input "already charged", and it is expected among the remeasurement maxima. The remeasurement does not measure that quantity. rev8 derives wall days as binding core-hours/192 (8 workers × 24 h), so the ×2 overhead cannot be checked here. | Take the maximum of each measured input across every completed run (R9-reading-1 kept). Compute binding core-hours in the evaluator with the rev8 binding expression, using the locked units per population, manifest counts, inventory `p_gate`, 4,000 executions and ×2. Derive wall days as binding core-hours/192. Keep the fixed-8-worker and memory checks as they are. Tests: changing only `p_gate` flips the fit; maxima taken from different runs are combined per input; CPU passes at exactly 2,880 and fails just above it. |
| R1-03 | major | `provenance.py:39`; `evaluator.py:85-104`; README l.8 ("The final host cannot change") | R9-reading-2 is not enforced for the event it governs. Only the remeasurement runs are compared with `binding_host`. Neither the records nor the bundle identify which host ran the binding OC, so OC output from another host passes. | The sealed OC output must carry hw.model, hw.memsize, logical CPU count and OS build. The evaluator must refuse when these differ from the remeasurement host; that is a deviation that pauses the work (P §11). Test: measurements from host A plus valid OC output from host B must fail. |
| R1-04 | major | `evaluator.py:40-55`; `test_evaluator.py:92-104` | The §7 ordering is broken for unmeasured intervals. The check for finite values and lower ≤ upper runs **before** steps 1–2. So an integrity failure with a NaN interval raises an error instead of returning `inconclusive`, and a violating case with a NaN interval raises instead of returning `not supported`. An unmeasured interval with nothing higher-priority should be inadequate, giving `inconclusive` at step 3 (P l.115, l.131). The flags are also tested for truthiness, so `integrity="false"` counts as passing. The step-4 boundary `upper == .10` has no test. | Decide steps 1–2 before validating the interval. Treat an unmeasured interval as inadequate (step 3). Require exact `bool` flags. Add tests: integrity False with NaN gives `inconclusive`; violating with NaN gives `not supported`; adequate with NaN gives `inconclusive`; upper = .10 does not give `not supported`; upper just below .10 does. |
| R1-05 | minor | `evaluator.py:103,133,140,143`; `provenance.py:48` | Malformed values get past signed failure conditions because types are not checked: `reachable="false"` passes the reference reachability check; `mask_reachable="false"` passes the missing-cell conjunction; `realised_gain=True` satisfies ≥ .20; `p_gate=True` is accepted as 1; `completed="false"` is treated as completed. The manifest's `final_evaluator_contract` requires malformed or unmeasured input to fail closed. | Use `is True` for predicates and completion flags. Reject booleans and non-finite values for gains, `p_gate` and measurements. Add one mutant test per field. |
| R1-06 | major | `test_evaluator.py:51-57` | The counter mutants don't isolate the caps. `executions=4001`, `attempts=8001` and `eligible=1999` each also break the counter reconciliation at `evaluator.py:125` (inadequate = e−n and gate failures = a−e). Deleting `e<=4000`, `a<=ceil(4000/p_gate)` or `n==2000` alone would therefore still pass the suite [INFERENCE: from reading; not executed]. These are untested decision boundaries for P l.163 and l.177. | Use reconciled fixtures at each boundary with `p_gate=.5`. Passing: `e=4000, n=2000, inadequate=2000, a=4000`, and `a=8000, e=2000, gate_failures=6000`. Failing: `e=4001, inadequate=2001, a=4001`, and `a=8001, gate_failures=6001`. Also a reconciled `n=1999` with `inadequate=e−1999`. |
| R1-07 | major | `test_evaluator.py:38-66` | The binding statistical gates have no adverse or boundary fixtures. Every fixture has 100/10,000 misses, 9,500 support and `inadequate=0`. Untested: each tail's Wilson upper ≤ .035, power support lower ≥ .80, the null false-support boundary, per-population power inadequacy upper ≤ .05, null realised ≤ 0 (e.g. realised .001), the missing-cell `mask_reachable` and realised-target checks, extra identities or out-of-range population indices, and manifest substitution. P l.169, l.175 and l.177 apply. | Add last-pass and first-fail fixtures, computed with the module's `wilson` and asserted against the hand values [INFERENCE: hand arithmetic]: pooled tail misses 313 pass / 314 fail; power support 8,079 pass / 8,078 fail; null false support 457 pass / 458 fail; per-population inadequacy at its boundary (≈84/2,084 pass, 85/2,085 fail). Test each tail separately, and test coverage, reference and missing cells separately. |
| R1-08 | major | `test_evaluator.py:68-90` | Resource, rerun and selection boundaries are untested: resources ≠ maxima; `memory_bytes ≠ hw.memsize`; memory exactly at hw.memsize − 8 GiB (60,129,542,144 bytes); a rerun after a successful completed run ("one create-once run", l.181); host mismatch; duplicate run ID; missing failure or OS evidence; an interrupted run with no cause; a trailing unresolved failure; M̂ exactly 400 or 700. | Add one passing and one failing fixture for each. The CPU fixtures follow from R1-02. |
| R1-09 | minor | `evaluator.py:85,152-158` | The protocol has two decision points but the draft has one. Before dev there is no M̂, but `evaluate` requires one, so a caller has to invent it. The rule "none passes → not testable before dev" (P l.175; manifest `row_timing.before_dev`) only appears as the side field `any_row_passed_before_dev`, which is untested. After dev, rows are recomputed from whatever records are supplied, so "must already have passed" (l.175, l.206) is not bound to a sealed pre-dev result. | Split the evaluator into (a) a pre-dev row evaluation with no M̂ input that returns row results, status and a result digest, and (b) a post-dev selection that takes the sealed pre-dev digest and the bound M̂ evidence. Test: no rows pass gives "not testable before dev"; tampering with the sealed rows is refused. |
| R1-10 | minor | `provenance.py:21-60`; `evaluator.py:156-159` | P l.181 says "Report every completed run". The output contains only the maxima: no run IDs, causes, completion flags or per-run inputs. | Add the full run list and its classification to the output. |
| R1-11 | minor | `evaluator.py:9-27`; README l.19 (Open choice 4) | The binding-cell identity is the HMAC seed message (manifest l.47: compact sorted-key cell JSON). It is defined only here. No test pins it, so a changed key name or dimension silently changes the seeds, and digest stability is untested. No reviewed binding-cell enumerator exists elsewhere: the sibling controller enumerates only safety groups. | Keep this schema as the one canonical binding-cell identity. Assert the per-kind counts (16,128 / 1,792 / 252). Pin the SHA-256 of the sorted canonical identity list in a test. State in the README that the stage-3 generator must import `expected_cells` and `identity`. |

## Open-choice verdicts

1. **Synthetic-only and non-authoritative: accept.**
   - This is consistent with P l.220 ("Only draft/synthetic work is authorised now") and l.228. It only makes the gate stricter.
   - Accepting the draft does **not** satisfy round 9 l.451's "reviewed final evaluator". The authoritative variant, with real trust anchors and the ability to emit `binding_gate_established`, needs its own scoped zero-issue review.
   - This choice does not defer R1-01: tying the bundle to the evaluated inputs can be implemented and tested now on synthetic data.
2. **Per-population power-cell inadequacy: accept.**
   - It is provably stricter than the per-cell reading of l.169. If p̂ᵢ ≤ c − z√(c(1−c)/nᵢ) for every i, then p̂_pool ≤ c − z√(c(1−c))·Σ√nᵢ/N ≤ c − z√(c(1−c)/N), because Σ√nᵢ ≥ √Σnᵢ.
   - It is therefore fail-closed. Its boundary test is part of R1-07.
3. **Resource input conventions: issue (R1-02, and R1-01 for provenance).**
   - Accepted parts: max-phase lifetime high-water RSS, "no planning slope/intercept substitutes for locked RSS", ×2 applied once, and the memory check.
   - Recommendation: take per-input measured maxima and have the evaluator compute the rev8 binding arithmetic, with wall days = binding core-hours/192. Don't accept supplied totals or pre-charged wall days.
4. **Canonical dictionaries, zero-based populations, separate counters: partly accepted.**
   - Zero-based indices are accepted (manifest l.48). Separate gate and adequacy counters are accepted (manifest l.43, l.60).
   - The dictionary serialisation is accepted as consistent with manifest l.47, but it must be pinned (R1-11).
   - External bundle membership is an issue (R1-01).

## Checked and correct

- Wilson z = 1.959963984540054, the same convention as `validation/reproduce_reports.py:30-34`, `ranking.py:300-304` and `controller.py:28-35`.
- Cell counts: 16,128 + 1,792 + 252 = 18,172 cells, × 5 = 90,860 records.
- `kind` labels match the manifest's `cell_identity_kind`.
- Exact-membership and duplicate refusal.
- B = 1,999 is checked at both manifest and record level.
- Counter reconciliation logic.
- The pooled tail gate over 10,000.
- Null false-support upper ≤ .05.
- Largest registered M ≤ M̂ is chosen before its pass status is consulted, so no lower passing row can be used to bypass a failing one (tested: 848 selects 700).
- Safety is kept out of the gate.
- The fixed 8 workers, 8 × max RSS ≤ hw.memsize − 8 GiB.
- The rerun-cause rules.
- R9-reading-1: completed closure-mismatch runs enter the maxima (tested).

## Considered and not counted

- Exceptions raised in place of a recorded "not testable": fail-closed.
- `lower_miss + upper_miss ≤ n` is unchecked: can only fail more cells.
- A completed run tagged as a crash: it enters the maxima and still requires a rerun.
- The controller's `report()` lacks the evaluator's record fields: the generator/scorer is a separate stage-3 item (P l.224).
- `__pycache__/` is git-ignored.
- The authenticity of the trust anchors: deferred under Open choice 1.

## Verification run

- **No tests, mutations or hashes were run in this round.** Neither this session nor either peer had a shell or write tool. Everything above comes from reading the files, `grep`, and checking `oc-manifest.json`, `rev9-digests.json` l.405–412 and `reproduce_reports.py:205-330`.
- The README's "6 tests passed" and "126/131 literal / 5/5 groups / 131/131 unchanged" results **rest on Main's report** and were not reproduced.
- Mutation-survival claims in R1-06 and R1-07 are marked [INFERENCE].
- Commands for Main to run from a scratch copy:

```sh
W=/Users/adamjackson/Projects/work/scout-x1g; S=$(mktemp -d /tmp/x1eval.XXXX)
mkdir -p $S/evaluation && cp -R $W/evaluation/x1-triage-ranking $S/evaluation/
cd $S && PYTHONDONTWRITEBYTECODE=1 /Users/adamjackson/Projects/work/scout/.tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/test_evaluator.py -v
cd $W && PYTHONDONTWRITEBYTECODE=1 /Users/adamjackson/Projects/work/scout/.tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py
shasum -a 256 evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/{README.md,evaluator.py,provenance.py,test_evaluator.py,check_frozen.py}
```

- Mutants to confirm each test gap in the scratch copy. Each is expected to still pass today:
  - delete `n <= e <= 4000`;
  - delete the attempt cap;
  - change `.035` to `.5`;
  - change `interval[0] >= .80` to `True`;
  - delete the inadequacy condition at l.138;
  - delete `r["mask_reachable"]`;
  - delete the CPU limit;
  - delete `(i and previous_failure is None)`;
  - delete the host check at `provenance.py:39`.

## Recommendation, alternatives, risks and fix order

- **Recommendation:** keep the small evaluator. Fix R1-01 through R1-11 in the same draft, then resubmit it for round 2 of this scoped review.
- **Alternative considered:** a separately reviewed loader and projection component that feeds one digest-bound normalised record to an arithmetic-only evaluator. It is viable, but it adds a second artefact and a second review, and the evaluator must not be callable without it. Rejected for that extra surface.
- **Risks:**
  - Any change to the identity schema (R1-11) moves every seed. Pin it now, before the controller or generator consumes it.
  - R1-02 makes the fit depend on the locked `p_gate` and units, as P intends. That is reversible only through a new council round.
- **Fix order:**
  1. R1-01 and R1-03: bundle-to-input binding and the OC host.
  2. R1-02: rev8 binding arithmetic.
  3. R1-09 and R1-10: two decision points and the run report.
  4. R1-04 and R1-05: verdict order and type checks.
  5. R1-11: identity pin.
  6. R1-06 to R1-08: tests.
  7. Rerun the tests and the mutants above, and record the SHA-256s.

## How the peers changed the outcome

- The challenger added R1-05 and the verdict truthiness part of R1-04.
- The peer added R1-09 and the identity-pin framing of R1-11.
- Both independently confirmed R1-01 to R1-03 and R1-06 to R1-08.
- Neither changed the verdicts on Open choices 1 and 2.

## Verdict

**11 issues open** (8 major, 3 minor). The draft is not accepted for its stage-3 role. This review gives no authority to run anything.
