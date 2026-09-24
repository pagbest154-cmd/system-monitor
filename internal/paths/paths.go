package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

var (
	RootDir     string
	ConfigDir   string
	WebDir      string
	DataDir     string
	PkgConfigDir string

	SensorsConfig      string
	DashboardConfig    string
	AgentsConfig       string
	AgentConfig        string
	AgentSensorsConfig string
	AgentTokenFile     string
	AgentStatusFile    string
	AgentUpdateCache   string
	HubUpdateCache     string
	HubConfig          string
	DBPath             string
)

func init() {
	initPaths()
}

func initPaths() {
	if useDevLayout() {
		root := devRoot()
		RootDir = root
		ConfigDir = filepath.Join(root, "config")
		WebDir = filepath.Join(root, "web")
		DataDir = filepath.Join(root, "data")
		PkgConfigDir = ConfigDir
	} else if runtime.GOOS == "windows" {
		RootDir = envOr("SYSTEM_MONITOR_ROOT", windowsProgramFiles())
		ConfigDir = envOr("SYSTEM_MONITOR_CONFIG_DIR", windowsProgramData())
		WebDir = filepath.Join(RootDir, "web")
		DataDir = envOr("SYSTEM_MONITOR_DATA_DIR", filepath.Join(windowsProgramData(), "data"))
		PkgConfigDir = filepath.Join(RootDir, "config")
	} else {
		RootDir = envOr("SYSTEM_MONITOR_ROOT", "/usr/share/system-monitor")
		ConfigDir = envOr("SYSTEM_MONITOR_CONFIG_DIR", "/etc/system-monitor")
		WebDir = filepath.Join(RootDir, "web")
		DataDir = envOr("SYSTEM_MONITOR_DATA_DIR", "/var/lib/system-monitor")
		PkgConfigDir = filepath.Join(RootDir, "config")
	}

	SensorsConfig = filepath.Join(ConfigDir, "sensors.yaml")
	DashboardConfig = filepath.Join(ConfigDir, "dashboard.yaml")
	AgentsConfig = filepath.Join(ConfigDir, "agents.yaml")
	AgentConfig = filepath.Join(ConfigDir, "agent.yaml")
	AgentSensorsConfig = filepath.Join(ConfigDir, "agent_sensors.yaml")
	AgentTokenFile = filepath.Join(ConfigDir, "agent.token")
	AgentStatusFile = filepath.Join(ConfigDir, "agent.status.json")
	AgentUpdateCache = filepath.Join(ConfigDir, "agent.update.json")
	HubUpdateCache = filepath.Join(DataDir, "hub.update.json")
	HubConfig = filepath.Join(ConfigDir, "hub.yaml")
	DBPath = filepath.Join(DataDir, "metrics.db")
}

func useDevLayout() bool {
	root := devRoot()
	return dirExists(filepath.Join(root, "config")) && dirExists(filepath.Join(root, "web"))
}

func devRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func windowsProgramFiles() string {
	if v := os.Getenv("ProgramFiles"); v != "" {
		return filepath.Join(v, "system-monitor-agent")
	}
	return filepath.Join("C:", "Program Files", "system-monitor-agent")
}

func windowsProgramData() string {
	if v := os.Getenv("ProgramData"); v != "" {
		return filepath.Join(v, "system-monitor")
	}
	return filepath.Join("C:", "ProgramData", "system-monitor")
}

// OverrideForTest resets paths for unit tests.
func OverrideForTest(root string) {
	RootDir = root
	ConfigDir = filepath.Join(root, "config")
	WebDir = filepath.Join(root, "web")
	DataDir = filepath.Join(root, "data")
	PkgConfigDir = ConfigDir
	SensorsConfig = filepath.Join(ConfigDir, "sensors.yaml")
	DashboardConfig = filepath.Join(ConfigDir, "dashboard.yaml")
	AgentsConfig = filepath.Join(ConfigDir, "agents.yaml")
	AgentConfig = filepath.Join(ConfigDir, "agent.yaml")
	AgentSensorsConfig = filepath.Join(ConfigDir, "agent_sensors.yaml")
	AgentTokenFile = filepath.Join(ConfigDir, "agent.token")
	AgentStatusFile = filepath.Join(ConfigDir, "agent.status.json")
	AgentUpdateCache = filepath.Join(ConfigDir, "agent.update.json")
	HubUpdateCache = filepath.Join(DataDir, "hub.update.json")
	HubConfig = filepath.Join(ConfigDir, "hub.yaml")
	DBPath = filepath.Join(DataDir, "metrics.db")
}
