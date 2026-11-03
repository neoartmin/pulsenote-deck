package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllModelsAutoDetectAndPreservePN12Settings(t *testing.T) {
	c := defaults()
	if c.Selected != 15 || len(c.Models["12"].Profiles[0].Keys) != 12 || len(c.Models["15"].Profiles[0].Keys) != 15 {
		t.Fatal("PN15 must be initial model and PN12 must remain available")
	}
	c.Selected = 12
	mutateKey(&c, 11, func(k *Key) { k.Color = "노랑"; k.Type = "website"; k.Value = "https://example.org" })
	legacy := newModel(8)
	legacy.Profiles[0].Keys[0].Value = "KEEP_FOR_ARCHIVE"
	c.Models["8"] = legacy
	c.Selected = 8 // old installed configuration
	c = normalize(c)
	if c.Selected != 15 {
		t.Fatal("legacy unsupported selection should default to PN15")
	}
	if c.Models["12"].Profiles[0].Keys[11].Value != "https://example.org" {
		t.Fatal("PN12 configuration was lost")
	}
	if c.Models["8"].Profiles[0].Keys[0].Value != "KEEP_FOR_ARCHIVE" {
		t.Fatal("old data must not be destroyed")
	}
	if countValid(8) || !countValid(12) {
		t.Fatal("8-key runtime must be disabled")
	}
}
func TestSaveLoadColors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.json")
	c := defaults()
	c.Selected = 12
	mutateKey(&c, 7, func(k *Key) { k.Color = "보라" })
	if e := saveJSON(path, c); e != nil {
		t.Fatal(e)
	}
	got, e := readJSON(path)
	if e != nil || got.Models["12"].Profiles[0].Keys[7].Color != "보라" {
		t.Fatal(got, e)
	}
}
func TestSevenColors(t *testing.T) {
	if len(palette) != 7 {
		t.Fatal(palette)
	}
	for _, s := range palette {
		if !isColor(s) {
			t.Fatal(s)
		}
	}
	if isColor("갈색") {
		t.Fatal("invalid")
	}
}
func TestProtocol(t *testing.T) {
	if parseInfo("PND:INFO:12:C3") != 12 || parseInfo("PND:INFO:8:C3") != 0 {
		t.Fatal("info")
	}
	if n, ok := parseKey("KEY:12"); !ok || n != 12 {
		t.Fatal("key")
	}
	if _, ok := parseKey("KEY:16"); ok {
		t.Fatal("bound")
	}
}
func TestLegacyDoesNotOverwriteOriginal(t *testing.T) {
	p := filepath.Join(t.TempDir(), "old.json")
	raw := []byte(`{"profiles":[{"name":"Old","keys":[{"label":"A","type":"shortcut","value":"CTRL+S"}]}]}`)
	os.WriteFile(p, raw, 0600)
	c := defaults()
	if e := importLegacy(&c, p); e != nil {
		t.Fatal(e)
	}
	c = normalize(c)
	b, _ := os.ReadFile(p)
	if string(raw) != string(b) {
		t.Fatal("legacy modified")
	}
	if c.Models["12"].Profiles[0].Keys[0].Value != "CTRL+S" {
		t.Fatal("migration")
	}
}

func TestSaveLoadMultilineTextForEachDevice(t *testing.T) {
    sample := "첫 줄\n두 번째 줄 😀\n끝"
    for _, tc := range []struct {
        count int
        side string
        model string
    }{{12, "R", "12"}, {15, "L", "15L"}, {15, "R", "15"}} {
        c := defaults()
        c.Selected, c.SelectedSide = tc.count, tc.side
        if !mutateKey(&c, 0, func(k *Key) {
            k.Type, k.Value = "text", sample
        }) {
            t.Fatalf("could not edit model %s", tc.model)
        }
        path := filepath.Join(t.TempDir(), tc.model+".json")
        if err := saveJSON(path, c); err != nil {
            t.Fatal(err)
        }
        read, err := readJSON(path)
        if err != nil {
            t.Fatal(err)
        }
        key := read.Models[tc.model].Profiles[0].Keys[0]
        if key.Type != "text" || key.Value != sample {
            t.Fatalf("model %s: got %#v", tc.model, key)
        }
    }
}
