# Stage-3 OC controller — Option G draft

**DRAFT, synthetic only. No authority until this artefact's scoped council review has zero issues.** No preparation action (a)–(e), scans, models or real-repository data was used. No binding OC was executed.

Implements PREREGISTRATION.md lines **163, 169, 171, 175, 181, 185, 224**; round 8 X1-R8-02, round 9 §§8–9 (including the ordered-group SHA-256 at freeze).

- `controller.py`: independently enforces attempts ≤ceil(4000/p_gate) and gate-passing executions ≤4000. Rejections never invoke the trial callback. Inadequacy consumes executions. Binding stops at 2,000 eligible; safety stops at 2,000 executions, including inadequate. Monotone attempt indices feed the manifest's phase-separated HMAC seed recipe.
- `safety_groups()` enumerates exactly 2,016 common-mask groups: 336 complete and 1,680 full-vertex. `SafetySchedule` freezes ascending keyed order, its ordered-list SHA-256 and the policy SHA-256 before dispatch. It rejects subsequent order/policy mutation. Criticality completion precedes advancing to the next group.
- Reports separate measured, attempt-exhausted `unmeasured; cell fails`, and `not completed by gold`, preserving exhaustion precedence, partial counters and pooled-measured-group membership. The scheduler accepts resource/time/gold controls but **no safety outcomes**.
- `test_controller.py`: independent counter exhaustion, rejection non-execution, inadequate safety stopping, complete group dimensions, independently computed order/digest, criticality sequencing, outcome-blind cap termination, gold reporting and frozen-policy mutation refusal.

`Population.step(gate, trial)` is the execution boundary; callbacks receive the fixed attempt index. Resource inputs to `next_group` are cumulative **binding plus safety**, already charged with the 2× factor. Caller must reserve work before invoking it; the controller never grants extra capacity. There are no subprocess/model interfaces.

## Open choices for review

1. Start immediately after all binding rows have completed (not after a favourable safety result). Reserve eight workers and dispatch **one group at a time**, with canonical-criticality complete before the next group. No worker-count adaptation.
2. No elective pauses. Terminate on gold or the shared 2,880-core-hour/15-OC-day boundary; record cause and current group. Integrity incidents require affected-work review under §11, not an outcome-conditioned scheduling exception.
3. HMAC ties break by canonical identity bytes; SHA-256 of the ordered list uses compact sorted-key JSON of the ordered group objects. Freeze policy separately as compact sorted-key JSON. Review these concrete serialization conventions.

## Synthetic validation

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/oc-controller/test_controller.py
```

From repository root: **5 tests passed** (CPython 3.11.11). Synthetic callbacks only; no audit/bootstrap/real OC was run. Baseline unavailable. The review must bind the final ordered-list digest, policy and exact source digests before real safety output.
