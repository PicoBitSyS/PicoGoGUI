package layout

import (
	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

// Canvas is an absolute-positioning surface for designer-generated layouts.
type Canvas struct {
	id       string
	children []controls.Component
}

// NewCanvas creates an absolute-positioning surface.
func NewCanvas(children ...controls.Component) *Canvas {
	return &Canvas{id: controls.AllocateID("canvas"), children: children}
}

func (c *Canvas) ID(id string) *Canvas { c.id = id; return c }
func (c *Canvas) CompID() string       { return c.id }
func (c *Canvas) Kind() string         { return "canvas" }
func (c *Canvas) ChildComponents() []controls.Component {
	return c.children
}
func (c *Canvas) Node() controls.Node {
	children := make([]controls.Node, 0, len(c.children))
	for _, child := range c.children {
		if child != nil {
			children = append(children, child.Node())
		}
	}
	return controls.Node{
		ID: c.id, Kind: c.Kind(),
		Props:    map[string]any{"visible": true, "enabled": true},
		Children: children,
	}
}
func (c *Canvas) CollectHandlers(reg *events.Registry) {
	for _, child := range c.children {
		if child != nil {
			child.CollectHandlers(reg)
		}
	}
}

// Positioned places one component at explicit canvas coordinates.
type Positioned struct {
	id                  string
	child               controls.Component
	x, y, width, height int
	zIndex              int
}

// ZIndex sets the stacking order inside the canvas.
func (p *Positioned) ZIndex(value int) *Positioned {
	p.zIndex = value
	return p
}

// At places child at x/y with width/height in CSS pixels.
func At(child controls.Component, x, y, width, height int) *Positioned {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	id := controls.AllocateID("positioned")
	if child != nil {
		id = child.CompID() + "-position"
	}
	return &Positioned{id: id, child: child, x: x, y: y, width: width, height: height}
}

func (p *Positioned) CompID() string { return p.id }
func (p *Positioned) Kind() string   { return "positioned" }
func (p *Positioned) ChildComponents() []controls.Component {
	if p.child == nil {
		return nil
	}
	return []controls.Component{p.child}
}
func (p *Positioned) Node() controls.Node {
	var children []controls.Node
	if p.child != nil {
		children = []controls.Node{p.child.Node()}
	}
	return controls.Node{
		ID: p.id, Kind: p.Kind(),
		Props: map[string]any{
			"x": p.x, "y": p.y, "width": p.width, "height": p.height,
			"zIndex":  p.zIndex,
			"visible": true, "enabled": true,
		},
		Children: children,
	}
}
func (p *Positioned) CollectHandlers(reg *events.Registry) {
	if p.child != nil {
		p.child.CollectHandlers(reg)
	}
}
