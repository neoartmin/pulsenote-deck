//go:build windows

package main

import "unsafe"

// GDI's RoundRect does not anti-alias. GDI+ draws two filled rounded paths
// with smooth edges while keeping the original v7.5.17 fill/border colors.
// GDI+ graphics objects are short-lived because the frame deliberately
// interleaves GDI text and GDI+ icons on one back-buffered HDC.
var (
    aaSetSmoothing  = gdiPlus.NewProc("GdipSetSmoothingMode")
    aaCreatePath    = gdiPlus.NewProc("GdipCreatePath")
    aaDeletePath    = gdiPlus.NewProc("GdipDeletePath")
    aaPathLine      = gdiPlus.NewProc("GdipAddPathLineI")
    aaPathBezier    = gdiPlus.NewProc("GdipAddPathBezierI")
    aaPathRect      = gdiPlus.NewProc("GdipAddPathRectangleI")
    aaClosePath     = gdiPlus.NewProc("GdipClosePathFigure")
    aaSolidBrush    = gdiPlus.NewProc("GdipCreateSolidFill")
    aaDeleteBrush   = gdiPlus.NewProc("GdipDeleteBrush")
    aaFillPath      = gdiPlus.NewProc("GdipFillPath")
)

func gdipArg(v int) uintptr { return uintptr(int64(v)) }
func gdipARGB(color uint32) uintptr {
    // Windows COLORREF is 0x00BBGGRR; GDI+ ARGB is 0xAARRGGBB.
    return uintptr(0xFF000000 | ((color & 0xff) << 16) | (color & 0xff00) | ((color >> 16) & 0xff))
}

func newRoundedGdipPath(x, y, w, h, radius int) (uintptr, bool) {
    if w <= 0 || h <= 0 {
        return 0, false
    }
    var path uintptr
    status, _, _ := aaCreatePath.Call(0, uintptr(unsafe.Pointer(&path)))
    if status != 0 || path == 0 { return 0, false }
    success := false
    defer func() {
        if !success { aaDeletePath.Call(path) }
    }()
    r := minInt(radius, minInt(w,h)/2)
    if r <= 1 {
        status, _, _ = aaPathRect.Call(path, gdipArg(x), gdipArg(y), gdipArg(w), gdipArg(h))
        if status != 0 { return 0, false }
    } else {
        // Four cubic Beziers, accurate to <0.1% of a circle using k=0.552.
        k := (r*5523 + 5000) / 10000
        x2, y2 := x+w-1, y+h-1
        line := func(x1,y1,x2,y2 int) bool {
            status,_,_ := aaPathLine.Call(path, gdipArg(x1),gdipArg(y1),gdipArg(x2),gdipArg(y2))
            return status == 0
        }
        curve := func(x1,y1,c1x,c1y,c2x,c2y,x2,y2 int) bool {
            status,_,_ := aaPathBezier.Call(path,
                gdipArg(x1),gdipArg(y1),gdipArg(c1x),gdipArg(c1y),
                gdipArg(c2x),gdipArg(c2y),gdipArg(x2),gdipArg(y2))
            return status == 0
        }
        if !line(x+r,y,x2-r,y) ||
           !curve(x2-r,y,x2-r+k,y,x2,y+r-k,x2,y+r) ||
           !line(x2,y+r,x2,y2-r) ||
           !curve(x2,y2-r,x2,y2-r+k,x2-r+k,y2,x2-r,y2) ||
           !line(x2-r,y2,x+r,y2) ||
           !curve(x+r,y2,x+r-k,y2,x,y2-r+k,x,y2-r) ||
           !line(x,y2-r,x,y+r) ||
           !curve(x,y+r,x,y+r-k,x+r-k,y,x+r,y) {
            return 0,false
        }
    }
    status, _, _ = aaClosePath.Call(path)
    if status != 0 { return 0,false }
    success=true
    return path,true
}

func fillGdipRounded(g uintptr, x,y,w,h,r int,color uint32) bool {
    path, ok := newRoundedGdipPath(x,y,w,h,r)
    if !ok {return false}
    defer aaDeletePath.Call(path)
    var brush uintptr
    status,_,_ := aaSolidBrush.Call(gdipARGB(color), uintptr(unsafe.Pointer(&brush)))
    if status != 0 || brush == 0 {return false}
    defer aaDeleteBrush.Call(brush)
    status,_,_ = aaFillPath.Call(g,brush,path)
    return status == 0
}

func (d *drawCtx) antiAliasedRounded(x,y,w,h,r int,fill,border uint32) bool {
    if gdipToken == 0 {return false}
    q:=d.ab(x,y,w,h)
    width,height:=int(q.Right-q.Left),int(q.Bottom-q.Top)
    if width < 3 || height < 3 {return false}
    var graphics uintptr
    status,_,_:=gdipFromHDC.Call(d.hdc,uintptr(unsafe.Pointer(&graphics)))
    if status!=0||graphics==0{return false}
    defer gdipDeleteGraphics.Call(graphics)
    status,_,_=aaSetSmoothing.Call(graphics,4) // SmoothingModeAntiAlias
    if status!=0{return false}
    radius:=int(float64(r)*d.scale+0.5)
    if radius<1{radius=1}
    if !fillGdipRounded(graphics,int(q.Left),int(q.Top),width,height,radius,border) {
        return false
    }
    if fill!=border {
        edge:=maxInt(1,int(d.scale+0.5))
        if width>edge*2 && height>edge*2 {
            if !fillGdipRounded(graphics,int(q.Left)+edge,int(q.Top)+edge,
                    width-edge*2,height-edge*2,maxInt(0,radius-edge),fill){
                return false
            }
        }
    }
    return true
}
