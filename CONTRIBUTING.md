# Contributing

## Install pinned tools and native hooks

Install a Go bootstrap toolchain first (for example, `mise use go@1.26.10`).
The installer fails clearly if Go is unavailable; Go's toolchain downloader
then selects the version in `toolchain/versions.env`.

```sh
git submodule update --init
scripts/bootstrap-tools.sh
source .tools/activate
git config core.hooksPath .githooks
```

The last command installs this checkout's tracked native hooks. It is an explicit
local setup step, not a side effect of running checks. No hook bypass is supported.
Tool pins live only in `toolchain/versions.env`; when changing an allowed developer
tool pin, update its package metadata and run `scripts/bootstrap-tools.sh --relock`.
Scanner and Clawpatch upgrades require their compatibility review.

## One gate locally and in CI

```sh
bash scripts/gate.sh ci
bash scripts/gate.sh commit
bash scripts/gate.sh path internal/domain/types.go
```

`.omp/gate.json` schema version 2 owns `checks.commit`, `checks.ci` and
`pathCheck` argv arrays. CI's required `gate` job executes the exact `checks.ci`
argv on pushes and pull requests, without modifying sources. Failures propagate.
CI reads the Go pin from `toolchain/versions.env` before setup-go and runs
`bash scripts/ci-sandbox.sh --strict` before bootstrap. The shared script installs
Linux bubblewrap and handles the Ubuntu user-namespace restriction; an unusable
sandbox blocks the gate immediately. Guardian provisioning calls the same script
without `--strict`, preserving its warning-only behavior when the probe fails.
CI subchecks discover their maintained scope without forwarding the full
repository inventory as process arguments. Commit snapshots receive their changed
paths over NUL-separated stdin; `pathCheck` retains its explicit changed-path argv.
The config allows 10 seconds for path checks: more than ten times the
recorded 929.494 ms maximum baseline. Sensitive folder patterns end in `/`;
file patterns are basename prefixes, so `bootstrap-tools.sh` protects
`scripts/bootstrap-tools.sh` and `scout-design.md` protects `docs/scout-design.md`.
Sensitive changes require owner merge even with `agentMerge` set to `allowed`;
that setting never grants permission to perform Git writes.
The pre-commit hook transfers the staged index in bulk into an independent shared
clone, so partially staged files are checked as committed, not as edited. Its
commit tier runs static checks, every T1 package, and T2 seam packages affected by
staged changes, using `.omp/test-tiers.json`. Secret scanning uses the original
staged index, including Git's alternate index. Deleted files do not become
formatter inputs. Every snapshot has its own Git directory, index, refs and
config, and checker/test subprocesses receive no inherited `GIT_*` variables.
There is no pre-push hook or push-range runner: CI owns the complete T1 and T2 suite.
Failed tests print concise failure blocks and a `go test JSON:` path to the full
temporary JSON log. The ledger's `hook-check.py pre-commit` installation check
proves wiring, failure propagation and snapshot isolation in an isolated checkout.

T1 tests retain provisional owner-approved limits of 159 seconds wall time and
46 seconds per case, based on five recorded baseline runs. Speed failures require
removing needless work or injecting clocks/runners, not moving tests to another tier.
Growth mappings and owning tiers are maintained in `.omp/test-map.json` and
`.omp/test-tiers.json`. Missing/invalid/stale measurements fail closed.
Five complete commit-tier measurements with one staged association-package change
are recorded in `.omp/guardrails.json` under `budget.commitTierMeasurement`.
The provisional 159-second ceiling is retained; no budget increase was applied.

Python's maintained roots/packages, supported versions, exclusion rationales,
canonical test groups, first-party names and missing-stub override evidence live
in `scripts/guardrails/python/settings.json`. Checkers consume that target policy;
frozen `evaluation/**` remains excluded. Skill frontmatter follows the target-owned
metadata schema in `.omp/guardrails.json`. `.agents/skills/` contains project-scoped
skills only; no shared-skill metadata field is required here.

Prettier 3.6.2 checks maintained tracked Markdown/YAML. All `evaluation/**`,
`docs/design-reviews/**`, historical `docs/archive/**`, scanner rule assets under
`internal/analyzers/rules/**`, governed design and ledger documents, third-party
code, instruction/migration files and `.codegraph` directories remain excluded.
Markdown fixtures under `internal/**/testdata` are also excluded. Checks never run frozen studies.
The dependency audit is strict: unavailable tooling/network and vulnerability
findings both fail. Runtime seam tests record revision, report paths and SHA-256
artifact digests under ignored `.tools/guardrails-seam-*`.

Host protection is separate from offline checks. After the first green CI run,
a maintainer must require `gate` on the default branch and verify host readback;
`.omp/host-settings.json` records pending versus observed state. Fixture successes
are never evidence that GitHub protection is installed.

## Complete live host protection evidence

After protection is applied, record successful live GitHub protection and
repository-settings queries as JSON under `.omp/evidence/`, including the command,
exit code, stdout/stderr, query time and checked revision. The readback must prove
that the default branch requires `gate` and the supported auto/rebase merge
settings are enabled; offline fixtures are not host evidence.

Update only the `GIT.required-checks` row in `.omp/ledger.d/c.json`: point
`host_query` at the successful record with assertions against its actual output,
then change `status` to `implemented`. Rebuild the aggregate explicitly:

```sh
.tools/venv/bin/python scripts/guardrails/merge-ledger.py
bash scripts/gate.sh ci
```

The CI gate checks that the aggregate exactly matches its fragments, then runs
the byte-preserved validator/catalog in `scripts/guardrails/vendor/`. Neither
check rewrites the ledger. The validator checks recorded live host-query evidence
for `GIT.required-checks`; offline fixtures never replace that evidence.
