# X1: council experiment review, round 5 (revision 5 with ADR-014 round 3)

Date: 2026-11-08. Reviewer: innovation council. The lead wrote the synthesis. Two read-only peers ran in parallel:
- `innovation-challenger` (X1Council5.AdversarialChallenger);
- `experiment-peer` (X1Council5.ExperimentPeer).

This record is X1 round 5. It is also ADR-014's third council review. It must be saved byte for byte as `docs/adrs/2026-11-08-x1-delegated-decisions/round-3.md`.

**Governance.**
- `CLAUDE.md` experiment rules 1–7.
- Design rules 1–3, for ADR-014.
- The ledger rules, and the General sections "Toolchain", "Jev labelling", "Model use" and "Secrets".
- Standing owner delegation R-064: where an owner choice arises, this record gives the council's recommended option as the fix.

**Scope.** Revision 5, as listed in `reviews/rev5-digests.json` (77 files; the manifest excludes itself):
- `PREREGISTRATION.md`, `README.md`, `oc-manifest.json` and `inputs/x11-rev4-digests.json`;
- the nine `preparation/` sources;
- `reviews/` (README, rev3 and rev4 digests, rounds 1–4 and their responses);
- 45 `validation/` scripts, outputs and source snapshots;
- ADR-014's `ADR.md`, `round-1.md` and `round-2.md`;
- the shared files `docs/STATUS.md`, `docs/adrs/README.md`, `docs/research-ledger.md`, `docs/hypothesis-ledger.md` and `docs/scout-design.html`.

Each round-4 item was checked against the fix that `reviews/round-4.md` required.

**What was and was not run.**
- No shell, model call, self-test, simulation, benchmark or digest recomputation was run.
- Evidence comes from reading files, jq queries over the JSON, and hand arithmetic.
- One web lookup was made: the CPython 3.12 "What's New" entry on `sum()`.
- `codegraph` was not used. Its reference documents were read first, but this read-only session had no way to run the executable.
- Main reports an independent self-test run: exit 0, 146 checks, 5.4 s, with digests unchanged before and after. This round relies on that report and did not repeat the run.
- Paths are relative to `evaluation/x1-triage-ranking/` unless stated otherwise. "P l.N" means a line of revision 5's `PREREGISTRATION.md`; "M l.N" means a line of `oc-manifest.json`.

**Reviewed files (SHA-256).** Main checked all 77 entries of `reviews/rev5-digests.json` with shasum. Main reports the manifest's own SHA-256 as `ba1d2ae4efdf1c66e6d5f3002b5d0f892bac079a3efe4cbc2d98fd7826f69a95`. This round did not recompute any digest. Both peers found that the 77 entries match the files on disk exactly: 8 in docs, 4 at top level, 9 in preparation, 11 in reviews and 45 in validation.

