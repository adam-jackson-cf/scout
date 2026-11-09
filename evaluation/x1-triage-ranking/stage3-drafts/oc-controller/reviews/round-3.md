# X1 stage-3 draft review — oc-controller — round 3

**Scope.** This is round 3 of the scoped review of `stage3-drafts/oc-controller/`. It covers PR #21 at head `38cead909de6de815ef8e9453630f404f6e2628f` in worktree `/Users/adamjackson/Projects/work/scout-x1g`. The review was read-only: no edits, no model calls and no real data. **It gives no run authority.** Round 9 §12 still lists the blocking stage-3 prerequisites.

## Reviewed files

Main recomputed these SHA-256s with `shasum -a 256` at the head above. Paths are relative to `stage3-drafts/oc-controller/`.

| File | SHA-256 |
|---|---|
| `README.md` | `db902fcc72ac24f41cdbfa40b17a8be57091b52285afdbd4ff9a8073f5299a8c` |
| `controller.py` | `331770134461efeb09828bbc6e906251b3e5082ab36b8758e915f11e6f0061a9` |
| `derive_known_answers.py` | `cf86e79f65565388ed029cc5184d71e7874750f1d33f30ac8635cafe8d2e1b8b` |
| `known_answers.json` | `012699629663f464af546f8e5048e730af711b49f9932e3e4b6db115a5a24449` |
| `policy.json` | `decd2f3aa55fa61e8a8e5cf2a4abc7de33403a480f959fd8fa832030cfd0f918` |
| `mutate_round2.py` | `7732b41ca846c0eeedde88b91e0714da1f7841963d6a343f205b80e3a3fe9c66` |
| `test_controller.py` | `e029d1ca0375b0388692b341f310a9c5dbb9cfaa9ecc0fff4b66473c125ba7cf` |
| `reviews/round-2.md` | `d3b770d948fa2b7162f57a281b27a9201c3283e3f22f94436c40a4499945e22e` |
| `reviews/round-2-response.md` | `02f0cac73fd13f66699c974e8a4dd7dab64cfa5281928b1c25acdd41c1084dbf` |
| `reviews/round-2-test-output.txt` | `5b25a2d02a93e746da046c9233da84c2df22295243279f8eb181151962db79dd` |
| `reviews/round-2-derivation-output.json` | `012699629663f464af546f8e5048e730af711b49f9932e3e4b6db115a5a24449` (byte-identical to `known_answers.json`) |
| `reviews/round-2-mutant-output.txt` | `48d9c17a3e41adfd12186ecc2559d842ad54d4dae778383afe26d5884fe58467` |

Constants pinned in `controller.py:21–25`:

| Constant | Value |
|---|---|
| `ORDER_SHA256` | `04aa2184f79cb3795f60305a6b8fa31687d5acd7069846e78b6acf1b1f4dd4ff` (independently re-derived by `derive_known_answers.py`) |
| `POLICY_SHA256` | `71c848cdf686b5424629098f4a8bd30ea8b8ceae6ea050652e2a7f1d230dc33b` |
| `KERNEL_SHA256` | `ed670a6b…`, the signed `fixed_population.py` |
| `RANKING_SHA256` | `8bc2e7cb…`, matching the signed `ranking.py` in `rev8-executed-sources.json` |

## Signed basis (unchanged)

| Item | SHA-256 |
|---|---|
| `PREREGISTRATION.md` (P) | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `oc-manifest.json` (M) | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |

**Governing text used this round:**
- P l.163 (seeds, counters), l.169 (calibration, "Record sys.flags.optimize and refuse non-zero optimization", "Missing/safety use the same rules"), l.171 ("`safety-profile` identities include critical share"; "Run **after all binding rows**"; the frozen scheduling policy), l.181 (binding resource stop rule; "Workers are fixed at 8"; "Failed or unmeasured CPU or memory fit means not testable"), §11.
- M l.10 ("executable path/digest and sys.flags.optimize=0 in every new output; non-zero optimization refused"), l.46–54 (seed recipe; l.51 identity kinds; l.53 "profile includes both, interval-group omits critical_share; construction base excludes both axes"), l.113–149, l.200.
- `validation/compute-projection-rev8.json` `measured_inputs`: gate check 1.911e-05 s; maximum complete trial 0.00114 s.

