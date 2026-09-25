//go:build windows

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pagbest154-cmd/system-monitor/internal/branding"
	"github.com/pagbest154-cmd/system-monitor/internal/hiddenexec"
)

const toastAppID = "pagbest154.system-monitor.agent"

func showToast(title, message string) {
	if title == "" {
		title = branding.AgentName
	}
	payload, err := json.Marshal(map[string]string{
		"title": title,
		"body":  message,
		"appId": toastAppID,
	})
	if err != nil {
		return
	}
	b64 := base64.StdEncoding.EncodeToString(payload)
	script := fmt.Sprintf(`
$p = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('%s')) | ConvertFrom-Json
$title = [Security.SecurityElement]::Escape($p.title)
$body = [Security.SecurityElement]::Escape($p.body)
$appId = $p.appId
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
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($appId).Show($toast)
`, b64)
	if err := hiddenexec.Command("powershell.exe", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script).Start(); err != nil {
		trayLog("toast failed: %v", err)
	}
}