| File | SHA-256 |
|---|---|
| `docs/STATUS.md` | `f4220d5239e4fc7abc7b763e45c2f04c5c5837d0bd498497069106deb683b8eb` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md` | `cd13f82102bc71ad4212b6c8a384ba820644754dce0e54fe44b599effe99a35a` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-1.md` | `7bc229fd392215971646bfc547883050e82512c21777fbd311cf779e3e394710` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-2.md` | `5a298d77bb778d79c17d158657d868ccdadb0ce8a5eef9d01a8d170a22f976ed` |
| `docs/adrs/README.md` | `c277db8d8703c087b8c922693abcf1be24b7e0d5006d7798e6aef2b4492ec602` |
| `docs/hypothesis-ledger.md` | `ee32984061b9213140a72dcc5c950fe53ae193da367f0aa52fd953916dbd5cc1` |
| `docs/research-ledger.md` | `033fa837abe5d0f227c89fe4c8c3545c59bc6bd17ba2d4d6aab5cb256abcf01e` |
| `docs/scout-design.html` | `f7180b872e7380b8c2a644b8c7fe33b90c9ccf3e3cc0a4c8c3662227dbb3d8a9` |
| `PREREGISTRATION.md` | `d52cf50b21bbbb10a7bf19079d9a11f2ecc87c2fc756dafbee2a4f53c1f1833f` |
| `README.md` | `53267008b73f071de427186d57f8f432ecc6bbb8cb8e564d95adf0e228ac5b49` |
| `inputs/x11-rev4-digests.json` | `e9edba67c885623df6900a08aee1cde463e5fa78f828bbb04ca3d8efd7a0324c` |
| `oc-manifest.json` | `18e5e67e24ba209cfd1d738aceeef34cdff6eaac649bbac8ff3a82dc73ad528b` |
| `preparation/README.md` | `55f6bd97678406391b06d14bcf750502228d1c2043969b3933b92b5902c1e8c7` |
| `preparation/dev-census.json` | `c2c54ee30d2726839c40cfc5bb7f8241dc5bf41e01b1cfb17dd39770bbc4a6e5` |
| `preparation/exposure-recipe.json` | `21ef61d1a782311e4fd09d2554c41784324141ae9e8bb01f3dbbd877194e981f` |
| `preparation/exposure-sources.json` | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` |
| `preparation/frozen-exposure-inputs.json` | `ec92fba6d3f91c671301244539a8caefc5025c0d6b03efca77c5bdfc4bd96425` |
| `preparation/guide-author-isolation.json` | `c1ccff494fde292edc8a556adedb368066c4f6543f0a161921118f9057fb53bd` |
| `preparation/model-free-policy.json` | `87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f` |
| `preparation/pins.json` | `6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc` |
| `preparation/runtime-attestation.json` | `3a924a0c8484eb51b9f62a8b0a805f9d6e5b023a3043e1a0ddd2b64d05e653ab` |
| `reviews/README.md` | `d95e896f9189b5925356c86b60251d43eee56f080cfe445dc05b682a21e08d6f` |
| `reviews/rev3-digests.json` | `894dc63b0522a23589187c6de9fa4e672878fc1c07b85bfbbdcd224e5ab2f448` |
| `reviews/rev4-digests.json` | `01c790837f961cd208fed98b3bd653b514da36e6d8bcee3b2fa4270f88e1876c` |
| `reviews/round-1-response.md` | `66340fd83af645782d5a8772fa7c74e4261d34a7e9d9ff01b8383fc5d5443ffe` |
| `reviews/round-1.md` | `89eab4a5d0f63487da200ab34db28dc0528ed9fcc73745c9d1075272339e172d` |
| `reviews/round-2-response.md` | `4c926bfa0111e458cb23b006101fa605ef57005cb2065a1f6e7260144bc14986` |
| `reviews/round-2.md` | `4021d173533876e775bf9200bc3a1aa1164f04feffbe9f3beb56dfb3b400d1ef` |
| `reviews/round-3-response.md` | `0117db6c3af240beddbef0ebc2a77a9ea886bb00f7aec4a4a984e89184e5b6ad` |
| `reviews/round-3.md` | `7bc229fd392215971646bfc547883050e82512c21777fbd311cf779e3e394710` |
| `reviews/round-4-response.md` | `be9579d04d03512605ad4bd6350cca331eecb39a71a1b9eee0d1888692063739` |
| `reviews/round-4.md` | `5a298d77bb778d79c17d158657d868ccdadb0ce8a5eef9d01a8d170a22f976ed` |
| `validation/calibration-overhead-rev5-final.json` | `52d8f63527643395e0487535b0a23e06fc791357554244f19dadc64e66892f3b` |
| `validation/calibration-overhead-rev5-final2.json` | `5a340ebdfa26f35560e8ccce36f06f437cb1e6ad3e045153101e152b829c73dc` |
| `validation/calibration-overhead-rev5.json` | `b01d5a000bdb8a81b816bdee48f2cd240ed5f0b47a7a9100da0984ede9807073` |
| `validation/calibration-sources-rev5-before-opt.json` | `795283eb79ab1aa480f90ba0de613bcc0620f6b97281b9ece452449bf8ff11eb` |
| `validation/calibration-sources-rev5-before-transition-fix.json` | `5334c47b0260556a1e5a2e19608a202013aa8a4f34ef2ab33a8ac0fde0cf24cd` |
| `validation/compute-baseline-rev4.json` | `44a3b3e8ca82cbb98c3b9a9b8ef04e82053b0184eacf4d4d35f237ba6ffb3497` |
| `validation/compute-benchmark-rev4.json` | `e5f068d7badf2f43afcfbd2ee0c0d0a270efe605329a43a865ae114a7aab1ca8` |
| `validation/compute-benchmark-rev5.json` | `fc086226d92a622f32b289dc6b88a6716fd31573fd98d27dedb200fb9f0ed5df` |
| `validation/compute-projection-rev4-final.json` | `b422c2a6fc7de6e548fa94cc4f36beec403ea9b169aff0429fb01386c44947dd` |
| `validation/compute-projection-rev4-total.json` | `65752e13309281fc8d1019034428103e5f2a445248bc1c142a958b402713a3c1` |
| `validation/compute-projection-rev4.json` | `4d2ca24912b6c94689138ff626869bd7473747a2dfde86497a59925f6e10dc0b` |
| `validation/compute-projection-rev5-final.json` | `3d6c2ff4c76e96c9c215025e4237744e5d6e53540f10966d3e09df6e603638cc` |
| `validation/compute-projection-rev5-final2.json` | `7ebbbb5da4798aac8acc93c6bb1616042075c77a7f62e912027a1ffb7cf8e121` |
| `validation/compute-projection-rev5.json` | `ee616c45674f9beb82bbc8c3951576cf955e1e3c62a2281089fc0b6ec474ad4b` |
| `validation/compute_benchmark.py` | `dd05a8ba08d27120868b515b341a56c6a86bf923da24d92fe6c5e0efafb04c8e` |
| `validation/fixed_population.py` | `5226ad84a708478eed3182b844f1c200812ead9f3308ed86e4def6481b84645e` |
| `validation/interval-diagnostic-rev4.json` | `50d41cc86c0342e05110d322b5d452ffcad7570c612efbbaee705ce11cc18a54` |
| `validation/interval-diagnostic-sources-rev4.json` | `d12d80627cecd6e7ee0b291dee939638cf7752c24ae62164bb495bf0ab693722` |
| `validation/interval-diagnostic-spec-rev4.json` | `e5a93a49c14a4345cba2fd6f2d4ddc55a998eaa199d371299d5965443fe54bb6` |
| `validation/interval_diagnostic.go` | `24f2eaa78e76c1acb7ebe625e3aefa5395b8f76001a9faf6d38206b1720b873c` |
| `validation/interval_diagnostic.py` | `a4b233374bdfec466bfc5862166497a4bbac48b13668aee2697fcb90e2b7c345` |
| `validation/outer-results.json` | `b851c2c1b6ffacee1957ade0071aa9fd38ede0dce5ebd2e3577a2967923f6352` |
| `validation/outer_simulation.py` | `91d07760b61551810282622c2bda53c34fcfcb72a9ffd1798b1a5ffb8ece23f6` |
| `validation/pilot-summary-rev4.json` | `05da900b22b738a61c36d4f02efc76c722109aa2dcffff40fba04f828c14475d` |
| `validation/pilot-summary-rev5.json` | `c21a5a13ea18e521c7fd8469d93b4547471924dd740c75f1c2368fc2ea8ed17e` |
| `validation/ranking.py` | `cad370785167328c95fcbb78390aafe2b1b24d34e4ac8e4aae9544e9d99e276e` |
| `validation/reproduce_reports-initial-rev5.py` | `0b7781db46f78e81a0ac091bf123a3fe1e4045925c09396a6299065119a10ed4` |
| `validation/reproduce_reports-optimized-rev5.py` | `d680ee085bc20525761fdedf924f9d5ac0be12984be9a10c1e3ac4ff5b102887` |
| `validation/reproduce_reports-pilot-rev5.py` | `3ee6476c7d142bc1e7340fa9ba31d668aa42cc92d744c66cc5dc910638cf0ee2` |
| `validation/reproduce_reports.py` | `e2593bcb4112ae80cb82a1f3125d0e3316c9b49e2750875487f6446b0ebf8855` |
| `validation/self_test.py` | `282ffe41e81aab003cbf1f6f57392e03468b79f15a1fd0b90f12779e9cccf5f4` |
| `validation/strong-effect-OC-rev5-closure.json` | `379f2c28ac5fa79ae738d2f1ad18acf823d389ff717616a0d8e37147a4e582fb` |
| `validation/strong-effect-OC-rev5-final.json` | `6f26bad6dcf9f9d7414757ae3ca73b68990187b3d3b52edadd856cb6613a4375` |
| `validation/strong-effect-OC-rev5-optimized.json` | `19c78ee133a87c3f05a9319e58b7bff59031e886837733edb668c566d945615b` |
| `validation/strong-effect-OC-rev5.json` | `089941397a9036830532ffe52410f3f3ca65a153db4e6cfa3e57ec2960ddbb8b` |
| `validation/synthetic-results-rev1.json` | `fbc4cc0de2c948574a76216cffd12cb33b488f0fedd3c3519de4ddf0b5a82277` |
| `validation/synthetic-results-rev2.json` | `24549f3ed2fcaecb357bf747ec31058ae115592f8bf63a1319763d2a84faa77d` |
| `validation/synthetic-results-rev4-percentile.json` | `b206c43600f80dab426f303a915695176418680296105144545e332dca04c56d` |
| `validation/synthetic-results-rev4.json` | `9f3ec381a945028c5a246ea723f03ee771815ff3bfdf609c917505e12c66c2f2` |
| `validation/synthetic-results-rev5-closure.json` | `f0ae8b5c7b341d2f5d932c72a416ae8451f99429c973e7079daad1869324c134` |
| `validation/synthetic-results-rev5-final.json` | `f5ea12d6d3df4043610921e98fb42a8ceea47f973ca0bd5311ba388fdeabe124` |
| `validation/synthetic-results-rev5-optimized.json` | `b235853e5a6b185c8908543a82430c0cb569261dfb200434e3d8d27a19c34392` |
| `validation/synthetic-results-rev5.json` | `a7a2c26a3fbf7eedf60216bddea196364022453769c95c06f7029946385b552b` |
| `validation/synthetic-results.json` | `e65ec36c7d57a0d15f32bc3d2d7bd5f1f76f4148189ec2676653ef996901a350` |
| `validation/synthetic-sources-rev5-initial.json` | `a77aff4f00d87a5d98dafe3114df227b9de27ffb215c45a82ef6074781aee4c2` |

