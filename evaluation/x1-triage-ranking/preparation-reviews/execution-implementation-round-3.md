# X1 execution-implementation council — round 3 (remedy for the binding-v2 rehearsal drift)

**Study:** `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/` (STUDY). OUT = `STUDY/preparation-run`. SCRATCH = `/Users/adamjackson/Projects/work/x1-preparation-scratch`.
**Date:** 2026-11-10.
**Action:** preparation action (c), binding v2.
**Authority:** R-064 is in force, so this council recommends one remedy and does not pose a choice.

The first fresh scratch rehearsal, authorised by `execution-implementation-round-2.md` step 5, exited 1 with `ValueError: installed runtime file/link/mode set drift`. This round decides:
- how to restore the installed file;
- the `binding/**` repair under `preparation/model-free-binding-amendment.md:74`;
- the integrity and deviation consequences;
- the authorised sequence from here.

## Inputs

SHA-256 values marked "…" are the prefixes supplied to the council. The council did not recompute any digest.

| Input | SHA-256 |
|---|---|
| Round-2 record `preparation-reviews/execution-implementation-round-2.md` | `d547b10382b03920373e431e506f711511bab336169473e01860fc292f40338d` |
| `execution/binding/acceptance.py` (signed, round 2) | `6e56fb5f6f2f8f1041c2d06498b6f2d01c0e7480a2911def12902094e955de63` |
| `execution/binding/interlock.py` (signed, round 2) | `19464e8940d0458be3606ed0438ae778fd01b48b73f8c2fee801856b80b2e6a1` |
| `execution/tests/test_binding.py` | `be91c678512402244ff27847177046511caa129ac6d8834b079c2add35ccc155` |
| `OUT/reviews/implementation.json` (9 rows) | `8711ffaecc7662609f6afa75a9e50ba6ff1fcf8f1b9d41312a0639efed11a0d4` |
| `OUT/execution-source-manifest.json` | `25ef2b4e6387f5ed496acdc4d28c187fcf1a38f71fdb2d783bda5bbf78a3d147` |
| `OUT/binding/runtime.json` | `647381c553c78516fe19fe7d83e7401be1692e6a6922e93f4218606b52ebbc6d` |
| `OUT/closure.json` | `b5df46b1b92a3cc9ef2905589db120a008881bc6b5519f37b6d38e3705cc35b2` |
| Rehearsal run record `preparation-reviews/binding-v2-rehearsal-run-2026-11-10.json` | `58d8856b…` |
| Rehearsal session record `preparation-reviews/binding-v2-rehearsal-session-2026-11-10.json` | `084e61cd…` |
| Failed report `SCRATCH/binding-v2-rehearsal/report.json` | `4bb73d07…` |
| Diagnosis script `SCRATCH/binding-v2-rehearsal-diag/compare_enumeration.py` | `5b23f238…` |
| Its stdout | `e7b5ea36…` |
| Frozen pyc (closure, two enumeration records) | `71093d3edaad0475b353d8b8515d0888fe84fbe74f0fac3080a907dbc5b01d12` (12015 B) |
| Current (drifted) pyc | `9c815f6b…` (12018 B) |

**Method:** three independent read-only analyses: Main council (Opus), `innovation-challenger` (GPT Astra) and `experiment-peer` (Opus). Nobody launched, built, tested or changed a file.

## Findings

### F1. Cause: the positive-control stub ran the attested base interpreter without `-B`

The council considers this established; attribution to a PID is [INFERENCE].

- **The stub has no interpreter flags.**
  - `interlock.py:13–21` writes the shebang as `'#!'+str(python)` with no flags.
  - `acceptance.py:116` passes `runtime['python']['path']`, which is the base interpreter `scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11`.
  - The retained stub `SCRATCH/binding-v2-rehearsal/report.json.work/logs/positive/parent/stubs/claude` has exactly that bare shebang (read by Main council).
- **The startup hook imports the module.** `…/cpython-3.11.11-macos-aarch64-none/lib/python3.11/site-packages/distutils-precedence.pth` runs `__import__('_distutils_hack').add_shim()` on every start that processes site. `-I` does not disable that.
- **The environment does not stop bytecode writes.**
  - `private_env` (`acceptance.py:26–38`) sets neither `PYTHONDONTWRITEBYTECODE` nor a cache prefix.
  - The controller's own `-I -B` is not inherited by its children.
