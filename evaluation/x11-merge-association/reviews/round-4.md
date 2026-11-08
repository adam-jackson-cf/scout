# Council round 4: X11a merge-time association accuracy, preregistration rev 4, export-preparation scope

| Field | Value |
|---|---|
| Study | X11a merge-time association accuracy (HD8). It covers the accuracy part of REQ-176/REQ-177. Guard and recovery evidence belongs to X11b (`../x11b-guard-recovery/PLAN.md`, a plan only, not preregistered). |
| Digest manifest | `reviews/rev4-digests.json`, SHA-256 **e9edba67c885623df6900a08aee1cde463e5fa78f828bbb04ca3d8efd7a0324c**. Main computed it, and the manifest excludes itself. The reviewers have no hashing tool. Every digest below is quoted from the manifest and none was recomputed here. Read-only jq checks confirmed that `inventory-inputs.json` `.recipe_sources`, `.export_preparation` and the extraction seal pin the same digests as the manifest. |
| Extraction seal | `exposure-extraction-v4/seal.json`, SHA-256 96fcfcbfada87faaa6624d9124fb3c88d55d714a8cd0bf79ce72a114e32f8510 (schema `x11a-exposure-extraction-seal-v2`) |
| Requested scope | **Export preparation only.** This means model-free Clawpatch heuristic ownership scans at the pinned commits, plus the council-gated frontend build (EXPORT_RECIPES §"Scoped authority" item 1), producing sealed ownership exports. It does not cover inventory or selection, maps, labels or model calls. |
| Date | 2026-11-08 |
| Reviewer | innovation-council (anthropic/claude-opus-5-5) |
| Independent peers | innovation-challenger (openai-codex/gpt-6-astra:high); experiment-peer (anthropic/claude-opus-5-5:high) |
| Scope of review | (1) Status of R1-01 and R3-01…R3-17. (2) New issues in rev 4: the 2,600-per-cell hash-label cap, the recall missingness cap, the pooled-only verdict, the 3,000-hour ceiling and its justification, extraction v4 completeness and the 27-versus-28 correction, scanner/build/runtime closures, and strict schemas. (3) Whether export preparation may be authorised. |
| Method | Read-only. Nobody edited a file or ran a script, test or git command, and nobody touched the candidate clones. The only computations were read-only jq/grep queries over the sealed JSON and the repository files. All arithmetic below was recomputed by hand. |
| Line references | `PREREGISTRATION.md` (PREREG), `prepare_exports.py` (PREP), `inventory.py` (INV), `stages.py` (STG), `pin_tool_closure.py` (PIN), `extract_exposures.py` (EXT), `frontend/main.go` (FE), `frontend/main_test.go` (FET), `selftest.py` (ST), `inventory-inputs.json` (INP), all as of the manifest digests |

## 1. Status of prior items

"Closed" means the round-3 required fix is present in the rev-4 files. Each item was checked against the files, not taken from the response. Remaining gaps are tracked once, as R4 issues.

| ID | Status | Evidence / remaining gap |
|---|---|---|
| R1-01 | **Open (carried, deferred as planned)** | Still absent: the GUIDE, full driver, replay, label, group and scorer code, the seal writer, the planted-effect and planted-null validation and the operating-characteristic (OC) implementation. Only inventory is implemented (STG:11 `IMPLEMENTED = {'inventory'}`), and later stages refuse. PREREG §2, §9 and §11 disclose this. It counts as open until the implementation round. |
| R3-01 | **Closed** | v4 scoped projection: **fastapi/fastapi association 8,866** and tokio-rs/axum association 3,097 (jq group-by over `exposure-identifiers.json`). Scoped records total 50,016. Inventory consumes every unscoped full commit and blob SHA and excludes the empty blob (INV `annotate_exposure`, `validate_unscoped`). There is a positive X2d-SHA fixture (ST:148). |
| R3-02 | **Closed for enumeration and extraction**; reporting residual in R4-17 | All 74 pinned inertia paths are enumerated. There are 36 data files, and the 1,350 scoped records match the seal summary. The repository-level screening identity records are extracted (`additional-identifiers.json`) but never reported by inventory (R4-17). |
| R3-03 | **Superseded by R4-01, R4-03, R4-14, R4-15** | Every generator input is now fixed (PREREG §9 items 1–7). But the false-support event contradicts the power target (R4-01), the frozen simulation is probably not computable (R4-03), and seed, copula and edge-case details remain ambiguous (R4-14, R4-15). |
| R3-04 | **Closed** | The recall-denominator cap U_all/(E_known+U_all) ≤ .05 and the separate precision cap are at PREREG l.120. The wrong .855/.831 recall rationale was removed (l.147). Checked here: .9×.95 = .855, and the R3-04 counterexample gives 855/1,075 = .7953. Consequences follow in R4-12 and R4-13. |
| R3-05 | **Superseded by R4-02, R4-04, R4-05** | Present: kind composition, deterministic thinning (INV `projection_vectors`), the census-replay-plus-hash-label decision, the rate-limit threat and the latency basis. Open: the power-feasibility band (R4-02), wall clock and stop handling (R4-04), and the grouping budget (R4-05). |
| R3-06 | **Closed** | Rule 3 has only per-age adequacy gates (PREREG l.139), rule 4 uses the pooled estimand (l.140), and the .78/.92 fixture passes when pooled (ST:182). OC coverage of the pooled estimand continues in R4-15. |
| R3-07 | **Closed**; location residual in R4-09 | Whole-install manifest: **2,667 files and 109 symlinks** (jq `.files|length`, `.symlinks|length`), all under the tree root. PREP:150–153 recomputes it before `--version`, `init` and `map`. |
| R3-08 | **Closed** | Descriptive high-none difference with block bootstrap, the exact phrase "high-gold-none condition not realised", and no re-selection (PREREG §6 high-none paragraph). |
| R3-09 | **Closed** | `association.Run(..., nil)` parity (FET:151). The `LoadFeatures`/`!IsAgent` differential runs against the Python `heuristic_paths` and covers absent, skipped, agent and duplicate paths (FET:241–278). Normalised directories are emitted by the product (FE:118–129). |
| R3-10 | **Closed**; case nit in R4-14 | The patterns and asymmetries at PREREG §3 match INV `role()`, with fixtures at ST:56–66. INV:88 lowercases paths, which PREREG does not state. |
| R3-11 | **Closed** | Global and system git config are neutralised. There is an independent rename-skip warning check that reports the count and affected merges (INV `check_rename_detection`, `git_env`), with a fixture at ST:168. |
| R3-12 | **Partially fixed → R4-06** | Files are checked for `export-ignore`/`export-subst` (PREP:156–167). Directory entries are not, although `git archive` honours `export-ignore` on directories. |
| R3-13 | **Partially fixed → R4-08** | The rule is at PREREG l.34. The artifact schema and inventory require ownership from all five repositories (INV:288, :476–479), so the candidate-ineligible branch cannot run. |
| R3-14 | **Closed** | The freeze from preparation authority and the enumerated later completions are at PREREG §2. The original preparation authority is re-verified at STG:70. |
| R3-15 | **Partially fixed → R4-11** | Present: the Go recipe, toolchain, module and embed closure (251 packages, 1,463 files, `terra_prompt.txt` included), the pinned Python binary, and the v4 seal's runtime identity and argv. Open: the Python standard library and startup isolation are not pinned, and the Go build checks only the listed files, not the file set it actually selects. |
| R3-16 | **Closed** | Exact envelopes for inputs, artifacts and authorisation (`exact_keys`). FE:171 sets `DisallowUnknownFields`. Unmocked authorisation-chain fixtures are at ST:217 onward. |
| R3-17 | **Closed** | PREREG l.5–6 and l.143 and the PLAN's final paragraph: the REQ-176/177 decision waits for X11b or an amending ADR; refresh recommendations are recorded but never acted on; reuse of X11a data by X11b is declared as prior exposure. |

