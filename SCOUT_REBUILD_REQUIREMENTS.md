# Scout — product requirements (aligned to design revision 2.2)

Status: rewritten 2026-10-22, synced to revision 2.2 on 2026-10-23. `docs/scout-design.md` revision 2.2 (sha256 `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4`, council-signed in `docs/design-reviews/2026-10-22-rescope/round-6.md`) is authoritative: where this file and the design disagree, the design wins. Current implementation state and the ordered plan are in `docs/STATUS.md` (§1, §5). Design changes need council sign-off (`AGENTS.md` design rules). This file holds testable requirements only; mechanisms, formulas and phase scenarios live in the design and are cited, not copied. Per design §10.3 this file is archived at Phase 6 and replaced by `docs/scout-product-contract.md` and `docs/scout-adapter-contracts.md`.

Requirement IDs `REQ-n` are local to this file (not ledger IDs).

## 1. Product

- **REQ-1** Scout is a lightweight, extensible, **discovery-only CI guardian for a single repository**. It runs on pull requests, default-branch pushes, nightly and locally, and surfaces evidence-backed issues and descriptive measurements into a durable, triageable ledger (design §1, §7.1).
- **REQ-2** Four analyzer families — **architecture, security, code quality, performance** — with code quality primary and the others lower priority (design §1, §5.1). Language support is opt-in and minimal: tier 1 = Go, Python, TypeScript; CC/NLOC/duplication on all seven baseline languages; tools provisioned only for the detected stack (design §5.3, Appendix 5).
- **REQ-3** Two operating modes over the same state: **CI/nightly** (differential, baseline restored from the `scout/state` branch) and **persistent/local** (resumable, SQLite WAL) (design §3.1, §7.1–7.2).
- **REQ-4** Non-goals (design §1): no remediation or approval workflow; no execution of the target's tests; **no inertia score** as gate, severity, forecast or headline; no claim that agents degrade an adopter's repository.

## 2. Ledger: issues, identity, dispositions, evidence

- **REQ-5** Issues have an opaque durable ID (`issue:<family>:<ulid>`) anchored to source (rename-followed path, qualified symbol, rule key, normalised mechanism, evidence fingerprint). Line/byte positions and Clawpatch feature/agent IDs never enter identity (design §3.3).
- **REQ-6** Matching follows the design §3.3 ladder: line shifts and renames keep the ID; changed evidence is `evidence_drift` with disposition intact; distinct occurrences are distinct issues; ambiguous candidates stay open with `ambiguous_with` and are never auto-merged; a reappearing resolved issue is `regressed`.
- **REQ-7** Dispositions are audited and survive evidence drift; evidence is sanitised before persistence, logging, model requests or export; secrets match by digest so moves stay `pre-existing` (design §3.3, §6, §7.3).
- **REQ-8** Three record types: Signal (append-only, never gates), Issue, Coverage. Signals become issues only through a named, versioned rule with a `claim` (`delta`, `trend`, `policy`); no severity is derived from an aggregate (design §3.2).
- **REQ-9** Coverage is explicit: `inspected/reused/skipped/blocked/incomplete` with reasons; an incomplete or skipped unit is never reported clean and never closes an issue (design §3.2, §3.4, §7.3).
- **REQ-10** A migration re-anchors the v1 store and lists ambiguous rows (design §11 Phase 1).
- **REQ-11** A resolution is recorded only after a relevant successful re-check; absence from a partial/skipped run is not resolution (design §3.2 Coverage row; §7.3 attribution).

## 3. Feature mapping

- **REQ-12** Clawpatch **0.7.3** (`734f81bc…`) heuristic map is the canonical feature vocabulary; agent-mapper IDs never enter issue identity (design §4; E3 `evaluation/map-stability-v1/RESULTS.md:3`).
- **REQ-13** Per-repo mapping uses both mappers at onboarding and on debounced default-branch structural events (default 1 h); never on PR runs; `scout map --refresh` forces one. Scout-side **areas** absorb agent-ID churn; time series are keyed to heuristic features plus directory scopes and never re-key on area change (design §4, Appendix 1).
- **REQ-14** Unowned files are reported as a mapping-coverage gap but still receive every file-level lane and a directory scope; the heuristic map runs offline (design §4, §10.2).
- **REQ-15** Scout never runs `clawpatch fix` and never invokes the registry verifier (design §4 "Failure semantics"; CP07 `evaluation/clawpatch-0.7-compat/RESULTS.md:24`).

