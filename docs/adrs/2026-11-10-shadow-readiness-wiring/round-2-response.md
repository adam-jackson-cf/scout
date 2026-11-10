# Round 2 response (not a sign-off)

Round-2.md is preserved verbatim. The integration peer completed after the lead's record; its complete independent findings are preserved in round-2-integration.md. Both sets are addressed below; final approval remains for the next council.

| Finding | Resolution |
|---|---|
| R2-01 | Signed machine contract now has literal operation conditions, each existing guard, pure validate/normalization helper, input grammar/event regex/control bounds, scoped exits, token-env map, protected-job condition/60m ceiling, tag rule, artifact regex syntax, exact approved REQ row SHA/extraction and unchanged row hashes. W2/W6 define token-free askpass and fail-closed restore. |
| R2-02 | Explicit !cancelled() assessment guard; inventory moved after upload/enforcement so it cannot starve them. |
| R2-03 | Protected-default reachable git commit/path/content-hash receipt; schema/bindings/API checks and explicit operational attestations. Empty-history genesis=null is allowed only for initial normal S3 genesis; existing lineage never resets. |
| R2-04 | No unpublished-action rerun authorization in v1; run_attempt>1 is not_published/2 with fresh dispatch. Preserve terminal deduplication. Append R-104 correcting R-102; historical row unchanged. |
| R2-05 | No-marker reconcile is neutral not_started/2/no writes. Explicit documented shadow-vs-assessment fail-on-incomplete scope in W2, machine exit map and CI table/notes. |
| R2-06 | No pretending an advancing API clock exists. Integration F1/F3 exposes deeper identity/record prerequisites: v1 now authorizes no reassessment/attempt collection at all; later source/reassessment ADR must solve all of them, not just time. No inert 2h/40m settings remain. |
| R2-07 | Never tokens in argv; S6 includes other process argv/environment and disk on Linux/macOS, not merely scrubbed child env. |
| R2-08 | Select highest exactly named round-N.md, one matching signoff block there; higher unsigned round refuses. Lower historical signed rounds remain immutable records (rather than deleting them to satisfy a global-single-block rule); no non-round block may authorize. |
| Integration F1/F3/F4/F5/F7 | Take the offered conservative option: remove multi-head reassessment from W5. V1 inventory is diagnostic; exact durable types are start, existing-event triage, immutable action refusals/idempotence/publications only. No PRRunFact/new event/link/gap/interval/frontier; no attempt identity collision, invented B or generated veto backlog. Horizon is run creation at/after marker; later full source/identity/B/deadline/lag/capacity ADR required. |
| Integration F2 | Last-step placement, 2m timeout, 100-request cap and stop below 200 provider remaining. No API resource cap is presented as measurement completeness or a cross-run reservation. |
| Integration F6 | Later complete source continues the same lineage and creates an interval origin/first frontier, never new lineage genesis. |
| Integration F8 | Explicit token-free workflow askpass helper (no new binary dependency for baseline Git); distinguish ls-remote 0/2 from auth/network errors. Old shadow binary fails early normalization before assessment/publication; compatible tag coupling required. |
| Integration F9/F10/F11 | Receipt includes last_intact_v2_commit/migration_result; whitelist includes cache/prepare-auth/normalization; CI table and CLI stale paragraph corrected; design/trackers/reference summaries and R-100–R-108 sources aligned. |

The actual integration review is design/code analysis, not an experiment or live proof. This decision is a complete conservative control-plane wiring contract, **not full certifying collection**, and keeps that explicit prerequisite and ADR-012 acceptance open. No Go or workflow YAML changes are made here.
