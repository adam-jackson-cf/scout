# Council round 5: X11a merge-time association accuracy, preregistration rev 5, export-preparation scope

| Field | Value |
|---|---|
| Study | X11a merge-time association accuracy (HD8). It covers the accuracy part of REQ-176/REQ-177. Guard and recovery evidence belongs to X11b (`../x11b-guard-recovery/PLAN.md`, a plan only, not preregistered). |
| Digest manifest | `reviews/rev5-digests.json`, SHA-256 **38fd6c5f874faa81238db9dd9411b7ca0a624259a13ea2027c919631f90798d7**, computed by Main. The manifest excludes itself and has 1,534 entries. The reviewers have no hashing tool. Every digest below is quoted from the manifest; none was recomputed. Read-only jq checks confirmed that `inventory-inputs.json .recipe_sources` lists every rev-5 study script, spec, timing file, `cost-feasibility-v5.json`, `latency-basis-v4.json` and all seven `tool-pins-v5/*` files. |
| Requested scope | **Export preparation only.** This covers the model-free Clawpatch heuristic ownership scans at the pinned commits, the council-gated frontend build and the exposure-binding operation, producing sealed ownership exports. It does not cover inventory or selection, maps, labels, replay, OC runs or model calls. |
| Owner decisions in force | R4-02: "(b) High-none only on selected repo" (2026-11-08; alternative D). Standing direction R-064: owner choices are delegated to the agent's recommended option. Where this round would otherwise escalate, the recommended option is stated as the required fix. |
| Date | 2026-11-08 |
| Reviewer | innovation-council (anthropic/claude-opus-5-5) |
| Independent peers | innovation-challenger (openai-codex/gpt-6-astra:high); experiment-peer (anthropic/claude-opus-5-5:high) |
| Scope of review | (1) Status of R1-01 and R4-01…R4-17. (2) New issues in rev 5, specifically: per-arm false support, the alternative-D grid and band, the finite OC budget, the 42-day/65% wall ceiling, hash-order labelling and the unsampled tail, grouping chunking, seals and the assembler, the refusal union, persistent locations, the Python/Go closure, the preparation fixtures and each delegated decision. (3) Whether export preparation may be authorised. |
| Method | Read-only. Nobody edited a file or ran a script, test, build or git command, and nobody touched the candidate clones. The only computations were jq/grep over the study files and hand arithmetic. The drafter's reports of passing fixtures (19 + 14 Python, Go `ok`) are recorded as drafter claims and were not re-executed. |
| Line references | `PREREGISTRATION.md` (PREREG, 214 lines), `prepare_exports.py` (PREP), `assemble_preparation.py` (ASM), `inventory.py` (INV), `stages.py` (STG), `pin_tool_closure.py` (PIN), `prep_selftest.py` (PST), `selftest.py` (ST), `oc_v5.py`/`oc_v5.go` (OCPY/OCGO), `oc-v5-spec.json` (SPEC), `oc-v5-fixtures.json` (OCFX), `EXPORT_RECIPES.md` (REC), `round-4-response.md` (RESP), all as of the manifest digests. |

## 1. Status of prior items

"Closed" means the round-4 required fix is present in the rev-5 files. Each item was checked against the files, not taken from the response. Remaining gaps are tracked once, as R5 issues.

| ID | Status | Evidence / remaining gap |
|---|---|---|
| R1-01 | **Open (carried, deferred as planned)** | The GUIDE, full driver, replay, label, group, scorer and seal writer, the whole-pipeline plants and the inventory-conditioned OC tables are still absent. STG:11 `IMPLEMENTED = {'inventory'}`. PREREG §2, §9 and §11 disclose this. Open until the implementation round. |
| R4-01 | **Closed** | Per-arm definition: PREREG l.159, SPEC `false_support`, OCPY `false_support` (an arm counts only when its own min true metric is < .80). The union runs over defective arms only. Single-arm and both-arm defect truths: SPEC `truth_table` 3–38. Four literal fixtures: OCFX `false_support_cases`. Text residual: R5-10. |
| R4-02 | **Closed** | Owner alternative D is at PREREG l.7 and l.153. The price c = .001674 is derived from the pinned cost-only rows (`cost-feasibility-v5.json`, recomputed: ⌈1.25×.0013386404×10⁶⌉/10⁶). The band and counterexamples are at l.155–157. Post-winner floors 1404/2106 with no reselection: INV `select_repositories`, fixtures at ST:300–314. Disclosure residual: R5-06. |
| R4-03 | **Partially fixed → R5-03, R5-04** | Computability is resolved: 28,080 scenarios, fixed looks, the 399-draw approximation, a 384 h cap and timings that recompute to 100.02 h (×3 = 300.05 h). Open: the error control used by the gate (R5-04) and the frozen wrapper's input binding (R5-03). |
| R4-04 | **Partially fixed → R5-06, R5-07** | Present: hash-first labelling, the unsampled tail, clock start, Jev waiting, and the wall formula at l.199. Open: the projection cannot fail on its own (R5-06), and stop semantics for begun cases and shared labels are unspecified (R5-07). |
| R4-05 | **Partially fixed → R5-08** | Chunking and the 45-session reserve are present (l.115). The 12-case sample cannot observe duplication. |
| R4-06 | **Closed** | `ls-tree -r -z` plus `cat-file blob` materialisation (PREP:169–192), post-verification of path/type/mode/blob identity (PREP `verify_materialized`), and a directory/file `export-ignore`/`export-subst` fixture (PST:156–164). Scale residual: R5-01. Replacement-ref residual: R5-12. |
| R4-07 | **Partially fixed → R5-13, R5-01** | Create-once completed/refused seals exist (PREP:606–648), and the assembler builds the manifest only from verified seals (ASM:10–36). However, `verify_preparation_seal` accepts a hand-written seal with `commands: []` and no trace consistency; PST:25–89 demonstrates this by passing hand-built seals through the assembler (R5-13). Trace volume: R5-01. |
| R4-08 | **Closed** | Completed/refused union in PREP, ASM and INV. Operational refusals are rejected at INV:357. A content-refused candidate becomes ineligible and a refused mandatory repository becomes `not_testable_no_substitution` (INV:486–492, ST:246–261). The end-to-end fixture gap is in R5-13. |
| R4-09 | **Partially fixed → R5-02** | Persistent clones and an immutable content-addressed Node/Clawpatch root are present (PREREG l.40, `persistent-locations-v5.json`). There is an operational/content split (l.38). Open: misclassification, no terminal rule, and pins that cannot be restored. |
| R4-10 | **Partially fixed → R5-01, R5-13** | 14 preparation fixtures exist (PST:156–390). There is no scale fixture. The build success path, the exposures operation, the scanner-rerun branch and end-to-end refused seals are not exercised. |
| R4-11 | **Partially fixed → R5-12** | Whole-CPython pin, `-I -B` checks and exact Go selected-set re-derivation with private caches are present (PIN:96–213). Open: replacement objects, automatic PGO, and launcher import before verification. |
| R4-12 | **Closed** | Correct-preferential shared-mask removal is binding and error-preferential is a sensitivity (l.179; SPEC `missingness`; OCGO:96–120). |
| R4-13 | **Closed** | Tolerances q/(19+q) = .0206/.0306/.0404 are stated (l.126). There is one same-model replacement per invalid A/B session (l.105). |
| R4-14 | **Partially fixed → R5-10** | Mixed-radix index, seeds, PCG, open uniform, draw order and lowercase roles are all specified. The normative copula formula at l.173 contradicts the appendix. |
| R4-15 | **Closed** | ρ_age ∈ {.5, 1} is binding, with a heterogeneous .76/.84 boundary (l.159, l.171). Rationale residual: R5-09. |
| R4-16 | **Partially fixed → R5-11** | Exact count plants and .79/.81 are present, but the plants cannot tell the denominators apart. Full-driver plants are deferred under R1-01. |
| R4-17 | **Closed** | Repository-only exposure is reported per candidate/study (INV:579–582, :632–633). PREREG l.57 has 36 = 27 + 8 + 1, the `startswith("inertia-bench/")` filter and the mis-scoped record disclosure. |

