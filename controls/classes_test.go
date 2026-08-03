package controls

import "testing"

func TestClassIsSerializedForBuiltInControls(t *testing.T) {
	components := []Component{
		NewLabel("Label").Class("title  muted"),
		NewButton("Button").Class("primary"),
		NewTextBox().Class("input-wide"),
		NewNumberBox().Class("numeric"),
		NewCheckBox("Check").Class("choice"),
		NewComboBox("A").Class("selector"),
		NewTable().Class("striped"),
		NewTree().Class("navigation"),
		NewDesignSurface().Class("designer"),
		NewDropZone(NewLabel("Drop")).Class("drop-target"),
	}
	want := []string{"title muted", "primary", "input-wide", "numeric", "choice", "selector", "striped", "navigation", "designer", "drop-target"}
	for i, component := range components {
		if got := component.Node().Props["class"]; got != want[i] {
			t.Errorf("%s class=%v, want %q", component.Kind(), got, want[i])
		}
	}
}
