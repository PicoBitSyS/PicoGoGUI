package designer

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/plugin"
)

func TestGenerateGo(t *testing.T) {
	d := NewDocument("Demo")
	d.Add(Widget{Kind: KindLabel, Text: "Hello", Class: "hero  muted"})
	d.Add(Widget{Kind: KindButton, Text: "OK"})
	src := d.GenerateGo()
	for _, need := range []string{
		`gui.New(gui.Options{Theme: gui.ThemeSystem()})`,
		`app.NewWindow("Demo")`,
		`win.SetSize(480, 360)`,
		`gui.Label("Hello")`,
		`.Class("hero muted")`,
		`gui.Button("OK")`,
		`app.Run()`,
	} {
		if !strings.Contains(src, need) {
			t.Fatalf("missing %q in:\n%s", need, src)
		}
	}
}

func TestGenerateGoContainers(t *testing.T) {
	d := NewDocument("Demo")
	d.Add(Widget{Kind: KindColumn, ID: "column1", Text: "Main", X: 10, Y: 10, Width: 200, Height: 160})
	d.Add(Widget{Kind: KindButton, Text: "OK", Parent: "column1", X: 8, Y: 24, Width: 80, Height: 28})
	src := d.GenerateGo()
	for _, need := range []string{
		`gui.Column(`,
		`gui.Button("OK")`,
		`.ID("column1")`,
	} {
		if !strings.Contains(src, need) {
			t.Fatalf("missing %q in:\n%s", need, src)
		}
	}
}

func TestJSONRoundTrip(t *testing.T) {
	d := NewDocument("Win")
	d.Add(Widget{Kind: KindTextBox, ID: "host", Value: "localhost", X: 20, Y: 40, Width: 120, Height: 28})
	raw, err := d.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.WindowTitle != "Win" || len(got.Widgets) != 1 || got.Widgets[0].Value != "localhost" {
		t.Fatalf("got %#v", got)
	}
	if got.Widgets[0].X != 20 || got.Widgets[0].Width != 120 {
		t.Fatalf("geometry %#v", got.Widgets[0])
	}
}

func TestRemoveCascade(t *testing.T) {
	d := NewDocument("W")
	d.Add(Widget{Kind: KindColumn, ID: "c1"})
	d.Add(Widget{Kind: KindButton, Parent: "c1"})
	d.RemoveAt(0)
	if len(d.Widgets) != 0 {
		t.Fatalf("expected empty, got %#v", d.Widgets)
	}
}

func TestSetGeometry(t *testing.T) {
	d := NewDocument("W")
	d.Add(Widget{Kind: KindButton})
	d.SetGeometry(0, 30, 40, 90, 28)
	w := d.Widgets[0]
	if w.X != 30 || w.Y != 40 || w.Width != 90 || w.Height != 28 {
		t.Fatalf("%#v", w)
	}
}

func TestGenerateGoPluginKind(t *testing.T) {
	plugin.ResetForTest()
	if err := plugin.Use(&stubBadgePlugin{}); err != nil {
		t.Fatal(err)
	}
	d := NewDocument("Demo")
	d.Add(Widget{Kind: "badge", Text: "LIVE", Value: "success", ID: "badge1"})
	src := d.GenerateGo()
	for _, need := range []string{
		`github.com/PicoBitSyS/PicoGoGUI/examples/plugins/badge`,
		`badge.New("LIVE").Tone("success").ID("badge1")`,
	} {
		if !strings.Contains(src, need) {
			t.Fatalf("missing %q in:\n%s", need, src)
		}
	}
}

func TestGenerateGoPreservesGeometryAndParses(t *testing.T) {
	d := NewDocument("Geometry")
	d.Add(Widget{Kind: KindButton, ID: "save", Text: "Save", X: 30, Y: 40, Width: 90, Height: 32})
	src := d.GenerateGo()
	if !strings.Contains(src, "gui.Canvas(") || !strings.Contains(src, "30, 40, 90, 32") {
		t.Fatalf("geometry missing from generated source:\n%s", src)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", src, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, src)
	}
}

func TestValidateDuplicateAndUndoRedo(t *testing.T) {
	d := NewDocument("History")
	d.Add(Widget{Kind: KindLabel, ID: "first"})
	d.Add(Widget{Kind: KindButton, ID: "second"})
	if !d.CanUndo() || !d.Undo() || len(d.Widgets) != 1 {
		t.Fatalf("undo failed: %#v", d.Widgets)
	}
	if !d.CanRedo() || !d.Redo() || len(d.Widgets) != 2 {
		t.Fatalf("redo failed: %#v", d.Widgets)
	}
	d.Widgets[1].ID = "first"
	if err := d.Validate(); err == nil {
		t.Fatal("expected duplicate id validation error")
	}
}

func TestContainersStayBehindControls(t *testing.T) {
	d := NewDocument("Layers")
	d.Add(Widget{Kind: KindButton, ID: "button1", ZIndex: -500})
	d.Add(Widget{Kind: KindColumn, ID: "panel1", ZIndex: 500})
	d.Add(Widget{Kind: KindColumn, ID: "panel2", ZIndex: 600})

	roots := d.ChildrenOf("")
	if len(roots) != 3 || roots[0].ID != "panel1" || roots[1].ID != "panel2" || roots[2].ID != "button1" {
		t.Fatalf("unexpected layer order: %#v", roots)
	}
	if effectiveZIndex(d.Widgets[1]) >= effectiveZIndex(d.Widgets[0]) {
		t.Fatalf("container crossed the control layer: container=%d control=%d",
			effectiveZIndex(d.Widgets[1]), effectiveZIndex(d.Widgets[0]))
	}
	if !d.BringToFront(1) || d.Widgets[1].ZIndex <= d.Widgets[2].ZIndex {
		t.Fatalf("container bring-to-front failed: %#v", d.Widgets)
	}
	if effectiveZIndex(d.Widgets[1]) >= effectiveZIndex(d.Widgets[0]) {
		t.Fatal("container bring-to-front crossed into the control layer")
	}
}

