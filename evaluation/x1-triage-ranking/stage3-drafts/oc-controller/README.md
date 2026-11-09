# Draft stage-3 OC controller

**Synthetic-only, non-authoritative draft.** Implements the P l.163 counters/seeds, l.169 calibration disclosure, l.171 safety order/scheduling/statuses, l.181 fixed eight binding workers, l.185 actual OC caps, and §11 integrity pauses/duplicate-stage refusal. It carries no authority until its scoped review reports zero issues; acceptance itself does not authorize preparation or a real run. No candidate data, scan or model call is used here.

## One controller, one durable stage

`SafetySchedule(p_gate=..., inventory_sha256=..., meter=..., oc_started_at=..., clock=..., run_root=...)` owns the fixed `x1-oc-controller/` directory under the root. The default root is this draft folder; root injection is for isolated synthetic tests, not a second production-stage identifier. A process lock prevents concurrent owners. CPython 3.11.11 with `sys.flags.optimize == 0` is mandatory before freezing. The create-once freeze binds executable real path/SHA-256/version/optimization, absolute full-OC origin, p_gate, inventory digest, source digests, ordered identities and policy checksum. Replay under different interpreter provenance refuses.

The hash-chained journal uses compact counter deltas keyed by a short frozen population index. Gate starts/classifications are coalesced in memory until the next durable point. Trial start and result each write exactly one fsync'd record; a rejected attempt writes none. Control transitions (including dispatch, completion, pause, resume and termination) are durable. Meter readings are folded into records, not separate meter records. Lost gate deltas are harmless: replay repeats the same deterministic gate index/seeds. Counter records do not repeat population identities. Population creation fixes the index mapping in journal order.

Any write (including a short write), flush, fsync or reducer failure poisons the writer under the shared state lock and wakes waiters. All subsequent work refuses **without appending a pause to an uncertain chain**. Only restart/replay recovers: a complete final record is authoritative even if in-memory application failed; a torn tail is preserved by digest in a reviewed process-loss pause. A terminated replay is immutable. Counter commits stay on disk, not duplicated in RAM.

A final JSON record that parses and chains is authoritative even if only its newline was lost. Replay hashes it with the restored newline, then restores that delimiter before the process-loss pause; its `tail_sha256` is null. Only an unparseable or non-chaining tail is torn. The fault test includes `raw[:-1]` and two consecutive restarts, comparing every durable population field.

Freeze records a conservative disk reservation for all 90,860 binding and 10,080 safety populations at frozen p_gate (4,000 execution cap, two records per execution plus control allowance). Insufficient run-volume free space writes `freeze-refused.json` with the projection and observed free bytes, and creates no `freeze.json`. The reservation is deliberately separate from the benchmark's measured, sample-scaled estimate; it is not a claim of measured full-OC fit.

Raw actual cumulative binding-plus-safety core-hours are compared with 2,880 using `>=`; ×2 is solely the separate projection convention. Unpaused wall time starts at the frozen first-full-OC origin and stops at 15 days using `>=`. Meter faults take a recorded integrity path that does not reread the faulty meter. Decision state and the nested runtime manifest are exposed only as read-only values; `finish`, `next_group` and `mark_unreachable` no longer exist publicly.

## Binding lanes and the controller-owned safety loop

`binding_population(canonical_cell_json, population_index, slot=0..7)` returns a read-only lease with `step(gate, trial)`, `report()`, `complete` and `exhausted`. Each population has its own nonreentrant lock and fixed worker slot. Eight binding slots may execute concurrently under one shared meter/state lock. A ninth slot, a reused live slot, or concurrent calls on one population refuse. Binding stops at 2,000 eligible trials, 4,000 executions or `ceil(4000/p_gate)` attempts. Completion wins if the last execution simultaneously reaches both limits.

`run(gate, trial, calibration, criticality)` owns safety pacing through termination; no caller-driven per-step safety handle is exposed. Acquiring the run lock closes binding admission, even while paused, and the completion check is repeated atomically with first dispatch. Every leased binding population must already be complete/exhausted and free of pending faults. Gold arrives asynchronously through `stop_gold()`; integrity stop through `stop_integrity(evidence)`. The loop owns group dispatch, increasing-index serial safety lanes 0–4 (slots 5–7 reserved idle), and four serial criticality phases. Callback signatures:

- `calibration(profile_identity, population_index, population_seed)` returns **prepared population data**, not an asserted reachability record. The controller calls the digest-pinned signed kernel's `calibrate_population` itself. All five populations are calibrated before the first attempt. It retains identity, population index, target, realised, excess, latent gain, reachability, kernel SHA and the controller-computed record digest. The first five records per base profile are reused across all six κ groups; conflicting disclosures refuse. Any unreachable population makes the group `unmeasured (unreachable target)` with actual zero counters and advances the frozen order.
- `gate(group_identity, population_index, attempt, audit_seed)` returns an exact bool. Rejection performs no trial work.
- `trial(group_identity, population_index, attempt, seeds)` returns an exact adequacy bool; `seeds` contains audit/bootstrap/missing-mask/canonical-safety seeds. Safety stops at 2,000 executions including inadequacy, independent of labels/outcomes.
- `criticality(group_identity, j, key)` returns a JSON object with the exact signed `key` and `completed: true`; all additional descriptive outcome fields are retained. All four j records are required for `measured`.

