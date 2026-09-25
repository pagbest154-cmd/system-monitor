//go:build !windows

package agent

import (
	"fmt"

	"github.com/pagbest154-cmd/system-monitor/internal/release"
)

func ApplyUpdate(result release.ReleaseCheckResult) error {
	return fmt.Errorf("автоустановка обновлений не поддерживается на %s", "этой платформе")
}

func ApplyStagedUpdate(version string) error {
	return fmt.Errorf("автоустановка обновлений не поддерживается на %s", "этой платформе")
}
