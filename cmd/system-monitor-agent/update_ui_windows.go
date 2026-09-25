//go:build windows

package main

import (
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/pagbest154-cmd/system-monitor/internal/agent"
	"github.com/pagbest154-cmd/system-monitor/internal/release"
)

var updateMu sync.Mutex

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
}

func checkUpdatesFromTray() {
	result := agent.CheckForUpdates(true)
	if result.Error != nil {
		showUpdateResult(result, formatUpdateDialog(result))
		return
	}
	if !result.UpdateAvailable {
		showUpdateResult(result, formatUpdateDialog(result))
		return
	}
	if !updateMu.TryLock() {
		showInfo("system-monitor agent", "Обновление уже выполняется")
		return
	}
	go runBackgroundUpdate(result)
}

func runBackgroundUpdate(result release.ReleaseCheckResult) {
	defer updateMu.Unlock()

	title := "system-monitor agent"
	systray.SetTooltip("Загрузка обновления " + result.LatestVersion + "...")

	if err := agent.ApplyUpdate(result); err != nil {
		systray.SetTooltip("system-monitor agent")
		msg := "Не удалось установить обновление " + result.LatestVersion + ":\n" + err.Error()
		showError(title, msg)
		openUpdateFallback(result)
		return
	}

	systray.SetTooltip("Установка обновления " + result.LatestVersion + "...")
	// Обновление идёт в отдельном elevated-скрипте; закрываем трей, чтобы не блокировать exe.
	time.AfterFunc(800*time.Millisecond, func() { systray.Quit() })
}

func openUpdateFallback(result release.ReleaseCheckResult) {
	url := result.ReleaseURL
	if result.DownloadURL != nil && *result.DownloadURL != "" {
		url = *result.DownloadURL
	}
	if url != "" {
		_ = agent.OpenUpdatePage(url)
	}
}
