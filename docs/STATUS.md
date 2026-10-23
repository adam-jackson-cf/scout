# Scout — status and handoff (2026-10-23)

Written for a fresh session. Every claim cites a file in this repo unless marked `[INFERENCE]` or `[UNVERIFIED]`. History: `git log` (initial import `41ea1eb`, Phase 1 `945ac4d`, toolchain `1235132`, Phase 2 part `fccc137`, Phase 3 part `4bab5c0`, Gitleaks batching `36d1023`). `.codegraph/`, `.scout/`, `.tools/`, `__pycache__/` and `*.pyc` are git-ignored. Never touch `.codegraph` (`AGENTS.md`).

## 1. Current state (verified 2026-10-23)

**Design: revision 2.2 adopted 2026-10-23.** `docs/scout-design.md` SHA-256 is `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4`. That equals the digest signed off with zero issues in `docs/design-reviews/2026-10-22-rescope/round-6.md` after six rounds (25 → 9 → 12 → 7 → 2 → 0 issues). Revision 2.2 does the product-only re-scope:
- C1: reasoning-LLM lanes run only through `claude -p`/`codex exec` (Clawpatch bound to `codex`/`claude` in host auth, no API keys);
- C2: research moves to inertia-bench;
- C3: §8.7 subject-interface principles;
- C4: the performance bar is decided;
- C5: E1b-v2 is grandfathered;
- C6: §15 renumbered;
- C7a: the Jev adapter design (version key, failure handling, primitives, audit principles, labelling protocol).

C7b, the experiment estimators and samplers, is carried forward as binding input to the X1/X3/X4/X5/X10 preregistrations (`round-3-response.md`, "C7b carry-forward"). Companion edits adopted with it: `docs/hypothesis-ledger.md` (HA2 → inertia-bench; H\*/HA\*/HQ\* frozen), research ledger R-038, the requirements rows synced in rounds 1–5, and §5 row 8. Rev 2.1 rounds are in `docs/design-reviews/2026-10-21-e1b-v2/`; v1 is `docs/archive/scout-design-v1.md`.

**Code.** `go build ./...`, `go vet ./...` and `go test ./...` pass with the toolchain activated.
- **Phase 1 is done (`945ac4d`).**
  - `plan_jobs` (schema v4) makes per-job commits flat in job count. `TestCommitRunJobCostIsFlatInJobCount`: 5,000 jobs against a 5 MB plan.
  - Design §3.3 issue identity: ULID IDs, anchor ladder, `evidence_drift`, `regressed`, rename and delete lifecycle from the latest recorded inventory, secret digests, and a schema v5 migration.
  - The heuristic map runs offline.
  - Scanners are batched per adapter (`36d1023` adds Gitleaks batching with per-file trace proof).
  - Approval machinery and `internal/codex` are removed.
