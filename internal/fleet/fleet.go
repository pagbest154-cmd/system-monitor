package fleet

import (
	"os"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
	"github.com/pagbest154-cmd/system-monitor/internal/storage"
)

func DefaultAgentID(configured string) string {
	if configured = trim(configured); configured != "" {
		return configured
	}
	hostname, _ := os.Hostname()
	return hostname
}

func VerifyAgentToken(agentID, token string) bool {
	if token == "" {
		return false
	}
	entry := config.FindAgentEntry(agentID, nil)
	if entry == nil {
		return false
	}
	return entry.Token == token
}

func BuildAgentConfigResponse(agentID string, configVersion int) protocol.AgentConfigResponse {
	entry := config.FindAgentEntry(agentID, nil)
	if entry == nil {
		return protocol.AgentConfigResponse{Version: configVersion}
	}
	overrides := make([]protocol.SensorOverride, 0, len(entry.Overrides))
	for _, item := range entry.Overrides {
		overrides = append(overrides, protocol.SensorOverride{
			SensorID:      item.SensorID,
			Enabled:       item.Enabled,
			IntervalSec:   item.IntervalSec,
			WarnAbove:     item.WarnAbove,
			CriticalAbove: item.CriticalAbove,
			Params:        item.Params,
		})
	}
	version := configVersion
	if len(overrides) > 0 && version < 1 {
		version = 1
	}
	return protocol.AgentConfigResponse{Version: version, Overrides: overrides}
}

type LiveHub interface {
	ScheduleBroadcast(payload map[string]interface{})
}

func IngestAgentReport(report *protocol.AgentReport, store *storage.MetricStore, liveHub LiveHub) map[string]interface{} {
	entry := config.FindAgentEntry(report.AgentID, nil)
	displayName := report.AgentID
	if entry != nil && entry.Name != "" {
		displayName = entry.Name
	}
	rows := make([]storage.MetricRow, 0)
	for _, point := range report.Metrics {
		if point.Value == nil {
			continue
		}
		rows = append(rows, storage.MetricRow{
			SensorID: protocol.PrefixedSensorID(report.AgentID, point.SensorID),
			TS:       point.TS,
			Value:    point.Value,
			Status:   point.Status,
		})
	}
	_ = store.InsertBatch(rows)
	sensorsPayload := make([]map[string]interface{}, 0, len(report.Sensors))
	for _, s := range report.Sensors {
		sensorsPayload = append(sensorsPayload, sensorMetaMap(s))
	}
	_ = store.UpsertAgent(storage.AgentUpsert{
		AgentID:  report.AgentID,
		Name:     displayName,
		Hostname: report.Hostname,
		Status:   "online",
		System:   report.System,
		Sensors:  sensorsPayload,
	})
	if liveHub != nil {
		latest := map[string]interface{}{}
		for _, point := range report.Metrics {
			if point.Value == nil {
				continue
			}
			fullID := protocol.PrefixedSensorID(report.AgentID, point.SensorID)
			latest[fullID] = map[string]interface{}{
				"sensor_id": fullID,
				"value":     point.Value,
				"status":    point.Status,
				"ts":        point.TS,
			}
		}
		if len(latest) > 0 {
			liveHub.ScheduleBroadcast(map[string]interface{}{
				"type": "update", "data": latest, "agent_id": report.AgentID,
			})
		}
	}
	return map[string]interface{}{"status": "ok", "metrics": len(rows)}
}

func sensorMetaMap(s protocol.SensorMeta) map[string]interface{} {
	m := map[string]interface{}{
		"id": s.ID, "name": s.Name, "type": s.Type, "unit": s.Unit, "enabled": s.Enabled,
	}
	if s.IntervalSec != nil {
		m["interval_sec"] = *s.IntervalSec
	}
	if s.Params != nil {
		m["params"] = s.Params
	}
	if s.WarnAbove != nil {
		m["warn_above"] = *s.WarnAbove
	}
	if s.CriticalAbove != nil {
		m["critical_above"] = *s.CriticalAbove
	}
	return m
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
