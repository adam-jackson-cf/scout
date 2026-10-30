# X2d — round 0 record and council critique

## Main-agent record (before any label; 2026-10-30)

**Sign-off conditions (council round 11, `reviews/round-11.md` §5 and Addendum).**
1. All 44 frozen-file digests and the `tune_threshold.py` pin were recomputed and equal the round-11 table.
2. `signoff.json` was written (round 11): SHA-256 `e7059f2a9407e38c4d5e8aff0480913fab36a32a69ccbb9721c9633934fdbc2d`. Its `code_under_test` holds 17 rows, equal row for row to the round-9 record. `x2d.verify_pins()` returns `[]`.
3. Validation ran on these bytes:
   - (a) with `signoff.json` moved aside: **651 checks, 0 failed** (`/tmp/x2d-validate-r11-norecord.log`). The record's digest was `e7059f2a…dc2d` both before and after the move.
   - (b) with the record in place: **650 passed, 1 failed**. The failure is `C without the council's sign-off record the collector refuses before any clone or model call`, which is expected per the Addendum (`/tmp/x2d-validate-r11-signoff.log`).
4. Commit `0b5659440c7368ac18b82f60a4edf3956408d2bf`, pushed.
5. The run worktree `../scout-x2d` was moved to `0b56594` by checkout only. `data/` stayed untracked and untouched; the catalogue digest is unchanged.

**Two commits (PREREGISTRATION §14 item 18; for RESULTS `Deviations`).**
- These stages ran at `730674e` under the rev-9 sign-off: onboarding (16 maps, all rc 0), consensus and catalogue freeze (114 areas, sha256 `1220016c26115ccd3e3a521de76ec45e68893f4ab6edf0f4889e8c8ce2c84199`), and quality with the intruder audit (40/40 both labellers, κ 1.0).
- The first `gates` attempt refused with the P2 shortfall (`reviews/pilot-round-0-p2-gate-memo.md`). Nothing after the refusal ran.
- Everything from `gates-build` onward runs at `0b56594`.

**R10-03 comparison: `gates-build` alone (`/tmp/x2d-gates-build.log`) against `reviews/round-10-response.md` §2 and `reviews/round-11.md` §2.** Keys are `(kind, source_case_id)`. The file is in blinded order, so sets are compared, not file order.

| item | recorded | live build | equal |
|---|---|---|---|
| kind counts | p1 20, p2 12, dup 9, planted_new 3, prose 1, trap 1 | same | yes |
| P1 source cluster ids (set) | the 20 ids of round-10-response §2 | same 20 | yes |
| P2 source cluster ids (set) | the 12 ids of round-10-response §2 (4 P unplaceable, 2 P leftover, 6 D) | same 12; stdout `p2_donors` {P_unplaceable 4, P_leftover 2, D 6} | yes |
| `ablated_area_id` set / `ablated_area_ids` | 10 areas `26ac01437eaf`, `73232e7095cf`, `95b28e00298a`, `9a4a0e49b8c6`, `c207aaf8cb60`, `c4f3d5556193`, `d56ca48225f1`, `ea1d02de2a93`, `ea75266c1719`, `f36cd1216200` | same 10, identical list on every P2 row | yes |
| `gap_positive` | 12 of 12 | 12 of 12 | yes |
| shared ablated areas | `26ac01437eaf` (`2c4db772`, `af2f04c8`: P/P path twin); `73232e7095cf` (`d636de1b`, `84f3f93b`) | same | yes |
| `catalogue-p1` / `-p2` / `-n1` areas | 114 / 104 / 114 | 114 / 104 / 114 | yes |
| every P1 gold area in `catalogue-p1` | true | true | yes |
| `units.jsonl` / `units.contrast.jsonl` lines | 96 / 20 | 96 / 20 | yes |
| `data/catalogue/catalogue.sha256` | `1220016c…4199` | file and recomputed digest both `1220016c…4199` | yes |
| stdout | eligible 26, eligible_d 19, p1_unplaceable 4, ablated_areas 10 | same | yes |

