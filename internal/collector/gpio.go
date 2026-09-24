//go:build linux && arm

package collector

import (
	"fmt"
	"math"
)

type dht22Sensor struct{ baseSensor }

func (s dht22Sensor) Read() SensorReading {
	pin := 4
	if s.cfg.Params != nil {
		if p, ok := s.cfg.Params["pin"].(int); ok {
			pin = p
		} else if p, ok := s.cfg.Params["pin"].(float64); ok {
			pin = int(p)
		}
	}
	value, err := readDHT22(pin)
	if err != nil {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: err.Error()}
	}
	v := math.Round(value*10) / 10
	return SensorReading{SensorID: s.ID(), Value: &v, Status: evaluateStatus(s.cfg, &v)}
}

func readDHT22(pin int) (float64, error) {
	return 0, fmt.Errorf("DHT22: установите periph.io на Raspberry Pi")
}
