package collector

import (
	"github.com/pagbest154-cmd/system-monitor/internal/config"
)

type SensorReading struct {
	SensorID string
	Value    *float64
	Status   string
	Error    string
}

func (r SensorReading) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"sensor_id": r.SensorID,
		"value":     r.Value,
		"status":    r.Status,
	}
	if r.Error != "" {
		m["error"] = r.Error
	}
	return m
}

type Sensor interface {
	ID() string
	Config() config.SensorConfig
	Read() SensorReading
}

type baseSensor struct {
	cfg config.SensorConfig
}

func (s baseSensor) ID() string              { return s.cfg.ID }
func (s baseSensor) Config() config.SensorConfig { return s.cfg }

func evaluateStatus(cfg config.SensorConfig, value *float64) string {
	if value == nil {
		return "unknown"
	}
	if cfg.CriticalAbove != nil && *value >= *cfg.CriticalAbove {
		return "critical"
	}
	if cfg.WarnAbove != nil && *value >= *cfg.WarnAbove {
		return "warning"
	}
	return "ok"
}
