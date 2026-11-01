package main

import "time"

// Login autostart stays silent, and a second launch delegates to the already
// running instance rather than pretending to initialize the application again.
func shouldShowStartupSplash(background, primary bool) bool {
	return !background && primary
}

// Keep the start screen perceptible without imposing a fixed delay on slower
// machines: once setup has finished, wait only for the remaining time.
func remainingSplashTime(start time.Time, now time.Time, minimum time.Duration) time.Duration {
	if minimum <= 0 {
		return 0
	}
	remaining := minimum - now.Sub(start)
	if remaining < 0 {
		return 0
	}
	return remaining
}