## Round-2 disposition

C = `controller.py`, T = `test_controller.py`.

| ID | Fixed | Code (C:line) and catching test (T:line) | Residual |
|---|---|---|---|
| R2-01 | **yes** | Per-population lane locks (C:423–425). Slots 0–7 (C:388) plus a busy-slot set (C:433). Shared state lock and meter (C:357–372). `policy.json` `binding_workers: 8`, re-pinned. Test: `test_eight_concurrent_binding_slots_equal_serial_and_refuse_reentry` (T:132): eight trial callbacks held at once by `Barrier(9)`; slot 8 and re-entry on one population are refused. | — |
| R2-02 | **yes** | One state lock covers every commit and stop. After the gate, `_boundary` re-authorises before the durable trial start (C:469–475). The post-trial check is at C:485–487. The in-flight rule is in hashed `policy.json` `inflight`. A single terminate path (C:683–686). Tests: T:184 (gate-window gold, pause and cap never start the trial), T:211 (trial window: cap keeps the result, gold, pause and integrity stop leave it unclassified), T:518 (single termination), `assert_report` (2,016 groups in order, no duplicates, one terminate). | The gate-result classification guard (C:456–457) has no isolating test (R3-05). |
| R2-03 | **yes** | `run` owns dispatch, lanes and criticality (C:596–644). Every journal record is timestamped (C:291–297). `next_group`, `step`, `finish` and `run_criticality` are removed. Test: T:274 checks the removed APIs, the event kinds, timestamps and outcome-independent event projections. | — |
| R2-04 | **yes** | The controller runs the signed `calibrate_population` for all five populations before lane 0 (C:530–568, C:617). It computes the record digest itself (C:549). Records are kept in every snapshot (C:659–663) and reused per base profile across κ (C:533, C:560–568). `mark_unreachable` is removed. Tests: T:303 (unreachable population 3, zero attempts, reuse without recalibration, asserted record pauses), T:343 (the run advances in frozen order). | The construction seed cell is a new, undisclosed author choice (R3-04). |
| R2-05 | **yes** | No asserted-outcome API exists. `resume` (C:510–517) leaves `pending` in place, and the same phase, index and seeds re-execute (C:439–445, C:471–475). Tests: `fault_case` (T:243–266: exactly one re-invocation at the original index; no `adequate` parameter, no `resolve_crash`); T:463 (real `os._exit(17)` mid-safety; replay gives the same index and seeds; a cap after the recovered result keeps it). | — |
| R2-06 | **yes** | One latch, `_fault` (C:411–420), is reached from gate (C:450), trial (C:480), calibration (C:550) and criticality (C:585) via `except BaseException`. `gate_faults` is kept out of the Wilson denominators (C:652–657). Tests: T:268–272, T:381–383, T:546. | `BaseException` is exercised only on the trial and criticality paths (R3-05). |
| R2-07 | **partly** | Fixed directory with `flock` (C:162–171). The create-once freeze includes the absolute origin, `p_gate`, the inventory digest and source digests (C:183–203). Hash-chained journal, appended and fsync'd before apply (C:279–304). Replay, process-loss pause and torn-tail handling (C:204–216, C:251–277). Terminated replay is immutable. Meter faults are recorded without re-reading the meter (C:362–368). Tests: T:385, T:410, T:428, T:463. | A journal I/O failure leaves the writer live and can make the stage permanently unreplayable (R3-01). Per-transition record volume was never costed (R3-02). |
| R2-08 | **yes** | Read-only properties (C:221–240). Immutable nested manifest and `critical_j` tuple (C:150, C:154). `_complete` is private (C:665–669). Test: T:449. | — |
| R2-09 | **yes** | Guards at C:37–43, C:151–152, C:362–365, C:447–448 and C:667–668; interior Wilson at C:85–94. Tests: T:100, T:119, T:269, T:428, T:449. Nine assertion-killed mutants (`round-2-mutant-output.txt`: "Guard mutants killed: 9/9"). | — |
| R2-10 | **yes** | `derive_known_answers.py` imports neither the controller nor the kernel and uses OpenSSL for all 2,016 HMACs and the array SHA. Its output is committed byte-identical as `known_answers.json`. The README now calls the policy digest a checksum. Test: T:100. | The suite checks the committed fixture and does not run the derivation (Main reproduces it; see Verification). |
| R2-11 | **partly** | A canonical string-cell check (C:67–70; T:106–107; recorded mutant). The binding check before safety (C:603–605; T:501). | The check and the first dispatch are separate critical sections (R3-03). |
| R1-03 (carried) | **yes** | Gate-window authorisation, controller-owned pacing and private `terminated`; see R2-02, R2-03 and R2-08. | — |
| R1-04 (carried) | **yes** | Eight binding lanes; the frozen origin is tested (T:428 with origin earlier than construction, killed by the "wall origin" mutant). | — |
| R1-06 (carried) | **yes** | See R2-04. | — |
| R1-10 (carried) | **yes** | Interior Wilson(50, 2000) agrees with 60-digit Decimal to 15 places; the council re-derived it by hand as about [0.019014, 0.032807]. The checksum wording is corrected. | — |