**Status count:**
- Open: R1-01.
- Closed: R3-01, R3-04, R3-06 to R3-11, R3-14, R3-16, R3-17, and R3-02 for its extraction scope.
- Partially fixed: R3-12, R3-13, R3-15.
- Superseded by R4 issues: R3-03 and R3-05.

**Council correction (27 vs 28).** Round 3 (R3-02) said the sealed listing had "28 JSON/JSONL/CSV data files under `docs/design/screening/`". That was a council miscount.
- jq over `exposure-extraction-v4/inertia-enumeration.json` finds **27** JSON/JSONL/CSV files under `docs/design/screening/`.
- It finds **28** in the whole inertia listing; the 28th is `docs/design/candidates.csv`, outside `screening/`.
- All 28 are among v4's 36 parsed data files. The 36 are the 27 screening JSON/JSONL/CSV files, 8 screening text and checksum files (`SHA256SUMS`, three `SEAL.txt`, two `FREEZE.txt`, `SUPERSEDED.txt` and `slugs.txt`) and `docs/design/candidates.csv`.

Rev 4's "27" is correct. No file was missed. The published breakdown "36 = 27 + 8" leaves out the `candidates.csv` term (R4-17).

## 2. New issues

| ID | Severity | Location | Problem | Required fix |
|---|---|---|---|---|
| R4-01 | **blocking** | PREREG l.149 ("Any-arm family-wise false support ≤ .05 in every defective .75/.78 … scenario"), l.151, l.163 ("Family-wise support counts either adoption arm supporting") | **The defective-scenario criterion contradicts the power target, so the pre-map stop fires whenever the design is adequately powered.** Defects rotate across repository × adoption arm × metric, and everything else is held at P=R=.90. In a rotation that puts the defect on a Jev-only metric, the primary arm sits at .90 everywhere. l.149 requires primary support ≥ .80 there, so "either arm supporting" has probability ≥ .80, against the ≤ .05 limit. The reverse holds for primary-defect rotations if Jev-only is powered. Both peers found this independently, and it is verified here from the text. | Define false support per arm: an arm is falsely supported when it is supported and at least one of its own true metrics is < .80. Family-wise false support is the union over defective arms only. Keep the all-.80 boundary for both arms. Add scenarios where one arm is defective and where both are, plus a deterministic truth-and-verdict fixture. |
| R4-02 | **major** | PREREG l.50 (gate N ≥ 1,800), l.122 (≥ 800 per age), l.149 (every binding scenario, including none share .6 for all repositories), l.161 (C0 retention ≤ 80%), l.176 (Jev .005×4P + 20, ≤ USD60 per repository) | **The binding target is predictably unreachable within the frozen ceilings (R3-05 residual).** At none share .6 (q=.4) and P=R=.90, reaching ≥ 800 known gold-existing and ≥ 800 known assignments after the recall-unknown fraction of up to .05 needs n ≥ 800/(.4×.95) ≈ **2,105** sampled cases per C0 cell. The Jev ceiling .005×4×P_r ≤ 60 caps each repository at **P_r ≤ 3,000** proxy clusters, *including* exposed continuity. C0 thinning keeps 80% of all-added clusters, 50% of renamed clusters without modified members, and 0% of anything else. The target therefore needs .8α + .5β ≥ .70 (α = all-added share, β = renamed-without-modified share). With β = 0 that means α ≳ .88, and a support probability ≥ .80 needs more still. Realistic merge structure is unlikely to meet this [INFERENCE]. The N ≥ 1,800 gate is not the floor that decides feasibility. At C0 the 2,600 cap cannot bind in projection, since .8×3,000 = 2,400. The .005 USD per cluster-run constant is not derived from anything: X2d measured about $0.0009–0.0013 per unit (`x2d-postrun-tuning/axum/README.md:11`, per experiment-peer). | Before preparation freezes PREREG: (1) state the band in §8/§10. (2) Derive the Jev constant from pinned cost-only metadata, as was done for the latency basis. (3) If the band still excludes realistic composition, the owner records one of: (a) acceptance of a probable pre-map not-testable stop; (b) none share .6 binding only for the selected high-none repository, or lower per-age targets with stated consequences; (c) a recomputed Jev ceiling. (4) Align the supply gates with the resulting floor. |
| R4-03 | **major** | PREREG l.151 (10,000 replicates per scenario), l.164–170 (10,000 block replicates per estimator in every replicate; pinned `random.Random` with a fixed sequential draw order), l.172 (an invalid generator run → not testable) | **The frozen OC simulation is probably not computable.** Binding scenarios: 27 none-share configurations × 28 truth settings (4 uniform + 24 defective) × 2 ρ × 3 c × 3 m × 2 mechanisms = **27,216**, plus the boundary scenario and sensitivities. At 10,000 replicates that is about 2.7×10⁸ synthetic studies. Each has up to about 46,800 cases drawn in one fixed sequential MT stream with Beta/Gaussian inverse CDFs at 1e-12 tolerance, plus 10⁴ bootstrap replicates for each of about 24 estimators. Months to years of core time [INFERENCE; not benchmarked]. Under l.172 a run that cannot finish makes X11a not testable. The m=0 cells are also duplicated across both mechanisms. | Preregister a computable design now: fewer within-simulation bootstrap replicates with the Monte Carlo error stated, or an analytic or percentile shortcut proven equivalent; a fixed sequential Monte Carlo stopping rule; removal of the m=0 duplicates; and an exact-equivalent vectorisable stream definition. Add a compute budget and a fixed consequence if it is exceeded. Resolve this together with R4-01 and R4-15, which change the scenario count. |
| R4-04 | **major** | PREREG l.176, l.178, l.180 (labels "ordered by position/cell/case hash"), l.181; `latency-basis-v4.json .planning` | **The wall clock, not the 3,000-hour ceiling, binds, and any resource stop disables every cell at once.** Checked here: (60×149,769 + 31×40,000 + 900×6)/3600 = **2,842.09 h**, under 3,000 h. But 6 concurrent sessions × 24 h × 21 days = **3,024 h**, so the worst case needs **94.0%** round-the-clock utilisation, and raising 1,500 → 3,000 buys no feasibility. Label time of 60 s is an assumption: no X2d label-session latency exists (X2d `sessions.json` records counts only). At 90 s, worst-case label hours alone would be 3,744 h. Labelling runs position-first across all 18 cells, so a wall, hours or rate-limit stop leaves the late segment unlabelled everywhere. Those cases count as unknown. At none share .6 the recall cap tolerates only about 2.1% residual missingness (R4-13), so every cell fails adequacy together. Jev call time and the start of the 21-day clock are not defined. The raise was made before any supply facts, so it was procedurally permissible. | (1) Label each cell in ascending `x11a-label-v4` hash order, round-robin across cells. (2) Preregister that a stop-truncated tail counts as **unsampled**: the realised lowest-hash sample of size k is the sample and π is recomputed. It does not count as missing. (3) Add a wall-day projection with a fixed utilisation factor to the pre-map stop. (4) Define the clock start, and include Jev time or state why it is excluded. Optionally, coordinate samples across arms by dropping the arm from the label hash key; this keeps per-cell validity, raises label reuse and pairing, and lowers load. |
| R4-05 | **major** | PREREG l.109 (grouping takes all deduplicated gold-none cases per repository across ages and arms), l.176 ("plus 9 grouping") | **Nine grouping sessions cannot handle the case volume.** At none share .6 and 2,600 sampled cases per cell, a repository can have up to about 2,600×6×.6 = **9,360** gold-none cases before deduplication, all for a single grouper session. That is not feasible in one context [INFERENCE]. Only 150,000 − 149,769 = **231** spare sessions exist under the frozen ceiling, so the preregistered duplication and new-concern outcomes would be unmeasured, and the fix cannot be made without touching frozen caps. | Preregister a chunking rule now (for example by normalised directory and time, with a maximum number of cases per session), a deterministic cross-chunk linking pass, and a matching reserve within the ceilings. Alternatively, state now that grouping runs on a fixed-size hash subsample, with its outcomes declared sample-bound. |
| R4-06 | **major** | PREP:156–167 (`ls-tree -r --name-only`, without `-t`), PREP:216–219 (`git archive`); EXPORT_RECIPES "requires zero non-unspecified attributes" | **Directory-level `export-ignore` bypasses the archive check (R3-12 residual).** Only file paths are checked. Directory patterns do not recursively match their contents, yet `git archive` skips a directory carrying `export-ignore` (installed Git docs: `gitattributes`, "patterns that match a directory do not recursively match paths inside"; `git-archive`, "Files and directories with the attribute export-ignore won't be added"; challenger citations; verified against the code path here). A pattern such as `/docs export-ignore` gives zero reported attributes and a silently truncated scanned tree, and therefore wrong per-commit ownership and U. Whether any candidate has such a pattern was not checked: no clone was touched. `git archive` also applies working-tree conversions (eol, ident, filter) [INFERENCE: probably immaterial to path ownership]. EXPORT_RECIPES says "zero non-unspecified", but the code also accepts `unset`. | Prefer materialising exact blobs through `ls-tree -r -z` plus `cat-file --batch`, with no attributes or conversions. Otherwise, after extraction, require path, type, mode and blob-hash equality with `ls-tree -r -t` and refuse on any difference. Add a synthetic directory-`export-ignore` fixture that refuses before the scanner runs. Align the wording on `unset`. |
| R4-07 | **major** | PREREG §2 l.23 (each artifact seal contains stage, authority digest, predecessors, exact input and output SHA-256s and argv); PREP:234–239, :266–270, :318–322; STG:88–98 (`verify_seal`: input digests and argv not required, predecessors optional) | **Preparation produces no provenance seal.** The ownership output carries scanner digests, a placeholder argv (`{commit_archive}`, `{fresh_external_state}`), revisions and attribute counts. It records none of these: the export-preparation authorisation digest, the `inventory-inputs.json` digest, PREP's own digest, the actual `sys.argv`, the init/archive/check-attr commands, the runtime identities, the repository's Cold/C0/tip, timing, or a completion status. The build result is equally bare. The artifact manifest (`status: "SEALED"`, INV:281–293) has no generator, so it would be hand-assembled and self-asserted. The inventory round cannot later establish execution provenance that was never recorded. | PREP writes a create-once seal for each operation: stage `export-preparation`, operation, authorisation SHA-256, inputs SHA-256, PREP and imported-module SHA-256s, full argv, the exact per-step command templates, runtime identities, repository pins, revision count, start and end UTC, output path and SHA-256, and a status of completed or refused with a reason. Add a reviewed assembler that builds the artifact manifest from verified seals. Extend `validate_heuristic` and `verify_seal` to match. Failed operations leave explicitly unsealed partial output or none. |
| R4-08 | **major** | PREREG l.34; INV:288 (`exact_keys(value['ownership'], set(repositories))`), INV:476–479; PREP main (a refusal exits 2 with no record) | **The refused-candidate branch cannot run (R3-13 residual).** PREREG says a refused candidate export makes that candidate ineligible. The artifact schema demands an ownership pin for every repository, and inventory loads and validates all of them. One refused candidate therefore blocks the whole inventory, and fixing that would mean changing frozen code after the exports are known. | Define a sealed success-or-refusal record per repository (repository, revision, refusal class, message). Make the manifest accept either one. Implement in INV: candidate refused → ineligible and excluded from ranking; mandatory refused → `not_testable_no_substitution`. Add fixtures for both branches. |
| R4-09 | **major** | INP `.repositories[].path` (`/var/folders/…/T/x11-repos/…`); INP `.export_preparation.tools.node`/`.clawpatch` and `clawpatch-tree-v4.json .root` (all under `/Users/adamjackson/Projects/work/scout/.tools/…`, a different working copy); EXPORT_RECIPES l.26 (reproduction uses `--scanner-root …/scout-x11/.tools/…`); PREREG l.34, l.36 | **Operational failures can trigger the refused-export rule, and the signed inputs bind volatile locations.** The clones sit in macOS per-user temp storage, which the OS purges [INFERENCE: periodic `$TMPDIR` cleanup]. PREREG l.36 calls clone paths non-normative, but PREP and INV use the signed `path`. A purge therefore forces either a changed input digest (a new review) or missing objects, and missing objects make PREP refuse. The scanner and Node pins point into the owner's main `scout` checkout, where unrelated bootstrap work can change bytes. The tree verifier then refuses, which is correct fail-closed behaviour, but under l.34 a refused mandatory export makes X11a not testable. The documented reproduction command targets the `scout-x11` install, so it would produce a manifest with a different root, not the pinned one. PREREG does not separate content-determined refusals from operational ones. | Move the clones to a persistent non-temp root and re-pin the paths only (SHAs unchanged). Pin the x11 worktree's own install, or an immutable copy, and fix the EXPORT_RECIPES command. Preregister that only content-determined refusals fail gates (case or Unicode collision, unsafe symlink, attribute mismatch, unsupported entry, non-UTF-8 path, a scanner error that reproduces on one identical rerun). Operational failures (missing objects, disk or memory exhaustion, interruption, tool-tree drift) write no export, are recorded as deviations, and permit an identical-recipe rerun into fresh output after the pinned bytes are restored by refetching identical objects from the canonical URL. |
| R4-10 | **major** | ST (no PREP fixtures beyond `heuristic_paths`, via FET:275); experiment rule 4; EXPORT_RECIPES claims of "safe extraction checks" | **The preparation path being authorised has no synthetic validation.** No fixture exercises `authorise` for scope `export-preparation`, `extract` (traversal, absolute or escaping symlink, case collision, unsupported entries), the 6N-field `check-attr` parse, scanner-tree drift refusal before `init`/`map`, the end-to-end `ownership` operation, create-once outputs, or `build_frontend` refusals. R4-06 shows that such fixtures would have caught a real defect. | Before re-review, add synthetic fixtures that run the real pinned scanner offline on synthetic repositories only: directory and file `export-ignore` and `export-subst` refusal; each extraction refusal; tree-drift refusal; unsigned-file and wrong-scope refusals; seal contents (R4-07); refusal records (R4-08); and Go build refusal on closure mismatch or an added source file (R4-11). Report the exercised results in the response. |
| R4-11 | minor | PIN:85–89, `python-runtime-v4.json` (executable only); PREP:241–270, PIN:98–132; build recipe `.environment` (real HOME, no `GOCACHE`/`GOTMPDIR`); `__pycache__/prepare_exports.cpython-314.pyc` present in the study folder | **Execution-closure residual (R3-15).** (a) The Python standard library modules the run executes (`tarfile`, `json`, `subprocess`, `hashlib`, `pathlib`) are not pinned, and launch is not isolated. PYTHON* variables, user site and `sitecustomize` apply, and cached bytecode is validated by timestamp, not by hash. (b) The Go build verifies the listed closure files but does not check the file set it actually selects, so a `.go` file added to a package directory would compile unpinned (challenger counterexample) [INFERENCE: low practical risk]. (c) The shared build cache and temp directory are outside the preparation root. | Pin a CPython install manifest (`lib/python3.14/**`, including `lib-dynload`) and verify it before running. Launch with isolation (for example `-I -B` with explicit module paths and a scrubbed environment) and clear stale `__pycache__`. Immediately before the build, rerun `go list -deps -json` in the recipe environment and require exact equality with the recorded packages and files. Use private `GOCACHE`/`GOTMPDIR` under the preparation root. |
| R4-12 | minor | PREREG §9 item 6 (l.166: w=4 on wrong assignments and abstentions) vs §8 l.147 | **The binding "preferential" missingness mechanism is the least adverse direction for support power.** The worst-case estimators already count errors as wrong, so removing errors first raises worst-case support: at the cap, error-preferential gives about .865, uniform .855 and correct-preferential .851 (experiment-peer; consistent with the algebra here). l.147 itself says correct-preferential "can do worse", but it is not binding. | Add correct-preferential removal (w=4 on correct assignments and correct existing cases) as a binding mechanism alongside uniform, or stop describing the binding set as MNAR stress. |
| R4-13 | minor | PREREG l.99, l.120, l.181 | **The implied residual-missingness tolerance is not stated, and there is no rule for replacing an invalid session.** Under uniform missingness the recall cap means a residual missing share m ≤ q/(19+q): **2.06%** at none share .6, 3.06% at .4 and 4.04% at .2. One invalid A or B session can make a case unrecoverable. The 3-session reserve funds a replacement, but no rule uses it. | State the tolerance. Preregister one fresh same-model replacement for an invalid A or B session, within the cell reserve, before a case is declared unmeasured. |
| R4-14 | minor | PREREG l.164, l.169; INV:88 vs PREREG §3 role patterns | **Remaining specification ambiguities (R3-03 residual):** the scenario-index order (q-vector and truth/rotation encoding) does not fix the seeds uniquely; sequential age→arm draws conflict with a Kronecker copula (it needs to say "iid normals in loop order, then Cholesky"); `random()` can return 0, where Φ⁻¹(0) is undefined; the bootstrap seed formula's `%2^32` is ambiguous; the bootstrap stream is reused across all scenarios and replicates, which is undisclosed; the missingness race runs per cell while labels are shared across arms; and `role()` lowercases paths, which the text does not say. | Specify each of these exactly in PREREG, and add a seed-index fixture. |
| R4-15 | minor | PREREG l.164 (cross-age copula fixed at .5), l.163 (homogeneous boundary only) | **OC coverage of the pooled-only verdict is thin.** The same lane judges overlapping clusters at both ages, so the cross-age correlation may approach 1, where pooled variance scales by (1+ρ_age)/2, i.e. 1.33× the .5 value. No boundary scenario splits truth by age (for example Cold .72 and C0 .88, pooled .80), which is the case the pooled estimand could mask. | Make ρ_age ∈ {.5, 1} binding, and add an age-heterogeneous boundary false-support scenario. Coordinate with R4-03. |
| R4-16 | minor | PREREG l.157 | **The planted values (.95, .70, .50, .60) are too gross to catch small scoring bugs.** A defect that shifts estimates by a few points would still pass. | Require exact equality of every numerator and denominator (worst-case and best-case) against the planted counts in each cell, plus deterministic near-boundary plants at .79 and .81. |
| R4-17 | minor | EXT `additional-identifiers.json` (records with a repository but no commit, blob or path identifier: django/django, gin-gonic/gin, microsoft/TypeScript, prometheus/prometheus in `inertia-bench/docs/design/screening/{t2-r2,t2-r3,archive}`, plus django and prometheus in `change-cost-v2`; jq verified); INV (never reads that file); EXPORT_RECIPES l.70; `exposure-identifiers.json` | **Exposure-reporting residuals (R3-02).** (a) Repository-level other-task screening exposure is extracted but never reported, so inventory can show zero resolved records for screening studies that did examine these repositories. (b) The breakdown should read 36 = 27 + 8 + 1 (`docs/design/candidates.csv`). (c) The study filter behind "1,350 scoped inertia records" is not named. (d) A mis-scoped record, `{"path":"human-confirmation-51.csv","repository":"microsoft/TypeScript"}`, shows that the path allowlist can attribute paths to the wrong repository. It is harmless, because it is other-task and reported only. | Add a descriptive inventory field for repository-only exposure by study. It must not drive score exclusion or claim commit overlap. Correct the wording, name the filter, and disclose the mis-scoped record. |

