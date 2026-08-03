// Package win32 provides small Windows helpers used by the native host.
package win32

import "errors"

// ErrWebView2Missing is returned when the WebView2 Runtime cannot be loaded.
var ErrWebView2Missing = errors.New("picogogui: WebView2 Runtime is not available; install the Evergreen WebView2 Runtime from Microsoft")
