# Round 7 response — revision 8

Date: 2026-11-08. **DRAFT; export preparation/inventory remain unauthorised.** This submission implements R7-01/R7-02/R7-03 for council review; it does not declare them closed or invent an authorising verdict. R1-01 remains explicitly open/deferred. Manifest: `rev8-digests.json`, self excluded. All earlier receipts/reviews/manifests remain preserved. Main saved study round7 and ADR-013 round2 with identical bytes; the council ratifies ADR roundN=study roundN+5. ADR round1 is unchanged.

## Disposition

| Issue | Revision8 implementation and evidence |
|---|---|
| R1-01 carried | Unchanged deferral: full GUIDE/real orchestration/calibration/replay/labels/scorer/stage seals/whole-pipeline plants/actual conditional OC still required in the implementation round. No map/label/model authority |
| R7-01 major | Follow council's recommended **single held-out TypeScript direct bound**, not a language-family refit. Three Go design points rerun on final PREP; prediction frozen before Main's once-only held-out run. Reuse v7 size facts, no candidate access/new collector. Held-out comparison passes all unchanged caps; no refit, rerun, fixture change or ceiling change. Detailed clocks/ownership below |
| R7-02 minor | Canonical PREREG/REC/ADR-013 table/card/R-079 state final rule **and measured result**, not a proposed/pending capacity gate; they become effective unedited only on the authorising zero-issue round. Draft headers/README/STATUS still truthfully withhold authority. Historical failed max-rule evidence and R-064 direction stay in history/ledger/response. R-079 includes final coefficients, frozen prediction, held-out receipt and direct-bound result |
| R7-03 minor | Follow council's smaller **canonical paths** remedy, no extra ledger. `canonical_paths(inputs,operation,repository=None)` derives fixed create-once artifacts/preparation-v8/<operation>[--owner__repo].{output,seal}.json. Execute rejects alternate paths, repeat invocation/any existing seal, and prior canonical interrupt/budget stop. Empty crash reservation also blocks repeat. ASM discovers every configured ownership seal plus frontend/exposures (seven for real inputs); `--seal` list and old API removed. Interrupt/terminal refusal cannot be hidden by omission/fresh filenames. Host-failure continuation requires council-reviewed deviation. Fixtures exercise alternates, repeat terminal invocation and global-stop omission/continuation |

Canonical budget stays `artifacts/preparation-budget-v7.json`, preserving reviewed schema,42day clock and≤five locked slots;≤three internal attempts,14day attempt/3600s subprocess group-kill remain unchanged. Canonical paths are derived from that budget directory, not an operator-selected new root. No candidate clone was touched by this worker or its implementation slices.

## Final-PREP Go reruns

Final PREP SHA256 **77446e97386561cf62f5afb8cfea96f5809f96445b8d724e223d2a57cb3eca44**. All three v8 receipts use genuine execute→verify_preparation_seal→built frontend, unchanged Go shape,5%churn for populated points; no mocks. They retain exact original/snapshot source/authority/runtime derivations.

| Receipt | Shape | Whole elapsed seconds | Outer exit |
|---|---|---:|---:|
|[construct](../prep-scale-construct-v8.json)|41files/2247B×3|13.361451208998915|0|
|[file-heavy](../prep-scale-files-v8.json)|80001files×20|2465.5741741670063|0|
|[byte-heavy](../prep-scale-bytes-v8.json)|2001files/2000120047B×3|39.49403787497431|0|

All v7 receipts and failed original max-ratio projections remain unchanged; their old≈1463day figure is a projection-method artefact, not measured candidate capacity. V8 preserves its diagnostic max-rule calculations too; the adopted gate is additive plus held-out direct bound.

Fit whole elapsed/revision with **all three final-PREP points**, no dropping/refit. Exact fractions are in [capacity-v8](../preparation-capacity-v8.json); decimal coefficients: a=.001485893272630103s/file,b=2.899084947714978e−9s/byte,c=4.392888931244594s/revision. Three points/three coefficients give **no residual degrees of freedom, error estimate or uncertainty interval**. Fixed1.5× is the only safety factor, not a confidence bound. Go/Django use sums over Main's pinned actual scoped sizes; Django's unexercised Python mapper remains a threat.

## Frozen held-out TypeScript point — no post-run tuning

Council R7-01's fixture is informed only by already-recorded TypeScript maxima. Actual fixture: **81402files,395061604B per revision**,81400TSsource files, exactly40700of those tests (20350tests/ tree and20350sibling*.test.ts),2JSONconfigs. Nested≥three levels, TS imports/exports/functions;4code files≥1050000B, normal code files≥4800B, no padding comments; root package.json has test script and tsconfig.json. Exactly3revisions and4070changed sources/revision after base (5%). No candidate file content/diff/scanner/metadata collection informed it beyond the already-authorised size distributions.

