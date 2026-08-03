package designer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/plugin"
)

// Supported widget kinds.
const (
	KindLabel     = "label"
	KindButton    = "button"
	KindTextBox   = "textbox"
	KindNumberBox = "numberbox"
	KindCheckBox  = "checkbox"
	KindComboBox  = "combobox"
	KindColumn    = "column"
	KindRow       = "row"
	KindStack     = "stack"
)

// Widget is one control on the designer canvas.
// X/Y are relative to the parent container (or form root). Width/Height are CSS pixels.
type Widget struct {
	Kind       string              `json:"kind"`
	ID         string              `json:"id,omitempty"`
	Text       string              `json:"text,omitempty"`
	Value      string              `json:"value,omitempty"`
	Class      string              `json:"class,omitempty"`
	Parent     string              `json:"parent,omitempty"`
	X          int                 `json:"x,omitempty"`
	Y          int                 `json:"y,omitempty"`
	Width      int                 `json:"width,omitempty"`
	Height     int                 `json:"height,omitempty"`
	ZIndex     int                 `json:"zIndex,omitempty"`
	Appearance controls.Appearance `json:"appearance,omitempty"`
	Locked     bool                `json:"locked,omitempty"`
	Hidden     bool                `json:"hidden,omitempty"`
}

// GeometryChange is one atomic geometry update, used by group moves.
type GeometryChange struct {
	Index               int
	X, Y, Width, Height int
}

// AlignMode identifies an alignment command.
type AlignMode string

const (
	AlignLeft    AlignMode = "left"
	AlignHCenter AlignMode = "hcenter"
	AlignRight   AlignMode = "right"
	AlignTop     AlignMode = "top"
	AlignVCenter AlignMode = "vcenter"
	AlignBottom  AlignMode = "bottom"
)

// DistributeMode identifies an equal-spacing command.
type DistributeMode string

const (
	DistributeHorizontal DistributeMode = "horizontal"
	DistributeVertical   DistributeMode = "vertical"
)

// Document is a saved designer layout.
type Document struct {
	WindowTitle string   `json:"windowTitle"`
	Width       int      `json:"width,omitempty"`
	Height      int      `json:"height,omitempty"`
	Widgets     []Widget `json:"widgets"`
	history     []documentSnapshot
	future      []documentSnapshot
}

type documentSnapshot struct {
	WindowTitle string
	Width       int
	Height      int
	Widgets     []Widget
}

const designerZLimit = 99999

// IsContainer reports whether kind is a layout container.
func IsContainer(kind string) bool {
	switch kind {
	case KindColumn, KindRow, KindStack:
		return true
	default:
		return false
	}
}

// NewDocument creates an empty document.
func NewDocument(title string) *Document {
	if title == "" {
		title = "MyWindow"
	}
	return &Document{WindowTitle: title, Width: 480, Height: 360, Widgets: nil}
}

// Add appends a widget with defaults for id, text, size, and placement.
func (d *Document) Add(w Widget) {
	d.pushHistory()
	w.Appearance = normalizeAppearance(w.Appearance)
	w.Class = normalizeClass(w.Class)
	w.ZIndex = normalizeZIndex(w.ZIndex)
	if w.ID == "" {
		w.ID = d.nextID(w.Kind)
	} else if d.hasID(w.ID) {
		w.ID = d.nextID(w.Kind)
	}
	if w.Text == "" {
		if def := pluginKindDefault(w.Kind); def["text"] != "" {
			w.Text = def["text"]
		} else {
			switch w.Kind {
			case KindLabel:
				w.Text = "Label"
			case KindButton:
				w.Text = "Button"
			case KindCheckBox:
				w.Text = "CheckBox"
			case KindComboBox:
				w.Text = "ComboBox"
			case KindColumn:
				w.Text = "Column"
			case KindRow:
				w.Text = "Row"
			case KindStack:
				w.Text = "Stack"
			}
		}
	}
	if w.Value == "" {
		if def := pluginKindDefault(w.Kind); def["value"] != "" {
			w.Value = def["value"]
		}
	}
	if w.Kind == KindComboBox && w.Value == "" {
		w.Value = "Item1,Item2,Item3"
	}
	if w.Width <= 0 || w.Height <= 0 {
		dw, dh := defaultSize(w.Kind)
		if w.Width <= 0 {
			w.Width = dw
		}
		if w.Height <= 0 {
			w.Height = dh
		}
	}
	if w.X == 0 && w.Y == 0 {
		w.X, w.Y = d.nextPlacement(w.Parent)
	}
	if w.Parent != "" && !d.hasID(w.Parent) {
		w.Parent = ""
	}
	if w.Parent != "" && w.Parent == w.ID {
		w.Parent = ""
	}
	if w.ZIndex == 0 {
		w.ZIndex = d.nextZIndex(w.Parent, IsContainer(w.Kind))
	}
	d.Widgets = append(d.Widgets, w)
}

