# Council round 2: X11 merge-time association, preregistration rev 2 (part 1) and inventory stage

| Field | Value |
|---|---|
| Study | X11 merge-time association (HD8), required by REQ-176/REQ-177 |
| Reviewed files | `evaluation/x11-merge-association/` `PREREGISTRATION.md` (Revision 2, part 1, DRAFT), `README.md`, `reviews/round-1-response.md`, `inventory.py`, `inventory-inputs.json`, `stages.py`, `selftest.py` |
| SHA-256 (computed by main; the reviewers have no hashing tool) | `PREREGISTRATION.md` eeb542ab5ba8b559681223b04adcec2a2f534acfc633fe7f5f754d6ffe470c1a; `README.md` 18a246d000705d3c359b5b32e581824b9015a61a846711b215f862cf4b7a7a30; `reviews/round-1-response.md` f4e5d19abe1645087aff074e14163df87a5e1904bf8a6b5c89eeec3e2d5e0d6f; `inventory.py` ece1f6389dacfb270c3d4a1944ac95e953a648822f9c71b7bc77c76a69da906b; `inventory-inputs.json` e3ddb9ec50440ed587f0b13deab73c5e1ec34b125a6c8fe0b846d7a9f1bf1454; `stages.py` 8318f49ef283017d784c01e037fb335879165c238651e0c1cc2603ac5e9b458a; `selftest.py` 47e21abdbe72ee6024561822846054c118925fec4ef61662c5f60a86b2f7c326 |
| Digest caveat | Main confirmed after the review that `inventory.py` still hashes to ece1f638…906b and `PREREGISTRATION.md` to eeb542ab…0c1a. |
| Date | 2026-11-08 |
| Reviewer | innovation-council (anthropic/claude-opus-5-5). Independent peers: innovation-challenger (openai-codex/gpt-6-astra:high) and experiment-peer (anthropic/claude-opus-5-5:high). |
| Scope | Whether rev 2 resolves R1-01 to R1-30; new issues; whether the model-free inventory stage may be authorised. Read-only: no file was edited and nothing was run. `inventory.py` and `selftest.py` were not executed, and the candidate clones were only inspected (`django/.git/config`). |
| Line references | `PREREGISTRATION.md` (PREREG), `inventory.py` (INV), `stages.py` (STG), `inventory-inputs.json` (INP), as of this review |

## 1. Round-1 issue status

"Resolved" means the protocol text now meets the round-1 required fix. Code, fixtures and simulations it implies are checked in the implementation round under R1-01.