**Pre-run freeze:** [prediction](../prep-scale-heldout-ts-prediction-v8.json) SHA256 **4b993052ef109fdcf0c53dce2f991392c0f6e163fa5091d8c2394be30406ccfe**,126.49289025945677s/revision. The freeze command exited0 before scanner work; it pins actual ls-tree sizes/tree digests, generated working-file manifest, source snapshots/context, exact Go observations/coefficients and prediction. Runner SHA256 **721f91473786519326ae916e8f3922d0306499ffec71cc8e9e6f4fe44f911eba**. [Exact Main run request](rev8-heldout-request.json) records argv/cwd/digests. The run validates frozen fixture/source/input/prediction and creates a one-shot invocation marker; no rerun path.

**Main ran once**, output absent beforehand, exit0,365.607s wrapper wall, stderr empty (e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855). [Held-out receipt](../prep-scale-heldout-ts-v8.json) SHA256 **c47075f934c1beaaa5fe7038aaea5eb4124413f3c9ec9470d0d6df8387ffe0c9**; [run/clock record](../prep-scale-heldout-ts-v8-run.json). Main's stdout SHA2567f122afcb66657eb94d05072374229418ed8c81ac077eb3e83dad4e6b0ad1e65. Peak RSS725483520B (max child/self, not simultaneous process-group total); frontend0.5509349170024507s,2merge outputs. This worker generated/froze inputs but did **not** run the TS scanner.

### Why419.987s exceeds Main's365.607s wall

The receipt's `elapsed_seconds` is **charged compute across the earlier freeze/setup and later run**, not one continuous process wall span. This definition was in the frozen runner before Main's run: `run_frozen` passes earlier setup+freeze+current validation time to execute_stage; execute_stage offsets its monotonic accounting start by that charge. Operator waiting between freeze and run is excluded.

| Component | Seconds |
|---|---:|
|Earlier synthetic generation/setup|46.08128170896089|
|Earlier fixture freeze/manifest accounting|9.647321375028696|
|Prior-phase charge, sum|55.72860308398958|
|Receipt charged elapsed|419.9873540849658|
|Receipt minus prior-phase charge (later runner/validation/operations)|364.2587510009762|
|Main wrapper continuous wall, run invocation only|365.607|
|Wrapper minus internal later phase|1.3482489990238378|

The residual is outside the internal clock; launch/bootstrap attribution is **[INFERENCE]**, not separately profiled. `started_utc` is set after run_frozen validation, so its span also is not the receipt's charged-elapsed clock. The comparison uses **all419.9873540849658s/3** as frozen, not a favourable post-run subtraction. This conservatively includes earlier setup/freeze and excludes only human waiting.

### Zero owned tests matches the pre-frozen construct

Before any run, reading pinned mapper-source behaviour showed Node source groups exclude test paths (`third_party/clawpatch/src/mappers/node.ts`, isReviewableNodeSourceFile) and own group.files, while tests are association references. Its packageTestFiles also caps association candidates at200 before per-group filtering. Therefore the frozen metadata explicitly requires **genuine nonempty source ownership, not ownership of every test**, and records that40,700test files do not imply40,700association-filter candidates per group. The Go all-source ownership assertion is unchanged.

Observed each revision: **40702total owned paths,40700planted non-test source paths,0planted test paths**. Zero test ownership is consistent with that pre-frozen construct; the200cap limits references, not an expectation that200tests become owned files. We do not claim all40700tests entered every inner filter. TypeScript language/package detection and production execution exercise the Node path; correspondence of pinned built scanner to inspected v0.7.3 source retains the council's stated source-version **[INFERENCE]** caveat. No expectation was relaxed after seeing results.

## Frozen-rule comparison

S_obs=419.9873540849658/3=**139.9957846949886s/revision**; frozen prediction126.49289025945677; ratio**1.1067482481255293**. Apply S_bound=max(S_obs,prediction), not a new fit. TypeScript1.5×S_bound=209.9936770424829s;1.5×1668×S_bound=350269.45330686145s/attempt,three attempts1050808.3599205844s. All within unchanged3600s/14days/42days. Other repositories retain additive sums; no new candidate facts.

| Repository | Attempt seconds, margin1 | Attempt seconds, margin1.5 | Three attempts seconds, margin1.5 | Largest revision seconds, margin1.5 |
|---|---:|---:|---:|---:|
| django/django | 29539.608706423667 | 44309.4130596355 | 132928.23917890652 | 22.421464597394674 |
| microsoft/TypeScript | 233512.968871241 | 350269.45330686145 | 1050808.3599205844 | 209.9936770424829 |
| gin-gonic/gin | 1456.2277103544461 | 2184.341565531669 | 6553.024696595007 | 6.8852974507599845 |
| spf13/cobra | 323.5364174089855 | 485.3046261134783 | 1455.9138783404348 | 6.743899729673268 |
| prometheus/prometheus | 13009.248577783846 | 19513.87286667577 | 58541.618600027316 | 10.12186676565733 |

Final status **projected-feasible-not-authorised**. A failed branch would be **“preparation not feasible”**, not rerun/refit/fixture/repository/ceiling change. Plants exercise attempt, three-attempt and subprocess booleans independently, linked direct-bound boundary and aggregate failure. Passing planning gates does not grant authority or guarantee runtime/power. No fitted uncertainty is invented.