// AddE validates and adds a widget without silently repairing invalid input.
func (d *Document) AddE(w Widget) error {
	if strings.TrimSpace(w.Kind) == "" {
		return fmt.Errorf("designer: widget kind is required")
	}
	if w.ID != "" && d.hasID(w.ID) {
		return fmt.Errorf("designer: duplicate widget id %q", w.ID)
	}
	if w.Parent != "" {
		parent := d.IndexOfID(w.Parent)
		if parent < 0 {
			return fmt.Errorf("designer: parent %q does not exist", w.Parent)
		}
		if !IsContainer(d.Widgets[parent].Kind) {
			return fmt.Errorf("designer: parent %q is not a container", w.Parent)
		}
	}
	d.Add(w)
	return nil
}

func defaultSize(kind string) (w, h int) {
	switch kind {
	case KindLabel:
		return 80, 22
	case KindButton:
		return 100, 32
	case KindTextBox, KindNumberBox, KindComboBox:
		return 160, 28
	case KindCheckBox:
		return 120, 24
	case "badge":
		return 72, 22
	case KindColumn, KindRow, KindStack:
		return 220, 160
	default:
		return 100, 32
	}
}

func pluginKindDefault(kind string) map[string]string {
	_ = plugin.Activate()
	for _, dk := range plugin.DesignerKinds() {
		if dk.Kind == kind && dk.Default != nil {
			return dk.Default
		}
	}
	return nil
}

func (d *Document) nextPlacement(parent string) (x, y int) {
	n := 0
	for _, w := range d.Widgets {
		if w.Parent == parent {
			n++
		}
	}
	return 12 + (n%5)*12, 12 + n*24
}

func (d *Document) nextZIndex(parent string, container bool) int {
	next := 0
	for _, w := range d.Widgets {
		if w.Parent == parent && IsContainer(w.Kind) == container && w.ZIndex >= next {
			next = w.ZIndex + 1
		}
	}
	return normalizeZIndex(next)
}

func normalizeZIndex(value int) int {
	if value < -designerZLimit {
		return -designerZLimit
	}
	if value > designerZLimit {
		return designerZLimit
	}
	return value
}

func normalizeAppearance(value controls.Appearance) controls.Appearance {
	value.FontFamily = strings.TrimSpace(value.FontFamily)
	value.Color = strings.TrimSpace(value.Color)
	value.Background = strings.TrimSpace(value.Background)
	value.BorderColor = strings.TrimSpace(value.BorderColor)
	value.TextAlign = strings.ToLower(strings.TrimSpace(value.TextAlign))
	switch value.TextAlign {
	case "", "left", "center", "right", "justify":
	default:
		value.TextAlign = "left"
	}
	if value.FontSize < 0 {
		value.FontSize = 0
	}
	if value.FontSize > 512 {
		value.FontSize = 512
	}
	if value.BorderWidth < 0 {
		value.BorderWidth = 0
	}
	if value.BorderRadius < 0 {
		value.BorderRadius = 0
	}
	if value.Opacity < 0 {
		value.Opacity = 0
	}
	if value.Opacity > 1 {
		value.Opacity = 1
	}
	return value
}

