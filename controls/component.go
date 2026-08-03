// Package controls defines the declarative UI component model for PicoGoGUI.
package controls

import (
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/PicoBitSyS/PicoGoGUI/events"
)

var idSeq atomic.Uint64

// AllocateID returns a unique component identifier.
func AllocateID(prefix string) string {
	n := idSeq.Add(1)
	return prefix + "-" + strconv.FormatUint(n, 10)
}

func nextID(prefix string) string { return AllocateID(prefix) }

// Patcher applies property updates to a mounted component.
type Patcher interface {
	Patch(id string, props map[string]any) error
}

// HostAware is implemented by components that need the window patch bus
// (e.g. for two-way binding).
type HostAware interface {
	AttachHost(p Patcher)
}

// Disposable is implemented by components that own subscriptions or other
// resources which must be released when the component leaves the UI tree.
type Disposable interface {
	Dispose()
}

// Container is implemented by layout components that own children.
type Container interface {
	Component
	ChildComponents() []Component
}

// Component is the common interface implemented by every UI control.
type Component interface {
	// CompID returns the unique identifier of this component.
	CompID() string
	// Kind returns the control type name used by the web runtime (e.g. "label").
	Kind() string
	// Node serializes this component (and children) into a bridge-friendly tree node.
	Node() Node
	// CollectHandlers registers event handlers into the registry.
	CollectHandlers(reg *events.Registry)
}

// Node is the JSON-serializable representation of a component in the UI tree.
type Node struct {
	ID       string         `json:"id"`
	Kind     string         `json:"kind"`
	Props    map[string]any `json:"props,omitempty"`
	Children []Node         `json:"children,omitempty"`
}

// base holds fields shared by all controls.
type base struct {
	id         string
	class      string
	visible    bool
	enabled    bool
	appearance Appearance
}

func newBase(prefix string) base {
	return base{
		id:      nextID(prefix),
		visible: true,
		enabled: true,
	}
}

func (b *base) CompID() string { return b.id }

func (b *base) applyCommonProps(props map[string]any) {
	props["visible"] = b.visible
	props["enabled"] = b.enabled
	props["appearance"] = b.appearance
	if b.class != "" {
		props["class"] = b.class
	}
}

func (b *base) setClass(value string) {
	b.class = strings.Join(strings.Fields(value), " ")
}

// AttachHosts walks the component tree and attaches the patcher to HostAware nodes.
func AttachHosts(p Patcher, children ...Component) {
	for _, c := range children {
		attachHost(p, c)
	}
}

func attachHost(p Patcher, c Component) {
	if c == nil {
		return
	}
	if h, ok := c.(HostAware); ok {
		h.AttachHost(p)
	}
	if cont, ok := c.(Container); ok {
		for _, ch := range cont.ChildComponents() {
			attachHost(p, ch)
		}
	}
}

// DisposeAll walks component trees and releases resources owned by every
// Disposable component.
func DisposeAll(children ...Component) {
	for _, c := range children {
		dispose(c)
	}
}

func dispose(c Component) {
	if c == nil {
		return
	}
	if cont, ok := c.(Container); ok {
		for _, ch := range cont.ChildComponents() {
			dispose(ch)
		}
	}
	if d, ok := c.(Disposable); ok {
		d.Dispose()
	}
}