**Byte-preservation checks against the round-4 table.**
- **Unchanged:**
  - every earlier council round and response;
  - `rev3-digests.json` and `inputs/x11-rev4-digests.json`;
  - `synthetic-results-rev1/-rev2/.json` and `outer-results.json`;
  - every revision-4 output (diagnostic spec, result and sources; compute baseline, benchmark and three projections; pilot summary; the two rev4 synthetic results);
  - `interval_diagnostic.py`, `outer_simulation.py` and `docs/scout-design.html`.
- **Identical pairs:** `round-3.md` equals ADR-014 `round-1.md` (`7bc229fd…`), and `round-4.md` equals ADR-014 `round-2.md` (`5a298d77…`).
- **Changed executables that revision 4 had reviewed:**
  - `interval_diagnostic.go` (`184fa317…` → `24f2eaa7…`). The old bytes are archived in `interval-diagnostic-sources-rev4.json`.
  - `ranking.py` (`405f2a4a…` → `cad37078…`), `self_test.py` (`3c47f232…` → `282ffe41…`) and `compute_benchmark.py` (`dabebd88…` → `dd05a8ba…`). Their old bytes are **not** archived (X1-R5-06).

**Verdicts.**
- **X1: 9 issues open (0 blocking, 1 major, 8 minor). No sign-off.**
- **ADR-014: 8 issues open (X1-R5-02 to X1-R5-09). Each needs a change to ADR-014's table or text, and each is shared with X1. No sign-off.** ADR-014 can be adopted only on the joint zero-issue final round.
- **Joint signed set: 9 distinct issues open.**
- No preparation, inventory or draw, Jev, labelling or analysis stage is authorised (§12).

**Severity scale.** Unchanged from rounds 1–4:
- **blocking:** can produce a wrong or unreachable verdict, or breaks experiment rules 1–7 or REQ-169.
- **major:** misleads the result, leaves a preregistered quantity undefined, or makes the study unrepeatable or likely not testable.
- **minor:** text, consistency or provenance. Minor issues must still be fixed before sign-off (rule 1).

---

## 1. Recommendation

**Revise to revision 6, closing X1-R5-01 to X1-R5-09, and submit it for X1 round 6 / ADR-014 round 4. Do not sign off. Authorise no stage beyond drafting and synthetic validation.**

Revision 5 is a strong response. All 13 round-4 findings are met; one (X1-R4-05) is met only in part because of a mismatch between the protocol and the code. Keep all of the following:
- the genuine-population benchmark;
- the fixed-key truth-only calibration with exact IEEE-754 transitions;
- the preserved failed projection;
- the R4-02 acceptance redesign, with OC-verified strong-effect support at 2,000/2,000;
- the 16,128 + 1,792 + 252 grid;
- the κ/M̄ and j/TP̄ safety axes and their schedule;
- the signed-infinity pivots;
- the gradient and verdict-boundary fixtures;
- the full 18-cell diagnostic table;
- the dev source and inventory throughput screen;
- the internal-study rows R-080 and R-081;
- the honest, non-claiming preparation sources.

One defect must be settled before preparation: **the compute gate charges a full trial for every attempt up to the cap, including sampling-gate rejections that never run a trial** (X1-R5-02). That makes the locked-size fit, a binding not-testable stop, depend on p̂_gate falsely. At 1.2 M units the fit fails for every p̂_gate below .4497 [INFERENCE], even though the real trial work does not depend on p̂_gate.

The eight minor issues are cheap synthetic, text or recipe work:
- the zero-variance tolerance;
- the scope of the calibration-equivalence evidence;
- the projection envelope and the cost of candidate enumeration;
- the numerical interpreter;
- revision-4 source provenance;
- the dev-census boundary;
- the preparation scope and model-free checks;
- run-together tokens.

## 2. Alternatives considered

| Option | Description | Tradeoff | Council position |
|---|---|---|---|
| A. Sign off revision 5 and carry the fixes into the implementation round | Freezes the protocol now | Rule 1 requires zero issues. X1-R5-02 sits in a binding stop rule. | Rejected |
| B. **Revision 6 closes all nine issues; then a zero-issue round** | Settles the gate accounting and the preparation scope before the 45-day clock starts | One more round, mostly synthetic and text work | **Recommended** |
| C. Authorise a scoped preparation now | Starts pinning and fetching | Issues are open (rule 1). The preparation actions and clock-start action are undefined (X1-R5-08). | Rejected |
| D. Authorise a model-free inventory or SRS draw | Locks sizes early | Inventory needs signed preparation outputs, an effective model-free policy, runtime closure and an exposure delta review (P l.219; `preparation/README.md`). None exists yet. | Rejected |
| E. Keep the cap-based trial charge and lower p̂_gate's influence some other way | Smaller contract change | Without a hard, p-independent bound on trial executions, the locked fit stays falsely p-dependent | Rejected in favour of X1-R5-02's trial-execution cap |
| F. Cache transition candidates per (locked sizes, prevalence, a_size) instead of measuring their worst case | Removes the enumeration cost altogether | Needs a code change and an equality fixture | Acceptable alternative inside X1-R5-04 |
| G. Finish model-free engineering before preparation: policy-binding harness on synthetic repositories, interpreter and runtime closure, heterogeneous-size benchmark, OC generator and final-evaluator drafts | Shortens the clocked path | More work before the clock starts, none of it touching real data | **Recommended practice**, not a sign-off condition |

## 3. Rule checks

- **Rules 1–2.** This round is recorded with digests. Issues are open, so iteration continues.
- **Rule 3.** No predictor–outcome association was computed on real data. All revision-5 evidence is marked `synthetic_only`. The historical method-selection diagnostic was deliberately not rerun (`self_test.py:363–364`).
- **Rule 4.**
  - Planted effect: support at 2,000/2,000, Wilson MC95 [.99808, 1.0], plus support on fixed seeds 169/170/171.
  - Planted null: not supported, with non-zero width (`strong-effect-OC-rev5-closure.json`).
  - Kernel weighting mutants are retained.
  - The calibration-equivalence evidence is narrower than claimed (X1-R5-03), and the zero-variance pivot does not implement P (X1-R5-01).
