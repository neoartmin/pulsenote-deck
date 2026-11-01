package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallerFileReplaceAndNoStrayTemp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "PulseNoteDeck.exe")
	if err := writeFile(path, []byte("old")); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(path, []byte("new binary")); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(path, []byte("new binary")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, []byte("new binary")) {
		t.Fatalf("invalid replacement: %v, %q", err, got)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("stale installation temporary files: %v", entries)
	}
}
