//go:build windows

package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	_ = os.Remove(installerPath)
	if err := release.DownloadFile(*result.DownloadURL, installerPath, "system-monitor-agent"); err != nil {
		return fmt.Errorf("ошибка загрузки: %w", err)
	}
	if err := validateWindowsInstaller(installerPath); err != nil {
		return fmt.Errorf("некорректный установщик: %w", err)
	}

	return launchStagedInstaller(installerPath, result.LatestVersion)
}

func ApplyStagedUpdate(version string) error {
	stagingDir := filepath.Join(paths.ConfigDir, updateStagingDir)
	installerPath := filepath.Join(stagingDir, AgentAssetName(version))
	if err := validateWindowsInstaller(installerPath); err != nil {
		return fmt.Errorf("скачанный установщик не найден: %w", err)
	}
	return launchStagedInstaller(installerPath, version)
}

func launchStagedInstaller(installerPath, version string) error {
	stagingDir := filepath.Dir(installerPath)
	logPath := filepath.Join(stagingDir, "install.log")
	writeUpdateJournal(stagingDir, "launching "+version+" installer: "+installerPath)
	return shellExecuteElevated(installerPath, fmt.Sprintf(
		"/VERYSILENT /SUPPRESSMSGBOXES /NORESTART /CLOSEAPPLICATIONS /LOG=\"%s\"",
		logPath,
	), stagingDir)
}

func writeUpdateJournal(stagingDir, line string) {
	path := filepath.Join(stagingDir, "update-journal.txt")
	msg := fmt.Sprintf("%s  %s\n", time.Now().Format(time.RFC3339), line)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(msg)
	_ = f.Close()
	_ = os.WriteFile(filepath.Join(stagingDir, "update-status.txt"), []byte(line), 0o644)
}

func validateWindowsInstaller(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sig := make([]byte, 2)
	if _, err := f.Read(sig); err != nil {
		return err
	}
	if sig[0] != 'M' || sig[1] != 'Z' {
		return fmt.Errorf("файл не является Windows-установщиком")
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() < 500_000 {
		return fmt.Errorf("слишком маленький файл (%d байт)", info.Size())
	}
	return nil
}

func shellExecuteElevated(file, parameters, workingDir string) error {
	verb, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return err
	}
	filePtr, err := windows.UTF16PtrFromString(file)
	if err != nil {
		return err
	}
	paramPtr, err := windows.UTF16PtrFromString(parameters)
	if err != nil {
		return err
	}
	dirPtr, err := windows.UTF16PtrFromString(workingDir)
	if err != nil {
		return err
	}
	if err := windows.ShellExecute(0, verb, filePtr, paramPtr, dirPtr, windows.SW_HIDE); err != nil {
		return fmt.Errorf("не удалось запустить установщик (отклонён UAC?): %w", err)
	}
	return nil
}
