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
