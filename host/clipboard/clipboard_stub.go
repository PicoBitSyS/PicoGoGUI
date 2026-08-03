//go:build !windows

package clipboard

func getText() (string, error) { return "", ErrUnsupportedPlatform }
func setText(string) error     { return ErrUnsupportedPlatform }
