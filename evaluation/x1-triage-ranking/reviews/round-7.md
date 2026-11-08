# X1: council experiment review, round 7 (revision 7 with ADR-014 round 5)

Date: 2026-11-08. Reviewer: innovation council. The lead wrote the synthesis. Two read-only peers ran in parallel:
- `innovation-challenger` (X1Council7.X1Council7Challenger);
- `experiment-peer` (X1Council7.X1Council7ExperimentPeer).

This record is X1 round 7 and ADR-014's fifth council review. Save it byte for byte as both `evaluation/x1-triage-ranking/reviews/round-7.md` and `docs/adrs/2026-11-08-x1-delegated-decisions/round-5.md`.

**Governance.**
- `CLAUDE.md` experiment rules 1–7.
- Design rules 1–3, for ADR-014.
- The ledger rules.
- The General sections "Toolchain", "Jev labelling", "Model use" and "Secrets".
- Standing owner delegation R-064: where an owner choice comes up, this record gives the council's recommended option as the fix. Wording preferences alone are not counted as issues.

**Scope.** Revision 7, as listed in `reviews/rev7-digests.json`. The manifest has 110 files and excludes itself. Each round-6 item was checked against the fix that `reviews/round-6.md` §5 required.

**What was and was not run.**
- No shell, model call, self-test, simulation, benchmark, disassembly or digest recomputation was run. No web lookup was made this round.
- The evidence comes from reading files, grep and hand arithmetic.
- `codegraph` was not used. Its reference documents (`readiness-and-bootstrap.md`, `query-selection.md`) were read first, but this read-only session had no way to run the executable.
- Main reports an independent self-test by a separate agent:
  - README entrypoint with pinned Python 3.11.11;
  - exit 0, 249 checks passed, 14.0 s;
  - planted effect supported, planted null not supported;
  - 110 of 110 digests matched before and after, and `git status` was unchanged.
  
  This round relies on that report. The report names only the self-test entrypoint. The supplementary `kernel_rounding.py` run was not independently repeated [INFERENCE].
- Paths are relative to `evaluation/x1-triage-ranking/` unless stated otherwise. "P l.N" means a line of revision 7's `PREREGISTRATION.md`. "M" means `oc-manifest.json`. "ADR l.N" means a line of `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md`.

**Reviewed files (SHA-256).** Main checked all 110 entries of `reviews/rev7-digests.json`. Main reports the manifest's own SHA-256 as `b108f52422c2724d95eb1dca2ced6bdd7da8245ec478af9888aae6864287f76d`. This round did not recompute any digest.

