package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gui "github.com/PicoBitSyS/PicoGoGUI"
	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/designer"
	"github.com/PicoBitSyS/PicoGoGUI/host/clipboard"
	"github.com/PicoBitSyS/PicoGoGUI/plugin"

	_ "github.com/PicoBitSyS/PicoGoGUI/examples/plugins/badge"
)

const layoutFile = "designer-layout.json"

func main() {
	app := gui.New(gui.Options{Theme: gui.ThemeDark()})
	win := app.NewWindow("PicoGoGUI Designer")
	win.SetOuterSize(1440, 840)
	if configDir, err := os.UserConfigDir(); err == nil {
		_ = win.PersistSize(filepath.Join(configDir, "PicoGoGUI", "designer-window.json"))
	}

	doc := designer.NewDocument("MyWindow")
	selected := -1
	selection := []int{}
	layoutPath := layoutFile

	titleBox := gui.TextBox().ID("winTitle").Value(doc.WindowTitle)
	widthBox := gui.TextBox().ID("winW").Value(strconv.Itoa(doc.Width))
	heightBox := gui.TextBox().ID("winH").Value(strconv.Itoa(doc.Height))
	idBox := gui.TextBox().ID("propID")
	classBox := gui.TextBox().ID("propClass")
	textBox := gui.TextBox().ID("propText")
	valueBox := gui.TextBox().ID("propValue")
	parentBox := gui.TextBox().ID("propParent")
	xBox := gui.TextBox().ID("propX")
	yBox := gui.TextBox().ID("propY")
	wBox := gui.TextBox().ID("propWidth")
	hBox := gui.TextBox().ID("propHeight")
	zBox := gui.TextBox().ID("propZIndex")
	fontBox := gui.ComboBox("Theme default", "Segoe UI", "Arial", "Calibri", "Consolas", "Georgia", "Times New Roman").
		ID("propFont").Value("Theme default")
	fontSizeBox := gui.TextBox().ID("propFontSize")
	colorBox := gui.TextBox().ID("propColor")
	backgroundBox := gui.TextBox().ID("propBackground")
	alignBox := gui.ComboBox("Theme default", "left", "center", "right", "justify").
		ID("propTextAlign").Value("Theme default")
	boldBox := gui.CheckBox("Bold").ID("propBold")
	italicBox := gui.CheckBox("Italic").ID("propItalic")
	underlineBox := gui.CheckBox("Underline").ID("propUnderline")
	borderColorBox := gui.TextBox().ID("propBorderColor")
	borderWidthBox := gui.TextBox().ID("propBorderWidth")
	borderRadiusBox := gui.TextBox().ID("propBorderRadius")
	opacityBox := gui.TextBox().ID("propOpacity")
	lockedBox := gui.CheckBox("Locked").ID("propLocked")
	hiddenBox := gui.CheckBox("Hidden in app").ID("propHidden")

	statusControls := gui.Label("0").ID("statusControls")
	statusSelected := gui.Label("(none)").ID("statusSelected")
	statusForm := gui.Label("480 x 360").ID("statusForm")
	statusMsg := gui.Label("Ready — drag to move, corners to resize").ID("statusMsg")

	surface := gui.DesignSurface().ID("surface").
		Title(doc.WindowTitle).
		FormSize(doc.Width, doc.Height).
		Widgets(doc.SurfaceWidgets())
	outline := gui.Tree().ID("documentOutline")
	outlineExpanded := map[string]bool{"outline:form": true}

	var buildOutlineNodes func() []*controls.TreeNode
	buildOutlineNodes = func() []*controls.TreeNode {
		var childrenOf func(string) []*controls.TreeNode
		childrenOf = func(parent string) []*controls.TreeNode {
			children := make([]*controls.TreeNode, 0)
			for _, widget := range doc.ChildrenOf(parent) {
				prefix := ""
				if widget.Locked {
					prefix += "🔒 "
				}
				if widget.Hidden {
					prefix += "◌ "
				}
				label := prefix + widget.ID + "  [" + widget.Kind + "]"
				if widget.Text != "" {
					label += " — " + truncate(widget.Text, 24)
				}
				nodeID := "outline:" + widget.ID
				node := gui.TreeNode(label, childrenOf(widget.ID)...).WithID(nodeID)
				expanded, known := outlineExpanded[nodeID]
				node.Expanded = !known || expanded
				children = append(children, node)
			}
			return children
		}
		form := gui.TreeNode(doc.WindowTitle+"  [form]", childrenOf("")...).WithID("outline:form")
		form.Expanded = outlineExpanded["outline:form"]
		return []*controls.TreeNode{form}
	}

	refreshOutline := func() {
		outline.Nodes(buildOutlineNodes()...)
		if selected >= 0 && selected < len(doc.Widgets) {
			outline.Selected("outline:" + doc.Widgets[selected].ID)
		} else {
			outline.Selected("outline:form")
		}
		_ = win.Apply(outline)
	}

	syncFormMeta := func() {
		doc.WindowTitle = titleBox.GetValue()
		if w, err := strconv.Atoi(strings.TrimSpace(widthBox.GetValue())); err == nil && w > 0 {
			doc.Width = w
		}
		if h, err := strconv.Atoi(strings.TrimSpace(heightBox.GetValue())); err == nil && h > 0 {
			doc.Height = h
		}
	}

	refreshSurface := func() {
		syncFormMeta()
		surface.Title(doc.WindowTitle).
			FormSize(doc.Width, doc.Height).
			Widgets(doc.SurfaceWidgets()).
			Selection(selection...).
			Refresh()
		refreshOutline()
		statusControls.Text(strconv.Itoa(len(doc.Widgets)))
		statusForm.Text(strconv.Itoa(doc.Width) + " x " + strconv.Itoa(doc.Height))
		_ = win.Apply(statusControls)
		_ = win.Apply(statusForm)
	}

	// Live-update form chrome when title / W / H change.
	titleBox.OnChange(func(string) { refreshSurface() })
	widthBox.OnChange(func(string) { refreshSurface() })
	heightBox.OnChange(func(string) { refreshSurface() })

	setStatus := func(msg string) {
		statusMsg.Text(msg)
		_ = win.Apply(statusMsg)
	}

	setPropBoxes := func(w designer.Widget) {
		idBox.Value(w.ID)
		classBox.Value(w.Class)
		textBox.Value(w.Text)
		valueBox.Value(w.Value)
		parentBox.Value(w.Parent)
		xBox.Value(strconv.Itoa(w.X))
		yBox.Value(strconv.Itoa(w.Y))
		wBox.Value(strconv.Itoa(w.Width))
		hBox.Value(strconv.Itoa(w.Height))
		zBox.Value(strconv.Itoa(w.ZIndex))
		font := w.Appearance.FontFamily
		if font == "" {
			font = "Theme default"
		}
		fontBox.Value(font)
		fontSizeBox.Value(strconv.Itoa(w.Appearance.FontSize))
		colorBox.Value(w.Appearance.Color)
		backgroundBox.Value(w.Appearance.Background)
		align := w.Appearance.TextAlign
		if align == "" {
			align = "Theme default"
		}
		alignBox.Value(align)
		boldBox.Checked(w.Appearance.Bold)
		italicBox.Checked(w.Appearance.Italic)
		underlineBox.Checked(w.Appearance.Underline)
		borderColorBox.Value(w.Appearance.BorderColor)
		borderWidthBox.Value(strconv.Itoa(w.Appearance.BorderWidth))
		borderRadiusBox.Value(strconv.Itoa(w.Appearance.BorderRadius))
		if w.Appearance.Opacity == 0 {
			opacityBox.Value("")
		} else {
			opacityBox.Value(strconv.FormatFloat(w.Appearance.Opacity, 'f', -1, 64))
		}
		lockedBox.Checked(w.Locked)
		hiddenBox.Checked(w.Hidden)
		_ = win.Apply(idBox)
		_ = win.Apply(classBox)
		_ = win.Apply(textBox)
		_ = win.Apply(valueBox)
		_ = win.Apply(parentBox)
		_ = win.Apply(xBox)
		_ = win.Apply(yBox)
		_ = win.Apply(wBox)
		_ = win.Apply(hBox)
		_ = win.Apply(zBox)
		_ = win.Apply(fontBox)
		_ = win.Apply(fontSizeBox)
		_ = win.Apply(colorBox)
		_ = win.Apply(backgroundBox)
		_ = win.Apply(alignBox)
		_ = win.Apply(boldBox)
		_ = win.Apply(italicBox)
		_ = win.Apply(underlineBox)
		_ = win.Apply(borderColorBox)
		_ = win.Apply(borderWidthBox)
		_ = win.Apply(borderRadiusBox)
		_ = win.Apply(opacityBox)
		_ = win.Apply(lockedBox)
		_ = win.Apply(hiddenBox)
	}

	clearPropBoxes := func() {
		setPropBoxes(designer.Widget{})
	}

	loadSelectionSet := func(indices []int) {
		source := append([]int(nil), indices...)
		seen := map[int]bool{}
		selection = selection[:0]
		for _, index := range source {
			if index >= 0 && index < len(doc.Widgets) && !seen[index] {
				seen[index] = true
				selection = append(selection, index)
			}
		}
		selected = -1
		if len(selection) > 0 {
			selected = selection[len(selection)-1]
		}
		surface.Selection(selection...)
		_ = win.Apply(surface)
		refreshOutline()
		if selected < 0 {
			clearPropBoxes()
			statusSelected.Text("(none)")
			_ = win.Apply(statusSelected)
			return
		}
		w := doc.Widgets[selected]
		setPropBoxes(w)
		if len(selection) > 1 {
			statusSelected.Text(strconv.Itoa(len(selection)) + " controls")
		} else {
			statusSelected.Text(w.Kind + " #" + w.ID)
		}
		_ = win.Apply(statusSelected)
		setStatus("Selected " + strconv.Itoa(len(selection)) + " — Ctrl/Shift-click to modify selection")
	}

	loadSelection := func(i int) {
		if i < 0 {
			loadSelectionSet(nil)
			return
		}
		loadSelectionSet([]int{i})
	}

	surface.OnSelection(loadSelectionSet)
	surface.OnLayout(func(i, x, y, width, height int) {
		if i < 0 || i >= len(doc.Widgets) {
			return
		}
		doc.SetGeometry(i, x, y, width, height)
		selected = i
		if !containsIndex(selection, i) {
			selection = []int{i}
		}
		setPropBoxes(doc.Widgets[i])
		setStatus("Moved " + doc.Widgets[i].ID)
		// Keep JS geometry; only sync selection class / props — full refresh optional
		surface.Widgets(doc.SurfaceWidgets()).Selection(selection...)
		refreshOutline()
	})
	surface.OnLayouts(func(changes []controls.DesignLayoutChange) {
		modelChanges := make([]designer.GeometryChange, 0, len(changes))
		for _, change := range changes {
			modelChanges = append(modelChanges, designer.GeometryChange{
				Index: change.Index, X: change.X, Y: change.Y,
				Width: change.Width, Height: change.Height,
			})
		}
		if doc.SetGeometries(modelChanges) {
			surface.Widgets(doc.SurfaceWidgets()).Selection(selection...)
			refreshOutline()
			setStatus("Moved " + strconv.Itoa(len(modelChanges)) + " controls")
		}
	})
	outline.OnSelect(func(nodeID string) {
		if nodeID == "outline:form" {
			loadSelection(-1)
			return
		}
		id := strings.TrimPrefix(nodeID, "outline:")
		loadSelection(doc.IndexOfID(id))
	})
	outline.OnToggle(func(id string, expanded bool) {
		outlineExpanded[id] = expanded
	})

	parentForNew := func() string {
		if selected >= 0 && selected < len(doc.Widgets) {
			w := doc.Widgets[selected]
			if designer.IsContainer(w.Kind) {
				return w.ID
			}
			return w.Parent
		}
		return ""
	}

	add := func(kind string) {
		syncFormMeta()
		doc.Add(designer.Widget{Kind: kind, Parent: parentForNew()})
		refreshSurface()
		loadSelection(len(doc.Widgets) - 1)
		setStatus("Added " + kind)
	}

	parseIntBox := func(get func() string, fallback int) int {
		n, err := strconv.Atoi(strings.TrimSpace(get()))
		if err != nil {
			return fallback
		}
		return n
	}
	parseFloatBox := func(get func() string, fallback float64) float64 {
		n, err := strconv.ParseFloat(strings.TrimSpace(get()), 64)
		if err != nil {
			return fallback
		}
		return n
	}

	applyProps := func() {
		syncFormMeta()
		if selected < 0 || selected >= len(doc.Widgets) {
			refreshSurface()
			setStatus("Form updated")
			return
		}
		cur := doc.Widgets[selected]
		font := strings.TrimSpace(fontBox.GetValue())
		if font == "Theme default" {
			font = ""
		}
		align := strings.TrimSpace(alignBox.GetValue())
		if align == "Theme default" {
			align = ""
		}
		w := designer.Widget{
			Kind:   cur.Kind,
			ID:     idBox.GetValue(),
			Class:  classBox.GetValue(),
			Text:   textBox.GetValue(),
			Value:  valueBox.GetValue(),
			Parent: strings.TrimSpace(parentBox.GetValue()),
			X:      parseIntBox(xBox.GetValue, cur.X),
			Y:      parseIntBox(yBox.GetValue, cur.Y),
			Width:  parseIntBox(wBox.GetValue, cur.Width),
			Height: parseIntBox(hBox.GetValue, cur.Height),
			ZIndex: parseIntBox(zBox.GetValue, cur.ZIndex),
			Locked: lockedBox.IsChecked(),
			Hidden: hiddenBox.IsChecked(),
			Appearance: gui.Appearance{
				FontFamily:   font,
				FontSize:     parseIntBox(fontSizeBox.GetValue, cur.Appearance.FontSize),
				Color:        strings.TrimSpace(colorBox.GetValue()),
				Background:   strings.TrimSpace(backgroundBox.GetValue()),
				Bold:         boldBox.IsChecked(),
				Italic:       italicBox.IsChecked(),
				Underline:    underlineBox.IsChecked(),
				TextAlign:    align,
				BorderColor:  strings.TrimSpace(borderColorBox.GetValue()),
				BorderWidth:  parseIntBox(borderWidthBox.GetValue, cur.Appearance.BorderWidth),
				BorderRadius: parseIntBox(borderRadiusBox.GetValue, cur.Appearance.BorderRadius),
				Opacity:      parseFloatBox(opacityBox.GetValue, cur.Appearance.Opacity),
			},
		}
		if err := doc.UpdateAtE(selected, w); err != nil {
			gui.Message(win, "Properties", err.Error())
			setStatus("Invalid properties")
			return
		}
		refreshSurface()
		// Selection index may change if ID list order same
		if idx := doc.IndexOfID(w.ID); idx >= 0 {
			loadSelection(idx)
		} else {
			loadSelection(selected)
		}
		setStatus("Applied properties")
	}

	deleteSelected := func() {
		if len(selection) == 0 {
			gui.Message(win, "Delete", "Select a control first.")
			return
		}
		deletable := make([]int, 0, len(selection))
		for _, index := range selection {
			if index >= 0 && index < len(doc.Widgets) && !doc.Widgets[index].Locked {
				deletable = append(deletable, index)
			}
		}
		if len(deletable) == 0 {
			gui.Message(win, "Delete", "The selected controls are locked.")
			return
		}
		gui.Confirm(win, "Delete", "Remove "+strconv.Itoa(len(deletable))+" selected control(s) and their children?", func(ok bool) {
			if !ok {
				return
			}
			doc.RemoveIndices(deletable)
			selected = -1
			selection = nil
			refreshSurface()
			loadSelection(-1)
			setStatus("Deleted " + strconv.Itoa(len(deletable)) + " control(s)")
		})
	}

	duplicateSelected := func() {
		if len(selection) == 0 {
			gui.Message(win, "Duplicate", "Select a control first.")
			return
		}
		source := append([]int(nil), selection...)
		newSelection := make([]int, 0, len(source))
		for _, index := range source {
			if index < 0 || index >= len(doc.Widgets) {
				continue
			}
			src := doc.Widgets[index]
			doc.Add(designer.Widget{
				Kind:       src.Kind,
				Text:       src.Text,
				Value:      src.Value,
				Class:      src.Class,
				Parent:     src.Parent,
				X:          src.X + 16,
				Y:          src.Y + 16,
				Width:      src.Width,
				Height:     src.Height,
				Appearance: src.Appearance,
				Hidden:     src.Hidden,
			})
			newSelection = append(newSelection, len(doc.Widgets)-1)
		}
		refreshSurface()
		loadSelectionSet(newSelection)
		setStatus("Duplicated " + strconv.Itoa(len(newSelection)) + " control(s)")
	}

	undo := func() {
		if !doc.Undo() {
			setStatus("Nothing to undo")
			return
		}
		selected = -1
		refreshSurface()
		loadSelection(-1)
		setStatus("Undo")
	}

	redo := func() {
		if !doc.Redo() {
			setStatus("Nothing to redo")
			return
		}
		selected = -1
		refreshSurface()
		loadSelection(-1)
		setStatus("Redo")
	}

	bringToFront := func() {
		if selected < 0 || selected >= len(doc.Widgets) {
			setStatus("Select a control first")
			return
		}
		if !doc.BringToFront(selected) {
			setStatus("Already at the front of its layer")
			return
		}
		refreshSurface()
		setPropBoxes(doc.Widgets[selected])
		setStatus("Brought " + doc.Widgets[selected].ID + " to front")
	}

	sendToBack := func() {
		if selected < 0 || selected >= len(doc.Widgets) {
			setStatus("Select a control first")
			return
		}
		if !doc.SendToBack(selected) {
			setStatus("Already at the back of its layer")
			return
		}
		refreshSurface()
		setPropBoxes(doc.Widgets[selected])
		setStatus("Sent " + doc.Widgets[selected].ID + " to back")
	}

	resetAppearance := func() {
		if selected < 0 || selected >= len(doc.Widgets) {
			setStatus("Select a control first")
			return
		}
		w := doc.Widgets[selected]
		w.Appearance = gui.Appearance{}
		if err := doc.UpdateAtE(selected, w); err != nil {
			gui.Message(win, "Appearance", err.Error())
			return
		}
		refreshSurface()
		setPropBoxes(doc.Widgets[selected])
		setStatus("Reset appearance for " + w.ID)
	}

	alignSelection := func(mode designer.AlignMode) {
		if !doc.Align(selection, mode) {
			setStatus("Alignment needs at least 2 unlocked controls with the same parent")
			return
		}
		refreshSurface()
		loadSelectionSet(selection)
		setStatus("Aligned " + strconv.Itoa(len(selection)) + " control(s)")
	}

	distributeSelection := func(mode designer.DistributeMode) {
		if !doc.Distribute(selection, mode) {
			setStatus("Distribution needs at least 3 unlocked controls with the same parent")
			return
		}
		refreshSurface()
		loadSelectionSet(selection)
		setStatus("Distributed " + strconv.Itoa(len(selection)) + " control(s)")
	}

	toggleLock := func() {
		if len(selection) == 0 {
			setStatus("Select one or more controls first")
			return
		}
		lock := false
		for _, index := range selection {
			if index >= 0 && index < len(doc.Widgets) && !doc.Widgets[index].Locked {
				lock = true
				break
			}
		}
		if doc.SetLocked(selection, lock) {
			refreshSurface()
			loadSelectionSet(selection)
		}
		if lock {
			setStatus("Locked selection")
		} else {
			setStatus("Unlocked selection")
		}
	}

	toggleHidden := func() {
		if len(selection) == 0 {
			setStatus("Select one or more controls first")
			return
		}
		hide := false
		for _, index := range selection {
			if index >= 0 && index < len(doc.Widgets) && !doc.Widgets[index].Hidden {
				hide = true
				break
			}
		}
		if doc.SetHidden(selection, hide) {
			refreshSurface()
			loadSelectionSet(selection)
		}
		if hide {
			setStatus("Hidden in generated UI; retained as ghost in Designer")
		} else {
			setStatus("Selection is visible")
		}
	}

	loadSample := func() {
		*doc = *designer.NewDocument("SampleForm")
		doc.Width = 480
		doc.Height = 320
		doc.Add(designer.Widget{Kind: designer.KindColumn, ID: "column1", Text: "Main", X: 16, Y: 16, Width: 280, Height: 240})
		doc.Add(designer.Widget{
			Kind: designer.KindLabel, Text: "Connection settings", Parent: "column1",
			X: 12, Y: 20, Width: 220, Height: 32,
			Appearance: gui.Appearance{
				FontFamily: "Segoe UI", FontSize: 18, Color: "#174a7e",
				Bold: true, Underline: true,
			},
		})
		doc.Add(designer.Widget{Kind: designer.KindLabel, Text: "Host", Parent: "column1", X: 12, Y: 62, Width: 60, Height: 22})
		doc.Add(designer.Widget{Kind: designer.KindTextBox, Value: "localhost", Parent: "column1", X: 80, Y: 58, Width: 160, Height: 28})
		doc.Add(designer.Widget{Kind: designer.KindCheckBox, Text: "Use SSL", Parent: "column1", X: 12, Y: 98, Width: 120, Height: 24})
		doc.Add(designer.Widget{Kind: designer.KindRow, ID: "row1", Text: "Actions", Parent: "column1", X: 12, Y: 140, Width: 240, Height: 56})
		doc.Add(designer.Widget{Kind: designer.KindButton, Text: "Connect", Parent: "row1", X: 8, Y: 12, Width: 90, Height: 32})
		doc.Add(designer.Widget{Kind: designer.KindButton, Text: "Cancel", Parent: "row1", X: 110, Y: 12, Width: 90, Height: 32})
		titleBox.Value(doc.WindowTitle)
		widthBox.Value(strconv.Itoa(doc.Width))
		heightBox.Value(strconv.Itoa(doc.Height))
		_ = win.Apply(titleBox)
		_ = win.Apply(widthBox)
		_ = win.Apply(heightBox)
		selected = -1
		refreshSurface()
		loadSelection(-1)
		setStatus("Loaded sample form with containers")
	}

	toolbar := gui.Row(
		gui.Label("PicoGoGUI Designer"),
		gui.Button("Sample").OnClick(loadSample),
		gui.Button("Save").OnClick(func() {
			syncFormMeta()
			if err := doc.Validate(); err != nil {
				gui.Message(win, "Save", err.Error())
				return
			}
			path, accepted, err := gui.SaveFile(gui.FileOptions{
				Title:       "Save PicoGoGUI layout",
				DefaultName: filepath.Base(layoutPath),
				DefaultExt:  "json",
				Filters:     []gui.FileFilter{{Name: "PicoGoGUI layout", Pattern: "*.json"}},
			})
			if err != nil {
				gui.Message(win, "Save", err.Error())
				return
			}
			if !accepted {
				return
			}
			layoutPath = path
			raw, err := doc.MarshalJSON()
			if err != nil {
				gui.Message(win, "Save", err.Error())
				return
			}
			if err := os.WriteFile(layoutPath, raw, 0o644); err != nil {
				gui.Message(win, "Save", err.Error())
				return
			}
			setStatus("Saved " + filepath.Base(layoutPath))
		}),
		gui.Button("Load").OnClick(func() {
			path, accepted, err := gui.OpenFile(gui.FileOptions{
				Title:   "Open PicoGoGUI layout",
				Filters: []gui.FileFilter{{Name: "PicoGoGUI layout", Pattern: "*.json"}},
			})
			if err != nil {
				gui.Message(win, "Load", err.Error())
				return
			}
			if !accepted {
				return
			}
			layoutPath = path
			raw, err := os.ReadFile(layoutPath)
			if err != nil {
				gui.Message(win, "Load", err.Error())
				return
			}
			parsed, err := designer.ParseDocument(raw)
			if err != nil {
				gui.Message(win, "Load", err.Error())
				return
			}
			*doc = *parsed
			titleBox.Value(doc.WindowTitle)
			widthBox.Value(strconv.Itoa(doc.Width))
			heightBox.Value(strconv.Itoa(doc.Height))
			_ = win.Apply(titleBox)
			_ = win.Apply(widthBox)
			_ = win.Apply(heightBox)
			selected = -1
			refreshSurface()
			loadSelection(-1)
			setStatus("Loaded " + filepath.Base(layoutPath))
		}),
		gui.Button("Export Go").OnClick(func() {
			syncFormMeta()
			if err := doc.Validate(); err != nil {
				gui.Message(win, "Export Go", err.Error())
				return
			}
			src := doc.GenerateGo()
			if err := clipboard.SetText(src); err != nil {
				gui.Message(win, "Export Go", "Clipboard failed: "+err.Error()+"\n\n"+truncate(src, 800))
				return
			}
			gui.Message(win, "Export Go", "Source copied to clipboard ("+strconv.Itoa(len(src))+" bytes).")
			setStatus("Exported Go to clipboard")
		}),
		gui.Button("Duplicate").OnClick(duplicateSelected),
		gui.Button("Delete").OnClick(deleteSelected),
		gui.Button("Send Back").OnClick(sendToBack),
		gui.Button("Bring Front").OnClick(bringToFront),
		gui.Button("Undo").OnClick(undo),
		gui.Button("Redo").OnClick(redo),
	).Class("pico-designer-toolbar")

	arrangeToolbar := gui.Row(
		gui.Label("Arrange"),
		gui.Button("Left").OnClick(func() { alignSelection(designer.AlignLeft) }),
		gui.Button("H Center").OnClick(func() { alignSelection(designer.AlignHCenter) }),
		gui.Button("Right").OnClick(func() { alignSelection(designer.AlignRight) }),
		gui.Button("Top").OnClick(func() { alignSelection(designer.AlignTop) }),
		gui.Button("V Center").OnClick(func() { alignSelection(designer.AlignVCenter) }),
		gui.Button("Bottom").OnClick(func() { alignSelection(designer.AlignBottom) }),
		gui.Button("Distribute H").OnClick(func() { distributeSelection(designer.DistributeHorizontal) }),
		gui.Button("Distribute V").OnClick(func() { distributeSelection(designer.DistributeVertical) }),
		gui.Button("Lock / Unlock").OnClick(toggleLock),
		gui.Button("Hide / Show").OnClick(toggleHidden),
	).Class("pico-designer-toolbar pico-designer-arrangebar")

	paletteChildren := []gui.Component{
		gui.Label("Controls"),
		gui.Button("Label").OnClick(func() { add(designer.KindLabel) }),
		gui.Button("Button").OnClick(func() { add(designer.KindButton) }),
		gui.Button("TextBox").OnClick(func() { add(designer.KindTextBox) }),
		gui.Button("NumberBox").OnClick(func() { add(designer.KindNumberBox) }),
		gui.Button("CheckBox").OnClick(func() { add(designer.KindCheckBox) }),
		gui.Button("ComboBox").OnClick(func() { add(designer.KindComboBox) }),
		gui.Label("Containers"),
		gui.Button("Column").OnClick(func() { add(designer.KindColumn) }),
		gui.Button("Row").OnClick(func() { add(designer.KindRow) }),
		gui.Button("Stack").OnClick(func() { add(designer.KindStack) }),
	}
	_ = plugin.Activate()
	seenPluginGroup := map[string]bool{}
	for _, pk := range plugin.DesignerKinds() {
		pk := pk
		if !seenPluginGroup[pk.Group] {
			paletteChildren = append(paletteChildren, gui.Label(pk.Group))
			seenPluginGroup[pk.Group] = true
		}
		paletteChildren = append(paletteChildren, gui.Button(pk.Label).OnClick(func() {
			w := designer.Widget{Kind: pk.Kind, Parent: parentForNew()}
			if pk.Default != nil {
				if t := pk.Default["text"]; t != "" {
					w.Text = t
				}
				if v := pk.Default["value"]; v != "" {
					w.Value = v
				}
			}
			syncFormMeta()
			doc.Add(w)
			refreshSurface()
			loadSelection(len(doc.Widgets) - 1)
			setStatus("Added " + pk.Kind)
		}))
	}
	palette := gui.Column(paletteChildren...).Class("pico-designer-palette")

	canvasPane := gui.Column(
		gui.Row(
			gui.Label("Form W:"),
			widthBox,
			gui.Label("H:"),
			heightBox,
			gui.Label("Title:"),
			titleBox,
		).Class("pico-designer-formbar"),
		surface,
	).Class("pico-designer-canvas")

	props := gui.Column(
		gui.Label("Document Outline"),
		outline,
		gui.Label("Properties"),
		gui.Form(
			gui.Field("ID", idBox),
			gui.Field("CSS class", classBox),
			gui.Field("Text", textBox),
			gui.Field("Value", valueBox),
			gui.Field("Parent", parentBox),
			gui.Field("X", xBox),
			gui.Field("Y", yBox),
			gui.Field("Width", wBox),
			gui.Field("Height", hBox),
			gui.Field("Z index", zBox),
			gui.Field("Font", fontBox),
			gui.Field("Font size", fontSizeBox),
			gui.Field("Text color", colorBox),
			gui.Field("Background", backgroundBox),
			gui.Field("Align", alignBox),
			gui.Field("Text style", gui.Row(boldBox, italicBox, underlineBox).Class("pico-designer-stylechecks")),
			gui.Field("Border", borderColorBox),
			gui.Field("Border px", borderWidthBox),
			gui.Field("Radius", borderRadiusBox),
			gui.Field("Opacity", opacityBox),
			gui.Field("Designer", gui.Row(lockedBox, hiddenBox).Class("pico-designer-stylechecks")),
		),
		gui.Row(
			gui.Button("Apply").OnClick(applyProps),
			gui.Button("Reset style").OnClick(resetAppearance),
			gui.Button("Send Back").OnClick(sendToBack),
			gui.Button("Bring Front").OnClick(bringToFront),
			gui.Button("Delete").OnClick(deleteSelected),
		).Class("pico-designer-actions"),
	).Class("pico-designer-props")

	status := gui.Row(
		gui.Label("Controls:"),
		statusControls,
		gui.Label("Selected:"),
		statusSelected,
		gui.Label("Form:"),
		statusForm,
		gui.Label("·"),
		statusMsg,
	).Class("pico-designer-statusbar")

	win.Add(
		gui.Column(
			toolbar,
			arrangeToolbar,
			gui.Row(palette, canvasPane, props).Class("pico-designer-shell"),
			status,
		).Class("pico-designer-app"),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func containsIndex(indices []int, target int) bool {
	for _, index := range indices {
		if index == target {
			return true
		}
	}
	return false
}
