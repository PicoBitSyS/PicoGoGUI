// Package layout provides layout containers for arranging components.
package layout

import (
	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

type baseLayout struct {
	id         string
	kind       string
	class      string
	children   []controls.Component
	visible    bool
	enabled    bool
	appearance controls.Appearance
}

func newBaseLayout(kind string, children []controls.Component) baseLayout {
	return baseLayout{
		id:       controls.AllocateID(kind),
		kind:     kind,
		children: children,
		visible:  true,
		enabled:  true,
	}
}

func (b *baseLayout) CompID() string { return b.id }

func (b *baseLayout) Kind() string { return b.kind }

func (b *baseLayout) ChildComponents() []controls.Component { return b.children }

func (b *baseLayout) Node() controls.Node {
	nodes := make([]controls.Node, 0, len(b.children))
	for _, ch := range b.children {
		if ch == nil {
			continue
		}
		nodes = append(nodes, ch.Node())
	}
	props := map[string]any{
		"visible": b.visible, "enabled": b.enabled, "appearance": b.appearance,
	}
	if b.class != "" {
		props["class"] = b.class
	}
	return controls.Node{
		ID:       b.id,
		Kind:     b.kind,
		Props:    props,
		Children: nodes,
	}
}

func (b *baseLayout) CollectHandlers(reg *events.Registry) {
	for _, ch := range b.children {
		if ch == nil {
			continue
		}
		ch.CollectHandlers(reg)
	}
}

// Column stacks children vertically.
type Column struct{ baseLayout }

// NewColumn creates a vertical stack of components.
//
// Example:
//
//	layout.NewColumn(gui.Label("A"), gui.Button("B"))
func NewColumn(children ...controls.Component) *Column {
	return &Column{baseLayout: newBaseLayout("column", children)}
}

// ID sets the column identifier.
func (c *Column) ID(id string) *Column {
	c.id = id
	return c
}

// Class sets extra CSS classes for the column container.
func (c *Column) Class(class string) *Column {
	c.class = class
	return c
}

// Appearance replaces the column visual appearance.
func (c *Column) Appearance(value controls.Appearance) *Column {
	c.appearance = value
	return c
}

// Visible sets column visibility.
func (c *Column) Visible(value bool) *Column {
	c.visible = value
	return c
}

// Row arranges children horizontally.
type Row struct{ baseLayout }

// NewRow creates a horizontal row of components.
func NewRow(children ...controls.Component) *Row {
	return &Row{baseLayout: newBaseLayout("row", children)}
}

// ID sets the row identifier.
func (r *Row) ID(id string) *Row {
	r.id = id
	return r
}

// Class sets extra CSS classes for the row container.
func (r *Row) Class(class string) *Row {
	r.class = class
	return r
}

// Appearance replaces the row visual appearance.
func (r *Row) Appearance(value controls.Appearance) *Row {
	r.appearance = value
	return r
}

// Visible sets row visibility.
func (r *Row) Visible(value bool) *Row {
	r.visible = value
	return r
}

// Stack overlays children in a single cell (first visible wins for flow;
// all are mounted and stacked).
type Stack struct{ baseLayout }

// NewStack creates a stacked layout.
func NewStack(children ...controls.Component) *Stack {
	return &Stack{baseLayout: newBaseLayout("stack", children)}
}

// ID sets the stack identifier.
func (s *Stack) ID(id string) *Stack {
	s.id = id
	return s
}

// Class sets extra CSS classes for the stack container.
func (s *Stack) Class(class string) *Stack {
	s.class = class
	return s
}

// Appearance replaces the stack visual appearance.
func (s *Stack) Appearance(value controls.Appearance) *Stack {
	s.appearance = value
	return s
}

// Visible sets stack visibility.
func (s *Stack) Visible(value bool) *Stack {
	s.visible = value
	return s
}

// Grid arranges children in a CSS-style grid.
type Grid struct {
	baseLayout
	columns int
}

// NewGrid creates a grid with the given column count.
//
// Example:
//
//	layout.NewGrid(2, gui.Label("A"), gui.TextBox())
func NewGrid(columns int, children ...controls.Component) *Grid {
	if columns <= 0 {
		columns = 1
	}
	g := &Grid{
		baseLayout: newBaseLayout("grid", children),
		columns:    columns,
	}
	return g
}

// ID sets the grid identifier.
func (g *Grid) ID(id string) *Grid {
	g.id = id
	return g
}

// Class sets extra CSS classes for the grid container.
func (g *Grid) Class(class string) *Grid {
	g.class = class
	return g
}

// Columns sets the number of columns.
func (g *Grid) Columns(n int) *Grid {
	if n > 0 {
		g.columns = n
	}
	return g
}

// Node implements controls.Component.
func (g *Grid) Node() controls.Node {
	n := g.baseLayout.Node()
	n.Props["columns"] = g.columns
	return n
}
