package main

import (
	"strings"
	"time"
)

// prioritizeActivePorts must never turn a remembered COM number into a live
// device. Windows can retain user preferences after USB removal.
func prioritizeActivePorts(ports []string, preferred, lastGood string) []string {
	live := make(map[string]bool, len(ports))
	for _, port := range ports {
		live[strings.ToUpper(strings.TrimSpace(port))] = true
	}
	result := make([]string, 0, len(live))
	seen := make(map[string]bool, len(live))
	add := func(port string) {
		port = strings.ToUpper(strings.TrimSpace(port))
		if port != "" && live[port] && !seen[port] {
			result = append(result, port)
			seen[port] = true
		}
	}
	add(preferred)
	add(lastGood)
	for _, port := range ports {
		add(port)
	}
	return result
}

// A handshake alone does not prove a stable link: the port can disappear just
// after identification. Announce connection only after a short stable session.
const (
	deviceNoticeStableDuration = 1200 * time.Millisecond
	deviceHeartbeatInterval    = 3 * time.Second
	deviceHeartbeatTimeout     = 8 * time.Second
)

type heartbeatState struct {
	lastAcknowledged time.Time
	lastProbe        time.Time
}

func newHeartbeatState(now time.Time) heartbeatState {
	return heartbeatState{lastAcknowledged: now}
}
func (s *heartbeatState) acknowledge(now time.Time) { s.lastAcknowledged = now }
func (s *heartbeatState) shouldProbe(now time.Time) bool {
	return s.lastProbe.IsZero() || now.Sub(s.lastProbe) >= deviceHeartbeatInterval
}
func (s *heartbeatState) probed(now time.Time) { s.lastProbe = now }
func (s *heartbeatState) timedOut(now time.Time) bool {
	return now.Sub(s.lastAcknowledged) > deviceHeartbeatTimeout
}

func containsPort(ports []string, expected string) bool {
	for _, port := range ports {
		if strings.EqualFold(strings.TrimSpace(port), strings.TrimSpace(expected)) {
			return true
		}
	}
	return false
}

// COM modem lines matter on ESP32-C3: requesting DTR and RTS together
// can be interpreted by native USB/JTAG or auto-reset wiring as a boot/reset
// request. Enable only DTR for the CDC connection; never assert RTS at open.
const (
	dcbFlagBinary        uint32 = 1
	dcbFlagDTREnable     uint32 = 1 << 4
	dcbFlagRTSEnable     uint32 = 1 << 12
	bootRecoveryCooldown        = 90 * time.Second
)

func normalCDCFlags() uint32 { return dcbFlagBinary | dcbFlagDTREnable }
func allowBootRecovery(native, trusted bool, lastAttempt, now time.Time) bool {
	return native && trusted && (lastAttempt.IsZero() || now.Sub(lastAttempt) >= bootRecoveryCooldown)
}

// A valid CDC session requires *both* identification records. Receiving READY
// before INFO is normal on slower USB hosts, so no fixed short grace period.
type handshakeVerifier struct {
	ready       bool
	model       int
	side        string
	conflicting bool
}

func (h *handshakeVerifier) accept(line string) {
	if line == "PND:READY:2.1" {
		h.ready = true
	}
	if n, side := parseDeviceInfo(line); n != 0 {
		if h.model != 0 && (h.model != n || h.side != side) {
			h.conflicting = true
		}
		h.model = n
		h.side = side
	}
}
func (h handshakeVerifier) verified() bool {
	return h.ready && countValid(h.model) && !h.conflicting
}
