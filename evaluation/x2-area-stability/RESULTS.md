# X2 — Do Scout areas stabilise agent mapper output? Results

Preregistration: `PREREGISTRATION.md` revision 10 (= signed-off rev 8, `reviews/round-8.md`, + Amendment 1, `reviews/round-10.md`; both with zero council issues). Hypothesis: **HD5** (`docs/hypothesis-ledger.md`). Governed by `AGENTS.md` experiment rules.
Data: collected 2026-10-23 18:42–19:56Z (28 of 28 maps `rc` 0, one attempt each; `logs/x2-collect.log`), labelled 19:56Z to about 21:20Z (`logs/x2-label.log`; the end time is as reported by the operator, the log itself carries no timestamps), analysed with `run.sh analysis` (`data/results.json`, `data/analysis_output.txt`).
This study supersedes no earlier study and is superseded by none. The first, rev-8 collection (`data-rev8-unpatched/`) is **excluded** (Deviations, D1). Every number below is copied from `data/results.json` or `data/analysis_output.txt` (full text in Appendix A); nothing was recomputed outside those files except the tabulations noted. `[INFERENCE]` marks statements that are not read from the outputs.

## 1. Headline

**Arm A (codex, decides): `adopt`, `t* = 0.30`.** Citable for a design change: **yes** (probe-consistent: one served model, `gpt-6.1-sol`, across all four codex probes, no workspace config file, no probe error). Reason string: "largest passing threshold (survival, spurious-merge and precision criteria in both repositories)". Pass set = `{0.30}` only; the other 12 thresholds `fail` (C1). Non-contiguity flag: **contiguous: true** (a single-point pass set).

| | axum | fastapi |
|---|---|---|
| C1 mean per-refresh survival at 0.30 (exact) | **0.8035423898426994** (205559/255816; ≥ 4/5) | **0.8119702273411623** (862897/1062720; ≥ 4/5) |
| C2 spurious rate (`different` / labelled matches) at 0.30 | 0/228 = 0.000 | 0/208 = 0.000 |
| C3 Wilson 95 % upper bound of spurious (≤ 0.10) | [0.000, 0.017] | [0.000, 0.018] |
| pooled survival (areas as independent; Wilson) | 0.797 [0.747, 0.840] | 0.794 [0.741, 0.838] |
| worst transition survival | 0.6470588235294118 (T4) | 0.5740740740740741 (T5) |
| survival at 0.35 (fails C1) | 0.7806665728492354 (66569/85272) | 0.7989931496537188 (424553/531360) |
| survival at the design default 0.50 (fails C1) | 0.7207172342621259 (5587/7752) | 0.7662488708220415 (67859/88560) |

- **Survival margin at `t*`** (`survival_margin_at_t_star` = min over repos of mean survival − 4/5): **0.00354238984269944** (exact 4531/1279080), set by axum. `t*` sits at the survival edge by construction (§4); the margin is 0.35 percentage points of survival. [INFERENCE] With a margin this small, C1 at 0.30 is sensitive to which refreshes were drawn (see LOTO below and the Threats).
- **`t*` is the lowest point of the grid (0.30 = `T_MIN`)**. Labels exist only for matches at or above `T_MIN`, so nothing below 0.30 was measured; the study does not say whether a lower threshold would also pass.
- **The design default 0.50 does not meet C1 in either repository** (table above), so `adopt t*` would propose a default below the current assumption (consequence of the rule, see §4; a later council design round decides).
- **Spurious merges: zero `different_concern` labels** among the labelled matches at every threshold, in every stratum, in both arms (asserted over all tables when this file was generated). Unlabelled share at `t*`: 0 in both repos; worst-case spurious rate at `t*` ((different + unlabelled)/matches): 0.0 in both repos. The strict sensitivity (`different` + `overlapping`) at 0.30 is **0.32456140350877194** (axum) and **0.32211538461538464** (fastapi), which would fail C2 by a wide margin; per §10 both are reported and the verdict stays the preregistered one (primary count).
- **Leave-one-transition-out (joint `t*`, descriptive, §3.7)** for arm A: only holding out T1 leaves a non-empty pass set (`t* = 0.30`; held-out survival of T1: axum 0.7254901960784313, fastapi 0.8222222222222222). Holding out any of T2–T6 leaves **no** passing threshold on the remaining five transitions (`t* = none`). **Caveat (§3.7):** this view understates selection, because `t*` sits at the survival edge by construction and the adopted rule's selection on the same transitions is not removed; it is not an out-of-sample estimate. [INFERENCE] The adopted `t*` depends on the full set of six refreshes.
- **Arm B (claude, replication, non-gating): `fallback`**, no threshold passes (all 13 `fail`), gates ok (class, label; control gate pooled and ok), citable: yes (one served model `claude-opus-5-5`, no probe error). At 0.30 claude's mean survival is axum 0.7687468030690537 (15029/19550, fails C1) and fastapi 0.8 (4/5, meets C1 exactly), with zero `different` labels in both repos; claude's own `t*` is **none**. **Replication status: `not replicated`** (§4 rule clause iv: arm A probe-consistent, arm B gates pass, `t*` is not `pass` in arm B).
- **§4 ledger mapping (fixed in the preregistration; the ledger is not edited by this document):** arm A `adopt` and probe-consistent → HD5 **`supported`, scoped to the codex provider**, because arm B is `not replicated`; arm B never moves the HD5 state. Consequence of arm B (§4): `adopt t*` proposes a provider-general default only if B is `replicated`; since it is not, **the proposal is scoped to the codex provider and A1 stays an assumption for claude**. The consequence `adopt t*` is "proposes `mapping.areas.match_jaccard` default `t*`" and needs its own council design round (design rule 1); this study proposes nothing further. Any design change citing X2 must state arm B's status (`not replicated`, claude `fallback`), the contiguity flag (contiguous), and the scope limits below.
- **Scope limits (§4, §10):** one-step (per-refresh) result on the **clean state layout** (state directory outside the workspace); **two repositories** (fastapi `d666ccb6…`, Python; axum `562fa206…`, Rust), no cross-repository claim (two clusters, no cluster-robust interval); **codex-scoped** (Clawpatch 0.7.3, codex-cli 0.162.1 with the served model `gpt-6.1-sol`); one run of each of 7 maps per (arm, repo); labels are model-made and unreviewed by a human (`agent_reviewed`), under guide `x2-concern-v1/guide-1`; a mapper truly at 5 % spurious is not adopted at a given passing threshold about three times in four (§7 operating characteristics, below).

## 2. Checklist: PREREGISTRATION §11 reporting contract → sections here

| §11 item | Section |
|---|---|
| Per arm, repo, stratum (decisive, remap, mutation, fresh, same_state) and threshold: the tables of `analysis_output.txt` (survival per transition with IDs, mean and pooled with Wilson, worst transition, spurious with Wilson, strict and worst-case, unlabelled split, unique labelled pairs, created/retired/kept, Jaccard summaries) | §3 (decisive, all four arm×repo, plus per-transition tables) and **Appendix A** (the complete `analysis_output.txt` verbatim: every stratum, every threshold) |
| Area-ID retention per transition (`survival_kept_inclusive`) | §3.2 (the `ret` value in each cell) |
| Coverage diagnostics | §5 |
| Chained survival | §6 |
| Label agreement: adjudication rate, C-equals-A share with denominator, codex sessions, pairs with an exec event, served models per role | §7.1 |
| Controls: misses by kind, per arm, by source-pair band, shortfalls; over-call check | §7.2, §7.4 |
| Label integrity | §7.3 |
| Label distribution by class and threshold band | §7.4 |
| Map inventory (retries, failures, stops, aborts, skipped mutations, attempt start times) | §8 |
| `unmeasured` counts with reasons | §9 |
| Tool versions (CLI versions after the after-probe, workspace-config presence) | §10.1 |
| Model probes and per-arm probe-consistency flag with reasons | §10.2 |
| Replay provenance, digests | §10.3, §10.4 |
| Wall times | §8 |
| Verdict for both arms, arm B replication status | §1, §4 |
| Non-contiguity flag, `t*` survival margin, `AT t*` unlabelled share and worst-case rate | §1, §4 |
| Joint leave-one-transition-out `t*` with held-out survival, matched, labelled, `different` counts | §3.3 |
| Operating characteristics (§7) | §4.3 |
| Deviations (Amendment 1 and any runtime deviation) | §11 |
| Reproduce commands | §12 |
| Superseding / superseded links | header (none) |

**Gaps (§11 items the outputs do not contain): none identified.** Two partial items are stated here rather than invented: (a) the end time of the labelling pass is not recorded in `x2-label.log` (a single line, `labelled 649 pairs`); (b) the per-attempt split of codex sessions with exec events is not kept by design (§6: `pairs_with_exec_event` is the preregistered measure).

## 3. Survival and spurious merge by threshold (decisive stratum, all six transitions)

### 3.1 Decisive stratum, per arm and repo (verbatim rows of `analysis_output.txt`)
Columns: `prev` prior areas; `match` matches; `surv(mean)` mean of per-transition survival; `worstT` worst transition; `pooled [Wilson]`; `crtd/retd/kept` created, retired, kept areas; `Jmed` median Jaccard of matches (upper median); `J<.8` share of matches with Jaccard < 0.8; `lab(occ/uniq)` labelled occurrences / unique labelled pairs; `diff`, `overl` the `different` and `overlapping` counts; `unlab(ct/unp/never)` unlabelled (cannot_tell / unparsable / never labelled); `spur [Wilson]` spurious rate; `strict`, `worst` the strict and worst-case spurious rates. The 13 thresholds are on the rows.

**Arm A (codex), axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   286   228  0.804  0.647  0.797 [0.747,0.840]    22    31    27 0.833 0.474 |   228/228      0    74     0 (0/0/0)  0.000 [0.000,0.017] 0.325 0.000
   0.35   286   221  0.781  0.588  0.773 [0.721,0.817]    29    35    30 0.857 0.457 |   221/221      0    68     0 (0/0/0)  0.000 [0.000,0.017] 0.308 0.000
   0.40   286   217  0.766  0.588  0.759 [0.706,0.805]    33    37    32 0.857 0.447 |   217/217      0    64     0 (0/0/0)  0.000 [0.000,0.017] 0.295 0.000
   0.45   286   205  0.724  0.529  0.717 [0.662,0.766]    45    45    36 1.000 0.415 |   205/205      0    53     0 (0/0/0)  0.000 [0.000,0.018] 0.259 0.000
   0.50   286   204  0.721  0.529  0.713 [0.658,0.763]    46    45    37 1.000 0.412 |   204/204      0    52     0 (0/0/0)  0.000 [0.000,0.018] 0.255 0.000
   0.55   286   173  0.617  0.373  0.605 [0.547,0.660]    77    66    47 1.000 0.306 |   173/173      0    32     0 (0/0/0)  0.000 [0.000,0.022] 0.185 0.000
   0.60   286   160  0.569  0.353  0.559 [0.501,0.616]    90    77    49 1.000 0.250 |   160/160      0    20     0 (0/0/0)  0.000 [0.000,0.023] 0.125 0.000
   0.65   286   150  0.532  0.353  0.524 [0.467,0.582]   100    85    51 1.000 0.200 |   150/150      0    11     0 (0/0/0)  0.000 [0.000,0.025] 0.073 0.000
   0.70   286   137  0.487  0.333  0.479 [0.422,0.537]   113    93    56 1.000 0.124 |   137/137      0     7     0 (0/0/0)  0.000 [0.000,0.027] 0.051 0.000
   0.75   286   134  0.476  0.333  0.469 [0.411,0.526]   116    94    58 1.000 0.104 |   134/134      0     5     0 (0/0/0)  0.000 [0.000,0.028] 0.037 0.000
   0.80   286   120  0.427  0.314  0.420 [0.364,0.477]   130   100    66 1.000 0.000 |   120/120      0     1     0 (0/0/0)  0.000 [0.000,0.031] 0.008 0.000
   0.85   286   112  0.398  0.275  0.392 [0.337,0.449]   138   105    69 1.000 0.000 |   112/112      0     0     0 (0/0/0)  0.000 [0.000,0.033] 0.000 0.000
   0.90   286   104  0.370  0.216  0.364 [0.310,0.421]   146   112    70 1.000 0.000 |   104/104      0     0     0 (0/0/0)  0.000 [0.000,0.036] 0.000 0.000
```
**Arm A (codex), fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   262   208  0.812  0.574  0.794 [0.741,0.838]    55    15    39 0.909 0.327 |   208/208      0    67     0 (0/0/0)  0.000 [0.000,0.018] 0.322 0.000
   0.35   262   205  0.799  0.574  0.782 [0.729,0.828]    58    16    41 0.909 0.317 |   205/205      0    64     0 (0/0/0)  0.000 [0.000,0.018] 0.312 0.000
   0.40   262   203  0.792  0.574  0.775 [0.720,0.821]    60    17    42 0.909 0.310 |   203/203      0    63     0 (0/0/0)  0.000 [0.000,0.019] 0.310 0.000
   0.45   262   198  0.770  0.556  0.756 [0.700,0.804]    65    19    45 0.933 0.293 |   198/198      0    60     0 (0/0/0)  0.000 [0.000,0.019] 0.303 0.000
   0.50   262   197  0.766  0.556  0.752 [0.696,0.800]    66    20    45 0.933 0.289 |   197/197      0    59     0 (0/0/0)  0.000 [0.000,0.019] 0.299 0.000
   0.55   262   192  0.744  0.556  0.733 [0.676,0.783]    71    21    49 1.000 0.271 |   192/192      0    54     0 (0/0/0)  0.000 [0.000,0.020] 0.281 0.000
   0.60   262   182  0.705  0.537  0.695 [0.636,0.747]    81    23    57 1.000 0.231 |   182/182      0    45     0 (0/0/0)  0.000 [0.000,0.021] 0.247 0.000
   0.65   262   165  0.636  0.500  0.630 [0.570,0.686]    98    27    70 1.000 0.152 |   165/165      0    33     0 (0/0/0)  0.000 [0.000,0.023] 0.200 0.000
   0.70   262   156  0.601  0.463  0.595 [0.535,0.653]   107    31    75 1.000 0.103 |   156/156      0    30     0 (0/0/0)  0.000 [0.000,0.024] 0.192 0.000
   0.75   262   147  0.565  0.463  0.561 [0.501,0.620]   116    35    80 1.000 0.048 |   147/147      0    23     0 (0/0/0)  0.000 [0.000,0.025] 0.156 0.000
   0.80   262   140  0.539  0.444  0.534 [0.474,0.594]   123    36    86 1.000 0.000 |   140/140      0    18     0 (0/0/0)  0.000 [0.000,0.027] 0.129 0.000
   0.85   262   117  0.451  0.378  0.447 [0.388,0.507]   146    46    99 1.000 0.000 |   117/117      0     4     0 (0/0/0)  0.000 [0.000,0.032] 0.034 0.000
   0.90   262   108  0.418  0.333  0.412 [0.354,0.473]   155    49   105 1.000 0.000 |   108/108      0     1     0 (0/0/0)  0.000 [0.000,0.034] 0.009 0.000
```
**Arm B (claude), axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   121    88  0.769  0.280  0.727 [0.642,0.799]    29    24     9 0.636 0.614 |    88/88       0    42     0 (0/0/0)  0.000 [0.000,0.042] 0.477 0.000
   0.35   121    79  0.690  0.200  0.653 [0.565,0.732]    38    28    14 0.705 0.570 |    79/79       0    33     0 (0/0/0)  0.000 [0.000,0.046] 0.418 0.000
   0.40   121    79  0.690  0.200  0.653 [0.565,0.732]    38    28    14 0.705 0.570 |    79/79       0    33     0 (0/0/0)  0.000 [0.000,0.046] 0.418 0.000
   0.45   121    76  0.666  0.200  0.628 [0.539,0.709]    41    29    16 0.727 0.553 |    76/76       0    30     0 (0/0/0)  0.000 [0.000,0.048] 0.395 0.000
   0.50   121    75  0.659  0.200  0.620 [0.531,0.701]    42    29    17 0.727 0.547 |    75/75       0    29     0 (0/0/0)  0.000 [0.000,0.049] 0.387 0.000
   0.55   121    60  0.540  0.080  0.496 [0.408,0.584]    57    41    20 0.900 0.433 |    60/60       0    17     0 (0/0/0)  0.000 [0.000,0.060] 0.283 0.000
   0.60   121    55  0.499  0.040  0.455 [0.369,0.543]    62    44    22 0.900 0.382 |    55/55       0    14     0 (0/0/0)  0.000 [0.000,0.065] 0.255 0.000
   0.65   121    43  0.379  0.000  0.355 [0.276,0.444]    74    52    26 1.000 0.209 |    43/43       0     8     0 (0/0/0)  0.000 [0.000,0.082] 0.186 0.000
   0.70   121    41  0.355  0.000  0.339 [0.261,0.427]    76    54    26 1.000 0.171 |    41/41       0     7     0 (0/0/0)  0.000 [0.000,0.086] 0.171 0.000
   0.75   121    37  0.317  0.000  0.306 [0.231,0.393]    80    56    28 1.000 0.081 |    37/37       0     4     0 (0/0/0)  0.000 [0.000,0.094] 0.108 0.000
   0.80   121    34  0.295  0.000  0.281 [0.209,0.367]    83    59    28 1.000 0.000 |    34/34       0     2     0 (0/0/0)  0.000 [0.000,0.102] 0.059 0.000
   0.85   121    31  0.271  0.000  0.256 [0.187,0.341]    86    61    29 1.000 0.000 |    31/31       0     0     0 (0/0/0)  0.000 [0.000,0.110] 0.000 0.000
   0.90   121    30  0.264  0.000  0.248 [0.180,0.332]    87    62    29 1.000 0.000 |    30/30       0     0     0 (0/0/0)  0.000 [0.000,0.114] 0.000 0.000
```
**Arm B (claude), fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    68    54  0.800  0.667  0.794 [0.684,0.873]    11     4    10 0.826 0.481 |    54/54       0    13     0 (0/0/0)  0.000 [0.000,0.066] 0.241 0.000
   0.35    68    52  0.772  0.583  0.765 [0.651,0.850]    13     4    12 0.826 0.462 |    52/52       0    11     0 (0/0/0)  0.000 [0.000,0.069] 0.212 0.000
   0.40    68    51  0.758  0.500  0.750 [0.636,0.838]    14     5    12 0.826 0.451 |    51/51       0    11     0 (0/0/0)  0.000 [0.000,0.070] 0.216 0.000
   0.45    68    47  0.697  0.417  0.691 [0.574,0.788]    18     8    13 0.867 0.404 |    47/47       0     9     0 (0/0/0)  0.000 [0.000,0.076] 0.191 0.000
   0.50    68    45  0.669  0.333  0.662 [0.543,0.763]    20     9    14 0.909 0.378 |    45/45       0     7     0 (0/0/0)  0.000 [0.000,0.079] 0.156 0.000
   0.55    68    40  0.600  0.333  0.588 [0.470,0.697]    25    11    17 1.000 0.300 |    40/40       0     4     0 (0/0/0)  0.000 [0.000,0.088] 0.100 0.000
   0.60    68    39  0.586  0.333  0.574 [0.455,0.684]    26    12    17 1.000 0.282 |    39/39       0     3     0 (0/0/0)  0.000 [0.000,0.090] 0.077 0.000
   0.65    68    34  0.511  0.333  0.500 [0.384,0.616]    31    14    20 1.000 0.176 |    34/34       0     2     0 (0/0/0)  0.000 [0.000,0.102] 0.059 0.000
   0.70    68    33  0.494  0.333  0.485 [0.371,0.602]    32    15    20 1.000 0.152 |    33/33       0     1     0 (0/0/0)  0.000 [0.000,0.104] 0.030 0.000
   0.75    68    30  0.453  0.250  0.441 [0.329,0.559]    35    17    21 1.000 0.067 |    30/30       0     0     0 (0/0/0)  0.000 [0.000,0.114] 0.000 0.000
   0.80    68    28  0.419  0.250  0.412 [0.303,0.530]    37    18    22 1.000 0.000 |    28/28       0     0     0 (0/0/0)  0.000 [0.000,0.121] 0.000 0.000
   0.85    68    24  0.361  0.167  0.353 [0.250,0.472]    41    20    24 1.000 0.000 |    24/24       0     0     0 (0/0/0)  0.000 [0.000,0.138] 0.000 0.000
   0.90    68    23  0.344  0.167  0.338 [0.237,0.457]    42    21    24 1.000 0.000 |    23/23       0     0     0 (0/0/0)  0.000 [0.000,0.143] 0.000 0.000
```
The `remap`, `mutation`, `fresh` and `same_state` strata, per repo and arm, are in Appendix A with the same columns.

### 3.2 Per-transition survival and area-ID retention, all thresholds (decisive)
Each cell is `matched/prior areas = survival (ret = kept-inclusive survival = (matched + kept)/prior areas, the design's "area-ID retention")`. Transition IDs: T1, T2 = F1→F2, F1→F3 (remap/fresh); T3, T4 = F1→R1, R1→R2 (remap/same_state); T5, T6 = R2→M1, M1→M2 (mutation). Mean is the exact-rational mean of the six per-transition survivals; pooled is the area-pooled proportion with its Wilson interval.

