//go:build windows

package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
)

func TestPruneStagedInstallers(t *testing.T) {
	dir := t.TempDir()
	paths.OverrideForTest(dir)
	staging := filepath.Join(paths.ConfigDir, updateStagingDir)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(staging, "system-monitor-agent-1.0.70-setup.exe")
	old := filepath.Join(staging, "system-monitor-agent-1.0.69-setup.exe")
	journal := filepath.Join(staging, "update-journal.txt")
	for _, p := range []string{keep, old, journal} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := pruneStagedInstallers(staging, keep); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("expected old installer removed")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("expected kept installer: %v", err)
	}
	if _, err := os.Stat(journal); err != nil {
		t.Fatalf("expected journal kept: %v", err)
	}
	if err := CleanupUpdateStaging(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(keep); !os.IsNotExist(err) {
		t.Fatal("expected all installers removed after cleanup")
	}
}
