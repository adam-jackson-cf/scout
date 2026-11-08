# X1: council experiment review, round 6 (revision 6 with ADR-014 round 4)

Date: 2026-11-08. Reviewer: innovation council. The lead wrote the synthesis. Two read-only peers ran in parallel:
- `innovation-challenger` (X1Council6.X1Council6Challenger);
- `experiment-peer` (X1Council6.X1Council6ExperimentPeer).

This record is X1 round 6 and ADR-014's fourth council review. Save it byte for byte as both `evaluation/x1-triage-ranking/reviews/round-6.md` and `docs/adrs/2026-11-08-x1-delegated-decisions/round-4.md`.

**Governance.**
- `CLAUDE.md` experiment rules 1–7.
- Design rules 1–3, for ADR-014.
- The ledger rules.
- The General sections "Toolchain", "Jev labelling", "Model use" and "Secrets".
- Standing owner delegation R-064: where an owner choice comes up, this record gives the council's recommended option as the fix.

**Scope.** Revision 6, as listed in `reviews/rev6-digests.json`. The manifest has 91 files and excludes itself. Each round-5 item was checked against the fix that `reviews/round-5.md` §5 required.

**What was and was not run.**
- No shell, model call, self-test, simulation, benchmark, disassembly or digest recomputation was run.
- The evidence comes from:
  - reading files and grep;
  - hand arithmetic;
  - two web lookups on Go floating-point fusion: the Go specification, and `cmd/compile/internal/ssa/_gen/ARM64.rules` at https://go.dev/src/cmd/compile/internal/ssa/_gen/ARM64.rules, lines 1796–1806.
- `codegraph` was not used. Its reference documents were read first, but this read-only session had no way to run the executable.
- Main reports an independent self-test by a separate agent. It used the README entrypoint with `.tools/venv/bin/python` 3.11.11 and passed 241 checks (planted effect supported, planted null not supported), with exit 0. It took 13.1 s, and 91 of 91 digests matched before and after. This round relies on that report and did not repeat the run.
- Paths are relative to `evaluation/x1-triage-ranking/` unless stated otherwise. "P l.N" means a line of revision 6's `PREREGISTRATION.md`; "M l.N" means a line of `oc-manifest.json`.

**Reviewed files (SHA-256).** Main checked all 91 entries of `reviews/rev6-digests.json` with shasum. Main reports the manifest's own SHA-256 as `cca1baf432389b094588b6085888e8116442f0ee9baef15cbc750ef30086a746`. This round did not recompute any digest.

