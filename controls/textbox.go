package controls

import (
	"github.com/PicoBitSyS/PicoGoGUI/binding"
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

// TextBox is a single-line text input.
type TextBox struct {
	base
	value    string
	bound    *binding.Var[string]
	onChange func(string)
	patcher  Patcher
	unsub    func()
	origin   binding.Origin
}

// NewTextBox creates an empty text box.
//
// Example:
//
//	gui.TextBox().Bind(host)
func NewTextBox() *TextBox {
	return &TextBox{base: newBase("textbox"), origin: binding.NewOrigin()}
}

// ID sets the component identifier.
func (t *TextBox) ID(id string) *TextBox {
	t.id = id
	return t
}

// Value sets the text value.
func (t *TextBox) Value(v string) *TextBox {
	t.value = v
	if t.bound != nil {
		t.bound.SetFrom(t.origin, v)
	}
	return t
}

// GetValue returns the current text.
func (t *TextBox) GetValue() string {
	if t.bound != nil {
		return t.bound.Get()
	}
	return t.value
}

// Visible sets visibility.
func (t *TextBox) Visible(v bool) *TextBox {
	t.visible = v
	return t
}

// Enabled sets enabled state.
func (t *TextBox) Enabled(v bool) *TextBox {
	t.enabled = v
	return t
}

// Appearance replaces the text box visual appearance.
func (t *TextBox) Appearance(value Appearance) *TextBox {
	t.appearance = value
	return t
}

// OnChange registers a change handler.
func (t *TextBox) OnChange(fn func(string)) *TextBox {
	t.onChange = fn
	return t
}

// Bind attaches a binding.Var for two-way sync.
func (t *TextBox) Bind(v *binding.Var[string]) *TextBox {
	if t.unsub != nil {
		t.unsub()
		t.unsub = nil
	}
	t.bound = v
	if v != nil {
		t.value = v.Get()
		t.unsub = v.SubscribeFrom(t.origin, func(val string) {
			t.value = val
			if t.patcher != nil {
				_ = t.patcher.Patch(t.id, map[string]any{"value": val})
			}
		})
	}
	return t
}

// AttachHost implements HostAware.
func (t *TextBox) AttachHost(p Patcher) { t.patcher = p }

// Kind implements Component.
func (t *TextBox) Kind() string { return "textbox" }

// Node implements Component.
func (t *TextBox) Node() Node {
	props := map[string]any{"value": t.GetValue()}
	t.applyCommonProps(props)
	return Node{ID: t.id, Kind: t.Kind(), Props: props}
}

// CollectHandlers implements Component.
func (t *TextBox) CollectHandlers(reg *events.Registry) {
	reg.OnChange(t.id, func(v any) {
		s, _ := v.(string)
		t.value = s
		if t.bound != nil {
			t.bound.SetFrom(t.origin, s)
		}
		if t.onChange != nil {
			t.onChange(s)
		}
	})
}

// Dispose releases the binding subscription and host reference.
func (t *TextBox) Dispose() {
	if t.unsub != nil {
		t.unsub()
		t.unsub = nil
	}
	t.patcher = nil
}
