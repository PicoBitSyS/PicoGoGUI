//go:build !windows

package dialog

func openFile(FileOptions, bool) (string, bool, error) {
	return "", false, ErrUnsupportedPlatform
}

func selectFolder(string) (string, bool, error) {
	return "", false, ErrUnsupportedPlatform
}
