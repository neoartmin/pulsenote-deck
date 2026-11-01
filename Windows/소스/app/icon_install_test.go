package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func samplePNG(t *testing.T, filename string, alpha uint8) {
	t.Helper()
	f, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	pic := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			pic.SetNRGBA(x, y, color.NRGBA{R: 10, G: 90, B: 80, A: 255})
		}
	}
	pic.SetNRGBA(0, 0, color.NRGBA{R: 20, G: 100, B: 200, A: alpha})
	if err = png.Encode(f, pic); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestInstallTransparentIconAndRestore(t *testing.T) {
	home := t.TempDir()
	source := filepath.Join(home, "custom.png")
	app := filepath.Join(home, "Arduino IDE.lnk")
	samplePNG(t, source, 0)
	installed, err := installTransparentIcon(app, home, source)
	if err != nil || installed != transparentIconPaths(app, home)[0] {
		t.Fatalf("install: %v %v", installed, err)
	}
	if !pngHasTransparency(installed) {
		t.Fatal("alpha lost")
	}
	if err = removeManagedTransparentIcon(app, home); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(installed); !os.IsNotExist(err) {
		t.Fatal("managed override not removed", err)
	}
}
func TestInstallRejectsOpaqueAndKeepsExisting(t *testing.T) {
	home := t.TempDir()
	app := "C:/Program Files/Some App.exe"
	source := filepath.Join(home, "custom.png")
	samplePNG(t, source, 0)
	saved, err := installTransparentIcon(app, home, source)
	if err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(saved)
	if err != nil {
		t.Fatal(err)
	}
	samplePNG(t, source, 255)
	if _, err = installTransparentIcon(app, home, source); err == nil {
		t.Fatal("opaque PNG accepted")
	}
	got, err := os.ReadFile(saved)
	if err != nil || string(got) != string(old) {
		t.Fatal("previous image modified")
	}
}
func TestSameNameDoesNotCollide(t *testing.T) {
	home := t.TempDir()
	src := filepath.Join(home, "custom.png")
	samplePNG(t, src, 0)
	a, _ := installTransparentIcon(`C:\First\Editor.lnk`, home, src)
	b, _ := installTransparentIcon(`C:\Second\Editor.lnk`, home, src)
	if a == b {
		t.Fatal("two apps share override")
	}
}
