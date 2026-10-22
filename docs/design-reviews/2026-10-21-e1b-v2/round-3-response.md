# Design review E1b v2 — Response to round 3

All three round-3 issues are fixed in `scout-design.proposed.md` (now `55810f1a…`). The base file and `docs/scout-design.md` are unchanged (`901b97cb…`).

| # | Fix (proposed file line) |
|---|---|
| 1 | Lane 3 at `:108` now excludes bot commits by the `[bot]` author suffix or an exact `history.bots` entry, and never by substring. It points to §8.3, so it now matches `:207`. |
| 2 | **PR inclusion:** at `:207`, PR-level measures now require ≤ 100 commits as well as ≤ 50 files. **Incomplete API data:** if the API returns an incomplete file list or commit list (fewer entries than its total count, or no commits), the PR is marked `linkage: incomplete`, counted and excluded. This is the same rule as E1b's `collect.py:146-149`. |
| 3 | `bot_reason` now has values for both row types: `app_suffix` or `configured` for commits, and `api_bot_type` or `login_list` for PRs (`:207`). `:257` refers to §8.3 for these values. |

`proposed.diff` has been regenerated.
