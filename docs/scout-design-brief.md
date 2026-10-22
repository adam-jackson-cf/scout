# Scout design brief for the innovation council

## 1. Aim (from the user; authoritative)

Scout is a **discovery-only guardian** that sits in an individual project's CI pipeline (and runs nightly / locally) and **surfaces issues for others to decide how to deal with**. It does not fix code. Its primary focus is **code quality as the driver of long-term codebase inertia**; architecture, security and performance remain families but quality/inertia is the priority. Findings are persisted in a durable, triageable store (stable IDs, dispositions, evidence) so the same issue is tracked across runs. Feature mapping uses Clawpatch; the user wants **both** Clawpatch mappers (deterministic heuristics + agent exploration) always used so areas the deterministic mappers miss still get a definition issues can attach to. The user wants cheap typed decision models (TypeSafe Jev, via OpenRouter) considered as replacements for LLM judgment where they fit. The user likes SlopCodeBench's "solve rate"-style outcome metric framing (without hidden tests) and SCB's quality signals.

## 2. Current state of Scout (observed)

Repo: /Users/adamjackson/Projects/inflight/scout (Go, 13,777 LOC, no git commits yet). `go build/vet/test ./...` pass (8 suites; engine/eval/codex/domain untested). Requirements: `SCOUT_REBUILD_REQUIREMENTS.md` (62 KB).

| Area | State |
|---|---|
| Change detection & planning (`internal/inventory` scan/diff/plan) | Real: content hashes incl. dirty/untracked, renames, explicit dependency closure, per-check cache keys incl. tool/rule/model/mapper versions. |
| Durable issues (`internal/state`, `internal/issues/correlate.go`) | Real: SQLite WAL, stable issue IDs, states open/investigating/accepted/false-positive/won't-fix/pending-verification/resolved/unverified/regressed, audited triage, `verify` guards closure. |
| Clawpatch (`internal/clawpatch`, `scripts/build-clawpatch.sh` + `clawpatch-safe.patch`) | Real: builds pinned 0.4.0 rev 587b9a1 + redaction patch, imports feature IDs, per-feature `review`. Uses `map --source auto` (`client.go:178`), so agent mapping runs only when heuristic map is "weak". |
| Contextual LLM review per feature×category (`internal/contextual`) | Built; never completed a run (provider egress never authorised). |
| CLI / web | `assess, issues, issue, triage, approve, verify, coverage, mapping, readiness, serve` (read-only web). JSON + exit codes 0/1/2/3. |
| Analyzers | Security: Gitleaks + 4 Semgrep rules. Quality: Ruff C90 (Python only). Performance: Ruff PERF + 1 CSS rule. **Architecture: none** (`catalog_architecture.go` returns nil). |
| Evaluation | 7 pinned repos × 4 families; 9 validated true issues, 13 FP, all 28 entries `incomplete`. Custom rules were written to match the exact corpus issues (`python-multipart-cve-2024-24762`, `current-script-dom-clobbering`, `css-height-transition`) → the 9/9 v2 match is overfit, not general recall. |
| CI | `.github/workflows/scout-nightly.yml`: push-to-main/nightly/dispatch only, assesses Scout itself, needs S3 + AWS OIDC baseline, LLM egress off by default ⇒ every normal run exits 2 (partial) and fails. No PR trigger, no reusable workflow/action, gate is on all unresolved issues not new/regressed. |
| Dead code | `internal/codex/exec.go` (sandboxed Codex CLI runner) has no callers. |

## 3. Comparison sources studied

### Atlas (xSolutions365/architectural-harness @0652b9f; already cited in requirements as "Architectural Harness")
- Org-wide multi-repo profiler (Python): tech-profiling/EOL, security-scan (Semgrep/Gitleaks/Checkov/OSV), commit-team (ownership), provisioned-resources (IaC), estate joins, repo-scorecard. Static Astro dashboard over committed typed JSON "signals".
- Incrementality: repo re-pull by pushed_at; per-analyzer cache keyed by **git HEAD only** (`core/cache.py:97-110`) — misses dirty edits and analyzer upgrades. Scout's planner is strictly finer.
- **No finding comparison** across runs (no identity, no new/fixed classification, no triage); snapshots committed to git and diffed. No CI.
- Worth taking: committed-to-git results as a baseline store (vs S3); commit-team ownership signals (deterministic input for architecture "team ownership"); typed signal/view contracts; `plans/07-analyzer-building-skill.md` LLM-judge + human-gate analyzer refinement loop. Atlas is estate-level, Scout repo-level → complementary (Atlas could consume Scout stores).

