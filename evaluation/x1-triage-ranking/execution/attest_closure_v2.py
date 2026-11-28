#!/usr/bin/env python3
"""Digest-guarded append-only entrypoint for the reviewed weak-link collector."""
from pathlib import Path
import sys

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import attest_closure as base
from closure import native_v2
from verify_inputs import pin

SIGNED_BASE = {
    'attest_closure.py': '113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892',
    'closure/native.py': 'c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb',
}
_SIGNED_OWN_SOURCES = base.own_sources


def own_sources():
    return [*_SIGNED_OWN_SOURCES(), pin(HERE/'attest_closure_v2.py'),
            pin(HERE/'closure/native_v2.py')]


def native_closure(files, commands, cwd, cache_root=None):
    return native_v2.collect(files, commands, cwd, cache_root, base.safe_resolve, base.pin)


def install():
    for name, sha in SIGNED_BASE.items():
        if base.digest(HERE/name) != sha:
            raise ValueError('signed closure base drift: '+name)
    base.native_closure = native_closure
    base.own_sources = own_sources


def main():
    install()
    return base.main()


if __name__ == '__main__':
    main()
