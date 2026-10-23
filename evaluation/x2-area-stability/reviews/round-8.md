# X2 — Innovation council experiment review, round 8

Reviewed files (SHA-256): `PREREGISTRATION.md` `a39513d6eabddce467cca26cf0ea5085e15ab3e615e34a571f9603e5c2a3c426`; `GUIDE.md` `9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799`; `common.py` `2cd742020c609c02fa09ad504812382493583a4b5f0c6fac831e86dac391741c`; `collect.py` `7e22b60360c4082d62400a0582f6b369cc9c89f556372065d30b79cba698a133`; `replay/main.go` `f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e`; `build_pairs.py` `face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a`; `label.py` `87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f`; `analysis.py` `7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63`; `run.sh` `5157b15be6013c350b1dde006b0ee4833766d46981b131df756b1f838441c41b`; `validation/validate.py` `beb32dc2527aa01abda63a0650c93390dc38e14e5dd24f56f9452dc465d286f6`; `validation/stub_clawpatch.js` `0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226`; `validation/credcheck/main.go` `e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779`

Scope: `evaluation/x2-area-stability/` revision 8 (PREREGISTRATION.md rev 8, GUIDE.md, README.md, DRAFT-NOTES.md incl. interpretation choices 40–41, `common.py`, `collect.py`, `replay/main.go`, `build_pairs.py`, `label.py`, `analysis.py`, `run.sh`, `validation/validate.py`, `validation/stub_clawpatch.js`, `validation/credcheck/main.go`), `reviews/round-7.md` (2 minor issues) and `reviews/round-7-response.md`, `docs/hypothesis-ledger.md` HD5 (`:29`, history `:33`), against `AGENTS.md` experiment rules 1–7. Since round 7 only `PREREGISTRATION.md` and `validation/validate.py` changed; the eleven other hand-over digests equal the round-7 digests (`DRAFT-NOTES.md:8-20` against `round-7.md:72` and `round-6-response.md:27-39`, transcription agreement; not recomputed by the council). Council: chair (independent pass), `X2R8Challenger` (innovation-challenger, GPT family) and `X2R8Peer` (experiment-peer, Opus) in parallel, synthesised in §7. No model call, no data, no file edited, no secret printed, `.codegraph` untouched; the 225 passing checks are hand-over evidence (`DRAFT-NOTES.md:4`), not exercised by the council.

## 1. Verdict: ZERO ISSUES — signed off (conditional on the main agent's digest recomputation, §10).

Both round-7 issues are fixed as required at the named locations, no decision logic changed (`analysis.py` digest `7d70e630…` unchanged since round 3), nothing regressed in S1's unplanted assertions, and every single-token mutant named in round 7 issue 1 is now killed by a named assertion through `A.analyze`. The edited lines (validate.py `:568-602`, prereg `:3`, `:144`, `:199`, DRAFT-NOTES `:3-4`, `:72-73`, ledger `:29`, `:33`) introduce no contradiction that could change a verdict, mislead `RESULTS.md`, break rules 1–7 or make a run unrepeatable.

## 2. Round-7 verification

