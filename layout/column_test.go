package layout

import (
	"testing"

	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

func TestRowAndGridNodes(t *testing.T) {
	row := NewRow(controls.NewLabel("A"), controls.NewButton("B")).ID("r1")
	n := row.Node()
	if n.Kind != "row" || n.ID != "r1" || len(n.Children) != 2 {
		t.Fatalf("row = %+v", n)
	}

	grid := NewGrid(2, controls.NewLabel("A"), controls.NewLabel("B"))
	gn := grid.Node()
	if gn.Kind != "grid" || gn.Props["columns"] != 2 {
		t.Fatalf("grid = %+v", gn)
	}
}

func TestColumnCollectsChildHandlers(t *testing.T) {
	called := false
	col := NewColumn(controls.NewButton("Go").ID("btn").OnClick(func() { called = true }))
	d := events.NewDispatcher()
	d.SetRegistry(controls.CollectAllHandlers(col))
	if !d.Dispatch("btn", "click", nil) || !called {
		t.Fatal("nested click not dispatched")
	}
}

func TestAdvancedLayouts(t *testing.T) {
	tabs := NewTabs(
		NewTab("One", controls.NewLabel("A")),
		NewTab("Two", controls.NewLabel("B")),
	).Selected(1)
	node := tabs.Node()
	if node.Kind != "tabs" || node.Props["selected"] != 1 || len(node.Children) != 2 {
		t.Fatalf("tabs node = %#v", node)
	}
	split := NewSplit(controls.NewLabel("A"), controls.NewLabel("B")).Vertical(true).Ratio(70)
	if split.Node().Props["ratio"] != 70 || split.Node().Props["vertical"] != true {
		t.Fatalf("split node = %#v", split.Node())
	}
	dock := NewDock(DockItem{Region: DockCenter, Child: controls.NewLabel("Center")})
	if dock.Node().Kind != "dock" || len(dock.Node().Children) != 1 {
		t.Fatalf("dock node = %#v", dock.Node())
	}
}
