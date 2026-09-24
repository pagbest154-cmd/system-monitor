package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"gopkg.in/yaml.v3"
)

func CurrentPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	case "darwin":
		return "macos"
	default:
		return runtime.GOOS
	}
}

type SettingsConfig struct {
	RetentionDays      int  `yaml:"retention_days" json:"retention_days"`
	DefaultIntervalSec int  `yaml:"default_interval_sec" json:"default_interval_sec"`
	AutoDiscoverDisks  bool `yaml:"auto_discover_disks" json:"auto_discover_disks"`
}

func defaultSettings() SettingsConfig {
	return SettingsConfig{RetentionDays: 7, DefaultIntervalSec: 5, AutoDiscoverDisks: true}
}

type SensorConfig struct {
	ID            string                 `yaml:"id" json:"id"`
	Name          string                 `yaml:"name" json:"name"`
	Type          string                 `yaml:"type" json:"type"`
	Enabled       bool                   `yaml:"enabled" json:"enabled"`
	IntervalSec   *int                   `yaml:"interval_sec" json:"interval_sec,omitempty"`
	Unit          string                 `yaml:"unit" json:"unit"`
	Params        map[string]interface{} `yaml:"params" json:"params,omitempty"`
	Platforms     []string               `yaml:"platforms" json:"platforms,omitempty"`
	WarnAbove     *float64               `yaml:"warn_above" json:"warn_above,omitempty"`
	CriticalAbove *float64               `yaml:"critical_above" json:"critical_above,omitempty"`
}

func (s SensorConfig) IsSupportedOnPlatform() bool {
	if len(s.Platforms) == 0 {
		return true
	}
	platform := CurrentPlatform()
	for _, p := range s.Platforms {
		if p == platform {
			return true
		}
	}
	return false
}

type SensorsFile struct {
	Settings SettingsConfig `yaml:"settings" json:"settings"`
	Sensors  []SensorConfig `yaml:"sensors" json:"sensors"`
}

type DashboardMeta struct {
	Title      string `yaml:"title" json:"title"`
	RefreshSec int    `yaml:"refresh_sec" json:"refresh_sec"`
}

type PanelConfig struct {
	ID      string   `yaml:"id" json:"id"`
	Title   string   `yaml:"title" json:"title"`
	Type    string   `yaml:"type" json:"type"`
	Sensors []string `yaml:"sensors" json:"sensors"`
	Period  string   `yaml:"period" json:"period"`
	Col     int      `yaml:"col" json:"col"`
	Row     int      `yaml:"row" json:"row"`
	Span    int      `yaml:"span" json:"span"`
}

type DashboardFile struct {
	Dashboard DashboardMeta `yaml:"dashboard" json:"dashboard"`
	Panels    []PanelConfig `yaml:"panels" json:"panels"`
}

type SensorOverrideConfig struct {
	SensorID      string                 `yaml:"sensor_id" json:"sensor_id"`
	Enabled       *bool                  `yaml:"enabled" json:"enabled,omitempty"`
	IntervalSec   *int                   `yaml:"interval_sec" json:"interval_sec,omitempty"`
	WarnAbove     *float64               `yaml:"warn_above" json:"warn_above,omitempty"`
	CriticalAbove *float64               `yaml:"critical_above" json:"critical_above,omitempty"`
	Params        map[string]interface{} `yaml:"params" json:"params,omitempty"`
}

type AgentEntry struct {
	ID        string                 `yaml:"id" json:"id"`
	Name      string                 `yaml:"name" json:"name"`
	Token     string                 `yaml:"token" json:"token"`
	Overrides []SensorOverrideConfig `yaml:"overrides" json:"overrides,omitempty"`
}

type AgentsFile struct {
	Agents []AgentEntry `yaml:"agents" json:"agents"`
}

type HubSettings struct {
	Domain        string   `yaml:"domain" json:"domain"`
	PublicURL     string   `yaml:"public_url" json:"public_url"`
	UseHTTPS      bool     `yaml:"use_https" json:"use_https"`
	TrustedHosts  []string `yaml:"trusted_hosts" json:"trusted_hosts"`
}

type HubFile struct {
	Hub HubSettings `yaml:"hub" json:"hub"`
}

type AgentFileConfig struct {
	HubURL      string `yaml:"hub_url" json:"hub_url"`
	AgentID     string `yaml:"agent_id" json:"agent_id"`
	Token       string `yaml:"token" json:"token"`
	TokenFile   string `yaml:"token_file" json:"token_file"`
	IntervalSec int    `yaml:"interval_sec" json:"interval_sec"`
	Transport   string `yaml:"transport" json:"transport"`
}

func loadYAML(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, err
	}
	var out map[string]interface{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("config %s: %w", path, err)
	}
	if out == nil {
		return map[string]interface{}{}, nil
	}
	return out, nil
}