- **The drift happened in the positive-parent run.**
  - The work tree has `resolve/*` (6) and `logs/positive/parent`, but no `logs/positive/head`.
  - `run_revision` verifies the closure before it creates the run directory (`acceptance.py:113–115`). So the check passed before positive-parent and refused before positive-head (`execute_fixture` order, `acceptance.py:222–240`).
  - The peer found 80 `claude -p …` attempts in `stubs/attempts.jsonl`. Their argv matches the contextual launcher `internal/contextual/cli.go:24–38,81–84`, which passes `modelcli.Env()` (inherited environment).
  - The pyc mtime, 18:32:59Z, falls between the Clawpatch `createdAt` 18:32:58.873Z and the first scanner start at 18:33:00.949Z.
  - The codex positive control never ran.
- **Why the pyc was stale.**
  - The frozen pyc embeds the python-build-standalone build path (144 characters). The local path is 147 characters, which matches the 12015 → 12018 B growth.
  - Its header size field (6755) equals the source size, but its mtime field predates the installed source mtime.
  - So the pyc has been stale since installation. Any start of this interpreter that processes site without `-B` rewrites it. `closure-impl-round-1.md:68` (S10) had warned about exactly this.
- **Rejected alternatives.**
  - Sandboxed scanners: `internal/sandbox/sandbox.go:76–83` denies writes outside the private state directory, and the scanners started after the write.
  - Venv Python: `scout-x1/.tools/venv/pyvenv.cfg` has `include-system-site-packages = false`, so the venv never imports base site-packages.
  - A copy from another checkout: the build path rules it out.
  - Still open: an unrelated outside process inside the same ~2 s window. T4 below settles this operationally.

### F2. Concrete facts for the restore

These were read by Main council from `closure.json` and the filesystem.

- `retained_files` has exactly one row for the frozen bytes: `SCRATCH/closure-retained/sha256/71093d3edaad…`, 12015 B. The file is present (glob).
- `enumeration` has two records for the pyc path, under the `.tools` root and the overlapping cpython root. Both have mode 420 (0644), size 12015 and sha256 `71093d3e…`. `original_runtime_files` pins it too.
- The `__pycache__` directory is recorded with size 96. An in-place write does not change its entries; a temp file plus rename inside that directory would.
- `Inventory` (`attest_closure.py:85–129`) records path, type, `S_IMODE` mode, size and file sha256, plus link fields for symlinks. It records no mtime, ctime or inode, so an in-place byte restore that keeps mode 0644 restores equality.
- `retain()` sets 0444 on retained objects (`:147`), and `verify_installed_closure` refuses if a retained object is writable (`:486–488`). The restore must read the retained object and never move or chmod it. A copy that carries 0444 across would itself be mode drift.

### F3. Which processes can write into closure roots

| Process | Confinement | Exposure |
|---|---|---|
| Claude/codex stubs (base Python) | None; started by the contextual lane or Clawpatch | **The incident.** Fixed by C1. |
| Frontend (Go), Clawpatch (node), git | None (`acceptance.py:119`, "No outer sandbox") | Expected writes go to the fixture, run directory and private HOME/XDG only. No closure write is known. C3 makes any write fatal at the launch that made it. |
| Scanners (semgrep/pysemgrep, ruff, gitleaks, mypy, lizard, jscpd, prettier, govulncheck) | `sandbox-exec` with writes denied outside state (`sandbox.go:76–83`); fresh environment (`run.go:40–48`) | Python cache writes are denied by the kernel. |
| uv, npm/npx, pnpm | None if launched; no launch is identified on the binding route | `private_env` passes no `UV_CACHE_DIR`, so caches resolve to the private XDG and HOME. The bootstrap's `UV_CACHE_DIR=.tools/cache/uv` applies only to bootstrap, which P3 forbids. C3 covers this. |
| go list (Clawpatch Go mapper) | None | GOCACHE and GOMODCACHE default to the private HOME/XDG. Go does not write GOROOT. C3 covers this. |
| Other unsandboxed Python started through PATH (`python3` → `scout-x1/.tools/venv/bin/python`) | None | No launcher was found in Scout or Clawpatch. The venv holds about 4,400 `.py` files and one `.pyc`, so any such start would add files. C2 covers inheriting children; C3 detects the rest. |

