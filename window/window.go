// Package window implements native application windows backed by the runtime host.
package window

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/PicoBitSyS/PicoGoGUI/bridge"
	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/events"
	"github.com/PicoBitSyS/PicoGoGUI/runtime"
	"github.com/PicoBitSyS/PicoGoGUI/theme"
)

var requestSequence atomic.Uint64

const windowEventTarget = "__pico_window__"

// Size is a width and height expressed in logical pixels.
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ResizeEvent describes a DPI-aware native resize.
type ResizeEvent struct {
	Outer  Size
	Client Size
	DPI    int
}

type rpcResult struct {
	payload json.RawMessage
	err     error
}

// Window is a top-level application window.
type Window struct {
	mu          sync.Mutex
	title       string
	width       int
	height      int
	debug       bool
	theme       theme.Name
	children    []controls.Component
	host        *runtime.Host
	dispatcher  *events.Dispatcher
	dialogs     map[string]func(map[string]any)
	mounted     bool
	ready       bool
	closed      bool
	pending     []bridge.Message
	errs        chan error
	firstErr    error
	onClosing   func() bool
	onClosed    func()
	onResize    func(ResizeEvent)
	outerSize   Size
	lastOuter   Size
	persistPath string
	requests    map[string]chan rpcResult
}

// Config holds window construction options.
type Config struct {
	Title       string
	Width       int
	Height      int
	OuterWidth  int
	OuterHeight int
	Debug       bool
	Theme       theme.Name
}

// New creates a window that has not yet been shown.
func New(cfg Config) *Window {
	if cfg.Title == "" {
		cfg.Title = "PicoGoGUI"
	}
	if cfg.Width <= 0 {
		cfg.Width = 900
	}
	if cfg.Height <= 0 {
		cfg.Height = 600
	}
	if cfg.Theme == "" {
		cfg.Theme = theme.Light
	}
	return &Window{
		title:      cfg.Title,
		width:      cfg.Width,
		height:     cfg.Height,
		debug:      cfg.Debug,
		theme:      cfg.Theme,
		outerSize:  Size{Width: cfg.OuterWidth, Height: cfg.OuterHeight},
		dispatcher: events.NewDispatcher(),
		dialogs:    make(map[string]func(map[string]any)),
		errs:       make(chan error, 16),
		requests:   make(map[string]chan rpcResult),
	}
}

// Add appends components to the window content.
func (w *Window) Add(components ...controls.Component) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.children = append(w.children, components...)
	controls.AttachHosts(w, components...)
	w.rebuildHandlersLocked()
	if w.mounted && w.host != nil {
		_ = w.mountLocked()
	}
}

// SetSize updates the preferred window size before or after show.
func (w *Window) SetSize(width, height int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.width = width
	w.height = height
	if w.host != nil {
		w.host.SetSize(width, height)
	}
}

// SetOuterSize updates the DPI-aware outer window size in logical pixels.
func (w *Window) SetOuterSize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	w.mu.Lock()
	w.outerSize = Size{Width: width, Height: height}
	host := w.host
	w.mu.Unlock()
	if host != nil {
		host.SetOuterSize(width, height)
	}
}

// OuterSize returns the current DPI-aware outer native size. Before Show it
// returns the size configured with SetOuterSize, when available.
func (w *Window) OuterSize() (Size, error) {
	w.mu.Lock()
	host := w.host
	configured := w.outerSize
	last := w.lastOuter
	w.mu.Unlock()
	if host != nil {
		size, err := host.OuterSize()
		if err == nil {
			return Size{Width: size.Width, Height: size.Height}, nil
		}
	}
	if last.Width > 0 && last.Height > 0 {
		return last, nil
	}
	if configured.Width > 0 && configured.Height > 0 {
		return configured, nil
	}
	return Size{}, errors.New("picogogui: outer window size is unavailable before Show")
}

// OnResize registers a DPI-aware native window resize handler.
func (w *Window) OnResize(fn func(ResizeEvent)) {
	w.mu.Lock()
	w.onResize = fn
	w.mu.Unlock()
}

// SetTitle updates the window title before or after show.
func (w *Window) SetTitle(title string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.title = title
	if w.host != nil {
		w.host.SetTitle(title)
	}
}

// OnClosing registers a callback for programmatic close requests. Returning
// false cancels the close.
func (w *Window) OnClosing(fn func() bool) {
	w.mu.Lock()
	w.onClosing = fn
	w.mu.Unlock()
}

// OnClosed registers a callback invoked after the native loop exits.
func (w *Window) OnClosed(fn func()) {
	w.mu.Lock()
	w.onClosed = fn
	w.mu.Unlock()
}

