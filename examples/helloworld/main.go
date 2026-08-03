package main

import (
	"log"

	gui "github.com/PicoBitSyS/PicoGoGUI"
	"github.com/PicoBitSyS/PicoGoGUI/app"
)

func main() {
	application := app.New(app.Options{Debug: true})
	win := application.NewWindow("Hello PicoGoGUI")

	label := gui.Label("Hello, PicoGoGUI").ID("greeting")

	win.Add(
		label,
		gui.Button("Click me").OnClick(func() {
			label.Text("Clicked!")
			_ = win.Apply(label)
		}),
	)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
