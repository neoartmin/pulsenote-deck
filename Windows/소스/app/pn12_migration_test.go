package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExistingPN12ProfilesSurviveSingleModelMigration(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.json")
	original := defaults()
	original.Selected = 8
	original.Models["8"] = newModel(8)
	original.Models["12"].Profiles[0].Keys[2] = Key{Label: "실제 버튼", Type: "shortcut", Value: "CTRL+SHIFT+S", Color: "보라"}
	original.Models["8"].Profiles[0].Keys[1].Label = "historical"
	b, e := json.Marshal(original)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p, b, 0600); e != nil {
		t.Fatal(e)
	}
	got, e := readJSON(p)
	if e != nil {
		t.Fatal(e)
	}
	if got.Selected != 15 || got.Models["12"].Profiles[0].Keys[2].Value != "CTRL+SHIFT+S" {
		t.Fatal("lost PN12 key configuration")
	}
	if got.Models["8"].Profiles[0].Keys[1].Label != "historical" {
		t.Fatal("legacy data erased")
	}
	if err := saveJSON(p, got); err != nil {
		t.Fatal(err)
	}
	again, e := readJSON(p)
	if e != nil || again.Models["12"].Profiles[0].Keys[2].Color != "보라" {
		t.Fatal("round-trip failed", e)
	}
}
func TestPN12OnlyRejectsPN8DeviceInfo(t *testing.T) {
	if parseInfo("PND:INFO:8:C3") != 0 || parseInfo("PND:INFO:12:C3") != 12 {
		t.Fatal("PN8 incorrectly recognized")
	}
}

func TestPN12UIHasNoModelSwitch(t *testing.T) {
	src, e := os.ReadFile("gui_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(src, []byte(`"8키"`)) || bytes.Contains(src, []byte(`"mode", 8`)) || bytes.Contains(src, []byte(`"model", 0`)) {
		t.Fatal("old 8-key selector still accessible")
	}
}
