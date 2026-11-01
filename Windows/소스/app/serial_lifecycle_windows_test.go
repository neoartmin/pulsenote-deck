//go:build windows

package main

import (
	"os"
	"strings"
	"testing"
)

// A reopened window must wake discovery without disconnecting any
// currently connected USB session.
func TestWindowReopenDoesNotRequestHardReconnect(t *testing.T) {
	gui, err := os.ReadFile("gui_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(gui)
	start := strings.Index(source, "func (g *ui) show()")
	end := strings.Index(source[start:], "func (g *ui) tray(")
	if start < 0 || end < 0 {
		t.Fatal("show/tray functions not found")
	}
	body := source[start : start+end]
	if !strings.Contains(body, "requestDiscoveryWake()") {
		t.Fatal("reopen must wake discovery")
	}
	if strings.Contains(body, "requestReconnect()") {
		t.Fatal("reopen must not force disconnect")
	}
}

func TestDiscoveryAndReconnectSignalsAreIndependent(t *testing.T) {
	before := reconnectFlag.Load()
	wakeBefore := discoveryWakeFlag.Load()
	requestDiscoveryWake()
	if reconnectFlag.Load() != before {
		t.Fatal("discovery wake forced disconnect")
	}
	if discoveryWakeFlag.Load() != wakeBefore+1 {
		t.Fatal("discovery wake signal missing")
	}
}
