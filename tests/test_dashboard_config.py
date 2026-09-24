from __future__ import annotations

from system_monitor.config_loader import PanelConfig, enrich_dashboard_panels


def test_enrich_dashboard_panels_restores_default_sensors() -> None:
    panels = enrich_dashboard_panels(
        [
            PanelConfig(
                id="system_chart",
                title="Загрузка системы",
                type="line",
                sensors=[],
            ),
            PanelConfig(
                id="cpu_gauge",
                title="Процессор",
                type="gauge",
                sensors=[],
            ),
        ]
    )

    assert panels[0].sensors == ["cpu_percent", "ram_used"]
    assert panels[1].sensors == ["cpu_percent"]
