from __future__ import annotations

import sys


def main() -> int:
    sys.stdout.write("report\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