**Arm A (codex), axum**

| t | T1 | T2 | T3 | T4 | T5 | T6 | mean (exact) | pooled [Wilson] |
|---|---|---|---|---|---|---|---|---|
| 0.30 | 37/51 = 0.725 (ret 0.804) | 42/51 = 0.824 (ret 0.882) | 45/51 = 0.882 (ret 0.922) | 33/51 = 0.647 (ret 0.961) | 36/38 = 0.947 (ret 0.947) | 35/44 = 0.795 (ret 0.841) | 0.8035 (205559/255816) | 0.797 [0.747, 0.840] |
| 0.35 | 36/51 = 0.706 (ret 0.784) | 42/51 = 0.824 (ret 0.882) | 42/51 = 0.824 (ret 0.863) | 30/51 = 0.588 (ret 0.961) | 36/38 = 0.947 (ret 0.947) | 35/44 = 0.795 (ret 0.841) | 0.7807 (66569/85272) | 0.773 [0.721, 0.817] |
| 0.40 | 35/51 = 0.686 (ret 0.784) | 41/51 = 0.804 (ret 0.882) | 41/51 = 0.804 (ret 0.843) | 30/51 = 0.588 (ret 0.961) | 35/38 = 0.921 (ret 0.921) | 35/44 = 0.795 (ret 0.841) | 0.7665 (65359/85272) | 0.759 [0.706, 0.805] |
| 0.45 | 33/51 = 0.647 (ret 0.745) | 40/51 = 0.784 (ret 0.863) | 39/51 = 0.765 (ret 0.804) | 27/51 = 0.529 (ret 0.941) | 33/38 = 0.868 (ret 0.895) | 33/44 = 0.750 (ret 0.818) | 0.7240 (16837/23256) | 0.717 [0.662, 0.766] |
| 0.50 | 33/51 = 0.647 (ret 0.745) | 40/51 = 0.784 (ret 0.863) | 38/51 = 0.745 (ret 0.804) | 27/51 = 0.529 (ret 0.941) | 33/38 = 0.868 (ret 0.895) | 33/44 = 0.750 (ret 0.818) | 0.7207 (5587/7752) | 0.713 [0.658, 0.763] |
| 0.55 | 27/51 = 0.529 (ret 0.686) | 38/51 = 0.745 (ret 0.824) | 26/51 = 0.510 (ret 0.588) | 19/51 = 0.373 (ret 0.902) | 31/38 = 0.816 (ret 0.842) | 32/44 = 0.727 (ret 0.795) | 0.6167 (78875/127908) | 0.605 [0.547, 0.660] |
| 0.60 | 24/51 = 0.471 (ret 0.627) | 36/51 = 0.706 (ret 0.804) | 25/51 = 0.490 (ret 0.569) | 18/51 = 0.353 (ret 0.902) | 27/38 = 0.711 (ret 0.737) | 30/44 = 0.682 (ret 0.750) | 0.5687 (18184/31977) | 0.559 [0.501, 0.616] |
| 0.65 | 22/51 = 0.431 (ret 0.588) | 33/51 = 0.647 (ret 0.765) | 24/51 = 0.471 (ret 0.549) | 18/51 = 0.353 (ret 0.902) | 24/38 = 0.632 (ret 0.658) | 29/44 = 0.659 (ret 0.750) | 0.5321 (136121/255816) | 0.524 [0.467, 0.582] |
| 0.70 | 19/51 = 0.373 (ret 0.529) | 32/51 = 0.627 (ret 0.745) | 20/51 = 0.392 (ret 0.510) | 17/51 = 0.333 (ret 0.882) | 23/38 = 0.605 (ret 0.658) | 26/44 = 0.591 (ret 0.727) | 0.4869 (15571/31977) | 0.479 [0.422, 0.537] |
| 0.75 | 19/51 = 0.373 (ret 0.529) | 32/51 = 0.627 (ret 0.745) | 19/51 = 0.373 (ret 0.490) | 17/51 = 0.333 (ret 0.882) | 23/38 = 0.605 (ret 0.658) | 24/44 = 0.545 (ret 0.727) | 0.4761 (20299/42636) | 0.469 [0.411, 0.526] |
| 0.80 | 16/51 = 0.314 (ret 0.490) | 29/51 = 0.569 (ret 0.706) | 16/51 = 0.314 (ret 0.490) | 16/51 = 0.314 (ret 0.882) | 21/38 = 0.553 (ret 0.632) | 22/44 = 0.500 (ret 0.705) | 0.4271 (2483/5814) | 0.420 [0.364, 0.477] |
| 0.85 | 15/51 = 0.294 (ret 0.471) | 28/51 = 0.549 (ret 0.686) | 15/51 = 0.294 (ret 0.490) | 14/51 = 0.275 (ret 0.863) | 19/38 = 0.500 (ret 0.605) | 21/44 = 0.477 (ret 0.682) | 0.3982 (1787/4488) | 0.392 [0.337, 0.449] |
| 0.90 | 13/51 = 0.255 (ret 0.431) | 27/51 = 0.529 (ret 0.667) | 15/51 = 0.294 (ret 0.490) | 11/51 = 0.216 (ret 0.824) | 18/38 = 0.474 (ret 0.579) | 20/44 = 0.455 (ret 0.659) | 0.3704 (1316/3553) | 0.364 [0.310, 0.421] |

**Arm A (codex), fastapi**

| t | T1 | T2 | T3 | T4 | T5 | T6 | mean (exact) | pooled [Wilson] |
|---|---|---|---|---|---|---|---|---|
| 0.30 | 37/45 = 0.822 (ret 0.956) | 36/45 = 0.800 (ret 0.911) | 34/45 = 0.756 (ret 0.889) | 39/41 = 0.951 (ret 1.000) | 31/54 = 0.574 (ret 0.926) | 31/32 = 0.969 (ret 1.000) | 0.8120 (862897/1062720) | 0.794 [0.741, 0.838] |
| 0.35 | 37/45 = 0.822 (ret 0.956) | 35/45 = 0.778 (ret 0.889) | 34/45 = 0.756 (ret 0.889) | 38/41 = 0.927 (ret 1.000) | 31/54 = 0.574 (ret 0.926) | 30/32 = 0.938 (ret 1.000) | 0.7990 (424553/531360) | 0.782 [0.729, 0.828] |
| 0.40 | 37/45 = 0.822 (ret 0.956) | 35/45 = 0.778 (ret 0.889) | 32/45 = 0.711 (ret 0.867) | 38/41 = 0.927 (ret 1.000) | 31/54 = 0.574 (ret 0.926) | 30/32 = 0.938 (ret 1.000) | 0.7916 (420617/531360) | 0.775 [0.720, 0.821] |
| 0.45 | 37/45 = 0.822 (ret 0.956) | 35/45 = 0.778 (ret 0.889) | 31/45 = 0.689 (ret 0.844) | 37/41 = 0.902 (ret 1.000) | 30/54 = 0.556 (ret 0.926) | 28/32 = 0.875 (ret 0.969) | 0.7703 (68219/88560) | 0.756 [0.700, 0.804] |
| 0.50 | 37/45 = 0.822 (ret 0.956) | 35/45 = 0.778 (ret 0.889) | 31/45 = 0.689 (ret 0.844) | 36/41 = 0.878 (ret 0.976) | 30/54 = 0.556 (ret 0.926) | 28/32 = 0.875 (ret 0.969) | 0.7662 (67859/88560) | 0.752 [0.696, 0.800] |
| 0.55 | 37/45 = 0.822 (ret 0.956) | 33/45 = 0.733 (ret 0.867) | 31/45 = 0.689 (ret 0.844) | 35/41 = 0.854 (ret 0.976) | 30/54 = 0.556 (ret 0.926) | 26/32 = 0.812 (ret 0.969) | 0.7444 (4883/6560) | 0.733 [0.676, 0.783] |
| 0.60 | 37/45 = 0.822 (ret 0.956) | 31/45 = 0.689 (ret 0.844) | 28/45 = 0.622 (ret 0.822) | 32/41 = 0.780 (ret 0.976) | 29/54 = 0.537 (ret 0.926) | 25/32 = 0.781 (ret 0.969) | 0.7054 (749591/1062720) | 0.695 [0.636, 0.747] |
| 0.65 | 34/45 = 0.756 (ret 0.933) | 30/45 = 0.667 (ret 0.844) | 25/45 = 0.556 (ret 0.822) | 28/41 = 0.683 (ret 0.927) | 27/54 = 0.500 (ret 0.907) | 21/32 = 0.656 (ret 0.969) | 0.6362 (225353/354240) | 0.630 [0.570, 0.686] |
| 0.70 | 32/45 = 0.711 (ret 0.911) | 29/45 = 0.644 (ret 0.844) | 24/45 = 0.533 (ret 0.800) | 27/41 = 0.659 (ret 0.927) | 25/54 = 0.463 (ret 0.889) | 19/32 = 0.594 (ret 0.938) | 0.6007 (127673/212544) | 0.595 [0.535, 0.653] |
| 0.75 | 30/45 = 0.667 (ret 0.889) | 28/45 = 0.622 (ret 0.822) | 22/45 = 0.489 (ret 0.778) | 24/41 = 0.585 (ret 0.902) | 25/54 = 0.463 (ret 0.889) | 18/32 = 0.562 (ret 0.938) | 0.5648 (60019/106272) | 0.561 [0.501, 0.620] |
| 0.80 | 28/45 = 0.622 (ret 0.867) | 26/45 = 0.578 (ret 0.822) | 21/45 = 0.467 (ret 0.778) | 23/41 = 0.561 (ret 0.902) | 24/54 = 0.444 (ret 0.889) | 18/32 = 0.562 (ret 0.938) | 0.5391 (19097/35424) | 0.534 [0.474, 0.594] |
| 0.85 | 25/45 = 0.556 (ret 0.844) | 21/45 = 0.467 (ret 0.800) | 17/45 = 0.378 (ret 0.733) | 17/41 = 0.415 (ret 0.829) | 21/54 = 0.389 (ret 0.852) | 16/32 = 0.500 (ret 0.906) | 0.4506 (2494/5535) | 0.447 [0.388, 0.507] |
| 0.90 | 24/45 = 0.533 (ret 0.844) | 20/45 = 0.444 (ret 0.778) | 16/45 = 0.356 (ret 0.733) | 14/41 = 0.341 (ret 0.829) | 18/54 = 0.333 (ret 0.815) | 16/32 = 0.500 (ret 0.906) | 0.4180 (617/1476) | 0.412 [0.354, 0.473] |

**Arm B (claude), axum**

| t | T1 | T2 | T3 | T4 | T5 | T6 | mean (exact) | pooled [Wilson] |
|---|---|---|---|---|---|---|---|---|
| 0.30 | 17/23 = 0.739 (ret 0.783) | 21/23 = 0.913 (ret 0.913) | 17/23 = 0.739 (ret 0.783) | 16/17 = 0.941 (ret 0.941) | 7/25 = 0.280 (ret 0.560) | 10/10 = 1.000 (ret 1.000) | 0.7687 (15029/19550) | 0.727 [0.642, 0.799] |
| 0.35 | 17/23 = 0.739 (ret 0.783) | 19/23 = 0.826 (ret 0.870) | 15/23 = 0.652 (ret 0.696) | 14/17 = 0.824 (ret 0.882) | 5/25 = 0.200 (ret 0.560) | 9/10 = 0.900 (ret 1.000) | 0.6902 (5397/7820) | 0.653 [0.565, 0.732] |
| 0.40 | 17/23 = 0.739 (ret 0.783) | 19/23 = 0.826 (ret 0.870) | 15/23 = 0.652 (ret 0.696) | 14/17 = 0.824 (ret 0.882) | 5/25 = 0.200 (ret 0.560) | 9/10 = 0.900 (ret 1.000) | 0.6902 (5397/7820) | 0.653 [0.565, 0.732] |
| 0.45 | 16/23 = 0.696 (ret 0.739) | 18/23 = 0.783 (ret 0.870) | 15/23 = 0.652 (ret 0.696) | 13/17 = 0.765 (ret 0.882) | 5/25 = 0.200 (ret 0.560) | 9/10 = 0.900 (ret 1.000) | 0.6659 (5207/7820) | 0.628 [0.539, 0.709] |
| 0.50 | 16/23 = 0.696 (ret 0.739) | 17/23 = 0.739 (ret 0.870) | 15/23 = 0.652 (ret 0.696) | 13/17 = 0.765 (ret 0.882) | 5/25 = 0.200 (ret 0.560) | 9/10 = 0.900 (ret 1.000) | 0.6586 (15451/23460) | 0.620 [0.531, 0.701] |
| 0.55 | 12/23 = 0.522 (ret 0.609) | 13/23 = 0.565 (ret 0.696) | 13/23 = 0.565 (ret 0.609) | 12/17 = 0.706 (ret 0.824) | 2/25 = 0.080 (ret 0.480) | 8/10 = 0.800 (ret 1.000) | 0.5397 (15826/29325) | 0.496 [0.408, 0.584] |
| 0.60 | 12/23 = 0.522 (ret 0.609) | 12/23 = 0.522 (ret 0.652) | 12/23 = 0.522 (ret 0.565) | 10/17 = 0.588 (ret 0.765) | 1/25 = 0.040 (ret 0.480) | 8/10 = 0.800 (ret 1.000) | 0.4989 (29261/58650) | 0.455 [0.369, 0.543] |
| 0.65 | 9/23 = 0.391 (ret 0.522) | 11/23 = 0.478 (ret 0.609) | 10/23 = 0.435 (ret 0.565) | 8/17 = 0.471 (ret 0.647) | 0/25 = 0.000 (ret 0.440) | 5/10 = 0.500 (ret 0.800) | 0.3792 (593/1564) | 0.355 [0.276, 0.444] |
| 0.70 | 9/23 = 0.391 (ret 0.522) | 10/23 = 0.435 (ret 0.565) | 10/23 = 0.435 (ret 0.565) | 8/17 = 0.471 (ret 0.647) | 0/25 = 0.000 (ret 0.440) | 4/10 = 0.400 (ret 0.700) | 0.3552 (1389/3910) | 0.339 [0.261, 0.427] |
| 0.75 | 7/23 = 0.304 (ret 0.478) | 9/23 = 0.391 (ret 0.522) | 10/23 = 0.435 (ret 0.565) | 8/17 = 0.471 (ret 0.647) | 0/25 = 0.000 (ret 0.440) | 3/10 = 0.300 (ret 0.700) | 0.3168 (7433/23460) | 0.306 [0.231, 0.393] |
| 0.80 | 6/23 = 0.261 (ret 0.435) | 9/23 = 0.391 (ret 0.522) | 8/23 = 0.348 (ret 0.478) | 8/17 = 0.471 (ret 0.647) | 0/25 = 0.000 (ret 0.440) | 3/10 = 0.300 (ret 0.700) | 0.2951 (301/1020) | 0.281 [0.209, 0.367] |
| 0.85 | 5/23 = 0.217 (ret 0.391) | 9/23 = 0.391 (ret 0.522) | 7/23 = 0.304 (ret 0.435) | 7/17 = 0.412 (ret 0.647) | 0/25 = 0.000 (ret 0.440) | 3/10 = 0.300 (ret 0.700) | 0.2708 (6353/23460) | 0.256 [0.187, 0.341] |
| 0.90 | 5/23 = 0.217 (ret 0.391) | 8/23 = 0.348 (ret 0.478) | 7/23 = 0.304 (ret 0.435) | 7/17 = 0.412 (ret 0.647) | 0/25 = 0.000 (ret 0.440) | 3/10 = 0.300 (ret 0.700) | 0.2636 (2061/7820) | 0.248 [0.180, 0.332] |

**Arm B (claude), fastapi**

| t | T1 | T2 | T3 | T4 | T5 | T6 | mean (exact) | pooled [Wilson] |
|---|---|---|---|---|---|---|---|---|
| 0.30 | 10/12 = 0.833 (ret 0.917) | 8/12 = 0.667 (ret 1.000) | 9/12 = 0.750 (ret 0.917) | 9/10 = 0.900 (ret 1.000) | 9/12 = 0.750 (ret 0.833) | 9/10 = 0.900 (ret 1.000) | 0.8000 (4/5) | 0.794 [0.684, 0.873] |
| 0.35 | 9/12 = 0.750 (ret 0.917) | 7/12 = 0.583 (ret 1.000) | 9/12 = 0.750 (ret 0.917) | 9/10 = 0.900 (ret 1.000) | 9/12 = 0.750 (ret 0.833) | 9/10 = 0.900 (ret 1.000) | 0.7722 (139/180) | 0.765 [0.651, 0.850] |
| 0.40 | 9/12 = 0.750 (ret 0.917) | 6/12 = 0.500 (ret 0.917) | 9/12 = 0.750 (ret 0.917) | 9/10 = 0.900 (ret 1.000) | 9/12 = 0.750 (ret 0.833) | 9/10 = 0.900 (ret 1.000) | 0.7583 (91/120) | 0.750 [0.636, 0.838] |
| 0.45 | 9/12 = 0.750 (ret 0.917) | 5/12 = 0.417 (ret 0.833) | 9/12 = 0.750 (ret 0.917) | 7/10 = 0.700 (ret 0.800) | 8/12 = 0.667 (ret 0.833) | 9/10 = 0.900 (ret 1.000) | 0.6972 (251/360) | 0.691 [0.574, 0.788] |
| 0.50 | 9/12 = 0.750 (ret 0.917) | 4/12 = 0.333 (ret 0.750) | 8/12 = 0.667 (ret 0.917) | 7/10 = 0.700 (ret 0.800) | 8/12 = 0.667 (ret 0.833) | 9/10 = 0.900 (ret 1.000) | 0.6694 (241/360) | 0.662 [0.543, 0.763] |
| 0.55 | 8/12 = 0.667 (ret 0.917) | 4/12 = 0.333 (ret 0.750) | 6/12 = 0.500 (ret 0.917) | 7/10 = 0.700 (ret 0.800) | 6/12 = 0.500 (ret 0.667) | 9/10 = 0.900 (ret 1.000) | 0.6000 (3/5) | 0.588 [0.470, 0.697] |
| 0.60 | 8/12 = 0.667 (ret 0.917) | 4/12 = 0.333 (ret 0.750) | 5/12 = 0.417 (ret 0.833) | 7/10 = 0.700 (ret 0.800) | 6/12 = 0.500 (ret 0.667) | 9/10 = 0.900 (ret 1.000) | 0.5861 (211/360) | 0.574 [0.455, 0.684] |
| 0.65 | 7/12 = 0.583 (ret 0.833) | 4/12 = 0.333 (ret 0.750) | 4/12 = 0.333 (ret 0.750) | 6/10 = 0.600 (ret 0.800) | 5/12 = 0.417 (ret 0.667) | 8/10 = 0.800 (ret 1.000) | 0.5111 (23/45) | 0.500 [0.384, 0.616] |
| 0.70 | 7/12 = 0.583 (ret 0.833) | 4/12 = 0.333 (ret 0.750) | 4/12 = 0.333 (ret 0.750) | 5/10 = 0.500 (ret 0.700) | 5/12 = 0.417 (ret 0.667) | 8/10 = 0.800 (ret 1.000) | 0.4944 (89/180) | 0.485 [0.371, 0.602] |
| 0.75 | 6/12 = 0.500 (ret 0.833) | 3/12 = 0.250 (ret 0.667) | 3/12 = 0.250 (ret 0.667) | 5/10 = 0.500 (ret 0.700) | 5/12 = 0.417 (ret 0.667) | 8/10 = 0.800 (ret 1.000) | 0.4528 (163/360) | 0.441 [0.329, 0.559] |
| 0.80 | 6/12 = 0.500 (ret 0.833) | 3/12 = 0.250 (ret 0.667) | 3/12 = 0.250 (ret 0.667) | 4/10 = 0.400 (ret 0.700) | 5/12 = 0.417 (ret 0.667) | 7/10 = 0.700 (ret 0.900) | 0.4194 (151/360) | 0.412 [0.303, 0.530] |
| 0.85 | 6/12 = 0.500 (ret 0.833) | 2/12 = 0.167 (ret 0.583) | 3/12 = 0.250 (ret 0.667) | 4/10 = 0.400 (ret 0.700) | 3/12 = 0.250 (ret 0.583) | 6/10 = 0.600 (ret 0.900) | 0.3611 (13/36) | 0.353 [0.250, 0.472] |
| 0.90 | 6/12 = 0.500 (ret 0.833) | 2/12 = 0.167 (ret 0.583) | 3/12 = 0.250 (ret 0.667) | 3/10 = 0.300 (ret 0.600) | 3/12 = 0.250 (ret 0.583) | 6/10 = 0.600 (ret 0.900) | 0.3444 (31/90) | 0.338 [0.237, 0.457] |

### 3.3 Joint leave-one-transition-out `t*` (descriptive; not out-of-sample)
For each held-out transition `h` (held out in every repository) the thresholds are classified jointly on the remaining transitions by C1–C3; `t* = max(pass)`; held-out survival of `h` is reported with its matched/prior areas, labelled and `different` counts. **Caveat (§3.7):** `t*` sits at the survival edge by construction, so this view understates the selection optimism and is the cheapest check only; it is not an out-of-sample estimate.

