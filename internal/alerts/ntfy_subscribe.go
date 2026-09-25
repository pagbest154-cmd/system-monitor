package alerts

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pagbest154-cmd/system-monitor/internal/branding"
)

type NtfySubscriber struct {
	BaseURL     string
	Topic       string
	Token       string
	OnMessage   func(title, body string)
	OnConnected func()
	OnError     func(error)

	mu     sync.Mutex
	stopCh chan struct{}
}

type ntfyWSMessage struct {
	Event   string `json:"event"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

func NtfyWebSocketURL(baseURL, topic string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	topic = strings.TrimSpace(topic)
	if base == "" || topic == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(base, "https://"):
		base = "wss://" + strings.TrimPrefix(base, "https://")
	case strings.HasPrefix(base, "http://"):
		base = "ws://" + strings.TrimPrefix(base, "http://")
	case strings.HasPrefix(base, "wss://") || strings.HasPrefix(base, "ws://"):
		// already websocket scheme
	default:
		base = "wss://" + base
	}
	return base + "/" + topic + "/ws"
}

func (s *NtfySubscriber) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopCh != nil {
		return
	}
	s.stopCh = make(chan struct{})
	go s.run()
}

func (s *NtfySubscriber) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopCh == nil {
		return
	}
	close(s.stopCh)
	s.stopCh = nil
}

func (s *NtfySubscriber) run() {
	s.mu.Lock()
	stopCh := s.stopCh
	s.mu.Unlock()
	if stopCh == nil {
		return
	}

	backoff := time.Second
	for {
		err := s.connectOnce(stopCh)
		select {
		case <-stopCh:
			return
		default:
		}
		if err != nil {
			log.Printf("[ntfy] subscribe error: %v", err)
			if s.OnError != nil {
				s.OnError(err)
			}
		}
		select {
		case <-stopCh:
			return
		case <-time.After(backoff):
		}
		if backoff < 60*time.Second {
			backoff *= 2
			if backoff > 60*time.Second {
				backoff = 60 * time.Second
			}
		}
	}
}

func (s *NtfySubscriber) connectOnce(stopCh chan struct{}) error {
	url := NtfyWebSocketURL(s.BaseURL, s.Topic)
	if url == "" {
		return fmt.Errorf("empty ntfy websocket url")
	}
	header := http.Header{}
	if token := strings.TrimSpace(s.Token); token != "" {
		header.Set("Authorization", "Bearer "+token)
	}
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Printf("[ntfy] subscribed to %s", s.Topic)
	if s.OnConnected != nil {
		s.OnConnected()
	}

	for {
		select {
		case <-stopCh:
			return nil
		default:
		}
		// ntfy keepalive defaults to 45s; shorter deadlines drop the socket constantly.
		_ = conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		_, data, err := conn.ReadMessage()
		if err != nil {
			select {
			case <-stopCh:
				return nil
			default:
			}
			return err
		}
		var msg ntfyWSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		switch msg.Event {
		case "open", "keepalive", "poll_request":
			continue
		case "message":
			// handled below
		default:
			continue
		}
		title := strings.TrimSpace(msg.Title)
		body := strings.TrimSpace(msg.Message)
		if title == "" {
			title = branding.AgentName
		}
		if body == "" {
			body = title
		}
		if s.OnMessage != nil {
			s.OnMessage(title, body)
		}
	}
}

func ParseNtfyWSMessage(data []byte) (title, body string, ok bool) {
	var msg ntfyWSMessage
	if err := json.Unmarshal(data, &msg); err != nil || msg.Event != "message" {
		return "", "", false
	}
	title = strings.TrimSpace(msg.Title)
	body = strings.TrimSpace(msg.Message)
	if title == "" {
		title = branding.AgentName
	}
	if body == "" {
		body = title
	}
	return title, body, true
}
