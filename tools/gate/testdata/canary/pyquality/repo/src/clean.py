from __future__ import annotations

import json
from dataclasses import dataclass


@dataclass(frozen=True)
class Total:
    name: str
    amount: int


def totals(rows: list[dict[str, int]]) -> list[Total]:
    result: list[Total] = []
    for row in rows:
        for name, amount in row.items():
            result.append(Total(name, amount))
    return result


def render(rows: list[dict[str, int]]) -> str:
    return json.dumps([total.amount for total in totals(rows)])