func saveYAML(path string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func decodeYAML[T any](path string, target *T) error {
	raw, err := loadYAML(path)
	if err != nil {
		return err
	}
	buf, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buf, target)
}

func LoadSensorsConfig(path string) (*SensorsFile, error) {
	if path == "" {
		path = paths.SensorsConfig
	}
	cfg := &SensorsFile{Settings: defaultSettings()}
	if err := decodeYAML(path, cfg); err != nil {
		return nil, err
	}
	if cfg.Settings.RetentionDays == 0 {
		cfg.Settings = defaultSettings()
	}
	return cfg, nil
}

func SaveSensorsConfig(cfg *SensorsFile, path string) error {
	if path == "" {
		path = paths.SensorsConfig
	}
	return saveYAML(path, cfg)
}

func defaultPanelSensors() map[string][]string {
	data, err := loadYAML(filepath.Join(paths.PkgConfigDir, "dashboard.yaml"))
	if err != nil {
		return map[string][]string{}
	}
	panels, _ := data["panels"].([]interface{})
	mapping := map[string][]string{}
	for _, p := range panels {
		panel, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := panel["id"].(string)
		sensorsRaw, _ := panel["sensors"].([]interface{})
		if id == "" || len(sensorsRaw) == 0 {
			continue
		}
		sensors := make([]string, 0, len(sensorsRaw))
		for _, s := range sensorsRaw {
			sensors = append(sensors, fmt.Sprint(s))
		}
		mapping[id] = sensors
	}
	return mapping
}

func EnrichDashboardPanels(panels []PanelConfig) []PanelConfig {
	defaults := defaultPanelSensors()
	enriched := make([]PanelConfig, len(panels))
	for i, panel := range panels {
		if len(panel.Sensors) == 0 && defaults[panel.ID] != nil {
			panel.Sensors = defaults[panel.ID]
		}
		enriched[i] = panel
	}
	return enriched
}

func LoadDashboardConfig(path string) (*DashboardFile, error) {
	if path == "" {
		path = paths.DashboardConfig
	}
	cfg := &DashboardFile{
		Dashboard: DashboardMeta{Title: "Мониторинг системы", RefreshSec: 3},
	}
	if err := decodeYAML(path, cfg); err != nil {
		return nil, err
	}
	cfg.Panels = EnrichDashboardPanels(cfg.Panels)
	return cfg, nil
}

func SaveDashboardConfig(cfg *DashboardFile, path string) error {
	if path == "" {
		path = paths.DashboardConfig
	}
	return saveYAML(path, cfg)
}

func DefaultDiskPath() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}
	return "/"
}

