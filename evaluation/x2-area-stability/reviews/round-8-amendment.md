# X2 amendment 1 (prereg rev 9) — request for council round 9

Status: **requested, not reviewed.** Rev 8 was signed off in `round-8.md`. Collection started under it on 2026-10-23T17:24Z and was stopped by Main
after codex/fastapi (7 maps, rc 0) and a claude/fastapi probe refusal. This amendment changes frozen files, so rule 5 requires a new council round with
zero issues before any collection resumes. Nothing was computed on predictor–outcome data (only map success/failure and counts were inspected, §11).
Validation after the change: `run.sh validate` **232 checks, all passing** (225 at rev 8 + 7 new); `go vet ./evaluation/x2-area-stability/...` clean; `gofmt -l evaluation` empty.

## Issues, changes, locations, validation

| # | Issue | Change | Location | Validation |
|---|---|---|---|---|
| 1 | The rev-8 codex/fastapi maps were produced by the **unpatched** local Clawpatch build (`build-clawpatch.sh` staged inside Scout's worktree, so `git apply` applied nothing; fixed in `06c9154`): agent-map prompts not redacted, feature metadata not sanitised; not "as Scout invokes" (§2). Evidence: its `tool-versions` has `clawpatch_cli_sha256 4eab263b…`, the unpatched build. | Maps excluded from every analysis, preserved untouched: `data/` moved whole to `data-rev8-unpatched/` (+ `x2-collect.log`, + a README note; no file edited). Fresh `data/`; all four (arm, repo) collections re-run in the preregistered ABBA order. | `PREREGISTRATION.md` status line, §2 (tool pins), §11 Amendment 1; `README.md` (layout, restart note); `data-rev8-unpatched/README.md` | The restart is a procedure, not code. Preflight (issue 3) refuses a non-patched build; the old build fails it (case "nomarker"/"nosanitizer"). |
| 2 | Patch is now `scout-review-sanitize-v3` (sha256 `dc867503…`): `claudeEnv` host context also copies `USER`, `LOGNAME`. Without them Claude Code 2.1.309/macOS fails every call (the claude/fastapi probe refusal). The port in `common.py` lacked them. | `common.claude_provider_env` copies `USER` and `LOGNAME` (non-empty only) after `CLAUDE_CONFIG_DIR`, as the patch's host branch does. | `common.py` (`claude_provider_env`, header comment); prereg §2 (probe environment list) | Unit port check extended (USER/LOGNAME copied, empty not copied, `SHELL` still not copied); probe allow-list in the collector check extended with the two names. |
| 3 | Drift guards and port checks read the **unpatched** submodule working tree, not the source Scout builds. | `common.patched_sources(rels)`: `git archive` of the gitlink revision (only files the patch touches + the requested ones), patch applied in a throw-away repository exactly as `build-clawpatch.sh` does (`--check`, apply, `--check --reverse`); the guards on `claude.ts`, `codex.ts`, `config.ts` use it. | `common.py`; `validation/validate.py` (unit section) | Existing drift guards still pass on the patched text; new guard: patched host branch has `copyEnv(env,"USER"); copyEnv(env,"LOGNAME")` right after `CLAUDE_CONFIG_DIR`, patched `claude.ts` ≠ submodule `claude.ts`, sanitizer token present in patched `prompt.ts` and absent from the submodule one. |
| 4 | Nothing stopped a collection on an unpatched build (recurrence). | `collect.py` preflight (before any probe, clone or data file, both arms) refuses unless `<build>/.scout-review-safety` (`<build>` = dirname(dirname(cli)), as Scout's `verifyReviewSafetyPatch`) equals `<version> <sha256>` AND a `*.js` under `<build>/dist` contains `SCOUT_REDACTED_SECRET`. Pins come from `scripts/build-clawpatch.sh` (what `bootstrap-tools.sh` compares), must equal `internal/clawpatch/review.go`, and the patch file's digest must equal the pinned sha (`common.review_safety_pins`, `common.build_safety_check`; no second copy of the pins). The result is recorded in tool-versions as `clawpatch_build_safety`. | `common.py`; `collect.py` (docstring, `preflight`, `tool_versions`) | New unit cases: pins agree with `review.go`/patch digest; check passes a patched build (sanitizer in the entry file or another dist file); refuses no marker, old marker (`-v2 f39d73da…`), right version with wrong sha, and a marker over an unpatched dist. New collector cases: the three refusal builds are refused before the probe/clone/data with the cause named; the normal run records `clawpatch_build_safety` ok with marker = expected. The stub Clawpatch now runs from a fixture build directory with the pinned marker. The real `.tools` build was checked to pass (not part of the suite). |
| 5 | Bookkeeping | DRAFT-NOTES digests, README, HD5 history line (wording only, amendment pending council). | `DRAFT-NOTES.md`, `README.md`, `docs/hypothesis-ledger.md` | n/a |

## Judgement calls for the council to confirm
1. **Both arms and both repos restart**, including codex/fastapi, though only its first run is tainted: the arm must be run on one build, and claude/axum/codex/axum had not started.
2. **No reuse or comparison of the discarded run**: it is preserved but never read; the new codex/fastapi run is a fresh draw (arm–time confound text extended in §11 Amendment 1).
3. **Marker + dist token** rather than marker alone: the marker alone is what the unpatched build lacked silently; the token proves the content independently.
4. **Pins read, not duplicated**; a disagreement between `build-clawpatch.sh` and `review.go`, or a patch file whose digest is not the pin, makes the check refuse (loud) rather than pass.
5. **Patched-source drift guards use `git archive` at the gitlink revision**, so a modified submodule working tree cannot influence them.
6. `validation/stub_clawpatch.js` is **unchanged**; only the way validate.py places it (a fixture build directory) changed.
7. `/tmp/scout-x2/codex/fastapi` (the first run's workspace) must be removed before the restart (existing preflight rule); documented in the README restart note, not automated.

## Digests at hand-over (SHA-256)
```
fecbe60772d1029d62159c0e4e5f23d02aa09c5a803eca3aa497aabba0664993  PREREGISTRATION.md (rev 9)
9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799  GUIDE.md (unchanged)
e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3  common.py
89948f3c6d1b308b91a5ac5a8060088269270806546fe5c577dea307a4dfc83d  collect.py
f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e  replay/main.go (unchanged)
face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a  build_pairs.py (unchanged)
87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f  label.py (unchanged)
7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63  analysis.py (unchanged)
5157b15be6013c350b1dde006b0ee4833766d46981b131df756b1f838441c41b  run.sh (unchanged)
218f4162dc040a1b22866138cf48e3d1c9d7e9a102fd184a28010fd614f8cd95  validation/validate.py
0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226  validation/stub_clawpatch.js (unchanged)
e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779  validation/credcheck/main.go (unchanged)
e3b7b7277e5e83dde5286884672ecaf75c359e476419956d21030266dc437ebb  README.md
dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3  scripts/clawpatch-safe.patch (code under test; equals the pin)
99e3f6b3c670cd8c297b4ed498eafc735b7b5045aea8b97c7b55dc6343c21653  scripts/build-clawpatch.sh (code under test; holds the pins)
2f4ca016d67d9bbfb646a73fe96f740e6982abb7f93112011d9d30efacd98f1e  internal/clawpatch/review.go (pins cross-checked)
```
Round-9 verdict: pending.
