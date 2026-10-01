//go:build debinstall_integration

package agentinstall

import (
	"os"
	"testing"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
)

// TestDebInstalledConfigOnDisk is run from scripts/test-deb-agent-install-config.sh after dpkg -i.
func TestDebInstalledConfigOnDisk(t *testing.T) {
	agentYAML := os.Getenv("AGENT_YAML")
	tokenFile := os.Getenv("TOKEN_FILE")
	wantHub := os.Getenv("HUB_URL")
	wantID := os.Getenv("AGENT_ID")
	wantToken := os.Getenv("AGENT_TOKEN")
	if agentYAML == "" || tokenFile == "" || wantHub == "" || wantID == "" {
		t.Fatal("set AGENT_YAML, TOKEN_FILE, HUB_URL, AGENT_ID (and AGENT_TOKEN) for integration run")
	}

	cfg, err := config.LoadAgentConfig(agentYAML)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HubURL != wantHub {
		t.Fatalf("hub_url: %q want %q", cfg.HubURL, wantHub)
	}
	if cfg.AgentID != wantID {
		t.Fatalf("agent_id: %q want %q", cfg.AgentID, wantID)
	}
	if config.LoadAgentToken(cfg) != wantToken {
		t.Fatal("token mismatch via LoadAgentToken")
	}
	raw, err := os.ReadFile(tokenFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != wantToken {
		t.Fatalf("token file: %q", raw)
	}
}
