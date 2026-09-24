//go:build windows

package main

import (
	"strings"

	"github.com/pagbest154-cmd/system-monitor/internal/hiddenexec"
)

func serviceRunning() bool {
	out, err := hiddenexec.Command("sc", "query", "system-monitor-agent").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "RUNNING")
}
