# X1 preparation evidence

No inventory, model evaluation, gold labels or study analysis results are reported here.

## Deviations

### Producer bytes unrecorded (X1C-R1-06)

The eight passed preparation operations below predate automatic producer-byte capture. Their immutable journals do not pin the entry script and loaded execution modules. **Producer bytes unrecorded**; the later execution-source manifest must not be interpreted as the bytes that produced these earlier records.

- Action (a): `172d394fab834f629a7587d62fdd4ee7`, `85d89edca80243a69e634e924d4fa179`, `bb92520b7b7849458ff4b21e7b2a5a12`, `bf5490fb8c7d46b38ab036c4c7d855ab`, `bf2bbe582def4439a1d56b745d995d8e`.
- Action (b): `d8414b51d4ae4420a1dbbf4e300f67b5`, `3f969fafa01244478350d2d37e52c657`.
- Action (d): `906389d7212a401b8c7f781d37960d84`.

The [Core round-1 review](preparation-reviews/core-impl-round-1.md#effect-on-executed-evidence) retains the input-gate and executed action-(a) evidence as valid. Object identities, headers, boundaries and fsck are rechecked by the seal; old create-once records are not rewritten or rerun. The seal lists each affected passed journal with reason `producer bytes unrecorded`. New journals capture producer-file pins and the prior journal inventory.

### Offline Go graph refusal and add-only cache remedy (4124…; B′)

Preflight `412402435e9f4ac9bb810a5520d102d5` refused because offline `go list -m -json all` required `.info` metadata missing for 15 unselected modules in the 28-module graph. The original cache copy and refusal were retained unchanged. [Remedy round 1](preparation-reviews/closure-preflight-remedy-round-1.md) selected B′: a new `closure-modules-r2` copy, with add-only metadata population under separate setup journals, followed by an offline preflight. No signed implementation was edited. The [gate evidence](preparation-reviews/closure-preflight-remedy-gates-2026-11-10.json) records preservation of the 14,494 original entries and addition of 15 `.info` files plus two directories. The [recovery receipt](preparation-run/reviews/operation-recovery-412402435e9f4ac9bb810a5520d102d5.json) retains the refusal; it is not passing closure evidence.

### Weak-link refusal and additive collector (0824…; F1)

Preflight `082476b0335b4deebe7390ca6b426c0a` refused at RemoteProcessingBlock. The signed parser discarded link attributes and treated an absent weak dependency as required: CoreAudio's edge was `delay-init weak-link`, and AudioDSPGraph's was `weak-link`. The missing-image probe exited 0 with empty stdout, so exit status alone did not establish identity or absence. [Remedy round 2](preparation-reviews/closure-preflight-remedy-round-2.md) and its [recovery receipt](preparation-run/reviews/operation-recovery-082476b0335b4deebe7390ca6b426c0a.json) preserve that refusal and its causal records unchanged.

The additive `execution/attest_closure_v2.py` entrypoint and `execution/closure/native_v2.py` collector were [signed off in implementation round 3](preparation-reviews/closure-weaklink-impl-round-3.md). The wrapper guards the signed base digests; no pinned base file changed. Required dependencies still refuse. Weak absence requires negative active-cache membership, no filesystem entry at the target or enumerated cryptex overlays, and a successful empty-stdout probe. The full runtime rehearsal and canonical collection recorded four absent target names:

- `/System/Library/PrivateFrameworks/RemoteProcessingBlock.framework/Versions/A/RemoteProcessingBlock`;
- `/System/Library/PrivateFrameworks/IntelligencePlatformLibrary_AppleInternal.framework/Versions/A/IntelligencePlatformLibrary_AppleInternal`;
- `/System/Library/PrivateFrameworks/MLCompilerRuntime.framework/Versions/A/Libraries/libmlc_rt.dylib`;
- `/usr/lib/libobjc-env.dylib`.

The [execution gate](preparation-reviews/closure-weaklink-gates-2026-11-10.json), SHA-256 `9e34808c05b3a521e98ee286c22446a34e61c78a28255887a3df4cb253170426`, retains each absence proof, command argv/output digest and timestamp. Passed preflight journal `5a6010880cad41a680d000a350807417` produced SHA-256 `7b060348a482fe2a0356c00166b587474aae7cbb56346a86a2e7bc3129c55232`; canonical journal `c779b7b5666e4099b093003c373f11f5` created [closure.json](preparation-run/closure.json) once, SHA-256 `b5df46b1b92a3cc9ef2905589db120a008881bc6b5519f37b6d38e3705cc35b2`.

### Present system alternates and loaded-image identity (I7)

Three reachable present files, `/usr/lib/system/libsystem_{kernel,platform,pthread}.dylib`, are also active-cache members. Their multiple arm64-family slices make the signed disk classifier refuse. The reviewed I7 amendment resolves these cached-parent edges as cached images and selects the architecture from the UUID-matched cache magic. `dyld_info -arch arm64e` still reads the disk slice, not necessarily the cache. The collector therefore compares its UUID case-insensitively with the unique matching image's in-memory `LC_UUID`, requiring its entire header/commands within the active cache range. A mismatch or unestablished identity refuses. Only these present-member records carry `identity_check`.

The [round-2 response](preparation-reviews/closure-weaklink-impl-round-2-response.md) records equal loaded/disk UUIDs for all three images and the synthetic refusal tests; the [round-3 response](preparation-reviews/closure-weaklink-impl-round-3-response.md) adds the isolated below-range guard test. This is a UUID identity policy, not a claim that loaded and disk bytes were compared byte-for-byte. Sources: R-088–R-091.

### Real-null criterion replaced (R3 → R3′a–c)

The original `/bin/sh` passing-null comparison cannot complete on signed v1: it refuses at RemoteProcessingBlock and would also encounter the fat system alternates. [Implementation round 1](preparation-reviews/closure-weaklink-impl-round-1.md#q2-r3-replaced-not-accepted) explicitly replaced, rather than silently relaxed, R3. R3′a preserves the exact signed-v1 refusal; R3′b compares both collectors on the real dependency-free `race_darwin_arm64.syso`, pinned SHA-256 `6ca6a32e8b650ac03b6a11b40c7a0cdd938d00046bd1c942f7c530e0fdaf23b2`; R3′c compares every cached image reached before v1's refusal and multiset inclusion of its actual metadata command records against completed v2. Synthetic S12 remains the visit-order parity check. All five real-byte tests ran without skips in the signed-off 339-test suite; retained run manifests bind 1,590 command records.

### Update prevention replaced by detection-and-stop (M7 → P3-D)

M7 recorded `AutomaticDownload=1` and `SplatEnabled=1`. Prevention could not be achieved on the managed Mac without root; no settings change or privilege escalation was attempted. The council replaced prevention with [P3-D](preparation-reviews/closure-weaklink-impl-round-3.md#p3-d-p3-and-p4-restated-with-the-v2-argv): retain OS/build, boot-session, active-cache UUID, scheduled cache-family hashes, software-update preferences/managed preferences and history; stop before the next step on drift or failed observation. O0–O4 passed, the preflight/canonical outputs matched O0, and the boot session stayed constant from O1 through O4. The protocol remains required through seal, including session checks before binding launches. The owner does not restart or accept update prompts from O1 until seal; an MDM deferral request is non-blocking. Detection does not prevent a forced update: availability/recovery risk remains. Sources: R-092–R-093.

### Native and verification rehearsals (GR and GV)

The unjournalled GR called the real attestation with exact parsed preflight arguments and stub ID `remedy-r2-gr`, stopping only at the reviewed retention sentinel. It saved the full native result, proved the weak-absence records and all three `identity_check` conditions, and passed. It created no Journal. After the passed v2 preflight, read-only GV passed `verify_installed_closure` in `28.39171587501187` seconds and `verify_tree_pins` over 92,066 pins in `15.660364750074223` seconds. Neither duration exceeded the one-hour clock-risk threshold. The gate retains the harness, outputs, exact argv and journal pins; no study inventory or model evaluation was authorised by these preparation gates.

### Frontend build outside the protected-cache command rule (X1EI-R1-02)

After canonical closure (b), the binding frontend was built directly and unjournalled using Go 1.26.10, with `GOMODCACHE=/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-modules-r2`, contrary to closure-preflight-remedy-round-1 Step 11's restriction on direct Go commands against that cache. The build's exact argv is retained in `/Users/adamjackson/Projects/work/x1-preparation-scratch/binding-v2-build/preparation-report.json`; driver SHA-256 is `5742f0004c49c005313d88cd06df964b09cad6bcf2eac4506bdb90abc923b1f4`, source `main.go` SHA-256 is `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078`, and binary SHA-256 is `05fa048c4911666e107e8e272770308cef3b23535b6c0dd208f36f7768a0a139`. The binary was not rebuilt during repair.

Council-requested read-only integrity evidence passed: installed closure verification and 92,066 recursive pins; a complete 14,511-entry cache stat listing with zero mtime/ctime values later than the build-directory creation time; and `go version -m` under a fresh empty module cache, with all 12 dependency h1 sums matching closure-verified sums. The cutoff was the build directory's birth time, `1794333048955812608` ns, not its mutable mtime/ctime. The stat listing is timestamp evidence, not a byte-for-byte snapshot of uncovered download files. Metadata-output SHA-256 is `339b679c3aba00a9dfb84ace408ac70b705d9a2c739e49b3745af4c85badce1d`.

Full [integrity evidence](preparation-reviews/execution-implementation-round-1-evidence.json), SHA-256 `e6b624b5c7eebcf590575d1fd8431fa062bbc82664236d3160321fd1cbd30d6c`, and the [repair response](preparation-reviews/execution-implementation-round-1-response.md#x1ei-r1-02--read-only-integrity-evidence) preserve the observations. They support retaining the existing binary but do not erase the command-rule breach or claim reproducible binary identity from a second build. Acceptance remains subject to the execution-implementation zero-issue review, not these observations alone.

### Installed runtime pyc rewritten during the first binding-v2 scratch rehearsal (X1EI-R3-03)

On 2026-11-10 the first fresh binding-v2 scratch rehearsal ran unjournalled from 18:28:03Z to 18:33:30Z (`preparation-reviews/binding-v2-rehearsal-run-2026-11-10.json`).

**What happened.**
- Its positive-control `claude` stub ran the attested base interpreter `scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11` without `-B`.
- Site startup (`distutils-precedence.pth`) imported `_distutils_hack`, found the shipped pyc's header stale, and at 18:32:59Z rewrote `…/site-packages/_distutils_hack/__pycache__/__init__.cpython-311.pyc`.
- The file went from 12015 B `71093d3e…` to 12018 B `9c815f6b…`. The decoded code differs only in `co_filename`.

**Consequence.**
- This breached P3 byte stability for one closure-enumerated file, which has two overlapping root records.
- The next launch's `verify_installed_closure` refused. The rehearsal exited 1 with every predicate false. It is void and superseded, not evidence.
- No journalled, canonical or sealed operation ran while the file differed. P3-D observations (OS, boot session, dyld cache) were unaffected.

**Restore.**
- Under [execution-implementation round 3](preparation-reviews/execution-implementation-round-3.md), Main restored the closure bytes in place from the retained object `SCRATCH/closure-retained/sha256/71093d3e…` at `2026-11-10T18:54:01.536494227Z`.
- Evidence: [binding-v2-restore-2026-11-10.json](preparation-reviews/binding-v2-restore-2026-11-10.json), SHA-256 `1c5a5dce4c6ddb2ec2aea34be391dfa7964f9b7fe96933de0174c982c3c74690`.
- Afterwards `verify_installed_closure` and `verify_tree_pins` passed, and the enumeration comparison reported zero differences.
- The drifted bytes are retained at `SCRATCH/binding-v2-restore/drifted/9c815f6bb55ad4e93c1b74a6c729e234f3838ef8067f201a2441c32cfffdabaf`.

**Limit.** The closure check compares bytes, modes and file sets at the moment it runs. It cannot exclude a write-then-restore between checks.

**Binding repair.** The stubs now run with `-ISB`, the binding environment sets `PYTHONDONTWRITEBYTECODE=1`, and the closure is re-verified after every launch. Signed off with zero issues in [execution-implementation round 4](preparation-reviews/execution-implementation-round-4.md). Sources: R-094–R-095.

### Reproduce commands and create-once limits

The exact historical GR, preflight, GV and canonical argv—including all 32 external roots—are in the gate's `operations` rows. Print them without rerunning or reserving any output:

```sh
/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python -I -B - <<'PY'
import json, shlex
from pathlib import Path
gate = json.loads(Path('/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-weaklink-gates-2026-11-10.json').read_bytes())
for row in gate['operations']:
    if row['step'] in ('Step-4-GR', 'Step-5-v2-preflight', 'Step-6-GV', 'Step-7-canonical-once'):
        print(row['step'] + ': ' + shlex.join(row['argv']))
PY
```

The original frontend build command and isolated read-only metadata command are retained, with their environment and cwd, in the [repair response](preparation-reviews/execution-implementation-round-1-response.md#s5-original-build-provenance) and integrity evidence. These are provenance/reproduce records, not permission to rerun the direct protected-cache build, create-once preflight/canonical outputs, or the GR harness at its already-used paths. Signed reviews, receipts, copied runtimes and failed-operation environments remain retained unchanged until seal.


## Operator retention and seal preflight

`brew pin git` completed successfully with no output on 2026-11-09. Retain that Git keg unchanged until sealing; do not upgrade or clean it up. This is retention, not a claim of study authority.

Until seal, never start either the base Python under `scout-x1/.tools/python/` or `scout-x1/.tools/venv/bin/python` without `-B`; study commands require the pinned interpreter with `-I -B`. Never run bootstrap, uv, npm or pnpm against `scout-x1`. The restored pyc intentionally remains stale. These are [round-3 rule O](preparation-reviews/execution-implementation-round-3.md#o-operator-rule-added-to-p3), in force before any further binding launch.

After all required reviews and outputs exist, Main must run the read-only preflight immediately before the real create-once seal:

```sh
"$PY" -I -B "$EXEC/preparation.py" seal --check --config "$EXEC/run-config.json" --clock "$OUT/clock.json"
"$PY" -I -B "$EXEC/preparation.py" seal --config "$EXEC/run-config.json" --clock "$OUT/clock.json"
```

Do not run the second command if the first refuses. Neither a passing preflight nor a seal authorizes inventory.
