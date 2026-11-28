# Closure implementation review round 3 — response

Status: R3-01 fixed, submitted for round-4 review. This response is not a council sign-off and does not authorise a preflight or canonical closure. No bootstrap, real scanner/model invocation, network fetch, canonical closure or journalled preflight was run. The historical installation and signed inputs were not modified.

## Required issue R3-01

`execution/closure/native.py:43–55` now tests only the fat interpretation's file bounds before skipping Java-range CAFEBABE bytes:

1. If the interpreted fat32 architecture table exceeds the file, return non-native.
2. Otherwise read every entry, without allocating the full table. Return non-native only if an interpreted slice extends outside the file.
3. If every interpreted slice is bounded, refuse `ambiguous Java/fat Mach-O image`, regardless of whether the attestor's stricter count, alignment, thin-header or CPU checks would reject it.

Thus a native image cannot disappear merely because its architecture count collides with a Java version word. No filename/suffix exemption or full Java parser was added. Java 52.1 also no longer refuses solely because its minor version is nonzero (S5).

Catching tests in `execution/tests/test_closure_round3.py`:

- `test_r301_45_slice_native_collision_never_skipped`: 45 complete, nonoverlapping, 16 KiB-aligned thin slices, exactly one arm64. Classification must raise the ambiguity refusal, never return None. Reverting the round-3 change would return None and fail this test [INFERENCE from the round-2 branch].
- `test_r301_java_52_0_and_45_3_remain_non_native`: complete tiny Java 52.0 and 45.3 class fixtures remain non-native; includes 52.1.
- `test_r301_java_only_skipped_when_a_fat_slice_escapes_file`: the whole table fits and only its final interpreted slice escapes, proving every entry is checked.
- `test_r301_bounded_fat_refuses_even_if_stricter_checks_would_fail`: a bounded interpretation with zero-sized/non-thin slices must still refuse ambiguity, not exploit stricter validation to skip bytes.

S4 classifier coverage was also expanded: non-native null; truncated fat version; zero/excess architecture count; bad slice magic; bad thin filetype and command bounds; fat64 reserved field; excessive alignment/table overlap; empty slice; CPU mismatch; overlapping slices; missing/duplicate arm64; and passing x86_64 + arm64e and swapped-endian fat fixtures. Existing round-2 tests retain truncated-table/header, out-of-bounds and misalignment coverage. The other optional suggestions were not adopted in this narrowly scoped fix; in particular, this does not claim new retention-destination tests, inherited-rpath support, broader dry-check inventory, or recheck-cost changes.

## Shared preparation.py dependency

Read the final `execution/preparation.py`, including `adopt`, load-time adoption/journal checks, `operation_lock`, `Journal`, and the seal path. Its SHA-256 was checked before and after the authorised real-byte dry check and full Python suite and remained:

`599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427`

`verify_inputs.py` remained `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`.

No closure API adjustment was necessary [INFERENCE from the read plus the passing synthetic integration suite]. `attest_closure.main()` holds `operation_lock(config)` across Journal construction, attest, output creation and finish. `adopt` requires the exclusive lock, while Journal's nested shared locks are reentrant. Journal also reloads and compares effective configuration before reservation; a configuration loaded before adoption therefore refuses instead of recording stale effective state. The reviewed-Go receipt's three-key schema matches the final `review_receipt` interface. Closure own_sources still pins this exact shared file in both the source journal and closure output; changing it after (b) remains forbidden.

The full suite against these bytes passed **198 tests in 14.738s**, exit 0. It includes the closure main/journal wiring and real orchestration tests, `test_adoption_racing_journal_is_serialized`, the handwritten/backdated adoption refusal test, and binding's real-adopt/load authority-chain integration tests. This is synthetic integration evidence, not a real canonical closure or a zero-issue review of Core.

## Real-byte check and limits

The authorised read-only dry check passed against `/Users/adamjackson/Projects/work/scout-x1` with `/opt/homebrew/Cellar/gettext/1.0`. It repeated the real Python RECORD/requirements, libpython, pnpm and npm checks and classified **all 2,493 gettext regular files: 32 Mach-O images and four Java classes**. The complete stdout below names and pins every classified image/class. It used -I -B and produced no canonical writes, journals or retained copies.

This is not proof of a successful full native load graph, dyld cache extraction, offline Go metadata walk, all-roots dry check, journalled preflight or canonical closure.

## Required run prerequisites (P1–P4 remain in force)

- P1: Main must obtain zero-issue Core/identifiers review closing R6-01 on the final shared digests, record them, and keep every source pinned by closure unchanged through the seal. The full-suite rerun against 599a128d… is supplied here; it is not a substitute for that sign-off.
- P2: Binding v2 must sign off the final main.go digest. Main supplies OUT/reviews/closure-model-free.json with exactly scope/record/reviewed; scope is closure-model-free-go, record pins the authentic binding sign-off, and reviewed contains exactly main.go, not main_test.go. This closure review must not be used as its Go authority.
- P3: Freeze the OS, dyld cache, Homebrew/mise kegs, CLIs and other attested runtime bytes until seal. Keep preflight outputs, retained bytes and closure-modules.
- P4: After zero-issue round 4 and the required all-roots read-only dry check, Main alone runs one journalled preflight with every runtime-dependency keg from the pinned Git installation receipt (the review records 16), explicit external-root arguments and the active dyld-cache-root. Run canonical closure once only after that preflight passes. Main supplies the complete approved root set, active cache directory/read privileges and retention storage. A refused/stale journal still requires reviewed recovery; no bypass or cleanup was attempted.

No product design or signed recipe bytes were changed. Changed production file: execution/closure/native.py. New synthetic test file: execution/tests/test_closure_round3.py. This response is outside execution/ so it does not alter its source-closure surface.

## Exact commands and complete outputs

All commands below ran in `/Users/adamjackson/Projects/work/scout-x1prep`. Empty stdout/stderr are explicitly retained. Offline Go checks used the pinned activation, disabled proxy/sumdb and toolchain download, and passed; go test reported its cached result, as shown. All checks ran once for this round, without failures.

### Read-only historical runtime / gettext dry check

