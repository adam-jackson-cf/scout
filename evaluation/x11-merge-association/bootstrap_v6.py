#!/usr/bin/env python3
"""Stdlib-only study launcher; the caller supplies the reviewed config digest."""
import argparse
import hashlib
import json
from pathlib import Path
import runpy
import sys


def digest(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def verified_helper(config, directory):
    directory = Path(directory).resolve()
    # -B forbids writes, not timestamp-valid cache reads. Reject before import.
    if any(directory.rglob('__pycache__')) or any(directory.glob('*.pyc')):
        raise ValueError('stale study bytecode cache')
    helper = directory / 'pin_tool_closure.py'
    pins = [p['sha256'] for p in config['recipe_sources']
            if Path(p['path']).resolve() == helper]
    if len(pins) != 1 or digest(helper) != pins[0]:
        raise ValueError('reviewed runtime helper source drift')
    sys.path.insert(0, str(directory))
    import pin_tool_closure
    return pin_tool_closure


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--config', required=True)
    parser.add_argument('--config-sha256', required=True)
    parser.add_argument('--script', required=True)
    parser.add_argument('args', nargs=argparse.REMAINDER)
    args = parser.parse_args()
    if not sys.flags.isolated or not sys.dont_write_bytecode:
        raise ValueError('bootstrap requires -I -B')
    if digest(args.config) != args.config_sha256:
        raise ValueError('reviewed inventory-inputs digest mismatch')
    config = json.loads(Path(args.config).read_text())
    directory = Path(__file__).resolve().parent
    script = Path(args.script).resolve()
    if script.parent != directory:
        raise ValueError('study entrypoint outside reviewed module directory')
    pin = verified_helper(config, directory)
    prep = config.get('export_preparation', config)
    pin.verify_python_runtime(prep)
    pin.verify_preparation_closure(config)
    source_pins = {str(Path(p['path']).resolve()): p['sha256'] for p in config['recipe_sources']}
    if source_pins.get(str(script)) != digest(script):
        raise ValueError('reviewed entrypoint source drift')
    sys.argv = [str(script), *(args.args[1:] if args.args[:1] == ['--'] else args.args)]
    runpy.run_path(str(script), run_name='__main__')


if __name__ == '__main__':
    main()
