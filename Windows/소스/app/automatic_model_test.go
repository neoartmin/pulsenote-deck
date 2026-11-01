package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutomaticModelOnFirstStartAndLegacyPreview(t *testing.T) {
	first := defaults()
	if first.Selected != 15 || len(first.Models["12"].Profiles[0].Keys) != 12 || len(first.Models["15"].Profiles[0].Keys) != 15 {
		t.Fatalf("wrong default: %+v", first)
	}
	legacy := defaults()
	legacy.Selected = 12 // Prior versions allowed a manual selection.
	legacy.Models["12"].Profiles[0].Keys[0].Label = "SAVED PN12"
	restored := restoreAutomaticModel(legacy)
	if restored.Selected != 12 || restored.Models["12"].Profiles[0].Keys[0].Label != "SAVED PN12" {
		t.Fatal("legacy settings lost or old manual model retained")
	}
}
func TestEditorSelectionUnaffectedByLastVerifiedHardware(t *testing.T) {
	for _, model := range []int{12, 15} {
		c := defaults()
		c.Selected = 15
		c.LastVerifiedModel = model
		c = restoreAutomaticModel(c)
		if c.Selected != 15 {
			t.Fatalf("model %d not restored", model)
		}
		file := filepath.Join(t.TempDir(), "config.json")
		if err := saveJSON(file, c); err != nil {
			t.Fatal(err)
		}
		loaded, err := readJSON(file)
		if err != nil {
			t.Fatal(err)
		}
		if restoreAutomaticModel(loaded).Selected != 15 {
			t.Fatalf("model %d lost on restart", model)
		}
	}
}
func TestOfflineModelEditorHasThreeTabs(t *testing.T) {
	b, err := os.ReadFile("gui_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `case "editmodel"`) || !strings.Contains(string(b), `15 왼손`) {
		t.Fatal("offline model editor missing")
	}
}
