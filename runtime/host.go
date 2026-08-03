package runtime

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/PicoBitSyS/PicoGoGUI/bridge"
)

// ErrUnsupportedPlatform is returned when the native host is unavailable.
var ErrUnsupportedPlatform = errors.New("picogogui: native host is only supported on Windows")

// Options configures a native host window.
type Options struct {
	Title       string
	Width       int
	Height      int
	OuterWidth  int
	OuterHeight int
	Debug       bool
}

// WindowSize is an outer native window size in logical Windows pixels.
type WindowSize struct {
	Width  int
	Height int
	DPI    int
}

// Handler receives decoded bridge messages from the web runtime.
type Handler func(bridge.Message)

// Host is the native WebView2 window and message bridge.
type Host struct {
	mu           sync.Mutex
	handler      Handler
	pending      []bridge.Message
	send         func(bridge.Message) error
	load         func()
	run          func()
	destroy      func()
	close        func()
	setTitle     func(string)
	setSize      func(int, int)
	setOuterSize func(int, int)
	outerSize    func() (WindowSize, error)
	ready        chan struct{}
	once         sync.Once
	destroyOnce  sync.Once
}

// SetHandler registers the inbound message handler.
// Any messages received before the handler was set are delivered immediately.
func (h *Host) SetHandler(fn Handler) {
	h.mu.Lock()
	h.handler = fn
	pending := h.pending
	h.pending = nil
	h.mu.Unlock()
	for _, m := range pending {
		fn(m)
	}
}

func (h *Host) dispatch(m bridge.Message) {
	if m.Kind == bridge.KindReady {
		h.once.Do(func() { close(h.ready) })
	}
	h.mu.Lock()
	fn := h.handler
	if fn == nil {
		h.pending = append(h.pending, m)
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()
	fn(m)
}

// Load navigates the webview to the embedded runtime document.
// Call SetHandler before Load so the initial ready event is not missed.
func (h *Host) Load() {
	if h != nil && h.load != nil {
		h.load()
	}
}

// Ready returns a channel closed when the web runtime signals ready.
func (h *Host) Ready() <-chan struct{} { return h.ready }

// Send delivers a message to the web runtime.
func (h *Host) Send(m bridge.Message) error {
	if h == nil || h.send == nil {
		return errors.New("picogogui: host is not started")
	}
	return h.send(m)
}

// SendJSON marshals payload into a message of the given kind.
func (h *Host) SendJSON(kind string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return h.Send(bridge.Message{Kind: kind, Payload: raw})
}

// Run starts the native message loop. It blocks until the window closes.
func (h *Host) Run() {
	if h != nil && h.run != nil {
		h.run()
	}
}

// Destroy releases native resources.
func (h *Host) Destroy() {
	if h != nil && h.destroy != nil {
		h.destroyOnce.Do(h.destroy)
	}
}

// Close requests that the native window close on its UI thread.
func (h *Host) Close() {
	if h != nil && h.close != nil {
		h.close()
	}
}

// SetTitle updates the native window title.
func (h *Host) SetTitle(title string) {
	if h != nil && h.setTitle != nil {
		h.setTitle(title)
	}
}

// SetSize updates the native window client size.
func (h *Host) SetSize(width, height int) {
	if h != nil && h.setSize != nil {
		h.setSize(width, height)
	}
}

// SetOuterSize updates the DPI-aware outer window size in logical pixels.
func (h *Host) SetOuterSize(width, height int) {
	if h != nil && h.setOuterSize != nil {
		h.setOuterSize(width, height)
	}
}

// OuterSize returns the DPI-aware outer native window size.
func (h *Host) OuterSize() (WindowSize, error) {
	if h == nil || h.outerSize == nil {
		return WindowSize{}, errors.New("picogogui: native window size is unavailable")
	}
	return h.outerSize()
}

func defaultOptions(opt Options) Options {
	if opt.Title == "" {
		opt.Title = "PicoGoGUI"
	}
	if opt.Width <= 0 {
		opt.Width = 900
	}
	if opt.Height <= 0 {
		opt.Height = 600
	}
	return opt
}
