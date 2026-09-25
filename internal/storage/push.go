package storage

import (
	"database/sql"
)

type AlertState struct {
	AgentID        string
	Key            string
	LastBreached   bool
	LastNotifiedAt float64
}

func (s *MetricStore) initAlertStateSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS alert_state (
    agent_id TEXT NOT NULL,
    key TEXT NOT NULL,
    last_breached INTEGER NOT NULL DEFAULT 0,
    last_notified_at REAL NOT NULL DEFAULT 0,
    PRIMARY KEY (agent_id, key)
);`
	_, err := db.Exec(schema)
	return err
}

func (s *MetricStore) GetAlertState(agentID, key string) (AlertState, error) {
	state := AlertState{AgentID: agentID, Key: key}
	err := s.withDB(func(db *sql.DB) error {
		if err := s.initAlertStateSchema(db); err != nil {
			return err
		}
		var breached int
		err := db.QueryRow(
			`SELECT last_breached, last_notified_at FROM alert_state WHERE agent_id = ? AND key = ?`,
			agentID, key,
		).Scan(&breached, &state.LastNotifiedAt)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}
		state.LastBreached = breached != 0
		return nil
	})
	return state, err
}

func (s *MetricStore) SetAlertState(agentID, key string, breached bool, notifiedAt float64) error {
	return s.withDB(func(db *sql.DB) error {
		if err := s.initAlertStateSchema(db); err != nil {
			return err
		}
		b := 0
		if breached {
			b = 1
		}
		_, err := db.Exec(`
INSERT INTO alert_state (agent_id, key, last_breached, last_notified_at) VALUES (?, ?, ?, ?)
ON CONFLICT(agent_id, key) DO UPDATE SET
    last_breached = excluded.last_breached,
    last_notified_at = excluded.last_notified_at`,
			agentID, key, b, notifiedAt,
		)
		return err
	})
}