## Issues

### X1EI-R3-01 — Major — the positive-control stubs run the attested base interpreter without bytecode suppression and write into a P3-frozen closure root

**Evidence:**
- F1.
- `interlock.py:13–21`; `acceptance.py:26–38,116`.
- Unit tests run stubs with `sys.executable` (`test_binding.py:70`) and mock `verify_installed_closure` (`:128`; `test_binding_boundaries.py:84`). That is why the suite could not catch the defect.

**Consequence:** every synthetic operation, both rehearsal and canonical, runs the claude positive control. Each would rewrite the stale pyc, and the next launch would refuse.

**Required fix:** C1, C2, T1 and T2 below.

### X1EI-R3-02 — Moderate — a closure write during the final launch of an operation is caught only at seal, after the create-once canonical receipt is `passed`

**Evidence:**
- The installed closure is verified only before each launch (`acceptance.py:113–115`).
- The last launch of each operation is reference/head for both synthetic and Django (`acceptance.py:235–240`).
- The seal re-verifies the closure (`preparation.py:702–703`).
- This incident shows that the binding's own launches can write into closure roots. Detection worked here only because the write was not in the final launch.

**Required fix:** C3 and T3 below.

### X1EI-R3-03 — Major — one closure-enumerated runtime file differs from `closure.json`, so P3 byte stability is breached

**Evidence:** F1, F2 and the diagnosis.

**Consequence:**
- Every binding launch and the seal refuse until the file holds `71093d3e…` again.
- The episode must be disclosed, because a later passing check cannot show that it happened.

**Required fix:**
- the restore R below;
- the `RESULTS.md` deviation D below;
- the operator rule O below.

## Recommendation

### R. Restore the file in place from the retained object (X1EI-R3-03)

This record freezes the restore script below, which runs once. Main writes it verbatim (UTF-8, LF) to `SCRATCH/binding-v2-restore/restore_pyc.py` and records its sha256.

It must not be placed under `STUDY/execution/`. The seal requires the execution manifest to equal every file there (`preparation.py:711–713`).

```python
"""X1EI-R3-03: return one P3-frozen pyc to its closure bytes. Run once, with -I -B."""
import hashlib, json, os, stat, sys
from pathlib import Path

TARGET = Path('/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/lib/python3.11/site-packages/_distutils_hack/__pycache__/__init__.cpython-311.pyc')
RETAINED = Path('/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-retained/sha256/71093d3edaad0475b353d8b8515d0888fe84fbe74f0fac3080a907dbc5b01d12')
STORE = Path('/Users/adamjackson/Projects/work/x1-preparation-scratch/binding-v2-restore/drifted')
FROZEN_SHA, FROZEN_BYTES, MODE = '71093d3edaad0475b353d8b8515d0888fe84fbe74f0fac3080a907dbc5b01d12', 12015, 0o644
DRIFT_PREFIX, DRIFT_BYTES = '9c815f6b', 12018


def facts(path):
    info = os.lstat(path)
    data = path.read_bytes()
    return data, {'path': str(path), 'sha256': hashlib.sha256(data).hexdigest(), 'bytes': len(data),
                  'mode': stat.S_IMODE(info.st_mode), 'regular': stat.S_ISREG(info.st_mode),
                  'inode': info.st_ino, 'mtime_ns': info.st_mtime_ns, 'ctime_ns': info.st_ctime_ns}


def main():
    if not (sys.flags.isolated and sys.flags.dont_write_bytecode):
        raise SystemExit('run with -I -B')
    if 'OPENROUTER_API_KEY' in os.environ:
        raise SystemExit('key present')
    good, retained = facts(RETAINED)
    if (retained['sha256'], retained['bytes'], retained['mode'], retained['regular']) != (FROZEN_SHA, FROZEN_BYTES, 0o444, True):
        raise SystemExit('retained object mismatch')
    bad, before = facts(TARGET)
    if (not before['regular'] or before['mode'] != MODE or before['bytes'] != DRIFT_BYTES
            or not before['sha256'].startswith(DRIFT_PREFIX)):
        raise SystemExit('target is not in the diagnosed drift state')
    STORE.mkdir(mode=0o700, parents=True, exist_ok=True)
    kept = STORE/before['sha256']
    with kept.open('xb') as out:
        out.write(bad); out.flush(); os.fsync(out.fileno())
    kept.chmod(0o444)
    _, kept_facts = facts(kept)
    if kept_facts['sha256'] != before['sha256']:
        raise SystemExit('drifted copy mismatch')
    fd = os.open(TARGET, os.O_WRONLY | os.O_TRUNC | os.O_NOFOLLOW)
    try:
        view = memoryview(good)
        while view:
            view = view[os.write(fd, view):]
        os.fsync(fd)
    finally:
        os.close(fd)
    _, after = facts(TARGET)
    if (after['sha256'], after['bytes'], after['mode'], after['regular'], after['inode']) != (
            FROZEN_SHA, FROZEN_BYTES, MODE, True, before['inode']):
        raise SystemExit('restore postcondition failed')
    json.dump({'schema': 'scout.x1.closure-file-restore.v1', 'retained': retained, 'before': before,
               'drifted_copy': kept_facts, 'after': after}, sys.stdout, sort_keys=True, indent=2)
    sys.stdout.write('\n')


if __name__ == '__main__':
    main()
```

