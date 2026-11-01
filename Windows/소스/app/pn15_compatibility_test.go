package main

import (
	"path/filepath"
	"testing"
)

func TestPN15ProtocolAndRange(t *testing.T) {
	if parseInfo("PND:INFO:15:C3") != 15 {
		t.Fatal("PN15 identification missing")
	}
	if parseInfo("PND:INFO:12:C3") != 12 {
		t.Fatal("PN12 recognition regressed")
	}
	if parseInfo("PND:INFO:8:C3") != 0 {
		t.Fatal("PN8 must stay unsupported")
	}
	for k := 1; k <= 15; k++ {
		key, ok := parseKey("KEY:" + fmtKey(k))
		if !ok || key != k {
			t.Fatalf("KEY:%d rejected", k)
		}
	}
	for _, s := range []string{"KEY:0", "KEY:16", "KEY:-1", "KEY:11X", "KEY:999", "BTN:1"} {
		if _, ok := parseKey(s); ok {
			t.Fatalf("unexpected accepted key %s", s)
		}
	}
}
func fmtKey(n int) string {
	if n >= 10 {
		return string([]byte{byte('0' + n/10), byte('0' + n%10)})
	}
	return string([]byte{byte('0' + n)})
}
func TestPN15PhysicalGridDescendsRight(t *testing.T) {
	g := gridForKeys(15)
	if g.Columns != 5 || g.Rows != 3 || g.ColumnDrop <= 0 {
		t.Fatalf("invalid PN15 grid: %+v", g)
	}
	for i := 0; i < 15; i++ {
		x, y := g.pos(i)
		if x < 282 || x+g.CellW > 874 || y < 285 || y+g.CellH > 762 {
			t.Fatalf("PN15 KEY:%02d beyond layout: %d,%d", i+1, x, y)
		}
	}
	x0, y0 := g.pos(0)
	x4, y4 := g.pos(4)
	if x4 <= x0 || y4 <= y0 {
		t.Fatalf("right-descending layout absent: %d,%d to %d,%d", x0, y0, x4, y4)
	}
	x5, y5 := g.pos(5)
	if x5 != x0 || y5 <= y0 {
		t.Fatal("second row incorrect")
	}
}
func TestPN15SeparateProfilesAndPN12Unchanged(t *testing.T) {
	c := defaults()
	c.Models["12"].Profiles[0].Keys[0].Label = "PN12 KEEP"
	c.Models["12"].Profiles[0].Keys[11].Value = "CTRL+S"
	c.Selected = 15
	c = normalize(c)
	if !countValid(15) || countValid(8) {
		t.Fatal("supported device list wrong")
	}
	if c.Selected != 15 || len(c.Models["15"].Profiles[0].Keys) != 15 {
		t.Fatal("PN15 model not initialized")
	}
	if c.Models["12"].Profiles[0].Keys[0].Label != "PN12 KEEP" {
		t.Fatal("PN12 modified")
	}
	if !mutateKey(&c, 14, func(k *Key) { k.Label = "PN15 only" }) {
		t.Fatal("15th key not editable")
	}
	file := filepath.Join(t.TempDir(), "config.json")
	if e := saveJSON(file, c); e != nil {
		t.Fatal(e)
	}
	round, e := readJSON(file)
	if e != nil {
		t.Fatal(e)
	}
	if round.Models["15"].Profiles[0].Keys[14].Label != "PN15 only" {
		t.Fatal("PN15 lost")
	}
	if round.Models["12"].Profiles[0].Keys[11].Value != "CTRL+S" {
		t.Fatal("PN12 changed")
	}
	round.Selected = 12
	round = normalize(round)
	if round.Models["15"].Profiles[0].Keys[14].Label != "PN15 only" {
		t.Fatal("PN15 data not kept after switching")
	}
}
