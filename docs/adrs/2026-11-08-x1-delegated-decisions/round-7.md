# X1: council experiment review, round 9 (revision 9 with ADR-014 round 7)

Date: 2026-11-08. Reviewer: innovation council. The lead wrote the synthesis. Two read-only peer audits ran, and neither saw the lead's leaning:
- `innovation-challenger` (X1Council9.X1R9Challenger): an independent adversarial audit;
- `experiment-peer` (X1Council9.X1R9Peer): an independent methods audit looking for forking paths, confounds and falsifying observations.

This record is X1 round 9 and ADR-014's seventh council review. Save it byte for byte as both `evaluation/x1-triage-ranking/reviews/round-9.md` and `docs/adrs/2026-11-08-x1-delegated-decisions/round-7.md`. **It is the final round for both.**

**Governance.**
- `CLAUDE.md` experiment rules 1–7.
- Design rules 1–3, for ADR-014.
- The ledger rules 1–3.
- The General sections "Toolchain", "Jev labelling", "Model use" and "Secrets".
- Standing owner delegation R-064: where an owner choice comes up, this record gives the council's recommended option.
- Wording preferences alone are not counted as issues.
- Risks already accepted by name in earlier rounds are not reopened without new evidence.

**Scope.**
- Revision 9, as listed in `reviews/rev9-digests.json`. That manifest has 131 files and excludes itself.
- Each round-8 item was checked against the fix required in `reviews/round-8.md` §5. The required text was read raw at l.321 and l.340.
- Earlier closures were re-checked against the revision-9 bytes.

**What was and was not run.**
- This round ran no shell, model call, self-test, simulation, benchmark or digest recomputation, and made no web lookup.
- The evidence comes from reading files, grep and hand arithmetic.
- `codegraph` was not used. Its reference documents (`readiness-and-bootstrap.md` and `query-selection.md`) were read first, but this read-only session had no way to run the executable.
- Main reports independent checks by a separate agent (scratch `x1-rev9-independent-8zbfi9h7`). This round relies on that report:
  - Interpreter: pinned CPython 3.11.11, executable SHA-256 `570b4105…1d6d`.
  - README self-test: exit 0, 249 checks passed, 11.5 s. Planted effect supported. The planted census null and the unequal-π null were not supported. Strong effect 2,000/2,000 support.
  - `kernel_rounding.py` supplement: exit 0, 7 checks.
  - Digests: 131 of 131 matched before and after.
  - `git status` was unchanged.

**Notation.**
- Paths are relative to `evaluation/x1-triage-ranking/` unless stated otherwise.
- "P l.N" is a line of revision 9's `PREREGISTRATION.md`.
- "M l.N" is a line of `oc-manifest.json`.
- "ADR l.N" is a line of `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md`.
- "D9" is `reviews/rev9-digests.json` and "D8" is `reviews/rev8-digests.json`.

**Reviewed files (SHA-256).** Main checked all 131 entries of D9. Main reports the manifest's own SHA-256 as `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`. This round did not recompute any digest. The values below are transcribed from D9 `files_sha256`.