| ID | Status | Evidence / remaining gap |
|---|---|---|
| R1-01 | **Open** (deferred as planned; disclosed honestly) | PREREG l.3, l.31 and l.178, and README, say that only the inventory and a refusal skeleton exist. Jev limits are fixed at l.57. The driver, GUIDE, protocol manifest, seal writer, fixtures and simulations are still missing. Only misstatement: the "faithful port" claim at l.39 (→ R2-02). |
| R1-02 | Resolved in the protocol | Two-part freeze (l.13–15), inventory-only authority (l.17), maps only after final sign-off (l.22), catalogue-check stage (l.23), all arms forbidden before replay (l.24), branch-label stage (l.28), rule-3 recording (l.17, l.188). The seal contract between the two authorisation scopes is inconsistent (→ R2-15). |
| R1-03 | Partially resolved | Non-overlapping blocks, H=40, mutually exclusive timely/false/censored with a "late" subtype, no credit beyond H (l.128–132). The block interval is degenerate (→ R2-06). Calibration is deferred (expected). |
| R1-04 | Resolved (specification); numbers deferred | Binding .75/.78/.80 scenarios, any-arm support, MC upper bounds, ICC grid, stochastic guard nulls (l.158–176). |
| R1-05 | Partially resolved | Three-zone rule with L=.80, false-support ≤.05 and false not-supported ≤.10 (l.140–166). No accepted failure-to-support rate and no expected support probability (→ R2-10). |
| R1-06 | Resolved | Loss weights 5/1/2/2 plus 0.25 per escalation; thresholds not claimed loss-optimal (l.138). Completeness gaps → R2-17. |
| R1-07 | Resolved (fixtures deferred) | Ordered table with per-metric adequacy (l.144–152). The scope of a label-gate failure is ambiguous (→ R2-09). |
| R1-08 | Resolved | Caps of ≤5% missing gold and ≤10% pr_cap; worst-case P_assign and recall used for support (l.116). Which estimator the bounds use → R2-08. |
| R1-09 | Resolved | Assignment per arm (l.73–86); membership scoring, A∩B, adjudication triggers, sensitivities and chance baseline (l.96). |
| R1-10 | Resolved (implementation deferred) | Blind grouping, ARI, first-new origin, later-gap attribution, new→gap fixture (l.102). |
| R1-11 | Resolved | S/V/A/C terminal table, with pre-fallback Jev-none reported separately (l.71–88). |
| R1-12 | Partially resolved | Audit, A-only/B-only gold, Jev-/Terra-decided splits, adjudicator siding (l.98–100). The gate's semantics and scope are inconsistent (→ R2-09). |
| R1-13 | Resolved | Product run as-is with `PRMetadata=absent`; no token; no PR text for labellers (l.69). Matches `lane.go:109`. |
| R1-14 | Resolved (fixtures deferred) | l.126 matches `guard.go` filters (l.56–62), whole-merge ineligibility (l.84–86) and decided-time/SHA order (l.88–95). Directory guard N/A; Jev-only pending-new rate unmeasured (l.124). |
| R1-15 | Resolved | Descriptive case series of at most 6, coverage-based staleness, precision-only and combined false-fire rates reported separately (l.124, l.130, l.132). |
| R1-16 | Partially resolved | Paired refreshed-minus-no-refresh contrast, map-draw confound named, midpoint always run (l.134). The recovery window and the eligibility clock differ (→ R2-07). |
| R1-17 | Resolved | Fixed 2,000-commit segment, Cold-start disclosure, no-refresh estimand, age buckets (l.45, l.51, l.53). |
| R1-18 | Partially resolved (projection deferred) | Fixed segment, 120 h summed model time / 7-day wall clock, deterministic order (l.45, l.182–184). Order under budget exhaustion → R2-16. Inventory not runnable → R2-01. |
| R1-19 | Partially resolved | Supply gates and the +0.05 margin (l.43). Not implemented in code and "does not discriminate" undefined (→ R2-03). Proxy construct → R2-04. Effectively one candidate → R2-20. |
| R1-20 | Resolved | Wrong-origin propagation endpoint (l.114). |
| R1-21 | Resolved at cell level | Circular moving-block bootstrap fully specified, plus DEFF and missingness-bounded strata (l.118–120). The guard-block interval is not valid (→ R2-06). |
| R1-22 | Resolved | Conclusion conditional on these draws (l.7, l.149). |
| R1-23 | Partially resolved | All 12 `evaluation/*` studies listed (INP:74–86, matching the directory). Inertia-bench still PENDING (INP:87). E1b-v2 disclosed (l.47). Matching gaps → R2-11. |
| R1-24 | Resolved in text; recipe missing | Heuristic scan pinned per commit (l.39, l.65). The scan recipe does not exist (→ R2-01). |
| R1-25 | Resolved | l.57. ≤3 Jev asks per cluster × 10 clusters = 30, and the cap check is `calls >= MaxPerRun` (`lane.go:309`), so the call cap cannot bind. |
| R1-26 | Resolved | Any-arm support across arm-error correlations 0/.5/1 (l.154, l.174). |
| R1-27 | Resolved | Gap/new evidence method specified (l.94). |
| R1-28 | Resolved | Evidence gap disclosed as a threat (l.69, l.186). |
| R1-29 | Resolved | Per-cell operationalisation (l.7); R-063 sentence removed (l.9); removal path handled (l.154). |
| R1-30 | Resolved | Baseline writes its own product-format records, and its own assignments count as owners (l.63). |

Every partially resolved item is tracked by the R2 issue named in its row or by R1-01. None is counted twice.

## 2. New issues

