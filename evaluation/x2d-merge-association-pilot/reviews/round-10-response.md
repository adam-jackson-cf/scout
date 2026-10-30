# Round 10 response (the drafter's response that council round 11 reviews): R10-01 … R10-03, notes (a)–(e), PREREGISTRATION revision 11

Source: `reviews/round-10.md` §2–§3. No model call, no label, no `data/` write, no run-worktree write, no commit, no product file (`internal/`, `cmd/`, `scripts/`, `toolchain/`, `third_party/`, `go.*` untouched). `../scout-x2d/…/data/` (catalogue, units, prepass) was only read, in memory, for membership; no labels exist.

## 1. Per-step rows (line numbers are in the files as they now stand)

| Id | What changed | File:line | Exact text / evidence |
|---|---|---|---|
| R10-01 code, placeable loop | `c1 = _filter_catalogue(cat, removed \| set(r["members"]))` is computed first; if `a` is not an `area_id` of `c1["areas"]` (A would lose its last owned file) the pair is appended to `unplaceable` and the loop `continue`s; otherwise `o` is built from `c1` as before | `build_controls.py:264-275` | `if a not in {x["area_id"] for x in c1["areas"]}:  # A would lose its last owned file: Jev could not recover an area that is not in catalogue-p1` |
| R10-01 code, stabilisation loop | an item is also `bad` when its gold area is not an area of `cat_p1` | `build_controls.py:298` | `bad = [i for i, t in enumerate(placements) if t is None or p1[i][1] not in {x["area_id"] for x in cat_p1["areas"]}]`; no other allocation line changed |
| R10-01 §5 sentence | council's text inserted after "remove an owned file of area A from the catalogue" | `PREREGISTRATION.md:106` | "(a P unit is a P1 candidate only if A keeps at least one owned file in `catalogue-p1` after every P1 removal — Jev could not recover an area that is not in the catalogue; a unit that would empty its area goes to the P2 source pool as unplaceable)" |
| R10-01 record | change log rev 10 → rev 11 (area `26ac01437eaf`, the two `image01.py` clusters, 113 of 114); status line; §14 item 18 (population finding, to be stated in RESULTS §0) | `PREREGISTRATION.md:3,13,240` | status "Revision 11 … DRAFT for council round 11" |
| R10-01 §5 pool wording | "unplaceable ones first — including a unit that would empty its area in `catalogue-p1` (R10-01) — then placeable ones beyond the 20 of P1" | `PREREGISTRATION.md:107` | |
| R10-01 fixture (α) | on the real stub `gates-build`: ≥ 20 `kind == "p1"` rows, every `area_id` in `rounds/0/catalogue-p1.json`, and `catalogue-p1` has exactly as many areas as the catalogue | `validation/validate.py:2663-2668` | log line 413: "(alpha) … 20 P1 items, … exactly as many areas as the catalogue (15)" PASS |
| R10-01 fixture (β) | planted catalogue (patched `x2d.read_catalogue`, restored in `finally`): the area of a real P1 source unit is reduced to that unit's members; the unit is not a P1 item, is in `p2_source_ids`, P1 still has 20 items, `catalogue-p1` keeps every area (names and count) and every P1 gold area is in it | `validation/validate.py:2669-2695`; `import copy` at `:23` | log line 414 PASS |
| R10-01 fixture (γ) | rev-10 fixtures (i)/(ii) unchanged | `validation/validate.py:2600-2661` | log line 412 PASS (the other rev-10 check lines PASS) |
| R10-02 | P2 descriptive disclosures preregistered, descriptive only, gate stays pooled: donor counts by set (`P_unplaceable` / `P_leftover` / `D`); ablated area ids, count, `catalogue-p2` size out of the catalogue's areas; clones sharing an ablated area; gap positives, `none` answers, `none_correct_gap`, `retained_covered`, `no_agreed_gold` split P vs D joined from `control-key.jsonl` `source_case_id` → `units.jsonl` `set`, `gold.jsonl`, `arms-p2/V-sem.jsonl`, join command as reproduce command; path twins (D/P, P/P) | `PREREGISTRATION.md:107` | no analysis code change |
| R10-03 | the step-6 record is recomputed in §2 below; procedure (`gates-build` alone, compare `control-key.jsonl` fields, `catalogue-p*.json` area counts, `units*.jsonl` line counts, `catalogue.sha256`; record in `reviews/pilot-round-0.md`; mismatch = new council round, never a re-build; only then `gates-label`, `gates-jev`, `gates-analyse`) | `PREREGISTRATION.md:240` (item 18), `DRAFT-NOTES.md:119` | no code change |
| (a) | "planted-effect and planted-null **checks** are unchanged; the stub world now builds 12 clones (8 P, 4 D)" | `PREREGISTRATION.md:12` (change log rev 9 → 10), `:187` | `:187` adds eligible_P 28, eligible_D 20 and "exercises the D path through gold and analysis end to end" |
| (b) | the `:186` wording at the fixture sentence: "the dup rows common to both builds, … identical to the P-only build, with the sequential control ids `0:ctl-<n>` of the controls dropped" (in place of "byte-identical") | `PREREGISTRATION.md:176` (§11 P2 ablation), `:187` | |
| (c) | "the duplicate controls are a hash-order prefix whose length can change by one with the P2 count (`ceil(10 %)`), and on the frozen catalogue it is 9 either way" | `PREREGISTRATION.md:107` | replaces "the duplicate controls are unaffected" |
| (d) | "P beyond 20 plus eligible D = 7 gives the same refusal (the fixture's 7 case has 0 eligible D)" | `PREREGISTRATION.md:176,187` | |
| (e) | "Two commits": onboarding, catalogue freeze, quality and intruder audit ran at `730674e` under the rev-9 sign-off, everything from `gates-build` at the rev-11 commit; data reused, not re-run, because every file those stages read has the same digest under both sign-offs and `rounds/0/version.json` had not been written; the worktree moves by checkout only (no `git clean`, no `stash -u`; `data/` untracked); the same sentence goes to RESULTS `Deviations` | `PREREGISTRATION.md:240`, `DRAFT-NOTES.md:119` | |
| README | module row: a P unit whose area would lose its last owned file in `catalogue-p1` is a P2 source, not a P1 item | `README.md:56` | |
| DRAFT-NOTES, ledger | rev-11 entry; HD6 History line (HD6 stays `preregistered`) | `DRAFT-NOTES.md:119`, `docs/hypothesis-ledger.md:37` | |

