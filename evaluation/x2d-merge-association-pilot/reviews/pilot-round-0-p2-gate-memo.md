# Decision memo — X2d round 0: the planted‑gap (P2) gate cannot be built

## 1. Recommendation

**Top up the P2 clone sources from set D (option f1), keep every other rule as signed off, and run one short council round (round 10) on that single change before resuming at `run.sh gates`.**

Concretely: P2 sources become *eligible P units not used by P1 (hash order), then eligible D units (hash order)*, `n_p2 = min(12, eligible_P − 20 + eligible_D)`; the refusal below 8 and `P1 ≥ 20` stay. With the frozen catalogue that gives 20 P1 items + 12 P2 clones (6 from P, 6 from D), restoring the slack R3‑04 asked for. Nothing before `gates` (onboarding maps, catalogue `1220016c…4199`, intruder audit) is touched or re‑run; no label exists, so no outcome is affected. The original refusal is reported in RESULTS.md as the population finding it is.

All three analyses (mine, the experiment‑peer, and the innovation‑challenger on the follow‑up) converge on this. The challenger first preferred G donors; it conceded D after the code check (see §6).

## 2. Why it happened (plain language)

The planted‑gap gate needs units that the catalogue assigns **wholly to one area** (only those can have their area ablated). P has 30 units; the design needed 20 for P1 and ≥ 8 for P2, i.e. **28 of 30 eligible — two misses allowed**. Four missed:

| P cluster | file | why not uniquely owned | knowable before the run? |
|---|---|---|---|
| `30e1d6bd…` | `docs/en/mkdocs.insiders.yml` | outside every onboarding scope | **yes** — `docs/` deliberately unscoped (PREREGISTRATION §4.1 line 81) |
| `9582f013…` | `docs/en/overrides/partials/copyright.html` | outside every onboarding scope | **yes** — same |
| `ac049d4f…` | `docs/language_names.yml` | in the infra scope, but *contested* → no area | no (model outcome) |
| `e7bf5ec4…` | `tests/test_webhooks_security.py` | in scope, *unowned* after Jev | no (model outcome) |

So the two "allowed" misses were already spent by two frozen files (`prepass/reserve.json` + `prepass/scopes.json`) before any model ran; the real ceiling was 28/28 in‑scope units, zero slack. The drafter even knew the number ("at least 28 P units have one unique owner", PREREGISTRATION line 182) and the stub world had to be tuned to `features=3` to reach it (round‑2‑response line 16; the council wrote "the real run can also fall short", round‑3.md line 51) — but nobody computed the scope ceiling on the real reserve. Even at perfect eligibility P could yield only 10 clones, never the 12 the design advertises. This is a design defect in allocation, not a Jev result and not an onboarding failure.

## 3. Options and consequences

Every option changes a frozen file (`build_controls.py` hard‑exits; the prereg routes a real refusal to "the council and the user decide", not to the UNMEASURED line), so **every option needs a prereg revision 10 + council round with zero issues (rule 5)**. The differences are in what else moves.

| Option | Needs | Round‑0 evidence | Cost |
|---|---|---|---|
| **(a) 6 clones, accept UNMEASURED** | lower the build floor (`build_controls.py:272`) | P2 can **never PASS** (needs ≥ 8 label‑derived positives); only FAIL or UNMEASURED; RESULTS must say "R_new on catalogue gaps has no planted test". P1/K5/K6/N1 unchanged. | smallest code change; weakest evidence |
| **(b) P1 = 18** | breaks the preregistered "≥ 20 units of P" (prereg line 104; `build_controls.py:269,294`; analysis enforces 20 too) | exactly 8 clones → re‑creates the zero slack R3‑04 removed; weaker path‑cue test | small; bad trade |
| **(c) extend the reserve** | regenerate `prepass/reserve.json` (+ `prepass.py`); sets are cut by global hash rank (`prepass.py:129‑131`) | reshuffles D/P/G → can move K4/K5/K6/N1 units and break "GUIDE examples from PRs with no D/P cluster"; the extra PRs are older (pre‑2023) survivor‑selected material | largest; touches the most frozen files |
| **(d) G donors (non‑GUIDE PRs)** | G sources must be added as real units and labelled (gold needed for `derive_ablation`); `required_gold_ids` names D/P only (`gold.py:307‑319`); G is the K4 re‑run pool ("fresh draw from G"); `prepass.py:130` says G is never labelled as units | sound, but spends session cap and the K4 reserve | medium |
| **(e) P2_MIN = 6** | constant also read by `analysis_x2d.py` | undoes R3‑04; 5/6 at the 0.80 bound ≈ Wilson lb 0.44 [INFERENCE] | small; weak |
| **(f1) D top‑up — recommended** | `build_controls.py` allocation + prereg §5 sentence + fixtures | 12 clones, ≥ 8 positives possible → gate can PASS/FAIL/UNMEASURED as designed; D already needs gold (prereg line 106), so **no new real units, no new labels beyond the 6 extra clones**, no data‑file change | small |
| (g) reuse P1 sources for P2 | code change | couples the two gates | rejected |

