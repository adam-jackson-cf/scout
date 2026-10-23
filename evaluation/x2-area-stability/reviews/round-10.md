# X2 — Innovation council experiment review, round 10 (Amendment 1, prereg rev 10)

Reviewed files (SHA-256): `PREREGISTRATION.md` `ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355`; `GUIDE.md` `9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799`; `common.py` `e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3`; `collect.py` `89948f3c6d1b308b91a5ac5a8060088269270806546fe5c577dea307a4dfc83d`; `replay/main.go` `f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e`; `build_pairs.py` `face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a`; `label.py` `87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f`; `analysis.py` `7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63`; `run.sh` `5157b15be6013c350b1dde006b0ee4833766d46981b131df756b1f838441c41b`; `validation/validate.py` `29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784`; `validation/stub_clawpatch.js` `0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226`; `validation/credcheck/main.go` `e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779`

Reviewed files (SHA-256 as handed over in `DRAFT-NOTES.md:8-31` and `reviews/round-9-response.md`; not recomputed by the council, which has no shell): `PREREGISTRATION.md` (rev 10) `ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355`; `validation/validate.py` `29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784`; `README.md` `13e04e61720fa0542d391a48f2d63db4e1973f29096114ba9120bddef84c4048`; all other entries unchanged from round 9 (listed in §10).

Scope: second round of the Amendment 1 review (AGENTS.md experiment rule 5). Inputs: `reviews/round-9.md` (1 minor issue, `:29`; §4 Not raised; §9 outline), `reviews/round-9-response.md`, `PREREGISTRATION.md` rev 10 (`:3`, `:221`, `:226-234`, the fixed sentence `:232`), `README.md` (`:4`, `:18`, `:24`, `:28`, `:41`, `:44-45`), `DRAFT-NOTES.md:1-32`, `validation/validate.py` (edited regions `:476-501`, `:1070-1082`, `:1084-1130`, `:1157-1159`, `:1300-1350`; fake-CLI logging `:1092-1094`, `:1099-1101`), `common.py:160-212`, `collect.py:267-295`, `:385-406`, `validation/stub_clawpatch.js` (token absence), `docs/hypothesis-ledger.md:29,33`. Since round 9 only `PREREGISTRATION.md`, `validation/validate.py` and `README.md` changed (hand-over claim; consistent with the digest block and with the line-offset trace in §6). No model calls, no secrets printed, `.codegraph` untouched.

## 1. Verdict: ZERO ISSUES — signed off (conditional on the main agent's digest recomputation and a green `run.sh validate`, §10)

Round-9 issue 1 is closed by option (ii): the three missing assertions exist at the named lines and each discriminates (a plausible mutant per gap is killed, §2). The rev-10 `:232` sentence matches `validate.py` claim for claim (§2, table 2). The edited lines introduce no new inaccuracy: revision and round numbers are consistent across prereg `:3`, `:221`, `:227`, `:232`, README `:4`, `:18`, DRAFT-NOTES `:3-4` and the ledger `:33`; the check count 232 → 233 is structurally explained (one more iteration of the collector-refusal loop; the new unit case extends an existing aggregate check, §6); no previous assertion was weakened (§2, H5). Chair, peer and challenger independently reached zero issues.

## 2. Round-9 issue verification

### 2.1 The three gaps

