"""Deterministic synthetic-only size-source acceptance checks."""
import ast
from pathlib import Path
import unittest
from size_input import SizeInput, REPOSITORIES
from population import prepare_population
from fixed_population import prepare_population as original


class SizeTests(unittest.TestCase):
    def test_equal_vectors_preserve_exact_population(self):
        sizes = tuple([50, 60]*10)
        inputs = SizeInput({r: sizes for r in REPOSITORIES})
        self.assertEqual(prepare_population(size_input=inputs, seed=169, size_dependence=.05),
                         original(sizes=sizes, run_count=20, seed=169, size_dependence=.05))

    def test_asymmetric_complete_sizes_and_global_offsets(self):
        vectors = {r: [50+i*25]* (20+i) for i, r in enumerate(REPOSITORIES)}
        inputs = SizeInput(vectors)
        pop = prepare_population(size_input=inputs, size_dependence=.20)
        self.assertEqual(len(pop.runs), 63)
        for r in pop.runs:
            self.assertEqual(r.n, inputs.size(r.repo, int(r.run)))
            self.assertEqual(r.size_offset, inputs.offset(r.repo, int(r.run), .20))
        self.assertEqual(inputs.offset(REPOSITORIES[-1], 0, .20), .20)
        self.assertEqual([len(v) for v in inputs.measurement_sizes().values()], [20, 21, 22])

    def test_gate_draw_and_digest_determinism(self):
        good = SizeInput({r: [1000]*20 for r in REPOSITORIES})
        bad = SizeInput({r: [2000]*20 for r in REPOSITORIES})
        self.assertTrue(good.draw(169)[1])
        self.assertFalse(bad.draw(169)[1])
        self.assertEqual(good.draw(169), good.draw(169))
        self.assertEqual(good.sha256, SizeInput(dict(reversed(list(good.vectors.items())))).sha256)
        with self.assertRaises(ValueError):
            SizeInput({r: [True]*20 for r in REPOSITORIES})

    def test_only_size_source_changed_in_kernel(self):
        validation = Path(__file__).resolve().parents[2]/"validation/fixed_population.py"
        draft = Path(__file__).with_name("population.py").read_text()
        old = ast.parse(validation.read_text())
        new = ast.parse(draft)
        a = next(n for n in old.body if isinstance(n, ast.FunctionDef) and n.name == "prepare_population")
        b = next(n for n in new.body if isinstance(n, ast.FunctionDef) and n.name == "prepare_population")
        text = ast.unparse(b).replace('size_input,', 'run_count=20, sizes=(50,),')
        text = text.replace('range(len(size_input.vectors[repo]))', 'range(run_count)')
        text = text.replace('size_input.size(repo, run)', 'sizes[run % len(sizes)]')
        text = text.replace('size_input.offset(repo, run, size_dependence)', 'size_dependence * (n - min(sizes)) / max(1, max(sizes) - min(sizes))')
        self.assertEqual(ast.dump(ast.parse(ast.unparse(a))), ast.dump(ast.parse(text)))


if __name__ == "__main__":
    unittest.main()
