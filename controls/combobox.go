package controls

import (
	"github.com/PicoBitSyS/PicoGoGUI/binding"
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

// ComboBox is a drop-down selection control.
type ComboBox struct {
	base
	items    []string
	value    string
	bound    *binding.Var[string]
	onChange func(string)
	patcher  Patcher
	unsub    func()
	origin   binding.Origin
}

// NewComboBox creates a combo box with the given items.
//
// Example:
//
//	gui.ComboBox("TCP", "UDP").Bind(proto)
func NewComboBox(items ...string) *ComboBox {
	c := &ComboBox{
		base:   newBase("combobox"),
		items:  append([]string(nil), items...),
		origin: binding.NewOrigin(),
	}
	if len(items) > 0 {
		c.value = items[0]
	}
	return c
}

// ID sets the component identifier.
func (c *ComboBox) ID(id string) *ComboBox {
	c.id = id
	return c
}

// Items replaces the option list.
func (c *ComboBox) Items(items ...string) *ComboBox {
	c.items = append([]string(nil), items...)
	return c
}

// Value sets the selected value.
func (c *ComboBox) Value(v string) *ComboBox {
	c.value = v
	if c.bound != nil {
		c.bound.SetFrom(c.origin, v)
	}
	return c
}

// GetValue returns the selected value.
func (c *ComboBox) GetValue() string {
	if c.bound != nil {
		return c.bound.Get()
	}
	return c.value
}

// Visible sets visibility.
func (c *ComboBox) Visible(v bool) *ComboBox {
	c.visible = v
	return c
}

// Enabled sets enabled state.
func (c *ComboBox) Enabled(v bool) *ComboBox {
	c.enabled = v
	return c
}

// Appearance replaces the combo box visual appearance.
func (c *ComboBox) Appearance(value Appearance) *ComboBox {
	c.appearance = value
	return c
}

// OnChange registers a change handler.
func (c *ComboBox) OnChange(fn func(string)) *ComboBox {
	c.onChange = fn
	return c
}

// Bind attaches a binding.Var[string] for two-way sync.
func (c *ComboBox) Bind(v *binding.Var[string]) *ComboBox {
	if c.unsub != nil {
		c.unsub()
		c.unsub = nil
	}
	c.bound = v
	if v != nil {
		c.value = v.Get()
		c.unsub = v.SubscribeFrom(c.origin, func(val string) {
			c.value = val
			if c.patcher != nil {
				_ = c.patcher.Patch(c.id, map[string]any{"value": val})
			}
		})
	}
	return c
}

// AttachHost implements HostAware.
func (c *ComboBox) AttachHost(p Patcher) { c.patcher = p }

// Kind implements Component.
func (c *ComboBox) Kind() string { return "combobox" }

// Node implements Component.
func (c *ComboBox) Node() Node {
	props := map[string]any{
		"value": c.GetValue(),
		"items": append([]string(nil), c.items...),
	}
	c.applyCommonProps(props)
	return Node{ID: c.id, Kind: c.Kind(), Props: props}
}

// CollectHandlers implements Component.
func (c *ComboBox) CollectHandlers(reg *events.Registry) {
	reg.OnChange(c.id, func(v any) {
		s, _ := v.(string)
		c.value = s
		if c.bound != nil {
			c.bound.SetFrom(c.origin, s)
		}
		if c.onChange != nil {
			c.onChange(s)
		}
	})
}

// Dispose releases the binding subscription and host reference.
func (c *ComboBox) Dispose() {
	if c.unsub != nil {
		c.unsub()
		c.unsub = nil
	}
	c.patcher = nil
}
