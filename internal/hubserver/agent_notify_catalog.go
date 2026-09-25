package hubserver

import (
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/pagbest154-cmd/system-monitor/internal/alerts"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
)

func (s *Server) handleAgentNotifyCatalog(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if !s.requireAgentToken(w, r, agentID) {
		return
	}
	alertsCfg, err := config.LoadAlertsConfig("")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		return
	}
	agentsCfg, _ := config.LoadAgentsConfig("")

	items := make([]protocol.NotifyCatalogItem, 0)
	for id, alertCfg := range alertsCfg.Alerts {
		if !alertCfg.Enabled {
			continue
		}
		topic := strings.TrimSpace(alertCfg.Ntfy.Topic)
		if topic == "" {
			continue
		}
		name := id
		if entry := config.FindAgentEntry(id, agentsCfg); entry != nil && strings.TrimSpace(entry.Name) != "" {
			name = strings.TrimSpace(entry.Name)
		}
		items = append(items, protocol.NotifyCatalogItem{
			AgentID: id,
			Name:    name,
			Enabled: true,
			Topic:   topic,
			Token:   strings.TrimSpace(alertCfg.Ntfy.Token),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Name == items[j].Name {
			return items[i].AgentID < items[j].AgentID
		}
		return items[i].Name < items[j].Name
	})
	writeJSON(w, http.StatusOK, protocol.NotifyCatalogResponse{
		NtfyBaseURL: alerts.NtfyPublicBaseURL(),
		Items:       items,
	})
}