## New issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R3-01 | major | C:279–304 (write :299, fsync :301, `_apply` :302), always called under `_state_lock`; replay C:262–273 | **A journal I/O failure is not latched.** `_emit` has no failure handling.<br>(1) If `write`, `flush` or `fsync` raises (ENOSPC, EIO), or `_apply` raises after the line is written, then `_sequence` and `_previous` do not advance. The `OSError` leaves `step` or `run`; `run` catches only `Stopped` (C:640). Nothing is recorded.<br>(2) Other binding lanes keep committing. Their next record either reuses the same `seq` after a buffered or durable copy of the failed record, or appends to a partial line.<br>(3) On restart, replay raises "journal chain mismatch" or "unreviewed corrupt journal interior" (C:268–273). The run directory can then never reopen, and there is no §11 pause. Binding rows become unrecoverable, against R2-07's "rebuild … and resume" and P l.171 "every pause or termination is recorded".<br>Transient ENOSPC is plausible given R3-02 [INFERENCE]. | **Poison on failure.** Set a failed latch under the state lock before the write. Any exception in write, flush, fsync or `_apply` leaves the latch set. Every later `_emit`, `_boundary`, `step`, `run`, `pause` and `resume` then refuses without writing, and waiters are notified. Do not try to journal a pause with an uncertain sequence. Recovery goes only through restart and replay: a complete final record is authoritative (append-before-apply), and a torn tail uses the existing reviewed process-loss pause.<br>**Tests:** inject failures at write, flush, fsync and `_apply` while a second binding lane is mid-attempt. Assert that no bytes are appended after the failure, that the second lane cannot commit, and that a rebuilt controller replays without a chain error into a process-loss pause, with counters equal to the durable records. Add these to `mutate_round2.py`. |
| R3-02 | major | C:369 (one `meter` record per `_boundary`); `_attempt` emits at C:429, 443, 464, 465, 469, 475, 493, 495; fsync is held under the shared lock | **The journal's volume and serialisation were never costed against the binding workload.**<br>A gate-passing attempt writes 8 fsync'd records and a rejected attempt writes 4. Each population record carries the identity twice.<br>Binding has (16,128 + 1,792 + 252) × 5 = 90,860 populations with at least 2,000 executions each. At the rev-8 projection's p_gate = .5 that is about 2.2 × 10⁹ fsync'd records and about 1 TB of journal (about 710 B per population record, about 170 B per meter record) [INFERENCE, static trace].<br>All fsyncs are serialised by the one state lock that all eight workers share. The work being wrapped is a 19 µs gate check (4 fsync'd records) or a trial of at most 1.14 ms (8 records), so the overhead may exceed the work [INFERENCE].<br>None of this is in the binding fit (P l.181 measures only the benchmark commands). Disk is unbudgeted, and replay must parse every record while the 15-day clock runs. Likely result: binding OC halts on disk exhaustion, or wall time grows by an amount never measured, leaving binding rows unmeasured or hitting the cap. | One recommended design:<br>(a) Fsync only at linearisation points: trial start, trial result, dispatch, completion, pause, resume and termination. Fold the meter reading into those records and drop standalone `meter` records.<br>(b) Write a compact counter delta keyed by a short frozen population index instead of two copies of the identity.<br>(c) Make gate classifications durable at the next fsync'd record. Losing them in a crash is harmless because same-seed re-execution reproduces them (C:439–445).<br>(d) Measure records, bytes and µs per attempt on a synthetic binding lane. Record the projected journal size for 90,860 binding plus 10,080 safety populations at the frozen p_gate in the freeze, and refuse to freeze if the run-root volume has less free space.<br>Tests: pin the number of fsync'd records per gate-passing and rejected attempt; replay equivalence after a crash between gate and trial start; the free-space refusal. |
| R3-03 | minor | C:388, C:603–605, C:606–616 | **Safety admission is not atomic with the binding-completion check.** The check (C:603–605) and the first `dispatch` (C:614) take the state lock separately, with `_wait_ready` in between. `binding_population` refuses only once `_started_safety` is set at dispatch.<br>Sequence: `run()` is called while the controller is paused, passes the check and blocks in `_wait_ready`. On `resume`, a binding worker leases a new population before the run thread dispatches. Safety then starts with an unfinished binding population, and that lease keeps stepping concurrently, against P l.171 "after all binding rows" and the R2-11 required fix. T:501 creates its lease *before* `run`, so it cannot catch this. | Re-check binding completion inside the same critical section that emits the first `dispatch` whenever `_started_safety` is False. Refuse `binding_population` while `_run_lock` is held. Add a deterministic threaded test (paused, `run` waiting, resume, lease) asserting that exactly one of the lease and the safety start succeeds. |
| R3-04 | minor | C:525–528, C:541; `policy.json` `calibration`; README l.19; `derive_known_answers.py:35–36`; `known_answers.json` `seeds[0]` | **The construction seed cell is an undisclosed choice that conflicts with the signed `safety-profile` definition.** Every safety population is seeded from `population_identity("safety-profile", six base dimensions)`, with neither κ nor j.<br>P l.171 says `safety-profile` identities include critical share, and M l.53 says "profile includes both". M l.53 also says the "construction base excludes both axes", but M l.51 has no construction-base kind. The choice is therefore forced, but it is not stated in the hashed policy, the README or the response (R-064).<br>It is also not independently pinned: the only population-phase known answer uses an interval-group cell that the controller never seeds construction with. A reproducer following P l.171 would derive different seeds for all 1,680 constructions, and nothing would catch it. | **Council verdict on the choice: keep the current bytes.** They use only the signed constructor and a listed kind, and their dimension set is exactly M l.53's construction base. A new kind would need an unsigned code path.<br>Required: state in hashed `policy.json` and in the README that this is the κ/j-free construction-base seed cell, not an outcome-cell `safety-profile` identity, citing M l.53, and re-pin `POLICY_SHA256`. Add an independent OpenSSL population-phase vector for one construction base to `derive_known_answers.py`, and test that the seed passed at C:541 equals it. Test that there are 336 distinct bases, each shared by exactly six κ groups. |
| R3-05 | minor | C:456–457, C:450, C:550, C:431–432 | **Four guards have no test that would fail if they were removed [INFERENCE, static trace].**<br>(a) **Gate-result classification after a stop or pause (C:456–457).** Delete it and `_boundary` at C:465/469 still raises, so T:184 passes on `executions == 0` and `called == []`. Yet the gate result would be classified after gold or a pause, and a pause plus resume during the gate would let the trial run. That contradicts hashed `policy.json` `inflight`.<br>(b) **`except BaseException` on gate (C:450) and calibration (C:550).** A mutant to `except Exception` survives, because only trial (T:272) and criticality (T:383) raise `KeyboardInterrupt`. R2-06 required one test per path including `BaseException`.<br>(c) **The safety-step guard (C:431–432)** is never exercised. | Add isolating tests:<br>(a) Gold, cap, pause, and pause plus resume during the gate: assert `gate_passes == 0`, pending phase `gate`, no `population_state` after `terminate`, and the gate is re-invoked after resume.<br>(b) `KeyboardInterrupt` from gate and from calibration latches a pause with an incident.<br>(c) `BindingLane(s, safety_key).step` outside an active `run` raises.<br>Add the four reversion mutants to `mutate_round2.py` and record that each is killed by an assertion. |
| R3-06 | minor | Freeze C:183–186; report C:671–681; compare `locked-size-input/locked_commands.py:175–178, 325–326` and `validation/ranking.py:391–396` | **No interpreter provenance and no refusal under `-O`.** The controller now invokes the numerical kernel (C:542), so P l.169's "Record sys.flags.optimize and refuse non-zero optimization" applies directly. M l.10 requires "executable path/digest and sys.flags.optimize=0 in every new output; non-zero optimization refused". The freeze, journal and report record neither, and `python -O` proceeds. | Before freezing, refuse any interpreter other than CPython 3.11.11 and any non-zero `sys.flags.optimize`. Add executable path, executable SHA-256, version and `python_optimize` to the freeze's `expected` block, so replay under a different interpreter refuses. Test in a subprocess: under `-O`, construction raises and no `freeze.json` exists. |

**Counts:** 0 blocking, 2 major (R3-01, R3-02), 4 minor (R3-03 to R3-06).

### Rejected or not raised

- **The caller-supplied prepared population as a skip lever** (peer N-01, main claim): not an issue.
  - Callbacks are the work functions, and the concrete product harness is reviewed separately (M l.200).
  - The same lever already exists through a gate callback that always returns False.
  - R2-04 required removing asserted *verdict records*, and that is done.
  - Having the controller construct populations needs the authentic locked-size input, which round 9 lists as a blocking prerequisite. The binding part of N-01 is suggestion 3.
- **A pause plus reviewed resume between the gate commit and the trial start** (challenger R3-C2): a suggestion, not an issue.
  - The window between two lock acquisitions is microseconds, and a reviewed resume cannot plausibly happen inside it.
  - The effect is fail-closed: the trial re-executes with the same seed.
- **A hand-built `BindingLane` stepping a safety lane** (peer N-05): it needs the private `_owner`/`_key` fields, so it is treated like round 2's deliberate-misuse items. The untested guard is folded into R3-05(c); hardening it is suggestion 4.
- **A binding identity colliding with a safety identity** (peer N-06): it needs a harness to lease a `safety-interval-group` identity as a binding cell. Suggestion 5.
- **Self-asserted pause/resume evidence:** a risk already accepted under R1-04(b). No new evidence.

## Non-blocking suggestions

1. Add an epoch check under the lock immediately before the trial-start transition (C:469–475) (challenger R3-C2).
2. P l.171 says "after all binding rows", but `policy.json` `start` covers only *leased* populations. Because `binding_population` refuses after dispatch (C:388), an early `run()` permanently strands binding cells that were never leased. Consider requiring the leased binding population count to equal the signed 90,860 before the first dispatch.
3. Pass the controller's calibration record (`latent_gain`, `record_sha256`) to `trial` and `criticality`, and journal a digest of each prepared population. Today the harness must recalibrate (against M l.142's one calibration) or read `events` to find the gain the trials should use. When the locked-size input exists, have the controller call the signed `prepare_population` itself.
4. Bind the safety-step guard (C:431) to the run thread's identity rather than to `_run_lock.locked()`.
5. Accept binding identities only of the binding kinds in M l.51, refuse a key collision whose `safety` flag differs, and require exactly five populations for `measured`.
6. On macOS, use `fcntl(F_FULLFSYNC)` for real durability. Losing the tail is already safe because the same seeds re-execute.
7. Check `sys.modules["ranking"].__file__` and its digest. `ranking` is a generic module name, unlike the private name that round 2 rejected.
8. Record the run root's real path in the freeze. Root injection is accepted for tests but not enforced.
9. Use a clock that advances in the test fixtures (T:29–31). Today every timestamp is identical, so T:296 checks only types.
10. Optionally run `derive_known_answers.py` from the suite.