| File | SHA-256 |
|---|---|
| `docs/STATUS.md` | `90166fb05fb88b610ce1c58055c559d3716162e4de217ebf8d8eda9091a4456e` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md` | `b7f260ade5c61d8860158ba93a8e489eb02e1163ea26fc33962b51ba831edbf3` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-1.md` | `7bc229fd392215971646bfc547883050e82512c21777fbd311cf779e3e394710` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-2.md` | `5a298d77bb778d79c17d158657d868ccdadb0ce8a5eef9d01a8d170a22f976ed` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-3.md` | `9530480f20aa59b2c4c72047dfab46c263d73dd86e716a5a566fee4d54f87038` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-4.md` | `0682fcf9a04f0e84e70b3d56e6275bfcb17bb5d9463b519b0c39c2d3dcf26d28` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-5.md` | `b7149420e1c21e245b2d1a1fb96ca4085fa6b84a1c5ea6c231a5a13aab1b0949` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-6.md` | `1f6b3509c1026c03bb502ca70ec3df254b79b8f554d78abf4ff4962f8f9158b3` |
| `docs/adrs/README.md` | `c277db8d8703c087b8c922693abcf1be24b7e0d5006d7798e6aef2b4492ec602` |
| `docs/hypothesis-ledger.md` | `72d8e4a97d671decf19bcffd7ab141360508224762b1421afc64572048cc932b` |
| `docs/research-ledger.md` | `f024b03e282e3bb19d3b504119e3a17b82547ce69c08ac0640a724e4c4665eb5` |
| `docs/scout-design.html` | `f7180b872e7380b8c2a644b8c7fe33b90c9ccf3e3cc0a4c8c3662227dbb3d8a9` |
| `PREREGISTRATION.md` | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `inputs/x11-rev4-digests.json` | `e9edba67c885623df6900a08aee1cde463e5fa78f828bbb04ca3d8efd7a0324c` |
| `oc-manifest.json` | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
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
| `reviews/README.md` | `a42592fab58149f0a96e821b4892aef92c2986a5afe2712a14bdd8a9d4b83984` |
| `reviews/rev3-digests.json` | `894dc63b0522a23589187c6de9fa4e672878fc1c07b85bfbbdcd224e5ab2f448` |
| `reviews/rev4-digests.json` | `01c790837f961cd208fed98b3bd653b514da36e6d8bcee3b2fa4270f88e1876c` |
| `reviews/rev5-digests.json` | `ba1d2ae4efdf1c66e6d5f3002b5d0f892bac079a3efe4cbc2d98fd7826f69a95` |
| `reviews/rev6-digests.json` | `cca1baf432389b094588b6085888e8116442f0ee9baef15cbc750ef30086a746` |
| `reviews/rev7-digests.json` | `b108f52422c2724d95eb1dca2ced6bdd7da8245ec478af9888aae6864287f76d` |
| `reviews/rev8-digests.json` | `8cbdb2114d79daea2d892da1cc294fe79ebfb305d56c37fd4c51d24bb1dd419e` |
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
| `reviews/round-8-response.md` | `9ad889519329559db1e989725fb7a928e56d56dd3ae59d6c98e83e2f72d9dd4e` |
| `reviews/round-8.md` | `1f6b3509c1026c03bb502ca70ec3df254b79b8f554d78abf4ff4962f8f9158b3` |
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
| `validation/kernel-rounding-rev9.json` | `d667c6cff997fdb847d151a93cfe1cb69bd5d3960a3b1d6d601c46f24048c51c` |
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
| `validation/revision9-checks.json` | `d650a484c805692d5d233a5319c470e08057df8af138085098450b778177f18a` |
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
| `validation/strong-effect-OC-rev9.json` | `9da6b7a14b66d163a2140268c8ee73b1bbdd9abaf67b219f7495f81ced901b3b` |
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
| `validation/synthetic-results-rev9.json` | `2ca9a6ac4611e22c7e73007478b55f01b7411ccf5eb2b03cd2ba87693dae2dcd` |
| `validation/synthetic-results.json` | `e65ec36c7d57a0d15f32bc3d2d7bd5f1f76f4148189ec2676653ef996901a350` |
| `validation/synthetic-sources-rev5-initial.json` | `a77aff4f00d87a5d98dafe3114df227b9de27ffb215c45a82ef6074781aee4c2` |

**ADR-014 owned extracts (D9 `owned_extracts_sha256`).**
- ADR-014 index row (`docs/adrs/README.md`): `343ca28ba725d15ebcf235136ea930871a19f9e3f7177cd8726d63a6c4644233`.
- ADR-014 Appendix B card (`docs/scout-design.html`): `b20a7636892feed1f288894e8f8a39b1dfd38ebeedf6635c75bbbdb577313681`.
- HD9 register row: `e937be022de233692782a8a5dae7546c76d1ab122bf65368339ad25d326d89bd`.
- R-081 row: `c02da10a410ecb37cf878f3c1145e48bd8e5b931d4fc04f34fe3eac08cba02e5`.
- HD9 history lines:
  - revision 9: `db0dbafac959e3d8468c003352b518ea06cd9a7d8df5f7dbfdc61bc9d8848593`;
  - revision 8: `ed47dca4e65bb7ce8226b3aa601d278592dca217caedaf12162ddfb4ff24a192`;
  - revision 7: `ae6d74fdfa9a0c7c7bc2139431da062bd42b2ebefcfe8aee194057b9265d1501`;
  - revision 6: `ec3ceabb5d5fd9ccd8834d33ba1a558f645bcad834817c98fe4bde275277bb45`.
- X1 tracker line (`docs/STATUS.md`): `06e927e65336ab92b86be7eca51625aff371f49f8610f6de99d9001db1eca5aa`.
- The remaining extracts are R-064, R-065, R-066, R-067, R-068, R-071, R-072, R-078, R-080, R-085 and R-086, as listed in D9.

**Byte-preservation checks against D8.** The lead compared entries and the challenger compared them independently (D8 l.378–502 against D9 l.392–524).
- **Unchanged (115 files):**
  - every earlier council round and response from rounds 1–7;
  - ADR-014 `round-1.md` to `round-5.md`;
  - `rev3-digests.json` to `rev7-digests.json` and `inputs/x11-rev4-digests.json`;
  - `docs/adrs/README.md` and `docs/scout-design.html`;
  - all 11 preparation files;
  - every validation output, projection and source snapshot from revisions 1 to 8;
  - **all 13 reviewed executables**: `compute_benchmark.py`, `fixed_population.py`, `interval_diagnostic.go`, `interval_diagnostic.py`, `kernel_rounding.py`, `kernel_rounding_test.go`, `outer_simulation.py`, `ranking.py`, the three `reproduce_reports-*-rev5.py` snapshots, `reproduce_reports.py` and `self_test.py`.
- **Changed mutable documents (8):**
  - `docs/STATUS.md`, ADR-014 `ADR.md` and both ledgers;
  - `PREREGISTRATION.md`, `README.md`, `oc-manifest.json` and `reviews/README.md`.
- **New files (8):**
  - ADR-014 `round-6.md`;
  - `rev8-digests.json`, `round-8.md` and `round-8-response.md`;
  - `kernel-rounding-rev9.json`, `revision9-checks.json`, `strong-effect-OC-rev9.json` and `synthetic-results-rev9.json`.

  115 + 8 + 8 = 131, and 123 + 8 = 131. No file was removed.
- **Identical pairs:**
  - `round-8.md` = ADR-014 `round-6.md` = D9 `joint_previous_record_sha256` (`1f6b3509…`, D9 l.17).
  - `rev8-digests.json` = `8cbdb211…`, the manifest digest Main reported in round 8 (= D9 `previous_submission_sha256`).
  - `kernel-rounding-rev9.json` = `kernel-rounding-rev8.json` (`d667c6cf…`). This is consistent with a deterministic rerun:
    - the output has no time, path or output-name field (`kernel_rounding.py:77–83`);
    - its kernel `binary_sha256` `c6adbe8a…` also appears in `strong-effect-OC-rev9.json` provenance (l.583), which is consistent with reproducible Go builds [INFERENCE];
    - Main's separate agent reran the supplement independently (exit 0, 7 checks).
- **Cross-references:**
  - All 12 snapshot archives in `revision9-checks.json` match D9.
  - The source and interpreter hashes in `kernel-rounding-rev9.json`, `synthetic-results-rev9.json` (l.851–857) and `strong-effect-OC-rev9.json` (l.586–594) equal the table entries and D9 `numerical_runtime` (`570b4105…`, go1.26.10 darwin/arm64, optimize 0).

**Verdicts.**
- **X1: zero issues (0 blocking, 0 major, 0 minor). Signed off** (§13).
- **ADR-014: zero issues. Signed off.** ADR-014 is adopted when this record is saved as `round-7.md` in its folder (ADR l.73).
- **Joint signed set: zero issues.**
- **Next stage authorised: preparation only, limited to the actions P l.222 lists (a)–(e)** (§12).

**Severity scale.** Unchanged from rounds 1–8:
- **blocking:** can produce a wrong or unreachable verdict, or breaks experiment rules 1–7 or REQ-169;
- **major:** misleads the result, leaves a preregistered quantity undefined, or makes the study unrepeatable or likely not testable;
- **minor:** text, consistency or provenance. Minor issues must still be fixed before sign-off.
- A fail-closed default that only makes a gate harder to pass is not a defect, unless it makes the gate unreachable for a compliant case.

---

## 1. Recommendation

**Sign off X1 revision 9 and ADR-014 now, jointly, with zero issues. Authorise only preparation actions (a)–(e) (P l.222). The 45-day clock starts at the first such action.**

- Both round-8 fixes are adopted word for word and mirrored at every required location (§4).
- Every earlier closure still holds in the revision-9 bytes (§5).
- No executable, benchmark, calibration or projection byte changed.
- The rule-4 evidence was regenerated as create-once revision-9 outputs and reproduced independently.
- Neither peer nor the lead found a new defect.
- Two readings of the protocol text are recorded in §6 so that no later operator choice arises. Both follow the literal text, which is fail-closed, and they change no byte.

**Recommended practice (not a condition):** finish option G before the first preparation action, on synthetic data only. Option G covers:
- the reviewed locked-size input change;
- the stage-3 OC controller, enforcing both counters and the frozen safety order and scheduling policy;
- a draft final evaluator.

The clock starts at preparation, and throughput is still the tightest constraint (§9).

## 2. Alternatives considered

| Option | Description | Tradeoff | Council position |
|---|---|---|---|
| **A. Joint zero-issue sign-off now; preparation (a)–(e) only** | Freezes revision 9 and adopts ADR-014 | Stage-3 artefacts still need scoped reviews | **Recommended** |
| B. Revision 10 to write §6's two readings into the protocol, and to tag ADR rows 40 and 45 with R8-01/02 | Explicit text | The literal text already decides both cases, and ADR l.15 records the R8 provenance. A round for no defect. | Rejected: readings are recorded here, and the tags are a wording preference |
| C. Defer HD9's `preregistered` transition so that no signed shared file changes | Keeps whole-file digests stable | Breaks ledger rule 2, which says to record the transition when the hypothesis is preregistered | Rejected. Condition 2 in §13 allows exactly this ledger record. |
| D. Authorise inventory or the model-free scans together with preparation | Saves a review cycle | P l.223 requires sealed preparation outputs, a passing binding acceptance with positive control, and a scoped zero-issue review | Rejected |
| E. Option G before the first preparation action | Shortens the clocked path. Its artefacts are pre-reviewed against synthetic sizes. | Pre-clock work on no real data | **Recommended practice**, not a sign-off condition |

## 3. Rule checks

- **Rules 1–2.** This round records scope, issues (none), severities, verdicts and every digest. The sign-off in §13 names the preregistration, the OC manifest, every executable, the evidence, the manifest digest and ADR-014's governed files.
- **Rule 3.** No predictor–outcome association has been computed on real data. All evidence is synthetic. The historical method selection was not rerun (D9 `historical_method_selection_rerun: false`). The benchmark, calibration and projection were not rerun (D9 `benchmark_calibration_projection_rerun_in_revision9: false`).
- **Rule 4.**
  - Planted effect: supported. Planted census null and unequal-π null: not supported (`synthetic-results-rev9.json` l.253, 729–785, 725, 842).
  - Strong effect: 2,000/2,000, Wilson MC95 lower .9980829527187469. That equals 2000/(2000 + 1.96²) = 2000/2003.8416, as expected for n = 2000 [hand arithmetic]. Realised gain .463, zero inadequate (`strong-effect-OC-rev9.json` l.581, 596, 4086–4092).
  - Both were reproduced independently by Main's separate agent.
  - These runs went against round 8 §1's "no rerun" advice. Main explicitly asked for them, and they were written as new create-once outputs (D9 `reason_for_rerun`). That is not a defect.
- **Rule 5.**
  - Nothing has been changed after a sign-off; this round is the first sign-off.
  - Every executable was archived before modification in earlier revisions, and all 12 archives close (`revision9-checks.json`).
  - After this round, the frozen set is defined by §13.
- **Rule 6.**
  - Binding and total envelopes, the reserve-adjusted memory envelope, the timing drift, the safety statuses and the scheduling are all preregistered (P l.171, l.181–185, l.240; README l.93–110).
  - The disclosures of earlier revisions' byte loss and superseded outputs are preserved (README l.83–85; D9 `source_loss`, `superseded_outputs`, `preserved_superseded_projections`, which now lists all three revision-7 projections, adopting round 8's advisory).
- **Rule 7.** All superseded projections and outputs are kept, with labels (README l.106).
- **Model use and secrets.** No model was called. The binding recipe checks the Jev key by presence only and uses stub-only positive controls. Model calls happen only via `claude -p` / `codex exec` and only under later stage authority.
- **Toolchain.** `.tools/venv/bin/python` 3.11.11 (`570b4105…`) and `go1.26.10 darwin/arm64`. `toolchain_pin_changed: false`.
- **Design rules.**
  - ADR-014 has sequential ID 014 and an adopted-form status matching ADR-011 at `7bb73e52` (D9 `adr_status_check`).
  - Its Review section defers the outcome and digests to this folder's final round (ADR l.73).
  - Its Records list includes `round-6.md` (ADR l.83).
  - The decision text names revision 9 and "R3–R8 are also this folder's rounds 1–6" (ADR l.15).
  - The index row and design card are byte-unchanged and still accurate (`docs/adrs/README.md:28`; `docs/scout-design.html:1006–1010`).
  - The decisions cite only signed-off X1 rounds, R-IDs and synthetic evidence; no product result is cited.
- **Ledgers.**
  - HD9 is proposed (`docs/hypothesis-ledger.md:36`). A new dated revision-9 history line was added (l.40), and earlier history lines are unchanged (extract digests above).
  - R-081 mirrors the new suffix, under Main's standing permission for this unmerged row (`docs/research-ledger.md:83`).
  - The next free ID is R-087 (`docs/STATUS.md:128`), and no ID was consumed.
  - This round uses no new source, so no new R-ID is needed.
  - HD9's transition to `preregistered` follows this sign-off (§11 step 3).

## 4. Verification of round-8 fixes

| Issue | Status | Evidence |
|---|---|---|
| X1-R8-01 | **Fixed** | **Exact text.** `round-8.md` §5's text (raw l.321) appears character for character in P l.181, from "Locked remeasurement is one create-once run, on the host…" to "…maximum over all completed runs." The lead compared raw lines and so did both peers.<br>**Mirrors.** Identical in ADR l.45, M l.165 (`resources.binding_resource_stop_rule`), D9 l.181 and `revision9-checks.json` l.2.<br>**Paraphrases.** README l.108 and R-081 (`docs/research-ledger.md:83`) carry the full new suffix.<br>**No conflicting rule.** "Binding resource stop rule:" occurs once in each of P, M and ADR. No current prose still contains the old phrases "registered replicate structure" or "(crash, closure mismatch or host fault)"; they survive only in historical rounds.<br>**Elements present.** Host binding with hw.model, hw.memsize, logical CPU count and OS build. The revision-7 replicate structure, concretely backed by README l.122 and l.124 and by `compute_benchmark.py:247–253` (eight cells: two L values × complete/missing × two fixture types, each at 2,000 trials with one and two workers, giving 32,000 trials). Size-source-only code change under rule 5. Integrity failures defined, with OS-evidenced host faults. A resource-exhaustion crash is a failed fit. Operator termination or host load leaves the fit unmeasured, which means not testable (P l.181; P l.240 "absent capacity/fit fails closed").<br>**Checks.** `revision9-checks.json` items "new locked suffix equals council verbatim text", "rule mirrors exactly in ADR/OC manifest", "one operational rule in protocol" and "locked suffix matches README and R081". |
| X1-R8-02 | **Fixed** | **Exact text.** `round-8.md` §5's sentence (raw l.340) appears right after "with no reordering after the first output." in P l.171, ADR l.40, M l.143 (`final_safety_OC.schedule`) and README l.93. It is also copied at D9 l.182 and `revision9-checks.json` l.36.<br>**Unchanged context.** The keyed order, the group-measured rule, the "unmeasured; cell fails" result on exhaustion, the "not completed by gold" status and exclusion from pooled Wilson summaries, and the non-gating role all follow unchanged (P l.171).<br>**Stage text.** Stage 3 still routes safety through §8 (P l.224). |

**ADR-014 follow-through.** Records adds `round-6.md` (ADR l.83). ADR l.15 records that revision 9 takes both R8 recommended options under R-064.

## 5. Earlier closures re-confirmed

- **X1-R7-01** (supplement closure):
  - `kernel_rounding.py` is unchanged (`906d385f…`).
  - The revision-9 output reports seven counted checks and records the interpreter, Go version, kernel binary and test binary (`kernel-rounding-rev9.json`).
- **X1-R7-02** (reserve and maximum-phase RSS): the memory prefix of the stop rule is unchanged and character-identical across P l.181, ADR l.45 and M l.165. The arithmetic is unchanged (§8).
- **X1-R7-03 and X1-R7-04:** their residual gaps were X1-R8-01 and X1-R8-02, now closed.
- **X1-R7-05:** `docs/hypothesis-ledger.md:43` reads "worst-case enumeration and L400 RSS projection". The revision-6 extract `ec3ceabb…` is unchanged.
- **Rounds 1–6:**
  - every earlier council record and response is byte-identical;
  - the single binding stop rule, the safety statuses, the staged authority (P l.220–226), source-loss disclosure, snapshot discipline and the pinned interpreter are all unchanged.
- **Trackers:**
  - P l.3, README l.3, `reviews/README.md:3` and `docs/STATUS.md:128` name revision 9 and the X1 round 9 / ADR-014 round 7 review.
  - The draft banners follow house convention: the signed bytes are the submitted draft, and authority comes from this record, as P l.3 states. The x2d preregistration was signed with the same pattern ("DRAFT for council round N").

## 6. New issues

**None.**

### Readings recorded, with no change to the text

Both readings follow the literal protocol text and are fail-closed. They are recorded so that no operator choice arises later.

- **R9-reading-1: a completed run later found to have a closure mismatch.**
  - P l.181 says "each binding input is the maximum over all completed runs". A run that reached its end is a completed run, even if its closure check then fails.
  - Its values therefore enter the maximum, and the closure mismatch also permits one rerun.
  - The protocol text governs. Round 8's §2 D and §10 narrative, which said closure-invalid measurements are "not forced into the maximum", is superseded on this point. That narrative was the council's rationale, not protocol, and the reading adopted here can only make the fit harder to pass. (Peer candidate 2.)
- **R9-reading-2: changing host after the remeasurement.**
  - The protocol allows one remeasurement run, with reruns only after a qualifying integrity failure. A host change is not a rerun cause.
  - After a completed remeasurement, the binding OC must run on the recorded host. Running it elsewhere is a deviation under P §11 (pause the affected work, review, record under Deviations).
  - Round 8 §9's phrase "a host change means a new locked remeasurement" is corrected accordingly. The protocol is stricter, and the stricter reading governs. (Peer candidate 1.)

### Considered and not counted

- **How locked sizes map onto the eight-cell benchmark** (peer).
  - The trial count, cell set and worker-equality structure are now fixed. The size input itself is reviewed under rule 5 before the run.
  - This is the residue that round 8 §2 E deliberately closed with one clause. There is no new evidence.
- **The stage table does not name the OC controller** (peer). P l.224 routes safety through §8, whose sentence requires the reviewed stage-3 controller. Rounds 7 and 8 accepted the controller as a stage-3 prerequisite.
- **No evidence type for a "crash not caused by resource exhaustion"** (peer). Not counted:
  - a crash produces no completed run, so it cannot lower the maximum;
  - memory demand at fixed sizes is deterministic in practice, so reclassifying an out-of-memory crash buys at most a rerun that fails again [INFERENCE];
  - deliberate misclassification is misconduct, not a design gap;
  - round 8 adopted this text with the classification as worded.
- **ADR rows 40 and 45 have no R8-01/02 tags in their source column** (peer). ADR l.15 records the provenance. This is a wording preference.
- **`kernel-rounding-rev9.json` is byte-identical to the revision-8 output** (lead). This is consistent with deterministic output (see the byte-preservation checks above). It was reproduced independently.
- **The README invocation includes `--oc-days 15`, but the rule's quoted invocation omits it** (lead). The rule names "README benchmark invocation", which identifies README l.122 unambiguously. The text is the council's own verbatim wording.
- **Stale "draft" and "waits" statements in signed trackers after sign-off** (lead). Adoption is recorded outside the signed files (route (a), `docs/adrs/README.md:12`; ADR-011 precedent). The only permitted in-set edit is the ledger-rule HD9 record (§13 condition 2).

## 7. Checked and sound

- **Counts** agree across P, M, D9 and the projection:
  - 17,920 complete/reference cells and 252 missing cells;
  - 336 safety profiles and 8,064 outcomes;
  - 336 complete + 1,680 full-vertex interval groups;
  - hard caps of 4,000 binding and 2,000 safety executions per population, and 8,000 attempts at p_gate = .5 (D9 `resources`).
- **The p_gate estimate** (100,000 keyed size-only redraws at inventory; P l.163) is a separate inventory step and does not conflict with the size-source-only remeasurement.
- **Preparation actions (a)–(e)** are bounded at P l.222:
  - exact pinned objects with canonical URLs and SHA verification;
  - closure attestation;
  - model-free binding implementation and fixture test;
  - identifier delta review for the test repositories only;
  - sealing.
  
  Frame or dev collection, SRS, model calls and score or label extraction are excluded.
- **Inventory and later stages** remain separately gated (P l.223–226).
- **No over-authorisation** appears in P, README, ADR l.15, STATUS or `reviews/README.md`.

## 8. Evidence index

- **Arithmetic (unchanged from round 8 §8; inputs are byte-identical):**
  - Binding = 397.1563 + 306.9737 + 7.7177 = 711.8477 h, which ÷ 192 = 3.7075 days.
  - Total = 1,288.4371 h (6.7106 days).
  - Memory break-even U = 11,219,137.75 [INFERENCE: linear].
  - At U = 12 M, memory is 64,297,238,528 B, which exceeds the 60,129,542,144 B budget.
  - D9 `resources` repeats these values.
- **X1-R8-01:**
  - `round-8.md` raw l.321;
  - P l.181; ADR l.45; M l.165; D9 l.181; README l.108, l.122, l.124; `docs/research-ledger.md:83`;
  - `compute_benchmark.py:247–253`; `reproduce_reports.py:126–152`.
- **X1-R8-02:** `round-8.md` raw l.340; P l.171; ADR l.40; M l.143; README l.93; D9 l.182.
- **Rule 4:** `synthetic-results-rev9.json` l.253–254, 725, 729–785, 842, 851; `strong-effect-OC-rev9.json` l.2, 581–596, 4086–4092; `kernel-rounding-rev9.json`.
- **Accounting:** D8 l.378–502 against D9 l.392–524; D9 l.17–19, l.476–477.

## 9. Risks

- **Throughput and clock.** This is carried from rounds 7 and 8 and is still the tightest constraint. The 45-day clock starts at the first preparation action and is never paused. Option G before that action reduces exposure.
- **Stage-3 artefacts are not yet written:**
  - the locked-size input change;
  - the OC controller, with both counters, the frozen safety order and scheduling policy, and the SHA-256 of the ordered group list recorded at freeze;
  - the final evaluator;
  - the product-backed rule-4 harness;
  - GUIDE and isolation records.

  Each needs its scoped zero-issue review. Their absence blocks only later stages.
- **Host binding** (R9-reading-2). Losing the remeasurement host after a passing fit leads to a deviation and review, not a fresh timing run.
- **Shared-file staleness and merge order.**
  - STATUS, the ledgers, the index and `scout-design.html` are shared.
  - A sibling change landing first alters their whole-file digests. Condition 2 limits what that may touch.
  - A change to any X1-owned extract or X1 file voids the sign-off.
- **Attestation.** The digests and the independent runs rest on Main's and its separate agent's reports. The council had no shell.
- **Reversibility.** Everything is still pre-data. A defect found during preparation is fixed by a new review round (rule 5). ADR-014 changes no product default, flag, threshold or requirement.

## 10. Where the peers changed the council's recommendation

- **Challenger:**
  - Found no issues.
  - Independently reproduced the 123 + 8 = 131 accounting and the 115 / 8 / 8 split.
  - Confirmed that the replicate structure is concrete in code (`compute_benchmark.py:247–253`, 32,000 trials).
  - Confirmed that the kernel output's byte identity is not a provenance contradiction.
  - It did not change the verdict. It strengthened the lead's confidence in the accounting.
- **Experiment peer:**
  - Found no issues.
  - Its candidates 1 and 2 became the recorded readings in §6. The lead had not noticed that round 8's own narrative (§2 D, §10 and §9) diverged from the literal adopted text.
  - Its candidates 3–6 were judged accepted residue or wording, and were not counted.
- **Lead:**
  - Verified both verbatim insertions against the raw required text.
  - Checked the kernel output's byte identity against the generator.
  - Checked the draft-banner convention against the signed x2d precedent.
  - Wrote the sign-off conditions, including the ledger-rule exception for HD9's transition (option C rejected).

## 11. Execution outline (one integration owner)

1. **Save this record** byte for byte as `evaluation/x1-triage-ranking/reviews/round-9.md` and `docs/adrs/2026-11-08-x1-delegated-decisions/round-7.md`.
   - *Test:* the two SHA-256s are equal.
2. **Re-verify** the manifest SHA-256 `2253911b…` and all 131 file digests in §13 before committing.
   - *Test:* 131/131 match, with no extra modified file in the signed set.
3. **Ledger-rule record (condition 2):**
   - In `docs/hypothesis-ledger.md`, set HD9's state to `preregistered` and its date cell to the sign-off date.
   - Append one dated history line: "HD9: proposed → preregistered; X1 revision 9 signed off in X1 round 9 / ADR-014 round 7 (X1; R-064, R-081)."
   - Change nothing else in the signed set. Record ADR-014 adoption only outside the signed files (route (a)).
   - *Test:* the only diff inside the 131 files is those two cells plus one appended line.
4. **Commit and merge** on base `7bb73e52`.
   - If origin/main has moved, rebase. Any resulting change to a signed byte outside condition 2 voids this sign-off and needs a new round.
5. **(Recommended) Option G before the first preparation action:**
   - write the size-input change, OC controller and final-evaluator draft against synthetic sizes;
   - submit each for its scoped review.
6. **Preparation actions (a)–(e) only:**
   - record the timestamp of the first action as the 45-day clock start;
   - make no frame or dev collection, SRS, model call or score or label extraction.
7. **Seal the preparation outputs** and submit them for the scoped zero-issue inventory review. That review must include the binding acceptance with its positive control (P l.223).

## 12. Stage authorisation

| Stage | Status now | Earliest path |
|---|---|---|
| Drafting, synthetic validation, option G engineering on synthetic repositories | **Permitted** | — |
| **Preparation, actions (a)–(e) of P l.222 only** | **Authorised** by this record | The 45-day clock starts at the first such action |
| Inventory, model-free scans and SRS draw (including dev-sample collection) | **Not authorised** | All of:<br>• sealed preparation outputs;<br>• a passing binding acceptance with positive control;<br>• runtime closure;<br>• the identifier review for the test repositories;<br>• a scoped zero-issue inventory review |
| Full OC and implementation (including GUIDE-author sessions) | **Not authorised** | All of:<br>• locked sizes;<br>• a binding CPU and memory fit under P l.181, with a rule-5-reviewed size input;<br>• a reviewed OC controller with the frozen safety order and scheduling policy;<br>• a reviewed final evaluator;<br>• product-backed rule-4 frames;<br>• GUIDE and isolation records reviewed to zero issues before any dev label |
| Dev Jev and labels; test scoring and labels | **Not authorised** | Under full authority only |
| Gold unblinding and analysis | **Not authorised** | — |
| ADR-014 adoption | **Adopted** on saving this record as `round-7.md` in its folder | Record adoption outside the signed files |
| Default, threshold or product-sampler change | **Not authorised** | A separate signed ADR |

**Next stage authorised:** preparation, limited to actions (a)–(e).

## 13. Sign-off statement

**Zero issues. Signed off.**

**X1.** The innovation council signs off X1 (REQ-169, HD9) at revision 9 with zero open issues. This round is X1's final pre-preparation round. The submission is:
- worktree `scout-x1`;
- uncommitted changes on base origin/main `7bb73e52ffcc4daa74247bee1ecc2008ffd1d830`;
- reviewed preregistration `PREREGISTRATION.md` SHA-256 `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`;
- OC manifest `oc-manifest.json` `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d`.

The reviewed scripts are:
- `compute_benchmark.py` `d39f40a2…c6ea90d`;
- `fixed_population.py` `ed670a6b…49a2`;
- `interval_diagnostic.go` `437f9c25…9417`;
- `interval_diagnostic.py` `a4b23337…345`;
- `kernel_rounding.py` `906d385f…7c23`;
- `kernel_rounding_test.go` `33bde17b…ff8def`;
- `outer_simulation.py` `91d07760…ece23f6`;
- `ranking.py` `8bc2e7cb…0a780`;
- `reproduce_reports-initial-rev5.py` `0b7781db…0ed4`;
- `reproduce_reports-optimized-rev5.py` `d680ee08…102887`;
- `reproduce_reports-pilot-rev5.py` `3ee6476c…5e2`;
- `reproduce_reports.py` `023e2e5b…0168e`;
- `self_test.py` `36d07d56…eb8a`.

Full digests are in the reviewed-files table.

**ADR-014.** The innovation council signs off ADR-014 ("Capture X1 delegated experiment decisions") with zero open issues. This record is the folder's final round. It covers:
- ADR summary `ADR.md` `b7f260ade5c61d8860158ba93a8e489eb02e1163ea26fc33962b51ba831edbf3`;
- index `docs/adrs/README.md` `c277db8d8703c087b8c922693abcf1be24b7e0d5006d7798e6aef2b4492ec602` (ADR-014 row extract `343ca28b…4644233`);
- design card in `docs/scout-design.html` `f7180b872e7380b8c2a644b8c7fe33b90c9ccf3e3cc0a4c8c3662227dbb3d8a9` (ADR-014 article extract `b20a7636…313681`);
- the X1 preregistration, OC manifest, implementation and evidence named above.

**Scope.** The signed set is exactly the 131 files at the SHA-256s in the reviewed-files table, bound by `reviews/rev9-digests.json` with manifest SHA-256 `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`. Main computed and verified those digests.

**Conditions of this sign-off:**

1. **The bytes are frozen.** Under experiment rule 5 and design rule 2, any change to a signed file voids this sign-off and needs a new zero-issue round. This covers:
   - the preregistration, OC manifest, README, preparation recipes and executables;
   - the evidence, inputs, earlier records and manifests;
   - ADR-014 `ADR.md`, apart from the status-line updates design rule 2 allows when a later ADR is adopted;
   - the X1-owned extracts.

   The only exceptions are the two in condition 2.
2. **Permitted later edits in shared files:**
   - **(a) The ledger record.** Ledger rule 2 requires HD9's state and date cells to change to `preregistered`, with one appended dated history line (§11 step 3).
   - **(b) Sibling integration.** Changes in `docs/STATUS.md`, `docs/research-ledger.md`, `docs/hypothesis-ledger.md`, `docs/adrs/README.md` and `docs/scout-design.html` outside the X1-owned extracts listed in D9 `owned_extracts_sha256` are allowed. The exception is the HD9 cells changed under (a).
   - Any change to HD9's hypothesis text or sources, the R-081 row, an earlier HD9 history line, the ADR-014 index row or the ADR-014 design card voids this sign-off.
3. **Adoption happens outside the signed set** (route (a), `docs/adrs/README.md:12`). No signed file is edited to mark adoption beyond condition 2(a).
4. **Readings.** §6's R9-reading-1 and R9-reading-2 govern how P l.181 is applied.
5. **Authority** is limited to preparation actions (a)–(e) (§12). Every later stage needs its scoped zero-issue review.
6. **Attestation.** The digests, the unchanged base and the independent reruns rest on Main's reports. The council had no shell.

Round 9 closes as follows:
- **X1: zero issues, signed off.**
- **ADR-014: zero issues, signed off, adopted on saving this record as its folder's `round-7.md`.**
