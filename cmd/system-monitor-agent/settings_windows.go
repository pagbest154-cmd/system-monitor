//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/paths"
)

func runSettings(configPath string) error {
	return runSettingsDialog(configPath)
}

func runSettingsDialog(configPath string) error {
	cfg, err := config.LoadAgentConfig(configPath)
	if err != nil {
		logSettingsError("LoadAgentConfig: " + err.Error())
		return err
	}
	token := config.LoadAgentToken(cfg)

	var hubEdit, agentEdit, tokenEdit *walk.LineEdit
	var intervalEdit *walk.NumberEdit

	if cfg.IntervalSec <= 0 {
		cfg.IntervalSec = 5
	}

	window := MainWindow{
		Title:   "Настройки агента system-monitor",
		MinSize: Size{Width: 520, Height: 280},
		Layout:  VBox{},
		Children: []Widget{
			GroupBox{
				Title:  "Подключение",
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Hub URL:"},
					LineEdit{AssignTo: &hubEdit, Text: cfg.HubURL},
					Label{Text: "Agent ID:"},
					LineEdit{AssignTo: &agentEdit, Text: cfg.AgentID},
					Label{Text: "Token:"},
					LineEdit{AssignTo: &tokenEdit, Text: token, PasswordMode: true},
					Label{Text: "Интервал (сек):"},
					NumberEdit{AssignTo: &intervalEdit, Value: float64(cfg.IntervalSec), MinValue: 1, MaxValue: 3600},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "Сохранить",
						OnClicked: func() {
							cfg.HubURL = hubEdit.Text()
							cfg.AgentID = agentEdit.Text()
							cfg.IntervalSec = int(intervalEdit.Value())
							_ = config.SaveAgentConfig(cfg, configPath)
							_ = config.SaveAgentToken(tokenEdit.Text(), "")
							walk.App().Exit(0)
						},
					},
					PushButton{
						Text: "Отмена",
						OnClicked: func() {
							walk.App().Exit(0)
						},
					},
				},
			},
		},
	}
	if icon := settingsIconPath(); icon != "" {
		window.Icon = icon
	}

	_, err = window.Run()
	if err != nil {
		logSettingsError("settings UI: " + err.Error())
		return fmt.Errorf("settings UI: %w", err)
	}
	return nil
}

func logSettingsError(msg string) {
	_ = os.MkdirAll(paths.ConfigDir, 0755)
	path := filepath.Join(paths.ConfigDir, "settings.err.log")
	line := fmt.Sprintf("%s  %s\n", time.Now().Format(time.RFC3339), msg)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line)
	_ = f.Close()
}

func settingsIconPath() string {
	exe, err := os.Executable()
	if err == nil {
		iconPath := filepath.Join(filepath.Dir(exe), "app-icon.ico")
		if _, err := os.Stat(iconPath); err == nil {
			return iconPath
		}
	}
	return ""
}
