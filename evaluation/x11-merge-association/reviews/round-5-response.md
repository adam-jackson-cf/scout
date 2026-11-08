# X11a revision6b — response to council round5

Date:2026-11-08. **Proposed fixes, not council closure/sign-off. Export preparation and inventory remain unauthorised.** Take alternative A: resolve preparation/statistical defects together before another preparation-only review. No candidate scan, inventory, real map/label/replay or model call occurred. Historical reviews/manifests/extraction v3/v4/timings are preserved.

## Authority and delegated decisions

R-064 standing direction, verbatim:

> can you stop posing councile decisions to me please, go with your recommended each time and just capture them as adr's i can review later - so we can complete our entired implementation plan autonomously

Owner R-073, verbatim: **“(b) High-none only on selected repo”**. Targets800 retained; remaining composition-dependent stops accepted. R-074 records pinned cost-only evidence; R-075 synthetic OC timings; R-076 round5 statistical/preparation recommendations; R-079 synthetic preparation-scale calibration. After fast-forward to main7bb73e52, the canonical main R-064 row is retained and the temporary older source-branch duplicate removed; no main ledger row rewritten.

[ADR-013](../../../docs/adrs/2026-11-08-x11a-delegated-decisions/ADR.md) lists questions/options/choices/rationales/source rounds for all carried/delegated choices. Required omitted choices now explicit:

- **48→384 one-core binding hours:** measured beta minimum254.57h could not fit48h; rev5 two-point maximum100.01669184h/old3×300.05007552h justified a bounded larger band. Rev6 directory estimator is retimed separately; its singleton full maximum860.3267h can exceed384, so no carried-forward3×headroom/universal-feasibility claim.
- **Maximum cumulative look4096→1024:** fixed256/512/1024 sequence with undecidable→not testable, finite workload, no extra favourable looks.
- **.65 utilisation:** explicit pre-inventory engineering allowance for subscription idle/queue/availability, not measured saturation or guaranteed quotas.42days and3000 model-hours remain frozen; new real calibration determines whether the band is usable.

ADR status is **Accepted — agent decision under owner delegation (R-064); owner review pending**. It is not council acceptance. ADR/preregistration must be reviewed and signed in the same future council packet; no sign-off/digest is fabricated. Existing accepted ADRs/design HTML/index/cards are untouched.

## Issue-by-issue changes

Fixture descriptions below identify implemented acceptance cases, not test-pass claims; exercised evidence is separate.

