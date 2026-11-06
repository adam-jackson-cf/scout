# Response to round 4 — revision 2.3 (round-5 draft)

All five round-4 issues are fixed as text-only changes. X2's decision rule, A1 and HD5 are unchanged. Line numbers refer to `scout-design.proposed.md` (still 466 lines).

| # | Sev | Fix | Where |
|---|---|---|---|
| 1 | minor | The Phase 4 ratio is renamed "area-ID retention" and the council's HD5 clause added: "this ratio counts kept areas and is not X2's survival, which HD5 (`docs/hypothesis-ledger.md`) operationalises as prior areas matched by a next feature". The surviving areas' Jaccard becomes "retained areas'". X2's thresholds are now "matched-area survival". A1's "0.8 survival" is untouched (it is X2's). `:3` "X2-dependent figures" became "area-ID retention and Jaccard figures". | `:3`, `:7` item (4), `:372`; PROPOSAL D6 and "Explicitly not changed" (A1 parenthetical: "is X2's matched-area survival (HD5), not the Phase 4 ratio, which the design now calls area-ID retention"; "observed area-ID retention"); DRAFT-NOTES |
| 2 | nit | "PR runs" | `:7` item (4), `:189`, `:425`; PROPOSAL D4; DRAFT-NOTES Issue 1 |
| 3 | nit | "`comparison` is not persisted in Scout's state (only in each run's `scout-assessment.json` artifact)" | `:186`; PROPOSAL D4 |
| 4 | nit | "install bubblewrap (Linux runners)" | `:197`; PROPOSAL D8 |
| 5 | nit | PROPOSAL D6 "does not separate them"; DRAFT-NOTES Issue 4 now says one retention ratio and one pooled Jaccard distribution | PROPOSAL, DRAFT-NOTES |

Also: DRAFT-NOTES gains a "Round-4 judgement calls" section, a round-5 header and refreshed digests. `requirements.proposed.diff` is unchanged (none of its hunks touch the edited sentences). `proposed.diff` regenerated (236 lines, 42 insertions, 38 deletions). `docs/scout-design.md` is untouched. STATUS (main-agent files, not the design): added `cdbf56c` (`pr_effort` unmeasured reason when PR metadata holds no merged PRs) and dogfood run 39969572544 (`v0.1.2`, exit 0, all scanner lanes, `scout/state` published, `pr_effort` unmeasured: no merged PRs) to the Published bullet.

## Digests (`shasum -a 256`, recomputed after the last edit)

- `scout-design.base.md` = `docs/scout-design.md`: `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4`
- `scout-design.proposed.md`: `ed0faacf09bd6ba6cb5b4bc1c7231d29f54ef9b7d43ea6120653ce93a2efe48b`
- `proposed.diff`: `e2d72e04a156b12492215c2292d1e7c7c15f4f73a34591d0ad22bd683cda37b9`
- `requirements.proposed.diff`: `4e7adaccaa3ffe0e627fdd06c068a8469afbefc5a6dd09d985ca73803ab927c3`
- `PROPOSAL.md`: `d227702ce129867e9f3cd2d7df7469ffc499589054fa520b6d7686125ffc6d69`
- `DRAFT-NOTES.md`: `8df33727907fcf0a2e082d6f3d701550764bcc4ce9d14951774fb229b5f4075d`
