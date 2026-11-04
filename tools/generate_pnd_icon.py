#!/usr/bin/env python3
"""Build glossy PulseNote shortcut and new action artwork in the v7.5.17 icon style.

Existing assets and user color profiles remain untouched.
Generated assets are embedded in the executable by the Windows CI build.
"""
from pathlib import Path
from PIL import Image, ImageDraw, ImageFilter

ROOT = Path(__file__).resolve().parents[1]
APP = ROOT / "Windows" / "소스" / "app"
INSTALLER = ROOT / "Windows" / "소스" / "installer"
ART = APP / "assets"
S = 1024

def rgb(v):
    v = v.lstrip("#")
    return tuple(int(v[i:i + 2], 16) for i in (0, 2, 4))

def mix(a, b, t):
    return tuple(round(a[i] * (1 - t) + b[i] * t) for i in range(3))

def rr(d, rect, radius, fill=None, outline=None, width=1):
    d.rounded_rectangle(tuple(round(z * 2) for z in rect),
                        radius=round(radius * 2), fill=fill,
                        outline=outline, width=round(width * 2))

def ellipse(d, rect, fill=None, outline=None, width=1):
    d.ellipse(tuple(round(z * 2) for z in rect),
              fill=fill, outline=outline, width=round(width * 2))

def line(d, points, fill, width):
    d.line([(round(x * 2), round(y * 2)) for x, y in points],
           fill=fill, width=round(width * 2), joint="curve")

def polygon(d, points, fill):
    d.polygon([(round(x * 2), round(y * 2)) for x, y in points],
              fill=fill)

def gradient_tile(top, bottom, outline, inset, box=(48, 40, 464, 459),
                  radius=103):
    im = Image.new("RGBA", (S, S), (0, 0, 0, 0))
    mask = Image.new("L", (S, S), 0)
    rr(ImageDraw.Draw(mask), box, radius, fill=255)
    shadow = Image.new("RGBA", (S, S), (0, 0, 0, 0))
    shadow.paste((4, 7, 23, 125), (0, 0), mask.filter(ImageFilter.GaussianBlur(24)))
    im.alpha_composite(shadow)
    gradient = Image.new("RGBA", (S, S), (0, 0, 0, 0))
    draw = ImageDraw.Draw(gradient)
    a, b = rgb(top), rgb(bottom)
    for y in range(52, 916):
        color = mix(a, b, (y - 52) / 864)
        draw.line((40, y, 940, y), fill=color + (255,), width=1)
    im.paste(gradient, (0, 0), mask)
    draw = ImageDraw.Draw(im)
    rr(draw, box, radius, outline=outline, width=3)
    rr(draw, (57, 49, 455, 450), 96, outline=inset, width=2)
    overlay = Image.new("RGBA", (S, S), (0, 0, 0, 0))
    rr(ImageDraw.Draw(overlay), (66, 58, 446, 167), 48,
       fill=(255, 255, 255, 42))
    overlay.putalpha(Image.composite(overlay.getchannel("A"),
                                    Image.new("L", (S, S), 0), mask))
    im.alpha_composite(overlay)
    return im

def draw_autoclick():
    im = gradient_tile("#30C7C7", "#12608A", "#79F0E9", "#228EB1")
    d = ImageDraw.Draw(im)
    rr(d, (181, 119, 331, 362), 71, outline="#FFFFFF", width=18)
    line(d, [(256, 121), (256, 205)], "#E8FFFF", 10)
    rr(d, (242, 161, 270, 222), 13, fill="#FFFFFF")
    line(d, [(196, 258), (318, 258)], "#D4FCFF", 9)
    ellipse(d, (342, 124, 390, 172), outline="#FFF2A0", width=8)
    ellipse(d, (357, 139, 375, 157), fill="#FFFFFF")
    line(d, [(401, 126), (422, 105)], "#FFFFFF", 12)
    line(d, [(410, 162), (441, 162)], "#FFFFFF", 12)
    line(d, [(378, 102), (381, 78)], "#FFFFFF", 11)
    return im.resize((512, 512), Image.Resampling.LANCZOS)

def draw_audio_output():
    im = gradient_tile("#587AE9", "#292D8F", "#9FBBFF", "#647CEB")
    d = ImageDraw.Draw(im)
    rr(d, (132, 207, 184, 305), 12, fill="#FFFFFF")
    polygon(d, [(176, 207), (261, 148), (261, 363), (176, 305)], "#FFFFFF")
    line(d, [(307, 194), (391, 194)], "#F7FCFF", 14)
    polygon(d, [(373, 167), (411, 194), (373, 221)], "#FFFFFF")
    line(d, [(395, 291), (313, 291)], "#FFFFFF", 14)
    polygon(d, [(330, 264), (294, 291), (330, 318)], "#FFFFFF")
    d.arc((452, 408, 666, 620), -55, 55, fill="#DDE8FF", width=24)
    return im.resize((512, 512), Image.Resampling.LANCZOS)

def draw_desktop_logo():
    im = gradient_tile("#F25665", "#AA2849", "#FF8795", "#D44561",
                       box=(38, 36, 474, 475), radius=107)
    d = ImageDraw.Draw(im)
    for yy in (169, 268):
        for xx in (171, 271):
            rr(d, (xx+3, yy+7, xx+78, yy+82), 21, fill="#93223E")
            rr(d, (xx, yy, xx+76, yy+76), 20, fill="#F5F7FF")
            rr(d, (xx+8, yy+7, xx+67, yy+26), 10, fill="#FFFFFF")
    return im

def main():
    ART.mkdir(parents=True, exist_ok=True)
    INSTALLER.mkdir(parents=True, exist_ok=True)
    for name, draw_icon in [("auto_click", draw_autoclick),
                            ("audio_output", draw_audio_output)]:
        target = ART / (name + ".png")
        draw_icon().save(target, "PNG", optimize=True)
        print("Generated action icon", name, target.stat().st_size)
    logo = draw_desktop_logo()
    sizes = [(n, n) for n in (16, 24, 32, 48, 64, 128, 256)]
    for path in (APP / "PulseNoteDeck.ico", INSTALLER / "icon.ico"):
        path.parent.mkdir(parents=True, exist_ok=True)
        logo.save(path, format="ICO", sizes=sizes)
        print("Generated application ICO", path.stat().st_size)

if __name__ == "__main__":
    main()
