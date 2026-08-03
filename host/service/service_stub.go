//go:build !windows

package service

import "log"

// HandleCommand processes known verbs; SCM verbs error on non-Windows.
func HandleCommand(cfg Config, args []string) (handled bool, err error) {
	cmd := ParseCommand(args)
	if cmd == "" {
		return false, nil
	}
	switch cmd {
	case "install", "start", "stop", "remove", "uninstall":
		return true, ErrUnsupportedPlatform
	case "run", "console":
		log.Println("[INFO] running in console mode")
		return true, runConsole(cfg)
	default:
		return false, nil
	}
}

// MaybeRun is a no-op outside Windows.
func MaybeRun(Config) (bool, error) { return false, nil }

// Run returns ErrUnsupportedPlatform outside Windows.
func Run(Config) error { return ErrUnsupportedPlatform }

// Install returns ErrUnsupportedPlatform outside Windows.
func Install(Config) error { return ErrUnsupportedPlatform }

// Start returns ErrUnsupportedPlatform outside Windows.
func Start(Config) error { return ErrUnsupportedPlatform }

// Stop returns ErrUnsupportedPlatform outside Windows.
func Stop(Config) error { return ErrUnsupportedPlatform }

// Uninstall returns ErrUnsupportedPlatform outside Windows.
func Uninstall(Config) error { return ErrUnsupportedPlatform }