- **Rule 5.** Nothing is signed yet. The failed and intermediate projections and calibrations are preserved with their reporter snapshots. The revision-4 executed sources were overwritten without archive or disclosure (X1-R5-06).
- **Rule 6.** All 18 diagnostic cells, the L100 exceedance and the asymmetry reversal are reported (README). Reproduce commands are repository-relative, but they invoke an unpinned interpreter (X1-R5-05).
- **Model use and secrets.**
  - Runtime attestation ran `--version` only.
  - No session was created.
  - Jev-key handling is unchanged (P l.212).
- **Toolchain rule.** The numerical evidence ran on Python 3.14.10 from outside the pinned toolchain (X1-R5-05).
- **Design rules.**
  - ADR-014's format is correct: adopted-form status, an extended route-(a) note and a concurrent-ID note in the index, and STATUS showing effectiveness on the final round.
  - Eight issues require changes to its table or text.
- **Ledgers.**
  - R-080 and R-081 are present (`docs/research-ledger.md:82–83`) and cited.
  - HD9 remains proposed.
  - Next free ID is R-082 (`docs/STATUS.md:128`).

## 4. Verification of round-4 fixes

| Issue | Status | Evidence |
|---|---|---|
| X1-R4-01 (major) | **Fixed.** Residuals in X1-R5-02, -03, -04 | (a) Fixed-key path: P l.169, M l.77, `fixed_population.py:47–195`.<br>(b) Genuine independent L40 and L400 populations (`compute-benchmark-rev5.json` `population_measurements`: `independent_runs: true`, `run_template_replication: false`, 1.2 M units at L400).<br>(c) Committed reporter `reproduce_reports.py:67–121, 155–193` with repository-relative commands; per-unit costs recorded.<br>(d) Fits caps; the counts are unchanged.<br>(e) Locked remeasurement of construction, calibration, trials and attempt cap (P l.163, l.179, l.181).<br>ADR compute row updated. |
| X1-R4-02 (major) | **Fixed** | P l.185. Strong-effect OC precedes the seed checks (`self_test.py:58–61, 74–82`). Null: verdict ≠ support with non-zero width (`:82`). ADR "product parity/recovery" row; `reviews/README.md`. |
| X1-R4-03 | **Fixed** | Smallest attainable ≥ .10/.20 (P l.169; `fixed_population.py:173–196`). Identity kinds, phases, `safety_identity` and `criticality_identity` at M l.40–56. Fixtures at `self_test.py:300–303, 315–340`. |
| X1-R4-04 | **Fixed** | Coverage size factors {0, .05, .20}, 16,128 cells (P l.155, l.161; M l.30–35). Timing reworded as author-recorded. |
| X1-R4-05 | **Partly fixed** | Signed infinities retained and infinite quantiles clipped (`ranking.py:186–203`; `interval_diagnostic.go:19, 28, 54`). Hand fixture and Go differential at `self_test.py:203–209`. **Residual:** P l.113's 1e−14 tolerance is not implemented (X1-R5-01). |
| X1-R4-06 | **Fixed** | No row ≤ M̂ → not testable (P l.175). p̂_gate from 100,000 redraws; cap ceil(4,000/p̂_gate); zero → not testable (P l.163; M l.38, l.58). The cost model for that cap is in X1-R5-02. |
| X1-R4-07 | **Fixed** | κ/M̄ and j/TP̄ axes, the zero-positive degeneracy, interval reuse, and scheduling after binding rows until gold freeze (P l.171, l.220). |
| X1-R4-08 | **Fixed** | Nine gradient fixtures (exact count, prefix discrepancy < 1, first-half concentration) at `self_test.py:284–291`. Five verdict boundaries at `:259–263`. |
| X1-R4-09 | **Fixed as specified** | (a) All 18 cells, with the L100 .03562 exceedance and the 8/9 asymmetry (README).<br>(b) Realised-gain bins (P l.137).<br>(c) Generator and repository-relative commands (README "Reproduce from repository root").<br>(d) Revision-3 loss disclosed (README "Historical source limitation").<br>The **revision-4** sources were lost the same way, without disclosure (X1-R5-06). |
| X1-R4-10 | **Fixed** | Dev source and n_max = 300 (P l.105). Inventory screen (P l.206; 25.909 s × 1.2 ≈ 31.09 s per reserved case). Four threats (P §12, last paragraph). The proposed census boundary is in X1-R5-07. |
| X1-R4-11 | **Fixed** | P l.43, l.137, l.228; `docs/hypothesis-ledger.md:43`. New run-together tokens appear elsewhere (X1-R5-09). |
| X1-R4-12 | **Fixed** | R-080 and R-081 (`docs/research-ledger.md:82–83`); P l.13; ADR "Informed by"; HD9 sources. No collision with sibling allocations (peer check). |
| X1-R4-13 | **Fixed** | ADR claims qualified as author-recorded. `docs/STATUS.md:7, 126, 128` (ADR-014 effective on its final round; R-078/R-080/R-081; next free R-082). `docs/adrs/README.md:12` (route (a) extended; concurrent-ID gap note). |

## 5. New issues

### X1-R5-01: minor. The zero-variance pivot tolerance in P is not implemented, and the two kernels detect zero variance differently

- **Location:** P l.113; `validation/ranking.py:188–193`; `validation/interval_diagnostic.go:9` (`fit`), `:19`, `:54`; fixtures at `self_test.py:203–209`.
- **Problem:**
  - **What P requires.** P l.113 says that a zero-variance replicate with |D*−D̂| ≤ 1e−14 has pivot 0; otherwise the signed infinity is kept.
  - **What the code does.** Python (`elif value == point`) and both Go paths (`d==point`) use exact floating-point equality. Revision 4's executed diagnostic did use the tolerance (`interval-diagnostic-sources-rev4.json`, `ranking.py` `7202f906…`: `abs(value-point) <= 1e-14`).
  - **Consequence.** A difference of one or a few ulps now becomes ±∞, and selecting that quantile replaces an endpoint with ±1.
  - **Different zero tests.** Go computes the variance in one pass as `n·max(0, s2 − s²/n)`. Python sums squared deviations in two passes. The two can disagree on exact zero.
  - **Missing test.** No fixture covers 0 < |Δ| ≤ 1e−14. The existing fixture uses |Δ| = .6.
  - **When it matters** [INFERENCE]. Only when all three repositories have zero replicate variance. That arises in degenerate constant-ratio or census frames, which is the regime P l.137 cites for deterministic support. Python/Go parity cannot detect the deviation, because both share it.
