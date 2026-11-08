# Council round 7: X11a merge-time association accuracy, preregistration rev 7, export-preparation scope (also ADR-013 round 2)

| Field | Value |
|---|---|
| Study | X11a merge-time association accuracy (HD8). This is the accuracy part of REQ-176/REQ-177. Guard and recovery evidence belongs to X11b (`../x11b-guard-recovery/PLAN.md`). X11b is a plan only and is not preregistered. |
| Digest manifest | `reviews/rev7-digests.json` has 1,749 entries and excludes itself. Main computed its SHA-256 as **892b3ab79834b856d44b4e6b12bda384be8f1feb1927297349579e900e8786b4** and verified all 1,749 entries. The council has no hashing tool. Every digest below is copied from the manifest; none was recomputed. |
| Requested scope | **Export preparation only.** This covers the model-free Clawpatch heuristic ownership scans at the pinned commits, the council-gated frontend build and the exposure-binding operation, which together produce sealed ownership exports. Inventory, OC runs, maps, labels, replay and model calls are out of scope. ADR-013 is reviewed in the same packet as its round 2. |
| Owner decisions in force | R-073: "(b) High-none only on selected repo" (alternative D). R-064 is a standing delegation: wherever an owner choice arises, the recommended option is stated as the required fix. |
| Independent checks (Main and independent agents, not the drafter) | **Report:** `/var/folders/g5/r809r2p97nz9gtj25s1c37t80000gn/T/x11-rev7-checks-agi08wdc/report.txt`.<br>• **Integrated selftest:** 29 outer and 26 nested preparation tests OK (181.8 s).<br>• **Pinned PGO-off frontend Go package:** `ok`.<br>• **Digests:** 1,749/1,749 matched before and after the checks; git status unchanged.<br>• **R6-02 envelope generator:** run once by Main. Exit 0 and empty stderr. Receipt SHA-256 f8ebcdcb….<br>• **Git-only capacity collector `preparation_capacity_v7.py` (1599b49b…):** Main audited it and found it runs only `rev-list` and `ls-tree -r -l`. Main ran it once: exit 0, 335.7 s. Receipt SHA-256 141c6764…, copied byte-identically into the study folder.<br>The council records these as exercised evidence and did not re-execute them. |
| Date | 2026-11-08 |
| Reviewer | innovation-council (anthropic/claude-opus-5-5) |
| Independent peers | innovation-challenger (openai-codex/gpt-6-astra:high) and experiment-peer (anthropic/claude-opus-5-5:high). Neither was told the lead's view. |
| Method | Read-only. Nobody edited a file or ran a script, test, build, git command or model, and nobody touched the candidate clones.<br>• Computations were limited to jq and grep over study files, plus hand arithmetic.<br>• Clawpatch mapper behaviour was read in the `third_party/clawpatch` submodule. That the pinned read-only scanner tree was built from the same v0.7.3 source is [INFERENCE] (STATUS l.120; PREP checks `clawpatch_version`). |
| Line references | `PREREGISTRATION.md` (PREREG), `prepare_exports.py` (PREP), `assemble_preparation.py` (ASM), `preparation_capacity_v7.py` (CAP), `prep_scale_v7.py` (SCALE7), `selftest.py` (ST), `oc_v5.go` (OCGO), `EXPORT_RECIPES.md` (REC), ADR-013 `ADR.md` (ADR), `third_party/clawpatch/src` (CP), `reviews/round-6-response.md` (RESP). All refer to the bytes pinned in the manifest. |

## 1. Status of prior items

"Closed" means the round-6 required fix is present in the rev-7 files. Each status was checked against the files, not taken from RESP.

