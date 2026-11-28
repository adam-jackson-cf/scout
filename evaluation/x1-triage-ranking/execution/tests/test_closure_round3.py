"""Round-3 native ambiguity and classifier branch regressions; no image loading."""
from pathlib import Path
import struct
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from closure.native import image_filetype


class RoundThreeTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.path = Path(temporary.name)/'opaque.data'

    def classify(self, data):
        self.path.write_bytes(data)
        return image_filetype(self.path)

    def refuses(self, data, message):
        with self.assertRaisesRegex(ValueError, message):
            self.classify(data)

    def thin(self, cpu=0x100000c, subtype=0, kind=6, count=0, command_bytes=0):
        return struct.pack('<8I', 0xfeedfacf, cpu, subtype, kind, count, command_bytes, 0, 0)

    def fat(self, entries, wide=False, endian='>'):
        magic = 0xcafebabf if wide else 0xcafebabe
        data = bytearray(struct.pack(endian+'II', magic, len(entries)))
        for cpu, subtype, offset, image, align, reserved in entries:
            values = (cpu, subtype, offset, len(image), align)
            if wide:
                values += (reserved,)
            data.extend(struct.pack(endian+('IIQQII' if wide else 'IIIII'), *values))
        for _, _, offset, image, _, _ in entries:
            if len(data) < offset+len(image):
                data.extend(b'\0'*(offset+len(image)-len(data)))
            data[offset:offset+len(image)] = image
        return bytes(data)

    def test_r301_45_slice_native_collision_never_skipped(self):
        entries = []
        for index in range(45):
            cpu = 0x100000c if index == 44 else index+1
            entries.append((cpu, 0, (index+1)*16384, self.thin(cpu), 14, 0))
        self.refuses(self.fat(entries), 'ambiguous Java/fat')

    def test_r301_java_52_0_and_45_3_remain_non_native(self):
        for minor, major in ((0, 52), (3, 45), (1, 52)):
            with self.subTest(minor=minor, major=major):
                # Complete tiny class: constant pool, class/super and zero members.
                data = (struct.pack('>IHHH', 0xcafebabe, minor, major, 5)
                        + b'\x01\x00\x04Test\x07\x00\x01\x01\x00\x10java/lang/Object\x07\x00\x03'
                        + struct.pack('>7H', 0x21, 2, 4, 0, 0, 0, 0))
                self.assertIsNone(self.classify(data))

    def test_r301_java_only_skipped_when_a_fat_slice_escapes_file(self):
        table = bytearray(struct.pack('>II', 0xcafebabe, 52)+b'\0'*(52*20))
        # The table fits; all preceding interpreted slices are bounded.
        struct.pack_into('>II', table, 8+51*20+8, len(table)+1, 0)
        self.assertIsNone(self.classify(table))

    def test_r301_bounded_fat_refuses_even_if_stricter_checks_would_fail(self):
        # All interpreted slices are inside the file; their zero sizes and
        # invalid thin headers must not be used as evidence that this is Java.
        self.refuses(struct.pack('>II', 0xcafebabe, 52)+b'\0'*(52*20), 'ambiguous Java/fat')

    def test_s4_non_native_null(self):
        self.assertIsNone(self.classify(b'not Mach-O'))

    def test_s4_truncated_fat_version(self):
        self.refuses(b'\xca\xfe\xba\xbe', 'truncated fat')

    def test_s4_invalid_architecture_counts(self):
        for count in (0, 33):
            with self.subTest(count=count):
                self.refuses(struct.pack('>II', 0xcafebabe, count), 'architecture count')

    def test_s4_invalid_slice_magic(self):
        self.refuses(self.fat([(0x100000c, 0, 256, b'X'*32, 8, 0)]), 'slice magic')

    def test_s4_invalid_thin_filetype_and_command_bounds(self):
        for data in (self.thin(kind=0), self.thin(kind=13),
                     self.thin(command_bytes=8), self.thin(count=1)):
            with self.subTest(data=data):
                self.refuses(data, 'header bounds/filetype')

    def test_s4_fat64_reserved_field(self):
        self.refuses(self.fat([(0x100000c, 0, 256, self.thin(), 8, 1)], True), 'reserved field')

    def test_s4_alignment_exponent_and_table_overlap(self):
        self.refuses(self.fat([(0x100000c, 0, 256, self.thin(), 64, 0)]), 'bounds/alignment')
        data = struct.pack('>II5I', 0xcafebabe, 1, 0x100000c, 0, 8, 32, 0)+b'\0'*32
        self.refuses(data, 'bounds/alignment')

    def test_s4_empty_slice(self):
        self.refuses(struct.pack('>II5I', 0xcafebabe, 1, 0x100000c, 0, 256, 0, 8).ljust(256, b'\0'), 'bounds/alignment')

    def test_s4_cpu_disagreement(self):
        self.refuses(self.fat([(0x1000007, 0, 256, self.thin(), 8, 0)]), 'CPU disagrees')

    def test_s4_overlapping_slices(self):
        self.refuses(self.fat([(0x100000c, 0, 256, self.thin(), 8, 0)]*2), 'overlapping')

    def test_s4_missing_or_duplicate_arm64(self):
        for cpus in ((0x1000007,), (0x100000c, 0x100000c)):
            entries = [(cpu, 0, (index+1)*256, self.thin(cpu), 8, 0)
                       for index, cpu in enumerate(cpus)]
            with self.subTest(cpus=cpus):
                self.refuses(self.fat(entries), 'no unique arm64')

    def test_s4_x86_and_arm64e_null_and_swapped_fat(self):
        entries = [(0x1000007, 3, 256, self.thin(0x1000007, 3), 8, 0),
                   (0x100000c, 2, 512, self.thin(subtype=2), 8, 0)]
        self.assertEqual(self.classify(self.fat(entries)), 6)
        self.assertEqual(self.classify(self.fat(entries, endian='<')), 6)


if __name__ == '__main__':
    unittest.main()
