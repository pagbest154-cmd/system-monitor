//go:build windows

package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateWindowsInstaller(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.exe")
	if err := os.WriteFile(bad, []byte("not-an-exe"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateWindowsInstaller(bad); err == nil {
		t.Fatal("expected error for invalid installer")
	}

	good := filepath.Join(dir, "good.exe")
	payload := []byte{'M', 'Z'}
	payload = append(payload, make([]byte, 600_000))
	if err := os.WriteFile(good, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateWindowsInstaller(good); err != nil {
		t.Fatalf("expected valid installer: %v", err)
	}
}
