//go:build windows

package main

import (
	"time"
	"unsafe"
)

const (
	WM_DEVICE_NOTICE = WM_APP + 14
	NIF_INFO         = 0x10
	NIIF_INFO        = 0x1
	NIIF_WARNING     = 0x2
)

var liveDeviceNotices deviceNoticeTracker
var balloonLimiter noticeRateLimiter
var noticeEvents = make(chan deviceNotice, 16)

func enqueueDeviceNotice(n deviceNotice) {
	select {
	case noticeEvents <- n:
	default:
		// Prefer a recent transition over an old balloon if the UI is busy.
		select {
		case <-noticeEvents:
		default:
		}
		select {
		case noticeEvents <- n:
		default:
		}
	}
	if mainUI != nil && mainUI.hwnd != 0 {
		postMessage.Call(mainUI.hwnd, WM_DEVICE_NOTICE, 0, 0)
	}
}
func recordDeviceAttached(port string, keys int) {
	if n, ok := liveDeviceNotices.attach(port, keys); ok {
		enqueueDeviceNotice(n)
	}
}
func recordDeviceDetached(manual bool) {
	if n, ok := liveDeviceNotices.detach(manual); ok {
		enqueueDeviceNotice(n)
	}
}
func (g *ui) trayBalloon(n deviceNotice) {
	if g == nil || g.hwnd == 0 || !notificationsEnabled() || !balloonLimiter.allow(n, time.Now()) {
		return
	}
	var v nid
	v.Size = uint32(unsafe.Sizeof(v))
	v.Hwnd = g.hwnd
	v.ID = 1
	v.Flags = NIF_INFO
	if n.Connected {
		uint16copy(v.InfoTitle[:], "PulseNote Deck · 장치 연결")
		uint16copy(v.Info[:], deviceModelLabel(n.Keys)+" 장치가 "+n.Port+"에 연결되었습니다.")
		v.InfoFlags = NIIF_INFO
	} else {
		uint16copy(v.InfoTitle[:], "PulseNote Deck · 장치 통신 종료")
		uint16copy(v.Info[:], deviceModelLabel(n.Keys)+" 장치의 "+n.Port+" 통신이 중단됐습니다. 자동 재연결을 시도합니다.")
		v.InfoFlags = NIIF_WARNING
	}
	notifyIcon.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&v)))
}
func deviceModelLabel(keys int) string {
	switch keys {
	case 12:
		return "PN12"
	case 15:
		return "PN15"
	default:
		return "PulseNote Deck"
	}
}
