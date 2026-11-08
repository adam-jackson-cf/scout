# Round 6 response — revision 7

Date: 2026-11-08. **DRAFT; not signed off. No candidate preparation/inventory/model authority.** This response continues the interrupted writer's work rather than replacing it. The packet is `rev7-digests.json` (self excluded); exact exercised commands/output are in [rev7-checks.json](rev7-checks.json). Only council can close issues. R1-01 remains explicitly deferred; R6-05 remains **open**, with an explicitly changed projection rule submitted for review.

## Interrupted work and provenance

Initial `git status --short` showed uncommitted X11a/X11b and ADR/ledger/design work. Directory mtimes and existing v7 source/recipe/timing/construct artifacts showed substantial partial revision7 work; there was no round-6-response draft or rev7 manifest. Those artifacts and all historical reviews/manifests/extractions were retained. No surviving benchmark process existed; `x11-tools/benchmark-v7-files` had an empty ownership seal and partial revision records, so it was not treated as completed evidence. A new create-once full file benchmark used `benchmark-v7-files-rerun`; byte-heavy used `benchmark-v7-bytes`.

The first attempted restart refused before benchmark execution: exit1,1.94s, “reviewed entrypoint source drift”. The interrupted source-pin packet was stale. All current executable sources and v7 build recipes were pinned before the successful runs and final integrated check; no trust check was bypassed. The final bootstrap/input pins are recorded in rev7-checks. Codegraph was unavailable (`command not found`); ordinary targeted reads were used, never candidate scans.

The already-created [envelope stdout](../envelope-v7.json) matches Main's source receipt byte-for-byte, SHA256 **f8ebcdcb8aad65d1b288433a1180ae12a36fa7ff3cfb2b088c526a7dcb89d5ca**. The generator was **not rerun**. Its SHA256 is00b6bf253f5ea2547854f228526fff34535084e008874725540769b9600bd1f5; Main's17.9s/exit0/empty-stderr execution is preserved in [run record](../envelope-v7-run.json). The original input bytes used in that run are retained as `envelope-v7-inputs.json`, SHA256e492da740af2aa778304394dd326c269099e7e7aa30a9207e680c92b0f34c74b. Historical argv still names the original active input path; it is not relabelled as a run against today's inputs.

ADR-013 `round-1.md` already preserves study round6 byte-identically. It was checked and left unchanged, as instructed; no reverse-chronology renumbering or invented review record was added.

## Issue-by-issue disposition

