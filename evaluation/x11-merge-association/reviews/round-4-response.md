# Round 4 response — revision 5

2026-11-08. **Draft submitted for another export-preparation-only review. Not signed off; preparation and inventory remain unauthorised.** This replaces the earlier owner-choice-gate-only response. R4 fixes below are drafter claims for council re-review, not council closure. R1-01 remains explicitly open/deferred.

## Owner decisions and source record

Owner R4-02 choice, verbatim: **“(b) High-none only on selected repo”**.

Adopt council alternative D: mandatory repositories bind at none shares {.2,.4}; selected high-none repository at {.2,.4,.6}; mandatory .6 is sensitivity. Keep 800 known assignments and 800 known existing cases per age. The owner accepts remaining composition-dependent pre-map stops. No candidate supply or composition was measured to choose this target.

Standing owner direction, verbatim:

> can you stop posing councile decisions to me please, go with your recommended each time and just capture them as adr's i can review later - so we can complete our entired implementation plan autonomously

**Research-ledger handoff for Main (no invented R-ID):**

| Source title | Link/provenance | Type | Date | Informs | Claim used | Limitation |
| --- | --- | --- | --- | --- | --- | --- |
| Owner high-none choice | Owner conversation relayed by Main; verbatim quote above and in PREREG | user | 2026-11-08 | HD8/X11a R4-02 | Select alternative D, retain targets 800, accept remaining pre-map stops | Direction, not empirical supply/power evidence |
| Owner autonomous recommendation/ADR direction | Same dated owner conversation; verbatim quote above | user | 2026-11-08 | Delegated decisions below | Take recommended choices without repeatedly asking; Main records ADRs for owner review | Does not waive council sign-off or authorise real execution |
| Pinned cost-only X2d metadata | `../cost-feasibility-v5.json`, exact source paths/raw SHA-256s at Scout `04d5e632ae34d8056c3df5116b9ba87cfbf99282` | internal-study | 2026-11-08 | X11a resource constant | 444 known-cost requests; maximum round USD/unit .0013386404347826087; 25%-margin price .001674 | Cost-only projection, not future tariff or latency evidence |
| Synthetic OC timing/redesign | `../oc-v5-spec.json`, final-r2 timing artifacts and retained earlier timings | internal-study | 2026-11-08 | Bounded OC implementation | Final worst-programme projection 100.01669184 core-hours; conservative threefold 300.05007552 under 384 | Timing only; no candidate OC probabilities or runtime guarantee |

Main owns actual append-only ledger IDs and ADR records. This worker does not modify ledgers/design/ADRs or claim those entries already exist.

## R4-02 feasibility outcome

Cost-only evidence projects no predictor–outcome relationship:

| Cohort | Units | Known requests | Unknown request costs | Known USD | USD/unit |
| --- | ---: | ---: | ---: | ---: | ---: |
| Fastapi postrun 8–10 | 92 | 200 | 0 | .117830244 | .0012807635217391305 |
| Axum postrun 1–5 | 109 | 244 | 0 | .097129326 | .0008910947339449541 |

Metadata headers excluded. Price **c = ceil(1.25 × maximum round USD/unit × 10^6)/10^6 = USD .001674**. Jev USD200 total (60/repository plus20 shared) permits total proxy P≤26881.72043010753 and per-repository P≤8960.57347670251; **fallback 4P≤40000 still binds total P≤10000**, including exposed continuity.

Alternative D's continuous necessary summed sample floor is **4912.2807017543855**, requiring mean C0 retention at least .49122807017543857 at that cap. Integer both-age floors are **1404 mandatory / 2106 selected**, evaluated after fixing the concordant winner, never promoting a runner-up. A 50%-added/25%-renamed composition needs P≈9356.725146198829; 50%-added/no-renames still needs P≈12280.701754385964 and fails. These are algebraic assumptions, not candidate facts or sufficient support-power proof.

## Issue-by-issue response

