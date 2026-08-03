package controls

import "github.com/PicoBitSyS/PicoGoGUI/events"

// Root builds a root column node from a list of components.
func Root(children ...Component) Node {
	nodes := make([]Node, 0, len(children))
	for _, c := range children {
		if c == nil {
			continue
		}
		nodes = append(nodes, c.Node())
	}
	return Node{
		ID:       "root",
		Kind:     "column",
		Props:    map[string]any{"visible": true, "enabled": true},
		Children: nodes,
	}
}

// CollectAllHandlers gathers event handlers from every component into a registry.
func CollectAllHandlers(children ...Component) *events.Registry {
	reg := events.NewRegistry()
	for _, c := range children {
		if c == nil {
			continue
		}
		c.CollectHandlers(reg)
	}
	return reg
}
