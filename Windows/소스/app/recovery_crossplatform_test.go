package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupNotOverwrittenByCorruption(t *testing.T) {
	p := filepath.Join(t.TempDir(), "profile.json")
	first := defaults()
	if err := saveJSON(p, first); err != nil {
		t.Fatal(err)
	}
	second := copyConfig(first)
	second.Models["12"].Profiles[0].Keys[0].Color = "빨강"
	if err := saveJSON(p, second); err != nil {
		t.Fatal(err)
	}
	bak, err := readJSON(p + ".bak")
	if err != nil || bak.Models["12"].Profiles[0].Keys[0].Color != "회색" {
		t.Fatalf("bad backup: %+v %v", bak, err)
	}
	if err := os.WriteFile(p, []byte("damaged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(p, second); err == nil {
		t.Fatal("overwrite of corrupt primary was allowed")
	}
	b, err := os.ReadFile(p + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	restored, err := readJSON(p + ".bak")
	if err != nil || restored.Models["12"].Profiles[0].Keys[0].Color != "회색" || len(b) == 0 {
		t.Fatalf("backup lost: %+v %v", restored, err)
	}
}
