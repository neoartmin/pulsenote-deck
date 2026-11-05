//go:build windows

package main

import "unsafe"

const WM_ACTION_NOTICE = WM_APP + 23

type actionStateNotice struct{
    title string
    message string
    warning bool
}
var actionStateEvents=make(chan actionStateNotice,32)

func enqueueActionStateNotice(title,message string,warning bool){
    n:=actionStateNotice{title:title,message:message,warning:warning}
    select{
    case actionStateEvents<-n:
    default:
        select{case <-actionStateEvents:default:}
        select{case actionStateEvents<-n:default:}
    }
    if mainUI!=nil&&mainUI.hwnd!=0{
        postMessage.Call(mainUI.hwnd,WM_ACTION_NOTICE,0,0)
    }
}

func (g *ui) showActionStateNotice(n actionStateNotice){
    if g==nil||g.hwnd==0||!notificationsEnabled(){return}
    var v nid
    v.Size=uint32(unsafe.Sizeof(v))
    v.Hwnd=g.hwnd
    v.ID=1
    v.Flags=NIF_INFO
    v.InfoFlags=NIIF_INFO
    if n.warning {v.InfoFlags=NIIF_WARNING}
    uint16copy(v.InfoTitle[:],"PulseNote · "+n.title)
    uint16copy(v.Info[:],n.message)
    notifyIcon.Call(NIM_MODIFY,uintptr(unsafe.Pointer(&v)))
}
