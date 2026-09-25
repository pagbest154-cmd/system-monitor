package hubserver

import (
	"net/http"
	"testing"
)

func TestIsPublicPath(t *testing.T) {
	cases := []struct {
		path   string
		method string
		want   bool
	}{
		{"/api/version", http.MethodGet, true},
		{"/api/agents/homepc/metrics", http.MethodPost, false},
		{"/api/agents/homepc/notify/catalog", http.MethodGet, false},
		{"/api/dashboard", http.MethodGet, false},
	}
	for _, tc := range cases {
		if got := IsPublicPath(tc.path, tc.method); got != tc.want {
			t.Errorf("IsPublicPath(%q, %s) = %v, want %v", tc.path, tc.method, got, tc.want)
		}
	}
}

func TestAgentPathInfo(t *testing.T) {
	cases := []struct {
		path     string
		method   string
		agentID  string
		wantOK   bool
	}{
		{"/api/agents/homepc/metrics", http.MethodPost, "homepc", true},
		{"/api/agents/homepc/notify", http.MethodGet, "homepc", true},
		{"/api/agents/homepc/notify/catalog", http.MethodGet, "homepc", true},
		{"/api/agents/homepc/metrics", http.MethodGet, "", false},
		{"/api/agents/homepc/system", http.MethodGet, "", false},
		{"/api/agents/homepc/notify/catalog", http.MethodPost, "", false},
	}
	for _, tc := range cases {
		agentID, ok := agentPathInfo(tc.path, tc.method)
		if ok != tc.wantOK || agentID != tc.agentID {
			t.Errorf("agentPathInfo(%q, %s) = (%q, %v), want (%q, %v)", tc.path, tc.method, agentID, ok, tc.agentID, tc.wantOK)
		}
	}
}
