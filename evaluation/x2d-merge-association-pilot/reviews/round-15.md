# X2d — council experiment review, round 15 (rev 12 after the round-14 fixes: R14-01 README execution provenance, R14-02 HD6 state, consequential edits (c)–(e))

Date: 2026-10-30. Reviewer: innovation council. Opus 5.5 wrote the lead synthesis. Two peers worked read-only and received the inputs without the lead's leaning:
- `innovation-challenger` = R15Adversary (GPT Astra);
- `experiment-peer` = R15FixVerify (Opus 5.5).

Governance: `CLAUDE.md` experiment rules 1–7; user decision R-052.

Scope:
- the fixes in `reviews/round-14-response.md` §1, checked against `reviews/round-14.md` §2 (raw rows `:56-57`, consequential edits (c)–(e));
- the §4 conditions of round 14;
- the changed lines (`README.md:15`, `PREREGISTRATION.md:3,14,16`, `docs/hypothesis-ledger.md:37`, `DRAFT-NOTES.md:122`);
- any frozen-file statement that rev 12 makes false.

Reviewed files (SHA-256): `DRAFT-NOTES.md` `32c931aa2b66c3485a3d2c9a31f0308bdb5d138cc442e611fd99474fc8134feb`; `GUIDE.md` `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980`; `PREREGISTRATION.md` `f93304879e670d059028e837bab46054b757b08dbf71f429305e24bf1419bdbb`; `README.md` `809bb76aefdfdcda4ea7547889d0249918840865b8f95cc73d8b1a8d6fa5d14a`; `analysis_x2d.py` `b0a66eedb739c2952b14f8468ae56da65d9628f6fa345f8dc51380520b397ade`; `b2.py` `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f`; `baselines.py` `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0`; `build_controls.py` `a71c4e634b5a0d0b42eda15f8e58220fef159c077fd3bf22d36041dd3e4857ab`; `build_units.py` `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243`; `collect_onboard.py` `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff`; `compose.py` `296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1`; `frontend.py` `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8`; `gold.py` `6b0afaad2a3ae6652a02207813c5a3492ba6954bd0a745b58a3f98c960bb321e`; `jev/main.go` `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0`; `label_x2d.py` `fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb`; `measurement/lever-result.json` `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c`; `measurement/lever_check.py` `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42`; `onboard/main.go` `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b`; `power_sim.py` `a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279`; `prepass/commits.jsonl` `1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c`; `prepass/k0.json` `f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab`; `prepass/reserve.json` `a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a`; `prepass/rounds.json` `0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516`; `prepass/scopes.json` `a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512`; `prepass/summary.json` `e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1`; `prepass.py` `1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696`; `quality.py` `6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade`; `questions/merge.area_covers@v1.json` `cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0`; `questions/merge.covered@v1.json` `f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9`; `questions/merge.which_area@v1.json` `cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928`; `questions/onboard.covered@v1.json` `ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70`; `questions/onboard.same_responsibility@v1.json` `31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1`; `questions/onboard.which_file_area@v1.json` `425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57`; `reference_arm.py` `d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739`; `run.sh` `f48cbf01a9e69c67f58c98e3c78880db814f7e28452587d0981176385d937595`; `scopes.py` `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072`; `signoff.json` `e7059f2a9407e38c4d5e8aff0480913fab36a32a69ccbb9721c9633934fdbc2d`; `thresholds.json` `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb`; `tune.py` `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c`; `validation/r4_checks.py` `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52`; `validation/stub_clawpatch_x2d.js` `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f`; `validation/stub_jev.py` `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d`; `validation/stub_labeller.py` `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5`; `validation/validate.py` `a8ce75417cc883b0ad73d36b95c4536744271686c2da746859dfc8e02eeef7c4`; `versioning.py` `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168`; `x2d.py` `62e8cc1de402e221d6a57d04f9ebfc50c9e7bf4875ac5f3d65709830a60665ca`

