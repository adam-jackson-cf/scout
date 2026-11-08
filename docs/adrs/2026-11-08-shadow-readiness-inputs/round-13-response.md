# ADR-012 round 13 response

Date: 2026-11-08. Responds to `round-13.md`; not a sign-off. Adopt the recommended options under R-064. Historical review/response records remain unchanged.

| Issue | Fix | Acceptance obligation |
|---|---|---|
| ADR012-R13-01 | Replace the round-12 Amendments clause with the recommended accurate provenance: the adopted lower bound is the cancelled attempt’s job start, deliberately stricter than round 12’s recommended run-creation bound. A cancelled run that never started is assessed/gapped; absent proof never excludes a head. This records the fail-closed reason and departure instead of attributing it to the council. | No operative behaviour changes; the signed summary now accurately records the delegated choice. |
| ADR012-R13-02 | At all five rule locations (ADR population, REQ-27/78, CI/CLI), date the candidate by original run created_at at its first attempt, never rerun-attempt time. A rerun of an earlier-created run never qualifies; the window is after the cancelled attempt’s job start and before its cancellation. Explicitly assess/gap if original run creation cannot be distinguished from rerun-attempt time. Record round-13 decision in Amendments, rerun GITHUB_SHA/GITHUB_REF fact in R-070 using its existing source, and round-13 link in R-077. | Add fixture (4) in ADR acceptance, B4-S and CI, keeping every earlier fixture: pinned=true/unedited, terminal R1 on H1, started R2 on H2, maintainer reruns R1 and rerun cancels R2 before assessment, no H3. With provider cause and separately no actor/no provider cause, assess/gap H2. Dating R1 by its new attempt incorrectly excludes H2 and fails this fixture. |

No new mechanism, source ID or shipped behaviour. Other summaries remain accurate. No checks/builds/tests/formatters or experiments were run. Changes are staged, not committed; requested hashes and staged diff statistics accompany the handoff for round 14 review.
