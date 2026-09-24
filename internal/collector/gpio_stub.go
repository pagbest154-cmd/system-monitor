//go:build !linux || !arm

package collector

type dht22Sensor struct{ baseSensor }

func (s dht22Sensor) Read() SensorReading {
	return SensorReading{
		SensorID: s.ID(),
		Status:   "error",
		Error:    "gpio.dht22 доступен только на Linux ARM (Raspberry Pi)",
	}
}