**P1 `target_dir` list:** `.github`, `.github/actions/notify-translations`, `.github/actions/people`, `.github/workflows`, `docs_src/app_testing/app_b_an_py310`, `docs_src/bigger_applications/app_an_py39/internal`, `docs_src/settings/app01`, `docs_src/templates`, `fastapi/security`, `tests` (×2), `tests/test_modules_same_name_body` (×2), `tests/test_tutorial` (×2), `tests/test_tutorial/test_advanced_middleware`, `tests/test_tutorial/test_configure_swagger_ui`, `tests/test_tutorial/test_header_params`, `tests/test_tutorial/test_query_params`, `tests/test_tutorial/test_request_forms_and_files`.

**K2/K3 from the live build (stdout line 2, membership only):**
- K2 = 0.585 (B0 unique owner on 193 of 330 scored clusters), below the 0.90 ceiling, so there is room for Jev; K2 for B2 = 0.806.
- K3 = 0.052 (17 stratum-3 clusters).
- `real_none_proxy` = 39; core gate not enabled.

No mismatch, so the resume continues with `gates-label`.

## `gates-label` and the `gates-jev` cost stop (main agent, 2026-10-30, at `0b56594`)

- `gates-label`: 116 cases labelled (96 main + 20 contrast units), 317 of 429 CLI sessions used; labels-check `ok` (κ_answer 1.0 on 50 units; report-only at round 0); `gold.jsonl` and `gold.contrast.jsonl` written.
- `gates-jev` stopped with exit 4 at the preregistered $1 per-round cap: 7,755 of the 7,872 requests of the `main` run answered, $1.0001 billed; 7 rows `kind: network`, `cost_known: false` (requests cancelled at the stop; their cost is unknown, not zero); the `n1`, `p1` and `p2` V-sem runs had not started; `compose` had not run. Measured cost per request ≈ $0.000092 (fan-out `merge.area_covers`) and ≈ $0.000589 (other), against the preregistered $0.000055. The user raised the caps to $2 per round / $20 total (R-052); PREREGISTRATION rev 12, council rounds 12–16, signed off in `reviews/round-16.md`.
- **Attestation (R12-02 (d)).** The main agent attests that no Jev answer was compared with a label or gold before the user's decision, nor since: no stage that reads both (`gates-analyse`, `tune`, `analysis`) has run, round 0 has no `results.json`, and its `scores*.csv` tune inputs hold only their 67-byte headers.

## Third run commit (item 19, R15-01)

