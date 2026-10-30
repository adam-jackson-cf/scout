# X2d RESULTS review — round 2 (innovation council, CLAUDE.md experiment rule 6)

Reviewed files (SHA-256): `RESULTS.md` `3bbd3133275c0281394c7ca130e95123f6eaf66d03a587aa2d16ab54615b87c3`

Scope: the changes made to `evaluation/x2d-merge-association-pilot/RESULTS.md` (now 813 lines) after `reviews/results-review-1.md` (15 issues, R1-01 to R1-15). Covered:
- every changed line, and every figure the fixes added, checked against the read-only data in `../scout-x2d/evaluation/x2d-merge-association-pilot/data/` (`<data>`);
- the drafter's three departures from the round-1 wording;
- whether the go/no-go section (§1.4) is still correct;
- the `docs/hypothesis-ledger.md` HD6 change, as a side check.

Method: the lead re-read each changed line. Every number below was checked against data read in rounds 1 and 2. No shell, no model call, no `.codegraph` access.

## 1. Verification of the round-1 fixes

| Round-1 issue | Where it now stands | Check | Result |
|---|---|---|---|
| R1-01 baselines built in; 0/38 decisive | lines 30, 35, 613 | text as specified; 0/38 and 0/35 match `regate/gates.json` `.baselines.B*`; `[INFERENCE]` marks the mechanism | fixed |
| R1-02 legacy statement | lines 28, 35, 39, 613 | present beside the re-gate figures | fixed |
| R1-03 P2 fragility and competing explanations | lines 29, 51, 539–542 | 8/11 = 0.727 and 2/12 = 0.167 are right. `p2_items` confidences are ctl-26 0.31, ctl-70 0.31, ctl-10 0.33, ctl-41 0.33. 12/35 = 0.343 is right (ctl-28 0.67, ctl-35 0.63, ctl-60 0.65 are stratum-1 `below_t_choice`). But "any single change fails it" is too broad. | **issue R2-01** |
| R1-04 raw path pull | lines 31, 371 | 6/35 items: ctl-3, 33, 60, 65 unresolved; ctl-75 `none`; ctl-21 assigned. Compared with rev 12's 3/20. | fixed |
| R1-05 abstention population and thresholds | lines 41–47 | 15/2 and 13/2 match §6.8. ctl-3, 26, 34 are `t_choice` misses; ctl-6, 25 are `t_cov` misses (`arms-p1/V-sem.jsonl`). Memo `:104` is cited correctly. | fixed |
| R1-06 area denominators | lines 48, 480 | 26 / 19 areas and 13/19 are right; "two largest areas … (five areas have more than one item)" is right | fixed (line 480 punctuation is awkward but not misleading) |
| R1-07 power `u` | lines 248, 250, 304 | the `mode` values and the 14 `u_mean` values match `power.json` (rounded to 3 decimals) | fixed |
| R1-08 ablated areas, overlap, ctl-9 | lines 439, 489, 507 | 16 areas, clone counts and titles match `p2_items` and `catalogue.json`; ctl-9 is the dup control of `0:373fe757150aa6cc` | fixed |
| R1-09 obligations table | lines 791, 800, 809, 812 | there is no `labels-check.json` under `regate/rounds/0/`; the `PREREGISTRATION.md:158` and `:71` citations are right; the unmeasured list now includes the re-gate κ | fixed |
| R1-10 baselines on the labelled sets | line 192 | 3/20 = 0.150 [0.052, 0.360] (ctl-1, ctl-23, ctl-31) | fixed |
| R1-11 reproduce command | lines 708–739 | the script reads fields that exist (`control-key` `area_id`, `units` `stratum`, `V-sem` `raw_choice`/`confidence`, `B*` `area_id`, `labels` `acceptable`/`agreement`); the counts it prints match §4.3 | fixed |
| R1-12 costs | lines 27, 52, 587 | $1.377808 as a lower bound with 7 unknown, marked `[INFERENCE]`; 1.7× and 10.7× are right (V20, round 1) | fixed |
| R1-13 numeric slips | lines 138, 160, 173, 205, 211, 346 | 0.92, [0.250, 0.842], [0.008, 0.227], the column rename, and "5 of the 12 clones" (ctl-7, 11, 22, 44, 45) | fixed |
| R1-14 duplicate-check scope | lines 67, 98, 615 | 261 of 1,268 pairs, consistent with `onboard/main.go:25,1504-1514` | fixed |
| R1-15 `regate/` path | line 6 | defined | fixed |
| Ledger (side check) | `docs/hypothesis-ledger.md:32,38` | 0/38 is given as decisive, with 0/35 in stratum 1 | consistent |

## 2. The drafter's departures

1. **R1-13, the parenthesis on line 160.** It now reads "Two corrections († below; and ctl-23's confidence, 0.92 … against the record's .91): no other mismatch". It is accurate: the † covers the three B cells as one correction, and the confidence is the second. Accepted.
2. **R1-08, the `catalogue-p2` sentence placed before the §6.6 table (line 439).** The meaning is unchanged and the citation `reviews/round-18.md:81` is kept. Accepted.
3. **R1-09, only the status cells replaced.** The obligation labels are unchanged and the replaced cells carry the full content. The re-gate κ addition to the "Unmeasured within what ran" list is present (line 791). Accepted.

## 3. Issues and exact fixes

### R2-01 (low–moderate): "Any single change fails it" overstates how fragile the P2 PASS is
**Problem**
- This wording came from the round-1 fix text for R1-03. It is false as a general claim.
- Changes on the other three positives would keep the PASS:
  - ctl-42 or ctl-68 changing from an assignment to `none` gives 10/11;
  - the abstained ctl-23 answering `none` gives 10/12 answered with abstention 0/12.
- Only a change to one of the nine `none` answers fails P2 (to an assignment: 8/11 = 0.727; to `unresolved`: abstention 2/12 = 0.167).
- The same overstatement is repeated at line 51.

**Fix**
- Line 29: replace "It passes by one answer: 8 of 11 would be 0.727. Any single change fails it: one `none` becoming an assignment gives 8/11 = 0.727, and one becoming `unresolved` gives abstention 2/12 = 0.167 > 0.15." with "It passes by one answer: if any one of the nine `none` answers changed, P2 would fail. Becoming an assignment gives 8/11 = 0.727; becoming `unresolved` gives abstention 2/12 = 0.167 > 0.15."
- Line 51: replace "(any single changed answer fails it, §1.2)" with "(any one of its nine `none` answers changing fails it, §1.2)".

## 4. Go/no-go (§1.4) after the fixes

Still correct:
- **"No GO" is a rule.** It follows from the §14 map at `PREREGISTRATION.md:253`.
- **"[INFERENCE] no full trial of this instrument as frozen"** follows from the abstention miss:
  - 15/35 = 0.429 [0.280, 0.591];
  - 12/35 even if the near-threshold abstentions are counted as answered;
  - point estimates above 0.15 under every one-per-area selection.

  It is scoped to the frozen instrument and to legacy files, and HD6 stays `inconclusive`.
- **The evidence items now match the data.** They cover the two-threshold mechanics, the 26/19 area counts, the P2 fragility and the lower-bound cost.
- **R2-01 does not change the go/no-go.** It only corrects how fragile the P2 PASS is described as being.

Outside RESULTS (not counted): `docs/hypothesis-ledger.md:40` (the HD7 history line) still says the check "compares only areas of one scope (261 pairs …)". If a corrective history line is added, it should use the R1-14 wording.

## 5. Verdict

1 ISSUES