| Gap | Assertion (exists?) | Discriminates? (mutant → killing line) |
|---|---|---|
| (a) collector wrong-sha refusal | **Yes.** Fixture `stubbuild-wrongsha` `validate.py:1331-1335`: stub as `dist/cli.js`, token in `dist/prompt.js`, marker `"<pinned version> <0×64>"`. In the loop `:1343` as `("wrongsha", wsb, "not the pinned")`; refusal asserted `:1347-1350`. Reason string traced: `common.py:199` `marker_ok = marker == expected` → False → `:206-207` `"build marker %r is not the pinned %r"`; `collect.py:282` wraps it in `"… is not the review-safety build (%s)"`, so both `"review-safety build"` and `why_` appear. | **Yes.** Mutant: compare the marker's version only (`marker.split()[0] == expected.split()[0]`) at `common.py:199`. `wrongsha` then passes the build check, the collector proceeds (exit 0) → `rb_.returncode != 0` fails at `:1348`; also killed at the unit check `:499` (`"right version, wrong sha"` must be refused). No other fixture catches this: `oldmarker` differs in version, `nomarker`/`nosanitizer` fail elsewhere. |
| (b) "before any probe" | **Yes.** `probe_calls()` `:1337-1342` sums lines of `bins/codex.calls` and `bins/claude.calls`; `:1345` baseline; `:1349` requires `probe_calls() == calls_before_ and calls_before_ > 0`. Wiring verified: both fake CLIs `sys.exit(0)` on `--version` before the `.calls` write (`:1092-1093`, `:1099`) and append exactly one JSON line per other invocation (`:1094`, `:1101`) to `abspath(__file__) + ".calls"` = `bins/<cli>.calls`; `collect()` prepends `bins` to `PATH` (`:1117`); refusal runs use the default codex arm (`:1114`) and `probe_model` invokes `codex exec` through PATH. `calls_before_ > 0` is guaranteed by the earlier successful runs (`:1258`, `:1270` already read both files; `:1272` requires ≥ 4 codex calls). No `wait=False` collector is live at `:1343` (first background run starts later, peer trace). The stub Clawpatch logs to `X2_STUB_ENVLOG` (`<data>-envlog.jsonl`, `:1116`), never to `.calls`. | **Yes.** Mutant: move `C.build_safety_check` (`collect.py:280-282`) after `probe_model` (`:394`) and before `os.makedirs(base)` (`:399`). The refusal still leaves no workspace, no `maps/`, no `tool-versions-*` (round 9's exact objection), but the fake codex logs one line → `probe_calls() != calls_before_` → `:1349` fails. This conjunct is the only one that kills the mutant. |
| (c) sanitizer in the entry file | **Yes.** Unit case `"patched, sanitizer in the entry file": fake_build("ok3", exp_marker, "", {"cli.js": san})` `:493`; `fake_build` writes only the supplied files (`:480-488`), so `build-ok3/dist/` holds a single `.js` (`cli.js`) with the token; required `ok` at `:498`. Check name `:497` now says "in the entry file cli.js, and in a non-entry file". | **Yes.** Mutant: skip the entry file in the dist scan (`common.py:202` `and name != "cli.js"`). `ok3` → `sanitizer_in_dist` False → `ok` False → `:498` fails. The converse mutant (scan `cli.js` only) is killed by `patched` (token in `prompt.js`) and `ok2` (token in `sub.js`). |

### 2.2 `PREREGISTRATION.md:232` (rev 10), claim by claim

| Claim | Backing |
|---|---|
| USER/LOGNAME port: copied; empty not copied; `SHELL` not copied | `validate.py:441-449` (exact key set, `USER="op"`/`LOGNAME="op2"`, empty `USER` dropped, `SHELL` absent) — unchanged since round 9 |
| patched-source drift guards | `:456-474` |
| pins / `review.go` / patch-digest agreement | `:476-478`; `common.py:168-178` |
| build-safety check (unit) passes: token in the entry file `cli.js`, and in another dist file | `:492-493` (`patched` → `prompt.js`; `ok2` → `sub.js`; `ok3` → `cli.js`) → `:498` |
| unit refuses: no marker, old marker, wrong sha, marker over unpatched `dist/` | `:494-495` → `:499-501` |
| collector refuses each of those four | loop `:1343`: `nomarker`, `oldmarker`, `wrongsha`, `nosanitizer` |
| "before any probe" | `:1349` `probe_calls()` unchanged |
| "clone" | `not os.path.exists(wb_)` `:1348`; the single `git clone` (`collect.py:400`) targets `base/ws` under `a.work` created at `:399`, so a clone implies `wb_` exists |
| "data file" | no `maps/`, no `tool-versions-*` `:1348-1349`; `os.makedirs(a.data)` is at `:405`, after the clone |
| "(asserted: nonzero exit, the build-safety reason, no workspace, no `maps/`, no `tool-versions-*`, no model call logged by the fake CLIs)" | `:1348-1349`, each conjunct present; the parenthesis enumerates exactly what is asserted, nothing more |
| `clawpatch_build_safety` recorded as `ok` | `:1157-1159` (`ok`, `marker_ok`, `sanitizer_in_dist`, `marker == expected == "<pv> <ps>"`) |
| stub lives in a fixture build directory carrying the pinned marker | `:1077-1082` (`stubbuild/dist/cli.js`, `prompt.js` with token, marker = real pins) |

Round-9's three defects are gone: "the last three" is replaced by the explicit four-item list that equals the fixtures; "before any probe" is now asserted; the entry-file claim is now tested. No assertion round 9 described at `:1333-1335` was dropped (H5: `returncode != 0`, `"review-safety build"`, `why_`, no `wb_`, no `maps/`, no `tool-versions-*` all retained).

## 3. Issues

None.

## 4. Not raised (recorded, not issues)

- **Sha-only marker comparison survives every fixture** (peer). If `common.py:199` compared only the sha half, `oldmarker` (`f39d73da`) and `wrongsha` (`0×64`) would still be refused on the sha; no fixture varies the version alone. `:232` makes no version-only claim and the sha pins the patch content, so no reportable consequence. Optional fixture: `"<v2> <pinned sha>"`.
- **Refusal order shown for the codex arm only.** The build check (`collect.py:280-282`) precedes the claude-only branch (`:283`) and `main`'s probe (`:394`) unconditionally, so arm does not matter; a claude-arm refusal run would be optional hardening.
- **"No data file" is indirect**: it rests on the asserted artifacts plus the ordering `:399 < :405 < :406`; `stopped-*`, `attempts/` and the data directory itself are not asserted absent. The `:232` parenthesis states exactly what is asserted, so the text is accurate.
- **Prereg `:3`** calls the round-9 finding "one validation-coverage wording issue"; the fix also changed `validate.py` (option ii). Not false — round 9 classed it as frozen-text accuracy — and `round-9-response.md` records both halves.
- **Timing**: README `:24` "about 3 min" vs DRAFT-NOTES `:4` "~2.5 min" (measured 155 s per the response). Both approximate; README is outside the frozen set.
- **`probe_calls()` counts every non-`--version` fake-CLI invocation**, not only probes (e.g. the direct fake-control calls at `:1277-1283`), and ignores `altbin/claude-alt.calls`. Harmless: the assertion is a synchronous delta around a single collector run on the standard fakes.
- **Round-9 §9 step 3** named `internal/clawpatch/review.go` for the round-10 sign-off; prereg `:221` freezes only `scripts/clawpatch-safe.patch` and `scripts/build-clawpatch.sh`. Settled in round 9: `review.go` is cross-checked (`common.py:171-173`), not frozen; its digest is named in §10 as cross-checked.
- **Static review cannot prove the absence of a same-line-count edit elsewhere in `validate.py`**; the line-offset trace (§6) fits edits only at `:493`, `:1331-1342` and the loop. The main agent's `git diff` against the round-9 commit (§9 step 1) closes this.
- Round-9 §4 items (build-safety not bound to the executable; negative `review_safety_pins` branches; `patched_sources` failure paths; time-of-check gap at `collect.py:338`; `:6`/`:228` qualifiers; `__pycache__/`) are unchanged and not reopened by these edits.
- 233 passing checks, `go vet`, `gofmt -l` and the digests are hand-over evidence, not exercised by the council.

## 5. Judgement calls

Round-9 judgement calls 1–7 stand (accepted in round 9 §5). One new call: **option (ii) over option (i)** — adding assertions rather than narrowing the text. **Accepted** (chair, peer, challenger): the three new assertions each discriminate a real regression (§2.1) without touching collection or analysis behaviour; the S1/S12/P sections are claimed untouched (to be confirmed by diff, §9).

## 6. Checked and found correct (static traces; [INFERENCE: not executed])

- **Check-count arithmetic**: round 9 had 232; the loop `:1343` now has 4 iterations (was 3) → +1 `check(` call; the `ok3` case is a new key inside the existing aggregate check `:497-501` → +0. Total 233, as DRAFT-NOTES `:4` and the response state.
- **Line-offset trace** (peer): anchors before `:493` unmoved (`:447`, `:456-474`, `:477`); `:496→:497`, `:1157→:1158`, `:1258→:1259`, `:1318→:1319` (+1 = the `ok3` line); loop `:1330→:1343` (+13 = 1 + 5 fixture lines + 7 for `probe_calls` and its blank line). Consistent with edits only at the three claimed places.
- **Reason strings are branch-unique in combination**: `"not the pinned"` also appears in the `--version` refusal (`collect.py:279`) but that message lacks `"review-safety build"`, which `:1348` also requires.
- **No fixture-name collisions**: `build-wrongsha` (`:495`), `stubbuild-wrongsha` (`:1331`), `col-wrongsha` / `col-wrongsha-work` (`:1344-1346`) are distinct; the four `stubbuild-*` directories are separate.
- **`--cli` override still last-wins**: `*args` follows the default `--cli stub` at `:1124`.
- **`stub_clawpatch.js` contains no `SCOUT_REDACTED_SECRET`** (grep: no match), so the `nosanitizer` fixture's `dist/` is genuinely token-free and `nomarker`'s refusal is on the marker, not the token.
- **Digest block**: DRAFT-NOTES `:8-31` differs from round 9 only in `PREREGISTRATION.md`, `validation/validate.py`, `README.md`; the three values equal those in `round-9-response.md`.
- **Ledger**: `docs/hypothesis-ledger.md:29` HD5 `preregistered`, Tests cell names rev 8 (correct while rev 10 is pending); `:33` dated wording-only line names the three added validations accurately and "round 10 pending".

## 7. Peer synthesis

Chair's independent pass: all three gaps closed with discriminating assertions; `:232` accurate; zero issues. **Peer** (`X2R10Peer`): six hypotheses (gaps closed, `:232` claim-for-claim, only the claimed edits, no stale text, nothing weakened, ledger consistent) all held; contributed the clone-implies-workspace proof (`collect.py:399-400`), the line-offset trace and the sha-only-comparison surviving mutant (Not raised). **Challenger** (`X2R10Challenger`): adversarial pass on `probe_calls` semantics (which invocations log, which directory, non-vacuous baseline, stub not writing `.calls`, no fixture collision, argparse last-wins) — all held; corrected the README anchors to `:24`/`:41`/`:45` (adopted above); zero issues. Neither peer changed the verdict; both sharpened the evidence and the Not-raised list. Disagreements: none.

## 8. Risks (recorded, not issues)

Unchanged from round 9 §8: later collection window than the first attempt (§10 `:218`, §11 `:233`); `.tools` build attested by marker + token, not a reproducible artifact digest; `review.go` cross-checked but not frozen (divergence refuses loudly). The round-9 risk "probe-before-preflight ordering is a code fact rather than a validated one" is now retired by `:1349`.

## 9. Execution outline (main agent, after this sign-off)

1. Recompute every digest with the `DRAFT-NOTES.md:32` commands and confirm equality with §10; `git diff` `validation/validate.py` against the round-9 commit to confirm only `:493`, `:1331-1350` and the `:497` check name changed (S1, S12, P untouched, rule 4); re-run `source .tools/activate && evaluation/x2-area-stability/run.sh validate` → 233 checks, 0 FAIL. Any mismatch voids this sign-off.
2. Ledger HD5: Tests cell → rev 10, signed in `reviews/round-10.md`; dated history line; state stays `preregistered`.
3. Confirm `.tools` holds the v3 build (`scripts/bootstrap-tools.sh` rebuilds on marker mismatch; `collect.py` refuses otherwise); `rm -rf /tmp/scout-x2`; `data/` empty.
4. Run the README `:28-34` chain under `nohup` in ABBA order; inspect only map success/failure and counts (§11).
5. `RESULTS.md` `Deviations`: Amendment 1, `data-rev8-unpatched/` and the reason, the later window; report `clawpatch_build_safety` and `clawpatch_cli_sha256` per (arm, repo) beside the probe lines (rule 6).

## 10. Sign-off

**ZERO ISSUES — signed off.** The innovation council signs off X2 preregistration revision 10 (`PREREGISTRATION.md`, signed-off rev 8 + Amendment 1 as revised after round 9) and its implementation, conditional on the main agent recomputing the digests below and finding each equal, and on `run.sh validate` reporting 233 checks with 0 FAIL. Frozen set and code under test (SHA-256, from `DRAFT-NOTES.md:8-31`):

```
ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355  PREREGISTRATION.md (rev 10)
9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799  GUIDE.md
e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3  common.py
89948f3c6d1b308b91a5ac5a8060088269270806546fe5c577dea307a4dfc83d  collect.py
f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e  replay/main.go
face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a  build_pairs.py
87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f  label.py
7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63  analysis.py
5157b15be6013c350b1dde006b0ee4833766d46981b131df756b1f838441c41b  run.sh
29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784  validation/validate.py
0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226  validation/stub_clawpatch.js
e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779  validation/credcheck/main.go
dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3  scripts/clawpatch-safe.patch (code under test, frozen from rev 9; equals the pinned sha)
99e3f6b3c670cd8c297b4ed498eafc735b7b5045aea8b97c7b55dc6343c21653  scripts/build-clawpatch.sh (code under test, frozen from rev 9; holds the pins)
```

Code under test and ported/drift-guarded sources at sign-off (not in the frozen set; a change refuses or fails validation loudly rather than silently):

```
8ed2e70815fdcc605d60402fb603cfb905d58457a5869b3795d59507a93cb484  internal/areas/areas.go
57344e025d5e2d0b7a10abf543b20b55f899602c1dd95d2b6a2b62d9b823dfbe  internal/clawpatch/records.go
bf10ddd58e0c2ecf08372d79505177b00c187955f3d64b3be9a3e8eb7c8617b8  internal/modelcli/modelcli.go
209639260283c373b7376bab352f580723a348bb2d6949bed7549c4ab03d910b  internal/clawpatch/client.go
615fcb0eebf047d9637d4d71cfe4e2faff7e730b840b8f8cbd4e032ed6384363  internal/clawpatch/exclude.go
2f4ca016d67d9bbfb646a73fe96f740e6982abb7f93112011d9d30efacd98f1e  internal/clawpatch/review.go (review-safety pins; cross-checked by common.review_safety_pins)
fadf1e447d9b0b9a32ba7db6ca9101d29a84acce5c575f37423995456d03d059  third_party/clawpatch/src/providers/claude.ts (unpatched submodule file)
994e8066be12c3051c4cb63b32e2144bcf4bd495c5d17a8033f94814bd7a6e92  third_party/clawpatch/src/providers/codex.ts (unpatched submodule file)
97413a275363a52e05c3b17d08449b211fdc26584bd60155c893a647720381ce  third_party/clawpatch/src/config.ts (unpatched submodule file)
13e04e61720fa0542d391a48f2d63db4e1973f29096114ba9120bddef84c4048  README.md (not frozen; recorded)
```

Rule 5 applies from this point: any change to a frozen file, however small, needs a new council round with zero issues before collection continues. Unavoidable runtime deviations go under `Deviations` in `RESULTS.md`.