| ID | Severity | Location | Problem | Required fix |
|---|---|---|---|---|
| R2-01 | **blocker** (blocks inventory authorisation) | INP:3, :16–18 (and the four other `heuristic` blocks), :87–93; PREREG l.17, l.41, l.47; STG:45–53; INV:139; response §2 | **The inventory inputs cannot be run, and the export recipes do not exist.** INP status is not `PINNED`. The heuristic exports, the exposure export and the inertia-bench enumeration are all `PENDING`, so INV:139 refuses by design and authorising digest e3ddb9… would authorise nothing. l.17 requires council to name the ownership and exposure **export recipes**, but neither the Clawpatch heuristic-only per-commit scan recipe nor the identifier-only exposure export recipe exists. The clones are `partialclonefilter = blob:none` (`x11-repos/django/.git/config`), and INV:103 reads blob contents with lazy fetch disabled, so the run would fail. No materialisation recipe exists. INP gives no clone URLs, although response §2 claims "origin" pins. The git version is unpinned, and the clones sit in a purgeable macOS temp directory. STG does not require recipe or scanner digests, and the status-only check would accept leftover `PENDING` tokens such as INP:87. | (1) Write the heuristic scan recipe and submit it for review: pinned Clawpatch build and scanner digests, non-agent mappers only, network off, per-commit worktree procedure (handle case collisions on the case-insensitive filesystem, e.g. TypeScript), output schema as in l.41. (2) Write the exposure export recipe: exactly what it reads from each `evaluation/*` study and from inertia-bench, allowed keys only, and every study enumerated, including those with zero intersections. (3) Produce both artifacts. Pin their digests and the recipe and scanner digests in INP, with no `PENDING` token, and set status to `PINNED`. (4) Record canonical clone URLs, the git version and the clone procedure. Either materialise blobs or remove the need for them (R2-02c). Treat the local path as non-normative, with SHA checks governing. (5) Make STG require the recipe and scanner digests, have INV check `scanner_sha256` against the pin, and reject any `PENDING` string. |
| R2-02 | major | PREREG l.39 ("faithful port"); INV:43–45, :97–105, :103, :49, :67; `internal/association/cluster.go:32–40,166`; `git.go:71–109` | **The inventory is not a faithful port of the product front end, and no test compares the two.** (a) Dotfiles: the product `splitName` gives `.gitignore` the stem "", so the dotfiles in a directory form one cluster. INV uses `PurePosixPath.stem` and keeps them separate, which changes cluster IDs and cap order. (b) The product treats type changes (`T`) as Modified, i.e. a coverage gap unless heuristic-owned (`git.go:106`). INV admits `T` as residual. (c) The product detects empty files by comparing blob IDs from `--raw`. INV runs `git show` on content: this needs blobs and is untested on gitlinks [INFERENCE: it would fail there]. (d) Media detection: the product uses the last dot of the whole path; INV uses the filename suffix (edge cases such as a file named `.png`). (e) INV decodes paths strictly as UTF-8, while the product reads bytes with `core.quotepath=off`. | Use the product code itself: a small Go harness calling `Repo.Merges`, `Classify` (no area owners, heuristic set from the export), `BuildClusters` and `Cap`, which is the boring option. Or fix (a)–(e) and add a Go-versus-Python differential fixture covering dotfiles, trailing-dot names, `T`, `C`, empty renames, gitlinks, dotted directories and non-UTF-8 paths. Include the fixture in the authorised digests. Correct l.39 until parity is shown. |
| R2-03 | major | INV:211–216; PREREG l.43, l.45 | **Selection, eligibility and segment rules are not in reviewed code, and several are ambiguous.** INV outputs only envelope totals. `first_parent_commits` includes exposure-excluded commits. There is no cap-excluded share (and its denominator is undefined), no ranking, no discrimination verdict, no segment cut and no in-segment ≥400/≥100 check. It is not stated whether U/N and N≥800 are measured over the envelope or the segment, or what "eligible first-parent commits" means. "The proxy does not discriminate" has no operational test. The consequence of a mandatory repository (Django, TypeScript) failing a gate is "stop for amendment", which leaves the redesign open after supply is known. These decisions would therefore be made by unreviewed arithmetic after the data are seen. | Define each population and statistic exactly. Define discrimination numerically, or state that the +0.05 margin is the whole test. Preregister the consequence of each gate for each mandatory repository, e.g. "HD8 not testable in X11, no substitution". Implement a deterministic selection function in INV that outputs per-gate pass/fail, the ranking, the margin verdict, the segment start and end SHAs and the in-segment counts. Add a fixture for each branch. |
| R2-04 | major | PREREG l.41, l.43; INV:52–62, :173; Clawpatch `src/mappers/go.ts:160,284`, `python.ts:2173,2217` | **Competing hypothesis: U/N may measure test layout and scanner coverage, not likely gold-none share.** The Go mapper separates `_test.go` files (go.ts:160, :284), but Go tests sit next to owned sources, so their directories are owned. Django and TypeScript keep large separate test trees (TypeScript `tests/baselines`), which count toward U if unowned [INFERENCE]. Labellers will often assign test clusters to existing areas, so U inflates the mandatory repositories, and the "exceed both by 0.05" gate may fail, or pass, on layout. `role()` checks only directory segments and misses `_test.go`, `.test.ts`, `test_*.py` and `_test.py` filenames, so the reported role distribution cannot reveal the bias. | Fix `role()` for filename test patterns. Before the inventory runs, preregister whether U/N uses all clusters or non-test clusters only. Require the ranking and the margin to hold on both, and on added-only clusters, and fix now what a disagreement leads to. Report role-stratified U/N for all five repositories. |
| R2-05 | major | STG:45–47; PREREG l.17, l.43–45, l.166, l.182 | **No fixed post-inventory freeze scope (forking paths).** The inventory-scope authorisation does not require the PREREG digest. The preregistration does not list which parameters may change once supply facts are visible. Rule 3 permits distribution checks, but the segment length, L, the adequacy minima (200/100/400), the missingness caps, the block lengths, H and the OC limits could all be retuned to supply without any audit trail. | Add `PREREGISTRATION.md` to the digests required for inventory scope. Add a clause listing exactly what the inventory output sets (the selected repository and segment SHAs by rule; merge-size vectors for simulation; session and time budgets by a stated formula) and freeze everything else from inventory authorisation on. Any other change is a disclosed post-inventory amendment needing a fresh zero-issue round. |
| R2-06 | major | PREREG l.128 | **The guard-block robust interval is degenerate.** A 20-slot block resampled in 5-slot sub-blocks has 4 resampling units. The percentile interval at 20/20 is [1,1]. At 19/20 its lower bound depends on which sub-block holds the error (peer: .85 or .80; Wilson lower bound .764). So whether a block counts as a healthy reference or a fall depends on where errors land. At true .90, P(20/20)=.12 and P(≥19/20)=.39 (binomial, computed here). The share of sequences that establish a healthy reference is neither specified nor reported in calibration. | Replace it with a small-sample-valid interval (e.g. Wilson or Clopper–Pearson with a design effect from the cell-level bootstrap), or use larger blocks (≥40 assignments). Keep the calibration binding, and report the share of sequences that establish a healthy reference in every null and degraded scenario. |
| R2-07 | major | PREREG l.126, l.132, l.134 | **The recovery contrast is on a different clock from its eligibility and may be unmeasurable by construction.** Eligibility is a fall within H=40 **shared guard slots**, but the contrast observes 40 **first-parent commits**. Excluded and non-residual commits use up commits but no slots, so a qualifying fall can come after the contrast window ends. At the supply floor (≥400 clusters per 2,000 commits ≈ 0.2 per commit), 40 commits give about 8 clusters, below the ≥20 labelled A required in each branch [INFERENCE]. The midpoint branch's length and comparison are unspecified. | Define the follow-up in shared slots aligned with H, or require the qualifying fall to lie inside the contrast bounds. Fix a deterministic follow-up rule now, before the inventory, e.g. until ≥20 labelled A in both branches (counts, not correctness) or segment end. Specify the midpoint branch's length and comparator. Update the map and label budgets. |
| R2-08 | major | PREREG l.116, l.118, l.140, l.147, l.149 | **The decision bounds do not name their missingness estimator.** Support uses worst-case **points**. It is unstated whether the robust lower bound is bootstrapped on the labelled-only rate or on the worst-case rate, and whether the priority-2 upper bound uses labelled-only or best-case data. A labelled-only lower bound can clear .80 while the worst-case one does not. | Name the estimator for every decision bound and bootstrap exactly that: the worst-case rate for support lower bounds, and the best-case rate (missing counted correct) for decisive upper bounds, or justify the alternative under the 5% cap. Add near-threshold missing-gold fixtures and use the same estimators in the simulations. |
| R2-09 | major | PREREG l.96, l.98, l.146 | **The label-quality gate's semantics and scope are inconsistent.** l.96 accepts a nonempty A∩B without adjudication, but l.98 gates on an "exact joint-label adjudication rate" defined by exact sorted sets. So A={x,y}, B={x,z} counts as agreed under one reading and as a disagreement under the other. l.98 makes a gate failure cell-level, while priority 1 (l.146) lists label-quality invalidity as global. "Only a new review/fresh sample protocol" leaves an open route to extend the study after labels exist. | Define the gate's numerator and denominator (exact-label disagreement rate versus actual adjudication rate) and keep one routing rule. State that a cell's gate failure makes only that cell's metrics inadequate (priority 3), and list priority-1 conditions exhaustively. State that a gate failure ends that cell as inconclusive within X11, and that any fresh sample is a separate, superseding preregistration. Add fixtures for overlapping-but-unequal sets and for one cell failing while another shows a decisive fall. |
| R2-10 | major | PREREG l.140, l.163, l.166; residual of R1-05 | **No accepted failure-to-support rate and no stated support probability.** Requiring every robust lower bound in 6 cells × 2 metrics to be ≥.80 makes support unlikely even for a good lane. Peer approximation (independent metrics, n=200 each): uniform .90 gives P(support)≈.90 at DEFF 1 and ≈.17 at DEFF 2; .88 gives ≈.30; .85 gives ≈0. "No minimum support power promised" leaves the study, which costs tens of thousands of sessions, likely to end inconclusive even when the lane clears the bar. That undercuts REQ-176's purpose of a signed-off default decision. | State the expected support probabilities in §9, from the binding simulation. Preregister a pre-map rule: if simulated P(support | uniform .90, ICC .10) falls below a stated value, stop for redesign before maps. That value is an **owner decision** to record. Alternatively, record the owner's explicit acceptance that X11 is mainly a falsification test. |
| R2-11 | minor | INV:178–179, :191–198, :210 | Exposure handling has gaps. Matching compares the pre-change blob only for renames, never for ordinary modifications. Other-task exposure is reported only when a record's commit is in the window, so path-only and blob-only records disappear. Record keys are checked only for the repositories being inventoried. The heuristic export schema is not validated beyond presence checks. | Match pre- and post-images of every changed file. Report other-task commit, path and blob matches (without excluding them). Validate every exposure record globally and the full heuristic schema, including key set, types and study membership. Add negative fixtures. |
| R2-12 | minor | INV:175; PREREG l.37, l.45 | The window is selected by committer timestamp, not by chain position after C0. Non-monotone timestamps could put pre-C0 commits in the window or leave post-C0 commits out of it, so the window would not match the replay sequence (C0..tip). | Define the envelope as the first-parent chain after C0, limited by the end time. Assert contiguity, and report any timestamp-order violations. |
| R2-13 | minor | PREREG l.47, l.45 | Merges excluded for association exposure are removed "before segment cutting", but closed-loop replay semantics for them are unstated. Skipping a merge breaks rename and deletion state continuity. | State whether excluded merges are replayed for state but not scored (recommended), or skipped, and how. |
| R2-14 | minor | PREREG l.39 vs INV:173, :211–216 | l.39 promises projections "for both Cold/C0". The code produces one age-independent projection, with U measured against C0 only. | Align the text with the code, or compute both. |
| R2-15 | minor | STG:57–73, :88–91; `selftest.py:81` | The stage and seal contract is inconsistent and the refusal tests are weak. `check_stage` binds every predecessor seal to the *current* authorisation digest, but the inventory runs under the inventory-scope authorisation and is never sealed, so `map` could never verify a "sealed reviewed inventory" (l.22). The `replay` refusal test passes only because the scope is wrong. The missing-seal, wrong-seal and NOT IMPLEMENTED paths are never exercised. `assertGreaterEqual(result['U'], 9)` is not exact. | Specify how the protocol manifest binds the reviewed inventory output and authorisation digests. Add protocol-scope synthetic fixtures for missing, wrong and stale seals and for NOT IMPLEMENTED. Make the U assertion exact. Implementation round. |
| R2-16 | minor | PREREG l.182, l.184 | Under budget exhaustion, the processing order (mandatory repositories first, the Go repository and directory arm last) removes whole cells, most likely the high gold-none repository that REQ-176 requires. The 30,000-session cap (assumption) and "PENDING INVENTORY approved final budgets" overlap without a stated relationship. | Interleave by segment position across cells, or reserve budget per cell. State that the inventory-derived budget replaces or bounds the 30,000 figure, using a stated formula (R2-05). |
| R2-17 | minor | PREREG l.138 | Loss weights are missing for a wrong Terra existing assignment (not silent), a wrong directory assignment and a pending proposal on a gold-gap case. The REQ-102 link to the bar is not stated: 5(1−p) < 1 ⇔ p > .80 is the break-even precision (computed here). | Complete the weights and state the break-even justification. |
| R2-18 | minor | PREREG l.186 | Threat missing: training-data memorisation. The 2024–2026 public history of Django, TypeScript and Prometheus may be known to labellers, Terra and the mapper. That would inflate agreement and Terra-decided P_assign and bias gap/new judgements. | Add it as a threat. Add a descriptive split by commit date against documented model training cutoffs, or state that the cutoffs are unknown. |
| R2-19 | minor | PREREG l.57; `internal/policy/consent.go:26,31`; `lane.go:309–313` | "Product per-run Jev limits" of 30 calls / USD 0.25 are study settings: the product defaults are 2000 calls / USD 0.50. Separately, the product's unknown-cost latch caps every later call in a run after one reply of unknown cost. | Call them study settings and give the cannot-bind arithmetic. Disclose the unknown-cost latch as a possible source of Jev-cap rows. |
| R2-20 | minor | PREREG l.35, l.43, l.47 | gin and cobra are unlikely to reach N≥800 and ≥2,000 first-parent commits in the envelope [INFERENCE: small repositories]. Selection then effectively defaults to Prometheus, which is already exposed to E1b-v2, and the winner's-curse protections become nominal. | State this expectation now, and state whether defaulting to Prometheus is acceptable or which rule applies (R2-03). |

