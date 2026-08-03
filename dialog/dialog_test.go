package dialog

import (
	"testing"

	"github.com/PicoBitSyS/PicoGoGUI/bridge"
)

func TestDialogCallPayload(t *testing.T) {
	msg, err := bridge.NewCall("dialog.open", map[string]any{
		"id":      "dialog-1",
		"kind":    "confirm",
		"title":   "Delete",
		"body":    "Remove?",
		"buttons": []string{"Cancel", "OK"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.Kind != bridge.KindCall || msg.Event != "dialog.open" {
		t.Fatalf("msg = %+v", msg)
	}
}
