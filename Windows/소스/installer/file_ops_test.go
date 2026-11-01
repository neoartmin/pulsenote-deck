package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveInstalledFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "item.exe")
	if err := removeInstalledFile(path); err != nil {
		t.Fatal("missing file", err)
	}
	if err := os.WriteFile(path, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := removeInstalledFile(path); err != nil {
		t.Fatal(err)
	}
	if err := removeInstalledFile(path); err != nil {
		t.Fatal("repeated removal", err)
	}
	if err := removeInstalledFile(filepath.Dir(path)); err == nil {
		t.Fatal("directory deletion must report an error")
	}
}
