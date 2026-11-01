package main

import (
	"image/png"
	"os"
	"testing"
)

func TestColorfulIconAlphaAndPalette(t *testing.T) {
	for _, name := range []string{"none", "program", "folder", "website", "shortcut", "volume_up", "volume_down", "volume_mute", "media_play", "media_next", "media_prev"} {
		f, err := os.Open("assets/" + name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			t.Fatal(name, err)
		}
		if img.Bounds().Dx() != 512 || img.Bounds().Dy() != 512 {
			t.Fatalf("%s dimensions changed", name)
		}
		if _, _, _, a := img.At(0, 0).RGBA(); a != 0 {
			t.Errorf("%s missing transparent padding", name)
		}
		palette := map[[3]uint8]struct{}{}
		visible := 0
		for y := 50; y < 450; y += 8 {
			for x := 50; x < 450; x += 8 {
				r, g, b, a := img.At(x, y).RGBA()
				if a > 0xE000 {
					visible++
					palette[[3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}] = struct{}{}
				}
			}
		}
		if visible < 300 || len(palette) < 80 {
			t.Errorf("%s insufficient detailed artwork: visible=%d colors=%d", name, visible, len(palette))
		}
	}
}
