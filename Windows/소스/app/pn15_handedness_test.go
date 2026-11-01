package main

import "testing"

func TestPN15FirmwareHandednessIdentification(t *testing.T) {
	for _, tc := range []struct {
		info  string
		count int
		side  string
	}{
		{"PND:INFO:12:C3", 12, "R"},
		{"PND:INFO:15:C3", 15, "R"},
		{"PND:INFO:15:R:C3", 15, "R"},
		{"PND:INFO:15:L:C3", 15, "L"},
		{"PND:INFO:15:X:C3", 0, ""},
	} {
		n, s := parseDeviceInfo(tc.info)
		if n != tc.count || s != tc.side {
			t.Fatalf("%s -> %d/%s want %d/%s", tc.info, n, s, tc.count, tc.side)
		}
	}
}
func TestPN15HandednessHandshakeStrict(t *testing.T) {
	for _, hand := range []string{"L", "R"} {
		var h handshakeVerifier
		h.accept("PND:READY:2.1")
		h.accept("PND:INFO:15:" + hand + ":C3")
		if !h.verified() || h.side != hand {
			t.Fatal("hand verification failed", hand)
		}
		h.accept("PND:INFO:15:C3")
		if hand == "L" && !h.conflicting {
			t.Fatal("hand mismatch accepted")
		}
	}
}
func TestPN15IndependentSideSettingsAndGrid(t *testing.T) {
	c := defaults()
	c.Selected = 15
	c.SelectedSide = "R"
	c = normalize(c)
	if !mutateKey(&c, 0, func(k *Key) { k.Label = "RIGHT" }) {
		t.Fatal("R save")
	}
	c.SelectedSide = "L"
	c = normalize(c)
	if !mutateKey(&c, 0, func(k *Key) { k.Label = "LEFT" }) {
		t.Fatal("L save")
	}
	if c.Models["15"].Profiles[0].Keys[0].Label != "RIGHT" || c.Models["15L"].Profiles[0].Keys[0].Label != "LEFT" {
		t.Fatal("hand profiles not isolated")
	}
	g := gridForKeys(15)
	_, r0 := g.posForSide(0, "R")
	_, r4 := g.posForSide(4, "R")
	_, l0 := g.posForSide(0, "L")
	_, l4 := g.posForSide(4, "L")
	if r0 >= r4 || l0 <= l4 {
		t.Fatal("not mirrored")
	}
	c.LastVerifiedModel = 15
	c.LastVerifiedSide = "L"
	c = restoreAutomaticModel(c)
	if c.Selected != 15 || c.SelectedSide != "L" {
		t.Fatal("restore hand failed")
	}
}
