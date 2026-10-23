# X2 — response to council round 7

Round 7 (`reviews/round-7.md`): 2 minor issues, text and validation only. Each is fixed as the council's required fix; no code under `analysis.py`, `collect.py`, `common.py` changed, and no decision logic changed. Preregistration is now revision 8.

## Issue 1 — `analyze → replication` wiring never exercised with real probe files (fixed)
- `validation/validate.py` (after the S1 citability check): `plant_versions` writes probe-consistent `tool-versions-{codex,claude}-{fastapi,axum}.json` (reusing the `tvf`/`ok_p` shapes; `workspace_config_presence` all false) into a copy of S1's data directory. S1's existing unplanted assertions stay on `r1`.
  - (a) `A.analyze` on the planted copy: both `probe_consistency` flags true, both `decision.citable` true, `replication.status == "replicated"`, `arm_A_citable` and `arm_B_citable` true, `arm_B_verdict == arms.claude.decision.verdict == "adopt"`, `arm_B_threshold_pct == 75`.
  - (c, recommended, done) the same data with the claude axum after-probe rolled (`n1` → `n2`): `arms.claude.decision.citable` false, `arms.codex.decision.citable` true, status `not citable (arm B probe-inconsistent)`, `arm_A_citable` true, `arm_B_citable` false. This also fails a `:500` `.get("codex")` mutant.
  - (b) new scenario S1b (only the claude arm's real pairs labelled `different_concern`, files planted, both arms consistent): arm A `adopt` 75, claude verdict differs from arm A's, status `not replicated`, `arm_B_verdict == arms.claude.decision.verdict`, `arm_B_threshold_pct` equals the claude decision's, `arm_B_citable` true. This fails the `dec, dec` mutant at `:504`.
- §8 gains a "Round-7 additions" bullet describing exactly these fixtures. DRAFT-NOTES choices 38 and 40 describe the fixture that exists.

## Issue 2 — §4 citability sentence (fixed)
- `PREREGISTRATION.md:144`: "Citability covers the served-model name only: …" replaced with the council's wording: "Citability is the §2 probe flag (served model identical across all probes of the arm, every probe taken without error, every `tool-versions` file present, no workspace-level configuration file present); a listed CLI-version change or a labeller-model change is reported beside the verdict and the ledger outcome, not folded into the flag." Consistent with §2 `:39`, §10 `:216` and `analysis.py:448-450`. The optional scoping of `:139` was not done.
- DRAFT-NOTES 39 and 41 aligned.

## Revision and consistency
- Prereg status line: revision 8, rounds 1–7, links this file. `docs/hypothesis-ledger.md` HD5 row (rev 8, rounds 1–7) and a dated wording-only history line (state stays `proposed`). README unchanged.

## Verification
- `run.sh validate`: 225 checks, all passing (three added); `gofmt -l evaluation` empty.

## Digests after the fix (SHA-256)
```
a39513d6eabddce467cca26cf0ea5085e15ab3e615e34a571f9603e5c2a3c426  PREREGISTRATION.md (changed)
9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799  GUIDE.md
2cd742020c609c02fa09ad504812382493583a4b5f0c6fac831e86dac391741c  common.py
7e22b60360c4082d62400a0582f6b369cc9c89f556372065d30b79cba698a133  collect.py
f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e  replay/main.go
face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a  build_pairs.py
87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f  label.py
7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63  analysis.py
5157b15be6013c350b1dde006b0ee4833766d46981b131df756b1f838441c41b  run.sh
beb32dc2527aa01abda63a0650c93390dc38e14e5dd24f56f9452dc465d286f6  validation/validate.py (changed)
0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226  validation/stub_clawpatch.js
e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779  validation/credcheck/main.go
94ea650f24854869f8351ccf4a4fdf80b9bbaf1956c722e8c6494dbaf53a4744  README.md
```