## Exercised checks and unit-fixture correction

[Complete commands/output](rev8-checks.json):

- Initial integrated run **exit1**,162.4913732080022s: one direct UNIT fixture (`test_revision_scope_excludes_unobserved_content_unit_only`) failed FileNotFoundError because it calls ownership directly, bypassing execute's canonical parent-directory reservation. This was a fixture setup defect, not a scanner/capacity or production-PREP failure.
- Fix: create that fixture's canonical output parent before its direct call. **Only prep_selftest.py changed**, no PREP/runner/model/prediction/fixture change; exact scope/sequence/assertions remain. Refresh its source pin. Preserve original06676fb3…input bytes as [heldout inputs](../prep-scale-heldout-ts-inputs-v8.json), so Main's frozen packet remains auditable. Benchmark snapshots retain pre-fix test bytes; final PREP digest is unchanged, no benchmark rerun.
- Final isolated pinned integrated run **exit0**,169.39888904202962s: **30outer tests**, nested **28preparation** tests (133.185s) and **6capacity** tests (.289s) OK, including bootstrap/OC contracts, canonical repeat/alternate/global-stop and all gate plants. Failed output is retained, not overwritten.
- Pinned Go1.26.10 `go test -v -pgo=off ./evaluation/x11-merge-association/frontend` with recipe environment+pinned Git: **exit0**,5top-level tests pass,3.541415582993068s wall,stderr empty.

## Delegated decisions and hand-off

Under R-064, recorded in ADR-013/R-079/R-084:

1. Take council's held-out direct TypeScript bound, rather than separate-family refit; keep all earlier receipts, all three final-PREP Go points, frozen prediction and once-only Main result; unchanged ceilings.
2. Use council's canonical operation paths/all-seal assembly remedy rather than an additional attempt ledger. Global stops cannot be omitted; host continuation requires review.
3. Use route-(a) canonical effective-rule/result text. Pending authority remains in draft headers/trackers, not in the capacity rule that will freeze.

Request export-preparation-only round8 plus ADR-013 round3. R1-01 stays deferred, HD8 stays proposed, no candidate preparation/inventory/model execution or product/default adoption is authorised. **No extra ledger IDs requested**; R-079/R-084 updated within their draft allocations; next freeR-087 remains for Main to allocate. No commits.

## Main integration before round8

Main requested integration before review so round8 signs final bytes. Fast-forwarded the uncommitted X11 branch from7bb73e52 to **584da4a2**, incorporating ADR-012/PR19 (c420d6f0) and X1/ADR-014/PR18. Used fetch→stash→ff-only merge→stash pop; reported five shared-file conflicts before resolution. Kept the stash for recovery, resolved additively, and marked those conflicts resolved; no reset/clean/restore or commit.

`git diff origin/main -- docs/STATUS.md docs/adrs/README.md docs/hypothesis-ledger.md docs/research-ledger.md docs/scout-design.html` shows **33insertions/0deletions**, all X11 additions. Main's lines, ADR-012/ADR-014 cards/index entries and signed texts are unchanged. X11 adds ADR-013, its card, HD8 history/tracking and R-073–R-076/R-079/R-082–R-084. Preserve main's HD8 register row; its proposed state and X11a conditional mapping are explained by the added X11 history/tracker. All86 research IDs are unique and ascending; **R-087 remains next free**, ADR-013 belongs to X11 and ADR-014 to X1.

Compared all20 `owned_extracts_sha256` selectors from main's X1 `rev9-digests.json`: **19match exactly**. The sole difference is main's adopted HD9 register row: current `preregistered` row SHA256 **cd3f7d8a48dfe01b6c75909ffc2b674cbb3fec77ca6c9a8498007c72eb70612b**, signed pre-adoption `proposed` row SHA256 **e937be022de233692782a8a5dae7546c76d1ab122bf65368339ad25d326d89bd**. Changing only that state in memory reproduces the signed hash. Main confirmed this is the authorised post-sign-off change under X1 round9 §13 condition2(a): HD9 state/date→preregistered plus one history line, landed at584da4a2. **Keep main's bytes**, not a reversion to the signed draft state; no X1-owned extract was edited.

Repeated the same pinned commands on integrated bytes: integrated **exit0**,168.17355641600443s wall;28preparation tests132.196s,6capacity tests.295s,30outer tests166.920s, all OK. Frontend Go **exit0**,5top-level tests pass,2.705706958018709s wall,stderr empty. [Exact commands, outputs, extract hashes and ID checks](rev8-integration-checks.json). No production script/input or synthetic benchmark receipt changed, and no benchmark rerun occurred. The pre-integration manifest (9eb55af605304d7ec0f321dd3db4c8c4af491a289dd3286cdd1e55a31306429f) is preserved in `rev8-pre-integration-digests.json`; active `rev8-digests.json` is regenerated over integrated bytes.
