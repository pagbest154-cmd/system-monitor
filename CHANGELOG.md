# История изменений

Формат: [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/).  
Версия `0.0.N` — N совпадает с числом коммитов в `main`.

## [Unreleased]

### Добавлено

- Fleet-модель: **hub** (Docker) + **agents** (deb `system-monitor-agent`)
- HTTP ingest: `POST /api/agents/{id}/metrics`, SyncConfig, Bearer token
- Страница «Хосты», селектор агента на дашборде
- debconf при установке agent: hub-url, agent-id, token
- Docker-образ hub в GHCR; standalone через `docker-compose.standalone.yml`
- Протокол AgentReport (Pydantic + `proto/agent.proto`), transport HTTP/gRPC abstraction

### Изменено

- Hub/дашборд — только Docker (не .deb)
- APT repo — только `system-monitor-agent`

## [0.0.1] — 2026-09-24

Первая публичная версия **system-monitor** — лёгкая система мониторинга с веб-панелью на русском языке.

### Добавлено

- Веб-дашборд: gauge, графики, live-обновления (WebSocket)
- Датчики: CPU, RAM, диск, сеть, температура CPU/GPU
- Блок «Система»: процессор по ядрам, физические диски (HDD/SSD/NVMe), разделы, GPU, CUDA, сеть, батарея
- Автообнаружение дисков для графиков занятости
- Настройка датчиков и панелей через YAML и веб-интерфейс
- Плагины: HTTP JSON, MQTT, GPIO (Raspberry Pi / DHT22)
- SQLite-хранилище с автоочисткой истории
- Периоды графиков: 1ч, 6ч, 1д, 1н, 2н, 1мес
- Debian-пакет (`.deb`) и systemd-сервис
- Docker и docker-compose
- GitHub Actions (Node 24): сборка `.deb`, Release, APT-репозиторий на Pages

### Установка

```bash
# .deb из Releases
sudo apt install ./system-monitor_0.0.1-1_amd64.deb

# или Python
pip install .
python -m system_monitor
```

Веб-интерфейс: http://127.0.0.1:8080
