//go:build windows

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pagbest154-cmd/system-monitor/internal/hiddenexec"
)

func showToast(title, message string) {
	payload, err := json.Marshal(map[string]string{
		"title": title,
		"body":  message,
	})
	if err != nil {
		return
	}
	b64 := base64.StdEncoding.EncodeToString(payload)
	script := fmt.Sprintf(`
$p = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('%s')) | ConvertFrom-Json
Add-Type -AssemblyName System.Windows.Forms
$n = New-Object System.Windows.Forms.NotifyIcon
$n.Icon = [System.Drawing.SystemIcons]::Information
$n.Visible = $true
$n.ShowBalloonTip(8000, $p.title, $p.body, [System.Windows.Forms.ToolTipIcon]::Info)
Start-Sleep -Seconds 8
$n.Dispose()
`, b64)
	_ = hiddenexec.Command("powershell.exe", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script).Start()
}
