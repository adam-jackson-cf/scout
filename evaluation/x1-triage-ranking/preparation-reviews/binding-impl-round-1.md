# X1 preparation model-free binding implementation review — round 1

Scope: X1 preparation action (c) only, which is the model-free binding implementation and its fixture-acceptance harness. The review was read-only. I made no model calls, built nothing, ran no tests and touched no real inputs. Two independent analyses fed this record: an adversarial review by `innovation-challenger` and a confounds/falsifiers review by `experiment-peer`. Where they changed my conclusions is noted in the council synthesis section.

## Reviewed files (SHA-256, paths relative to `evaluation/x1-triage-ranking/execution/`)

Main computed these digests. I could not run a shell, so I did not recompute them. Main must confirm the bytes are unchanged before round 2.

| File | SHA-256 |
|---|---|
| model_free/main.go | d285dc5c6f173c65ac33f2f8519c303c140df408badbbb132ad6757731b03773 |
| model_free/main_test.go | 90841bdc62e081458ee1c1d9d8937f04033aa649700e38cadf75dde6b4c48b69 |
| binding/REVIEW_REQUEST.md | 1811759c5a44c65620f8b13e03ea23152868dac149d8d43d3f4fd3f0e8c24bc7 |
| binding/acceptance.py | b1d81d5366f1d5802183a3b30f59801eecbbad079b18cea01e392c996207cbc3 |
| binding/descendant_fixture.py | 07e9a58d735b5b6bacb14957f644cba07d2a332ea24ca3226df6cc586a0d1ddb |
| binding/fixtures.py | c84b8ad7883efb3f57d527bd16cd8eb58f60f6f05810e158b730fea1d6395c76 |
| binding/gate.py | 72006f7874f481dcbbc7ee61c7472aca9b08c0bd7286052e39c1960794860ff0 |
| binding/interlock.py | 8279db5d34e7ea1fdeaab00da13554f1522400e9a5772079f7c243ab4cda5226 |
| binding/observer.py | 2254ff9f6e75aea28de01bd5ae5ebbf79d4021fdbca575e6aa0d5ea1f0eaaae9 |
| binding/unprivileged-probe.json | b56da4e245a76137a250ccd8d62132135bc62f341bfad9f22b3a6f454f721e1e |
| binding/fixtures/common/go.mod | ec75b6e87719a439d7ec4959302e25fd588319c7f7cfd1e6e20b39a324aff0b0 |
| binding/fixtures/common/index.js | 2fa9c2c5a47269dd22d1da68d24423b1a2d3a2da635c4c60798c59e17ade5d23 |
| binding/fixtures/common/main.go | d37528f6027aa09adca50fa54e40ffc247c5597ffb7d4effa633ccba6134536c |
| binding/fixtures/common/package.json | 514ef8299051029ada3268f1fb2eba8b698548d59bb6d438d71c42c03a627f9e |
| binding/fixtures/common/unchanged.py | d48e6536db1160b997626ab56e18cc0a616ff6f7ddd09c0dd978dd0dad28b6cf |
| binding/fixtures/head/.scout/policy.yaml | 6189fea6ad039f0eb0206a7986aa94a44e3bdf46669184298b8a558d06fa7255 |
| binding/fixtures/head/changed.py | 9403b4c63b9c64adf4569bc2cac670d3572138cb6d485dc552916229e83b02c3 |
| binding/fixtures/parent/.scout/policy.yaml | 3d3dae7d063c4fe94127de4e67ae9315e1495e4fa510ea587ae4c5a058604ec3 |
| binding/fixtures/parent/changed.py | 66976b08ac880af1f462c07aa630f8eb659127b8ca873c036dcfde0b5ac15041 |
| tests/test_binding.py | b4d0c4a8442947bd0d3f37025818d4f68ad6603911b6073f0d3c49b6190ed6aa |
| preparation.py (shared) | 1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf |
| run-config.json (shared) | c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54 |

## Basis digests

- Authority: `reviews/round-9.md` 6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc. It authorises (a)–(e) only. The 45-day clock started 2026-11-09T03:22:58.701909Z.
- `PREREGISTRATION.md` abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb. The binding is defined in §11 item 1(c) and item 2 (l.222–223).
- `reviews/rev9-digests.json` 2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da
- `preparation/pins.json` 6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc
- `preparation/model-free-policy.json` 87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f. This matches `frozenDigest` at main.go:25.
- `preparation/model-free-binding-check.json` 5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c
- `preparation/runtime-attestation-rev6.json` 19463957becfc23a391eac03582b32eaa0ff392069bccd9aed9000f390ad2d62
- Worktree `/Users/adamjackson/Projects/work/scout-x1prep`, base 5a8c0a155cca6bc2cbb9e71c83c7284598adc6e0. Product source was read at that base and is unchanged by this component.

