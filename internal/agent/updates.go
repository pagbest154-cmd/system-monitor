package agent

import (
	"fmt"
	"runtime"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/release"
	"github.com/pagbest154-cmd/system-monitor/internal/version"
)

const UpdateCheckIntervalSec = 24 * 60 * 60

func AgentAssetName(v string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("system-monitor-agent_%s_setup.exe", v)
	}
	return fmt.Sprintf("system-monitor-agent_%s-1_amd64.deb", v)
}

func CheckForUpdates(force bool) release.ReleaseCheckResult {
	return release.CheckReleaseUpdates(
		version.Version,
		paths.AgentUpdateCache,
		force,
		"system-monitor-agent",
		UpdateCheckIntervalSec,
		AgentAssetName,
	)
}

func FormatUpdateMessage(result release.ReleaseCheckResult) string {
	if result.UpdateAvailable {
		return fmt.Sprintf("Доступно обновление %s (установлена %s)", result.LatestVersion, result.CurrentVersion)
	}
	return fmt.Sprintf("Версия %s актуальна", result.CurrentVersion)
}

func OpenUpdatePage(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	return osStart(cmd, args)
}

func osStart(cmd string, args []string) error {
	// platform-specific in exec helper
	return execCommand(cmd, args)
}
