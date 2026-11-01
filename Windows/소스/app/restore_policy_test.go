package main

import "testing"

func TestShouldRescanOnShow(t *testing.T) {
	tests := []struct {
		name, port string
		connected  bool
		live       []string
		want       bool
	}{
		{"connected and still present", "COM11", true, []string{"COM3", "COM11"}, false},
		{"case insensitive live port", "com11", true, []string{"COM11"}, false},
		{"disconnected", "", false, []string{"COM11"}, true},
		{"stale connected COM", "COM11", true, []string{"COM8"}, true},
		{"no COM devices", "COM11", true, nil, true},
		{"empty connected port", "", true, []string{"COM11"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRescanOnShow(tt.port, tt.connected, tt.live); got != tt.want {
				t.Fatalf("rescan=%v want=%v", got, tt.want)
			}
		})
	}
}
