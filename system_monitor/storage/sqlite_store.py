from __future__ import annotations

import json
import sqlite3
import threading
import time
from pathlib import Path
from typing import Any

PERIOD_SECONDS = {
    "1h": 60 * 60,
    "6h": 6 * 60 * 60,
    "1d": 24 * 60 * 60,
    "1w": 7 * 24 * 60 * 60,
    "2w": 14 * 24 * 60 * 60,
    "1mo": 30 * 24 * 60 * 60,
}

_PERIOD_ALIASES = {
    "5m": "1h",
    "24h": "1d",
    "7d": "1w",
}

AGENT_OFFLINE_SEC = 90


class MetricStore:
    def __init__(self, db_path: Path) -> None:
        self.db_path = db_path
        self._lock = threading.Lock()
        self._init_db()

    def _connect(self) -> sqlite3.Connection:
        conn = sqlite3.connect(self.db_path, check_same_thread=False)
        conn.row_factory = sqlite3.Row
        return conn

    def _init_db(self) -> None:
        self.db_path.parent.mkdir(parents=True, exist_ok=True)
        with self._lock, self._connect() as conn:
            try:
                conn.execute("PRAGMA journal_mode=WAL")
            except sqlite3.OperationalError:
                conn.execute("PRAGMA journal_mode=DELETE")
            conn.execute(
                """
                CREATE TABLE IF NOT EXISTS metrics (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    sensor_id TEXT NOT NULL,
                    ts REAL NOT NULL,
                    value REAL,
                    status TEXT NOT NULL
                )
                """
            )
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_metrics_sensor_ts ON metrics(sensor_id, ts)"
            )
            conn.execute(
                """
                CREATE TABLE IF NOT EXISTS agents (
                    agent_id TEXT PRIMARY KEY,
                    name TEXT NOT NULL DEFAULT '',
                    hostname TEXT NOT NULL DEFAULT '',
                    status TEXT NOT NULL DEFAULT 'unknown',
                    last_seen REAL NOT NULL DEFAULT 0,
                    system_json TEXT,
                    sensors_json TEXT
                )
                """
            )
            conn.commit()

    def insert(self, sensor_id: str, value: float | None, status: str, ts: float | None = None) -> None:
        ts = time.time() if ts is None else ts
        with self._lock, self._connect() as conn:
            conn.execute(
                "INSERT INTO metrics (sensor_id, ts, value, status) VALUES (?, ?, ?, ?)",
                (sensor_id, ts, value, status),
            )
            conn.commit()

    def insert_batch(self, rows: list[tuple[str, float, float | None, str]]) -> None:
        if not rows:
            return
        with self._lock, self._connect() as conn:
            conn.executemany(
                "INSERT INTO metrics (sensor_id, ts, value, status) VALUES (?, ?, ?, ?)",
                rows,
            )
            conn.commit()

    def get_history(self, sensor_id: str, period: str = "1h") -> list[dict[str, Any]]:
        period = _PERIOD_ALIASES.get(period, period)
        seconds = PERIOD_SECONDS.get(period, PERIOD_SECONDS["1h"])
        since = time.time() - seconds
        with self._lock, self._connect() as conn:
            rows = conn.execute(
                """
                SELECT ts, value, status
                FROM metrics
                WHERE sensor_id = ? AND ts >= ?
                ORDER BY ts ASC
                """,
                (sensor_id, since),
            ).fetchall()
        return [
            {"ts": row["ts"], "value": row["value"], "status": row["status"]}
            for row in rows
        ]

    def get_latest(self, sensor_id: str) -> dict[str, Any] | None:
        with self._lock, self._connect() as conn:
            row = conn.execute(
                """
                SELECT ts, value, status
                FROM metrics
                WHERE sensor_id = ?
                ORDER BY ts DESC
                LIMIT 1
                """,
                (sensor_id,),
            ).fetchone()
        if row is None:
            return None
        return {"ts": row["ts"], "value": row["value"], "status": row["status"]}

    def cleanup(self, retention_days: int) -> int:
        cutoff = time.time() - retention_days * 24 * 60 * 60
        with self._lock, self._connect() as conn:
            cursor = conn.execute("DELETE FROM metrics WHERE ts < ?", (cutoff,))
            conn.commit()
            return cursor.rowcount

    def upsert_agent(
        self,
        agent_id: str,
        *,
        name: str = "",
        hostname: str = "",
        status: str = "online",
        last_seen: float | None = None,
        system: dict[str, Any] | None = None,
        sensors: list[dict[str, Any]] | None = None,
    ) -> None:
        last_seen = time.time() if last_seen is None else last_seen
        with self._lock, self._connect() as conn:
            conn.execute(
                """
                INSERT INTO agents (agent_id, name, hostname, status, last_seen, system_json, sensors_json)
                VALUES (?, ?, ?, ?, ?, ?, ?)
                ON CONFLICT(agent_id) DO UPDATE SET
                    name = excluded.name,
                    hostname = excluded.hostname,
                    status = excluded.status,
                    last_seen = excluded.last_seen,
                    system_json = excluded.system_json,
                    sensors_json = excluded.sensors_json
                """,
                (
                    agent_id,
                    name,
                    hostname,
                    status,
                    last_seen,
                    json.dumps(system, ensure_ascii=False) if system is not None else None,
                    json.dumps(sensors, ensure_ascii=False) if sensors is not None else None,
                ),
            )
            conn.commit()

    def touch_agent(self, agent_id: str, status: str = "online") -> None:
        with self._lock, self._connect() as conn:
            conn.execute(
                """
                UPDATE agents SET status = ?, last_seen = ? WHERE agent_id = ?
                """,
                (status, time.time(), agent_id),
            )
            conn.commit()

    def get_agent(self, agent_id: str) -> dict[str, Any] | None:
        with self._lock, self._connect() as conn:
            row = conn.execute(
                "SELECT * FROM agents WHERE agent_id = ?",
                (agent_id,),
            ).fetchone()
        if row is None:
            return None
        return self._agent_row_to_dict(row)

    def list_agents(self) -> list[dict[str, Any]]:
        now = time.time()
        with self._lock, self._connect() as conn:
            rows = conn.execute("SELECT * FROM agents ORDER BY name, agent_id").fetchall()
        result = []
        for row in rows:
            item = self._agent_row_to_dict(row)
            if now - item["last_seen"] > AGENT_OFFLINE_SEC:
                item["status"] = "offline"
            result.append(item)
        return result

    def _agent_row_to_dict(self, row: sqlite3.Row) -> dict[str, Any]:
        system = json.loads(row["system_json"]) if row["system_json"] else None
        sensors = json.loads(row["sensors_json"]) if row["sensors_json"] else []
        return {
            "id": row["agent_id"],
            "name": row["name"] or row["agent_id"],
            "hostname": row["hostname"],
            "status": row["status"],
            "last_seen": row["last_seen"],
            "system": system,
            "sensors": sensors,
        }
