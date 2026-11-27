from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Protocol, overload


class Reader(Protocol):
    def read(self) -> str: ...


class Store(ABC):
    @abstractmethod
    def save(self, value: str) -> None:
        pass


@overload
def parse(value: str) -> str: ...
@overload
def parse(value: int) -> int: ...
def parse(value: str | int) -> str | int:
    return value
