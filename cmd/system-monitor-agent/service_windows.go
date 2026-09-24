//go:build windows

package main

import (
	"os"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
)

// Служба пишет agent.status.json каждые interval_sec; sc query каждые 3 с давал мигание консоли.
func serviceRunning() bool {
	info, err := os.Stat(paths.AgentStatusFile)
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) < 30*time.Second
}