## 2. The recomputed real round-0 allocation (R10-03)

Computed by calling `build_controls.make_controls(SRC, data)` unchanged (the code of this revision) with `data = ../scout-x2d/evaluation/x2d-merge-association-pilot/data` and `SRC = /tmp/scout-x2c-smoke/_src/fastapi`, in memory; nothing was written. Membership only. The script is `/tmp/x2d_r10_alloc.py` (not committed).

- Catalogue digest: SHA-256 `1220016c26115ccd3e3a521de76ec45e68893f4ab6edf0f4889e8c8ce2c84199` (file digest equals `data/catalogue/catalogue.sha256`); 114 areas.
- eligible_P 26, eligible_D 19, `p_slots` 6, `n_p2 = min(12, 6 + 19)` = 12; refusal not triggered; 12 of 12 membership gap positives.
- `p1_unplaceable` (the value `build_controls` prints) **4**; P1 20 items; donors `{P_unplaceable 4, P_leftover 2, D 6}`.
- The 20 P1 source cluster ids, in order (`0:<id>`): `ad64cd8d312c55f5`, `3bd1063c5c57cd0a`, `634a3315f5e624fb`, `7e510eda2a7de453`, `e13b5aaa4e5cf4b2`, `de6004542631febf`, `373fe757150aa6cc`, `7e381413bc181529`, `1f52dcef64ad8251`, `3ea28ee9c73e0be5`, `60c20dc7c517a2c4`, `4ffa0cf617d9d2d0`, `de57771884770f8b`, `38886da598e1d4fd`, `fecbae9a3e4d6536`, `723abe0752a9cd76`, `bda457209e0f1fa1`, `c73cc2eb10ae3269`, `fbfc5080c49eb089`, `e83a36870dec905f`.
- The 12 P2 source cluster ids in order (`p2_source_ids`):