| ID | Status | Evidence / remaining gap |
|---|---|---|
| R1-01 | **Open (carried; deferred as planned)** | These are still absent: the full GUIDE, real orchestration, calibration, replay, labels, scorer, stage seals, whole-pipeline plants and inventory-conditioned OC tables. PREREG l.37 discloses them, and later stages refuse. It stays open until the implementation round. |
| R6-01 | **Closed** | **Rule:** L_r = min(2000, envelope), floor 1,500, with repositories and anchors unchanged (PREREG l.41, l.46). Prometheus is disclosed as the sole length-eligible Go candidate, along with its E1b-v2 exposure (l.41, l.66).<br>**Inventory:** the gate is 1500 ≤ length ≤ 2000 (`inventory.py`:548).<br>**OC:** rejects lengths outside 1500..2000 (OCGO:165). Commit bootstrap uses ⌈L/20⌉ blocks with a truncated final block (OCGO:178–182, 348–349). Retimed at 1500/1666/2000 (`oc-v7-timing-*.json`). |
| R6-02 | **Closed** | The reviewed generator `envelope_check.py` is pinned. Main's single run reproduces envelopes 2791/1666/316/70/3248 and capped lengths 2000/1666/316/70/2000, equal to round-6 §3. The challenger checked the object counts against command `stdout_lines` and the anchors against `segment_end`/`Cold`: consistent. The stderr digest e3b0c442… is the empty-file digest. The four false claims are corrected (REC:13, README, `inventory-inputs.json`:2, ADR "Informed by"), and the Prometheus HEAD note is present. |
| R6-03 | **Closed** (residual enforcement gap → R7-03) | • **Subprocess kill:** process-group session, `killpg` on timeout or any exception, 3,600 s clipped to the active deadline (PREP:204–246).<br>• **Deadlines:** 14-day attempt deadline and canonical create-once 42-day shared budget via `ITIMER_REAL` (PREP:52–68, 870–877).<br>• **Interrupts and budget:** `KeyboardInterrupt` → `PreparationInterruptedOperational`; budget expiry → `PreparationBudgetExhausted`. Both break the loop (PREP:893–906), and the assembler stops the study globally (ASM:17–19).<br>• **Slots:** five locked slots (PREP:933–950).<br>• **Text:** PREREG l.42 and l.46 updated; the timeout/interrupt/deadline/reset/sixth-slot fixtures pass per Main's run. |
| R6-04 | **Closed** | The full-L H/d vetoes are deleted. Fixed-part feasibility and the uniform a/pool/k gate remain (PREREG l.225–231). The RESP table and ADR row are corrected. |
| R6-05 | **Partially fixed → R7-01, R7-02** | **Done:** the required production-path benchmarks exist. Final PREP d52135b0… ran through `execute()` → `verify_preparation_seal` → built-frontend NDJSON, with 5% churn, in an 80,001-file × 20 run and a 2,000,120,047 B × 3 run. Both mandated max-ratio projections fail and are preserved.<br>**Changed rule:** the replacement rule (PREREG l.48) is accepted **in principle** (§3). But the decisive TypeScript projection rests on a scanner path that no design point exercised (R7-01), and the signed texts state the rule as pending (R7-02). |
| R6-06 | **Closed** | A content refusal needs a structured revision, a `failed_step` ∈ {ls-tree, cat-file}, the scoped chain, a matching successful tree trace, and a batch trace for cat-file (PREP:637–663). The refusal schema is exact (PREP:851–855). Forged-seal plants refuse. |
| R6-07 | **Closed** | PREREG says "robust to either dimension singly, not conservative against two-way dependence" and lists the threats. K, the effective directory count 1/Σshare² and the largest share are reported (l.64). |
| R6-08 | **Closed** | PREREG states the minimum 256-replicate basis. `recordStop` keeps the first reason and appends the others without duplicates (OCGO:177, 293). `stop_reasons` and `compute_projection_basis` are in the output (OCGO:347, 350). |
| R6-09 | **Closed** | PREREG l.161 has the HD8 state map, and l.16 says "Main records under R-064". HD8 stays `proposed`. |
| R6-10 | **Closed** (one content residual folded into R7-02) | • R-073 links to the session.<br>• R-075 adds the r3 and v7 profiles.<br>• R-079 adds the v6b/v7 receipts and numbers.<br>• R-082 (Berger 1982) and R-083 (Cameron, Gelbach & Miller 2011) are added with "cited through the council".<br>• R-084 records provenance.<br>• `Informed` = HD8/X11a/ADR-013.<br>• STATUS points to the next free ID, R-086.<br>Residual: R-079 records the additive rule as "needs council approval" and omits the coefficients and the capacity result (R7-02). |
| R6-11 | **Closed** (numbering instruction corrected by the council) | **Form fixes:** (a), (b) and (f) are fixed: the status line, metadata and heading follow ADR-011 (ADR:1–5). (c) The Review section names the final round, not a manifest. (e) The index row (`docs/adrs/README.md`:28) and the Appendix B card (`scout-design.html`:1006–1023) are in the signed set. (g) The envelope observation is disclosed.<br>**Numbering (d):** the folder's `round-1.md` (7746296a…) equals study `round-6.md`. Round 6 told the drafter to file study round 5 as ADR round 1, but round 5 never reviewed ADR-013; it only asked for it to be recorded (round-5.md:63, :231). That instruction was the council's error. **The council ratifies the drafter's mapping: ADR round N = study round N+5.** This record is ADR-013 `round-2.md`. Pending wording in the card and ADR row: R7-02. |

**Status count:**
- Open: R1-01.
- Partially fixed: R6-05.
- Closed: R6-01 to R6-04 and R6-06 to R6-11.

## 2. New issues