The rev-12 sign-off commit is `b6a5bf30324201bc88caae75d5774255c6cac54d` (pushed to `origin main`). The run worktree `../scout-x2d` moved to it by `git checkout --detach` only (no `git clean`, no `stash -u`; `data/` untracked and untouched); worktree `git rev-parse HEAD` = `b6a5bf30324201bc88caae75d5774255c6cac54d`. Sign-off conditions: 44 digests and tune pin equal to `reviews/round-16.md`; `signoff.json` (`round` `reviews/round-16.md`) with `code_under_test` equal to round 11's 17 rows, `verify_pins()` = `[]`; validation (a) 651 checks / 0 failed with the record aside (`/tmp/x2d-validate-r15a.log`; same frozen bytes); validation (b) 651 checks / 1 failed with the record in place, the one expected failure `C without the council's sign-off record the collector refuses before any clone or model call` (`/tmp/x2d-validate-r16b.log`). From here to `verdict` the worktree is not checked out again; each critique is copied into the worktree's `reviews/` before the stage that reads it, and the copy's SHA-256 is recorded in the next record.

## Round-0 re-freeze (R12-01, `PREREGISTRATION.md` §14 item 19)

1. Before: `rounds/0/version.json` `6355be3fbcd058b87ba4057e1939eb3ee2435cd28c7de08305fb5c3ebc1e3ca8`, `rounds/0/settings.json` `c165a695a0a7bfd6ae91fcee630b80db2c58949186dd1e0ab1ea3a77decef6e2`. `version.json` renamed to `version.rev11.json` (kept; SHA-256 after the rename `6355be3fbcd058b87ba4057e1939eb3ee2435cd28c7de08305fb5c3ebc1e3ca8`, unchanged).
2. Command, from the run-worktree root at `b6a5bf3`, `X2D_DATA` unset: `.tools/venv/bin/python evaluation/x2d-merge-association-pilot/versioning.py --round 0` → `froze version of round 0: …/data/rounds/0/version.json`, exit 0.
3. Key diff (`version.rev11.json` → `version.json`): exactly `thresholds_sha256` (`5c211eb939b4880ae21e79fdfcfc25060722f73a96b99d6f27aaceb2153094a8` → `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb`, the signed rev-12 digest) and `question_version_keys` (`null` → the `versions` of the `rounds/0/jev/scores.jsonl` header, equal: 7 entries). `settings.json` byte-unchanged (`c165a695a0a7bfd6ae91fcee630b80db2c58949186dd1e0ab1ea3a77decef6e2`).
4. `versioning.py --round 0 --drift` → `round 0: no drift from version.json`, exit 0.
5. After: `version.json` `e079e2d1df7d851552119213d7e4117ac0a7c679a6946c510ea144e39657b64f`, `version.rev11.json` `6355be3fbcd058b87ba4057e1939eb3ee2435cd28c7de08305fb5c3ebc1e3ca8`, `settings.json` `c165a695a0a7bfd6ae91fcee630b80db2c58949186dd1e0ab1ea3a77decef6e2`. RESULTS `Deviations` will state the re-freeze and why (the rev-12 cap change altered the `thresholds.json` digest pinned by round 0's version record).

The resume continues with `gates-jev`.

## `gates-jev` resumed, `gates-analyse`, `power` (main agent, 2026-10-30, at `b6a5bf3`)

`gates-jev` (`infisical run … -- run.sh gates-jev`, log `/tmp/x2d-gates-jev-resume.log`), exit 0:

```
round 0: no drift from version.json
jev: 64 units, 7872 requests (117 pending at start), 117 answered this run, 0 failed, 0 with unknown billed cost, cumulative billed cost 1.009834 USD of which 1.000115 earlier (model typesafe/jev-1.14-20261014)
jev: 50 units, 110 requests (110 pending at start), 110 answered this run, 0 failed, 0 with unknown billed cost, cumulative billed cost 0.048770 USD of which 0.000000 earlier (model typesafe/jev-1.14-20261014)
jev: 20 units, 43 requests (43 pending at start), 43 answered this run, 0 failed, 0 with unknown billed cost, cumulative billed cost 0.025070 USD of which 0.000000 earlier (model typesafe/jev-1.14-20261014)
jev: 12 units, 26 requests (26 pending at start), 26 answered this run, 0 failed, 0 with unknown billed cost, cumulative billed cost 0.014860 USD of which 0.000000 earlier (model typesafe/jev-1.14-20261014)
composed V-sem: 64 clusters -> /Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data/rounds/0/arms/V-sem.jsonl
composed V-sib: 64 clusters -> /Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data/rounds/0/arms/V-sib.jsonl
composed V-title: 64 clusters -> /Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data/rounds/0/arms/V-title.jsonl
composed fan-out: 64 clusters -> /Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data/rounds/0/arms/fan-out.jsonl
composed V-sem: 50 clusters -> /Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data/rounds/0/arms-n1/V-sem.jsonl
composed V-sem: 20 clusters -> /Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data/rounds/0/arms-p1/V-sem.jsonl
composed V-sem: 12 clusters -> /Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data/rounds/0/arms-p2/V-sem.jsonl
```

Round-0 Jev spend: `main` $1.009834 (of which $1.000115 before the stop), `n1` $0.048770, `p1` $0.025070, `p2` $0.014860; total ≈ $1.0985 of the $2 round cap. No failure and no unknown cost in the resumed requests (the 7 `network` / `cost_known: false` rows of the stop remain, as recorded above).

`gates-analyse` (`run.sh gates-analyse`, log `/tmp/x2d-gates-analyse.log`), **exit 1**:

```
PASS K2 value=0.5848484848484848 bound=0.9 n=None
PASS K3 value=0.051515151515151514 bound=0.4 n=None
PASS K4 value={"kappa_answer": 1.0, "kappa_covered": 1.0, "raw_agreement_covered": 1.0, "prevalence_covered": 1.0, "raw_agreement_answer": 1.0} bound=0.6 n=20
PASS K5 value=0.04 bound=0.15 n=50
PASS K6_N1 value={"jev_accuracy": 0.0, "wilson_ub90": 0.03180284473263897, "K": 114, "B2_unaffected": true, "units_without_arm_row": 0} bound=0.2 n=50
FAIL P1 value={"V-sem": 0.45, "B1": 0.05, "B2": 0.05} bound={"jev_min": 0.6, "baseline_max": 0.2} n=20
PASS planted_new value={"none_rate": 1.0, "reported_as": "none_correct_new", "reported_as_n": 3} bound=0.8 n=3
FAIL P2 value={"none_rate": 0.5714285714285714, "none_correct_gap": 4, "none_correct_gap_rate": 0.5714285714285714, "gap_positives": 7, "retained_covered": {"n": 5, "assigned_correct": 1}, "p2_clones": 12, "no_agreed_gold": 0} bound=0.8 n=7
INFO real_none_proxy value=39 bound=null n=98
fan_out_primary=False
GATES FAIL: P1, P2
```

`power` (`run.sh power`, log `/tmp/x2d-power.log`), exit 0 (no null cell above its bar): `n4_max 0.054; checks {"P3_u0": true, "P3_u05": true, "P3_u10": true, "N3": true, "N3_PR": true, "N4a": true, "N4b": true, "N4b_cap": true, "N4b_rho0": true}; seconds 34.9`.

**Gate stop.** P1 (V-sem 0.45 < 0.60; B1 and B2 0.05 ≤ 0.20) and P2 (7 label-derived gap positives, below 8; 4 answered `none`, so 3 ≥ 2 misses → FAIL, not UNMEASURED) fail. Under `PREREGISTRATION.md` §5 (`:110`) "Failing a gate stops the pilot before round 1 (the failure is the finding: it is reported, not repaired silently) … a guide or question fix needs a new review round." Round 1 has not started; HD6 stays `preregistered`. The next step is a council memo on the stop and the user's decision.

## Gate-failure disclosure (from the council memo §2; preregistered P1/P2 descriptive disclosures)

#### P1: per item (the preregistered descriptive disclosure)

How to read the columns:
- **A** is the gold area, the file's owner before the move (`control-key.jsonl` line N = ctl-N).
- **B** is the area the directory baselines assign (`data/rounds/0/arms/B2.jsonl`; B1 is the same on every item). Stratum 3 means no owned sibling.
- **Answer** is V-sem's composed output (`arms-p1/V-sem.jsonl:1-20`), and **raw** is its unthresholded permutation-1 choice.
- **Labelled set** is the agreed acceptable set in `gold.jsonl`.

| Item | Stratum | File moved | A | Answer (conf.) | Raw | B | Labelled set | Category |
|---|---|---|---|---|---|---|---|---|
| ctl-1 | 3 | separate-schemas tests | ad8836 | **A** 1.00 | A | (none) | A, e287e9 | hit |
| ctl-2 | 1 | `tests/test_forms_single_model.py` | 3de70a Parameter declarations | 158cb3 Request form model tutorials 0.91 | 158cb3 | 3e0c34 Advanced middleware | **158cb3**, A, 7d2013 | gold narrower than labels |
| ctl-3 | 3 | dependency 008c tests | c3534a Dependency injection (tutorial) | unresolved 0.44 | A | (none) | 9a4a0e, A | threshold; probability split with the core area also titled "Dependency injection" |
| ctl-6 | 2 | playwright `image02.py` → `.github/workflows` | ef2b9b | `none` (p_true .67) | A | tie | A | covered threshold |
| ctl-8 | 1 | `tests/test_ws_dependencies.py` | 36295a | **A** .91 | A | ade7d5 | A, 9a4a0e | hit |
| ctl-12 | 2 | `fastapi/cli.py` → `fastapi/security` | 1c4aa7 | **A** 1.00 | A | 637a9b | A | hit |
| ctl-13 | 1 | form-model test | 158cb3 | **A** .86 | A | 749bd1 | A, 3de70a | hit |
| ctl-20 | 1 | `docs_src/request_form_models/…` → `docs_src/settings/app01` | 158cb3 | unresolved (the two option orders disagree) | 7d2013 Request forms | 9ce9c5 | A, 3de70a, **7d2013** | near-duplicate areas |
| ctl-21 | 2 | dependency 008d tests | c3534a | **A** .83 | A | e287e9 | 9a4a0e, A | hit |
| ctl-23 | 3 | `tests/test_additional_properties_bool.py` | e287e9 | **A** .91 | A | e287e9 (baselines fall back to `tests/`, which A owns) | A | hit; baseline leak |
| ctl-24 | 1 | recursive response tests → `test_header_params` | 617fde | unresolved (orders disagree) | **B** | ea1d02 Header parameters | A | path pull |
| ctl-25 | 1 | playwright `image01.py` → `notify-translations` | ef2b9b | `none` (p_true .66) | A | d1459f | A | covered threshold |
| ctl-26 | 1 | `docs_src/dependencies/tutorial008d*` → `docs_src/templates` | c3534a | unresolved 0.52 | A | 29a30e | 9a4a0e, A | threshold; duplicate "Dependency injection" split |
| ctl-27 | 2 | `scripts/mkdocs_hooks.py` → `.github` | 30399e | **A** 1.00 | A | tie | A | hit |
| ctl-31 | 2 | schema-example tests → `tests` | 95b28e | **A** .98 | A | e287e9 | A, e287e9 | hit |
| ctl-33 | 1 | playwright cookie-models `image01.py` → `.github/actions/people` | f0abaf | unresolved 0.49 | **B** | 342c6b People GitHub Action | A | path pull (real) |
| ctl-34 | 1 | `docs_src/sql_databases/tutorial002*` | 8a632c | unresolved 0.69 | A | 3d2f5f | A | threshold (0.01 short) |
| ctl-36 | 3 | header-models test | b58662 | **A** .96 | A | (none) | 3de70a, A | hit |
| ctl-39 | 1 | `docs_src/header_param_models/tutorial002*` → `bigger_applications/…/internal` | b58662 | unresolved 0.68 | **B** | 2874a6 Bigger applications | 3de70a, A | path pull (real) |
| ctl-46 | 1 | `tests/test_response_model_data_filter_no_inheritance.py` | 617fde Response model validation | 10c16a Response models .86 | 10c16a | 233f9b | **10c16a**, A | gold narrower than labels |

**Counts.**
- 9 hits.
- 0 composed answers equal B. Raw choice equals B on 3 of 20 (ctl-24, 33, 39; Wilson [0.05, 0.36]).
- Raw choice equals A on 14 of 20, and lies in the labelled set on 17 of 20.
- By stratum: stratum 1 recovered 2 of 11 ([0.05, 0.48]); strata 2 and 3 recovered 7 of 9 (`units.jsonl` `stratum`).
- The 20 items come from only 13 PRs, so the effective n is below 20.

**What is construction or measurement:**
- **Gold narrower than the labels.** Gold is the single owner A (`build_controls.py:316`; scored by `analysis_x2d.py:1355-1357`). The blind labellers' agreed sets are wider on 12 of 20 items and contain A on all 20. Jev's two confident "errors" (ctl-2, ctl-46) are in those sets.
- **Untuned threshold, abstention counted as a miss.** Five raw-A answers fell under the provisional 0.70 threshold (`data/rounds/0/thresholds.json` `provisional-0.7`; `compose.py:68-90`), and abstention counts as a miss. P1 dates from rev 1 and was never realigned with R-051, which judges accuracy on answered clusters and caps abstention separately (`PREREGISTRATION.md:134-136`).
- **Near-duplicate areas across scope shards.** There are two areas titled "Dependency injection" (c3534a and 9a4a0e), and "Request forms" sits next to "Request form model tutorials". This is the known scope-shard threat (`:211`); it also affects the main rounds.
- **Baseline leak on ctl-23.** This did not decide the gate.

**What looks like a real Jev weakness:**
- ctl-33 and ctl-39: the content plainly names A's topic, yet the raw choice is B.
- ctl-24: the same pull, but more diffuse.
- Jev's low confidence in stratum 1.

**Sensitivity (descriptive only, not a re-scoring):**
- Against the labelled sets, V-sem scores 11 of 20 = 0.55 ([0.34, 0.74]), which still fails. Baselines score 3 of 20 = 0.15, still within the 0.20 bar.
- Lowering both thresholds to 0.5 would give exactly 12 of 20 on P1 (the experiment-peer's calculation), but it would also send ctl-39 to B. It would also cancel P2's `none` answers on ctl-11, 22 and 44, whose p_true values are .68, .54 and .59 (`data/rounds/0/jev-p/scores.p2.jsonl:9,15,23`). The threshold pulls the two gates in opposite directions, so it must not be moved on the strength of these data.

#### P2: the R10-02 split by source set

Sources are joined `control-key.jsonl` → `units.jsonl` `set` → `gold.jsonl` → `arms-p2/V-sem.jsonl`. Donor categories come from `reviews/round-10-response.md:39-52`. In this table, "derived gold" is what the gate used, and "direct label" is the clone's own blind label on `catalogue-p2` (`data/rounds/0/labels.jsonl`), which the gate discarded.

| Clone | Set | File | Ablated area | Derived gold | Direct label | V-sem | Gate |
|---|---|---|---|---|---|---|---|
| ctl-11 | P unplaceable | `pdm_build.py` | c4f3d5 Package build | gap | gap, agreed (:207) | `none` (p_true .68) | hit |
| ctl-18 | P unplaceable | `scripts/playwright/request_form_models/image01.py` | 26ac01 Playwright example: request forms | gap | **covered** f0abaf, 158cb3, adjudicated; labeller B said gap (:80-83) | unresolved .58, raw 158cb3 | miss |
| ctl-40 | P unplaceable | the same file (path twin) | 26ac01 | gap | **covered** f0abaf, 158cb3, adjudicated (:22-25) | assigned 158cb3 (.70) | miss |
| ctl-45 | P unplaceable | `CITATION.cff` | 73232e | gap | none (new), adjudicated (:138) | `none` (p_true .10) | hit |
| ctl-22 | D | `requirements-tests.txt` | c207aa Test suite and CI | gap | gap, agreed (:210) | `none` (p_true .54) | hit |
| ctl-35 | D | `tests/test_generic_parameterless_depends.py` | 9a4a0e Dependency injection (core) | gap | **covered** c3534a "Dependency injection" (tutorial), agreed (:286-288) | unresolved .68, raw c3534a | miss |
| ctl-44 | D | `tests/test_compat.py` | ea7526 Public API and compatibility | gap | none (new), agreed (:122) | `none` (p_true .59) | hit |

| Set | Clones | Gap positives | `none` | Misses | Retained covered (assigned correct) | No agreed gold |
|---|---|---|---|---|---|---|
| P (4 unplaceable, 2 leftover) | 6 | 4 | 2 | 2 (one twin pair) | 2: ctl-5, ctl-7 (0) | 0 |
| D | 6 | 3 | 2 | 1 | 3: ctl-10, ctl-16, ctl-30 (1, ctl-10) | 0 |
| Total | 12 | 7 | 4 | 3 | 5 (1) | 0 |

Ten areas were ablated, leaving `catalogue-p2` with 104 of 114. Two ablated areas are shared: 26ac01 (ctl-18 and ctl-40, a P/P path twin) and 73232e (ctl-7 and ctl-45).

**What is construction or measurement:**
- **Derived gold contradicts direct labels.** `derive_ablation` assumes the source's acceptable set (made on the full catalogue) lists every area that would fit once the best one is removed (`gold.py:92-104`). The study's own blind labels contradict this on all 3 misses, and in each case they accept the area Jev picked. This is a label-versus-label fact; Jev is not involved.
- **Twin double count.** The path twin counts one item twice. Round 11 kept the per-clone count as conservative (`reviews/round-11.md:54`).
- **A thin denominator by design.** Five of 12 clones keep a second acceptable area (`gold.jsonl:15,25,36,86,89`). The build checks ownership, not labels.

**Inherited D label errors: not observed.** The one D miss (ctl-35) had a correct source label on the full catalogue (`gold.jsonl:27`). Its problem is the derivation, not inheritance.

**Real signals:**
- On the retained-covered clones, Jev made one confident wrong assignment: ctl-5 → b58662 where the gold is 3de70a (`arms-p2/V-sem.jsonl:1`; `labels.jsonl:180`).
- Jev's `covered` reading is high (p_true 0.54–0.68) on 3 of the 4 true gaps. Its `none` answers depend on the untuned threshold.

**Sensitivity (descriptive only):**
- Collapsing the twins gives 6 positives and 2 misses: still FAIL.
- Using direct-label gold gives 4 positives and 0 misses: UNMEASURED, not PASS.

#### Classification

| Gate | Construction or measurement | Real Jev weakness | Confidence |
|---|---|---|---|
| P2 | Derived gold contradicts direct labels on 3 of 3 misses; twin double count; thin denominator built in | Overconfident `covered` reading on gaps; one confident wrong answer on a retained item | moderate to high |
| P1 | Narrow single-A gold (2–3 items); untuned threshold with abstention scored as a miss (5); near-duplicate areas; baseline leak | Path pull in stratum 1 (ctl-24, 33, 39); low confidence on displaced files | moderate |

Two more confounds:
- The labellers see more than Jev: 300 lines and 20 files per area, against Jev's 60 lines and 5 sample files (`PREREGISTRATION.md:129`; `questions/merge.covered@v1.json`).
- Wording is unlikely to be the cause. The question already says "never from their directory" (`questions/merge.which_area@v1.json:7`).

**Council-seen disclosure (rule 3).** In writing `reviews/pilot-round-0-gate-stop-memo.md` the council saw the per-item P1 and P2 Jev answers, their choice and covered probabilities, the P1 labelled sets and the P2 clones' direct labels. No other Jev-versus-label comparison was made. Any rev-13 gate repair is therefore outcome-informed, and RESULTS discloses it.

**User decision (R-053), 2026-10-30:** option (c): repair both gates and re-test once on fresh held-out items, with the Jev side frozen. The rev-12 FAILs stand as recorded and are never re-scored.

## Fourth run commit and re-gate build (rev 13, R-053; main agent, 2026-10-30)

- **Rev-13 sign-off commit (fourth run commit):** `8c7148cbb39b58440b39d18e0dddd0668e3e249c` (council round 18, `reviews/round-18.md`, zero issues; pushed). Conditions: 45 digests and tune pin equal; `signoff.json` (`round` `reviews/round-18.md`, SHA-256 `1ef1402d301b241d0d05d711f83773a5398688abbb844e08801d6ca8e2d1171b`) with `code_under_test` equal to round 11's 17 rows, `verify_pins()` = `[]`; validation (a) 715 checks / 0 failed with the record aside, SHA-256 equal before and after (`/tmp/x2d-validate-r18a.log`); validation (b) 715 / 1 failed, the expected `C without the council's sign-off record the collector refuses before any clone or model call` (`/tmp/x2d-validate-r18b.log`). Worktree moved by `git checkout --detach` only; worktree HEAD `8c7148cbb39b58440b39d18e0dddd0668e3e249c`.
- **`run.sh regate-build`** (alone; log `/tmp/x2d-regate-build.log`), exit 0: fresh pool 129 clusters / 129 PRs / 69 areas (D 3, G 7, `in_no_unit` 119); P1 40 of 40, strata 37 / 0 / 3; P2 40 of 40 clones in 16 ablated areas (D 2, G 2, `in_no_unit` 36); problems: none; arms written for 120 units.
- **Plan comparison (R17-07):** `regate/supply.json` P1 items (cluster ids, members, target directories, strata) and P2 items (cluster ids, members, ablated areas) equal `reviews/regate-supply.json` as committed at `8c7148c`; every top-level key equal; the two files are byte-identical. SHA-256: `reviews/regate-supply.json` `e499c52b3a3837b3b1202902b5d8242dd5488bfbefadf8bb46ea98e4e04eaa51`; `regate/supply.json` `e499c52b3a3837b3b1202902b5d8242dd5488bfbefadf8bb46ea98e4e04eaa51`; `regate/regate.json` `2b42b918cf34a6cfbf95bbb09a0a3fd6c5150698289083958397d71ec5d183ea` (`utc` 2026-10-30T22:54:55Z); `regate/rounds/0/control-key.jsonl` `5946b690b530612cc079a57aa66ab24d8ec76e26facc52d9094b64956769b81e`. No mismatch: the run continues with `regate-label`.

