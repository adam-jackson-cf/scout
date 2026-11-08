#!/usr/bin/env python3
"""Assemble an X11a artifact manifest exclusively from verified preparation seals."""
import argparse
import json
from pathlib import Path
from prepare_exports import (absolute, authorise, canonical_paths, canonical_seals,
                             digest, file_pin, load, pin, require,
                             require_no_global_stop, verify_preparation_seal)


def assemble(inputs, authorization):
    config, _, _, signed_file = authorise(inputs, authorization)
    signed_file(Path(__file__).resolve())
    require_no_global_stop(inputs)
    seals = canonical_seals(inputs)
    ownership, operations, operation_pins = {}, {}, []
    for seal in seals:
        entry = file_pin(seal)
        record = verify_preparation_seal(entry, config, inputs, authorization)
        require(seal == canonical_paths(inputs, record['operation'], record['repository'])[1],
                'preparation seal operation/repository differs from canonical path')
        require(not (record['status'] == 'refused' and record['refusal']['code'] in {
            'PreparationInterruptedOperational', 'PreparationBudgetExhausted'}),
            'X11a not testable (operational): study-wide interruption/deadline; no candidate exclusion or substitution')
        if record['operation'] == 'ownership':
            slug = record['repository']
            require(slug not in ownership, 'duplicate repository preparation')
            ownership[slug] = {'status': record['status'], 'seal': entry}
        else:
            name = record['operation']
            require(name not in operations, 'duplicate preparation operation')
            require(record['status'] == 'completed',
                    'X11a not testable (operational): terminal '+name+' preparation failure; no retry or substitution')
            operations[name] = load(pin(record['output']))
            operation_pins.append(entry)
    require(set(ownership) == {r['repository'] for r in config['repositories']},
            'missing repository success/refusal seal')
    require(set(operations) == {'exposures', 'build-frontend'}, 'missing preparation operation')
    result = {'status': 'SEALED', 'source_inputs_sha256': digest(inputs),
              'preparation_authorization': file_pin(authorization), 'ownership': ownership,
              'preparation_seals': operation_pins,
              **operations['exposures'], **operations['build-frontend']}
    from inventory import validate_artifacts, verify_artifact_bindings
    validate_artifacts(result, [r['repository'] for r in config['repositories']])
    verify_artifact_bindings(result, inputs, inputs.parent)
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inputs', required=True)
    parser.add_argument('--authorization', required=True)
    parser.add_argument('--output', required=True)
    args = parser.parse_args()
    output = absolute(args.output)
    require(not output.exists(), 'manifest is create-once')
    result = assemble(absolute(args.inputs), absolute(args.authorization))
    with output.open('x') as stream:
        json.dump(result, stream, sort_keys=True, indent=2)
        stream.write('\n')


if __name__ == '__main__':
    main()
