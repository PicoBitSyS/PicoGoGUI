// Package bridge implements the JSON message protocol between Go and the web runtime.
package bridge

import "encoding/json"

// Message kinds exchanged over the RPC bridge.
const (
	KindReady    = "ready"
	KindMount    = "mount"
	KindPatch    = "patch"
	KindCall     = "call"
	KindEvent    = "event"
	KindRequest  = "request"
	KindResponse = "response"
	KindError    = "error"
)

// Message is a single bridge frame.
type Message struct {
	Kind    string          `json:"kind"`
	ID      string          `json:"id,omitempty"`
	Event   string          `json:"event,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Encode marshals a message to JSON bytes.
func Encode(m Message) ([]byte, error) {
	return json.Marshal(m)
}

// Decode unmarshals JSON bytes into a message.
func Decode(data []byte) (Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	return m, err
}

// EventPayload is carried by KindEvent messages from the web runtime.
type EventPayload struct {
	Target string `json:"target"`
	Name   string `json:"name"`
	Value  any    `json:"value,omitempty"`
}

// PatchPayload describes a partial UI update.
type PatchPayload struct {
	ID    string         `json:"id"`
	Props map[string]any `json:"props"`
}

// NewMount builds a mount message with the given tree payload.
func NewMount(tree any) (Message, error) {
	raw, err := json.Marshal(tree)
	if err != nil {
		return Message{}, err
	}
	return Message{Kind: KindMount, Payload: raw}, nil
}

// NewPatch builds a patch message for a single component.
func NewPatch(id string, props map[string]any) (Message, error) {
	raw, err := json.Marshal(PatchPayload{ID: id, Props: props})
	if err != nil {
		return Message{}, err
	}
	return Message{Kind: KindPatch, Payload: raw}, nil
}

// NewCall builds a call message (theme, dialog.open, …).
func NewCall(event string, payload any) (Message, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Message{}, err
	}
	return Message{Kind: KindCall, Event: event, Payload: raw}, nil
}

// NewRequest builds a correlated RPC request.
func NewRequest(id, method string, payload any) (Message, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Message{}, err
	}
	return Message{Kind: KindRequest, ID: id, Event: method, Payload: raw}, nil
}

// ParseEvent extracts an EventPayload from a message.
func ParseEvent(m Message) (EventPayload, error) {
	var p EventPayload
	if len(m.Payload) > 0 {
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			return EventPayload{}, err
		}
	}
	if p.Target == "" {
		p.Target = m.ID
	}
	if p.Name == "" {
		p.Name = m.Event
	}
	return p, nil
}
