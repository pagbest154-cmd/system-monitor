package alerts

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/config"
)

type webhookPayload struct {
	AgentID  string `json:"agent_id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Kind     string `json:"kind"`
	Severity string `json:"severity"`
	SensorID string `json:"sensor_id,omitempty"`
}

func sendWebhook(cfg config.WebhookAlertConfig, payload webhookPayload) {
	url := strings.TrimSpace(cfg.URL)
	if url == "" {
		return
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if secret := strings.TrimSpace(cfg.Secret); secret != "" {
		req.Header.Set("X-Sysmon-Secret", secret)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[alerts] webhook failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("[alerts] webhook HTTP %d", resp.StatusCode)
	}
}
