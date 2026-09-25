package hubserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"

	"github.com/pagbest154-cmd/system-monitor/internal/alerts"
	"github.com/pagbest154-cmd/system-monitor/internal/collector"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/enrich"
	"github.com/pagbest154-cmd/system-monitor/internal/fleet"
	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
	"github.com/pagbest154-cmd/system-monitor/internal/release"
	"github.com/pagbest154-cmd/system-monitor/internal/storage"
	"github.com/pagbest154-cmd/system-monitor/internal/sysinfo"
	"github.com/pagbest154-cmd/system-monitor/internal/version"
)

type Server struct {
	Mode          string
	Store         *storage.MetricStore
	Hub           *LiveHub
	Fleet         *FleetState
	Collector     *collector.Collector
	RetentionDays int
	AlertEngine   *alerts.Engine
}

func NewServer(mode string) (*Server, error) {
	if err := os.MkdirAll(paths.DataDir, 0o755); err != nil {
		return nil, err
	}
	store, err := storage.NewMetricStore(paths.DBPath)
	if err != nil {
		return nil, err
	}
	sender := alerts.NewNtfySenderFromEnv()
	s := &Server{
		Mode: mode, Store: store, Hub: NewLiveHub(), Fleet: NewFleetState(),
		RetentionDays: 31,
		AlertEngine:   alerts.NewEngine(store, sender),
	}
	if mode == "hub" {
		s.AlertEngine.StartOfflineLoop()
	}
	if mode == "standalone" {
		cfg, _ := config.LoadSensorsConfig("")
		if cfg != nil {
			s.RetentionDays = cfg.Settings.RetentionDays
		}
		s.Collector = collector.NewCollector(store, func(data map[string]interface{}) {
			s.Hub.ScheduleBroadcast(map[string]interface{}{"type": "update", "data": data})
		}, cfg, nil)
	}
	go s.retentionLoop()
	return s, nil
}

func (s *Server) retentionLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if s.Mode == "hub" || s.Collector != nil {
			_, _ = s.Store.Cleanup(s.RetentionDays)
		}
	}
}

func (s *Server) StartCollector() {
	if s.Collector != nil {
		s.Collector.Start()
	}
}

