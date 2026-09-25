//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/pagbest154-cmd/system-monitor/internal/agent"
	"github.com/pagbest154-cmd/system-monitor/internal/branding"
	"github.com/pagbest154-cmd/system-monitor/internal/hiddenexec"
	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/version"
)

func runTray(configPath string) error {
	systray.Run(func() { onTrayReady(configPath) }, func() {})
	return nil
}

func onTrayReady(configPath string) {
	systray.SetTitle(branding.AgentShort)
	systray.SetTooltip(branding.AgentName)
	systray.SetIcon(trayIcon("idle"))

	mStatus := systray.AddMenuItem(trayStatusText(nil, serviceRunning()), "")
	mStatus.Disable()
	systray.AddSeparator()

	mSettings := systray.AddMenuItem("Настройки "+branding.AgentName, "Открыть настройки")
	mRestart := systray.AddMenuItem("Перезапустить службу", "")
	mLog := systray.AddMenuItem("Открыть лог", "")
	mConfig := systray.AddMenuItem("Открыть папку конфигурации", "")
	mTrayLog := systray.AddMenuItem("Открыть лог трея", "")
	mPush := systray.AddMenuItem(currentPushNotifyStatus(), "")
	mPush.Disable()
	registerPushNotifyMenu(func(title string) { mPush.SetTitle(title) })
	systray.AddSeparator()
	systray.AddMenuItem("Версия "+version.Version, "").Disable()
	mUpdate := systray.AddMenuItem("Проверить обновления", "")
	mQuit := systray.AddMenuItem("Выход", "")

	startNtfyListener(configPath)

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			refreshTrayIcon(mStatus)
		}
	}()

	go func() {
		for {
			select {
			case <-mSettings.ClickedCh:
				go func() {
					if err := launchSettings(configPath); err != nil {
						systray.SetTooltip("Настройки: " + err.Error())
					}
				}()
			case <-mRestart.ClickedCh:
				_ = restartService()
			case <-mLog.ClickedCh:
				_ = openPath(filepath.Join(paths.ConfigDir, "agent.log"))
			case <-mConfig.ClickedCh:
				_ = openPath(paths.ConfigDir)
			case <-mTrayLog.ClickedCh:
				_ = openPath(filepath.Join(paths.ConfigDir, "tray.log"))
			case <-mUpdate.ClickedCh:
				go checkUpdatesFromTray()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func refreshTrayIcon(mStatus *systray.MenuItem) {
	status, _ := agent.ReadStatus()
	running := serviceRunning()
	mStatus.SetTitle(trayStatusText(status, running))

	iconKey := "idle"
	if status != nil {
		if status.Connected && !status.IsStale() {
			iconKey = "ok"
		} else if status.LastError != nil {
			iconKey = "error"
		}
	}
	systray.SetTooltip(trayTooltipText(status, running))
	systray.SetIcon(trayIcon(iconKey))
}

func trayStatusText(status *agent.Status, running bool) string {
	if !running {
		return "Служба: не запущена"
	}
	if status == nil {
		return "Статус: ожидание"
	}
	if status.Connected && !status.IsStale() {
		if status.AgentID != "" {
			return "Статус: подключён (" + status.AgentID + ")"
		}
		return "Статус: подключён"
	}
	if status.LastError != nil && *status.LastError != "" {
		return "Статус: ошибка"
	}
	return "Статус: ожидание"
}

func trayTooltipText(status *agent.Status, running bool) string {
	if !running {
		return branding.AgentName + " — служба не запущена"
	}
	if status == nil {
		return branding.AgentName + " — ожидание данных"
	}
	parts := []string{branding.AgentName}
	if status.Connected && !status.IsStale() {
		parts = append(parts, "подключён")
	} else if status.LastError != nil && *status.LastError != "" {
		parts = append(parts, *status.LastError)
	} else {
		parts = append(parts, "ожидание")
	}
	if status.AgentID != "" {
		parts = append(parts, status.AgentID)
	}
	if status.HubURL != "" {
		parts = append(parts, status.HubURL)
	}
	text := strings.Join(parts, " — ")
	if len(text) > 127 {
		return text[:126] + "…"
	}
	return text
}

func trayIcon(key string) []byte {
	if data := branding.TrayICO(key); len(data) > 0 {
		return data
	}
	if data := loadInstalledIcon(); len(data) > 0 {
		return data
	}
	return branding.TrayPNG(key)
}

func loadInstalledIcon() []byte {
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(exe), "app-icon.ico"))
	if err != nil {
		return nil
	}
	return data
}

func openPath(path string) error {
	return exec.Command("explorer", path).Start()
}

func restartService() error {
	return hiddenexec.Command("sc", "stop", "system-monitor-agent").Run()
}
