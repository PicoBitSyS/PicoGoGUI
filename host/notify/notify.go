// Package notify shows simple user notifications.
package notify

import "errors"

// ErrUnsupportedPlatform is returned on non-Windows builds.
var ErrUnsupportedPlatform = errors.New("picogogui/notify: notifications are only supported on Windows")

// Show displays a notification with title and body.
func Show(title, body string) error {
	return show(title, body)
}
