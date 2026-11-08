# Council round 8: X11a merge-time association accuracy, preregistration rev 8, export-preparation scope (also ADR-013 round 3)

| Field | Value |
|---|---|
| Study | X11a merge-time association accuracy (HD8). This is the accuracy part of REQ-176/REQ-177. Guard and recovery evidence belongs to X11b (`../x11b-guard-recovery/PLAN.md`), which is a plan only and is not preregistered. |
| Digest manifest | `reviews/rev8-digests.json` has 1,978 entries and excludes itself. Main computed its SHA-256 as **056ddf7f428364230020800e2e4c36d4f8d2a2643640c91664b93686ba9cd135** and verified all 1,978 entries before and after the independent checks. The council has no hashing tool. Every digest below is copied from the manifest; none was recomputed. |
| Base | Uncommitted worktree `scout-x11`, integrated onto origin/main `584da4a2`. That base includes ADR-012 (PR #19) and X1/ADR-014 (PR #18). |
| Requested scope | **Export preparation only.** This covers the model-free Clawpatch heuristic ownership scans at the pinned commits, the council-gated frontend build and the exposure-binding operation, which together produce sealed ownership exports. Inventory, OC runs, maps, labels, replay and model calls are out of scope. ADR-013 is reviewed in the same packet as its round 3. |
| Owner decisions in force | R-073: "(b) High-none only on selected repo" (alternative D). R-064 is a standing delegation: wherever an owner choice arises, the recommended option is stated as the required fix. |
| Independent checks (Main and independent agents, not the drafter) | **Report:** `/var/folders/g5/r809r2p97nz9gtj25s1c37t80000gn/T/x11-rev8-checks-jpxe1glv/report.txt`.<br>• **Integrated selftest:** 30 outer, 28 preparation and 6 capacity tests OK (156.5 s).<br>• **Pinned frontend Go package:** `ok`.<br>• **Digests:** 1,978/1,978 matched before and after; git status unchanged.<br>• **R7-01 held-out TypeScript run:** Main ran it once, using the drafter's recorded argv (`reviews/rev8-heldout-request.json` e196f8af…), after the prediction was frozen. Exit 0, empty stderr, wrapper wall 365.607 s. The output `prep-scale-heldout-ts-v8.json` c47075f9… did not exist before the run.<br>• **Integration:** the five shared files show additions only against origin/main (33 added, 0 deleted). 19 of X1's 20 signed extracts match; the 20th, HD9's row, differs only by the authorised X1 round-9 §13 condition 2(a) change. Research IDs R-001..R-086 are unique; the next free ID is R-087. |
| Date | 2026-11-08 |
| Reviewer | innovation-council (anthropic/claude-opus-5-5) |
| Independent peers | innovation-challenger (openai-codex/gpt-6-astra:high) and experiment-peer (anthropic/claude-opus-5-5:high). Neither was told the lead's view. |
| Method | Read-only. Nobody edited a file or ran a script, test, build, git command or model, and nobody touched the candidate clones.<br>• Computations were limited to jq and grep over study files, plus hand arithmetic.<br>• Clawpatch mapper behaviour was read in the `third_party/clawpatch` submodule. That the pinned built scanner matches that v0.7.3 source is still [INFERENCE] (standing caveat). The held-out ownership counts match what the source predicts exactly, which supports it (§3). |
| Line references | `PREREGISTRATION.md` (PREREG), `prepare_exports.py` (PREP), `assemble_preparation.py` (ASM), `stages.py` (STG), `prep_selftest.py` (PST), `prep_scale_v8.py` (SCALE8), `preparation_capacity_v8.py` (CAP8), `capacity_selftest_v8.py` (CST), `EXPORT_RECIPES.md` (REC), ADR-013 `ADR.md` (ADR), `third_party/clawpatch/src` (CP), `reviews/round-7-response.md` (RESP). All refer to the bytes pinned in the manifest. SCALE8 and CST line numbers marked "(peer)" were read by a peer and spot-checked through the receipts, not line by line by the lead. |

## 1. Status of prior items

"Closed" means the round-7 required fix (round-7.md l.44–46) is present in the rev-8 files. Each status was checked against the files, not taken from RESP.

| ID | Status | Evidence / remaining gap |
|---|---|---|
| R1-01 | **Open (carried; deferred as planned)** | These are still absent: the full GUIDE, real orchestration, calibration, replay, labels, scorer, stage seals, whole-pipeline plants and inventory-conditioned OC tables. PREREG l.35–37 discloses them, and `stages.py` lists only `inventory` as implemented, so later stages refuse (STG:11).<br>**It does not block export-preparation authorisation.**<br>• Preparation is model-free and outcome-free.<br>• Every stage that R1-01 governs refuses in code.<br>• The R1-01 completions PREREG permits (l.37) cannot change the preparation bytes this round freezes.<br>It stays open until the implementation round. |
| R7-01 | **Closed** | Each element of the required fix, in order:<br>**Fixture.** Per revision: 81,402 files and 395,061,604 B, against required minima of 81,369 and 380,592,869. The 81,400 `.ts` files include 40,700 tests, split 20,350 under `tests/` and 20,350 sibling `*.test.ts`; the other 2 files are `package.json` (with a test script) and `tsconfig.json`. The tree is nested at least three levels, with imports, exports and functions. Four code files are ≥ 1,050,000 B, normal files are ≥ 4,800 B, and there is no padding comment (`prep-scale-heldout-ts-prediction-v8.json`:13–62; receipt :39–77; SCALE8:28–40, 243–247, 285–292 (peer)).<br>**Run.** 3 revisions, with 4,070 changed sources after the base (5%). It ran on the final PREP 77446e97… through the production execute → seal verifier → built frontend. CAP8:89–91 refuses a receipt pinning any other PREP.<br>**Main ran it once and the receipt is retained** (run record `prep-scale-heldout-ts-v8-run.json`, `runner: "Main"`).<br>**Freeze first.** The prediction was created at 19:21:47Z and the run started at 19:22:22Z (prediction :11; receipt :795). The prediction digest 4b993052… was cited in the pre-run request. The runner refuses on prediction drift or a mismatch in fixture, source or tree. CAP8:79–84 refuses unless the prediction's observations and exact coefficients equal its own fit.<br>**Rule.** CAP8:52–59, 96–101 implement 1.5 × 1668 × max(S_obs, prediction) for one attempt, three attempts and the subprocess proxy, with unchanged ceilings. The status string is at CAP8:62–64.<br>**Plants.** CST plants each gate failing alone, the linked boundary and the aggregate failure (Main's run: 6 capacity tests OK).<br>**Size facts.** Reused from 141c6764… (CAP8:91–93), with no new candidate metadata.<br>**Result under the frozen rule:** pass (§3). |
| R7-02 | **Closed** | The canonical texts now state the final rule and the recorded result in adopted form:<br>• PREREG l.50 (gate), l.52 (held-out bound), l.54 (recorded result);<br>• REC:124–126;<br>• ADR:73 (R6-05/R7-01 row) and :74 (R7-03 row);<br>• the ADR-013 card, `scout-design.html`:1021–1023.<br>No canonical text calls the rule proposed or pending. The only grep hits are unrelated ("council micro-round", PREREG l.46; "council-reviewed deviation", ADR:74).<br>R-079 (`research-ledger.md`:88) holds the final coefficients, the three v8 Go elapsed times, the frozen prediction, the held-out result and ratio, the direct-bound projections, the gate status, the clock caveat and the ownership counts.<br>Pending authority stays only in the draft header (PREREG l.3) and the trackers (index row :33, STATUS l.135), as the fix allowed. |
| R7-03 | **Closed** | **Canonical paths.** `canonical_paths` and `canonical_seals` derive create-once `artifacts/preparation-v8/<operation>[--owner__repo].{output,seal}.json` from the canonical budget directory (PREP:31–54).<br>**Execute** rejects non-canonical output or seal paths before any command (PREP:981–983). Under the budget lock it calls `require_no_global_stop`, refuses an existing output, and opens the seal with `'x'` (PREP:994–998).<br>**Assembler.** It discovers exactly the configured ownership seals plus frontend and exposures; the `--seal` list is gone. It refuses an interrupt or budget stop, a duplicate, a missing seal or a non-canonical seal (ASM:11–41).<br>**Relocation.** Moving `inventory-inputs.json` to get a fresh canonical directory fails, because the authority must cover the *resolved* inputs path (STG:60–61). Only a new authority artefact could do it (§2, considered).<br>**Fixtures:**<br>• alternate paths refuse (PST:306–318);<br>• a repeat invocation after a terminal operational refusal raises `FileExistsError` (PST:337–341);<br>• after an interrupt, both assembly and any further operation refuse, and the old list API is gone (PST:508–545);<br>• budget exhaustion behaves the same way (PST:547–570).<br>**Benchmarks.** The construct, file-heavy and byte-heavy points were rerun on final PREP 77446e97…, and the v7 receipts are preserved (§3). |

