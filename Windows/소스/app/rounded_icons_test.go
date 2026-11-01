package main

import (
	"image/png"
	"os"
	"strings"
	"testing"
)

func TestEmbeddedActionArtworkQuality(t *testing.T) {
	names := []string{"none", "program", "folder", "website", "shortcut", "volume_up", "volume_down", "volume_mute", "media_play", "media_next", "media_prev"}
	for _, name := range names {
		f, err := os.Open("assets/" + name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		config, err := png.DecodeConfig(f)
		f.Close()
		if err != nil {
			t.Fatal(name, err)
		}
		if config.Width != 512 || config.Height != 512 {
			t.Fatalf("%s: expected 512px, got %dx%d", name, config.Width, config.Height)
		}
	}
}
func TestShellIconUsesHighestAvailableResolution(t *testing.T) {
	b, e := os.ReadFile("gui_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	s := string(b)
	for _, wanted := range []string{"SHGetImageList", "highResShellIcon(path)", "gdipBitmapFromHICON.Call", "gdipImageRect.Call", "gdipPixelOffsetMode.Call", "drawAppIcon(d, key.Value"} {
		if !strings.Contains(s, wanted) {
			t.Errorf("missing icon quality feature: %s", wanted)
		}
	}
	if strings.Contains(s, "0x100|1") {
		t.Error("16px SHGFI_SMALLICON regression")
	}
}

func TestAppIconHasNoDarkBackplateOrJaggedClip(t *testing.T) {
	b, err := os.ReadFile("gui_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	a := strings.Index(s, "func (g *ui) drawAppIcon(")
	z := strings.Index(s[a:], "\nfunc scanInstalledPrograms()")
	if a < 0 || z < 0 {
		t.Fatal("drawAppIcon source not found")
	}
	body := s[a : a+z]
	for _, bad := range []string{"createRoundRectRgn.Call", "drawIconEx.Call", "d.rounded(x-2, y-2"} {
		if strings.Contains(body, bad) {
			t.Errorf("low-quality app icon renderer reintroduced: %s", bad)
		}
	}
	for _, need := range []string{"gdipBitmapFromHICON.Call", "gdipImageRect.Call", "gdipPixelOffsetMode.Call"} {
		if !strings.Contains(body, need) {
			t.Errorf("GDI+ high-quality icon rendering missing: %s", need)
		}
	}
	if !strings.Contains(s, "// One clean focus ring") {
		t.Error("double focus ring may have returned")
	}
}
