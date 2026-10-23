# Jev adapter audit (`internal/decisions`, 2026-10-23)

Scope: the decisions adapter, its ranking hook (`internal/engine/ranking.go`), the decisions policy defaults
(`internal/policy/consent.go`) and the `scout init --decisions` default. Standard: the `jev-actions` skill
(`SKILL.md` and `references/*.md`) and the 12 adapter requirements of
`docs/design-reviews/2026-10-22-rescope/jev-actions-review.md` §B (cited `B1`–`B12`). The design document
(`docs/scout-design.md`) is council-governed and was not edited; gaps that need a design change are marked
**design-level**.

Skill file abbreviations: `DP` docs-and-programming-model, `DPT` design-patterns, `QD` question-design,
`CI` compose-and-integrate, `BED` building-eval-data, `LC` labeling-cases, `TT` threshold-tuning,
`CC` calibration-check, `EA` error-analysis. `S` = status: **Met** (already conformant), **Fixed** (gap closed in this
change, with a test), **Design-level** (needs a design or data decision, reason given), **N/A** (practice has no
subject at launch).

Live docs read for this audit: `docs.typesafe.ai/models.md` (model IDs, aliases, pricing, context limits) and
`model-jaggedness/jev-1.13.md`. The Jev 1.14 jaggedness page returned HTTP 404, so the 1.13 page is applied as the
nearest published one. The HTTP API page was not re-read; the OpenRouter `alpha/decisions` contract stays
`[INFERENCE]` (design R14).

## Pinned model slug

`scout init --decisions` now pins **`typesafe/jev-1.14-20261014`** (`policy.DefaultModel`,
`internal/policy/consent.go:27-31`). Sources, checked 2026-10-23:

1. `GET https://openrouter.ai/api/v1/models/typesafe/jev-1.14-20261014/endpoints` returns HTTP 200: model
   `typesafe/jev-1.14` ("TypeSafe: Jev 1.14"), one endpoint named `TypeSafe | typesafe/jev-1.14-20261014`, provider
   `TypeSafe`, prompt price `0.000000042` USD per token (= $0.042 per Mtok, the `models.md` price for Jev 1.14).
2. `evaluation/jev-triage-v1/metrics.json` and `evaluation/change-cost-v2/DESIGN-NOTES.md`: runs that requested
   `typesafe/jev-1.14-20261014` had it echoed back as the resolved model (66 of 66 calls, 0 failed, per variant).
3. The undated id `typesafe/jev-1.14` is the moving release line; `GET /api/v1/models` lists only
   `typesafe/jev-router`. The undated line is therefore rejected as an alias (below). TypeSafe's own ID `jev-1.14.0`
   (`models.md`) belongs to `api.typesafe.ai`, not OpenRouter, and is not the default.

Not verifiable `[INFERENCE]`: whether OpenRouter ever retires or re-points the dated slug. The resolved-model check
(row 15) is what protects against it.

## Audit table