## 3. Inventory-stage authorisation

What the code does:
- **No labels or outcomes.** The exposure export may contain only `studies` and `records` (INV:144–148), and records only identifier keys (INV:178–179; see R2-11).
- **No model calls.** Only `git` is invoked. Lazy fetch and prompts are disabled (INV:26–28).
- **No predictor–outcome association.** Outputs are counts, U/N, Wilson intervals, role and extension distributions, and exposure identifiers (INV:211–216). No gold variable exists.
- **Refusals hold.** Every non-inventory stage refuses (STG:20, :67–73). Output is create-once (STG:88) and deliberately unsealed (STG:91). Rule-3 recording is required (PREREG l.17, l.188).

Why it may not run:
- Its inputs are not runnable, and the recipes that must be reviewed do not exist (R2-01).
- It is not a faithful port (R2-02).
- The rules that consume its outputs are not implemented or fully defined (R2-03, R2-04).
- The parameters frozen at that point are not fixed (R2-05).

Authorising the submitted digests would be either meaningless (INV:139 refuses) or premature.

**When it can be authorised.** A later zero-issue inventory-scope round can authorise it once that round names the SHA-256 of:
- `inventory.py` (or the Go harness), `stages.py`, `selftest.py` and `PREREGISTRATION.md`;
- the PINNED `inventory-inputs.json`;
- both export artifacts and both recipes;
- the scanner and Clawpatch binaries.

