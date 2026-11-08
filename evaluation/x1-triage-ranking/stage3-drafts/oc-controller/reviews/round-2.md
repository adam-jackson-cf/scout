# X1 stage-3 draft review: OC controller (`stage3-drafts/oc-controller/`), round 2

**Scope.** This is round 2 of the scoped review. It covers PR #21 at head `ed7f21ce2bc58f03c4dee6e0eeaa7092ab9dbc4f` in worktree `/Users/adamjackson/Projects/work/scout-x1g`. The review was read-only: no edits, no model calls and no real data. **It gives no run authority.** Round 9 §12 still lists the blocking stage-3 prerequisites.

## Reviewed files

Main recomputed these SHA-256s at the head above. Paths are relative to `stage3-drafts/oc-controller/`.

| File | SHA-256 |
|---|---|
| `README.md` | `583967f896c4ee0b602769ff9986173e125003b53966605f5180afa9bd1897b4` |
| `controller.py` | `e2740ddc4e7d7bb5ffe5a826c292a1b87dd9ad74c97463aaf13b1f81491c94bb` |
| `known_answers.json` | `24b0b48497cac0955737bef05661b109462223304eac068898151a13cd15a694` |
| `test_controller.py` | `41419d66799b525b82c0f857fd0dede4f813a21351cb63744ea4e16b3a405f5e` |
| `reviews/round-1.md` | `f5a1c845f81d3d22e9d357c03e7c233c382317d8894e67551a10a3558a90539b` |
| `reviews/round-1-response.md` | `682af9a486173b56c3a302e93251134f2592a3a0a9100ac174c4d6c8296b8ce8` |
| `reviews/round-1-test-output.txt` | `4d970a1f1c77544377c8dff41a23a6a3efae447b0ec28b8e324d467eb534f91a` |
| `reviews/round-1-cap-output.txt` | `1d442d16d51c04f4120587d56c4638fa9035e88280376da4fccd70cb690cfeb4` |

## Signed basis (unchanged)

| Item | SHA-256 |
|---|---|
| `PREREGISTRATION.md` (P) | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `oc-manifest.json` (M) | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |
| `validation/fixed_population.py`, as pinned in `controller.py:20` and `rev9-digests.json` l.468 | `ed670a6bf8076d791c06cfc2426aff03a0a2f9b1730982cf3fdfb2ec13bc49a2` |

**Governing text:**
- P l.163: counters, Wilson intervals, fixed seed order.
- P l.169: calibration disclosure; unreachable targets are unmeasured; "Missing/safety use the same rules".
- P l.171: safety order, frozen scheduling policy, statuses.
- P l.181: "Workers are fixed at 8, with no reduction".
- P l.185: caps, ×2 overhead, "Keep … all descriptive safety outcomes".
- P §11: "Refuse … duplicate stages"; "pause affected work for review".
- M l.46–54 (seed recipe), l.113–149 (`final_safety_OC`, `construction_reuse` "336*5", `reachability`), l.167.
- `fixed_population.py:199–210`; round 9 l.391; round 8 §2 F.

## Round-1 issue disposition

