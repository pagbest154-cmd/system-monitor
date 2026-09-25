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
	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
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

	ownID := fleet.DefaultAgentID(cfg.AgentID)
	selected, _ := config.LoadNotifySubscriptions()
	selectedSet := map[string]bool{}
	for _, id := range selected {
		selectedSet[id] = true
	}
	if len(selectedSet) == 0 {
		selectedSet[ownID] = true
	}

	var catalog protocol.NotifyCatalogResponse
	var catalogErr error
	if token != "" {
		transport := agent.NewTransport(cfg.HubURL, token)
		catalog, catalogErr = transport.FetchNotifyCatalog(ownID)
	}

	notifyStatus := fetchNotifyStatus(cfg, token)

	var mw *walk.MainWindow
	var hubEdit, agentEdit, tokenEdit *walk.LineEdit
	var intervalEdit *walk.NumberEdit
	var notifyGroup *walk.GroupBox
	notifyChecks := map[string]*walk.CheckBox{}

	const dlgW, dlgH = 480, 520

	decl := MainWindow{
		AssignTo: &mw,
		Title:    "Настройки " + branding.AgentName,
		Font:     Font{Family: "Segoe UI", PointSize: 9},
		Size:     Size{Width: dlgW, Height: dlgH},
		MinSize:  Size{Width: dlgW, Height: dlgH},
		MaxSize:  Size{Width: dlgW, Height: dlgH},
		Layout:   VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}},
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
			GroupBox{
				AssignTo: &notifyGroup,
				Title:    "Уведомления — хосты",
				Layout:   VBox{Spacing: 6},
			},
			Label{Text: notifyStatus},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "Тест уведомления",
						OnClicked: func() {
							if err := showTestToast(); err != nil {
								showError(branding.AgentName, "Не удалось показать уведомление:\n"+err.Error())
							}
						},
					},
					HSpacer{},
				},
			},
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
						Text:    "Сохранить",
						MaxSize: Size{Width: 100, Height: 0},
						OnClicked: func() {
							cfg.HubURL = hubEdit.Text()
							cfg.AgentID = agentEdit.Text()
							cfg.IntervalSec = int(intervalEdit.Value())
							_ = config.SaveAgentConfig(cfg, configPath)
							_ = config.SaveAgentToken(tokenEdit.Text(), "")
							ids := make([]string, 0, len(notifyChecks))
							for id, cb := range notifyChecks {
								if cb.Checked() {
									ids = append(ids, id)
								}
							}
							_ = config.SaveNotifySubscriptions(ids)
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

	if notifyGroup != nil {
		if catalogErr != nil {
			lbl, _ := walk.NewLabel(notifyGroup)
			lbl.SetText("Список хостов недоступен.\nОбновите hub до последней версии.")
		} else if len(catalog.Items) == 0 {
			lbl, _ := walk.NewLabel(notifyGroup)
			lbl.SetText("Нет хостов с включёнными алертами.\nВключите алерты на странице Хосты.")
		} else {
			for _, item := range catalog.Items {
				cb, err := walk.NewCheckBox(notifyGroup)
				if err != nil {
					continue
				}
				label := item.Name
				if label == "" {
					label = item.AgentID
				}
				if label != item.AgentID {
					label = fmt.Sprintf("%s (%s)", label, item.AgentID)
				}
				cb.SetText(label)
				cb.SetChecked(selectedSet[item.AgentID])
				notifyChecks[item.AgentID] = cb
			}
		}
	}

	if err := mw.Run(); err != nil {
		logSettingsError("settings UI: " + err.Error())
		return fmt.Errorf("settings UI: %w", err)
	}
	return nil
}

func fetchNotifyStatus(cfg *config.AgentFileConfig, token string) string {
	if token == "" {
		return "Уведомления: укажите token и сохраните настройки"
	}
	ownID := fleet.DefaultAgentID(cfg.AgentID)
	selected, _ := config.LoadNotifySubscriptions()
	if len(selected) == 0 {
		selected = []string{ownID}
	}
	transport := agent.NewTransport(cfg.HubURL, token)
	catalog, err := transport.FetchNotifyCatalog(ownID)
	if err != nil {
		return "Уведомления: hub не отвечает (нужен hub 1.0.32+)"
	}
	enabled := 0
	byID := map[string]bool{}
	for _, item := range catalog.Items {
		if item.Enabled && item.Topic != "" {
			byID[item.AgentID] = true
			enabled++
		}
	}
	if enabled == 0 {
		return "Уведомления: выключены — включите алерты на странице Хосты"
	}
	subscribed := 0
	for _, id := range selected {
		if byID[id] {
			subscribed++
		}
	}
	return fmt.Sprintf("Уведомления: %d из %d хостов · %s", subscribed, enabled, catalog.NtfyBaseURL)
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
