package main

import "testing"

func TestPhysicalGrid12IsThreeAcrossFourDown(t *testing.T) {
	g := gridForKeys(12)
	if g.Columns != 3 || g.Rows != 4 || g.CellW != g.CellH {
		t.Fatalf("12-key must be 3 columns x 4 rows, got %+v", g)
	}
	for i := 0; i < 12; i++ {
		x, y := g.pos(i)
		if x != g.StartX+(i%3)*(g.CellW+g.GapX) || y != g.StartY+(i/3)*(g.CellH+g.GapY) {
			t.Fatalf("KEY:%d position wrong", i+1)
		}
		if x < 282 || x+g.CellW > 874 || y < 285 || y+g.CellH > 762 {
			t.Fatalf("KEY:%d outside card bounds (%d,%d)", i+1, x, y)
		}
	}
	if x, y := g.pos(3); x != g.StartX || y != g.StartY+g.CellH+g.GapY {
		t.Fatal("KEY4 must begin the second row")
	}
	if x, y := g.pos(11); x != g.StartX+2*(g.CellW+g.GapX) || y != g.StartY+3*(g.CellH+g.GapY) {
		t.Fatal("KEY12 must end fourth row")
	}
}
func TestPN12OnlyGrid(t *testing.T) {
	g := gridForKeys(12)
	if g.Columns != 3 || g.Rows != 4 || g.CellW != 114 {
		t.Fatal(g)
	}
}