## Re-gate label, pre-Jev check, Jev and analysis (rev 13; main agent, 2026-10-30, at `8c7148c`)

- **`run.sh regate-label`** (log `/tmp/x2d-regate-label.log`), exit 0: 120 cases labelled (wide); 246 of the cap of 330 sessions; 120 gold rows.
- **Pre-Jev check, run alone first** (read-only; `X2D_DATA=<data>/rounds/0/regate .tools/venv/bin/python evaluation/x2d-merge-association-pilot/analysis_x2d.py --regate-precheck`), exit 0, no `PROBLEM:` line:
  `regate precheck: {"baseline_all_eligible": {"B1": {"k": 0, "n": 38, "value": 0.0, "wilson95": [0.0, 0.09180986822558396]}, "B2": {"k": 0, "n": 38, "value": 0.0, "wilson95": [0.0, 0.09180986822558396]}}, "direct_label_positives": 12, "items_built": 40, "items_eligible": 38, "rules_resolved_units": 0, "stratum1_eligible": 35, "twins_excluded": 0}`
- **`run.sh regate-jev`** under Infisical (log `/tmp/x2d-regate-jev.log`), exit 0: it repeated the same precheck line before any request; p1 run 87 requests, 0 failed, $0.049920; p2 run 87 requests, 0 failed, $0.043979; re-gate Jev spend $0.093899 (charged to the round-0 cap: round 0 total ≈ $1.192 of $2).
- **`run.sh regate-analyse`** (log `/tmp/x2d-regate-analyse.log`), exit 1: `regate/gates.json` SHA-256 `081be905f088fd0f67e15995166ea28279bea340c7bb7d451c4014f6c0532e7c`.

