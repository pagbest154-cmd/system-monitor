//go:build windows

package main

import (
	"sync"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/agent"
	"github.com/pagbest154-cmd/system-monitor/internal/alerts"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/fleet"
)

const ntfyConfigPollSec = 30

type ntfyTrayListener struct {
	configPath string
	mu         sync.Mutex
	sub        *alerts.NtfySubscriber
	cfgKey     string
}

func startNtfyListener(configPath string) {
	go (&ntfyTrayListener{configPath: configPath}).run()
}

func (l *ntfyTrayListener) run() {
	l.refresh()
	ticker := time.NewTicker(time.Duration(ntfyConfigPollSec) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		l.refresh()
	}
}

func (l *ntfyTrayListener) refresh() {
	setPushNotifyStatus("Уведомления: проверка…")

	cfg, err := config.LoadAgentConfig(l.configPath)
	if err != nil {
		setPushNotifyStatus("Уведомления: ошибка конфига")
		trayLog("ntfy: load config: %v", err)
		return
	}
	agentID := fleet.DefaultAgentID(cfg.AgentID)
	token := config.LoadAgentToken(cfg)
	if token == "" {
		l.stopSub()
		setPushNotifyStatus("Уведомления: нет token")
		trayLog("ntfy: empty agent token")
		return
	}

	transport := agent.NewTransport(cfg.HubURL, token)
	notify, err := transport.FetchNotifyConfig(agentID)
	if err != nil {
		l.stopSub()
		setPushNotifyStatus("Уведомления: hub недоступен")
		trayLog("ntfy: fetch notify config: %v", err)
		return
	}
	if !notify.Enabled || notify.Topic == "" {
		l.stopSub()
		setPushNotifyStatus("Уведомления: выкл. на hub")
		trayLog("ntfy: alerts disabled for %s", agentID)
		return
	}

	key := notify.NtfyBaseURL + "|" + notify.Topic + "|" + notify.Token
	l.mu.Lock()
	defer l.mu.Unlock()
	if key == l.cfgKey && l.sub != nil {
		return
	}
	l.stopSubLocked()
	l.cfgKey = key
	sub := &alerts.NtfySubscriber{
		BaseURL: notify.NtfyBaseURL,
		Topic:   notify.Topic,
		Token:   notify.Token,
		OnConnected: func() {
			setPushNotifyStatus("Уведомления: подключены")
			trayLog("ntfy: websocket connected to %s", notify.Topic)
		},
		OnError: func(err error) {
			setPushNotifyStatus("Уведомления: переподключение…")
			trayLog("ntfy: websocket error: %v", err)
		},
		OnMessage: func(title, body string) {
			trayLog("ntfy message: %s — %s", title, body)
			showToast(title, body)
		},
	}
	sub.Start()
	l.sub = sub
	setPushNotifyStatus("Уведомления: подключение…")
	trayLog("ntfy: listening %s via %s", notify.Topic, notify.NtfyBaseURL)
}

func (l *ntfyTrayListener) stopSub() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stopSubLocked()
}

func (l *ntfyTrayListener) stopSubLocked() {
	if l.sub != nil {
		l.sub.Stop()
		l.sub = nil
	}
	l.cfgKey = ""
}

var (
	pushStatusMu sync.Mutex
	pushStatus   = "Уведомления: …"
	pushMenuItem func(string)
)

func registerPushNotifyMenu(setTitle func(string)) {
	pushMenuItem = setTitle
}

func setPushNotifyStatus(text string) {
	pushStatusMu.Lock()
	pushStatus = text
	pushStatusMu.Unlock()
	if pushMenuItem != nil {
		pushMenuItem(text)
	}
}

func currentPushNotifyStatus() string {
	pushStatusMu.Lock()
	defer pushStatusMu.Unlock()
	return pushStatus
}