| ID | Severity | Location | Problem | Required fix (recommended option under R-064) |
|---|---|---|---|---|
| **R7-01** | **major** | **Fixtures:** SCALE7:96–104 (every fixture is Go: one `go.mod`, flat 1,000-file packages, one function per file, no test files, byte padding as a single `/*x…*/` comment); `fixture.language: "Go"` in all three receipts.<br>**Model and gate:** CAP:69–83 (fit), :48–56 (gate); PREREG l.48; `preparation-capacity-v7.json` TypeScript row.<br>**Clawpatch:** `mapper.ts` `collectSeeds`/`shouldRunNodeMappers` (≈l.316–343) runs Node mappers only when JS/TS or a Node package manager is detected; `mapper.ts`:≈99–121 runs `findNearbyTests` per seed; `mappers/node.ts`:214–215 calls `associatedTests` per ≤12-file group; `node.ts`:454–463 filters **every** test per group. | **The decisive capacity projection extrapolates Go-only coefficients onto a scanner path that no design point exercised.**<br>• **Share:** TypeScript is 81.4% of projected single-attempt work (210,949 of 259,052 s unmargined).<br>• **Unexercised path:** TypeScript runs the Node/TypeScript mappers, which the three Go fixtures never executed. Feature count (Node groups of ≤12 files, against 81 Go features for 80,001 files) and test association, which is O(groups × tests), are both unidentified. The fixtures contain zero tests.<br>• **Break-even:** the unmargined 14-day gate fails at ≈10.1 ms per file, against the fitted 1.70 ms (5.9×). The margined gate fails at ≈6.7 ms (3.9×). These are well within plausible per-path differences [INFERENCE].<br>• **Bounded alternatives:** file count (1.017× the design maximum), bytes (inside the design range), a files×bytes interaction (bounded by the byte-heavy point to ≤ +72 s/rev), concurrency (serial sum fits) and fixed costs (conservatively charged) are all bounded by the data and leave TypeScript feasible (§3). Only this term is unbounded.<br>• **Consequence:** if the projection is wrong, mandatory TypeScript can burn up to 42 days and end `PreparationInfeasibleOperational` or `PreparationBudgetExhausted`, leaving X11a not testable. The runtime deadlines fail closed, so this is a waste and stop risk, not a validity threat. But R6-05 existed to show that the authorised operation fits **before** it is authorised and PREREG freezes (PREREG l.35).<br>• **Other repositories:** Django (Python mapper) has 24.8× headroom on the margined projection. Go repositories are inside the design. | **Add one preregistered, held-out, TypeScript-language synthetic point and bound TypeScript directly with it.**<br>**Fixture**, generated without candidate inspection:<br>• a root `package.json` with a `test` script, and `tsconfig.json`;<br>• `.ts` sources only, in a nested directory tree at least three levels deep;<br>• per revision, files ≥ **81,369** and bytes ≥ **380,592,869** (TypeScript's scoped maxima);<br>• TypeScript code text (imports, exports, functions), including at least a few files ≥ 1 MB, not a single padding comment;<br>• **half the files are tests** (a `tests/` tree plus sibling `*.test.ts`). A 50/50 split maximises groups × tests.<br>**Run:** 3 revisions with 5% churn, on the **final** PREP, through the same production `execute` → seal verification → frontend runner. Main runs it once and retains the receipt whatever the outcome.<br>**Freeze first:** before the run, record the additive prediction at the fixture's actual files and bytes in a create-once file whose digest is cited in the response.<br>**Rule:** let S_obs = whole elapsed/revision. TypeScript per-attempt projection = 1.5 × 1,668 × max(S_obs, prediction). TypeScript passes only if:<br>• one attempt ≤ 14 days, i.e. max(S_obs, prediction) ≤ 483.45 s;<br>• three attempts ≤ 42 days, which is the same bound;<br>• 1.5 × max(S_obs, prediction) ≤ 3,600 s.<br>Otherwise the result is "preparation not feasible", with no refit, rerun, fixture change or ceiling change. Report S_obs/prediction descriptively as the test of the margin.<br>**Other repositories:** they keep the additive model, and the language-path threat is disclosed for Django.<br>**Plants:** the projection code plants each failing branch (attempt, three attempts, subprocess) and the aggregate `preparation not feasible` status.<br>**Size facts:** reuse the size facts pinned in `preparation-capacity-v7.json` (141c6764…). Collect no new candidate metadata. |
| R7-02 | minor | PREREG l.48 ("proposed R6-05 rule amendment…council review required…R6-05 remains open until council accepts"); REC:122 ("Proposed replacement requiring council approval"); ADR final decision row ("R6-05 remains open until council approves"); `scout-design.html`:1013, ADR-013 card bullet 4 ("requires council review"); `research-ledger.md`:79 R-079 ("changed projection rule needs council approval"; no coefficients or result) | **Canonical texts that freeze at authorisation describe their own capacity gate as unapproved.** Under route (a), PREREG, ADR.md and the card become effective, unedited, on the zero-issue round. Signing them as written would permanently freeze a preregistration and an accepted ADR that each declare an open issue and a "proposed" rule. That contradicts experiment rule 1 and the frozen-status precedent (ADR011-R1-02). The accepted ADR body and card could not be corrected afterwards. | Rewrite PREREG l.48, REC:122, the ADR row and card bullet 4 in adopted form. They should state the final rule, including R7-01's held-out direct bound, effective on the authorising zero-issue round. Keep the history (the failed max-ratio receipts, Main's R-064 direction) in RESP, the study README history and the ledger. Before landing, complete R-079 with the final coefficients, the capacity receipt's result and the held-out receipt and result. Change no other wording. |
| R7-03 | minor | PREP:860–867 (seal and output are create-once at **caller-chosen** paths); PREP:933–950 (slots lock concurrency but record no invocations); ASM:13–30 (duplicates rejected only among the seals supplied); REC:54, :68, :71; PREREG l.42 ("never…silently retry") | **The attempt cap and the interrupt global stop depend only on operator disclosure.**<br>• **Retry after a terminal refusal:** after `PreparationInfeasibleOperational`, the same operation can be run again with new `--output/--seal` paths, which gives three fresh attempts.<br>• **Hidden interrupt:** a `PreparationInterruptedOperational` or `PreparationBudgetExhausted` seal can be left out of assembly simply by not supplying it.<br>• **Trial runs:** a run at an ad-hoc path is invisible.<br>The heuristic map is deterministic and gin and cobra are length-ineligible, so this cannot steer the Go winner or change an estimate [INFERENCE]. It does leave the R6-03 commitments unenforced: a terminal mandatory refusal or a global stop could be converted into a testable outcome without a trace. | **Enforce canonical paths:**<br>• Derive create-once seal and output paths from the canonical budget directory for each operation, and for each repository in the case of ownership (e.g. `artifacts/preparation-v7/<operation>[--<owner>__<repo>].{seal,output}.json`).<br>• `execute` refuses any other path and any existing seal.<br>• The assembler reads exactly the seven canonical seals instead of a `--seal` list.<br>**Fixtures:**<br>• a second invocation refuses;<br>• a non-canonical path refuses;<br>• an interrupt seal cannot be omitted.<br>Host-failure continuation stays a council-reviewed deviation.<br>**Because PREP changes, rerun the construct, file-heavy and byte-heavy benchmarks on the final PREP** (round-6 R6-05 required benchmarks on the final signed PREP; CAP:84–85 enforces it) before computing R7-01's prediction. Preserve the v7 receipts. |

