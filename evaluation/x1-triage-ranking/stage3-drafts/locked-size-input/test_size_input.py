"""Synthetic-only regression suite. Run explicitly; imports perform no measurement."""
import ast
import copy
from dataclasses import FrozenInstanceError
import hashlib
import itertools
import json
import math
import time
from pathlib import Path
import subprocess
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch

from size_input import (SizeInput, REPOSITORIES, synthetic_record, canonical,
                        GATE_CELL, controller_seed, gate_passes)
from population import prepare_population
import fixed_population as kernel
import locked_commands as commands


def inputs(vectors=None):
    return SizeInput(synthetic_record(vectors or {r: [50]*20 for r in REPOSITORIES}))


def asymmetric():
    # Neither outer repository contains the global extrema.
    return {REPOSITORIES[0]: [50, 100]*11,
            REPOSITORIES[1]: [500]*9+[20, 2000]+[500]*12,
            REPOSITORIES[2]: [75, 150]*12}


class SizeTests(unittest.TestCase):
    def test_equal_vectors_preserve_exact_population(self):
        sizes = tuple([50, 60]*10)
        source = inputs({r: sizes for r in REPOSITORIES})
        self.assertEqual(prepare_population(size_input=source, seed=169, size_dependence=.05),
                         kernel.prepare_population(sizes=sizes, run_count=20, seed=169, size_dependence=.05))

    def test_asymmetric_offsets_and_constant_zero(self):
        vectors = asymmetric()
        source = inputs(vectors)
        for factor in (.05, .20):
            prepared = prepare_population(size_input=source, size_dependence=factor)
            self.assertEqual(len(prepared.runs), 69)
            for run in prepared.runs:
                n = vectors[run.repo][int(run.run)]
                self.assertEqual(run.n, n)
                self.assertEqual(run.size_offset, factor*(n-20)/1980)
        constant = inputs({r: [77]*(20+i) for i, r in enumerate(REPOSITORIES)})
        self.assertEqual([r.size_offset for r in prepare_population(size_input=constant, size_dependence=.20).runs], [0.]*63)

    def test_planted_effect_null_and_truth_identity(self):
        for factor, target in itertools.product((.05, .20), (0., .20)):
            with self.subTest(factor=factor, target=target):
                prepared = prepare_population(size_input=inputs(asymmetric()), seed=169,
                                              size_dependence=factor, prevalence=.20, cap=2000)
                result = kernel.calibrate_population(prepared, target)
                self.assertTrue(result['reachable'])
                self.assertGreaterEqual(result['realised'], .20) if target else self.assertLessEqual(result['realised'], 0)
                self.assertEqual(kernel._population_gain(prepared, result['latent_gain']),
                                 kernel.realised_gain(result['frame']))

    def test_mismatched_truth_mutant_raises(self):
        prepared = prepare_population(size_input=inputs(asymmetric()), size_dependence=.20)
        original = kernel._population_gain
        with patch.object(kernel, '_population_gain', side_effect=lambda *args: original(*args)+.001):
            with self.assertRaisesRegex(RuntimeError, 'truth/full-frame mismatch'):
                kernel.calibrate_population(prepared, .20)

    def test_constructor_run_count_and_size_boundaries(self):
        for length in (19, 20, 2000, 2001):
            vectors = {r: [50]*length for r in REPOSITORIES}
            with self.subTest(length=length):
                if length in (20, 2000):
                    self.assertEqual(inputs(vectors).lengths[REPOSITORIES[0]], length)
                else:
                    with self.assertRaises(ValueError):
                        inputs(vectors)
        for n in (19, 20, 2000, 2001, True, 50.0):
            with self.subTest(n=n):
                vectors = {r: [n]*20 for r in REPOSITORIES}
                if type(n) is int and n in (20, 2000):
                    self.assertEqual(inputs(vectors).size(REPOSITORIES[0], 0), n)
                else:
                    with self.assertRaises(ValueError):
                        inputs(vectors)

    def test_direct_malformed_records_and_position_boundary(self):
        original = synthetic_record({r: [50]*20 for r in REPOSITORIES})
        for count in (2, 4):
            record = copy.deepcopy(original)
            record["repositories"] = (record["repositories"]*2)[:count]
            with self.assertRaisesRegex(ValueError, "exactly three"):
                SizeInput(record)
        for level in ("record", "repository", "run"):
            for mutation in ("extra", "missing"):
                record = copy.deepcopy(original)
                target = record if level == "record" else record["repositories"][0]
                if level == "run":
                    target = target["runs"][0]
                if mutation == "extra":
                    target["unexpected"] = 0
                else:
                    del target[next(iter(target))]
                with self.subTest(level=level, mutation=mutation), self.assertRaises(ValueError):
                    SizeInput(record)
        original["repositories"][0]["runs"][-1]["first_parent_position"] = 1999
        self.assertEqual(SizeInput(original).lengths[REPOSITORIES[0]], 20)
        original["repositories"][0]["runs"][-1]["first_parent_position"] = 2000
        with self.assertRaisesRegex(ValueError, "positions"):
            SizeInput(original)

    def test_identity_digest_and_invalid_order(self):
        record = synthetic_record({r: [50]*20 for r in REPOSITORIES})
        source = SizeInput(record)
        # Independently generated via openssl dgst -sha256; see README.
        self.assertEqual(source.sha256, '7f97e552a3e6ec6b4b056c3905056ea58ce1c82251a9cd0cfba0cddc12aa496a')
        self.assertEqual(source.sha256, SizeInput(dict(reversed(list(record.items())))).sha256)
        for mutation in ('size', 'sha', 'seal'):
            changed = copy.deepcopy(record)
            if mutation == 'size':
                changed['repositories'][0]['runs'][0]['size'] = 51
            elif mutation == 'sha':
                changed['repositories'][0]['runs'][0]['commit_sha'] = 'f'*40
            else:
                changed['inventory_seal_sha256'] = '1'*64
            self.assertNotEqual(source.sha256, SizeInput(changed).sha256)
        for mutation in ('runs', 'repositories', 'tip', 'mapping', 'sha', 'duplicate', 'schema', 'real'):
            changed = copy.deepcopy(record)
            runs = changed['repositories'][0]['runs']
            if mutation == 'runs':
                runs[0], runs[1] = runs[1], runs[0]
            elif mutation == 'repositories':
                changed['repositories'][0], changed['repositories'][1] = changed['repositories'][1], changed['repositories'][0]
            elif mutation == 'tip':
                changed['repositories'][0]['pinned_tip'] = 'f'*40
            elif mutation == 'mapping':
                changed['repositories'][0]['repository'] = 'other/other'
            elif mutation == 'sha':
                runs[0]['commit_sha'] = 'short'
            elif mutation == 'duplicate':
                runs[0]['commit_sha'] = runs[1]['commit_sha']
            elif mutation == 'schema':
                changed['schema'] = 'other'
            else:
                changed['synthetic_only'] = False
            self.assertNotEqual(hashlib.sha256(canonical(changed)).hexdigest(), source.sha256)
            with self.assertRaises(ValueError):
                SizeInput(changed)

    def test_immutable_snapshot(self):
        record = synthetic_record({r: [50]*20 for r in REPOSITORIES})
        source = SizeInput(record)
        record['repositories'][0]['runs'][0]['size'] = 2000
        self.assertEqual(source.minimum, 50)
        self.assertEqual(source.maximum, 50)
        self.assertEqual(source.size(REPOSITORIES[0], 0), 50)
        with self.assertRaises(TypeError):
            source.vectors[REPOSITORIES[0]] = (2000,)*20
        with self.assertRaises(TypeError):
            source.vectors[REPOSITORIES[0]][0] = 2000
        with self.assertRaises(FrozenInstanceError):
            source.sha256 = 'stale'

    def test_seed_and_heterogeneous_selection_known_answers(self):
        message = bytes.fromhex('7b226b696e64223a2273616d706c696e672d676174652d657374696d617465227d003000300073616d706c696e672d676174652d657374696d617465')
        expected = 'f98555ac5c97329b205df7cf34ef7c58400eafd0ee5b3013253ee37552e418a3'
        digest = subprocess.check_output(['openssl', 'dgst', '-sha256', '-hmac', '169-x1-oc-v4'], input=message).decode().split()[-1]
        self.assertEqual(digest, expected)
        self.assertEqual(controller_seed(GATE_CELL, 0, 0, 'sampling-gate-estimate'), 17979871286122721947)
        self.assertEqual(controller_seed({'kind': 'sampling-gate-estimate'}, 0, 0, 'sampling-gate-estimate'), 17979871286122721947)
        last_message = bytes.fromhex('7b226b696e64223a2273616d706c696e672d676174652d657374696d617465227d00300039393939390073616d706c696e672d676174652d657374696d617465')
        last_digest = subprocess.check_output(['openssl', 'dgst', '-sha256', '-hmac', '169-x1-oc-v4'], input=last_message).decode().split()[-1]
        self.assertEqual(last_digest, 'a66f0e557732e338843ccea658cde74eb941e7fc5f6d0e83d1a8879a8725d494')
        self.assertEqual(controller_seed(GATE_CELL, 0, 99999, 'sampling-gate-estimate'),
                         int('a66f0e557732e338', 16))
        source = inputs({r: [50+i*10 for i in range(n)] for r, n in zip(REPOSITORIES, (23, 27, 31))})
        selected, passed = source.gate_selection(0)
        expected_ids = (
            [10,5,14,7,18,8,0,2,12,11,19,3,13,6,20,1,22,21,9,17],
            [26,9,10,6,2,24,7,14,16,18,8,20,23,12,13,21,22,0,25,11],
            [29,20,10,22,6,17,13,18,5,4,27,28,16,12,21,14,26,8,25,30])
        self.assertEqual(list(selected.values()), [tuple(v) for v in expected_ids])
        self.assertTrue(passed)

    def test_gate_boundaries(self):
        self.assertFalse(gate_passes([999, 1000, 1000]))
        self.assertTrue(gate_passes([1000, 1000, 1000]))
        self.assertTrue(gate_passes([20000, 20000, 20000]))
        self.assertFalse(gate_passes([20000, 20000, 20001]))
        self.assertFalse(inputs({r: [20]*21 for r in REPOSITORIES}).gate_selection(0)[1])
        self.assertTrue(inputs({r: [50]*21 for r in REPOSITORIES}).gate_selection(0)[1])

    def test_full_100000_key_estimates(self):
        import size_input
        original, attempts = size_input.controller_seed, []
        def observed(cell, population, attempt, phase):
            self.assertEqual((cell, population, phase), (GATE_CELL, 0, 'sampling-gate-estimate'))
            attempts.append(attempt)
            return original(cell, population, attempt, phase)
        with patch.object(size_input, 'controller_seed', side_effect=observed):
            certain = inputs({r: [50]*(21+i) for i, r in enumerate(REPOSITORIES)}).estimate_p_gate()
        self.assertEqual(attempts, list(range(100000)))
        self.assertEqual((certain['successes'], certain['p_gate'], certain['attempt_cap_per_population']), (100000, 1., 4000))
        impossible = inputs({r: [2000]*(21+i) for i, r in enumerate(REPOSITORIES)}).estimate_p_gate()
        self.assertEqual((impossible['successes'], impossible['p_gate'], impossible['attempt_cap_per_population'], impossible['status']), (0, 0., None, 'not testable'))
        for result in (certain, impossible):
            p, n, z = result['p_gate'], 100000, 1.959963984540054
            centre = (2*n*p+z*z)/(2*(n+z*z))
            radius = z*((z*z+4*n*p*(1-p))**.5)/(2*(n+z*z))
            for got, expected in zip(result['Wilson_MC95'], (centre-radius, centre+radius)):
                self.assertAlmostEqual(got, expected)

    def test_interior_gate_wilson_and_ceiling(self):
        source = inputs()
        with patch.object(SizeInput, "gate_selection", side_effect=lambda i: ({}, i < 3)) as selection:
            result = source.estimate_p_gate()
        self.assertEqual(selection.call_count, 100000)
        self.assertEqual(result["attempt_cap_per_population"], 133333334)
        n, k, z = 100000, 3, 1.959963984540054
        centre = (2*k+z*z)/(2*(n+z*z))
        radius = z*math.sqrt(z*z+4*k*(1-k/n))/(2*(n+z*z))
        for actual, expected in zip(result["Wilson_MC95"], (centre-radius, centre+radius)):
            self.assertAlmostEqual(actual, expected, places=15)

    def test_only_size_source_changed_in_kernel(self):
        old = ast.parse((commands.VALIDATION/'fixed_population.py').read_text())
        new = ast.parse(Path(__file__).with_name('population.py').read_text())
        a = next(n for n in old.body if isinstance(n, ast.FunctionDef) and n.name == 'prepare_population')
        b = next(n for n in new.body if isinstance(n, ast.FunctionDef) and n.name == 'prepare_population')
        text = ast.unparse(b).replace('size_input,', 'run_count=20, sizes=(50,),')
        text = text.replace('range(len(size_input.vectors[repo]))', 'range(run_count)')
        text = text.replace('size_input.size(repo, run)', 'sizes[run % len(sizes)]')
        text = text.replace('size_input.offset(repo, run, size_dependence)', 'size_dependence * (n - min(sizes)) / max(1, max(sizes) - min(sizes))')
        self.assertEqual(ast.dump(ast.parse(ast.unparse(a))), ast.dump(ast.parse(text)))


