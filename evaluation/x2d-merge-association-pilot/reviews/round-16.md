# X2d — council experiment review, round 16 (rev 12 after the round-15 fixes: R15-01 third run commit, consequential edits (c)–(e))

Date: 2026-10-30. Reviewer: innovation council. Opus 5.5 wrote the lead synthesis. Two peers worked read-only and received the inputs without the lead's leaning:
- `innovation-challenger` = R16Adversary (GPT Astra);
- `experiment-peer` = R16FixVerify (Opus 5.5).

Governance: `CLAUDE.md` experiment rules 1–7; user decision R-052.

Scope:
- the fixes in `reviews/round-15-response.md` §1, checked against `reviews/round-15.md` §2 (raw row `:54`, consequential edits (c)–(e), note (i));
- the §4 items of round 15;
- the changed lines (`README.md:15`, `PREREGISTRATION.md:3,14,242`, `docs/hypothesis-ledger.md:37`, `DRAFT-NOTES.md:122`);
- any frozen-file statement that rev 12 makes false.

Reviewed files (SHA-256): `DRAFT-NOTES.md` `0dcb1972ecdc31eef3267aefb62d747493c4c968b93dc1e3ae6eb4c54fa6799e`; `GUIDE.md` `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980`; `PREREGISTRATION.md` `de553c5265407f0a8aae566b9b0f59c11b670a5d8d99386da39de45d945c8120`; `README.md` `bd1d47ab62513ae97a218c0aeedc6211c644d854401050a6e6df435c29f858df`; `analysis_x2d.py` `b0a66eedb739c2952b14f8468ae56da65d9628f6fa345f8dc51380520b397ade`; `b2.py` `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f`; `baselines.py` `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0`; `build_controls.py` `a71c4e634b5a0d0b42eda15f8e58220fef159c077fd3bf22d36041dd3e4857ab`; `build_units.py` `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243`; `collect_onboard.py` `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff`; `compose.py` `296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1`; `frontend.py` `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8`; `gold.py` `6b0afaad2a3ae6652a02207813c5a3492ba6954bd0a745b58a3f98c960bb321e`; `jev/main.go` `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0`; `label_x2d.py` `fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb`; `measurement/lever-result.json` `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c`; `measurement/lever_check.py` `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42`; `onboard/main.go` `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b`; `power_sim.py` `a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279`; `prepass/commits.jsonl` `1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c`; `prepass/k0.json` `f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab`; `prepass/reserve.json` `a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a`; `prepass/rounds.json` `0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516`; `prepass/scopes.json` `a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512`; `prepass/summary.json` `e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1`; `prepass.py` `1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696`; `quality.py` `6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade`; `questions/merge.area_covers@v1.json` `cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0`; `questions/merge.covered@v1.json` `f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9`; `questions/merge.which_area@v1.json` `cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928`; `questions/onboard.covered@v1.json` `ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70`; `questions/onboard.same_responsibility@v1.json` `31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1`; `questions/onboard.which_file_area@v1.json` `425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57`; `reference_arm.py` `d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739`; `run.sh` `f48cbf01a9e69c67f58c98e3c78880db814f7e28452587d0981176385d937595`; `scopes.py` `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072`; `signoff.json` `e7059f2a9407e38c4d5e8aff0480913fab36a32a69ccbb9721c9633934fdbc2d`; `thresholds.json` `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb`; `tune.py` `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c`; `validation/r4_checks.py` `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52`; `validation/stub_clawpatch_x2d.js` `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f`; `validation/stub_jev.py` `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d`; `validation/stub_labeller.py` `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5`; `validation/validate.py` `a8ce75417cc883b0ad73d36b95c4536744271686c2da746859dfc8e02eeef7c4`; `versioning.py` `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168`; `x2d.py` `62e8cc1de402e221d6a57d04f9ebfc50c9e7bf4875ac5f3d65709830a60665ca`

