package events

import "testing"

func TestDispatchClickAndChange(t *testing.T) {
	d := NewDispatcher()
	reg := NewRegistry()
	clicked := false
	var changed any
	reg.OnClick("btn", func() { clicked = true })
	reg.OnChange("host", func(v any) { changed = v })
	d.SetRegistry(reg)

	if !d.Dispatch("btn", "click", nil) || !clicked {
		t.Fatal("click failed")
	}
	if !d.Dispatch("host", "change", "localhost") || changed != "localhost" {
		t.Fatalf("change failed: %v", changed)
	}
	if d.Dispatch("missing", "click", nil) {
		t.Fatal("unexpected handler")
	}
}

func TestDispatchSelect(t *testing.T) {
	d := NewDispatcher()
	reg := NewRegistry()
	var idx any
	reg.OnSelect("tbl", func(v any) { idx = v })
	d.SetRegistry(reg)
	if !d.Dispatch("tbl", "select", float64(2)) || idx != float64(2) {
		t.Fatalf("select failed: %v", idx)
	}
}
