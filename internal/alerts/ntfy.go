package alerts

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type NtfySender struct {
	baseURL string
	client  *http.Client
}

func NtfyBaseURL() string {
	base := strings.TrimSpace(os.Getenv("NTFY_BASE_URL"))
	if base == "" {
		return "https://ntfy.sh"
	}
	return strings.TrimRight(base, "/")
}

// URL for clients (agents, phones). Use when NTFY_BASE_URL is internal (docker network).
func NtfyPublicBaseURL() string {
	base := strings.TrimSpace(os.Getenv("NTFY_PUBLIC_URL"))
	if base == "" {
		return NtfyBaseURL()
	}
	return strings.TrimRight(base, "/")
}

func NewNtfySenderFromEnv() *NtfySender {
	return &NtfySender{
		baseURL: NtfyBaseURL(),
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *NtfySender) Send(topic, token, title, body string, data map[string]string) error {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil
	}
	url := fmt.Sprintf("%s/%s", s.baseURL, topic)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	req.Header.Set("Title", title)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	if priority := ntfyPriority(data["severity"]); priority != "" {
		req.Header.Set("Priority", priority)
	}
	if tag := ntfyTag(data["severity"]); tag != "" {
		req.Header.Set("Tags", tag)
	}
	for key, value := range data {
		if value == "" {
			continue
		}
		req.Header.Set(httpHeaderName(key), value)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy POST %s: HTTP %d", url, resp.StatusCode)
	}
	log.Printf("[alerts] ntfy sent to %s: %s", topic, title)
	return nil
}

func ntfyPriority(severity string) string {
	switch severity {
	case "critical":
		return "5"
	case "warning":
		return "4"
	case "ok":
		return "3"
	default:
		return ""
	}
}

func ntfyTag(severity string) string {
	switch severity {
	case "critical":
		return "skull"
	case "warning":
		return "warning"
	case "ok":
		return "white_check_mark"
	default:
		return ""
	}
}

func httpHeaderName(key string) string {
	parts := strings.Split(strings.ReplaceAll(key, "_", "-"), "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, "-")
}