### factory-x-loops (/Users/adamjackson/Projects/inflight/factory-x-loops/factory-x-loops)
- Agentic-workflow benchmark; remediation-leaning. Transferable:
  - **"Against the base" PR rules** (`docs/references/verify-gate.md:33-93`): `CONFIG.ignores-added`, `CONFIG.strictness-lowered`, `SUPPRESS` (new `noqa`/`fmt: off`/`no_type_check`), `TESTQ.constant-assert/literal-assert/no-assertion/swallowed-assert/skipped/delegating-test/generated-test`, `GATE.coverage-pragma-added`, `TEST.changed-line-uncovered`, `BASE.test-failed/missing/skipped`. Deterministic guardrail-erosion findings. Python-only today (raidar `python-guardrails@1`).
  - Scoring discipline (`docs/references/scoring.md`): `count/attempts` with Wilson 95% intervals; unmeasured ≠ failed; refuse to pool incomparable runs.
  - **Assurance matrix**: claim vs independent oracle → assured / false assurance / flagged / over-caution; flag precision/recall. Apply to Scout's own "covered/clean" claims.
  - Fault catalog record shape (FCR-NNN, detection signal, evidence, hypothesis, measurable resolution criterion, status open/targeted/mitigated/resolved/reopened/not-reproducible) → template for Scout's gap→rule records.

### SlopCodeBench (/Users/adamjackson/Projects/inflight/slop-code-bench) — not previously mined
- Canonical components in [0,1], higher better (code: `src/slop_code/metrics/scoring/`):
  - **Erosion** `1 − Σ_{CC>10} CC·√SLOC / Σ CC·√SLOC` (`production_quality.py:779-803`)
  - **Verbosity** `1 − |pinned-Ruff-flagged lines ∪ cloned lines| / SLOC` (`production_quality.py:669-703`)
  - **Architecture** `1 − cyclic import-edge mass / total edge mass` (`production_graph.py:222-235`)
  - **Rework** `1 − repeated (re-changed) lines / changed lines` via symbol lineage (`lineage.py:121-407`)
  - **Regression breadth** `1 − changed symbols hit by newly broken tests / changed symbols`
  - Inertia `I = .15V + .15E + .20H + .30R + .20G`; score `S = 100·C·(.70+.30·I)`.
  - Supporting: cc_concentration (Gini), nesting depth, lint_per_loc, cloned_pct, single-use functions, trivial wrappers, unused variables, propagation cost, dependency entropy, churn ratio.
- **45-pattern LLM-judge rubric** (`configs/rubrics/llm_judge.jsonl`: name, category, type verbosity|erosion, description, positive/negative indicators), e.g. silent_fallbacks, retry_loops_hiding_failures, validation_theater, error_message_destruction, stub_implementations, strategy_never_varies, stringly_typed_logic, narration_comments, delegation_only_class.
- **Claims ladder** (`docs/metrics/measurement-validity.md`): correctness / descriptive quality / predictive quality / causal inertia; report a measurement only with the claim its evidence supports. All SCB quality components are currently *descriptive*; predictive validation preregistered (`experiment_analysis/011026-predictive-validity-preregistration.md`) but no valid cohort result yet.
- Metric code is Python-only (`metrics/languages/python`).

### TypeSafe Jev (decision models)
- Typed noul/choice/score answers, no generation. Measured locally (`~/.omp/jev-lab/fastino-bench/results/REPORT.md`, 2026-10-02): Jev $0.10/1k decisions, p50 0.26 s; reasoning set incl. code tracing 75% (GLiDE 98% at 9× cost, p95 10 s). Jev 1.13 known weaknesses: indirection, numbers, large irrelevant state, adversarial content in state (`docs.typesafe.ai/model-jaggedness/jev-1.13.md`). OpenRouter now also resolves jev-1.14-20261014.
- Judgment inventory (where typed decisions fit): triage of deterministic signals (TP/FP/cannot-tell), observation→existing-issue matching (choice incl. none), §12.3 severity rubric dimensions (closed choices), coverage status after LLM review, unowned-file → existing feature assignment (choice incl. none), commit classification (fix/revert/feature) for history signals, SCB rubric patterns per changed function. Keep LLM: finding discovery, new feature grouping, FN discovery. Keep deterministic: eligibility, dependency closure, quote validation. Keep human: dispositions, approvals, rule promotion.
- Safety rule proposed: decision models **route/rank only**; never close or dismiss (adversarial code comments in a PR could steer them). Enforce resolved-model pinning per response; mismatch ⇒ unit incomplete + cache invalidation.

