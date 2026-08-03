package service

import (
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	cases := map[string]string{
		"install":   "install",
		"INSTALL":   "install",
		"uninstall": "uninstall",
		"remove":    "remove",
		"run":       "run",
		"console":   "console",
		"foo":       "",
		"":          "",
	}
	for in, want := range cases {
		args := []string{}
		if in != "" {
			args = []string{in}
		}
		if got := ParseCommand(args); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestHandleCommandUnknown(t *testing.T) {
	handled, err := HandleCommand(Config{Name: "x"}, []string{"not-a-cmd"})
	if handled || err != nil {
		t.Fatalf("handled=%v err=%v", handled, err)
	}
}

func TestRunRequiresConfig(t *testing.T) {
	err := Run(Config{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Name") && err != ErrUnsupportedPlatform {
		// On Windows Name required; on stub unsupported
		t.Log(err)
	}
}
