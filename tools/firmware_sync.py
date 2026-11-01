#!/usr/bin/env python3
"""Synchronize deployment copies from the three authoritative Firmware/ sketches.

Run from any directory:
  python tools/firmware_sync.py --check  # CI, never modifies files
  python tools/firmware_sync.py --write  # refresh Windows firmware and embed copies
"""
from __future__ import annotations

import argparse
import hashlib
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
COPIES = {
    "PN12": (
        "Firmware/PN12/PulseNoteDeck_C3_12Key.ino",
        "Windows/펌웨어/PN12/PulseNoteDeck_PN12.ino",
        "Windows/소스/installer/firmware_pn12.ino",
    ),
    "PN15_RIGHT": (
        "Firmware/PN15/PulseNoteDeck_C3_15Key.ino",
        "Windows/펌웨어/PN15_RIGHT/PulseNoteDeck_PN15_RIGHT.ino",
        "Windows/소스/installer/firmware_pn15.ino",
    ),
    "PN15_LEFT": (
        "Firmware/PN15_LEFT/PulseNoteDeck_C3_15Key_LEFT.ino",
        "Windows/펌웨어/PN15_LEFT/PulseNoteDeck_PN15_LEFT.ino",
        "Windows/소스/installer/firmware_pn15_left.ino",
    ),
}
IDENTIFIERS = {
    "PN12": "PND:INFO:12:C3",
    "PN15_RIGHT": "PND:INFO:15:R:C3",
    "PN15_LEFT": "PND:INFO:15:L:C3",
}

def validate_firmware(model: str, data: bytes) -> str:
    code = data.decode("utf-8")
    identity = IDENTIFIERS[model]
    if code.count(identity) != 1:
        raise ValueError(f"{model}: missing/duplicate expected identity {identity}")
    for item in ("PND:READY:2.1", "PND:HELLO", "KEY:", "DEBOUNCE_MS"):
        if item not in code:
            raise ValueError(f"{model}: missing {item}")
    if model == "PN12":
        pin_sets = (r"ROWS\[4\]\s*=\s*\{0,\s*1,\s*6,\s*7\}",
                    r"COLS\[3\]\s*=\s*\{3,\s*4,\s*21\}")
    else:
        pin_sets = (r"ROWS\[\]\s*=\s*\{0,\s*1,\s*6\}",
                    r"COLS\[\]\s*=\s*\{3,\s*4,\s*21,\s*7,\s*5\}")
    if not all(re.search(pattern, code) for pattern in pin_sets):
        raise ValueError(f"{model}: GPIO matrix differs from approved pinout")
    return code

def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--check", action="store_true", help="fail on mismatched firmware copies")
    mode.add_argument("--write", action="store_true", help="replace copies with canonical bytes")
    args = parser.parse_args()
    failures = []
    variants = {}
    for model, paths in COPIES.items():
        source = ROOT / paths[0]
        if not source.is_file():
            failures.append(f"MISSING canonical firmware: {source}")
            continue
        data = source.read_bytes()
        try:
            variants[model] = validate_firmware(model, data)
        except (ValueError, UnicodeDecodeError) as error:
            failures.append(str(error))
            continue
        digest = hashlib.sha256(data).hexdigest()
        print(f"{model}: Firmware source SHA256 {digest}")
        for mirror in paths[1:]:
            target = ROOT / mirror
            if args.write:
                target.parent.mkdir(parents=True, exist_ok=True)
                if not target.is_file() or target.read_bytes() != data:
                    target.write_bytes(data)
                    print(f"  UPDATED {mirror}")
            elif not target.is_file() or target.read_bytes() != data:
                failures.append(f"OUT OF SYNC: {mirror} (expected {paths[0]})")
            else:
                print(f"  MATCH   {mirror}")
    if "PN15_LEFT" in variants and "PN15_RIGHT" in variants:
        def normalize(text: str) -> str:
            text = re.sub(r"^// PN15 GRIP (LEFT|RIGHT) hand firmware;[^\n]*\n", "", text)
            text = re.sub(r"PND:INFO:15:[LR]:C3", "PND:INFO:15:X:C3", text)
            return text
        if normalize(variants["PN15_LEFT"]) != normalize(variants["PN15_RIGHT"]):
            failures.append("LEFT and RIGHT PN15 firmware differ beyond identity and heading")
    for failure in failures:
        print("ERROR: " + failure, file=sys.stderr)
    if failures:
        return 1
    print("All PN12 / PN15 LEFT / PN15 RIGHT firmware bytes match.")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