| Issue | Revision6 change | Scope / evidence |
|---|---|---|
| R1-01 carried | **Intentionally remains open/deferred.** Full GUIDE/real orchestration/calibration/replay/label/scorer/stage-seal writer, whole-pipeline plants and inventory-conditioned OC are not claimed complete | Later experiment stages refuse; preparation-only review does not authorise model collection |
| R5-01 blocking | Cold+C0+up to first2000 only; one tree+batch/revision; compact trace; disk-backed per-revision ownership digest index; NDJSON frontend stream; BaseException operational seals;8192MiB heap | Real80k-file×20 synthetic runner, fixed100k-file×2002 per-repository assumptions and separate measured/projection artifacts; no candidate scan. Envelope facts outside scanned segment relabelled unscanned |
| R5-02 major | All scanner errors operational, structural content classes only; max3 identical-recipe attempts; terminal operational infeasibility; immutable relocated Git+dylibs/helpers/templates; OS build/sandbox binding/micro-round; atomic build promotion/recovery; gitlinks empty/count-reported; sorted filesystem-independent aliases | PREP/ASM/INV retain refusal provenance; mandatory operational not testable, candidate preparation-ineligible with reason. Fixtures include real synthetic build recovery, scanner failures, alias/gitlink/mode drift and interruption |
| R5-03 major | Protocol requires reviewed_inventory; OC validates original inventory authority/export artifact manifest/seal/sole output, exact non-null count+directory vectors, inventory/vector digests and pinned recipe;399 only, no timing overrides | `oc_selftest_v6.py` refusal fixtures; active v5-named source files now implement rev6 contract |
| R5-04 major | IUT one-sided CPα=.05/3 each look, no scenario/event gate penalty; simultaneous bounds descriptive only | n256 support≥219/error≤5;512≥429/≤15;1024≥847/≤36. Historical12-study10000-draw comparison explicitly smoke, not calibration |
| R5-05 major | Per-case normalised-directory hashes/shared age keys/top share; paired directory bootstrap plus block bootstrap, min support lower/max best-case upper; independent directory latentρ0/.10 crossed through every scenario |56160 binding/198 sensitivity scenarios; additive convex latent preserves merge stress; exact count-vector coalescing preserves seeded key draws. New timing, fixed384h/8h and honest possible capacity stop. Cameron/Gelbach/Miller cited via R-076, no general coverage theorem |
| R5-06 major | Future180-session synthetic-GUIDE-only calibration (60 each role,12×5), pinned models/medium efforts/renderers; λ=1.25×maximum role mean; measuredQ and actual calibration hours; exact H/day/k gates before maps | Static inventory no longer claims vacuous subscription-hour/wall/session passes; it marks calibration required. Old break-even63.7037037037037s/71.96052938627196s disclosed;4P worst all-escalate+retry retained because mid-replay failure cannot gracefully truncate census |
| R5-07 major |≤6 in-flight cases, sequential requests within slot, prepay6, complete canonical/replacements/adjudication/audit before refill; finishing reserve36×max(λ,930s),≥9.3h | Six repo×age pools24960 each, exact reuse charged once; own-cell eligible hash frontiers only. Start-time20% SHA audit threshold, minimum40/pool top-up from remaining budget/cases only |
| R5-08 major | Remove12-case grouping/k/duplicate rates/duplicate loss and obsolete group stage; **concern duplication unmeasured** | X11b PLAN explicitly retains outstanding REQ-176 grouping/coverage/state-deduplication requirement. No design/default waiver |
| R5-09 minor | Keep simpler abstain-only generator; call .72/.88 infeasible **under that generator**, not physical impossibility; report precision-defect companion adjustment | Limitation explicit in PREREG/spec/ADR; no unnecessary wrong-ID machinery |
| R5-10 minor | Correct final copula term Z_age_arm; require both primary/Jev-only uniform.90 power consistently | PREREG/spec/code agree on conservative implemented rule |
| R5-11 minor | Discriminating count plants include unequal A_known/E_known, none assignments, positive unassigned unknowns and A_unknown≠U_all | Frozen literal counts/four estimator checks, not merely all-assigned supportive plants; full-driver validation still R1-01 |
| R5-12 minor | No-replace/no-lazy-fetch environment, reject graft/shallow/replace/promisor; PGO off both recipes; reviewed stdlib-first bootstrap rejects cache before helper import; whole immutable tool/platform preflight | Sentinel fixtures; differential test uses verified bootstrap helper. Five full ordinary canonical clones actually restored/verified, not flag clearing; `tool-pins-v6/object-restoration-v6.json` records commands/URLs/HEAD/pins/full connectivity |
| R5-13 minor | Exact per-operation trace consistency, tree/batch/OID digest/count/revision/argv validation; codegraph exclusions counted | Fixtures use genuine execute-produced synthetic build/exposure/content/terminal-operational seals through ASM→artifact bindings→inventory, not hand-built success envelopes; empty/tampered traces refuse |
| R5-14 minor | R-064/R-073–076/R-079 ledger citations before freeze; comprehensive ADR-013 with omitted compute/look/utilisation decisions | No historical ledger rows rewritten; no false pending-ID/source claim; owner and council review remain distinct |

## Calibration arithmetic, not measured model evidence

No calibration request has run. These are **hypothetical arithmetic counterexamples** using P_all10000, full149760 label reserve, all role means equal and2600 projected capacity/cell. Exact fractions avoid hiding gate boundaries:

| Hypothetical role mean s | HypotheticalQ sessions/h | λs | H model h | d days | projectedk | Result |
|---:|---:|---:|---:|---:|---:|---|
|48|360|60|128414/45|313645/8424|2600|passes resource projection only|
|52|360|65|137783/45|332365/8424|2540|fails3000h before maps|
|48|290|60|128414/45|10668185/244296|2468|fails42days before maps|

These demonstrate non-vacuous measured gates, **not observed throughput/quality or a power claim**. All six pools must also satisfy actual projected age×arm1404/2106 floors.149760 label/audit+180 calibration=149940,60 spare; no grouping budget or pool transfer.

## Resource/evidence interpretation

Final OC r3 profiles each120 synthetic studies: balanced2.103530166s, concentrated2.10516475s, age-disjoint singleton5200-key6.462790666s; RSS16416768/19070976/19578880bytes. Corresponding256/1024 full-grid projections are70.0055/280.0219h,70.0599/280.2395h,215.0817/860.3267h (display approximations; exact JSON governs).384h remains a fail-closed cap, no scenario/directory filtering/reselection. Sensitivities may stop at8h with remainder unmeasured. Timings do not establish OC support/false-support probabilities; actual inventory-conditioned OC remains unmeasured.

Git restoration evidence SHA-256: `95854874cfde17d52a806e442fb7ea3597c75732d1e2beb073ef808e49e2094c`. It records fresh ordinary clones at `x11-candidates-v6`, exact original anchor SHAs/canonical origins, current HEAD, no promisor/shallow/replacement/graft state, full/observed-envelope object connectivity and fsck. No source/diff/supply analysis occurred; old partial clones are untouched.