class LockedModeTests(unittest.TestCase):
    def test_construction_calibration_consume_every_locked_size(self):
        source = inputs(asymmetric())
        functions = commands.locked_functions(source)
        constructed = []
        def observe(**kwargs):
            result = prepare_population(**kwargs)
            self.assertEqual([r.n for r in result.runs], [n for r in REPOSITORIES for n in asymmetric()[r]])
            constructed.append(result)
            return result
        functions['locked_prepare'] = observe
        benchmark = functions['benchmark_population'](40)
        calibration = functions['calibration_measurement'](400)
        self.assertEqual(len(constructed), 2)
        self.assertEqual(benchmark['units'], source.units)
        self.assertEqual(len(benchmark['fixed_key_evaluations']), 22)
        self.assertTrue(all(e['units'] == source.units for e in benchmark['fixed_key_evaluations']))
        self.assertEqual(calibration['units'], source.units)
        self.assertEqual(calibration['frame_evaluations'], 1)
        self.assertLessEqual(calibration['truth_only_evaluations'], 21)
        self.assertEqual(benchmark['size_only_sampling_gate']['draws'], 100000)

    def test_trial_population_lengths_and_selected_ids(self):
        source = inputs({r: [50]*(21+i*3) for i, r in enumerate(REPOSITORIES)})
        functions = commands.locked_functions(source)
        for fixture in ('uniform-srs', 'mixed-explicit'):
            requests = []
            def helper(command, *, input, **kwargs):
                requests.append(json.loads(input))
                return SimpleNamespace(stdout=b'{"Combinations":512}\n')
            with patch.object(commands.benchmark.subprocess, 'run', side_effect=helper):
                result = functions['execute_cell']((40, True, 'studentized', 1, '/unused-synthetic-helper', fixture))
            expected_lengths = list(source.lengths[r] for r in sorted(REPOSITORIES)) if fixture == 'uniform-srs' else [20]*3
            self.assertEqual(list(map(len, requests[0]['Population'])), expected_lengths)
            self.assertEqual(requests[0]['B'], 1999)
            self.assertEqual(result['scientific']['selected_units_every_trial'], 60000)
            _, selected = functions['trial_frame'](40, False, 0, fixture)
            for repo, ids in selected.items():
                self.assertEqual(len(set(ids)), 20)
                self.assertTrue(all(0 <= i < source.lengths[repo] for i in ids))
                self.assertLess(max(ids), 40)

    def test_actual_helper_one_two_worker_science(self):
        # Small synthetic L_r; retain the signed helper's full replicate structure.
        source = inputs({r: [50]*(21+i) for i, r in enumerate(REPOSITORIES)})
        with tempfile.TemporaryDirectory() as temporary:
            binary = str(Path(temporary)/'interval-kernel')
            subprocess.run(['go', 'build', '-o', binary, str(commands.VALIDATION/'interval_diagnostic.go')], check=True)
            tasks = [(slot, missing, 'studentized', 2000, binary, fixture)
                     for slot, missing, fixture in itertools.product((40, 400), (False, True), ('uniform-srs', 'mixed-explicit'))]
            runs = [commands.execute_trials(source, tasks, workers) for workers in (1, 2)]
        self.assertEqual(canonical(commands.benchmark.science(runs[0])), canonical(commands.benchmark.science(runs[1])))
        self.assertEqual(len(runs[0]['cells']), 8)
        for cell in runs[0]['cells']:
            self.assertEqual(cell['cell']['L_r'], source.lengths)
            self.assertEqual(cell['measured_trials'], 2000)
            self.assertEqual(len(cell['scientific']['trial_outputs']), 2000)
            if cell['cell']['missing_labels']:
                self.assertTrue(all(output['Combinations'] == 512
                                    for output in cell['scientific']['trial_outputs']))

    def test_size_site_drift_refused(self):
        with patch.object(commands, 'SIZE_SITES', {**commands.SIZE_SITES, 'benchmark_population': (('123456789', 'size_input'),)}):
            with self.assertRaisesRegex(RuntimeError, 'signed size site drift'):
                commands.locked_functions(inputs())

    def test_extraction_uses_locked_maxima_not_slopes(self):
        source, data, calibration, gate = measurement_fixture()
        result = commands.measurement_run(source, data, calibration, gate, {"synthetic": "1"*64})
        values = result["binding_inputs"]
        self.assertEqual(values["construction_cpu_per_unit"], 1/source.units)
        self.assertEqual(values["requested_base_cpu_per_unit"], 2/source.units)
        self.assertEqual(values["calibration_cpu_per_unit"], 3/source.units)
        self.assertEqual(values["enumeration_cpu_per_population"], 4)
        self.assertEqual(values["construction_rss"], 1000)
        self.assertEqual(values["calibration_rss"], 2000)
        self.assertEqual(values["trial_plus_helper_rss"], 3000)
        self.assertNotIn("fits_caps", result)
        calibration["measurements"][0]["peak_rss_bytes"] = 0
        with self.assertRaisesRegex(ValueError, "unmeasured"):
            commands.measurement_run(source, data, calibration, gate, {})

    def test_frozen_method_and_cell_set_refused(self):
        for mutation in ("method", "cell_method", "missing_cell", "duplicate_cell", "workers", "count"):
            source, data, calibration, gate = measurement_fixture()
            if mutation == "method":
                data["method"] = "percentile"
            elif mutation == "cell_method":
                data["executions"][0]["cells"][0]["cell"]["method"] = "percentile"
            elif mutation == "missing_cell":
                data["executions"][0]["cells"].pop()
            elif mutation == "duplicate_cell":
                data["executions"][0]["cells"][-1] = copy.deepcopy(data["executions"][0]["cells"][0])
            elif mutation == "workers":
                data["executions"][1]["workers"] = 8
            else:
                data["executions"][0]["cells"][0]["measured_trials"] = 1999
            with self.subTest(mutation=mutation), self.assertRaises(ValueError):
                commands.measurement_run(source, data, calibration, gate, {})

    def test_patched_signed_digest_refused_before_measurement(self):
        for name in commands.SIGNED:
            with self.subTest(source=name), patch.dict(commands.SIGNED, {name: "0"*64}):
                with self.assertRaisesRegex(ValueError, "substituted signed source"):
                    commands.locked_functions(inputs())

    def test_evidence_single_read_hashes_parsed_bytes(self):
        path = SimpleNamespace(read_bytes=unittest.mock.Mock(side_effect=[b'{"v":1}', b'{"v":2}']))
        records, digests = commands.read_evidence([path])
        self.assertEqual(records, [{"v": 1}])
        self.assertEqual(list(digests.values()), [hashlib.sha256(b'{"v":1}').hexdigest()])
        path.read_bytes.assert_called_once_with()

    def test_closure_failure_published_create_once_then_nonzero(self):
        source = inputs()
        basis = json.loads(canonical(commands.provenance(source)))
        with tempfile.TemporaryDirectory() as temporary:
            input_path, output = Path(temporary)/"input.json", Path(temporary)/"result.json"
            closure_source = Path(temporary)/"synthetic-source.py"
            closure_source.write_bytes(b"original synthetic source\n")
            def closure(_):
                value = copy.deepcopy(basis)
                value["source_sha256"]["synthetic-source.py"] = hashlib.sha256(closure_source.read_bytes()).hexdigest()
                return value
            start = closure(source)
            # Measurements have finished; mutate an actual disposable closure source.
            measured = {"values": [123]}
            closure_source.write_bytes(b"changed after measurement\n")
            end = closure(source)
            input_path.write_bytes(source.record_bytes)
            with patch.object(commands, "provenance", side_effect=closure), patch.object(
                    commands, "host_load", return_value={"load_averages": [0., 0., 0.], "concurrent_process_count": 1}):
                with self.assertRaisesRegex(RuntimeError, "completed evidence published"):
                    commands.finish(source, input_path, output, "benchmark", measured,
                                    start, {"synthetic": True}, {}, {}, time.perf_counter())
                result = json.loads(output.read_bytes())
                self.assertTrue(result["completed"])
                self.assertEqual(result["integrity_failure"], "closure mismatch")
                self.assertEqual(result["values"], [123])
                self.assertEqual(result["provenance"], start)
                self.assertEqual(result["end_provenance"], end)
                with self.assertRaises(FileExistsError):
                    commands.finish(source, input_path, output, "benchmark", {}, start, {}, {}, {}, time.perf_counter())

    def test_evaluator_provenance_roundtrip(self):
        source, data, calibration, gate = measurement_fixture()
        record = commands.measurement_run(source, data, calibration, gate, {"synthetic": "1"*64})
        check_roundtrip(self, record, commands.locked_inventory(source, gate))
        data["integrity_failure"] = "closure mismatch"
        invalid = commands.measurement_run(source, data, calibration, gate, {"invalid": "2"*64})
        invalid["binding_inputs"]["construction_rss"] = 9000
        provenance, evaluator = evaluator_modules()
        maxima, reports = provenance.maximum_completed_inputs(
            json.loads(canonical([invalid, record])), record["host"])
        self.assertEqual(maxima["construction_rss"], 9000)
        self.assertEqual(reports[0]["classification"], "completed-invalid-included")


