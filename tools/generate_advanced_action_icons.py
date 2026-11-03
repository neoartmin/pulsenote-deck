#!/usr/bin/env python3
"""512px, antialiased PNG assets for the advanced macro actions.

All shared PNGs are rendered once and written byte-identically to both
Windows and Android. No runtime SVG/primitive drawing on small keycaps.
"""
from pathlib import Path
from math import pi,cos,sin
from PIL import Image,ImageDraw,ImageOps
ROOT=Path(__file__).resolve().parent.parent
DESKTOP=ROOT/"Windows"/"소스"/"app"/"assets"
ANDROID=ROOT/"android-pn12"/"app"/"src"/"main"/"assets"/"action_icons"
FACTOR=3
SIZE=512*FACTOR
X=lambda x:round(x*FACTOR)
WHITE=(252,250,255,255)

PALETTES={
    "web_search":((57,145,232),(19,61,143)),
    "compose_email":((225,102,145),(111,40,107)),
    "qr_code":((76,168,121),(28,97,104)),
    "sequence":((155,106,230),(53,50,158)),
    "share_text":((55,180,211),(19,90,164)),
    "clipboard_history":((228,154,82),(142,68,61)),
    "window_control":((98,127,233),(59,62,131)),
}
SHARED=["web_search","compose_email","qr_code","sequence"]

def line(draw,points,width=18,color=WHITE):
    poly=[(X(a),X(b)) for a,b in points]
    draw.line(poly,fill=color,width=X(width),joint="curve")
    r=X(width)//2
    for x,y in poly:draw.ellipse((x-r,y-r,x+r,y+r),fill=color)

def canvas(name):
    top,bottom=PALETTES[name]
    gradient=Image.new("RGBA",(SIZE,SIZE),(0,0,0,0))
    d=ImageDraw.Draw(gradient)
    for row in range(X(54),X(460),3):
        t=(row-X(54))/max(1,X(406))
        rgb=tuple(round(top[i]*(1-t)+bottom[i]*t) for i in range(3))
        d.rectangle((X(54),row,X(459),min(row+3,X(460))),fill=rgb+(255,))
    mask=Image.new("L",(SIZE,SIZE),0)
    ImageDraw.Draw(mask).rounded_rectangle((X(54),X(54),X(459),X(459)),radius=X(88),fill=255)
    gradient.putalpha(mask)
    d=ImageDraw.Draw(gradient)
    d.rounded_rectangle((X(62),X(62),X(451),X(451)),radius=X(81),outline=(255,248,255,125),width=X(9))
    d.rounded_rectangle((X(75),X(74),X(437),X(174)),radius=X(48),outline=(255,249,255,55),width=X(4))
    return gradient,d

def paint(name,d):
    if name=="web_search":
        d.ellipse((X(145),X(140),X(324),X(321)),outline=WHITE,width=X(28))
        line(d,[(312,313),(375,377)],27)
        d.ellipse((X(178),X(175),X(289),X(286)),outline=(255,255,255,95),width=X(7))
    elif name=="compose_email":
        d.rounded_rectangle((X(123),X(178),X(389),X(347)),radius=X(24),outline=WHITE,width=X(22))
        line(d,[(132,197),(256,276),(380,197)],20)
        line(d,[(130,334),(206,274)],14)
        line(d,[(382,334),(306,274)],14)
    elif name=="qr_code":
        for x,y in ((130,145),(288,145),(130,303)):
            d.rounded_rectangle((X(x),X(y),X(x+94),X(y+94)),radius=X(8),outline=WHITE,width=X(16))
            d.rectangle((X(x+29),X(y+29),X(x+65),X(y+65)),fill=WHITE)
        for x,y in ((299,309),(344,309),(294,354),(340,356),(385,355),(390,298)):
            d.rounded_rectangle((X(x),X(y),X(x+25),X(y+25)),radius=X(4),fill=WHITE)
    elif name=="sequence":
        for x,y in ((112,181),(280,181),(280,317)):
            d.rounded_rectangle((X(x),X(y),X(x+111),X(y+84)),radius=X(17),outline=WHITE,width=X(16))
        line(d,[(228,223),(276,223)],15)
        line(d,[(266,209),(282,223),(266,237)],12)
        line(d,[(335,274),(335,313)],15)
        line(d,[(322,299),(335,315),(348,299)],12)
    elif name=="share_text":
        for x,y in ((158,256),(336,160),(336,356)):
            d.ellipse((X(x-30),X(y-30),X(x+30),X(y+30)),fill=WHITE)
        line(d,[(184,244),(305,176)],15)
        line(d,[(184,267),(306,342)],15)
    elif name=="clipboard_history":
        d.rounded_rectangle((X(150),X(130),X(352),X(393)),radius=X(26),outline=WHITE,width=X(20))
        d.rounded_rectangle((X(207),X(108),X(295),X(155)),radius=X(14),fill=WHITE)
        d.ellipse((X(202),X(214),X(335),X(347)),outline=WHITE,width=X(17))
        line(d,[(269,280),(269,239)],13)
        line(d,[(269,280),(304,298)],13)
    elif name=="window_control":
        d.rounded_rectangle((X(115),X(153),X(335),X(322)),radius=X(17),outline=WHITE,width=X(18))
        line(d,[(118,207),(332,207)],13)
        d.rounded_rectangle((X(191),X(230),X(402),X(394)),radius=X(18),outline=WHITE,width=X(18))
        line(d,[(194,283),(399,283)],13)
        for x in (224,245):d.ellipse((X(x),X(255),X(x+10),X(265)),fill=WHITE)

def render(name):
    img,d=canvas(name)
    paint(name,d)
    result=img.resize((512,512),Image.Resampling.LANCZOS)
    assert result.mode=="RGBA" and result.getpixel((0,0))[3]==0
    return result

def main():
    for name in PALETTES:
        image=render(name)
        targets=[DESKTOP,ANDROID] if name in SHARED else ([ANDROID] if name=="share_text" else [DESKTOP])
        for directory in targets:
            directory.mkdir(parents=True,exist_ok=True)
            path=directory/(name+".png")
            image.save(path,format="PNG",optimize=True)
            print(name,path.stat().st_size,path)
if __name__=="__main__":main()
