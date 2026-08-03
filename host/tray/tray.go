// Package tray provides the system-tray shell channel.
//
// On Windows, call Icon.Run() from the main goroutine — the native message
// loop must own the main thread. Icon.Start() is a non-blocking helper with
// a timeout and is less reliable on Windows.
//
// Note: Windows Session 0 (services) cannot host a real tray icon.
package tray

import "errors"

// ErrUnsupportedPlatform is returned on non-Windows builds.
var ErrUnsupportedPlatform = errors.New("picogogui/tray: system tray is only supported on Windows")

// ErrStartTimeout is returned when the tray native loop never becomes ready.
var ErrStartTimeout = errors.New("picogogui/tray: timed out waiting for tray ready (use Icon.Run on the main thread)")

// ErrAlreadyRunning is returned when another tray icon owns the global loop.
var ErrAlreadyRunning = errors.New("picogogui/tray: another tray icon is already running")

// Item is a context-menu entry.
type Item struct {
	Label   string
	OnClick func()
	// Separator renders a menu separator when true (Label ignored).
	Separator bool
	Disabled  bool
	Checked   bool
	Children  []Item
}

// Icon is the tray icon configuration and lifecycle handle.
type Icon struct {
	Tooltip string
	Menu    []Item
	// OnOpen is invoked from an "Open" menu item when set.
	OnOpen func()
	// OnExit is invoked when the tray loop ends (after Quit).
	OnExit func()
	// IconPNG optional tray icon bytes (PNG). Empty uses the default.
	IconPNG []byte
}

// New creates a tray icon configuration.
func New(tooltip string) *Icon {
	return &Icon{Tooltip: tooltip}
}

// Add appends menu items.
func (i *Icon) Add(items ...Item) *Icon {
	i.Menu = append(i.Menu, items...)
	return i
}

// Action creates a clickable menu item.
func Action(label string, fn func()) Item {
	return Item{Label: label, OnClick: fn}
}

// Separator creates a menu separator.
func Separator() Item {
	return Item{Separator: true}
}

// Submenu creates a nested menu.
func Submenu(label string, children ...Item) Item {
	return Item{Label: label, Children: children}
}

// Quit requests the tray message loop to exit (unblocks Run).
func Quit() { quit() }
