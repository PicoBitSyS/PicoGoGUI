package controls

import (
	"github.com/PicoBitSyS/PicoGoGUI/binding"
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

// CheckBox is a labeled boolean toggle.
type CheckBox struct {
	base
	text     string
	checked  bool
	bound    *binding.Var[bool]
	onChange func(bool)
	patcher  Patcher
	unsub    func()
	origin   binding.Origin
}

// NewCheckBox creates a checkbox with the given label.
//
// Example:
//
//	gui.CheckBox("Use SSL").Bind(ssl)
func NewCheckBox(text string) *CheckBox {
	return &CheckBox{
		base:   newBase("checkbox"),
		text:   text,
		origin: binding.NewOrigin(),
	}
}

// ID sets the component identifier.
func (c *CheckBox) ID(id string) *CheckBox {
	c.id = id
	return c
}

// Text sets the label text.
func (c *CheckBox) Text(text string) *CheckBox {
	c.text = text
	return c
}

// Checked sets the checked state.
func (c *CheckBox) Checked(v bool) *CheckBox {
	c.checked = v
	if c.bound != nil {
		c.bound.SetFrom(c.origin, v)
	}
	return c
}

// IsChecked returns the checked state.
func (c *CheckBox) IsChecked() bool {
	if c.bound != nil {
		return c.bound.Get()
	}
	return c.checked
}

// Visible sets visibility.
func (c *CheckBox) Visible(v bool) *CheckBox {
	c.visible = v
	return c
}

// Enabled sets enabled state.
func (c *CheckBox) Enabled(v bool) *CheckBox {
	c.enabled = v
	return c
}

// Appearance replaces the check box visual appearance.
func (c *CheckBox) Appearance(value Appearance) *CheckBox {
	c.appearance = value
	return c
}

// OnChange registers a change handler.
func (c *CheckBox) OnChange(fn func(bool)) *CheckBox {
	c.onChange = fn
	return c
}

// Bind attaches a binding.Var[bool] for two-way sync.
func (c *CheckBox) Bind(v *binding.Var[bool]) *CheckBox {
	if c.unsub != nil {
		c.unsub()
		c.unsub = nil
	}
	c.bound = v
	if v != nil {
		c.checked = v.Get()
		c.unsub = v.SubscribeFrom(c.origin, func(val bool) {
			c.checked = val
			if c.patcher != nil {
				_ = c.patcher.Patch(c.id, map[string]any{"checked": val})
			}
		})
	}
	return c
}

// AttachHost implements HostAware.
func (c *CheckBox) AttachHost(p Patcher) { c.patcher = p }

// Kind implements Component.
func (c *CheckBox) Kind() string { return "checkbox" }

// Node implements Component.
func (c *CheckBox) Node() Node {
	props := map[string]any{
		"text":    c.text,
		"checked": c.IsChecked(),
	}
	c.applyCommonProps(props)
	return Node{ID: c.id, Kind: c.Kind(), Props: props}
}

// CollectHandlers implements Component.
func (c *CheckBox) CollectHandlers(reg *events.Registry) {
	reg.OnChange(c.id, func(v any) {
		b, _ := v.(bool)
		c.checked = b
		if c.bound != nil {
			c.bound.SetFrom(c.origin, b)
		}
		if c.onChange != nil {
			c.onChange(b)
		}
	})
}

// Dispose releases the binding subscription and host reference.
func (c *CheckBox) Dispose() {
	if c.unsub != nil {
		c.unsub()
		c.unsub = nil
	}
	c.patcher = nil
}
