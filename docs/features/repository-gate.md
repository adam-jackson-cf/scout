# Repository gate

## Sub-features

- **Whole-repository static checks.** Commit checks report invalid unstaged tracked sources and untracked new sources, including Go formatting, maintained Python lint and contributor Markdown formatting.
- **Empty-index checks and test selection.** Static checks and every fast-tier package run even when nothing is staged. A failing fast test blocks the gate; a conforming repository succeeds after running both fixture packages. Affected seam packages run at commit; unknown ownership selects all seam packages. Both commit and CI run every `scripts/gate/test_*.py` boundary suite.
- **Staged-byte protection.** Partially staged sources are checked as committed without changing working-tree edits. Deleted sources are not formatter inputs, and an alternate Git index retains its staged bytes.
- **Source containment.** Tracked and untracked source symlinks targeting files outside the repository are rejected, even when the target exists and is otherwise valid source. The external pinned tool environment is not a source input.
- **Maintained Python typing.** Commit checks reject type errors in both `scripts/gate` and `internal/association/testdata/gen`. Normal mypy import following checks an imported maintained dependency even when only its caller is an explicit target.
- **Dependency direction.** Unmapped Go dependency targets and forbidden Python gate-module edges are rejected with the source file, edge and repair. Proof: [test_structure_naming.py](../../scripts/gate/test_structure_naming.py).
- **Folder placement.** Unmapped locations and grab-bag folder names are rejected with a concrete owner destination. Proof: [test_structure_naming.py](../../scripts/gate/test_structure_naming.py).
- **Path and symbol naming.** Changed paths and Python symbols follow declared conventions; diagnostics give the corrected name. Proof: [test_structure_naming.py](../../scripts/gate/test_structure_naming.py).
- **Test-file names.** Tests use their owner's naming pattern without ticket, version or date tokens. Proof: [test_structure_naming.py](../../scripts/gate/test_structure_naming.py).
- **Test ownership and layout.** Every test has a valid owner and tier; fixtures and shared support stay in their declared layout. Findings name the file and destination or tier-policy repair. Proof: [test_structure_naming.py](../../scripts/gate/test_structure_naming.py).
- **Changed-path diagnostics.** The `path` command checks structure, naming and test placement, printing one actionable finding per line. Conforming paths and excluded frozen history remain accepted. Proof: [test_structure_naming.py](../../scripts/gate/test_structure_naming.py).
- **Python ownership and naming (PYQ100–107).** Maintained roots, approved kebab-case commands, canonical test groups, restricted name parts, snake-case packages/modules and kebab-case skill names have refusal canaries. Proof: [test_python_policy.py](../../scripts/gate/test_python_policy.py).
- **Python class and collaborator contracts (PYQ200–203, PYQ210–211, PYQ220–221, PYQ230–231).** Canaries cover the 700-line limit, configured public-method cap, eight constructor parameters, twelve attributes, approved analyzer bases and entrypoints, process wrappers, dynamic-import adapters, named callable aliases and concrete collaborator types. Every refusal must identify the file, rule and repair. Proof: [test_python_policy.py](../../scripts/gate/test_python_policy.py).
- **Python scalability (PYS250–253).** Canaries cover loop/comprehension nesting above two, analyzer filesystem/process/environment effects outside approved seams, and missing ownership markers. Proof: [test_python_policy.py](../../scripts/gate/test_python_policy.py).
- **Python architectural fanout.** Distinct non-standard-library imports are capped at ten, or three above 800 non-comment lines; boundary controls preserve standard-library and duplicate-import handling. Proof: [test_python_policy.py](../../scripts/gate/test_python_policy.py).
- **Python typing policy.** New unchecked maintained packages, package-list/import-root drift, global mypy exceptions and unjustified module overrides have independent refusal canaries. Proof: [test_python_policy.py](../../scripts/gate/test_python_policy.py).
- **Ruff suppression approval.** New inline `noqa` and `per-file-ignores` entries require approval rather than silently weakening checks. Proof: [test_python_policy.py](../../scripts/gate/test_python_policy.py).

The Python policy checks run over the whole maintained inventory at commit and in CI; conforming boundary controls remain accepted.

Proved by [test_inventory.py](../../scripts/gate/test_inventory.py): temporary Git repositories exercise the real gate and pinned checkers, with conforming, deletion, partial-staging and alternate-index acceptance controls alongside the refusal canaries.

## How to get to it (user POV)

1. Install the pinned toolchain and native hook as described in [CONTRIBUTING.md](../../CONTRIBUTING.md).
2. Edit the repository normally, stage the intended commit, and run `bash scripts/gate.sh commit` (also invoked by the native pre-commit hook).
3. Repair findings before committing. CI runs the same entry point in `ci` mode.

## Driving it with `bash scripts/gate.sh`

```bash
bash scripts/gate.sh commit
bash scripts/gate.sh ci
bash scripts/gate.sh path scripts/gate/run.py internal/domain/types_test.go
```

Run the repository-gate boundary canaries independently, without test-result caching:

```bash
source .tools/activate
.tools/venv/bin/python -B scripts/gate/test_inventory.py
.tools/venv/bin/python -B scripts/gate/test_structure_naming.py
.tools/venv/bin/python -B scripts/gate/test_python_policy.py
```

Each test creates and removes its own fixture repository. Subprocesses have bounded timeouts; the fixtures disable Go dependency and toolchain downloads and use the installed tools, not network services.

## Gotchas

- Static checks and tests run with an empty index too; staging no files does not skip validation.
- These are process-boundary tests, not fast-tier in-process tests. Both commit and CI select every `scripts/gate/test_*.py` suite.
- Fixtures copy the actual production checker modules and configuration. They use a tiny dependency-free Go module and two declared fast packages rather than running Scout's product suite recursively.
- Missing tools or prerequisite failures are setup failures, not evidence that a rule rejected its intended violation. See [CONTRIBUTING.md](../../CONTRIBUTING.md) for bootstrap commands.
- Frozen studies, third-party sources and generated tool environments are not maintained source inputs. These tests never execute studies or dependency audits.
