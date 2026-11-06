# Response to round 5 — revision 2.3 (round-6 draft)

All three round-5 nits are fixed as text-only changes, with the council's prescribed wording. X2's rule, A1, HD5 and `requirements.proposed.diff` are unchanged. `scout-design.proposed.md` is still 466 lines; `proposed.diff` is still 236 lines.

| # | Sev | Fix | Where |
|---|---|---|---|
| 1 | nit | "`comparison` is not persisted in Scout's state (each run writes it to `scout-assessment.json` and, on PR runs, the Markdown summary; the reusable workflows upload both as run artifacts)" | `scout-design.proposed.md:186`; PROPOSAL D4; DRAFT-NOTES "Issue 3" line; `proposed.diff` regenerated |
| 2 | nit | PROPOSAL D6: "area-ID retention" (was "active-area retention"); "X2's matched-area survival and spurious-merge thresholds"; "Observed area-ID retention 0.95 (fastapi) and 0.85 (axum) stays in STATUS (STATUS's row "Area-ID retention" is this ratio; the council's wording named the old row label, which this round renames in STATUS)". PROPOSAL D8 header-line-3 row: "area-ID retention and Jaccard figures reported" | PROPOSAL D6, D8 |
| 3 | nit | PROPOSAL status "round-5 draft (round-4 issues addressed, `round-4-response.md`; round 3 in `round-3-response.md`; …)"; DRAFT-NOTES second "Round-4 judgement calls" heading is now "Round-5 judgement calls (round-4 issues; see `round-4-response.md`)"; line-map label "(proposed file; line numbers unchanged since the round-3 draft)" | PROPOSAL line 3; DRAFT-NOTES headings |

Also: `docs/STATUS.md` (main-agent file, not the design) rows now say "Area-ID retention (area IDs still active after the refresh)", "Retained areas' file-set Jaccard" and "93 retained areas", so the Phase 4 ratio is no longer called "surviving". STATUS `:51` bubblewrap wording is untouched (a follow-up carried from round 4). The DRAFT-NOTES digest table is refreshed. `docs/scout-design.md` is untouched.

## Digests (`shasum -a 256`, recomputed after the last edit)

- `scout-design.base.md` = `docs/scout-design.md`: `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4`
- `scout-design.proposed.md`: `75ef1e66521b1f129e872f651bd0f8bb9989e38867fb5335ec5ad43158fb3061`
- `proposed.diff`: `73da0f865fb72158ee5ee3e367b6b5dfcb7f8e0945c2e10144f121145c1eb6c1`
- `requirements.proposed.diff`: `4e7adaccaa3ffe0e627fdd06c068a8469afbefc5a6dd09d985ca73803ab927c3`
- `PROPOSAL.md`: `5751c1eed1cd29c91d208b0cee0da90b8afea1a6cd5449b0d2a5a5cf00ef61f4`
- `DRAFT-NOTES.md`: `dfff4e8412801349e0ec9bce1f7716087e27740f46977af17ed6fd44eca0dd64` (after the digest-table refresh)
