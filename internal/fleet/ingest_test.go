package fleet

import (
	"path/filepath"
	"testing"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
	"github.com/pagbest154-cmd/system-monitor/internal/storage"
)

func TestIngestAgentReportMarksAgentOnline(t *testing.T) {
	dir := t.TempDir()
	paths.OverrideForTest(dir)
	store, err := storage.NewMetricStore(paths.DBPath)
	if err != nil {
		t.Fatal(err)
	}
	v := 12.5
	report := &protocol.AgentReport{
		AgentID:  "win-pc",
		Hostname: "win-pc",
		Metrics: []protocol.MetricPoint{
			{SensorID: "cpu_percent", Value: &v, Status: "ok", TS: 1_700_000_000},
		},
	}
	result := IngestAgentReport(report, store, nil)
	if result["status"] != "ok" {
		t.Fatalf("unexpected ingest result: %v", result)
	}
	agents, err := store.ListAgents()
	if err != nil || len(agents) != 1 {
		t.Fatalf("agents: %v err=%v", agents, err)
	}
	if agents[0]["status"] != "online" {
		t.Fatalf("expected online, got %v", agents[0]["status"])
	}
	lastSeen, _ := agents[0]["last_seen"].(float64)
	if lastSeen <= 0 {
		t.Fatalf("expected last_seen set, got %v", agents[0]["last_seen"])
	}
}

func TestVerifyAgentTokenRequiresConfig(t *testing.T) {
	if VerifyAgentToken("missing", "secret") {
		t.Fatal("expected false without agents config entry")
	}
}
