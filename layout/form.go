// Package layout provides layout containers for arranging components.
package layout

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PicoBitSyS/PicoGoGUI/controls"
	"github.com/PicoBitSyS/PicoGoGUI/events"
)

// Field pairs a label with a control for forms.
type Field struct {
	id       string
	label    string
	control  controls.Component
	visible  bool
	enabled  bool
	required bool
	validate func(controls.Component) error
	errText  string
}

// NewField creates a labeled field.
//
// Example:
//
//	layout.NewField("Host", gui.TextBox().Bind(host))
func NewField(label string, control controls.Component) *Field {
	return &Field{
		id:      controls.AllocateID("field"),
		label:   label,
		control: control,
		visible: true,
		enabled: true,
	}
}

// ID sets the field identifier.
func (f *Field) ID(id string) *Field {
	f.id = id
	return f
}

// Required marks the field as requiring a non-empty value.
func (f *Field) Required(v bool) *Field {
	f.required = v
	return f
}

// ValidateWith sets a custom field validator.
func (f *Field) ValidateWith(fn func(controls.Component) error) *Field {
	f.validate = fn
	return f
}

// Error returns the last validation error.
func (f *Field) Error() string { return f.errText }

// CompID implements controls.Component.
func (f *Field) CompID() string { return f.id }

// Kind implements controls.Component.
func (f *Field) Kind() string { return "field" }

// ChildComponents implements controls.Container.
func (f *Field) ChildComponents() []controls.Component {
	if f.control == nil {
		return nil
	}
	return []controls.Component{f.control}
}

// Node implements controls.Component.
func (f *Field) Node() controls.Node {
	children := []controls.Node{
		{
			ID:    f.id + "-label",
			Kind:  "label",
			Props: map[string]any{"text": f.label, "visible": true, "enabled": true},
		},
	}
	if f.control != nil {
		children = append(children, f.control.Node())
	}
	if f.errText != "" {
		children = append(children, controls.Node{
			ID:   f.id + "-error",
			Kind: "label",
			Props: map[string]any{
				"text":    f.errText,
				"class":   "pico-field-error",
				"visible": true,
				"enabled": true,
			},
		})
	}
	return controls.Node{
		ID:       f.id,
		Kind:     "field",
		Props:    map[string]any{"text": f.label, "visible": f.visible, "enabled": f.enabled},
		Children: children,
	}
}

func (f *Field) runValidation() error {
	f.errText = ""
	if f.required && strings.TrimSpace(componentValue(f.control)) == "" {
		f.errText = f.label + " is required"
		return errors.New(f.errText)
	}
	if f.validate != nil {
		if err := f.validate(f.control); err != nil {
			f.errText = err.Error()
			return err
		}
	}
	return nil
}

func componentValue(c controls.Component) string {
	switch value := c.(type) {
	case interface{ GetValue() string }:
		return value.GetValue()
	case interface{ GetValue() float64 }:
		return strings.TrimSpace(strings.TrimRight(strings.TrimRight(
			fmt.Sprintf("%f", value.GetValue()), "0"), "."))
	case interface{ IsChecked() bool }:
		if value.IsChecked() {
			return "true"
		}
	}
	return ""
}

// CollectHandlers implements controls.Component.
func (f *Field) CollectHandlers(reg *events.Registry) {
	if f.control != nil {
		f.control.CollectHandlers(reg)
	}
}

// Form stacks labeled fields vertically.
type Form struct {
	id      string
	fields  []*Field
	visible bool
	enabled bool
}

// ValidationError identifies an invalid form field.
type ValidationError struct {
	FieldID string
	Message string
}

// NewForm creates a form from fields.
//
// Example:
//
//	layout.NewForm(layout.NewField("Host", tb), layout.NewField("Port", nb))
func NewForm(fields ...*Field) *Form {
	return &Form{
		id:      controls.AllocateID("form"),
		fields:  fields,
		visible: true,
		enabled: true,
	}
}

// ID sets the form identifier.
func (f *Form) ID(id string) *Form {
	f.id = id
	return f
}

// CompID implements controls.Component.
func (f *Form) CompID() string { return f.id }

// Kind implements controls.Component.
func (f *Form) Kind() string { return "form" }

// ChildComponents implements controls.Container.
func (f *Form) ChildComponents() []controls.Component {
	out := make([]controls.Component, 0, len(f.fields))
	for _, field := range f.fields {
		if field != nil {
			out = append(out, field)
		}
	}
	return out
}

// Node implements controls.Component.
func (f *Form) Node() controls.Node {
	children := make([]controls.Node, 0, len(f.fields))
	for _, field := range f.fields {
		if field == nil {
			continue
		}
		children = append(children, field.Node())
	}
	return controls.Node{
		ID:       f.id,
		Kind:     "form",
		Props:    map[string]any{"visible": f.visible, "enabled": f.enabled},
		Children: children,
	}
}

// CollectHandlers implements controls.Component.
func (f *Form) CollectHandlers(reg *events.Registry) {
	for _, field := range f.fields {
		if field != nil {
			field.CollectHandlers(reg)
		}
	}
}

// Validate evaluates required and custom validators and returns all errors.
func (f *Form) Validate() []ValidationError {
	var out []ValidationError
	for _, field := range f.fields {
		if field == nil {
			continue
		}
		if err := field.runValidation(); err != nil {
			out = append(out, ValidationError{FieldID: field.id, Message: err.Error()})
		}
	}
	return out
}
