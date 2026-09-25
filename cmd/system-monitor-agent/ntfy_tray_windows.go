//go:build windows

package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/agent"
	"github.com/pagbest154-cmd/system-monitor/internal/alerts"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/fleet"
	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
)

const ntfyConfigPollSec = 30

type ntfyTrayListener struct {
	configPath string
	mu         sync.Mutex
	subs       map[string]*alerts.NtfySubscriber
	cfgKeys    map[string]string
}

func startNtfyListener(configPath string) {
	go (&ntfyTrayListener{
		configPath: configPath,
		subs:       map[string]*alerts.NtfySubscriber{},
		cfgKeys:    map[string]string{},
	}).run()
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
	ownID := fleet.DefaultAgentID(cfg.AgentID)
	token := config.LoadAgentToken(cfg)
	if token == "" {
		l.stopAll()
		setPushNotifyStatus("Уведомления: нет token")
		trayLog("ntfy: empty agent token")
		return
	}

	selected, err := config.LoadNotifySubscriptions()
	if err != nil {
		trayLog("ntfy: load subscriptions: %v", err)
	}
	if len(selected) == 0 {
		selected = []string{ownID}
	}

	transport := agent.NewTransport(cfg.HubURL, token)
	catalog, catalogErr := transport.FetchNotifyCatalog(ownID)
	if catalogErr != nil {
		trayLog("ntfy: fetch catalog: %v (fallback single host)", catalogErr)
		l.refreshLegacy(transport, ownID)
		return
	}

	byID := map[string]protocol.NotifyCatalogItem{}
	for _, item := range catalog.Items {
		if item.Enabled && strings.TrimSpace(item.Topic) != "" {
			byID[item.AgentID] = item
		}
	}
	if len(byID) == 0 {
		l.stopAll()
		setPushNotifyStatus("Уведомления: выкл. на hub")
		trayLog("ntfy: no enabled alert topics on hub")
		return
	}

	active := 0
	want := map[string]protocol.NotifyCatalogItem{}
	for _, id := range selected {
		item, ok := byID[id]
		if !ok {
			trayLog("ntfy: host %s not in catalog or alerts disabled", id)
			continue
		}
		want[id] = item
		active++
	}
	if active == 0 {
		l.stopAll()
		setPushNotifyStatus("Уведомления: выберите хосты в настройках")
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for id, sub := range l.subs {
		if _, ok := want[id]; !ok {
			sub.Stop()
			delete(l.subs, id)
			delete(l.cfgKeys, id)
		}
	}

	for id, item := range want {
		key := catalog.NtfyBaseURL + "|" + item.Topic + "|" + item.Token
		if l.cfgKeys[id] == key && l.subs[id] != nil {
			continue
		}
		if sub := l.subs[id]; sub != nil {
			sub.Stop()
		}
		hostLabel := item.Name
		if hostLabel == "" {
			hostLabel = item.AgentID
		}
		agentID := id
		topic := item.Topic
		sub := &alerts.NtfySubscriber{
			BaseURL: catalog.NtfyBaseURL,
			Topic:   item.Topic,
			Token:   item.Token,
			OnConnected: func() {
				trayLog("ntfy: websocket connected %s (%s)", hostLabel, topic)
			},
			OnError: func(err error) {
				trayLog("ntfy: websocket error %s: %v", agentID, err)
			},
			OnMessage: func(title, body string) {
				trayLog("ntfy message [%s]: %s — %s", hostLabel, title, body)
				showToast(title, body)
			},
		}
		sub.Start()
		l.subs[id] = sub
		l.cfgKeys[id] = key
		trayLog("ntfy: listening %s via %s", item.Topic, catalog.NtfyBaseURL)
	}

	if active == 1 {
		setPushNotifyStatus("Уведомления: 1 хост")
	} else {
		setPushNotifyStatus(fmt.Sprintf("Уведомления: %d хостов", active))
	}
}

func (l *ntfyTrayListener) refreshLegacy(transport *agent.Transport, ownID string) {
	notify, err := transport.FetchNotifyConfig(ownID)
	if err != nil {
		l.stopAll()
		setPushNotifyStatus("Уведомления: hub недоступен")
		trayLog("ntfy: fetch notify config from %s: %v", transport.HubURL, err)
		return
	}
	if !notify.Enabled || notify.Topic == "" {
		l.stopAll()
		setPushNotifyStatus("Уведомления: выкл. на hub")
		trayLog("ntfy: alerts disabled for %s (enabled=%v topic=%q)", ownID, notify.Enabled, notify.Topic)
		return
	}

	item := protocol.NotifyCatalogItem{
		AgentID: ownID,
		Name:    ownID,
		Enabled: true,
		Topic:   notify.Topic,
		Token:   notify.Token,
	}
	catalog := protocol.NotifyCatalogResponse{
		NtfyBaseURL: notify.NtfyBaseURL,
		Items:       []protocol.NotifyCatalogItem{item},
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for id, sub := range l.subs {
		if id != ownID {
			sub.Stop()
			delete(l.subs, id)
			delete(l.cfgKeys, id)
		}
	}

	key := catalog.NtfyBaseURL + "|" + item.Topic + "|" + item.Token
	if l.cfgKeys[ownID] == key && l.subs[ownID] != nil {
		setPushNotifyStatus("Уведомления: подключены")
		return
	}
	if sub := l.subs[ownID]; sub != nil {
		sub.Stop()
	}
	sub := &alerts.NtfySubscriber{
		BaseURL: catalog.NtfyBaseURL,
		Topic:   item.Topic,
		Token:   item.Token,
		OnConnected: func() {
			setPushNotifyStatus("Уведомления: подключены")
			trayLog("ntfy: websocket connected to %s", item.Topic)
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
	l.subs[ownID] = sub
	l.cfgKeys[ownID] = key
	setPushNotifyStatus("Уведомления: подключение…")
	trayLog("ntfy: listening %s via %s", item.Topic, catalog.NtfyBaseURL)
}

func (l *ntfyTrayListener) stopAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for id, sub := range l.subs {
		sub.Stop()
		delete(l.subs, id)
	}
	l.cfgKeys = map[string]string{}
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