| # | set | donor | first member |
|---|---|---|---|
| 1 | P (unplaceable) | `2c4db77221b8e719` | `scripts/playwright/request_form_models/image01.py` |
| 2 | P (unplaceable) | `d636de1b7509049b` | `CITATION.cff` |
| 3 | P (unplaceable) | `d8a31a4c12fa6633` | `pdm_build.py` |
| 4 | P (unplaceable) | `af2f04c8c181c116` | `scripts/playwright/request_form_models/image01.py` |
| 5 | P (leftover placeable) | `cc8803ec299cc7c3` | `tests/test_tutorial/test_header_params/test_tutorial003.py` (+ `_an`) |
| 6 | P (leftover placeable) | `84f3f93b13f228da` | `.github/ISSUE_TEMPLATE/privileged.yml` |
| 7 | D | `dcc1b446bf7afc14` | `docs_src/schema_extra_example/tutorial005.py` (+ `_an`) |
| 8 | D | `1aa8788262775a7b` | `tests/test_compat.py` |
| 9 | D | `5eedae70bdf178fc` | `tests/test_router_redirect_slashes.py` |
| 10 | D | `df1a80257920dafa` | `tests/test_tutorial/test_query_param_models/test_tutorial002.py` |
| 11 | D | `92757f9e933787d3` | `tests/test_generic_parameterless_depends.py` |
| 12 | D | `a71e28eebe80d0c9` | `requirements-tests.txt` |

- Ablated areas: **10** (`26ac01437eaf`, `73232e7095cf`, `95b28e00298a`, `9a4a0e49b8c6`, `c207aaf8cb60`, `c4f3d5556193`, `d56ca48225f1`, `ea1d02de2a93`, `ea75266c1719`, `f36cd1216200`); two pairs of clones share an area (computed from `control-key.jsonl` `ablated_area_id`): the two `image01.py` clones (`2c4db772`, `af2f04c8`; a P/P path twin) share `26ac01437eaf`, and `d636de1b` (`CITATION.cff`) and `84f3f93b` (`.github/ISSUE_TEMPLATE/privileged.yml`) share `73232e7095cf`; 12 donors over 10 areas.
- Area counts: `catalogue-p1` **114**, `catalogue-p2` **104**, `catalogue-n1` 114. Every P1 gold area is in `catalogue-p1` (computed: true).
- Kind counts (`control-key.jsonl`): p1 20, p2 12, dup 9, planted_new 3, prose 1, trap 1 (46 controls); `units.jsonl` 96 (20 D, 30 P, 46 controls) + `units.contrast.jsonl` 20 → W = 116, J 0, R 0, I 40, C 0, cap 429 (unchanged; used 80 by the intruder audit; 80 + 232 = 312 ≤ 429).

**Changed from `round-9-response.md` §2:**

| Quantity | round-9 | now |
|---|---|---|
| `p1_unplaceable` | 2 | **4** |
| P2 donors by set | P_unplaceable 2, P_leftover 4, D 6 | **P_unplaceable 4, P_leftover 2, D 6** |
| P2 source ids, order | `d636de1b`, `d8a31a4c`, `fbfc5080`, `e83a3687`, `cc8803ec`, `84f3f93b`, then D | **`2c4db772`, `d636de1b`, `d8a31a4c`, `af2f04c8`, `cc8803ec`, `84f3f93b`**, then the same six D (donors 7–12 unchanged) |
| P1 sources | previous set incl. `2c4db772`, `af2f04c8` | the previous 18 plus **`fbfc5080`, `e83a3687`** (those two moved from P2 leftovers into P1; `2c4db772`, `af2f04c8` left P1) |
| ablated areas | 11 (…, `158cb37f0800`, `1c4aa748f480`, …) | **10**: `26ac01437eaf` added; `158cb37f0800`, `1c4aa748f480` (areas of `fbfc5080`, `e83a3687`) removed; the other 8 equal |
| `catalogue-p2` | 103 | **104** |
| `catalogue-p1` | 113 | **114** |

