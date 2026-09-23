from __future__ import annotations

import json
import threading
import time
from typing import Any

import httpx

from .base import Sensor, SensorReading


def _extract_json_path(data: Any, path: str) -> Any:
    current = data
    for part in path.split("."):
        if isinstance(current, dict):
            current = current.get(part)
        else:
            return None
    return current


class HttpJsonSensor(Sensor):
    def read(self) -> SensorReading:
        url = self.config.params.get("url")
        json_path = self.config.params.get("json_path", "")
        timeout = float(self.config.params.get("timeout", 5))

        if not url:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="error",
                error="Не указан URL в params.url",
            )

        try:
            response = httpx.get(url, timeout=timeout)
            response.raise_for_status()
            payload = response.json()
            raw = _extract_json_path(payload, json_path) if json_path else payload
            value = float(raw)
        except Exception as exc:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="error",
                error=str(exc),
            )

        return SensorReading(
            sensor_id=self.id,
            value=round(value, 3),
            status=self.evaluate_status(value),
        )


class _MqttListener:
    _instance: _MqttListener | None = None
    _lock = threading.Lock()

    def __init__(self) -> None:
        self._values: dict[str, float] = {}
        self._topics: dict[str, tuple[str, int]] = {}
        self._client: Any = None
        self._thread: threading.Thread | None = None
        self._started = False

    @classmethod
    def instance(cls) -> _MqttListener:
        with cls._lock:
            if cls._instance is None:
                cls._instance = cls()
            return cls._instance

    def register(self, cache_key: str, topic: str, broker: str, port: int) -> None:
        self._topics[cache_key] = (topic, broker, port)
        self._ensure_running()

    def get(self, cache_key: str) -> float | None:
        return self._values.get(cache_key)

    def _ensure_running(self) -> None:
        if self._started:
            return
        try:
            import paho.mqtt.client as mqtt  # type: ignore[import-untyped]
        except ImportError:
            return

        def on_message(_client: Any, _userdata: Any, msg: Any) -> None:
            topic = msg.topic
            for key, (registered_topic, _, _) in self._topics.items():
                if registered_topic != topic:
                    continue
                try:
                    payload = msg.payload.decode("utf-8")
                    try:
                        data = json.loads(payload)
                        if isinstance(data, dict) and "value" in data:
                            self._values[key] = float(data["value"])
                        else:
                            self._values[key] = float(data)
                    except (json.JSONDecodeError, TypeError, ValueError):
                        self._values[key] = float(payload)
                except (UnicodeDecodeError, ValueError):
                    pass

        client = mqtt.Client()
        client.on_message = on_message
        first = next(iter(self._topics.values()), None)
        if first is None:
            return
        broker, port = first[1], first[2]
        client.connect(broker, port, keepalive=30)
        for topic, _, _ in self._topics.values():
            client.subscribe(topic)

        def run() -> None:
            client.loop_forever()

        self._client = client
        self._thread = threading.Thread(target=run, name="system-monitor-mqtt", daemon=True)
        self._thread.start()
        self._started = True
        time.sleep(0.1)


class MqttSensor(Sensor):
    def read(self) -> SensorReading:
        topic = self.config.params.get("topic")
        broker = self.config.params.get("broker", "localhost")
        port = int(self.config.params.get("port", 1883))

        if not topic:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="error",
                error="Не указан topic в params.topic",
            )

        listener = _MqttListener.instance()
        listener.register(self.id, topic, broker, port)

        if listener._client is None:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="error",
                error="Установите paho-mqtt: pip install paho-mqtt",
            )

        value = listener.get(self.id)
        if value is None:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="unknown",
                error="Нет данных по MQTT-топику",
            )

        return SensorReading(
            sensor_id=self.id,
            value=round(value, 3),
            status=self.evaluate_status(value),
        )
