from __future__ import annotations

from collections.abc import Callable
from dataclasses import dataclass
from typing import Protocol

Formatter = Callable[[int], str]


class Sink(Protocol):
    def write(self, text: str) -> None: ...


@dataclass(frozen=True)
class Report:
    formatter: Formatter
    sink: Sink


def publish(report: Report, value: int) -> None:
    report.sink.write(report.formatter(value))
