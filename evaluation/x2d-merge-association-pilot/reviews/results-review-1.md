# X2d RESULTS review — round 1 (innovation council, CLAUDE.md experiment rule 6)

Reviewed files (SHA-256): `RESULTS.md` `2a8beea1d23e0fc9b94dd22d9943b4e0365584c65508cee184fc598f40160c0f`; `PREREGISTRATION.md` `b9110486014bca580be3f14629c0beda54f35109a2c89ca5ca04ad95085140a8`; `reviews/pilot-round-0.md` `cb16eca0096f441d40cbff6ba0569362ca3138c4477b8eac27d2621d312ecb94`

Scope: `evaluation/x2d-merge-association-pilot/RESULTS.md` (764 lines) for the stopped X2d pilot.

Checked against:
- `PREREGISTRATION.md` rev 13, especially `:71`, `:111-114`, `:129`, `:158`, `:253`, `:255`;
- `reviews/pilot-round-0.md` (the scout-repo copy; the `../scout-x2d` copy ends before the re-gate record, and RESULTS correctly cites the scout copy);
- `reviews/pilot-round-0-gate-stop-memo.md`;
- `reviews/round-16.md`, `reviews/round-17.md` §3–§4, `reviews/round-18.md` §3.6 and the run steps;
- the read-only data in `../scout-x2d/evaluation/x2d-merge-association-pilot/data/` (`<data>`).

Method:
- The lead read the data directly.
- An `innovation-challenger` (GPT) and an `experiment-peer` (Opus) ran independent audits.
- Every issue below was re-verified by the lead against the cited file before inclusion.
- No shell, no model call, no `.codegraph` access.

## 1. Verification table