Rule 3 check: no predictor–outcome association has been computed. The council read the following, all read-only:
- the `../scout-x2d/…/data/rounds/0/` directory listing. It has no `version.rev11.json`, `results.json`, `jev-n1/` or `jev-p/`. Its only Jev artefact is `jev/scores.jsonl`;
- `signoff.json` (`round` `reviews/round-11.md`; 44 `files` keys; 17 `code_under_test` rows; `tune_script_sha256` `2f850933…2e4c`);
- `/tmp/x2d-validate-r15a.log`.

The council compared no Jev answer with any label or gold.

## 1. Verification

Status: **verified** = as claimed; **not verifiable read-only** = left to the conditions.

| item | status | evidence (file:line) | gap → issue |
|---|---|---|---|
| R15-01 (a) | verified | `README.md:15` matches the round-15 text word for word: "…quality at `730674e`, `gates-build`, `gates-label` and the `gates-jev` run up to the cost stop at `0b56594`, the round-0 re-freeze and every later stage at the rev-12 sign-off commit; `PREREGISTRATION.md` §14 items 18–19)". "`gates-build` onward at `0b56594`" has no match in the study tree outside `reviews/` | — |
| R15-01 (b) | verified | `PREREGISTRATION.md:242`: after "…states the re-freeze and why." comes the "**Third run commit (council round 15, R15-01).**" paragraph, word for word. It names the stages (the rest of `gates-jev`, `gates-analyse`, `power`, rounds 1–10, `verdict`), "checkout only (no `git clean`, no `stash -u`; `data/` untracked)", the $1 / $10 reason, the item-18 limit, and "`RESULTS.md` `Deviations` names the three run commits (`730674e`, `0b56594`, the rev-12 commit)". Item 19 still ends "HD6 stays `preregistered`." The escaped `\"` of the review text was applied as plain quotes, which is correct | — |
| (c) | verified | `PREREGISTRATION.md:14` contains all three: "council rounds 12–15: R12-01–R12-03, R13-01, R13-02, R14-01, R14-02, R15-01)", "`:3` and `:15` state what exists, R13-01, R14-01, R15-01)" and "§14 item 19 (the stop and the decision, verbatim; the third run commit, R15-01)". None of the three old substrings matches | — |
| (d) | verified | `PREREGISTRATION.md:3` "DRAFT for council round 16" | — |
| (e) | verified | `docs/hypothesis-ledger.md:37` ends "council rounds 12–15 raised R12-01–R12-03, R13-01–R13-02, R14-01–R14-02 and R15-01, fixed in revision 12; council round 16 pending." | — |
| Note (i) `DRAFT-NOTES.md:122` (not frozen) | verified | Lists `reviews/round-15-response.md`; "(R13-01, R14-01, R15-01)"; "`PREREGISTRATION.md:242`: the third run commit (R15-01)"; council rounds 12–15, naming `reviews/round-15.md` | — |
| §4.1 only these edits | verified as text and structure | `PREREGISTRATION.md` has 249 lines and `README.md` 74, as in round 15; (b) appends to the existing line 242. Byte identity cannot be proven read-only | condition (§5) |
| §4.2 diff and numstat | verified | Response §2 prints the `-U0` diff unfiltered, with `signoff.json` in place. Numstat: `PREREGISTRATION.md` 6/4 with hunks {3, 14, 16, 156, 192, 242}; `README.md` 3/3 with {3, 15, 69}; `validate.py` 20/13 with {218, 678, 768, 771–779, 783, 786, 789–794} (1/1+1/1+1/1+9/4+1/1+1/1+6/4). The `run.sh` 1/1, `thresholds.json` 3/3, `x2d.py` 2/2, `DRAFT-NOTES.md` 3/0 and both ledger 1/0 rows equal round-14-response §2, and so do the index blobs of `run.sh`, `thresholds.json`, `validate.py`, `x2d.py` and `docs/research-ledger.md`. There is no `signoff.json` row | — |
| §4.3 validation (a) | verified from the log | `/tmp/x2d-validate-r15a.log` ends "651 checks, 0 failed" / "all validation checks passed", and no line starts with FAIL. The log has PASS for "thresholds.json minimums, PR cap, round table, caps and seed equal the x2d constants" and the R3-17 check with the rev-12 label "charged to the x2d.CAPS total only". The response gives the record digest `e7059f2a…dc2d` before and after the move (equal). That the run used exactly the rev-12 bytes is not verifiable read-only | condition (§5) |
| §4.4 digest table | verified as text | 44 rows (`round-15-response.md:198-241`). The lead and both peers each compared them row by row with `round-14-response.md:200-243`: the 42 rows other than `PREREGISTRATION.md` and `README.md` are equal. That includes `validate.py` `a8ce7541…f7c4`, `x2d.py` `62e8cc1d…65ca`, `run.sh` `f48cbf01…7595` and `thresholds.json` `7142d7ac…95eb`. PREREG moves `f9330487…bdbb` → `de553c52…8120`; README moves `809bb76a…d14a` → `bd1d47ab…58df`. The 44 names equal the `signoff.json` `files` keys. The tune pin `2f850933…2e4c` equals `tune_script_sha256`. The values cannot be hashed read-only | condition (§5) |
| §4.5 no write to the run data | consistent | See the Rule 3 check: no re-freeze or analysis artefact. That nothing was written cannot be proven read-only | condition (§5) |
| New paragraph against the rest of rev 12 | consistent | Item 18 (`:241`, "everything from `gates-build` runs at the rev-11 commit") is now limited explicitly by item 19 ("holds up to the cost stop only"). R12-01's "Once, after the rev-12 sign-off and before `gates-jev`" agrees with "the re-freeze and every stage after it". The stage order agrees with `run.sh` (`gates-jev` → `gates-analyse` → `power`, `:315-335`; `round`, `rescore-v7`, `verdict` `:353-363`). "`0b56594` cannot run them": `x2d.jev_budget` under $1 / $10 leaves round 0 nothing after $1.0001, and every later stage depends on `gates-jev`. `verify_pins` checks file sets, file digests, code digests and the tune pin, not HEAD or the record's `round` (`x2d.py:182-217`), so nothing ties the run to a particular commit hash | — |
| Other frozen statements made false by rev 12 | none found | All three reviewers searched the frozen files for `$1`, `$10`, `10.0`, `8.2`, "two commits", `0b56594`, `730674e`, "rev-11 commit" and "round 15/16". Every remaining `$1` / `$10` is labelled historical (`PREREGISTRATION.md:14,16,156,192,242`; `README.md:69`; `validate.py:768`). The values in force agree with the caps: `x2d.py:45`, `thresholds.json:181-183`, `validate.py:218`, `run.sh:248`, `x2d.py:521`. "Two commits" at `:241` is limited by `:242`. `reviews/pilot-round-0.md:17` ("Everything from `gates-build` onward runs at `0b56594`") is a run record, not frozen; see note (ii) | — |

