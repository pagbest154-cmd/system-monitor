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

const toastAppID = "pagbest154.system-monitor.agent"

func toastNotifierTarget() string {
	if shortcut := findToastShortcut(); shortcut != "" {
		return shortcut
	}
	return toastAppID
}

func findToastShortcut() string {
	appData := os.Getenv("APPDATA")
	programData := os.Getenv("ProgramData")
	candidates := []string{
		filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "SysMon agent", "SysMon agent.lnk"),
		filepath.Join(programData, "Microsoft", "Windows", "Start Menu", "Programs", "SysMon agent", "SysMon agent.lnk"),
		// legacy folder name before v1.0.24
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
	_ = runToastScript(title, message, false)
}

func showTestToast() error {
	return runToastScript(
		branding.AgentName,
		"Тестовое уведомление. Если видите это — push на Windows работает.",
		true,
	)
}

func runToastScript(title, message string, wait bool) error {
	if title == "" {
		title = branding.AgentName
	}
	notifier := strings.ReplaceAll(toastNotifierTarget(), "'", "''")
	payload, err := json.Marshal(map[string]string{
		"title":    title,
		"body":     message,
		"notifier": notifier,
	})
	if err != nil {
		return err
	}
	b64 := base64.StdEncoding.EncodeToString(payload)
	script := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$p = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('%s')) | ConvertFrom-Json
$title = [Security.SecurityElement]::Escape($p.title)
$body = [Security.SecurityElement]::Escape($p.body)
$notifier = $p.notifier
$xml = @"
<toast>
  <visual>
    <binding template="ToastGeneric">
      <text>$title</text>
      <text>$body</text>
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
`, b64)

	cmd := hiddenexec.Command("powershell.exe", "-NoProfile", "-STA", "-WindowStyle", "Hidden", "-Command", script)
	if wait {
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if msg == "" {
				msg = err.Error()
			}
			trayLog("toast failed: %v: %s", err, msg)
			return fmt.Errorf("%s", msg)
		}
		return nil
	}
	if err := cmd.Start(); err != nil {
		trayLog("toast failed: %v", err)
		return err
	}
	return nil
}