- **Phase 2 is done** (`fccc137` plus the Phase 2 completion commit):
  - `scout export/import` (state/v1, byte-identical round trip) and `scout state publish|restore` on `scout/state`, with expected-parent compare-and-swap; publishing requires `--trusted`.
  - Lane registry and health; exit codes 0/1/2/3; `scout diff --base` attribution against the published baseline (`comparison: baseline`, or `recomputed` with a reason).
  - Report/guardian profiles; `.scout/policy.yaml` read from the base ref; shadow window; maintainer-only overrides.
  - `would_block` (schema v6); `guardrails.SUPPRESS.added.unscoped`.
  - `internal/decisions`: typed `noul|choice|score` questions with the §6 version key; OpenRouter `alpha/decisions` client; resolved-model pin (mismatch → `incomplete`); retries for 429/529 only; the 2,000-call cap in §6 order; the deterministic baseline; the audit draw with stored inclusion probability. Schema v8 adds the annotation, ranking, audit and disposition-context tables.
  - Ranking hook (`engine.RankingOf`): Jev only when consented, keyed, egress allowed and not a fork. Otherwise the run reports `ranking: deterministic (<reason>)`, and a fork is reported first.
  - Model-CLI adapter (`internal/modelcli`, `internal/contextual`): tool-less `claude -p` or `codex exec` with a credential-scrubbed environment. The `SCOUT_PROVIDER_*` HTTP client is removed.
  - Clawpatch runs only with `--provider codex|claude` and `CLAWPATCH_CLAUDE_AUTH_CONTEXT=host`. The `contextual` and `clawpatch-review` lanes run only when opted in through `model_cli.lanes`; otherwise they report `skipped (not opted in)` or `skipped (no model CLI)`, and the exit code is unchanged.
  - `scout init` records consent: profile `report`, `decisions`, `model_cli`. It is idempotent and keeps activations and comments.
  - `.github/workflows/scout-guardian.yml` is the reusable workflow. It installs a pinned, sha256-verified release, never builds from the PR, and only the trusted job sees the Jev secret and publishes. `.github/actions/scout/` is the composite action. `scout-nightly.yml` is now a dogfood caller; AWS, S3 and the model API keys are removed. `release.yml` writes basename checksums, so its `shasum -c` step works.
  - **§11 Phase 2 scenario passed on 2026-10-23** (`/tmp/scout-p2/phase2-scenario.sh` plus `report.sh`, throwaway):
    - PR A exits 0 with 2 `would_block` (PAT and `SUPPRESS`).
    - After activation, the identical PR exits 1 with only `SUPPRESS` active, and the PAT is still `would_block`.
    - PR B exits 0.
    - Fork PR C reports `deterministic (fork)`, and its untrusted publish exits 3.
    - PR D's `scout-override` from a non-maintainer is rejected and logged; the maintainer control exits 0.
    - A tampered baseline digest gives `comparison: recomputed`.
    - Two concurrent publishes give 2 commits (one retried via CAS).
    - Export → import → export is byte-identical.
    - Moving the activation's version back returns the check to shadow (exit 0).
  - **Deviation from the §11 text:** the digest-mismatch PR is attributed `introduced`, not `unknown`. The paired recompute rescans the merge base under head's digests, which meets §7.3's definition of `introduced`. `unknown` is produced only when no comparable base exists (`gate.go:178`). Any change to the §11 wording needs a council round (§6).
  - Workflow assumptions: `release-repo` defaults to `inflight/scout` because there is no git remote; the dogfood caller pins `v0.1.0`, which is not tagged yet; `refresh-map` forces `--full` until Phase 4; SARIF is uploaded only if present.
- **Phase 3, part done (`4bab5c0`):**
  - Signal store (schema v7) and `scout signals`; `scout.cc`, `complex_burden`, `burden_density`, `dup_pct` (lizard and jscpd); `cycle_mass` (Go/Python/TS import graph); §8.3 history measures in calendar quarters.
  - `ActiveHours` reproduces E1b `active_hours` for all 598 nest/webpack panel PRs (max error 8.9e-16 h).
  - **Not started:** artifacts (LCOV/Cobertura/JUnit, CRAP), the remaining guardrail mechanisms, PR-metadata ingestion, `delta` rules, `readiness --matrix`, the 9/13 corpus regression check.
- **Acceptance figures:** fastapi `d666ccb`, every lane, no egress. 455 s after Phase 1, 537 s after Phase 3's lanes, 143 s after Gitleaks batching, **125 s** after Phase 2 (2026-10-23); 74 issues, exit 0 each time. With no policy, the contextual and clawpatch-review lanes read `skipped (not opted in)` and decisions reads `skipped (decisions not enabled)`. Phase 1 scenario: `/tmp/scout-p1/scenario.sh`. Phase 2 scenario: `/tmp/scout-p2/phase2-scenario.sh`.
- **Known gaps:**
  - If agent mapping succeeds with zero features, heuristic output is not re-merged (Phase 4).
  - An issue left unmatched during a rename keeps its old anchor path.
  - Policy `version` strings for Gitleaks checks are the long rule-pack version.

**Toolchain.** Pinned, project-local, in `.tools/` (§7). Clawpatch is a submodule at `third_party/clawpatch` (`734f81bc…`, v0.7.3). Pins are in `toolchain/`, and `internal/toolchain` enforces parity with the Go code. A fresh clone bootstraps and passes every test in about 44 s.

**Jev labels.** All 52 E2 triage cases carry human labels: the user reviewed agent proposals on 2026-10-23 (`evaluation/jev-triage-labels-v1/`; real_defect 28, false_positive 23, cannot_tell 1). No Jev-versus-label comparison has been run; that needs a council-signed preregistration.

## 2. Decisions in force