## Judgements on the four questions

**(1) Does the frontend bind the frozen policy at both revisions, using existing APIs and no product edits? Yes.**
- Each revision is a separate frontend invocation. It digest-checks and parses the frozen policy and refuses anything that is not model-free (main.go:109–113). It then calls `engine.ApplyModelCLIPolicy`, which sets lanes, CLI and Clawpatch provider (`internal/engine/model_cli_policy.go:17–25`).
- `engine.Assess` never loads a policy itself (`internal/engine/assess.go:27–513`). Only `diff.go:152` and `cmd/scout/ranking.go:20` call `policy.Load`, and the frontend uses neither.
- Association is off: a nil `Options.Association` gives a zero `LaneOptions`, so the disabled return fires (assess.go:491–501).
- Mapping is hard-coded heuristic (assess.go:551, 578).
- With `ModelLanes=[]`, contextual and Clawpatch-review jobs become `skipped (not opted in)` (assess.go:336–339, 358–361).
- `modelcli.Detect` and `Config.Fingerprint` only look the CLI up on PATH; they never start it (`internal/modelcli/modelcli.go:44–61`).
- No ranking or decisions stage is reached.

**Is the reference route structurally independent? Yes, but its inputs are not equivalent.** It calls `clawpatch.New`/`Map(heuristic)`, `inventory.Scan`, `Registry.Specs`, `Spec.Batches` and `Runner.Run` (main.go:79–94). It never touches the override, `ApplyModelCLIPolicy`, `Assess`, correlation, ranking or association. Its inputs differ from Assess's; see R1-08.

**(2) Issue identity: I accept the reading, with an amendment (R1-07).**
- Durable IDs cannot match between independent runs. They are minted by `mintID = domain.NewIssueID(family, time.Now())` (`internal/issues/correlate.go:28–29`), assigned at :413, and each is a millisecond timestamp plus 80 random bits (`internal/domain/issue_anchor.go:77–100`).
- Comparing the deterministic identity Scout itself computes is therefore the only faithful reading of "issue-ID sets". Scout derives that identity with `domain.AnchorOf` (correlate.go:322).
- The current key is incomplete, though. Scout mints one durable issue per occurrence, keyed on `{family, category, RuleKey, Secret, Fingerprint, path, StartLine, StartByte}` (correlate.go:344). The frontend key (main.go:27–31, 68–75) leaves out category, StartLine and StartByte, and does not tie the bound route's actual durable issues to the compared set.

**(3) Process observation: the current design cannot pass. I recommend the setup below** (R1-01 to R1-03, R1-11).
- An eslogger NOTIFY exec/fork/exit consumer is the right observer for successful starts.
- The worker's hard gate demands complete argv for *denied* execs (observer.py:120–124). That goes beyond the signed text: argv capture applies to "child-process starts", and attempts are captured by the logging stubs (binding-check.json `required_checks_per_fixture[2]`).
- The denied-attempt gap exists only because an outer `sandbox-exec` denies process-exec. That outer sandbox also breaks every product scanner, because macOS refuses to apply a second sandbox inside a process that is already sandboxed (R1-01).

**Recommended minimum privileged setup, provided by the owner:**
1. **Full Disk Access** for the app that launches the controller (Terminal or Orca). `man eslogger` requires it for the responsible process.
2. **A temporary sudoers drop-in limited to one exact command.** Create `/etc/sudoers.d/x1-eslogger` with mode 0440, check it with `visudo -cf`, and add:
   `adamjackson ALL=(root) NOPASSWD: /usr/bin/eslogger exec fork exit`
   Remove it after the last (c) acceptance run. Journal its presence and removal, never any credential.
