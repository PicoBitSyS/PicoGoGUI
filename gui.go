// Package gui is the public facade for PicoGoGUI.
//
// Application developers should import this package only:
//
//	import gui "github.com/PicoBitSyS/PicoGoGUI"
package gui

import (
	"github.com/PicoBitSyS/PicoGoGUI/app"
	"github.com/PicoBitSyS/PicoGoGUI/binding"
	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/dialog"
	"github.com/PicoBitSyS/PicoGoGUI/layout"
	"github.com/PicoBitSyS/PicoGoGUI/plugin"
	"github.com/PicoBitSyS/PicoGoGUI/theme"
	"github.com/PicoBitSyS/PicoGoGUI/window"
)

// App is the application root.
type App = app.App

// Window is a top-level window.
type Window = window.Window

// Size is a width and height expressed in logical pixels.
type Size = window.Size

// ResizeEvent describes a DPI-aware native window resize.
type ResizeEvent = window.ResizeEvent

type FileOptions = dialog.FileOptions
type FileFilter = dialog.FileFilter

// Component is implemented by every control.
type Component = controls.Component

// Appearance configures font, colors, border, alignment, and opacity.
type Appearance = controls.Appearance

// New creates a PicoGoGUI application.
//
// Example:
//
//	app := gui.New(gui.Options{Theme: gui.ThemeSystem()})
//	win := app.NewWindow("Hello")
//	win.Add(gui.Label("Hello"), gui.Button("Click"))
//	app.Run()
func New(opts ...Options) *App {
	return app.New(opts...)
}

// Options configures application creation.
type Options = app.Options

// Label creates a text label.
func Label(text string) *controls.Label {
	return controls.NewLabel(text)
}

// Button creates a button.
func Button(text string) *controls.Button {
	return controls.NewButton(text)
}

// TextBox creates a single-line text input.
func TextBox() *controls.TextBox {
	return controls.NewTextBox()
}

// NumberBox creates a numeric input.
func NumberBox() *controls.NumberBox {
	return controls.NewNumberBox()
}

// CheckBox creates a labeled checkbox.
func CheckBox(text string) *controls.CheckBox {
	return controls.NewCheckBox(text)
}

// ComboBox creates a drop-down with the given items.
func ComboBox(items ...string) *controls.ComboBox {
	return controls.NewComboBox(items...)
}

// Animate applies a built-in motion effect to a component.
func Animate(component Component, animation controls.Animation) Component {
	return controls.Animate(component, animation)
}

// Animation configures a built-in motion effect.
type Animation = controls.Animation

// AnimationName identifies a built-in motion effect.
type AnimationName = controls.AnimationName

const (
	AnimationFadeIn  = controls.AnimationFadeIn
	AnimationSlideUp = controls.AnimationSlideUp
	AnimationScaleIn = controls.AnimationScaleIn
	AnimationPulse   = controls.AnimationPulse
)

// DropZone creates a Windows Explorer file drop target.
func DropZone(child Component) *controls.DropZone { return controls.NewDropZone(child) }

// DroppedFile is a file received by a DropZone.
type DroppedFile = controls.DroppedFile

// Column stacks children vertically.
func Column(children ...Component) *layout.Column {
	return layout.NewColumn(children...)
}

// Row arranges children horizontally.
func Row(children ...Component) *layout.Row {
	return layout.NewRow(children...)
}

// Stack overlays children.
func Stack(children ...Component) *layout.Stack {
	return layout.NewStack(children...)
}

// Grid arranges children in columns.
func Grid(columns int, children ...Component) *layout.Grid {
	return layout.NewGrid(columns, children...)
}

// Canvas creates an absolute-positioning surface.
func Canvas(children ...Component) *layout.Canvas { return layout.NewCanvas(children...) }

// At places a component at explicit canvas coordinates.
func At(child Component, x, y, width, height int) *layout.Positioned {
	return layout.At(child, x, y, width, height)
}

// Split creates a two-pane layout.
func Split(first, second Component) *layout.Split { return layout.NewSplit(first, second) }

// Tab creates one titled page for Tabs.
func Tab(title string, child Component) *layout.Tab { return layout.NewTab(title, child) }