## Author choices (R-064)

None are posed back as open choices.

**Accepted:**
- Durable trial start as the linearisation point.
- A cap keeps a returned result; gold, pause or integrity stop leaves work unclassified.
- A torn tail is preserved and needs a reviewed resume.
- A clean reopen records a process-loss pause (fail-closed).
- Completion wins when binding reaches both limits at once.
- Any unreachable population makes the group unmeasured.
- A re-executed trial counts as one execution.
- Gate faults are counted separately and re-executed without a new attempt.

**Not accepted as sized:** per-transition fsync'd records including meter readings (R3-02).

**Accepted on condition of disclosure:** the construction seed cell (R3-04).

## Recommendation

Keep the round-2 architecture. Fix R3-01 and R3-02 together with one change to `_emit`: poison the writer on failure, and fsync compact records only at linearisation points. Then close R3-03 to R3-06.

**Alternatives considered:**
- **Truncate to the last good offset and journal a pause on an I/O failure** (peer): rejected. After an fsync failure the bytes may already be durable, so truncation can destroy evidence, and journaling with an uncertain sequence repeats the defect.
- **A SQLite WAL journal:** transactional, but adds a dependency and is still bound by fsync.
- **Add only a disk preflight:** rejected. It still leaves about 1 TB of journal and serialised fsync [INFERENCE].