## 4. Analyzer lanes

- **REQ-16** Lanes are declared as `{family, trigger, mode, requires}` and report health (ran/reused/skipped/blocked/failed); lane health drives exit 2/3, findings exit 1 (design §3.4, §7.4).
- **REQ-17** Lane catalogue and launch languages follow design §5.1: `guardrails`, `measures`, `history`, `security` (Gitleaks exact ids, Semgrep packs), `artifacts`, `rubric`, `architecture-policy`, `performance`, `contextual`/`clawpatch-review` (opt-in). Findings need an observable mechanism and an evidence location; tool output absence is not proof of safety.
- **REQ-18** `scout readiness --matrix` prints the capability matrix (design §5.3); a missing optional tool never blocks a ready assessment, and JSON and exit status agree.
- **REQ-19** Fixture classes (positive, clean negative, boundary, redaction, plus attribution fixtures: line move, rename, version bump, config migration, tool upgrade, new file) exist per adopted check; the v1 corpus regression (9 true detected, 13 FP not flagged; `evaluation/reference-v1.json`) holds (design §7.3 Track A, §9.8).

## 5. CI guardian contract

- **REQ-20** Differential gate: head vs same-policy base; only `introduced`/`regressed` issues in **activated** allow-listed checks can block; `pre-existing` never blocks; `unknown` attribution reports with a banner (design §7.3).
- **REQ-21** Profiles `report` (adoption default; never fails on findings) and `guardian`. Allow-list = `docs/scout-allowlist.md`, eligible-not-active, no wildcards (design §7.3, Appendix 3).
- **REQ-22** Per-repo activation after a shadow window ≥ 14 days **and** ≥ 5 PR runs with comparable base (configurable), with `would_block` persisted and maintainer triage; auto-demote on ≥ 3 confirmed-unwarranted blocks in the trailing 20; version bump returns a check to shadow; rare security checks may activate `provisional` on track A alone (design §7.3).
- **REQ-23** Overrides accepted only from a maintainer, stored as dispositions, findings never deleted (design §7.3).
- **REQ-24** Exit codes `0 clean / 1 gated / 2 incomplete / 3 unavailable`, precedence 3 > 2 > 1 > 0; JSON always carries `gate`, `health`, `would_block`, `ranking`; ≤ 60-line Markdown summary; optional SARIF (design §7.4).
- **REQ-25** Trust boundary: binary and Clawpatch bundle from a pinned release (never built from the PR checkout); policy read from the base branch; no provider/decision secrets on fork PRs; only trusted default-branch runs publish state (design §7.2, §7.4).
- **REQ-26** Baseline = protected `scout/state` git branch (JSONL export/import, expected-parent compare-and-swap); SQLite and WAL never committed; S3 deferred (design §7.2, Appendix 2).
- **REQ-27** First adoption: `scout init` → consent → full default-branch run → baseline; nothing is `introduced` before a baseline exists (design §7.4).

## 6. Placement of judgment

- **REQ-28** Deterministic Go owns eligibility, cache keys, identity, gate, measurements, trend flags, redaction; Jev ranks/routes **on from launch** and only annotates (output is an `annotation` row; `issues` does not import `decisions`); reasoning LLMs are opt-in discovery; humans own dispositions, policy, activation (design §6, Appendix 4).
- **REQ-29** Jev guardrails: typed `noul|choice|score` questions with `question_id@<sha256 version key over instructions, criteria, state shape, option-order rule, resolved model>`; rank-formula and baseline versions stored with every ranking and disposition; resolved-model pinning (mismatch → discarded, unit `incomplete`); cap `policy.decisions.max_per_run` (default 2,000); unranked view always available; nightly audit is a probability sample with a stored inclusion probability per unit, presented blind (annotation, stratum and rank hidden); the interim draw is independent per unit at that probability under a stored audit-sampler version, with a provisional 5 % total until X1's preregistration is signed off (sampling unit, strata, draws and floor fixed there) and Jev-quality lines `insufficient` until then; every disposition records its view (audit or not) and annotation visibility; deterministic ranking baseline used and disclosed on fork PRs or without key/egress, and stored for every finding so X1 can measure Jev's gain (design §6, §13 X1).
- **REQ-30** Jev never closes or dismisses issues, never defines identity, and `P(real_defect)` is not presented as calibrated confidence (design §6 "Rejected"; E2 `evaluation/jev-triage-v1/RESULTS.md`).
- **REQ-31** **Model-use policy (user rule, 2026-10-22):** OpenRouter is used only for TypeSafe Jev. Every other model use goes through local subscription CLIs (`claude -p`, `codex exec`) with Anthropic/OpenAI models only; Clawpatch-driven lanes use Clawpatch provider `codex` or `claude` (host auth context, no API-key environment). Decided (`docs/STATUS.md` §6.1, R-036; design rev 2.2 §4, §6): a CLI-subprocess adapter replaces the HTTP contextual provider product-wide; reasoning-LLM lanes run only where a model CLI is logged in and otherwise report `skipped (no model CLI)`, which does not change the exit code; CI carries no Anthropic/OpenAI API keys.
- **REQ-32** The Jev/OpenRouter key comes from Infisical, injected per command, for local runs, and from a GitHub Actions secret on trusted CI runs (never on fork PRs; design §6, §7.4); never printed, exported or committed (`AGENTS.md:52-61`). `claude -p`/`codex exec` use the CLI's subscription login and need no Scout-managed secret.

