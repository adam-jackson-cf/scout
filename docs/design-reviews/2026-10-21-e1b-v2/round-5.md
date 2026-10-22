# Design review — E1b v2 revision 2.1 — round 5

Scope: `docs/design-reviews/2026-10-21-e1b-v2/scout-design.proposed.md` (SHA-256 `7ff21a7fba70a5e3af7a735b0aa40f1ad654d17714df8c3549da6ebcf89b89c5`, as computed by the main agent with `shasum -a 256` and reported in `round-4-response.md:3`; this council session has no shell and did not re-hash) vs `scout-design.base.md` (`901b97cb…` = `docs/scout-design.md`), `proposed.diff`, rounds 1–4 and responses; evidence `evaluation/change-cost-v2/{PREREGISTRATION.md, AMENDMENT-1.md, RESULTS.md, results.json, sensitivity.json, validation-gates.json, power-results.json, common.py, collect.py, build_panel.py, fetch_prs.py}`, `docs/{hypothesis,research}-ledger.md`, `AGENTS.md`. Read-only; `.codegraph` untouched; no secrets; no API calls. Independent challenger `history://DesignE1bV2R5.E1bV2R5Challenger` ran blind; synthesis in §4.

## 1. Verdict: ZERO ISSUES — sign-off granted (§5)

All three round-4 issues are resolved exactly as prescribed; the fresh full review of every changed hunk and the whole-document consistency pass found nothing. The challenger independently reached the same verdict with no new issue.

## 2. Round-4 resolution check

| # | Status | Evidence |
|---|---|---|
| 1 Bot PRs defined but not excluded | **Resolved** | `proposed.md:207` "**PR-level measures** (I4, `batch_size`, the active-hours part of I5): merged, **non-bot** PRs with ≤ 50 changed files and ≤ 100 commits …" and "bot PRs are counted and excluded (`collect.py:144-145`)". Matches `collect.py:144-145` (`if is_bot(...): return dict(base, skip="bot")`), `common.py:85-86` (`author_type == "Bot" or login.lower() in BOT_LOGINS`), `PREREGISTRATION.md:41-42`. |
| 2 Over-cap PRs counted only on API truncation | **Resolved** | `proposed.md:207` "A PR **over either cap**, or whose file list or commit list the API returns incomplete (fewer entries than its total count, or no commits), is marked `linkage: incomplete`, counted and excluded (as E1b `collect.py:146-149`)". Matches `collect.py:146-149` (`n_files > BULK_FILES or len(files) < n_files` → `bulk`; `n_commits > len(commits) or not commits` → `many_commits`); 100-commit cap = `PREREGISTRATION.md:42`, `fetch_prs.py` `commits(first:100)`. |
| 3 Ledger HA1 Tests cell stale | **Resolved** | `docs/hypothesis-ledger.md:21` HA1 Tests = "— (X8 tests HA2 only; HA1 needs a human comparison)", Updated 2026-10-21; history line `:32` "2026-10-21 — HA1: Tests corrected from X8 to none … (design review …/round-4.md)". HA2 `:22` keeps X8; HA3 `:23` "—". Agrees with `proposed.md:251` (b), `:374` X8, `:400` item 2. |

`proposed.diff` is regenerated: its `:207` hunk carries both round-4 phrases and reconstructs `base.md:206` → `proposed.md:207`; no other hunk changed relative to round 4 (`round-4-response.md:7-13`).

## 3. Issues

None.

Checked and found correct (changed text + whole document):

