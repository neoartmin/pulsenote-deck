#!/usr/bin/env python3
"""Validate newly added macro illustrations on each target.

The four shared actions are byte-compared by check_action_icon_sync.py.
All seven unique assets must also remain correctly sized transparent PNGs.
"""
from pathlib import Path
import struct

ROOT=Path(__file__).resolve().parent.parent
WIN=ROOT/"Windows"/"소스"/"app"/"assets"
DROID=ROOT/"android-pn12"/"app"/"src"/"main"/"assets"/"action_icons"
SETS={
    WIN:("web_search","compose_email","qr_code","sequence","clipboard_history","window_control"),
    DROID:("web_search","compose_email","qr_code","sequence","share_text"),
}
def main():
    for directory,names in SETS.items():
        for name in names:
            file=directory/(name+".png")
            data=file.read_bytes()
            if len(data)<36 or data[:8]!=b"\x89PNG\r\n\x1a\n":
                raise RuntimeError(f"Not a PNG: {file}")
            if data[12:16]!=b"IHDR":raise RuntimeError(f"Invalid IHDR: {file}")
            w,h,bit,ctype=struct.unpack(">IIBB",data[16:26])
            if (w,h,bit,ctype)!=(512,512,8,6):
                raise RuntimeError(f"Icon must be 512px RGBA: {file}, {(w,h,bit,ctype)}")
            print(f"OK {name} ({'Windows' if directory==WIN else 'Android'})")
    print("Advanced action illustrations validated")

if __name__=="__main__":main()
