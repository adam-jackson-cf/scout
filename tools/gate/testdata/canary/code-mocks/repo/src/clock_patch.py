from unittest import mock

PATCHED = mock.patch("time.time", return_value=0)
