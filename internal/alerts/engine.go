package alerts

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
	"github.com/pagbest154-cmd/system-monitor/internal/storage"
)

const offlineStateKey = "__offline__"

type Engine struct {
	store  *storage.MetricStore
	sender Sender
}

type Sender interface {
	Send(topic, token, title, body string, data map[string]string) error
}

type MetricInput struct {
	AgentID    string
	AgentName  string
	SensorID   string
	SensorName string
	Unit       string
	Value      float64
}

func NewEngine(store *storage.MetricStore, sender Sender) *Engine {
	return &Engine{store: store, sender: sender}
}

func (e *Engine) EvaluateMetric(input MetricInput) {
	cfg, err := config.LoadAlertsConfig("")
	if err != nil || cfg == nil {
		return
	}
	agentCfg, ok := cfg.Alerts[input.AgentID]
	if !ok || !agentCfg.Enabled {
		return
	}
	rule := findSensorRule(agentCfg, input.SensorID)
	if rule == nil || !rule.Enabled {
		return
	}
	breached := input.Value >= rule.Threshold
	stateKey := "sensor:" + input.SensorID
	prev, _ := e.store.GetAlertState(input.AgentID, stateKey)
	if breached && !prev.LastBreached {
		if e.inCooldown(agentCfg, prev) {
			_ = e.store.SetAlertState(input.AgentID, stateKey, true, prev.LastNotifiedAt)
			return
		}
		unit := strings.TrimSpace(input.Unit)
		valueText := formatValue(input.Value, unit)
		thresholdText := formatValue(rule.Threshold, unit)
		title := input.AgentName
		body := fmt.Sprintf("%s — %s (порог %s)", input.SensorName, valueText, thresholdText)
		e.dispatch(agentCfg, title, body, map[string]string{
			"agent_id":  input.AgentID,
			"kind":      "sensor",
			"sensor_id": input.SensorID,
			"severity":  "warning",
		})
		now := float64(time.Now().UnixNano()) / 1e9
		_ = e.store.SetAlertState(input.AgentID, stateKey, true, now)
		return
	}
	if !breached && prev.LastBreached {
		_ = e.store.SetAlertState(input.AgentID, stateKey, false, prev.LastNotifiedAt)
		if agentCfg.NotifyRecovery {
			e.dispatch(agentCfg, input.AgentName, fmt.Sprintf("%s вернулся в норму", input.SensorName), map[string]string{
				"agent_id":  input.AgentID,
				"kind":      "recovery",
				"sensor_id": input.SensorID,
				"severity":  "ok",
			})
		}
	}
}

func (e *Engine) EvaluateOfflineAgents() {
	cfg, err := config.LoadAlertsConfig("")
	if err != nil || cfg == nil {
		return
	}
	agents, err := e.store.ListAgents()
	if err != nil {
		return
	}
	now := float64(time.Now().UnixNano()) / 1e9
	for _, agent := range agents {
		agentID, _ := agent["id"].(string)
		if agentID == "" {
			continue
		}
		agentCfg, ok := cfg.Alerts[agentID]
		if !ok || !agentCfg.Enabled || !agentCfg.Offline.Enabled {
			continue
		}
		afterSec := agentCfg.Offline.AfterSec
		if afterSec <= 0 {
			afterSec = storage.AgentOfflineSec
		}
		lastSeen, _ := agent["last_seen"].(float64)
		offline := now-lastSeen > float64(afterSec)
		prev, _ := e.store.GetAlertState(agentID, offlineStateKey)
		if offline && !prev.LastBreached {
			if e.inCooldown(agentCfg, prev) {
				_ = e.store.SetAlertState(agentID, offlineStateKey, true, prev.LastNotifiedAt)
				continue
			}
			name, _ := agent["name"].(string)
			if name == "" {
				name = agentID
			}
			e.dispatch(agentCfg, name, "Хост недоступен", map[string]string{
				"agent_id": agentID,
				"kind":     "offline",
				"severity": "critical",
			})
			_ = e.store.SetAlertState(agentID, offlineStateKey, true, now)
			continue
		}
		if !offline && prev.LastBreached {
			_ = e.store.SetAlertState(agentID, offlineStateKey, false, prev.LastNotifiedAt)
			if agentCfg.NotifyRecovery {
				name, _ := agent["name"].(string)
				if name == "" {
					name = agentID
				}
				e.dispatch(agentCfg, name, "Хост снова в сети", map[string]string{
					"agent_id": agentID,
					"kind":     "recovery",
					"severity": "ok",
				})
			}
		}
	}
}

func (e *Engine) StartOfflineLoop() {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			e.EvaluateOfflineAgents()
		}
	}()
}

func (e *Engine) dispatch(agentCfg config.AgentAlertConfig, title, body string, data map[string]string) {
	topic := strings.TrimSpace(agentCfg.Ntfy.Topic)
	if topic == "" {
		log.Printf("[alerts] no ntfy topic: %s — %s", title, body)
		return
	}
	if e.sender == nil {
		log.Printf("[alerts] %s: %s", title, body)
		return
	}
	if err := e.sender.Send(topic, agentCfg.Ntfy.Token, title, body, data); err != nil {
		log.Printf("[alerts] send failed: %v", err)
	}
}

func (e *Engine) inCooldown(agentCfg config.AgentAlertConfig, prev storage.AlertState) bool {
	cooldown := agentCfg.CooldownSec
	if cooldown <= 0 {
		cooldown = 900
	}
	now := float64(time.Now().UnixNano()) / 1e9
	return prev.LastNotifiedAt > 0 && now-prev.LastNotifiedAt < float64(cooldown)
}

func findSensorRule(agentCfg config.AgentAlertConfig, sensorID string) *config.AlertSensorRule {
	for i := range agentCfg.Sensors {
		if agentCfg.Sensors[i].SensorID == sensorID {
			return &agentCfg.Sensors[i]
		}
	}
	return nil
}

func formatValue(value float64, unit string) string {
	text := fmt.Sprintf("%.1f", value)
	if unit != "" {
		return text + " " + unit
	}
	return text
}
