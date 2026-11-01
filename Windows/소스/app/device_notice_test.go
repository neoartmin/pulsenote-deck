package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDeviceNoticeTransitions(t *testing.T) {
	tkr := &deviceNoticeTracker{}
	if _, ok := tkr.detach(false); ok {
		t.Fatal("initial scan must not show disconnect")
	}
	a, ok := tkr.attach("com7", 12)
	if !ok || a.Port != "COM7" || !a.Connected || a.Keys != 12 {
		t.Fatal(a, ok)
	}
	if _, ok := tkr.attach("COM7", 12); ok {
		t.Fatal("repeated discovery must not notify")
	}
	if _, ok := tkr.attach("COM7", 12); ok {
		t.Fatal("repeated info packet must not notify")
	}
	b, ok := tkr.detach(false)
	if !ok || b.Connected || b.Port != "COM7" || b.Keys != 12 {
		t.Fatal(b, ok)
	}
	if _, ok := tkr.detach(false); ok {
		t.Fatal("double detach must not notify")
	}
	if _, ok := tkr.attach("COM8", 12); !ok {
		t.Fatal("new port must notify")
	}
	if _, ok := tkr.detach(true); ok {
		t.Fatal("manual reconnect must be silent")
	}
	if _, ok := tkr.attach("COM8", 12); !ok {
		t.Fatal("reconnect must notify when ready")
	}
}
func TestNoticeSettingsPersistence(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nested", "notify.json")
	if err := loadNotificationPreference(p); err != nil {
		t.Fatal(err)
	}
	if !notificationsEnabled() {
		t.Fatal("default must be enabled")
	}
	if err := saveNotificationPreference(p, false); err != nil {
		t.Fatal(err)
	}
	if notificationsEnabled() {
		t.Fatal("disabled setting not saved")
	}
	if err := loadNotificationPreference(p); err != nil {
		t.Fatal(err)
	}
	if notificationsEnabled() {
		t.Fatal("disabled setting not restored")
	}
	if err := os.WriteFile(p, []byte("bad"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := loadNotificationPreference(p); err == nil {
		t.Fatal("invalid preference should report error")
	}
	if notificationsEnabled() {
		t.Fatal("invalid preference must not silently toggle settings")
	}
}

func TestDeviceBalloonRateLimit(t *testing.T) {
	limiter := &noticeRateLimiter{}
	t0 := time.Unix(0, 0)
	connected := deviceNotice{Port: "COM11", Connected: true}
	detached := deviceNotice{Port: "COM11", Connected: false}
	if !limiter.allow(connected, t0) {
		t.Fatal("first connection should notify")
	}
	if limiter.allow(connected, t0.Add(time.Second)) {
		t.Fatal("duplicate connection should be suppressed")
	}
	if !limiter.allow(detached, t0.Add(2*time.Second)) {
		t.Fatal("first disconnect should notify")
	}
	if limiter.allow(detached, t0.Add(3*time.Second)) {
		t.Fatal("duplicate disconnect should be suppressed")
	}
	if !limiter.allow(connected, t0.Add(16*time.Second)) {
		t.Fatal("connection cooldown should expire")
	}
}
