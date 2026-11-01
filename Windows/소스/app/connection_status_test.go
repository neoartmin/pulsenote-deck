package main

import (
	"strings"
	"testing"
	"time"
)

func TestConnectionPhaseDisplay(t *testing.T) {
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)
	cases := []struct {
		phase      connectionPhase
		title, sub string
	}{
		{phaseSearching, "장치 검색 중", "ESP32-C3 · 자동 검색 중"},
		{phaseOpening, "포트 연결 중", "COM11 · 포트 여는 중"},
		{phaseHandshake, "장치 확인 중", "COM11 · 응답 확인 중"},
		{phaseConnected, "PN12 연결됨", "COM11 · 12키 통신 중"},
		{phaseRetry, "재연결 대기", "COM11 · 자동 재시도"},
	}
	for _, tc := range cases {
		s := connectionProgress{Phase: tc.phase, Port: "COM11", CheckedAt: now}
		if got := s.title(); got != tc.title {
			t.Fatalf("%v title = %q", tc.phase, got)
		}
		if got := s.subtitle(); got != tc.sub {
			t.Fatalf("%v sub = %q", tc.phase, got)
		}
	}
}
func TestConnectionRetryReasonAndCheckTime(t *testing.T) {
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)
	s := connectionProgress{Phase: phaseRetry, Port: "COM11", Detail: "포트 사용 중", CheckedAt: now, RetryAt: now.Add(4 * time.Second)}
	if !strings.Contains(s.description(now), "12:00:04 재시도") {
		t.Fatal(s.description(now))
	}
	if !strings.Contains(s.description(now.Add(5*time.Second)), "포트 사용 중") {
		t.Fatal("missing error detail")
	}
	if !strings.Contains(s.checkedLabel(), "12:00:00") {
		t.Fatal(s.checkedLabel())
	}
}
