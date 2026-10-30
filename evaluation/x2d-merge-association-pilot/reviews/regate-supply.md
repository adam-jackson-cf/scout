# X2d re-gate: fresh supply (membership only; rev 13, R-053; regenerated after council round 17, R17-01 and R17-03)

Produced from the frozen membership, catalogue and units data only. No Jev score, label or model output was read, and nothing was written under `../scout-x2d`. The round-17 version (R17-01: presented paths and rev-12 target directories excluded; R17-03: item dates recorded) replaces the first record (P1 strata 34/2/4); the pool, the 40 P1 clusters and the 40 P2 clusters are the same, the P1 placements differ.

## Command

```
# (signoff.json moved aside for the run and restored byte-identical; SHA-256 578a2462…2e61 before and after)
cd /Users/adamjackson/Projects/inflight/scout && source .tools/activate
X2D_SKIP_SIGNOFF=1 X2D_JEV_ENDPOINT=http://127.0.0.1:9/ \
X2D_DATA=/Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data \
.tools/venv/bin/python evaluation/x2d-merge-association-pilot/build_controls.py \
  --repo /tmp/scout-x2c-smoke/_src/fastapi --regate-supply \
  --out evaluation/x2d-merge-association-pilot/reviews/regate-supply.json
```

Full table: `reviews/regate-supply.json`.

## Funnel (clusters seen -> fresh pool)

| origin | clusters | used as rev-12 source | path twin | PR episode | guide `#N` PR | no PR number | not all files at C0 | not one owner | same PR as earlier | **pool** |
|---|---|---|---|---|---|---|---|---|---|---|
| D (remaining eligible D units) | 20 | 6 | 1 | 8 | 0 | 0 | 0 | 1 | 1 | **3** |
| G (non-guide clusters) | 59 | 0 | 1 | 34 | 12 | 0 | 0 | 3 | 2 | **7** |
| P (reserve P units) | 30 | 26 | 0 | 0 | 0 | 0 | 0 | 4 | 0 | **0** |
| in_no_unit (C0-owned files in no unit) | 553 | 0 | 0 | 0 | 0 | 55 | 42 | 12 | 325 | **119** |
| **total** | | | | | | | | | | **129** (129 PRs, 69 areas) |

The reserve sets alone give 10 candidates; the 40 + 40 minimums need the `in_no_unit` clusters.

## Exposure recorded for R17-01

Presented paths of the round-0 controls: 83. Rev-12 P1 target directories: 17. Directory candidates skipped as rev-12 targets, summed over the final trial's 40 placements: 662. Placements rejected for a presented path in the final trial: 0. Counts are in `regate-supply.json` `exposed`.

## P1 candidates by stratum (clusters whose area has a free, non-excluded target directory of that kind, before the removal sets and the leak check)

| stratum | can host |
|---|---|
| 1 | 126 |
| 2 | 10 |
| 3 | 0 |

Planned P1 (leak-checked, targets cycled 1,1,1,2,3, falling back when none is free): **40 of 40 items, achieved strata 1: 37, 2: 0, 3: 3** (minimum: 40 items, at least 20 in stratum 1: met). Strata 2–3 are small and report-only.

## P2

16 areas own at least 2 fresh candidates after P1. Planned: **40 of 40 clones in 16 ablated areas** (clones per area 4, 4, 4, 3, 3, 3, 2 × 9, 1), origins D 2, G 2, in_no_unit 36. Sizing: round-0 yield 4 direct-label positives in 12 clones (1/3); P(X ≥ 8) = 0.98 (0.82 at 1/4) for independent clones, about 0.88 (0.71) with yield fully correlated within an area [INFERENCE]. The direct-label positive yield is unmeasured until labels exist; construction counts are met, and a shortfall is caught by `regate-jev`'s pre-Jev check.

## Population (R17-03)

All 40 planned P1 items are `in_no_unit` clusters with PRs below #10000, dated 2019-03-03 to 2022-09-01 (2019: 22, 2020: 9, 2021: 5, 2022: 4). Of the 40 P2 clones, 36 are `in_no_unit` clusters of the same kind (2019: 15, 2020: 11, 2021: 3, 2022: 7) and 4 are dated 2023–2024 (D 2, G 2). C0 is dated 2024-12-30.

## Result

Both minimums are met (40 P1 items, 37 in stratum 1; 40 P2 clones). `problems: none`. No stop-and-report.
