# Architecture decision records

Each folder is one architecture decision record (ADR). `ADR.md` in each folder summarises the decision. The `round-<N>.md` files are the council review rounds, and the final round names the signed digests; an `ADR.md` that is itself among the signed files points to that round instead of holding its digests. ADR-000 is the exception: it records the v1 user decisions of 2026-10-19 and has no council round. The cards in Appendix B of `docs/scout-design.html` summarise these files.

Rules (`AGENTS.md` design rules):

- A new design change gets the next ID and a folder `docs/adrs/<YYYY-MM-DD>-<topic>/`.
- It is adopted only after an innovation-council round with zero issues.
- IDs are never reused.
- An accepted ADR is never edited, except its status line. The commit that adopts a later ADR updates that line to name what supersedes or amends it.

ADR-011 uses adopted-form route (a): its canonical files take effect on its folder's explicit final zero-issue council round, which names their signed SHA-256 digests; adoption is recorded outside signed files. ADR-011 was adopted in PR #17 at 7bb73e52; ADR-012 is based on it and is reviewed independently. No patch application or merge-method restriction is required. Owner confirmation is a new `user` ledger row and tracker update; rejection requires a superseding ADR.

ADR-012 uses canonical adopted-form route (a), effective on its final zero-issue council round, with adoption outside signed files. Based on adopted ADR-011 (PR #17, 7bb73e52); separate ordinary PRs replace the superseded patch route. Owner confirmation is a new user ledger row and editable trackers; rejection requires a superseding ADR.

ADR-014 uses adopted-form route (a): it takes effect on its folder's explicit final zero-issue council round, which names its signed SHA-256 digests; adoption is recorded outside signed files. IDs were allocated across concurrent work, so a gap does not imply a missing record.

| ID      | Folder                              | Decision                                                                                                    | Status                 |
| ------- | ----------------------------------- | ----------------------------------------------------------------------------------------------------------- | ---------------------- |
| ADR-000 | `2026-10-19-v1-decisions/`          | v1 baseline decisions: map refresh, baseline branch, allow-list, Jev, language tiers, spec cutover, pin     | partly superseded; amended by ADR-001, ADR-002, ADR-003, ADR-012 (readiness observations/eligibility, maintainer triage history, trusted readiness publication) |
| ADR-001 | `2026-10-21-e1b-v2/`                | Replace invalidated inertia evidence with E1b v2 (rev 2.1)                                                  | partly superseded      |
| ADR-002 | `2026-10-22-rescope/`               | Re-scope Scout to a product and an inertia-bench subject (rev 2.2)                                          | partly superseded      |
| ADR-003 | `2026-10-23-decisions/` | Record user decisions and align the design with shipped behaviour (rev 2.3) | partly superseded by ADR-007; amended by ADR-010 (triage minimums), ADR-012 (readiness observations/eligibility, maintainer triage history, trusted readiness publication) |
| ADR-004 | `2026-10-23-clawpatch-patch-v3/`    | Correct the Clawpatch review-safety patch pin (rev 2.4)                                                     | accepted               |
| ADR-005 | `2026-10-23-codex-luna-mapping/`    | Pin agent mapping to codex `gpt-6-luna` at `xhigh` (rev 2.5)                                                | partly superseded      |
| ADR-006 | `2026-10-24-area-refresh-fallback/` | Run the agent map only at onboarding and on manual refresh (rev 2.6)                                        | partly superseded      |
| ADR-007 | `2026-10-24-cutover-cleanup/` | Consolidate state, baseline, model-CLI and CI-installation cutovers (rev 2.7) | partly superseded by ADR-008; amended by ADR-012 (13→14 migration, export manifest v2, trusted readiness publication) |
| ADR-008 | `2026-10-31-jev-merge-association/` | Add provisional merge-time area association (rev 2.8) | accepted (provisional); amended by ADR-012 (schema 14, readiness export/merge) |
| ADR-009 | `2026-11-06-design-restructure/`    | Replace the design and requirements documents with `docs/scout-design.html`; rename review folders to ADRs | accepted               |
| ADR-010 | `2026-11-08-d1-jev-minimums/` | Record owner-agreed Jev triage minimums; retain provisional threshold and preregistered review | accepted |
| ADR-011 | `2026-11-08-gitleaks-policy-version/` | Keep the exact Gitleaks rule-pack activation version; enforce binary-segment invariant in later code | accepted; agent-decided under R-064; owner review pending; effective on final zero-issue council round |
| ADR-012 | `2026-11-08-shadow-readiness-inputs/` | Persist CI observations and authorized triage history; compute readiness under bounded evidence invariants | accepted; agent-decided under R-064; owner review pending; implementation pending; based on adopted ADR-011 |
| ADR-013 | `2026-11-08-x11a-delegated-decisions/` | Record X11a assignment-study choices, supply-informed1500–2000 segments, bounded preparation and uniform label-prefix gates | accepted-form; agent-decided under R-064; owner review pending; not effective until final zero-issue council round |
| ADR-014 | `2026-11-08-x1-delegated-decisions/` | Capture delegated X1 experiment-protocol choices; no product-default, flag, sampler or threshold adoption | accepted; decided by the agent under owner delegation (R-064) |

## Old paths and citations

Until 2026-11-06 this folder was `docs/design-reviews/`. Folder names did not change, so an older path `docs/design-reviews/<folder>/<file>` resolves to `docs/adrs/<folder>/<file>`. Frozen evaluation records, signed review rounds and earlier ledger rows keep the old path (R-062).

ADR-009 deleted the old design documents: `docs/scout-design.md`, `SCOUT_REBUILD_REQUIREMENTS.md`, `docs/scout-allowlist.md`, `docs/scout-design-brief.md`, `docs/inertia-research.md`, `docs/enaible-findings.md`, `docs/archive/`, and `docs/STATUS.md` §4 and §6. Commit `9a3d19f5` holds the last version of each. A frozen record that cites one of them by line or section means the revision current when the record was written: read that folder's `scout-design.base.md` or `scout-design.proposed.md` snapshot, or the file in git history at that date. For section-style citations (`design §n`, `Appendix N`, `AL §n`, `req §n`) to the last version, `2026-11-06-design-restructure/design-section-map.json` gives the current REQ and ADR IDs.
