// Package plugin provides the Phase 6 compile-time plugin system for PicoGoGUI.
//
// Plugins are Go packages that call Register from init (or the host calls Use).
// They contribute custom control kinds, CSS, JS renderers, and designer metadata.
// Native Go plugin DLLs are not supported on Windows; blank-import the package instead.
//
// Example:
//
//	import _ "myapp/plugins/badge"
//
//	func main() {
//	    app := gui.New()
//	    // plugins activate automatically before the window document is built
//	}
package plugin

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// APIVersion is the compile-time extension API compatibility level.
const APIVersion = 1

// Info describes a registered plugin.
type Info struct {
	Name         string
	Version      string
	Description  string
	MinAPI       int
	MaxAPI       int
	Dependencies []string
	Capabilities []string
}

// Plugin is implemented by every PicoGoGUI plugin package.
type Plugin interface {
	Info() Info
	Contribute(h Host) error
}

// Lifecycle is optionally implemented by plugins that own runtime resources.
type Lifecycle interface {
	OnActivate() error
	OnDeactivate() error
}

// Host is the contribution API passed to Plugin.Contribute.
type Host interface {
	RegisterControl(spec ControlSpec) error
	RegisterCSS(id, css string) error
	RegisterJS(id, js string) error
	RegisterDesignerKind(kind DesignerKind) error
}

// ControlSpec registers a custom UI kind rendered by plugin JS.
type ControlSpec struct {
	// Kind is the bridge kind name (e.g. "badge"). Must not collide with built-ins.
	Kind string
	// CSS is optional stylesheet for this control (also registerable via RegisterCSS).
	CSS string
	// CreateJS is a JavaScript function expression:
	//   function(node, props, api) { return HTMLElement; }
	CreateJS string
	// PatchJS is an optional JavaScript function expression:
	//   function(el, props, api) { ... }
	PatchJS string
}

// DesignerKind describes a palette entry for the visual designer.
type DesignerKind struct {
	Kind    string            // bridge kind, e.g. "badge"
	Label   string            // palette button text
	Group   string            // e.g. "Plugins"
	Default map[string]string // optional default props (text, value, …)
	// GoImport is added to generated main when this kind is used (optional).
	GoImport string
	// GoExpr formats Text, Value, ID with fmt — e.g. `badge.New(%[1]q).Tone(%[2]q).ID(%[3]q)`.
	// If empty, codegen falls back to gui.Label.
	GoExpr string
}

var (
	mu             sync.Mutex
	pending        []Plugin
	activated      bool
	activating     bool
	activationDone chan struct{}
	activationErr  error
	activePlugins  []Plugin

	infos         []Info
	cssBlocks     []namedBlock
	jsBlocks      []namedBlock
	controls      map[string]ControlSpec
	designerKinds []DesignerKind
	builtinKinds  = map[string]bool{
		"column": true, "row": true, "stack": true, "grid": true,
		"split": true, "tabs": true, "tab": true, "dock": true, "dockitem": true,
		"canvas": true, "positioned": true,
		"form": true, "field": true,
		"label": true, "button": true, "textbox": true, "numberbox": true,
		"checkbox": true, "combobox": true, "table": true, "tree": true,
		"designsurface": true, "dropzone": true,
	}
)

type namedBlock struct {
	ID   string
	Body string
}

type host struct{}

// Register queues a plugin for activation (typically from init()).
//
// Example:
//
//	func init() { plugin.Register(&MyPlugin{}) }
func Register(p Plugin) {
	_ = RegisterE(p)
}

// RegisterE queues a plugin and reports late or invalid registration.
func RegisterE(p Plugin) error {
	if p == nil {
		return fmt.Errorf("plugin: nil plugin")
	}
	mu.Lock()
	defer mu.Unlock()
	if activated || activating {
		return fmt.Errorf("plugin: Register after activation started is not allowed")
	}
	pending = append(pending, p)
	return nil
}

// Use registers plugins explicitly and activates contributions immediately.
//
// Example:
//
//	_ = plugin.Use(&badge.Plugin{})
func Use(plugins ...Plugin) error {
	mu.Lock()
	if activated || activating {
		mu.Unlock()
		return fmt.Errorf("plugin: Use after activation started is not allowed")
	}
	for _, p := range plugins {
		if p != nil {
			pending = append(pending, p)
		}
	}
	mu.Unlock()
	return Activate()
}

