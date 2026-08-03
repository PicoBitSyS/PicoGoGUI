//go:build windows

package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// HandleCommand processes install/start/stop/remove/run CLI verbs.
// handled is true when args[0] was a known service command.
func HandleCommand(cfg Config, args []string) (handled bool, err error) {
	cmd := ParseCommand(args)
	if cmd == "" {
		return false, nil
	}
	switch cmd {
	case "install":
		return true, Install(cfg)
	case "start":
		return true, Start(cfg)
	case "stop":
		return true, Stop(cfg)
	case "remove", "uninstall":
		return true, Uninstall(cfg)
	case "run", "console":
		log.Println("[INFO] running in console mode")
		return true, runConsole(cfg)
	default:
		return false, nil
	}
}

// MaybeRun runs the Windows service control handler when started by the SCM.
func MaybeRun(cfg Config) (ran bool, err error) {
	inService, err := svc.IsWindowsService()
	if err != nil {
		return false, err
	}
	if !inService {
		return false, nil
	}
	return true, Run(cfg)
}

// Install registers the current executable as a Windows Service (requires admin).
func Install(cfg Config) error {
	if cfg.Name == "" {
		return fmt.Errorf("picogogui/service: Config.Name is required")
	}
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("eval symlinks: %w", err)
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect SCM: %w (run as Administrator)", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(cfg.Name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists; use remove first", cfg.Name)
	}

	display := cfg.DisplayName
	if display == "" {
		display = cfg.Name
	}
	config := mgr.Config{
		DisplayName:      display,
		Description:      cfg.Description,
		StartType:        mgr.StartAutomatic,
		ServiceStartName: "LocalSystem",
	}
	s, err = m.CreateService(cfg.Name, exePath, config)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	defer s.Close()

	if err := s.SetRecoveryActions([]mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 5 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 10 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 15 * time.Second},
	}, 86400); err != nil {
		fmt.Printf("warning: recovery actions not set: %v\n", err)
	}
	fmt.Printf("service %s installed\n", cfg.Name)
	return nil
}

// Uninstall removes a previously installed Windows Service.
func Uninstall(cfg Config) error {
	if cfg.Name == "" {
		return fmt.Errorf("picogogui/service: Config.Name is required")
	}
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect SCM: %w (run as Administrator)", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(cfg.Name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", cfg.Name)
	}
	defer s.Close()

	status, err := s.Query()
	if err == nil && status.State == svc.Running {
		_, _ = s.Control(svc.Stop)
		time.Sleep(2 * time.Second)
	}
	if err := s.Delete(); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	fmt.Printf("service %s removed\n", cfg.Name)
	return nil
}

// Start starts an installed Windows Service.
func Start(cfg Config) error {
	if cfg.Name == "" {
		return fmt.Errorf("picogogui/service: Config.Name is required")
	}
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(cfg.Name)
	if err != nil {
		return fmt.Errorf("service %s is not installed; run install first", cfg.Name)
	}
	defer s.Close()

	status, err := s.Query()
	if err == nil && status.State == svc.Running {
		fmt.Println("service already running")
		return nil
	}
	if err := s.Start(); err != nil {
		return fmt.Errorf("start service: %w", err)
	}
	fmt.Printf("start issued for %s\n", cfg.Name)
	return nil
}

// Stop stops a running Windows Service.
func Stop(cfg Config) error {
	if cfg.Name == "" {
		return fmt.Errorf("picogogui/service: Config.Name is required")
	}
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(cfg.Name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", cfg.Name)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return fmt.Errorf("query status: %w", err)
	}
	if status.State != svc.Running {
		fmt.Printf("service not running (state %v)\n", status.State)
		return nil
	}
	if _, err := s.Control(svc.Stop); err != nil {
		return fmt.Errorf("stop service: %w", err)
	}
	fmt.Printf("stop issued for %s\n", cfg.Name)
	return nil
}

// Run executes the service control handler (blocking). Prefer MaybeRun from main.
func Run(cfg Config) error {
	if cfg.Name == "" {
		return fmt.Errorf("picogogui/service: Config.Name is required")
	}
	if cfg.Run == nil {
		return fmt.Errorf("picogogui/service: Config.Run is required")
	}
	return svc.Run(cfg.Name, &windowsService{cfg: cfg})
}

type windowsService struct {
	cfg Config
}

func (s *windowsService) Execute(args []string, requests <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	_ = args
	changes <- svc.Status{State: svc.StartPending}

	stop := make(chan struct{})
	var once sync.Once
	requestStop := func() { once.Do(func() { close(stop) }) }

	done := make(chan error, 1)
	go func() {
		done <- s.cfg.Run(stop)
	}()

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case req := <-requests:
			switch req.Cmd {
			case svc.Interrogate:
				changes <- req.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				requestStop()
			}
		case err := <-done:
			if err != nil {
				return false, 1
			}
			return false, 0
		}
	}
}
