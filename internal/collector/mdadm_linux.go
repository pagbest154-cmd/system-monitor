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
	value := mdadmdiscovery.HealthMetricValue(*arr)
	status := s.healthStatus(*arr, &value)
	devices := make([]map[string]interface{}, 0, len(arr.Devices))
	for _, d := range arr.Devices {
		devices = append(devices, map[string]interface{}{
			"name": d.Name, "slot": d.Slot, "state": d.State,
		})
	}
	details := map[string]interface{}{
		"device":         arr.Device,
		"raid_level":     arr.RaidLevel,
		"state":          arr.State,
		"active_devices": arr.ActiveDevices,
		"failed_devices": arr.FailedDevices,
		"devices":        devices,
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
	v := mdadmdiscovery.HealthMetricValue(arr)
	if v >= 2 {
		return "critical"
	}
	if v >= 1 {
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
