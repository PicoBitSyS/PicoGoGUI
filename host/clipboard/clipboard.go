// Package clipboard provides text clipboard access.
package clipboard

import "errors"

// ErrUnsupportedPlatform is returned on non-Windows builds.
var ErrUnsupportedPlatform = errors.New("picogogui/clipboard: clipboard is only supported on Windows")

// GetText returns the current clipboard text.
func GetText() (string, error) { return getText() }

// SetText sets the clipboard text.
func SetText(s string) error { return setText(s) }
