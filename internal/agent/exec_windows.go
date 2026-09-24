//go:build windows

package agent

import (
	"github.com/pagbest154-cmd/system-monitor/internal/hiddenexec"
)

func execCommand(cmd string, args []string) error {
	c := hiddenexec.Command(cmd, args...)
	return c.Start()
}
