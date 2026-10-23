# Scout — status and handoff (2026-10-22)

Written for a fresh session. Every claim cites a file in this repo unless marked `[INFERENCE]` or `[UNVERIFIED]`. Initial commit `41ea1eb` (2026-10-22) holds the whole tree; `.codegraph/`, `.scout/`, `__pycache__/` and `*.pyc` are git-ignored. Never touch `.codegraph` (`AGENTS.md:54`).

## 1. Current state (verified 2026-10-22)

**Design: done and signed off.** `docs/scout-design.md` is revision 2.1; `shasum -a 256` = `7ff21a7fba70a5e3af7a735b0aa40f1ad654d17714df8c3549da6ebcf89b89c5` (matches the sign-off digest in `docs/design-reviews/2026-10-21-e1b-v2/round-5.md:40`). Council rounds 1–5 are in that directory (`round-1..5.md`, responses, `proposed.diff`, `scout-design.base.md`). v1 is `docs/archive/scout-design-v1.md`.

**Code: v1 Go implementation with design-v2 Phase 1 implemented, not yet committed (2026-10-22, §5 row 1).** `go build ./...`, `go vet ./...` and `go test ./...` pass. `internal/{analyzers,clawpatch,contextual,domain,inventory,issues,state,web}` and `cmd/scout` have tests; `engine` and `eval` have none.

| Exists (v1, kept/rewritten per design §10.2) | Evidence |
|---|---|
| CLI `readiness assess issues issue mapping coverage triage serve` (`approve` and `verify` removed in Phase 1; resolution is automatic after a relevant re-check) | `cmd/scout/main.go:90-108` |
| Inventory/plan/diff; analyzers (Semgrep, Gitleaks, Ruff adapters; catalogs; architecture catalog returns nil) | `internal/inventory/`, `internal/analyzers/` |
| Clawpatch client pinned **0.7.3** / `734f81bc…`, 900 s map timeout, sanitising review patch | `internal/clawpatch/client.go:19-22`, `scripts/build-clawpatch.sh`, `scripts/clawpatch-safe.patch` |
| SQLite state, issues, progress; read-only web | `internal/state/`, `internal/web/web.go` |
| Contextual LLM client (generic JSON/HTTP endpoint from `SCOUT_PROVIDER_*` env) | `internal/contextual/client.go`, `cmd/scout/main.go:128-130` |
| Self-eval corpus replay (v1 corpus: 9 true / 13 FP) | `internal/eval/`, `evaluation/{reference-v1,outcomes-v1,outcomes-v2,scout-requests-v1,scout-requests-v2}.json`, `evaluation/replay.go` |
| Workflows: `scout-nightly.yml` (237 lines, builds `./cmd/scout` from the checkout; passes `OPENAI_API_KEY`/`ANTHROPIC_API_KEY` secrets at `:172-173`), `release.yml` | `.github/workflows/` |

| Design-v2 components that do **not** exist | Design ref |
|---|---|
| `internal/{measures,history,artifacts,guardrails,decisions,baseline,selfeval}`; mapping areas/refresh trigger; `scout init/diff/export/import/signals/trends/annotate/selfeval`; `.scout/policy.yaml`; `scout-guardian.yml`; `docs/scout-product-contract.md`, `docs/scout-adapter-contracts.md`, `docs/evidence/` | §3.1, §10.2–10.3, §11 |