| File | SHA-256 |
|---|---|
| `docs/STATUS.md` | `4f811845662beb848e112ef25f01381254ba434885889d1891d7f996c7de6ae9` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md` | `fd724da98a0ce670a1c4ec862e4495f17474d3c330e98e9c780a45534516ce37` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-1.md` | `7bc229fd392215971646bfc547883050e82512c21777fbd311cf779e3e394710` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-2.md` | `5a298d77bb778d79c17d158657d868ccdadb0ce8a5eef9d01a8d170a22f976ed` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-3.md` | `9530480f20aa59b2c4c72047dfab46c263d73dd86e716a5a566fee4d54f87038` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-4.md` | `0682fcf9a04f0e84e70b3d56e6275bfcb17bb5d9463b519b0c39c2d3dcf26d28` |
| `docs/adrs/README.md` | `c277db8d8703c087b8c922693abcf1be24b7e0d5006d7798e6aef2b4492ec602` |
| `docs/hypothesis-ledger.md` | `8bda09c7dfa705e5fc754c0b23d9af1cb06d934a9503a709d912ac8957473910` |
| `docs/research-ledger.md` | `77d84f564ea1ac1727487d005e7f760820610002c8105784860d9bacd68d006c` |
| `docs/scout-design.html` | `f7180b872e7380b8c2a644b8c7fe33b90c9ccf3e3cc0a4c8c3662227dbb3d8a9` |
| `PREREGISTRATION.md` | `5db9e331e0e1bb1719b18a4f41daac99360caef60251b012858e7b1191db7c77` |
| `README.md` | `c9985161e7542e050f6ca4e224ea231918e40281360acd0d04a2f8bfeb234b6b` |
| `inputs/x11-rev4-digests.json` | `e9edba67c885623df6900a08aee1cde463e5fa78f828bbb04ca3d8efd7a0324c` |
| `oc-manifest.json` | `72e07be486cb2d4f04bd74247f073ceaf4f07b8ed11623d6660c7a55fc2d268f` |
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
| `reviews/README.md` | `d5edef1e8d340c6c0d3850393922e6c6b1d96c5eb2e99fddc291bf997b14ccb6` |
| `reviews/rev3-digests.json` | `894dc63b0522a23589187c6de9fa4e672878fc1c07b85bfbbdcd224e5ab2f448` |
| `reviews/rev4-digests.json` | `01c790837f961cd208fed98b3bd653b514da36e6d8bcee3b2fa4270f88e1876c` |
| `reviews/rev5-digests.json` | `ba1d2ae4efdf1c66e6d5f3002b5d0f892bac079a3efe4cbc2d98fd7826f69a95` |
| `reviews/rev6-digests.json` | `cca1baf432389b094588b6085888e8116442f0ee9baef15cbc750ef30086a746` |
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
| `validation/compute_benchmark.py` | `d39f40a28a6fd70698ea9e53e30448dec193c395d264002436dad9278ea6c90d` |
| `validation/fixed_population.py` | `ed670a6bf8076d791c06cfc2426aff03a0a2f9b1730982cf3fdfb2ec13bc49a2` |
| `validation/interval-diagnostic-rev4.json` | `50d41cc86c0342e05110d322b5d452ffcad7570c612efbbaee705ce11cc18a54` |
| `validation/interval-diagnostic-sources-rev4.json` | `d12d80627cecd6e7ee0b291dee939638cf7752c24ae62164bb495bf0ab693722` |
| `validation/interval-diagnostic-spec-rev4.json` | `e5a93a49c14a4345cba2fd6f2d4ddc55a998eaa199d371299d5965443fe54bb6` |
| `validation/interval_diagnostic.go` | `437f9c2559156f8befb613d3e17f19e2aad5437217081dbd4a1d0fb95c069417` |
| `validation/interval_diagnostic.py` | `a4b233374bdfec466bfc5862166497a4bbac48b13668aee2697fcb90e2b7c345` |
| `validation/kernel-rounding-rev7-final.json` | `c974259794fb559b00187767562fe0d552268aac3608da5544e6812bb2f26446` |
| `validation/kernel-rounding-rev7.json` | `133319e305f9c23f7c5e660860cd648afe6336c09d6fccb5fd9af930333d4ff9` |
| `validation/kernel-sources-rev6.json` | `66193e0fee0da8e8a2e386cac973329c81267a893657314bfcf0f42616579e22` |
| `validation/kernel_rounding.py` | `395fb9e08e92b042e7ab0c7cbfcd2a9baa88f615103e9f1012f4530e3d0ce8cf` |
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
| `validation/reproduce_reports.py` | `dd749c32c5ae2326266317edd66d91b9520338afe59fb58af50ba6255d6241d7` |
| `validation/rev5-reviewed-sources.json` | `fa6820c87e068c16b102e2b0d7647c31c6a8f58eccca8f3d3bb8b437fe2bf89d` |
| `validation/rev6-executed-sources.json` | `fafdcbc291a2521e1bf2b5f90ea43e7046135140c9bbb579c69ad41c4e9e6086` |
| `validation/rev6-reviewed-sources.json` | `33d0fc21db11c381d051a331e7c5db3a765889abba52a79b294336f080eac7cb` |
| `validation/rev7-executed-sources.json` | `13c805f30d5b8ce2442464e3a754f8df684664669a40659e211bf55d5f21a88a` |
| `validation/self_test.py` | `36d07d5648fdb0db325cb7583b02c4117cab1ada07b5fefe4590c66c8f03eb8a` |
| `validation/source-snapshot-checks-rev7.json` | `bfa5a445272441fa11cb95fe92a988a1cfd74c3dfbf71d61c37449aa2f21a5b1` |
| `validation/strong-effect-OC-rev5-closure.json` | `379f2c28ac5fa79ae738d2f1ad18acf823d389ff717616a0d8e37147a4e582fb` |
| `validation/strong-effect-OC-rev5-final.json` | `6f26bad6dcf9f9d7414757ae3ca73b68990187b3d3b52edadd856cb6613a4375` |
| `validation/strong-effect-OC-rev5-optimized.json` | `19c78ee133a87c3f05a9319e58b7bff59031e886837733edb668c566d945615b` |
| `validation/strong-effect-OC-rev5.json` | `089941397a9036830532ffe52410f3f3ca65a153db4e6cfa3e57ec2960ddbb8b` |
| `validation/strong-effect-OC-rev6.json` | `96cbbeaae60431010a97964a39ced7522d2c80d90eeb81d49ebc69945355b476` |
| `validation/strong-effect-OC-rev7.json` | `da5cc7ad3a833788407046ff6317e0d5e8bcd1acf0098d6b642d1b1a771d1b41` |
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
| `validation/synthetic-results.json` | `e65ec36c7d57a0d15f32bc3d2d7bd5f1f76f4148189ec2676653ef996901a350` |
| `validation/synthetic-sources-rev5-initial.json` | `a77aff4f00d87a5d98dafe3114df227b9de27ffb215c45a82ef6074781aee4c2` |

**Byte-preservation checks against the round-6 table.**
- **Unchanged (72 files):**
  - every earlier council round and response;
  - ADR-014 `round-1.md` to `round-3.md`;
  - `rev3-digests.json`, `rev4-digests.json`, `rev5-digests.json` and `inputs/x11-rev4-digests.json`;
  - `docs/adrs/README.md` and `docs/scout-design.html`;
  - from preparation: `exposure-sources.json`, `frozen-exposure-inputs.json`, `model-free-policy.json`, `pins.json`, `runtime-attestation.json` and `runtime-attestation-rev6.json`;
  - every revision-4, revision-5 and revision-6 validation output, reporter snapshot and source snapshot, including `kernel-sources-rev6.json` and `rev6-executed-sources.json`;
  - `interval_diagnostic.py` and `outer_simulation.py`.
- **Changed mutable documents (13):**
  - `docs/STATUS.md`, ADR-014 `ADR.md`, both ledgers;
  - `PREREGISTRATION.md`, `README.md`, `oc-manifest.json`, `reviews/README.md`;
  - `preparation/README.md`, `dev-census.json`, `exposure-recipe.json`, `guide-author-isolation.json` and `model-free-binding-check.json`.
- **Changed executables (6):**
  - `compute_benchmark.py` (`5c2557be…` → `d39f40a2…`);
  - `fixed_population.py` (`090e9348…` → `ed670a6b…`);
  - `interval_diagnostic.go` (`27c79a31…` → `437f9c25…`);
  - `ranking.py` (`9fc05fa5…` → `8bc2e7cb…`);
  - `reproduce_reports.py` (`4942de74…` → `dd749c32…`);
  - `self_test.py` (`06e1791f…` → `36d07d56…`).
  
  Each old digest is a `sha256` field in `validation/rev6-reviewed-sources.json`. Each new digest is in `validation/rev7-executed-sources.json`. `source-snapshot-checks-rev7.json` reports that content matches its SHA-256 for all five archives (11, 4, 6, 11 and 11 entries).
- **New files (19):**
  - ADR-014 `round-4.md`, `rev6-digests.json`, `round-6.md`, `round-6-response.md`;
  - `calibration-overhead-rev7.json`, `compute-benchmark-rev7.json` and the three `compute-projection-rev7*.json`;
  - `kernel-rounding-rev7.json`, `kernel-rounding-rev7-final.json`, `kernel_rounding.py`, `kernel_rounding_test.go`;
  - `optimization-refusal-rev7.json`, `rev6-reviewed-sources.json`, `rev7-executed-sources.json`, `source-snapshot-checks-rev7.json`;
  - `strong-effect-OC-rev7.json` and `synthetic-results-rev7.json`.
  
  72 + 13 + 6 + 19 = 110, and 91 + 19 = 110.
- **Identical pairs:**
  - `round-6.md` = ADR-014 `round-4.md` = manifest `joint_previous_record_sha256` (`0682fcf9…`).
  - `rev6-digests.json` = `cca1baf4…`, the manifest digest Main reported in round 6.
- **Cross-references:**
  - The projection's `benchmark_sha256` (`f268997d…`), `calibration_overhead_sha256` (`415e3ac6…`), `reporter_sha256` (`dd749c32…`) and `snapshot_sha256` (`13c805f3…`) equal the manifest entries.
  - The kernel-rounding output's source hashes (`437f9c25…`, `395fb9e0…`, `33bde17b…`, `8bc2e7cb…`) equal the manifest entries.
  - `optimization-refusal-rev7.json` records `ranking_sha256` `8bc2e7cb…`.

**Verdicts.**
- **X1: 5 issues open (0 blocking, 0 major, 5 minor). No sign-off.**
- **ADR-014: 3 issues open (X1-R7-02, -03 and -04).** Each needs a change to ADR-014's table, and each is shared with X1. **No sign-off.** ADR-014 can be adopted only on the joint zero-issue final round.
- **Joint signed set: 5 distinct issues open.**
- No preparation, inventory or draw, Jev, labelling or analysis stage is authorised (§12).

**Severity scale.** Unchanged from rounds 1–6:
- **blocking:** can produce a wrong or unreachable verdict, or breaks experiment rules 1–7 or REQ-169.
- **major:** misleads the result, leaves a preregistered quantity undefined, or makes the study unrepeatable or likely not testable.
- **minor:** text, consistency or provenance. Minor issues must still be fixed before sign-off (rule 1).
- **Calibration:** a fail-closed default that only makes a gate harder to pass is not a defect unless it makes the gate unreachable for a compliant case.
- **Undefined rules counted as minor.** X1-R7-02 to -04 leave procedural details of stop or reporting rules undefined. They are graded minor, not major, because:
  - none can change the verdict;
  - the two stop rules fail closed;
  - the undefined part can only make passing easier or reporting incomplete.

---

## 1. Recommendation

**Revise to revision 8, closing X1-R7-01 to X1-R7-05, and submit it for X1 round 8 / ADR-014 round 6. Do not sign off. Authorise no stage beyond drafting, synthetic validation and option G model-free engineering on synthetic repositories.**

All ten round-6 items are fixed as required (§4). The revision-7 resource arithmetic reproduces exactly (§8).

The drafter's open question is the 12 M boundary. U = 12 M is within the registered population envelope but exceeds the binding CPU fit. **This is not an issue.**
- Any population larger than the binding break-even (≈ 9.68 M at p_gate = .5) fails the preregistered sampling gate (ΣN ≤ 60,000) almost surely, so p̂_gate ≈ 0 and the study is already not testable on that ground (§6).
- At U = 12 M exactly, every run has 2,000 units, ΣN = 120,000, and p̂_gate = 0.
- The binding CPU stop therefore removes no case that the sampling gate would leave testable.
- The text makes no promise that the whole registered range fits (P l.181: "Neither historical inference nor registration guarantees fit").

The five open issues are minor and fall into three groups:
- **Evidence provenance:** the supplementary kernel-rounding run lacks runtime and binary closure, and its superseded predecessor's source was neither archived nor labelled (X1-R7-01).
- **Locked-fit procedure:** the memory threshold is undefined (X1-R7-02), and the remeasurement procedure is undefined despite a 30–66 % run-to-run CPU drift (X1-R7-03).
- **Reporting and ledger:**
  - no fixed order or status rule for a safety grid left partial at gold (X1-R7-04);
  - the revision-6 HD9 history line still claims a memory envelope (X1-R7-05).

Keep all of the following:
- the binding-only stop and the separate total-planning envelope;
- safety charged at 2,000 executions;
- explicit Go `float64()` rounding and the disassembly guard;
- the non-strippable calibration refusal and optimization refusal;
- the canonical keyed strings;
- the dev eligibility rules;
- the stub-CLI positive control;
- the three-session GUIDE bound;
- source snapshots before and after modification;
- the stage table.

## 2. Alternatives considered

| Option | Description | Tradeoff | Council position |
|---|---|---|---|
| A. Sign off revision 7 and carry the minors | Starts the clock sooner | Rule 1 requires zero issues. X1-R7-02 and -03 sit in a binding stop rule. | Rejected |
| B. **Revision 8 closes the five minors; then a zero-issue round** | Text changes, one supplementary rerun, and one arithmetic-only projection from the existing rev7 benchmark | One more round | **Recommended** |
| C. Authorise preparation now, since no open issue touches actions (a)–(e) | Starts fetch and attestation | Rule 1 applies to the whole reviewed preregistration, and the clock would start under a protocol that is still changing | Rejected |
| D. Restrict the registered envelope to U ≤ 9.68 M, or raise caps or shrink the grid, because of the 12 M result | Removes the apparent gap | Unneeded: the sampling gate already excludes those populations (§6). It also changes the grid or caps, which round 6 forbade. | Rejected |
| E. For X1-R7-01, fold the supplement into `self_test.py` | One provenance convention | Changes the 249-check self-test bytes and forces a full rerun | Acceptable. The recommended option is to give the supplement the self-test's provenance pattern. |
| F. For X1-R7-02, use "available memory" measured at launch | Reflects host state | Depends on host load, so it is itself a forking path | Rejected in favour of a fixed reserve |
| G. For X1-R7-03, use the median of k runs | Less sensitive to noise | A rerun could still help a failing fit pass | Rejected in favour of one planned run, with the maximum over all completed runs binding |
| H. For X1-R7-04, require safety to finish before gold | Complete descriptive grid | Lets descriptive work gate the study, against P l.171 and round 6's X1-R6-04 | Rejected in favour of a keyed order and an explicit "not completed" status |
| I. Option G: finish model-free engineering before preparation (OC controller enforcing both counters, final-evaluator draft, binding harness on synthetic repositories) | Shortens the clocked path. The throughput margin is tight (§9). | Work before the clock, on no real data | **Recommended practice**, not a sign-off condition |

## 3. Rule checks

- **Rules 1–2.** This round is recorded with digests. Issues are open, so iteration continues.
- **Rule 3.** No predictor–outcome association was computed on real data. All revision-7 evidence is synthetic. The historical method selection was not rerun (manifest `historical_method_selection_rerun: false`). The dev size-distribution report (P l.192) is distributional only.
- **Rule 4.**
  - Planted effect: 2,000/2,000 support, Wilson MC95 lower .9980829527187469, realised gain .463, zero inadequate (`strong-effect-OC-rev7.json`).
  - Planted null: not supported (independent run reported by Main).
  - New heterogeneous-T differential, mismatched-gain mutant and optimization refusal.
  - The supplement's provenance gap (X1-R7-01) weakens one piece of compiled-parity evidence.
- **Rule 5.** Nothing is signed yet. The revision-6 projection is preserved and labelled superseded. Every changed reviewed executable was archived before modification. One new script's pre-change bytes were not kept (X1-R7-01(b)).
- **Rule 6.** Both envelopes, the boundary projections at U = 6 M and 12 M, the memory envelope and the host memory are reported. A partial safety grid has no reporting rule (X1-R7-04).
- **Model use and secrets.**
  - No model was called.
  - The binding recipe checks the key by presence only, and its positive control uses stubs, never a real CLI.
- **Toolchain.**
  - Numerical evidence ran on `.tools/venv/bin/python` 3.11.11, SHA-256 `570b4105…`, and `go1.26.10 darwin/arm64`.
  - `toolchain/versions.env` is unchanged (manifest `toolchain_pin_changed: false`).
- **Design rules.**
  - ADR-014 has a sequential ID and an adopted-form status matching ADR-011.
  - Its Review section defers to the final round in its folder, and its Records list includes `round-4.md`.
  - The index and STATUS are correct.
  - Three issues require table changes.
- **Ledgers.**
  - R-086 is present (`docs/research-ledger.md:85`), and the next free ID is R-087 (`docs/STATUS.md:128`).
  - HD9 lists R-085 and R-086 and remains proposed (`docs/hypothesis-ledger.md:36`).
  - The revision-6 history line is inaccurate (X1-R7-05).

## 4. Verification of round-6 fixes

| Issue | Status | Evidence |
|---|---|---|
| X1-R6-01 | **Fixed** | ADR l.48 reads "M = 2,000 with no binding cap, failure 0, fully annotated". A re-grep of P, README, M, ADR and the recipes found only historical cell IDs (README l.26–47). |
| X1-R6-02 | **Fixed.** Provenance residual in X1-R7-01 | **Explicit rounding:** explicit `float64()` on every product that feeds an addition or subtraction:<br>• `fit` (`interval_diagnostic.go:11`);<br>• the moment terms (`:23`);<br>• `endpoint` (`:30`);<br>• the population accumulations (`:86–91`).<br>The only remaining products feed divisions, `Floor` or `Sqrt` and cannot fuse.<br>**Differential fixtures:**<br>• T = 10/30 with inexact ratio .2 and mixed resamples gives hex-identical Go complete and vertex intervals (`self_test.py:259–272`);<br>• Go classification test (`kernel_rounding_test.go:5–15`);<br>• actual Python vertex and complete paths, with 6,000 zero-variance classifications (`kernel-rounding-rev7-final.json`).<br>**Disassembly:** recorded (`synthetic-results-rev7.json`, `fit_disassembly`) and guarded (`self_test.py:273–275`).<br>**Text:** comment, M `zero_variance_rule` and ADR l.26 say "identical after explicit float64 rounding". |
| X1-R6-03 | **Fixed** | `raise RuntimeError` (`fixed_population.py:192–193`). The mismatched-gain mutant must raise (`self_test.py:415–425`). `sys.flags.optimize` is recorded and non-zero values are refused (`ranking.py:393–396`; `compute_benchmark.py:233–234`; `reproduce_reports.py:339–340`). The PYTHONOPTIMIZE=1 refusal exits 1 (`optimization-refusal-rev7.json`). P l.169 and ADR l.33 agree. |
| X1-R6-04 | **Fixed** | (i) Safety is charged at 2,000 executions (`reproduce_reports.py:224–227`), and the 4,000 cap is inert as a guard (P l.171; M `counter_rule`).<br>(ii) `fits_caps` = binding hours ≤ 2,880 and the memory fit (`:319`).<br>(iii) Safety uses remaining capacity and is unmeasured if unfinished by gold (P l.171, l.183).<br>(iv) Both envelopes are published (`binding_break_even_envelope`; `break_even_envelope`, scoped "never a testability stop").<br>Revision 6 is preserved as superseded. ADR l.40 and l.45 agree. Arithmetic in §8. |
| X1-R6-05 | **Fixed** | Host memory comes from `sysctl -n hw.memsize` (`compute_benchmark.py:260`): 68,719,476,736 B, also in M `observed_host_physical_memory_bytes`. Slope, intercept and break-even (`reproduce_reports.py:251–255`) are published in P l.181, README l.102 and ADR l.45, marked [INFERENCE]. The locked-fit definition is X1-R7-02, and the ledger residual is X1-R7-05. |
| X1-R6-06 | **Fixed** | Only first-occurrence, repeated, pre-existing-at-C0_dev and unknown-attribution exclusions apply. Test freshness/exposure and run-size eligibility do not. Examined units are flagged and counted, with current-guide labels only. The greedy skip rule and the source-commit size report are stated (P l.192; `dev-census.json:24–28`; `exposure-recipe.json` step 6). Action (d) is test-only (P l.220). The §12 threats cover unequal inclusion and the 2023 vs 2024+ era gap (P l.238). ADR l.22 agrees. |
| X1-R6-07 | **Fixed** | Canonical strings are defined in P l.39, M `keyed_order_strings` and `dev-census.json:26`. One fixture per key has an expected digest recorded as computed with `openssl dgst -sha256 -hmac` (`self_test.py:122–127`). That derivation is author-recorded. |
| X1-R6-08 | **Fixed** | Synthetic-only unbound positive control: key absent, binding deliberately unapplied, logging non-zero-exit stubs on PATH and at configured paths, starts captured and interlock failure required, enabled policy sealed, and the bound run must show no starts (P l.221; `model-free-binding-check.json:14, :19`; ADR l.61). Marked unexecuted. |
| X1-R6-09 | **Fixed** | At most three fresh sessions under stage 3, each receiving the packet, the previous draft and verbatim issues with no data. All are logged before labels, an exposed replacement uses a slot, and exhaustion means not testable before dev (P l.222; `guide-author-isolation.json:5, :21–22`; ADR l.60). No "exactly one" wording remains. |
| X1-R6-10 | **Fixed** | R-085 (and R-086) are in HD9's sources, and HD9 remains proposed (`docs/hypothesis-ledger.md:36`). The owned-extract digest has been refreshed (`rev7-digests.json`, `owned_extracts_sha256`; digest verification as reported by Main). |

## 5. New issues

### X1-R7-01: minor. The supplementary kernel-rounding evidence lacks runtime and binary closure, and its superseded predecessor is unlabelled and its source unarchived

- **Location:**
  - `validation/kernel_rounding.py:19–22, 53–64`;
  - `validation/kernel-rounding-rev7-final.json` (`provenance`);
  - `validation/kernel-rounding-rev7.json:2, :34`;
  - `README.md:13`; `round-6-response.md:8`;
  - P l.179; ADR l.59.
- **Problem:**
  - **(a) Missing closure.** Only this supplement exercises the actual Python vertex path and the Go zero/non-zero classification that X1-R6-02 required. Its provenance (`ranking.provenance`, `ranking.py:389–398`) records only the Python version, the optimization flag and source hashes. Unlike `self_test.py:102–113`, it:
    - does not refuse a non-3.11.11 interpreter;
    - does not record the interpreter path or digest, the Go version, the digest of the built interval kernel, or the identity of the compiled `go test` binary.
    
    This falls short of P l.179 ("source/binary/runtime closure") and ADR l.59 ("path/digest in every new output"). `checks_passed` is also a hard-coded literal (`6`), which is how the predecessor misreported `7`.
  - **(b) Superseded predecessor.**
    - The superseded `kernel-rounding-rev7.json` was produced by `kernel_rounding.py` `d6d4701e…`. Those bytes exist nowhere in the worktree (grep); round 6 §11 step 7 required a snapshot of every executable before modification.
    - The supersession is stated only in the author response, not in the README or the manifest.
    - The scientific fields of the two outputs agree (peer spot check).
  - **Why minor.** No numerical error is alleged. The main self-test, which has full closure, covers the Go paths.
- **Required fix** (recommended option):
  - Give `kernel_rounding.py` the self-test's provenance pattern:
    - refuse versions other than 3.11.11;
    - record the interpreter path and SHA-256 and the `go version`;
    - build the classification test with `go test -c -o <tmp>`, hash it and run it;
    - hash the built interval kernel;
    - count executed checks instead of using a literal.
  - Snapshot `kernel_rounding.py` and `kernel_rounding_test.go` before modification and after execution.
  - Produce a create-once `kernel-rounding-rev8.json`.
  - Label both revision-7 kernel-rounding outputs superseded, with reasons, in the README and a manifest supersession list.
  - Disclose that the `d6d4701e…` bytes were not retained, as X1-R5-06 did.

### X1-R7-02: minor. The locked memory fit, part of the binding stop, has no operational definition

- **Location:**
  - P l.179 ("failed or unmeasured binding-critical-path runtime/memory fit means not testable") and l.181 ("available-memory remeasurement still binds");
  - M `locked_runtime_and_memory_gate` ("locked available-memory capacity") and `final_evaluator_contract` ("runtime and memory limits");
  - `reproduce_reports.py:255, :319` (physical memory ÷ 8, no reserve, calibration slope only);
  - ADR l.45.
- **Problem:**
  - "Available memory" has no defined measure, operating-system reserve, RSS phase set or worker rule.
  - The final evaluator would therefore fix the threshold after the locked sizes are known. That is a forking path in a stop rule.
  - The planning `fits_caps` uses a different, reserve-free test.
  - **Why minor.** Plausible locked U leaves a wide margin (U = 6 M needs about 32.3 GB for 8 workers on a 68.7 GB host [INFERENCE]), and the fit fails closed.
- **Required fix** (recommended option):
  - **Rule.** Locked memory fit = 8 × max(locked-size peak RSS of construction, calibration, and trial worker plus helper), using the existing lifetime high-water method, must be ≤ `hw.memsize` − 8 GiB. On this host that is 60,129,542,144 B.
  - **Workers.** Fixed at 8, with no reduction.
  - **Failure.** Failed or unmeasured means not testable.
  - **Planning test.** Apply the same reserve in the planning test, and publish the reserve-adjusted break-even (about 11.22 M units [INFERENCE]). Produce it as a create-once revision-8 projection computed only from the existing `compute-benchmark-rev7.json` and `calibration-overhead-rev7.json`, with no new timing run.
  - **Where to change it:** P §8, M, README and ADR l.45.

### X1-R7-03: minor. The locked-size remeasurement procedure is unspecified, although measured CPU drifts by 30–66 % between runs

- **Location:** P l.179 ("Remeasure these costs, RSS and both hard counters on locked sizes"); M `preparation_gate`, `locked_runtime_and_memory_gate`; ADR l.45.
- **Evidence:** same host, same benchmark specification, and executable changes not on the timed paths. Between `compute-projection-rev6.json` and `compute-projection-rev7.json` (`measured_inputs`):

  | Input | Revision 6 | Revision 7 | Change |
  |---|---|---|---|
  | Construction CPU per unit | 4.13415e−7 | 6.872933e−7 | +66.2 % |
  | Re-evaluation CPU per unit | 3.271078e−6 | 4.263523e−6 | +30.3 % |
  | Gate check CPU | 1.455533e−5 | 1.911161e−5 | +31.3 % |
  | Complete trial CPU | 1.0558725e−3 | 1.1410945e−3 | +8.1 % |

  - Revision-6 inputs under revision-7 accounting give exactly round 6's binding per-unit value, 1.97202e−4. So the move from about 12.65 M to 9.68 M comes entirely from timing drift, not from the accounting.
  - The cause is probably host load [INFERENCE].
- **Problem:**
  - Nothing fixes how many locked runs are made, which statistic binds, or whether a failing run may be repeated.
  - "Remeasure until it fits" is therefore open at a binding stop.
- **Required fix** (recommended option):
  - **One run.** Locked remeasurement is one create-once run of the reviewed benchmark and calibration commands at the locked sizes, with the registered replicate structure.
  - **Reruns.** A rerun is allowed only after a recorded integrity failure (crash, closure mismatch or host fault), with its reason recorded. A failed fit is never a reason to rerun.
  - **Reporting.** Report every completed run. Each binding input is the maximum over all completed runs.
  - **Host state.** Record host load (load averages and the count of concurrent processes) in provenance.
  - **Disclosure.** Disclose the revision-6 to revision-7 drift in P §12 and the README.
  - **Where to change it:** P §8, M and ADR l.45.

### X1-R7-04: minor. A safety grid left partial at gold has no preregistered execution order or status rule

- **Location:** P l.171 ("Finish before gold or report unmeasured"), l.183, l.222; M `final_safety_OC.schedule`, `attempt_and_execution_projection`; ADR l.40.
- **Problem:**
  - The total-planning fit fails above about 5.66 M units, and safety overlaps labelling under the 45-day clock, so a partial grid is foreseeable.
  - With no fixed order, which profiles are measured can be chosen during the run, possibly after safety outputs have been seen.
  - "Unmeasured" also collides with the trial-level outcome category "safety-unmeasured" (P l.171).
  - This is a reporting risk under rule 6 for descriptive results. It cannot change the verdict.
- **Required fix** (recommended option):
  - **Order.** Fix the execution order before the first safety output: ascending HMAC-SHA-256 with UTF-8 key `169-x1-safety-order` over the canonical `safety-interval-group` identity. Run the canonical-criticality phase per group in the same order.
  - **Measured groups.** A group counts as measured only when all five populations reach the 2,000 stop. Attempt exhaustion keeps its existing "unmeasured; cell fails" meaning.
  - **Unfinished groups.** Report every unstarted or unfinished group at gold freeze as "not completed by gold": a separate status, with counts, partial counts shown separately and excluded from pooled Wilson summaries.
  - **No reordering** after the first safety output.
  - **Where to change it:** P l.171, M and ADR l.40.

### X1-R7-05: minor. The revision-6 HD9 history line still credits revision 6 with a memory envelope

- **Location:** `docs/hypothesis-ledger.md:41`: "worst-case enumeration and L400 RSS projection and memory envelope".
- **Problem:**
  - Round 6 (X1-R6-05) found that revision 6 published no memory envelope.
  - Its fix wording ("reword the ledger line to 'L400 RSS projection and memory envelope' once the envelope exists") was applied to the revision-6 line, so the inaccurate claim remains.
  - The revision-7 line (l.40) is correct. The council acknowledges that its own round-6 wording caused this.
- **Required fix:**
  - With Main's permission for this unmerged row, change line 41 to "… worst-case enumeration and L400 RSS projection …", without the memory envelope.
  - Leave line 40 unchanged.
  - Refresh any owned extract that covers HD9 history.

## 6. Assessment of the targeted questions

- **U = 12 M and the binding CPU fit** (drafter's note). Not an issue. The council's derivation:
  - Every L_r ≤ 2,000 (P l.41), so the expected sampled total satisfies E[ΣN] = 20·Σ_r U_r/L_r ≥ U/100. Above the binding break-even, U > 9,675,589 gives E[ΣN] > 96,756.
  - With unit counts in [20, 2,000], Hoeffding's bound (which also holds for sampling without replacement, independently across repositories) gives P(ΣN ≤ 60,000) ≤ exp(−2·36,756²/(60·1,980²)) = exp(−11.49) ≈ 1.0e−5.
  - So p_gate ≤ 1.0e−5. The attempt cap is then at least 4 × 10⁸, and the binding gate term alone is at least 4 × 10⁸ × 9.647e−4 ≈ 3.9 × 10⁵ core-hours. Alternatively, 100,000 redraws give p̂_gate = 0. Either way the study is not testable regardless of the CPU stop.
  - At U = 12 M every run has N = 2,000 and ΣN = 120,000, so p̂_gate = 0 exactly (challenger).
  - Populations with p_gate near .5 need a mean run size near 1,000, so U ≤ about 6 M. The binding fit holds there: 1,939.74 h ≤ 2,880 h.
  - The stop is calibrated and fails closed. Remeasured locked costs must still decide the outcome (X1-R7-03).
- **Binding-only stop.** It is correct and consistently implemented (§4, §8). The total envelope is planning-only.
- **Memory.** No gate is unreachable for a reachable U: the break-even is 12.83 M with no reserve and about 11.22 M with an 8 GiB reserve, both above any gate-feasible U. The only defect is the missing definition (X1-R7-02).
- **Dev change.** The keyed order is independent of whether a unit was examined, so examined units enter at their natural rate. No new confound for c_U, M̂, r_L or the throughput bound beyond the disclosed commit-balanced mixture and era gap (peer).
- **GUIDE bound.** Consistent with the stage table and with P l.222.
- **Stage-3 prerequisites carried from round 6.** These are not issues at this stage:
  - an executing OC controller that enforces both counters (the Go population path still loops `for range q.Trials`);
  - a reviewed final evaluator;
  - a product-backed rule-4 harness.
- **ADR-014.**
  - The status matches adopted ADR-011, and the Records list is current.
  - The table matches P on every round-6 fix.
  - The design card and index are unchanged and still accurate.
  - Three issues (X1-R7-02, -03, -04) need table edits at l.40 and l.45.

## 7. Checked and sound

- **Cell counts** agree across P, M, the projection and the ADR:
  - 17,920 complete/reference and 252 missing cells;
  - 336 safety profiles, 8,064 outcomes, 2,016 interval groups;
  - 11,900 calibrated populations;
  - 90,860 binding constructions;
  - 80,640 requested and 40,320 canonical evaluations.
- **Go fusion.** Every multiply feeding an add or subtract is wrapped in `float64()`. The recorded `main.fit` disassembly shows separate `FMULD` and `FSUBD`/`FADDD`.
- **Fixture sensitivity.** In the heterogeneous fixture, Python rounds fl(0.2)×30 to exactly 6.0. A fused kernel would give a residual of about −3.3e−16, so the fixture is sensitive to fusion.
- **Snapshot discipline.**
  - All six changed reviewed executables are archived with their round-6 digests (`rev6-reviewed-sources.json`).
  - The executed bytes are sealed (`rev7-executed-sources.json`).
  - All five archives close on content and hash.
- **Interpreter.** Explicit `.tools/venv/bin/python` 3.11.11 appears in every README command. `synthetic-results-rev7.json` records the interpreter digest, the Go version and the binary digest.
- **Dev window.** The C0_dev and root edge cases are unchanged and well defined.
- **No over-authorisation.** P l.220–224, the recipes and ADR l.15 grant no stage.

## 8. Evidence index

- **Binding envelope** (`compute-projection-rev7.json`, `reproduce_reports.py:234–250`):
  - fixed hours = 227.20459377777783 + 133.9095352 + 41.96692344444444 × 10,220/11,900 (= 36.04218131) = 397.1563102888889;
  - per unit = (90,860 × 6.872933e−7 + 80,640 × 4.2635225e−6 + 10,220 × 5.3035817e−6) × 2/3,600 = 2.558114e−4;
  - gate per attempt = 90,860 × 1.911161e−5 × 2/3,600 = 9.647116e−4.
- **At U = 1.2 M, p_gate = .5:**
  - binding 397.1563 + 306.9737 + 7.7177 = 711.8477 h, which ÷ 192 = 3.7075 days;
  - total 1,288.4371 h (6.7106 days);
  - safety 576.5894 h, which includes the 2,000-execution charges of 2.1300 + 446.3651 h.
- **Break-even U at p_gate = .5:**
  - binding (2,880 − 397.1563 − 7.7177)/2.558114e−4 = 9,675,588.9;
  - total 5,659,336.2.
- **Boundary projections:**
  - U = 6 M: binding 1,939.7424 h (passes), total 3,001.5849 h;
  - U = 12 M: binding 3,474.6109 h, total 5,143.0196 h (`-L2000.json`, `-U12000000.json`).
- **Memory** (`calibration-overhead-rev7.json`):
  - slope (831,799,296 − 111,263,744)/1,080,000 = 667.1625 B per unit; intercept 31,204,238 B;
  - 8-worker break-even (68,719,476,736/8 − intercept)/slope = 12,828,553;
  - with an 8 GiB reserve, (7,516,192,768 − 31,204,238)/667.1625 ≈ 11,219,138 [INFERENCE].
- **X1-R7-03 drift:** the ratios in §5. The revision-6 inputs under revision-7 binding accounting give 1.97202e−4 per unit.
- **X1-R7-01:**
  - `kernel_rounding.py:21` (`ranking.provenance` only), `:22` and `:55` (unrecorded `go test` and `go build`), `:62` (literal count);
  - `kernel-rounding-rev7.json:34` (`d6d4701e…`, not archived).
- **§6 bound:** E[ΣN] ≥ U/100; Hoeffding exp(−11.49) ≈ 1.03e−5.

## 9. Risks

- **The throughput screen is the tightest constraint** (peer). The minimum labelling bound is 31.09 s per reserved case. If the labelled-case reservation approaches the 60,000-unit sampling cap, that is about 21.6 days. At U = 6 M (10.10 binding OC days), only about 6.3 days would remain for preparation and inventory [INFERENCE]. The screen is preregistered and fails closed. Option G reduces exposure.
- **Timing noise.** A locked fit near the boundary could flip with host load. X1-R7-03 makes the outcome conservative and auditable.
- **Stage-3 artefacts not written:** the OC controller, the final evaluator and the product harness (§6).
- **Implementing the binding** (action (c)) may touch product code. If so, those digests join the seals and the pinned Scout identity reviewed at inventory.
- **Shared-file staleness.** Whole-file digests of STATUS, the ledgers and the index go stale if a sibling lands first. The owned extracts mitigate this.
- **Reversibility.** All fixes are pre-data text or synthetic changes. No product setting, default, accepted ADR or requirement changes.

## 10. Where the peers changed the council's recommendation

- **Challenger:**
  - Found the supplement's missing runtime and binary closure (**X1-R7-01(a)**).
  - Supplied the U = 12 M ⇒ ΣN = 120,000 ⇒ p̂_gate = 0 argument. This settled that the drafter's 12 M note is not an issue; the lead generalised it with the Hoeffding bound (§6).
  - Independently reproduced every envelope and memory figure.
  - Rated the disassembly guard's scope (main.fit only; no FNMADDD or single-precision opcodes) as optional hardening. **Adopted as advice**, because explicit conversion already guarantees the property in source.
- **Experiment peer:**
  - Found the undefined memory fit (**X1-R7-02**), the partial-safety ordering gap (**X1-R7-04**) and the unarchived, unlabelled predecessor (**X1-R7-01(b)**). The lead had found none of these.
  - Independently found the timing drift (**X1-R7-03**), with the decomposition showing that the change to the break-even is entirely due to timing.
  - Independently found the HD9 revision-6 line (**X1-R7-05**).
  - Raised the throughput-screen risk (§9).
- **Lead:**
  - Independently found X1-R7-03 and X1-R7-05.
  - Merged the challenger's and peer's kernel-rounding findings into one issue with a single fix.
  - Decided not to raise:
    - the missing comma between R-081 and R-085 in ADR "Informed by" (editorial; fix it while editing l.40/45);
    - the "12 M fails CPU" phrasing in P l.181 and README l.100 (true under the stated p_gate = .5 assumption; optionally add that p̂_gate = 0 would apply first);
    - the HMAC fixtures using Python's `hmac` against literals (meets X1-R6-07 as written).

## 11. Execution outline for revision 8 (one integration owner)

1. **Supplement provenance (X1-R7-01).**
   - Snapshot `kernel_rounding.py` and `kernel_rounding_test.go`.
   - Add the version refusal, interpreter path and digest, `go version`, kernel and test-binary digests, and counted checks.
   - Produce a create-once `kernel-rounding-rev8.json`.
   - Label both revision-7 outputs superseded, and disclose the loss of `d6d4701e…`.
   - *Test:* six or more checks pass, and every closure field is present.
2. **Locked memory fit (X1-R7-02).**
   - Define the 8-worker maximum-phase RSS ≤ `hw.memsize` − 8 GiB rule.
   - Apply the reserve in the reporter (snapshot it first), and generate a create-once revision-8 projection from the existing rev7 benchmark and calibration only.
   - *Test:* `fits_caps` uses the reserve, and the break-even is about 11.22 M.
3. **Locked remeasurement (X1-R7-03).**
   - Register one planned run, integrity-only reruns, the maximum over all completed runs, and host-load provenance.
   - Disclose the drift in §12 and the README.
4. **Safety partial grid (X1-R7-04).**
   - Add the `169-x1-safety-order` keyed order, the all-five-populations completion rule, the "not completed by gold" status, and no reordering.
5. **Ledger (X1-R7-05).**
   - Correct `hypothesis-ledger.md:41` with Main's permission, and refresh the extracts.
6. **ADR-014.**
   - Update l.40 (X1-R7-04) and l.45 (X1-R7-02, -03).
   - Add the missing comma in "Informed by".
7. **Do not rerun** the self-test, benchmark or calibration unless their bytes change. Any rerun uses `.tools/venv/bin/python` 3.11.11 with create-once outputs and snapshots before modification.
8. **Submit** `rev8-digests.json` and the response for X1 round 8, to be saved also as ADR-014 `round-6.md`.
9. **Main checks before round 8:**
   - recompute `content → sha256` for every snapshot archive, including the new supplement snapshot;
   - check that the revision-8 projection's inputs are byte-identical to the revision-7 benchmark and calibration outputs.

## 12. Stage authorisation

| Stage | Status now | Earliest path |
|---|---|---|
| Drafting, synthetic validation, and option G model-free engineering on synthetic repositories | **Permitted** (unchanged) | — |
| Preparation | **Not authorised**: 5 open issues (rule 1) | A zero-issue round on revision 8; then only P l.220 actions (a)–(e). The 45-day clock starts at the first such action. |
| Inventory, model-free scans and SRS draw (including dev-sample collection) | **Not authorised** | Sealed preparation outputs, a passing binding check with positive control, runtime closure, the test-repository identifier review, and a scoped zero-issue inventory review |
| Full OC and implementation (including GUIDE-author sessions) | **Not authorised** | Locked sizes; a binding-only CPU and memory fit under the X1-R7-02/03 rules; a reviewed OC controller enforcing both counters; a reviewed final evaluator; product-backed rule-4 frames; GUIDE and isolation reviewed to zero issues before any dev label |
| Dev Jev and labels; test scoring and labels | **Not authorised** | Under full authority only |
| Gold unblinding and analysis | **Not authorised** | — |
| ADR-014 adoption | **Not authorised** | The joint zero-issue final round saved in its folder |
| Default, threshold or product-sampler change | **Not authorised** | A separate signed ADR |

**Next stage authorised:** none beyond drafting, synthetic validation and option G engineering on synthetic repositories.

**Sign-off statement:** none. Round 7 closes as follows.
- **X1: 5 issues open** (0 blocking, 0 major, 5 minor: X1-R7-01 to X1-R7-05).
- **ADR-014: 3 issues open** (X1-R7-02, -03 and -04, each needing ADR-014 table changes at l.40 or l.45). Adoption also waits for the joint zero-issue X1 round.

The reviewed files are those in `reviews/rev7-digests.json` (manifest SHA-256 `b108f52422c2724d95eb1dca2ced6bdd7da8245ec478af9888aae6864287f76d`, as reported by Main), with the per-file digests in the table above. No preparation, inventory, draw, Jev, labelling or analysis stage is authorised, and ADR-014 is not adopted.