**Status count:**
- Open: R1-01 (deferred; does not block this scope).
- Closed: R7-01, R7-02, R7-03.

## 2. New issues

**None.** No critical, major or minor issue was found in rev 8 or in its integration onto main. All three reviewers reached this independently.

**Considered and not raised:**

- **The 200-candidate cap: round 7's 50/50 rationale was wrong, but the held-out result still stands** (challenger, peer).
  - CP `mappers/node.ts`:424–431 `packageTestFiles` walks the source roots and `test/tests/__tests__`, keeps test paths, then applies `.slice(0, 200)`. `associatedTests` (:454–464) filters those ≤ 200 candidates per source group and keeps at most 8. Source groups set `skipNearbyTests: true`, so `findNearbyTests` does not run for them (CP `mapper.ts`:108–117; `node.ts`:196–257).
  - Round 7's premise, that a 50/50 split maximises groups × tests, therefore does not hold. Per-group association cost is capped at 200 candidates.
  - **The fixture still reaches that cap.** The walk is sorted, and `src/` sorts before `tests/`, so about 3,392 groups each filter 200 candidates (peer) [INFERENCE: source reading]. No repository can exceed 200 candidates per group.
  - **Why ownership volume does not matter:**
    - The held-out ownership operation took 344.2347 s / 3 = 114.745 s per revision, 1.410 ms per tree file, owning 40,700 files.
    - The v8 Go file-heavy point took 2,288.8379 s / 20 = 114.442 s per revision, 1.431 ms per tree file, owning all 80,000 sources (receipts :128 and :136).
    - Per-revision cost follows tree files (materialisation and walk), not owned count. An all-source composition would therefore not be expected to approach the 483.45 s bound [INFERENCE].
  - The cap and the zero-test-ownership construct were recorded in the frozen prediction **before** the run (prediction :13, :44–50). Nothing was relaxed afterwards.
  - The result must not be described as an exhaustive worst-case bound. PREREG l.52 already names synthetic complexity as a threat.
