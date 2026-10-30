# X2d — merge-association pilot: RESULTS

**Status: the pilot stopped at round 0.** The round-0 gates P1 and P2 failed. The one re-gate permitted by PREREGISTRATION rev 13 (R-053) then did not pass: **P1 FAIL, P2 PASS**. Rounds 1–10, threshold tuning and the §9 verdict never ran.
**HD6 → `inconclusive`; HD6b stays `proposed`; HD7 `proposed`. The pilot gives no GO and recommends no full trial of this instrument as frozen** (§1.4).

Paths. `<data>` = `../scout-x2d/evaluation/x2d-merge-association-pilot/data/` (the run worktree; `data/` is untracked and was only read to write this file). `<study>` = `evaluation/x2d-merge-association-pilot/`. `regate/` = `<data>/rounds/0/regate/` (so `regate/gates.json` is `<data>/rounds/0/regate/gates.json`). Short area ids are the first 6 characters of the 12-character ids. Intervals are Wilson 95 % unless stated, and treat items as independent (§0.6). Where a figure is not stored in a data file it was computed from the stored rows by the commands in §10, and the table says so. `[INFERENCE]` marks anything that is my reading and not a measured fact.

## 1. Summary

### 1.1 What was tested
Scout keeps a catalogue of **areas** (named groups of files) for a repository. HD6 asks whether Jev (a model decision service, pinned `typesafe/jev-1.14-20261014`) can, when a merged PR adds files that no area owns, assign them to the right existing area **by meaning rather than by folder**. To count as good it had to be right on at least 80 % of the files it answers, decline to answer on at most 15 %, and beat a folder-based baseline. The plan was a 10-round pilot on fastapi merges from 2025-01 to 2026-10, ending in a go/no-go on whether a full trial is worth running. Before round 1, **round 0** had to pass a set of checks called **gates**, so that the window data were not spent on an instrument that was already broken.

| Term | Meaning here |
|---|---|
| **P1** ("move a file") | Take a file whose owner area A is known. Put it in the catalogue as a *new* file in a folder whose neighbours belong to a different area B. Ask Jev which area it belongs to. Jev passes if it names A (or an area the human labellers also accept) and does not follow the folder. |
| **Stratum** | How the folder's neighbours are owned. **1** = all owned by one other area (the hard case, and the one that decides P1). **2** = owned by several areas. **3** = no owned neighbours. |
| **P2** ("remove an area") | Delete area A from the catalogue, then ask about a file A owned. If no area is left that fits, the right answer is `none`. |
| **Answered / abstained** | Jev *answers* with an area or `none`. It *abstains* (`unresolved`) when its top choice has probability under the provisional 0.70 threshold (`below_t_choice`) or when the two option orders disagree (`permutation_disagreement`). |
| **B1, B2** | Folder-based baselines (nearest owned directory; B2 adds a test-mirror rule). |
| **Gold** | The labellers' agreed answer. Labellers are two blind agent sessions (`claude -p`, `codex exec`) plus an adjudicator; provenance `agent_reviewed`. |
| **Wilson 95 %** | A confidence interval for a proportion. |

### 1.2 What happened
- **Onboarding worked.** 16 Clawpatch agent maps (15 scoped + 1 contrast) were consolidated into **114 areas** (catalogue sha256 `1220016c…4199`) for $0.185 of Jev. The intruder audit found the odd file out on 40/40 items for both labellers (κ 1.0) (§3).
- **Round 0 passed most checks and failed two.** Passed: K2, K3, K4, K5, K6/N1 (Jev follows option text, not position or id), the planted new concern (3/3 `none`), and every power cell. **P1 failed** (Jev recovered the original area on 9 of 20 files, 0.45, bar 0.60; baselines 0.05). **P2 failed** (Jev answered `none` on 4 of 7 label-derived gap positives, 0.57, bar 0.80) (§4).
- **Part of both failures was a construction fault, not Jev.** P1's gold was one area while the labellers accepted several; P2's gold was *derived* and contradicted the clones' own blind labels on all 3 misses; one file was counted twice (a path twin). Part was real: in stratum 1 Jev recovered the original area on only 2 of 11 files (§4.3–4.4).
- **A $1 Jev cost stop** (measured cost per request ≈ 1.7× the budget assumption for the fan-out and ≈ 10.7× for the other questions, §7) and the rev-12 caps ($2 per round, $20 total) came first (§8).
- **One repaired re-gate on fresh legacy items (rev 13; 2019–2022 files, not window-era files, §0.3).** P1: 40 items, 35 eligible in stratum 1. P2: 40 clones, 12 direct-label positives. Result (§6):
  - **P2 PASS:** `none` on 9 of 11 answered positives = 0.818 [0.523, 0.949] (bar 0.80); abstention 1/12 (cap 0.15). It passes by one answer: if any one of the nine `none` answers changed, P2 would fail. Becoming an assignment gives 8/11 = 0.727; becoming `unresolved` gives abstention 2/12 = 0.167 > 0.15. Four of the nine `none` answers have `1 − p_true` of 0.31–0.33 (ctl-26 and ctl-70 at 0.31, ctl-10 and ctl-41 at 0.33), within 0.03 of the 0.30 cut (§6.9).
  - **P1 FAIL:** stratum 1, accuracy on answered items 14/20 = 0.700 [0.481, 0.855] (bar 0.80); **abstention 15/35 = 0.429 [0.280, 0.591] (cap 0.15)**; baselines 0/38 over all eligible items, the decisive conjunct (0/35 in stratum 1; bar ≤ 0.20; zero by construction, §1.3). Two of the four conditions failed.
  - Of the 6 wrong answers in stratum 1, 5 were `none` and 1 named a sibling area. Path-following: 1/20 of answered items; 6/35 of unthresholded first choices, 4 of which abstained (§6.4).
- **Rule: no second re-gate.** The pilot stopped; HD6 was never under test on window merges.

### 1.3 What it means
The pilot did not show that Jev can do this, and it did not show that it cannot. It measured the *entry gates* and the instrument as frozen (v1 questions, untuned 0.70 thresholds, the 114-area catalogue) did not clear them: on files that sit in a folder owned by another area (legacy 2019–2022 files, not window-era files), Jev abstained far more than the 15 % cap, and when it did answer it was right 70 % of the time, with an interval that straddles the 80 % bar. The folder-based baselines recovered no acceptable area (0/38 eligible items), but that is built in: the rev-13 build rejects any target directory for which B1 or B2 would still answer A (`PREREGISTRATION.md:112`), and the pre-Jev check showed 0/38 before any Jev request (`reviews/pilot-round-0.md:222`). [INFERENCE] The correct answers did not follow the folder (no acceptable area is a sibling area, `sibling_in_acceptable` 0; composed path-following 1/20). Which cues produced them is not measured: each item also showed Jev its file name and the PR text (`PREREGISTRATION.md:111`), and the raw first choice named the sibling area on 6/35 items (§6.4).

### 1.4 Go/no-go for a full trial (the pilot's stated output)
- **No GO.** The §9 verdict was never computed (no `verdict.json`; rounds 8–10 did not run). The §14 map for a re-gate that is not PASS for both P1 and P2 gives HD6 `inconclusive`, never GO (`PREREGISTRATION.md:253`).
- **[INFERENCE] The pilot recommends no full trial of this instrument as frozen.** Nothing measured here shows the instrument clears its own entry bars. The two failing quantities (abstention 0.43 against the 0.15 cap; accuracy on answered items 0.70 against 0.80 (re-gate, legacy 2019–2022 files)) are the ones the GO rule itself bounds (`u ≤ 0.15`, `R_assign_res ≥ 0.80`, `PREREGISTRATION.md:112`).
- **[INFERENCE] Evidence needed before any full trial**, each tied to a figure above:
  1. **Abstention under tuned thresholds.**
     - Of the 17 abstentions among the 38 eligible re-gate P1 items, 15 were `below_t_choice` and 2 `permutation_disagreement`; in decisive stratum 1 the split was 13 and 2 (§6.8).
     - The 5 wrong `none` answers in stratum 1 come from the covered threshold `t_cov`, not from `t_choice` (`PREREGISTRATION.md:129`).
     - Of the 11 round-0 P1 misses, 5 had raw choice A: 3 abstained under `t_choice` (ctl-3, ctl-26, ctl-34) and 2 were `none` under `t_cov` (ctl-6, ctl-25) (§4.3).
     - No threshold was tuned (rounds 1–8 never ran). The stop memo's sensitivity lowered both thresholds to 0.5 and found it would cancel three round-0 P2 `none` answers (`reviews/pilot-round-0-gate-stop-memo.md:104`).
     - [INFERENCE] P1's `none` errors and P2's `none` answers pull `t_cov` in opposite directions, while P1's abstentions depend mainly on `t_choice`. The data therefore cannot say whether tuned thresholds meet the 0.15 cap without breaking P2.
     - A new preregistration would have to calibrate both thresholds on separate labelled items first.
  2. **Enough answered stratum-1 items from more areas.** The 35 eligible stratum-1 items come from 26 areas and the 20 answered items from 19. Accuracy is 14/20 = 0.700 [0.481, 0.855]; one answered item per area gives 13/19 = 0.684 [0.460, 0.846] (§6.6).
  3. **Window-era items.** The re-gate population is legacy 2019–2022 files (§0.3); nothing here speaks to 2023–2026 merges.
  4. **Near-duplicate areas resolved or disclosed (HD7).** Two areas are titled "Dependency injection" (`c3534a`, `9a4a0e`); both re-gate stratum-1 items whose original area is one of them abstained (§6.6).
  5. **New-area detection.** P2 passes by one answer from 12 positives (any one of its nine `none` answers changing fails it, §1.2), and both wrong assignments sit in one ablated area (`a0f998`, §6.9), so HD6b (new-concern detection) stays `proposed`.
- [INFERENCE] Cost is not the obstacle. Known Jev spend to this point is $1.377808, a lower bound because 7 cancelled requests have unknown cost (§7). The whole pilot was projected at ≈ $9.74 of Jev (R-052) against a $20 cap (§7).

## 0. Threats and scope statements the preregistration requires
(Numbered §0 because `PREREGISTRATION.md` §14 requires them there; the plain-language summary comes first.)

Sources: `PREREGISTRATION.md:108-114, 215-228, 249, 255`; `reviews/pilot-round-0.md:208`; `reviews/round-18.md:80-85`.

