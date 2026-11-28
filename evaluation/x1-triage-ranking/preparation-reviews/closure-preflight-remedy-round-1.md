# X1 preparation runtime closure (b): remedy for the refused preflight — round 1

**Scope.** This round chooses and signs off the remedy for the refused journalled closure preflight `412402435e9f4ac9bb810a5520d102d5`, and records how that refusal is recovered at seal. It does not reopen the closure implementation, which stays signed at the digests in `binding-closure-confirm-round-2.md`. Under R-064, Main adopts the recommended option without asking the owner.

**Method.**
- Read-only. The council ran no command, test, hash, journal, attestation, adoption or seal. Its only model calls were the two peer analyses (`innovation-challenger`, `experiment-peer`).
- All digests are Main's (`shasum -a 256`), as recorded in the journals, the refused receipt and `closure-p4-attempt-2-2026-11-10.json`.
- Paths are relative to `evaluation/x1-triage-ranking/` unless absolute. `SCRATCH` = `/Users/adamjackson/Projects/work/x1-preparation-scratch`; `ROOT` = `/Users/adamjackson/Projects/work/scout-x1prep`.
- Main's read-only diagnostics (`go list -m -e -json all` against `SCRATCH/closure-modules`; `image_filetype` over the real selected roots) are measurement checks. Rule 3 allows them, and this record records them. They look at no predictor–outcome association.

## Evidence reviewed

| Item | SHA-256 / location | Use |
|---|---|---|
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` | `go_metadata` `:410–461`; `dry_check` `:752–806`; `main` `:809–856`; env `:639–645`; SCRATCH cache check `:611–612` |
| `execution/preparation.py` | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` | `verify_tree_pins` `:423–434`; `review_receipt` `:437–461`; `collect_journals` `:596–639`; `journal` subcommand `:827–839` |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | SCRATCH/OUT paths |
| `preparation/pins.json` | `6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc` | pinned by `4124…` |
| `preparation-reviews/binding-closure-confirm-round-2.md` | `c2bf06cff7830cc085b4a5f0496b4835394047332019ddfa3a76d0e7278ff944` | current closure sign-off; P3/P4 still standing |
| `preparation-reviews/closure-impl-round-5.md` | digest not supplied | P3 `:132–135`, P4 `:137–140` |
| `preparation-reviews/closure-p4-attempt-2-2026-11-10.json` | `539abde46dd7557cbc66fd7f56b71804309f934ee906c8415c21edeab68c30f3` | argv, failing command, journal digests |
| `preparation-reviews/closure-p4-attempt-2026-11-10.json` | `2e26611a50303a3069fa1fc416dfc98c11c3109b939ae160734ddef8dcc55acf` (43778 B) | earlier unjournalled attempt; not a signed step |
| `preparation-run/journal/412402435e9f4ac9bb810a5520d102d5.start.json` | `98c5d6e75958d5b3bcf02dc632f8f2545082b911e2a63760c4a6d586d5e57f19` (19029 B) | refused operation, start |
| `preparation-run/journal/412402435e9f4ac9bb810a5520d102d5.json` | `6228531e352a3463ee7154d82dbb13cf0bf208730bc0842752e8e658642fe597` (19910 B) | refused operation, end: `status: refused`, `exit_code: 1` |
| `SCRATCH/closure-preflight/412402435e9f4ac9bb810a5520d102d5.json` | `839b973a573a52b6d833cd0bedd274410d3dc977e18355e9cd44b44ea3172390` (2366 B) | refused receipt; pins command records `000001`–`000005` |
| `SCRATCH/closure/412402435e9f4ac9bb810a5520d102d5/commands/000005.json` | `aa678a7628d1af8acb5a4c6196069e49caee7f10f88c110f9e61b6a0348a26ba` (2166 B) | failing `go list -m -json all`; stderr `go: module lookup disabled by GOPROXY=off` (repeated) |
| `SCRATCH/closure/412402435e9f4ac9bb810a5520d102d5/commands/000004.json` | `d75c74c4edbec038ef285471eef0df83e0faf1b95b2fef6c0da0fe519799d53e` | `go list -pgo=off -deps -json …/model_free` passed offline |
| `preparation-run/journal/3f969fafa01244478350d2d37e52c657.json` | `da008e6d25978a9b40cca804e13cea241a8a6407af5d2efa8d714d0071c47344` | setup copy `/bin/cp -R -p ~/go/pkg/mod SCRATCH/closure-modules`; `output_pins: []` |
| `ROOT/go.mod`, `ROOT/go.sum` | not hashed here | `go 1.26.0`, `toolchain go1.26.10`, no `replace`; `go.sum` has no `github.com/google/go-cmp` line |
| Go 1.26.10 source, `~/.local/share/mise/installs/go/1.26.10/src/cmd/go/internal/` | not hashed | `modload/build.go` `completeFromModCache` (read at `:319–403`); `modcmd/verify.go` `runVerify`/`verifyMod` (read at `:55–150`) |

