package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestListAgentsMarksStaleOffline(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "metrics.db")
	store, err := NewMetricStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	stale := float64(time.Now().UnixNano())/1e9 - AgentOfflineSec - 10
	if err := store.UpsertAgent(AgentUpsert{
		AgentID: "stale", Name: "Stale", Status: "online", LastSeen: stale,
	}); err != nil {
		t.Fatal(err)
	}
	agents, err := store.ListAgents()
	if err != nil || len(agents) != 1 {
		t.Fatalf("agents %v err=%v", agents, err)
	}
	if agents[0]["status"] != "offline" {
		t.Fatalf("expected offline for stale agent, got %v", agents[0]["status"])
	}
}

func TestTouchAgentRefreshesOnline(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "metrics.db")
	store, err := NewMetricStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	stale := float64(time.Now().UnixNano())/1e9 - AgentOfflineSec - 5
	if err := store.UpsertAgent(AgentUpsert{
		AgentID: "a1", Status: "offline", LastSeen: stale,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.TouchAgent("a1", "online"); err != nil {
		t.Fatal(err)
	}
	agents, err := store.ListAgents()
	if err != nil || len(agents) != 1 {
		t.Fatal(err)
	}
	if agents[0]["status"] != "online" {
		t.Fatalf("expected online after touch, got %v", agents[0]["status"])
	}
}
