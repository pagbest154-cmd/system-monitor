//go:build linux

package collector

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func readLinuxHwmon() *float64 {
	entries, err := os.ReadDir("/sys/class/hwmon")
	if err != nil {
		return nil
	}
	temps := make([]float64, 0)
	for _, entry := range entries {
		hwmonDir := filepath.Join("/sys/class/hwmon", entry.Name())
		files, err := os.ReadDir(hwmonDir)
		if err != nil {
			continue
		}
		for _, file := range files {
			name := file.Name()
			if !strings.HasPrefix(name, "temp") || !strings.HasSuffix(name, "_input") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(hwmonDir, name))
			if err != nil {
				continue
			}
			raw, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				continue
			}
			celsius := float64(raw) / 1000.0
			if celsius > 0 && celsius < 150 {
				temps = append(temps, celsius)
			}
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