3. **Nothing else is needed:** no `log stream`, DTrace, SIP change or AUTH-exec entitlement. Once the outer sandbox is gone, no exec is denied, so every exec the kernel performs is an ES NOTIFY_EXEC event with argv and sequence-loss accounting. Model-name launches always land on stubs:
   - the contextual lane resolves the CLI through PATH (modelcli.go:54–58), where the stubs are first (acceptance.py:74);
   - Clawpatch's claude provider honours `CLAWPATCH_CLAUDE_BIN` (`third_party/clawpatch/src/providers/claude.ts:320–323`), which must point at the stub;
   - a lookup of a nonexistent path creates no process and cannot reach a model.

This covers "capture all child-process starts and argv", the positive control (an ES-observed stub start correlated with the stub's own log) and the bound zero-start check.

**(4) Positive control and gates.**
- The production path to a stub is genuine: unbound policy → contextual lane → `provider.Review` (assess.go:341) → `prerequisites`/`Detect` → `exec.CommandContext(binary)` (`internal/contextual/cli.go:81`).
- It depends on the heuristic map producing a feature from `go.mod`/`main.go` (`third_party/clawpatch/src/mappers/go.ts:19–35`). [INFERENCE] I have not observed this.
- Gates fail closed, with one exception: R1-12.
- Acceptance cannot currently pass at all, because of R1-01, R1-02, R1-04 and R1-05.

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R1-01 | Critical | acceptance.py:79–83; interlock.py:14–28 | The frontend runs inside an outer `sandbox-exec -f`. Every product scanner then runs `sandbox-exec -p …` itself (`internal/sandbox/sandbox.go:45–49`, `internal/analyzers/runner.go:554, 570`). macOS refuses a nested sandbox (`sandbox_apply: Operation not permitted`; external source: anthropics/sandbox-runtime#67). Every adapter fails, so no bound or reference run can be healthy, on either fixture. The outer exec-deny rule is also the only reason denied attempts can go unseen. | Remove the outer `sandbox-exec` wrapper and its exec-deny profile. Keep the private HOME/XDG/TMPDIR and minimal env. Set `CLAWPATCH_CLAUDE_BIN=<stubdir>/claude` in `private_env` for every route. Keep stubs first on PATH. Enforce with ES: the existing check at acceptance.py:101–113 must refuse any start outside the attested executables plus the stubs. The product scanner sandbox still confines scanner network and writes. Record the unprivileged nesting probe `sandbox-exec -p '(version 1)(allow default)' /usr/bin/sandbox-exec -p '(version 1)(allow default)' /usr/bin/true` in `unprivileged-probe.json`. Drop the `denied` mode from `descendant_fixture.py`. |
| R1-02 | Critical | observer.py:120–124; acceptance.py:240–246 | `require_attempt_coverage` always raises. As built, no passing receipt is possible: required behaviour (complete process observation feeding acceptance) is missing, not merely gated. Running the controller now would permanently use up the canonical receipt with a refusal. | After R1-01, delete the gate. Replace it with a pinned, non-canonical observer-proof run (scratch output, same ES session and parser) that is required before acceptance. It must show: a shebang stub start (interpreter executable, script and args shape); a grandchild still visible after its parent exits; a pinned-Node `child_process.spawn` grandchild; contiguous `global_seq_num`; and no live descendants at the end. Acceptance must verify that proof's pin. |
| R1-03 | Major | observer.py:126–136; REVIEW_REQUEST.md:44 | `sudo -n eslogger` starts once per revision run (6 for synthetic plus a preflight, 4 for Django) from a new session with no controlling terminal. [INFERENCE] macOS sudo ties its cached credential to the terminal and expires it after 5 minutes by default, so `sudo -v` followed by `sudo -n` fails. That gives "eslogger unavailable" mid-operation and an irrevocable refusal. | Use the sudoers drop-in from setup step 2. Start one eslogger session per acceptance operation and attach the revision trees one after another, so sequence contiguity covers the whole operation. Update REVIEW_REQUEST's privileged instructions (remove `sudo -v` and `log stream`). |
| R1-04 | Critical | main.go:139, 147–151 vs 80–91 | Bound health is `Run.State==RunComplete`, but `skipped` does not count as accounted (`internal/domain/types.go:131–137`). If the map finds features, every contextual and review job is skipped (assess.go:336–339, 358–361; `internal/inventory/plan.go:128–146`). If it finds none, mapping gaps become omissions (assess.go:174–179). Either way the run stays partial (assess.go:435–450): the frontend exits 1 and acceptance.py:115 refuses, Django included. The reference route uses a different health rule (main.go:84, 91). | Use one deterministic-scanner health rule for both routes: every registered spec present; every file-analysis coverage row accounted; no adapter, inventory or source-drift failure; map status not failed; omissions and mapping gaps recorded and required to be equal between routes, not zero. The bound route must also show every feature and review row as `skipped` with reason `ReasonNotOptedIn`, which records the binding as an observation. Unit-test each boundary. |
| R1-05 | Critical | fixtures/common/unchanged.py, parent/changed.py, head/changed.py, index.js, main.go | No registered rule matches the fixtures. Registered rules: semgrep unsafe YAML loaders (`internal/analyzers/rules/security/semgrep-python-unsafe-yaml-load.yaml:14–18`), multipart CVE, `document.currentScript`, Actions mutable refs, CSS transitions, gitleaks, ruff C90 (default threshold) and PERF (`catalog_quality.go:18`, `catalog_performance.go:27`). The fixtures contain `eval`, unused imports and `== true`. [INFERENCE] Both sets will be empty, so the nonempty parity check (interlock.py:85–86) always fails. The sets are also the same at parent and head, so a parent/head swap could not be detected. | Add harmless triggers. `yaml.load(data, Loader=yaml.Loader)` in `unchanged.py`, unchanged at both revisions (bare `yaml.load(data)` does not match the rule). A PERF401 append loop in `changed.py`, moved by one line at head. A function with complexity above 10 added only at head. Keep `go.mod`/`main.go` for the heuristic feature. Measure the trigger set on synthetic data during fixture review. |
| R1-06 | Major | fixtures.py:69–71, 113–114; acceptance.py:197, 179; preparation.py:158–169 | Repository-relative `{path, sha256}` manifest rows go into the receipt. `verify_tree_pins` treats every such dict as a filesystem pin and resolves it against the working directory. The Django prerequisite (acceptance.py:179) and the seal therefore refuse valid evidence. | Store manifest rows under keys that are not pins (e.g. `repo_path`, `content_sha256`), tagged with their revision and absolute root. Add a test that runs `verify_tree_pins` on a synthetic receipt. |
| R1-07 | Minor | main.go:27–31, 68–75; interlock.py:81–96 | The comparison key is not Scout's complete deterministic issue identity (correlate.go:344). An evidence location shift or a category remap would pass undetected. The bound route's durable issues are only counted. | Key on the correlate.go:344 occurrence tuple plus the full `AnchorOf` and `rule_version`, keeping multiplicity. At parent (fresh store), require the bound durable issues to map one-to-one onto those keys. At head, require each key to link to exactly one durable issue in that run. Disclose IDs carried over from parent. |
| R1-08 | Minor | main.go:86–91 | The reference batches every inventory file through every spec (main.go:88), while Assess batches only planned files in job order (assess.go:261–286). Root path lengths also differ by route, and batching counts `len(root)` (`internal/analyzers/batch.go:32–37`). On Django the 96 KB argv cap forces multiple batches, so boundaries differ. Separately, `splitAdapter` drops off-file observations (assess.go:684–689). The result could be a false, irreversible refusal, and nothing shows the two routes scanned identical inputs. | Filter reference inputs the way the planner does (`Spec.Applies`, not excluded, readable). Use equal-length roots for both routes. Record each adapter's accounted coverage keys in both receipts and require them equal before comparing findings. |
| R1-09 | Minor | acceptance.py:139–160 | The signed text requires both effective policies to be captured before any fixture scanner runs. Head's effective policy is recorded only after the parent scans. | Add a frontend `--resolve-only` mode that writes the effective record without calling Assess. Run it for both revisions (bound and synthetic-unbound) before any scan, and require each later per-run record to match it byte for byte. |
| R1-10 | Minor | acceptance.py:123–136 | Django's head has no tracked `.scout/`, and `iterdir` order is arbitrary. If a file such as `state.sqlite` comes first, `shutil.copy2` raises FileNotFoundError and the run refuses. | Create `head/.scout` (mode 0700) when it is absent. Copy in sorted order. Add a test for a head with no `.scout`. |
| R1-11 | Minor | observer.py:85–86 | [INFERENCE] eslogger writes `"script": null` for execs that are not scripts. `event['script']['path']` would then raise TypeError, and every run would refuse. The parser has never been checked against real output. | Use `s = event[kind].get('script')` and record only when it is not null. Add real-shape fixtures to the tests and validate through the R1-02 proof. |
| R1-12 | Minor | acceptance.py:254 | The except clause misses `subprocess.CalledProcessError` (raised by the git helpers), `TimeoutExpired`, TypeError and others. Then no refused receipt is written, the journal entry stays open, and the same canonical path can be re-run. That weakens the create-once refusal safeguard. | Reserve the canonical receipt exclusively before doing any work, or catch `Exception` and always write a refused receipt. Test it by injecting a `CalledProcessError`. |
| R1-13 | Minor | main_test.go:11–24; tests/test_binding.py | Decision boundaries with no tests: digest mismatch, the not-model-free refusal, key present, the `X1_SYNTHETIC_CONTROL` gate, the enabled-policy precondition, the health function, the executable allowlist, interpreter-argument and Node-payload refusals, multiplicity mismatch, and a positive control whose attempt pid does not match. `TestCanonicalAnchorUnchangedAndMultiplicity` is circular. | Add table tests for each boundary, with expected outcomes fixed independently of the code under test. |

## Suggestions (not blocking)

- **S1, defence in depth after R1-01:** run the frontend tree as a dedicated standard (non-admin) account that has never logged in to claude or codex, optionally with a pf rule blocking that user's outbound traffic. Then an accidental absolute-path real-CLI start cannot authenticate. No reachable code path makes such a start (modelcli.go:44–61; claude.ts:320–323), so this is residual risk only.
- **S2, positive-control attribution:** require claude at parent and codex at head, the production contextual argv (`claudeArgs`/`codexArgs`, cli.go:40–58), frontend ancestry and stub exit 97. Also run the bound-route predicate (`evaluate(..., positive=False)`) on the control's observation and require it to raise.
- **S3:** confirm heuristic-only mapping from the observed Clawpatch argv (`map --source heuristic`, never `review` or `--source agent`), in addition to the constant at main.go:132.
- **S4:** cross-check descendant attribution against the exec target's `parent_audit_token`/`ppid`. Treat any tracked process that died from a signal, or forked without exec, as a diagnostic.
- **S5:** check the `script` field of interpreter execs against the pinned payload set. Add python `-c` to the restricted interpreter arguments.
- **S6:** compute reference keys after the same JSON round trip the store applies (invalid UTF-8 in quotes).
- **S7:** derive each of the six acceptance flags from its own evidence instead of setting them all true at acceptance.py:250.

## Council synthesis

- **Mine:** R1-07, R1-10, and confirming R1-01 against product code (`internal/sandbox`, runner) plus an external source.
- **Challenger:** raised R1-04, R1-05, R1-06, R1-08 and S2. It recommended keeping the exec-denying sandbox; I rejected that because of R1-01. It also confirmed that the frontend binding is structurally sound.
- **Peer:** raised nested sandboxing as an inference, which I confirmed, plus R1-03, R1-09, R1-11, R1-12 and R1-13. It proposed a SIGKILL deny modifier; I did not adopt it, because the modifier is unverified and becomes moot once the outer sandbox goes. Its fixture example `yaml.load(data)` was wrong; corrected in R1-05.
- **Agreement:** all three accept the issue-identity reading with the occurrence-key amendment, and agree that capturing argv for denied attempts goes beyond the signed text.

## Verification

- Read the reviewed files and the relevant product source: engine assess/plan/lanes/model_cli_policy, contextual, modelcli, issues/correlate, domain issue anchor and types, analyzers batch/runner/catalogs/rules, inventory scan/plan, sandbox, clawpatch client, Clawpatch providers, `cmd/scout` main and ranking.
- Read the preregistration (l.218–228), the binding-check recipe and the frozen policy.
- External sources: anthropics/sandbox-runtime#67 (nested `sandbox-exec` fails on macOS) and Apple Endpoint Security documentation for fork/exec events.
- I did not recompute any SHA-256, build, test, probe or run anything. Main's report of 58 tests OK and clean `go vet`/build/test is accepted as given, but it does not cover any of the issues above.

## Verdict

**13 issues open:** 4 critical (R1-01, R1-02, R1-04, R1-05), 2 major (R1-03, R1-06) and 7 minor. Component (c) must not be run. Running `binding/acceptance.py` now would permanently write refused receipts at the canonical paths. After fixing, re-review in round 2 with new digests, the updated `unprivileged-probe.json` including the nesting probe, and the owner-provided setup in place (Full Disk Access plus the sudoers drop-in). That review must reach zero issues before the observer-proof run or any fixture scan.
