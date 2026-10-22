# Design review — E1b v2 revision 2.1 — round 3

Scope: `docs/design-reviews/2026-10-21-e1b-v2/scout-design.proposed.md` (sha256 `d702b468…` per `round-2-response.md`; not independently hashed — read-only, no shell) vs `scout-design.base.md` (`901b97cb…`), `proposed.diff`, `round-1.md`, `round-1-response.md`, `round-2.md`, `round-2-response.md`; evidence `evaluation/change-cost-v2/{PREREGISTRATION.md, RESULTS.md, AMENDMENT-1.md, results.json, sensitivity.json, validation-gates.json, power-results.json, common.py, build_panel.py, fetch_prs.py, collect.py, panel-final.csv.gz, raw/}`, `docs/{hypothesis,research}-ledger.md`, `AGENTS.md`. Read-only; `.codegraph` untouched; no secrets; no API calls. Independent challenger `history://DesignE1bV2R3.E1bV2R3Challenger` ran blind to this analysis; synthesis in §4.

## 1. Verdict: ISSUES FOUND (3 minor). Not signed off.

All eight round-2 issues and the three round-1 residues are substantively fixed; every number re-checked matches the evidence files (§5). What remains are three definition/consistency gaps in the bot and PR-linkage text that round 2's fixes introduced. Each is a one-sentence edit; no evidence or conclusion changes.

## 2. Round-2 resolution check

| # | Status | Evidence (proposed file) |
|---|---|---|
| 1 (major) units | Resolved | `:233` "SESOI 0.10 per SD — log-hours for T1, T2; relative dispersion for T3, T4", MDE80 range "in each test's units"; `:240` T4 "+0.006 relative dispersion/SD"; `:247` and `:403` state both thresholds with their units; `:389` A8 labels T1/T2 log-h/SD. Matches `PREREGISTRATION.md:128-130` |
| 2 `:39` marker | Resolved | `:39` "(Revision 2.1: provisional pending §15 item 5, after E1b, §8.5.)"; consistent with `:249`, `:403` |
| 3 T1/T3 CIs | Resolved | `:241` T1 [−0.029, +0.007], T3 [−0.003, +0.038] = `results.json .estimates.T1.ci90`, `.T3.ci90` |
| 4 bot seed / provenance | **Partly** | `:207` drops the underivable no-reply seed and `web-flow`; `[bot]` suffix + exact `history.bots` list; `bot_reason` stored (`:207`, `:257`). Residue: `:108` not updated (issue 1); `bot_reason` has no value for PR-level bots (issue 3) |
| 5 PR→scope linkage | **Partly** | `:207` membership by GraphQL `PullRequest.files` (= `fetch_prs.py:31`), ≤ 50 files, truncated file list → `linkage: incomplete` (counted, excluded). Residue: commit-list completeness, which E1b enforces, is not specified (issue 2) |
| 6 I5 missingness | Resolved | `:217` "`unmeasured` whenever I4 is", churn "git, always"; `:358` R11 and `:402` item 4 include I5's active-hours part |
| 7 floor denominator | Resolved | `:247` "38 % of the 7,030 PRs with a measured `Y_active`" = `results.json .gates.G3` {n 7030, floor_share 0.3815}. 38 % is a correct rounding; 38.2 % (`RESULTS.md:29`) optional |
| 8 associational | Resolved | `:247` "associational, not causal (time-varying confounding remains; prereg §6, RESULTS §6)"; `:403` "no material association … it is associational" |

Round-1 residues (issues 3, 7, 8, 9 marked Partly in round 2): 7, 8, 9 now resolved (`:39`, `:233`/`:240`/`:247`/`:403`, `:241`); 3 remains partly, folded into issues 1 and 3 below.

## 3. Issues

1. **Minor — `:108` describes a different bot rule from `:207`.** Lane 3 (`:108`): "bot commits excluded by an exact author-identity list, never substring — §8.3". §8.3 (`:207`): author name ending in the literal suffix `[bot]` **or** an exact entry of `history.bots`, which is *empty by default*. Under `:108`'s wording the default configuration excludes no bots; under `:207` it excludes every GitHub-App author. The two sections produce different populations. **Fix:** `:108` → "bot commits excluded by the `[bot]` author-name suffix or an exact author-identity list, never substring — §8.3".

2. **Minor — PR commit-list completeness is unspecified, so active hours can be computed on a partial list.** `:207` marks only a truncated *file* list `linkage: incomplete`; `:216` computes I4 from "the PR commit list" with no completeness condition. E1b caps PRs at ≤ 100 commits (`PREREGISTRATION.md:42`), fetches `commits(first:100)` (`fetch_prs.py:29`) and skips any PR whose commit list is truncated or empty (`collect.py:148-149`, `skip="many_commits"`). Without that guard a truncated list silently understates active hours and contaminates I5's active-hours ratio, contradicting `:216`'s "E1b prereg §3.3" construct claim and R4's "PR effort defined as in E1b" (`:351`). **Fix:** `:207` add "a PR whose API commit list is empty or truncated (E1b: > 100 commits, prereg §2) is `linkage: incomplete` for I4 and I5's active-hours part (counted, excluded); `batch_size` still uses the complete file list"; `:257` record commit-list completeness separately from file-linkage completeness.

