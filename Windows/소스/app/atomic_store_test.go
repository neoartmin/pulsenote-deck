package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtomicSettingsAndBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	c := defaults()
	for n := 0; n < 10; n++ {
		color := "빨강"
		if n%2 == 1 {
			color = "보라"
		}
		c.Models["12"].Profiles[0].Keys[0].Color = color
		if err := saveJSON(path, c); err != nil {
			t.Fatal(err)
		}
		got, err := readJSON(path)
		if err != nil || got.Selected != 15 || got.Models["12"].Profiles[0].Keys[0].Color != color {
			t.Fatalf("iteration %d: %v", n, err)
		}
		if n > 0 {
			bak, err := readJSON(path + ".bak")
			if err != nil || bak.Models["12"].Profiles[0].Keys[0].Color == color {
				t.Fatalf("backup iteration %d: %v", n, err)
			}
		}
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".pnd-settings-") {
			t.Fatal("stale temporary file", file.Name())
		}
	}
}
func TestRefuseToOverwriteDamagedConfigAndKeepBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	c := defaults()
	if err := saveJSON(path, c); err != nil {
		t.Fatal(err)
	}
	c.Models["12"].Profiles[0].Keys[0].Color = "빨강"
	if err := saveJSON(path, c); err != nil {
		t.Fatal(err)
	}
	backupBefore, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("damaged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(path, c); err == nil {
		t.Fatal("damaged settings overwritten")
	}
	backupAfter, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if string(backupBefore) != string(backupAfter) {
		t.Fatal("existing valid backup overwritten")
	}
}