| ID | Fixed | Evidence (file:line) and catching test | Residual |
|---|---|---|---|
| R1-01 | **yes** | **Code:** the signed `population_identity`/`criticality_key` load after a digest check (`controller.py:35–49`). The identity is built from the six M dimensions plus κ (`:62–68`). HMAC runs over the identity's UTF-8 bytes with a byte tie-break, and the ordered-list SHA is taken over the identity-string array (`:185–186`). `criticality_key` is used at `:360`. README l.9 lists the keys.<br>**Tests:** `test_signed_identity_exact_dimensions_and_independent_order_pin` (`test_controller.py:54–80`: key set, signed-kernel equality, first and last identity, list SHA via `json.dumps`, OpenSSL HMAC of the first identity); `test_seed_known_answers_openssl_and_input_refusals` (`:82–93`). | The pinned value's independent derivation is not committed (R2-10). |
| R1-02 | **yes** | **Code:** M digest check (`:178–180`); the 2,016 / unique / 336-at-κ=0 check (`:183–184`); pinned order and policy check (`:223–225`); a create-once `open("x")` freeze record written before dispatch (`:206–213`).<br>**Tests:** `test_signed_identity_…` (`:71–78`) and `test_manifest_substitution_and_create_once_freeze` (`:95–107`). | The freeze path is chosen by the caller and leaves out clock origin and p_gate (R2-07). Refusal on an order or policy mismatch is untested (R2-09). |
| R1-03 | **no** | **Code:** each attempt is authorised before the gate (`:120–127`, `:274–298`, `:227–239`); termination snapshot (`:408–420`); a result arriving after termination is discarded (`:143–144`); grid and integrity causes are logged (`:405–406`, `:426–428`).<br>**Tests:** `test_gold_before_during_and_after_last_trial_is_irreversible` (`:219–237`), `test_caps_at_execution_boundary_and_wall_clock_unpaused` (`:239–256`), `test_cap_during_last_trial_and_criticality_freezes_partial_counts` (`:258–282`), `test_serial_lease_refuses_reentry_and_gold_inside_trial_discards_result` (`:109–126`). | The required fix said "refuses work when terminated, paused or over a cap" and "freeze the active group's counts at that moment". Three gaps remain: a trial still starts after a stop that arrives during the gate, and the snapshot can be torn or lost (R2-02); the caller's pacing still decides how many groups are measured (R2-03); `terminated` is a public flag that can be set back (R2-08). |
| R1-04 | **no** (partly) | **Code:** (a) raw meter compared with `>=`, no ×2 (`:231–236`); (b) `pause`/`resume` with evidence and a running clock (`:311–339`); (c) fixed safety lane order (`:293–296`).<br>**Tests:** `test_caps_…` (`:239–256`), `test_integrity_pause_resume_preserves_attempt_and_group` (`:284–300`), `test_population_lane_mapping_and_finish_membership` (`:179–186`). | The global non-blocking lock forces binding onto one worker (R2-01, introduced by this fix). The required fix said "test each part", but the wall-clock start point (`oc_started_at` different from construction time) is untested (R2-09) and is not recorded (R2-07). |
| R1-05 | **yes** | **Code:** identity-bound snapshots (`:396–406`); exhaustion is judged only from the active group (`:414–416`); the report requires a termination event (`:430–434`).<br>**Tests:** `test_exhaustion_is_bound_only_to_active_group` (`:319–330`), `test_gold_…` (`:220–224`), `test_criticality_exact_keys_order_and_auto_completion` (`:197–217`). | — |
| R1-06 | **no** | **Code:** `mark_unreachable` (`:384–394`).<br>**Test:** `test_unreachable_provenance_advances_in_order_and_grid_completion_logged` (`:332–351`) builds each record itself and passes `sha256(canonical(record))` as the "expected" digest (`:345`). | The record is not bound to the calibration provenance. Per-population unreachability found after lane 0 has started still stalls, which contradicts the required fix's "keep the actual counters". See R2-04. |
| R1-07 | **yes** (structural) | **Code:** leases are keyed by `(identity, index)` (`:286–298`). `finish` (`:370–382`) can only succeed from the fourth criticality phase (`:366–368`), because `active["criticality"]` never holds four keys outside that call.<br>**Tests:** `test_population_lane_mapping_…` (`:179–195`), `test_criticality_…` (`:197–217`). | The membership tests do not isolate their guards (R2-09). A mutable `manifest` reopens `finish` (R2-08). |
| R1-08 | **yes** (trial crashes) | **Code:** counters are `init=False` (`:97–102`); `crashed` is charged at `:140` and cleared at `:152`; invariant at `:158`; the crash latches a pause (`:147–150`); `resume` is refused while a crash is unresolved (`:336`).<br>**Test:** `test_crash_and_nonbool_latch_integrity_with_reconciled_counts` (`:302–317`). | Gate, criticality and `BaseException` paths are not latched (R2-06). The resolution step accepts an asserted outcome (R2-05). |
| R1-09 | **yes** | **Code:** phases, integer types, signs and population-phase attempt 0 are validated (`:52–59`).<br>**Test:** OpenSSL vectors in `known_answers.json`, checked by `test_seed_known_answers_…` (`:82–93`). | Any string cell is accepted without a canonical-form check (R2-11). |
| R1-10 | **no** (partly) | **Code and tests:** the listed counter, probability, meter, gold, mutation and outcome-independence boundaries are covered (`:151–177`, `:219–256`, `:353–378`). | Both Wilson known answers sit at k=0 and k=n, so the variance term is untested (R2-09). README l.13 still says the policy digest was "independently computed" (R2-10). |

