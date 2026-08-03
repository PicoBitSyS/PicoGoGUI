# PicoGoGUI

Windows-first desktop GUI framework for Go. Declarative, strongly typed API inspired by WPF/WinUI. Internally powered by WebView2 — HTML/CSS/JS stay hidden from application developers.

Repository: https://github.com/PicoBitSyS/PicoGoGUI

## Requirements

- Windows 10/11
- Go 1.22+
- [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/) (preinstalled on modern Windows)

## Quick start

```bash
go get github.com/PicoBitSyS/PicoGoGUI
```

```go
package main

import (
	"log"

	gui "github.com/PicoBitSyS/PicoGoGUI"
)

func main() {
	app := gui.New()
	win := app.NewWindow("Hello PicoGoGUI")

	label := gui.Label("Hello, PicoGoGUI").ID("greeting")
	win.Add(
		label,
		gui.Button("Click me").OnClick(func() {
			label.Text("Clicked!")
			_ = win.Apply(label)
		}),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Run the bundled examples:

```bash
go run ./examples/helloworld
go run ./examples/settings
go run ./examples/table
go run ./examples/servermanager
go run ./examples/designer
go run ./examples/pluginhost
go run ./examples/nativewindow
```

Development hot reload:

```bash
go run ./cmd/picogogui run ./examples/helloworld
```

The development command watches Go and embedded UI assets, uses the Go build
cache, and restarts the application after a successful rebuild.

## Plugin System (Phase 6)

Plugins are **compile-time Go packages** (blank-import or `gui.UsePlugins`). They keep a single `.exe` on Windows — native Go `plugin` DLLs are not used.

Plugin metadata supports extension API compatibility (`MinAPI` / `MaxAPI`),
dependencies, capabilities, and optional activation/deactivation lifecycle.

```go
import _ "github.com/PicoBitSyS/PicoGoGUI/examples/plugins/badge"

app := gui.New() // activates registered plugins before the UI document is built
```

A plugin implements `plugin.Plugin`, registers from `init()`, and contributes:

- **Control kinds** — `RegisterControl` with CSS + JS `create` / `patch`
- **Designer kinds** — palette entry + optional Go codegen (`GoImport` / `GoExpr`)
- Extra CSS/JS blocks via `RegisterCSS` / `RegisterJS`

Sample: [`examples/plugins/badge`](examples/plugins/badge) and host [`examples/pluginhost`](examples/pluginhost).

## Visual Designer (Phase 5)

```bash
go run ./examples/designer
```

- **Toolbar** — file/edit actions plus align left/center/right/top/middle/bottom, horizontal/vertical distribution, lock and hide
- **Toolbox (left)** — controls + containers (`Column` / `Row` / `Stack`)
- **Canvas (center)** — Ctrl/Shift multi-select, group drag, **corner handles to resize**, arrow-key nudge (Shift = 10 px) and edge/center smart guides
- **Containers** — select a container, then add children (Parent is set automatically); container layers are always behind regular controls
- **Document Outline** — synchronized form hierarchy with container children plus locked/hidden state markers
- **Properties (right)** — identity, content, geometry, Z-index, lock/hide, font family/size, text/background colors, bold/italic/underline, alignment, border, radius and opacity
- **Status bar** — control count, selection, form size
- **Export Go** — geometry, stacking and appearance-preserving `gui.Canvas` / `gui.At` output with nested containers
- **Save / Load JSON** — native Windows file dialogs, validation, geometry + parent + appearance

Codegen lives in package [`designer`](designer) (`Document.GenerateGo()`).

CSS classes assigned to built-in controls in Designer are saved in JSON,
previewed on the design surface, and exported as strongly typed
`.Class("...")` calls.

## CSS classes

All built-in controls accept additional CSS classes, including `Label`:

```go
title := gui.Label("Dashboard").Class("page-title muted")
button := gui.Button("Save").Class("primary compact")
```

Classes are serialized through the normal component tree and can be updated at
runtime with `win.Apply(component)`. Layout containers continue to support the
same `.Class(...)` convention.

## DPI-aware native window geometry

```go
win.SetOuterSize(640, 480) // logical Windows pixels, stable across DPI/RDP

win.OnResize(func(event gui.ResizeEvent) {
    fmt.Printf("outer=%dx%d client=%dx%d dpi=%d\n",
        event.Outer.Width, event.Outer.Height,
        event.Client.Width, event.Client.Height,
        event.DPI,
    )
})

