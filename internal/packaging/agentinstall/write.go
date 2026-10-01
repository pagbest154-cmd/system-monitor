// Package agentinstall implements first-install agent.yaml generation matching
// debian/system-monitor-agent.postinst and packaging/windows/installer.iss.
package agentinstall

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultIntervalSec = 5

// EscapeYAMLDoubleQuoted matches Inno Setup EscapeYaml in installer.iss.
func EscapeYAMLDoubleQuoted(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return value
}

// LinuxDebPaths are default install locations on Linux (.deb).
type LinuxDebPaths struct {
	AgentYAML string
	TokenFile string
}

func DefaultLinuxDebPaths() LinuxDebPaths {
	return LinuxDebPaths{
		AgentYAML: "/etc/system-monitor/agent.yaml",
		TokenFile: "/etc/system-monitor/agent.token",
	}
}

// WriteLinuxDebInstallConfig writes agent.yaml and token like postinst on fresh install.
func WriteLinuxDebInstallConfig(paths LinuxDebPaths, hubURL, agentID, token string) error {
	if strings.TrimSpace(hubURL) == "" {
		hubURL = "http://127.0.0.1:8080"
	}
	if strings.TrimSpace(agentID) == "" {
		return fmt.Errorf("agent id is empty")
	}
	tokenFile := paths.TokenFile
	if tokenFile == "" {
		tokenFile = DefaultLinuxDebPaths().TokenFile
	}
	agentYAML := paths.AgentYAML
	if agentYAML == "" {
		agentYAML = DefaultLinuxDebPaths().AgentYAML
	}

	body := fmt.Sprintf(
		"hub_url: \"%s\"\nagent_id: \"%s\"\ntoken_file: %s\ninterval_sec: %d\ntransport: http\n",
		hubURL,
		agentID,
		tokenFile,
		defaultIntervalSec,
	)
	if err := os.MkdirAll(filepath.Dir(agentYAML), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(agentYAML, []byte(body), 0o644); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(tokenFile), 0o755); err != nil {
		return err
	}
	if token != "" {
		return os.WriteFile(tokenFile, []byte(token), 0o600)
	}
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return os.WriteFile(tokenFile, nil, 0o600)
	}
	return nil
}

// WindowsPaths are default install locations on Windows (installer.iss).
type WindowsPaths struct {
	ConfigDir string
}

func DefaultWindowsPaths() WindowsPaths {
	return WindowsPaths{ConfigDir: filepath.Join("ProgramData", "system-monitor")}
}

// WriteWindowsInstallConfig writes agent.yaml and token like Inno Setup WriteAgentConfig.
func WriteWindowsInstallConfig(paths WindowsPaths, hubURL, agentID, token string) error {
	if strings.TrimSpace(hubURL) == "" {
		hubURL = "http://127.0.0.1:8080"
	}
	if strings.TrimSpace(agentID) == "" {
		return fmt.Errorf("agent id is empty")
	}
	configDir := paths.ConfigDir
	if configDir == "" {
		configDir = DefaultWindowsPaths().ConfigDir
	}
	configFile := filepath.Join(configDir, "agent.yaml")
	tokenFile := filepath.Join(configDir, "agent.token")

	body := fmt.Sprintf(
		"hub_url: \"%s\"\nagent_id: \"%s\"\ntoken_file: \"%s\"\ninterval_sec: %d\ntransport: http\n",
		EscapeYAMLDoubleQuoted(hubURL),
		EscapeYAMLDoubleQuoted(agentID),
		EscapeYAMLDoubleQuoted(tokenFile),
		defaultIntervalSec,
	)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(configFile, []byte(body), 0o644); err != nil {
		return err
	}
	if token != "" {
		return os.WriteFile(tokenFile, []byte(token), 0o600)
	}
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return os.WriteFile(tokenFile, nil, 0o600)
	}
	return nil
}
