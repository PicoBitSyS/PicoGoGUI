package layout

import (
	"testing"

	"github.com/PicoBitSyS/PicoGoGUI/controls"
)

func TestFormNode(t *testing.T) {
	f := NewForm(
		NewField("Host", controls.NewTextBox().ID("host")),
		NewField("Port", controls.NewNumberBox().ID("port")),
	).ID("settings-form")
	n := f.Node()
	if n.Kind != "form" || n.ID != "settings-form" || len(n.Children) != 2 {
		t.Fatalf("form = %+v", n)
	}
	if n.Children[0].Kind != "field" {
		t.Fatalf("field kind = %q", n.Children[0].Kind)
	}
}

func TestFormValidation(t *testing.T) {
	field := NewField("Host", controls.NewTextBox()).Required(true)
	form := NewForm(field)
	if errors := form.Validate(); len(errors) != 1 || field.Error() == "" {
		t.Fatalf("validation errors = %#v", errors)
	}
}