| Issue | Revision7 response | Evidence / remaining scope |
|---|---|---|
| R1-01 carried | **Open, deferred as planned.** Full GUIDE, real orchestration/calibration/replay/labels/scorer/stage seals, whole-pipeline plants and actual inventory-conditioned OC remain absent | PREREG§§2/9/11 disclose prerequisites; later real stages refuse. Preparation-only review is not overall sign-off |
| R6-01 blocking | L_r=min(2000,post-C0 envelope), floor1500, unchanged repositories/anchors; parameterised inventory/vector dimensions/OC streams/circular bootstrap/plants/replay bounds | Git receipt reproduces2791/1666/316/70/3248 envelopes; capped2000/1666/316/70/2000. Prometheus alone passes Go length, still subject to every other gate and E1b-v2 disclosure. No replacement/C0 movement. OC tests exercise1500/1666/2000 and reject1499/age mismatch; changed OC was retimed |
| R6-02 major | Reviewed Git-only producer and verbatim once-only receipt/argv/environment; correct all “object-only/no supply facts” claims | PREREG§3, recipes, README, active inputs and ADR openly disclose earlier unreviewed capped eligibility/object counts and supply-informed amendment. Fresh-clone Prometheus default HEAD is separate from original resolved-tip pin, unused; anchors unchanged |
| R6-03 major | Enforce3600s subprocess process-group kill,14day attempt, canonical create-once shared42day budget,≤five locked concurrent operations | Ordinary operational failures get≤three fresh attempts; interrupt seals `PreparationInterruptedOperational`, whole deadline `PreparationBudgetExhausted`, both global stop rather than selective exclusion. Host interruption only for documented failure under Deviations. Integrated timeout/interrupt/deadline/reset/sixth-slot fixtures pass |
| R6-04 major | Delete full149760-label H/d vetoes; keep fixed-part feasibility and uniform pool/k frontier with1404/2106 floors | PREREG§10/ADR; corrected arithmetic below. Full-reserve H/d descriptive only, no transfer/new allocation machinery |
| R6-05 minor | **OPEN.** Genuine final-source file-heavy and≥2GB×3 production-path benchmarks completed; both original mandated max-rule projections fail and are preserved unchanged | Fail-closed gate retained. Main's delegated additive-model replacement and once-only actual Git-size receipt are proposed, not silently substituted or council-approved; detailed record below |
| R6-06 minor | Content refusal names structured in-scope revision/failed_step; requires successful matching tree and applicable batch trace plus scoped-chain binding | Forged preflight-only/no-batch/wrong-revision/wrong-step content-seal plants refuse. Genuine content/operational seals still pass through verifier/assembler/inventory fixtures |
| R6-07 minor | “Robust to either dimension singly”, not conservative against two-way dependence; add simultaneous dependence,ρ_dir>.10,few-directory threats | Inventory/OC report K,1/Σshare² and largest share by repository×age; R-083 caveat, no general multiway coverage theorem |
| R6-08 minor | Specify minimum256-replicate full-grid compute projection; `stop_reasons` preserves first gate failure and independent compute stop | Contract plant checks retained first status, separate reason and deduplication; actual exhaustion still fail-closed |
| R6-09 minor | Exact primary verdict/high-gold-none→HD8 mapping, Main records under R-064 | Supported+realised→supported scoped draws; supported+unrealised→inconclusive/accuracy supported, high-none clause untested; not-supported→not-supported; otherwise inconclusive. HD8 remains proposed until actual sign-off, then preregistered/tested by X11a |
| R6-10 minor | Extend R-075/R-079 with exact versioned receipts/numbers; R-073 session link; R-082 Berger1982/R-083 Cameron–Gelbach–Miller2011 through council; R-084 provenance/choices | Informed IDs are HD8/X11a/ADR-013, not issue IDs. STATUS points to next unallocatedR-086, coordinating concurrent allocations. No extra ID requested |
| R6-11 minor | ADR-011 accepted-form status/metadata, effective only on folder's final explicit zero-issue round; review section names that round, not superseded manifest | Pending-owner state only in editable index/STATUS trackers. ADR-013 summary/index/AppendixB card and ledgers/STATUS are in signed submission set now, no unsigned landing edit. Existing ADR round1 unchanged; future council rounds continue chronology. No current acceptance asserted |

## Corrected calibration table (supersedes round-5-response table, not its historical bytes)

Hypothetical arithmetic only: P_all10000, projected capacity2600/cell, equal role means. No calibration/model session ran. R=345.944444…h, J400h, F9.3h; fixed elapsed=89005/8424days. Exact fractions govern.

| Mean seconds | Q/h | λs | Fixed model hours | A_H | A_D | a | Uniform pool | k | New resource result |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
|48|360|60|16094/45|158541|176535|149760|24960|2600|fixed parts/floors pass|
|52|360|65|16103/45|146334|162955|146334|24389|2540|fixed parts/floors pass; full-reserve hours are not a veto|
|48|290|60|16094/45|158541|142209|142209|23701|2468|fixed parts/floors pass; full-reserve days are not a veto|

Old full-reserve diagnostics remain128414/45h and313645/8424days;137783/45h and332365/8424days;128414/45h and10668185/244296days. They do not decide the new gates. Real measured λ/Q and actual proxy capacities still can fail before maps.

## R6-05: measured benchmarks, failed original rule, proposed additive replacement