**Totals:** 17 new issues (1 blocking, 9 major, 7 minor) plus R1-01 carried.

Considered and not raised:
- **Nested frontend row schemas (challenger P-08).** The rows come from a pinned binary built from reviewed code, and R3-16 asked for exact envelopes, which now exist.
- **A write allowlist (challenger P-04).** No preregistered promise of confinement exists. The private-cache part of that point is in R4-11.
- **Dropping the arm from the label hash (experiment-peer P-08).** This is an optimisation, not a defect, and is offered as an option in R4-04.

## 3. Review of the specific rev-4 elements

- **2,600 per cell hash-label cap (PREREG l.97).** The sample does not depend on outputs or gold. π is constant within each cell, and ratio estimators with commit-block resampling are valid. Sessions per cell: 3×2,600 + ⌈.2×2,600⌉ = **8,320**; 18×8,320 + 9 = **149,769** ≤ 150,000. The defects are in its interaction with supply and resources (R4-02, R4-04, R4-05), not in the cap itself.
- **Recall missingness cap (PREREG l.120).** The definition is correct, and worst-case recall at the cap with known recall .90 is .855. Consequences: R4-12 and R4-13.
- **Pooled-only verdict (PREREG l.133–143).** Coherent. Rules 2–4 use the equal-age pooled estimand, the per-age gates are adequacy-only, and the .78/.92 fixture exists. REQ-177 is a "starting reference only", so it permits this. Masking risk is disclosed (l.183). Simulation coverage: R4-15.
- **3,000-hour ceiling.** The arithmetic is correct (2,842.09 h, 5.3% headroom). The latency basis is honestly labelled as an allowance, not measured label latency. Fallback 31 = ⌈3×10.3⌉, and the label figure of 60 exceeds 5×10.3 = 51.5. The JSON means check: 111.7/25 = 4.468 and 186.3/32 = 5.822. But the ceiling is effectively the 3,024-hour wall-concurrency limit (R4-04).
- **Extraction v4.** jq totals:
  - 24,001 parsed, 321 non-data and 3 unclassified (sum 24,325).
  - Unscoped records: 177,483 blobs + 61,135 commits + 23,462 paths = 262,080.
  - All 74 inertia paths are accounted for.
  - 27 is correct (§1).
  - The owner-directed extraction ran outside council authority. It read only pinned Scout and inertia blobs, read no candidate outcome and computed no association. It is accepted as review evidence and supersedes v3 for binding, subject to R4-17.