- **Zero owned tests.** Node source groups own only reviewable non-test files under the source roots (`node.ts`:202–215, 559–567), plus `package.json` and `tsconfig.json`.
  - The observed counts were 40,700 / 0 / 40,702 for all three revisions (receipt :560–574). That is exactly the prediction from the source.
  - The R7-01 floors apply to tree files, which the model's `files` term also counts. They are met.
  - The runner's own ownership assertion is weak: it only checks `source_owned > 0` (SCALE8:371, peer). That has no effect, because the observed value is all 40,700 planted sources.
- **Clocks.**
  - S_obs uses the receipt's charged elapsed time, 419.9873540849658 s. That is setup 46.081 s + freeze 9.647 s + timed run 364.259 s. The definition was in the frozen runner before the run (SCALE8:153–154, 344–345, peer; run record `clock`).
  - It matches the Go points' "whole elapsed", which also includes fixture generation, and it is the conservative choice.
  - Main's 365.607 s wrapper wall time covers the run phase only. The 1.348 s difference is launch overhead [INFERENCE].
  - Under every alternative accounting the ratio stays ≤ 1.107 and S_bound ≤ 139.9958 s:

    | Accounting | s/revision | Ratio to prediction |
    |---|---:|---:|
    | All charged time | 140.00 | 1.107 |
    | Excluding setup | 124.64 | 0.985 |
    | Excluding setup and freeze | 121.42 | 0.960 |
    | Ownership operation only | 114.74 | 0.907 |

    The decision is the same in every case.
- **Who ran what.** The drafter ran the three Go reruns and the capacity script; Main ran only the held-out point.
  - R7-01 named Main only for the held-out run.
  - Favourable Go coefficients cannot pull TypeScript's S_bound below S_obs.
  - The other repositories have ≥ 27× headroom.
  - One v8 receipt per scenario is pinned. The council cannot rule out undisclosed attempts [INFERENCE], but they could not change any gate.
- **Inputs digest 06676fb3… → 18ef246b….** The selftest fixture fix (RESP "Exercised checks") changed only `prep_selftest.py` and its pin.
  - The peer compared every `recipe_sources` pin against the preserved `prep-scale-heldout-ts-inputs-v8.json`; `prep_selftest.py` is the only difference.
  - PREP is unchanged, and CAP8 checks the PREP digest, not the inputs digest. No benchmark depends on the selftest bytes.
  - The initial failed output is preserved (`rev8-checks.json`).
- **Relocating through a second authority.** The only remaining way to get a fresh canonical directory is to write a new authority artefact that covers a copied inputs file. That is forging the council authority, not a silent retry.
  - Condition 3 below fixes the authority as a single Main act.
  - The inventory artifact-binding round can check that the assembled manifest's `preparation_authorization` pin is that artefact.
- **Unplanted override.** The rule "held-out status not completed → not feasible" (CAP8:104–105) has no plant. R7-01 did not require one, and the run completed (peer).
- **Byte term on loose objects versus real packs.** The byte term is 1.15 s per revision. Breaking the 14-day gate would need about 0.9 µs/B (≈ 1.1 MB/s) of extra cost (peer) [INFERENCE: implausible for `cat-file`]. The runtime deadlines remain the backstop.
- **Directory fan-out.** The fixture has at most 100 files per directory. Real TypeScript's very large directories most likely sit outside the Node source roots [INFERENCE]. This falls under the named synthetic-complexity threat.
- **PREREG draft header** (l.3: "DRAFT; not signed off… take effect, unedited, only on the authorising zero-issue council round").
  - This is the accepted route-(a) form; round 7 accepted it.
  - The study remains not signed off overall (R1-01). This record holds the authority, so the header becomes neither false nor binding.
- **STATUS ADR range** (l.7, l.131: "ADR-000…ADR-012"). These are main's bytes. They are accurate, because neither ADR-013 nor ADR-014 was effective at that base. ADR-013 is tracked at l.135 and ADR-014 at l.11 and l.132. Tracker matter only.
- **Index.** There is no route paragraph for ADR-013 like the ADR-011/012/014 paragraphs (`docs/adrs/README.md`:12–16). Its row (:33) carries the same meaning. Wording preference.
- **Integration.** Additions only (33 added, 0 deleted). The ADR-012, ADR-013 and ADR-014 index rows and cards coexist (index :32–34). HD8 is still `proposed` and HD9 `preregistered`, as on main (`hypothesis-ledger.md`:35–36, :40–41). The single X1 extract difference is the authorised HD9 adoption. Study `round-7.md` and ADR `round-2.md` are byte-identical (795e4525…).
- **Accepted risks, not reopened** (no new evidence):
  - real-content Node paths (routes, workspaces, React);
  - Django's unexercised Python mapper;
  - disk use;
  - exposure-binding timing;
  - one run per point, with no fitted error estimate;
  - the signed-set snapshot breadth (round 7 §2).

