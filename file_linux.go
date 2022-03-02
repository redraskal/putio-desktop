//go:build linux

package main

import (
	"path/filepath"
)

func displayFile(path string) error {
	return exec.Command("xdg-open", filepath.Dir(path)).Run()
}
