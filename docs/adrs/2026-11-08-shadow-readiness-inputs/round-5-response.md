# ADR-012 round-5 response

Date: 2026-11-08. Response to `round-5.md` (full restored 392-line council record). These are proposed fixes for council review, not a sign-off or implementation validation. ADR-012 remains decided by the agent under owner delegation (R-064); owner review pending.

## Delegated choices

Use every recommended option: keep the narrowed invariants-only ADR; send requirement-changing wiring to a separate adopted, zero-issue wiring ADR; restore the reviewed protected-base population; seal terminal-complete provider-clock frontiers; retain the exact-version false-positive veto across gaps; use the three-term publishing-CAS timestamp formula; retain and explicitly declare the inspected-only Days origin. No trigger, transport, cadence, workflow, policy key, command or sandbox mechanism is selected. Wiring wholly within existing requirements needs no additional ADR. No product code or experiment changes.

## Issue responses

| Issue | Fix and implementation acceptance invariant |
|---|---|
| ADR012-R5-01 (major) | ADR scope, REQ-161, CI reference, STATUS gap and B4-S now require a separate adopted wiring ADR with zero-issue council review before implementation merges if any trigger/dispatch input/workflow/permission/policy key/release artifact or REQ-160–163/168 text changes. B4-S depends on adopted ADR-012 and that wiring ADR where required. Acceptance: a requirement-changing proposal cannot merge without it; unchanged-requirement wiring needs none. |
| ADR012-R5-02 (major) | ADR counting item 1, REQ-27/78 and CI reference define each eligible guardian-run/would-run PR event targeting the protected default branch, including forks/bots, with no content/label/author/outcome selector. Reassessments add attempts only to existing observations. Provider enumeration determines completeness; omitted events create gaps and selector changes start intervals. Non-default/unreachable base is disclosed not_assessed/non_default_base. Acceptance fixtures: omit a failing eligible PR and certification fails; a non-default-base PR is not_assessed, never evidence. |
| ADR012-R5-03 (major) | ADR item 4, REQ-78 and CLI reference fix started_at/intervals/C to the provider clock. Advancing C requires all eligible attempts started≤C terminal and published in that state or in a recorded gap. Later states retain exactly the same pre-C facts/links; late additions are refused/logged. Acceptance: an attempt/link arriving after sealed C1 with started_at≤C1 cannot change the as_of=C1 dossier. Unreliable provider start time cannot certify coverage. |
| ADR012-R5-04 (major) | ADR items 5/ShadowReady clause, REQ-22/78 and policy/CLI references apply zero-not_warranted to all persisted CI events of the repository/exact cohort linked to assessed cohort attempts started≤as_of, using triage at as_of regardless of interval. Other window counts remain interval-scoped; report prior-interval not_warranted. Acceptance: a known not_warranted in interval 1 still vetoes that exact version after a gap and a clean qualifying interval 2. |
| ADR012-R5-05 (major) | ADR triage, REQ-78 and policy/CLI references assign recorded_at=max(writer UTC now, parent C+1ms, parent maximum recorded_at+1ms), recomputed on each CAS retry. Publication order determines the latest authorized verdict while preserving cutoff safety. Report in-window events with a superseded authorized not_warranted row in history by as_of. Acceptance: sequential writers with skewed clocks leave the later publication current. |
| ADR012-R5-06 | ADR triage/CLI, REQ-78 and CLI reference require C advancement whenever terminal completeness permits, at triage publication and periodic trusted reconciliation, including quiet repositories; cadence remains implementation-chosen. Default is named triage at as_of; report frontier age and authorized rows after cutoff. Quiet-repository reconciliation is an explicit acceptance fixture. |
| ADR012-R5-07 | Each persisted frontier retains the protected-default-tip policy revision/effective minima. ADR, REQ-78 and CLI/policy references use the latest frontier≤as_of for default and historical queries, report the revision and label other-policy evaluation policy re-evaluation. Acceptance fixture: a minima change after T cannot change historical reproduction at T. |
| ADR012-R5-08 | Days=floor((as_of−origin)/24h) on the provider clock; origin is the earliest in-interval represented comparable observation with the selected check inspected. ADR amendments, REQ-22/78, card and references explicitly declare this inspected-only origin and exclusion of both not_inspected and not_applicable from PRRuns. No change to the settled origin. |
| ADR012-R5-09 | ADR items 1–4, REQ-27/78 and CI reference distinguish authenticated started assessment output failure (missing/invalid/irrecoverable means comparison-none) from trusted collector enumeration/attestation/publication failure (gap). Acceptance fixture: scanner deletion of its own output cannot reset an interval. |
| ADR012-R5-10 | ADR counting/identity/CLI, REQ-78 and CLI reference restrict cohort attempts to the exact pinned check version and links to matching events; choose representatives within that cohort. Default version is the latest-started in-scope attempt’s pin≤as_of, including rollback, not newest version. C<as_of≤reader now exits 2; as_of>reader now exits 3. Report every distinct Scout/tool/policy pin tuple and represented-observation count; mixed-pin counts are explicitly non-additive. |
| ADR012-R5-11 | Card now says Selects no workflow/key/commands/transport/retention/install mechanics, with REQ-162/163/168 unchanged. ADR revision/card declare the overview Activation callout. ADR amendments explicitly include trusted readiness publication (REQ-26/51/161) for ADR-000/003/007. Accepted ADR-011 summary/card/README row and R-064/R-069 remain untouched. |
| ADR012-R5-12 | Main restored round-4.md from the original council output to its full 392-line record, including the missing safety recheck, peer account, verdict and memo. Main attested that the existing 299-line bytes are an exact prefix and only the missing tail was appended. Main likewise restored round-5.md to its full 392 lines; its tail and required submission items were read for this response. Both remain unsigned review records, outside the canonical signed set. |

All five major fixes are observable invariants in ADR-012’s Implementation acceptance invariants section. B4-S’s acceptance column also names the eight requested fixtures: omitted failing PR, non-default base, late attempt, cross-gap not_warranted, skewed-clock publications, quiet-repository triage, policy change after T and scanner-deleted output. No fixtures or product checks were executed in this document-only unit.

## Review set

The canonical signed set remains these 13 adopted-form files (all under `docs/`):

1. `adrs/2026-11-08-shadow-readiness-inputs/ADR.md`
2. `scout-design.html`
3. `adrs/README.md`
4. `STATUS.md`
5. `plans/2026-11-06-remaining-work.md`
6. `references/policy-file.md`
7. `references/ci-integration.md`
8. `references/cli-commands.md`
9. `research-ledger.md` (R-077 now also cites round 5)
10. `adrs/2026-10-19-v1-decisions/ADR.md` (status only)
11. `adrs/2026-10-23-decisions/ADR.md` (status only)
12. `adrs/2026-10-24-cutover-cleanup/ADR.md` (status only)
13. `adrs/2026-10-31-jev-merge-association/ADR.md` (status only)

Historical rounds/responses, restored `round-4.md`, restored `round-5.md` and this response are review records, not canonical signed files. Main supplies their requested digests separately. Adoption remains outside the signed set. No commit or sign-off is asserted.
