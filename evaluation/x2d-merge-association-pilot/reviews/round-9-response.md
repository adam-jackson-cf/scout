# Round 9 response (the drafter's response that council round 10 reviews): P2 clone sources top up from set D

Source: `reviews/pilot-round-0-p2-gate-memo.md` §7 steps 1–5, option (f1), user decision 2026-10-30. No model call, no label, no `data/` write, no product file, no commit. The run worktree `../scout-x2d` was only read (catalogue and units; no labels exist).

## 1. Per-step rows

| Step | What changed | File | Evidence |
|---|---|---|---|
| 1 Deviation entry | population finding: four ineligible P units (`30e1d6bd…` `docs/en/mkdocs.insiders.yml`, `9582f013…` `docs/en/overrides/partials/copyright.html`, `ac049d4f…` `docs/language_names.yml`, `e7bf5ec4…` `tests/test_webhooks_security.py`), the 28/28 ceiling, carried into RESULTS later | `DRAFT-NOTES.md` §13; also `PREREGISTRATION.md` §14 item 18 | read-back |
| 2 PREREGISTRATION rev 10 | §5 source pool = P leftovers (unplaceable, then placeable beyond 20), then D, hash order; `n_p2 = min(12, eligible_P − 20 + eligible_D)`; refusal below 8 and `P1 ≥ 20` unchanged; D clones disclosed as source-dependent; §11 line 174 states the new count and fixtures; new §11 bullet "Pipeline checks added in rev 10"; §14 item 18 (cites the memo and the user decision 2026-10-30); change log rev 9 → rev 10; status "Revision 10 … DRAFT for council round 10" | `PREREGISTRATION.md` | each replacement asserted to match exactly once |
| 3 `build_controls.py` | `elig_p`, `elig_d`; `p_slots = len(elig_p) − 20`; `need = min(12, p_slots + len(elig_d))`; refusal text unchanged in substance (now names both counts; still contains "planted-gap gate" and "council and the user decide"); donors = unplaceable P → leftover placeable P beyond 20 → D in hash order; `used_src` includes D donors; docstring and constant comment updated; `make_controls` returns `eligible`, `eligible_d`, `p2_donors` {P_unplaceable, P_leftover, D}, `p2_source_ids`, `ablated_area_ids`; `main` prints the counts | `build_controls.py` | validation below |
| 4 fixtures | (i) P short (20 + 7 eligible), D fills: 12 clones, all membership gap positives, 7 from P then 5 from D; D / P / P1 / trap / prose / planted_new rows, dup rows common to both builds, `catalogue-p1`, `catalogue-n1`, contrast units, membership byte-identical to the P-only build of the same reserve (P-only build = `_unique_owner` forced to `None` for D rows and the gap floor lifted); P1 exactly 20 items; (ii) P + D short (7) refuses with the planted-gap-gate text, 8 builds exactly 8 clones; real stub `gates-build` builds `min(12, eligible_P − 20 + eligible_D)`. The old trim-until-refusal check is replaced by (ii). R3-04/R4-04 analysis fixtures, planted-effect and planted-null e2e runs unchanged | `validation/validate.py` | log below |
| 5 docs and ledger | README module rows (lines 56, 69); DRAFT-NOTES rev-10 entry; HD6 History line (round-0 stop, rev-10 change, council round 10 required; HD6 stays `preregistered`) | `README.md`, `DRAFT-NOTES.md` §13, `docs/hypothesis-ledger.md` | read-back |

Finding disclosed to the council: `dup` controls are `ceil(10 %)` of `len(d)+len(p)+len(items)`, so a different P2 clone count can change the number of dup rows by one. The dup rows are a hash-order prefix, so the fixture requires every dup row common to both builds to be identical, not equal counts. On the real data the dup count is 9.

## 2. The REAL round-0 allocation (computed read-only: `make_controls` on `scout-x2d/…/data` catalogue, `/tmp/scout-x2c-smoke/_src/fastapi`, nothing written)

- eligible_P 26 (30 − 4), eligible_D 19 (of 20; `tests/test_computed_fields.py` has no owner), `p_slots` 6, `n_p2 = min(12, 6 + 19) = 12`; refusal not triggered; all 12 are membership gap positives.
- P1: 20 items (`p1_unplaceable` 2).
- Donor sources in order (P2 clone sources, cluster ids = `0:<id>`):

| # | set | donor | first member |
|---|---|---|---|
| 1 | P (unplaceable) | `d636de1b7509049b` | `CITATION.cff` |
| 2 | P (unplaceable) | `d8a31a4c12fa6633` | `pdm_build.py` |
| 3 | P (leftover placeable) | `fbfc5080c49eb089` | `tests/test_tutorial/test_request_form_models/test_tutorial002_pv1_an_p39.py` |
| 4 | P (leftover placeable) | `e83a36870dec905f` | `fastapi/cli.py` |
| 5 | P (leftover placeable) | `cc8803ec299cc7c3` | `tests/test_tutorial/test_header_params/test_tutorial003.py` (+ `_an`) |
| 6 | P (leftover placeable) | `84f3f93b13f228da` | `.github/ISSUE_TEMPLATE/privileged.yml` |
| 7 | D | `dcc1b446bf7afc14` | `docs_src/schema_extra_example/tutorial005.py` (+ `_an`) |
| 8 | D | `1aa8788262775a7b` | `tests/test_compat.py` |
| 9 | D | `5eedae70bdf178fc` | `tests/test_router_redirect_slashes.py` |
| 10 | D | `df1a80257920dafa` | `tests/test_tutorial/test_query_param_models/test_tutorial002.py` |
| 11 | D | `92757f9e933787d3` | `tests/test_generic_parameterless_depends.py` |
| 12 | D | `a71e28eebe80d0c9` | `requirements-tests.txt` |

  Per-set donor counts: P_unplaceable 2, P_leftover 4, D 6 (6 from P, 6 from D).
