# Changelog

## Unreleased

- Keep issue identity across file renames even when the rename run has no relevant successful coverage; unmatched issues follow the new path without being resolved.
- Preserve heuristic features and report incomplete exploration when an agent map succeeds with zero features (REQ-56).
- Add merge-time area association: opt-in Jev `association` lane, separately consented codex Terra fallback, `scout areas list --pending`, `accept` and `merge`, and a refresh guard that recommends manual mapping without starting it.
- Add `scout decisions orderflip`; REQ-99's live `merge.which_area` record is 11/92 flipped (0.1196), 0 unmeasured (2026-11-07; `docs/jev-adapter-audit.md`). This is a required record, not a pass/fail threshold.
- Association lanes remain **off by default and provisional**. Fastapi P_assign 0.922 (71/77) did not replicate on axum (0.622). Closed-loop error accumulation and duplicate pending areas are unmeasured; refresh-guard defaults are assumptions. Evidence and phase records: `docs/STATUS.md` §1.
