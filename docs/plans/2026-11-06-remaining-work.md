# Remaining work (2026-11-06)

This is the single list of Scout work still to do. It replaces `HANDOFF.md` (deleted) and the former remaining-work table in `docs/STATUS.md`. The detailed specifications stay where they are, and this plan points to them:

- product requirements and experiments in `docs/scout-design.html` Appendix A (REQ-1…REQ-183);
- the per-phase specification in `docs/plans/2026-11-01-rev-2.8-merge-association.md`.

Where this plan and the design differ, the design wins. Machine setup, accounts and the repository gate are in `CONTRIBUTING.md`.

## Where things stand

| Area                                                | State                                                                                                                                         | Evidence                                         |
| --------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ |
| Design                                              | `docs/scout-design.html`, ADR-009 adopted 2026-11-06                                                                                          | `docs/adrs/2026-11-06-design-restructure/ADR.md` |
| Product roadmap (REQ-167, REQ-168, REQ-178–REQ-182) | Phases 1–4 done, with three open items (track B)                                                                                              | `docs/STATUS.md` §1                              |
| Merge-time association (rev 2.8 plan)               | Phases 0–2 done (`1d4e0cc`, `3e89f23`); the lane is not wired into `scout assess` yet                                                         | `docs/STATUS.md` §1                              |
| Repository gate                                     | Gate schema v2: the pre-commit hook runs `checks.commit`, the required CI check `gate` runs `checks.ci`, and `main` is protected (`5776d102`) | `.omp/gate.json`; `CONTRIBUTING.md`              |
| Toolchain                                           | Go 1.26.10, Clawpatch 0.7.3                                                                                                                   | `toolchain/versions.env`                         |

Two phase series exist. Track A numbers follow the rev 2.8 plan (A3–A7), and track B numbers follow the product roadmap (REQ-167, REQ-168, REQ-180–REQ-182; Phases 3–6). They are different phases.

```mermaid
flowchart LR
  A3[A3 association lane] --> A4[A4 fallback lane]
  A3 --> A5[A5 pending areas]
  A4 --> A6[A6 refresh guard]
  A3 --> A6
  A4 --> A7[A7 orderflip record + docs]
  A5 --> A7
  A6 --> A7
  C[Track C preregistrations] --> B4[B4 Phase 5]
  D1[D1 Jev costs and minimums] --> C
  B4 --> B5[B5 Phase 6]
  A3 --> X11[X11 run]
```

## Track A: merge-time association (rev 2.8 plan Phases 3–7)

The lane ships off by default as a provisional opt-in. The fastapi result did not replicate on axum (P_assign 0.922 vs 0.622, R-061).

| #   | Work                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | Acceptance                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | Depends on              |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| A3  | **`association` lane (Jev).** <br>1. Register `merge.covered@v1` and `merge.which_area@v1` with the pilot's frozen question text and criteria (`evaluation/x2d-merge-association-pilot/questions/`). <br>2. Add the `mapping.association.*` policy keys. <br>3. Add a push-only optional lane with `requires: decision-egress`. <br>4. Wire `association.Run` into `scout assess`. <br>5. Add a store-backed `association.Ownership` with all four methods (`AreaOwners`, which must include association-derived owners, `PendingAreas`, `HeuristicOnly`, `SiblingFiles`). Phase 2 has only the `overlay` wrapper and a test fixture. <br>6. Call `Decider` once per merge, and map results to `domain.Association`. <br>7. Add staleness marking. | Replay of the saved fastapi rounds 8–10 scores (`evaluation/x2d-merge-association-pilot/data/postrun/rounds/{8,9,10}/jev/scores.jsonl`) gives **64/67 assigned correct and 10/92 unresolved**. Per round that is 14/14, 8/10 and 42/43. Settings: covered score 0.60 (`none` when P(true) ≤ 0.40), `t_choice` 0.535, and a second permutation when `sha256("x2d-flip"+cluster) mod 5 == 0`. A recomputation on 2026-11-06 reproduced these numbers exactly. The pilot's provisional `thresholds.json` (0.70) is not the frozen setting. The plan's scenario and test list also applies. | —                       |
| A4  | **`association-fallback` lane** (`codex exec -m gpt-6.1-terra`, effort high). Its prompt is `evaluation/x2d-postrun-tuning/terra_prompt_v1.txt`, digest-checked. Consent is captured in `scout init`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | Replay with a fake `codex` gives **71/77**, 0 unresolved, and 13 pending on gold-`none`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | A3                      |
| A5  | **Pending areas:** `scout areas list --pending`, `accept` and `merge`. Follow the `scout triage --actor/--reason/--authorize` pattern.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | Plan Phase 5                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | A3 (state from Phase 1) |
| A6  | **Refresh guard** (`internal/association/guard.go`). Its defaults (20/20/0.40/0.20) are assumptions.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | Plan Phase 6                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | A3, A4                  |
| A7  | **Release gate:** wire `orderflip`, record the `merge.which_area` flip rate in `docs/jev-adapter-audit.md` (B-3), and update the docs and changelog. No release may enable the lane before this is done.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Plan Phase 7                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | A3–A6                   |