## 3. Review of the specific rev-8 elements

- **Final-PREP Go refit.** All three points ran on PREP 77446e97…:
  - construct: 13.361451209 s / 3;
  - file-heavy: 2,465.574174167 s / 20;
  - byte-heavy: 39.494037875 s / 3.

  Exact fit: a = 0.001485893272630103 s/file, b = 2.899084947714978 × 10⁻⁹ s/B, c = 4.392888931244594 s/rev. All are positive. Compared with v7, a fell 12.4% and c fell 1.5%. Reproduction:

  | Point | Measured s/rev | a·f + b·B + c |
  |---|---:|---|
  | construct | 4.45382 | 0.06092 + 0.00001 + 4.39289 = 4.45382 |
  | file-heavy | 123.27871 | 118.87295 + 0.01287 + 4.39289 = 123.27871 |
  | byte-heavy | 13.16468 | 2.97327 + 5.79852 + 4.39289 = 13.16468 |

- **Frozen prediction.** a·81,402 + b·395,061,604 + c = 120.95469 + 1.14532 + 4.39289 = **126.49289 s/rev**. It equals the frozen 126.49289025945677.

- **Held-out comparison under the frozen rule (R7-01, round-7.md l.44):**

  | Quantity | Value | Limit | Pass |
  |---|---:|---:|---|
  | S_obs = 419.9873540849658 / 3 | 139.9957846949886 s/rev | — | — |
  | S_obs / prediction (descriptive) | 1.1067482481255293 | 1.5 margin (descriptive) | within |
  | S_bound = max(S_obs, prediction) | 139.9957846949886 | 483.45 (= 1,209,600 / 2,502) | ✓ (3.45× headroom) |
  | 1.5 × S_bound (subprocess proxy) | 209.9936770424829 s | 3,600 s | ✓ |
  | 1.5 × 1668 × S_bound, one attempt | 350,269.45330686145 s (4.054 d) | 14 d | ✓ |
  | Three attempts | 1,050,808.3599205844 s (12.162 d) | 42 d | ✓ |

  **R7-01 is satisfied under the frozen comparison rule.** There was no refit, rerun, fixture change or ceiling change. The direct bound is about 26% above the additive TypeScript sum (≈ 185.8 k s unmargined), so the substitution is conservative.

- **The other repositories** keep the additive model. Hand checks:
  - Django unmargined ≈ 20,485.4 + 259.4 + 8,794.6 = 29,539.5 s, against the receipt's 29,539.61.
  - Margined single attempts: Django 44,309.41, gin 2,184.34, cobra 485.30, Prometheus 19,513.87 s. All are far inside the ceilings.

- **Fully serial execution of all five:**
  - one margined attempt each: 416,762.37 s = 4.82 days;
  - three attempts each: 1,250,287.15 s = 14.47 days, under 42 days.

- **Experiment rule 3.** No predictor–outcome relationship was computed. The new measurements are all synthetic, and the size facts are the pinned pre-freeze Git metadata (141c6764…). No candidate was touched (CAP8:91–93; README l.15).

- **ADR-013 form.**
  - The status line, metadata, heading, Review section and card footer follow the ADR-011 precedent, as in round 7 §3.
  - Rows :73–74 add the R7-01/R7-03 choices, sources R7-01/R7-03, R-079 and R-084.
  - Card bullets 4–6 match PREREG l.50–54.
  - Round numbering: ADR round N = study round N+5, so this record is ADR `round-3.md`.

## 4. Export-preparation authorisation

What rev 8 does well:
- the council's recommended held-out direct bound, implemented with a create-once pre-run freeze, digest-chained to Main's once-only run and to the capacity script;
- honest clock accounting, plus disclosure of the cap and the ownership construct before the run;
- canonical create-once operation paths and auto-discovering assembly, which close the omission and fresh-path routes;
- adopted-form canonical texts and a complete R-079;
- clean additive integration onto main, with X1's signed extracts preserved.

PREREG freezes at this authority (l.35). The council finds no open issue that blocks it.

## Verdicts

**(a) Export-preparation authorisation: `authorised`**, under the conditions in the sign-off statement.

**(b) Overall: `1 issue open`.** The open issue is R1-01: carried and deferred by design until the implementation round. **It does not block export-preparation authorisation.**
- Rev 8 has zero new issues.
- The study as a whole is not signed off, and HD8 stays `proposed`. The ledger mapping at `hypothesis-ledger.md`:40 moves it to `preregistered` only at final study sign-off.

**ADR-013 verdict: `zero issues — signed off`.**
- ADR.md 2f273489c1e15f82af4a2433dcbfbb38ecde8a012bf5073fad12befd3898814c; index d15b4d89…; card in `scout-design.html` 9a54215d….
- This record is the folder's final round and is saved byte-identically as `docs/adrs/2026-11-08-x11a-delegated-decisions/round-3.md`.
- Under route (a), ADR-013 takes effect when this record is saved. Adoption is recorded outside the signed set.
- Owner review stays pending in the trackers (R-064).
- R1-01 is a study implementation deferral that ADR-013 itself records (ADR "Execute missing full experiment machinery early?" row). It is not an ADR-013 defect.