- **PR population consistency.** `:207` is now the only operational PR population; `:216` I4 ("from the PR commit list … single-commit PRs … counted separately"; `unmeasured` < 5 linked PRs), `:217` I5 (`unmeasured` whenever I4 is), `:261` §8.6 change table ("bot flag and `bot_reason` (§8.3), linkage completeness"), `:358` R11 and `:402` §15 item 4 all propagate the same missingness and provenance. Bot PRs excluded from measures yet present in the change table with `bot_reason` (`api_bot_type`/`login_list`) is coherent with `:261`. `:108` lane 3 states the commit-side bot predicate (`[bot]` suffix / exact `history.bots`, never substring), distinct from and consistent with `:207`'s API predicate for PRs; "E1b-v1's substring regex … is retired" is the only remaining v1 mention besides `:5` ("superseded and invalidated, R-033; not cited as evidence").
- **Design rule 3 (only signed-off experiments cited).** `:5` cites E1b = `change-cost-v2/{PREREGISTRATION,AMENDMENT-1,RESULTS}.md` (R-035); sign-offs exist: `reviews/round-10.md:5` "ZERO ISSUES — sign-off granted" (prereg + scripts, digests `:39-`), `reviews/round-14.md:5` "ZERO ISSUES — Amendment 1 revision 2 is signed off", `:36` prereg digest byte-identical to round 10; `RESULTS.md:84-87` records Amendment 1 under Deviations. `:5` confines E1/E2/E3/CP07 to pre-rule engineering observations and states "no change in revision 2.1 rests on them".
- **Numbers** (all re-derived from JSON, not from prior rounds): `results.json .estimates` T1 −0.01120 ci90 [−0.02898, +0.00658]; T2 +0.02280 [+0.00445, +0.04115], p 0.0412; T3 +0.01755 [−0.00321, +0.03831]; T4 +0.00609 [−0.01036, +0.02255]; P1_rel_gain −0.00153 ci95 [−0.00691, +0.00384]; P1_base_vs_const 0.3334; n_T1 7,028 / 451 clusters, n_T2 6,110 / 417; G 480; G3 n 7,030, floor 0.3815 → "38 %"; G3b ρ 0.4597 → "0.46"; G2 twelve `pass: true` repositories (axum, caddy, django, jellyfin, nest, pandas, prometheus, scikit-learn, terraform, tokio, vite, webpack) sum to 7,717 = `power-results.json .n`. Holm p 0.165 (`RESULTS.md:13`); "council decision b" read-together rule (`RESULTS.md:21`). `power-results.json` empirical MDE80 T1 0.04488, T2 0.06289, T3 0.04191, T4 0.05276, SESOI 0.1, all `adequate` → `:233` "0.042–0.063", `:389` "0.045 … 0.063". `validation-gates.json` G5 `jev_recall_ipw` 0.422, G7 `kappa_pairs` 0.514, `hidden_precision` 0.63. All match `:233-243`, `:249`, `:389`, `:403`, `hypothesis-ledger.md:14-17`, `research-ledger.md:43`.
- **Phase 3 acceptance executable.** `evaluation/change-cost-v2/raw/prs-{nest,webpack}.json.gz` and `panel-final.csv.gz` exist; `build_panel.active_hours` (`build_panel.py:56-62`) produces the `active_hours` column (`:137`); `:216`'s session wording (gap ≤ 4 h added, else new session, 0.5 h first credit) matches.
- **§15 vs body.** Items 1–5 (`:399-403`) agree with `:259-263` (export/consent), `:374` X8/HA2, `:216-217` and `:358` (PR ingestion → I4/`batch_size`/I5 active-hours `unmeasured`), `:247-249` (provisional framing, "structure to watch", no `claim` changes), `:39` (revision 2.1 provisional note). X7 (`:373`) and the fixed decision rule (`:253`) agree; A8 (`:389`) defers sizing to a preregistered power simulation and cites E1b figures "for scale only".
- **Ledgers.** R-033 (invalidated v1), R-034 (user direction), R-035 (v2) carry distinct meanings (`research-ledger.md:41-43`); H1/H2 `not-supported`, H4 `inconclusive` with history `:33-34` matching `RESULTS.md:7-19`.

Observation, not an issue (prior round accepted this text unchanged): `:247` "(5) agent-authored change (HA1–HA3; X8)" lists the hypotheses E1b leaves untested and the related external study; `:374` and `:251` immediately scope X8 to HA2, so no reader can take X8 as testing HA1/HA3.

## 4. Challenger synthesis

The challenger (different model family, blind to my pass) returned ZERO ISSUES with the same three resolutions quoted at the same lines, the same numeric cross-checks to more digits (T2 Holm p 0.164779; floor 0.3815078; max sensitivity UB 0.0602542 pre-2024 T2), and confirmed `proposed.diff:85` reconstructs the `:207` change. It did not change my recommendation. Its one caution is adopted into §6: sign-off of this revision does not resolve §15 item 5 (the outcome-first vs structure-first framing), which E1b cannot settle because its scope excludes other quality measures, defects and agent effects (`:247`); the item stays open for the user. Neither reviewer independently hashed the file (no shell in either session).

## 5. Sign-off statement

The innovation council signs off, with zero issues, Scout design revision 2.1 as contained in `docs/design-reviews/2026-10-21-e1b-v2/scout-design.proposed.md`, SHA-256 **`7ff21a7fba70a5e3af7a735b0aa40f1ad654d17714df8c3549da6ebcf89b89c5`**, against base `docs/scout-design.md` (`901b97cb…`). The sign-off is of that exact byte content: before adoption the main agent must confirm `shasum -a 256` of the file still prints this digest (neither council session could hash it); any other digest requires a new round (`AGENTS.md` design rule 2 ↔ experiment rule 5). The accompanying edit to `docs/hypothesis-ledger.md:21,32` (HA1 Tests cell + history line) is part of the adopted change set.

Design-rule compliance: the revision cites only E1b v2 (R-035), signed off in `evaluation/change-cost-v2/reviews/round-10.md` and `round-14.md` and run under the experiment rules; E1b-v1 appears only as "superseded and invalidated … not cited as evidence".

## 6. Risks and reversibility

- Digest provenance: as above; trivial to confirm, blocking if it differs.
- §15 item 5 remains an open user decision; adopting revision 2.1 adopts the *provisional* framing (`:39`, `:249`, `:403`), not a resolution. Reversible by a later design round.
- `linkage: incomplete` labels over-cap PRs as well as API-truncated ones (round-3/4 uniform rule); if a future need arises to distinguish them, the change table's counts make that a non-breaking addition.

## 7. Execution outline (adoption)

1. Main agent: `shasum -a 256 docs/design-reviews/2026-10-21-e1b-v2/scout-design.proposed.md` → must equal `7ff21a7f…`; `docs/scout-design.md` must still equal `901b97cb…`.
2. Copy `scout-design.proposed.md` over `docs/scout-design.md`; keep the ledger edit at `docs/hypothesis-ledger.md:21,32`.
3. Record this round as `docs/design-reviews/2026-10-21-e1b-v2/round-5.md` (this report) per design rule 2.
4. No further review needed unless any byte of the adopted text changes.
