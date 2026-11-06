# Round 2 response — revision 2.3 decisions proposal

Drafter's response to `round-2.md` (9 issues: 0 blocker, 1 major, 5 minor, 3 nits). All nine are addressed in the round-3 draft. Line numbers are in `scout-design.proposed.md` (466 lines). Not signed off; the council must re-read every hunk of both diffs and recompute the digests.

Digests of the round-3 draft (`shasum -a 256`):

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` (= `docs/scout-design.md`, rev 2.2) | 462 | `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4` |
| `scout-design.proposed.md` | 466 | `07ccff552d42859447b1654353f8c4a83fa24003e899629a352019e59e37f236` |
| `proposed.diff` | 236 | `bd42d686fb17038a382db12d23d654b6ef8b9af9f29ab8d6f0b3ff79bbea44e3` |
| `requirements.proposed.diff` | 49 | `b4fd5d8d1de93ae391f1c447af34c43c31efd3541c5651a3b3cbe73f698ed397` |
| `PROPOSAL.md` | — | `8890be1ab05218b4da8e0ff688db931441645a415ada6bd2c44cfe0269e18563` |

## Facts that moved (committed on main)

- `480949a` "Attribution: findings in lanes the base did not inspect are unknown, never introduced": paired recompute carries base lane health (lanes `blocked`/`failed` are uninspected, family granularity); imported baselines carry an optional manifest `uninspected` list of `{family, reason}`; a legacy baseline without it is recomputed with a banner; affected head findings are `unknown`, excluded from `would_block` and the gate with one gate reason per family; `guardrails.*` exempt (`internal/engine/{diff,gate}.go`, `base_gaps_test.go`, `internal/baseline/load.go`, `internal/state/export.go`).
- `404f69b` "CI: provision pinned scanners from the release tag": both reusable workflows and the composite action shallow-clone `release-repo` at `scout-version` outside the workspace, run its `scripts/bootstrap-tools.sh` into `$RUNNER_TEMP`, add it to `PATH` via `$GITHUB_PATH`, cache keyed on the tag, install bubblewrap; nothing from the PR checkout runs (reason: dogfood run 39962565095 incomplete, scanners missing). Described as shipped and not yet verified in CI (verification run in progress).

## Issues

| # | Sev | Fix | Where |
|---|---|---|---|
| 1 | major | Design now describes the shipped coverage-aware attribution. §7.3 `introduced` requires a base whose scanner lane inspected the finding's family (recompute, or a baseline that records lane health); `unknown` is the whole run when no comparable base can be produced (`comparison: none`, banner `no baseline available (…); nothing is classified introduced`) or per finding when the base did not inspect its family (lane `blocked`/`failed`, or manifest `uninspected`; legacy baseline recomputed with a banner); excluded from `would_block` and the gate, one gate reason per family; guardrails exempt. §7.2 states the comparability conditions (tool and policy digests match, revision is the merge base, lane health recorded; otherwise recompute with a banner naming the reason) and documents the manifest `uninspected` field. Counting unit of the ≤ 10 % rule stated: runs, one per run, that are `comparison: none` or carry at least one base-gap `unknown`, over all PR runs in the window; recomputed runs stay comparable-base runs. "Known gap … outside this revision" deleted; `:7` item (4), §7.4 (also adds the CI toolchain sentence), Phase 2 acceptance, A2, REQ-27 (and REQ-25's CI clause), PROPOSAL D4/"Explicitly not changed"/new D8 row, DRAFT-NOTES call updated. | `:7`, `:182`, `:186`, `:189`, `:197`, `:360`, `:425`; REQ-25, REQ-27; PROPOSAL D4 |
| 2 | minor | Row list corrected everywhere: "Design-level rows 5, 35 and 38 remain open; backlog B-1–B-4 records the follow-ups (B-3 runs the row-30 flip-rate harness on the first Choice that gates anything)". | `:7`, `:166`, `:466`; REQ-29; PROPOSAL D2 |
| 3 | minor | Phase 4 survival is "active areas before the refresh still active after ÷ active areas before" (retired areas keep IDs and do not count); match wording is "matched IDs kept, unmatched agent features become new areas, unmatched active areas kept or retired per §4"; matched and kept counts reported separately, kept areas contributing Jaccard 1.0 apart from the matched distribution. | `:372`; PROPOSAL D6 |
| 4 | minor | Option (a): "the Phase 4 scenario reports …" (measured by the scenario harness and recorded in STATUS §1; `scout map` does not print it, `cmd/scout/map.go:102` discards the report). Choice recorded in PROPOSAL D6. No STATUS gap entry added (that is option (b)). | `:372`; PROPOSAL D6 |
| 5 | minor | One scope list in four places (`:166`, `:440`, `:466`, REQ-29): defaults substitute only for precision/recall point minimums (X3, X10, point values in X5); X1's gain threshold and the lower-bound forms (X4, X5) are confirmed or replaced by their preregistrations without a default substitute; no new X1 number. PROPOSAL D2 mirrors it. | `:166`, `:440`, `:466`; REQ-29 |
| 6 | minor | Precedence scoped: "`human` overrides `agent_reviewed` on the same unit for the same question; within one provenance the latest adjudicated label under the current guide version wins"; per-session rows (provenance `agent`) are inputs to the adjudicated label, not labels. Mirrored in Appendix 15 and PROPOSAL D3. | `:168`, `:462` |
| 7 | nit | `:3`: "its X2-dependent figures are reported (not a Phase 4 pass criterion, §11)". PROPOSAL header row and bullet: "wired; `pr_effort` in trusted CI not yet observed (STATUS §1)". | `:3`; PROPOSAL D8/"Explicitly not changed" |
| 8 | nit | The §11 `[ASSUMPTION, Xn]` sentence and the `:7` clause are dropped; the Phase 4 text states the rule in words. | §11 intro; `:7`; PROPOSAL D6 |
| 9 | nit | DRAFT-NOTES line map and digest table refreshed against the round-3 file. | `DRAFT-NOTES.md` |

## Other files touched

- `docs/STATUS.md` §1: Known-gaps attribution entry changed from "being fixed" to fixed (`480949a`) with the mechanism; the Phase 2 "Deviation" bullet (flagged stale in round 2 §3 issue 1) now says per-finding `unknown` also arises from base lane gaps; the workflow-assumptions bullet records the CI toolchain provisioning (`404f69b`), shipped and unverified in CI. Factual only.
- `requirements.proposed.diff` regenerated by applying the previous diff to a copy of `SCOUT_REBUILD_REQUIREMENTS.md` (untouched in the repo), editing REQ-25, 27, 29, and re-diffing with `a/`/`b/` headers and no timestamps. `proposed.diff` regenerated with `diff -u` (header timestamps as before).
- `docs/scout-design.md` not edited; no commits.

## Choices made

- Counting unit: runs, one per run (not per finding), matching `ShadowStats.UnknownAttribution`/`TotalRuns` (`internal/policy/policy.go:386-394`, `policy_test.go:157-158`).
- Issue 4 option (a) (text-only) over (b).
- Issue 2's row-30 harness wording taken verbatim from the council's fix.
- The §4 and `:3` Phase 4 figures remain in STATUS only.

## Not run

No tests, builds or council re-review in this step. Main agent: confirm `shasum -a 256 docs/scout-design.md` is still `f81b2601…`; reproduce `proposed.diff` with `diff -u scout-design.base.md scout-design.proposed.md` in the review directory; send round 3 to the council.
