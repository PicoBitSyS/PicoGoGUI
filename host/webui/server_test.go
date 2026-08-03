package webui

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestServerStaticAndStop(t *testing.T) {
	s := New()
	s.Addr = "127.0.0.1:0"
	s.Static = fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("hello-webui")},
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = s.Stop(ctx)
	}()

	resp, err := http.Get(s.URL() + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "hello-webui") {
		t.Fatalf("body = %q", body)
	}
}
