package main

import (
	"testing"
	"time"
)

func TestSplashOnlyOnFreshInteractiveLaunch(t *testing.T) {
	cases := []struct {
		background, primary, want bool
	}{
		{false, true, true},
		{true, true, false},
		{false, false, false},
		{true, false, false},
	}
	for _, c := range cases {
		if got := shouldShowStartupSplash(c.background, c.primary); got != c.want {
			t.Fatalf("splash background=%v primary=%v -> %v", c.background, c.primary, got)
		}
	}
}

func TestSplashMinimumWaitOnlyAfterFastInitialization(t *testing.T) {
	start := time.Unix(0, 0)
	min := 550 * time.Millisecond
	if got := remainingSplashTime(start, start.Add(200*time.Millisecond), min); got != 350*time.Millisecond {
		t.Fatalf("fast startup wants 350ms, got %v", got)
	}
	if got := remainingSplashTime(start, start.Add(700*time.Millisecond), min); got != 0 {
		t.Fatalf("slow startup must not be delayed, got %v", got)
	}
	if got := remainingSplashTime(start, start, 0); got != 0 {
		t.Fatalf("disabled splash must have zero wait, got %v", got)
	}
}
