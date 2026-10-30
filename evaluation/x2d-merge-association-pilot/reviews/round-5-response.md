# X2d — drafter's response to council round 5 (`reviews/round-5.md`)

Date: 2026-10-30. Responds to the 12 issues of round 5 (2 blocking, 4 major, 6 minor) and the four rulings of its §2. PREREGISTRATION is now **revision 6**. Every fix uses the council's exact text where the round gives it. Nothing was committed; no model call, Jev, `claude`, `codex` or Clawpatch map was run; no file under `internal/`, `cmd/`, `scripts/`, `toolchain/`, `third_party/` or `go.*` and no `.codegraph` was touched. HD6 stays `proposed`.

Validation: `source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate` → **647 checks, 0 failed** (exit 0; round 4: 631). Three full runs (`/tmp/x2d-validate-r6.log`, `-r6b.log`, `-r6c.log`) gave the identical 647 / 0 and identical power figures (deterministic). The last full run (`-r6c`) is on every code file at the digests below. After it, one wording-only fix was made to `README.md` (rev-6 paragraph: the symbol names `analysis_x2d.GO` / `x2d.POWER_CHECK_KEYS`); `validation/validate.py` does not read `README.md` content (grep-checked), so no check can depend on it, but that README digest is not covered by a run.

## 1. Per-issue table