// Activate runs Contribute on every pending plugin exactly once.
// Safe to call repeatedly; subsequent calls are no-ops.
//
// Example:
//
//	_ = plugin.Activate()
func Activate() error {
	mu.Lock()
	if activated {
		mu.Unlock()
		return nil
	}
	if activating {
		done := activationDone
		mu.Unlock()
		<-done
		mu.Lock()
		err := activationErr
		mu.Unlock()
		return err
	}
	activating = true
	activationDone = make(chan struct{})
	list := append([]Plugin(nil), pending...)
	controls = make(map[string]ControlSpec)
	cssBlocks = nil
	jsBlocks = nil
	designerKinds = nil
	infos = nil
	h := host{}
	mu.Unlock()

	ordered, err := orderPlugins(list)
	if err != nil {
		return finishActivation(list, nil, err)
	}
	started := make([]Plugin, 0, len(ordered))
	for _, p := range ordered {
		info := p.Info()
		if err := p.Contribute(h); err != nil {
			deactivateReverse(started)
			return finishActivation(list, nil, fmt.Errorf("plugin %q: %w", info.Name, err))
		}
		if lifecycle, ok := p.(Lifecycle); ok {
			if err := lifecycle.OnActivate(); err != nil {
				deactivateReverse(started)
				return finishActivation(list, nil, fmt.Errorf("plugin %q activate: %w", info.Name, err))
			}
		}
		started = append(started, p)
		mu.Lock()
		infos = append(infos, info)
		mu.Unlock()
	}

	return finishActivation(nil, started, nil)
}

func finishActivation(pendingOnError, active []Plugin, err error) error {
	mu.Lock()
	if err != nil {
		pending = append([]Plugin(nil), pendingOnError...)
		controls = nil
		cssBlocks = nil
		jsBlocks = nil
		designerKinds = nil
		infos = nil
		activePlugins = nil
		activated = false
	} else {
		pending = nil
		activePlugins = append([]Plugin(nil), active...)
		activated = true
	}
	activationErr = err
	activating = false
	done := activationDone
	activationDone = nil
	mu.Unlock()
	close(done)
	return err
}

// Shutdown invokes plugin cleanup in reverse dependency order.
func Shutdown() error {
	mu.Lock()
	active := append([]Plugin(nil), activePlugins...)
	activePlugins = nil
	mu.Unlock()
	var errs []error
	for i := len(active) - 1; i >= 0; i-- {
		if lifecycle, ok := active[i].(Lifecycle); ok {
			if err := lifecycle.OnDeactivate(); err != nil {
				errs = append(errs, fmt.Errorf("plugin %q deactivate: %w", active[i].Info().Name, err))
			}
		}
	}
	return errors.Join(errs...)
}

func deactivateReverse(plugins []Plugin) {
	for i := len(plugins) - 1; i >= 0; i-- {
		if lifecycle, ok := plugins[i].(Lifecycle); ok {
			_ = lifecycle.OnDeactivate()
		}
	}
}

