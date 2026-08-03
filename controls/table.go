package controls

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/PicoBitSyS/PicoGoGUI/events"
)

// Table displays tabular data with optional slice binding.
type Table struct {
	base
	columns    []string
	fields     []string // struct field names aligned with columns
	rows       []map[string]any
	selected   int
	slicePtr   any
	onSelect   func(int)
	onActivate func(int)
	onSort     func(column string, ascending bool)
	patcher    Patcher
	sortable   bool
	filter     string
	lastErr    error
}

// NewTable creates an empty table.
//
// Example:
//
//	gui.Table().Columns("Host", "Port").Bind(&rows)
func NewTable() *Table {
	return &Table{
		base:     newBase("table"),
		selected: -1,
	}
}

// ID sets the component identifier.
func (t *Table) ID(id string) *Table {
	t.id = id
	return t
}

// Columns sets visible column titles (also used as default field names).
func (t *Table) Columns(names ...string) *Table {
	t.columns = append([]string(nil), names...)
	if len(t.fields) == 0 {
		t.fields = append([]string(nil), names...)
	}
	return t
}

// Fields sets struct field names corresponding to Columns.
func (t *Table) Fields(names ...string) *Table {
	t.fields = append([]string(nil), names...)
	return t
}

// Visible sets visibility.
func (t *Table) Visible(v bool) *Table {
	t.visible = v
	return t
}

// Enabled sets enabled state.
func (t *Table) Enabled(v bool) *Table {
	t.enabled = v
	return t
}

// OnSelect registers a row-select handler (0-based index).
func (t *Table) OnSelect(fn func(int)) *Table {
	t.onSelect = fn
	return t
}

// OnActivate registers a double-click handler.
func (t *Table) OnActivate(fn func(int)) *Table {
	t.onActivate = fn
	return t
}

// Bind attaches a pointer to a slice of structs or maps and loads rows.
//
// Example:
//
//	gui.Table().Columns("Host", "Port").Bind(&connections)
func (t *Table) Bind(slicePtr any) *Table {
	_ = t.BindE(slicePtr)
	return t
}

// BindE attaches a slice and reports invalid binding shapes.
func (t *Table) BindE(slicePtr any) error {
	rows, err := sliceToRows(slicePtr, t.columns, t.fields)
	t.slicePtr = slicePtr
	t.lastErr = err
	if err != nil {
		return err
	}
	t.rows = rows
	return nil
}

// Refresh rebuilds rows from the bound slice and patches the UI when hosted.
func (t *Table) Refresh() {
	_ = t.RefreshE()
}

// RefreshE rebuilds rows and reports binding errors.
func (t *Table) RefreshE() error {
	t.reloadFromBind()
	if t.lastErr != nil {
		return t.lastErr
	}
	if t.patcher != nil {
		if err := t.patcher.Patch(t.id, t.props()); err != nil {
			return err
		}
	}
	return nil
}

// Err returns the most recent binding error.
func (t *Table) Err() error { return t.lastErr }

// Sortable enables clickable column sorting in the runtime.
func (t *Table) Sortable(v bool) *Table {
	t.sortable = v
	return t
}

// Filter limits visible rows to cells containing text, case-insensitively.
func (t *Table) Filter(text string) *Table {
	t.filter = text
	return t
}

// OnSort registers a column sort handler.
func (t *Table) OnSort(fn func(column string, ascending bool)) *Table {
	t.onSort = fn
	return t
}

// Selected returns the selected row index, or -1.
func (t *Table) Selected() int { return t.selected }

// AttachHost implements HostAware.
func (t *Table) AttachHost(p Patcher) { t.patcher = p }

// Kind implements Component.
func (t *Table) Kind() string { return "table" }

func (t *Table) props() map[string]any {
	props := map[string]any{
		"columns":  append([]string(nil), t.columns...),
		"rows":     t.rows,
		"selected": t.selected,
		"sortable": t.sortable,
		"filter":   t.filter,
	}
	t.applyCommonProps(props)
	return props
}

