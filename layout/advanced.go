package layout

import (
	"strconv"

	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

// Split arranges two panes horizontally or vertically.
type Split struct {
	id       string
	first    controls.Component
	second   controls.Component
	vertical bool
	ratio    int
}

// NewSplit creates a horizontal two-pane split.
func NewSplit(first, second controls.Component) *Split {
	return &Split{id: controls.AllocateID("split"), first: first, second: second, ratio: 50}
}

func (s *Split) ID(id string) *Split { s.id = id; return s }

// Vertical changes the split direction.
func (s *Split) Vertical(value bool) *Split { s.vertical = value; return s }

// Ratio sets the first pane percentage from 10 through 90.
func (s *Split) Ratio(percent int) *Split {
	if percent < 10 {
		percent = 10
	}
	if percent > 90 {
		percent = 90
	}
	s.ratio = percent
	return s
}

func (s *Split) CompID() string { return s.id }
func (s *Split) Kind() string   { return "split" }
func (s *Split) ChildComponents() []controls.Component {
	return []controls.Component{s.first, s.second}
}
func (s *Split) Node() controls.Node {
	children := make([]controls.Node, 0, 2)
	for _, child := range s.ChildComponents() {
		if child != nil {
			children = append(children, child.Node())
		}
	}
	return controls.Node{
		ID: s.id, Kind: s.Kind(),
		Props:    map[string]any{"vertical": s.vertical, "ratio": s.ratio, "visible": true, "enabled": true},
		Children: children,
	}
}
func (s *Split) CollectHandlers(reg *events.Registry) {
	for _, child := range s.ChildComponents() {
		if child != nil {
			child.CollectHandlers(reg)
		}
	}
}

// Tab is one titled page in a Tabs layout.
type Tab struct {
	title string
	child controls.Component
}

// NewTab creates a titled tab page.
func NewTab(title string, child controls.Component) *Tab {
	return &Tab{title: title, child: child}
}

// Tabs displays one selected page at a time.
type Tabs struct {
	id       string
	pages    []*Tab
	selected int
	onChange func(int)
	patcher  controls.Patcher
}

// NewTabs creates a tab layout.
func NewTabs(pages ...*Tab) *Tabs {
	return &Tabs{id: controls.AllocateID("tabs"), pages: pages}
}

func (t *Tabs) ID(id string) *Tabs { t.id = id; return t }

// Selected sets the active tab index.
func (t *Tabs) Selected(index int) *Tabs {
	if index >= 0 && index < len(t.pages) {
		t.selected = index
		if t.patcher != nil {
			_ = t.patcher.Patch(t.id, map[string]any{"selected": index})
		}
	}
	return t
}

// OnChange registers an active-tab handler.
func (t *Tabs) OnChange(fn func(int)) *Tabs { t.onChange = fn; return t }

func (t *Tabs) AttachHost(p controls.Patcher) { t.patcher = p }
func (t *Tabs) CompID() string                { return t.id }
func (t *Tabs) Kind() string                  { return "tabs" }
func (t *Tabs) ChildComponents() []controls.Component {
	out := make([]controls.Component, 0, len(t.pages))
	for _, page := range t.pages {
		if page != nil && page.child != nil {
			out = append(out, page.child)
		}
	}
	return out
}
func (t *Tabs) Node() controls.Node {
	children := make([]controls.Node, 0, len(t.pages))
	for index, page := range t.pages {
		if page == nil {
			continue
		}
		var content []controls.Node
		if page.child != nil {
			content = []controls.Node{page.child.Node()}
		}
		children = append(children, controls.Node{
			ID:       t.id + "-page-" + strconv.Itoa(index),
			Kind:     "tab",
			Props:    map[string]any{"title": page.title, "index": index, "visible": true, "enabled": true},
			Children: content,
		})
	}
	return controls.Node{
		ID: t.id, Kind: t.Kind(),
		Props:    map[string]any{"selected": t.selected, "visible": true, "enabled": true},
		Children: children,
	}
}
func (t *Tabs) CollectHandlers(reg *events.Registry) {
	reg.OnChange(t.id, func(value any) {
		index := -1
		switch raw := value.(type) {
		case float64:
			index = int(raw)
		case int:
			index = raw
		}
		if index < 0 || index >= len(t.pages) {
			return
		}
		t.selected = index
		if t.onChange != nil {
			t.onChange(index)
		}
	})
	for _, child := range t.ChildComponents() {
		child.CollectHandlers(reg)
	}
}

// DockRegion identifies a Dock area.
type DockRegion string

const (
	DockTop    DockRegion = "top"
	DockRight  DockRegion = "right"
	DockBottom DockRegion = "bottom"
	DockLeft   DockRegion = "left"
	DockCenter DockRegion = "center"
)

// DockItem assigns a component to a Dock region.
type DockItem struct {
	Region DockRegion
	Child  controls.Component
}

// Dock arranges components around a center region.
type Dock struct {
	id    string
	items []DockItem
}

// NewDock creates a dock layout.
func NewDock(items ...DockItem) *Dock {
	return &Dock{id: controls.AllocateID("dock"), items: items}
}

func (d *Dock) ID(id string) *Dock { d.id = id; return d }
func (d *Dock) CompID() string     { return d.id }
func (d *Dock) Kind() string       { return "dock" }
func (d *Dock) ChildComponents() []controls.Component {
	out := make([]controls.Component, 0, len(d.items))
	for _, item := range d.items {
		if item.Child != nil {
			out = append(out, item.Child)
		}
	}
	return out
}
func (d *Dock) Node() controls.Node {
	children := make([]controls.Node, 0, len(d.items))
	for index, item := range d.items {
		if item.Child == nil {
			continue
		}
		children = append(children, controls.Node{
			ID: d.id + "-item-" + strconv.Itoa(index), Kind: "dockitem",
			Props:    map[string]any{"region": item.Region, "visible": true, "enabled": true},
			Children: []controls.Node{item.Child.Node()},
		})
	}
	return controls.Node{ID: d.id, Kind: d.Kind(), Props: map[string]any{"visible": true, "enabled": true}, Children: children}
}
func (d *Dock) CollectHandlers(reg *events.Registry) {
	for _, child := range d.ChildComponents() {
		child.CollectHandlers(reg)
	}
}
