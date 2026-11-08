# Council round 6: X11a merge-time association accuracy, preregistration rev 6b, export-preparation scope

| Field | Value |
|---|---|
| Study | X11a merge-time association accuracy (HD8). This is the accuracy part of REQ-176/REQ-177. Guard and recovery evidence belongs to X11b (`../x11b-guard-recovery/PLAN.md`, a plan only, not preregistered). |
| Digest manifest | `reviews/rev6b-digests.json`, 1,573 entries, self excluded. Main computed its SHA-256 as **7e873019ac128958347d2e6cb80b148481eedc026050fb3c6abca93b2d2b1acd**. The council has no hashing tool. Every digest below is copied from the manifest and none was recomputed. |
| Requested scope | **Export preparation only.** That is the model-free Clawpatch heuristic ownership scans at the pinned commits, the council-gated frontend build and the exposure-binding operation, producing sealed ownership exports. Inventory, selection, OC runs, maps, labels, replay and model calls are out of scope. ADR-013 is reviewed in the same packet. |
| Owner decisions in force | R-073: "(b) High-none only on selected repo" (alternative D). R-064 standing delegation: wherever an owner choice arises, the recommended option is stated as the required fix. |
| Independent check (Main) | `/var/folders/g5/r809r2p97nz9gtj25s1c37t80000gn/T/x11-checks-rev6b-w6bong_v/check-report.txt`. Preparation suite: 22 tests OK. Integrated suite: 27 tests OK, including the imported `BootstrapTrustTests` and `OCContractTests`. Pinned PGO-off frontend Go tests: `ok`. All 1,573 digests matched before and after the checks. Launcher pins were `INPUTS_SHA256` e492da74… and `BOOTSTRAP_SHA256` 5b8735ad…, both equal to the manifest entries. The failed rev-6 run and its root causes are preserved in `reviews/rev6b-checks.json`. The council records these as Main's exercised evidence and did not re-execute them. |
| Date | 2026-11-08 |
| Reviewer | innovation-council (anthropic/claude-opus-5-5) |
| Independent peers | innovation-challenger (openai-codex/gpt-6-astra:high); experiment-peer (anthropic/claude-opus-5-5:high). Neither was told the lead's view. |
| Method | Read-only. Nobody edited a file or ran a script, test, build or git command, and nobody touched the candidate clones. The only computations were jq/grep over study files and hand arithmetic. Two bibliographic facts were checked by web search. |
| Line references | `PREREGISTRATION.md` (PREREG), `prepare_exports.py` (PREP), `inventory.py` (INV), `oc_v5.go` (OCGO), `oc_v5.py` (OCPY), `prep_selftest.py` (PST), `oc_selftest_v6.py` (OCST), `prep_scale_v6b.py` (SCALE), `tool-pins-v6/persistent-locations-v6b.json` (LOC), `tool-pins-v6/object-restoration-v6.json` (RESTORE), ADR-013 `ADR.md` (ADR), `docs/research-ledger.md` (LEDGER), `reviews/round-5-response.md` (RESP). All refer to the bytes pinned in the manifest. |

## 1. Status of prior items

"Closed" means the round-5 required fix is present in the rev-6b files. Each status was checked against the files, not taken from the response. Residual gaps appear only once, as R6 issues.

