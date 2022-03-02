//go:build windows

package main

import "os/exec"

func displayFile(path string) error {
	return exec.Command("explorer.exe", "/select,"+path).Run()
}
