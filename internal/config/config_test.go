package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
)

func TestMergeAgentConfig(t *testing.T) {
	base := &SensorsFile{
		Settings: defaultSettings(),
		Sensors: []SensorConfig{
			{ID: "cpu_percent", Name: "CPU", Type: "system.cpu_percent", Enabled: true},
		},
	}
	disabled := false
	merged := MergeAgentConfig(base, []SensorOverrideConfig{
		{SensorID: "cpu_percent", Enabled: &disabled},
	})
	if merged.Sensors[0].Enabled {
		t.Fatal("expected sensor disabled")
	}
}

func TestPrepareAgentsForSave(t *testing.T) {
	dir := t.TempDir()
	paths.OverrideForTest(dir)
	agentsPath := filepath.Join(dir, "config", "agents.yaml")
	_ = os.MkdirAll(filepath.Dir(agentsPath), 0o755)
	incoming := &AgentsFile{Agents: []AgentEntry{{ID: "a1", Name: "A", Token: "change-me"}}}
	prepared, err := PrepareAgentsForSave(incoming)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Agents[0].Token == "" || prepared.Agents[0].Token == "change-me" {
		t.Fatal("expected generated token")
	}
}