Rule 3 check: no predictor–outcome association has been computed. The council read the following, all read-only:
- the `../scout-x2d/…/data/rounds/0/` directory listing. It has no `version.rev11.json`, `results.json`, `jev-n1/` or `jev-p/`, and every entry is about an hour old;
- `rounds/0/version.json:160,180` (`question_version_keys: null`, `thresholds_sha256` `5c211eb9…94a8`);
- `rounds/0/labels-check.json` (`ok: true`; a label-only statistic);
- the `rounds/0/thresholds.json` tags (`provisional`, `provisional-0.7`);
- `signoff.json` in both trees.

The council compared no Jev answer with any label or gold.

## 1. Verification

Status: **verified** = as claimed; **partial** = as claimed with a gap; **not verifiable read-only** = left to the conditions.

| item | status | evidence (file:line) | gap → issue |
|---|---|---|---|
| R14-01 (a) | applied as worded; the wording is incomplete | `README.md:15` carries the round-14 replacement word for word ("…run on live data since the rev-9 sign-off: onboarding, catalogue freeze and quality at `730674e`, `gates-build` onward at `0b56594`; `PREREGISTRATION.md` §14 item 18)"). The text "from the rev-11 sign-off" has no match. "`gates-build` onward at `0b56594`" is false for every stage from the R12-01 re-freeze on, which runs at the rev-12 commit. The defect is in the council's own round-14 text | **R15-01** |
| R14-01 (b) | verified | `PREREGISTRATION.md:14` "`:3` and `:15` state what exists, R13-01, R14-01" | — |
| R14-02 (a) | verified | `PREREGISTRATION.md:16` "Hypothesis **HD6** (`docs/hypothesis-ledger.md`; `preregistered` since the rev-9 sign-off, council round 9; it stays `preregistered` at the rev-12 sign-off, §14 item 19)." The text "`proposed`; becomes" has no match. This agrees with `reviews/round-9.md:42-46` (rev 9 signed off with zero issues), ledger `:32` (state `preregistered`), ledger `:39` (proposed → preregistered, round 9) and `PREREGISTRATION.md:242` ("HD6 stays `preregistered`") | — |
| R14-02 (b) | verified | `PREREGISTRATION.md:14` "the ledger-sources line (R-052; the HD6 state, R14-02)," | — |
| (c) | verified | `PREREGISTRATION.md:14` "council rounds 12–14: R12-01–R12-03, R13-01, R13-02, R14-01, R14-02" | — |
| (d) | verified | `PREREGISTRATION.md:3` "DRAFT for council round 15" | — |
| (e) | verified | `docs/hypothesis-ledger.md:37` ends "council rounds 12–14 raised R12-01–R12-03, R13-01–R13-02 and R14-01–R14-02, fixed in revision 12; council round 15 pending." | — |
| Note (i) `DRAFT-NOTES.md:122` (not frozen) | verified | It cites the round-13 and round-14 responses. It reads "`:3` and `:15` state what exists (R13-01, R14-01)" and "`PREREGISTRATION.md:16`: the HD6 state (R14-02)", and it names council rounds 12–14 | — |
| Note (ii) unfiltered diff | verified | Response §2 prints the full `-U0` diff, numstat and hunk list with `signoff.json` in place. There is no `signoff.json` row, and the parenthetical about it is gone | — |
| Note (iii) | respected | `README.md:3` ("The files marked *exists*…") and `README.md:14` are unchanged | — |
| §4.1 only these edits | verified as text and structure | `PREREGISTRATION.md` has 249 lines against the base's 247: inserts at 14 and 242, and every heading sits at base + 1 or + 2. `README.md` has 74 lines in both trees. `validate.py` has 3,832 lines, and `:218` and `:789-794` hold the expected text. Byte identity is not provable read-only | condition (§5) |
| §4.2 diff and numstat | verified as text | Numstat: `PREREGISTRATION.md` 6/4 with hunks {3, 14, 16, 156, 192, 242}; `README.md` 3/3 with {3, 15, 69}; `validate.py` 20/13 with {218, 678, 768, 771,9, 783, 786, 789,6}. The `run.sh` 1/1, `thresholds.json` 3/3, `x2d.py` 2/2, `DRAFT-NOTES.md` 3/0 and ledger 1/0 rows, and their index lines, equal round-13-response §3. Of the base `signoff.json` (round-11 record) `files` digests, exactly the six frozen files in the diff differ: PREREG, README, `run.sh`, `thresholds.json`, `validate.py`, `x2d.py` | — |
| §4.3 validation (a) | verified from the log | `/tmp/x2d-validate-r14a.log:715-716` "651 checks, 0 failed" / "all validation checks passed". No line starts with FAIL. `:9` PASS (caps equal the x2d constants); `:79` PASS with the rev-12 label "x2d.CAPS total only". The response reports the record digest `e7059f2a…dc2d` before and after the move (= `pilot-round-0.md:7`). That this run used the exact rev-12 bytes is not verifiable read-only | condition (§5) |
| §4.4 digest table | verified as text | 44 rows, whose names equal the `signoff.json` `files` keys. The lead and the peer each compared the 42 rows other than `PREREGISTRATION.md` and `README.md` one by one with round-13-response §4: all equal. That includes `validate.py` `a8ce7541…f7c4`, `x2d.py` `62e8cc1d…65ca`, `run.sh` `f48cbf01…7595` and `thresholds.json` `7142d7ac…95eb`. PREREG moves `8b3e6375…f9db` → `f9330487…bdbb`; README moves `c170b870…fa74` → `809bb76a…d14a`. Tune pin `2f850933…2e4c` = `tune_script_sha256`. The values are not hashable read-only | condition (§5) |
| §4.5 no write to the run data | consistent | See the Rule 3 check above: the rev-11 `thresholds_sha256`, `question_version_keys: null`, and no re-freeze or analysis artefact. That nothing was written cannot be proven read-only | condition (§5) |
| R12-01 re-freeze | stands as signed | The `:242` steps are unchanged. `versioning.py:115-138` (`build_version`): `code_digests` is `x2d.go_code_digests()` and `state_builder` covers only the `STATE_BUILDER` modules. Neither changes in rev 12, so the only keys that can move are `thresholds_sha256` and `question_version_keys`. `freeze_version` reuses an existing `settings.json` and `inputs/` (`:85-87`, `:140-151`) | — |
| Other frozen statements made false by rev 12 | 1 found | In the changed lines: `README.md:15` "`gates-build` onward at `0b56594`". In base text: `PREREGISTRATION.md:241` (item 18) "everything from `gates-build` runs at the rev-11 commit". Item 19 (`:242`), the status line and §15 name no third commit. The remaining stale-looking text is historical or below the bar (notes) | **R15-01** |

