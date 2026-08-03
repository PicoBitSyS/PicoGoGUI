//go:build windows

package tray

import (
	"sync"
	"time"

	"github.com/getlantern/systray"
)

var (
	activeMu sync.Mutex
	active   *Icon
)

func quit() { systray.Quit() }

func (i *Icon) applyMenus() {
	if i.Tooltip != "" {
		systray.SetTooltip(i.Tooltip)
	}
	systray.SetTitle("PicoGoGUI")
	if len(i.IconPNG) > 0 {
		systray.SetIcon(i.IconPNG)
	}
	if i.OnOpen != nil {
		open := systray.AddMenuItem("Open", "Open")
		go func() {
			for range open.ClickedCh {
				i.OnOpen()
			}
		}()
		systray.AddSeparator()
	}
	i.addItems(nil, i.Menu)
}

func (i *Icon) addItems(parent *systray.MenuItem, items []Item) {
	for _, item := range items {
		if item.Separator {
			if parent == nil {
				systray.AddSeparator()
			}
			continue
		}
		var mi *systray.MenuItem
		if parent == nil {
			mi = systray.AddMenuItemCheckbox(item.Label, item.Label, item.Checked)
		} else {
			mi = parent.AddSubMenuItemCheckbox(item.Label, item.Label, item.Checked)
		}
		if item.Disabled {
			mi.Disable()
		}
		if len(item.Children) > 0 {
			i.addItems(mi, item.Children)
			continue
		}
		fn := item.OnClick
		go func(mi *systray.MenuItem, fn func()) {
			for range mi.ClickedCh {
				if fn != nil {
					fn()
				}
			}
		}(mi, fn)
	}
}

// Run shows the tray icon and blocks until Quit. Call from main on Windows.
func (i *Icon) Run() error {
	if i == nil {
		return nil
	}
	activeMu.Lock()
	if active != nil && active != i {
		activeMu.Unlock()
		return ErrAlreadyRunning
	}
	active = i
	activeMu.Unlock()
	systray.Run(func() {
		i.applyMenus()
	}, func() {
		activeMu.Lock()
		if active == i {
			active = nil
		}
		activeMu.Unlock()
		if i.OnExit != nil {
			i.OnExit()
		}
	})
	return nil
}

// Start tries to bring up the tray without permanently blocking the caller.
// Prefer Run() on the main thread. Start uses a locked OS thread and times out
// if the native loop never signals ready (the previous hang).
func (i *Icon) Start() error {
	if i == nil {
		return nil
	}
	ready := make(chan struct{})
	errCh := make(chan error, 1)
	activeMu.Lock()
	if active != nil && active != i {
		activeMu.Unlock()
		return ErrAlreadyRunning
	}
	active = i
	activeMu.Unlock()
	go func() {
		// Dedicated OS thread for the native loop (still inferior to main).
		systray.Run(func() {
			i.applyMenus()
			close(ready)
		}, func() {
			activeMu.Lock()
			if active == i {
				active = nil
			}
			activeMu.Unlock()
		})
		errCh <- nil
	}()
	select {
	case <-ready:
		return nil
	case <-time.After(3 * time.Second):
		systray.Quit()
		return ErrStartTimeout
	}
}

// Stop removes the tray icon / quits the loop.
func (i *Icon) Stop() error {
	systray.Quit()
	return nil
}