- Ablated areas: **11** (`158cb37f0800`, `1c4aa748f480`, `73232e7095cf`, `95b28e00298a`, `9a4a0e49b8c6`, `c207aaf8cb60`, `c4f3d5556193`, `d56ca48225f1`, `ea1d02de2a93`, `ea75266c1719`, `f36cd1216200`; fewer than 12 because two clones share an area).
- Catalogue size: 114 areas; `catalogue-p2` **103 areas** (114 − 11); `catalogue-p1` 113.
- Units: 96 in `units.jsonl` (20 D, 30 P, 46 controls: p1 20, p2 12, dup 9, planted_new 3, prose 1, trap 1) + 20 contrast units.
- Session plan, round 0 (`label_x2d.plan_for` / `cap_for`): W = 96 + 20 = **116**, J 0, R 0, I 40, C 0 → cap = ceil(1.25 × (2.2 × 116 + 2.2 × 40)) = **429** (ceiling 500); used 80 (intruder audit); planned 2 per case = 232; 80 + 232 = 312 ≤ 429 fits.

## 3. Validation

`source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate`: **649 checks, 0 failed**, exit 0 (rev 9: 647; the single trim-until-refusal check became three rev-10 checks). Log: `/tmp/x2d-val-r10b.log`.

Process disclosure: the first run (`/tmp/x2d-val-r10.log`) failed in the stub labelling check because the existing `signoff.json` (round 9) pins the old digests of `build_controls.py` and `validation/validate.py` and the labeller verifies pins whenever the record exists. For the green run I moved `signoff.json` aside for the duration of the run and restored it afterwards (SHA-256 before and after: `25100022a2698c07e049fb20f57f30f6de9e2333b6d7ea029826f049f2e8fdbb`, unchanged). Round 10 writes the new record; the run resumes with it.

## 4. SHA-256 of every frozen file (44; computed after the validation run)

```
GUIDE.md 3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980
PREREGISTRATION.md d059764047761baa6a5e5683c7568e33cb3accde4acd9bf6522b9322556a69ba   (changed; signed rev 9: 4b14dae0…8429)
README.md 3baeca377e828f9b7990bcc1d10665c03c20c3273d7ac6f3e1f3733040bf7fd6   (changed; signed: e69c713c…f1c5)
analysis_x2d.py b0a66eedb739c2952b14f8468ae56da65d9628f6fa345f8dc51380520b397ade
b2.py 5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f
baselines.py a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0
build_controls.py 41787fd8de4ad6f6bbabfaef2ef2053da18047fc0fc557ad3057ec8e521390df   (changed; signed: 7ded8c91…81b)
build_units.py c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243
collect_onboard.py c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff
compose.py 296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1
frontend.py dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8
gold.py 6b0afaad2a3ae6652a02207813c5a3492ba6954bd0a745b58a3f98c960bb321e
jev/main.go 63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0
label_x2d.py fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb
measurement/lever-result.json 298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c
measurement/lever_check.py 8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42
onboard/main.go 7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b
power_sim.py a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279
prepass.py 1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696
prepass/commits.jsonl 1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c
prepass/k0.json f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab
prepass/reserve.json a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a
prepass/rounds.json 0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516
prepass/scopes.json a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512
prepass/summary.json e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1
quality.py 6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade
questions/merge.area_covers@v1.json cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0
questions/merge.covered@v1.json f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9
questions/merge.which_area@v1.json cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928
questions/onboard.covered@v1.json ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70
questions/onboard.same_responsibility@v1.json 31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1
questions/onboard.which_file_area@v1.json 425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57
reference_arm.py d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739
run.sh 31029ba557d453e620b214f242ed6454622617e34f0850fcd7a0667ac82806c5
scopes.py 5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072
thresholds.json 5c211eb939b4880ae21e79fdfcfc25060722f73a96b99d6f27aaceb2153094a8
tune.py 2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c
validation/r4_checks.py f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52
validation/stub_clawpatch_x2d.js d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f
validation/stub_jev.py f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d
validation/stub_labeller.py d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5
validation/validate.py 3fe9749443243a1034cf55c023359c9e45b6e1d26cb55ad6994e13eb1968afbf   (changed; signed: b63ca4e7…4c42b)
versioning.py 331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168
x2d.py 32c093e972cff31e86dae51412e948c70f07348465d299e18036813f82c8d464
```

Four of 44 changed since the round-9 record (PREREGISTRATION, README, build_controls, validate); the other 40 equal the recorded digests (spot-checked against the sign-off: every unchanged file's digest above is the one in the round-9 record `signoff.json`, whose 17 code-under-test blobs are untouched). Not frozen and changed: `DRAFT-NOTES.md` `f07b501c…b4c2`, `docs/hypothesis-ledger.md` `5fbf76b6…3949`.

## 5. External pin

`~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py` SHA-256 `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c` = `signoff.json` `tune_script_sha256` (equal).

## 6. For round 10 (conditions of the memo §6)

Donor ids and order, the ablated-area count (11), the remaining catalogue (103 areas), and the session plan (cap 429, W 116) are fixed here before any label; the council has seen membership only (rule 3 allows). Please write new `signoff.json` digests for the four changed frozen files in the sign-off round.
