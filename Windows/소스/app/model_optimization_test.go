package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUnchangedSettingsDoNotRotateBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	first := defaults()
	if err := saveJSON(path, first); err != nil {
		t.Fatal(err)
	}
	second := copyConfig(first)
	second.Models["12"].Profiles[0].Keys[0].Color = "빨강"
	if err := saveJSON(path, second); err != nil {
		t.Fatal(err)
	}
	backupBefore, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	statBefore, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if err := saveJSON(path, second); err != nil {
		t.Fatal(err)
	}
	backupAfter, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	statAfter, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(backupBefore, backupAfter) {
		t.Fatal("no-op replaced recovery backup")
	}
	if !statBefore.ModTime().Equal(statAfter.ModTime()) {
		t.Fatal("no-op rewrote main configuration")
	}
	third := copyConfig(second)
	third.Models["12"].Profiles[0].Keys[0].Color = "보라"
	if err := saveJSON(path, third); err != nil {
		t.Fatal(err)
	}
	restored, err := readJSON(path + ".bak")
	if err != nil || restored.Models["12"].Profiles[0].Keys[0].Color != "빨강" {
		t.Fatalf("real change did not preserve previous config: %v, %+v", err, restored)
	}
}
