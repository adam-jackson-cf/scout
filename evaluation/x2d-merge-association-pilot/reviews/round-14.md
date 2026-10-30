# X2d — council experiment review, round 14 (rev 12 after the round-13 fixes: R13-01 README statements of what exists, R13-02 change record, status line and ledger)

Date: 2026-10-30. Reviewer: innovation council. Opus 5.5 wrote the lead synthesis. Two peers worked read-only and received the inputs without the lead's leaning:
- `innovation-challenger` = R14Adversary (GPT Astra);
- `experiment-peer` = R14FixVerify (Opus 5.5).

Governance: `CLAUDE.md` experiment rules 1–7; user decision R-052.

Scope:
- the fixes in `reviews/round-13-response.md` §1, checked against `reviews/round-13.md` §2;
- the §4 conditions of round 13;
- the changed lines (`README.md:3,15`, `PREREGISTRATION.md:3,14`, `docs/hypothesis-ledger.md:37`, `DRAFT-NOTES.md:122`);
- any frozen-file statement that rev 12 makes false;
- whether the R12-01 re-freeze (`PREREGISTRATION.md:242`) still stands as signed in substance.

Reviewed files (SHA-256): `DRAFT-NOTES.md` `52785d99aa12cbf126e7d07513fe89ca741729abb38fb3e296d44877f3a0ac6a`; `GUIDE.md` `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980`; `PREREGISTRATION.md` `8b3e6375c44e48528c56074194a7a85c4d753c44898ecc3eaa88ceb11098f9db`; `README.md` `c170b870988f21845129ae86e60337cefe5847fea626591a534bb69d1157fa74`; `analysis_x2d.py` `b0a66eedb739c2952b14f8468ae56da65d9628f6fa345f8dc51380520b397ade`; `b2.py` `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f`; `baselines.py` `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0`; `build_controls.py` `a71c4e634b5a0d0b42eda15f8e58220fef159c077fd3bf22d36041dd3e4857ab`; `build_units.py` `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243`; `collect_onboard.py` `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff`; `compose.py` `296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1`; `frontend.py` `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8`; `gold.py` `6b0afaad2a3ae6652a02207813c5a3492ba6954bd0a745b58a3f98c960bb321e`; `jev/main.go` `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0`; `label_x2d.py` `fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb`; `measurement/lever-result.json` `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c`; `measurement/lever_check.py` `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42`; `onboard/main.go` `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b`; `power_sim.py` `a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279`; `prepass/commits.jsonl` `1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c`; `prepass/k0.json` `f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab`; `prepass/reserve.json` `a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a`; `prepass/rounds.json` `0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516`; `prepass/scopes.json` `a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512`; `prepass/summary.json` `e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1`; `prepass.py` `1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696`; `quality.py` `6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade`; `questions/merge.area_covers@v1.json` `cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0`; `questions/merge.covered@v1.json` `f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9`; `questions/merge.which_area@v1.json` `cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928`; `questions/onboard.covered@v1.json` `ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70`; `questions/onboard.same_responsibility@v1.json` `31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1`; `questions/onboard.which_file_area@v1.json` `425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57`; `reference_arm.py` `d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739`; `run.sh` `f48cbf01a9e69c67f58c98e3c78880db814f7e28452587d0981176385d937595`; `scopes.py` `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072`; `signoff.json` `e7059f2a9407e38c4d5e8aff0480913fab36a32a69ccbb9721c9633934fdbc2d`; `thresholds.json` `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb`; `tune.py` `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c`; `validation/r4_checks.py` `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52`; `validation/stub_clawpatch_x2d.js` `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f`; `validation/stub_jev.py` `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d`; `validation/stub_labeller.py` `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5`; `validation/validate.py` `a8ce75417cc883b0ad73d36b95c4536744271686c2da746859dfc8e02eeef7c4`; `versioning.py` `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168`; `x2d.py` `62e8cc1de402e221d6a57d04f9ebfc50c9e7bf4875ac5f3d65709830a60665ca`

