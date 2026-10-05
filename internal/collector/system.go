package collector

import (
	"math"
	"strings"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	netps "github.com/shirou/gopsutil/v4/net"
)

var netCounters = map[string]struct {
	bytes uint64
	ts    float64
}{}

// networkRateMbps converts byte counter delta to megabits per second (decimal, as in ISP speed tests).
func networkRateMbps(prevBytes, currBytes uint64, elapsedSec float64) float64 {
	elapsedSec = math.Max(elapsedSec, 0.001)
	if currBytes < prevBytes {
		return 0
	}
	delta := float64(currBytes - prevBytes)
	mbps := delta * 8 / 1_000_000 / elapsedSec
	if mbps > 200_000 {
		return 0
	}
	return math.Round(mbps*100) / 100
}

// maxNetworkDeltaBytes — верхняя граница прироста счётчика за интервал (~100 Гбит/с).
func maxNetworkDeltaBytes(elapsedSec float64) uint64 {
	const bytesPerSec100G = 12_500_000_000.0
	return uint64(elapsedSec * bytesPerSec100G * 1.05)
}

func networkSensorIntervalSec(cfg config.SensorConfig, fallback int) float64 {
	if cfg.IntervalSec != nil && *cfg.IntervalSec > 0 {
		return float64(*cfg.IntervalSec)
	}
	if fallback > 0 {
		return float64(fallback)
	}
	return 2
}

func networkInterfaceParam(params map[string]interface{}) string {
	if params == nil {
		return ""
	}
	iface, ok := params["interface"].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(iface)
}

func matchNetworkInterface(name, want string) bool {
	return strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(want))
}

func networkCounterBytes(counters []netps.IOCountersStat, direction, ifaceFilter string) (uint64, bool) {
	var current uint64
	matched := false
	for _, c := range counters {
		if ifaceFilter != "" && !matchNetworkInterface(c.Name, ifaceFilter) {
			continue
		}
		matched = true
		if direction == "sent" {
			current += c.BytesSent
		} else {
			current += c.BytesRecv
		}
	}
	if ifaceFilter != "" {
		return current, matched
	}
	return current, len(counters) > 0
}

// SensorDisplayName adds interface hint for network sensors (legend / tooltips).
func SensorDisplayName(cfg config.SensorConfig) string {
	iface := networkInterfaceParam(cfg.Params)
	if iface != "" && cfg.Type == "system.network_bytes" {
		return cfg.Name + " (" + iface + ")"
	}
	return cfg.Name
}

type cpuPercentSensor struct{ baseSensor }

func (s cpuPercentSensor) Read() SensorReading {
	v, _ := cpu.Percent(0, false)
	value := 0.0
	if len(v) > 0 {
		value = v[0]
	}
	value = math.Round(value*100) / 100
	return SensorReading{SensorID: s.ID(), Value: &value, Status: evaluateStatus(s.cfg, &value)}
}

type memoryPercentSensor struct{ baseSensor }

func (s memoryPercentSensor) Read() SensorReading {
	vm, err := mem.VirtualMemory()
	if err != nil || vm == nil {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: "memory stats unavailable"}
	}
	value := math.Round(vm.UsedPercent*100) / 100
	return SensorReading{SensorID: s.ID(), Value: &value, Status: evaluateStatus(s.cfg, &value)}
}

type diskUsageSensor struct{ baseSensor }

func (s diskUsageSensor) Read() SensorReading {
	path := config.DefaultDiskPath()
	if s.cfg.Params != nil {
		if p, ok := s.cfg.Params["path"].(string); ok && p != "" {
			path = p
		}
	}
	usage, err := disk.Usage(path)
	if err != nil {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: err.Error()}
	}
	value := math.Round(usage.UsedPercent*100) / 100
	details := map[string]interface{}{
		"used_gb":  bytesToGB(usage.Used),
		"total_gb": bytesToGB(usage.Total),
		"free_gb":  bytesToGB(usage.Free),
	}
	return SensorReading{
		SensorID: s.ID(),
		Value:    &value,
		Status:   evaluateStatus(s.cfg, &value),
		Details:  details,
	}
}

func bytesToGB(value uint64) float64 {
	return math.Round(float64(value)/(1024*1024*1024)*10) / 10
}

type networkBytesSensor struct{ baseSensor }

func (s networkBytesSensor) Read() SensorReading {
	direction := "recv"
	if s.cfg.Params != nil {
		if d, ok := s.cfg.Params["direction"].(string); ok && d != "" {
			direction = d
		}
	}
	ifaceFilter := networkInterfaceParam(s.cfg.Params)
	perNic := ifaceFilter != ""
	counters, err := netps.IOCounters(perNic)
	if err != nil {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: err.Error()}
	}
	now := float64(time.Now().UnixNano()) / 1e9
	current, okCounters := networkCounterBytes(counters, direction, ifaceFilter)
	if !okCounters {
		msg := "network counters unavailable"
		if ifaceFilter != "" {
			msg = "interface not found: " + ifaceFilter
		}
		return SensorReading{SensorID: s.ID(), Status: "error", Error: msg}
	}
	prev, ok := netCounters[s.ID()]
	if !ok {
		netCounters[s.ID()] = struct {
			bytes uint64
			ts    float64
		}{current, now}
		zero := 0.0
		return SensorReading{SensorID: s.ID(), Value: &zero, Status: "ok"}
	}
	interval := networkSensorIntervalSec(s.Config(), 2)
	elapsed := now - prev.ts
	if elapsed <= 0 || elapsed > interval*20 {
		netCounters[s.ID()] = struct {
			bytes uint64
			ts    float64
		}{current, now}
		zero := 0.0
		return SensorReading{SensorID: s.ID(), Value: &zero, Status: "ok"}
	}
	if elapsed < interval*0.25 {
		elapsed = interval
	}
	if current < prev.bytes {
		netCounters[s.ID()] = struct {
			bytes uint64
			ts    float64
		}{current, now}
		zero := 0.0
		return SensorReading{SensorID: s.ID(), Value: &zero, Status: "ok"}
	}
	delta := current - prev.bytes
	if delta > maxNetworkDeltaBytes(elapsed) {
		netCounters[s.ID()] = struct {
			bytes uint64
			ts    float64
		}{current, now}
		zero := 0.0
		return SensorReading{SensorID: s.ID(), Value: &zero, Status: "ok"}
	}
	mbps := networkRateMbps(prev.bytes, current, elapsed)
	netCounters[s.ID()] = struct {
		bytes uint64
		ts    float64
	}{current, now}
	return SensorReading{SensorID: s.ID(), Value: &mbps, Status: evaluateStatus(s.cfg, &mbps)}
}