| # | RESULTS claim (line) | Source checked | Observed | Match |
|---|---|---|---|---|
| V1 | Re-gate P1 stratum-1 accuracy 14/20 = 0.700 [0.481, 0.855] (30, 316) | `<data>/rounds/0/regate/gates.json` `.gates.P1.value.stratum1.accuracy_answered` | k 14, n 20, [0.48103, 0.85452] | yes |
| V2 | Abstention 15/35 = 0.429 [0.280, 0.591] (30) | `.stratum1.abstention_rate` | [0.27985, 0.59143] | yes |
| V3 | B1 and B2 0/38 [0.000, 0.092] (decisive); 0/35 [0.000, 0.099] | `.baselines.B*.all_eligible`, `.stratum1` | as stated | yes (summary quotes 0/35: issue 1) |
| V4 | Path-following 1/20 [0.009, 0.236], 6/35 [0.081, 0.327]; recount from `p1_items` (ctl-3, 21, 33, 60, 65, 75) | `.stratum1.path_following_*`; `p1_items` | as stated | yes |
| V5 | All eligible: 15/21, 17/38, 15/38, 14/38; path-following 1/21 [0.009, 0.227] (336) | `.all_eligible` | lower bound of 1/21 is 0.008456 | **no** (issue 13) |
| V6 | Strata 37/0/3; stratum 3 1/1 answered, 2/3 abstained; excluded ctl-38, ctl-64 | `.strata`, `.excluded.gold_excludes_A` | as stated | yes |
| V7 | P2: 9/11 [0.523, 0.949]; 1/12 [0.015, 0.354]; 9/12 [0.468, 0.911]; gap 11 / new 1; origin D 2 / `in_no_unit` 10 | `.gates.P2.value` | as stated | yes |
| V8 | P2 retained-covered 18/0/3/7 of 28; `derived_vs_direct` 13/15/12 | `.gates.P2.value.retained_covered`, `.derived_vs_direct` | as stated | yes |
| V9 | Clones per ablated area 4,4,4,3,3,3, nine areas with 2, one with 1; 8 areas hold the 12 positives; per-positive confidences | recount of `p2_items` | as stated | yes |
| V10 | §6.5 year tables (P1 and P2) | recount of `p1_items` / `p2_items` dates | as stated | yes |
| V11 | §6.6: 26 areas; 13 all-correct / 6 all-wrong / 0 mixed; 13/19; 7/26; 12/26; 10/27 | recount of `p1_items` grouped by `area_id` | as stated | yes (wording: issue 6) |
| V12 | §6.8 reasons: P1 15/2, stratum 1 13/2, P2 positives 1, retained 7 | `p1_items` / `p2_items` `reason` | as stated | yes |
| V13 | Rev-12 gates: K2 0.5848, K3 0.0515, K4 1.0, K5 0.04, K6 UB90 0.0318, planted_new 3/3, P1 0.45/0.05/0.05, P2 4/7, real_none_proxy 39/98 | `<data>/rounds/0/gates.json` `.gates` | as stated | yes |
| V14 | **Corrected B cells:** ctl-1, ctl-3, ctl-36 → `e287e9` | `rounds/0/arms/B1.jsonl` and `B2.jsonl` lines 68, 71, 53: `assigned e287e950ae41` | correction right (B0 reads `none`, so the record's "(none)" was B0, not B2) | yes |
| V15 | Raw = B on 4/20 (ctl-23, 24, 33, 39); raw = B ≠ A on 3/20 | `arms-p1/V-sem.jsonl` `raw_choice` against B2 | as stated | yes |
| V16 | ctl-23 confidence ".91" (167) | `arms-p1/V-sem.jsonl:10` `confidence` 0.92 (record `:126` has .91) | 0.92 | **no** (issue 13) |
| V17 | 4/7 = 0.571 [0.251, 0.842] (132) | Wilson z = 1.96 | [0.2505, 0.8418] | **no** (issue 13) |
| V18 | Power: `n4_max` 0.054, `all_pass`, P3_u0 894/1000 [0.873, 0.912], cell counts | `<data>/power.json` | as stated; the `u` column of the N4 cells is not their abstention (`u_mean` 0.050–0.150) | values yes; labelling **no** (issue 7) |
| V19 | **Totals:** known Jev cost $1.377808 = 0.185375 + 1.098534 + 0.093899; 9,759 answers; **7 unknown** | score files; `rounds/0/jev/scores.jsonl:7757-7763` (7 × `type: failure`, `cost_known: false`) | as stated | yes (§1.4 omits the caveat: issue 12) |
| V20 | Per-request prices: fan-out $0.670496 / 7,296; `which_area` $0.211143 / 384; `covered` $0.128194 / 192 | `rounds/0/jev/scores.jsonl` | as stated; fan-out ≈ 1.7×, other ≈ 10.7×, overall ≈ 2.3× the $0.000055 assumption | values yes; "≈ 10×" **no** (issue 12) |
| V21 | **Unmeasured split-half ARI**: three rotations unmeasured; policy (e) excludes 1,217 files | `onboard/quality.json` `.split_half_consensus_ari.rotations[*].status`, `.unmeasured` | as stated; the `[INFERENCE]` fits the error text | yes |
| V22 | **Stale wording**: `quality.json` intruder status "labels pending"; `PREREGISTRATION.md:3` "has **not been built**" | the two files | present, and disclosed in RESULTS (§3, §8.7.5–6) | yes |
| V23 | Intruder 40/40 [0.912, 1.000], κ 1.0, unmeasured 0 | `quality.intruder.json` | as stated | yes |
| V24 | HD6 reason string | `PREREGISTRATION.md:253` template; `docs/hypothesis-ledger.md:38` | identical | yes |
| V25 | Duplicate check "261 same-scope pairs" (61, 92, 601) | `consensus-report.json` `.jev.duplicate_pairs_checked` 261; `onboard/main.go:25,1504-1514` | the 261 are the same-scope pairs whose titles share ≥ 1 token, of 1,268 same-scope pairs | value yes; description **no** (issue 14) |
| V26 | Labels-check "round 0 and re-gate reported" (751) | `<data>/rounds/0/regate/rounds/0/` has `labels.jsonl` and `gold.jsonl` but no `labels-check.json` | there is no re-gate labels-check | **no** (issue 9) |

## 2. Issues and exact fixes

Line numbers are those of the current `RESULTS.md`.

### R1-01 (high): the claim that "the answers came from content, not path" is unsupported, and the baseline conjunct has the wrong denominator
**Problem**
- Line 35 states, without `[INFERENCE]`, that the baselines recovering nothing shows the answers came from content. This does not follow:
  - the rev-13 build rejects any target directory for which B1 or B2 would still answer A (`PREREGISTRATION.md:112`, `leak_check=True`);
  - the pre-Jev check already showed 0/38 before any Jev request (`reviews/pilot-round-0.md:222`).
- Each item is also shown to Jev "with their text and PR at the adding commit" (`PREREGISTRATION.md:111`), so it carries a file name and the PR title and body (e.g. `rounds/0/units.jsonl` ctl-23: `members`, `pr_title`, `pr_body`).
- Lines 30 and 599 give "baselines 0/35". The decisive conjunct is over ALL eligible items, 0/38 (`PREREGISTRATION.md:112`; `.baselines.B*.all_eligible`).

**Fix**
- Line 35: replace "The folder-based baselines recovered nothing on these items, so the answers it did give came from content, not path (stratum-1 path-following 1/20 answered)." with:

  "The folder-based baselines recovered no acceptable area (0/38 eligible items), but that is built in: the rev-13 build rejects any target directory for which B1 or B2 would still answer A (`PREREGISTRATION.md:112`), and the pre-Jev check showed 0/38 before any Jev request (`reviews/pilot-round-0.md:222`). [INFERENCE] The correct answers did not follow the folder (no acceptable area is a sibling area, `sibling_in_acceptable` 0; composed path-following 1/20). Which cues produced them is not measured: each item also showed Jev its file name and the PR text (`PREREGISTRATION.md:111`), and the raw first choice named the sibling area on 6/35 items (§6.4)."
- Line 30: replace "baselines 0/35 (bar ≤ 0.20)" with "baselines 0/38 over all eligible items, the decisive conjunct (0/35 in stratum 1; bar ≤ 0.20; zero by construction, §1.3)".
- Line 599: replace "baselines 0/35" with "baselines 0/38 (all eligible; 0/35 in stratum 1)".

### R1-02 (moderate): the legacy-population statement is missing beside re-gate figures
**Problem**
- `PREREGISTRATION.md:111` requires that "`RESULTS.md` states this beside every re-gate figure".
- §1.2 (lines 28–31), §1.3 (line 35), §1.4 (line 39) and the §9 HD6 row (line 599) give re-gate figures without the statement.

**Fix**
- Line 28: replace "**One repaired re-gate on fresh items (rev 13).**" with "**One repaired re-gate on fresh legacy items (rev 13; 2019–2022 files, not window-era files, §0.3).**".
- Line 35: after "on files that sit in a folder owned by another area" insert " (legacy 2019–2022 files, not window-era files)".
- Line 39: after "accuracy on answered items 0.70 against 0.80" insert " (re-gate, legacy 2019–2022 files)".
- Line 599: after "re-gate P1 stratum 1" insert " (legacy 2019–2022 items, not window-era files)".

### R1-03 (moderate): the P2 PASS is fragile, and the round-17 §4 competing explanations are not carried
**Problem**
- Round 18 §3.6 says "The round-17 §4 competing explanations still apply" (`reviews/round-18.md:85`; list at `reviews/round-17.md:172-174`).
- RESULTS omits several of them: run-to-run noise near the thresholds, cues in the PR text, and old central files.
- RESULTS shows only one way P2 flips (8/11). It does not say that one more abstention also fails P2 (2/12 = 0.167 > 0.15).
- Four of the nine `none` answers sit within 0.03 of the 0.30 cut (`p2_items`: ctl-26 0.31, ctl-70 0.31, ctl-10 0.33, ctl-41 0.33).

**Fix**
- Line 29: append "Any single change fails it: one `none` becoming an assignment gives 8/11 = 0.727, and one becoming `unresolved` gives abstention 2/12 = 0.167 > 0.15. Four of the nine `none` answers have `1 − p_true` of 0.31–0.33 (ctl-26 and ctl-70 at 0.31, ctl-10 and ctl-41 at 0.33), within 0.03 of the 0.30 cut (§6.9)."
- §6.9: after the positives table (after line 527), add the bullet:

  "- **Competing explanations carried over (round-17 §4, `reviews/round-17.md:172-174`; round-18 §3.6).**
    - For the P2 PASS: run-to-run noise near the provisional threshold (four `none` answers within 0.03 of the cut; [INFERENCE] Jev's repeatability was not measured, so a re-run could fail P2); cues in the PR text, which Jev sees with each legacy item (`PREREGISTRATION.md:111`); old, central files; labellers seeing more context than Jev (§0.7).
    - For the P1 FAIL: the untuned threshold (§6.8), cross-scope near-duplicate areas (§0.9), legacy content (§0.3), topic cues in test-directory paths (§6.4) and run-to-run noise.
    - The P1 abstention conjunct is not near its bar. If the three stratum-1 abstentions with choice confidence ≥ 0.60 had been answered (ctl-28 0.67, ctl-35 0.63, ctl-60 0.65), abstention would still be 12/35 = 0.343."
- Line 45 (§1.4 item 5): after "P2 passes by one answer from 12 positives" insert " (any single changed answer fails it, §1.2)".

### R1-04 (moderate): §6.4 dismisses path pull although the raw choices show it
**Problem**
- Line 361 says "The data give no support for a path-pull story (1/20 composed)".
- In fact the raw first choice named the non-acceptable sibling area on 6/35 stratum-1 items (`p1_items`):
  - 4 abstained: ctl-3, ctl-33, ctl-60, ctl-65;
  - 1 answered `none`: ctl-75;
  - 1 assigned: ctl-21.
- 6/35 is close to rev 12's real path pull, raw = B ≠ A on 3/20 (§4.3).

**Fix**
- Line 361: replace "The data give no support for a path-pull story (1/20 composed), and no power to rank prefixes." with:

  "Composed path-following is rare (1/20), but the raw first choice named the non-acceptable sibling area on 6/35 stratum-1 items: 4 abstained (ctl-3, ctl-33, ctl-60, ctl-65), 1 was answered `none` (ctl-75) and 1 was assigned (ctl-21). That is close to rev 12's raw = B ≠ A rate of 3/20 (§4.3). The threshold turned most of this pull into abstention, so the data do not rule path pull out, and they give no power to rank prefixes."
- Line 31: replace "Path-following was low: 1/20 of answered items, 6/35 of unthresholded first choices." with "Path-following: 1/20 of answered items; 6/35 of unthresholded first choices, 4 of which abstained (§6.4)."

### R1-05 (moderate): §1.4 evidence item 1 mixes populations and the two thresholds
**Problem**
- "15 of the 17 re-gate abstentions" covers the 38 eligible P1 items only. In decisive stratum 1 the split is 13 and 2, and P2 adds another 8 abstentions (§6.8).
- The two thresholds act separately (`PREREGISTRATION.md:129`):
  - P1 abstentions come from `t_choice`;
  - the 5 wrong stratum-1 `none` answers and P2's `none` answers come from `t_cov`.
- The stop memo's sensitivity lowered **both** thresholds (`reviews/pilot-round-0-gate-stop-memo.md:104`).
- The "5 raw-A misses" are 3 `t_choice` abstentions (ctl-3, ctl-26, ctl-34) and 2 `t_cov` `none` answers (ctl-6, ctl-25).

**Fix**
- Line 41: replace item 1 with:

  "1. **Abstention under tuned thresholds.**
     - Of the 17 abstentions among the 38 eligible re-gate P1 items, 15 were `below_t_choice` and 2 `permutation_disagreement`; in decisive stratum 1 the split was 13 and 2 (§6.8).
     - The 5 wrong `none` answers in stratum 1 come from the covered threshold `t_cov`, not from `t_choice` (`PREREGISTRATION.md:129`).
     - Of the 11 round-0 P1 misses, 5 had raw choice A: 3 abstained under `t_choice` (ctl-3, ctl-26, ctl-34) and 2 were `none` under `t_cov` (ctl-6, ctl-25) (§4.3).
     - No threshold was tuned (rounds 1–8 never ran). The stop memo's sensitivity lowered both thresholds to 0.5 and found it would cancel three round-0 P2 `none` answers (`reviews/pilot-round-0-gate-stop-memo.md:104`).
     - [INFERENCE] P1's `none` errors and P2's `none` answers pull `t_cov` in opposite directions, while P1's abstentions depend mainly on `t_choice`. The data therefore cannot say whether tuned thresholds meet the 0.15 cap without breaking P2.
     - A new preregistration would have to calibrate both thresholds on separate labelled items first."

### R1-06 (moderate): the area denominators are misstated
**Problem**
- Line 42 says "20 answered items from 26 areas (19 answered areas)". The 35 eligible items come from 26 areas; the 20 answered items come from 19 (recount of `p1_items`).
- Line 470 says "the two multi-item areas `e287e9` and `9f08e6`", but five areas have more than one item: `e287e9` 5, `9f08e6` 3, and `3de70a`, `4d46c3`, `617fde` 2 each (§6.6 table).

**Fix**
- Line 42: replace with "2. **Enough answered stratum-1 items from more areas.** The 35 eligible stratum-1 items come from 26 areas and the 20 answered items from 19. Accuracy is 14/20 = 0.700 [0.481, 0.855]; one answered item per area gives 13/19 = 0.684 [0.460, 0.846] (§6.6)."
- Line 470: replace "Dropping the two multi-item areas `e287e9` and `9f08e6`" with "Dropping the two largest areas, `e287e9` (5 items) and `9f08e6` (3 items) (five areas have more than one item),".

### R1-07 (moderate): the power table labels selective-abstention cells as zero abstention
**Problem**
- §5 (line 240) defines the column as "abstention `u`", and the N4 rows show 0.0.
- In `power.json` those cells abstain selectively, so `u` is a random-abstention input, not the abstention rate:
  - modes are `b2_errors` (N4a), `own_errors` (N4b, N4b_rho0, N4c) and `own_errors_cap` (N4b_cap);
  - the realised `u_mean` is 0.050–0.150.

**Fix**
- Line 240: replace "abstention `u`" with "random-abstention input `u` (`power.json` `.u`). The N4 cells abstain selectively (`mode` `b2_errors` for N4a; `own_errors` for N4b, N4b_rho0 and N4c; `own_errors_cap` for N4b_cap), so their `u` = 0.0 is not their abstention rate".
- Rename the table header `u` to `u input`.
- After the table (after line 294), add: "Realised mean abstention (`u_mean`) of the selective cells, in table order: N4a 0.142, 0.131, 0.097, 0.051; N4b 0.150, 0.120, 0.101, 0.050; N4b_cap 0.142, 0.132; N4b_rho0 0.149, 0.120, 0.099; N4c 0.102 (`power.json` `.N4a[].u_mean`, `.N4b[].u_mean`, `.N4b_cap[].u_mean`, `.N4b_rho0[].u_mean`, `.N4c.u_mean`)."

### R1-08 (moderate): preregistered P2 disclosures and a round-18 §3.6 fact are missing
**Problem**
- `PREREGISTRATION.md:113` carries the R10-02 disclosures over to the re-gate, including "the ablated areas and clones per area". §6.9 gives counts only and names only the 8 areas that hold positives, not all 16.
- Round-18 §3.6 (`reviews/round-18.md:81`) requires the fact that `e287e9`, `9f08e6`, `3de70a` and `617fde` are also P2 ablated areas, but on the separate `catalogue-p2`. It is absent.
- `reviews/round-18.md:82` names "ctl-24/ctl-9" for `tests/test_validate_response_recursive`, but §6.7 lists only ctl-24. ctl-9 is the dup control of the same source unit, `0:373fe757150aa6cc` (`control-key.jsonl:9,24`).

**Fix**
- §6.9, "Donors by origin" bullet (line 497): after "then nine areas with 2 and one with 1;" insert:

  "the 16 ablated areas, with clones each: `3de70a` Parameter declarations and request input extraction 4, `e287e9` OpenAPI schema and interactive documentation 4, `9a4a0e` Dependency injection 4, `9f08e6` Custom responses 3, `f36cd1` HTTP routing and route registration 3, `617fde` Response model validation and encoding 3, `063f15` Client generation tutorials 2, `30399e` Documentation site build and tooling 2, `73232e` Repository metadata and contribution configuration 2, `86949d` Python types tutorials 2, `a0f998` Request file upload tutorials 2, `70d6ec` Handling errors tutorials 2, `c6a286` WebSocket tutorials 2, `c79f1b` Testing tutorial examples 2, `dc7550` Response classes and status behavior 2, `2266c6` Query parameter string validations tutorials 1 (`p2_items` `ablated_area_id`; titles from `<data>/catalogue/catalogue.json`);"
- §6.6, after the first sentence (line 437): insert "`e287e9`, `9f08e6`, `3de70a` and `617fde` are also re-gate P2 ablated areas, but only on the separate `catalogue-p2`; the P1 items were asked on `catalogue-p1` (`reviews/round-18.md:81`)."
- §6.7 table, row ctl-80, column "rev-12 item(s) that came from it": replace "ctl-24" with "ctl-24 (and its dup control ctl-9, same source unit `373fe757`)".

### R1-09 (moderate): the §11 obligations table claims items are "reported" when they are not
**Problem**
- Line 751 says the labels-check was reported for round 0 **and the re-gate**. There is no re-gate labels-check (no `labels-check.json` under `<data>/rounds/0/regate/rounds/0/`), so re-gate κ is unmeasured, and §11 does not list it as unmeasured.
- Line 760 says the planted new-area figures were reported. The preregistered planted figures pool frozen-round traps (`pooled.planted_new`, `PREREGISTRATION.md:158`), and none were produced.
- Line 763 marks the round-10 role mix "not applicable". `PREREGISTRATION.md:71` states it unconditionally and it is membership-only. It was not computed: `k23.json` `.per_round."10"` and `prepass/rounds.json` carry no role field.

**Fix**
- Line 751: replace the status cell with "adjudication: round 0 and re-gate **reported** (§4.5; §6.1: 115 agreed / 5 adjudicated of 120); labels-check and κ: round 0 **reported** (§4.5), re-gate **unmeasured** (`regate-label` writes labels and gold only; no `rounds/0/regate/rounds/0/labels-check.json` exists); rounds 1–10 not applicable".
- Line 760: replace the status cell with "**not applicable**: the preregistered planted figures pool frozen-round traps (`pooled.planted_new`, `PREREGISTRATION.md:158`) and the real figures need window rounds; none ran. Gate analogues are reported: planted_new 3/3 (§4.1); re-gate P2 `none` on all positives 9/12 = 0.750 [0.468, 0.911] (§6.9)".
- Line 763: replace the status cell with "role mix **not computed** (membership-only, `PREREGISTRATION.md:71`; no stage that ran writes it: `k23.json` `.per_round."10"` has no role field); episode sensitivity **not applicable** (no window round)".
- §11, "Unmeasured within what ran" (line 740): after "the re-gate's stratum 2 (0 items);" insert " the re-gate labeller κ and labels-check (not run by `regate-label`);".

### R1-10 (low–moderate): §4.3 shows only Jev's side of the labelled-set sensitivity
**Problem**
- §4.3 reports V-sem at 11/20 against the labelled sets but omits the baselines on the same sets: 3/20 (ctl-1, ctl-23, ctl-31), as given in the stop memo's P1 sensitivity ("Baselines score 3 of 20 = 0.15").
- The corrected B cells (V14) are what make ctl-1 count.
- Without the baseline figure, a reader compares 0.55 against the baselines' 0.05.

**Fix**
- Line 187: after "which still fails 0.60." insert "On the same sets B1 and B2 each score 3/20 = 0.150 [0.052, 0.360] (ctl-1, ctl-23, ctl-31; the corrected B cells count here), still within the 0.20 bar (`reviews/pilot-round-0-gate-stop-memo.md`, P1 sensitivity)."

### R1-11 (moderate): §4.3 corrects the run record but has no reproduce command
**Problem**
- Line 6 promises that every figure not stored in a data file was computed "by the commands in §10".
- §4.3's re-derived counts and the B-cell correction to `reviews/pilot-round-0.md` have no command.
- These counts are raw = A 14/20, raw in the labelled set 17/20, raw = B 4/20, raw = B ≠ A 3/20, the labelled-set scores 11/20 and 3/20, stratum 1 2/11 and strata 2–3 7/9.

**Fix**
- §10: after the "P2 join" block (after line 693), add:

**Rev-12 P1 per item** (§4.3; rev-12 round 0; any Python 3):
```
X2D_DATA=$D python3 - <<'E'
import json, math, os
D = os.environ["X2D_DATA"] + "/rounds/0/"
rd = lambda f: [json.loads(l) for l in open(D + f) if l.strip()]
def W(k, n, z=1.96):
    p = k / n; d = 1 + z * z / n; m = (p + z * z / (2 * n)) / d
    h = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / d
    return f"{k}/{n} = {p:.3f} [{m - h:.3f}, {m + h:.3f}]"
ck = [r for r in rd("control-key.jsonl") if r["kind"] == "p1"]
U = {u["case_id"]: u for u in rd("units.jsonl")}
V = {a["case_id"]: a for a in rd("arms-p1/V-sem.jsonl")}
B1 = {a["case_id"]: a["area_id"] for a in rd("arms/B1.jsonl")}
B2 = {a["case_id"]: a["area_id"] for a in rd("arms/B2.jsonl")}
L = {l["case_id"]: set(l["acceptable"]) for l in rd("labels.jsonl") if "agreement" in l}
c = {}
def add(k, t): c[k] = c.get(k, 0) + bool(t)
for r in ck:
    i, A = r["case_id"], r["area_id"]; v, b, s, acc = V[i], B2[i], U[i]["stratum"], L[i]
    print(i, s, A[:6], v["output"], (v["area_id"] or "-")[:6], v["confidence"], (v["raw_choice"] or "-")[:6],
          (B1[i] or "-")[:6], (b or "-")[:6], sorted(a[:6] for a in acc))
    add("hit", v["area_id"] == A); add("abstained", v["output"] == "unresolved"); add("none", v["output"] == "none")
    add("raw_A", v["raw_choice"] == A); add("raw_in_labelled", v["raw_choice"] in acc)
    add("raw_B", v["raw_choice"] == b); add("raw_B_not_A", v["raw_choice"] == b and b != A)
    add("V_labelled", v["area_id"] in acc); add("B1_labelled", B1[i] in acc); add("B2_labelled", b in acc)
    add("s1", s == 1); add("s1_hit", s == 1 and v["area_id"] == A)
n = len(ck)
for k in ("hit", "abstained", "raw_A", "raw_in_labelled", "raw_B", "raw_B_not_A", "V_labelled", "B1_labelled", "B2_labelled"):
    print(k, W(c[k], n))
print("none", c["none"], "| stratum 1", W(c["s1_hit"], c["s1"]), "| strata 2-3", W(c["hit"] - c["s1_hit"], n - c["s1"]))
E
```
- Expected output: hit 9/20, abstained 7/20, raw_A 14/20, raw_in_labelled 17/20, raw_B 4/20, raw_B_not_A 3/20, V_labelled 11/20, B1_labelled and B2_labelled 3/20, none 2, stratum 1 2/11, strata 2–3 7/9.

### R1-12 (low): the cost reporting drops a caveat and overstates the price rise
**Problem**
- Line 46 gives "$1.378" as all Jev spend, without the lower-bound caveat that §7 carries (7 `cost_known: false` rows, `rounds/0/jev/scores.jsonl:7757-7763`).
- "Cost is not the obstacle" is a judgement but is not marked `[INFERENCE]`.
- Lines 27 and 573 say per-request cost was "≈ 10×" the assumption. By question type, fan-out was ≈ 1.7× ($0.000092), the other questions ≈ 10.7× ($0.000589), and overall ≈ 2.3× (V20; `reviews/pilot-round-0.md:47`).

**Fix**
- Line 46: replace with "- [INFERENCE] Cost is not the obstacle. Known Jev spend to this point is $1.377808, a lower bound because 7 cancelled requests have unknown cost (§7). The whole pilot was projected at ≈ $9.74 of Jev (R-052) against a $20 cap (§7)."
- Line 27: replace "(measured cost ≈ 10× the budget assumption)" with "(measured cost per request ≈ 1.7× the budget assumption for the fan-out and ≈ 10.7× for the other questions, §7)".
- Line 573: replace "The measured cost per request was ≈ 10× the budget assumption;" with "The measured cost per request was ≈ $0.000092 for the fan-out (≈ 1.7× the $0.000055 assumption) and ≈ $0.000589 for the other questions (≈ 10.7×);".

### R1-13 (low): numeric slips in reported figures
**Problem**
- Line 167: ctl-23's confidence is 0.92 (`arms-p1/V-sem.jsonl:10`), not .91. This contradicts "One correction … no other mismatch" at line 157.
- Line 132: the Wilson lower bound of 4/7 is 0.2505, which rounds to 0.250, not 0.251.
- Line 336: the lower bound of 1/21 is 0.008456 (`.all_eligible.path_following_answered.wilson95[0]`), so 0.008 and not 0.009. This contradicts line 114's "match the file to 4 decimals".
- Lines 199–203: the column "`none` answers" counts only `none` on gap positives. The P set has 3 `none` answers (ctl-7, ctl-11, ctl-45), and all 12 clones have 5.

**Fix**
- Line 167: ".91" → ".92".
- Line 157: replace "One correction (†" with "Two corrections (†; and ctl-23's confidence, 0.92 in `arms-p1/V-sem.jsonl` against the record's .91".
- Line 132: "[0.251, 0.842]" → "[0.250, 0.842]".
- Line 336: "1/21 = 0.048 [0.009, 0.227]" → "1/21 = 0.048 [0.008, 0.227]".
- Line 199: rename the header "`none` answers" to "`none` on gap positives".
- After line 203, add: "V-sem also answered `none` on 1 retained-covered P clone (ctl-7, gold `214b68`), so it answered `none` on 5 of the 12 clones."

### R1-14 (low): the duplicate-check scope is overstated, and HD7 rests on it
**Problem**
- Lines 61, 92 and 601 imply every same-scope pair was compared.
- The check compares only same-scope pairs whose titles share at least one token (`onboard/main.go:25,1504-1514`; `-dup-min-shared` default 1, not overridden in `run.sh`).
- That is 261 of the 1,268 same-scope pairs (22, 29, 21, 22 and 20 areas per scope).

**Fix**
- Line 61: replace "compared only areas of one scope (261 pairs, 0 flagged)" with "compared only same-scope pairs whose titles share at least one token (`onboard/main.go:25`): 261 of the 1,268 same-scope pairs, 0 flagged".
- Line 92: replace "0 duplicates in 261 same-scope pairs" with "0 duplicates in the 261 same-scope pairs whose titles share a token (of 1,268 same-scope pairs)".
- Line 601: replace "compared 261 same-scope pairs and flagged 0" with "compared the 261 title-token-sharing same-scope pairs (of 1,268) and flagged 0".

### R1-15 (low): the `regate/` path shorthand is never defined
**Problem**
- RESULTS cites `regate/gates.json`, `regate/rounds/0/...` and, in the §7 cost table, `regate/rounds/0/jev-p/scores.p*.jsonl`.
- No `regate/` exists under `<data>`. The files are under `<data>/rounds/0/regate/`, while the other §7 rows are relative to `<data>`.

**Fix**
- Line 5 (the Paths paragraph): after "`<study>` = `evaluation/x2d-merge-association-pilot/`." insert " `regate/` = `<data>/rounds/0/regate/` (so `regate/gates.json` is `<data>/rounds/0/regate/gates.json`)."

## 3. Go/no-go assessment (§1.4)

The section is honest and follows from the figures, once R1-01, R1-03, R1-05, R1-06 and R1-12 are applied.

- **"No GO" is a rule, not an inference.** The re-gate is not PASS for both gates, so the §14 map (`PREREGISTRATION.md:253`) gives `inconclusive` and no verdict exists. RESULTS states this correctly.
- **"[INFERENCE] no full trial of this instrument as frozen"** is supported, because the abstention miss does not depend on the uncertainty or the selection:
  - 15/35 = 0.429 [0.280, 0.591]: the lower bound clears the 0.15 cap;
  - 12/35 = 0.343 even if the three near-threshold abstentions are answered (R1-03);
  - 7/26 = 0.269 at best one item per area, though with an interval [0.137, 0.461] that reaches the cap. Clustering widens all of these.

  Accuracy (0.700 [0.481, 0.855]) is not decisive, and §1.3 says so.
- **It does not overclaim.** It is scoped to "as frozen" (untuned 0.70 thresholds, v1 questions, the 114-area catalogue), keeps HD6 `inconclusive` rather than `not-supported`, and the population limit is item 3.
- **It does not underclaim.** It leaves a new preregistration open, and every evidence-needed item maps to a measured figure once items 1 and 2 are corrected.
- The weak points were inside the evidence items:
  - the "content, not path" mechanism (R1-01);
  - the mixed abstention population and the single-threshold framing (R1-05);
  - the wrong area denominator (R1-06);
  - the P2 PASS being described as more robust than it is (R1-03).

## 4. Synthesis: where each reviewer changed the outcome


- **Lead.** Found R1-01, R1-03 (the P2 near-threshold fact and the missing round-17 explanations), R1-08 (the 16 ablated areas and the overlap fact), R1-11 (no reproduce command for §4.3) and R1-14 (the duplicate-pair filter, traced to `onboard/main.go`). Verified the four flagged points (V14, V19, V21, V22).
- **innovation-challenger.** Added R1-05 (the population mix and the two-threshold framing), R1-07 (the N4 `u` labelling, verified in `power.json`), the "≈ 10×" overstatement in R1-12, and R1-15. It independently found R1-01, R1-06 and the cost caveat. Its recommendation to keep the go/no-go line was adopted.
- **experiment-peer.** Added R1-02 (legacy statement), the 0/38 denominator in R1-01, R1-04 (raw path pull), the abstention-flip route in R1-03, R1-09 (labels-check, planted table, role mix), R1-10, R1-13 and ctl-9 in R1-08. This widened R1-03 from one route to failure to two, and turned R1-09 from a lead suspicion into a verified issue.
- **Rejected or not raised.**
  - The §1.4 citation of `PREREGISTRATION.md:112` for the GO bounds: :112 itself equates answered-item accuracy with `R_assign_res` on the planted population, so the citation is acceptable.
  - The near-duplicate link in §1.4 item 4: it is stated as a fact (both items abstained) under an `[INFERENCE]` heading, so it is acceptable.
- **Outside RESULTS.** `docs/hypothesis-ledger.md:38` repeats "baselines 0/35". Align it with R1-01 when RESULTS is fixed (ledger skill).

## 5. Verdict

**15 ISSUES**: R1-01 to R1-15. RESULTS is not accepted. Apply the fixes above, then run a round-2 review of the changed lines and of any figure the fixes add.
