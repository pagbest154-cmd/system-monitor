package enrich

import (
	"testing"

	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
)

func TestEnrichAgentReportPreservesVersion(t *testing.T) {
	report := &protocol.AgentReport{
		AgentID:      "vpn",
		Hostname:     "ams-1-vm-nyc1",
		AgentVersion: "1.0.36",
		Platform:     "linux",
		System: map[string]interface{}{
			"cpu":     map[string]interface{}{"percent": 12.5},
			"memory":  map[string]interface{}{"percent": 40.0},
		},
	}
	enriched := EnrichAgentReport(report)
	if enriched.AgentVersion != "1.0.36" {
		t.Fatalf("agent_version lost: %q", enriched.AgentVersion)
	}
	if enriched.Platform != "linux" {
		t.Fatalf("platform lost: %q", enriched.Platform)
	}
}
