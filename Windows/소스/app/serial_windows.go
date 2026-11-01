//go:build windows

package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

var adv = syscall.NewLazyDLL("advapi32.dll")
var regEnumValue = adv.NewProc("RegEnumValueW")
var getCommState = k32.NewProc("GetCommState")
var setCommState = k32.NewProc("SetCommState")
var setCommTimeouts = k32.NewProc("SetCommTimeouts")
var escapeCommFunction = k32.NewProc("EscapeCommFunction")
var reconnectFlag atomic.Uint32

// A window reopen is a request to refresh discovery, NOT permission to
// close a working USB session. Keep this generation separate from manual
// reconnects, which intentionally restart the serial session.
var discoveryWakeFlag atomic.Uint32

func requestReconnect()     { reconnectFlag.Add(1) }
func requestDiscoveryWake() { discoveryWakeFlag.Add(1) }

type dcb struct {
	Len, Baud, Flags                                                           uint32
	Reserved, XonLim, XoffLim                                                  uint16
	ByteSize, Parity, StopBits, XonChar, XoffChar, ErrorChar, EofChar, EvtChar byte
	Reserved2                                                                  uint16
}
type commTimeouts struct{ ReadInterval, ReadTotalMultiplier, ReadTotalConstant, WriteTotalMultiplier, WriteTotalConstant uint32 }

