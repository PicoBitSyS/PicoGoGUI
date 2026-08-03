//go:build windows

package win32

import (
	"golang.org/x/sys/windows"
)

// SetDPIAware enables per-monitor DPI awareness when available.
func SetDPIAware() {
	user := windows.NewLazySystemDLL("user32.dll")
	if proc := user.NewProc("SetProcessDpiAwarenessContext"); proc.Find() == nil {
		const dpiAwarenessContextPerMonitorAwareV2 = ^uintptr(3) + 1 // -4
		_, _, _ = proc.Call(dpiAwarenessContextPerMonitorAwareV2)
		return
	}
	shcore := windows.NewLazySystemDLL("shcore.dll")
	if proc := shcore.NewProc("SetProcessDpiAwareness"); proc.Find() == nil {
		const processPerMonitorDPIAware = 2
		_, _, _ = proc.Call(processPerMonitorDPIAware)
		return
	}
	if proc := user.NewProc("SetProcessDPIAware"); proc.Find() == nil {
		_, _, _ = proc.Call()
	}
}
