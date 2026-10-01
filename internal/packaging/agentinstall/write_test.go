package agentinstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
)

var installFixture = struct {
	hubURL  string
	agentID string
	token   string
}{
	hubURL:  "https://hub.example:8443/monitor",
	agentID: "agent-01_special",
	token:   "Bearer tok\"en\\with\\slashes",
}

func TestEscapeYAMLDoubleQuoted(t *testing.T) {
	got := EscapeYAMLDoubleQuoted(`a\b"c`)
	want := `a\\b\"c`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLinuxDebInstallConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	agentYAML := filepath.Join(dir, "agent.yaml")
	tokenFile := filepath.Join(dir, "agent.token")
	f := installFixture
	if err := WriteLinuxDebInstallConfig(LinuxDebPaths{
		AgentYAML: agentYAML,
		TokenFile: tokenFile,
	}, f.hubURL, f.agentID, f.token); err != nil {
		t.Fatal(err)
	}
	assertInstallRoundTrip(t, agentYAML, tokenFile, f.hubURL, f.agentID, f.token)
}

func TestWindowsInstallConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	f := installFixture
	if err := WriteWindowsInstallConfig(WindowsPaths{ConfigDir: dir}, f.hubURL, f.agentID, f.token); err != nil {
		t.Fatal(err)
	}
	agentYAML := filepath.Join(dir, "agent.yaml")
	tokenFile := filepath.Join(dir, "agent.token")
	assertInstallRoundTrip(t, agentYAML, tokenFile, f.hubURL, f.agentID, f.token)
}

func TestLinuxDebMatchesPostinstShape(t *testing.T) {
	dir := t.TempDir()
	agentYAML := filepath.Join(dir, "agent.yaml")
	tokenFile := filepath.Join(dir, "etc", "agent.token")
	if err := WriteLinuxDebInstallConfig(LinuxDebPaths{
		AgentYAML: agentYAML,
		TokenFile: tokenFile,
	}, "http://h", "id1", "tok"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(agentYAML)
	if err != nil {
		t.Fatal(err)
	}
	want := "hub_url: \"http://h\"\nagent_id: \"id1\"\ntoken_file: " + tokenFile + "\ninterval_sec: 5\ntransport: http\n"
	if string(raw) != want {
		t.Fatalf("agent.yaml:\n%s\nwant:\n%s", raw, want)
	}
}

func assertInstallRoundTrip(t *testing.T, agentYAML, tokenFile, hubURL, agentID, token string) {
	cfg, err := config.LoadAgentConfig(agentYAML)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HubURL != hubURL {
		t.Fatalf("hub_url: %q", cfg.HubURL)
	}
	if cfg.AgentID != agentID {
		t.Fatalf("agent_id: %q", cfg.AgentID)
	}
	if config.LoadAgentToken(cfg) != token {
		t.Fatalf("token via LoadAgentToken")
	}
	onDisk, err := os.ReadFile(tokenFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(onDisk) != token {
		t.Fatalf("token file: %q", onDisk)
	}
	if strings.TrimSpace(token) != "" && len(onDisk) != len(token) {
		t.Fatalf("token file length mismatch (extra whitespace?)")
	}
}