| Gate | Status | Decisive figures (Wilson 95 %) | Bars |
|---|---|---|---|
| P1 (stratum 1 decides; 35 eligible stratum-1 items of 38 eligible, 40 built; 2 excluded because the agreed gold excludes A: ctl-38, ctl-64) | **FAIL** | accuracy on answered stratum-1 items 14/20 = 0.700 [0.481, 0.855]; abstention 15/35 = 0.429 [0.280, 0.591]; B1 and B2 0/35 | accuracy ≥ 0.80; abstention ≤ 0.15; baselines ≤ 0.20 |
| P2 (12 direct-label positives of 40 clones; 11 gap, 1 new) | **PASS** | `none` on answered positives 9/11 = 0.818 [0.523, 0.949]; abstention 1/12 = 0.083 [0.015, 0.354] | `none` ≥ 0.80; abstention ≤ 0.15; positives ≥ 8 |

Reported figures (not decisive): P1 all eligible: recovery 15/38 = 0.395 [0.256, 0.553], A-only 14/38 = 0.368 [0.234, 0.527]; path-following on answered items 1/21 = 0.048 [0.008, 0.227], on the raw choice 6/38 = 0.158 [0.074, 0.304]; `none` answers on covered stratum-1 items 5; stratum 3 (one directory): 1/1 answered, 2/3 abstained. P2: `none` over all positives 9/12 = 0.750 [0.468, 0.911]; retained-covered clones 28: 18 assigned correctly, 0 wrong, 3 `none`, 7 abstained; derived vs direct gold {"covered|covered": 13, "covered|none": 15, "none|none": 12}.

**Outcome (PREREGISTRATION §5, rev 13, one shot):** the re-gate is not PASS for both P1 and P2, so the pilot stops before round 1. HD6 → `inconclusive`, reason: "round-0 gates P1 and P2 failed and the one permitted re-gate did not pass (re-gate P1 FAIL, P2 PASS); HD6 never under test". HD6b stays `proposed`. No round-0 critique is written; `RESULTS.md` follows.