## Findings

**F1. Cause.** In `completeFromModCache`, every listed module with a version gets a `Query` that needs the cached `.info` file. An error there becomes `m.Error`, and without `-e` that makes `go list -m` fail. Reading the module's `go.mod` for `GoVersion` is gated by `checksumOk("/go.mod")`, and its errors are ignored.

So `go list -m -json all` needs `.info` for all 28 graph modules. The 15 unselected modules lack it, so the command fails under `GOPROXY=off`. Package loading needs no `.info` for unselected modules, which is why `go list -deps` passed (`000004.json`).

The developer cache lacks the same files (Main), so the SCRATCH copy faithfully reproduces a pre-existing gap. It is not a copy fault.

**F2. Why the dry check did not catch it.**
- `dry_check()` runs no Go command (`attest_closure.py:752–806`).
- The Go tests fake any `-m` argv (`tests/test_closure_review.py:332–334`, `tests/test_closure_round2.py:87–88`).
- Closure rounds 1–5 listed "offline `go list -m all` / `go mod verify`" as unverified on real bytes (`closure-impl-round-1.md:109`, `-round-2.md:108`, `-round-3.md:166`, `-round-4.md:222`).

The risk was known and is now realised. It is not a new implementation defect.

**F3. Remedy A in place cannot be sealed.**
- `collect_journals` runs `verify_tree_pins(value)` on every terminal record (`preparation.py:618`) before the status/recovery branch (`:621–630`).
- `4124…` pins `execution/attest_closure.py` at `113d2d72…` in both `producer_pins` and `source_pins`.
- Any byte change to that path therefore makes the seal refuse with `pin mismatch`, and no recovery receipt can clear it.
- `preparation.py` cannot be relaxed: it is pinned by `4124…` and by CC2 journal `0378e156…`.

**F4. `go mod verify` does not need `.info`.**
- `verifyMod` skips modules with neither zip nor directory, which covers go-cmp.
- It refuses `missing ziphash` when a zip or directory exists without `.ziphash`, and it hash-checks every downloaded module in the build list, unselected ones included.
- No real run has reached it yet (`attest_closure.py:455–457`). Its result does not depend on adding `.info` files, so it can be pre-checked on the current cache (G0 below).

**F5. Recovery path.** The Core recovery path covers `4124…`, provided every pin in its end record stays valid until seal (`preparation.py:618–630`).

## Decision

**Adopt remedy B′: an add-only, metadata-only cache population in a new SCRATCH copy, with the signed code unchanged.**

