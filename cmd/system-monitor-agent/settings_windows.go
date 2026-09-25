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

	var mw *walk.MainWindow
	var hubEdit, agentEdit, tokenEdit *walk.LineEdit
	var intervalEdit *walk.NumberEdit
	var titleLabel, subtitleLabel, notifyLabel *walk.Label
	var lblHub, lblAgent, lblToken, lblInterval *walk.Label
	var saveBtn, cancelBtn *walk.PushButton

	const dlgW, dlgH = 500, 340

	decl := MainWindow{
		AssignTo:   &mw,
		Title:      "system-monitor agent",
		Background: solidBrush(colorBg),
		Font:       Font{Family: "Segoe UI", PointSize: 9},
		Size:       Size{Width: dlgW, Height: dlgH},
		MinSize:    Size{Width: dlgW, Height: dlgH},
		MaxSize:    Size{Width: dlgW, Height: dlgH},
		Layout:     VBox{Margins: Margins{Left: 20, Top: 18, Right: 20, Bottom: 16}, Spacing: 10},
		Children: []Widget{
			Label{AssignTo: &titleLabel, Text: "system-monitor agent"},
			Label{AssignTo: &subtitleLabel, Text: "Подключение к hub · v" + version.Version},
			Composite{
				Layout: VBox{Spacing: 8},
				Children: []Widget{
					Composite{
						Layout: Grid{Columns: 2, Spacing: 10},
						Children: []Widget{
							Label{AssignTo: &lblHub, Text: "Hub URL"},
							LineEdit{AssignTo: &hubEdit, Text: cfg.HubURL},
							Label{AssignTo: &lblAgent, Text: "Agent ID"},
							LineEdit{AssignTo: &agentEdit, Text: cfg.AgentID},
							Label{AssignTo: &lblToken, Text: "Token"},
							LineEdit{AssignTo: &tokenEdit, Text: token, PasswordMode: true},
							Label{AssignTo: &lblInterval, Text: "Интервал (сек)"},
							NumberEdit{AssignTo: &intervalEdit, Value: float64(cfg.IntervalSec), MinValue: 1, MaxValue: 3600},
						},
					},
					Label{AssignTo: &notifyLabel, Text: notifyStatus},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo:  &cancelBtn,
						Text:      "Отмена",
						MinSize:   Size{Width: 96, Height: 32},
						OnClicked: func() { walk.App().Exit(0) },
					},
					PushButton{
						AssignTo: &saveBtn,
						Text:     "Сохранить",
						MinSize:  Size{Width: 110, Height: 32},
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

	if err := decl.Create(); err != nil {
		logSettingsError("settings UI create: " + err.Error())
		return fmt.Errorf("settings UI: %w", err)
	}

	styleHeading(titleLabel)
	styleLabel(subtitleLabel, true)
	styleLabel(notifyLabel, false)
	notifyLabel.SetTextColor(notifyStatusColor(notifyStatus))
	for _, lbl := range []*walk.Label{lblHub, lblAgent, lblToken, lblInterval} {
		styleLabel(lbl, true)
	}
	styleLineEdit(hubEdit)
	styleLineEdit(agentEdit)
	styleLineEdit(tokenEdit)
	styleNumberEdit(intervalEdit)
	stylePushButton(saveBtn, true)
	stylePushButton(cancelBtn, false)

	_, err = mw.Run()
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