| # | Practice | Skill source | S | Evidence | Fix / note |
|---|---|---|---|---|---|
| 1 | Primitive follows what the answer means: a yes/no condition is a Noul | QD:14-22 | Met | `triage.go:13-27` (Noul, P(true) = P(real_defect)) | none |
| 2 | Untrusted state is data, not instructions | QD:31-32 | Met | `triage.go:16`; asserted in `decisions_test.go:202` (`never instructions` in the request) | none |
| 3 | Named JSON fields, backticked paths, state shape stated | QD:26 | Met | `triage.go:16-18` (`state.excerpt`, `state.tool_signal`, `state.facts`), `triage.go:23-26` shape list | none |
| 4 | Question IDs are for code and never sent | QD:29-30; B1 | Met | wire keys are `q1…qn` (`question.go:190-192`); `decisions_test.go:202` asserts the ID is absent | none |
| 5 | One narrow judgment per question; one Noul per reason | QD:33-37 | Design-level | `triage.go:16-18` joins "mechanism occurs" ∧ "shipped" ∧ "hand-written" ∧ "non-test" ∧ "no visible guard" | Splitting or rewording changes the version key and needs a labelled dev/test set (BED:62-69, EA:21); the label set is `none` and E2's evidence (n = 22) is for this wording. Backlog **B-1**: split into per-reason Nouls once ≥ 30 labels per label exist; path role is already a code fact (`triage.go:38-43`) |
| 6 | Binary criteria; criteria agree with the instruction | QD:38; jaggedness 1.13 §7 | Met | `triage.go:19-22` (`true` means yes, `false` means no) | none |
| 7 | Jaggedness: literal reading, indirection, large state, adversarial content | QD:49-51 | Met (one note) | state bounded to 8 excerpt lines / 600 bytes (`triage.go:31-35`); injection fixture `ranking_test.go:253`; indirection see row 5 | 1.14 page is absent, so the 1.13 list is applied |
| 8 | Supply the reference evidence the judgment needs; keep derivations and unsupported factuality out | QD:25, 52-56 | Met | excerpt, rule, mechanism, path role are sent (`triage.go:54-100`); nothing asks Jev to compute or verify a derivation | none |
| 9 | Pairwise comparisons judged in both orders | QD:57-58 | N/A | no pairwise question at launch | X10 not admitted yet |
| 10 | Independent questions over one state go in **one request**; a second request only when an answer builds the next state | CI:11-15; B4 | **Fixed** | was: `Ask` took one question and rejected `len(answers) != 1`. Now `Unit.Questions` (`run.go`), `Client.Ask` (`client.go:129`), `BuildRequest` (`question.go:210`), reply must answer every key (`client.go:303-304`) | tests `adapter_practices_test.go:42` (3 questions, 1 request, 3 rows) and `:77` (missing answer ⇒ schema-invalid; duplicate IDs rejected) |
| 11 | Version key = hash of instructions, criteria, state shape, option order, model — not the ID | CI:35-36; B2 | **Fixed** | `question.go:111-132`: the question **kind** was not in the payload (a Noul and a Score with the same text collided) | kind added; test `adapter_practices_test.go:144`; existing mutation test `decisions_test.go:101` |
| 12 | Exact (redacted) request stored beside every annotation; realised option order stored | CI:35; B2, B7 | Met | `run.go` sets `RequestBody`, `OptionOrder`; `decisions_test.go:158`, `ranking_test.go:211` | none |
| 13 | Pin a versioned model, never an alias; reject at config load | CI:32-33; B3 | **Fixed** | policy rejected `jev-latest`/`-preview`/`:latest` but accepted the **undated release line** `typesafe/jev-1.14` (`consent.go:76-92`), which OpenRouter re-points and reports back as the dated slug, so the pin could never match | `policy.IsModelAlias` now also rejects `^(vendor/)?jev-N.M$`; `decisions.Config.Validate` uses the same function (`run.go:43-50`), one source. Tests `decisions_defaults_test.go:14,27`, `adapter_practices_test.go:153`, `init_test.go` (undated and preview rejected) |
| 14 | Default pin for `scout init --decisions` is Jev 1.14 via OpenRouter; an existing pin is never replaced | user decision (b) | **Fixed** | was: a missing model was an init error. Now `init.go:81-93` defaults to `policy.DefaultModel` for a new pin only; interactive empty answer takes it | tests `init_test.go` `TestInitDecisionsPinsJev114ByDefaultAndKeepsAnExistingPin` (also asserts the slug), `decisions_defaults_test.go:8` |
| 15 | Record the model the response reports; compare to the pin on every call; mismatch ⇒ discard, unit `incomplete`, counted | CI:34; B3 | Met | `client.go:300-302`; `run.go` `UnitIncomplete`; health text `incomplete (model mismatch)` | tests `decisions_test.go:226`, `ranking_test.go:338` |
| 16 | A model bump creates a new version key; prior annotations and thresholds go stale | CI:37-38; B3, B8 | Met / **Fixed** (thresholds) | annotations: `ranking.go:153` (`MarkAnnotationsStale`); thresholds: row 21 | annotation test in `ranking_test.go` model-bump case; threshold test `ranking_thresholds_test.go:40` |
| 17 | Failed or invalid call is a failed decision, never a pass or low score: network, HTTP 4xx/5xx, 429 with `retry-after`, timeout, schema-invalid, missing key, fork, egress off, cap reached | CI:39-40; B5 | Met | `client.go:129-207` (400/422 non-retryable, 429/529 backoff honouring `Retry-After`), `run.go:63` (`Eligibility.Skip`), per-finding disclosure `run.go:293` | tests `decisions_test.go:265,288,326,366`; `ranking_test.go:155,299` |
| 18 | A unit with several rows that fails to persist | CI:39 | Met (accepted) | `run.go` persists rows in order; a sink error fails the unit | rows already written are valid answers of the same call; only a store failure triggers it; the unit ranks deterministic and is disclosed |
| 19 | Cap in **calls** and in **billed cost** | B6 | **Fixed** (cost) | calls: `run.go:207`. Cost was absent | `Config.MaxCostUSD`, policy `decisions.max_cost_usd` (default $0.50 `[ASSUMPTION]`, five times the ≈ $0.10 the call cap costs), `--max-cost-usd` in init. Checked before each call (overshoot ≤ one call); a discarded or mismatched reply is still billed and counted; a reply that reports no `usage.cost` stops the run (`billed cost unknown`) because the cap cannot be enforced. Tests `adapter_practices_test.go:92,115,124,133`, `ranking_thresholds_test.go:81` |
| 20 | Record billed `usage` per call; measure cost and latency | CI:16; B6 | **Fixed** (latency) | usage on the first row of a unit (`run.go`) so sums never double-count; latency was not measured | `Reply.Latency` (end to end incl. retries) → `Annotation.LatencyMS` (first row); test `adapter_practices_test.go:166` |
| 21 | Annotation row fields | B7 | Met / **Fixed** (digest) | finding id, question id, version key, resolved model, full `Answer` JSON incl. probabilities, run id, timestamps: `domain/annotation.go`. Ranking provenance: `RankingRecord` (`ranking.go:224`). Audit stratum and inclusion probability: `AuditDraw` (`audit.go`). Shown-at-disposition: `disposition_contexts.annotation_visible` (`state/annotations.go:25`) | **state digest** was missing: `Annotation.StateDigest` (SHA-256 of the exact state sent) added; test `adapter_practices_test.go:42`, `ranking_thresholds_test.go:24`. Provenance, audit and visibility sit in sibling tables joined by issue id by design (schema V8), not on the row |
| 22 | Thresholds are named configuration beside minimums, metric, label set, version key, score space, status, cost rule, assumptions — never literals | TT:90-92; BED:85-91; B8 | **Fixed** | none existed | new `internal/decisions/thresholds` (embedded `defaults.json`; committed override `.scout/decisions/thresholds.json`). Default for `triage.real_defect`: `status: provisional`, label set `none`, score space `raw`, minimums precision 0.8 / recall 0.9 (`assumed`: skill default 0.8/0.8, recall raised because a miss costs more, TT:43-46), metric F-beta β = 2 (TT:24), threshold 0.3 for a costlier miss (BED:94-97), `gating: false`, cost rule, assumptions and a visible TODO naming the missing cases (BED:99). Tests `thresholds_test.go:13-91` |
| 23 | A threshold set against another version key is stale ⇒ provisional and disclosed | CI:37-38; B8 | **Fixed** | `thresholds.Resolve` (`thresholds.go:143`); `ranking.go:124-134,203-206`; disclosure `jev (unvalidated, X1 pending; thresholds stale)` (`run.go:288-289`); `Ranking.Thresholds` | tests `thresholds_test.go:37`, `ranking_thresholds_test.go:40`. An invalid or unreadable file is a disclosed deterministic fallback with no Jev call (`:60`) |
| 24 | Configuration a pull request authors must not steer its own ranking | (trust model; policy is read from a ref) | **Fixed** | override is read with `git show <ref>:<path>` (`thresholds.ReadCommitted`), at the base SHA in `scout diff` (`diff.go:255`) and HEAD in `scout` (`cmd/scout/ranking.go:36`), never the working tree | test `thresholds_test.go:91` |
| 25 | Thresholds only where behaviour is gated by a probability; ordering alone needs none | TT:9; CC:13; CI:17-19 | Met | ranking is ordinal (`rank.go:19-26`); the threshold is recorded but `gating: false`, nothing reads it | revisit when a Noul gates behaviour |
| 26 | Rank composition in code over stored answers; changing a weight re-ranks with no model call; baseline stored for every finding | CI:25, 29-30; B9 | Met | `rank.go` (`Formula`, `Score`, `Rank`, versioned); `RankingRecord.BaselineScore` (`ranking.go:224`) | test `decisions_test.go:463` |
| 27 | Mixed weights (rule weight × Jev probability) need a calibration check first | CI:27-28; CC:10-12 | Met | no mixed term exists (`rank.go:13-17`) | none |
| 28 | Probabilities are used for ranking only; no displayed likelihood; raw score space | CC:33-35, 40-42 | **Fixed** | no renderer prints a probability (no use in `cmd/scout`, `internal/web`); `score_space: raw`; the product JSON (`RankingItem`) no longer carries `probability`, and `score` is present only for deterministic items, so `scout-assessment.json` never exposes P(real_defect) (design §6 `:162`) | `ranking_test.go` asserts Jev-ranked items have no score; the probability stays in the state store (`annotations`, `ranking_records`) for X1; test `ranking_thresholds_test.go:98` (the PR summary contains no probability) |
| 29 | No escalation band at launch; any later band is per question, from saved scores, fallback measured on the same cases | EA:36-52; B12 | Met | no band, cut-offs or fallback judge in code or config | none |
| 30 | Choice option order: realised order stored; **two-order flip-rate harness** | QD:57-58; jaggedness 1.13 §8; B11; design §6 | **Fixed** | order stored (`question.go` `RealisedOrder`, `Annotation.OptionOrder`); no harness existed | new `internal/decisions/orderflip`: asks a Choice in an order and its reverse, reports flip rate, max probability shift and cost, leaves failed units `unmeasured`. Tests `orderflip_test.go:46,65,73,93`. It lives outside `internal/decisions` because the adapter's boundary test forbids quality-metric identifiers (`decisions_test.go:568`). Not wired to a command: there is no Choice question at launch; run it on a Choice's own labelled set when X3 admits one (**design-level**) |
| 31 | Typed output is a guaranteed interface, not guaranteed truth | CI:31 | Met | output only orders findings and can only lower a rank; test `ranking_test.go:253` | none |
| 32 | Tests: forced mismatch, injection fixture, 429/timeout never a low rank, version-key change ⇒ stale, flip rate | B11 | Met / **Fixed** | `decisions_test.go:226,288,326`, `ranking_test.go:253,338`, `ranking_thresholds_test.go:40`, `orderflip_test.go` | none |
| 33 | Credentials stay server-side; never in logs, errors or stored requests | CI:41 | Met | `client.go` header only; `decisions_test.go:326` (no key in errors), `ranking_test.go:211` (no key in stored body) | none |
| 34 | Record the state, score, model and version key of every decision | EA:62-64 | Met | `RequestBody`, `Answer`, `ModelResolved`, `QuestionVersion` on every row | none |
| 35 | `scores.csv`-compatible self-eval export (`case_id, label, score, model, question_version, split, stratum`) | B10; BED:76-84 | Design-level | needs blind labels and dispositions, none exist (label set `none`); belongs to the X1 preregistration and `scout selfeval triage` | the row keys it needs (`QuestionVersion`, `ModelResolved`, `StateDigest`, `UnitKey`, ranking and audit joins) are now all stored |
| 36 | Labelling: blind, frozen guide, provenance, two independent labellers | LC:20-38; user decision (c) | N/A (constraint recorded) | no labelling happens in this change | When labels are collected they are `agent_reviewed`: two independent sessions (ideally different model families) from the frozen guide that never see Jev's answer; no user review. Recorded in `defaults.json` assumptions |
| 37 | Escalate or tune only with ≥ ~30 cases per label; below that set a provisional threshold and a visible TODO | BED:92-100; TT:9-11 | Met | row 22 | none |
| 38 | OpenRouter route contract | DP:24-31 | Design-level | `client.go:18-21` marks the contract `[INFERENCE]` (design R14) | unchanged; no live call was made in this audit |

## Backlog recorded (not defects)

- **B-1** Split the triage Noul into per-reason Nouls (row 5) once labels exist.
- **B-2** Label ≥ 30 cases per label (`agent_reviewed`), tune with `--beta 2`, replace the default entry (`defaults.json` TODO).
- **B-3** Run `orderflip` on the first Choice that gates anything (row 30).
- **B-4** `scores.csv` export with the self-evaluation (row 35).

## Reproduce

`source .tools/activate && gofmt -l . && go build ./... && go vet ./... && go test ./...`. No test calls OpenRouter; all
use `httptest`.
