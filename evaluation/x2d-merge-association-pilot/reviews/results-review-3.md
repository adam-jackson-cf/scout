# X2d RESULTS review — round 3 (innovation council, CLAUDE.md experiment rule 6)

Reviewed files (SHA-256): `RESULTS.md` `8217504127ff96f8dc6c2a581cb753c399437a888cce78ce87674406e9a001da`; `docs/hypothesis-ledger.md` `10f7eeb5d89f67e2304926f9d6bcf2c0990e4ffc0987b82153af2cef8b1d8903`

Scope: the R2-01 fix in `evaluation/x2d-merge-association-pilot/RESULTS.md` (lines 29 and 51), and the corrective HD7 History line in `docs/hypothesis-ledger.md`. The lead re-read each line against the data checked in rounds 1–2. No shell, no model call, no `.codegraph` access.

## 1. Checks

| Item | Text now | Check | Result |
|---|---|---|---|
| RESULTS:29 (R2-01) | "It passes by one answer: if any one of the nine `none` answers changed, P2 would fail. Becoming an assignment gives 8/11 = 0.727; becoming `unresolved` gives abstention 2/12 = 0.167 > 0.15." | Matches the round-2 fix word for word. The claim is limited to the nine `none` answers, so it no longer covers ctl-42, ctl-68 or ctl-23. The figures are right: 8/11 = 0.727; 2/12 = 0.167 > the 0.15 cap (`regate/gates.json` `.bars.P2.abstention_max`). The four near-cut confidences (0.31, 0.31, 0.33, 0.33) still match `p2_items`. | fixed |
| RESULTS:51 (R2-01) | "P2 passes by one answer from 12 positives (any one of its nine `none` answers changing fails it, §1.2)" | Matches the round-2 fix and is consistent with line 29. | fixed |
| `docs/hypothesis-ledger.md:38` (HD7 History) | "HD7: still proposed; evidence wording corrected (X2d RESULTS review, R1-14): the onboarding duplicate check runs only on the 261 same-scope area pairs whose titles share a token (of 1,268 same-scope pairs), and never on cross-scope pairs (R-054)." | Uses the R1-14 wording, which matches `consensus-report.json` `.jev.duplicate_pairs_checked` 261 and `onboard/main.go:25,1504-1514`. The HD7 state stays `proposed`. It is placed newest-first, as the rest of the History section is. The original line (`:41`) is kept, so the history stays append-only. | fixed |

## 2. Go/no-go

Nothing in §1.4 changed in this round except item 5, which is now worded correctly. The round-2 assessment stands: "No GO" follows from the rule, and "[INFERENCE] no full trial of this instrument as frozen" is supported by the abstention miss and correctly scoped.

## 3. Issues

None.

## 4. Verdict

ZERO ISSUES — RESULTS accepted