| # | Round-7 issue | Status | Evidence read (council) |
|---|---|---|---|
| 1 | `analyze → replication` wiring (`analysis.py:502-504`) never exercised with real probe files; required (a) planted probe-consistent files → `replicated`, (b) two-arm case with B's decision ≠ A's → `not replicated`, recommended (c) rolled claude probe → `not citable (arm B …)` | **verified** (a, b and c) | `validate.py:569-575` `plant_versions(data, claude_axum_after="n1")` writes all four `tool-versions-{codex,claude}-{fastapi,axum}.json` with `model_probe_before/after = {"model": m_, "error": None}` (codex `m1`, claude `n1`), `cli_versions_after` plus the same CLI keys top-level (`codex: c1`; `claude: v1`, `claude_executable_version: e1`, matching `analysis.py:412` `ARM_CLI_KEYS`), `workspace_config_presence` all `False` — the `tvf`/`ok_p` shape (`:470-472`). **(a)** `:576-585`: `shutil.copytree(d1/data, dv_ok)`, plant, `rv = A.analyze(dv_ok, rp1, p1)`; asserts both `probe_consistency[arm]["consistent"]`, both `decision.citable is True`, `replication.status == "replicated"`, `arm_A_citable is True`, `arm_B_citable is True`, `arm_B_verdict == arms.claude.decision.verdict == "adopt"`, `arm_B_threshold_pct == 75`. S1's unplanted assertions on `r1` are unchanged at `:562-566` (`dA1/dB1 citable False`, `r1["replication"]["status"] == "not citable (arm A probe-inconsistent)"`). **(c)** `:586-593`: copy of `dv_ok`, replanted with `claude_axum_after="n2"`, `rr_ = A.analyze(dv_roll, rp1, p1)`; asserts codex `citable True`, claude `citable False`, status `not citable (arm B probe-inconsistent)`, `arm_A_citable True`, `arm_B_citable False`. **(b)** `:594-602`: new scenario `S1b` via `run_scenario(tmp, "S1b", s1, rule=lambda k: "different_concern" if k["arm"] == "claude" else default_rule(k))`, files planted, `rb_ = A.analyze(...)`; asserts codex `adopt` 75, `arms.claude.decision.verdict != arms.codex.decision.verdict`, status `not replicated`, `arm_B_verdict == arms.claude.decision.verdict`, `arm_B_threshold_pct == arms.claude.decision.get("threshold_pct")`, `arm_B_citable True`. S1b's construction is sound: `truth_labels` applies `rule` to real pairs only (`:162`) and controls get `c["expected"]` (`:163-164`), so the pooled control gate passes (confirmed by arm A still adopting at `:598`); pair ids are hashes of the evidence (`build_pairs.py:15`) and every synthetic feature carries `tok = sha1("arm-repo-map")` (`validate.py:85-86`), so no `pair_id` is shared across arms and the per-arm rule cannot leak. `analyze` loads planted files through `inventory` (`analysis.py:407-408`) and `probe_consistency` keys them by `tool-versions-%s-%s.json` (`:425`). §8 `:199` new "Round-7 additions" bullet names the file pattern, the three fixtures and their expected statuses; DRAFT-NOTES 40 (`:72`) and 38 (`:70`) describe the fixture that exists; "three checks added (225)" is consistent with 222 + 3. |
| 2 | §4 `:144` "Citability covers the served-model name only" contradicted §2/§10/code | **verified** | `PREREGISTRATION.md:144` now reads, word for word as required by `round-7.md:34`: "Citability is the §2 probe flag (served model identical across all probes of the arm, every probe taken without error, every `tool-versions` file present, no workspace-level configuration file present); a listed CLI-version change or a labeller-model change is reported beside the verdict and the ledger outcome, not folded into the flag." Agrees with §2 `:39` (file present, probes taken without error, identical values, no workspace configuration file; CLI versions "does not enter the flag"), §10 `:217` (roll, failed probe, split, workspace file → inconsistent), and `analysis.py` reasons: missing file `:427`, probe not taken `:439-440`, failed probe `:441-442`, workspace config present `:448-450`, differing models `:451-452`, `cli_versions_changed` outside `reasons` `:453`; `:500` folds the flag into `decision.citable`. DRAFT-NOTES 39 (`:71`) records the rev-7 sentence as wrong and 41 (`:73`) the replacement; the optional `:139` scoping was declined (allowed). |

Revision metadata: `PREREGISTRATION.md:3` "DRAFT revision 8 (council rounds 1–7 addressed: … `reviews/round-7-response.md`), not signed off"; `DRAFT-NOTES.md:3` "draft revision 8, awaiting council round 8", `:4` 225 checks; `docs/hypothesis-ledger.md:29` HD5 Tests cell "draft preregistration rev 8, council rounds 1–7 addressed", state `proposed`; history `:33` dated 2026-10-23, "wording only, no state change", names `round-7-response.md`, the §2-flag wording and the planted fixture. Response digest block equals `DRAFT-NOTES.md:8-20` string for string; README `94ea650f…` equals round 6/7.

## 3. New issues

None.

## 4. Not raised (recorded, not issues)

