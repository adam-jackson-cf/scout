# Stage-3 OC controller — Option G draft, round-1 fixes

**DRAFT; no authority until a new zero-issue scoped review.** All validation is synthetic. No preparation action, real repository data, inventory, model, scan or binding OC was used. Acceptance of this draft alone cannot authorise any later stage.

Implements the counter/seed portions of P163, unreachable-safety handling in P169, safety ordering/scheduling/statuses in P171, actual shared caps in P185 and §11 integrity pauses. It does not implement P175's binding gate or P181's locked resource remeasurement.

## Signed identities and immutable freeze

`controller.py` imports the signed `fixed_population.population_identity` and `criticality_key` by explicit path after checking source SHA-256. An interval group's dimensions are **exactly** `cap_M`, `prevalence`, `realised_gain`, `scanner_control_true_at_prevalence`, `rank_structure`, `size_dependence`, `unresolved_kappa`. Fixed factors and critical share are not identity dimensions.

The constructor verifies the exact OC-manifest byte SHA-256 (`1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d`) and enumerates all **2,016** identities: 336 κ=0 and 1,680 κ>0. HMAC uses each signed identity's UTF-8 bytes with key `169-x1-safety-order`; ties use identity bytes. The ordered-list digest hashes the compact JSON array of **identity strings**, without re-parsing them.

Pinned values, independently computed with OpenSSL and stored in `known_answers.json`:

- Ordered list SHA-256: `04aa2184f79cb3795f60305a6b8fa31687d5acd7069846e78b6acf1b1f4dd4ff`.
- Policy SHA-256: `e2c8c9ec1da43b4eeface3cf3c52c711ecad8cef11be97c00defe51fc12353ca`.

`SafetySchedule(freeze_path=..., p_gate=..., meter=..., oc_started_at=..., clock=...)` refuses mismatches and writes a **create-once** freeze record containing both digests, source digests and timestamp before dispatch. The list, policy and nested policy values refuse mutation. That synthetic freeze is not an experiment stage seal or approval.

## Operational policy (hashed in source)

- Start immediately after all binding rows complete. One group at a time in frozen order.
- Reserve eight worker slots. Population i uses slot i (0–4); slots 5–7 remain reserved idle. Execute population lanes synchronously in increasing index. A helper is a child of its active lane. No overlapping/reentrant attempt or criticality call is permitted. This is the conservative fixed allocation; it does not promise eightfold speedup or full utilisation.
- Meter **raw measured cumulative binding-plus-safety core-hours**. Wall time is seconds since the **first full-OC process start**, never paused. There is **no ×2 multiplier on actual meters**; ×2 remains the preregistered resource-projection factor.
- Check meters before every attempt, dispatch, criticality phase and completion, and after executed work. Stop at `>=2880` actual core-hours or `>=15` actual wall days. `stop_gold()` is irreversible. Grid completion and §11 integrity termination also record causes. At termination, identity-bound counters and statuses are snapshotted; later calls cannot make a partial group measured. Work interrupted by a stop remains charged and unclassified, never silently added to eligible results.
- `pause(evidence)` requires an outcome-independent §11 integrity kind and evidence digest. `resume(reviewed_resolution)` preserves group/attempt order; clock continues. Every pause/resume/termination records its cause and time. Trial crashes/non-bool adequacy latch an integrity pause. Completion remains refused until reviewed resolution reconciles `executions == eligible + inadequate + crashed`.
- Criticality follows all five population lanes and runs j=(0,.1,1,3) in that fixed order. Completion records use the signed `criticality_key(identity,j)`; all four records are required before automatic group completion/next dispatch.

## Counters, statuses and reporting

Population counters cannot be supplied to the constructor. Execution leases bind `(group identity, population index)`; stale, duplicate, foreign and incomplete populations are refused. Attempts stop at `ceil(4000/p_gate)` independently of the 4,000-execution limit. Gate rejection never calls the trial. Inadequate/crashed executions remain charged. Binding stops at 2,000 eligible; safety at 2,000 executions including inadequacy. Completion on the last permitted counter wins over exhaustion.

Final reports retain all five population snapshots, counters, Wilson intervals, criticality records and calibration provenance. Statuses are `measured`, `unmeasured; cell fails`, `unmeasured (unreachable target)` and `not completed by gold`. Exhaustion precedence applies only to the active group's own populations. Unstarted groups have no invented counters. Unreachable profiles require identity/target/source/digest-bound calibration records and advance without reordering. Only measured groups enter pooled membership.

## Independent known answers and reproduction

`known_answers.json` records full OpenSSL HMACs, the first/last frozen identities, ordered-list/policy digests and seed64 vectors. Seed messages use identity bytes, NUL separators, zero-based decimal indices, manifest phases and first64 big-endian conversion. Population-phase attempts must be zero. Tests reproduce HMAC vectors with `openssl dgst -sha256 -hmac` over explicit stdin bytes, independently of production `canonical`.

```sh
source .tools/activate
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/oc-controller/test_controller.py -v
```

**Result: 19 tests passed in 0.795 seconds.** Full exercised output, issue-to-fix/test mapping and per-file SHA-256s are recorded in `reviews/round-1-response.md` and `reviews/round-1-test-output.txt`. The suite exercises exact counter stops, all group identities, known-answer seeds/digests/Wilson endpoints, execution-boundary caps/gold, immutable snapshots, criticality/membership refusal, integrity pauses/crashes, unreachable progression/grid termination and outcome-independent scheduling. An initial hand-entered Wilson fixture was corrected using an independent 60-digit Decimal calculation of z²/(2000+z²); the decision rule was unchanged.

## Conservative choices submitted for review

1. Synchronous population lanes and criticality, with eight reserved slots and no outcome-driven throttling or reassignment.
2. Raw actual meters; an unpaused wall clock from first full-OC start; §11 evidence-only pauses/resolution and irreversible snapshot termination.
3. Compact JSON array of signed identity strings; byte tie-break; separately pinned operational-policy digest; create-once synthetic freeze before dispatch.
4. Crash resolution must be externally reviewed, recorded and reconciled before resume; it is never an automatic retry or a licence to invent a missing outcome.
