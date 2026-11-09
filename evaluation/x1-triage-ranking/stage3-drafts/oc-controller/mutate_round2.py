"""Scratch-only reversion probes: original and signed files are read-only."""
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

HERE = Path(__file__).resolve().parent
EXPERIMENT = HERE.parents[1]
MUTANTS = [
    ("kernel digest", "    if digest(path.read_bytes()) != KERNEL_SHA256:\n        raise RuntimeError(\"signed population identity source changed\")\n", "", "test_source_order_policy_refusals_before_freeze"),
    ("order pin", "digest(canonical(self._ordered)) != ORDER_SHA256 or ", "", "test_source_order_policy_refusals_before_freeze"),
    ("policy pin", " or digest(canonical(POLICY)) != POLICY_SHA256", "", "test_source_order_policy_refusals_before_freeze"),
    ("criticality completion", "        if status == \"measured\" and len(self._active[\"criticality\"]) != 4:\n            raise RuntimeError(\"all criticality phases required\")\n", "", "test_private_readonly_state_and_completion_guard"),
    ("nonbool gate", "                    if type(passes) is not bool:\n                        raise ValueError(\"gate must report a bool\")\n", "", "test_nonbool_gate_is_latched_not_gate_failure"),
    ("wall origin", "(self._clock() - self._oc_started_at) / 86400", "(self._clock() - self._freeze[\"created_at\"]) / 86400", "test_meter_origin_decrease_and_fault_latch_without_meter_read"),
    ("decreasing meter", " or core < self._last_core or wall < self._last_wall", "", "test_meter_origin_decrease_and_fault_latch_without_meter_read"),
    ("Wilson variance", "p*(1-p)/n+", "", "test_independent_known_answers_and_canonical_seed_refusals"),
    ("canonical seed", "        if canonical(json.loads(cell)) != cell_bytes:\n            raise ValueError(\"cell string is not canonical JSON\")\n", "", "test_independent_known_answers_and_canonical_seed_refusals"),
    ("R3-01 poison latch", '        raw = canonical(event) + b"\\n"\n        self._poisoned = True\n', '        raw = canonical(event) + b"\\n"\n        self._poisoned = False\n', "test_journal_failure_poison_blocks_other_lane_and_replays"),
    ("R3-01 poisoned writer guard", '        if self._poisoned:\n            raise JournalPoisoned("journal writer poisoned; restart and replay required")\n', "", "test_journal_failure_poison_blocks_other_lane_and_replays"),
    ("R3-03 binding admission", 'self._started_safety or self._run_lock.locked() or ', 'self._started_safety or ', "test_paused_run_atomically_closes_binding_admission"),
    ("R3-05 gate classification", '                    if self._terminated or self._paused or epoch != self._epoch:\n                        raise Stopped("gate result remains unclassified after stop/pause")\n', "", "test_gate_window_classifications_are_isolated"),
    ("R3-05 gate BaseException", '                        raise ValueError("gate must report a bool")\n                except BaseException as error:', '                        raise ValueError("gate must report a bool")\n                except Exception as error:', "test_gate_baseexception_is_latched"),
    ("R3-05 calibration BaseException", '                record["record_sha256"] = digest(canonical(record))\n            except BaseException as error:', '                record["record_sha256"] = digest(canonical(record))\n            except Exception as error:', "test_calibration_baseexception_is_latched"),
    ("R3-05 safety handle", '                if p["safety"] and not self._run_lock.locked():\n                    raise RuntimeError("safety work belongs to run loop")\n', "", "test_safety_handle_guard"),
]


def main():
    original = (HERE / "controller.py").read_text()
    killed = 0
    for name, old, new, test in MUTANTS:
        if original.count(old) != 1:
            raise RuntimeError(f"nonunique mutation: {name}")
        with tempfile.TemporaryDirectory(prefix="x1g-controller-mutant-") as directory:
            experiment = Path(directory) / "evaluation/x1-triage-ranking"
            target = experiment / "stage3-drafts/oc-controller"
            target.mkdir(parents=True)
            (experiment / "validation").mkdir()
            shutil.copy2(EXPERIMENT / "oc-manifest.json", experiment / "oc-manifest.json")
            shutil.copy2(EXPERIMENT / "validation/fixed_population.py", experiment / "validation/fixed_population.py")
            shutil.copy2(EXPERIMENT / "validation/ranking.py", experiment / "validation/ranking.py")
            for filename in ("test_controller.py", "known_answers.json", "policy.json"):
                shutil.copy2(HERE / filename, target / filename)
            (target / "controller.py").write_text(original.replace(old, new, 1))
            command = [sys.executable, str(target / "test_controller.py"), "-v", "ControllerTests." + test]
            result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                                    text=True, env={**os.environ, "PYTHONDONTWRITEBYTECODE": "1"}, timeout=120)
            assertion = result.returncode == 1 and "FAIL:" in result.stdout and "AssertionError" in result.stdout and "ERROR:" not in result.stdout
            print(f"=== {name}: {'KILLED' if assertion else 'INVALID OR SURVIVED'} (exit {result.returncode}) ===", flush=True)
            print(f"controller.py: {old!r} -> {new!r}", flush=True)
            print("command:", " ".join(command), flush=True)
            print(result.stdout, flush=True)
            if not assertion:
                raise RuntimeError(f"mutation not killed by an assertion: {name}")
            killed += 1
    print(f"Guard mutants killed: {killed}/{len(MUTANTS)}")


if __name__ == "__main__":
    main()
