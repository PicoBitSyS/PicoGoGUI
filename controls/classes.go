package controls

// Class sets additional CSS classes on the rendered label.
func (l *Label) Class(value string) *Label { l.setClass(value); return l }

// Class sets additional CSS classes on the rendered button.
func (b *Button) Class(value string) *Button { b.setClass(value); return b }

// Class sets additional CSS classes on the rendered text box.
func (t *TextBox) Class(value string) *TextBox { t.setClass(value); return t }

// Class sets additional CSS classes on the rendered number box.
func (n *NumberBox) Class(value string) *NumberBox { n.setClass(value); return n }

// Class sets additional CSS classes on the rendered check box.
func (c *CheckBox) Class(value string) *CheckBox { c.setClass(value); return c }

// Class sets additional CSS classes on the rendered combo box.
func (c *ComboBox) Class(value string) *ComboBox { c.setClass(value); return c }

// Class sets additional CSS classes on the rendered table.
func (t *Table) Class(value string) *Table { t.setClass(value); return t }

// Class sets additional CSS classes on the rendered tree.
func (t *Tree) Class(value string) *Tree { t.setClass(value); return t }

// Class sets additional CSS classes on the rendered design surface.
func (d *DesignSurface) Class(value string) *DesignSurface { d.setClass(value); return d }

// Class sets additional CSS classes on the rendered drop zone.
func (d *DropZone) Class(value string) *DropZone { d.setClass(value); return d }
