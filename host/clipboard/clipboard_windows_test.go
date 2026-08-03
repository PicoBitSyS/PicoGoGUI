//go:build windows

package clipboard

import "testing"

func TestClipboardRoundTrip(t *testing.T) {
	previous, err := GetText()
	if err != nil {
		t.Skipf("clipboard text cannot be safely preserved: %v", err)
	}
	t.Cleanup(func() {
		if err := SetText(previous); err != nil {
			t.Errorf("restore clipboard: %v", err)
		}
	})
	const want = "picogogui-clipboard-test"
	if err := SetText(want); err != nil {
		t.Fatal(err)
	}
	got, err := GetText()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
