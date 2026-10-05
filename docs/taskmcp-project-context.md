# Workflow агента (TaskMCP)

## Перед работой
1. Прочитай `project://context` (этот текст + правила проекта ниже).
2. Для сводки бэклога — `get_planning_context` или `get_status`.
3. Не выдумывай задачи: статусы только `backlog | in_progress | done | cancelled`.

## Взять задачу
1. `claim_next_task` (exclude_blocked: true) или `claim_task` по ключу.
2. `get_task` — прочитай acceptance, notes, epic.
3. Если не можешь продолжить — `release_task`, не оставляй «висящий» claim.
4. Если упёрся в другую задачу — `block_task` + `add_note`, не молчи.

## В процессе
- Существенные решения — `add_note`.
- Мелкие правки scope — `update_task`, не плоди дубликаты.
- Новую работу без задачи — сначала спроси или `create_task`.

## Завершение (обязательно)
1. Проверь acceptance из `get_task`.
2. `complete_task` с `result` (2–5 предложений: что сделано, где смотреть).
3. Если был PR/commit — `completion_meta`: pr_url, commit, files_changed, test_status.
4. **Никогда** не оставляй задачу `in_progress` без итога.

## Ingest (если пользователь дал переписку/diff)
1. `ingest_diff` — preview без записи.
2. `ingest_transcript` — сохранить ingest.
3. `ingest://{id}` — review proposed_tasks.
4. `apply_ingest` / `decide_ingest_task` — только после явного OK пользователя.

## Эпики
- Крупная фича из нескольких задач — `create_epic` (минимум 2 задачи).
- Статус эпика — `update_epic_status`, не только задач.

---

## Проект: system-monitor (sysmon)

**Продукт:** лёгкая кроссплатформенная система мониторинга с веб-панелью на русском языке.  
Датчики, графики, пороги, live WebSocket, история в SQLite, fleet hub + agents, Android-клиент SysMon.

| | |
|---|---|
| Репозиторий | https://github.com/pagbest154-cmd/system-monitor |
| Локально | `d:\BEST154\simv1` |
| Версия | **1.0.N** (тег `v1.0.N`; см. `internal/version/version.go`, `CHANGELOG.md`) |
| Архив Python | теги `v0.0.*` (до миграции на Go) |
| Лицензия | MIT |
| Hub Docker | `ghcr.io/pagbest154-cmd/system-monitor:${VERSION}` |
| Prod hub | https://mon.ferrumnst.ru · VPS `72.56.95.92`, `/opt/system-monitor` |
| APT agent | `https://pagbest154-cmd.github.io/system-monitor/apt stable main` |

### Стек (текущий)

**Go 1.23+** · chi HTTP · gorilla/websocket · gopsutil · modernc SQLite · YAML (`gopkg.in/yaml.v3`) · vanilla JS (`web/`) · **Android** Kotlin + Compose (`android/`)

Сборка: `go test ./...`, CI на Ubuntu + Windows installer + deb + Docker + APK.

### Режимы работы

| Режим | Запуск | Описание |
|-------|--------|----------|
| **standalone** | `go run ./cmd/system-monitor --mode standalone` или `docker-compose.standalone.yml` | Collector + SQLite + UI на одной машине |
| **hub** | `docker compose up -d` | Центральный дашборд; метрики с агентов |
| **agent** | `system-monitor-agent` (.deb / Windows installer) | Push HTTP + Bearer token, без локальной БД метрик |

**Важно:** hub **не** опрашивает железо сам. Датчики на агентах; overrides — `config/agents.yaml` на hub.  
Sensor ID в fleet: `{agent_id}:{sensor_id}`.

**RAID (mdadm):** только Linux-агент с `/proc/mdstat` (например fleet **`4ov_server`**), не hub-VPS. Панель `type: raid`, пресет `auto_mdadm`, алерты «Включить все RAID».

### Дерево репозитория (ключевое)