// Errors returns asynchronous mount, theme, and bridge errors.
func (w *Window) Errors() <-chan error { return w.errs }

// SetTheme updates the active theme and notifies the runtime when mounted.
func (w *Window) SetTheme(name theme.Name) {
	w.mu.Lock()
	w.theme = name
	var err error
	if w.mounted && w.host != nil {
		err = w.sendThemeLocked()
	}
	w.mu.Unlock()
	w.recordError(err)
}

// Patch sends a property update for a mounted component.
//
// Example:
//
//	win.Patch(label.CompID(), map[string]any{"text": "Updated"})
func (w *Window) Patch(id string, props map[string]any) error {
	msg, err := bridge.NewPatch(id, props)
	if err != nil {
		return err
	}
	return w.sendOrQueue(msg)
}

// Call sends a runtime call message (theme, dialog.open, …).
func (w *Window) Call(event string, payload any) error {
	msg, err := bridge.NewCall(event, payload)
	if err != nil {
		return err
	}
	return w.sendOrQueue(msg)
}

// CallRPC invokes a registered JavaScript runtime method and waits for its
// correlated response, cancellation, or error.
func (w *Window) CallRPC(ctx context.Context, method string, payload any) (json.RawMessage, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	id := fmt.Sprintf("rpc-%d", requestSequence.Add(1))
	message, err := bridge.NewRequest(id, method, payload)
	if err != nil {
		return nil, err
	}
	result := make(chan rpcResult, 1)
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil, errors.New("picogogui: window is closed")
	}
	w.requests[id] = result
	w.mu.Unlock()
	if err := w.sendOrQueue(message); err != nil {
		w.mu.Lock()
		delete(w.requests, id)
		w.mu.Unlock()
		return nil, err
	}
	select {
	case response := <-result:
		return response.payload, response.err
	case <-ctx.Done():
		w.mu.Lock()
		delete(w.requests, id)
		w.mu.Unlock()
		return nil, ctx.Err()
	}
}

// OnDialog registers a one-shot result handler for a dialog id.
func (w *Window) OnDialog(id string, fn func(map[string]any)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if fn == nil {
		delete(w.dialogs, id)
		return
	}
	if w.dialogs == nil {
		w.dialogs = make(map[string]func(map[string]any))
	}
	w.dialogs[id] = fn
}

// Apply patches the UI from a component's current Node props.
func (w *Window) Apply(c controls.Component) error {
	if c == nil {
		return nil
	}
	n := c.Node()
	return w.Patch(n.ID, n.Props)
}

// Refresh re-sends the full component tree (e.g. after table data changes).
func (w *Window) Refresh() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.mounted || w.host == nil {
		return nil
	}
	return w.mountLocked()
}

// Show creates the native host and mounts the component tree once the runtime is ready.
func (w *Window) Show() error {
	w.mu.Lock()
	if w.host != nil {
		w.mu.Unlock()
		return nil
	}
	title, width, height, outerSize, debug := w.title, w.width, w.height, w.outerSize, w.debug
	w.mu.Unlock()

	host, err := runtime.NewHost(runtime.Options{
		Title:       title,
		Width:       width,
		Height:      height,
		OuterWidth:  outerSize.Width,
		OuterHeight: outerSize.Height,
		Debug:       debug,
	})
	if err != nil {
		return err
	}

	host.SetHandler(func(m bridge.Message) {
		switch m.Kind {
		case bridge.KindReady:
			w.mu.Lock()
			errMount := w.mountLocked()
			errTheme := w.sendThemeLocked()
			w.ready = errMount == nil
			pending := append([]bridge.Message(nil), w.pending...)
			if w.ready {
				w.pending = nil
			}
			w.mu.Unlock()
			w.recordError(errMount)
			w.recordError(errTheme)
			if errMount == nil {
				for _, msg := range pending {
					w.recordError(host.Send(msg))
				}
			}
		case bridge.KindEvent:
			p, err := bridge.ParseEvent(m)
			if err != nil {
				return
			}
			if p.Name == "dialog" {
				w.handleDialogResult(p.Target, p.Value)
				return
			}
			if p.Target == windowEventTarget && p.Name == "resize" {
				size, sizeErr := host.OuterSize()
				if sizeErr == nil {
					w.handleResize(size, p.Value)
				}
				return
			}
			w.dispatcher.Dispatch(p.Target, p.Name, p.Value)
		case bridge.KindResponse, bridge.KindError:
			w.handleRPCResult(m)
		}
	})

	w.mu.Lock()
	w.host = host
	controls.AttachHosts(w, w.children...)
	w.mu.Unlock()
	host.Load()
	return nil
}

