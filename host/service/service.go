// Package service hosts PicoGoGUI apps as a Windows Service.
//
// Interactive builds should use Desktop + WebUI + Tray; service mode runs
// core + WebUI without a GUI thread.
//
// Note: Windows does not allow a proper system tray inside Session 0.
// Tray/Desktop typically need an interactive companion process.
//
// CLI pattern (from go_service.md):
//
//	service.FixWorkingDir("myapp.exe")
//	if handled, err := service.HandleCommand(cfg, os.Args[1:]); handled { ... }
//	if ran, err := service.MaybeRun(cfg); ran || err != nil { ... }
//	cfg.Run(stop) // interactive fallback
package service

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

// ErrUnsupportedPlatform is returned for SCM operations on non-Windows builds.
var ErrUnsupportedPlatform = errors.New("picogogui/service: Windows services are only supported on Windows")

// Config describes a Windows Service registration/run request.
type Config struct {
	// Name is the service name (e.g. "PicoGoGUIDemo").
	Name string
	// DisplayName is shown in services.msc.
	DisplayName string
	// Description is the service description.
	Description string
	// ExeName is the production executable base name for FixWorkingDir
	// (e.g. "servermanager.exe"). Optional for SCM helpers.
	ExeName string
	// Run is invoked while the service or console mode is running.
	// It must block until stop is closed (or the work finishes).
	Run func(stop <-chan struct{}) error
}

// FixWorkingDir sets the current working directory to the executable location
// in production. Under `go run` (exe name mismatch) the current directory is kept.
func FixWorkingDir(exeName string) {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("picogogui/service: cannot get executable path: %v", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		log.Fatalf("picogogui/service: cannot eval symlinks: %v", err)
	}
	execDir := filepath.Dir(execPath)
	currentExe := strings.ToLower(filepath.Base(execPath))
	if currentExe != strings.ToLower(exeName) {
		log.Printf("[INFO] dev mode detected (%s), keeping current working directory", currentExe)
		return
	}
	if err := os.Chdir(execDir); err != nil {
		log.Fatalf("picogogui/service: cannot chdir to exe dir: %v", err)
	}
	log.Printf("[INFO] working directory set to %s", execDir)
}

// ParseCommand returns the normalized first CLI verb, or "" if none/unknown.
func ParseCommand(args []string) string {
	if len(args) == 0 {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "install", "start", "stop", "remove", "uninstall", "run", "console":
		return strings.ToLower(strings.TrimSpace(args[0]))
	default:
		return ""
	}
}

func runConsole(cfg Config) error {
	if cfg.Run == nil {
		return errors.New("picogogui/service: Config.Run is required")
	}
	stop := make(chan struct{})
	result := make(chan error, 1)
	go func() { result <- cfg.Run(stop) }()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	select {
	case err := <-result:
		return err
	case <-signals:
		close(stop)
		return <-result
	}
}
