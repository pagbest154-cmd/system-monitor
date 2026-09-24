<div align="center">

# system-monitor

**Лёгкая кроссплатформенная система мониторинга с веб-панелью на русском языке**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Python](https://img.shields.io/badge/Python-3.11%2B-3776AB?logo=python&logoColor=white)](https://www.python.org/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)](#установка)
[![Release](https://img.shields.io/github/v/release/pagbest154-cmd/system-monitor?label=release)](https://github.com/pagbest154-cmd/system-monitor/releases)

Датчики, графики и пороги — через YAML или веб-интерфейс.  
Live-обновления по WebSocket, история в SQLite.

[Установка](#установка) ·
[Возможности](#возможности) ·
[Скриншот](#интерфейс) ·
[API](#api) ·
[Changelog](CHANGELOG.md)

</div>

---

## Интерфейс

![Превью веб-панели system-monitor](docs/dashboard.svg)

---

## Возможности

| Категория | Что отслеживается |
|-----------|-------------------|
| **Процессор** | Загрузка %, модель, ядра, частота, загрузка по каждому ядру |
| **Память** | RAM, swap, объёмы и проценты |
| **Диски** | HDD / SSD / NVMe, разделы, тома, занятость (автообнаружение всех дисков) |
| **GPU** | Модель, VRAM, загрузка, CUDA (NVIDIA) |
| **Сеть** | Интерфейсы, IP, MAC, трафик ↑↓ |
| **Прочее** | Температура CPU, батарея, uptime |

**Плагины:** HTTP JSON · MQTT · GPIO (Raspberry Pi / DHT22)

**UI:** gauge · линейные графики · столбцы · live WebSocket · страница настроек

---

## Архитектура

**Fleet (hub + agents):**

```mermaid
flowchart TB
    subgraph machines [Машины]
        A1[system-monitor-agent]
        A2[system-monitor-agent]
    end
    subgraph hubHost [Hub Docker]
        API[FastAPI + WebSocket]
        DB[(SQLite)]
        UI[Dashboard]
    end
    A1 -->|HTTP push| API
    A2 --> API
    API --> DB
    DB --> UI
```

**Standalone (одна машина):**

```mermaid
flowchart LR
    COL[Collector] --> DB[(SQLite)] --> API[FastAPI] --> UI[Dashboard]
```

---

## Установка

### Hub (Docker) — центральный дашборд

```bash
git clone https://github.com/pagbest154-cmd/system-monitor.git
cd system-monitor
cp .env.example .env
# HUB_NAME и HUB_KEY — имя хаба и ключ для входа в веб-интерфейс
docker compose up -d
```

Образ: `ghcr.io/pagbest154-cmd/system-monitor:latest` · интерфейс: http://127.0.0.1:8080

**Привязка домена (HTTPS):**

```bash
cp .env.example .env
# DOMAIN=monitor.example.com  — DNS A-запись на IP сервера
docker compose --profile domain up -d
```

В настройках hub укажите тот же домен — «URL для агентов» появится автоматически.  
Caddy получит сертификат Let's Encrypt и проксирует на hub.

**Standalone (одна машина):** `docker compose -f docker-compose.standalone.yml up -d`

**Обновление hub:**

```bash
docker compose pull
docker compose up -d
```

Конкретная версия: `VERSION=0.0.20 docker compose pull && docker compose up -d`  
В футере веб-интерфейса — установленная версия и статус обновления с GitHub.

> На каждом релизе собираются **оба агента** — `.deb` (Linux) и `.exe` (Windows). Docker-образ hub пересобирается только при изменениях hub-кода. Смотрите блок «Сборка релиза» в [Releases](https://github.com/pagbest154-cmd/system-monitor/releases).

### Подключение агентов к hub

1. В [`config/agents.yaml`](config/agents.yaml) добавьте агента (id, name, token):

```yaml
agents:
  - id: homepc
    name: Домашний ПК
    token: "длинный-секретный-токен"
```

2. Установите агент на машине с **тем же** Hub URL, Agent ID и token.
3. На панели hub в выпадающем списке **Хост** выберите агента — без этого графики и блок «Система» пустые (режим «Все хосты» только для списка, не для графиков).
4. Агентов можно добавлять и в **Настройки → Агенты** в веб-интерфейсе hub.

Конфиг hub монтируется с хоста: `./config/` (в т.ч. `agents.yaml`, `dashboard.yaml`).

### Agent — slim-пакет на машинах

**Linux (.deb):**

```bash
sudo apt install ./system-monitor-agent_*_amd64.deb
```

Автономный бинарник — **Python на машине не нужен**.  
При установке debconf спросит **Hub URL**, **Agent ID** и **token** (по одному вопросу).  
На hub добавьте агента в [`config/agents.yaml`](config/agents.yaml) с тем же token.

Если вопросы не появились (узкий терминал, повторная установка):

```bash
sudo dpkg-reconfigure system-monitor-agent
```

Без интерактива:

```bash
sudo DEBIAN_FRONTEND=noninteractive \
  HUB_URL=https://monitor.example.com \
  AGENT_ID=my-laptop \
  AGENT_TOKEN=your-secret-token \
  apt install ./system-monitor-agent_*_amd64.deb
```

После установки: `sudo systemctl status system-monitor-agent` · логи: `journalctl -u system-monitor-agent -f`

APT-репозиторий (GitHub Pages при релизе):

```bash
echo "deb [trusted=yes] https://pagbest154-cmd.github.io/system-monitor/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/system-monitor.list
sudo apt update
sudo apt install system-monitor-agent
```

**Windows (установщик):**

1. Скачайте `system-monitor-agent_*_setup.exe` из [Releases](https://github.com/pagbest154-cmd/system-monitor/releases).
2. Запустите установщик — укажите **Hub URL**, **Agent ID** и **token** (как debconf на Linux).
3. Агент регистрируется как служба Windows `system-monitor-agent` и стартует автоматически.
4. После установки появляется **иконка в трее** (зелёная — подключён, красная — ошибка).

| Компонент | Способ | Путь конфигурации |
|-----------|--------|-------------------|
| Hub | Docker | `./config/` + volume `hub-data` |
| Agent (Linux) | apt | `/etc/system-monitor/agent.yaml` |
| Agent (Windows) | setup.exe | `%ProgramData%\system-monitor\agent.yaml` |

**Windows — трей и настройки:**

| Действие | Как |
|----------|-----|
| Иконка в трее | Пуск → **system-monitor agent** или `system-monitor-agent --tray` |
| Настройки | Пуск → **Настройки агента** или `system-monitor-agent --settings` |
| Логи службы | `%ProgramData%\system-monitor\agent.log` |
| Статус службы | `%ProgramData%\system-monitor\agent.status.json` |

В меню трея: статус подключения, настройки, перезапуск службы, проверка обновлений.

**Переустановка Windows-агента:** перед обновлением остановите службу и трей, иначе установщик не сможет заменить exe:

```powershell
Stop-Service system-monitor-agent -ErrorAction SilentlyContinue
taskkill /IM system-monitor-agent.exe /F
```

Затем запустите новый `system-monitor-agent_*_setup.exe` из [Releases](https://github.com/pagbest154-cmd/system-monitor/releases).

**Проверка версии и обновлений (Windows и Linux):**

```bash
system-monitor-agent --version
system-monitor-agent --check-update
```

```bash
system-monitor-agent --check-update
```

- код `0` — версия актуальна  
- код `2` — доступно обновление (подсказка по установке в выводе)  
- код `1` — ошибка проверки  

Служба агента проверяет обновления автоматически раз в сутки. На Linux через APT:

```bash
sudo apt update && sudo apt install --only-upgrade system-monitor-agent
```

### Из исходников (Python)

```bash
git clone https://github.com/pagbest154-cmd/system-monitor.git
cd system-monitor
python -m venv .venv

# Windows
.venv\Scripts\activate

# Linux / macOS
source .venv/bin/activate

pip install -r requirements.txt
pip install ".[hub]"
python -m system_monitor --mode hub --host 0.0.0.0 --port 8080
```

**Agent (dev):**

```bash
pip install -r requirements-agent.txt
pip install ".[windows]"   # трей и WMI на Windows
pip install .
system-monitor-agent --config config/agent.yaml
```

**Флаги агента:**

| Флаг | Описание |
|------|----------|
| `--config PATH` | Путь к `agent.yaml` |
| `--tray` | Иконка в трее (Windows) |
| `--settings` | Окно настроек (Windows) |
| `--check-update` | Проверить обновления |
| `--version` | Показать версию и выйти |

**Опционально:**

```bash
pip install -r requirements-pi.txt   # GPIO на Raspberry Pi
pip install paho-mqtt                # MQTT-датчики
```

Откройте в браузере: **http://127.0.0.1:8080**

Standalone без fleet: `python -m system_monitor --mode standalone --host 0.0.0.0 --port 8080`

---

## Диагностика

| Симптом | Что проверить |
|---------|----------------|
| Графики и gauge пустые, внизу «Система» есть данные | Выбран ли **конкретный хост** в шапке (не «Все хосты») |
| То же после обновления hub | `docker compose pull && up -d`, затем **Ctrl+F5** в браузере |
| `Нет данных` на gauge при онлайн-агенте | Обновите hub до **0.0.20+** (исправлен пустой `sensors` в `dashboard.yaml`) |
| Агент онлайн, API пустой | `GET /api/sensors?agent=<id>` — есть ли `current` с `value` |
| История пустая | `GET /api/metrics/cpu_percent?agent=<id>&period=1h` — копятся ли `points` |
| Windows: ошибка PyInstaller PKG archive | Установите агент **0.0.17+** (сломанные сборки 0.0.14–0.0.16) |
| Linux: `pydantic_core._pydantic_core` missing (status 1) | Обновите deb до **0.0.25+** (автономный бинарник) |
| Linux: служба не стартует после обновления | `journalctl -u system-monitor-agent -f` |
| Служба не стартует после обновления | Логи: `%ProgramData%\system-monitor\agent.log` |

**Быстрая проверка API** (с cookie сессии или Basic Auth `HUB_NAME:HUB_KEY`):

```bash
curl -u "$HUB_NAME:$HUB_KEY" "https://your-hub/api/sensors?agent=homepc"
curl -u "$HUB_NAME:$HUB_KEY" "https://your-hub/api/metrics/cpu_percent?agent=homepc&period=1h"
```

Статус агента на машине: `%ProgramData%\system-monitor\agent.status.json` (поле `metrics_count` > 0).

---

## Конфигурация

<details>
<summary><strong>Датчики — <code>config/sensors.yaml</code></strong></summary>

```yaml
settings:
  retention_days: 7
  default_interval_sec: 5

sensors:
  - id: cpu_percent
    name: Загрузка процессора
    type: system.cpu_percent
    enabled: true
    interval_sec: 5
    unit: "%"
    warn_above: 80
    critical_above: 95
```

</details>

<details>
<summary><strong>Панели — <code>config/dashboard.yaml</code></strong></summary>

```yaml
dashboard:
  title: Мониторинг системы
  refresh_sec: 3

panels:
  - id: cpu_gauge
    title: Процессор
    type: gauge
    sensors: [cpu_percent]
    col: 1
    row: 1
```

</details>

---

## Типы датчиков

| Тип | Описание |
|-----|----------|
| `system.cpu_percent` | Загрузка CPU (%) |
| `system.memory_percent` | Использование RAM (%) |
| `system.disk_usage` | Занятость диска (`params.path`) |
| `system.network_bytes` | Сеть МБ/с (`params.direction`: recv/sent) |
| `system.temperature` | Температура CPU |
| `gpio.dht22` | DHT22 на Raspberry Pi (`params.pin`) |
| `remote.http_json` | HTTP API (`params.url`, `params.json_path`) |
| `remote.mqtt` | MQTT-топик (`params.topic`, `params.broker`) |

> Диски также обнаруживаются автоматически — отдельный датчик на каждый том (`disk_auto_*`).

---

## API

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/hub/info` | Домен и public URL hub |
| GET | `/api/config/hub` | Конфиг домена |
| PUT | `/api/config/hub` | Сохранить домен |
| GET | `/api/mode` | Режим: `standalone` / `hub` |
| GET | `/api/version` | Версия hub и проверка обновления на GitHub |
| GET | `/api/agents` | Список агентов (hub) |
| GET | `/api/agents/{id}/system` | Snapshot железа агента |
| POST | `/api/agents/{id}/metrics` | Ingest метрик (agent, Bearer token) |
| POST | `/api/agents/{id}/config` | SyncConfig overrides (agent) |
| GET | `/api/sensors?agent=` | Датчики агента |
| GET | `/api/metrics/{id}?agent=` | История метрик агента |
| GET | `/api/sensors` | Список датчиков и текущие значения |
| GET | `/api/metrics/{id}?period=1h` | История метрик |
| GET | `/api/system` | Информация о железе (CPU, RAM, диски, GPU…) |
| GET | `/api/dashboard` | Конфигурация панелей |
| PUT | `/api/config/sensors` | Обновить датчики |
| PUT | `/api/config/dashboard` | Обновить панели |
| WS | `/ws/live` | Live-обновления |

---

<div align="center">

**[Changelog](CHANGELOG.md)** · **[Security](SECURITY.md)** · **[License](LICENSE)**

Сделано с ❤️ для мониторинга своих серверов и ПК

</div>
