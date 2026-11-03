#!/usr/bin/env python3
"""Fail the build if Android's packaged action art diverges from Windows.

The 20 PNG files, including Android/Windows parity actions, must match byte-for-byte.
"""
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SOURCE = ROOT / "Windows" / "소스" / "app" / "assets"
ANDROID = ROOT / "android-pn12" / "app" / "src" / "main" / "assets" / "action_icons"
ICONS = (
    "none", "volume_up", "volume_down", "volume_mute",
    "media_play", "media_next", "media_prev", "website",
    "program", "shortcut", "folder", "capture",
    "copy_text", "system_settings", "open_timers", "open_alarms",
    "web_search", "compose_email", "qr_code", "sequence",
)

def main() -> int:
    failed = []
    for name in ICONS:
        source = SOURCE / (name + ".png")
        target = ANDROID / (name + ".png")
        if not source.is_file() or not target.is_file():
            failed.append(f"{name}: source or Android asset missing")
            continue
        expected = source.read_bytes()
        actual = target.read_bytes()
        if not expected.startswith(b"\x89PNG\r\n\x1a\n"):
            failed.append(f"{name}: not a PNG")
        elif expected != actual:
            failed.append(f"{name}: different bytes")
    for fail in failed:
        print(f"ERROR: {fail}")
    if failed:
        return 1
    print(f"OK: {len(ICONS)} shared PNG icons match across PC and Android")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
