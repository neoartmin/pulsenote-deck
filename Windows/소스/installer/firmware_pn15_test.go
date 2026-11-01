package main

import (
	"bytes"
	"testing"
)

func TestPN15FirmwareEmbedded(t *testing.T) {
	for _, part := range [][]byte{
		[]byte("PND:INFO:15:R:C3"), []byte("PND:HELLO"),
		[]byte("ROWS[]={0,1,6}"), []byte("COLS[]={3,4,21,7,5}"),
		[]byte("waitingForRelease"),
	} {
		if !bytes.Contains(firmwarePN15, part) {
			t.Fatalf("PN15 firmware missing %q", part)
		}
	}
}

func TestPN15LeftHandFirmwareIncluded(t *testing.T) {
	if !bytes.Contains(firmwarePN15Left, []byte("PND:INFO:15:L:C3")) {
		t.Fatal("PN15 LEFT firmware missing")
	}
}