- **Required fix** (recommended option): implement P as written; do not amend P.
  - Use |D*−D̂| ≤ 1e−14 → 0 in Python and in both Go paths.
  - Make zero-variance detection identical in both kernels: one shared variance formula, or one shared exact-zero rule applied to the same computed quantity.
  - Add fixtures with independent expected results at |Δ| ∈ {0, 5e−15, 1e−14, 2e−14}, both signs, on the complete path and the vertex-envelope path, plus a Go/Python differential in that band.
  - Refresh the affected closure evidence. Keep the method selection unchanged.

### X1-R5-02: major. The compute gate charges every capped attempt as a full trial, which makes the binding locked-size fit depend on p̂_gate without cause

- **Location:** `reproduce_reports.py:170–171, 184–191`; `compute-projection-rev5-final2.json` `gate_attempt_contract` and `projected_trial_attempt_multiplier: 4.0`; P l.163, l.181; M l.38, l.58 and the resources section.
- **Problem:**
  - **How trials are charged.** Every population is charged ceil(4,000/p)/2,000 × the maximum trial CPU: four full trials per eligible trial at p = .5.
  - **What a gate rejection costs.** A rejection depends only on run sizes and never runs an audit or bootstrap. The measured cost is 1.33 s per 100,000 checks, about 13 µs each.
  - **How many trials can actually run.** At most the gate-passing attempts within the cap, about p × ceil(4,000/p) ≈ 4,000, whatever p is.
  - **Effect on the fit** [INFERENCE, holding the other inputs fixed]. The fit needs 341.16 + 570.84 × ceil(4,000/p)/2,000 ≤ 2,880, i.e. p̂_gate ≥ .4497. At .44 the projection is about 2,936 h.
  - **Effect on the outcome.** P l.163 and l.181 make a failed locked fit "not testable" after the clock has started. A plausible locked p̂_gate below .45 would end a viable study on an accounting artefact.
  - **Overstatement.** At p = .5 trial CPU is overstated about 2× (2,283 h charged against at most about 1,142 h runnable).
  - **Gap in P.** P never defines the cost model for gate-rejected attempts.
- **Required fix** (recommended option):
  - Keep the R4-06 attempt cap.
  - Add a **p-independent cap of 4,000 trial executions per population**. Trial executions are gate-passing attempts, including adequacy failures. Exhaustion means unmeasured and the cell fails.
  - Project gate checks at measured gate-check CPU × attempt cap, and trials at 4,000 × the maximum trial CPU, plus 2× overhead.
  - Recompute the authoritative projection with the committed reporter and preserve `-final2` as superseded.
  - Update P l.163 and l.181, the manifest, README, the ADR rows on attempt cap and compute, and R-081.

### X1-R5-03: minor. The calibration-equivalence evidence is narrower than claimed: the "full-regeneration" reference is the code under test, and the fixtures never exercise caps, failures or offsets

- **Location:** `ranking.py:349–359` (`synthetic` = `prepare_population` + `evaluate_population`); `self_test.py:292–314`; claims in P l.169, M l.77, the README, the ADR "power/null calibration" row and the response.
- **Problem:**
  - **The regeneration fixture is near-tautological.** `self_test.py:296` compares a reused fixed-key population with `synthetic()`, which calls the same two helpers. It shows only deterministic fresh construction of a frozen dataclass. It is not equivalence with an independent regeneration path.
  - **The other fixtures are narrow.**
    - The meaningful checks are truth-only = full-frame truth at three gains (`:297–299`) and optimised calibration = full-frame-oracle calibration (`:304–314`).
    - Both use only run_count 2, sizes {50, 100}, prevalence .20, cap 400 (which never binds), failure 0, overlap 0, a_size 0, scanner false and the uniform default.
    - The benchmarked profile (cap 400 binding, failure .02, overlap .5, prevalence .55, a_size .05, linear gradient; `reproduce_reports.py:73–77`) is not among them.
  - **What the code still guarantees.** On reading, equality holds structurally: both paths share `_run_order` and integer per-repository sums. The binary search's monotonicity also holds on reading [INFERENCE: one positive enters and one negative leaves the desired set per step, so captured top-k positives cannot fall]. The evidence and the signed claims overstate what was tested.
- **Required fix:**
  - Rename the `:296` check as a fresh-construction determinism check, and correct the five claims.
  - Extend both equality fixtures, including targets 0/.10/.20, to populations with:
    - binding caps (sizes ≥ 500 with caps 400 and 850);
    - failure .02, overlap .5, prevalence .10 and .55, a_size .05 and .20;
    - linear gradient, scanner_true and heterogeneity .10;
    - the benchmarked profile at reduced L.
  - Add a fail-closed assertion in `calibrate_population`: `_population_gain(prepared, latent_gain) == realised_gain(frame)` for the returned frame. Every OC calibration then checks itself.

### X1-R5-04: minor. The authoritative fit holds only inside an unstated envelope, and candidate enumeration was benchmarked on three distinct run sizes

- **Location:** `compute-projection-rev5-final2.json` `population_extrapolation` (400 runs × mean 1,000 = 1.2 M units); `reproduce_reports.py:73` (sizes 2,000/500/1,000 only); `fixed_population.py:148–170`; `calibration-overhead-rev5-final2.json` (`candidate_truth_aggregation_and_bookkeeping_cpu_seconds` .032 s at L40 and .242 s at L400).
- **Problem:**
  - **Population-dependent cost.** It is 341.16 h. At p = .5 the current fit holds up to about 2.10 M units per population (about L700 at mean 1,000). L2,000 at mean 1,000 gives about 3,989 h [INFERENCE: linear per-unit scaling].
  - **Candidate enumeration.** Its cost scales with the number of distinct (positives, min(k, P), offset) kinds, not with units. The search runs over about 63 float indices per transition. With every size from 20 to 2,000 (≤ 601,136 candidates, about 5.4 × 10⁵ searched transitions) it costs roughly 9–20 s per calibration, or about 60–130 core-hours across 11,900 calibrations at 2× overhead [INFERENCE]. That is up to half of the 255 h margin. The benchmark's three sizes cannot show it, although P l.169 requires "measure all candidate/truth/final-frame work".
  - **Memory.** Peak memory per worker is unmeasured.
- **Required fix:**
  - After X1-R5-02, publish the break-even envelope (units per population and p̂_gate) in P §8, the README and the ADR compute row.
  - Benchmark candidate enumeration on a worst-case vector with every size from 20 to 2,000, at prevalence .55 and a_size .05, and charge it per calibrated population. Alternatively, register caching per (locked sizes, prevalence, a_size) with an equality fixture and charge per cache key.
  - Record peak resident memory for L400 construction, calibration and trials, and project the 8-worker footprint.

### X1-R5-05: minor. The numerical interpreter is outside the pinned toolchain

- **Location:**
  - `preparation/runtime-attestation.json`: Python 3.14.10 at `~/.local/bin/python`;
  - `toolchain/versions.env:11` (`PYTHON_VERSION=3.11.11`);
  - `scripts/bootstrap-tools.sh:156–164`: `.tools/bin` links only semgrep, ruff, lizard and mypy from the 3.11.11 venv, not `python`;
  - every `reproduce_cli` and the README commands (`source .tools/activate && python …`).
