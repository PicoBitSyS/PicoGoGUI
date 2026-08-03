// Package web embeds the PicoGoGUI web runtime assets.
package web

import "embed"

// FS contains index.html, app.js, and theme.css.
//
//go:embed index.html app.js theme.css
var FS embed.FS

// MustRead returns the contents of a runtime file or panics.
func MustRead(name string) string {
	b, err := FS.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return string(b)
}
