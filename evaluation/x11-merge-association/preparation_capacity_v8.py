#!/usr/bin/env python3
"""Final-PREP additive fit and frozen held-out TypeScript direct bound; no candidate access."""
import argparse
from datetime import datetime, timezone
from fractions import Fraction
import json
from pathlib import Path
import sys

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
from preparation_capacity_v7 import fit, pin, projected

SIZE_FACTS_SHA256 = '141c6764dee809dd0bb198d53e1a414c2f744b1403d9d99c1e282af530c8e037'
GATES = ('attempt_fits', 'three_attempts_fit', 'largest_revision_fits_subprocess_ceiling')


def load(path):
    return json.loads(Path(path).read_text())


def observations(paths):
    rows = []
    seen = set()
    final = pin(HERE/'prepare_exports.py')['sha256']
    for path in paths:
        receipt = load(path)
        if (receipt['status'] != 'completed' or receipt['synthetic_only'] is not True
                or receipt['preparation_script']['sha256'] != final
                or receipt['schema'] != 'x11a-preparation-scale-v8'):
            raise ValueError('three successful final-PREP v8 Go receipts required')
        fixture = receipt['fixture']
        if fixture['language'] != 'Go' or receipt['scenario'] in seen:
            raise ValueError('distinct Go design points required')
        seen.add(receipt['scenario'])
        rows.append({'receipt': pin(Path(path)), 'files': fixture['files_per_revision'],
                     'bytes': fixture['bytes_per_revision_min'], 'revisions': fixture['revision_count'],
                     'elapsed_seconds': receipt['elapsed_seconds']})
    if seen != {'construct', 'files', 'bytes'} or len(rows) != 3:
        raise ValueError('construct, files and bytes receipts required exactly once')
    return rows


def gates(attempt, three_attempts, largest_revision):
    return {'per_attempt_seconds': float(attempt), 'three_attempt_seconds': float(three_attempts),
            'largest_revision_seconds': float(largest_revision),
            'attempt_fits': attempt <= 14*86400,
            'three_attempts_fit': three_attempts <= 42*86400,
            'largest_revision_fits_subprocess_ceiling': largest_revision <= 3600}


def direct_bound(observed, prediction, revisions=1668, margin=Fraction(3, 2)):
    bound = max(Fraction(str(observed)), Fraction(str(prediction)))
    if bound < 0 or revisions != 1668:
        raise ValueError('invalid held-out direct bound')
    result = gates(margin*revisions*bound, 3*margin*revisions*bound, margin*bound)
    result.update(revision_count=revisions, margin=str(margin), bound_seconds_per_revision=float(bound),
                  rule='margin * 1668 * max(observed whole elapsed/revision, frozen additive prediction)')
    return result


def status(projections):
    return ('projected-feasible-not-authorised' if all(all(p[k] for k in GATES) for p in projections)
            else 'preparation not feasible')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go-receipts', nargs=3, type=Path, required=True)
    parser.add_argument('--heldout', type=Path, required=True)
    parser.add_argument('--prediction', type=Path, required=True)
    parser.add_argument('--prediction-sha256', required=True)
    args = parser.parse_args()
    rows = observations(args.go_receipts)
    coefficients = fit(rows)
    frozen = load(args.prediction)
    if pin(args.prediction)['sha256'] != args.prediction_sha256:
        raise ValueError('frozen prediction drift')
    if frozen['go_observations'] != rows or frozen['coefficients_exact'] != list(map(str, coefficients)):
        raise ValueError('prediction must precede held-out and use these final-PREP receipts without refit')
    heldout = load(args.heldout)
    if (heldout['scenario'] != 'heldout-ts' or heldout['prediction'] != pin(args.prediction)
            or heldout['preparation_script']['sha256'] != pin(HERE/'prepare_exports.py')['sha256']
            or heldout['fixture'] != frozen['fixture']):
        raise ValueError('held-out receipt does not match frozen final-PREP fixture')
    source = HERE/'preparation-capacity-v7.json'
    if pin(source)['sha256'] != SIZE_FACTS_SHA256:
        raise ValueError('preserved v7 size facts drift')
    repositories = []
    observed = heldout['elapsed_seconds']/heldout['fixture']['revision_count']
    prediction = Fraction(frozen['prediction_seconds_per_revision_exact'])
    for old in load(source)['repositories']:
        item = {k: old[k] for k in ('repository', 'anchors', 'segment_length', 'size_facts')}
        if item['repository'] == 'microsoft/TypeScript':
            if len(item['size_facts']) != 1668:
                raise ValueError('TypeScript revision count drift')
            item['projection'] = direct_bound(observed, prediction)
            item['projection_without_margin'] = direct_bound(observed, prediction, margin=Fraction(1))
        else:
            item['projection'] = projected(item['size_facts'], coefficients)
            item['projection_without_margin'] = projected(item['size_facts'], coefficients, Fraction(1))
        repositories.append(item)
    result = status([r['projection'] for r in repositories])
    if heldout['status'] != 'completed':
        result = 'preparation not feasible'
    report = {'schema': 'x11a-additive-preparation-capacity-v8', 'script': pin(Path(__file__)),
        'size_facts_source': pin(source), 'candidate_access': False, 'observations': rows,
        'prediction': pin(args.prediction), 'heldout': pin(args.heldout),
        'model': {'coefficients_exact': dict(zip(('a', 'b', 'c'), map(str, coefficients))),
                  'coefficients_seconds': dict(zip(('a', 'b', 'c'), map(float, coefficients))),
                  'margins': ['1', '3/2'], 'no_refit_or_rerun': True,
                  'limitations': 'Exact three-point fit, no confidence interval; Django Python mapper remains unexercised. TypeScript uses a held-out direct bound, not extrapolated Go coefficients.'},
        'heldout_observed_seconds_per_revision': observed,
        'heldout_observed_to_prediction_ratio': observed/float(prediction),
        'repositories': repositories, 'status': result,
        'budget': {'attempt_seconds': 14*86400, 'whole_preparation_seconds': 42*86400,
                   'subprocess_seconds': 3600, 'max_concurrent_repositories': 5},
        'recorded_utc': datetime.now(timezone.utc).isoformat()}
    print(json.dumps(report, sort_keys=True, indent=2))


if __name__ == '__main__':
    main()
