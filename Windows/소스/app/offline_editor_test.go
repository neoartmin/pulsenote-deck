package main

import "testing"

func TestOfflineEditingAcrossAllThreeModels(t *testing.T) {
	c := defaults()
	cases := []struct {
		model     int
		side, key string
	}{{12, "R", "12"}, {15, "L", "15L"}, {15, "R", "15"}}
	for _, tt := range cases {
		if !setEditorModel(&c, tt.model, tt.side) && tt.key != "15" {
			t.Fatal("cannot select", tt.key)
		}
		if !mutateKey(&c, 0, func(k *Key) { k.Label = "EDIT " + tt.key }) {
			t.Fatal("cannot edit", tt.key)
		}
	}
	for _, tt := range cases {
		got := c.Models[tt.key].Profiles[0].Keys[0].Label
		if got != "EDIT "+tt.key {
			t.Fatalf("%s got %q", tt.key, got)
		}
	}
	if k, ok := resolveHardwareKey(c, 12, "R", 1); !ok || k.Label != "EDIT 12" {
		t.Fatal("PN12 route changed")
	}
	if k, ok := resolveHardwareKey(c, 15, "L", 1); !ok || k.Label != "EDIT 15L" {
		t.Fatal("PN15 L route changed")
	}
	if k, ok := resolveHardwareKey(c, 15, "R", 1); !ok || k.Label != "EDIT 15" {
		t.Fatal("PN15 R route changed")
	}
}
func TestLargerKeycapsFitGrid(t *testing.T) {
	for _, n := range []int{12, 15} {
		for _, side := range []string{"R", "L"} {
			g := gridForKeys(n)
			for i := 0; i < n; i++ {
				x, y := g.posForSide(i, side)
				if x < 282 || y < 285 || x+g.CellW > 874 || y+g.CellH > 762 {
					t.Fatalf("overflow n=%d side=%s key=%d position=%d,%d", n, side, i+1, x, y)
				}
			}
		}
	}
}
