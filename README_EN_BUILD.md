# PulseNote Deck v7.5.23 — English Edition

This distribution is a **translation-only variant** of the approved v7.5.23
Windows desktop application. The original Korean edition remains unchanged.

## What's translated

The Windows UI (sidebar, buttons, action chooser, dialogs, tooltips, key
descriptions, status notifications, errors), all action help, splash and
Windows setup/uninstall prompts. The PN12/PN15 firmware, icon artwork,
layout coordinates, DPIs, audio switching, auto-click and macro execution
code are not changed. Keyboard shortcuts use the same format.

## Compatibility

The program **intentionally shares the original data folder and application
identity** so existing profiles continue to work:
    %APPDATA%\PulseNoteDeck\config_unified.json
    %APPDATA%\PulseNoteDeck\device_notifications.json

**Before installing English Setup:** Exit PulseNote Deck completely from the
tray menu and back up %APPDATA%\PulseNoteDeck. The English setup updates the
same installation; it does not create a second independent installation.

The seven persisted Korean keycap color IDs (파랑, 빨강, 회색, 보라,
주황, 초록, 노랑) and legacy firmware filenames are **not translated**:
changing these would break saved configurations and firmware payloads.
Existing user-created Korean names and labels also remain as saved.

## Files

- PulseNoteDeck_v7.5.23_EN.exe: English app, standalone
- PulseNoteDeck_v7.5.23_EN_Setup.exe: English installer
- app/, installer/: translated Windows Go source used for these binaries
- Firmware/: original firmware, unmodified
- english_translation_catalogs/: Korean→English source string mappings

## Build reproducibility

On Windows with Go 1.23.x, Python 3.11, Pillow 11.3.0:

1. Run tools/generate_pnd_icon.py and tools/prepare_pretendard.py.
2. Add PE resources with akavel/rsrc v0.10.2 and app.manifest.
3. Run tools/build_english_version.py on an isolated v7.5.23 checkout.
4. Build app and installer using the English workflow under
   .github/workflows/windows-english-v7523.yml.

The English build intentionally does not alter application behavior,
including v7.5.23's existing background auto-click and keyboard-hold
policies. A passing source build is not a substitute for real Windows
device and mixed-DPI visual testing.

Source branch: feature/pnd-v7.5.23-english-only
Original Korean source: feature/pnd-auto-click-audio-notifications-dark-theme