func TestAppearanceRoundTripAndCodegen(t *testing.T) {
	d := NewDocument("Styled")
	d.Add(Widget{
		Kind: KindLabel,
		ID:   "heading",
		Text: "Heading",
		Appearance: controls.Appearance{
			FontFamily: "Segoe UI",
			FontSize:   20,
			Color:      "#123456",
			Bold:       true,
			Italic:     true,
			TextAlign:  "center",
		},
	})
	raw, err := d.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	got := parsed.Widgets[0].Appearance
	if got.FontFamily != "Segoe UI" || got.FontSize != 20 || !got.Bold || got.Color != "#123456" {
		t.Fatalf("appearance was not preserved: %#v", got)
	}
	src := parsed.GenerateGo()
	for _, want := range []string{
		`Appearance(gui.Appearance{`,
		`FontFamily: "Segoe UI"`,
		`FontSize: 20`,
		`Color: "#123456"`,
		`Bold: true`,
		`.ZIndex(`,
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("generated source missing %q:\n%s", want, src)
		}
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "styled.go", src, parser.AllErrors); err != nil {
		t.Fatalf("styled generated source does not parse: %v\n%s", err, src)
	}
}

func TestAlignDistributeAndLockedWidgets(t *testing.T) {
	d := NewDocument("Arrange")
	d.Widgets = []Widget{
		{Kind: KindButton, ID: "a", X: 10, Y: 10, Width: 20, Height: 20},
		{Kind: KindButton, ID: "b", X: 100, Y: 40, Width: 20, Height: 20},
		{Kind: KindButton, ID: "c", X: 260, Y: 80, Width: 20, Height: 20},
	}
	if !d.Align([]int{0, 1, 2}, AlignTop) {
		t.Fatal("expected alignment to change geometry")
	}
	for _, widget := range d.Widgets {
		if widget.Y != 10 {
			t.Fatalf("top alignment failed: %#v", d.Widgets)
		}
	}
	if !d.Undo() || d.Widgets[1].Y != 40 {
		t.Fatalf("alignment was not one undo step: %#v", d.Widgets)
	}
	if !d.Distribute([]int{0, 1, 2}, DistributeHorizontal) {
		t.Fatal("expected horizontal distribution")
	}
	if d.Widgets[1].X != 135 {
		t.Fatalf("unexpected distributed position: %#v", d.Widgets)
	}

	d.Widgets[1].Locked = true
	before := d.Widgets[1]
	if !d.Align([]int{0, 1, 2}, AlignLeft) {
		t.Fatal("expected unlocked controls to align")
	}
	if d.Widgets[1] != before {
		t.Fatal("locked widget moved")
	}
	if d.Widgets[2].X != d.Widgets[0].X {
		t.Fatal("unlocked widgets did not align")
	}
}

func TestGroupGeometryLockHideAndCodegen(t *testing.T) {
	d := NewDocument("Group")
	d.Add(Widget{Kind: KindLabel, ID: "one", X: 10, Y: 10, Width: 80, Height: 22})
	d.Add(Widget{Kind: KindButton, ID: "two", X: 100, Y: 10, Width: 80, Height: 28})
	if !d.SetGeometries([]GeometryChange{
		{Index: 0, X: 20, Y: 30, Width: 80, Height: 22},
		{Index: 1, X: 110, Y: 30, Width: 80, Height: 28},
	}) {
		t.Fatal("expected group geometry update")
	}
	if !d.Undo() || d.Widgets[0].X != 10 || d.Widgets[1].X != 100 {
		t.Fatalf("group move was not atomic: %#v", d.Widgets)
	}
	if !d.SetLocked([]int{0}, true) || !d.SetHidden([]int{1}, true) {
		t.Fatal("lock/hide update failed")
	}
	if d.SetGeometries([]GeometryChange{{Index: 0, X: 99, Y: 99, Width: 80, Height: 22}}) {
		t.Fatal("locked widget accepted a geometry update")
	}
	raw, err := d.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.Widgets[0].Locked || !parsed.Widgets[1].Hidden {
		t.Fatalf("lock/hide were not persisted: %#v", parsed.Widgets)
	}
	if src := parsed.GenerateGo(); !strings.Contains(src, `.Visible(false)`) {
		t.Fatalf("hidden state missing from codegen:\n%s", src)
	}
}

type stubBadgePlugin struct{}

func (stubBadgePlugin) Info() plugin.Info {
	return plugin.Info{Name: "badge-stub", Version: "0"}
}

func (stubBadgePlugin) Contribute(h plugin.Host) error {
	return h.RegisterDesignerKind(plugin.DesignerKind{
		Kind:     "badge",
		Label:    "Badge",
		GoImport: "github.com/PicoBitSyS/PicoGoGUI/examples/plugins/badge",
		GoExpr:   `badge.New(%[1]q).Tone(%[2]q).ID(%[3]q)`,
	})
}
