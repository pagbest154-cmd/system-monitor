package collector

import (
	"math"
	"strings"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/process"
)

type loadAverageSensor struct{ baseSensor }

func (s loadAverageSensor) Read() SensorReading {
	avg, err := load.Avg()
	if err != nil || avg == nil {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: "load average unavailable"}
	}
	value := math.Round(avg.Load1*100) / 100
	return SensorReading{SensorID: s.ID(), Value: &value, Status: evaluateStatus(s.cfg, &value)}
}

type processCPUPercentSensor struct{ baseSensor }

func (s processCPUPercentSensor) Read() SensorReading {
	name := processNameParam(s.cfg)
	if name == "" {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: "params.name required"}
	}
	procs, err := process.Processes()
	if err != nil {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: err.Error()}
	}
	var total float64
	for _, p := range procs {
		pname, err := p.Name()
		if err != nil {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(pname), name) {
			continue
		}
		cpu, err := p.CPUPercent()
		if err != nil {
			continue
		}
		total += cpu
	}
	value := math.Round(total*100) / 100
	return SensorReading{SensorID: s.ID(), Value: &value, Status: evaluateStatus(s.cfg, &value)}
}

func processNameParam(cfg config.SensorConfig) string {
	if cfg.Params == nil {
		return ""
	}
	if n, ok := cfg.Params["name"].(string); ok {
		return strings.TrimSpace(n)
	}
	return ""
}
