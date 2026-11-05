//go:build windows

package main

import "unsafe"

const WM_SHORTCUT_TOGGLE = WM_APP+18

type shortcutToggleNotice struct{
    label string
    active bool
}
var shortcutToggleEvents=make(chan shortcutToggleNotice,256)

// One event is emitted exactly once for each real state transition.
// The UI thread owns Shell_NotifyIcon, never the USB action worker.
func enqueueShortcutToggleNotice(label string,active bool){
    n:=shortcutToggleNotice{label:label,active:active}
    select {
    case shortcutToggleEvents<-n:
    default:
        // Keep newest transitions instead of blocking keyboard release.
        select{case <-shortcutToggleEvents:default:}
        shortcutToggleEvents<-n
    }
    if mainUI!=nil&&mainUI.hwnd!=0{
        postMessage.Call(mainUI.hwnd,WM_SHORTCUT_TOGGLE,0,0)
    }
}

func (g *ui) showShortcutToggleNotice(n shortcutToggleNotice){
    if g==nil||g.hwnd==0||!notificationsEnabled(){return}
    var v nid
    v.Size=uint32(unsafe.Sizeof(v))
    v.Hwnd=g.hwnd
    v.ID=1
    v.Flags=NIF_INFO
    v.InfoFlags=NIIF_INFO
    if n.active {
        uint16copy(v.InfoTitle[:],"PulseNote · 토글 활성화")
        uint16copy(v.Info[:],n.label+" · 키 누름 유지 시작")
    }else{
        uint16copy(v.InfoTitle[:],"PulseNote · 토글 비활성화")
        uint16copy(v.Info[:],n.label+" · 키 누름 유지 해제")
    }
    notifyIcon.Call(NIM_MODIFY,uintptr(unsafe.Pointer(&v)))
}