| ID | Status | Evidence / remaining gap |
|---|---|---|
| R1-01 | **Open (carried; deferred as planned)** | These are still absent: full GUIDE, real orchestration, calibration, replay, labels, scorer, stage seals, whole-pipeline plants and inventory-conditioned OC tables. PREREG §2, §9 and §11 disclose this, and later stages refuse. It stays open until the implementation round. |
| R5-01 | **Closed** (residuals in R6-03 and R6-05) | Revisions are bounded to Cold, C0 and the first 2,000 post-C0 commits (PREP:401–403). Each revision gets one `ls-tree -r -z` and one disk-backed `cat-file --batch` (PREP:192–242), a compact trace and a per-revision disk record (PREP:403–431). Ownership is streamed to the frontend as NDJSON, one record at a time (INV:651–658; frontend `main.go` `next`). `BaseException` failures are sealed (PREP:765). Node heap is 8,192 MiB. A populated synthetic benchmark exists: 80,001 files × 20 revisions, with time and RSS reported (`prep-scale-v6b.json`). |
| R5-02 | **Partially fixed → R6-03** | Scanner errors are always operational, and only structural causes count as content refusals (PREP:50–58). Attempts are capped at three, with terminal `PreparationInfeasibleOperational` (PREP:742–779, :548). The immutable Git root, OS build and sandbox binding are in place. Gitlinks are empty directories and alias checks are sorted. **Termination is still not guaranteed:** `subprocess.run` has no timeout (PREP:153), so a hung attempt never reaches the terminal rule. Operator interrupts also consume attempts (R6-03). |
| R5-03 | **Closed** | Protocol scope requires `reviewed_inventory` (STG:48–55). OC checks the original inventory authority, export manifest, sole sealed output and exact non-null vectors (OCPY:99–137). Bootstrap is fixed at 399 and timing overrides refuse (OCPY:156–161). Refusal plants: OCST:76–130. |
| R5-04 | **Closed** | Intersection–union gate with one-sided Clopper–Pearson at α = .05/3 per look (PREREG l.189–197; OCGO:234–235, 289–297). Peer exact-binomial checks: P(Bin(256,.05) ≤ 5) = .0107 and P(Bin(512,.05) ≤ 15) = .0150, both ≤ 1/60, with 6 and 16 failing; 36/1024 passes and 37 fails. 219/256 and 847/1024 hold with margin. 429/512 is within hand-arithmetic error, but the code computes it exactly and OCST:147–155 pins it together with its neighbours. |
| R5-05 | **Closed** (wording residual in R6-07) | Per-case directory hash and largest share (INV:242–264). Paired commit and directory bootstraps combined as min lower / max upper, each needing ≥ 95% finite draws (PREREG l.136; OCGO:190–225). Directory ICC {0, .10} is crossed through the grid: 56,160 scenarios, recomputed. |
| R5-06 | **Closed as specified; defect in the gate composition → R6-04** | The 180-session synthetic-GUIDE calibration, λ = 1.25 × the largest role mean, measured Q, the finishing allowance and the H/d/k gates are frozen (PREREG l.213–225). The RESP calibration table (l.47–51) reproduces exactly. Before this fix, the static projection was vacuous. Now the full-L H/d gates make the k gate dead (R6-04). |
| R5-07 | **Closed** | At most six cases in flight; requests are sequential within a case; six credits are prepaid (l.111). Frontiers are per cell with charge-once pools (l.108, l.209). Audit inclusion is decided at case start (l.119). Finishing reserve 36τ (l.217). |
| R5-08 | **Closed** | Duplication is declared unmeasured, not zero (PREREG l.5, l.121). The REQ-176 duplication requirement stays outstanding in the X11b PLAN (l.13–19). |
| R5-09 | **Closed** | ".72/.88 infeasible under the frozen abstain-only generator" (l.165). |
| R5-10 | **Closed** | `Z_age_arm` (l.179). Both arms are required for the uniform-.90 pass (l.167; OCGO:292–293). |
| R5-11 | **Closed** | The discriminating plant is C 70, A_known 110, E_known 90, A_unknown 7, U_unassigned 13. It gives 70/117, 70/110, 77/117 and 77/97, all hand-verified (`oc-v5-fixtures.json` `count_plants`; OCST:157–161). |
| R5-12 | **Closed** | The environment sets no-replace and no-lazy-fetch, and repositories with grafts, shallow history, replacement refs or promisor packs are refused (PREP:139–145, 179–189). PGO is off in both recipes (`pin_tool_closure.py`:176–178). The stdlib-first bootstrap rejects caches before importing the helper (`bootstrap_v6.py`:16–27), and Main's launcher exercised it. Fresh ordinary full clones exist (RESTORE). |
| R5-13 | **Closed** (refusal-seal residual in R6-06) | Exact per-attempt preflight, plus the tree → batch → materialise → init → map chain with digest, count and argv checks (PREP:526–580). Build and exposure trace checks are present. Genuine execute-produced build, exposure, content and terminal-operational seals pass through ASM → artifact bindings → inventory (PST:27–77, 417–467, 480–497). Empty or tampered traces refuse. |
| R5-14 | **Partially fixed → R6-10, R6-11** | R-073–R-076 and R-079 exist and PREREG cites them (l.11), and ADR-013 records the three omitted delegated decisions (ADR rows "Old 48 coreh…", "Maximum cumulative look…", "Maintain 21 days…"). However, R-075 and R-079 do not record the evidence they are cited for, and the ADR has form defects. |

**Status count:**
- Open: R1-01.
- Partially fixed: R5-02, R5-14.
- Closed: R5-01, R5-03 to R5-13.

## 2. New issues

