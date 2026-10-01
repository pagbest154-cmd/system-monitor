package hubserver

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleHubBackup(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"detail": "store unavailable"})
		return
	}
	filename := fmt.Sprintf("system-monitor-backup-%s.db", time.Now().UTC().Format("20060102-150405"))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if err := s.Store.WriteBackup(w); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
}

func (s *Server) handlePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	var b strings.Builder
	b.WriteString("# HELP sysmon_up Hub process is running.\n")
	b.WriteString("# TYPE sysmon_up gauge\n")
	b.WriteString("sysmon_up 1\n")
	if s.Store == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(b.String()))
		return
	}
	agents, _ := s.Store.ListAgents()
	b.WriteString("# HELP sysmon_agent_online Agent seen within offline window.\n")
	b.WriteString("# TYPE sysmon_agent_online gauge\n")
	now := float64(time.Now().UnixNano()) / 1e9
	for _, agent := range agents {
		id, _ := agent["id"].(string)
		if id == "" {
			continue
		}
		lastSeen, _ := agent["last_seen"].(float64)
		online := 0
		if now-lastSeen <= float64(180) {
			online = 1
		}
		b.WriteString(fmt.Sprintf("sysmon_agent_online{agent=%q} %d\n", id, online))
	}
	b.WriteString("# HELP sysmon_sensor_value Latest sensor reading.\n")
	b.WriteString("# TYPE sysmon_sensor_value gauge\n")
	for _, agent := range agents {
		id, _ := agent["id"].(string)
		if id == "" {
			continue
		}
		ids, _ := s.Store.ListAgentSensorIDs(id)
		for _, fullID := range ids {
			localID := protocol.StripAgentPrefix(id, fullID)
			latest, err := s.Store.GetLatest(fullID)
			if err != nil || latest == nil {
				continue
			}
			value, ok := latest["value"].(float64)
			if !ok {
				continue
			}
			b.WriteString(fmt.Sprintf(
				"sysmon_sensor_value{agent=%q,sensor=%q} %g\n",
				id, localID, value,
			))
		}
	}
	_, _ = w.Write([]byte(b.String()))
}

func (s *Server) handleListAlertEvents(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"events": []interface{}{}})
		return
	}
	agentID := strings.TrimSpace(r.URL.Query().Get("agent_id"))
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	events, err := s.Store.ListAlertEvents(agentID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent_id": agentID,
		"events":   events,
	})
}

// auditLog writes a single-line audit record for config changes (no secrets).
func auditLog(r *http.Request, action string) {
	role := SessionRole(r)
	user := sessionUserLabel(r)
	log.Printf("[audit] user=%q role=%q action=%q path=%s", user, role, action, r.URL.Path)
}

func sessionUserLabel(r *http.Request) string {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		if _, user := ParseSessionToken(cookie.Value); user != "" {
			return user
		}
	}
	if user, _, ok := ParseBasicAuth(r.Header.Get("Authorization")); ok {
		return user
	}
	return "anonymous"
}
