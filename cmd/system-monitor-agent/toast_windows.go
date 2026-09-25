//go:build windows

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pagbest154-cmd/system-monitor/internal/branding"
	"github.com/pagbest154-cmd/system-monitor/internal/hiddenexec"
)

func toastNotifierTarget() string {
	if shortcut := ensureToastShortcut(); shortcut != "" {
		return shortcut
	}
	return toastAppID
}

func findToastShortcut() string {
	appData := os.Getenv("APPDATA")
	programData := os.Getenv("ProgramData")
	candidates := []string{
		filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "SysMon agent", branding.AgentName+".lnk"),
		filepath.Join(programData, "Microsoft", "Windows", "Start Menu", "Programs", "SysMon agent", branding.AgentName+".lnk"),
		filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "SysMon agent", "SysMon agent.lnk"),
		filepath.Join(programData, "Microsoft", "Windows", "Start Menu", "Programs", "SysMon agent", "SysMon agent.lnk"),
		filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "system-monitor", "system-monitor agent.lnk"),
		filepath.Join(programData, "Microsoft", "Windows", "Start Menu", "Programs", "system-monitor", "system-monitor agent.lnk"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func showToast(title, message string) {
	go func() {
		if err := runNotificationScript(title, message, true, false); err != nil {
			trayLog("alert notification failed: %v", err)
			return
		}
		trayLog("alert notification shown: %s — %s", title, message)
	}()
}

func showTestToast() error {
	ensureToastShortcut()
	err := runNotificationScript(
		branding.AgentName,
		"Тестовое уведомление. Если видите это — push на Windows работает.",
		true,
		true,
	)
	if err != nil {
		return err
	}
	showInfo(
		branding.AgentName,
		"Уведомление отправлено.\n\nПроверьте всплывашку у часов (справа внизу) или центр уведомлений Windows.",
	)
	return nil
}

func runNotificationScript(title, message string, wait, requireVisible bool) error {
	if title == "" {
		title = branding.AgentName
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	notifier := strings.ReplaceAll(toastNotifierTarget(), "'", "''")
	exePath := strings.ReplaceAll(exe, "'", "''")
	payload, err := json.Marshal(map[string]string{
		"title":    title,
		"body":     message,
		"notifier": notifier,
		"exe":      exePath,
	})
	if err != nil {
		return err
	}
	b64 := base64.StdEncoding.EncodeToString(payload)
	holdSec := "0"
	if wait {
		holdSec = "9"
	}
	script := fmt.Sprintf(`
$ErrorActionPreference = 'Continue'
$p = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('%s')) | ConvertFrom-Json
$title = $p.title
$body = $p.body
$notifier = $p.notifier
$exe = $p.exe
$shown = $false

function Show-Balloon {
  try {
    Add-Type -AssemblyName System.Windows.Forms
    Add-Type -AssemblyName System.Drawing
    $icon = [System.Drawing.Icon]::ExtractAssociatedIcon($exe)
    $n = New-Object System.Windows.Forms.NotifyIcon
    $n.Icon = $icon
    $n.Visible = $true
    $n.BalloonTipTitle = $title
    $n.BalloonTipText = $body
    $n.BalloonTipIcon = [System.Windows.Forms.ToolTipIcon]::Info
    $n.ShowBalloonTip(10000)
    Start-Sleep -Seconds %s
    $n.Dispose()
    return $true
  } catch {
    return $false
  }
}

function Show-WinToast {
  try {
    $titleEsc = [Security.SecurityElement]::Escape($title)
    $bodyEsc = [Security.SecurityElement]::Escape($body)
    $xml = @"
<toast>
  <visual>
    <binding template="ToastGeneric">
      <text>$titleEsc</text>
      <text>$bodyEsc</text>
    </binding>
  </visual>
</toast>
"@
    [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
    [Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
    $doc = New-Object Windows.Data.Xml.Dom.XmlDocument
    $doc.LoadXml($xml)
    $toast = [Windows.UI.Notifications.ToastNotification]::new($doc)
    [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($notifier).Show($toast)
    return $true
  } catch {
    return $false
  }
}

if (Show-WinToast) { $shown = $true }
if (Show-Balloon) { $shown = $true }
if (-not $shown) { exit 2 }
exit 0
`, b64, holdSec)

	cmd := hiddenexec.Command("powershell.exe", "-NoProfile", "-STA", "-WindowStyle", "Hidden", "-Command", script)
	if wait {
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if msg == "" {
				msg = err.Error()
			}
			trayLog("notification failed: %v: %s", err, msg)
			if requireVisible {
				showError(branding.AgentName, "Не удалось показать уведомление:\n"+msg)
			}
			return fmt.Errorf("%s", msg)
		}
		trayLog("notification sent via %s", notifier)
		return nil
	}
	if err := cmd.Start(); err != nil {
		trayLog("notification failed: %v", err)
		return err
	}
	return nil
}