| ID | Severity | Location | Problem | Required fix (recommended option under R-064) |
|---|---|---|---|---|
| **R6-01** | **blocking** | LOC:10223–10224 (TypeScript `segment_commits: 1666`, `segment_end` = pinned tip `6161772c…`). Same values in `persistent-locations-v6.json`:174–175. PREREG l.46 ("an envelope < 2000 fails inventory's gate"), l.54, l.60, l.41 ("If mandatory repository … fail any required gate, X11a is not testable"). INV:538 (`'segment_length': r['segment_commits'] == 2000`). ST:43–45 (1,999 → `not_testable_no_substitution`). | **Packet evidence already shows that mandatory microsoft/TypeScript fails the fixed segment gate.** The active, signed-for-review metadata records only **1,666** first-parent commits after C0 through the pinned tip. That count uses the same `chain[:2000]` semantics as INV:685. Under the frozen rules, inventory would return `not_testable_no_substitution` whatever the ownership exports contain. Authorising roughly 85 h per repository of preparation (R6-03) would therefore buy nothing. Worse, it would freeze, at preparation authority (PREREG l.32), a preregistration that is known to be untestable. The same metadata shows gin at 316 and cobra at 70, so **Prometheus is the only Go candidate that can pass**. PREREG l.41 anticipates this only as an [INFERENCE]. Both peers' readings agree. The challenger found the TypeScript count independently. | **Amend the segment rule before the freeze, and disclose that the amendment was informed by eligibility data.** Nothing is frozen yet. The envelope length carries no predictor–outcome information, and experiment rule 3 permits such distribution checks if they are recorded, which this round does (§3).<br>• **Rule:** for every repository, segment length L_r = min(2000, post-C0 envelope_r), with eligibility requiring **L_r ≥ 1,500** (three quarters of the design segment). Any floor in (316, 1666] gives the same eligibility as the original rule for every candidate except TypeScript, so the change admits exactly one repository and cannot steer the Go selection. All supply, cap, composition and post-winner floor gates stay unchanged.<br>• **Parameterise every 2000-dependent constant by L_r:** INV:538 and the projection vectors; OC vector length, the `IntN(2000)` streams and the commit bootstrap (⌈L_r/20⌉ circular starts of 20, truncated to L_r); the near-boundary plants ("one per 2000 positions"); replay "all 2000 commits/cell"; and the 24,000 Jev-run bound, which stays an upper bound. Re-time the OC if the code changes.<br>• Recompute the five envelope counts with reviewed code (R6-02) before relying on them.<br>• State in PREREG §3 and ADR-013 that the rule was changed after observing these counts and nothing else. Disclose that Prometheus is the sole viable Go candidate, together with its E1b-v2 exposure.<br>• Do not replace TypeScript, and do not move C0. |
| **R6-02** | **major** | LOC `connectivity.segment_commits`/`segment_end` for all five repositories, plus `object_lines`. No producing script exists in the worktree: grep for `segment_commits`, `all_reachable_connectivity` and `restoration_method` outside JSON finds no generator. False or incomplete claims: EXPORT_RECIPES.md:13 ("not supply/diff/inventory analysis"); `inventory-inputs.json`:2 ("no candidate scan/supply analysis/inventory"); README.md:10 ("Object-only evidence"); ADR:72 ("No candidate inventory … informed these choices"); RESP (no mention of the counts). | **A supply-gate quantity was measured on the candidates, was not disclosed, and came from unreviewed code.** Computing each "observed-envelope" segment end means walking the post-C0 first-parent chain and recording its length. That length is exactly the inventory eligibility gate. The packet says the opposite. Experiment rule 3 allows such checks only if they are recorded in the review round. The drafter did not flag them, and the numbers' provenance cannot be audited. Separately, Prometheus `result_HEAD` `1d4faeaf…` differs from `resolved_remote_tip` `2f5baf82…` without explanation. This is harmless because the anchors are pinned and HEAD is unused, but it is unexplained. | Add the generator (or a reviewed `envelope_check.py`) to the packet. Main reruns it once, Git-only and read-only, and records verbatim output, argv and environment in the round-7 packet. Correct the four statements above so they say that envelope lengths and object counts were observed. Cite this round's §3 record. Add a one-line explanation of the Prometheus HEAD discrepancy. Carry the disclosure into PREREG §3 and ADR-013. |
| **R6-03** | **major** | PREP:150–155 (`subprocess.run`, no `timeout`); PREP:742–777 (any `BaseException`, including `KeyboardInterrupt`, starts the next attempt); PREREG l.42, l.44, l.227 ("preparation/OC have their own compute caps"); `prep-scale-projection-v6b.json` | **Preparation has no compute cap and no enforced termination.** l.227 is false for preparation: only the OC has a numeric cap (OCGO:20). The populated projection is 307,279 s ≈ **85.36 h** per repository per attempt, and 921,837 s ≈ 256.07 h for three attempts. Over five repositories run serially that is ≈ 426.8 h (17.8 days) for one attempt each and ≈ 1,280.3 h (53.3 days) for three. There is no ceiling to compare these against. This answers the question of whether the projection fits a preregistered ceiling: none exists. The 42-day clock starts later (l.227), so it does not apply. A hung scanner never ends. Three operator interrupts turn a candidate into `PreparationInfeasibleOperational` and exclude it before ranking. That is an unregistered experimenter degree of freedom over which Go candidate wins (peer). | Freeze the following in PREREG §3 and enforce them in code:<br>• a per-subprocess timeout of **3,600 s** (≈ 31× the measured 115.6 s per 80k-file revision), killing the whole process group;<br>• a per-attempt wall deadline of **14 days** (≈ 3.9× the projection);<br>• a whole-preparation ceiling of **42 days**, with up to five repositories concurrently and the concurrency stated.<br>Timeouts and deadlines produce sealed operational attempt failures, never content failures. `KeyboardInterrupt` must not silently start the next attempt: seal it as an operational deviation. Interruption is permitted only for documented host failure, never chosen per candidate, and is reported under Deviations. Add fixtures for timeout → operational seal and for the interrupt policy. Correct l.227. |
| **R6-04** | **major** | PREREG l.219–223; RESP l.47–51; ADR row "Vague projected capacity…" | **The H and d gates require the full 149,760-session label reserve to fit, which makes the k gate's A_H and A_D terms dead code.** H ≤ 3000 ⟺ λL/3600 ≤ 3000 − H_C − R − F ⟺ A_H ≥ L, and likewise d ≤ 42 ⟹ A_D ≥ L. Whenever both gates pass, a = L, and the k gate reduces to min(proxy N, 2600). RESP rows 2–3 (k = 2540/2468, above the floors) are therefore stops that serve no purpose. This contradicts l.211, which says labels "degrade gracefully like an unstarted label tail". Break-even, using RESP row 1 (H_C 2.4 h, R 345.944 h, F 9.3 h): the hours gate allows λ ≤ **63.5 s** and the wall gate λ ≤ **70.7 s**. Dropping the full-L gates and keeping the frozen uniform pool split, the selected-repository floor binds at a ≥ 6 × ⌈2106 × 9.6⌉ = **121,308**, which allows λ ≈ **78.4 s** (hours) and ≈ **87.3 s** (wall). The current rules thus build in a predictable band of avoidable pre-map not-testable outcomes, and that band is frozen at preparation authority (found by the peer; the council recomputed it for the uniform pool split). | Delete the stand-alone full-L H and d inequalities. Keep two things:<br>• fixed-part feasibility: H_C + R + F ≤ 3000, and 2 + [(R+J)/6 + F]/(24u) ≤ 42;<br>• the existing per-cell k gate with a = min(L, A_H, A_D) and pool allowance = min(24,960, ⌊a/6⌋) ≥ ⌈floor × 9.6⌉.<br>Labels then stop at their hash frontier as the unsampled tail l.108 already defines. Keep the uniform pools; no new allocation machinery is needed. Correct the RESP table and the ADR row. |
| R6-05 | minor | SCALE:63–114 (one identical tree; `prep.materialize`/`run`/`heuristic_paths`, bypassing `ownership()`/`execute()`/`verify_preparation_seal`; map-response JSON not decoded, unlike PREP:419); `prep-scale-v6b.json` `preparation_script` = rev-6 digest **45fc5eb7…**, whereas rev 6b pins **d7f091fb…**; projection multiplier = files × revisions only | **The scale evidence does not establish capacity for the production path or for the candidate byte volumes.** The fixture holds 4,471,247 B per revision (≈ 56 B per file), with identical trees and warm object reuse. It does not execute the frontend. It measured the previous `prepare_exports.py`. [INFERENCE] Candidate trees such as TypeScript and Django carry far more bytes per revision. Materialisation, `cat-file --batch` and possibly the scanner scale with bytes, and the 100,000-file assumption is not checked against the candidates. These are disclosed as limitations, but the projection is the only evidence that preparation fits the R6-03 ceiling. | Re-run the benchmark against the final signed `prepare_exports.py` through `execute()` on a synthetic repository, with about 5% of files changed per revision. Include `verify_preparation_seal` and the built frontend consuming the NDJSON stream. Add a byte-heavy synthetic run (recommended ≥ 2 GB per revision × 3 revisions) to measure per-byte cost. Project with max(file ratio, byte ratio) under a fixed 2 GB per revision assumption, and show the result fits R6-03's deadlines. No candidate is inspected. |
| R6-06 | minor | PREP:543–548 (`verify_trace` returns after preflight checks for refused seals); PREP:767–775 (content refusal records a free-text message and `failed_step`) | **Content-refusal seals are not tied to the trace.** A seal that contains only valid preflight entries and claims `class: content` passes verification. A content refusal is scientifically decisive: a mandatory repository becomes not testable, and a Go candidate becomes ineligible before ranking. So a self-asserted content seal could steer selection. Exploiting this needs deliberate forgery. | For content refusals, require the final attempt's trace to contain the `ls-tree` (and the batch, where applicable) for an in-scope revision named in the refusal, with the failing step and revision recorded as structured fields. Reject content refusals that have no traced Git step. Add a forged-content-seal fixture. |
| R6-07 | minor | PREREG l.136 ("This conservative stress estimator"); l.229 (threats) | **The min-of-one-way bound is not conservative against two-way dependence.** The two-way variance V_c + V_d − V_cd (Cameron, Gelbach & Miller 2011) can exceed either one-way variance. The grid tests ρ_dir only up to .10, and the few-directory case (small K) is not listed as a threat (peer). | Replace "conservative" with "robust to either dimension singly". Add "two-way dependence; directory ρ > .10; few-directory bootstrap" to the threats. Report K and the effective directory count 1/Σshare² per repository × age. |
| R6-08 | minor | PREREG l.199 ("elapsed/work projections beyond the cap"); OCGO:281, 305–307 | (a) The text does not say which projection triggers the stop. The code uses the 256-replicate minimum (`scenarios*256`). (b) `status` is overwritten: an earlier `not-testable-OC-gate` becomes `not-testable-compute-budget`, so the stop reason is misattributed (peer, confirmed in code). | State "minimum 256-replicate projection over all scenarios" in §9. Record a `stop_reasons` list that keeps the first gate failure and the compute stop separately. |
| R6-09 | minor | PREREG l.16, l.139; `hypothesis-ledger.md`:35 (HD8 "…every repository tested, including a high gold-`none` one"; Tested by "X11") | **No mapping from verdict to HD8 ledger state.** Under alternative D the high-none check is descriptive, so "supported" combined with "high-gold-none condition not realised" has no defined HD8 state. "Owner records" also conflicts with R-064. | Preregister the mapping:<br>• supported + realised → `supported` (these repositories and draws);<br>• supported + not realised → `inconclusive — accuracy supported; high-gold-none clause untested`;<br>• not-supported → `not supported`;<br>• otherwise `inconclusive`.<br>Change the text to "Main records under R-064". At sign-off, record HD8 as `preregistered`, tested by X11a. |
| R6-10 | minor | LEDGER:77 (R-075), :79 (R-079), :75 (R-073), :78 (R-076); PREREG l.199, l.11; `STATUS.md`:128 | **The ledger rows do not support what they are cited for.**<br>• R-075 records only the rev-5 timing, but PREREG l.199 cites it for the `oc-v6-timing-*-r3.json` profiles.<br>• R-079 does not link `prep-scale-v6b.json` or `prep-scale-projection-v6b.json`, nor their 2,455.8 s and 307,279 s results.<br>• R-073's Link points to the PREREG rather than the session (the R-063/R-064 pattern).<br>• Berger (1982) and Cameron, Gelbach & Miller (2011) shaped the IUT gate and the directory estimator, but have no rows or links of their own (ledger rule 1; X1 precedent R-072/R-078).<br>• The `Informed` column uses review-issue IDs (`X11a R4-02`, `R5-01`) rather than decision or hypothesis IDs (ADR011-R3-02 precedent).<br>• The STATUS pointer is stale ("next free R-078"; R-079 unlisted). | These rows have not landed on main, so they are still editable. Extend R-075 with the r3 profiles and metadata and their numbers, or add a new row. Add the receipt, the projection and their results to R-079. Change R-073's Link to "This project's session, 2026-11-08". Add paper rows: Berger 1982, *Technometrics* 24(4):295–300, https://doi.org/10.1080/00401706.1982.10487790, and Cameron, Gelbach & Miller 2011, *JBES* 29(2):238–249, https://doi.org/10.1198/jbes.2010.07136, with the caveat "cited through the council". Put HD8, X11a and ADR-013 in `Informed`. Update the STATUS pointer. |
| R6-11 | minor | ADR:1, :3, :9, :72, :76, :80; folder holds only `ADR.md`; `docs/adrs/README.md` ("adopted only after an innovation-council round with zero issues") | **ADR-013's form breaks the ADR rules and precedent.**<br>• (a) The frozen status line carries "owner review pending", which rulings ADR011-R1-02 and ADR012-R1-15 forbid.<br>• (b) "Accepted" is claimed before any zero-issue round.<br>• (c) The Review section names the superseded `rev6-digests.json` (f62a12c9…), not rev 6b.<br>• (d) There are no `round-N.md` files (design rule 2).<br>• (e) The index row and Appendix B card are deferred to "PR landing". That would be an unsigned edit to the governed `docs/scout-design.html` (design rule 1).<br>• (f) The heading and metadata differ from the sibling form (`# ADR-NNN: …`, `- **Status:**`, `- **Decided:** … **sources:**`).<br>• (g) ADR:72 omits the envelope observation (R6-02). | Follow route (a), as ADR-011 and ADR-014 do:<br>• Status: "accepted; decided by the agent under owner delegation (R-064)", effective on the folder's final zero-issue round. Keep pending-owner state only in the README and STATUS trackers.<br>• The Review section names the folder's final round, not a manifest.<br>• Save study round 5 and this round byte-identically in the folder as `round-1.md` and `round-2.md`.<br>• Put the index row and the Appendix B card in the signed set.<br>• Align the heading and metadata format.<br>• Add rows for the R6-01 and R6-04 decisions, and correct ADR:72. |