| ID | Rev5 change/evidence | Status for re-review |
| --- | --- | --- |
| R1-01 | PREREG §§2/9/11 retain full GUIDE, real orchestration/replay/label/group/scorer/seal writer, whole-pipeline plants and actual inventory-conditioned OC tables as later prerequisites. Conditional OC source now exists; it is not the real experiment driver. | Open, intentionally deferred |
| R4-01 blocking | PREREG §8, OC truth/event helpers: union counts only supported arms with an **own** true pooled metric below .80. Healthy companion support is not false support. Separate both-arm defects and homogeneous/heterogeneous boundary unions. Literal false-support fixtures passed. | Addressed; council decides closure |
| R4-02 major | Owner alternative D, cost-derived .001674, unchanged 800 targets/fallback ceiling; postwinner both-age 1404/2106 projection/hash-capacity gates, explicit remaining stops. Inventory emits unthinned cost counts and capped OC vectors. | Addressed with owner choice |
| R4-03 major | Fixed 28,080 scenarios; 256/512/1024 looks, simultaneous exact CP bounds, Go PCG streams, 399-draw OC approximation, 384-hour serial compute cap plus8 sensitivity hours. Full-workload timings and standalone Go closure/build wrapper exist. | Addressed; OC probabilities unmeasured |
| R4-04 major | PREREG §§5/10: ascending-hash round-robin labels; unstarted tail unsampled, begun unresolved gold missing; realised fraction k/N disclosed without claiming unconditional inclusion under informative stopping. Fixed 42-day clock, .65 utilisation,2 local days, separate Jev waiting and inventory projection gates. | Addressed; throughput/informative-stop threats retained |
| R4-05 major | At most12 hashed gold-none cases/repository, chunks≤3,≤4 chunks plus linking; bounded lines/summaries,≤15 sessions/repository=45. Reserve8330×18+45=149985. Concern/duplication outcomes explicitly sample-bound; ungrouped concerns unknown. | Addressed |
| R4-06 major | PREP uses ls-tree/cat-file exact blob materialisation, not archive. It postverifies selected path/type/mode/blob/symlink identity; `.codegraph` excluded. File/directory export-ignore/subst fixtures pass without altering bytes. | Addressed |
| R4-07 major | Create-once completed/refused seals bind authority/input/script/runtime pins, original argv/command traces/timing, anchors/revisions and output/refusal. `assemble_preparation.py` derives manifest only from verified seals; inventory rechecks them/original authority. | Addressed; seal/assembler fixtures passed |
| R4-08 major | Ownership manifest is a sealed completed/refused union for all five repositories. Content-refused candidate is excluded before ranking; mandatory refusal stops; operational refusal cannot become a gate. No substitution after projected-winner failure. | Addressed; union fixtures passed |
| R4-09 major | Persistent unchanged-object clones and immutable copied tools. Preserve pnpm internal relative symlinks; no links into original checkout. Storage/missing objects/tool/runtime and post-write drift operational; unsupported content/unsafe links remain content. Identical restored-input reruns use fresh outputs. | Addressed; real offline synthetic scan passed |
| R4-10 major | `prep_selftest.py` covers scoped authority, safe/unsafe materialisation, attributes, modes/bytes/path-set/links, case/Unicode, real pinned offline scanner, drift-before-call, seals/refusals/assembler/create-once, frontend build refusals/selected-file drift. Integrated suite invokes it once. | Addressed;14 preparation tests passed |
| R4-11 minor | Whole CPython install pinned; isolated -I -B bootstrap rejects stale local bytecode. Go selected package/file set re-derived before build, private HOME/cache/temp. Go differential now uses v5 runtime, -I -B and whole-install verification. | Addressed; actual stale-cache refusal observed/fixed |
| R4-12 minor | Correct-preferential shared-case removal w4 is binding alongside uniform; error-preferential is sensitivity. Both arms' denominator caps recomputed before each removal. | Addressed in fixed spec/source |
| R4-13 minor | Recall-cap residual tolerances q/(19+q):≈2.06%/3.06%/4.04%. One fresh same-model replacement for invalid A/B, within cell reserve; no retry valid unfavourable labels. Prepay6 slots/case, return unused; no replacement case. | Addressed |
| R4-14 minor | Full mixed-radix tables, seed/PCG/open-uniform/factor/draw order, isolated paired bootstrap streams, shared evidence-case masks, lowercase roles, literal seed/index/count fixtures. Real-analysis bootstrap seed/draw order also fixed. | Addressed; deterministic fixtures passed |
| R4-15 minor | Cross-age latent correlation {.5,1} binding; heterogeneous boundary Cold.76/C0.84 pooled.80, plus homogeneous.80. The illustrative .72/.88 P=R is infeasible at q=.8; change disclosed rather than clamped. | Addressed |
| R4-16 minor | Exact numerator/denominator fixtures, worst/best counts and deterministic .79/.81 checks exist; §9 additionally fixes exact full-driver near-boundary plants and required verdicts for R1-01. | Addressed at prereg scope; full plants still deferred |
| R4-17 minor | Inventory reports repository-only exposure per candidate/study without exclusion/overlap claims. Correct36=27+8+1; exact `study.startswith('inertia-bench/')`1350 filter and mis-scoped other-task human-confirmation-51.csv disclosed. Sealed v4 unchanged. | Addressed |

## Delegated decisions

All choices below use the standing owner direction quoted above; Main records corresponding ADRs. They were made before candidate ownership/supply facts.