Both new receipts measure final PREP SHA256 **d52135b07b13100fb0f0092de759c9bd17b06821d5c92eb5a4fa1d0e3c8e0e8d** through genuine execute→verify_preparation_seal→built frontend NDJSON consumption, with5% changed sources per revision. No execute/authorise/verifier/frontend mock; location-only synthetic runtime metadata is explicitly derived/pinned in each receipt.

| Receipt | Shape | Whole elapsed seconds | Original max-rule seconds/attempt | Three attempts | Fits unchanged14/42days? |
|---|---|---:|---:|---:|---|
|[file-heavy](../prep-scale-files-v7.json)|80,001files×20,4,440,047B recorded tree size;80,001ownedpaths every revision;1,600,020streamedpaths,19frontend merges|2803.829540457984|126423588.3087923|379270764.9263769|no/no|
|[byte-heavy](../prep-scale-bytes-v7.json)|2,001files,2,000,120,047B/revision×3|40.76915474998532|1359650.9713388407|4078952.914016522|no/no|

Both outer commands exited0 (2805.29s/42.73s wall); receipts are unchanged. File-heavy max RSS699711488B (max child/self, not simultaneous process-group RSS). The file-heavy projection is about **1,463 days**, not candidate runtime evidence: max(file ratio,byte ratio) multiplies a file-driven80k-small-file workload by a roughly450× byte ratio. Neither that extrapolation nor its failure proves candidate preparation infeasible.

**Main's delegated direction under R-064**, preserved here: replace the max rule with additive time/revision=a·files+b·bytes+c; retain both receipts; use actual pinned candidate Git file/byte metadata, no source scan; add a conservative1.5×margin; keep14/42day ceilings and stop “preparation not feasible” if they fail. Main required stdout-only read-only collector, pinned Git in `env -i`, ran it once personally, and required reporting both unmarginated/marginated projections and no fitted error estimate. This is an explicit **change to R6-05's rule requiring council review**, not permission to tune ceilings or claim closure.

The successful same-PREP [41-file/2247B construct](../prep-scale-construct-v7-attempt2.json),13.584968417009804s/three revisions, supplies the third independent point: two observations cannot identify three unknowns. Fit whole elapsed/revision, so setup/build/verifier/frontend costs are charged again on every projected revision. Coefficients are a=.001696478572992594s/file,b=2.867979796619823e−9s/byte,c=4.458760739826635s/revision; exact fractions and all observations are in the receipt. **Three design points/three coefficients leave zero residual degrees of freedom: no error estimate or uncertainty interval. The1.5× multiplier is the only safety factor, not a confidence bound.** Linear scaling/complexity/host variability remain threats. The synthetic runner's `bytes_per_revision_min` field records its final generated tree total; two-digit file-fixture revisions add small source-text bytes, so this field is not independently audited as the true minimum. Receipts are retained rather than silently relabelled/refit.

