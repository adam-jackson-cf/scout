# Draft stage-3 OC controller

**Synthetic-only, non-authoritative draft.** Implements the P l.163 counters/seeds, l.169 calibration disclosure, l.171 safety order/scheduling/statuses, l.181 fixed eight binding workers, l.185 actual OC caps, and §11 integrity pauses/duplicate-stage refusal. It carries no authority until its scoped review reports zero issues; acceptance itself does not authorize preparation or a real run. No candidate data, scan or model call is used here.

## One controller, one durable stage

`SafetySchedule(p_gate=..., inventory_sha256=..., meter=..., oc_started_at=..., clock=..., run_root=...)` owns the fixed `x1-oc-controller/` directory under the root. The default root is this draft folder; root injection is for isolated synthetic tests, not a second production-stage identifier. A process lock prevents concurrent owners. The create-once `freeze.json` binds the absolute first-full-OC origin, p_gate, inventory digest, source digests, ordered identity-list digest and operational-policy digest. An existing freeze is verified, never replaced; contradictory origins, inputs or sources refuse.

The hash-chained `journal.jsonl` is flushed and fsync'd **before** each transition is applied. Replay restores the same attempt, phase, counters, calibration disclosures and stops. A nonterminal replay records a process-loss integrity pause requiring reviewed resume. A torn trailing record is preserved byte-for-byte and identified by digest in the recovery pause; unexplained interior corruption refuses. A terminated replay returns the identical frozen report without dispatching. Counter commits remain in the durable journal, not millions of duplicated in-memory event objects. Reports retain all population counters and control events.

Raw actual cumulative binding-plus-safety core-hours are compared with 2,880 using `>=`; ×2 is solely the separate projection convention. Unpaused wall time starts at the frozen first-full-OC origin and stops at 15 days using `>=`. Meter faults take a recorded integrity path that does not reread the faulty meter. Decision state and the nested runtime manifest are exposed only as read-only values; `finish`, `next_group` and `mark_unreachable` no longer exist publicly.

## Binding lanes and the controller-owned safety loop

`binding_population(canonical_cell_json, population_index, slot=0..7)` returns a read-only lease with `step(gate, trial)`, `report()`, `complete` and `exhausted`. Each population has its own nonreentrant lock and fixed worker slot. Eight binding slots may execute concurrently under one shared meter/state lock. A ninth slot, a reused live slot, or concurrent calls on one population refuse. Binding stops at 2,000 eligible trials, 4,000 executions or `ceil(4000/p_gate)` attempts. Completion wins if the last execution simultaneously reaches both limits.

`run(gate, trial, calibration, criticality)` owns safety pacing through termination; no caller-driven per-step safety handle is exposed. Every leased binding population must already be complete/exhausted and free of pending faults. Gold arrives asynchronously through `stop_gold()`; integrity stop through `stop_integrity(evidence)`. The loop owns group dispatch, increasing-index serial safety lanes 0–4 (slots 5–7 reserved idle), and four serial criticality phases. Callback signatures:

- `calibration(profile_identity, population_index, population_seed)` returns **prepared population data**, not an asserted reachability record. The controller calls the digest-pinned signed kernel's `calibrate_population` itself. All five populations are calibrated before the first attempt. It retains identity, population index, target, realised, excess, latent gain, reachability, kernel SHA and the controller-computed record digest. The first five records per base profile are reused across all six κ groups; conflicting disclosures refuse. Any unreachable population makes the group `unmeasured (unreachable target)` with actual zero counters and advances the frozen order.
- `gate(group_identity, population_index, attempt, audit_seed)` returns an exact bool. Rejection performs no trial work.
- `trial(group_identity, population_index, attempt, seeds)` returns an exact adequacy bool; `seeds` contains audit/bootstrap/missing-mask/canonical-safety seeds. Safety stops at 2,000 executions including inadequacy, independent of labels/outcomes.
- `criticality(group_identity, j, key)` returns a JSON object with the exact signed `key` and `completed: true`; all additional descriptive outcome fields are retained. All four j records are required for `measured`.

The journal timestamps dispatch/completion, lane start/stop, calibration and criticality start/result (the result is the normal phase stop), and explicit phase stops at interruptions. The source of identities is the pinned signed `population_identity`; group dimensions are the six manifest dimensions plus `unresolved_kappa`, never `critical_share`. Order is ascending HMAC-SHA256 (`169-x1-safety-order`) over exact identity UTF-8 bytes with a byte tie-break. The freeze hashes the compact JSON **array of identity strings**. String seed cells must themselves be compact sorted-key canonical JSON; phases and integer indices are strict.

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
```

`derive_known_answers.py` imports neither controller nor signed kernel. It independently enumerates the seven identity axes, invokes OpenSSL for every order HMAC and the final array SHA, independently reproduces four seed vectors, and computes an interior Wilson known answer using 60-digit Decimal. Its complete output is committed as `reviews/round-2-derivation-output.json` and as the current `known_answers.json`. The policy digest is explicitly a **checksum of the published policy**, not an independently derived design claim. `policy.json` specifies the reviewed operational choices and is checksum-pinned in the controller.

`reviews/round-2-response.md` maps all findings and named round-1 residuals to implementation/test lines, commands, complete observed output and file SHA-256s. Raw suite and scratch-only guard-mutant output are preserved alongside it. Tests include eight concurrent binding lanes, stop-window races and invariant snapshots, actual process loss mid-safety and same-seed replay, criticality/gate/BaseException faults, counter boundaries, unreachable population 3, immutable state, meter origins/decreases and independent pins. Some transition tests seed feasible prior counter states through the private journal reducer; they are explicitly unit fixtures, not claims that full OC ran.

## Open choices for review

None are deferred to an operator. The round-2 recommended design is implemented here. The conservative torn-tail handling preserves evidence and requires reviewed resume; no recovery resets the origin or counters. Real trust anchors, a production run root and stage authority remain later scoped prerequisites, not authority supplied by this synthetic draft.