- **Scanner, build and runtime closures.** The scanner tree is sound (R3-07 closed), apart from where it lives (R4-09). Go: R4-11(b). Python: R4-11(a).
- **Strict schemas.** Input, artifact and authorisation envelopes are exact, and FE rejects unknown fields. Preparation outputs lack seal fields (R4-07), and the refusal union is missing (R4-08).
- **Model-free and no outcome leakage.** Confirmed:
  - The scanner runs `map --source heuristic` with fresh state and HOME.
  - PATH holds only the pinned git, and model CLIs are absent.
  - `sandbox-exec` denies network access.
  - The pinned scanner sets provider null and returns before agent mapping (challenger: `dist/app.js:59`, `dist/agent-mapper.js:76–79`).
  - PREP computes no selection, gold or predictor–outcome relation.

## 4. Export-preparation authorisation

What the reviewed code does well:
- Fail-closed authorisation with exact digest coverage (STG:36–85, PREP:56–74).
- Whole-scanner-tree verification before every scanner call.
- Per-commit fresh state.
- Network denial and model-free execution.
- Create-once outputs.
- A hermetic Go build recipe with a mandatory buildinfo artifact.

Why it may not run now:
1. **The preparation authority freezes PREREG and every preparation and inventory byte** (PREREG §2, STG:12–14, STG:70). R4-01 to R4-05 and R4-12 to R4-16 are PREREG defects. Fixing them after ownership exports exist would need a new preparation round with exposure disclosed. R4-01 alone makes the study untestable by construction.
2. **The recipe itself has correctness and provenance defects:** silent tree truncation (R4-06), no seal (R4-07), an unimplementable refusal branch in frozen INV (R4-08), refusal triggered by operational failure and volatile pinned locations (R4-09), and no synthetic validation of the authorised path (R4-10).

