//go:build windows

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigBackupValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings", "config.json")
	c := defaults()
	if err := saveJSON(path, c); err != nil {
		t.Fatal(err)
	}
	c.Models["12"].Profiles[0].Keys[0].Color = "빨강"
	if err := saveJSON(path, c); err != nil {
		t.Fatal(err)
	}
	old, err := readJSON(path + ".bak")
	if err != nil || old.Models["12"].Profiles[0].Keys[0].Color != "회색" {
		t.Fatalf("backup not preserved: %v %v", old.Selected, err)
	}
	newer, err := readJSON(path)
	if err != nil || newer.Selected != 15 {
		t.Fatalf("updated settings invalid: %v %v", newer.Selected, err)
	}
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(path, c); err == nil {
		t.Fatal("corrupt settings overwritten")
	}
}
func TestLineParserPartialOverflow(t *testing.T) {
	p := &lineParser{}
	var got []string
	emit := func(s string) { got = append(got, s) }
	p.push([]byte("KEY:"), emit)
	p.push([]byte("12\r\n"), emit)
	if len(got) != 1 || got[0] != "KEY:12" {
		t.Fatalf("partial: %v", got)
	}
	p.push(make([]byte, 201), emit)
	p.push([]byte("\nKEY:1\n"), emit)
	if len(got) != 2 || got[1] != "KEY:1" {
		t.Fatalf("overflow: %v", got)
	}
}
