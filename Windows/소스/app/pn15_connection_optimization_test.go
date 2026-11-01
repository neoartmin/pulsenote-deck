package main

import (
	"testing"
)

func TestHandshakeVerifierFragmentAndOrdering(t *testing.T) {
	for _, frames := range [][]string{
		{"PND:READY:2.1", "PND:INFO:15:C3"},
		{"PND:INFO:15:C3", "PND:READY:2.1"},
		{"PND:READY:2.1", "noise", "PND:INFO:15:C3"},
	} {
		var h handshakeVerifier
		for i, line := range frames {
			h.accept(line)
			if i < len(frames)-1 && h.verified() {
				t.Fatal("premature verification")
			}
		}
		if !h.verified() || h.model != 15 {
			t.Fatalf("did not verify: %+v", h)
		}
	}
	var old handshakeVerifier
	old.accept("PND:INFO:12:C3")
	old.accept("PND:READY:2.1")
	if !old.verified() || old.model != 12 {
		t.Fatal("PN12 regression")
	}
}
func TestHandshakeVerifierRejectsContradictoryModels(t *testing.T) {
	var h handshakeVerifier
	h.accept("PND:READY:2.1")
	h.accept("PND:INFO:15:C3")
	h.accept("PND:INFO:12:C3")
	if h.verified() || !h.conflicting {
		t.Fatalf("model changed during handshake: %+v", h)
	}
}
func TestPN15ConnectionStatusDisplay(t *testing.T) {
	p := connectionProgress{Phase: phaseConnected, Port: "COM8", Keys: 15}
	if p.title() != "PN15 연결됨" || p.subtitle() != "COM8 · 15키 통신 중" {
		t.Fatalf("wrong PN15 display: %q %q", p.title(), p.subtitle())
	}
}
