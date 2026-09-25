//go:build windows

package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/release"
	"golang.org/x/sys/windows"
)

const updateStagingDir = "update-staging"

func ApplyUpdate(result release.ReleaseCheckResult) error {
	if !result.UpdateAvailable {
		return fmt.Errorf("обновление недоступно")
	}
	if result.DownloadURL == nil || *result.DownloadURL == "" {
		return fmt.Errorf("нет ссылки на установщик")
	}

	stagingDir := filepath.Join(paths.ConfigDir, updateStagingDir)
	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		return fmt.Errorf("не удалось создать каталог загрузки: %w", err)
	}

	installerPath := filepath.Join(stagingDir, AgentAssetName(result.LatestVersion))
	if err := release.DownloadFile(*result.DownloadURL, installerPath, "system-monitor-agent"); err != nil {
		return fmt.Errorf("ошибка загрузки: %w", err)
	}

	return runElevatedInstaller(installerPath)
}

func runElevatedInstaller(installerPath string) error {
	verb, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return err
	}
	exePtr, err := windows.UTF16PtrFromString(installerPath)
	if err != nil {
		return err
	}
	params, err := windows.UTF16PtrFromString("/VERYSILENT /SUPPRESSMSGBOXES /NORESTART")
	if err != nil {
		return err
	}
	dirPtr, err := windows.UTF16PtrFromString(filepath.Dir(installerPath))
	if err != nil {
		return err
	}

	if err := windows.ShellExecute(0, verb, exePtr, params, dirPtr, windows.SW_HIDE); err != nil {
		return fmt.Errorf("не удалось запустить установщик (отклонён UAC?): %w", err)
	}
	return nil
}