_ = win.PersistSize(filepath.Join(configDir, "MyApp", "window.json"))
```

`SetSize` remains the low-level client-size API for compatibility.
`SetOuterSize`, `OuterSize`, `OnResize`, and `PersistSize` use logical pixels
and account for the real native frame and per-monitor DPI.

The native geometry API is also the foundation for a future detached Designer
preview: the editable canvas can remain declarative while a secondary native
window renders and measures the real generated form on the active monitor.

## Themes

```go
app := gui.New(gui.Options{Theme: gui.ThemeSystem()}) // or ThemeDark / ThemeLight
app.SetTheme(gui.ThemeDark())
```

`ThemeSystem` follows Windows **Settings → Personalization → Colors → Choose your mode**. The Settings example includes a Dark theme checkbox.
System mode is monitored while the application runs. Strongly typed custom
theme tokens are available through `gui.CustomTheme(gui.ThemeDefinition{...})`.

## Animations and file drag & drop

```go
win.Add(gui.Animate(
    gui.Label("Ready"),
    gui.Animation{Name: gui.AnimationFadeIn, DurationMS: 180},
))

win.Add(gui.DropZone(gui.Label("Drop files here")).
    MaxBytes(4 << 20).
    OnDrop(func(files []gui.DroppedFile) {
        // Metadata plus optional bounded file content.
    }))
```

Built-in animations honor the Windows/browser reduced-motion preference.

## Native dialogs

```go
path, accepted, err := gui.OpenFile(gui.FileOptions{
    Title: "Open database",
    Filters: []gui.FileFilter{{Name: "SQLite", Pattern: "*.db"}},
})
```

`OpenFile`, `SaveFile`, and `SelectFolder` use native Windows dialogs.

## Multi-shell model (Desktop / Web / Tray / Service)

PicoGoGUI apps can expose **three UI channels** plus an optional Windows Service host. They share **Go business logic**, not the declarative `gui` tree:

| Channel | Package | Role |
|---------|---------|------|
| Desktop | `gui` | PicoGoGUI windows (WebView2) |
| Web UI | [`host/webui`](host/webui) | HTTP/API + static/embed pages (separate HTML) |
| Tray | [`host/tray`](host/tray) | System tray menu (NotifyIcon) |
| Service | [`host/service`](host/service) | Windows Service (`install`/`start`/`stop`/`remove`/`run`) |
| Notify | [`host/notify`](host/notify) | Balloon notifications |
| Clipboard | [`host/clipboard`](host/clipboard) | Get/Set text |

```go
import "github.com/PicoBitSyS/PicoGoGUI/host/webui"