**Totals:** 3 new issues (1 major, 2 minor), plus R1-01 carried.

**Considered and not raised:**
- **Inputs digest.** The collector ran on the preserved inputs e492da74… rather than the current e958f0f1…. The council compared the two `repositories` arrays: path, Cold, C0, tip and resolved tip are identical for all five repositories, and so is the Git pin (peer).
- **Fixed-cost charging.** Whole elapsed time includes fixture generation (2.07/165.8/10.6 s) and the frontend build (7.34/7.70/7.82 s). So c = 4.4588 s/rev charges ≈ 7,437 s of TypeScript's 210,949 s.
  - An ownership-only refit gives ≈ 117.9 s/rev for TypeScript, against 126.5 s/rev (peer). The whole-elapsed fit is therefore conservative.
  - The challenger notes that c is not an identified per-revision constant: C_run/N differs across N = 3/20/3. That is accurate, but it concerns the description, not a defect.
  - Exposure binding was not benchmarked. It only verifies preserved extraction pins (PREP:576–600) under the 3,600 s cap [INFERENCE: seconds].
- **`bytes_per_revision_min` label.** The file-heavy receipt records the final tree, 4,440,047 B; the generator implies 4,400,047–4,440,047 B. Using the maximum overstates bytes, and the effect on b·bytes is ≈ 10⁻⁴ s. Disclosed in RESP.
- **Verifier cost.** The per-revision filter over the trace (PREP:679–681) is quadratic in revisions. It runs in seal verification or assembly, outside the preparation deadlines: ≈ 1,668 × 8.3 k comparisons [INFERENCE: seconds].
- **Joint extrapolation.** 1,227 of 1,668 TypeScript revisions lie outside the convex hull of the three design points; the maximum barycentric sum is 1.1997 (challenger). Files are 1.017× the design maximum and bytes are inside the range. The interaction bound leaves ≥ 3.6× headroom (peer).
  - Revision count, 20 → 1,668, scales linearly by construction: each revision gets fresh roots and state and its own commands.
- **Concurrency.** Fully serial execution of all five repositories fits:
  - one attempt each at 1.5×: 388,578 s = 4.50 days, under 14 days;
  - three attempts each: 1,165,734 s = 13.49 days, under 42 days.
- **Capacity-gate plants.** ST:458–473 plants only `attempt_fits=False`. The council hand-checked all three booleans for all five repositories in Main's receipt. The new rule's plants are part of R7-01.
- **ADR-013 signed-set breadth (peer).** The signed set is a snapshot. ADR-011 likewise signs files that later change (`scout-design.html`). PREREG is normative, and later R1-01 completions get their own rounds without editing ADR.md.
- **Index token.** "accepted-form; … not effective until final zero-issue council round" means the same as ADR-011's token. This is a wording preference.
- **STATUS ADR range** (l.7, l.126). ADR-012 and ADR-014 sit in other worktrees, and l.128 states ADR-013's state accurately. Reconciling the range is a tracker matter at landing.
- **Disk.** [INFERENCE] TypeScript ownership records ≈ 119.1 M paths × ~37 B ≈ 4.4 GB, using the v6b ratio. Exhaustion is an operational failure and is disclosed.
- **OC timing variability (challenger).** OC profiles are not preparation replicates and do not bear on this gate.

## 3. Review of the specific rev-7 elements and rule-3 record

- **Experiment rule 3 record.** These are pre-freeze distribution facts collected by Main's Git-only, read-only collector (`ls-tree -r -l` sizes; no blob contents). No predictor or outcome was involved. Source: `preparation-capacity-v7.json` `repositories[].size_facts`.

  | Repository | Scoped revisions (Cold+C0+L_r) | Files per revision | Bytes per revision | Σ files | Σ bytes |
  |---|---:|---:|---:|---:|---:|
  | django/django | 2,002 | 6,661–7,014 | 41,755,002–45,930,981 | 13,786,781 | 89,462,686,201 |
  | microsoft/TypeScript | 1,668 | 65,704–81,369 | 208,413,944–380,592,869 | 119,119,904 | 497,759,062,718 |
  | gin-gonic/gin | 318 | 117–131 | 715,471–916,617 | 39,392 | 261,021,128 |
  | spf13/cobra | 72 | 66–68 | 547,468–705,271 | 4,783 | 48,769,470 |
  | prometheus/prometheus | 2,002 | 1,113–1,534 | 23,091,052–29,361,867 | 2,738,159 | 50,385,864,014 |

  - Zero `.codegraph` entries were excluded.
  - The receipt holds 6,067 Git commands; the longest took 0.238 s.
  - R7-01's fixture shape is informed by TypeScript's maxima above and must say so.
- **R6-05 rule change: legitimacy.**
  - **Accepted in principle.** The round-6 max(file ratio, byte ratio) rule multiplied file-driven time by a ≈ 450× byte ratio under the council's fixed 2 GB assumption. Its ≈ 1,463-day result is an artefact of that method, not evidence of infeasibility.
  - **Not a goalpost move.** The replacement is pre-freeze, disclosed, physically motivated and uses all three successful same-PREP receipts, with none dropped. Both failed receipts are preserved, the ceilings are unchanged, and the fit refuses singular or negative designs (CAP:23–45). The gate is fail-closed, and a pass says `projected-feasible-not-authorised`.