| Question/options | Choice | Rationale |
| --- | --- | --- |
| Fix protocol and prep together (A), move freeze (B), or accept broken feasibility (C)? | Recommended A | Avoid supply-adjacent exposure before statistical decisions; no retrospective authority. |
| Apply projection floors before ranking or to fixed winner? | Postwinner, no reselection | Prevent composition/resource gates from opportunistically selecting a different high-none repository. |
| Keep21days requiring near-continuous throughput, only raise hours, or change wall planning? |42days, .65 utilisation,2 local contingency days;3000 model-hours unchanged | A meaningful wall band, with Jev60s/run allowance counted separately; worst projected≈36.68days. Unknown subscription latency can still stop earlier. |
| Chronological labels versus hash-prefix degradation; assume arbitrary inclusion weights? | Hash-first, prepaid6 slots and smaller begun prefixes; report k/N as realised fraction | Do not turn unsampled tails into unknowns or hide begun failures. Informative cost-dependent stopping remains an explicit inference threat, not a false unbiasedness claim. |
| Group every none case in nine sessions versus bounded/chunked grouping? | Hash12/repository, chunks3, linking,45 total sessions | Feasible bounded contexts; do not pretend complete concern discovery. 8330 sessions/cell leaves total149985. |
| Uncomputable beta10k×10k versus a bounded method? | Exact-ICC two-point latent distribution, PCG/factor copula, finite CP looks,399 OC bootstraps,384h+8h | Original measured beta minimum254.57h exceeded48h. Final maximum projection100.0167h, conservative threefold300.0501h fits384. Distribution changed explicitly, not claimed equivalent beta. Real intervals retain10000 draws; approximation error is not covered by event MC bounds. |
| Heterogeneous .72/.88 versus a feasible pooled-.80 boundary? | .76/.84 | P=R=.72 with q=.8 needs none false-positive probability1.12, impossible; .76/.84 preserves pooled boundary across every binding margin. |
| Flatten copied package links versus preserve dependency resolution? | New immutable root with internal pnpm links | Real synthetic --version exposed missing graceful-fs. Preserve all internal links and unchanged Node/CLI/marker bytes; do not patch missing packages ad hoc. |
| Materialised byte/mode drift as content failure or operation failure? | Operational | Written-tree corruption/mismatch is not an inherent property of pinned repository content and must not select candidates. |

## Exercised validation and corrections

Final isolated command is the launcher in `../EXPORT_RECIPES.md`, targeting `selftest.py`, with child TMPDIR on a temporary case-sensitive HFSX synthetic volume. **19 inventory/helper tests passed**, including the wrapper that ran **14 preparation tests, all passed** (final runs21.171s and20.749s respectively). The real pinned scanner ran only on synthetic repositories, offline; no candidate clone was scanned.

Pinned offline Go command, cwd `/Users/adamjackson/Projects/work/scout-x11`:

```sh
/Users/adamjackson/.local/share/mise/installs/go/1.26.10/bin/go test ./evaluation/x11-merge-association/frontend
```

Environment was the exact frontend recipe environment plus `/opt/homebrew/Cellar/git/2.56.0_2/bin` appended to PATH. **Passed** (`ok scout/evaluation/x11-merge-association/frontend 1.539s`). No prepared candidate frontend build ran.

Observed integration failures were fixed, not suppressed:

1. PREP passed resolved tool Paths to a runtime verifier expecting pin dictionaries; fixed the call to use configured pins.
2. Flattened tool-copy links lost transitive package resolution (`graceful-fs`). Published a new immutable root preserving pnpm layout; offline --version and real synthetic init/map then passed. Current root: `/Users/adamjackson/Projects/work/x11-tools/0a6e4750c3caeb121b49d1d6c09c35533238e095e477451bf6caa0635ea897bf`,2668 files including Node,109 internal links. Scanner-tree SHA-256 `d07a43d895415241f462fa654ce51ae531f51ff91878661584a1d88abd601140`; persistent-location metadata `48532bc6be66ac54c9714c8c0e79b2a79dd7155b78a42da0ab79ca165d928df3`.
3. Python3.14 resolve(strict=False) did not reject a cyclic symlink. Explicit ELOOP detection now rejects cycles while retaining safe dangling links; added coverage. Post-write mismatch is operational, and known module/runtime-loader failures cannot be content-gate candidates.
4. Initial Go fixture invocation used build-only PATH and lacked Git; test environment now explicitly adds pinned Git. The old Python differential created local bytecode; the isolated bootstrap rejected it. Removed that own fixture cache and fixed the differential to -I -B/v5 whole-runtime verification. Final integrated Python and Go runs passed. Regenerated frontend Go metadata after the test-source edit; selected build closure digest remained `5872bc692e73066642658c270ec22ec6efac44a6c559a5a2d72a679affb922db` (test source is not selected for build).

OC timing was separately authorised synthetic work: final-r2 artifacts record120 full studies/1872000 unique cases/14364000 block starts in1.502654625s,10000-bootstrap comparison, and360-study stress4.200638542s. Go source SHA-256 `9d34fd44f8b2fcd839a7e38cddb862ad6be15873e7c9ee9c8c146c8e5985c876`. Standalone OC closure99packages/745files; SHA-256 `bbd5a9b8203612d05b871e0852a669c3a821f1d4eb4808947855725af1f5ed05`. **No inventory-conditioned OC study ran; no support/false-support probability is claimed.** Wrapper build/refusal paths and full real-driver plants still require implementation-round exercise/sign-off.

## Review request and remaining boundary

Please review the final `rev5-digests.json` packet for **export preparation only**, including all R4 proposed fixes. Manifest excludes itself; the old gate-only rev5 manifest is obsolete. Prior review/digest/extraction v3/v4 files remain unchanged. R1-01 is not waived. Real ownership preparation, inventory, maps, labels/replay/model calls and product/default changes remain blocked pending their scoped zero-issue authority.
