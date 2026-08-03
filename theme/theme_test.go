package theme

import "testing"

func TestResolveConcrete(t *testing.T) {
	if Resolve(Dark) != Dark {
		t.Fatal("dark")
	}
	if Resolve(Light) != Light {
		t.Fatal("light")
	}
}

func TestResolveSystem(t *testing.T) {
	got := Resolve(System)
	if got != Dark && got != Light {
		t.Fatalf("system resolved to %q", got)
	}
}

func TestIsDark(t *testing.T) {
	if !IsDark(Dark) {
		t.Fatal("expected dark")
	}
	if IsDark(Light) {
		t.Fatal("expected light")
	}
}

func TestCustomTheme(t *testing.T) {
	name, err := Register(Definition{
		Name: "ocean", Dark: true, Background: "#001122", Accent: "#00aaff",
	})
	if err != nil {
		t.Fatal(err)
	}
	if Resolve(name) != name || !IsDark(name) {
		t.Fatalf("custom theme not resolved: %q", Resolve(name))
	}
	vars := Variables(name)
	if vars["--pico-bg"] != "#001122" || vars["--pico-accent"] != "#00aaff" {
		t.Fatalf("variables = %#v", vars)
	}
}