## Sign-off statement

**X11a export preparation: zero new issues; export preparation authorised. ADR-013: zero issues; signed off.**

The submission is:
- worktree `scout-x11`, with uncommitted changes on base origin/main `584da4a2`;
- reviewed preregistration `PREREGISTRATION.md` SHA-256 `9b4be6ef008c057e7891865fafeb1fbf6b0246dde0b7a0ab6552e9312378a082`;
- signed set: exactly the 1,978 files at the SHA-256s in `reviews/rev8-digests.json`, manifest SHA-256 `056ddf7f428364230020800e2e4c36d4f8d2a2643640c91664b93686ba9cd135`, computed and verified by Main.

The principal reviewed bytes, copied from that manifest:

**Protocol and recipes**
- `PREREGISTRATION.md` 9b4be6ef008c057e7891865fafeb1fbf6b0246dde0b7a0ab6552e9312378a082
- `EXPORT_RECIPES.md` 74f01a4df6fff3cd5cb83e51228d851b346043238ae20b6f1b9dd5904f0fbabc
- `README.md` bc54a21d11a9a20d57900a9564e8c2cfbafcd9ced844e53674c59d97a5077e50
- `inventory-inputs.json` 18ef246b0eb10e5ead090698dfe139778a2a806941b33135049635fa6e703da8

**Scripts**
- `prepare_exports.py` 77446e97386561cf62f5afb8cfea96f5809f96445b8d724e223d2a57cb3eca44
- `assemble_preparation.py` 3e3288a4f0cb5f237f2e7160e7b3d81ebf4454958baab6bf81eb5a23e0899d3d
- `stages.py` e58dc239e13615c989410502afe47a8ef64f84a1e810bd6749551c654cb06a02
- `inventory.py` 88cdd3b1a76b60c7394f43df58fe52a0573f671f7d3c12c15f2f75ea2f1db9cc
- `pin_tool_closure.py` 39dff5b8a75e41852af7d6f5f534728ed0667626e1eb42e8ab38e657109e8348
- `extract_exposures.py` 966f7ed55d0a30c3d5cc4e2deea5c78b912f98827e261bbdb9b265c44749c6b8
- `bootstrap_v6.py` 5b8735ad1c03ca0afebf54dfe3f1123a96dabb72002db53de16f60d2ae0244e9
- `bootstrap_selftest_v6.py` a9901432ace306ce867aa1e09ef617a24a4e8b6347359e86415a831784978ff7
- `selftest.py` 1132cb745575ba3e14485d37b9a1ac69dd8a4cce09a2cbd531cc0506e97f2e19
- `prep_selftest.py` ed65b441b6e6d0e05dc18789a361931cc4b47c1f069d171a54acfe2febf9d793
- `capacity_selftest_v8.py` 25ad58bc1950185f9e57f7b18878f02d651a2dfc82b9527468e930a28d93c3b4
- `preparation_capacity_v8.py` 140eee8d70ea08cb88c07feddf57264a55ec907815e2230cf6b05afe31dcf083
- `prep_scale_v8.py` 721f91473786519326ae916e8f3922d0306499ffec71cc8e9e6f4fe44f911eba
- `frontend/main.go` 8bbe00b526dd39ef0905fb85e47d85326418e57c062ebb241d8fd55770c4c6f9
- `frontend/main_test.go` 1b49b60b85568ef8d3b440d12590335e5a9e17eb7b9997514af9827b833cf3c0
- `envelope_check.py` 00b6bf253f5ea2547854f228526fff34535084e008874725540769b9600bd1f5
- `preparation_capacity_v7.py` 1599b49bbb41b48454bc8394fdb1dc6398d760f1957a69250f0c1fe9b561453b
- `prep_scale_v7.py` 956624bb7e26abe215c8e75b3ea0e93069de443867dcb3256a5ef083e3ad6ed0
- `oc_v5.go` 34e51550f3f0a4d7d406e4302200af9a1f7302b33186b97c1734450373bb4966
- `oc_v5.py` 5886200086055b9e403a823f465e6342981466d229f660fed8f55a8abfefcbcf
- `oc_selftest_v6.py` 0eba0a6330c8ca7bc52819684e94dd6b57387960f4fe757a03d71e35438789b4
- `oc_timing_v7.py` ff283ad8d923ada9d282b8d777b3f69750f2f4c5bb749aae87171199272e0a9f

