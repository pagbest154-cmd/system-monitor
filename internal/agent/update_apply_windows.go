//go:build windows

package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pagbest154-cmd/system-monitor/internal/branding"
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
	if err := validateWindowsInstaller(installerPath); err != nil {
		if err := release.DownloadFile(*result.DownloadURL, installerPath, "system-monitor-agent"); err != nil {
			return fmt.Errorf("ошибка загрузки: %w", err)
		}
		if err := validateWindowsInstaller(installerPath); err != nil {
			return fmt.Errorf("некорректный установщик: %w", err)
		}
	}

	return launchStagedInstaller(installerPath)
}

func ApplyStagedUpdate(version string) error {
	stagingDir := filepath.Join(paths.ConfigDir, updateStagingDir)
	installerPath := filepath.Join(stagingDir, AgentAssetName(version))
	if err := validateWindowsInstaller(installerPath); err != nil {
		return fmt.Errorf("скачанный установщик не найден: %w", err)
	}
	return launchStagedInstaller(installerPath)
}

func launchStagedInstaller(installerPath string) error {
	stagingDir := filepath.Dir(installerPath)
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
	updateLog := strings.ReplaceAll(filepath.Join(filepath.Dir(installerPath), "update.log"), "'", "''")
	statusFile := strings.ReplaceAll(filepath.Join(filepath.Dir(installerPath), "update-status.txt"), "'", "''")
	agentName := strings.ReplaceAll(branding.AgentName, "'", "''")

	script := fmt.Sprintf(`$ErrorActionPreference = 'Stop'
$scriptPath = $MyInvocation.MyCommand.Path
$installer = '%s'
$log = '%s'
$tray = '%s'
$updateLog = '%s'
$statusFile = '%s'

function Write-Status([string]$Value) {
  try { Set-Content -Path $statusFile -Value $Value -Encoding UTF8 } catch {}
}

function Show-Error([string]$Message) {
  Write-Status ('failed: ' + $Message)
  try {
    Add-Type -AssemblyName System.Windows.Forms
    [System.Windows.Forms.MessageBox]::Show(
      $Message,
      '%s',
      [System.Windows.Forms.MessageBoxButtons]::OK,
      [System.Windows.Forms.MessageBoxIcon]::Error
    ) | Out-Null
  } catch {}
}

function Restart-Tray {
  if (-not (Test-Path $tray)) { return }
  try {
    $shell = New-Object -ComObject Shell.Application
    $shell.ShellExecute($tray, '--tray', '', '', 0) | Out-Null
  } catch {}
}

$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
  Start-Process powershell.exe -ArgumentList @('-NoProfile','-ExecutionPolicy','Bypass','-File',$scriptPath) -Verb RunAs
  exit 0
}

Write-Status 'running'
Start-Transcript -Path $updateLog -Force | Out-Null
try {
  & sc.exe stop system-monitor-agent 2>$null | Out-Null
  Start-Sleep -Seconds 2
  Get-Process -Name 'system-monitor-agent' -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
  Start-Sleep -Seconds 2

  Write-Status 'installing'
  $setupArgs = @(
    '/VERYSILENT',
    '/SUPPRESSMSGBOXES',
    '/NORESTART',
    '/CLOSEAPPLICATIONS',
    ('/LOG=' + $log)
  )
  $p = Start-Process -FilePath $installer -ArgumentList $setupArgs -Wait -PassThru
  if ($null -eq $p) {
    throw 'Установщик не запущен'
  }
  if ($p.ExitCode -ne 0) {
    throw ('Установщик завершился с кодом ' + $p.ExitCode + '. Лог: ' + $log)
  }

  Write-Status 'done'
  exit 0
} catch {
  $err = $_.Exception.Message
  if (-not $err) { $err = $_.ToString() }
  Show-Error ($err + [Environment]::NewLine + 'Подробности: ' + $updateLog)
  Restart-Tray
  exit 1
} finally {
  try { Stop-Transcript | Out-Null } catch {}
}
`, installerPath, logPath, trayExe, updateLog, statusFile, agentName)

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
