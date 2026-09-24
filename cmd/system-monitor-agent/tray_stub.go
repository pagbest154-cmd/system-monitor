//go:build !windows

package main

import "fmt"

func runTray(configPath string) error {
	return fmt.Errorf("--tray доступен только на Windows")
}

func runSettings(configPath string) error {
	return fmt.Errorf("--settings доступен только на Windows")
}
