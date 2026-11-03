#!/usr/bin/env python3
"""Assert the shipped standard APK declares no accessibility or overlay access.

The extended edition intentionally preserves old opt-in advanced capabilities.
The standard edition must retain USB, screen capture, and notifications.
"""
from pathlib import Path
import xml.etree.ElementTree as ET

ROOT=Path(__file__).resolve().parent.parent
INTERMEDIATES=ROOT / "android-pn12" / "app" / "build" / "intermediates"
ANDROID="{http://schemas.android.com/apk/res/android}"
SENSITIVE="android.permission.BIND_ACCESSIBILITY_SERVICE"
OVERLAY="android.permission.SYSTEM_ALERT_WINDOW"
PROJECTION="android.permission.FOREGROUND_SERVICE_MEDIA_PROJECTION"

def find_merged(edition):
    candidates=[
        p for p in INTERMEDIATES.rglob("AndroidManifest.xml")
        if edition+"Debug" in str(p)
        and ("merged_manifest" in str(p) or "merged_manifests" in str(p))
    ]
    if not candidates:
        raise RuntimeError(f"No merged {edition}Debug manifest; build that APK first")
    return max(candidates,key=lambda p:p.stat().st_mtime)

def check(edition,expected_sensitive):
    filename=find_merged(edition)
    root=ET.parse(filename).getroot()
    permissions={p.attrib.get(ANDROID+"name") for p in root.findall("uses-permission")}
    services=root.findall("./application/service")
    accessibility=[
        s for s in services
        if s.attrib.get(ANDROID+"name","").endswith(".TextMacroAccessibilityService")
        or s.attrib.get(ANDROID+"permission")==SENSITIVE
    ]
    has_overlay=OVERLAY in permissions
    has_accessibility=bool(accessibility)
    if has_overlay!=expected_sensitive or has_accessibility!=expected_sensitive:
        raise RuntimeError(
            f"{edition} contains wrong sensitive features: "
            f"overlay={has_overlay}, accessibility={has_accessibility}: {filename}")
    if PROJECTION not in permissions:
        raise RuntimeError(f"{edition} lost screen capture foreground-service permission")
    if not any(s.attrib.get(ANDROID+"name","").endswith(".Pn12UsbService")
        for s in services):
        raise RuntimeError(f"{edition} lost USB connection service")
    if not any(s.attrib.get(ANDROID+"name","").endswith(".ScreenshotCaptureService")
        for s in services):
        raise RuntimeError(f"{edition} lost screenshot capture service")
    print(f"{edition}: overlay={has_overlay}, accessibility={has_accessibility}; "
          "USB and capture intact")

if __name__=="__main__":
    check("standard",False)
    check("extended",True)
