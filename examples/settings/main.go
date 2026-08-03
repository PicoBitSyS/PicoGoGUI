package main

import (
	"fmt"
	"log"

	gui "github.com/PicoBitSyS/PicoGoGUI"
	"github.com/PicoBitSyS/PicoGoGUI/binding"
)

func main() {
	app := gui.New(gui.Options{Theme: gui.ThemeSystem()})
	win := app.NewWindow("Settings")
	win.SetOuterSize(480, 420)

	host := binding.New("localhost")
	port := binding.New(25)
	ssl := binding.New(false)
	proto := binding.New("TCP")
	dark := binding.New(app.ResolvedTheme() == gui.ThemeDark())
	status := gui.Label("Ready").ID("status")
	win.OnResize(func(event gui.ResizeEvent) {
		status.Text(fmt.Sprintf("Window: %d x %d @ %d DPI", event.Outer.Width, event.Outer.Height, event.DPI))
		_ = win.Apply(status)
	})

	win.Add(
		gui.Column(
			gui.Label("Server"),
			gui.TextBox().ID("host").Bind(host),
			gui.Label("Port"),
			gui.NumberBox().ID("port").Value(25).Bind(port),
			gui.Label("Protocol"),
			gui.ComboBox("TCP", "UDP", "TLS").ID("proto").Bind(proto),
			gui.CheckBox("Use SSL").ID("ssl").Bind(ssl),
			gui.CheckBox("Dark theme").ID("dark").Bind(dark).OnChange(func(on bool) {
				if on {
					app.SetTheme(gui.ThemeDark())
				} else {
					app.SetTheme(gui.ThemeLight())
				}
				status.Text(fmt.Sprintf("Theme: %s", app.ResolvedTheme()))
				_ = win.Apply(status)
			}),
			gui.Row(
				gui.Button("Save").OnClick(func() {
					status.Text(fmt.Sprintf("Saved %s:%d (%s) ssl=%v",
						host.Get(), port.Get(), proto.Get(), ssl.Get()))
					_ = win.Apply(status)
				}),
				gui.Button("Reset").OnClick(func() {
					host.Set("localhost")
					port.Set(25)
					ssl.Set(false)
					proto.Set("TCP")
					status.Text("Reset")
					_ = win.Apply(status)
				}),
			),
			status,
		),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
