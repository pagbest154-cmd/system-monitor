package hubserver

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pagbest154-cmd/system-monitor/internal/alerts"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
)

func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadAlertsConfig("")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"alerts":        cfg.Alerts,
		"ntfy_base_url": alerts.NtfyBaseURL(),
	})
}

func (s *Server) handleGetAgentAlerts(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	cfg, err := config.LoadAlertsConfig("")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		return
	}
	agentCfg, ok := cfg.Alerts[agentID]
	if !ok {
		agentCfg = config.DefaultAgentAlertConfig()
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent_id":      agentID,
		"config":        agentCfg,
		"ntfy_base_url": alerts.NtfyBaseURL(),
	})
}

func (s *Server) handlePutAgentAlerts(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	var agentCfg config.AgentAlertConfig
	if err := json.NewDecoder(r.Body).Decode(&agentCfg); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": err.Error()})
		return
	}
	config.EnsureAgentNtfyTopic(&agentCfg, agentID)
	cfg, err := config.LoadAlertsConfig("")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		return
	}
	if cfg.Alerts == nil {
		cfg.Alerts = map[string]config.AgentAlertConfig{}
	}
	cfg.Alerts[agentID] = agentCfg
	if err := config.SaveAlertsConfig(cfg, ""); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent_id":      agentID,
		"config":        agentCfg,
		"ntfy_base_url": alerts.NtfyBaseURL(),
	})
}

func (s *Server) handleRegenerateNtfyTopic(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	cfg, err := config.LoadAlertsConfig("")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		return
	}
	if cfg.Alerts == nil {
		cfg.Alerts = map[string]config.AgentAlertConfig{}
	}
	agentCfg, ok := cfg.Alerts[agentID]
	if !ok {
		agentCfg = config.DefaultAgentAlertConfig()
	}
	agentCfg.Ntfy.Topic = config.GenerateNtfyTopic(agentID)
	cfg.Alerts[agentID] = agentCfg
	if err := config.SaveAlertsConfig(cfg, ""); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent_id":      agentID,
		"config":        agentCfg,
		"ntfy_base_url": alerts.NtfyBaseURL(),
	})
}
