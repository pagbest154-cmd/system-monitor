//go:build !windows

package collector

func readWindowsACPI() *float64 {
	return nil
}
