package main

import "testing"

func TestInstanceAction(t *testing.T) {
	for _, tc := range []struct {
		title      string
		background bool
		want       InstanceAction
	}{
		{"PulseNote Deck 7.4.14 · 8/12 Key Controller", false, InstanceReplace},
		{"PulseNote Deck 7.4.13 · 8/12 Key Controller", false, InstanceReplace},
		{"PulseNote Deck 7.4.1 · 8/12 Key Controller", false, InstanceReplace},
		{"PulseNote Deck 7.4.14 · 8/12 Key Controller", true, InstanceSkip},
		{"PulseNote Deck 7.4.17 · 8/12 Key Controller", false, InstanceReplace},
		{"PulseNote Deck 7.4.23 · 8/12 Key Controller", false, InstanceReplace},
		{"PulseNote Deck 7.4.25 · 8/12 Key Controller", false, InstanceReplace},
		{"PulseNote Deck 7.5.2 · PN12 Controller", false, InstanceReplace},
		{"PulseNote Deck 7.5.3 · PN12 / PN15 Controller", false, InstanceReplace},
		{"PulseNote Deck 7.5.5 · PN12 / PN15 Controller", false, InstanceReplace},
		{"PulseNote Deck legacy", false, InstanceReplace},
	} {
		if got := instanceAction(tc.title, version, tc.background); got != tc.want {
			t.Fatalf("instanceAction(%q,%v) = %d want %d", tc.title, tc.background, got, tc.want)
		}
	}
}

func TestVersionParsing(t *testing.T) {
	tests := []struct {
		title string
		want  [3]int
		valid bool
	}{
		{"PulseNote Deck 7.4.13 · 8/12 Key Controller", [3]int{7, 4, 13}, true},
		{"PulseNote Deck 7.4.14 · 8/12 Key Controller", [3]int{7, 4, 14}, true},
		{"PulseNote Deck 7.4.100 · 8/12 Key Controller", [3]int{7, 4, 100}, true},
		{"Some Other App 7.4.14", [3]int{}, false},
	}
	for _, tc := range tests {
		got, valid := parseWindowVersion(tc.title)
		if got != tc.want || valid != tc.valid {
			t.Fatalf("%q got %v %v", tc.title, got, valid)
		}
	}
}
