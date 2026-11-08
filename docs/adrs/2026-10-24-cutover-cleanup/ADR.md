# ADR-007: Consolidate state, baseline, model CLI and CI installation cutovers

- **Status:** partly superseded by ADR-008; amended by ADR-012 (13→14 migration, export manifest v2, trusted readiness publication)
- **Design revision:** 2.7 (adopted 2026-10-28, commit `9640370`)
- **Decided:** 2026-10-24; **sources:** R-044–R-047, user direction to address the coding-guidance audit findings

## Decision

- Create new SQLite stores directly at the one current schema, `user_version` 12.
- Keep only the v1 re-anchor migration from version 3 to 12, preserving `legacy_id` and recording ambiguous rows without merging them.
- Refuse versions 1, 2 and 4–11 with a re-baseline remedy after a read-only version check before the writable open.
- Require non-null `uninspected` in differential baselines and refuse missing lane health with `unavailable`, exit 3, instead of recomputing around it.
- Import records from older manifests without comparison identifiers, replacing or deleting `export_base` in the import transaction according to whether lane health is present.
- Export identifiers and lane health from the latest finalized run, else a recorded imported base with non-null `uninspected`, else empty identifiers and `uninspected: []`.
- Detect model CLIs in one order, `claude` then `codex`, while opted-in lanes use the policy provider and Scout no longer reads `CLAWPATCH_PROVIDER`.
- Remove the composite action and share the release tag's `scripts/install-release.sh` and `scripts/ci-toolchain.sh` between both reusable workflows.

## Informed by

- R-044 records the user's direction to address four design-bound findings from the coding-guidance audit of HEAD `1e07f84`.
- R-046 records that every tag v0.1.0–v0.1.5 writes schema 12.
- R-046 records 17 observed baseline manifests with `uninspected` and 69 anchored issues with no `legacyId` at the branch tip.
- The proposal retains version-3 support because the v1 nightly workflow bundled SQLite state to S3.
- The proposal finds that opted-in model lanes already require an explicit provider, so the review lane's separate default has no product caller.
- R-045 documents that `./` resolves in the runner workspace and that self-repository action references do not solve release-tag skew.
- R-047 documents that a read-only WAL open may create `-wal` and `-shm` side files.
- This revision cites no experiment and changes no hypothesis.

## Changes

- Amend ADR-000's v1 migration commitment to keep the version-3 re-anchor without the historical migration ladder.
- Amend ADR-002's model-CLI design and ADR-005's historical provider-variable account by removing Scout's `CLAWPATCH_PROVIDER` use.
- Replace ADR-003's legacy-baseline recompute clause in `scout-design.base.md` (revision 2.6) §7.2 and §7.3 with explicit refusal.
- Replace the composite-action commitment in `scout-design.base.md` (revision 2.6) §7.4 with release-tag scripts and an explicit workflow-to-tag interface.
- Update `scout-design.proposed.md` (signed revision 2.7) §3.3, §4, §10.2 and §11 Phase 1, add risk R17 and assumption A13, and record Appendix decision 22.
- Update `SCOUT_REBUILD_REQUIREMENTS.md` REQ-10, REQ-25, REQ-27, REQ-31 and its header through the signed `requirements.proposed.diff`.
- Preserve ADR-005's codex-only agent-map pin and ADR-006's onboarding/manual refresh schedule and Jaccard assumption.
- ADR-008 later changes the current schema to 13 and refuses version 12 with the re-baseline remedy (`docs/scout-design.md`@9a3d19f5, revision 2.8 header; `docs/STATUS.md`@9a3d19f5 §1).

## Conditions

- Re-baselining a refused local store loses unpublished local state.
- Assumption A13 leaves other repositories' baselines and surviving S3 bundles unobserved.
- Compare new-store and re-anchored schemas using a genuine v1 DDL fixture, not a rewound current store.
- Ship decisions 1–3 in one release.
- Ship the CI scripts in release N before merging workflows and caller pins that require tag N.
- Move each caller's workflow revision and `scout-version` together to versions containing both scripts.
- The final review requires matching digests, an exactly reproduced design diff and a clean requirements dry run before adoption.

## Review

- Seven rounds: 16 → 13 → 8 → 6 → 3 → 3 → 0 issues.
- The rounds resolve baseline contradictions, imported-identifier leakage, read-only WAL guarantees, migration-fixture validity, provider reporting and workflow-to-tag skew.
- Final sign-off: `round-7.md`; signed digests: `scout-design.base.md`: `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f`; requirements base `SCOUT_REBUILD_REQUIREMENTS.md` as reviewed in round 7: `bf6b8d2944abb757dea13c36ee85ebf4159ebf5d9535ce31827357362f8f649b`; `scout-design.proposed.md`: `00ffd3677acd7dbaec94d7c0a88be0b0d5a18841ceed5f7dbffa6df08d8bf005`; `proposed.diff`: `f3d055f04691375b714e04f26eb41bac081a6582e6259cf87b828b9fb7f72973`; `requirements.proposed.diff`: `f8d5cfc68c0d9602397923090ec619a6f43b6ad16749f105a28e5eed428d6ca8`; requirements patched result: `50b683f9c91fe6b3424187091306ea31754e4da675be8aa0b5b1034b754fcca9`; `PROPOSAL.md`: `a1fb5f7c19eb8ccd1815b513652f4dc411a10a1ca932a42e8a087a206cc25efa`; `DRAFT-NOTES.md`: `eacd00a46bfb3f3468f76bf8610770c159937a63172d28a604a121238bf69bac`.
- Adoption record: `round-7-response.md` records all three sign-off checks passing and adoption of the ledgers, design and requirements.
- `docs/STATUS.md`@9a3d19f5 §1 records adoption on 2026-10-28 at `9640370`.

## Records in this folder

- `ADR.md`: reader summary of the signed decision and its later schema amendment.
- `PROPOSAL.md`: four cutover decisions, alternatives, evidence, code inventories and adoption sequence.
- `DRAFT-NOTES.md`: changed-line inventory, artifact digests, judgement calls and unobserved evidence.
- `scout-design.base.md`: preserved revision-2.6 design snapshot.
- `scout-design.proposed.md`: signed revision-2.7 design snapshot.
- `proposed.diff`: signed design changes from the base snapshot.
- `requirements.proposed.diff`: signed requirements alignment patch.
- `round-1.md`: initial council review with 16 issues.
- `round-1-response.md`: fixes for all 16 issues and regenerated artifacts.
- `round-2.md`: second review with 13 issues.
- `round-2-response.md`: fixes for all 13 issues and SQLite source additions.
- `round-3.md`: third review with eight issues.
- `round-3-response.md`: fixes for all eight issues and regenerated artifacts.
- `round-4.md`: fourth review with six issues.
- `round-4-response.md`: fixes for all six issues and corrections to earlier records.
- `round-5.md`: fifth review with three issues.
- `round-5-response.md`: readiness wording, import-fixture and bookkeeping fixes.
- `round-6.md`: sixth review with three issues.
- `round-6-response.md`: complete readiness-capability wording, mapping-refresh fixture and artifact fixes.
- `round-7.md`: zero-issue council sign-off with digests and adoption conditions.
- `round-7-response.md`: passing pre-adoption checks and adoption record.