1. **P1 and P2 test only cleanly owned units.** A unit can be a P1 item or a P2 source only if every member file has exactly one owning area. Docs-config files, unscoped files and contested files are untested by these gates. The reserve shows how much is excluded: of the 30 P units, 4 had no single owner (`docs/en/mkdocs.insiders.yml` and `docs/en/overrides/partials/copyright.html`, outside every onboarding scope; `docs/language_names.yml`, contested; `tests/test_webhooks_security.py`, unowned), leaving 26 eligible against the 28 the design needed (`PREREGISTRATION.md:249`; `reviews/pilot-round-0-p2-gate-memo.md:13-22`). Onboarding coverage of those roles is in §3 (docs-config: 1 of 28 files area-owned).
2. **D clones are source-dependent (round 0).** Six of the twelve rev-12 P2 clones were cloned from set-D units, and the rev-12 gold of every P2 clone was *derived* from its source's labels (`derive_ablation`). A label error in a D source would propagate to its clone, so D clones are not an independent replication. No inherited D label error was observed (the one D miss, `ctl-35`, had a correct source label; the fault was the derivation, §4.4). In the re-gate the gate gold is each clone's **own** blind label, so the source dependence is removed from the decisive figure. The derived gold is only reported beside it (`derived_vs_direct`, §6.9).
3. **The re-gate population is legacy 2019–2022 files, not window-era files.** All 40 P1 items are `in_no_unit` clusters (C0-owned files that belong to no reserve unit) dated 2019-03-03 to 2022-09-01 (2019: 22, 2020: 9, 2021: 5, 2022: 4 by year). Of the 40 P2 clones, 36 are `in_no_unit` clusters dated 2019–2022 (2019: 15, 2020: 11, 2021: 3, 2022: 7) and 4 are dated 2023–2024 (D 2, G 2). C0 is dated 2024-12-30, so the items are 2–6 years older than C0 and than the rev-12 P items (`reviews/regate-supply.md:49-51`). **Every re-gate figure in §6 is a test of the repaired instrument on legacy files. It is not a re-test of the rev-12 population and is not evidence about window-era files (2023→2026).** Every 2023→C0 cluster was already in the reserve and so was excluded as an exposed source.
4. **The repairs were outcome-informed.** In writing the stop memo the council and the drafter saw the per-item round-0 Jev answers, their choice and covered probabilities, the P1 labelled sets and the P2 clones' direct labels (`reviews/pilot-round-0.md:208`; `PREREGISTRATION.md:114`). Each repair rests on a fact that does not involve Jev's answers (gold narrower than the labels, the baseline leak, derived gold contradicting the clones' labels, the path twin, the thin denominator), the evidence is fresh items only, no bar moved towards the round-0 data, and it was one shot. **Which constructs were repaired was nevertheless guided by where Jev failed.** No other Jev-versus-label comparison was made before the re-gate.
5. **Re-gate bars and easing, disclosed (R17-04).** P2's decisive rate moved to *answered* positives with a separate abstention cap, which is easier on abstention than rev 12 (at abstention 0.15 the bars imply `none` on ≥ 0.68 of all positives against rev 12's 0.80 with abstention a miss). P1's bars imply an acceptable-set recovery of at least 0.68 on stratum 1, but this is not dominance over rev 12: the gold is wider than rev 12's single A and the population and strata differ. P2 stays binding although R-051 made new-area detection non-vetoing in the GO rule (council-flagged, ruled in round 17/18) (`PREREGISTRATION.md:112-113`).
6. **Items are not independent.** The 20 rev-12 P1 items come from 13 PRs. The 35 eligible re-gate stratum-1 items come from distinct PRs but only 26 areas, and area `e287e9` holds 5 and `9f08e6` holds 3 (§6.6). Wilson intervals throughout treat items as independent and so are too narrow. K2 and K3 are over 330 clusters from fewer PRs, also treated as independent.
7. **Labels are agent labels** (R-039, `agent_reviewed`) from a Claude labeller (`claude-opus-5-5`), a GPT labeller (`gpt-6.1-sol`) and a Claude adjudicator (`claude-sonnet-5-5`); the mapper is codex `gpt-6-luna`. The wide-evidence labellers saw more than Jev (up to 300 lines and 20 files per area against Jev's 60 lines and 5 sample files; `reviews/pilot-round-0.md:205`). Agreement is an internal-consistency result, not proof.
8. **The Jev side was never tuned.** Thresholds were the provisional 0.70 (`rounds/0/thresholds.json`, status `provisional-0.7`, no cases per label) because `tune` runs from round 1. Both the round-0 gates and the re-gate ran on those thresholds; the re-gate froze them by design.
9. **Near-duplicate areas across scope shards.** The onboarding duplicate check compared only same-scope pairs whose titles share at least one token (`onboard/main.go:25`): 261 of the 1,268 same-scope pairs, 0 flagged. Cross-scope pairs were never compared (R-054; HD7). Two areas share the title "Dependency injection" (`9a4a0e` core, `c3534a` tutorial-1) (`<data>/catalogue/consensus-report.json` `.jev.duplicates`, `.jev.duplicate_pairs_checked`).
10. **One repository, one catalogue, one model pin.** Nothing transfers beyond fastapi at C0 `1d50bad4…` with `typesafe/jev-1.14-20261014`.
11. **Measurement and construct checks made while drafting** (disclosed in `PREREGISTRATION.md:22-26`; none touched a label, a Jev score or any predictor–outcome relationship): the model-free pre-pass, one real agent map (the Clawpatch scope-lever check, codex `gpt-6-luna` `xhigh`, scope `fastapi/**`), an illustrative power calculation with assumed accuracies, and an eligible-PRs-per-year yield count.
12. **`R_new` on catalogue gaps has a planted test** (rev-12 P2 FAIL on derived gold; re-gate P2 PASS on 12 direct-label positives). The sentence "R_new on catalogue gaps has no planted test" applies only when P2 is UNMEASURED, which it never was.

## 2. What ran, in order (run record)
Source: `reviews/pilot-round-0.md` (line numbers in the table); stage logs in `/tmp/x2d-*.log` (not committed; §10).

| # | Date (2026-10-30) | Stage | Commit | Result | Record |
|---|---|---|---|---|---|
| 1 | 12:40–13:10Z | `onboard-collect` (16 maps) | `730674e` | 16/16 rc 0, 1 attempt each, 71.6–153.7 s per map | `pilot-round-0.md:15`; `<data>/onboard/versions.json` |
| 2 | 13:10Z | `onboard` (consensus, Jev, freeze) | `730674e` | 114 areas, 17 contested, 526 unowned, 440 assigned by Jev; sha `1220016c…4199` | `/tmp/x2d-onboard.log` |
| 3 | 13:13Z | `quality` + intruder audit | `730674e` | 40/40 both labellers, κ 1.0, 80 sessions | `pilot-round-0.md:15`; `/tmp/x2d-quality.log` |
| 4 | 13:13Z | `gates` (aggregate) | rev-9 sign-off | **refused at the P2 build** (26 eligible P units leave 6 for P2, 8 needed); nothing written | `/tmp/x2d-gates.log`; `pilot-round-0.md:16` |
| 5 | 16:04Z | `gates-build` | `0b56594` | matched the round-10 table on every item | `pilot-round-0.md:19-42` |
| 6 | 16:13Z | `gates-label` | `0b56594` | 116 cases; 317 of 429 sessions; labels-check ok | `pilot-round-0.md:46` |
| 7 | 16:13–16:17Z | `gates-jev` | `0b56594` | **cost stop, exit 4**: 7,755 of 7,872 `main` requests, $1.0001 | `pilot-round-0.md:47` |
| 8 | before row 9 | re-freeze of `rounds/0/version.json` | `b6a5bf3` | exact key diff as required | `pilot-round-0.md:54-62` |
| 9 | 18:30Z | `gates-jev` (resumed) | `b6a5bf3` | all 4 runs complete, 0 failed | `pilot-round-0.md:64-83` |
| 10 | 18:31Z | `gates-analyse` | `b6a5bf3` | exit 1: **P1 and P2 FAIL** | `pilot-round-0.md:85-99` |
| 11 | 18:31Z | `power` | `b6a5bf3` | exit 0, `n4_max` 0.054, all checks true | `pilot-round-0.md:101` |
| 12 | 22:54Z | `regate-build` | `8c7148c` | 40 P1 items (strata 37/0/3), 40 P2 clones in 16 areas; plan equals `reviews/regate-supply.json` byte for byte | `pilot-round-0.md:215-216` |
| 13 | before row 14 | `regate-label` + pre-Jev check | `8c7148c` | 120 cases; 246 of 330 sessions; precheck clean (§6.1) | `pilot-round-0.md:220-222` |
| 14 | 23:02Z | `regate-jev` | `8c7148c` | 87 + 87 requests, 0 failed, $0.0939 | `pilot-round-0.md:223` |
| 15 | 23:02Z | `regate-analyse` | `8c7148c` | exit 1: **P1 FAIL, P2 PASS**; pilot stops | `pilot-round-0.md:224-233` |

Times are in-log UTC stamps or log-file mtimes (which agree where both exist); rows 8 and 13 have no stamp. The first `gates` attempt (row 4) ran in the same session as rows 1–3, under the rev-9 sign-off; the record does not name a separate commit for the refused attempt.

## 3. Onboarding and quality (the §4.3 report the preregistration places in §0)
Sources: `<data>/onboard/quality.json`, `quality.intruder.json`, `versions.json`; `<data>/catalogue/consensus-report.json`, `catalogue.sha256`; `<data>/onboard-jev/scores.jsonl`; `/tmp/x2d-onboard-collect.log`, `/tmp/x2d-onboard.log`, `/tmp/x2d-quality.log`. Run at `730674e`. Label-free except the intruder audit.

**Maps and catalogue.** 16 maps ran (15 scoped = 5 scopes × K=3 draws, plus 1 contrast draw), all `rc 0`, one attempt each. Clawpatch 0.7.3, `codex-cli 0.164.0`, `--source agent --provider codex --model gpt-6-luna --reasoning-effort xhigh`; 12:40:49Z to 13:10:27Z (sum of map times 1,775 s). The code-first consensus (quorum 2 of 3, τ = 0.5) gave **114 areas** in 5 scopes (core 22, tutorial-1 29, tutorial-2 21, tutorial-3 22, infra 20; `consensus-report.json` `.scopes`), **17 contested** files and **526 unowned** files. Jev placed the contested and unowned files: **543 files asked, 440 assigned, 74 below the covered threshold, 29 below the choice threshold**, on 1,534 answered requests. The duplicate check found **0 duplicates in the 261 same-scope pairs whose titles share a token (of 1,268 same-scope pairs)**. **Catalogue sha256 `1220016c26115ccd3e3a521de76ec45e68893f4ab6edf0f4889e8c8ce2c84199`** (file and recomputed digest equal).

**Onboarding Jev cost: $0.185375** (1,534 answered requests, 0 failed, 0 unknown-cost rows) (`onboard-jev/scores.jsonl`, summed `cost_usd` over answer rows; `consensus-report.json` `.jev.billed_cost_usd` 0.185374518).

**Intruder audit** (agent protocol, `agent_reviewed`): 40 items; both labellers (A `claude-opus-5-5`, B `gpt-6.1-sol`) found the intruder on **40/40 = 1.000 [0.912, 1.000]**, consensus 40/40; **κ(A vs B) = 1.0**; unmeasured 0 (`quality.intruder.json`; labels sha256 `0ff91dba…9d7840`). 80 sessions (cap 110 at the time). Note: `quality.json` `.intruder_audit.status` still reads "items written; labels pending" because it was written before labelling; the result is in `quality.intruder.json`.

**Coverage by file role** (C0 inventory: 2,572 files, 811 source, 537 test, 1,355 eligible; `quality.json` `.coverage_by_role`):

| role | heuristic-owned only | area-owned only | both | unowned | files |
|---|---|---|---|---|---|
| core | 0 | 0 | 44 | 1 | 45 |
| example | 0 | 685 | 0 | 2 | 687 |
| test | 91 | 258 | 181 | 7 | 537 |
| docs-config | 0 | 1 | 0 | 27 | 28 |
| infra | 0 | 12 | 33 | 0 | 45 |
| config | 1 | 9 | 2 | 1 | 13 |

- **Core gate (decided once): not enabled.** 44 of 44 core `fastapi/` source files are area-owned (share 1.000 ≥ 0.50), so core units stay in the GO population (`quality.json` `.core_gate`; `rounds/0/gates.json` `.core_gate`).
- **Area-owned share at C0 inventory position > 500 (the H-W blind spot):** 1,196/1,307 = 0.915 (positions ≤ 500: 29/48 = 0.604); by source-file position: > 500 309/311 = 0.994, ≤ 500 464/476 = 0.975 (`.area_owned_share_position_gt_500`).
- **Areas at the 60-file cap:** 2 of 114 at or over the cap. **Overlap** (files in ≥ 2 areas): 0 of 1,225 owned files (share 0.000).
- **Co-change lift** (first-parent commits before C0, since 2023-01-01; 145 of 2,773 commits used, 8 skipped for > 30 owned files): observed same-area probability 0.345 against a size-preserving permutation null 0.0225 (sd 0.0045, 200 permutations): **lift 15.3**, permutation p = 0.005.
- **Leaf-directory entropy:** 178 leaf dirs, 1,225 files; H(area | leaf dir) = 0.339 nats, H(leaf dir | area) = 0.659 nats.
- **Granularity:** 114 areas; files per area median 6, quartiles [2, 6, 12.75], max 93, **14 singletons**; exp(H) = 64.6.
- **Reproducibility between draws** (median over scope pairs): owned-file Jaccard 0.956, nestedness 0.789 / 0.946, adjusted Rand index 0.786, VI split H(X|Y) 0.000 / H(Y|X) 0.226. Single-draw pooled ARI pairs 0.996, 0.823, 0.836 (median 0.836). Weakest pairs are `tutorial-3` draws 1–3 (ARI 0.441) and 2–3 (0.470) (`.reproducibility`).
- **UNMEASURED: split-half consensus ARI.** All three rotations returned `unmeasured` (`go consensus failed rc=1: scope core: draw 3 is missing … the quorum needs K=3 draws`); the H-G prediction that the consensus ARI exceeds the single-draw median is therefore untested (`.split_half_consensus_ari.rotations[*]`). [INFERENCE] a harness limitation of the held-out-draw rotation, not a catalogue property. Also unmeasured: 1,217 inventory files excluded by policy (e) (`.unmeasured`).
- **κ.** Intruder κ above. Round-0 labelling κ is in §4.5 (K4: κ_answer 1.0 on 20 D units; labels-check 1.0 on 50 units). The §4.3 catalogue stratum per unit is carried in `units.jsonl` `stratum`; no per-round strata table exists because no round ran.

## 4. Round-0 gates (rev 12, run on the signed-off rev-12 bytes)
Gate statuses: `<data>/rounds/0/gates.json` `.gates.*` (SHA-256 `d7c1b98b…bd48e`, confirmed by `regate/regate.json` `.rev12_gates_sha256`); log `/tmp/x2d-gates-analyse.log` (`pilot-round-0.md:88-98`). `fan_out_primary` False. Core gate not enabled. Intervals not in the file were computed with the Wilson formula used by the analysis code (the re-gate intervals in §6 match the file to 4 decimals).

### 4.1 Gate table
| gate | status | value (k/n [Wilson 95 %]) | bound | n |
|---|---|---|---|---|
| K2 | PASS | B0 assigns on 193/330 = 0.585 [0.531, 0.637]; B2 assigns on 266/330 = 0.806 [0.760, 0.845] | stop if B0 share ≥ 0.90 | 330 residual clusters, rounds 1–10 (membership only) |
| K3 | PASS | stratum-3 clusters 17/330 = 0.052 [0.032, 0.081]; `new_dir` 93 and `core_gate` 0 reported apart | stop if share > 0.40 | 330 |
| K4 | PASS | κ_answer 1.0; κ_covered 1.0; raw agreement (answer) 20/20 = 1.000 [0.839, 1.000]; prevalence of `covered` 1.0 | κ_answer ≥ 0.6 | 20 D units |
| K5 | PASS | order-flip rate 0.04 = 2/50 = 0.040 [0.011, 0.135] (k = 0.04 × 50, derived) | flip rate ≤ 0.15 | 50 units (D ∪ P) |
| K6/N1 | PASS | Jev accuracy on the shuffled catalogue 0/50 = 0.000 [0.000, 0.071]; one-sided 90 % Wilson upper bound 0.0318; K = 114; B2 unaffected; 0 units without an arm row | UB90 ≤ max(2/K, 0.20) = 0.20 | 50 gold-covered units |
| planted_new | PASS | `none` on planted new concerns 3/3 = 1.000 [0.439, 1.000] (all scored `none_correct_new`) | `none` rate ≥ 0.80 | 3 |
| **P1** (rev 12) | **FAIL** | V-sem recovers A on 9/20 = 0.450 [0.258, 0.658]; B1 1/20 = 0.050 [0.009, 0.236]; B2 1/20 = 0.050 [0.009, 0.236]; `rules_resolved_units` 0 | V-sem ≥ 0.60 and baselines ≤ 0.20 | 20 items |
| **P2** (rev 12) | **FAIL** | `none` on label-derived gap positives 4/7 = 0.571 [0.250, 0.842] (4 scored `none_correct_gap`); retained-covered clones: 1 of 5 assigned correctly; 12 clones; `no_agreed_gold` 0 | `none` ≥ 0.80 over ≥ 8; below 8, FAIL if ≥ 2 misses | 7 gap positives (below the floor of 8; 3 misses → FAIL, not UNMEASURED) |
| real_none_proxy | INFO | 39 of 98 frozen-round (8–10) clusters (39/98 = 0.398 [0.307, 0.497]), descriptive share; indicates, does not prove, gold-`none` positives available | none (INFO, never a stop) | 98 |

`gates-analyse` exit 1: `GATES FAIL: P1, P2`. The failure of P1 and P2 stopped the pilot before round 1 (`PREREGISTRATION.md:116`). Rev 13 allowed one re-gate of P1 and P2 only, because no gate other than those two failed.

### 4.2 K2 and K3 by round (membership only; `rounds/0/k23.json` `.per_round`)
| round | clusters | B0 unique owner | B2 assigns | stratum 3 | `new_dir` |
|---|---|---|---|---|---|
| 1 | 23 | 12 (0.522) | 14 (0.609) | 2 (0.087) | 0 |
| 2 | 25 | 16 (0.640) | 23 (0.920) | 0 (0.000) | 8 |
| 3 | 22 | 18 (0.818) | 20 (0.909) | 0 (0.000) | 3 |
| 4 | 26 | 14 (0.538) | 26 (1.000) | 0 (0.000) | 12 |
| 5 | 49 | 32 (0.653) | 36 (0.735) | 0 (0.000) | 16 |
| 6 | 59 | 47 (0.797) | 56 (0.949) | 6 (0.102) | 6 |
| 7 | 28 | 8 (0.286) | 15 (0.536) | 0 (0.000) | 18 |
| 8 | 25 | 8 (0.320) | 10 (0.400) | 1 (0.040) | 7 |
| 9 | 23 | 4 (0.174) | 16 (0.696) | 1 (0.043) | 14 |
| 10 | 50 | 34 (0.680) | 50 (1.000) | 7 (0.140) | 9 |

The frozen rounds 8, 9, 10 have K2 (B0) 0.320, 0.174, 0.680 and K3 0.040, 0.043, 0.140. K2 counts the clusters on which a baseline *produces* an assignment, not whether it is right, so it bounds the room Jev has to add value and says nothing about baseline accuracy. B2 assigns on 0.806 of all clusters (0.40–1.00 per round; 0.92–1.00 in rounds 2, 3, 4, 6 and 10). The new-directory share (`new_dir` 93 of 330) is reported apart from K3 as the preregistration requires.

### 4.3 P1 per item (the preregistered descriptive disclosure) and counts
Table copied from `reviews/pilot-round-0.md:115-136` (columns defined at `:109-113`). I re-derived every row's stratum, A, answer, raw choice, labelled set and B from `rounds/0/control-key.jsonl`, `units.jsonl`, `arms-p1/V-sem.jsonl`, `arms/B1.jsonl`, `arms/B2.jsonl` and the final rows of `labels.jsonl`. Two corrections († below; and ctl-23's confidence, 0.92 in `arms-p1/V-sem.jsonl` against the record's .91): no other mismatch.

| Item | Stratum | File moved | A | Answer (conf.) | Raw | B | Labelled set | Category |
|---|---|---|---|---|---|---|---|---|
| ctl-1 | 3 | separate-schemas tests | ad8836 | **A** 1.00 | A | e287e9 (`tests/` fallback)† | A, e287e9 | hit |
| ctl-2 | 1 | `tests/test_forms_single_model.py` | 3de70a Parameter declarations | 158cb3 Request form model tutorials 0.91 | 158cb3 | 3e0c34 Advanced middleware | **158cb3**, A, 7d2013 | gold narrower than labels |
| ctl-3 | 3 | dependency 008c tests | c3534a Dependency injection (tutorial) | unresolved 0.44 | A | e287e9 (`tests/` fallback)† | 9a4a0e, A | threshold; probability split with the core area also titled "Dependency injection" |
| ctl-6 | 2 | playwright `image02.py` → `.github/workflows` | ef2b9b | `none` (p_true .67) | A | tie | A | covered threshold |
| ctl-8 | 1 | `tests/test_ws_dependencies.py` | 36295a | **A** .91 | A | ade7d5 | A, 9a4a0e | hit |
| ctl-12 | 2 | `fastapi/cli.py` → `fastapi/security` | 1c4aa7 | **A** 1.00 | A | 637a9b | A | hit |
| ctl-13 | 1 | form-model test | 158cb3 | **A** .86 | A | 749bd1 | A, 3de70a | hit |
| ctl-20 | 1 | `docs_src/request_form_models/…` → `docs_src/settings/app01` | 158cb3 | unresolved (the two option orders disagree) | 7d2013 Request forms | 9ce9c5 | A, 3de70a, **7d2013** | near-duplicate areas |
| ctl-21 | 2 | dependency 008d tests | c3534a | **A** .83 | A | e287e9 | 9a4a0e, A | hit |
| ctl-23 | 3 | `tests/test_additional_properties_bool.py` | e287e9 | **A** .92 | A | e287e9 (baselines fall back to `tests/`, which A owns) | A | hit; baseline leak |
| ctl-24 | 1 | recursive response tests → `test_header_params` | 617fde | unresolved (orders disagree) | **B** | ea1d02 Header parameters | A | path pull |
| ctl-25 | 1 | playwright `image01.py` → `notify-translations` | ef2b9b | `none` (p_true .66) | A | d1459f | A | covered threshold |
| ctl-26 | 1 | `docs_src/dependencies/tutorial008d*` → `docs_src/templates` | c3534a | unresolved 0.52 | A | 29a30e | 9a4a0e, A | threshold; duplicate "Dependency injection" split |
| ctl-27 | 2 | `scripts/mkdocs_hooks.py` → `.github` | 30399e | **A** 1.00 | A | tie | A | hit |
| ctl-31 | 2 | schema-example tests → `tests` | 95b28e | **A** .98 | A | e287e9 | A, e287e9 | hit |
| ctl-33 | 1 | playwright cookie-models `image01.py` → `.github/actions/people` | f0abaf | unresolved 0.49 | **B** | 342c6b People GitHub Action | A | path pull (real) |
| ctl-34 | 1 | `docs_src/sql_databases/tutorial002*` | 8a632c | unresolved 0.69 | A | 3d2f5f | A | threshold (0.01 short) |
| ctl-36 | 3 | header-models test | b58662 | **A** .96 | A | e287e9 (`tests/` fallback)† | 3de70a, A | hit |
| ctl-39 | 1 | `docs_src/header_param_models/tutorial002*` → `bigger_applications/…/internal` | b58662 | unresolved 0.68 | **B** | 2874a6 Bigger applications | 3de70a, A | path pull (real) |
| ctl-46 | 1 | `tests/test_response_model_data_filter_no_inheritance.py` | 617fde Response model validation | 10c16a Response models .86 | 10c16a | 233f9b | **10c16a**, A | gold narrower than labels |

† The record's B column reads `(none)` on the three stratum-3 items ctl-1, ctl-3 and ctl-36. The stored B1 and B2 rows are `assigned e287e9` on all three (the same `tests/` fallback the record names for ctl-23, whose A is `e287e9`). Corrected here. No conclusion changes: B1 and B2 recover A only on ctl-23, 1/20.

**Counts** (re-derived from the same files):
- 9 hits (9/20 = 0.450 [0.258, 0.658]). Abstained (`unresolved`): 7 of 20 (7/20 = 0.350 [0.181, 0.567]); `none` answers: 2 (ctl-6, ctl-25). 1 composed answer equals B (ctl-23, where B = A through the baseline leak); 0 do when B ≠ A.
- Raw (unthresholded permutation-1) choice equals A on 14/20 = 0.700 [0.481, 0.855] and lies in the labelled set on 17/20 = 0.850 [0.640, 0.948]; equals B on 4/20 = 0.200 [0.081, 0.416]: ctl-23 (B = A, the leak) and the three path-pull items ctl-24, 33, 39 (raw = B ≠ A on 3/20 = 0.150 [0.052, 0.360], the record's figure).
- By stratum: **stratum 1 recovered 2/11 = 0.182 [0.051, 0.477]**; strata 2–3 recovered 7/9 = 0.778 [0.453, 0.937] (stratum counts 11 / 5 / 4).
- The 20 items come from 13 PRs, so the effective n is below 20.
- Against the labellers' agreed acceptable sets (descriptive, not a re-score) V-sem scores 11/20 = 0.550 [0.342, 0.742], which still fails 0.60. On the same sets B1 and B2 each score 3/20 = 0.150 [0.052, 0.360] (ctl-1, ctl-23, ctl-31; the corrected B cells count here), still within the 0.20 bar (`reviews/pilot-round-0-gate-stop-memo.md`, P1 sensitivity). The labelled sets are wider than the single gold A on 12 of 20 items and contain A on all 20.
- Failure analysis (`pilot-round-0-gate-stop-memo.md:11-23`, drawn from the table): 5 threshold misses (raw A, probability under 0.70), 2 contested misses (a confident answer inside the labelled set: ctl-2, ctl-46), 1 near-duplicate miss (ctl-20), 3 path-pull misses (ctl-24, 33, 39); baseline leak on ctl-23 (the baselines fell back to `tests/`, which A owns; it did not decide the gate).
- Sensitivity from the memo: lowering both thresholds to 0.5 would give exactly 12 of 20 on P1 (experiment-peer calculation) but would also send ctl-39 to B and cancel P2's `none` answers on ctl-11, 22, 44 (p_true .68, .54, .59; `jev-p/scores.p2.jsonl:9,15,23`).

**P1 target directories** (20 items): `.github`, `.github/actions/notify-translations`, `.github/actions/people`, `.github/workflows`, `docs_src/app_testing/app_b_an_py310`, `docs_src/bigger_applications/app_an_py39/internal`, `docs_src/settings/app01`, `docs_src/templates`, `fastapi/security`, `tests` (×2), `tests/test_modules_same_name_body` (×2), `tests/test_tutorial` (×2), `tests/test_tutorial/test_advanced_middleware`, `tests/test_tutorial/test_configure_swagger_ui`, `tests/test_tutorial/test_header_params`, `tests/test_tutorial/test_query_params`, `tests/test_tutorial/test_request_forms_and_files` (`pilot-round-0.md:35`; matches `control-key.jsonl` `target_dir`).

### 4.4 P2 (rev 12): disclosures required by R10-02
Join (`control-key.jsonl` `source_case_id` → `units.jsonl` `set` → `gold.jsonl` → `arms-p2/V-sem.jsonl`; command in §10, "P2 join"). In rev 12 `gold.jsonl` holds the **derived** gold the gate used; the clone's own **direct label** is the final row of `labels.jsonl` for its case id (rows carrying `agreement`).

**Donors by set** (build stdout, `/tmp/x2d-gates-build.log:1`; `pilot-round-0.md:25,33`): `{P_unplaceable: 4, P_leftover: 2, D: 6}` = 12 clones; eligible P 26, eligible D 19, `p1_unplaceable` 4. The two P/P path-twin donors are `2c4db772` and `af2f04c8` (both exactly `scripts/playwright/request_form_models/image01.py`).

**Split by source set** (computed by the join; gap positive = derived gold `none`/gap):

| source set | clones | gap positives | `none` on gap positives | `none_correct_gap` | retained covered | `no_agreed_gold` |
|---|---|---|---|---|---|---|
| P (4 unplaceable + 2 leftover) | 6 | 4 | 2 | 2 | 2 (0 assigned correct) | 0 |
| D (6) | 6 | 3 | 2 | 2 | 3 (1 assigned correct) | 0 |
| Total | 12 | 7 | 4 | 4 | 5 (1 assigned correct) | 0 |

V-sem also answered `none` on 1 retained-covered P clone (ctl-7, gold `214b68`), so it answered `none` on 5 of the 12 clones.

**Per clone** (V-sem confidence is `1 − p_true` for a `none`):

| clone | set | source cluster | ablated area | derived gold | direct label | V-sem |
|---|---|---|---|---|---|---|
| ctl-5 | P | cc8803ec | ea1d02 | covered 3de70a | covered 3de70a (agreed) | assigned → b58662 (conf 0.76) |
| ctl-7 | P | 84f3f93b | 73232e | covered 214b68 | covered 214b68 (agreed) | none (conf 0.33) |
| ctl-10 | D | 5eedae70 | f36cd1 | covered bfd60b | covered bfd60b (agreed) | assigned → bfd60b (conf 0.94) |
| ctl-11 | P | d8a31a4c | c4f3d5 | gap | gap (agreed) | none (conf 0.32) |
| ctl-16 | D | dcc1b446 | 95b28e | covered 3de70a | covered 3de70a, e287e9 (agreed) | unresolved (conf 0.66) |
| ctl-18 | P | 2c4db772 | 26ac01 | gap | covered f0abaf, 158cb3 (adjudicated) | unresolved (conf 0.58) |
| ctl-22 | D | a71e28ee | c207aa | gap | gap (agreed) | none (conf 0.46) |
| ctl-30 | D | df1a8025 | d56ca4 | covered 3de70a | covered 233f9b, 3de70a (agreed) | unresolved (conf 0.61) |
| ctl-35 | D | 92757f9e | 9a4a0e | gap | covered c3534a (agreed) | unresolved (conf 0.68) |
| ctl-40 | P | af2f04c8 | 26ac01 | gap | covered f0abaf, 158cb3 (adjudicated) | assigned → 158cb3 (conf 0.70) |
| ctl-44 | D | 1aa87882 | ea7526 | gap | new (agreed) | none (conf 0.41) |
| ctl-45 | P | d636de1b | 73232e | gap | new (adjudicated) | none (conf 0.90) |

**Ablated areas: 10**, the same list on every P2 row (`control-key.jsonl` `.ablated_area_ids`): `26ac01` Playwright documentation example: request forms, `73232e` Repository metadata and contribution configuration, `95b28e` Schema extra data and examples, `9a4a0e` Dependency injection, `c207aa` Test suite and CI, `c4f3d5` Package build and publishing, `d56ca4` Query parameter models tutorials, `ea1d02` Header parameters tutorials, `ea7526` Public package API and compatibility, `f36cd1` HTTP routing and route registration. **`catalogue-p2` has 104 of the catalogue's 114 areas**; `catalogue-p1` 114; `catalogue-n1` 114 (`rounds/0/catalogue-p*.json`, `.areas` length).
**Clones sharing an ablated area (2 pairs):** `26ac01` (ctl-18 and ctl-40, the `image01.py` P/P path twin) and `73232e` (ctl-7 and ctl-45). **Path twins among donors:** one pair (ctl-18/ctl-40), found by intersecting donor member paths with the other D/P units' members; no D/P donor twin exists.

**What is construction and what is Jev** (`pilot-round-0.md:182-195`):
- **Derived gold contradicts direct labels on all 3 misses** (ctl-18, ctl-40, ctl-35): the clone's own blind label says an area still fits, and it is the area Jev picked or leaned to (`158cb3`, `158cb3`, `c3534a`). This is a label-versus-label fact; Jev is not involved. Direct labels also show ctl-44 and ctl-45 were `new`, not `gap`.
- The path twin counts one item twice (round 11 kept the per-clone count, `round-11.md:54`). 5 of 12 clones keep a second acceptable area (`gold.jsonl:15,25,36,86,89`): a thin denominator by design.
- **Inherited D label errors: not observed.**
- **Real signals:** on the retained-covered clones Jev made one confident wrong assignment (ctl-5 → `b58662`, gold `3de70a`, confidence 0.76). On 3 of the 4 true gaps Jev's `covered` reading was high (p_true 0.54–0.68), so its `none` answers depended on the untuned threshold.
- **Sensitivity (descriptive only, no re-scoring):** collapsing the twins gives 6 positives and 2 misses, still FAIL. Scoring against the direct labels (ctl-11, 22, 44, 45) gives 4 positives, all answered `none` (4/4 = 1.000 [0.510, 1.000]), which is below the floor of 8 and would have been UNMEASURED, not PASS.

### 4.5 Round-0 labels
`<data>/rounds/0/labels-check.json` (`ok` true; `/tmp/x2d-gates-label.log`): **κ_answer 1.0 on 50 units** (both labellers parsed on all 50 D ∪ P units; `n_units_all` 50, `kappa_answer_all_units` 1.0), κ on `covered` 1.0, raw agreement (answer and covered) 1.0, prevalence of `covered` 1.0, **adjudication rate 0.0** on those units, guide `x2d-association-v1/guide-3e036ff81b1b`, labels sha256 `36139293…1f401`. Round-0 label rows: 96 main cases, 93 agreed / 3 adjudicated (ctl-18, ctl-40, ctl-45), plus 20 contrast cases. The rule-citation concentration (`rule_top_share_over_0_40`) is true: rule R1 is cited in 91 of 100 parsed sessions (0.91); it is report-only by design (R3-02, `PREREGISTRATION.md:136`).

## 5. Power on the live frozen-round structure (`run.sh power`; model-free)
Source: `<data>/power.json` (schema `x2d-power-v5`, seed 20261029, 1,000 pilots per cell, arm correlation 0.5, PR weight 0.3 where stated); `/tmp/x2d-power.log`; `pilot-round-0.md:101`. Structure: rounds 8–10, **98 clusters, 78 covered, 44 PRs, 20 positives inside the population** (`real_none_proxy` 39). The simulation **assumes** the perception check, N1 and informativeness pass (`assumed_passing_in_simulation`); they were never measured.

- **`n4_max` = 0.054** (max over N4a, N4b, N4b-cap, N4b-ρ0; bar ≤ 0.10). **`all_pass` true**; every binding check true: P3_u0, P3_u05, P3_u10, N3, N3_PR, N4a, N4b, N4b_cap, N4b_rho0.
- **`p3_real_structure`:** P3-u0 GO share **0.894 [0.873, 0.912]** (bar ≥ 0.80), pass. **`p3_abstention`:** 0.876 at u = 0.05, 0.847 at u = 0.10.
- Columns: Jev accuracy `p_J`, baseline accuracy `p_B2`, random-abstention input `u` (`power.json` `.u`; the N4 cells abstain selectively (`mode` `b2_errors` for N4a; `own_errors` for N4b, N4b_rho0 and N4c; `own_errors_cap` for N4b_cap), so their `u` = 0.0 is not their abstention rate), error correlation ρ, PR weight, positives; GO share with Wilson 95 %; verdict counts GO / INCONCLUSIVE / NO-GO.

| cell | p_J | p_B2 | u input | ρ | PR w | pos | GO | GO share [Wilson] | bar | check | GO/INC/NO-GO |
|---|---|---|---|---|---|---|---|---|---|---|---|
| P3_u0 | 0.95 | 0.8 | 0.0 | 0.5 | 0.0 | 20 | 894/1000 | 0.894 [0.873, 0.912] | ≥ 0.8 | PASS | 894/0/106 |
| P3_u05 | 0.95 | 0.8 | 0.05 | 0.5 | 0.0 | 20 | 876/1000 | 0.876 [0.854, 0.895] | ≥ 0.75 | PASS | 876/1/123 |
| P3_u10 | 0.95 | 0.8 | 0.1 | 0.5 | 0.0 | 20 | 847/1000 | 0.847 [0.823, 0.868] | ≥ 0.6 | PASS | 847/1/152 |
| N3 | 0.5 | 0.5 | 0.0 | 0.0 | 0.0 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N3 | 0.5 | 0.5 | 0.1 | 0.0 | 0.0 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N3 | 0.6 | 0.6 | 0.0 | 0.0 | 0.0 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N3 | 0.6 | 0.6 | 0.1 | 0.0 | 0.0 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N3 | 0.7 | 0.7 | 0.0 | 0.0 | 0.0 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N3 | 0.7 | 0.7 | 0.1 | 0.0 | 0.0 | 20 | 1/1000 | 0.001 [0.000, 0.006] | ≤ 0.1 | PASS | 1/0/999 |
| N3 | 0.8 | 0.8 | 0.0 | 0.0 | 0.0 | 20 | 20/1000 | 0.020 [0.013, 0.031] | ≤ 0.1 | PASS | 20/1/979 |
| N3 | 0.8 | 0.8 | 0.1 | 0.0 | 0.0 | 20 | 12/1000 | 0.012 [0.007, 0.021] | ≤ 0.1 | PASS | 12/4/984 |
| N3 | 0.85 | 0.85 | 0.0 | 0.0 | 0.0 | 20 | 15/1000 | 0.015 [0.009, 0.025] | ≤ 0.1 | PASS | 15/1/984 |
| N3 | 0.85 | 0.85 | 0.1 | 0.0 | 0.0 | 20 | 24/1000 | 0.024 [0.016, 0.035] | ≤ 0.1 | PASS | 24/6/970 |
| N3 | 0.9 | 0.9 | 0.0 | 0.0 | 0.0 | 20 | 11/1000 | 0.011 [0.006, 0.020] | ≤ 0.1 | PASS | 11/0/989 |
| N3 | 0.9 | 0.9 | 0.1 | 0.0 | 0.0 | 20 | 12/1000 | 0.012 [0.007, 0.021] | ≤ 0.1 | PASS | 12/2/986 |
| N3 | 0.95 | 0.95 | 0.0 | 0.0 | 0.0 | 20 | 4/1000 | 0.004 [0.002, 0.010] | ≤ 0.1 | PASS | 4/0/996 |
| N3 | 0.95 | 0.95 | 0.1 | 0.0 | 0.0 | 20 | 3/1000 | 0.003 [0.001, 0.009] | ≤ 0.1 | PASS | 3/0/997 |
| N3_PR | 0.5 | 0.5 | 0.0 | 0.0 | 0.3 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N3_PR | 0.5 | 0.5 | 0.1 | 0.0 | 0.3 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N3_PR | 0.6 | 0.6 | 0.0 | 0.0 | 0.3 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N3_PR | 0.6 | 0.6 | 0.1 | 0.0 | 0.3 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N3_PR | 0.7 | 0.7 | 0.0 | 0.0 | 0.3 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N3_PR | 0.7 | 0.7 | 0.1 | 0.0 | 0.3 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/1/999 |
| N3_PR | 0.8 | 0.8 | 0.0 | 0.0 | 0.3 | 20 | 21/1000 | 0.021 [0.014, 0.032] | ≤ 0.1 | PASS | 21/7/972 |
| N3_PR | 0.8 | 0.8 | 0.1 | 0.0 | 0.3 | 20 | 24/1000 | 0.024 [0.016, 0.035] | ≤ 0.1 | PASS | 24/13/963 |
| N3_PR | 0.85 | 0.85 | 0.0 | 0.0 | 0.3 | 20 | 37/1000 | 0.037 [0.027, 0.051] | ≤ 0.1 | PASS | 37/11/952 |
| N3_PR | 0.85 | 0.85 | 0.1 | 0.0 | 0.3 | 20 | 40/1000 | 0.040 [0.030, 0.054] | ≤ 0.1 | PASS | 40/12/948 |
| N3_PR | 0.9 | 0.9 | 0.0 | 0.0 | 0.3 | 20 | 21/1000 | 0.021 [0.014, 0.032] | ≤ 0.1 | PASS | 21/4/975 |
| N3_PR | 0.9 | 0.9 | 0.1 | 0.0 | 0.3 | 20 | 21/1000 | 0.021 [0.014, 0.032] | ≤ 0.1 | PASS | 21/5/974 |
| N3_PR | 0.95 | 0.95 | 0.0 | 0.0 | 0.3 | 20 | 5/1000 | 0.005 [0.002, 0.012] | ≤ 0.1 | PASS | 5/0/995 |
| N3_PR | 0.95 | 0.95 | 0.1 | 0.0 | 0.3 | 20 | 2/1000 | 0.002 [0.001, 0.007] | ≤ 0.1 | PASS | 2/1/997 |
| N4a | 0.8 | 0.8 | 0.0 | 0.5 | 0.0 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N4a | 0.85 | 0.85 | 0.0 | 0.5 | 0.0 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N4a | 0.9 | 0.9 | 0.0 | 0.5 | 0.0 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N4a | 0.95 | 0.95 | 0.0 | 0.5 | 0.0 | 20 | 0/1000 | 0.000 [0.000, 0.004] | ≤ 0.1 | PASS | 0/0/1000 |
| N4b | 0.85 | 0.85 | 0.0 | 0.5 | 0.0 | 20 | 42/1000 | 0.042 [0.031, 0.056] | ≤ 0.1 | PASS | 42/246/712 |
| N4b | 0.88 | 0.88 | 0.0 | 0.5 | 0.0 | 20 | 34/1000 | 0.034 [0.024, 0.047] | ≤ 0.1 | PASS | 34/256/710 |
| N4b | 0.9 | 0.9 | 0.0 | 0.5 | 0.0 | 20 | 24/1000 | 0.024 [0.016, 0.035] | ≤ 0.1 | PASS | 24/171/805 |
| N4b | 0.95 | 0.95 | 0.0 | 0.5 | 0.0 | 20 | 3/1000 | 0.003 [0.001, 0.009] | ≤ 0.1 | PASS | 3/4/993 |
| N4b_cap | 0.8 | 0.8 | 0.0 | 0.5 | 0.0 | 20 | 51/1000 | 0.051 [0.039, 0.066] | ≤ 0.1 | PASS | 51/384/565 |
| N4b_cap | 0.85 | 0.85 | 0.0 | 0.5 | 0.0 | 20 | 37/1000 | 0.037 [0.027, 0.051] | ≤ 0.1 | PASS | 37/397/566 |
| N4b_rho0 | 0.85 | 0.85 | 0.0 | 0.0 | 0.0 | 20 | 54/1000 | 0.054 [0.042, 0.070] | ≤ 0.1 | PASS | 54/395/551 |
| N4b_rho0 | 0.88 | 0.88 | 0.0 | 0.0 | 0.0 | 20 | 43/1000 | 0.043 [0.032, 0.057] | ≤ 0.1 | PASS | 43/543/414 |
| N4b_rho0 | 0.9 | 0.9 | 0.0 | 0.0 | 0.0 | 20 | 48/1000 | 0.048 [0.036, 0.063] | ≤ 0.1 | PASS | 48/420/532 |
| N4c (report-only) | 0.9 | 0.8 | 0.0 | 0.5 | 0.0 | 20 | 694/1000 | 0.694 [0.665, 0.722] |  |  | 694/192/114 |
| P3_PR (report-only) | 0.95 | 0.8 | 0.05 | 0.5 | 0.3 | 20 | 786/1000 | 0.786 [0.760, 0.810] |  |  | 786/14/200 |
| P3_rho0 (report-only) | 0.95 | 0.8 | 0.05 | 0.0 | 0.0 | 20 | 830/1000 | 0.830 [0.805, 0.852] |  |  | 830/4/166 |
| P3 curve Δ0.15 (report-only) | 0.85 | 0.7 | 0.05 | 0.5 | 0.0 | 20 | 346/1000 | 0.346 [0.317, 0.376] |  |  | 346/0/654 |
| P3 curve Δ0.15 (report-only) | 0.9 | 0.75 | 0.05 | 0.5 | 0.0 | 20 | 718/1000 | 0.718 [0.689, 0.745] |  |  | 718/0/282 |
| P3 curve Δ0.15 (report-only) | 0.95 | 0.8 | 0.05 | 0.5 | 0.0 | 20 | 876/1000 | 0.876 [0.854, 0.895] |  |  | 876/1/123 |
| P3 curve Δ0.15 (report-only) | 1.0 | 0.85 | 0.05 | 0.5 | 0.0 | 20 | 916/1000 | 0.916 [0.897, 0.932] |  |  | 916/0/84 |

Realised mean abstention (`u_mean`) of the selective cells, in table order: N4a 0.142, 0.131, 0.097, 0.051; N4b 0.150, 0.120, 0.101, 0.050; N4b_cap 0.142, 0.132; N4b_rho0 0.149, 0.120, 0.099; N4c 0.102 (`power.json` `.N4a[].u_mean`, `.N4b[].u_mean`, `.N4b_cap[].u_mean`, `.N4b_rho0[].u_mean`, `.N4c.u_mean`).

[INFERENCE] These cells bound what a GO verdict would have been able to show, not what Jev does: the power to reach GO was 0.89 at p_J 0.95 vs p_B2 0.80 and u 0, and the null cells (Jev = B2) stayed at or below 0.054.

## 6. Re-gate (rev 13, R-053): the one permitted re-test of P1 and P2 on fresh items
**Population statement (R17-03; `PREREGISTRATION.md:111`): these figures test the repaired instrument on legacy 2019–2022 files. They are not a re-test of the rev-12 population and are not evidence about window-era files.** All 40 P1 items and 36 of 40 P2 clones are `in_no_unit` legacy clusters (§0.3).
Sources: `<data>/rounds/0/regate/gates.json` (gate set `x2d-regate-v1`, SHA-256 `081be905f088fd0f67e15995166ea28279bea340c7bb7d451c4014f6c0532e7c`) at the JSON paths named in each table; `/tmp/x2d-regate-*.log`; `pilot-round-0.md:212-233`. Figures marked *computed* come from `p1_items` / `p2_items` by the command in §10 ("Re-gate tables") and reproduce the file's own totals exactly (checked: 14/20, 15/35, 6/35, 1/20, 15/38, 14/38, 9/11, 1/12).

### 6.1 Build, labelling and the pre-Jev check
- **Fresh pool:** 129 clusters from 129 PRs in 69 areas (D 3, G 7, `in_no_unit` 119); the P pool was exhausted (26 eligible, all used by rev 12). Exposure excluded: 83 presented round-0 paths, 17 rev-12 P1 target directories (`regate-supply.md:19-33`).
- **P1:** 40 of 40 items, achieved strata **37 / 0 / 3** (target pattern 24 / 8 / 8; the stratum-2 targets fell back, `regate-supply.md:43`). **P2:** 40 of 40 clones in **16 ablated areas** (D 2, G 2, `in_no_unit` 36). `problems: none`. Plan equals `reviews/regate-supply.json` byte for byte (SHA-256 `e499c52b…4eaa51`); `regate/regate.json` SHA-256 `2b42b918…d183ea` (`utc` 2026-10-30T22:54:55Z); `regate/rounds/0/control-key.jsonl` `5946b690…769b81e`.
- **Labelling:** 120 cases, 120 gold rows; 115 agreed, 5 adjudicated; 246 of the cap of 330 sessions (`regate/rounds/0/sessions.json`; `labels.jsonl`).
- **Pre-Jev check** (read-only, repeated by `regate-jev` before any request; `pilot-round-0.md:222`): items built 40, eligible 38, stratum-1 eligible 35, direct-label positives 12, `rules_resolved_units` 0, baselines B1 and B2 0/38 each ([0.000, 0.092]), `twins_excluded` 0. No `PROBLEM:` line, so Jev ran.
- **Excluded P1 items (2):** `ctl-38` (`tests/test_tutorial/test_path_operation_configurations`) and `ctl-64` (`docs_src/cookie_params`), both stratum 1, because the agreed gold excludes A (`gates.json` `.gates.P1.value.excluded`). Their Jev answers are not scored. `gold_not_covered` and `no_agreed_gold` are empty.

### 6.2 Decisive figures
Source: `.gates.P1`, `.gates.P2`, `.bars`.

| gate | conjunct | value | bar | result |
|---|---|---|---|---|
| P1 | stratum-1 accuracy on answered items | 14/20 = 0.700 [0.481, 0.855] | ≥ 0.80 | **not met** |
| P1 | stratum-1 abstention rate | 15/35 = 0.429 [0.280, 0.591] | ≤ 0.15 | **not met** |
| P1 | B1 and B2 recovery over all eligible items (each) | 0/38 = 0.000 [0.000, 0.092] (B1 and B2 both) | ≤ 0.20 | met |
| P1 | `rules_resolved_units` | 0 | = 0 | met |
| P1 | eligible stratum-1 items | 35 | ≥ 20 (else UNMEASURED) | met |
| P1 | **status** |  |  | **FAIL** |
| P2 | direct-label positives | 12 (11 gap, 1 new) | ≥ 8 (else UNMEASURED) | met |
| P2 | `none` / answered positives | 9/11 = 0.818 [0.523, 0.949] | ≥ 0.80 | met |
| P2 | abstention (unresolved / positives) | 1/12 = 0.083 [0.015, 0.354] | ≤ 0.15 | met |
| P2 | **status** |  |  | **PASS** |

### 6.3 Reported P1 figures (`.gates.P1.value`)
Stratum 1 (35 eligible items) beside all eligible items (38; 3 more from stratum 3):

| figure | stratum 1 (35) | all eligible (38) |
|---|---|---|
| answered-item accuracy (decisive) | 14/20 = 0.700 [0.481, 0.855] | 15/21 = 0.714 [0.500, 0.862] |
| abstention rate (decisive) | 15/35 = 0.429 [0.280, 0.591] | 17/38 = 0.447 [0.301, 0.603] |
| `recovery_all_items` (acceptable set, abstention = miss; the rev-12 figure) | 14/35 = 0.400 [0.256, 0.564] | 15/38 = 0.395 [0.256, 0.553] |
| `recovery_all_items_A_only` (named exactly A) | 13/35 = 0.371 [0.232, 0.537] | 14/38 = 0.368 [0.234, 0.527] |
| `sibling_in_acceptable` (items whose sibling area is also acceptable) | 0 | 0 |
| `none` answers (a resolved wrong answer on a covered item) | 5 | 5 |
| path-following, composed answers over answered items | 1/20 = 0.050 [0.009, 0.236] | 1/21 = 0.048 [0.008, 0.227] |
| path-following, unthresholded permutation-1 choice over all items | 6/35 = 0.171 [0.081, 0.327] | 6/38 = 0.158 [0.074, 0.304] |
| B1 recovers an acceptable area | 0/35 = 0.000 [0.000, 0.099] | 0/38 = 0.000 [0.000, 0.092] |
| B2 recovers an acceptable area | 0/35 = 0.000 [0.000, 0.099] | 0/38 = 0.000 [0.000, 0.092] |

Not decisive: strata 2 and 3, which are report-only. The decisive accuracy rests on 20 answered items from 19 areas (§6.6).

**By stratum** (`.gates.P1.value.strata`; stratum 2 has **no** items, so the 37/0/3 build left stratum 2 untested):

| stratum | items | answered | correct / answered | abstained | `none` answers | `recovery_all_items` | `recovery_all_items_A_only` | `sibling_in_acceptable` |
|---|---|---|---|---|---|---|---|---|
| 1 | 35 | 20 | 14/20 | 15 | 5 | 14/35 = 0.400 [0.256, 0.564] | 13/35 = 0.371 [0.232, 0.537] | 0 |
| 2 | 0 | 0 | — | 0 | 0 | — | — | 0 |
| 3 | 3 | 1 | 1/1 | 2 | 0 | 1/3 = 0.333 [0.061, 0.792] | 1/3 = 0.333 [0.061, 0.792] | 0 |

### 6.4 Path-following (report-only, no bar)
Path-following = the answer names a sibling area of the target directory (`sibling_area_ids`) that is not acceptable. Composed: 1/20 of answered stratum-1 items (1/20 = 0.050 [0.009, 0.236]) and 1/21 over all answered; the one case is `ctl-21` (`tests/test_tutorial/test_conditional_openapi`, answered `c2b9d0`, acceptable `10c16a`, `617fde`). Raw first choice: 6/35 = 0.171 [0.081, 0.327] of stratum-1 items (6/38 = 0.158 [0.074, 0.304] all eligible). **By `target_dir` prefix** (*computed*; stratum 1 only; the 37 stratum-1 targets are 24 `tests/test_tutorial/test_<topic>`, 9 `docs_src/<topic>` and 4 other before the 2 exclusions, 23 / 8 / 4 after):

| group | items | answered | correct / answered | abstained | `none` answers | path-following (answered) | path-following (raw) |
|---|---|---|---|---|---|---|---|
| `tests/test_tutorial/<topic>` | 23 | 13 | 11/13 = 0.846 [0.578, 0.957] | 10 | 1 | 1/13 | 4/23 |
| `docs_src/<topic>` | 8 | 3 | 1/3 = 0.333 [0.061, 0.792] | 5 | 2 | 0/3 | 2/8 |
| other | 4 | 4 | 2/4 = 0.500 [0.150, 0.850] | 0 | 2 | 0/4 | 0/4 |
| all stratum 1 | 35 | 20 | 14/20 = 0.700 [0.481, 0.855] | 15 | 5 | 1/20 | 6/35 |

"Other" = `.github/ISSUE_TEMPLATE`, `scripts/playwright/separate_openapi_schemas`, `tests/test_filter_pydantic_sub_model`, `tests/test_validate_response_recursive`. [INFERENCE] The cells are small (3–23 items): `tests/test_tutorial` answers are mostly right (11/13) but 10 of 23 abstain, and `docs_src` items mostly abstain (5/8). Composed path-following is rare (1/20), but the raw first choice named the non-acceptable sibling area on 6/35 stratum-1 items: 4 abstained (ctl-3, ctl-33, ctl-60, ctl-65), 1 was answered `none` (ctl-75) and 1 was assigned (ctl-21). That is close to rev 12's raw = B ≠ A rate of 3/20 (§4.3). The threshold turned most of this pull into abstention, so the data do not rule path pull out, and they give no power to rank prefixes.

### 6.5 By year and by origin (*computed* from `p1_items` / `p2_items`)
**Origin.** Every P1 item is `in_no_unit` (40/40; 38 eligible; 35 stratum-1), so P1 has no origin split. P2 is split below.

**P1 by year, eligible items** (answered = assigned or `none`; correct = assigned to an acceptable area; path-following composed over answered / raw over all):

| group | items | answered | correct / answered | abstained | `none` answers | path-following (answered) | path-following (raw) |
|---|---|---|---|---|---|---|---|
| 2019 | 21 | 12 | 10/12 = 0.833 [0.552, 0.953] | 9 | 2 | 0/12 | 3/21 |
| 2020 | 9 | 5 | 2/5 = 0.400 [0.118, 0.769] | 4 | 3 | 0/5 | 1/9 |
| 2021 | 5 | 3 | 2/3 = 0.667 [0.208, 0.939] | 2 | 0 | 1/3 | 2/5 |
| 2022 | 3 | 1 | 1/1 = 1.000 [0.206, 1.000] | 2 | 0 | 0/1 | 0/3 |
| all | 38 | 21 | 15/21 = 0.714 [0.500, 0.862] | 17 | 5 | 1/21 | 6/38 |

**P1 by year, stratum 1 (decisive stratum):**

| group | items | answered | correct / answered | abstained | `none` answers | path-following (answered) | path-following (raw) |
|---|---|---|---|---|---|---|---|
| 2019 | 19 | 11 | 9/11 = 0.818 [0.523, 0.949] | 8 | 2 | 0/11 | 3/19 |
| 2020 | 9 | 5 | 2/5 = 0.400 [0.118, 0.769] | 4 | 3 | 0/5 | 1/9 |
| 2021 | 4 | 3 | 2/3 = 0.667 [0.208, 0.939] | 1 | 0 | 1/3 | 2/4 |
| 2022 | 3 | 1 | 1/1 = 1.000 [0.206, 1.000] | 2 | 0 | 0/1 | 0/3 |
| all | 35 | 20 | 14/20 = 0.700 [0.481, 0.855] | 15 | 5 | 1/20 | 6/35 |

[INFERENCE] 2019 carries 19 of the 35 stratum-1 items and 9 of the 14 correct answers; 2020 is 2 correct of 5 answered with 3 `none`. With 3–19 items per year this describes the sample's age mix, not a trend.

**P2 positives by year** (direct-label positives, n = 12):

| year | positives | answered | `none` | assigned | abstained |
|---|---|---|---|---|---|
| 2019 | 2 | 2 | 1 | 1 | 0 |
| 2020 | 4 | 4 | 4 | 0 | 0 |
| 2021 | 1 | 1 | 1 | 0 | 0 |
| 2022 | 3 | 2 | 1 | 1 | 1 |
| 2023 | 1 | 1 | 1 | 0 | 0 |
| 2024 | 1 | 1 | 1 | 0 | 0 |
| all | 12 | 11 | 9 | 2 | 1 |

**P2 retained-covered clones by year** (28; correct answer is the retained area):

| year | clones | answered | `none` | assigned | abstained |
|---|---|---|---|---|---|
| 2019 | 13 | 10 | 0 | 10 | 3 |
| 2020 | 7 | 6 | 1 | 5 | 1 |
| 2021 | 2 | 2 | 1 | 1 | 0 |
| 2022 | 4 | 1 | 0 | 1 | 3 |
| 2023 | 1 | 1 | 1 | 0 | 0 |
| 2024 | 1 | 1 | 0 | 1 | 0 |
| all | 28 | 21 | 3 | 18 | 7 |

**P2 by origin** (donor origin from `control-key.jsonl`; gate positives, then retained):

| origin | positives | answered | `none` | assigned | abstained |
|---|---|---|---|---|---|
| in_no_unit | 10 | 9 | 7 | 2 | 1 |
| D | 2 | 2 | 2 | 0 | 0 |
| G | 0 | 0 | 0 | 0 | 0 |

| origin | retained clones | answered | `none` | assigned | abstained |
|---|---|---|---|---|---|
| in_no_unit | 26 | 19 | 2 | 17 | 7 |
| D | 0 | 0 | 0 | 0 | 0 |
| G | 2 | 2 | 1 | 1 | 0 |

Only 2 of the 12 positives (both D) come from outside the 2019–2022 `in_no_unit` pool; the two G clones are retained-covered.

### 6.6 Area concentration: per-area accuracy and the one-item-per-area sensitivity (round-18 §3.6)
Area `e287e9` is the original area A for **5** stratum-1 items and `9f08e6` for **3**; the 35 items come from **26** areas (*computed*, stratum 1 only, grouped by A). `e287e9`, `9f08e6`, `3de70a` and `617fde` are also re-gate P2 ablated areas, but only on the separate `catalogue-p2`; the P1 items were asked on `catalogue-p1` (`reviews/round-18.md:81`):

| area | title | stratum-1 items | answered | correct | abstained | `none` |
|---|---|---|---|---|---|---|
| e287e9 | OpenAPI schema and interactive documentation | 5 | 2 | 2 | 3 | 0 |
| 9f08e6 | Custom responses | 3 | 1 | 1 | 2 | 0 |
| 3de70a | Parameter declarations and request input extraction | 2 | 1 | 0 | 1 | 1 |
| 4d46c3 | Advanced path operation configuration tutorials | 2 | 1 | 1 | 1 | 0 |
| 617fde | Response model validation and encoding | 2 | 1 | 0 | 1 | 0 |
| 11da39 | HTTP authentication schemes | 1 | 1 | 1 | 0 | 0 |
| 1c212a | Path operation configuration tutorials | 1 | 0 | 0 | 1 | 0 |
| 26c1c4 | Security dependencies and authentication examples | 1 | 1 | 1 | 0 | 0 |
| 2874a6 | Bigger applications | 1 | 1 | 0 | 0 | 1 |
| 3d3e91 | Extra models tutorials | 1 | 0 | 0 | 1 | 0 |
| 3e0c34 | Advanced middleware | 1 | 1 | 1 | 0 | 0 |
| 3efbd5 | Lint and format commands | 1 | 1 | 0 | 0 | 1 |
| 4b7221 | Approved-label automation | 1 | 1 | 0 | 0 | 1 |
| 6790cf | Body updates | 1 | 0 | 0 | 1 | 0 |
| 70d6ec | Handling errors tutorials | 1 | 1 | 1 | 0 | 0 |
| 722160 | Additional response models | 1 | 1 | 1 | 0 | 0 |
| 7823f3 | Async tests | 1 | 1 | 1 | 0 | 0 |
| 8d1a9d | Additional status codes | 1 | 1 | 1 | 0 | 0 |
| 95b28e | Schema extra data and examples | 1 | 0 | 0 | 1 | 0 |
| 9a4a0e | Dependency injection | 1 | 0 | 0 | 1 | 0 |
| 9ce9c5 | Application settings examples | 1 | 1 | 0 | 0 | 1 |
| a1080e | Data structures | 1 | 1 | 1 | 0 | 0 |
| bd67ed | Applications behind a proxy | 1 | 0 | 0 | 1 | 0 |
| c2f26e | Application events | 1 | 1 | 1 | 0 | 0 |
| c3534a | Dependency injection | 1 | 0 | 0 | 1 | 0 |
| d1f466 | Change response status code | 1 | 1 | 1 | 0 | 0 |

| `e287e9` item | target dir | V-sem | acceptable set |
|---|---|---|---|
| ctl-28 | `tests/test_tutorial/test_response_headers` | unresolved (`below_t_choice`, 0.67; raw `ade7d5`) | `ade7d5`, `e287e9` |
| ctl-30 | `tests/test_tutorial/test_bigger_applications` | assigned `e287e9` (0.87) | `e287e9` |
| ctl-33 | `tests/test_tutorial/test_request_files` | unresolved (`below_t_choice`, 0.39) | `e287e9` |
| ctl-4 | `tests/test_tutorial/test_sub_applications` | assigned `95b28e` (0.85), acceptable though not A | `3de70a`, `95b28e`, `e287e9` |
| ctl-73 | `tests/test_tutorial/test_openapi_callbacks` | unresolved (`permutation_disagreement`) | `722160`, `e287e9` |

`9f08e6`: ctl-14 unresolved, ctl-34 assigned `9f08e6` (0.99), ctl-60 unresolved.

**One-item-per-area sensitivity.** §3.6 names the check but not the selection rule, so I report the whole range instead of choosing. Result: every area's *answered* items are either all correct or all wrong (13 areas all correct, 6 areas all wrong, 0 mixed), so **any one-answered-item-per-area selection gives 13/19 = 0.684 [0.460, 0.846]** (19 areas have an answered item), against the item-level 14/20 = 0.700 [0.481, 0.855]. Weighting areas equally gives 0.684. Dropping the two largest areas, `e287e9` (5 items) and `9f08e6` (3 items) (five areas have more than one item), (3 correct answers) leaves 11/17 = 0.647 [0.413, 0.827]. Abstention is not selection-invariant: of 26 areas, 7 have every item abstained and 12 have at least one, so one item per area gives an abstention rate between 7/26 = 0.269 [0.137, 0.461] and 12/26 = 0.462 [0.288, 0.645], and without `e287e9` and `9f08e6` 10/27 = 0.370 [0.215, 0.558] at item level. [INFERENCE] Area concentration lowers the one-per-area accuracy by 1.6 points (0.700 → 0.684) and does not rescue either bar: accuracy stays below 0.80 and abstention stays above 0.15 under every selection.

### 6.7 Rev-12 source directories as targets, and stratum 3 (round-18 §3.6)
**Three stratum-1 targets are rev-12 *source* directories** (the directory a rev-12 P1 file was moved *from*), so their sibling cue points at a rev-12 item's area. The basenames are new and C0 paths are refused, so this is not a path exposure (`round-18.md:82`). *Computed* by intersecting `regate/gates.json` `target_dir` with the dirnames of `rounds/0/control-key.jsonl` `original_paths`:

| re-gate item | target dir | rev-12 item(s) that came from it | its sibling area | re-gate original area A | V-sem |
|---|---|---|---|---|---|
| ctl-6 | `tests/test_tutorial/test_separate_openapi_schemas` | ctl-1 | `ad8836` Separate OpenAPI input and output schemas | `3de70a` Parameter declarations and request input extraction | unresolved |
| ctl-18 | `scripts/playwright/separate_openapi_schemas` | ctl-6, ctl-25 | `ef2b9b` Playwright OpenAPI schema examples | `3efbd5` Lint and format commands | `none` (wrong; raw `3efbd5`) |
| ctl-80 | `tests/test_validate_response_recursive` | ctl-24 (and its dup control ctl-9, same source unit `373fe757`) | `617fde` Response model validation and encoding | `722160` Additional response models | assigned `722160` (correct) |

[INFERENCE] 1 of the 3 is correct, 1 abstained, 1 `none`; none followed the sibling area, so there is no sign that these targets biased the figures, but n = 3.

**Stratum 3 rests on one directory.** All 3 stratum-3 items (`ctl-13`, `ctl-32`, `ctl-36`) target `tests/test_tutorial/test_response_change_status_code`, with three different original areas (`11da39`, `f36cd1`, `533bc8`), so the stratum-3 figure (accuracy 1/1 answered, abstained 2/3 = 2/3 = 0.667 [0.208, 0.939], `recovery_all_items` 1/3 = 0.333 [0.061, 0.792]) is **not a stratum-level result**.

### 6.8 Abstention by reason (`reason` in `p1_items` / `p2_items`)
| group | items | abstained | `below_t_choice` | `permutation_disagreement` | `below_t_cov` | `failure` |
|---|---|---|---|---|---|---|
| P1, all eligible | 38 | 17 | 15 | 2 | 0 | 0 |
| P1, stratum 1 | 35 | 15 | 13 | 2 | 0 | 0 |
| P2 positives | 12 | 1 | 1 | 0 | 0 | 0 |
| P2 retained-covered | 28 | 7 | 7 | 0 | 0 | 0 |

(*computed*. `none` answers are not abstentions: a `none` is composed when `1 − p_true ≥ 0.30`, i.e. the covered read falls under the provisional 0.70.) The two P1 `permutation_disagreement` items are `ctl-65` (`docs_src/debugging`) and `ctl-73` (`tests/test_tutorial/test_openapi_callbacks`), both stratum 1; the 2 stratum-3 abstentions are `below_t_choice`.

### 6.9 P2 in detail (`.gates.P2.value`, `.p2_items`)
- **Gate:** 40 clones; **12 direct-label positives** (`positives_by_none_kind`: gap 11, new 1); `none` on answered positives 9/11 = 0.818 [0.523, 0.949]; abstention 1/12 = 0.083 [0.015, 0.354]; **`none_rate_all_positives` 9/12 = 0.750 [0.468, 0.911]**; 2 positives assigned an area. `no_agreed_gold` 0; `twins_excluded` `[]` (no path twin remained).
- **Donors by origin:** `in_no_unit` 36, D 2, G 2, P 0 (the P pool was exhausted). **Ablated areas: 16**; `catalogue-p2` has **98 of 114 areas** (`regate/rounds/0/catalogue-p2.json`). Clones per ablated area 4, 4, 4, 3, 3, 3, then nine areas with 2 and one with 1; the 16 ablated areas, with clones each: `3de70a` Parameter declarations and request input extraction 4, `e287e9` OpenAPI schema and interactive documentation 4, `9a4a0e` Dependency injection 4, `9f08e6` Custom responses 3, `f36cd1` HTTP routing and route registration 3, `617fde` Response model validation and encoding 3, `063f15` Client generation tutorials 2, `30399e` Documentation site build and tooling 2, `73232e` Repository metadata and contribution configuration 2, `86949d` Python types tutorials 2, `a0f998` Request file upload tutorials 2, `70d6ec` Handling errors tutorials 2, `c6a286` WebSocket tutorials 2, `c79f1b` Testing tutorial examples 2, `dc7550` Response classes and status behavior 2, `2266c6` Query parameter string validations tutorials 1 (`p2_items` `ablated_area_id`; titles from `<data>/catalogue/catalogue.json`); **15 areas are shared by 39 clones**, so the clones are strongly clustered by area.
- **Positives by ablated area (`positives_by_ablated_area`, with outcomes, *computed*)** — 8 areas hold the 12 positives:

| ablated area | title | clones in area | positives | `none` | assigned | abstained | none kind |
|---|---|---|---|---|---|---|---|
| 9f08e6 | Custom responses | 3 | 3 | 3 | 0 | 0 | gap |
| 86949d | Python types tutorials | 2 | 2 | 2 | 0 | 0 | gap |
| a0f998 | Request file upload tutorials | 2 | 2 | 0 | 2 | 0 | gap |
| 063f15 | Client generation tutorials | 2 | 1 | 1 | 0 | 0 | gap |
| 30399e | Documentation site build and tooling | 2 | 1 | 1 | 0 | 0 | gap |
| 3de70a | Parameter declarations and request input extraction | 4 | 1 | 0 | 0 | 1 | gap |
| 73232e | Repository metadata and contribution configuration | 2 | 1 | 1 | 0 | 0 | new |
| f36cd1 | HTTP routing and route registration | 3 | 1 | 1 | 0 | 0 | gap |

  Two areas hold 5 of the 9 `none` answers (`9f08e6` 3, `86949d` 2); both wrong assignments are in `a0f998`. [INFERENCE] At area level 6 of 7 areas with an answer were all-`none` (0.857, [0.487, 0.974] *computed*), which gives no more support for the 0.80 bar than the item-level 9/11.
- **The 12 positives:**

| clone | date | origin | ablated area | none kind | V-sem | conf (`1 − p_true` for `none`) | abstain reason |
|---|---|---|---|---|---|---|---|
| ctl-47 | 2023-12-26 | D | 063f15 | gap | none | 0.68 |  |
| ctl-10 | 2024-08-05 | D | 30399e | gap | none | 0.33 |  |
| ctl-23 | 2022-08-18 | in_no_unit | 3de70a | gap | unresolved | 0.3 | below_t_choice |
| ctl-17 | 2020-05-17 | in_no_unit | 73232e | new | none | 0.82 |  |
| ctl-11 | 2020-06-12 | in_no_unit | 86949d | gap | none | 0.42 |  |
| ctl-41 | 2022-05-14 | in_no_unit | 86949d | gap | none | 0.33 |  |
| ctl-26 | 2021-05-10 | in_no_unit | 9f08e6 | gap | none | 0.31 |  |
| ctl-51 | 2020-03-01 | in_no_unit | 9f08e6 | gap | none | 0.46 |  |
| ctl-57 | 2020-06-13 | in_no_unit | 9f08e6 | gap | none | 0.36 |  |
| ctl-42 | 2022-01-23 | in_no_unit | a0f998 | gap | assigned → 749bd1 | 0.86 |  |
| ctl-68 | 2019-04-14 | in_no_unit | a0f998 | gap | assigned → 749bd1 | 0.98 |  |
| ctl-70 | 2019-08-24 | in_no_unit | f36cd1 | gap | none | 0.31 |  |

- **Competing explanations carried over (round-17 §4, `reviews/round-17.md:172-174`; round-18 §3.6).**
  - For the P2 PASS: run-to-run noise near the provisional threshold (four `none` answers within 0.03 of the cut; [INFERENCE] Jev's repeatability was not measured, so a re-run could fail P2); cues in the PR text, which Jev sees with each legacy item (`PREREGISTRATION.md:111`); old, central files; labellers seeing more context than Jev (§0.7).
  - For the P1 FAIL: the untuned threshold (§6.8), cross-scope near-duplicate areas (§0.9), legacy content (§0.3), topic cues in test-directory paths (§6.4) and run-to-run noise.
  - The P1 abstention conjunct is not near its bar. If the three stratum-1 abstentions with choice confidence ≥ 0.60 had been answered (ctl-28 0.67, ctl-35 0.63, ctl-60 0.65), abstention would still be 12/35 = 0.343.
- **Retained-covered clones** (own label keeps an area; 28; reported, not decisive): assigned correctly 18/28 = 0.643 [0.458, 0.793], assigned wrong 0/28 = 0.000 [0.000, 0.121], `none` 3/28 = 0.107 [0.037, 0.272], abstained 7/28 = 0.250 [0.127, 0.434]; answered 21/28 = 0.750 [0.566, 0.873].
- **`derived_vs_direct`** (key = direct label | rev-12 derived gold; the derived gold is reported, never scored): `covered|covered` 13, `covered|none` 15, `none|none` 12. So the rev-12 derivation called 27 of the 40 clones gap/new and the clones' own labels agreed on only 12 of them (12/27 = 0.444 [0.276, 0.627]); it never called an area covered that the label called `none` (0 of 13).

## 7. Costs
Sources: summed `cost_usd` over `type: answer` rows of each score file; `pilot-round-0.md:47,70-83,223`. Model `typesafe/jev-1.14-20261014`. Cap: $2 per Jev round (round 0 included) and $20 total (rev 12, R-052).

| stage | file | requests | Jev cost | note |
|---|---|---|---|---|
| Onboarding (consensus: contested/unowned files) | `onboard-jev/scores.jsonl` | 1,534 answers | $0.185375 |  |
| Round 0 `main` (64 units: D 20, P 30, controls: 9 duplicate, 3 planted-new, 1 prose, 1 trap) | `rounds/0/jev/scores.jsonl` | 7,872 answers + 7 `failure` rows | $1.009834 (known rows) | first run to the stop: 7,755 answers, $1.000115 billed; resumed run: 117 answers, $0.009719; the 7 `failure` rows have `cost_known: false` |
| Round 0 `n1` (50 units, shuffled catalogue) | `rounds/0/jev-n1/scores.jsonl` | 110 | $0.048770 |  |
| Round 0 `p1` (20 units) | `rounds/0/jev-p/scores.p1.jsonl` | 43 | $0.025070 |  |
| Round 0 `p2` (12 units) | `rounds/0/jev-p/scores.p2.jsonl` | 26 | $0.014860 |  |
| **Round 0 subtotal** |  | 8,051 answers | **$1.098534** | of the $2 per-round cap (rev 12) |
| Re-gate `p1` (40 units) | `regate/rounds/0/jev-p/scores.p1.jsonl` | 87 | $0.049920 | charged to the round-0 cap and the total |
| Re-gate `p2` (40 units) | `regate/rounds/0/jev-p/scores.p2.jsonl` | 87 | $0.043979 |  |
| **Round 0 incl. re-gate** |  | 8,225 answers | **$1.192433** | of $2 |
| **All Jev spend (onboarding + round 0 + re-gate)** |  | 9,759 answers | **$1.377808** | of the $20 total cap; rounds 1–10 spent nothing |

- **Unknown, not zero:** 7 `main`-run requests were cancelled at the cost stop and are stored as `type: failure`, `kind: network`, `cost_known: false` (their `cost_usd` field reads 0 as a placeholder, not a measurement; `rounds/0/jev/scores.jsonl`, request ids e.g. `0:ctl-43|merge.which_area|V-title|2|-`). They were re-asked in the resumed run and answered (7,872 of 7,872), but the cost of the cancelled attempts is unknown, so every total above is a lower bound of what was billed by up to those 7 requests. The stop log's "0 with unknown billed cost" counts answers only.
- **Measured price per request (round-0 `main`):** fan-out `merge.area_covers` 7,296 requests, $0.670496 (≈ $0.000092 each); `merge.which_area` 384, $0.211143 (≈ $0.000550); `merge.covered` 192, $0.128194 (≈ $0.000668). The preregistered budget assumed $0.000055 per request (R-052).
- **Billed vs cap at the stop:** $1.0001 billed against `-max-cost-usd 1.0000` (requests in flight overshot, ruled in round 17).
- **CLI sessions** (`claude -p` and `codex exec`, no API keys): 16 Clawpatch agent maps (codex `gpt-6-luna`, `xhigh`); labelling sessions **317 of the cap of 429 in round 0** (80 intruder audit of cap 110, then 196 + 41 for the 116 gate cases; `rounds/0/sessions.json`: plan W 116, I 40) and **246 of the cap of 330 in the re-gate** (W 120). Labeller rows: re-gate A 120 / B 120 / adjudicator 5 (115 agreed, 5 adjudicated); round-0 main A 96 / B 96 / adjudicator 3.
- No model call was made to write this file.

## 8. Deviations
Everything below is a departure from, or a recorded exception to, the frozen plan. Each was reviewed by the council or ruled by the user as stated.

### 8.1 The four run commits and the stages run at each
`PREREGISTRATION.md:249-251`; `pilot-round-0.md:5-17,50-52,212-216`; digests from `reviews/round-9.md`, `round-11.md`, `round-16.md`, `round-18.md`. The current `PREREGISTRATION.md` has SHA-256 `b9110486…140a8`, equal to the round-18 sign-off.

| commit | sign-off (council round; PREREGISTRATION SHA-256) | stages run | note |
|---|---|---|---|
| `730674e54bab2bc4e76f59cd5a476031732204df` | rev 9 (round 9, 0 issues; `4b14dae0…38429`) | `onboard-collect`, `onboard` (consensus + catalogue freeze), `quality` + intruder audit; the first `gates` attempt, which refused | the refused `gates` attempt's own commit is not recorded separately; reused unchanged after the rev-10/11 change, because every file those stages read has the same digest under both sign-offs and `rounds/0/version.json` had not been written |
| `0b5659440c7368ac18b82f60a4edf3956408d2bf` | rev 11 (round 11, 0 issues; `b5029a6f…fdec22`; `signoff.json` SHA-256 `e7059f2a…dc2d`) | `gates-build`, `gates-label` (+ labels-check, gold), `gates-jev` up to the cost stop | validation (a) 651 checks / 0 failed; (b) 650 passed / 1 failed |
| `b6a5bf30324201bc88caae75d5774255c6cac54d` | rev 12 (round 16, 0 issues; `de553c52…8120`) | the round-0 re-freeze, `gates-jev` (resumed), `gates-analyse`, `power` | `0b56594` could not run them: its `x2d.CAPS` is $1 / $10 and round 0 had spent $1.0001; validation (a) 651 / 0, (b) 651 / 1 |
| `8c7148cbb39b58440b39d18e0dddd0668e3e249c` | rev 13 (round 18, 0 issues; `b9110486…140a8`; `signoff.json` `1ef1402d…1171b`) | `regate-build`, `regate-label`, `regate-precheck`, `regate-jev`, `regate-analyse` | validation (a) 715 checks / 0 failed (record aside, SHA-256 equal before and after); (b) 715 / 1 failed |

The run worktree moved between commits by `git checkout --detach` only (no `git clean`, no `stash -u`); `data/` stayed untracked and untouched.

### 8.2 P2 build refusal and the rev-10/11 top-up (user decision 2026-10-30, option f1)
The first `gates` attempt (`/tmp/x2d-gates.log`) **refused**: 26 eligible P units left 6 for P2 after the 20 P1 items, 8 needed. Four of the 30 P units had no single owner (named in §0.1). Even with every in-scope unit eligible the ceiling was 28 of 28, zero slack: a design defect in allocation, not a Jev result (`reviews/pilot-round-0-p2-gate-memo.md:11-22`). No label, score or outcome existed. Rev 10 made the P2 source pool P leftovers then D (`n_p2 = min(12, eligible_P − 20 + eligible_D)`; refusal below 8 unchanged). Council round 10 gave 3 issues; **R10-01** found area `26ac01437eaf` owns exactly one file (`scripts/playwright/request_form_models/image01.py`) and two P clusters consist of just that file, so rev 10 would have built two P1 items whose gold area was absent from `catalogue-p1`; rev 11 sends such units to the P2 pool as unplaceable. Round 11: 0 issues. The recomputed allocation (P1 20, P2 12 = P unplaceable 4, P leftover 2, D 6; 10 ablated areas; `catalogue-p2` 104) matched the live `gates-build` output on every item (`pilot-round-0.md:19-42`).

### 8.3 The $1 cost stop and the rev-12 caps (R-052)
`gates-jev` stopped (exit 4) at the preregistered $1 per-round cap after **7,755 of 7,872** `main` requests ($1.0001 billed). Labels and gold already existed; thresholds were the provisional 0.70; no Jev answer had been compared with a label or gold. The measured cost per request was ≈ $0.000092 for the fan-out (≈ 1.7× the $0.000055 assumption) and ≈ $0.000589 for the other questions (≈ 10.7×); the projection for the whole pilot was ≈ $9.74 (R-052). The user decided, verbatim: **"$2 per round, $20 total"**. Rev 12 changed only the `caps` block of `thresholds.json` (`usd_per_round` 2.0, `usd_total` 20.0) plus text and two fixtures; council rounds 12–16 gave 3, 2, 2, 1 and 0 issues. The 7 `cost_known: false` requests are reported as **unknown, not zero** (§7). Attestation (R12-02 (d)): no Jev answer was compared with a label or gold before the user's decision, nor since, until `gates-analyse`: `scores*.csv` held only their headers (`pilot-round-0.md:48`).

### 8.4 The round-0 re-freeze (R12-01)
Once, after the rev-12 sign-off and before the resumed `gates-jev`, because the rev-12 cap change altered the `thresholds.json` digest pinned by round 0's version record. `rounds/0/version.json` (`6355be3f…3ca8`) was kept as `version.rev11.json`; `versioning.py --round 0` re-froze it (`e079e2d1…b64f`). The key diff was exactly `thresholds_sha256` (`5c211eb9…94a8` → `7142d7ac…95eb`) and `question_version_keys` (`null` → the 7 `versions` of the score-file header); `settings.json` was byte-unchanged (`c165a695…f6e2`); `--drift` printed `round 0: no drift from version.json`. All four digests above were re-read from the data files for this report and are equal (`pilot-round-0.md:54-62`).

### 8.5 Validation (b): the expected single FAIL
With the council's `signoff.json` in place, `run.sh validate` fails exactly one check by design: **"C without the council's sign-off record the collector refuses before any clone or model call"**. The sign-off run is the no-record run. Counts: rev 11 650 passed / 1 failed (651 with the record aside); rev 12 651 / 1; rev 13 715 / 1 (`/tmp/x2d-validate-r11-norecord.log`, `-r11-signoff.log`, `-r15a.log`, `-r16b.log`, `-r18a.log`, `-r18b.log`). Rule 4 (planted effect recovered, planted null not): the validated implementation includes a planted-effect and planted-null pair for the GO rule (`/tmp/x2d-validate-r18a.log:150-153`), N2 label permutation (`:206`), the verdict (`:755-763`), P1 (`:404-405`), P2 (`:412-413`) and the whole re-gate end to end (`:493, :497`); the log ends `715 checks, 0 failed`.

### 8.6 The gate stop and the rev-13 re-gate (R-053)
`gates-analyse` exited 1 with P1 and P2 FAIL. Under `PREREGISTRATION.md` §5 the pilot stopped before round 1. The council memo (`reviews/pilot-round-0-gate-stop-memo.md`) showed both gates partly mis-built (P1: single-owner gold narrower than the labelled sets, an untuned threshold with abstention counted as a miss, a baseline leak; P2: derived gold contradicting the clones' own labels, a path twin counted twice, a thin denominator) and a real stratum-1 weakness on P1. **User decision, option (c):** repair both gates and re-test them once on fresh held-out items with the Jev side (model, v1 questions, guide, 0.70 thresholds) frozen; the round-0 FAIL stays on record and is never re-scored (`rounds/0/gates.json` unchanged, SHA-256 `d7c1b98b…bd48e`). Rev 13: council round 17 gave 7 issues (R17-01…R17-07), round 18 gave 0. The re-gate lived in `rounds/0/regate/`, a self-contained data root; one shot; a re-gate that is not PASS for both stops the pilot. It was not PASS (P1 FAIL), so the pilot stopped and no round-0 critique was written (`pilot-round-0.md:233`; `round-18.md:187`).

### 8.7 Other deviations and record notes
1. **No round-0 critique, no rounds 1–10, no `tune`, `rescore-v7` or `verdict`.** By rule: a stop before round 1 (`PREREGISTRATION.md:116`).
2. **P1 re-gate achieved strata 37 / 0 / 3** against the target 24 / 8 / 8: the build falls back when no host of the target kind is free; the minimum of 20 stratum-1 items was met (35 eligible). Stratum 2 is therefore untested in the re-gate, and stratum 3 is one directory (§6.7).
3. **Two P1 items excluded** because the agreed gold excludes A (`ctl-38`, `ctl-64`); the build refuses to shrink, so 40 were built (§6.1).
4. **`reviews/regate-supply.json` was produced with `signoff.json` moved aside** (byte-identical restore, SHA-256 `578a2462…2e61` before and after) and `X2D_SKIP_SIGNOFF=1`, as a model-free membership run (`regate-supply.md:7-15`). It wrote nothing under `data/`.
5. **`quality.json` is stale on the intruder audit** (status "items written; labels pending") because it was written before labelling; the result is in `quality.intruder.json`. **The split-half consensus ARI is unmeasured** (§3).
6. **`PREREGISTRATION.md:3` still reads "DRAFT for council round 18 … the rev-13 re-gate … has **not been built**".** The file is frozen after sign-off (its SHA-256 equals the signed one), so the status line was not updated. It is stale; the record is `reviews/pilot-round-0.md:212-233` and this file.
7. **Logs are not committed.** `/tmp/x2d-*.log` are the stage logs; they may be cleaned. The data files, `reviews/pilot-round-0.md` and the `signoff.json` digests are the durable record.
8. **Council record:** rounds 1–9 (rev 1–9): 29, 18, 18, 12, 12, 7, 3, 2, 0 issues; round 10: 3; round 11: 0; rounds 12–16 (rev 12): 3, 2, 2, 1, 0; rounds 17–18 (rev 13): 7, 0 (`reviews/round-N.md`).

## 9. Ledger outcomes
Sources: `docs/hypothesis-ledger.md:32-34,38`; `docs/research-ledger.md` R-049…R-054; `PREREGISTRATION.md:253`.

| ID | State | Reason / content |
|---|---|---|
| **HD6** | **`inconclusive`** (was `preregistered`; never `under-test`) | "round-0 gates P1 and P2 failed and the one permitted re-gate did not pass (re-gate P1 FAIL, P2 PASS); HD6 never under test". Not `not-supported` (a gate stop is not a verdict) and not `refuted` (one repository, a pilot). Evidence: round-0 P1 9/20 = 0.45, P2 4/7; re-gate P1 stratum 1 (legacy 2019–2022 items, not window-era files) accuracy 14/20 = 0.70 [0.48, 0.85] (bar 0.80), abstention 15/35 = 0.43 [0.28, 0.59] (cap 0.15), baselines 0/38 (all eligible; 0/35 in stratum 1); P2 9/11 = 0.82 [0.52, 0.95], abstention 1/12. Record `reviews/pilot-round-0.md`; `regate/gates.json` SHA-256 `081be905…2e7c`. |
| **HD6b** | `proposed` (unchanged) | New-concern detection is not tested by X2d. Only the planted tests exist: planted-new 3/3 `none` (round 0) and re-gate P2 `none` 9/12 over all positives [0.468, 0.911]; real `P_new` / `R_new` are unmeasured (no round ran). |
| **HD7** | `proposed` (new) | Cross-scope near-duplicate areas (R-054). The onboarding duplicate check compared the 261 title-token-sharing same-scope pairs (of 1,268) and flagged 0; cross-scope pairs were never compared. One identical title spans scopes ("Dependency injection": `core` `9a4a0e`, `tutorial-1` `c3534a`) and 6 cross-scope pairs have title-token Jaccard ≥ 0.5. In round-0 P1 the duplicate split Jev's probability on 2 items and a "Request forms" / "Request form model tutorials" pair left 1 unresolved; in the re-gate both stratum-1 items whose A is a "Dependency injection" area abstained. No effect on accuracy has been measured; title overlap is a proxy. Not testable inside X2d (frozen catalogue); a design change needs council sign-off. |

R-049 (user direction: semantic association, 10-round pilot, no dismissal on one or two directional failures, no large trial before proof), R-051 (GO rule revision), R-052 (caps), R-053 (re-gate), R-054 (cross-scope duplicates) are the sources used. X2d proposes no design change (`PREREGISTRATION.md:16`), and none follows from a stopped pilot.

## 10. Reproduce
**Everything here is read from the stored data; the model-dependent stages are append-only and non-deterministic.** Re-running them costs ≈ $1.38 of Jev, ≈ 560 labelling sessions and 16 Clawpatch maps, refuses when its outputs exist (use a fresh `X2D_DATA`), and will not reproduce identical model outputs. The figures in this file are reproduced from the stored files by the model-free commands below.

**Toolchain and validation** (model-free; `CLAUDE.md`):
```
cd /Users/adamjackson/Projects/inflight/scout
git submodule update --init && scripts/bootstrap-tools.sh && source .tools/activate
# signoff.json moved aside -> "715 checks, 0 failed"; with the record in place -> 715 / 1 failed
# (the expected failure: "C without the council's sign-off record the collector refuses before any clone or model call")
evaluation/x2d-merge-association-pilot/run.sh validate
```

**The stages as run** (run worktree `../scout-x2d`, moved between commits by `git checkout --detach` only; the Jev key is injected by Infisical into the one command that needs it and is never exported or printed; `data/` is untracked):
```
W=/Users/adamjackson/Projects/inflight/scout-x2d
INF='infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent --'
R=evaluation/x2d-merge-association-pilot/run.sh
export X2D_SRC=/tmp/scout-x2c-smoke/_src/fastapi        # any clone holding the pinned revisions
cd $W && source .tools/activate

git checkout --detach 730674e54bab2bc4e76f59cd5a476031732204df          # rev-9 sign-off
$R onboard-collect $X2D_SRC             # 16 maps (codex gpt-6-luna xhigh)       /tmp/x2d-onboard-collect.log
$INF $R onboard                         # consensus, Jev, freeze, catalogue sha  /tmp/x2d-onboard.log
$R quality                              # section 4.3 + intruder audit (labellers: claude -p, codex exec)  /tmp/x2d-quality.log
$INF $R gates                           # refused: P2 shortfall                  /tmp/x2d-gates.log

git checkout --detach 0b5659440c7368ac18b82f60a4edf3956408d2bf          # rev-11 sign-off
$R gates-build                          # /tmp/x2d-gates-build.log
$R gates-label                          # /tmp/x2d-gates-label.log
$INF $R gates-jev                       # exit 4 at the $1 cap                   /tmp/x2d-gates-jev.log

git checkout --detach b6a5bf30324201bc88caae75d5774255c6cac54d          # rev-12 sign-off
D=$W/evaluation/x2d-merge-association-pilot/data
mv $D/rounds/0/version.json $D/rounds/0/version.rev11.json               # re-freeze (R12-01)
.tools/venv/bin/python evaluation/x2d-merge-association-pilot/versioning.py --round 0
.tools/venv/bin/python evaluation/x2d-merge-association-pilot/versioning.py --round 0 --drift
$INF $R gates-jev                       # /tmp/x2d-gates-jev-resume.log
$R gates-analyse                        # exit 1: P1, P2 FAIL                    /tmp/x2d-gates-analyse.log
$R power                                # /tmp/x2d-power.log

git checkout --detach 8c7148cbb39b58440b39d18e0dddd0668e3e249c          # rev-13 sign-off
$R regate-build                         # alone first                            /tmp/x2d-regate-build.log
$R regate-label                         # /tmp/x2d-regate-label.log
X2D_DATA=$D/rounds/0/regate .tools/venv/bin/python evaluation/x2d-merge-association-pilot/analysis_x2d.py --regate-precheck
$INF $R regate-jev                      # repeats the precheck first             /tmp/x2d-regate-jev.log
$R regate-analyse                       # exit 1: P1 FAIL, P2 PASS               /tmp/x2d-regate-analyse.log
```
Validation logs: `/tmp/x2d-validate-r11-norecord.log`, `-r11-signoff.log`, `-r15a.log`, `-r16b.log`, `-r18a.log`, `-r18b.log`.

**Integrity checks** (model-free): `shasum -a 256` of `<data>/catalogue/catalogue.json` = `1220016c…4199`; `<data>/rounds/0/gates.json` = `d7c1b98b…bd48e`; `<data>/rounds/0/regate/gates.json` = `081be905…2e7c`; `regate/supply.json` = `reviews/regate-supply.json` = `e499c52b…4eaa51`.

**P2 join** (§4.4; the R10-02 command; rev-12 round 0; runs under any Python 3; `D` is the data directory set above, `/Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data`):
```
X2D_DATA=$D python3 - <<'E'
import collections, json, os
D = os.environ["X2D_DATA"] + "/rounds/0/"          # the study data directory
rd = lambda f: [json.loads(l) for l in open(D + f) if l.strip()]
ck = [r for r in rd("control-key.jsonl") if r["kind"] == "p2"]
U = {u["case_id"]: u for u in rd("units.jsonl")}
G = {g["case_id"]: g for g in rd("gold.jsonl")}      # rev 12: the DERIVED gold the gate used
V = {a["case_id"]: a for a in rd("arms-p2/V-sem.jsonl")}
direct = {l["case_id"]: l for l in rd("labels.jsonl") if "agreement" in l}   # the clone's own final label
t = collections.defaultdict(collections.Counter)
for r in ck:
    s, g, v = U[r["source_case_id"]]["set"], G.get(r["case_id"]), V[r["case_id"]]
    k = t[s]; k["clones"] += 1
    if g is None: k["no_agreed_gold"] += 1; continue
    if not g["covered"]:
        k["gap_positives"] += 1; k["none"] += v["output"] == "none"
        k["none_correct_gap"] += v["output"] == "none" and g["none_kind"] == "gap"
    else:
        k["retained_covered"] += 1
        k["retained_assigned_correct"] += v["output"] == "assigned" and v["area_id"] in g["acceptable"]
for s, k in sorted(t.items()): print("set", s, dict(k))
abl = {a for r in ck for a in r["ablated_area_ids"]}
cat = json.load(open(D + "catalogue-p2.json"))["areas"]
print("ablated areas", len(abl), sorted(abl), "catalogue-p2 areas", len(cat))
print("clones sharing an ablated area", {a: [r["case_id"] for r in ck if r["ablated_area_id"] == a]
      for a in {r["ablated_area_id"] for r in ck} if sum(r["ablated_area_id"] == a for r in ck) > 1})
paths = collections.defaultdict(set)
for u in rd("units.jsonl"):
    if not u["control"]:
        for m in u["members"]: paths[m].add(u["case_id"])
print("donor path twins", [(r["case_id"], m, sorted(paths[m] - {r["source_case_id"]}))
      for r in ck for m in U[r["source_case_id"]]["members"] if paths[m] - {r["source_case_id"]}])
print("direct labels", {r["case_id"]: (direct[r["case_id"]]["covered"], direct[r["case_id"]]["none_kind"]) for r in ck})
E
```

**Rev-12 P1 per item** (§4.3; rev-12 round 0; any Python 3; prints the per-item rows and the counts quoted in §4.3):
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
Expected output: hit 9/20, abstained 7/20, raw_A 14/20, raw_in_labelled 17/20, raw_B 4/20, raw_B_not_A 3/20, V_labelled 11/20, B1_labelled and B2_labelled 3/20, none 2, stratum 1 2/11, strata 2–3 7/9 (run and confirmed for this revision).

**Re-gate tables** (§6.4–§6.9; by year, origin, `target_dir` prefix, per area, rev-12 source directories, abstention reasons, P2 by area):
```
X2D_DATA=$D python3 - <<'E'
import collections, json, os, posixpath
D = os.environ["X2D_DATA"] + "/rounds/0/"
g = json.load(open(D + "regate/gates.json"))
P1 = [i for i in g["p1_items"] if not i["excluded"]]; P2 = g["p2_items"]; S1 = [i for i in P1 if i["stratum"] == 1]
def p1(items):
    v = lambda i: i["V-sem"]; ans = [i for i in items if v(i)["output"] in ("assigned", "none")]
    return dict(n=len(items), answered=len(ans),
        correct=sum(v(i)["output"] == "assigned" and v(i)["area_id"] in i["acceptable"] for i in ans),
        abstained=sum(v(i)["output"] == "unresolved" for i in items), none=sum(v(i)["output"] == "none" for i in items),
        pf_answered=sum(v(i)["output"] == "assigned" and v(i)["area_id"] in i["sibling_area_ids"] and v(i)["area_id"] not in i["acceptable"] for i in ans),
        pf_raw=sum(v(i)["raw_choice"] in i["sibling_area_ids"] and v(i)["raw_choice"] not in i["acceptable"] for i in items),
        reasons=dict(collections.Counter(v(i)["reason"] for i in items if v(i)["output"] == "unresolved")))
def p2(items):
    return dict(n=len(items), none=sum(i["output"] == "none" for i in items), assigned=sum(i["output"] == "assigned" for i in items),
                abstained=sum(i["output"] == "unresolved" for i in items))
print("stratum 1", p1(S1)); print("all eligible", p1(P1))
for y in sorted({i["date"][:4] for i in P1}):
    print("year", y, "all", p1([i for i in P1 if i["date"][:4] == y]), "s1", p1([i for i in S1 if i["date"][:4] == y]))
pref = lambda t: "tests/test_tutorial" if t.startswith("tests/test_tutorial/") else "docs_src" if t.startswith("docs_src/") else "other"
for p in ("tests/test_tutorial", "docs_src", "other"): print("prefix", p, p1([i for i in S1 if pref(i["target_dir"]) == p]))
by = collections.defaultdict(list)
for i in S1: by[i["area_id"]].append(i)
for a, its in sorted(by.items(), key=lambda kv: (-len(kv[1]), kv[0])): print("area", a[:6], p1(its))
ck = [json.loads(l) for l in open(D + "control-key.jsonl")]            # rev-12 round 0
src = {posixpath.dirname(p) for r in ck if r["kind"] == "p1" for p in r["original_paths"]}
print("rev-12 source dirs used as targets", [(i["case_id"], i["target_dir"], i["stratum"]) for i in g["p1_items"] if i["target_dir"] in src])
print("stratum-3 target dirs", sorted({i["target_dir"] for i in P1 if i["stratum"] == 3}))
POS = [i for i in P2 if i["direct_label"] == "none"]; RET = [i for i in P2 if i["direct_label"] == "covered"]
print("P2 positives", p2(POS), "retained", p2(RET))
for k in ("date", "origin"):
    for x in sorted({(i["date"][:4] if k == "date" else i[k]) for i in P2}):
        f = lambda i: (i["date"][:4] if k == "date" else i[k]) == x
        print("P2", k, x, "positives", p2([i for i in POS if f(i)]), "retained", p2([i for i in RET if f(i)]))
pa = collections.defaultdict(list)
for i in POS: pa[i["ablated_area_id"]].append(i)
for a, its in sorted(pa.items(), key=lambda kv: (-len(kv[1]), kv[0])): print("P2 area", a[:6], p2(its))
E
```
Both scripts were run against the stored data for this file and print the totals quoted here.

## 11. Unmeasured and not run
**Not run (the pilot stopped at round 0; `PREREGISTRATION.md:116`):** rounds 1–10 (`units`, `label`, `labels-check`, `gold`, `baselines`, `jev`, `tune`, `analysis`), the round-0 critique, `rescore-v7`, `verdict` (no `verdict.json`), the reference arm (`claude -p`), perception, the contrast comparison (the 20 round-0 contrast units were built and labelled, not analysed), the replayed free-running divergence, the refinement envelope, and any threshold tuning (thresholds stayed the provisional 0.70). **No GO / INCONCLUSIVE / NO-GO line was produced; the twelve GO conditions a–l were never evaluated.**

**Built or answered in round 0 but never analysed** (the `analysis` stage did not run): the control figures (9 duplicate, 1 empty-rename trap, 1 misleading-prose control in the `main` run; they were labelled and answered but not compared).

**Unmeasured within what ran:** split-half consensus ARI (§3); 1,217 inventory files excluded by policy (e) (§3); the re-gate's stratum 2 (0 items); the re-gate labeller κ and labels-check (not run by `regate-label`); the cost of the 7 cancelled requests (§7); per-reason and per-area intervals that treat items as independent (§0.6); real-world `P_new` / `R_new` (no window round); `unmeasured` count of the gates themselves: **0** for K2–K6, P1, P2, planted_new in round 0; **0** for the re-gate (both gates decided, neither UNMEASURED).

**Preregistered RESULTS obligations (`PREREGISTRATION.md:255`) and where each stands:**

| obligation | status |
|---|---|
| round-0 report (§4.3, §5), real-`none` proxy, live-structure power (`p3_real_structure`, `p3_abstention`, `n4_max`, the cells) | **reported** (§3, §4, §5) |
| per round the §9 table for every arm | **not applicable**: no round ran |
| κ (joint answer; κ on `covered`, raw agreement, prevalence; all scored units and the GO population) | round 0 only, **reported** (§4.1 K4, §4.5); per-round and GO-population κ not applicable |
| adjudication; labels-check | adjudication: round 0 and re-gate **reported** (§4.5; §6.1: 115 agreed / 5 adjudicated of 120); labels-check and κ: round 0 **reported** (§4.5), re-gate **unmeasured** (`regate-label` writes labels and gold only; no `rounds/0/regate/rounds/0/labels-check.json` exists); rounds 1–10 not applicable |
| strata; `modified_unowned` | P1 strata **reported** (§6.3); per-round strata and `modified_unowned` not applicable |
| session cap and used count | **reported** (§7) |
| primary arm | V-sem (K5 PASS, `fan_out_primary` false); the critique verdict is not applicable (no round-0 critique: it is written only on a re-gate PASS) |
| control figures (dup / trap / prose) | **not measured** (answered, not analysed) |
| replayed free-running divergence; contrast (window and D); reference arm; perception; leave-one-round-out | **not applicable / not run** |
| N1 / P1 results | **reported** (§4.1, §6) |
| P2 status; "R_new on catalogue gaps has no planted test" | **reported** (§4.4, §6.9); the sentence applies only to UNMEASURED and does not apply |
| `u` with per-reason split and interval; abstention table; `gain_res`, `gain_guess`, `gain_all` | **not applicable** for window rounds. Abstention by reason **is** reported for the gates (§6.8) |
| new-area table (`P_new`, `R_new`, real and planted, Wilson 95 %, "no veto", under HD6b) | **not applicable**: the preregistered planted figures pool frozen-round traps (`pooled.planted_new`, `PREREGISTRATION.md:158`) and the real figures need window rounds; none ran. Gate analogues are reported: planted_new 3/3 (§4.1); re-gate P2 `none` on all positives 9/12 = 0.750 [0.468, 0.911] (§6.9) |
| the GO / INCONCLUSIVE / NO-GO line naming each failed condition | **no verdict computed**; the §14 map gives HD6 `inconclusive` (§9) |
| HD6 ledger state per the §14 map | **reported** (§9) |
| §3 role mix of the folded round 10; episode sensitivity (`#14569` …) | role mix **not computed** (membership-only, `PREREGISTRATION.md:71`; no stage that ran writes it: `k23.json` `.per_round."10"` has no role field); episode sensitivity **not applicable** (no window round) |
| `Deviations`, reproduce commands | **reported** (§8, §10) |