## 2. Issues

None.

Notes (below the bar; no frozen file changes):
- (i) **Critique transport (R16FixVerify).** `critique_ok` reads `reviews/pilot-round-<n>.md` from the run worktree's own `reviews/` (`x2d.py:106-118`; `X2D_REVIEWS_DIR` is honoured only in validation). `reviews/` is outside the frozen set (`x2d.py:129`), so `verify_pins` passes either way. But if the critiques were committed and the worktree checked out again, HEAD would move past the rev-12 commit, and item 19 and `README.md:15` say "every later stage at the rev-12 sign-off commit". The procedure in §5 (step 1, item c) keeps both statements true without any frozen change.
- (ii) **Provenance record (R16FixVerify).** `reviews/pilot-round-0.md:14-17` records two commits and is append-only. Do not edit those lines. The rev-12 commit is recorded by appending at resume step 1 (§5).
- (iii) `power` is model-free and needs only `rounds/0/k23.json`, which already exists, so strictly speaking `0b56594` could run it. Item 19 still has it run at rev 12, after `gates-analyse`, and no figure or Deviations entry changes.
- (iv) Resume with the named sub-stages (`gates-jev`, `gates-analyse`, `power`), never the aggregate `gates`, which starts with `gates-build` (`run.sh:351`) (R16Adversary).
- (v) Round-15 notes (ii)–(vi) still hold and are not re-raised. Note (v) (the ledger HD6 evidence column) is done at resume step 10. Note (vi) (the re-freeze command) is done at resume step 1.

