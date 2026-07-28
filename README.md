# PicoGoGUI

> **A modern Windows-first GUI framework for Go.**
>
> Build beautiful desktop applications using a declarative Go API inspired by WPF and WinUI—without writing HTML, CSS, or JavaScript.

---

## Why PicoGoGUI?

Building desktop applications in Go is often a compromise:

- Learn a frontend framework (React, Vue, Svelte...)
- Write HTML and CSS
- Manage JavaScript
- Configure Node.js
- Bundle assets
- Deal with frontend tooling

PicoGoGUI eliminates all of that.

You write **Go**.

PicoGoGUI handles everything else.

Internally it uses modern web technologies powered by **Microsoft WebView2**, but these implementation details are completely hidden from the developer.

---

## Features

- Pure Go API
- Declarative UI
- Automatic Data Binding
- Reactive Components
- Windows 11 Design
- Dark & Light Themes
- High DPI Support
- Native File Dialogs
- Native Notifications
- System Tray
- Context Menus
- Multiple Windows
- Drag & Drop
- Clipboard Integration
- Fast Startup
- Small Executables
- Zero Electron
- Zero Chromium Embedded Framework

---

# Philosophy

PicoGoGUI follows one simple rule:

> **Write Go. Build Windows apps. Nothing else.**

No HTML.

No CSS.

No JavaScript.

No Node.js.

No Webpack.

No Vite.

Just Go.

---

# Example

```go
package main

import "github.com/PicoGoGUI/picogogui/gui"

func main() {

    app := gui.New()

    window := app.NewWindow("Hello")

    window.Add(

        gui.Label("Welcome to PicoGoGUI"),

        gui.TextBox().
            ID("username"),

        gui.Button("Login").
            OnClick(Login),
    )

    app.Run()
}
```

---

# Architecture

```
Application
│
├── Go Models
├── Go Services
├── Go Events
│
▼
Declarative Go API
│
├── Window
├── Button
├── TextBox
├── Table
├── Tree
├── Dialog
└── Layouts
│
▼
Binding Engine
│
Go State
↕
UI State
│
▼
RPC Bridge
│
Go ↔ JavaScript
│
▼
Web Runtime
│
HTML
CSS
Components
│
▼
Native Host
│
WebView2
```

---

# Design Goals

PicoGoGUI is inspired by:

- Windows Presentation Foundation (WPF)
- WinUI
- SwiftUI
- Jetpack Compose

while remaining completely idiomatic Go.

The public API should feel natural for Go developers.

---

# Roadmap

## Phase 1

- Window
- Application
- Event Loop
- WebView2 Host
- RPC Bridge

## Phase 2

- Layout System
- Button
- Label
- TextBox
- CheckBox
- ComboBox

## Phase 3

- Tables
- TreeView
- Forms
- Dialogs
- Notifications

## Phase 4

- Themes
- Animations
- System Tray
- Clipboard
- Drag & Drop

## Phase 5

- Visual Designer

## Phase 6

- Plugin System

---

# Core Components

- Window
- Label
- TextBox
- PasswordBox
- NumberBox
- Button
- ToggleButton
- CheckBox
- RadioButton
- ComboBox
- ListBox
- TreeView
- DataTable
- Tabs
- SplitView
- Grid
- Stack
- Row
- Column
- Dock
- Dialog
- Notification
- StatusBar
- Toolbar
- Menu
- Sidebar

---

# Data Binding

Automatic two-way binding.

```go
type Settings struct {
    Host string
    Port int
    SSL  bool
}
```

```go
gui.TextBox().
    Bind(&settings.Host)

gui.NumberBox().
    Bind(&settings.Port)

gui.CheckBox().
    Bind(&settings.SSL)
```

Changing the Go model updates the UI automatically.

Changing the UI updates the Go model automatically.

No manual synchronization required.

---

# Themes

Built-in themes:

- Light
- Dark

Future:

- Windows Accent
- Mica
- Acrylic
- Custom Themes

---

# Platform

Current target:

- Windows 10
- Windows 11

Future support:

- Linux
- macOS

---

# Why WebView2?

PicoGoGUI uses Microsoft Edge WebView2 because it provides:

- Modern rendering
- Small memory footprint
- Native Windows integration
- Excellent performance
- Long-term Microsoft support

Web technologies remain completely hidden behind the Go API.

---

# Project Structure

```
picogogui/

    app/
    binding/
    bridge/
    controls/
    events/
    layout/
    runtime/
    theme/
    web/
    window/
    internal/
    examples/
    cmd/
```

---

# Examples

Planned example applications:

- Hello World
- Dashboard
- SQLite CRUD
- File Explorer
- Server Manager
- Mail Server
- Configuration Tool
- Monitoring Dashboard
- System Tray App

---

# Principles

- Go First
- Windows First
- Performance First
- Simplicity First
- Developer Experience First

---

# Contributing

Contributions are welcome.

Please open an issue before implementing large features.

Coding style:

- Idiomatic Go
- Small packages
- Extensive documentation
- Unit tests
- Backward compatibility

---

# License

MIT License

See the LICENSE file for details.

---

# Vision

Our long-term vision is simple:

Become the de facto desktop GUI framework for Go developers on Windows.

Create beautiful native desktop applications with the simplicity that made Go successful.

**Write Go. Build Windows Apps.**
