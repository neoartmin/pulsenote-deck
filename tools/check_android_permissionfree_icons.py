#!/usr/bin/env python3
"""Validate all four Android-only permission-free action PNG resources.

PNG validation does not depend on Pillow at build time. Ensures image is 512
pixel RGBA, with regular PNG headers and valid 32-bit image dimensions.
"""
from pathlib import Path
import struct

ROOT=Path(__file__).resolve().parent.parent
ICONS=ROOT/"android-pn12"/"app"/"src"/"main"/"assets"/"action_icons"
REQUIRED=("copy_text","system_settings","open_timers","open_alarms")

def validate():
    for name in REQUIRED:
        file=ICONS/(name+".png")
        if not file.is_file():
            raise RuntimeError(f"Missing Android macro icon: {file}")
        with file.open("rb") as inp:
            header=inp.read(33)
        if len(header)<33 or header[:8]!=b"\x89PNG\r\n\x1a\n" or header[12:16]!=b"IHDR":
            raise RuntimeError(f"Not a PNG: {file}")
        width,height,depth,color=struct.unpack(">IIBB",header[16:26])
        if (width,height,depth,color)!=(512,512,8,6):
            raise RuntimeError(f"Icon not 512x512 RGBA: {file}, {(width,height,depth,color)}")
        print(f"OK {name}: 512x512 RGBA")
    print(f"All {len(REQUIRED)} permission-free action icons verified")

if __name__=="__main__": validate()
