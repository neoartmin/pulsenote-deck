package main

import (
	"bytes"
	"testing"
)

func TestPN15FirmwareEmbedded(t *testing.T) {
	// Verify semantics without depending on whitespace/style of the canonical
	// Arduino sketch. The installer embeds the exact Firmware/ source bytes.
	normalized := bytes.Join(bytes.Fields(firmwarePN15), nil)
	for _, part := range [][]byte{
		[]byte("PND:INFO:15:R:C3"), []byte("PND:HELLO"),
		[]byte("ROWS[]={0,1,6}"), []byte("COLS[]={3,4,21,7,5}"),
		[]byte("waitingForRelease"),
	} {
		if !bytes.Contains(normalized, bytes.Join(bytes.Fields(part), nil)) {
			t.Fatalf("PN15 firmware missing %q", part)
		}
	}
}

func TestPN15LeftHandFirmwareIncluded(t *testing.T) {
	if !bytes.Contains(firmwarePN15Left, []byte("PND:INFO:15:L:C3")) {
		t.Fatal("PN15 LEFT firmware missing")
	}
}
