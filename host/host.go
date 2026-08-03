// Package host groups multi-shell channels for PicoGoGUI apps:
// Desktop (gui), WebUI (HTTP), Tray, and Windows Service.
//
// Web UI does not render the declarative gui component tree — it shares
// Go business logic only.
//
// Service mode uses Session 0 (no tray). Interactive builds may combine
// WebUI + Tray (+ Desktop). See examples/servermanager.
package host

import (
	"github.com/PicoBitSyS/PicoGoGUI/host/service"
	"github.com/PicoBitSyS/PicoGoGUI/host/tray"
	"github.com/PicoBitSyS/PicoGoGUI/host/webui"
)

// Shell flags select which channels an interactive process enables.
type Shell uint

const (
	// Desktop enables PicoGoGUI windows.
	Desktop Shell = 1 << iota
	// WebUI enables the HTTP/Web channel.
	WebUI
	// Tray enables the system tray channel.
	Tray
)

// Common re-exports for convenience.
type (
	WebServer = webui.Server
	TrayIcon  = tray.Icon
	SvcConfig = service.Config
)

// NewWebUI constructs a Web UI server.
func NewWebUI() *webui.Server { return webui.New() }

// NewTray constructs a tray icon configuration.
func NewTray(tooltip string) *tray.Icon { return tray.New(tooltip) }