**Properties:**
- No `O_CREAT`, no temp file and no chmod on the target, so the inode, mode and directory entries stay the same.
- The retained object is opened read-only.
- The drifted bytes are kept, content-addressed and read-only, under SCRATCH.
- Every precondition fails closed before any write.
- The mtime changes; `Inventory` does not record it, and it is reported in `after`.

**Authority:** this restore is a recovery of action-(b) state, outside amendment `:74`. It returns a P3-frozen file to its signed bytes and changes no signed artefact. This record authorises the one run of the frozen script. Round 4 must review its evidence before any launch.

### C. Binding repair (X1EI-R3-01/02), a `binding/**` repair under `:74`

**C1** `interlock.py` `stub_source`:
- The first line becomes `'#!'+str(python)+' -ISB\n'`. Nothing else in the source changes.
- `-I` ignores `PYTHON*` variables, `-S` skips site and `.pth` processing so `_distutils_hack` is never imported, and `-B` suppresses bytecode writes.
- The single grouped token avoids differences in how shebang arguments are split.
- The stub needs only the stdlib `json`, `os` and `sys`. Its logged `argv` (`sys.argv`, the stub path first) is unchanged, so `evaluate` (`interlock.py:36–60`) is unchanged.

**C2** `acceptance.py` `private_env`:
- Add `'PYTHONDONTWRITEBYTECODE': '1'` to the environment literal. This is the same setting as the signed attestor environment (`attest_closure.py:645`).
- It covers any unsandboxed Python child that inherits the environment, such as the contextual lane through `modelcli.Env()`.
- It is not the stub's guarantee; C1 is. Clawpatch's claude provider replaces the child environment (`.tools/clawpatch/734f81bc…/dist/providers/claude.js:164`, `replaceEnv: true`), and `-I` ignores it anyway.

**C3** `acceptance.py` `run_revision`:
- Immediately after the `with (run/'stdout')… subprocess.run(…)` block closes, and before `attempts=` is read, add `verify_installed_closure(closure_path)`, using the function already imported at `:113`.
- This runs on every route, including the positive control's allowed nonzero exit. A drift raises `ValueError('installed runtime file/link/mode set drift')`.
- The existing pre-launch call stays.
- Cost: one extra closure check per launch, 12 per synthetic operation and 8 per Django operation. At the measured 28.4–54.5 s per check (GV and round-2 evidence), that adds about 6–11 min and 4–7 min.

**No other changes:**
- No other file under `execution/` changes, and no file is added there.
- `runtime.json`, `main.go`, the frontend binary, the closure, `attest_closure.py`, `preparation.py`, the amendment and its receipts stay unchanged.

### T. Tests