**Phase 1 (design §10.1, §11) — done in the working tree:**
- **Persistence.** `plan_jobs` table (schema v4). `CommitRunJob` reads one job row and no longer rewrites `runs.payload`. Run progress is rebuilt from `coverage` and `run_job_commits` (`internal/state/progress.go`). Proven by `TestCommitRunJobCostIsFlatInJobCount`: 5,000 jobs against a 5 MB plan, first and last 500-job buckets 232.8 µs vs 223.5 µs median, with a corrupted `plans.payload` and an abort-trigger on `runs.payload` updates. `TestInterruptedRunResumesToSameFinalRun` proves resume equivalence.
- **Identity.** Design §3.3 ladder (`internal/issues/correlate.go`, `internal/domain/issue_anchor.go`, `internal/state/issue_match.go`). IDs are `issue:<family>:<ulid>`. The evidence fingerprint includes the normalised message. Secrets match by an in-memory SHA-256 digest (`internal/analyzers/gitleaks.go`). Renames and deletions use the latest *recorded* inventory, partial runs included, while coverage reuse keeps the complete-run baseline. Migration to schema v5 re-anchors v1 issues, keeps `legacy_id` and lists ambiguous rows (`internal/state/issue_migration.go`). Approval machinery is removed.
- **Offline map.** `Map(..., MapSourceHeuristic)` runs `clawpatch map --source heuristic` with no egress. Agent mode needs egress and falls back to heuristic output with `ExplorationIncomplete` (`internal/clawpatch/client.go`). `internal/codex` is cut (§10.2).
- **Scanner batching (found during toolchain verification).** File-analysis jobs used to run one scanner subprocess per (file, category) job, re-checking the version every time. That made fastapi exceed 20 min once Semgrep actually ran. Jobs are now grouped per adapter spec and chunked by the spec's input limits (`internal/analyzers/batch.go`), the version is verified once per run, and each job still commits its own per-file slice through `CommitRunJob` (`internal/engine/assess.go`).
- **Acceptance evidence.**
  - fastapi `d666ccb`, toolchain activated (pinned Semgrep, Ruff, Gitleaks and Clawpatch), no provider env, no `--allow-egress`: **455 s** (< 600 s), `mappingStatus: mapped`, 74 issues. Every scanner category is `inspected`; the only non-inspected rows are the opt-in contextual and Clawpatch-review lanes (`blocked`). Earlier timings of 389–448 s are void, because Semgrep lanes had exited on the version mismatch.
  - Throwaway scenario (`/tmp/scout-p1/scenario.sh`; ruff C90/PERF fixture, 3 runs): one ID across a 20-line insertion with C901 drift (`created, triaged, evidence_drift, matched`, stays `accepted`); one ID across `git mv`; two IDs for two PERF401 occurrences in one file; `resolved` then `regressed` on the same ID. Rerun after batching with the same result.
- **Known gaps.** Real runs still end `partial`/exit 2, because the opt-in contextual and review lanes report `blocked`; lane health and optional lanes are Phase 2 (§3.4, §7.3). Gitleaks stays at one file per call (`MaxFiles: 1`) and is the largest remaining cost in the 455 s. If agent mapping succeeds with zero features, it does not re-merge heuristic output (Phase 4). An issue left unmatched during a rename keeps its old anchor path.

**Toolchain (2026-10-22).** Pinned project-local toolchain in `.tools/` (§7). Clawpatch is a git submodule at `third_party/clawpatch` (gitlink `734f81bc…`). Pins live in `toolchain/`, and `internal/toolchain` tests enforce parity with the Go code. Both workflows provision through `scripts/bootstrap-tools.sh`. Go moves from 1.25.0 (old CI pin) to 1.25.6 (`go.mod` `toolchain`).

**Jev labels (2026-10-22/23).** All 52 E2 triage cases are labelled (§6.8). Agents did the first pass: A/B agreement 48/52 (κ 0.856), 4 adjudicated. The user reviewed every case on 2026-10-23 and confirmed all 52 proposals, recorded as `human` rows (`by: user`) appended to `evaluation/jev-triage-labels-v1/labels.jsonl`. Final labels: real_defect 28, false_positive 23, cannot_tell 1. No Jev-versus-label comparison has been run; that needs a council-signed preregistration. `E2-COMPARISON.md` (proposed labels vs the original E2 adjudication: 1 of 22 differs) can now be read.

## 2. Decisions in force

