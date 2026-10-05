package collector

import (
	"math"
	"testing"
)

func TestNetworkRateMbps(t *testing.T) {
	// ~204 Мбит/с over 5 s (same order as a typical speed test)
	bytesPer5s := uint64(204_000_000 / 8 * 5)
	got := networkRateMbps(0, bytesPer5s, 5)
	if math.Abs(got-204) > 1 {
		t.Fatalf("expected ~204 Mbps, got %v", got)
	}

	if networkRateMbps(100, 50, 1) != 0 {
		t.Fatal("counter reset should yield 0")
	}
}
