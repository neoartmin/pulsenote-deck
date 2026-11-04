#!/usr/bin/env python3
"""Generate the PulseNote Deck flat app icon (black / light gray / red).

Build-time only. Requires Pillow (pip install pillow).
The same ICO is embedded by the installer and used by the Windows UI.
"""
from pathlib import Path
from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parents[1]
APP_ICON = ROOT / "Windows" / "소스" / "app" / "PulseNoteDeck.ico"
SETUP_ICON = ROOT / "Windows" / "소스" / "installer" / "icon.ico"

SIZE = 1024
im = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
# No drop shadows, gradients, glow effects or small lettering.
d.rounded_rectangle((38, 38, 986, 986), radius=194,
                    fill="#111114", outline="#AFAFB6", width=26)
d.rounded_rectangle((91, 91, 933, 933), radius=147, fill="#18181B")
d.rounded_rectangle((127, 244, 193, 780), radius=31, fill="#D62839")
for j, y in enumerate((199, 414, 629)):
    for i, x in enumerate((259, 473, 687)):
        color = "#D62839" if (i, j) == (0, 0) else "#D9D9DE"
        d.rounded_rectangle((x, y, x + 162, y + 162),
                            radius=35, fill=color)

sizes = [(x, x) for x in (16, 24, 32, 48, 64, 128, 256)]
for path in (APP_ICON, SETUP_ICON):
    path.parent.mkdir(parents=True, exist_ok=True)
    im.save(path, format="ICO", sizes=sizes)
    print(f"Generated app/installer icon: {path.stat().st_size} bytes")