func (s *Server) StopCollector() {
	if s.Collector != nil {
		s.Collector.Stop()
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	if s.Mode == "hub" && HubAuthEnabled() {
		r.Use(s.authMiddleware)
	}
	r.Get("/api/mode", s.handleMode)
	r.Get("/api/version", s.handleVersion)
	r.Get("/api/sensors", s.handleSensors)
	r.Get("/api/metrics/*", s.handleMetrics)
	r.Get("/api/system", s.handleSystem)
	r.Get("/api/agents", s.handleListAgents)
	r.Get("/api/agents/{agentID}", s.handleGetAgent)
	r.Get("/api/agents/{agentID}/system", s.handleAgentSystem)
	r.Post("/api/agents/{agentID}/metrics", s.handlePushMetrics)
	r.Post("/api/agents/{agentID}/heartbeat", s.handleHeartbeat)
	r.Post("/api/agents/{agentID}/config", s.handleSyncConfig)
	r.Get("/api/agents/{agentID}/notify", s.handleAgentNotify)
	r.Get("/api/agents/{agentID}/notify/catalog", s.handleAgentNotifyCatalog)
	r.Get("/api/dashboard", s.handleDashboard)
	r.Put("/api/config/sensors", s.handleUpdateSensors)
	r.Put("/api/config/dashboard", s.handleUpdateDashboard)
	r.Get("/api/config/agents", s.handleGetAgentsConfig)
	r.Put("/api/config/agents", s.handleUpdateAgentsConfig)
	r.Get("/api/config/hub", s.handleGetHubConfig)
	r.Put("/api/config/hub", s.handleUpdateHubConfig)
	r.Get("/api/hub/info", s.handleHubInfo)
	r.Get("/api/auth/status", s.handleAuthStatus)
	r.Post("/api/auth/login", s.handleLogin)
	r.Post("/api/auth/logout", s.handleLogout)
	r.Get("/api/sensor-types", s.handleSensorTypes)
	r.Get("/api/alerts", s.handleListAlerts)
	r.Get("/api/alerts/{agentID}", s.handleGetAgentAlerts)
	r.Put("/api/alerts/{agentID}", s.handlePutAgentAlerts)
	r.Post("/api/alerts/{agentID}/ntfy/topic", s.handleRegenerateNtfyTopic)
	r.Get("/ws/live", s.handleLiveWS)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(paths.WebDir))))
	r.Get("/", s.serveHTML("index.html"))
	r.Get("/settings", s.serveHTML("settings.html"))
	r.Get("/hosts", s.serveHTML("hosts.html"))
	r.Get("/login", s.serveHTML("login.html"))
	return r
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.Mode != "hub" || !HubAuthEnabled() || IsPublicPath(r.URL.Path, r.Method) {
			next.ServeHTTP(w, r)
			return
		}
		if IsAuthenticated(r) {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"detail": "Требуется авторизация"})
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func (s *Server) serveHTML(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(paths.WebDir, name))
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"mode": s.Mode})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	result := release.CheckReleaseUpdates(version.Version, paths.HubUpdateCache, false, "system-monitor-hub", release.HubVersionCheckInterval, nil)
	payload := map[string]interface{}{
		"current_version":  result.CurrentVersion,
		"latest_version":   result.LatestVersion,
		"update_available": result.UpdateAvailable,
		"release_url":      result.ReleaseURL,
		"error":            result.Error,
	}
	if s.Mode == "hub" && result.UpdateAvailable {
		payload["update_hint"] = fmt.Sprintf(
			"docker pull ghcr.io/pagbest154-cmd/system-monitor:%s && docker compose up -d",
			result.LatestVersion,
		)
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleSensors(w http.ResponseWriter, r *http.Request) {
	agent := r.URL.Query().Get("agent")
	if agent != "" {
		payload, ok := s.sensorsForAgent(agent)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"detail": "Агент не найден"})
			return
		}
		writeJSON(w, http.StatusOK, payload)
		return
	}
	if s.Collector == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"settings": map[string]int{"retention_days": s.RetentionDays},
			"sensors":  []interface{}{},
		})
		return
	}
	configs := s.Collector.GetSensorConfigs()
	latest := s.Collector.GetLatestSnapshot()
	active := s.Collector.GetActiveSensorIDs()
	items := make([]map[string]interface{}, 0, len(configs))
	for _, cfg := range configs {
		supported := cfg.IsSupportedOnPlatform()
		items = append(items, map[string]interface{}{
			"id": cfg.ID, "name": cfg.Name, "type": cfg.Type, "enabled": cfg.Enabled,
			"interval_sec": cfg.IntervalSec, "unit": cfg.Unit, "params": cfg.Params,
			"platforms": cfg.Platforms, "warn_above": cfg.WarnAbove, "critical_above": cfg.CriticalAbove,
			"supported": supported, "available": supported && active[cfg.ID],
			"current": latest[cfg.ID],
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"settings": s.Collector.GetSettings(),
		"sensors":  items,
	})
}

