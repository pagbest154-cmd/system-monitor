from __future__ import annotations

from .base import Sensor, SensorReading


class Dht22Sensor(Sensor):
    def read(self) -> SensorReading:
        pin = int(self.config.params.get("pin", 4))
        try:
            from gpiozero import DHT22  # type: ignore[import-untyped]
            from gpiozero.pins.native import NativeFactory  # type: ignore[import-untyped]
        except ImportError:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="error",
                error="Установите gpiozero: pip install -r requirements-pi.txt",
            )

        try:
            NativeFactory()
            sensor = DHT22(pin)
            value = sensor.temperature
        except Exception as exc:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="error",
                error=str(exc),
            )

        if value is None:
            return SensorReading(
                sensor_id=self.id,
                value=None,
                status="unknown",
                error="Не удалось считать DHT22",
            )

        value = round(float(value), 1)
        return SensorReading(
            sensor_id=self.id,
            value=value,
            status=self.evaluate_status(value),
        )
