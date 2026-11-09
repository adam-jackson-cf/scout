# X1 stage-3 draft review — draft final evaluator (`stage3-drafts/final-evaluator/`) — round 3

## Scope and authority

- This is a scoped innovation-council review, round 3, of one Option G stage-3 draft. It checks the fixes for the round-2 issues in `reviews/round-2.md`, as described in the author's `reviews/round-2-response.md`. It also judges the new choices the author made while fixing them.
- Worktree: `/Users/adamjackson/Projects/work/scout-x1g`. PR #21 head: `38cead909de6de815ef8e9453630f404f6e2628f`.
- **This review gives no authority to run anything.** That includes the OC, the inventory, model calls, scans and anything on real repositories.
- How the round was done:
  - The lead read every source, test, runner and record file and the signed text.
  - An `innovation-challenger` (adversarial pass) and an `experiment-peer` (competing readings) worked in parallel and were not told the lead's view. Neither had a shell.
  - The lead checked every finding below against the files.

## Reviewed files (SHA-256 recomputed by Main with `shasum -a 256` at the head above)

| File | SHA-256 |
|---|---|
| `README.md` | `1acc107a64e1306fbe346528ab8e2f46760cf37ff7655532a5af08ed2162c30f` |
| `check_frozen.py` | `8130568a8e3e49fdf420c110bc9825ad20f483237f1f8745a2d96fa4009a56a1` |
| `evaluator.py` | `940480b9190971549acb2abd20cd9853c67a377a45cf3b1ab2ccd927d96f7eed` |
| `provenance.py` | `5742d827f9e080a9934b39f94670cdecaba8bf74c6b58bcfd65d3a230e8c3932` |
| `mutate_round2.py` | `e09b53c2a19ec2f09c207bfbec0c08fd633b932839f4ecff334d860c2a3191a9` |
| `test_evaluator.py` | `9948edea3d09902c6609362fc13cbeb010813097526a1b1d023859369ea60703` |
| `reviews/round-2.md` | `301bb41f025f2f592b9907a87a725fb645550a8ef491abd5496bab2e2b0e00a4` |
| `reviews/round-2-response.md` | `c9022ef164f6a620206c44d72ba013fb83392f491123c0d2d9834b55e4ae1426` |
| `reviews/round-2-test-output.txt` | `fe20fd766c02f16537145bce289c5a7a7447d1d7004b215931a15999e4277dab` |
| `reviews/round-2-mutant-output.txt` | `6a9747089ee3dd0736833941f3fcf1a36401843ee9bda32765ef2fdfa4478302` |
| `reviews/round-2-frozen-output.json` | `2c5d8a155e57cdd60a2b6ea295e9c0c1a2453acccf0ce662a60e711a85df50d0` |

- `check_frozen.py` is byte-identical to rounds 1 and 2.
- The first four lines of `round-2-mutant-output.txt` record the digests of `evaluator.py`, `provenance.py`, `test_evaluator.py` and `mutate_round2.py`. All four equal the digests above, so the mutant evidence was produced against the reviewed bytes.

## Signed basis (unchanged)

| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| X1 `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `oc-manifest.json` | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |

Signed text applied:
- P §7 l.132 and l.135.
- P §8 l.163, l.175, l.177 and l.181.
- P §10 l.206.
- Manifest `final_evaluator_contract` (l.196–200).

## Round-2 issue disposition

Paths are relative to the artefact folder. Test names are methods of `EvaluatorTests` in `test_evaluator.py`. "Mutant output" means `reviews/round-2-mutant-output.txt`.

| ID | Fixed | Evidence (code; catching test; mutant) |
|---|---|---|
| R2-01 | **Yes** | **Code:** every row starts from the resource-fit result (`evaluator.py:146`). Each cell's result is then ANDed into its row.<br>**Each part of the required fix:**<br>• The CPU-failing `evaluate` call now asserts status `not testable before dev` and that every row is `False` (`test_evaluator.py:279-282`).<br>• `test_resource_fit_propagates_to_every_row` (:284-297) runs two `evaluate` fixtures:<br>&nbsp;&nbsp;– CPU exactly 2,880 core-hours, which must pass with every row `True`. The other inputs are 1e-300, which keeps them positive without changing the 2,880 total.<br>&nbsp;&nbsp;– Memory only: `construction_rss = budget/8+1`, so 8×RSS = budget + 8, with CPU asserted below 2,880. It must give `not testable before dev` with every row `False`.<br>**Mutant:** `passed = {m: True` is killed by an assertion failure (mutant output l.6-25).<br>**Other reason ruled out:** the 2,880 fixture uses the same records and passes, so the rows fail because of memory, not something else. |
| R2-02 | **Yes** | **Test:** `test_signed_manifest_pin_with_consistent_bundle_resealing` (:134-145) changes the manifest bytes (`+b"\n"`). The changed bytes are used as both the argument and the bundle member. The test updates the bundle digest and the OC output's `signed_manifest_sha256`, reseals the OC output, and then asserts the exact message raised only at `provenance.py:78-79`.<br>**Mutant:** with that check deleted, the test reports `ValueError not raised`, so no earlier check catches the change (mutant output l.27-43). |
| R2-03 (and the R1-05 residual) | **Yes** | **Test:** `test_nonfinite_exponent_gain_bytes_fail_cells` (:226-244) builds OC bytes by hand with `1e400` and `-1e400` gains, for a reference cell and a missing cell, at targets 0 and .2. That covers all 8 required combinations. Each run goes through `evaluate` and asserts both the cell and row 700 fail. Line :240 confirms the sealed bytes are the bytes the evaluator reads.<br>**Mutants:**<br>• Deleting `math.isfinite` at `evaluator.py:177` fails exactly the two cases that would wrongly pass: missing/0/−inf and missing/.2/+inf (mutant output l.45-71).<br>• The reference-branch mutant (`:165`) fails the matching reference cases (l.73-99).<br>**Correction to round 2:** R2-03 said `isfinite` was the only guard. Under the round-2 code that was not accurate [INFERENCE]: the round-2 seal check recomputed `digest(encode(records))`, and `encode` refuses ±inf, so evaluation refused as a whole. The fix meets the required fix as written. The new exact-bytes seal is what makes `isfinite` reachable (judged below). |
| R2-04 | **Yes** | **Test:** `test_complete_host_and_load_provenance` (:315-329) covers:<br>• each of the four host fields missing;<br>• wrong types, including bool and string `logical_cpus`, bool and string `hw.memsize`, zero values and empty strings;<br>• `load_averages` with 2 entries, or containing −1, `"0"` or `True`, or missing;<br>• `concurrent_process_count` of 0, `True`, or missing.<br>**Hosts kept equal:** the bad host is passed as both the OC host and the run host (:321), so the equality check cannot be what catches it.<br>**Mutants:**<br>• `validate_host` as a no-op: 12 assertion failures plus 2 downstream errors (mutant output l.101-253).<br>• Load clauses removed: 8 assertion failures (l.255-335).<br>This matches P l.181. |
| R2-05 | **Yes** (exact required values) | **Test (`test_ordered_typed_verdict_and_interval_boundaries`):**<br>• `point=nextafter(.10,0), lower=.001, upper=.4` → `inconclusive` (:409).<br>• `lower=.3, upper=.2` → `inconclusive` (:410).<br>• The passing side, `point=.10, lower=.001` → `support` (:411).<br>**Mutants:** both are killed, each with `'support' != 'inconclusive'` (mutant output l.337-377). |
| R2-06 | **Yes** | **Code:** `complete_bootstrap_evidence` (`evaluator.py:195-202`) requires exactly the five recommended fields, with exact types and values. It is wired into selection at `:219`.<br>**Test:** `test_complete_typed_bootstrap_evidence` (:374-394) checks:<br>• the complete schema passes;<br>• `True`, `"false"`, `None`, `{}` and an extra field are refused;<br>• for each field: missing, a wrong value, the wrong numeric type and a bool are refused.<br>**Mutants:**<br>• Truthiness restored: 23 assertion failures.<br>• Check inverted: the positive case fails (mutant output l.379-625). |

Round-1 residuals: R1-05 was the only one carried, as R2-03, and it is closed above. No regression was found in any earlier accepted behaviour: identities, rev8 arithmetic, maxima, ordered verdicts, or no-bypass post-dev selection.

## New author choices

- **Exact-bytes records seal** (`provenance.py:36-64`; `evaluator.py:129`). Accepted: it does not weaken a signed safeguard.
  - **No signed definition to conflict with.** Neither P nor the manifest defines how `records_sha256` is computed. The manifest requires "authentic locked-inventory provenance" and "signed manifest identity", not a canonical hash of the record array.
  - **The digest still covers every byte.** The whole OC payload stays pinned by its external digest (`provenance.py:72-80`).
  - **Extraction is safe for the cases examined:**
    - A strict `decode` runs first and refuses duplicate keys, including escaped equivalents, and the NaN/Infinity literals (`:38`).
    - Navigation uses `raw_decode` and compares top-level keys after unescaping, so a nested `records` key or misleading string content cannot redirect it. `test_exact_record_member_bytes` (:246-253) covers these cases.
    - A BOM or non-UTF-8 input raises, so it fails closed [INFERENCE: Python `json` behaviour].
  - **One behaviour change:** a non-finite value in a field that a decision reads now fails that cell and its row, instead of refusing the whole evaluation. That is what P l.163 and the manifest's "unmeasured input fails closed" require.
  - Non-finite values in fields no decision reads are now accepted silently. They change no outcome.
  - Producers that emit compact sorted-key JSON get the same digest as before.
- **Typed bootstrap schema:** it matches the R2-06 recommended fix field for field and only makes the check stricter. Recomputing c_U from receipts stays with the dev-gate stage, and anchor authenticity stays under round-1 Open choice 1.
- **Owner-facing open choices:** `README.md` and `round-2-response.md` pose none (R-064 satisfied).

## New issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| — | — | — | None. No concrete input was found that leads to a wrong outcome against the signed text. | — |

### Considered and not counted

- **Only `construction_rss` is pushed past the memory budget in the tests** (raised by the peer).
  - `evaluator.py:84` takes the maximum of all three RSS inputs, as P l.181 requires. The tests (:292, :303-305) only ever push `construction_rss` over the limit. So a mutant that reads only `construction_rss` would survive [INFERENCE].
  - Not counted, for three reasons:
    - the current code is correct for every input, so no input gives a wrong outcome;
    - the memory decision boundary itself is tested (budget/8 passes, +1 fails, including through `evaluate`);
    - the per-input arithmetic was accepted under R1-02 and R1-08 in round 2 without isolating each term, and there is no new evidence to reopen those verdicts.
  - Recorded as suggestion S-01.
- **`one_sided_level` does not say upper vs lower** (raised by the peer).
  - This schema is the council's own R2-06 recommendation, and the declared fields are self-reported either way.
  - Checking that c_U really is the upper bound means recomputing it from receipts, which belongs to the dev-gate stage. Recorded as suggestion S-02.

## Non-blocking suggestions

- **S-01:** repeat the budget/8 (pass) and budget/8 + 1 (fail) fixtures for `calibration_rss` and `trial_plus_helper_rss`, with at least one going through `evaluate`.
- **S-02:** when the dev-gate stage defines the c_U recomputation, consider adding the bound direction (`"upper"`) and the statistic (mean billed cost) to the evidence. Changing the frozen field set needs a new council round.
- **S-03:** `README.md:43` says "Round-2 edits have not been verified during implementation", but `README.md:53` reports the round-2 results. Reword l.43.
- **S-04:** add extractor fixtures for an escaped duplicate key (`"rec\u006frds"`), an escaped target key, multibyte UTF-8 content and whitespace inside the records value.
- **S-05:** `round-2-frozen-output.json` and `round-1-frozen-output.json` have the same digest. That is expected, because the signed files did not change, but a one-line note in the response would save a reader from suspecting a copy.

## Verification

- **The lead and both peers had no shell.** In this round they ran no tests, mutants or hashes.
- **Main's independent run** at head `38cead90…` (clean tree, pinned `.tools/venv/bin/python` 3.11.11, `PYTHONDONTWRITEBYTECODE=1`): `final-evaluator/test_evaluator.py` ran 20 tests in 258.191 s, OK. Main also recomputed every file digest above.
- **The author's committed outputs** were read but not reproduced:
  - 20 tests OK in 256.070 s;
  - 10/10 mutants killed by assertion failures. The runner counts a mutant as killed only if the test run exits nonzero and prints `FAIL:` (`mutate_round2.py:68-73`), so a crash alone does not count;
  - frozen check: 126/131 literal matches, 5/5 groups valid, 131/131 unchanged against `e2b94e3c`, manifest digest `2253911b…`.
- **Every mutant listed in round 2's verification section is in `mutate_round2.py:14-42`.** One qualification: the reference-gain mutant replaces the whole validity predicate rather than deleting one `isfinite` call.
- **Reproduce:**

```sh
cd /Users/adamjackson/Projects/work/scout-x1g
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/test_evaluator.py -v
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/mutate_round2.py
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py
```

## Recommendation, alternatives and risks

- **Recommendation:** accept this evaluator for its stage-3 role as it stands. S-01 to S-05 can be taken up in the authoritative-variant review.
- **Alternatives:**
  - **Return to whole-payload strictness:** decode records with a `parse_float` that refuses overflow. Rejected: `isfinite` would become unreachable again and need explicit acceptance, and the signed text gains nothing.
  - **A single strict loader** that returns both the parsed object and the exact byte span of the records value. This is a maintainability option only; adopt it only with fixtures showing it behaves the same.
- **Risks:**
  - The field names of the typed bootstrap schema and the exact-bytes definition of `records_sha256` are now part of the contract with later stages. Changing either needs a new council round.
  - The authoritative variant still needs its own scoped zero-issue review: real trust anchors, binding to the controller and the source/runtime closure, and `binding_gate_established`.

## How the peers changed the outcome

- **Challenger:**
  - independently confirmed every R2 disposition;
  - found no way to bypass the seal through `member_bytes`;
  - added S-04 and the single-loader alternative;
  - noted that the reference-gain mutant is not an isolated deletion.
- **Peer:**
  - supplied the correction to R2-03's premise, which the lead verified against `provenance.py:19` and the round-2 seal expression;
  - raised the RSS-coverage candidate, which the lead did not count (reasons above) and recorded as S-01;
  - raised the c_U direction point (S-02), the README wording (S-03) and the identical frozen-output digests (S-05).
- Neither peer changed the verdict.

## Verdict

**Zero issues — draft accepted for its stage-3 role, conferring no run authority.**

**Sign-off statement.** The innovation council has reviewed the X1 Option G stage-3 draft final evaluator at PR #21 head `38cead909de6de815ef8e9453630f404f6e2628f` and found zero open issues. The files reviewed:
- `README.md` `1acc107a64e1306fbe346528ab8e2f46760cf37ff7655532a5af08ed2162c30f`
- `check_frozen.py` `8130568a8e3e49fdf420c110bc9825ad20f483237f1f8745a2d96fa4009a56a1`
- `evaluator.py` `940480b9190971549acb2abd20cd9853c67a377a45cf3b1ab2ccd927d96f7eed`
- `provenance.py` `5742d827f9e080a9934b39f94670cdecaba8bf74c6b58bcfd65d3a230e8c3932`
- `mutate_round2.py` `e09b53c2a19ec2f09c207bfbec0c08fd633b932839f4ecff334d860c2a3191a9`
- `test_evaluator.py` `9948edea3d09902c6609362fc13cbeb010813097526a1b1d023859369ea60703`

The review was against the unchanged signed basis:
- `PREREGISTRATION.md` `abd206c3…6fdb`
- `oc-manifest.json` `1caea1a5…761d`
- `reviews/rev9-digests.json` `2253911b…c2da`
- `reviews/round-9.md` `6f1d6ca2…c6cc`

This acceptance covers only the draft's stage-3 role. It gives no authority for preparation, inventory, the OC, model calls, scans or real-repository data. Any change to the files above needs a new scoped review round with zero issues.
