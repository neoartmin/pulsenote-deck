#!/usr/bin/env python3
"""Download pinned original Pretendard OTFs for embedding in PulseNote Windows.

No files are installed into the OS. The EXE registers the bundled fonts
per-process via AddFontMemResourceEx. Git blob checks prevent drift.
Upstream: https://github.com/orioncactus/pretendard
License: SIL OFL 1.1, retained in installer/Pretendard_OFL.txt
"""
from pathlib import Path
from urllib.request import Request, urlopen
from hashlib import sha1

REV = "7aeb0698819be2b4097dae8ec8fe6a795e5cf3ae"
FONTS = {
    "Pretendard-Regular.otf": "08bf4cfc2164a0bff74a4bf844128d3b843bbf87",
    "Pretendard-Medium.otf": "057506983f0a8b0438fc5f24382db64497bf6eae",
    "Pretendard-Bold.otf": "8e5e30a28b4435359945cf08b261c86d0de8da53",
}

DEST = Path(__file__).resolve().parents[1] / "Windows" / "소스" / "app" / "assets"

def git_blob_sha(data: bytes) -> str:
    return sha1(b"blob " + str(len(data)).encode() + bytes((0,)) + data).hexdigest()

def main():
    DEST.mkdir(parents=True, exist_ok=True)
    for filename, expected_sha in FONTS.items():
        dst = DEST / filename
        if dst.is_file() and git_blob_sha(dst.read_bytes()) == expected_sha:
            print("Already present", filename)
            continue
        url = (
            "https://raw.githubusercontent.com/orioncactus/pretendard/"
            + REV + "/packages/pretendard/dist/public/static/" + filename
        )
        req = Request(url, headers={"User-Agent": "PulseNoteDeck-Font-Prep/1.0"})
        with urlopen(req, timeout=35) as response:
            content = response.read(8 * 1024 * 1024)
        sha = git_blob_sha(content)
        if sha != expected_sha:
            raise ValueError(f"Unexpected upstream content for {filename}: {sha}")
        dst.write_bytes(content)
        print("Verified bundled font", filename, len(content))

if __name__ == "__main__":
    main()
