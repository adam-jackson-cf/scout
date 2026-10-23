# X2 — Response to council round 9 (Amendment 1)

Round 9 (`reviews/round-9.md`) reported one minor issue: `PREREGISTRATION.md:232` described validation coverage that `validation/validate.py` lacked. Prereg is now rev 10; no product code, collector, or analysis code changed.

## Issue 1 — validation-coverage sentence (option ii, then wording)

| Part | Council finding | Fix |
|---|---|---|
| (a) collector set | "the last three" ≠ the fixtures (`nomarker`, `oldmarker`, `nosanitizer`); wrong-sha unit-only | Added a collector fixture `wrongsha` (`stubbuild-wrongsha`: pinned version, sha `0`×64, patched `dist/`); the refusal loop now covers `nomarker`, `oldmarker`, `wrongsha`, `nosanitizer` (4 collector refusals; was 3) |
| (b) "before any probe" unasserted | only no workspace / `maps/` / `tool-versions-*` | New helper `probe_calls()` counts lines of the fake CLIs' `*.calls` logs (every model call, hence a probe, appends one; `--version` does not). Each refusal asserts the count is unchanged and that it was positive beforehand (the log works). |
| (c) "sanitizer in the entry file" untested | `"patched"` has the token in `prompt.js` only | Added unit case `"patched, sanitizer in the entry file"` (`cli.js` holds the token) to `U build-safety check …`, whose name now says entry file and non-entry file |

`PREREGISTRATION.md` §11 Amendment 1 "Validation" now says: the build-safety check (unit) on a patched build (sanitizer in the entry file `cli.js`, and in another dist file), no marker, an old marker, a wrong sha and a marker over an unpatched `dist/`; the collector refusing each of no marker, an old marker, a wrong sha and a marker over an unpatched `dist/` before any probe, clone or data file (asserted: nonzero exit, the build-safety reason, no workspace, no `maps/`, no `tool-versions-*`, no model call logged by the fake CLIs).

Other claims of validation in prereg §11/§2, README, DRAFT-NOTES and the amendment were checked against `validate.py` by name: USER/LOGNAME port (the claude-environment port check, exact key set, empty `USER` not copied, `SHELL` not copied); patched-source drift guards (`claude.ts`/`codex.ts`/`config.ts`, working tree vs `patched_sources`); pins/`review.go`/patch digest (`U … pins` check); `clawpatch_build_safety` recorded as `ok` (the collector tool-versions check); stub in a fixture build dir (`stubbuild`). No further discrepancy. README "about 1 min" corrected to "about 3 min" (measured 155 s).

## Other edits
Status line, Amendment 1 heading, "Validation (rev 10)" and the sign-off reference (round 10); README banner/`data/` row; DRAFT-NOTES status, 233 checks, digests; ledger HD5 wording-only history line. Optional Not-raised items (`:6`, `:228` qualifiers; negative `review_safety_pins` fixtures) were not taken.

## Evidence
`run.sh validate`: 233 checks (232 + the `wrongsha` collector refusal), all passing, 0 FAIL; `gofmt -l evaluation` empty; `go vet ./evaluation/x2-area-stability/...` clean. S1, S12, P sections unchanged (only the unit build-safety and collector-refusal sections of `validate.py` were edited).

## Digests (SHA-256)
- `PREREGISTRATION.md` (rev 10) `ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355`
- `validation/validate.py` `29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784`
- `README.md` `13e04e61720fa0542d391a48f2d63db4e1973f29096114ba9120bddef84c4048`
- All other files: unchanged from `DRAFT-NOTES.md` digest block.
