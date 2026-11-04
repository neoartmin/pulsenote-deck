//go:build windows

package main

import (
    "fmt"
    "runtime"
    "syscall"
    "unsafe"
)

type audioGUID struct {
    D1 uint32
    D2 uint16
    D3 uint16
    D4 [8]byte
}
var (
    clsidAudioEnumerator = audioGUID{0xBCDE0395,0xE52F,0x467C,[8]byte{0x8E,0x3D,0xC4,0x57,0x92,0x91,0x69,0x2E}}
    iidAudioEnumerator = audioGUID{0xA95664D2,0x9614,0x4F35,[8]byte{0xA7,0x46,0xDE,0x8D,0xB6,0x36,0x17,0xE6}}
    clsidAudioPolicy = audioGUID{0x870AF99C,0x171D,0x4F9E,[8]byte{0xAF,0x0D,0xE6,0x3D,0xF4,0x0C,0x2B,0xC9}}
    iidAudioPolicy = audioGUID{0xF8679F50,0x850A,0x41CF,[8]byte{0x9C,0x72,0x43,0x0F,0x29,0x02,0x90,0xC8}}
    pkeyFriendlyName = audioPropertyKey{audioGUID{0xA45C254E,0xDF1C,0x4EFD,[8]byte{0x80,0x20,0x67,0xD1,0x46,0xA8,0x50,0xE0}},14}
    coCreateAudio = ole32.NewProc("CoCreateInstance")
    coTaskFreeAudio = ole32.NewProc("CoTaskMemFree")
    clearAudioVariant = ole32.NewProc("PropVariantClear")
)

type audioPropertyKey struct{
    Fmtid audioGUID
    PID uint32
}
type audioVariant struct{
    VT uint16
    Reserved [6]byte
    Data [16]byte
}
type audioEndpoint struct {ID,Name string}