## New issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R2-01 | major | `controller.py:23` (`parallel_attempts: False`, `worker_mapping`), `:197`, `:241–252`, `:274–276` | Binding leases share the schedule's single non-blocking `_execution_lock`. A second binding `step` running at the same time raises "parallel or reentrant execution refused". If eight separate schedule instances were used instead, there would be eight freezes, meters and termination states, which breaks the shared caps. Binding OC is therefore one worker, against P l.181 "Workers are fixed at 8, with no reduction". On revision-8 arithmetic, the binding path (711.85 ×2 core-hours) is about 14.8 raw days on one worker, which is at the 15-day cap [INFERENCE]. | Give each population lane its own lock. Allow up to 8 concurrent binding lanes (slots 0–7), each serial in attempt index. Hold a short state lock only around `_boundary`, counter commits and transitions (shared with R2-02). Update the hashed policy and re-pin `POLICY_SHA256`. Tests: 8 concurrent binding lanes succeed with counts identical to a serial run; a 9th concurrent lane, or one population on two threads, is refused. |
| R2-02 | major | `:129–155`, `:311–317`, `:402–420`, `:422–428` | **Execution is not atomic with stops and pauses.**<br>(a) Authorisation happens only before `gate`. A gold, pause or cap that arrives during `gate(index)` (`:134`) still lets `trial(index)` start (`:142`).<br>(b) `stop_gold`/`_terminate` take no lock. If `_terminate` runs between `:139` and `:140`, or between `:152` and `:153–154`, `report()` raises inside `_terminate` after `terminated=True` has been set (`:411`). `_final` then stays `None`, and `gold_report()` returns `None`.<br>(c) If it runs between `:403` and `:404`, the active group is snapshotted twice and `ordered[n+1]` is skipped.<br>(d) Gold racing a cap can log two terminations.<br>(e) A pause during an in-flight trial commits that trial's result.<br>The draft's own policy says "snapshot counts immediately". | Serialise every transition (termination, pause, completion, counter commit, snapshot) under one state lock. Re-authorise under that lock immediately before the trial starts. Write the in-flight rule into the hashed policy: a cap observed after a trial returns keeps that result; gold, an integrity stop or a pause that arrives during an attempt leaves it charged and unclassified, pending R2-05. Threaded tests inject gold, pause and cap during the gate and during the trial, and check: the trial is never invoked after a stop or pause in the gate; exactly one termination event; the snapshot invariant holds; the report lists the 2,016 groups in order with no duplicates. |
| R2-03 | major | `:254–272`, `:120`, `:344`; events at `:264`, `:317`, `:338`, `:412` | **The caller still sets the safety cadence.** The caller decides when to call `next_group`, `step` and `run_criticality`, and gaps between calls are not recorded. The caller sees trial results, so it can slow down after an adverse outcome and push the active group into "not completed by gold". That is the polling-time selection path that R1-03 and X1-R8-02 closed. It is also an unrecorded pause, against P l.171. No dispatch, lane or criticality timestamps are logged. | The controller owns the loop: `run(gate, trial, calibration, criticality)` runs safety to termination, and gold arrives asynchronously through `stop_gold`. Do not expose a per-step safety API. Journal timestamps for group dispatch, lane start and stop, calibration and criticality phase start and stop, and group completion (R2-07). Tests: the event sequence carries these timestamps, and no caller action between controller-issued work changes dispatch order or status. |
| R2-04 | major | `:384–394`; `test_controller.py:332–351`; `fixed_population.py:173–196` | **The unreachable status is self-certified and group-level.**<br>(a) `expected_sha256` comes with the record, so the digest check binds nothing. The other checks need only fields the caller already knows. After observing earlier groups, a caller can skip any later group, which is an outcome-dependent exclusion forbidden by P l.171.<br>(b) Calibration is per population (M `construction_reuse` "336*5"; `calibrate_population` returns `reachable`). If population k>0 is found unreachable after lane 0 has run, `mark_unreachable` refuses (`:387`) and the group stalls. That is R1-06 again.<br>(c) P l.169 requires target, realised and excess to be disclosed; neither the unreachable record nor any reachable group retains them. | Remove `mark_unreachable`. Inside the R2-03 loop, run a controller-leased calibration phase for all five populations before lane 0's first attempt. The controller records each kernel result (identity, population index, target, realised, excess, latent gain, `reachable`, kernel SHA) and computes the record digest itself, then keeps the records in every group's snapshot. Fix reachability per base profile at its first group and reuse it for that profile's other five κ groups; a conflicting record is refused. Any unreachable population gives "unmeasured (unreachable target)" and the run advances in order. Tests: an asserted record cannot skip a group; an unreachable population 3 gives the terminal status and order continues; profile consistency across κ; disclosure fields are present for reachable groups. |
| R2-05 | major | `:319–331`; README "Conservative choices" 4; `test_controller.py:313` | **`resolve_crash` takes `adequate` as a free argument.** An operator-chosen eligible or inadequate classification for a trial that never finished enters the P l.163 counters. The only evidence is any 64-hex string. This contradicts M `trial_sequence` and the README's own "never … a licence to invent a missing outcome". | After the reviewed evidence is recorded, resolve by re-executing the same attempt index under the lease with the same seed. The controller records that trial's own bool (and re-runs the gate for a gate fault, R2-06). No API accepts an asserted outcome. Test: resolution invokes `trial(attempt)` exactly once with the original index, and there is no `adequate` parameter. |
| R2-06 | minor | `:134–137`, `:147`, `:163`, `:361–364`, `:114`; `test_controller.py:205–208` | **Integrity faults outside the trial are not latched.**<br>- A gate exception (`:134`, outside the `try`) is charged as a gate failure, with no incident and no pause.<br>- A non-bool gate pauses the run, but records no population incident and is counted in `gate_failures` and its Wilson interval, misreporting a P l.163 counter. `resume` is then allowed straight away.<br>- A criticality `work` exception or malformed result has no incident or pause, and is retried immediately (the test does this).<br>- `BaseException` escapes `except Exception` with `crashed=1` and no pause. The safety population then runs past 2,000 executions, can never satisfy `== 2000` (`:114`), and ends as "unmeasured; cell fails". | Use one latch for gate, trial and criticality faults, including `BaseException`. It records a population or phase incident with the attempt or phase index, keeps the fault out of `gate_failures` (report a separate gate-fault count), pauses, and requires the R2-05 resolution. Add a separate test for each path. |
| R2-07 | major | `:176`, `:194`, `:198–213`, `:232–233`, `:408–420` | **No controller-owned durable record.**<br>- The freeze path is chosen by the caller, and every event, snapshot and termination exists only in memory.<br>- After `stop_gold()`, a second instance with another path refreezes and dispatches `ordered[0]` again. That produces two reports to choose from, against P §11 "Refuse … duplicate stages" and P l.171 "every pause or termination is recorded".<br>- If the process dies, every record is lost.<br>- The freeze leaves out `oc_started_at`, so a restart can reset the 15-day clock. It also leaves out p_gate, which sets the attempt caps.<br>- A meter fault raises `ValueError` with no recorded event, and `pause` cannot record it because `pause` reads the meter. | Use a fixed, controller-owned run directory (injectable root for tests) with a create-once freeze that includes the absolute `oc_started_at` and p_gate with its inventory source digest. Write an fsync'd, append-only journal entry before each transition takes effect. On construction, if a freeze exists, verify it and replay the journal instead of refreezing. Record an unjournaled process loss as a §11 integrity pause that needs a reviewed resume. A terminated journal refuses work and returns the same report. Latch meter faults as recorded integrity events through a path that does not read the meter. Tests: rebuild mid-group and resume at the same attempt; rebuild after gold and get an identical report with no dispatch; a second freeze is refused; a clock origin different from the frozen one is refused. |
| R2-08 | minor | `:181`, `:194`, `:202–203`, `:357`, `:370–382` | **Decision state is public and mutable.**<br>- Setting `schedule.manifest["final_safety_OC"]["critical_j"] = []` and then calling `finish(pops, {})` marks a group measured with no criticality work.<br>- Setting `schedule.terminated = False` after gold lets the same group become measured.<br>- `oc_started_at` and `paused` can also be overwritten. | Keep this state private and expose read-only properties. Capture runtime manifest values as immutable (for example a `critical_j` tuple) at construction. Make `finish` private, so completion happens only in the fourth criticality phase or the run loop. Test that assignment raises and that both sequences fail. |
| R2-09 | minor | `:37–38`, `:223–225`, `:377–379`; `test_controller.py:179–195`, `:376–377`; all fixtures use `oc_started_at == clock()` | **Refusal paths untested.**<br>- Kernel-digest refusal.<br>- Order or policy mismatch refusal.<br>- `finish`'s criticality check when the populations are complete.<br>- The non-bool gate.<br>- Wall-clock origin later than the OC start.<br>- An interior Wilson value (k=0 and k=n do not exercise the variance term).<br>Each check would survive deletion [INFERENCE, static trace]. | Add isolating negative tests: an injected kernel path; a patched `POLICY` or `ORDER_SHA256` gives `RuntimeError` and no freeze file; an OC start earlier than construction; decreasing meters; an interior Wilson known answer (for example `wilson(50, 2000)`) with independent provenance. |
| R2-10 | minor | `known_answers.json` (`method`); `round-1-response.md:24`; README l.13 | **The pinned digest's independent derivation cannot be reproduced.** The "separate construction" behind `04aa2184…` and the first and last identities is described but not committed. The policy digest is a hash of the controller's own `canonical(POLICY)`, yet the README calls it "independently computed". | Commit the standalone construction (no controller import) with its output and a reproduce command next to `known_answers.json`. Reword the policy-digest claim. |
| R2-11 | minor | `:254–256`, `:243–244`, `:57` | **Two caller inputs are trusted.**<br>- `next_group` accepts a caller Boolean while a leased binding population is unfinished; that population can never run afterwards (P l.171 "after all binding rows").<br>- `seed` accepts any `str` as cell bytes without checking it is the "UTF-8 canonical compact sorted-key cell JSON" that M l.47 requires. | Refuse to start safety while any leased binding population is neither complete nor exhausted, or has an unresolved crash. Require a string cell to equal `canonical(json.loads(cell))`. Test both. |