// Node implements Component.
func (t *Table) Node() Node {
	return Node{ID: t.id, Kind: t.Kind(), Props: t.props()}
}

// CollectHandlers implements Component.
func (t *Table) CollectHandlers(reg *events.Registry) {
	reg.OnSelect(t.id, func(v any) {
		idx := asInt(v)
		t.selected = idx
		if t.patcher != nil {
			_ = t.patcher.Patch(t.id, map[string]any{"selected": idx})
		}
		if t.onSelect != nil {
			t.onSelect(idx)
		}
	})
	reg.On(t.id, "activate", func(v any) {
		idx := asInt(v)
		t.selected = idx
		if t.patcher != nil {
			_ = t.patcher.Patch(t.id, map[string]any{"selected": idx})
		}
		if t.onActivate != nil {
			t.onActivate(idx)
		}
	})
	reg.On(t.id, "sort", func(v any) {
		m, _ := v.(map[string]any)
		if m == nil || t.onSort == nil {
			return
		}
		column, _ := m["column"].(string)
		ascending, _ := m["ascending"].(bool)
		t.onSort(column, ascending)
	})
}

func (t *Table) reloadFromBind() {
	if t.slicePtr == nil {
		return
	}
	rows, err := sliceToRows(t.slicePtr, t.columns, t.fields)
	if err != nil {
		t.lastErr = err
		return
	}
	t.lastErr = nil
	t.rows = rows
}

func asInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case float32:
		return int(x)
	default:
		return -1
	}
}

func sliceToRows(slicePtr any, columns, fields []string) ([]map[string]any, error) {
	rv := reflect.ValueOf(slicePtr)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return nil, fmt.Errorf("Bind requires non-nil pointer to slice")
	}
	sv := rv.Elem()
	if sv.Kind() != reflect.Slice {
		return nil, fmt.Errorf("Bind requires pointer to slice")
	}
	if len(fields) == 0 {
		fields = columns
	}
	out := make([]map[string]any, 0, sv.Len())
	for i := 0; i < sv.Len(); i++ {
		item := sv.Index(i)
		if item.Kind() == reflect.Pointer {
			if item.IsNil() {
				return nil, fmt.Errorf("row %d is a nil pointer", i)
			}
			item = item.Elem()
		}
		if item.Kind() == reflect.Struct {
			out = append(out, structToRow(item, columns, fields))
			continue
		}
		if item.Kind() == reflect.Map {
			out = append(out, mapToRow(item, columns))
			continue
		}
		return nil, fmt.Errorf("unsupported slice element type %s", item.Kind())
	}
	return out, nil
}

func structToRow(item reflect.Value, columns, fields []string) map[string]any {
	row := make(map[string]any, len(columns))
	typ := item.Type()
	for i, col := range columns {
		fieldName := col
		if i < len(fields) && fields[i] != "" {
			fieldName = fields[i]
		}
		fv := fieldByNameFold(item, typ, fieldName)
		if fv.IsValid() && fv.CanInterface() {
			row[col] = fv.Interface()
		} else {
			row[col] = ""
		}
	}
	return row
}

func fieldByNameFold(item reflect.Value, typ reflect.Type, name string) reflect.Value {
	if f := item.FieldByName(name); f.IsValid() {
		return f
	}
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if strings.EqualFold(sf.Name, name) {
			return item.Field(i)
		}
	}
	return reflect.Value{}
}

func mapToRow(item reflect.Value, columns []string) map[string]any {
	row := make(map[string]any, len(columns))
	for _, col := range columns {
		var found reflect.Value
		for _, key := range item.MapKeys() {
			if strings.EqualFold(fmt.Sprint(key.Interface()), col) {
				found = item.MapIndex(key)
				break
			}
		}
		if found.IsValid() && found.CanInterface() {
			row[col] = found.Interface()
		} else {
			row[col] = ""
		}
	}
	return row
}
