# PulseNote Deck — Windows · Android · Firmware

Single repository for **PN12**, **PN15 LEFT**, and **PN15 RIGHT**.
The three USB CDC devices are designed to operate concurrently (42 keys total).
Physical three-device verification is still pending.

## Source folders (original relative paths preserved)

```text
Windows/
  설치/PulseNoteDeck_v7.5.8_Setup.exe
  소스/app/                       Go Windows program, assets and tests
  소스/installer/                 Go Windows installer, assets and tests
  펌웨어/PN12/
  펌웨어/PN15_LEFT/
  펌웨어/PN15_RIGHT/
  README_KO.txt
  SHA256SUMS.txt
  검증보고서.txt

android-pn12/                     Full Android Java + Gradle source
Firmware/
  PN12/                           12-key ESP32-C3 sketch
  PN15/                           RIGHT 15-key ESP32-C3 sketch
  PN15_LEFT/                      LEFT 15-key ESP32-C3 sketch
.github/workflows/
  pn12-android-apk.yml            Android APK + full ZIP release
  pn12-pn15-windows.yml           Windows install EXE + full ZIP release
```

Windows original **v7.5.8 complete ZIP** was reassembled and verified by
GitHub Actions using SHA-256
`270f8839b092ca8726f159fd9474a440cb305d56cb687434cc1a2fb59b7dca25`.
All **119 original ZIP entries** were extracted into the `Windows/` tree.
One legacy Go test assertion was updated to match the new independent
USB device event queue; this does not change the runtime code.

## Download complete platform archives

- **Windows v7.5.8**:
  [Setup EXE, Go source, PN12/PN15 L/R firmware and verification data](https://github.com/neoartmin/pulsenote-deck/releases/tag/pnd-windows-v7.5.8)
- **Android v1.3.1**:
  [APK, Android source, PN12/PN15 L/R firmware and instructions](https://github.com/neoartmin/pulsenote-deck/releases/tag/pn12-pn15-android-v1.3.1)

## Firmware USB identities

| Board | ID | Keys |
|---|---|---:|
| PN12 | `PND:INFO:12:C3` | 12 |
| PN15 RIGHT | `PND:INFO:15:R:C3` | 15 |
| PN15 LEFT | `PND:INFO:15:L:C3` | 15 |

Legacy PN15 identifier `PND:INFO:15:C3` is treated as RIGHT.
Offline editing allows all three models to be configured before connection.
Hardware connection, reconnection, and concurrent inputs still require
real-device tests on Windows and Android.

## Build

Both platform workflows build independently using GitHub Actions.
Windows uses Go 1.23 on `windows-latest`; Android uses Java 17 / Gradle and
Android SDK 35. The published APK is debug-signed; it may not update an APK
signed with a different key.


## Single canonical firmware (PN12 / PN15 LEFT / PN15 RIGHT)

Only the three sketches in [Firmware/](Firmware/README.md) are authoritative.
The copies in `Windows/펌웨어/` and in
`Windows/소스/installer/firmware_*.ino` are **exact binary-identical mirrors**.
Windows Go `embed` includes these verified mirrored sketches in the installer;
Android full ZIP packages include `Firmware/` directly.

```sh
python tools/firmware_sync.py --check
python tools/firmware_sync.py --write  # regenerate Windows mirrors after firmware changes
python tools/firmware_sync.py --check
```

Both GitHub Actions workflows verify byte equality before building. If a
mirror drifts, **CI fails**, rather than silently publishing different
firmware for Windows and Android. The two PN15 sketches differ only in the
LEFT/RIGHT identifier and first comment; actual GPIO mapping and scan logic
remain identical.
