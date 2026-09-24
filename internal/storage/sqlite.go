package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const AgentOfflineSec = 180

var PeriodSeconds = map[string]int{
	"1h":  60 * 60,
	"6h":  6 * 60 * 60,
	"1d":  24 * 60 * 60,
	"1w":  7 * 24 * 60 * 60,
	"2w":  14 * 24 * 60 * 60,
	"1mo": 30 * 24 * 60 * 60,
}

var periodAliases = map[string]string{
	"5m":  "1h",
	"24h": "1d",
	"7d":  "1w",
}

type MetricStore struct {
	dbPath string
	mu     sync.Mutex
}

func NewMetricStore(dbPath string) (*MetricStore, error) {
	store := &MetricStore{dbPath: dbPath}
	if err := store.initDB(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *MetricStore) initDB() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.dbPath), 0o755); err != nil {
		return err
	}
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_, _ = db.Exec("PRAGMA journal_mode=DELETE")
	}
	schema := `
CREATE TABLE IF NOT EXISTS metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id TEXT NOT NULL,
    ts REAL NOT NULL,
    value REAL,
    status TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_metrics_sensor_ts ON metrics(sensor_id, ts);
CREATE TABLE IF NOT EXISTS agents (
    agent_id TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    hostname TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'unknown',
    last_seen REAL NOT NULL DEFAULT 0,
    system_json TEXT,
    sensors_json TEXT
);`
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	return nil
}

func (s *MetricStore) withDB(fn func(*sql.DB) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	return fn(db)
}

func (s *MetricStore) Insert(sensorID string, value *float64, status string, ts float64) error {
	if ts == 0 {
		ts = float64(time.Now().UnixNano()) / 1e9
	}
	return s.withDB(func(db *sql.DB) error {
		_, err := db.Exec(
			"INSERT INTO metrics (sensor_id, ts, value, status) VALUES (?, ?, ?, ?)",
			sensorID, ts, value, status,
		)
		return err
	})
}

type MetricRow struct {
	SensorID string
	TS       float64
	Value    *float64
	Status   string
}

func (s *MetricStore) InsertBatch(rows []MetricRow) error {
	if len(rows) == 0 {
		return nil
	}
	return s.withDB(func(db *sql.DB) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		stmt, err := tx.Prepare("INSERT INTO metrics (sensor_id, ts, value, status) VALUES (?, ?, ?, ?)")
		if err != nil {
			tx.Rollback()
			return err
		}
		for _, row := range rows {
			if _, err := stmt.Exec(row.SensorID, row.TS, row.Value, row.Status); err != nil {
				stmt.Close()
				tx.Rollback()
				return err
			}
		}
		stmt.Close()
		return tx.Commit()
	})
}

func (s *MetricStore) GetHistory(sensorID, period string) ([]map[string]interface{}, error) {
	if alias, ok := periodAliases[period]; ok {
		period = alias
	}
	seconds := PeriodSeconds[period]
	if seconds == 0 {
		seconds = PeriodSeconds["1h"]
	}
	since := float64(time.Now().UnixNano())/1e9 - float64(seconds)
	var points []map[string]interface{}
	err := s.withDB(func(db *sql.DB) error {
		rows, err := db.Query(
			`SELECT ts, value, status FROM metrics WHERE sensor_id = ? AND ts >= ? ORDER BY ts ASC`,
			sensorID, since,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var ts float64
			var value sql.NullFloat64
			var status string
			if err := rows.Scan(&ts, &value, &status); err != nil {
				return err
			}
			point := map[string]interface{}{"ts": ts, "status": status}
			if value.Valid {
				point["value"] = value.Float64
			} else {
				point["value"] = nil
			}
			points = append(points, point)
		}
		return rows.Err()
	})
	return points, err
}

func (s *MetricStore) ListAgentSensorIDs(agentID string) ([]string, error) {
	prefix := agentID + ":"
	var ids []string
	err := s.withDB(func(db *sql.DB) error {
		rows, err := db.Query(
			`SELECT DISTINCT sensor_id FROM metrics WHERE sensor_id LIKE ? ORDER BY sensor_id`,
			prefix+"%",
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var fullID string
			if err := rows.Scan(&fullID); err != nil {
				return err
			}
			if len(fullID) > len(prefix) {
				ids = append(ids, fullID[len(prefix):])
			}
		}
		return rows.Err()
	})
	return ids, err
}