```sh
/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python -I -B evaluation/x1-triage-ranking/execution/attest_closure.py --dry-check --runtime-root /Users/adamjackson/Projects/work/scout-x1 --external-root /opt/homebrew/Cellar/gettext/1.0
```

stdout:
```text
{
  "R1-01": {
    "applicable_names": 53,
    "deprecated": "1.3.1",
    "installed_distributions": 53,
    "librt": "0.16.0"
  },
  "R1-02": {
    "load_command": "@executable_path/../lib/libpython3.11.dylib",
    "python": {
      "bytes": 49968,
      "expected_resolved_path": null,
      "mode": 493,
      "path": "/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11",
      "requested_path": "/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python",
      "sha256": "570b41051d135bcfc011b41272224f7cc66b0d0e7f735761dcb8611029f51d6d",
      "shebang": null,
      "symlink_chain": [
        {
          "link_path": "/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python",
          "link_target": "/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11"
        }
      ]
    },
    "resolved_libpython": {
      "bytes": 18495136,
      "path": "/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/lib/libpython3.11.dylib",
      "sha256": "9bb0abed613dca8b5a578d2ebf0465a59e3c25946dddade90c2dbd6a90853efb"
    }
  },
  "R1-03": {
    "installed_packages": 54,
    "locked_packages": 133,
    "signed_assets": [
      {
        "bytes": 1562,
        "expected_resolved_path": null,
        "mode": 420,
        "path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/package.json",
        "requested_path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/package.json",
        "sha256": "b878e7ecf4f988a11ad9de8ae4412a32abdeff89ebf6fcc89296aea39ff4e818",
        "shebang": null,
        "symlink_chain": []
      },
      {
        "bytes": 45009,
        "expected_resolved_path": null,
        "mode": 420,
        "path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/pnpm-lock.yaml",
        "requested_path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/pnpm-lock.yaml",
        "sha256": "f546ec43c1dced67f3b3015dd1ba387fa3a999c648a30eda80475dc546de2b44",
        "shebang": null,
        "symlink_chain": []
      },
      {
        "bytes": 69,
        "expected_resolved_path": null,
        "mode": 420,
        "path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/pnpm-workspace.yaml",
        "requested_path": "/Users/adamjackson/Projects/work/scout-x1/.tools/clawpatch/734f81bc1514b52a0627edb0c24ebe1a18b36105/pnpm-workspace.yaml",
        "sha256": "beb856bbf5b1db5da0186b42aeb420f2a09526acdc76ce3db2843ec56844d45f",
        "shebang": null,
        "symlink_chain": []
      }
    ]
  },
  "R1-04": {
    "lock_listed_packages": 122
  },
  "R2-01": {
    "external_root_classifications": [
      {
        "mach_o_images": [
          {
            "filetype": 2,
            "image": {
              "bytes": 96304,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/envsubst",
              "sha256": "1a2d86e1e3d0554727a5d30aabf7c1e4313846b8fe5481a3c89a03de50193dae"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 77392,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/gettext",
              "sha256": "bb664022d6606c2f5754d1d1007caf0c6333c46f0ff58e62bafe1ac11a8bd43e"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 53648,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgattrib",
              "sha256": "b150d1ed528d5f4bc1ae3b3acfefc39f7fbb18b3c453c2a4b22d5d318b6137f6"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 53808,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgcat",
              "sha256": "91f5034966982ea5d050cec88dae8d1534d8377c3c635130f38042dcc144aab4"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 54528,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgcmp",
              "sha256": "5fa5ea0849cc9a46ef20c4a65c618e99889c4d05a9da1f34619408331fd1e4c5"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 53840,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgcomm",
              "sha256": "b5ac00c5fbd6c504d0b3d97a5fe20a219a64ea3242cb619f337958aff247aa69"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 53296,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgconv",
              "sha256": "00b28039dc9661b1e8a840ec4e97b2c7158d095c3f169475e3d4786d1519bb71"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 70224,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgen",
              "sha256": "98d0e0d581a3d878b68d63998d1b994dd69c33e0c4d6b798776516f5a4b546c3"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 52944,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgexec",
              "sha256": "84671f800b15a12e1c9c70f6a1421979b3d7f99812c51d773c6ef5e72075b009"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 54768,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgfilter",
              "sha256": "edabcf147e1b6bad16b7236005f7f9405f53caf9927a4ab39e8f4e7e19afa4e7"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 133328,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgfmt",
              "sha256": "d6d5d37e8557ef5b7051f44987c317de329b86f1fbfb6367c4fd1c683cb271f2"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 146736,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msggrep",
              "sha256": "0733ffcd3b185be78e1551b1d6ea1c63c5138294acafdc7bb800882bbf6b1f2d"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 94192,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msginit",
              "sha256": "27b055a9dc6ef470990eb7bd5c593fcab272c65b5e45b3675ed1c604f69395c4"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 75664,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgmerge",
              "sha256": "200a9d7d032d99617f33099dc2565929b8832a6fe9c073b33fc6babb3f5d2b58"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 54256,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgpre",
              "sha256": "4b779771ce2bc7e697e478498f9755b4d3111c69f90264622e93ecb527558a90"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 73696,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msgunfmt",
              "sha256": "2647c9ef6bde4989f1322d6ef7789350e39ddb9b1f1607607827838219dbf12c"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 53648,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/msguniq",
              "sha256": "35da51730aa86fcaa8da137819a14059d4fc38a5acbcbedd0ea33b7cc9c1584f"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 77520,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/ngettext",
              "sha256": "b2f2d701f536f0b7481a539fc98e76e9e1634d2268bfcca287710ab1dd2b1a76"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 131952,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/printf_gettext",
              "sha256": "2f08a3abc6e6bd799b242d837ff1cfd99f9952d5d41a34fb23b50eb1ea51cb79"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 132128,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/printf_ngettext",
              "sha256": "d0aaa3f8d417fee721cdfc4c05404c28ee5115a4c7e404a156a172d4c02848a6"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 35840,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/recode-sr-latin",
              "sha256": "f7c161a6978f58ba6716daedb677d899ff490e6885e71ac9d95b938b84656e10"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 71008,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/spit",
              "sha256": "740a4bda983551534511dcedb2eb2a9855dbb16b7352fc94987b9a7b3ba9a3b4"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 13751632,
              "path": "/opt/homebrew/Cellar/gettext/1.0/bin/xgettext",
              "sha256": "f1c1dd881de54317de2b0d8990023eba69b0c0a452d7c0387733bdbb1da394db"
            }
          },
          {
            "filetype": 6,
            "image": {
              "bytes": 139936,
              "path": "/opt/homebrew/Cellar/gettext/1.0/lib/libasprintf.0.dylib",
              "sha256": "c1cfe0e252ae64bd71bde5d0d77ff8fe5e3a88e0babd3194634d28720cfc91b3"
            }
          },
          {
            "filetype": 6,
            "image": {
              "bytes": 445952,
              "path": "/opt/homebrew/Cellar/gettext/1.0/lib/libgettextlib-1.0.dylib",
              "sha256": "a3f37cf55767df3722f666eb1922b0bbcfff9832c6bc6470a4d7423af6eb6886"
            }
          },
          {
            "filetype": 6,
            "image": {
              "bytes": 412304,
              "path": "/opt/homebrew/Cellar/gettext/1.0/lib/libgettextpo.0.dylib",
              "sha256": "792ca7a3d928bfa457f39342c11ae9ca2f1b7143de4ec46df3629622548c86b1"
            }
          },
          {
            "filetype": 6,
            "image": {
              "bytes": 361264,
              "path": "/opt/homebrew/Cellar/gettext/1.0/lib/libgettextsrc-1.0.dylib",
              "sha256": "8afef315fd512b2f2a729d075f8b464b7974be1c4b429714dacd0272f47d625f"
            }
          },
          {
            "filetype": 6,
            "image": {
              "bytes": 212080,
              "path": "/opt/homebrew/Cellar/gettext/1.0/lib/libintl.8.dylib",
              "sha256": "be8f2b23ec17bd294c7bd4fceaa00526c135b6110a55598b82b3aceeb533a601"
            }
          },
          {
            "filetype": 6,
            "image": {
              "bytes": 660208,
              "path": "/opt/homebrew/Cellar/gettext/1.0/lib/libtextstyle.0.dylib",
              "sha256": "bf37523629858b7d8b41504c295d1a489b5b9813a1dd26d531a1a05741ee7658"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 55184,
              "path": "/opt/homebrew/Cellar/gettext/1.0/libexec/gettext/cldr-plurals",
              "sha256": "0da549a06ccd23a2019f6ee619053f547dbfd0e25676b0e84df88320651ae3e6"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 35440,
              "path": "/opt/homebrew/Cellar/gettext/1.0/libexec/gettext/hostname",
              "sha256": "73f34891ebcd64e0c8e5c2d44f04df2c361396ebfbd3fcee71380ded1fbb1d62"
            }
          },
          {
            "filetype": 2,
            "image": {
              "bytes": 35696,
              "path": "/opt/homebrew/Cellar/gettext/1.0/libexec/gettext/urlget",
              "sha256": "60627bda647aa4a81529c8ee3790ad9a40426df74481562e462ad3e2f5289bb7"
            }
          }
        ],
        "non_mach_o_classes": [
          {
            "bytes": 449,
            "path": "/opt/homebrew/Cellar/gettext/1.0/share/doc/gettext/examples/hello-java-awt/m4/TestAWT.class",
            "sha256": "d7010787a901da8ac0d0dcc958aafda0ae32275f9a15e8a785bdee15e6579302"
          },
          {
            "bytes": 1125,
            "path": "/opt/homebrew/Cellar/gettext/1.0/share/doc/gettext/examples/hello-java-qtjambi/m4/Test15.class",
            "sha256": "7b628b26b9e4ea4a14597310fc2109ac9d9a7eb8f989ba65714eb763d8a5c3a6"
          },
          {
            "bytes": 449,
            "path": "/opt/homebrew/Cellar/gettext/1.0/share/doc/gettext/examples/hello-java-swing/m4/TestAWT.class",
            "sha256": "d7010787a901da8ac0d0dcc958aafda0ae32275f9a15e8a785bdee15e6579302"
          },
          {
            "bytes": 510,
            "path": "/opt/homebrew/Cellar/gettext/1.0/share/gettext/javaversion.class",
            "sha256": "2bee26b0b761a52a418a6488d94be8af824a30a3f87f50cb8d35f4b7a54a13a6"
          }
        ],
        "regular_files": 2493,
        "root": "/opt/homebrew/Cellar/gettext/1.0"
      }
    ]
  },
  "canonical_writes": false,
  "commands": [
    {
      "argv": [
        "/usr/bin/otool",
        "-L",
        "/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11"
      ],
      "cwd": "/Users/adamjackson/Projects/work/scout-x1",
      "ended_at": "2026-11-09T13:02:37.880741Z",
      "exit_code": 0,
      "started_at": "2026-11-09T13:02:37.785746Z",
      "stderr": "",
      "stdout": "/Users/adamjackson/Projects/work/scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11:\n\t/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation (compatibility version 150.0.0, current version 2202.0.0)\n\t@executable_path/../lib/libpython3.11.dylib (compatibility version 3.11.0, current version 3.11.0)\n\t/usr/lib/libSystem.B.dylib (compatibility version 1.0.0, current version 1336.61.1)\n"
    }
  ],
  "runtime_root": "/Users/adamjackson/Projects/work/scout-x1",
  "schema": "scout.x1.closure-review-dry-check.v1",
  "status": "passed"
}
```

