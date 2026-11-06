# ADR-001: Replace invalidated inertia evidence with E1b v2

- **Status:** partly superseded by ADR-002; amended by ADR-003
- **Design revision:** 2.1 (adopted 2026-10-21 [INFERENCE] from the revision date; the folder records no separate adoption date or commit)
- **Decided:** 2026-10-21; **sources:** R-033, R-034, R-035, H1–H4, HA1–HA3, user direction

## Decision

- Replace E1b v1 with council-signed E1b v2 as evidence for the inertia design.
- Keep structural measures descriptive and present them as structure to watch, not cost already being paid.
- Keep code quality as the primary family provisionally, pending the user's framing decision.
- Add scope-level `scout.burden_density@1` without treating its values as interchangeable with E1b's per-PR `Q`.
- Measure PR effort from author active hours in the API-sourced PR commit list and report created-to-merged queue time separately.
- Separate commit and PR populations, record bot-predicate provenance, and count and exclude bot, over-cap and incompletely linked PRs.
- Fix directory measurement scopes per map version rather than reuse E1b's PR-area clustering rule.
- Withdraw per-repository E1b re-estimation because Scout's change table lacks the study's predictors and controls.

## Informed by

- R-033 records the superseded and invalidated E1b v1 study.
- R-034 supplies the user's inertia-monitoring direction rather than experimental evidence for family priority.
- R-035 supplies E1b v2: 7,717 merged PRs in 12 repositories and 480 directory areas.
- T2 estimates +0.023 log-hours/SD, with 90% CI [+0.004, +0.041] and Holm p 0.165.
- T4 estimates +0.006 relative dispersion/SD, with 90% CI [−0.010, +0.023].
- Exploratory P1 estimates −0.15% relative MSE gain, with 95% CI [−0.69, +0.38].
- Failed validation leaves H4 undecided: G5 recall 0.422, G7 pair κ 0.514 and hidden precision 0.630.
- E1b's associations concern comparable-size, comparable-task PRs and do not identify causation or amplification-mediated cost.
- The single-commit floor affects 38% of 7,030 measured active-time rows, so the council requires reading T2 and T4 together.
- The round-1 response records active-hours reproduction for all 598 nest/webpack panel PRs with zero mismatches above 1e-9 hours.

## Changes

- Amend ADR-000's inertia framing in `scout-design.proposed.md` (signed revision 2.1) §§1, 2, 8.1 and 8.4 without reopening its Appendix decisions.
- Replace `scout-design.base.md` §8.5's invalidated v1 results and underpowered-study rationale with v2 estimates, equivalence conclusions and explicit limits.
- Amend `scout-design.proposed.md` (signed revision 2.1) §§4, 5.1, 8.2, 8.3 and 8.6 to define stable scopes, measurement populations, effort sources and provenance.
- Replace Phase 3's v1 outcome-comparison acceptance with active-hours reproduction joined on `repo` and `number` within 1e-9 hours.
- Amend X7 and A8 to require a separately powered, council-signed pooled study rather than reuse v1 sample-size assumptions.
- Narrow X8 to HA2 and correct the hypothesis ledger so X8 no longer claims to test HA1 or HA3.
- ADR-002 later removes the further-test list, X7, X8 and A8 when hypothesis studies move to inertia-bench (`docs/scout-design.md`@9a3d19f5, revision 2.2 history).
- ADR-003 later confirms the structure-to-watch framing and trusted-run PR ingestion, removing their provisional markers (`docs/scout-design.md`@9a3d19f5, revision 2.3 history).

## Conditions

- At sign-off, `docs/adrs/2026-10-21-e1b-v2/scout-design.proposed.md` §15 item 5 leaves primary-family priority and the provisional framing for the user to decide.
- Without PR metadata, I4, `batch_size` and I5's active-hours component remain `unmeasured` rather than zero.
- Keep `change_coupling.unexplained` exploratory until a dependency check validates it.
- Require a council-signed pooled test with at least 5% relative MSE reduction and a CI lower bound above zero before upgrading a named metric-outcome claim to predictive.
- Require consistency across all but one repository with at least 10 units and multiplicity control for that upgrade.
- Before adoption, confirm the signed proposal digest because neither final reviewer independently hashed it.

## Review

- Five rounds: 12 → 8 → 3 → 3 → 0 issues.
- Rounds 1–2 correct comparability promises, estimand limits, units, bot identity and PR linkage.
- Rounds 3–4 close population, completeness, provenance and hypothesis-ledger inconsistencies.
- Final sign-off: `round-5.md`; signed digests: `scout-design.proposed.md`: `7ff21a7fba70a5e3af7a735b0aa40f1ad654d17714df8c3549da6ebcf89b89c5`.
- The final round also adopts the HA1 Tests-cell correction and history entry in `docs/hypothesis-ledger.md`.
- Adoption record: no post-sign-off response appears in this folder; `round-4-response.md` records the final fixes and author-computed digest before sign-off.

## Records in this folder

- `ADR.md`: this decision summary.
- `scout-design.base.md`: pre-revision design snapshot.
- `scout-design.proposed.md`: signed revision 2.1 design snapshot.
- `proposed.diff`: word diff from the base to the final proposal.
- `round-1.md`: initial review with eight major and four minor issues.
- `round-1-response.md`: fixes the initial twelve issues and records active-hours reproduction.
- `round-2.md`: review with one major and seven minor issues.
- `round-2-response.md`: records units, provenance, linkage and missingness fixes.
- `round-3.md`: review with three minor bot and PR-linkage issues.
- `round-3-response.md`: records bot-rule alignment, commit-list checks and provenance values.
- `round-4.md`: review with three minor population and ledger issues.
- `round-4-response.md`: records final population and ledger fixes with the author-computed digest.
- `round-5.md`: zero-issue council sign-off of the exact proposal bytes.
