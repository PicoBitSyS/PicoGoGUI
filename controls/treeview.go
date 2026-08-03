package controls

import (
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

// TreeNode is a hierarchical tree entry.
type TreeNode struct {
	ID       string
	Text     string
	Expanded bool
	Children []*TreeNode
}

// NewTreeNode creates a tree node with generated id.
//
// Example:
//
//	gui.TreeNode("Servers", gui.TreeNode("smtp"))
func NewTreeNode(text string, children ...*TreeNode) *TreeNode {
	return &TreeNode{
		ID:       AllocateID("treenode"),
		Text:     text,
		Expanded: true,
		Children: children,
	}
}

// WithID sets the node id.
func (n *TreeNode) WithID(id string) *TreeNode {
	n.ID = id
	return n
}

// Tree displays a hierarchical list.
type Tree struct {
	base
	nodes    []*TreeNode
	onSelect func(string)
	onToggle func(id string, expanded bool)
	patcher  Patcher
	selected string
}

// NewTree creates an empty tree.
func NewTree() *Tree {
	return &Tree{base: newBase("tree")}
}

// ID sets the component identifier.
func (t *Tree) ID(id string) *Tree {
	t.id = id
	return t
}

// Nodes sets the root nodes.
func (t *Tree) Nodes(nodes ...*TreeNode) *Tree {
	t.nodes = nodes
	return t
}

// Visible sets visibility.
func (t *Tree) Visible(v bool) *Tree {
	t.visible = v
	return t
}

// OnSelect registers a selection handler (node id).
func (t *Tree) OnSelect(fn func(string)) *Tree {
	t.onSelect = fn
	return t
}

// OnToggle registers an expand/collapse handler.
func (t *Tree) OnToggle(fn func(id string, expanded bool)) *Tree {
	t.onToggle = fn
	return t
}

// Selected sets the selected node id.
func (t *Tree) Selected(id string) *Tree {
	t.selected = id
	return t
}

// SelectedID returns the selected node id.
func (t *Tree) SelectedID() string { return t.selected }

// AttachHost implements HostAware.
func (t *Tree) AttachHost(p Patcher) { t.patcher = p }

// Kind implements Component.
func (t *Tree) Kind() string { return "tree" }

func (t *Tree) props() map[string]any {
	props := map[string]any{
		"nodes":    serializeTreeNodes(t.nodes),
		"selected": t.selected,
	}
	t.applyCommonProps(props)
	return props
}

// Node implements Component.
func (t *Tree) Node() Node {
	return Node{ID: t.id, Kind: t.Kind(), Props: t.props()}
}

// CollectHandlers implements Component.
func (t *Tree) CollectHandlers(reg *events.Registry) {
	reg.OnSelect(t.id, func(v any) {
		id, _ := v.(string)
		t.selected = id
		if t.patcher != nil {
			_ = t.patcher.Patch(t.id, map[string]any{"nodes": serializeTreeNodes(t.nodes), "selected": id})
		}
		if t.onSelect != nil {
			t.onSelect(id)
		}
	})
	reg.On(t.id, "toggle", func(v any) {
		m, _ := v.(map[string]any)
		if m == nil {
			return
		}
		id, _ := m["id"].(string)
		expanded, _ := m["expanded"].(bool)
		setExpanded(t.nodes, id, expanded)
		if t.onToggle != nil {
			t.onToggle(id, expanded)
		}
		if t.patcher != nil {
			_ = t.patcher.Patch(t.id, t.props())
		}
	})
}

func serializeTreeNodes(nodes []*TreeNode) []map[string]any {
	out := make([]map[string]any, 0, len(nodes))
	for _, n := range nodes {
		if n == nil {
			continue
		}
		item := map[string]any{
			"id":       n.ID,
			"text":     n.Text,
			"expanded": n.Expanded,
		}
		if len(n.Children) > 0 {
			item["children"] = serializeTreeNodes(n.Children)
		}
		out = append(out, item)
	}
	return out
}

func setExpanded(nodes []*TreeNode, id string, expanded bool) bool {
	for _, n := range nodes {
		if n == nil {
			continue
		}
		if n.ID == id {
			n.Expanded = expanded
			return true
		}
		if setExpanded(n.Children, id, expanded) {
			return true
		}
	}
	return false
}