1. **Code and runbook.** No signed file changes. The closure keeps running offline with the signed environment (`GOPROXY=off`, `-mod=readonly`).
2. **New copy.** Keep `SCRATCH/closure-modules` byte-stable, as P3 requires and as the environment in which `4124…` refused. Make a new journalled copy, `SCRATCH/closure-modules-r2`.
3. **Network step.** Populate the missing `.info` files with one journalled, network-enabled run of the very command that failed: `go list -m -json all`. It runs against `closure-modules-r2`, from a staging copy of `go.mod`/`go.sum` outside `ROOT`, with a fixed HTTPS proxy and no `direct` fallback.
   - By F1, this writes only the 15 missing `.info` files.
   - Go may also create lock files and missing download directories [INFERENCE].
   - It does not download or extract zips, and it fetches no go-cmp `go.mod` (`checksumOk` is false because `go.sum` has no go-cmp line).
4. **Preflight.** Rerun the preflight with the signed argv, changing only the module-cache prefix.

**Reasons.**
- **Closure claim.**
  - The selected module bytes are those already classified: copies of the same files.
  - The pins, `go.sum` binding and `go mod verify` are unchanged, and verify still covers every downloaded graph module.
  - `module_graph` keeps its meaning and becomes complete. Its 15 new rows carry proxy-sourced `Time` values, which `go.sum` does not authenticate. The same is already true of the 12 selected rows.
- **Irreversibility.** This adds two create-once setup journals and one preflight journal, and it invalidates no existing pin. A failed setup journal costs one recovery review. The canonical `OUT/closure.json` still runs only after a passing preflight.
- **Clock.** No code, tests or closure re-sign-off are needed; the remedy takes hours, not rounds. The deadline is 2026-12-24T03:22:58Z.
- **Blast radius.**
  - Writes go only to new SCRATCH directories.
  - The network is reached once, only `proxy.golang.org` and `sum.golang.org`, under a clean `env -i`.
  - `ROOT/go.mod` and `go.sum` are never in the working directory of a networked Go command; staging copies are compared byte-for-byte with them afterwards.

## Alternatives rejected

| Option | Why rejected |
|---|---|
| **A. Edit `go_metadata` in place** (select paths or `-e`) | Fatal at seal (F3). The `-e` variant also turns refusal into tolerated `Error` rows: today's code neither rejects module `Error` nor requires every selected module to be covered. |
| **A′. New wrapper or file that monkeypatches `go_metadata`** | Possible, but needs new code, tests, a closure round and a binding-confirm round. `own_sources()` would not pin the wrapper, `module_graph` would change meaning, and monkeypatching signed code is itself a review concern. Kept only as the fallback named under Risks. |
| **B in place** (add files to `closure-modules`) | Breaks P3's byte-stability wording and makes journal `3f969faf…` misdescribe the directory. It also destroys the environment in which `4124…` refused. Its only gain, an unchanged argv, is small. |
| **B with `go mod download` of the 15 modules** (peer variant) | Downloads and extracts unselected trees the closure does not need. That widens what `go mod verify` must hash and adds sumdb traffic for every module. Larger footprint for no gain over the `.info`-only fetch. |
| **Hand-written `.info` files** | They would fabricate timestamps. |
| **`GOPROXY=file://` or a freshly downloaded cache** | `file://` changes the hard-coded environment (`:642`) and still needs `.info` from somewhere. A fresh cache replaces the provenance of the selected module bytes. |

## Exact next steps (ordered; stop at the first failure and open a new council round)

Fixed paths:
- `PY=/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python`
- `EXEC=$ROOT/evaluation/x1-triage-ranking/execution`
- `OUT=$ROOT/evaluation/x1-triage-ranking/preparation-run`
- `GO=/Users/adamjackson/.local/share/mise/installs/go/1.26.10/bin/go`

Nothing below writes under `OUT` except the two journal pairs and the recovery receipt.

