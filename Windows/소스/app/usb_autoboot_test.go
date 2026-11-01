package main

import (
	"testing"
	"time"
)

func TestESP32CDCOpensWithoutAssertingRTS(t *testing.T) {
	flags := normalCDCFlags()
	if flags&dcbFlagDTREnable == 0 || flags&dcbFlagRTSEnable != 0 || flags&dcbFlagBinary == 0 {
		t.Fatalf("unsafe C3 modem state: %#x", flags)
	}
}
func TestRecoveryIsRestrictedAndRateLimited(t *testing.T) {
	now := time.Unix(1234567890, 0)
	for _, tc := range []struct {
		native, trusted bool
		elapsed         time.Duration
		want            bool
	}{
		{false, true, 0, false}, {true, false, 0, false}, {true, true, 0, true},
		{true, true, 5 * time.Second, false}, {true, true, bootRecoveryCooldown, true},
	} {
		last := time.Time{}
		if tc.elapsed > 0 {
			last = now.Add(-tc.elapsed)
		}
		got := allowBootRecovery(tc.native, tc.trusted, last, now)
		if got != tc.want {
			t.Fatalf("native=%v trusted=%v elapsed=%s: got %v want %v", tc.native, tc.trusted, tc.elapsed, got, tc.want)
		}
	}
}
