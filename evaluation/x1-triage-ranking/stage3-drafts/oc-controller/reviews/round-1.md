# X1 stage-3 draft review — OC controller (`stage3-drafts/oc-controller/`) — round 1

**Scope.** This is the scoped innovation-council review that X1 round 9 recommended (§11 step 5; §9 l.391: "the OC controller, with both counters, the frozen safety order and scheduling policy, and the SHA-256 of the ordered group list recorded at freeze"). The draft is in PR #21: worktree `/Users/adamjackson/Projects/work/scout-x1g`, branch `agent/x1-option-g`, head `7168c0e1b2cbfaacbd3ab81da0910eeaa09b531e`. The review was read-only. **This review gives no authority to run OC, inventory, scans, models or any real-repository work.** Round 9 §12 authorises preparation actions (a)–(e) only.

## Reviewed files

These SHA-256s come from Main's report. The council did not recompute them (see Verification).

| File | SHA-256 |
|---|---|
| `stage3-drafts/oc-controller/README.md` | `6b298d3b3492a941abb9bfc04448d3ee306d2216beb8c3ddc09a833e3f3913bd` |
| `stage3-drafts/oc-controller/controller.py` | `451866837cf3aa575a5c6b6ecdfb9f1f36a93138778c3cb809da8c024db4d48b` |
| `stage3-drafts/oc-controller/test_controller.py` | `7d1d86161398d08e30bdc53244109b9ff4369b704cfe88d27fc3f1a43a625c92` |

## Signed basis

| Item | SHA-256 |
|---|---|
| `PREREGISTRATION.md` (P) | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `reviews/rev9-digests.json` (manifest) | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |
| `oc-manifest.json` (M), as recorded at rev9-digests l.409 | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| `validation/fixed_population.py`, as recorded at l.468 | `ed670a6bf8076d791c06cfc2426aff03a0a2f9b1730982cf3fdfb2ec13bc49a2` |
| `validation/self_test.py`, as recorded at l.500 | `36d07d5648fdb0db325cb7583b02c4117cab1ada07b5fefe4590c66c8f03eb8a` |

