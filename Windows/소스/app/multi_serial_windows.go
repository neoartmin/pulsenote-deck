//go:build windows

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Each USB port owns its file descriptor, heartbeat, input parser and identity.
// Three firmware identities are supported at once: PN12, PN15 RIGHT, PN15 LEFT.
type deviceKeyEvent struct {
	Key, Model int
	Side, Port string
}
type deviceLink struct {
	port       string
	file       *os.File
	model      int
	side       string
	generation uint32
}
type multiSerialManager struct {
	sync.Mutex
	links  map[string]*deviceLink
	failed map[string]time.Time
}

func deviceIdentityLabel(model int, side string) string {
	if model == 12 {
		return "PN12"
	}
	if model == 15 && side == "L" {
		return "PN15 LEFT"
	}
	if model == 15 {
		return "PN15 RIGHT"
	}
	return "Unknown"
}
func (m *multiSerialManager) activeSnapshot() []deviceLink {
	m.Lock()
	defer m.Unlock()
	result := make([]deviceLink, 0, len(m.links))
	for _, l := range m.links {
		if l.model != 0 {
			result = append(result, *l)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].port < result[j].port })
	return result
}
func (m *multiSerialManager) updateAggregateStatus() {
	devices := m.activeSnapshot()
	if len(devices) == 0 {
		setState("", 0, false, "PN12 / PN15 LEFT / PN15 RIGHT · 자동 검색 중")
		return
	}
	var names []string
	total := 0
	for _, d := range devices {
		names = append(names, deviceIdentityLabel(d.model, d.side)+" "+d.port)
		total += d.model
	}
	status := strings.Join(names, " · ")
	// Info stores total physical keys; UI distinguishes multi-unit counts > 15.
	setState(status, total, true, fmt.Sprintf("%d대 연결 · 총 %d키 · %s", len(devices), total, status))
}
func (m *multiSerialManager) disconnect(port string, expected *deviceLink, reason string, manual bool) {
	m.Lock()
	current := m.links[port]
	if current != expected {
		m.Unlock()
		return
	}
	delete(m.links, port)
	if !manual {
		m.failed[port] = time.Now().Add(3 * time.Second)
	}
	connected := current.model != 0
	model := current.model
	m.Unlock()
	if current.file != nil {
		_ = current.file.Close()
	}
	if connected {
		enqueueDeviceNotice(deviceNotice{Port: port, Keys: model, Connected: false})
		logUSBEvent(port, reason)
	}
	m.updateAggregateStatus()
}
func (m *multiSerialManager) register(port string, current *deviceLink, count int, side string) bool {
	m.Lock()
	if m.links[port] != current {
		m.Unlock()
		return false
	}
	// One of each model/hardware hand avoids ambiguous shortcuts.
	for _, l := range m.links {
		if l != current && l.model == count && l.side == side {
			m.Unlock()
			return false
		}
	}
	current.model = count
	current.side = side
	delete(m.failed, port)
	m.Unlock()
	m.updateAggregateStatus()
	// Existing configuration is retained by model+side; this merely records
	// the last physically verified device for offline presentation.
	if snap().LastVerifiedModel != count || snap().LastVerifiedSide != side || snap().PreferredCOM != port {
		if err := update(func(c *Config) {
			c.LastVerifiedModel = count
			c.LastVerifiedSide = side
			c.PreferredCOM = port
		}); err != nil {
			logUSBEvent(port, "장치 정보 저장 실패: "+err.Error())
		}
	}
	enqueueDeviceNotice(deviceNotice{Port: port, Keys: count, Connected: true})
	return true
}
func (m *multiSerialManager) serve(current *deviceLink) {
	port := current.port
	defer func() {
		m.disconnect(port, current, "장치 세션 종료", reconnectFlag.Load() != current.generation)
	}()
	f, err := openCOM(port)
	if err != nil {
		logUSBEvent(port, "포트 열기 실패: "+err.Error())
		return
	}
	m.Lock()
	if m.links[port] != current {
		m.Unlock()
		f.Close()
		return
	}
	current.file = f
	m.Unlock()
	count, side, verified, reason := handshake(f)
	if !verified {
		logUSBEvent(port, "장치 식별 실패: "+reason)
		return
	}
	if reconnectFlag.Load() != current.generation {
		return
	}
	if !m.register(port, current, count, side) {
		logUSBEvent(port, "중복 장치 유형 또는 종료 중인 포트: "+deviceIdentityLabel(count, side))
		return
	}
	p := &lineParser{}
	health := newHeartbeatState(time.Now())
	buffer := make([]byte, 256)
	seen := time.Now()
	for reconnectFlag.Load() == current.generation {
		now := time.Now()
		if health.timedOut(now) {
			logUSBEvent(port, "USB 장치 응답 시간 초과")
			return
		}
		if health.shouldProbe(now) {
			if _, err = f.Write([]byte("PND:HELLO\n")); err != nil {
				return
			}
			health.probed(now)
		}
		n, err := readCOM(f, buffer)
		if err != nil {
			logUSBEvent(port, "USB 수신 오류: "+err.Error())
			return
		}
		modelChanged := false
		if n > 0 {
			p.push(buffer[:n], func(line string) {
				if line == "PND:READY:2.1" {
					health.acknowledge(time.Now())
					return
				}
				if mod, orient := parseDeviceInfo(line); mod != 0 {
					if mod != count || orient != side {
						modelChanged = true
					}
					return
				}
				if key, ok := parseKey(line); ok && key <= count {
					ev := deviceKeyEvent{Key: key, Model: count, Side: side, Port: port}
					select {
					case multiActionQueue <- ev:
					default:
						logUSBEvent(port, "입력 대기열 가득 참")
					}
				}
			})
		}
		if modelChanged {
			logUSBEvent(port, "펌웨어 장치 식별 정보 변경")
			return
		}
		if time.Since(seen) > 2500*time.Millisecond {
			if !containsPort(portList(), port) {
				return
			}
			seen = time.Now()
		}
	}
}
func multiSerialLoop() {
	m := &multiSerialManager{links: map[string]*deviceLink{}, failed: map[string]time.Time{}}
	currentRevision := reconnectFlag.Load()
	m.updateAggregateStatus()
	for {
		revision := reconnectFlag.Load()
		if revision != currentRevision {
			m.Lock()
			for _, link := range m.links {
				if link.file != nil {
					_ = link.file.Close()
				}
			}
			clear(m.failed)
			m.Unlock()
			currentRevision = revision
		}
		ports := prioritizeActivePorts(portList(), snap().PreferredCOM, "")
		for _, port := range ports {
			m.Lock()
			_, running := m.links[port]
			if !running && len(m.links) < 3 && time.Now().After(m.failed[port]) {
				link := &deviceLink{port: port, generation: revision}
				m.links[port] = link
				go m.serve(link)
			}
			m.Unlock()
		}
		time.Sleep(700 * time.Millisecond)
	}
}
