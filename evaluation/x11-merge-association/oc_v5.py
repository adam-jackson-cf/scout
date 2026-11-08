"""Frozen X11a OC helpers and fail-closed launcher; no candidate access."""
import hashlib
import json
from pathlib import Path

RADICES = (12, 39, 2, 3, 2, 5, 2)


def scenario_index(q_index, truth_index, rho_index, crossarm_index, agecorr_index, missing_index, directory_index):
    value = 0
    for index, radix in zip((q_index, truth_index, rho_index, crossarm_index, agecorr_index, missing_index, directory_index), RADICES):
        if type(index) is not int or not 0 <= index < radix:
            raise ValueError('scenario coordinate out of range')
        value = value * radix + index
    return value


def scenario_seed(index, replicate, stream='study'):
    if not 0 <= index < 56160 or replicate < 0:
        raise ValueError('invalid scenario/replicate')
    raw = f'x11a-oc-v6\0{index}\0{replicate}\0{stream}'.encode()
    return int.from_bytes(hashlib.sha256(raw).digest()[:8], 'big')


def false_support(truth_by_arm, support_by_arm):
    return any(support_by_arm[arm] and min(metrics) < .8 for arm, metrics in truth_by_arm.items())


def boundary_support(truth_by_arm, support_by_arm):
    return any(support_by_arm[arm] and min(metrics) == .8 for arm, metrics in truth_by_arm.items())


def score_counts(correct, known_assigned, known_existing, unknown_assigned, unknown_unassigned):
    if min(correct, known_assigned, known_existing, unknown_assigned, unknown_unassigned) < 0:
        raise ValueError('negative count')
    if correct > min(known_assigned, known_existing):
        raise ValueError('correct exceeds known denominator')
    return {'worst_precision': [correct, known_assigned + unknown_assigned],
            'worst_recall': [correct, known_existing + unknown_assigned + unknown_unassigned],
            'best_precision': [correct + unknown_assigned, known_assigned + unknown_assigned],
            'best_recall': [correct + unknown_assigned, known_existing + unknown_assigned]}


def combined_bounds(block, directory, draws=399):
    import math
    if any(not isinstance(values, (list, tuple)) or any(
            type(value) not in (int, float) or not math.isfinite(value) or not 0 <= value <= 1
            for value in values) for values in (block, directory)):
        return 0., 1., False
    if len(block) < math.ceil(.95 * draws) or len(directory) < math.ceil(.95 * draws):
        return 0., 1., False
    block, directory = sorted(block), sorted(directory)
    return (min(block[int(.025 * (len(block)-1))], directory[int(.025 * (len(directory)-1))]),
            max(block[int(.975 * (len(block)-1))], directory[int(.975 * (len(directory)-1))]), True)


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def resolve_pin(entry, base):
    if set(entry) != {'path', 'sha256'}:
        raise ValueError('invalid exact pin')
    path = Path(entry['path'])
    path = (path if path.is_absolute() else base / path).resolve()
    if sha(path) != entry['sha256']:
        raise ValueError('pin digest mismatch')
    return path


def validate_vectors(vectors):
    if not isinstance(vectors, dict) or set(vectors) != {'segment_lengths', 'counts', 'directory_cases'}:
        raise ValueError('invalid exact vectors schema')
    lengths = vectors['segment_lengths']
    if not isinstance(lengths, list) or len(lengths) != 3 or any(
            type(length) is not int or not 1500 <= length <= 2000 for length in lengths):
        raise ValueError('invalid segment lengths')
    for field in ('counts', 'directory_cases'):
        if not isinstance(vectors[field], list) or len(vectors[field]) != 3:
            raise ValueError('invalid repository vector dimension')
        for repo in vectors[field]:
            if not isinstance(repo, list) or len(repo) != 2:
                raise ValueError('invalid age vector dimension')
    for repo in range(3):
        for age in range(2):
            counts = vectors['counts'][repo][age]
            cases = vectors['directory_cases'][repo][age]
            if not isinstance(counts, list) or len(counts) != lengths[repo] or any(type(n) is not int or n < 0 for n in counts):
                raise ValueError('invalid count vector')
            if not isinstance(cases, list) or len(cases) > 2600:
                raise ValueError('invalid directory cases')
            actual, last = [0] * lengths[repo], -1
            for case in cases:
                if not isinstance(case, dict) or set(case) != {'position', 'directory_hash'}:
                    raise ValueError('invalid directory case')
                pos, key = case['position'], case['directory_hash']
                if type(pos) is not int or not last <= pos < lengths[repo] or pos < 0:
                    raise ValueError('invalid directory position/order')
                if not isinstance(key, str) or len(key) != 64 or any(c not in '0123456789abcdef' for c in key):
                    raise ValueError('invalid directory hash')
                actual[pos] += 1
                last = pos
            if actual != counts:
                raise ValueError('counts/directory_cases mismatch')