func audioHResult(hr uintptr)error{
    if int32(hr)<0{return fmt.Errorf("Windows Audio COM HRESULT 0x%08X",uint32(hr))}
    return nil
}
func audioCOMMethod(obj uintptr,slot uintptr)uintptr{
    vtable:=*(*uintptr)(unsafe.Pointer(obj))
    return *(*uintptr)(unsafe.Pointer(vtable+slot*unsafe.Sizeof(uintptr(0))))
}
func audioCall(obj uintptr,slot uintptr,args ...uintptr)(uintptr,error){
    if obj==0{return 0,fmt.Errorf("COM 객체가 비어 있습니다")}
    all:=make([]uintptr,0,len(args)+1)
    all=append(all,obj)
    all=append(all,args...)
    hr,_,_:=syscall.SyscallN(audioCOMMethod(obj,slot),all...)
    return hr,audioHResult(hr)
}
func audioRelease(obj uintptr){
    if obj!=0 {_,_,_=syscall.SyscallN(audioCOMMethod(obj,2),obj)}
}
func newAudioCOM(clsid,iid *audioGUID)(uintptr,error){
    var obj uintptr
    hr,_,_:=coCreateAudio.Call(uintptr(unsafe.Pointer(clsid)),0,1,uintptr(unsafe.Pointer(iid)),uintptr(unsafe.Pointer(&obj)))
    if err:=audioHResult(hr);err!=nil{return 0,err}
    return obj,nil
}
func withAudioCOM(fn func()error)error{
    runtime.LockOSThread()
    defer runtime.UnlockOSThread()
    hr,_,_:=coInit.Call(0,0) // COINIT_MULTITHREADED
    if uint32(hr)!=0x80010106 {
        if e:=audioHResult(hr);e!=nil{return e}
        defer coUninit.Call()
    }
    return fn()
}
func wideAudioString(ptr uintptr)string{
    if ptr==0{return ""}
    buf:=make([]uint16,0,128)
    for i:=0;i<4096;i++{
        ch:=*(*uint16)(unsafe.Pointer(ptr+uintptr(2*i)))
        if ch==0{break}
        buf=append(buf,ch)
    }
    return syscall.UTF16ToString(buf)
}
func audioDeviceID(device uintptr)(string,error){
    var idPtr uintptr
    _,err:=audioCall(device,5,uintptr(unsafe.Pointer(&idPtr)))
    if err!=nil{return "",err}
    defer coTaskFreeAudio.Call(idPtr)
    return wideAudioString(idPtr),nil
}
func audioDeviceName(device uintptr)string{
    var props uintptr
    _,err:=audioCall(device,4,0,uintptr(unsafe.Pointer(&props)))
    if err!=nil{return "출력 장치"}
    defer audioRelease(props)
    var data audioVariant
    _,err=audioCall(props,5,uintptr(unsafe.Pointer(&pkeyFriendlyName)),uintptr(unsafe.Pointer(&data)))
    if err!=nil{return "출력 장치"}
    defer clearAudioVariant.Call(uintptr(unsafe.Pointer(&data)))
    if data.VT!=31{return "출력 장치"} // VT_LPWSTR
    str:=wideAudioString(*(*uintptr)(unsafe.Pointer(&data.Data[0])))
    if str==""{return "출력 장치"}
    return str
}
func enumerateAudioDevices(enumerator uintptr)([]audioEndpoint,error){
    var collection uintptr
    _,err:=audioCall(enumerator,3,0,1,uintptr(unsafe.Pointer(&collection))) // eRender / DEVICE_STATE_ACTIVE
    if err!=nil{return nil,err}
    defer audioRelease(collection)
    var count uint32
    _,err=audioCall(collection,3,uintptr(unsafe.Pointer(&count)))
    if err!=nil{return nil,err}
    if count>128{return nil,fmt.Errorf("출력 장치 수가 비정상적입니다")}
    result:=make([]audioEndpoint,0,count)
    for i:=uint32(0);i<count;i++{
        var device uintptr
        _,err=audioCall(collection,4,uintptr(i),uintptr(unsafe.Pointer(&device)))
        if err!=nil{return nil,err}
        id,err:=audioDeviceID(device)
        if err!=nil{audioRelease(device);return nil,err}
        name:=audioDeviceName(device)
        audioRelease(device)
        result=append(result,audioEndpoint{ID:id,Name:name})
    }
    return result,nil
}
func currentAudioDevice(enumerator uintptr)(string,error){
    var device uintptr
    _,err:=audioCall(enumerator,4,0,0,uintptr(unsafe.Pointer(&device))) // eRender / eConsole
    if err!=nil{return "",err}
    defer audioRelease(device)
    return audioDeviceID(device)
}
func listAudioOutputs()([]audioEndpoint,string,error){
    var outputs []audioEndpoint
    var current string
    err:=withAudioCOM(func()error{
        enumerator,e:=newAudioCOM(&clsidAudioEnumerator,&iidAudioEnumerator)
        if e!=nil{return e}
        defer audioRelease(enumerator)
        outputs,e=enumerateAudioDevices(enumerator)
        if e!=nil{return e}
        current,e=currentAudioDevice(enumerator)
        return e
    })
    return outputs,current,err
}
func chooseAudioOutput(devices []audioEndpoint,current,value string)(audioEndpoint,error){
    if len(devices)==0{return audioEndpoint{},fmt.Errorf("사용 가능한 출력 장치가 없습니다")}
    if value!=""&&value!="cycle"{
        for _,dev:=range devices{if dev.ID==value{return dev,nil}}
        return audioEndpoint{},fmt.Errorf("지정한 출력 장치가 연결되지 않았습니다")
    }
    if len(devices)<2{return audioEndpoint{},fmt.Errorf("순환할 출력 장치가 두 개 이상 필요합니다")}
    for i,dev:=range devices{
        if dev.ID==current{return devices[(i+1)%len(devices)],nil}
    }
    return devices[0],nil
}
func switchAudioOutput(value string)error{
    var chosen audioEndpoint
    err:=withAudioCOM(func()error{
        enumerator,e:=newAudioCOM(&clsidAudioEnumerator,&iidAudioEnumerator)
        if e!=nil{return e}
        defer audioRelease(enumerator)
        devices,e:=enumerateAudioDevices(enumerator)
        if e!=nil{return e}
        current,e:=currentAudioDevice(enumerator)
        if e!=nil{return e}
        chosen,e=chooseAudioOutput(devices,current,value)
        if e!=nil{return e}
        if chosen.ID==current{return nil}
        policy,e:=newAudioCOM(&clsidAudioPolicy,&iidAudioPolicy)
        if e!=nil{return fmt.Errorf("Windows 출력 전환 인터페이스를 사용할 수 없습니다: %w",e)}
        defer audioRelease(policy)
        ptr,e:=syscall.UTF16PtrFromString(chosen.ID)
        if e!=nil{return e}
        // IPolicyConfig is not an officially documented Microsoft API.
        // Its SetDefaultEndpoint method is at IPolicyConfig vtable slot 13.
        for _,role:=range []uintptr{0,1}{ // Console and Multimedia
            _,e=audioCall(policy,13,uintptr(unsafe.Pointer(ptr)),role)
            if e!=nil{return fmt.Errorf("출력 장치 변경 실패 (role=%d): %w",role,e)}
        }
        runtime.KeepAlive(ptr)
        actual,e:=currentAudioDevice(enumerator)
        if e!=nil{return e}
        if actual!=chosen.ID{return fmt.Errorf("출력 장치 변경 확인 실패")}
        return nil
    })
    if err!=nil{
        enqueueActionStateNotice("스피커 전환 실패",err.Error(),true)
        return err
    }
    enqueueActionStateNotice("스피커 전환",chosen.Name,false)
    return nil
}
