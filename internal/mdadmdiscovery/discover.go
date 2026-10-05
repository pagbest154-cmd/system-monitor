package mdadmdiscovery

import (
	"fmt"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
)

// DiscoverMdadmSensors builds sensor configs for each array in /proc/mdstat.
func DiscoverMdadmSensors() []config.SensorConfig {
	arrays := ReadMdstat()
	if len(arrays) == 0 {
		return nil
	}
	sensors := make([]config.SensorConfig, 0, len(arrays))
	warn := 1.0
	critical := 2.0
	interval := 30
	for _, arr := range arrays {
		level := arr.RaidLevel
		if level == "" {
			level = "unknown"
		}
		sensors = append(sensors, config.SensorConfig{
			ID:            DeviceToSensorID(arr.Device),
			Name:          fmt.Sprintf("RAID %s (%s)", arr.Device, level),
			Type:          "system.mdadm_status",
			Enabled:       true,
			IntervalSec:   &interval,
			Unit:          "",
			Platforms:     []string{"linux"},
			Params:        map[string]interface{}{"device": arr.Device},
			WarnAbove:     &warn,
			CriticalAbove: &critical,
		})
	}
	return sensors
}

// FindArray returns parsed array by device name (e.g. md0).
func FindArray(device string) *Array {
	for _, arr := range ReadMdstat() {
		if arr.Device == device {
			return &arr
		}
	}
	return nil
}
