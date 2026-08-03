package app

import (
	"errors"
	"testing"

	"github.com/PicoBitSyS/PicoGoGUI/plugin"
)

type failingPlugin struct{}

func (failingPlugin) Info() plugin.Info { return plugin.Info{Name: "failing"} }
func (failingPlugin) Contribute(plugin.Host) error {
	return errors.New("activation failed")
}

func TestRunReturnsPluginActivationError(t *testing.T) {
	plugin.ResetForTest()
	t.Cleanup(plugin.ResetForTest)
	plugin.Register(failingPlugin{})
	application := New()
	if err := application.Run(); err == nil {
		t.Fatal("expected plugin activation error")
	}
}