**Capacity and benchmark evidence**
- `preparation-capacity-v8.json` efff9344728e6ba37932a53e557aefba44ae51a727599da30631cbd4f1b7692f
- `prep-scale-construct-v8.json` 1f7ac4d27f0e11a2e49acdbcdb8fc281bca5c4bc43c9b4b06954ee427f6617d7
- `prep-scale-files-v8.json` fc659eb9a719e603eaa401c0434e10d2662dde2645b73f062671b14a82606e39
- `prep-scale-bytes-v8.json` 1b3bd639f7a03caf4d0627bcd77d09f6bcde0b6122e35d1cc2f79541eb1a08fd
- `prep-scale-heldout-ts-prediction-v8.json` 4b993052ef109fdcf0c53dce2f991392c0f6e163fa5091d8c2394be30406ccfe
- `prep-scale-heldout-ts-v8.json` c47075f934c1beaaa5fe7038aaea5eb4124413f3c9ec9470d0d6df8387ffe0c9
- `prep-scale-heldout-ts-v8-run.json` c8405cda963c8f155ab82c4d1fbfd33841a302462ef3ca462b3a321e82f63690
- `prep-scale-heldout-ts-inputs-v8.json` 06676fb31747c554b44841fb76aa334f759596c77f58f169f10d8ebabbceae95
- `preparation-capacity-v7.json` 141c6764dee809dd0bb198d53e1a414c2f744b1403d9d99c1e282af530c8e037
- `preparation-capacity-v7-run.json` f50871675e3295f8e2522ab4b73ec8b54e8e885ab2d35f0c0c5f3108d0a2c42f
- `prep-scale-files-v7.json` 4aa468fd5dd4c96ff815a24d48fd17d175b4ea8fea3bd2ba9b20a16d671de2b2
- `prep-scale-bytes-v7.json` fd2fa231a08d8d55353088a2c9712eb129810f4e6b594ef00d88749e59f11cd2
- `prep-scale-construct-v7-attempt2.json` 19fc7fd814a43329c0d92b72de43f5b449e7b6a3632f387404793592bfb4d8b5
- `prep-scale-construct-v7-attempt1.json` b3c55ccbb318c337ba54e08b10a0df988511b92b7c5423bb7cc19c98ee1a5cc7
- `envelope-v7.json` f8ebcdcb8aad65d1b288433a1180ae12a36fa7ff3cfb2b088c526a7dcb89d5ca
- `envelope-v7-run.json` 30aa18632d5e31069ad9dbcb1a85bd62ef321b4c2917abc90ea25d0e2fecd6b0
- `envelope-v7-inputs.json` e492da740af2aa778304394dd326c269099e7e7aa30a9207e680c92b0f34c74b
- `envelope-v7.stderr` e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855

**OC and tool pins**
- `oc-v5-spec.json` 7f0f2f83887634b0ca303e189e5f9c8ec4505e5b06a646bfbb10b2c4b08999b4
- `oc-v5-fixtures.json` b5c2a1e6cd35284b633489f26e70db3d5c65c7ce145abdb280853c950609137e
- `oc-v7-timing-metadata.json` c81e284bc6d3309d55a2f4a8eafaf2b2c9b84f44d052182a205f23248ed075d5
- `oc-v7-timing-go-closure.json` 2df0ec6ded049c4d1a30650804aac78aa089a1c3f07ce9bc9fa7d87b6000a16e
- the six `oc-v7-timing-{balanced,concentrated,singleton}[-maximum].json` files, as pinned in the manifest
- `tool-pins-v7/frontend-build-recipe-v7.json` 8b149edf4a67ef9797d8acc9361414fb1dfa8529e510ae45f72bc2eb29c33426
- `tool-pins-v7/frontend-go-closure-v7.json` 788f2a3be52a471602a4d622c90dae88b52d79a5bd8b8c1d9383c2ae42a22e37
- `tool-pins-v7/oc-build-recipe-v7.json` e11bad826e142715036fdf66dbc166b672be0b731f8a1be2e8cb035ab2197c81

**Exposure and cost evidence**
- `exposure-source-index.json` bccbfd4b8a2e42c9a91cdd9ead0c3f939c5606950a5fcb9844f8d4b6a141c3d8
- `exposure-contexts-v4.json` 2dd5b6b7d6665d64519585acdb63ed6eafadc3092bda25a38433616f896199fb
- `latency-basis-v4.json` bdc3cef935d65a20b4b0b9f34bb68c881d29da4960e74e5ee5b60549f0894104
- `cost-feasibility-v5.json` 1637256f02d54164c73eff82abb8648b4f70f4924701c9c29ee5a54e6c76c1d0
- `exposure-extraction-v4/seal.json` 96fcfcbfada87faaa6624d9124fb3c88d55d714a8cd0bf79ce72a114e32f8510, with its five outputs:
  - `additional-identifiers.json` 03f6628b66f5d52f9c2f1b3ed1540e4d99541f028c3778a90ce99721a073f3ca
  - `coverage.json` d2a29358c31abb66034fe7dfeda5e88e4741f37ceae14fbdb8ce3764d79b4777
  - `exposure-identifiers.json` 3a48e75f39a2a773cfadb2e9bc9ea12eb1f4b6ec6d33026dabde6a5f501f0ab8
  - `inertia-enumeration.json` 9d6f99c4efee874f7162e8d6576207e289347999998df38184d463f8f926497d
  - `unattributed-identifiers.json` b15ab49722bb9f40538cd1b16e02082e6b4be31c16348cbd1e2af4c76c1c633d