**Totals:** 11 new issues (1 blocking, 3 major, 7 minor), plus R1-01 carried.

Considered and not raised:
- **Test weakening between rev 6 and rev 6b.** Unchanged: `selftest.py` (223da673…) and `frontend/main_test.go` (1b49b60b…), per both manifests. The exact-SHA plus zero-output assertion remains (`main_test.go`:241–251), and `main.go`:191–197 now wraps the decode error.
  - Canonicalisation happens where fixture roots are created (`Path(...).resolve()`, PST:87, OCST:78), not by loosening assertions. The exact `assert_any_call` and `assert_called_with` checks remain (OCST:102–103).
  - The preflight fix is on the producer and verifier side (PREP:745–746, 536–541), and missing or tampered preflight plants were added.
  - The rev-6 bytes of the two changed test files are not preserved; they are untracked, with only digests 59f0d177… and 2eb61b82…. A full diff is therefore impossible. No rule requires keeping an unsigned intermediate, and every fix site was inspected, so this is not raised.
- **Python closure.** The relocated read-only CPython 3.14.10 root (2,505 files and 9 links) and the stdlib-first bootstrap were exercised by Main's launcher, so there is no relocation failure. The stdlib `.pyc` files are pinned rather than assumed to be bypassed by `-B`. `local_module_paths` is manifest metadata, not a live `sys.path` comparison, and read-only modes do not stop a hostile process running as the same owner. Neither is a defect for the stated threat model (challenger).
- **Clone restoration.** These are genuine ordinary clones: canonical URLs, `--no-checkout --no-hardlinks`, pinned anchors verified, `rev-list --objects --all --missing=error`, `fsck --connectivity-only` exit 0, and no replacement refs, grafts, shallow history or promisor packs (RESTORE, all five repositories). The supply disclosure gap is R6-02.
- **Calibration measures burst throughput rather than sustained throughput; u = .65 is applied on top of Q; the 36τ wall reserve is six times the true need.** All three are conservative and disclosed (peer).
- **The OC `--recipe` check accepts any pinned recipe source.** Other recipes fail source-drift and PGO checks, or build an unrelated binary (peer). No validity risk.
- **The 384 h OC cap does not predictably bind.** The balanced maximum is 280 h. For a singleton profile, 384 h allows 11.29 M studies beyond the 256 minimum, so a compute stop needs more than 26.2% of scenarios to reach 1024 replicates. That would mostly coincide with a gate failure [INFERENCE]. The stop is disclosed.

