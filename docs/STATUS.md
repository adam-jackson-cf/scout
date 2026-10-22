# Scout — status and handoff (2026-10-22)

Written for a fresh session. Every claim cites a file in this repo unless marked `[INFERENCE]` or `[UNVERIFIED]`. The repo has no commits; all 14 top-level entries are untracked (`git status --short`: `.agents/ .claude/ .github/ .gitignore AGENTS.md CLAUDE.md SCOUT_REBUILD_REQUIREMENTS.md cmd/ docs/ evaluation/ go.mod go.sum internal/ scripts/`; `.codegraph/` and `.scout/` are git-ignored, `.gitignore:1-2`). Never touch `.codegraph` (`AGENTS.md:54`).

## 1. Current state (verified 2026-10-22)

**Design: done and signed off.** `docs/scout-design.md` is revision 2.1; `shasum -a 256` = `7ff21a7fba70a5e3af7a735b0aa40f1ad654d17714df8c3549da6ebcf89b89c5` (matches the sign-off digest in `docs/design-reviews/2026-10-21-e1b-v2/round-5.md:40`). Council rounds 1–5 are in that directory (`round-1..5.md`, responses, `proposed.diff`, `scout-design.base.md`). v1 is `docs/archive/scout-design-v1.md`.

**Code: the v1 Go implementation only; the v2 design is not started.** `go build ./...` and `go test ./...` pass (run 2026-10-22; `internal/{analyzers,clawpatch,contextual,inventory,issues,state,web}` and `cmd/scout` have tests; `domain`, `engine`, `eval`, `codex` have none). ≈13.8 k lines under `cmd/` + `internal/`.

| Exists (v1, kept/rewritten per design §10.2) | Evidence |
|---|---|
| CLI `readiness assess issues issue mapping coverage triage approve verify serve` | `cmd/scout/main.go:90-110` |
| Inventory/plan/diff; analyzers (Semgrep, Gitleaks, Ruff adapters; catalogs; architecture catalog returns nil) | `internal/inventory/`, `internal/analyzers/` |
| Clawpatch client pinned **0.7.3** / `734f81bc…`, 900 s map timeout, sanitising review patch | `internal/clawpatch/client.go:19-22`, `scripts/build-clawpatch.sh`, `scripts/clawpatch-safe.patch` |
| SQLite state, issues, progress; read-only web | `internal/state/`, `internal/web/web.go` |
| Contextual LLM client (generic JSON/HTTP endpoint from `SCOUT_PROVIDER_*` env) | `internal/contextual/client.go`, `cmd/scout/main.go:128-130` |
| Self-eval corpus replay (v1 corpus: 9 true / 13 FP) | `internal/eval/`, `evaluation/{reference-v1,outcomes-v1,outcomes-v2,scout-requests-v1,scout-requests-v2}.json`, `evaluation/replay.go` |
| Workflows: `scout-nightly.yml` (237 lines, builds `./cmd/scout` from the checkout; passes `OPENAI_API_KEY`/`ANTHROPIC_API_KEY` secrets at `:172-173`), `release.yml` | `.github/workflows/` |

| Design-v2 components that do **not** exist | Design ref |
|---|---|
| `internal/{measures,history,artifacts,guardrails,decisions,baseline,selfeval}`; mapping areas/refresh trigger; `scout init/diff/export/import/signals/trends/annotate/selfeval`; `.scout/policy.yaml`; `scout-guardian.yml`; `docs/scout-product-contract.md`, `docs/scout-adapter-contracts.md`, `docs/evidence/` | §3.1, §10.2–10.3, §11 |

**P0/P1 defects still present** (design §3.3, §10.1; re-confirmed by grep): `CommitRunJob` calls `loadPlanTx` per job commit (`internal/state/progress.go:129,331-333`); issue matching requires feature intersection (`internal/issues/correlate.go:188,214-217,323`); `ApprovalRef`/`ProposedAction` (approval machinery the design cuts) remain (`internal/domain/types.go:892-893`, `internal/state/issues.go:457-464,510-511,570`). `internal/codex/exec.go` still present (design §10.2 says cut).

## 2. Decisions in force

