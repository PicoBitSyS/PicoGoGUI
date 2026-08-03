package dialog

import "errors"

// ErrUnsupportedPlatform is returned for native dialogs outside Windows.
var ErrUnsupportedPlatform = errors.New("picogogui/dialog: native file dialogs are only supported on Windows")

// FileFilter is one entry in a native file dialog filter list.
type FileFilter struct {
	Name    string
	Pattern string
}

// FileOptions configures native open/save dialogs.
type FileOptions struct {
	Title       string
	InitialDir  string
	DefaultName string
	DefaultExt  string
	Filters     []FileFilter
}

// OpenFile shows the native Windows open-file dialog.
func OpenFile(options FileOptions) (path string, accepted bool, err error) {
	return openFile(options, false)
}

// SaveFile shows the native Windows save-file dialog.
func SaveFile(options FileOptions) (path string, accepted bool, err error) {
	return openFile(options, true)
}

// SelectFolder shows the native Windows folder picker.
func SelectFolder(title string) (path string, accepted bool, err error) {
	return selectFolder(title)
}
