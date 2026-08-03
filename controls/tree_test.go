package controls

import (
	"testing"

	"github.com/PicoBitSyS/PicoGoGUI/binding"
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

func TestLabelNode(t *testing.T) {
	l := NewLabel("Hello").ID("greeting")
	n := l.Node()
	if n.ID != "greeting" || n.Kind != "label" || n.Props["text"] != "Hello" {
		t.Fatalf("node = %+v", n)
	}
}

func TestButtonHandlers(t *testing.T) {
	called := false
	b := NewButton("Go").ID("btn").OnClick(func() { called = true })
	d := events.NewDispatcher()
	d.SetRegistry(CollectAllHandlers(b))
	if !d.Dispatch("btn", "click", nil) || !called {
		t.Fatal("click handler not invoked")
	}
}

func TestTree(t *testing.T) {
	root := Root(NewLabel("A"), NewButton("B"))
	if root.Kind != "column" || len(root.Children) != 2 {
		t.Fatalf("tree = %+v", root)
	}
}

func TestTextBoxBind(t *testing.T) {
	host := binding.New("localhost")
	tb := NewTextBox().ID("host").Bind(host)
	if tb.Node().Props["value"] != "localhost" {
		t.Fatal("initial bind value missing")
	}
	reg := CollectAllHandlers(tb)
	d := events.NewDispatcher()
	d.SetRegistry(reg)
	d.Dispatch("host", "change", "127.0.0.1")
	if host.Get() != "127.0.0.1" {
		t.Fatalf("host = %q", host.Get())
	}
}

func TestCheckBoxBind(t *testing.T) {
	ssl := binding.New(false)
	cb := NewCheckBox("SSL").ID("ssl").Bind(ssl)
	d := events.NewDispatcher()
	d.SetRegistry(CollectAllHandlers(cb))
	d.Dispatch("ssl", "change", true)
	if !ssl.Get() {
		t.Fatal("expected true")
	}
}

func TestTwoTextBoxesShareUIChanges(t *testing.T) {
	value := binding.New("one")
	first := NewTextBox().ID("first").Bind(value)
	second := NewTextBox().ID("second").Bind(value)
	host := &recordingPatcher{}
	first.AttachHost(host)
	second.AttachHost(host)

	d := events.NewDispatcher()
	d.SetRegistry(CollectAllHandlers(first, second))
	d.Dispatch("first", "change", "two")

	if value.Get() != "two" {
		t.Fatalf("binding value = %q", value.Get())
	}
	if len(host.calls) != 1 || host.calls[0].id != "second" || host.calls[0].props["value"] != "two" {
		t.Fatalf("patches = %#v", host.calls)
	}
	host.calls = nil
	first.Value("three")
	if value.Get() != "three" || len(host.calls) != 1 || host.calls[0].id != "second" {
		t.Fatalf("programmatic setter did not propagate: value=%q patches=%#v", value.Get(), host.calls)
	}
}

type patchCall struct {
	id    string
	props map[string]any
}

type recordingPatcher struct {
	calls []patchCall
}

func (p *recordingPatcher) Patch(id string, props map[string]any) error {
	p.calls = append(p.calls, patchCall{id: id, props: props})
	return nil
}