**Counts:** 0 blocking, 6 major (R2-01, 02, 03, 04, 05, 07), 5 minor (R2-06, 08, 09, 10, 11).

### Rejected or not raised

- **Cached `sys.modules` substitution** (challenger): not an issue. It needs someone to plant a module under the controller's private name on purpose. The freeze record digests the source files.
- **Gold versus cap handling of in-flight work:** both rules are outcome-independent. R2-02 only requires the rule to be written into the policy.
- **Running criticality for a group that is already exhausted:** outcome-independent. The extra cost was not raised.

## Open-choice verdicts (R-064: one recommendation each)

1. **Serial increasing-index safety lanes, eight reserved slots with 5–7 idle, serial j phases: accepted** (round 1; round 8 §2 F). The risk is accepted by name: on revision-8 arithmetic, serial lanes need about 12.0 raw worker-days for the full grid [INFERENCE: 576.589/2 core-hours ÷ 24], so gold or the 15-day cap may cut the grid short. This does not depend on outcomes. Binding must change regardless (R2-01).
2. **Raw actual meter; clock from the first full-OC start, never paused; `>=` caps: accepted.** The origin must be frozen (R2-07).
3. **Compact JSON array of identity strings, byte tie-break, separately pinned policy digest: accepted.** Re-pin after the R2-01 and R2-02 policy changes. Create-once freeze before dispatch is accepted only with a controller-owned path (R2-07).
4. **Crash resolution by an asserted outcome: rejected** (R2-05). Use a reviewed same-seed re-execution.
5. **Unreachable only before any attempt: accepted as a guard,** provided the calibration of all five populations precedes lane 0 (R2-04).
6. **"Measured" requires all four criticality records: accepted.** It is stricter than P l.171.