## Verdicts

**(a) Export-preparation authorisation: `not authorised`.** This covers:
- `PREREGISTRATION.md` 974e15dee5412c4aa7cb5d541fab6759290713c7e26b52996de88b738a71f7e2
- `EXPORT_RECIPES.md` 198fae6c8812bb5ece2b78c11f08f75680ca3ce8da37fd074c8b1673ef5c4890
- `README.md` 5c443200d85959bcea1daee468a06bbd847fbf8657a0ce20b361cd5e035bd0bb
- `prepare_exports.py` a29b75fa76de8c71f269ba776bd02f4cb5704470169a5d8809026b73fc06e124
- `pin_tool_closure.py` 3792fcc23a55ab10b6474a0c83a8efabaf59f479af3ad3a8dc445a2fe8250cb3
- `stages.py` 7647cd9c77cc0e6d5b273325ea319d44a63101547e7ebe46164fa464bb7e8496
- `inventory.py` e8b1e382e6cc8ff2bc21eda14c59c2fdf6d0f9758ccd29f892aef23329062022
- `selftest.py` 8447eab07a20cefc6f49f37bcb7d249502a638624643182eb44e5b2719a27d18
- `extract_exposures.py` 966f7ed55d0a30c3d5cc4e2deea5c78b912f98827e261bbdb9b265c44749c6b8
- `frontend/main.go` bd83e1c3a5c4c68a6daee2fa2b592e9a83fca889e6ddfaba9336c079f0b965c0
- `frontend/main_test.go` 39a6c86e03eed976befd6ce04ae686b2e8e237b554e0fa5da1d4c9e9ff8d0b2d
- `inventory-inputs.json` fe137bed1c5720250988d7895974b56a05cdd5f8c3ec8fc5f093d8f3f3ec1ed8
- `exposure-source-index.json` fd490c0545b8684abff2d462b47a97b7e7d4074c90f583556539c9dddb0ac331
- `exposure-contexts-v4.json` 2dd5b6b7d6665d64519585acdb63ed6eafadc3092bda25a38433616f896199fb
- `latency-basis-v4.json` bdc3cef935d65a20b4b0b9f34bb68c881d29da4960e74e5ee5b60549f0894104
- `tool-pins-v4/clawpatch-tree-v4.json` a0121b5d6a7bf6e27b341cac6328daf91e41d7c129e00c9d16acf2a10f1a8b90
- `tool-pins-v4/frontend-build-recipe-v4.json` 12692e0d0e8d3b04f4af78c94361693c9a5cb6a3d086acc2d5fb644c21eb5d30
- `tool-pins-v4/frontend-go-closure-v4.json` f5f7b82f81796f923c030ed91d6352e8d4000d53fb4c8a2316e1e7b9c79533d1
- `tool-pins-v4/python-runtime-v4.json` c8ce8516ba0c1ec0f22ff5b345931be636cfb01040a2fae2be2067a466e6be79
- `exposure-extraction-v4/seal.json` 96fcfcbfada87faaa6624d9124fb3c88d55d714a8cd0bf79ce72a114e32f8510
- the five v4 extraction outputs as pinned in that seal and the manifest
- `../x11b-guard-recovery/PLAN.md` d3a5c666d50e1e6087aeb43896b7f5ad663911b88e3ccde0d885fa6f47791609

