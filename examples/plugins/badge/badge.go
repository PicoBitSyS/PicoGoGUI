// Package badge is a sample PicoGoGUI plugin that provides a Badge control.
package badge

import (
	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/events"
	"github.com/PicoBitSyS/PicoGoGUI/plugin"
)

func init() {
	plugin.Register(&Plugin{})
}

// Plugin contributes the Badge control to PicoGoGUI.
type Plugin struct{}

// Info implements plugin.Plugin.
func (Plugin) Info() plugin.Info {
	return plugin.Info{
		Name:        "badge",
		Version:     "1.0.0",
		Description: "Compact status badge control",
	}
}

// Contribute implements plugin.Plugin.
func (Plugin) Contribute(h plugin.Host) error {
	if err := h.RegisterControl(plugin.ControlSpec{
		Kind: "badge",
		CSS: `
.pico-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 22px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  line-height: 18px;
  border: 1px solid transparent;
  user-select: none;
}
.pico-badge-info {
  background: color-mix(in srgb, var(--pico-accent) 18%, transparent);
  color: var(--pico-accent);
  border-color: color-mix(in srgb, var(--pico-accent) 35%, transparent);
}
.pico-badge-success {
  background: color-mix(in srgb, #3ecf8e 22%, transparent);
  color: #1f8f5f;
  border-color: color-mix(in srgb, #3ecf8e 40%, transparent);
}
.pico-badge-warn {
  background: color-mix(in srgb, #f0ad4e 22%, transparent);
  color: #9a6b12;
  border-color: color-mix(in srgb, #f0ad4e 40%, transparent);
}
[data-theme="dark"] .pico-badge-success { color: #6ee7b0; }
[data-theme="dark"] .pico-badge-warn { color: #ffd28a; }
`,
		CreateJS: `function(node, props) {
  var el = document.createElement("span");
  var tone = (props && props.tone) || "info";
  el.className = "pico-badge pico-badge-" + tone;
  el.textContent = (props && props.text) || "";
  return el;
}`,
		PatchJS: `function(el, props) {
  if (!props) return;
  if (Object.prototype.hasOwnProperty.call(props, "text")) {
    el.textContent = props.text || "";
  }
  if (Object.prototype.hasOwnProperty.call(props, "tone")) {
    var tone = props.tone || "info";
    el.className = "pico-badge pico-badge-" + tone;
  }
}`,
	}); err != nil {
		return err
	}
	return h.RegisterDesignerKind(plugin.DesignerKind{
		Kind:     "badge",
		Label:    "Badge",
		Group:    "Plugins",
		Default:  map[string]string{"text": "Badge", "value": "info"},
		GoImport: "github.com/PicoBitSyS/PicoGoGUI/examples/plugins/badge",
		GoExpr:   `badge.New(%[1]q).Tone(%[2]q).ID(%[3]q)`,
	})
}

// Tone values for Badge.
const (
	ToneInfo    = "info"
	ToneSuccess = "success"
	ToneWarn    = "warn"
)

// Badge is a compact status label contributed by this plugin.
type Badge struct {
	id      string
	text    string
	tone    string
	visible bool
	enabled bool
	patcher controls.Patcher
}

// New creates a Badge with the given text.
//
// Example:
//
//	badge.New("LIVE").Tone(badge.ToneSuccess)
func New(text string) *Badge {
	return &Badge{
		id:      controls.AllocateID("badge"),
		text:    text,
		tone:    ToneInfo,
		visible: true,
		enabled: true,
	}
}

// ID sets the component identifier.
func (b *Badge) ID(id string) *Badge {
	b.id = id
	return b
}

// Text sets the badge caption.
func (b *Badge) Text(text string) *Badge {
	b.text = text
	return b
}

// Tone sets the visual tone (info, success, warn).
func (b *Badge) Tone(tone string) *Badge {
	if tone == "" {
		tone = ToneInfo
	}
	b.tone = tone
	return b
}

// Visible sets visibility.
func (b *Badge) Visible(v bool) *Badge {
	b.visible = v
	return b
}

// Enabled sets enabled state.
func (b *Badge) Enabled(v bool) *Badge {
	b.enabled = v
	return b
}

// AttachHost implements controls.HostAware.
func (b *Badge) AttachHost(p controls.Patcher) { b.patcher = p }

// CompID implements controls.Component.
func (b *Badge) CompID() string { return b.id }

// Kind implements controls.Component.
func (b *Badge) Kind() string { return "badge" }

func (b *Badge) props() map[string]any {
	return map[string]any{
		"text":    b.text,
		"tone":    b.tone,
		"visible": b.visible,
		"enabled": b.enabled,
	}
}

// Node implements controls.Component.
func (b *Badge) Node() controls.Node {
	return controls.Node{ID: b.id, Kind: b.Kind(), Props: b.props()}
}

// CollectHandlers implements controls.Component.
func (b *Badge) CollectHandlers(*events.Registry) {}

// Refresh patches the mounted badge.
func (b *Badge) Refresh() {
	if b.patcher != nil {
		_ = b.patcher.Patch(b.id, b.props())
	}
}
