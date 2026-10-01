package storage

import (
	"database/sql"
	"time"
)

type AlertEvent struct {
	ID       int64   `json:"id"`
	TS       float64 `json:"ts"`
	AgentID  string  `json:"agent_id"`
	Kind     string  `json:"kind"`
	Severity string  `json:"severity"`
	SensorID string  `json:"sensor_id,omitempty"`
	Title    string  `json:"title"`
	Body     string  `json:"body"`
}

func (s *MetricStore) initAlertEventsSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS alert_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ts REAL NOT NULL,
    agent_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    severity TEXT NOT NULL,
    sensor_id TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    body TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_alert_events_agent_ts ON alert_events(agent_id, ts DESC);`
	_, err := db.Exec(schema)
	return err
}

func (s *MetricStore) InsertAlertEvent(ev AlertEvent) error {
	if ev.TS == 0 {
		ev.TS = float64(time.Now().UnixNano()) / 1e9
	}
	return s.withDB(func(db *sql.DB) error {
		if err := s.initAlertEventsSchema(db); err != nil {
			return err
		}
		_, err := db.Exec(
			`INSERT INTO alert_events (ts, agent_id, kind, severity, sensor_id, title, body) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			ev.TS, ev.AgentID, ev.Kind, ev.Severity, ev.SensorID, ev.Title, ev.Body,
		)
		return err
	})
}

func (s *MetricStore) ListAlertEvents(agentID string, limit int) ([]AlertEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	var events []AlertEvent
	err := s.withDB(func(db *sql.DB) error {
		if err := s.initAlertEventsSchema(db); err != nil {
			return err
		}
		var rows *sql.Rows
		var err error
		if agentID != "" {
			rows, err = db.Query(
				`SELECT id, ts, agent_id, kind, severity, sensor_id, title, body FROM alert_events WHERE agent_id = ? ORDER BY ts DESC LIMIT ?`,
				agentID, limit,
			)
		} else {
			rows, err = db.Query(
				`SELECT id, ts, agent_id, kind, severity, sensor_id, title, body FROM alert_events ORDER BY ts DESC LIMIT ?`,
				limit,
			)
		}
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var ev AlertEvent
			if err := rows.Scan(&ev.ID, &ev.TS, &ev.AgentID, &ev.Kind, &ev.Severity, &ev.SensorID, &ev.Title, &ev.Body); err != nil {
				return err
			}
			events = append(events, ev)
		}
		return rows.Err()
	})
	return events, err
}
