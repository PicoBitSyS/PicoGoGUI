package controls

import "github.com/PicoBitSyS/PicoGoGUI/events"

// Label displays read-only text.
type Label struct {
	base
	text string
}

// NewLabel creates a label with the given text.
//
// Example:
//
//	gui.Label("Hello")
func NewLabel(text string) *Label {
	return &Label{
		base: newBase("label"),
		text: text,
	}
}

// ID sets the component identifier and returns the label for chaining.
func (l *Label) ID(id string) *Label {
	l.id = id
	return l
}

// Text sets the label text and returns the label for chaining.
func (l *Label) Text(text string) *Label {
	l.text = text
	return l
}

// GetText returns the current label text.
func (l *Label) GetText() string { return l.text }

// Visible sets visibility and returns the label for chaining.
func (l *Label) Visible(v bool) *Label {
	l.visible = v
	return l
}

// Enabled sets the enabled state and returns the label for chaining.
func (l *Label) Enabled(v bool) *Label {
	l.enabled = v
	return l
}

// Appearance replaces the visual appearance and returns the label for chaining.
func (l *Label) Appearance(value Appearance) *Label {
	l.appearance = value
	return l
}

// Font sets the font family.
func (l *Label) Font(family string) *Label {
	l.appearance.FontFamily = family
	return l
}

// FontSize sets the font size in CSS pixels.
func (l *Label) FontSize(size int) *Label {
	l.appearance.FontSize = size
	return l
}

// Color sets the text color using any valid CSS color.
func (l *Label) Color(color string) *Label {
	l.appearance.Color = color
	return l
}

// Background sets the label background using any valid CSS color.
func (l *Label) Background(color string) *Label {
	l.appearance.Background = color
	return l
}

// Bold enables or disables bold text.
func (l *Label) Bold(value bool) *Label {
	l.appearance.Bold = value
	return l
}

// Italic enables or disables italic text.
func (l *Label) Italic(value bool) *Label {
	l.appearance.Italic = value
	return l
}

// Underline enables or disables underlined text.
func (l *Label) Underline(value bool) *Label {
	l.appearance.Underline = value
	return l
}

// TextAlign sets horizontal alignment: left, center, right, or justify.
func (l *Label) TextAlign(value string) *Label {
	l.appearance.TextAlign = value
	return l
}

// Border sets the border width, color, and radius in CSS pixels.
func (l *Label) Border(width int, color string, radius int) *Label {
	l.appearance.BorderWidth = width
	l.appearance.BorderColor = color
	l.appearance.BorderRadius = radius
	return l
}

// Opacity sets opacity from 0 (transparent) to 1 (opaque).
func (l *Label) Opacity(value float64) *Label {
	l.appearance.Opacity = value
	return l
}

// Kind implements Component.
func (l *Label) Kind() string { return "label" }

// Node implements Component.
func (l *Label) Node() Node {
	props := map[string]any{"text": l.text}
	l.applyCommonProps(props)
	return Node{ID: l.id, Kind: l.Kind(), Props: props}
}

// CollectHandlers implements Component.
func (l *Label) CollectHandlers(*events.Registry) {}
