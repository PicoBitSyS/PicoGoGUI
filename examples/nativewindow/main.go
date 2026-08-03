package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	gui "github.com/PicoBitSyS/PicoGoGUI"
)

func main() {
	app := gui.New(gui.Options{Theme: gui.ThemeSystem()})
	win := app.NewWindow("Native Window")
	win.SetOuterSize(640, 480)

	statePath := os.Getenv("PICOGOGUI_WINDOW_STATE")
	if statePath == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			log.Fatal(err)
		}
		statePath = filepath.Join(configDir, "PicoGoGUI", "examples", "native-window.json")
	}
	if err := win.PersistSize(statePath); err != nil {
		log.Fatal(err)
	}

	metrics := gui.Label("Resize the native window").Class("window-metrics")
	win.OnResize(func(event gui.ResizeEvent) {
		metrics.Text(fmt.Sprintf(
			"Outer %d x %d · Client %d x %d · DPI %d",
			event.Outer.Width, event.Outer.Height,
			event.Client.Width, event.Client.Height,
			event.DPI,
		))
		_ = win.Apply(metrics)
	})

	win.Add(gui.Column(
		gui.Label("DPI-aware native geometry").Class("window-title").Bold(true),
		metrics,
		gui.Button("Close").OnClick(func() { _ = win.Close() }),
	))

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