The journal timestamps dispatch/completion, lane start/stop, calibration and criticality start/result (the result is the normal phase stop), and explicit phase stops at interruptions. The source of identities is the pinned signed `population_identity`; group dimensions are the six manifest dimensions plus `unresolved_kappa`, never `critical_share`. Order is ascending HMAC-SHA256 (`169-x1-safety-order`) over exact identity UTF-8 bytes with a byte tie-break. The freeze hashes the compact JSON **array of identity strings**. String seed cells must themselves be compact sorted-key canonical JSON; phases and integer indices are strict.

**Construction seed disclosure (M l.53).** Calibration uses the existing signed `safety-profile` kind with exactly the six construction-base dimensions, excluding κ and j. This is the κ/j-free **construction-base seed cell**, not the outcome-cell `safety-profile` identity (which includes `critical_share`). The council required keeping these current seed bytes rather than inventing a new unsigned identity kind. The hashed policy records the choice. There are 336 distinct construction bases, each shared by six κ groups; the independent OpenSSL vector pins the actual seed passed to population preparation.

## Atomic stops and integrity recovery

One state lock serializes every counter commit, pause, completion, termination and snapshot. It is short-lived and never held over callback work. After a gate returns, the controller reauthorizes before the durable trial-start transition (the defined linearization point).

The hashed policy fixes the in-flight rule: a newly observed cap **after** a trial returns retains its result. Gold, an integrity stop or a pause during an attempt leaves it charged and unclassified; no trial starts after a gate-window stop/pause. Exactly one termination freezes all 2,016 identities in order, without duplicates. A later callback cannot change that report.

Gate/trial/calibration/criticality faults, malformed results and `BaseException` all latch the same integrity mechanism. Gate faults have their own count and do not masquerade as gate rejections or enter a measured gate-failure denominator. `resume({"reviewed": true, "evidence_sha256": ...})` journals the reviewed resolution. The owned loop then reruns the **same pending phase/index with the same seeds**; a binding lease likewise retries its pending phase on the next `step`. There is no `adequate` parameter or asserted-outcome resolution API. Paused in-flight work is retried, never classified by an operator.

## Independent derivation and validation

From the repository root after `source .tools/activate`:

```sh
D=evaluation/x1-triage-ranking/stage3-drafts/oc-controller
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/derive_known_answers.py"
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/test_controller.py" -v
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/mutate_round2.py"
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python "$D/benchmark_journal.py" --attempts 1000 --p-gate 0.5
```

`derive_known_answers.py` imports neither controller nor signed kernel. It independently enumerates the seven identity axes, invokes OpenSSL for every order HMAC and final array SHA, reproduces five seed vectors (including the construction base), and computes an interior Wilson answer using 60-digit Decimal. Its stdout is the reproducible current `known_answers.json`; preserve each round's full derivation output separately. Historical round-2 evidence remains unchanged. The policy digest is a **checksum of the published policy**, not an independent design derivation.

`reviews/round-3-response.md` maps the six findings and R2-07/R2-11 residuals to implementation/test lines, commands, complete observed output and file SHA-256s. Historical round-2 records remain unchanged. Raw suite and scratch-only guard-mutant output are preserved alongside the response. Tests include eight concurrent binding lanes, stop-window races and invariant snapshots, actual process loss mid-safety and same-seed replay, criticality/gate/BaseException faults, counter boundaries, unreachable population 3, immutable state, meter origins/decreases and independent pins. Some transition tests seed feasible prior counter states through the private journal reducer; they are explicitly unit fixtures, not claims that full OC ran.

`benchmark_journal.py` uses deterministic synthetic callbacks and real temporary-file writes/fsyncs, with no controller-side randomness. It reports per-attempt records/bytes/microseconds, passing/rejected/mixed cases, final deferred-gate flush separately, interpreter/source provenance, population-scaled estimates at the supplied frozen p_gate, and the conservative freeze reservation. Its sample size is instrumentation, **never a reduced scientific trial count**. Only this isolated synthetic fixture mocks full-run disk admission; production admission remains enforced. This is journal I/O evidence, not full benchmark CPU/memory fit. Reproduce at the actual frozen p_gate when that input exists; default `.5` is the signed rev-8 projection's synthetic comparison point.

Round-3 isolating tests cover partial/short write, flush, fsync and apply failures with a second lane in flight; no post-failure append and replay authority; compact fsync counts and lost gate replay; disk refusal; paused-run admission; gate-window gold/cap/pause/pause+resume classification; gate/calibration `KeyboardInterrupt`; safety-handle refusal; 336×6 construction sharing; optimization/version refusal and provenance mismatch. `mutate_round2.py` adds seven named R3 reversion probes to the nine preserved round-2 probes. Parent verification generates current known answers before running suite/mutants and preserves all raw outputs; no checks were run during this implementation.

Round-4 fixes and complete validation evidence are recorded in `reviews/round-4-response.md`. Runtime-refusal subprocesses mock ample disk space and require the exact interpreter diagnostic for optimization, version and non-CPython cases. Four additional assertion-killed reversion probes cover these guards and newline-less complete-record replay; the existing policy, seed bytes and disk reservation are unchanged.

## Open choices for review

No choices are deferred to an operator. **Conservative choice for council review:** reserve 2,048 bytes per record, 4,000 executions and 16 additional control records per population across all 100,940 populations, rather than admitting from the sample-scaled estimate. This fixes a 1,657,108,561,920-byte reservation at p_gate .5 and may refuse a volume that could hold the measured extrapolation; review must approve this deliberately conservative reservation. The recommended journal design and torn-tail handling preserve evidence and require reviewed resume; recovery never resets the origin or counters. Real trust anchors, a production run root and stage authority remain later scoped prerequisites, not authority supplied by this synthetic draft.
