//go:build !windows

package main

import (
	"os"
	"os/exec"
)

func disableEcho() string {
	// save and disable echo
	cmd := exec.Command("stty", "-echo")
	cmd.Stdin = os.Stdin
	cmd.Run()
	return ""
}

func restoreEcho(_ string) {
	cmd := exec.Command("stty", "echo")
	cmd.Stdin = os.Stdin
	cmd.Run()
}
