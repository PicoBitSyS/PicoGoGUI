//go:build !windows

package tray

func quit() {}

// Run returns ErrUnsupportedPlatform outside Windows.
func (i *Icon) Run() error { return ErrUnsupportedPlatform }

// Start returns ErrUnsupportedPlatform outside Windows.
func (i *Icon) Start() error { return ErrUnsupportedPlatform }

// Stop is a no-op outside Windows.
func (i *Icon) Stop() error { return nil }