## 3. Synthesis

- **Fixes:** all three reviewers found every R15-01 replacement in place word for word, one match each, with the old text gone and no extra hunk.
- **Digests and diff:** all three reviewers compared the 44 rows independently, and all agree: 42 rows equal, and only PREREG and README differ. They also agree on the numstat and hunk sets and on the 651/0 log.
- **Both peers report zero issues.** The lead's own pass also found none.
- **Where the peers changed the recommendation:**
  - The experiment-peer's critique-transport point (note i) and provenance point (note ii) did not reach the bar: a compliant procedure exists under the frozen text, and `verify_pins` binds the frozen bytes, not the commit. They are still adopted as resume-procedure requirements (§5, step 1), so that RESULTS' "three run commits" stays true.
  - The peer's `code_under_test` pre-check is already covered by sign-off conditions 2 and 4: the new record's `code_under_test` must equal round 11's 17 rows, and validation (b) with the record in place exercises `verify_pins`, including the code check. A new file under `internal/decisions` would add a key and fail condition 2.
  - The challenger's sub-stage point became note (iv).
- **R12-01:** kept as signed. The peer re-traced `versioning.py:115-138`: only `thresholds_sha256` and `question_version_keys` can differ. The `jev/scores.jsonl` header carries a non-null `versions`, so step (3) can be met.

## 4. Verdict

**ZERO ISSUES — signed off.**

### Sign-off statement

The innovation council signs off X2d **PREREGISTRATION revision 12** and its implementation with zero open issues (`CLAUDE.md` experiment rules 1–5; user decision R-052). The sign-off covers exactly these 44 frozen files (SHA-256, paths relative to `evaluation/x2d-merge-association-pilot/`) and the external tune pin.

| file | sha256 |
|---|---|
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `de553c5265407f0a8aae566b9b0f59c11b670a5d8d99386da39de45d945c8120` |
| `README.md` | `bd1d47ab62513ae97a218c0aeedc6211c644d854401050a6e6df435c29f858df` |
| `analysis_x2d.py` | `b0a66eedb739c2952b14f8468ae56da65d9628f6fa345f8dc51380520b397ade` |
| `b2.py` | `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f` |
| `baselines.py` | `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0` |
| `build_controls.py` | `a71c4e634b5a0d0b42eda15f8e58220fef159c077fd3bf22d36041dd3e4857ab` |
| `build_units.py` | `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243` |
| `collect_onboard.py` | `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff` |
| `compose.py` | `296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1` |
| `frontend.py` | `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8` |
| `gold.py` | `6b0afaad2a3ae6652a02207813c5a3492ba6954bd0a745b58a3f98c960bb321e` |
| `jev/main.go` | `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0` |
| `label_x2d.py` | `fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb` |
| `measurement/lever-result.json` | `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c` |
| `measurement/lever_check.py` | `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42` |
| `onboard/main.go` | `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b` |
| `power_sim.py` | `a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279` |
| `prepass/commits.jsonl` | `1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c` |
| `prepass/k0.json` | `f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab` |
| `prepass/reserve.json` | `a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a` |
| `prepass/rounds.json` | `0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516` |
| `prepass/scopes.json` | `a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512` |
| `prepass/summary.json` | `e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1` |
| `prepass.py` | `1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696` |
| `quality.py` | `6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade` |
| `questions/merge.area_covers@v1.json` | `cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0` |
| `questions/merge.covered@v1.json` | `f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9` |
| `questions/merge.which_area@v1.json` | `cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928` |
| `questions/onboard.covered@v1.json` | `ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70` |
| `questions/onboard.same_responsibility@v1.json` | `31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1` |
| `questions/onboard.which_file_area@v1.json` | `425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57` |
| `reference_arm.py` | `d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739` |
| `run.sh` | `f48cbf01a9e69c67f58c98e3c78880db814f7e28452587d0981176385d937595` |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` |
| `thresholds.json` | `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb` |
| `tune.py` | `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c` |
| `validation/r4_checks.py` | `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `a8ce75417cc883b0ad73d36b95c4536744271686c2da746859dfc8e02eeef7c4` |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` |
| `x2d.py` | `62e8cc1de402e221d6a57d04f9ebfc50c9e7bf4875ac5f3d65709830a60665ca` |

Tune pin (`tune_script_sha256`): `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c`.

**The sign-off holds only if all of the following are true before resume step 1.** If any one fails, the sign-off lapses and a new council round is needed.
1. The 44 digests above are recomputed from the committed bytes and each one is equal. The tune pin is equal.
2. A new `signoff.json` has `round` `reviews/round-16.md`. Its `files` equal the table above, its `tune_script_sha256` equals the pin, and its `code_under_test` equals round 11's 17 rows exactly (same keys, same digests).
3. Validation (a), with `signoff.json` moved aside, gives 651 checks and 0 failed. Record the log path, and the record's SHA-256 before and after the move.
4. Validation (b), with the new record in place, shows exactly 1 failure, the expected one: `C without the council's sign-off record the collector refuses before any clone or model call`. Record the log path.
5. Everything goes in one commit: the rev-12 files, `signoff.json`, this round and the response. Record its full hash: it is the **rev-12 sign-off commit**, the third run commit.
6. The run worktree `../scout-x2d` moves to that commit by checkout only: no `git clean`, no `stash -u`; `data/` untracked and untouched. Record `git rev-parse HEAD` of the worktree.

