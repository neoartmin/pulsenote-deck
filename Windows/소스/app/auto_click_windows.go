//go:build windows

package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
    "unsafe"
)

const (
    mouseInputType = 0
    leftMouseDown = 0x0002
    leftMouseUp = 0x0004
    rightMouseDown = 0x0008
    rightMouseUp = 0x0010
)

type mouseInput struct {
    DX int32
    DY int32
    Data uint32
    Flags uint32
    Time uint32
    ExtraInfo uintptr
}
type windowsInput struct {
    Type uint32
    _ uint32 // Windows x64: MOUSEINPUT is aligned to 8 bytes
    Mouse mouseInput
}

var sendMouseInputs = u.NewProc("SendInput")
var asyncKeyState = u.NewProc("GetAsyncKeyState")
var autoClickControl = struct{
    sync.Mutex
    done chan struct{}
    label string
}{}

func clickMouse(button string) error {
    down,up:=uint32(leftMouseDown),uint32(leftMouseUp)
    if button=="right" {down,up=rightMouseDown,rightMouseUp}
    inputs:=[2]windowsInput{
        {Type:mouseInputType,Mouse:mouseInput{Flags:down}},
        {Type:mouseInputType,Mouse:mouseInput{Flags:up}},
    }
    if unsafe.Sizeof(windowsInput{})!=40 {
        return fmt.Errorf("Windows 입력 구조 크기가 예상과 다릅니다")
    }
    sent,_,e:=sendMouseInputs.Call(2,uintptr(unsafe.Pointer(&inputs[0])),uintptr(unsafe.Sizeof(inputs[0])))
    runtime.KeepAlive(inputs)
    if sent!=2 {return fmt.Errorf("Windows SendInput 실패: %v",e)}
    return nil
}

func toggleAutoClick(raw,label string) error{
    settings,e:=parseAutoClick(raw)
    if e!=nil{return e}
    autoClickControl.Lock()
    if autoClickControl.done!=nil {
        close(autoClickControl.done)
        autoClickControl.done=nil
        previous:=autoClickControl.label
        autoClickControl.Unlock()
        enqueueActionStateNotice("오토클릭 중지",previous,false)
        return nil
    }
    done:=make(chan struct{})
    autoClickControl.done=done
    autoClickControl.label=label
    autoClickControl.Unlock()
    countLabel:="무제한"
    if settings.Count>0 {countLabel=fmt.Sprintf("%d회",settings.Count)}
    enqueueActionStateNotice("오토클릭 시작",fmt.Sprintf("%s · %dms · %s",label,settings.IntervalMS,countLabel),false)
    go runAutoClick(settings,done,label)
    return nil
}

func stopAutoClick(reason string){
    autoClickControl.Lock()
    if autoClickControl.done==nil{
        autoClickControl.Unlock()
        return
    }
    close(autoClickControl.done)
    autoClickControl.done=nil
    autoClickControl.Unlock()
    enqueueActionStateNotice("오토클릭 중지",reason,false)
}

func finishAutoClick(done chan struct{},label string,err error){
    autoClickControl.Lock()
    if autoClickControl.done!=done {
        autoClickControl.Unlock()
        return
    }
    autoClickControl.done=nil
    autoClickControl.Unlock()
    if err!=nil{
        enqueueActionStateNotice("오토클릭 오류",err.Error(),true)
    }else{
        enqueueActionStateNotice("오토클릭 완료",label,false)
    }
}

func runAutoClick(settings autoClickSpec,done chan struct{},label string){
    ticker:=time.NewTicker(time.Duration(settings.IntervalMS)*time.Millisecond)
    defer ticker.Stop()
    for n:=0;settings.Count==0||n<settings.Count;n++ {
        select {
        case <-done:return
        case <-ticker.C:
        }
        // F8 is a nonblocking emergency stop.
        held,_,_:=asyncKeyState.Call(0x77)
        if (uint16(held)&0x8000)!=0 {
            stopAutoClick("F8 긴급 중지")
            return
        }
        select{case <-done:return;default:}
        if err:=clickMouse(settings.Button);err!=nil{
            finishAutoClick(done,label,err)
            return
        }
    }
    finishAutoClick(done,label,nil)
}
