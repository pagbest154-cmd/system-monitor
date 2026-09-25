package agent

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/collector"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/diskdiscovery"
	"github.com/pagbest154-cmd/system-monitor/internal/fleet"
	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
	"github.com/pagbest154-cmd/system-monitor/internal/sysinfo"
	"github.com/pagbest154-cmd/system-monitor/internal/version"
)

type Runner struct {
	ConfigPath     string
	Config         *config.AgentFileConfig
	Transport      *Transport
	AgentID        string
	configVersion  int
	configMTime    time.Time
	sensorsMTime   time.Time
	lastUpdateCheck float64
	collector      *collector.Collector
	stop            chan struct{}
	done            chan struct{}
	status          Status
	wasConnected    bool
	pushErrorCount  int
	lastPushError   string
	syncErrorCount  int
	lastSyncError   string
}

func NewRunner(configPath string) (*Runner, error) {
	cfg, err := config.LoadAgentConfig(configPath)
	if err != nil {
		return nil, err
	}
	token := config.LoadAgentToken(cfg)
	transport := NewTransport(cfg.HubURL, token)
	hostname, _ := os.Hostname()
	r := &Runner{
		ConfigPath: configPath,
		Config:     cfg,
		Transport:  transport,
		AgentID:    fleet.DefaultAgentID(cfg.AgentID),
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
		status: Status{
			AgentID: cfg.AgentID, HubURL: cfg.HubURL, Hostname: hostname,
		},
	}
	if info, err := os.Stat(configPath); err == nil {
		r.configMTime = info.ModTime()
	}
	if info, err := os.Stat(paths.AgentSensorsConfig); err == nil {
		r.sensorsMTime = info.ModTime()
	}
	return r, nil
}

func (r *Runner) Start() {
	go r.run()
}

func (r *Runner) Stop() {
	close(r.stop)
	<-r.done
	if r.collector != nil {
		r.collector.Stop()
	}
	r.Transport.Close()
}

