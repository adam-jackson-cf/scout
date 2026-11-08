# X1: council experiment review, round 8 (revision 8 with ADR-014 round 6)

Date: 2026-11-08. Reviewer: innovation council. The lead wrote the synthesis. Three read-only peer passes ran:
- `innovation-challenger` (X1Council8.X1Council8Challenger): an independent audit;
- `experiment-peer` (X1Council8.X1Council8ExperimentPeer): an independent audit;
- `innovation-challenger` (X1Council8.X1Council8Rebuttal): an adversarial check of the experiment peer's candidate findings.

This record is X1 round 8 and ADR-014's sixth council review. Save it byte for byte as both `evaluation/x1-triage-ranking/reviews/round-8.md` and `docs/adrs/2026-11-08-x1-delegated-decisions/round-6.md`.

**Governance.**
- `CLAUDE.md` experiment rules 1–7.
- Design rules 1–3, for ADR-014.
- The ledger rules.
- The General sections "Toolchain", "Jev labelling", "Model use" and "Secrets".
- Standing owner delegation R-064: where an owner choice comes up, this record gives the council's recommended option as the fix.
- Wording preferences alone are not counted as issues.
- Risks already accepted by name in earlier rounds are not reopened without new evidence.

**Scope.** Revision 8, as listed in `reviews/rev8-digests.json`. That manifest has 123 files and excludes itself. Each round-7 item was checked against the fix required in `reviews/round-7.md` §5.

**What was and was not run.**
- This round ran no shell, model call, self-test, simulation, benchmark or digest recomputation, and made no web lookup.
- The evidence comes from reading files, grep and hand arithmetic.
- `codegraph` was not used. Its reference documents (`readiness-and-bootstrap.md`, `query-selection.md`) were read first, but this read-only session had no way to run the executable.
- Main reports independent checks by a separate agent. This round relies on that report:
  - Interpreter: pinned CPython 3.11.11, executable SHA-256 `570b4105…1d6d`.
  - README self-test: exit 0, 249 checks passed, 12.0 s. Planted effect supported, planted null not supported, strong effect 2,000/2,000 support.
  - `kernel_rounding.py` supplement: exit 0, 7 checks. Unlike round 7, the supplement was repeated independently.
  - Digests: 123 of 123 matched before and after.
  - `git status` was unchanged.
- Paths are relative to `evaluation/x1-triage-ranking/` unless stated otherwise.
- "P l.N" is a line of revision 8's `PREREGISTRATION.md`. "M l.N" is a line of `oc-manifest.json`. "ADR l.N" is a line of `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md`. "D8" is `reviews/rev8-digests.json`.

**Reviewed files (SHA-256).** Main checked all 123 entries of D8. Main reports the manifest's own SHA-256 as `8cbdb2114d79daea2d892da1cc294fe79ebfb305d56c37fd4c51d24bb1dd419e`. This round did not recompute any digest.

