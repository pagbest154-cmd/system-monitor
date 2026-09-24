package collector

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/pagbest154-cmd/system-monitor/internal/hiddenexec"
)

var cpuNameRE = regexp.MustCompile(`(?i)cpu|core|package|tctl|processor|xeon|ryzen`)
var gpuNameRE = regexp.MustCompile(`(?i)gpu|graphics|nvidia|geforce|radeon|video`)

type temperatureSensor struct{ baseSensor }
type gpuTemperatureSensor struct{ baseSensor }

func (s temperatureSensor) Read() SensorReading {
	value := readCPUTemperature()
	if value == nil {
		return SensorReading{SensorID: s.ID(), Status: "unknown"}
	}
	v := math.Round(*value*10) / 10
	return SensorReading{SensorID: s.ID(), Value: &v, Status: evaluateStatus(s.cfg, &v)}
}

func (s gpuTemperatureSensor) Read() SensorReading {
	value := readGPUTemperature()
	if value == nil {
		return SensorReading{SensorID: s.ID(), Status: "unknown"}
	}
	v := math.Round(*value*10) / 10
	return SensorReading{SensorID: s.ID(), Value: &v, Status: evaluateStatus(s.cfg, &v)}
}

func readCPUTemperature() *float64 {
	if runtime.GOOS == "linux" {
		return readLinuxThermal()
	}
	if runtime.GOOS == "windows" {
		if v := readWindowsACPI(); v != nil {
			return v
		}
		readings := readHWMonitor("root/LibreHardwareMonitor")
		if len(readings) == 0 {
			readings = readHWMonitor("root/OpenHardwareMonitor")
		}
		return pickCPUTemp(readings)
	}
	return nil
}

func readGPUTemperature() *float64 {
	if v := readNvidiaSMITemp(); v != nil {
		return v
	}
	if runtime.GOOS == "windows" {
		readings := readHWMonitor("root/LibreHardwareMonitor")
		if len(readings) == 0 {
			readings = readHWMonitor("root/OpenHardwareMonitor")
		}
		return pickGPUTemp(readings)
	}
	return nil
}

func readLinuxThermal() *float64 {
	thermalDir := "/sys/class/thermal"
	entries, err := os.ReadDir(thermalDir)
	if err != nil {
		return nil
	}
	temps := make([]float64, 0)
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "thermal_zone") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(thermalDir, entry.Name(), "temp"))
		if err != nil {
			continue
		}
		raw, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			continue
		}
		temps = append(temps, float64(raw)/1000.0)
	}
	if len(temps) == 0 {
		return nil
	}
	max := temps[0]
	for _, t := range temps[1:] {
		if t > max {
			max = t
		}
	}
	return &max
}

func readWindowsACPI() *float64 {
	script := "Get-CimInstance -Namespace root/wmi -ClassName MSAcpi_ThermalZoneTemperature | Select-Object -ExpandProperty CurrentTemperature"
	out, err := hiddenexec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		return nil
	}
	temps := make([]float64, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		raw, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
		if err != nil {
			continue
		}
		celsius := (raw/10.0) - 273.15
		if celsius > 0 && celsius < 150 {
			temps = append(temps, celsius)
		}
	}
	if len(temps) == 0 {
		return nil
	}
	max := temps[0]
	for _, t := range temps[1:] {
		if t > max {
			max = t
		}
	}
	return &max
}

func readHWMonitor(namespace string) []struct {
	name  string
	value float64
} {
	script := "Get-CimInstance -Namespace " + namespace + " -ClassName Sensor -ErrorAction SilentlyContinue | Where-Object { $_.SensorType -eq 'Temperature' -and $_.Value -ne $null } | Select-Object Name, Value | ConvertTo-Json -Compress"
	out, err := hiddenexec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		return nil
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(out, &items); err != nil {
		var single map[string]interface{}
		if err2 := json.Unmarshal(out, &single); err2 == nil {
			items = []map[string]interface{}{single}
		}
	}
	readings := make([]struct {
		name  string
		value float64
	}, 0)
	for _, item := range items {
		name, _ := item["Name"].(string)
		v, ok := toFloat(item["Value"])
		if !ok || v <= 0 || v >= 150 {
			continue
		}
		readings = append(readings, struct {
			name  string
			value float64
		}{name, v})
	}
	return readings
}

func pickCPUTemp(readings []struct {
	name  string
	value float64
}) *float64 {
	cpuVals := make([]float64, 0)
	for _, r := range readings {
		if cpuNameRE.MatchString(r.name) {
			cpuVals = append(cpuVals, r.value)
		}
	}
	if len(cpuVals) > 0 {
		max := cpuVals[0]
		for _, v := range cpuVals[1:] {
			if v > max {
				max = v
			}
		}
		return &max
	}
	if len(readings) > 0 {
		max := readings[0].value
		for _, r := range readings[1:] {
			if r.value > max {
				max = r.value
			}
		}
		return &max
	}
	return nil
}

func pickGPUTemp(readings []struct {
	name  string
	value float64
}) *float64 {
	gpuVals := make([]float64, 0)
	for _, r := range readings {
		if gpuNameRE.MatchString(r.name) {
			gpuVals = append(gpuVals, r.value)
		}
	}
	if len(gpuVals) == 0 {
		return nil
	}
	max := gpuVals[0]
	for _, v := range gpuVals[1:] {
		if v > max {
			max = v
		}
	}
	return &max
}

func readNvidiaSMITemp() *float64 {
	out, err := hiddenexec.Command("nvidia-smi", "--query-gpu=temperature.gpu", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil
	}
	line := strings.TrimSpace(string(out))
	if line == "" {
		return nil
	}
	v, err := strconv.ParseFloat(strings.Split(line, "\n")[0], 64)
	if err != nil {
		return nil
	}
	return &v
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