3. **Minor — `bot_reason` has no value for PR-level bots.** `:207` fixes `bot_reason` ∈ {`app_suffix`, `configured`} — both commit-side predicates — while `:257` stores `bot_reason` for "deduplicated change/PR keys", and `:207` defines PR bots by API type `Bot` or `common.BOT_LOGINS` (`common.py:81-86`). A PR row flagged as bot has no admissible reason, so the provenance requirement of round-1 issue 3 is unmet for PRs. **Fix:** `:207` "`bot_reason` (`app_suffix` | `configured` for commits; `api_bot_type` | `api_login` for PRs)".

No other defects found in the changed text or in whole-document consistency (checked: E1b-v1 references removed throughout — `change-cost-v1`, "O1–O5", "60 areas" and the substring regex appear only as "retired/not cited"; §1, §2, §4, §5, §8.1–8.6, §11 Phase 3, §12 R4/R6/R11, §13 X7/X8, §14 A8, §15 items 2/4/5 agree with one another; ledger IDs R-033, R-034, R-035, H1–H4, HA1–HA3 exist with matching meaning, `research-ledger.md:41-43`, `hypothesis-ledger.md:14-24`; `:5`'s pre-rule note matches `hypothesis-ledger.md:33-35`; E1b sign-offs recorded, `RESULTS.md:3`, `PREREGISTRATION.md:3-5`; Phase 3 acceptance `:328` is executable — `raw/prs-{nest,webpack}.json.gz` and `panel-final.csv.gz` exist, the CSV carries `repo`, `number`, `active_hours`, `slim()` keeps `authoredDate` (`fetch_prs.py:97-98`), and `:216`'s session rule equals `build_panel.active_hours` (`build_panel.py:18-19,56-62`)).

## 4. Challenger synthesis

Agreed and adopted: issue 1 (`:108` conflict) and issue 2, where the challenger's `collect.py:148-149` evidence strengthened my finding beyond the prereg cap — E1b's guard is an executed exclusion, not just a fetch limit. Rejected on evidence: its minor "38 % must read 38.2 %" — round-2 issue 7 was the missing denominator, now present; 0.3815 rounds to 38 %, so the statement is correct as written (38.2 % is a free alignment with `RESULTS.md:29`, not a defect). Issue 3 (`bot_reason` for PR rows) is mine; the challenger did not raise it. Numeric checks agree in full.

## 5. Evidence (numeric cross-checks)

`results.json`: T1 −0.0112 ci90 [−0.0290, +0.0066]; T2 +0.0228 [+0.0044, +0.0412], p 0.041; T3 +0.0175 [−0.0032, +0.0383]; T4 +0.0061 [−0.0104, +0.0226]; P1_rel_gain −0.00153 ci95 [−0.0069, +0.0038]; P1_base_vs_const 0.333; n_T1 7,028 / 451 clusters, n_T2 6,110 / 417; G 480; Σ gate-passing repos (12) = 7,717; G3 n 7,030, floor 0.3815; G3b ρ 0.4597. `power-results.json`: empirical MDE80 T1 0.0449, T2 0.0629, T3 0.0419, T4 0.0528. `sensitivity.json`: max T2 ci95 UB 0.0603 (`d_cohort_merged_before_2024`); all other variants incl. `a_multicommit…` (0.0433), `f_author_clusters` (0.0531), LORO max (0.0511, scikit-learn) ≤ 0.061; per-repo pandas +0.0960 [+0.0479, +0.1442], vite +0.0611, tokio −0.1143. `validation-gates.json`: G5 recall 0.422, G7 κ 0.514, hidden precision 0.63. `PREREGISTRATION.md:193` council decision b; `:74-77` control set as `:233`. Holm 0.165 per `RESULTS.md`/round 2 (not recomputed).

## 6. Risks

- Issue 2 may slightly reduce I4 coverage on repositories with very long PRs; it mirrors E1b, so comparability improves, not worsens.
- All fixes are documentation edits; fully reversible; no §15 decision text changes.

## 7. Execution outline (round 4)

1. `:108` wording (issue 1).
2. `:207`, `:257` commit-list completeness rule and missingness propagation (issue 2).
3. `:207` `bot_reason` enum for PR rows (issue 3).
4. Regenerate `proposed.diff`, record the new SHA-256, request round 4 (expected sign-off if nothing else changes).
