//go:build !windows

package main

import "fmt"

func openBrowser(url string) error {
	return fmt.Errorf("open browser unsupported: %s", url)
}
