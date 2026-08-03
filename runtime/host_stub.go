//go:build !windows

package runtime

// NewHost returns ErrUnsupportedPlatform on non-Windows builds.
func NewHost(Options) (*Host, error) {
	return nil, ErrUnsupportedPlatform
}
