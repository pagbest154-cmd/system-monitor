//go:build !windows

package main

import (
	"github.com/pagbest154-cmd/system-monitor/internal/agent"
	"github.com/pagbest154-cmd/system-monitor/internal/release"
)

func formatUpdateDialog(result release.ReleaseCheckResult) string {
	msg := agent.FormatUpdateMessage(result)
	if result.Error != nil {
		msg += "\n\n" + *result.Error
	}
	return msg
}

func showUpdateResult(result release.ReleaseCheckResult, msg string) {}

func checkUpdatesFromTray() {}

func showInfo(title, message string) {}

func showError(title, message string) {}
