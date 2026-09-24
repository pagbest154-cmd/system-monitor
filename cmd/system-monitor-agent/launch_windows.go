//go:build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
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
	cmd := exec.Command(exe, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}