Each phase also needs `go test ./...`, the commit hook, its phase scenario, and a record in `docs/STATUS.md` §1.

## Track B: product roadmap (REQ-167, REQ-168, REQ-180–REQ-182)

| #   | Work                                                                                                                              | Entry condition                     | Notes                                                                                                                     |
| --- | --------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| B1  | Phase 3 CI item: observe `pr_effort` in a trusted default-branch CI run                                                           | a trusted run on `main`             | Wired already (`scout-guardian-trusted.yml`, `pull-requests: read`) but never observed                                    |
| B2  | Re-run Phase 4 under the rev 2.6 acceptance on fastapi `d666ccb` and axum `562fa20`, and record the result in `docs/STATUS.md` §1 | `ead1a3f` (done)                    | The scenario script `/tmp/scout-p4/scenario.sh` was lost with the old machine. Rebuild it against the amended acceptance. |
| B3  | Phase 3b: the exact schema for the subject interface                                                                              | inertia-bench design signed off     | Needs its own council round                                                                                               |
| B4  | Phase 5: trends, exploratory measures, the rubric lane, self-evaluation, and agent-labelled eval data                             | track C preregistrations signed off | REQ-34, REQ-67–REQ-69, REQ-108–REQ-110, REQ-123–REQ-124, REQ-136–REQ-138, REQ-182                                         |
| B5  | Phase 6: study export (the dogfood ledger is dataset #1), web inertia panels, release, dogfood in `guardian`                      | B4                                  | REQ-36, REQ-168. The spec cutover is done by ADR-009.                                                                     |

## Track C: experiments that need a council-signed preregistration

None of these is preregistered yet. Each one follows the experiment rules in `CLAUDE.md`: preregister, plant an effect and a null, get council sign-off with zero issues, then freeze.

| ID              | Question (REQ-169–REQ-177)                                                   | Feeds                             |
| --------------- | ---------------------------------------------------------------------------- | --------------------------------- |
| X1              | Jev ranking vs deterministic ranking                                         | B4; REQ-169                       |
| X3, X4, X5, X10 | Phase 5 measures, the rubric lane and self-evaluation                        | B4                                |
| X6              | Replay of unwarranted guardrail blocks                                       | allow-list promotion              |
| X9              | Hand-check of 200 `moved/pasted/new` labels and 200 `stale_reference` labels | B4                                |
| X11             | Merge-time association plus the refresh guard (tests HD8)                    | upgrades track A from provisional |

HD6b and HD7 are `proposed` and have no test yet (`docs/hypothesis-ledger.md`).

## Owner decisions

| #   | Decision                                                                                                                                                          | Blocks                                                      |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| D1  | Jev false-action costs and per-decision minimums (REQ-98). The provisional defaults 0.8/0.8, raised to 0.9 on the costlier side, are in force until you set them. | the minimums in the X1, X3, X4, X5 and X10 preregistrations |

## Known gaps (from `docs/STATUS.md` §1)

- When agent mapping succeeds with zero features, the heuristic output is not re-merged.
- An issue left unmatched during a rename keeps its old anchor path.
- The policy `version` strings for Gitleaks checks are the long rule-pack version.
- The shadow-window bound (`unknown` ≤ 10 % of PR runs) cannot be computed yet. `ShadowStats` has no producer, and `comparison` is not persisted (REQ-22, REQ-78).

## Candidate source (not in the ledger)

MOSS, arXiv:2610.26902: an LLM-judge monitoring loop whose thresholds transferred across benchmarks. It may matter for X1 and nightly mode. Read the full paper before using it. If it informs a decision, record it under the next free research-ledger ID.

## Docs that stay

- **Governed design:** `docs/scout-design.html`. Appendix A holds REQ-1…REQ-183; Appendix B indexes ADR-000…ADR-009.
- **Decisions and review evidence:** `docs/adrs/`, indexed by `docs/adrs/README.md`.
- **Reference:** `docs/references/` describes the CLI, lanes, policy, environment and CI.
- **Plans:** `docs/plans/` keeps this remaining-work list and the per-phase merge-association specification.
- **Current state:** `docs/STATUS.md` keeps code/phase state, evidence and resume pointers.
- **Ledgers:** `docs/research-ledger.md` and `docs/hypothesis-ledger.md` keep sources and hypothesis state.
- **Jev audit:** `docs/jev-adapter-audit.md` keeps the adapter audit and open design-level gaps.
- **Deleted by ADR-009:** `docs/scout-design.md`, `SCOUT_REBUILD_REQUIREMENTS.md`, `docs/scout-allowlist.md`, `docs/archive/`, `docs/scout-design-brief.md`, `docs/inertia-research.md` and `docs/enaible-findings.md`. Git keeps their history.
- `HANDOFF.md` was deleted in the earlier tidy-up. Its setup and account notes moved to `CONTRIBUTING.md`, and its remaining work moved here.
