package diskdiscovery

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/shirou/gopsutil/v4/disk"
)

var ignoreFSTypes = map[string]bool{
	"tmpfs": true, "devtmpfs": true, "squashfs": true, "overlay": true,
	"proc": true, "sysfs": true, "devfs": true, "autofs": true,
	"cgroup": true, "cgroup2": true, "pstore": true, "bpf": true,
	"tracefs": true, "debugfs": true, "securityfs": true, "configfs": true,
	"fusectl": true, "mqueue": true, "hugetlbfs": true,
}

func MountToSensorID(mountpoint string) string {
	cleaned := strings.Trim(strings.TrimRight(mountpoint, "\\/"), " ")
	if cleaned == "" {
		return "disk_auto_root"
	}
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.ReplaceAll(cleaned, "\\", "_")
	cleaned = strings.ReplaceAll(cleaned, "/", "_")
	re := regexp.MustCompile(`[^a-z0-9_]+`)
	cleaned = strings.ToLower(cleaned)
	cleaned = re.ReplaceAllString(cleaned, "_")
	re2 := regexp.MustCompile(`_+`)
	cleaned = strings.Trim(re2.ReplaceAllString(cleaned, "_"), "_")
	if cleaned == "" {
		cleaned = "root"
	}
	return "disk_auto_" + cleaned
}

func DiscoverDiskSensors() []config.SensorConfig {
	partitions, err := disk.Partitions(true)
	if err != nil {
		partitions, _ = disk.Partitions(false)
	}
	sensors := make([]config.SensorConfig, 0)
	seen := map[string]bool{}
	for _, part := range partitions {
		if ignoreFSTypes[part.Fstype] {
			continue
		}
		if _, err := disk.Usage(part.Mountpoint); err != nil {
			continue
		}
		sensorID := MountToSensorID(part.Mountpoint)
		if seen[sensorID] {
			suffix := 2
			for seen[sensorID+"_"+strconv.Itoa(suffix)] {
				suffix++
			}
			sensorID = sensorID + "_" + strconv.Itoa(suffix)
		}
		seen[sensorID] = true
		label := part.Mountpoint
		if part.Device != "" {
			label = part.Mountpoint + " (" + part.Device + ")"
		}
		warn := 85.0
		critical := 95.0
		interval := 30
		sensors = append(sensors, config.SensorConfig{
			ID:            sensorID,
			Name:          "Диск " + label,
			Type:          "system.disk_usage",
			Enabled:       true,
			IntervalSec:   &interval,
			Unit:          "%",
			Params:        map[string]interface{}{"path": part.Mountpoint},
			WarnAbove:     &warn,
			CriticalAbove: &critical,
		})
	}
	return sensors
}
