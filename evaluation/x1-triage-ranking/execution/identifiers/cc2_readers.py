"""Finite-schema readers: excluded values are scanned, never decoded/materialised."""
import json
import re

SKIP = object()
ATOM = object()
NUMBER = re.compile(rb'-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?')


class SelectiveJSON:
    def __init__(self, raw):
        self.raw, self.pos = raw, 0

    def whitespace(self):
        while self.pos < len(self.raw) and self.raw[self.pos] in b' \t\r\n':
            self.pos += 1

    def take(self, token):
        self.whitespace()
        if self.raw[self.pos:self.pos + len(token)] != token:
            raise ValueError('unexpected JSON structure')
        self.pos += len(token)

    def string_token(self, capture=True):
        self.take(b'"')
        start = self.pos - 1
        while self.pos < len(self.raw):
            char = self.raw[self.pos]
            self.pos += 1
            if char == 34:
                return self.raw[start:self.pos] if capture else None
            if char < 32:
                raise ValueError('invalid JSON string framing')
            if char == 92:
                if self.pos >= len(self.raw):
                    break
                escape = self.raw[self.pos]
                self.pos += 1
                if escape == 117:
                    digits = self.raw[self.pos:self.pos + 4]
                    if len(digits) != 4 or any(c not in b'0123456789abcdefABCDEF' for c in digits):
                        raise ValueError('invalid JSON escape')
                    self.pos += 4
                elif escape not in b'"\\/bfnrt':
                    raise ValueError('invalid JSON escape')
        raise ValueError('unterminated JSON string')

    def value(self, spec, depth=0):
        if depth > 128:
            raise ValueError('JSON nesting limit')
        self.whitespace()
        char = self.raw[self.pos:self.pos + 1]
        if not char:
            raise ValueError('missing JSON value')
        if spec is SKIP:
            # Structural scan only: no json.loads, numeric conversion or string
            # decoding on the excluded value or any of its descendants.
            if char == b'{':
                self.take(b'{')
                self.whitespace()
                if self.raw[self.pos:self.pos + 1] != b'}':
                    while True:
                        self.string_token(False)
                        self.take(b':')
                        self.value(SKIP, depth + 1)
                        self.whitespace()
                        if self.raw[self.pos:self.pos + 1] != b',':
                            break
                        self.pos += 1
                self.take(b'}')
            elif char == b'[':
                self.take(b'[')
                self.whitespace()
                if self.raw[self.pos:self.pos + 1] != b']':
                    while True:
                        self.value(SKIP, depth + 1)
                        self.whitespace()
                        if self.raw[self.pos:self.pos + 1] != b',':
                            break
                        self.pos += 1
                self.take(b']')
            else:
                self.scalar(False)
            return None
        if isinstance(spec, dict):
            self.take(b'{')
            result, seen = {}, set()
            self.whitespace()
            if self.raw[self.pos:self.pos + 1] != b'}':
                while True:
                    key = json.loads(self.string_token())
                    if key not in spec or key in seen:
                        raise ValueError('unknown or duplicate schema key')
                    seen.add(key)
                    self.take(b':')
                    value = self.value(spec[key], depth + 1)
                    if spec[key] is not SKIP:
                        result[key] = value
                    self.whitespace()
                    if self.raw[self.pos:self.pos + 1] != b',':
                        break
                    self.pos += 1
            self.take(b'}')
            return result
        if isinstance(spec, list):
            self.take(b'[')
            result = []
            self.whitespace()
            if self.raw[self.pos:self.pos + 1] != b']':
                while True:
                    result.append(self.value(spec[0], depth + 1))
                    self.whitespace()
                    if self.raw[self.pos:self.pos + 1] != b',':
                        break
                    self.pos += 1
            self.take(b']')
            return result
        if spec is not ATOM or char in (b'{', b'['):
            raise ValueError('unexpected identifier structure')
        return self.scalar(True)

    def scalar(self, materialise):
        self.whitespace()
        start = self.pos
        if self.raw[self.pos:self.pos + 1] == b'"':
            token = self.string_token(materialise)
        else:
            match = NUMBER.match(self.raw, self.pos)
            if match:
                self.pos = match.end()
            else:
                for token in (b'null', b'true', b'false'):
                    if self.raw.startswith(token, self.pos):
                        self.pos += len(token)
                        break
                else:
                    raise ValueError('invalid JSON scalar framing')
            token = self.raw[start:self.pos] if materialise else None
        return json.loads(token) if materialise else None

    def document(self, spec):
        result = self.value(spec)
        self.whitespace()
        if self.pos != len(self.raw):
            raise ValueError('trailing JSON bytes')
        return result


def json_records(raw, schema, jsonl):
    if jsonl:
        for index, line in enumerate(raw.splitlines()):
            if not line.strip():
                raise ValueError('blank JSONL record')
            yield index, SelectiveJSON(line).document(schema)
    else:
        for index, row in enumerate(SelectiveJSON(raw).document([schema])):
            yield index, row


def csv_cells(raw):
    """Yield raw cell spans; quoted escapes are decoded only after selection."""
    pos, row, start = 0, [], 0
    while pos < len(raw):
        start = pos
        if raw[pos:pos + 1] == b'"':
            pos += 1
            while pos < len(raw):
                if raw[pos:pos + 1] == b'"':
                    pos += 1
                    if raw[pos:pos + 1] != b'"':
                        break
                    pos += 1
                else:
                    pos += 1
            else:
                raise ValueError('unterminated CSV quoted field')
            if raw[pos:pos + 1] not in (b',', b'\r', b'\n', b''):
                raise ValueError('unexpected CSV structure')
        else:
            while pos < len(raw) and raw[pos] not in b',\r\n':
                if raw[pos] == 34:
                    raise ValueError('unexpected CSV quote')
                pos += 1
        row.append((start, pos))
        sep = raw[pos:pos + 1]
        pos += 1
        if sep == b',':
            if pos == len(raw):
                row.append((pos, pos))
                yield row
            continue
        if sep == b'\r' and raw[pos:pos + 1] == b'\n':
            pos += 1
        yield row
        row = []


def csv_decode(raw, span):
    value = raw[span[0]:span[1]]
    if value.startswith(b'"'):
        value = value[1:-1].replace(b'""', b'"')
    return value.decode('utf-8', errors='strict')


def csv_records(raw, allowed_columns, selected):
    rows = iter(csv_cells(raw))
    try:
        header = [csv_decode(raw, cell) for cell in next(rows)]
    except StopIteration:
        raise ValueError('empty CSV') from None
    if len(header) != len(set(header)) or set(header) != set(allowed_columns):
        raise ValueError('unknown CSV schema')
    for index, spans in enumerate(rows):
        if len(spans) != len(header):
            raise ValueError('CSV column count mismatch')
        yield index, {key: csv_decode(raw, cell) for key, cell in zip(header, spans) if key in selected}
