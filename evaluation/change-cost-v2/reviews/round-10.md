# E1b-v2 — Innovation council review, round 10

Scope: `evaluation/change-cost-v2/` after the round-9 fixes — delta check plus a fresh full review (prereg §1–10, `analysis.py`, `validate.py`, `constructs.py`, `tests.py`, `gates.py`, `annotate.py`/`jev_*.py` secret handling, stored JSON, test log, digests, all nine review rounds and responses). Read-only; `.codegraph` untouched (no script references it: `grep codegraph *.py` → no matches); no secrets printed; no predictor–outcome association computed (only stored distributions, marginals and text read). Main-agent evidence (2026-10-21): `/tmp/e1bv2-tests-r10.log`, `/tmp/e1bv2-constructs-r10.json`, `/tmp/e1bv2-digests-round10.txt`. An independent challenger (different model family) reviewed in parallel; synthesis below. Standing decisions (a)–(c) not reopened. No shell was available to council or challenger: digests were compared entry by entry, not recomputed; they remain main-agent evidence as in rounds 5–9.

## Verdict: ZERO ISSUES — sign-off granted.

## Round-9 resolution check

| # | Status | Evidence |
|---|---|---|
| 1 (nit, dead conditional) | **Resolved** | `constructs.py:58` now reads `Q_zero_share=float((g.Q == 0).mean()), Q_q=qs(g.Q), Q_within_sd=within_sd(g, "Q"),` — no `or True`, no unreachable branch; `grep "or True" *.py *.md` → no matches. Reproduction: `/tmp/e1bv2-constructs-r10.json` and `constructs.json` agree on every spot-checked value (`n_panel` 7717; `overall.ALL.Q_within_sd` 0.6463035722536351; `y_active_q` {−0.6931, −0.6931, 0.0, 0.9327, 2.1866}; floor share 0.38150782361308677; `repos.django.Q_within_sd` 0.5938573720328045; `repos.tokio.D_within_sd` 0.02185727841853647; the whole `variance` block incl. `y_active_area_icc` 0.006050002468321489); byte identity (`cmp`) is the main agent's evidence (`round-9-response.md`). `constructs.json` digest 68325887… unchanged in §10. |
| 2 (nit, `validate.py:89` comment) | **Resolved** | `validate.py:89`: "G4: 30 PRs per Jev class (dev classes from panel PRs; `nondev` from all typed PRs, since nondev PRs never enter the panel)" — matches `:92` `if k in inpanel or v == "nondev"` and prereg §2 (`nondev` excluded from the panel) / §5 G4 (30 per class incl. `nondev`). No sampling change. |
| 3 (nit, `tests.py:331` fixture comment) | **Resolved** | `tests.py:331`: "G7: hidden items: python 8/10 no (D000, D002 yes), then D004 unparsable twice -> 7 no / 2 yes / 1 unresolved; go 7/10 no; explained: 8/10 yes" — matches the fixture (`:316` even i = python/odd i = go, i < 20 hidden; `:335` labels; `:339` D004 `unparsable`) and the assertions `:360-362` (`by_language.python.n == 9`, unresolved 1/30). |

**Tests.** `/tmp/e1bv2-tests-r10.log`: 64 `ok` lines (counted; two interleaved `jackknife 0 / 16` progress lines are not checks), last line `all tests passed`. The log covers the planted logit recovery (0.5 ± 0.12), ridge-vs-λ=0 < 2 %, all `decide` branches, G3b refusal, G4/G5/G7 fixtures incl. partner aggregation and the per-language failure rule, jackknife df, and non-convergence error.

**Digests.** All 48 lines of `/tmp/e1bv2-digests-round10.txt` equal `DESIGN-NOTES.md` §10 (`:121-168`) entry by entry (council and challenger independently). Exactly `constructs.py`, `tests.py` (8aa4bdbb… → 39e98ad7…) and `validate.py` (81ce2777… → f8643b2d…) differ from the round-8/9 list; `analysis.py` 651cc1ba…, `power.py` 6ac3d720…, `harness.py` 233b91a1… are the digests recorded in `reviews/round-6.md:17`, `round-7.md:3`, `round-8.md:3`, so the stored `harness-results.json` (0ce59a21…) and `power-results.json` (2c924dcb…) stand without re-run. No `validation/`, `results.json`, `sensitivity.json` or `RESULTS.md` exists; `reviews/round-1.md … round-9.md` and all nine responses exist (rule 2).

## New issues

**None.**

