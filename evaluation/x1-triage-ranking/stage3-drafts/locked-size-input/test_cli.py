"""Actual synthetic CLI integration; full scientific counts, no mocked measurements.

Set X1G_CLI_OUTPUT to a nonexistent directory to preserve all evidence.
"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

import locked_commands as commands
from test_size_input import check_roundtrip


class SyntheticCLITests(unittest.TestCase):
    def test_full_cli_sequence_and_provenance_roundtrip(self):
        destination = os.environ.get("X1G_CLI_OUTPUT")
        temporary = None
        if destination:
            folder = Path(destination)
            folder.mkdir(parents=True, exist_ok=False)
        else:
            temporary = tempfile.TemporaryDirectory(prefix="x1g-cli-")
            folder = Path(temporary.name)
            self.addCleanup(temporary.cleanup)
        driver = str(commands.HERE/"locked_commands.py")
        paths = {name: folder/(name+".json") for name in
                 ("synthetic-input", "gate", "benchmark", "calibration", "fit")}
        invocations = []
        for mode in paths:
            argv = [sys.executable, driver, mode, "--output", str(paths[mode])]
            if mode != "synthetic-input":
                argv += ["--input", str(paths["synthetic-input"])]
            if mode == "benchmark":
                argv += ["--trials", "2000"]
            if mode == "fit":
                for name in ("benchmark", "calibration", "gate"):
                    argv += ["--"+name, str(paths[name])]
            completed = subprocess.run(argv, capture_output=True, text=True)
            invocations.append({"argv": argv, "returncode": completed.returncode,
                                "stdout": completed.stdout, "stderr": completed.stderr})
            self.assertEqual(completed.returncode, 0, completed.stderr)
        records = {name: json.loads(path.read_bytes()) for name, path in paths.items()}
        for mode in ("gate", "benchmark", "calibration", "fit"):
            result = records[mode]
            self.assertTrue(result["completed"])
            self.assertIsNone(result["integrity_failure"])
            self.assertEqual(result["python_optimize"], 0)
            self.assertTrue(result["python_executable"])
            self.assertEqual(result["provenance"], result["end_provenance"])
            self.assertEqual(set(result["host"]), {"hw.model", "hw.memsize", "logical_cpus", "os_build"})
            for load in result["host_load"].values():
                self.assertEqual(len(load["load_averages"]), 3)
                self.assertGreater(load["concurrent_process_count"], 0)
        benchmark = records["benchmark"]
        commands.validate_benchmark(benchmark)
        self.assertEqual(benchmark["trials"], 2000)
        self.assertEqual(benchmark["replicates"], 1999)
        self.assertEqual(sum(c["measured_trials"] for r in benchmark["executions"] for c in r["cells"]), 32000)
        fit = records["fit"]
        self.assertEqual(fit["locked_inventory"], records["gate"]["locked_inventory"])
        self.assertEqual(fit["measurement_runs"]["locked_inventory_sha256"],
                         hashlib.sha256(commands.canonical(fit["locked_inventory"])).hexdigest())
        evaluated = check_roundtrip(self, fit["measurement_runs"]["runs"][0], fit["locked_inventory"])
        for name in ("benchmark", "calibration", "gate"):
            self.assertEqual(fit["evidence_sha256"][str(paths[name])], hashlib.sha256(paths[name].read_bytes()).hexdigest())
        before = paths["gate"].read_bytes()
        duplicate = subprocess.run([sys.executable, driver, "gate", "--input", str(paths["synthetic-input"]),
                                    "--output", str(paths["gate"])], capture_output=True, text=True)
        self.assertNotEqual(duplicate.returncode, 0)
        self.assertEqual(paths["gate"].read_bytes(), before)
        commands.publish(folder/"cli-roundtrip.json", {"synthetic_only": True, "invocations": invocations,
                         "resource_fit": evaluated, "duplicate_publication_exit": duplicate.returncode})


if __name__ == "__main__":
    unittest.main()
