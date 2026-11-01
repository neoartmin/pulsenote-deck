package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type deviceNotice struct {
	Port      string
	Keys      int
	Connected bool
}

type deviceNoticeTracker struct {
	mu     sync.Mutex
	active bool
	port   string
	keys   int
}

// A state transition, not a poll, produces a notification. Never announce
// device discovery retries or repeated INFO packets.
func (t *deviceNoticeTracker) attach(port string, keys int) (deviceNotice, bool) {
	port = strings.TrimSpace(strings.ToUpper(port))
	if port == "" {
		return deviceNotice{}, false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.active && t.port == port {
		t.keys = keys
		return deviceNotice{}, false
	}
	t.active, t.port, t.keys = true, port, keys
	return deviceNotice{Port: port, Keys: keys, Connected: true}, true
}

// silent means a deliberate software reconnect/reconfiguration; it must
// reset the state but should not produce a misleading "cable unplugged" toast.
func (t *deviceNoticeTracker) detach(silent bool) (deviceNotice, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return deviceNotice{}, false
	}
	n := deviceNotice{Port: t.port, Keys: t.keys, Connected: false}
	t.active, t.port, t.keys = false, "", 0
	return n, !silent
}

type notificationPreference struct {
	Enabled bool `json:"enabled"`
}

var notificationPref = struct {
	sync.RWMutex
	enabled bool
}{enabled: true}

func notificationPreferencePath() string {
	return filepath.Join(filepath.Dir(getConfigPath()), "device_notifications.json")
}
func notificationsEnabled() bool {
	notificationPref.RLock()
	defer notificationPref.RUnlock()
	return notificationPref.enabled
}
func loadNotificationPreference(path string) error {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		notificationPref.Lock()
		notificationPref.enabled = true
		notificationPref.Unlock()
		return nil
	}
	if err != nil {
		return err
	}
	var cfg notificationPreference
	if err := json.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("장치 알림 설정 읽기 오류: %w", err)
	}
	notificationPref.Lock()
	notificationPref.enabled = cfg.Enabled
	notificationPref.Unlock()
	return nil
}
func saveNotificationPreference(path string, enabled bool) error {
	b, err := json.Marshal(notificationPreference{Enabled: enabled})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if err := atomicWriteFile(path, b); err != nil {
		return err
	}
	notificationPref.Lock()
	notificationPref.enabled = enabled
	notificationPref.Unlock()
	return nil
}

// Each actual state transition is still tracked. Only duplicate Windows
// balloons of the same kind on the same port are rate-limited.
type noticeRateLimiter struct{ recent map[string]time.Time }

func (l *noticeRateLimiter) allow(n deviceNotice, now time.Time) bool {
	if l.recent == nil {
		l.recent = make(map[string]time.Time)
	}
	key := n.Port
	if n.Connected {
		key += "/up"
	} else {
		key += "/down"
	}
	if t, exists := l.recent[key]; exists && now.Sub(t) < 15*time.Second {
		return false
	}
	l.recent[key] = now
	if len(l.recent) > 40 {
		for k, t := range l.recent {
			if now.Sub(t) > time.Minute {
				delete(l.recent, k)
			}
		}
	}
	return true
}
