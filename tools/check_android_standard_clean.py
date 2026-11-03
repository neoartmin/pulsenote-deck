#!/usr/bin/env python3
"""Reject a standard Android APK that still contains accessibility implementation.

Run after Gradle assembleStandardDebug and assembleExtendedDebug.
The standard edition must not package AccessibilityService, the edit-shortcut
parser or text-macro helper, even as unused code. The extended edition must keep
all three, preserving prior opt-in functionality.
"""
from pathlib import Path
from zipfile import ZipFile

ROOT=Path(__file__).resolve().parent.parent
APK_DIR=ROOT/"android-pn12"/"app"/"build"/"outputs"/"apk"
SENSITIVE_CLASSES=(
    b"TextMacroAccessibilityService",
    b"AndroidEditShortcut",
    b"TextMacro;",
)

def check(edition,expect_present):
    path=APK_DIR/edition/"debug"/f"app-{edition}-debug.apk"
    if not path.is_file():
        raise RuntimeError(f"Missing {edition} APK: {path}")
    with ZipFile(path) as zipfile:
        dex_names=[name for name in zipfile.namelist() if name.endswith(".dex")]
        if not dex_names:
            raise RuntimeError(f"Missing compiled .dex in {path}")
        dex=b"".join(zipfile.read(name) for name in dex_names)
        for needle in SENSITIVE_CLASSES:
            actual=needle in dex
            if actual!=expect_present:
                raise RuntimeError(
                    f"{edition}: {needle.decode('ascii')} found={actual}, "
                    f"expected={expect_present}"
                )
        xmls=[name for name in zipfile.namelist() if "text_macro_accessibility" in name]
        if bool(xmls)!=expect_present:
            raise RuntimeError(
                f"{edition}: accessibility XML present={bool(xmls)}, "
                f"expected={expect_present}"
            )
    print(f"{edition}: accessibility implementation present={expect_present}; "
          f"APK contains {len(dex_names)} DEX file(s)")

if __name__=="__main__":
    check("standard",False)
    check("extended",True)
