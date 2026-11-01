package main

import (
	"math"
	"os"
	"strings"
	"testing"
)

func TestSoftKeycapColorsPreserveNamesAndReadableContrast(t *testing.T) {
	if len(palette) != 7 || len(paletteRGB) != 7 || len(paletteHighlightRGB) != 7 {
		t.Fatal("palette length")
	}
	for i, name := range palette {
		if got := colorValue(name); got != paletteRGB[i] {
			t.Fatalf("%s wrong color: %X", name, got)
		}
		bgr := paletteRGB[i]
		r, g, b := float64(bgr&255), float64((bgr>>8)&255), float64((bgr>>16)&255)
		// Muted palette uses light fills; check that bright
		// fills use dark foreground, while darker fills use white foreground.
		brightness := 0.2126*r + 0.7152*g + 0.0722*b
		if brightness >= 130 && !lightKeycap(name) {
			t.Fatalf("%s soft keycap needs dark foreground: %.1f", name, brightness)
		}
		if name == "회색" && brightness < 190 {
			t.Fatalf("light-gray requirement regressed: %.1f", brightness)
		}
		if keycapHighlight(name) == paletteRGB[i] {
			t.Fatalf("%s highlight not distinguishable", name)
		}
	}
}

func TestBackBufferedWin32RenderingContracts(t *testing.T) {
	contents, err := os.ReadFile("gui_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(contents)
	for _, required := range []string{"createCompatibleDC.Call", "createCompatibleBitmap.Call", "bitBlt.Call", "case WM_ERASEBKGND:", "WS_CLIPCHILDREN", "!g.editsPlaced || g.labelRect != label", "!g.editsPlaced || g.valueRect != value"} {
		if !strings.Contains(src, required) {
			t.Fatalf("flicker guard missing: %s", required)
		}
	}
	if strings.Contains(src, "case WM_TIMER:\n\t\tg.redraw()") {
		t.Fatal("periodic full repaint regression")
	}
}

func TestSoftKeycapsReadableText(t *testing.T) {
	for i, colorName := range palette {
		v := paletteRGB[i]
		r, g, b := float64(v&0xff)/255.0, float64((v>>8)&0xff)/255.0, float64((v>>16)&0xff)/255.0
		linear := func(c float64) float64 {
			if c <= 0.04045 {
				return c / 12.92
			}
			return math.Pow((c+0.055)/1.055, 2.4)
		}
		light := 0.2126*linear(r) + 0.7152*linear(g) + 0.0722*linear(b)
		fg := []float64{1, 1, 1}
		if lightKeycap(colorName) {
			fg = []float64{10.0 / 255, 15.0 / 255, 24.0 / 255}
		}
		foreground := 0.2126*linear(fg[0]) + 0.7152*linear(fg[1]) + 0.0722*linear(fg[2])
		var ratio float64
		if light > foreground {
			ratio = (light + 0.05) / (foreground + 0.05)
		} else {
			ratio = (foreground + 0.05) / (light + 0.05)
		}
		if ratio < 4.5 {
			t.Fatalf("keycap %s has insufficient foreground contrast %.2f", colorName, ratio)
		}
	}
}
