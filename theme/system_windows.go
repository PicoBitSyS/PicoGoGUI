//go:build windows

package theme

import (
	"golang.org/x/sys/windows/registry"
)

// detectSystem reads the Windows light/dark apps preference.
func detectSystem() Name {
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return Light
	}
	defer k.Close()

	v, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return Light
	}
	if v == 0 {
		return Dark
	}
	return Light
}
