package main

import (
	"fmt"
	"log"

	gui "github.com/PicoBitSyS/PicoGoGUI"
)

type connection struct {
	Host   string
	Port   int
	Status string
}

func main() {
	app := gui.New(gui.Options{Theme: gui.ThemeSystem()})
	win := app.NewWindow("Connections")
	win.SetSize(720, 480)

	rows := []connection{
		{Host: "localhost", Port: 25, Status: "up"},
		{Host: "mail.example", Port: 587, Status: "up"},
		{Host: "backup", Port: 2525, Status: "down"},
	}

	status := gui.Label("Select a row").ID("status")
	table := gui.Table().
		ID("conns").
		Columns("Host", "Port", "Status").
		Bind(&rows).
		OnSelect(func(i int) {
			if i < 0 || i >= len(rows) {
				return
			}
			status.Text(fmt.Sprintf("Selected %s:%d (%s)", rows[i].Host, rows[i].Port, rows[i].Status))
			_ = win.Apply(status)
		})

	tree := gui.Tree().ID("nav").Nodes(
		gui.TreeNode("Servers",
			gui.TreeNode("SMTP"),
			gui.TreeNode("IMAP"),
		),
		gui.TreeNode("Clients"),
	).OnSelect(func(id string) {
		status.Text("Tree: " + id)
		_ = win.Apply(status)
	})

	win.Add(
		gui.Row(
			gui.Column(
				gui.Label("Directory"),
				tree,
			),
			gui.Column(
				gui.Label("Connections"),
				table,
				gui.Row(
					gui.Button("Add").OnClick(func() {
						rows = append(rows, connection{Host: "new", Port: 25, Status: "up"})
						table.Bind(&rows).Refresh()
						_ = win.Refresh()
						gui.Message(win, "Added", "New connection appended.")
					}),
					gui.Button("Delete").OnClick(func() {
						i := table.Selected()
						if i < 0 || i >= len(rows) {
							gui.Message(win, "Delete", "Select a row first.")
							return
						}
						host := rows[i].Host
						gui.Confirm(win, "Delete", "Remove "+host+"?", func(ok bool) {
							if !ok {
								return
							}
							rows = append(rows[:i], rows[i+1:]...)
							table.Bind(&rows).Refresh()
							_ = win.Refresh()
							status.Text("Deleted " + host)
							_ = win.Apply(status)
						})
					}),
				),
				status,
			),
		),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
