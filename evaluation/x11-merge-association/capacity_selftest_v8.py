#!/usr/bin/env python3
"""Pure synthetic plants for the frozen held-out capacity rule; no candidate access."""
from fractions import Fraction
import unittest

import preparation_capacity_v8 as capacity
import prep_scale_v8 as scale


class CapacityTests(unittest.TestCase):
    def test_identified_additive_coefficients(self):
        rows = [{'files': f, 'bytes': b, 'revisions': n, 'elapsed_seconds': n*(2*f+3*b+5)}
                for f, b, n in ((1, 1, 3), (10, 1, 20), (1, 10, 3))]
        self.assertEqual(capacity.fit(rows), (2, 3, 5))
        self.assertEqual(capacity.projected([{'files': 1, 'bytes': 1}], (2, 3, 5), Fraction(1))['per_attempt_seconds'], 10)
        self.assertEqual(capacity.projected([{'files': 1, 'bytes': 1}], (2, 3, 5))['per_attempt_seconds'], 15)

    def test_direct_bound_uses_larger_observation_or_prediction(self):
        for observed, prediction in ((100, 200), (200, 100)):
            value = capacity.direct_bound(observed, prediction)
            self.assertEqual(value['per_attempt_seconds'], 1.5*1668*200)
            self.assertEqual(value['three_attempt_seconds'], 3*1.5*1668*200)
            self.assertEqual(value['largest_revision_seconds'], 1.5*200)
            self.assertTrue(all(value[k] for k in capacity.GATES))
            self.assertEqual(capacity.status([value]), 'projected-feasible-not-authorised')
        self.assertEqual(capacity.direct_bound(100, 200, margin=Fraction(1))['per_attempt_seconds'], 1668*200)

    def test_linked_attempt_bounds_and_exact_boundary(self):
        boundary = Fraction(14*86400, 1)/Fraction(3, 2)/1668
        equal = capacity.direct_bound(boundary, 1)
        self.assertTrue(equal['attempt_fits'])
        self.assertTrue(equal['three_attempts_fit'])
        over = capacity.direct_bound(boundary+Fraction(1, 1000000), 1)
        self.assertFalse(over['attempt_fits'])
        self.assertFalse(over['three_attempts_fit'])
        self.assertTrue(over['largest_revision_fits_subprocess_ceiling'])
        self.assertEqual(capacity.status([over]), 'preparation not feasible')

    def test_actual_subprocess_failure_plant(self):
        value = capacity.direct_bound(2401, 1)
        self.assertFalse(value['attempt_fits'])
        self.assertFalse(value['three_attempts_fit'])
        self.assertFalse(value['largest_revision_fits_subprocess_ceiling'])
        self.assertEqual(capacity.status([value]), 'preparation not feasible')

    def test_each_independent_gate_and_aggregate_failure(self):
        # Actual three-attempt and one-attempt limits are algebraically equivalent.
        # Handcrafted per-branch inputs plant each boolean independently, including
        # a subprocess-only failure that cannot occur for the 1668-revision point.
        passing = capacity.gates(1, 3, 1)
        for key, arguments in (
                ('attempt_fits', (14*86400+1, 3, 1)),
                ('three_attempts_fit', (1, 42*86400+1, 1)),
                ('largest_revision_fits_subprocess_ceiling', (1, 3, 3601))):
            with self.subTest(key=key):
                value = capacity.gates(*arguments)
                self.assertFalse(value[key])
                for other in capacity.GATES:
                    if other != key:
                        self.assertTrue(value[other])
                self.assertEqual(capacity.status([value]), 'preparation not feasible')
                self.assertEqual(capacity.status([passing, value, passing]), 'preparation not feasible')
        self.assertTrue(all(capacity.gates(14*86400, 42*86400, 3600)[key] for key in capacity.GATES))

    def test_typescript_shape_and_code_not_comment_padding(self):
        count = scale.TS_SPEC['source_files']
        self.assertGreaterEqual(count+2, 81369)
        self.assertEqual(count % 20, 0)
        paths = [scale.ts_path(i) for i in range(count)]
        self.assertEqual(len(set(paths)), count)
        tests = [p for p in paths if p.name.endswith('.test.ts')]
        self.assertEqual(len(tests), count//2)
        self.assertEqual(sum(p.parts[0] == 'tests' for p in tests), count//4)
        self.assertTrue(all(len(p.parts) >= 4 and p.suffix == '.ts' for p in paths))
        for position, minimum in ((0, 1050000), (4, 4800), (count-1, 4800)):
            data = scale.ts_source(position, 0)
            self.assertGreaterEqual(len(data), minimum)
            self.assertIn(b'import ', data)
            self.assertIn(b'export function ', data)
            self.assertNotIn(b'/*', data)
            self.assertEqual(len(data), len(scale.ts_source(position, 2)))


def run():
    result = unittest.TextTestRunner(verbosity=2).run(unittest.defaultTestLoader.loadTestsFromTestCase(CapacityTests))
    if not result.wasSuccessful():
        raise AssertionError('capacity v8 selftest failed')


if __name__ == '__main__':
    run()
