#!/usr/bin/env python3
"""Render PulseNote Deck's flat black / light-grey / red desktop icon.

Uses only Python's standard library. Produces a multi-resolution ICO for
Windows shortcuts and a matching PNG preview. The design intentionally avoids
glows, embossed effects, decorative waves and microtext.
"""
from __future__ import annotations

from pathlib import Path
import struct
import zlib


ROOT = Path(__file__).resolve().parents[1]
SIZES = (16, 24, 32, 48, 64, 128, 256)


def rounded(x, y, left, top, right, bottom, radius):
    px = max(left + radius, min(x, right - radius))
    py = max(top + radius, min(y, bottom - radius))
    return (x - px) ** 2 + (y - py) ** 2 <= radius ** 2


def pixel(x, y):
    # Normalized 256px coordinates; black surround remains transparent.
    if not rounded(x, y, 9, 9, 247, 247, 45):
        return (0, 0, 0, 0)
    color = (18, 18, 20, 255)
    if rounded(x, y, 14, 14, 242, 242, 40):
        color = (24, 24, 27, 255)
    else:
        return (143, 143, 151, 255)
    # A red vertical marker distinguishes the desktop icon from generic grids.
    for row in range(3):
        for col in range(5):
            left = 34 + col * 39
            top = 69 + row * 41
            if rounded(x, y, left, top, left + 31, top + 31, 7):
                if (row, col) == (0, 0):
                    return (214, 40, 57, 255)
                return (217, 217, 222, 255)
    return color


def render_rgba(size):
    samples = 2 if size <= 64 else 1
    raw = bytearray()
    for y in range(size):
        for x in range(size):
            counts = [0, 0, 0, 0]
            for sy in range(samples):
                for sx in range(samples):
                    nx = (x + (sx + 0.5) / samples) * 256 / size
                    ny = (y + (sy + 0.5) / samples) * 256 / size
                    rgba = pixel(nx, ny)
                    for i in range(4):
                        counts[i] += rgba[i]
            raw.extend(value // (samples * samples) for value in counts)
    return bytes(raw)


def png_chunk(kind, data):
    return struct.pack('>I', len(data)) + kind + data + struct.pack(
        '>I', zlib.crc32(kind + data) & 0xffffffff)


def encode_png(size, rgba):
    rows = bytearray()
    for y in range(size):
        rows.append(0)
        rows.extend(rgba[y * size * 4: (y + 1) * size * 4])
    return (b'\x89PNG\r\n\x1a\n'
            + png_chunk(b'IHDR', struct.pack('>2I5B', size, size, 8, 6, 0, 0, 0))
            + png_chunk(b'IDAT', zlib.compress(bytes(rows), 9))
            + png_chunk(b'IEND', b''))


def main():
    pictures = [(size, encode_png(size, render_rgba(size))) for size in SIZES]
    directory = struct.pack('<HHH', 0, 1, len(pictures))
    offset = 6 + len(pictures) * 16
    entries = bytearray()
    content = bytearray()
    for size, payload in pictures:
        entries.extend(struct.pack('<BBBBHHII', 0 if size == 256 else size,
                                    0 if size == 256 else size,
                                    0, 0, 1, 32, len(payload), offset))
        content.extend(payload)
        offset += len(payload)
    ico = directory + entries + content
    for target in (ROOT / 'Windows/소스/app/PulseNoteDeck.ico',
                   ROOT / 'Windows/소스/installer/icon.ico'):
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(ico)
        print(target.relative_to(ROOT), len(ico), 'bytes')
    preview = ROOT / 'Windows/소스/app/PulseNoteDeck_icon_preview.png'
    preview.write_bytes(pictures[-1][1])


if __name__ == '__main__':
    main()