func (s *Server) sensorsForAgent(agent string) (map[string]interface{}, bool) {
	record, _ := s.Store.GetAgent(agent)
	if record == nil {
		return nil, false
	}
	latest := s.Fleet.GetAgentSnapshot(agent)
	system, _ := record["system"].(map[string]interface{})
	metas := make([]map[string]interface{}, 0)
	if sensors, ok := record["sensors"].([]interface{}); ok && len(sensors) > 0 {
		for _, item := range sensors {
			if m, ok := item.(map[string]interface{}); ok {
				metas = append(metas, m)
			}
		}
	} else if system != nil {
		metas = enrich.SensorMetasFromSystem(system)
	} else {
		ids, _ := s.Store.ListAgentSensorIDs(agent)
		for _, id := range ids {
			metas = append(metas, map[string]interface{}{"id": id, "name": id, "type": "unknown", "unit": "%"})
		}
	}
	items := make([]map[string]interface{}, 0, len(metas))
	for _, meta := range metas {
		sensorID := fmt.Sprint(meta["id"])
		fullID := protocol.PrefixedSensorID(agent, sensorID)
		current := latest[fullID]
		if current == nil || current["value"] == nil {
			stored, _ := s.Store.GetLatest(fullID)
			if stored != nil && stored["value"] != nil {
				current = map[string]interface{}{"sensor_id": fullID}
				for k, v := range stored {
					current[k] = v
				}
				latest[fullID] = current
			}
		}
		if (current == nil || current["value"] == nil) && system != nil {
			fromSystem := enrich.ReadingFromSystem(system, sensorID, 0)
			if fromSystem != nil {
				current = fromSystem
				current["sensor_id"] = fullID
				latest[fullID] = current
			}
		}
		item := map[string]interface{}{}
		for k, v := range meta {
			item[k] = v
		}
		item["id"] = sensorID
		item["full_id"] = fullID
		item["supported"] = true
		item["available"] = true
		item["current"] = current
		items = append(items, item)
	}
	if len(latest) > 0 {
		s.Fleet.UpdateAgent(agent, latest)
	}
	return map[string]interface{}{
		"settings": map[string]int{"retention_days": s.RetentionDays},
		"sensors":  items,
		"agent_id": agent,
	}, true
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	sensorID := strings.TrimPrefix(r.URL.Path, "/api/metrics/")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "1h"
	}
	agent := r.URL.Query().Get("agent")
	if agent != "" {
		sensorID = protocol.PrefixedSensorID(agent, sensorID)
	}
	points, _ := s.Store.GetHistory(sensorID, period)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sensor_id": sensorID, "period": period, "points": points,
	})
}

func (s *Server) handleSystem(w http.ResponseWriter, r *http.Request) {
	agent := r.URL.Query().Get("agent")
	if agent != "" {
		record, _ := s.Store.GetAgent(agent)
		if record == nil || record["system"] == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"detail": "Данные системы агента недоступны"})
			return
		}
		writeJSON(w, http.StatusOK, record["system"])
		return
	}
	if s.Mode == "hub" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": "Укажите agent=<id> в hub-режиме"})
		return
	}
	writeJSON(w, http.StatusOK, sysinfo.GetSystemInfo())
}

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	agents, _ := s.Store.ListAgents()
	agentRelease := release.CheckReleaseUpdates(
		version.Version,
		paths.HubAgentLatestCache,
		false,
		"system-monitor-hub",
		release.HubAgentVersionCheckInterval,
		nil,
	)
	latest := agentRelease.LatestVersion
	releaseURL := agentRelease.ReleaseURL
	for _, agent := range agents {
		current, _ := agent["agent_version"].(string)
		if current != "" && latest != "" {
			agent["update_available"] = release.IsNewerVersion(latest, current)
		}
	}
	payload := map[string]interface{}{
		"agents":               agents,
		"latest_agent_version": latest,
		"agent_release_url":    releaseURL,
	}
	if latest != "" {
		payload["agent_deb_url"] = release.ReleaseDownloadURL("", latest, release.AgentDebAssetName(latest))
		payload["agent_windows_url"] = release.ReleaseDownloadURL("", latest, release.AgentWindowsSetupAssetName(latest))
		payload["agent_apt_command"] = "sudo apt update && sudo apt install --only-upgrade system-monitor-agent"
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	record, _ := s.Store.GetAgent(agentID)
	if record == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"detail": "Агент не найден"})
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *Server) handleAgentSystem(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	record, _ := s.Store.GetAgent(agentID)
	if record == nil || record["system"] == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"detail": "Данные системы агента недоступны"})
		return
	}
	writeJSON(w, http.StatusOK, record["system"])
}

func (s *Server) requireAgentToken(w http.ResponseWriter, r *http.Request, agentID string) bool {
	token := ""
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		token = strings.TrimSpace(auth[7:])
	}
	if !fleet.VerifyAgentToken(agentID, token) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"detail": "Неверный токен агента"})
		return false
	}
	return true
}

