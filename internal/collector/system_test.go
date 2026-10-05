package collector

import (
	"math"
	"testing"

	netps "github.com/shirou/gopsutil/v4/net"
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

func TestNetworkCounterBytesInterface(t *testing.T) {
	stats := []netps.IOCountersStat{
		{Name: "eth0", BytesRecv: 100, BytesSent: 10},
		{Name: "wlan0", BytesRecv: 200, BytesSent: 20},
	}
	rx, ok := networkCounterBytes(stats, "recv", "wlan0")
	if !ok || rx != 200 {
		t.Fatalf("wlan0 recv: got %d ok=%v", rx, ok)
	}
	_, ok = networkCounterBytes(stats, "recv", "missing")
	if ok {
		t.Fatal("expected missing interface")
	}
}