Time cost of (f1): one council round on ~5 files; then `run.sh gates` + `power` from scratch (nothing of round 0 after `quality` exists). Extra labelling ≈ 6 clones × 2.2 × 1.25 ≈ 17 sessions; a few extra Jev requests on `catalogue‑p2`.

## 4. Evidence

- Refusal logic: `build_controls.py:255` (`elig` from `p_rows` only), `:271‑274` (`need = min(12, len(elig) − 20)`, refusal < 8), `:189‑196` (`_unique_owner`: every member exactly one owner, same owner). Eligibility is defined only in code; prereg line 105 says "P2 clones covered units" without naming the pool.
- Four ineligible P files: area membership in the run's `catalogue/catalogue.json` (0 owners each); `docs/language_names.yml` in `contested`; `tests/test_webhooks_security.py` in `unowned`; the two `docs/en/…` files in none of areas/unowned/contested/heuristic_only; `prepass/scopes.json` `unscoped_by_top_dir.docs = 1240`; PREREGISTRATION line 80–81 (infra scope admits only `docs/en/mkdocs.yml`, `docs/language_names.yml`).
- D eligibility: all 36 listed D files checked by directory group; only `tests/test_computed_fields.py` has 0 owners → 19/20 D clusters eligible. (Members are replayed at build time and can only grow, `build_units.py:384‑385`; the 12‑file P cluster `3ea28ee9…` is wholly owned.)
- Refusal is a preregistered stop, distinct from UNMEASURED: PREREGISTRATION line 105 ("a real `build_controls` refusal stops the pilot before any label … the council and the user decide"); `x2d.py:71` (`p2_unmeasured` is a critique line for the `gates()` status).
- Foreseeability: PREREGISTRATION line 182; `reviews/round-2-response.md:16`; `reviews/round-3.md:51,62`; `DRAFT-NOTES.md:95`.
- Isolation of D clones (challenger follow‑up, verified in code): `_clone` sets `control=True, set="ctl"`, fresh ids (`build_controls.py:192‑195`); K4 κ and K6/N1 select original non‑control D/P ids (`analysis_x2d.py:1307‑1320, 1337‑1349`); P2 selects `kind=="p2"` (`:1373‑1376`); contrast set uses original D rows (`build_units.py:408‑430`); dup compares against its own source (`analysis_x2d.py:920‑945`); `tune --round 0` has no P2 branch (`tune.py:68‑84, 357‑389`); dup controls already clone D/P (`build_controls.py:352‑356`).
- Nothing binds round‑0 data to the run commit: `scout_head` is recorded only (`collect_onboard.py:170`); `rounds/0/version.json` is written by `versioning.py --round 0` inside `gates-build`, which has not run; the catalogue digest (`versioning.py:127,202`) is what later rounds check.
- Intruder audit is complete: `onboard/quality.intruder.json` 40/40 both labellers, κ 1.0, `unmeasured 0`.

## 5. Risks

