//go:build windows

package main

import (
	"gopkg.in/toast.v1"
)

const toastAppID = "pagbest154.system-monitor.agent"

func showToast(title, message string) {
	if title == "" {
		title = "system-monitor"
	}
	if err := toast.Notification{
		AppID:   toastAppID,
		Title:   title,
		Message: message,
	}.Push(); err != nil {
		trayLog("toast failed: %v", err)
	}
}
