# ADR-012 round 12 response

Date: 2026-11-08. Responds to `round-12.md`; no sign-off claimed. Adopt the recommended two-bound window under R-064, recorded as the round-12 delegated decision in ADR.md’s Amendments section. Historical records remain unchanged.

| Issue | Fix | Discriminating acceptance obligations |
|---|---|---|
| ADR012-R12-01 | ADR population, REQ-27/78 and CI/CLI require the qualifying superseding run to be created after the cancelled run started and before its cancellation. If no qualifying run can be identified within both bounds, assess/gap the head, never exclude it; missing proof cannot certify readiness. Keep all other same-PR/pinned-lane/configuration/cause checks unchanged. | ADR acceptance, B4-S and CI add all three recommended cases with other exclusion conditions satisfied: (1) pinned=true, unedited configuration, no recorded actor/no provider cause; R1 cancelled before assessment and H2’s R2 created only afterwards: assess/gap H1. This fails an implementation omitting only the upper bound. (2) Pinned=true, R2 legitimately supersedes R1, then PR-added job cancels R2 before assessment with provider cause and no H3: assess/gap H2, never select older R1. (3) Same lower-bound case with manual cancel, no recorded actor and no provider cause. Cases 2–3 fail naive earlier-run selection for each evidence path. All existing fixtures remain. |

R-077 now links round 12. R-070 remains unchanged: no new platform claim or source is used; the change binds an implementation window and fail-closed fallback. Other summaries remain accurate without changes. No new ledger ID or mechanism.

No checks/builds/tests/formatters or experiments were run. Changes are staged, not committed. Requested hashes and staged diff statistics accompany the handoff for round 13 review.
