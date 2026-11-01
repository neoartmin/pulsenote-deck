package main

import (
	"bytes"
	"testing"
)

func TestPN12FirmwareOnly(t *testing.T) {
	if !bytes.Contains(firmwarePN12, []byte("PND:INFO:12:C3")) {
		t.Fatal("PN12 firmware identification missing")
	}
	if !bytes.Contains(firmwareWiring, []byte("GPIO21")) || !bytes.Contains(firmwarePN12, []byte("{3,4,21}")) {
		t.Fatal("PN12 wiring guide missing")
	}
}