All are as listed in manifest e9edba67c885623df6900a08aee1cde463e5fa78f828bbb04ca3d8efd7a0324c. No ownership scan, frontend build, inventory, map, label, replay or model call may run. Inventory also still needs its own artifact-binding round after a future authorised preparation.

**(b) Overall: `18 issues open`.** That is R1-01 carried (deferred implementation; open until the implementation round) plus 17 new (1 blocking, 9 major, 7 minor). Not signed off. HD8 stays `proposed`. The ledgers, design and product are unchanged.

## Council synthesis

### 1. Recommendation
Revise to rev 5 and request a second export-preparation round. Fix all PREREG defects before the freeze, and fix the preparation recipe defects at the same time.

1. **R4-01 (blocking):** redefine family-wise false support per defective arm.
2. **Make the binding programme feasible** (R4-02, R4-03, R4-15). Derive the Jev constant from pinned cost metadata. Make the OC computable, with a stated compute budget. **Escalation:** if the band still excludes realistic composition, the owner must choose before the freeze among R4-02 options (a), (b) and (c). The council does not choose a scientific target on the owner's behalf.
3. **Make resource stops degrade gracefully** (R4-04, R4-05, R4-13): hash-order labelling, a truncated tail treated as unsampled, a wall-day projection, a grouping chunking rule and a reserve.
4. **Fix the recipe:** cat-file materialisation (R4-06), seals and a manifest assembler (R4-07), the refusal union in PREP and INV (R4-08), persistent clone and tool locations plus the operational-versus-content refusal rule (R4-09), Python and Go closure enforcement (R4-11), and the preparation fixtures (R4-10).
5. **Clear the minors:** R4-12, R4-14, R4-16, R4-17.