**Arm A (codex)**

| held-out | joint t* | axum: held-out survival (matched/prior; labelled; different) | fastapi: held-out survival (matched/prior; labelled; different) |
|---|---|---|---|
| T1 | 0.30 | 0.7254901960784313 (37/51; 37; 0) | 0.8222222222222222 (37/45; 37; 0) |
| T2 | none | n/a (no t*) | n/a (no t*) |
| T3 | none | n/a (no t*) | n/a (no t*) |
| T4 | none | n/a (no t*) | n/a (no t*) |
| T5 | none | n/a (no t*) | n/a (no t*) |
| T6 | none | n/a (no t*) | n/a (no t*) |

**Arm B (claude)** (verdict `fallback`; shown for completeness, no `t*` is adopted)

| held-out | joint t* | axum: held-out survival (matched/prior; labelled; different) | fastapi: held-out survival (matched/prior; labelled; different) |
|---|---|---|---|
| T1 | none | n/a (no t*) | n/a (no t*) |
| T2 | none | n/a (no t*) | n/a (no t*) |
| T3 | none | n/a (no t*) | n/a (no t*) |
| T4 | none | n/a (no t*) | n/a (no t*) |
| T5 | 0.30 | 0.28 (7/25; 7; 0) | 0.75 (9/12; 9; 0) |
| T6 | none | n/a (no t*) | n/a (no t*) |

## 4. Decision details

### 4.1 Gates, arm A and arm B (from `results.json`, `arms.<arm>.decision`)
| gate | arm A (codex) | arm B (claude) |
|---|---|---|
| class gate (per repo: remap ≥ 2, mutation ≥ 1 measured) | axum: remap 4, mutation 2; fastapi: remap 4, mutation 2; ok | identical counts; ok |
| label gate (unlabelled share at `T_MIN` ≤ 1/10) | axum 0.0, fastapi 0.0; ok | axum 0.0, fastapi 0.0; ok |
| control gate (pooled across arms) | ok (reason null) | ok (reason null) |
| feasible thresholds (`pass`) | 0.30 | none |
| `imprecise` thresholds | none | none |
| verdict | **adopt**, `t* = 0.30` | **fallback** |
| `tripped_gate` | none | none |
| pass set contiguous | true | true (empty) |
| survival margin at `t*` | 0.00354238984269944 (4531/1279080) | none |
| citable | true | true |

Per-threshold status, arm A: 0.30 `pass`; 0.35 … 0.90 `fail`. Arm B: all 13 `fail`.
Arm A `AT t*=0.30` (§7, prominent): axum labelled 228 of 228 matched, unlabelled 0 (share 0.0), worst-case spurious 0.0; fastapi labelled 208 of 208, unlabelled 0 (share 0.0), worst-case spurious 0.0.

### 4.2 Replication and ledger mapping
`replication` (from `results.json`): gating false; arm A threshold 0.30; arm B verdict `fallback`; arm B threshold none; arm B gates ok true; arm B status at the arm-A threshold `fail`; arm A citable true; arm B citable true; **status `not replicated`**.
Applying the ordered rule of §4: (i) arm A adopted, so not `not applicable`; (ii) arm A is probe-consistent; (iii) arm B's `t*` is not `pass`; (iv) `replicated` requires `t*` to be `pass` in arm B, so `not replicated`. Arm B's own verdict is `fallback` and its own `t*` is none.
Ledger state fixed by §4: arm A `adopt` + probe-consistent → **HD5 `supported`, scoped to the codex provider** (arm B is not `replicated`); arm B never moves HD5. Not edited here (ledgers are updated by the main agent).

### 4.3 Operating characteristics (§7; from `analysis_output.txt`)
One threshold, conditional on C1, independent matches (anticonservative; selection inflation on real maps stays unmeasured). P(adopt at a passing threshold, both repos) at a true spurious rate of 3 % / 4 % / 5 % / 7 % / 10 %:
n = 150 per repo (adoption needs ≤ 7 spurious): 0.8401 / 0.5578 / 0.2733 / 0.0285 / 0.0002;
n = 216 (≤ 10): 0.8784 / 0.5639 / 0.2318 / 0.0109 / 0.0;
n = 270 (≤ 13): 0.9314 / 0.6456 / 0.2671 / 0.0088 / 0.0;
C3 stops binding at n = 140. Observed labelled occurrences at 0.30: axum 228, fastapi 208 (inside the 216–270 expected band for axum, below it for fastapi). [INFERENCE] These figures describe the pipeline's power given a true rate; they do not say what this labeller's real sensitivity was. Per §7, `adopt` means point spurious ≤ 5 % and Wilson upper ≤ 10 % in both repos *under the label protocol*; it does not demonstrate a true rate ≤ 5 %.

## 5. Coverage diagnostics (descriptive; every replay row)
Per transition: counts of prior areas, next agent features, distinct owned files in each map and in both, prior areas with zero overlap with every next agent feature (a survival ceiling no threshold can lift). Not used in any decision.

**Arm A (codex), axum**

| T | class/sub | prior areas | next agent features | prev owned files | next owned files | owned files in both | prior areas with zero overlap |
|---|---|---|---|---|---|---|---|
| T1 | remap/fresh | 51 | 37 | 173 | 170 | 170 | 0 |
| T2 | remap/fresh | 51 | 42 | 173 | 170 | 170 | 0 |
| T3 | remap/same_state | 51 | 51 | 173 | 221 | 170 | 0 |
| T4 | remap/same_state | 51 | 38 | 221 | 167 | 167 | 0 |
| T5 | mutation/mutation | 38 | 44 | 167 | 179 | 166 | 0 |
| T6 | mutation/mutation | 44 | 38 | 179 | 178 | 173 | 0 |

**Arm A (codex), fastapi**

| T | class/sub | prior areas | next agent features | prev owned files | next owned files | owned files in both | prior areas with zero overlap |
|---|---|---|---|---|---|---|---|
| T1 | remap/fresh | 45 | 38 | 436 | 370 | 369 | 5 |
| T2 | remap/fresh | 45 | 46 | 436 | 625 | 436 | 0 |
| T3 | remap/same_state | 45 | 41 | 436 | 628 | 435 | 0 |
| T4 | remap/same_state | 41 | 54 | 628 | 612 | 612 | 2 |
| T5 | mutation/mutation | 54 | 32 | 612 | 337 | 335 | 14 |
| T6 | mutation/mutation | 32 | 52 | 337 | 636 | 334 | 0 |

**Arm B (claude), axum**

| T | class/sub | prior areas | next agent features | prev owned files | next owned files | owned files in both | prior areas with zero overlap |
|---|---|---|---|---|---|---|---|
| T1 | remap/fresh | 23 | 19 | 161 | 157 | 157 | 0 |
| T2 | remap/fresh | 23 | 25 | 161 | 161 | 161 | 0 |
| T3 | remap/same_state | 23 | 17 | 161 | 156 | 156 | 0 |
| T4 | remap/same_state | 17 | 25 | 156 | 222 | 156 | 0 |
| T5 | mutation/mutation | 25 | 10 | 222 | 172 | 167 | 0 |
| T6 | mutation/mutation | 10 | 21 | 172 | 163 | 160 | 0 |

**Arm B (claude), fastapi**

| T | class/sub | prior areas | next agent features | prev owned files | next owned files | owned files in both | prior areas with zero overlap |
|---|---|---|---|---|---|---|---|
| T1 | remap/fresh | 12 | 12 | 134 | 102 | 102 | 0 |
| T2 | remap/fresh | 12 | 9 | 134 | 65 | 65 | 1 |
| T3 | remap/same_state | 12 | 10 | 134 | 100 | 99 | 0 |
| T4 | remap/same_state | 10 | 12 | 100 | 98 | 95 | 0 |
| T5 | mutation/mutation | 12 | 10 | 98 | 81 | 75 | 1 |
| T6 | mutation/mutation | 10 | 12 | 81 | 119 | 79 | 0 |

## 6. Chained survival (label-free; reported, not gating)
Fold F1→R1→R2→M1→M2 (T3–T6) with accumulating areas, at `t*` for arm A (0.30) and at the design default 0.50 for arm B (nothing adopted); `F1 areas matched at every step so far` is the share of the original F1 areas matched at every step so far. The one-step and chained directions are not ordered a priori (§10).

**Arm A (codex), axum**

Threshold 0.30 (adopted t*):

| step | status | active prior | matched | created | kept | retired | F1 areas matched at every step so far / F1 areas |
|---|---|---|---|---|---|---|---|
| T3 | ok | 51 | 45 | 6 | 2 | 4 | 45/51 |
| T4 | ok | 53 | 33 | 5 | 18 | 2 | 33/51 |
| T5 | ok | 56 | 38 | 6 | 16 | 2 | 31/51 |
| T6 | ok | 60 | 36 | 2 | 17 | 7 | 28/51 |

**Arm A (codex), fastapi**

Threshold 0.30 (adopted t*):

| step | status | active prior | matched | created | kept | retired | F1 areas matched at every step so far / F1 areas |
|---|---|---|---|---|---|---|---|
| T3 | ok | 45 | 34 | 7 | 6 | 5 | 34/45 |
| T4 | ok | 47 | 44 | 10 | 3 | 0 | 34/45 |
| T5 | ok | 57 | 31 | 1 | 22 | 4 | 25/45 |
| T6 | ok | 54 | 47 | 5 | 7 | 0 | 24/45 |

**Arm B (claude), axum**

Threshold 0.50 (design default 0.5; nothing adopted):

| step | status | active prior | matched | created | kept | retired | F1 areas matched at every step so far / F1 areas |
|---|---|---|---|---|---|---|---|
| T3 | ok | 23 | 15 | 2 | 1 | 7 | 15/23 |
| T4 | ok | 18 | 13 | 12 | 3 | 2 | 12/23 |
| T5 | ok | 28 | 6 | 4 | 11 | 11 | 4/23 |
| T6 | ok | 21 | 11 | 10 | 10 | 0 | 4/23 |

**Arm B (claude), fastapi**

Threshold 0.50 (design default 0.5; nothing adopted):

| step | status | active prior | matched | created | kept | retired | F1 areas matched at every step so far / F1 areas |
|---|---|---|---|---|---|---|---|
| T3 | ok | 12 | 8 | 2 | 3 | 1 | 8/12 |
| T4 | ok | 13 | 7 | 5 | 4 | 2 | 6/12 |
| T5 | ok | 16 | 8 | 2 | 6 | 2 | 5/12 |
| T6 | ok | 16 | 11 | 1 | 5 | 0 | 5/12 |

## 7. Labels

