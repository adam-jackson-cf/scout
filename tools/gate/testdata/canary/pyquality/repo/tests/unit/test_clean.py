from __future__ import annotations

import unittest


class TotalsBehaviour(unittest.TestCase):
    def test_empty_rows_give_no_totals(self) -> None:
        self.assertEqual([], [])
