package collector

import (
	"sync"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/diskdiscovery"
	"github.com/pagbest154-cmd/system-monitor/internal/storage"
	"github.com/shirou/gopsutil/v4/cpu"
)

type UpdateCallback func(map[string]interface{})

type Collector struct {
	store        *storage.MetricStore
	onUpdate     UpdateCallback
	configLoader func() (*config.SensorsFile, error)
	stop         chan struct{}
	done         chan struct{}
	mu           sync.Mutex
	cfg          *config.SensorsFile
	sensors      map[string]Sensor
	lastPoll     map[string]float64
	latest       map[string]map[string]interface{}
}

func NewCollector(store *storage.MetricStore, onUpdate UpdateCallback, cfg *config.SensorsFile, loader func() (*config.SensorsFile, error)) *Collector {
	if loader == nil {
		loader = func() (*config.SensorsFile, error) { return config.LoadSensorsConfig("") }
	}
	c := &Collector{
		store: store, onUpdate: onUpdate, configLoader: loader,
		stop: make(chan struct{}), done: make(chan struct{}),
		sensors: map[string]Sensor{}, lastPoll: map[string]float64{},
		latest: map[string]map[string]interface{}{},
	}
	if cfg != nil {
		c.cfg = cfg
	} else {
		c.cfg, _ = loader()
	}
	c.ReloadConfig(nil)
	return c
}

func (c *Collector) ReloadConfig(cfg *config.SensorsFile) {
	if cfg != nil {
		c.cfg = cfg
	} else {
		c.cfg, _ = c.configLoader()
	}
	sensors := map[string]Sensor{}
	configuredPaths := map[string]bool{}
	for _, item := range c.cfg.Sensors {
		if item.Type == "system.disk_usage" && item.Params != nil {
			if p, ok := item.Params["path"].(string); ok {
				configuredPaths[p] = true
			}
		}
	}
	for _, item := range c.cfg.Sensors {
		if !item.Enabled || !item.IsSupportedOnPlatform() {
			continue
		}
		sensor, err := CreateSensor(item)
		if err != nil {
			continue
		}
		sensors[item.ID] = sensor
	}
	if c.cfg.Settings.AutoDiscoverDisks {
		for _, item := range diskdiscovery.DiscoverDiskSensors() {
			path, _ := item.Params["path"].(string)
			if configuredPaths[path] {
				continue
			}
			sensor, err := CreateSensor(item)
			if err != nil {
				continue
			}
			sensors[item.ID] = sensor
		}
	}
	c.mu.Lock()
	c.sensors = sensors
	c.mu.Unlock()
}

func (c *Collector) Start() {
	go c.run()
}

func (c *Collector) Stop() {
	close(c.stop)
	<-c.done
}

func (c *Collector) run() {
	_, _ = cpu.Percent(0, false)
	lastCleanup := 0.0
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	defer close(c.done)
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			now := float64(time.Now().UnixNano()) / 1e9
			updates := map[string]interface{}{}
			c.mu.Lock()
			sensors := make(map[string]Sensor, len(c.sensors))
			for k, v := range c.sensors {
				sensors[k] = v
			}
			c.mu.Unlock()
			for sensorID, sensor := range sensors {
				cfg := sensor.Config()
				interval := c.intervalFor(cfg)
				last := c.lastPoll[sensorID]
				if now-last < float64(interval) {
					continue
				}
				reading := c.safeRead(sensor)
				if c.store != nil {
					_ = c.store.Insert(sensorID, reading.Value, reading.Status, now)
				}
				payload := reading.ToMap()
				payload["name"] = cfg.Name
				payload["unit"] = cfg.Unit
				payload["type"] = cfg.Type
				payload["ts"] = now
				c.mu.Lock()
				c.latest[sensorID] = payload
				c.mu.Unlock()
				updates[sensorID] = payload
				c.lastPoll[sensorID] = now
			}
			if len(updates) > 0 && c.onUpdate != nil {
				c.onUpdate(updates)
			}
			if c.store != nil && now-lastCleanup > 3600 {
				_, _ = c.store.Cleanup(c.cfg.Settings.RetentionDays)
				lastCleanup = now
			}
		}
	}
}

func (c *Collector) intervalFor(cfg config.SensorConfig) int {
	if cfg.IntervalSec != nil {
		return *cfg.IntervalSec
	}
	return c.cfg.Settings.DefaultIntervalSec
}

func (c *Collector) safeRead(sensor Sensor) SensorReading {
	reading := sensor.Read()
	if reading.Status == "" {
		reading.Status = "ok"
	}
	return reading
}

func (c *Collector) GetLatestSnapshot() map[string]map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]map[string]interface{}, len(c.latest))
	for k, v := range c.latest {
		out[k] = v
	}
	return out
}

func (c *Collector) GetSensorConfigs() []config.SensorConfig {
	configs := append([]config.SensorConfig{}, c.cfg.Sensors...)
	configuredPaths := map[string]bool{}
	for _, cfg := range configs {
		if cfg.Type == "system.disk_usage" && cfg.Params != nil {
			if p, ok := cfg.Params["path"].(string); ok {
				configuredPaths[p] = true
			}
		}
	}
	for _, item := range diskdiscovery.DiscoverDiskSensors() {
		path, _ := item.Params["path"].(string)
		if configuredPaths[path] {
			continue
		}
		configs = append(configs, item)
	}
	return configs
}

func (c *Collector) GetSettings() map[string]interface{} {
	return map[string]interface{}{
		"retention_days":       c.cfg.Settings.RetentionDays,
		"default_interval_sec": c.cfg.Settings.DefaultIntervalSec,
		"auto_discover_disks":  c.cfg.Settings.AutoDiscoverDisks,
	}
}

func (c *Collector) GetActiveSensorIDs() map[string]bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]bool, len(c.sensors))
	for id := range c.sensors {
		out[id] = true
	}
	return out
}

func (c *Collector) BuildSensorMeta() []map[string]interface{} {
	active := c.GetActiveSensorIDs()
	items := make([]map[string]interface{}, 0)
	for _, cfg := range c.GetSensorConfigs() {
		if !active[cfg.ID] {
			continue
		}
		item := map[string]interface{}{
			"id": cfg.ID, "name": cfg.Name, "type": cfg.Type, "unit": cfg.Unit,
			"enabled": cfg.Enabled, "params": cfg.Params,
		}
		if cfg.IntervalSec != nil {
			item["interval_sec"] = *cfg.IntervalSec
		}
		if cfg.WarnAbove != nil {
			item["warn_above"] = *cfg.WarnAbove
		}
		if cfg.CriticalAbove != nil {
			item["critical_above"] = *cfg.CriticalAbove
		}
		items = append(items, item)
	}
	return items
}