**Step 1. G0 (unjournalled, read-only).** Pre-check `go mod verify` on the current cache, before any journal.
- Environment: exactly the `attest()` environment keys (`attest_closure.py:639–645`), with `GOMODCACHE=SCRATCH/closure-modules`. Put `HOME`, `GOCACHE` and `GOTMPDIR` in a fresh `SCRATCH/closure-remedy-checks/g0/`.
- Run `"$GO" mod verify` with cwd `ROOT`.
- Pass: exit 0 and stdout exactly `all modules verified`. A failure means the remedy is incomplete (F4): stop.
- Also record, without gating, the output of `find SCRATCH/closure-modules -newer $OUT/journal/3f969fafa01244478350d2d37e52c657.json`. If any listed path lies inside a selected module directory or its `cache/download/<module>/@v/` entry, stop.

**Step 2. J1 (journalled, action b, no network).** Copy the cache.
```sh
"$PY" -I -B "$EXEC/preparation.py" journal --config "$EXEC/run-config.json" --clock "$OUT/clock.json" --action b \
  /bin/sh -eu -c '[ ! -e /Users/adamjackson/Projects/work/x1-preparation-scratch/closure-modules-r2 ]; /bin/cp -R -p /Users/adamjackson/Projects/work/x1-preparation-scratch/closure-modules /Users/adamjackson/Projects/work/x1-preparation-scratch/closure-modules-r2'
```

**Step 3. J2 (journalled, action b, the one network use).** Fetch `.info` metadata only.
```sh
"$PY" -I -B "$EXEC/preparation.py" journal --config "$EXEC/run-config.json" --clock "$OUT/clock.json" --action b \
  /bin/sh -eu -c '
S=/Users/adamjackson/Projects/work/x1-preparation-scratch
R=/Users/adamjackson/Projects/work/scout-x1prep
F="$S/closure-modules-r2-fetch"
/bin/mkdir "$F" "$F/module" "$F/home" "$F/gocache" "$F/tmp"
/bin/cp -p "$R/go.mod" "$R/go.sum" "$F/module/"
/usr/bin/env -i PATH=/usr/bin:/bin LANG=C LC_ALL=C HOME="$F/home" GOCACHE="$F/gocache" GOTMPDIR="$F/tmp" GOMODCACHE="$S/closure-modules-r2" GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org GOTOOLCHAIN=local GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 GOARM64=v8.0 GOEXPERIMENT= /Users/adamjackson/.local/share/mise/installs/go/1.26.10/bin/go -C "$F/module" list -m -json all >"$F/module-list.json" 2>"$F/module-list.stderr"
/usr/bin/cmp "$R/go.mod" "$F/module/go.mod"
/usr/bin/cmp "$R/go.sum" "$F/module/go.sum"
'
```
Pass conditions:
- the journal is `passed` with exit 0;
- `module-list.stderr` is empty;
- `module-list.json` holds 28 objects, none with an `Error` key.

**Step 4. G4 (unjournalled, read-only on `-r2`).** Rehearse the signed `go_metadata` exactly.
- In a Python harness run with `"$PY" -I -B`, import `execution/attest_closure.py` without editing it.
- Build `allowed` and `reviewed_go` as `attest()` does at `:670–672`: `pins['source']['files']` mapped onto `ROOT`, plus `reviewed_go_sources(ROOT, OUT)`.
- Build `env` with exactly the keys and values of `:639–645`, except:
  - `GOMODCACHE=SCRATCH/closure-modules-r2`;
  - `HOME`, `GOCACHE` and `GOTMPDIR` under a fresh `SCRATCH/closure-remedy-checks/g4/` (created first).
- Call `go_metadata(runtime['go']['binary'], ROOT, Commands(env, SCRATCH/closure-remedy-checks/g4/commands), allowed, reviewed_go)`. `runtime` is `preparation/runtime-attestation-rev6.json`.
- Pass conditions:
  - no exception;
  - `version` equals `runtime['go']['version']`;
  - `module_verification == 'all modules verified'`;
  - `module_graph` has 28 rows, none with `Error`;
  - `modules` holds the main module plus exactly the 12 selected non-main paths, each with `verified_sum` and `verified_gomod_sum`;
  - the returned roots are GOROOT plus the 12 module directories under `SCRATCH/closure-modules-r2/`.