def bind_oc_inputs(authorization, inputs, inventory, vectors, recipe, stages):
    authorization, inputs = Path(authorization).resolve(), Path(inputs).resolve()
    record = stages.verify_authorization(authorization, inputs, scope='protocol')
    adopted = record.get('reviewed_inventory')
    if not isinstance(adopted, dict) or set(adopted) != {'authorization', 'seal'}:
        raise ValueError('mandatory original reviewed_inventory absent')
    original = resolve_pin(adopted['authorization'], authorization.parent)
    sealpath = resolve_pin(adopted['seal'], authorization.parent)
    seal = stages.verify_seal(sealpath, 'inventory', sha(original), inputs)
    # Inventory authority originally reviewed the export manifest, not a later substitute.
    candidates = []
    for entry in seal['inputs']:
        path = resolve_pin(entry, sealpath.parent)
        if path != inputs:
            data = json.loads(path.read_text())
            if isinstance(data, dict) and data.get('status') == 'SEALED' and data.get('source_inputs_sha256') == sha(inputs):
                candidates.append(path)
    if len(candidates) != 1:
        raise ValueError('inventory seal must bind exactly one original export manifest')
    stages.verify_authorization(original, inputs, scope='inventory', artifacts=candidates[0])
    output = Path(inventory).resolve()
    if len(seal['artifacts']) != 1 or resolve_pin(seal['artifacts'][0], sealpath.parent) != output:
        raise ValueError('inventory must be original sealed sole output')
    data = json.loads(output.read_text())
    expected = data.get('oc_vectors')
    if expected is None:
        raise ValueError('inventory oc_vectors is null')
    actual = json.loads(Path(vectors).read_text())
    validate_vectors(expected)
    validate_vectors(actual)
    if actual != expected:
        raise ValueError('vectors differ from reviewed inventory oc_vectors')
    config = json.loads(inputs.read_text())
    recipe_path = Path(recipe).resolve()
    pins = config.get('recipe_sources', [])
    if not any(resolve_pin(entry, inputs.parent) == recipe_path for entry in pins):
        raise ValueError('OC recipe is not source-pinned')
    return {'vectors_sha256': sha(vectors), 'inventory_sha256': sha(output),
            'inventory_authorization_sha256': sha(original), 'inventory_seal_sha256': sha(sealpath),
            'protocol_authorization_sha256': sha(authorization)}


def main():
    """Build and run the pinned standalone OC (timing is synthetic only)."""
    import argparse
    import subprocess
    import sys
    parser = argparse.ArgumentParser(description=main.__doc__)
    for name in ('recipe', 'build-record', 'out'):
        parser.add_argument('--' + name, type=Path, required=True)
    parser.add_argument('--mode', choices=('timing', 'oc', 'sensitivity'), required=True)
    for name in ('vectors', 'inventory', 'inputs', 'authorization'):
        parser.add_argument('--' + name, type=Path)
    parser.add_argument('--bootstrap', type=int, default=399)
    parser.add_argument('--timing-replicates', type=int)
    parser.add_argument('--directory-profile', choices=('balanced', 'concentrated', 'singleton'))
    parser.add_argument('--segment-lengths', help='timing only: three comma-separated lengths1500..2000')
    args = parser.parse_args()
    if not sys.flags.isolated or not sys.dont_write_bytecode:
        raise ValueError('launch with pinned isolated Python -I -B')
    if 'pin_tool_closure' not in sys.modules:
        raise ValueError('launch through reviewed stdlib-only bootstrap_v6.py')
    if args.mode != 'timing' and (args.bootstrap != 399 or args.timing_replicates is not None or args.directory_profile is not None or args.segment_lengths is not None):
        raise ValueError('non-timing OC fixes bootstrap399 and forbids timing overrides')
    recipe = json.loads(args.recipe.read_text())
    for path, digest in recipe['sources'].items():
        if sha(path) != digest:
            raise ValueError('OC wrapper/source drift: ' + path)
    binding = {}
    if args.mode != 'timing':
        if not all((args.inputs, args.authorization, args.vectors, args.inventory)):
            raise ValueError('non-timing OC requires authority, inputs, inventory and vectors')
        import stages
        binding = bind_oc_inputs(args.authorization, args.inputs, args.inventory, args.vectors, args.recipe, stages)
    # The stdlib-only bootstrap must verify this helper before importing it.
    import pin_tool_closure as pin
    runtime = pin.load_pin(recipe['python_runtime'])
    pin.verify_python_runtime({'python_runtime': recipe['python_runtime'], 'tools': {'python': {'path': runtime['path'], 'sha256': runtime['sha256']}}})
    closure = pin.load_pin(recipe['closure'])
    pin.verify_go_closure(recipe, closure)
    if '-pgo=off' not in recipe['argv']:
        raise ValueError('OC recipe must disable automatic PGO')
    binary = Path(recipe['binary'])
    if binary.exists() or args.build_record.exists() or args.out.exists():
        raise ValueError('OC build/result paths must be fresh')
    binary.parent.mkdir(parents=True, exist_ok=True)
    for key in ('HOME', 'GOCACHE', 'GOTMPDIR'):
        Path(recipe['environment'][key]).mkdir(parents=True, exist_ok=True)
    subprocess.run(recipe['argv'], cwd=recipe['cwd'], env=recipe['environment'], check=True)
    buildinfo = subprocess.check_output(recipe['version_m_argv'], cwd=recipe['cwd'], env=recipe['environment']).decode()
    argv = [str(binary), '--mode', args.mode, '--source', recipe['source'], '--out', str(args.out), '--bootstrap', str(args.bootstrap)]
    if args.mode == 'timing':
        argv += ['--timing-replicates', str(args.timing_replicates or 20), '--directory-profile', args.directory_profile or 'balanced']
        argv += ['--segment-lengths', args.segment_lengths or '1500,1666,2000']
    if args.vectors:
        argv += ['--vectors', str(args.vectors)]
    record = {'schema': 'x11a-oc-v6-build-v1', 'recipe_sha256': sha(args.recipe), 'closure': recipe['closure'],
              'binary_sha256': sha(binary), 'build_argv': recipe['argv'], 'run_argv': argv,
              'environment': recipe['environment'], 'buildinfo': buildinfo, 'runtime': recipe['python_runtime'], **binding}
    with args.build_record.open('x') as stream:
        json.dump(record, stream, sort_keys=True, indent=2)
        stream.write('\n')
    subprocess.run(argv, cwd=recipe['cwd'], env=recipe['environment'], check=True)


if __name__ == '__main__':
    main()
