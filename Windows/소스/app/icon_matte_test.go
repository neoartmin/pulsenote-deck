package main

import "testing"

func fixtureBGRA(w, h int, matte [4]byte) []byte {
	p := make([]byte, w*h*4)
	for i := 0; i < len(p); i += 4 {
		copy(p[i:i+4], matte[:])
	}
	return p
}
func setPx(p []byte, w, x, y int, c [4]byte) { copy(p[(y*w+x)*4:], c[:]) }
func TestRecoverBlackMatteWithoutErasingLogo(t *testing.T) {
	const w, h = 32, 32
	p := fixtureBGRA(w, h, [4]byte{1, 1, 1, 255})
	for y := 6; y < 26; y++ {
		for x := 6; x < 26; x++ {
			setPx(p, w, x, y, [4]byte{148, 156, 12, 255})
		}
	}
	// A black feature INSIDE the logo must remain opaque.
	setPx(p, w, 12, 12, [4]byte{0, 0, 0, 255})
	if !recoverOpaqueBlackMatteBGRA(p, w, h, w*4) {
		t.Fatal("black border matte not recovered")
	}
	if p[3] != 0 {
		t.Fatal("corner not transparent")
	}
	if p[(12*w+12)*4+3] != 255 {
		t.Fatal("internal dark logo detail was erased")
	}
	if p[(10*w+10)*4+3] != 255 {
		t.Fatal("logo color was erased")
	}
}
func TestNoEraseForColoredSquareOrAlreadyTransparentLogo(t *testing.T) {
	for _, bg := range [][4]byte{{80, 170, 70, 255}, {0, 0, 0, 0}, {30, 30, 30, 100}} {
		p := fixtureBGRA(32, 32, bg)
		for y := 7; y < 25; y++ {
			for x := 7; x < 25; x++ {
				setPx(p, 32, x, y, [4]byte{130, 200, 230, 255})
			}
		}
		if recoverOpaqueBlackMatteBGRA(p, 32, 32, 32*4) {
			t.Fatalf("mutated uncertain icon with background %v", bg)
		}
	}
}
func TestRejectMalformedMatteBuffer(t *testing.T) {
	if recoverOpaqueBlackMatteBGRA(make([]byte, 12), 64, 64, 256) {
		t.Fatal("accepted truncated pixel buffer")
	}
}
