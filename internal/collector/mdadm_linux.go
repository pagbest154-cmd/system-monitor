//go:build linux

package collector

import (
	"fmt"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/mdadmdiscovery"
)

type mdadmStatusSensor struct{ baseSensor }

func (s mdadmStatusSensor) Read() SensorReading {
	device := mdadmDeviceParam(s.cfg)
	if device == "" {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: "params.device required"}
	}
	arr := mdadmdiscovery.FindArray(device)
	if arr == nil {
		return SensorReading{SensorID: s.ID(), Status: "unknown", Error: fmt.Sprintf("массив %s не найден", device)}
	}
	failed := float64(arr.FailedDevices)
	value := failed
	status := s.healthStatus(*arr, &value)
	details := map[string]interface{}{
		"device":         arr.Device,
		"raid_level":     arr.RaidLevel,
		"state":          arr.State,
		"active_devices": arr.ActiveDevices,
		"failed_devices": arr.FailedDevices,
	}
	if arr.CheckProgress != nil {
		details["check_progress"] = *arr.CheckProgress
	} else {
		details["check_progress"] = nil
	}
	return SensorReading{
		SensorID: s.ID(),
		Value:    &value,
		Status:   status,
		Details:  details,
	}
}

func (s mdadmStatusSensor) healthStatus(arr mdadmdiscovery.Array, value *float64) string {
	if arr.FailedDevices > 0 {
		return "critical"
	}
	if arr.State == "degraded" {
		return "warning"
	}
	return evaluateStatus(s.cfg, value)
}

func mdadmDeviceParam(cfg config.SensorConfig) string {
	if cfg.Params == nil {
		return ""
	}
	v, ok := cfg.Params["device"].(string)
	if !ok {
		return ""
	}
	return v
}
