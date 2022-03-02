//go:build darwin

package main

import "os/exec"

func displayFile(path string) error {
	return exec.Command("open", "-R", path).Run()
}
