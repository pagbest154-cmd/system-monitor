package hubserver

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/pagbest154-cmd/system-monitor/internal/alerts"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
)

func (s *Server) handleAgentNotify(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if !s.requireAgentToken(w, r, agentID) {
		return
	}
	resp := protocol.AgentNotifyResponse{
		NtfyBaseURL: alerts.NtfyPublicBaseURL(),
	}
	cfg, err := config.LoadAlertsConfig("")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		return
	}
	agentCfg, ok := cfg.Alerts[agentID]
	if !ok || !agentCfg.Enabled {
		writeJSON(w, http.StatusOK, resp)
		return
	}
	topic := strings.TrimSpace(agentCfg.Ntfy.Topic)
	if topic == "" {
		writeJSON(w, http.StatusOK, resp)
		return
	}
	resp.Enabled = true
	resp.Topic = topic
	resp.Token = strings.TrimSpace(agentCfg.Ntfy.Token)
	writeJSON(w, http.StatusOK, resp)
}
