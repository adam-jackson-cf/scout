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
bash scripts/gate.sh full
bash scripts/gate.sh commit
bash scripts/gate.sh fast
bash scripts/gate.sh path internal/domain/types.go
```

`.omp/gate.json` owns the four argv arrays. CI's required `gate` job executes the
exact full argv, without modifying sources. Failures propagate to the caller.
CI reads the Go pin from `toolchain/versions.env` before setup-go and runs
`bash scripts/ci-sandbox.sh --strict` before bootstrap. The shared script installs
Linux bubblewrap and handles the Ubuntu user-namespace restriction; an unusable
sandbox blocks the gate immediately. Guardian provisioning calls the same script
without `--strict`, preserving its warning-only behavior when the probe fails.
Full/fast subchecks discover their maintained scope without forwarding the full
repository inventory as process arguments. Commit snapshots receive their changed
paths over NUL-separated stdin; `pathCheck` retains its explicit changed-path argv.
The schema-1 config allows 10 seconds for path checks: more than ten times the
recorded 929.494 ms maximum baseline. Sensitive folder patterns end in `/`;
file patterns are basename prefixes, so `bootstrap-tools.sh` protects
`scripts/bootstrap-tools.sh` and `scout-design.md` protects `docs/scout-design.md`.
Sensitive changes require owner merge even with `agentMerge` set to `allowed`;
that setting never grants permission to perform Git writes.
The pre-commit hook transfers the staged index in bulk into an independent shared
clone, so partially staged files are checked as committed, not as edited. Its
cheap tier checks affected maintained files only, never `.tools`, `.git`, or
`third_party`. Secret scanning uses the original staged index, including Git's
alternate index. Deleted files do not become formatter inputs.
The pre-push hook forwards Git's ref updates and push remote to the fast runner;
each pushed commit not on the configured base (normally `origin/main`) is checked,
including non-HEAD refs. Every snapshot has its own Git directory, index, refs and
config, and checker/test subprocesses receive no inherited `GIT_*` variables.
The installed runner checks each checkout without changing the pusher's refs,
config, staged/working-tree state or stash. Deletions need no source checks.
Commits predating `.omp/gate.json` are explicitly reported and skipped; every
configured commit uses the installed test checker against its own source and
configuration, comparing growth against the same default-branch base.
Failed tests print concise failure blocks and a `go test JSON:` path to the full
temporary JSON log.
If the configured base is absent, the runner queries the push remote's advertised
default HEAD and uses it only when that commit already exists locally. It never
fetches or repairs refs/config automatically: fetch the advertised default branch
and retry when the diagnostic requests it.

Fast tests have provisional owner-approved limits of 159 seconds wall time and
46 seconds per case, based on five recorded baseline runs. Speed failures require
removing needless work or injecting clocks/runners, not moving tests to another tier.
Growth mappings and owning tiers are maintained in `.omp/test-map.json` and
`.omp/test-tiers.json`. Missing/invalid/stale measurements fail closed.

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
bash scripts/gate.sh full
```

The full gate checks that the aggregate exactly matches its fragments, then runs
the byte-preserved validator/catalog in `scripts/guardrails/vendor/`. Neither
check rewrites the ledger. Until live protection evidence is recorded, the
validator intentionally fails on `GIT.required-checks`.