func normalizeClass(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func effectiveZIndex(w Widget) int {
	z := normalizeZIndex(w.ZIndex)
	if IsContainer(w.Kind) {
		return z
	}
	return designerZLimit*2 + z + 1
}

func (d *Document) hasID(id string) bool {
	for _, w := range d.Widgets {
		if w.ID == id {
			return true
		}
	}
	return false
}

func (d *Document) nextID(kind string) string {
	n := 1
	for {
		id := fmt.Sprintf("%s%d", kind, n)
		if !d.hasID(id) {
			return id
		}
		n++
	}
}

// IndexOfID returns the widget index or -1.
func (d *Document) IndexOfID(id string) int {
	for i, w := range d.Widgets {
		if w.ID == id {
			return i
		}
	}
	return -1
}

// ChildrenOf returns widgets whose Parent matches parentID, sorted by visual layer.
// Containers are always below leaf controls, regardless of their individual ZIndex.
func (d *Document) ChildrenOf(parentID string) []Widget {
	out := make([]Widget, 0)
	for _, w := range d.Widgets {
		if w.Parent == parentID {
			out = append(out, w)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		zi, zj := effectiveZIndex(out[i]), effectiveZIndex(out[j])
		if zi != zj {
			return zi < zj
		}
		if out[i].Y != out[j].Y {
			return out[i].Y < out[j].Y
		}
		return out[i].X < out[j].X
	})
	return out
}

// RemoveAt deletes a widget and all descendants.
func (d *Document) RemoveAt(i int) {
	d.RemoveIndices([]int{i})
}

// RemoveIndices deletes selected widgets and their descendants as one undo step.
func (d *Document) RemoveIndices(indices []int) bool {
	remove := map[string]bool{}
	for _, i := range uniqueValidIndices(indices, len(d.Widgets)) {
		remove[d.Widgets[i].ID] = true
	}
	if len(remove) == 0 {
		return false
	}
	changed := true
	for changed {
		changed = false
		for _, w := range d.Widgets {
			if remove[w.ID] {
				continue
			}
			if w.Parent != "" && remove[w.Parent] {
				remove[w.ID] = true
				changed = true
			}
		}
	}
	d.pushHistory()
	keep := make([]Widget, 0, len(d.Widgets))
	for _, w := range d.Widgets {
		if !remove[w.ID] {
			keep = append(keep, w)
		}
	}
	d.Widgets = keep
	return true
}

// UpdateAt replaces a widget at index.
func (d *Document) UpdateAt(i int, w Widget) {
	_ = d.UpdateAtE(i, w)
}

// UpdateAtE validates and replaces a widget at index.
func (d *Document) UpdateAtE(i int, w Widget) error {
	if i < 0 || i >= len(d.Widgets) {
		return fmt.Errorf("designer: widget index %d is out of range", i)
	}
	if w.ID == "" {
		w.ID = d.Widgets[i].ID
	}
	if duplicate := d.IndexOfID(w.ID); duplicate >= 0 && duplicate != i {
		return fmt.Errorf("designer: duplicate widget id %q", w.ID)
	}
	if w.Parent == w.ID {
		return fmt.Errorf("designer: widget %q cannot parent itself", w.ID)
	}
	if w.Parent != "" && !d.hasID(w.Parent) {
		return fmt.Errorf("designer: parent %q does not exist", w.Parent)
	}
	if w.Parent != "" {
		parent := d.IndexOfID(w.Parent)
		if parent >= 0 && !IsContainer(d.Widgets[parent].Kind) {
			return fmt.Errorf("designer: parent %q is not a container", w.Parent)
		}
	}
	// Prevent cycles: parent cannot be a descendant.
	if w.Parent != "" && d.isDescendant(w.ID, w.Parent) {
		return fmt.Errorf("designer: parent %q creates a cycle", w.Parent)
	}
	w.Appearance = normalizeAppearance(w.Appearance)
	w.Class = normalizeClass(w.Class)
	w.ZIndex = normalizeZIndex(w.ZIndex)
	d.pushHistory()
	d.Widgets[i] = w
	return nil
}

// BringToFront moves a widget above siblings of the same layer.
// Containers remain below non-container controls by design.
func (d *Document) BringToFront(i int) bool {
	if i < 0 || i >= len(d.Widgets) {
		return false
	}
	current := d.Widgets[i]
	maxZ := current.ZIndex
	maxCount := 0
	for _, w := range d.Widgets {
		if w.Parent != current.Parent || IsContainer(w.Kind) != IsContainer(current.Kind) {
			continue
		}
		if w.ZIndex > maxZ {
			maxZ = w.ZIndex
			maxCount = 1
		} else if w.ZIndex == maxZ {
			maxCount++
		}
	}
	if current.ZIndex == maxZ && maxCount == 1 {
		return false
	}
	next := normalizeZIndex(maxZ + 1)
	if next == current.ZIndex {
		return false
	}
	d.pushHistory()
	current.ZIndex = next
	d.Widgets[i] = current
	return true
}

// SendToBack moves a widget below siblings of the same layer.
// A container can never cover a non-container control.
func (d *Document) SendToBack(i int) bool {
	if i < 0 || i >= len(d.Widgets) {
		return false
	}
	current := d.Widgets[i]
	minZ := current.ZIndex
	minCount := 0
	for _, w := range d.Widgets {
		if w.Parent != current.Parent || IsContainer(w.Kind) != IsContainer(current.Kind) {
			continue
		}
		if w.ZIndex < minZ {
			minZ = w.ZIndex
			minCount = 1
		} else if w.ZIndex == minZ {
			minCount++
		}
	}
	if current.ZIndex == minZ && minCount == 1 {
		return false
	}
	next := normalizeZIndex(minZ - 1)
	if next == current.ZIndex {
		return false
	}
	d.pushHistory()
	current.ZIndex = next
	d.Widgets[i] = current
	return true
}

func (d *Document) isDescendant(ancestorID, maybeChildID string) bool {
	cur := maybeChildID
	for cur != "" {
		if cur == ancestorID {
			return true
		}
		idx := d.IndexOfID(cur)
		if idx < 0 {
			return false
		}
		cur = d.Widgets[idx].Parent
	}
	return false
}

// SetGeometry updates X/Y/Width/Height for the widget at index.
func (d *Document) SetGeometry(i, x, y, width, height int) {
	d.SetGeometries([]GeometryChange{{Index: i, X: x, Y: y, Width: width, Height: height}})
}

// SetGeometries applies a group move/resize as one undo step.
func (d *Document) SetGeometries(changes []GeometryChange) bool {
	next := append([]Widget(nil), d.Widgets...)
	changed := false
	for _, change := range changes {
		if change.Index < 0 || change.Index >= len(next) || next[change.Index].Locked {
			continue
		}
		width, height := change.Width, change.Height
		if width < 16 {
			width = 16
		}
		if height < 16 {
			height = 16
		}
		w := next[change.Index]
		if w.X == change.X && w.Y == change.Y && w.Width == width && w.Height == height {
			continue
		}
		w.X, w.Y, w.Width, w.Height = change.X, change.Y, width, height
		next[change.Index] = w
		changed = true
	}
	if !changed {
		return false
	}
	d.pushHistory()
	d.Widgets = next
	return true
}

func uniqueValidIndices(indices []int, size int) []int {
	seen := make(map[int]bool, len(indices))
	out := make([]int, 0, len(indices))
	for _, index := range indices {
		if index < 0 || index >= size || seen[index] {
			continue
		}
		seen[index] = true
		out = append(out, index)
	}
	return out
}

// SetLocked locks or unlocks widgets as one undo step.
func (d *Document) SetLocked(indices []int, value bool) bool {
	valid := uniqueValidIndices(indices, len(d.Widgets))
	changed := false
	for _, index := range valid {
		if d.Widgets[index].Locked != value {
			changed = true
			break
		}
	}
	if !changed {
		return false
	}
	d.pushHistory()
	for _, index := range valid {
		d.Widgets[index].Locked = value
	}
	return true
}

// SetHidden shows or hides widgets as one undo step.
func (d *Document) SetHidden(indices []int, value bool) bool {
	valid := uniqueValidIndices(indices, len(d.Widgets))
	changed := false
	for _, index := range valid {
		if d.Widgets[index].Hidden != value {
			changed = true
			break
		}
	}
	if !changed {
		return false
	}
	d.pushHistory()
	for _, index := range valid {
		d.Widgets[index].Hidden = value
	}
	return true
}

// Align aligns unlocked widgets in each parent coordinate space.
func (d *Document) Align(indices []int, mode AlignMode) bool {
	groups := d.movableGroups(indices, 2)
	next := append([]Widget(nil), d.Widgets...)
	changed := false
	for _, group := range groups {
		minX, minY := next[group[0]].X, next[group[0]].Y
		maxX := next[group[0]].X + next[group[0]].Width
		maxY := next[group[0]].Y + next[group[0]].Height
		for _, index := range group[1:] {
			w := next[index]
			minX = min(minX, w.X)
			minY = min(minY, w.Y)
			maxX = max(maxX, w.X+w.Width)
			maxY = max(maxY, w.Y+w.Height)
		}
		for _, index := range group {
			w := next[index]
			x, y := w.X, w.Y
			switch mode {
			case AlignLeft:
				x = minX
			case AlignHCenter:
				x = minX + (maxX-minX-w.Width)/2
			case AlignRight:
				x = maxX - w.Width
			case AlignTop:
				y = minY
			case AlignVCenter:
				y = minY + (maxY-minY-w.Height)/2
			case AlignBottom:
				y = maxY - w.Height
			default:
				continue
			}
			if x != w.X || y != w.Y {
				w.X, w.Y = x, y
				next[index] = w
				changed = true
			}
		}
	}
	if !changed {
		return false
	}
	d.pushHistory()
	d.Widgets = next
	return true
}

// Distribute places unlocked widgets at equal gaps in each parent coordinate space.
func (d *Document) Distribute(indices []int, mode DistributeMode) bool {
	groups := d.movableGroups(indices, 3)
	next := append([]Widget(nil), d.Widgets...)
	changed := false
	for _, group := range groups {
		sort.SliceStable(group, func(i, j int) bool {
			a, b := next[group[i]], next[group[j]]
			if mode == DistributeVertical {
				return a.Y < b.Y
			}
			return a.X < b.X
		})
		first, last := next[group[0]], next[group[len(group)-1]]
		totalSize := 0
		for _, index := range group {
			if mode == DistributeVertical {
				totalSize += next[index].Height
			} else {
				totalSize += next[index].Width
			}
		}
		span := last.X + last.Width - first.X
		cursor := float64(first.X)
		if mode == DistributeVertical {
			span = last.Y + last.Height - first.Y
			cursor = float64(first.Y)
		}
		gap := float64(span-totalSize) / float64(len(group)-1)
		for position, index := range group {
			if position == 0 || position == len(group)-1 {
				if mode == DistributeVertical {
					cursor += float64(next[index].Height) + gap
				} else {
					cursor += float64(next[index].Width) + gap
				}
				continue
			}
			w := next[index]
			if mode == DistributeVertical {
				y := int(math.Round(cursor))
				if y != w.Y {
					w.Y = y
					changed = true
				}
				cursor += float64(w.Height) + gap
			} else {
				x := int(math.Round(cursor))
				if x != w.X {
					w.X = x
					changed = true
				}
				cursor += float64(w.Width) + gap
			}
			next[index] = w
		}
	}
	if !changed {
		return false
	}
	d.pushHistory()
	d.Widgets = next
	return true
}

func (d *Document) movableGroups(indices []int, minimum int) map[string][]int {
	groups := map[string][]int{}
	for _, index := range uniqueValidIndices(indices, len(d.Widgets)) {
		w := d.Widgets[index]
		if w.Locked || w.Hidden {
			continue
		}
		groups[w.Parent] = append(groups[w.Parent], index)
	}
	for parent, group := range groups {
		if len(group) < minimum {
			delete(groups, parent)
		}
	}
	return groups
}

// MarshalJSON returns indented JSON.
func (d *Document) MarshalJSON() ([]byte, error) {
	type alias Document
	return json.MarshalIndent((*alias)(d), "", "  ")
}

// ParseDocument loads a document from JSON.
func ParseDocument(data []byte) (*Document, error) {
	var d Document
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	if d.WindowTitle == "" {
		d.WindowTitle = "MyWindow"
	}
	if d.Width <= 0 {
		d.Width = 480
	}
	if d.Height <= 0 {
		d.Height = 360
	}
	for i := range d.Widgets {
		w := &d.Widgets[i]
		w.Appearance = normalizeAppearance(w.Appearance)
		w.Class = normalizeClass(w.Class)
		w.ZIndex = normalizeZIndex(w.ZIndex)
		if w.Width <= 0 || w.Height <= 0 {
			dw, dh := defaultSize(w.Kind)
			if w.Width <= 0 {
				w.Width = dw
			}
			if w.Height <= 0 {
				w.Height = dh
			}
		}
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return &d, nil
}

// Validate checks ids, parents, container relationships, and cycles.
func (d *Document) Validate() error {
	ids := make(map[string]int, len(d.Widgets))
	for index, widget := range d.Widgets {
		if strings.TrimSpace(widget.Kind) == "" {
			return fmt.Errorf("designer: widget %d has no kind", index)
		}
		if strings.TrimSpace(widget.ID) == "" {
			return fmt.Errorf("designer: widget %d has no id", index)
		}
		if previous, exists := ids[widget.ID]; exists {
			return fmt.Errorf("designer: duplicate id %q at widgets %d and %d", widget.ID, previous, index)
		}
		ids[widget.ID] = index
	}
	for _, widget := range d.Widgets {
		if widget.Parent == "" {
			continue
		}
		parentIndex, exists := ids[widget.Parent]
		if !exists {
			return fmt.Errorf("designer: parent %q for %q does not exist", widget.Parent, widget.ID)
		}
		if !IsContainer(d.Widgets[parentIndex].Kind) {
			return fmt.Errorf("designer: parent %q for %q is not a container", widget.Parent, widget.ID)
		}
		seen := map[string]bool{widget.ID: true}
		parent := widget.Parent
		for parent != "" {
			if seen[parent] {
				return fmt.Errorf("designer: parent cycle involving %q", widget.ID)
			}
			seen[parent] = true
			parent = d.Widgets[ids[parent]].Parent
		}
	}
	return nil
}

// CanUndo reports whether Undo can restore an earlier document state.
func (d *Document) CanUndo() bool { return len(d.history) > 0 }

// CanRedo reports whether Redo can restore a reverted document state.
func (d *Document) CanRedo() bool { return len(d.future) > 0 }

// Undo restores the previous document state.
func (d *Document) Undo() bool {
	if len(d.history) == 0 {
		return false
	}
	d.future = append(d.future, d.snapshot())
	last := len(d.history) - 1
	d.restore(d.history[last])
	d.history = d.history[:last]
	return true
}

// Redo reapplies the most recently undone state.
func (d *Document) Redo() bool {
	if len(d.future) == 0 {
		return false
	}
	d.history = append(d.history, d.snapshot())
	last := len(d.future) - 1
	d.restore(d.future[last])
	d.future = d.future[:last]
	return true
}

func (d *Document) pushHistory() {
	d.history = append(d.history, d.snapshot())
	if len(d.history) > 100 {
		d.history = append([]documentSnapshot(nil), d.history[len(d.history)-100:]...)
	}
	d.future = nil
}

func (d *Document) snapshot() documentSnapshot {
	return documentSnapshot{
		WindowTitle: d.WindowTitle,
		Width:       d.Width,
		Height:      d.Height,
		Widgets:     append([]Widget(nil), d.Widgets...),
	}
}

func (d *Document) restore(snapshot documentSnapshot) {
	d.WindowTitle = snapshot.WindowTitle
	d.Width = snapshot.Width
	d.Height = snapshot.Height
	d.Widgets = append([]Widget(nil), snapshot.Widgets...)
}

// SurfaceWidgets returns maps for DesignSurface preview.
func (d *Document) SurfaceWidgets() []map[string]any {
	out := make([]map[string]any, 0, len(d.Widgets))
	for index, w := range d.Widgets {
		out = append(out, map[string]any{
			"kind":            w.Kind,
			"id":              w.ID,
			"text":            w.Text,
			"value":           w.Value,
			"class":           w.Class,
			"parent":          w.Parent,
			"x":               w.X,
			"y":               w.Y,
			"width":           w.Width,
			"height":          w.Height,
			"zIndex":          w.ZIndex,
			"effectiveZIndex": effectiveZIndex(w),
			"appearance":      w.Appearance,
			"index":           index,
			"locked":          w.Locked,
			"hidden":          w.Hidden,
		})
	}
	return out
}

// Rows returns table-friendly maps for the designer canvas list.
func (d *Document) Rows() []map[string]any {
	out := make([]map[string]any, 0, len(d.Widgets))
	for _, w := range d.Widgets {
		summary := w.Text
		if summary == "" {
			summary = w.Value
		}
		out = append(out, map[string]any{
			"Kind": w.Kind,
			"ID":   w.ID,
			"Text": summary,
		})
	}
	return out
}

// GenerateGo emits a runnable main package using the PicoGoGUI facade.
func (d *Document) GenerateGo() string {
	_ = plugin.Activate()
	imports := d.collectImports()

	var b bytes.Buffer
	b.WriteString("package main\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"log\"\n\n")
	b.WriteString("\tgui \"github.com/PicoBitSyS/PicoGoGUI\"\n")
	for _, imp := range imports {
		b.WriteString("\t")
		b.WriteString(strconv.Quote(imp))
		b.WriteString("\n")
	}
	b.WriteString(")\n\n")
	b.WriteString("func main() {\n")
	b.WriteString("\tapp := gui.New(gui.Options{Theme: gui.ThemeSystem()})\n")
	b.WriteString(fmt.Sprintf("\twin := app.NewWindow(%s)\n", strconv.Quote(d.WindowTitle)))
	w, h := d.Width, d.Height
	if w <= 0 {
		w = 480
	}
	if h <= 0 {
		h = 360
	}
	b.WriteString(fmt.Sprintf("\twin.SetSize(%d, %d)\n", w, h))
	roots := d.ChildrenOf("")
	b.WriteString("\twin.Add(\n")
	b.WriteString("\t\tgui.Canvas(\n")
	if len(roots) == 0 {
		b.WriteString("\t\t\tgui.At(gui.Label(\"Empty designer canvas\"), 12, 12, 180, 24),\n")
	} else {
		for _, root := range roots {
			b.WriteString(emitTree(d, root, "\t\t\t"))
			b.WriteString(",\n")
		}
	}
	b.WriteString("\t\t),\n")
	b.WriteString("\t)\n")
	b.WriteString("\tif err := app.Run(); err != nil {\n")
	b.WriteString("\t\tlog.Fatal(err)\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	if formatted, err := format.Source(b.Bytes()); err == nil {
		return string(formatted)
	}
	return b.String()
}

func (d *Document) collectImports() []string {
	used := map[string]bool{}
	for _, w := range d.Widgets {
		for _, dk := range plugin.DesignerKinds() {
			if dk.Kind == w.Kind && dk.GoImport != "" {
				used[dk.GoImport] = true
			}
		}
	}
	out := make([]string, 0, len(used))
	for imp := range used {
		out = append(out, imp)
	}
	sort.Strings(out)
	return out
}

func emitTree(d *Document, w Widget, indent string) string {
	var b strings.Builder
	b.WriteString(indent)
	b.WriteString("gui.At(\n")
	b.WriteString(emitCore(d, w, indent+"\t"))
	b.WriteString(",\n")
	b.WriteString(fmt.Sprintf("%s\t%d, %d, %d, %d,\n", indent, w.X, w.Y, w.Width, w.Height))
	b.WriteString(indent)
	b.WriteString(fmt.Sprintf(").ZIndex(%d)", effectiveZIndex(w)))
	return b.String()
}

func emitCore(d *Document, w Widget, indent string) string {
	if !IsContainer(w.Kind) {
		return indent + widgetExpr(w)
	}
	kids := d.ChildrenOf(w.ID)
	var b strings.Builder
	ctor := "gui.Column"
	switch w.Kind {
	case KindRow:
		ctor = "gui.Row"
	case KindStack:
		ctor = "gui.Stack"
	}
	b.WriteString(indent)
	b.WriteString(ctor)
	b.WriteString("(\n")
	if len(kids) == 0 {
		b.WriteString(indent + "\tgui.Label(" + strconv.Quote(w.Text) + "),\n")
	} else {
		for _, kid := range kids {
			b.WriteString(emitTree(d, kid, indent+"\t"))
			b.WriteString(",\n")
		}
	}
	b.WriteString(indent + ")")
	id := strings.TrimSpace(w.ID)
	if id != "" {
		b.WriteString(fmt.Sprintf(".ID(%s)", strconv.Quote(id)))
	}
	if appearance := appearanceExpr(w.Appearance); appearance != "" {
		b.WriteString(".Appearance(" + appearance + ")")
	}
	if w.Class != "" {
		b.WriteString(fmt.Sprintf(".Class(%s)", strconv.Quote(w.Class)))
	}
	if w.Hidden {
		b.WriteString(".Visible(false)")
	}
	return b.String()
}

func widgetExpr(w Widget) string {
	if expr, ok := pluginWidgetExpr(w); ok {
		return expr
	}
	id := strings.TrimSpace(w.ID)
	switch w.Kind {
	case KindLabel:
		s := fmt.Sprintf("gui.Label(%s)", strconv.Quote(w.Text))
		if id != "" {
			s += fmt.Sprintf(".ID(%s)", strconv.Quote(id))
		}
		return withWidgetState(s, w)
	case KindButton:
		s := fmt.Sprintf("gui.Button(%s)", strconv.Quote(w.Text))
		if id != "" {
			s += fmt.Sprintf(".ID(%s)", strconv.Quote(id))
		}
		return withWidgetState(s, w)
	case KindTextBox:
		s := "gui.TextBox()"
		if id != "" {
			s += fmt.Sprintf(".ID(%s)", strconv.Quote(id))
		}
		if w.Value != "" {
			s += fmt.Sprintf(".Value(%s)", strconv.Quote(w.Value))
		}
		return withWidgetState(s, w)
	case KindNumberBox:
		s := "gui.NumberBox()"
		if id != "" {
			s += fmt.Sprintf(".ID(%s)", strconv.Quote(id))
		}
		if w.Value != "" {
			if n, err := strconv.ParseFloat(w.Value, 64); err == nil {
				s += fmt.Sprintf(".Value(%v)", n)
			}
		}
		return withWidgetState(s, w)
	case KindCheckBox:
		s := fmt.Sprintf("gui.CheckBox(%s)", strconv.Quote(w.Text))
		if id != "" {
			s += fmt.Sprintf(".ID(%s)", strconv.Quote(id))
		}
		return withWidgetState(s, w)
	case KindComboBox:
		items := splitCSV(w.Value)
		if len(items) == 0 {
			items = []string{"Item1", "Item2"}
		}
		args := make([]string, 0, len(items))
		for _, it := range items {
			args = append(args, strconv.Quote(it))
		}
		s := "gui.ComboBox(" + strings.Join(args, ", ") + ")"
		if id != "" {
			s += fmt.Sprintf(".ID(%s)", strconv.Quote(id))
		}
		return withWidgetState(s, w)
	default:
		return fmt.Sprintf("gui.Label(%s)", strconv.Quote("unknown:"+w.Kind))
	}
}

func withWidgetState(expr string, widget Widget) string {
	if appearance := appearanceExpr(widget.Appearance); appearance != "" {
		expr += ".Appearance(" + appearance + ")"
	}
	if widget.Class != "" {
		expr += ".Class(" + strconv.Quote(widget.Class) + ")"
	}
	if widget.Hidden {
		expr += ".Visible(false)"
	}
	return expr
}

func appearanceExpr(value controls.Appearance) string {
	value = normalizeAppearance(value)
	fields := make([]string, 0, 12)
	addString := func(name, v string) {
		if v != "" {
			fields = append(fields, name+": "+strconv.Quote(v))
		}
	}
	addInt := func(name string, v int) {
		if v != 0 {
			fields = append(fields, fmt.Sprintf("%s: %d", name, v))
		}
	}
	addString("FontFamily", value.FontFamily)
	addInt("FontSize", value.FontSize)
	addString("Color", value.Color)
	addString("Background", value.Background)
	if value.Bold {
		fields = append(fields, "Bold: true")
	}
	if value.Italic {
		fields = append(fields, "Italic: true")
	}
	if value.Underline {
		fields = append(fields, "Underline: true")
	}
	addString("TextAlign", value.TextAlign)
	addString("BorderColor", value.BorderColor)
	addInt("BorderWidth", value.BorderWidth)
	addInt("BorderRadius", value.BorderRadius)
	if value.Opacity != 0 {
		fields = append(fields, fmt.Sprintf("Opacity: %s", strconv.FormatFloat(value.Opacity, 'f', -1, 64)))
	}
	if len(fields) == 0 {
		return ""
	}
	return "gui.Appearance{" + strings.Join(fields, ", ") + "}"
}

func pluginWidgetExpr(w Widget) (string, bool) {
	for _, dk := range plugin.DesignerKinds() {
		if dk.Kind != w.Kind || strings.TrimSpace(dk.GoExpr) == "" {
			continue
		}
		text := w.Text
		value := w.Value
		id := w.ID
		if value == "" {
			value = "info"
		}
		return fmt.Sprintf(dk.GoExpr, text, value, id), true
	}
	return "", false
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
