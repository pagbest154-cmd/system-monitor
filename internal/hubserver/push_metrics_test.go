package hubserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
)

func TestPushMetricsUpdatesAgentOnline(t *testing.T) {
	dir := t.TempDir()
	paths.OverrideForTest(dir)
	if err := os.MkdirAll(filepath.Dir(paths.AgentsConfig), 0o755); err != nil {
		t.Fatal(err)
	}
	agentsYAML := "agents:\n  - id: homepc\n    name: Home\n    token: test-token-secret\n"
	if err := os.WriteFile(paths.AgentsConfig, []byte(agentsYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := NewServer("hub")
	if err != nil {
		t.Fatal(err)
	}
	handler := s.Router()

	v := 55.0
	body, _ := json.Marshal(protocol.AgentReport{
		AgentID: "homepc",
		Metrics: []protocol.MetricPoint{
			{SensorID: "cpu_percent", Value: &v, Status: "ok", TS: 1_700_000_100},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/agents/homepc/metrics", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-token-secret")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("push metrics status %d: %s", rec.Code, rec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/agents", nil)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list agents status %d: %s", listRec.Code, listRec.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(listRec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	agents, _ := payload["agents"].([]interface{})
	if len(agents) != 1 {
		t.Fatalf("expected one agent, got %v", payload)
	}
	agent, _ := agents[0].(map[string]interface{})
	if agent["status"] != "online" {
		t.Fatalf("expected online after push, got %v", agent)
	}

	wrongReq := httptest.NewRequest(http.MethodPost, "/api/agents/unknown/metrics", bytes.NewReader(body))
	wrongReq.Header.Set("Authorization", "Bearer test-token-secret")
	wrongRec := httptest.NewRecorder()
	handler.ServeHTTP(wrongRec, wrongReq)
	if wrongRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unknown agent, got %d", wrongRec.Code)
	}
}