**Status count:**
- Open: R1-01.
- Closed: R4-01, R4-02, R4-06, R4-08, R4-12, R4-13, R4-15, R4-17.
- Partially fixed: R4-03, R4-04, R4-05, R4-07, R4-09, R4-10, R4-11, R4-14, R4-16.

## 2. New issues

| ID | Severity | Location | Problem | Required fix |
|---|---|---|---|---|
| R5-01 | **blocking** | PREP:156–159 (`TRACE.append` for every command), PREP:169–192 (one sandboxed `git cat-file blob` process per blob), PREP:367 (revisions = Cold, C0 and **every** post-C0 commit to tip), PREP:641 (`'commands': list(TRACE)` in the seal), PREP:630 (no `MemoryError` handling); INV `inventory()` (sends every revision's full owned-path list to the frontend in one JSON request) | **As implemented, the authorised operation probably cannot finish on the mandatory repositories.** Every revision rebuilds the whole tree with one `sandbox-exec`+git spawn per file, keeps a trace record per spawn in memory, writes them all into the seal, and emits a full owned-path list per revision. Process spawns and trace entries ≈ files × revisions. With assumed sizes of about 7×10³ files × about 3×10³ revisions (django) and 10⁴–10⁵ files (TypeScript) [INFERENCE: public tree sizes and commit rates, not measured], this means about 10⁷–10⁸ spawns (days to weeks) and tens to hundreds of GB of trace and output. A `MemoryError` escapes the caught exceptions and leaves an empty create-once seal. Anything that fails is operational, so it reruns without end (R5-02), the assembler rejects it, and inventory never runs. The fixtures use one-commit synthetic repositories (PST:110–126), so they cannot detect this. Fixing it after preparation has started would change signed preparation bytes after supply-adjacent exports exist for the smaller repositories, which is exactly what alternative A was chosen to avoid. | (1) Limit revisions to Cold, C0 and the first 2,000 post-C0 commits. Inventory then takes the envelope length from `rev-list` and drops or relabels `envelope_metrics` as unscanned. (2) Materialise each revision with one `ls-tree -r -z` plus one `cat-file --batch`, or incrementally from `diff-tree` with full post-verification. (3) Record one trace entry per revision step (command template, count, digest of the OID list, exit code, stdout/stderr digests). (4) Write ownership as per-revision records or deltas behind a digest index, and stream them to INV and the frontend (FE request schema updated). (5) Catch `MemoryError`/`BaseException` (except `SystemExit`) as operational and still write the seal. (6) Add a synthetic scale fixture (≥ 80,000 files × ≥ 20 revisions with the real pinned scanner offline) recording elapsed time and peak RSS, plus a stated per-repository preparation projection at 2,002 revisions × an assumed file bound. Pin the Node heap size. |
| R5-02 | **major** | PREP:56–63 (blacklist classifier), PREP:386–398 (rerun), PREP:189 (gitlink → content), PREP:207 (Unicode probe), PREP:443–447 (binary written before buildinfo); PREREG l.38, l.40, l.46 ("gitlinks … follow product"); INV:357; `inventory-inputs.json` tools (`/opt/homebrew/Cellar/git/2.56.0_2/…`, `/usr/bin/sandbox-exec`) | **Refusal classes are misassigned and operational refusals have no end.** (a) Any exit 1–127 whose stderr lacks a blacklisted marker is a content candidate. `EROFS` and `EDQUOT` are not on the list (challenger), so a storage fault that reproduces becomes a scientific exclusion. In the other direction, a deterministic content failure whose stderr contains `ENOENT` (for example a scanner stat on a retained dangling symlink) or `out of memory` is classed operational forever [INFERENCE: scanner behaviour not observed]. Stderr that differs between runs (timestamps, paths) can never "reproduce identically". (b) No rule bounds operational reruns, so the study can stall with no defined outcome. (c) Some pins cannot be restored without repinning: a macOS update replaces `/usr/bin/sandbox-exec`, and `brew upgrade`/`cleanup` removes the Cellar git (with unpinned Homebrew dylibs). PREREG l.40 forbids repinning. (d) An interruption between `go build` and `go version -m` leaves the fixed-path binary, and every later attempt then fails create-once (challenger). (e) A gitlink at any revision refuses the repository, contradicting l.46. Unicode-alias refusal depends on the scratch filesystem's normalisation. A content defect in a commit beyond segment position 2,000 can make a mandatory repository not testable for content outside observation. | (a) Classify every scanner process failure as **operational**; content classes become the structural materialisation refusals only (unsafe/absolute symlink, non-UTF-8 path, unsupported non-gitlink entry, NFD/case alias evaluated as a property of the paths whatever the filesystem). (b) Preregister a terminal rule now: at most 3 identical-recipe attempts per operation after restoration. After the third identical-step failure the repository is **preparation-infeasible (operational)**: mandatory → X11a not testable (operational, reported as such); candidate → excluded before ranking with that reason. This is fixed before any scan. (c) Copy git (with its dylib closure) into the immutable tool root. Record the macOS build and the `sandbox-exec` digest in every seal. Preregister that a change to an OS-provided binary needs a council micro-round to repin that binary only. (d) Write the build to a per-attempt directory and promote it atomically after buildinfo succeeds, or quarantine unsealed outputs with a frozen recovery procedure. (e) Materialise gitlinks as empty directories (what a non-recursive checkout produces) and count them. Limit content gating to scanned revisions (R5-01). Fixtures for each. |
| R5-03 | **major** | OCPY `main` (non-timing branch: `stages.verify_authorization(..., scope='protocol')`, then `--vectors` passed through unchecked; `--bootstrap`, `--timing-replicates` and `--recipe` taken from argv); STG:41–46 (`reviewed_inventory` optional in protocol scope); SPEC `input_vector_contract` | **The frozen OC wrapper does not bind its inputs (challenger).** Any vector file, any bootstrap count and an unpinned recipe can drive an apparently authorised `oc` run. The build record does not hash the vectors. Under PREREG §2 these bytes are frozen at preparation authority, so a later fix needs a renewed preparation round after exports exist. | In `oc`/`sensitivity` mode: require `reviewed_inventory`, verify the original inventory authority and output seal, and require the vectors to equal that output's non-null `oc_vectors`. Bind the inventory and vector digests into the build record. Reject `--bootstrap` ≠ 399 and any `--timing-replicates`. Require `--recipe` to equal the pinned `oc-build-recipe-v5.json` path and digest from `recipe_sources`. Add refusal fixtures (altered vectors, unreviewed inventory, overridden bootstrap). |
| R5-04 | **major** | PREREG l.181, l.183 ("Bonferroni family error ≤ .05 across all looks/scenarios/events"); SPEC `sequential_MC`; OCGO:172; `oc-v5-timing-10000-r2.json` (12 studies, 2 per scenario) | **The OC gate uses the wrong error control, and its effective thresholds are not disclosed.** The pre-map decision is "feasible iff every scenario and every relevant event passes". That is an intersection–union decision: the probability of a false "feasible" is at most the per-scenario per-event test level, summed only over the 3 looks (Berger 1982, *Technometrics* 24:295). Bonferroni over 505,440 intervals (α = 9.89×10⁻⁸, z ≈ 5.33) is not needed for that error. It raises the pass requirement at n = 1,024 from observed support ≈ .83 to ≈ .86 and lowers the tolerated false-event rate from ≈ .034 to ≈ .02 [normal/Poisson approximations]. That turns adequately powered scenarios into "unresolved → not testable". The thresholds are not stated (experiment-peer). The "paired small-grid 10,000-draw comparison" (l.181) has n = 2 per scenario and is a smoke test. | Gate with the intersection–union rule: one-sided per-look level .05/3 for every relevant event in every scenario, with no cross-scenario or cross-event correction. Keep the Bonferroni-simultaneous intervals as descriptive reporting only. State the resulting effective thresholds at 256/512/1,024 in §9. Relabel the 10,000-draw artifact as a smoke test, or size it to ≥ 256 paired studies on the uniform, boundary and .78 scenarios with verdict discordance reported. Recompute the compute projection, since earlier resolution reduces work. |
| R5-05 | **major** | PREREG l.130 ("Beyond 20-commit correlation remains a threat"), l.173 (merge-keyed latent), l.187; INV `projection_vectors` (per-position counts only); SPEC `latent_distribution` | **Directory-keyed long-range dependence is neither generated by the OC nor captured by the primary interval (experiment-peer).** Residual clusters recur in the same normalised directories across distant commits, and thinning retains all-added clusters preferentially [INFERENCE: e.g. TypeScript test-case directories]. Given a fixed catalogue and own history, the decisions in a directory are strongly correlated. The commit-block bootstrap treats them as independent across blocks, so support lower bounds can be anti-conservative and boundary false adoption can exceed .05 without detection. Inventory is frozen at preparation authority, so the needed inputs must be emitted now. | Inventory emits a per-case normalised-directory hash in the projection, and the top-directory share of retained proxy cases is reported as a distribution check under rule 3. Recommended (R-064): the primary support lower bound is the **minimum** of the commit-block bound and a directory-cluster bootstrap bound (same 10,000 draws, fixed seed; the best-case negative bound uses the maximum). The OC adds a binding directory-keyed latent effect (ICC_dir ∈ {0, .10}) and computes that same min-bound estimator. The grid and compute budget are recomputed, and the change is stated before freeze. Cite Cameron, Gelbach & Miller (2011, *JBES* 29:238) for multiway clustering. |
| R5-06 | **major** | PREREG l.197–199; INV:528–538; `latency-basis-v4.json` | **The wall/hour projection cannot fail on its own, and the binding risk is never measured (experiment-peer).** Only P_all varies. Once the fallback gate holds (4P_all ≤ 40,000): H ≤ 2,845.69 h, wall ≤ 36.68 days and Jev ≤ USD 86.96. So `subscription_model_hours`, `collection_wall_days`, `jev_total_cost` and `reserved_sessions` are vacuous. The real driver is mean label-session time. Break-even is (3,000×3,600 − 31×40,000 − 5,400)/149,985 = **63.7 s** for 3,000 h and **72.0 s** for 42 days. l.199's "90 s labels can break this budget" understates this. Subscription quota and throughput are not projected. The 4P projection assumes every selected proxy cluster escalates, with one retry, in both ages; that assumption is not stated in the text. | Preregister a pre-map **latency and throughput calibration**, run after the GUIDE freezes and before maps, on the GUIDE's synthetic examples only. It covers a fixed N (e.g. 60) of A, B and adjudicator sessions with the pinned CLIs, models and efforts, and records sustained sessions per hour. The measured mean × 1.25 replaces 60 s in H and days, and gives a projected k per cell. Stop before maps if H > 3,000, days > 42, or any cell's projected k is below 1,404 (mandatory) or 2,106 (selected). Correct l.199. State the 4P assumption, why it is retained (a mid-replay ceiling breach cannot degrade gracefully), and the quota threat. |
| R5-07 | **major** | PREREG l.102, l.105, l.113, l.195, l.201 | **The hash-prefix stop can recreate the R4-04 failure, and shared-label accounting is undefined.** (a) At an hours or wall stop "no new requests" means begun cases stay unresolved and count as missing. The number of begun cases is unbounded (a pipelining scheduler can have many routed but unadjudicated cases), so every cell can breach the recall cap of about 2% at once. (b) Labels are per repository×age (arm omitted) but reserves are per arm-cell. Neither the paying cell nor the 6-slot prepayment is assigned, and nothing says whether a label beyond one cell's frontier enters that cell (which would give a non-prefix sample biased towards shared cases; experiment-peer). (c) "Smallest 20%" audit selection needs the final agreed set, which conflicts with prepayment and stops (experiment-peer). | (a) At most one in-flight case per concurrency slot (≤ 6 in total). Each case runs to completion (A/B, replacements, adjudication, audit) before its slot starts another. Hours and wall stops become "stop starting" once the remaining allowance falls below a fixed finishing allowance (6 cases × 6 sessions × the calibrated latency from R5-06), so begun cases finish. (b) Pool reserves per repository×age (3 × 8,330) with a charge-once rule. Each cell's analysed sample is exactly its own eligible cases at or below its own frontier hash. (c) Audit inclusion iff SHA-256(`x11a-audit\0`+case ID) < 0.2·2²⁵⁶, decided at case start. The minimum-40 top-up comes only from remaining budget; otherwise report the shortfall. |
| R5-08 | **major** | PREREG l.115; REQ-176 (`scout-design.html` l.693: "report duplicate pending areas per real new concern"); `../x11b-guard-recovery/PLAN.md` | **The grouping sample cannot observe duplication, so the REQ-176 duplication outcome is biased towards zero by construction (experiment-peer).** Twelve hash-random gold-none cases per repository, taken from thousands, rarely include two cases of one concern in one arm×age state, so k > 1 is almost never seen. Evidence-digest deduplication keeps Cold and C0 copies. How k is counted across states is undefined. Chunks of 3 are not justified by any context budget: 12 cases at ≤ 3×30 lines fit one session. | Recommended (R-064, conservative): declare concern-level duplication **unmeasured in X11a**. Record it as an outstanding REQ-176 element in the X11b PLAN (as with guard/recovery), and drop the 12-case k and duplication rates. Retain or remove the 12-case grouping as a descriptive novelty illustration in a single session. Acceptable alternative, only if fully specified before freeze: group the origin clusters of **all** pending areas within each primary-arm×age state under a stated context budget, with k counted within that state and deduplicated by cluster identity. |
| R5-09 | minor | SPEC `truth_margins` (`existing_failure: abstain`, `infeasible_companion`); PREREG l.159; RESP delegated row ".72/.88 … impossible" | **The generator has no wrong-ID-on-existing mode.** Precision errors come only from gold-none cases. At none share .2 the .75 precision defect therefore forces recall down to .75, so the intended precision-only defect is never simulated. ".72 impossible" is an artefact of the generator, not a property of the data. The likely effect on pass/fail is small [INFERENCE]. | Split existing failures into a fixed wrong-ID fraction and abstention, so precision targets hold with recall at .90 for every q; this also makes .72/.88 feasible. At minimum, reword "impossible" to "infeasible under the abstain-only generator" and list it as a limitation. |
| R5-10 | minor | PREREG l.173 vs SPEC `latent_distribution.normal_factor`/OCGO; PREREG l.161 vs l.183/SPEC `uniform90_pass` | **The normative texts contradict each other.** (a) The copula's last term reads `Z_age`, where it should be `Z_age_arm`; read literally, the variance is 1.5 at a = c = .5 (challenger, experiment-peer). (b) l.161 makes Jev-only power a condition only for the Jev-only recommendation, while l.183, the spec and the code require both arms for the uniform-.90 pass. | (a) Correct the term to `Z_age_arm`. (b) Keep both arms (conservative, as implemented) and edit l.161 to match. |
| R5-11 | minor | OCFX `count_plants` | **The count plants cannot catch swapped denominators.** Every plant has A_known = E_known and U_unassigned = 0, so swapped precision/recall denominators, or U_all confused with A_unknown, still pass (experiment-peer). | Add plants with A_known ≠ E_known, U_unassigned > 0, A_unknown ≠ U_all, and none-case assignments. |
| R5-12 | minor | PREP:140–150 and INV `git_env` (no `GIT_NO_REPLACE_OBJECTS`); frontend and OC recipes (`-pgo` unset, so `auto`); REC launcher plus PIN `isolated_argv` (helper digest computed live, parent imports helper first); `persistent-locations-v5.json` (only anchor commit objects verified) | **Execution-closure residuals.** (a) Replacement refs and grafts can make `ls-tree`/`rev-list` read substituted objects for a signed commit SHA (challenger; installed `git-replace` docs). (b) A `default.pgo` in a main package would change compilation without changing the recorded closure (challenger; Go `alldocs`). (c) The launcher imports `pin_tool_closure` before any check. The child's "helper digest" is the helper's current hash, not a reviewed pin, and timestamp-valid local bytecode can be read before rejection (challenger, council). (d) Envelope object connectivity, shallow/partial-clone status and absence of `refs/replace` were not verified. | (a) Set `GIT_NO_REPLACE_OBJECTS=1` everywhere. Refuse (operational) when `refs/replace/*`, `info/grafts` or a `shallow` file is present. Add a replacement-ref fixture. (b) Set `-pgo=off` in both recipes. (c) Use a stdlib-only bootstrap that checks the helper against the digest in `inventory-inputs.json`/authorisation and rejects any `__pycache__` before the first study import. Add a planted-cache sentinel fixture. (d) Record `rev-list --objects --missing=error` connectivity over Cold..segment end, plus non-shallow, non-promisor status, in the persistent-location metadata. |
| R5-13 | minor | PREP `verify_preparation_seal`; PST:25–89 (hand-built seals with `commands: []` pass the assembler); PST overall; PREP:182 (silent `.codegraph` skip) | **Seal verification and fixture residuals.** (a) The verifier never checks that the trace matches the operation: one ls-tree, materialisation and sandboxed `init`+`map --source heuristic` per revision with exit 0 and `revision_count` matching; go list, build and `version -m` for the build. So self-asserted seals pass. (b) Not exercised: `build_frontend` success (only refusals), the `exposures` operation through `execute()`, the scanner-rerun branch, and content-refused or operational seals produced by `execute()` going through ASM → `verify_artifact_bindings` → `inventory()` (ST:246–261 uses hand-built dicts). (c) Committed `.codegraph` entries in candidates are skipped without being counted. | (a) Add per-operation trace-consistency checks to `verify_preparation_seal`. (b) Add those fixtures; the synthetic Go project in PST:352–378 can be built for real through `build_frontend`. (c) Count excluded `.codegraph` entries per revision in the output and seal. |
| R5-14 | minor | PREREG l.7, l.9, §8 (owner directions and cost/OC sources cited "pending" with no ledger ID); `docs/research-ledger.md` (no R-063/R-064 entries in either worktree); RESP "Delegated decisions" | **Ledger rule 3 is not met, and some delegated decisions are not recorded.** PREREG freezes at preparation authority while citing un-ledgered sources. Delegated changes with no recorded rationale: the OC budget 48 h → 384 h, looks 4,096 → 1,024, and the .65 utilisation figure (experiment-peer). | Main appends the research-ledger entries (owner R4-02 choice; the standing direction R-064; cost-feasibility-v5; OC timing), and PREREG cites their IDs before sign-off. Add the three missing delegated decisions with rationale to the response. Record the ADR(s) Main promised for owner review. |

**Totals:** 14 new issues (1 blocking, 7 major, 6 minor) plus R1-01 carried.

Considered and not raised:
- **Boundary scenarios are homogeneous (all cells at .80 or .76/.84), the least adverse boundary configuration.** The council prescribed this in R4-01. The .78 single-cell defects cover near-boundary false support, and a single-cell .80 boundary would be mostly undecidable by design.
- **Unsafe-symlink refusal versus product parity.** This was prescribed in R4-09. It is kept as a content class, limited to scanned revisions (R5-02e).
- **Model drift confounded with replay position (experiment-peer).** It is listed under "subscription drift" in the threats. Recording the reported model per call is already required for labels and should extend to Terra in the implementation round.
- **4P fallback conservatism deciding feasibility.** The owner accepted composition-dependent stops. Only the disclosure in R5-06 is required.

## 3. Review of the specific rev-5 elements

- **Per-arm false support.** Correct and implemented consistently. The defective-arm union, both-arm defects and separate boundary events are present (§1, R4-01). The text contradiction is in R5-10(b).
- **Alternative-D grid and feasibility band.** Grid: 2×2×3 = 12 none vectors. Scenarios: 12×39×2×3×2×5 = **28,080**. Truths: 1 + 1 + 1 + 24 + 12 = 39.
  - Floors are ⌈800/(.6×.95)⌉ = 1,404 and ⌈800/(.4×.95)⌉ = 2,106. The summed floor is 4,912.28, so average C0 retention must be ≥ .4912 at P = 10,000.
  - Examples: 50% added/25% renamed needs P ≈ 9,356.7; 50% added with no renames needs 12,280.7 and fails. The feasible band for typical compositions is therefore narrow (about 9.4k–10k total proxy clusters). This is owner-accepted and must be disclosed as such (R5-06).
- **Finite OC compute budget.**
  - 28,753,920 studies × 1.502654625 s/120 = 360,060 s = **100.02 h**; ×3 = 300.05 h < 384 h.
  - Stress: 93.20 h. The 10,000-draw timing projects about 630 h maximum, which justifies 399 draws in the OC.
  - Computable. Error control: R5-04. Input binding: R5-03.
- **42-day wall at 65%.**
  - H = (60×149,985 + 31×40,000 + 5,400)/3,600 = **2,845.69 h**.
  - Wall: (2,845.69 + 400)/93.6 + 2 = **36.68 days**. The arithmetic holds.
  - The gate is vacuous given the fallback cap (R5-06).
- **Hash-order labelling and the unsampled tail.** The design is sound: the arm-omitted key, the tail counted as unsampled rather than unknown, π = k/N_actual, and the informative-stopping caveat at l.102. Stop semantics and accounting: R5-07.
- **Grouping chunking.** Reserve: 8,330×18 + 45 = **149,985**. Feasible, but the duplication outcome is unobservable (R5-08).
- **Seals and assembler.** Structure is correct. Verifier strength: R5-13. Volume: R5-01.
- **Refusal union.** Correct for PREP/ASM/INV. Classification and termination: R5-02.
- **Persistent locations.** Clones are under `x11-candidates/` with canonical origins and no alternates. The Node/Clawpatch immutable root (modes 0444/0555, internal pnpm links, offline `--version` 0.7.3) is accepted. Residuals: git/sandbox-exec (R5-02c) and connectivity (R5-12d).
- **Python/Go closure.** Whole-install CPython pin and isolated launch are accepted, as is the Go selected-set equality in the private environment. Residuals: R5-12.
- **Preparation fixtures.** Good coverage of materialisation safety, attributes, drift, authorisation scope and create-once. Gaps: R5-01 and R5-13.
- **Delegated decisions (R-064):**

| Decision | Council finding |
|---|---|
| Alternative A (fix the protocol and preparation together) | Endorsed |
| Post-winner floors, no reselection | Endorsed |
| 42 days / .65 / 2 local days | Endorsed as a ceiling; projection vacuous (R5-06) |
| Hash-first labels, 6-slot prepayment, k/N | Endorsed in principle; R5-07 |
| 12-case, chunk-of-3 grouping | Not endorsed (R5-08) |
| Two-point latent, PCG, CP looks, 399 draws, 384 h | Endorsed for computability; R5-03, R5-04, R5-05 |
| .76/.84 boundary | Acceptable as a boundary; rationale wrong (R5-09) |
| Immutable root with internal pnpm links | Endorsed |
| Byte/mode drift classed operational | Endorsed |
| Unrecorded: budget 48→384 h, looks 4,096→1,024, .65 | R5-14 |

- **Model-free, no outcome leakage.** Confirmed for the reviewed code:
  - `map --source heuristic` with fresh state and HOME.
  - PATH holds only the pinned git.
  - The sandbox denies network access.
  - No model CLI and no selection, gold or predictor–outcome computation in PREP or ASM.

## 4. Export-preparation authorisation

What the reviewed revision does well:
- The R4-01 contradiction is resolved.
- A computable OC.
- Exact-blob materialisation.
- Create-once seals and a verified-seal assembler.
- A refusal union through to inventory.
- Persistent, content-addressed tool locations.
- Whole-runtime Python and selected-set Go enforcement.
- The owner's alternative D is implemented with post-winner floors.

Why it may not run now:
1. **R5-01:** the authorised operation probably cannot complete on the mandatory repositories, and a fix mid-preparation would create exactly the supply-adjacent exposure that alternative A exists to avoid.
2. **R5-02:** operational and content classification can turn storage faults into scientific exclusions and has no terminal path.
3. **PREREG freezes at this authority** (§2, STG:12–14), so the statistical defects (R5-04 to R5-08) and the frozen-helper defect (R5-03) must be fixed first.

## Verdicts

**(a) Export-preparation authorisation: `not authorised`.** This covers:
- `PREREGISTRATION.md` 9eb2c2ddb512be6ef11b5ace716df65fd6e38037da67b1256f6e6e6596642bba
- `EXPORT_RECIPES.md` 0bd6d1b6a2d8dee7b0be633f9536bc5a7fde0edfb3cb5d972eed6d1a7c5f14cf
- `README.md` 8f3c2a057aa132c5fa8dcaefe7784b9125c10388e50e014280b87b5b047fde9c
- `prepare_exports.py` 5b69f7ca5d6fa49aeb5209ce1b5f70b4b05acaf027823b0236f661aad7ba1a1f
- `assemble_preparation.py` 54022f21ace9c43f8bfbcb1dcb5db92b1ae025d46a575b40b760dd0cfb618885
- `prep_selftest.py` c0963951e2c8c6d1a1962672d9b6e58033ec7957367cbe5af35d9ac2cb6a42c1
- `pin_tool_closure.py` d5461dfc14a49435fce9b8d546feb67e7392afafedcc9c4bab221e6a0b0c6eb6
- `stages.py` 6bb1fa6e24dc30a77c8c0fc8fdaca83aa7be50f32b9acaf5e0938ea6654bc682
- `inventory.py` 29b072c34b820e2bd52d51467a80c9edf58341cf0553a4b6e816111e80108fcd
- `selftest.py` dc6e294976c447e19030ff17d70afb41a6c5fb936b56a6d06916e35a9734ee66
- `extract_exposures.py` 966f7ed55d0a30c3d5cc4e2deea5c78b912f98827e261bbdb9b265c44749c6b8
- `frontend/main.go` bd83e1c3a5c4c68a6daee2fa2b592e9a83fca889e6ddfaba9336c079f0b965c0
- `frontend/main_test.go` 58b2f0720633a391672084265397bff7575121d3dfa032b4839cf0dda1dc5661
- `inventory-inputs.json` 9263ac7fb22e49317357e1996160a6fb86476202f5c0022e1b98382a09904b45
- `exposure-source-index.json` bccbfd4b8a2e42c9a91cdd9ead0c3f939c5606950a5fcb9844f8d4b6a141c3d8
- `exposure-contexts-v4.json` 2dd5b6b7d6665d64519585acdb63ed6eafadc3092bda25a38433616f896199fb
- `latency-basis-v4.json` bdc3cef935d65a20b4b0b9f34bb68c881d29da4960e74e5ee5b60549f0894104
- `cost-feasibility-v5.json` 1637256f02d54164c73eff82abb8648b4f70f4924701c9c29ee5a54e6c76c1d0
- `oc-v5-spec.json` c3618c05bddfd8ffd6386902468b1418cb6bebf57bfdced5763f111fba610943
- `oc-v5-fixtures.json` b4772b40fae20a04785efac8dc07deff6cc9f5645cacbbeb108006c16f21c1bc
- `oc_v5.go` 9d34fd44f8b2fcd839a7e38cddb862ad6be15873e7c9ee9c8c146c8e5985c876
- `oc_v5.py` fcc1f7dc99a2b87288f84dc32a351400bb1c7ad651290569ba233ce7cfe6c55a
- `oc-v5-timing-final-r2.json` d3ce6d8cc343fb49487ffdb3a91e0e08beaa1664f8d7abd415936399103bb86d
- `oc-v5-timing-10000-r2.json` 2e8c5dc2be122032ce271cb8568fc5fa81650427ff86cbd8bbb4d2abe5cbe93b
- `oc-v5-timing-stress-r2.json` 68bdc4b7b229b5bd344335ae3f6d9d67a1d494b06ffb0976a8faf779e6783d4e
- `tool-pins-v5/clawpatch-tree-v5.json` d07a43d895415241f462fa654ce51ae531f51ff91878661584a1d88abd601140
- `tool-pins-v5/frontend-build-recipe-v5.json` 2260f639e36eeb1577c06f0fa0acb4972562ff1f0b9c770a3eae2dcb666a1e72
- `tool-pins-v5/frontend-go-closure-v5.json` 5872bc692e73066642658c270ec22ec6efac44a6c559a5a2d72a679affb922db
- `tool-pins-v5/python-runtime-v5.json` 50119d209fe607e2db940a8ea7be5b07eb2b8897c727224e07928aa2f7ffb1e3
- `tool-pins-v5/oc-build-recipe-v5.json` c963611cf55eb95be94cb5940b84f346e7c460a3b97bb5c06d503c4e6e9c87db
- `tool-pins-v5/oc-go-closure-v5.json` bbd5a9b8203612d05b871e0852a669c3a821f1d4eb4808947855725af1f5ed05
- `tool-pins-v5/persistent-locations-v5.json` 48532bc6be66ac54c9714c8c0e79b2a79dd7155b78a42da0ab79ca165d928df3
- `exposure-extraction-v4/seal.json` 96fcfcbfada87faaa6624d9124fb3c88d55d714a8cd0bf79ce72a114e32f8510, and the five v4 outputs as pinned in it
- `reviews/round-4-response.md` adb61dc48957cfa68d35f921175edeb0dbba29b619e02e4980407076bc36a19a
- `../x11b-guard-recovery/PLAN.md` d3a5c666d50e1e6087aeb43896b7f5ad663911b88e3ccde0d885fa6f47791609

All are as listed in manifest 38fd6c5f874faa81238db9dd9411b7ca0a624259a13ea2027c919631f90798d7. No ownership scan, frontend build, exposure operation, inventory, OC run, map, label, replay or model call may run. Inventory still needs its own artifact-binding round after a future authorised preparation.

**(b) Overall: `15 issues open`.** That is R1-01 carried (deferred implementation; open until the implementation round) plus 14 new (1 blocking, 7 major, 6 minor). Not signed off. HD8 stays `proposed`. The ledgers, design and product are unchanged by this round.

## Council synthesis

### 1. Recommendation
Revise to rev 6 and request export-preparation round 6. Rebuild the preparation path so it is feasible and terminates (R5-01, R5-02). Bind the frozen OC wrapper (R5-03). Make the four statistical fixes that the preparation freeze would otherwise lock in (R5-04 intersection–union gate, R5-05 directory-robust bound, R5-06 latency calibration gate, R5-07 stop semantics), and choose the R5-08 option. Clear R5-09 to R5-14. Keep everything closed in §1.

### 2. Alternatives
- **A. Recommended: fix everything, then one preparation round.** Nothing is tuned after supply-adjacent facts exist. It costs one more round.
- **B. Authorise preparation now for gin, cobra and prometheus only (small trees).** It would surface real-scale behaviour early. Rejected: it creates candidate supply-adjacent exports before PREREG freezes, and the R5-01 fix changes signed preparation bytes anyway.
- **C. Permit a content-neutral capacity check on the clones (`ls-tree` counts at C0).** Rejected as unnecessary. Restricting revisions to Cold, C0 and 2,000 bounds R by construction, and a synthetic benchmark gives the per-revision cost without touching candidates.
- **D. R5-08: measure duplication via all-pending-origin grouping instead of declaring it unmeasured.** Scientifically richer, but it needs a new, budgeted design before freeze. Acceptable only if it is fully specified in rev 6.
- **E. R5-04: keep Bonferroni and only disclose the thresholds.** Valid but needlessly conservative, and it raises the probability of a pre-map stop with no error-control benefit. Not recommended.

### 3. Evidence
- **Scale:**
  - PREP:156–159 trace per command; PREP:192 one `cat-file blob` per blob; PREP:367 revisions to tip; PREP:641 trace written to the seal; PREP:630 `MemoryError` not caught.
  - PST fixtures use one-commit repositories (PST:110–126).
- **Classification:**
  - PREP:56–63 blacklist; PREP:386–398 rerun.
  - `/usr/bin/sandbox-exec` and `/opt/homebrew/Cellar/git/2.56.0_2/bin/git` pins in INP; PREREG l.40 "no repin".
- **OC wrapper:** OCPY `main` passes `--vectors` and `--bootstrap` through; STG:41–46 makes `reviewed_inventory` optional.
- **Gate arithmetic:**
  - α = .05/(28,080×3×6) = 9.892×10⁻⁸ (PREREG l.183; OCGO:172).
  - Intersection–union principle: Berger (1982) *Technometrics* 24:295–300.
- **Projection vacuity:** H ≤ 2,845.69 h and wall ≤ 36.68 days whenever 4P ≤ 40,000 (INV:528–538). Break-even 63.7 s / 72.0 s.
- **Grouping:** PREREG l.115; REQ-176 at `scout-design.html` l.693.
- **Count plants:** OCFX `count_plants` all have A_known = E_known and U_unassigned = 0.
- **Comparison size:** `oc-v5-timing-10000-r2.json` studies 12, scenarios 6.
- **Ledger:** no R-063/R-064 or rev-5 source entries in `docs/research-ledger.md` (grep, both worktrees).
- **Recomputed constants:**
  - c = .001674; P_r ≤ 8,960.57; total P ≤ 26,881.72; fallback P ≤ 10,000.
  - Floors 1,404/2,106; summed floor 4,912.28.
  - Reserve 149,985; OC 100.02 h / 300.05 h.

### 4. Risks
- Even after the fixes, the alternative-D band is narrow (total proxy about 9.4k–10k for typical mixes), so a pre-map not-testable outcome remains likely. The owner accepted this.
- The R5-05 directory-robust bound may lower power, which makes the OC stop more likely. That is the honest consequence of concentrated dependence.
- The R5-06 calibration may show label latency above 64 s, ending X11a before maps. That is cheaper than finding out after replay.
- R5-01's incremental materialisation must still verify exact tree identity per revision, or R4-06 regresses.
- The R5-02 terminal rule makes persistent operational failures decisive. It is fixed in advance and reported as operational, never as content.
- Reversibility: everything is reversible. No candidate scan, build, inventory or model call has run. The v3/v4 extraction artifacts and all earlier reviews are preserved.

### 5. Execution outline
1. **PREP/INV/FE (R5-01, R5-02, R5-12, R5-13).**
   - Revision scope Cold + C0 + 2,000; batch or incremental materialisation; per-revision trace; streamed per-revision ownership records with INV/FE consumption; `BaseException` sealing.
   - Scanner failures operational, terminal rule (3 attempts), gitlinks as empty directories, NFD alias as a path property.
   - Atomic build promotion; `GIT_NO_REPLACE_OBJECTS`, `-pgo=off`, pre-import helper and cache verification; git copied into the immutable root; OS build recorded in seals; trace-consistency verification.
   - Check: synthetic scale benchmark (≥ 80k files × ≥ 20 revisions) with time and peak RSS reported. New fixtures for each refusal path, build success, exposures, scanner rerun and end-to-end refused seals through `inventory()`.
2. **OCPY (R5-03).** Inventory-seal and vector binding; fixed bootstrap and recipe. Check: refusal fixtures.
3. **PREREG and SPEC (R5-04 to R5-11).**
   - Intersection–union gate with stated effective thresholds.
   - Directory hash in the projection, a min-bound estimator and the OC directory mode, with the budget recomputed.
   - Latency/throughput calibration gate and the 4P disclosure.
   - Bounded in-flight work, finishing allowance, pooled charge-once reserves, prefix-only samples, per-case audit inclusion.
   - R5-08 option (recommended: unmeasured in X11a, moved into the X11b PLAN).
   - Wrong-ID generator mode or reworded rationale; `Z_age_arm`; l.161 alignment; discriminating count plants.
   - Check: a scenario-count and compute table recomputed in the response.
4. **Records (R5-14).** Main appends ledger entries, PREREG cites their IDs, the response lists all delegated decisions, and Main records the ADRs.
5. **Main** recomputes all SHA-256s and runs the pinned isolated selftest and offline Go tests once, on synthetic data only. Main records the exercised results and requests export-preparation round 6.
6. After zero issues: run preparation create-once, assemble, request the inventory artifact-binding round, run inventory once, and record its distributions.

### Where the peers changed this memo
- **innovation-challenger:**
  - Found the EROFS/EDQUOT misclassification (merged into R5-02a with my ENOENT/termination finding).
  - Found replacement refs and automatic PGO (new to me; adopted as minor R5-12a/b).
  - Found the stale-cache read in the launcher (extended my self-referential digest finding, R5-12c).
  - Found the partial frontend publish (R5-02d).
  - Found the OC wrapper vector binding (new; adopted as major R5-03, extended with bootstrap and recipe binding).
  - Found the copula typo (R5-10a).
  - I overruled its "fixed" calls on R4-04, R4-05, R4-07, R4-10 and R4-16 with the evidence in §1.
- **experiment-peer:**
  - Independently confirmed the scale defect, which I upgraded to blocking.
  - Supplied the vacuous projection and break-even latencies (R5-06; I had only the in-flight angle).
  - Supplied the duplication bias argument (R5-08), which is stronger than my chunk-efficiency point.
  - Supplied directory-keyed dependence (R5-05; new to me).
  - Supplied the count-plant gap (R5-11) and the audit timing point (merged into R5-07c).
  - Supplied the undocumented budget/look changes (R5-14).
  - Its effective-threshold disclosure point was merged with my intersection–union argument and upgraded to major (R5-04).
  - We agreed independently on the generator (R5-09), Jev-only power (R5-10b) and shared-label accounting (R5-07b).
  - I overruled its "fixed" calls on R4-07, R4-09, R4-10, R4-11 and R4-14.
- **My own additions:**
  - The intersection–union error-control argument.
  - Unbounded begun cases at a stop (R5-07a).
  - Content gating beyond the observation window, and the gitlink contradiction with l.46.
  - Pins that cannot be restored (sandbox-exec/Homebrew) under "no repin".
  - The uncaught `MemoryError`.
  - The verifier accepting self-asserted seals.
  - The ledger-citation gap.
  - The fixture success-path gaps.
  - The `.codegraph` skip count.