Keep the rev-4 core: the Go frontend, selection, estimators, equal-age pooling, the label gate, the census-replay plus hash-label design, extraction v4 and the X11a/X11b split are sound, and 12 R3 items are closed (R3-02 for its extraction scope).

### 2. Alternatives
- **A. Recommended: fix everything, then one preparation round.** Nothing is tuned after supply-adjacent facts exist. It costs one more round.
- **B. Move the freeze from preparation authority to inventory-binding authority** (a PREREG §2 amendment). Preparation could then run once R4-06 to R4-11 are fixed, while the PREREG fixes continue. The drawback: ownership exports plus the clones let anyone derive supply facts that bear directly on R4-02-type choices. A rule that the exports stay unopened cannot be verified. Not recommended unless the owner records the exposure and accepts it.
- **C. Owner accepts a probable not-testable stop** (R4-02 option a, with R4-04 left as is). Honest, but likely to waste the preparation and implementation effort.
- **D. Narrow the binding grid:** none share .6 binding only for the selected high-none repository, with mandatory repositories at {.2, .4} and .6 as sensitivity. This widens the feasible band and keeps REQ-177's high-none condition explicit. It is a design choice for the owner (part of R4-02 option b).

### 3. Evidence
- **False-support contradiction:** PREREG l.149, l.151 and l.163, quoted in R4-01.
- **Feasibility arithmetic:**
  - Jev ceiling: .005×4×P_r ≤ 60 gives P_r ≤ 3,000 (l.176). C0 retention ≤ .80 (l.161).
  - Floor: n ≥ 800/(.4×.95) = 2,105. The requirement .8α + .5β ≥ .702 gives α ≥ .877 when β = 0.
  - Grouping headroom: 150,000 − 149,769 = 231.
  - Hours: (60×149,769 + 31×40,000 + 900×6)/3600 = 2,842.09.
  - Wall: 6×24×21 = 3,024 → 94.0%.
  - Residual tolerance m ≤ q/(19+q) = .0206 at q = .4.