Rule 3 check: no predictor–outcome association has been computed. The council read the following, all read-only:
- `rounds/0/labels-check.json` (ok, κ_answer 1.0, 50 units; a label-only statistic);
- the `rounds/0/thresholds.json` tags (`provisional-0.7`);
- `rounds/0/version.json` `:160,180` (`question_version_keys: null`, `thresholds_sha256` `5c211eb9…94a8`);
- the header of `rounds/0/jev/scores.jsonl` (64 `main` units, seven question-version entries);
- the round-0 directory listing (no `version.rev11.json`, `results.json`, `arms/V-*`, `jev-n1/` or `jev-p/`);
- `signoff.json` in both trees.

The council compared no Jev answer with any label or gold.

## 1. Verification

Status: **verified** = as claimed; **partial** = as claimed with a gap; **not verifiable read-only** = left to the conditions.

| item | status | evidence (file:line) | gap → issue |
|---|---|---|---|
| R13-01 (a) | verified | `README.md:3`: "(revision 12, after the round-0 cost stop, user decision R-052)". The rev-9 clause is gone | — |
| R13-01 (b) | verified | `README.md:3` carries the round-13 (b) sentence word for word, including "(`PREREGISTRATION.md` status line, `reviews/pilot-round-0.md`)"; only the response's table shortens it to "(…)". The live round-0 state agrees: labels and gold exist, labels-check is ok, thresholds are provisional, and there is no `results.json` and no `V-*` arm | — |
| R13-01 (c) | applied as worded; the wording is false | `README.md:15` "run on live data from the rev-11 sign-off". The modules ran on live data from the **rev-9** sign-off: onboarding (16 maps), consensus and catalogue freeze, and quality with the intruder audit ran at `730674e` (`PREREGISTRATION.md:241` "Two commits"; `reviews/pilot-round-0.md:14-17`). Only `gates-build` onward ran at `0b56594`. The defect is in the council's own round-13 text | **R14-01** |
| R13-02 (a), (b) | verified | `PREREGISTRATION.md:14`: "council rounds 12–13: R12-01–R12-03, R13-01, R13-02"; "`README.md` (the rev-4 note's total; `:3` and `:15` state what exists, R13-01), `x2d.py` (one docstring figure), `run.sh` (one comment, `:248`)". The `run.sh` hunk is `@@ -248 +248 @@` | — |
| R13-02 (c) | verified | `PREREGISTRATION.md:3` "DRAFT for council round 14" | — |
| R13-02 (d) | verified | `docs/hypothesis-ledger.md:37` ends "council rounds 12–13 raised R12-01–R12-03 and R13-01–R13-02, fixed in revision 12; council round 14 pending." | — |
| Note (iii) `DRAFT-NOTES.md:122` (not frozen) | partial (below the bar) | It cites `reviews/round-12-response.md` and rounds 12–13, and says "not touched before the round-0 re-freeze (R12-01)". It does not cite `reviews/round-13-response.md`, and its README entry omits `:3`/`:15` | note (i) |
| §4.1 `validate.py` unchanged | verified as text | The response states `a8ce7541…f7c4`, equal to the round-13 expectation. The `validate.py` hunks are identical to round 13's. The hash itself is not verifiable read-only | condition (§5) |
| §4.2 `validate.py` `-U0` | verified | Compared line by line with the `0b56594` copy in `../scout-x2d` (3,825 → 3,832 lines). Hunks: 218 (1/1), 678 (1/1), 768 (1/1), base 771–774 → 771–779 (4/9), 778 → 783 (1/1), 781 → 786 (1/1), 784–787 → 789–794 (4/6). That is **20 added / 13 deleted**, new-file hunks exactly {218, 678, 768, 771–779, 783, 786, 789–794}. The response's `+771,9` / `+789,6` agree | — |
| §4.2 other numstat rows | consistent | `PREREGISTRATION.md` 6/4: hunks 3, 14 (+), 16, 156, 192, 242 (+); 247 → 249 lines. `README.md` 3/3: hunks 3, 15, 69; 74 = 74 lines. The other rows check against base: `thresholds.json` 3/3 (`:181-183`), `x2d.py` 2/2, `run.sh` 1/1, `DRAFT-NOTES.md` 3/0, ledgers 1/0 each | — |
| §4.2 `signoff.json` parenthetical | presentation gap (below the bar) | Response §3 says `signoff.json` "shows deleted", but no such row or deletion appears in the block. The live and base `signoff.json` have identical content (round-11 record; 44 + 17 digests; same tune pin) | note (ii) |
| §4.3 validation (a) | verified from the log | `/tmp/x2d-validate-r13a.log:715-716` "651 checks, 0 failed" / "all validation checks passed". `:9` PASS (thresholds pin); `:78-79` PASS (budget). The `:79` label carries the rev-12 wording ("x2d.CAPS total only"), so the run used rev-12 `validate.py`. Exact byte identity is not verifiable read-only | condition (§5) |
| §4.4 digest table | verified as text | 44 rows, whose names equal the `signoff.json` `files` keys. The lead and the peer each compared the 42 rows other than `PREREGISTRATION.md` and `README.md` one by one with `reviews/round-12-response.md` §3: all equal. `PREREGISTRATION.md` `fd30e0f4…2a79` → `8b3e6375…f9db`; `README.md` `fc51642b…6124` → `c170b870…fa74`. `validate.py` `a8ce7541…f7c4`. Tune pin `2f850933…2e4c` = `signoff.json` `tune_script_sha256`. The values are not hashable read-only | condition (§5) |
| §4.5 no write to the run data | consistent | `version.json` still holds the rev-11 `thresholds_sha256` and `question_version_keys: null`. There is no `version.rev11.json`, `jev-n1/` or `jev-p/`. That nothing was written cannot be proven read-only | condition (§5) |
| R12-01 re-freeze | stands as signed | The `PREREGISTRATION.md:242` text is unchanged from round 13 (the R13 fixes touched only `:3`, `:14`). The challenger re-traced `versioning.py:79-83,132-144` (it reuses `inputs/` and `settings.json`), `:117-129` (exactly two keys differ) and `:162-166,205,240-249` (the drift and the round-1 check). No defect | — |
| Other frozen statements made false by rev 12 | 1 found | Grep over the 44 frozen files for `$1`/`$10`/"unrun"/"no label"/"nothing"/"draft for council"/"not re-run". The remaining cap figures are historical (`:14`, `:156`, `:192`, `:242`, `README.md:69` "($10 in rev 4, $20 from rev 12)"); Go defaults are overridden by `run.sh`. One false statement: `PREREGISTRATION.md:16` "HD6 (`docs/hypothesis-ledger.md`, `proposed`; becomes `preregistered` at sign-off)". The line was rewritten in rev 12 (the R-052 hunk). It contradicts `:242` "HD6 stays `preregistered`" and ledger `:32`/`:39` (preregistered since round 9) | **R14-02** |
| R12-02 (d) attestation | not yet due | `reviews/pilot-round-0.md` still ends "continues with `gates-label`" (`:42`). It is scheduled after sign-off (§5 step 1) | condition (§5) |