| File | SHA-256 |
|---|---|
| `docs/STATUS.md` | `b0896f16e1fdbab2e6c1d73590642357267e927524b8f96783fb6645dd2407e4` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md` | `424669ed87407e56103496bdb1733466d551c6a9c233b256136c5ae783011ac1` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-1.md` | `7bc229fd392215971646bfc547883050e82512c21777fbd311cf779e3e394710` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-2.md` | `5a298d77bb778d79c17d158657d868ccdadb0ce8a5eef9d01a8d170a22f976ed` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-3.md` | `9530480f20aa59b2c4c72047dfab46c263d73dd86e716a5a566fee4d54f87038` |
| `docs/adrs/README.md` | `c277db8d8703c087b8c922693abcf1be24b7e0d5006d7798e6aef2b4492ec602` |
| `docs/hypothesis-ledger.md` | `595450122ff2c67efb071b185bc1451870e396e84ccefde35d34c9f92862b50a` |
| `docs/research-ledger.md` | `c9e078717d011c91b7a90a0f10f22854e895ce19f5829068d8a94bfb5611f9d7` |
| `docs/scout-design.html` | `f7180b872e7380b8c2a644b8c7fe33b90c9ccf3e3cc0a4c8c3662227dbb3d8a9` |
| `PREREGISTRATION.md` | `893dfa24fdfe3f7e5ccffce0280338351dec9479f3f48d016a28e54dc5ae7bbd` |
| `README.md` | `1e3a12f60e6d917a868364ec2dcd4d1d78099aada228028b5667abeb39967ab2` |
| `inputs/x11-rev4-digests.json` | `e9edba67c885623df6900a08aee1cde463e5fa78f828bbb04ca3d8efd7a0324c` |
| `oc-manifest.json` | `983199eed88ea51bdf2754749b58bdacc354e1fd3cfe5a958be7c27f8f909ab7` |
| `preparation/README.md` | `3fa35613a020a18f776127286498b31b833975568361be323f9518da4519bf2b` |
| `preparation/dev-census.json` | `8b5aba45ff98db94c516eea23c38ba0702f586e677669de15e9c2a7ff21152ab` |
| `preparation/exposure-recipe.json` | `21ef61d1a782311e4fd09d2554c41784324141ae9e8bb01f3dbbd877194e981f` |
| `preparation/exposure-sources.json` | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` |
| `preparation/frozen-exposure-inputs.json` | `ec92fba6d3f91c671301244539a8caefc5025c0d6b03efca77c5bdfc4bd96425` |
| `preparation/guide-author-isolation.json` | `3418a1ce247e021bf5e689207d40e73c51aafd3ba3cfaa8674edc008233b55ac` |
| `preparation/model-free-binding-check.json` | `ab8649a6763cf8b928010b16c95c095133a8520524bf407f455d8841b03c70c1` |
| `preparation/model-free-policy.json` | `87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f` |
| `preparation/pins.json` | `6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc` |
| `preparation/runtime-attestation-rev6.json` | `19463957becfc23a391eac03582b32eaa0ff392069bccd9aed9000f390ad2d62` |
| `preparation/runtime-attestation.json` | `3a924a0c8484eb51b9f62a8b0a805f9d6e5b023a3043e1a0ddd2b64d05e653ab` |
| `reviews/README.md` | `5bb2ebf04f696ffddaf3af23903bba686a42a8215dd5a806017811388e7bcc3c` |
| `reviews/rev3-digests.json` | `894dc63b0522a23589187c6de9fa4e672878fc1c07b85bfbbdcd224e5ab2f448` |
| `reviews/rev4-digests.json` | `01c790837f961cd208fed98b3bd653b514da36e6d8bcee3b2fa4270f88e1876c` |
| `reviews/rev5-digests.json` | `ba1d2ae4efdf1c66e6d5f3002b5d0f892bac079a3efe4cbc2d98fd7826f69a95` |
| `reviews/round-1-response.md` | `66340fd83af645782d5a8772fa7c74e4261d34a7e9d9ff01b8383fc5d5443ffe` |
| `reviews/round-1.md` | `89eab4a5d0f63487da200ab34db28dc0528ed9fcc73745c9d1075272339e172d` |
| `reviews/round-2-response.md` | `4c926bfa0111e458cb23b006101fa605ef57005cb2065a1f6e7260144bc14986` |
| `reviews/round-2.md` | `4021d173533876e775bf9200bc3a1aa1164f04feffbe9f3beb56dfb3b400d1ef` |
| `reviews/round-3-response.md` | `0117db6c3af240beddbef0ebc2a77a9ea886bb00f7aec4a4a984e89184e5b6ad` |
| `reviews/round-3.md` | `7bc229fd392215971646bfc547883050e82512c21777fbd311cf779e3e394710` |
| `reviews/round-4-response.md` | `be9579d04d03512605ad4bd6350cca331eecb39a71a1b9eee0d1888692063739` |
| `reviews/round-4.md` | `5a298d77bb778d79c17d158657d868ccdadb0ce8a5eef9d01a8d170a22f976ed` |
| `reviews/round-5-response.md` | `cf517de83a7ed13a5f5fe4ee6cbf63170e5729e5d0e29210ef5437cbab53611c` |
| `reviews/round-5.md` | `9530480f20aa59b2c4c72047dfab46c263d73dd86e716a5a566fee4d54f87038` |
| `validation/calibration-overhead-rev5-final.json` | `52d8f63527643395e0487535b0a23e06fc791357554244f19dadc64e66892f3b` |
| `validation/calibration-overhead-rev5-final2.json` | `5a340ebdfa26f35560e8ccce36f06f437cb1e6ad3e045153101e152b829c73dc` |
| `validation/calibration-overhead-rev5.json` | `b01d5a000bdb8a81b816bdee48f2cd240ed5f0b47a7a9100da0984ede9807073` |
| `validation/calibration-overhead-rev6.json` | `0a299b0014898a25c92e8a111bd68a9d1ddcf3f1edf8d547ca4319cd13c3221d` |
| `validation/calibration-sources-rev5-before-opt.json` | `795283eb79ab1aa480f90ba0de613bcc0620f6b97281b9ece452449bf8ff11eb` |
| `validation/calibration-sources-rev5-before-transition-fix.json` | `5334c47b0260556a1e5a2e19608a202013aa8a4f34ef2ab33a8ac0fde0cf24cd` |
| `validation/compute-baseline-rev4.json` | `44a3b3e8ca82cbb98c3b9a9b8ef04e82053b0184eacf4d4d35f237ba6ffb3497` |
| `validation/compute-benchmark-rev4.json` | `e5f068d7badf2f43afcfbd2ee0c0d0a270efe605329a43a865ae114a7aab1ca8` |
| `validation/compute-benchmark-rev5.json` | `fc086226d92a622f32b289dc6b88a6716fd31573fd98d27dedb200fb9f0ed5df` |
| `validation/compute-benchmark-rev6.json` | `9662eb5e9d9344effed4f3141041cf86856c869c9a5219b0f0311eba32076a1f` |
| `validation/compute-projection-rev4-final.json` | `b422c2a6fc7de6e548fa94cc4f36beec403ea9b169aff0429fb01386c44947dd` |
| `validation/compute-projection-rev4-total.json` | `65752e13309281fc8d1019034428103e5f2a445248bc1c142a958b402713a3c1` |
| `validation/compute-projection-rev4.json` | `4d2ca24912b6c94689138ff626869bd7473747a2dfde86497a59925f6e10dc0b` |
| `validation/compute-projection-rev5-final.json` | `3d6c2ff4c76e96c9c215025e4237744e5d6e53540f10966d3e09df6e603638cc` |
| `validation/compute-projection-rev5-final2.json` | `7ebbbb5da4798aac8acc93c6bb1616042075c77a7f62e912027a1ffb7cf8e121` |
| `validation/compute-projection-rev5.json` | `ee616c45674f9beb82bbc8c3951576cf955e1e3c62a2281089fc0b6ec474ad4b` |
| `validation/compute-projection-rev6.json` | `972b3620d76f8f2d344aadc47e07b7b226c445828a166f264a215cceeb7b9199` |
| `validation/compute_benchmark.py` | `5c2557beff77369a3291f4ea20d83acc00e4faf7fc24f2480d1471c347771e11` |
| `validation/fixed_population.py` | `090e93489f4fbf7487a05bc99ae3f77032cc515d3381b41dd80156d1d7c46d9a` |
| `validation/interval-diagnostic-rev4.json` | `50d41cc86c0342e05110d322b5d452ffcad7570c612efbbaee705ce11cc18a54` |
| `validation/interval-diagnostic-sources-rev4.json` | `d12d80627cecd6e7ee0b291dee939638cf7752c24ae62164bb495bf0ab693722` |
| `validation/interval-diagnostic-spec-rev4.json` | `e5a93a49c14a4345cba2fd6f2d4ddc55a998eaa199d371299d5965443fe54bb6` |
| `validation/interval_diagnostic.go` | `27c79a310e03cafa769c22e90c1a8736de1fd0d32ef0ddab42df6ea28838b63b` |
| `validation/interval_diagnostic.py` | `a4b233374bdfec466bfc5862166497a4bbac48b13668aee2697fcb90e2b7c345` |
| `validation/kernel-sources-rev6.json` | `66193e0fee0da8e8a2e386cac973329c81267a893657314bfcf0f42616579e22` |
| `validation/outer-results.json` | `b851c2c1b6ffacee1957ade0071aa9fd38ede0dce5ebd2e3577a2967923f6352` |
| `validation/outer_simulation.py` | `91d07760b61551810282622c2bda53c34fcfcb72a9ffd1798b1a5ffb8ece23f6` |
| `validation/pilot-summary-rev4.json` | `05da900b22b738a61c36d4f02efc76c722109aa2dcffff40fba04f828c14475d` |
| `validation/pilot-summary-rev5.json` | `c21a5a13ea18e521c7fd8469d93b4547471924dd740c75f1c2368fc2ea8ed17e` |
| `validation/ranking.py` | `9fc05fa5ccb9d0d41ff231cec7bcb14c67653399171ddff7a9448a400f483ed7` |
| `validation/reproduce_reports-initial-rev5.py` | `0b7781db46f78e81a0ac091bf123a3fe1e4045925c09396a6299065119a10ed4` |
| `validation/reproduce_reports-optimized-rev5.py` | `d680ee085bc20525761fdedf924f9d5ac0be12984be9a10c1e3ac4ff5b102887` |
| `validation/reproduce_reports-pilot-rev5.py` | `3ee6476c7d142bc1e7340fa9ba31d668aa42cc92d744c66cc5dc910638cf0ee2` |
| `validation/reproduce_reports.py` | `4942de744a4f9f1537c21d60936c2ff2c3fed781a7c16160300b1f2c4cb8a094` |
| `validation/rev5-reviewed-sources.json` | `fa6820c87e068c16b102e2b0d7647c31c6a8f58eccca8f3d3bb8b437fe2bf89d` |
| `validation/rev6-executed-sources.json` | `fafdcbc291a2521e1bf2b5f90ea43e7046135140c9bbb579c69ad41c4e9e6086` |
| `validation/self_test.py` | `06e1791ff1ebd7ea94059c29e457805bc771549ce639750af6c7e0a2ef45b85b` |
| `validation/strong-effect-OC-rev5-closure.json` | `379f2c28ac5fa79ae738d2f1ad18acf823d389ff717616a0d8e37147a4e582fb` |
| `validation/strong-effect-OC-rev5-final.json` | `6f26bad6dcf9f9d7414757ae3ca73b68990187b3d3b52edadd856cb6613a4375` |
| `validation/strong-effect-OC-rev5-optimized.json` | `19c78ee133a87c3f05a9319e58b7bff59031e886837733edb668c566d945615b` |
| `validation/strong-effect-OC-rev5.json` | `089941397a9036830532ffe52410f3f3ca65a153db4e6cfa3e57ec2960ddbb8b` |
| `validation/strong-effect-OC-rev6.json` | `96cbbeaae60431010a97964a39ced7522d2c80d90eeb81d49ebc69945355b476` |
| `validation/synthetic-results-rev1.json` | `fbc4cc0de2c948574a76216cffd12cb33b488f0fedd3c3519de4ddf0b5a82277` |
| `validation/synthetic-results-rev2.json` | `24549f3ed2fcaecb357bf747ec31058ae115592f8bf63a1319763d2a84faa77d` |
| `validation/synthetic-results-rev4-percentile.json` | `b206c43600f80dab426f303a915695176418680296105144545e332dca04c56d` |
| `validation/synthetic-results-rev4.json` | `9f3ec381a945028c5a246ea723f03ee771815ff3bfdf609c917505e12c66c2f2` |
| `validation/synthetic-results-rev5-closure.json` | `f0ae8b5c7b341d2f5d932c72a416ae8451f99429c973e7079daad1869324c134` |
| `validation/synthetic-results-rev5-final.json` | `f5ea12d6d3df4043610921e98fb42a8ceea47f973ca0bd5311ba388fdeabe124` |
| `validation/synthetic-results-rev5-optimized.json` | `b235853e5a6b185c8908543a82430c0cb569261dfb200434e3d8d27a19c34392` |
| `validation/synthetic-results-rev5.json` | `a7a2c26a3fbf7eedf60216bddea196364022453769c95c06f7029946385b552b` |
| `validation/synthetic-results-rev6.json` | `92b935e4685b597c1fe120c5057a6e0fb1cc2dd81657a43ebae4c02e3f8c6073` |
| `validation/synthetic-results.json` | `e65ec36c7d57a0d15f32bc3d2d7bd5f1f76f4148189ec2676653ef996901a350` |
| `validation/synthetic-sources-rev5-initial.json` | `a77aff4f00d87a5d98dafe3114df227b9de27ffb215c45a82ef6074781aee4c2` |

**Byte-preservation checks against the round-5 table.**
- **Unchanged:**
  - every earlier council round and response;
  - ADR-014 `round-1.md` and `round-2.md`;
  - `rev3-digests.json`, `rev4-digests.json` and `inputs/x11-rev4-digests.json`;
  - `docs/adrs/README.md` and `docs/scout-design.html`;
  - from preparation: `exposure-recipe.json`, `exposure-sources.json`, `frozen-exposure-inputs.json`, `model-free-policy.json`, `pins.json` and `runtime-attestation.json`;
  - every revision-4 and revision-5 validation output, reporter snapshot and source snapshot;
  - `interval_diagnostic.py` and `outer_simulation.py`.
- **Identical pairs:** `round-5.md` = ADR-014 `round-3.md` (`9530480f…`). The submission records `rev5-digests.json` as `ba1d2ae4…`, which equals the manifest digest Main reported in round 5.
- **Changed executables:**
  - `compute_benchmark.py` (`dd05a8ba…` → `5c2557be…`);
  - `fixed_population.py` (`5226ad84…` → `090e9348…`);
  - `interval_diagnostic.go` (`24f2eaa7…` → `27c79a31…`);
  - `ranking.py` (`cad37078…` → `9fc05fa5…`);
  - `reproduce_reports.py` (`e2593bcb…` → `4942de74…`);
  - `self_test.py` (`282ffe41…` → `06e1791f…`).
  
  Each old digest appears as a `sha256` field in `validation/rev5-reviewed-sources.json`. The council did not hash the embedded `content` fields (see §11, step 9).
- **14 new files:**
  - ADR-014 `round-3.md`;
  - `rev5-digests.json`;
  - `round-5.md` and `round-5-response.md`;
  - `model-free-binding-check.json` and `runtime-attestation-rev6.json`;
  - eight revision-6 validation files.
  
  77 + 14 = 91.
- **Cross-references:** the projection's `benchmark_sha256` (`9662eb5e…`), `calibration_overhead_sha256` (`0a299b00…`), `reporter_sha256` (`4942de74…`) and `snapshot_sha256` (`fafdcbc2…`) all equal the manifest entries.

**Verdicts.**
- **X1: 10 issues open (0 blocking, 0 major, 10 minor). No sign-off.**
- **ADR-014: 7 issues open (X1-R6-01, -02, -04, -05, -06, -08 and -09). Each needs a change to ADR-014's table or text, and each is shared with X1. No sign-off.** ADR-014 can be adopted only on the joint zero-issue final round.
- **Joint signed set: 10 distinct issues open.**
- No preparation, inventory or draw, Jev, labelling or analysis stage is authorised (§12).

**Severity scale.** Unchanged from rounds 1–5:
- **blocking:** can produce a wrong or unreachable verdict, or breaks experiment rules 1–7 or REQ-169.
- **major:** misleads the result, leaves a preregistered quantity undefined, or makes the study unrepeatable or likely not testable.
- **minor:** text, consistency or provenance. Minor issues must still be fixed before sign-off (rule 1).
- **Calibration:** a fail-closed default that only makes a gate harder to pass is not a defect unless it makes the gate unreachable for a compliant case.

---

## 1. Recommendation

**Revise to revision 7, closing X1-R6-01 to X1-R6-10, and submit it for X1 round 7 / ADR-014 round 5. Do not sign off. Authorise no stage beyond drafting, synthetic validation and option G model-free engineering on synthetic repositories.**

Revision 6 closes the single round-5 major (X1-R5-02) correctly, and the projection arithmetic reproduces exactly (§8). Seven round-5 items are fully fixed. Two are fixed only in part:
- X1-R5-09 still has a run-together token in the ADR.
- X1-R5-01 is fixed as specified in source, but its identical-classification goal is not guaranteed in the compiled Go helper.

Every open issue is minor. They fall into four groups:
- **Numerical:**
  - Go floating-point fusion (X1-R6-02);
  - a refusal check written as a strippable `assert` (X1-R6-03).
- **Resource accounting:**
  - the descriptive safety grid is charged at an execution cap it can never reach, and it enters a testability fit that P l.171 says it cannot gate (X1-R6-04);
  - no memory envelope (X1-R6-05).
- **Recipe definitions:**
  - dev-sample eligibility (X1-R6-06);
  - the strings fed to the keyed orders (X1-R6-07);
  - the argv positive control (X1-R6-08);
  - the single GUIDE-author session (X1-R6-09).
- **Text and ledger:** X1-R6-01 and X1-R6-10.

Keep all of the following:
- the two hard per-population counters;
- cheap gate costing kept separate from the expensive trials;
- the worst-case enumeration charge;
- the L400 RSS record;
- the explicit pinned 3.11.11 interpreter and its digests;
- the disclosure of lost revision-4 sources and the snapshot-before-overwrite rule;
- the windowed keyed dev sample;
- the five-action preparation scope and its clock start;
- the falsifiable binding recipe;
- the stage names.

## 2. Alternatives considered

| Option | Description | Tradeoff | Council position |
|---|---|---|---|
| A. Sign off revision 6 and carry the minors into the next authority review | Starts the clock sooner | Rule 1 requires zero issues. X1-R6-04 and X1-R6-06 sit in binding stop rules or binding inputs. | Rejected |
| B. **Revision 7 closes the ten minors; then a zero-issue round** | All fixes are text, recipe or small synthetic changes | One more round | **Recommended** |
| C. Authorise a scoped preparation now | Starts fetch and attestation | Issues are open (rule 1). The binding recipe lacks a positive control (X1-R6-08), and the identifier-review scope for dev is undefined (X1-R6-06). | Rejected |
| D. For X1-R6-04, keep the total-grid fit as the testability gate and amend P l.171 so that safety resources can gate | Smaller text change | A descriptive grid that cannot change the verdict would end the study as not testable | Rejected in favour of a binding-only testability fit |
| E. For X1-R6-02, accept Go/Python differences and make Python the sole authority for zero-variance classification | No Go change | The Go helper produces the OC trials, so its classification must match the evaluator's | Rejected. Explicit `float64()` rounding is cheap and guaranteed by the language. |
| F. For X1-R6-09, keep one GUIDE session and handle review issues by protocol amendment | No recipe change | Every GUIDE issue costs an amendment round inside the 45-day clock | Rejected. Register a bounded set of fresh sessions instead. |
| G. Finish model-free engineering before preparation: an OC controller that enforces both counters, a policy-binding harness on synthetic repositories, the final-evaluator draft | Shortens the clocked path | Work before the clock starts, none of it touching real data | **Recommended practice**, not a sign-off condition |

## 3. Rule checks

- **Rules 1–2.** This round is recorded with digests. Issues are open, so iteration continues.
- **Rule 3.** No predictor–outcome association was computed on real data. All revision-6 evidence is synthetic, and the historical method selection was not rerun (manifest `historical_method_selection_rerun: false`).
- **Rule 4.**
  - Planted effect: 2,000/2,000 support, Wilson MC95 lower .9980829527187469, realised gain .463, zero inadequate (`strong-effect-OC-rev6.json`).
  - Planted null: not supported (independent run reported by Main).
  - New boundary fixtures for the zero-SE tolerance.
  - Nine expanded calibration-equivalence profiles.
  - The fusion gap (X1-R6-02) and the strippable refusal (X1-R6-03) weaken two of these checks.
- **Rule 5.** Nothing is signed yet. Superseded projections are preserved and labelled. Revision-5 executables were archived before modification.
- **Rule 6.** The envelope, RSS, both counters and the failing L2,000 example are reported (README). Reproduce commands name the pinned interpreter explicitly.
- **Model use and secrets.**
  - No model was called and no session was created.
  - The binding recipe checks the key by presence only and never runs a model to test the interlock.
- **Toolchain rule.**
  - The numerical evidence ran under `.tools/venv/bin/python` 3.11.11, SHA-256 `570b4105…`. It is recorded in all five revision-6 outputs.
  - `toolchain/versions.env` is unchanged.
- **Design rules.**
  - ADR-014 has a sequential ID and an adopted-form status, unchanged and matching ADR-011.
  - Its Review section defers to the final round in its folder, and its Records list now includes `round-3.md`.
  - The index and STATUS are correct.
  - Seven issues require table or text changes.
- **Ledgers.**
  - R-085 is present (`docs/research-ledger.md:84`).
  - Next free ID is R-086 (`docs/STATUS.md:128`).
  - HD9 remains proposed, but its sources column omits R-085 (X1-R6-10).

## 4. Verification of round-5 fixes

| Issue | Status | Evidence |
|---|---|---|
| X1-R5-01 | **Fixed as specified.** Compiled-arithmetic residual in X1-R6-02 | Tolerance `abs(value-point) <= 1e-14` at `validation/ranking.py:192` and in both Go paths (`interval_diagnostic.go:21, 57`). Same two-pass centred variance in source (`ranking.py:167–179`; `interval_diagnostic.go:11`). Fixtures at \|Δ\| ∈ {0, 5e−15, 1e−14, 2e−14}, both signs, on the complete path, Python vertex path and Go complete/vertex paths, with independent expected intervals (`self_test.py:216–251`). Method selection not rerun. |
| X1-R5-02 (major) | **Fixed** | Attempts ≤ ceil(4,000/p̂_gate) and gate-passing executions ≤ 4,000, including inadequacy. Rejection does no audit or bootstrap. Either exhaustion fails the cell (P l.163, l.171, l.181; M l.40–42, l.126–129). The reporter charges 4,000 × maximum trial CPU, plus gate CPU × attempt cap (`reproduce_reports.py:222–229`). `-final2` is preserved and labelled superseded. ADR rows (l.30, l.45) and R-081 are updated. The 4,000-execution count is right for binding cells; its use for safety cells is in X1-R6-04. **Scope:** the cap is registered and costed, but no executing OC controller exists yet (§6). |
| X1-R5-03 | **Fixed.** Refusal mechanism in X1-R6-03 | Determinism renamed (P l.169; `self_test.py:336–341`). Expanded profiles at targets 0/.10/.20 cover binding caps 400/850, sizes ≥ 500, failure .02, overlap .5, prevalence .10/.55, size factors .05/.20, gradient, scanner controls, heterogeneity .10 and the reduced-L benchmark profile (`self_test.py:359–390`). Returned-gain equality at `fixed_population.py:192`. Claims corrected in P, M, README, ADR l.33 and the response. |
| X1-R5-04 | **Fixed.** Memory-envelope residual in X1-R6-05 | Worst case over every size 20–2,000, p = .55, size factor .05: 530,767 candidates, 6.546945999999999 CPU s, charged 11,900 times = 43.282587444444445 h (`calibration-overhead-rev6.json`; projection `candidate_enumeration`). Envelope published (projection l.8–18; README; P l.179 refers to it). L400 RSS and an 8-worker footprint of 6,639,583,232 B recorded. |
| X1-R5-05 | **Fixed** | Explicit `.tools/venv/bin/python` in every README command (`README.md:115–119`). Interpreter version and digest recorded in synthetic (`:815–817`), strong (`:585–587`), benchmark (`:369442–369444`), calibration (`:70–72`) and projection (`:129–130`). Framing corrected (`preparation/runtime-attestation-rev6.json`). R-085 recorded. No pin change. |
| X1-R5-06 | **Fixed** (disclosure alternative) | Lost revision-4 digests `405f2a4a…`, `3c47f232…` and `dabebd88…` disclosed in the README, manifest `source_loss` and ADR l.56. The earlier diagnostic snapshot is not misattributed. `rev5-reviewed-sources.json` lists all 11 revision-5 executables with the round-5 digests, including the six changed in revision 6. That recovery was impossible is author-reported. |
| X1-R5-07 | **Fixed against the required text.** Eligibility gap in X1-R6-06 | Window, exclusive C0_dev ("or the root if shorter (exclude the root)"), HMAC key and message, 150 per repository, ≤ 5 per commit, ≥ 30 commits, insufficiency stop, no extension (P l.190; `dev-census.json:24–27`). Within-commit dependence in the threats (P l.236). Renamed in P, ADR and the preparation README. Dev scans counted in inventory (P l.206, l.219). |
| X1-R5-08 | **Fixed.** Residuals in X1-R6-08 and X1-R6-09 | (a) Five actions and clock start (P l.218).<br>(b) Synthetic plus pinned-pair binding check covering both policies, key presence, argv log and issue-ID parity (P l.219; `model-free-binding-check.json`).<br>(c) Dev collection named as inventory; GUIDE session named as stage 3, with zero-issue review before dev labels (P l.219–220; `guide-author-isolation.json:5`). |
| X1-R5-09 | **Partly fixed.** Residual in X1-R6-01 | The cited tokens are corrected, but ADR l.48 still reads "no-cap M2,000/failure0/fully annotated". |

## 5. New issues

### X1-R6-01: minor. Run-together quantities remain in the signed ADR (X1-R5-09 residual)

- **Location:** `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md:48`, row "What product parity/recovery?".
- **Problem:** the row reads "no-cap M2,000/failure0/fully annotated". This is mutable normative prose, so the round-5 fix ("add the missing spaces in every signed document") is incomplete.
- **Required fix:**
  - Replace it with "M = 2,000 with no binding cap, failure 0, fully annotated", matching P l.185.
  - Re-grep every mutable signed document for digits joined to words.
  - Generated outputs, review records and snapshots stay unchanged.

### X1-R6-02: minor. The Go helper may fuse multiply–add, so Python and Go can still classify zero variance differently on darwin/arm64

- **Location:** `validation/interval_diagnostic.go`:
  - comment at l.9–10 ("identical multiplication grouping");
  - `fit` at l.11 (`e:=a[1]-a[2]-d*a[0]`, `squares+=z*z`);
  - `endpoint` at l.30 (`point-pivot*math.Sqrt(variance)`);
  - the moment loop at l.23.
  
  Also the fixtures at `self_test.py:220`, M `zero_variance_rule`, and ADR l.26 ("identical two-pass variance … in Python/Go").
- **Problem:**
  - **The language allows fusion.** The Go specification permits an implementation to fuse `x*y ± z` into one rounding. Only an explicit `float64(...)` conversion prevents it.
  - **The arm64 backend does fuse.** Its rewrite rules turn `a − x*y` into `FMSUBD` and `a + x*y` into `FMADDD` (ARM64.rules l.1796–1806, cited above). The helper is built without flags for `go1.26.10 darwin/arm64` (`synthetic-results-rev6.json:814`).
  - **Example.** Take two runs with T = 10 and T = 30, J−B = 2 and 6, so d = fl(0.2) [INFERENCE: the compiler fuses this expression].
    - Python rounds fl(0.2)×30 to 6.0, so both residuals are 0 and the variance is exactly 0.
    - Fused Go gives residuals of about −1.1e−16 and −3.3e−16, so the variance is greater than 0.
  - **The branches diverge.** In that case Python takes the zero-variance branch (pivot 0 or a signed infinity), while Go computes a finite pivot. In a mixed resample with \|D*−D̂\| ≤ 1e−14 but D* ≠ D̂, Go can produce a large finite pivot where Python produces 0.
  - **Why it matters.** This is the degenerate regime R5-01 targeted. P l.179's parity requirement is ≤ 1e−12, and a classification flip is a discontinuity that tolerance cannot absorb.
  - **Why the fixtures miss it.** Every tolerance fixture uses T = 1, where d·T is exact.
- **Required fix** (recommended option):
  - Prevent fusion explicitly in every Go product that feeds an addition or subtraction: `float64(d*a[0])`, `float64(z*z)`, `float64(pivot*math.Sqrt(variance))`, and the `e*e` and `e*e*e` moment terms.
  - Add a differential fixture with heterogeneous T and a constant inexact ratio (for example T ∈ {10, 30}, J−B ∈ {2, 6}, plus a mixed resample). Require the same zero/non-zero variance classification and byte-identical intervals in both kernels, on the complete and vertex paths.
  - Correct the comment and ADR l.26 to say "identical after explicit rounding".
  - Record a disassembly of `main.fit` showing no `FMADDD`/`FMSUBD/FNMSUBD` in the next submission.

### X1-R6-03: minor. The calibration refusal is a bare `assert`, and no fixture shows it firing

- **Location:** `validation/fixed_population.py:192`; P l.169 ("otherwise refuse it"); ADR l.33 and l.55.
- **Problem:**
  - Python removes `assert` statements under `-O` or a non-empty `PYTHONOPTIMIZE`, so the refusal P requires depends on interpreter flags.
  - No negative fixture shows a mismatch being refused.
  - A guard that has never been seen to fail is weak rule-4 evidence.
- **Required fix:**
  - Replace it with `if _population_gain(...) != realised: raise RuntimeError(...)`.
  - Add a mutant fixture (monkeypatched `_population_gain` returning a mismatched value) that must raise.
  - Record `sys.flags.optimize` in output provenance and refuse a non-zero value.

### X1-R6-04: minor. The descriptive safety grid is charged at an execution cap it cannot reach, and it enters the testability fit that P says it cannot gate

- **Location:**
  - P l.171 ("10,000 gate-passing trials/profile … cannot gate dev/test or alter the verdict");
  - P l.179 ("failed or unmeasured runtime/memory fit means not testable") and l.181;
  - M l.126–129 (`gate_passing_trials_per_population: 2000`);
  - `reproduce_reports.py:222–229` (all four trial classes × 4,000), `:232` and `:290` (`fits_caps` on the total);
  - README envelope paragraph ("descriptive safety … remains counted"; L2,000 "fails");
  - ADR l.40 and l.45.
- **Problem:**
  - **(a) Overcharge.** A safety population stops at 2,000 gate-passing executions, and inadequate executions count toward that 2,000 because they stay in the denominator. Its 4,000-execution cap can therefore never bind. The projection still charges 4,000:
    - 3.941924 + 875.9159827 = 879.8580 h are charged;
    - the reachable maximum is 439.9290 h;
    - the overstatement is 439.93 h [INFERENCE: P l.171 and M l.126–129 as written].
  - **(b) Contradiction.** P l.171 says the safety grid cannot gate dev or test and is reported unmeasured if it is not finished by gold. Yet the published envelope and `fits_caps` test the total (binding plus safety) against 2,880 h, and the README treats a total-fit failure as not testable. P l.206 already uses binding-only wall days for the throughput screen.
  - **Effect** [INFERENCE: linear per-unit scaling].
    - At p = .5 the total fit fails above 5,833,787 units per population (L ≈ 1,944.6 at mean 1,000).
    - With safety charged correctly, the total fit holds to about 7.43 M units.
    - A binding-only fit holds to about 12.65 M units, which exceeds the largest registered population of 3 × 2,000 × 2,000 = 12 M.
    - So compliant inventories between about 5.8 M and 12 M units would be declared not testable after the clock starts, purely because of descriptive work.
  - **Why minor.** Such populations are far above the 1.2 M planning point and probably implausible for introduced-issue runs [INFERENCE]. There is no effect on the verdict.
- **Required fix** (recommended option):
  - **(i) Charge safety correctly.** Charge safety trials at 2,000 gate-passing executions per population. State that the 4,000 cap is inert for safety and remains only as a guard.
  - **(ii) Make the testability fit binding-only.** Define the fit as the binding critical path: binding construction, binding calibration and enumeration, requested-base evaluation, binding trials and binding gate checks, all × 2, must fit 2,880 core-hours and 15 OC days, together with the locked memory fit.
  - **(iii) Keep safety descriptive.** Safety runs in the remaining capacity and is reported unmeasured if not finished by gold, as P l.171 already says.
  - **(iv) Publish both envelopes.** The binding envelope is the stop rule; the total envelope is for planning safety completion.
  - **Where to change it:** P l.171/179/181, M, README, the reporter (a create-once revision-7 projection, keeping revision 6 as superseded), and ADR l.40 and l.45.

### X1-R6-05: minor. The published envelope covers CPU only; memory scaling and host memory are missing

- **Location:**
  - `compute-projection-rev6.json:69–77` (`memory`);
  - `calibration-overhead-rev6.json:16, 49` (L40: 110,329,856 B at 120,000 units; L400: 829,947,904 B at 1,200,000 units);
  - the README "Peak RSS" paragraph;
  - M resources (cores recorded, memory not);
  - `docs/hypothesis-ledger.md:40` ("memory envelope");
  - ADR l.45.
- **Problem** [INFERENCE: linear in units]:
  - The calibration RSS slope is 719,618,048 / 1,080,000 ≈ 666.3 B per unit, with an intercept of about 30.4 MB.
  - At 6 M units that is about 4.03 GB per worker, or about 32.2 GB for 8 workers. At 12 M units it is about 64 GB.
  - The machine's physical memory is not recorded, so readers cannot tell whether memory, not CPU, binds first.
  - The "conservative" 6.64 GB holds at L400 only.
  - The HD9 history line claims a "memory envelope" that was not published.
  - The locked fit still fails closed, so the verdict is unaffected.
- **Required fix:**
  - Record the OC host's physical memory as a value only, from a system query, in the benchmark output and M.
  - Publish the RSS slope and intercept, and the 8-worker memory break-even U, next to the CPU envelope in P §8, the README and the ADR compute row.
  - Reword the ledger line to "L400 RSS projection and memory envelope", once the envelope exists.

### X1-R6-06: minor. Dev-sample eligibility is not uniquely defined

- **Location:** P l.23, l.39–41, l.190, l.192, l.236; `preparation/dev-census.json:25–26`; `preparation/exposure-recipe.json:11–12`; ADR l.22.
- **Problem:**
  - **(a) Exposure rules contradict each other.**
    - P l.190 applies "§3 first-occurrence/exclusion rules". §3 counts exposure exclusions, and §2's freshness rules exclude examined issue IDs, conservative commit and blob matches, and unknown freshness. The recipe also counts "exposure … exclusions".
    - The same P line then says "Examined re-labels require disclosure", which implies examined units may be sampled.
    - The dev tips `d666ccb6` and `562fa206` are the examined E2/D1 snapshots, so the whole windows are examined material by construction (P §2).
    - Under the strict reading, conservative matching plus unknown-freshness ineligibility could exhaust supply and trigger the study-ending insufficiency stop [INFERENCE].
  - **(b) Run-size eligibility is unstated.** P does not say whether the 20 ≤ N_run ≤ 2,000 eligibility rule (l.41) applies to dev commits.
  - **(c) Composition is undisclosed.**
    - The ≤ 5-per-commit cap gives units in large commits a lower inclusion probability, so the unweighted dev estimates (l.192) describe a commit-balanced mixture.
    - The dev windows end in 2023-03, before the 2024-onward test windows.
    - The greedy "skip a unit once its commit has five" step is implied but not stated.
  - **Why it matters.** The dev eligible set feeds binding quantities: c_U → M̂, r_L futility and the throughput bound.
- **Required fix** (recommended option):
  - **(a)** Dev applies first-occurrence, repeated, pre-existing-at-C0_dev and unknown-attribution exclusions only. It does **not** apply the test freshness and exposure exclusions, because §2 already declares fastapi and axum examined and development-only. Examined units stay eligible, are flagged and counted, and are labelled under the current guide; legacy labels never count. State that preparation action (d) covers the test repositories only.
  - **(b)** Run-size eligibility does not apply to dev; dev quantities are per unit.
  - **(c)** State the greedy per-commit skip rule. Add the unequal per-commit inclusion and the 2023-versus-2024+ era gap to §12. Report the size distribution of dev source commits against the test frame.
  - Update P §9, the recipe and ADR l.22.

### X1-R6-07: minor. The repository and run strings fed to the keyed orders are undefined

- **Location:**
  - P l.41 (`169-x1-runs`: "repository NUL full commit SHA");
  - P l.60 (`169-x1-audit`: "repository NUL run ID NUL issue ID");
  - P l.190 (`169-x1-dev`: "repository NUL full commit SHA NUL issue ID");
  - `preparation/dev-census.json:26`;
  - M seed and recipe strings.
- **Problem:**
  - "Repository" could be `django`, `django/django` or the URL; "run ID" and the SHA's letter case are also unstated.
  - Each choice yields a different SRS, audit draw and dev sample. This leaves a forking path in preregistered randomisation, including on the supply gates, which P l.41 says must never be redrawn to improve supply.
  - The synthetic `select_runs` (`ranking.py:314–316`) takes the string as given.
- **Required fix:**
  - Define:
    - repository = exact GitHub `owner/name` as shown in the §3 table and the dev recipe (`django/django`, `microsoft/TypeScript`, `prometheus/prometheus`, `fastapi/fastapi`, `tokio-rs/axum`), UTF-8 and case-preserved;
    - commit SHA and run ID = the full 40-character lowercase hexadecimal commit SHA;
    - issue ID = the exact UTF-8 Scout issue ID.
  - Apply this in P, M and the recipe.
  - Add one fixture per key with an independently computed expected digest for a fixed message.

### X1-R6-08: minor. The model-free binding check has no positive control for its process interlock

- **Location:** `preparation/model-free-binding-check.json:14` ("Any attempted or actual claude/codex start fails acceptance"); P l.219; ADR l.61.
- **Problem:**
  - The policy check has built-in sensitivity: the synthetic repository enables model settings before the binding (`:6`).
  - The argv check does not. A broken or empty process capture passes without detecting anything, so the check is not demonstrably falsifiable, which is what X1-R5-08(b) asked for.
- **Required fix:**
  - On the synthetic fixture only, with the key absent and the binding deliberately **not** applied, put logging stub `claude` and `codex` executables first on PATH. If the unbound policy names CLI paths, point it at the stubs. Each stub records its argv and exits non-zero.
  - Require the capture to record the attempted start and the interlock to report failure. Then require the bound run to record none.
  - No real CLI or model runs.
  - Record the unbound resolved policy, showing the model settings enabled.
  - Update P l.219, the recipe and ADR l.61.

### X1-R6-09: minor. Exactly one GUIDE-author session, with no path to fix the GUIDE after review

- **Location:** P l.220; `preparation/guide-author-isolation.json:5` and `future_source_only_protocol` (the "exactly one isolated author session" and "obtain a different author" entries); ADR l.60.
- **Problem:**
  - The GUIDE must pass a zero-issue review before any dev label. Any review issue needs a revised GUIDE, but only one session is authorised, and the recipe's "obtain a different author" contradicts "exactly one".
  - The only remedy is a protocol-amendment round inside the 45-day clock [INFERENCE: a first-pass zero-issue review is unlikely; X1 rounds 1–5 each found at least nine issues].
  - The wording follows the council's own X1-R5-08(c), which this round corrects.
- **Required fix** (recommended option):
  - Register **at most three** fresh isolated author sessions under stage-3 authority. Each receives:
    - the reviewed packet;
    - the previous GUIDE draft;
    - the council's issue text verbatim, with no empirical data.
  - Log every session, all before any dev label.
  - Replacing an author because of exposure uses up one of the three.
  - Exhausting all three means not testable before dev.
  - Update P l.220, the recipe and ADR l.60.

### X1-R6-10: minor. R-085 is missing from HD9's sources

- **Location:** `docs/hypothesis-ledger.md:36`, which lists R-078, R-080 and R-081 but not R-085; `docs/research-ledger.md:84` says R-085 is used for "HD9 / X1".
- **Required fix:** add R-085 to HD9's sources column, and refresh the `HD9` owned-extract digest in the next manifest.

## 6. Assessment of the targeted questions

- **Trial-execution cap.**
  - The protocol, manifest and projection are correct for binding cells. No executing full-OC controller enforces the two counters yet: the Go population path loops `for range q.Trials` (`interval_diagnostic.go:72–75`).
  - That is the declared boundary of numerical validation (P l.177, l.220; manifest `full_product_rule4_harness`), so it is **not** an issue at this stage.
  - It is a stage-3 prerequisite (§12): one shared controller (attempt → size gate → execution counter → audit and bootstrap → eligibility), or a reviewed outer coordinator, with fixtures showing:
    - zero audit or bootstrap calls on rejected attempts;
    - inadequacy consuming executions;
    - exhaustion at either counter;
    - success on the final permitted execution;
    - distinct stopping counts for eligible cells and for safety cells.
- **Projection of 1,602.2 h / 8.34 days.** It reproduces exactly (§8). The binding path is 621.3 h. After X1-R6-04 the indicative total is about 1,162.3 h [INFERENCE].
- **Calibration.** It is exact as implemented and the evidence is now broad. The 21-call budget (20 search calls plus the gain check) is charged correctly (`reproduce_reports.py:109–114`).
- **Interpreter.** Closed. The benchmark refuses any version other than 3.11.11 (`compute_benchmark.py:231–232`).
- **Preparation contract.** It is well scoped. Two definitional gaps remain (X1-R6-06, X1-R6-08) and one deadlock risk (X1-R6-09).
- **ADR-014.**
  - The status matches adopted ADR-011, and the Records list is updated.
  - The table matches P on the counters, enumeration, interpreter, preparation actions and dev window.
  - The design card is unchanged and still accurate.
  - Seven issues need table or text edits.

## 7. Checked and sound

- **Cell counts** agree across P, M, the projection and the ADR:
  - 17,920 complete/reference cells and 252 missing;
  - 336 safety profiles, 8,064 outcomes and 2,016 interval groups (336 complete + 1,680 full-vertex);
  - 11,900 calibrated populations;
  - 90,860 constructions = (17,920 + 252) × 5;
  - 80,640 requested and 40,320 canonical evaluations.
- **Binding allocation:** (41.62343655555556 + 43.282587444444445) × 10,220/11,900 = 72.9193 h. Adding binding construction, requested-base evaluation, trials and gate checks gives 621.3154967724441 h.
- **Gate-check cost does not grow with L** (`compute_benchmark.py:190–191`).
- **Throughput constant:** 7 × 900 × ln 40 / (3 × 299) = 25.909 s; × 1.2 ≈ 31.09 s per reserved case.
- **Snapshot discipline:**
  - all six changed executables are listed in the pre-change archive with round-5 digests;
  - every revision-6 output's source digests equal the current files and `rev6-executed-sources.json`.
- **Lost revision-4 sources:** disclosed without reconstruction.
- **Dev window edge case:** a chain of exactly 2,000 commits has no 2,000th ancestor, so C0_dev is the root, which is well defined.
- **No new forking path from the execution cap.** Adequacy failure near 50% is implausible at sampled N ≈ 20,000 per repository [INFERENCE].

## 8. Evidence index

- **Total** (`compute-projection-rev6.json:87–112`):
  - construction 25.0419246 + 0.4630248;
  - enumeration 43.2825874;
  - reevaluation 87.9265856 + 175.8531712 + 41.6234366;
  - gate checks 5.7962559 + 0.0815098 + 0.1086798 + 0.5433990;
  - trials 210.2359467 + 131.3873974 + 3.9419240 + 875.9159827;
  - sum = 1,602.2018254 h; ÷ 192 = 8.3448 days.
- **Trial terms:**
  - binding complete: 17,920 × 5 × 4,000 × 0.0010558725 × 2 / 3,600 = 210.236 h;
  - missing: 252 × 5 × 4,000 × 0.0469240705 × 2 / 3,600 = 131.387 h;
  - safety full-512: 1,680 × 5 × 4,000 × 0.0469240705 × 2 / 3,600 = 875.916 h.
- **Envelope:**
  - fixed hours = 1,221.4812 + 43.2826 = 1,264.7638;
  - per-unit hours = 0.4963612 s × 2 / 3,600 = 2.7575679e−4;
  - U_max at p = .5: (2,880 − 1,264.7638 − 6.5298) / 2.7575679e−4 = 5,833,787;
  - maximum attempt cap at 1.2 M units is 1,573,486, so p_min = .0025421;
  - with 100,000 draws the smallest passing estimate is .00255 (challenger);
  - L2,000 at mean 1,000: 2,925.834 h.
- **X1-R6-04** [INFERENCE]:
  - safety overcharge = (3.941924 + 875.9159827) / 2 = 439.9290 h;
  - corrected total = 1,162.2728 h;
  - corrected total-fit U_max = (2,880 − 824.8348 − 6.5298) / 2.7575679e−4 ≈ 7.43 M;
  - L2,000 corrected = 2,485.905 h;
  - binding per-unit hours = (90,860 × 4.13415e−7 + 80,640 × 3.2710783e−6 + 10,220 × 5.2466517e−6) × 2 / 3,600 = 1.97202e−4;
  - binding fixed hours = 341.6233 + 37.1720 + 5.8778 = 384.673;
  - binding U_max = (2,880 − 384.673) / 1.97202e−4 ≈ 12.65 M.
- **X1-R6-05** [INFERENCE]:
  - RSS slope = (829,947,904 − 110,329,856) / 1,080,000 = 666.3 B per unit; intercept ≈ 30.4 MB;
  - at 6 M units: 4.03 GB per worker, 32.2 GB for 8 workers.
- **X1-R6-02:**
  - the Go specification on floating-point fusion and explicit conversion;
  - ARM64.rules l.1796–1806 (`FADDD a (FMULD x y)` → `FMADDD`; `FSUBD a (FMULD x y)` → `FMSUBD`);
  - `interval_diagnostic.go:11, 23, 30`;
  - fixtures with T = 1 at `self_test.py:220`.

## 9. Risks

- **Locked-size fit.** CPU, memory or 8-worker efficiency can still fail legitimately after the clock starts. X1-R6-04 and X1-R6-05 remove the artefactual part and expose the memory part before preparation.
- **Host capacity is unknown.** Only 1 and 2 workers were measured. The host's physical memory is unrecorded (X1-R6-05).
- **OC controller not yet written.** Cap enforcement rests on a future reviewed artefact (§6).
- **Implementing the binding** (preparation action (c)) may touch product code. If so, its digests must join the seals and the pinned Scout identity reviewed at inventory (peer question). Consider using the Django C0 and its parent rather than the tip as the pinned fixture pair, so preparation never scans a frame commit. P l.219 already permits the tip as fixture-only, so this is advice, not an issue.
- **Shared-file staleness.** Whole-file digests of STATUS, both ledgers and the index go stale if a sibling lands first. The owned extracts in the manifest mitigate this.
- **Clock.** Every remaining fix is pre-clock synthetic or text work. Option G reduces work inside the clock.
- **Reversibility.** All fixes are pre-data. No product, default, accepted ADR or requirement changes.

## 10. Where the peers changed the council's recommendation

- **Experiment peer:**
  - Found the contradiction in dev exposure and eligibility (**X1-R6-06**) and the undefined keyed-order strings (**X1-R6-07**). The lead had not found either.
  - Proposed the argv positive control (**X1-R6-08**), the bounded set of GUIDE sessions (**X1-R6-09**), the RSS slope and host memory (**X1-R6-05**), and the R-085 ledger gap (**X1-R6-10**). All adopted.
  - Independently raised Go fusion with a stronger heterogeneous-T example, which was adopted into **X1-R6-02**.
  - Raised the bare `assert`. The lead had treated it as advice for the final generator, and the peer's point that no fixture shows it firing made it **X1-R6-03**.
  - Rated R5-09 fixed. The council overruled this on the challenger's evidence.
  - Its point that the binding-check JSON says "action 3 … after its scoped review" and disagrees with P was **not adopted**: it is only stricter (calibration rule). Its open questions are recorded in §9.
- **Challenger:**
  - Found the residual `M2,000/failure0` token (**X1-R6-01**), so X1-R5-09 is partly fixed.
  - Independently reproduced the projection and envelope, including the .00255 quantisation.
  - Flagged that no executing controller enforces the counters. This is recorded as a stage-3 prerequisite (§6, §12), not an issue, because the boundary is declared.
  - Rated X1-R5-01 fixed without qualification. The lead and the peer added X1-R6-02.
- **Lead only:**
  - The safety overcharge and the conflict between a descriptive grid and the testability fit (**X1-R6-04**), with the binding-only envelope arithmetic.
  - The initial Go-fusion finding and the citation of the arm64 rules (X1-R6-02).

## 11. Execution outline for revision 7 (one integration owner)

1. **Kernel parity (X1-R6-02).**
   - Add explicit `float64()` rounding in the Go helper.
   - Add the heterogeneous-T constant-ratio differential on the complete and vertex paths.
   - Record a `main.fit` disassembly.
   - *Test:* identical classification and byte-identical intervals.
2. **Calibration refusal (X1-R6-03).**
   - Raise explicitly instead of asserting; add the mutant fixture; record the `optimize` flag.
   - *Test:* the mutant raises.
3. **Resource accounting (X1-R6-04, X1-R6-05).**
   - Charge safety at 2,000 executions; make the testability fit binding-only.
   - Add host memory, the RSS slope and intercept, and the memory break-even.
   - Produce a create-once revision-7 projection, keeping revision 6 as superseded.
   - *Test:* both envelopes are published, and the binding envelope covers U ≤ 12 M at p = .5 [INFERENCE until recomputed].
4. **Dev and keyed-order definitions (X1-R6-06, X1-R6-07).**
   - Set the dev eligibility rules, state the greedy rule, add the era and inclusion threats and the size report.
   - Define the canonical strings and add digest fixtures.
5. **Preparation recipes (X1-R6-08, X1-R6-09).**
   - Add the stub-CLI positive control.
   - Allow at most three fresh GUIDE sessions.
6. **Text and ledgers (X1-R6-01, X1-R6-10).**
   - Respace the ADR token; add R-085 to HD9; refresh the extracts.
   - Update ADR-014 rows l.22, 26, 40, 45, 48, 60 and 61.
7. **Re-run once** under `.tools/venv/bin/python` 3.11.11 with create-once outputs. Snapshot every executable before modification.
8. **Submit** `rev7-digests.json` and the response for X1 round 7, to be saved also as ADR-014 `round-5.md`.
9. **Main checks before round 7:**
   - recompute `content → sha256` for every entry of `rev5-reviewed-sources.json`, `kernel-sources-rev6.json`, `rev6-executed-sources.json` and the revision-7 snapshots;
   - verify the disassembly from step 1.

## 12. Stage authorisation

| Stage | Status now | Earliest path |
|---|---|---|
| Drafting, synthetic validation, and option G model-free engineering on synthetic repositories | **Permitted** (unchanged) | — |
| Preparation | **Not authorised**: 10 open issues (rule 1) | A zero-issue round on revision 7; then only P l.218 actions (a)–(e) |
| Inventory, model-free scans and SRS draw (including dev-sample collection) | **Not authorised** | Sealed preparation outputs, a passing binding check with positive control, runtime closure, the test-repository identifier review, and a scoped zero-issue inventory review |
| Full OC and implementation (including GUIDE-author sessions) | **Not authorised** | Locked sizes; a binding-only CPU and memory fit; **a reviewed OC controller enforcing both counters**; a reviewed final evaluator; product-backed rule-4 frames; GUIDE and isolation reviewed to zero issues before any dev label |
| Dev Jev and labels; test scoring and labels | **Not authorised** | Under full authority only |
| Gold unblinding and analysis | **Not authorised** | — |
| ADR-014 adoption | **Not authorised** | The joint zero-issue final round saved in its folder |
| Default, threshold or product-sampler change | **Not authorised** | A separate signed ADR |

**Next stage authorised:** none beyond drafting, synthetic validation and option G engineering on synthetic repositories.

**Sign-off statement:** none. Round 6 closes with:
- **X1: 10 issues open**:
  - **0 blocking**;
  - **0 major**;
  - **10 minor:** X1-R6-01 to X1-R6-10.
- **ADR-014: 7 issues open** (X1-R6-01, -02, -04, -05, -06, -08 and -09, each needing ADR-014 table or text changes). Adoption also waits for the joint zero-issue X1 round.

The reviewed files are those in `reviews/rev6-digests.json` (manifest SHA-256 `cca1baf432389b094588b6085888e8116442f0ee9baef15cbc750ef30086a746`, as reported by Main), with the per-file digests in the table above. No preparation, inventory, draw, Jev, labelling or analysis stage is authorised, and ADR-014 is not adopted.
