package main

import (
	"encoding/binary"
	"hash/crc32"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestRejectOversizedPNGBeforeDecompression(t *testing.T) {
	// Fully valid IHDR metadata, with dimensions far beyond the supported
	// icon size. DecodeConfig must recognize it without decompressing pixels.
	b := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	chunk := make([]byte, 4+4+13+4)
	binary.BigEndian.PutUint32(chunk[0:4], 13)
	copy(chunk[4:8], "IHDR")
	binary.BigEndian.PutUint32(chunk[8:12], 50000)
	binary.BigEndian.PutUint32(chunk[12:16], 50000)
	chunk[16], chunk[17] = 8, 6
	binary.BigEndian.PutUint32(chunk[len(chunk)-4:], crc32.ChecksumIEEE(chunk[4:len(chunk)-4]))
	b = append(b, chunk...)
	path := filepath.Join(t.TempDir(), "huge.png")
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := png.DecodeConfig(f)
	f.Close()
	if err != nil || info.Width != 50000 {
		t.Fatalf("expected valid header metadata: %+v, %v", info, err)
	}
	if pngHasTransparency(path) {
		t.Fatal("unsupported enormous PNG accepted")
	}
}