## 4. Ideas raised so far (proposals, not decisions)
1. Minimal guardian shape: PR trigger; gate only on **new or regressed** issues (pre-existing debt reported, not blocking); git-based baseline (dedicated branch or committed snapshot) instead of S3/OIDC; deterministic lanes by default, LLM lanes opt-in; reusable workflow/action for adopting repos.
2. Two record types: **signals** (descriptive measurements with per-feature history/trends: erosion, verbosity, architecture cycles, rework, churn, coupling, change-success) and **issues** (concrete mechanisms: rubric hit, guardrail weakened, new import cycle, tool finding). Issues may be raised from signal combinations only with stated evidence. No CI gating on inertia scores without predictive support.
3. "Solve rate" analogue without hidden tests: per-feature **change success rate** = changes with no follow-up fix/revert of the same lines within N days / changes, Wilson interval, unmeasured when too small. Git history as the oracle (SZZ-style), commit classification by Jev.
4. Reimplement SCB formulas in Go on multi-language tools (lizard for CC/SLOC, jscpd clones, per-language import graphs) rather than wrapping Python SCB.
5. PR-time lanes: guardrail-weakening ("against the base") rules ported per language; SCB rubric via Semgrep where structural, via Jev per changed function where semantic (≈$0.90 per 200 functions × 45 patterns, unvalidated).
6. Always run Clawpatch both mappers (`--source agent`); risk: agent-feature ID drift orphaning issue anchors; agent pass with zero features is a hard error in `agent` mode.
7. Replace corpus-specific custom rules with broad maintained rule packs; keep corpus as regression check.
8. Scout self-evaluation: add false-assurance rate and flag precision/recall for its "clean" claims, alongside Enaible-parity matching.
9. Read existing CI artifacts (coverage, JUnit) rather than running project tests (requirements forbid unauthorised test execution) → changed-line coverage, regression breadth only when artifacts exist.

## 5. Open contradictions / gaps
- Requirements treat four families equally; aim now prioritises quality/inertia.
- Requirements §10/§12.5 define only a JSON LLM provider adapter; decision models need a second adapter contract (typed questions, escalation threshold, version pin, cost budget).
- Requirements' pinned Clawpatch 0.4.0 vs installed 0.7.2+: upgrade cost per compatibility review.
- reference-v1.json `behavior` text leaks verdicts; evaluations must rebuild states from raw signals.
- Spec size (62 KB) vs delivered breadth: risk of over-specification.

## 6. Evidence gathered (this session)
All three under `/Users/adamjackson/Projects/inflight/scout/evaluation/`; each has RESULTS.md, scripts, raw data.

### E1 — Do SCB-style inertia signals predict later fixes on real repos? (`inertia-v1/`, preregistered)
- 7 pinned repos, 1,732 files; T0 = pinned − 365 d; outcome = log1p(fix commits touching file) in following year; baseline = size, prior commits/fixes/churn, age, authors; quality = erosion, cc_max, high-CC count, clone_pct, rework. Ridge, leave-one-repo-out, 2,000 bootstrap.
- **No predictive support.** Quality adds 1.18% relative MSE reduction, 95% CI [−0.62%, +3.24%] (needed ≥5%, lower>0); 6/7 repos improved. Y_any AUROC 0.642 → 0.623 (worse). Best single predictor: count of CC>10 functions, +1.91% [0.39%, 3.68%] — real but small. Nothing construct-invalid.
- Regex fix-commit classifier: precision 0.52, recall 0.92 on 100 hand labels (Renovate `fix(deps)`, docs "Fix typo" false positives) → history-based oracles need a better commit classifier (Jev candidate).
- Caveats: single labeller, single time split, eShopOnWeb has 0 fix commits in window, file-level not feature-level, fix count conflated with usage/centrality.
- Implication: inertia/quality scores stay **descriptive**; do not gate CI or set severity from them; prior history (churn, prior fixes) carries most of the predictive signal.