func orderPlugins(list []Plugin) ([]Plugin, error) {
	byName := make(map[string]Plugin, len(list))
	infoByName := make(map[string]Info, len(list))
	for _, candidate := range list {
		info := candidate.Info()
		info.Name = strings.TrimSpace(info.Name)
		if info.Name == "" {
			return nil, fmt.Errorf("plugin: name is required")
		}
		if _, exists := byName[info.Name]; exists {
			return nil, fmt.Errorf("plugin: duplicate name %q", info.Name)
		}
		if info.MinAPI > APIVersion || (info.MaxAPI > 0 && info.MaxAPI < APIVersion) {
			return nil, fmt.Errorf("plugin %q: API %d is outside supported range %d..%d",
				info.Name, APIVersion, info.MinAPI, info.MaxAPI)
		}
		byName[info.Name] = candidate
		infoByName[info.Name] = info
	}

	state := make(map[string]uint8, len(list))
	ordered := make([]Plugin, 0, len(list))
	var visit func(string) error
	visit = func(name string) error {
		switch state[name] {
		case 1:
			return fmt.Errorf("plugin: dependency cycle involving %q", name)
		case 2:
			return nil
		}
		state[name] = 1
		for _, dependency := range infoByName[name].Dependencies {
			dependency = strings.TrimSpace(dependency)
			if _, exists := byName[dependency]; !exists {
				return fmt.Errorf("plugin %q: missing dependency %q", name, dependency)
			}
			if err := visit(dependency); err != nil {
				return err
			}
		}
		state[name] = 2
		ordered = append(ordered, byName[name])
		return nil
	}
	for _, candidate := range list {
		if err := visit(strings.TrimSpace(candidate.Info().Name)); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

// All returns info for activated plugins (empty before Activate).
func All() []Info {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Info, len(infos))
	copy(out, infos)
	return out
}

// DesignerKinds returns palette metadata contributed by plugins.
func DesignerKinds() []DesignerKind {
	mu.Lock()
	defer mu.Unlock()
	out := make([]DesignerKind, len(designerKinds))
	copy(out, designerKinds)
	return out
}

// ControlKinds returns registered custom control kind names (sorted).
func ControlKinds() []string {
	mu.Lock()
	defer mu.Unlock()
	out := make([]string, 0, len(controls))
	for k := range controls {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CSS returns concatenated plugin stylesheets for document injection.
func CSS() string {
	_ = Activate()
	mu.Lock()
	defer mu.Unlock()
	var b strings.Builder
	for _, block := range cssBlocks {
		b.WriteString("\n/* plugin:")
		b.WriteString(block.ID)
		b.WriteString(" */\n")
		b.WriteString(block.Body)
		b.WriteString("\n")
	}
	for _, spec := range controls {
		if strings.TrimSpace(spec.CSS) == "" {
			continue
		}
		b.WriteString("\n/* control:")
		b.WriteString(spec.Kind)
		b.WriteString(" */\n")
		b.WriteString(spec.CSS)
		b.WriteString("\n")
	}
	return b.String()
}

// JS returns concatenated plugin scripts including control create/patch registration.
func JS() string {
	_ = Activate()
	mu.Lock()
	defer mu.Unlock()
	var b strings.Builder
	b.WriteString("\nwindow.__picoPlugins = window.__picoPlugins || { controls: {} };\n")
	b.WriteString("window.__picoPlugins.controls = window.__picoPlugins.controls || {};\n")
	for _, block := range jsBlocks {
		b.WriteString("\n/* plugin-js:")
		b.WriteString(block.ID)
		b.WriteString(" */\n")
		b.WriteString(block.Body)
		b.WriteString("\n")
	}
	// Stable order for tests
	kinds := make([]string, 0, len(controls))
	for k := range controls {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		spec := controls[kind]
		b.WriteString("\n(function(){\n")
		b.WriteString("  var api = window.__picoPlugins;\n")
		b.WriteString("  api.controls[")
		b.WriteString(strconvQuote(kind))
		b.WriteString("] = {\n")
		if strings.TrimSpace(spec.CreateJS) != "" {
			b.WriteString("    create: ")
			b.WriteString(spec.CreateJS)
			b.WriteString(",\n")
		} else {
			b.WriteString("    create: function(node, props) {\n")
			b.WriteString("      var el = document.createElement('div');\n")
			b.WriteString("      el.className = 'pico-plugin ' + (node.kind || '');\n")
			b.WriteString("      el.textContent = (props && props.text) || node.kind || '';\n")
			b.WriteString("      return el;\n")
			b.WriteString("    },\n")
		}
		if strings.TrimSpace(spec.PatchJS) != "" {
			b.WriteString("    patch: ")
			b.WriteString(spec.PatchJS)
			b.WriteString("\n")
		} else {
			b.WriteString("    patch: function(el, props) {\n")
			b.WriteString("      if (props && Object.prototype.hasOwnProperty.call(props, 'text')) {\n")
			b.WriteString("        el.textContent = props.text || '';\n")
			b.WriteString("      }\n")
			b.WriteString("    }\n")
		}
		b.WriteString("  };\n")
		b.WriteString("})();\n")
	}
	return b.String()
}

func strconvQuote(s string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
}

func (host) RegisterControl(spec ControlSpec) error {
	kind := strings.TrimSpace(spec.Kind)
	if kind == "" {
		return fmt.Errorf("control kind is required")
	}
	if builtinKinds[kind] {
		return fmt.Errorf("control kind %q is reserved", kind)
	}
	mu.Lock()
	defer mu.Unlock()
	if _, exists := controls[kind]; exists {
		return fmt.Errorf("control kind %q already registered", kind)
	}
	spec.Kind = kind
	controls[kind] = spec
	return nil
}

func (host) RegisterCSS(id, css string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("css id is required")
	}
	mu.Lock()
	defer mu.Unlock()
	for _, b := range cssBlocks {
		if b.ID == id {
			return fmt.Errorf("css id %q already registered", id)
		}
	}
	cssBlocks = append(cssBlocks, namedBlock{ID: id, Body: css})
	return nil
}

func (host) RegisterJS(id, js string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("js id is required")
	}
	mu.Lock()
	defer mu.Unlock()
	for _, b := range jsBlocks {
		if b.ID == id {
			return fmt.Errorf("js id %q already registered", id)
		}
	}
	jsBlocks = append(jsBlocks, namedBlock{ID: id, Body: js})
	return nil
}

func (host) RegisterDesignerKind(kind DesignerKind) error {
	kind.Kind = strings.TrimSpace(kind.Kind)
	if kind.Kind == "" {
		return fmt.Errorf("designer kind is required")
	}
	if kind.Label == "" {
		kind.Label = kind.Kind
	}
	if kind.Group == "" {
		kind.Group = "Plugins"
	}
	mu.Lock()
	defer mu.Unlock()
	for _, existing := range designerKinds {
		if existing.Kind == kind.Kind {
			return fmt.Errorf("designer kind %q already registered", kind.Kind)
		}
	}
	designerKinds = append(designerKinds, kind)
	return nil
}

// ResetForTest clears registry state (tests only).
func ResetForTest() {
	mu.Lock()
	defer mu.Unlock()
	pending = nil
	activated = false
	activating = false
	activationDone = nil
	activationErr = nil
	activePlugins = nil
	infos = nil
	cssBlocks = nil
	jsBlocks = nil
	controls = nil
	designerKinds = nil
}