After the run, the output digest and its U/N, supply, role, extension and exposure summaries must be recorded in the next round, before any further PREREG change.

## Verdicts

**(a) Inventory-stage authorisation: `not authorised`.** This applies to `inventory.py` ece1f6389dacfb270c3d4a1944ac95e953a648822f9c71b7bc77c76a69da906b, `inventory-inputs.json` e3ddb9ec50440ed587f0b13deab73c5e1ec34b125a6c8fe0b846d7a9f1bf1454 and `stages.py` 8318f49ef283017d784c01e037fb335879165c238651e0c1cc2603ac5e9b458a, reviewed with `PREREGISTRATION.md` eeb542ab5ba8b559681223b04adcec2a2f534acfc633fe7f5f754d6ffe470c1a and `selftest.py` 47e21abdbe72ee6024561822846054c118925fec4ef61662c5f60a86b2f7c326. No real inventory, Clawpatch scan export, map, label or replay may run.

**(b) Overall: `21 issues open`**: R1-01 carried (deferred as planned) and 20 new (1 blocker, 9 major, 10 minor). Not signed off. HD8 stays `proposed`.

## Council synthesis

### 1. Recommendation
Revise rev 2 in place to rev 2 part 2. Fix R2-01 to R2-05 first, because they gate the inventory: the recipes, a parity harness built on the product's Go front end, the selection function, the proxy concordance rule and the post-inventory freeze scope. Then submit for an inventory-scope round. Fix R2-06 to R2-20 in the same revision, so that every rule that uses inventory facts is settled before those facts are seen. The protocol work since round 1 is substantial and should be kept: 23 of 30 round-1 issues are resolved in text, and the deferral of the remaining stages is described honestly.

