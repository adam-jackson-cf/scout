# Design review E1b v2 — Response to round 2

All 8 round-2 issues are fixed in `scout-design.proposed.md`, which is now `d702b468…`. The base and `docs/scout-design.md` are unchanged at `901b97cb…`.

| # | Fix (proposed file line) |
|---|---|
| 1 (major) | **Units are now stated for each test:** log-hours for the effort tests (T1, T2) and relative dispersion for the predictability tests (T3, T4). This applies to the SESOI and MDE wording in `:235`, the "settles" statement in `:247`, the T4 row in `:240`, and §15 item 5 in `:403`. |
| 2 | `:39` adds "(Revision 2.1: provisional pending §15 item 5, after E1b, §8.5.)". |
| 3 | `:241` now gives the T1 and T3 90 % CIs: [−0.029, +0.007] and [−0.003, +0.038], from `results.json`. |
| 4 | **Commit bots (`:207`):** the author name must end in the literal suffix `[bot]`, which is GitHub's reserved App naming and the git-side counterpart of the API `Bot` type. Alternatively, the name or e-mail must exactly equal an entry in `history.bots` (repository policy, empty by default). No other pattern matching. **Recording:** the change table stores `bot_reason` (`app_suffix` \| `configured`) in `:207` and `:257`. **Removed:** the derived no-reply seed and the `web-flow` mapping are gone. **PR bots:** API `Bot` type, or a login in the exact list `common.BOT_LOGINS`. |
| 5 | **PR→scope linkage (`:207`):** a PR is in scope when the GraphQL `PullRequest.files` list contains at least one eligible file in scope, as in E1b's `fetch_prs.py`; PRs with more than 50 changed files are excluded. A truncated file list sets `linkage: incomplete`; such PRs are counted and excluded. |
| 6 | **I5 (`:217`):** the active-hours part is `unmeasured` whenever I4 is; the churn part always comes from git. R11 (`:358`) and §15 item 4 (`:402`) now include I5's active-hours part. |
| 7 | `:247` now reads "38 % of the 7,030 PRs with a measured `Y_active`". |
| 8 | **Causal scope:** `:247` and `:403` now say the result is associational, not causal: time-varying confounding remains (prereg §6, RESULTS §6). |

The word diff is regenerated in `proposed.diff`.
