//go:build windows

package main

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

// The documented MMDevice API discovers the current and available playback
// endpoints. Windows does not expose a documented public setter for the
// system's default endpoint. The guarded IPolicyConfig call below is an
// undocumented Windows interface: return a visible error if unavailable.
var (
	coCreateInstance=ole32.NewProc("CoCreateInstance")
	propVariantClear=ole32.NewProc("PropVariantClear")
	clsidMMDeviceEnumerator=syscall.GUID{Data1:0xBCDE0395,Data2:0xE52F,Data3:0x467C,Data4:[8]byte{0x8E,0x3D,0xC4,0x57,0x92,0x91,0x69,0x2E}}
	iidIMMDeviceEnumerator=syscall.GUID{Data1:0xA95664D2,Data2:0x9614,Data3:0x4F35,Data4:[8]byte{0xA7,0x46,0xDE,0x8D,0xB6,0x36,0x17,0xE6}}
	clsidPolicyConfig=syscall.GUID{Data1:0x870AF99C,Data2:0x171D,Data3:0x4F9E,Data4:[8]byte{0xAF,0x0D,0xE6,0x3D,0xF4,0x0C,0x2B,0xC9}}
	iidPolicyConfig=syscall.GUID{Data1:0xF8679F50,Data2:0x850A,Data3:0x41CF,Data4:[8]byte{0x9C,0x72,0x43,0x0F,0x29,0x02,0x90,0xC8}}
	pkeyFriendlyName=struct{Fmtid syscall.GUID;Pid uint32}{
		Fmtid:syscall.GUID{Data1:0xA45C254E,Data2:0xDF1C,Data3:0x4EFD,Data4:[8]byte{0x80,0x20,0x67,0xD1,0x46,0xA8,0x50,0xE0}},
		Pid:14,
	}
)

type audioPropVariant struct {
	Vt uint16
	_ [6]byte
	Value uintptr
	_2 uintptr
}
type playbackEndpoint struct{ id,name string }

func failedHRESULT(hr uintptr) bool{return uint32(hr)&0x80000000!=0}
func audioHRESULT(what string,hr uintptr) error {return fmt.Errorf("%s 실패 (HRESULT 0x%08X)",what,uint32(hr))}
func comMethod(obj uintptr,index int,params ...uintptr) uintptr {
	table:=*(*uintptr)(unsafe.Pointer(obj))
	method:=*(*uintptr)(unsafe.Pointer(table+uintptr(index)*unsafe.Sizeof(uintptr(0))))
	args:=append([]uintptr{obj},params...)
	r,_,_:=syscall.SyscallN(method,args...)
	return r
}
func createCOM(clsid,iid *syscall.GUID)(uintptr,error) {
	var instance uintptr
	hr,_,_:=coCreateInstance.Call(uintptr(unsafe.Pointer(clsid)),0,1,uintptr(unsafe.Pointer(iid)),uintptr(unsafe.Pointer(&instance)))
	if failedHRESULT(hr)||instance==0{return 0,audioHRESULT("COM 인터페이스 초기화",hr)}
	return instance,nil
}
func releaseCOM(obj uintptr){if obj!=0{comMethod(obj,2)}}
func audioEndpointID(device uintptr)(string,error){
	var p uintptr
	hr:=comMethod(device,5,uintptr(unsafe.Pointer(&p)))
	if failedHRESULT(hr)||p==0{return "",audioHRESULT("오디오 장치 ID",hr)}
	defer ole32.NewProc("CoTaskMemFree").Call(p)
	return syscall.UTF16PtrToString((*uint16)(unsafe.Pointer(p))),nil
}
func friendlyAudioName(device uintptr) string {
	var store uintptr
	hr:=comMethod(device,4,0,uintptr(unsafe.Pointer(&store))) // STGM_READ
	if failedHRESULT(hr)||store==0{return ""}
	defer releaseCOM(store)
	var v audioPropVariant
	hr=comMethod(store,5,uintptr(unsafe.Pointer(&pkeyFriendlyName)),uintptr(unsafe.Pointer(&v)))
	if failedHRESULT(hr){return ""}
	defer propVariantClear.Call(uintptr(unsafe.Pointer(&v)))
	if v.Vt!=31||v.Value==0{return ""} // VT_LPWSTR
	return syscall.UTF16PtrToString((*uint16)(unsafe.Pointer(v.Value)))
}
func enumerateOutputDevices()(endpoints []playbackEndpoint,current string,err error){
	en,err:=createCOM(&clsidMMDeviceEnumerator,&iidIMMDeviceEnumerator)
	if err!=nil{return nil,"",err}
	defer releaseCOM(en)
	var defaultDevice uintptr
	hr:=comMethod(en,4,0,0,uintptr(unsafe.Pointer(&defaultDevice))) // eRender, eConsole
	if !failedHRESULT(hr)&&defaultDevice!=0{
		current,_=audioEndpointID(defaultDevice)
		releaseCOM(defaultDevice)
	}
	var collection uintptr
	hr=comMethod(en,3,0,1,uintptr(unsafe.Pointer(&collection))) // eRender, ACTIVE
	if failedHRESULT(hr)||collection==0{return nil,current,audioHRESULT("재생 장치 목록",hr)}
	defer releaseCOM(collection)
	var count uint32
	hr=comMethod(collection,3,uintptr(unsafe.Pointer(&count)))
	if failedHRESULT(hr){return nil,current,audioHRESULT("재생 장치 개수",hr)}
	for index:=uint32(0);index<count;index++{
		var device uintptr
		hr=comMethod(collection,4,uintptr(index),uintptr(unsafe.Pointer(&device)))
		if failedHRESULT(hr)||device==0{continue}
		id,idErr:=audioEndpointID(device)
		name:=friendlyAudioName(device)
		releaseCOM(device)
		if idErr!=nil||id==""{continue}
		if name==""{name=fmt.Sprintf("출력 장치 %d",index+1)}
		endpoints=append(endpoints,playbackEndpoint{id:id,name:name})
	}
	return endpoints,current,nil
}
func setDefaultOutput(id string)error {
	policy,err:=createCOM(&clsidPolicyConfig,&iidPolicyConfig)
	if err!=nil{return fmt.Errorf("Windows 오디오 출력 변경 API를 사용할 수 없습니다: %w",err)}
	defer releaseCOM(policy)
	utf,err:=syscall.UTF16PtrFromString(id)
	if err!=nil{return err}
	for _,role:=range []uintptr{0,1}{ // Console and Multimedia, not Communications
		// IMPORTANT: slot 13 belongs ONLY to IPolicyConfig (IID F8679F50).
		hr:=comMethod(policy,13,uintptr(unsafe.Pointer(utf)),role)
		if failedHRESULT(hr){return audioHRESULT("기본 출력 장치 변경",hr)}
	}
	return nil
}
func switchAudioOutput()error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	hr,_,_:=coInit.Call(0,0) // COINIT_MULTITHREADED on this worker thread
	if failedHRESULT(hr){return audioHRESULT("Windows 오디오 초기화",hr)}
	defer coUninit.Call()
	endpoints,current,err:=enumerateOutputDevices()
	if err!=nil{return err}
	index:=-1
	for i,p:=range endpoints{if p.id==current{index=i;break}}
	next:=nextOutputIndex(len(endpoints),index)
	if next<0{return fmt.Errorf("활성 출력 장치가 2개 이상 있어야 전환할 수 있습니다")}
	if err:=setDefaultOutput(endpoints[next].id);err!=nil{return err}
	enqueueActionNotice("스피커 전환 완료",endpoints[next].name)
	return nil
}
