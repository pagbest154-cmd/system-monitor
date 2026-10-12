package agent

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
)

func TestTransportPushReportUsesBearerAndAgentURL(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotBody protocol.AgentReport
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	}))
	defer srv.Close()

	transport := NewTransport(srv.URL+"/", "agent-token")
	v := 1.0
	err := transport.PushReport(&protocol.AgentReport{
		AgentID: "host-a",
		Metrics: []protocol.MetricPoint{{SensorID: "cpu_percent", Value: &v, Status: "ok"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/agents/host-a/metrics" {
		t.Fatalf("unexpected path %q", gotPath)
	}
	if gotAuth != "Bearer agent-token" {
		t.Fatalf("unexpected auth %q", gotAuth)
	}
	if gotBody.AgentID != "host-a" {
		t.Fatalf("unexpected agent id %q", gotBody.AgentID)
	}
}

func TestTransportPushReportHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "not found")
	}))
	defer srv.Close()

	transport := NewTransport(srv.URL, "t")
	err := transport.PushReport(&protocol.AgentReport{AgentID: "x"})
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected 404 error, got %v", err)
	}
}
