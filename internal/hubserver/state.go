package hubserver

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type LiveHub struct {
	mu          sync.Mutex
	connections map[*websocket.Conn]bool
}

func NewLiveHub() *LiveHub {
	return &LiveHub{connections: map[*websocket.Conn]bool{}}
}

func (h *LiveHub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	h.connections[conn] = true
	h.mu.Unlock()
}

func (h *LiveHub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.connections, conn)
	h.mu.Unlock()
}

func (h *LiveHub) Broadcast(payload map[string]interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.mu.Lock()
	dead := make([]*websocket.Conn, 0)
	for conn := range h.connections {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			dead = append(dead, conn)
		}
	}
	for _, conn := range dead {
		delete(h.connections, conn)
		_ = conn.Close()
	}
	h.mu.Unlock()
}

func (h *LiveHub) ScheduleBroadcast(payload map[string]interface{}) {
	h.Broadcast(payload)
}

type FleetState struct {
	mu             sync.Mutex
	latestByAgent  map[string]map[string]map[string]interface{}
}

func NewFleetState() *FleetState {
	return &FleetState{latestByAgent: map[string]map[string]map[string]interface{}{}}
}

func (f *FleetState) UpdateAgent(agentID string, latest map[string]map[string]interface{}) {
	f.mu.Lock()
	current := f.latestByAgent[agentID]
	if current == nil {
		current = map[string]map[string]interface{}{}
		f.latestByAgent[agentID] = current
	}
	for k, v := range latest {
		current[k] = v
	}
	f.mu.Unlock()
}

func (f *FleetState) GetAgentSnapshot(agentID string) map[string]map[string]interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	current := f.latestByAgent[agentID]
	out := make(map[string]map[string]interface{}, len(current))
	for k, v := range current {
		out[k] = v
	}
	return out
}

func (f *FleetState) RemoveAgent(agentID string) {
	f.mu.Lock()
	delete(f.latestByAgent, agentID)
	f.mu.Unlock()
}

func (f *FleetState) GetCombinedSnapshot() map[string]map[string]interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	combined := map[string]map[string]interface{}{}
	for _, agentLatest := range f.latestByAgent {
		for k, v := range agentLatest {
			combined[k] = v
		}
	}
	return combined
}
