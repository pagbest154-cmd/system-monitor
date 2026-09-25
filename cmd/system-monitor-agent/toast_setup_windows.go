//go:build windows

package main

import (
	"os"
	"path/filepath"

	"github.com/pagbest154-cmd/system-monitor/internal/branding"
	"github.com/pagbest154-cmd/system-monitor/internal/hiddenexec"
)

const toastAppID = "pagbest154.system-monitor.agent"

func ensureToastShortcut() string {
	exe, err := os.Executable()
	if err != nil {
		return findToastShortcut()
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return findToastShortcut()
	}
	dir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "SysMon agent")
	shortcut := filepath.Join(dir, branding.AgentName+".lnk")
	if _, err := os.Stat(shortcut); err == nil {
		return shortcut
	}

	script := `
$ErrorActionPreference = 'Stop'
$dir = $env:TOAST_SHORTCUT_DIR
$shortcut = $env:TOAST_SHORTCUT_PATH
$exe = $env:TOAST_EXE_PATH
$appId = $env:TOAST_APP_ID
New-Item -ItemType Directory -Force -Path $dir | Out-Null
$wsh = New-Object -ComObject WScript.Shell
$link = $wsh.CreateShortcut($shortcut)
$link.TargetPath = $exe
$link.Arguments = '--tray'
$link.WorkingDirectory = [System.IO.Path]::GetDirectoryName($exe)
$link.Description = 'SysMon agent'
$icon = Join-Path ([System.IO.Path]::GetDirectoryName($exe)) 'app-icon.ico'
if (Test-Path $icon) { $link.IconLocation = $icon }
$link.Save()
$reg = "HKCU:\Software\Classes\AppUserModelId\$appId"
New-Item -Path $reg -Force | Out-Null
Set-ItemProperty -Path $reg -Name DisplayName -Value 'SysMon agent'
if (Test-Path $icon) { Set-ItemProperty -Path $reg -Name IconUri -Value $icon }
`
	cmd := hiddenexec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-WindowStyle", "Hidden", "-Command", script)
	cmd.Env = append(os.Environ(),
		"TOAST_SHORTCUT_DIR="+dir,
		"TOAST_SHORTCUT_PATH="+shortcut,
		"TOAST_EXE_PATH="+exe,
		"TOAST_APP_ID="+toastAppID,
	)
	if err := cmd.Run(); err != nil {
		trayLog("toast shortcut setup failed: %v", err)
		return findToastShortcut()
	}
	return shortcut
}
