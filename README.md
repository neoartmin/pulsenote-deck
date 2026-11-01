# PulseNote Deck — PN12 · PN15 LEFT · PN15 RIGHT

Single repository for Windows, Android and ESP32-C3 firmware.

## Repository paths

```text
android-pn12/                   Android native application v1.3.1 (complete Java/Gradle source)
Firmware/
  PN12/                         PN12 12-key firmware
  PN15/                         PN15 right-hand 15-key firmware
  PN15_LEFT/                    PN15 left-hand 15-key firmware
.github/workflows/
  pn12-android-apk.yml          Android APK + source/firmware full ZIP release
  pn12-pn15-windows.yml         Windows build when Windows files are imported
tools/
  import-windows-v758.ps1       Exact Windows v7.5.8 ZIP importer
Windows/                        Full Windows package after ZIP import (see below)
  README_KO.txt
  SHA256SUMS.txt
  검증보고서.txt
  설치/PulseNoteDeck_v7.5.8_Setup.exe
  소스/app/                    Original Go source, icon assets and tests
  소스/installer/              Original Go installer source, icons and tests
  펌웨어/PN12/
  펌웨어/PN15_LEFT/
  펌웨어/PN15_RIGHT/
```

**Current import status (2026-11-01):** Android source and all three common firmware
sketches are committed in this branch. The Windows v7.5.8 original ZIP has
been inspected and verified locally, but its complete binary/source contents
are **not committed to GitHub yet**. Do not treat the presence of this README,
the import script or the Windows workflow as proof that Windows source is
already in the remote repository.

## Exact Windows source import

Original local Windows archive:

`PND_Windows_v7.5.8_PN12_PN15_3Devices_전체파일집.zip`

SHA-256:
`270f8839b092ca8726f159fd9474a440cb305d56cb687434cc1a2fb59b7dca25`

In a clone on branch `feature/pn15-grip` run:

```powershell
powershell -ExecutionPolicy Bypass -File .\tools\import-windows-v758.ps1 -ArchivePath "C:\path\to\PND_Windows_v7.5.8_PN12_PN15_3Devices_전체파일집.zip"
```

The importer copies **every** file, including the Windows installer EXE, Go
sources and 3 firmware sketches, into `Windows/` **without renaming any
path inside the original archive**, commits it, and pushes to the same branch.
The Windows workflow then runs tests and builds a Windows setup EXE and ZIP.

## USB identifiers

| Firmware | Identifier | Keys |
|---|---|---:|
| PN12 | `PND:INFO:12:C3` | 12 |
| PN15 RIGHT | `PND:INFO:15:R:C3` | 15 |
| PN15 LEFT | `PND:INFO:15:L:C3` | 15 |

Firmware determines physical device identity automatically, while the
offline editor supports all three profiles without attached hardware.
Hardware testing with three simultaneous devices is still required.