**Review records (preserved evidence, not authority)**
- `reviews/round-7.md` 795e45259cbbebd16e2fe6f9f2e36417dcf34da4bc3e62be608205f39baee173
- `reviews/round-7-response.md` 193f2ac0ef3ab499eb8ddd15ba6b066198c65bffa95e55484d002e610772c31e
- `reviews/rev8-checks.json` d7d86b07561f0ef9e871074690ad6e771462d055a61b2509b460d82ecb25d8b1
- `reviews/rev8-integration-checks.json` 1637af57d5c2031c00112cfa186f2b373f7af0d507ff089f290e7c74fbd94f3d
- `reviews/rev8-heldout-request.json` e196f8af705d82b090e57bb676776ebe1d5310e011eabb35166ef0bdf3512489
- `reviews/rev8-pre-integration-digests.json` 9eb55af605304d7ec0f321dd3db4c8c4af491a289dd3286cdd1e55a31306429f

**X11b plan**
- `../x11b-guard-recovery/PLAN.md` 034284255f303ad09308f30fc6212efac6d2730aa50ac724c4a8243df3d48753

**ADR-013 and shared files**
- `docs/adrs/2026-11-08-x11a-delegated-decisions/ADR.md` 2f273489c1e15f82af4a2433dcbfbb38ecde8a012bf5073fad12befd3898814c
- `docs/adrs/2026-11-08-x11a-delegated-decisions/round-1.md` 7746296a149ef67bddf3aa4671797117013379d50b832510790d5313ffb3368b
- `docs/adrs/2026-11-08-x11a-delegated-decisions/round-2.md` 795e45259cbbebd16e2fe6f9f2e36417dcf34da4bc3e62be608205f39baee173
- `docs/adrs/README.md` d15b4d89a0d9592697ec36469e9d3fe882e7d13777b007b2df03c78c01308297 (ADR-013 row :33)
- `docs/scout-design.html` 9a54215d65aa0fbca0dcc9604c6eef1bbf47e3cb483728d579cd1f51350290c8 (ADR-013 article :1014–1029)
- `docs/research-ledger.md` fc7b6e22ff860d74031f0de3cb7a6d30fd3b5e8ffeae942b4359f9c54cf55dfe
- `docs/hypothesis-ledger.md` 90644b6f035d4e1e0ed025862274fda787340910f0dd74707fe9cfc79d07351c
- `docs/STATUS.md` 4e902ea9ba632499d6597b66992874d370055c67f166e228d53e3808e9c40552

All other manifest entries are covered at their manifest digests: the Go toolchain, the Clawpatch tree, the Python runtime and persistent-location pins, and the historical receipts and records.

**Conditions of this authorisation and sign-off:**

1. **The bytes are frozen.** Under experiment rule 5 and design rule 2, any change to a signed file voids this authorisation and needs a new zero-issue round. This covers:
   - the preregistration, recipes and README;
   - the executables, inputs and evidence;
   - earlier records and manifests;
   - ADR-013 `ADR.md`, apart from the status-line updates design rule 2 allows when a later ADR is adopted.

   The only exceptions are those in condition 2.
2. **Permitted later edits in shared files** (`docs/STATUS.md`, `docs/research-ledger.md`, `docs/hypothesis-ledger.md`, `docs/adrs/README.md`, `docs/scout-design.html`):
   - Sibling edits are allowed except in the X11-owned spans:
     - the ADR-013 index row;
     - the ADR-013 card;
     - research rows R-073–R-076, R-079 and R-082–R-084;
     - the HD8 register row and existing HD8 history lines.
   - STATUS stays an editable tracker, including its X11 lines.
   - A change to an X11-owned span voids this authorisation.
   - Main checks any landing by diffing these five files against their signed bytes.
   - `reviews/README.md` may gain this round's row.
3. **The authority artefact is a single Main act**, written once:
   - `verdict: "ZERO ISSUES"`, `scope: "export-preparation"`, `review` = this record's pin.
   - `files` = the manifest entries the verifier requires (STG:12–15 `REVIEWED`, plus `inventory-inputs.json` at its resolved canonical path and its tool, recipe, source and exposure-extraction closure). Any other immutable signed entries may be added.
   - **Exclude** the five shared files and `reviews/README.md`, because sibling edits during the multi-day preparation would make preparation refuse (STG:58–59).
4. **Authority is limited to:**
   - (a) one `start_budget` at the canonical path;
   - (b) `build-frontend`;
   - (c) `exposures`;
   - (d) `ownership` for each of the five configured repositories at their canonical paths, with ≤ 5 concurrent;
   - (e) one `assemble_preparation.py` run.

   No inventory, OC run, map, label, replay or model call is authorised. Inventory needs its own artifact-binding round, which must check that the manifest's `preparation_authorization` equals the condition-3 artefact.
5. **Outcomes.**
   - Every refusal, interrupt or budget stop is sealed and reported, with no external rerun beyond the three internal attempts.
   - Continuing after a host failure needs a council-reviewed deviation.
   - "Preparation not feasible" or not-testable outcomes are reported, never substituted.
6. **Adoption happens outside the signed set** (route (a)). No ledger edit is required. HD8 stays `proposed` until final study sign-off.
7. **Attestation.** The digests, the base, the integration facts and the independent runs rest on Main's reports. The council had no shell.

## Council synthesis

### 1. Recommendation