### 2. Alternatives
- **A. A Go harness on the product front end (recommended for the inventory).** Parity holds by construction, and blob-ID emptiness removes the need to materialise blobs. The cost is a small new Go file under review.
- **B. Keep the Python port and add a differential fixture.** It works, but leaves two implementations to keep in sync.
- **C. Split authorisation further.** Review the recipes first, then the inventory. This adds a round but isolates provenance review. It is reasonable if the recipe review proves heavy.
- **D. Accept a low-power falsification design (R2-10).** This is legitimate only with recorded owner acceptance.

### 3. Evidence
- Inputs pending: INP:3, :16–18, :87–93.
- Partial clone: `x11-repos/django/.git/config` (`partialclonefilter = blob:none`).
- Port gaps: `cluster.go:32–40` (`splitName`), `git.go:94–109` (blob-ID emptiness, `T`→Modified) versus INV:43–45, :97–105.
- Guard ordering checked: `guard.go:56–95`.
- Jev cap: `lane.go:309–313`, `policy/consent.go:26,31`.
- Test separation: Clawpatch `go.ts:160,284`, `python.ts:2173,2217`.
- Arithmetic: binomial P(20/20)=.12 and P(≥19/20)=.39 at .90 (computed here); support probabilities are the peer's normal approximation; break-even 5(1−p)<1 (computed here).

