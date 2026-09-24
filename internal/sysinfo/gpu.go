package sysinfo

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/hiddenexec"
)

const nvidiaCacheTTL = 30 * time.Second

var (
	nvidiaCache struct {
		gpus []map[string]interface{}
		at   time.Time
	}
	nvidiaCacheMu sync.Mutex
)

func getGPUs() []map[string]interface{} {
	return getNvidiaGPUsCached()
}

func getNvidiaGPUsCached() []map[string]interface{} {
	nvidiaCacheMu.Lock()
	if nvidiaCache.gpus != nil && time.Since(nvidiaCache.at) < nvidiaCacheTTL {
		gpus := nvidiaCache.gpus
		nvidiaCacheMu.Unlock()
		return gpus
	}
	nvidiaCacheMu.Unlock()

	gpus := fetchNvidiaGPUs()
	nvidiaCacheMu.Lock()
	nvidiaCache.gpus = gpus
	nvidiaCache.at = time.Now()
	nvidiaCacheMu.Unlock()
	return gpus
}

// NvidiaGPUTemperature returns cached nvidia-smi temperature when available.
func NvidiaGPUTemperature() *float64 {
	gpus := getNvidiaGPUsCached()
	if len(gpus) == 0 {
		return nil
	}
	temp, ok := gpus[0]["temperature_c"]
	if !ok || temp == nil {
		return nil
	}
	switch v := temp.(type) {
	case float64:
		return &v
	case int:
		f := float64(v)
		return &f
	default:
		return nil
	}
}

func fetchNvidiaGPUs() []map[string]interface{} {
	cmd := hiddenexec.Command("nvidia-smi",
		"--query-gpu=name,memory.total,memory.used,memory.free,utilization.gpu,temperature.gpu,driver_version",
		"--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return []map[string]interface{}{}
	}
	gpus := make([]map[string]interface{}, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.Split(line, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) < 7 {
			continue
		}
		totalMB := parseFloat(parts[1])
		usedMB := parseFloat(parts[2])
		var memPercent interface{}
		if totalMB > 0 {
			memPercent = round1(usedMB / totalMB * 100)
		}
		var temp interface{}
		if parts[5] != "N/A" && parts[5] != "[N/A]" {
			temp = parseFloat(parts[5])
		}
		gpus = append(gpus, map[string]interface{}{
			"name":                 parts[0],
			"vendor":               "NVIDIA",
			"driver_version":       parts[6],
			"cuda_version":         nil,
			"cuda_toolkit_version": nil,
			"memory_total_gb":      round1(totalMB / 1024),
			"memory_used_gb":       round1(usedMB / 1024),
			"memory_free_gb":       round1(parseFloat(parts[3]) / 1024),
			"memory_percent":       memPercent,
			"utilization_percent":  parseFloat(parts[4]),
			"temperature_c":        temp,
			"resolution":           nil,
		})
	}
	return gpus
}

func parseFloat(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}
