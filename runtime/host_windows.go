//go:build windows

package runtime

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jchv/go-webview2"

	"github.com/PicoBitSyS/PicoGoGUI/bridge"
	"github.com/PicoBitSyS/PicoGoGUI/internal/win32"
)

// NewHost creates a Windows WebView2 host window.
func NewHost(opt Options) (*Host, error) {
	opt = defaultOptions(opt)
	win32.SetDPIAware()
	initialWidth, initialHeight := opt.Width, opt.Height
	if opt.OuterWidth > 0 && opt.OuterHeight > 0 {
		if width, height, sizeErr := win32.ClientSizeForOuter(0, opt.OuterWidth, opt.OuterHeight); sizeErr == nil {
			initialWidth, initialHeight = width, height
		}
	}
	document, err := DocumentE()
	if err != nil {
		return nil, err
	}

	wv := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     opt.Debug,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  opt.Title,
			Width:  uint(initialWidth),
			Height: uint(initialHeight),
			Center: true,
		},
	})
	if wv == nil {
		return nil, win32.ErrWebView2Missing
	}

	h := &Host{ready: make(chan struct{})}

	// Queue Go→JS messages until the page runtime installs __picoReceive.
	wv.Init(`window.__picoPending = window.__picoPending || [];
window.__picoReceive = function(msg){ (window.__picoPending = window.__picoPending || []).push(msg); };`)

	if err := wv.Bind("picoGoSend", func(msg string) {
		m, err := bridge.Decode([]byte(msg))
		if err != nil {
			return
		}
		h.dispatch(m)
	}); err != nil {
		wv.Destroy()
		return nil, fmt.Errorf("picogogui: bind bridge: %w", err)
	}

	h.send = func(m bridge.Message) error {
		raw, err := bridge.Encode(m)
		if err != nil {
			return err
		}
		// Ensure payload is an object for the JS runtime when present.
		var envelope map[string]any
		if err := json.Unmarshal(raw, &envelope); err != nil {
			return err
		}
		if len(m.Payload) > 0 {
			var payload any
			if err := json.Unmarshal(m.Payload, &payload); err == nil {
				envelope["payload"] = payload
			}
		}
		body, err := json.Marshal(envelope)
		if err != nil {
			return err
		}
		script := `(function(msg){
  if (typeof window.__picoReceive === "function") { window.__picoReceive(msg); }
  else { (window.__picoPending = window.__picoPending || []).push(msg); }
})(JSON.parse(` + strconv.Quote(string(body)) + `))`
		wv.Dispatch(func() { wv.Eval(script) })
		return nil
	}
	h.load = func() {
		width, height := initialWidth, initialHeight
		if opt.OuterWidth > 0 && opt.OuterHeight > 0 {
			if clientWidth, clientHeight, sizeErr := win32.ClientSizeForOuter(uintptr(wv.Window()), opt.OuterWidth, opt.OuterHeight); sizeErr == nil {
				width, height = clientWidth, clientHeight
			}
		}
		wv.SetSize(width, height, webview2.HintNone)
		wv.SetHtml(document)
	}
	h.run = wv.Run
	h.destroy = wv.Destroy
	h.close = func() { wv.Dispatch(wv.Destroy) }
	h.setTitle = func(title string) {
		wv.Dispatch(func() { wv.SetTitle(title) })
	}
	h.setSize = func(width, height int) {
		wv.Dispatch(func() { wv.SetSize(width, height, webview2.HintNone) })
	}
	h.setOuterSize = func(width, height int) {
		wv.Dispatch(func() {
			clientWidth, clientHeight, sizeErr := win32.ClientSizeForOuter(uintptr(wv.Window()), width, height)
			if sizeErr == nil {
				wv.SetSize(clientWidth, clientHeight, webview2.HintNone)
			}
		})
	}
	h.outerSize = func() (WindowSize, error) {
		size, sizeErr := win32.OuterWindowSize(uintptr(wv.Window()))
		return WindowSize{Width: size.Width, Height: size.Height, DPI: size.DPI}, sizeErr
	}
	return h, nil
}