// Run shows the window (if needed) and blocks on the native message loop.
func (w *Window) Run() error {
	if err := w.Show(); err != nil {
		return err
	}
	w.mu.Lock()
	host := w.host
	w.mu.Unlock()
	defer host.Destroy()
	host.Run()
	persistErr := w.savePersistedSize()
	w.mu.Lock()
	w.closed = true
	w.ready = false
	onClosed := w.onClosed
	err := errors.Join(w.firstErr, persistErr)
	w.mu.Unlock()
	w.Dispose()
	if onClosed != nil {
		onClosed()
	}
	return err
}

func (w *Window) handleResize(size runtime.WindowSize, value any) {
	client := sizeFromValue(value)
	event := ResizeEvent{
		Outer:  Size{Width: size.Width, Height: size.Height},
		Client: client,
		DPI:    size.DPI,
	}
	w.mu.Lock()
	w.lastOuter = event.Outer
	fn := w.onResize
	w.mu.Unlock()
	if fn != nil {
		fn(event)
	}
}

func sizeFromValue(value any) Size {
	raw, _ := value.(map[string]any)
	if raw == nil {
		return Size{}
	}
	return Size{Width: intFromAny(raw["width"]), Height: intFromAny(raw["height"])}
}

func intFromAny(value any) int {
	switch number := value.(type) {
	case float64:
		return int(number)
	case int:
		return number
	case json.Number:
		result, _ := number.Int64()
		return int(result)
	default:
		return 0
	}
}

// Close requests that the native window close. OnClosing may cancel it.
func (w *Window) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	onClosing := w.onClosing
	host := w.host
	w.mu.Unlock()
	if onClosing != nil && !onClosing() {
		return errors.New("picogogui: window close cancelled")
	}
	if host == nil {
		w.mu.Lock()
		w.closed = true
		w.mu.Unlock()
		w.Dispose()
		return nil
	}
	host.Close()
	return nil
}

// Dispose releases component subscriptions owned by the window.
func (w *Window) Dispose() {
	w.mu.Lock()
	children := append([]controls.Component(nil), w.children...)
	requests := w.requests
	w.requests = make(map[string]chan rpcResult)
	w.mu.Unlock()
	for _, result := range requests {
		result <- rpcResult{err: errors.New("picogogui: window disposed")}
	}
	controls.DisposeAll(children...)
}

func (w *Window) mountLocked() error {
	controls.AttachHosts(w, w.children...)
	w.rebuildHandlersLocked()
	tree := controls.Root(w.children...)
	msg, err := bridge.NewMount(tree)
	if err != nil {
		return err
	}
	if err := w.host.Send(msg); err != nil {
		return err
	}
	w.mounted = true
	return nil
}

func (w *Window) sendThemeLocked() error {
	if w.host == nil {
		return nil
	}
	payload, err := json.Marshal(map[string]any{
		"name":      string(w.theme),
		"variables": theme.Variables(w.theme),
	})
	if err != nil {
		return err
	}
	return w.host.Send(bridge.Message{
		Kind:    bridge.KindCall,
		Event:   "theme",
		Payload: payload,
	})
}

func (w *Window) rebuildHandlersLocked() {
	w.dispatcher.SetRegistry(controls.CollectAllHandlers(w.children...))
}

func (w *Window) handleDialogResult(id string, value any) {
	w.mu.Lock()
	fn := w.dialogs[id]
	delete(w.dialogs, id)
	w.mu.Unlock()
	if fn == nil {
		return
	}
	res, _ := value.(map[string]any)
	if res == nil {
		res = map[string]any{}
	}
	fn(res)
}

func (w *Window) sendOrQueue(msg bridge.Message) error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return errors.New("picogogui: window is closed")
	}
	host := w.host
	if host == nil || !w.ready {
		w.pending = append(w.pending, msg)
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()
	return host.Send(msg)
}

func (w *Window) recordError(err error) {
	if err == nil {
		return
	}
	w.mu.Lock()
	if w.firstErr == nil {
		w.firstErr = err
	}
	errs := w.errs
	w.mu.Unlock()
	select {
	case errs <- err:
	default:
	}
}

func (w *Window) handleRPCResult(message bridge.Message) {
	w.mu.Lock()
	result := w.requests[message.ID]
	delete(w.requests, message.ID)
	w.mu.Unlock()
	if result == nil {
		return
	}
	if message.Kind == bridge.KindError {
		var payload struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(message.Payload, &payload)
		if payload.Message == "" {
			payload.Message = "runtime RPC failed"
		}
		result <- rpcResult{err: errors.New(payload.Message)}
		return
	}
	result <- rpcResult{payload: append(json.RawMessage(nil), message.Payload...)}
}
