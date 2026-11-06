# ADR-004: Correct the Clawpatch review-safety patch pin

- **Status:** accepted
- **Design revision:** 2.4 (adopted 2026-10-23)
- **Decided:** 2026-10-23; **sources:** code defect and fix `06c9154`; no new R-/H- ledger source or user decision (`PROPOSAL.md:3`).

## Decision

- Pin the review-safety patch to `scout-review-sanitize-v3`, SHA-256 `dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3`.
- Keep Clawpatch at v0.7.3, commit `734f81bc1514b52a0627edb0c24ebe1a18b36105`.
- Record that v3 adds `USER` and `LOGNAME` to the host auth context copied by the patched `claudeEnv`.
- Correct the status-line dogfood release pin to `v0.1.4`.
- Correct `scout-design.base.md` (revision 2.3) §4's timeout symbol to `defaultMapTimeoutMS = "900000"` at `internal/clawpatch/client.go:23`.
- Leave every other substantive design statement unchanged.

## Informed by

- `PROPOSAL.md` fact 1 records that local builds staged inside Scout's worktree and silently applied no review-safety patch.
- Commit `06c9154` gives staging its own git repository and verifies application with `--check --reverse`.
- Bootstrap rebuilds when the installed marker differs from the pinned patch version and digest.
- `PROPOSAL.md` fact 2 records authentication failures on Claude Code 2.1.309 on macOS and a single-variable observation that adding `USER` fixed them.
- The same fact records a fastapi agent map exiting 0 with 11 agent features on the new build and exiting 1 on the old build.
- Round 1 excludes the host-specific failure claim from the design because the observation does not establish a universal failure mode.
- No experiment supports this revision; the council treats the pin as a code fact (`round-2.md`, §3).

## Changes

- Supersede ADR-000's Appendix 7 patch pin by replacing v2 with v3 in ADR-003's revision 2.3 snapshot, `scout-design.base.md` §4 and Appendix decision 7.
- Refresh the pin's code citations and cite `toolchain/versions.env:10` for pnpm 11.25.0.
- Refresh the status-line dogfood pin and `scout-design.base.md` (revision 2.3) §4 timeout symbol because they were stale in revision 2.3.
- Extend the evidence notation to certify that revision 2.4 does not rest on the historical CP07 study.
- Keep CP07 §7 unchanged as the historical v2 port record.
- Make no requirements change because `SCOUT_REBUILD_REQUIREMENTS.md`@9a3d19f5 does not state the patch pin.
- Later ADR-006 changes `docs/scout-design.md`@9a3d19f5 §4's refresh policy around the corrected timeout reference; it does not replace this patch-pin decision (revision 2.6 paragraph).

## Conditions

- Round 5 requires the main agent to recompute the recorded digests and reproduce `proposed.diff` byte for byte before adoption.
- Round 5 requires an adoption-time commit-scope check and re-reading affected code facts if cited files changed.
- Keep Clawpatch 0.8.x for a separate compatibility review.
- Round 1 records that the v3 fingerprint invalidates keyed caches once and triggers a local rebuild.
- Round 1 records that the existing `v0.1.4` release bundle still carries v2 and needs a later release for use with main's v3 provider-review check.

## Review

- Five rounds: 10 → 9 → 3 → 1 → 0 issues.
- Rounds 1–3 fix citations, stale code facts, overgeneralised evidence, scope wording and review bookkeeping.
- Round 5 resolves round 4's line-shift concern from fixer transcripts and records that the council's earlier insertion inference was wrong.
- Final sign-off: `round-5.md`; signed digests: `scout-design.proposed.md`: `5c16a807513d54667390278755b8daa504e8db5ae0733236f87938c51a9a90ae`; `proposed.diff`: `101f0af4fd1d48cdb70c2f1139071768bd77fa12be283c09f52648e1b00f8759`; `scout-design.base.md`: `75ef1e66521b1f129e872f651bd0f8bb9989e38867fb5335ec5ad43158fb3061`.
- Round 5 also records reviewed digests: `PROPOSAL.md`: `7b298dbc782b2da6ae8a6c30a8af2f49af563010cf7644c1d035c9da9e923d1a`; `DRAFT-NOTES.md`: `02d1fc068c600675a8ce4643ff6cb2a07a2f2cb0ca60cd06a63659d67ec2588a`.
- Adoption record: `docs/STATUS.md`@9a3d19f5, §1, revision 2.4 history entry; it records adoption on 2026-10-23 and byte-for-byte diff reproduction.
- No adoption response or adoption commit SHA is recorded in this folder.

## Records in this folder

- `ADR.md`: decision summary and record index.
- `PROPOSAL.md`: defect evidence, change set, exclusions and adoption procedure.
- `DRAFT-NOTES.md`: snapshot digests, line mapping and drafter judgement calls.
- `scout-design.base.md`: signed revision 2.3 baseline and rollback snapshot.
- `scout-design.proposed.md`: revision 2.4 snapshot signed off in round 5.
- `proposed.diff`: exact base-to-proposed design changes.
- `round-1.md`: ten issues and initial code-fact verification.
- `round-1-response.md`: fixes for round 1 and historical draft digests.
- `round-2.md`: nine remaining nits and verification of round 1 fixes.
- `round-2-response.md`: fixes for round 2 and consistency checks.
- `round-3.md`: three remaining nits and the line-shift verification condition.
- `round-3-response.md`: fixes for round 3 and corrected line-history account.
- `round-4.md`: one remaining review-record issue and conditional adoption checks.
- `round-4-response.md`: transcript-backed resolution and recorded digest, diff and HEAD checks.
- `round-5.md`: zero-issue council sign-off and adoption conditions.
