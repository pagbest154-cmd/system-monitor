//go:build windows

package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	if err := validateWindowsInstaller(installerPath); err != nil {
		return fmt.Errorf("некорректный установщик: %w", err)
	}

	logPath := filepath.Join(stagingDir, "install.log")
	scriptPath := filepath.Join(stagingDir, "run-update.ps1")
	if err := writeUpdateScript(scriptPath, installerPath, logPath); err != nil {
		return err
	}
	return launchElevatedPowerShell(scriptPath)
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

func writeUpdateScript(scriptPath, installerPath, logPath string) error {
	installerPath = strings.ReplaceAll(installerPath, "'", "''")
	logPath = strings.ReplaceAll(logPath, "'", "''")
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = "C:\\Program Files"
	}
	trayExe := strings.ReplaceAll(filepath.Join(programFiles, "system-monitor-agent", "system-monitor-agent.exe"), "'", "''")

	script := fmt.Sprintf(`$ErrorActionPreference = 'Stop'
$installer = '%s'
$log = '%s'
$tray = '%s'

Get-Process -Name 'system-monitor-agent' -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 2

$args = @(
  '/VERYSILENT',
  '/SUPPRESSMSGBOXES',
  '/NORESTART',
  '/CLOSEAPPLICATIONS',
  ('/LOG=' + $log)
)
$p = Start-Process -FilePath $installer -ArgumentList $args -Wait -PassThru
if ($p.ExitCode -ne 0) {
  Add-Type -AssemblyName System.Windows.Forms
  [System.Windows.Forms.MessageBox]::Show(
    "Не удалось установить обновление (код $($p.ExitCode)).`nЛог: $log",
    'system-monitor agent',
    [System.Windows.Forms.MessageBoxButtons]::OK,
    [System.Windows.Forms.MessageBoxIcon]::Error
  ) | Out-Null
  exit $p.ExitCode
}

if (Test-Path $tray) {
  $shell = New-Object -ComObject Shell.Application
  $shell.ShellExecute($tray, '--tray', '', '', 0) | Out-Null
}
exit 0
`, installerPath, logPath, trayExe)

	return os.WriteFile(scriptPath, []byte(script), 0o644)
}

func launchElevatedPowerShell(scriptPath string) error {
	verb, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return err
	}
	exePtr, err := windows.UTF16PtrFromString("powershell.exe")
	if err != nil {
		return err
	}
	params, err := windows.UTF16PtrFromString(
		"-NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -File \"" + scriptPath + "\"",
	)
	if err != nil {
		return err
	}
	dirPtr, err := windows.UTF16PtrFromString(filepath.Dir(scriptPath))
	if err != nil {
		return err
	}

	if err := windows.ShellExecute(0, verb, exePtr, params, dirPtr, windows.SW_HIDE); err != nil {
		return fmt.Errorf("не удалось запустить обновление (отклонён UAC?): %w", err)
	}
	return nil
}
