# Round 5 response — revision 2.2

Proposed file: 462 lines, SHA-256 `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4`. Base, `docs/scout-design.md` and Appendix 1–7 untouched. Only the two cited lines edited.

| # | Fix | Where |
|---|---|---|
| 1 | Cap order now covers every question family: triage units in baseline order (ties by issue id; `changed-in-PR` = 1 outside PR runs), then rubric functions, commit and unowned-file units in a stable unit-key order (path then function start line; commit hash; file path); units a family does not reach are counted in that lane's health as `n/N`. | L157 |
| 2 | REQ-29: independence confined to the interim draw; the audit is stated as a probability sample with a stored inclusion probability per unit. | `SCOUT_REBUILD_REQUIREMENTS.md:52` |

Files changed outside the proposed file: `proposed.diff`, `DRAFT-NOTES.md`, `round-5-response.md`, `SCOUT_REBUILD_REQUIREMENTS.md`.
