//go:build windows

package main

import "syscall"

var captureEllipse = syscall.NewLazyDLL("gdi32.dll").NewProc("Ellipse")

func (d *drawCtx) captureCircle(x, y, diameter int, fill, border uint32) {
    q := d.ab(x, y, diameter, diameter)
    pen, _, _ := createPen.Call(PS_SOLID,
        uintptr(maxInt(1, int(d.scale))), uintptr(border))
    brush, _, _ := createSolidBrush.Call(uintptr(fill))
    oldPen, _, _ := selectObject.Call(d.hdc, pen)
    oldBrush, _, _ := selectObject.Call(d.hdc, brush)
    captureEllipse.Call(d.hdc, uintptr(q.Left), uintptr(q.Top),
        uintptr(q.Right), uintptr(q.Bottom))
    selectObject.Call(d.hdc, oldPen)
    selectObject.Call(d.hdc, oldBrush)
    deleteObject.Call(pen)
    deleteObject.Call(brush)
}

// Camera-with-capture-corners motif shared with Android and assets/capture.svg.
func (d *drawCtx) drawCaptureArtwork(x, y, size int) {
    bg := RGB(69, 43, 132)
    white := RGB(247, 246, 255)
    violet := RGB(167, 134, 241)
    d.rounded(x, y, size, size, maxInt(4,size/5), bg, violet)
    thickness := maxInt(2, size/17)
    strokeLength := size/8
    for _, c := range []struct{ px, py, sx, sy int }{
        {15, 19, 1, 1}, {85, 19, -1, 1}, {15, 81, 1, -1}, {85, 81, -1, -1},
    } {
        cx := x+size*c.px/100
        cy := y+size*c.py/100
        horizontalX := cx
        if c.sx<0 { horizontalX-=strokeLength }
        verticalY := cy
        if c.sy<0 { verticalY-=strokeLength }
        d.rounded(horizontalX, cy, strokeLength, thickness, thickness, white, white)
        d.rounded(cx, verticalY, thickness, strokeLength, thickness, white, white)
    }
    d.rounded(x+size*22/100, y+size*40/100,
        size*56/100,size*31/100,maxInt(3,size/13),white,white)
    d.rounded(x+size*26/100, y+size*44/100,
        size*48/100,size*23/100,maxInt(2,size/14),bg,bg)
    d.rounded(x+size*36/100, y+size*34/100,
        size*28/100,size*15/100,maxInt(2,size/16),white,white)
    d.captureCircle(x+size*40/100, y+size*46/100,
        size*20/100,white,white)
    d.captureCircle(x+size*45/100, y+size*51/100,
        size*10/100,bg,bg)
}