## 2. Issues

| id | severity | finding | exact fix |
|---|---|---|---|
| R15-01 | minor (rule 6, reported deviations and reproduce commands; the run's repeatability. Same class as R14-01, and again introduced by the council's own prescribed wording) | The run now spans **three** commits, but the frozen files name two. `README.md:15` says "`gates-build` onward at `0b56594`", and item 18 (`:241`) says "everything from `gates-build` runs at the rev-11 commit". Under rev 12 both statements hold only up to the cost stop. The R12-01 re-freeze (`:242` step (2): "under the signed-off rev-12 bytes"), the rest of `gates-jev` (`main` remainder, `n1`, `p1`, `p2`), `gates-analyse`, `power`, rounds 1–10 and `verdict` all run at the rev-12 sign-off commit. They cannot run at `0b56594`: its `x2d.CAPS` is $1 / $10 (response §2, `x2d.py` hunk `-45`), and round 0 has already spent $1.0001, so the driver refuses there (`/tmp/x2d-validate-r14a.log:77`, exit 2 once the spend exceeds the cap). No frozen file names the third commit. Round 14's own §5 step 9 plans RESULTS `Deviations` for "the two commits". RESULTS would therefore omit a run commit, and anyone reproducing from the frozen README would check out a commit that cannot run the later stages | (a) `README.md:15`: "`gates-build` onward at `0b56594`; `PREREGISTRATION.md` §14 item 18)" → "`gates-build`, `gates-label` and the `gates-jev` run up to the cost stop at `0b56594`, the round-0 re-freeze and every later stage at the rev-12 sign-off commit; `PREREGISTRATION.md` §14 items 18–19)". (b) `PREREGISTRATION.md:242`: "`RESULTS.md` `Deviations` states the re-freeze and why." → "`RESULTS.md` `Deviations` states the re-freeze and why. **Third run commit (council round 15, R15-01).** The re-freeze and every stage after it (the rest of `gates-jev`, `gates-analyse`, `power`, rounds 1–10 and `verdict`) run at the rev-12 sign-off commit, to which the run worktree moves by checkout only (no `git clean`, no `stash -u`; `data/` untracked); `0b56594` cannot run them, because its `x2d.CAPS` is $1 / $10 and round 0 has spent $1.0001. Item 18's \"everything from `gates-build` runs at the rev-11 commit\" therefore holds up to the cost stop only, and `RESULTS.md` `Deviations` names the three run commits (`730674e`, `0b56594`, the rev-12 commit) and the stages run at each." Both lines are already rev-12 hunks, so no hunk is added |

