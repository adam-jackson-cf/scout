# X1 preparation model-free binding implementation review — round 2

Scope: preparation action (c) only, which is the model-free binding frontend, its fixture-acceptance controller, the observer and the observer proof. I read files only. I made no model calls and ran no tests, scanners, observer proof or acceptance, and touched no real inputs. Two independent analyses fed this record: `innovation-challenger` (adversarial) and `experiment-peer` (confounds and falsifiers). How each changed this record is noted under "Council synthesis".

## Reviewed files (SHA-256 from Main; I did not recompute them)

Paths are relative to `evaluation/x1-triage-ranking/`.

| File | SHA-256 |
|---|---|
| execution/model_free/main.go | 235ce0fc95e8bffbc33ac1929c22d440838bd9db9dfebddb664f716bb852222f |
| execution/model_free/main_test.go | c013b037fb26d61374a0edff84a316bda05bafa93a428946f8271783457ca205 |
| execution/binding/acceptance.py | 5e75e907392663394f01afa9d9d0f365568c3ae0f8bdc7cd0c222aa5ec59180c |
| execution/binding/descendant_fixture.py | 669c2a58dedc063e21a617479020b2eeacbaca684fd0123b871c89b68493420c |
| execution/binding/fixtures.py | 48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b |
| execution/binding/gate.py | 72006f7874f481dcbbc7ee61c7472aca9b08c0bd7286052e39c1960794860ff0 |
| execution/binding/interlock.py | dc12b4c45be9e01f182c6253941ab51543787ba1b436a1ce037871cf9ecb4116 |
| execution/binding/observer.py | af674a9bbf22c94df02eac5970a6d1f414be99e8dc2b13ce983c588ffa905553 |
| execution/binding/observer_proof.py | 91dad45fc0e8f6fe43315dcafc8bb5566589b2e392e2d11ddf6af0e62595a4c3 |
| execution/binding/unprivileged-probe.json | 10a0c930a3b3d006e7a9c368ea7a93f709ebfda13a149b7521dc1c5721ef4881 |
| execution/binding/fixtures/common/go.mod | ec75b6e87719a439d7ec4959302e25fd588319c7f7cfd1e6e20b39a324aff0b0 |
| execution/binding/fixtures/common/index.js | 2fa9c2c5a47269dd22d1da68d24423b1a2d3a2da635c4c60798c59e17ade5d23 |
| execution/binding/fixtures/common/main.go | d37528f6027aa09adca50fa54e40ffc247c5597ffb7d4effa633ccba6134536c |
| execution/binding/fixtures/common/package.json | 514ef8299051029ada3268f1fb2eba8b698548d59bb6d438d71c42c03a627f9e |
| execution/binding/fixtures/common/unchanged.py | 534f740e38f9a8c5e7226a676a427bc5b44ec64ab605c33c595a68aedf5a97c4 |
| execution/binding/fixtures/head/.scout/policy.yaml | 6189fea6ad039f0eb0206a7986aa94a44e3bdf46669184298b8a558d06fa7255 |
| execution/binding/fixtures/head/changed.py | 0950bd4989dfc76b1c8ed8641c91b4994cefee6bc5bb6d5105510f99f25ec109 |
| execution/binding/fixtures/parent/.scout/policy.yaml | 3d3dae7d063c4fe94127de4e67ae9315e1495e4fa510ea587ae4c5a058604ec3 |
| execution/binding/fixtures/parent/changed.py | 0ecf4e2fd0b76f8305f0729807a67bc7d2eeabe27ed637a98bda8b9d87738e3b |
| execution/tests/test_binding.py | 870e6eaf514f557703399ed8649e1a401856e58ff6f9a07f387bf43a46ac05d9 |
| execution/tests/test_binding_observer.py | 645133a7a0231714b0a648c31b66c4a01432405d825f675aacb568d5afdc85db |
| execution/preparation.py (shared, read for its contract) | 1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf |
| execution/run-config.json (shared) | c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54 |
| execution/verify_inputs.py (imported) | 7d6f89ce… (abbreviated in Main's handoff) |
| preparation-reviews/binding-impl-REVIEW_REQUEST.md | b11f46ee492915771576f5bb0376535753c614f1db6c6551ecad3f7deb22eb64 |
| preparation-reviews/binding-impl-round-1.md | cbcbb8e5… (abbreviated in Main's handoff) |
| preparation-reviews/binding-impl-round-1-response.md | fd412e99… (abbreviated in Main's handoff) |

## Basis (unchanged)

- Authority: `reviews/round-9.md` 6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc. It authorises (a)–(e). The clock started 2026-11-09T03:22:58.701909Z.
- PREREGISTRATION abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb. The binding is defined in §11 item 1(c) (l.222) and item 2 (l.223).
- rev9-digests 2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da
- pins.json 6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc
- model-free-policy 87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f. This equals `frozenDigest`, main.go:28.
- model-free-binding-check 5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c
- runtime-attestation-rev6 19463957becfc23a391eac03582b32eaa0ff392069bccd9aed9000f390ad2d62
- exposure-recipe c065347e08a8571228764dae19fc686f3a6909e612a4119aad7bb58d86c5d383
- exposure-sources 67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60
- frozen-exposure-inputs ec92fba6d3f91c671301244539a8caefc5025c0d6b03efca77c5bdfc4bd96425
- Product source at base 5a8c0a155cca6bc2cbb9e71c83c7284598adc6e0. This component does not change it.

## Prior-round disposition

| ID | Status | Evidence and catching test |
|---|---|---|
| R1-01 | **Fixed** | No outer sandbox: the gate launches directly (acceptance.py:96–101). Stubs are first on PATH, and `CLAWPATCH_CLAUDE_BIN`/`CLAWPATCH_CODEX_BIN` point at them (acceptance.py:26–31). The ES allowlist is at interlock.py:35–54. The nesting probe is recorded (unprivileged-probe.json, last probe, exit 0, no failure claimed). The `denied` mode is removed (descendant_fixture.py). Tests: `test_private_env_stubs_all_provider_paths_and_no_outer_sandbox` and `test_executable_interpreter_and_script_boundaries`. The new allowlist rule caused new issue R2-02. |
| R1-02 | **Fixed as specified** | The same-session scratch proof covers: shebang interpreter, script and argv; the Python orphan; the pinned-Node spawn; contiguous sequence; exit 97; zero live descendants (observer_proof.py:23–81). Acceptance verifies the proof's pin and session before every launch (acceptance.py:83–86). Tests: `test_complete_proof_and_both_shebang_argv_shapes`, `test_proof_refuses_each_missing_boundary`, `test_runtime_pin_mismatch_refuses_before_launch`. These change one event at a time and are genuine. Where the proof sits relative to the canonical reservation is suggestion S1. |
| R1-03 | **Fixed** | One session per operation. The exact noninteractive command is at observer.py:139–141. The sequence baseline carries across trees (observer.py:186–193). Tests: `test_exact_sudoers_command_and_no_stderr_capture`, `test_sequential_trees_keep_global_sequence_between_trees`, `test_sequence_gap_between_trees_is_fatal`, `test_idle_stream_checks_sequence_without_retaining_payload`. |
| R1-04 | **Partially fixed. The fix introduced R2-01.** | There is one shared predicate (main.go:123–145) and a separate check that binding rows are skipped as not opted in (main.go:147–162; `TestBindingRowsBoundaries`, `TestDuplicateBindingRowsRefuse`). But the predicate rejects the product's healthy `AdapterIneligible` outcome, and the test hard-codes that wrong outcome. See R2-01. |
| R1-05 | **Partially fixed. Still open as Minor.** | Read against the rules, the triggers fire: <ul><li>`yaml.load(data, Loader=yaml.Loader)` matches the semgrep rule pattern.</li><li>`complex_only_at_head` has complexity 12, above ruff's `--isolated` default limit of 10.</li><li>The append loop is PERF402, not PERF401, and is still selected by `--select PERF`.</li></ul>Parent and head findings differ: PERF moves from line 4 to line 5, and C901 exists only at head. acceptance.py:208 refuses equal sets. The required "measure the trigger set on synthetic data during fixture review" has not been done, and no non-canonical way to do it exists. See the R1-05 residual below. |
| R1-06 | **Fixed** | Rows use `repo_path`/`content_sha256`/`revision`/`root` (fixtures.py:66–70, 112–114). `test_source_manifest_not_mistaken_for_filesystem_pin` materialises a fixture and calls `verify_tree_pins`. |
| R1-07 | **Fixed** | The key is the full occurrence tuple: family, sanitised category, StartLine, StartByte, path, full `AnchorOf` and rule version (main.go:32–40, 82–88). It is a superset of correlate.go:344. Links come from persisted provenance: fresh-parent bijection, one ID per head occurrence, carried IDs disclosed (main.go:96–119, 306–327). Tests: `TestOccurrenceKeyIndependentExpectedValues` (literal expected key), `TestSecretOccurrenceKeyIndependentExpectedValue`, `TestDurableLinkBoundaries`, `TestOneOccurrenceCannotLinkToTwoIssues`. |
| R1-08 | **Fixed** | Planner eligibility and path order (main.go:166–173) match plan.go:107–118 and plan.go:`sortedJobs` (key string order equals path order). Keeping nonapplicable files inside batches correctly mirrors assess.go:265–290 and runner.go:343–352. Off-file filtering mirrors `splitAdapter` (assess.go:641–670). Roots have equal length (acceptance.py:23). Inputs and coverage are compared before findings (interlock.py:100–104). Test: `TestPlannerInputsAndScannerManifest`, with independently fixed batches. |
| R1-09 | **Fixed** | `--resolve-only` returns before the registry, state or Assess (main.go:295–298). All six records are resolved before any scanner (acceptance.py:180–190), and later bytes must match exactly (acceptance.py:72–76, 120–121). Tests: `test_both_revisions_resolved_before_any_scanner`, `test_resolved_policy_must_match_later_bytes_and_pin`, `TestResolveOnlyWritesExclusiveMatchingRecordWithoutScanner`. |
| R1-10 | **Fixed** | Head `.scout` is created at 0700, copied in sorted order, with policy and symlinks excluded (acceptance.py:128–141). The reference head works because the reference parent always gains `.scout/clawpatch` (clawpatch/client.go:136). Test: `test_equal_length_roots_and_missing_head_state_directory`. |
| R1-11 | **Fixed** | `script: null` is accepted (observer.py:91–95). Test: `test_real_shape_null_and_script`. |
| R1-12 | **Fixed** | The canonical receipt is reserved exclusively and fsynced before work. `Exception` produces a terminal refused receipt and closes the journal (acceptance.py:226–239, 253–255). Test: `test_calledprocesserror_always_terminates_refused_receipt`. |
| R1-13 | **Fixed apart from two gaps, both moved to R2-01/R2-02** | Digest, policy, key, control-authorisation, multiplicity, link, executable and Node boundaries are now tested independently. Two gaps remain: `ineligible` is tested as unhealthy (main_test.go:134), and no test passes a real-shape positive-control shebang start through `validate_processes`. Both are now required catching tests under R2-01 and R2-02. |

### R1-05 residual — Minor, open

- **Finding:** the trigger set is unmeasured, and the only executable entrypoint writes the canonical receipt (acceptance.py:249–255).
- **Why it matters:** the only way to "measure during fixture review" is the irreversible canonical run. Issues R2-01 and R2-02 below are exactly the kind of composition defect such a measurement would catch.
- **Required fix:** add a non-canonical synthetic rehearsal mode to `acceptance.py`. It must:
  - use the same reviewed runtime and closure, ES session, observer proof, routes and interlock code;
  - write only under `SCRATCH`;
  - never open `OUT/binding/*` or a Journal, and never be sealed;
  - report per route: the occurrence keys and multiplicity, the scanner-health reasons, and every observed exec against the allowlist;
  - report the acceptance predicate's outcome without asserting a pass.

  Test that rehearsal mode cannot create or reserve a canonical path.

## New issues

| ID | Severity | Location | Finding (input → wrong outcome) | Required fix and catching test |
|---|---|---|---|---|
| R2-01 | Critical | model_free/main.go:131; main_test.go:134 | `scannerHealth` treats every adapter status except `completed`/`no_applicable_input` as unhealthy.<ul><li>The planner schedules every eligible file for every registered category (plan.go:107–118), and Assess batches them per spec (assess.go:265–290).</li><li>A nonempty batch with no applicable file returns `AdapterIneligible` with `not_applicable` rows and no failure (runner.go:164–172).</li><li>The synthetic fixture has no `.css`, TOML/requirements or `.github/workflows` files, so the CSS-transition, multipart-CVE and Actions-mutable-ref specs are always `ineligible`.</li></ul>Result: both routes are unhealthy on every revision, the frontend exits 1 (main.go:346), and acceptance refuses (acceptance.py:118–119). Django is affected the same way. The test encodes the wrong outcome. | Treat `AdapterIneligible` as healthy only when the adapter has no failures and all its coverage rows are `CoverageNotApplicable`. The footprint variant (runner.go:186–192) still fails, because its rows are incomplete and it carries a failure. Do not add fixture languages or drop planned nonapplicable inputs. Tests: <ul><li>Build the no-applicable result from the real `Runner` on synthetic nonmatching files. That path returns before any sandbox or scanner start (runner.go:164–172). Require it to be healthy.</li><li>Require an `ineligible` result with a failure or incomplete rows to be unhealthy.</li><li>Correct the table row.</li></ul> |
| R2-02 | Critical | interlock.py:45–48, called at acceptance.py:112 before `evaluate` at :116 | The interpreter-flag rule scans every argument, including the script's own arguments.<ul><li>The parent fixture selects `claude` (parent/.scout/policy.yaml:8).</li><li>Production starts it with `-p --output-format json …` (contextual/cli.go:24–26, 81).</li><li>The stub is a `#!<python>` script (interlock.py:17), so the ES exec has executable `python…`, script = stub, and argv containing `-p` in both accepted shapes (observer_proof.py:45–48).</li></ul>Result: `validate_processes` raises "unrestricted interpreter argument refused", so the parent positive control, and with it synthetic acceptance, always fails. The test `positive()` (test_binding.py:52–57) uses `/synthetic/claude` as the executable and never calls `validate_processes`. | For Python and Node interpreters, examine only the interpreter options before the first non-option operand, which is the script or payload. That operand must be a pinned payload or the ES `script`. Keep refusing `-c/-e/--eval/-p/--print` *as interpreter options*. Do not allow-list `-p` globally or skip validation for controls. Test: for both providers and both shebang argv shapes, run the real-shape exec through `validate_processes` then `evaluate(positive=True)`, require a pass, and require the bound predicate to reject it. Keep the negative interpreter-option cases. |
| R2-03 | Minor | observer.py:150–161, 197–218; acceptance.py:281–285; no test calls `finish()` | The controlled-stop path is an untested decision boundary that can refuse a receipt whose evidence otherwise passed. <ul><li>After `stopping` is set and `terminate()` is sent, a trailing unterminated line reaches `json.loads` (observer.py:155). That sets `tree.error`, and acceptance raises "observer session failed at completion".</li><li>If `terminate()` on the root-owned sudo front end raises `PermissionError`, the exception escapes `finish` uncaught, the receipt is refused and eslogger is left running. [INFERENCE: depends on eslogger's write buffering and on how this host's sudo treats signals; neither has been observed. The proof never reaches `finish()`.]</li></ul> | <ul><li>Define the controlled stop: every complete line is still sequence-checked; one trailing fragment with no newline after `stopping` is recorded, not failed.</li><li>If signalling fails, close the read pipe so eslogger exits on EPIPE. Then require that the process has exited, or record an explicit failure naming the live privileged process.</li><li>Test each case.</li></ul> |

## Suggestions (not blocking)

- **S1, ordering of the privileged prerequisite:**
  - Before `canonical.open('x')`, run an unprivileged `sudo -n -l /usr/bin/eslogger exec fork exit` presence check, with no output relayed.
  - Run the R1-05 rehearsal, which includes the same-session proof, before any canonical operation.
  - Reason: the host currently lacks the drop-in. A canonical run now writes a refused `synthetic.json`, and preparation.py:286 makes any refused operation block the seal.
  - Not raised as an issue, because the implementation gates this prerequisite fail-closed.
- **S2:** `validate_processes` pins only shebang `script` and Node `args[0]`, so a plain `python /unpinned.py` passes. The round-1 response says unpinned scripts are refused, which overstates this. The R2-02 operand rule fixes it if Python operands are pinned too.
- **S3:** refuse execs whose `path_truncated` is true for the executable or script.
- **S4:**
  - acceptance.py:101 calls `observer.tree.attach` directly; use `ESLogger.attach`, which checks for errors and stop.
  - Give `child.wait()` (acceptance.py:105) a timeout.
  - Add a test of the actual `Popen` argv (no wrapper), not only `frontend_argv`.
- **S5:** the proof builds its stub with the resolved Python (observer_proof.py:100), while acceptance uses `runtime['python']['path']` (acceptance.py:90). Use the identical string so the proof exercises the shebang form acceptance actually uses.
- **S6:** rename PERF401 to PERF402 in the response and tests ([ruff PERF402](https://docs.astral.sh/ruff/rules/manual-list-copy/)).
- **S7:** after a feature is removed, the bound head can report `feature removed` gaps and their omissions (inventory/diff.go:167–168; assess.go:174–179) that the history-free reference cannot produce, so `compare` would refuse. Neither frozen pair triggers this: no synthetic feature is removed, and Django tip a05e16a7230fe33d3f43f4a3aaed103cebb42efc changes only `docs/releases/6.1.txt` ([commit](https://github.com/django/django/commit/a05e16a7230fe33d3f43f4a3aaed103cebb42efc)). Document it, or compare current-derivable gaps separately.
- **S8 (round-1 S3, still open):** confirm heuristic-only mapping from the observed Clawpatch exec argv as well as from the constant `mapping_argv` (main.go:263).
- **S9:** add a direct test of off-file observation filtering (main.go:221).

## Council synthesis

- **Mine:**
  - Found R2-02 independently before either peer reported.
  - Confirmed R2-01 against runner.go:164–172 and plan.go:107–118.
  - Checked that correlation runs on partial runs (assess.go:466) and that Clawpatch review is blocked without egress before any start (clawpatch/review.go:177–179), so the contextual stubs produce the positive-control attempts.
  - Confirmed policy.Load reads from git (policy.go:141–163).
  - Found S7 and established that the Django tip commit only touches docs.
- **Challenger:**
  - Raised R2-01 and R2-02, which I confirmed.
  - Rated R2-02 Major. I rate it Critical, because it deterministically prevents acceptance, which is round 1's Critical criterion.
  - Its note on testing the real launch wrapper became S4.
- **Peer:**
  - Raised R2-01, R2-02, the PERF402 correction (S6) and S2–S5.
  - Its N4 became R2-03. I accepted it as Minor, matching round-1 R1-11, which was an [INFERENCE]-level host-shape parser issue.
  - Its N3 (proof after reservation) became S1, not an issue. The prerequisite is gated fail-closed, and moving the reservation would not avoid the reviewed-recovery rule, because preparation.py:286 blocks the seal on any refused operation. I changed my recommendation to add the unprivileged presence check and the rehearsal.
  - Its S6 (non-canonical dry run) became the required fix for the R1-05 residual.

## Verification

- **Read:**
  - all listed component files;
  - preparation.py (pins, journal, seal);
  - product source: engine/assess.go, engine/model_cli_policy.go, inventory/plan.go and diff.go, analyzers/runner.go, domain/types.go and issue_anchor.go, issues/correlate.go, state/store.go, policy/policy.go, contextual/cli.go and client.go, modelcli/modelcli.go, clawpatch/client.go and review.go, Clawpatch id/mapper sources;
  - PREREGISTRATION l.222–223 and model-free-binding-check.json.
- **External:** GitHub commit page for the Django tip.
- **Not done:** I did not recompute any SHA-256, build, test, probe or run anything. Main's report (121 Python tests OK, `go vet` clean, `go test` ok) is accepted as given. The passing suite does not catch R2-01 or R2-02: the first is hard-coded into a test, and the second's real-shape composition is untested.
- **Host:** Main reports `/etc/sudoers.d/` is empty and that `sudo -n /usr/bin/eslogger exec fork exit` asks for a password. The implementation gates this fail-closed (observer.py:141–143).

## Verdict

**4 issues open:** 2 Critical (R2-01, R2-02) and 2 Minor (R1-05 residual, R2-03).

Component (c) must not be run: not the observer proof, not a rehearsal, not a canonical acceptance. Each Critical issue on its own makes synthetic acceptance impossible. Because the controller reserves the canonical receipt first, a run now would permanently write a refused `binding/synthetic.json`.

Round 3 needs:
- the fixes and their catching tests, with new digests;
- the rehearsal mode;
- the owner's prerequisites actually in place: Full Disk Access for the launching app, and `/etc/sudoers.d/x1-eslogger` with exactly `adamjackson ALL=(root) NOPASSWD: /usr/bin/eslogger exec fork exit`, mode 0440, checked with `visudo -cf`.

A round-3 sign-off would first authorise only the scratch rehearsal. The canonical synthetic acceptance follows only after its report shows:
- nonempty, differing parent and head occurrence sets;
- healthy routes;
- every observed exec inside the allowlist.