## 7. Inertia monitoring (descriptive only)

- **REQ-33** Per scope, store versioned `metric@version` signals with numerator, denominator, population digest, tool/config digests and missingness (`unmeasured` ≠ 0, `insufficient` ≠ good): principal measures (design §8.2), exposure/interest measures (§8.3). Every panel carries `claim: descriptive` (§8.4).
- **REQ-34** Trend flag only after ≥ 3 consecutive comparable windows adverse past a practical delta on a stable population; `n < 20` changes → `insufficient`; no composite score; the watch list is descriptive; interventions can be annotated and excluded from flags (design §8.4).
- **REQ-35** A metric's `claim` moves to `predictive` only by a preregistered, council-signed pooled test meeting the design §8.5 rule; Scout never upgrades a claim itself. E1b found no material PR-effort or predictability effect from complexity burden (`evaluation/change-cost-v2/RESULTS.md:5-21`), so principal series are framed as "structure to watch", not cost (design §8.1, §8.5; framing provisional pending design §15.3).
- **REQ-36** Pooled-study export is a separate opt-in: pseudonymous, scope-level aggregates plus a sanitised change table; no source, messages, emails or reviewer identities (design §8.6). Scout owns the export and its consent (design §15.1: contents and first dataset still open); inertia-bench runs any pooled studies (`docs/STATUS.md` §6.4; design Appendix 11).

## 8. Subject interface for inertia-bench (new; not yet in design 2.1)

- **REQ-37** Scout is a future *subject* of inertia-bench. It must expose `capabilities --out caps.json` and `measure --request request.json --snapshot <dir> --out measures.jsonl` per `inertia-bench/docs/design/05-harness-design.md:879-920`: schemas `ib-subject-caps/1`, `ib-subject-request/1`, `ib-subject-measure/1`; read-only snapshot, no network, wall-clock cap; one line per requested (unit, measure) sorted by (`unit_id`, `measure_id`); `value` finite or `null` with `unmeasured_reason` from the closed set; two runs byte-identical; opaque ids, no outcome or arm labels. Adoption of this requirement needs a council-signed design change and inertia-bench sign-off (`docs/STATUS.md` §5 item 3b, §6.3).

## 9. Non-functional

- **REQ-38** Persistence performance: per-job commit cost flat in job count; no plan decode or run-payload rewrite per commit; interruption-resume yields the same finalised run; `scout assess` on the pinned fastapi revision completes in < 10 min (heuristic map, egress off, development Mac) — target confirmed by the user 2026-10-22 (design §10.1, Appendix 14; `docs/STATUS.md` §6.7).
- **REQ-39** Interfaces: CLI authoritative and scriptable (`--json`, stable exit codes); optional read-only web over the same state, local bind by default (design §3.1 `web`, §10.2). New CLI surface per design §10.2: `init`, `diff`, `map --refresh`, `export`/`import [--study]`, `signals`, `trends`, `annotate`, `selfeval`, `readiness --matrix`, and `subject capabilities|measure` (after Phase 3, per §8.7); `approve` removed.
- **REQ-40** Build: Go module with subprocess adapters and SQLite; Node ≥ 22 only for the pinned Clawpatch bundle; checksummed macOS/Linux binaries; no auto-install into an analysed repository; Clawpatch upgrades need a compatibility review (CP07) (design §3.1, §7.4, R9).
- **REQ-41** Scout's own scorecard: per-line `count/attempts` with a Wilson 95 % interval for unweighted binomial lines and a design-based interval for weighted or paired lines, plus `unmeasured` counts, no single solve rate (design §9).

