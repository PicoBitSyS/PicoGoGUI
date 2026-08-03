package controls

import (
	"github.com/PicoBitSyS/PicoGoGUI/binding"
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

// NumberBox is a numeric input.
type NumberBox struct {
	base
	value    float64
	bound    *binding.Var[int]
	boundF   *binding.Var[float64]
	onChange func(float64)
	patcher  Patcher
	unsub    func()
	origin   binding.Origin
}

// NewNumberBox creates a number box with value 0.
//
// Example:
//
//	gui.NumberBox().Value(25).Bind(port)
func NewNumberBox() *NumberBox {
	return &NumberBox{base: newBase("numberbox"), origin: binding.NewOrigin()}
}

// ID sets the component identifier.
func (n *NumberBox) ID(id string) *NumberBox {
	n.id = id
	return n
}

// Value sets the numeric value.
func (n *NumberBox) Value(v float64) *NumberBox {
	n.value = v
	if n.bound != nil {
		n.bound.SetFrom(n.origin, int(v))
	}
	if n.boundF != nil {
		n.boundF.SetFrom(n.origin, v)
	}
	return n
}

// GetValue returns the current number.
func (n *NumberBox) GetValue() float64 {
	if n.bound != nil {
		return float64(n.bound.Get())
	}
	if n.boundF != nil {
		return n.boundF.Get()
	}
	return n.value
}

// Visible sets visibility.
func (n *NumberBox) Visible(v bool) *NumberBox {
	n.visible = v
	return n
}

// Enabled sets enabled state.
func (n *NumberBox) Enabled(v bool) *NumberBox {
	n.enabled = v
	return n
}

// Appearance replaces the number box visual appearance.
func (n *NumberBox) Appearance(value Appearance) *NumberBox {
	n.appearance = value
	return n
}

// OnChange registers a change handler.
func (n *NumberBox) OnChange(fn func(float64)) *NumberBox {
	n.onChange = fn
	return n
}

// Bind attaches a binding.Var[int] for two-way sync.
func (n *NumberBox) Bind(v *binding.Var[int]) *NumberBox {
	n.clearSub()
	n.bound = v
	n.boundF = nil
	if v != nil {
		n.value = float64(v.Get())
		n.unsub = v.SubscribeFrom(n.origin, func(val int) {
			n.value = float64(val)
			if n.patcher != nil {
				_ = n.patcher.Patch(n.id, map[string]any{"value": val})
			}
		})
	}
	return n
}

// BindFloat attaches a binding.Var[float64] for two-way sync.
func (n *NumberBox) BindFloat(v *binding.Var[float64]) *NumberBox {
	n.clearSub()
	n.boundF = v
	n.bound = nil
	if v != nil {
		n.value = v.Get()
		n.unsub = v.SubscribeFrom(n.origin, func(val float64) {
			n.value = val
			if n.patcher != nil {
				_ = n.patcher.Patch(n.id, map[string]any{"value": val})
			}
		})
	}
	return n
}

func (n *NumberBox) clearSub() {
	if n.unsub != nil {
		n.unsub()
		n.unsub = nil
	}
}

// AttachHost implements HostAware.
func (n *NumberBox) AttachHost(p Patcher) { n.patcher = p }

// Kind implements Component.
func (n *NumberBox) Kind() string { return "numberbox" }

// Node implements Component.
func (n *NumberBox) Node() Node {
	props := map[string]any{"value": n.GetValue()}
	n.applyCommonProps(props)
	return Node{ID: n.id, Kind: n.Kind(), Props: props}
}

// CollectHandlers implements Component.
func (n *NumberBox) CollectHandlers(reg *events.Registry) {
	reg.OnChange(n.id, func(v any) {
		f, ok := asFloat64(v)
		if !ok {
			return
		}
		n.value = f
		if n.bound != nil {
			n.bound.SetFrom(n.origin, int(f))
		}
		if n.boundF != nil {
			n.boundF.SetFrom(n.origin, f)
		}
		if n.onChange != nil {
			n.onChange(f)
		}
	})
}

// Dispose releases the binding subscription and host reference.
func (n *NumberBox) Dispose() {
	n.clearSub()
	n.patcher = nil
}

func asFloat64(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case int32:
		return float64(x), true
	default:
		return 0, false
	}
}