Unchanged: eligible_P 26, eligible_D 19, `p_slots` 6, `n_p2` 12, 12 of 12 gap positives, P 6 / D 6 split, the six D donors and their order, dup 9, planted_new 3, prose 1, trap 1, 96 units, 20 contrast, W 116, cap 429, `catalogue-n1` 114.

## 3. Validation

`source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate` once, with the stale round-9 `signoff.json` moved aside and restored afterwards (SHA-256 before and after `25100022a2698c07e049fb20f57f30f6de9e2333b6d7ea029826f049f2e8fdbb`, unchanged): **651 checks, 0 failed**, exit 0 (round 9 response: 649; +2 = fixtures α and β). Log: `/tmp/x2d-val-r11.log` (lines 413–414 are α and β; 412 is the rev-10 refusal / exactly-8 check). The sign-off run (condition of round 11) is with the new record in place.

## 4. SHA-256 of every frozen file (44; computed after the validation run, no frozen file edited since)

```
GUIDE.md 3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980
PREREGISTRATION.md b5029a6fede44f3c88438f04075b86f3ff24f5ab9afdb19cfcf61865ecfdec22   (changed; replaces round-9 response d0597640…69ba / signed rev 9 4b14dae0c3875eddd98667bf6da3935785f4817b209dd5d78145bbc0dca38429)
README.md 0b9fc29c8a7d18ae39df21e7340cd04d4ade10bdfbe73955ee6b9041b8528959   (changed; replaces 3baeca37…7fd6 / signed e69c713c7ef486d5023460a0ed2dbf4d1881982c5cc11fbf39962fc38019f1c5)
analysis_x2d.py b0a66eedb739c2952b14f8468ae56da65d9628f6fa345f8dc51380520b397ade
b2.py 5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f
baselines.py a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0
build_controls.py a71c4e634b5a0d0b42eda15f8e58220fef159c077fd3bf22d36041dd3e4857ab   (changed; replaces 41787fd8…90df / signed 7ded8c9177d8e21131ed91ee0aa423a4dd8c64ac35b215f82e93b1bfbfe6c81b)
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
validation/validate.py b9fc5de8d47a5a45077953bf72568ebdc30451e91eaed3a458c35e2910e08396   (changed; replaces 3fe97494…bf / signed b63ca4e7e606e7284ad9f565c02655554cfe81873950ae8227ef44946ca4c42b)
versioning.py 331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168
x2d.py 32c093e972cff31e86dae51412e948c70f07348465d299e18036813f82c8d464
```

Four of 44 changed (the same four files as round 9: PREREGISTRATION, README, build_controls, validate); the other 40 equal the round-9 `signoff.json` row for row (compared by script against `signoff.json` `files`; every unchanged row printed with no "CHANGED"). The "replaces" digest in each changed row is the round-9 *response* digest (abbreviated) and the signed rev-9 digest (full, from `signoff.json`). Not frozen and changed: `DRAFT-NOTES.md` `7ed7da237004da821fd97a469b1d6dc64111b121cabf48ad16439685426474c6` (round-9 response: `f07b501c…b4c2`), `docs/hypothesis-ledger.md` `e9a59db1ab7c7054382b0a204a704df271c9fe06a84f752003e376ea821b65d4` (round-9 response: `5fbf76b6…3949`). `reviews/round-10-response.md` itself is a review file and not in the frozen set.

## 5. External pin

`~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py` SHA-256 `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c` = `signoff.json` `tune_script_sha256` (equal).

## 6. For round 11

Round-10 §3 item 5 (the `image01.py` P/P path twin and the `CITATION.cff` / `privileged.yml` shared-area pair are membership facts recorded by the council) is left to the council's record. The sign-off conditions of round-10 §4 apply: new `signoff.json`, one green `run.sh validate` with it in place after the last frozen-file edit, `run.sh gates-build` alone and the §2 comparison before any label.