- Design revision 2.2 is authoritative (`docs/scout-design.md`, `f81b2601…`). Appendix decisions 1–7 (2026-10-19) and 8–15 (2026-10-22/23) are not reopened. Open user decisions: design §15 items 1–3 (study-export consent, `pull-requests: read`, inertia framing), and the remaining `jev-actions-review.md` §D questions (minimums, costs, Jev version pin), which feed the experiment preregistrations.
- Scout is a **discovery-only CI guardian**; no remediation, no inertia score, no claim that agents degrade a repo (`docs/scout-design.md:9-19`). All inertia measures are `descriptive` (`:185`).
- Design changes need innovation-council sign-off with zero issues, rounds recorded under `docs/design-reviews/<date>-<topic>/` (`AGENTS.md:30-38`). Experiments likewise, preregistered, frozen after sign-off (`AGENTS.md:3-28`). Design changes may cite only signed-off experiments (`AGENTS.md:37`).
- Ledgers must be updated when a source/hypothesis is used (`AGENTS.md` ledger rules): `docs/research-ledger.md` (R-001…R-038), `docs/hypothesis-ledger.md` (HD\* live; H\*/HA\*/HQ\* frozen).
- Secrets only via Infisical per command; never print/export/commit (`AGENTS.md:52-61`).
- **User decisions of 2026-10-22/23** are recorded in §6 (R-036). The design-affecting ones are adopted in revision 2.2 (Appendix 8–15).

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
| 0 | ~~Decide the §6 conflicts; council re-scope round~~ **done**: rev 2.2 adopted 2026-10-23 (round 6, zero issues) | — | — | — |
| 1 | ~~**Phase 1**~~ **done** (`945ac4d`, `36d1023`) | — | — | fastapi 143 s; §1 |
| 2 | ~~**Phase 2**~~ **done** 2026-10-23: §11 scenario passed (§1); one wording deviation (digest mismatch → `introduced` after a valid recompute) | — | — | §11 Phase 2 |
| 3 | **Phase 3 remainder**: artifacts, CRAP, remaining guardrails (all three languages), PR metadata, `delta` rules, `readiness --matrix`, corpus regression, TS `paths` acceptance | #2 for PR metadata in CI | Impl | §11 Phase 3 |
| 3b | **Subject interface** principles adopted (rev 2.2 §8.7); exact schema needs its own council round after inertia-bench signs off | bench design signed off; #3 measures | Council → Impl | §11 Phase 3b |
| 4 | **Phase 4** — mapping refresh triggers + areas (X2) | #1, #3 scopes | Impl | §11 Phase 4 |
| 5 | **Phase 5** — trends, exploratory measures, rubric lane, self-evaluation (X1, X3–X5, X9, X10), blind-label collection | #2–#4; ≥100 blind labels for X1 | Impl | §11 Phase 5; X1 decision rule recorded (§13) |
| 6 | **Phase 6** — study export, spec cutover (archive `SCOUT_REBUILD_REQUIREMENTS.md`, write `docs/scout-product-contract.md`, `scout-adapter-contracts.md`, `docs/evidence/`), web inertia panels, release, dogfood in `guardian` | #5, §15.1 consent decision | Impl | §11 Phase 6 |
| 7 | ~~Update `AGENTS.md`~~ **done 2026-10-23** (§6.6) | — | — | — |
| 8 | **Candidate reference (not a ledger entry yet):** MOSS, arXiv:2610.26902 (Nesen et al., published 2026-10-16; verified abstract: production LLM-judge monitoring loop with targeted sampling, strategy-aware prompting, structured alerting; a single judge/score scale/threshold transferred from a 498-pair recommendation benchmark to a 466-pair search benchmark with only a surface-specific prompt change). Relevance: Scout persistent/nightly mode (§7.1, §8.4 trend flags, 5 % audit sample §6) and Jev triage thresholds (X1, §13: are thresholds transferable across rules/repos?) | read the full paper before use | Impl/User | If it informs a decision or hypothesis, add the next free `R-NNN` (R-039) per `AGENTS.md:46-49` at that time |
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
- `shasum -a 256 docs/scout-design.md` — must print `f81b2601…` (rev 2.2) until a new council round adopts a change.
- Clawpatch upgrade: move the submodule to the new commit and update `clawpatch.Revision`/`Version` and `CLAWPATCH_VERSION` together. This needs a compatibility review (CP07 pattern, design R9). The parity test blocks a partial bump.
- Jev or study runs: `infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- <command>` (`AGENTS.md:56-58`); presence checks only.
- Rules to keep: council sign-off before any design change or experiment; ledgers updated when a source/hypothesis is used; never print secrets; never touch `.codegraph`; commits in logical pieces with a staged Gitleaks scan first (`gitleaks git --pre-commit --staged --no-banner --redact .`).
