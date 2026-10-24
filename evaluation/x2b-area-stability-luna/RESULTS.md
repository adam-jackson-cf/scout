# X2b — Is there a Jaccard match threshold for Scout's production mapper (codex `gpt-6-luna`, effort `xhigh`)? Results

Preregistration: `PREREGISTRATION.md` revision 4 (signed off with zero council issues, `reviews/round-4.md`; digests below). Hypothesis: **HD5b** (`docs/hypothesis-ledger.md`). Governed by `AGENTS.md` experiment rules.
Run commit: `f2baaa80b784fba98b561c5eccf135e1d54b3d5c` (worktree `/Users/adamjackson/Projects/inflight/scout-x2b`, detached; created after sign-off, so PREREGISTRATION §2.6's "if created after it, in `RESULTS.md`" branch applies). Code under test: `2a2a346b518cd68c479ac8c25a65d85f3c128c37`.
Data: collected 2026-10-24, codex/fastapi first attempt 00:41:25Z (before-probe 00:41:22Z) to the axum after-probe 01:12:20Z (14 of 14 maps `rc` 0, one attempt each; `logs/x2b-collect.log`), then `run.sh replay`, `run.sh pairs`, `run.sh label` (`logs/label.log`: `labelled 584 pairs`), `run.sh analysis` (`data/results.json`, `data/analysis_output.txt`). No commit was made in the run worktree before this document was written.
This study supersedes no earlier study and is superseded by none. It is a codex-only repeat of X2 (`../x2-area-stability/RESULTS.md`, R-041) on a different model and effort (cross-model confound, PREREGISTRATION §4). Every number below is copied from `data/results.json` or `data/analysis_output.txt` (complete text in Appendix A) or is a stated tabulation of those; nothing is recomputed from raw data. `[INFERENCE]` marks statements not read from the outputs.

## 1. Headline

**Verdict (codex `gpt-6-luna` `xhigh`, the only arm): `fallback`; no threshold on the 0.30–0.90 grid passes C1–C3** (all 13 thresholds `fail`; reason string: "no threshold meets the criteria: refresh only at onboarding and manually"). **Citable for a design change: yes** (probe-consistency CONSISTENT with one served model `gpt-6-luna` and effort `xhigh` in all four probes; served/flags check `ok`, reasons `[]`; unexpected arms `[]`). Gates ok: class gate (remap 4 and mutation 2 measured in both repositories; required 2 and 1), label gate (unlabelled share at 0.30 is 0.0 in both repositories), control gate ok. Pass set empty (contiguity flag `true` for an empty set); feasible thresholds `[]`; imprecise thresholds `[]`; survival margin at t\*: none (no t\*); gradient (Jaccard vs `different`): `r = None` (no `different` label exists, so the correlation is undefined).

**HD5b state per PREREGISTRATION §4b: `fallback` and citable → `not-supported`** for codex `gpt-6-luna` `xhigh` on this grid and guide (not `refuted`: existential claim, this grid, this guide, two repositories, one run each). HD5 itself (X2, `gpt-6.1-sol`, adopt at 0.30) is unchanged by X2b. The ledger is not edited by this document.

**§4a decision (PROPOSAL, for a design round, NOT adopted here):** "no threshold on the 0.30-0.90 grid meets C1-C3 for gpt-6-luna xhigh (nothing below 0.30 measured): the design's fallback applies (agent map at onboarding and manually); X2's codex adopt is not applied to a model Scout no longer uses". Concretely: the agent map runs at onboarding and manually only; refresh-on-merge is removed (§4a). `mapping.areas.match_jaccard` stays 0.5 in the design until a design round (`AGENTS.md` design rules) adopts or rejects this proposal. X2 cross-check (reported, never decisive): X2's codex arm `adopt` at t\* = 0.30, pass set {0.30}, model `gpt-6.1-sol`, citable; **X2b's own status at X2's 0.30 is `fail`**; `x2b_t_star_in_x2_pass_set`: none (there is no X2b t\*). X2's t\* = 0.30 does not carry over to the production mapper.

**Why it fails (the effect, per repository; exact means from `results.json`):**

| | axum | fastapi |
|---|---|---|
| C1 mean per-refresh survival at 0.30 (lowest grid point; exact) | **0.9102** (2503211/2750280; passes ≥ 4/5) | **0.5721** (4943/8640; **fails**) |
| worst transition at 0.30 | T3 0.8488 | T4 0.3750 |
| pooled survival at 0.30 (areas as independent; Wilson 95 %) | 0.910 [0.881, 0.932] | 0.544 [0.457, 0.629] |
| spurious (`different` / labelled) at 0.30; Wilson | 0/453 = 0.000 [0.000, 0.008] | 0/68 = 0.000 [0.000, 0.053] |
| strict (`overlapping` counted) at 0.30 | 0.130 (59/453) | 0.471 (32/68) |
| C1 mean survival at the design default 0.50 (exact) | **0.8800** (29042261/33003360; passes) | **0.4154** (3589/8640; fails) |
| worst transition at 0.50 | T3 0.8256 | T4 0.2083 |
| pooled survival at 0.50; Wilson | 0.880 [0.848, 0.905] | 0.400 [0.318, 0.488] |
| spurious at 0.50; Wilson | 0/438 = 0.000 [0.000, 0.009] | 0/50 = 0.000 [0.000, 0.071] |
| strict at 0.50 | 0.107 (47/438) | 0.380 (19/50) |
| per-threshold status at 0.50 | fail (joint, driven by fastapi) | fail |

- **axum passes C1 at every threshold t ≤ 0.50** (mean survival 0.9102 at 0.30, 0.8939 at 0.35 and 0.40, 0.8841 at 0.45, 0.8800 at 0.50) and fails from 0.55 up (0.5950 at 0.55); C2 and C3 pass at every threshold (zero `different` labels; largest Wilson upper bound 0.017). Across the drop 0.50 → 0.55 the matched count falls 438 → 294 and the Jaccard band [0.50, 0.55) holds 144 of axum's 521 labelled pairs (122 `same_concern`, 22 `overlapping_concern`; §7.3). [INFERENCE] Axum's 0.50-default pass sits on a cliff, not on a margin: a large group of axum matches have Jaccard in [0.50, 0.55), so the same mapper output survives at 0.50 and mostly does not at 0.55.
- **fastapi never reaches C1**: its mean survival is 0.5721 at 0.30 (worst transition 0.3750) and falls monotonically to 0.1296 (7/54) at 0.90. The joint (both repositories) criterion therefore fails at every threshold, including 0.30. [INFERENCE] The binding constraint is fastapi's map-to-map instability, not the threshold: luna's fastapi maps vary in feature count from 15 to 38 across the transitions' next maps (and prior-area counts 15–32; §5), with a prior area of zero overlap with every next feature in 1, 3, 4, 7, 0, 4 prior areas per transition. Even counting kept areas (area-ID retention, the `ret` values in §3.2), fastapi's mean at 0.30 is 0.8841.
- **Spurious merges: zero `different_concern` labels** among 521 labelled real pairs in every stratum, threshold and repository (asserted by `analysis_output.txt`: over-call check in the [0.90, 1.0] band 0 of 227 axum, 0 of 17 fastapi). C2/C3 do not bind. The strict rate (`overlapping` counted) is large where matches are loose (fastapi 0.471 at 0.30), reported as a sensitivity; the verdict stays the preregistered primary count. Labels are model-made and unreviewed by a human (R-039).
- **Leave-one-transition-out (joint t\*, descriptive)**: `t* = None` for every held-out transition T1–T6 (§3.3). No adopted t\* to leave out, and no subset of five transitions has a passing threshold.
- **Scope limits (PREREGISTRATION §4, §10):** one run of 7 maps per repository, one arm, no replication; the verdict is one draw of the agent per repository (the one-arm costs of §4 restated in §10 below); two repositories, no cross-repository claim; model and effort changed together relative to X2, so any difference from X2 is descriptive only; the served model is inferred from probes that echo their flags **[INFERENCE]**; nothing below 0.30 was measured, so the study does not say whether a lower threshold would pass; labels are model-made, labellers are the same family as the mapper (GPT-6) for B (§7.1).

## 2. Checklist: PREREGISTRATION §11 reporting contract → sections here

| §11 item | Section |
|---|---|
| Per stratum, repo and threshold tables | §3.1 (decisive, verbatim), §3.4 (strata at 0.30 and 0.50), Appendix A (every stratum, every threshold) |
| Per-transition survival and area-ID retention | §3.2 |
| Leave-one-transition-out | §3.3 |
| Gates, citability, served/flags check with reasons | §4.1, §4.2 |
| Combination block, `PROPOSAL`, X2's t\* status in X2b, X2 cross-check | §1, §4.3 |
| §4b outcome fields (0.50 and 0.30 figures, per-repo mean survival with pooled Wilson) | §1, §4.4 |
| Operating characteristics; X2b expected n | §4.5 |
| Coverage diagnostics | §5 |
| Chained survival at 0.50 | §6 |
| Labels: agreement, adjudication, served models, controls, integrity, labelled occurrences vs unique pairs, distribution | §7 |
| Map inventory, wall times | §8 |
| Unmeasured counts | §9 |
| Tool versions, probes with effort, build safety, CLI digest vs X2 | §10 |
| Run-commit provenance (conditions (4)–(8), §2.6) | §11 |
| Deviations | §12 |
| Reproduce | §13 |
| Threats; one-arm costs restated | §14 |
| Superseding / superseded links | header (none) |

Quantities required by §11 that the outputs do not contain: none identified. `survival_pooled` is printed beside every Wilson interval below, and labelled "pooled" (round-3 observation (b)): the Wilson interval is for pooled matched/prior-area survival, **not** for the per-repository mean.

## 3. Survival and spurious merge by threshold

### 3.1 Decisive stratum (all six transitions), verbatim rows of `analysis_output.txt`
Columns: `prev` prior areas; `match` matches; `surv(mean)` mean of per-transition survival; `worstT` worst transition; `pooled [Wilson]` pooled survival with Wilson 95 %; `crtd/retd/kept` created, retired, kept areas; `Jmed` upper-median Jaccard of matches; `J<.8` share of matches with Jaccard < 0.8; `lab(occ/uniq)` labelled occurrences / unique labelled pairs; `diff`, `overl` the `different` and `overlapping` counts; `unlab(ct/unp/never)` unlabelled; `spur [Wilson]` spurious rate; `strict` and `worst` the strict and worst-case spurious rates.

**axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   498   453  0.910  0.849  0.910 [0.881,0.932]    61    25    20 1.000 0.470 |   453/453      0    59     0 (0/0/0)  0.000 [0.000,0.008] 0.130 0.000
   0.35   498   445  0.894  0.849  0.894 [0.863,0.918]    69    28    25 1.000 0.461 |   445/445      0    54     0 (0/0/0)  0.000 [0.000,0.009] 0.121 0.000
   0.40   498   445  0.894  0.849  0.894 [0.863,0.918]    69    28    25 1.000 0.461 |   445/445      0    54     0 (0/0/0)  0.000 [0.000,0.009] 0.121 0.000
   0.45   498   440  0.884  0.826  0.884 [0.852,0.909]    74    32    26 1.000 0.455 |   440/440      0    49     0 (0/0/0)  0.000 [0.000,0.009] 0.111 0.000
   0.50   498   438  0.880  0.826  0.880 [0.848,0.905]    76    32    28 1.000 0.452 |   438/438      0    47     0 (0/0/0)  0.000 [0.000,0.009] 0.107 0.000
   0.55   498   294  0.595  0.341  0.590 [0.547,0.633]   220    89   115 1.000 0.184 |   294/294      0    25     0 (0/0/0)  0.000 [0.000,0.013] 0.085 0.000
   0.60   498   293  0.593  0.341  0.588 [0.545,0.631]   221    90   115 1.000 0.181 |   293/293      0    24     0 (0/0/0)  0.000 [0.000,0.013] 0.082 0.000
   0.65   498   282  0.571  0.329  0.566 [0.522,0.609]   232    99   117 1.000 0.149 |   282/282      0    17     0 (0/0/0)  0.000 [0.000,0.013] 0.060 0.000
   0.70   498   264  0.535  0.280  0.530 [0.486,0.574]   250   110   124 1.000 0.091 |   264/264      0    10     0 (0/0/0)  0.000 [0.000,0.014] 0.038 0.000
   0.75   498   263  0.533  0.268  0.528 [0.484,0.572]   251   111   124 1.000 0.087 |   263/263      0    10     0 (0/0/0)  0.000 [0.000,0.014] 0.038 0.000
   0.80   498   240  0.487  0.207  0.482 [0.438,0.526]   274   120   138 1.000 0.000 |   240/240      0     5     0 (0/0/0)  0.000 [0.000,0.016] 0.021 0.000
   0.85   498   228  0.463  0.171  0.458 [0.415,0.502]   286   128   142 1.000 0.000 |   228/228      0     0     0 (0/0/0)  0.000 [0.000,0.017] 0.000 0.000
   0.90   498   227  0.461  0.171  0.456 [0.413,0.500]   287   129   142 1.000 0.000 |   227/227      0     0     0 (0/0/0)  0.000 [0.000,0.017] 0.000 0.000
```

**fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   125    68  0.572  0.375  0.544 [0.457,0.629]    85    15    42 0.636 0.735 |    68/68       0    32     0 (0/0/0)  0.000 [0.000,0.053] 0.471 0.000
   0.35   125    61  0.516  0.250  0.488 [0.402,0.575]    92    18    46 0.667 0.705 |    61/61       0    27     0 (0/0/0)  0.000 [0.000,0.059] 0.443 0.000
   0.40   125    58  0.491  0.208  0.464 [0.379,0.551]    95    18    49 0.667 0.690 |    58/58       0    25     0 (0/0/0)  0.000 [0.000,0.062] 0.431 0.000
   0.45   125    54  0.456  0.208  0.432 [0.348,0.520]    99    19    52 0.667 0.667 |    54/54       0    22     0 (0/0/0)  0.000 [0.000,0.066] 0.407 0.000
   0.50   125    50  0.415  0.208  0.400 [0.318,0.488]   103    19    56 0.714 0.640 |    50/50       0    19     0 (0/0/0)  0.000 [0.000,0.071] 0.380 0.000
   0.55   125    43  0.355  0.111  0.344 [0.266,0.431]   110    22    60 0.750 0.581 |    43/43       0    12     0 (0/0/0)  0.000 [0.000,0.082] 0.279 0.000
   0.60   125    37  0.295  0.111  0.296 [0.223,0.381]   116    22    66 0.778 0.514 |    37/37       0     9     0 (0/0/0)  0.000 [0.000,0.094] 0.243 0.000
   0.65   125    32  0.255  0.111  0.256 [0.188,0.339]   121    23    70 1.000 0.438 |    32/32       0     5     0 (0/0/0)  0.000 [0.000,0.107] 0.156 0.000
   0.70   125    25  0.197  0.111  0.200 [0.139,0.279]   128    26    74 1.000 0.280 |    25/25       0     1     0 (0/0/0)  0.000 [0.000,0.133] 0.040 0.000
   0.75   125    23  0.182  0.056  0.184 [0.126,0.261]   130    26    76 1.000 0.217 |    23/23       0     1     0 (0/0/0)  0.000 [0.000,0.143] 0.043 0.000
   0.80   125    18  0.141  0.000  0.144 [0.093,0.216]   135    26    81 1.000 0.000 |    18/18       0     0     0 (0/0/0)  0.000 [0.000,0.176] 0.000 0.000
   0.85   125    18  0.141  0.000  0.144 [0.093,0.216]   135    26    81 1.000 0.000 |    18/18       0     0     0 (0/0/0)  0.000 [0.000,0.176] 0.000 0.000
   0.90   125    17  0.130  0.000  0.136 [0.087,0.207]   136    26    82 1.000 0.000 |    17/17       0     0     0 (0/0/0)  0.000 [0.000,0.184] 0.000 0.000
```

Exact mean survival per threshold (`results.json`, `survival_exact`), axum (C1 ≥ 0.8; fastapi's means are the `surv(mean)` column above and fail C1 at all 13 thresholds, 0.5721 at 0.30):

| t | mean survival (exact) | = | C1 (≥ 0.8) |
|---|---|---|---|
| 0.30 | 2503211/2750280 | 0.9102 | pass |
| 0.35 | 29502781/33003360 | 0.8939 | pass |
| 0.40 | 29502781/33003360 | 0.8939 | pass |
| 0.45 | 29179861/33003360 | 0.8841 | pass |
| 0.50 | 29042261/33003360 | 0.8800 | pass |
| 0.55 | 4909409/8250840 | 0.5950 | FAIL |
| 0.60 | 4893419/8250840 | 0.5931 | FAIL |
| 0.65 | 2095071/3667040 | 0.5713 | FAIL |
| 0.70 | 1104433/2062710 | 0.5354 | FAIL |
| 0.75 | 2200481/4125420 | 0.5334 | FAIL |
| 0.80 | 268081/550056 | 0.4874 | FAIL |
| 0.85 | 5097863/11001120 | 0.4634 | FAIL |
| 0.90 | 52864/114595 | 0.4613 | FAIL |

fastapi exact means: 0.30: 4943/8640, 0.35: 2231/4320, 0.40: 707/1440, 0.45: 3941/8640, 0.50: 3589/8640, 0.55: 383/1080, 0.60: 319/1080, 0.65: 2207/8640, 0.70: 63/320, 0.75: 197/1080, 0.80: 19/135, 0.85: 19/135, 0.90: 7/54

### 3.2 Per-transition survival and area-ID retention (decisive)
Cells: `matched/prior areas = survival (ret = kept-inclusive survival, (matched + kept)/prior areas, the design's "area-ID retention")`. Transitions: T1, T2 = F1→F2, F1→F3 (remap/fresh); T3, T4 = F1→R1, R1→R2 (remap/same_state); T5, T6 = R2→M1, M1→M2 (mutation).

**axum**

| threshold | T1 | T2 | T3 | T4 | T5 | T6 | mean (exact) | worst |
|---|---|---|---|---|---|---|---|---|
| 0.30 | 78/86 = 0.907 (ret 0.942) | 81/86 = 0.942 (ret 0.977) | 73/86 = 0.849 (ret 0.942) | 76/80 = 0.950 (ret 0.950) | 73/82 = 0.890 (ret 0.902) | 72/78 = 0.923 (ret 0.987) | 2503211/2750280 = 0.9102 | T3 0.8488 |
| 0.50 | 74/86 = 0.860 (ret 0.919) | 80/86 = 0.930 (ret 0.965) | 71/86 = 0.826 (ret 0.919) | 73/80 = 0.912 (ret 0.950) | 70/82 = 0.854 (ret 0.890) | 70/78 = 0.897 (ret 0.974) | 29042261/33003360 = 0.8800 | T3 0.8256 |
| 0.70 | 63/86 = 0.733 (ret 0.837) | 26/86 = 0.302 (ret 0.849) | 26/86 = 0.302 (ret 0.872) | 64/80 = 0.800 (ret 0.850) | 23/82 = 0.280 (ret 0.354) | 62/78 = 0.795 (ret 0.910) | 1104433/2062710 = 0.5354 | T5 0.2805 |
| 0.90 | 60/86 = 0.698 (ret 0.802) | 18/86 = 0.209 (ret 0.826) | 19/86 = 0.221 (ret 0.837) | 56/80 = 0.700 (ret 0.775) | 14/82 = 0.171 (ret 0.305) | 60/78 = 0.769 (ret 0.897) | 52864/114595 = 0.4613 | T5 0.1707 |

**fastapi**

| threshold | T1 | T2 | T3 | T4 | T5 | T6 | mean (exact) | worst |
|---|---|---|---|---|---|---|---|---|
| 0.30 | 13/18 = 0.722 (ret 0.889) | 11/18 = 0.611 (ret 0.889) | 7/18 = 0.389 (ret 1.000) | 9/24 = 0.375 (ret 0.625) | 13/15 = 0.867 (ret 0.933) | 15/32 = 0.469 (ret 0.969) | 4943/8640 = 0.5721 | T4 0.3750 |
| 0.50 | 9/18 = 0.500 (ret 0.833) | 10/18 = 0.556 (ret 0.889) | 4/18 = 0.222 (ret 0.944) | 5/24 = 0.208 (ret 0.583) | 9/15 = 0.600 (ret 0.933) | 13/32 = 0.406 (ret 0.938) | 3589/8640 = 0.4154 | T4 0.2083 |
| 0.70 | 2/18 = 0.111 (ret 0.833) | 2/18 = 0.111 (ret 0.833) | 2/18 = 0.111 (ret 0.833) | 4/24 = 0.167 (ret 0.583) | 6/15 = 0.400 (ret 0.933) | 9/32 = 0.281 (ret 0.812) | 63/320 = 0.1969 | T1 0.1111 |
| 0.90 | 0/18 = 0.000 (ret 0.833) | 1/18 = 0.056 (ret 0.833) | 1/18 = 0.056 (ret 0.833) | 2/24 = 0.083 (ret 0.583) | 5/15 = 0.333 (ret 0.933) | 8/32 = 0.250 (ret 0.812) | 7/54 = 0.1296 | T1 0.0000 |

### 3.3 Leave-one-transition-out (joint t\*, descriptive; not out-of-sample)
For each held-out transition the 13 thresholds were classified jointly on the other five transitions. Result for **every** held-out transition T1, T2, T3, T4, T5, T6: `t* = None` per repository; held-out survival, matched, labelled and `different` are `null` for both repositories (nothing to evaluate). (Lines 26–31 of `analysis_output.txt`.)

### 3.4 Strata (remap, mutation, fresh, same_state) at 0.30 and 0.50
All strata and thresholds are in Appendix A. `different` is 0 in every stratum at every threshold.

| repo | stratum | prior areas | matched | survival mean (exact) at 0.30 | at 0.50 | `different` | `overlapping` at 0.30 |
|---|---|---|---|---|---|---|---|
| axum | remap | 338 | 308 | 3137/3440 = 0.9119 | 12139/13760 = 0.8822 | 0 | 41 |
| axum | mutation | 160 | 145 | 1933/2132 = 0.9067 | 1400/1599 = 0.8755 | 0 | 18 |
| axum | fresh | 172 | 159 | 159/172 = 0.9244 | 77/86 = 0.8953 | 0 | 21 |
| axum | same_state | 166 | 149 | 1547/1720 = 0.8994 | 5979/6880 = 0.8690 | 0 | 20 |
| fastapi | remap | 78 | 40 | 151/288 = 0.5243 | 107/288 = 0.3715 | 0 | 21 |
| fastapi | mutation | 47 | 28 | 641/960 = 0.6677 | 161/320 = 0.5031 | 0 | 11 |
| fastapi | fresh | 36 | 24 | 2/3 = 0.6667 | 19/36 = 0.5278 | 0 | 13 |
| fastapi | same_state | 42 | 16 | 55/144 = 0.3819 | 31/144 = 0.2153 | 0 | 8 |

## 4. Decision details

### 4.1 Gates and citability (from `results.json`, `arms.codex.decision` and `x2b.served_check`)
- Class gate: axum remap 4 / mutation 2, fastapi remap 4 / mutation 2 measured (required 2 / 1): ok.
- Label gate: unlabelled share at T_MIN 0.0 (axum), 0.0 (fastapi); limit 1/10: ok.
- Control gate: ok, reason none (§7.2).
- Verdict `fallback`; `pass_set_contiguous` true; `tripped_gate` none; `feasible_pct` `[]`; `imprecise_pct` `[]`.
- Citable: `true`. Probe-consistency CONSISTENT (models `["gpt-6-luna"]`, CLI versions changed between probes: `{}`); served/flags check `ok` with reasons `[]` (every successful map's `meta.json` records `--model gpt-6-luna --reasoning-effort xhigh`, `agent-start` reported `provider=codex` and the option `gpt-6-luna`, rc 0); unexpected arms `[]`.
- Per-threshold status: `fail` at 0.30, 0.35, 0.40, 0.45, 0.50, 0.55, 0.60, 0.65, 0.70, 0.75, 0.80, 0.85, 0.90.

### 4.2 Why each threshold fails
C1 (mean survival ≥ 4/5 in each repository) fails on fastapi at all 13 thresholds (§3.1); it also fails on axum from 0.55 up. C2 (spurious ≤ 1/20) passes in both repositories at all 13 thresholds (0 `different`). C3 (Wilson upper ≤ 0.10) passes in axum at all 13 (largest upper bound 0.017 at 0.90) and in fastapi at 0.30–0.60 (upper bounds 0.053 … 0.094) but **fails in fastapi from 0.65 up** (0.107 at 0.65, 0.133 at 0.70, 0.143 at 0.75, 0.176 at 0.80 and 0.85, 0.184 at 0.90), purely because few matches remain labelled (n = 32 down to 17) while no `different` exists. Every threshold from 0.55 up is already C1-failed in both repositories, so C3 never decides a verdict: C1 alone does, and for t ≤ 0.50 the fastapi repository alone.

### 4.3 Combination with X2 (§4a; `results.json` `x2b.combination`)
X2b verdict `fallback`, citable true, `t*` None; X2 codex (frozen, digest-pinned) `{"available": true, "citable": true, "models": ["gpt-6.1-sol"], "pass_pct": [30], "threshold_pct": 30, "verdict": "adopt"}`; `x2b_t_star_in_x2_pass_set` None; `x2_t_star_status_in_x2b` `fail`. The preregistered branch is "X2b `fallback`, citable". PROPOSAL line (verbatim): "no threshold on the 0.30-0.90 grid meets C1-C3 for gpt-6-luna xhigh (nothing below 0.30 measured): the design's fallback applies (agent map at onboarding and manually); X2's codex adopt is not applied to a model Scout no longer uses". Not adopted here.

### 4.4 §4b outcome fields (fixed for `fallback`)
Per-threshold status at 0.50 (default in force): `fail`. Per-repository mean survival (exact), pooled Wilson and spurious with Wilson at **0.50** and at **0.30** (the lowest grid point): see the table in §1 (0.30 rows and 0.50 rows). Survival and spurious quantities with their intervals at 0.30: axum mean 0.9102, pooled 0.910 [0.881, 0.932], spurious 0/453 [0.000, 0.008], strict 0.130; fastapi mean 0.5721, pooled 0.544 [0.457, 0.629], spurious 0/68 [0.000, 0.053], strict 0.471. The inconclusive/not-citable lists of §4b do not apply (round-3 observation (e)): the verdict is `fallback`, citable, and neither list applies.

### 4.5 Expected n and operating characteristics (PREREGISTRATION §4(iv); `results.json` `x2b`)
- **Expected labelled n per repository** (`expected_labelled_n_per_repo`, 0.8 × 6 transitions × the smoke-map agent-feature counts **[INFERENCE]**): axum 365, fastapi 125. **Actual labelled occurrences at T_MIN** (the decisive row at 0.30): axum 453, fastapi 68 (521 total; unique pairs also 521, so occurrences = unique pairs; against the 226 minimum for the control gate: 521 ≥ 226). The expectation held up for axum (453 vs ≈ 365) and not for fastapi (68 vs ≈ 125): fastapi's prior-area total over the six transitions is 125 (18+18+18+24+15+32) and only 68 of them matched at 0.30, against the 0.8 matched share assumed. [INFERENCE] The shortfall comes from the lower matched share, not from fewer features than the smoke map. This is an expectation, not a preregistered rule; no decision rule depends on it (§12).
- **Operating characteristics** (probability of adopting at one passing threshold, conditional on C1, independent matches, anticonservative; P(both repos) at true spurious rate 3 % / 4 % / 5 % / 7 % / 10 %):

| n per repo | max spurious count | 0.03 | 0.04 | 0.05 | 0.07 | 0.10 |
|---|---|---|---|---|---|---|
| 100 | 4 | 0.6689 | 0.3955 | 0.1901 | 0.0266 | 0.0006 |
| 125 | 5 | 0.6817 | 0.3794 | 0.1612 | 0.0150 | 0.0001 |
| 150 | 7 | 0.8401 | 0.5578 | 0.2733 | 0.0285 | 0.0002 |
| 216 | 10 | 0.8784 | 0.5639 | 0.2318 | 0.0109 | 0.0 |
| 270 | 13 | 0.9314 | 0.6456 | 0.2671 | 0.0088 | 0.0 |

  `c3_stops_binding_at_n` = 140 is X2's multiple-of-20 search; at integer n C3 can bind only for n ≤ 126 (PREREGISTRATION §4(iv)). Line 20 of `analysis_output.txt` prints X2's "216-270 labelled occurrences per repo" band; line 203 states that this band is X2's, not X2b's (as preregistered). The actual fastapi n = 68 is below 126, so C3 binds there in principle; at 0/68 spurious its Wilson upper bound (0.053) passes C3 anyway.

## 5. Coverage diagnostics (descriptive; not used in any decision)
Per transition: prior areas, next agent features, distinct owned files in each map and in both, prior areas with zero overlap with every next agent feature (a survival ceiling no threshold can lift).

**axum**

| T | class/sub | prior areas | next agent features | prior owned files | next owned files | shared owned files | zero-overlap prior areas |
|---|---|---|---|---|---|---|---|
| T1 | remap/fresh | 86 | 89 | 224 | 235 | 216 | 1 |
| T2 | remap/fresh | 86 | 94 | 224 | 179 | 168 | 0 |
| T3 | remap/same_state | 86 | 80 | 224 | 174 | 161 | 3 |
| T4 | remap/same_state | 80 | 82 | 174 | 174 | 165 | 0 |
| T5 | mutation/mutation | 82 | 78 | 174 | 230 | 167 | 2 |
| T6 | mutation/mutation | 78 | 91 | 230 | 242 | 221 | 1 |

**fastapi**

| T | class/sub | prior areas | next agent features | prior owned files | next owned files | shared owned files | zero-overlap prior areas |
|---|---|---|---|---|---|---|---|
| T1 | remap/fresh | 18 | 17 | 243 | 247 | 162 | 1 |
| T2 | remap/fresh | 18 | 27 | 243 | 212 | 148 | 3 |
| T3 | remap/same_state | 18 | 24 | 243 | 217 | 100 | 4 |
| T4 | remap/same_state | 24 | 15 | 217 | 331 | 107 | 7 |
| T5 | mutation/mutation | 15 | 32 | 331 | 468 | 284 | 0 |
| T6 | mutation/mutation | 32 | 38 | 468 | 227 | 168 | 4 |

Agent feature counts of the next maps (T1–T6 → F2, F3, R1, R2, M1, M2): axum T1: 89, T2: 94, T3: 80, T4: 82, T5: 78, T6: 91; fastapi T1: 17, T2: 27, T3: 24, T4: 15, T5: 32, T6: 38; prior-area counts (F1, F1, F1, R1, R2, M1): axum 86, 86, 86, 80, 82, 78; fastapi 18, 18, 18, 24, 15, 32. **[INFERENCE]** The map labels follow the state layout of X2 §2 (F1 F2 F3 R1 R2 M1 M2); the seven fastapi maps then have 18 / 17 / 27 / 24 / 15 / 32 / 38 areas (F1 … M2), so luna's feature count for the same repository state (F1, F2, F3, and R1, R2) ranges 15–27 and across all seven maps 15–38. The smoke map (one per repository, before this study) had 26 fastapi and 76 axum features.

## 6. Chained survival at 0.50 (label-free; design default, nothing adopted; reported, not gating)
Fold F1→R1→R2→M1→M2 (T3–T6) with accumulating areas; `F1 areas matched at every step so far` is the share of original F1 areas matched at every step so far.

**axum**

| step | status | active prior | matched | created | kept | retired | F1 areas matched at every step so far / F1 areas |
|---|---|---|---|---|---|---|---|
| T3 | ok | 86 | 71 | 9 | 8 | 7 | 71/86 |
| T4 | ok | 88 | 79 | 3 | 5 | 4 | 69/86 |
| T5 | ok | 87 | 70 | 8 | 8 | 9 | 64/86 |
| T6 | ok | 86 | 73 | 18 | 11 | 2 | 60/86 |

**fastapi**

| step | status | active prior | matched | created | kept | retired | F1 areas matched at every step so far / F1 areas |
|---|---|---|---|---|---|---|---|
| T3 | ok | 18 | 4 | 20 | 13 | 1 | 4/18 |
| T4 | ok | 37 | 5 | 10 | 22 | 10 | 1/18 |
| T5 | ok | 37 | 17 | 15 | 19 | 1 | 1/18 |
| T6 | ok | 51 | 14 | 24 | 35 | 2 | 1/18 |

[INFERENCE] axum's chain keeps 60 of 86 original F1 areas matched at every step by T6 (60/86); fastapi's keeps 1 of 18 by T4 and stays at 1 of 18, while its active-area count grows from 18 to 51 (created 20, 10, 15, 24). At 0.50, fastapi's one-step survival is already 0.4154 (§1). The one-step and chained directions are not ordered a priori (X2 §10).

## 7. Labels

### 7.1 Agreement, adjudication and served models (real pairs, round 1)
- Unique real pairs labelled: **521** (one arm); controls **63**; call rows 584; `labelled 584 pairs` (`logs/label.log`).
- A/B raw agreement **0.9712092130518234**; Cohen's κ **0.8997369940342548** (521 pairs, both parsed 521, unparsable final 0, codex sessions 521, pairs with an exec event 0).
- Adjudicated pairs (C sessions): **15**, adjudication rate **0.028790786948176585** (15/521); C answered validly in 15 of 15; C's served model equals A's in 0 of 15.
- Served models per role (every call row incl. controls; final attempt's model per call row): **A `claude-opus-5-5` (584 call rows), B `gpt-6.1-sol` (584), C `claude-sonnet-5-5` (16)**. Labeller B was **not** served `gpt-6-luna`, so the same-model caveat of PREREGISTRATION §6 does not apply; the same-family caveat (labellers B and the mapper are GPT-6 models; X2 §10 shared blind spots) remains **[INFERENCE: family inferred from model names]**. Labellers read the mapper's own titles and summaries, so luna's writing style may shift labeller behaviour (descriptive).
- Label integrity: `labelled_pairs_without_call_row` 0.
- Controls agreement: A/B raw 0.9841269841269841, κ **0.9678899082568806** (63 pairs, adjudicated 1, C valid 1 of 1).
- No human checks these labels (R-039; label status `agent_reviewed`).

### 7.2 Known-answer controls (pooled over the one arm; gate ok)
521 unique real pairs, 63 control pairs, shortfall 0, `labelled_pairs_without_call_row` 0.
- **Swapped** (expected `different_concern`): n = 42 (wanted 42, shortfall 0); usable 42 (minimum 21); misses (not `different_concern`) **3**, rate **0.071** (gate limit 20 %); cannot_tell 0, unparsable 0, never labelled 0; labels {"different_concern": 39, "overlapping_concern": 3}. By source-pair Jaccard band (the misses' bands): 0.40: 0/1; 0.50: 0/18; 0.65: 0/2; 0.75: 0/1; **0.90: 3/20**.
- **Identical** (expected `same_concern`): n = 21 (wanted 21, shortfall 0); usable 21 (minimum 11); misses (not `same_concern`) **0**, rate 0.000; cannot_tell 0, unparsable 0, never labelled 0; labels {"same_concern": 21}.
- Per arm (descriptive): codex identical 0/21, swapped 3/42.
- The three swapped-control misses are all `overlapping_concern` and all in source-pair Jaccard band 0.90 (3 of 20). [INFERENCE] Where two source pairs share most files, labellers may call a swapped (different-concern) control `overlapping`, so `overlapping` is not always a strict synonym for a loose match. Controls do not test the 0.30–0.60 regime (X2 §13).

### 7.3 Label distribution by Jaccard band (real matches at T_MIN; counts of final labels)
No `different_concern` and no `cannot_tell` among real pairs. By class: axum mutation {"overlapping_concern": 18, "same_concern": 127}, remap {"overlapping_concern": 41, "same_concern": 267}; fastapi mutation {"overlapping_concern": 11, "same_concern": 17}, remap {"overlapping_concern": 21, "same_concern": 19}. Over-call check ([0.90, 1.0] band; `different` share of labelled matches): axum 0 of 227 (share 0.000), fastapi 0 of 17 (share 0.000).

| band lower edge (Jaccard) | axum same / overlapping | fastapi same / overlapping |
|---|---|---|
| 0.30 | 3 / 5 | 2 / 5 |
| 0.35 | 0 / 0 | 1 / 2 |
| 0.40 | 0 / 5 | 1 / 3 |
| 0.45 | 0 / 2 | 1 / 3 |
| 0.50 | 122 / 22 | 0 / 7 |
| 0.55 | 0 / 1 | 3 / 3 |
| 0.60 | 4 / 7 | 1 / 4 |
| 0.65 | 11 / 7 | 3 / 4 |
| 0.70 | 1 / 0 | 2 / 0 |
| 0.75 | 18 / 5 | 4 / 1 |
| 0.80 | 7 / 5 | 0 / 0 |
| 0.85 | 1 / 0 | 1 / 0 |
| 0.90 | 227 / 0 | 17 / 0 |

## 8. Map inventory, attempts and wall times
14 of 14 maps recorded `ok`; retried, failed, skipped, unrecorded or workspace-dirty maps: none; stop records (`data/stopped-*.json`): none; every map succeeded on its first attempt (`attempts=1`, `rc=0`).

| repo | map | first attempt started (UTC) | seconds total |
|---|---|---|---|
| fastapi | F1 | 2026-10-24T00:41:25Z | 95.6 |
| fastapi | F2 | 2026-10-24T00:43:02Z | 96.7 |
| fastapi | F3 | 2026-10-24T00:44:38Z | 140.6 |
| fastapi | R1 | 2026-10-24T00:46:59Z | 135.4 |
| fastapi | R2 | 2026-10-24T00:49:15Z | 121.1 |
| fastapi | M1 | 2026-10-24T00:51:17Z | 147.9 |
| fastapi | M2 | 2026-10-24T00:53:46Z | 89.4 |
| axum | F1 | 2026-10-24T00:55:25Z | 112.9 |
| axum | F2 | 2026-10-24T00:57:18Z | 139.3 |
| axum | F3 | 2026-10-24T00:59:38Z | 194.3 |
| axum | R1 | 2026-10-24T01:02:52Z | 136.3 |
| axum | R2 | 2026-10-24T01:05:09Z | 152.3 |
| axum | M1 | 2026-10-24T01:07:42Z | 141.8 |
| axum | M2 | 2026-10-24T01:10:05Z | 131.7 |

Sum of map wall times: fastapi 826.7 s, axum 1008.6 s.

## 9. Unmeasured counts and reasons
Transitions lost to failed, stopped or skipped maps: **0** (no `unmeasured_transitions`); maps retried or failed: 0; mutation skipped: 0; stops: 0; aborts: 0; unlabelled pairs: 0 (cannot_tell 0, unparsable 0, never labelled 0 at every row).

## 10. Tool versions, probes, provenance

### 10.1 Tool versions (`data/tool-versions-codex-<repo>.json`)
| | fastapi | axum |
|---|---|---|
| started (UTC) | 2026-10-24T00:41:25Z | 2026-10-24T00:55:25Z |
| codex CLI | codex-cli 0.162.1 | codex-cli 0.162.1 |
| Clawpatch (node) | 0.7.3 (v22.18.0) | 0.7.3 (v22.18.0) |
| `clawpatch_cli_sha256` | `4eab263b87d6b0d0af1eb6792e1e38827e0b13ab53078e62f5ffed76cc4320eb` | same |
| `clawpatch_build_safety` | ok=true, marker_ok=true, sanitizer_in_dist=true, marker `scout-review-sanitize-v3 dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3` | same |
| `scout_head` | `f2baaa80b784fba98b561c5eccf135e1d54b3d5c` | same |
| workspace config presence (all five names) | none present | none present |
| presence `CODEX_HOME` / `CLAWPATCH_CLAUDE_BIN` / `ANTHROPIC_MODEL` | true / false / false | true / false / false |
| provider timeouts codex/claude, per-call, wall, grace | 900000 / 900000 ms, 1800 s, 28800 s, 10 s | same |
| `x2b.map_flags` | `--model gpt-6-luna --reasoning-effort xhigh` | same |

`clawpatch_cli_sha256` equals X2's recorded value for the same field (`4eab263b…20eb`; round-3 observation (d); descriptive: the same built CLI as X2).

### 10.2 Model probes (arm codex: CONSISTENT; models `["gpt-6-luna"]`; CLI versions changed between probes: `{}`)
| repo | before | after |
|---|---|---|
| fastapi | 2026-10-24T00:41:22Z, `gpt-6-luna`, effort `xhigh`, 4.2 s, error null | 2026-10-24T00:55:20Z, `gpt-6-luna`, `xhigh`, 4.7 s, error null |
| axum | 2026-10-24T00:55:24Z, `gpt-6-luna`, `xhigh`, 3.2 s, error null | 2026-10-24T01:12:20Z, `gpt-6-luna`, `xhigh`, 3.5 s, error null |

CLI before and after in every case `codex-cli 0.162.1`. The probe passes `--model` and `-c model_reasoning_effort="xhigh"` itself, so its banner echoes the requested values; a same-name server-side roll of `gpt-6-luna` is invisible before and after (PREREGISTRATION §2.2, §10) **[INFERENCE]**.

### 10.3 Replay provenance (`data/replay-meta.json`)
Scout HEAD `f2baaa80b784fba98b561c5eccf135e1d54b3d5c`; `internal/areas/areas.go` sha256 `8ed2e70815fdcc605d60402fb603cfb905d58457a5869b3795d59507a93cb484`; `internal/clawpatch/records.go` sha256 `57344e025d5e2d0b7a10abf543b20b55f899602c1dd95d2b6a2b62d9b823dfbe`; working tree dirty for those two files: false; Go `go1.25.6`; grid `[30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90]` percent; transitions file sha256 `18d3d618ace78f00a8be9943597a666a09e4b9c393d1c7caada7b38e2e502603` (12 codex rows); maps directory `evaluation/x2b-area-stability-luna/data/maps`.

## 11. Run-commit provenance (PREREGISTRATION §2.6; `reviews/round-4.md` conditions (4)–(9))

**Run commit: `f2baaa80b784fba98b561c5eccf135e1d54b3d5c`** ("X2b: preregistration rev 4 and implementation signed off (council round 4, zero issues); HD5b preregistered"), a child of `42bf9cc` (hypothesis ledger: HD5b proposed), which is a descendant of `2a2a346`. It contains the 18 study files of `evaluation/x2b-area-stability-luna/` (git tree) and the sign-off round. The evidence below was taken by the main agent before the first collection (`logs/preconditions.txt`) and re-read from the worktree and the versions files for this report.

- **(4) Worktree and shell.** Dedicated detached worktree at the run commit; bootstrapped with `SCOUT_TOOLS_ACTIVE` unset (`preconditions.txt`: `SCOUT_TOOLS_ACTIVE=unset`; `logs/launch.sh` and `label.sh` both `unset SCOUT_TOOLS_ACTIVE` then `source .tools/activate`), so `.tools/activate` and `.tools/venv/bin/python` belong to the run worktree (own `.tools`).
- **(5) Diffs.** `git diff --quiet 2a2a346 f2baaa80… -- internal/ scripts/ third_party/clawpatch`: empty (`preconditions.txt` line 3; exit status 0 re-run for this report). `go.mod go.sum toolchain/` diff: empty (`preconditions.txt` line 4, the list is empty). `git submodule status third_party/clawpatch`: ` 734f81bc1514b52a0627edb0c24ebe1a18b36105 third_party/clawpatch (v0.7.3)` (no `+`/`-` prefix; same output re-run after the collections).
- **(6) Status.** `git status --porcelain internal/ scripts/ evaluation/x2b-area-stability-luna/` **before the first collection**: empty (`preconditions.txt`, `status_before_first_collection:` followed by `(end)`). **Before the second (axum) collection** (`logs/status-before-second.txt`, captured by `launch.sh` between the two collections): exactly `?? evaluation/x2b-area-stability-luna/data/`. No commit in the run worktree until both collections finished: `HEAD` is still `f2baaa80…` after the collections, and both `scout_head` values equal it. After the collections and analysis `git status --porcelain` (run for this report, whole tree) shows only `?? evaluation/x2b-area-stability-luna/data/` and `?? x2b-collect.log`.
- **(7) Frozen files.** `shasum -a 256` in the run worktree (`preconditions.txt`; recomputed after the run for this report with identical results) equals the sign-off digests:
```
7efb06334d4a7f6086e536286decd4443076e988d724443ff8e626a197ab207d  PREREGISTRATION.md
0b9f6de2f3250bfb389ff03368007bc38a31cf536f01e8064bbfce3b1d9bdf3a  README.md
2f63debd7363715d0f4d17019d3713de3a6ce8f01e31ed7aa46a28d3ab0f23ad  x2b.py
c1aa630764de2d17dabc2fc296b7f3e7542f84d5c9343da584459102ae369fec  collect_x2b.py
f1302f6cb95e59fba8a46586d5779be58956654d2937bd5a5fc1ecfeff3b0d09  analysis_x2b.py
50f09fd6bd2f2bc6ee19497ba71d04f0445dfeecd2afbed94999a90162bcfc92  run.sh
b1e26ffc3d3e505a60344bd94af975fe9a2f95ea5f9379b329fa3fd9d27c4fd5  validation/validate.py
3ac67b968c77995c23ba1965d28ea5b0037d170eda0a6a04039a167c730a135a  validation/stub_clawpatch_x2b.js
efef0f82b3d2493bbcbc8851f40f2e2db19c8627197e8eada0e7ca83d8615a24  validation/collect-x2b.diff
```
- **(8) After collection, per repository (both versions files, fastapi and axum, identical):** `scout_head` = `f2baaa80b784fba98b561c5eccf135e1d54b3d5c` = the run commit. Recorded Go digests equal the `2a2a346` blobs (`git show 2a2a346:<path> | shasum -a 256`, recomputed for this report):

| field | recorded value (both repos) | `2a2a346` blob |
|---|---|---|
| `x2b.client_go_sha256` (`internal/clawpatch/client.go`) | `ff05791c317eef52061755f2d9124c118c53c09004ec81ea6a2102d6c683fd88` | equal |
| `x2b.provider_go_sha256` (`internal/clawpatch/provider.go`) | `3b3a97346b87c676263ce80337d9bb911493a82c31f8662f3f1d12a022226e7f` | equal |
| `areas_go_sha256` (`internal/areas/areas.go`) | `8ed2e70815fdcc605d60402fb603cfb905d58457a5869b3795d59507a93cb484` | equal |
| `records_go_sha256` (`internal/clawpatch/records.go`) | `57344e025d5e2d0b7a10abf543b20b55f899602c1dd95d2b6a2b62d9b823dfbe` | equal |

  `x2b.collect_x2b_sha256` = `c1aa630764de2d17dabc2fc296b7f3e7542f84d5c9343da584459102ae369fec` = its sign-off digest. `x2b.x2_pins` (10 entries in each versions file) = `x2b.X2_PINS`, with the O1 chain stated rather than the tautology: the recorded `x2_pins` equal the `X2_PINS` literal of `x2b.py` at its sign-off digest `2f63debd…0f23ad` (item (7)), and the collector started, so `verify_pins()` passed and the X2 files on disk equalled those pins at run time (main agent's check). `clawpatch_cli_sha256` `4eab263b…20eb` and `clawpatch_build_safety` (ok) recorded; the CLI digest equals X2's recorded value (§10.1). The other two code-under-test blobs: `scripts/clawpatch-safe.patch` `dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3` and `scripts/build-clawpatch.sh` `99e3f6b3c670cd8c297b4ed498eafc735b7b5045aea8b97c7b55dc6343c21653` equal at the run commit (empty `git diff` of item (5); no recorded field).
- **(9) Ledger.** HD5b is `preregistered` with the sign-off History line in `docs/hypothesis-ledger.md` at the run commit (rows 30, 34, 35). `docs/scout-design.md:416,430` queueing for a design round is not verifiable from this document's inputs (not a RESULTS quantity).
- Conditions (1)–(3) (digest recomputation, blob check, validation) belong to the sign-off commit, not to this report.

## 12. Deviations

**None found.** (The ordering "replay, pairs, label, analysis, no commit until after both collections" is as reported by the main agent; the files above carry no timestamps for labelling or analysis.) Checks made: (a) all of conditions (4)–(8) hold with the evidence of §11; the run is citable under §2.6. (b) Collection: 14 of 14 maps first attempt, no stop, retry, abort, skip. (c) The preregistered order (codex/fastapi, then codex/axum, sequential), the arm, the flags and the probes with effort were as preregistered. (d) The labelling ran once (round 1) on 521 real pairs and 63 controls (584 call rows; the 16 `C` rows are adjudications: 15 real, 1 control); `replay` and `pairs` were not re-run after labels existed. (e) The §4(iii) control gate sizing (≥ 226 unique real pairs) is met by 521 unique pairs; all 21 identical controls were usable.
Expectations stated in the preregistration that were not met (none is a rule, none changes a verdict, none is a deviation):
- §4(iv) **expected labelled n** (≈ 125 fastapi, ≈ 365 axum, **[INFERENCE]**): actual 68 and 453 (§4.5). The preregistration flagged fastapi near 125 as a possible `inconclusive`-through-C3 case; it did not arise (68 is below 126, but 0/68 spurious has Wilson upper 0.053, and C1 failed anyway).
- `analysis_output.txt` line 20 prints X2's "expected n 216-270" band; `analysis_output.txt` line 203 corrects it for X2b, as the preregistration provides for.
- §4: "X2's codex arm had 436 unique labelled pairs ... the smoke maps suggest a similar or larger total" **[INFERENCE from one map each]**: actual total 521 labelled pairs, larger.
- §2.3: about 35 min of collection: actual 00:41:22Z–01:12:20Z, about 31 min.

## 13. Reproduce
From the run worktree (`git worktree add ../scout-x2b f2baaa80b784fba98b561c5eccf135e1d54b3d5c`), toolchain per `README.md` bootstrapped with `SCOUT_TOOLS_ACTIVE` unset (`git submodule update --init && scripts/bootstrap-tools.sh && source .tools/activate`):

    source .tools/activate && evaluation/x2b-area-stability-luna/run.sh validate   # no model calls

Collection as run (`logs/launch.sh`, copied from `/tmp/x2b-run/` after the run with the other files in `logs/`; `SOURCE_FASTAPI=/tmp/scout-p1/fastapi`, `SOURCE_AXUM=/tmp/scout-p4/axum-src`; `rm -rf /tmp/scout-x2b` first):

    PY=.tools/venv/bin/python; C=evaluation/x2b-area-stability-luna/collect_x2b.py
    nohup sh -c "$PY $C --arm codex --repo fastapi --source $SOURCE_FASTAPI &&
                 { git status --porcelain internal/ scripts/ evaluation/x2b-area-stability-luna/ > /tmp/x2b-run/status-before-second.txt; } &&
                 $PY $C --arm codex --repo axum    --source $SOURCE_AXUM" > x2b-collect.log 2>&1
    evaluation/x2b-area-stability-luna/run.sh replay
    evaluation/x2b-area-stability-luna/run.sh pairs
    evaluation/x2b-area-stability-luna/run.sh label       # via /tmp/x2b-run/label.sh (nohup), log /tmp/x2b-run/label.log
    evaluation/x2b-area-stability-luna/run.sh analysis    # data/results.json, data/analysis_output.txt

The `git status` line between the collections is the round-4 condition (6) capture; it is not part of `README.md`. Re-running the analysis on the existing `data/` (no model calls): `evaluation/x2b-area-stability-luna/run.sh analysis`. Model calls: codex (`gpt-6-luna` `xhigh` maps and probes; B labeller) and `claude -p` (A and C labellers); the keys are not used by the collector.

## 14. Threats, and the one-arm costs of §4 restated
- **One run per repository, one arm.** Run-to-run variance of the agent is inside the estimate; the verdict is one draw per repository (X2b adds no second draw to separate noise from effect). Here the draw is not near the edge for fastapi (mean 0.5721 vs 0.8 at the most favourable grid point), so the verdict does not look like noise at the edge; it does not exclude that a different draw would be higher **[INFERENCE]**. The luna fastapi maps vary from 15 to 38 features across the seven maps (§5): the run-to-run variance may itself be large **[INFERENCE]**.
- **No replication.** A design cannot claim this holds for another model; that is acceptable because Scout's map is codex `gpt-6-luna` only (R-040).
- **Model and effort changed together vs X2**; any difference from X2 (adopt at 0.30 there, fail here) cannot be attributed to either; descriptive only.
- **Same-name model roll** and the **served-model inference** (§10.2) are unchanged threats.
- **Two repositories**, no cluster-robust interval; shared prior maps (F1 in three transitions; R1, R2, M1 in two each) make the effective number of refreshes smaller than six.
- **Selection on the same labels** is not removed (X2 §7); moot here (no t\*).
- **Labels are model-made and unreviewed**; labellers see file overlap; the swapped controls missed 3 of 42 (all at band 0.90, §7.2). C2 passes vacuously (zero `different`); C3 fails in fastapi from 0.65 up from small n alone (§4.2); the strict rate is large at low thresholds, mainly in fastapi.
- **Knowledge of X2** (disclosure in the preregistration). Nothing below 0.30 measured.

## Appendix A. `data/analysis_output.txt` (verbatim, complete)

```
replay: scout f2baaa80b784 areas.go 8ed2e70815fd records.go 57344e025d5e go1.25.6 grid [30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90]
label agreement (real pairs, round 1): {"pairs": 521, "both_parsed": 521, "adjudicated": 15, "adjudication_rate": 0.028790786948176585, "adjudicated_c_ok": 15, "adjudicated_c_model_equals_a": 0, "adjudicated_c_model_equals_a_of": 15, "adjudicated_c_model_equals_a_share": 0.0, "codex_sessions": 521, "pairs_with_exec_event": 0, "unparsable_final": 0, "raw_agreement": 0.9712092130518234, "cohen_kappa": 0.8997369940342548}
label call rows by role, served model (final attempt's model per call row; every call row incl. controls and later rounds; a model seen only on a failed first attempt is not recorded): {"A": {"claude-opus-5-5": 584}, "B": {"gpt-6.1-sol": 584}, "C": {"claude-sonnet-5-5": 16}}
controls: pooled across arms; 521 unique real pairs; 63 pairs, gate ok
   swapped   n=42 (wanted 42, shortfall 0) usable=42 (min 21) misses(not different_concern)=3 rate=0.071 cannot_tell=0 unparsable=0 never=0 labels={"different_concern": 39, "overlapping_concern": 3}
   identical n=21 (wanted 21, shortfall 0) usable=21 (min 11) misses(not same_concern)=0 rate=0.000 cannot_tell=0 unparsable=0 never=0 labels={"same_concern": 21}
   per arm (descriptive): {"codex": {"identical": {"miss_rate": 0.0, "misses": 0, "n": 21, "usable": 21}, "swapped": {"miss_rate": 0.07142857142857142, "misses": 3, "n": 42, "usable": 42}}}
   swapped controls by source-pair Jaccard band: {"40": {"misses": 0, "n": 1, "usable": 1}, "50": {"misses": 0, "n": 18, "usable": 18}, "65": {"misses": 0, "n": 2, "usable": 2}, "75": {"misses": 0, "n": 1, "usable": 1}, "90": {"misses": 3, "n": 20, "usable": 20}}
   label integrity: {"labelled_pairs_without_call_row": 0}
   control agreement: {"pairs": 63, "both_parsed": 63, "adjudicated": 1, "adjudication_rate": 0.015873015873015872, "adjudicated_c_ok": 1, "adjudicated_c_model_equals_a": 0, "adjudicated_c_model_equals_a_of": 1, "adjudicated_c_model_equals_a_share": 0.0, "codex_sessions": 63, "pairs_with_exec_event": 0, "unparsable_final": 0, "raw_agreement": 0.9841269841269841, "cohen_kappa": 0.9678899082568806}
map inventory: 14 recorded ok of 14; retried/failed/skipped/unrecorded/workspace dirty after the map: []
probe consistency (section 2, arm codex): CONSISTENT; models ["gpt-6-luna"]; CLI versions changed between the probes: {}
   axum probes: before {"error": null, "model": "gpt-6-luna", "reasoning_effort": "xhigh", "seconds": 3.2, "timestamp": "2026-10-24T00:55:24Z"}, after {"error": null, "model": "gpt-6-luna", "reasoning_effort": "xhigh", "seconds": 3.5, "timestamp": "2026-10-24T01:12:20Z"}; CLI versions before {"codex": "codex-cli 0.162.1"}, after {"codex": "codex-cli 0.162.1"}
   axum workspace config presence (names only): {".claude/settings.json": false, ".claude/settings.local.json": false, ".clawpatch/config.json": false, ".codex": false, "clawpatch.config.json": false}
   fastapi probes: before {"error": null, "model": "gpt-6-luna", "reasoning_effort": "xhigh", "seconds": 4.2, "timestamp": "2026-10-24T00:41:22Z"}, after {"error": null, "model": "gpt-6-luna", "reasoning_effort": "xhigh", "seconds": 4.7, "timestamp": "2026-10-24T00:55:20Z"}; CLI versions before {"codex": "codex-cli 0.162.1"}, after {"codex": "codex-cli 0.162.1"}
   fastapi workspace config presence (names only): {".claude/settings.json": false, ".claude/settings.local.json": false, ".clawpatch/config.json": false, ".codex": false, "clawpatch.config.json": false}
operating characteristics (n=150 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 7 spurious; both repos): {"0.03": 0.8401, "0.04": 0.5578, "0.05": 0.2733, "0.07": 0.0285, "0.1": 0.0002}
operating characteristics (n=216 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 10 spurious; both repos): {"0.03": 0.8784, "0.04": 0.5639, "0.05": 0.2318, "0.07": 0.0109, "0.1": 0.0}
operating characteristics (n=270 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 13 spurious; both repos): {"0.03": 0.9314, "0.04": 0.6456, "0.05": 0.2671, "0.07": 0.0088, "0.1": 0.0}
   expected n 216-270 labelled occurrences per repo; C3 stops binding at n=140
== arm codex: verdict fallback, threshold None (no threshold meets the criteria: refresh only at onboarding and manually) [citable for a design change: yes]
   citable for a design change (X2b section 4): yes -- probe-consistency ok; served/flags ok; unexpected arms []  (X2's own bracket on the arm line above says 'arm probe-inconsistent' for any false value; this line gives the reason)
   per-threshold status: {"30": "fail", "35": "fail", "40": "fail", "45": "fail", "50": "fail", "55": "fail", "60": "fail", "65": "fail", "70": "fail", "75": "fail", "80": "fail", "85": "fail", "90": "fail"}
   gradient (Jaccard vs different, exploratory): {'n': 521, 'r': None}
   survival margin at t* (min over repos of mean survival - 4/5): None (None)
   leave-one-transition-out, joint t* (descriptive) hold out T1: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T2: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T3: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T4: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T5: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T6: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
-- axum coverage (T_MIN rows): {"T1": {"class": "remap", "sub": "fresh", "prevAreas": 86, "nextAgentFeatures": 89, "prevOwnedFiles": 224, "nextOwnedFiles": 235, "sharedOwnedFiles": 216, "zeroOverlapPriorAreas": 1}, "T2": {"class": "remap", "sub": "fresh", "prevAreas": 86, "nextAgentFeatures": 94, "prevOwnedFiles": 224, "nextOwnedFiles": 179, "sharedOwnedFiles": 168, "zeroOverlapPriorAreas": 0}, "T3": {"class": "remap", "sub": "same_state", "prevAreas": 86, "nextAgentFeatures": 80, "prevOwnedFiles": 224, "nextOwnedFiles": 174, "sharedOwnedFiles": 161, "zeroOverlapPriorAreas": 3}, "T4": {"class": "remap", "sub": "same_state", "prevAreas": 80, "nextAgentFeatures": 82, "prevOwnedFiles": 174, "nextOwnedFiles": 174, "sharedOwnedFiles": 165, "zeroOverlapPriorAreas": 0}, "T5": {"class": "mutation", "sub": "mutation", "prevAreas": 82, "nextAgentFeatures": 78, "prevOwnedFiles": 174, "nextOwnedFiles": 230, "sharedOwnedFiles": 167, "zeroOverlapPriorAreas": 2}, "T6": {"class": "mutation", "sub": "mutation", "prevAreas": 78, "nextAgentFeatures": 91, "prevOwnedFiles": 230, "nextOwnedFiles": 242, "sharedOwnedFiles": 221, "zeroOverlapPriorAreas": 1}}
-- axum label distribution: {"by_band_pct": {"30": {"overlapping_concern": 5, "same_concern": 3}, "40": {"overlapping_concern": 5}, "45": {"overlapping_concern": 2}, "50": {"overlapping_concern": 22, "same_concern": 122}, "55": {"overlapping_concern": 1}, "60": {"overlapping_concern": 7, "same_concern": 4}, "65": {"overlapping_concern": 7, "same_concern": 11}, "70": {"same_concern": 1}, "75": {"overlapping_concern": 5, "same_concern": 18}, "80": {"overlapping_concern": 5, "same_concern": 7}, "85": {"same_concern": 1}, "90": {"same_concern": 227}}, "by_class": {"mutation": {"overlapping_concern": 18, "same_concern": 127}, "remap": {"overlapping_concern": 41, "same_concern": 267}}, "over_call_check_band_90": {"different": 0, "labelled": 227, "matches": 227, "share": 0.0}}
-- axum over-call check (`different` share of labelled matches in the [0.90, 1.0] band): 0 of 227 labelled (227 matches): share 0.000
-- axum chained survival at 50% (design default; no adopted threshold): [{"step": "T3", "status": "ok", "activePrior": 86, "matched": 71, "created": 9, "kept": 8, "retired": 7, "f1MatchedAllSoFar": 71, "f1Areas": 86}, {"step": "T4", "status": "ok", "activePrior": 88, "matched": 79, "created": 3, "kept": 5, "retired": 4, "f1MatchedAllSoFar": 69, "f1Areas": 86}, {"step": "T5", "status": "ok", "activePrior": 87, "matched": 70, "created": 8, "kept": 8, "retired": 9, "f1MatchedAllSoFar": 64, "f1Areas": 86}, {"step": "T6", "status": "ok", "activePrior": 86, "matched": 73, "created": 18, "kept": 11, "retired": 2, "f1MatchedAllSoFar": 60, "f1Areas": 86}]
-- fastapi coverage (T_MIN rows): {"T1": {"class": "remap", "sub": "fresh", "prevAreas": 18, "nextAgentFeatures": 17, "prevOwnedFiles": 243, "nextOwnedFiles": 247, "sharedOwnedFiles": 162, "zeroOverlapPriorAreas": 1}, "T2": {"class": "remap", "sub": "fresh", "prevAreas": 18, "nextAgentFeatures": 27, "prevOwnedFiles": 243, "nextOwnedFiles": 212, "sharedOwnedFiles": 148, "zeroOverlapPriorAreas": 3}, "T3": {"class": "remap", "sub": "same_state", "prevAreas": 18, "nextAgentFeatures": 24, "prevOwnedFiles": 243, "nextOwnedFiles": 217, "sharedOwnedFiles": 100, "zeroOverlapPriorAreas": 4}, "T4": {"class": "remap", "sub": "same_state", "prevAreas": 24, "nextAgentFeatures": 15, "prevOwnedFiles": 217, "nextOwnedFiles": 331, "sharedOwnedFiles": 107, "zeroOverlapPriorAreas": 7}, "T5": {"class": "mutation", "sub": "mutation", "prevAreas": 15, "nextAgentFeatures": 32, "prevOwnedFiles": 331, "nextOwnedFiles": 468, "sharedOwnedFiles": 284, "zeroOverlapPriorAreas": 0}, "T6": {"class": "mutation", "sub": "mutation", "prevAreas": 32, "nextAgentFeatures": 38, "prevOwnedFiles": 468, "nextOwnedFiles": 227, "sharedOwnedFiles": 168, "zeroOverlapPriorAreas": 4}}
-- fastapi label distribution: {"by_band_pct": {"30": {"overlapping_concern": 5, "same_concern": 2}, "35": {"overlapping_concern": 2, "same_concern": 1}, "40": {"overlapping_concern": 3, "same_concern": 1}, "45": {"overlapping_concern": 3, "same_concern": 1}, "50": {"overlapping_concern": 7}, "55": {"overlapping_concern": 3, "same_concern": 3}, "60": {"overlapping_concern": 4, "same_concern": 1}, "65": {"overlapping_concern": 4, "same_concern": 3}, "70": {"same_concern": 2}, "75": {"overlapping_concern": 1, "same_concern": 4}, "85": {"same_concern": 1}, "90": {"same_concern": 17}}, "by_class": {"mutation": {"overlapping_concern": 11, "same_concern": 17}, "remap": {"overlapping_concern": 21, "same_concern": 19}}, "over_call_check_band_90": {"different": 0, "labelled": 17, "matches": 17, "share": 0.0}}
-- fastapi over-call check (`different` share of labelled matches in the [0.90, 1.0] band): 0 of 17 labelled (17 matches): share 0.000
-- fastapi chained survival at 50% (design default; no adopted threshold): [{"step": "T3", "status": "ok", "activePrior": 18, "matched": 4, "created": 20, "kept": 13, "retired": 1, "f1MatchedAllSoFar": 4, "f1Areas": 18}, {"step": "T4", "status": "ok", "activePrior": 37, "matched": 5, "created": 10, "kept": 22, "retired": 10, "f1MatchedAllSoFar": 1, "f1Areas": 18}, {"step": "T5", "status": "ok", "activePrior": 37, "matched": 17, "created": 15, "kept": 19, "retired": 1, "f1MatchedAllSoFar": 1, "f1Areas": 18}, {"step": "T6", "status": "ok", "activePrior": 51, "matched": 14, "created": 24, "kept": 35, "retired": 2, "f1MatchedAllSoFar": 1, "f1Areas": 18}]
-- axum / decisive
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   498   453  0.910  0.849  0.910 [0.881,0.932]    61    25    20 1.000 0.470 |   453/453      0    59     0 (0/0/0)  0.000 [0.000,0.008] 0.130 0.000
   0.35   498   445  0.894  0.849  0.894 [0.863,0.918]    69    28    25 1.000 0.461 |   445/445      0    54     0 (0/0/0)  0.000 [0.000,0.009] 0.121 0.000
   0.40   498   445  0.894  0.849  0.894 [0.863,0.918]    69    28    25 1.000 0.461 |   445/445      0    54     0 (0/0/0)  0.000 [0.000,0.009] 0.121 0.000
   0.45   498   440  0.884  0.826  0.884 [0.852,0.909]    74    32    26 1.000 0.455 |   440/440      0    49     0 (0/0/0)  0.000 [0.000,0.009] 0.111 0.000
   0.50   498   438  0.880  0.826  0.880 [0.848,0.905]    76    32    28 1.000 0.452 |   438/438      0    47     0 (0/0/0)  0.000 [0.000,0.009] 0.107 0.000
   0.55   498   294  0.595  0.341  0.590 [0.547,0.633]   220    89   115 1.000 0.184 |   294/294      0    25     0 (0/0/0)  0.000 [0.000,0.013] 0.085 0.000
   0.60   498   293  0.593  0.341  0.588 [0.545,0.631]   221    90   115 1.000 0.181 |   293/293      0    24     0 (0/0/0)  0.000 [0.000,0.013] 0.082 0.000
   0.65   498   282  0.571  0.329  0.566 [0.522,0.609]   232    99   117 1.000 0.149 |   282/282      0    17     0 (0/0/0)  0.000 [0.000,0.013] 0.060 0.000
   0.70   498   264  0.535  0.280  0.530 [0.486,0.574]   250   110   124 1.000 0.091 |   264/264      0    10     0 (0/0/0)  0.000 [0.000,0.014] 0.038 0.000
   0.75   498   263  0.533  0.268  0.528 [0.484,0.572]   251   111   124 1.000 0.087 |   263/263      0    10     0 (0/0/0)  0.000 [0.000,0.014] 0.038 0.000
   0.80   498   240  0.487  0.207  0.482 [0.438,0.526]   274   120   138 1.000 0.000 |   240/240      0     5     0 (0/0/0)  0.000 [0.000,0.016] 0.021 0.000
   0.85   498   228  0.463  0.171  0.458 [0.415,0.502]   286   128   142 1.000 0.000 |   228/228      0     0     0 (0/0/0)  0.000 [0.000,0.017] 0.000 0.000
   0.90   498   227  0.461  0.171  0.456 [0.413,0.500]   287   129   142 1.000 0.000 |   227/227      0     0     0 (0/0/0)  0.000 [0.000,0.017] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "78/86", "0.942"], ["T2", "81/86", "0.977"], ["T3", "73/86", "0.942"], ["T4", "76/80", "0.950"], ["T5", "73/82", "0.902"], ["T6", "72/78", "0.987"]], "50": [["T1", "74/86", "0.919"], ["T2", "80/86", "0.965"], ["T3", "71/86", "0.919"], ["T4", "73/80", "0.950"], ["T5", "70/82", "0.890"], ["T6", "70/78", "0.974"]], "70": [["T1", "63/86", "0.837"], ["T2", "26/86", "0.849"], ["T3", "26/86", "0.872"], ["T4", "64/80", "0.850"], ["T5", "23/82", "0.354"], ["T6", "62/78", "0.910"]], "90": [["T1", "60/86", "0.802"], ["T2", "18/86", "0.826"], ["T3", "19/86", "0.837"], ["T4", "56/80", "0.775"], ["T5", "14/82", "0.305"], ["T6", "60/78", "0.897"]]}
-- axum / remap
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   338   308  0.912  0.849  0.911 [0.876,0.937]    37    16    14 0.889 0.471 |   308/308      0    41     0 (0/0/0)  0.000 [0.000,0.012] 0.133 0.000
   0.35   338   302  0.894  0.849  0.893 [0.856,0.922]    43    18    18 1.000 0.460 |   302/302      0    38     0 (0/0/0)  0.000 [0.000,0.013] 0.126 0.000
   0.40   338   302  0.894  0.849  0.893 [0.856,0.922]    43    18    18 1.000 0.460 |   302/302      0    38     0 (0/0/0)  0.000 [0.000,0.013] 0.126 0.000
   0.45   338   298  0.882  0.826  0.882 [0.843,0.912]    47    21    19 1.000 0.453 |   298/298      0    34     0 (0/0/0)  0.000 [0.000,0.013] 0.114 0.000
   0.50   338   298  0.882  0.826  0.882 [0.843,0.912]    47    21    19 1.000 0.453 |   298/298      0    34     0 (0/0/0)  0.000 [0.000,0.013] 0.114 0.000
   0.55   338   202  0.602  0.372  0.598 [0.545,0.649]   143    33   103 1.000 0.193 |   202/202      0    18     0 (0/0/0)  0.000 [0.000,0.019] 0.089 0.000
   0.60   338   201  0.599  0.360  0.595 [0.542,0.646]   144    34   103 1.000 0.189 |   201/201      0    17     0 (0/0/0)  0.000 [0.000,0.019] 0.085 0.000
   0.65   338   192  0.573  0.337  0.568 [0.515,0.620]   153    42   104 1.000 0.151 |   192/192      0    11     0 (0/0/0)  0.000 [0.000,0.020] 0.057 0.000
   0.70   338   179  0.534  0.302  0.530 [0.476,0.582]   166    50   109 1.000 0.089 |   179/179      0     7     0 (0/0/0)  0.000 [0.000,0.021] 0.039 0.000
   0.75   338   179  0.534  0.302  0.530 [0.476,0.582]   166    50   109 1.000 0.089 |   179/179      0     7     0 (0/0/0)  0.000 [0.000,0.021] 0.039 0.000
   0.80   338   163  0.487  0.244  0.482 [0.429,0.535]   182    57   118 1.000 0.000 |   163/163      0     4     0 (0/0/0)  0.000 [0.000,0.023] 0.025 0.000
   0.85   338   154  0.460  0.209  0.456 [0.403,0.509]   191    63   121 1.000 0.000 |   154/154      0     0     0 (0/0/0)  0.000 [0.000,0.024] 0.000 0.000
   0.90   338   153  0.457  0.209  0.453 [0.400,0.506]   192    64   121 1.000 0.000 |   153/153      0     0     0 (0/0/0)  0.000 [0.000,0.024] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "78/86", "0.942"], ["T2", "81/86", "0.977"], ["T3", "73/86", "0.942"], ["T4", "76/80", "0.950"]], "50": [["T1", "74/86", "0.919"], ["T2", "80/86", "0.965"], ["T3", "71/86", "0.919"], ["T4", "73/80", "0.950"]], "70": [["T1", "63/86", "0.837"], ["T2", "26/86", "0.849"], ["T3", "26/86", "0.872"], ["T4", "64/80", "0.850"]], "90": [["T1", "60/86", "0.802"], ["T2", "18/86", "0.826"], ["T3", "19/86", "0.837"], ["T4", "56/80", "0.775"]]}
-- axum / mutation
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   160   145  0.907  0.890  0.906 [0.851,0.942]    24     9     6 1.000 0.469 |   145/145      0    18     0 (0/0/0)  0.000 [0.000,0.026] 0.124 0.000
   0.35   160   143  0.894  0.878  0.894 [0.836,0.933]    26    10     7 1.000 0.462 |   143/143      0    16     0 (0/0/0)  0.000 [0.000,0.026] 0.112 0.000
   0.40   160   143  0.894  0.878  0.894 [0.836,0.933]    26    10     7 1.000 0.462 |   143/143      0    16     0 (0/0/0)  0.000 [0.000,0.026] 0.112 0.000
   0.45   160   142  0.888  0.866  0.887 [0.829,0.928]    27    11     7 1.000 0.458 |   142/142      0    15     0 (0/0/0)  0.000 [0.000,0.026] 0.106 0.000
   0.50   160   140  0.876  0.854  0.875 [0.815,0.918]    29    11     9 1.000 0.450 |   140/140      0    13     0 (0/0/0)  0.000 [0.000,0.027] 0.093 0.000
   0.55   160    92  0.581  0.341  0.575 [0.498,0.649]    77    56    12 1.000 0.163 |    92/92       0     7     0 (0/0/0)  0.000 [0.000,0.040] 0.076 0.000
   0.60   160    92  0.581  0.341  0.575 [0.498,0.649]    77    56    12 1.000 0.163 |    92/92       0     7     0 (0/0/0)  0.000 [0.000,0.040] 0.076 0.000
   0.65   160    90  0.568  0.329  0.562 [0.485,0.637]    79    57    13 1.000 0.144 |    90/90       0     6     0 (0/0/0)  0.000 [0.000,0.041] 0.067 0.000
   0.70   160    85  0.538  0.280  0.531 [0.454,0.607]    84    60    15 1.000 0.094 |    85/85       0     3     0 (0/0/0)  0.000 [0.000,0.043] 0.035 0.000
   0.75   160    84  0.532  0.268  0.525 [0.448,0.601]    85    61    15 1.000 0.083 |    84/84       0     3     0 (0/0/0)  0.000 [0.000,0.044] 0.036 0.000
   0.80   160    77  0.488  0.207  0.481 [0.405,0.558]    92    63    20 1.000 0.000 |    77/77       0     1     0 (0/0/0)  0.000 [0.000,0.048] 0.013 0.000
   0.85   160    74  0.470  0.171  0.463 [0.387,0.540]    95    65    21 1.000 0.000 |    74/74       0     0     0 (0/0/0)  0.000 [0.000,0.049] 0.000 0.000
   0.90   160    74  0.470  0.171  0.463 [0.387,0.540]    95    65    21 1.000 0.000 |    74/74       0     0     0 (0/0/0)  0.000 [0.000,0.049] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T5", "73/82", "0.902"], ["T6", "72/78", "0.987"]], "50": [["T5", "70/82", "0.890"], ["T6", "70/78", "0.974"]], "70": [["T5", "23/82", "0.354"], ["T6", "62/78", "0.910"]], "90": [["T5", "14/82", "0.305"], ["T6", "60/78", "0.897"]]}
-- axum / fresh
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   172   159  0.924  0.907  0.924 [0.875,0.955]    24     7     6 0.800 0.484 |   159/159      0    21     0 (0/0/0)  0.000 [0.000,0.024] 0.132 0.000
   0.35   172   156  0.907  0.884  0.907 [0.854,0.942]    27     9     7 1.000 0.474 |   156/156      0    18     0 (0/0/0)  0.000 [0.000,0.024] 0.115 0.000
   0.40   172   156  0.907  0.884  0.907 [0.854,0.942]    27     9     7 1.000 0.474 |   156/156      0    18     0 (0/0/0)  0.000 [0.000,0.024] 0.115 0.000
   0.45   172   154  0.895  0.860  0.895 [0.841,0.933]    29    10     8 1.000 0.468 |   154/154      0    16     0 (0/0/0)  0.000 [0.000,0.024] 0.104 0.000
   0.50   172   154  0.895  0.860  0.895 [0.841,0.933]    29    10     8 1.000 0.468 |   154/154      0    16     0 (0/0/0)  0.000 [0.000,0.024] 0.104 0.000
   0.55   172   102  0.593  0.395  0.593 [0.518,0.664]    81    16    54 1.000 0.196 |   102/102      0     8     0 (0/0/0)  0.000 [0.000,0.036] 0.078 0.000
   0.60   172   102  0.593  0.395  0.593 [0.518,0.664]    81    16    54 1.000 0.196 |   102/102      0     8     0 (0/0/0)  0.000 [0.000,0.036] 0.078 0.000
   0.65   172    96  0.558  0.360  0.558 [0.483,0.630]    87    22    54 1.000 0.146 |    96/96       0     4     0 (0/0/0)  0.000 [0.000,0.038] 0.042 0.000
   0.70   172    89  0.517  0.302  0.517 [0.443,0.591]    94    27    56 1.000 0.079 |    89/89       0     2     0 (0/0/0)  0.000 [0.000,0.041] 0.022 0.000
   0.75   172    89  0.517  0.302  0.517 [0.443,0.591]    94    27    56 1.000 0.079 |    89/89       0     2     0 (0/0/0)  0.000 [0.000,0.041] 0.022 0.000
   0.80   172    82  0.477  0.256  0.477 [0.403,0.551]   101    30    60 1.000 0.000 |    82/82       0     1     0 (0/0/0)  0.000 [0.000,0.045] 0.012 0.000
   0.85   172    78  0.453  0.209  0.453 [0.381,0.528]   105    32    62 1.000 0.000 |    78/78       0     0     0 (0/0/0)  0.000 [0.000,0.047] 0.000 0.000
   0.90   172    78  0.453  0.209  0.453 [0.381,0.528]   105    32    62 1.000 0.000 |    78/78       0     0     0 (0/0/0)  0.000 [0.000,0.047] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "78/86", "0.942"], ["T2", "81/86", "0.977"]], "50": [["T1", "74/86", "0.919"], ["T2", "80/86", "0.965"]], "70": [["T1", "63/86", "0.837"], ["T2", "26/86", "0.849"]], "90": [["T1", "60/86", "0.802"], ["T2", "18/86", "0.826"]]}
-- axum / same_state
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   166   149  0.899  0.849  0.898 [0.842,0.935]    13     9     8 1.000 0.456 |   149/149      0    20     0 (0/0/0)  0.000 [0.000,0.025] 0.134 0.000
   0.35   166   146  0.881  0.849  0.880 [0.821,0.921]    16     9    11 1.000 0.445 |   146/146      0    20     0 (0/0/0)  0.000 [0.000,0.026] 0.137 0.000
   0.40   166   146  0.881  0.849  0.880 [0.821,0.921]    16     9    11 1.000 0.445 |   146/146      0    20     0 (0/0/0)  0.000 [0.000,0.026] 0.137 0.000
   0.45   166   144  0.869  0.826  0.867 [0.807,0.911]    18    11    11 1.000 0.438 |   144/144      0    18     0 (0/0/0)  0.000 [0.000,0.026] 0.125 0.000
   0.50   166   144  0.869  0.826  0.867 [0.807,0.911]    18    11    11 1.000 0.438 |   144/144      0    18     0 (0/0/0)  0.000 [0.000,0.026] 0.125 0.000
   0.55   166   100  0.611  0.372  0.602 [0.526,0.674]    62    17    49 1.000 0.190 |   100/100      0    10     0 (0/0/0)  0.000 [0.000,0.037] 0.100 0.000
   0.60   166    99  0.605  0.360  0.596 [0.520,0.668]    63    18    49 1.000 0.182 |    99/99       0     9     0 (0/0/0)  0.000 [0.000,0.037] 0.091 0.000
   0.65   166    96  0.587  0.337  0.578 [0.502,0.651]    66    20    50 1.000 0.156 |    96/96       0     7     0 (0/0/0)  0.000 [0.000,0.038] 0.073 0.000
   0.70   166    90  0.551  0.302  0.542 [0.466,0.616]    72    23    53 1.000 0.100 |    90/90       0     5     0 (0/0/0)  0.000 [0.000,0.041] 0.056 0.000
   0.75   166    90  0.551  0.302  0.542 [0.466,0.616]    72    23    53 1.000 0.100 |    90/90       0     5     0 (0/0/0)  0.000 [0.000,0.041] 0.056 0.000
   0.80   166    81  0.497  0.244  0.488 [0.413,0.563]    81    27    58 1.000 0.000 |    81/81       0     3     0 (0/0/0)  0.000 [0.000,0.045] 0.037 0.000
   0.85   166    76  0.467  0.221  0.458 [0.384,0.534]    86    31    59 1.000 0.000 |    76/76       0     0     0 (0/0/0)  0.000 [0.000,0.048] 0.000 0.000
   0.90   166    75  0.460  0.221  0.452 [0.378,0.528]    87    32    59 1.000 0.000 |    75/75       0     0     0 (0/0/0)  0.000 [0.000,0.049] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T3", "73/86", "0.942"], ["T4", "76/80", "0.950"]], "50": [["T3", "71/86", "0.919"], ["T4", "73/80", "0.950"]], "70": [["T3", "26/86", "0.872"], ["T4", "64/80", "0.850"]], "90": [["T3", "19/86", "0.837"], ["T4", "56/80", "0.775"]]}
-- fastapi / decisive
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   125    68  0.572  0.375  0.544 [0.457,0.629]    85    15    42 0.636 0.735 |    68/68       0    32     0 (0/0/0)  0.000 [0.000,0.053] 0.471 0.000
   0.35   125    61  0.516  0.250  0.488 [0.402,0.575]    92    18    46 0.667 0.705 |    61/61       0    27     0 (0/0/0)  0.000 [0.000,0.059] 0.443 0.000
   0.40   125    58  0.491  0.208  0.464 [0.379,0.551]    95    18    49 0.667 0.690 |    58/58       0    25     0 (0/0/0)  0.000 [0.000,0.062] 0.431 0.000
   0.45   125    54  0.456  0.208  0.432 [0.348,0.520]    99    19    52 0.667 0.667 |    54/54       0    22     0 (0/0/0)  0.000 [0.000,0.066] 0.407 0.000
   0.50   125    50  0.415  0.208  0.400 [0.318,0.488]   103    19    56 0.714 0.640 |    50/50       0    19     0 (0/0/0)  0.000 [0.000,0.071] 0.380 0.000
   0.55   125    43  0.355  0.111  0.344 [0.266,0.431]   110    22    60 0.750 0.581 |    43/43       0    12     0 (0/0/0)  0.000 [0.000,0.082] 0.279 0.000
   0.60   125    37  0.295  0.111  0.296 [0.223,0.381]   116    22    66 0.778 0.514 |    37/37       0     9     0 (0/0/0)  0.000 [0.000,0.094] 0.243 0.000
   0.65   125    32  0.255  0.111  0.256 [0.188,0.339]   121    23    70 1.000 0.438 |    32/32       0     5     0 (0/0/0)  0.000 [0.000,0.107] 0.156 0.000
   0.70   125    25  0.197  0.111  0.200 [0.139,0.279]   128    26    74 1.000 0.280 |    25/25       0     1     0 (0/0/0)  0.000 [0.000,0.133] 0.040 0.000
   0.75   125    23  0.182  0.056  0.184 [0.126,0.261]   130    26    76 1.000 0.217 |    23/23       0     1     0 (0/0/0)  0.000 [0.000,0.143] 0.043 0.000
   0.80   125    18  0.141  0.000  0.144 [0.093,0.216]   135    26    81 1.000 0.000 |    18/18       0     0     0 (0/0/0)  0.000 [0.000,0.176] 0.000 0.000
   0.85   125    18  0.141  0.000  0.144 [0.093,0.216]   135    26    81 1.000 0.000 |    18/18       0     0     0 (0/0/0)  0.000 [0.000,0.176] 0.000 0.000
   0.90   125    17  0.130  0.000  0.136 [0.087,0.207]   136    26    82 1.000 0.000 |    17/17       0     0     0 (0/0/0)  0.000 [0.000,0.184] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "13/18", "0.889"], ["T2", "11/18", "0.889"], ["T3", "7/18", "1.000"], ["T4", "9/24", "0.625"], ["T5", "13/15", "0.933"], ["T6", "15/32", "0.969"]], "50": [["T1", "9/18", "0.833"], ["T2", "10/18", "0.889"], ["T3", "4/18", "0.944"], ["T4", "5/24", "0.583"], ["T5", "9/15", "0.933"], ["T6", "13/32", "0.938"]], "70": [["T1", "2/18", "0.833"], ["T2", "2/18", "0.833"], ["T3", "2/18", "0.833"], ["T4", "4/24", "0.583"], ["T5", "6/15", "0.933"], ["T6", "9/32", "0.812"]], "90": [["T1", "0/18", "0.833"], ["T2", "1/18", "0.833"], ["T3", "1/18", "0.833"], ["T4", "2/24", "0.583"], ["T5", "5/15", "0.933"], ["T6", "8/32", "0.812"]]}
-- fastapi / remap
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    78    40  0.524  0.375  0.513 [0.404,0.621]    43    13    25 0.591 0.900 |    40/40       0    21     0 (0/0/0)  0.000 [0.000,0.088] 0.525 0.000
   0.35    78    35  0.465  0.250  0.449 [0.343,0.559]    48    16    27 0.600 0.886 |    35/35       0    17     0 (0/0/0)  0.000 [0.000,0.099] 0.486 0.000
   0.40    78    32  0.427  0.208  0.410 [0.308,0.521]    51    16    30 0.636 0.875 |    32/32       0    15     0 (0/0/0)  0.000 [0.000,0.107] 0.469 0.000
   0.45    78    30  0.399  0.208  0.385 [0.284,0.496]    53    16    32 0.640 0.867 |    30/30       0    14     0 (0/0/0)  0.000 [0.000,0.114] 0.467 0.000
   0.50    78    28  0.372  0.208  0.359 [0.261,0.470]    55    16    34 0.667 0.857 |    28/28       0    13     0 (0/0/0)  0.000 [0.000,0.121] 0.464 0.000
   0.55    78    22  0.288  0.111  0.282 [0.194,0.390]    61    18    38 0.688 0.818 |    22/22       0     7     0 (0/0/0)  0.000 [0.000,0.149] 0.318 0.000
   0.60    78    18  0.233  0.111  0.231 [0.151,0.336]    65    18    42 0.750 0.778 |    18/18       0     6     0 (0/0/0)  0.000 [0.000,0.176] 0.333 0.000
   0.65    78    14  0.181  0.111  0.179 [0.110,0.279]    69    18    46 0.750 0.714 |    14/14       0     3     0 (0/0/0)  0.000 [0.000,0.215] 0.214 0.000
   0.70    78    10  0.125  0.111  0.128 [0.071,0.220]    73    19    49 0.778 0.600 |    10/10       0     1     0 (0/0/0)  0.000 [0.000,0.278] 0.100 0.000
   0.75    78     9  0.111  0.056  0.115 [0.062,0.205]    74    19    50 0.778 0.556 |     9/9        0     1     0 (0/0/0)  0.000 [0.000,0.299] 0.111 0.000
   0.80    78     4  0.049  0.000  0.051 [0.020,0.125]    79    19    55 1.000 0.000 |     4/4        0     0     0 (0/0/0)  0.000 [0.000,0.490] 0.000 0.000
   0.85    78     4  0.049  0.000  0.051 [0.020,0.125]    79    19    55 1.000 0.000 |     4/4        0     0     0 (0/0/0)  0.000 [0.000,0.490] 0.000 0.000
   0.90    78     4  0.049  0.000  0.051 [0.020,0.125]    79    19    55 1.000 0.000 |     4/4        0     0     0 (0/0/0)  0.000 [0.000,0.490] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "13/18", "0.889"], ["T2", "11/18", "0.889"], ["T3", "7/18", "1.000"], ["T4", "9/24", "0.625"]], "50": [["T1", "9/18", "0.833"], ["T2", "10/18", "0.889"], ["T3", "4/18", "0.944"], ["T4", "5/24", "0.583"]], "70": [["T1", "2/18", "0.833"], ["T2", "2/18", "0.833"], ["T3", "2/18", "0.833"], ["T4", "4/24", "0.583"]], "90": [["T1", "0/18", "0.833"], ["T2", "1/18", "0.833"], ["T3", "1/18", "0.833"], ["T4", "2/24", "0.583"]]}
-- fastapi / mutation
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    47    28  0.668  0.469  0.596 [0.453,0.724]    42     2    17 0.875 0.500 |    28/28       0    11     0 (0/0/0)  0.000 [0.000,0.121] 0.393 0.000
   0.35    47    26  0.619  0.438  0.553 [0.412,0.686]    44     2    19 1.000 0.462 |    26/26       0    10     0 (0/0/0)  0.000 [0.000,0.129] 0.385 0.000
   0.40    47    26  0.619  0.438  0.553 [0.412,0.686]    44     2    19 1.000 0.462 |    26/26       0    10     0 (0/0/0)  0.000 [0.000,0.129] 0.385 0.000
   0.45    47    24  0.570  0.406  0.511 [0.372,0.647]    46     3    20 1.000 0.417 |    24/24       0     8     0 (0/0/0)  0.000 [0.000,0.138] 0.333 0.000
   0.50    47    22  0.503  0.406  0.468 [0.333,0.608]    48     3    22 1.000 0.364 |    22/22       0     6     0 (0/0/0)  0.000 [0.000,0.149] 0.273 0.000
   0.55    47    21  0.487  0.375  0.447 [0.314,0.588]    49     4    22 1.000 0.333 |    21/21       0     5     0 (0/0/0)  0.000 [0.000,0.155] 0.238 0.000
   0.60    47    19  0.421  0.375  0.404 [0.276,0.547]    51     4    24 1.000 0.263 |    19/19       0     3     0 (0/0/0)  0.000 [0.000,0.168] 0.158 0.000
   0.65    47    18  0.405  0.344  0.383 [0.258,0.526]    52     5    24 1.000 0.222 |    18/18       0     2     0 (0/0/0)  0.000 [0.000,0.176] 0.111 0.000
   0.70    47    15  0.341  0.281  0.319 [0.204,0.462]    55     7    25 1.000 0.067 |    15/15       0     0     0 (0/0/0)  0.000 [0.000,0.204] 0.000 0.000
   0.75    47    14  0.325  0.250  0.298 [0.187,0.440]    56     7    26 1.000 0.000 |    14/14       0     0     0 (0/0/0)  0.000 [0.000,0.215] 0.000 0.000
   0.80    47    14  0.325  0.250  0.298 [0.187,0.440]    56     7    26 1.000 0.000 |    14/14       0     0     0 (0/0/0)  0.000 [0.000,0.215] 0.000 0.000
   0.85    47    14  0.325  0.250  0.298 [0.187,0.440]    56     7    26 1.000 0.000 |    14/14       0     0     0 (0/0/0)  0.000 [0.000,0.215] 0.000 0.000
   0.90    47    13  0.292  0.250  0.277 [0.169,0.418]    57     7    27 1.000 0.000 |    13/13       0     0     0 (0/0/0)  0.000 [0.000,0.228] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T5", "13/15", "0.933"], ["T6", "15/32", "0.969"]], "50": [["T5", "9/15", "0.933"], ["T6", "13/32", "0.938"]], "70": [["T5", "6/15", "0.933"], ["T6", "9/32", "0.812"]], "90": [["T5", "5/15", "0.933"], ["T6", "8/32", "0.812"]]}
-- fastapi / fresh
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    36    24  0.667  0.611  0.667 [0.503,0.798]    20     4     8 0.591 0.958 |    24/24       0    13     0 (0/0/0)  0.000 [0.000,0.138] 0.542 0.000
   0.35    36    23  0.639  0.611  0.639 [0.476,0.775]    21     5     8 0.591 0.957 |    23/23       0    12     0 (0/0/0)  0.000 [0.000,0.143] 0.522 0.000
   0.40    36    21  0.583  0.556  0.583 [0.422,0.729]    23     5    10 0.600 0.952 |    21/21       0    11     0 (0/0/0)  0.000 [0.000,0.155] 0.524 0.000
   0.45    36    20  0.556  0.556  0.556 [0.396,0.705]    24     5    11 0.600 0.950 |    20/20       0    10     0 (0/0/0)  0.000 [0.000,0.161] 0.500 0.000
   0.50    36    19  0.528  0.500  0.528 [0.370,0.680]    25     5    12 0.600 0.947 |    19/19       0    10     0 (0/0/0)  0.000 [0.000,0.168] 0.526 0.000
   0.55    36    15  0.417  0.389  0.417 [0.271,0.578]    29     5    16 0.667 0.933 |    15/15       0     6     0 (0/0/0)  0.000 [0.000,0.204] 0.400 0.000
   0.60    36    11  0.306  0.278  0.306 [0.180,0.469]    33     5    20 0.667 0.909 |    11/11       0     5     0 (0/0/0)  0.000 [0.000,0.259] 0.455 0.000
   0.65    36     8  0.222  0.167  0.222 [0.117,0.381]    36     5    23 0.714 0.875 |     8/8        0     3     0 (0/0/0)  0.000 [0.000,0.324] 0.375 0.000
   0.70    36     4  0.111  0.111  0.111 [0.044,0.253]    40     6    26 0.778 0.750 |     4/4        0     1     0 (0/0/0)  0.000 [0.000,0.490] 0.250 0.000
   0.75    36     3  0.083  0.056  0.083 [0.029,0.218]    41     6    27 0.778 0.667 |     3/3        0     1     0 (0/0/0)  0.000 [0.000,0.561] 0.333 0.000
   0.80    36     1  0.028  0.000  0.028 [0.005,0.142]    43     6    29 1.000 0.000 |     1/1        0     0     0 (0/0/0)  0.000 [0.000,0.793] 0.000 0.000
   0.85    36     1  0.028  0.000  0.028 [0.005,0.142]    43     6    29 1.000 0.000 |     1/1        0     0     0 (0/0/0)  0.000 [0.000,0.793] 0.000 0.000
   0.90    36     1  0.028  0.000  0.028 [0.005,0.142]    43     6    29 1.000 0.000 |     1/1        0     0     0 (0/0/0)  0.000 [0.000,0.793] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "13/18", "0.889"], ["T2", "11/18", "0.889"]], "50": [["T1", "9/18", "0.833"], ["T2", "10/18", "0.889"]], "70": [["T1", "2/18", "0.833"], ["T2", "2/18", "0.833"]], "90": [["T1", "0/18", "0.833"], ["T2", "1/18", "0.833"]]}
-- fastapi / same_state
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    42    16  0.382  0.375  0.381 [0.250,0.532]    23     9    17 0.500 0.812 |    16/16       0     8     0 (0/0/0)  0.000 [0.000,0.194] 0.500 0.000
   0.35    42    12  0.292  0.250  0.286 [0.172,0.436]    27    11    19 0.750 0.750 |    12/12       0     5     0 (0/0/0)  0.000 [0.000,0.242] 0.417 0.000
   0.40    42    11  0.271  0.208  0.262 [0.153,0.411]    28    11    20 0.750 0.727 |    11/11       0     4     0 (0/0/0)  0.000 [0.000,0.259] 0.364 0.000
   0.45    42    10  0.243  0.208  0.238 [0.135,0.385]    29    11    21 0.750 0.700 |    10/10       0     4     0 (0/0/0)  0.000 [0.000,0.278] 0.400 0.000
   0.50    42     9  0.215  0.208  0.214 [0.117,0.359]    30    11    22 0.750 0.667 |     9/9        0     3     0 (0/0/0)  0.000 [0.000,0.299] 0.333 0.000
   0.55    42     7  0.160  0.111  0.167 [0.083,0.306]    32    13    22 0.750 0.571 |     7/7        0     1     0 (0/0/0)  0.000 [0.000,0.354] 0.143 0.000
   0.60    42     7  0.160  0.111  0.167 [0.083,0.306]    32    13    22 0.750 0.571 |     7/7        0     1     0 (0/0/0)  0.000 [0.000,0.354] 0.143 0.000
   0.65    42     6  0.139  0.111  0.143 [0.067,0.278]    33    13    23 1.000 0.500 |     6/6        0     0     0 (0/0/0)  0.000 [0.000,0.390] 0.000 0.000
   0.70    42     6  0.139  0.111  0.143 [0.067,0.278]    33    13    23 1.000 0.500 |     6/6        0     0     0 (0/0/0)  0.000 [0.000,0.390] 0.000 0.000
   0.75    42     6  0.139  0.111  0.143 [0.067,0.278]    33    13    23 1.000 0.500 |     6/6        0     0     0 (0/0/0)  0.000 [0.000,0.390] 0.000 0.000
   0.80    42     3  0.069  0.056  0.071 [0.025,0.190]    36    13    26 1.000 0.000 |     3/3        0     0     0 (0/0/0)  0.000 [0.000,0.561] 0.000 0.000
   0.85    42     3  0.069  0.056  0.071 [0.025,0.190]    36    13    26 1.000 0.000 |     3/3        0     0     0 (0/0/0)  0.000 [0.000,0.561] 0.000 0.000
   0.90    42     3  0.069  0.056  0.071 [0.025,0.190]    36    13    26 1.000 0.000 |     3/3        0     0     0 (0/0/0)  0.000 [0.000,0.561] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T3", "7/18", "1.000"], ["T4", "9/24", "0.625"]], "50": [["T3", "4/18", "0.944"], ["T4", "5/24", "0.583"]], "70": [["T3", "2/18", "0.833"], ["T4", "4/24", "0.583"]], "90": [["T3", "1/18", "0.833"], ["T4", "2/24", "0.583"]]}

== X2b (X2b PREREGISTRATION sections 2, 4a)
served/flags check: ok
X2b expected labelled n per repository ~ {"axum": 365, "fastapi": 125} [INFERENCE: 0.8 x 6 transitions x the smoke-map agent-feature counts {"axum": 76, "fastapi": 26}]; the '216-270' band printed above is X2's, not X2b's; X2's c3_stops_binding_at_n=140 is a multiple-of-20 search; at integer n C3 can bind only for n <= 126 (PREREGISTRATION section 4(iv))
X2b operating characteristics (n=100 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 4 spurious; both repos): {"0.03": 0.6689, "0.04": 0.3955, "0.05": 0.1901, "0.07": 0.0266, "0.1": 0.0006}
X2b operating characteristics (n=125 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 5 spurious; both repos): {"0.03": 0.6817, "0.04": 0.3794, "0.05": 0.1612, "0.07": 0.015, "0.1": 0.0001}
X2b operating characteristics (n=150 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 7 spurious; both repos): {"0.03": 0.8401, "0.04": 0.5578, "0.05": 0.2733, "0.07": 0.0285, "0.1": 0.0002}
X2b operating characteristics (n=216 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 10 spurious; both repos): {"0.03": 0.8784, "0.04": 0.5639, "0.05": 0.2318, "0.07": 0.0109, "0.1": 0.0}
X2b operating characteristics (n=270 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 13 spurious; both repos): {"0.03": 0.9314, "0.04": 0.6456, "0.05": 0.2671, "0.07": 0.0088, "0.1": 0.0}
combination: X2b verdict fallback, t* None, citable True; X2 codex (frozen) {"available": true, "citable": true, "models": ["gpt-6.1-sol"], "pass_pct": [30], "threshold_pct": 30, "verdict": "adopt"}; X2b t* in X2 pass set: None; X2's t* status in X2b (per-threshold at X2's t*): fail
PROPOSAL (for a design round, not adopted here): no threshold on the 0.30-0.90 grid meets C1-C3 for gpt-6-luna xhigh (nothing below 0.30 measured): the design's fallback applies (agent map at onboarding and manually); X2's codex adopt is not applied to a model Scout no longer uses
```