**Recommended remediation (one design).** Extend the schedule-owned state machine into a controller-owned run loop:
- It has a durable run directory: a create-once freeze that includes the clock origin and p_gate, plus an fsync'd journal that is replayed on restart.
- One state lock covers every transition.
- Each lane has its own lock: up to 8 binding lanes, and the frozen serial safety lanes.
- Each group gets a leased calibration phase before its first attempt.
- One latch covers gate, trial and criticality faults, resolved by same-seed re-execution.
- All decision state is private and immutable.

The rejected alternative was keeping a caller-driven library and adding audits. It cannot make pacing or duplicate instances non-discretionary.

## Peer contributions

**Challenger (GPT Astra):**
- Found the gate-window trial start, folded into R2-02.
- Found the public `terminated`/`manifest` bypass sequences (R2-08).
- Found the unlatched gate, criticality and `BaseException` paths (R2-06).
- Found the self-certified unreachable digest (R2-04).
- Found the non-isolating tests (R2-09).
- Rated R2-06 and R2-08 major. The council rates them minor: R2-08 needs deliberate assignment, R2-06 has no outcome-dependent path, and both follow the round-1 calibration of R1-07/R1-08.
- Its cached-module finding was rejected (see above).

**Experiment peer (Opus):**
- Found binding forced onto one worker (R2-01) and caller pacing (R2-03); the lead had found neither.
- Extended the lead's clock-origin point to duplicate instances and the durable journal (R2-07).
- Recommended same-seed re-execution (R2-05). This replaced the lead's weaker record-digest binding.
- Found per-population calibration from `fixed_population.py:173–196` (R2-04b), and raised the interior-Wilson, unretained-construction and binding-Boolean points.