Each test has a planted effect and a passing null. Tests go in the existing `test_binding.py` and `test_binding_boundaries.py`. Every write a test makes stays inside its own temp directory; no test touches `scout-x1`.

**T1 (C1) stub flags.**
- **T1a, static.**
  - Null: the first line of `stub_source('/synthetic/python', log)` equals `#!/synthetic/python -ISB`.
  - Planted effect: the pre-repair line `#!/synthetic/python` fails the same assertion.
- **T1b, real execution**, for both `claude` and `codex`, with `env={'PATH':'/usr/bin:/bin','PYTHONPYCACHEPREFIX':str(T/'pycache')}`:
  - Planted effect: a stub written from `stub_source(sys.executable, log)` with ` -ISB` removed from line 1 exits 97, logs one row, **and** leaves at least one `*.pyc` under `T/pycache`.
  - Null: `make_stubs(T/'stubs', sys.executable)` exits 97, logs exactly one row that passes `evaluate(rows, positive=True, expected_provider=…)`, **and** creates no `T/pycache`.
  - T1a carries the `-B` guarantee; T1b proves that bytecode writes are observable and that the repaired stub still meets its contract.

**T2 (C2) environment.**
- Null:
  - `private_env(...)['PYTHONDONTWRITEBYTECODE'] == '1'`;
  - `[sys.executable, '-c', 'import json']` under `private_env(...)` plus `PYTHONPYCACHEPREFIX=T/pycache` creates no `T/pycache`.
- Planted effect: the same command with the `PYTHONDONTWRITEBYTECODE` key removed creates at least one `*.pyc` there.

**T3 (C3) post-launch verification**, against a **real** installed-closure receipt over `T/runtime`, built as in `test_closure_review.py:238–248` (`Inventory`, `retain`, `checked`):
- **Null:**
  - a frontend that writes only its outputs in the run directory passes;
  - a spy around the real `verify_installed_closure` records the order verify → launch → verify, with exactly two calls.
- **Planted effects**, each raising `ValueError` matching `file/link/mode set drift` with exactly one launch:
  - (a) route `bound`, exit 0; the frontend appends a byte to `T/runtime/file`;
  - (b) route `synthetic-unbound`, not resolve-only, exit 1, with a valid claude attempt row; same mutation. This shows the positive-control path does not bypass the check;
  - (c) route `reference`; the frontend creates `T/runtime/added`.
- **Existing assertions updated:**
  - `test_binding.py:131` becomes `closure.call_args_list == [call(root/'closure')]*2`;
  - `test_binding_boundaries.py:91` becomes `closure.call_count == 2`;
  - the key-present branch keeps `closure.assert_not_called()`.

**T4 (integration)** is the fresh rehearsal in step 9 of the sequence. It must:
- pass with all six predicates true;
- show both providers' positive controls (`claude` and `codex`; the codex control has never run);
- be followed by a passing real closure check.

### O. Operator rule (added to P3)

Until the seal:
- Never start `scout-x1/.tools/python/…/python3.11` or `scout-x1/.tools/venv/bin/python` without `-B`, or without `-I -B` for any study command.
- Never run bootstrap, uv, npm or pnpm against `scout-x1`.

The restored pyc is still stale by design, so any such start rewrites it again (closure-impl round 1, S10).

### D. Deviation text for `RESULTS.md` `## Deviations`

Main writes this after the restore, filling only the measured values marked `<…>`:

> ### Installed runtime pyc rewritten during the first binding-v2 scratch rehearsal (X1EI-R3-03)
>
> On 2026-11-10 the first fresh binding-v2 scratch rehearsal ran unjournalled from 18:28:03Z to 18:33:30Z (`preparation-reviews/binding-v2-rehearsal-run-2026-11-10.json`).
>
> **What happened.**
> - Its positive-control `claude` stub ran the attested base interpreter `scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11` without `-B`.
> - Site startup (`distutils-precedence.pth`) imported `_distutils_hack`, found the shipped pyc's header stale, and at 18:32:59Z rewrote `…/site-packages/_distutils_hack/__pycache__/__init__.cpython-311.pyc`.
> - The file went from 12015 B `71093d3e…` to 12018 B `9c815f6b…`. The decoded code differs only in `co_filename`.
>
> **Consequence.**
> - This breached P3 byte stability for one closure-enumerated file, which has two overlapping root records.
> - The next launch's `verify_installed_closure` refused. The rehearsal exited 1 with every predicate false. It is void and superseded, not evidence.
> - No journalled, canonical or sealed operation ran while the file differed. P3-D observations (OS, boot session, dyld cache) were unaffected.
>
> **Restore.**
> - Under [execution-implementation round 3](preparation-reviews/execution-implementation-round-3.md), Main restored the closure bytes in place from the retained object `SCRATCH/closure-retained/sha256/71093d3e…` at `<UTC time>`.
> - Evidence: `<restore record path, sha256>`.
> - Afterwards `verify_installed_closure` and `verify_tree_pins` passed, and the enumeration comparison reported zero differences.
> - The drifted bytes are retained at `SCRATCH/binding-v2-restore/drifted/<sha256>`.
>
> **Limit.** The closure check compares bytes, modes and file sets at the moment it runs. It cannot exclude a write-then-restore between checks.
>
> **Binding repair.** The stubs now run with `-ISB`, the binding environment sets `PYTHONDONTWRITEBYTECODE=1`, and the closure is re-verified after every launch. This was signed off in `<round-4 record>`.

## Answers to the questions

1. **The installed file.**
   - Yes, it can be restored byte-identically from the retained object (F2), by R.
   - `Inventory` does not compare mtime.
   - The restore is recorded in a restore evidence JSON plus deviation D.
   - P3-D needs no deviation: no OS, boot or cache item changed, and the P3-D session checks passed and matched O0.
   - P3 does need a deviation (D): the breach, not the restore, is the deviation.
2. **The binding code.** C1–C3 with tests T1–T4, as specified above.
   - Outer write-confinement of the frontend tree is not adopted (see Alternatives).
   - The tool matrix is F3. No launched tool other than the stubs is known to write into closure roots, and C3 makes any write fatal at the launch that made it.
3. **The closure's integrity claim.**
   - The attestation in `closure.json` stays true for the time it was recorded. The round-1 evidence `.checks[0]` passed the real check at 18:10:40Z, after closure and before the rehearsal.
   - Continuous byte stability (P3) was broken from 18:32:59Z until the restore. That requires D.
   - No recorded attestation, receipt or journal is invalidated. The only process known to have loaded the rewritten bytes was the stub, which exited 97.
4. **The authorised sequence** is below.

## Alternatives

