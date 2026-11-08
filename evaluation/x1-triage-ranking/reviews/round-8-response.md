# X1 round-8 response — revision 9 / ADR-014 round 7 submission

2026-11-08. Both fixes adopt round 8's exact recommended text under R-064. This response cannot close council findings or grant authority; preparation remains unauthorised pending a joint zero-issue round. No executable was changed.

| Issue | Recommended fix implemented | Evidence |
| --- | --- | --- |
| X1-R8-01 | Replace the locked-remeasurement suffix with round 8 §5's verbatim text: bind the remeasurement and binding-OC host, record host identity, preserve revision-7 benchmark/calibration replicate structure, allow only reviewed locked-size source substitution, classify resource-exhaustion crashes as failed fit, and classify operator termination/host load as non-integrity failures. Qualifying host faults need contemporaneous independent OS evidence; retain interrupted output and evidence. Nonqualifying termination leaves fit unmeasured; failed fit never permits rerun; completed-run maxima remain binding. | Exactly mirrored in protocol §8, ADR-014 compute row, OC manifest and submission manifest. README and R-081 use the same new suffix. `revision9-checks.json` verifies literal adoption, one operational protocol rule and all 13 reviewed executable digests unchanged. Protocol §12's timing-drift sentence is unchanged. |
| X1-R8-02 | Insert the exact outcome-independent scheduling sentence after “with no reordering after the first output.”: freeze start, worker allocation, pauses and termination with the reviewed stage-3 OC controller before first safety output; none may depend on partial/completed safety outcomes; record every pause/termination and cause. | Verbatim in protocol §8, ADR-014 safety row, OC manifest and README. Existing keyed order, group-completion/exhaustion/gold statuses and non-gating descriptive role remain unchanged. `revision9-checks.json` verifies all four literal copies and placement. |

## Checks and preservation

Main explicitly requested the README self-test and supplement reruns despite round 8 recommending no rerun for these text-only fixes. Both use create-once revision-9 output paths and explicit pinned `.tools/venv/bin/python` 3.11.11:

- README self-test: **249 checks passed**, **11.658710999996401 seconds**, **exit 0**; planted effect **support**, planted null **not supported**. `synthetic-results-rev9.json` and `strong-effect-OC-rev9.json` preserve this requested run. Strong effect: **2,000/2,000** support, Wilson MC95 **[0.9980829527187469, 0.9999999999999998]**, realised gain **0.463**, zero inadequate.
- README `kernel_rounding.py`: **seven counted checks passed, exit 0**, with full existing source/interpreter/Go/kernel/test-binary closure in `kernel-rounding-rev9.json`.
- Create-once `revision9-checks.json`: **30 checks passed**, covering both verbatim fixes, unchanged **13** reviewed executables, content→SHA-256 closure for all **12** snapshot archives, earlier HD9 histories and earlier rounds/responses unchanged, and byte-identical round-8/ADR round-6 records.
- Benchmark, calibration and projection were **not rerun or changed**; revision-8 resource arithmetic remains authoritative. No historical method-selection rerun.

ADR-014 Records now includes its round 6; current draft trackers name revision 9 and the next joint review. A dated HD9 revision-9 history line was added with state still proposed; earlier history is unchanged. Under Main's standing permission, unmerged R-081 now mirrors the new remeasurement suffix. No new R-ID is needed or consumed; next free remains R-087.

The optional boundary-projection advisory is adopted: README and the submission's supersession list name all three reserve-free revision-7 planning projections; only planning memory is superseded. No reporter changes, commits, model calls, real data or toolchain changes.
