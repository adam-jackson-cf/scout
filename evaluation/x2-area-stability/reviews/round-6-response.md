# X2 — response to council round 6

Round 6 (`reviews/round-6.md`): 3 minor issues, all text and validation only. Each is fixed exactly as the council's required fix; no code under `analysis.py`, `collect.py`, `common.py` changed, and no decision logic changed. Preregistration is now revision 7.

## Issue 1 — replication-status precedence (fixed)
- `PREREGISTRATION.md` §4 (`:142`) is rewritten as an ordered rule, the first clause that applies deciding: (i) `not applicable` when arm A did not adopt `t*`; (ii) else `not citable (arm A probe-inconsistent)`; (iii) else `not citable (arm B probe-inconsistent)` when B's gates pass, `t*` is `pass` in B and B is probe-inconsistent; (iv) else `replicated` iff arm A is probe-consistent, B's gates pass, `t*` is `pass` in B and B is probe-consistent, otherwise `not replicated`. This is the order `analysis.py:285-296` implements.
- `validation/validate.py:356` adds `rep(dB_no, bad_a_) == "not citable (arm A probe-inconsistent)"` to the "U replication citability" check.

## Issue 2 — positive citability path (fixed)
- `validation/validate.py:1112-1114`: the collector's probe-consistent `analyze` run now also asserts `res["arms"]["codex"]["decision"]["citable"] is True` and `"[citable for a design change: yes]" in A.render(res)`. A wrong key or arm name at `analysis.py:500` would now fail.
- §8 (`:197`) reworded: the S1 case is "a direct `replication` call with the claude probe flag set inconsistent on S1's decisions", which is what `validate.py:558` does; a new "Round-6 additions" bullet (`:198`) records the two additions.
- The optional planted S1 `tool-versions` fixture was not added (judgement call; the council marked it optional).

## Issue 3 — `fallback → not-supported` scope (fixed)
- §4 (`:144`): `fallback` and probe-consistent → `not-supported` **for the codex provider (arm A)**; "Arm B never moves the HD5 state; the ledger outcome and `RESULTS.md` state arm B's verdict, its own `t*` and its citability beside arm A's."
- The council's optional sentence (citability covers the served-model name only; a listed CLI-version change or a labeller-model change is reported beside the verdict and ledger outcome) is added to the same §4 paragraph rather than §10, so that no other frozen paragraph changes.
- `docs/hypothesis-ledger.md`: HD5 Tests cell (rev 7, rounds 1–6, codex-only `not-supported`, arm B never moves the state) and a dated "wording only, no state change" history line.

## Revision and consistency
- Prereg status line: revision 7, rounds 1–6, links `reviews/round-6-response.md`. `DRAFT-NOTES.md`: header, digests, choices 37–39. README has no revision or mapping text and is unchanged. §2/§10 carry no replication or ledger wording that changes.

## Verification
- `run.sh validate`: 222 checks, all passing (two existing checks extended, none added); `gofmt -l evaluation` empty.

## Digests after the fix (SHA-256)
```
02e59089646875940e0d08ebd9553e0835aea4808f9dbf9eb7cf24781d6f5e85  PREREGISTRATION.md (changed)
9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799  GUIDE.md
2cd742020c609c02fa09ad504812382493583a4b5f0c6fac831e86dac391741c  common.py
7e22b60360c4082d62400a0582f6b369cc9c89f556372065d30b79cba698a133  collect.py
f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e  replay/main.go
face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a  build_pairs.py
87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f  label.py
7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63  analysis.py
5157b15be6013c350b1dde006b0ee4833766d46981b131df756b1f838441c41b  run.sh
39200a38bdbb781f0204c6e05f24aec11addcd9e56982ec5578e474a8565b271  validation/validate.py (changed)
0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226  validation/stub_clawpatch.js
e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779  validation/credcheck/main.go
94ea650f24854869f8351ccf4a4fdf80b9bbaf1956c722e8c6494dbaf53a4744  README.md
```
