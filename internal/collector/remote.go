package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type httpJsonSensor struct{ baseSensor }

func (s httpJsonSensor) Read() SensorReading {
	url, _ := s.cfg.Params["url"].(string)
	jsonPath, _ := s.cfg.Params["json_path"].(string)
	timeout := 5.0
	if t, ok := s.cfg.Params["timeout"].(float64); ok {
		timeout = t
	}
	if url == "" {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: "Не указан URL в params.url"}
	}
	client := &http.Client{Timeout: time.Duration(timeout * float64(time.Second))}
	resp, err := client.Get(url)
	if err != nil {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: err.Error()}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: err.Error()}
	}
	var payload interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: err.Error()}
	}
	raw := payload
	if jsonPath != "" {
		raw = extractJSONPath(payload, jsonPath)
	}
	value, err := toFloatValue(raw)
	if err != nil {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: err.Error()}
	}
	v := math.Round(value*1000) / 1000
	return SensorReading{SensorID: s.ID(), Value: &v, Status: evaluateStatus(s.cfg, &v)}
}

func extractJSONPath(data interface{}, path string) interface{} {
	current := data
	for _, part := range strings.Split(path, ".") {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}
	return current
}

func toFloatValue(v interface{}) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case json.Number:
		return n.Float64()
	default:
		return 0, fmt.Errorf("cannot convert value to float")
	}
}

type mqttSensor struct{ baseSensor }

var mqttListener = &mqttListenerState{
	values: map[string]float64{},
	topics: map[string]mqttTopic{},
}

type mqttListenerState struct {
	mu      sync.Mutex
	values  map[string]float64
	topics  map[string]mqttTopic
	client  mqtt.Client
	started bool
}

type mqttTopic struct {
	topic  string
	broker string
	port   int
}

func (l *mqttListenerState) register(key, topic, broker string, port int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.topics[key] = mqttTopic{topic: topic, broker: broker, port: port}
	l.ensureRunning()
}

func (l *mqttListenerState) get(key string) *float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	if v, ok := l.values[key]; ok {
		return &v
	}
	return nil
}

func (l *mqttListenerState) ensureRunning() {
	if l.started || len(l.topics) == 0 {
		return
	}
	if l.values == nil {
		l.values = map[string]float64{}
	}
	opts := mqtt.NewClientOptions()
	var first mqttTopic
	for _, t := range l.topics {
		first = t
		break
	}
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", first.broker, first.port))
	opts.SetDefaultPublishHandler(func(_ mqtt.Client, msg mqtt.Message) {
		l.mu.Lock()
		defer l.mu.Unlock()
		for key, t := range l.topics {
			if t.topic != msg.Topic() {
				continue
			}
			payload := string(msg.Payload())
			var data interface{}
			if err := json.Unmarshal(msg.Payload(), &data); err == nil {
				if m, ok := data.(map[string]interface{}); ok {
					if v, ok := m["value"]; ok {
						if f, err := toFloatValue(v); err == nil {
							l.values[key] = f
							continue
						}
					}
				}
				if f, err := toFloatValue(data); err == nil {
					l.values[key] = f
					continue
				}
			}
			if f, err := strconvParse(payload); err == nil {
				l.values[key] = f
			}
		}
	})
	l.client = mqtt.NewClient(opts)
	if token := l.client.Connect(); token.Wait() && token.Error() != nil {
		return
	}
	seen := map[string]bool{}
	for _, t := range l.topics {
		if seen[t.topic] {
			continue
		}
		seen[t.topic] = true
		l.client.Subscribe(t.topic, 0, nil)
	}
	l.started = true
}

func strconvParse(s string) (float64, error) {
	var v float64
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%f", &v)
	return v, err
}

func (s mqttSensor) Read() SensorReading {
	topic, _ := s.cfg.Params["topic"].(string)
	broker, _ := s.cfg.Params["broker"].(string)
	if broker == "" {
		broker = "localhost"
	}
	port := 1883
	if p, ok := s.cfg.Params["port"].(int); ok {
		port = p
	} else if p, ok := s.cfg.Params["port"].(float64); ok {
		port = int(p)
	}
	if topic == "" {
		return SensorReading{SensorID: s.ID(), Status: "error", Error: "Не указан topic в params.topic"}
	}
	mqttListener.register(s.ID(), topic, broker, port)
	value := mqttListener.get(s.ID())
	if value == nil {
		return SensorReading{SensorID: s.ID(), Status: "unknown", Error: "Нет данных по MQTT-топику"}
	}
	v := math.Round(*value*1000) / 1000
	return SensorReading{SensorID: s.ID(), Value: &v, Status: evaluateStatus(s.cfg, &v)}
}
