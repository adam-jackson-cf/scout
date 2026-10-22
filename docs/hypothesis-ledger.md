# Hypothesis ledger

Every hypothesis that shapes Scout, its current state, the experiments that test it and their outcome.
Sources are research-ledger IDs (`docs/research-ledger.md`).

States: `proposed` → `preregistered` → `under-test` → `supported` | `not-supported` | `refuted` |
`inconclusive`; any state → `revised` (link the new ID) or `withdrawn`. Only a signed-off experiment
(AGENTS.md experiment rules) moves a hypothesis to `supported`, `not-supported` or `refuted`.

## Register

| ID | Hypothesis | Sources | State | Tests | Outcome | Updated |
|---|---|---|---|---|---|---|
| H1 | Quality degradation in an area raises the effort and complexity of later changes there | R-005, R-010, R-011, R-012, R-018, R-034, R-035 | not-supported | E1b v1 (invalid); E1b v2 T2 | v2: equivalent — +0.023 log-h/SD, 90% CI [0.004, 0.041] inside SESOI ±0.10; Holm p 0.165 (merged PRs, author active time, 12 repos) | 2026-10-21 |
| H2 | Degradation makes change effort less predictable (wider variance, longer tail) | R-010, R-034, R-035 | not-supported | E1b v1 (invalid); E1b v2 T4 | v2: equivalent — +0.006/SD, 90% CI [−0.010, 0.023] inside SESOI ±0.10 | 2026-10-21 |
| H3 | Degradation raises the chance that a change introduces a bug | R-010, R-015, R-034 | proposed | — (E1 tested fix-proneness, not introduction) | — | 2026-10-20 |
| H4 | Hidden coupling (co-change without a static dependency) causes incomplete changes | R-015, R-017, R-022, R-035 | inconclusive | E1b v1 (invalid); E1b v2 T5 | v2: not decided — gates G5 (Jev recall 0.42) and G7 (proxy κ 0.51, precision 0.63) failed; exploratory −0.024 [−0.151, 0.102] | 2026-10-21 |
| H5 | Concentrated ownership slows change and raises risk | R-016, R-017 | proposed | — | — | 2026-10-20 |
| H6 | Specific code patterns cost runtime performance or energy | R-026 | proposed | — | — | 2026-10-20 |
| H7 | Larger change batches reduce delivery stability | R-025 | proposed | — | — | 2026-10-20 |
| HA1 | Agents degrade code faster than humans | R-006, R-024, R-025, R-034 | proposed | — (X8 tests HA2 only; HA1 needs a human comparison) | Benchmark evidence only (R-006) | 2026-10-21 |
| HA2 | Degraded code makes agents less reliable | R-027, R-034 | proposed | X8 | One refactoring study, partial (R-027) | 2026-10-20 |
| HA3 | Agent velocity multiplies the interest paid on debt | R-028, R-034 | proposed | — | — | 2026-10-20 |
| HA4 | Agents misapply patterns and are misled by contradictory context | R-034 | proposed | — | No study found | 2026-10-20 |
| HQ1 | Quality levels predict where bug fixes land next, beyond size and history | R-005 | not-supported | E1 inertia-v1 (pre-dates experiment rules) | +1.18% [−0.62, +3.24]; CC>10 count alone +1.91% [0.39, 3.68] | 2026-10-17 |
| HD2 | A typed decision model can triage analyzer findings well enough to route them | R-008, R-009 | inconclusive | E2 jev-triage-v1 (pre-dates experiment rules) | 17–21/22 correct; n too small | 2026-10-17 |
| HD3 | Always running Clawpatch's agent mapper gives stable feature IDs for issue anchoring | R-007, R-034 | refuted | E3 map-stability-v1; CP07 | Agent IDs 5–34% survival on 0.4.0 and 0.7.3 | 2026-10-19 |
| HD4 | Clawpatch 0.7.x is compatible with Scout without breaking feature identity | R-007 | supported | CP07 clawpatch-0.7-compat (pre-dates experiment rules) | 100% heuristic IDs retained on 5 repos | 2026-10-19 |

## History

- 2026-10-21 — HA1: Tests corrected from X8 to none. X8 (healthy vs degraded agent continuations) tests HA2 only; HA1 needs a human comparison (design review `docs/design-reviews/2026-10-21-e1b-v2/round-4.md`).
- 2026-10-21 — H1, H2 → `not-supported`: E1b v2 (R-035) returned `equivalent` for T2 and T4 with adequate power; effects ≥ 0.10 log-h per SD excluded. A small positive T2 association (≈ 2 %/SD) remains compatible. H4 → `inconclusive`: T5 not decided because its measurement gates failed. Run used Amendment 1 (annotator C truncation fix, rounds 11–14).
- 2026-10-21 — H1, H2, H4: E1b v2 preregistration and scripts signed off by the innovation council with zero issues (`evaluation/change-cost-v2/reviews/round-10.md`); state → `preregistered`. Run blocked at annotation (step 11) until `OPENROUTER_API_KEY` is provided.
- 2026-10-20 — H1, H2, H4: E1b v1 invalidated by method critique (R-033); state stays `under-test` pending E1b v2.
- 2026-10-20 — Ledger created; E1, E2, E3 and CP07 outcomes back-filled. Those studies pre-date the experiment rules.
