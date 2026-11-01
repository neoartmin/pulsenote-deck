package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransparentIconPathsAreSharedForAllPrograms(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "Program A.lnk")
	choices := transparentIconPaths(path, base)
	if len(choices) != 2 || !strings.HasSuffix(choices[1], filepath.Join("icons", "Program A.png")) {
		t.Fatalf("unexpected icon override path: %v", choices)
	}
	if choices[0] == choices[1] || len(transparentIconPaths("", base)) != 0 {
		t.Fatal("bad common icon path convention")
	}
}

func TestOnlyRealTransparentPNGsAreSelected(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "Any App.lnk")
	file := transparentIconPaths(path, base)[1]
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		t.Fatal(err)
	}
	write := func(alpha uint8) {
		t.Helper()
		f, err := os.Create(file)
		if err != nil {
			t.Fatal(err)
		}
		img := image.NewNRGBA(image.Rect(0, 0, 5, 5))
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				img.SetNRGBA(x, y, color.NRGBA{R: 20, G: 200, B: 140, A: 255})
			}
		}
		img.SetNRGBA(0, 0, color.NRGBA{R: 0, G: 0, B: 0, A: alpha})
		if err = png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	write(255)
	if pngHasTransparency(file) {
		t.Fatal("opaque image incorrectly accepted")
	}
	write(0)
	if !pngHasTransparency(file) {
		t.Fatal("real alpha dropped")
	}
	if err := os.WriteFile(file, []byte("not a png"), 0600); err != nil {
		t.Fatal(err)
	}
	if pngHasTransparency(file) {
		t.Fatal("corrupt PNG accepted")
	}
}
