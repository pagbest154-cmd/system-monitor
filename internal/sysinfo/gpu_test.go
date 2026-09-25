package sysinfo

import "testing"

func TestParseNvidiaSmiCUDA(t *testing.T) {
	sample := `
| NVIDIA-SMI 535.161.08             Driver Version: 535.161.08   CUDA Version: 12.2     |
+-------------------------------+----------------------+----------------------+
`
	if v := parseNvidiaSmiCUDA(sample); v != "12.2" {
		t.Fatalf("expected 12.2, got %q", v)
	}
	if v := parseNvidiaSmiCUDA("no cuda here"); v != "" {
		t.Fatalf("expected empty, got %q", v)
	}
}

func TestParseNVCCVersion(t *testing.T) {
	sample := "Cuda compilation tools, release 12.4, V12.4.131"
	if v := parseNVCCVersion(sample); v != "12.4" {
		t.Fatalf("expected 12.4, got %q", v)
	}
}