**Risks:**
- The R3-02 numbers are [INFERENCE] until measured on a synthetic lane.
- The fix relies on gate and trial determinism given seeds, which the keyed seed recipe already assumes.
- This is a draft only, and fully reversible.

**Execution outline:**
1. Add the `_emit` failure latch and its replay semantics, with the R3-01 tests.
2. Compact the linearisation-point records, fold in the meter and add the disk preflight, with record-count and replay tests.
3. Make the binding check atomic with dispatch (R3-03).
4. Disclose and pin the construction seed cell, re-pin the policy digest and add the independent vector (R3-04).
5. Add the R3-05 tests and mutants.
6. Add interpreter provenance and the `-O` refusal (R3-06).
7. Measure synthetic per-attempt records, bytes and µs, and commit the full outputs and digests.

## Peer contributions

**Challenger (GPT Astra):**
- Found the admission race (folded into R3-03).
- Found the unlatched journal-write failure (R3-01).
- Argued for poisoning the writer over journaling an ordinary pause; the council adopted this.
- Found the stale-epoch gap (downgraded to suggestion 1).
- Noted that the suite does not run the derivation.

**Experiment peer (Opus):**
- Found the construction-seed contradiction (R3-04). The council rates it minor, not major: the signed text forces a choice, the contradiction is in descriptive safety only, and the council accepts the bytes once disclosed.
- Costed the journal volume (R3-02).
- Gave the paused-run interleaving that made R3-03 plausible. The lead had first treated the race as suggestion-level.
- Found the untested gate re-check and the `BaseException` paths (R3-05).
- Found the missing interpreter and `-O` provenance (R3-06). The lead had not considered this, and it changed the issue count.
- N-01, N-05 and N-06 were moved to "Rejected or not raised" or to the suggestions.

