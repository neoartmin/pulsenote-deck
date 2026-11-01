package main

import (
	"reflect"
	"testing"
	"time"
)

func TestPortPrioritizationNeverInventsDevice(t *testing.T) {
	if got := prioritizeActivePorts(nil, "COM11", "COM11"); len(got) != 0 {
		t.Fatalf("unplugged COM must not be scanned: %v", got)
	}
	got := prioritizeActivePorts([]string{"COM3", "COM7"}, "COM11", "COM3")
	want := []string{"COM3", "COM7"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	got = prioritizeActivePorts([]string{"com7", "COM3", "com7"}, "COM7", "COM3")
	want = []string{"COM7", "COM3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
func TestHeartbeatRequiresRecentDeviceReply(t *testing.T) {
	now := time.Unix(1730000000, 0)
	h := newHeartbeatState(now)
	if !h.shouldProbe(now) || h.timedOut(now) {
		t.Fatal("initial probe or timeout wrong")
	}
	h.probed(now)
	if h.shouldProbe(now.Add(2 * time.Second)) {
		t.Fatal("premature heartbeat")
	}
	if !h.shouldProbe(now.Add(3 * time.Second)) {
		t.Fatal("heartbeat not due")
	}
	h.acknowledge(now.Add(5 * time.Second))
	if h.timedOut(now.Add(12 * time.Second)) {
		t.Fatal("recent reply should keep session")
	}
	if !h.timedOut(now.Add(14 * time.Second)) {
		t.Fatal("unresponsive session should end")
	}
}
