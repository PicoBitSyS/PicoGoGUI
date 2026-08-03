//go:build windows

package clipboard

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")
	openClipboard    = user32.NewProc("OpenClipboard")
	closeClipboard   = user32.NewProc("CloseClipboard")
	emptyClipboard   = user32.NewProc("EmptyClipboard")
	getClipboardData = user32.NewProc("GetClipboardData")
	setClipboardData = user32.NewProc("SetClipboardData")
	globalAlloc      = kernel32.NewProc("GlobalAlloc")
	globalLock       = kernel32.NewProc("GlobalLock")
	globalUnlock     = kernel32.NewProc("GlobalUnlock")
	globalFree       = kernel32.NewProc("GlobalFree")
	copyMemory       = kernel32.NewProc("RtlMoveMemory")
	lstrlenW         = kernel32.NewProc("lstrlenW")
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

func openCB() error {
	runtime.LockOSThread()
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		r, _, err := openClipboard.Call(0)
		if r != 0 {
			return nil
		}
		if time.Now().After(deadline) {
			runtime.UnlockOSThread()
			return fmt.Errorf("OpenClipboard: %w", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func getText() (string, error) {
	if err := openCB(); err != nil {
		return "", err
	}
	defer closeClipboard.Call()
	defer runtime.UnlockOSThread()

	h, _, err := getClipboardData.Call(cfUnicodeText)
	if h == 0 {
		return "", fmt.Errorf("GetClipboardData: %w", err)
	}
	ptr, _, err := globalLock.Call(h)
	if ptr == 0 {
		return "", fmt.Errorf("GlobalLock: %w", err)
	}
	defer globalUnlock.Call(h)
	length, _, _ := lstrlenW.Call(ptr)
	if length == 0 {
		return "", nil
	}
	text := make([]uint16, int(length)+1)
	copyMemory.Call(uintptr(unsafe.Pointer(&text[0])), ptr, length*2)
	return syscall.UTF16ToString(text), nil
}

func setText(s string) error {
	utf16, err := syscall.UTF16FromString(s)
	if err != nil {
		return err
	}
	bytes := len(utf16) * 2

	if err := openCB(); err != nil {
		return err
	}
	defer closeClipboard.Call()
	defer runtime.UnlockOSThread()

	if r, _, callErr := emptyClipboard.Call(); r == 0 {
		return fmt.Errorf("EmptyClipboard: %w", callErr)
	}
	h, _, err := globalAlloc.Call(gmemMoveable, uintptr(bytes))
	if h == 0 {
		return fmt.Errorf("GlobalAlloc: %w", err)
	}
	ptr, _, err := globalLock.Call(h)
	if ptr == 0 {
		globalFree.Call(h)
		return fmt.Errorf("GlobalLock: %w", err)
	}
	copyMemory.Call(ptr, uintptr(unsafe.Pointer(&utf16[0])), uintptr(bytes))
	globalUnlock.Call(h)

	r, _, err := setClipboardData.Call(cfUnicodeText, h)
	if r == 0 {
		globalFree.Call(h)
		return fmt.Errorf("SetClipboardData: %w", err)
	}
	return nil
}
