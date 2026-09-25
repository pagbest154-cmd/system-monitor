//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/pagbest154-cmd/system-monitor/internal/agent"
	"github.com/pagbest154-cmd/system-monitor/internal/branding"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/fleet"
	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/version"
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
	if cfg.IntervalSec <= 0 {
		cfg.IntervalSec = 5
	}
	if exe, err := os.Executable(); err == nil {
		_ = os.Chdir(filepath.Dir(exe))
	}

	notifyStatus := fetchNotifyStatus(cfg, token)

	var hubEdit, agentEdit, tokenEdit *walk.LineEdit
	var intervalEdit *walk.NumberEdit
	const dlgW, dlgH = 460, 280

	window := MainWindow{
		Title:   "Настройки " + branding.AgentName,
		Font:    Font{Family: "Segoe UI", PointSize: 9},
		Size:    Size{Width: dlgW, Height: dlgH},
		MinSize: Size{Width: dlgW, Height: dlgH},
		MaxSize: Size{Width: dlgW, Height: dlgH},
		Layout:  VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}},
		Children: []Widget{
			Label{Text: branding.AgentName + " · v" + version.Version},
			GroupBox{
				Title:  "Подключение",
				Layout: Grid{Columns: 2, Spacing: 10},
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
			Label{Text: notifyStatus},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text:      "Отмена",
						MaxSize:   Size{Width: 100, Height: 0},
						OnClicked: func() { walk.App().Exit(0) },
					},
					PushButton{
						Text: "Сохранить",
						MaxSize: Size{Width: 100, Height: 0},
						OnClicked: func() {
							cfg.HubURL = hubEdit.Text()
							cfg.AgentID = agentEdit.Text()
							cfg.IntervalSec = int(intervalEdit.Value())
							_ = config.SaveAgentConfig(cfg, configPath)
							_ = config.SaveAgentToken(tokenEdit.Text(), "")
							walk.App().Exit(0)
						},
					},
				},
			},
		},
	}

	_, err = window.Run()
	if err != nil {
		logSettingsError("settings UI: " + err.Error())
		return fmt.Errorf("settings UI: %w", err)
	}
	return nil
}

func fetchNotifyStatus(cfg *config.AgentFileConfig, token string) string {
	if token == "" {
		return "Уведомления: укажите token и сохраните настройки"
	}
	agentID := fleet.DefaultAgentID(cfg.AgentID)
	transport := agent.NewTransport(cfg.HubURL, token)
	notify, err := transport.FetchNotifyConfig(agentID)
	if err != nil {
		return "Уведомления: hub не отвечает (обновите hub до 1.0.18+)"
	}
	if !notify.Enabled || notify.Topic == "" {
		return "Уведомления: выключены — включите алерты на странице Хосты"
	}
	return "Уведомления: включены · " + notify.NtfyBaseURL
}

func logSettingsError(msg string) {
	_ = os.MkdirAll(paths.ConfigDir, 0o755)
	path := filepath.Join(paths.ConfigDir, "settings.err.log")
	line := fmt.Sprintf("%s  %s\n", time.Now().Format(time.RFC3339), msg)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line)
	_ = f.Close()
}