func (s *Server) handlePushMetrics(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if !s.requireAgentToken(w, r, agentID) {
		return
	}
	var report protocol.AgentReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": err.Error()})
		return
	}
	if report.AgentID != agentID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": "agent_id в URL и теле не совпадают"})
		return
	}
	enriched := enrich.EnrichAgentReport(&report)
	result := fleet.IngestAgentReport(enriched, s.Store, s.Hub)
	latest := map[string]map[string]interface{}{}
	for _, point := range enriched.Metrics {
		if point.Value == nil {
			continue
		}
		fullID := protocol.PrefixedSensorID(agentID, point.SensorID)
		latest[fullID] = map[string]interface{}{
			"sensor_id": fullID, "value": point.Value, "status": point.Status, "ts": point.TS,
		}
	}
	s.Fleet.UpdateAgent(agentID, latest)
	s.evaluatePushAlerts(agentID, enriched)
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) evaluatePushAlerts(agentID string, report *protocol.AgentReport) {
	if s.AlertEngine == nil || report == nil {
		return
	}
	agentName := agentID
	entry := config.FindAgentEntry(agentID, nil)
	if entry != nil && entry.Name != "" {
		agentName = entry.Name
	}
	sensorNames := map[string]string{}
	sensorUnits := map[string]string{}
	for _, sensor := range report.Sensors {
		sensorNames[sensor.ID] = sensor.Name
		sensorUnits[sensor.ID] = sensor.Unit
	}
	for _, point := range report.Metrics {
		if point.Value == nil {
			continue
		}
		name := sensorNames[point.SensorID]
		if name == "" {
			name = point.SensorID
		}
		s.AlertEngine.EvaluateMetric(alerts.MetricInput{
			AgentID:    agentID,
			AgentName:  agentName,
			SensorID:   point.SensorID,
			SensorName: name,
			Unit:       sensorUnits[point.SensorID],
			Value:      *point.Value,
		})
	}
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if !s.requireAgentToken(w, r, agentID) {
		return
	}
	_ = s.Store.TouchAgent(agentID, "online")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSyncConfig(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if !s.requireAgentToken(w, r, agentID) {
		return
	}
	configVersion := 0
	if v := r.URL.Query().Get("config_version"); v != "" {
		_, _ = fmt.Sscanf(v, "%d", &configVersion)
	}
	resp := fleet.BuildAgentConfigResponse(agentID, configVersion)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadDashboardConfig("")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleUpdateSensors(w http.ResponseWriter, r *http.Request) {
	if s.Mode == "hub" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": "Датчики настраиваются на агентах"})
		return
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": err.Error()})
		return
	}
	current, _ := config.LoadSensorsConfig("")
	raw, _ := json.Marshal(current)
	var data map[string]interface{}
	_ = json.Unmarshal(raw, &data)
	if body["settings"] != nil {
		data["settings"] = body["settings"]
	}
	if body["sensors"] != nil {
		data["sensors"] = body["sensors"]
	}
	buf, _ := json.Marshal(data)
	cfg := &config.SensorsFile{}
	_ = json.Unmarshal(buf, cfg)
	_ = config.SaveSensorsConfig(cfg, "")
	if s.Collector != nil {
		s.Collector.ReloadConfig(cfg)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleUpdateDashboard(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": err.Error()})
		return
	}
	current, _ := config.LoadDashboardConfig("")
	if body["dashboard"] != nil {
		raw, _ := json.Marshal(body["dashboard"])
		_ = json.Unmarshal(raw, &current.Dashboard)
	}
	if body["panels"] != nil {
		defaults := map[string][]string{}
		for _, panel := range current.Panels {
			if len(panel.Sensors) > 0 {
				defaults[panel.ID] = panel.Sensors
			}
		}
		panelsRaw, _ := json.Marshal(body["panels"])
		var panels []config.PanelConfig
		_ = json.Unmarshal(panelsRaw, &panels)
		for i, panel := range panels {
			if len(panel.Sensors) == 0 && defaults[panel.ID] != nil {
				panels[i].Sensors = defaults[panel.ID]
			}
		}
		current.Panels = config.EnrichDashboardPanels(panels)
	}
	_ = config.SaveDashboardConfig(current, "")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetAgentsConfig(w http.ResponseWriter, r *http.Request) {
	cfg, _ := config.LoadAgentsConfig("")
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleUpdateAgentsConfig(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Agents []config.AgentEntry `json:"agents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": err.Error()})
		return
	}
	incoming := &config.AgentsFile{Agents: body.Agents}
	prepared, _ := config.PrepareAgentsForSave(incoming)
	_ = config.SaveAgentsConfig(prepared, "")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok", "agents": prepared.Agents,
	})
}

func (s *Server) handleGetHubConfig(w http.ResponseWriter, r *http.Request) {
	cfg, _ := config.LoadHubConfig("")
	data := map[string]interface{}{"hub": cfg.Hub}
	hubMap, _ := data["hub"].(config.HubSettings)
	_ = hubMap
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"hub": map[string]interface{}{
			"domain":              cfg.Hub.Domain,
			"public_url":          cfg.Hub.PublicURL,
			"use_https":           cfg.Hub.UseHTTPS,
			"trusted_hosts":       cfg.Hub.TrustedHosts,
			"public_url_resolved": config.ResolvePublicURL(cfg.Hub),
		},
	})
}

func (s *Server) handleUpdateHubConfig(w http.ResponseWriter, r *http.Request) {
	if s.Mode != "hub" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": "Доступно только в hub-режиме"})
		return
	}
	var body struct {
		Hub map[string]interface{} `json:"hub"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": err.Error()})
		return
	}
	current, _ := config.LoadHubConfig("")
	raw, _ := json.Marshal(current.Hub)
	var settings config.HubSettings
	_ = json.Unmarshal(raw, &settings)
	if body.Hub != nil {
		patch, _ := json.Marshal(body.Hub)
		_ = json.Unmarshal(patch, &settings)
	}
	settings.Domain = config.NormalizeDomain(settings.Domain)
	if settings.Domain != "" {
		settings.TrustedHosts = []string{settings.Domain, "*." + settings.Domain}
	}
	cfg := &config.HubFile{Hub: settings}
	_ = config.SaveHubConfig(cfg, "")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"hub": map[string]interface{}{
			"domain":              settings.Domain,
			"public_url":          settings.PublicURL,
			"use_https":           settings.UseHTTPS,
			"trusted_hosts":       settings.TrustedHosts,
			"public_url_resolved": config.ResolvePublicURL(settings),
		},
	})
}

func (s *Server) handleHubInfo(w http.ResponseWriter, r *http.Request) {
	if s.Mode != "hub" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"mode": s.Mode, "public_url": "", "domain": "",
		})
		return
	}
	cfg, _ := config.LoadHubConfig("")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"mode":        "hub",
		"domain":      cfg.Hub.Domain,
		"public_url":  config.ResolvePublicURL(cfg.Hub),
		"use_https":   cfg.Hub.UseHTTPS,
	})
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	enabled := s.Mode == "hub" && HubAuthEnabled()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"auth_required": enabled,
		"hub_name":      HubName(),
		"authenticated": IsAuthenticated(r),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if s.Mode != "hub" || !HubAuthEnabled() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": "Авторизация не настроена"})
		return
	}
	var body struct {
		Name string `json:"name"`
		Key  string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"detail": err.Error()})
		return
	}
	if !VerifyCredentials(strings.TrimSpace(body.Name), body.Key) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"detail": "Неверное имя хаба или ключ"})
		return
	}
	token := CreateSessionToken()
	w.Header().Set("Set-Cookie", SessionCookieHeader(token, RequestIsSecure(r)))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Set-Cookie", ClearSessionCookieHeader(RequestIsSecure(r)))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSensorTypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"types": collector.GetRegisteredTypes()})
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func (s *Server) handleLiveWS(w http.ResponseWriter, r *http.Request) {
	if s.Mode == "hub" && HubAuthEnabled() && !IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.Hub.Add(conn)
	agent := r.URL.Query().Get("agent")
	var snapshot map[string]map[string]interface{}
	if agent != "" {
		raw := s.Fleet.GetAgentSnapshot(agent)
		snapshot = map[string]map[string]interface{}{}
		for key, value := range raw {
			snapshot[protocol.StripAgentPrefix(agent, key)] = value
		}
	} else if s.Collector != nil {
		snapshot = s.Collector.GetLatestSnapshot()
	} else if s.Mode == "hub" {
		snapshot = s.Fleet.GetCombinedSnapshot()
	} else {
		snapshot = map[string]map[string]interface{}{}
	}
	_ = conn.WriteJSON(map[string]interface{}{
		"type": "snapshot", "data": snapshot, "agent_id": agent,
	})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
	s.Hub.Remove(conn)
}