## 3. Review of the specific rev-6b elements and rule-3 record

- **Experiment rule 3 record (distribution and eligibility facts observed before the freeze; no predictor–outcome association involved):**

  | Repository | Post-C0 first-parent segment (`chain[:2000]`) | Segment end |
  |---|---|---|
  | django | 2000 (capped) | `220db1c7…` |
  | microsoft/TypeScript | 1666 | pinned tip `6161772c…` |
  | gin | 316 | tip `0f09c3a9…` |
  | cobra | 70 | tip `adbc8813…` |
  | prometheus | 2000 (capped) | `be419d80…` |

  - Cold-to-segment-end `object_lines` counts are also recorded (for example, django 36,977).
  - Source: LOC and `persistent-locations-v6.json`. The producer was not reviewed (R6-02).
- **Populated scale benchmark.**
  - Actual figures: 80,001 owned paths per revision across 81 features, 1,600,020 streamed paths, 59,208,700 record bytes, elapsed 2,455.807 s, processing 2,311.121 s (115.56 s per revision), peak child/self RSS 188,481,536 B.
  - Multiplier: (100,000/80,001) × (2,002/20) = 125.1234, consistent.
  - It is genuine component-capacity evidence. It is not a production-path or byte-volume bound (R6-05), and there is no ceiling to test it against (R6-03).
- **Immutable Python closure.** Accepted (see above).
- **Rev 6 → 6b test fixes.** No weakening found (see above). The rev-6 failure record and the intermediate failure, which exposed a genuine pre-command trace defect, are preserved verbatim in `rev6b-checks.json`. That disclosure is commendable.
- **Clone restoration evidence.** Accepted as object and connectivity evidence. The undisclosed supply quantity is R6-01 and R6-02.
- **IUT thresholds and constants.**
  - All recomputed: 12 × 39 × 2 × 3 × 2 × 5 × 2 = 56,160 scenarios.
  - Descriptive α = .05/1,010,880 = 4.946 × 10⁻⁸.
  - Quantile Monte Carlo standard errors .0078160 and .0015612.
  - Projections: 70.006/280.02 h balanced and 215.08/860.33 h singleton.
  - Sessions: 6 × 3 × 8,320 = 149,760; plus 180 = 149,940 ≤ 150,000.
  - The three RESP calibration rows reproduce exactly: 128414/45 h and 313645/8424 days; 137783/45 h; 10668185/244296 days.
- **Ledger rows R-073 to R-076 and R-079.** Format: 8 columns. Links resolve. No ID collisions: X1 holds R-078 and R-080/R-081, and ADR-012 holds R-070 and R-077. The canonical main R-064 row is retained and the X11 duplicate removed, which satisfies ADR-011 round-4 condition C2. Content residuals are in R6-10.
- **ADR-013 content.** The decision table is complete for rev-6 choices and its rationales match PREREG. It must add R6-01 and R6-04 and fix its form (R6-11). ID-013 is sequential: ADR-012 is in `scout-adr11` and ADR-014 in `scout-x1`.
- **Model-free, no outcome leakage.** Confirmed for PREP, ASM and SCALE:
  - `map --source heuristic` runs with a fresh state and HOME;
  - the sandbox denies network access;
  - PATH holds only the pinned Git;
  - no model CLI, gold, selection or predictor–outcome computation.