stderr:
```text
```

Exit code: 0

### Full Python suite against final Core

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py' -v
```

stdout:
```text
```

stderr:
```text
test_both_revisions_resolved_before_any_scanner (test_binding.CompletionTests.test_both_revisions_resolved_before_any_scanner) ... ok
test_bound_policy_refuses_each_input_boundary (test_binding.CompletionTests.test_bound_policy_refuses_each_input_boundary) ... ok
test_calledprocesserror_always_terminates_refused_receipt (test_binding.CompletionTests.test_calledprocesserror_always_terminates_refused_receipt) ... ok
test_invalid_reference_head_resolution_prevents_every_scanner (test_binding.CompletionTests.test_invalid_reference_head_resolution_prevents_every_scanner) ... ok
test_resolved_policy_must_match_later_bytes_and_pin (test_binding.CompletionTests.test_resolved_policy_must_match_later_bytes_and_pin) ... ok
test_all_six_flags_derived_from_complete_evidence (test_binding.DeterministicProofTests.test_all_six_flags_derived_from_complete_evidence) ... ok
test_config_proof_and_actual_lanes_each_required (test_binding.DeterministicProofTests.test_config_proof_and_actual_lanes_each_required) ... ok
test_layout_and_identity_cannot_pass_vacuously (test_binding.DeterministicProofTests.test_layout_and_identity_cannot_pass_vacuously) ... ok
test_actual_frontend_launch_is_direct_and_private (test_binding.EnvironmentAndSeamTests.test_actual_frontend_launch_is_direct_and_private) ... ok
test_missing_attempt_log_refuses_instead_of_passing_zero (test_binding.EnvironmentAndSeamTests.test_missing_attempt_log_refuses_instead_of_passing_zero) ... ok
test_private_env_stubs_all_provider_paths_and_no_outer_sandbox (test_binding.EnvironmentAndSeamTests.test_private_env_stubs_all_provider_paths_and_no_outer_sandbox) ... ok
test_static_provider_binary_seams (test_binding.EnvironmentAndSeamTests.test_static_provider_binary_seams) ... ok
test_stub_is_logging_only (test_binding.EnvironmentAndSeamTests.test_stub_is_logging_only) ... ok
test_equal_length_roots_and_missing_head_state_directory (test_binding.FixtureAndParityTests.test_equal_length_roots_and_missing_head_state_directory) ... ok
test_exact_sets_health_multiplicity_coverage_and_durable_link (test_binding.FixtureAndParityTests.test_exact_sets_health_multiplicity_coverage_and_durable_link) ... ok
test_fixture_trigger_constructs_and_revision_difference (test_binding.FixtureAndParityTests.test_fixture_trigger_constructs_and_revision_difference) ... ok
test_path_escape_and_codegraph_refused (test_binding.FixtureAndParityTests.test_path_escape_and_codegraph_refused) ... ok
test_source_manifest_not_mistaken_for_filesystem_pin (test_binding.FixtureAndParityTests.test_source_manifest_not_mistaken_for_filesystem_pin) ... ok
test_actual_failing_stub_evidence (test_binding.ModelAttemptTests.test_actual_failing_stub_evidence) ... ok
test_positive_boundaries (test_binding.ModelAttemptTests.test_positive_boundaries) ... ok
test_zero_and_positive_logs (test_binding.ModelAttemptTests.test_zero_and_positive_logs) ... ok
test_canonical_gate_happens_before_lock_reservation_and_journal (test_binding.RehearsalTests.test_canonical_gate_happens_before_lock_reservation_and_journal) ... ok
test_rehearsal_main_never_opens_out_or_journal (test_binding.RehearsalTests.test_rehearsal_main_never_opens_out_or_journal) ... ok
test_rehearsal_rejects_canonical_escape_and_symlinks (test_binding.RehearsalTests.test_rehearsal_rejects_canonical_escape_and_symlinks) ... ok
test_reviewed_rehearsal_requires_current_pins_and_every_predicate (test_binding.RehearsalTests.test_reviewed_rehearsal_requires_current_pins_and_every_predicate) ... ok
test_wrong_or_missing_rehearsal_clock_refuses_before_scratch_writes (test_binding.RehearsalTests.test_wrong_or_missing_rehearsal_clock_refuses_before_scratch_writes) ... ok
test_adopted_sha_mismatch_refuses_load_and_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_adopted_sha_mismatch_refuses_load_and_binding) ... ok
test_adoption_review_mismatched_pin_refuses_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_adoption_review_mismatched_pin_refuses_binding) ... ok
test_loaded_configuration_requires_exact_adopted_row (test_binding_adoption.BindingAdoptionIntegrationTests.test_loaded_configuration_requires_exact_adopted_row) ... ok
test_one_source_pin_missing_from_amendment_review_refuses_binding (test_binding_adoption.BindingAdoptionIntegrationTests.test_one_source_pin_missing_from_amendment_review_refuses_binding) ... ok
test_proposal_digest_missing_refuses_real_adoption (test_binding_adoption.BindingAdoptionIntegrationTests.test_proposal_digest_missing_refuses_real_adoption) ... ok
test_proposal_path_missing_refuses_binding_even_with_reviewed_digest (test_binding_adoption.BindingAdoptionIntegrationTests.test_proposal_path_missing_refuses_binding_even_with_reviewed_digest) ... ok
test_real_adopt_load_and_entire_binding_authority_chain (test_binding_adoption.BindingAdoptionIntegrationTests.test_real_adopt_load_and_entire_binding_authority_chain) ... ok
test_every_launch_checks_key_presence_before_direct_execution (test_binding_boundaries.ControllerKeyBoundaryTests.test_every_launch_checks_key_presence_before_direct_execution) ... ok
test_rehearsal_entry_checks_presence_before_output_reservation (test_binding_boundaries.ControllerKeyBoundaryTests.test_rehearsal_entry_checks_presence_before_output_reservation) ... ok
test_reviewed_inputs_checks_presence_not_truthiness (test_binding_boundaries.ControllerKeyBoundaryTests.test_reviewed_inputs_checks_presence_not_truthiness) ... ok
test_django_evidence_pins_exact_synthetic_receipt_and_detects_change (test_binding_boundaries.DjangoFixtureBoundaryTests.test_django_evidence_pins_exact_synthetic_receipt_and_detects_change) ... ok
test_django_manifest_refusal_table (test_binding_boundaries.DjangoFixtureBoundaryTests.test_django_manifest_refusal_table) ... ok
test_equal_synthetic_parent_head_issue_sets_refuse (test_binding_boundaries.DjangoFixtureBoundaryTests.test_equal_synthetic_parent_head_issue_sets_refuse) ... ok
test_each_bound_reference_attempt_count_must_be_zero (test_binding_boundaries.PredicateBoundaryTests.test_each_bound_reference_attempt_count_must_be_zero) ... ok
test_positive_control_false_boundaries (test_binding_boundaries.PredicateBoundaryTests.test_positive_control_false_boundaries) ... ok
test_resolve_attempt_and_exit_status_boundaries (test_binding_boundaries.PredicateBoundaryTests.test_resolve_attempt_and_exit_status_boundaries) ... ok
test_codegraph_refused_even_via_link (test_closure.ClosureTests.test_codegraph_refused_even_via_link) ... ok
test_cycle_refused (test_closure.ClosureTests.test_cycle_refused) ... ok
test_explicit_external_root_counted (test_closure.ClosureTests.test_explicit_external_root_counted) ... ok
test_incomplete_retained_object_refused (test_closure.ClosureTests.test_incomplete_retained_object_refused) ... ok
test_inventory_preserves_links_and_files (test_closure.ClosureTests.test_inventory_preserves_links_and_files) ... ok
test_json_stream_rejects_partial (test_closure.ClosureTests.test_json_stream_rejects_partial) ... ok
test_launcher_bytes_never_normalized (test_closure.ClosureTests.test_launcher_bytes_never_normalized) ... ok
test_link_chain (test_closure.ClosureTests.test_link_chain) ... ok
test_missing_node_dependency_refused (test_closure.ClosureTests.test_missing_node_dependency_refused) ... ok
test_model_cli_never_launched (test_closure.ClosureTests.test_model_cli_never_launched) ... ok
test_native_missing_dependency_refused (test_closure.ClosureTests.test_native_missing_dependency_refused) ... ok
test_native_self_identity_not_dependency (test_closure.ClosureTests.test_native_self_identity_not_dependency) ... ok
test_npm_identity_not_just_version (test_closure.ClosureTests.test_npm_identity_not_just_version) ... ok
test_python_record_verified (test_closure.ClosureTests.test_python_record_verified) ... ok
test_python_tamper_refused (test_closure.ClosureTests.test_python_tamper_refused) ... ok
test_requirements_frozen_hashes (test_closure.ClosureTests.test_requirements_frozen_hashes) ... ok
test_requirements_unhashed_refused (test_closure.ClosureTests.test_requirements_unhashed_refused) ... ok
test_retention_deduplicates_and_preserves_original_mode (test_closure.ClosureTests.test_retention_deduplicates_and_preserves_original_mode) ... ok
test_unknown_external_link_refused (test_closure.ClosureTests.test_unknown_external_link_refused) ... ok
test_r101_duplicate_applicable_row (test_closure_review.ClosureReviewTests.test_r101_duplicate_applicable_row) ... ok
test_r101_real_marker_lines (test_closure_review.ClosureReviewTests.test_r101_real_marker_lines) ... ok
test_r101_unknown_marker_variable_and_operator (test_closure_review.ClosureReviewTests.test_r101_unknown_marker_variable_and_operator) ... ok
test_r102_executable_context_carried_through_rpath (test_closure_review.ClosureReviewTests.test_r102_executable_context_carried_through_rpath) ... ok
test_r102_missing_executable_context_refuses (test_closure_review.ClosureReviewTests.test_r102_missing_executable_context_refuses) ... ok
test_r103_nested_peer_keys_are_not_packages (test_closure_review.ClosureReviewTests.test_r103_nested_peer_keys_are_not_packages) ... ok
test_r104_internal_tarball_node_modules_pass (test_closure_review.ClosureReviewTests.test_r104_internal_tarball_node_modules_pass) ... ok
test_r104_stray_nested_dependency_refuses (test_closure_review.ClosureReviewTests.test_r104_stray_nested_dependency_refuses) ... ok
test_r104_stray_top_level_package_refuses (test_closure_review.ClosureReviewTests.test_r104_stray_top_level_package_refuses) ... ok
test_r105_consistently_substituted_installed_locks_refuse (test_closure_review.ClosureReviewTests.test_r105_consistently_substituted_installed_locks_refuse) ... ok
test_r106_git_agreement_and_helper_root (test_closure_review.ClosureReviewTests.test_r106_git_agreement_and_helper_root) ... ok
test_r106_git_disagreement_refuses (test_closure_review.ClosureReviewTests.test_r106_git_disagreement_refuses) ... ok
test_r107_reviewed_frontend_bytes_must_stay_signed (test_closure_review.ClosureReviewTests.test_r107_reviewed_frontend_bytes_must_stay_signed) ... ok
test_r107_unpinned_selected_source_refuses (test_closure_review.ClosureReviewTests.test_r107_unpinned_selected_source_refuses) ... ok
test_r108_own_closure_is_not_whole_execution_manifest (test_closure_review.ClosureReviewTests.test_r108_own_closure_is_not_whole_execution_manifest) ... ok
test_r109_interpreter_outside_inventory_refuses (test_closure_review.ClosureReviewTests.test_r109_interpreter_outside_inventory_refuses) ... ok
test_r109_relocated_runtime_refuses (test_closure_review.ClosureReviewTests.test_r109_relocated_runtime_refuses) ... ok
test_r109_resolved_path_must_equal_signed_path (test_closure_review.ClosureReviewTests.test_r109_resolved_path_must_equal_signed_path) ... ok
test_r109_unisolated_interpreter_refuses (test_closure_review.ClosureReviewTests.test_r109_unisolated_interpreter_refuses) ... ok
test_r109_wrong_running_interpreter_refuses (test_closure_review.ClosureReviewTests.test_r109_wrong_running_interpreter_refuses) ... ok
test_r110_nonexecutable_node_addon_is_inspected (test_closure_review.ClosureReviewTests.test_r110_nonexecutable_node_addon_is_inspected) ... ok
test_r111_deferred_count_refuses (test_closure_review.ClosureReviewTests.test_r111_deferred_count_refuses) ... ok
test_r111_deferred_empty_still_verifies_working_files (test_closure_review.ClosureReviewTests.test_r111_deferred_empty_still_verifies_working_files) ... ok
test_r111_deferred_field_refuses (test_closure_review.ClosureReviewTests.test_r111_deferred_field_refuses) ... ok
test_r111_deferred_nonempty_null_pass (test_closure_review.ClosureReviewTests.test_r111_deferred_nonempty_null_pass) ... ok
test_r111_deferred_set_refuses (test_closure_review.ClosureReviewTests.test_r111_deferred_set_refuses) ... ok
test_r111_go_metadata_planted_null_pass (test_closure_review.ClosureReviewTests.test_r111_go_metadata_planted_null_pass) ... ok
test_r111_go_replace_refuses (test_closure_review.ClosureReviewTests.test_r111_go_replace_refuses) ... ok
test_r111_go_sum_mismatch_refuses (test_closure_review.ClosureReviewTests.test_r111_go_sum_mismatch_refuses) ... ok
test_r111_recheck_added_file_refuses (test_closure_review.ClosureReviewTests.test_r111_recheck_added_file_refuses) ... ok
test_r111_recheck_entrypoint_resolution_drift_refuses (test_closure_review.ClosureReviewTests.test_r111_recheck_entrypoint_resolution_drift_refuses) ... ok
test_r111_recheck_mode_change_refuses (test_closure_review.ClosureReviewTests.test_r111_recheck_mode_change_refuses) ... ok
test_r111_recheck_planted_null_pass (test_closure_review.ClosureReviewTests.test_r111_recheck_planted_null_pass) ... ok
test_r111_recheck_retargeted_link_refuses (test_closure_review.ClosureReviewTests.test_r111_recheck_retargeted_link_refuses) ... ok
test_r111_recheck_writable_retained_object_refuses (test_closure_review.ClosureReviewTests.test_r111_recheck_writable_retained_object_refuses) ... ok
test_r201_fat32_uses_arm64_slice (test_closure_round2.RoundTwoTests.test_r201_fat32_uses_arm64_slice) ... ok
test_r201_fat64_uses_arm64_slice (test_closure_round2.RoundTwoTests.test_r201_fat64_uses_arm64_slice) ... ok
test_r201_java_is_non_native_without_suffix_filter (test_closure_round2.RoundTwoTests.test_r201_java_is_non_native_without_suffix_filter) ... ok
test_r201_misaligned_fat_refuses (test_closure_round2.RoundTwoTests.test_r201_misaligned_fat_refuses) ... ok
test_r201_out_of_bounds_fat_refuses (test_closure_round2.RoundTwoTests.test_r201_out_of_bounds_fat_refuses) ... ok
test_r201_thin_arm64 (test_closure_round2.RoundTwoTests.test_r201_thin_arm64) ... ok
test_r201_truncated_fat_refuses (test_closure_round2.RoundTwoTests.test_r201_truncated_fat_refuses) ... ok
test_r201_truncated_thin_refuses (test_closure_round2.RoundTwoTests.test_r201_truncated_thin_refuses) ... ok
test_r202_go_metadata_escaping_file_symlink_refuses (test_closure_round2.RoundTwoTests.test_r202_go_metadata_escaping_file_symlink_refuses) ... ok
test_r202_go_metadata_logical_source_null (test_closure_round2.RoundTwoTests.test_r202_go_metadata_logical_source_null) ... ok
test_r202_go_metadata_unpinned_inside_root_refuses (test_closure_round2.RoundTwoTests.test_r202_go_metadata_unpinned_inside_root_refuses) ... ok
test_r203_drifted_council_record_refuses (test_closure_round2.RoundTwoTests.test_r203_drifted_council_record_refuses) ... ok
test_r203_drifted_receipt_refuses (test_closure_round2.RoundTwoTests.test_r203_drifted_receipt_refuses) ... ok
test_r203_exact_selected_frontend_review_set (test_closure_round2.RoundTwoTests.test_r203_exact_selected_frontend_review_set) ... ok
test_r203_main_journal_binds_dedicated_receipt_and_record (test_closure_round2.RoundTwoTests.test_r203_main_journal_binds_dedicated_receipt_and_record) ... ok
test_r203_missing_dedicated_receipt_refuses (test_closure_round2.RoundTwoTests.test_r203_missing_dedicated_receipt_refuses) ... ok
test_r203_unrelated_reviewed_byte_refuses (test_closure_round2.RoundTwoTests.test_r203_unrelated_reviewed_byte_refuses) ... ok
test_r203_valid_receipt_and_council_record_are_pinned (test_closure_round2.RoundTwoTests.test_r203_valid_receipt_and_council_record_are_pinned) ... ok
test_r203_wrong_scope_refuses (test_closure_round2.RoundTwoTests.test_r203_wrong_scope_refuses) ... ok
test_r204_attest_producer_wires_helpers_and_review_inputs (test_closure_round2.RoundTwoTests.test_r204_attest_producer_wires_helpers_and_review_inputs)
Fake only external metadata; execute the real attestation orchestration. ... ok
test_r204_claw_assets_precede_pnpm_null (test_closure_round2.RoundTwoTests.test_r204_claw_assets_precede_pnpm_null) ... ok
test_r204_git_head_mismatch_refuses (test_closure_round2.RoundTwoTests.test_r204_git_head_mismatch_refuses) ... ok
test_r204_git_non_bin_refuses (test_closure_round2.RoundTwoTests.test_r204_git_non_bin_refuses) ... ok
test_r204_git_shim_refuses (test_closure_round2.RoundTwoTests.test_r204_git_shim_refuses) ... ok
test_r204_git_version_drift_refuses (test_closure_round2.RoundTwoTests.test_r204_git_version_drift_refuses) ... ok
test_r204_keg_git_version_head_null_and_command_identity (test_closure_round2.RoundTwoTests.test_r204_keg_git_version_head_null_and_command_identity) ... ok
test_r204_missing_git_core_refuses (test_closure_round2.RoundTwoTests.test_r204_missing_git_core_refuses) ... ok
test_r204_native_inventory_selection_addon_failure_caught (test_closure_round2.RoundTwoTests.test_r204_native_inventory_selection_addon_failure_caught) ... ok
test_r204_native_inventory_selection_null (test_closure_round2.RoundTwoTests.test_r204_native_inventory_selection_null) ... ok
test_r204_nonpassing_object_receipt_refuses (test_closure_round2.RoundTwoTests.test_r204_nonpassing_object_receipt_refuses) ... ok
test_r204_runtime_inventory_includes_keg_null (test_closure_round2.RoundTwoTests.test_r204_runtime_inventory_includes_keg_null) ... ok
test_r204_runtime_inventory_keg_escape_refuses (test_closure_round2.RoundTwoTests.test_r204_runtime_inventory_keg_escape_refuses) ... ok
test_r204_substituted_claw_refuses_before_pnpm (test_closure_round2.RoundTwoTests.test_r204_substituted_claw_refuses_before_pnpm) ... ok
test_r301_45_slice_native_collision_never_skipped (test_closure_round3.RoundThreeTests.test_r301_45_slice_native_collision_never_skipped) ... ok
test_r301_bounded_fat_refuses_even_if_stricter_checks_would_fail (test_closure_round3.RoundThreeTests.test_r301_bounded_fat_refuses_even_if_stricter_checks_would_fail) ... ok
test_r301_java_52_0_and_45_3_remain_non_native (test_closure_round3.RoundThreeTests.test_r301_java_52_0_and_45_3_remain_non_native) ... ok
test_r301_java_only_skipped_when_a_fat_slice_escapes_file (test_closure_round3.RoundThreeTests.test_r301_java_only_skipped_when_a_fat_slice_escapes_file) ... ok
test_s4_alignment_exponent_and_table_overlap (test_closure_round3.RoundThreeTests.test_s4_alignment_exponent_and_table_overlap) ... ok
test_s4_cpu_disagreement (test_closure_round3.RoundThreeTests.test_s4_cpu_disagreement) ... ok
test_s4_empty_slice (test_closure_round3.RoundThreeTests.test_s4_empty_slice) ... ok
test_s4_fat64_reserved_field (test_closure_round3.RoundThreeTests.test_s4_fat64_reserved_field) ... ok
test_s4_invalid_architecture_counts (test_closure_round3.RoundThreeTests.test_s4_invalid_architecture_counts) ... ok
test_s4_invalid_slice_magic (test_closure_round3.RoundThreeTests.test_s4_invalid_slice_magic) ... ok
test_s4_invalid_thin_filetype_and_command_bounds (test_closure_round3.RoundThreeTests.test_s4_invalid_thin_filetype_and_command_bounds) ... ok
test_s4_missing_or_duplicate_arm64 (test_closure_round3.RoundThreeTests.test_s4_missing_or_duplicate_arm64) ... ok
test_s4_non_native_null (test_closure_round3.RoundThreeTests.test_s4_non_native_null) ... ok
test_s4_overlapping_slices (test_closure_round3.RoundThreeTests.test_s4_overlapping_slices) ... ok
test_s4_truncated_fat_version (test_closure_round3.RoundThreeTests.test_s4_truncated_fat_version) ... ok
test_s4_x86_and_arm64e_null_and_swapped_fat (test_closure_round3.RoundThreeTests.test_s4_x86_and_arm64e_null_and_swapped_fat) ... ok
test_adopted_amendment_without_review_refused_at_load (test_core.CoreTests.test_adopted_amendment_without_review_refused_at_load) ... ok
test_adoption_racing_journal_is_serialized (test_core.CoreTests.test_adoption_racing_journal_is_serialized) ... ok
test_amendment_chain_applies_in_order_and_preserves_original (test_core.CoreTests.test_amendment_chain_applies_in_order_and_preserves_original) ... ok
test_amendment_review_scope_and_exact_digest_required (test_core.CoreTests.test_amendment_review_scope_and_exact_digest_required) ... ok
test_clock_pinned_original_config_cannot_drift (test_core.CoreTests.test_clock_pinned_original_config_cannot_drift) ... ok
test_deadline_exact_seconds_and_leap_boundary (test_core.CoreTests.test_deadline_exact_seconds_and_leap_boundary) ... ok
test_edit_delete_or_rename_adopted_amendment_refused (test_core.CoreTests.test_edit_delete_or_rename_adopted_amendment_refused) ... ok
test_exception_classification (test_core.CoreTests.test_exception_classification) ... ok
test_exclusive_creation_preserves_original (test_core.CoreTests.test_exclusive_creation_preserves_original) ... ok
test_hand_written_and_backdated_adoption_receipts_refused (test_core.CoreTests.test_hand_written_and_backdated_adoption_receipts_refused) ... ok
test_hd9_only_state_date_normalized (test_core.CoreTests.test_hd9_only_state_date_normalized) ... ok
test_initialized_submodule_checks_working_bytes (test_core.CoreTests.test_initialized_submodule_checks_working_bytes) ... ok
test_initialized_submodule_mismatch_refused (test_core.CoreTests.test_initialized_submodule_mismatch_refused) ... ok
test_load_and_seal_accept_legacy_ab_and_d_before_adoption (test_core.CoreTests.test_load_and_seal_accept_legacy_ab_and_d_before_adoption) ... ok
test_merge_first_parent_and_offpath_refusal (test_core.CoreTests.test_merge_first_parent_and_offpath_refusal) ... ok
test_original_config_is_unchanged (test_core.CoreTests.test_original_config_is_unchanged) ... ok
test_promisor_and_extra_ref_refused (test_core.CoreTests.test_promisor_and_extra_ref_refused) ... ok
test_proposed_amendment_is_ignored (test_core.CoreTests.test_proposed_amendment_is_ignored) ... ok
test_recursive_pin_mismatch_and_missing (test_core.CoreTests.test_recursive_pin_mismatch_and_missing) ... ok
test_seal_deferred_rows_must_all_be_verified (test_core.CoreTests.test_seal_deferred_rows_must_all_be_verified) ... ok
test_seal_journal_validator_accepts_pre_and_post_adoption (test_core.CoreTests.test_seal_journal_validator_accepts_pre_and_post_adoption) ... ok
test_seal_missing_and_mismatched_canonical_record (test_core.CoreTests.test_seal_missing_and_mismatched_canonical_record) ... ok
test_seal_missing_records_reservation_not_replaced (test_core.CoreTests.test_seal_missing_records_reservation_not_replaced) ... ok
test_start_binds_uninitialized_submodule_receipt_without_hashing_deferred_rows (test_core.CoreTests.test_start_binds_uninitialized_submodule_receipt_without_hashing_deferred_rows) ... ok
test_uninitialized_submodule_deferred_not_passed (test_core.CoreTests.test_uninitialized_submodule_deferred_not_passed) ... ok
test_unreviewed_amendment_refused_at_seal (test_core.CoreTests.test_unreviewed_amendment_refused_at_seal) ... ok
test_unsafe_existing_repository_refused (test_core.CoreTests.test_unsafe_existing_repository_refused) ... ok
test_wrong_prior_digest_refused (test_core.CoreTests.test_wrong_prior_digest_refused) ... ok
test_context_identity_refuses (test_identifiers.PipelineTests.test_context_identity_refuses) ... ok
test_coverage_and_enumeration_identity_refuse (test_identifiers.PipelineTests.test_coverage_and_enumeration_identity_refuse) ... ok
test_dropped_record_refuses (test_identifiers.PipelineTests.test_dropped_record_refuses) ... ok
test_existing_root_refuses_and_preserves_bytes (test_identifiers.PipelineTests.test_existing_root_refuses_and_preserves_bytes) ... ok
test_frozen_pin_duplicate_and_count_refusals (test_identifiers.PipelineTests.test_frozen_pin_duplicate_and_count_refusals) ... ok
test_frozen_valid_eight_file_fixture (test_identifiers.PipelineTests.test_frozen_valid_eight_file_fixture) ... ok
test_group_conservation_pin_once_and_source_table (test_identifiers.PipelineTests.test_group_conservation_pin_once_and_source_table) ... ok
test_independent_sources_digest_and_canonical_input_refusals (test_identifiers.PipelineTests.test_independent_sources_digest_and_canonical_input_refusals) ... ok
test_listing_mismatch_and_duplicates_refuse (test_identifiers.PipelineTests.test_listing_mismatch_and_duplicates_refuse) ... ok
test_metadata_delta_and_duplicate (test_identifiers.PipelineTests.test_metadata_delta_and_duplicate) ... ok
test_missing_and_ambiguous_source_lookup (test_identifiers.PipelineTests.test_missing_and_ambiguous_source_lookup) ... ok
test_nonparsed_statuses_all_unresolved (test_identifiers.PipelineTests.test_nonparsed_statuses_all_unresolved) ... ok
test_partial_publication_retained (test_identifiers.PipelineTests.test_partial_publication_retained) ... ok
test_planted_emitted_range_loss_refuses (test_identifiers.PipelineTests.test_planted_emitted_range_loss_refuses) ... ok
test_required_header_and_table_mutants_fail_oracle (test_identifiers.PipelineTests.test_required_header_and_table_mutants_fail_oracle) ... ok
test_runtime_noncanonical_signed_and_source_count_refuse (test_identifiers.PipelineTests.test_runtime_noncanonical_signed_and_source_count_refuse) ... ok
test_seal_identity_and_crosslinks_refuse (test_identifiers.PipelineTests.test_seal_identity_and_crosslinks_refuse) ... ok
test_successful_publication_groups_and_pin_once (test_identifiers.PipelineTests.test_successful_publication_groups_and_pin_once) ... ok
test_summary_count_refusals (test_identifiers.PipelineTests.test_summary_count_refusals) ... ok
test_aliases_and_case (test_identifiers.ProjectionTests.test_aliases_and_case) ... ok
test_any_scope_sha_and_repository_scoped_path_pr (test_identifiers.ProjectionTests.test_any_scope_sha_and_repository_scoped_path_pr) ... ok
test_rejection_boundaries (test_identifiers.ProjectionTests.test_rejection_boundaries) ... ok
test_scoped_paths_prs_and_familiarity (test_identifiers.ProjectionTests.test_scoped_paths_prs_and_familiarity) ... ok
test_unattributed_boundaries (test_identifiers.ProjectionTests.test_unattributed_boundaries) ... ok
test_unknown_top_level_and_duplicate_json (test_identifiers.ProjectionTests.test_unknown_top_level_and_duplicate_json) ... ok

