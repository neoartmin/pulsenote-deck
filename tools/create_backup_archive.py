#!/usr/bin/env python3
"""Archive the entire tracked PulseNote repository + generated assets and EXEs."""
import argparse
import hashlib
import subprocess
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
GENERATED = [
    "Windows/소스/app/assets/auto_click.png",
    "Windows/소스/app/assets/audio_output.png",
    "Windows/소스/app/assets/app_logo.png",
    "Windows/소스/app/PulseNoteDeck.ico",
    "Windows/소스/installer/icon.ico",
    "Windows/소스/app/rsrc_windows_amd64.syso",
    "Windows/소스/installer/rsrc_windows_amd64.syso",
    "Windows/소스/installer/app.bin",
]
README = """PulseNote Deck Control Center v7.5.23 - 전체 백업

이 압축파일에는 v7.5.23 프리뷰 개발 브랜치의 모든 Git 추적 파일,
빌드에 필요한 생성 아이콘/리소스, 앱 EXE 및 설치 EXE가 있습니다.
폰트 원본 파일은 포함하지 않습니다. tools/prepare_pretendard.py를 실행하면
공식 Pretendard 배포본을 해시 검증 후 다시 가져올 수 있습니다.
Android/펌웨어/웹사이트/과거 설치기/Windows 소스를 모두 보존합니다.

PROJECT/  전체 프로젝트 소스 및 부속 파일
BINARIES/  빌드한 프로그램과 Windows 설치 파일
BACKUP_SHA256.txt  모든 구성 파일의 SHA-256 목록

중요: 실제 사용자의 PC에 저장된 개인 프로필은 여기에 없습니다.
프로그램을 완전히 종료한 다음
PROJECT/tools/backup_user_settings.cmd 를 실행하세요.
%APPDATA%\\PulseNoteDeck 의 파일들을 날짜가 포함된
USER_SETTINGS_BACKUP_yyyyMMdd_HHmmss 폴더로 복사합니다.

개인 설정 파일:
%APPDATA%\\PulseNoteDeck\\config_unified.json
%APPDATA%\\PulseNoteDeck\\device_notifications.json

복원 시 프로그램을 완전히 종료하고 기존 설정을 별도 보관한 뒤,
백업한 USER_SETTINGS_BACKUP 폴더 내용물을
%APPDATA%\\PulseNoteDeck 로 복사하십시오.

설치: BINARIES/PulseNoteDeck_v7.5.23_Setup_preview.exe
코드 위치: PROJECT/Windows/소스/app, PROJECT/Windows/소스/installer
펌웨어: PROJECT/Firmware 및 PROJECT/Windows/펌웨어
Android: PROJECT/android-pn12
웹사이트: PROJECT/website
빌드 명령: PROJECT/.github/workflows/windows-source-check.yml
GitHub PR: https://github.com/neoartmin/pulsenote-deck/pull/15

이 아카이브는 Git 기록(.git)을 제외한 소스 스냅샷이며,
실물 USB 장치에서의 모든 동작을 보증하는 정식 릴리스는 아닙니다.
"""

def digest(raw):
    return hashlib.sha256(raw).hexdigest()

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--app", type=Path, required=True)
    parser.add_argument("--installer", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    for exe in (args.app, args.installer):
        with exe.open("rb") as handle:
            if handle.read(2) != b"MZ":
                raise RuntimeError("Invalid Windows executable: " + str(exe))
    tracked = subprocess.check_output(["git", "ls-files", "-z"], cwd=ROOT)
    names = [x.decode("utf-8") for x in tracked.split(b"\0") if x]
    names = sorted(set(names) | set(GENERATED))
    commit = subprocess.check_output(["git", "rev-parse", "HEAD"],
                                     cwd=ROOT, text=True).strip()
    rows = []
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(args.output, "w", zipfile.ZIP_DEFLATED,
                         compresslevel=6, allowZip64=True) as archive:
        def add(entry, data):
            archive.writestr(entry, data)
            rows.append((digest(data), entry, len(data)))
        for name in names:
            path = ROOT / name
            if not path.is_file():
                raise RuntimeError("Missing source file: " + name)
            add("PROJECT/" + Path(name).as_posix(), path.read_bytes())
        add("BINARIES/PulseNoteDeck_v7.5.23.exe", args.app.read_bytes())
        add("BINARIES/PulseNoteDeck_v7.5.23_Setup_preview.exe",
            args.installer.read_bytes())
        add("README_BACKUP_KO.txt", ("SOURCE_COMMIT: " + commit
                                    + "\n\n" + README).encode("utf-8"))
        manifest = (f"PulseNote Deck v7.5.23 BACKUP - Git {commit}\n"
                    + f"Entries: {len(rows)}\n\n"
                    + "".join(f"{sha}  {name}  ({size} bytes)\n"
                              for sha, name, size in rows))
        archive.writestr("BACKUP_SHA256.txt", manifest.encode("utf-8"))
    with zipfile.ZipFile(args.output) as archive:
        if archive.testzip() is not None:
            raise RuntimeError("ZIP archive CRC verification failed")
        if len(archive.namelist()) != len(rows) + 1:
            raise RuntimeError("ZIP manifest entry mismatch")
    print("Full backup created", args.output, "files", len(rows) + 1,
          "bytes", args.output.stat().st_size, "commit", commit)

if __name__ == "__main__":
    main()
