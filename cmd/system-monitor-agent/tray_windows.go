//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/getlantern/systray"
	"github.com/pagbest154-cmd/system-monitor/internal/agent"
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
				_ = runSettingsDialog(configPath)
			case <-mRestart.ClickedCh:
				_ = restartService()
			case <-mLog.ClickedCh:
				_ = openPath(filepath.Join(paths.ConfigDir, "agent.log"))
			case <-mConfig.ClickedCh:
				_ = openPath(paths.ConfigDir)
			case <-mUpdate.ClickedCh:
				result := agent.CheckForUpdates(true)
				fmt.Println(agent.FormatUpdateMessage(result))
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
	switch key {
	case "ok":
		return greenIconPNG
	case "error":
		return redIconPNG
	default:
		return grayIconPNG
	}
}

func openPath(path string) error {
	return exec.Command("explorer", path).Start()
}

func restartService() error {
	return exec.Command("sc", "stop", "system-monitor-agent").Run()
}

var greenIconPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0xf3, 0xff,
	0x61, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x60, 0x00, 0x02, 0x00,
	0x00, 0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44,
	0xae, 0x42, 0x60, 0x82,
}
var redIconPNG = greenIconPNG
var grayIconPNG = greenIconPNG