Consequential edits (exact; one match each):
- (c) `PREREGISTRATION.md:14`, three edits:
  - "council rounds 12–14: R12-01–R12-03, R13-01, R13-02, R14-01, R14-02)" → "council rounds 12–15: R12-01–R12-03, R13-01, R13-02, R14-01, R14-02, R15-01)";
  - "`:3` and `:15` state what exists, R13-01, R14-01)" → "`:3` and `:15` state what exists, R13-01, R14-01, R15-01)";
  - "§14 item 19 (the stop and the decision, verbatim)" → "§14 item 19 (the stop and the decision, verbatim; the third run commit, R15-01)".
- (d) `PREREGISTRATION.md:3`: "DRAFT for council round 15" → "DRAFT for council round 16".
- (e) `docs/hypothesis-ledger.md:37` (main agent; not frozen): "council rounds 12–14 raised R12-01–R12-03, R13-01–R13-02 and R14-01–R14-02, fixed in revision 12; council round 15 pending." → "council rounds 12–15 raised R12-01–R12-03, R13-01–R13-02, R14-01–R14-02 and R15-01, fixed in revision 12; council round 16 pending."

Notes (below the bar; no fix required for sign-off):
- (i) `DRAFT-NOTES.md:122` (not frozen):
  - add `reviews/round-15-response.md` to the responses;
  - "`:3` and `:15` state what exists (R13-01, R14-01)" → "…(R13-01, R14-01, R15-01)";
  - add "`PREREGISTRATION.md:242`: the third run commit (R15-01)";
  - "council rounds 12–14" → "council rounds 12–15", naming `reviews/round-15.md`.
- (ii) `README.md:15` lists "onboarding, catalogue freeze and quality". Item 18 and `pilot-round-0.md:14-15` also name consensus and the intruder audit. The item-18 pointer covers them, and no figure moves.
- (iii) `PREREGISTRATION.md:244` "at the run commit HD6 → `preregistered`" could be read as a second transition at the rev-12 commit. `:16` and `:242` ("stays") settle it.
- (iv) `PREREGISTRATION.md:3` "…sign-off … before the round-0 run". This is unchanged governance wording from rev 1, true of the original sign-off, and the same line states what has run.
- (v) The ledger HD6 row (`docs/hypothesis-ledger.md:32`, not frozen) still cites "PREREGISTRATION rev 9; council sign-off round 9". Update its evidence column at resume step 10. Its state does not change.
- (vi) Round-13 note (iv) and round-14 note (iv) still hold. Run the re-freeze step (2) from the run-worktree root as `.tools/venv/bin/python evaluation/x2d-merge-association-pilot/versioning.py --round 0`, with `X2D_DATA` unset or equal to `run.sh`'s `D`, and record the command.

## 3. Synthesis

- **R14-01, R14-02, (c)–(e), note (i):** all three reviewers found every replacement in place word for word, one match each, and the old text gone.
- **R15-01:**
  - The lead found it while checking the `README.md:15` wording against the post-sign-off resume. The peer raised it independently, with the same evidence (`x2d.py` base CAPS, `:242` step (2), round-14 §5 step 9 "two commits").
  - The challenger reported zero issues. Its check of `README.md:15` against item 18 and `pilot-round-0.md` is correct for the stages already run, but it did not test the stages still to run. The lead had no channel to put the point back to the challenger in this session. The lead and the peer hold it above the bar because it is a false statement in a frozen file about the run's provenance, and RESULTS `Deviations` and the reproduce path depend on it.
  - The council owns the error: its round-14 fix text described only the commits already used. The R15-01 fix states the whole commit timeline (`730674e` → `0b56594` → the rev-12 commit) so that no further wording round is needed. The peer's fix wording was adopted, with item 19's sentence extended to name the stages and the checkout procedure.
