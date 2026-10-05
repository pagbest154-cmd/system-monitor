//go:build !linux

package collector

import (
	"github.com/pagbest154-cmd/system-monitor/internal/config"
)

type mdadmStatusSensor struct{ baseSensor }

func (s mdadmStatusSensor) Read() SensorReading {
	return SensorReading{
		SensorID: s.ID(),
		Status:   "unknown",
		Error:    "system.mdadm_status доступен только на Linux",
	}
}

func mdadmDeviceParam(cfg config.SensorConfig) string {
	if cfg.Params == nil {
		return ""
	}
	v, _ := cfg.Params["device"].(string)
	return v
}
