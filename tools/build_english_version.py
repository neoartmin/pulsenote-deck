#!/usr/bin/env python3
"""Build-time English localization for PulseNote Deck v7.5.23.

Translates Go UI/help/notifications/installer string literals only. Immutable
storage keys, macro protocol, artwork, geometry and device operations remain
unchanged. Apply this script to an isolated checkout of the English branch
before compiling. Do not apply it to the Korean release branch.
"""
from __future__ import annotations
import argparse
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "Windows" / "소스"
CATALOGS = (
    "en_ui.tsv", "en_help.tsv", "en_status.tsv", "en_installer.tsv",
)
DO_NOT_TRANSLATE = {
    "맑은 고딕",  # Font face name, not a UI string.
    "파랑", "빨강", "회색", "보라", "주황", "초록", "노랑", # Persistent keycap color IDs.
    "버튼 ",     # Existing Korean default-key label detection in config migration.
    "ESP32_C3_배선.txt", # Firmware file name in the installer payload.
}
# Preserve Go string syntax and escaped characters.
STRINGS = re.compile(r'"(?:\\.|[^"\\])*"|' + chr(96) + r'[^' + chr(96) + r']*' + chr(96), re.S)
HANGUL = re.compile(r"[\uac00-\ud7a3]")
GO_TEST = "_test.go"

def load_catalog() -> dict[str, str]:
    mapping: dict[str, str] = {}
    for filename in CATALOGS:
        path = ROOT / "tools" / "locales" / filename
        with path.open("r", encoding="utf-8") as handle:
            for line_no, row in enumerate(handle, 1):
                row = row.rstrip("\r\n")
                if not row or row.startswith("#"):
                    continue
                if "\t" not in row:
                    raise ValueError(f"{path}:{line_no}: no tab separator")
                source, target = row.split("\t", 1)
                if not source or not target:
                    raise ValueError(f"{path}:{line_no}: empty translation")
                if HANGUL.search(target):
                    raise ValueError(f"{path}:{line_no}: target contains Hangul")
                old = mapping.get(source)
                if old is not None and old != target:
                    raise ValueError(f"{path}:{line_no}: conflicting translation for {source!r}")
                mapping[source] = target
    return mapping

def rewrite_go(content: str, mapping: dict[str, str]) -> tuple[str, list[str], int]:
    missing: set[str] = set()
    count = 0
    def change(match: re.Match[str]) -> str:
        nonlocal count
        original = match.group()
        literal = original[1:-1]
        if not HANGUL.search(literal) or literal in DO_NOT_TRANSLATE:
            return original
        translated = mapping.get(literal)
        if translated is None:
            missing.add(literal)
            return original
        count += 1
        return original[0] + translated + original[-1]
    result = STRINGS.sub(change, content)
    return result, sorted(missing), count

def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true",
                        help="report missing strings without editing source")
    parser.add_argument("--allow-incomplete", action="store_true",
                        help="use only during translation development")
    args = parser.parse_args()
    mapping = load_catalog()
    pending: dict[Path, str] = {}
    missing: dict[Path, list[str]] = {}
    total = 0
    for path in sorted(SOURCE.rglob("*.go")):
        if path.name.endswith(GO_TEST):
            continue
        original = path.read_text(encoding="utf-8")
        localized, unresolved, count = rewrite_go(original, mapping)
        if unresolved:
            missing[path] = unresolved
        if localized != original:
            pending[path] = localized
        total += count
    print(f"Loaded {len(mapping)} English translations; matched {total} Go literals")
    if missing:
        for path, items in missing.items():
            print(f"\n{path.relative_to(ROOT)} : {len(items)} untranslated:")
            for item in items:
                print("  " + repr(item))
        print(f"\n{sum(map(len, missing.values()))} untranslated literals across {len(missing)} files")
    if missing and not args.allow_incomplete:
        raise SystemExit(2)
    if not args.check:
        for path, content in pending.items():
            path.write_text(content, encoding="utf-8")
        print(f"English localized sources written: {len(pending)} files")

if __name__ == "__main__":
    main()
