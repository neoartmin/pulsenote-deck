#!/usr/bin/env python3
"""Generate one shared 512px screenshot action icon for Windows and Android.

Use the same high-quality transparent PNG pipeline as the 11 existing
action icons, not the old Win32/Android primitive rectangle renderer.
"""
from pathlib import Path
import sys
from PIL import Image, ImageDraw, ImageFilter, ImageChops, ImageOps

ROOT = Path(__file__).resolve().parent.parent
DEFAULT_TARGETS = (
    ROOT / "Windows" / "소스" / "app" / "assets" / "capture.png",
    ROOT / "android-pn12" / "app" / "src" / "main" / "assets"
         / "action_icons" / "capture.png",
)


def render_icon():
    scale = 3
    size = 512 * scale
    c = lambda v: round(v * scale)
    rect = (c(57), c(57), c(455), c(455))
    radius = c(86)
    mask = Image.new("L", (size, size), 0)
    ImageDraw.Draw(mask).rounded_rectangle(rect, radius=radius, fill=255)

    background = ImageOps.colorize(
        Image.linear_gradient("L").resize((size, size)),
        black="#291647", white="#9863ed"
    ).convert("RGBA")
    background.putalpha(mask)

    shine = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    hi = ImageDraw.Draw(shine)
    hi.ellipse((c(-5), c(-188), c(542), c(305)), fill=(249, 227, 255, 23))
    hi.rounded_rectangle((c(75), c(68), c(437), c(172)),
                         radius=c(70), fill=(249, 228, 255, 16))
    shine.putalpha(ImageChops.multiply(shine.getchannel("A"), mask))
    background = Image.alpha_composite(background, shine)

    glow = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    ImageDraw.Draw(glow).rounded_rectangle(
        (c(59), c(59), c(453), c(453)), radius=c(85),
        outline=(172, 115, 255, 145), width=c(12))
    background = Image.alpha_composite(
        background, glow.filter(ImageFilter.GaussianBlur(c(11))))

    d = ImageDraw.Draw(background)
    d.rounded_rectangle(rect, radius=radius,
                        outline=(187, 155, 251, 250), width=c(9))
    d.rounded_rectangle((c(69), c(68), c(443), c(442)),
                        radius=c(82), outline=(249, 230, 255, 80),
                        width=c(4))
    white = (253, 250, 255, 255)
    shadow = (55, 22, 116, 190)

    def stroke(points, width, fill):
        pts = [(c(x), c(y)) for x, y in points]
        d.line(pts, fill=fill, width=c(width), joint="curve")
        rr = c(width / 2)
        for x, y in pts:
            d.ellipse((x - rr, y - rr, x + rr, y + rr), fill=fill)

    for left, top, dx, dy in (
        (129, 150, 1, 1), (383, 150, -1, 1),
        (129, 378, 1, -1), (383, 378, -1, -1),
    ):
        corner = [(left + dx * 28, top),
                  (left, top), (left, top + dy * 28)]
        stroke([(x + 2, y + 3) for x, y in corner], 18, shadow)
        stroke(corner, 18, white)

    d.rounded_rectangle((c(171), c(221), c(342), c(343)),
                        radius=c(26), outline=shadow, width=c(29))
    d.rounded_rectangle((c(170), c(218), c(342), c(340)),
                        radius=c(26), outline=white, width=c(23))
    stroke([(213, 219), (229, 195), (285, 195), (301, 219)], 21, white)
    d.ellipse((c(219), c(239), c(293), c(313)),
              outline=white, width=c(19))
    d.ellipse((c(236), c(256), c(250), c(270)),
              fill=(233, 218, 255, 170))

    icon = background.resize((512, 512), Image.Resampling.LANCZOS)
    if icon.getpixel((0, 0))[3] != 0:
        raise RuntimeError("Icon lost transparent padding")
    return icon


def main():
    icon = render_icon()
    targets = [Path(arg) for arg in sys.argv[1:]] or DEFAULT_TARGETS
    for target in targets:
        target.parent.mkdir(parents=True, exist_ok=True)
        icon.save(target, format="PNG", optimize=True, compress_level=9)
        print(f"{target}: {target.stat().st_size} bytes, 512x512 RGBA")


if __name__ == "__main__":
    main()
