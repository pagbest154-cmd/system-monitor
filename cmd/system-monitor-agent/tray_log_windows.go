//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
)

func trayLog(format string, args ...interface{}) {
	_ = os.MkdirAll(paths.ConfigDir, 0o755)
	path := filepath.Join(paths.ConfigDir, "tray.log")
	line := fmt.Sprintf("%s  %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line)
	_ = f.Close()
}
