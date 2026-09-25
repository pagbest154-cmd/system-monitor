package storage

import (
	"path/filepath"
	"testing"
)

func TestMetricStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "metrics.db")
	store, err := NewMetricStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	v := 42.0
	if err := store.Insert("cpu_percent", &v, "ok", 1000); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetLatest("cpu_percent")
	if err != nil || latest == nil {
		t.Fatal("expected latest value")
	}
	if err := store.UpsertAgent(AgentUpsert{
		AgentID: "host1", Name: "Host", Hostname: "host1",
		AgentVersion: "1.0.35", Platform: "linux",
	}); err != nil {
		t.Fatal(err)
	}
	agents, err := store.ListAgents()
	if err != nil || len(agents) != 1 {
		t.Fatalf("expected one agent, got %v", agents)
	}
	if agents[0]["agent_version"] != "1.0.35" || agents[0]["platform"] != "linux" {
		t.Fatalf("unexpected agent metadata: %v", agents[0])
	}
	if err := store.Insert("host1:cpu_percent", &v, "ok", 1001); err != nil {
		t.Fatal(err)
	}
	if err := store.SetAlertState("host1", "cpu_percent", true, 1001); err != nil {
		t.Fatal(err)
	}
	found, metricsDeleted, err := store.DeleteAgent("host1")
	if err != nil || !found || metricsDeleted != 1 {
		t.Fatalf("DeleteAgent: found=%v metrics=%d err=%v", found, metricsDeleted, err)
	}
	agents, err = store.ListAgents()
	if err != nil || len(agents) != 0 {
		t.Fatalf("expected no agents after delete, got %v", agents)
	}
	latest, err = store.GetLatest("host1:cpu_percent")
	if err != nil || latest != nil {
		t.Fatalf("expected metric history removed, got %v", latest)
	}
}
