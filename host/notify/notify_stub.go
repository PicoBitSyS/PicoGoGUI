//go:build !windows

package notify

func show(string, string) error { return ErrUnsupportedPlatform }
