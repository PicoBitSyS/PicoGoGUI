package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/PicoBitSyS/PicoGoGUI/host/notify"
	"github.com/PicoBitSyS/PicoGoGUI/host/service"
	"github.com/PicoBitSyS/PicoGoGUI/host/tray"
	"github.com/PicoBitSyS/PicoGoGUI/host/webui"
)

const (
	exeName     = "servermanager.exe"
	serviceName = "PicoGoGUIServerManager"
	listenAddr  = "127.0.0.1:8080"
)

//go:embed web/*
var webFS embed.FS

type statusResponse struct {
	OK      bool   `json:"ok"`
	Service string `json:"service"`
	Addr    string `json:"addr"`
	Uptime  string `json:"uptime"`
}

func main() {
	service.FixWorkingDir(exeName)

	cfg := service.Config{
		Name:        serviceName,
		DisplayName: "PicoGoGUI Server Manager",
		Description: "Demo multi-shell service with HTTP WebUI (no desktop GUI in Session 0).",
		ExeName:     exeName,
		Run:         runCore,
	}

	if handled, err := service.HandleCommand(cfg, os.Args[1:]); handled {
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if ran, err := service.MaybeRun(cfg); ran || err != nil {
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Println("[INFO] interactive mode (WebUI + tray on main thread)")
	if err := runInteractive(cfg); err != nil {
		log.Fatal(err)
	}
}

func runCore(stop <-chan struct{}) error {
	started := time.Now()
	mux := http.NewServeMux()
	static, err := fs.Sub(webFS, "web")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(static)))
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(statusResponse{
			OK:      true,
			Service: serviceName,
			Addr:    listenAddr,
			Uptime:  time.Since(started).Round(time.Second).String(),
		})
	})

	srv := webui.New()
	srv.Addr = listenAddr
	srv.Handler = mux
	if err := srv.Start(); err != nil {
		return err
	}
	log.Printf("[INFO] WebUI listening on http://%s", listenAddr)

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Stop(ctx)
}

func runInteractive(cfg service.Config) error {
	stop := make(chan struct{})
	var once atomic.Bool
	requestStop := func() {
		if once.CompareAndSwap(false, true) {
			close(stop)
		}
	}

	errCh := make(chan error, 1)
	go func() {
		err := cfg.Run(stop)
		errCh <- err
		// Ensure tray unblocks if the core exits first (e.g. port in use).
		tray.Quit()
	}()

	// If WebUI failed immediately (port busy), surface the error without hanging.
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-time.After(300 * time.Millisecond):
		// core still running — continue to tray
	}

	icon := tray.New("PicoGoGUI Server Manager")
	icon.OnOpen = func() {
		_ = openBrowser("http://" + listenAddr + "/")
	}
	icon.OnExit = requestStop
	icon.Add(
		tray.Action("Open Web UI", func() {
			_ = openBrowser("http://" + listenAddr + "/")
		}),
		tray.Action("Notify", func() {
			go func() { _ = notify.Show("Server Manager", "WebUI is running on "+listenAddr) }()
		}),
		tray.Separator(),
		tray.Action("Exit", func() {
			requestStop()
			tray.Quit()
		}),
	)

	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = notify.Show("Server Manager", "Running — open http://"+listenAddr+"/")
	}()

	// Block on main thread (required on Windows). Exit menu / Quit unblocks.
	_ = icon.Run()
	requestStop()

	select {
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Second):
		return nil
	}
}
