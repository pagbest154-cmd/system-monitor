//go:build windows

package main

import (
	"log"
	"sync"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/agent"
	"github.com/pagbest154-cmd/system-monitor/internal/alerts"
	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/fleet"
)

const ntfyConfigPollSec = 60

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
	cfg, err := config.LoadAgentConfig(l.configPath)
	if err != nil {
		return
	}
	agentID := fleet.DefaultAgentID(cfg.AgentID)
	token := config.LoadAgentToken(cfg)
	if token == "" {
		l.stopSub()
		return
	}
	transport := agent.NewTransport(cfg.HubURL, token)
	notify, err := transport.FetchNotifyConfig(agentID)
	if err != nil {
		log.Printf("ntfy tray: fetch notify config: %v", err)
		return
	}
	if !notify.Enabled || notify.Topic == "" {
		l.stopSub()
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
		OnMessage: func(title, body string) {
			showToast(title, body)
		},
	}
	sub.Start()
	l.sub = sub
	log.Printf("ntfy tray: listening on topic %s", notify.Topic)
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
