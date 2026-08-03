// Package dialog provides in-window modal dialogs over the PicoGoGUI bridge.
package dialog

import (
	"fmt"
	"sync/atomic"

	"github.com/PicoBitSyS/PicoGoGUI/window"
)

var dlgSeq atomic.Uint64

func nextID() string {
	return fmt.Sprintf("dialog-%d", dlgSeq.Add(1))
}

// Message shows a simple OK dialog.
//
// Example:
//
//	dialog.Message(win, "Saved", "Settings written.")
func Message(w *window.Window, title, body string) {
	_ = MessageE(w, title, body)
}

// MessageE shows a simple OK dialog and reports argument or bridge errors.
func MessageE(w *window.Window, title, body string) error {
	if w == nil {
		return fmt.Errorf("dialog: window is required")
	}
	id := nextID()
	return w.Call("dialog.open", map[string]any{
		"id":      id,
		"kind":    "message",
		"title":   title,
		"body":    body,
		"buttons": []string{"OK"},
	})
}

// Confirm shows a Cancel/OK dialog and invokes fn with whether OK was chosen.
//
// Example:
//
//	dialog.Confirm(win, "Delete", "Remove this row?", func(ok bool) { ... })
func Confirm(w *window.Window, title, body string, fn func(ok bool)) {
	_ = ConfirmE(w, title, body, fn)
}

// ConfirmE shows a Cancel/OK dialog and reports argument or bridge errors.
func ConfirmE(w *window.Window, title, body string, fn func(ok bool)) error {
	if w == nil {
		return fmt.Errorf("dialog: window is required")
	}
	id := nextID()
	w.OnDialog(id, func(result map[string]any) {
		if fn == nil {
			return
		}
		ok, _ := result["ok"].(bool)
		fn(ok)
	})
	if err := w.Call("dialog.open", map[string]any{
		"id":      id,
		"kind":    "confirm",
		"title":   title,
		"body":    body,
		"buttons": []string{"Cancel", "OK"},
	}); err != nil {
		w.OnDialog(id, nil)
		return err
	}
	return nil
}

// Prompt shows a text-input dialog and invokes fn when the user completes it.
func Prompt(w *window.Window, title, body, value string, fn func(value string, ok bool)) {
	_ = PromptE(w, title, body, value, fn)
}

// PromptE shows a text-input dialog and reports argument or bridge errors.
func PromptE(w *window.Window, title, body, value string, fn func(value string, ok bool)) error {
	if w == nil {
		return fmt.Errorf("dialog: window is required")
	}
	id := nextID()
	w.OnDialog(id, func(result map[string]any) {
		if fn == nil {
			return
		}
		ok, _ := result["ok"].(bool)
		text, _ := result["value"].(string)
		fn(text, ok)
	})
	if err := w.Call("dialog.open", map[string]any{
		"id":          id,
		"kind":        "prompt",
		"title":       title,
		"body":        body,
		"value":       value,
		"inputLabel":  title,
		"dismissible": true,
		"buttons":     []string{"Cancel", "OK"},
	}); err != nil {
		w.OnDialog(id, nil)
		return err
	}
	return nil
}