## 4. Export-preparation authorisation

What rev 6b does well:
- bounded, batch-materialised, trace-verified preparation;
- refusal classes that do not turn tool faults into exclusions;
- OC input binding;
- honest intersection–union error control;
- a directory-robust estimator;
- measured calibration gates;
- bounded stopping;
- duplication honestly unmeasured;
- relocated immutable tool closures;
- genuine end-to-end synthetic seals;
- independently exercised checks.

Why it may not run now:
1. **R6-01.** Packet evidence already makes X11a, as written, not testable at inventory. Preparing exports under that design would waste multi-day compute and freeze a known-failing preregistration.
2. **R6-02.** The supply observation behind R6-01 was undisclosed and its producer is unreviewed.
3. **R6-03.** The authorised operation has no ceiling and no guaranteed termination.
4. **R6-04.** A frozen stopping rule produces avoidable not-testable outcomes.
5. PREREG freezes at this authority (§2, l.32), so every minor item must be cleared first as well.

## Verdicts

**(a) Export-preparation authorisation: `not authorised`.** Reviewed bytes, all copied from manifest `reviews/rev6b-digests.json` (SHA-256 7e873019ac128958347d2e6cb80b148481eedc026050fb3c6abca93b2d2b1acd, computed by Main):

- `PREREGISTRATION.md` 5dd3893b6eda7f8718d3da0cca55755f3e151e686d60af626588122396bdb823
- `EXPORT_RECIPES.md` c78f23d769953ce46dc5a16d215c085c21e777ecf9f27de2b2a2ae8a4ee8104d
- `README.md` 0962b85210eea46e213088657a8f13e4d19af1e1470be8f347b6eadbfdf4da3d
- `prepare_exports.py` d7f091fb273ae57f926feeb5621793c3224de16162ffde82460b67c28b228b9c
- `assemble_preparation.py` d250dbbb2cb7bf3463864d580273dfa62a864d6322ce7388e67c8e61ce6ba0ad
- `prep_selftest.py` 83702d9925ce5877ad1f29d3125dcc33e2fc1bb09ed89c7de5b5da4851633a05
- `pin_tool_closure.py` 39dff5b8a75e41852af7d6f5f534728ed0667626e1eb42e8ab38e657109e8348
- `bootstrap_v6.py` 5b8735ad1c03ca0afebf54dfe3f1123a96dabb72002db53de16f60d2ae0244e9
- `bootstrap_selftest_v6.py` a9901432ace306ce867aa1e09ef617a24a4e8b6347359e86415a831784978ff7
- `stages.py` e58dc239e13615c989410502afe47a8ef64f84a1e810bd6749551c654cb06a02
- `inventory.py` 85979ccdf1d33540612eb3eb7064ad6c554c2aafa224edbdf6f8a4002b9808d4
- `selftest.py` 223da6736087b8efd0cbcfb962825f7646f6c39b14e3b4741bc9c140ca21b391
- `extract_exposures.py` 966f7ed55d0a30c3d5cc4e2deea5c78b912f98827e261bbdb9b265c44749c6b8
- `frontend/main.go` 8bbe00b526dd39ef0905fb85e47d85326418e57c062ebb241d8fd55770c4c6f9
- `frontend/main_test.go` 1b49b60b85568ef8d3b440d12590335e5a9e17eb7b9997514af9827b833cf3c0
- `inventory-inputs.json` e492da740af2aa778304394dd326c269099e7e7aa30a9207e680c92b0f34c74b
- `exposure-source-index.json` bccbfd4b8a2e42c9a91cdd9ead0c3f939c5606950a5fcb9844f8d4b6a141c3d8
- `exposure-contexts-v4.json` 2dd5b6b7d6665d64519585acdb63ed6eafadc3092bda25a38433616f896199fb
- `latency-basis-v4.json` bdc3cef935d65a20b4b0b9f34bb68c881d29da4960e74e5ee5b60549f0894104
- `cost-feasibility-v5.json` 1637256f02d54164c73eff82abb8648b4f70f4924701c9c29ee5a54e6c76c1d0
- `oc-v5-spec.json` 5c373731eb67c9bc251eb187f2d68ea673c09c81965825a396bf46663da2e732
- `oc-v5-fixtures.json` 3c68cbb3d3e66796db9efaeb2175197fcd9b54f4658027dab7a6706040f561ee
- `oc_v5.go` ebc60a467dfc574c57ab5961c7ceafab2cef1b479ed1e7812faaf8fba40cdb22
- `oc_v5.py` 405a244ff9733f94ce22906d6e6c6e6b4ab402e2a40382a8e1f4f6920597797b
- `oc_selftest_v6.py` 7bba25f325a731fd2009e73a4520737c1bc558f680ec62bd87b4f97be7f273c7
- `oc-v6-timing-metadata.json` f99f67a43b7450fa3bcfd79cf569544b43a84e503dc028456325020b76961511
- `oc-v6-timing-balanced-r3.json` 7650a4a9234709681ca029bb6a2ab6f5300cc76efe0ba1edc510b52b95248c66
- `oc-v6-timing-concentrated-r3.json` 207ddd56e3c4ee7b0fa30170fa8085df728c1abb70e1e55375b836c8d0ea8c67
- `oc-v6-timing-singleton-r3.json` f144bbb4b7b1fb7667f4cdaa0fdf087a2f2f6eff62279df2fa4cf3cd8dc84e20
- `prep_scale_v6b.py` 1a89246556d5367d0f4eb955d483d294390481c61dbb4a19b0e1fc095cf7db41
- `prep-scale-v6b.json` 49d4f502f70c1b134e8a9a0903bdeface84205a881ea4f843417869c4ea75ded
- `prep-scale-construct-v6b.json` 20077b7ecdf571648871038ae7be3e51e90fd289420617ab8a5f424c6d4fd3ae
- `prep-scale-projection-v6b.json` b448460c758d009d761401db2e53f3f74667708c618747bd410eded4d93f7538
- `prep_scale_v6.py` ddea776867bba3e5eeb2e15b2589c7f2b0264df727a72eb9689b78f5e7cb7448
- `prep-scale-v6.json` 27d3612d8ddf59958040adbe4604d57b724fe460acd206602ff42c35f1d07104
- `prep-scale-projection-v6.json` f56243f117224e8e94da473bd21fa063b57a543d513590f96f5bd922dd6460e4
- `prep-scale-assumptions-v6.json` 57c5e0c08922b7d5857845e36f0d603efa9e0a013c45415668c8c8252cdd733c
- `tool-pins-v6/clawpatch-tree-v6.json` 423a786aef8874ac34e01c9f9c3d0b52f78ddffb6af774a80801c3bd5c3b9f29
- `tool-pins-v6/frontend-build-recipe-v6.json` f5f18ea51218d69cfe058a8348f3841b40b47875de1ce4b993a600ba4441f41c
- `tool-pins-v6/frontend-go-closure-v6.json` 788f2a3be52a471602a4d622c90dae88b52d79a5bd8b8c1d9383c2ae42a22e37
- `tool-pins-v6/oc-build-recipe-v6.json` fc261bbe4fab4ca713374d271f4d8aa54a7690fb1257a4c5a4b7b11fca69f5ef
- `tool-pins-v6/oc-go-closure-v6.json` d62d84cbf2b46a107231005c61b9ed2a3858080874d4ac6794ac039e4b36bf21
- `tool-pins-v6/persistent-locations-v6b.json` 469e54fffc2d2628a457200d33945ee9bc1f4a16702d87096fb2c77e243c0ddb
- `tool-pins-v6/persistent-locations-v6.json` ab610f59d9a335762442894226b19b2fc1ba93d9aa185be3ee469289c746a130
- `tool-pins-v6/python-runtime-v6b.json` a33cccdec6af44cfb8fd34006d89a633840afef2444c00a88fa3249fe80978ac
- `tool-pins-v6/object-restoration-v6.json` 95854874cfde17d52a806e442fb7ea3597c75732d1e2beb073ef808e49e2094c
- `exposure-extraction-v4/seal.json` 96fcfcbfada87faaa6624d9124fb3c88d55d714a8cd0bf79ce72a114e32f8510, and the five v4 outputs as pinned in it
- `reviews/round-5-response.md` e109cf65d25078ff9d67289eda1eeb3d829371036f9fa8ee673e5d39818a28a3
- `reviews/rev6b-checks.json` 348d4c33c8c0456e32734c38dba8e03e0b8a2e5ca7687dea91bfaad9e7137037
- `reviews/rev6-digests.json` (superseded) f62a12c9a26bc2038cc2609f5912a833006471b940efdf52f2ea0f733a5e3f81
- `../x11b-guard-recovery/PLAN.md` 034284255f303ad09308f30fc6212efac6d2730aa50ac724c4a8243df3d48753
- `docs/research-ledger.md` 537193107dcbf996914decb0db8e2905187287184575d468779735ceaf4bb66b
- `docs/adrs/2026-11-08-x11a-delegated-decisions/ADR.md` be3f5bed72a79617bc11ce8bb8ba438e08eb0523d55c587d08c8e08d010dde97