**Lead:**
- Independently found the I/O desynchronisation and the duplicate-sequence replay refusal.
- Grounded R3-02 in the measured rev-8 inputs (19 µs gate, 1.14 ms trial).
- Found the construction-base seed issue and decided the council's verdict on it.
- Found the unleased-binding stranding (suggestion 2) and the unbound latent gain (suggestion 3).

## Verification

- **No shell was used in this council session or either peer session.** Every race, mutant-survival and resource claim here is a static trace [INFERENCE].
- **Main's independent run** at `38cead90` (clean tree, pinned `.tools/venv/bin/python` 3.11.11, `PYTHONDONTWRITEBYTECODE=1`): `oc-controller/test_controller.py` ran 25 tests in 2.175s, OK. File digests are Main's.
- The mutant results (9/9 killed) and the derivation output are the author's records. The council did not rerun them.
- **Files read:** all of `controller.py`, `test_controller.py`, `README.md`, `policy.json`, `derive_known_answers.py`, `mutate_round2.py`, `known_answers.json`, `round-2.md` and `round-2-response.md`; P l.161–186; M l.1–19 and l.37–201; `fixed_population.py:150–210`; `compute-projection-rev8.json` `measured_inputs`.
- **Main should run before round 4:**
  - `shasum -a 256` on the revised files;
  - the README test command;
  - `derive_known_answers.py`, compared byte-for-byte with `known_answers.json`;
  - `mutate_round2.py`, extended with the R3-01, R3-03 and R3-05 mutants;
  - the synthetic per-attempt records, bytes and µs measurement for R3-02.

## Verdict

**6 issues open** (0 blocking, 2 major, 4 minor). R2-07 and R2-11 are only partly fixed, and four guards are not isolated by tests. The draft is not accepted for its stage-3 role. A revised draft needs a new scoped round with zero issues, and that round must pin the re-derived policy SHA-256, the construction-base seed vector and the source digests. Acceptance would still confer no run authority.
