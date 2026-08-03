package runtime

import (
	"strings"
	"testing"
)

func TestDocumentInlinesAssets(t *testing.T) {
	doc := Document()
	if doc == "" {
		t.Fatal("empty document")
	}
	for _, need := range []string{"pico-root", "__picoReceive", "--pico-bg", "picoGoSend", "signalReady", "__pico_window__", "picoCustomClass"} {
		if !strings.Contains(doc, need) {
			t.Fatalf("document missing %q", need)
		}
	}
	if strings.Contains(doc, `href="theme.css"`) || strings.Contains(doc, `src="app.js"`) {
		t.Fatal("document still references external assets")
	}
	if strings.Contains(doc, "chrome.webview.postMessage(data)") {
		t.Fatal("document must not prefer chrome.webview.postMessage for Go bridge")
	}
}