func LoadAgentsConfig(path string) (*AgentsFile, error) {
	if path == "" {
		path = paths.AgentsConfig
	}
	cfg := &AgentsFile{}
	if err := decodeYAML(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func SaveAgentsConfig(cfg *AgentsFile, path string) error {
	if path == "" {
		path = paths.AgentsConfig
	}
	return saveYAML(path, cfg)
}

func NormalizeDomain(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(value, prefix) {
			value = value[len(prefix):]
		}
	}
	if idx := strings.Index(value, "/"); idx >= 0 {
		value = value[:idx]
	}
	return strings.TrimSpace(value)
}

func ResolvePublicURL(settings HubSettings) string {
	if strings.TrimSpace(settings.PublicURL) != "" {
		return strings.TrimRight(strings.TrimSpace(settings.PublicURL), "/")
	}
	domain := NormalizeDomain(settings.Domain)
	if domain == "" {
		return ""
	}
	scheme := "https"
	if !settings.UseHTTPS {
		scheme = "http"
	}
	return scheme + "://" + domain
}

func LoadHubConfig(path string) (*HubFile, error) {
	if path == "" {
		path = paths.HubConfig
	}
	cfg := &HubFile{Hub: HubSettings{UseHTTPS: true, TrustedHosts: []string{"*"}}}
	if err := decodeYAML(path, cfg); err != nil {
		return nil, err
	}
	if len(cfg.Hub.TrustedHosts) == 0 {
		cfg.Hub.TrustedHosts = []string{"*"}
	}
	return cfg, nil
}

func SaveHubConfig(cfg *HubFile, path string) error {
	if path == "" {
		path = paths.HubConfig
	}
	return saveYAML(path, cfg)
}

func HubTrustedHosts(settings HubSettings) []string {
	hosts := make([]string, 0)
	seen := map[string]bool{}
	for _, item := range settings.TrustedHosts {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		hosts = append(hosts, item)
	}
	domain := NormalizeDomain(settings.Domain)
	if domain != "" {
		for _, h := range []string{domain, "*." + domain} {
			if !seen[h] {
				seen[h] = true
				hosts = append(hosts, h)
			}
		}
	}
	if len(hosts) == 0 {
		return []string{"*"}
	}
	return hosts
}

func normalizeTokenFile(tokenFile string) string {
	value := strings.TrimSpace(tokenFile)
	if value == "" {
		return paths.AgentTokenFile
	}
	if _, err := os.Stat(value); err == nil {
		return value
	}
	if _, err := os.Stat(paths.AgentTokenFile); err == nil {
		return paths.AgentTokenFile
	}
	return value
}

func LoadAgentConfig(path string) (*AgentFileConfig, error) {
	if path == "" {
		path = paths.AgentConfig
	}
	cfg := &AgentFileConfig{
		HubURL:      "http://127.0.0.1:8080",
		IntervalSec: 5,
		Transport:   "http",
	}
	if err := decodeYAML(path, cfg); err != nil {
		return nil, err
	}
	cfg.TokenFile = normalizeTokenFile(cfg.TokenFile)
	return cfg, nil
}

func SaveAgentConfig(cfg *AgentFileConfig, path string) error {
	if path == "" {
		path = paths.AgentConfig
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return saveYAML(path, cfg)
}

func LoadAgentToken(cfg *AgentFileConfig) string {
	if cfg == nil {
		c, err := LoadAgentConfig("")
		if err != nil {
			return ""
		}
		cfg = c
	}
	if strings.TrimSpace(cfg.Token) != "" {
		return strings.TrimSpace(cfg.Token)
	}
	tokenPath := cfg.TokenFile
	if strings.TrimSpace(tokenPath) == "" {
		tokenPath = paths.AgentTokenFile
	}
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func SaveAgentToken(token string, path string) error {
	if path == "" {
		path = paths.AgentTokenFile
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.TrimSpace(token)), 0o600)
}

func LoadAgentSensorsConfig(path string) (*SensorsFile, error) {
	if path == "" {
		path = paths.AgentSensorsConfig
	}
	if _, err := os.Stat(path); err == nil {
		return LoadSensorsConfig(path)
	}
	bundled := filepath.Join(paths.PkgConfigDir, "agent_sensors.yaml")
	if _, err := os.Stat(bundled); err == nil {
		return LoadSensorsConfig(bundled)
	}
	return LoadSensorsConfig(paths.SensorsConfig)
}

func SaveAgentSensorsConfig(cfg *SensorsFile, path string) error {
	if path == "" {
		path = paths.AgentSensorsConfig
	}
	return SaveSensorsConfig(cfg, path)
}

func MergeAgentConfig(base *SensorsFile, overrides []SensorOverrideConfig) *SensorsFile {
	if len(overrides) == 0 {
		return base
	}
	overrideMap := map[string]SensorOverrideConfig{}
	for _, o := range overrides {
		overrideMap[o.SensorID] = o
	}
	merged := make([]SensorConfig, 0, len(base.Sensors))
	for _, sensor := range base.Sensors {
		override, ok := overrideMap[sensor.ID]
		if !ok {
			merged = append(merged, sensor)
			continue
		}
		s := sensor
		if override.Enabled != nil {
			s.Enabled = *override.Enabled
		}
		if override.IntervalSec != nil {
			s.IntervalSec = override.IntervalSec
		}
		if override.WarnAbove != nil {
			s.WarnAbove = override.WarnAbove
		}
		if override.CriticalAbove != nil {
			s.CriticalAbove = override.CriticalAbove
		}
		if override.Params != nil {
			if s.Params == nil {
				s.Params = map[string]interface{}{}
			}
			for k, v := range override.Params {
				s.Params[k] = v
			}
		}
		merged = append(merged, s)
	}
	return &SensorsFile{Settings: base.Settings, Sensors: merged}
}

func FindAgentEntry(agentID string, cfg *AgentsFile) *AgentEntry {
	if cfg == nil {
		c, err := LoadAgentsConfig("")
		if err != nil {
			return nil
		}
		cfg = c
	}
	for _, entry := range cfg.Agents {
		if entry.ID == agentID {
			return &entry
		}
	}
	return nil
}

var placeholderTokens = map[string]bool{"": true, "change-me": true, "changeme": true}

func needsNewToken(token string) bool {
	return placeholderTokens[strings.ToLower(strings.TrimSpace(token))]
}

func GenerateAgentToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("token-%d", os.Getpid())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func PrepareAgentsForSave(incoming *AgentsFile) (*AgentsFile, error) {
	previous, _ := LoadAgentsConfig("")
	prevMap := map[string]AgentEntry{}
	if previous != nil {
		for _, entry := range previous.Agents {
			prevMap[entry.ID] = entry
		}
	}
	prepared := make([]AgentEntry, 0, len(incoming.Agents))
	for _, agent := range incoming.Agents {
		token := strings.TrimSpace(agent.Token)
		if needsNewToken(token) {
			if prev, ok := prevMap[agent.ID]; ok && !needsNewToken(prev.Token) {
				token = prev.Token
			} else {
				token = GenerateAgentToken()
			}
		}
		agent.Token = token
		prepared = append(prepared, agent)
	}
	return &AgentsFile{Agents: prepared}, nil
}
