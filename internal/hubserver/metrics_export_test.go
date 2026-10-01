package hubserver

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteMetricsCSV(t *testing.T) {
	points := []map[string]interface{}{
		{"ts": 1000.5, "value": 42.0, "status": "ok"},
		{"ts": 1001.0, "value": nil, "status": "unknown"},
	}
	rec := httptest.NewRecorder()
	writeMetricsCSV(rec, "homepc:cpu_percent", "1h", points)
	body := rec.Body.String()
	if !strings.HasPrefix(body, "ts,value,status\n") {
		t.Fatalf("unexpected csv header: %q", body)
	}
	if !strings.Contains(body, "1000.5,42,ok") {
		t.Fatalf("expected data row in csv: %q", body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/csv") {
		t.Fatalf("expected csv content type, got %q", ct)
	}
}

func TestSanitizeExportFilename(t *testing.T) {
	if got := sanitizeExportFilename("a:b/c"); got != "a_b_c" {
		t.Fatalf("unexpected sanitize: %q", got)
	}
}