- **Identification.** The design is not collinear: centred determinant 1.599 × 10¹⁴, and 0.99995 after scaling (challenger). All coefficients are positive: a = 0.001696478573 s/file, b = 2.867979797 × 10⁻⁹ s/B, c = 4.458760740 s/rev. The council reproduced each design point exactly:

  | Point | Measured s/rev | a·f + b·B + c |
  |---|---|---|
  | file-heavy | 2,803.8295/20 = 140.1915 | 135.7200 + 0.0127 + 4.4588 |
  | byte-heavy | 40.7692/3 = 13.5897 | 3.3946 + 5.7363 + 4.4588 |
  | construct | 13.5850/3 = 4.5283 | 0.0696 + 0.0000 + 4.4588 |

- **Worst case, recomputed.** For TypeScript, T = a·Σf + b·ΣB + c·n = 202,084.36 + 1,427.56 + 7,437.21 = **210,949.14 s**.
  - With the 1.5× margin: **316,423.71 s** per attempt (3.662 days) and **949,271.13 s** for three attempts (10.987 days). Both equal the receipt.
  - Headroom against the 14-day ceiling: 3.82× beyond the margined figure and 5.73× beyond the raw one. The real protection is that gap, not the 1.5× factor, and the margin is not a confidence bound (disclosed).
  - The decisive gap is R7-01.
- **Envelope receipt.** Accepted (R6-02 row). The run record `envelope-v7-run.json` gives argv, environment, exit 0, and the input e492da74…, preserved as `envelope-v7-inputs.json`.
- **Capacity receipt.** Accepted as distribution evidence.
  - Script 1599b49b…, receipt 141c6764…, run record f5087167….
  - Only `rev-list --first-parent --reverse` and `ls-tree -r -l -z` were run, with an explicit empty-allowlist Git environment and a 3,600 s timeout (CAP:86–103).
  - The scope matches production: Cold, C0, then the first ≤ 2000 positions (CAP:108–110, the same as PREP:485–487).
- **Model-free, no outcome leakage.** Confirmed:
  - all benchmarks are synthetic (SCALE7);
  - the collector is Git-metadata only;
  - no scanner, map, inventory ranking, label or model touched a candidate.
- **ADR-013 form against the adopted ADR-011 precedent.**
  - Status line: "accepted; decided by the agent under owner delegation (R-064)", identical in pattern.
  - Decided/sources metadata, heading and Review section name the folder's final zero-issue round, not a manifest.
  - Pending-owner state is kept only in the README row and STATUS l.128 trackers.
  - The card has status `accepted`, and its footer reads "effective on its final zero-issue council round; adoption recorded outside signed files", as ADR-011's does.
  - The index row and card are in the signed set, and ID 013 is sequential.
  - **Defects:** R7-02 (pending wording in the row and card), plus the R7-01 content that the R6-05 row must absorb.

## 4. Export-preparation authorisation

What rev 7 does well:
- a disclosed, supply-informed segment amendment that leaves Go eligibility unchanged;
- a reviewed, once-only envelope reproduction;
- enforced subprocess, attempt and whole-preparation deadlines with global interrupt stops;
- usable uniform label-prefix gates;
- trace-bound content refusals;
- honest dependence wording;
- separate stop reasons;
- an explicit HD8 map;
- complete ledger evidence;
- an ADR in adopted form;
- genuine final-PREP production-path benchmarks, with failed projections preserved;
- a candid, fail-closed proposal for the capacity rule.

Why it may not run now:
1. **R7-01.** The only pre-authorisation evidence that preparation fits its ceilings does not exercise the scanner path that dominates the worst case.
2. **R7-02.** Signing now would freeze a PREREG and an accepted ADR that call their own gate unapproved.
3. **R7-03.** The attempt-cap and global-stop commitments are unenforced.
4. PREREG freezes at this authority (l.35), so every issue must be cleared first.

## Verdicts

**(a) Export-preparation authorisation: `not authorised`.** Reviewed bytes, all copied from manifest `reviews/rev7-digests.json` (SHA-256 892b3ab79834b856d44b4e6b12bda384be8f1feb1927297349579e900e8786b4, computed by Main):