### 7.1 Agreement, adjudication and served models (real pairs, round 1)
- Unique real pairs labelled: **578** (pooled across arms); controls **71**; call rows 649; `labelled 649 pairs` (`logs/x2-label.log`).
- A/B raw agreement **0.9653979238754326**; Cohen's κ **0.9235944956311386** (578 pairs, both parsed 578, unparsable final 0).
- Adjudicated pairs (C sessions): **20**, adjudication rate **0.03460207612456748** (20/578); C answered validly in 20 of 20.
- **C-equals-A model share: 0.0 (0 of 20)**: C's resolved model (`claude-sonnet-5-5`) never equalled A's (`claude-opus-5-5`) on the adjudicated pairs.
- Codex sessions (sum of B's attempts over the 578 real round-1 pairs): **578**; pairs in which any codex attempt printed an `exec` event: **0** (real pairs, round 1).
- Served models per role (final attempt's model per call row, all 649 call rows including controls): **A `claude-opus-5-5` (649), B `gpt-6.1-sol` (649), C `claude-sonnet-5-5` (20)**. A model seen only on a failed first attempt is not recorded (§6).
- Control pairs, A/B: raw agreement 1.0, κ 1.0 (71 pairs, adjudicated 0).
- [INFERENCE] The labeller families (claude-opus-5-5, gpt-6.1-sol) are the same families as the two mapper arms; shared blind spots are possible and no human checked any label (§6, §10).

### 7.2 Known-answer controls (pooled across arms; gate ok)
578 unique real pairs, 71 control pairs, shortfalls 0.
- **Swapped** (expected `different_concern`): n = 47 (wanted 47, shortfall 0); usable 47 (minimum 24); misses 2 (rate 0.043; gate limit 20 %); cannot_tell 0, unparsable 0, never labelled 0; labels {"different_concern": 45, "overlapping_concern": 2}.
- **Identical** (expected `same_concern`): n = 24 (wanted 24, shortfall 0); usable 24 (minimum 12); misses 0 (rate 0.000); cannot_tell 0, unparsable 0, never labelled 0; labels {"same_concern": 24}.
- Per arm (descriptive): `{"claude": {"identical": {"miss_rate": 0.0, "misses": 0, "n": 6, "usable": 6}, "swapped": {"miss_rate": 0.0, "misses": 0, "n": 8, "usable": 8}}, "codex": {"identical": {"miss_rate": 0.0, "misses": 0, "n": 18, "usable": 18}, "swapped": {"miss_rate": 0.05128205128205128, "misses": 2, "n": 39, "usable": 39}}}`.
- Swapped controls by source-pair Jaccard band (misses/usable, band lower edge in %): 30: 0/1; 35: 0/1; 40: 0/4; 45: 0/1; 50: 0/3; 55: 1/3; 60: 0/3; 65: 0/1; 70: 0/1; 75: 0/2; 80: 0/2; 85: 0/3; 90: 1/22. The two misses are in the 55 and 90 bands.
- The controls test inattention and gross anchoring at the source pair's overlap, **not** the 0.30–0.60 regime where `t*` is chosen, and not concern judgement itself (§5, §10).

### 7.3 Label integrity
`labelled_pairs_without_call_row`: 0. No pair is unlabelled, unparsable or never labelled (§9).

### 7.4 Label distribution by class and threshold band; over-call check
Counts of final labels on real matches at `T_MIN` (matches at 0.30 and above), by Jaccard band lower edge, per arm and repo. There are no `different_concern` and no `cannot_tell` labels among real pairs.

**Arm A (codex), axum**

| band (Jaccard ≥) | same | overlapping | different |
|---|---|---|---|
| 0.30 | 1 | 6 | 0 |
| 0.35 | 0 | 4 | 0 |
| 0.40 | 1 | 11 | 0 |
| 0.45 | 0 | 1 | 0 |
| 0.50 | 11 | 20 | 0 |
| 0.55 | 1 | 12 | 0 |
| 0.60 | 1 | 9 | 0 |
| 0.65 | 9 | 4 | 0 |
| 0.70 | 1 | 2 | 0 |
| 0.75 | 10 | 4 | 0 |
| 0.80 | 7 | 1 | 0 |
| 0.85 | 8 | 0 | 0 |
| 0.90 | 104 | 0 | 0 |

By class: mutation: overlapping_concern=21, same_concern=50; remap: overlapping_concern=53, same_concern=104. Over-call check, band [0.90, 1.0]: 0 `different` of 104 labelled (104 band matches), share 0.0.

**Arm A (codex), fastapi**

| band (Jaccard ≥) | same | overlapping | different |
|---|---|---|---|
| 0.30 | 0 | 3 | 0 |
| 0.35 | 1 | 1 | 0 |
| 0.40 | 2 | 3 | 0 |
| 0.45 | 0 | 1 | 0 |
| 0.50 | 0 | 5 | 0 |
| 0.55 | 1 | 9 | 0 |
| 0.60 | 5 | 12 | 0 |
| 0.65 | 6 | 3 | 0 |
| 0.70 | 2 | 7 | 0 |
| 0.75 | 2 | 5 | 0 |
| 0.80 | 9 | 14 | 0 |
| 0.85 | 6 | 3 | 0 |
| 0.90 | 107 | 1 | 0 |

By class: mutation: overlapping_concern=15, same_concern=47; remap: overlapping_concern=52, same_concern=94. Over-call check, band [0.90, 1.0]: 0 `different` of 108 labelled (108 band matches), share 0.0.

**Arm B (claude), axum**

| band (Jaccard ≥) | same | overlapping | different |
|---|---|---|---|
| 0.30 | 0 | 9 | 0 |
| 0.40 | 0 | 3 | 0 |
| 0.45 | 0 | 1 | 0 |
| 0.50 | 3 | 12 | 0 |
| 0.55 | 2 | 3 | 0 |
| 0.60 | 6 | 6 | 0 |
| 0.65 | 1 | 1 | 0 |
| 0.70 | 1 | 3 | 0 |
| 0.75 | 1 | 2 | 0 |
| 0.80 | 1 | 2 | 0 |
| 0.85 | 1 | 0 | 0 |
| 0.90 | 30 | 0 | 0 |

By class: mutation: overlapping_concern=11, same_concern=6; remap: overlapping_concern=31, same_concern=40. Over-call check, band [0.90, 1.0]: 0 `different` of 30 labelled (30 band matches), share 0.0.

**Arm B (claude), fastapi**

| band (Jaccard ≥) | same | overlapping | different |
|---|---|---|---|
| 0.30 | 0 | 2 | 0 |
| 0.35 | 1 | 0 | 0 |
| 0.40 | 2 | 2 | 0 |
| 0.45 | 0 | 2 | 0 |
| 0.50 | 2 | 3 | 0 |
| 0.55 | 0 | 1 | 0 |
| 0.60 | 4 | 1 | 0 |
| 0.65 | 0 | 1 | 0 |
| 0.70 | 2 | 1 | 0 |
| 0.75 | 2 | 0 | 0 |
| 0.80 | 4 | 0 | 0 |
| 0.85 | 1 | 0 | 0 |
| 0.90 | 23 | 0 | 0 |

By class: mutation: overlapping_concern=2, same_concern=16; remap: overlapping_concern=11, same_concern=25. Over-call check, band [0.90, 1.0]: 0 `different` of 23 labelled (23 band matches), share 0.0.

Over-call check (§5, descriptive): the share of `different_concern` among labelled matches in the [0.90, 1.0] band is 0 in all four (arm, repo) cells (counts above). It cannot separate an under-calling labeller from a genuinely concern-identical near-identical file set, and it never changes the verdict.
Gradient (Jaccard vs `different`, exploratory): arm A n = 436, r = None; arm B n = 142, r = None; undefined because no `different` label exists.
[INFERENCE] The absence of any `different` label among 578 real pairs (while 45 of 47 swapped controls are called `different`) means the spurious-merge criterion C2 never binds in this data set; the decision is carried entirely by survival (C1). Whether this reflects a mapper that never merges across concerns at overlap ≥ 0.30 or a labeller that calls partial overlaps `overlapping_concern` rather than `different_concern` cannot be separated here; the strict sensitivity is reported in §1 for that reason.

## 8. Map inventory, attempts and wall times
28 of 28 maps recorded `ok`; retried, failed, skipped, unrecorded or workspace-dirty maps: none; stop records (`data/stopped-*.json`): none; aborts: none; mutation skipped: none (M2 ran in all four collections); every map succeeded on its first attempt.

| arm | repo | map | rc | attempts | first attempt started (UTC) | seconds total | status | skipped | workspace dirty |
|---|---|---|---|---|---|---|---|---|---|
| codex | fastapi | F1 | 0 | 1 | 2026-10-23T18:42:14Z | 236.2 | ok | None | False |
| codex | fastapi | F2 | 0 | 1 | 2026-10-23T18:46:11Z | 212.5 | ok | None | False |
| codex | fastapi | F3 | 0 | 1 | 2026-10-23T18:49:44Z | 294.0 | ok | None | False |
| codex | fastapi | R1 | 0 | 1 | 2026-10-23T18:54:38Z | 290.0 | ok | None | False |
| codex | fastapi | R2 | 0 | 1 | 2026-10-23T18:59:29Z | 280.3 | ok | None | False |
| codex | fastapi | M1 | 0 | 1 | 2026-10-23T19:04:10Z | 168.6 | ok | None | False |
| codex | fastapi | M2 | 0 | 1 | 2026-10-23T19:06:59Z | 300.2 | ok | None | False |
| claude | fastapi | F1 | 0 | 1 | 2026-10-23T19:12:10Z | 88.3 | ok | None | False |
| claude | fastapi | F2 | 0 | 1 | 2026-10-23T19:13:39Z | 86.4 | ok | None | False |
| claude | fastapi | F3 | 0 | 1 | 2026-10-23T19:15:05Z | 50.1 | ok | None | False |
| claude | fastapi | R1 | 0 | 1 | 2026-10-23T19:15:56Z | 72.6 | ok | None | False |
| claude | fastapi | R2 | 0 | 1 | 2026-10-23T19:17:09Z | 81.6 | ok | None | False |
| claude | fastapi | M1 | 0 | 1 | 2026-10-23T19:18:31Z | 64.1 | ok | None | False |
| claude | fastapi | M2 | 0 | 1 | 2026-10-23T19:19:36Z | 92.1 | ok | None | False |
| claude | axum | F1 | 0 | 1 | 2026-10-23T19:21:16Z | 103.4 | ok | None | False |
| claude | axum | F2 | 0 | 1 | 2026-10-23T19:22:59Z | 183.5 | ok | None | False |
| claude | axum | F3 | 0 | 1 | 2026-10-23T19:26:03Z | 100.6 | ok | None | False |
| claude | axum | R1 | 0 | 1 | 2026-10-23T19:27:44Z | 156.5 | ok | None | False |
| claude | axum | R2 | 0 | 1 | 2026-10-23T19:30:20Z | 244.1 | ok | None | False |
| claude | axum | M1 | 0 | 1 | 2026-10-23T19:34:25Z | 72.8 | ok | None | False |
| claude | axum | M2 | 0 | 1 | 2026-10-23T19:35:38Z | 100.0 | ok | None | False |
| codex | axum | F1 | 0 | 1 | 2026-10-23T19:37:27Z | 184.7 | ok | None | False |
| codex | axum | F2 | 0 | 1 | 2026-10-23T19:40:32Z | 152.1 | ok | None | False |
| codex | axum | F3 | 0 | 1 | 2026-10-23T19:43:04Z | 135.6 | ok | None | False |
| codex | axum | R1 | 0 | 1 | 2026-10-23T19:45:20Z | 182.9 | ok | None | False |
| codex | axum | R2 | 0 | 1 | 2026-10-23T19:48:24Z | 150.2 | ok | None | False |
| codex | axum | M1 | 0 | 1 | 2026-10-23T19:50:54Z | 129.7 | ok | None | False |
| codex | axum | M2 | 0 | 1 | 2026-10-23T19:53:06Z | 151.1 | ok | None | False |

Collection windows and probes (UTC); the sum of the 28 maps' `seconds_total` is 4364.2 s:

| arm/repo | before-probe (UTC, model) | after-probe (UTC, model) | sum of map seconds | maps |
|---|---|---|---|---|
| codex/fastapi | 2026-10-23T18:42:11Z (gpt-6.1-sol) | 2026-10-23T19:12:04Z (gpt-6.1-sol) | 1781.8 | 7 |
| claude/fastapi | 2026-10-23T19:12:08Z (claude-opus-5-5) | 2026-10-23T19:21:11Z (claude-opus-5-5) | 535.2 | 7 |
| claude/axum | 2026-10-23T19:21:15Z (claude-opus-5-5) | 2026-10-23T19:37:21Z (claude-opus-5-5) | 960.9 | 7 |
| codex/axum | 2026-10-23T19:37:26Z (gpt-6.1-sol) | 2026-10-23T19:55:42Z (gpt-6.1-sol) | 1086.3 | 7 |

## 9. Unmeasured counts and reasons
- Transitions lost to failed, stopped or skipped maps: **0** in every (arm, repo) (`unmeasured_transitions: []` in every table).
- Maps retried or failed: **0**; mutation skipped: **0**; stops: **0**; aborts: **0**.
- Areas with no prior areas: none (every transition has prior areas, §5).
- Matches without a usable label, at every threshold and stratum, both arms: `cannot_tell` **0**, unparsable **0**, never labelled **0** (every unlabelled column in Appendix A is `0 (0/0/0)`). Unlabelled share at `T_MIN` per repo, both arms: 0.0.
- Controls: cannot_tell 0, unparsable 0, never labelled 0; shortfall 0.

## 10. Tool versions, probes, provenance

### 10.1 Tool versions (`data/tool-versions-<arm>-<repo>.json`)
| arm/repo | started | claude CLI before → after | codex CLI before → after | clawpatch (node) | clawpatch_cli_sha256 | build safety | areas.go / records.go sha256 (12) | presence CODEX_HOME / CLAWPATCH_CLAUDE_BIN / ANTHROPIC_MODEL | codex/claude provider timeout, per-call, wall, grace | workspace config presence |
|---|---|---|---|---|---|---|---|---|---|---|
| codex/fastapi | 2026-10-23T18:42:14Z | 2.1.309 (Claude Code) → 2.1.309 (Claude Code) | codex-cli 0.162.1 → codex-cli 0.162.1 | 0.7.3 (v22.18.0) | `4eab263b87d6b0d0af1eb6792e1e38827e0b13ab53078e62f5ffed76cc4320eb` | ok=True, marker_ok=True, sanitizer_in_dist=True, marker `scout-review-sanitize-v3 dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3` | 8ed2e70815fd / 57344e025d5e | True / False / False | 900000 / 900000 ms, 1800 s, 28800 s, 10 s | none present (all 5 false) |
| codex/axum | 2026-10-23T19:37:27Z | 2.1.309 (Claude Code) → 2.1.309 (Claude Code) | codex-cli 0.162.1 → codex-cli 0.162.1 | 0.7.3 (v22.18.0) | `4eab263b87d6b0d0af1eb6792e1e38827e0b13ab53078e62f5ffed76cc4320eb` | ok=True, marker_ok=True, sanitizer_in_dist=True, marker `scout-review-sanitize-v3 dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3` | 8ed2e70815fd / 57344e025d5e | True / False / False | 900000 / 900000 ms, 1800 s, 28800 s, 10 s | none present (all 5 false) |
| claude/fastapi | 2026-10-23T19:12:10Z | 2.1.309 (Claude Code) → 2.1.309 (Claude Code) | codex-cli 0.162.1 → codex-cli 0.162.1 | 0.7.3 (v22.18.0) | `4eab263b87d6b0d0af1eb6792e1e38827e0b13ab53078e62f5ffed76cc4320eb` | ok=True, marker_ok=True, sanitizer_in_dist=True, marker `scout-review-sanitize-v3 dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3` | 8ed2e70815fd / 57344e025d5e | True / False / False | 900000 / 900000 ms, 1800 s, 28800 s, 10 s | none present (all 5 false) |
| claude/axum | 2026-10-23T19:21:16Z | 2.1.309 (Claude Code) → 2.1.309 (Claude Code) | codex-cli 0.162.1 → codex-cli 0.162.1 | 0.7.3 (v22.18.0) | `4eab263b87d6b0d0af1eb6792e1e38827e0b13ab53078e62f5ffed76cc4320eb` | ok=True, marker_ok=True, sanitizer_in_dist=True, marker `scout-review-sanitize-v3 dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3` | 8ed2e70815fd / 57344e025d5e | True / False / False | 900000 / 900000 ms, 1800 s, 28800 s, 10 s | none present (all 5 false) |

CLI versions after the after-probe equal those before in all four files (`cli_versions_changed: {}` for both arms). Names (not values) of the environment passed to Clawpatch are in the versions files; presence flags are recorded without values.

### 10.2 Model probes and probe consistency (§2)
- **Arm codex: CONSISTENT**; models `["gpt-6.1-sol"]`; CLI versions changed between the probes: `{}`.
  - fastapi: before 2026-10-23T18:42:11Z `gpt-6.1-sol` (5.7 s), after 2026-10-23T19:12:04Z `gpt-6.1-sol` (4.2 s); CLI before and after `codex-cli 0.162.1`.
  - axum: before 2026-10-23T19:37:26Z `gpt-6.1-sol` (4.9 s), after 2026-10-23T19:55:42Z `gpt-6.1-sol` (4.5 s); CLI before and after `codex-cli 0.162.1`.
  - Workspace config presence (names only): all five false in both repos (`.claude/settings.json`, `.claude/settings.local.json`, `.clawpatch/config.json`, `.codex`, `clawpatch.config.json`).
- **Arm claude: CONSISTENT**; models `["claude-opus-5-5"]`; CLI versions changed between the probes: `{}`.
  - fastapi: before 2026-10-23T19:12:08Z `claude-opus-5-5` (2.6 s), after 2026-10-23T19:21:11Z `claude-opus-5-5` (2.9 s); CLI `2.1.309 (Claude Code)` before and after.
  - axum: before 2026-10-23T19:21:15Z `claude-opus-5-5` (3.1 s), after 2026-10-23T19:37:21Z `claude-opus-5-5` (2.4 s); CLI `2.1.309 (Claude Code)` before and after.
  - Workspace config presence: all five false in both repos.
- Reasons for inconsistency: none (`reasons` empty for both arms).
- [INFERENCE] The served model behind each map is taken to be the one named by the bracketing probes (Clawpatch reports only its option value `default`, recorded per map as `model_option`). A transient roll reverted by the after-probe, or a weight change under the same model name, would not be seen (§2, §10).

### 10.3 Replay provenance (`data/replay-meta.json`)
Scout HEAD `3d78ab7f423e6ee813235bda6fbf0d3e5bb2ef1c`; `internal/areas/areas.go` sha256 `8ed2e70815fdcc605d60402fb603cfb905d58457a5869b3795d59507a93cb484`; `internal/clawpatch/records.go` sha256 `57344e025d5e2d0b7a10abf543b20b55f899602c1dd95d2b6a2b62d9b823dfbe`; working tree dirty for those two files: false; Go `go1.25.6`; grid `[30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90]` percent; transitions file sha256 `7356ab0d4121785d5e2ac2c050d66f17a5368812415cebdd3562c90e5ced6ac7`; maps directory `evaluation/x2-area-stability/data/maps`. The two source digests equal those recorded in the round-10 sign-off. Scout HEAD is later than the preregistration commit (`01ed665`; HEAD `3d78ab7` is the design rev 2.4 adoption commit); neither file under test changed (equal digests).

### 10.4 Digests (SHA-256)
Frozen files and code under test, recomputed on 2026-10-23 at the time of writing; each equals the round-10 sign-off digest (`reviews/round-10.md` §10):
```
ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355  PREREGISTRATION.md (rev 10)
9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799  GUIDE.md
e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3  common.py
89948f3c6d1b308b91a5ac5a8060088269270806546fe5c577dea307a4dfc83d  collect.py
f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e  replay/main.go
face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a  build_pairs.py
87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f  label.py
7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63  analysis.py
5157b15be6013c350b1dde006b0ee4833766d46981b131df756b1f838441c41b  run.sh
29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784  validation/validate.py
0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226  validation/stub_clawpatch.js
e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779  validation/credcheck/main.go
13e04e61720fa0542d391a48f2d63db4e1973f29096114ba9120bddef84c4048  README.md (not frozen)
dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3  scripts/clawpatch-safe.patch
99e3f6b3c670cd8c297b4ed498eafc735b7b5045aea8b97c7b55dc6343c21653  scripts/build-clawpatch.sh
```
Recorded `clawpatch_cli_sha256` (same in all four collections): `4eab263b87d6b0d0af1eb6792e1e38827e0b13ab53078e62f5ffed76cc4320eb`. The rev-8 README note in `data-rev8-unpatched/` quotes the prefix `4eab263b…` for the unpatched build, i.e. this digest does not distinguish the patched from the unpatched build [INFERENCE: it appears to cover a file the patch does not change]; build safety rests on the recorded `clawpatch_build_safety` check (marker `scout-review-sanitize-v3 dc867503…`, sanitizer present in `dist/`), which is `ok` in all four files. The `validate` run itself was not repeated for this document; its 233 checks and 0 FAIL are those recorded in `reviews/round-10.md`.

## 11. Deviations
- **D1. Amendment 1 (rev 9, revised to rev 10; signed off in `reviews/round-10.md`, zero issues).** The rev-8 collection started 2026-10-23T17:24Z on a local Clawpatch build that lacked the review-safety patch and (claude/fastapi) failed at the before-probe because the host auth context lacked `USER`/`LOGNAME`. Collection was restarted on the review-safety build (`scout-review-sanitize-v3`, patch sha256 `dc867503…6ce3`; build safety `ok` in all four versions files). All four (arm, repo) collections re-ran in the preregistered ABBA order into a fresh `data/`: codex/fastapi 18:42Z, claude/fastapi 19:12Z, claude/axum 19:21Z, codex/axum 19:37Z. Nothing else in §2–§10 changed.
- **D2. Excluded data.** `data-rev8-unpatched/` (7 codex/fastapi maps, all rc 0, plus `x2-collect.log` of that run and its versions file; unpatched build per its README note) is **excluded from every analysis** and was not read by any stage; no predictor–outcome quantity was computed from it. The directory is preserved untouched.
- **D3. Later collection window / arm–time confound.** The restart took place about 1 h 18 min after the first attempt (17:24Z vs 18:42Z). The four arm–repo collections are sequential (ABBA), so no arm was always later, but model or load change over the window is not separable from arm or repository (§10). [INFERENCE] The agents may differ between the first and the restarted window; the probes bracket each restarted collection and show one served model per arm throughout.
- **D4. Scout commit.** Replay ran at Scout HEAD `3d78ab7` (design rev 2.4 adoption), later than the preregistration commit `01ed665`; `internal/areas/areas.go` and `internal/clawpatch/records.go`, the code under test, have the signed-off digests. Not a preregistration deviation; recorded for provenance.
- **D5. Log location.** Neither `README.md` nor the preregistration names a location for `x2-collect.log` / `x2-label.log`, so copies were placed in `logs/` (the originals remain at the repository root; `data/` was left untouched).
- **Runtime deviations: none.** No CLI outage, repository unavailability, map failure or retry, skipped mutation, stop, abort, or label-session failure occurred (inventory in §8; every label unparsable-final count 0, §7.1). The §11 rule against peeking was followed: only map success/failure and counts were inspected before the full pipeline had run [INFERENCE: relies on the operator's account, not on a recorded check].

## 12. Reproduce
From the repository root, toolchain per `README.md` (`git submodule update --init && scripts/bootstrap-tools.sh && source .tools/activate`):

    source .tools/activate && evaluation/x2-area-stability/run.sh validate      # no model calls, ~3 min
    PY=.tools/venv/bin/python; C=evaluation/x2-area-stability/collect.py
    nohup sh -c "$PY $C --arm codex  --repo fastapi --source $SOURCE_FASTAPI &&
                 $PY $C --arm claude --repo fastapi --source $SOURCE_FASTAPI &&
                 $PY $C --arm claude --repo axum    --source $SOURCE_AXUM    &&
                 $PY $C --arm codex  --repo axum    --source $SOURCE_AXUM" > x2-collect.log 2>&1 &
    evaluation/x2-area-stability/run.sh replay
    evaluation/x2-area-stability/run.sh pairs
    evaluation/x2-area-stability/run.sh label        # both arms in one pass; needs logged-in claude and codex CLIs
    evaluation/x2-area-stability/run.sh analysis     # data/results.json, data/analysis_output.txt

`SOURCE_*` = any clone or URL containing the pinned revisions (fastapi `d666ccb62216e45ca78643b52c235ba0d2c53986`, axum `562fa206f09514e7503e898decf73024f5ce6fd4`). `collect.py` refuses to start over existing data, so a re-run needs a fresh `data/` and `rm -rf /tmp/scout-x2`. Collection and labelling are model-call runs: re-running draws new agent maps and labels. The replay, pairs and analysis stages are deterministic given `data/maps` and `data/pairs`: `run.sh replay`, `run.sh pairs` (refuses once `labels.jsonl` has rows) and `run.sh analysis` reproduce `data/replay.jsonl`, `data/chain.jsonl`, `data/pairs/*` (not labels) and `data/results.json`/`data/analysis_output.txt` from the recorded data.

## 13. Threats and caveats that bear on reading the headline (§10)
- **Two repositories, one run of each map**; run-to-run variance of the agent is inside the estimates. Shared prior maps (F1 in three transitions; R1, R2, M1 in two each) make the effective number of refreshes smaller than six. The decisive survival margin (0.00354238984269944) is smaller than the spread across transitions (axum per-transition survival at 0.30 ranges 0.647–0.947; fastapi 0.574–0.969, §3.2).
- **Selection on the same labels**: `t*` is chosen on the data that evaluates it; selection optimism is not removed (§7); the LOTO view understates it (§3.3).
- **Labels are model-made and unreviewed**; labellers see file overlap; controls do not test the 0.30–0.60 regime. All real pairs received no `different` label, so C2 and C3 pass vacuously at every threshold; the strict sensitivity (0.32456140350877194 axum, 0.32211538461538464 fastapi at 0.30) shows that a large share of matches at 0.30 are `overlapping_concern`, not `same_concern` (§7.4).
- **M2 is a rename**; mutation transitions enter the decisive mean as designed (§10).
- **Python and Rust only; Clawpatch 0.7.3 and default models only**; arm A is codex-scoped; arm B did not replicate.
- **One-step on the clean layout**; the chained picture (§6) is label-free and its direction relative to one-step survival is unknown (§4).

## Appendix A. `data/analysis_output.txt` (verbatim, complete)

```
replay: scout 3d78ab7f423e areas.go 8ed2e70815fd records.go 57344e025d5e go1.25.6 grid [30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90]
label agreement (real pairs, round 1): {"pairs": 578, "both_parsed": 578, "adjudicated": 20, "adjudication_rate": 0.03460207612456748, "adjudicated_c_ok": 20, "adjudicated_c_model_equals_a": 0, "adjudicated_c_model_equals_a_of": 20, "adjudicated_c_model_equals_a_share": 0.0, "codex_sessions": 578, "pairs_with_exec_event": 0, "unparsable_final": 0, "raw_agreement": 0.9653979238754326, "cohen_kappa": 0.9235944956311386}
label call rows by role, served model (final attempt's model per call row; every call row incl. controls and later rounds; a model seen only on a failed first attempt is not recorded): {"A": {"claude-opus-5-5": 649}, "B": {"gpt-6.1-sol": 649}, "C": {"claude-sonnet-5-5": 20}}
controls: pooled across arms; 578 unique real pairs; 71 pairs, gate ok
   swapped   n=47 (wanted 47, shortfall 0) usable=47 (min 24) misses(not different_concern)=2 rate=0.043 cannot_tell=0 unparsable=0 never=0 labels={"different_concern": 45, "overlapping_concern": 2}
   identical n=24 (wanted 24, shortfall 0) usable=24 (min 12) misses(not same_concern)=0 rate=0.000 cannot_tell=0 unparsable=0 never=0 labels={"same_concern": 24}
   per arm (descriptive): {"claude": {"identical": {"miss_rate": 0.0, "misses": 0, "n": 6, "usable": 6}, "swapped": {"miss_rate": 0.0, "misses": 0, "n": 8, "usable": 8}}, "codex": {"identical": {"miss_rate": 0.0, "misses": 0, "n": 18, "usable": 18}, "swapped": {"miss_rate": 0.05128205128205128, "misses": 2, "n": 39, "usable": 39}}}
   swapped controls by source-pair Jaccard band: {"30": {"misses": 0, "n": 1, "usable": 1}, "35": {"misses": 0, "n": 1, "usable": 1}, "40": {"misses": 0, "n": 4, "usable": 4}, "45": {"misses": 0, "n": 1, "usable": 1}, "50": {"misses": 0, "n": 3, "usable": 3}, "55": {"misses": 1, "n": 3, "usable": 3}, "60": {"misses": 0, "n": 3, "usable": 3}, "65": {"misses": 0, "n": 1, "usable": 1}, "70": {"misses": 0, "n": 1, "usable": 1}, "75": {"misses": 0, "n": 2, "usable": 2}, "80": {"misses": 0, "n": 2, "usable": 2}, "85": {"misses": 0, "n": 3, "usable": 3}, "90": {"misses": 1, "n": 22, "usable": 22}}
   label integrity: {"labelled_pairs_without_call_row": 0}
   control agreement: {"pairs": 71, "both_parsed": 71, "adjudicated": 0, "adjudication_rate": 0.0, "adjudicated_c_ok": 0, "adjudicated_c_model_equals_a": 0, "adjudicated_c_model_equals_a_of": 0, "adjudicated_c_model_equals_a_share": null, "codex_sessions": 71, "pairs_with_exec_event": 0, "unparsable_final": 0, "raw_agreement": 1.0, "cohen_kappa": 1.0}
map inventory: 28 recorded ok of 28; retried/failed/skipped/unrecorded/workspace dirty after the map: []
probe consistency (section 2, arm claude): CONSISTENT; models ["claude-opus-5-5"]; CLI versions changed between the probes: {}
   axum probes: before {"error": null, "model": "claude-opus-5-5", "seconds": 3.1, "timestamp": "2026-10-23T19:21:15Z"}, after {"error": null, "model": "claude-opus-5-5", "seconds": 2.4, "timestamp": "2026-10-23T19:37:21Z"}; CLI versions before {"claude": "2.1.309 (Claude Code)"}, after {"claude": "2.1.309 (Claude Code)"}
   axum workspace config presence (names only): {".claude/settings.json": false, ".claude/settings.local.json": false, ".clawpatch/config.json": false, ".codex": false, "clawpatch.config.json": false}
   fastapi probes: before {"error": null, "model": "claude-opus-5-5", "seconds": 2.6, "timestamp": "2026-10-23T19:12:08Z"}, after {"error": null, "model": "claude-opus-5-5", "seconds": 2.9, "timestamp": "2026-10-23T19:21:11Z"}; CLI versions before {"claude": "2.1.309 (Claude Code)"}, after {"claude": "2.1.309 (Claude Code)"}
   fastapi workspace config presence (names only): {".claude/settings.json": false, ".claude/settings.local.json": false, ".clawpatch/config.json": false, ".codex": false, "clawpatch.config.json": false}
probe consistency (section 2, arm codex): CONSISTENT; models ["gpt-6.1-sol"]; CLI versions changed between the probes: {}
   axum probes: before {"error": null, "model": "gpt-6.1-sol", "seconds": 4.9, "timestamp": "2026-10-23T19:37:26Z"}, after {"error": null, "model": "gpt-6.1-sol", "seconds": 4.5, "timestamp": "2026-10-23T19:55:42Z"}; CLI versions before {"codex": "codex-cli 0.162.1"}, after {"codex": "codex-cli 0.162.1"}
   axum workspace config presence (names only): {".claude/settings.json": false, ".claude/settings.local.json": false, ".clawpatch/config.json": false, ".codex": false, "clawpatch.config.json": false}
   fastapi probes: before {"error": null, "model": "gpt-6.1-sol", "seconds": 5.7, "timestamp": "2026-10-23T18:42:11Z"}, after {"error": null, "model": "gpt-6.1-sol", "seconds": 4.2, "timestamp": "2026-10-23T19:12:04Z"}; CLI versions before {"codex": "codex-cli 0.162.1"}, after {"codex": "codex-cli 0.162.1"}
   fastapi workspace config presence (names only): {".claude/settings.json": false, ".claude/settings.local.json": false, ".clawpatch/config.json": false, ".codex": false, "clawpatch.config.json": false}
operating characteristics (n=150 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 7 spurious; both repos): {"0.03": 0.8401, "0.04": 0.5578, "0.05": 0.2733, "0.07": 0.0285, "0.1": 0.0002}
operating characteristics (n=216 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 10 spurious; both repos): {"0.03": 0.8784, "0.04": 0.5639, "0.05": 0.2318, "0.07": 0.0109, "0.1": 0.0}
operating characteristics (n=270 per repo, one threshold, conditional on C1, independent matches: anticonservative; C2+C3 need <= 13 spurious; both repos): {"0.03": 0.9314, "0.04": 0.6456, "0.05": 0.2671, "0.07": 0.0088, "0.1": 0.0}
   expected n 216-270 labelled occurrences per repo; C3 stops binding at n=140
replication: {"gating": false, "arm_A_threshold_pct": 30, "arm_B_verdict": "fallback", "arm_B_threshold_pct": null, "arm_B_gates_ok": true, "arm_B_status_at_threshold": "fail", "arm_A_citable": true, "arm_B_citable": true, "status": "not replicated"} (arm A citable: True, arm B citable: True)
== arm claude: verdict fallback, threshold None (no threshold meets the criteria: refresh only at onboarding and manually) [citable for a design change: yes]
   per-threshold status: {"30": "fail", "35": "fail", "40": "fail", "45": "fail", "50": "fail", "55": "fail", "60": "fail", "65": "fail", "70": "fail", "75": "fail", "80": "fail", "85": "fail", "90": "fail"}
   gradient (Jaccard vs different, exploratory): {'n': 142, 'r': None}
   survival margin at t* (min over repos of mean survival - 4/5): None (None)
   leave-one-transition-out, joint t* (descriptive) hold out T1: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T2: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T3: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T4: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T5: t*=30 per repo (held-out survival, matched, labelled, different): {"axum": {"different": 0, "held_out_survival": 0.28, "labelled": 7, "matched": 7, "measured": true, "prev_areas": 25}, "fastapi": {"different": 0, "held_out_survival": 0.75, "labelled": 9, "matched": 9, "measured": true, "prev_areas": 12}}
   leave-one-transition-out, joint t* (descriptive) hold out T6: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
-- axum coverage (T_MIN rows): {"T1": {"class": "remap", "sub": "fresh", "prevAreas": 23, "nextAgentFeatures": 19, "prevOwnedFiles": 161, "nextOwnedFiles": 157, "sharedOwnedFiles": 157, "zeroOverlapPriorAreas": 0}, "T2": {"class": "remap", "sub": "fresh", "prevAreas": 23, "nextAgentFeatures": 25, "prevOwnedFiles": 161, "nextOwnedFiles": 161, "sharedOwnedFiles": 161, "zeroOverlapPriorAreas": 0}, "T3": {"class": "remap", "sub": "same_state", "prevAreas": 23, "nextAgentFeatures": 17, "prevOwnedFiles": 161, "nextOwnedFiles": 156, "sharedOwnedFiles": 156, "zeroOverlapPriorAreas": 0}, "T4": {"class": "remap", "sub": "same_state", "prevAreas": 17, "nextAgentFeatures": 25, "prevOwnedFiles": 156, "nextOwnedFiles": 222, "sharedOwnedFiles": 156, "zeroOverlapPriorAreas": 0}, "T5": {"class": "mutation", "sub": "mutation", "prevAreas": 25, "nextAgentFeatures": 10, "prevOwnedFiles": 222, "nextOwnedFiles": 172, "sharedOwnedFiles": 167, "zeroOverlapPriorAreas": 0}, "T6": {"class": "mutation", "sub": "mutation", "prevAreas": 10, "nextAgentFeatures": 21, "prevOwnedFiles": 172, "nextOwnedFiles": 163, "sharedOwnedFiles": 160, "zeroOverlapPriorAreas": 0}}
-- axum label distribution: {"by_band_pct": {"30": {"overlapping_concern": 9}, "40": {"overlapping_concern": 3}, "45": {"overlapping_concern": 1}, "50": {"overlapping_concern": 12, "same_concern": 3}, "55": {"overlapping_concern": 3, "same_concern": 2}, "60": {"overlapping_concern": 6, "same_concern": 6}, "65": {"overlapping_concern": 1, "same_concern": 1}, "70": {"overlapping_concern": 3, "same_concern": 1}, "75": {"overlapping_concern": 2, "same_concern": 1}, "80": {"overlapping_concern": 2, "same_concern": 1}, "85": {"same_concern": 1}, "90": {"same_concern": 30}}, "by_class": {"mutation": {"overlapping_concern": 11, "same_concern": 6}, "remap": {"overlapping_concern": 31, "same_concern": 40}}, "over_call_check_band_90": {"different": 0, "labelled": 30, "matches": 30, "share": 0.0}}
-- axum over-call check (`different` share of labelled matches in the [0.90, 1.0] band): 0 of 30 labelled (30 matches): share 0.000
-- axum chained survival at 50% (design default; no adopted threshold): [{"step": "T3", "status": "ok", "activePrior": 23, "matched": 15, "created": 2, "kept": 1, "retired": 7, "f1MatchedAllSoFar": 15, "f1Areas": 23}, {"step": "T4", "status": "ok", "activePrior": 18, "matched": 13, "created": 12, "kept": 3, "retired": 2, "f1MatchedAllSoFar": 12, "f1Areas": 23}, {"step": "T5", "status": "ok", "activePrior": 28, "matched": 6, "created": 4, "kept": 11, "retired": 11, "f1MatchedAllSoFar": 4, "f1Areas": 23}, {"step": "T6", "status": "ok", "activePrior": 21, "matched": 11, "created": 10, "kept": 10, "retired": 0, "f1MatchedAllSoFar": 4, "f1Areas": 23}]
-- fastapi coverage (T_MIN rows): {"T1": {"class": "remap", "sub": "fresh", "prevAreas": 12, "nextAgentFeatures": 12, "prevOwnedFiles": 134, "nextOwnedFiles": 102, "sharedOwnedFiles": 102, "zeroOverlapPriorAreas": 0}, "T2": {"class": "remap", "sub": "fresh", "prevAreas": 12, "nextAgentFeatures": 9, "prevOwnedFiles": 134, "nextOwnedFiles": 65, "sharedOwnedFiles": 65, "zeroOverlapPriorAreas": 1}, "T3": {"class": "remap", "sub": "same_state", "prevAreas": 12, "nextAgentFeatures": 10, "prevOwnedFiles": 134, "nextOwnedFiles": 100, "sharedOwnedFiles": 99, "zeroOverlapPriorAreas": 0}, "T4": {"class": "remap", "sub": "same_state", "prevAreas": 10, "nextAgentFeatures": 12, "prevOwnedFiles": 100, "nextOwnedFiles": 98, "sharedOwnedFiles": 95, "zeroOverlapPriorAreas": 0}, "T5": {"class": "mutation", "sub": "mutation", "prevAreas": 12, "nextAgentFeatures": 10, "prevOwnedFiles": 98, "nextOwnedFiles": 81, "sharedOwnedFiles": 75, "zeroOverlapPriorAreas": 1}, "T6": {"class": "mutation", "sub": "mutation", "prevAreas": 10, "nextAgentFeatures": 12, "prevOwnedFiles": 81, "nextOwnedFiles": 119, "sharedOwnedFiles": 79, "zeroOverlapPriorAreas": 0}}
-- fastapi label distribution: {"by_band_pct": {"30": {"overlapping_concern": 2}, "35": {"same_concern": 1}, "40": {"overlapping_concern": 2, "same_concern": 2}, "45": {"overlapping_concern": 2}, "50": {"overlapping_concern": 3, "same_concern": 2}, "55": {"overlapping_concern": 1}, "60": {"overlapping_concern": 1, "same_concern": 4}, "65": {"overlapping_concern": 1}, "70": {"overlapping_concern": 1, "same_concern": 2}, "75": {"same_concern": 2}, "80": {"same_concern": 4}, "85": {"same_concern": 1}, "90": {"same_concern": 23}}, "by_class": {"mutation": {"overlapping_concern": 2, "same_concern": 16}, "remap": {"overlapping_concern": 11, "same_concern": 25}}, "over_call_check_band_90": {"different": 0, "labelled": 23, "matches": 23, "share": 0.0}}
-- fastapi over-call check (`different` share of labelled matches in the [0.90, 1.0] band): 0 of 23 labelled (23 matches): share 0.000
-- fastapi chained survival at 50% (design default; no adopted threshold): [{"step": "T3", "status": "ok", "activePrior": 12, "matched": 8, "created": 2, "kept": 3, "retired": 1, "f1MatchedAllSoFar": 8, "f1Areas": 12}, {"step": "T4", "status": "ok", "activePrior": 13, "matched": 7, "created": 5, "kept": 4, "retired": 2, "f1MatchedAllSoFar": 6, "f1Areas": 12}, {"step": "T5", "status": "ok", "activePrior": 16, "matched": 8, "created": 2, "kept": 6, "retired": 2, "f1MatchedAllSoFar": 5, "f1Areas": 12}, {"step": "T6", "status": "ok", "activePrior": 16, "matched": 11, "created": 1, "kept": 5, "retired": 0, "f1MatchedAllSoFar": 5, "f1Areas": 12}]
-- axum / decisive
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   121    88  0.769  0.280  0.727 [0.642,0.799]    29    24     9 0.636 0.614 |    88/88       0    42     0 (0/0/0)  0.000 [0.000,0.042] 0.477 0.000
   0.35   121    79  0.690  0.200  0.653 [0.565,0.732]    38    28    14 0.705 0.570 |    79/79       0    33     0 (0/0/0)  0.000 [0.000,0.046] 0.418 0.000
   0.40   121    79  0.690  0.200  0.653 [0.565,0.732]    38    28    14 0.705 0.570 |    79/79       0    33     0 (0/0/0)  0.000 [0.000,0.046] 0.418 0.000
   0.45   121    76  0.666  0.200  0.628 [0.539,0.709]    41    29    16 0.727 0.553 |    76/76       0    30     0 (0/0/0)  0.000 [0.000,0.048] 0.395 0.000
   0.50   121    75  0.659  0.200  0.620 [0.531,0.701]    42    29    17 0.727 0.547 |    75/75       0    29     0 (0/0/0)  0.000 [0.000,0.049] 0.387 0.000
   0.55   121    60  0.540  0.080  0.496 [0.408,0.584]    57    41    20 0.900 0.433 |    60/60       0    17     0 (0/0/0)  0.000 [0.000,0.060] 0.283 0.000
   0.60   121    55  0.499  0.040  0.455 [0.369,0.543]    62    44    22 0.900 0.382 |    55/55       0    14     0 (0/0/0)  0.000 [0.000,0.065] 0.255 0.000
   0.65   121    43  0.379  0.000  0.355 [0.276,0.444]    74    52    26 1.000 0.209 |    43/43       0     8     0 (0/0/0)  0.000 [0.000,0.082] 0.186 0.000
   0.70   121    41  0.355  0.000  0.339 [0.261,0.427]    76    54    26 1.000 0.171 |    41/41       0     7     0 (0/0/0)  0.000 [0.000,0.086] 0.171 0.000
   0.75   121    37  0.317  0.000  0.306 [0.231,0.393]    80    56    28 1.000 0.081 |    37/37       0     4     0 (0/0/0)  0.000 [0.000,0.094] 0.108 0.000
   0.80   121    34  0.295  0.000  0.281 [0.209,0.367]    83    59    28 1.000 0.000 |    34/34       0     2     0 (0/0/0)  0.000 [0.000,0.102] 0.059 0.000
   0.85   121    31  0.271  0.000  0.256 [0.187,0.341]    86    61    29 1.000 0.000 |    31/31       0     0     0 (0/0/0)  0.000 [0.000,0.110] 0.000 0.000
   0.90   121    30  0.264  0.000  0.248 [0.180,0.332]    87    62    29 1.000 0.000 |    30/30       0     0     0 (0/0/0)  0.000 [0.000,0.114] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "17/23", "0.783"], ["T2", "21/23", "0.913"], ["T3", "17/23", "0.783"], ["T4", "16/17", "0.941"], ["T5", "7/25", "0.560"], ["T6", "10/10", "1.000"]], "50": [["T1", "16/23", "0.739"], ["T2", "17/23", "0.870"], ["T3", "15/23", "0.696"], ["T4", "13/17", "0.882"], ["T5", "5/25", "0.560"], ["T6", "9/10", "1.000"]], "70": [["T1", "9/23", "0.522"], ["T2", "10/23", "0.565"], ["T3", "10/23", "0.565"], ["T4", "8/17", "0.647"], ["T5", "0/25", "0.440"], ["T6", "4/10", "0.700"]], "90": [["T1", "5/23", "0.391"], ["T2", "8/23", "0.478"], ["T3", "7/23", "0.435"], ["T4", "7/17", "0.647"], ["T5", "0/25", "0.440"], ["T6", "3/10", "0.700"]]}
-- axum / remap
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    86    71  0.833  0.739  0.826 [0.732,0.891]    15    13     2 0.727 0.563 |    71/71       0    31     0 (0/0/0)  0.000 [0.000,0.051] 0.437 0.000
   0.35    86    65  0.760  0.652  0.756 [0.655,0.834]    21    17     4 0.786 0.523 |    65/65       0    25     0 (0/0/0)  0.000 [0.000,0.056] 0.385 0.000
   0.40    86    65  0.760  0.652  0.756 [0.655,0.834]    21    17     4 0.786 0.523 |    65/65       0    25     0 (0/0/0)  0.000 [0.000,0.056] 0.385 0.000
   0.45    86    62  0.724  0.652  0.721 [0.618,0.805]    24    18     6 0.800 0.500 |    62/62       0    22     0 (0/0/0)  0.000 [0.000,0.058] 0.355 0.000
   0.50    86    61  0.713  0.652  0.709 [0.606,0.795]    25    18     7 0.800 0.492 |    61/61       0    21     0 (0/0/0)  0.000 [0.000,0.059] 0.344 0.000
   0.55    86    50  0.590  0.522  0.581 [0.476,0.680]    36    28     8 0.900 0.380 |    50/50       0    13     0 (0/0/0)  0.000 [0.000,0.071] 0.260 0.000
   0.60    86    46  0.538  0.522  0.535 [0.430,0.637]    40    31     9 1.000 0.326 |    46/46       0    10     0 (0/0/0)  0.000 [0.000,0.077] 0.217 0.000
   0.65    86    38  0.444  0.391  0.442 [0.342,0.547]    48    36    12 1.000 0.184 |    38/38       0     6     0 (0/0/0)  0.000 [0.000,0.092] 0.158 0.000
   0.70    86    37  0.433  0.391  0.430 [0.331,0.536]    49    37    12 1.000 0.162 |    37/37       0     6     0 (0/0/0)  0.000 [0.000,0.094] 0.162 0.000
   0.75    86    34  0.400  0.304  0.395 [0.299,0.501]    52    39    13 1.000 0.088 |    34/34       0     4     0 (0/0/0)  0.000 [0.000,0.102] 0.118 0.000
   0.80    86    31  0.368  0.261  0.360 [0.267,0.466]    55    42    13 1.000 0.000 |    31/31       0     2     0 (0/0/0)  0.000 [0.000,0.110] 0.065 0.000
   0.85    86    28  0.331  0.217  0.326 [0.236,0.430]    58    44    14 1.000 0.000 |    28/28       0     0     0 (0/0/0)  0.000 [0.000,0.121] 0.000 0.000
   0.90    86    27  0.320  0.217  0.314 [0.226,0.418]    59    45    14 1.000 0.000 |    27/27       0     0     0 (0/0/0)  0.000 [0.000,0.125] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "17/23", "0.783"], ["T2", "21/23", "0.913"], ["T3", "17/23", "0.783"], ["T4", "16/17", "0.941"]], "50": [["T1", "16/23", "0.739"], ["T2", "17/23", "0.870"], ["T3", "15/23", "0.696"], ["T4", "13/17", "0.882"]], "70": [["T1", "9/23", "0.522"], ["T2", "10/23", "0.565"], ["T3", "10/23", "0.565"], ["T4", "8/17", "0.647"]], "90": [["T1", "5/23", "0.391"], ["T2", "8/23", "0.478"], ["T3", "7/23", "0.435"], ["T4", "7/17", "0.647"]]}
-- axum / mutation
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    35    17  0.640  0.280  0.486 [0.330,0.644]    14    11     7 0.600 0.824 |    17/17       0    11     0 (0/0/0)  0.000 [0.000,0.184] 0.647 0.000
   0.35    35    14  0.550  0.200  0.400 [0.256,0.564]    17    11    10 0.609 0.786 |    14/14       0     8     0 (0/0/0)  0.000 [0.000,0.215] 0.571 0.000
   0.40    35    14  0.550  0.200  0.400 [0.256,0.564]    17    11    10 0.609 0.786 |    14/14       0     8     0 (0/0/0)  0.000 [0.000,0.215] 0.571 0.000
   0.45    35    14  0.550  0.200  0.400 [0.256,0.564]    17    11    10 0.609 0.786 |    14/14       0     8     0 (0/0/0)  0.000 [0.000,0.215] 0.571 0.000
   0.50    35    14  0.550  0.200  0.400 [0.256,0.564]    17    11    10 0.609 0.786 |    14/14       0     8     0 (0/0/0)  0.000 [0.000,0.215] 0.571 0.000
   0.55    35    10  0.440  0.080  0.286 [0.163,0.451]    21    13    12 0.667 0.700 |    10/10       0     4     0 (0/0/0)  0.000 [0.000,0.278] 0.400 0.000
   0.60    35     9  0.420  0.040  0.257 [0.142,0.421]    22    13    13 0.667 0.667 |     9/9        0     4     0 (0/0/0)  0.000 [0.000,0.299] 0.444 0.000
   0.65    35     5  0.250  0.000  0.143 [0.063,0.294]    26    16    14 0.944 0.400 |     5/5        0     2     0 (0/0/0)  0.000 [0.000,0.434] 0.400 0.000
   0.70    35     4  0.200  0.000  0.114 [0.045,0.260]    27    17    14 1.000 0.250 |     4/4        0     1     0 (0/0/0)  0.000 [0.000,0.490] 0.250 0.000
   0.75    35     3  0.150  0.000  0.086 [0.030,0.224]    28    17    15 1.000 0.000 |     3/3        0     0     0 (0/0/0)  0.000 [0.000,0.561] 0.000 0.000
   0.80    35     3  0.150  0.000  0.086 [0.030,0.224]    28    17    15 1.000 0.000 |     3/3        0     0     0 (0/0/0)  0.000 [0.000,0.561] 0.000 0.000
   0.85    35     3  0.150  0.000  0.086 [0.030,0.224]    28    17    15 1.000 0.000 |     3/3        0     0     0 (0/0/0)  0.000 [0.000,0.561] 0.000 0.000
   0.90    35     3  0.150  0.000  0.086 [0.030,0.224]    28    17    15 1.000 0.000 |     3/3        0     0     0 (0/0/0)  0.000 [0.000,0.561] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T5", "7/25", "0.560"], ["T6", "10/10", "1.000"]], "50": [["T5", "5/25", "0.560"], ["T6", "9/10", "1.000"]], "70": [["T5", "0/25", "0.440"], ["T6", "4/10", "0.700"]], "90": [["T5", "0/25", "0.440"], ["T6", "3/10", "0.700"]]}
-- axum / fresh
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    46    38  0.826  0.739  0.826 [0.693,0.909]     6     7     1 0.700 0.605 |    38/38       0    18     0 (0/0/0)  0.000 [0.000,0.092] 0.474 0.000
   0.35    46    36  0.783  0.739  0.783 [0.644,0.877]     8     8     2 0.727 0.583 |    36/36       0    16     0 (0/0/0)  0.000 [0.000,0.096] 0.444 0.000
   0.40    46    36  0.783  0.739  0.783 [0.644,0.877]     8     8     2 0.727 0.583 |    36/36       0    16     0 (0/0/0)  0.000 [0.000,0.096] 0.444 0.000
   0.45    46    34  0.739  0.696  0.739 [0.597,0.844]    10     9     3 0.727 0.559 |    34/34       0    14     0 (0/0/0)  0.000 [0.000,0.102] 0.412 0.000
   0.50    46    33  0.717  0.696  0.717 [0.575,0.827]    11     9     4 0.727 0.545 |    33/33       0    13     0 (0/0/0)  0.000 [0.000,0.104] 0.394 0.000
   0.55    46    25  0.543  0.522  0.543 [0.402,0.678]    19    16     5 0.900 0.400 |    25/25       0     7     0 (0/0/0)  0.000 [0.000,0.133] 0.280 0.000
   0.60    46    24  0.522  0.522  0.522 [0.381,0.659]    20    17     5 0.900 0.375 |    24/24       0     6     0 (0/0/0)  0.000 [0.000,0.138] 0.250 0.000
   0.65    46    20  0.435  0.391  0.435 [0.302,0.578]    24    20     6 1.000 0.250 |    20/20       0     4     0 (0/0/0)  0.000 [0.000,0.161] 0.200 0.000
   0.70    46    19  0.413  0.391  0.413 [0.283,0.557]    25    21     6 1.000 0.211 |    19/19       0     4     0 (0/0/0)  0.000 [0.000,0.168] 0.211 0.000
   0.75    46    16  0.348  0.304  0.348 [0.227,0.492]    28    23     7 1.000 0.062 |    16/16       0     2     0 (0/0/0)  0.000 [0.000,0.194] 0.125 0.000
   0.80    46    15  0.326  0.261  0.326 [0.209,0.470]    29    24     7 1.000 0.000 |    15/15       0     1     0 (0/0/0)  0.000 [0.000,0.204] 0.067 0.000
   0.85    46    14  0.304  0.217  0.304 [0.191,0.448]    30    25     7 1.000 0.000 |    14/14       0     0     0 (0/0/0)  0.000 [0.000,0.215] 0.000 0.000
   0.90    46    13  0.283  0.217  0.283 [0.173,0.425]    31    26     7 1.000 0.000 |    13/13       0     0     0 (0/0/0)  0.000 [0.000,0.228] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "17/23", "0.783"], ["T2", "21/23", "0.913"]], "50": [["T1", "16/23", "0.739"], ["T2", "17/23", "0.870"]], "70": [["T1", "9/23", "0.522"], ["T2", "10/23", "0.565"]], "90": [["T1", "5/23", "0.391"], ["T2", "8/23", "0.478"]]}
-- axum / same_state
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    40    33  0.840  0.739  0.825 [0.681,0.913]     9     6     1 0.786 0.515 |    33/33       0    13     0 (0/0/0)  0.000 [0.000,0.104] 0.394 0.000
   0.35    40    29  0.738  0.652  0.725 [0.572,0.839]    13     9     2 0.833 0.448 |    29/29       0     9     0 (0/0/0)  0.000 [0.000,0.117] 0.310 0.000
   0.40    40    29  0.738  0.652  0.725 [0.572,0.839]    13     9     2 0.833 0.448 |    29/29       0     9     0 (0/0/0)  0.000 [0.000,0.117] 0.310 0.000
   0.45    40    28  0.708  0.652  0.700 [0.546,0.819]    14     9     3 0.900 0.429 |    28/28       0     8     0 (0/0/0)  0.000 [0.000,0.121] 0.286 0.000
   0.50    40    28  0.708  0.652  0.700 [0.546,0.819]    14     9     3 0.900 0.429 |    28/28       0     8     0 (0/0/0)  0.000 [0.000,0.121] 0.286 0.000
   0.55    40    25  0.636  0.565  0.625 [0.470,0.758]    17    12     3 0.900 0.360 |    25/25       0     6     0 (0/0/0)  0.000 [0.000,0.133] 0.240 0.000
   0.60    40    22  0.555  0.522  0.550 [0.398,0.693]    20    14     4 1.000 0.273 |    22/22       0     4     0 (0/0/0)  0.000 [0.000,0.149] 0.182 0.000
   0.65    40    18  0.453  0.435  0.450 [0.307,0.602]    24    16     6 1.000 0.111 |    18/18       0     2     0 (0/0/0)  0.000 [0.000,0.176] 0.111 0.000
   0.70    40    18  0.453  0.435  0.450 [0.307,0.602]    24    16     6 1.000 0.111 |    18/18       0     2     0 (0/0/0)  0.000 [0.000,0.176] 0.111 0.000
   0.75    40    18  0.453  0.435  0.450 [0.307,0.602]    24    16     6 1.000 0.111 |    18/18       0     2     0 (0/0/0)  0.000 [0.000,0.176] 0.111 0.000
   0.80    40    16  0.409  0.348  0.400 [0.263,0.554]    26    18     6 1.000 0.000 |    16/16       0     1     0 (0/0/0)  0.000 [0.000,0.194] 0.062 0.000
   0.85    40    14  0.358  0.304  0.350 [0.221,0.505]    28    19     7 1.000 0.000 |    14/14       0     0     0 (0/0/0)  0.000 [0.000,0.215] 0.000 0.000
   0.90    40    14  0.358  0.304  0.350 [0.221,0.505]    28    19     7 1.000 0.000 |    14/14       0     0     0 (0/0/0)  0.000 [0.000,0.215] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T3", "17/23", "0.783"], ["T4", "16/17", "0.941"]], "50": [["T3", "15/23", "0.696"], ["T4", "13/17", "0.882"]], "70": [["T3", "10/23", "0.565"], ["T4", "8/17", "0.647"]], "90": [["T3", "7/23", "0.435"], ["T4", "7/17", "0.647"]]}
-- fastapi / decisive
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    68    54  0.800  0.667  0.794 [0.684,0.873]    11     4    10 0.826 0.481 |    54/54       0    13     0 (0/0/0)  0.000 [0.000,0.066] 0.241 0.000
   0.35    68    52  0.772  0.583  0.765 [0.651,0.850]    13     4    12 0.826 0.462 |    52/52       0    11     0 (0/0/0)  0.000 [0.000,0.069] 0.212 0.000
   0.40    68    51  0.758  0.500  0.750 [0.636,0.838]    14     5    12 0.826 0.451 |    51/51       0    11     0 (0/0/0)  0.000 [0.000,0.070] 0.216 0.000
   0.45    68    47  0.697  0.417  0.691 [0.574,0.788]    18     8    13 0.867 0.404 |    47/47       0     9     0 (0/0/0)  0.000 [0.000,0.076] 0.191 0.000
   0.50    68    45  0.669  0.333  0.662 [0.543,0.763]    20     9    14 0.909 0.378 |    45/45       0     7     0 (0/0/0)  0.000 [0.000,0.079] 0.156 0.000
   0.55    68    40  0.600  0.333  0.588 [0.470,0.697]    25    11    17 1.000 0.300 |    40/40       0     4     0 (0/0/0)  0.000 [0.000,0.088] 0.100 0.000
   0.60    68    39  0.586  0.333  0.574 [0.455,0.684]    26    12    17 1.000 0.282 |    39/39       0     3     0 (0/0/0)  0.000 [0.000,0.090] 0.077 0.000
   0.65    68    34  0.511  0.333  0.500 [0.384,0.616]    31    14    20 1.000 0.176 |    34/34       0     2     0 (0/0/0)  0.000 [0.000,0.102] 0.059 0.000
   0.70    68    33  0.494  0.333  0.485 [0.371,0.602]    32    15    20 1.000 0.152 |    33/33       0     1     0 (0/0/0)  0.000 [0.000,0.104] 0.030 0.000
   0.75    68    30  0.453  0.250  0.441 [0.329,0.559]    35    17    21 1.000 0.067 |    30/30       0     0     0 (0/0/0)  0.000 [0.000,0.114] 0.000 0.000
   0.80    68    28  0.419  0.250  0.412 [0.303,0.530]    37    18    22 1.000 0.000 |    28/28       0     0     0 (0/0/0)  0.000 [0.000,0.121] 0.000 0.000
   0.85    68    24  0.361  0.167  0.353 [0.250,0.472]    41    20    24 1.000 0.000 |    24/24       0     0     0 (0/0/0)  0.000 [0.000,0.138] 0.000 0.000
   0.90    68    23  0.344  0.167  0.338 [0.237,0.457]    42    21    24 1.000 0.000 |    23/23       0     0     0 (0/0/0)  0.000 [0.000,0.143] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "10/12", "0.917"], ["T2", "8/12", "1.000"], ["T3", "9/12", "0.917"], ["T4", "9/10", "1.000"], ["T5", "9/12", "0.833"], ["T6", "9/10", "1.000"]], "50": [["T1", "9/12", "0.917"], ["T2", "4/12", "0.750"], ["T3", "8/12", "0.917"], ["T4", "7/10", "0.800"], ["T5", "8/12", "0.833"], ["T6", "9/10", "1.000"]], "70": [["T1", "7/12", "0.833"], ["T2", "4/12", "0.750"], ["T3", "4/12", "0.750"], ["T4", "5/10", "0.700"], ["T5", "5/12", "0.667"], ["T6", "8/10", "1.000"]], "90": [["T1", "6/12", "0.833"], ["T2", "2/12", "0.583"], ["T3", "3/12", "0.667"], ["T4", "3/10", "0.600"], ["T5", "3/12", "0.583"], ["T6", "6/10", "0.900"]]}
-- fastapi / remap
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    46    36  0.787  0.667  0.783 [0.644,0.877]     7     2     8 0.714 0.556 |    36/36       0    11     0 (0/0/0)  0.000 [0.000,0.096] 0.306 0.000
   0.35    46    34  0.746  0.583  0.739 [0.597,0.844]     9     2    10 0.778 0.529 |    34/34       0     9     0 (0/0/0)  0.000 [0.000,0.102] 0.265 0.000
   0.40    46    33  0.725  0.500  0.717 [0.575,0.827]    10     3    10 0.778 0.515 |    33/33       0     9     0 (0/0/0)  0.000 [0.000,0.104] 0.273 0.000
   0.45    46    30  0.654  0.417  0.652 [0.508,0.773]    13     6    10 0.867 0.467 |    30/30       0     7     0 (0/0/0)  0.000 [0.000,0.114] 0.233 0.000
   0.50    46    28  0.613  0.333  0.609 [0.465,0.736]    15     7    11 0.958 0.429 |    28/28       0     5     0 (0/0/0)  0.000 [0.000,0.121] 0.179 0.000
   0.55    46    25  0.550  0.333  0.543 [0.402,0.678]    18     7    14 1.000 0.360 |    25/25       0     4     0 (0/0/0)  0.000 [0.000,0.133] 0.160 0.000
   0.60    46    24  0.529  0.333  0.522 [0.381,0.659]    19     8    14 1.000 0.333 |    24/24       0     3     0 (0/0/0)  0.000 [0.000,0.138] 0.125 0.000
   0.65    46    21  0.463  0.333  0.457 [0.322,0.598]    22    10    15 1.000 0.238 |    21/21       0     2     0 (0/0/0)  0.000 [0.000,0.155] 0.095 0.000
   0.70    46    20  0.438  0.333  0.435 [0.302,0.578]    23    11    15 1.000 0.200 |    20/20       0     1     0 (0/0/0)  0.000 [0.000,0.161] 0.050 0.000
   0.75    46    17  0.375  0.250  0.370 [0.245,0.514]    26    13    16 1.000 0.059 |    17/17       0     0     0 (0/0/0)  0.000 [0.000,0.184] 0.000 0.000
   0.80    46    16  0.350  0.250  0.348 [0.227,0.492]    27    13    17 1.000 0.000 |    16/16       0     0     0 (0/0/0)  0.000 [0.000,0.194] 0.000 0.000
   0.85    46    15  0.329  0.167  0.326 [0.209,0.470]    28    14    17 1.000 0.000 |    15/15       0     0     0 (0/0/0)  0.000 [0.000,0.204] 0.000 0.000
   0.90    46    14  0.304  0.167  0.304 [0.191,0.448]    29    15    17 1.000 0.000 |    14/14       0     0     0 (0/0/0)  0.000 [0.000,0.215] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "10/12", "0.917"], ["T2", "8/12", "1.000"], ["T3", "9/12", "0.917"], ["T4", "9/10", "1.000"]], "50": [["T1", "9/12", "0.917"], ["T2", "4/12", "0.750"], ["T3", "8/12", "0.917"], ["T4", "7/10", "0.800"]], "70": [["T1", "7/12", "0.833"], ["T2", "4/12", "0.750"], ["T3", "4/12", "0.750"], ["T4", "5/10", "0.700"]], "90": [["T1", "6/12", "0.833"], ["T2", "2/12", "0.583"], ["T3", "3/12", "0.667"], ["T4", "3/10", "0.600"]]}
-- fastapi / mutation
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    22    18  0.825  0.750  0.818 [0.615,0.927]     4     2     2 0.909 0.333 |    18/18       0     2     0 (0/0/0)  0.000 [0.000,0.176] 0.111 0.000
   0.35    22    18  0.825  0.750  0.818 [0.615,0.927]     4     2     2 0.909 0.333 |    18/18       0     2     0 (0/0/0)  0.000 [0.000,0.176] 0.111 0.000
   0.40    22    18  0.825  0.750  0.818 [0.615,0.927]     4     2     2 0.909 0.333 |    18/18       0     2     0 (0/0/0)  0.000 [0.000,0.176] 0.111 0.000
   0.45    22    17  0.783  0.667  0.773 [0.566,0.899]     5     2     3 0.909 0.294 |    17/17       0     2     0 (0/0/0)  0.000 [0.000,0.184] 0.118 0.000
   0.50    22    17  0.783  0.667  0.773 [0.566,0.899]     5     2     3 0.909 0.294 |    17/17       0     2     0 (0/0/0)  0.000 [0.000,0.184] 0.118 0.000
   0.55    22    15  0.700  0.500  0.682 [0.473,0.836]     7     4     3 1.000 0.200 |    15/15       0     0     0 (0/0/0)  0.000 [0.000,0.204] 0.000 0.000
   0.60    22    15  0.700  0.500  0.682 [0.473,0.836]     7     4     3 1.000 0.200 |    15/15       0     0     0 (0/0/0)  0.000 [0.000,0.204] 0.000 0.000
   0.65    22    13  0.608  0.417  0.591 [0.387,0.767]     9     4     5 1.000 0.077 |    13/13       0     0     0 (0/0/0)  0.000 [0.000,0.228] 0.000 0.000
   0.70    22    13  0.608  0.417  0.591 [0.387,0.767]     9     4     5 1.000 0.077 |    13/13       0     0     0 (0/0/0)  0.000 [0.000,0.228] 0.000 0.000
   0.75    22    13  0.608  0.417  0.591 [0.387,0.767]     9     4     5 1.000 0.077 |    13/13       0     0     0 (0/0/0)  0.000 [0.000,0.228] 0.000 0.000
   0.80    22    12  0.558  0.417  0.545 [0.347,0.731]    10     5     5 1.000 0.000 |    12/12       0     0     0 (0/0/0)  0.000 [0.000,0.242] 0.000 0.000
   0.85    22     9  0.425  0.250  0.409 [0.233,0.613]    13     6     7 1.000 0.000 |     9/9        0     0     0 (0/0/0)  0.000 [0.000,0.299] 0.000 0.000
   0.90    22     9  0.425  0.250  0.409 [0.233,0.613]    13     6     7 1.000 0.000 |     9/9        0     0     0 (0/0/0)  0.000 [0.000,0.299] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T5", "9/12", "0.833"], ["T6", "9/10", "1.000"]], "50": [["T5", "8/12", "0.833"], ["T6", "9/10", "1.000"]], "70": [["T5", "5/12", "0.667"], ["T6", "8/10", "1.000"]], "90": [["T5", "3/12", "0.583"], ["T6", "6/10", "0.900"]]}
-- fastapi / fresh
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    24    18  0.750  0.667  0.750 [0.551,0.880]     3     1     5 0.833 0.500 |    18/18       0     6     0 (0/0/0)  0.000 [0.000,0.176] 0.333 0.000
   0.35    24    16  0.667  0.583  0.667 [0.467,0.820]     5     1     7 1.000 0.438 |    16/16       0     4     0 (0/0/0)  0.000 [0.000,0.194] 0.250 0.000
   0.40    24    15  0.625  0.500  0.625 [0.427,0.788]     6     2     7 1.000 0.400 |    15/15       0     4     0 (0/0/0)  0.000 [0.000,0.204] 0.267 0.000
   0.45    24    14  0.583  0.417  0.583 [0.388,0.755]     7     3     7 1.000 0.357 |    14/14       0     3     0 (0/0/0)  0.000 [0.000,0.215] 0.214 0.000
   0.50    24    13  0.542  0.333  0.542 [0.351,0.721]     8     4     7 1.000 0.308 |    13/13       0     2     0 (0/0/0)  0.000 [0.000,0.228] 0.154 0.000
   0.55    24    12  0.500  0.333  0.500 [0.314,0.686]     9     4     8 1.000 0.250 |    12/12       0     2     0 (0/0/0)  0.000 [0.000,0.242] 0.167 0.000
   0.60    24    12  0.500  0.333  0.500 [0.314,0.686]     9     4     8 1.000 0.250 |    12/12       0     2     0 (0/0/0)  0.000 [0.000,0.242] 0.167 0.000
   0.65    24    11  0.458  0.333  0.458 [0.279,0.649]    10     5     8 1.000 0.182 |    11/11       0     1     0 (0/0/0)  0.000 [0.000,0.259] 0.091 0.000
   0.70    24    11  0.458  0.333  0.458 [0.279,0.649]    10     5     8 1.000 0.182 |    11/11       0     1     0 (0/0/0)  0.000 [0.000,0.259] 0.091 0.000
   0.75    24     9  0.375  0.250  0.375 [0.212,0.573]    12     6     9 1.000 0.000 |     9/9        0     0     0 (0/0/0)  0.000 [0.000,0.299] 0.000 0.000
   0.80    24     9  0.375  0.250  0.375 [0.212,0.573]    12     6     9 1.000 0.000 |     9/9        0     0     0 (0/0/0)  0.000 [0.000,0.299] 0.000 0.000
   0.85    24     8  0.333  0.167  0.333 [0.180,0.533]    13     7     9 1.000 0.000 |     8/8        0     0     0 (0/0/0)  0.000 [0.000,0.324] 0.000 0.000
   0.90    24     8  0.333  0.167  0.333 [0.180,0.533]    13     7     9 1.000 0.000 |     8/8        0     0     0 (0/0/0)  0.000 [0.000,0.324] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "10/12", "0.917"], ["T2", "8/12", "1.000"]], "50": [["T1", "9/12", "0.917"], ["T2", "4/12", "0.750"]], "70": [["T1", "7/12", "0.833"], ["T2", "4/12", "0.750"]], "90": [["T1", "6/12", "0.833"], ["T2", "2/12", "0.583"]]}
-- fastapi / same_state
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    22    18  0.825  0.750  0.818 [0.615,0.927]     4     1     3 0.714 0.611 |    18/18       0     5     0 (0/0/0)  0.000 [0.000,0.176] 0.278 0.000
   0.35    22    18  0.825  0.750  0.818 [0.615,0.927]     4     1     3 0.714 0.611 |    18/18       0     5     0 (0/0/0)  0.000 [0.000,0.176] 0.278 0.000
   0.40    22    18  0.825  0.750  0.818 [0.615,0.927]     4     1     3 0.714 0.611 |    18/18       0     5     0 (0/0/0)  0.000 [0.000,0.176] 0.278 0.000
   0.45    22    16  0.725  0.700  0.727 [0.518,0.868]     6     3     3 0.778 0.562 |    16/16       0     4     0 (0/0/0)  0.000 [0.000,0.194] 0.250 0.000
   0.50    22    15  0.683  0.667  0.682 [0.473,0.836]     7     3     4 0.778 0.533 |    15/15       0     3     0 (0/0/0)  0.000 [0.000,0.204] 0.200 0.000
   0.55    22    13  0.600  0.500  0.591 [0.387,0.767]     9     3     6 0.867 0.462 |    13/13       0     2     0 (0/0/0)  0.000 [0.000,0.228] 0.154 0.000
   0.60    22    12  0.558  0.417  0.545 [0.347,0.731]    10     4     6 0.958 0.417 |    12/12       0     1     0 (0/0/0)  0.000 [0.000,0.242] 0.083 0.000
   0.65    22    10  0.467  0.333  0.455 [0.269,0.653]    12     5     7 1.000 0.300 |    10/10       0     1     0 (0/0/0)  0.000 [0.000,0.278] 0.100 0.000
   0.70    22     9  0.417  0.333  0.409 [0.233,0.613]    13     6     7 1.000 0.222 |     9/9        0     0     0 (0/0/0)  0.000 [0.000,0.299] 0.000 0.000
   0.75    22     8  0.375  0.250  0.364 [0.197,0.570]    14     7     7 1.000 0.125 |     8/8        0     0     0 (0/0/0)  0.000 [0.000,0.324] 0.000 0.000
   0.80    22     7  0.325  0.250  0.318 [0.164,0.527]    15     7     8 1.000 0.000 |     7/7        0     0     0 (0/0/0)  0.000 [0.000,0.354] 0.000 0.000
   0.85    22     7  0.325  0.250  0.318 [0.164,0.527]    15     7     8 1.000 0.000 |     7/7        0     0     0 (0/0/0)  0.000 [0.000,0.354] 0.000 0.000
   0.90    22     6  0.275  0.250  0.273 [0.132,0.482]    16     8     8 1.000 0.000 |     6/6        0     0     0 (0/0/0)  0.000 [0.000,0.390] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T3", "9/12", "0.917"], ["T4", "9/10", "1.000"]], "50": [["T3", "8/12", "0.917"], ["T4", "7/10", "0.800"]], "70": [["T3", "4/12", "0.750"], ["T4", "5/10", "0.700"]], "90": [["T3", "3/12", "0.667"], ["T4", "3/10", "0.600"]]}
== arm codex: verdict adopt, threshold 30 (largest passing threshold (survival, spurious-merge and precision criteria in both repositories)) [citable for a design change: yes]
   per-threshold status: {"30": "pass", "35": "fail", "40": "fail", "45": "fail", "50": "fail", "55": "fail", "60": "fail", "65": "fail", "70": "fail", "75": "fail", "80": "fail", "85": "fail", "90": "fail"}
   gradient (Jaccard vs different, exploratory): {'n': 436, 'r': None}
   survival margin at t* (min over repos of mean survival - 4/5): 0.00354238984269944 (4531/1279080)
   AT t*=30 (section 7, prominent): unlabelled share of matches and worst-case spurious rate ((different + unlabelled) / matches) per repo: {"axum": {"labelled": 228, "matched": 228, "spurious_worst_case": 0.0, "unlabelled": 0, "unlabelled_share": 0.0}, "fastapi": {"labelled": 208, "matched": 208, "spurious_worst_case": 0.0, "unlabelled": 0, "unlabelled_share": 0.0}}
   leave-one-transition-out, joint t* (descriptive) hold out T1: t*=30 per repo (held-out survival, matched, labelled, different): {"axum": {"different": 0, "held_out_survival": 0.7254901960784313, "labelled": 37, "matched": 37, "measured": true, "prev_areas": 51}, "fastapi": {"different": 0, "held_out_survival": 0.8222222222222222, "labelled": 37, "matched": 37, "measured": true, "prev_areas": 45}}
   leave-one-transition-out, joint t* (descriptive) hold out T2: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T3: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T4: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T5: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
   leave-one-transition-out, joint t* (descriptive) hold out T6: t*=None per repo (held-out survival, matched, labelled, different): {"axum": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}, "fastapi": {"different": null, "held_out_survival": null, "labelled": null, "matched": null, "measured": true, "prev_areas": null}}
-- axum coverage (T_MIN rows): {"T1": {"class": "remap", "sub": "fresh", "prevAreas": 51, "nextAgentFeatures": 37, "prevOwnedFiles": 173, "nextOwnedFiles": 170, "sharedOwnedFiles": 170, "zeroOverlapPriorAreas": 0}, "T2": {"class": "remap", "sub": "fresh", "prevAreas": 51, "nextAgentFeatures": 42, "prevOwnedFiles": 173, "nextOwnedFiles": 170, "sharedOwnedFiles": 170, "zeroOverlapPriorAreas": 0}, "T3": {"class": "remap", "sub": "same_state", "prevAreas": 51, "nextAgentFeatures": 51, "prevOwnedFiles": 173, "nextOwnedFiles": 221, "sharedOwnedFiles": 170, "zeroOverlapPriorAreas": 0}, "T4": {"class": "remap", "sub": "same_state", "prevAreas": 51, "nextAgentFeatures": 38, "prevOwnedFiles": 221, "nextOwnedFiles": 167, "sharedOwnedFiles": 167, "zeroOverlapPriorAreas": 0}, "T5": {"class": "mutation", "sub": "mutation", "prevAreas": 38, "nextAgentFeatures": 44, "prevOwnedFiles": 167, "nextOwnedFiles": 179, "sharedOwnedFiles": 166, "zeroOverlapPriorAreas": 0}, "T6": {"class": "mutation", "sub": "mutation", "prevAreas": 44, "nextAgentFeatures": 38, "prevOwnedFiles": 179, "nextOwnedFiles": 178, "sharedOwnedFiles": 173, "zeroOverlapPriorAreas": 0}}
-- axum label distribution: {"by_band_pct": {"30": {"overlapping_concern": 6, "same_concern": 1}, "35": {"overlapping_concern": 4}, "40": {"overlapping_concern": 11, "same_concern": 1}, "45": {"overlapping_concern": 1}, "50": {"overlapping_concern": 20, "same_concern": 11}, "55": {"overlapping_concern": 12, "same_concern": 1}, "60": {"overlapping_concern": 9, "same_concern": 1}, "65": {"overlapping_concern": 4, "same_concern": 9}, "70": {"overlapping_concern": 2, "same_concern": 1}, "75": {"overlapping_concern": 4, "same_concern": 10}, "80": {"overlapping_concern": 1, "same_concern": 7}, "85": {"same_concern": 8}, "90": {"same_concern": 104}}, "by_class": {"mutation": {"overlapping_concern": 21, "same_concern": 50}, "remap": {"overlapping_concern": 53, "same_concern": 104}}, "over_call_check_band_90": {"different": 0, "labelled": 104, "matches": 104, "share": 0.0}}
-- axum over-call check (`different` share of labelled matches in the [0.90, 1.0] band): 0 of 104 labelled (104 matches): share 0.000
-- axum chained survival at 30%: [{"step": "T3", "status": "ok", "activePrior": 51, "matched": 45, "created": 6, "kept": 2, "retired": 4, "f1MatchedAllSoFar": 45, "f1Areas": 51}, {"step": "T4", "status": "ok", "activePrior": 53, "matched": 33, "created": 5, "kept": 18, "retired": 2, "f1MatchedAllSoFar": 33, "f1Areas": 51}, {"step": "T5", "status": "ok", "activePrior": 56, "matched": 38, "created": 6, "kept": 16, "retired": 2, "f1MatchedAllSoFar": 31, "f1Areas": 51}, {"step": "T6", "status": "ok", "activePrior": 60, "matched": 36, "created": 2, "kept": 17, "retired": 7, "f1MatchedAllSoFar": 28, "f1Areas": 51}]
-- fastapi coverage (T_MIN rows): {"T1": {"class": "remap", "sub": "fresh", "prevAreas": 45, "nextAgentFeatures": 38, "prevOwnedFiles": 436, "nextOwnedFiles": 370, "sharedOwnedFiles": 369, "zeroOverlapPriorAreas": 5}, "T2": {"class": "remap", "sub": "fresh", "prevAreas": 45, "nextAgentFeatures": 46, "prevOwnedFiles": 436, "nextOwnedFiles": 625, "sharedOwnedFiles": 436, "zeroOverlapPriorAreas": 0}, "T3": {"class": "remap", "sub": "same_state", "prevAreas": 45, "nextAgentFeatures": 41, "prevOwnedFiles": 436, "nextOwnedFiles": 628, "sharedOwnedFiles": 435, "zeroOverlapPriorAreas": 0}, "T4": {"class": "remap", "sub": "same_state", "prevAreas": 41, "nextAgentFeatures": 54, "prevOwnedFiles": 628, "nextOwnedFiles": 612, "sharedOwnedFiles": 612, "zeroOverlapPriorAreas": 2}, "T5": {"class": "mutation", "sub": "mutation", "prevAreas": 54, "nextAgentFeatures": 32, "prevOwnedFiles": 612, "nextOwnedFiles": 337, "sharedOwnedFiles": 335, "zeroOverlapPriorAreas": 14}, "T6": {"class": "mutation", "sub": "mutation", "prevAreas": 32, "nextAgentFeatures": 52, "prevOwnedFiles": 337, "nextOwnedFiles": 636, "sharedOwnedFiles": 334, "zeroOverlapPriorAreas": 0}}
-- fastapi label distribution: {"by_band_pct": {"30": {"overlapping_concern": 3}, "35": {"overlapping_concern": 1, "same_concern": 1}, "40": {"overlapping_concern": 3, "same_concern": 2}, "45": {"overlapping_concern": 1}, "50": {"overlapping_concern": 5}, "55": {"overlapping_concern": 9, "same_concern": 1}, "60": {"overlapping_concern": 12, "same_concern": 5}, "65": {"overlapping_concern": 3, "same_concern": 6}, "70": {"overlapping_concern": 7, "same_concern": 2}, "75": {"overlapping_concern": 5, "same_concern": 2}, "80": {"overlapping_concern": 14, "same_concern": 9}, "85": {"overlapping_concern": 3, "same_concern": 6}, "90": {"overlapping_concern": 1, "same_concern": 107}}, "by_class": {"mutation": {"overlapping_concern": 15, "same_concern": 47}, "remap": {"overlapping_concern": 52, "same_concern": 94}}, "over_call_check_band_90": {"different": 0, "labelled": 108, "matches": 108, "share": 0.0}}
-- fastapi over-call check (`different` share of labelled matches in the [0.90, 1.0] band): 0 of 108 labelled (108 matches): share 0.000
-- fastapi chained survival at 30%: [{"step": "T3", "status": "ok", "activePrior": 45, "matched": 34, "created": 7, "kept": 6, "retired": 5, "f1MatchedAllSoFar": 34, "f1Areas": 45}, {"step": "T4", "status": "ok", "activePrior": 47, "matched": 44, "created": 10, "kept": 3, "retired": 0, "f1MatchedAllSoFar": 34, "f1Areas": 45}, {"step": "T5", "status": "ok", "activePrior": 57, "matched": 31, "created": 1, "kept": 22, "retired": 4, "f1MatchedAllSoFar": 25, "f1Areas": 45}, {"step": "T6", "status": "ok", "activePrior": 54, "matched": 47, "created": 5, "kept": 7, "retired": 0, "f1MatchedAllSoFar": 24, "f1Areas": 45}]
-- axum / decisive
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   286   228  0.804  0.647  0.797 [0.747,0.840]    22    31    27 0.833 0.474 |   228/228      0    74     0 (0/0/0)  0.000 [0.000,0.017] 0.325 0.000
   0.35   286   221  0.781  0.588  0.773 [0.721,0.817]    29    35    30 0.857 0.457 |   221/221      0    68     0 (0/0/0)  0.000 [0.000,0.017] 0.308 0.000
   0.40   286   217  0.766  0.588  0.759 [0.706,0.805]    33    37    32 0.857 0.447 |   217/217      0    64     0 (0/0/0)  0.000 [0.000,0.017] 0.295 0.000
   0.45   286   205  0.724  0.529  0.717 [0.662,0.766]    45    45    36 1.000 0.415 |   205/205      0    53     0 (0/0/0)  0.000 [0.000,0.018] 0.259 0.000
   0.50   286   204  0.721  0.529  0.713 [0.658,0.763]    46    45    37 1.000 0.412 |   204/204      0    52     0 (0/0/0)  0.000 [0.000,0.018] 0.255 0.000
   0.55   286   173  0.617  0.373  0.605 [0.547,0.660]    77    66    47 1.000 0.306 |   173/173      0    32     0 (0/0/0)  0.000 [0.000,0.022] 0.185 0.000
   0.60   286   160  0.569  0.353  0.559 [0.501,0.616]    90    77    49 1.000 0.250 |   160/160      0    20     0 (0/0/0)  0.000 [0.000,0.023] 0.125 0.000
   0.65   286   150  0.532  0.353  0.524 [0.467,0.582]   100    85    51 1.000 0.200 |   150/150      0    11     0 (0/0/0)  0.000 [0.000,0.025] 0.073 0.000
   0.70   286   137  0.487  0.333  0.479 [0.422,0.537]   113    93    56 1.000 0.124 |   137/137      0     7     0 (0/0/0)  0.000 [0.000,0.027] 0.051 0.000
   0.75   286   134  0.476  0.333  0.469 [0.411,0.526]   116    94    58 1.000 0.104 |   134/134      0     5     0 (0/0/0)  0.000 [0.000,0.028] 0.037 0.000
   0.80   286   120  0.427  0.314  0.420 [0.364,0.477]   130   100    66 1.000 0.000 |   120/120      0     1     0 (0/0/0)  0.000 [0.000,0.031] 0.008 0.000
   0.85   286   112  0.398  0.275  0.392 [0.337,0.449]   138   105    69 1.000 0.000 |   112/112      0     0     0 (0/0/0)  0.000 [0.000,0.033] 0.000 0.000
   0.90   286   104  0.370  0.216  0.364 [0.310,0.421]   146   112    70 1.000 0.000 |   104/104      0     0     0 (0/0/0)  0.000 [0.000,0.036] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "37/51", "0.804"], ["T2", "42/51", "0.882"], ["T3", "45/51", "0.922"], ["T4", "33/51", "0.961"], ["T5", "36/38", "0.947"], ["T6", "35/44", "0.841"]], "50": [["T1", "33/51", "0.745"], ["T2", "40/51", "0.863"], ["T3", "38/51", "0.804"], ["T4", "27/51", "0.941"], ["T5", "33/38", "0.895"], ["T6", "33/44", "0.818"]], "70": [["T1", "19/51", "0.529"], ["T2", "32/51", "0.745"], ["T3", "20/51", "0.510"], ["T4", "17/51", "0.882"], ["T5", "23/38", "0.658"], ["T6", "26/44", "0.727"]], "90": [["T1", "13/51", "0.431"], ["T2", "27/51", "0.667"], ["T3", "15/51", "0.490"], ["T4", "11/51", "0.824"], ["T5", "18/38", "0.579"], ["T6", "20/44", "0.659"]]}
-- axum / remap
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   204   157  0.770  0.647  0.770 [0.707,0.822]    11    22    25 0.778 0.510 |   157/157      0    53     0 (0/0/0)  0.000 [0.000,0.024] 0.338 0.000
   0.35   204   150  0.735  0.588  0.735 [0.671,0.791]    18    26    28 0.800 0.487 |   150/150      0    47     0 (0/0/0)  0.000 [0.000,0.025] 0.313 0.000
   0.40   204   147  0.721  0.588  0.721 [0.655,0.778]    21    27    30 0.800 0.476 |   147/147      0    44     0 (0/0/0)  0.000 [0.000,0.025] 0.299 0.000
   0.45   204   139  0.681  0.529  0.681 [0.615,0.741]    29    33    32 0.857 0.446 |   139/139      0    37     0 (0/0/0)  0.000 [0.000,0.027] 0.266 0.000
   0.50   204   138  0.676  0.529  0.676 [0.610,0.737]    30    33    33 0.857 0.442 |   138/138      0    36     0 (0/0/0)  0.000 [0.000,0.027] 0.261 0.000
   0.55   204   110  0.539  0.373  0.539 [0.471,0.606]    58    51    43 1.000 0.300 |   110/110      0    19     0 (0/0/0)  0.000 [0.000,0.034] 0.173 0.000
   0.60   204   103  0.505  0.353  0.505 [0.437,0.573]    65    56    45 1.000 0.252 |   103/103      0    13     0 (0/0/0)  0.000 [0.000,0.036] 0.126 0.000
   0.65   204    97  0.475  0.353  0.475 [0.408,0.544]    71    61    46 1.000 0.206 |    97/97       0     8     0 (0/0/0)  0.000 [0.000,0.038] 0.082 0.000
   0.70   204    88  0.431  0.333  0.431 [0.365,0.500]    80    68    48 1.000 0.125 |    88/88       0     5     0 (0/0/0)  0.000 [0.000,0.042] 0.057 0.000
   0.75   204    87  0.426  0.333  0.426 [0.361,0.495]    81    69    48 1.000 0.115 |    87/87       0     4     0 (0/0/0)  0.000 [0.000,0.042] 0.046 0.000
   0.80   204    77  0.377  0.314  0.377 [0.314,0.446]    91    73    54 1.000 0.000 |    77/77       0     1     0 (0/0/0)  0.000 [0.000,0.048] 0.013 0.000
   0.85   204    72  0.353  0.275  0.353 [0.291,0.421]    96    76    56 1.000 0.000 |    72/72       0     0     0 (0/0/0)  0.000 [0.000,0.051] 0.000 0.000
   0.90   204    66  0.324  0.216  0.324 [0.263,0.390]   102    81    57 1.000 0.000 |    66/66       0     0     0 (0/0/0)  0.000 [0.000,0.055] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "37/51", "0.804"], ["T2", "42/51", "0.882"], ["T3", "45/51", "0.922"], ["T4", "33/51", "0.961"]], "50": [["T1", "33/51", "0.745"], ["T2", "40/51", "0.863"], ["T3", "38/51", "0.804"], ["T4", "27/51", "0.941"]], "70": [["T1", "19/51", "0.529"], ["T2", "32/51", "0.745"], ["T3", "20/51", "0.510"], ["T4", "17/51", "0.882"]], "90": [["T1", "13/51", "0.431"], ["T2", "27/51", "0.667"], ["T3", "15/51", "0.490"], ["T4", "11/51", "0.824"]]}
-- axum / mutation
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    82    71  0.871  0.795  0.866 [0.776,0.923]    11     9     2 1.000 0.394 |    71/71       0    21     0 (0/0/0)  0.000 [0.000,0.051] 0.296 0.000
   0.35    82    71  0.871  0.795  0.866 [0.776,0.923]    11     9     2 1.000 0.394 |    71/71       0    21     0 (0/0/0)  0.000 [0.000,0.051] 0.296 0.000
   0.40    82    70  0.858  0.795  0.854 [0.761,0.914]    12    10     2 1.000 0.386 |    70/70       0    20     0 (0/0/0)  0.000 [0.000,0.052] 0.286 0.000
   0.45    82    66  0.809  0.750  0.805 [0.706,0.876]    16    12     4 1.000 0.348 |    66/66       0    16     0 (0/0/0)  0.000 [0.000,0.055] 0.242 0.000
   0.50    82    66  0.809  0.750  0.805 [0.706,0.876]    16    12     4 1.000 0.348 |    66/66       0    16     0 (0/0/0)  0.000 [0.000,0.055] 0.242 0.000
   0.55    82    63  0.772  0.727  0.768 [0.666,0.846]    19    15     4 1.000 0.317 |    63/63       0    13     0 (0/0/0)  0.000 [0.000,0.057] 0.206 0.000
   0.60    82    57  0.696  0.682  0.695 [0.589,0.784]    25    21     4 1.000 0.246 |    57/57       0     7     0 (0/0/0)  0.000 [0.000,0.063] 0.123 0.000
   0.65    82    53  0.645  0.632  0.646 [0.538,0.741]    29    24     5 1.000 0.189 |    53/53       0     3     0 (0/0/0)  0.000 [0.000,0.068] 0.057 0.000
   0.70    82    49  0.598  0.591  0.598 [0.489,0.697]    33    25     8 1.000 0.122 |    49/49       0     2     0 (0/0/0)  0.000 [0.000,0.073] 0.041 0.000
   0.75    82    47  0.575  0.545  0.573 [0.465,0.675]    35    25    10 1.000 0.085 |    47/47       0     1     0 (0/0/0)  0.000 [0.000,0.076] 0.021 0.000
   0.80    82    43  0.526  0.500  0.524 [0.418,0.629]    39    27    12 1.000 0.000 |    43/43       0     0     0 (0/0/0)  0.000 [0.000,0.082] 0.000 0.000
   0.85    82    40  0.489  0.477  0.488 [0.383,0.594]    42    29    13 1.000 0.000 |    40/40       0     0     0 (0/0/0)  0.000 [0.000,0.088] 0.000 0.000
   0.90    82    38  0.464  0.455  0.463 [0.360,0.571]    44    31    13 1.000 0.000 |    38/38       0     0     0 (0/0/0)  0.000 [0.000,0.092] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T5", "36/38", "0.947"], ["T6", "35/44", "0.841"]], "50": [["T5", "33/38", "0.895"], ["T6", "33/44", "0.818"]], "70": [["T5", "23/38", "0.658"], ["T6", "26/44", "0.727"]], "90": [["T5", "18/38", "0.579"], ["T6", "20/44", "0.659"]]}
-- axum / fresh
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   102    79  0.775  0.725  0.775 [0.684,0.845]     0    16     7 1.000 0.430 |    79/79       0    27     0 (0/0/0)  0.000 [0.000,0.046] 0.342 0.000
   0.35   102    78  0.765  0.706  0.765 [0.674,0.836]     1    17     7 1.000 0.423 |    78/78       0    26     0 (0/0/0)  0.000 [0.000,0.047] 0.333 0.000
   0.40   102    76  0.745  0.686  0.745 [0.653,0.820]     3    17     9 1.000 0.408 |    76/76       0    24     0 (0/0/0)  0.000 [0.000,0.048] 0.316 0.000
   0.45   102    73  0.716  0.647  0.716 [0.622,0.794]     6    20     9 1.000 0.384 |    73/73       0    21     0 (0/0/0)  0.000 [0.000,0.050] 0.288 0.000
   0.50   102    73  0.716  0.647  0.716 [0.622,0.794]     6    20     9 1.000 0.384 |    73/73       0    21     0 (0/0/0)  0.000 [0.000,0.050] 0.288 0.000
   0.55   102    65  0.637  0.529  0.637 [0.541,0.724]    14    25    12 1.000 0.308 |    65/65       0    13     0 (0/0/0)  0.000 [0.000,0.056] 0.200 0.000
   0.60   102    60  0.588  0.471  0.588 [0.491,0.679]    19    29    13 1.000 0.250 |    60/60       0     9     0 (0/0/0)  0.000 [0.000,0.060] 0.150 0.000
   0.65   102    55  0.539  0.431  0.539 [0.443,0.633]    24    33    14 1.000 0.182 |    55/55       0     5     0 (0/0/0)  0.000 [0.000,0.065] 0.091 0.000
   0.70   102    51  0.500  0.373  0.500 [0.405,0.595]    28    37    14 1.000 0.118 |    51/51       0     3     0 (0/0/0)  0.000 [0.000,0.070] 0.059 0.000
   0.75   102    51  0.500  0.373  0.500 [0.405,0.595]    28    37    14 1.000 0.118 |    51/51       0     3     0 (0/0/0)  0.000 [0.000,0.070] 0.059 0.000
   0.80   102    45  0.441  0.314  0.441 [0.349,0.538]    34    41    16 1.000 0.000 |    45/45       0     0     0 (0/0/0)  0.000 [0.000,0.079] 0.000 0.000
   0.85   102    43  0.422  0.294  0.422 [0.330,0.519]    36    43    16 1.000 0.000 |    43/43       0     0     0 (0/0/0)  0.000 [0.000,0.082] 0.000 0.000
   0.90   102    40  0.392  0.255  0.392 [0.303,0.489]    39    46    16 1.000 0.000 |    40/40       0     0     0 (0/0/0)  0.000 [0.000,0.088] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "37/51", "0.804"], ["T2", "42/51", "0.882"]], "50": [["T1", "33/51", "0.745"], ["T2", "40/51", "0.863"]], "70": [["T1", "19/51", "0.529"], ["T2", "32/51", "0.745"]], "90": [["T1", "13/51", "0.431"], ["T2", "27/51", "0.667"]]}
-- axum / same_state
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   102    78  0.765  0.647  0.765 [0.674,0.836]    11     6    18 0.667 0.590 |    78/78       0    26     0 (0/0/0)  0.000 [0.000,0.047] 0.333 0.000
   0.35   102    72  0.706  0.588  0.706 [0.611,0.786]    17     9    21 0.750 0.556 |    72/72       0    21     0 (0/0/0)  0.000 [0.000,0.051] 0.292 0.000
   0.40   102    71  0.696  0.588  0.696 [0.601,0.777]    18    10    21 0.750 0.549 |    71/71       0    20     0 (0/0/0)  0.000 [0.000,0.051] 0.282 0.000
   0.45   102    66  0.647  0.529  0.647 [0.551,0.733]    23    13    23 0.750 0.515 |    66/66       0    16     0 (0/0/0)  0.000 [0.000,0.055] 0.242 0.000
   0.50   102    65  0.637  0.529  0.637 [0.541,0.724]    24    13    24 0.750 0.508 |    65/65       0    15     0 (0/0/0)  0.000 [0.000,0.056] 0.231 0.000
   0.55   102    45  0.441  0.373  0.441 [0.349,0.538]    44    26    31 1.000 0.289 |    45/45       0     6     0 (0/0/0)  0.000 [0.000,0.079] 0.133 0.000
   0.60   102    43  0.422  0.353  0.422 [0.330,0.519]    46    27    32 1.000 0.256 |    43/43       0     4     0 (0/0/0)  0.000 [0.000,0.082] 0.093 0.000
   0.65   102    42  0.412  0.353  0.412 [0.321,0.509]    47    28    32 1.000 0.238 |    42/42       0     3     0 (0/0/0)  0.000 [0.000,0.084] 0.071 0.000
   0.70   102    37  0.363  0.333  0.363 [0.276,0.459]    52    31    34 1.000 0.135 |    37/37       0     2     0 (0/0/0)  0.000 [0.000,0.094] 0.054 0.000
   0.75   102    36  0.353  0.333  0.353 [0.267,0.449]    53    32    34 1.000 0.111 |    36/36       0     1     0 (0/0/0)  0.000 [0.000,0.096] 0.028 0.000
   0.80   102    32  0.314  0.314  0.314 [0.232,0.409]    57    32    38 1.000 0.000 |    32/32       0     1     0 (0/0/0)  0.000 [0.000,0.107] 0.031 0.000
   0.85   102    29  0.284  0.275  0.284 [0.206,0.378]    60    33    40 1.000 0.000 |    29/29       0     0     0 (0/0/0)  0.000 [0.000,0.117] 0.000 0.000
   0.90   102    26  0.255  0.216  0.255 [0.180,0.347]    63    35    41 1.000 0.000 |    26/26       0     0     0 (0/0/0)  0.000 [0.000,0.129] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T3", "45/51", "0.922"], ["T4", "33/51", "0.961"]], "50": [["T3", "38/51", "0.804"], ["T4", "27/51", "0.941"]], "70": [["T3", "20/51", "0.510"], ["T4", "17/51", "0.882"]], "90": [["T3", "15/51", "0.490"], ["T4", "11/51", "0.824"]]}
-- fastapi / decisive
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   262   208  0.812  0.574  0.794 [0.741,0.838]    55    15    39 0.909 0.327 |   208/208      0    67     0 (0/0/0)  0.000 [0.000,0.018] 0.322 0.000
   0.35   262   205  0.799  0.574  0.782 [0.729,0.828]    58    16    41 0.909 0.317 |   205/205      0    64     0 (0/0/0)  0.000 [0.000,0.018] 0.312 0.000
   0.40   262   203  0.792  0.574  0.775 [0.720,0.821]    60    17    42 0.909 0.310 |   203/203      0    63     0 (0/0/0)  0.000 [0.000,0.019] 0.310 0.000
   0.45   262   198  0.770  0.556  0.756 [0.700,0.804]    65    19    45 0.933 0.293 |   198/198      0    60     0 (0/0/0)  0.000 [0.000,0.019] 0.303 0.000
   0.50   262   197  0.766  0.556  0.752 [0.696,0.800]    66    20    45 0.933 0.289 |   197/197      0    59     0 (0/0/0)  0.000 [0.000,0.019] 0.299 0.000
   0.55   262   192  0.744  0.556  0.733 [0.676,0.783]    71    21    49 1.000 0.271 |   192/192      0    54     0 (0/0/0)  0.000 [0.000,0.020] 0.281 0.000
   0.60   262   182  0.705  0.537  0.695 [0.636,0.747]    81    23    57 1.000 0.231 |   182/182      0    45     0 (0/0/0)  0.000 [0.000,0.021] 0.247 0.000
   0.65   262   165  0.636  0.500  0.630 [0.570,0.686]    98    27    70 1.000 0.152 |   165/165      0    33     0 (0/0/0)  0.000 [0.000,0.023] 0.200 0.000
   0.70   262   156  0.601  0.463  0.595 [0.535,0.653]   107    31    75 1.000 0.103 |   156/156      0    30     0 (0/0/0)  0.000 [0.000,0.024] 0.192 0.000
   0.75   262   147  0.565  0.463  0.561 [0.501,0.620]   116    35    80 1.000 0.048 |   147/147      0    23     0 (0/0/0)  0.000 [0.000,0.025] 0.156 0.000
   0.80   262   140  0.539  0.444  0.534 [0.474,0.594]   123    36    86 1.000 0.000 |   140/140      0    18     0 (0/0/0)  0.000 [0.000,0.027] 0.129 0.000
   0.85   262   117  0.451  0.378  0.447 [0.388,0.507]   146    46    99 1.000 0.000 |   117/117      0     4     0 (0/0/0)  0.000 [0.000,0.032] 0.034 0.000
   0.90   262   108  0.418  0.333  0.412 [0.354,0.473]   155    49   105 1.000 0.000 |   108/108      0     1     0 (0/0/0)  0.000 [0.000,0.034] 0.009 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "37/45", "0.956"], ["T2", "36/45", "0.911"], ["T3", "34/45", "0.889"], ["T4", "39/41", "1.000"], ["T5", "31/54", "0.926"], ["T6", "31/32", "1.000"]], "50": [["T1", "37/45", "0.956"], ["T2", "35/45", "0.889"], ["T3", "31/45", "0.844"], ["T4", "36/41", "0.976"], ["T5", "30/54", "0.926"], ["T6", "28/32", "0.969"]], "70": [["T1", "32/45", "0.911"], ["T2", "29/45", "0.844"], ["T3", "24/45", "0.800"], ["T4", "27/41", "0.927"], ["T5", "25/54", "0.889"], ["T6", "19/32", "0.938"]], "90": [["T1", "24/45", "0.844"], ["T2", "20/45", "0.778"], ["T3", "16/45", "0.733"], ["T4", "14/41", "0.829"], ["T5", "18/54", "0.815"], ["T6", "16/32", "0.906"]]}
-- fastapi / remap
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   176   146  0.832  0.756  0.830 [0.767,0.878]    33    11    19 0.909 0.329 |   146/146      0    52     0 (0/0/0)  0.000 [0.000,0.026] 0.356 0.000
   0.35   176   144  0.821  0.756  0.818 [0.755,0.868]    35    12    20 0.909 0.319 |   144/144      0    50     0 (0/0/0)  0.000 [0.000,0.026] 0.347 0.000
   0.40   176   142  0.809  0.711  0.807 [0.742,0.858]    37    13    21 0.909 0.310 |   142/142      0    49     0 (0/0/0)  0.000 [0.000,0.026] 0.345 0.000
   0.45   176   140  0.798  0.689  0.795 [0.730,0.848]    39    14    22 0.909 0.300 |   140/140      0    48     0 (0/0/0)  0.000 [0.000,0.027] 0.343 0.000
   0.50   176   139  0.792  0.689  0.790 [0.724,0.843]    40    15    22 0.909 0.295 |   139/139      0    47     0 (0/0/0)  0.000 [0.000,0.027] 0.338 0.000
   0.55   176   136  0.775  0.689  0.773 [0.705,0.828]    43    16    24 0.929 0.279 |   136/136      0    44     0 (0/0/0)  0.000 [0.000,0.027] 0.324 0.000
   0.60   176   128  0.728  0.622  0.727 [0.657,0.788]    51    18    30 1.000 0.234 |   128/128      0    37     0 (0/0/0)  0.000 [0.000,0.029] 0.289 0.000
   0.65   176   117  0.665  0.556  0.665 [0.592,0.730]    62    21    38 1.000 0.162 |   117/117      0    28     0 (0/0/0)  0.000 [0.000,0.032] 0.239 0.000
   0.70   176   112  0.637  0.533  0.636 [0.563,0.704]    67    23    41 1.000 0.125 |   112/112      0    25     0 (0/0/0)  0.000 [0.000,0.033] 0.223 0.000
   0.75   176   104  0.591  0.489  0.591 [0.517,0.661]    75    27    45 1.000 0.058 |   104/104      0    19     0 (0/0/0)  0.000 [0.000,0.036] 0.183 0.000
   0.80   176    98  0.557  0.467  0.557 [0.483,0.628]    81    28    50 1.000 0.000 |    98/98       0    15     0 (0/0/0)  0.000 [0.000,0.038] 0.153 0.000
   0.85   176    80  0.454  0.378  0.455 [0.383,0.528]    99    35    61 1.000 0.000 |    80/80       0     3     0 (0/0/0)  0.000 [0.000,0.046] 0.037 0.000
   0.90   176    74  0.419  0.341  0.420 [0.350,0.494]   105    36    66 1.000 0.000 |    74/74       0     1     0 (0/0/0)  0.000 [0.000,0.049] 0.014 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "37/45", "0.956"], ["T2", "36/45", "0.911"], ["T3", "34/45", "0.889"], ["T4", "39/41", "1.000"]], "50": [["T1", "37/45", "0.956"], ["T2", "35/45", "0.889"], ["T3", "31/45", "0.844"], ["T4", "36/41", "0.976"]], "70": [["T1", "32/45", "0.911"], ["T2", "29/45", "0.844"], ["T3", "24/45", "0.800"], ["T4", "27/41", "0.927"]], "90": [["T1", "24/45", "0.844"], ["T2", "20/45", "0.778"], ["T3", "16/45", "0.733"], ["T4", "14/41", "0.829"]]}
-- fastapi / mutation
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    86    62  0.771  0.574  0.721 [0.618,0.805]    22     4    20 1.000 0.323 |    62/62       0    15     0 (0/0/0)  0.000 [0.000,0.058] 0.242 0.000
   0.35    86    61  0.756  0.574  0.709 [0.606,0.795]    23     4    21 1.000 0.311 |    61/61       0    14     0 (0/0/0)  0.000 [0.000,0.059] 0.230 0.000
   0.40    86    61  0.756  0.574  0.709 [0.606,0.795]    23     4    21 1.000 0.311 |    61/61       0    14     0 (0/0/0)  0.000 [0.000,0.059] 0.230 0.000
   0.45    86    58  0.715  0.556  0.674 [0.570,0.764]    26     5    23 1.000 0.276 |    58/58       0    12     0 (0/0/0)  0.000 [0.000,0.062] 0.207 0.000
   0.50    86    58  0.715  0.556  0.674 [0.570,0.764]    26     5    23 1.000 0.276 |    58/58       0    12     0 (0/0/0)  0.000 [0.000,0.062] 0.207 0.000
   0.55    86    56  0.684  0.556  0.651 [0.546,0.743]    28     5    25 1.000 0.250 |    56/56       0    10     0 (0/0/0)  0.000 [0.000,0.064] 0.179 0.000
   0.60    86    54  0.659  0.537  0.628 [0.522,0.723]    30     5    27 1.000 0.222 |    54/54       0     8     0 (0/0/0)  0.000 [0.000,0.066] 0.148 0.000
   0.65    86    48  0.578  0.500  0.558 [0.453,0.658]    36     6    32 1.000 0.125 |    48/48       0     5     0 (0/0/0)  0.000 [0.000,0.074] 0.104 0.000
   0.70    86    44  0.528  0.463  0.512 [0.408,0.614]    40     8    34 1.000 0.045 |    44/44       0     5     0 (0/0/0)  0.000 [0.000,0.080] 0.114 0.000
   0.75    86    43  0.513  0.463  0.500 [0.397,0.603]    41     8    35 1.000 0.023 |    43/43       0     4     0 (0/0/0)  0.000 [0.000,0.082] 0.093 0.000
   0.80    86    42  0.503  0.444  0.488 [0.386,0.592]    42     8    36 1.000 0.000 |    42/42       0     3     0 (0/0/0)  0.000 [0.000,0.084] 0.071 0.000
   0.85    86    37  0.444  0.389  0.430 [0.331,0.536]    47    11    38 1.000 0.000 |    37/37       0     1     0 (0/0/0)  0.000 [0.000,0.094] 0.027 0.000
   0.90    86    34  0.417  0.333  0.395 [0.299,0.501]    50    13    39 1.000 0.000 |    34/34       0     0     0 (0/0/0)  0.000 [0.000,0.102] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T5", "31/54", "0.926"], ["T6", "31/32", "1.000"]], "50": [["T5", "30/54", "0.926"], ["T6", "28/32", "0.969"]], "70": [["T5", "25/54", "0.889"], ["T6", "19/32", "0.938"]], "90": [["T5", "18/54", "0.815"], ["T6", "16/32", "0.906"]]}
-- fastapi / fresh
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    90    73  0.811  0.800  0.811 [0.718,0.879]    11     6    11 1.000 0.260 |    73/73       0    19     0 (0/0/0)  0.000 [0.000,0.050] 0.260 0.000
   0.35    90    72  0.800  0.778  0.800 [0.706,0.870]    12     7    11 1.000 0.250 |    72/72       0    18     0 (0/0/0)  0.000 [0.000,0.051] 0.250 0.000
   0.40    90    72  0.800  0.778  0.800 [0.706,0.870]    12     7    11 1.000 0.250 |    72/72       0    18     0 (0/0/0)  0.000 [0.000,0.051] 0.250 0.000
   0.45    90    72  0.800  0.778  0.800 [0.706,0.870]    12     7    11 1.000 0.250 |    72/72       0    18     0 (0/0/0)  0.000 [0.000,0.051] 0.250 0.000
   0.50    90    72  0.800  0.778  0.800 [0.706,0.870]    12     7    11 1.000 0.250 |    72/72       0    18     0 (0/0/0)  0.000 [0.000,0.051] 0.250 0.000
   0.55    90    70  0.778  0.733  0.778 [0.682,0.851]    14     8    12 1.000 0.229 |    70/70       0    16     0 (0/0/0)  0.000 [0.000,0.052] 0.229 0.000
   0.60    90    68  0.756  0.689  0.756 [0.658,0.833]    16     9    13 1.000 0.206 |    68/68       0    15     0 (0/0/0)  0.000 [0.000,0.053] 0.221 0.000
   0.65    90    64  0.711  0.667  0.711 [0.610,0.795]    20    10    16 1.000 0.156 |    64/64       0    12     0 (0/0/0)  0.000 [0.000,0.057] 0.188 0.000
   0.70    90    61  0.678  0.644  0.678 [0.576,0.765]    23    11    18 1.000 0.115 |    61/61       0    10     0 (0/0/0)  0.000 [0.000,0.059] 0.164 0.000
   0.75    90    58  0.644  0.622  0.644 [0.542,0.736]    26    13    19 1.000 0.069 |    58/58       0     8     0 (0/0/0)  0.000 [0.000,0.062] 0.138 0.000
   0.80    90    54  0.600  0.578  0.600 [0.497,0.695]    30    14    22 1.000 0.000 |    54/54       0     6     0 (0/0/0)  0.000 [0.000,0.066] 0.111 0.000
   0.85    90    46  0.511  0.467  0.511 [0.410,0.612]    38    16    28 1.000 0.000 |    46/46       0     0     0 (0/0/0)  0.000 [0.000,0.077] 0.000 0.000
   0.90    90    44  0.489  0.444  0.489 [0.388,0.590]    40    17    29 1.000 0.000 |    44/44       0     0     0 (0/0/0)  0.000 [0.000,0.080] 0.000 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T1", "37/45", "0.956"], ["T2", "36/45", "0.911"]], "50": [["T1", "37/45", "0.956"], ["T2", "35/45", "0.889"]], "70": [["T1", "32/45", "0.911"], ["T2", "29/45", "0.844"]], "90": [["T1", "24/45", "0.844"], ["T2", "20/45", "0.778"]]}
-- fastapi / same_state
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30    86    73  0.853  0.756  0.849 [0.758,0.909]    22     5     8 0.833 0.397 |    73/73       0    33     0 (0/0/0)  0.000 [0.000,0.050] 0.452 0.000
   0.35    86    72  0.841  0.756  0.837 [0.745,0.900]    23     5     9 0.833 0.389 |    72/72       0    32     0 (0/0/0)  0.000 [0.000,0.051] 0.444 0.000
   0.40    86    70  0.819  0.711  0.814 [0.719,0.882]    25     6    10 0.833 0.371 |    70/70       0    31     0 (0/0/0)  0.000 [0.000,0.052] 0.443 0.000
   0.45    86    68  0.796  0.689  0.791 [0.693,0.863]    27     7    11 0.857 0.353 |    68/68       0    30     0 (0/0/0)  0.000 [0.000,0.053] 0.441 0.000
   0.50    86    67  0.783  0.689  0.779 [0.681,0.854]    28     8    11 0.857 0.343 |    67/67       0    29     0 (0/0/0)  0.000 [0.000,0.054] 0.433 0.000
   0.55    86    66  0.771  0.689  0.767 [0.668,0.844]    29     8    12 0.862 0.333 |    66/66       0    28     0 (0/0/0)  0.000 [0.000,0.055] 0.424 0.000
   0.60    86    60  0.701  0.622  0.698 [0.594,0.785]    35     9    17 0.900 0.267 |    60/60       0    22     0 (0/0/0)  0.000 [0.000,0.060] 0.367 0.000
   0.65    86    53  0.619  0.556  0.616 [0.511,0.712]    42    11    22 1.000 0.170 |    53/53       0    16     0 (0/0/0)  0.000 [0.000,0.068] 0.302 0.000
   0.70    86    51  0.596  0.533  0.593 [0.487,0.691]    44    12    23 1.000 0.137 |    51/51       0    15     0 (0/0/0)  0.000 [0.000,0.070] 0.294 0.000
   0.75    86    46  0.537  0.489  0.535 [0.430,0.637]    49    14    26 1.000 0.043 |    46/46       0    11     0 (0/0/0)  0.000 [0.000,0.077] 0.239 0.000
   0.80    86    44  0.514  0.467  0.512 [0.408,0.614]    51    14    28 1.000 0.000 |    44/44       0     9     0 (0/0/0)  0.000 [0.000,0.080] 0.205 0.000
   0.85    86    34  0.396  0.378  0.395 [0.299,0.501]    61    19    33 1.000 0.000 |    34/34       0     3     0 (0/0/0)  0.000 [0.000,0.102] 0.088 0.000
   0.90    86    30  0.349  0.341  0.349 [0.257,0.454]    65    19    37 1.000 0.000 |    30/30       0     1     0 (0/0/0)  0.000 [0.000,0.114] 0.033 0.000
   per-transition survival (matched/prior areas; kept-inclusive (matched+kept)/prior areas) by threshold: {"30": [["T3", "34/45", "0.889"], ["T4", "39/41", "1.000"]], "50": [["T3", "31/45", "0.844"], ["T4", "36/41", "0.976"]], "70": [["T3", "24/45", "0.800"], ["T4", "27/41", "0.927"]], "90": [["T3", "16/45", "0.733"], ["T4", "14/41", "0.829"]]}
```
