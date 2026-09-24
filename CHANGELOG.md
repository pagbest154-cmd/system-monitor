# История изменений

Формат: [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/).  
Версия `0.0.N` — N совпадает с числом коммитов в `main`.

## [Unreleased]

### Исправлено

- Windows: запуск службы агента через NSSM (`AppNoConsole`, проверка статуса после старта)
- Windows: `token_file` в `agent.yaml` записывается в кавычках (корректный путь)
- Агент: fallback на встроенный `agent_sensors.yaml`, если файл в ProgramData отсутствует

### Изменено

- Release workflow: Docker-образ hub пересобирается только при изменениях hub-кода (не при релизах только агента)

## [0.0.4] — 2026-09-24

### Добавлено

- Windows-агент: иконка в системном трее (`--tray`) со статусом подключения к hub
- Windows-агент: окно настроек (`--settings`) — Hub URL, Agent ID, token, интервал
- Проверка обновлений агента (`--check-update`) для Windows и Linux (.deb / APT)
- Автоматическая проверка обновлений раз в сутки в фоновой службе агента
- Файл статуса агента (`agent.status.json`) для связи службы и трея

## [0.0.3] — 2026-09-24

### Добавлено

- Windows-установщик агента (`system-monitor-agent_*_setup.exe`): мастер Hub URL / Agent ID / token, служба через NSSM
- Сборка Windows-агента в Release workflow (PyInstaller + Inno Setup)

## [0.0.2] — 2026-09-24

Fleet hub + agents и авторизация веб-интерфейса hub.

### Добавлено

- Fleet-модель: **hub** (Docker) + **agents** (deb `system-monitor-agent`)
- Авторизация hub: `HUB_NAME` и `HUB_KEY` в `.env`, страница `/login`
- HTTP ingest: `POST /api/agents/{id}/metrics`, SyncConfig, Bearer token
- Страница «Хосты», селектор агента на дашборде
- debconf при установке agent: hub-url, agent-id, token
- Docker-образ hub в GHCR; standalone через `docker-compose.standalone.yml`
- Протокол AgentReport (Pydantic + `proto/agent.proto`), transport HTTP/gRPC abstraction

### Изменено

- Hub/дашборд — только Docker (не .deb)
- APT repo — только `system-monitor-agent`

### Исправлено

- CI: сборка `.deb` — удалён дублирующий `debian/compat`

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
