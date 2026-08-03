package tray

import "testing"

func TestMenuBuilder(t *testing.T) {
	icon := New("Demo").Add(
		Action("Open", func() {}),
		Separator(),
		Action("Exit", func() {}),
	)
	if icon.Tooltip != "Demo" || len(icon.Menu) != 3 {
		t.Fatalf("icon = %+v", icon)
	}
}