None of the following may run: ownership scan, frontend build, exposure operation, inventory, OC run, map, label, replay or model call. Inventory still needs its own artifact-binding round after a future authorised preparation.

**(b) Overall: `12 issues open`.** That is R1-01, carried and deferred until the implementation round, plus 11 new: 1 blocking, 3 major and 7 minor. The study is not signed off. HD8 stays `proposed`. This round changes no ledger, design or product file.

**ADR-013 verdict: `not signed off` (ADR.md be3f5bed72a79617bc11ce8bb8ba438e08eb0523d55c587d08c8e08d010dde97).**
- One issue is its own: R6-11, minor.
- Its decision table also has to absorb the R6-01 and R6-04 decisions and the R6-02 disclosure.
- It must be re-reviewed with the rev-7 packet and cannot carry an "accepted" status before its folder's final zero-issue round.
- Owner review stays pending, recorded in the trackers only.

## Council synthesis

### 1. Recommendation
Revise to rev 7 and request export-preparation round 7, with ADR-013 in the same packet:
- Amend the segment rule before the freeze, disclosing the eligibility data that informed it: L_r = min(2000, envelope), floor 1,500 (R6-01). Reproduce and disclose the envelope counts with reviewed code (R6-02).
- Freeze and enforce preparation timeouts, deadlines and the interrupt policy (R6-03).
- Drop the full-L H/d gates in favour of the k gate (R6-04).
- Clear R6-05 to R6-11.
- Keep everything that §1 marks closed.

