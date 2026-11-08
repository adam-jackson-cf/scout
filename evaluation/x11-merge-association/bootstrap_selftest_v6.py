"""Synthetic preimport trust fixtures; run only in the main integration check."""
import hashlib
import importlib.util
from pathlib import Path
import py_compile
import tempfile
import unittest

import bootstrap_v6


class BootstrapTrustTests(unittest.TestCase):
    def test_timestamp_valid_cache_never_imported(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            marker = root / 'IMPORTED'
            helper = root / 'pin_tool_closure.py'
            helper.write_text('from pathlib import Path\nPath(' + repr(str(marker)) + ').touch()\n')
            py_compile.compile(str(helper), doraise=True)
            cached = Path(importlib.util.cache_from_source(str(helper)))
            self.assertTrue(cached.exists())
            config = {'recipe_sources': [{'path': str(helper), 'sha256': hashlib.sha256(helper.read_bytes()).hexdigest()}]}
            with self.assertRaisesRegex(ValueError, 'stale study bytecode'):
                bootstrap_v6.verified_helper(config, root)
            self.assertFalse(marker.exists())

    def test_helper_digest_comes_from_reviewed_pin(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            marker = root / 'IMPORTED'
            helper = root / 'pin_tool_closure.py'
            helper.write_text('from pathlib import Path\nPath(' + repr(str(marker)) + ').touch()\n')
            config = {'recipe_sources': [{'path': str(helper), 'sha256': '0' * 64}]}
            with self.assertRaisesRegex(ValueError, 'reviewed runtime helper source drift'):
                bootstrap_v6.verified_helper(config, root)
            self.assertFalse(marker.exists())


if __name__ == '__main__':
    unittest.main()
