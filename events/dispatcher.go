// Package events provides UI event dispatch helpers.
package events

// ChangeHandler receives a value from a valued UI event (change, select, dialog, …).
type ChangeHandler func(value any)

// Registry collects click and valued-event handlers by component ID.
type Registry struct {
	clicks map[string]func()
	valued map[string]map[string]ChangeHandler
}

// NewRegistry creates an empty handler registry.
func NewRegistry() *Registry {
	return &Registry{
		clicks: make(map[string]func()),
		valued: make(map[string]map[string]ChangeHandler),
	}
}

// OnClick registers a click handler for a component.
func (r *Registry) OnClick(id string, fn func()) {
	if r == nil || id == "" || fn == nil {
		return
	}
	r.clicks[id] = fn
}

// On registers a valued-event handler (change, select, toggle, dialog, …).
func (r *Registry) On(id, event string, fn ChangeHandler) {
	if r == nil || id == "" || event == "" || fn == nil {
		return
	}
	if r.valued[id] == nil {
		r.valued[id] = make(map[string]ChangeHandler)
	}
	r.valued[id][event] = fn
}

// OnChange registers a change handler for a component.
func (r *Registry) OnChange(id string, fn ChangeHandler) {
	r.On(id, "change", fn)
}

// OnSelect registers a select handler for a component.
func (r *Registry) OnSelect(id string, fn ChangeHandler) {
	r.On(id, "select", fn)
}

// Dispatcher routes bridge events to registered handlers.
type Dispatcher struct {
	reg *Registry
}

// NewDispatcher creates an empty dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{reg: NewRegistry()}
}

// SetRegistry replaces the handler registry.
func (d *Dispatcher) SetRegistry(reg *Registry) {
	if reg == nil {
		reg = NewRegistry()
	}
	d.reg = reg
}

// Registry returns the active registry (may be empty, never nil after New).
func (d *Dispatcher) Registry() *Registry {
	if d.reg == nil {
		d.reg = NewRegistry()
	}
	return d.reg
}

// Dispatch invokes the handler for target/event.
func (d *Dispatcher) Dispatch(target, event string, value any) bool {
	if d == nil || d.reg == nil {
		return false
	}
	if event == "click" {
		fn := d.reg.clicks[target]
		if fn == nil {
			return false
		}
		fn()
		return true
	}
	byEvent := d.reg.valued[target]
	if byEvent == nil {
		return false
	}
	fn := byEvent[event]
	if fn == nil {
		return false
	}
	fn(value)
	return true
}
