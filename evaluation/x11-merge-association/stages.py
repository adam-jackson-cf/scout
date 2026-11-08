#!/usr/bin/env python3
"""X11a scoped authorisation and fail-closed driver. No experiment implementation."""
import argparse
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
PREREQUISITES = {'inventory': [], 'map': ['inventory'], 'catalogue-check': ['map'],
    'replay': ['catalogue-check'], 'label': ['replay'], 'analyse': ['label']}
IMPLEMENTED = {'inventory'}
REVIEWED = ('PREREGISTRATION.md', 'inventory.py', 'stages.py', 'selftest.py',
    'prepare_exports.py', 'extract_exposures.py', 'EXPORT_RECIPES.md', 'exposure-source-index.json',
    'frontend/main.go', 'frontend/main_test.go', 'assemble_preparation.py', 'prep_selftest.py',
    'pin_tool_closure.py', 'oc_v5.py', 'oc_v5.go', 'oc-v5-spec.json', 'oc-v5-fixtures.json',
    'oc-v5-timing-final-r2.json', 'oc-v5-timing-10000-r2.json', 'oc-v5-timing-stress-r2.json',
    'bootstrap_v6.py', 'bootstrap_selftest_v6.py', 'oc_selftest_v6.py', 'prep_scale_v6.py',
    'prep-scale-assumptions-v6.json', 'prep-scale-v6.json', 'prep-scale-projection-v6.json',
    'prep_scale_v6b.py', 'prep-scale-construct-v6b.json', 'prep-scale-v6b.json',
    'prep-scale-projection-v6b.json',
    'oc-v6-timing-metadata.json', 'oc-v6-timing-balanced-r3.json',
    'oc-v6-timing-concentrated-r3.json', 'oc-v6-timing-singleton-r3.json')


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def verify_files(entries, base):
    if not entries:
        raise ValueError('empty reviewed file set')
    covered = set()
    for entry in entries:
        from inventory import validate_pin
        validate_pin(entry)
        p = Path(entry['path'])
        p = (p if p.is_absolute() else base / p).resolve()
        if sha(p) != entry['sha256']:
            raise ValueError('reviewed digest mismatch: ' + str(p))
        covered.add(str(p))
    return covered


def verify_authorization(path, inputs, scope='protocol', artifacts=None):
    path, inputs = Path(path).resolve(), Path(inputs).resolve()
    record = json.loads(path.read_text())
    from inventory import exact_keys, validate_inputs, validate_artifacts
    required_fields = {'verdict', 'scope', 'review', 'files'}
    if scope == 'protocol':
        required_fields.add('reviewed_inventory')
    exact_keys(record, required_fields)
    if 'reviewed_inventory' in record:
        exact_keys(record['reviewed_inventory'], {'authorization', 'seal'})
        for entry in record['reviewed_inventory'].values():
            exact_keys(entry, {'path', 'sha256'})
    if record.get('verdict') != 'ZERO ISSUES' or record.get('scope') != scope:
        raise ValueError('missing zero-issue authorisation for scope')
    verify_files([record['review']], path.parent)
    covered = verify_files(record['files'], path.parent)
    required = {str((ROOT / name).resolve()) for name in REVIEWED}
    required.add(str(inputs))
    config = json.loads(inputs.read_text())
    validate_inputs(config)
    tools = list(config['export_preparation']['tools'].values())
    required.update(verify_files(tools, inputs.parent))
    sources = config.get('recipe_sources', [])
    required.update(verify_files(sources, inputs.parent))
    prep = config['export_preparation']
    required.update(verify_files([prep['clawpatch_tree'], prep['frontend_build_recipe'],
        prep['python_runtime'], prep['exposure_extraction']['seal'],
        *prep['exposure_extraction']['files']], inputs.parent))
    if artifacts is not None:
        artifacts = Path(artifacts).resolve()
        required.add(str(artifacts))
        generated = json.loads(artifacts.read_text())
        validate_artifacts(generated, [r['repository'] for r in config['repositories']])
        if generated['status'] != 'SEALED' or generated['source_inputs_sha256'] != sha(inputs):
            raise ValueError('stale export source seal')
        preparation = Path(generated['preparation_authorization']['path'])
        preparation = preparation if preparation.is_absolute() else artifacts.parent / preparation
        verify_files([generated['preparation_authorization']], artifacts.parent)
        verify_authorization(preparation, inputs, scope='export-preparation')
        verify_files([generated['exposure'], generated['unscoped_exposure'],
            generated['repository_exposure'], generated['frontend'],
            *generated['preparation_seals'], *[union['seal'] for union in generated['ownership'].values()],
            *generated['frontend_build'].values()], artifacts.parent)
        from inventory import verify_artifact_bindings
        verify_artifact_bindings(generated, inputs, artifacts.parent)
        if generated['frontend_build']['recipe'] != prep['frontend_build_recipe']:
            raise ValueError('frontend recipe differs from preparation')
        from inventory import load_pinned
        recipe = load_pinned(generated['frontend_build']['recipe'], artifacts.parent)
        if generated['frontend_build']['closure'] != recipe['closure']:
            raise ValueError('frontend closure differs from build recipe')
        if generated['frontend_build']['buildinfo']['path'] != recipe['required_buildinfo_path']:
            raise ValueError('frontend buildinfo path differs from recipe')
        argv = recipe['argv']
        if argv.count('-o') != 1 or generated['frontend']['path'] != argv[argv.index('-o') + 1]:
            raise ValueError('frontend executable path differs from recipe')
    elif scope == 'inventory':
        raise ValueError('inventory requires reviewed export artifact manifest')
    if not required <= covered:
        raise ValueError('authorisation omits protocol/recipe/tool/source/artifact bytes')
    return record