## 2. Issues

| id | severity | finding | exact fix |
|---|---|---|---|
| R14-01 | minor (rule 6; the R13-01 class, introduced by the council's own round-13 wording) | `README.md:15` "run on live data from the rev-11 sign-off" misstates which sign-off governed the live runs. Onboarding, catalogue freeze and quality ran at `730674e` under the rev-9 sign-off; `gates-build` onward ran at `0b56594` under rev 11 (`PREREGISTRATION.md:241`; `reviews/pilot-round-0.md:14-17`). This is the two-commit fact that RESULTS `Deviations` must report, and the frozen README would contradict it | (a) `README.md:15`: "exists (validated model-free by `run.sh validate`; run on live data from the rev-11 sign-off)" → "exists (validated model-free by `run.sh validate`; run on live data since the rev-9 sign-off: onboarding, catalogue freeze and quality at `730674e`, `gates-build` onward at `0b56594`; `PREREGISTRATION.md` §14 item 18)". (b) `PREREGISTRATION.md:14`: "`:3` and `:15` state what exists, R13-01" → "`:3` and `:15` state what exists, R13-01, R14-01". No other README byte moves |
| R14-02 | minor (rules 5–6 and ledger rule 2: the rev-12 document contradicts itself and the ledger on the hypothesis state that resume step 10 records) | `PREREGISTRATION.md:16`, rewritten in rev 12, still says HD6 is "`proposed`; becomes `preregistered` at sign-off". HD6 has been `preregistered` since round 9 (ledger `:39`), and item 19 of the same revision (`:242`) says "HD6 stays `preregistered`". Read at the rev-12 sign-off, `:16` calls for a false "proposed → preregistered" ledger transition | (a) `PREREGISTRATION.md:16`: "Hypothesis **HD6** (`docs/hypothesis-ledger.md`, `proposed`; becomes `preregistered` at sign-off)." → "Hypothesis **HD6** (`docs/hypothesis-ledger.md`; `preregistered` since the rev-9 sign-off, council round 9; it stays `preregistered` at the rev-12 sign-off, §14 item 19)." (b) `PREREGISTRATION.md:14`: "the ledger-sources line," → "the ledger-sources line (R-052; the HD6 state, R14-02)," |

Consequential edits for both issues (exact; one match each):
- (c) `PREREGISTRATION.md:14`: "council rounds 12–13: R12-01–R12-03, R13-01, R13-02" → "council rounds 12–14: R12-01–R12-03, R13-01, R13-02, R14-01, R14-02".
- (d) `PREREGISTRATION.md:3`: "DRAFT for council round 14" → "DRAFT for council round 15".
- (e) `docs/hypothesis-ledger.md:37` (main agent; not frozen): "council rounds 12–13 raised R12-01–R12-03 and R13-01–R13-02, fixed in revision 12; council round 14 pending" → "council rounds 12–14 raised R12-01–R12-03, R13-01–R13-02 and R14-01–R14-02, fixed in revision 12; council round 15 pending".

Notes (below the bar; no fix required for sign-off):
- (i) `DRAFT-NOTES.md:122` (not frozen):
  - "Responses: `reviews/round-11-response.md`, `reviews/round-12-response.md`" → add "`reviews/round-13-response.md`, `reviews/round-14-response.md`";
  - "`README.md`: the rev-4 note's total." → "`README.md`: the rev-4 note's total; `:3` and `:15` state what exists (R13-01, R14-01).";
  - add "`PREREGISTRATION.md:16`: the HD6 state (R14-02)";
  - name council round 14.
- (ii) Compute the round-15 `-U0` diff and numstat with `signoff.json` in place, and print them unfiltered. Then delete the response's parenthetical about `signoff.json`.
- (iii) `README.md:3` "The files marked *exists* below were produced by the model-free pre-pass and one Clawpatch measurement map" (the challenger's R14-C01) and `README.md:14` "`data/` | empty until sign-off" date from rev 1. Rev 12 did not make them false, and they move no gate or RESULTS figure. Do not change them.
- (iv) Round-13 notes (i)–(iv) still hold. In particular, run the re-freeze step (2) from the run-worktree root as `.tools/venv/bin/python evaluation/x2d-merge-association-pilot/versioning.py --round 0`, with `X2D_DATA` unset or equal to `run.sh`'s `D`, and record the command.

## 3. Synthesis

- **R13-01, R13-02:** all three reviewers found every replacement in place word for word and the old text gone.
- **R14-01:** raised by the peer; the lead had marked `:15` verified as applied. The peer's check of the two-commit record (`:241`, `pilot-round-0.md:14-17`) changed the lead's verdict. Round 13's prescribed wording was wrong. The council owns the error and gives a fix tied to item 18.
- **R14-02:** raised by the challenger. The peer and, at first, the lead held it below the bar, reading it as a condition already met. The lead upgraded it for two reasons. First, rev 12 rewrote this line (it is the `:16` hunk). Second, item 19 of the same revision now says "stays `preregistered`", so rev 12 leaves an internal contradiction about the state that resume step 10 writes to the ledger. Fixing it adds no hunk, because `:16` is already in the rev-12 diff.
- **R14-C01 (README `:3` "exists" provenance):** the challenger's second point. The lead and the peer hold it below the bar (note iii). It predates rev 12, the Modules row says what ran, and it changes no gate or RESULTS figure.
- **R12-01:** all three reviewers keep the preregistered re-freeze. The challenger again judged a dedicated transition command viable but without benefit.
- **Agreed by all three:** the `validate.py` hunks and the 20/13 numstat; 42 of 44 digests equal round-12-response §3; the 651/0 log; no write to the run data seen.

## 4. Required for round 15

1. Apply the fixes for R14-01 and R14-02 and the consequential edits (c)–(e) only, using exact-substring replacement with one match each. No other frozen byte moves. `validation/validate.py` stays at `a8ce75417cc883b0ad73d36b95c4536744271686c2da746859dfc8e02eeef7c4`; `x2d.py`, `run.sh` and `thresholds.json` keep their round-13-response §4 digests.
2. Recompute `git diff -U0 0b5659440c7368ac18b82f60a4edf3956408d2bf -- evaluation/x2d-merge-association-pilot docs/research-ledger.md docs/hypothesis-ledger.md` and `--numstat` with `signoff.json` in place, unfiltered. Expected:

   | file | numstat | new-file hunks |
   |---|---|---|
   | `PREREGISTRATION.md` | 6/4 | {3, 14, 16, 156, 192, 242} |
   | `README.md` | 3/3 | {3, 15, 69} |
   | `validate.py` | 20/13 | {218, 678, 768, 771–779, 783, 786, 789–794} |

   Every other row must be as in round-13-response §3, with no `signoff.json` row. Any other result reopens review.
3. Run validation (a) with `signoff.json` moved aside: 651 checks, 0 failed. Record the log path and the record's SHA-256 before and after the move.
4. Recompute the 44 frozen-file digests after (3). Expected: 42 rows equal `reviews/round-13-response.md` §4 (equal to `reviews/round-12-response.md` §3), and new digests for `PREREGISTRATION.md` and `README.md` only. Tune pin `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c`.
5. No write to `../scout-x2d/…/data` before sign-off.

## 5. Resume order after a future zero-issue sign-off (for reference; not authorised by this round)

Sign-off conditions, all before step 1:
- the 44 digests recomputed and equal to the signed table, with the tune pin;
- a new `signoff.json` (round `reviews/round-15.md`) whose `code_under_test` equals round 11's (17 rows);
- validation (a) 0 failed;
- validation (b) with the new record in place, showing exactly the one expected failure (`C without the council's sign-off record…`);
- one commit;
- the run worktree moved by checkout only.

Order:
1. The R12-01 round-0 re-freeze, steps (1)–(5) of `PREREGISTRATION.md:242`, with the command (note iv). Record it, together with the R12-02 (d) attestation, in `reviews/pilot-round-0.md`. The attestation is appended with the `gates-label` record and the `gates-jev` stop record (exit 4; 7,755 of the `main` run's 7,872; $1.0001; seven `network` / `cost_known: false` failures; `n1` / `p1` / `p2` not started; R-052). If the main agent cannot attest, or step (3) shows any other difference, stop for a new council round.
2. Resume `gates-jev`.
3. `gates-analyse`.
4. `power`.
5. The round-0 critique (its `version_sha256` is the re-frozen file's digest).
6. Rounds 1–7, each with its critique.
7. Frozen rounds 8–10.
8. `verdict`.
9. `RESULTS.md`. Its `Deviations` cover the two commits, the cost stop and R-052, and the re-freeze.
10. Ledger states (HD6 per the §14 map; it stays `preregistered` at the rev-12 sign-off).

## Verdict

**2 ISSUES** (R14-01 minor, R14-02 minor). Rev 12 is not signed off. `signoff.json` stays at the round-11 record and no stage may run. Re-review in round 15 after the fixes in §2 and the items in §4.
