// Package theme provides Go-facing theme selection. CSS variables stay internal.
package theme

import (
	"errors"
	"strings"
	"sync"
)

// Name identifies a built-in or custom theme.
type Name string

const (
	// Dark is the dark theme.
	Dark Name = "dark"
	// Light is the light theme.
	Light Name = "light"
	// System follows the Windows AppsUseLightTheme setting.
	System Name = "system"
)

// Definition contains strongly typed visual tokens for a custom theme.
type Definition struct {
	Name         Name
	Dark         bool
	Background   string
	Foreground   string
	Muted        string
	Accent       string
	Control      string
	ControlHover string
	Input        string
	Border       string
	BorderBottom string
	Radius       string
	Gap          string
	Padding      string
	ControlWidth string
}

var customThemes = struct {
	sync.RWMutex
	values map[Name]Definition
}{values: make(map[Name]Definition)}

// Register validates and registers a custom theme definition.
func Register(def Definition) (Name, error) {
	def.Name = Name(strings.TrimSpace(string(def.Name)))
	if def.Name == "" || def.Name == Light || def.Name == Dark || def.Name == System {
		return "", errors.New("theme: custom name must be non-empty and not reserved")
	}
	customThemes.Lock()
	customThemes.values[def.Name] = def
	customThemes.Unlock()
	return def.Name, nil
}

// Lookup returns a registered custom theme.
func Lookup(name Name) (Definition, bool) {
	customThemes.RLock()
	def, ok := customThemes.values[name]
	customThemes.RUnlock()
	return def, ok
}

// Variables returns internal CSS variable values for a custom theme.
func Variables(name Name) map[string]string {
	def, ok := Lookup(name)
	if !ok {
		return nil
	}
	values := map[string]string{
		"--pico-bg":            def.Background,
		"--pico-fg":            def.Foreground,
		"--pico-muted":         def.Muted,
		"--pico-accent":        def.Accent,
		"--pico-control-bg":    def.Control,
		"--pico-control-hover": def.ControlHover,
		"--pico-input-bg":      def.Input,
		"--pico-border":        def.Border,
		"--pico-border-bottom": def.BorderBottom,
		"--pico-radius":        def.Radius,
		"--pico-gap":           def.Gap,
		"--pico-pad":           def.Padding,
		"--pico-control-width": def.ControlWidth,
	}
	for key, value := range values {
		if strings.TrimSpace(value) == "" {
			delete(values, key)
		}
	}
	return values
}

// DarkTheme returns the dark theme name.
//
// Example:
//
//	gui.ThemeDark()
func DarkTheme() Name { return Dark }

// LightTheme returns the light theme name.
//
// Example:
//
//	gui.ThemeLight()
func LightTheme() Name { return Light }

// SystemTheme returns the system-following theme name.
//
// Example:
//
//	gui.ThemeSystem()
func SystemTheme() Name { return System }

// Resolve maps System (or empty) to the concrete Dark or Light theme.
// Dark and Light are returned unchanged.
func Resolve(name Name) Name {
	switch name {
	case Dark, Light:
		return name
	case System, "":
		return detectSystem()
	default:
		if _, ok := Lookup(name); ok {
			return name
		}
		return Light
	}
}

// IsDark reports whether the resolved theme is dark.
func IsDark(name Name) bool {
	resolved := Resolve(name)
	if resolved == Dark {
		return true
	}
	def, ok := Lookup(resolved)
	return ok && def.Dark
}
