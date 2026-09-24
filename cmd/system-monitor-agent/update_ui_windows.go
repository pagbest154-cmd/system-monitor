//go:build windows

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

func showUpdateResult(result release.ReleaseCheckResult, msg string) {
	title := "system-monitor agent"
	if result.Error != nil {
		showError(title, msg)
		return
	}
	showInfo(title, msg)
	if result.UpdateAvailable {
		url := result.ReleaseURL
		if result.DownloadURL != nil && *result.DownloadURL != "" {
			url = *result.DownloadURL
		}
		if url != "" {
			_ = agent.OpenUpdatePage(url)
		}
	}
}

func checkUpdatesFromTray() {
	result := agent.CheckForUpdates(true)
	msg := formatUpdateDialog(result)
	showUpdateResult(result, msg)
}
