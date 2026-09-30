package config

import (
	"os"
	"strings"
	"testing"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
)

func TestSaveAgentSettingsWritesTokenToYAMLAndFile(t *testing.T) {
	dir := t.TempDir()
	paths.OverrideForTest(dir)
	configPath := paths.AgentConfig
	cfg := &AgentFileConfig{
		HubURL:      "http://hub.test",
		AgentID:     "pc1",
		IntervalSec: 5,
		Transport:   "http",
		TokenFile:   paths.AgentTokenFile,
	}
	if err := SaveAgentSettings(cfg, configPath, "secret-token"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "secret-token") {
		t.Fatalf("agent.yaml missing token: %s", raw)
	}
	fileToken, err := os.ReadFile(paths.AgentTokenFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(fileToken)) != "secret-token" {
		t.Fatalf("agent.token: %q", fileToken)
	}
	loaded, err := LoadAgentConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if LoadAgentToken(loaded) != "secret-token" {
		t.Fatal("LoadAgentToken after save")
	}
}