- `PREREGISTRATION.md` 97cdf755e148f65889184a2a250b6bbed592b68d93e7c1e67ce95b28f4679550
- `EXPORT_RECIPES.md` 6bb59e859aee9f727db5edca350fdb9f00db8eb9bff7f9c3e4b720d2294040e3
- `README.md` cce54fb35d5bcb8cdc3bf8f65ca77c6a95d6fb2c3448be4e35ecd46a7dc2764e
- `prepare_exports.py` d52135b07b13100fb0f0092de759c9bd17b06821d5c92eb5a4fa1d0e3c8e0e8d
- `assemble_preparation.py` b8b6a04899651058b1b61b65b2178dfdf8eff1019bb40ee3a124b3d98f3d51a0
- `prep_selftest.py` 72946bfdf9e997df659d4fe0f6fd5c73502a2d41d29e1999c1c9fef1cba894d7
- `selftest.py` b7d1ef33f5051cac08218c8c60955600082075e50d335b7d8cf4693a58c6740a
- `pin_tool_closure.py` 39dff5b8a75e41852af7d6f5f534728ed0667626e1eb42e8ab38e657109e8348
- `bootstrap_v6.py` 5b8735ad1c03ca0afebf54dfe3f1123a96dabb72002db53de16f60d2ae0244e9
- `bootstrap_selftest_v6.py` a9901432ace306ce867aa1e09ef617a24a4e8b6347359e86415a831784978ff7
- `stages.py` e58dc239e13615c989410502afe47a8ef64f84a1e810bd6749551c654cb06a02
- `inventory.py` 88cdd3b1a76b60c7394f43df58fe52a0573f671f7d3c12c15f2f75ea2f1db9cc
- `extract_exposures.py` 966f7ed55d0a30c3d5cc4e2deea5c78b912f98827e261bbdb9b265c44749c6b8
- `frontend/main.go` 8bbe00b526dd39ef0905fb85e47d85326418e57c062ebb241d8fd55770c4c6f9
- `frontend/main_test.go` 1b49b60b85568ef8d3b440d12590335e5a9e17eb7b9997514af9827b833cf3c0
- `inventory-inputs.json` e958f0f1ba876a3ba83ab8c15df7825c9b5b06d86eee1a7e6322988492bbd6cc
- `envelope_check.py` 00b6bf253f5ea2547854f228526fff34535084e008874725540769b9600bd1f5
- `envelope-v7.json` f8ebcdcb8aad65d1b288433a1180ae12a36fa7ff3cfb2b088c526a7dcb89d5ca
- `envelope-v7-run.json` 30aa18632d5e31069ad9dbcb1a85bd62ef321b4c2917abc90ea25d0e2fecd6b0
- `envelope-v7-inputs.json` e492da740af2aa778304394dd326c269099e7e7aa30a9207e680c92b0f34c74b
- `envelope-v7.stderr` e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
- `preparation_capacity_v7.py` 1599b49bbb41b48454bc8394fdb1dc6398d760f1957a69250f0c1fe9b561453b
- `preparation-capacity-v7.json` 141c6764dee809dd0bb198d53e1a414c2f744b1403d9d99c1e282af530c8e037
- `preparation-capacity-v7-run.json` f50871675e3295f8e2522ab4b73ec8b54e8e885ab2d35f0c0c5f3108d0a2c42f
- `prep_scale_v7.py` 956624bb7e26abe215c8e75b3ea0e93069de443867dcb3256a5ef083e3ad6ed0
- `prep-scale-files-v7.json` 4aa468fd5dd4c96ff815a24d48fd17d175b4ea8fea3bd2ba9b20a16d671de2b2
- `prep-scale-bytes-v7.json` fd2fa231a08d8d55353088a2c9712eb129810f4e6b594ef00d88749e59f11cd2
- `prep-scale-construct-v7-attempt2.json` 19fc7fd814a43329c0d92b72de43f5b449e7b6a3632f387404793592bfb4d8b5
- `prep-scale-construct-v7-attempt1.json` b3c55ccbb318c337ba54e08b10a0df988511b92b7c5423bb7cc19c98ee1a5cc7
- `oc_v5.go` 34e51550f3f0a4d7d406e4302200af9a1f7302b33186b97c1734450373bb4966
- `oc_v5.py` 5886200086055b9e403a823f465e6342981466d229f660fed8f55a8abfefcbcf
- `oc_selftest_v6.py` 0eba0a6330c8ca7bc52819684e94dd6b57387960f4fe757a03d71e35438789b4
- `oc-v5-spec.json` 7f0f2f83887634b0ca303e189e5f9c8ec4505e5b06a646bfbb10b2c4b08999b4
- `oc-v5-fixtures.json` b5c2a1e6cd35284b633489f26e70db3d5c65c7ce145abdb280853c950609137e
- `oc_timing_v7.py` ff283ad8d923ada9d282b8d777b3f69750f2f4c5bb749aae87171199272e0a9f
- `oc-v7-timing-metadata.json` c81e284bc6d3309d55a2f4a8eafaf2b2c9b84f44d052182a205f23248ed075d5
- `oc-v7-timing-balanced.json` 78ea75275753eff946a82ec15620b395137cfba62e410dea9582e90c7b6bacb2
- `oc-v7-timing-balanced-maximum.json` 29ca943bc69e130d3344f459fab91f7f680137a14bffd993ddc1ade9988d4de9
- `oc-v7-timing-concentrated.json` b82d33928cb6f114c3805e65ffed19ff9b5ec8f079912cd6fbedb041e0715afa
- `oc-v7-timing-concentrated-maximum.json` bdd5d4f275beff62f1b2b8776054f1cf1c87df4554ed9892a95f60a5a5fe0643
- `oc-v7-timing-singleton.json` e0d0723b6576a08694d324a862fc00d0b5956e067109d020ec07d2be24d6745e
- `oc-v7-timing-singleton-maximum.json` eae42ae1fc9a55ecf94ef26daba6478704315a2b4bac6f61398b33253a54a47d
- `oc-v7-timing-go-closure.json` 2df0ec6ded049c4d1a30650804aac78aa089a1c3f07ce9bc9fa7d87b6000a16e
- `tool-pins-v7/frontend-build-recipe-v7.json` 8b149edf4a67ef9797d8acc9361414fb1dfa8529e510ae45f72bc2eb29c33426
- `tool-pins-v7/frontend-go-closure-v7.json` 788f2a3be52a471602a4d622c90dae88b52d79a5bd8b8c1d9383c2ae42a22e37
- `tool-pins-v7/oc-build-recipe-v7.json` e11bad826e142715036fdf66dbc166b672be0b731f8a1be2e8cb035ab2197c81
- `exposure-source-index.json` bccbfd4b8a2e42c9a91cdd9ead0c3f939c5606950a5fcb9844f8d4b6a141c3d8
- `exposure-contexts-v4.json` 2dd5b6b7d6665d64519585acdb63ed6eafadc3092bda25a38433616f896199fb
- `latency-basis-v4.json` bdc3cef935d65a20b4b0b9f34bb68c881d29da4960e74e5ee5b60549f0894104
- `cost-feasibility-v5.json` 1637256f02d54164c73eff82abb8648b4f70f4924701c9c29ee5a54e6c76c1d0
- `exposure-extraction-v4/seal.json` 96fcfcbfada87faaa6624d9124fb3c88d55d714a8cd0bf79ce72a114e32f8510, plus the five v4 outputs as pinned in it
- `reviews/round-6-response.md` 9e6af4631e56c32ecfee445e0433e406f2c1e9de9e7816fa4ac9898d5924393d
- `reviews/rev7-checks.json` 7c55e546dd815f188ec11fd0155316f50d63ddac1f3484598372510c4552d7ed
- `reviews/rev7-metadata-extraction.json` e23c21b6c2f5427ecd8a7fb60ed9a2fdf5a0c1d46a003cf39ee783ad588c9a40
- `reviews/round-6.md` 7746296a149ef67bddf3aa4671797117013379d50b832510790d5313ffb3368b
- `../x11b-guard-recovery/PLAN.md` 034284255f303ad09308f30fc6212efac6d2730aa50ac724c4a8243df3d48753
- `docs/research-ledger.md` ddb30253b89ee2386c26035217ae9d6ff362e6ccd953371024282dc2b0387005
- `docs/hypothesis-ledger.md` 1eae27b9c2e8f39c98c835fad0ab06a8267b15fecf1b6b7c13d3f2216518c7ea
- `docs/STATUS.md` 378f0a2a9fc60d5c893b6e2e2b01a9331bc0384ed439bb21af0a410138f6eaf6
- `docs/adrs/README.md` 10599a4a97e7b8dc18dab3d9d305747c756c97f5566abadeceed5af90f3aa3e2
- `docs/scout-design.html` ca8c2f9a8332bf6cb60cc7dd9545d3dcba8e5266072cbb04064b7c9f4a63c864
- `docs/adrs/2026-11-08-x11a-delegated-decisions/ADR.md` 9ccc106e5419a0d0c8e37de59f8110b6079c1c9e325b7636131aadd195ccc03a
- `docs/adrs/2026-11-08-x11a-delegated-decisions/round-1.md` 7746296a149ef67bddf3aa4671797117013379d50b832510790d5313ffb3368b