## 10. Open decisions (not requirements until closed)

Design §15 items 1 (study-export contents and first dataset), 2 (`pull-requests: read`) and 3 (inertia framing); the former §15.2 (X8) moved to inertia-bench and the former §15.3 is decided (REQ-38). Product experiments X1–X6, X9, X10 with their decision rules are in design §13.

## Appendix — Removed / superseded from the previous requirements

One line each; the former text is recoverable from the working-tree history only if the user committed it, so the design carries the surviving substance.

- Clawpatch **0.4.0** pin, build and `map --source auto` default (§§3, 10, 12.1, 13) — superseded by 0.7.3 and both-mappers-at-onboarding (design §4, Appendix 7).
- "Clawpatch features are the LLM work queue; every feature × category row accountable" (§§3, 6, 12.3, 12.5) — contextual review is opt-in nightly and demoted (design §3.1 `contextual`, §6).
- Mandatory contextual inspection for completeness of the four families — now opt-in discovery (design §6).
- Severity rubrics `security-impact-likelihood@v1` etc. (§12.3) and "persist all low-to-critical findings" as severity-driven design — severity is rule-defined, never derived from aggregates (design §3.2).
- Enaible issue-level parity as the acceptance bar (§§5, 11, 12.2, 12.8; 28-entry corpus incomplete) — replaced by Track A fixtures and the 9/13 regression corpus; parity is not a prerequisite (design §7.3, §9.8). Enaible inventory moves to `docs/evidence/` (design §10.3).
- Seven-language parity (§5) — tier 1 Go/Python/TypeScript (design §5.3 "Rejected", Appendix 5).
- Wide tool selection roadmap (Checkov, OSV, Trivy, CodeQL, Knip, SpotBugs, dependency-cruiser, Lighthouse, EXPLAIN, pprof, etc., §5) — launch catalogue is design §5.1; others are not scheduled.
- Remediation: issue states `pending verification`, `approval`, `ProposedAction`, `ApprovalRef`, authorised-change verification and `scout approve` (§7, §12.6) — cut (design §1 non-goal 1, §10.2).
- Gap→rule promotion workflow with explicit approval (§8) — kept only as FCR-shaped promotion and fixtures per design §10.3; no approval machinery.
- "Exit 1 when finding meets severity threshold" gate (§12.7) — replaced by attribution/profile gate (design §7.3–7.4).
- Durable baseline "bundle from persistent storage"/CI cache and the AWS/OIDC prerequisite (§12.7; `scout-nightly.yml`) — replaced by the `scout/state` branch; S3 deferred (design §7.2).
- Zero-call no-change nightly as a hard contract (§§4, 12.4) — retained only as reuse by cache key (design §3.4); wall-time promises are not made.
- Full rescans as the default nightly shape and full-repo `.scout/clawpatch` state attachment options (§§3, 12.1) — superseded by lanes and the state branch.
- Research/hypothesis-testing duties (predictive-validity studies of inertia, controlled agent continuations) and "weak-model labelling" — not Scout requirements; research lives in inertia-bench (STATUS §4). Design 2.1 still lists X7/X8 pending its re-scope (STATUS §6.4). Weak-model labelling appears nowhere in the old requirements text or design, so nothing further was removed under that name.
- Non-Jev OpenRouter/vendor-neutral model use as a product assumption (§10 "Contextual models": generic JSON/HTTP provider; workflow API-key secrets) — replaced by REQ-31; the code and workflow still carry the old shape until the user decides (STATUS §6.1).
- Sections 11.1–11.2 measured offline corpus statistics (6/9 matches, 28 incomplete comparisons) — historical evidence, kept in `evaluation/outcomes-v*.json`, not requirements.
- Section 13 source baselines (Enaible paths, Harness revision, Clawpatch 0.4.0 checkout) — moved to the research ledger (R-002, R-003, R-007) and `docs/evidence/` at Phase 6.