### E2 — Can Jev triage deterministic signals as real vs false positive? (`jev-triage-v1/`)
- 22 adjudicated reference signals (9 true, 13 FP), states rebuilt from raw tool signal + source excerpt at pinned revision (de-leaked). 12 held-out / 10 tuning, 3 reps, jev-1.13 and jev-1.14 (both resolve correctly).
- Choice v1 (pre-written): 17/22, held-out 10/12, AUROC ~0.95. Noul v2: 14–16/22, routes 4–6 to cannot_tell. Choice v3 (revised on tuning errors): 21/22, held-out 11/12, AUROC 0.99 (optimistic: author saw all adjudications).
- Claude Opus via `claude -p`: 21/22 (v1 criteria), ~$55–65 per 1k incl. CLI overhead, ~8–9 s p50. Jev: ~$0.05 per 1k, p50 ~0.3 s, 0/396 failures → ~1000× cheaper, ~30× faster at comparable accuracy on this tiny set.
- Weaknesses: 6/9 positives are the same mutable-Action pattern; unlabelled sample showed v1 inconsistency on that pattern (15/42 reps FP). n=22 → wide Wilson intervals.
- Implication: Jev viable as a **router/ranker** for triage queues; needs a larger blind label set; route-only, never close.
- Access: OPENROUTER key via `infisical run` from `~/.omp/jev-lab/fastino-bench`; scout dir has no Infisical project.

### E3 — Is always-agent Clawpatch mapping stable enough to anchor issues? (`map-stability-v1/`)
- Pinned Clawpatch 0.4.0 built; provider codex. fastapi (154 heuristic features) and axum (52).
- Heuristic IDs: Jaccard 1.00 across 3 fresh runs; unchanged by content-only edit; new file absorbed by existing group feature (fastapi).
- Agent IDs: fresh-run Jaccard 0.05–0.15 (fastapi), 0.13–0.34 (axum); only 9–29% survive same-state reruns; stale features accumulate (fastapi 187→296 features over 5 runs). Owning feature of 44–81% of files changes between fresh runs. A one-line content edit produced ~20 new agent IDs unrelated to the edit. Agent run 2–5 min vs ~1 s heuristic.
- `auto`: fastapi heuristic map is weak (9% owned) → agent would run; axum not weak (32%).
- **Implication: "always use both mappers" as-is would orphan issues.** Supported mitigations: anchor issues to file+symbol/content fingerprint with feature as secondary label; heuristic map canonical; agent only for unowned files and run rarely (map inputs changed), with Scout-side pinning of accepted agent features and re-attachment by file overlap; or Jev choice "assign unowned file to existing feature or none" as a cheaper, deterministic-ID alternative.

## 7. What the council must produce
A final design and planned approach for Scout that serves the aim in §1, decides the open questions below, and is grounded in §2–6. Expected output: product definition and non-goals; architecture (components, data model: signals vs issues vs coverage, anchors/identity); feature-mapping strategy given E3; analyzer lanes and priorities given the quality/inertia focus and E1; where decision models / LLMs / deterministic code / humans sit given E2; CI guardian contract (triggers, baseline storage, gate semantics, exit codes, adoption path for other repos); self-evaluation metrics (solve-rate-like, false assurance); what to keep, cut, or rewrite from current code and the 62 KB spec; phased plan with each phase ending in a runnable scenario and explicit acceptance evidence; top risks and the experiments still needed.

Open questions:
1. Signals vs issues model and how a descriptive signal becomes an issue without overclaiming (SCB claims ladder; E1 = no predictive support).
2. Feature mapping: heuristic-canonical + constrained agent vs always-both; identity/anchoring of issues.
3. Decision-model role and guardrails (route-only, version pinning, eval set growth).
4. Guardian gate: new/regressed only? which lanes may fail a PR (guardrail weakening, new import cycle, security), which only report?
5. Baseline storage for adopting repos (git branch / committed snapshot / artifact vs S3).
6. Language strategy: breadth (7 languages) vs depth first; multi-language quality tools (lizard, jscpd, import graphs, Semgrep packs).
7. Outcome metric for Scout's value ("solve rate" analogue) given noisy git oracles (classifier P=0.52).
8. Scope of the existing spec/code to keep vs simplify.
