//go:build windows

package main

import (
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
)

func runSettings(configPath string) error {
	return runSettingsDialog(configPath)
}

func runSettingsDialog(configPath string) error {
	cfg, err := config.LoadAgentConfig(configPath)
	if err != nil {
		return err
	}
	token := config.LoadAgentToken(cfg)

	var hubEdit, agentEdit, tokenEdit *walk.LineEdit
	var intervalEdit *walk.NumberEdit

	_, err = MainWindow{
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
	}.Run()
	if err != nil {
		return fmt.Errorf("settings UI: %w", err)
	}
	return nil
}
