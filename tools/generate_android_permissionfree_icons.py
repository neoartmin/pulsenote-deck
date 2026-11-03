#!/usr/bin/env python3
"""Produce 512px PNG artwork for four Android-only permission-free actions.

Existing Windows/Android shared assets are untouched; icons 12..15 use the
same high-resolution rounded colored artwork pipeline as other action types.
"""
from pathlib import Path
from math import cos, sin, pi
from PIL import Image, ImageDraw, ImageChops, ImageFilter

ROOT=Path(__file__).resolve().parent.parent
OUTPUT=ROOT/"android-pn12"/"app"/"src"/"main"/"assets"/"action_icons"
SCALE=3
N=512*SCALE
def px(value):return round(value*SCALE)

def stroke(draw,points,color,width=20):
    coords=[(px(x),px(y)) for x,y in points]
    draw.line(coords,fill=color,width=px(width),joint="curve")
    r=px(width/2)
    for x,y in coords:
        draw.ellipse((x-r,y-r,x+r,y+r),fill=color)

def make_canvas(upper,lower):
    # True alpha padding and high quality antialiased shape.
    mask=Image.new("L",(N,N),0)
    ImageDraw.Draw(mask).rounded_rectangle((px(52),px(52),px(460),px(460)),
                                           radius=px(91),fill=255)
    image=Image.new("RGBA",(N,N),(0,0,0,0))
    d=ImageDraw.Draw(image)
    for y in range(px(52),px(461),3):
        t=max(0,min(1,(y-px(52))/px(408)))
        rgb=tuple(round(upper[i]*(1-t)+lower[i]*t) for i in range(3))
        d.rectangle((px(52),y,px(460),min(px(461),y+3)),fill=rgb+(255,))
    image.putalpha(mask)
    d=ImageDraw.Draw(image)
    d.rounded_rectangle((px(58),px(58),px(454),px(454)),radius=px(86),
                        outline=(242,235,255,115),width=px(9))
    d.rounded_rectangle((px(76),px(75),px(436),px(172)),radius=px(47),
                        outline=(255,255,255,34),width=px(5))
    return image,d

WHITE=(251,250,255,255)
SUBTLE=(227,239,255,235)
def copy_icon():
    image,d=make_canvas((68,154,232),(24,76,157))
    d.rounded_rectangle((px(144),px(142),px(320),px(334)),radius=px(21),
                        outline=SUBTLE,width=px(20))
    d.rounded_rectangle((px(192),px(183),px(370),px(374)),radius=px(23),
                        fill=(40,111,181,255),outline=WHITE,width=px(22))
    for y,w in ((241,92),(277,103),(312,79)):
        stroke(d,[(218,y),(218+w,y)],WHITE,13)
    return image

def gear_icon():
    image,d=make_canvas((145,103,232),(62,37,148))
    points=[]
    cx=cy=256
    teeth=12
    for i in range(teeth*4):
        theta=-pi/2+2*pi*i/(teeth*4)
        rad=149 if (i%4 in (0,1)) else 121
        points.append((px(cx+rad*cos(theta)),px(cy+rad*sin(theta))))
    d.polygon(points,fill=WHITE)
    d.ellipse((px(204),px(204),px(308),px(308)),fill=(77,47,149,255))
    return image

def timer_icon():
    image,d=make_canvas((38,180,177),(20,98,135))
    d.rounded_rectangle((px(229),px(103),px(283),px(150)),
                        radius=px(13),fill=WHITE)
    stroke(d,[(256,167),(256,195)],WHITE,19)
    d.ellipse((px(134),px(168),px(378),px(412)),outline=WHITE,width=px(23))
    d.ellipse((px(246),px(278),px(266),px(298)),fill=WHITE)
    stroke(d,[(256,288),(256,222)],WHITE,20)
    stroke(d,[(256,288),(307,314)],WHITE,19)
    return image

def alarm_icon():
    image,d=make_canvas((241,141,103),(182,62,102))
    d.ellipse((px(150),px(181),px(362),px(393)),outline=WHITE,width=px(22))
    stroke(d,[(256,286),(256,215)],WHITE,20)
    stroke(d,[(256,286),(304,316)],WHITE,20)
    d.ellipse((px(246),px(277),px(267),px(298)),fill=WHITE)
    d.arc((px(128),px(123),px(242),px(237)),188,298,
          fill=WHITE,width=px(23))
    d.arc((px(270),px(123),px(384),px(237)),245,354,
          fill=WHITE,width=px(23))
    stroke(d,[(183,392),(164,418)],WHITE,20)
    stroke(d,[(327,392),(346,418)],WHITE,20)
    return image

ICONS={
    "copy_text":copy_icon,
    "system_settings":gear_icon,
    "open_timers":timer_icon,
    "open_alarms":alarm_icon,
}
def main():
    OUTPUT.mkdir(parents=True,exist_ok=True)
    for name,render in ICONS.items():
        icon=render().resize((512,512),Image.Resampling.LANCZOS)
        assert icon.mode=="RGBA"
        assert icon.getpixel((0,0))[3]==0
        dst=OUTPUT/(name+".png")
        icon.save(dst,"PNG",optimize=True)
        print(name, dst.stat().st_size,"bytes")
if __name__=="__main__":main()