- Design revision 2.1 is authoritative; the Appendix decisions (map refresh, `scout/state` branch, allow-list, Jev on from launch, tier-1 = Go/Python/TypeScript, spec cutover, Clawpatch 0.7.3) are not reopened (`docs/scout-design.md:409-417`). Of the five §15 decisions (`:397-403`), §15.3 is now decided and §15.2 leaves with the re-scope; §15.1, §15.4, §15.5 stay open (§6).
- Scout is a **discovery-only CI guardian**; no remediation, no inertia score, no claim that agents degrade a repo (`docs/scout-design.md:9-19`). All inertia measures are `descriptive` (`:185`).
- Design changes need innovation-council sign-off with zero issues, rounds recorded under `docs/design-reviews/<date>-<topic>/` (`AGENTS.md:30-38`). Experiments likewise, preregistered, frozen after sign-off (`AGENTS.md:3-28`). Design changes may cite only signed-off experiments (`AGENTS.md:37`).
- Ledgers must be updated when a source/hypothesis is used (`AGENTS.md:40-50`): `docs/research-ledger.md` (R-001…R-037), `docs/hypothesis-ledger.md`.
- Secrets only via Infisical per command; never print/export/commit (`AGENTS.md:52-61`).
- **User decisions of 2026-10-22** are recorded in §6 (R-036). Design-affecting ones take effect only after the council re-scope round `docs/design-reviews/2026-10-22-rescope/` signs off.

## 3. Evidence to date (one line per study)

| Study | Outcome | Source |
|---|---|---|
| E1 inertia-v1 (pre-rules) | Quality set adds +1.18 % [−0.62, +3.24] relative MSE to fix-proneness prediction: fails the ≥5 % rule; signals stay descriptive | `evaluation/inertia-v1/RESULTS.md:7-19` |
| E1b-v1 change-cost-v1 | **Invalidated** (baseline worse than constant, outcomes not effort); not cited as evidence | `docs/research-ledger.md` R-033; `evaluation/change-cost-v1/METHOD-CRITIQUE.md` |
| E1b-v2 change-cost-v2 (council-signed prereg round 10; Amendment 1 annotator-C truncation round 14; run 2026-10-21) | 7,717 PRs / 12 repos. H1 (T2) +0.023 log-h/SD, 90 % CI [+0.004,+0.041], `equivalent`; H2 (T4) +0.006 [−0.010,+0.023], `equivalent`; level tests T1/T3 equivalent; P1 incremental gain −0.15 %; **H4 not decided** (gate G5 recall 0.422 < 0.60; G7 κ 0.514, hidden precision 0.630 fail) | `evaluation/change-cost-v2/RESULTS.md:5-33`, `reviews/round-10.md`, `reviews/round-14.md` |
| E2 jev-triage-v1 (pre-rules) | Jev choice-variant 17–21/22 correct, AUROC ≈ 0.94–0.99, ≈ $0.05/1k decisions, p50 ≈ 0.3 s; n = 22, held-out 12; v3 was tuned on tuning split; route-only | `evaluation/jev-triage-v1/RESULTS.md:16-43` |
| E3 map-stability-v1 (pre-rules) | Heuristic feature IDs 100 % stable; agent-mapper IDs 5–31 % recur → agent IDs cannot anchor issues | `evaluation/map-stability-v1/RESULTS.md:3` |
| CP07 clawpatch-0.7-compat (pre-rules) | Upgrade to 0.7.3 recommended; heuristic IDs retained on fastapi/caddy/nest/webpack, axum 50/52 records change once; patch ported; agent churn unchanged | `evaluation/clawpatch-0.7-compat/RESULTS.md:3-5`, design §4 |

**Hypothesis ledger now** (`docs/hypothesis-ledger.md:14-28`): H1 not-supported, H2 not-supported, H3 proposed, H4 inconclusive, H5 proposed, H6 proposed, H7 proposed; HA1 proposed (no test), HA2 proposed (X8), HA3 proposed, HA4 proposed; HQ1 not-supported; HD2 inconclusive; HD3 refuted; HD4 supported. Latest research entry **R-035** (`docs/research-ledger.md:43`).

## 4. What moved to inertia-bench, and the subject interface Scout must provide

Per the task brief, inertia-bench (`/Users/adamjackson/Projects/tools/inertia-bench`) is now the research lab on how codebase inertia affects AI coding agents; Scout is a product and a future **subject**. Its design is in council review; the latest recorded round in `docs/reviews/design/` is `round-5.md` (round 6 in progress per the brief; not signed off). Consequences for Scout: hypothesis-testing studies of H1–H7/HA* (design §8.5 items (a)–(d), X7, X8, §15.2) are no longer Scout's work to run; Scout keeps only product-validation experiments (X1–X6, X9, X10) and the data-export hooks (§8.6).

