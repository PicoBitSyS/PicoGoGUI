package controls

import "github.com/PicoBitSyS/PicoGoGUI/events"

// LayoutHandler receives geometry updates from the design surface.
type LayoutHandler func(index, x, y, width, height int)

// DesignLayoutChange is one geometry update emitted by the design surface.
type DesignLayoutChange struct {
	Index               int
	X, Y, Width, Height int
}

// LayoutsHandler receives an atomic group move.
type LayoutsHandler func([]DesignLayoutChange)

// DesignSurface is a WinForms-style form preview used by the visual designer.
type DesignSurface struct {
	base
	title       string
	width       int
	height      int
	widgets     []map[string]any
	selected    int
	selection   []int
	onSelect    func(int)
	onSelection func([]int)
	onLayout    LayoutHandler
	onLayouts   LayoutsHandler
	patcher     Patcher
}

// NewDesignSurface creates an empty design surface.
func NewDesignSurface() *DesignSurface {
	return &DesignSurface{
		base:     newBase("designsurface"),
		title:    "MyWindow",
		width:    480,
		height:   360,
		selected: -1,
	}
}

// ID sets the component identifier.
func (d *DesignSurface) ID(id string) *DesignSurface {
	d.id = id
	return d
}

// Title sets the preview window title.
func (d *DesignSurface) Title(title string) *DesignSurface {
	d.title = title
	return d
}

// FormSize sets the preview client size in CSS pixels.
func (d *DesignSurface) FormSize(width, height int) *DesignSurface {
	if width > 0 {
		d.width = width
	}
	if height > 0 {
		d.height = height
	}
	return d
}

// Widgets replaces the preview widget list.
func (d *DesignSurface) Widgets(items []map[string]any) *DesignSurface {
	d.widgets = items
	return d
}

// Selected sets the selected widget index (-1 = none).
func (d *DesignSurface) Selected(i int) *DesignSurface {
	d.selected = i
	if i < 0 {
		d.selection = nil
	} else {
		d.selection = []int{i}
	}
	return d
}

// Selection sets all selected widget indices. The last index is primary.
func (d *DesignSurface) Selection(indices ...int) *DesignSurface {
	d.selection = append([]int(nil), indices...)
	d.selected = -1
	if len(indices) > 0 {
		d.selected = indices[len(indices)-1]
	}
	return d
}

// OnSelect registers a widget-select handler (0-based index, -1 clears).
func (d *DesignSurface) OnSelect(fn func(int)) *DesignSurface {
	d.onSelect = fn
	return d
}

// OnSelection registers a multi-selection handler. The last index is primary.
func (d *DesignSurface) OnSelection(fn func([]int)) *DesignSurface {
	d.onSelection = fn
	return d
}

// OnLayout registers a geometry-change handler (after drag/resize).
func (d *DesignSurface) OnLayout(fn LayoutHandler) *DesignSurface {
	d.onLayout = fn
	return d
}

// OnLayouts registers an atomic group geometry-change handler.
func (d *DesignSurface) OnLayouts(fn LayoutsHandler) *DesignSurface {
	d.onLayouts = fn
	return d
}

// Visible sets visibility.
func (d *DesignSurface) Visible(v bool) *DesignSurface {
	d.visible = v
	return d
}

// Enabled sets enabled state.
func (d *DesignSurface) Enabled(v bool) *DesignSurface {
	d.enabled = v
	return d
}

// Refresh patches the mounted surface.
func (d *DesignSurface) Refresh() {
	if d.patcher != nil {
		_ = d.patcher.Patch(d.id, d.props())
	}
}

// AttachHost implements HostAware.
func (d *DesignSurface) AttachHost(p Patcher) { d.patcher = p }

// Kind implements Component.
func (d *DesignSurface) Kind() string { return "designsurface" }

func (d *DesignSurface) props() map[string]any {
	widgets := d.widgets
	if widgets == nil {
		widgets = []map[string]any{}
	}
	props := map[string]any{
		"title":     d.title,
		"width":     d.width,
		"height":    d.height,
		"widgets":   widgets,
		"selected":  d.selected,
		"selection": append([]int(nil), d.selection...),
	}
	d.applyCommonProps(props)
	return props
}

// Node implements Component.
func (d *DesignSurface) Node() Node {
	return Node{ID: d.id, Kind: d.Kind(), Props: d.props()}
}

// CollectHandlers implements Component.
func (d *DesignSurface) CollectHandlers(reg *events.Registry) {
	reg.OnSelect(d.id, func(v any) {
		idx := asInt(v)
		d.selected = idx
		if idx < 0 {
			d.selection = nil
		} else {
			d.selection = []int{idx}
		}
		if d.onSelect != nil {
			d.onSelect(idx)
		}
	})
	reg.On(d.id, "selection", func(v any) {
		indices := asIntSlice(v)
		d.selection = indices
		d.selected = -1
		if len(indices) > 0 {
			d.selected = indices[len(indices)-1]
		}
		if d.onSelection != nil {
			d.onSelection(append([]int(nil), indices...))
		}
		if d.onSelect != nil {
			d.onSelect(d.selected)
		}
	})
	reg.On(d.id, "layout", func(v any) {
		idx, x, y, w, h, ok := parseLayoutValue(v)
		if !ok {
			return
		}
		d.selected = idx
		if d.onLayout != nil {
			d.onLayout(idx, x, y, w, h)
		}
	})
	reg.On(d.id, "layouts", func(v any) {
		changes := parseLayoutValues(v)
		if len(changes) == 0 {
			return
		}
		if d.onLayouts != nil {
			d.onLayouts(changes)
		}
	})
}

func asIntSlice(v any) []int {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]int, 0, len(raw))
	seen := map[int]bool{}
	for _, item := range raw {
		index := asInt(item)
		if index < 0 || seen[index] {
			continue
		}
		seen[index] = true
		out = append(out, index)
	}
	return out
}

func parseLayoutValues(v any) []DesignLayoutChange {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]DesignLayoutChange, 0, len(raw))
	for _, item := range raw {
		index, x, y, width, height, valid := parseLayoutValue(item)
		if valid {
			out = append(out, DesignLayoutChange{
				Index: index, X: x, Y: y, Width: width, Height: height,
			})
		}
	}
	return out
}

func parseLayoutValue(v any) (index, x, y, width, height int, ok bool) {
	m, isMap := v.(map[string]any)
	if !isMap {
		return 0, 0, 0, 0, 0, false
	}
	index = asInt(m["index"])
	x = asInt(m["x"])
	y = asInt(m["y"])
	width = asInt(m["width"])
	height = asInt(m["height"])
	if width < 1 || height < 1 {
		return 0, 0, 0, 0, 0, false
	}
	return index, x, y, width, height, true
}