func openCOM(p string) (*os.File, error) {
	n, e := syscall.UTF16PtrFromString(`\\.\` + p)
	if e != nil {
		return nil, e
	}
	h, e := syscall.CreateFile(n, syscall.GENERIC_READ|syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if e != nil {
		return nil, e
	}
	f := os.NewFile(uintptr(h), p)
	d := dcb{Len: uint32(unsafe.Sizeof(dcb{}))}
	ret, _, err := getCommState.Call(uintptr(h), uintptr(unsafe.Pointer(&d)))
	if ret == 0 {
		f.Close()
		return nil, fmt.Errorf("GetCommState: %v", err)
	}
	d.Baud = 115200
	d.ByteSize = 8
	d.Parity = 0
	d.StopBits = 0
	// DTR for Arduino USB CDC, but keep RTS *deasserted*. Simultaneously
	// asserting both lines can interfere with C3 normal boot.
	d.Flags = normalCDCFlags()
	ret, _, err = setCommState.Call(uintptr(h), uintptr(unsafe.Pointer(&d)))
	if ret == 0 {
		f.Close()
		return nil, fmt.Errorf("SetCommState: %v", err)
	}
	tm := commTimeouts{ReadInterval: 100, ReadTotalConstant: 120, WriteTotalConstant: 350}
	ret, _, err = setCommTimeouts.Call(uintptr(h), uintptr(unsafe.Pointer(&tm)))
	if ret == 0 {
		f.Close()
		return nil, fmt.Errorf("SetCommTimeouts: %v", err)
	}
	return f, nil
}
func portList() []string {
	var h syscall.Handle
	e := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, syscall.StringToUTF16Ptr(`HARDWARE\DEVICEMAP\SERIALCOMM`), 0, syscall.KEY_READ, &h)
	if e != nil {
		return nil
	}
	defer syscall.RegCloseKey(h)
	ports := []string{}
	seen := map[string]bool{}
	for i := uint32(0); i < 100; i++ {
		name := make([]uint16, 512)
		nlen := uint32(len(name))
		data := make([]byte, 1024)
		dlen := uint32(len(data))
		var typ uint32
		ret, _, _ := regEnumValue.Call(uintptr(h), uintptr(i), uintptr(unsafe.Pointer(&name[0])), uintptr(unsafe.Pointer(&nlen)), 0, uintptr(unsafe.Pointer(&typ)), uintptr(unsafe.Pointer(&data[0])), uintptr(unsafe.Pointer(&dlen)))
		if ret == 259 {
			break
		}
		if ret != 0 || typ != 1 {
			continue
		}
		dev := strings.ToUpper(syscall.UTF16ToString(name[:nlen]))
		if strings.Contains(dev, "BTH") || strings.Contains(dev, "BLUETOOTH") || strings.Contains(dev, "RFCOMM") {
			continue
		}
		if dlen < 2 {
			continue
		}
		chars := make([]uint16, dlen/2)
		for j := range chars {
			chars[j] = uint16(data[j*2]) | uint16(data[j*2+1])<<8
		}
		p := strings.ToUpper(syscall.UTF16ToString(chars))
		if !strings.HasPrefix(p, "COM") || len(p) > 6 || seen[p] {
			continue
		}
		seen[p] = true
		ports = append(ports, p)
	}
	sort.Slice(ports, func(i, j int) bool {
		var a, b int
		fmt.Sscanf(ports[i], "COM%d", &a)
		fmt.Sscanf(ports[j], "COM%d", &b)
		return a < b
	})
	return ports
}

type lineParser struct {
	partial  []byte
	overflow bool
}

func (p *lineParser) push(b []byte, handle func(string)) {
	for _, c := range b {
		if c == '\n' {
			if !p.overflow {
				handle(strings.TrimSpace(string(p.partial)))
			}
			p.partial = p.partial[:0]
			p.overflow = false
		} else if c != '\r' && !p.overflow {
			if len(p.partial) < 200 {
				p.partial = append(p.partial, c)
			} else {
				p.partial = p.partial[:0]
				p.overflow = true
			}
		}
	}
}
func handshake(f *os.File) (int, string, bool, string) {
	parser := &lineParser{}
	verifier := &handshakeVerifier{}
	deadline := time.Now().Add(7500 * time.Millisecond)
	lastSent := time.Time{}
	buf := make([]byte, 256)
	// ESP32-C3 may reset when USB CDC opens; do not mistake boot delay for loss.
	time.Sleep(350 * time.Millisecond)
	for time.Now().Before(deadline) {
		if time.Since(lastSent) >= 500*time.Millisecond {
			if _, err := f.Write([]byte("PND:HELLO\n")); err != nil {
				return 0, "", false, "USB 식별 요청 송신 오류: " + err.Error()
			}
			lastSent = time.Now()
		}
		n, err := readCOM(f, buf)
		if err != nil {
			return 0, "", false, "USB 수신 오류: " + err.Error()
		}
		if n > 0 {
			parser.push(buf[:n], verifier.accept)
			if verifier.conflicting {
				return 0, "", false, "장치 정보가 연결 확인 중 변경됨"
			}
			if verifier.verified() {
				return verifier.model, verifier.side, true, ""
			}
		}
	}
	if !verifier.ready {
		return 0, "", false, "PND:READY:2.1 응답 없음"
	}
	return 0, "", false, "PN12/PN15 장치 정보 불일치 또는 INFO 응답 없음"
}
func connectedSession(f *os.File, port string, hardwareCount int, hardwareSide string, reconn uint32) (bool, string) {
	setState(port, hardwareCount, true, "USB 연결됨 · "+port)
	parser := &lineParser{}
	buf := make([]byte, 256)
	started := time.Now()
	lastCheck := started
	health := newHeartbeatState(started)
	currentCount := hardwareCount
	announced := false
	sawHeartbeatReply := false
	missingPortChecks := 0
	for {
		if reconnectFlag.Load() != reconn {
			return announced, "사용자 재검색"
		}
		now := time.Now()
		if !announced && sawHeartbeatReply && now.Sub(started) >= deviceNoticeStableDuration {
			recordDeviceAttached(port, currentCount)
			announced = true
		}
		// PN12 firmware replies to HELLO, including while idle.
		// Thus a stale COM handle cannot remain connected indefinitely.
		if health.timedOut(now) {
			return announced, "장치 응답 시간 초과"
		}
		if health.shouldProbe(now) {
			if _, err := f.Write([]byte("PND:HELLO\n")); err != nil {
				return announced, "USB 송신 오류: " + err.Error()
			}
			health.probed(now)
		}
		n, e := readCOM(f, buf)
		if e != nil {
			return announced, "USB 수신 오류: " + e.Error()
		}
		if n > 0 {
			modelMismatch := false
			parser.push(buf[:n], func(s string) {
				if s == "PND:READY:2.1" {
					health.acknowledge(time.Now())
					recordConnectionReply()
					sawHeartbeatReply = true
					return
				}
				if cnt, side := parseDeviceInfo(s); cnt != 0 {
					if cnt != hardwareCount || side != hardwareSide {
						modelMismatch = true
						return
					}
					currentCount = cnt
					// Avoid needless window refreshes for every heartbeat.
					return
				}
				if key, ok := parseKey(s); ok {
					// Never dispatch a PN15 key through an unrelated PN12 profile.
					if key > currentCount || snap().Selected != currentCount {
						return
					}
					select {
					case actionQueue <- key:
					default:
						setState(port, currentCount, true, "버튼 입력 대기열 가득 참")
					}
				}
			})
			if modelMismatch {
				return announced, "장치 모델 정보가 연결 중 변경됨"
			}
		}
		if time.Since(lastCheck) > 2500*time.Millisecond {
			if !containsPort(portList(), port) {
				missingPortChecks++
				if missingPortChecks >= 2 {
					return announced, "Windows COM 포트 목록에서 사라짐"
				}
			} else {
				missingPortChecks = 0
			}
			lastCheck = time.Now()
		}
	}
}

// Device discovery runs independently of each active COM connection.
func serialLoop() { multiSerialLoop() }

var _ = bytes.Equal
