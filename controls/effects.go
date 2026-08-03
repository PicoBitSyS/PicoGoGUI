package controls

import (
	"encoding/base64"
	"strings"

	"github.com/PicoBitSyS/PicoGoGUI/events"
)

// AnimationName identifies a built-in motion effect.
type AnimationName string

const (
	AnimationFadeIn  AnimationName = "fade-in"
	AnimationSlideUp AnimationName = "slide-up"
	AnimationScaleIn AnimationName = "scale-in"
	AnimationPulse   AnimationName = "pulse"
)

// Animation configures a built-in component animation.
type Animation struct {
	Name       AnimationName
	DurationMS int
	DelayMS    int
	Iterations int
}

// Animated applies motion metadata to an existing component.
type Animated struct {
	child     Component
	animation Animation
}

// Animate wraps a component with a built-in animation.
func Animate(child Component, animation Animation) *Animated {
	if animation.DurationMS <= 0 {
		animation.DurationMS = 180
	}
	if animation.Iterations <= 0 {
		animation.Iterations = 1
	}
	return &Animated{child: child, animation: animation}
}

func (a *Animated) CompID() string {
	if a == nil || a.child == nil {
		return ""
	}
	return a.child.CompID()
}

func (a *Animated) Kind() string {
	if a == nil || a.child == nil {
		return ""
	}
	return a.child.Kind()
}

func (a *Animated) Node() Node {
	if a == nil || a.child == nil {
		return Node{}
	}
	node := a.child.Node()
	if node.Props == nil {
		node.Props = make(map[string]any)
	}
	node.Props["animation"] = map[string]any{
		"name":       a.animation.Name,
		"durationMS": a.animation.DurationMS,
		"delayMS":    a.animation.DelayMS,
		"iterations": a.animation.Iterations,
	}
	return node
}

func (a *Animated) CollectHandlers(reg *events.Registry) {
	if a != nil && a.child != nil {
		a.child.CollectHandlers(reg)
	}
}

func (a *Animated) ChildComponents() []Component {
	if a == nil || a.child == nil {
		return nil
	}
	return []Component{a.child}
}

// DroppedFile contains browser-safe file metadata and optional content.
type DroppedFile struct {
	Name         string
	Size         int64
	Type         string
	LastModified int64
	Data         []byte
	Truncated    bool
}

// DropZone accepts files dragged from Windows Explorer.
type DropZone struct {
	base
	child    Component
	prompt   string
	maxBytes int
	onDrop   func([]DroppedFile)
}

// NewDropZone creates a file drop target around child.
func NewDropZone(child Component) *DropZone {
	return &DropZone{
		base:     newBase("dropzone"),
		child:    child,
		prompt:   "Drop files here",
		maxBytes: 4 << 20,
	}
}

func (d *DropZone) ID(id string) *DropZone {
	d.id = id
	return d
}

// Prompt sets accessible drop instructions.
func (d *DropZone) Prompt(text string) *DropZone {
	d.prompt = text
	return d
}

// MaxBytes limits content transferred per file. Zero transfers metadata only.
func (d *DropZone) MaxBytes(n int) *DropZone {
	if n >= 0 {
		d.maxBytes = n
	}
	return d
}

// OnDrop registers the file drop handler.
func (d *DropZone) OnDrop(fn func([]DroppedFile)) *DropZone {
	d.onDrop = fn
	return d
}

func (d *DropZone) Kind() string { return "dropzone" }

func (d *DropZone) ChildComponents() []Component {
	if d.child == nil {
		return nil
	}
	return []Component{d.child}
}

func (d *DropZone) Node() Node {
	props := map[string]any{"prompt": d.prompt, "maxBytes": d.maxBytes}
	d.applyCommonProps(props)
	node := Node{ID: d.id, Kind: d.Kind(), Props: props}
	if d.child != nil {
		node.Children = []Node{d.child.Node()}
	}
	return node
}

func (d *DropZone) CollectHandlers(reg *events.Registry) {
	reg.On(d.id, "drop", func(value any) {
		if d.onDrop == nil {
			return
		}
		list, _ := value.([]any)
		files := make([]DroppedFile, 0, len(list))
		for _, raw := range list {
			item, _ := raw.(map[string]any)
			if item == nil {
				continue
			}
			file := DroppedFile{
				Name:         stringValue(item["name"]),
				Type:         stringValue(item["type"]),
				Size:         int64(asFloat(item["size"])),
				LastModified: int64(asFloat(item["lastModified"])),
				Truncated:    boolValue(item["truncated"]),
			}
			if encoded := stringValue(item["data"]); encoded != "" {
				if comma := strings.IndexByte(encoded, ','); comma >= 0 {
					encoded = encoded[comma+1:]
				}
				file.Data, _ = base64.StdEncoding.DecodeString(encoded)
			}
			files = append(files, file)
		}
		d.onDrop(files)
	})
	if d.child != nil {
		d.child.CollectHandlers(reg)
	}
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func boolValue(value any) bool {
	result, _ := value.(bool)
	return result
}

func asFloat(value any) float64 {
	result, _ := value.(float64)
	return result
}
