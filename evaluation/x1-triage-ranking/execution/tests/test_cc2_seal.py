"""Synthetic CC2 implementation/output authority and canonical seal integration."""
import copy
import json
from pathlib import Path
import unittest

import preparation as prep
import test_binding_seal as binding_fixture
from verify_inputs import pin


class DeltaSealTests(unittest.TestCase):
    def ready(self):
        fixture = binding_fixture.BindingSealIntegrationTests('test_real_adopt_rehearse_canonical_django_and_v2_seal')
        fixture.setUp(); self.addCleanup(fixture.doCleanups)
        fixture.prepare()
        report, _ = fixture.rehearsal('cc2-seal.json')
        fixture.canonical(report)
        return fixture

    def passed(self, fixture):
        clock = prep.read_clock(fixture.config, fixture.config_path, fixture.clock_path)
        return prep.collect_journals(fixture.config, fixture.config_path, fixture.clock_path, clock)[4]

    def collect_delta(self, fixture, passed=None):
        return prep.collect_cc2_delta(fixture.config, fixture.config_path, fixture.clock_path,
                                      self.passed(fixture) if passed is None else passed)

    def test_full_seal_retains_baseline_and_canonical_delta_roles_and_scopes(self):
        fixture = self.ready()
        baseline = [Path(row['path']).read_bytes() for row in fixture.cc2_baseline_outputs]
        seal = fixture.collect()
        outputs = {row['path']: row for row in seal['outputs']}
        for name, kind in (('manifest.json', 'scout.x1.identifier-delta-cc2.v1'),
                           ('coverage.json', 'scout.x1.identifier-delta-cc2-coverage.v1'),
                           ('familiarity.jsonl', 'scout.x1.familiarity-delta-cc2.v1')):
            self.assertEqual(outputs[str(fixture.out/'identifiers-delta-cc2'/name)]['role'], kind)
        scopes = {row['scope'] for row in seal['scoped_reviews']}
        self.assertTrue({'identifier-delta-cc2-implementation', 'identifier-delta-cc2-output'} <= scopes)
        self.assertEqual(baseline, [Path(row['path']).read_bytes() for row in fixture.cc2_baseline_outputs])
        self.assertEqual(fixture.clock_path.read_bytes(), fixture.original_clock)

    def test_full_seal_refuses_missing_delta_output(self):
        fixture = self.ready()
        path = fixture.out/'identifiers-delta-cc2/coverage.json'
        path.unlink()
        with self.assertRaisesRegex(FileNotFoundError, 'coverage.json'):
            fixture.collect()

    def test_full_seal_refuses_changed_execution_source(self):
        fixture = self.ready()
        (fixture.execution/'preparation.py').write_text('changed after review\n')
        with self.assertRaisesRegex(ValueError, 'pin mismatch'):
            fixture.collect()

    def test_both_review_receipts_are_required_and_scope_specific(self):
        fixture = self.ready(); passed = self.passed(fixture)
        for path in (fixture.cc2_implementation_review, fixture.cc2_output_review):
            original = path.read_bytes()
            try:
                path.unlink()
                with self.assertRaisesRegex(FileNotFoundError, path.name):
                    self.collect_delta(fixture, passed)
                value = json.loads(original); value['scope'] = 'execution-implementation'
                path.write_text(json.dumps(value))
                with self.assertRaisesRegex(ValueError, 'invalid scoped review receipt'):
                    self.collect_delta(fixture, passed)
            finally:
                path.write_bytes(original)

    def test_implementation_review_must_bind_every_control(self):
        fixture = self.ready(); passed = self.passed(fixture)
        path = fixture.cc2_implementation_review; original = path.read_bytes()
        for index in range(16):
            with self.subTest(control=index):
                value = json.loads(original); value['reviewed'].pop(index)
                path.write_text(json.dumps(value))
                with self.assertRaisesRegex(ValueError, 'CC2 implementation review omits bound controls'):
                    self.collect_delta(fixture, passed)
        path.write_bytes(original)

    def test_output_review_must_bind_all_outputs_and_the_passed_journal(self):
        fixture = self.ready(); passed = self.passed(fixture)
        path = fixture.cc2_output_review; original = path.read_bytes()
        for index in range(4):
            with self.subTest(evidence=index):
                value = json.loads(original); value['reviewed'].pop(index)
                path.write_text(json.dumps(value))
                with self.assertRaisesRegex(ValueError, 'CC2 output review omits delta outputs/producer journal'):
                    self.collect_delta(fixture, passed)
        path.write_bytes(original)

    def test_delta_schema_clock_chain_and_union_semantics_are_required(self):
        fixture = self.ready(); passed = self.passed(fixture)
        path = fixture.out/'identifiers-delta-cc2/manifest.json'; original = path.read_bytes()
        cases = [('schema', 'wrong', 'schema/status/clock'), ('status', 'refused', 'schema/status/clock'),
                 ('clock', {}, 'schema/status/clock'), ('implementation_review', {}, 'implementation/source/baseline'),
                 ('sources_manifest', {}, 'implementation/source/baseline'), ('predecessors', [], 'implementation/source/baseline'),
                 ('development_freshness_filter', True, 'failed/unmeasured/freshness'),
                 ('unmeasured', ['missing'], 'failed/unmeasured/freshness'),
                 ('failures', ['failed'], 'failed/unmeasured/freshness'),
                 ('merge_semantics', 'replace', 'conservative append-union'),
                 ('unknown_freshness', 'fresh', 'conservative append-union'),
                 ('candidate_exclusions_computed', True, 'conservative append-union')]
        for field, changed, message in cases:
            with self.subTest(field=field):
                value = json.loads(original); value[field] = changed; path.write_text(json.dumps(value))
                with self.assertRaisesRegex(ValueError, message):
                    self.collect_delta(fixture, passed)
        path.write_bytes(original)

    def test_familiarity_header_must_match_manifest_sources(self):
        fixture = self.ready(); passed = self.passed(fixture)
        path = fixture.out/'identifiers-delta-cc2/familiarity.jsonl'; original = path.read_bytes()
        for field, changed in (('entry_kind', 'wrong'), ('sources', [{'source_id': 7}])):
            with self.subTest(field=field):
                value = json.loads(original); value[field] = changed; path.write_text(json.dumps(value)+'\n')
                with self.assertRaisesRegex(ValueError, 'CC2 familiarity source-table mismatch'):
                    self.collect_delta(fixture, passed)
        path.write_bytes(original)

    def test_unique_passed_action_d_and_producer_control_pins_are_required(self):
        fixture = self.ready(); passed = self.passed(fixture)
        producer = fixture.cc2_journal['path']
        for mode in ('missing', 'wrong-action', 'duplicate', 'missing-source'):
            with self.subTest(mode=mode):
                changed = copy.deepcopy(passed)
                message = 'CC2 delta lacks unique passed action-d producer'
                if mode == 'missing': changed.pop(producer)
                elif mode == 'wrong-action': changed[producer]['action'] = 'c'
                elif mode == 'duplicate': changed[producer+'-duplicate'] = changed[producer]
                else:
                    changed[producer]['source_pins'].pop()
                    message = 'CC2 delta producer omits implementation/source pins'
                with self.assertRaisesRegex(ValueError, message):
                    self.collect_delta(fixture, changed)

    def test_baseline_outputs_cannot_be_substituted(self):
        fixture = self.ready(); passed = self.passed(fixture)
        Path(fixture.cc2_baseline_outputs[0]['path']).write_bytes(b'substituted')
        with self.assertRaisesRegex(ValueError, 'pin mismatch: .*identifiers/manifest.json'):
            self.collect_delta(fixture, passed)


if __name__ == '__main__':
    unittest.main()