def measurement_fixture():
    """Fabricated unit-test records, never actual timing/RSS evidence."""
    source = inputs()
    common = {"provenance": json.loads(canonical(commands.provenance(source))),
              "host": {"hw.model": "synthetic", "hw.memsize": 64*1024**3, "logical_cpus": 8, "os_build": "synthetic"},
              "host_load": {"start": {"load_averages": [0., 0., 0.], "concurrent_process_count": 1},
                            "end": {"load_averages": [0., 0., 0.], "concurrent_process_count": 1}},
              **commands.interpreter(), "completed": True, "integrity_failure": None}
    cells = [{"cell": {"L": slot, "missing_labels": missing, "method": "studentized", "fixture": fixture},
              "measured_trials": 2000, "trial_timing": [{"cpu_seconds": .000001}],
              "peak_rss_bytes": 3000}
             for slot, missing, fixture in itertools.product((40, 400), (False, True),
                                                             ("uniform-srs", "mixed-explicit"))]
    data = dict(common, method="studentized", trials=2000, replicates=1999,
                worker_count_equality={"byte_identical_scientific_json": True},
                population_measurements=[{"construction": {"cpu_seconds": 1., "peak_rss_bytes": 1000},
                                          "fixed_key_evaluations": [{"cpu_seconds": 2.}],
                                          "size_only_sampling_gate": {"cpu_seconds": 1.}}],
                executions=[{"workers": w, "cells": copy.deepcopy(cells)} for w in (1, 2)])
    calibration = dict(common, measurements=[{"conservative_21_truth_one_frame_no_enumeration_cpu_seconds": 3.,
                                              "frame_evaluation_cpu_seconds": [2.], "peak_rss_bytes": 2000}],
                       worst_case_candidate_enumeration={"cpu_seconds": 4.})
    gate = dict(common, draws=100000, successes=100000, p_gate=1., attempt_cap_per_population=4000)
    return source, data, calibration, gate


def evaluator_modules():
    import sys
    folder = commands.HERE.parent/"final-evaluator"
    with patch.object(sys, "path", [str(folder), *sys.path]):
        provenance = commands.load_file("provenance", folder/"provenance.py")
        evaluator = commands.load_file("x1_size_evaluator", folder/"evaluator.py")
    return provenance, evaluator


def check_roundtrip(case, record, inventory):
    provenance, evaluator = evaluator_modules()
    record, inventory = json.loads(canonical(record)), json.loads(canonical(inventory))
    maxima, reports = provenance.maximum_completed_inputs([record], record["host"])
    manifest = json.loads((commands.VALIDATION.parent/"oc-manifest.json").read_bytes())
    fit = evaluator.resource_fit(maxima, inventory, manifest, record["host"])
    case.assertEqual(fit["maximum_inputs"], record["binding_inputs"])
    case.assertEqual(fit["units_per_population"], sum(map(sum, inventory["eligible_sizes"].values())))
    case.assertEqual(reports[0]["classification"], "completed-valid-included")
    return fit


if __name__ == '__main__':
    unittest.main()
