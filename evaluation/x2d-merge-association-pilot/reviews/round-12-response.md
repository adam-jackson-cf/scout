# X2d — drafter's response to council round 12

Date: 2026-10-30. Responds to `reviews/round-12.md` §2 (R12-01, R12-02, R12-03) and §4. The council's text was applied exactly as worded; no other byte moved in the frozen files except the DRAFT-NOTES §14 update the council asked for under R12-01 and note (ii). PREREGISTRATION stays **revision 12** (the council did not ask for a bump). No model call, no write to `../scout-x2d`, no commit. The round-0 re-freeze is preregistered only; it runs after the round-13 sign-off.

## 1. Per-issue rows

| id | file:line (new file) | text now |
|---|---|---|
| R12-01 | `PREREGISTRATION.md:242` (§14 item 19, "are reused, not re-run … , except that round 0's `version.json` is re-frozen once as follows.") | the council's "**Round-0 re-freeze (council round 12, R12-01).** …" paragraph verbatim: steps (1)–(5), exact keys allowed to differ `thresholds_sha256` and `question_version_keys`, `settings.json` byte-unchanged, "any other difference is a stop for a new council round" |
| R12-01 | `PREREGISTRATION.md:14` | "Also a one-time round-0 re-freeze of `rounds/0/version.json` after sign-off and before `gates-jev` (R12-01; §14 item 19)." |
| R12-01 | `DRAFT-NOTES.md:122` | "`run.sh`: one comment (`:248`) now says \"$20 since rev 12\". Round-0 re-freeze of `rounds/0/version.json` after sign-off (R12-01, §14 item 19). Not touched: …" (replaces the stale "Not touched: `run.sh`") |
| R12-02 (a) | `PREREGISTRATION.md:3` | the council's replacement sentence ("Round 0 has run through `gates-label` and part of `gates-jev` … the rev-12 decision used billed cost only.") |
| R12-02 (b) | `PREREGISTRATION.md:242` | "round-0 labels and gold already existed (`gates-label` complete; labels-check ok, κ_answer 1.0 on 50 units), thresholds were the provisional 0.7, no Jev answer had been compared with a label or gold and no predictor–outcome relationship existed" |
| R12-02 (b) | `docs/hypothesis-ledger.md:37` | the same clause, in parentheses, replacing "(no label, gold, tuned threshold or outcome existed)" |
| R12-02 (c) | `PREREGISTRATION.md:14`, `:192`, `:242`; `docs/hypothesis-ledger.md:37` | "after 7,755 of the 7,872 requests of the `main` run (64 units; the `n1` (50 units), `p1` (20) and `p2` (12) V-sem runs had not started)" (`PREREGISTRATION.md:3` states "7,755 answered requests of the round-0 `main` Jev run"); additionally `DRAFT-NOTES.md:122` carries the same qualifier |
| R12-02 (d) | — | attestation is the main agent's, in the stop record of `reviews/pilot-round-0.md` (not written here) |
| R12-03 (1) | `validation/validate.py:774-780` | `pr = os.path.join(tmp, "budget-data-pr")`, the two `wjl` rows (`usd_per_round - 0.5`, 0.20), `b_pr = x2d.jev_budget(3, …, pr)` with the council's comment, and `and abs(b_pr - (x2d.CAPS["usd_per_round"] - 0.20)) < 1e-9 and b_pr < x2d.CAPS["usd_total"] - 0.20` in the return expression |
| R12-03 (2) | `validation/validate.py:218` | `th["caps"]["usd_per_round"] == x2d.CAPS["usd_per_round"] == 2.0 and th["caps"]["usd_total"] == x2d.CAPS["usd_total"] == 20.0` |
| R12-03 (3) | `PREREGISTRATION.md:14` | "where two fixtures derive their expected budgets from `x2d.CAPS`, one adds a per-round-binding case, the cap values are pinned at `validate.py:218`; check count unchanged." (replaces "same properties, same check count"); `DRAFT-NOTES.md:122` likewise |

## 2. Validation (a)

`signoff.json` moved aside (`/tmp/signoff.json.aside`), `run.sh validate` run, file moved back. SHA-256 of `signoff.json` before and after: `e7059f2a9407e38c4d5e8aff0480913fab36a32a69ccbb9721c9633934fdbc2d` (equal). Result: **651 checks, 0 failed**, "all validation checks passed", exit 0 (check count unchanged; the new conjuncts are inside the existing checks). Log: `/tmp/x2d-validate-r12b.log`.

## 3. Frozen digests (computed after validation; 44 rows = the `signoff.json` `files` keys)

Tune pin `~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py`: `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c` = `signoff.json` `tune_script_sha256` (equal).