**Governing text:**
- P l.163: both hard counters, Wilson reporting, fixed seed order.
- P l.169: unreachable targets are unmeasured; "Missing/safety use the same rules".
- P l.171: safety grid, keyed order, frozen scheduling policy, statuses.
- P l.179–185: caps, 2× overhead, "Keep … all descriptive safety outcomes".
- P §10 clocks; P §11 ("pause affected work for review"); P l.224.
- M l.46–54 (seed recipe, identity kinds), l.113–149 (`final_safety_OC`), l.167 (`runtime_overhead_multiplier_in_projection`).
- Rounds 7 and 8: X1-R7-04 and X1-R8-02.

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R1-01 | major | `controller.py:10–11, 19–25, 89, 94, 96`; `README.md:18`; `test_controller.py:45–46` | **A second "canonical `safety-interval-group` identity".** The signed kernel already defines this identity, and it is self-tested at `self_test.py:448–451`: `population_identity` returns `json.dumps([kind, values], sort_keys=True, separators=(',', ':'))` (`fixed_population.py:199–205`), and `criticality_key` is at l.208–210. The draft instead HMACs `canonical({"kind": "safety-interval-group", **dimensions, **fixed_factors, "unresolved_kappa": κ})`. That produces different bytes, so a different P l.171 order and a different ordered-list SHA-256 from anything the signed code can rebuild. The draft also puts fixed factors into the identity, but M l.53 defines the identity as the dimensions plus `unresolved_kappa` (critical share omitted). The key set is stated nowhere. | Build each identity with `fixed_population.population_identity('safety-interval-group', dims)`. `dims` is exactly the six M `final_safety_OC.dimensions` keys plus `unresolved_kappa`, and the README lists that key set. HMAC the identity's UTF-8 bytes. Use the same identity bytes as the seed cell JSON, and `criticality_key` for the criticality phase. Hash the ordered list as the compact JSON array of identity strings in frozen order. Add known-answer HMAC and SHA vectors computed independently (e.g. `openssl dgst -sha256 -hmac`), following `self_test.py:122–127`. |
| R1-02 | major | `controller.py:88–108`; `README.md:26`; `test_controller.py:43–47` | **The freeze is not bound to reviewed values.** `SafetySchedule` freezes whatever list it receives: a 10-group synthetic list or any subset. That conflicts with P l.171 ("Reuse 336 complete plus 1,680 full-vertex groups exactly") and P l.185 ("No grid/cap reduction"). The ordered-list and policy digests are computed from the inputs, held in memory, and checked only against themselves. No create-once record is written at freeze, which round 9 l.391 requires ("recorded at freeze"). The README postpones the real 2,016-group digest even though it is computable now from signed inputs, without outcomes. The test recomputes its expected value with the code under test. | Inside the controller, build the groups from M after verifying M's SHA-256 `1caea1a5…`. Assert 2,016 groups (336 with κ = 0, 1,680 with κ > 0). Require the expected ordered-list and policy SHA-256s, stated in the README and pinned in reviewed source, and refuse to start on any mismatch. Write a create-once freeze record (both digests, source digests, timestamp) before the first dispatch. Test the real 2,016-group digest against the pinned constant. |
| R1-03 | major | `controller.py:60–72, 110–128, 141–149` | **Termination is not enforced where work executes.** Gold and the caps are checked only when the caller polls `next_group`. `Population.step` never consults the schedule, so after gold or a cap trials keep running. Resources are never checked during a group. `gold_report` neither terminates the schedule nor fixes the result. One sequence shows this: dispatch, `gold_report()` (group "not completed by gold"), then `finish()`, then `gold_report()` (the same group is now "measured"). Polling time therefore decides which groups are measured: the selection path X1-R8-02 closed. Finishing the whole grid ends the run silently (`:125–126`) with no termination event. P l.171 requires "every pause or termination is recorded with its cause". README l.12 says "Caller must reserve work", but no reservation interface exists. | Route every attempt through a schedule check that refuses work when terminated, paused or over a cap. Make gold and cap termination irreversible: freeze the active group's counts at that moment. Record a cause for every termination, including grid completion and any stop before gold. Test termination before, during, and right after the last trial and the criticality work. |
| R1-04 | major | `controller.py:95`; `README.md:12, 16–17` | **The frozen policy is prose, not operational.** P l.171 freezes start, worker allocation, pauses and termination. (a) **Meter undefined.** `core_hours`/`wall_days` are numbers the caller supplies, "already charged with the 2× factor". But M l.167 names that factor `runtime_overhead_multiplier_in_projection`, and doubling measured actuals would stop safety at about 7.5 real days [INFERENCE]. That works against P l.185 ("Keep … all descriptive safety outcomes"). The OC wall-day clock also has no stated start point. (b) **No pause path.** `"pauses": "none"`, yet P §11 requires pausing affected work for review, and such a pause would go unrecorded. (c) **Workers.** "workers: 8" never says how 8 workers map onto at most five serial population lanes. An unfixed allocation lets throttling during a group move that group past gold. | Put operational content into the hashed policy: raw measured cumulative OC core-hours (binding plus safety); wall-clock time from the first full-OC process start, never paused; caps compared with `>=` on actual values, with no 2× on actuals; checks before every attempt; a `pause(cause)`/`resume()` limited to outcome-independent §11 integrity evidence, with start, end and cause recorded, resuming at the same group and attempt index while the clock keeps running; and a fixed worker mapping. If attempts run in parallel, commit results in attempt order and discard work past the stop. Test each part. |
| R1-05 | major | `controller.py:130–149`; `test_controller.py:65–68` | **Statuses and counts depend on an optional, unbound argument.** `gold_report` applies an exhaustion status to group `self.index` even when no group is active. The test enshrines this: an unstarted group is reported "unmeasured; cell fails", which contradicts P l.171 ("every unstarted or unfinished group is 'not completed by gold'"). If the argument is left out while a group is active, the exhaustion and the "partial counts shown separately" are silently dropped. `finish()` keeps only a status string, discarding the counters, gate failures and Wilson intervals that P l.163 requires each population to report. | Keep identity-bound population snapshots for completed and active groups, and derive statuses, counts and pooled membership from them. Apply exhaustion precedence only to the active group's own five populations. Require a termination event before the gold report. Replace the fixture at `test_controller.py:65–68` with a dispatched, bound group. |
| R1-06 | major | `controller.py:130–139` | **An unreachable calibrated profile has no terminal path.** M l.136 ("unreachable profiles unmeasured") and P l.169 ("Missing/safety use the same rules … unreachable unmeasured") say what happens. In the draft, `finish()` accepts only complete or counter-exhausted populations, so an unreachable group raises "unfinished populations". The serial schedule then stalls, and every later group becomes "not completed by gold". | Add an explicit terminal status, "unmeasured (unreachable target)", bound to the calibration provenance. Keep the actual counters (zero if no trials ran), exclude the group from pooling, and advance in frozen order. Test that later groups keep their order. |
| R1-07 | minor | `controller.py:38–45, 130–139`; `test_controller.py:53–54` | **`finish()` does not tie its inputs to the dispatched group.** `Population` carries no group identity or population index, so `finish([p]*5, criticality_complete=True)` marks a group measured using one population, and stale or foreign populations are also accepted. This undermines P l.171 ("measured only when all five populations reach the 2,000 stop"). Criticality completion is a bare Boolean that is never tested as `False`, and it is not checked against the four M `critical_j` values. | Give each population a `(group identity, index 0–4)`. `finish` requires each index exactly once for the active group. Take per-j criticality completion records keyed by `criticality_key(identity, j)` for exactly `critical_j`. Test duplicates, a foreign group, four populations, a binding population and incomplete criticality. |
| R1-08 | minor | `controller.py:38–45, 60–72` | **A crash leaves an execution charged but unclassified.** If `trial` raises (or returns a non-bool after `executions += 1`), the execution is counted but neither `eligible` nor `inadequate` is. A safety population can then report "measured" with 1,999 classified outcomes, which breaks the P l.163/171 denominator invariant. The dataclass constructor also accepts arbitrary counters, and tests 4 and 5 rely on that. | Latch the crash as an integrity incident (R1-04 pause path) and refuse to complete until it is resolved. Keep the invariant `executions == eligible + inadequate + crashed` with a reported `crashed` count. Make the counters `field(init=False)` and drive the fixtures through `step`. |
| R1-09 | minor | `controller.py:14–16`; `test_controller.py:69–70` | **The seed recipe is checked only with inequalities.** A wrong key, separator, endianness, slice or index encoding would still pass, even though M l.46–50 fixes all of them. `phase` is not validated against M l.49, and the population index type is not checked. | Add independent known-answer seed vectors, including population separation and attempt 0 of the population phase (M l.50). Refuse phases outside M l.49 and non-integer or negative indices. Use the R1-01 identity bytes as the cell. |
| R1-10 | minor | `test_controller.py:11–72`; `README.md:5, 10` | **Untested boundaries and README overclaims.** No test covers: a non-integer cap (`Population(.3)` → 13,334); completion on exactly the last allowed attempt or execution counting as measured; `step` after binding or safety completion; `wall_days` = 15 and the `gold` cause; NaN, negative or >1 `p_gate`; non-bool adequacy; known-answer Wilson values; mutation of the frozen order; outcome independence (same dispatch, events and statuses under opposite synthetic outcomes). The README claims "independently computed order/digest", but the test reuses `controller.canonical`. It claims "complete group dimensions", but the contents are never checked. It cites P l.169/175/181, which are not implemented. | Add these tests. Correct the README claims and line citations. |

