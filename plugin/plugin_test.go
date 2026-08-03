package plugin_test

import (
	"strings"
	"testing"

	"github.com/PicoBitSyS/PicoGoGUI/plugin"
)

type samplePlugin struct {
	name string
	fail bool
}

func (p *samplePlugin) Info() plugin.Info {
	return plugin.Info{Name: p.name, Version: "1.0.0", Description: "test"}
}

func (p *samplePlugin) Contribute(h plugin.Host) error {
	if p.fail {
		return errFail
	}
	if err := h.RegisterControl(plugin.ControlSpec{
		Kind: "samplebadge",
		CSS:  ".pico-samplebadge{color:red}",
		CreateJS: `function(node, props) {
			var el = document.createElement("div");
			el.className = "pico-samplebadge";
			el.textContent = (props && props.text) || "";
			return el;
		}`,
		PatchJS: `function(el, props) {
			if (props && Object.prototype.hasOwnProperty.call(props, "text")) {
				el.textContent = props.text || "";
			}
		}`,
	}); err != nil {
		return err
	}
	if err := h.RegisterCSS("sample-extra", ".x{display:block}"); err != nil {
		return err
	}
	if err := h.RegisterDesignerKind(plugin.DesignerKind{
		Kind:    "samplebadge",
		Label:   "Sample Badge",
		Group:   "Plugins",
		Default: map[string]string{"text": "Badge"},
	}); err != nil {
		return err
	}
	return nil
}

type errString string

func (e errString) Error() string { return string(e) }

const errFail = errString("contribute failed")

func TestActivateAndAssets(t *testing.T) {
	plugin.ResetForTest()
	plugin.Register(&samplePlugin{name: "sample"})
	if err := plugin.Activate(); err != nil {
		t.Fatal(err)
	}
	infos := plugin.All()
	if len(infos) != 1 || infos[0].Name != "sample" {
		t.Fatalf("All = %#v", infos)
	}
	css := plugin.CSS()
	if !strings.Contains(css, ".pico-samplebadge") || !strings.Contains(css, ".x{display:block}") {
		t.Fatalf("CSS missing pieces:\n%s", css)
	}
	js := plugin.JS()
	if !strings.Contains(js, `__picoPlugins`) || !strings.Contains(js, `controls["samplebadge"]`) {
		t.Fatalf("JS missing registration:\n%s", js)
	}
	kinds := plugin.ControlKinds()
	if len(kinds) != 1 || kinds[0] != "samplebadge" {
		t.Fatalf("ControlKinds = %#v", kinds)
	}
	dk := plugin.DesignerKinds()
	if len(dk) != 1 || dk[0].Label != "Sample Badge" {
		t.Fatalf("DesignerKinds = %#v", dk)
	}
}

func TestDuplicateKind(t *testing.T) {
	plugin.ResetForTest()
	plugin.Register(&samplePlugin{name: "a"})
	plugin.Register(&dupPlugin{})
	if err := plugin.Activate(); err == nil {
		t.Fatal("expected duplicate kind error")
	}
}

type dupPlugin struct{}

func (dupPlugin) Info() plugin.Info { return plugin.Info{Name: "dup"} }
func (dupPlugin) Contribute(h plugin.Host) error {
	return h.RegisterControl(plugin.ControlSpec{Kind: "samplebadge", CreateJS: `function(){return document.createElement("div")}`})
}

func TestReservedKind(t *testing.T) {
	plugin.ResetForTest()
	plugin.Register(&reservedPlugin{})
	if err := plugin.Activate(); err == nil {
		t.Fatal("expected reserved kind error")
	}
}

type reservedPlugin struct{}

func (reservedPlugin) Info() plugin.Info { return plugin.Info{Name: "bad"} }
func (reservedPlugin) Contribute(h plugin.Host) error {
	return h.RegisterControl(plugin.ControlSpec{Kind: "button"})
}

func TestUse(t *testing.T) {
	plugin.ResetForTest()
	if err := plugin.Use(&samplePlugin{name: "via-use"}); err != nil {
		t.Fatal(err)
	}
	if len(plugin.All()) != 1 {
		t.Fatalf("All = %#v", plugin.All())
	}
}

type lifecyclePlugin struct {
	name  string
	deps  []string
	order *[]string
}

func (p *lifecyclePlugin) Info() plugin.Info {
	return plugin.Info{Name: p.name, MinAPI: 1, MaxAPI: 1, Dependencies: p.deps}
}
func (p *lifecyclePlugin) Contribute(plugin.Host) error {
	*p.order = append(*p.order, "contribute:"+p.name)
	return nil
}
func (p *lifecyclePlugin) OnActivate() error {
	*p.order = append(*p.order, "activate:"+p.name)
	return nil
}
func (p *lifecyclePlugin) OnDeactivate() error {
	*p.order = append(*p.order, "deactivate:"+p.name)
	return nil
}

func TestDependenciesAndLifecycle(t *testing.T) {
	plugin.ResetForTest()
	var order []string
	plugin.Register(&lifecyclePlugin{name: "feature", deps: []string{"core"}, order: &order})
	plugin.Register(&lifecyclePlugin{name: "core", order: &order})
	if err := plugin.Activate(); err != nil {
		t.Fatal(err)
	}
	if err := plugin.Shutdown(); err != nil {
		t.Fatal(err)
	}
	got := strings.Join(order, ",")
	want := "contribute:core,activate:core,contribute:feature,activate:feature,deactivate:feature,deactivate:core"
	if got != want {
		t.Fatalf("lifecycle order = %q want %q", got, want)
	}
}

func TestMissingDependency(t *testing.T) {
	plugin.ResetForTest()
	var order []string
	plugin.Register(&lifecyclePlugin{name: "feature", deps: []string{"missing"}, order: &order})
	if err := plugin.Activate(); err == nil {
		t.Fatal("expected missing dependency error")
	}
}