- **OC scale:** 27×28×2×3×3×2 = 27,216 binding scenarios; ×10⁴ replicates; ×10⁴ block replicates per estimator (l.151, l.170).
- **Archive check:** PREP:157 (`ls-tree -r --name-only` without `-t`); PREP:216–219 `git archive`; installed Git documentation per the challenger.
- **Seals:** PREP:234–239 (placeholder argv; no authority or input digest); STG:88–98; no assembler for `status: "SEALED"` (only ST:260 builds one).
- **Refusal branch:** INV:288, INV:476–479; PREREG l.34.
- **Locations:** INP repository paths under `/var/folders/g5/…/T/x11-repos/`; tools and tree root under `/Users/adamjackson/Projects/work/scout/.tools/…`; EXPORT_RECIPES l.26 points at the `scout-x11` install.
- **Fixtures:** no PREP coverage in ST; FET:275 covers only `heuristic_paths`.
- **Extraction (jq):**
  - fastapi association 8,866; axum 3,097.
  - Screening JSON/JSONL/CSV 27; repository-wide 28 (`docs/design/candidates.csv`); 36 data files.
  - Repository-identity records for all four screened candidates.
  - Mis-scoped `human-confirmation-51.csv` record under TypeScript.
  - gin's pinned tip `0f09c3a9…` appears in inertia screening as other-task exposure.
- **X2d session metadata:** `data/rounds/0/regate/rounds/0/sessions.json` has the keys cap, ceiling, formula, plan, round and used, with no durations. X2d labelling unmeasured gold was 0 of 120, with 115 agreed and 5 adjudicated (X2d RESULTS l.315), which makes a residual tolerance of about 2% plausible but untested at scale.

### 4. Risks
- Even after the fixes, real thinned supply may fall below the floor, giving a pre-map not-testable outcome. That is acceptable if the owner chose it knowingly.
- Subscription throughput across about 10⁵ sessions in 21 days is untested. Hash-order labelling limits the damage to smaller samples but does not remove it.
- A candidate tree may contain content that is legitimately refused (absolute symlinks, odd filenames). Only a content refusal of a mandatory repository should end X11a.
- Directory `export-ignore` in a candidate is unknown until a safe materialisation runs.
- The OC compute redesign may need its own review iteration.

Reversibility: everything is reversible. Only the owner-directed identifier extraction has run, and its v3 and v4 artifacts are kept.

### 5. Execution outline
1. **PREREG.** Apply the R4-01 per-arm definition, the R4-02 band and Jev derivation plus the owner's choice, a computable R4-03 design with a budget, R4-04 hash-order labelling and the unsampled-tail rule with a wall projection, R4-05 grouping chunking and reserve, the R4-12 binding correct-preferential mechanism, the R4-13 tolerance and replacement rule, the R4-14 seed and copula details, the R4-15 ρ_age grid and heterogeneous boundary, and R4-16 exact plants.
   - Check: a scenario-index table and a compute-budget statement exist; recompute the totals.
2. **PREP and INV.**
   - Cat-file materialisation, or a post-extraction equality check (R4-06).
   - A per-operation seal plus the manifest assembler (R4-07).
   - A success-or-refusal union and the INV eligibility branches (R4-08).
   - Content-versus-operational refusal classes (R4-09).
   - Python isolation and stdlib manifest; Go `list -deps` equality and private caches (R4-11).
   - The repository-only exposure field (R4-17).
   - Check: all R4-10 fixtures pass on synthetic repositories with the real pinned scanner offline, and none touches a candidate clone.
3. **Locations.** Re-clone to a persistent root (same SHAs), pin tools from an immutable or x11-local install, regenerate `tool-pins-*` with the corrected command, and update INP.
   - Check: the regenerated tree root equals the pinned tool parent, and the jq SHA cross-check passes.
4. **Docs.** Fix the EXPORT_RECIPES wording (36 = 27 + 8 + 1, `unset`, the filter name, the mis-scoped record).
5. **Main** recomputes all SHA-256s, runs the pinned Python selftest and the offline Go tests once, records the exercised results, and requests export-preparation round 5.
6. After zero issues: run preparation create-once, seal it, request the artifact-binding inventory round, run inventory once, and record its distributions before any further PREREG change.

### Where the peers changed this memo
- **innovation-challenger:**
  - Found the false-support contradiction (R4-01), which I had missed.
  - Found the directory `export-ignore` bypass (R4-06). I had flagged archive content conversions but accepted the file-level check. I verified their reading against the code path and adopted it, overruling experiment-peer's "R3-12 fixed".
  - Found the Go set-membership hole (part of R4-11).
  - Found the repository-only screening exposure gap (R4-17), which I verified with jq.
  - Independently matched my own findings on seals (R4-07) and the refusal branch (R4-08).
  - I did not raise the write-confinement (P-04) and nested-schema (P-08) points as separate issues; the reasons are given in §2.
- **experiment-peer:**
  - Independently found R4-01.
  - Supplied the Jev-ceiling × thinning feasibility band (R4-02). I had seen that the 2,105 floor exceeds the 1,800 gate, but not the P_r ≤ 3,000 interaction.
  - Found OC computability (R4-03) and the grouping volume (R4-05), both of which I had missed.
  - Supplied the hash-order and unsampled-tail fix for the wall-clock stop. I had found the 94% utilisation and the simultaneous truncation.
  - Raised R4-13 replacement, R4-14 to R4-16 and the extraction wording.
  - We agreed independently on the missingness direction (R4-12).
  - I overrode the peer's "fixed" ratings for R3-12, R3-13 and R3-15 with the code evidence above, and folded the peer's arm-hash point into R4-04 as an option.
- **My own additions:**
  - Volatile clone and tool locations, and the operational-versus-content refusal rule (R4-09).
  - No synthetic validation of the authorised path (R4-10).
  - The stale `__pycache__` and the Python stdlib (with the challenger).
  - The 27/28 verification and council correction.
  - The hours-versus-wall arithmetic.
  - The repository-identity and mis-scoped record checks.
  - The archive conversion note in R4-06.
