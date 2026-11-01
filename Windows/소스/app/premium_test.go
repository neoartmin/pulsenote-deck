package main

import (
	"os"
	"strings"
	"testing"
)

func TestPremiumModeSquareKeycaps(t *testing.T) {
	for _, n := range []int{12} {
		g := gridForKeys(n)
		if g.CellW != g.CellH {
			t.Fatalf("%d-key is not square", n)
		}
		if g.Columns*g.Rows != n {
			t.Fatalf("%d-key row/column mismatch", n)
		}
	}
}
func TestPremiumWin32TypographyAndIcons(t *testing.T) {
	b, e := os.ReadFile("gui_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	src := string(b)
	for _, s := range []string{"Pretendard", "Segoe MDL2 Assets", "d.actionArtwork(key.Type", "d.textBold(", "g.navPill(", "createCompatibleBitmap.Call"} {
		if !strings.Contains(src, s) {
			t.Fatalf("required UI feature missing: %s", s)
		}
	}
	for _, s := range []string{"ESP32-WROOM-32", "▦   버튼 설정", "▣   프로필 관리", "♫+"} {
		if strings.Contains(src, s) {
			t.Fatalf("legacy placeholder: %s", s)
		}
	}
}
func TestC3TwelveKeyInfo(t *testing.T) {
	if parseInfo("PND:INFO:12:C3") != 12 {
		t.Fatal("C3 12key metadata not recognized")
	}
}