- **Shared‑source dependence.** P2 gold for D‑sourced clones derives from D labels (same as for P‑sourced clones from P labels). A source label error propagates; disclose, don't present as independent replication.
- **Label‑derived positives are not guaranteed.** 12 membership gap positives can still yield < 8 label‑derived ones (source gold keeps a non‑ablated acceptable area, or no agreed gold) → UNMEASURED via the existing path. Expected to be rarer with 12 than with 8; not eliminated.
- **Catalogue‑p2 thins** by up to ~12 of 114 areas (≈ 10 %); a shared D/G construct risk (easier task, absence cue), not a blindness break. Report the exact ablated count.
- **Allocation side‑effects.** The `spare/moved` logic (`build_controls.py:276‑279`) moves placeable P units out of P1 when `need` rises; the rev‑10 fixture must assert P1 stays exactly as the P‑only build would give it (20 items) and that only P2 rows differ.
- **Selection effect (all options).** P1/P2 test only cleanly‑owned units; docs‑config/unscoped/contested files are untested by these gates. State it in RESULTS.md §0.
- **Path twins** between D and P (e.g. `tests/test_forms_single_model.py` in both, PRs 12127/12129) pre‑exist; D‑sourcing adds no new twin class, but the clone of a twin should be reported.
- **Reversibility:** full — no label or score exists; the change is one allocation rule plus fixtures; the P‑only behaviour remains as the first branch.
- **Process record:** this council has now seen which D/P/G units are owned (membership only; allowed under rule 3). Record it in round‑10.md.

## 6. Where the peers changed the memo

- **Experiment‑peer** supplied the foreseeability chain (prereg line 182, round‑2‑response:16, round‑3:51) and the `spare/moved` falsification test; it also ruled out member‑replay and rule‑strictness as causes. Adopted.
- **Innovation‑challenger** argued for G donors; its own follow‑up found no code path by which D clones contaminate K4/N1/contrast/dup, and that G needs new real units, gold‑completeness changes and the K4 reserve. I kept D. Its demand to freeze donor ids/order, ablated‑area count and session plan before any outcome is adopted as a round‑10 condition.
- Not adopted: the challenger's option (g) (coupled gates) and option (c) (reserve extension — most invasive, reshuffles D/P/G).

## 7. Execution outline

1. **Report the refusal** (rule 6): a `Deviations`/population‑finding entry for RESULTS.md with the four ineligible P units and the 28/28 ceiling.
2. **PREREGISTRATION rev 10** (§5 line 105: P2 source pool = P leftovers then D, `n_p2 = min(12, eligible_P − 20 + eligible_D)`, refusal < 8 unchanged, P1 ≥ 20 unchanged, D clones disclosed as source‑dependent; §11 line 182 stub note; §14 change log; status line). No other rule moves.
3. **`build_controls.py`**: build `elig_P` and `elig_D`; donors = unplaceable P → leftover placeable P beyond 20 → D in hash order; `used_src` includes D donors; docstring lines 20–24; return the per‑set donor counts for the manifest/report.
4. **`validation/validate.py`**: (i) P‑short/D‑fills fixture → 12 clones, all membership gap positives, D/P/P1/dup/trap/prose/planted_new rows and `catalogue‑p1` byte‑identical to a P‑only build; (ii) P+D‑short → the same refusal text; keep the existing R3‑04 slack/UNMEASURED/FAIL fixtures. Planted‑effect and planted‑null end‑to‑end runs unchanged (rule 4).
5. **README.md** lines 56/69 and **DRAFT‑NOTES.md** rev‑10 entry; **hypothesis‑ledger** HD6 history line for round 10.
6. **Council round 10** (`reviews/round-10.md`, zero issues; new `signoff.json` digests). Condition: donor ids/order, ablated‑area count, remaining catalogue size and session plan are fixed and recorded before any label.
7. **Resume**: `run.sh validate` once, then `run.sh gates` → `power` in the same worktree (onboard/catalogue/quality reused by catalogue digest `1220016c…4199`), then the round‑0 critique.

**Escalation to the user:** choose between (f1) as recommended and (a) as the fallback if the council will not accept any clone‑source change; (b), (c), (e), (g) are not recommended.
