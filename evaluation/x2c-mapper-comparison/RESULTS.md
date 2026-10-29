# X2c — Which agent mapper and which area-matching method keep Scout's areas stable, cheaply and quickly? Results

Preregistration: `PREREGISTRATION.md` revision 3 (signed off with zero council issues in `reviews/round-5.md`, after round 4's zero-issue review of the same revision; digests in §11). Hypothesis: **HD5c** (`docs/hypothesis-ledger.md`). Governed by `CLAUDE.md` experiment rules 1–7.
Run commit: `996541061e295101d197437544e6c0d31c085496` (worktree `/Users/adamjackson/Projects/inflight/scout-x2c`, pushed to `origin/main`). The run worktree is `git status`-clean apart from `?? evaluation/x2c-mapper-comparison/data/` after the run (checked for this report); no commit was made in it after the run commit.
Data: collected 2026-10-29. First before-probe 00:29:38Z (luna-xhigh, fastapi); first map attempt 00:29:40Z; last map attempt started 08:08:05Z; last after-probe 08:12:03Z (luna-xhigh, axum). 56 of 56 maps `ok`, rc 0, one attempt each (`/tmp/x2c-collect.log`, the collector's stdout). Then `run.sh replay`, `run.sh pairs` (1,986 unique real pairs, 239 controls), `run.sh label --workers 8` (interrupted) and `run.sh label --workers 8 --new-round resume-1`, `run.sh jev` (once, then one verbatim resume), `run.sh analysis` once (`data/results.json` sha256 `831ff555a61480260265fd0f89731ad1d41fff0d54850b812c55d13a1bfad4bc`, `data/analysis_output.txt` sha256 `9dc72a11146ac77963fd73ed23a770a1243075c0fd1ed5ae3d888785b6a3ec61`).
This study supersedes no earlier study and is superseded by none. It follows X2 (`../x2-area-stability/RESULTS.md`) and X2b (`../x2b-area-stability-luna/RESULTS.md`). Every number below is copied from `data/results.json` or `data/analysis_output.txt` or is a stated tabulation of those; nothing is recomputed from raw data except where a section says so (the best-threshold table of §1 is a maximum over the per-threshold values of `results.json`, and §11's digests are recomputed from git objects). `[INFERENCE]` marks statements not read from the outputs. The complete text of the analysis is `data/analysis_output.txt` (committed beside this file; it is 1,496 lines and is not copied here), and every table below is a rendering of `data/results.json`.

## 1. Headline

**SELECTION: `no-adoption`.** No cell is eligible. All 8 cells (4 mapper arms × {Jaccard, Jev}) have verdict `fallback`; all 13 thresholds of every cell have status `fail`. Reason string of every cell: "no threshold meets the criteria: refresh only at onboarding and manually". `results.json` `x2c.selection`: `{"cell": null, "outcome": "no-adoption", "proposal": "none: no cell is eligible; the design's fallback stays (agent map at onboarding and manually; no refresh-on-merge)"}`. Cells examined 8, eligible 0.

In plain terms: for whole-tree agent re-maps reconciled one-to-one, neither a cheaper or faster mapper (`claude-haiku-5-5`; `gpt-6.1-terra` at effort `low` or `none`) nor a Jev concern judgment on **file-sharing candidate pairs** (every candidate shares at least one owned file, §5 of the preregistration) makes areas stable enough across refreshes to refresh on merge. The design's fallback stays: the agent map runs at onboarding and on manual `scout map --refresh` only. Nothing in this study proposes a design change; `docs/scout-design.md` is not touched.

**Scope (added after the post-run critical review, 2026-10-29).** This study did **not** test matching by meaning without a file gate, nor associating merged changes with existing areas at merge time: every map is a full re-map of a fixed or synthetically mutated tree, not a merged PR; Jev judged pairs of map features (title, summary, kind, entrypoints, owned and shared files), never a PR, a diff or the area catalogue; and candidates without a shared owned file were never scored. The result cannot be cited for or against semantic association at merge.

**X2b comparison:** `agree`. X2b's luna `xhigh` Jaccard result was `fallback`; the luna-xhigh Jaccard cell here is `fallback` too (`x2b_comparison`: `{"agree": true, "cell": "luna-xhigh", "x2b_verdict": "fallback", "x2c_verdict": "fallback"}`).

**HD5c state per PREREGISTRATION §7: `no-adoption` → `not-supported`** (existential claim; two repositories, one run per arm; this grid, this guide; spurious merges as labelled by `agent_reviewed` labellers). The ledger is not edited by this document; the History line is a later step of the run order (h). HD5 and HD5b are unchanged.

**Why: no threshold meets the survival criterion in both repositories.** The criterion (PREREGISTRATION §7, "Per-cell verdict"; X2 §4): "C1 per-repository mean survival ≥ 4/5 (exact), C2 per-repository spurious rate ≤ 1/20, C3 Wilson upper bound ≤ 0.10; class gate, label gate (≤ 1/10 unlabelled), control gate (pooled over every arm and method: one labelling session, one set of controls), then `adopt t* = max(pass)`, else `inconclusive` (imprecise), else `fallback`." C1 is the binding criterion: in every cell fastapi's mean survival is below 0.8 at every one of the 13 thresholds (the largest value anywhere is 0.7887, terra-none+jev at 0.30), while axum's is at or above 0.8 at the best threshold in all cells. **Spurious merges were 0 in every cell**: zero `different_concern` labels among the labelled matches in every cell, repository, stratum and threshold, so C2 (≤ 1/20) passes everywhere. C3 (Wilson upper ≤ 0.10) additionally fails at some high thresholds from small labelled n alone (haiku from 0.45 up, luna-xhigh at 0.80–0.90), never on its own: C1 already fails at those thresholds. Labelled `different` = 0 in all 8 cells.

Best threshold per cell: the largest over the 13 thresholds of the smaller of the two repositories' decisive-stratum mean per-refresh survival (from `results.json` `arms.<cell>.tables.<repo>.decisive.<t>.survival`; the exact fraction is in the per-cell tables of §3). Where several thresholds tie, the lowest is shown and the plateau is given.

| cell | best t | fastapi survival | axum survival | min over repos | plateau (thresholds at the same min) | C1 (≥ 0.8) | `different` / labelled at best t (fastapi, axum) |
|---|---|---|---|---|---|---|---|
| haiku | 0.30 | 0.411 | 0.865 | 0.411 | 0.30 | fail (fastapi) | 0/42, 0/99 |
| haiku+jev | 0.30 | 0.547 | 0.910 | 0.547 | 0.30, 0.35, 0.40, 0.45, 0.50 | fail (fastapi) | 0/56, 0/104 |
| luna-xhigh | 0.30 | 0.473 | 0.936 | 0.473 | 0.30 | fail (fastapi) | 0/80, 0/442 |
| luna-xhigh+jev | 0.30 | 0.589 | 0.954 | 0.589 | 0.30, 0.35, 0.40, 0.45, 0.50, 0.55, 0.60, 0.65, 0.70, 0.75 | fail (fastapi) | 0/100, 0/451 |
| terra-low | 0.30 | 0.629 | 0.933 | 0.629 | 0.30 | fail (fastapi) | 0/80, 0/179 |
| terra-low+jev | 0.30 | 0.661 | 0.938 | 0.661 | 0.30, 0.35, 0.40, 0.45, 0.50, 0.55, 0.60, 0.65, 0.70, 0.75, 0.80, 0.85, 0.90 | fail (fastapi) | 0/84, 0/180 |
| terra-none | 0.30 | 0.752 | 0.905 | 0.752 | 0.30 | fail (fastapi) | 0/166, 0/203 |
| terra-none+jev | 0.30 | 0.789 | 0.918 | 0.789 | 0.30, 0.35, 0.40, 0.45, 0.50, 0.55, 0.60, 0.65, 0.70, 0.75, 0.80 | fail (fastapi) | 0/173, 0/206 |

(Columns "fastapi survival" and "axum survival" are the per-repository means at that threshold; every min is on fastapi. The eight (fastapi / axum) pairs: haiku 0.411 / 0.865, haiku+jev 0.547 / 0.910, luna-xhigh 0.473 / 0.936, luna-xhigh+jev 0.589 / 0.954, terra-low 0.629 / 0.933, terra-low+jev 0.661 / 0.938, terra-none 0.752 / 0.905, terra-none+jev 0.789 / 0.918 (fastapi / axum, all attained at t = 0.30; the Jev cells' survival is flat over a plateau of thresholds that begins at 0.30, listed above), each re-derived from `results.json` for this report.)

The best cell, terra-none with Jev, reaches 0.789 on fastapi: 0.011 below the 0.8 minimum, and still `fallback`. [INFERENCE] That is a miss, not an eligible cell, and the same draw could land either side of 0.8; the study has one run per arm and repository, so it cannot say how often.

**Candidate ceiling.** `1 − zeroOverlapPriorAreas / prevAreas`, mean over transitions (`results.json` `x2c.candidate_ceiling`): the most survival any matcher can reach on the candidate set, because a prior area that shares no owned file with any next-map feature is never a candidate for any matcher.

| arm | axum | fastapi | Jev-cell fastapi best survival (§1 table) | gap to ceiling |
|---|---|---|---|---|
| haiku | 0.9852 | 0.6336 | 0.547 (Jaccard 0.411) | 0.086 |
| luna-xhigh | 0.9819 | 0.7787 | 0.589 (Jaccard 0.473) | 0.190 |
| terra-low | 1.0000 | 0.7622 | 0.661 (Jaccard 0.629) | 0.101 |
| terra-none | 1.0000 | 0.9592 | 0.789 (Jaccard 0.752) | 0.171 |

[INFERENCE] On fastapi, prior areas often share no files with the next map. The ceiling is 0.6336 for haiku, 0.7787 for luna-xhigh and 0.7622 for terra-low, all below the 0.8 minimum, so no matcher, however good its notion of "same concern", can reach 80 % there; for those three arms C1 is unreachable on this candidate set before any threshold or Jev score is applied. Only terra-none on fastapi (0.9592) and every arm on axum (0.9819–1.0) have headroom above 0.8, and terra-none's Jev cell reaches 0.789 of 0.9592 on fastapi. The ceiling is a property of the pair of consecutive maps, not of the matching method, so the question of whether fastapi's maps are intrinsically unstable (map-to-map variance of the agent) is separate from the question this study asks of the matcher. It is read from the outputs; the explanation (that the maps redraw the area boundaries between runs, and area counts vary run to run, §5) is inferred from the counts of §5.

**Jev vs Jaccard (design question Q2):** the Jev method raised the minimum-over-repositories survival in every arm (0.411 → 0.547, 0.473 → 0.589, 0.629 → 0.661, 0.752 → 0.789) and never changed a verdict (all `fallback`). The rise is on fastapi, the binding repository; axum's survival at the best threshold moved by at most 0.045 (haiku 0.865 → 0.910, luna 0.936 → 0.954, terra-low 0.933 → 0.938, terra-none 0.905 → 0.918). Spurious was 0 for both methods in every arm. Jev scores the same 1,986 candidate pairs the labellers labelled; the labels, not Jev's score, decide the spurious rate (PREREGISTRATION §11).

**Scope limits (PREREGISTRATION §11):** one run per arm and repository; four arms; eight cells share labels; two repositories; haiku `none` is "the model default"; the coverage guard is a coarse proxy (haiku fails it, §6); the Jev time is an 8-worker service-time estimate (§4); the labels are agent labels from the mapper families, with Jev answering the same question from the same text, so Jev's agreement with the labels is an internal-consistency result, not proof of correctness; nothing below 0.30 was measured; the synthetic planted Jev scores in `validation/` test arithmetic and wiring, not Jev.

`<arm>+jev` rows: created/kept/retired and kept-inclusive survival are schema placeholders (kept = prior − matched, retired = 0), unmeasured; the inventory's 'not recorded' entries for `+jev` arms are the same placeholders (Jev cells run no maps).

## 2. Checklist: PREREGISTRATION §12 reporting contract → sections here

| §12 item | Section |
|---|---|
| Per cell: X2 tables, verdict, pass set and contiguity flag, margin, unlabelled share, confirmation, coverage ratio, refresh time, Jev cost/time | §3.1 (tables), §3.2 (cell summary) |
| Per cell: leave-one-transition-out; per-transition survival; strata; chained survival; label distribution; over-call check | §3.3–§3.6, §7.3 |
| Served checks and the Jev check with reasons; probe lines per arm | §4.1, §4.2, §9 |
| X2b's luna result beside the luna-xhigh cells; the selection line; HD5c consequence | §1, §3.2 |
| Labelled occurrences against unique pairs | §7.1 |
| Per (arm, repo) unscored ÷ candidate occurrences; per-refresh Jev cost, `seconds`, `seconds_sequential`, requests, transitions | §4.3, §4.4 |
| Jev run: answers, failure rows by kind, `cost_unknown_rows`, billed USD, attempts > 1 share, `foreign_model_failure_rows` | §4.2 |
| Per (arm, repo) map seconds n, mean, min, max and per map id | §8 |
| Candidate ceiling beside each Jev cell's C1; area ratio | §1, §6 |
| Jev code digests (two lines) | §11.3 |
| Cells examined and eligible; per-cell false-pass figures; sum of maxima | §5 |
| Labels: agreement, adjudication, served models, controls, integrity | §7 |
| Unmeasured items (mapper cost) | §10 |
| Run commit, digest equality, reproduce commands | §11, §13 |
| Deviations | §12 |
| Threats | §14 |

## 3. Survival and spurious merge by cell

### 3.1 Decisive stratum, per cell and repository (all thirteen thresholds)

Verbatim rows of `data/analysis_output.txt`. Columns: `prev` prior areas; `match` matches; `surv(mean)` mean of per-transition survival; `worstT` worst transition; `pooled [Wilson]` pooled matched/prior with Wilson 95 % (not the per-repository mean's interval); `crtd/retd/kept` created, retired, kept areas (schema placeholders for `+jev` rows); `Jmed` upper-median Jaccard of matches; `J<.8` share of matches with Jaccard < 0.8; `lab(occ/uniq)` labelled occurrences / unique labelled pairs; `diff`, `overl` the `different` and `overlapping` counts; `unlab(ct/unp/never)` unlabelled; `spur [Wilson]` spurious rate; `strict` and `worst` the strict and worst-case spurious rates. Below each block, the exact mean survival (`survival_exact`) per threshold. All other strata (remap, mutation, fresh, same_state) and the per-transition survival with kept-inclusive values at 0.30, 0.50, 0.70, 0.90 are in `data/analysis_output.txt`.

#### haiku (jaccard)

**haiku / fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   101    42  0.411  0.100  0.416 [0.325,0.513]    64    22    37 0.625 0.738 |    42/42       0    10     0 (0/0/0)  0.000 [0.000,0.084] 0.238 0.000
   0.35   101    38  0.365  0.083  0.376 [0.288,0.474]    68    23    40 0.636 0.711 |    38/38       0     7     0 (0/0/0)  0.000 [0.000,0.092] 0.184 0.000
   0.40   101    38  0.365  0.083  0.376 [0.288,0.474]    68    23    40 0.636 0.711 |    38/38       0     7     0 (0/0/0)  0.000 [0.000,0.092] 0.184 0.000
   0.45   101    32  0.307  0.050  0.317 [0.234,0.413]    74    26    43 0.714 0.656 |    32/32       0     4     0 (0/0/0)  0.000 [0.000,0.107] 0.125 0.000
   0.50   101    32  0.307  0.050  0.317 [0.234,0.413]    74    26    43 0.714 0.656 |    32/32       0     4     0 (0/0/0)  0.000 [0.000,0.107] 0.125 0.000
   0.55   101    27  0.266  0.050  0.267 [0.191,0.361]    79    27    47 0.733 0.593 |    27/27       0     4     0 (0/0/0)  0.000 [0.000,0.125] 0.148 0.000
   0.60   101    24  0.241  0.050  0.238 [0.165,0.329]    82    29    48 0.769 0.542 |    24/24       0     3     0 (0/0/0)  0.000 [0.000,0.138] 0.125 0.000
   0.65   101    18  0.177  0.000  0.178 [0.116,0.264]    88    32    51 0.909 0.389 |    18/18       0     1     0 (0/0/0)  0.000 [0.000,0.176] 0.056 0.000
   0.70   101    16  0.160  0.000  0.158 [0.100,0.242]    90    34    51 1.000 0.312 |    16/16       0     1     0 (0/0/0)  0.000 [0.000,0.194] 0.062 0.000
   0.75   101    13  0.134  0.000  0.129 [0.077,0.208]    93    37    51 1.000 0.154 |    13/13       0     1     0 (0/0/0)  0.000 [0.000,0.228] 0.077 0.000
   0.80   101    11  0.118  0.000  0.109 [0.062,0.185]    95    38    52 1.000 0.000 |    11/11       0     1     0 (0/0/0)  0.000 [0.000,0.259] 0.091 0.000
   0.85   101    10  0.109  0.000  0.099 [0.055,0.173]    96    38    53 1.000 0.000 |    10/10       0     0     0 (0/0/0)  0.000 [0.000,0.278] 0.000 0.000
   0.90   101    10  0.109  0.000  0.099 [0.055,0.173]    96    38    53 1.000 0.000 |    10/10       0     0     0 (0/0/0)  0.000 [0.000,0.278] 0.000 0.000
```
Exact mean survival: 0.30: 4879/11880, 0.35: 542/1485, 0.40: 542/1485, 0.45: 365/1188, 0.50: 365/1188, 0.55: 631/2376, 0.60: 1429/5940, 0.65: 1051/5940, 0.70: 238/1485, 0.75: 133/990, 0.80: 233/1980, 0.85: 433/3960, 0.90: 433/3960

**haiku / axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   115    99  0.865  0.783  0.861 [0.786,0.913]    34    13     3 0.667 0.646 |    99/99       0    59     0 (0/0/0)  0.000 [0.000,0.037] 0.596 0.000
   0.35   115    95  0.828  0.765  0.826 [0.747,0.885]    38    17     3 0.692 0.632 |    95/95       0    55     0 (0/0/0)  0.000 [0.000,0.039] 0.579 0.000
   0.40   115    90  0.784  0.737  0.783 [0.699,0.848]    43    22     3 0.714 0.611 |    90/90       0    50     0 (0/0/0)  0.000 [0.000,0.041] 0.556 0.000
   0.45   115    81  0.707  0.591  0.704 [0.615,0.780]    52    30     4 0.750 0.568 |    81/81       0    41     0 (0/0/0)  0.000 [0.000,0.045] 0.506 0.000
   0.50   115    78  0.680  0.591  0.678 [0.588,0.757]    55    32     5 0.778 0.551 |    78/78       0    38     0 (0/0/0)  0.000 [0.000,0.047] 0.487 0.000
   0.55   115    66  0.579  0.409  0.574 [0.483,0.660]    67    44     5 0.824 0.470 |    66/66       0    26     0 (0/0/0)  0.000 [0.000,0.055] 0.394 0.000
   0.60   115    61  0.532  0.409  0.530 [0.440,0.619]    72    47     7 0.833 0.426 |    61/61       0    23     0 (0/0/0)  0.000 [0.000,0.059] 0.377 0.000
   0.65   115    57  0.496  0.353  0.496 [0.406,0.586]    76    50     8 0.833 0.386 |    57/57       0    19     0 (0/0/0)  0.000 [0.000,0.063] 0.333 0.000
   0.70   115    47  0.409  0.294  0.409 [0.323,0.500]    86    57    11 0.900 0.255 |    47/47       0    12     0 (0/0/0)  0.000 [0.000,0.076] 0.255 0.000
   0.75   115    44  0.384  0.294  0.383 [0.299,0.474]    89    60    11 0.900 0.205 |    44/44       0     9     0 (0/0/0)  0.000 [0.000,0.080] 0.205 0.000
   0.80   115    35  0.308  0.227  0.304 [0.228,0.394]    98    69    11 1.000 0.000 |    35/35       0     3     0 (0/0/0)  0.000 [0.000,0.099] 0.086 0.000
   0.85   115    28  0.248  0.174  0.243 [0.174,0.329]   105    76    11 1.000 0.000 |    28/28       0     0     0 (0/0/0)  0.000 [0.000,0.121] 0.000 0.000
   0.90   115    25  0.219  0.118  0.217 [0.152,0.301]   108    79    11 1.000 0.000 |    25/25       0     0     0 (0/0/0)  0.000 [0.000,0.133] 0.000 0.000
```
Exact mean survival: 0.30: 282641/326876, 0.35: 270497/326876, 0.40: 384563/490314, 0.45: 231175/326876, 0.50: 222397/326876, 0.55: 189201/326876, 0.60: 522041/980628, 0.65: 162031/326876, 0.70: 66779/163438, 0.75: 94091/245157, 0.80: 100807/326876, 0.85: 20278/81719, 0.90: 35749/163438

#### haiku+jev (jev)

**haiku+jev / fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   101    56  0.547  0.167  0.554 [0.457,0.648]    50     0    45 0.995 0.071 |    56/56       0    23     0 (0/0/0)  0.000 [0.000,0.064] 0.411 0.000
   0.35   101    56  0.547  0.167  0.554 [0.457,0.648]    50     0    45 0.995 0.071 |    56/56       0    23     0 (0/0/0)  0.000 [0.000,0.064] 0.411 0.000
   0.40   101    56  0.547  0.167  0.554 [0.457,0.648]    50     0    45 0.995 0.071 |    56/56       0    23     0 (0/0/0)  0.000 [0.000,0.064] 0.411 0.000
   0.45   101    56  0.547  0.167  0.554 [0.457,0.648]    50     0    45 0.995 0.071 |    56/56       0    23     0 (0/0/0)  0.000 [0.000,0.064] 0.411 0.000
   0.50   101    56  0.547  0.167  0.554 [0.457,0.648]    50     0    45 0.995 0.071 |    56/56       0    23     0 (0/0/0)  0.000 [0.000,0.064] 0.411 0.000
   0.55   101    55  0.539  0.167  0.545 [0.448,0.638]    51     0    46 0.995 0.055 |    55/55       0    22     0 (0/0/0)  0.000 [0.000,0.065] 0.400 0.000
   0.60   101    55  0.539  0.167  0.545 [0.448,0.638]    51     0    46 0.995 0.055 |    55/55       0    22     0 (0/0/0)  0.000 [0.000,0.065] 0.400 0.000
   0.65   101    55  0.539  0.167  0.545 [0.448,0.638]    51     0    46 0.995 0.055 |    55/55       0    22     0 (0/0/0)  0.000 [0.000,0.065] 0.400 0.000
   0.70   101    53  0.517  0.167  0.525 [0.428,0.619]    53     0    48 0.995 0.019 |    53/53       0    20     0 (0/0/0)  0.000 [0.000,0.068] 0.377 0.000
   0.75   101    52  0.508  0.111  0.515 [0.419,0.610]    54     0    49 0.995 0.000 |    52/52       0    19     0 (0/0/0)  0.000 [0.000,0.069] 0.365 0.000
   0.80   101    52  0.508  0.111  0.515 [0.419,0.610]    54     0    49 0.995 0.000 |    52/52       0    19     0 (0/0/0)  0.000 [0.000,0.069] 0.365 0.000
   0.85   101    52  0.508  0.111  0.515 [0.419,0.610]    54     0    49 0.995 0.000 |    52/52       0    19     0 (0/0/0)  0.000 [0.000,0.069] 0.365 0.000
   0.90   101    52  0.508  0.111  0.515 [0.419,0.610]    54     0    49 0.995 0.000 |    52/52       0    19     0 (0/0/0)  0.000 [0.000,0.069] 0.365 0.000
```
Exact mean survival: 0.30: 271/495, 0.35: 271/495, 0.40: 271/495, 0.45: 271/495, 0.50: 271/495, 0.55: 427/792, 0.60: 427/792, 0.65: 427/792, 0.70: 2047/3960, 0.75: 6031/11880, 0.80: 6031/11880, 0.85: 6031/11880, 0.90: 6031/11880

**haiku+jev / axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   115   104  0.910  0.783  0.904 [0.837,0.946]    29     0    11 0.995 0.010 |   104/104      0    64     0 (0/0/0)  0.000 [0.000,0.036] 0.615 0.000
   0.35   115   104  0.910  0.783  0.904 [0.837,0.946]    29     0    11 0.995 0.010 |   104/104      0    64     0 (0/0/0)  0.000 [0.000,0.036] 0.615 0.000
   0.40   115   104  0.910  0.783  0.904 [0.837,0.946]    29     0    11 0.995 0.010 |   104/104      0    64     0 (0/0/0)  0.000 [0.000,0.036] 0.615 0.000
   0.45   115   104  0.910  0.783  0.904 [0.837,0.946]    29     0    11 0.995 0.010 |   104/104      0    64     0 (0/0/0)  0.000 [0.000,0.036] 0.615 0.000
   0.50   115   104  0.910  0.783  0.904 [0.837,0.946]    29     0    11 0.995 0.010 |   104/104      0    64     0 (0/0/0)  0.000 [0.000,0.036] 0.615 0.000
   0.55   115   104  0.910  0.783  0.904 [0.837,0.946]    29     0    11 0.995 0.010 |   104/104      0    64     0 (0/0/0)  0.000 [0.000,0.036] 0.615 0.000
   0.60   115   104  0.910  0.783  0.904 [0.837,0.946]    29     0    11 0.995 0.010 |   104/104      0    64     0 (0/0/0)  0.000 [0.000,0.036] 0.615 0.000
   0.65   115   104  0.910  0.783  0.904 [0.837,0.946]    29     0    11 0.995 0.010 |   104/104      0    64     0 (0/0/0)  0.000 [0.000,0.036] 0.615 0.000
   0.70   115   103  0.902  0.783  0.896 [0.826,0.939]    30     0    12 0.995 0.000 |   103/103      0    63     0 (0/0/0)  0.000 [0.000,0.036] 0.612 0.000
   0.75   115   103  0.902  0.783  0.896 [0.826,0.939]    30     0    12 0.995 0.000 |   103/103      0    63     0 (0/0/0)  0.000 [0.000,0.036] 0.612 0.000
   0.80   115   103  0.902  0.783  0.896 [0.826,0.939]    30     0    12 0.995 0.000 |   103/103      0    63     0 (0/0/0)  0.000 [0.000,0.036] 0.612 0.000
   0.85   115   102  0.892  0.783  0.887 [0.816,0.933]    31     0    13 0.995 0.000 |   102/102      0    62     0 (0/0/0)  0.000 [0.000,0.036] 0.608 0.000
   0.90   115   100  0.874  0.783  0.870 [0.796,0.919]    33     0    15 0.995 0.000 |   100/100      0    60     0 (0/0/0)  0.000 [0.000,0.037] 0.600 0.000
```
Exact mean survival: 0.30: 223199/245157, 0.35: 223199/245157, 0.40: 223199/245157, 0.45: 223199/245157, 0.50: 223199/245157, 0.55: 223199/245157, 0.60: 223199/245157, 0.65: 223199/245157, 0.70: 442097/490314, 0.75: 442097/490314, 0.80: 442097/490314, 0.85: 218645/245157, 0.90: 71448/81719

#### luna-xhigh (jaccard)

**luna-xhigh / fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   169    80  0.473  0.400  0.473 [0.400,0.548]    67    33    56 0.700 0.662 |    80/80       0    33     0 (0/0/0)  0.000 [0.000,0.046] 0.412 0.000
   0.35   169    70  0.407  0.280  0.414 [0.343,0.490]    77    35    64 0.750 0.614 |    70/70       0    25     0 (0/0/0)  0.000 [0.000,0.052] 0.357 0.000
   0.40   169    66  0.388  0.280  0.391 [0.320,0.466]    81    36    67 0.750 0.591 |    66/66       0    21     0 (0/0/0)  0.000 [0.000,0.055] 0.318 0.000
   0.45   169    59  0.342  0.160  0.349 [0.281,0.424]    88    37    73 0.762 0.542 |    59/59       0    15     0 (0/0/0)  0.000 [0.000,0.061] 0.254 0.000
   0.50   169    57  0.331  0.160  0.337 [0.270,0.411]    90    37    75 0.769 0.526 |    57/57       0    14     0 (0/0/0)  0.000 [0.000,0.063] 0.246 0.000
   0.55   169    48  0.275  0.080  0.284 [0.221,0.356]    99    41    80 0.880 0.438 |    48/48       0     8     0 (0/0/0)  0.000 [0.000,0.074] 0.167 0.000
   0.60   169    46  0.264  0.040  0.272 [0.211,0.344]   101    42    81 0.896 0.413 |    46/46       0     7     0 (0/0/0)  0.000 [0.000,0.077] 0.152 0.000
   0.65   169    43  0.250  0.040  0.254 [0.195,0.325]   104    44    82 0.909 0.372 |    43/43       0     6     0 (0/0/0)  0.000 [0.000,0.082] 0.140 0.000
   0.70   169    40  0.233  0.040  0.237 [0.179,0.306]   107    45    84 0.955 0.325 |    40/40       0     6     0 (0/0/0)  0.000 [0.000,0.088] 0.150 0.000
   0.75   169    36  0.201  0.040  0.213 [0.158,0.281]   111    48    85 0.968 0.250 |    36/36       0     5     0 (0/0/0)  0.000 [0.000,0.096] 0.139 0.000
   0.80   169    27  0.146  0.040  0.160 [0.112,0.222]   120    51    91 1.000 0.000 |    27/27       0     2     0 (0/0/0)  0.000 [0.000,0.125] 0.074 0.000
   0.85   169    24  0.131  0.040  0.142 [0.097,0.203]   123    53    92 1.000 0.000 |    24/24       0     1     0 (0/0/0)  0.000 [0.000,0.138] 0.042 0.000
   0.90   169    22  0.122  0.040  0.130 [0.088,0.189]   125    53    94 1.000 0.000 |    22/22       0     1     0 (0/0/0)  0.000 [0.000,0.149] 0.045 0.000
```
Exact mean survival: 0.30: 2383/5040, 0.35: 1139/2800, 0.40: 3257/8400, 0.45: 8627/25200, 0.50: 2083/6300, 0.55: 289/1050, 0.60: 277/1050, 0.65: 131/525, 0.70: 839/3600, 0.75: 1691/8400, 0.80: 131/900, 0.85: 827/6300, 0.90: 767/6300

**luna-xhigh / axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   474   442  0.936  0.805  0.932 [0.906,0.952]    65    20    12 0.714 0.534 |   442/442      0    41     0 (0/0/0)  0.000 [0.000,0.009] 0.093 0.000
   0.35   474   434  0.919  0.793  0.916 [0.887,0.937]    73    25    15 0.750 0.525 |   434/434      0    34     0 (0/0/0)  0.000 [0.000,0.009] 0.078 0.000
   0.40   474   432  0.914  0.793  0.911 [0.882,0.934]    75    27    15 0.750 0.523 |   432/432      0    32     0 (0/0/0)  0.000 [0.000,0.009] 0.074 0.000
   0.45   474   422  0.893  0.782  0.890 [0.859,0.915]    85    32    20 0.750 0.512 |   422/422      0    26     0 (0/0/0)  0.000 [0.000,0.009] 0.062 0.000
   0.50   474   422  0.893  0.782  0.890 [0.859,0.915]    85    32    20 0.750 0.512 |   422/422      0    26     0 (0/0/0)  0.000 [0.000,0.009] 0.062 0.000
   0.55   474   247  0.521  0.270  0.521 [0.476,0.566]   260    82   145 1.000 0.166 |   247/247      0    14     0 (0/0/0)  0.000 [0.000,0.015] 0.057 0.000
   0.60   474   246  0.519  0.270  0.519 [0.474,0.564]   261    83   145 1.000 0.163 |   246/246      0    14     0 (0/0/0)  0.000 [0.000,0.015] 0.057 0.000
   0.65   474   239  0.505  0.257  0.504 [0.459,0.549]   268    89   146 1.000 0.138 |   239/239      0     9     0 (0/0/0)  0.000 [0.000,0.016] 0.038 0.000
   0.70   474   222  0.469  0.216  0.468 [0.424,0.513]   285    99   153 1.000 0.072 |   222/222      0     1     0 (0/0/0)  0.000 [0.000,0.017] 0.005 0.000
   0.75   474   220  0.464  0.216  0.464 [0.420,0.509]   287   101   153 1.000 0.064 |   220/220      0     1     0 (0/0/0)  0.000 [0.000,0.017] 0.005 0.000
   0.80   474   206  0.436  0.216  0.435 [0.391,0.480]   301   105   163 1.000 0.000 |   206/206      0     0     0 (0/0/0)  0.000 [0.000,0.018] 0.000 0.000
   0.85   474   201  0.425  0.203  0.424 [0.380,0.469]   306   106   167 1.000 0.000 |   201/201      0     0     0 (0/0/0)  0.000 [0.000,0.019] 0.000 0.000
   0.90   474   199  0.421  0.203  0.420 [0.376,0.465]   308   106   169 1.000 0.000 |   199/199      0     0     0 (0/0/0)  0.000 [0.000,0.019] 0.000 0.000
```
Exact mean survival: 0.30: 3416857/3650346, 0.35: 3354271/3650346, 0.40: 1668914/1825173, 0.45: 13042649/14601384, 0.50: 13042649/14601384, 0.55: 87463/167832, 0.60: 43565/83916, 0.65: 272869/540792, 0.70: 855782/1825173, 0.75: 1695121/3650346, 0.80: 3180503/7300692, 0.85: 3104623/7300692, 0.90: 3071737/7300692

#### luna-xhigh+jev (jev)

**luna-xhigh+jev / fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   169   100  0.589  0.480  0.592 [0.516,0.663]    47     0    69 0.995 0.010 |   100/100      0    49     0 (0/0/0)  0.000 [0.000,0.037] 0.490 0.000
   0.35   169   100  0.589  0.480  0.592 [0.516,0.663]    47     0    69 0.995 0.010 |   100/100      0    49     0 (0/0/0)  0.000 [0.000,0.037] 0.490 0.000
   0.40   169   100  0.589  0.480  0.592 [0.516,0.663]    47     0    69 0.995 0.010 |   100/100      0    49     0 (0/0/0)  0.000 [0.000,0.037] 0.490 0.000
   0.45   169   100  0.589  0.480  0.592 [0.516,0.663]    47     0    69 0.995 0.010 |   100/100      0    49     0 (0/0/0)  0.000 [0.000,0.037] 0.490 0.000
   0.50   169   100  0.589  0.480  0.592 [0.516,0.663]    47     0    69 0.995 0.010 |   100/100      0    49     0 (0/0/0)  0.000 [0.000,0.037] 0.490 0.000
   0.55   169   100  0.589  0.480  0.592 [0.516,0.663]    47     0    69 0.995 0.010 |   100/100      0    49     0 (0/0/0)  0.000 [0.000,0.037] 0.490 0.000
   0.60   169   100  0.589  0.480  0.592 [0.516,0.663]    47     0    69 0.995 0.010 |   100/100      0    49     0 (0/0/0)  0.000 [0.000,0.037] 0.490 0.000
   0.65   169   100  0.589  0.480  0.592 [0.516,0.663]    47     0    69 0.995 0.010 |   100/100      0    49     0 (0/0/0)  0.000 [0.000,0.037] 0.490 0.000
   0.70   169   100  0.589  0.480  0.592 [0.516,0.663]    47     0    69 0.995 0.010 |   100/100      0    49     0 (0/0/0)  0.000 [0.000,0.037] 0.490 0.000
   0.75   169   100  0.589  0.480  0.592 [0.516,0.663]    47     0    69 0.995 0.010 |   100/100      0    49     0 (0/0/0)  0.000 [0.000,0.037] 0.490 0.000
   0.80   169    99  0.584  0.480  0.586 [0.510,0.657]    48     0    70 0.995 0.000 |    99/99       0    48     0 (0/0/0)  0.000 [0.000,0.037] 0.485 0.000
   0.85   169    99  0.584  0.480  0.586 [0.510,0.657]    48     0    70 0.995 0.000 |    99/99       0    48     0 (0/0/0)  0.000 [0.000,0.037] 0.485 0.000
   0.90   169    98  0.579  0.480  0.580 [0.505,0.652]    49     0    71 0.995 0.000 |    98/98       0    47     0 (0/0/0)  0.000 [0.000,0.038] 0.480 0.000
```
Exact mean survival: 0.30: 1649/2800, 0.35: 1649/2800, 0.40: 1649/2800, 0.45: 1649/2800, 0.50: 1649/2800, 0.55: 1649/2800, 0.60: 1649/2800, 0.65: 1649/2800, 0.70: 1649/2800, 0.75: 1649/2800, 0.80: 701/1200, 0.85: 701/1200, 0.90: 4867/8400

**luna-xhigh+jev / axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.35   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.40   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.45   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.50   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.55   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.60   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.65   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.70   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.75   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.80   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.85   474   451  0.954  0.874  0.951 [0.928,0.967]    56     0    23 1.000 0.000 |   451/451      0    47     0 (0/0/0)  0.000 [0.000,0.008] 0.104 0.000
   0.90   474   450  0.952  0.874  0.949 [0.926,0.966]    57     0    24 1.000 0.000 |   450/450      0    46     0 (0/0/0)  0.000 [0.000,0.008] 0.102 0.000
```
Exact mean survival: 0.30: 1160923/1216782, 0.35: 1160923/1216782, 0.40: 1160923/1216782, 0.45: 1160923/1216782, 0.50: 1160923/1216782, 0.55: 1160923/1216782, 0.60: 1160923/1216782, 0.65: 1160923/1216782, 0.70: 1160923/1216782, 0.75: 1160923/1216782, 0.80: 1160923/1216782, 0.85: 1160923/1216782, 0.90: 1737629/1825173

#### terra-low (jaccard)

**terra-low / fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   129    80  0.629  0.440  0.620 [0.534,0.699]    43     6    43 0.882 0.500 |    80/80       0    25     0 (0/0/0)  0.000 [0.000,0.046] 0.312 0.000
   0.35   129    76  0.595  0.400  0.589 [0.503,0.670]    47     6    47 1.000 0.474 |    76/76       0    21     0 (0/0/0)  0.000 [0.000,0.048] 0.276 0.000
   0.40   129    70  0.555  0.360  0.543 [0.457,0.626]    53     9    50 1.000 0.429 |    70/70       0    17     0 (0/0/0)  0.000 [0.000,0.052] 0.243 0.000
   0.45   129    67  0.526  0.360  0.519 [0.434,0.604]    56     9    53 1.000 0.403 |    67/67       0    15     0 (0/0/0)  0.000 [0.000,0.054] 0.224 0.000
   0.50   129    66  0.520  0.360  0.512 [0.426,0.596]    57    10    53 1.000 0.394 |    66/66       0    14     0 (0/0/0)  0.000 [0.000,0.055] 0.212 0.000
   0.55   129    59  0.465  0.320  0.457 [0.374,0.543]    64    14    56 1.000 0.322 |    59/59       0    10     0 (0/0/0)  0.000 [0.000,0.061] 0.169 0.000
   0.60   129    55  0.430  0.320  0.426 [0.344,0.513]    68    15    59 1.000 0.273 |    55/55       0     7     0 (0/0/0)  0.000 [0.000,0.065] 0.127 0.000
   0.65   129    48  0.376  0.280  0.372 [0.294,0.458]    75    16    65 1.000 0.167 |    48/48       0     4     0 (0/0/0)  0.000 [0.000,0.074] 0.083 0.000
   0.70   129    46  0.362  0.280  0.357 [0.279,0.442]    77    16    67 1.000 0.130 |    46/46       0     3     0 (0/0/0)  0.000 [0.000,0.077] 0.065 0.000
   0.75   129    43  0.338  0.240  0.333 [0.258,0.418]    80    17    69 1.000 0.070 |    43/43       0     2     0 (0/0/0)  0.000 [0.000,0.082] 0.047 0.000
   0.80   129    40  0.317  0.238  0.310 [0.237,0.394]    83    19    70 1.000 0.000 |    40/40       0     0     0 (0/0/0)  0.000 [0.000,0.088] 0.000 0.000
   0.85   129    40  0.317  0.238  0.310 [0.237,0.394]    83    19    70 1.000 0.000 |    40/40       0     0     0 (0/0/0)  0.000 [0.000,0.088] 0.000 0.000
   0.90   129    39  0.310  0.200  0.302 [0.230,0.386]    84    19    71 1.000 0.000 |    39/39       0     0     0 (0/0/0)  0.000 [0.000,0.090] 0.000 0.000
```
Exact mean survival: 0.30: 134663/214200, 0.35: 509903/856800, 0.40: 475631/856800, 0.45: 50109/95200, 0.50: 148423/285600, 0.55: 49787/107100, 0.60: 52637/122400, 0.65: 23041/61200, 0.70: 51677/142800, 0.75: 96571/285600, 0.80: 271489/856800, 0.85: 271489/856800, 0.90: 265777/856800

**terra-low / axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   192   179  0.933  0.879  0.932 [0.888,0.960]     9    10     3 0.875 0.374 |   179/179      0    60     0 (0/0/0)  0.000 [0.000,0.021] 0.335 0.000
   0.35   192   178  0.928  0.879  0.927 [0.881,0.956]    10    11     3 0.889 0.371 |   178/178      0    59     0 (0/0/0)  0.000 [0.000,0.021] 0.331 0.000
   0.40   192   177  0.923  0.879  0.922 [0.875,0.952]    11    12     3 0.889 0.367 |   177/177      0    58     0 (0/0/0)  0.000 [0.000,0.021] 0.328 0.000
   0.45   192   170  0.887  0.818  0.885 [0.833,0.923]    18    13     9 0.909 0.341 |   170/170      0    54     0 (0/0/0)  0.000 [0.000,0.022] 0.318 0.000
   0.50   192   169  0.881  0.818  0.880 [0.827,0.919]    19    14     9 0.909 0.337 |   169/169      0    53     0 (0/0/0)  0.000 [0.000,0.022] 0.314 0.000
   0.55   192   158  0.823  0.774  0.823 [0.763,0.870]    30    22    12 1.000 0.291 |   158/158      0    43     0 (0/0/0)  0.000 [0.000,0.024] 0.272 0.000
   0.60   192   154  0.802  0.710  0.802 [0.740,0.852]    34    23    15 1.000 0.273 |   154/154      0    39     0 (0/0/0)  0.000 [0.000,0.024] 0.253 0.000
   0.65   192   147  0.766  0.710  0.766 [0.701,0.820]    41    29    16 1.000 0.238 |   147/147      0    32     0 (0/0/0)  0.000 [0.000,0.025] 0.218 0.000
   0.70   192   119  0.620  0.484  0.620 [0.549,0.685]    69    47    26 1.000 0.059 |   119/119      0    13     0 (0/0/0)  0.000 [0.000,0.031] 0.109 0.000
   0.75   192   116  0.605  0.484  0.604 [0.534,0.671]    72    50    26 1.000 0.034 |   116/116      0    11     0 (0/0/0)  0.000 [0.000,0.032] 0.095 0.000
   0.80   192   112  0.583  0.452  0.583 [0.513,0.651]    76    54    26 1.000 0.000 |   112/112      0     8     0 (0/0/0)  0.000 [0.000,0.033] 0.071 0.000
   0.85   192   103  0.536  0.419  0.536 [0.466,0.606]    85    61    28 1.000 0.000 |   103/103      0     4     0 (0/0/0)  0.000 [0.000,0.036] 0.039 0.000
   0.90   192    86  0.448  0.323  0.448 [0.379,0.519]   102    72    34 1.000 0.000 |    86/86       0     1     0 (0/0/0)  0.000 [0.000,0.043] 0.012 0.000
```
Exact mean survival: 0.30: 15103/16182, 0.35: 7508/8091, 0.40: 14929/16182, 0.45: 78901/89001, 0.50: 156779/178002, 0.55: 73267/89001, 0.60: 71411/89001, 0.65: 136405/178002, 0.70: 12259/19778, 0.75: 17939/29667, 0.80: 51928/89001, 0.85: 95459/178002, 0.90: 39877/89001

#### terra-low+jev (jev)

**terra-low+jev / fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.35   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.40   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.45   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.50   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.55   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.60   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.65   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.70   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.75   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.80   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.85   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
   0.90   129    84  0.661  0.440  0.651 [0.566,0.728]    39     0    45 1.000 0.000 |    84/84       0    29     0 (0/0/0)  0.000 [0.000,0.044] 0.345 0.000
```
Exact mean survival: 0.30: 47197/71400, 0.35: 47197/71400, 0.40: 47197/71400, 0.45: 47197/71400, 0.50: 47197/71400, 0.55: 47197/71400, 0.60: 47197/71400, 0.65: 47197/71400, 0.70: 47197/71400, 0.75: 47197/71400, 0.80: 47197/71400, 0.85: 47197/71400, 0.90: 47197/71400

**terra-low+jev / axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   192   180  0.938  0.879  0.938 [0.894,0.964]     8     0    12 1.000 0.017 |   180/180      0    61     0 (0/0/0)  0.000 [0.000,0.021] 0.339 0.000
   0.35   192   180  0.938  0.879  0.938 [0.894,0.964]     8     0    12 1.000 0.017 |   180/180      0    61     0 (0/0/0)  0.000 [0.000,0.021] 0.339 0.000
   0.40   192   180  0.938  0.879  0.938 [0.894,0.964]     8     0    12 1.000 0.017 |   180/180      0    61     0 (0/0/0)  0.000 [0.000,0.021] 0.339 0.000
   0.45   192   180  0.938  0.879  0.938 [0.894,0.964]     8     0    12 1.000 0.017 |   180/180      0    61     0 (0/0/0)  0.000 [0.000,0.021] 0.339 0.000
   0.50   192   180  0.938  0.879  0.938 [0.894,0.964]     8     0    12 1.000 0.017 |   180/180      0    61     0 (0/0/0)  0.000 [0.000,0.021] 0.339 0.000
   0.55   192   179  0.933  0.879  0.932 [0.888,0.960]     9     0    13 1.000 0.011 |   179/179      0    60     0 (0/0/0)  0.000 [0.000,0.021] 0.335 0.000
   0.60   192   179  0.933  0.879  0.932 [0.888,0.960]     9     0    13 1.000 0.011 |   179/179      0    60     0 (0/0/0)  0.000 [0.000,0.021] 0.335 0.000
   0.65   192   179  0.933  0.879  0.932 [0.888,0.960]     9     0    13 1.000 0.011 |   179/179      0    60     0 (0/0/0)  0.000 [0.000,0.021] 0.335 0.000
   0.70   192   179  0.933  0.879  0.932 [0.888,0.960]     9     0    13 1.000 0.011 |   179/179      0    60     0 (0/0/0)  0.000 [0.000,0.021] 0.335 0.000
   0.75   192   178  0.928  0.879  0.927 [0.881,0.956]    10     0    14 1.000 0.006 |   178/178      0    59     0 (0/0/0)  0.000 [0.000,0.021] 0.331 0.000
   0.80   192   177  0.923  0.879  0.922 [0.875,0.952]    11     0    15 1.000 0.000 |   177/177      0    58     0 (0/0/0)  0.000 [0.000,0.021] 0.328 0.000
   0.85   192   177  0.923  0.879  0.922 [0.875,0.952]    11     0    15 1.000 0.000 |   177/177      0    58     0 (0/0/0)  0.000 [0.000,0.021] 0.328 0.000
   0.90   192   176  0.918  0.879  0.917 [0.869,0.948]    12     0    16 1.000 0.000 |   176/176      0    57     0 (0/0/0)  0.000 [0.000,0.021] 0.324 0.000
```
Exact mean survival: 0.30: 83516/89001, 0.35: 83516/89001, 0.40: 83516/89001, 0.45: 83516/89001, 0.50: 83516/89001, 0.55: 166075/178002, 0.60: 166075/178002, 0.65: 166075/178002, 0.70: 166075/178002, 0.75: 7508/8091, 0.80: 18253/19778, 0.85: 18253/19778, 0.90: 81689/89001

#### terra-none (jaccard)

**terra-none / fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   231   166  0.752  0.559  0.719 [0.657,0.773]    91    16    49 0.762 0.524 |   166/166      0    66     0 (0/0/0)  0.000 [0.000,0.023] 0.398 0.000
   0.35   231   159  0.722  0.529  0.688 [0.626,0.745]    98    17    55 0.778 0.503 |   159/159      0    59     0 (0/0/0)  0.000 [0.000,0.024] 0.371 0.000
   0.40   231   148  0.663  0.529  0.641 [0.577,0.700]   109    19    64 0.818 0.466 |   148/148      0    51     0 (0/0/0)  0.000 [0.000,0.025] 0.345 0.000
   0.45   231   140  0.630  0.485  0.606 [0.542,0.667]   117    19    72 0.833 0.436 |   140/140      0    45     0 (0/0/0)  0.000 [0.000,0.027] 0.321 0.000
   0.50   231   139  0.624  0.485  0.602 [0.537,0.663]   118    19    73 0.833 0.432 |   139/139      0    44     0 (0/0/0)  0.000 [0.000,0.027] 0.317 0.000
   0.55   231   131  0.589  0.456  0.567 [0.503,0.629]   126    22    78 0.882 0.397 |   131/131      0    38     0 (0/0/0)  0.000 [0.000,0.028] 0.290 0.000
   0.60   231   126  0.567  0.441  0.545 [0.481,0.608]   131    23    82 0.900 0.373 |   126/126      0    34     0 (0/0/0)  0.000 [0.000,0.030] 0.270 0.000
   0.65   231   107  0.478  0.310  0.463 [0.400,0.528]   150    28    96 1.000 0.262 |   107/107      0    22     0 (0/0/0)  0.000 [0.000,0.035] 0.206 0.000
   0.70   231    94  0.418  0.310  0.407 [0.346,0.471]   163    29   108 1.000 0.160 |    94/94       0    11     0 (0/0/0)  0.000 [0.000,0.039] 0.117 0.000
   0.75   231    88  0.393  0.310  0.381 [0.321,0.445]   169    35   108 1.000 0.102 |    88/88       0     9     0 (0/0/0)  0.000 [0.000,0.042] 0.102 0.000
   0.80   231    79  0.351  0.241  0.342 [0.284,0.405]   178    37   115 1.000 0.000 |    79/79       0     3     0 (0/0/0)  0.000 [0.000,0.046] 0.038 0.000
   0.85   231    69  0.308  0.207  0.299 [0.243,0.361]   188    43   119 1.000 0.000 |    69/69       0     1     0 (0/0/0)  0.000 [0.000,0.053] 0.014 0.000
   0.90   231    64  0.287  0.172  0.277 [0.223,0.338]   193    44   123 1.000 0.000 |    64/64       0     0     0 (0/0/0)  0.000 [0.000,0.057] 0.000 0.000
```
Exact mean survival: 0.30: 676661/899232, 0.35: 432727/599488, 0.40: 596603/899232, 0.45: 566089/899232, 0.50: 1122811/1798464, 0.55: 55727/94656, 0.60: 339965/599488, 0.65: 859871/1798464, 0.70: 11734/28101, 0.75: 706171/1798464, 0.80: 631327/1798464, 0.85: 2711/8816, 0.90: 516637/1798464

**terra-none / axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   225   203  0.905  0.821  0.902 [0.856,0.935]    15    13     9 1.000 0.384 |   203/203      0    52     0 (0/0/0)  0.000 [0.000,0.019] 0.256 0.000
   0.35   225   198  0.883  0.769  0.880 [0.831,0.916]    20    18     9 1.000 0.369 |   198/198      0    47     0 (0/0/0)  0.000 [0.000,0.019] 0.237 0.000
   0.40   225   194  0.865  0.769  0.862 [0.811,0.901]    24    21    10 1.000 0.356 |   194/194      0    43     0 (0/0/0)  0.000 [0.000,0.019] 0.222 0.000
   0.45   225   185  0.824  0.692  0.822 [0.767,0.867]    33    27    13 1.000 0.324 |   185/185      0    37     0 (0/0/0)  0.000 [0.000,0.020] 0.200 0.000
   0.50   225   184  0.820  0.667  0.818 [0.762,0.863]    34    28    13 1.000 0.321 |   184/184      0    36     0 (0/0/0)  0.000 [0.000,0.020] 0.196 0.000
   0.55   225   163  0.727  0.538  0.724 [0.663,0.779]    55    44    18 1.000 0.233 |   163/163      0    22     0 (0/0/0)  0.000 [0.000,0.023] 0.135 0.000
   0.60   225   157  0.700  0.513  0.698 [0.635,0.754]    61    48    20 1.000 0.204 |   157/157      0    17     0 (0/0/0)  0.000 [0.000,0.024] 0.108 0.000
   0.65   225   150  0.669  0.487  0.667 [0.603,0.725]    68    52    23 1.000 0.167 |   150/150      0    13     0 (0/0/0)  0.000 [0.000,0.025] 0.087 0.000
   0.70   225   135  0.602  0.385  0.600 [0.535,0.662]    83    59    31 1.000 0.074 |   135/135      0    11     0 (0/0/0)  0.000 [0.000,0.028] 0.081 0.000
   0.75   225   134  0.598  0.385  0.596 [0.530,0.658]    84    59    32 1.000 0.067 |   134/134      0    10     0 (0/0/0)  0.000 [0.000,0.028] 0.075 0.000
   0.80   225   125  0.559  0.333  0.556 [0.490,0.619]    93    67    33 1.000 0.000 |   125/125      0     5     0 (0/0/0)  0.000 [0.000,0.030] 0.040 0.000
   0.85   225   116  0.518  0.333  0.516 [0.451,0.580]   102    72    37 1.000 0.000 |   116/116      0     1     0 (0/0/0)  0.000 [0.000,0.032] 0.009 0.000
   0.90   225   102  0.455  0.256  0.453 [0.390,0.519]   116    81    42 1.000 0.000 |   102/102      0     0     0 (0/0/0)  0.000 [0.000,0.036] 0.000 0.000
```
Exact mean survival: 0.30: 412/455, 0.35: 6829/7735, 0.40: 18521/21420, 0.45: 229577/278460, 0.50: 76129/92820, 0.55: 13495/18564, 0.60: 7499/10710, 0.65: 31049/46410, 0.70: 5986/9945, 0.75: 11887/19890, 0.80: 38893/69615, 0.85: 2404/4641, 0.90: 21137/46410

#### terra-none+jev (jev)

**terra-none+jev / fastapi**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.35   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.40   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.45   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.50   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.55   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.60   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.65   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.70   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.75   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.80   231   173  0.789  0.559  0.749 [0.689,0.800]    84     0    58 0.995 0.000 |   173/173      0    73     0 (0/0/0)  0.000 [0.000,0.022] 0.422 0.000
   0.85   231   172  0.783  0.559  0.745 [0.685,0.797]    85     0    59 0.995 0.000 |   172/172      0    72     0 (0/0/0)  0.000 [0.000,0.022] 0.419 0.000
   0.90   231   172  0.783  0.559  0.745 [0.685,0.797]    85     0    59 0.995 0.000 |   172/172      0    72     0 (0/0/0)  0.000 [0.000,0.022] 0.419 0.000
```
Exact mean survival: 0.30: 1418381/1798464, 0.35: 1418381/1798464, 0.40: 1418381/1798464, 0.45: 1418381/1798464, 0.50: 1418381/1798464, 0.55: 1418381/1798464, 0.60: 1418381/1798464, 0.65: 1418381/1798464, 0.70: 1418381/1798464, 0.75: 1418381/1798464, 0.80: 1418381/1798464, 0.85: 704507/899232, 0.90: 704507/899232

**terra-none+jev / axum**
```
   t     prev  match  surv(mean)  worstT pooled [Wilson]        crtd  retd  kept  Jmed  J<.8 | lab(occ/uniq) diff overl unlab(ct/unp/never)  spur  [Wilson]        strict worst
   0.30   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.35   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.40   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.45   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.50   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.55   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.60   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.65   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.70   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.75   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.80   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.85   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
   0.90   225   206  0.918  0.872  0.916 [0.872,0.945]    12     0    19 1.000 0.000 |   206/206      0    55     0 (0/0/0)  0.000 [0.000,0.018] 0.267 0.000
```
Exact mean survival: 0.30: 2507/2730, 0.35: 2507/2730, 0.40: 2507/2730, 0.45: 2507/2730, 0.50: 2507/2730, 0.55: 2507/2730, 0.60: 2507/2730, 0.65: 2507/2730, 0.70: 2507/2730, 0.75: 2507/2730, 0.80: 2507/2730, 0.85: 2507/2730, 0.90: 2507/2730

### 3.2 Cell summary (`results.json` `x2c.cells` and `arms.<cell>.decision`)

| cell | verdict | state | t* | margin | pass set | contiguous | feasible / imprecise | tripped gate | citable | confirmation | unlabelled share at 0.30 (axum, fastapi) | area ratio (axum, fastapi) | refresh time (s) | Jev cost USD / s (÷ 8) / s sequential |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| haiku | fallback | not-adopted | None | None | {} (empty) | True | [] / [] | None | True | None | 0.0, 0.0 | 0.2528, 0.6551 | 174.843 | – |
| haiku+jev | fallback | not-adopted | None | None | {} (empty) | True | [] / [] | None | True | None | 0.0, 0.0 | 0.2528, 0.6551 | 178.434 | 0.0055 / 3.591 / 28.726 |
| luna-xhigh | fallback | not-adopted | None | None | {} (empty) | True | [] / [] | None | True | None | 0.0, 0.0 | 1.0000, 1.0000 | 349.000 | – |
| luna-xhigh+jev | fallback | not-adopted | None | None | {} (empty) | True | [] / [] | None | True | None | 0.0, 0.0 | 1.0000, 1.0000 | 357.804 | 0.0123 / 8.804 / 70.434 |
| terra-low | fallback | not-adopted | None | None | {} (empty) | True | [] / [] | None | True | None | 0.0, 0.0 | 0.3874, 0.7975 | 146.271 | – |
| terra-low+jev | fallback | not-adopted | None | None | {} (empty) | True | [] / [] | None | True | None | 0.0, 0.0 | 0.3874, 0.7975 | 150.994 | 0.0069 / 4.723 / 37.782 |
| terra-none | fallback | not-adopted | None | None | {} (empty) | True | [] / [] | None | True | None | 0.0, 0.0 | 0.4516, 1.5443 | 247.814 | – |
| terra-none+jev | fallback | not-adopted | None | None | {} (empty) | True | [] / [] | None | True | None | 0.0, 0.0 | 0.4516, 1.5443 | 255.424 | 0.0117 / 7.609 / 60.876 |

Pass set empty in every cell (every `per_threshold` status is `fail`, 13 of 13); the contiguity flag is `true` for an empty set. `citable: true` in every cell (served checks ok for all four arms, Jev check ok; see §4 and the probe note in §9, which does not change this). Gates: class gate ok (remap and mutation measured in both repositories in every cell), label gate ok (unlabelled share 0.0), control gate ok (§7.2); no cell is `inconclusive`. `confirmation` is `null` because it applies only to an `adopt` cell. Refresh time is the arm's map wall time per refresh (sum of the mean map seconds over both repositories, §8) plus, for Jev cells, the Jev time per refresh at 8 workers. Per-refresh Jev cost and time by (arm, repo) are in §4.4.

X2b's luna `xhigh` Jaccard result (`fallback`, `../x2b-area-stability-luna/RESULTS.md`) beside the luna-xhigh cells: X2b fastapi mean survival at 0.30 was 0.5721 and axum 0.9102; here luna-xhigh Jaccard at 0.30 is fastapi 0.4728, axum 0.9360 (§1). Both `fallback`; `x2b_comparison.agree` true. [INFERENCE] Both studies draw one run of the same mapper per repository; the fastapi difference (0.5721 vs 0.4728) is of the size §5 shows the maps differ between runs, not evidence of a change in the mapper.

### 3.3 Leave-one-transition-out (joint t*, descriptive; not out-of-sample)

For 45 of the 48 rows (every cell × held-out transition T1–T6), `t* = None` per repository and held-out survival, matched, labelled and `different` are `null`: no threshold passes on the other five transitions. The three exceptions are all `terra-none+jev`: holding out T4 selects t* 0.90 (held-out survival axum 0.9714, fastapi 0.5588), holding out T5 selects 0.90 (axum 1.0, fastapi 0.6053), holding out T6 selects 0.80 (axum 0.8974, fastapi 0.7241). Each still fails C1 on fastapi, so the full-data verdict stands (`results.json` `arms.terra-none+jev.loto`; this paragraph corrected 2026-10-29: its first version stated that all 48 rows were `null`).

### 3.4 Per-transition survival, decisive stratum

Cells: `matched/prior = survival`. T1, T2 = F1→F2, F1→F3 (remap/fresh); T3, T4 = F1→R1, R1→R2 (remap/same_state); T5, T6 = R2→M1, M1→M2 (mutation).

**At threshold 0.30**

| cell | repo | T1 | T2 | T3 | T4 | T5 | T6 | mean (exact) |
|---|---|---|---|---|---|---|---|---|
| haiku | fastapi | 16/20 = 0.800 | 13/20 = 0.650 | 2/20 = 0.100 | 7/11 = 0.636 | 2/12 = 0.167 | 2/18 = 0.111 | 4879/11880 = 0.4107 |
| haiku | axum | 16/17 = 0.941 | 15/17 = 0.882 | 14/17 = 0.824 | 17/19 = 0.895 | 19/22 = 0.864 | 18/23 = 0.783 | 282641/326876 = 0.8647 |
| haiku+jev | fastapi | 20/20 = 1.000 | 16/20 = 0.800 | 5/20 = 0.250 | 9/11 = 0.818 | 3/12 = 0.250 | 3/18 = 0.167 | 271/495 = 0.5475 |
| haiku+jev | axum | 16/17 = 0.941 | 16/17 = 0.941 | 16/17 = 0.941 | 18/19 = 0.947 | 20/22 = 0.909 | 18/23 = 0.783 | 223199/245157 = 0.9104 |
| luna-xhigh | fastapi | 21/35 = 0.600 | 14/35 = 0.400 | 15/35 = 0.429 | 10/25 = 0.400 | 7/15 = 0.467 | 13/24 = 0.542 | 2383/5040 = 0.4728 |
| luna-xhigh | axum | 71/74 = 0.959 | 73/74 = 0.986 | 73/74 = 0.986 | 77/81 = 0.951 | 78/84 = 0.929 | 70/87 = 0.805 | 3416857/3650346 = 0.9360 |
| luna-xhigh+jev | fastapi | 23/35 = 0.657 | 18/35 = 0.514 | 23/35 = 0.657 | 12/25 = 0.480 | 9/15 = 0.600 | 15/24 = 0.625 | 1649/2800 = 0.5889 |
| luna-xhigh+jev | axum | 73/74 = 0.986 | 73/74 = 0.986 | 73/74 = 0.986 | 78/81 = 0.963 | 78/84 = 0.929 | 76/87 = 0.874 | 1160923/1216782 = 0.9541 |
| terra-low | fastapi | 11/25 = 0.440 | 20/25 = 0.800 | 15/25 = 0.600 | 10/21 = 0.476 | 12/16 = 0.750 | 12/17 = 0.706 | 134663/214200 = 0.6287 |
| terra-low | axum | 30/33 = 0.909 | 31/33 = 0.939 | 29/33 = 0.879 | 28/29 = 0.966 | 30/31 = 0.968 | 31/33 = 0.939 | 15103/16182 = 0.9333 |
| terra-low+jev | fastapi | 11/25 = 0.440 | 20/25 = 0.800 | 16/25 = 0.640 | 12/21 = 0.571 | 12/16 = 0.750 | 13/17 = 0.765 | 47197/71400 = 0.6610 |
| terra-low+jev | axum | 31/33 = 0.939 | 31/33 = 0.939 | 29/33 = 0.879 | 28/29 = 0.966 | 30/31 = 0.968 | 31/33 = 0.939 | 83516/89001 = 0.9384 |
| terra-none | fastapi | 28/32 = 0.875 | 28/32 = 0.875 | 30/32 = 0.938 | 38/68 = 0.559 | 22/38 = 0.579 | 20/29 = 0.690 | 676661/899232 = 0.7525 |
| terra-none | axum | 34/39 = 0.872 | 32/39 = 0.821 | 34/39 = 0.872 | 34/35 = 0.971 | 34/34 = 1.000 | 35/39 = 0.897 | 412/455 = 0.9055 |
| terra-none+jev | fastapi | 29/32 = 0.906 | 30/32 = 0.938 | 32/32 = 1.000 | 38/68 = 0.559 | 23/38 = 0.605 | 21/29 = 0.724 | 1418381/1798464 = 0.7887 |
| terra-none+jev | axum | 34/39 = 0.872 | 35/39 = 0.897 | 34/39 = 0.872 | 34/35 = 0.971 | 34/34 = 1.000 | 35/39 = 0.897 | 2507/2730 = 0.9183 |

**At threshold 0.50**

| cell | repo | T1 | T2 | T3 | T4 | T5 | T6 | mean (exact) |
|---|---|---|---|---|---|---|---|---|
| haiku | fastapi | 14/20 = 0.700 | 10/20 = 0.500 | 1/20 = 0.050 | 5/11 = 0.455 | 1/12 = 0.083 | 1/18 = 0.056 | 365/1188 = 0.3072 |
| haiku | axum | 11/17 = 0.647 | 12/17 = 0.706 | 12/17 = 0.706 | 14/19 = 0.737 | 13/22 = 0.591 | 16/23 = 0.696 | 222397/326876 = 0.6804 |
| haiku+jev | fastapi | 20/20 = 1.000 | 16/20 = 0.800 | 5/20 = 0.250 | 9/11 = 0.818 | 3/12 = 0.250 | 3/18 = 0.167 | 271/495 = 0.5475 |
| haiku+jev | axum | 16/17 = 0.941 | 16/17 = 0.941 | 16/17 = 0.941 | 18/19 = 0.947 | 20/22 = 0.909 | 18/23 = 0.783 | 223199/245157 = 0.9104 |
| luna-xhigh | fastapi | 15/35 = 0.429 | 13/35 = 0.371 | 9/35 = 0.257 | 4/25 = 0.160 | 4/15 = 0.267 | 12/24 = 0.500 | 2083/6300 = 0.3306 |
| luna-xhigh | axum | 65/74 = 0.878 | 72/74 = 0.973 | 69/74 = 0.932 | 73/81 = 0.901 | 75/84 = 0.893 | 68/87 = 0.782 | 13042649/14601384 = 0.8932 |
| luna-xhigh+jev | fastapi | 23/35 = 0.657 | 18/35 = 0.514 | 23/35 = 0.657 | 12/25 = 0.480 | 9/15 = 0.600 | 15/24 = 0.625 | 1649/2800 = 0.5889 |
| luna-xhigh+jev | axum | 73/74 = 0.986 | 73/74 = 0.986 | 73/74 = 0.986 | 78/81 = 0.963 | 78/84 = 0.929 | 76/87 = 0.874 | 1160923/1216782 = 0.9541 |
| terra-low | fastapi | 9/25 = 0.360 | 15/25 = 0.600 | 13/25 = 0.520 | 9/21 = 0.429 | 9/16 = 0.562 | 11/17 = 0.647 | 148423/285600 = 0.5197 |
| terra-low | axum | 30/33 = 0.909 | 29/33 = 0.879 | 27/33 = 0.818 | 26/29 = 0.897 | 28/31 = 0.903 | 29/33 = 0.879 | 156779/178002 = 0.8808 |
| terra-low+jev | fastapi | 11/25 = 0.440 | 20/25 = 0.800 | 16/25 = 0.640 | 12/21 = 0.571 | 12/16 = 0.750 | 13/17 = 0.765 | 47197/71400 = 0.6610 |
| terra-low+jev | axum | 31/33 = 0.939 | 31/33 = 0.939 | 29/33 = 0.879 | 28/29 = 0.966 | 30/31 = 0.968 | 31/33 = 0.939 | 83516/89001 = 0.9384 |
| terra-none | fastapi | 21/32 = 0.656 | 26/32 = 0.812 | 22/32 = 0.688 | 33/68 = 0.485 | 21/38 = 0.553 | 16/29 = 0.552 | 1122811/1798464 = 0.6243 |
| terra-none | axum | 31/39 = 0.795 | 26/39 = 0.667 | 33/39 = 0.846 | 32/35 = 0.914 | 29/34 = 0.853 | 33/39 = 0.846 | 76129/92820 = 0.8202 |
| terra-none+jev | fastapi | 29/32 = 0.906 | 30/32 = 0.938 | 32/32 = 1.000 | 38/68 = 0.559 | 23/38 = 0.605 | 21/29 = 0.724 | 1418381/1798464 = 0.7887 |
| terra-none+jev | axum | 34/39 = 0.872 | 35/39 = 0.897 | 34/39 = 0.872 | 34/35 = 0.971 | 34/34 = 1.000 | 35/39 = 0.897 | 2507/2730 = 0.9183 |

### 3.5 Strata at 0.30 (remap, mutation, fresh, same_state)

`prev` prior areas, `match` matches, `surv` mean survival (exact fraction), `diff` / `overl` the `different` / `overlapping` counts among labelled matches. Every other threshold: `data/analysis_output.txt`. `different` is 0 in every stratum at every threshold of every cell (checked programmatically for this report).

| cell | repo | stratum | prev | match | surv at 0.30 | diff | overl |
|---|---|---|---|---|---|---|---|
| haiku | fastapi | remap | 71 | 38 | 481/880 = 0.5466 | 0 | 9 |
| haiku | fastapi | mutation | 30 | 4 | 5/36 = 0.1389 | 0 | 1 |
| haiku | fastapi | fresh | 40 | 29 | 29/40 = 0.7250 | 0 | 6 |
| haiku | fastapi | same_state | 31 | 9 | 81/220 = 0.3682 | 0 | 3 |
| haiku | axum | remap | 70 | 62 | 286/323 = 0.8854 | 0 | 32 |
| haiku | axum | mutation | 45 | 37 | 833/1012 = 0.8231 | 0 | 27 |
| haiku | axum | fresh | 34 | 31 | 31/34 = 0.9118 | 0 | 16 |
| haiku | axum | same_state | 36 | 31 | 555/646 = 0.8591 | 0 | 16 |
| haiku+jev | fastapi | remap | 71 | 50 | 631/880 = 0.7170 | 0 | 20 |
| haiku+jev | fastapi | mutation | 30 | 6 | 5/24 = 0.2083 | 0 | 3 |
| haiku+jev | fastapi | fresh | 40 | 36 | 9/10 = 0.9000 | 0 | 13 |
| haiku+jev | fastapi | same_state | 31 | 14 | 47/88 = 0.5341 | 0 | 7 |
| haiku+jev | axum | remap | 70 | 66 | 609/646 = 0.9427 | 0 | 36 |
| haiku+jev | axum | mutation | 45 | 38 | 214/253 = 0.8458 | 0 | 28 |
| haiku+jev | axum | fresh | 34 | 32 | 16/17 = 0.9412 | 0 | 17 |
| haiku+jev | axum | same_state | 36 | 34 | 305/323 = 0.9443 | 0 | 19 |
| luna-xhigh | fastapi | remap | 130 | 60 | 16/35 = 0.4571 | 0 | 27 |
| luna-xhigh | fastapi | mutation | 39 | 20 | 121/240 = 0.5042 | 0 | 6 |
| luna-xhigh | fastapi | fresh | 70 | 35 | 1/2 = 0.5000 | 0 | 13 |
| luna-xhigh | fastapi | same_state | 60 | 25 | 29/70 = 0.4143 | 0 | 14 |
| luna-xhigh | axum | remap | 303 | 294 | 23275/23976 = 0.9708 | 0 | 26 |
| luna-xhigh | axum | mutation | 171 | 148 | 2111/2436 = 0.8666 | 0 | 15 |
| luna-xhigh | axum | fresh | 148 | 144 | 36/37 = 0.9730 | 0 | 13 |
| luna-xhigh | axum | same_state | 155 | 150 | 11611/11988 = 0.9686 | 0 | 13 |
| luna-xhigh+jev | fastapi | remap | 130 | 76 | 101/175 = 0.5771 | 0 | 40 |
| luna-xhigh+jev | fastapi | mutation | 39 | 24 | 49/80 = 0.6125 | 0 | 9 |
| luna-xhigh+jev | fastapi | fresh | 70 | 41 | 41/70 = 0.5857 | 0 | 19 |
| luna-xhigh+jev | fastapi | same_state | 60 | 35 | 199/350 = 0.5686 | 0 | 21 |
| luna-xhigh+jev | axum | remap | 303 | 297 | 7837/7992 = 0.9806 | 0 | 29 |
| luna-xhigh+jev | axum | mutation | 171 | 154 | 2195/2436 = 0.9011 | 0 | 18 |
| luna-xhigh+jev | axum | fresh | 148 | 146 | 73/74 = 0.9865 | 0 | 15 |
| luna-xhigh+jev | axum | same_state | 155 | 151 | 3895/3996 = 0.9747 | 0 | 14 |
| terra-low | fastapi | remap | 96 | 56 | 304/525 = 0.5790 | 0 | 19 |
| terra-low | fastapi | mutation | 33 | 24 | 99/136 = 0.7279 | 0 | 6 |
| terra-low | fastapi | fresh | 50 | 31 | 31/50 = 0.6200 | 0 | 11 |
| terra-low | fastapi | same_state | 46 | 25 | 113/210 = 0.5381 | 0 | 8 |
| terra-low | axum | remap | 128 | 118 | 589/638 = 0.9232 | 0 | 36 |
| terra-low | axum | mutation | 64 | 61 | 1951/2046 = 0.9536 | 0 | 24 |
| terra-low | axum | fresh | 66 | 61 | 61/66 = 0.9242 | 0 | 19 |
| terra-low | axum | same_state | 62 | 57 | 1765/1914 = 0.9222 | 0 | 17 |
| terra-low+jev | fastapi | remap | 96 | 59 | 429/700 = 0.6129 | 0 | 23 |
| terra-low+jev | fastapi | mutation | 33 | 25 | 103/136 = 0.7574 | 0 | 6 |
| terra-low+jev | fastapi | fresh | 50 | 31 | 31/50 = 0.6200 | 0 | 12 |
| terra-low+jev | fastapi | same_state | 46 | 28 | 106/175 = 0.6057 | 0 | 11 |
| terra-low+jev | axum | remap | 128 | 119 | 3563/3828 = 0.9308 | 0 | 37 |
| terra-low+jev | axum | mutation | 64 | 61 | 1951/2046 = 0.9536 | 0 | 24 |
| terra-low+jev | axum | fresh | 66 | 62 | 31/33 = 0.9394 | 0 | 20 |
| terra-low+jev | axum | same_state | 62 | 57 | 1765/1914 = 0.9222 | 0 | 17 |
| terra-none | fastapi | remap | 164 | 124 | 883/1088 = 0.8116 | 0 | 48 |
| terra-none | fastapi | mutation | 67 | 42 | 699/1102 = 0.6343 | 0 | 18 |
| terra-none | fastapi | fresh | 64 | 56 | 7/8 = 0.8750 | 0 | 16 |
| terra-none | fastapi | same_state | 100 | 68 | 407/544 = 0.7482 | 0 | 32 |
| terra-none | axum | remap | 152 | 134 | 2413/2730 = 0.8839 | 0 | 40 |
| terra-none | axum | mutation | 73 | 69 | 37/39 = 0.9487 | 0 | 12 |
| terra-none | axum | fresh | 78 | 66 | 11/13 = 0.8462 | 0 | 21 |
| terra-none | axum | same_state | 74 | 68 | 1258/1365 = 0.9216 | 0 | 19 |
| terra-none+jev | fastapi | remap | 164 | 129 | 1851/2176 = 0.8506 | 0 | 53 |
| terra-none+jev | fastapi | mutation | 67 | 44 | 1465/2204 = 0.6647 | 0 | 20 |
| terra-none+jev | fastapi | fresh | 64 | 59 | 59/64 = 0.9219 | 0 | 19 |
| terra-none+jev | fastapi | same_state | 100 | 70 | 53/68 = 0.7794 | 0 | 34 |
| terra-none+jev | axum | remap | 152 | 137 | 4931/5460 = 0.9031 | 0 | 43 |
| terra-none+jev | axum | mutation | 73 | 69 | 37/39 = 0.9487 | 0 | 12 |
| terra-none+jev | axum | fresh | 78 | 69 | 23/26 = 0.8846 | 0 | 24 |
| terra-none+jev | axum | same_state | 74 | 68 | 1258/1365 = 0.9216 | 0 | 19 |

### 3.6 Chained survival at 0.50 (label-free; design default; reported, not gating)

Fold F1→R1→R2→M1→M2 (T3–T6) with accumulating areas; `F1 matched all so far` is the share of the original F1 areas matched at every step so far. Jev cells' chain rows are the placeholders of §1.

| cell | repo | T3 matched/active | T4 | T5 | T6 | F1 areas matched at every step by T6 |
|---|---|---|---|---|---|---|
| haiku | fastapi | 1/20 | 5/23 | 8/28 | 2/30 | 1/20 |
| haiku | axum | 12/17 | 14/19 | 13/22 | 16/25 | 7/17 |
| haiku+jev | fastapi | – | – | – | – | – (empty in `results.json`: the `+jev` cells carry no chain) |
| haiku+jev | axum | – | – | – | – | – (empty in `results.json`: the `+jev` cells carry no chain) |
| luna-xhigh | fastapi | 9/35 | 4/43 | 12/49 | 20/60 | 0/35 |
| luna-xhigh | axum | 69/74 | 73/81 | 75/87 | 70/93 | 59/74 |
| luna-xhigh+jev | fastapi | – | – | – | – | – (empty in `results.json`: the `+jev` cells carry no chain) |
| luna-xhigh+jev | axum | – | – | – | – | – (empty in `results.json`: the `+jev` cells carry no chain) |
| terra-low | fastapi | 13/25 | 10/31 | 17/33 | 13/31 | 5/25 |
| terra-low | axum | 27/33 | 26/31 | 28/34 | 30/35 | 21/33 |
| terra-low+jev | fastapi | – | – | – | – | – (empty in `results.json`: the `+jev` cells carry no chain) |
| terra-low+jev | axum | – | – | – | – | – (empty in `results.json`: the `+jev` cells carry no chain) |
| terra-none | fastapi | 22/32 | 34/77 | 23/74 | 37/79 | 10/32 |
| terra-none | axum | 33/39 | 32/36 | 30/36 | 33/41 | 26/39 |
| terra-none+jev | fastapi | – | – | – | – | – (empty in `results.json`: the `+jev` cells carry no chain) |
| terra-none+jev | axum | – | – | – | – | – (empty in `results.json`: the `+jev` cells carry no chain) |

## 4. Served checks, Jev check, Jev cost and time

### 4.1 Served checks (`x2c.served`; `unexpected_arms` = `[]`)

| arm | served check | reasons |
|---|---|---|
| haiku | ok | [] |
| luna-xhigh | ok | [] |
| terra-low | ok | [] |
| terra-none | ok | [] |

Every probe (before and after, both repositories) names the arm's model and effort, every versions file records the arm's flags and every successful map's `agent-start` reported the arm's provider and model option (PREREGISTRATION §7, "Citable").

### 4.2 Jev check and the whole Jev run

Jev check: `ok` True; reasons `[]`; arm-only failures `{}`; failure rows by kind `{'schema': 12}`; `foreign_model_failure_rows` 0.

Jev requests (`x2c.jev_requests`): **3,972 answers**; failures 12 (kind `schema`: "choice probabilities incomplete or not normalised"); `cost_unknown_rows` 0; **total billed 0.219195 USD (≈ $0.219)**; share of answers with attempts > 1: 0.0; mean latency 0.2988 s. Model `typesafe/jev-1.14-20261014`, question `x2c.concern@2ff0e36e9db56d928d5b9beac8cb642b111e1ae805b5e2bea4133ec9c98bf1b2`. 1,986 pairs × 2 orders = 3,972 answers. The first `run.sh jev` answered 3,960 and failed 12 (`/tmp/x2c-jev.log`: "1986 pairs, 3960 answered this run, 12 failed, 0 with unknown billed cost, cumulative billed cost 0.218503 USD"; exit status 3); the one verbatim resume answered the remaining 12 and failed 0 (`/tmp/x2c-jev-resume1.log`: cumulative 0.219195 USD of which 0.218503 earlier). Both ran before the first analysis (§12). The 12 failure rows are counted in `failures` above (kind `schema`) and enter only the request-level counts, not the per-refresh cost and time (PREREGISTRATION §6: answers-only means). The budget estimate of PREREGISTRATION §9 was ≈ $0.30–0.40 billed; the measured figure is $0.219.

### 4.3 Unscored ratio: unscored occurrences ÷ candidate occurrences per (arm, repo) (`jev_check.per_arm`)

| arm | repo | unscored occurrences | candidate occurrences | share |
|---|---|---|---|---|
| haiku | fastapi | 0 | 89 | 0.0 |
| haiku | axum | 0 | 197 | 0.0 |
| luna-xhigh | fastapi | 0 | 184 | 0.0 |
| luna-xhigh | axum | 0 | 528 | 0.0 |
| terra-low | fastapi | 0 | 117 | 0.0 |
| terra-low | axum | 0 | 260 | 0.0 |
| terra-none | fastapi | 0 | 304 | 0.0 |
| terra-none | axum | 0 | 307 | 0.0 |

All 0 (limit 1 %). The candidate counts sum to 1,986 (= 197+89+528+184+260+117+307+304).

### 4.4 Jev refresh cost and time per (arm, repo) (`x2c.jev_refresh`)

`seconds` is the 8-worker service-time estimate (Σ request latencies ÷ 8; an assumption, PREREGISTRATION §6), `seconds_sequential` the sum of request latencies (the product client is sequential today, PREREGISTRATION §11); cost and time are answers-only means per refresh over the 6 measured transitions.

| arm / repo | cost per refresh (USD) | requests per refresh | seconds (÷ 8) | seconds sequential | transitions |
|---|---|---|---|---|---|
| haiku / fastapi | 0.001647 | 29.67 | 1.1126 | 8.9010 | 6 |
| haiku / axum | 0.003815 | 65.67 | 2.4781 | 19.8248 | 6 |
| luna-xhigh / fastapi | 0.003665 | 61.33 | 2.2688 | 18.1502 | 6 |
| luna-xhigh / axum | 0.008632 | 176.00 | 6.5355 | 52.2840 | 6 |
| terra-low / fastapi | 0.002346 | 39.00 | 1.4547 | 11.6373 | 6 |
| terra-low / axum | 0.004587 | 86.67 | 3.2681 | 26.1445 | 6 |
| terra-none / fastapi | 0.006326 | 101.33 | 3.8161 | 30.5288 | 6 |
| terra-none / axum | 0.005398 | 102.33 | 3.7934 | 30.3468 | 6 |

Per cell (sum over the two repositories; the cap is 0.10 USD and 300 s Jev per refresh, PREREGISTRATION §6): see the last column of §3.2. The largest is luna-xhigh+jev, 0.0123 USD, 8.804 s at 8 workers, 70.434 s sequential: far inside the caps. The caps are never the reason a cell fails.

## 5. Cells examined, eligible and false-pass figures (PREREGISTRATION §8)

Cells examined **8**; cells eligible **0**. Per cell, the largest over the 13 thresholds and the two repositories of the probability that one repository with that threshold's realised labelled count meets C2/C3 at the true spurious rate 0.08 / 0.10 (`x2c.family.per_cell.<cell>.p_false_pass`; single-threshold, conditional on C1, independent matches: an approximation to the family-wise false-support probability, not an upper bound):

| cell | p_false_pass at 0.08 | at 0.10 | at (0.08, 0.10): threshold, repo |
|---|---|---|---|
| haiku | 0.0948 | 0.0254 | 0.30 axum, 0.30 axum |
| haiku+jev | 0.0903 | 0.0259 | 0.90 axum, 0.70 fastapi |
| luna-xhigh | 0.0740 | 0.0242 | 0.35 fastapi, 0.35 fastapi |
| luna-xhigh+jev | 0.0948 | 0.0254 | 0.85 fastapi, 0.85 fastapi |
| terra-low | 0.0912 | 0.0242 | 0.65 axum, 0.40 fastapi |
| terra-low+jev | 0.0884 | 0.0264 | 0.90 fastapi, 0.90 fastapi |
| terra-none | 0.1204 | 0.0262 | 0.45 fastapi, 0.85 fastapi |
| terra-none+jev | 0.0619 | 0.0084 | 0.90 fastapi, 0.90 fastapi |

`sum_of_maxima`: **0.7159** at 0.08 and **0.1861** at 0.10 (not None). Operating characteristics (`results.json` `operating_characteristics`; probability of adopting at one passing threshold, conditional on C1, both repositories) for n = 150 / 216 / 270 per repository at true spurious 0.03 / 0.04 / 0.05 / 0.07 / 0.10: 0.8401 / 0.5578 / 0.2733 / 0.0285 / 0.0002; 0.8784 / 0.5639 / 0.2318 / 0.0109 / 0.0; 0.9314 / 0.6456 / 0.2671 / 0.0088 / 0.0; `c3_stops_binding_at_n` 140 (X2's multiple-of-20 search). The analysis prints X2's "expected n 216-270 labelled occurrences per repo" band, which is X2's and not X2c's (the realised labelled counts are in §3.1 and §7.1). No cell was eligible, so none of these figures supports a claim; they are reported because they are preregistered. No family-wise correction was applied (PREREGISTRATION §8).

## 6. Coverage diagnostics and area ratio (descriptive, except the coverage guard, which feeds eligibility)

Coverage guard (PREREGISTRATION §7): ratio of the arm's mean owned files per map to the reference's (reference = luna-xhigh, ratio 1.0 by definition), floor 0.5; area ratio = mean of (prior areas + next agent features) ÷ 2 per map relative to the reference's, a ratio below 0.8 sends the question to the design round. The guard only matters to `adopt` cells; no cell adopted, so it changes no verdict here.

| arm | mean areas per map (axum, fastapi) | mean owned files per map (axum, fastapi) | file ratio (axum, fastapi) | area ratio (axum, fastapi) | coverage ok | reasons |
|---|---|---|---|---|---|---|
| haiku | 20.67, 17.25 | 162.25, 116.08 | 0.8113, 0.4014 | 0.2528, 0.6551 | False | fastapi: mean owned files per map 116.1 is 0.40 of the reference's 289.2 (floor 0.50) |
| luna-xhigh | 81.75, 26.33 | 200.00, 289.17 | 1.0000, 1.0000 | 1.0000, 1.0000 | True | – |
| terra-low | 31.67, 21.00 | 170.58, 216.33 | 0.8529, 0.7481 | 0.3874, 0.7975 | True | – |
| terra-none | 36.92, 40.67 | 186.50, 471.83 | 0.9325, 1.6317 | 0.4516, 1.5443 | True | – |

**Haiku coverage ok = false**: fastapi's mean owned files per map is 116.1, 0.40 of the reference's 289.2 (floor 0.50); axum's file ratio 0.8113 passes. Haiku's area ratio is 0.2528 (axum) and 0.6551 (fastapi), terra-low's 0.3874 and 0.7975, terra-none's 0.4516 and 1.5443: below 0.8 for haiku (both), terra-low (both) and terra-none on axum. [INFERENCE] Those arms map fewer, larger areas than luna-xhigh on axum (luna-xhigh averages 81.75 areas per map there, terra-none 36.92, terra-low 31.67, haiku 20.67); a cell with fewer areas has fewer prior areas to lose but is also a coarser product. Terra-none on fastapi has 1.54 times the reference's areas and 1.63 times its owned files, i.e. a finer and broader map.

Per-transition coverage (prior areas, next agent features, owned files, shared owned files, zero-overlap prior areas) for every arm is in `data/analysis_output.txt` (`-- <repo> coverage (T_MIN rows)`). Zero-overlap prior areas per transition T1–T6 (a survival ceiling no threshold or matcher lifts; `1 − Σ/prev` per transition averages to the §1 ceiling):

| arm | repo | prior areas T1–T6 | zero-overlap prior areas T1–T6 | next agent features T1–T6 | ceiling (mean) |
|---|---|---|---|---|---|
| haiku | fastapi | 20, 20, 20, 11, 12, 18 | 0, 0, 12, 2, 9, 12 | 31, 19, 11, 12, 18, 15 | 0.6336 |
| haiku | axum | 17, 17, 17, 19, 22, 23 | 0, 0, 0, 0, 1, 1 | 22, 27, 19, 22, 23, 20 | 0.9852 |
| luna-xhigh | fastapi | 35, 35, 35, 25, 15, 24 | 5, 10, 4, 4, 5, 7 | 33, 20, 25, 15, 24, 30 | 0.7787 |
| luna-xhigh | axum | 74, 74, 74, 81, 84, 87 | 0, 1, 1, 1, 1, 5 | 89, 85, 81, 84, 87, 81 | 0.9819 |
| terra-low | fastapi | 25, 25, 25, 21, 16, 17 | 14, 0, 9, 3, 3, 3 | 19, 30, 21, 16, 17, 20 | 0.7622 |
| terra-low | axum | 33, 33, 33, 29, 31, 33 | 0, 0, 0, 0, 0, 0 | 32, 32, 29, 31, 33, 31 | 1.0000 |
| terra-none | fastapi | 32, 32, 32, 68, 38, 29 | 0, 1, 0, 2, 7, 0 | 37, 44, 68, 38, 29, 41 | 0.9592 |
| terra-none | axum | 39, 39, 39, 35, 34, 39 | 0, 0, 0, 0, 0, 0 | 35, 39, 35, 34, 39, 36 | 1.0000 |

Area counts per map on fastapi range, over the prior and next maps of the six transitions, from 11 to 31 (haiku), 15 to 35 (luna-xhigh), 16 to 30 (terra-low) and 29 to 68 (terra-none); on axum from 17 to 27, 74 to 89, 29 to 33 and 34 to 39. [INFERENCE] A fastapi arm that redraws its areas between runs leaves many prior areas without any file in common with the next map; that, not the match threshold, caps survival on fastapi, which is the reading §1 gives of the candidate ceiling.

## 7. Labels

### 7.1 Agreement, adjudication, served models, occurrences against unique pairs

- Unique real pairs labelled: **1,986** (`controls.n_real_pairs`); labelled occurrences = unique pairs (1,986 occurrences over 48 measured transitions, `/tmp/x2c-pairs.log`: "1986 unique real pairs, 1986 occurrences over 48 measured transitions, 239 control pairs (159 swapped, 80 identical)"). Control pairs 239. Call rows per role: **A 2,225, B 2,225, C 96** (2,225 = 1,986 + 239; the 96 C rows are adjudications over real pairs and controls, all rounds).
- Agreement covers **round-1 real pairs only: 260 pairs**, because new-round rows are excluded by design (§12 deviation 2). A/B raw agreement **0.9346153846153846**; Cohen's κ **0.8857231501111743**; both parsed 260, unparsable final 0, codex sessions 260, pairs with an exec event 0.
- Adjudicated pairs (C sessions), round-1 real pairs: **17** of 260, **adjudication rate 0.06538461538461539**; C answered validly in 17 of 17; C's served model equals A's in 0 of 17 (share 0.0). The analysis reports the adjudication rate for round 1 only.
- Served models per role (final attempt's model per call row, every call row incl. controls and later rounds): **A `claude-opus-5-5` (2,225 call rows), B `gpt-6.1-sol` (2,225), C `claude-sonnet-5-5` (96)**. A is the `claude` default, B the `codex` default, C the adjudicator. Labeller B is a GPT-6.1 model and the mappers luna and terra are GPT-6 family models; the same-family caveat (X2 §10 shared blind spots) applies to the codex arms **[INFERENCE: family inferred from model names]**; labeller A is the same family as the haiku arm.
- Label integrity: `labelled_pairs_without_call_row` **0**.
- Controls agreement (round 1, controls): 22 pairs, raw 1.0, κ 1.0, adjudicated 0.
- No human checks these labels (label status `agent_reviewed`; CLAUDE.md, Jev labelling).

### 7.2 Known-answer controls (pooled over every arm; gate ok)

Control gate: ok, reason `None`. 1,986 unique real pairs, 239 control pairs, shortfall 0, `labelled_pairs_without_call_row` 0.
- **Swapped** (expected `different_concern`): n = 159 (wanted 159, shortfall 0); usable 159 (minimum 80); misses (not `different_concern`) **7**, rate **0.0440** (gate limit 1/5); cannot_tell 0, unparsable 0, never labelled 0; labels {"different_concern": 152, "overlapping_concern": 6, "same_concern": 1}.
- **Identical** (expected `same_concern`): n = 80 (wanted 80, shortfall 0); usable 80 (minimum 40); misses **0**, rate 0.000; labels {"same_concern": 80}.
- Per arm (descriptive): haiku identical 0/7, swapped 0/24; luna-xhigh identical 0/26, swapped 4/53; terra-low identical 0/20, swapped 1/26; terra-none identical 0/27, swapped 2/56.
- Swapped controls by source-pair Jaccard band (misses/n): 0.30: 6/57, 0.35: 0/8, 0.40: 0/6, 0.45: 0/1, 0.50: 0/21, 0.55: 0/2, 0.60: 1/7, 0.65: 0/4, 0.75: 0/8, 0.80: 0/1, 0.85: 0/6, 0.90: 0/38. **Misses concentrate in the lowest band (0.30: 6 of 57)** and one at 0.60 (1 of 7): the swapped controls here include 57 source pairs in the 0.30 band (X2b's three misses were all in the 0.90 band, X2b RESULTS §7.2). Of the 7 misses, 6 are `overlapping_concern` and 1 is `same_concern`. [INFERENCE] Where two source areas share few files, labellers sometimes call a swapped pair `overlapping`; `overlapping` is not always a strict synonym for a loose match.

### 7.3 Label distribution by Jaccard band and over-call check

Real matches at T_MIN per base arm; counts of final labels (`same_concern` / `overlapping_concern`; no `different_concern`, no `cannot_tell` in any arm and repository). Over-call check: `different` share of labelled matches in the [0.90, 1.0] band. Complete per-band distributions: `data/analysis_output.txt` (`-- <repo> label distribution`).

| arm | repo | by class (mutation, remap) | over-call check [0.90, 1.0] |
|---|---|---|---|
| haiku | fastapi | mutation same 3 / overl 1; remap same 29 / overl 9 | 0 of 10 (share 0.000) |
| haiku | axum | mutation same 10 / overl 27; remap same 30 / overl 32 | 0 of 25 (share 0.000) |
| luna-xhigh | fastapi | mutation same 14 / overl 6; remap same 33 / overl 27 | 0 of 22 (share 0.000) |
| luna-xhigh | axum | mutation same 133 / overl 15; remap same 268 / overl 26 | 0 of 199 (share 0.000) |
| terra-low | fastapi | mutation same 18 / overl 6; remap same 37 / overl 19 | 0 of 39 (share 0.000) |
| terra-low | axum | mutation same 37 / overl 24; remap same 82 / overl 36 | 0 of 86 (share 0.000) |
| terra-none | fastapi | mutation same 24 / overl 18; remap same 76 / overl 48 | 0 of 64 (share 0.000) |
| terra-none | axum | mutation same 57 / overl 12; remap same 94 / overl 40 | 0 of 102 (share 0.000) |

## 8. Map inventory, attempts and wall times

56 of 56 real maps recorded `ok` (the 112 inventory rows include 56 `+jev` placeholders with status `not recorded`: the Jev cells run no maps); retried, failed, skipped or workspace-dirty maps: none (every real map `attempts` 1, `rc` 0, `workspace_dirty` false); stop records (`data/stopped-*.json`): none (`inventory.stopped` is empty). Collection order (`x2c.COLLECTION_ORDER`, sequential, `/tmp/x2c-collect.log`): luna-xhigh/fastapi, terra-none/fastapi, terra-low/fastapi, haiku/fastapi, haiku/axum, terra-low/axum, terra-none/axum, luna-xhigh/axum.

Map seconds (`x2c.map_seconds`): n, mean, min, max per (arm, repo) and per map id. Refresh time of a base cell is the sum over both repositories of the per-repository mean.

| arm | repo | n | mean | min | max | F1 | F2 | F3 | R1 | R2 | M1 | M2 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| haiku | fastapi | 7 | 70.53 | 38.7 | 108.3 | 70.2 | 108.3 | 86.7 | 38.7 | 42.0 | 94.7 | 53.1 |
| haiku | axum | 7 | 104.31 | 75.8 | 175.4 | 75.8 | 80.3 | 104.1 | 123.1 | 79.8 | 91.7 | 175.4 |
| luna-xhigh | fastapi | 7 | 149.49 | 90.4 | 208.3 | 189.6 | 164.1 | 208.3 | 90.4 | 92.5 | 110.6 | 190.9 |
| luna-xhigh | axum | 7 | 199.51 | 161.4 | 272.7 | 186.8 | 161.4 | 231.1 | 272.7 | 177.7 | 162.9 | 204.0 |
| terra-low | fastapi | 7 | 62.69 | 50.8 | 85.6 | 85.6 | 56.8 | 85.0 | 55.0 | 54.5 | 50.8 | 51.1 |
| terra-low | axum | 7 | 83.59 | 64.7 | 147.4 | 147.4 | 81.5 | 81.7 | 70.0 | 69.0 | 64.7 | 70.8 |
| terra-none | fastapi | 7 | 136.84 | 93.1 | 167.0 | 106.0 | 147.3 | 143.5 | 167.0 | 145.0 | 93.1 | 156.0 |
| terra-none | axum | 7 | 110.97 | 98.4 | 120.8 | 114.9 | 117.7 | 102.3 | 109.3 | 98.4 | 120.8 | 113.4 |

Sum over repositories of the per-repository mean (refresh time of the base cell): haiku 174.843 s; luna-xhigh 349.000 s; terra-low 146.271 s; terra-none 247.814 s.

## 9. Tool versions, probes, provenance

### 9.1 Tool versions (`data/tool-versions-<arm>-<repo>.json`)

Identical in all eight files: Claude Code `2.1.312 (Claude Code)`, `codex-cli 0.164.0`, Clawpatch `0.7.3` (node `v22.18.0`), `clawpatch_cli_sha256` `4eab263b87d6b0d0af1eb6792e1e38827e0b13ab53078e62f5ffed76cc4320eb` (the same value as X2b's, `4eab263b…20eb`), `clawpatch_build_safety` ok=True, marker_ok=True, sanitizer_in_dist=True, marker `scout-review-sanitize-v3 dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3`; `scout_head` `996541061e295101d197437544e6c0d31c085496` (= the run commit); `areas_go_sha256` `317564d0eba850d8375663c271070b38f85acd7317a83f3e2df2efe5509323da`; `records_go_sha256` `942d4efc31fd113ba831d8cd0d5759dc41d79c14f369f2d0632377706cecfe77`; provider timeouts claude/codex 900000/900000 ms, per-call 1800 s, wall 28800 s, grace 10 s.
(Equality across the eight files checked programmatically for this report: True.) Workspace config presence (all five names): none present in every file. The tool-versions files also record the names of the environment variables present (no values).

| arm | repo | started (UTC) | flags (`x2c`) |
|---|---|---|---|
| haiku | fastapi | 2026-10-29T01:13:40Z | `{"arm": "haiku", "collect_x2c_sha256": "7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff", "effort": "none", "map_flags": ["--model", "claude-haiku-5-5", "--reasoning-effort", "none"], "model": "claude-haiku-5-5", "provider": "claude", "x2_pins": {"GUIDE.md": "9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799", "PREREGISTRATION.md": "ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355", "analysis.py": "7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63", "build_pairs.py": "face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a", "common.py": "e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3", "label.py": "87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f", "replay/main.go": "f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e", "validation/credcheck/main.go": "e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779", "validation/stub_clawpatch.js": "0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226", "validation/validate.py": "29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784"}}` |
| haiku | axum | 2026-10-29T07:10:03Z | `{"arm": "haiku", "collect_x2c_sha256": "7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff", "effort": "none", "map_flags": ["--model", "claude-haiku-5-5", "--reasoning-effort", "none"], "model": "claude-haiku-5-5", "provider": "claude", "x2_pins": {"GUIDE.md": "9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799", "PREREGISTRATION.md": "ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355", "analysis.py": "7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63", "build_pairs.py": "face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a", "common.py": "e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3", "label.py": "87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f", "replay/main.go": "f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e", "validation/credcheck/main.go": "e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779", "validation/stub_clawpatch.js": "0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226", "validation/validate.py": "29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784"}}` |
| luna-xhigh | fastapi | 2026-10-29T00:29:40Z | `{"arm": "luna-xhigh", "collect_x2c_sha256": "7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff", "effort": "xhigh", "map_flags": ["--model", "gpt-6-luna", "--reasoning-effort", "xhigh"], "model": "gpt-6-luna", "provider": "codex", "x2_pins": {"GUIDE.md": "9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799", "PREREGISTRATION.md": "ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355", "analysis.py": "7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63", "build_pairs.py": "face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a", "common.py": "e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3", "label.py": "87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f", "replay/main.go": "f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e", "validation/credcheck/main.go": "e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779", "validation/stub_clawpatch.js": "0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226", "validation/validate.py": "29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784"}}` |
| luna-xhigh | axum | 2026-10-29T07:48:10Z | `{"arm": "luna-xhigh", "collect_x2c_sha256": "7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff", "effort": "xhigh", "map_flags": ["--model", "gpt-6-luna", "--reasoning-effort", "xhigh"], "model": "gpt-6-luna", "provider": "codex", "x2_pins": {"GUIDE.md": "9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799", "PREREGISTRATION.md": "ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355", "analysis.py": "7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63", "build_pairs.py": "face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a", "common.py": "e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3", "label.py": "87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f", "replay/main.go": "f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e", "validation/credcheck/main.go": "e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779", "validation/stub_clawpatch.js": "0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226", "validation/validate.py": "29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784"}}` |
| terra-low | fastapi | 2026-10-29T01:05:36Z | `{"arm": "terra-low", "collect_x2c_sha256": "7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff", "effort": "low", "map_flags": ["--model", "gpt-6.1-terra", "--reasoning-effort", "low"], "model": "gpt-6.1-terra", "provider": "codex", "x2_pins": {"GUIDE.md": "9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799", "PREREGISTRATION.md": "ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355", "analysis.py": "7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63", "build_pairs.py": "face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a", "common.py": "e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3", "label.py": "87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f", "replay/main.go": "f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e", "validation/credcheck/main.go": "e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779", "validation/stub_clawpatch.js": "0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226", "validation/validate.py": "29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784"}}` |
| terra-low | axum | 2026-10-29T07:23:05Z | `{"arm": "terra-low", "collect_x2c_sha256": "7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff", "effort": "low", "map_flags": ["--model", "gpt-6.1-terra", "--reasoning-effort", "low"], "model": "gpt-6.1-terra", "provider": "codex", "x2_pins": {"GUIDE.md": "9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799", "PREREGISTRATION.md": "ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355", "analysis.py": "7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63", "build_pairs.py": "face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a", "common.py": "e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3", "label.py": "87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f", "replay/main.go": "f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e", "validation/credcheck/main.go": "e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779", "validation/stub_clawpatch.js": "0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226", "validation/validate.py": "29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784"}}` |
| terra-none | fastapi | 2026-10-29T00:48:19Z | `{"arm": "terra-none", "collect_x2c_sha256": "7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff", "effort": "none", "map_flags": ["--model", "gpt-6.1-terra", "--reasoning-effort", "none"], "model": "gpt-6.1-terra", "provider": "codex", "x2_pins": {"GUIDE.md": "9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799", "PREREGISTRATION.md": "ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355", "analysis.py": "7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63", "build_pairs.py": "face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a", "common.py": "e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3", "label.py": "87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f", "replay/main.go": "f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e", "validation/credcheck/main.go": "e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779", "validation/stub_clawpatch.js": "0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226", "validation/validate.py": "29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784"}}` |
| terra-none | axum | 2026-10-29T07:34:01Z | `{"arm": "terra-none", "collect_x2c_sha256": "7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff", "effort": "none", "map_flags": ["--model", "gpt-6.1-terra", "--reasoning-effort", "none"], "model": "gpt-6.1-terra", "provider": "codex", "x2_pins": {"GUIDE.md": "9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799", "PREREGISTRATION.md": "ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355", "analysis.py": "7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63", "build_pairs.py": "face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a", "common.py": "e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3", "label.py": "87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f", "replay/main.go": "f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e", "validation/credcheck/main.go": "e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779", "validation/stub_clawpatch.js": "0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226", "validation/validate.py": "29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784"}}` |

### 9.2 Model probes and probe consistency (`probe_consistency`)

| arm | consistent | models | CLI versions changed | reasons |
|---|---|---|---|---|
| haiku | True | ['claude-haiku-5-5'] | {} | [] |
| haiku+jev | False | [] | {} | ['axum: no tool-versions file', 'fastapi: no tool-versions file'] |
| luna-xhigh | True | ['gpt-6-luna'] | {} | [] |
| luna-xhigh+jev | False | [] | {} | ['axum: no tool-versions file', 'fastapi: no tool-versions file'] |
| terra-low | True | ['gpt-6.1-terra'] | {} | [] |
| terra-low+jev | False | [] | {} | ['axum: no tool-versions file', 'fastapi: no tool-versions file'] |
| terra-none | True | ['gpt-6.1-terra'] | {} | [] |
| terra-none+jev | False | [] | {} | ['axum: no tool-versions file', 'fastapi: no tool-versions file'] |

Probe lines per base arm (`before` / `after`; model, requested or banner effort, seconds, UTC):

| arm | repo | before | after |
|---|---|---|---|
| haiku | fastapi | claude-haiku-5-5, none (requested), 4.5 s, 2026-10-29T01:13:37Z, error None | claude-haiku-5-5, none (requested), 5.3 s, 2026-10-29T07:09:56Z, error None |
| haiku | axum | claude-haiku-5-5, none (requested), 4.6 s, 2026-10-29T07:10:01Z, error None | claude-haiku-5-5, none (requested), 4.7 s, 2026-10-29T07:22:21Z, error None |
| luna-xhigh | fastapi | gpt-6-luna, xhigh (banner), 34.4 s, 2026-10-29T00:29:38Z, error None | gpt-6-luna, xhigh (banner), 33.4 s, 2026-10-29T00:47:43Z, error None |
| luna-xhigh | axum | gpt-6-luna, xhigh (banner), 33.1 s, 2026-10-29T07:48:09Z, error None | gpt-6-luna, xhigh (banner), 33.4 s, 2026-10-29T08:12:03Z, error None |
| terra-low | fastapi | gpt-6.1-terra, low (banner), 33.1 s, 2026-10-29T01:05:30Z, error None | gpt-6.1-terra, low (banner), 33.1 s, 2026-10-29T01:13:32Z, error None |
| terra-low | axum | gpt-6.1-terra, low (banner), 42.1 s, 2026-10-29T07:23:04Z, error None | gpt-6.1-terra, low (banner), 33.4 s, 2026-10-29T07:33:26Z, error None |
| terra-none | fastapi | gpt-6.1-terra, none (banner), 33.1 s, 2026-10-29T00:48:17Z, error None | gpt-6.1-terra, none (banner), 33.6 s, 2026-10-29T01:04:55Z, error None |
| terra-none | axum | gpt-6.1-terra, none (banner), 33.1 s, 2026-10-29T07:34:00Z, error None | gpt-6.1-terra, none (banner), 33.4 s, 2026-10-29T07:47:35Z, error None |

**Probe-consistency output inconsistency for the `+jev` arms (does not change any verdict).** The analysis prints `probe consistency … INCONSISTENT … not citable for a design change` for the four `+jev` pseudo-arms (`haiku+jev`, `luna-xhigh+jev`, `terra-low+jev`, `terra-none+jev`), with reason "no tool-versions file" for both repositories, because the X2 probe check looks for `tool-versions-<arm>-<repo>.json` under the `+jev` arm name, and the Jev cells run no maps and have no such file. The `+jev` cells reuse the base arms' maps, whose probes are `CONSISTENT` (table above; the served check of each base arm is `ok`, §4.1). The cell lines still print `citable: yes` (and `results.json` `x2c.cells.<arm>+jev.citable` is `true`). This is an analysis-output inconsistency; every cell is `fallback`, so it changes no verdict and no selection. [INFERENCE] The `+jev` probe lines are an artefact of the analysis key, not evidence about the Jev cells' models.

Probes pass `--model` and the effort themselves, so a banner echoes the requested values; a same-name server-side roll of a model is invisible before and after (PREREGISTRATION §11) **[INFERENCE]**. `haiku` has no effort banner (`reasoning_effort` `none`, source `requested`).

### 9.3 Replay provenance (`data/replay-meta.json`)

Scout HEAD `996541061e295101d197437544e6c0d31c085496`; `internal/areas/areas.go` sha256 `317564d0eba850d8375663c271070b38f85acd7317a83f3e2df2efe5509323da`; `internal/clawpatch/records.go` sha256 `942d4efc31fd113ba831d8cd0d5759dc41d79c14f369f2d0632377706cecfe77`; working tree dirty for those two files: False; Go `go1.25.6`; grid `[30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90]` percent; transitions file sha256 `9a7df933bb6d48f65be2385a319b4bea2f6add8884936d960cf090f4c3f0ffd9`; maps directory `evaluation/x2c-mapper-comparison/data/maps`.

## 10. Unmeasured counts and reasons

- **Mapper cost: not recorded** (`x2c.map_cost`: "not recorded: Clawpatch prints no tokens or cost (X2 section 9); map cost is not a decision input"). Only map wall seconds are reported (§8). Jev cost is measured (§4).
- Transitions lost to failed, stopped or skipped maps: 0 (48 of 48 transitions measured; `measured` true for every held-out row of §3.3); maps retried or failed: 0; stops: 0; mutation skipped: 0; unlabelled pairs: 0 (cannot_tell 0, unparsable 0, never labelled 0 at every row of the analysis); Jev unscored occurrences: 0 of 1,986.
- `<arm>+jev` created/kept/retired and kept-inclusive survival, and the inventory entries for the `+jev` arms: unmeasured schema placeholders (§1).

## 11. Run-commit provenance, sign-off and digest equality

### 11.1 Run commit and sign-off

**Run commit: `996541061e295101d197437544e6c0d31c085496`** ("X2c: preregistration rev 3 and implementation signed off (council round 5, zero issues); HD5c preregistered; R-048"), a child of `5d3122fcb4d06041147b4cfb7cb88360272039a7`; pushed to `origin/main`. It contains the 13 frozen study files, `reviews/` rounds 1–5 and responses 1–3, `DRAFT-NOTES.md`, the ledger appends (R-048, HD5c) and nothing under `internal/`, `scripts/`, `toolchain/`, `third_party/` or `go.*` that differs from its parent: `git diff --quiet 5d3122f 9965410… -- internal/ scripts/ third_party/clawpatch go.mod go.sum toolchain/` exits 0 (run for this report). `git submodule status third_party/clawpatch`: ` 734f81bc1514b52a0627edb0c24ebe1a18b36105 third_party/clawpatch (v0.7.3)` (no `+`/`-` prefix). `git status --porcelain internal/ scripts/ evaluation/x2c-mapper-comparison/` in the run worktree after the run shows only `?? evaluation/x2c-mapper-comparison/data/` (run for this report).

**Revision 3 was signed off in council round 5 (`reviews/round-5.md`), after round 4's zero-issue review of the revision.** Round 4's sign-off was voided by its own condition 1 when one value-identical line of `validation/validate.py` (l.45, the synthetic stub key split by concatenation so the pre-commit Gitleaks gate passes) changed; round 5 reviewed exactly that change and re-issued the complete sign-off. `validate.py`'s round-4 digest `22483acf31fd1ad5aef1e0a2e248d872e1f62d8bb63c61150c5074d66f3dd6ff` is **superseded** by `4a61b009c6cae8dbac0d68d06d00d4c0cb7eec2bad5d67f428e447abbf154f4d`. Rounds 1–5 are the review record (`reviews/round-1.md` … `round-5.md`, responses 1–3).

### 11.2 Digest equality

The SHA-256 of every frozen file and every code-under-test blob was recomputed from the run commit with `git -C /Users/adamjackson/Projects/inflight/scout-x2c show 996541061e295101d197437544e6c0d31c085496:<path> | shasum -a 256` and compared with the round-5 sign-off lists (`reviews/round-5.md`, "Reviewed and frozen" and "Code under test"). **Result: all 26 values are equal to the sign-off values** (13 frozen files and 13 code-under-test blobs; 0 mismatches, 0 missing). `PINNED_COLLECT_BASE` (`89948f3c…`) and the ten `x2c.X2_PINS` entries are pins of X2 files that `collect_x2c.py` and `x2c.py` verify at start; they were not recomputed here.

| path | sha256 (equal to round 5) |
|---|---|
| `evaluation/x2c-mapper-comparison/PREREGISTRATION.md` | `0a05fbca28bef210a59c8746895718917d7334cc03dcee3119cd71e9c0610657` |
| `evaluation/x2c-mapper-comparison/README.md` | `83479745585c65d5f8f2db8e2b1777466c154d06c59e5a8d379e33b9a370f1b0` |
| `evaluation/x2c-mapper-comparison/x2c.py` | `cbef182f37682630340ab0055c9c8e93c04dbf38be9036800fe4bfbb848c631e` |
| `evaluation/x2c-mapper-comparison/collect_x2c.py` | `7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff` |
| `evaluation/x2c-mapper-comparison/build_candidates.py` | `3597e005e1e1388de04e82a47b8e59164fcf65c60b13c9dbad6c3c58d7c75029` |
| `evaluation/x2c-mapper-comparison/jev_match.py` | `1234d1c22b7417b6ab5b7c02d0d913866ffeca9cc42e8711b2a3923a60e10fa0` |
| `evaluation/x2c-mapper-comparison/analysis_x2c.py` | `00f41631d9e27b72a0f7da5b6642f69e4db0181d8e4a2c63a546fa01aeed8648` |
| `evaluation/x2c-mapper-comparison/jev/main.go` | `6a6001a5710d1e9a661171874ab7c9a562bb1df4f7cd1b7c6d39512922a1af1f` |
| `evaluation/x2c-mapper-comparison/run.sh` | `a5f03458064ca61370f95927970f13c4322e3674ab42506efcdf005e55491514` |
| `evaluation/x2c-mapper-comparison/validation/validate.py` | `4a61b009c6cae8dbac0d68d06d00d4c0cb7eec2bad5d67f428e447abbf154f4d` |
| `evaluation/x2c-mapper-comparison/validation/stub_clawpatch_x2c.js` | `480340b2ca7be7cba69d14eac3017336d9d207e472f5e8b15692e271dd150b89` |
| `evaluation/x2c-mapper-comparison/validation/stub_jev.py` | `86e01ed2fb0349fdd2cbaff7facc3d60849125e8aa48eb5c36ff35097cbef1c6` |
| `evaluation/x2c-mapper-comparison/validation/collect-x2c.diff` | `b75cf8e937b9925bc0c9fc723d45a31e097f7da83bc4fe953960098c32d4e128` |
| `internal/decisions/audit.go` | `09ff2df8aece406181052c760a6798adaa0c4e37068aa4d9b724454bcb919cb2` |
| `internal/decisions/client.go` | `e56aadef3f245151e489f4a28de11ab953730a139ab4809085f5a95c0e6806ae` |
| `internal/decisions/question.go` | `832c87601e3ec22d0101bcb4a46160546dfdb56d4c10d7f3038f74eaaf3eee0b` |
| `internal/decisions/rank.go` | `605a39bb268b39b132aa2e8b851aa0ab17dd2faf3a10c4a4aced4575f61c36ab` |
| `internal/decisions/run.go` | `e696b6fa95b11b108763681462ef8f0002ae3dbcaa6d2c83af5928e15a3284ae` |
| `internal/decisions/triage.go` | `8d22dc18b5c0f74a9df3b44f0e81191b7e2d3e9e9009593875b224a70d4b161c` |
| `go.mod` | `7009be395a9f776345c3810af869a955a8ec569f0e9c969e9dd81b61ef710f24` |
| `go.sum` | `24d99eff5500559b16cff2acba29f1ceaf3d73110046cbc6906f61cc2a927255` |
| `internal/policy/consent.go` | `c8e55ff285e2ba3e5a38e66e09704f2d0ef607f552a773954b1a02c2ce7c3e4d` |
| `internal/areas/areas.go` | `317564d0eba850d8375663c271070b38f85acd7317a83f3e2df2efe5509323da` |
| `internal/clawpatch/records.go` | `942d4efc31fd113ba831d8cd0d5759dc41d79c14f369f2d0632377706cecfe77` |
| `scripts/clawpatch-safe.patch` | `dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3` |
| `scripts/build-clawpatch.sh` | `1b3eda9d166759be4dc04761757e93658cb8cb5e1088b1b6832a77639bc6b9b5` |

### 11.3 `jev_code_digests` (`data/jev/code-digests.jsonl`, one line per `run.sh jev` invocation)

| line | `utc` | invocation | digests |
|---|---|---|---|
| 1 | 2026-10-29T08:12:36Z | the first `run.sh jev` (3,960 answered, 12 failed) | 11 files |
| 2 | 2026-10-29T08:15:21Z | the verbatim resume (12 answered, 0 failed) | 11 files |

Both lines precede the analysis: `data/jev/code-digests.jsonl` was last written 08:15:21Z and `data/jev/scores.jsonl` 08:15:23Z (file times), while `data/results.json` and `data/analysis_output.txt` were written 09:02:39Z, after the last label row (09:02:30Z). The analysis ran once, on the final `scores.jsonl`; no `run.sh jev` ran after it. **Both lines are equal to each other and, for each of the 11 files they record (`go.mod`, `go.sum`, `internal/areas/areas.go`, `internal/clawpatch/records.go`, `internal/decisions/{audit,client,question,rank,run,triage}.go`, `internal/policy/consent.go`), to the digests in §11.2** (checked programmatically: 0 differing entries on either line; the two other code-under-test blobs, `scripts/clawpatch-safe.patch` and `scripts/build-clawpatch.sh`, are not part of the Jev driver and are not recorded in the lines; they are equal to round 5 in §11.2). Statement for every line: it precedes the analysis.

### 11.4 Run-order evidence (round-4 conditions and run order (a)–(g))

- (a) A dedicated worktree at the run commit: `/Users/adamjackson/Projects/inflight/scout-x2c`, `HEAD` = the run commit (checked for this report). Its toolchain is its own `.tools` (§12 deviation 1; `/tmp/x2c-bootstrap.log` ends "bootstrap-tools: ready: source /Users/adamjackson/Projects/inflight/scout-x2c/.tools/activate"; every `tool-versions-*.json` records Clawpatch 0.7.3 with the review-safety build marker, `ok` true).
- (b)–(f) Collection 00:29:40Z–08:12:03Z (`/tmp/x2c-collect.log`), replay (`data/replay.jsonl` written 08:12:23Z), pairs, then labelling (two invocations, §12) and Jev (two invocations at 08:12:36Z and 08:15:21Z, §11.3) in parallel (labels end 09:02:30Z), then the analysis once (09:02:39Z). Round-4 run order (e) allows labelling and Jev "in either order" (`jev` reads no labels); here their time spans overlapped (see §12 note 6).
- Validation before the run: `/tmp/x2c-validate.log` ends "X2c checks: 78 (plus X2's 233)" and "all validation checks passed" (78 + 233 = 311 PASS), as required by round 5's condition 2′; not re-run for this report.

## 12. Deviations and notes

Deviations from the round-4 run order (a)–(i), as recorded from the logs and files named; none changes a verdict, and no frozen file or code-under-test blob was changed (§11.2). Deviation 2 changes which agreement statistics exist (κ covers 260 of 1,986 real pairs).

1. **Toolchain bootstrap wrote to the main checkout first.** `scripts/bootstrap-tools.sh` in the new worktree first wrote to the main checkout's `.tools`, because `SCOUT_TOOLS_DIR` was inherited from an activated shell. It was re-run with every activation variable unset, so the worktree had its own `.tools` before any stage ran (`/tmp/x2c-bootstrap.log` ends "ready: source /Users/adamjackson/Projects/inflight/scout-x2c/.tools/activate"). The stages ran with the worktree's tools: every `tool-versions-*.json` records Clawpatch 0.7.3, the review-safety build marker (ok), `clawpatch_cli_sha256` `4eab263b…20eb`. (A `tools=` line is not present in `/tmp/x2c-collect.log` as retained; the evidence cited is the bootstrap log and the versions files.) `.tools/` is git-ignored.
2. **The first `run.sh label --workers 8` was interrupted and resumed in a new round.** It was stopped after 282 pairs had labels and call rows (260 real pairs and 22 controls; the agreement block of the analysis covers 260 real pairs and its control block 22). The operator interrupted it to consider a protocol change, then withdrew that change. Labelling resumed with `run.sh label --workers 8 --new-round resume-1`, the X2 §6 path (`/tmp/x2c-label-resume1.log`: "new round 'resume-1': 1943 pair(s) without a usable label … labelled 1943 pairs"; 282 + 1,943 = 2,225 = 1,986 + 239). **Label agreement (κ) therefore covers only round-1 real pairs (260), because new-round rows are excluded by design** (§7.1); agreement over all 1,986 real pairs is not reported. `labelled_pairs_without_call_row` = 0. The `labels.lock` left by the kill was deleted after confirming no labeller was running; no lock file is in the copied `data/` (and none would be tracked, X2b tracks none). `/tmp/x2c-label.log` (the interrupted run's log) is empty.
3. **`run.sh jev` was resumed once, verbatim, before the first analysis**, to retry the 12 schema failures ("choice probabilities incomplete or not normalised"); all 12 were answered (`/tmp/x2c-jev-resume1.log`). This is allowed by the run order (e), so it is recorded but is **not** a deviation. The first invocation exited with status 3 because of the 12 failures. Round-3 N5's inspection stop was not triggered (12 of 3,972 is not a `schema` surge). Cumulative billed cost 0.219195 USD of the 5 USD stop.
4. **Analysis-output inconsistency for the `+jev` probe lines.** The analysis prints `probe consistency … INCONSISTENT … not citable` for the four `+jev` pseudo-arms (§9.2). The X2 probe check looks for tool-versions files under the `+jev` arm name; the `+jev` cells reuse the base arms' maps, whose probes are `CONSISTENT`; the cell lines still print `citable: yes`. It does not change any verdict, because every cell is `fallback`.
5. **Served label models.** The label sessions were served `claude-opus-5-5` (A, the claude default), `gpt-6.1-sol` (B, the codex default) and `claude-sonnet-5-5` (C, 96 adjudications). The labeller models are those CLIs' defaults and were not pinned (PREREGISTRATION: "labellers use their CLIs' default models, reported by `label_models`").
6. **Concurrency of labelling and Jev.** `run.sh jev` ran 08:12:36Z–≈08:15:23Z (code-digest line 1 and the last `scores.jsonl` write), starting 33 seconds after the last after-probe, while labelling ran 08:12:40Z–09:02:30Z (first and last call timestamps in `data/pairs/calls.jsonl`). Run order (e) allows the two "in either order"; their time spans overlapped. `jev` reads no labels and the analysis ran after both finished. Recorded as a note, not a deviation. [INFERENCE: the overlap cannot influence the numbers because `jev` consumes only the candidate pairs.]
7. **Expectations in the preregistration that did not hold (rules unchanged, no verdict changed):** none found beyond the above. No stop records, no failed maps, no retried maps, no unlabelled pairs.

## 13. Reproduce

From the run worktree (`git worktree add ../scout-x2c 996541061e295101d197437544e6c0d31c085496`), toolchain per `README.md` bootstrapped with every activation variable unset (`git submodule update --init && scripts/bootstrap-tools.sh && source .tools/activate`). Verbatim from `README.md` "Reproduce" (the commands of the run order):

    source .tools/activate && evaluation/x2c-mapper-comparison/run.sh validate   # no model calls, no data

    SRC=/tmp/scout-x2c-smoke/_src      # any clone containing the pinned revisions; the collector verifies the SHA
    nohup evaluation/x2c-mapper-comparison/run.sh collect $SRC/fastapi $SRC/axum > x2c-collect.log 2>&1 &
    # once x2c-collect.log shows all 8 collections finished (map failures and stop records: data/stopped-*.json):
    evaluation/x2c-mapper-comparison/run.sh replay
    evaluation/x2c-mapper-comparison/run.sh pairs
    evaluation/x2c-mapper-comparison/run.sh label --workers 8        # appends data/pairs/{calls,labels}.jsonl
    infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- evaluation/x2c-mapper-comparison/run.sh jev
    evaluation/x2c-mapper-comparison/run.sh analysis                 # data/results.json, data/analysis_output.txt

As run: `run.sh collect` was as above; `run.sh label --workers 8` was interrupted and followed by `run.sh label --workers 8 --new-round resume-1` (§12 note 2); the `infisical … run.sh jev` command was run twice with the same verbatim command (§12 note 3); `run.sh analysis` once. Re-running the analysis on the existing `data/` makes no model call: `evaluation/x2c-mapper-comparison/run.sh analysis`. Model calls in the run: `claude -p` and `codex exec` (maps, probes, labellers; no API keys) and Jev through OpenRouter via Infisical (the one secret, injected into the one command; never printed). This document made no model call and re-ran no stage.

## 14. Threats

PREREGISTRATION §11 applies unchanged; restated with this run's numbers:

- **One run per arm and repository.** The verdict is one draw of each agent per repository. The luna-xhigh cell reproduces X2b's `fallback`, but the fastapi survival differs between the two draws of the same mapper (0.5721 in X2b, 0.4728 here); [INFERENCE] run-to-run variance is large on fastapi. The best cell (terra-none+jev, 0.789 on fastapi) is 0.011 from the minimum; a second draw could land on either side, and this study cannot say which.
- **Eight cells on shared labels, four arms.** No family-wise correction (PREREGISTRATION §8); §5 reports the per-cell false-pass figures and their sum of maxima (0.7159 at 0.08, 0.1861 at 0.10). No cell was eligible, so the false-support risk is moot for the outcome.
- **Labels are agent labels.** Labellers A and B and the Jev scorer answer the same question from the same text; Jev's agreement with the labels is an internal-consistency result, not proof of correctness. Labeller B is a GPT-6.1 model and two of the four mappers are GPT-6 family models (§7.1, inferred from names). C2 passes vacuously (zero `different`). The swapped controls missed 7 of 159 (§7.2), 6 of them in the 0.30 source-pair band; the strict rate `(different + overlapping)/labelled` is large at low thresholds in some cells (for example haiku axum 0.596 at 0.30; §3.1).
- **Haiku `none` is "the model default"**, not a fixed effort. The probe sees only the requested value; a server-side default change is invisible.
- **The coverage guard is a coarse proxy** (owned-file volume). Haiku fails it on fastapi (0.40 of the reference's mean owned files, floor 0.5); no cell adopted, so it changed no verdict. The area ratios (§6) show several arms map fewer, larger areas than luna-xhigh.
- **Jev time is an 8-worker estimate**, not an observed refresh time, and Jev cost and time were far inside the caps (≤ 0.0123 USD and ≤ 8.804 s per refresh, 70.434 s sequential). The synthetic planted Jev scores in `validation/` test arithmetic and wiring, not Jev.
- **Model names may roll** (`gpt-6-luna`, `gpt-6.1-terra`, `claude-haiku-5-5`); probes bracket each collection but cannot see a same-name change. The collection took about 7.7 hours across the eight sequential collections.
- **Nothing below 0.30 was measured**; this study does not say whether a lower threshold would pass. [INFERENCE] With the candidate ceilings in §1, a lower Jaccard threshold cannot lift fastapi above the ceiling of 0.6336 (haiku), 0.7787 (luna-xhigh) or 0.7622 (terra-low), so only terra-none (0.9592) could in principle reach 0.8 there.
- **Knowledge of X2/X2b/smoke** (disclosure in the preregistration). Fixture sources moved; the pinned revision SHAs, not paths, were verified by the collector.

## 15. Consequence for the design

`no-adoption` leaves design revision 2.7 as is: the agent map runs at onboarding and on manual `scout map --refresh`; there is no refresh-on-merge; `mapping.areas.match_jaccard` stays as in the design. X2c adopts nothing and proposes no change to `docs/scout-design.md`. HD5c: `not-supported` (§1); HD5 and HD5b unchanged. The ledger History line (run order (h)) and the design-document consequence are steps after this report.

Companion file: `data/analysis_output.txt` (the complete analysis text; sha256 `9dc72a11146ac77963fd73ed23a770a1243075c0fd1ed5ae3d888785b6a3ec61`), `data/results.json` (sha256 `831ff555a61480260265fd0f89731ad1d41fff0d54850b812c55d13a1bfad4bc`).
