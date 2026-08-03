//go:build !windows

package win32

// SetDPIAware is a no-op on non-Windows platforms.
func SetDPIAware() {}