### 2. Alternatives
- **A. Recommended: amend before the freeze, with disclosure.** It keeps the owner's repository set and the REQ-176/177 evidence path. It changes eligibility for TypeScript only. Cost: a parameterised OC and inventory, and one more round.
- **B. Declare X11a not testable on the existing evidence and supersede it later** (challenger's fix). This is valid and maximally conservative. But nothing is frozen yet, so any successor would be designed with the same knowledge: no protection is gained, and the X11a evidence path is lost. Not recommended.
- **C. Replace TypeScript with another mandatory repository.** This is a larger supply-informed change: a new repository's envelope, exposure and test patterns would all need fresh restoration. Rejected.
- **D. Move C0 earlier for TypeScript or for all repositories.** This changes the meaning of catalogue age and makes ages heterogeneous across repositories. Rejected.
- **R6-04 variant: keep the full-L gates and only disclose the break-even.** This is valid, but it builds in avoidable not-testable stops between λ 63.5 s and 78.4 s. Not recommended.

### 3. Evidence
- **Envelope:** LOC:10123–10124, 10223–10224, 10323–10324, 10423–10424, 10523–10524. INV:538, :685. ST:43–45. PREREG l.41, 46, 54, 60.
- **Undisclosed claims:** EXPORT_RECIPES.md:13; `inventory-inputs.json`:2; README.md:10; ADR:72. No generator found by grep.
- **No termination:** PREP:150–155, 742–777; PREREG l.227; OCGO:20 (the OC cap only).
- **Gate redundancy:** PREREG l.219–223 algebra. RESP l.47–51 rows reproduce. Break-even 63.5 / 70.7 s now, versus 78.4 / 87.3 s.
- **Scale:** `prep-scale-v6b.json` (`preparation_script` 45fc…, versus d7f0… in the manifest); SCALE:63–114.
- **Refusal seals:** PREP:543–548.
- **Thresholds:** peer exact-binomial checks; OCST:147–155.
- **Ledger:** LEDGER:75–79. X1 R-072/R-078 precedent for paper rows. DOIs verified by web search (doi.org/10.1080/00401706.1982.10487790; doi.org/10.1198/jbes.2010.07136).
- **ADR form:** `docs/adrs/README.md` rules; ADR-011:3; ADR-014:3, :15, :66, :74–77 (`scout-x1`).

### 4. Risks
- **Alternative A will look like supply-informed tuning.** Mitigation: the floor is outcome-equivalent anywhere in (316, 1666]; the change is disclosed in PREREG, ADR and this round; envelope length carries no information about accuracy.
- **TypeScript at 1,666 commits** supplies about 17% fewer clusters, which raises the chance of failing its 1,404 per-age floors. That would be a disclosed not-testable stop.
- **Prometheus is now the sole viable Go candidate.** If it fails any gate, X11a is not testable. Its E1b-v2 exposure must be disclosed.
- **Preparation may hit the 14-day deadline** on byte-heavy repositories. That ends as an operational not-testable, never as a content exclusion. R6-05's byte benchmark lowers this risk.
- **Reversibility:** everything is reversible. No candidate scan, build, inventory or model call has run, and all historical artifacts are preserved.

### 5. Execution outline
1. Main saves this record as `reviews/round-6.md`. It copies study round 5 and round 6 byte-identically into the ADR-013 folder as `round-1.md` and `round-2.md`.
2. **R6-02:** add a reviewed envelope generator. Main runs it once, Git-only and read-only, records verbatim output, and corrects the four claims. Check: the counts equal §3, or any discrepancy is explained before step 3.
3. **R6-01:** amend PREREG §3 (L_r rule and floor 1,500). Parameterise INV:538, the projection vectors, the OC spec, Go/Python code and fixtures (vector length, streams, ⌈L_r/20⌉ blocks) and the plants. Re-time the OC if the code changes. Check: fixtures with L_r ∈ {1500, 1666, 2000}, and 1,499 refused.
4. **R6-03:** implement the 3,600 s per-command timeout with process-group kill, the 14-day per-attempt deadline, the 42-day preparation ceiling with stated concurrency, and the interrupt policy. Correct PREREG l.227. Check: timeout and interrupt fixtures produce sealed operational records.
5. **R6-05:** re-run the synthetic benchmarks (production `execute()` with changed content, verifier plus frontend stream, and a byte-heavy run) with the final signed `prepare_exports.py`. Project with max(file, byte) ratio against R6-03.
6. **R6-06:** bind content refusals to the trace. Check: a forged content seal is refused.
7. **R6-04, R6-07, R6-08, R6-09:**
   - edit PREREG §10 gates;
   - correct the RESP table and ADR row;
   - change the wording and add the threats;
   - add `stop_reasons` and state the projection basis;
   - add the HD8 mapping.
8. **R6-10 and R6-11:**
   - add the ledger edits and the two paper rows, and update the STATUS pointer;
   - align ADR-013's form, status and Review section, and add its index row and Appendix B card to the signed set;
   - add the new decision rows.
9. Main recomputes the rev-7 manifest and runs the pinned isolated selftest, the OC fixtures and the PGO-off frontend Go tests once, on synthetic data only. Main records the exercised results and requests round 7 (export-preparation scope plus ADR-013).

### Where the peers changed this memo
- **innovation-challenger:**
  - Independently found the TypeScript 1,666 envelope failure, which confirms the lead's finding (R6-01). Its fix (record not testable) is kept as alternative B and overruled for the reasons in §2.
  - Found the missing preparation ceiling and subprocess timeout (merged into R6-03).
  - Found that the benchmark ran against the rev-6 `prepare_exports.py` and bypassed `execute()` and verification (R6-05).
  - Found the stale manifest in ADR-013 (R6-11c).
  - Its test-integrity limitation (rev-6 bytes not preserved) was considered and not raised, after the council inspected each fix site.
- **experiment-peer:**
  - Supplied the full-L H/d redundancy (R6-04), which was new to the lead. The council corrected its 94,352-session break-even to 121,308 (λ ≈ 78.4 s), because PREREG splits the pools uniformly.
  - Supplied the interrupt degree of freedom and the recommended timeout and deadline values (R6-03), the byte scaling (R6-05), the two-way-dependence wording (R6-07), the OC stop-reason overwrite, which the council confirmed at OCGO:305–307 (R6-08), and the HD8 mapping (R6-09).
  - Supplied the R-075/R-079 coverage gap and the STATUS pointer (R6-10), and the ADR precedent and form items (R6-11).
  - It missed the envelope failure; its "R5-01 closed" call was accepted.
- **Lead's own additions:**
  - The envelope failure, found independently.
  - The undisclosed supply measurement, the false "no supply analysis" claims and the unreviewed generator (R6-02).
  - The Prometheus HEAD discrepancy.
  - Forgeable content-refusal seals (R6-06).
  - The paper rows and `Informed`-column defects (R6-10).
  - The amend-before-freeze recommendation, with an outcome-equivalent floor.
  - The R5-02 partial status (hung attempts never reach the terminal rule).