----------------------------------------------------------------------
Ran 198 tests in 14.738s

OK
```

Exit code: 0

### Offline Go vet

```sh
/bin/bash -c 'source .tools/activate && GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 go vet ./evaluation/x1-triage-ranking/execution/...'
```

stdout:
```text
```

stderr:
```text
```

Exit code: 0

### Offline Go test

```sh
/bin/bash -c 'source .tools/activate && GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 go test ./evaluation/x1-triage-ranking/execution/...'
```

stdout:
```text
ok  	scout/evaluation/x1-triage-ranking/execution/model_free	(cached)
```

stderr:
```text
```

Exit code: 0

### Final source SHA-256s

```sh
/usr/bin/shasum -a 256 evaluation/x1-triage-ranking/execution/attest_closure.py evaluation/x1-triage-ranking/execution/closure/native.py evaluation/x1-triage-ranking/execution/tests/test_closure.py evaluation/x1-triage-ranking/execution/tests/test_closure_review.py evaluation/x1-triage-ranking/execution/tests/test_closure_round2.py evaluation/x1-triage-ranking/execution/tests/test_closure_round3.py evaluation/x1-triage-ranking/execution/preparation.py evaluation/x1-triage-ranking/execution/verify_inputs.py evaluation/x1-triage-ranking/execution/run-config.json
```

stdout:
```text
113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892  evaluation/x1-triage-ranking/execution/attest_closure.py
c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb  evaluation/x1-triage-ranking/execution/closure/native.py
3a1eae6082d0a9f42074e7ea5db8a2d74720808a91c2e2d5770bf44a82e599dd  evaluation/x1-triage-ranking/execution/tests/test_closure.py
8762b45d3e1db762b9094dfb174ec79c65ae5efa262a2903a2f3894bf8804743  evaluation/x1-triage-ranking/execution/tests/test_closure_review.py
7b22a59cbaf43cb0c97381d8ee6faf7fcc6a180c05b3d081ed88f04a4b99cc0c  evaluation/x1-triage-ranking/execution/tests/test_closure_round2.py
f6eec06e95a7227a02ffcd7ded9598957cffe2f840fe7976bde7e9e1e9e599e8  evaluation/x1-triage-ranking/execution/tests/test_closure_round3.py
599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427  evaluation/x1-triage-ranking/execution/preparation.py
7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310  evaluation/x1-triage-ranking/execution/verify_inputs.py
c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54  evaluation/x1-triage-ranking/execution/run-config.json
```

stderr:
```text
```

Exit code: 0