**Step 5. G5 (read-only).** Rerun the signed S11 dry check (`attest_closure.py --dry-check`) with the same 32 roots. The only change: each `--external-root` with prefix `SCRATCH/closure-modules/` gets prefix `SCRATCH/closure-modules-r2/`. Pass: exit 0 with zero classification failures.

**Step 6. G1 (read-only, after G4 and G5).** Prove the change was add-only.
- Every entry under `SCRATCH/closure-modules` exists under `SCRATCH/closure-modules-r2` with the same type and mode; regular files also need the same size and SHA-256, symlinks the same target.
- Every entry only in `-r2` must be one of:
  - a `cache/download/<escaped path>/@v/<escaped version>.info` file for one of the 15 graph modules that lacked `.info` in `closure-modules`;
  - a `*.lock` file under `cache/download/`;
  - a directory needed to hold those files (for go-cmp: `cache/download/github.com/google/go-cmp/` and `@v/`).
- The 15-module set derived from `module-list.json` must equal Main's diagnostic list:
  - `github.com/google/go-cmp v0.6.0`;
  - `github.com/google/pprof v0.0.0-20250317173921-a4b03ec1a45e`;
  - `golang.org/x/mod v0.27.0`, `golang.org/x/sync v0.16.0`, `golang.org/x/tools v0.36.0`;
  - `gopkg.in/check.v1`;
  - modernc.org `cc/v4`, `ccgo/v4`, `fileutil`, `gc/v2`, `goabi0`, `opt`, `sortutil`, `strutil`, `token`.
- Any other addition or difference: stop.

**Step 7. Gate record.** Main writes `preparation-reviews/closure-preflight-remedy-gates-2026-11-10.json`, never under `OUT`. It holds:
- the argv, exit code, stdout/stderr SHA-256 and timestamps for G0, J1, J2, G4, G5 and G1;
- both journal IDs and their digests;
- the SHA-256 of `module-list.json`;
- the G1 addition list.

No council round is needed before the preflight if every gate passes exactly as stated. A gate that fails or deviates in any way needs a new round.

**Step 8. Preflight.** Run one journalled `--preflight` with the argv of `4124…` and exactly two kinds of substitution:
- the `--module-cache` value becomes `SCRATCH/closure-modules-r2`;
- each `--external-root` with prefix `SCRATCH/closure-modules/` gets prefix `SCRATCH/closure-modules-r2/`.

No other argument, order or value changes. If it is refused, stop and open a new round; that refusal needs its own recovery record.

**Step 9. Canonical run.** Only after a `passed` preflight end record, run the canonical closure once with the same argv, `--preflight` removed and `--output $OUT/closure.json` added (P4, S16). Then follow the binding runbook (RB-1).

**Step 10. RESULTS.md, Deviations.** Add entries for:
- (i) refused preflight `4124…`: cause F1, evidence digests, recovery under this record;
- (ii) the earlier unjournalled over-broad attempt (`2e26611a…`), not a signed step;
- (iii) setup journals J1 and J2: one network use for `.info` metadata, the new cache `closure-modules-r2`, and the argv substitution;
- (iv) reproduce commands as above.

**Step 11. P3, restated.** Keep these byte-stable until seal:
- `SCRATCH/closure-modules` (unchanged; the environment of `4124…`);
- `SCRATCH/closure-modules-r2`;
- `SCRATCH/closure-modules-r2-fetch/`;
- `SCRATCH/closure-remedy-checks/`;
- `SCRATCH/closure/412402435e9f4ac9bb810a5520d102d5/`;
- `SCRATCH/closure-preflight/412402435e9f4ac9bb810a5520d102d5.json`;
- all later preflight and retained copies.