None of the following may run: ownership scan, frontend build, exposure operation, inventory, OC run, map, label, replay or model call. Inventory still needs its own artifact-binding round after a future authorised preparation.

**(b) Overall: `4 issues open`.** That is R1-01, carried and deferred until the implementation round, plus 3 new: R7-01 major, and R7-02 and R7-03 minor.
- The study is not signed off.
- HD8 stays `proposed`.
- This round changes no ledger, design or product file.

**ADR-013 verdict: `not signed off`** (ADR.md 9ccc106e5419a0d0c8e37de59f8110b6079c1c9e325b7636131aadd195ccc03a; index 10599a4a…; card in `scout-design.html` ca8c2f9a…).
- Its own issue is R7-02, minor: the pending wording in the final decision row and card bullet 4.
- Its R6-05 decision row must also absorb the R7-01 rule.
- The form otherwise matches ADR-011 (§3), and the round numbering is ratified (§1, R6-11).
- This record is saved byte-identically as ADR-013 `round-2.md`.
- Owner review stays pending in the trackers only.

## Council synthesis

### 1. Recommendation

Revise to rev 8 and request export-preparation round 8, with ADR-013 round 3 in the same packet:
- Keep every item §1 marks closed.
- Accept the additive capacity rule in principle, and close its one unbounded gap with a held-out TypeScript-language point that bounds TypeScript directly (R7-01).
- Enforce canonical create-once operation paths, then rerun the three fit benchmarks on the final PREP (R7-03).
- Rewrite the signed texts in adopted form (R7-02).

### 2. Alternatives

