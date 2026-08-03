package bridge

import (
	"encoding/json"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	raw, err := Encode(Message{Kind: KindReady})
	if err != nil {
		t.Fatal(err)
	}
	m, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m.Kind != KindReady {
		t.Fatalf("kind = %q", m.Kind)
	}
}

func TestNewMountAndPatch(t *testing.T) {
	mount, err := NewMount(map[string]any{"id": "root", "kind": "column"})
	if err != nil {
		t.Fatal(err)
	}
	if mount.Kind != KindMount {
		t.Fatalf("kind = %q", mount.Kind)
	}

	patch, err := NewPatch("label-1", map[string]any{"text": "Hi"})
	if err != nil {
		t.Fatal(err)
	}
	if patch.Kind != KindPatch {
		t.Fatalf("kind = %q", patch.Kind)
	}
}

func TestParseEventWithValue(t *testing.T) {
	payload, _ := json.Marshal(EventPayload{
		Target: "host",
		Name:   "change",
		Value:  "127.0.0.1",
	})
	m := Message{Kind: KindEvent, Payload: payload}
	p, err := ParseEvent(m)
	if err != nil {
		t.Fatal(err)
	}
	if p.Target != "host" || p.Name != "change" || p.Value != "127.0.0.1" {
		t.Fatalf("payload = %+v", p)
	}
}

func TestNewRequest(t *testing.T) {
	message, err := NewRequest("rpc-1", "runtime.info", map[string]any{"x": 1})
	if err != nil {
		t.Fatal(err)
	}
	if message.Kind != KindRequest || message.ID != "rpc-1" || message.Event != "runtime.info" {
		t.Fatalf("request = %#v", message)
	}
}