From Step 8 on, run no Go command against `-r2` except through `attest_closure.py` and the signed binding tools. The rest of P3 (OS, kegs, Go 1.26.10, casks, host clock) and P4 are unchanged.

## Recovery handling for `412402435e9f4ac9bb810a5520d102d5`

- The existing Core path covers this refusal: `collect_journals`, `preparation.py:621–630`.
  - At seal, it requires `OUT/reviews/operation-recovery-412402435e9f4ac9bb810a5520d102d5.json`.
  - It checks the receipt with `review_receipt(…, 'operation-recovery')`.
  - It lists the start record, end record, receipt and authority record under `refusals_or_deviations`, not under `operations`.
- **Precondition:** every pin in the end record must still validate (`:618`). That means these must never change before seal:
  - `attest_closure.py`, `closure/native.py`, `preparation.py`, `verify_inputs.py`, `run-config.json`;
  - `pins.json`, `reviews/closure-model-free.json`, `binding-closure-confirm-round-2.md`;
  - the clock, the prior journals and the refused receipt.

  B′ preserves all of these.
- **Authority.** This record is the authority for that receipt: its single operation-recovery scope block is at the end. Main writes it as UTF-8 with LF line endings and pins it.
- **Receipt.** Then Main writes the receipt with exactly these keys and values:
```json
{
  "scope": "operation-recovery",
  "record": {"path": "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-preflight-remedy-round-1.md", "sha256": "<sha256 of this file as written>", "bytes": "<size as integer>"},
  "reviewed": [
    {"path": "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/journal/412402435e9f4ac9bb810a5520d102d5.start.json", "sha256": "98c5d6e75958d5b3bcf02dc632f8f2545082b911e2a63760c4a6d586d5e57f19", "bytes": 19029},
    {"path": "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/journal/412402435e9f4ac9bb810a5520d102d5.json", "sha256": "6228531e352a3463ee7154d82dbb13cf0bf208730bc0842752e8e658642fe597", "bytes": 19910},
    {"path": "/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-preflight/412402435e9f4ac9bb810a5520d102d5.json", "sha256": "839b973a573a52b6d833cd0bedd274410d3dc977e18355e9cd44b44ea3172390", "bytes": 2366},
    {"path": "/Users/adamjackson/Projects/work/x1-preparation-scratch/closure/412402435e9f4ac9bb810a5520d102d5/commands/000005.json", "sha256": "aa678a7628d1af8acb5a4c6196069e49caee7f10f88c110f9e61b6a0348a26ba", "bytes": 2166}
  ]
}
```
  - `bytes` is an integer; fill in the measured value.
  - The fourth row binds the failing command's evidence; it lives in SCRATCH, which P3 already keeps frozen.
- **Pre-check.** Before any later journal, Main runs `review_receipt` on the receipt (read-only). It may be written right after this record. It must exist before seal; Core's rule (`closure-impl-round-2-response.md:16`) puts it before the canonical closure.
- **Disposition.** `4124…` is a retained refusal, not action-(b) evidence. `closure.json` still needs its own passed producer (`:729–731`). Nothing about `4124…` is deleted, rewritten or relabelled.

## Risks

- **J2 network failure or unexpected writes.** The proxy may fail to serve a pseudo-version, or Go may write more than `.info` [INFERENCE]. Either way J2 fails or G1 stops the run. A failed J2 needs a recovery record. Fallback: `go mod download -json` of the named modules in a new round, then A′ only if Go metadata still cannot be satisfied.
- **`go mod verify`.** It may refuse on an unselected module (missing `.ziphash`, or a modified zip or directory). G0 detects this before any journal.
- **Later closure stages never run on real bytes** [INFERENCE]: the native load walk over `.tools`, GOROOT and the kegs; `dyld_info` over the cache; retention. A refusal there costs another preflight recovery, not `closure.json`, which is the purpose of P4.
- **Unauthenticated metadata.** `module_graph` `Time` values come from the proxy and are not bound by `go.sum`. They are informational only and are recorded as a deviation.
- **Reversibility.** J1, J2 and each preflight are create-once journals, but they touch only new SCRATCH directories and invalidate no existing pin. The irreversible step remains the canonical `OUT/closure.json`, gated on a passed preflight.
- **Clock.** Deadline 2026-12-24T03:22:58Z; B′ needs no further council round if every gate passes.