```
cmd/system-monitor/          hub / standalone binary
cmd/system-monitor-agent/    fleet agent (+ Windows tray/service)
internal/hubserver/          HTTP API, auth, static web, WS live
internal/fleet/              ingest, agents, SyncConfig
internal/collector/          sensors (system, gpio, remote, mdadm)
internal/mdadmdiscovery/     parse /proc/mdstat, auto_discover_mdadm
internal/alerts/             пороги, ntfy, engine
internal/config/             YAML models, enrich dashboard/agents
internal/storage/            SQLite metrics + agents registry
internal/protocol/           AgentReport, MetricPoint (+ details JSON)
internal/agent/              runner, transport HTTP push
web/                         dashboard, hosts, settings (JS)
android/                     SysMon APK
config/                      sensors, dashboard, agents, alerts, hub.yaml
debian/                      system-monitor-agent .deb (+ conffiles agent.yaml)
scripts/                     release, deb, notify-taskmcp-release.sh
```

### Точки входа по типу задачи

| Задача | Файлы |
|--------|-------|
| Новый датчик | `internal/collector/`, `internal/collector/registry.go`, `config/sensors.yaml` |
| mdadm / RAID | `internal/mdadmdiscovery/`, `internal/collector/mdadm_linux.go`, `web/js/charts.js`, `RaidPanelCard.kt` |
| API / WebSocket | `internal/hubserver/server.go`, `auth.go`, `state.go` |
| Fleet ingest | `internal/fleet/fleet.go`, `internal/protocol/` |
| Agent push | `internal/agent/runner.go`, `internal/agent/transport.go` |
| YAML / enrich | `internal/config/config.go`, `config/*.yaml` |
| Web UI | `web/js/dashboard.js`, `hosts.js`, `settings.js`, `web/css/style.css` |
| Android | `android/app/src/main/kotlin/ru/ferrumnst/sysmon/` |
| Hub image | `Dockerfile`, `docker-compose.yml` |
| Agent packaging | `debian/`, `packaging/windows/`; deb install — `.cursor/rules/debian-agent-install.mdc` |
| Релиз | `internal/version/version.go`, `debian/changelog`, `android/gradle.properties`, `CHANGELOG.md`, тег `v1.0.N`, CI `.github/workflows/ci.yml` |

### Типы датчиков (примеры)

`system.cpu_percent` · `system.memory_percent` · `system.disk_usage` · `system.network_bytes` · `system.temperature` · `system.gpu_temperature` · `system.mdadm_status` · `gpio.dht22` · `remote.http_json` · `remote.mqtt`

Авто: `disk_auto_*`, `mdadm_md*`, пресеты `auto_disks`, `auto_mdadm` в dashboard.

### API (ключевое)

| Метод | Путь | Назначение |
|-------|------|------------|
| GET | `/api/mode` | standalone / hub |
| GET | `/api/sensors?agent=` | Датчики + `current` (value, status, **details**) |
| GET | `/api/metrics/{id}?period=1h&agent=` | История |
| GET | `/api/system?agent=` | Snapshot железа |
| GET | `/api/version` | Версия hub + GitHub release |
| PUT | `/api/config/dashboard` | Панели (`span`, `type: raid`, …) |
| GET/PUT | `/api/config/agents` | Реестр агентов |
| POST | `/api/agents/{id}/metrics` | Ingest (Bearer) |
| POST | `/api/agents/{id}/config` | SyncConfig overrides |
| WS | `/ws/live?agent=` | Live-обновления |

### Конвенции

- UI и пользовательские строки — **на русском**
- Sensor ID: snake_case; пороги: `warn_above`, `critical_above`
- Не ломать HTTP ingest (`AgentReport`, Bearer token, `details` на метриках)
- Hub — Docker; fleet agent — deb/Windows installer (не путать)
- Версия **1.0.N**: правки в `version.go` + changelog + deb/android sync; тег `v1.0.N` → CI (Docker, deb, APT Pages, APK, Windows)
- Автор коммитов: **pagbest154-cmd** only; не добавлять Co-authored-by Cursor
- CHANGELOG: Keep a Changelog; `[Unreleased]` перед релизом
- Канонический текст для TaskMCP: **`docs/taskmcp-project-context.md`** (синхронизировать в настройки проекта на taskmcp.ru)

