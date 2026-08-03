package main

import (
	"log"

	gui "github.com/PicoBitSyS/PicoGoGUI"
	"github.com/PicoBitSyS/PicoGoGUI/examples/plugins/badge"
)

// Blank-import style also works via badge's init(); we keep the named import
// so we can construct badge.New in this example.
func main() {
	app := gui.New(gui.Options{Theme: gui.ThemeSystem()})
	win := app.NewWindow("Plugin Host")
	win.SetSize(420, 280)

	status := badge.New("LIVE").Tone(badge.ToneSuccess).ID("live")
	info := badge.New("Phase 6").Tone(badge.ToneInfo).ID("phase")
	warn := badge.New("Beta").Tone(badge.ToneWarn).ID("beta")

	win.Add(
		gui.Column(
			gui.Label("PicoGoGUI plugin demo"),
			gui.Row(status, info, warn),
			gui.Button("Mark paused").OnClick(func() {
				status.Tone(badge.ToneWarn).Text("PAUSED")
				status.Refresh()
			}),
			gui.Label("Import examples/plugins/badge (init registers the plugin)."),
		),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