**Contract** (`inertia-bench/docs/design/05-harness-design.md:879-920`): a digest-pinned OCI image or CLI with two commands, run with the snapshot mounted **read-only, no network, wall-clock cap** (request `budget.wall_s` 900, `network:false`):

1. `capabilities --out caps.json` → `{"schema":"ib-subject-caps/1","provider","provider_version","measures":[{"measure_id","unit_kinds":["file"|"function"|"area"],"languages":[…],"direction":"higher_is_worse"|"higher_is_better","value_type":"number"|"integer"}]}`.
2. `measure --request request.json --snapshot /snapshot --out measures.jsonl`; request schema `ib-subject-request/1` (`snapshot_id`, `languages`, `units[{unit_id,kind,path,qualified_name,start_line,files}]`, `measures_requested` (`"*"` = all matching), `budget`). `area` units list member `files`.
3. Output lines `ib-subject-measure/1`: `provider, provider_version, provider_digest, unit_id, measure_id, value, unmeasured_reason`; exactly one line per requested (unit, measure); `value` finite or `null` with `unmeasured_reason ∈ {unsupported_language, parse_error, timeout, not_applicable, other}`; sorted by (`unit_id`, `measure_id`), UTF-8, LF; missing pair = `missing`; duplicates, non-finite values, unknown ids fail conformance.
4. Conformance: schema validation, two runs byte-identical, offline, within budget, coverage per measure/language. Blinding: opaque snapshot/unit ids, no arm labels or outcomes; "A Scout adapter lives outside this repository" (`:919-920`).
5. Planned evaluations of a provider (each preregistered in the bench): criterion validity vs lab reference measures, manipulation sensitivity (AUROC), incremental predictive validity on held-out repos, intervention mode (`:913-917`).

`[INFERENCE]` mapping: design §8.2 snapshot measures (`scout.cc@1`, `scout.complex_burden@1`, `scout.burden_density@1`, `scout.dup_pct@1`, `scout.cycle_mass@1`; all higher = worse) fit `file`/`function`/`area` (a Scout scope ↔ an `area` with `files`); Scout `unmeasured` ↔ `unmeasured_reason`. History/PR measures (§8.3) need git/PR data a bare read-only snapshot may not carry → likely out of the contract's first version; `scout.crap@1` needs a coverage artifact → `not_applicable`. Scout's state (`.scout/`, SQLite) must not write into the read-only snapshot, and the heuristic mapper must run offline (design §10.2 says it currently blocks without egress, `client.go:168-170`).

## 5. Remaining work (ordered)

Owners: **Council** = innovation-council review loop; **Impl** = a coding session; **User** = decision. Every implementation phase is a design-§11 phase; acceptance criteria are the ones in design §11 and are not restated here.

