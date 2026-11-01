package main

import (
	"bytes"
	"testing"
)

func TestPN12RuntimeFirmwareAnnouncesOnUSBOpen(t *testing.T) {
	for _, part := range [][]byte{[]byte("announceWhenUSBReady"), []byte("usbSessionAnnounced"), []byte("PND:READY:2.1")} {
		if !bytes.Contains(firmwarePN12, part) {
			t.Fatalf("firmware missing startup handshake %q", part)
		}
	}
}
