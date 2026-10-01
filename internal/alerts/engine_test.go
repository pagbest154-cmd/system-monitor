package alerts

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/storage"
)

type mockSender struct {
	mu    sync.Mutex
	sends []mockSend
}

type mockSend struct {
	title    string
	body     string
	severity string
}

func (m *mockSender) Send(topic, token, title, body string, data map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sends = append(m.sends, mockSend{title: title, body: body, severity: data["severity"]})
	return nil
}

func (m *mockSender) last() mockSend {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sends) == 0 {
		return mockSend{}
	}
	return m.sends[len(m.sends)-1]
}

func (m *mockSender) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sends)
}

func setupEngineTest(t *testing.T) (*Engine, *mockSender, func()) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	paths.OverrideForTest(root)
	store, err := storage.NewMetricStore(paths.DBPath)
	if err != nil {
		t.Fatal(err)
	}
	sender := &mockSender{}
	engine := NewEngine(store, sender)
	return engine, sender, func() {}
}

func writeAlerts(t *testing.T, cfg *config.AlertsFile) {
	t.Helper()
	if err := config.SaveAlertsConfig(cfg, paths.AlertsConfig); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateMetricManualThreshold(t *testing.T) {
	engine, sender, cleanup := setupEngineTest(t)
	defer cleanup()

	writeAlerts(t, &config.AlertsFile{
		Alerts: map[string]config.AgentAlertConfig{
			"host1": {
				Enabled:       true,
				ThresholdMode: config.AlertThresholdModeManual,
				CooldownSec:   0,
				Ntfy:          config.NtfyAlertConfig{Topic: "test-topic"},
				Sensors: []config.AlertSensorRule{
					{SensorID: "cpu_percent", Enabled: true, Threshold: 80},
				},
			},
		},
	})

	engine.EvaluateMetric(MetricInput{
		AgentID: "host1", AgentName: "Host", SensorID: "cpu_percent",
		SensorName: "CPU", Unit: "%", Value: 85,
	})
	if sender.count() != 1 {
		t.Fatalf("expected 1 send, got %d", sender.count())
	}
	if sender.last().severity != "warning" {
		t.Fatalf("expected warning severity, got %q", sender.last().severity)
	}

	engine.EvaluateMetric(MetricInput{
		AgentID: "host1", AgentName: "Host", SensorID: "cpu_percent",
		SensorName: "CPU", Unit: "%", Value: 85,
	})
	if sender.count() != 1 {
		t.Fatalf("expected no duplicate alert, got %d sends", sender.count())
	}
}

func TestEvaluateMetricSensorThresholdsCritical(t *testing.T) {
	engine, sender, cleanup := setupEngineTest(t)
	defer cleanup()

	warn := 80.0
	critical := 95.0
	writeAlerts(t, &config.AlertsFile{
		Alerts: map[string]config.AgentAlertConfig{
			"host1": {
				Enabled:       true,
				ThresholdMode: config.AlertThresholdModeSensor,
				CooldownSec:   0,
				Ntfy:          config.NtfyAlertConfig{Topic: "test-topic"},
				Sensors: []config.AlertSensorRule{
					{SensorID: "cpu_percent", Enabled: true},
				},
			},
		},
	})

	engine.EvaluateMetric(MetricInput{
		AgentID: "host1", AgentName: "Host", SensorID: "cpu_percent",
		SensorName: "CPU", Unit: "%", Value: 82,
		WarnAbove: &warn, CriticalAbove: &critical,
	})
	if sender.count() != 1 || sender.last().severity != "warning" {
		t.Fatalf("expected warning alert, got %+v", sender.last())
	}

	engine.EvaluateMetric(MetricInput{
		AgentID: "host1", AgentName: "Host", SensorID: "cpu_percent",
		SensorName: "CPU", Unit: "%", Value: 96,
		WarnAbove: &warn, CriticalAbove: &critical,
	})
	if sender.count() != 2 || sender.last().severity != "critical" {
		t.Fatalf("expected critical alert, got sends=%d last=%+v", sender.count(), sender.last())
	}
}

func TestNormalizeAlertThresholdMode(t *testing.T) {
	if config.NormalizeAlertThresholdMode("sensor") != config.AlertThresholdModeSensor {
		t.Fatal("expected sensor mode")
	}
	if config.NormalizeAlertThresholdMode("unknown") != config.AlertThresholdModeManual {
		t.Fatal("expected manual fallback")
	}
}