### Resume order (authorised once the conditions hold)

1. **R12-01 re-freeze and records.**
   - (a) Run steps (1)–(5) of `PREREGISTRATION.md:242`. Step (2) is run from the run-worktree root as `.tools/venv/bin/python evaluation/x2d-merge-association-pilot/versioning.py --round 0`, with `X2D_DATA` unset or equal to `run.sh`'s `D`; record the command.
   - (b) Append to `reviews/pilot-round-0.md`. Do not edit `:14-17`. The append contains:
     - the re-freeze record (both files' SHA-256, the key diff, the drift output);
     - the R12-02 (d) attestation, with the `gates-label` record and the `gates-jev` stop record (exit 4; 7,755 of the `main` run's 7,872; $1.0001; seven `network` / `cost_known: false` failures; `n1` / `p1` / `p2` not started; R-052);
     - a "Third run commit (item 19, R15-01)" entry giving the full rev-12 commit hash and the worktree HEAD.
   - (c) From here to `verdict` the worktree is not checked out again. Each critique `reviews/pilot-round-<n>.md` is copied into the run worktree's `reviews/` before the stage that reads it, and the SHA-256 of the copy is recorded in the next record (note i).
   - If the main agent cannot attest, or step (3) of `:242` shows any other difference, stop for a new council round.
2. Resume `run.sh gates-jev` (the named sub-stage, note iv).
3. `run.sh gates-analyse`.
4. `run.sh power`.
5. The round-0 critique (`reviews/pilot-round-0.md`; its `version_sha256` is the re-frozen file's digest).
6. Rounds 1–7, each with its critique (and `rescore-v7` after the round-7 critique).
7. Frozen rounds 8–10, each with its critique.
8. `run.sh verdict`.
9. `RESULTS.md`. Its `Deviations` name the three run commits (`730674e`, `0b56594`, the rev-12 commit) and the stages run at each, the cost stop and R-052, and the round-0 re-freeze.
10. Ledger states: HD6 follows the §14 map. It stays `preregistered` at the rev-12 sign-off, becomes `under-test` when round 1 starts, and takes its outcome state after `verdict`. Update the evidence column of the HD6 row (round-15 note v).