- **R12-01:** all three reviewers keep the preregistered re-freeze. The lead re-traced `build_version`: only `thresholds_sha256` and `question_version_keys` can differ.
- **Agreed by all three:**
  - the 6/4, 3/3 and 20/13 numstat and the hunk sets;
  - the 42 of 44 digests equal to round-13-response §4, and the tune pin;
  - the 651/0 log;
  - no write to the run data seen.

## 4. Required for round 16

1. Apply R15-01 (a), (b) and the consequential edits (c)–(e) only, by exact-substring replacement with one match each. No other frozen byte moves. `validation/validate.py` stays at `a8ce75417cc883b0ad73d36b95c4536744271686c2da746859dfc8e02eeef7c4`. `x2d.py`, `run.sh` and `thresholds.json` keep their round-14-response §4 digests.
2. Recompute `git diff -U0 0b5659440c7368ac18b82f60a4edf3956408d2bf -- evaluation/x2d-merge-association-pilot docs/research-ledger.md docs/hypothesis-ledger.md` and `--numstat` with `signoff.json` in place, and print both unfiltered. Expected: `PREREGISTRATION.md` 6/4 with hunks {3, 14, 16, 156, 192, 242}; `README.md` 3/3 with {3, 15, 69}; `validate.py` 20/13 with {218, 678, 768, 771–779, 783, 786, 789–794}. Every other row must be as in round-14-response §2, with no `signoff.json` row. Any other result reopens review.
3. Run validation (a) with `signoff.json` moved aside: 651 checks, 0 failed. Record the log path and the record's SHA-256 before and after the move.
4. Recompute the 44 frozen-file digests after (3). Expected: 42 rows equal round-14-response §4, and new digests for `PREREGISTRATION.md` and `README.md` only. Tune pin `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c`.
5. No write to `../scout-x2d/…/data` before sign-off.

## 5. Resume order after a future zero-issue sign-off (for reference; not authorised by this round)

Sign-off conditions, all before step 1:
- the 44 digests recomputed and equal to the signed table, with the tune pin;
- a new `signoff.json` (round `reviews/round-16.md`, or the round that signs off) whose `code_under_test` equals round 11's (17 rows);
- validation (a) 0 failed;
- validation (b) with the new record in place, showing exactly the one expected failure (`C without the council's sign-off record…`);
- one commit;
- the run worktree moved by checkout only.

Order:
1. The R12-01 round-0 re-freeze, steps (1)–(5) of `PREREGISTRATION.md:242`, with the command (note vi). Record it, together with the R12-02 (d) attestation, in `reviews/pilot-round-0.md`. The attestation is appended with the `gates-label` record and the `gates-jev` stop record (exit 4; 7,755 of the `main` run's 7,872; $1.0001; seven `network` / `cost_known: false` failures; `n1` / `p1` / `p2` not started; R-052). If the main agent cannot attest, or step (3) shows any other difference, stop for a new council round.
2. Resume `gates-jev`.
3. `gates-analyse`.
4. `power`.
5. The round-0 critique (its `version_sha256` is the re-frozen file's digest).
6. Rounds 1–7, each with its critique.
7. Frozen rounds 8–10.
8. `verdict`.
9. `RESULTS.md`. Its `Deviations` cover the **three run commits** (`730674e`, `0b56594`, the rev-12 commit) with the stages of each, the cost stop and R-052, and the re-freeze.
10. Ledger states (HD6 per the §14 map; it stays `preregistered` at the rev-12 sign-off; note v).

## Verdict

**1 ISSUE** (R15-01 minor). Rev 12 is not signed off. `signoff.json` stays at the round-11 record and no stage may run. Re-review in round 16 after the fix in §2 and the items in §4.