### 4. Risks
- The proxy may not discriminate on non-test clusters, forcing a candidate-list amendment.
- The Go candidate list may collapse to Prometheus.
- A per-commit Clawpatch scan of TypeScript history may be costly or hit filesystem case collisions [INFERENCE].
- The feasibility projection may exceed session caps.
- A low support probability.

Everything is reversible, because nothing real has run.

### 5. Execution outline
1. Write the scan and exposure recipes, materialise or pin the inputs, and set INP to PINNED (R2-01).
2. Replace or verify the inventory front end against the product Go code (R2-02). Add the selection function and its fixtures (R2-03), the role fix and stratified U/N (R2-04), and the window, exclusion and exposure fixes (R2-11 to R2-14).
3. Amend PREREG: post-inventory freeze scope (R2-05), guard interval (R2-06), recovery clock (R2-07), estimators (R2-08), gate (R2-09), support-probability rule after an owner decision (R2-10), and R2-16 to R2-20.
4. Pin PREREG in the inventory scope (R2-05) and fix the seal contract and refusal fixtures (R2-15). Compute SHA-256s and request the inventory-scope round.
5. After authorisation, run the inventory once and record its outputs in the next round. Then implement the remaining stages, the GUIDE and the simulations for the implementation round (R1-01).

### Where the peers changed this memo
- **innovation-challenger:** found the exposure pre-image and other-task reporting gaps and the unvalidated schemas (R2-11), the gate-semantics conflict (R2-09) and the missing estimator for the bounds (R2-08). It independently confirmed the dotfile and `T` divergences I found (R2-02), raised the recovery-clock mismatch I had only partly found (R2-07), and pointed out the missing clone URLs (folded into R2-01).
- **experiment-peer:**
  - Contributed the proxy test-layout confound (R2-04), which I checked against the Clawpatch mappers and kept as major.
  - Raised the PREREG pinning and freeze-scope gap (R2-05).
  - Supplied the support-probability numbers that turned my qualitative power concern into R2-10.
  - Raised the guard sub-block degeneracy (R2-06), the memorisation threat, the processing-order bias, the selftest refusal path and the Jev-limit wording.
  - Rated R2-18 (memorisation) major; I rated it minor, because it is a threat-disclosure fix and the descriptive split is cheap.
- I added the exposure-excluded replay semantics (R2-13), the seal-scope inconsistency (R2-15), the unknown-cost latch (R2-19) and the blob-dependence of the partial clones (R2-01).
