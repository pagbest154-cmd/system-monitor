//go:build windows

package main

import (
	"os/exec"
	"strings"
)

func serviceRunning() bool {
	out, err := exec.Command("sc", "query", "system-monitor-agent").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "RUNNING")
}