func (r *Runner) run() {
	defer close(r.done)
	AgentLogf(
		"starting v%s agent_id=%s hub=%s interval=%ds config=%s",
		version.Version, r.AgentID, r.Config.HubURL, r.Config.IntervalSec, r.ConfigPath,
	)
	r.ensureCollector()
	if err := r.reloadCollector(); err != nil {
		r.logSyncError("initial config sync", err)
	} else {
		AgentLogf("config synced version=%d", r.configVersion)
	}
	r.writeStatus(map[string]interface{}{"connected": false})
	r.maybeCheckUpdates(time.Now().Unix())
	lastConfigSync := time.Now().Unix()
	for {
		select {
		case <-r.stop:
			return
		default:
		}
		now := time.Now().Unix()
		r.maybeReloadConfig()
		r.maybeReloadSensors()
		r.maybeCheckUpdates(now)
		if now-lastConfigSync > 60 {
			if err := r.reloadCollector(); err != nil {
				r.logSyncError("periodic config sync", err)
				errStr := err.Error()
				r.writeStatus(map[string]interface{}{
					"connected": false, "last_error": errStr, "last_error_ts": float64(now),
				})
			} else {
				r.syncErrorCount = 0
				r.lastSyncError = ""
				AgentLogf("config synced version=%d", r.configVersion)
			}
			lastConfigSync = now
		}
		count, err := r.pushOnce()
		if err != nil {
			r.logPushError(err)
			errStr := err.Error()
			r.writeStatus(map[string]interface{}{
				"connected": false, "last_error": errStr, "last_error_ts": float64(now),
			})
			r.wasConnected = false
		} else {
			if !r.wasConnected {
				AgentLogf("hub connected, pushed %d metrics to %s", count, r.Config.HubURL)
			}
			r.wasConnected = true
			r.pushErrorCount = 0
			r.lastPushError = ""
			r.writeStatus(map[string]interface{}{
				"connected": true, "last_success_ts": float64(time.Now().Unix()),
				"last_error": nil, "last_error_ts": nil, "metrics_count": count,
			})
		}
		interval := r.Config.IntervalSec
		if interval < 1 {
			interval = 1
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func (r *Runner) writeStatus(updates map[string]interface{}) {
	if v, ok := updates["connected"].(bool); ok {
		r.status.Connected = v
	}
	if v, ok := updates["last_success_ts"].(float64); ok {
		r.status.LastSuccessTS = v
	}
	if v, ok := updates["last_error"].(string); ok {
		r.status.LastError = &v
	}
	if updates["last_error"] == nil {
		r.status.LastError = nil
	}
	if v, ok := updates["last_error_ts"].(float64); ok {
		r.status.LastErrorTS = v
	}
	if v, ok := updates["metrics_count"].(int); ok {
		r.status.MetricsCount = v
	}
	if v, ok := updates["update_available"].(bool); ok {
		r.status.UpdateAvailable = v
	}
	if v, ok := updates["latest_version"].(string); ok {
		r.status.LatestVersion = v
	}
	if v, ok := updates["release_url"].(string); ok {
		r.status.ReleaseURL = v
	}
	if v, ok := updates["update_checked_at"].(float64); ok {
		r.status.UpdateCheckedAt = v
	}
	_ = WriteStatus(r.status)
}

func (r *Runner) maybeReloadConfig() {
	info, err := os.Stat(r.ConfigPath)
	if err != nil || info.ModTime().Equal(r.configMTime) {
		return
	}
	r.configMTime = info.ModTime()
	cfg, err := config.LoadAgentConfig(r.ConfigPath)
	if err != nil {
		AgentLogf("config reload failed: %v", err)
		return
	}
	token := config.LoadAgentToken(cfg)
	agentID := fleet.DefaultAgentID(cfg.AgentID)
	if cfg.HubURL == r.Config.HubURL && cfg.IntervalSec == r.Config.IntervalSec &&
		agentID == r.AgentID && token == r.Transport.Token {
		r.Config = cfg
		return
	}
	r.Transport.Close()
	r.Config = cfg
	r.AgentID = agentID
	r.Transport = NewTransport(cfg.HubURL, token)
	r.status.AgentID = agentID
	r.status.HubURL = cfg.HubURL
	AgentLogf("config reloaded agent_id=%s hub=%s interval=%ds", agentID, cfg.HubURL, cfg.IntervalSec)
}

func (r *Runner) maybeReloadSensors() {
	info, err := os.Stat(paths.AgentSensorsConfig)
	if err != nil || info.ModTime().Equal(r.sensorsMTime) {
		return
	}
	r.sensorsMTime = info.ModTime()
	if err := r.reloadCollector(); err != nil {
		AgentLogf("sensor reload failed: %v", err)
		return
	}
	AgentLogf("agent_sensors.yaml reloaded")
}

func (r *Runner) maybeCheckUpdates(now int64) {
	if float64(now)-r.lastUpdateCheck < float64(UpdateCheckIntervalSec) {
		return
	}
	r.lastUpdateCheck = float64(now)
	result := CheckForUpdates(false)
	if result.Error != nil && *result.Error != "" {
		AgentLogf("update check failed: %s", *result.Error)
	} else if result.UpdateAvailable {
		AgentLogf("update available: %s (current %s)", result.LatestVersion, result.CurrentVersion)
	}
	r.writeStatus(map[string]interface{}{
		"update_available": result.UpdateAvailable,
		"latest_version":   result.LatestVersion,
		"release_url":      result.ReleaseURL,
		"update_checked_at": result.CheckedAt,
	})
}

func (r *Runner) reloadCollector() error {
	base, err := config.LoadAgentSensorsConfig("")
	if err != nil {
		return err
	}
	remote, err := r.Transport.SyncConfig(r.AgentID, r.configVersion)
	if err != nil {
		return err
	}
	r.configVersion = remote.Version
	overrides := make([]config.SensorOverrideConfig, 0, len(remote.Overrides))
	for _, item := range remote.Overrides {
		overrides = append(overrides, config.SensorOverrideConfig{
			SensorID: item.SensorID, Enabled: item.Enabled, IntervalSec: item.IntervalSec,
			WarnAbove: item.WarnAbove, CriticalAbove: item.CriticalAbove, Params: item.Params,
		})
	}
	merged := config.MergeAgentConfig(base, overrides)
	if r.collector == nil {
		r.collector = collector.NewCollector(nil, nil, merged, nil)
		r.collector.Start()
	} else {
		r.collector.ReloadConfig(merged)
	}
	return nil
}

func (r *Runner) ensureCollector() {
	if r.collector != nil {
		return
	}
	cfg, _ := config.LoadAgentSensorsConfig("")
	r.collector = collector.NewCollector(nil, nil, cfg, nil)
	r.collector.Start()
}

func (r *Runner) pushOnce() (int, error) {
	r.ensureCollector()
	snapshot := r.collector.GetLatestSnapshot()
	systemInfo := sysinfo.GetSystemInfo()
	systemSnapshot := r.snapshotFromSystem(systemInfo)
	if !snapshotHasValues(snapshot) {
		snapshot = systemSnapshot
	} else {
		snapshot = mergeSnapshot(snapshot, systemSnapshot)
	}
	metrics := make([]protocol.MetricPoint, 0, len(snapshot))
	for sensorID, item := range snapshot {
		value := metricValue(item)
		ts := float64(time.Now().Unix())
		if v, ok := item["ts"].(float64); ok {
			ts = v
		}
		status := "unknown"
		if v, ok := item["status"].(string); ok {
			status = v
		}
		sid := sensorID
		if v, ok := item["sensor_id"].(string); ok {
			sid = v
		}
		metrics = append(metrics, protocol.MetricPoint{
			SensorID: sid, TS: ts, Value: value, Status: status,
		})
	}
	sensors := make([]protocol.SensorMeta, 0)
	for _, item := range r.collector.BuildSensorMeta() {
		sensors = append(sensors, protocol.SensorMeta{
			ID: fmt.Sprint(item["id"]), Name: fmt.Sprint(item["name"]),
			Type: fmt.Sprint(item["type"]), Unit: fmt.Sprint(item["unit"]), Enabled: true,
		})
	}
	hostname, _ := os.Hostname()
	report := &protocol.AgentReport{
		AgentID: r.AgentID, Hostname: hostname, Metrics: metrics,
		AgentVersion: version.Version, Platform: runtime.GOOS,
		System: systemInfo, ConfigVersion: r.configVersion, Sensors: sensors,
	}
	if err := r.Transport.PushReport(report); err != nil {
		return 0, err
	}
	return len(metrics), nil
}

func snapshotHasValues(snapshot map[string]map[string]interface{}) bool {
	for _, item := range snapshot {
		if item["value"] != nil {
			return true
		}
	}
	return false
}

func mergeSnapshot(primary, fallback map[string]map[string]interface{}) map[string]map[string]interface{} {
	merged := map[string]map[string]interface{}{}
	for k, v := range primary {
		merged[k] = v
	}
	for sensorID, reading := range fallback {
		current := merged[sensorID]
		if current == nil || current["value"] == nil {
			merged[sensorID] = reading
		}
	}
	return merged
}

func (r *Runner) snapshotFromSystem(info map[string]interface{}) map[string]map[string]interface{} {
	now := float64(time.Now().Unix())
	snapshot := map[string]map[string]interface{}{}
	cpu, _ := info["cpu"].(map[string]interface{})
	if cpu != nil && cpu["percent"] != nil {
		v, _ := cpu["percent"].(float64)
		snapshot["cpu_percent"] = map[string]interface{}{
			"sensor_id": "cpu_percent", "value": v, "status": "ok", "ts": now, "unit": "%",
		}
	}
	mem, _ := info["memory"].(map[string]interface{})
	if mem != nil && mem["percent"] != nil {
		v, _ := mem["percent"].(float64)
		snapshot["ram_used"] = map[string]interface{}{
			"sensor_id": "ram_used", "value": v, "status": "ok", "ts": now, "unit": "%",
		}
	}
	for _, key := range []string{"disks", "partitions"} {
		parts, _ := info[key].([]map[string]interface{})
		if parts == nil {
			if raw, ok := info[key].([]interface{}); ok {
				for _, item := range raw {
					if m, ok := item.(map[string]interface{}); ok {
						addDiskSnapshot(snapshot, m, now)
					}
				}
			}
			continue
		}
		for _, part := range parts {
			addDiskSnapshot(snapshot, part, now)
		}
	}
	return snapshot
}

func addDiskSnapshot(snapshot map[string]map[string]interface{}, part map[string]interface{}, now float64) {
	mount := fmt.Sprint(part["mountpoint"])
	percent, ok := part["percent"].(float64)
	if !ok {
		return
	}
	sensorID := diskdiscovery.MountToSensorID(mount)
	snapshot[sensorID] = map[string]interface{}{
		"sensor_id": sensorID, "value": percent, "status": "ok", "ts": now, "unit": "%",
	}
}

func (r *Runner) logPushError(err error) {
	r.pushErrorCount++
	msg := err.Error()
	if msg == r.lastPushError && r.pushErrorCount%12 != 1 {
		return
	}
	r.lastPushError = msg
	AgentLogf("metrics push failed (attempt %d): %v", r.pushErrorCount, err)
}

func (r *Runner) logSyncError(stage string, err error) {
	r.syncErrorCount++
	msg := err.Error()
	if msg == r.lastSyncError && r.syncErrorCount%3 != 1 {
		return
	}
	r.lastSyncError = msg
	AgentLogf("%s failed (attempt %d): %v", stage, r.syncErrorCount, err)
}

func metricValue(item map[string]interface{}) *float64 {
	if item == nil {
		return nil
	}
	switch v := item["value"].(type) {
	case float64:
		return &v
	case *float64:
		return v
	case int:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	default:
		return nil
	}
}