def verify_seal(path, stage, authority_sha, inputs=None):
    path = Path(path).resolve()
    record = json.loads(path.read_text())
    from inventory import exact_keys
    exact_keys(record, {'stage', 'authorization_sha256', 'inputs', 'argv', 'artifacts', 'predecessors'})
    if record['stage'] != stage or record['authorization_sha256'] != authority_sha:
        raise ValueError('wrong stage/authority seal')
    if (not isinstance(record['argv'], list) or not record['argv']
            or not all(isinstance(value, str) and value and '\0' not in value for value in record['argv'])):
        raise ValueError('invalid exact stage argv')
    for field in ('inputs', 'artifacts', 'predecessors'):
        if not isinstance(record[field], list):
            raise ValueError('invalid stage provenance collection')
    covered_inputs = verify_files(record['inputs'], path.parent)
    if inputs is not None and str(Path(inputs).resolve()) not in covered_inputs:
        raise ValueError('stage seal omits bound source inputs')
    verify_files(record['artifacts'], path.parent)
    required = PREREQUISITES.get(stage)
    if required is None or len(record['predecessors']) != len(required):
        raise ValueError('wrong exact predecessor count')
    for predecessor, expected_stage in zip(record['predecessors'], required):
        verify_files([predecessor], path.parent)
        predecessor_path = Path(predecessor['path'])
        predecessor_path = predecessor_path if predecessor_path.is_absolute() else path.parent / predecessor_path
        previous = json.loads(predecessor_path.read_text())
        if previous.get('stage') != expected_stage:
            raise ValueError('wrong predecessor stage')
        if expected_stage != 'inventory' and previous.get('authorization_sha256') != authority_sha:
            raise ValueError('wrong predecessor authority')
        verify_seal(predecessor_path, expected_stage, previous.get('authorization_sha256'), inputs)
    return record


def check_stage(stage, authorization, inputs, seals, artifacts=None):
    scope = 'inventory' if stage == 'inventory' else 'protocol'
    record = verify_authorization(authorization, inputs, scope, artifacts=artifacts)
    for predecessor in PREREQUISITES[stage]:
        if predecessor == 'inventory':
            # Later protocol explicitly adopts the earlier authority/output seal.
            adopted = record['reviewed_inventory']
            verify_files([adopted['authorization'], adopted['seal']], authorization.parent)
            original = Path(adopted['authorization']['path'])
            original = original if original.is_absolute() else authorization.parent / original
            sealpath = Path(adopted['seal']['path'])
            sealpath = sealpath if sealpath.is_absolute() else authorization.parent / sealpath
            verify_seal(sealpath, 'inventory', sha(original), inputs)
        else:
            verify_seal(seals / (predecessor + '.json'), predecessor, sha(authorization), inputs)
    if stage not in IMPLEMENTED:
        raise ValueError('NOT IMPLEMENTED: ' + stage + '; no outputs written')


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('stage', choices=PREREQUISITES)
    p.add_argument('--authorization', type=Path, required=True)
    p.add_argument('--inputs', type=Path, required=True)
    p.add_argument('--artifacts', type=Path)
    p.add_argument('--seals', type=Path, required=True)
    p.add_argument('--output', type=Path, required=True)
    args = p.parse_args()
    check_stage(args.stage, args.authorization, args.inputs, args.seals, args.artifacts)
    from inventory import inventory
    result = inventory(args.inputs, args.artifacts)
    with args.output.open('x') as stream:
        json.dump(result, stream, indent=2, sort_keys=True)
        stream.write('\n')
    # Deliberately unsealed: only a reviewed output manifest can adopt inventory.


if __name__ == '__main__':
    main()
