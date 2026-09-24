# История изменений

Формат: [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/).  
Версия Go-эры: `1.0.N`. Python-история: теги `v0.0.*`.

## [Unreleased]

## [1.0.0] — 2026-09-24

### Изменено

- **Полная миграция на Go**: hub, standalone, agent, packaging (Docker, .deb, Windows installer)
- Удалён Python-стек (`system_monitor/`, PyInstaller)
- Версионирование: старт Go-эры с `v1.0.0` (Python-релизы остаются в GitHub Releases как архив)

### Добавлено

- Нативные Go-бинарники агента (Linux/Windows) без runtime-зависимостей
- Фоновая очистка метрик (`cleanup`) в hub-режиме

### Совместимость

- API, YAML-конфиги, SQLite `metrics.db` и фронтенд `web/` совместимы с Python-версией
- Требуется переустановка агентов (`.deb` / `.exe`)

## [0.0.25] — 2026-09-24

### Изменено

- Linux deb: автономный бинарник (PyInstaller), без зависимостей Python на целевой системе

## [0.0.24] — 2026-09-24

### Исправлено

- Linux deb: venv и зависимости (pydantic_core) создаются при установке на целевой системе
- Требуются `python3-venv`, `python3-pip` и интернет при первой установке

## [0.0.23] — 2026-09-24

### Исправлено

- Linux deb: агент запускается через системный `python3` (venv с путями CI больше не используется)
- Linux deb: debconf задаёт Hub URL, Agent ID и token по очереди

## [0.0.22] — 2026-09-24

### Исправлено

- Linux deb: debconf снова задаёт Hub URL, Agent ID и token при установке
- Linux deb: служба запускается через `dh_installsystemd` (#DEBHELPER#)
- postinst: переменные `HUB_URL`, `AGENT_ID`, `AGENT_TOKEN` для non-interactive установки

## [0.0.21] — 2026-09-24

### Исправлено

- Agent (.deb): убран дублирующий systemd unit в `/usr/lib/systemd/system` — `dpkg -i` больше не падает, если каталог отсутствует

## [0.0.20] — 2026-09-24

### Исправлено

- Hub: панели с пустым `sensors: []` снова получают датчики по умолчанию (графики и gauge)
- Настройки hub: сохранение панелей больше не стирает привязку датчиков
- Dashboard: WebSocket snapshot не затирает уже загруженные значения

## [0.0.19] — 2026-09-24

### Изменено

- CI: hub, .deb и .exe собираются независимо — только если менялся соответствующий код

## [0.0.18] — 2026-09-24

### Добавлено

- Hub: футер с версией и проверкой обновлений на GitHub (`/api/version`)

## [0.0.17] — 2026-09-24

### Исправлено

- Windows: убран post-build rcedit — он ломал exe (ошибка PyInstaller PKG archive при запуске)
- Иконка по-прежнему вшивается через PyInstaller; ярлыки используют app-icon.ico

## [0.0.16] — 2026-09-24

### Исправлено

- Датчики с `value: null` больше не блокируют показ данных из system snapshot
- Hub `/api/sensors` подставляет current из system, если в БД только null
- Агент не пишет null-метрики; snapshot дополняется из system info
- Dashboard: gauge/графики берут данные из system; WS не затирает значения null

## [0.0.15] — 2026-09-24

### Исправлено

- Hub: метрики CPU/RAM/дисков всегда берутся из system info (исправлен ключ `disks` вместо `storage`)
- Hub: `/api/sensors?agent=` возвращает датчики даже если список в БД пустой
- Dashboard: fallback gauge/bar из `/api/system` когда current ещё нет в БД

## [0.0.14] — 2026-09-24

### Исправлено

- Windows: кастомная иконка в exe, меню Пуск и трее (rcedit + ярлыки с app-icon.ico)

## [0.0.13] — 2026-09-24

### Исправлено

- Hub: метрики CPU/RAM/дисков синтезируются из блока «Система», если коллектор агента пустой
- Agent: fallback-метрики из system info; запуск коллектора при reload
- Dashboard: загрузка истории датчиков агента без фильтра sensorMeta

## [0.0.12] — 2026-09-24

### Исправлено

- Hub: пустые графики при выбранном агенте (запрос `/api/sensors?agent=…`)
- Hub: текущие значения датчиков восстанавливаются из БД после перезапуска

## [0.0.11] — 2026-09-24

### Исправлено

- Windows: кнопка «Проверить обновления» всегда запрашивает GitHub (не использует 15‑минутный кэш)

## [0.0.10] — 2026-09-24

### Исправлено

- Windows: установщик останавливает службу и процесс трея перед переустановкой
- Windows: улучшена иконка установщика (отдельный рендер для каждого размера ICO)
- Hub: плавное обновление графиков без мерцания (setOption вместо пересоздания)

## [0.0.9] — 2026-09-24

### Добавлено

- Windows: вкладка «Датчики» в настройках — выбор метрик и автообнаружение дисков

### Исправлено

- Hub: дашборд показывал 0% после live-обновления (несовпадение ключей WebSocket)
- Hub: gauge показывает «Нет данных» вместо 0% при отсутствии значения
- Windows: улучшена иконка установщика (supersampling, больше размеров ICO)

## [0.0.8] — 2026-09-24

### Исправлено

- Windows: трей и настройки больше не открывают чёрное консольное окно (`console=False`)
- Windows: настройки из трея запускаются без лишнего окна консоли

## [0.0.7] — 2026-09-24

### Исправлено

- Windows tray: совместимость pystray на Win32 (сигнатуры menu callbacks)
- Windows tray: tooltip ограничен 128 символами (лимит Windows)
- Перезапуск службы: понятное сообщение при нехватке прав администратора

## [0.0.6] — 2026-09-24

### Исправлено

- Windows: сохранение настроек без прав администратора (ACL на `%ProgramData%\\system-monitor`)
- Windows: вставка из буфера (Ctrl+V) в полях настроек
- Windows: проверка обновлений без GitHub API (без rate limit 403)
- Windows: остановка службы перед переустановкой
- Агент: автоматическое применение изменений `agent.yaml` без перезапуска службы

## [0.0.5] — 2026-09-24

Стабильный Windows-агент и обновлённый UI.

### Добавлено

- Иконки и обновлённый UI веб-панели (favicon, иконография в дашборде, хостах, настройках)
- Иконка Windows-агента в трее и установщике

### Исправлено

- Windows-агент: запуск службы после установки (PyInstaller entry point, импорт `protocol`, UTF-8 для логов)
- Windows: `token_file` в `agent.yaml` в кавычках; NSSM `AppNoConsole` и проверка статуса службы
- Агент: fallback на встроенный `agent_sensors.yaml`, если файл в ProgramData отсутствует
- Release workflow: корректный re-tag (без дублирования `.deb` в assets)

### Изменено

- Docker-образ hub пересобирается только при изменениях hub-кода

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