func (s *MetricStore) GetLatest(sensorID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := s.withDB(func(db *sql.DB) error {
		var ts float64
		var value sql.NullFloat64
		var status string
		err := db.QueryRow(
			`SELECT ts, value, status FROM metrics WHERE sensor_id = ? ORDER BY ts DESC LIMIT 1`,
			sensorID,
		).Scan(&ts, &value, &status)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}
		result = map[string]interface{}{"ts": ts, "status": status}
		if value.Valid {
			result["value"] = value.Float64
		} else {
			result["value"] = nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *MetricStore) Cleanup(retentionDays int) (int64, error) {
	cutoff := float64(time.Now().UnixNano())/1e9 - float64(retentionDays*24*60*60)
	var count int64
	err := s.withDB(func(db *sql.DB) error {
		res, err := db.Exec("DELETE FROM metrics WHERE ts < ?", cutoff)
		if err != nil {
			return err
		}
		count, _ = res.RowsAffected()
		return nil
	})
	return count, err
}

type AgentUpsert struct {
	AgentID  string
	Name     string
	Hostname string
	Status   string
	LastSeen float64
	System   map[string]interface{}
	Sensors  []map[string]interface{}
}

func (s *MetricStore) UpsertAgent(u AgentUpsert) error {
	if u.LastSeen == 0 {
		u.LastSeen = float64(time.Now().UnixNano()) / 1e9
	}
	if u.Status == "" {
		u.Status = "online"
	}
	var systemJSON, sensorsJSON interface{}
	if u.System != nil {
		b, _ := json.Marshal(u.System)
		systemJSON = string(b)
	}
	if u.Sensors != nil {
		b, _ := json.Marshal(u.Sensors)
		sensorsJSON = string(b)
	}
	return s.withDB(func(db *sql.DB) error {
		_, err := db.Exec(`
INSERT INTO agents (agent_id, name, hostname, status, last_seen, system_json, sensors_json)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(agent_id) DO UPDATE SET
    name = excluded.name,
    hostname = excluded.hostname,
    status = excluded.status,
    last_seen = excluded.last_seen,
    system_json = excluded.system_json,
    sensors_json = excluded.sensors_json`,
			u.AgentID, u.Name, u.Hostname, u.Status, u.LastSeen, systemJSON, sensorsJSON,
		)
		return err
	})
}

func (s *MetricStore) TouchAgent(agentID, status string) error {
	now := float64(time.Now().UnixNano()) / 1e9
	return s.withDB(func(db *sql.DB) error {
		_, err := db.Exec(`UPDATE agents SET status = ?, last_seen = ? WHERE agent_id = ?`, status, now, agentID)
		return err
	})
}

func (s *MetricStore) GetAgent(agentID string) (map[string]interface{}, error) {
	var record map[string]interface{}
	err := s.withDB(func(db *sql.DB) error {
		var name, hostname, status string
		var lastSeen float64
		var systemJSON, sensorsJSON sql.NullString
		err := db.QueryRow(`SELECT agent_id, name, hostname, status, last_seen, system_json, sensors_json FROM agents WHERE agent_id = ?`, agentID).Scan(
			&agentID, &name, &hostname, &status, &lastSeen, &systemJSON, &sensorsJSON,
		)
		if err == sql.ErrNoRows {
			return fmt.Errorf("not found")
		}
		if err != nil {
			return err
		}
		record = agentRowToDict(agentID, name, hostname, status, lastSeen, systemJSON, sensorsJSON)
		return nil
	})
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		return nil, err
	}
	return record, nil
}

func (s *MetricStore) ListAgents() ([]map[string]interface{}, error) {
	now := float64(time.Now().UnixNano()) / 1e9
	var agents []map[string]interface{}
	err := s.withDB(func(db *sql.DB) error {
		rows, err := db.Query(`SELECT agent_id, name, hostname, status, last_seen, system_json, sensors_json FROM agents ORDER BY name, agent_id`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var agentID, name, hostname, status string
			var lastSeen float64
			var systemJSON, sensorsJSON sql.NullString
			if err := rows.Scan(&agentID, &name, &hostname, &status, &lastSeen, &systemJSON, &sensorsJSON); err != nil { //nolint:rowserrcheck
				return err
			}
			item := agentRowToDict(agentID, name, hostname, status, lastSeen, systemJSON, sensorsJSON)
			if now-item["last_seen"].(float64) > AgentOfflineSec {
				item["status"] = "offline"
			}
			agents = append(agents, item)
		}
		return rows.Err()
	})
	return agents, err
}

func agentRowToDict(agentID, name, hostname, status string, lastSeen float64, systemJSON, sensorsJSON sql.NullString) map[string]interface{} {
	var system map[string]interface{}
	var sensors []interface{}
	if systemJSON.Valid && systemJSON.String != "" {
		_ = json.Unmarshal([]byte(systemJSON.String), &system)
	}
	if sensorsJSON.Valid && sensorsJSON.String != "" {
		_ = json.Unmarshal([]byte(sensorsJSON.String), &sensors)
	}
	displayName := name
	if displayName == "" {
		displayName = agentID
	}
	return map[string]interface{}{
		"id":        agentID,
		"name":      displayName,
		"hostname":  hostname,
		"status":    status,
		"last_seen": lastSeen,
		"system":    system,
		"sensors":   sensors,
	}
}
