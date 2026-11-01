package main

import (
	"fmt"
	"strings"
	"time"
)

type connectionPhase uint8

const (
	phaseSearching connectionPhase = iota
	phaseOpening
	phaseHandshake
	phaseConnected
	phaseRetry
)

type connectionProgress struct {
	Phase     connectionPhase
	Port      string
	Keys      int
	Detail    string
	CheckedAt time.Time
	RetryAt   time.Time
}

// The displayed status reflects a verified handshake, not just a COM port
// appearing in Windows. The stages do not initiate serial operations.
func (s connectionProgress) title() string {
	switch s.Phase {
	case phaseOpening:
		return "포트 연결 중"
	case phaseHandshake:
		return "장치 확인 중"
	case phaseConnected:
		if s.Keys > 15 {
			return fmt.Sprintf("%d대 연결됨", connectedDeviceCount(s.Keys))
		}
		if s.Keys == 15 {
			return "PN15 연결됨"
		}
		return "PN12 연결됨"
	case phaseRetry:
		return "재연결 대기"
	default:
		return "장치 검색 중"
	}
}
func (s connectionProgress) subtitle() string {
	switch s.Phase {
	case phaseOpening:
		return s.Port + " · 포트 여는 중"
	case phaseHandshake:
		return s.Port + " · 응답 확인 중"
	case phaseConnected:
		if s.Keys > 15 {
			return fmt.Sprintf("총 %d키 · 동시 입력 활성", s.Keys)
		}
		if s.Keys == 15 {
			return s.Port + " · 15키 통신 중"
		}
		return s.Port + " · 12키 통신 중"
	case phaseRetry:
		if s.Port != "" {
			return s.Port + " · 자동 재시도"
		}
		return "PN12/PN15 자동 재검색"
	default:
		return "ESP32-C3 · 자동 검색 중"
	}
}
func (s connectionProgress) description(now time.Time) string {
	d := strings.TrimSpace(s.Detail)
	if d == "" {
		d = "PN12/PN15 장치를 자동으로 검색합니다"
	}
	if s.Phase == phaseRetry && !s.RetryAt.IsZero() && now.Before(s.RetryAt) {
		return fmt.Sprintf("%s · %s 재시도", d, s.RetryAt.Format("15:04:05"))
	}
	return d
}
func (s connectionProgress) checkedLabel() string {
	if s.CheckedAt.IsZero() {
		return "최근 확인: 준비 중"
	}
	return "최근 확인: " + s.CheckedAt.Format("15:04:05")
}
