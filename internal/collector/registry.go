package collector

import (
	"fmt"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
)

type factory func(config.SensorConfig) Sensor

var registry = map[string]factory{
	"system.cpu_percent":      func(c config.SensorConfig) Sensor { return cpuPercentSensor{baseSensor{c}} },
	"system.memory_percent":   func(c config.SensorConfig) Sensor { return memoryPercentSensor{baseSensor{c}} },
	"system.disk_usage":       func(c config.SensorConfig) Sensor { return diskUsageSensor{baseSensor{c}} },
	"system.network_bytes":    func(c config.SensorConfig) Sensor { return networkBytesSensor{baseSensor{c}} },
	"system.temperature":      func(c config.SensorConfig) Sensor { return temperatureSensor{baseSensor{c}} },
	"system.gpu_temperature":  func(c config.SensorConfig) Sensor { return gpuTemperatureSensor{baseSensor{c}} },
	"gpio.dht22":              func(c config.SensorConfig) Sensor { return dht22Sensor{baseSensor{c}} },
	"remote.http_json":        func(c config.SensorConfig) Sensor { return httpJsonSensor{baseSensor{c}} },
	"remote.mqtt":             func(c config.SensorConfig) Sensor { return mqttSensor{baseSensor{c}} },
}

func GetRegisteredTypes() []string {
	types := make([]string, 0, len(registry))
	for t := range registry {
		types = append(types, t)
	}
	sortStrings(types)
	return types
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func CreateSensor(cfg config.SensorConfig) (Sensor, error) {
	f, ok := registry[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("неизвестный тип датчика: %s", cfg.Type)
	}
	return f(cfg), nil
}
