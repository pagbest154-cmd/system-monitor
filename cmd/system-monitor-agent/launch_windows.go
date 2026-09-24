//go:build windows

package main

import (
	"os"
	"os/exec"
)

func launchSettings(configPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := []string{"--settings"}
	if configPath != "" {
		args = append(args, "--config", configPath)
	}
	// GUI subprocess: do not set HideWindow — it suppresses the settings window.
	return exec.Command(exe, args...).Start()
}
