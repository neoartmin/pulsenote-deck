package main

import "testing"

func TestThreePhysicalDeviceRoutesDoNotShareKeys(t *testing.T) {
	c := defaults()
	for _, tc := range []struct {
		n           int
		side, label string
	}{
		{12, "R", "pn12"}, {15, "R", "right"}, {15, "L", "left"},
	} {
		model := c.Models[modelKeyFor(tc.n, tc.side)]
		model.Profiles[0].Keys[0].Label = tc.label
		c.Models[modelKeyFor(tc.n, tc.side)] = model
	}
	// GUI currently showing another device must not change event routing.
	c.Selected = 15
	c.SelectedSide = "L"
	for _, tc := range []struct {
		n          int
		side, want string
	}{
		{12, "R", "pn12"}, {15, "R", "right"}, {15, "L", "left"},
	} {
		k, ok := resolveHardwareKey(c, tc.n, tc.side, 1)
		if !ok || k.Label != tc.want {
			t.Fatalf("wrong profile for %d/%s: %+v", tc.n, tc.side, k)
		}
	}
	for _, key := range []int{0, 13, 16} {
		_, ok := resolveHardwareKey(c, 12, "R", key)
		if ok {
			t.Fatalf("accepted invalid PN12 key %d", key)
		}
	}
}
