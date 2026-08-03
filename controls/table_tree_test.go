package controls

import (
	"testing"

	"github.com/PicoBitSyS/PicoGoGUI/events"
)

type connRow struct {
	Host   string
	Port   int
	Status string
}

func TestTableBindStructSlice(t *testing.T) {
	rows := []connRow{
		{Host: "localhost", Port: 25, Status: "up"},
		{Host: "mail", Port: 587, Status: "down"},
	}
	tbl := NewTable().ID("conns").Columns("Host", "Port", "Status").Bind(&rows)
	n := tbl.Node()
	gotRows, _ := n.Props["rows"].([]map[string]any)
	if len(gotRows) != 2 {
		t.Fatalf("rows = %#v", n.Props["rows"])
	}
	if gotRows[0]["Host"] != "localhost" || gotRows[0]["Port"] != 25 {
		t.Fatalf("row0 = %#v", gotRows[0])
	}

	selected := -1
	tbl.OnSelect(func(i int) { selected = i })
	d := events.NewDispatcher()
	d.SetRegistry(CollectAllHandlers(tbl))
	d.Dispatch("conns", "select", float64(1))
	if selected != 1 {
		t.Fatalf("selected = %d", selected)
	}
}

func TestTreeSerializeAndToggle(t *testing.T) {
	child := NewTreeNode("smtp").WithID("n-smtp")
	root := NewTreeNode("Servers", child).WithID("n-root")
	tr := NewTree().ID("tree").Nodes(root)
	props := tr.Node().Props
	nodes, _ := props["nodes"].([]map[string]any)
	if len(nodes) != 1 || nodes[0]["text"] != "Servers" {
		t.Fatalf("nodes = %#v", props["nodes"])
	}

	d := events.NewDispatcher()
	d.SetRegistry(CollectAllHandlers(tr))
	d.Dispatch("tree", "toggle", map[string]any{"id": "n-root", "expanded": false})
	if root.Expanded {
		t.Fatal("expected collapsed")
	}
}