| # | Work | Entry condition | Owner | Acceptance |
|---|---|---|---|---|
| 0 | ~~Decide the §6 conflicts~~ **done 2026-10-22** (§6); council re-scope round | proposal written | Council | Round signed off with zero issues; design digest updated |
| 1 | **Phase 1** — fix `CommitRunJob` persistence, rewrite issue identity/anchoring, offline heuristic map, drop approval machinery, migration | met (§15.3 decided, §6.7) | Impl | **Done in working tree 2026-10-22, uncommitted** (§1). Design §10.1 + §11 Phase 1 (flat-per-commit benchmark, durable ID across line shift/rename, `evidence_drift`, `regressed`, heuristic map offline, fastapi assess < 10 min) |
| 2 | **Phase 2** — differential guardian, `scout/state` branch, `init`/consent, `decisions` adapter with deterministic fallback, `scout-guardian.yml` (replace nightly workflow built from checkout) | #1 | Impl | §11 Phase 2 (PR A–D scenarios, CAS publish, byte-identical import) |
| 3 | **Phase 3** — measures, history, artifacts, guardrails for Go/Python/TS | #2 (can start `measures` earlier; independent of identity) | Impl | §11 Phase 3, incl. reproducing `active_hours` in `evaluation/change-cost-v2/panel-final.csv.gz` within 1e-9 h |
| 3b | **Subject interface** (`capabilities`/`measure`) for inertia-bench, offline/deterministic, snapshot-only measures | inertia-bench design signed off; #3 `measures` exist; design change signed off by council (the interface is not in design 2.1) | Council → Impl | Passes bench conformance (§4 above); byte-identical double run |
| 4 | **Phase 4** — mapping refresh triggers + areas (X2) | #1, #3 scopes | Impl | §11 Phase 4 |
| 5 | **Phase 5** — trends, exploratory measures, rubric lane, self-evaluation (X1, X3–X5, X9, X10), blind-label collection | #2–#4; ≥100 blind labels for X1 | Impl | §11 Phase 5; X1 decision rule recorded (§13) |
| 6 | **Phase 6** — study export, spec cutover (archive `SCOUT_REBUILD_REQUIREMENTS.md`, write `docs/scout-product-contract.md`, `scout-adapter-contracts.md`, `docs/evidence/`), web inertia panels, release, dogfood in `guardian` | #5, §15.1 consent decision | Impl | §11 Phase 6 |
| 7 | ~~Update `AGENTS.md`~~ **done 2026-10-23** (§6.6) | — | — | — |
| 8 | **Candidate reference (not a ledger entry yet):** MOSS, arXiv:2610.26902 (Nesen et al., published 2026-10-16; verified abstract: production LLM-judge monitoring loop with targeted sampling, strategy-aware prompting, structured alerting; a single judge/score scale/threshold transferred from a 498-pair recommendation benchmark to a 466-pair search benchmark with only a surface-specific prompt change). Relevance: Scout persistent/nightly mode (§7.1, §8.4 trend flags, 5 % audit sample §6) and Jev triage thresholds (X1, §13: are thresholds transferable across rules/repos?) | read the full paper before use | Impl/User | If it informs a decision or hypothesis, add the next free `R-NNN` (R-038) per `AGENTS.md:42-45` at that time |
| 9 | ~~Repository hygiene~~ **done**: initial commit `41ea1eb`; `__pycache__/`, `*.pyc` ignored; study data (≈ 57 MB under `evaluation/`) committed because Phase 3 acceptance reads `change-cost-v2/raw` and `panel-final.csv.gz`. Pre-commit Gitleaks: 7 hits are public example tokens in corpus evidence (`evaluation/jev-triage-v1/unlabelled.jsonl`), 1 is a rule regex (`internal/analyzers/rules/security/gitleaks-default.toml:232`) | — | — | — |

## 6. Decisions on the former open questions (user, 2026-10-22; R-036)

