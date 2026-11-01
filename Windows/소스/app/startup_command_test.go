package main

import "testing"

func TestStartupMatchesCurrentVersionOnly(t *testing.T) {
	file := `C:\Users\User\Downloads\PulseNoteDeck_v7.4.14.exe`
	if !startupMatches(`"C:\Users\User\Downloads\PulseNoteDeck_v7.4.14.exe" --background`, file) {
		t.Fatal("valid registration not detected")
	}
	for _, cmd := range []string{`"C:\Users\User\Downloads\PulseNoteDeck_v7.4.12.exe" --background`, `"C:\Other\App.exe" --background`, `"C:\Users\User\Downloads\PulseNoteDeck_v7.4.14.exe"`, ""} {
		if startupMatches(cmd, file) {
			t.Fatal("stale registration accepted:", cmd)
		}
	}
}
