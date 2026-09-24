package enrich

import (
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/diskdiscovery"
	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
)

var defaultSensors = map[string]protocol.SensorMeta{
	"cpu_percent": {
		ID: "cpu_percent", Name: "Загрузка процессора", Type: "system.cpu_percent", Unit: "%", Enabled: true,
	},
	"ram_used": {
		ID: "ram_used", Name: "Использование памяти", Type: "system.memory_percent", Unit: "%", Enabled: true,
	},
}

func diskPartitions(system map[string]interface{}) []map[string]interface{} {
	for _, key := range []string{"disks", "partitions", "storage"} {
		if value, ok := system[key].([]interface{}); ok && len(value) > 0 {
			result := make([]map[string]interface{}, 0, len(value))
			for _, item := range value {
				if m, ok := item.(map[string]interface{}); ok {
					result = append(result, m)
				}
			}
			return result
		}
	}
	return nil
}

func ReadingFromSystem(system map[string]interface{}, sensorID string, ts float64) map[string]interface{} {
	if system == nil {
		return nil
	}
	if ts == 0 {
		ts = float64(time.Now().UnixNano()) / 1e9
	}
	if sensorID == "cpu_percent" {
		cpu, _ := system["cpu"].(map[string]interface{})
		if cpu == nil {
			return nil
		}
		if v, ok := toFloat(cpu["percent"]); ok {
			return map[string]interface{}{"value": v, "status": "ok", "ts": ts, "unit": "%"}
		}
	}
	if sensorID == "ram_used" {
		mem, _ := system["memory"].(map[string]interface{})
		if mem == nil {
			return nil
		}
		if v, ok := toFloat(mem["percent"]); ok {
			return map[string]interface{}{"value": v, "status": "ok", "ts": ts, "unit": "%"}
		}
	}
	if len(sensorID) > 10 && sensorID[:10] == "disk_auto_" {
		for _, part := range diskPartitions(system) {
			mount := ""
			if v, ok := part["mountpoint"].(string); ok {
				mount = v
			}
			if diskdiscovery.MountToSensorID(mount) != sensorID {
				continue
			}
			if v, ok := toFloat(part["percent"]); ok {
				return map[string]interface{}{"value": v, "status": "ok", "ts": ts, "unit": "%"}
			}
		}
	}
	return nil
}

func SensorMetasFromSystem(system map[string]interface{}) []map[string]interface{} {
	metas := make([]map[string]interface{}, 0)
	known := map[string]bool{}
	for _, meta := range defaultSensors {
		metas = append(metas, sensorMetaToMap(meta))
		known[meta.ID] = true
	}
	for _, part := range diskPartitions(system) {
		mount := ""
		if v, ok := part["mountpoint"].(string); ok {
			mount = v
		}
		sensorID := diskdiscovery.MountToSensorID(mount)
		if known[sensorID] {
			continue
		}
		known[sensorID] = true
		meta := protocol.SensorMeta{
			ID: sensorID, Name: "Диск " + mount, Type: "system.disk_usage", Unit: "%", Enabled: true,
		}
		metas = append(metas, sensorMetaToMap(meta))
	}
	return metas
}

func EnrichAgentReport(report *protocol.AgentReport) *protocol.AgentReport {
	system := report.System
	if system == nil {
		system = map[string]interface{}{}
	}
	now := float64(time.Now().UnixNano()) / 1e9
	metricsMap := map[string]protocol.MetricPoint{}
	for _, point := range report.Metrics {
		metricsMap[point.SensorID] = point
	}
	setMetric := func(sensorID string, value interface{}) {
		if value == nil {
			delete(metricsMap, sensorID)
			return
		}
		if v, ok := toFloat(value); ok {
			metricsMap[sensorID] = protocol.MetricPoint{
				SensorID: sensorID, TS: now, Value: &v, Status: "ok",
			}
		}
	}
	cpu, _ := system["cpu"].(map[string]interface{})
	if cpu != nil {
		setMetric("cpu_percent", cpu["percent"])
	}
	mem, _ := system["memory"].(map[string]interface{})
	if mem != nil {
		setMetric("ram_used", mem["percent"])
	}
	for _, part := range diskPartitions(system) {
		mount := ""
		if v, ok := part["mountpoint"].(string); ok {
			mount = v
		}
		sensorID := diskdiscovery.MountToSensorID(mount)
		setMetric(sensorID, part["percent"])
	}
	sensors := report.Sensors
	knownIDs := map[string]bool{}
	for _, s := range sensors {
		knownIDs[s.ID] = true
	}
	for sensorID := range metricsMap {
		if knownIDs[sensorID] {
			continue
		}
		if meta, ok := defaultSensors[sensorID]; ok {
			sensors = append(sensors, meta)
			knownIDs[sensorID] = true
			continue
		}
		if len(sensorID) > 10 && sensorID[:10] == "disk_auto_" {
			mount := sensorID[10:]
			for _, part := range diskPartitions(system) {
				m := ""
				if v, ok := part["mountpoint"].(string); ok {
					m = v
				}
				if diskdiscovery.MountToSensorID(m) == sensorID {
					mount = m
					break
				}
			}
			sensors = append(sensors, protocol.SensorMeta{
				ID: sensorID, Name: "Диск " + mount, Type: "system.disk_usage", Unit: "%", Enabled: true,
			})
		}
	}
	if len(sensors) == 0 && len(metricsMap) > 0 {
		for _, m := range SensorMetasFromSystem(system) {
			sensors = append(sensors, mapToSensorMeta(m))
		}
	}
	metrics := make([]protocol.MetricPoint, 0, len(metricsMap))
	for _, p := range metricsMap {
		metrics = append(metrics, p)
	}
	return &protocol.AgentReport{
		AgentID: report.AgentID, Hostname: report.Hostname, Metrics: metrics,
		System: system, ConfigVersion: report.ConfigVersion, Sensors: sensors,
	}
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

func sensorMetaToMap(meta protocol.SensorMeta) map[string]interface{} {
	return map[string]interface{}{
		"id": meta.ID, "name": meta.Name, "type": meta.Type, "unit": meta.Unit, "enabled": meta.Enabled,
	}
}

func mapToSensorMeta(m map[string]interface{}) protocol.SensorMeta {
	return protocol.SensorMeta{
		ID: fmtID(m["id"]), Name: fmtID(m["name"]), Type: fmtID(m["type"]),
		Unit: fmtID(m["unit"]), Enabled: true,
	}
}

func fmtID(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