## How the peers shaped this record

- **Both peers** found F3 independently; it matches my own reading of `preparation.py:618–630` and the pins of `4124…`. That finding removed A, which the question framed as the main candidate.
- **innovation-challenger**
  - Moved me from B in place to a separate cache copy.
  - Supplied the staging `go list -m -json all` approach and the Go 1.26.10 source reading. I checked that reading against `completeFromModCache` and `verifyMod` and adopted the `.info`-only fetch.
  - Warned that `-e` alone is unsafe.
- **experiment-peer**
  - Supplied the competing hypotheses (H1–H6), including H1, that the dry check runs no Go.
  - Flagged `go.sum` having no go-cmp line and the risk of writing `ROOT/go.sum`.
  - Supplied the end-record digest's provenance (Main's attempt-2 record), the minimal reviewed set, the hazard of writing new files under OUT, and G0.
  - Proposed signing the recovery in this same round, which I adopted.
  - Its `go mod download` variant was rejected in favour of the smaller `.info`-only fetch.
- **Mine**
  - The exact gate sequence G0, G4, G5, G1, with G1 placed after the rehearsals.
  - The fixed argv substitution rule.
  - Inclusion of `000005.json` in the receipt.

## New issues

None. The refusal is a known, recorded environment risk (F2), remedied without changing any signed byte.

**Zero issues. Signed off.**

Verdict: ZERO ISSUES — signed off

The innovation council signs off **remedy B′** exactly as specified under "Exact next steps":
- G0 → J1 → J2 → G4 → G5 → G1 → gate record → preflight with the stated substitution → canonical closure;
- P3 restated as above;
- any gate failure or deviation needs a new zero-issue round before the run continues.

No execution, test or configuration file changes. The closure implementation remains signed at the digests in `binding-closure-confirm-round-2.md` (`c2bf06cff7830cc085b4a5f0496b4835394047332019ddfa3a76d0e7278ff944`), including:
- `execution/attest_closure.py` `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892`
- `execution/preparation.py` `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`
- `preparation/pins.json` `6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc`

The evidence records reviewed for the operation recovery:
- `closure-p4-attempt-2-2026-11-10.json` `539abde46dd7557cbc66fd7f56b71804309f934ee906c8415c21edeab68c30f3`
- `closure-p4-attempt-2026-11-10.json` `2e26611a50303a3069fa1fc416dfc98c11c3109b939ae160734ddef8dcc55acf`

Signed-off scope: operation-recovery

Refused closure preflight `412402435e9f4ac9bb810a5520d102d5` (action b): reviewed and recovered as a retained refusal with cause F1, under remedy B′. The reviewed evidence:
- start record `preparation-run/journal/412402435e9f4ac9bb810a5520d102d5.start.json` `98c5d6e75958d5b3bcf02dc632f8f2545082b911e2a63760c4a6d586d5e57f19`
- end record `preparation-run/journal/412402435e9f4ac9bb810a5520d102d5.json` `6228531e352a3463ee7154d82dbb13cf0bf208730bc0842752e8e658642fe597`
- refused receipt `SCRATCH/closure-preflight/412402435e9f4ac9bb810a5520d102d5.json` `839b973a573a52b6d833cd0bedd274410d3dc977e18355e9cd44b44ea3172390`
- failing command record `SCRATCH/closure/412402435e9f4ac9bb810a5520d102d5/commands/000005.json` `aa678a7628d1af8acb5a4c6196069e49caee7f10f88c110f9e61b6a0348a26ba`
