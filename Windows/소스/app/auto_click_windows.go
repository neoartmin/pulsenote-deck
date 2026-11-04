//go:build windows

package main

import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

const (
	mouseLeftDown=0x0002
	mouseLeftUp=0x0004
	mouseRightDown=0x0008
	mouseRightUp=0x0010
)

type nativeMouseInput struct {
	Dx,Dy int32
	Data,Flags,Timestamp uint32
	ExtraInfo uintptr
}
type nativeInput struct {
	Kind uint32
	_ uint32
	Mouse nativeMouseInput
}
var sendNativeInput=u.NewProc("SendInput")

// The cancellation channel is scoped to one active run. Never dispatch clicks
// from the USB worker, because that would block ordinary macro key execution.
type clickSession struct {
	cancel chan struct{}
	settings autoClickSettings
}
var autoClickState=struct{
	sync.Mutex
	active *clickSession
}{}

func clickOnce(button string) error {
	var down,up uint32=mouseLeftDown,mouseLeftUp
	if button=="right"{down,up=mouseRightDown,mouseRightUp}
	inputs:=[2]nativeInput{{Kind:0,Mouse:nativeMouseInput{Flags:down}},{Kind:0,Mouse:nativeMouseInput{Flags:up}}}
	n,_,e:=sendNativeInput.Call(2,uintptr(unsafe.Pointer(&inputs[0])),unsafe.Sizeof(inputs[0]))
	if n!=2 {
		// Attempt a button-up even after partial injection.
		release:=nativeInput{Kind:0,Mouse:nativeMouseInput{Flags:up}}
		sendNativeInput.Call(1,uintptr(unsafe.Pointer(&release)),unsafe.Sizeof(release))
		return fmt.Errorf("Windows 마우스 입력 실패: %v",e)
	}
	return nil
}

func stopAutoClick() bool {
	autoClickState.Lock()
	old:=autoClickState.active
	if old!=nil {autoClickState.active=nil;close(old.cancel)}
	autoClickState.Unlock()
	if old!=nil {enqueueActionNotice("오토클릭 중지","반복 클릭이 중지되었습니다")}
	return old!=nil
}

func toggleAutoClick(raw string) error {
	// A second press always stops, even when the new settings are invalid.
	if stopAutoClick(){return nil}
	settings,err:=parseAutoClickSettings(raw)
	if err!=nil{return err}
	run:=&clickSession{cancel:make(chan struct{}),settings:settings}
	autoClickState.Lock()
	autoClickState.active=run
	autoClickState.Unlock()
	enqueueActionNotice("오토클릭 시작",fmt.Sprintf("%s 클릭 · %dms 간격",map[bool]string{true:"오른쪽",false:"왼쪽"}[settings.Button=="right"],settings.IntervalMS))
	go func(){
		var done int
		var fail error
		defer func(){
			autoClickState.Lock()
			owned:=autoClickState.active==run
			if owned{autoClickState.active=nil}
			autoClickState.Unlock()
			if owned{
				if fail!=nil {enqueueActionNotice("오토클릭 오류",fail.Error())}else{enqueueActionNotice("오토클릭 완료",fmt.Sprintf("%d회 클릭 완료",done))}
			}
		}()
		clock:=time.NewTicker(time.Duration(settings.IntervalMS)*time.Millisecond)
		defer clock.Stop()
		for {
			select {case <-run.cancel:return;default:}
			if err:=clickOnce(settings.Button);err!=nil{fail=err;return}
			done++
			if settings.Limit>0&&done>=settings.Limit{return}
			select{case <-run.cancel:return;case <-clock.C:}
		}
	}()
	return nil
}