| Option | Tradeoff | Decision |
|---|---|---|
| **Chosen:** restore R + C1 `-ISB` + C2 env + C3 post-launch check | Small, local change to the signed binding. Prevention where the mechanism is known; detection at the launch for everything else. Adds about 5–11 min per operation. | Adopted |
| `PYTHONDONTWRITEBYTECODE` only | Dropped by Clawpatch's `replaceEnv` and ignored under `-I`, so it does not fix the stub. | Rejected |
| `PYTHONPYCACHEPREFIX` (peer variant) | Redundant with C2: the cases where C2 is lost (`-I`, a replaced environment) also lose the prefix. It also forces in-memory recompilation of every import. | Rejected |
| Final check only, once after the last launch of `execute_fixture` | Same detection, because pre-launch k+1 already covers post-launch k. Cheaper, but needs more test patching (`test_binding.py:408`, `test_binding_boundaries.py:251`) and relies on reasoning about launch order. | Rejected in favour of a per-launch check |
| Outer kernel write-denial (`sandbox-exec`) around the frontend tree | Strongest prevention for all descendants. But it contradicts the reviewed "no outer sandbox" amendment, risks nesting with the scanners' `sandbox-exec` [INFERENCE], and needs a new protocol amendment and round. | Rejected now; available as an amendment if C3 ever trips again |
| `chmod`/`chflags uchg` on closure roots | Mode is in the inventory and would drift. `uchg` mutates `scout-x1` beyond read-only reuse. | Rejected |
| Make the pyc fresh (recompile or `touch` the source) | Changes pinned bytes or another closure file, and leaves about 5,800 uncached modules exposed. | Rejected |
| Re-attest the closure on the new bytes | `closure.json` is create-once and P4 allows one canonical run. That means a new action-(b) recovery, a new closure and new binding confirms. | Rejected |
| Restore only after round 4 (peer's order) | Round 4 could not then review the restore evidence, which costs an extra round. R is deterministic, guarded and keeps the drifted bytes, and no launch happens before round 4 either way. | Rejected |

## Evidence

- `execution/binding/interlock.py:13–21`.
- `execution/binding/acceptance.py:26–38,109–124,222–240`.
- `execution/attest_closure.py:85–129,132–162,474–493,543–545,645`.
- `execution/preparation.py:702–713`.
- `internal/sandbox/sandbox.go:76–83` and `internal/sandbox/run.go:40–48` (scout-x1).
- `internal/contextual/cli.go:81–84` and `internal/modelcli/modelcli.go:88–110`.
- `.tools/clawpatch/734f81bc…/dist/providers/claude.js:164`.
- `…/site-packages/distutils-precedence.pth`.
- The retained stub `SCRATCH/binding-v2-rehearsal/report.json.work/logs/positive/parent/stubs/claude` and the work-tree layout (no `logs/positive/head`).
- `closure.json` queries: the `retained_files` row, two enumeration records (mode 420, 12015 B), the `original_runtime_files` row, and `__pycache__` size 96.
- `preparation-reviews/execution-implementation-round-1-evidence.json` `.checks[0]`: a real closure check passed at 18:10:40Z.
- `closure-impl-round-1.md:68` (S10).
- `PREREGISTRATION.md` §11, the deviations rule.
- Python documentation: [command line `-I`, `-S`, `-B`](https://docs.python.org/3.11/using/cmdline.html) and [cached bytecode invalidation](https://docs.python.org/3.11/reference/import.html#cached-bytecode-invalidation).

## Risks

- **A partial restore.** If the write fails after `O_TRUNC`, the file is truncated and the script stops. A rerun then fails its precondition by design. The drifted bytes are already kept and the retained object is untouched, so recovery is a new round with the same inputs.
- **A write outside the known mechanism** during a later launch. C3 refuses the operation at that launch, before any canonical receipt passes. That costs a round, not integrity.
- **Write-then-restore between checks is undetectable.** This is stated in D.
- **H5**, an outside process in the same window, is not fully excluded. T4 under C1 settles it: if drift recurs with `-ISB` stubs, the cause is elsewhere and the run stops at the launch.
- **Cost:** about 5–11 min per operation. This is small against the 2026-12-24 deadline.
- **Reversibility:** the code is a three-line change plus tests. The restore is reversible from the kept drifted bytes. The canonical steps stay irreversible and come later.

## Authorised sequence

Stop at the first failure. Any failure means: retain everything, create nothing more, and open a new council round.

1. **Write this record.** Main writes it to `STUDY/preparation-reviews/execution-implementation-round-3.md` (UTF-8, LF) and records its sha256 and bytes.
2. **Freeze.**
   - No binding launch until step 9.
   - Rule O is in force.
   - Keep `SCRATCH/binding-v2-rehearsal/` (with `.work`), `SCRATCH/binding-v2-rehearsal-diag/` and both `preparation-reviews/binding-v2-rehearsal-*` records unchanged.
3. **Restore.**
   - Run P3-D session checks 1–3.
   - Write `SCRATCH/binding-v2-restore/restore_pyc.py` verbatim from R and record its sha256.
   - Run `/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B SCRATCH/binding-v2-restore/restore_pyc.py`, with stdout captured to `SCRATCH/binding-v2-restore/restore.json`. It must exit 0.
   - Then run, unchanged:
     - the argv of `execution-implementation-round-1-evidence.json` `.checks[0]` (real `verify_installed_closure` plus `verify_tree_pins`; both must pass);
     - `compare_enumeration.py` `5b23f238…` under `-I -B`, which must report zero differences.
   - Record everything in `preparation-reviews/binding-v2-restore-2026-11-10.json`: argv, cwd, environment with secrets checked by presence only, exit codes, stdout and digests, and timestamps.
4. **Implement** C1–C3 and T1–T3 in `binding/interlock.py`, `binding/acceptance.py`, `tests/test_binding.py` and `tests/test_binding_boundaries.py` only.
5. **Run the tests** with `/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest`:
   - the binding suites (`test_binding*.py`);
   - `test_binding_contract.py`;
   - the full Python suite.

   Then rerun `.checks[0]` to show the suite left the closure intact. Record argv, exit codes and timestamps in a checks artefact.
6. **Regenerate at fixed paths (`:74`).**
   - Move `OUT/execution-source-manifest.json` `25ef2b4e…` and `OUT/reviews/implementation.json` `8711ffae…` byte-identically to `SCRATCH/binding-v2-build/superseded-r2/`, with before/after pins.
   - Regenerate the manifest with round 2's producer and argv. Exactly the four changed files may differ, compared row by row.
   - `OUT/binding/runtime.json` stays `647381c5…`; recompute it to confirm.
7. **Prepare for round 4.**
   - Main writes `execution-implementation-round-3-response.md` with the new digests, the step-3 and step-5 evidence, and the deviation D as applied to `RESULTS.md`.
   - Main adds research-ledger rows, at the next free R-IDs, for the two Python documentation sources cited here.
8. **Council round 4** (execution-implementation plus the `:74` binding repair). It needs zero issues and must sign:
   - the new `interlock.py`, `acceptance.py` and test digests, and the manifest;
   - the restore evidence;
   - the 9-row scope.

   Main then writes `OUT/reviews/implementation.json` naming the round-4 record.
9. **Fresh rehearsal.**
   - Run P3-D session checks 1–3.
   - Run the fresh synthetic-only rehearsal with the round-2 step-5 argv, except `--output SCRATCH/binding-v2-rehearsal-2/report.json`. Neither that path nor its `.work` may exist.
   - Then rerun `.checks[0]` (T4).
10. **After the rehearsal.**
    - Council rehearsal review. On zero issues, write `OUT/reviews/binding-v2-rehearsal.json`.
    - Then canonical synthetic with `--rehearsal-report`, then Django, then `reviews/binding.json`.
    - Then the remaining reviews, `seal --check` and `seal`.
    - RB-1 stays in force: no further `adopt`.

## Suggestions (not conditions)

- **S1.** Round-2 S1 is carried forward: node shadowing or absence, runtime git ≠ closure git, closure-side duplicate support rows, and a refusal through `reviewed_inputs`. Add these if convenient in step 4.
- **S2.** `execution/closure/__pycache__/native.cpython-314.pyc`, about 2 h old, shows that some tool imported study code with an unpinned Python 3.14 without `-B`. It lies outside closure roots and both manifests exclude `__pycache__` (`preparation.py:711`, `acceptance.py:456`), so integrity is unaffected. Under rule O, run study code only with the pinned `-I -B` interpreter.
- **S3.** Rejected now, but the outer write-confinement option (Alternatives) is the escalation path if C3 ever trips in a later run.

## How the peers shaped this record

- **innovation-challenger:**
  - established that Clawpatch's claude provider replaces the environment (`replaceEnv: true`) and that `-I` ignores `PYTHON*` variables, which made C1, not C2, the decisive fix;
  - attributed the 80 attempts to the contextual launcher;
  - supplied the launch/write matrix and timing against the scanner starts;
  - proposed the post-launch check;
  - framed kernel write-confinement as the amendment-level alternative.
- **experiment-peer:**
  - replaced the council's planned shebang `-B` with `-ISB`;
  - read the pyc header and build-path `co_filename`, showing the pyc was stale since installation (refuting a copy from another checkout);
  - specified the in-place `O_TRUNC` restore and its mode hazard;
  - identified the seal-only gap as concrete through `preparation.py:702–703` (X1EI-R3-02);
  - reported S2.
  - The council did not adopt its `PYTHONPYCACHEPREFIX` addition (redundant), nor its order of restoring only after a zero-issue round (it would add a round with no safety gain, since no launch precedes round 4).
- **Main council:**
  - read the retained stub, the `.pth` hook, the enumeration, retained and original rows, the round-1 `.checks[0]` timing and `preparation.py:711`, which put the restore script outside `execution/`;
  - selected the per-launch check placement and the test design;
  - chose to authorise the restore in this round.

Verdict: REMEDY CHOSEN — implementation review required
