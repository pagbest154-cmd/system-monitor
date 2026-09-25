package config

import (
	"github.com/pagbest154-cmd/system-monitor/internal/paths"
)

type AlertOfflineRule struct {
	Enabled   bool `yaml:"enabled" json:"enabled"`
	AfterSec  int  `yaml:"after_sec" json:"after_sec"`
}

type AlertSensorRule struct {
	SensorID  string  `yaml:"sensor_id" json:"sensor_id"`
	Enabled   bool    `yaml:"enabled" json:"enabled"`
	Threshold float64 `yaml:"threshold" json:"threshold"`
}

type AgentAlertConfig struct {
	Enabled        bool              `yaml:"enabled" json:"enabled"`
	Offline        AlertOfflineRule  `yaml:"offline" json:"offline"`
	Sensors        []AlertSensorRule `yaml:"sensors" json:"sensors"`
	CooldownSec    int               `yaml:"cooldown_sec" json:"cooldown_sec"`
	NotifyRecovery bool              `yaml:"notify_recovery" json:"notify_recovery"`
	Ntfy           NtfyAlertConfig   `yaml:"ntfy" json:"ntfy"`
}

type AlertsFile struct {
	Alerts map[string]AgentAlertConfig `yaml:"alerts" json:"alerts"`
}

func LoadAlertsConfig(path string) (*AlertsFile, error) {
	if path == "" {
		path = paths.AlertsConfig
	}
	cfg := &AlertsFile{Alerts: map[string]AgentAlertConfig{}}
	if err := decodeYAML(path, cfg); err != nil {
		return nil, err
	}
	if cfg.Alerts == nil {
		cfg.Alerts = map[string]AgentAlertConfig{}
	}
	return cfg, nil
}

func SaveAlertsConfig(cfg *AlertsFile, path string) error {
	if path == "" {
		path = paths.AlertsConfig
	}
	if cfg.Alerts == nil {
		cfg.Alerts = map[string]AgentAlertConfig{}
	}
	return saveYAML(path, cfg)
}

func DefaultAgentAlertConfig() AgentAlertConfig {
	return AgentAlertConfig{
		Enabled: false,
		Offline: AlertOfflineRule{Enabled: false, AfterSec: 180},
		Sensors: []AlertSensorRule{},
		CooldownSec: 900,
		NotifyRecovery: false,
	}
}
