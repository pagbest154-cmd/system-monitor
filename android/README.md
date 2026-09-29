# SysMon — Android-клиент для system-monitor hub

Мобильное приложение на **Kotlin + Jetpack Compose** для мониторинга fleet-хостов через [system-monitor](https://github.com/pagbest154-cmd/system-monitor) hub.

Часть monorepo: каталог `android/` в репозитории hub/agent.

## Возможности

- Подключение к hub по URL (HTTP/HTTPS)
- Авторизация через `HUB_NAME` / `HUB_KEY` (Basic Auth)
- Список агентов с CPU/RAM и статусом online/offline
- Dashboard с gauge, line-графиками и live-обновлениями через WebSocket
- Выбор хоста и периода графиков

## Требования

- Android Studio Ladybug (2024.2+) или новее
- JDK 17
- Android SDK 35
- Min SDK 26

## Сборка

1. Откройте корень репозитория `system-monitor` в Android Studio (модуль `android/`)
2. Дождитесь синхронизации Gradle
3. Запустите на эмуляторе или устройстве (Run)

Из командной строки (после установки Android SDK):

```bash
cd android
./gradlew assembleDebug
```

APK: `app/build/outputs/apk/debug/app-debug.apk`

Release APK (как в CI):

```bash
./gradlew assembleRelease
```

## Релизы

На каждом GitHub Release публикуется `sysmon-{version}.apk` (например `sysmon-1.0.43.apk`).

Версия синхронизируется с hub/agent через `scripts/sync-android-version.sh` (вызывается из `bump-version.sh` и release workflow).

## Автообновление

Приложение проверяет GitHub Releases (`pagbest154-cmd/system-monitor`) при запуске и в **Настройки → Обновление приложения**. Если доступна новая версия — скачивает APK и открывает установщик Android.

Обновление поверх старой версии работает только при **одинаковой подписи APK**. CI подписывает релизы одним ключом (`android/release.keystore` из GitHub Secrets или закэшированный debug keystore). Если установщик пишет «Приложение не установлено» — удалите приложение и установите APK из релиза вручную.

Настройка подписи для CI (один раз, после `gh auth login`):

```powershell
.\scripts\setup-android-signing-secrets.ps1
```

## Первый запуск

1. Укажите URL хаба, например `https://monitor.example.com` или `http://192.168.1.10:8080`
2. Если на хабе включена авторизация — введите имя и ключ (`HUB_NAME` / `HUB_KEY`)
3. Выберите хост на вкладке «Хосты»
4. Смотрите метрики на вкладке «Панель»

## Архитектура

```
app/src/main/kotlin/ru/ferrumnst/sysmon/
├── data/
│   ├── api/          # Retrofit API + OkHttp
│   ├── models/       # JSON-модели hub API
│   ├── repository/   # HubRepository
│   ├── session/      # DataStore (URL, credentials, выбранный агент)
│   └── websocket/    # Live WebSocket /ws/live
└── ui/
    ├── screens/      # Setup, Login, HubUnavailable, Dashboard, Hosts, Settings
    ├── components/   # Gauge, LineChart, StatusBadge
    └── theme/        # Material 3
```

## API

Приложение использует те же endpoint'ы, что и веб-dashboard:

| Endpoint | Назначение |
|----------|------------|
| `GET /api/auth/status` | Проверка авторизации |
| `POST /api/auth/login` | Вход |
| `GET /api/agents` | Список хостов |
| `GET /api/sensors?agent=` | Датчики агента |
| `GET /api/metrics/{id}?agent=&period=` | История метрик |
| `GET /api/dashboard` | Конфиг панелей |
| `WS /ws/live?agent=` | Live snapshot/update |

Авторизация: HTTP Basic Auth (`Authorization: Basic ...`) для REST и WebSocket.

## Локальная разработка

Для HTTP без TLS в `AndroidManifest.xml` включён `usesCleartextTraffic`. Для production рекомендуется HTTPS.
