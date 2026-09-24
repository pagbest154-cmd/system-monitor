//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/getlantern/systray"
	"github.com/pagbest154-cmd/system-monitor/internal/agent"
	"github.com/pagbest154-cmd/system-monitor/internal/branding"
	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/version"
)

func runTray(configPath string) error {
	systray.Run(func() { onTrayReady(configPath) }, func() {})
	return nil
}

func onTrayReady(configPath string) {
	systray.SetTitle("system-monitor")
	systray.SetTooltip("system-monitor agent")
	systray.SetIcon(trayIcon("idle"))

	mSettings := systray.AddMenuItem("Настройки", "Открыть настройки")
	mRestart := systray.AddMenuItem("Перезапустить службу", "")
	mLog := systray.AddMenuItem("Открыть лог", "")
	mConfig := systray.AddMenuItem("Открыть папку конфигурации", "")
	systray.AddSeparator()
	systray.AddMenuItem("Версия "+version.Version, "").Disable()
	mUpdate := systray.AddMenuItem("Проверить обновления", "")
	mQuit := systray.AddMenuItem("Выход", "")

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			refreshTrayIcon()
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
			case <-mUpdate.ClickedCh:
				go checkUpdatesFromTray()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func refreshTrayIcon() {
	status, _ := agent.ReadStatus()
	iconKey := "idle"
	if status != nil {
		if status.Connected && !status.IsStale() {
			iconKey = "ok"
		} else if status.LastError != nil {
			iconKey = "error"
		}
		tooltip := status.AgentID
		if status.Connected {
			tooltip = "Подключён: " + tooltip
		} else {
			tooltip = "Ошибка: " + tooltip
		}
		systray.SetTooltip(tooltip)
	}
	systray.SetIcon(trayIcon(iconKey))
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
	return exec.Command("sc", "stop", "system-monitor-agent").Run()
}