Authorise export preparation on the rev-8 bytes, and sign off ADR-013, both under the conditions above.
- Main runs the preparation operations once, at canonical paths.
- Main then assembles the sealed manifest and requests the inventory artifact-binding round.
- R1-01 stays deferred to the implementation round.

### 2. Alternatives

- **A. Recommended: authorise now.** All round-7 fixes are present, the held-out bound passes with 3.45× headroom under the frozen rule, and the deadlines fail closed.
- **B. Require a source-heavy Node stress point first** (challenger's viable alternative). Not recommended:
  - The 200-candidate cap already saturates per-group association cost in the fixture.
  - Ownership time follows tree files, not owned count.
  - Adding a point after seeing a passing result would be exactly the post-hoc rerun that R7-01 forbids.
- **C. Have Main also rerun the Go points.** Not recommended. It cannot change TypeScript's S_bound, which floors at S_obs, and the other repositories have ≥ 27× headroom.
- **D. Withhold authorisation until R1-01 is complete.** Rejected. Preparation is model-free and outcome-free, R1-01 stages refuse in code, and the preparation-first sequence has been the agreed plan since round 1.

### 3. Evidence

- **Capacity:**
  - receipts `prep-scale-{construct,files,bytes}-v8.json`;
  - prediction, held-out receipt and run record;
  - `preparation-capacity-v8.json`;
  - CAP8:24–105; arithmetic in §3.
- **Scanner path:** CP `node.ts`:196–257, 424–431, 454–464, 559–567; `mapper.ts`:108–117, 314, 335; receipt :128, :560–574; v8 file-heavy :136.
- **Enforcement:** PREP:31–68, 72–92, 977–1006; ASM:11–41; STG:44–61; PST:306–318, 337–341, 508–570.
- **Canonical texts:** PREREG l.3, 35–37, 50–54, 167; REC:124–126; ADR:9, 73–74; `scout-design.html`:1015–1029; `research-ledger.md`:88; index :33; STATUS :7, :131, :135; `hypothesis-ledger.md`:35, :40.

### 4. Risks

- **Synthetic content cannot trigger real-content Node paths** (routes, workspaces, React) [INFERENCE]. The runtime deadlines (3,600 s, 14 days, 42 days) are the fail-closed backstop. The result would be an operational not-testable outcome, never a content exclusion.
- **The three-point fit has no error estimate, and each point ran once.** The 1.5× factor is a margin, not a confidence bound (disclosed).
- **Disk and host failure:** any of these ends as an operational stop that needs a reviewed deviation.
- **Reversibility:** preparation produces only sealed local artefacts. It runs no model and computes no predictor–outcome association. Once it has run, PREREG is frozen and any byte change needs a renewed round with exposure disclosed (PREREG l.37).

### 5. Execution outline

1. Main saves this record as `reviews/round-8.md`, and byte-identically as `docs/adrs/2026-11-08-x11a-delegated-decisions/round-3.md`.
2. **Commit the signed set** before the first preparation action.
   - If origin/main has moved, rebase.
   - Check: the five shared files differ from their signed bytes only outside the X11-owned spans, and every other signed byte is unchanged.
3. Write the condition-3 authority once.
   - Check: `verify_authorization(…, scope='export-preparation')` accepts it.
4. Run `start_budget` once, then `build-frontend`, `exposures` and the five `ownership` operations at canonical paths, with ≤ 5 concurrent.
   - Check: seven seals exist under `artifacts/preparation-v8/` and none is empty.
5. Run `assemble_preparation.py` once.
   - Check: a `SEALED` manifest, or a recorded not-testable stop.
6. Report the seals, elapsed times against the projections, and any deviations. Then request the inventory artifact-binding round.

### Where the peers changed this memo

- **innovation-challenger:**
  - Independently confirmed all three closures and that the comparison used the correct S_obs.
  - Showed that the 200-candidate cap makes round 7's 50/50 rationale wrong, while leaving the benchmark valid. This was adopted into §2 as a correction of the council's own round-7 reasoning.
  - Proposed the source-heavy stress alternative (B) and recommended against making it a prerequisite. Adopted.
- **experiment-peer:**
  - Supplied the ownership-only comparison (1.410 against 1.431 ms per tree file). It replaced the lead's planned linear-in-owned-count sensitivity bound: ownership count turned out not to drive cost.
  - Also supplied:
    - the sorted-walk reading showing the fixture saturates the cap;
    - the falsifiable H1 test of Node-path execution;
    - the accounting-variant table and the timing order;
    - the `recipe_sources` comparison for the inputs-digest change;
    - the unplanted-override, weak-assertion, fan-out and byte-term break-even points.

    All of these were checked and recorded as not raised (§2).
- **Lead's own additions:**
  - Traced the relocation route to the resolved-inputs binding (STG:60–61) and its residual (a forged second authority).
  - Found that `REVIEWED` excludes the shared files, which led to the authority-composition rule in condition 3.
  - Adapted the X1 round-9 sibling-integration and adoption conditions to X11's owned spans.
  - Confirmed that HD8 needs no ledger edit, because its mapping moves it only at final sign-off.
  - Read the PREREG draft header and the STATUS ADR range as accurate route-(a) and tracker states.
