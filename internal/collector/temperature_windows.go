//go:build windows

package collector

import (
	"github.com/yusufpapurcu/wmi"
)

type acpiThermalZone struct {
	CurrentTemperature uint32
}

func readWindowsACPI() *float64 {
	var zones []acpiThermalZone
	if err := wmi.QueryNamespace("SELECT CurrentTemperature FROM MSAcpi_ThermalZoneTemperature", &zones, "root/wmi"); err != nil {
		return nil
	}
	temps := make([]float64, 0, len(zones))
	for _, zone := range zones {
		celsius := (float64(zone.CurrentTemperature)/10.0) - 273.15
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
