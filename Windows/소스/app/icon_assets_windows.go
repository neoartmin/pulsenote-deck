//go:build windows

package main

import (
	"bytes"
	"embed"
	"image/png"
	"syscall"
	"unsafe"
)

// 512px, colorful illustrated rounded icons are embedded in the EXE.
//
//go:embed assets/*.png
var artwork embed.FS

const pixelFormat32bppARGB = 0x26200a

type roundedAsset struct {
	image uintptr
	data  []byte // GDI+ Bitmap references the source memory: keep it alive.
}

var (
	gdiPlus             = syscall.NewLazyDLL("gdiplus.dll")
	gdipStart           = gdiPlus.NewProc("GdiplusStartup")
	gdipStop            = gdiPlus.NewProc("GdiplusShutdown")
	gdipBitmapFromScan  = gdiPlus.NewProc("GdipCreateBitmapFromScan0")
	gdipDisposeImage    = gdiPlus.NewProc("GdipDisposeImage")
	gdipFromHDC         = gdiPlus.NewProc("GdipCreateFromHDC")
	gdipDeleteGraphics  = gdiPlus.NewProc("GdipDeleteGraphics")
	gdipInterpolation   = gdiPlus.NewProc("GdipSetInterpolationMode")
	gdipImageRect       = gdiPlus.NewProc("GdipDrawImageRectI")
	gdipBitmapFromHICON = gdiPlus.NewProc("GdipCreateBitmapFromHICON")
	gdipBitmapFromFile  = gdiPlus.NewProc("GdipCreateBitmapFromFile")
	gdipImageWidth      = gdiPlus.NewProc("GdipGetImageWidth")
	gdipImageHeight     = gdiPlus.NewProc("GdipGetImageHeight")
	gdipBitmapLockBits  = gdiPlus.NewProc("GdipBitmapLockBits")
	gdipBitmapUnlock    = gdiPlus.NewProc("GdipBitmapUnlockBits")
	gdipPixelOffsetMode = gdiPlus.NewProc("GdipSetPixelOffsetMode")
	gdipToken           uintptr
	roundedIcons        map[string]roundedAsset
)

type gdipStartupInput struct {
	Version                  uint32
	DebugEventCallback       uintptr
	SuppressBackgroundThread int32
	SuppressExternalCodecs   int32
}

func initRoundedArtwork() {
	startup := gdipStartupInput{Version: 1}
	if code, _, _ := gdipStart.Call(uintptr(unsafe.Pointer(&gdipToken)), uintptr(unsafe.Pointer(&startup)), 0); code != 0 {
		return
	}
	roundedIcons = map[string]roundedAsset{}
	for _, name := range []string{"none", "program", "folder", "website", "shortcut", "volume_up", "volume_down", "volume_mute", "media_play", "media_next", "media_prev", "capture", "copy_text", "system_settings", "open_timers", "open_alarms", "web_search", "compose_email", "qr_code", "sequence", "clipboard_history", "window_control"} {
		b, err := artwork.ReadFile("assets/" + name + ".png")
		if err != nil {
			continue
		}
		decoded, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			continue
		}
		w, h := decoded.Bounds().Dx(), decoded.Bounds().Dy()
		if w < 1 || h < 1 {
			continue
		}
		pixels := make([]byte, w*h*4)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				r, g, b, a := decoded.At(decoded.Bounds().Min.X+x, decoded.Bounds().Min.Y+y).RGBA()
				i := (y*w + x) * 4
				// RGBA() values are already premultiplied; un-premultiply for
				// GDI+ PixelFormat32bppARGB and preserve semitransparent corners.
				if a != 0 {
					pixels[i] = uint8((b * 65535 / a) >> 8)
					pixels[i+1] = uint8((g * 65535 / a) >> 8)
					pixels[i+2] = uint8((r * 65535 / a) >> 8)
					pixels[i+3] = uint8(a >> 8)
				}
			}
		}
		var bmp uintptr
		code, _, _ := gdipBitmapFromScan.Call(uintptr(w), uintptr(h), uintptr(w*4), pixelFormat32bppARGB, uintptr(unsafe.Pointer(&pixels[0])), uintptr(unsafe.Pointer(&bmp)))
		if code == 0 && bmp != 0 {
			roundedIcons[name] = roundedAsset{image: bmp, data: pixels}
		}
	}
}

func shutdownRoundedArtwork() {
	for _, entry := range roundedIcons {
		gdipDisposeImage.Call(entry.image)
	}
	roundedIcons = nil
	if gdipToken != 0 {
		gdipStop.Call(gdipToken)
		gdipToken = 0
	}
}

// drawTextMacroArtwork renders a dedicated T + text-lines symbol. This is
// deliberately different from the purple keyboard shortcut icon and uses
// scalable native GDI shapes so it remains sharp on both the 15-key grid and
// the right-side action picker, without requiring a downloaded font/image.
func (d *drawCtx) drawTextMacroArtwork(x, y, size int) {
    blue := RGB(35, 99, 235)
    border := RGB(100, 157, 255)
    white := RGB(255, 255, 255)
    d.rounded(x, y, size, size, maxInt(5, size/5), blue, border)
    d.textBold("T", x+size/10, y+size/8, size*47/100, size*76/100,
        maxInt(13, size*57/100), white, true)
    left := x+size*61/100
    lineWidth := maxInt(5,size*25/100)
    lineHeight := maxInt(2,size*7/100)
    for i:=0;i<3;i++ {
        lineY:=y+size*(31+17*i)/100
        d.rounded(left,lineY,lineWidth,lineHeight,maxInt(1,lineHeight/2),white,white)
    }
}

// The icon renderer targets the existing double-buffered HDC and reuses one
// GDI+ Graphics per paint, avoiding per-key font rasterization and upscaling.
func (d *drawCtx) actionArtwork(name string, x, y, size int) bool {
    if name == "text" {
        d.drawTextMacroArtwork(x, y, size)
        return true
    }
	asset, ok := roundedIcons[name]
	if !ok || asset.image == 0 {
		return false
	}
	// Dispose graphics before the next GDI font/text operation. Reusing a GDI+
	// context across interleaved GDI calls can reorder or overwrite drawings.
	var graphics uintptr
	code, _, _ := gdipFromHDC.Call(d.hdc, uintptr(unsafe.Pointer(&graphics)))
	if code != 0 || graphics == 0 {
		return false
	}
	defer gdipDeleteGraphics.Call(graphics)
	gdipInterpolation.Call(graphics, 7)   // HighQualityBicubic
	gdipPixelOffsetMode.Call(graphics, 4) // HighQuality
	xx, yy := d.xy(x, y)
	sz := maxInt(1, int(float64(size)*d.scale))
	result, _, _ := gdipImageRect.Call(graphics, asset.image, uintptr(xx), uintptr(yy), uintptr(sz), uintptr(sz))
	return result == 0
}
