# Semgrep 1.182.0 toolchain compatibility review

Source ledger: R-110 (`docs/research-ledger.md`).

**Recommendation: upgrade from 1.100.0 to 1.182.0 with the approved macOS null-sink permission.** Corpus findings and explicit target sets are unchanged. Sandboxed CLI and internal package tests pass after allowing writes to exactly `/dev/null`; all other scanner write restrictions remain. This is a toolchain compatibility review, not a Scout design experiment. Rule definitions and `docs/features/**` are unchanged.

## Selection and scope

The original candidate was 1.185.0. It is unsuitable for Scout's existing release targets: the upstream 1.183.0 changelog says Intel Mac binaries are no longer shipped and 1.182 is the last supported release. Scout's release matrix includes darwin/amd64 and its bootstrap accepts Darwin/x86_64. The owner therefore selected 1.182.0. A later move past 1.182 needs a separate platform decision; Docker would not preserve the current native scanner invocation.

Sources fetched in this session: [upstream changelog](https://github.com/semgrep/semgrep/blob/v1.185.0/CHANGELOG.md), [chosen release metadata](https://pypi.org/pypi/semgrep/1.182.0/json), and [rejected release metadata](https://pypi.org/pypi/semgrep/1.185.0/json). Copies are `raw/upstream-changelog.md` and `raw/pypi-*.json`. The rejected candidate installation and source-distribution listing are retained as selection evidence, not evidence of the adopted version's behaviour.

The worktree has its own `.tools` install, created by `scripts/bootstrap-tools.sh`; the main checkout's tools were not modified or symlinked. Baseline uses the base commit's hash lock in `.tools/semgrep-baseline`. Candidate uses the relocked worktree `.tools/venv`. Both are Python 3.11.11. Installation logs and `raw/runtime.txt` record the environment. Every measured number here comes from this session.

## Compatibility table

| Surface | Observation | Scout impact |
|---|---|---|
| CLI | `scan --config <local rules> --disable-version-check --metrics=off --no-git-ignore --json --error <explicit files>` works on both versions. Sample returns exit 1 for findings on both. | Invocation and exit interpretation unchanged. |
| JSON | Both samples contain `version`, `results[].check_id/path/start.line/end.line/extra.message/extra.severity`, `errors`, and `paths.scanned`. | These are all fields parsed by `internal/analyzers/semgrep.go`. Bump every catalog binary and `Output.SchemaVersion` to 1.182.0; otherwise exact schema equality rejects valid output. |
| Bundled rules | Candidate validation: 5 rules, 0 configuration errors (`raw/validate-after.log`). | All bundled rules accepted, with definitions unchanged. |
| Login | Upstream 1.101 changes fingerprint/metavariable availability; these extras are not parsed by Scout. Sample needs no login. | Parsed finding identity uses rule/path/line/message, not login-gated extras. Authenticated/Pro scans were not tested. |
| Python | Metadata requires Python >=3.10; installed Python 3.11.11 successfully runs version, validation and scans. | Existing Python pin qualifies. |
| Platforms | 1.182 metadata includes macOS x86_64 and arm64, Linux manylinux x86_64 and aarch64, plus musllinux wheels. Linux manylinux minimum is glibc 2.34. | All Scout release architectures have wheels. CI ubuntu-24.04's glibc 2.39 meets that floor. Older Linux hosts below glibc 2.34 cannot use these manylinux wheels; Linux and Intel Mac runtime execution was not measured locally. |

## Target selection and matcher changes

Semgrepignore v2 became the default in upstream 1.117. Scout does not ask Semgrep to recursively discover the repository: its catalog supplies `{input_paths}`, expanded into individual manifest paths by `internal/analyzers/runner.go:expandArgs`, and also passes `--no-git-ignore`. Thus directory-walk defaults must not be confused with Scout's invocation.

`run_compat.py` exercises all bundled rules with explicit files inside a Git repository, including a hidden path, a Git-ignored directory and a `.semgrepignore`-excluded directory. Both versions attest the same 9 scanned paths and return the same 8 findings, including ignored and hidden inputs; the safe YAML sample stays clean. Field-by-field comparison covers check ID, path, start/end line, message and severity. Both have empty errors. Raw JSON, stderr, command arguments and exact comparison are in `raw/sample-*`.

The upstream release range contains parser/matcher changes and the 1.168 switch to PCRE2. Scout's bundled rules include Python structural patterns plus generic/regex matching for Actions, multipart, JavaScript runtime text and CSS. Validation alone does not establish semantic equivalence: the corpus negative boundaries and positive mechanisms plus the all-rule sample provide measured regression coverage. They showed no user-visible difference. This is not an exhaustive equivalence proof over arbitrary source syntax.

## Corpus regression and sample performance

`go test -count=1 -v ./internal/eval` passes before and after (`raw/corpus-before.log`, `raw/corpus-after.log`). Both replay 9/9 true issues detected and 13/13 false positives unflagged. The live bundled-rule regression also passes both times, including its negative boundaries. This replay is recorded corpus evidence, not a fresh scan of the original remote repositories.

The same single sample takes 5.64105337485671 seconds before and 1.4537107918877155 seconds after (wall clock, `raw/sample-summary.json`). This is **provisional, n=1 per version**, with no interval, not a measured improvement or a benchmark; caches/order/machine load are uncontrolled. Live corpus test times are recorded separately in the raw Go logs and are not a performance comparison.

### Paired CLI scenario timing after the GitHub timeout

At the owner's request, run `go test -count=1 -json -p 1 -timeout 600s ./cmd/scout` once per version, without competing test suites. The upgraded branch ran first; baseline ran in a detached `origin/main` worktree at `d1fdcf38` with its own freshly bootstrapped `.tools`, then that scratch worktree was removed. Both pass. Raw `pass` events are `raw/cli-timings-1.100.0.jsonl` and `raw/cli-timings-1.182.0.jsonl`; all matched test deltas are `raw/cli-timing-comparison.json`. This comparison is **provisional, n=1 per version**, with no interval: single sequential local observations, not repeated or randomized measurements.

| Package/scenario | 1.100.0 seconds | 1.182.0 seconds | Delta seconds |
|---|---:|---:|---:|
| Entire cmd/scout package | 135.904 | 122.243 | -13.661 |
| TestDiffGateScenario | 44.84 | 38.43 | -6.41 |
| TestDiffDeltaIsReportOnlyUnderGuardianProfile | 24.98 | 22.72 | -2.26 |
| TestDiffReportsGuardrailErosion | 9.01 | 7.56 | -1.45 |
| TestDiffUsesPublishedBaselineUnlessDigestsDiffer | 20.96 | 19.56 | -1.40 |
| TestAssociationCLIOptionsSkipReasons | 6.10 | 6.82 | +0.72 |
| TestAssociationCLIOptionsPolicyAndPushRange | 2.50 | 3.18 | +0.68 |
| TestModelCLILaneOptedInUsesPolicyProviderNotFirstOnPath | 5.78 | 5.95 | +0.17 |

The final three rows are the largest observed top-level scenario increases; nested subtests are included in the JSON rather than double-counted here. The principal scanner scenarios have negative observed deltas in this one pair, not a measured improvement. The pair does not establish a performance change or explain why Linux GitHub runs exceeded the existing 150-second package deadline. No invocation flags, tests or deadlines were changed on the basis of these timings. The CI timeout is pass/fail evidence that the deadline was exceeded, not a performance conclusion; GitHub CI remains a separate acceptance requirement.

## Dependency outcome and remaining risk

The repository's normal `scripts/bootstrap-tools.sh --relock` path uses uv universal hash compilation, with no overrides or relaxed upstream bounds. The resulting lock contains:

- Semgrep 1.182.0;
- setuptools 84.1.0;
- protobuf 6.33.6;
- opentelemetry-api 1.37.0.

The chosen Semgrep metadata pins API/SDK/HTTP exporter to `~=1.37.0` and requests/threading instrumentation to `~=0.58b0`; this is the same relevant dependency regime as 1.185.0. Installation resolves fixed protobuf and runs with fixed setuptools. A source search of installed OpenTelemetry contains no `pkg_resources` reference (`raw/runtime.txt` command records the installed versions; the search yielded no matches). Version, validation and sample scan succeed without the legacy import failure. Tracing-enabled execution was not tested.

The protobuf and setuptools findings are addressed by these locked versions. Two API advisories remain: [GHSA-mf5v-w3w3-f6cg](https://github.com/advisories/GHSA-mf5v-w3w3-f6cg), fixed in 1.45.2, and [GHSA-xcq7-p3pf-p88h](https://github.com/advisories/GHSA-xcq7-p3pf-p88h), fixed in 1.45.1. Neither fixed release satisfies `~=1.37.0`. They require the separately reviewed, dated owner exception until 2027-01-15; this review does not itself suppress or authorize them. [INFERENCE] Offline scans without tracing have less direct exposure than a server parsing inbound trace headers, but the vulnerable dependency remains imported and non-reachability is not established.

## Reproduce

From the repository root, initialize submodules and run `scripts/bootstrap-tools.sh --relock`, then `source .tools/activate`. Recreate the baseline venv with uv and the base commit's `toolchain/python-requirements.lock`; run `.tools/venv/bin/python evaluation/semgrep-1.182-compat/run_compat.py`. Validate both bundled rule directories with `semgrep scan --validate --config internal/analyzers/rules/security --config internal/analyzers/rules/performance --metrics=off --disable-version-check`. Run `go test -count=1 -v ./internal/eval` for the corpus. Scoped acceptance is `go test -count=1 ./cmd/scout ./internal/...` (passes, `raw/go-test-supported.log`); `raw/sandbox-after.log` records the expanded confinement scenario passing. The original full-run failure is retained in `raw/go-test-all.log`; repository gates are retained as `raw/gate-*.log`, and exact-head CI is reported on the PR.

## Security boundary

The full acceptance run found CLI scenarios returning `output_schema`. A scrubbed-environment standalone scan still produces version 1.182.0, disproving an initial hypothesis that the native launcher selected an unpinned host `pysemgrep`. Adding the actual macOS sandbox profile reproduces the failure: `semgrep/rpc.py` calls `subprocess.Popen(stderr=subprocess.DEVNULL)`, and Python opens `/dev/null` with `O_RDWR`; the existing policy denies file writes outside the private scanner state. The process exits 2 with no JSON output (`raw/sandboxed-launcher.json`, `raw/sandboxed-launcher.stderr.txt`). The sanitized-only scan is `raw/sanitized-launcher.json`.

The owner approved allowing writes to exactly `/dev/null` in the macOS profile: `(allow file-write* (literal "/dev/null"))`, after the state write-deny rule. Linux already supplies a private writable `/dev` through `--dev /dev`; the decision brings the null sink into parity without granting any other device, path or subpath. Network denial, private-state-only filesystem writes, read roots and the scrubbed environment otherwise remain unchanged. The existing sandbox confinement scenario now checks read/write opening of `/dev/null`, refusal of `/dev/zero`, and refusal of ordinary writes outside state. No PATH change was needed or adopted.

The full `go test -count=1 ./...` log also includes the separately confirmed pre-existing X11 frontend absolute-path pin `KeyError`; its signed evidence is untouched. The owner scoped acceptance to the gate's packages plus `./cmd/scout ./internal/...`; `raw/go-test-supported.log` records that command after the null-sink fix. The original full-run log remains `raw/go-test-all.log`.

An initial post-fix CI-gate attempt ran concurrently with the separate supported-package acceptance and hit the existing CLI package timeout (151.252 seconds recorded in `raw/gate-ci-concurrent-timeout.jsonl`). No assertion or timeout was changed. A subsequent gate invocation runs without the competing suite; its output is `raw/gate-ci.log`. These concurrent-run timings are not performance evidence.

### Linux bubblewrap diagnostic

The owner requested a Linux reproduction after GitHub exceeded the package deadline. An independent scratch clone on the NUC, outside `~/.omp`, uses its own bootstrap-installed tools and pinned Go 1.26.11. The host is Linux x86_64 (Omarchy), glibc 2.44 and bubblewrap 0.12.0, not the Ubuntu GitHub runner; identities and installation outputs are in `raw/linux-environment.txt`, `raw/linux-go-identity.txt` and `raw/linux-bootstrap.log`. An initial invocation without the explicit mise Go selection failed before testing and is retained separately.

`go test -count=1 -json -p 1 -timeout 600s -run 'TestDiffGateScenario|TestDiffDelta' ./cmd/scout` passes both requested scenarios: DiffDelta 31.71 seconds, DiffGate 59.36 seconds, package 91.077 seconds (`raw/linux-cli-scenarios.jsonl`). These timings are **provisional, n=1**, with no interval or performance conclusion.

A direct Semgrep YAML sample under the same bubblewrap boundary (`--die-with-parent`, `--unshare-net`, read-only `/`, private writable scanner state, read-only rules, `--unshare-pid`, private `/proc` and `/dev`, scrubbed environment with HOME/TMPDIR set to state) returns the expected findings exit 1, one finding, `errors:[]` and `paths.scanned:[\"sample.py\"]`. Its captured stderr reports a successful scan, without RPC, network or filesystem errors (`raw/linux-sandbox-sample*`). The observed whole-second duration is **provisional, n=1**, not a benchmark. This diagnostic does not reproduce a Linux sandbox hang and does not waive the GitHub deadline failure.

A separate main-branch scratch clone with its own bootstrap-installed 1.100.0 tools also passes the same two scenarios. The sequential Linux comparison is **provisional, n=1 per arm**, without an interval or measured-improvement claim: baseline/upgraded DiffDelta 52.35/31.71 seconds, DiffGate 95.51/59.36 seconds and package 147.856/91.077 seconds. Raw baseline events, clone identity and all paired differences are `raw/linux-cli-baseline-scenarios.jsonl`, `raw/linux-baseline-identity.txt` and `raw/linux-scenario-comparison.json`. This host-specific single pair does not establish a performance change or explain the Ubuntu GitHub deadline failure.
