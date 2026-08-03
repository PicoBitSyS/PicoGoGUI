package window

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	minimumPersistedWidth  = 240
	minimumPersistedHeight = 180
	maximumPersistedWidth  = 7680
	maximumPersistedHeight = 4320
)

// PersistSize restores an outer logical size from path and saves later native
// resizes when the window closes normally.
func (w *Window) PersistSize(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("picogogui: persistence path is required")
	}
	w.mu.Lock()
	w.persistPath = path
	w.mu.Unlock()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var size Size
	if err := json.Unmarshal(data, &size); err != nil {
		return err
	}
	if !validPersistedSize(size) {
		return errors.New("picogogui: persisted window size is invalid")
	}
	w.SetOuterSize(size.Width, size.Height)
	return nil
}

func (w *Window) savePersistedSize() error {
	w.mu.Lock()
	path := w.persistPath
	size := w.lastOuter
	if !validPersistedSize(size) {
		size = w.outerSize
	}
	w.mu.Unlock()
	if path == "" || !validPersistedSize(size) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(size, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0600)
}

func validPersistedSize(size Size) bool {
	return size.Width >= minimumPersistedWidth && size.Width <= maximumPersistedWidth &&
		size.Height >= minimumPersistedHeight && size.Height <= maximumPersistedHeight
}
