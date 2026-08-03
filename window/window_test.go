package window

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/PicoBitSyS/PicoGoGUI/binding"
	"github.com/PicoBitSyS/PicoGoGUI/bridge"
	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/runtime"
)

func TestCallQueuesUntilReady(t *testing.T) {
	w := New(Config{})
	if err := w.Call("test", map[string]any{"value": 1}); err != nil {
		t.Fatal(err)
	}
	if len(w.pending) != 1 {
		t.Fatalf("pending messages = %d", len(w.pending))
	}
}

func TestCloseBeforeShowDisposesBindings(t *testing.T) {
	value := binding.New("initial")
	input := controls.NewTextBox().Bind(value)
	patcher := &testPatcher{}
	input.AttachHost(patcher)
	w := New(Config{})
	w.Add(input)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	value.Set("after-close")
	if patcher.calls != 0 {
		t.Fatalf("disposed control received %d patches", patcher.calls)
	}
	if err := w.Call("after-close", nil); err == nil {
		t.Fatal("expected closed-window error")
	}
}

func TestHandleRPCResult(t *testing.T) {
	w := New(Config{})
	result := make(chan rpcResult, 1)
	w.requests["rpc-1"] = result
	w.handleRPCResult(bridge.Message{
		Kind:    bridge.KindResponse,
		ID:      "rpc-1",
		Payload: json.RawMessage(`{"ok":true}`),
	})
	got := <-result
	if got.err != nil || string(got.payload) != `{"ok":true}` {
		t.Fatalf("rpc result = %#v", got)
	}
}

func TestHandleResizeUsesNativeOuterSize(t *testing.T) {
	w := New(Config{})
	var got ResizeEvent
	w.OnResize(func(event ResizeEvent) { got = event })
	w.handleResize(runtime.WindowSize{Width: 440, Height: 420, DPI: 192}, map[string]any{
		"width": 427.0, "height": 385.0,
	})
	if got.Outer != (Size{Width: 440, Height: 420}) || got.Client != (Size{Width: 427, Height: 385}) || got.DPI != 192 {
		t.Fatalf("resize event = %+v", got)
	}
}

func TestPersistSizeRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "window.json")
	w := New(Config{})
	if err := w.PersistSize(path); err != nil {
		t.Fatal(err)
	}
	w.lastOuter = Size{Width: 615, Height: 540}
	if err := w.savePersistedSize(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	restored := New(Config{})
	if err := restored.PersistSize(path); err != nil {
		t.Fatal(err)
	}
	if restored.outerSize != (Size{Width: 615, Height: 540}) {
		t.Fatalf("restored size = %+v", restored.outerSize)
	}
}

type testPatcher struct{ calls int }

func (p *testPatcher) Patch(string, map[string]any) error {
	p.calls++
	return nil
}
