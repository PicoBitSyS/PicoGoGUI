package controls

import "testing"

func TestLabelAppearance(t *testing.T) {
	label := NewLabel("Title").
		Font("Segoe UI").
		FontSize(18).
		Color("#123456").
		Background("#ffffff").
		Bold(true).
		Italic(true).
		Underline(true).
		TextAlign("center").
		Border(1, "#cccccc", 4).
		Opacity(0.8)

	value, ok := label.Node().Props["appearance"].(Appearance)
	if !ok {
		t.Fatalf("appearance has unexpected type: %#v", label.Node().Props["appearance"])
	}
	if value.FontFamily != "Segoe UI" || value.FontSize != 18 || !value.Bold || !value.Italic ||
		!value.Underline || value.TextAlign != "center" || value.BorderRadius != 4 || value.Opacity != 0.8 {
		t.Fatalf("unexpected appearance: %#v", value)
	}
}

func TestDesignSurfaceSelectionProps(t *testing.T) {
	surface := NewDesignSurface().Selection(1, 3, 5)
	props := surface.Node().Props
	selection, ok := props["selection"].([]int)
	if !ok || len(selection) != 3 || selection[2] != 5 {
		t.Fatalf("unexpected selection props: %#v", props["selection"])
	}
	if props["selected"] != 5 {
		t.Fatalf("unexpected primary selection: %#v", props["selected"])
	}
}