- **A. Recommended: held-out TypeScript-language point with a direct bound.** It tests the only path that could break the gate, costs well under an hour of synthetic compute [INFERENCE], and stays fail-closed with no refit. Because it bounds with max(observed, predicted), a model miss within the deadline headroom does not kill a feasible study.
- **B. Relabel the projection as advisory; the runtime deadlines are the only protection** (peer's alternative). This is honest and cheapest. But it authorises a possible 42-day burn with no evidence for the decisive path, and so abandons R6-05's purpose. Not recommended.
- **C. Authorise now on the additive model, with the language path disclosed as a threat.** The deadlines fail closed and there is 3.8× headroom. Rejected: the gap is cheap to close, and PREREG freezes at authorisation.
- **D. Add a fail-fast runtime check inside `execute` that stops after k revisions on a projected overrun.** This adds an adaptive stop rule to signed code and needs its own validation. More complex than A.
- **E. Revert to the max-ratio rule.** That makes X11a infeasible (≈ 1,463 days) because of a flaw in the method. Rejected.
- **F. Time a real candidate prefix.** That is a candidate ownership scan before the freeze. Rejected.
- **R7-03 variant: enforce only in the assembler.** Trial runs at ad-hoc paths would remain invisible. Rejected.

### 3. Evidence

- **Capacity:**
  - receipts `prep-scale-{construct-v7-attempt2,files-v7,bytes-v7}.json` (operations, fixture setup, build, elapsed);
  - `preparation-capacity-v7.json` (model, size facts, projections);
  - CAP:23–56, 69–110, 130–146;
  - SCALE7:96–104 (Go-only fixture), :199–211 (old rule).
- **Scanner path:** CP `mapper.ts` `collectSeeds`/`shouldRunNodeMappers` and `findNearbyTests`; `mappers/node.ts`:214–215, 454–463.
- **Canonical texts:** PREREG l.35, l.48; REC:122; ADR final row; `scout-design.html`:1013; `research-ledger.md`:79.
- **Enforcement:** PREP:52–68, 204–246, 860–931, 933–950; ASM:10–33; REC:54, 61, 68, 71.
- **Round-6 closures:** PREREG l.41–48, 64, 161, 225–231; `inventory.py`:548; OCGO:165, 177–182, 293, 347–350; PREP:637–663; `research-ledger.md`:75–82; ADR:1–5; README row :28; card :1006–1023; round-5.md:63, :231.
- **Arithmetic:** §3, reproduced independently by both peers.

### 4. Risks

- **The synthetic TypeScript fixture can still miss real-content paths** (route parsers, workspace manifests, React heuristics) [INFERENCE]. The runtime deadlines remain the backstop, and the result would be an operational not-testable outcome, never a content exclusion.
- **A failing held-out point ends "preparation not feasible".** X11a would then be not testable, with no refit. This is accepted as the honest outcome.
- **The fixture shape is informed by TypeScript's sizes.** Only sizes are used, and they are already disclosed (§3); no content or outcome is involved.
- **Rerunning the fit benchmarks** on the R7-03 PREP changes the coefficients. This is a new measurement under a reviewed rule, not tuning. The v7 receipts stay preserved.
- **Reversibility:** everything is reversible. No candidate scan, build, inventory or model call has run.

### 5. Execution outline

1. Main saves this record as `reviews/round-7.md` and byte-identically as `docs/adrs/2026-11-08-x11a-delegated-decisions/round-2.md`.
2. **R7-03:**
   - Implement the canonical create-once seal/output paths in `execute`, and the canonical seven-seal set in the assembler.
   - Update the REC commands.
   - Check: fixtures for a second invocation, a non-canonical path and an omitted interrupt seal all refuse.
3. **R7-01 implementation:**
   - Add a versioned runner that preserves `prep_scale_v7.py` and includes the TypeScript-language scenario.
   - Add a versioned projection script that reads the size facts from 141c6764… without touching Git, and implements the direct-bound rule and unchanged ceilings.
   - Freeze the rule in PREREG §3, the ADR row and the card.
   - Check: plants for every failing branch and for the status string.
4. **Main, once each, synthetic only, on the final PREP:**
   - Rerun construct, file-heavy and byte-heavy.
   - Compute the coefficients and write the TypeScript-point prediction to a create-once file (digest cited).
   - Run the held-out TypeScript point, then the projection script.
   - Retain every receipt whatever the outcome.
   - Check: either "projected-feasible" with max(S_obs, prediction) ≤ 483.45 s, or a recorded "preparation not feasible".
5. **R7-02:** apply the adopted-form wording, and complete R-079 with the coefficients and both receipts' results.
6. Main recomputes the rev-8 manifest and runs the pinned isolated selftest and the PGO-off frontend Go tests once. Main records the exercised results and requests round 8: export preparation plus ADR-013 round 3.

### Where the peers changed this memo

- **innovation-challenger:**
  - Independently judged the additive model identified but unvalidated, and recommended a held-out combined-load fixture. This was merged into R7-01, and its major severity was adopted.
  - Supplied the determinant check, the convex-hull and barycentric extrapolation counts, and an independent reproduction of 316,423.71 s and 949,271.13 s.
  - Supplied the `bytes_per_revision_min` mismatch, the fixed-cost confound and the quadratic verifier step. All three were checked and not raised (§2).
  - Its OC timing-variance point was not adopted, because those profiles are not preparation replicates.
- **experiment-peer:**
  - Supplied the Node-mapper dispatch evidence and the per-file break-even figures (10.1 and 6.7 ms). These confirmed the lead's independent finding of the language-path gap.
  - Supplied the alternative-model bounds: the f·B interaction ≤ +72 s/rev and the ownership-only refit of 117.9 s/rev.
  - Found the pending wording in the signed texts independently of the lead (R7-02). The council graded it minor, not major, because the fix is trivial and form-only, as R6-11 was.
  - Identified the ADR numbering departure. The council traced it to its own round-6 error and ratified the drafter's mapping instead of raising it.
  - Its capacity-plant gap was folded into R7-01, because the council hand-verified the existing booleans. Its signed-set breadth and index-token points were not raised (§2). Its list of rule-3 size facts was adopted into §3.
- **Lead's own additions:**
  - Found the language-path gap from the Clawpatch mapper sources and the Go-only fixture.
  - Designed the direct-bound held-out rule: fail-closed, but without killing a feasible study on a model miss.
  - Found the re-invocation and seal-omission gap (R7-03), and the sequencing consequence that PREP changes require the fit benchmarks to be rerun on the final PREP, with the pinned size facts reused and no new candidate access.
  - Checked that the anchors in e492da74… and e958f0f1… are identical.
  - Checked that a fully serial schedule still fits.
  - Recorded the R6-05 rule change as accepted in principle.