Two items were considered and judged not to be issues (recorded for transparency): (i) `validate.py:214` `cons.get(i) if i in cons else None` is redundant with `cons.get(i)` but both branches are reachable and the intent (missing → unresolved) is unambiguous — unlike round-9's `or True`, it contradicts nothing; rounds 7–9 read this function and did not flag it. (ii) Script docstrings give `uv run --with pandas …` usage hints while prereg §10 fixes `$U = uv run --python 3.13 --with-requirements requirements.txt python`; the preregistration governs the run command and the hints are not contradictory.

## Checked and found correct (fresh review)

- Prereg §3.2 controls = `analysis.CONTROLS` (`analysis.py:30`); SESOI dict `:31` = §4; T1/T3 FE repo+quarter+task (`:140-144`), T2/T4 area absorbed with quarter+task dummies (`:151-154`; `absorb = A["area"]` `:258` regardless of inference cluster); T5 logit on `log_hidden`, `log_partner_abs`, `Q`, controls, FE repo/quarter/task (`:157-162`), ridge 1e-2 off the intercept, step cap 2, non-convergence raises (`:96-110`); delete-one-cluster jackknife with full refit, `(G−1)/G Σ(θ₍₋g₎−θ̄)²`, per-test cluster presence, t with G−1 df, hard error on non-finite decision replicates (`:261-285`); Holm (`:290-296`); `decide` (`:299-327`) = §4 verdict rules incl. `equivalent` only with `adequate`; `main` (`:331-351`) refuses missing G3b, stops on G4 failure, requires `--power/--validation/--gates`, errors if a decision test has no sample.
- P1 (`:177-222`): per-repo 60/40 temporal split, train-period standardisation, `P1_base_vs_const` reported.
- `validate.py`: blinded files drop `pipeline/kind/repo/number/parent/follow/stratum/lang/proxy/partner_key` (`:140`); ids assigned after shuffling (`:101,120,137-138`); G5 strata Jev-pos ≥ 0.5 / Jev-neg (`:107-108`), IPW weights population/sampled (`:181`); G7 12 per (language, proxy) cell, `other` skipped, hidden partner ≤ 3 rng-drawn vias, explained → `via_ref` (`:121-136`); `kappa`/`consensus`/`disagree` treat `unparsable` as §5 (`:153-174`); G4/G5/G7 thresholds and pass rules = §5 (`:184-185,201-202,232-236`); G7 partner aggregation and per-language rule (`:213-230`).
- `constructs.py` computes distributions/ICCs only (docstring `:2`, `icc` `:25-35` outcome-marginal).
- `gates.py:33-37`: G3b Spearman only under `--g3b`; stored `gates.json` has G1–G3 only.
- Secrets: `annotate.py:62-64`, `jev_label.py:78-80`, `jev_type.py:75-77` presence-check `OPENROUTER_API_KEY` and exit; no `print` of the key anywhere.
- Prereg §10 commands match every CLI (`analysis.py` argparse `:333-338`; `validate.py sample|gates` `:258`; `gates.py --g3b`; `constructs.py [panel]`).

## Challenger synthesis

Challenger verdict: ZERO ISSUES. It independently confirmed the three round-9 resolutions with the same line evidence, the 64-check log (`:1-58,60,62-66` ok, `:67` "all tests passed"), the 48/48 digest match, the unchanged `analysis.py`/`harness.py`/`power.py` hashes against `/tmp/e1bv2-digests-round8.txt`, construct-reproduction agreement, and additionally re-verified `collect.py`/`build_panel.py`/`follow.py`/`finalize_panel.py` measures, `harness.py:130-171` non-vacuous checks (stored permutation FWER 0.075, passing), `power.py` marginal-only nuisance and the empirical-MDE adequacy rule (T5 inadequate). No disagreement between council and challenger; the challenger did not change the recommendation.

## Sign-off statement

The innovation council signs off E1b-v2 (`evaluation/change-cost-v2/`) with zero open issues on design and implementation. The reviewed, now-frozen preregistration and scripts are (SHA-256, `DESIGN-NOTES.md` §10; `shasum -a 256`, main-agent recomputation 2026-10-21):