| id | sev | status | what was done (files) | evidence |
|---|---|---|---|---|
| R5-01 | blocking | fixed | Guard h replaced by the one-sided 95 % PR-cluster-bootstrap lower bound of `gain_guess` > 0, kind `guard`, INCONCLUSIVE, with the council's reason text. `analysis_x2d.py`: `score_outcome` adds `raw_correct` / `guess_correct` (`none` iff `raw_covered is False`, else `raw_choice`, scored by the usual rule; a missing raw answer is wrong), `pooling_rows` adds `d_guess`, `gain_guess_stat`, `summarise` emits `gain_guess {point, n, lb95, lb90, ci95, jev_only, base_only, mcnemar_p}` and keeps `gain_all`; `go_verdict` h reads `gain_guess["lb95"]`; `GO["guard_level"] = 0.95` = `x2d.GO_GUARD_LEVEL` = `thresholds.json study.go.guard_level` (asserted equal in `validate.py`). `compose.py`: raw fields on fan-out rows (max area `p_true` ≥ 0.5; argmax, ties by area id; None on a failure row). `power_sim.py`: `d_guess` per PR for every cluster, both bounds computed only when the verdict with both set to pass is GO, schema `x2d-power-v5`, new binding cells N4b-cap (mode `own_errors_cap`, p ∈ {0.80, 0.85}, ρ 0.5) and N4b-ρ0 (own errors, ρ 0, p ∈ {0.85, 0.88, 0.90}), `checks` = nine keys, `n4_max` over N4a, N4b, N4b-cap, N4b-ρ0. `thresholds.json` (`go.guard_level`, `go.guard`, `all_of[h]`, `power.schema`, `power.bars`). | the full power table (§2): every binding bar holds, bars unchanged; `validate.py` checks "S the power report is schema x2d-power-v5 …", "S the power cells cover the preregistered grids …", the nine "S binding power bar …" checks |
| R5-02 | blocking | fixed | (1) `validate.py`: a binding cell on the wrong side of its bar is a failed check (the "MISSED … passes" branch is deleted; `_bar_met` + a synthetic null cell at 0.12 is asserted NOT met). (2) `power_sim.main` exits 3 when any null check (N3, N3_PR, N4a, N4b, N4b_cap, N4b_rho0) is false, 0 when only P3 cells miss (asserted by `_power_main_exit` with a stubbed `run`). (3) `x2d.critique_problems(0)`: a false null check is a problem no line clears ("null cell X above its bar on the live structure: the GO rule is unsound here; new council round"); a false P3 check or a thin frozen round needs `user_decision: proceed`; `checks` must be exactly `x2d.POWER_CHECK_KEYS`. (4) text: `thresholds.json pre_run_stop`, `x2d.py` comments, `power_sim.py` docstring, PREREGISTRATION §5, §11, README. (5) fixtures in `validation/r4_checks.py` `critique_checks`: `N4b=False` (+ line and without) → one problem naming N4b, "unsound"; `N4b_cap=False` + line → refused; `P3_u10=False` + line → ok; a missing and an extra `checks` key → refused. `round0_lines` only adds the decision line for P3 / thin. | checks "S R5-02 the bar evaluation …", "R5-02 round-0 critique: …"; e2e `run.sh power` exits 0 on the stub world |
| R5-03 | major | fixed | Text restated under R5-01: §9 row h and the precedence sentence, §11 (no measured number; the u-bars 0.75 / 0.60 stated as the council's estimates, [INFERENCE]), §13 threat text with the estimand statement, §14 item 16 (b) and new item 17, `thresholds.json` (`go.guard`, `go.all_of[h]`, `inconclusive`, `reported_beside_the_verdict` + `gain_guess`, `gain_all`), README, change-log rev 5 → rev 6. | `PREREGISTRATION.md` lines 3, 8 (change log), 25, 140–147, 157–181, 187, 212, 232–237; a grep for the stale phrases finds only dated history marked superseded |
| R5-04 | major | fixed | `rewrite_arm(…, abstain_raw)` sets `raw_covered` / `raw_choice` consistently with `decide` (right → gold answer, wrong → a wrong area / assignment on gold none, abstain → "keep" / "right" / "wrong" or a callable). Worlds: gaming (`abstain_raw="wrong"`), near-null (raw right on exactly 2 hash-chosen abstained clusters), honest 10 % with a strong B2 (9 abstentions "keep", B2 rewritten right on 80 % hash-chosen clusters), 5 % ("keep") → GO, 20 % → NO-GO. §11 and README say what is done. | log lines "INFO gaming world …", "INFO near-null world …", "INFO honest 10 % world …"; checks "E R5-04 gaming world …", "E R5-04 near-null world …", "E R5-04 honest 10 % …" (figures in `DRAFT-NOTES.md` §10) |
| R5-05 | major | fixed | `run.sh clarify-precheck R` (new stage = `label_x2d.py --round R --precheck-clarify` alone, rounds 1–7, refuses after Jev scores or gold); order and recovery text in `run.sh`, README, §10 (v); fixture (`r4_checks.e2e_r4_rounds`): pre-check refuses (cap 1) and writes `clarify-refused.json`; guide edited → `gold 5` refused, no gold; frozen bytes restored → `labels-check 5` exits 0, `gold 5` proceeds, the scored round is non-informative. | check "E R5-05 `clarify-precheck 5` …" |
| R5-06 | major | fixed | `docs/hypothesis-ledger.md`: HD6 Hypothesis cell replaced with the council's text, new row HD6b (`proposed`, not tested by X2d), two History lines appended, text copied from round 5. §14 map: "The map applies to HD6 only; HD6b stays `proposed` whatever the verdict …"; §1 names HD6b and the guard-h wording. `docs/research-ledger.md` not edited. HD6 state unchanged (`proposed`). | `docs/hypothesis-ledger.md` lines 32, 33, 37, 38; PREREGISTRATION line 25 and 235 |
| R5-07 | minor | fixed | `controls.trap` adds `none_correct {k, n}` and `none_n`; `_planted_new_from_rounds` pools them (R_new = Σk/Σn, P_new = Σk/Σ none_n, Wilson 95 %, `unmeasured`, the council's `source` string); `thresholds.json new_area`, §9, §14 say so. | check "S R5-07 the trap controls carry the joint counts …" (a trap labelled covered counts in neither numerator) |
| R5-08 | minor | fixed | `report.md`: the all-cluster table is titled "Paired gain vs B2, all clusters (abstention = miss; `gain_all`)"; new rows `gain_res` (n_res, jev-only, B2-only, point, McNemar, 95 % CI, lb90, ub80, PRs with jev-only) and `gain_guess` (with lb95) from `_gain_rows`, also stored in `results.json abstention.gain_res_row / gain_guess_row`; the critique pack carries them in its abstention JSON and its paired heading says `gain_all`. | `analysis_x2d.py` `_gain_rows`, `render_report`, pack; the e2e run builds `results.json` and `report.md` through it (no failure) |
| R5-09 | minor | fixed | `_require_complete`: `round_results` and `verdict()` refuse (naming the cases) when a GO-population case lacks a primary-arm row or a B2 outcome; `verdict()` also refuses when `gain_guess.n`, `gain_all.n`, `unresolved.n`, `n` differ. | check "S R5-09 an unfinished GO population is refused, naming the case …" (B2 row of c5 removed) |
| R5-10 | minor | fixed | `simulate(…, trace=)` (validation hook) records each pilot's latent; a traced non-degenerate pilot (p_J 0.88, p_B2 0.80; random u 0.10, `own_errors`, `b2_errors`) is rebuilt as cases and arm rows with raw fields from the latent and `A.summarise` equals the simulation's figures; at u = 0 the `gain_guess` lb95 equals the `gain_res` 95 % bound over the same rows in the simulation and in `summarise`, and the `gain_res` lb90 values agree. | check "R5-10 simulate matches analysis_x2d.summarise on a non-degenerate traced pilot …" |
| R5-11 | minor | fixed | On a reversion to the provisional 0.7, `tune._tune_dev` composes the provisional decisions and `tune_round` writes `composed_dev_provisional` beside `composed_dev` (also carried to rounds 9 and 10 by the reuse path); no decision change; §10 (iv) one sentence. | check "R5-11 when the tuned decisions revert …" (key present on the cap fixture, absent when tuned decisions stay) |
| R5-12 | minor | fixed | Stale text restated: README rev-3/4 symbols marked "(rev 3–4; replaced in rev 5 by `simulate(..., mode=)`)", the writer of `clarify-refused.json` is `label_x2d.py --precheck-clarify` (README, §8, §10 (v), §12, `run.sh`), `power_sim.py` docstring, `x2d.py` comments, README bars list. | grep of the six edited files for `x2d-power-v4`, `routed to the round-0`, `gains nothing`, `simulate_inside`, `gain_all > 0`: only dated history marked superseded remains |

## 2. Full measured power table (empirical 91 clusters / 40 PRs, 10 positives inside, 1,000 pilots per cell, boot_B = 1,000; `run.sh validate`, deterministic for seed 20261029)

Binding, minimum GO share (bar ≥):

| cell | u | GO share (GO / 1000) | Wilson 95 % | verdicts GO / INCONCLUSIVE / NO-GO | bar | status |
|---|---|---|---|---|---|---|
| P3-u0 (p_J 0.95, p_B2 0.80, ρ 0.5) | 0 | 0.830 (830) | [0.805, 0.852] | 830 / 0 / 170 | ≥ 0.80 | holds |
| P3-u05 | 0.05 | 0.876 (876) | [0.854, 0.895] | 876 / 0 / 124 | ≥ 0.75 | holds |
| P3-u10 | 0.10 | 0.806 (806) | [0.780, 0.829] | 806 / 1 / 193 | ≥ 0.60 | holds |

Binding, maximum GO share (bar ≤ 0.10 in every cell):

| cell (p_J = p_B2) | u / mode | GO share (GO / 1000) | verdicts GO / INCONCLUSIVE / NO-GO |
|---|---|---|---|
| N3 0.50, 0.60 (u 0 and 0.10) | random, ρ 0 | 0.000 each (0) | 0 / 0 / 1000 |
| N3 0.70 | u 0 | 0.005 (5) | 5 / 1 / 994 |
| N3 0.70 | u 0.10 | 0.000 (0) | 0 / 1 / 999 |
| N3 0.80 | u 0 | 0.021 (21) | 21 / 2 / 977 |
| N3 0.80 | u 0.10 | 0.019 (19) | 19 / 7 / 974 |
| N3 0.85 | u 0 | 0.013 (13) | 13 / 0 / 987 |
| N3 0.85 | u 0.10 | **0.023** (23, max) | 23 / 8 / 969 |
| N3 0.90 | u 0 | 0.011 (11) | 11 / 0 / 989 |
| N3 0.90 | u 0.10 | 0.011 (11) | 11 / 2 / 987 |
| N3 0.95 | u 0 | 0.001 (1) | 1 / 0 / 999 |
| N3 0.95 | u 0.10 | 0.002 (2) | 2 / 0 / 998 |
| N3-PR 0.50, 0.60 (u 0 and 0.10) | random, ρ 0, PR latent | 0.000 each (0) | 0 / 0 / 1000 |
| N3-PR 0.70 | u 0 | 0.006 (6) | 6 / 2 / 992 |
| N3-PR 0.70 | u 0.10 | 0.003 (3) | 3 / 4 / 993 |
| N3-PR 0.80 | u 0 | 0.034 (34) | 34 / 9 / 957 |
| N3-PR 0.80 | u 0.10 | 0.031 (31) | 31 / 21 / 948 |
| N3-PR 0.85 | u 0 | **0.044** (44, max) | 44 / 17 / 939 |
| N3-PR 0.85 | u 0.10 | 0.039 (39) | 39 / 20 / 941 |
| N3-PR 0.90 | u 0 | 0.023 (23) | 23 / 5 / 972 |
| N3-PR 0.90 | u 0.10 | 0.017 (17) | 17 / 12 / 971 |
| N3-PR 0.95 | u 0 | 0.003 (3) | 3 / 2 / 995 |
| N3-PR 0.95 | u 0.10 | 0.007 (7) | 7 / 0 / 993 |
| N4a 0.80, 0.85, 0.90, 0.95 | `b2_errors` (cap 13 clusters) | 0.000 each (0) | 0 / 0 / 1000 |
| N4b 0.85 | `own_errors`, ρ 0.5 | **0.047** (47) | 47 / 254 / 699 |
| N4b 0.88 | same | 0.043 (43) | 43 / 232 / 725 |
| N4b 0.90 | same | 0.026 (26) | 26 / 136 / 838 |
| N4b 0.95 | same | 0.004 (4) | 4 / 5 / 991 |
| N4b-cap 0.80 | `own_errors_cap`, ρ 0.5 | **0.061** (61, `n4_max`) | 61 / 400 / 539 |
| N4b-cap 0.85 | same | 0.051 (51) | 51 / 397 / 552 |
| N4b-ρ0 0.85 | `own_errors`, ρ 0 | 0.055 (55) | 55 / 407 / 538 |
| N4b-ρ0 0.88 | same | 0.054 (54) | 54 / 515 / 431 |
| N4b-ρ0 0.90 | same | 0.044 (44) | 44 / 409 / 547 |

`n4_max` (over N4a, N4b, N4b-cap, N4b-ρ0) = 0.061. Rev 5 under the point guard on `gain_all`: N4b 0.255 / 0.228 / 0.141 / 0.006. All nine binding checks are true (`all_pass`); the validation passes every binding bar, none was moved.

Report-only: N4c (p_J 0.90, p_B2 0.80, own-error abstention) GO 0.661 [0.631, 0.690], verdicts 661 / 220 / 119; P3-PR 0.812 [0.787, 0.835], 812 / 12 / 176; P3-ρ0 0.853 [0.830, 0.874], 853 / 3 / 144; δ = 0.15 curve (u 0.05) p_B2 0.70 / 0.75 / 0.80 / 0.85 → GO share 0.479 / 0.787 / 0.876 / 0.910.

## 3. Choices taken where the council offered one (recommendation adopted)

- **Guard:** the experiment-peer's G4 (`gain_guess`, lb95) as adopted in §2.1, not the lead's initial G1 (a 90 % bound on `gain_all`).
- **HD6:** the narrowed sentence plus a separate row HD6b (the council's recommendation) rather than one narrowed sentence.
- **Null vs P3 cells:** exit 3 and a critique stop for nulls; `user_decision: proceed` only for P3 cells and a thin frozen round.
- **Departures of round 4 as ruled:** `clarify-refused.json` writer accepted (with R5-05, R5-12); composed-row abstention worlds accepted (with R5-04); one additive pooling row set accepted; `power_sim` exit 0 rejected for null cells; `planted_new` as labelled rejected → R5-07.
- Drafter judgement calls for round 6 are in `DRAFT-NOTES.md` §10 (honest-10 % world sets Jev right on every answered cluster; the `trace` hook; where `composed_dev_provisional` is built; the two completeness refusals; `clarify-precheck` skips `require_version`; `POWER_*_KEYS` home).

## 4. Digests (SHA-256, recomputed 2026-10-30 after the last edit; `x2d.frozen_files()` plus `DRAFT-NOTES.md`)

| File | SHA-256 |
|---|---|
| `DRAFT-NOTES.md` | `f6bcf2f56c07fcab6d87a680825773b1bcc0471b9c40a76db0e53994bbabeb72` |
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `20354966575e3edca1776128322d906f241fcab5e6c111d83f6cae8a36f3dcb5` |
| `README.md` | `7ede45b8b3a6d74b3cfadebae16b20267447e3664429b51405947b85f03b60d1` |
| `analysis_x2d.py` | `ffbb3578d37950282144be8ee60b87b96053332a739cf3eecaf361e264c20021` |
| `b2.py` | `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f` |
| `baselines.py` | `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0` |
| `build_controls.py` | `7ded8c9177d8e21131ed91ee0aa423a4dd8c64ac35b215f82e93b1bfbfe6c81b` |
| `build_units.py` | `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243` |
| `collect_onboard.py` | `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff` |
| `compose.py` | `296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1` |
| `frontend.py` | `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8` |
| `gold.py` | `6b0afaad2a3ae6652a02207813c5a3492ba6954bd0a745b58a3f98c960bb321e` |
| `jev/main.go` | `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0` |
| `label_x2d.py` | `82fc06bede3b93ee7348d557290d4ed761d77c4d953e80cb981f00662f5250c4` |
| `measurement/lever-result.json` | `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c` |
| `measurement/lever_check.py` | `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42` |
| `onboard/main.go` | `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b` |
| `power_sim.py` | `a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279` |
| `prepass.py` | `1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696` |
| `prepass/commits.jsonl` | `1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c` |
| `prepass/k0.json` | `f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab` |
| `prepass/reserve.json` | `a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a` |
| `prepass/rounds.json` | `0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516` |
| `prepass/scopes.json` | `a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512` |
| `prepass/summary.json` | `e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1` |
| `quality.py` | `6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade` |
| `questions/merge.area_covers@v1.json` | `cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0` |
| `questions/merge.covered@v1.json` | `f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9` |
| `questions/merge.which_area@v1.json` | `cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928` |
| `questions/onboard.covered@v1.json` | `ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70` |
| `questions/onboard.same_responsibility@v1.json` | `31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1` |
| `questions/onboard.which_file_area@v1.json` | `425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57` |
| `reference_arm.py` | `d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739` |
| `run.sh` | `22913037722570352ae59f391dc2279da2e7aa8ee4bba2ba01d2c4f3f188812b` |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` |
| `thresholds.json` | `634618ce4f32e2e0c87e719573c20ecb857f22312440c53fbc076264fadd5ead` |
| `tune.py` | `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c` |
| `validation/r4_checks.py` | `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `f3d72623871fde720e3666b4c42bdca316ee79a086a07c3c3534f4a694f243ae` |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` |
| `x2d.py` | `32c093e972cff31e86dae51412e948c70f07348465d299e18036813f82c8d464` |

External pin: `tune_threshold.py` (`~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py`) = `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c` (recomputed 2026-10-30; unchanged since round 5).

Files changed this round outside the study directory: `docs/hypothesis-ledger.md` (R5-06 only). Unchanged: `GUIDE.md`, `questions/*.json`, the Go drivers, `label_x2d.py`, `versioning.py`, `docs/research-ledger.md`.

## 5. For the council / main agent

- Re-run `source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate` once after any further edit (expect 647 checks, 0 failed with these files; every binding bar holds, nulls included).
- The stub-world `run.sh power` exits 0 (no null cell above its bar), asserted in validation. The live round-0 `power` stage can exit 3; that stop is not clearable by `user_decision`.
- Sign-off remains conditional on the main agent recomputing the digests, obtaining a green validation, committing the frozen files, `signoff.json`, the code-under-test blobs and all review rounds and responses in one run commit (round 5 §5).
