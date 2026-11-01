package main

import (
	"fmt"
	"strings"
)

// InstanceAction controls what a second launch does.
type InstanceAction uint8

const (
	InstanceSkip    InstanceAction = iota // Login background launch: leave existing app untouched.
	InstanceShow                          // Same or newer version: reveal the running settings window.
	InstanceReplace                       // Older version: request graceful exit and take over.
)

func parseWindowVersion(title string) ([3]int, bool) {
	var v [3]int
	if !strings.HasPrefix(title, "PulseNote Deck ") {
		return v, false
	}
	n, err := fmt.Sscanf(title, "PulseNote Deck %d.%d.%d", &v[0], &v[1], &v[2])
	return v, n == 3 && err == nil
}

func compareVersion(a, b [3]int) int {
	for i := range a {
		if a[i] > b[i] {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
	}
	return 0
}

func instanceAction(title, desiredVersion string, background bool) InstanceAction {
	if background {
		return InstanceSkip
	}
	running, okRunning := parseWindowVersion(title)
	wanted, okWanted := parseWindowVersion("PulseNote Deck " + desiredVersion)
	if okRunning && okWanted && compareVersion(running, wanted) >= 0 {
		return InstanceShow
	}
	return InstanceReplace
}