- `analysis.py:284` `"arm_B_gates_ok": dec_b.get("gates_ok")` → `dec_a.get(...)` survives every check (peer): the field is reporting-only (rendered replication JSON line), the status reads `:290` independently, and S1b's two arms both have `gates_ok True`. Outside the edited lines; code unchanged since round 3. Optional one-line kill in a future round if the frozen set is ever reopened: assert `rb_["replication"]["arm_B_gates_ok"] == rb_["arms"]["claude"]["decision"]["gates_ok"]` in a fixture where the arms differ.
- §8 `:199` "the claude after-probe rolled" — the fixture rolls only the claude **axum** after-probe (`:572`, `:588`); less specific, not wrong; one differing probe suffices (`:451-452`).
- S1b asserts claude's verdict only as `!= codex` (`:599`); by construction it is `fallback` with `threshold_pct None`, `gates_ok True`, status `fail` at 75 [INFERENCE: traced from spurious = 1 at every threshold, C2 failing everywhere; not executed]. `:601` is therefore `None == None`, which still kills a `dec_a` swap at `:283` (75 ≠ None).
- `plant_versions` reproduces the `tvf`/`ok_p` shape rather than calling the unit-local helpers, and lists three presence names instead of §2's five; `probe_consistency` reacts only to truthy entries (`:448-450`), so the flag is unaffected; presence on real collector output is asserted by the round-4 fixtures.
- In S1, both arms carry identical planted data, so the positive fixture (a) alone cannot separate the arms; S1b (b) and the unit cases do.
- DRAFT-NOTES 29 (`:61`) still says the synthetic scenarios hold no versions files; choice 40 states the planted exceptions. Historical wording.
- §4 `:144` bracket "served model identical" summarises §2's claude value (the sorted `modelUsage` key set, `:39`), and the sentence opens with "is the §2 probe flag"; no residual contradiction.
- §4 `:139` `fallback` consequence still unscoped (declined optional edit; consequences are a design-round matter, no verdict changes) — unchanged from round 7.
- Round-7 Not-raised items (`:290` `!= "fail"` two-token mutant, ledger Tests-cell qualifiers, docstring at `:279-280`) unchanged; the `:500` `.get("codex")` survivor is now killed by `:591`.
- 225 passing checks and the digest block are hand-over evidence, not exercised by the council.

## 5. Adjudication of the response and DRAFT-NOTES 40–41

Response issue 1 — **accept** (a, b and the recommended c all present, through `A.analyze`, `r1` untouched). Issue 2 — **accept** (exact wording). "Verification: 225 checks, three added" — consistent with `:580`, `:590`, `:597`. DRAFT-NOTES 40 — accept. 41 — accept. Ledger row and history — accept.

## 6. Checked and found correct

Mutant kill table against the new assertions (static traces, chair + peer + challenger; [INFERENCE: not executed]): `:502` `.get("claude")`→`.get("codex")` and `:504` `(dec, dec)`/`(rep, rep)`/`(rep, dec)` → status `replicated` or `not applicable` with `arm_B_verdict` ≠ claude's → killed by `:600`; `:504` `res["probe_consistency"]`→`res["controls"]` (keys `per_arm`, `analysis.py:80`) or `res["inventory"]` (keys `maps`, `stopped`, `tool_versions`, `:391`) → `cit` false → `not citable (arm A …)` → killed by `:583`; `:500` `.get(arm)`→`.get("codex")` → claude citable on the rolled data → killed by `:591`; `.get("claude")` → killed by `:591` and the collector run `:1112`; `:291` `arm_A_citable`→`arm_B_citable` → killed by `:592` and U `:354`; `:293` drop `replicated and` → U `:354`; `:296` branches swapped → `:583`, `:600`. `validate.py:349-357` unit cases unchanged. `:565` on the unplanted `r1` unchanged. No later check scans `d1` or `tmp` (no `listdir`/`walk` on them), so the two planted copies under `d1` cannot leak into another scenario's inventory. §11 `:220-222` frozen list and freeze rule unchanged; README `:1-45` unchanged (digest).

## 7. Peer synthesis

Chair's independent pass: both fixes verified; S1b control-gate and pair-id isolation checked at `validate.py:162-164`, `:85-86`, `build_pairs.py:15`; verdict zero issues. **Challenger** (`X2R8Challenger`): same verification; traced the four named mutants to killing assertions with the concrete key sets of `controls`/`inventory`; established that `not applicable` cannot satisfy S1b because `:598` requires arm A's `t*`; verdict zero issues. **Peer** (`X2R8Peer`): same verification; full single-token mutant table over `:281-296` and `:500-504` — every mutant killed except the reporting-only `:284 arm_B_gates_ok` swap (adopted as Not raised); noted the §8 `:199` "claude after-probe" imprecision (Not raised); verdict zero issues. No disagreement to resolve.

