package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
)

type NotifySubscriptionsFile struct {
	Agents []string `yaml:"agents" json:"agents"`
}

func NotifySubscriptionsPath() string {
	return filepath.Join(paths.ConfigDir, "notify_subscriptions.yaml")
}

func LoadNotifySubscriptions() ([]string, error) {
	path := NotifySubscriptionsPath()
	cfg := &NotifySubscriptionsFile{}
	if err := decodeYAML(path, cfg); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]string, 0, len(cfg.Agents))
	for _, id := range cfg.Agents {
		id = strings.TrimSpace(id)
		if id != "" {
			out = append(out, id)
		}
	}
	return out, nil
}

func SaveNotifySubscriptions(agentIDs []string) error {
	path := NotifySubscriptionsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	unique := make([]string, 0, len(agentIDs))
	seen := map[string]bool{}
	for _, id := range agentIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}
	return saveYAML(path, &NotifySubscriptionsFile{Agents: unique})
}
