package controls

import "github.com/PicoBitSyS/PicoGoGUI/events"

// Button is a clickable button control.
type Button struct {
	base
	text    string
	onClick func()
}

// NewButton creates a button with the given caption.
//
// Example:
//
//	gui.Button("Start").OnClick(func() { ... })
func NewButton(text string) *Button {
	return &Button{
		base: newBase("button"),
		text: text,
	}
}

// ID sets the component identifier and returns the button for chaining.
func (b *Button) ID(id string) *Button {
	b.id = id
	return b
}

// Text sets the button caption and returns the button for chaining.
func (b *Button) Text(text string) *Button {
	b.text = text
	return b
}

// Visible sets visibility and returns the button for chaining.
func (b *Button) Visible(v bool) *Button {
	b.visible = v
	return b
}

// Enabled sets the enabled state and returns the button for chaining.
func (b *Button) Enabled(v bool) *Button {
	b.enabled = v
	return b
}

// Appearance replaces the button visual appearance.
func (b *Button) Appearance(value Appearance) *Button {
	b.appearance = value
	return b
}

// OnClick registers a click handler and returns the button for chaining.
func (b *Button) OnClick(fn func()) *Button {
	b.onClick = fn
	return b
}

// Kind implements Component.
func (b *Button) Kind() string { return "button" }

// Node implements Component.
func (b *Button) Node() Node {
	props := map[string]any{"text": b.text}
	b.applyCommonProps(props)
	return Node{ID: b.id, Kind: b.Kind(), Props: props}
}

// CollectHandlers implements Component.
func (b *Button) CollectHandlers(reg *events.Registry) {
	if b.onClick != nil {
		reg.OnClick(b.id, b.onClick)
	}
}
