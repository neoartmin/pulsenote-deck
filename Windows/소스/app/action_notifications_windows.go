//go:build windows

package main

import "unsafe"

const WM_ACTION_NOTICE = WM_APP + 21

type actionNotice struct{title,message string}
var actionNotices=make(chan actionNotice,64)

// Queue only significant user-visible transitions, never each repeated click.
func enqueueActionNotice(title,message string){
	if !notificationsEnabled(){return}
	n:=actionNotice{title:title,message:message}
	select{case actionNotices<-n:default:
		select{case <-actionNotices:default:}
		select{case actionNotices<-n:default:}
	}
	if mainUI!=nil&&mainUI.hwnd!=0 {
		postMessage.Call(mainUI.hwnd,WM_ACTION_NOTICE,0,0)
	}
}
func (g *ui) showActionNotice(n actionNotice){
	if g==nil||g.hwnd==0{return}
	var v nid
	v.Size=uint32(unsafe.Sizeof(v))
	v.Hwnd=g.hwnd
	v.ID=1
	v.Flags=NIF_INFO
	v.InfoFlags=NIIF_INFO
	uint16copy(v.InfoTitle[:],"PulseNote · "+n.title)
	uint16copy(v.Info[:],n.message)
	notifyIcon.Call(NIM_MODIFY,uintptr(unsafe.Pointer(&v)))
}