```
7d87913c535b286fe7e2c54419297fdfc29eebfae4a64269e33990963cb49bc1  PREREGISTRATION.md
651cc1ba3e95edff8f335ffcc0174b5e39d93e30d8ba71b6aa2e4659b0e428a9  analysis.py
d22f4d0feb54d71e8110477f1501bc61270d7b455de34844c3eca462edbe677f  annotate.py
f2e41c96c6868c81a19163abb5d079d48e99f1129ccbd42863c1af9a72461e71  build_panel.py
6ff799d402a261f295b810e3cea708d99e1fc4d00f6b0b36a3d4853ccd6eb087  collect.py
0ec6de5e41c69ab261c167fa172a10cd2f1f88794a2212afcd810fda34a85baa  common.py
80482f25a00c6214219e6674b432e28f78e1227c786f7c49bbead691a1af90d0  constructs.py
8db9613116fd62e803883e736cd8dfe55898a2947a7e954e89b5654e6160416a  fetch_prs.py
b2187dcd2b570b52d8891c18fe55cc978d128833120784cda3cccfbde93193e8  finalize_panel.py
e258ad83b96c066f00b137aa80426d97ac401f82f05e657570ccd1dd9bd48dcb  follow.py
800e84bb7c0f3eaf494f7e400aa4ac8b324ab240f2fd59107a759dd2b42716d5  gates.py
233b91a184fa6a788de4a2e97216b465e4805cd5a91b1d9e39886b07ea609016  harness.py
d94dfa090669c898766970b967b689cca1faddf30ab5e48c99099468b660ab1c  jev_label.py
d9accb9d419c4417998274db99c70fbc327331795de3ee64bba2eb70322c4bcb  jev_type.py
6ac3d72083a6cd2642155048f999da503c571868823b360e20f5888c48788506  power.py
687665d7946bce66a5b90d644f04f79ad5b34a64df895eba6993c174b49816c9  sensitivity.py
39e98ad7e3d38366d9cd07153d1979d7b9027bd9134567ffb363b97ab5a7cffe  tests.py
f8643b2d10ac1487a84afad00fcaac7cdfeec03a6c9e7fa4193e7fc0084c526d  validate.py
```
plus `requirements.txt` d01ffb51…, `prompts.json` 0b54289a…, and the 28 frozen data/result artefacts in §10 (`harness-results.json` 0ce59a21…, `power-results.json` 2c924dcb…, `gates.json` 2b184396… — regenerated by design at run time with G3b, `constructs.json` 68325887…, `panel-final.csv.gz` 66ae08f7…, `labels.jsonl` 4712eea7…, `types.jsonl` 7b8eb164…, `pins.json` 5005e10b…, `candidates.jsonl.gz` f9bff032…, `funnel.json`, `repo-gate.json`, `panel.csv.gz`, 16 `data/pr-*.jsonl.gz`). Any change to any of these requires a new review round with zero issues (AGENTS.md rule 5).

**Post-sign-off run order (prereg §10; `U = uv run --python 3.13 --with-requirements requirements.txt python`, `cd evaluation/change-cost-v2`).** Steps 1–9 are already executed and frozen (tests/harness outputs and the Jev-labelled inputs); the remaining steps run now, in this order:

1. `$U tests.py` — done (64 checks, `/tmp/e1bv2-tests-r10.log`)
2. `$U harness.py 200 harness-results.json` — done, stored (digest 0ce59a21…)
3. `$U fetch_prs.py` — done (`pins.json`, `raw/`)
4. `$U collect.py` — done (`data/`)
5. `$U jev_type.py types.jsonl` — done, frozen
6. `$U build_panel.py` — done (`panel.csv.gz`, `funnel.json`, `repo-gate.json`)
7. `$U follow.py` — done (`candidates.jsonl.gz`)
8. `$U jev_label.py candidates.jsonl.gz labels.jsonl` — done, frozen
9. `$U finalize_panel.py` — done (`panel-final.csv.gz`)
10. `$U validate.py sample` → `validation/*-sample.jsonl`, `*-blinded.jsonl`, `RUBRIC.json`, `pair-strata.json`
11. `$U annotate.py` (requires `OPENROUTER_API_KEY`, presence-checked, never printed)
12. `$U validate.py gates` → `validation-gates.json` (G4 fail ⇒ stop and return to council)
13. `$U constructs.py panel-final.csv.gz > constructs.json` — must reproduce digest 68325887…
14. `$U gates.py --g3b > gates.json` — first and only computation of G3b
15. `$U power.py panel-final.csv.gz 60 power-results.json` — must reproduce digest 2c924dcb…
16. `$U analysis.py panel-final.csv.gz results.json --power power-results.json --validation validation-gates.json --gates gates.json`
17. `$U sensitivity.py panel-final.csv.gz sensitivity.json`
18. Write `RESULTS.md` (everything preregistered, uncertainty intervals, `unmeasured` counts, `Deviations`, reproduce commands; T2 and T4 presented together per council decision b); update `docs/hypothesis-ledger.md` for H1/H2/H4.

File this report as `reviews/round-10.md` (the final round per AGENTS.md rule 2).