- **Problem:**
  - **Where `python` comes from.** `source .tools/activate && python` resolves to an unpinned interpreter.
  - **Wrong framing.** The attestation calls the gap a scanner-compatibility question. In fact the scanner venv is already built on 3.11.11, and the unpinned component is the numerical interpreter.
  - **Results are interpreter-dependent.** CPython 3.12 changed `sum()` of floats to Neumaier compensated summation ([What's New 3.12](https://docs.python.org/3.12/whatsnew/3.12.html); gh-100425). The kernels sum floats throughout, so the evidence is not byte-reproducible under the pinned 3.11.11.
  - **Costs are interpreter-dependent too.** Python-side construction and calibration costs depend on the interpreter.
  - P l.179 requires source, binary and runtime closure.
- **Required fix** (recommended option):
  - Run the X1 numerical scripts with the toolchain-managed CPython 3.11.11 by explicit path (`.tools/venv/bin/python`).
  - Regenerate the closure self-test, strong-effect, benchmark, calibration-overhead and projection outputs under it, preserving the 3.14.10 outputs as superseded.
  - Name the interpreter path and SHA-256 in the attestation and the reproduce commands, and correct the attestation's framing.
  - If the code needs ≥ 3.12, make a reviewed `versions.env` change that adds a dedicated numerical-interpreter pin instead.
  - Record the CPython `sum()` source in the research ledger (next free R-082) if the decision cites it.

### X1-R5-06: minor. The revision-4 executed sources were overwritten without archive or disclosure

- **Location:**
  - `reviews/rev4-digests.json:113, 122–123` and the round-4 table: `ranking.py` `405f2a4a…`, `self_test.py` `3c47f232…`, `compute_benchmark.py` `dabebd88…`;
  - preserved outputs `synthetic-results-rev4.json:276–278` and `compute-benchmark-rev4.json`;
  - README "Historical source limitation" (revision 3 only); the ADR provenance row.
- **Problem:**
  - Round 4 reviewed these bytes, and they produced preserved outputs.
  - None of the revision-5 snapshots contains them: `synthetic-sources-rev5-initial.json` and `calibration-sources-rev5-before-opt.json` already hold `cad37078…`.
  - This is the same loss X1-R4-09(d) found for revision 3, and this time it is not disclosed.
- **Required fix:**
  - Archive the three files' exact bytes if any preserved copy exists. Otherwise disclose the loss exactly as for revision 3, in the README, the next digest manifest and the ADR provenance row.
  - Adopt a snapshot-before-overwrite rule: every executable in a reviewed set is archived in a source-snapshot JSON before modification.

### X1-R5-07: minor. The proposed dev-census boundary takes the oldest units from the repository root

- **Location:** `preparation/dev-census.json` `chronological_census_recipe` [0] and [2]; P l.105, l.190, §10 (c_U, durations, r̂); ADR row "Which concrete preparation inputs?".
- **Problem:**
  - **What the recipe does.** It takes the first 150 eligible first-occurrence units per repository, oldest first, from complete first-parent ancestry starting at the root. The recipe itself labels this boundary "a proposed preparation choice requiring review".
  - **Different era** [INFERENCE]. The units come from fastapi's and axum's earliest history, a different era from the 2024-onward test windows (P §3).
  - **Clustering** [INFERENCE]. They may cluster in a few early bulk commits.
  - **Why it matters.** Dev quantities are binding, not only descriptive:
    - c_U sets M̂ and the selected cap row, and uses a per-run-cluster bootstrap that degenerates with few clusters;
    - Wilson r_L assumes independent cases;
    - Maurer–Pontil U95 drives the throughput stop.
  - **Construction mismatch.** The test frame is built differently: a 2,000-commit window after an exclusive C0.
- **Required fix** (recommended option):
  - **Window.** Mirror the test frame. The window is the last 2,000 first-parent commits ending at the pinned inclusive tip. Exclusive C0_dev is the 2,000th first-parent ancestor, or full ancestry minus the root if shorter. Use §3's first-occurrence and exclusion rules, with issues pre-existing at C0_dev excluded.
  - **Selection.** Take 150 eligible units per repository in ascending HMAC-SHA-256 order (UTF-8 key `169-x1-dev`, message repository NUL commit SHA NUL issue ID), at most 5 units per commit, so each repository has ≥ 30 distinct commits.
  - **Supply.** Insufficient supply means insufficient before dev models.
  - **Reporting and threats.** Report distinct commits per repository. Add within-commit dependence to the threats.
  - **Text.** Rename "census" to "keyed development sample" in P §5, §9 and §10, the ADR and the recipe. Count the dev scans in the inventory projection.

### X1-R5-08: minor. The preparation stage's actions, the clock-start action and the model-free check are prose, and the stages for dev collection and the GUIDE-author session are not named

- **Location:** P l.218–220; `preparation/README.md` ("`policy.Load` reads base-ref policy … a working-directory copy is not an effective-policy override"); `preparation/model-free-policy.json`; `preparation/guide-author-isolation.json`; ADR row "Which concrete preparation inputs?".
- **Problem:**
  - **Undefined actions.** P l.218 lists the reviews required before preparation, but not what preparation may do or which action starts the 45-day clock.
  - **No effective policy.** The model-free policy file has no effect as it stands, and no falsifiable check precedes scans.
  - **Unnamed stages.** No stage authorises dev-sample collection (model-free scans of fastapi and axum). No stage authorises the isolated GUIDE-author session, which is a model call.
- **Required fix** (recommended option):
  - **(a) Preparation permits only:**
    1. fetching pinned objects for the three test and two dev repositories from their canonical URLs, verifying every pinned SHA;
    2. bootstrapping and attesting the installed scanner, dependency and numerical-interpreter closure (X1-R5-05);
    3. implementing the effective model-free policy binding for commit and parent and testing it with fixtures;
    4. the X1 identifier completeness/delta review;
    5. sealing preparation outputs for inventory review.

    No frame or dev scans, SRS, model calls, or score or label extraction. The clock starts at the first action executed under preparation authority.
  - **(b) Before any scan, a falsifiable model-free check on a synthetic repository and on one pinned commit/parent pair:**
    - the effective policy at both commit and parent resolves `decisions.enabled=false`, `model_cli.lanes=[]`, `mapping.association.enabled=false` and heuristic mapping;
    - `OPENROUTER_API_KEY` is absent (presence check only);
    - no `claude` or `codex` process starts (argv log);
    - issue IDs are identical to a scan without the override.
  - **(c) Name the stages.** Dev-sample collection belongs to inventory. One isolated GUIDE-author session (`claude -p` or `codex exec`, per the recipe) belongs to stage 3, and the GUIDE plus its actual isolation record need a zero-issue review before any dev label session.

### X1-R5-09: minor. Run-together tokens in signed text (the same defect class as X1-R4-11)

- **Location:**
  - P l.181: "cap8,000", "total3,224.7242595555554", "remaining45-day".
  - ADR rows: "census150/repository (n_max300)", "inertia85e7329", "count3", "at601,136", "mean1,000/max2,000", "B1999", "All18", "R064", "ADR014".
  - R-081 row (`docs/research-ledger.md:83`): "projected2,624…", "within2,880/15", "initial3,224…".
  - Manifest strings, e.g. M l.77 "at most601136".
  - `preparation/README.md`: "Go1.26.10/Python3.14.10/Claude2.1.311/Codex0.166.0", "Python3.11.11", "used3.14.10".
- **Required fix:** add the missing spaces in every signed document. Generated outputs stay immutable.

## 6. Assessment of the targeted questions

- **Calibration optimisation: exact as implemented; the evidence is too narrow.**
  - Truth-only and materialised evaluation share `_run_order` and integer per-repository sums, and add the three repositories in the same order (`fixed_population.py:95–135`). The results therefore match the materialised path for any profile, on reading.
  - The exact-transition repair is sound: adjacent-float fixtures over P5, mixed sizes and size offsets (`self_test.py:322–334`), and the P5 regression (`:315–321`).
  - The measured API is the one the projection charges: candidates, ≤ 20 truth searches padded to 20 at the maximum observed step cost, and one final frame (`reproduce_reports.py:96–113, 167–183`).
  - Gaps: the fixture domain (X1-R5-03) and enumeration cost at realistic size diversity (X1-R5-04).
  - The benchmarked target was unreachable (realised .0767/.0770 against .20), so the search took 9 steps and was padded to 20. That is acceptable for cost.
- **Projection of 2,624.5 h / 13.7 days against 2,880 h / 15 days.**
  - The arithmetic reproduces exactly (§8), and every registered term is present.
  - The margin is 255.47 h (8.9% of the cap).
  - Measurement noise is small: maxima are used, worker-count differences are 1–3%, and `-final` and `-final2` differ by about 1.5%.
  - The binding threats are structural: the false p̂_gate dependence (X1-R5-02), the units and distinct-size envelope (X1-R5-04), and 8-worker efficiency (§9).
  - After X1-R5-02 the indicative total is about 1,489 h, and the break-even moves to about 6.1 M units per population [INFERENCE].
- **Preserved failed projection.**
  - Preserved honestly: `fits_caps: false`, 3,224.7242595555554 h / 16.795 d, with its reporter snapshot `0b7781db…`.
  - Only the calibration terms changed (597.98 + 41.61 h → 39.40 h). Counts, trial costs, overhead and p_gate are identical. This is a change of compute method, exact by construction, not a disguised grid or cap reduction.
  - The fix was chosen after seeing the failure, and that is fully disclosed. It is an acceptable forking path because the data are synthetic and no gate was loosened.
  - "Main approved" is author-reported and is not council approval. This round is the council review.
- **GUIDE, policy and runtime closure.**
  - The records are honest and claim nothing: no GUIDE exists, no isolation is attested, and the policy is a proposal with no effect.
  - They are prose rather than checks; X1-R5-08 makes them checks.
  - Runtime: Go, Claude and Codex digests are attested. The numerical interpreter is the real gap (X1-R5-05).
- **ADR-014.**
  - Sequential ID; adopted-form status; Review defers to the final round in its folder.
  - Records list rounds 1–2 and later rounds.
  - Index route (a) and gap note present; STATUS correct; the card (`docs/scout-design.html`, unchanged `f7180b87…`) is still accurate.
  - The table agrees with P on every row checked. It needs updates for X1-R5-02 to X1-R5-09.

## 7. Checked and sound

- **Cell counts:**
  - coverage 7 × 4 × 2 × 6 × 2 × 2 × 2 × 2 × 3 = 16,128;
  - references 896 × 2 = 1,792;
  - missing 7 × 2 × 3 × 2 × 3 = 252;
  - safety 336 × 24 = 8,064;
  - calibrated populations (1,792 + 252 + 336) × 5 = 11,900.
- **Binding path:** 25.0080 + 184.1913 + 39.4020 × 10,220/11,900 + 373.0193 + 248.2644 = 864.32 h.
- **Strong effect:**
  - realised .463; intervals about [.43, .49];
  - fixed-seed frames have 20 runs per repository with ≤ 60,000 units;
  - the null frames have unequal π and gain exactly 0, with intervals such as [−.0404, .0436].
- **Manifest completeness:** 77 entries equal the files on disk. The joint-record digest `5a298d77…` matches both copies.
- **Ledgers:** R-080 and R-081 rows have the correct links and digests; HD9 is proposed; no ID collision.
- **Delivery, gradient, verdict, identity and safety-axis fixtures:** as in §4.

## 8. Evidence index

- **Projection** (`compute-projection-rev5-final2.json`):
  - 25.0080 + 0.4624 + 184.1913 + 39.4020 + 92.0956 + 373.0193 + 248.2644 + 6.9941 + 1,655.0957 = 2,624.5328 h;
  - ÷ 192 = 13.6694 days.
- **p̂_gate break-even** [INFERENCE]:
  - trial cost per unit of multiplier = 2,283.3735 / 4 = 570.84 h;
  - (2,880 − 341.16) / 570.84 = 4.4475;
  - p ≥ 2 / 4.4475 = .4497.
- **Unit envelope** [INFERENCE]:
  - 1.2 M × (2,880 − 2,283.37) / 341.16 = 2.10 M units;
  - L2,000 at mean 1,000: 2,283.37 + 5 × 341.16 = 3,989.2 h.
- **After the X1-R5-02 fix** [INFERENCE]:
  - 341.16 + 1,141.69 + about 6 h of gate checks (100,940 populations × 8,000 × 13.3 µs × 2) ≈ 1,489 h.
- **Enumeration worst case** [INFERENCE]:
  - about 0.27 × Σn over sizes 20–2,000 ≈ 5.4 × 10⁵ transitions × about 63 steps ≈ 3.4 × 10⁷ evaluations;
  - at 0.25–0.6 µs each, 9–20 s per calibration;
  - × 11,900 × 2 ⇒ about 60–130 core-hours.
- **Eight-worker efficiency needed:** 2,624.53 / (15 × 24) = 7.29 effective workers, i.e. 91.1% [INFERENCE].
- **Zero-variance divergence:**
  - P l.113;
  - `ranking.py:190`;
  - `interval_diagnostic.go:19, 54`;
  - revision-4 code `abs(value-point) <= 1e-14` (`interval-diagnostic-sources-rev4.json`).
- **Interpreter:** `scripts/bootstrap-tools.sh:156–164` (no `python` link); `toolchain/versions.env:11`; attestation 3.14.10.

## 9. Risks

- **Resource fit.**
  - Even after X1-R5-02 and X1-R5-04, the locked-size remeasurement can still fail legitimately: very large envelopes, low efficiency, or memory. P makes that "not testable", without a grid cut or cap extension. This is correct, but it happens after the clock starts.
  - Option G and the published envelope reduce the surprise.
- **Eight-worker scaling.** Only 1 and 2 workers were measured. The machine's split of performance and efficiency cores, its thermals and its memory are unknown [INFERENCE].
- **Strong-effect acceptance is deliberately easy.** Realised .463 catches gross faults only. Near-threshold behaviour rests on the power references and the tail gate.
- **Shared-file staleness.** Whole-file digests of STATUS, both ledgers, the ADR index and the design document go stale if a sibling lands first. The final sign-off should also name the X1-owned extracts (the manifest already lists them).
- **Clock.** Preparation sub-stages, the inventory review and the implementation reviews all run inside the 45-day clock unless option G is followed.
- **Reversibility.** Every fix is pre-data and synthetic, or text or recipe work. No product, default, accepted ADR or requirement changes.

## 10. Where the peers changed the council's recommendation

- **Experiment peer:**
  - Rated the attempt-cost model **major**. The lead had it as minor. **Adopted as major X1-R5-02**, because it sits in a binding not-testable rule and the trigger (p̂_gate < .45) is plausible [INFERENCE]. The peer's trial-execution cap was adopted as the fix.
  - Supplied the CPython 3.12 `sum()` change. This strengthened **X1-R5-05** from a provenance gap into a byte-reproducibility defect.
  - Found the undisclosed revision-4 source loss. **Adopted as X1-R5-06** (lead had not found it).
  - Proposed the falsifiable model-free check and the clock-start definition (**X1-R5-08(b)**), the runtime self-check assertion (**X1-R5-03**) and peak-memory measurement (**X1-R5-04**). All adopted.
  - Proposed a keyed, per-commit-capped dev sample. Adopted and made concrete as **X1-R5-07**: a window mirroring the test frame, ≤ 5 units per commit, ≥ 30 commits.
- **Challenger:**
  - Rated the zero-variance tolerance **major**. **Downgraded to minor X1-R5-01**: a verdict effect needs every repository at exact zero replicate variance (a degenerate regime), and R4-05 itself was minor. The finding stands.
  - Its sensitivity figures (p ≥ .44969, 3,989 h at L2,000, 91.13% efficiency) were adopted into §6, §8 and §9.
  - Its candidate-caching alternative was adopted as option F.
  - Independently confirmed manifest completeness and the honesty of the failed projection.
- **Lead only:**
  - Go and Python detect zero variance with different formulas (part of X1-R5-01).
  - Monotonicity of the binary search, checked by reading.
  - The calibration benchmark's unreachable target and its three sizes; enumeration scales with distinct sizes (X1-R5-04, also raised by both peers).
  - `.tools/activate` provides no `python` (X1-R5-05).
  - No stage is named for dev collection or the GUIDE-author session (X1-R5-08(c)).

## 11. Execution outline for revision 6 (one integration owner)

1. **Gate accounting** (X1-R5-02, X1-R5-04):
   - trial-execution cap and gate-check costing;
   - worst-case enumeration benchmark (or registered caching with a fixture);
   - peak memory;
   - recomputed projection and published envelope.

   *Test:* the new total and envelope fit 2,880 h and 15 days, and the old `-final2` is preserved.
2. **Kernel** (X1-R5-01): tolerance in Python and Go, a unified zero-variance rule, boundary fixtures and a Go differential. *Test:* the new fixtures pass, and parity holds to ≤ 1e−12.
3. **Calibration evidence** (X1-R5-03): rename the determinism check, add the broad equality fixtures and the fail-closed assertion. *Test:* all equality checks pass on the benchmarked profile.
4. **Runtime** (X1-R5-05): run under the explicit pinned 3.11.11, regenerate the closure outputs, and correct the attestation and commands. If this fails, make a reviewed `versions.env` change. *Test:* every output records the interpreter path and digest.
5. **Provenance** (X1-R5-06): archive or disclose the revision-4 bytes; adopt snapshot-before-overwrite.
6. **Preparation contract** (X1-R5-07, X1-R5-08): the windowed keyed dev sample; preparation actions and clock start; the model-free check; stage names for dev collection and the GUIDE session.
7. **Text and trackers** (X1-R5-09): respace tokens; update the ADR-014 table for X1-R5-02 to -09, R-081, README and reviews README.
8. **Run and submit.** From the pinned toolchain, run the self-test, benchmark, calibration and projection once each, with create-once outputs. Record every digest and submit X1 round 6, saved also as ADR-014 `round-4.md`.

## 12. Stage authorisation

| Stage | Status now | Earliest path |
|---|---|---|
| Drafting, synthetic validation, and option G model-free engineering on synthetic repositories | **Permitted** (unchanged) | — |
| Preparation | **Not authorised**: 9 open issues (rule 1); preparation actions and clock start undefined (X1-R5-08); binding gate accounting defective (X1-R5-02) | A zero-issue round on revision 6, then preparation limited to the X1-R5-08(a) actions |
| Inventory, model-free scans and SRS draw (including dev-sample collection) | **Not authorised** | After sealed preparation outputs, a passing X1-R5-08(b) check, runtime closure, the exposure delta review and a scoped zero-issue inventory review |
| Full OC and implementation (including the isolated GUIDE session) | **Not authorised** | After inventory locks sizes. The locked-size fit, binding rows and product-backed rule-4 frames must pass; GUIDE and final evaluator reviewed before any dev label or model call |
| Dev Jev and labels; test scoring and labels | **Not authorised** | Under full authority only |
| Gold unblinding and analysis | **Not authorised** | — |
| ADR-014 adoption | **Not authorised** | The joint zero-issue final round saved in its folder |
| Default, threshold or product-sampler change | **Not authorised** | A separate signed ADR |

**Next stage authorised:** none beyond drafting and synthetic validation. A model-free draw cannot be the next stage under any outcome of this round, because it needs reviewed preparation outputs first.

**Sign-off statement:** none. Round 5 closes with:
- **X1: 9 issues open**:
  - **0 blocking**;
  - **1 major:** X1-R5-02;
  - **8 minor:** X1-R5-01 and X1-R5-03 to X1-R5-09.
- **ADR-014: 8 issues open** (X1-R5-02 to X1-R5-09, each needing ADR-014 table or text changes). Adoption also waits for the joint zero-issue X1 round.

The reviewed files are those in `reviews/rev5-digests.json` (manifest SHA-256 `ba1d2ae4efdf1c66e6d5f3002b5d0f892bac079a3efe4cbc2d98fd7826f69a95`, as reported by Main), with the per-file digests in the table above. No preparation, inventory, draw, Jev, labelling or analysis stage is authorised, and ADR-014 is not adopted.