### Preserved first scale measurement

[`prep-scale-v6.json`](../prep-scale-v6.json) / [projection](../prep-scale-projection-v6.json):80,000 TypeScript source files,80directories,3,572,400bytes/80,000lines per revision,20 fresh materialisations/scanner states, identical tree. Real init/map completed, but **zero owned paths in all20**: this is materialisation/offline traversal evidence, **not populated-ownership output-scale validation**. Elapsed2325.934470999986s includes setup/cleanup; processing1773.5792133308132s; maximum Darwin child/self RSS160808960bytes (not simultaneous aggregate RSS), heap8192MiB, Nodev22.18.0/Clawpatch0.7.3/Git2.56.0.

Preserved assumptions fixed100,000files×2,002revisions for each candidate. Conservative whole-elapsed125.125× extension gives80.84237518996478h/repository (three attempts873097.6520516197s); processing-only61.64419418556056h also disclosed. The original precompletion processing-only assumptions file is unchanged; using whole elapsed avoids favourable omission of cleanup. Neither estimate guarantees arbitrary language/byte/structural complexity. The empty-output receipt is not substituted for populated output evidence.

### Populated scale measurement

[Positive construct](../prep-scale-construct-v6b.json) first established three real features/seven owned paths. Frozen [v6b runner](../prep_scale_v6b.py), followed by [completed receipt](../prep-scale-v6b.json): **80,001 files×20 revisions,81 real features and80,001 owned paths every revision**;80packages×1000Go sourcefiles plus go.mod.160,000source lines,4,471,200source bytes/4,471,247total bytes per revision; identical trees, fresh materialisation/scanner state. Elapsed2455.8070487079676s, processing2311.1212765829987s, peak RSS188481536bytes, heap8192MiB. Disk ownership records59,208,700bytes/index3,621bytes; NDJSON59,201,755bytes read back20records/1,600,020ownedpaths one record at a time. This exercises actual scanner ownership, indexed persistence and Python stream emission/readback, **not frontend binary execution**.

[Derived projection](../prep-scale-projection-v6b.json): fixed100,000files×2,002revisions for every candidate; multiplier125.12343595705053. Whole elapsed including setup/cleanup gives307279.0159818847s/repository, three attempts921837.047945654s; processing-only289175.4350395097s separately. No candidate measurement, arbitrary source-complexity/changed-content guarantee or simultaneous aggregate-RSS claim. Original zero-owned receipt/source retained; v6b supersedes it for populated-output capacity.

### Observed metadata prerequisite and integration boundary

Requested `fetch`→`stash`→`merge --ff-only origin/main`→`stash pop` advanced this worktree04d5e632→7bb73e52. Ledger conflict resolved preserving all main rows plus R-073–076/R-079 and one canonical R-064; ledger staged solely to mark resolution. Git retained the conflict stash; no new commit.

Trusted Go metadata extraction initially refused **Python whole-install closure drift**:15new ambient stdlib caches, no changed/removed reviewed files or links. Root-cause fix: exact previously reviewed2505files/nine relative links copied to read-only content-addressed Python3.14.10 root; no ambient file deleted or newly generated cache repinned. Old v6 metadata preserved; active [Python/runtime v6b](../tool-pins-v6/python-runtime-v6b.json) and [platform/tools v6b](../tool-pins-v6/persistent-locations-v6b.json) bind the immutable installation and modes. Metadata extraction then succeeded through verified bootstrap/helper/Python/platform preflight; refreshed frontend dependency metadata accounts for main's changed `internal/clawpatch/client.go` and `internal/state/issue_match.go`.

Main's independent rev6 integrated checks failed (22 preparation tests:five errors;27 integrated:two failures/one error;frontend missing-SHA error context). Revision6b fixes the root causes without weakening assertions: canonicalise temporary roots so `/var` and `/private/var` identify one pinned file/mock path; explicitly bind each attempt's preflight input/authority even if no command runs; wrap streamed record decode errors with expected exact merge SHA while preserving the cause. Added missing/tampered-preflight rejection plants. A first authorised rerun exposed the genuine pre-command-refusal trace defect; after its producer/verifier fix, **22 preparation tests and27 integrated tests pass**, and pinned PGO-off frontend Go tests pass. [Verbatim commands/environments/results and all failures](rev6b-checks.json). No full experiment or council-pass claim.

## Review boundary

Request another **export-preparation-only** council round on `rev6b-digests.json` and ADR-013 in the same packet. Not an inventory grant, full experiment sign-off or default/design change. Any changed signed byte needs renewed scoped review. R1-01 remains the explicit later implementation prerequisite; real outcomes never fill validation gaps or tune choices. No commit.