**Counts:** 0 blocking, 6 major, 4 minor.

### Checked and sound (no issue)

- **Counters.** They are independent (`ceil(4000/p_gate)` attempts and 4,000 executions). A rejection never invokes the trial. Inadequacy consumes an execution. Binding stops at 2,000 eligible and safety at 2,000 executions. Completion takes precedence over exhaustion when both happen on the same step (`controller.py:52–72`; P l.163/171).
- **Float-to-ceiling arithmetic** for p̂_gate = count/100,000: the experiment peer proved by hand that there are zero mismatches against the exact rational ceiling. That proof is an [INFERENCE] and was not executed.
- **Wilson.** z = 1.959963984540054, the same constant as the signed kernel (peer). The formula has the standard Wilson form.
- **Pooled membership.** It excludes both non-measured statuses (P l.171).
- **Exhaustion versus "not completed by gold" precedence.** Settled in round 8 (X1-R8-02, "Not part of this issue"); the defect in R1-05 is how it is applied, not the precedence.
- **Planted effect / planted null.** Not applicable to a counter and scheduler module; they belong to the reviewed numerical and product harnesses. The metamorphic outcome-independence test in R1-10 is the analogue here.

## Open-choice verdicts

1. **Start immediately after all binding rows; eight reserved workers; one group at a time; criticality finishes before the next group: accepted.** This is consistent with P l.171 and round 8 §2 F (full use of every spare worker is not required). The worker allocation still has to be pinned down mechanically (R1-04(c)).
2. **No elective pauses; terminate on gold or the shared 2,880 core-hour / 15 OC-day boundary: accept the triggers, raise R1-03 and R1-04.**
   - Termination must be enforced before every attempt.
   - The meter is raw measured usage from a fixed start point.
   - §11 integrity pauses, which are not elective, must be recorded with their cause.
   - Grid completion and any stop before gold are recorded terminations.
