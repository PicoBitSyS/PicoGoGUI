// Package app manages application lifecycle for PicoGoGUI.
package app

import (
	"errors"
	"sync"
	"time"

	"github.com/PicoBitSyS/PicoGoGUI/plugin"
	"github.com/PicoBitSyS/PicoGoGUI/theme"
	"github.com/PicoBitSyS/PicoGoGUI/window"
)

// App is the root application object.
type App struct {
	mu         sync.Mutex
	theme      theme.Name // may be System; windows receive Resolve(theme)
	debug      bool
	windows    []*window.Window
	startupErr error
}

// Options configures a new application.
type Options struct {
	Theme theme.Name
	Debug bool
}

// New creates an application.
//
// Example:
//
//	a := app.New(app.Options{Theme: theme.SystemTheme()})
func New(opts ...Options) *App {
	a := &App{theme: theme.Light}
	if len(opts) > 0 {
		if opts[0].Theme != "" {
			a.theme = opts[0].Theme
		}
		a.debug = opts[0].Debug
	}
	a.startupErr = plugin.Activate()
	return a
}

// Theme returns the configured theme name (may be System).
func (a *App) Theme() theme.Name {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.theme
}

// ResolvedTheme returns the concrete Dark or Light theme in effect.
func (a *App) ResolvedTheme() theme.Name {
	a.mu.Lock()
	name := a.theme
	a.mu.Unlock()
	return theme.Resolve(name)
}

// NewWindow creates a top-level window owned by the application.
//
// Example:
//
//	win := a.NewWindow("Server")
func (a *App) NewWindow(title string) *window.Window {
	a.mu.Lock()
	defer a.mu.Unlock()
	w := window.New(window.Config{
		Title: title,
		Theme: theme.Resolve(a.theme),
		Debug: a.debug,
	})
	a.windows = append(a.windows, w)
	return w
}

// SetTheme sets the default theme for subsequently created windows
// and updates already created windows.
//
// Example:
//
//	a.SetTheme(theme.DarkTheme())
func (a *App) SetTheme(name theme.Name) {
	a.mu.Lock()
	a.theme = name
	resolved := theme.Resolve(name)
	windows := append([]*window.Window(nil), a.windows...)
	a.mu.Unlock()
	for _, w := range windows {
		w.SetTheme(resolved)
	}
}

// Run shows all windows and blocks until the primary window closes.
//
// Example:
//
//	if err := a.Run(); err != nil { log.Fatal(err) }
func (a *App) Run() error {
	a.mu.Lock()
	if a.startupErr != nil {
		err := a.startupErr
		a.mu.Unlock()
		return err
	}
	windows := append([]*window.Window(nil), a.windows...)
	a.mu.Unlock()
	if len(windows) == 0 {
		w := a.NewWindow("PicoGoGUI")
		stopTheme := a.startThemeWatcher()
		err := w.Run()
		close(stopTheme)
		return errors.Join(err, plugin.Shutdown())
	}
	shown := make([]*window.Window, 0, len(windows)-1)
	for i := 1; i < len(windows); i++ {
		if err := windows[i].Show(); err != nil {
			for _, opened := range shown {
				_ = opened.Close()
			}
			return errors.Join(err, plugin.Shutdown())
		}
		shown = append(shown, windows[i])
	}
	stopTheme := a.startThemeWatcher()
	err := windows[0].Run()
	close(stopTheme)
	for _, opened := range shown {
		_ = opened.Close()
		opened.Dispose()
	}
	return errors.Join(err, plugin.Shutdown())
}

func (a *App) startThemeWatcher() chan struct{} {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		last := theme.Name("")
		for {
			select {
			case <-ticker.C:
				a.mu.Lock()
				configured := a.theme
				windows := append([]*window.Window(nil), a.windows...)
				a.mu.Unlock()
				if configured != theme.System {
					last = ""
					continue
				}
				resolved := theme.Resolve(configured)
				if resolved == last {
					continue
				}
				last = resolved
				for _, win := range windows {
					win.SetTheme(resolved)
				}
			case <-stop:
				return
			}
		}
	}()
	return stop
}