## 8. Risks (recorded, not issues)

Unchanged from round 7 §8: ABBA places both claude collections mid-window; a roll inside a collection, a transient roll or a same-name weight change stays invisible; Claude Code self-updates are listed, not flagged; `--safe-mode`, `codex exec --cd <tmp>` and the output formats are first exercised in the real run; SIGKILL leaves no stop record; a labeller roll is visible, not consequential; controls bound inattention and gross anchoring only; `workspace_dirty` on a real map is descriptive and must be reported beside T1–T4. These are preregistered threats (§10), not open issues.

## 9. Execution outline (run)

1. Main agent recomputes the digests with the two `shasum -a 256` commands in `DRAFT-NOTES.md:29` and confirms they equal §10 below; any difference voids this sign-off (rule 5).
2. Ledger: HD5 `proposed` → `preregistered` with a dated history line naming this round (`docs/hypothesis-ledger.md:29`, `:31-`).
3. Run per `README.md`/§9 under `nohup`, in the ABBA order, with `OPENROUTER_API_KEY` injected only via `infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- <command>`; no peeking beyond map success/failure and counts (§11).
4. Any script or preregistration change after this round → new review round with zero issues before continuing (rule 5); runtime deviations → `Deviations` in `RESULTS.md`.
5. `RESULTS.md` reports everything preregistered incl. `unmeasured` counts, intervals, probe-consistency lines, replication status, `decision.citable` per arm, reproduce commands (rule 6).

## 10. Sign-off

The innovation council signs off X2 at revision 8 with zero issues, **conditional on the main agent's recomputation matching the following digests** (SHA-256, as handed over in `DRAFT-NOTES.md:8-28`; not recomputed by the council):

Frozen set (`evaluation/x2-area-stability/`):
- `PREREGISTRATION.md` `a39513d6eabddce467cca26cf0ea5085e15ab3e615e34a571f9603e5c2a3c426`
- `GUIDE.md` `9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799`
- `common.py` `2cd742020c609c02fa09ad504812382493583a4b5f0c6fac831e86dac391741c`
- `collect.py` `7e22b60360c4082d62400a0582f6b369cc9c89f556372065d30b79cba698a133`
- `replay/main.go` `f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e`
- `build_pairs.py` `face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a`
- `label.py` `87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f`
- `analysis.py` `7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63`
- `run.sh` `5157b15be6013c350b1dde006b0ee4833766d46981b131df756b1f838441c41b`
- `validation/validate.py` `beb32dc2527aa01abda63a0650c93390dc38e14e5dd24f56f9452dc465d286f6`
- `validation/stub_clawpatch.js` `0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226`
- `validation/credcheck/main.go` `e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779`
- `README.md` `94ea650f24854869f8351ccf4a4fdf80b9bbaf1956c722e8c6494dbaf53a4744`

Code under test / ported / drift-guarded (repository root):
- `internal/areas/areas.go` `8ed2e70815fdcc605d60402fb603cfb905d58457a5869b3795d59507a93cb484`
- `internal/clawpatch/records.go` `57344e025d5e2d0b7a10abf543b20b55f899602c1dd95d2b6a2b62d9b823dfbe`
- `internal/modelcli/modelcli.go` `bf10ddd58e0c2ecf08372d79505177b00c187955f3d64b3be9a3e8eb7c8617b8`
- `internal/clawpatch/client.go` `209639260283c373b7376bab352f580723a348bb2d6949bed7549c4ab03d910b`
- `internal/clawpatch/exclude.go` `615fcb0eebf047d9637d4d71cfe4e2faff7e730b840b8f8cbd4e032ed6384363`
- `third_party/clawpatch/src/providers/claude.ts` `fadf1e447d9b0b9a32ba7db6ca9101d29a84acce5c575f37423995456d03d059`
- `third_party/clawpatch/src/providers/codex.ts` `994e8066be12c3051c4cb63b32e2144bcf4bd495c5d17a8033f94814bd7a6e92`
- `third_party/clawpatch/src/config.ts` `97413a275363a52e05c3b17d08449b211fdc26584bd60155c893a647720381ce`

Signed-off preregistration: `PREREGISTRATION.md` revision 8 (`a39513d6…`); signed-off validation: `validation/validate.py` (`beb32dc2…`), 225 checks as handed over. The experiment may be run only against these exact digests; any change requires a new council round with zero issues (rule 5).