3. **Serialization.**
   - **HMAC tie-break by identity bytes: accepted.** It is deterministic.
   - **Ordered-list SHA-256 over JSON of the draft's group objects: issue (R1-01).** Recommended: SHA-256 of the compact JSON array of signed `population_identity` strings in frozen order.
   - **Policy as compact sorted-key JSON: encoding accepted.** The content must become operational (R1-04), and the resulting digest must be pinned (R1-02).

**Recommended remediation (one design, per R-064):** a schedule-owned state machine.
- It builds the 2,016 identities from the verified M through `population_identity`, checks them against pinned digests, and writes a create-once freeze record before the first dispatch.
- It hands out identity-bound population leases (indices 0–4).
- It authorises every attempt against the termination, pause and meter state.
- It has four terminal statuses: measured; unmeasured, cell fails; unmeasured (unreachable); not completed by gold.
- It logs start, pause, resume and termination events with causes, and derives the gold report from retained snapshots.

The alternative was separate components with revocable execution leases. It was rejected because it needs more integration evidence for the same guarantees.

## Peer contributions

- **Challenger (GPT Astra):**
  - Independently found the identity divergence (R1-01) and the termination/execution-boundary gap (R1-03).
  - Found three issues the lead had not identified: R1-06 (unreachable profiles stall the schedule), the reversible gold-then-finish sequence (R1-03), and the unstarted-group status in the test fixture (R1-05).
  - Rated R1-07 and R1-08 as major. The council rates them minor: each needs the caller to pass wrong objects or to catch and continue past an exception, and neither has a path that depends on outcomes.
- **Experiment peer (Opus):**
  - Wrote the per-test catch/miss trace and the hand proof of the ceiling arithmetic.
  - Made the meter, pause and worker-mapping points operational (R1-04).
  - Raised the stop-before-gold termination cause, folded into R1-03.
  - Rated R1-02 minor. The council rates it major because round 9 l.391 makes the freeze record a named requirement, and an unbound freeze accepts a reduced grid.

## Verification run

- **No commands were executed.** This council session had only read, grep, glob, LSP and web-search tools, with no shell or Python. Both peer sessions reported the same limitation (challenger's LSP: `No language server found for this action`).
  - Not run: the README test command (`PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/oc-controller/test_controller.py`), `shasum -a 256`, and the 1–100,000 ceiling enumeration.
  - The README's "5 tests passed" is unobserved. The peer's static trace predicts 5 passes [INFERENCE].
- **Basis digests** for P, M, `fixed_population.py` and `self_test.py` were checked only as values recorded in `reviews/rev9-digests.json` (l.406, 409, 468, 500), not recomputed from worktree bytes. The artefact digests and the round-9 digest rest on Main's report.
- **Static evidence** came from reading the three artefact files in full, P l.141–229 (raw l.163/169/171/179/181/185), M l.37–55 and l.113–149, `fixed_population.py:190–210`, the `validation/*.py` greps for identity and seed helpers, round 7 l.355–380, round 8 l.300–345, and round 9 l.200–514.
- **Main should run, once fixes land:** the test command above, `shasum -a 256` on the three files and the signed basis, and the ceiling enumeration snippet.

## Verdict

**10 issues open** (0 blocking, 6 major, 4 minor). The draft is not accepted for its stage-3 role. A fixed draft needs a new scoped round with zero issues, and that round binds the pinned ordered-list and policy SHA-256s and the source digests. Even then, acceptance gives no run authority: full OC still needs every stage-3 prerequisite in round 9 §12.
