#!/usr/bin/env python3
"""Revision-8 heterogeneous-T complete/vertex differential; synthetic only."""
import argparse
import hashlib
import json
from pathlib import Path
import random
import subprocess
import tempfile
import sys
import ranking


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    if sys.version_info[:3] != (3, 11, 11):
        raise RuntimeError('X1 numerical validation requires CPython 3.11.11')
    if args.output.exists():
        raise FileExistsError('fixture evidence is create-once')
    here = Path(__file__).resolve().parent
    paths = [here/name for name in ('kernel_rounding.py', 'kernel_rounding_test.go', 'interval_diagnostic.go', 'ranking.py')]
    provenance = ranking.provenance(paths)
    provenance['python_executable'] = sys.executable
    provenance['python_executable_sha256'] = hashlib.sha256(Path(sys.executable).read_bytes()).hexdigest()
    provenance['go_version'] = subprocess.check_output(['go', 'version'], text=True).strip()
    checks = []

    def check(name, condition):
        if not condition:
            raise AssertionError(name)
        checks.append(name)

    with tempfile.TemporaryDirectory(prefix='x1-rounding-test-') as temporary:
        test_binary = Path(temporary)/'classification-test'
        subprocess.run(['go', 'test', '-c', '-o', str(test_binary),
                        str(here/'interval_diagnostic.go'), str(here/'kernel_rounding_test.go')], check=True)
        provenance['classification_test_binary_sha256'] = hashlib.sha256(test_binary.read_bytes()).hexdigest()
        completed = subprocess.run([str(test_binary), '-test.v'], capture_output=True, text=True, check=True)
        check('Go heterogeneous zero/nonzero variance classification',
              '--- PASS: TestHeterogeneousInexactRatioClassification' in completed.stdout)
    frame, repositories = [], []
    for repo in ('a', 'b', 'c'):
        totals = []
        for run in range(20):
            total = 10 if run % 2 == 0 else 30
            retrieved = total//5
            positive_positions = list(range(1, retrieved+1))+list(range(31, 31+total-retrieved))
            others = [i for i in range(1, 101) if i not in positive_positions]
            for issue, position in enumerate(positive_positions+others):
                frame.append(ranking.Unit(repo, f'{run:02}', f'{run}/{issue}', 100, position,
                                          issue+31 if issue < total else
                                          ([i for i in range(1, 101) if i not in range(31, 31+total)])[issue-total],
                                          True, False, int(issue < total), False, 1., True))
            totals.append((float(total), float(retrieved), 0.))
        repositories.append(totals)
    rng = random.Random(169)
    draws = [[rng.choices(range(20), k=19) for _ in range(1999)] for _ in repositories]
    expected = ranking.bootstrap_totals(repositories, draws)
    vertex = ranking.analyse(frame, seed=169, include_conditional=False)
    check('Python vertex/complete interval parity', vertex['interval'] == expected[1])
    classifications = []
    for values, repo_draws in zip(repositories, draws):
        for sample in [values]+[[values[i] for i in indices] for indices in repo_draws]:
            classifications.append(ranking.ratio_variance(sample) == 0.)
    check('Python constant-ratio zero variance classification', all(classifications))
    check('Python nonzero variance control',
          ranking.ratio_variance([(10., 2., 0.), (30., 6.000000000001, 0.)]) > 0)
    outputs = {}
    with tempfile.TemporaryDirectory(prefix='x1-rounding-') as temporary:
        binary = Path(temporary)/'kernel'
        subprocess.run(['go', 'build', '-o', str(binary), str(here/'interval_diagnostic.go')], check=True)
        provenance['binary_sha256'] = hashlib.sha256(binary.read_bytes()).hexdigest()
        for field, values in (('Repositories', repositories), ('Vertices', [[v] for v in repositories])):
            result = json.loads(subprocess.check_output([str(binary)], input=json.dumps({'Method': 'studentized', field: values, 'Draws': draws}), text=True))
            check(field+' byte-identical interval',
                  [float(x).hex() for x in result['Interval']] == [float(x).hex() for x in expected[1]])
            outputs[field] = result
    check('source provenance unchanged during supplement',
          ranking.provenance(paths)['script_sha256'] == provenance['script_sha256'])
    report = {'schema': 'x1-kernel-rounding-fixture/rev8', 'synthetic_only': True,
              'checks_passed': len(checks), 'checks': checks,
              'python_zero_variance_classifications': len(classifications),
              'python_complete': expected, 'python_vertex': vertex, 'go': outputs,
              'go_classification_test': completed.stdout, 'provenance': provenance}
    with args.output.open('x') as stream:
        json.dump(report, stream, indent=2, sort_keys=True)
        stream.write('\n')
    print(json.dumps({'checks_passed': report['checks_passed'], 'output': str(args.output)}))

if __name__ == '__main__':
    main()
