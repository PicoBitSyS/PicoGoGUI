// Command picogogui provides development tooling for PicoGoGUI applications.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

var buildSequence atomic.Uint64

func main() {
	if len(os.Args) < 2 || os.Args[1] != "run" {
		fmt.Fprintln(os.Stderr, "usage: picogogui run [package] [-- app arguments]")
		os.Exit(2)
	}
	if err := run(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "picogogui:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	target := "."
	var appArgs []string
	for index, arg := range args {
		if arg == "--" {
			appArgs = append(appArgs, args[index+1:]...)
			break
		}
		if index == 0 {
			target = arg
		}
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	changes, err := snapshot(root)
	if err != nil {
		return err
	}
	var process *exec.Cmd
	var executable string

	restart := func() {
		next, buildErr := build(target)
		if buildErr != nil {
			fmt.Fprintln(os.Stderr, buildErr)
			return
		}
		stopProcess(process)
		if executable != "" {
			_ = os.Remove(executable)
		}
		executable = next
		process = exec.Command(executable, appArgs...)
		process.Stdout = os.Stdout
		process.Stderr = os.Stderr
		process.Stdin = os.Stdin
		if err := process.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "start:", err)
			process = nil
			return
		}
		fmt.Printf("[picogogui] running %s\n", target)
	}

	restart()
	defer func() {
		stopProcess(process)
		if executable != "" {
			_ = os.Remove(executable)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-signals:
			return nil
		case <-ticker.C:
			next, snapErr := snapshot(root)
			if snapErr != nil {
				fmt.Fprintln(os.Stderr, "watch:", snapErr)
				continue
			}
			if !sameSnapshot(changes, next) {
				changes = next
				fmt.Println("[picogogui] changes detected; rebuilding")
				restart()
			}
		}
	}
}

func build(target string) (string, error) {
	extension := ""
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}
	output := filepath.Join(os.TempDir(), fmt.Sprintf(
		"picogogui-dev-%d-%d%s", os.Getpid(), buildSequence.Add(1), extension))
	command := exec.Command("go", "build", "-o", output, target)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		_ = os.Remove(output)
		return "", fmt.Errorf("build failed: %w", err)
	}
	return output, nil
}

func stopProcess(command *exec.Cmd) {
	if command == nil || command.Process == nil {
		return
	}
	_ = command.Process.Kill()
	_ = command.Wait()
}

type fileState struct {
	modTime int64
	size    int64
}

func snapshot(root string) (map[string]fileState, error) {
	state := make(map[string]fileState)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch strings.ToLower(entry.Name()) {
			case ".git", "bin", "node_modules":
				if path != root {
					return filepath.SkipDir
				}
			}
			return nil
		}
		switch strings.ToLower(filepath.Ext(entry.Name())) {
		case ".go", ".html", ".css", ".js", ".json":
		default:
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		state[path] = fileState{modTime: info.ModTime().UnixNano(), size: info.Size()}
		return nil
	})
	return state, err
}

func sameSnapshot(left, right map[string]fileState) bool {
	if len(left) != len(right) {
		return false
	}
	for path, state := range left {
		if right[path] != state {
			return false
		}
	}
	return true
}
