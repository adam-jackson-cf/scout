# ADR-006: Run the agent map only at onboarding and on manual refresh

- **Status:** partly superseded by ADR-008
- **Design revision:** 2.6 (adopted 2026-10-24, commit `110388d`)
- **Decided:** 2026-10-24; **sources:** R-040, R-041, R-042, R-043, HD5, HD5b; user direction of 2026-10-24

## Decision

- Run the Clawpatch agent map only at onboarding and on a maintainer's manual refresh.
- Attempt onboarding on each default-branch push and nightly run until one agent map completes and is reconciled.
- Leave onboarding open after a `skipped`, `blocked` or `incomplete` attempt.
- Require `scout map --refresh --allow-egress` locally, or the trusted workflow's `refresh-map` dispatch with `authorize-egress`, for a manual map.
- Remove refresh-on-merge, structural-change triggers, debounce, nightly remapping and staleness-triggered remapping after onboarding.
- Keep `mapping.areas.match_jaccard` at 0.5 as an unsupported assumption for onboarding and manual maps.
- Keep the offline heuristic map canonical and preserve issue identity and time-series keys across agent maps.
- Never run the agent map on PR runs.

## Informed by

- R-042 records X2b's citable `fallback` result for codex `gpt-6-luna` at `xhigh`, run at commit `f2baaa8` under preregistration revision 4.
- All 13 thresholds on the 0.30–0.90 grid failed C1–C3.
- Fastapi's mean per-refresh survival was 0.5721 at 0.30 and 0.4154 at 0.50, below C1's 0.8 minimum.
- Axum's mean per-refresh survival was 0.9102 at 0.30 and 0.8800 at 0.50.
- X2b found zero `different_concern` labels among 521 labelled pairs.
- X2b's pass set was empty, its contiguity flag was vacuously `true`, and it produced no `t*` or survival margin.
- Unlabelled share at 0.30 was 0.0 in both repositories.
- The study used one run of seven maps per repository and measured nothing below 0.30.
- R-041's X2 result at 0.30 used another model, probe-served `gpt-6.1-sol`, and did not transfer to the production mapper.
- R-043 records the user's direction to adopt the preregistered fallback while retaining the 0.5 default.
- HD5b remained `not-supported`, not `refuted`; HD5 remained unchanged.

## Changes

- Supersede ADR-000's debounced merge-trigger clause in `docs/scout-design.md`@9a3d19f5 Appendix decision 1 while retaining its areas layer and PR exclusion.
- Close ADR-005's pending-X2b clause in `docs/scout-design.md`@9a3d19f5 Appendix decision 20 without changing its mapper model pin.
- Add Appendix decision 21 and amend `scout-design.proposed.md` (signed revision 2.6) §§1, 4, 7.1, 7.4, 10.2–10.3, 11 Phase 4, 12, 13 and 14.
- Update `SCOUT_REBUILD_REQUIREMENTS.md` REQ-13 and its authoritative-design header through the signed `requirements.proposed.diff`.
- Replace Phase 4's debounced-burst acceptance with one onboarding map, no merge/PR/nightly remaps, and one manual map.
- Add risk R16 for stale areas between agent maps without adding an automatic alarm.
- ADR-008 later reopens `docs/scout-design.md`@9a3d19f5 Appendix decision 21 only for pending-area association and adds a manual-refresh recommendation (revision 2.8 paragraph and Appendix decision 23).

## Conditions

- Final sign-off required matching digests, a byte-identical design diff and a clean requirements patch dry-run before adoption.
- `round-5-response.md` records all three conditions as satisfied.
- Adopt the design and requirements before committing the companion code change.
- Ship the changed workflow and a compatible released binary together, with both dogfood and documentation pins updated.
- Reject policies that still set `mapping.refresh.debounce` rather than accepting an ignored compatibility key.
- Re-run Phase 4 under its amended acceptance before claiming that acceptance passed.
- `docs/STATUS.md`@9a3d19f5 records companion code commit `ead1a3f` and says the amended Phase 4 scenario remains unrun.

## Review

- Five rounds: 12 → 10 → 4 → 4 → 0 issues.
- Reviews resolved onboarding semantics, preregistered disclosures, adoption ordering, cutover coverage and drifting code citations.
- Final sign-off: `round-5.md`; signed digests: `scout-design.proposed.md`: `be49efd5c5e0aed59f79dd416cc50312a3dc0b85b3f5a6397742e2de29771a6f`; `proposed.diff`: `fc69e6bf8e49f90123bd7203f07fa9bfa714f6e61895c7168a9e2ebd278c2c8d`; `requirements.proposed.diff`: `35df910ef6d512a6a96a8372394f0fa00ebfd25e7fea7e6023de467b7ccfc6a4`.
- Signed base digests: `scout-design.base.md`: `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9`; requirements base `SCOUT_REBUILD_REQUIREMENTS.md` as reviewed in round 5: `9efcea1b0248d43d0ab831f0924f055f16bb39a3e008365d1e64c7d68a6ece2d`.
- Signed supporting digests: `PROPOSAL.md`: `45ff11221040f65bdb9b7cd33b2d77f26f9725ad469028dd0659235df8c520d6`; `DRAFT-NOTES.md`: `17930c799b5c25d9010a5093fecf99c7f578f658f8ad705dc88f487f8c21192b`.
- Signed requirements patched-result digest: `bf6b8d2944abb757dea13c36ee85ebf4159ebf5d9535ce31827357362f8f649b`.
- Adoption record: `round-5-response.md` records the checks, design replacement, requirements patch and ledger updates.

## Records in this folder

- `ADR.md`: decision summary and later amendment pointer.
- `PROPOSAL.md`: evidence, change set, judgement calls and companion adoption work.
- `DRAFT-NOTES.md`: draft digests, changed-line inventory and adoption checks.
- `scout-design.base.md`: frozen revision 2.5 design before this decision.
- `scout-design.proposed.md`: signed revision 2.6 design.
- `proposed.diff`: revision 2.5-to-2.6 design patch.
- `requirements.proposed.diff`: REQ-13 and header patch for `SCOUT_REBUILD_REQUIREMENTS.md` as reviewed in round 5.
- `round-1.md`: twelve issues, including onboarding consistency and missing disclosures.
- `round-1-response.md`: fixes submitted for the twelve first-round issues.
- `round-2.md`: ten issues concerning semantics, requirements adoption and cutover details.
- `round-2-response.md`: fixes submitted for the ten second-round issues.
- `round-3.md`: four nits concerning moved code citations and adoption bookkeeping.
- `round-3-response.md`: fixes submitted for the four third-round nits.
- `round-4.md`: four remaining proposal and draft-note nits.
- `round-4-response.md`: fixes submitted for the four fourth-round nits.
- `round-5.md`: zero-issue conditional sign-off and signed digests.
- `round-5-response.md`: satisfied sign-off conditions and adoption record.