**Lead:**
- Found the torn or lost snapshot and the duplicated or skipped group under a concurrent stop (R2-02b–c).
- Traced the `BaseException` path to the `== 2000` consequence (R2-06).
- Showed that `finish`'s membership guarantee is structural, so `finish` should be private.
- Proposed the per-profile reachability consistency check.
- Assessed the non-canonical `seed` string check jointly with the peer (R2-11).

## Verification

- **This council session and both peer sessions had no shell** (read, grep, glob and LSP only). No command, test, hash, OpenSSL call or mutant was run here.
- Every "survives deletion" or race claim is a static trace [INFERENCE].
- File digests and test results come from Main's independent shell run at `ed7f21ce` with pinned CPython 3.11.11: `test_controller.py` ran 19 tests in 0.822s, OK. The ceiling enumeration (`round-1-cap-output.txt`) is the author's output and was not rerun by the council.
- The static evidence comes from reading `controller.py`, `test_controller.py`, `README.md` and `known_answers.json` in full; `round-1.md` (raw l.40–49) and `round-1-response.md`; P l.161–186 and l.187–230; M l.37–60 and l.110–170; `fixed_population.py:190–210`; and the round-8 §2 F and l.336 greps.
- **Main should run before round 3:**
  - `shasum -a 256` on the revised files;
  - the README test command;
  - an independent recomputation of `04aa2184…` with a script that does not import the controller;
  - reversion mutants for the R2-09 guards.

## Verdict

**11 issues open** (0 blocking, 6 major, 5 minor). R1-03, R1-04, R1-06 and R1-10 are not fully fixed. The draft is not accepted for its stage-3 role. A revised draft needs a new scoped round with zero issues, and that round must pin the re-derived order and policy SHA-256s and the source digests. Acceptance would still give no run authority.