- Design revision 2.1 is authoritative; the Appendix decisions (map refresh, `scout/state` branch, allow-list, Jev on from launch, tier-1 = Go/Python/TypeScript, spec cutover, Clawpatch 0.7.3) are not reopened (`docs/scout-design.md:409-417`). Five user decisions remain open (`:397-403`), see §6.
- Scout is a **discovery-only CI guardian**; no remediation, no inertia score, no claim that agents degrade a repo (`docs/scout-design.md:9-19`). All inertia measures are `descriptive` (`:185`).
- Design changes need innovation-council sign-off with zero issues, rounds recorded under `docs/design-reviews/<date>-<topic>/` (`AGENTS.md:30-38`). Experiments likewise, preregistered, frozen after sign-off (`AGENTS.md:3-28`). Design changes may cite only signed-off experiments (`AGENTS.md:37`).
- Ledgers must be updated when a source/hypothesis is used (`AGENTS.md:40-50`): `docs/research-ledger.md` (R-001…R-035), `docs/hypothesis-ledger.md`.
- Secrets only via Infisical per command; never print/export/commit (`AGENTS.md:52-61`).
- **User rules newly in force (this session, from the task brief; not yet in `AGENTS.md`):** OpenRouter only for TypeSafe Jev; all other model use via local subscription CLIs (`claude -p`, `codex exec`) with Anthropic/OpenAI models only. See §6 conflicts.

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
| 0 | Decide the §6 conflicts (esp. model-use policy, subject interface, spec cutover timing) and the five open design decisions | none | User | Answers recorded; any design change goes through a council round |
| 1 | **Phase 1** — fix `CommitRunJob` persistence, rewrite issue identity/anchoring, offline heuristic map, drop approval machinery, migration | #0 for P-perf target (§15.3) | Impl | Design §10.1 + §11 Phase 1 (flat-per-commit benchmark, durable ID across line shift/rename, `evidence_drift`, `regressed`, heuristic map offline, fastapi assess < 10 min) |
| 2 | **Phase 2** — differential guardian, `scout/state` branch, `init`/consent, `decisions` adapter with deterministic fallback, `scout-guardian.yml` (replace nightly workflow built from checkout) | #1 | Impl | §11 Phase 2 (PR A–D scenarios, CAS publish, byte-identical import) |
| 3 | **Phase 3** — measures, history, artifacts, guardrails for Go/Python/TS | #2 (can start `measures` earlier; independent of identity) | Impl | §11 Phase 3, incl. reproducing `active_hours` in `evaluation/change-cost-v2/panel-final.csv.gz` within 1e-9 h |
| 3b | **Subject interface** (`capabilities`/`measure`) for inertia-bench, offline/deterministic, snapshot-only measures | inertia-bench design signed off; #3 `measures` exist; design change signed off by council (the interface is not in design 2.1) | Council → Impl | Passes bench conformance (§4 above); byte-identical double run |
| 4 | **Phase 4** — mapping refresh triggers + areas (X2) | #1, #3 scopes | Impl | §11 Phase 4 |
| 5 | **Phase 5** — trends, exploratory measures, rubric lane, self-evaluation (X1, X3–X5, X9, X10), blind-label collection | #2–#4; ≥100 blind labels for X1 | Impl | §11 Phase 5; X1 decision rule recorded (§13) |
| 6 | **Phase 6** — study export, spec cutover (archive `SCOUT_REBUILD_REQUIREMENTS.md`, write `docs/scout-product-contract.md`, `scout-adapter-contracts.md`, `docs/evidence/`), web inertia panels, release, dogfood in `guardian` | #5, §15.1 consent decision | Impl | §11 Phase 6 |
| 7 | Update `AGENTS.md` for: model-use rule, research moved to inertia-bench (which experiment rules still apply to Scout), Infisical note for `claude -p`/`codex exec` | #0 | User approves; Impl edits | `AGENTS.md` consistent with §2 |
| 8 | **Candidate reference (not a ledger entry yet):** MOSS, arXiv:2610.26902 (Nesen et al., published 2026-10-16; verified abstract: production LLM-judge monitoring loop with targeted sampling, strategy-aware prompting, structured alerting; a single judge/score scale/threshold transferred from a 498-pair recommendation benchmark to a 466-pair search benchmark with only a surface-specific prompt change). Relevance: Scout persistent/nightly mode (§7.1, §8.4 trend flags, 5 % audit sample §6) and Jev triage thresholds (X1, §13: are thresholds transferable across rules/repos?) | read the full paper before use | Impl/User | If it informs a decision or hypothesis, add R-036 per `AGENTS.md:42-45` at that time |
| 9 | Repository hygiene: first commit (user's call), `.gitignore` covers `.scout/`, `.codegraph/` only — decide on `evaluation/**/__pycache__/` and bulky `evaluation/*/raw|data` | none | User | Decision recorded |

## 6. Open questions and conflicts needing the user

1. **Model-use policy vs current code/CI.** Policy: OpenRouter only for Jev; other models only via `claude -p`/`codex exec`, Anthropic/OpenAI only. Current: Scout's contextual client is a vendor-neutral JSON/HTTP endpoint (`cmd/scout/main.go:129`, `internal/contextual/client.go:31-50`); `scout-nightly.yml:172-173` passes OpenAI/Anthropic API-key secrets; Clawpatch provider defaults to `codex`. Subscription CLIs are not available on GitHub-hosted CI runners `[INFERENCE]`. Decide: (a) CI contextual/Clawpatch LLM lanes are local-only; (b) a CLI-subprocess provider adapter replaces the HTTP adapter; (c) API keys allowed in CI as an exception. Design §6 already makes reasoning LLMs opt-in, so (a) needs no design change; (b)/(c) do.
2. **Historical non-Jev vendor use.** E1b-v2 annotators A/B/C were `anthropic/claude-sonnet-4.5`, `openai/gpt-4.1`, `google/gemini-2.5-pro` through OpenRouter (`evaluation/change-cost-v2/annotate.py:7-9,27-28`); Gemini violates the new rule. The study is signed-off and frozen (`AGENTS.md:22-24`); it stays as-is, but cannot be reproduced under the new rule. Confirm it is grandfathered.
3. **Subject interface is not in design 2.1.** Needs a council-reviewed design change (`AGENTS.md:30-34`) before implementation; also settle where the adapter lives (bench says outside its repo; here or a third repo?).
4. **Research duties in design 2.1 now belong to the bench:** §8.5 (a)–(d), X7, X8, §15.2 and parts of §8.6 (pooled export). Design wins until revised; a council round should re-scope them.
5. **Spec-cutover timing.** Design §10.3/Phase 6 archive `SCOUT_REBUILD_REQUIREMENTS.md`; this session rewrote it as the live product spec instead. Keep it live until Phase 6, then archive.
6. Open design decisions (`docs/scout-design.md:399-403`): study-export consent; X8 (now bench territory); performance target; `pull-requests: read`; inertia framing (primary vs secondary family).
7. `AGENTS.md` experiment rules reference `evaluation/` studies "that produce evidence for Scout's design"; confirm they apply to Scout-side product experiments (X1–X6, X9, X10) but not to bench studies.

## 7. How to resume

Read first: `AGENTS.md`, this file, `docs/scout-design.md` (§§1, 3, 7, 8.5, 10–11, 15), `SCOUT_REBUILD_REQUIREMENTS.md`, `docs/design-reviews/2026-10-21-e1b-v2/round-5.md`, then `inertia-bench/docs/design/05-harness-design.md` §9.

Commands (run from repo root):
- `go build ./... && go test ./...` — baseline is green.
- `shasum -a 256 docs/scout-design.md` — must print `7ff21a7f…` until a new council round adopts a change.
- Phase 1 start points: `internal/state/progress.go:117-200`, `internal/issues/correlate.go:129,214-217,326-340`, `internal/state/issue_observations.go:94-102`, `internal/state/issues.go:455-461`.
- Clawpatch bundle: `scripts/build-clawpatch.sh` (pnpm 11.25.0).
- Jev or study runs: `infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- <command>` (`AGENTS.md:56-58`); presence checks only.
- Rules to keep: council sign-off before any design change or experiment; ledgers updated when a source/hypothesis is used; never print secrets; never touch `.codegraph`; no git commits unless the user asks.