38 of 44 rows equal the round-11 `signoff.json`; the six marked changed are `PREREGISTRATION.md`, `README.md`, `run.sh`, `thresholds.json`, `validation/validate.py`, `x2d.py` (the council's set).

| file | SHA-256 | = signoff.json |
|---|---|---|
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` | yes |
| `PREREGISTRATION.md` | `fd30e0f4d4b540008a803a4d646c941ee14e6866768448ece72d5c10312d2a79` | **changed** |
| `README.md` | `fc51642bf03d55847a2086590932ee5bf44e64e8847b2cdc8eaaadd36e5d6124` | **changed** |
| `analysis_x2d.py` | `b0a66eedb739c2952b14f8468ae56da65d9628f6fa345f8dc51380520b397ade` | yes |
| `b2.py` | `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f` | yes |
| `baselines.py` | `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0` | yes |
| `build_controls.py` | `a71c4e634b5a0d0b42eda15f8e58220fef159c077fd3bf22d36041dd3e4857ab` | yes |
| `build_units.py` | `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243` | yes |
| `collect_onboard.py` | `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff` | yes |
| `compose.py` | `296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1` | yes |
| `frontend.py` | `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8` | yes |
| `gold.py` | `6b0afaad2a3ae6652a02207813c5a3492ba6954bd0a745b58a3f98c960bb321e` | yes |
| `jev/main.go` | `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0` | yes |
| `label_x2d.py` | `fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb` | yes |
| `measurement/lever-result.json` | `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c` | yes |
| `measurement/lever_check.py` | `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42` | yes |
| `onboard/main.go` | `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b` | yes |
| `power_sim.py` | `a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279` | yes |
| `prepass.py` | `1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696` | yes |
| `prepass/commits.jsonl` | `1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c` | yes |
| `prepass/k0.json` | `f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab` | yes |
| `prepass/reserve.json` | `a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a` | yes |
| `prepass/rounds.json` | `0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516` | yes |
| `prepass/scopes.json` | `a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512` | yes |
| `prepass/summary.json` | `e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1` | yes |
| `quality.py` | `6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade` | yes |
| `questions/merge.area_covers@v1.json` | `cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0` | yes |
| `questions/merge.covered@v1.json` | `f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9` | yes |
| `questions/merge.which_area@v1.json` | `cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928` | yes |
| `questions/onboard.covered@v1.json` | `ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70` | yes |
| `questions/onboard.same_responsibility@v1.json` | `31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1` | yes |
| `questions/onboard.which_file_area@v1.json` | `425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57` | yes |
| `reference_arm.py` | `d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739` | yes |
| `run.sh` | `f48cbf01a9e69c67f58c98e3c78880db814f7e28452587d0981176385d937595` | **changed** |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` | yes |
| `thresholds.json` | `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb` | **changed** |
| `tune.py` | `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c` | yes |
| `validation/r4_checks.py` | `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52` | yes |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` | yes |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` | yes |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` | yes |
| `validation/validate.py` | `a8ce75417cc883b0ad73d36b95c4536744271686c2da746859dfc8e02eeef7c4` | **changed** |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` | yes |
| `x2d.py` | `62e8cc1de402e221d6a57d04f9ebfc50c9e7bf4875ac5f3d65709830a60665ca` | **changed** |

## 4. `git diff 0b5659440c7368ac18b82f60a4edf3956408d2bf -- evaluation/x2d-merge-association-pilot docs/ --stat`

(run as `--stat -- <paths>`; untracked `reviews/*` are not in a diff against a commit)

```
 docs/hypothesis-ledger.md                          |  1 +
 docs/research-ledger.md                            |  1 +
 .../x2d-merge-association-pilot/DRAFT-NOTES.md     |  3 ++
 .../x2d-merge-association-pilot/PREREGISTRATION.md | 10 ++++---
 evaluation/x2d-merge-association-pilot/README.md   |  2 +-
 evaluation/x2d-merge-association-pilot/run.sh      |  2 +-
 .../x2d-merge-association-pilot/thresholds.json    |  6 ++--
 .../validation/validate.py                         | 33 +++++++++++++---------
 evaluation/x2d-merge-association-pilot/x2d.py      |  4 +--
 9 files changed, 38 insertions(+), 24 deletions(-)
```

Changed hunks (new-file line ranges; `-U0`):

| file | lines | content |
|---|---|---|
| `docs/hypothesis-ledger.md` | 37 | HD6 History line (R-052; R12-02 b, c) |
| `docs/research-ledger.md` | 60 | R-052 row (unchanged this round) |
| `DRAFT-NOTES.md` | 120-122 | §14 rev-12 note (R12-01 re-freeze, `run.sh` comment, 7,755 qualifier, fixture wording) |
| `PREREGISTRATION.md` | 3 | status line (R12-02 a) |
| `PREREGISTRATION.md` | 14 | change log rev 11 → 12 (R12-01 naming, R12-02 c, R12-03 wording) |
| `PREREGISTRATION.md` | 16 | ledger-sources line (R-052; unchanged this round) |
| `PREREGISTRATION.md` | 156 | §10 re-score charged to the $20 total (unchanged this round) |
| `PREREGISTRATION.md` | 192 | §12 Jev bullet (R12-02 c) |
| `PREREGISTRATION.md` | 242 | §14 item 19 (R12-01 re-freeze, R12-02 b, c) |
| `README.md` | 69 | rev-4 note total (unchanged this round) |
| `run.sh` | 248 | comment, $20 (unchanged this round) |
| `thresholds.json` | 181-183 | caps 2.0 / 20.0 / note (unchanged this round) |
| `validation/validate.py` | 218 | R12-03 (2) |
| `validation/validate.py` | 678 | check label (unchanged this round) |
| `validation/validate.py` | 768, 771-795 (as listed by `git diff -U0`) | `jev_budget_test` / `jev_budget_v7_test` (R12-03 (1) at 774-780; the rest as round 11) |
| `x2d.py` | 45, 521 | `CAPS`; docstring (unchanged this round) |

This round's own hunks: `PREREGISTRATION.md` 3, 14, 192, 242; `docs/hypothesis-ledger.md` 37; `DRAFT-NOTES.md` 120-122; `validation/validate.py` 218, 774-780. No rule, bar, threshold, unit, label or other check moved.