### Локальный запуск

```bash
go test ./...

# hub / standalone
go run ./cmd/system-monitor --mode hub --host 127.0.0.1 --port 8080

# agent (dev)
go run ./cmd/system-monitor-agent --config config/agent.yaml
```

### Деплой

- **Hub:** `VERSION=x.y.z docker compose pull && docker compose up -d` в `/opt/system-monitor`; volumes `./config`, `hub-data`
- **Agent:** `apt install system-monitor-agent= x.y.z-1` + `systemctl restart system-monitor-agent`; conffiles сохраняют `agent.yaml`
- **Prod UI hotfix (без релиза):** `docker cp` в `/app/web/...` (теряется при recreate контейнера)

---

## System prompt (ingest / постановка задач)

Проект **system-monitor** (трекер: **sysmon**) — **Go**-мониторинг (chi HTTP, SQLite, YAML-конфиг), fleet-модель hub + agents, веб UI на vanilla JS, клиент Android SysMon. Архив Python-релизов: теги `v0.0.*`; текущие релизы: **`v1.0.N`**.

## При разборе переписки, diff и голосовых

### Различай компоненты
- **Hub** — Docker, Go `internal/hubserver`, SQLite, `config/agents.yaml`, страницы /, /settings, /hosts. Не собирает метрики сам.
- **Agent** — deb/Windows `system-monitor-agent`, `agent.yaml` + `agent_sensors.yaml`, push на hub. Без локальной БД метрик.
- **Standalone** — `cmd/system-monitor --mode standalone`: collector + SQLite + UI на одной машине; sensors/dashboard через API.
- **Android** — приложение SysMon: REST + WebSocket к hub; **не** заменяет fleet-agent.

### Формулировка задач
- Указывай режим: standalone / hub / agent / web / android.
- Acceptance: что меняется (config/API/UI/packaging), как проверить (`go test ./...`, localhost:8080, docker, deb; для RAID — Linux-агент, напр. `4ov_server`).
- Для датчиков: тип в `internal/collector/registry`, пример YAML, panel в `dashboard.yaml` при необходимости.
- Для fleet: token в `agents.yaml` на hub; prefixed sensor ID `{agent}:{sensor}`.
- Не предлагать .deb для hub (только Docker); agent — только `system-monitor-agent`.
- Не ссылаться на Python, FastAPI, pytest — только Go и `go test`.

### Типичные области
- Новый датчик → `internal/collector/` + registry + пример в `config/sensors.yaml`
- API → `internal/hubserver/server.go`, `auth.go`
- Live UI → `web/js/` + WebSocket `/ws/live`
- Fleet ingest → `internal/fleet/`, `internal/protocol/`
- Agent transport → `internal/agent/transport.go` (HTTP push — основной путь)
- Алерты / ntfy → `internal/alerts/`, `config/alerts.yaml`
- RAID mdadm → `internal/mdadmdiscovery/`, `system.mdadm_status`, панель `type: raid`
- Релиз → секция в `CHANGELOG.md`, `internal/version/version.go`, deb/android sync, тег **`v1.0.N`**, CI

### Не создавать задачи на
- Полную реализацию gRPC (только если явно в scope; сейчас не основной transport)
- Смену основного стека (не Next/React; UI — vanilla JS)
- Возврат к Python-стеку или пути `system_monitor/` / `server/routes.py`
- Дублирование hub и standalone логики без необходимости

### Приоритеты ingest
- Баги ingest/token/fleet, потеря `agent.yaml` при apt — **high**
- Новые `system.*` датчики, mdadm, алерты — **medium**
- UI polish — **low**, если не блокирует
- Документация README/CHANGELOG — вместе с фичей, не отдельным эпиком

Задачи на русском. Поля `description` и `acceptance` — конкретные, проверяемые.