// Tabs creates a tabbed layout.
func Tabs(pages ...*layout.Tab) *layout.Tabs { return layout.NewTabs(pages...) }

// Dock creates a dock layout.
func Dock(items ...layout.DockItem) *layout.Dock { return layout.NewDock(items...) }

// DockItem assigns a component to a dock region.
func DockItem(region layout.DockRegion, child Component) layout.DockItem {
	return layout.DockItem{Region: region, Child: child}
}

type DockRegion = layout.DockRegion

const (
	DockTop    = layout.DockTop
	DockRight  = layout.DockRight
	DockBottom = layout.DockBottom
	DockLeft   = layout.DockLeft
	DockCenter = layout.DockCenter
)

// Form stacks labeled fields.
func Form(fields ...*layout.Field) *layout.Form {
	return layout.NewForm(fields...)
}

// Field creates a labeled form field.
func Field(label string, control Component) *layout.Field {
	return layout.NewField(label, control)
}

// Table creates a data table.
func Table() *controls.Table {
	return controls.NewTable()
}

// DesignSurface creates a WinForms-style form preview for the designer.
func DesignSurface() *controls.DesignSurface {
	return controls.NewDesignSurface()
}

// Tree creates a hierarchical tree.
func Tree() *controls.Tree {
	return controls.NewTree()
}

// TreeNode creates a tree node.
func TreeNode(text string, children ...*controls.TreeNode) *controls.TreeNode {
	return controls.NewTreeNode(text, children...)
}

// Message shows an OK dialog on the window.
func Message(win *Window, title, body string) {
	dialog.Message(win, title, body)
}

// MessageE shows an OK dialog and reports argument or bridge errors.
func MessageE(win *Window, title, body string) error {
	return dialog.MessageE(win, title, body)
}

// Confirm shows a Cancel/OK dialog and invokes fn with the result.
func Confirm(win *Window, title, body string, fn func(ok bool)) {
	dialog.Confirm(win, title, body, fn)
}

// ConfirmE shows a Cancel/OK dialog and reports argument or bridge errors.
func ConfirmE(win *Window, title, body string, fn func(ok bool)) error {
	return dialog.ConfirmE(win, title, body, fn)
}

// Prompt shows a text-input dialog.
func Prompt(win *Window, title, body, value string, fn func(value string, ok bool)) {
	dialog.Prompt(win, title, body, value, fn)
}

// PromptE shows a text-input dialog and reports bridge errors.
func PromptE(win *Window, title, body, value string, fn func(value string, ok bool)) error {
	return dialog.PromptE(win, title, body, value, fn)
}

// OpenFile shows the native Windows open-file dialog.
func OpenFile(options FileOptions) (string, bool, error) { return dialog.OpenFile(options) }

// SaveFile shows the native Windows save-file dialog.
func SaveFile(options FileOptions) (string, bool, error) { return dialog.SaveFile(options) }

// SelectFolder shows the native Windows folder picker.
func SelectFolder(title string) (string, bool, error) { return dialog.SelectFolder(title) }

// Bind creates a reactive variable.
//
// Example:
//
//	host := gui.Bind("localhost")
func Bind[T any](v T) *binding.Var[T] {
	return binding.New(v)
}

// ThemeDark selects the dark theme.
func ThemeDark() theme.Name { return theme.DarkTheme() }

// ThemeLight selects the light theme.
func ThemeLight() theme.Name { return theme.LightTheme() }

// ThemeSystem follows the Windows light/dark apps preference.
func ThemeSystem() theme.Name { return theme.SystemTheme() }

// CustomTheme registers a strongly typed custom theme.
func CustomTheme(def theme.Definition) (theme.Name, error) { return theme.Register(def) }

// ThemeDefinition is a custom theme token set.
type ThemeDefinition = theme.Definition

// UsePlugins registers and activates compile-time plugins before the UI runs.
//
// Example:
//
//	_ = gui.UsePlugins(&badge.Plugin{})
func UsePlugins(plugins ...plugin.Plugin) error {
	return plugin.Use(plugins...)
}

// PluginInfo is metadata for an activated plugin.
type PluginInfo = plugin.Info