| File | SHA-256 |
|---|---|
| `docs/STATUS.md` | `1944051ef592b469cd0f72eebe6208156070eaa2faeeddad7d31660b71f3ea1a` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md` | `df26f688e3f86d767b20e18a92799fd8b5234e0d82c97a33d1005b48dbae69d8` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-1.md` | `7bc229fd392215971646bfc547883050e82512c21777fbd311cf779e3e394710` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-2.md` | `5a298d77bb778d79c17d158657d868ccdadb0ce8a5eef9d01a8d170a22f976ed` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-3.md` | `9530480f20aa59b2c4c72047dfab46c263d73dd86e716a5a566fee4d54f87038` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-4.md` | `0682fcf9a04f0e84e70b3d56e6275bfcb17bb5d9463b519b0c39c2d3dcf26d28` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-5.md` | `b7149420e1c21e245b2d1a1fb96ca4085fa6b84a1c5ea6c231a5a13aab1b0949` |
| `docs/adrs/README.md` | `c277db8d8703c087b8c922693abcf1be24b7e0d5006d7798e6aef2b4492ec602` |
| `docs/hypothesis-ledger.md` | `152cac84c3d94a72cbee598ce69c8ead446156bcd1e3fb3d2522100d85cd0753` |
| `docs/research-ledger.md` | `a40364be8998098fd97074c269c05df81eab3652fc27d14488724cca34fd851c` |
| `docs/scout-design.html` | `f7180b872e7380b8c2a644b8c7fe33b90c9ccf3e3cc0a4c8c3662227dbb3d8a9` |
| `PREREGISTRATION.md` | `9867c53b3cffad885d9fec70ad6f6855f48003693c343b2dc871b7c6a0685ef3` |
| `README.md` | `31dfaf249a393d47c94f60bbfa010ad7036321eacaf3530083551c32822572f9` |
| `inputs/x11-rev4-digests.json` | `e9edba67c885623df6900a08aee1cde463e5fa78f828bbb04ca3d8efd7a0324c` |
| `oc-manifest.json` | `99c95acada090fd9a1dad6d6110155a6c3b62bf4c9a288b2d9fb856714b47888` |
| `preparation/README.md` | `509ea3040fde082c32f167039b7252016e693a29d26a060da4613e74721a4058` |
| `preparation/dev-census.json` | `6faf0c4f5ca52345cc9399668829255996413d8b5b72e3b2df77d9cb2cad24a3` |
| `preparation/exposure-recipe.json` | `c065347e08a8571228764dae19fc686f3a6909e612a4119aad7bb58d86c5d383` |
| `preparation/exposure-sources.json` | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` |
| `preparation/frozen-exposure-inputs.json` | `ec92fba6d3f91c671301244539a8caefc5025c0d6b03efca77c5bdfc4bd96425` |
| `preparation/guide-author-isolation.json` | `c83c286df27f5a9edafdbcfe4213fcb7213d593812dbd6a4357e8369319d693f` |
| `preparation/model-free-binding-check.json` | `5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c` |
| `preparation/model-free-policy.json` | `87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f` |
| `preparation/pins.json` | `6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc` |
| `preparation/runtime-attestation-rev6.json` | `19463957becfc23a391eac03582b32eaa0ff392069bccd9aed9000f390ad2d62` |
| `preparation/runtime-attestation.json` | `3a924a0c8484eb51b9f62a8b0a805f9d6e5b023a3043e1a0ddd2b64d05e653ab` |
| `reviews/README.md` | `a6644b20d4abb49d5bb56bb3eeea45fbb5836788e1f7e4510fc3279af5a5c338` |
| `reviews/rev3-digests.json` | `894dc63b0522a23589187c6de9fa4e672878fc1c07b85bfbbdcd224e5ab2f448` |
| `reviews/rev4-digests.json` | `01c790837f961cd208fed98b3bd653b514da36e6d8bcee3b2fa4270f88e1876c` |
| `reviews/rev5-digests.json` | `ba1d2ae4efdf1c66e6d5f3002b5d0f892bac079a3efe4cbc2d98fd7826f69a95` |
| `reviews/rev6-digests.json` | `cca1baf432389b094588b6085888e8116442f0ee9baef15cbc750ef30086a746` |
| `reviews/rev7-digests.json` | `b108f52422c2724d95eb1dca2ced6bdd7da8245ec478af9888aae6864287f76d` |
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
| `reviews/round-6-response.md` | `143c06457606e01cdb61e63633f3144488956bdb281f0637793acadb14f58447` |
| `reviews/round-6.md` | `0682fcf9a04f0e84e70b3d56e6275bfcb17bb5d9463b519b0c39c2d3dcf26d28` |
| `reviews/round-7-response.md` | `9c77c8a8c026b73b4a99bca0e86175d43e711f651a9aaf1f880e376454f06835` |
| `reviews/round-7.md` | `b7149420e1c21e245b2d1a1fb96ca4085fa6b84a1c5ea6c231a5a13aab1b0949` |
| `validation/calibration-overhead-rev5-final.json` | `52d8f63527643395e0487535b0a23e06fc791357554244f19dadc64e66892f3b` |
| `validation/calibration-overhead-rev5-final2.json` | `5a340ebdfa26f35560e8ccce36f06f437cb1e6ad3e045153101e152b829c73dc` |
| `validation/calibration-overhead-rev5.json` | `b01d5a000bdb8a81b816bdee48f2cd240ed5f0b47a7a9100da0984ede9807073` |
| `validation/calibration-overhead-rev6.json` | `0a299b0014898a25c92e8a111bd68a9d1ddcf3f1edf8d547ca4319cd13c3221d` |
| `validation/calibration-overhead-rev7.json` | `415e3ac6c66b7bcca1d43f359f113417db2cb8f06a2f039ba87e71326bf9d954` |
| `validation/calibration-sources-rev5-before-opt.json` | `795283eb79ab1aa480f90ba0de613bcc0620f6b97281b9ece452449bf8ff11eb` |
| `validation/calibration-sources-rev5-before-transition-fix.json` | `5334c47b0260556a1e5a2e19608a202013aa8a4f34ef2ab33a8ac0fde0cf24cd` |
| `validation/compute-baseline-rev4.json` | `44a3b3e8ca82cbb98c3b9a9b8ef04e82053b0184eacf4d4d35f237ba6ffb3497` |
| `validation/compute-benchmark-rev4.json` | `e5f068d7badf2f43afcfbd2ee0c0d0a270efe605329a43a865ae114a7aab1ca8` |
| `validation/compute-benchmark-rev5.json` | `fc086226d92a622f32b289dc6b88a6716fd31573fd98d27dedb200fb9f0ed5df` |
| `validation/compute-benchmark-rev6.json` | `9662eb5e9d9344effed4f3141041cf86856c869c9a5219b0f0311eba32076a1f` |
| `validation/compute-benchmark-rev7.json` | `f268997dba849fd256916e556817764638d4a6c858e6458fccb7ad85a23801b8` |
| `validation/compute-projection-rev4-final.json` | `b422c2a6fc7de6e548fa94cc4f36beec403ea9b169aff0429fb01386c44947dd` |
| `validation/compute-projection-rev4-total.json` | `65752e13309281fc8d1019034428103e5f2a445248bc1c142a958b402713a3c1` |
| `validation/compute-projection-rev4.json` | `4d2ca24912b6c94689138ff626869bd7473747a2dfde86497a59925f6e10dc0b` |
| `validation/compute-projection-rev5-final.json` | `3d6c2ff4c76e96c9c215025e4237744e5d6e53540f10966d3e09df6e603638cc` |
| `validation/compute-projection-rev5-final2.json` | `7ebbbb5da4798aac8acc93c6bb1616042075c77a7f62e912027a1ffb7cf8e121` |
| `validation/compute-projection-rev5.json` | `ee616c45674f9beb82bbc8c3951576cf955e1e3c62a2281089fc0b6ec474ad4b` |
| `validation/compute-projection-rev6.json` | `972b3620d76f8f2d344aadc47e07b7b226c445828a166f264a215cceeb7b9199` |
| `validation/compute-projection-rev7-L2000.json` | `f1671be3649f8b7693c5c019a33521bcaa3d58c17ef94bd4c10ffde7088447af` |
| `validation/compute-projection-rev7-U12000000.json` | `7e3961e8af995cb44917ef26028383974661ce8809b71fe484dfc3fdc7adb0ec` |
| `validation/compute-projection-rev7.json` | `281270a3d91f054980b032ef99e1f05a7a41beed2e743feb8254439022d16121` |
| `validation/compute-projection-rev8.json` | `2854633dbef1e74d69571ca6f7ae626ed51b1047445b2463a1a3af962965882a` |
| `validation/compute_benchmark.py` | `d39f40a28a6fd70698ea9e53e30448dec193c395d264002436dad9278ea6c90d` |
| `validation/fixed_population.py` | `ed670a6bf8076d791c06cfc2426aff03a0a2f9b1730982cf3fdfb2ec13bc49a2` |
| `validation/interval-diagnostic-rev4.json` | `50d41cc86c0342e05110d322b5d452ffcad7570c612efbbaee705ce11cc18a54` |
| `validation/interval-diagnostic-sources-rev4.json` | `d12d80627cecd6e7ee0b291dee939638cf7752c24ae62164bb495bf0ab693722` |
| `validation/interval-diagnostic-spec-rev4.json` | `e5a93a49c14a4345cba2fd6f2d4ddc55a998eaa199d371299d5965443fe54bb6` |
| `validation/interval_diagnostic.go` | `437f9c2559156f8befb613d3e17f19e2aad5437217081dbd4a1d0fb95c069417` |
| `validation/interval_diagnostic.py` | `a4b233374bdfec466bfc5862166497a4bbac48b13668aee2697fcb90e2b7c345` |
| `validation/kernel-rounding-rev7-final.json` | `c974259794fb559b00187767562fe0d552268aac3608da5544e6812bb2f26446` |
| `validation/kernel-rounding-rev7.json` | `133319e305f9c23f7c5e660860cd648afe6336c09d6fccb5fd9af930333d4ff9` |
| `validation/kernel-rounding-rev8.json` | `d667c6cff997fdb847d151a93cfe1cb69bd5d3960a3b1d6d601c46f24048c51c` |
| `validation/kernel-sources-rev6.json` | `66193e0fee0da8e8a2e386cac973329c81267a893657314bfcf0f42616579e22` |
| `validation/kernel_rounding.py` | `906d385f59ec61d0ba1c508cad3295b561f7a3504e9933e335a246751b657c23` |
| `validation/kernel_rounding_test.go` | `33bde17be1cf67a0da47661f384901e79b60915f5cb8e6fd9d2f0652b1ff8def` |
| `validation/optimization-refusal-rev7.json` | `ef94cf73c2be6673905c773a71fa1f640987cc96b7135f64c09e9f87e099adc7` |
| `validation/outer-results.json` | `b851c2c1b6ffacee1957ade0071aa9fd38ede0dce5ebd2e3577a2967923f6352` |
| `validation/outer_simulation.py` | `91d07760b61551810282622c2bda53c34fcfcb72a9ffd1798b1a5ffb8ece23f6` |
| `validation/pilot-summary-rev4.json` | `05da900b22b738a61c36d4f02efc76c722109aa2dcffff40fba04f828c14475d` |
| `validation/pilot-summary-rev5.json` | `c21a5a13ea18e521c7fd8469d93b4547471924dd740c75f1c2368fc2ea8ed17e` |
| `validation/ranking.py` | `8bc2e7cbbc5da6ba47aeff3effb5dd04d5e5ecc9c1e01051647900095770a780` |
| `validation/reproduce_reports-initial-rev5.py` | `0b7781db46f78e81a0ac091bf123a3fe1e4045925c09396a6299065119a10ed4` |
| `validation/reproduce_reports-optimized-rev5.py` | `d680ee085bc20525761fdedf924f9d5ac0be12984be9a10c1e3ac4ff5b102887` |
| `validation/reproduce_reports-pilot-rev5.py` | `3ee6476c7d142bc1e7340fa9ba31d668aa42cc92d744c66cc5dc910638cf0ee2` |
| `validation/reproduce_reports.py` | `023e2e5bc6ca1ee282e4ebc75ef6b3bccdb96111d0f218c0341189d6a400168e` |
| `validation/rev5-reviewed-sources.json` | `fa6820c87e068c16b102e2b0d7647c31c6a8f58eccca8f3d3bb8b437fe2bf89d` |
| `validation/rev6-executed-sources.json` | `fafdcbc291a2521e1bf2b5f90ea43e7046135140c9bbb579c69ad41c4e9e6086` |
| `validation/rev6-reviewed-sources.json` | `33d0fc21db11c381d051a331e7c5db3a765889abba52a79b294336f080eac7cb` |
| `validation/rev7-executed-sources.json` | `13c805f30d5b8ce2442464e3a754f8df684664669a40659e211bf55d5f21a88a` |
| `validation/rev7-reviewed-sources-rev8.json` | `f402aa9e49fb744336a0d47cc4550c12c71f5071b53fbe9f07ba18fa2664ba27` |
| `validation/rev8-executed-sources.json` | `7412fae0f72de7a70cc208a10e6f65e3471b256fad4af79fd7e89a385af49d5d` |
| `validation/rev8-preexecution-sources.json` | `7412fae0f72de7a70cc208a10e6f65e3471b256fad4af79fd7e89a385af49d5d` |
| `validation/revision8-checks.json` | `01db1d1cb27ae862850f91443f90ce3f7439b808d4aac9136aad934676311101` |
| `validation/self_test.py` | `36d07d5648fdb0db325cb7583b02c4117cab1ada07b5fefe4590c66c8f03eb8a` |
| `validation/snapshot-closure-rev8.json` | `cc46b23623b049013abac27c6da3a9dd5f971b98ff9b1745f8f442e9d387efd1` |
| `validation/source-snapshot-checks-rev7.json` | `bfa5a445272441fa11cb95fe92a988a1cfd74c3dfbf71d61c37449aa2f21a5b1` |
| `validation/strong-effect-OC-rev5-closure.json` | `379f2c28ac5fa79ae738d2f1ad18acf823d389ff717616a0d8e37147a4e582fb` |
| `validation/strong-effect-OC-rev5-final.json` | `6f26bad6dcf9f9d7414757ae3ca73b68990187b3d3b52edadd856cb6613a4375` |
| `validation/strong-effect-OC-rev5-optimized.json` | `19c78ee133a87c3f05a9319e58b7bff59031e886837733edb668c566d945615b` |
| `validation/strong-effect-OC-rev5.json` | `089941397a9036830532ffe52410f3f3ca65a153db4e6cfa3e57ec2960ddbb8b` |
| `validation/strong-effect-OC-rev6.json` | `96cbbeaae60431010a97964a39ced7522d2c80d90eeb81d49ebc69945355b476` |
| `validation/strong-effect-OC-rev7.json` | `da5cc7ad3a833788407046ff6317e0d5e8bcd1acf0098d6b642d1b1a771d1b41` |
| `validation/strong-effect-OC-rev8.json` | `38bbcf56f0c1c570db32cbf61196a60bca05ffc7760f82b11c93f28b7f813a40` |
| `validation/synthetic-results-rev1.json` | `fbc4cc0de2c948574a76216cffd12cb33b488f0fedd3c3519de4ddf0b5a82277` |
| `validation/synthetic-results-rev2.json` | `24549f3ed2fcaecb357bf747ec31058ae115592f8bf63a1319763d2a84faa77d` |
| `validation/synthetic-results-rev4-percentile.json` | `b206c43600f80dab426f303a915695176418680296105144545e332dca04c56d` |
| `validation/synthetic-results-rev4.json` | `9f3ec381a945028c5a246ea723f03ee771815ff3bfdf609c917505e12c66c2f2` |
| `validation/synthetic-results-rev5-closure.json` | `f0ae8b5c7b341d2f5d932c72a416ae8451f99429c973e7079daad1869324c134` |
| `validation/synthetic-results-rev5-final.json` | `f5ea12d6d3df4043610921e98fb42a8ceea47f973ca0bd5311ba388fdeabe124` |
| `validation/synthetic-results-rev5-optimized.json` | `b235853e5a6b185c8908543a82430c0cb569261dfb200434e3d8d27a19c34392` |
| `validation/synthetic-results-rev5.json` | `a7a2c26a3fbf7eedf60216bddea196364022453769c95c06f7029946385b552b` |
| `validation/synthetic-results-rev6.json` | `92b935e4685b597c1fe120c5057a6e0fb1cc2dd81657a43ebae4c02e3f8c6073` |
| `validation/synthetic-results-rev7.json` | `616bf68453d1ed3ba01bf3286496175459c99e9fc42c33612890c011beaf707f` |
| `validation/synthetic-results-rev8.json` | `d1749e3edad7faeb5c982eb39d64fa8fb36651c4cda332523626f35404339948` |
| `validation/synthetic-results.json` | `e65ec36c7d57a0d15f32bc3d2d7bd5f1f76f4148189ec2676653ef996901a350` |
| `validation/synthetic-sources-rev5-initial.json` | `a77aff4f00d87a5d98dafe3114df227b9de27ffb215c45a82ef6074781aee4c2` |

**Byte-preservation checks against the round-7 table.**
- **Unchanged (100 files):**
  - every earlier council round and response from rounds 1–6;
  - ADR-014 `round-1.md` to `round-4.md`;
  - `rev3-digests.json` to `rev6-digests.json` and `inputs/x11-rev4-digests.json`;
  - `docs/adrs/README.md` and `docs/scout-design.html`;
  - all 11 preparation files;
  - every revision-4 to revision-7 validation output, reporter snapshot and source snapshot, including both revision-7 kernel-rounding outputs and all three revision-7 projections;
  - every executable except the two listed below.
- **Changed mutable documents (8):**
  - `docs/STATUS.md`, ADR-014 `ADR.md` and both ledgers;
  - `PREREGISTRATION.md`, `README.md`, `oc-manifest.json` and `reviews/README.md`.
- **Changed executables (2):**
  - `kernel_rounding.py` (`395fb9e0…` → `906d385f…`);
  - `reproduce_reports.py` (`dd749c32…` → `023e2e5b…`).

  Both old digests, and the unchanged `kernel_rounding_test.go` (`33bde17b…`), are archived with their content in `validation/rev7-reviewed-sources-rev8.json`. The new bytes are sealed in `rev8-preexecution-sources.json` and `rev8-executed-sources.json`.
- **New files (13):**
  - ADR-014 `round-5.md`;
  - `rev7-digests.json`, `round-7.md` and `round-7-response.md`;
  - `compute-projection-rev8.json` and `kernel-rounding-rev8.json`;
  - `rev7-reviewed-sources-rev8.json`, `rev8-executed-sources.json` and `rev8-preexecution-sources.json`;
  - `revision8-checks.json` and `snapshot-closure-rev8.json`;
  - `strong-effect-OC-rev8.json` and `synthetic-results-rev8.json`.

  100 + 8 + 2 + 13 = 123, and 110 + 13 = 123.
- **Identical pairs:**
  - `round-7.md` = ADR-014 `round-5.md` = D8 `joint_previous_record_sha256` (`b7149420…`).
  - `rev7-digests.json` = `b108f524…`, the manifest digest Main reported in round 7.
  - `rev8-preexecution-sources.json` = `rev8-executed-sources.json` (`7412fae0…`). The executed bytes did not change during execution.
- **Cross-references:**
  - The projection's `benchmark_sha256` (`f268997d…`), `calibration_overhead_sha256` (`415e3ac6…`), `reporter_sha256` (`023e2e5b…`) and `source_closure.snapshot_sha256` (`13c805f3…`) equal the table entries.
  - `kernel-rounding-rev8.json` records source hashes `437f9c25…`, `906d385f…`, `33bde17b…` and `8bc2e7cb…` and interpreter `570b4105…`. All equal the table entries and D8 `numerical_runtime`.

**Verdicts.**
- **X1: 2 issues open (0 blocking, 0 major, 2 minor: X1-R8-01 and X1-R8-02). No sign-off.**
- **ADR-014: 2 issues open (X1-R8-01 and X1-R8-02).** Both need edits to ADR-014's table (l.45 and l.40), and both are shared with X1. **No sign-off.** ADR-014 can be adopted only on the joint zero-issue final round.
- **Joint signed set: 2 distinct issues open.**
- No preparation, inventory or draw, Jev, labelling or analysis stage is authorised (§12).

**Severity scale.** Unchanged from rounds 1–7:
- **blocking:** can produce a wrong or unreachable verdict, or breaks experiment rules 1–7 or REQ-169.
- **major:** misleads the result, leaves a preregistered quantity undefined, or makes the study unrepeatable or likely not testable.
- **minor:** text, consistency or provenance. Minor issues must still be fixed before sign-off (rule 1).
- **Calibration:** a fail-closed default that only makes a gate harder to pass is not a defect, unless it makes the gate unreachable for a compliant case.
- **Undefined procedural details are minor.** Both open issues leave procedural details undefined. As with X1-R7-02/03 and X1-R7-04, they are minor, not major: neither can change the verdict logic, and the undefined part can only make the resource fit easier to pass (X1-R8-01) or descriptive reporting selective (X1-R8-02).

**Council attribution.** Both open issues are gaps left by wording the council itself required in round 7:
- X1-R8-01 arises from the X1-R7-03 phrases "registered replicate structure" and "(crash, closure mismatch or host fault)".
- X1-R8-02 arises from X1-R7-04, which fixed the execution order but not the scheduling.

The author implemented both round-7 fixes exactly as worded. To avoid another iteration, §5 gives each fix as verbatim replacement text.

---

## 1. Recommendation

**Revise to revision 9 with the two text-only fixes in §5. Do not change any executable, rerun anything or regenerate any output. Then submit revision 9 for X1 round 9 / ADR-014 round 7. Do not sign off now.**

Until then, authorise nothing beyond:
- drafting;
- synthetic validation;
- option G model-free engineering on synthetic repositories.

On a zero-issue round 9, authorise **only preparation actions (a)–(e)** (P l.222).

The five round-7 items are all fixed as required (§4). The revision-8 reserve and CPU arithmetic reproduce exactly by hand (§8). The single binding resource stop rule is character-identical in P l.181, ADR l.45 and M l.165.

Two minor gaps remain, both procedural and both in the round-7 fix text:
- **X1-R8-01:** the locked-remeasurement procedure still has two forking paths:
  - the replicate structure is undefined;
  - crash and host fault are unclassified, so a resource-exhaustion crash or an aborted run could buy a fresh timing run.
- **X1-R8-02:** the safety order is fixed, but nothing stops start, allocation, pausing or termination of descriptive safety from depending on safety outcomes.

Keep everything else, including:
- the reserve-adjusted memory rule;
- the one-run maximum rule;
- the keyed safety order and the "not completed by gold" status;
- the supplement's runtime and binary closure;
- the snapshot discipline;
- the stage table.

## 2. Alternatives considered

| Option | Description | Tradeoff | Council position |
|---|---|---|---|
| A. Sign off revision 8 and carry the two minors | Starts the clock sooner | Rule 1 requires zero issues. X1-R8-01 sits in a binding stop rule. | Rejected |
| B. **Revision 9: text-only fixes in §5, then a zero-issue round** | Four mirrored text edits. No executable, rerun or output change. | One more round | **Recommended** |
| C. Authorise preparation now, because neither issue touches actions (a)–(e) | Starts fetch and attestation | Rule 1 covers the whole reviewed preregistration, and the clock would start under a protocol that is still changing | Rejected |
| D. X1-R8-01: peer's broader fix. Every interrupted or integrity-failed measurement enters the maximum; soften P l.240. | Maximally conservative | Closure-invalid measurements are not trustworthy binding inputs (challenger) | Rejected in favour of fail-closed classification: a termination without a qualifying integrity failure leaves the fit unmeasured |
| E. X1-R8-01: leave the replicate structure to the stage-3 review of the locked-size command (challenger) | Avoids text now | That review happens after inventory reveals the sizes. The maximum trial CPU is an extreme-value statistic, so the trial count chosen then moves the binding input. | Rejected. One clause fixes it now. |
| F. X1-R8-02: compulsory continuous use of every spare worker (peer) | Removes stopping discretion | More restrictive than needed. Conflicts with labelling overlap within the same caps. | Rejected in favour of an outcome-independent scheduling policy frozen before the first safety output (challenger) |
| G. Option G: write the locked-size benchmark and calibration mode and the OC controller against synthetic sizes before preparation | Shortens the clocked path and pre-empts stage-3 review findings | Pre-clock work on no real data | **Recommended practice**, not a sign-off condition |

## 3. Rule checks

- **Rules 1–2.** This round is recorded with digests. Issues are open, so iteration continues.
- **Rule 3.** No predictor–outcome association was computed on real data, and all revision-8 evidence is synthetic. The historical method selection was not rerun (D8 `historical_method_selection_rerun: false`). The benchmark and calibration were not rerun (`benchmark_and_calibration_rerun_in_revision8: false`).
- **Rule 4.**
  - Planted effect: 2,000/2,000 support, Wilson MC95 lower .9980829527187469, realised gain .463, zero inadequate (`strong-effect-OC-rev8.json`).
  - Planted null: not supported.
  - Both were reproduced in Main's independent run.
  - Supplement: seven counted checks with full closure (`kernel-rounding-rev8.json`).
- **Rule 5.**
  - Nothing is signed yet.
  - Both changed executables were archived before modification (`rev7-reviewed-sources-rev8.json`), and the executed bytes were sealed before and after execution.
  - Both revision-7 supplement outputs are labelled superseded (README l.85; D8 `superseded_outputs`).
  - The loss of the `d6d4701e…` bytes is disclosed (README l.85; D8 `source_loss.revision7_supplement`).
- **Rule 6.**
  - Binding and total envelopes, the reserve-adjusted memory envelope and the timing drift are reported (README l.102–108; P l.183, l.240).
  - A partial safety grid now has status rules. Its scheduling independence is X1-R8-02.
- **Model use and secrets.** No model was called. The binding recipe still checks the key by presence only, with stub-only positive controls.
- **Toolchain.**
  - Numerical evidence ran on `.tools/venv/bin/python` 3.11.11 (`570b4105…`) and `go1.26.10 darwin/arm64`.
  - `toolchain_pin_changed: false`.
- **Design rules.**
  - ADR-014 has a sequential ID, an adopted-form status matching ADR-011 (D8 `adr_status_check`) and Review deferral to the final round in its folder (ADR l.73).
  - Its Records list includes `round-5.md` (ADR l.82).
  - The "Informed by" comma is fixed (ADR l.65).
  - The index row and design card are byte-unchanged and still accurate.
  - Two issues require table edits.
- **Ledgers.**
  - HD9 remains proposed (`docs/hypothesis-ledger.md:36`).
  - The revision-8 history line is new (l.40). The revision-7 line is byte-unchanged (l.41; extract `ae6d74…`). The revision-6 line is corrected (l.42).
  - The next free ID is R-087 (`docs/STATUS.md:128`), and no ID was consumed.
  - R-081 (`docs/research-ledger.md:83`) was updated in place under Main's standing permission.

## 4. Verification of round-7 fixes

| Issue | Status | Evidence |
|---|---|---|
| X1-R7-01 | **Fixed** | **Version refusal:** `kernel_rounding.py:18–19` refuses anything but 3.11.11. A non-zero optimization flag is refused through `ranking.provenance` (`ranking.py:389–398`).<br>**Interpreter and Go:** path and SHA-256 of the interpreter and the `go version` are recorded (`:25–27`).<br>**Test binary:** `go test -c -o` builds it; it is hashed and then that exact binary is run (`:36–43`).<br>**Kernel:** the interval kernel is hashed (`:66–70`).<br>**Counted checks:** checks are counted as they run (`:30–33`; `'checks_passed': len(checks)`, `:78`), plus a check that source digests are unchanged (`:75–76`).<br>**Create-once:** the output is refused if it exists and opened exclusively (`:20–21`, `:83`).<br>**Output:** `kernel-rounding-rev8.json` lists seven named checks, both binary digests, the interpreter, Go and source closure, and 6,000 zero-variance classifications.<br>**Snapshots and labels:** taken before modification and after execution; both revision-7 outputs labelled superseded, with the source loss disclosed. |
| X1-R7-02 | **Fixed** | **Rule:** 8 × max(construction, calibration, trial worker plus helper peak RSS), lifetime high-water, ≤ hw.memsize − 8 GiB (60,129,542,144 B). Workers fixed at 8; failed or unmeasured means not testable.<br>**Mirrored literally:** P l.181, ADR l.45 and M l.165 (`revision8-checks.json`: "ADR/manifest mirrors operational resource rule exactly").<br>**Planning:** the planning test applies the reserve (`reproduce_reports.py:259–261`). `fits_caps` requires both U ≤ the memory break-even and binding hours ≤ 2,880 (`:327`).<br>**Projection:** create-once revision-8 projection from byte-identical revision-7 benchmark and calibration inputs (`revision8-checks.json`). README l.104 publishes it. Arithmetic in §8. |
| X1-R7-03 | **Fixed as worded.** The residual is X1-R8-01. | **Rule text:** one create-once run, integrity-only reruns, "a failed fit never permits a rerun", every completed run reported, each input the maximum over completed runs, and load-average and concurrent-process provenance (P l.181 = ADR l.45 = M l.165).<br>**Drift disclosure:** P l.240 and README l.108 (+66.2 %, +30.3 %, +31.3 %, +8.1 %; host load [INFERENCE]). |
| X1-R7-04 | **Fixed as worded.** The residual is X1-R8-02. | **Text:** key `169-x1-safety-order` over the canonical `safety-interval-group` identity; the criticality phase runs in the same order with no reordering; a group is measured only when all five populations reach 2,000; exhaustion keeps "unmeasured; cell fails"; "not completed by gold" is a separate status, with partial counts shown separately and excluded from pooled Wilson summaries.<br>**Where:** P l.171, ADR l.40, M l.143, README l.93; cross-referenced at P l.224.<br>**Identity:** the canonical identity is defined (`fixed_population.py:199–205`; M l.47, l.51–53). |
| X1-R7-05 | **Fixed** | `docs/hypothesis-ledger.md:42` now reads "worst-case enumeration and L400 RSS projection", with no memory envelope. Line 41 is unchanged (`ae6d74…` in both rev7 and D8).<br>The revision-6 extract is refreshed (`ec3ceabb…`; the peer confirmed it differs from rev7's). A new extract covers the revision-8 history line (`ed47dca4…`). Main's permission is recorded (`round-7-response.md:11`). |

## 5. New issues

### X1-R8-01: minor. The locked-remeasurement procedure leaves the replicate structure undefined and does not classify crashes, host faults or operator aborts

- **Location:** P l.181, l.240; ADR l.45; M l.165 and l.191; README l.108; `docs/research-ledger.md:83` (R-081 paraphrase).
- **Evidence:**
  - **No locked-size mode exists.** `compute_benchmark.py` hard-codes its sizes (`:31–32`, `:163`, L = 40/400 at `:268`). Its CLI takes only `--method`, `--output`, `--trials`, `--cores` and `--oc-days` (`:220–225`). Calibration is fixed at `pool.map(calibration_measurement, (40, 400))` (`reproduce_reports.py:151`).
  - **The binding trial cost is a maximum** over every measured trial (`reproduce_reports.py:192–209`). The number of trials and cells measured therefore moves the binding input: fewer trials give a stochastically smaller maximum.
- **Problem:**
  - **(a) Undefined replicate structure.** "The registered replicate structure" is defined nowhere. A locked-size command must be written and reviewed after inventory reveals the sizes, so the trial and cell structure behind an extreme-value binding input would be chosen with the sizes known.
  - **(b) Unclassified crash and host fault.** "Crash" and "host fault" permit a rerun, while "failed fit" does not, but nothing classifies:
    - a resource-exhaustion (out-of-memory) crash, which is evidence of a failed memory fit;
    - an operator abort of a slow-looking run that is relabelled a host fault.

    Only completed runs enter the maximum, so either path can buy a fresh timing run. The CPU drift between revisions 6 and 7 was up to 66 %.
  - **(c) Host not tied.** The rule's thresholds are host-relative ("hw.memsize", "on this host"), but nothing ties the remeasurement host to the host that runs the binding OC.
- **Why minor:**
  - Each gap can only make the resource fit easier to pass; none changes the verdict logic.
  - Ordinary timing noise was accepted in round 7 (§9) and is not reopened. This issue concerns classification and definition only.
- **Required fix** (recommended option; text only, no executable change):
  - In P l.181, replace the sentences from "Locked remeasurement is one create-once run" through "each binding input is the maximum over all completed runs." with the text below. Mirror it verbatim in ADR l.45, M `resources.binding_resource_stop_rule` and D8's copy.

    > Locked remeasurement is one create-once run, on the host that will run the binding OC (record hw.model, hw.memsize, logical CPU count and OS build), of the reviewed benchmark and calibration commands with the revision-7 replicate structure (README benchmark invocation `--method studentized --trials 2000 --cores 8` with its cell set and one/two-worker equality; the calibration-overhead command's per-population measurements and worst-case enumeration), changing only the size source to the locked sizes; that size input is the only permitted code change and is reviewed under rule 5. Record load averages and concurrent-process count in provenance. Rerun only after a recorded integrity failure: a crash not caused by resource exhaustion, a closure mismatch, or a host fault evidenced by a contemporaneous OS-recorded malfunction independent of observed timings or partial results; record the reason and preserve that evidence and the interrupted run's available output. A resource-exhaustion crash is a failed fit. Operator termination and host load are not integrity failures; a run terminated without a qualifying integrity failure leaves the fit unmeasured. A failed fit never permits a rerun. Report every completed run; each binding input is the maximum over all completed runs.
  - Update the paraphrases in README l.108 and R-081 (`docs/research-ledger.md:83`) to match.
  - P l.240 needs no change once this rule is in place.

### X1-R8-02: minor. Starting, allocating, pausing or terminating descriptive safety work may still depend on safety outcomes

- **Location:** P l.171, l.185; ADR l.40; M l.143; README l.93.
- **Problem:**
  - The keyed order fixes *which* group comes next, but nothing fixes *when* safety starts, how many workers it gets, or when it pauses or stops.
  - An operator who has seen outputs could halt an in-progress group whose partial counts look poor. That group would then become "not completed by gold" and drop out of the pooled Wilson summaries.
  - Disclosing partial counts makes this visible, but it does not remove the selection from the pooled summaries.
  - The general "no efficacy peeking" text (P l.216) belongs to the clock and model-call rules. It does not clearly govern scheduling of synthetic descriptive outputs.
  - This is a rule-6 reporting risk. Safety cannot gate dev/test or the verdict.
- **Not part of this issue:**
  - Precedence between attempt exhaustion and "not completed by gold" is already settled: exhaustion "retains 'unmeasured; cell fails'" (P l.171), and "not completed" reports completion status only.
  - Continuous use of every spare worker is not required (§2 F).
- **Required fix** (recommended option; text only):
  - In P l.171, after "with no reordering after the first output.", insert the text below. Mirror it verbatim in ADR l.40, M `final_safety_OC.schedule` and README l.93.

    > The descriptive-safety scheduling policy (start, worker allocation, pauses and termination) is frozen with the reviewed stage-3 OC controller before the first safety output; none of these may depend on partial or completed safety outcomes, and every pause or termination is recorded with its cause.

### Considered and not counted as issues

- **Labelling of the revision-7 boundary projections** (peer's candidate, rejected after the challenger's rebuttal).
  - `compute-projection-rev7-L2000.json` and `-U12000000.json` still carry the reserve-free memory fields (12,828,553; "fit available memory"), and D8 `preserved_superseded_projections` lists only `compute-projection-rev7.json`.
  - However, P l.185 supersedes the "revision-7 reserve-free … projections" collectively, in the plural. M l.184 and README l.106 scope that supersession to planning memory. README l.104 publishes the 11.22 M break-even, from which the U = 12 M memory exceedance follows.
  - **Advisory:** when D8 is regenerated, list both boundary files in `preserved_superseded_projections` and name all three at README l.106.
- **The projection's own interpreter digest.** `compute-projection-rev8.json` carries the benchmark's interpreter digest (`reproduce_reports.py:184–185`) and the path in `reproduce_cli`, but not the reporter run's own digest. This is not raised:
  - the reporter refuses anything but 3.11.11 and refuses optimization at run time (`:345–348`);
  - the output is deterministic arithmetic over byte-sealed inputs;
  - round 7 reviewed the same pattern.

  **Advisory:** record `python_executable` and its SHA-256 the next time the reporter changes.
- **Sealing the frozen safety order** (peer advisory). The SHA-256 of the ordered group list can be recorded with the stage-3 controller. It sits at the same definitional level as the accepted OC seed JSON (M l.47).
- **Check-output generators.** `revision8-checks.json` and `snapshot-closure-rev8.json` have no committed generator. This is the same pattern round 7 accepted for `source-snapshot-checks-rev7.json`, and Main verified the digests independently.

## 6. Assessment of targeted points

- **Literal identity of the stop rule.** P l.181 and ADR l.45 are character-identical from "Binding resource stop rule:" to "…maximum over all completed runs." (lead reading). M l.165 matches (`revision8-checks.json`; challenger). P then continues with kernel-parity sentences that are not part of the mirrored rule.
- **Reserve.** The rule is internally consistent:
  - 2,880 core-hours ÷ (8 × 24) = 15 days;
  - workers cannot be reduced;
  - memory is charged at the maximum over phases, with lifetime high-water measurement already implemented (`compute_benchmark.py:45–47`, `:141–142`, `:170–172`; `reproduce_reports.py:120–122`).

  The reserve has teeth: at U = 12 M, projected memory of 64,297,238,528 B fits physical memory but exceeds the 60,129,542,144 B budget (`revision8-checks.json`).
- **Completed-run maximum.** For runs that complete, the maximum rule is fail-closed: rerunning can never lower a binding input. The residual paths are runs that never complete and the crash classification (X1-R8-01).
- **Safety statuses.**
  - Trial-level safety-unmeasured stays in the probability denominators.
  - Attempt exhaustion keeps its cell-failure meaning.
  - "Not completed by gold" is a group-completion status.
  - None of these collide. The remaining gap is scheduling only (X1-R8-02).
- **Supplement.** It now has the self-test's provenance pattern. The rerun was repeated independently.
- **Stale values.** No operative "physical memory ÷ 8", 12,828,553 or "available-memory" rule survives in current prose:
  - R-081's "available-memory fit" is an explicit disclaimer;
  - ADR l.45 names it only as a rejected option.
- **Preparation actions (a)–(e)** (P l.222) are well defined: pinned SHAs and URLs (P l.33–35; `dev-census.json`), recipes, clock start and prohibitions. Neither open issue touches them.

## 7. Checked and sound

- **Counts** agree across P, M, D8 and the projection:
  - 17,920 complete/reference cells and 252 missing cells;
  - 336 safety profiles and 8,064 outcomes;
  - 336 complete + 1,680 full-vertex interval groups;
  - 90,860 constructions;
  - 80,640 requested and 40,320 canonical evaluations.
- **Timing inputs and all CPU projections are unchanged** from revision 7 (`revision8-checks.json`). Binding 711.8476906426667 h and total 1,288.4370581039998 h equal round 7 §8.
- **Snapshot closure.** All 12 archives close on content and hash (`snapshot-closure-rev8.json`), and 23 affected-revision checks pass (`revision8-checks.json`).
- **ADR-014:**
  - the decision text names revision 8 and "R3–R7 are also this folder's rounds 1–5" (ADR l.15);
  - the safety row and compute row match P and M;
  - the Informed-by list matches HD9's sources.
- **STATUS and the reviews README** state revision 8 as a draft with no authority, and name the next joint round correctly.
- **No over-authorisation.** P l.220–226, the recipes and ADR l.15 grant no stage.

## 8. Evidence index

- **Memory** (`calibration-overhead-rev7.json`; `compute-projection-rev8.json` `memory`):
  - slope = (831,799,296 − 111,263,744)/1,080,000 = 667.1625481 B per unit;
  - intercept = 31,204,238.22 B;
  - budget = 68,719,476,736 − 8,589,934,592 = 60,129,542,144 B, which is 7,516,192,768 B per worker;
  - break-even = (7,516,192,768 − 31,204,238.22)/667.1625481 = 11,219,137.75 units [INFERENCE: linear];
  - at U = 1.2 M: 8 × 831,799,296 = 6,654,394,368 B (fits);
  - at U = 12 M: 8 × (31,204,238.22 + 8,005,950,577.78) = 64,297,238,528 B, which exceeds the budget.
- **CPU at U = 1.2 M, p_gate = .5:**
  - binding = 397.1563 + 306.9737 + 8,000 × 9.647116e−4 (= 7.7177) = 711.8477 h, which ÷ 192 = 3.7075 days;
  - total = 851.5762 + 428.2870 + 8.5739 = 1,288.4371 h (6.7106 days).
- **Break-even U at p_gate = .5:** binding 9,675,588.88; total 5,659,336.25. Both unchanged.
- **X1-R8-01:**
  - `compute_benchmark.py:31–33`, `:163`, `:220–225`, `:268`;
  - `reproduce_reports.py:151`, `:192–209`;
  - P l.181; README l.122 (the `--trials 2000` invocation).
- **X1-R8-02:** P l.171, l.185, l.216; M l.143.

## 9. Risks

- **Iteration risk.** The council's own fix wording has now left residual gaps in two consecutive rounds. §5 therefore gives verbatim text, and round 9 should check only for literal adoption and mirroring.
- **Throughput screen.** This is carried from round 7 §9 and is still the tightest constraint. Option G reduces exposure.
- **Stage-3 artefacts not yet written:**
  - the locked-size benchmark and calibration mode;
  - an OC controller that enforces both counters and the frozen safety policy;
  - the final evaluator;
  - the product-backed harness.

  Each needs a rule-5 review.
- **Host binding.** After the fix, the binding OC must run on the remeasurement host. A host change means a new locked remeasurement.
- **Shared-file staleness.** Whole-file digests of STATUS, the ledgers and the index go stale if a sibling change lands first. The owned extracts mitigate this.
- **Reversibility.** All fixes are pre-data text changes. No product setting, default, accepted ADR or requirement changes.

## 10. Where the peers changed the council's recommendation

- **Challenger (independent pass):**
  - Found no issues.
  - Independently reproduced every memory and CPU figure.
  - Supplied the U = 12 M memory figure, 64,297,238,528 B.
  - Confirmed the literal mirroring and the lifetime high-water implementation sites.
- **Experiment peer:**
  - Found the undefined replicate structure and the hard-coded sizes, with the extreme-value sensitivity of the trial-CPU maximum.
  - Found the abort/host-fault/out-of-memory classification paths and the missing host binding.
  - Found the safety stopping lever.

  These became X1-R8-01 and X1-R8-02. The lead had concluded that the completed-run maximum was fail-closed and had missed the paths for runs that never complete.

  The peer also proposed a third issue (boundary-projection labelling). It was **rejected** on the challenger's evidence (P l.185) and kept only as advice.
- **Challenger (rebuttal pass):**
  - Narrowed both fixes. Interrupted or closure-invalid measurements are not forced into the maximum; P l.240 is not softened; continuous full-worker scheduling is not required; the exhaustion precedence is already covered.
  - Argued that the replicate structure could wait for stage-3 review. The lead **did not adopt** this, because that review would see the locked sizes (§2 E).
  - Judged a separate same-host sentence unnecessary. The lead kept host identity as one clause inside X1-R8-01's rule, because the thresholds are host-relative.
- **Lead:**
  - Verified the 123-file accounting and the cross-references.
  - Checked the projection's interpreter provenance and decided not to raise it (§5).
  - Wrote both fixes as verbatim mirrored text.

## 11. Execution outline for revision 9 (one integration owner)

1. **Stop rule (X1-R8-01).**
   - Replace the locked-remeasurement sentences in P l.181 with §5's text.
   - Mirror verbatim in ADR l.45, M `resources.binding_resource_stop_rule` and D9's copy.
   - Update the paraphrases in README l.108 and R-081 (Main's in-place permission).
   - *Test:* literal equality across P, ADR and M, with exactly one operational rule in P.
2. **Safety scheduling (X1-R8-02).**
   - Insert §5's sentence in P l.171.
   - Mirror verbatim in ADR l.40, M `final_safety_OC.schedule` and README l.93.
   - *Test:* literal equality across all four.
3. **ADR-014.** Add to Records: "`round-6.md` — joint X1 round 8, two minor X1 issues, both shared with ADR-014."
4. **Ledgers.** Add a dated revision-9 HD9 history line (HD9 stays proposed), leave earlier lines unchanged and refresh the owned extracts. No new R-ID is needed.
5. **Optional advisories.**
   - List both revision-7 boundary projections as superseded in D9 and README l.106.
   - Record the reporter's own interpreter digest the next time the reporter changes. A change to the reporter would require a snapshot first and a create-once output.
6. **No executable change and no rerun.** The self-test, supplement, benchmark, calibration and projection bytes stay as they are.
7. **Create-once revision-9 checks.** Cover literal mirroring of both strings, unchanged executable digests and closure of every snapshot archive.
8. **Submit** `rev9-digests.json` and `round-8-response.md` for X1 round 9, to be saved also as ADR-014 `round-7.md`.

## 12. Stage authorisation

| Stage | Status now | Earliest path |
|---|---|---|
| Drafting, synthetic validation, and option G model-free engineering on synthetic repositories | **Permitted** (unchanged) | — |
| Preparation | **Not authorised**: 2 open issues (rule 1) | A zero-issue round on revision 9; then only P l.222 actions (a)–(e). The 45-day clock starts at the first such action. |
| Inventory, model-free scans and SRS draw (including dev-sample collection) | **Not authorised** | Sealed preparation outputs, a passing binding check with positive control, runtime closure, the test-repository identifier review, and a scoped zero-issue inventory review |
| Full OC and implementation (including GUIDE-author sessions) | **Not authorised** | All of:<br>• locked sizes;<br>• a binding CPU and memory fit under the corrected stop rule, with a reviewed locked-size command;<br>• a reviewed OC controller enforcing both counters and the frozen safety scheduling policy;<br>• a reviewed final evaluator;<br>• product-backed rule-4 frames;<br>• GUIDE and isolation records reviewed to zero issues before any dev label |
| Dev Jev and labels; test scoring and labels | **Not authorised** | Under full authority only |
| Gold unblinding and analysis | **Not authorised** | — |
| ADR-014 adoption | **Not authorised** | The joint zero-issue final round saved in its folder |
| Default, threshold or product-sampler change | **Not authorised** | A separate signed ADR |

**Next stage authorised:** none beyond drafting, synthetic validation and option G engineering on synthetic repositories.

**Sign-off statement:** none. Round 8 closes as follows:
- **X1: 2 issues open** (0 blocking, 0 major, 2 minor: X1-R8-01 and X1-R8-02).
- **ADR-014: 2 issues open** (X1-R8-01 and X1-R8-02, needing table changes at ADR l.45 and l.40). Adoption also waits for the joint zero-issue X1 round.

The reviewed files are those in `reviews/rev8-digests.json`, with manifest SHA-256 `8cbdb2114d79daea2d892da1cc294fe79ebfb305d56c37fd4c51d24bb1dd419e` as reported by Main. The per-file digests are in the table above. No preparation, inventory, draw, Jev, labelling or analysis stage is authorised, and ADR-014 is not adopted.
