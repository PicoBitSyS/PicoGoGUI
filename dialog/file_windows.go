//go:build windows

package dialog

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	comdlg32             = windows.NewLazySystemDLL("comdlg32.dll")
	shell32              = windows.NewLazySystemDLL("shell32.dll")
	ole32                = windows.NewLazySystemDLL("ole32.dll")
	getOpenFileNameW     = comdlg32.NewProc("GetOpenFileNameW")
	getSaveFileNameW     = comdlg32.NewProc("GetSaveFileNameW")
	commDlgExtendedError = comdlg32.NewProc("CommDlgExtendedError")
	shBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	shGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
	coTaskMemFree        = ole32.NewProc("CoTaskMemFree")
	coInitializeEx       = ole32.NewProc("CoInitializeEx")
	coUninitialize       = ole32.NewProc("CoUninitialize")
)

const (
	ofnExplorer             = 0x00080000
	ofnFileMustExist        = 0x00001000
	ofnPathMustExist        = 0x00000800
	ofnOverwritePrompt      = 0x00000002
	ofnNoChangeDir          = 0x00000008
	bifReturnOnlyFSDirs     = 0x0001
	bifNewDialogStyle       = 0x0040
	coinitApartmentThreaded = 0x2
)

type openFileName struct {
	structSize      uint32
	owner           uintptr
	instance        uintptr
	filter          *uint16
	customFilter    *uint16
	maxCustomFilter uint32
	filterIndex     uint32
	file            *uint16
	maxFile         uint32
	fileTitle       *uint16
	maxFileTitle    uint32
	initialDir      *uint16
	title           *uint16
	flags           uint32
	fileOffset      uint16
	fileExtension   uint16
	defaultExt      *uint16
	customData      uintptr
	hook            uintptr
	templateName    *uint16
	reserved        unsafe.Pointer
	reservedSize    uint32
	flagsEx         uint32
}

type browseInfo struct {
	owner       uintptr
	root        uintptr
	displayName *uint16
	title       *uint16
	flags       uint32
	callback    uintptr
	param       uintptr
	image       int32
}

func openFile(options FileOptions, save bool) (string, bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	buffer := make([]uint16, 32768)
	if options.DefaultName != "" {
		name, err := syscall.UTF16FromString(options.DefaultName)
		if err != nil {
			return "", false, err
		}
		copy(buffer, name)
	}
	filter := buildFilter(options.Filters)
	title, err := optionalUTF16(options.Title)
	if err != nil {
		return "", false, err
	}
	initialDir, err := optionalUTF16(options.InitialDir)
	if err != nil {
		return "", false, err
	}
	defaultExt, err := optionalUTF16(strings.TrimPrefix(options.DefaultExt, "."))
	if err != nil {
		return "", false, err
	}
	flags := uint32(ofnExplorer | ofnPathMustExist | ofnNoChangeDir)
	proc := getOpenFileNameW
	if save {
		flags |= ofnOverwritePrompt
		proc = getSaveFileNameW
	} else {
		flags |= ofnFileMustExist
	}
	request := openFileName{
		file:        &buffer[0],
		maxFile:     uint32(len(buffer)),
		filterIndex: 1,
		flags:       flags,
	}
	request.structSize = uint32(unsafe.Sizeof(request))
	if len(filter) > 0 {
		request.filter = &filter[0]
	}
	if len(title) > 0 {
		request.title = &title[0]
	}
	if len(initialDir) > 0 {
		request.initialDir = &initialDir[0]
	}
	if len(defaultExt) > 0 {
		request.defaultExt = &defaultExt[0]
	}
	ok, _, callErr := proc.Call(uintptr(unsafe.Pointer(&request)))
	if ok != 0 {
		return syscall.UTF16ToString(buffer), true, nil
	}
	code, _, _ := commDlgExtendedError.Call()
	if code == 0 {
		return "", false, nil
	}
	return "", false, fmt.Errorf("native file dialog failed (0x%x): %w", code, callErr)
}

func selectFolder(titleText string) (string, bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	initialized, _, _ := coInitializeEx.Call(0, coinitApartmentThreaded)
	if initialized == 0 || initialized == 1 {
		defer coUninitialize.Call()
	}
	title, err := optionalUTF16(titleText)
	if err != nil {
		return "", false, err
	}
	display := make([]uint16, windows.MAX_PATH)
	info := browseInfo{
		displayName: &display[0],
		flags:       bifReturnOnlyFSDirs | bifNewDialogStyle,
	}
	if len(title) > 0 {
		info.title = &title[0]
	}
	itemID, _, callErr := shBrowseForFolderW.Call(uintptr(unsafe.Pointer(&info)))
	if itemID == 0 {
		return "", false, nil
	}
	defer coTaskMemFree.Call(itemID)
	path := make([]uint16, windows.MAX_PATH)
	ok, _, pathErr := shGetPathFromIDListW.Call(itemID, uintptr(unsafe.Pointer(&path[0])))
	if ok == 0 {
		return "", false, fmt.Errorf("native folder dialog failed: %w / %v", callErr, pathErr)
	}
	return syscall.UTF16ToString(path), true, nil
}

func optionalUTF16(value string) ([]uint16, error) {
	if value == "" {
		return nil, nil
	}
	return syscall.UTF16FromString(value)
}

func buildFilter(filters []FileFilter) []uint16 {
	if len(filters) == 0 {
		filters = []FileFilter{{Name: "All files", Pattern: "*.*"}}
	}
	var value []rune
	for _, filter := range filters {
		name := filter.Name
		if name == "" {
			name = filter.Pattern
		}
		pattern := filter.Pattern
		if pattern == "" {
			pattern = "*.*"
		}
		value = append(value, []rune(name)...)
		value = append(value, 0)
		value = append(value, []rune(pattern)...)
		value = append(value, 0)
	}
	value = append(value, 0)
	return utf16.Encode(value)
}
