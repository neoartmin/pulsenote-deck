//go:build windows

package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var cmLocateDevNode = syscall.NewLazyDLL("cfgmgr32.dll").NewProc("CM_Locate_DevNodeW")

// Confirm a USB device instance is currently present, not merely a stale
// device-manager entry from a previously connected ESP32-C3.
func presentDeviceNode(instanceID string) bool {
	var instance uint32
	id := syscall.StringToUTF16Ptr(instanceID)
	code, _, _ := cmLocateDevNode.Call(uintptr(unsafe.Pointer(&instance)), uintptr(unsafe.Pointer(id)), 0)
	return code == 0
}

// Windows Device Manager's active USB hardware ID is required before we ever
// issue an automatic normal-boot reset. Never reset arbitrary COM equipment.
// ESP32-C3's integrated USB Serial/JTAG uses VID_303A&PID_1001.
func isESP32NativeUSBPort(port string) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	for _, path := range []string{
		`SYSTEM\CurrentControlSet\Enum\USB\VID_303A&PID_1001`,
		`SYSTEM\CurrentControlSet\Enum\USB\VID_303A&PID_1001&MI_00`,
	} {
		if espPortAtRegistryPath(path, port) {
			return true
		}
	}
	return false
}

func espPortAtRegistryPath(path, port string) bool {
	var parent syscall.Handle
	if err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, syscall.StringToUTF16Ptr(path), 0, syscall.KEY_READ, &parent); err != nil {
		return false
	}
	defer syscall.RegCloseKey(parent)
	for i := uint32(0); i < 64; i++ {
		buf := make([]uint16, 512)
		size := uint32(len(buf))
		if err := syscall.RegEnumKeyEx(parent, i, &buf[0], &size, nil, nil, nil, nil); err != nil {
			break
		}
		device := syscall.UTF16ToString(buf[:size])
		// The registry retains disconnected devices. Only accept a live node.
		if !presentDeviceNode(`USB\` + strings.TrimPrefix(path, `SYSTEM\CurrentControlSet\Enum\USB\`) + `\` + device) {
			continue
		}
		var key syscall.Handle
		if err := syscall.RegOpenKeyEx(parent, syscall.StringToUTF16Ptr(device+`\Device Parameters`), 0, syscall.KEY_READ, &key); err != nil {
			continue
		}
		var valType uint32
		n := uint32(256)
		data := make([]byte, n)
		err := syscall.RegQueryValueEx(key, syscall.StringToUTF16Ptr("PortName"), nil, &valType, &data[0], &n)
		syscall.RegCloseKey(key)
		if err != nil || valType != syscall.REG_SZ || n < 2 {
			continue
		}
		utf := make([]uint16, n/2)
		for j := range utf {
			utf[j] = uint16(data[j*2]) | uint16(data[j*2+1])<<8
		}
		if strings.EqualFold(strings.TrimSpace(syscall.UTF16ToString(utf)), port) {
			return true
		}
	}
	return false
}

// Reproduce Espressif's "reset into booting from flash" USB Serial/JTAG
// control-line sequence: DTR off, RTS off, pulse RTS on, RTS off.
// This is strictly a one-shot recovery after a failed firmware handshake on
// a confirmed native ESP32-C3 port, never on a working connection.
func requestNativeNormalBoot(f *os.File) error {
	handle := uintptr(f.Fd())
	for _, step := range []struct {
		action uintptr
		pause  time.Duration
	}{
		{6, 20 * time.Millisecond},  // CLRDTR
		{4, 20 * time.Millisecond},  // CLRRTS
		{3, 120 * time.Millisecond}, // SETRTS: request chip reset
		{4, 240 * time.Millisecond}, // CLRRTS: release reset
	} {
		ok, _, e := escapeCommFunction.Call(handle, step.action)
		if ok == 0 {
			return fmt.Errorf("EscapeCommFunction(%d): %v", step.action, e)
		}
		time.Sleep(step.pause)
	}
	return nil
}