1. **Model use — product-wide CLI adapter.** The vendor-neutral HTTP contextual client (`internal/contextual/client.go`, `SCOUT_PROVIDER_*`) is replaced by a subprocess adapter over `claude -p` / `codex exec` (Anthropic/OpenAI models only). Reasoning-LLM lanes (`contextual`, `clawpatch-review`, agent mapping) run only where a CLI is logged in (local or self-hosted runner); on GitHub-hosted runners they report `skipped (no model CLI)`. `scout-nightly.yml:172-173` drops `OPENAI_API_KEY`/`ANTHROPIC_API_KEY`. Jev stays on OpenRouter (Infisical locally, a GitHub secret in CI). Design change → re-scope round.
2. **E1b-v2 grandfathered with a note** (`evaluation/change-cost-v2/RESULTS.md:5`; R-036). Preregistration, amendment and scripts untouched.
3. **Subject interface in-repo, two-step.** `scout subject capabilities|measure` in the Scout binary; no third repo. The re-scope round adopts principles only (snapshot-only measures, offline, deterministic, never writes to the snapshot, `unmeasured` ↔ `unmeasured_reason`); the exact schema is a separate council round after inertia-bench's design signs off (bench round 6: issues found). Implementation after Phase 3 `measures`.
4. **Research moves to inertia-bench.** Remove §8.5 (a)–(d), X7, X8, §15.2 from the design and point to the bench; keep the §8.5 claim-upgrade rule, now satisfiable by a signed-off bench study; keep `scout export --study` as a product feature with Scout-owned consent (§15.1); the bench runs pooled analyses. Scout keeps product experiments X1–X6, X9, X10. Design change → re-scope round.
5. **Spec cutover at Phase 6 as designed.** `SCOUT_REBUILD_REQUIREMENTS.md` stays the live product spec until then.
6. **`AGENTS.md` full update — applied 2026-10-23** after user approval: model-use rule; experiment rules cover Scout product experiments (X1–X6, X9, X10), while hypothesis studies run in inertia-bench under its rules; hypothesis ledger frozen for H\*/HA\*/HQ\*, HD\* stays live; Infisical is the only local source of the Jev key, a GitHub secret in CI; `claude -p`/`codex exec` use subscription login; work inside the pinned toolchain; Jev labelling protocol (§6.8).
7. **Performance target (design §15.3) confirmed:** `scout assess` on fastapi `d666ccb`, heuristic map, egress off, < 10 min on the development Mac.
8. **Jev labelling: agents do the first pass, the user reviews every label** (user, 2026-10-22). Whenever Jev needs labelled cases (eval sets, threshold tuning, X1/X3/X4/X5/X10), independent labelling sessions propose labels and the user confirms or adjusts each one. Protocol, following the `jev-actions` skill (`labeling-cases.md`, `building-eval-data.md`):
   - **Blind.** Labellers, and the user's review sheet, show only the unit's evidence: never Jev's answer, rank, score, stratum or a prior verdict.
   - **Guided.** A one-page label guide per question, versioned with the question's version key and frozen before labelling. Its worked examples are synthetic, never taken from the cases being labelled.
   - **Independent first pass.** Two sessions with no shared context and no tools: one Anthropic via `claude -p`, one OpenAI via `codex exec` (REQ-31). A third adjudicating session proposes a label where they disagree, and records why.
   - **Human review of every case.** The user confirms or changes each proposed label; there is no bulk accept, because pre-labels only order the work (`labeling-cases.md:17`). Before review, labels are recorded as `agent_reviewed`; after it, as `human` (`by: user`, plus the agent proposal). The latest human label wins.
   - **Timing.** Building eval data (labels only, with no Jev-versus-label comparison) can proceed now. Any comparison or threshold, and the test set of an experiment, is used only under a council-signed preregistration (`AGENTS.md` experiment rules 1 and 3). Product dispositions and maintainer triage of `would_block` stay human (design §6, §7.3).

Still open (design §15): 1 study-export consent, 4 `pull-requests: read`, 5 inertia framing. §15.2 (X8) leaves Scout with the re-scope.

## 7. How to resume

Read first: `AGENTS.md`, this file, `docs/scout-design.md` (§§1, 3, 7, 8.5, 10–11, 15), `SCOUT_REBUILD_REQUIREMENTS.md`, `docs/design-reviews/2026-10-21-e1b-v2/round-5.md`, then `inertia-bench/docs/design/05-harness-design.md` §9.

Commands (run from repo root):
- First time on a machine: `git submodule update --init`, then `scripts/bootstrap-tools.sh` (≈ 40 s fresh, ≈ 1 s when nothing changed), then `source .tools/activate` (undo with `scout_deactivate`). This installs the pinned toolchain into git-ignored `.tools/`: uv, Python 3.11.11 venv with Semgrep, Ruff and lizard; Node 22.18.0 with pnpm and jscpd; Gitleaks; and the patched Clawpatch bundle from the `third_party/clawpatch` submodule. Activation also sets `SCOUT_CLAWPATCH_BINARY` and `GOTOOLCHAIN=go1.25.6`. Pins: `toolchain/versions.env` (single source), `toolchain/checksums.sha256`, `toolchain/python-requirements.lock`, `toolchain/node/package-lock.json`. `--relock` regenerates the locks. `go test ./internal/toolchain` fails on drift between these pins and the Go catalogs or `clawpatch.Revision`.
- `go build ./... && go test ./...` — run with the toolchain activated; baseline is green.
- `shasum -a 256 docs/scout-design.md` — must print `7ff21a7f…` until a new council round adopts a change.
- Clawpatch upgrade: move the submodule to the new commit and update `clawpatch.Revision`/`Version` and `CLAWPATCH_VERSION` together. This needs a compatibility review (CP07 pattern, design R9). The parity test blocks a partial bump.
- Jev or study runs: `infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- <command>` (`AGENTS.md:56-58`); presence checks only.
- Rules to keep: council sign-off before any design change or experiment; ledgers updated when a source/hypothesis is used; never print secrets; never touch `.codegraph`; no git commits unless the user asks.
