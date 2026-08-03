//go:build windows

package win32

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const wsOverlappedWindow = 0x00CF0000

type nativeRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// WindowSize is an outer native window size expressed in logical Windows
// pixels (DIP), plus the DPI used for the conversion.
type WindowSize struct {
	Width  int
	Height int
	DPI    int
}

var (
	windowUser32                 = windows.NewLazySystemDLL("user32.dll")
	adjustWindowRectExForDPIProc = windowUser32.NewProc("AdjustWindowRectExForDpi")
	getDPIForSystemProc          = windowUser32.NewProc("GetDpiForSystem")
	getDPIForWindowProc          = windowUser32.NewProc("GetDpiForWindow")
	getWindowRectProc            = windowUser32.NewProc("GetWindowRect")
)

// OuterWindowSize returns the real outer rectangle converted from physical
// pixels to logical Windows pixels.
func OuterWindowSize(hwnd uintptr) (WindowSize, error) {
	if hwnd == 0 {
		return WindowSize{}, fmt.Errorf("picogogui: native window is unavailable")
	}
	rect := nativeRect{}
	ok, _, callErr := getWindowRectProc.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	if ok == 0 {
		return WindowSize{}, fmt.Errorf("picogogui: GetWindowRect: %w", callErr)
	}
	dpi := windowDPI(hwnd)
	return WindowSize{
		Width:  scaleFromPhysical(int(rect.Right-rect.Left), dpi),
		Height: scaleFromPhysical(int(rect.Bottom-rect.Top), dpi),
		DPI:    dpi,
	}, nil
}

// ClientSizeForOuter converts an outer logical size to the physical client
// size expected by go-webview2.
func ClientSizeForOuter(hwnd uintptr, width, height int) (int, int, error) {
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("picogogui: outer window size must be positive")
	}
	dpi := windowDPI(hwnd)
	rect := nativeRect{}
	ok, _, callErr := adjustWindowRectExForDPIProc.Call(
		uintptr(unsafe.Pointer(&rect)),
		wsOverlappedWindow,
		0,
		0,
		uintptr(dpi),
	)
	if ok == 0 {
		return 0, 0, fmt.Errorf("picogogui: AdjustWindowRectExForDpi: %w", callErr)
	}
	physicalWidth := scaleToPhysical(width, dpi)
	physicalHeight := scaleToPhysical(height, dpi)
	clientWidth := physicalWidth - int(rect.Right-rect.Left)
	clientHeight := physicalHeight - int(rect.Bottom-rect.Top)
	if clientWidth <= 0 || clientHeight <= 0 {
		return 0, 0, fmt.Errorf("picogogui: outer window size is too small")
	}
	return clientWidth, clientHeight, nil
}

func windowDPI(hwnd uintptr) int {
	if hwnd != 0 {
		if dpi, _, _ := getDPIForWindowProc.Call(hwnd); dpi != 0 {
			return int(dpi)
		}
	}
	if dpi, _, _ := getDPIForSystemProc.Call(); dpi != 0 {
		return int(dpi)
	}
	return 96
}

func scaleToPhysical(value, dpi int) int {
	return (value*dpi + 48) / 96
}

func scaleFromPhysical(value, dpi int) int {
	return (value*96 + dpi/2) / dpi
}