web := webui.New()
web.Addr = "127.0.0.1:8080"
web.Handler = myAPIAndStatic // not gui components
_ = web.Start()
```

### Windows Service CLI (go_service pattern)

```bash
go build -o servermanager.exe ./examples/servermanager
servermanager.exe install    # admin
servermanager.exe start
servermanager.exe stop
servermanager.exe remove
servermanager.exe run        # console / foreground
servermanager.exe            # interactive: WebUI + tray
```

`FixWorkingDir` sets cwd next to the `.exe` in production (not `System32`). Service mode hosts **core + WebUI** only — Session 0 cannot show a tray icon.

## Status

**Phase 6** available: compile-time Plugin System (`plugin` package) — custom controls, assets, designer hooks, compatibility metadata, dependencies and lifecycle.

**Phase 5** available: advanced Visual Designer — multi-select, align/distribute, smart guides, lock/hide, Document Outline, layered containers, styled controls, undo/redo, validated native save/load and fidelity-preserving Go export.

**Phase 4** available: custom/system themes, animations, tray, asynchronous notifications, safe clipboard and file drag/drop.

**Phase 3:** Table sort/filter, accessible Tree, validated Form, prompt and native file/folder dialogs.

**Phase 2:** controls, origin-aware two-way binding, Grid/Canvas/Split/Tabs/Dock layouts.

Phase 1 foundation: application/window lifecycle, native WebView2 host, queued JSON bridge and error propagation.

The bridge also supports correlated request/response RPC through
`Window.CallRPC`. Runtime or extension JavaScript registers internal handlers
with `window.__picoRegisterRPC`; application developers remain in Go.

---

You are a senior Go software architect.

We are creating a brand new open-source framework called PicoGoGUI.

Mission
=======

PicoGoGUI is a Windows-first desktop GUI framework for Go.

The goal is NOT to create another wrapper over HTML.

The goal is to create the equivalent of Windows Presentation Foundation (WPF) but designed specifically for Go developers.

The developer should be able to build modern desktop applications without writing JavaScript, HTML or CSS.

Internally, PicoGoGUI will use:

- Go
- WebView2
- HTML
- CSS
- JavaScript

BUT these technologies must remain completely hidden from the developer.

Everything exposed publicly must look and feel like native Go.

The framework should generate and manage the UI automatically.

Architecture
============

Application

    Go Models
    Go Services
    Go Events

↓

Declarative Go API

    Window
    Button
    TextBox
    CheckBox
    ComboBox
    List
    Table
    Tree
    Tabs
    Form
    Dialog
    Notification

↓

Binding Engine

    Automatic synchronization

    Go State
        ↕
    UI State

↓

RPC Bridge

    Go ↔ JavaScript

↓

Web Runtime

    HTML
    CSS
    Components

↓

Native Host

    WebView2

Design Principles
=================

Everything must be:

- idiomatic Go
- clean
- declarative
- strongly typed
- easy to understand
- minimal boilerplate
- no code generation required
- no Node.js required by the application developer
- no JavaScript visible
- no HTML visible
- no CSS visible

The API should resemble:

app := gui.New()

window := app.NewWindow("Server")

window.Add(

    gui.Label("Server"),

    gui.TextBox().
        ID("host").
        Value("localhost"),

    gui.NumberBox().
        ID("port").
        Value(25),

    gui.Button("Start").
        OnClick(StartServer),

    gui.Table().
        Bind(&Connections),
)

app.Run()

Framework Goals
===============

Create a framework capable of building:

- server managers
- database tools
- administration consoles
- dashboards
- SQLite applications
- monitoring tools
- security applications
- office software
- configuration utilities

Windows Integration
===================

Support:

- WebView2
- Native Window
- Native File Dialog
- Native Notifications
- Tray Icon
- Context Menus
- Clipboard
- Drag & Drop
- Multiple Windows
- High DPI
- Dark Mode
- Acrylic / Mica (future)
- Windows 11 styling

Binding Engine
==============

Implement automatic reactive binding.

Example:

type Settings struct {
    Server string
    Port int
    SSL bool
}

Changing:

settings.Server

must automatically update every UI element bound to it.

Changing the UI must automatically update the Go struct.

No manual synchronization.

Component System
================

Every control must implement a common interface.

type Component interface {
    Render()
    Mount()
    Destroy()
}

Every component should support:

ID()

Visible()

Enabled()

Style()

Class()

Bind()

Events

Theme

Layout

Layouts
=======

Implement layout containers:

Column

Row

Grid

Stack

Dock

Split

Tabs

Accordion

Responsive Grid

Theme Engine
============

Implement a complete theme engine.

Support:

Dark

Light

Custom Themes

CSS variables should exist internally only.

Expose Go API only.

Example:

gui.ThemeDark()

or

gui.ThemeLight()

No CSS written by the developer.

Architecture
============

Repository layout

/picogogui

    /app

    /binding

    /bridge

    /controls

    /events

    /layout

    /runtime

    /theme

    /window

    /web

    /internal

    /examples

    /cmd

Examples
========

Create examples:

Hello World

Dashboard

SQLite CRUD

Server Manager

Settings Dialog

Table Example

Tree View

Tabs

Notifications

Goals for version 1.0
=====================

Stable public API

Excellent documentation

Fast startup

Low memory usage

Very small executable

No Electron

No Chromium Embedded Framework

Only WebView2

Long-term Vision
================

PicoGoGUI should become the de facto desktop framework for Go developers on Windows.

The framework should feel like:

WPF

+

WinUI

+

Go idioms

without exposing web technologies.

Implementation Rules
====================

Think as a framework architect.

Prioritize maintainability.

Favor composition over inheritance.

Use interfaces extensively.

Avoid global state.

Keep packages cohesive.

Write production-quality code.

Every public API must be documented.

Every exported function must include examples.

Every feature should include unit tests.

Every package should be independently testable.

Never implement quick hacks.

Always design for future extensibility.

Development Roadmap
===================

Phase 1
Core Window
WebView2 Host
RPC Bridge
Application lifecycle

Phase 2
Controls
Layouts
Events
Binding Engine

Phase 3
Dialogs
Tables
Trees
Forms
Themes

Phase 4
Notifications (host/notify)
Tray NotifyIcon (host/tray)
Clipboard text (host/clipboard)
Windows Service CLI + svc.Run (host/service, go_service pattern)
WebUI host (host/webui)
Drag & Drop
Animations

Phase 5
Advanced Visual Designer (examples/designer + designer package)
Multi-select/group drag, align/distribute, smart guides, lock/hide, Document Outline, deterministic layers, visual appearance, validation, undo/redo, native save/load, geometry/style codegen

Phase 6
Plugin System (compile-time registry, custom controls, designer hooks)

Important
=========

Before writing code, always evaluate whether the architecture remains clean, modular, and extensible.

Never sacrifice long-term design for short-term convenience.

PicoGoGUI should be something that Go developers enjoy using for the next 10 years.

Din experiența cu framework-uri GUI, acestea fac diferența:

Zero Dependencies Philosophy
aplicația finală să fie un singur .exe (în afară de WebView2 Runtime, deja instalat pe Windows moderne);
toate resursele (HTML, CSS, JS, fonturi, iconuri) să fie încorporate cu embed.FS.
Hot Reload pentru dezvoltare
picogogui run
modifici codul Go sau UI intern;
fereastra se actualizează instantaneu fără recompilare completă.
