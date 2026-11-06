# Design review proposal — Clawpatch review-safety patch pin v2 → v3 (revision 2.4)

Status: **round-5 draft (round-4 issue fixed, `round-4-response.md`; round 3 in `round-3-response.md`; round 2 in `round-2-response.md`), awaiting innovation-council review** (`AGENTS.md` design rules 1–2). Base: `docs/scout-design.md` revision 2.3, SHA-256 `75ef1e66521b1f129e872f651bd0f8bb9989e38867fb5335ec5ad43158fb3061` (signed off in `docs/design-reviews/2026-10-23-decisions/round-7.md`). Driver: a defect-and-fix pair in code on `main`, not a user decision and not an experiment. No experiment is cited; no ledger source is added. Proposed file digest: `DRAFT-NOTES.md`.

## Facts (verified 2026-10-23, commit `06c9154`, pushed)

1. `scripts/build-clawpatch.sh` staged Clawpatch source under `.tools` inside Scout's git worktree, so `git -C <stage> apply` applied nothing and `--check` passed vacuously. Local `.tools` Clawpatch builds therefore never carried the review-safety patch (prompt/output redaction via `sanitizeSensitiveText` in `prompt.ts`, `agent-mapper.ts`, `review.ts`; agent-map output metadata sanitisation; `.scout` skip in `mappers/shared.ts`). Release bundles build in `$RUNNER_TEMP` and were patched (v0.1.4 bundle: `SCOUT_REDACTED_SECRET` present, marker `scout-review-sanitize-v2 f39d73da…`). Fixed: the staging directory is its own git repo, apply is verified with `--check --reverse`, and bootstrap rebuilds when the marker differs from the pins.
2. The patch is bumped to `scout-review-sanitize-v3`, sha256 `dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3`: `claudeEnv` host auth context now also copies `USER` and `LOGNAME`. Without them Claude Code 2.1.309 on macOS failed every call (observed 2026-10-23; not in the design) (`Failed to authenticate: OAuth session expired and could not be refreshed`); a single-variable test showed `+USER` alone fixes it; a real claude agent map on fastapi exits 0 with 11 agent features on the new build and exits 1 on the old.

Where the pin lives in code (re-read 2026-10-23): `internal/clawpatch/review.go:20-21` (`reviewSafetyPatchVersion`, `reviewSafetyPatchSHA256`; `:488` source string `clawpatch@0.7.3+scout-review-sanitize-v3`), `scripts/build-clawpatch.sh:18-19`, `.github/workflows/release.yml:89` (`safetyPatchSha256`).

## Change set

| # | Where (rev 2.3) | Was | Now | Evidence |
|---|---|---|---|---|
| P1 | Header line 3 | "…revision 2.2 (2026-10-23), **revision 2.3 (2026-10-23)**." | "…revision 2.3 (2026-10-23), **revision 2.4 (2026-10-23)**." | revision bookkeeping |
| P2 | New paragraph after the Revision 2.3 paragraph (proposed line 9) | — | **Revision 2.4** paragraph: pin correction (v3, `dc867503…e6ce3`; §4, Appendix 7); v3 adds `USER`/`LOGNAME`; also refreshes two drifted code facts (P5, P6); apart from revision bookkeeping (status line, evidence notation), no other statement changes | facts 1–2 |
| P3 | §4 Pin bullet (line 104 → proposed 106) | patch `scout-review-sanitize-v2` sha256 `f39d73da…a865` (`review.go:19-20`, `scripts/build-clawpatch.sh:5-9`, pnpm 11.25.0, `client.go:19-20` cited on the same line) | patch `scout-review-sanitize-v3` sha256 `dc867503…e6ce3` (`internal/clawpatch/review.go:20-21`, `scripts/build-clawpatch.sh:18-19`, `.github/workflows/release.yml:89`, pnpm 11.25.0 `toolchain/versions.env:10`; Pin/Version cite corrected to `internal/clawpatch/client.go:20-21`) | `review.go:20-21`; `build-clawpatch.sh:18-19`; `release.yml:89` (read 2026-10-23). The old line cites (`review.go:19-20`, `build-clawpatch.sh:5-9`) no longer point at the pin |
| P4 | Appendix item 7 (line 454 → proposed 456) | ported patch `scout-review-sanitize-v2` (CP07 §7) | ported patch `scout-review-sanitize-v3` (CP07 §7; v2 → v3 by revision 2.4) | as P3. CP07 §7 is the original port record; it is left unchanged |
| P5 | Status line 3 (stale in revision 2.3, not from `06c9154`) | "release `v0.1.2` is published and pinned by the dogfood caller" | "release `v0.1.2` exists for CI to install and the dogfood caller pins `v0.1.4`, `docs/STATUS.md` §1" | `.github/workflows/scout-nightly.yml` `scout-version: v0.1.4` at `:40,50` on `main` (`git show HEAD:`; `:40,49` in the working tree, which carries uncommitted edits); `docs/STATUS.md:72,94` |
| P6 | §4 Map refresh (base 97) | `defaultMapCodexTimeoutMS = "900000"`, `client.go:22` | `defaultMapTimeoutMS = "900000"`, `internal/clawpatch/client.go:23` | no `defaultMapCodexTimeoutMS` under `internal/`; constant at `client.go:23` |
| P7 | Evidence notation line (proposed 11) | "revision 2.1, 2.2 or 2.3" | "revision 2.1, 2.2, 2.3 or 2.4" | rev 2.4 certified as not resting on CP07 |

The new abbreviation is `first 8…last 5`: `dc867503…e6ce3` (full digest above); the rev 2.3 line's `f39d73da…a865` used last 4.

## Explicitly not changed

- **Host-auth environment.** The design does not enumerate the host auth context's environment (§4 and §6 say only "host auth context", `CLAWPATCH_CLAUDE_AUTH_CONTEXT=host`), so no statement is false and none is added. The host-only block of `claudeEnv` sets `HOME` and copies `USERPROFILE`, `CLAUDE_CONFIG_DIR` and, since v3, `USER`, `LOGNAME` (`scripts/clawpatch-safe.patch:332-336`); upstream additionally copies `CLAUDE_CODE_OAUTH_TOKEN` in host context when auth is included (`third_party/clawpatch/src/providers/claude.ts:258,268-270`); it is recorded here and in the Revision 2.4 paragraph (`USER`/`LOGNAME`) only.
- §10.2's `scripts/build-clawpatch.sh` + patch "Keep" verdict, R9 ("`build-clawpatch.sh` verifies patch"; now true for local builds too), §6, §7.4, Phase 6 "Clawpatch 0.7.3 bundle", every other section and Appendix item.
- No design claim about the local-build defect is added: the design never said local builds were patched beyond "ported patch".

## Companion change

None. `SCOUT_REBUILD_REQUIREMENTS.md` does not mention the patch pin (searched `sanitize`, `f39d73da`, `HOME`, `USERPROFILE`: no matches), so there is no `requirements.proposed.diff`.

## Procedure

Per `AGENTS.md` design rules 1–2: round N by the innovation council, recorded as `round-<N>.md` / `round-<N>-response.md` in this directory, iterating to zero issues; the final round names the proposed file's SHA-256 (`DRAFT-NOTES.md`). Only then replace `docs/scout-design.md` with `scout-design.proposed.md` (base digest must still be `75ef1e66…`, otherwise rebase and re-review).