[Main's verbatim size/projection stdout](../preparation-capacity-v7.json), SHA256 **141c6764dee809dd0bb198d53e1a414c2f744b1403d9d99c1e282af530c8e037**, and [execution metadata](../preparation-capacity-v7-run.json) record one run of script SHA256 **1599b49bbb41b48454bc8394fdb1dc6398d760f1957a69250f0c1fe9b561453b**. Exit0,335.67s reported wall; `time -p` real335.56/user275.78/sys53.92. Stderr contains only those34 timing bytes, SHA2568a8ec9f3b87a4390f475b02944edc004d650c569539d32ec1adee35afc9d1a5a; no collector diagnostic. Command uses `env -i LANG=C`, pinned Python-I-B, preserved input e492da74…; Git subprocess environment is an explicit empty allowlist. Only rev-list and ls-tree-r-l-z; no blob contents, diffs, source scans or inventory ranking. Counts include materialised gitlink entries in the file term and exclude .codegraph metadata entries, never reading/modifying .codegraph directories.

Actual scoped revisions are Cold+C0+firstL_r, not an arbitrary worst-case source envelope:

| Repository | Revisions | Files range | Bytes range | Attempt seconds, margin1 | Attempt seconds, margin1.5 | Three attempts seconds, margin1.5 | Largest revision seconds, margin1.5 |
|---|---:|---:|---:|---:|---:|---:|---:|
| django/django | 2002 | 6661–7014 | 41755002–45930981 | 32571.99473475014 | 48857.99210212521 | 146573.97630637564 | 24.73370460889651 |
| microsoft/TypeScript | 1668 | 65704–81369 | 208413944–380592869 | 210949.14060242526 | 316423.7109036379 | 949271.1327109138 | 215.37676751802115 |
| gin-gonic/gin | 318 | 117–131 | 715471–916617 | 1485.4622025337892 | 2228.1933038006837 | 6684.579911402051 | 7.025442407888855 |
| spf13/cobra | 72 | 66–68 | 547468–705271 | 329.28490013679317 | 493.9273502051898 | 1481.7820506155692 | 6.864154899288182 |
| prometheus/prometheus | 2002 | 1113–1534 | 23091052–29361867 | 13716.172714107139 | 20574.259071160708 | 61722.77721348212 | 10.704013554812976 |

Under at mostfive concurrent repository operations, largest three-attempt projection949271.1327109138s; all candidates fit14days/attempt,42days/three attempts and3600s largest whole revision as a conservative per-command proxy. Status is **projected-feasible-not-authorised**, not observed candidate runtime feasibility. If this proposed gate fails on its reviewed receipt, X11a stops “preparation not feasible”, never higher ceilings/refit/reselection. R6-05 stays open until the council accepts the amended model and evidence.

## Delegated decisions

All are recorded in ADR-013; R-064 never waives council review:

1. Adopt council alternative A: pre-freeze1500–2000 variable lengths with candid eligibility disclosure, no repository/C0 substitution. Sole viable Go length candidate remains Prometheus, E1b-v2 exposure visible.
2. Adopt recommended3600s/14day/42day and interrupt/global-stop policy, bounded concurrency and concrete content-refusal trace evidence.
3. Adopt recommended fixed-part+uniform-k resource gates, not full-L affordability vetoes; retain labels' hash-prefix stopping and unchanged ceilings.
4. Keep single-dimension-robust wording, explicit HD8 map and exact minimum256 compute projection/stop reasons.
5. Follow adopted ADR-011 form and signed index/card set; preserve existing round6→ADR round1 bytes/chronology.
6. Following Main's explicit R-064 direction after the measured max-rule failure, propose additive file/byte/fixed-cost model with actual Git-only metadata, margin1 and1.5, no fitted error estimate and fail-closed unchanged-ceiling gate. **Council must review this rule change; R6-05 remains open.**

## Exercised checks and review request

- Integrated pinned isolated selftest **exit0**,208.45333741704235s:29outer tests OK (including imported bootstrap/OC contracts and additive-fit/negative/singular/refusal plants); its nested preparation suite26tests OK,163.299s. No separate duplicate preparation/OC suite run.
- Pinned Go1.26.10 frontend `go test -v -pgo=off ./evaluation/x11-merge-association/frontend`, recipe environment plus pinned Git on PATH: **exit0**,5top-level tests pass,4.189201375003904s command wall; stderr empty. `-v` exposes exercised count without another run.
- Variable-L OC retiming already present from interrupted writer: six120-study profiles, mixed1500/1666/2000 plus all2000; metadata/source/closure pins retained. It is synthetic compute evidence, not actual OC power/false-support validation.
- No model calls, candidate ownership scans/diffs/inventory ranking/maps/labels/replay. Main's once-only envelope and later once-only size distributions are explicitly recorded pre-freeze checks, not predictor–outcome associations.

Request another export-preparation-only council round, including ADR-013 and its signed index/AppendixB/ledger/STATUS set. Ten R6 fixes are implemented for review; **R6-05 remains open pending explicit acceptance of the changed projection rule**; R1-01 remains deferred. No experiment/ADR sign-off is invented. HD8 stays proposed. Ledger allocation remains R-073–R-076,R-079,R-082–R-084; **no additional ID request**. No commit.
