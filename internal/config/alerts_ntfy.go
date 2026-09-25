package config

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

type NtfyAlertConfig struct {
	Topic string `yaml:"topic" json:"topic"`
	Token string `yaml:"token" json:"token"`
}

var agentIDSlugRe = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func EnsureAgentNtfyTopic(cfg *AgentAlertConfig, agentID string) {
	if cfg == nil || !cfg.Enabled {
		return
	}
	if strings.TrimSpace(cfg.Ntfy.Topic) != "" {
		return
	}
	cfg.Ntfy.Topic = GenerateNtfyTopic(agentID)
}

func GenerateNtfyTopic(agentID string) string {
	slug := agentIDSlugRe.ReplaceAllString(strings.TrimSpace(agentID), "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "host"
	}
	if len(slug) > 32 {
		slug = slug[:32]
	}
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "sysmon-" + slug
	}
	return "sysmon-" + slug + "-" + hex.EncodeToString(buf)
}
