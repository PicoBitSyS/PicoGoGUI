// Package webui hosts a separate HTTP/Web UI channel.
//
// The Web UI does NOT use the declarative gui component tree. It serves
// developer-supplied handlers and/or static files that talk to the same
// Go business logic as the desktop shell.
package webui

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"sync"
)

// ErrNotRunning is returned when Stop is called before Start.
var ErrNotRunning = errors.New("picogogui/webui: server is not running")

// Server is a lightweight HTTP host for the Web UI channel.
type Server struct {
	mu sync.Mutex

	// Addr is the listen address (default "127.0.0.1:8080").
	Addr string
	// Handler is the root HTTP handler. If nil, Static (when set) or
	// a default mux is used.
	Handler http.Handler
	// Static optional filesystem served at "/".
	Static fs.FS

	server   *http.Server
	listener net.Listener
}

// New creates a Web UI server with defaults.
func New() *Server {
	return &Server{Addr: "127.0.0.1:8080"}
}

// URL returns the base URL once listening, or the configured Addr URL.
func (s *Server) URL() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	addr := s.Addr
	if s.listener != nil {
		addr = s.listener.Addr().String()
	}
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	return "http://" + addr
}

func (s *Server) handler() http.Handler {
	if s.Handler != nil {
		return s.Handler
	}
	mux := http.NewServeMux()
	if s.Static != nil {
		mux.Handle("/", http.FileServer(http.FS(s.Static)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("PicoGoGUI Web UI channel\n"))
		})
	}
	return mux
}

// Start begins listening. It does not block; serve runs in a goroutine.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.server != nil {
		return nil
	}
	addr := s.Addr
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: s.handler()}
	s.listener = ln
	s.server = srv
	s.Addr = ln.Addr().String()
	go func() { _ = srv.Serve(ln) }()
	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	srv := s.server
	s.server = nil
	s.listener = nil
	s.mu.Unlock()
	if srv == nil {
		return ErrNotRunning
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return srv.Shutdown(ctx)
}
