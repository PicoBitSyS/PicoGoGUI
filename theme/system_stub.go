//go:build !windows

package theme

// detectSystem defaults to Light on non-Windows builds.
func detectSystem() Name { return Light }
