//go:build !linux

package collector

func readLinuxHwmon() *float64 {
	return nil
}
