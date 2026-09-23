from __future__ import annotations

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

# Совместимость со старыми ключами периодов
_PERIOD_ALIASES = {
    "5m": "1h",
    "24h": "1d",
    "7d": "1w",
}


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
            conn.commit()

    def insert(self, sensor_id: str, value: float | None, status: str) -> None:
        with self._lock, self._connect() as conn:
            conn.execute(
                "INSERT INTO metrics (sensor_id, ts, value, status) VALUES (?, ?, ?, ?)",
                (sensor_id, time.time(), value, status),
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
