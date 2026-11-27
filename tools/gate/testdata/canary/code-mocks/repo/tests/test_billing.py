from unittest import mock

from billing import charge


def test_charge_returns_amount() -> None:
    with mock.patch("time.time", return_value=0):
        assert charge(3) == 3
